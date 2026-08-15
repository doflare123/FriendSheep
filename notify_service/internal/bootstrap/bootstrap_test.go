package bootstrap_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"notify_service/internal/application"
	"notify_service/internal/bootstrap"
	"notify_service/internal/config"
)

type fakeDatabase struct {
	closeErr  error
	closeCall int
}

func (*fakeDatabase) PingContext(context.Context) error { return nil }

func (db *fakeDatabase) Close() error {
	db.closeCall++
	return db.closeErr
}

func TestRunExecutesStartupInOrderAndClosesDatabase(t *testing.T) {
	t.Parallel()

	var order []string
	db := &fakeDatabase{}
	ctx := context.WithValue(context.Background(), bootstrapContextKey{}, "preserved")
	deps := bootstrap.Dependencies{
		OpenDatabase: func(gotCtx context.Context, _ config.Config) (bootstrap.Database, error) {
			assertBootstrapContext(t, gotCtx)
			order = append(order, "database")
			return db, nil
		},
		Migrate: func(gotCtx context.Context, gotDB bootstrap.Database) error {
			assertBootstrapContext(t, gotCtx)
			if gotDB != db {
				t.Fatal("Migrate получил другое подключение к базе данных")
			}
			order = append(order, "migrations")
			return nil
		},
		BuildApplication: func() (*application.Application, error) {
			order = append(order, "application")
			return application.New(), nil
		},
		ServeHTTP: func(gotCtx context.Context, _ config.Config, _ *slog.Logger, gotDB bootstrap.Database, app *application.Application) error {
			assertBootstrapContext(t, gotCtx)
			if gotDB != db {
				t.Fatal("ServeHTTP получил другое подключение к базе данных")
			}
			if app == nil {
				t.Fatal("ServeHTTP получил nil вместо приложения")
			}
			order = append(order, "http")
			return nil
		},
	}

	if err := bootstrap.Run(ctx, config.Config{}, discardLogger(), deps); err != nil {
		t.Fatalf("Run() завершился ошибкой: %v", err)
	}
	order = append(order, "closed")
	if db.closeCall != 1 {
		t.Fatalf("число вызовов Close базы данных = %d, ожидалось 1", db.closeCall)
	}
	wantOrder := []string{"database", "migrations", "application", "http", "closed"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Errorf("порядок запуска = %v, ожидался %v", order, wantOrder)
	}
}

func TestRunStopsOnMigrationFailureAndClosesDatabase(t *testing.T) {
	t.Parallel()

	migrationErr := errors.New("ошибка миграции")
	db := &fakeDatabase{}
	buildCalled := false
	serveCalled := false
	deps := bootstrap.Dependencies{
		OpenDatabase: func(context.Context, config.Config) (bootstrap.Database, error) { return db, nil },
		Migrate:      func(context.Context, bootstrap.Database) error { return migrationErr },
		BuildApplication: func() (*application.Application, error) {
			buildCalled = true
			return application.New(), nil
		},
		ServeHTTP: func(context.Context, config.Config, *slog.Logger, bootstrap.Database, *application.Application) error {
			serveCalled = true
			return nil
		},
	}

	err := bootstrap.Run(context.Background(), config.Config{}, discardLogger(), deps)
	if !errors.Is(err, migrationErr) {
		t.Fatalf("ошибка Run() = %v, ожидалась ошибка миграции", err)
	}
	if db.closeCall != 1 {
		t.Errorf("число вызовов Close базы данных = %d, ожидалось 1", db.closeCall)
	}
	if buildCalled || serveCalled {
		t.Errorf("после ошибки миграции вызваны следующие этапы запуска: сборка=%t сервер=%t", buildCalled, serveCalled)
	}
}

func TestRunJoinsServeAndDatabaseCloseFailures(t *testing.T) {
	t.Parallel()

	serveErr := errors.New("ошибка HTTP-сервера")
	closeErr := errors.New("ошибка закрытия")
	db := &fakeDatabase{closeErr: closeErr}
	deps := successfulDependencies(db)
	deps.ServeHTTP = func(context.Context, config.Config, *slog.Logger, bootstrap.Database, *application.Application) error {
		return serveErr
	}

	err := bootstrap.Run(context.Background(), config.Config{}, discardLogger(), deps)
	if !errors.Is(err, serveErr) {
		t.Errorf("ошибка Run() = %v, ожидалась ошибка HTTP-сервера", err)
	}
	if !errors.Is(err, closeErr) {
		t.Errorf("ошибка Run() = %v, ожидалась ошибка закрытия базы данных", err)
	}
	if db.closeCall != 1 {
		t.Errorf("число вызовов Close базы данных = %d, ожидалось 1", db.closeCall)
	}
}

func TestRunStopsOnApplicationCompositionFailureAndClosesDatabase(t *testing.T) {
	t.Parallel()

	composeErr := errors.New("ошибка сборки application-слоя")
	db := &fakeDatabase{}
	serveCalled := false
	deps := successfulDependencies(db)
	deps.BuildApplication = func() (*application.Application, error) {
		return nil, composeErr
	}
	deps.ServeHTTP = func(context.Context, config.Config, *slog.Logger, bootstrap.Database, *application.Application) error {
		serveCalled = true
		return nil
	}

	err := bootstrap.Run(context.Background(), config.Config{}, discardLogger(), deps)
	if !errors.Is(err, composeErr) {
		t.Fatalf("ошибка Run() = %v, ожидалась ошибка сборки приложения", err)
	}
	if db.closeCall != 1 {
		t.Errorf("число вызовов Close базы данных = %d, ожидалось 1", db.closeCall)
	}
	if serveCalled {
		t.Fatal("HTTP-сервер запущен после ошибки сборки приложения")
	}
}

func TestRunRejectsNilDatabase(t *testing.T) {
	t.Parallel()

	deps := successfulDependencies(nil)
	err := bootstrap.Run(context.Background(), config.Config{}, discardLogger(), deps)
	if err == nil {
		t.Fatal("Run() не вернул ошибку для nil-подключения к базе данных")
	}
}

func TestRunDoesNotContinueAfterDatabaseOpenFailure(t *testing.T) {
	t.Parallel()

	openErr := errors.New("ошибка открытия")
	called := false
	deps := bootstrap.Dependencies{
		OpenDatabase: func(context.Context, config.Config) (bootstrap.Database, error) { return nil, openErr },
		Migrate: func(context.Context, bootstrap.Database) error {
			called = true
			return nil
		},
		BuildApplication: func() (*application.Application, error) {
			called = true
			return application.New(), nil
		},
		ServeHTTP: func(context.Context, config.Config, *slog.Logger, bootstrap.Database, *application.Application) error {
			called = true
			return nil
		},
	}

	err := bootstrap.Run(context.Background(), config.Config{}, discardLogger(), deps)
	if !errors.Is(err, openErr) {
		t.Fatalf("ошибка Run() = %v, ожидалась ошибка открытия базы данных", err)
	}
	if called {
		t.Fatal("запуск продолжился после ошибки открытия базы данных")
	}
}

func TestRunRejectsIncompleteDependencies(t *testing.T) {
	t.Parallel()

	if err := bootstrap.Run(context.Background(), config.Config{}, discardLogger(), bootstrap.Dependencies{}); err == nil {
		t.Fatal("Run() не вернул ошибку для неполного набора зависимостей")
	}
	if err := bootstrap.Run(context.Background(), config.Config{}, nil, bootstrap.Dependencies{}); err == nil {
		t.Fatal("Run() не вернул ошибку для отсутствующего журнала")
	}
}

type bootstrapContextKey struct{}

func assertBootstrapContext(t *testing.T, ctx context.Context) {
	t.Helper()
	if got := ctx.Value(bootstrapContextKey{}); got != "preserved" {
		t.Errorf("значение контекста запроса = %v, ожидалось preserved", got)
	}
}

func successfulDependencies(db bootstrap.Database) bootstrap.Dependencies {
	return bootstrap.Dependencies{
		OpenDatabase:     func(context.Context, config.Config) (bootstrap.Database, error) { return db, nil },
		Migrate:          func(context.Context, bootstrap.Database) error { return nil },
		BuildApplication: func() (*application.Application, error) { return application.New(), nil },
		ServeHTTP: func(context.Context, config.Config, *slog.Logger, bootstrap.Database, *application.Application) error {
			return nil
		},
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
