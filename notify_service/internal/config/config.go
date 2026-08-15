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
)

type Getenv func(string) string

// Config содержит только конфигурацию процесса. Секреты читаются из окружения
// и никогда не включаются в ошибки валидации или журналы.
type Config struct {
	DatabaseURL            string
	Port                   string
	LogLevel               slog.Level
	HTTPReadTimeout        time.Duration
	HTTPReadHeaderTimeout  time.Duration
	HTTPWriteTimeout       time.Duration
	HTTPIdleTimeout        time.Duration
	HTTPShutdownTimeout    time.Duration
	DatabaseConnectTimeout time.Duration
	DatabaseReadyTimeout   time.Duration
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
		Port:                   valueOrDefault(getenv("PORT"), defaultPort),
		LogLevel:               slog.LevelInfo,
		HTTPReadTimeout:        defaultHTTPReadTimeout,
		HTTPReadHeaderTimeout:  defaultHTTPReadHeaderTimeout,
		HTTPWriteTimeout:       defaultHTTPWriteTimeout,
		HTTPIdleTimeout:        defaultHTTPIdleTimeout,
		HTTPShutdownTimeout:    defaultHTTPShutdownTimeout,
		DatabaseConnectTimeout: defaultDatabaseConnectTimeout,
		DatabaseReadyTimeout:   defaultDatabaseReadyTimeout,
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

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("требуется NOTIFY_DATABASE_URL")
	}
	parsed, err := url.Parse(c.DatabaseURL)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || parsed.Path == "" || parsed.Path == "/" {
		return errors.New("NOTIFY_DATABASE_URL должен быть корректным PostgreSQL URL с именем базы данных")
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
	} {
		if value <= 0 {
			return fmt.Errorf("%s должен быть положительным", name)
		}
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
