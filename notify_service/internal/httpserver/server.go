package httpserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"notify_service/internal/config"
	"notify_service/internal/reminders"
)

type Readiness interface {
	PingContext(context.Context) error
}

type Notifications interface {
	ListNotifications(context.Context, uint64, string, int, bool) (reminders.NotificationPage, error)
	CountUnread(context.Context, uint64) (int, error)
	MarkAsRead(context.Context, uint64, string) (reminders.NotificationRecord, error)
}

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
	logger          *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger, readiness Readiness, notifications Notifications) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Address(),
			Handler:           newHandler(logger, readiness, notifications, cfg.InternalToken, cfg.DatabaseReadyTimeout),
			ReadTimeout:       cfg.HTTPReadTimeout,
			ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
			WriteTimeout:      cfg.HTTPWriteTimeout,
			IdleTimeout:       cfg.HTTPIdleTimeout,
		},
		shutdownTimeout: cfg.HTTPShutdownTimeout,
		logger:          logger,
	}
}

func NewHandler(logger *slog.Logger, readiness Readiness, notifications Notifications, internalToken string) http.Handler {
	return newHandler(logger, readiness, notifications, internalToken, 2*time.Second)
}

func newHandler(logger *slog.Logger, readiness Readiness, notifications Notifications, internalToken string, readinessTimeout time.Duration) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if readinessTimeout <= 0 {
		readinessTimeout = 2 * time.Second
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		readyCtx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()
		if readiness == nil || readiness.PingContext(readyCtx) != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	if notifications != nil {
		internal := http.NewServeMux()
		internal.HandleFunc("GET /internal/v1/users/{userId}/notifications", func(w http.ResponseWriter, r *http.Request) {
			userID, ok := parseUserID(w, r)
			if !ok {
				return
			}
			query := r.URL.Query()
			after := strings.TrimSpace(query.Get("after"))
			limit := 50
			if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
				parsed, err := strconv.Atoi(raw)
				if err != nil {
					writeError(w, http.StatusBadRequest, reminders.ErrorCodeInvalidLimit)
					return
				}
				limit = parsed
			}
			unreadOnly := false
			if raw := strings.TrimSpace(query.Get("unread")); raw != "" {
				parsed, err := strconv.ParseBool(raw)
				if err != nil {
					writeError(w, http.StatusBadRequest, reminders.ErrorCodeInvalidUnreadFilter)
					return
				}
				unreadOnly = parsed
			}
			page, err := notifications.ListNotifications(r.Context(), userID, after, limit, unreadOnly)
			if err != nil {
				writeDomainError(w, err)
				return
			}
			writeAnyJSON(w, http.StatusOK, page)
		})
		internal.HandleFunc("GET /internal/v1/users/{userId}/notifications/unread-count", func(w http.ResponseWriter, r *http.Request) {
			userID, ok := parseUserID(w, r)
			if !ok {
				return
			}
			count, err := notifications.CountUnread(r.Context(), userID)
			if err != nil {
				writeDomainError(w, err)
				return
			}
			writeAnyJSON(w, http.StatusOK, map[string]int{"unreadCount": count})
		})
		internal.HandleFunc("PATCH /internal/v1/users/{userId}/notifications/{notificationId}/read", func(w http.ResponseWriter, r *http.Request) {
			userID, ok := parseUserID(w, r)
			if !ok {
				return
			}
			notificationID := strings.TrimSpace(r.PathValue("notificationId"))
			record, err := notifications.MarkAsRead(r.Context(), userID, notificationID)
			if err != nil {
				writeDomainError(w, err)
				return
			}
			writeAnyJSON(w, http.StatusOK, record)
		})
		mux.Handle("/internal/", requireInternalToken(internalToken, internal))
	}
	return requestLogger(logger, mux)
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("не удалось слушать %s: %w", s.httpServer.Addr, err)
	}
	return s.Serve(ctx, listener)
}

// Serve владеет goroutine HTTP-сервера и дожидается её завершения перед возвратом.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- s.httpServer.Serve(listener)
	}()

	select {
	case err := <-serveDone:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("ошибка HTTP-сервера: %w", err)
	case <-ctx.Done():
	}

	s.logger.Info("начато корректное завершение HTTP-сервера")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	shutdownErr := s.httpServer.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		s.logger.Warn("не удалось корректно завершить HTTP-сервер; соединения будут закрыты", "ошибка", shutdownErr)
		_ = s.httpServer.Close()
	}

	serveErr := <-serveDone
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("ошибка HTTP-сервера во время завершения: %w", serveErr)
	}
	if shutdownErr != nil {
		return fmt.Errorf("не удалось завершить HTTP-сервер: %w", shutdownErr)
	}
	s.logger.Info("HTTP-сервер корректно завершён")
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body map[string]string) {
	writeAnyJSON(w, status, body)
}

func writeAnyJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeAnyJSON(w, status, map[string]string{"error": code})
}

func writeDomainError(w http.ResponseWriter, err error) {
	var clientErr *reminders.ClientError
	if !errors.As(err, &clientErr) {
		writeError(w, http.StatusInternalServerError, reminders.ErrorCodeUnexpectedStatus)
		return
	}
	switch clientErr.Code {
	case reminders.ErrorCodeNotificationNotFound:
		writeError(w, http.StatusNotFound, clientErr.Code)
	case reminders.ErrorCodeInvalidCursor, reminders.ErrorCodeInvalidUserID, reminders.ErrorCodeInvalidLimit, reminders.ErrorCodeInvalidUnreadFilter:
		writeError(w, http.StatusBadRequest, clientErr.Code)
	default:
		if clientErr.Terminal {
			writeError(w, http.StatusBadRequest, clientErr.Code)
			return
		}
		writeError(w, http.StatusServiceUnavailable, clientErr.Code)
	}
}

func parseUserID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	userID, err := strconv.ParseUint(strings.TrimSpace(r.PathValue("userId")), 10, 64)
	if err != nil || userID == 0 {
		writeError(w, http.StatusBadRequest, reminders.ErrorCodeInvalidUserID)
		return 0, false
	}
	return userID, true
}

func requireInternalToken(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		if expected == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			writeError(w, http.StatusUnauthorized, reminders.ErrorCodeInvalidInternalToken)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		route := r.Pattern
		if route == "" {
			route = "не_сопоставлен"
		}
		logger.Info("HTTP-запрос",
			"метод", r.Method,
			"маршрут", route,
			"статус", wrapped.status,
			"длительность_мс", time.Since(started).Milliseconds(),
		)
	})
}
