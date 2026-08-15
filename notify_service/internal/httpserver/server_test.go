package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"notify_service/internal/config"
	"notify_service/internal/httpserver"
)

type readinessFunc func(context.Context) error

func (fn readinessFunc) PingContext(ctx context.Context) error {
	return fn(ctx)
}

func TestHealthIsIndependentOfDatabaseReadiness(t *testing.T) {
	t.Parallel()

	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error {
		return errors.New("база данных недоступна")
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	assertJSONStatus(t, response, http.StatusOK, "ok")
}

func TestReadyReflectsDatabasePing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		readiness  httpserver.Readiness
		wantCode   int
		wantStatus string
	}{
		{
			name: "готов",
			readiness: readinessFunc(func(context.Context) error {
				return nil
			}),
			wantCode:   http.StatusOK,
			wantStatus: "ready",
		},
		{
			name: "база данных недоступна",
			readiness: readinessFunc(func(context.Context) error {
				return errors.New("база данных недоступна")
			}),
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: "not_ready",
		},
		{
			name:       "зависимость готовности отсутствует",
			readiness:  nil,
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: "not_ready",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			response := httptest.NewRecorder()
			httpserver.NewHandler(discardLogger(), test.readiness).
				ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
			assertJSONStatus(t, response, test.wantCode, test.wantStatus)
		})
	}
}

func TestRequestLogDoesNotExposeHeadersOrQuerySecrets(t *testing.T) {
	t.Parallel()

	const secret = "unique-test-service-token"
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := httpserver.NewHandler(logger, readinessFunc(func(context.Context) error { return nil }))
	request := httptest.NewRequest(http.MethodGet, "/ready?token="+secret, nil)
	request.Header.Set("Authorization", "Bearer "+secret)
	request.Header.Set("X-Internal-Token", secret)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("HTTP-статус = %d, ожидался %d", response.Code, http.StatusOK)
	}
	if strings.Contains(logs.String(), secret) {
		t.Fatalf("структурированный журнал HTTP-запроса раскрывает секрет: %s", logs.String())
	}
}

func TestReadyStopsBlockedPingAtConfiguredDeadline(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("не удалось открыть тестовый listener: %v", err)
	}

	const readyTimeout = 40 * time.Millisecond
	deadlineSeen := make(chan time.Time, 1)
	readiness := readinessFunc(func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			return errors.New("контекст readiness не содержит deadline")
		}
		deadlineSeen <- deadline
		<-ctx.Done()
		return ctx.Err()
	})
	cfg := validConfig()
	cfg.DatabaseReadyTimeout = readyTimeout
	server := httpserver.New(cfg, discardLogger(), readiness)
	ctx, cancel := context.WithCancel(context.Background())
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(ctx, listener)
	}()
	defer func() {
		cancel()
		select {
		case err := <-serveDone:
			if err != nil {
				t.Errorf("HTTP-сервер завершился с ошибкой: %v", err)
			}
		case <-time.After(time.Second):
			t.Error("HTTP-сервер не завершился после отмены контекста")
		}
	}()

	client := &http.Client{Timeout: time.Second}
	baseURL := "http://" + listener.Addr().String()
	waitForHealthy(t, client, baseURL+"/health")
	started := time.Now()
	response, err := client.Get(baseURL + "/ready")
	if err != nil {
		t.Fatalf("запрос проверки готовности завершился ошибкой: %v", err)
	}
	defer response.Body.Close()
	elapsed := time.Since(started)

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("код проверки готовности = %d, ожидался %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("зависшая проверка готовности выполнялась %s, ожидалось ограничение таймаутом", elapsed)
	}
	select {
	case deadline := <-deadlineSeen:
		remaining := deadline.Sub(started)
		if remaining <= 0 || remaining > 200*time.Millisecond {
			t.Errorf("срок завершения проверки готовности установлен через %s, ожидалось около %s", remaining, readyTimeout)
		}
	default:
		t.Fatal("проверка готовности не получила контекст со сроком завершения")
	}
}

func TestServerServeStopsAfterContextCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("не удалось открыть тестовый listener: %v", err)
	}

	cfg := validConfig()
	server := httpserver.New(cfg, discardLogger(), readinessFunc(func(context.Context) error { return nil }))
	ctx, cancel := context.WithCancel(context.Background())
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(ctx, listener)
	}()

	client := &http.Client{Timeout: time.Second}
	baseURL := "http://" + listener.Addr().String()
	waitForHealthy(t, client, baseURL+"/health")
	cancel()

	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("Serve() после отмены контекста завершился ошибкой: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() не завершился после отмены контекста")
	}

	_, err = client.Get(baseURL + "/health")
	if err == nil {
		t.Fatal("HTTP-сервер продолжает принимать запросы после возврата из Serve()")
	}
}

func assertJSONStatus(t *testing.T, response *httptest.ResponseRecorder, wantCode int, wantStatus string) {
	t.Helper()
	if response.Code != wantCode {
		t.Fatalf("HTTP-статус = %d, ожидался %d; тело=%s", response.Code, wantCode, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, ожидался application/json", contentType)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}
	if body.Status != wantStatus {
		t.Errorf("поле status в теле = %q, ожидалось %q", body.Status, wantStatus)
	}
}

func waitForHealthy(t *testing.T, client *http.Client, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("HTTP-сервер не перешёл в исправное состояние")
}

func validConfig() config.Config {
	return config.Config{
		DatabaseURL:            "postgres://notify:secret@database/notifications",
		Port:                   "8080",
		LogLevel:               slog.LevelInfo,
		HTTPReadTimeout:        time.Second,
		HTTPReadHeaderTimeout:  time.Second,
		HTTPWriteTimeout:       time.Second,
		HTTPIdleTimeout:        time.Second,
		HTTPShutdownTimeout:    time.Second,
		DatabaseConnectTimeout: time.Second,
		DatabaseReadyTimeout:   time.Second,
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
