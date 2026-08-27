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
		"NOTIFY_DATABASE_URL":      "postgres://notify:secret@database:5432/notifications?sslmode=disable",
		"NOTIFY_MONOLITH_BASE_URL": "http://backend:8081",
		"NOTIFY_SERVICE_TOKEN":     "notify-service-token",
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
	assertDuration(t, "MonolithHTTPTimeout", cfg.MonolithHTTPTimeout, 5*time.Second)
	assertDuration(t, "SourcePollInterval", cfg.SourcePollInterval, 5*time.Second)
	assertDuration(t, "SourceRetryMinBackoff", cfg.SourceRetryMinBackoff, time.Second)
	assertDuration(t, "SourceRetryMaxBackoff", cfg.SourceRetryMaxBackoff, 30*time.Second)
	assertDuration(t, "WorkerScanInterval", cfg.WorkerScanInterval, time.Second)
	assertDuration(t, "WorkerRetryMinBackoff", cfg.WorkerRetryMinBackoff, time.Second)
	assertDuration(t, "WorkerRetryMaxBackoff", cfg.WorkerRetryMaxBackoff, 5*time.Minute)
	assertDuration(t, "JobLeaseDuration", cfg.JobLeaseDuration, 30*time.Second)
	if cfg.SourceBatchLimit != 100 {
		t.Fatalf("source batch limit = %d, want 100", cfg.SourceBatchLimit)
	}
}

func TestLoadReadsExplicitEnvironment(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(mapEnvironment(map[string]string{
		"NOTIFY_DATABASE_URL":             "postgresql://notify:secret@database:5432/notifications",
		"NOTIFY_MONOLITH_BASE_URL":        "https://backend.internal",
		"NOTIFY_SERVICE_TOKEN":            "notify-service-token",
		"PORT":                            "9090",
		"NOTIFY_LOG_LEVEL":                "debug",
		"NOTIFY_HTTP_READ_TIMEOUT":        "2s",
		"NOTIFY_HTTP_READ_HEADER_TIMEOUT": "3s",
		"NOTIFY_HTTP_WRITE_TIMEOUT":       "4s",
		"NOTIFY_HTTP_IDLE_TIMEOUT":        "5s",
		"NOTIFY_HTTP_SHUTDOWN_TIMEOUT":    "6s",
		"NOTIFY_DATABASE_CONNECT_TIMEOUT": "7s",
		"NOTIFY_DATABASE_READY_TIMEOUT":   "8s",
		"NOTIFY_MONOLITH_HTTP_TIMEOUT":    "9s",
		"NOTIFY_SOURCE_POLL_INTERVAL":     "10s",
		"NOTIFY_SOURCE_RETRY_MIN_BACKOFF": "11s",
		"NOTIFY_SOURCE_RETRY_MAX_BACKOFF": "12s",
		"NOTIFY_WORKER_SCAN_INTERVAL":     "13s",
		"NOTIFY_WORKER_RETRY_MIN_BACKOFF": "14s",
		"NOTIFY_WORKER_RETRY_MAX_BACKOFF": "15s",
		"NOTIFY_JOB_LEASE_DURATION":       "16s",
		"NOTIFY_SOURCE_BATCH_LIMIT":       "150",
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
	assertDuration(t, "MonolithHTTPTimeout", cfg.MonolithHTTPTimeout, 9*time.Second)
	assertDuration(t, "SourcePollInterval", cfg.SourcePollInterval, 10*time.Second)
	assertDuration(t, "SourceRetryMinBackoff", cfg.SourceRetryMinBackoff, 11*time.Second)
	assertDuration(t, "SourceRetryMaxBackoff", cfg.SourceRetryMaxBackoff, 12*time.Second)
	assertDuration(t, "WorkerScanInterval", cfg.WorkerScanInterval, 13*time.Second)
	assertDuration(t, "WorkerRetryMinBackoff", cfg.WorkerRetryMinBackoff, 14*time.Second)
	assertDuration(t, "WorkerRetryMaxBackoff", cfg.WorkerRetryMaxBackoff, 15*time.Second)
	assertDuration(t, "JobLeaseDuration", cfg.JobLeaseDuration, 16*time.Second)
	if cfg.SourceBatchLimit != 150 {
		t.Errorf("SourceBatchLimit = %d, ожидалось 150", cfg.SourceBatchLimit)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  map[string]string
	}{
		{name: "URL базы данных отсутствует", env: map[string]string{"NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken()}},
		{name: "URL монолита отсутствует", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken()}},
		{name: "internal token отсутствует", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL()}},
		{name: "неподдерживаемый URL базы данных", env: map[string]string{"NOTIFY_DATABASE_URL": "mysql://notify:secret@database/notifications", "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken()}},
		{name: "неподдерживаемый URL монолита", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": "postgres://database/internal", "NOTIFY_SERVICE_TOKEN": validInternalToken()}},
		{name: "некорректный порт", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "PORT": "70000"}},
		{name: "некорректный уровень логирования", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_LOG_LEVEL": "verbose"}},
		{name: "некорректная длительность", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_HTTP_READ_TIMEOUT": "soon"}},
		{name: "нулевая длительность", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_HTTP_SHUTDOWN_TIMEOUT": "0s"}},
		{name: "отрицательная длительность", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_DATABASE_CONNECT_TIMEOUT": "-1s"}},
		{name: "нулевой таймаут readiness", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_DATABASE_READY_TIMEOUT": "0s"}},
		{name: "отрицательный таймаут readiness", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_DATABASE_READY_TIMEOUT": "-1s"}},
		{name: "нулевой source batch limit", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_SOURCE_BATCH_LIMIT": "0"}},
		{name: "source retry minimum exceeds maximum", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_SOURCE_RETRY_MIN_BACKOFF": "2m", "NOTIFY_SOURCE_RETRY_MAX_BACKOFF": "1m"}},
		{name: "worker retry minimum exceeds maximum", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_WORKER_RETRY_MIN_BACKOFF": "2m", "NOTIFY_WORKER_RETRY_MAX_BACKOFF": "1m"}},
		{name: "job lease does not cover monolith request", env: map[string]string{"NOTIFY_DATABASE_URL": validDatabaseURL(), "NOTIFY_MONOLITH_BASE_URL": validMonolithURL(), "NOTIFY_SERVICE_TOKEN": validInternalToken(), "NOTIFY_MONOLITH_HTTP_TIMEOUT": "10s", "NOTIFY_JOB_LEASE_DURATION": "10s"}},
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
		"NOTIFY_DATABASE_URL":      validDatabaseURL(),
		"NOTIFY_MONOLITH_BASE_URL": validMonolithURL(),
		"NOTIFY_SERVICE_TOKEN":     validInternalToken(),
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
		"NOTIFY_DATABASE_URL":      validDatabaseURL(),
		"NOTIFY_MONOLITH_BASE_URL": validMonolithURL(),
		"NOTIFY_SERVICE_TOKEN":     validInternalToken(),
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
		"NOTIFY_DATABASE_URL":      "postgres://notify:" + password + "@database:5432/notifications",
		"NOTIFY_MONOLITH_BASE_URL": validMonolithURL(),
		"NOTIFY_SERVICE_TOKEN":     validInternalToken(),
		"NOTIFY_LOG_LEVEL":         "invalid",
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

func validMonolithURL() string {
	return "http://backend.internal:8080"
}

func validInternalToken() string {
	return "notify-service-internal-token"
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
