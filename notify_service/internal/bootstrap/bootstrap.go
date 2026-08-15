package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"notify_service/internal/application"
	"notify_service/internal/config"
)

type Database interface {
	PingContext(context.Context) error
	Close() error
}

type Dependencies struct {
	OpenDatabase     func(context.Context, config.Config) (Database, error)
	Migrate          func(context.Context, Database) error
	BuildApplication func() (*application.Application, error)
	ServeHTTP        func(context.Context, config.Config, *slog.Logger, Database, *application.Application) error
}

// Run выполняет последовательность запуска: база данных -> миграции -> application -> HTTP
// и всегда закрывает собственную базу данных сервиса перед возвратом.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, deps Dependencies) (runErr error) {
	if logger == nil {
		return errors.New("требуется logger")
	}
	if deps.OpenDatabase == nil || deps.Migrate == nil || deps.BuildApplication == nil || deps.ServeHTTP == nil {
		return errors.New("требуется полный набор bootstrap-зависимостей")
	}

	db, err := deps.OpenDatabase(ctx, cfg)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать базу данных: %w", err)
	}
	if db == nil {
		return errors.New("инициализация базы данных вернула nil")
	}
	logger.Info("соединение с базой данных установлено")
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("не удалось закрыть соединение с базой данных", "ошибка", err)
			runErr = errors.Join(runErr, fmt.Errorf("не удалось закрыть базу данных: %w", err))
			return
		}
		logger.Info("соединение с базой данных закрыто")
	}()

	if err := deps.Migrate(ctx, db); err != nil {
		return fmt.Errorf("не удалось применить миграции базы данных: %w", err)
	}
	logger.Info("миграции базы данных применены")
	app, err := deps.BuildApplication()
	if err != nil {
		return fmt.Errorf("не удалось собрать application-слой: %w", err)
	}
	if app == nil {
		return errors.New("сборка application-слоя вернула nil")
	}
	logger.Info("application-слой собран", "каналы_уведомлений", app.ChannelCount())

	if err := deps.ServeHTTP(ctx, cfg, logger, db, app); err != nil {
		return fmt.Errorf("не удалось запустить HTTP-сервер: %w", err)
	}
	return nil
}
