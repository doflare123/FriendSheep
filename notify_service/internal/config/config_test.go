package config_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"notify_service/internal/config"
)

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(mapEnvironment(map[string]string{
		"NOTIFY_DATABASE_URL": "postgres://notify:secret@database:5432/notifications?sslmode=disable",
	}))
	if err != nil {
		t.Fatalf("не удалось загрузить конфигурацию: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, ожидалось 8080", cfg.Port)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, ожидался info", cfg.LogLevel)
	}
	assertPositiveDuration(t, "HTTPReadTimeout", cfg.HTTPReadTimeout)
	assertPositiveDuration(t, "HTTPReadHeaderTimeout", cfg.HTTPReadHeaderTimeout)
	assertPositiveDuration(t, "HTTPWriteTimeout", cfg.HTTPWriteTimeout)
	assertPositiveDuration(t, "HTTPIdleTimeout", cfg.HTTPIdleTimeout)
	assertPositiveDuration(t, "HTTPShutdownTimeout", cfg.HTTPShutdownTimeout)
	assertPositiveDuration(t, "DatabaseConnectTimeout", cfg.DatabaseConnectTimeout)
	assertDuration(t, "DatabaseReadyTimeout", cfg.DatabaseReadyTimeout, 2*time.Second)
}

func TestLoadReadsExplicitEnvironment(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(mapEnvironment(map[string]string{
		"NOTIFY_DATABASE_URL":             "postgresql://notify:secret@database:5432/notifications",
		"PORT":                            "9090",
		"NOTIFY_LOG_LEVEL":                "debug",
		"NOTIFY_HTTP_READ_TIMEOUT":        "2s",
		"NOTIFY_HTTP_READ_HEADER_TIMEOUT": "3s",
		"NOTIFY_HTTP_WRITE_TIMEOUT":       "4s",
		"NOTIFY_HTTP_IDLE_TIMEOUT":        "5s",
		"NOTIFY_HTTP_SHUTDOWN_TIMEOUT":    "6s",
		"NOTIFY_DATABASE_CONNECT_TIMEOUT": "7s",
		"NOTIFY_DATABASE_READY_TIMEOUT":   "8s",
	}))
	if err != nil {
		t.Fatalf("не удалось загрузить конфигурацию: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, ожидалось 9090", cfg.Port)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, ожидался debug", cfg.LogLevel)
	}
	assertDuration(t, "HTTPReadTimeout", cfg.HTTPReadTimeout, 2*time.Second)
	assertDuration(t, "HTTPReadHeaderTimeout", cfg.HTTPReadHeaderTimeout, 3*time.Second)
	assertDuration(t, "HTTPWriteTimeout", cfg.HTTPWriteTimeout, 4*time.Second)
	assertDuration(t, "HTTPIdleTimeout", cfg.HTTPIdleTimeout, 5*time.Second)
	assertDuration(t, "HTTPShutdownTimeout", cfg.HTTPShutdownTimeout, 6*time.Second)
	assertDuration(t, "DatabaseConnectTimeout", cfg.DatabaseConnectTimeout, 7*time.Second)
	assertDuration(t, "DatabaseReadyTimeout", cfg.DatabaseReadyTimeout, 8*time.Second)
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  map[string]string
	}{
		{name: "URL базы данных отсутствует", env: map[string]string{}},
		{name: "неподдерживаемый URL базы данных", env: map[string]string{"NOTIFY_DATABASE_URL": "mysql://notify:secret@database/notifications"}},
		{name: "некорректный порт", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "PORT": "70000"}},
		{name: "некорректный уровень логирования", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_LOG_LEVEL": "verbose"}},
		{name: "некорректная длительность", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_HTTP_READ_TIMEOUT": "soon"}},
		{name: "нулевая длительность", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_HTTP_SHUTDOWN_TIMEOUT": "0s"}},
		{name: "отрицательная длительность", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_DATABASE_CONNECT_TIMEOUT": "-1s"}},
		{name: "нулевой таймаут readiness", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_DATABASE_READY_TIMEOUT": "0s"}},
		{name: "отрицательный таймаут readiness", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_DATABASE_READY_TIMEOUT": "-1s"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := config.Load(mapEnvironment(test.env)); err == nil {
				t.Fatal("Load() не вернул ошибку валидации")
			}
		})
	}
}

func TestValidateRejectsEmptyDatabaseURL(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(mapEnvironment(map[string]string{
		"NOTIFY_DATABASE_URL": validDatabaseURL(),
	}))
	if err != nil {
		t.Fatalf("не удалось загрузить корректную конфигурацию: %v", err)
	}
	cfg.DatabaseURL = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() не вернул ошибку отсутствующего URL базы данных")
	}
}

func TestValidateRejectsNonPositiveDatabaseReadyTimeout(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(mapEnvironment(map[string]string{
		"NOTIFY_DATABASE_URL": validDatabaseURL(),
	}))
	if err != nil {
		t.Fatalf("не удалось загрузить корректную конфигурацию: %v", err)
	}
	cfg.DatabaseReadyTimeout = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() не вернул ошибку для нулевого таймаута проверки готовности")
	}
}

func TestLoadErrorDoesNotExposeDatabaseCredentials(t *testing.T) {
	t.Parallel()

	const password = "do-not-log-this-password"
	_, err := config.Load(mapEnvironment(map[string]string{
		"NOTIFY_DATABASE_URL": "postgres://notify:" + password + "@database:5432/notifications",
		"NOTIFY_LOG_LEVEL":    "invalid",
	}))
	if err == nil {
		t.Fatal("Load() не вернул ошибку некорректного уровня логирования")
	}
	if strings.Contains(err.Error(), password) {
		t.Fatalf("ошибка конфигурации раскрывает пароль базы данных: %v", err)
	}
}

func validDatabaseURL() string {
	return "postgres://notify:secret@database:5432/notifications?sslmode=disable"
}

func mapEnvironment(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func assertPositiveDuration(t *testing.T, name string, value time.Duration) {
	t.Helper()
	if value <= 0 {
		t.Errorf("%s = %s, ожидалась положительная длительность", name, value)
	}
}

func assertDuration(t *testing.T, name string, got, want time.Duration) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %s, ожидалось %s", name, got, want)
	}
}
