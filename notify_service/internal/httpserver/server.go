package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"notify_service/internal/config"
)

type Readiness interface {
	PingContext(context.Context) error
}

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
	logger          *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger, readiness Readiness) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Address(),
			Handler:           newHandler(logger, readiness, cfg.DatabaseReadyTimeout),
			ReadTimeout:       cfg.HTTPReadTimeout,
			ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
			WriteTimeout:      cfg.HTTPWriteTimeout,
			IdleTimeout:       cfg.HTTPIdleTimeout,
		},
		shutdownTimeout: cfg.HTTPShutdownTimeout,
		logger:          logger,
	}
}

func NewHandler(logger *slog.Logger, readiness Readiness) http.Handler {
	return newHandler(logger, readiness, 2*time.Second)
}

func newHandler(logger *slog.Logger, readiness Readiness, readinessTimeout time.Duration) http.Handler {
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
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
