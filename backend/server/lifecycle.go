package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultShutdownTimeout = 30 * time.Second
	resourceCloseTimeout   = 5 * time.Second
)

// run управляет жизненным циклом процесса. Сигналы SIGINT и SIGTERM прекращают
// прием новых HTTP-соединений, ожидают активные запросы и затем закрывают ресурсы.
func (s *Server) run(addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Join(fmt.Errorf("listen on %s: %w", addr, err), s.closeResources())
	}

	return s.serve(ctx, listener, defaultShutdownTimeout)
}

// serve отделен от обработки сигналов, чтобы сценарий завершения можно было
// проверить в тестах с настоящим сетевым слушателем.
func (s *Server) serve(ctx context.Context, listener net.Listener, shutdownTimeout time.Duration) (resultErr error) {
	httpServer := &http.Server{
		Handler: s.engine,
	}

	defer func() {
		resultErr = errors.Join(resultErr, s.closeResources())
	}()
	return serveHTTPWithShutdownHook(ctx, httpServer, listener, shutdownTimeout, s.stopBackground, s.logger)
}

func serveHTTP(ctx context.Context, httpServer *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	return serveHTTPWithShutdownHook(ctx, httpServer, listener, shutdownTimeout, nil, nil)
}

func serveHTTPWithShutdownHook(
	ctx context.Context,
	httpServer *http.Server,
	listener net.Listener,
	shutdownTimeout time.Duration,
	beforeShutdown func(),
	log interface {
		Info(string, ...interface{})
		Error(string, ...interface{})
	},
) (resultErr error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultShutdownTimeout
	}

	requestCtx, cancelRequests := context.WithCancel(context.Background())
	httpServer.BaseContext = func(net.Listener) context.Context {
		return requestCtx
	}
	defer cancelRequests()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()

	if log != nil {
		log.Info("Server running", "addr", listener.Addr().String())
	}

	select {
	case err := <-serveErr:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		cancelRequests()
		_ = httpServer.Close()
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		if log != nil {
			log.Info("Graceful shutdown started", "timeout", shutdownTimeout)
		}
	}

	// Фоновые задачи должны прекратить создавать новую работу до ожидания HTTP-запросов.
	if beforeShutdown != nil {
		beforeShutdown()
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	shutdownErr := httpServer.Shutdown(shutdownCtx)
	cancelShutdown()
	if shutdownErr == nil {
		cancelRequests()
		<-serveErr
		if log != nil {
			log.Info("Graceful shutdown completed")
		}
		return nil
	}

	// Shutdown не отменяет контексты активных запросов. После истечения срока ожидания
	// общий контекст отменяется до закрытия соединений, чтобы транзакции базы данных,
	// использующие контекст, откатились и не продолжили выполнение.
	cancelRequests()
	closeErr := httpServer.Close()
	<-serveErr
	if log != nil {
		log.Error("Graceful shutdown deadline exceeded; active requests were canceled", "error", shutdownErr)
	}

	return errors.Join(
		fmt.Errorf("shutdown HTTP server: %w", shutdownErr),
		wrapCloseError("force close HTTP server", closeErr),
	)
}

func (s *Server) stopBackground() {
	s.backgroundStopOnce.Do(func() {
		if s.popularEventsService != nil {
			s.popularEventsService.Stop()
		}
	})
}

func (s *Server) closeResources() error {
	s.resourcesCloseOnce.Do(func() {
		// Этот вызов также покрывает ошибки запуска слушателя, при которых обычная
		// обработка сигнала не успевает остановить фоновую работу.
		s.stopBackground()

		var errs []error
		if s.redis != nil && s.redis.Client() != nil {
			errs = append(errs, wrapCloseError("close Redis", s.redis.Client().Close()))
		}
		if s.mongo != nil && s.mongo.DB() != nil && s.mongo.DB().Client() != nil {
			ctx, cancel := context.WithTimeout(context.Background(), resourceCloseTimeout)
			err := s.mongo.DB().Client().Disconnect(ctx)
			cancel()
			errs = append(errs, wrapCloseError("disconnect MongoDB", err))
		}
		if s.postgres != nil {
			errs = append(errs, wrapCloseError("close PostgreSQL", s.postgres.Close()))
		}

		s.resourcesCloseErr = errors.Join(errs...)
	})
	return s.resourcesCloseErr
}

func wrapCloseError(operation string, err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
