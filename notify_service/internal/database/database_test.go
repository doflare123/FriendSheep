package database_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"notify_service/internal/database"
)

func TestOpenDoesNotExposePasswordFromInvalidURL(t *testing.T) {
	t.Parallel()

	const password = "unikalnyj-parol-dlya-proverki"
	databaseURL := "postgres://notify:" + password + "%zz@database:5432/notifications"
	db, err := database.Open(context.Background(), databaseURL, 100*time.Millisecond)
	if db != nil {
		_ = db.Close()
		t.Fatal("Open() вернул соединение для синтаксически некорректного URL")
	}
	if err == nil {
		t.Fatal("Open() не вернул ошибку для синтаксически некорректного URL")
	}
	if strings.Contains(err.Error(), password) {
		t.Fatalf("ошибка Open() раскрывает пароль базы данных: %v", err)
	}
}
