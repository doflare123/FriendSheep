package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"notify_service/migrations"
)

const migrationLockID int64 = 731_047_292_641

func Migrate(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию миграций notify_service: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", migrationLockID); err != nil {
		return fmt.Errorf("не удалось получить блокировку миграций notify_service: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS notify_service;
		CREATE TABLE IF NOT EXISTS notify_service.schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("не удалось подготовить историю миграций notify_service: %w", err)
	}

	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("не удалось прочитать встроенные миграции notify_service: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	seenVersions := make(map[int64]string)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		version, err := migrationVersion(entry.Name())
		if err != nil {
			return err
		}
		if previousName, exists := seenVersions[version]; exists {
			return fmt.Errorf("версия миграции %d повторяется в %q и %q", version, previousName, entry.Name())
		}
		seenVersions[version] = entry.Name()

		body, err := fs.ReadFile(migrations.Files, entry.Name())
		if err != nil {
			return fmt.Errorf("не удалось прочитать миграцию notify_service %q: %w", entry.Name(), err)
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(body))

		var appliedName, appliedChecksum string
		err = tx.QueryRowContext(ctx,
			"SELECT name, checksum FROM notify_service.schema_migrations WHERE version = $1",
			version,
		).Scan(&appliedName, &appliedChecksum)
		switch {
		case err == nil:
			if appliedName != entry.Name() || appliedChecksum != checksum {
				return fmt.Errorf("применённая миграция версии %d не совпадает со встроенной историей", version)
			}
			continue
		case errors.Is(err, sql.ErrNoRows):
			// Миграция ещё не применена.
		case err != nil:
			return fmt.Errorf("не удалось проверить миграцию notify_service версии %d: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("не удалось применить миграцию notify_service %q: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO notify_service.schema_migrations (version, name, checksum) VALUES ($1, $2, $3)",
			version, entry.Name(), checksum,
		); err != nil {
			return fmt.Errorf("не удалось записать миграцию notify_service %q: %w", entry.Name(), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось зафиксировать миграции notify_service: %w", err)
	}
	return nil
}

func migrationVersion(name string) (int64, error) {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, fmt.Errorf("некорректное имя миграции notify_service %q", name)
	}
	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil || version <= 0 {
		return 0, fmt.Errorf("некорректное имя миграции notify_service %q", name)
	}
	return version, nil
}
