package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort                   = "8080"
	defaultHTTPReadTimeout        = 5 * time.Second
	defaultHTTPReadHeaderTimeout  = 5 * time.Second
	defaultHTTPWriteTimeout       = 10 * time.Second
	defaultHTTPIdleTimeout        = 60 * time.Second
	defaultHTTPShutdownTimeout    = 10 * time.Second
	defaultDatabaseConnectTimeout = 10 * time.Second
	defaultDatabaseReadyTimeout   = 2 * time.Second
	defaultMonolithHTTPTimeout    = 5 * time.Second
	defaultSourcePollInterval     = 5 * time.Second
	defaultSourceRetryMinBackoff  = 1 * time.Second
	defaultSourceRetryMaxBackoff  = 30 * time.Second
	defaultWorkerScanInterval     = 1 * time.Second
	defaultWorkerRetryMinBackoff  = 1 * time.Second
	defaultWorkerRetryMaxBackoff  = 5 * time.Minute
	defaultJobLeaseDuration       = 30 * time.Second
	defaultSourceBatchLimit       = 100
	maxSourceBatchLimit           = 500
)

type Getenv func(string) string

// Config содержит только конфигурацию процесса. Секреты читаются из окружения
// и никогда не включаются в ошибки валидации или журналы.
type Config struct {
	DatabaseURL            string
	MonolithBaseURL        string
	InternalToken          string
	Port                   string
	LogLevel               slog.Level
	HTTPReadTimeout        time.Duration
	HTTPReadHeaderTimeout  time.Duration
	HTTPWriteTimeout       time.Duration
	HTTPIdleTimeout        time.Duration
	HTTPShutdownTimeout    time.Duration
	DatabaseConnectTimeout time.Duration
	DatabaseReadyTimeout   time.Duration
	MonolithHTTPTimeout    time.Duration
	SourcePollInterval     time.Duration
	SourceRetryMinBackoff  time.Duration
	SourceRetryMaxBackoff  time.Duration
	WorkerScanInterval     time.Duration
	WorkerRetryMinBackoff  time.Duration
	WorkerRetryMaxBackoff  time.Duration
	JobLeaseDuration       time.Duration
	SourceBatchLimit       int
}

func LoadFromEnv() (Config, error) {
	return Load(os.Getenv)
}

func Load(getenv Getenv) (Config, error) {
	if getenv == nil {
		return Config{}, errors.New("требуется источник переменных окружения")
	}

	cfg := Config{
		DatabaseURL:            strings.TrimSpace(getenv("NOTIFY_DATABASE_URL")),
		MonolithBaseURL:        strings.TrimSpace(getenv("NOTIFY_MONOLITH_BASE_URL")),
		InternalToken:          strings.TrimSpace(getenv("NOTIFY_SERVICE_TOKEN")),
		Port:                   valueOrDefault(getenv("PORT"), defaultPort),
		LogLevel:               slog.LevelInfo,
		HTTPReadTimeout:        defaultHTTPReadTimeout,
		HTTPReadHeaderTimeout:  defaultHTTPReadHeaderTimeout,
		HTTPWriteTimeout:       defaultHTTPWriteTimeout,
		HTTPIdleTimeout:        defaultHTTPIdleTimeout,
		HTTPShutdownTimeout:    defaultHTTPShutdownTimeout,
		DatabaseConnectTimeout: defaultDatabaseConnectTimeout,
		DatabaseReadyTimeout:   defaultDatabaseReadyTimeout,
		MonolithHTTPTimeout:    defaultMonolithHTTPTimeout,
		SourcePollInterval:     defaultSourcePollInterval,
		SourceRetryMinBackoff:  defaultSourceRetryMinBackoff,
		SourceRetryMaxBackoff:  defaultSourceRetryMaxBackoff,
		WorkerScanInterval:     defaultWorkerScanInterval,
		WorkerRetryMinBackoff:  defaultWorkerRetryMinBackoff,
		WorkerRetryMaxBackoff:  defaultWorkerRetryMaxBackoff,
		JobLeaseDuration:       defaultJobLeaseDuration,
		SourceBatchLimit:       defaultSourceBatchLimit,
	}

	if raw := strings.TrimSpace(getenv("NOTIFY_LOG_LEVEL")); raw != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(raw)); err != nil {
			return Config{}, errors.New("некорректный NOTIFY_LOG_LEVEL")
		}
	}

	durations := []struct {
		key    string
		target *time.Duration
	}{
		{"NOTIFY_HTTP_READ_TIMEOUT", &cfg.HTTPReadTimeout},
		{"NOTIFY_HTTP_READ_HEADER_TIMEOUT", &cfg.HTTPReadHeaderTimeout},
		{"NOTIFY_HTTP_WRITE_TIMEOUT", &cfg.HTTPWriteTimeout},
		{"NOTIFY_HTTP_IDLE_TIMEOUT", &cfg.HTTPIdleTimeout},
		{"NOTIFY_HTTP_SHUTDOWN_TIMEOUT", &cfg.HTTPShutdownTimeout},
		{"NOTIFY_DATABASE_CONNECT_TIMEOUT", &cfg.DatabaseConnectTimeout},
		{"NOTIFY_DATABASE_READY_TIMEOUT", &cfg.DatabaseReadyTimeout},
		{"NOTIFY_MONOLITH_HTTP_TIMEOUT", &cfg.MonolithHTTPTimeout},
		{"NOTIFY_SOURCE_POLL_INTERVAL", &cfg.SourcePollInterval},
		{"NOTIFY_SOURCE_RETRY_MIN_BACKOFF", &cfg.SourceRetryMinBackoff},
		{"NOTIFY_SOURCE_RETRY_MAX_BACKOFF", &cfg.SourceRetryMaxBackoff},
		{"NOTIFY_WORKER_SCAN_INTERVAL", &cfg.WorkerScanInterval},
		{"NOTIFY_WORKER_RETRY_MIN_BACKOFF", &cfg.WorkerRetryMinBackoff},
		{"NOTIFY_WORKER_RETRY_MAX_BACKOFF", &cfg.WorkerRetryMaxBackoff},
		{"NOTIFY_JOB_LEASE_DURATION", &cfg.JobLeaseDuration},
	}
	for _, item := range durations {
		raw := strings.TrimSpace(getenv(item.key))
		if raw == "" {
			continue
		}
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("некорректный %s: требуется положительная длительность", item.key)
		}
		*item.target = parsed
	}

	ints := []struct {
		key    string
		target *int
	}{
		{"NOTIFY_SOURCE_BATCH_LIMIT", &cfg.SourceBatchLimit},
	}
	for _, item := range ints {
		raw := strings.TrimSpace(getenv(item.key))
		if raw == "" {
			continue
		}
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("некорректный %s: требуется положительное целое число", item.key)
		}
		*item.target = parsed
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("требуется NOTIFY_DATABASE_URL")
	}
	if c.MonolithBaseURL == "" {
		return errors.New("требуется NOTIFY_MONOLITH_BASE_URL")
	}
	if strings.TrimSpace(c.InternalToken) == "" {
		return errors.New("требуется NOTIFY_SERVICE_TOKEN")
	}
	parsed, err := url.Parse(c.DatabaseURL)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || parsed.Path == "" || parsed.Path == "/" {
		return errors.New("NOTIFY_DATABASE_URL должен быть корректным PostgreSQL URL с именем базы данных")
	}
	monolithURL, err := url.Parse(c.MonolithBaseURL)
	if err != nil ||
		(monolithURL.Scheme != "http" && monolithURL.Scheme != "https") ||
		monolithURL.Host == "" ||
		monolithURL.User != nil ||
		(monolithURL.Path != "" && monolithURL.Path != "/") ||
		monolithURL.RawQuery != "" ||
		monolithURL.Fragment != "" {
		return errors.New("NOTIFY_MONOLITH_BASE_URL должен быть HTTP(S) origin без credentials, path, query или fragment")
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("PORT должен быть целым числом от 1 до 65535")
	}

	for name, value := range map[string]time.Duration{
		"NOTIFY_HTTP_READ_TIMEOUT":        c.HTTPReadTimeout,
		"NOTIFY_HTTP_READ_HEADER_TIMEOUT": c.HTTPReadHeaderTimeout,
		"NOTIFY_HTTP_WRITE_TIMEOUT":       c.HTTPWriteTimeout,
		"NOTIFY_HTTP_IDLE_TIMEOUT":        c.HTTPIdleTimeout,
		"NOTIFY_HTTP_SHUTDOWN_TIMEOUT":    c.HTTPShutdownTimeout,
		"NOTIFY_DATABASE_CONNECT_TIMEOUT": c.DatabaseConnectTimeout,
		"NOTIFY_DATABASE_READY_TIMEOUT":   c.DatabaseReadyTimeout,
		"NOTIFY_MONOLITH_HTTP_TIMEOUT":    c.MonolithHTTPTimeout,
		"NOTIFY_SOURCE_POLL_INTERVAL":     c.SourcePollInterval,
		"NOTIFY_SOURCE_RETRY_MIN_BACKOFF": c.SourceRetryMinBackoff,
		"NOTIFY_SOURCE_RETRY_MAX_BACKOFF": c.SourceRetryMaxBackoff,
		"NOTIFY_WORKER_SCAN_INTERVAL":     c.WorkerScanInterval,
		"NOTIFY_WORKER_RETRY_MIN_BACKOFF": c.WorkerRetryMinBackoff,
		"NOTIFY_WORKER_RETRY_MAX_BACKOFF": c.WorkerRetryMaxBackoff,
		"NOTIFY_JOB_LEASE_DURATION":       c.JobLeaseDuration,
	} {
		if value <= 0 {
			return fmt.Errorf("%s должен быть положительным", name)
		}
	}
	if c.SourceRetryMinBackoff > c.SourceRetryMaxBackoff {
		return errors.New("NOTIFY_SOURCE_RETRY_MIN_BACKOFF не должен превышать NOTIFY_SOURCE_RETRY_MAX_BACKOFF")
	}
	if c.WorkerRetryMinBackoff > c.WorkerRetryMaxBackoff {
		return errors.New("NOTIFY_WORKER_RETRY_MIN_BACKOFF не должен превышать NOTIFY_WORKER_RETRY_MAX_BACKOFF")
	}
	if c.JobLeaseDuration <= c.MonolithHTTPTimeout {
		return errors.New("NOTIFY_JOB_LEASE_DURATION должен превышать NOTIFY_MONOLITH_HTTP_TIMEOUT")
	}
	if c.SourceBatchLimit < 1 {
		return errors.New("NOTIFY_SOURCE_BATCH_LIMIT должен быть положительным")
	}
	if c.SourceBatchLimit > maxSourceBatchLimit {
		return fmt.Errorf("NOTIFY_SOURCE_BATCH_LIMIT не должен превышать %d", maxSourceBatchLimit)
	}
	return nil
}

func (c Config) Address() string {
	return ":" + c.Port
}

func valueOrDefault(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}
