package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMigrationVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    int64
		wantErr bool
	}{
		{name: "000001_initial_schema.sql", want: 1},
		{name: "000042_add_jobs.sql", want: 42},
		{name: "000001.sql", wantErr: true},
		{name: "invalid.sql", wantErr: true},
		{name: "000000_invalid.sql", wantErr: true},
		{name: "-1_invalid.sql", wantErr: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := migrationVersion(test.name)
			if test.wantErr {
				if err == nil {
					t.Fatalf("migrationVersion(%q) не вернул ошибку", test.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("migrationVersion(%q) завершился ошибкой: %v", test.name, err)
			}
			if got != test.want {
				t.Errorf("migrationVersion(%q) = %d, ожидалось %d", test.name, got, test.want)
			}
		})
	}
}

func TestMigratePostgresIsIdempotent(t *testing.T) {
	dsn := os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NOTIFY_SERVICE_TEST_POSTGRES_DSN не задан; интеграционный тест миграций PostgreSQL пропущен")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := Open(ctx, dsn, 5*time.Second)
	if err != nil {
		t.Fatalf("не удалось открыть тестовый PostgreSQL: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("первый вызов Migrate() завершился ошибкой: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("второй вызов Migrate() завершился ошибкой: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notify_service.schema_migrations
		WHERE version IN (1, 2, 3)
	`).Scan(&count); err != nil {
		t.Fatalf("не удалось прочитать историю миграций: %v", err)
	}
	if count != 3 {
		t.Fatalf("число записей миграций 1, 2 и 3 = %d, ожидалось 3", count)
	}

	for _, table := range []string{
		"event_lifecycle_source_cursors",
		"event_lifecycle_source_messages",
		"event_lifecycle_jobs",
		"event_reminder_source_cursors",
		"event_reminder_source_messages",
		"event_reminder_jobs",
		"notifications",
		"notification_delivery_targets",
	} {
		var exists bool
		if err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'notify_service' AND table_name = $1
			)
		`, table).Scan(&exists); err != nil {
			t.Fatalf("не удалось проверить таблицу %s: %v", table, err)
		}
		if !exists {
			t.Errorf("миграции не создали таблицу notify_service.%s", table)
		}
	}
}
