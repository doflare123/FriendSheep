package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notify_service/internal/httpserver"
	"notify_service/internal/pushdevices"
)

type pushDevicesStub struct {
	registerInput   pushdevices.RegisterInput
	detachInput     pushdevices.DetachInput
	registerResult  pushdevices.MutationResult
	detachResult    pushdevices.MutationResult
	registerErr     error
	detachErr       error
	registerContext context.Context
	detachContext   context.Context
}

func (stub *pushDevicesStub) Register(ctx context.Context, input pushdevices.RegisterInput) (pushdevices.MutationResult, error) {
	stub.registerContext = ctx
	stub.registerInput = input
	return stub.registerResult, stub.registerErr
}

func (stub *pushDevicesStub) Detach(ctx context.Context, input pushdevices.DetachInput) (pushdevices.MutationResult, error) {
	stub.detachContext = ctx
	stub.detachInput = input
	return stub.detachResult, stub.detachErr
}

func TestPushDeviceRoutesRequireInternalTokenBeforeRegistryAvailability(t *testing.T) {
	t.Parallel()

	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", nil)
	for _, headers := range []http.Header{
		{},
		{"X-Internal-Token": {"wrong-token"}},
		{"X-Internal-Token": {"expected-token", "expected-token"}},
		{"Authorization": {"Bearer user-access-token"}},
	} {
		request := httptest.NewRequest(http.MethodPut, "/internal/v1/push-devices/example", strings.NewReader(`{}`))
		request.Header = headers
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		assertPushError(t, response, http.StatusUnauthorized, "invalid_internal_token")
	}
}

func TestPushDeviceRoutesReportUnavailableRegistryAfterAuthentication(t *testing.T) {
	t.Parallel()

	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", nil)
	request := httptest.NewRequest(http.MethodPut, "/internal/v1/push-devices/example", strings.NewReader(`{}`))
	request.Header.Set("X-Internal-Token", "expected-token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assertPushError(t, response, http.StatusServiceUnavailable, "push_registry_unavailable")
	if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "no-store" {
		t.Fatalf("Cache-Control = %q, ожидалось no-store", cacheControl)
	}
}

func TestPushDeviceRegisterMapsTrustedInternalPayloadAndResponse(t *testing.T) {
	t.Parallel()

	stub := &pushDevicesStub{registerResult: pushdevices.MutationResult{
		SchemaVersion:  pushdevices.SchemaVersion,
		InstallationID: "installation-id",
		Version:        "f0000000-0000-4000-8000-000000000001",
		State:          pushdevices.StateActive,
	}}
	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", stub)
	body := `{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null,"platform":"android","provider":"fcm","providerProject":"friendship-test","environment":"test","providerToken":"provider-token","notificationsEnabled":true}`
	request := httptest.NewRequest(http.MethodPut, "/internal/v1/push-devices/installation-id", strings.NewReader(body))
	request.Header.Set("X-Internal-Token", "expected-token")
	request.Header.Set("X-Installation-Secret", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("HTTP-статус = %d, ожидался 200, тело=%s", response.Code, response.Body.String())
	}
	input := stub.registerInput
	if input.InstallationID != "installation-id" || input.InstallationSecret != "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" || input.UserID != 7 || input.SessionID != "AAECAwQFBgcICQoLDA0ODxAR" {
		t.Fatalf("неверная идентичность RegisterInput: %#v", input)
	}
	if input.SchemaVersion != 1 || input.OperationID != "c0000000-0000-4000-8000-000000000001" || input.ExpectedVersion != nil || input.Platform != "android" || input.Provider != "fcm" || input.ProviderProject != "friendship-test" || input.Environment != "test" || input.ProviderToken != "provider-token" || !input.NotificationsEnabled {
		t.Fatalf("неверный RegisterInput: %#v", input)
	}
	var actual pushdevices.MutationResult
	if err := json.Unmarshal(response.Body.Bytes(), &actual); err != nil {
		t.Fatalf("не удалось декодировать успешный ответ: %v", err)
	}
	if actual != stub.registerResult {
		t.Fatalf("ответ = %#v, ожидался %#v", actual, stub.registerResult)
	}
}

func TestPushDeviceRoutesRejectAmbiguousOrUnsafeInputBeforeRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		headers map[string][]string
		status  int
		code    string
	}{
		{name: "неизвестное поле", method: http.MethodPut, path: "/internal/v1/push-devices/id", body: `{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null,"platform":"android","provider":"fcm","providerProject":"project","environment":"test","providerToken":"token","notificationsEnabled":true,"ignored":true}`, headers: map[string][]string{"X-Installation-Secret": {"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}}, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "повторённое JSON поле", method: http.MethodPost, path: "/internal/v1/push-devices/id/detach", body: `{"schemaVersion":1,"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null}`, headers: map[string][]string{"X-Installation-Secret": {"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}}, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "повторённый секрет", method: http.MethodPost, path: "/internal/v1/push-devices/id/detach", body: `{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null}`, headers: map[string][]string{"X-Installation-Secret": {"one", "two"}}, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "query запрещён", method: http.MethodPost, path: "/internal/v1/push-devices/id/detach?userId=99", body: `{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null}`, headers: map[string][]string{"X-Installation-Secret": {"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}}, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "неверный content type", method: http.MethodPost, path: "/internal/v1/push-devices/id/detach", body: `{}`, headers: map[string][]string{"X-Installation-Secret": {"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}, "Content-Type": {"text/plain"}}, status: http.StatusUnsupportedMediaType, code: "unsupported_media_type"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			stub := &pushDevicesStub{}
			handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", stub)
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			request.Header.Set("X-Internal-Token", "expected-token")
			if _, found := test.headers["Content-Type"]; !found {
				request.Header.Set("Content-Type", "application/json")
			}
			for name, values := range test.headers {
				request.Header.Del(name)
				for _, value := range values {
					request.Header.Add(name, value)
				}
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			assertPushError(t, response, test.status, test.code)
			if stub.registerContext != nil || stub.detachContext != nil {
				t.Fatal("реестр вызван до успешной валидации HTTP-запроса")
			}
		})
	}
}

func TestPushDeviceRouteRejectsOversizedBodyBeforeRegistry(t *testing.T) {
	t.Parallel()

	stub := &pushDevicesStub{}
	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", stub)
	request := httptest.NewRequest(http.MethodPut, "/internal/v1/push-devices/id", strings.NewReader(strings.Repeat("x", 16_385)))
	request.Header.Set("X-Internal-Token", "expected-token")
	request.Header.Set("X-Installation-Secret", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assertPushError(t, response, http.StatusRequestEntityTooLarge, "request_too_large")
	if stub.registerContext != nil || stub.detachContext != nil {
		t.Fatal("реестр вызван для тела больше лимита")
	}
}

func TestPushDeviceDetachMapsConflictAndPreservesRequestContext(t *testing.T) {
	t.Parallel()

	version := "d0000000-0000-4000-8000-000000000001"
	stub := &pushDevicesStub{detachErr: &pushdevices.DomainError{Code: pushdevices.ErrorCodePushVersionConflict, CurrentVersion: &version}}
	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", stub)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/push-devices/id/detach", strings.NewReader(`{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":"d0000000-0000-4000-8000-000000000000"}`)).WithContext(ctx)
	request.Header.Set("X-Internal-Token", "expected-token")
	request.Header.Set("X-Installation-Secret", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if !errors.Is(stub.detachContext.Err(), context.Canceled) {
		t.Fatalf("Detach не получил отменённый контекст: %v", stub.detachContext.Err())
	}
	if stub.detachInput.ExpectedVersion == nil || *stub.detachInput.ExpectedVersion != "d0000000-0000-4000-8000-000000000000" {
		t.Fatalf("Detach получил неверную expectedVersion: %#v", stub.detachInput.ExpectedVersion)
	}
	if response.Code != http.StatusConflict {
		t.Fatalf("HTTP-статус = %d, ожидался 409; тело=%s", response.Code, response.Body.String())
	}
	var body struct {
		Error   string `json:"error"`
		Details struct {
			CurrentVersion *string `json:"currentVersion"`
		} `json:"details"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось декодировать конфликт: %v", err)
	}
	if body.Error != pushdevices.ErrorCodePushVersionConflict || body.Details.CurrentVersion == nil || *body.Details.CurrentVersion != version {
		t.Fatalf("тело конфликта = %#v", body)
	}
}

func TestPushDeviceRegisterMapsInvalidatedTokenConflict(t *testing.T) {
	t.Parallel()

	stub := &pushDevicesStub{registerErr: &pushdevices.DomainError{Code: pushdevices.ErrorCodePushTokenInvalidated}}
	handler := httpserver.NewHandler(discardLogger(), readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", stub)
	request := httptest.NewRequest(http.MethodPut, "/internal/v1/push-devices/id", strings.NewReader(`{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":"d0000000-0000-4000-8000-000000000001","platform":"android","provider":"fcm","providerProject":"project","environment":"test","providerToken":"rejected-token","notificationsEnabled":true}`))
	request.Header.Set("X-Internal-Token", "expected-token")
	request.Header.Set("X-Installation-Secret", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assertPushError(t, response, http.StatusConflict, pushdevices.ErrorCodePushTokenInvalidated)
}

func TestPushDeviceRouteLogsDoNotRevealSecretOrProviderToken(t *testing.T) {
	t.Parallel()

	const secret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	const providerToken = "do-not-log-provider-token"
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	stub := &pushDevicesStub{registerResult: pushdevices.MutationResult{SchemaVersion: 1, InstallationID: "id", Version: "d0000000-0000-4000-8000-000000000001", State: pushdevices.StateActive}}
	handler := httpserver.NewHandler(logger, readinessFunc(func(context.Context) error { return nil }), nil, "expected-token", stub)
	send := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPut, "/internal/v1/push-devices/id", strings.NewReader(`{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR","operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null,"platform":"android","provider":"fcm","providerProject":"project","environment":"test","providerToken":"`+providerToken+`","notificationsEnabled":true}`))
		request.Header.Set("X-Internal-Token", "expected-token")
		request.Header.Set("X-Installation-Secret", secret)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	response := send()

	if response.Code != http.StatusOK {
		t.Fatalf("HTTP-статус = %d, ожидался 200; тело=%s", response.Code, response.Body.String())
	}
	stub.registerErr = errors.New("ошибка хранилища: " + secret + " " + providerToken)
	failed := send()
	assertPushError(t, failed, http.StatusServiceUnavailable, "push_registry_unavailable")
	for _, value := range []string{secret, providerToken} {
		if strings.Contains(response.Body.String(), value) || strings.Contains(failed.Body.String(), value) || strings.Contains(logs.String(), value) {
			t.Fatal("ответ или HTTP-журнал раскрыл секрет реестра")
		}
	}
}

func assertPushError(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if response.Code != wantStatus {
		t.Fatalf("HTTP-статус = %d, ожидался %d; тело=%s", response.Code, wantStatus, response.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось декодировать ошибку: %v", err)
	}
	if body.Error != wantCode {
		t.Fatalf("error = %q, ожидался %q", body.Error, wantCode)
	}
}
