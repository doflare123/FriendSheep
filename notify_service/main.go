package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"notify_service/internal/application"
	"notify_service/internal/bootstrap"
	"notify_service/internal/config"
	"notify_service/internal/database"
	"notify_service/internal/httpserver"
	"notify_service/internal/lifecycle"
	"notify_service/internal/logging"
)

func main() {
	if err := run(); err != nil {
		logging.New(os.Stderr, slog.LevelInfo).Error("notify_service остановлен с ошибкой", "ошибка", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}
	logger := logging.New(os.Stdout, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return bootstrap.Run(ctx, cfg, logger, bootstrap.Dependencies{
		OpenDatabase: func(ctx context.Context, cfg config.Config) (bootstrap.Database, error) {
			return database.Open(ctx, cfg.DatabaseURL, cfg.DatabaseConnectTimeout)
		},
		Migrate: func(ctx context.Context, connection bootstrap.Database) error {
			db, ok := connection.(*sql.DB)
			if !ok {
				return errors.New("неподдерживаемое соединение с базой данных уведомлений")
			}
			return database.Migrate(ctx, db)
		},
		BuildApplication: func() (*application.Application, error) {
			// Пользовательские каналы доставки намеренно отсутствуют в P0.2b.
			return application.New(), nil
		},
		ServeHTTP: func(
			ctx context.Context,
			cfg config.Config,
			logger *slog.Logger,
			connection bootstrap.Database,
			_ *application.Application,
		) (serveErr error) {
			db, ok := connection.(*sql.DB)
			if !ok {
				return errors.New("неподдерживаемое соединение с базой данных notify_service")
			}
			manager := lifecycle.NewManager(cfg, logger, db, nil)
			if err := manager.Start(ctx); err != nil {
				return err
			}
			defer func() {
				if err := manager.Stop(cfg.HTTPShutdownTimeout); err != nil {
					logger.Warn("lifecycle manager завершился по таймауту", "ошибка", err)
					serveErr = errors.Join(serveErr, err)
				}
			}()

			server := httpserver.New(cfg, logger, manager)
			logger.Info("HTTP-сервер запускается", "адрес", cfg.Address())
			return server.ListenAndServe(ctx)
		},
	})
}
