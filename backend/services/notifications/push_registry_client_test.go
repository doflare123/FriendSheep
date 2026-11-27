package notifications

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPPushRegistryRejectsUnexpectedUpstreamResponses(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"повтор версии внутри details", 409, `{"error":"push_version_conflict","details":{"currentVersion":null,"currentVersion":"d0a00000-0000-4000-8000-000000000001"}}`},
		{"пропущена версия", 409, `{"error":"push_version_conflict","details":{}}`},
		{"неожиданный код", 409, `{"error":"unexpected_integration_detail"}`},
		{"неполный ответ", 200, `{"schemaVersion":1,"state":"active"}`},
		{"слишком большой ответ", 200, strings.Repeat("x", pushResponseBodyLimit+1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = io.WriteString(w, test.body)
			}))
			defer server.Close()
			client, err := NewHTTPPushDeviceRegistryClient(server.URL, "test-internal", time.Second, "project", "test")
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Register(context.Background(), validPushRegistryInput())
			if !errors.Is(err, ErrPushRegistryUnavailable) {
				t.Fatalf("ожидалась безопасная ошибка недоступности, получено %v", err)
			}
		})
	}
}

func TestHTTPPushRegistryDoesNotFollowRedirects(t *testing.T) {
	var redirected atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirected.Store(true)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client, err := NewHTTPPushDeviceRegistryClient(server.URL, "test-internal", time.Second, "project", "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Register(context.Background(), validPushRegistryInput())
	if !errors.Is(err, ErrPushRegistryUnavailable) || redirected.Load() {
		t.Fatal("запрос пошёл по редиректу или клиент некорректно обработал ответ")
	}
}

func validPushRegistryInput() RegisterPushDeviceInput {
	return RegisterPushDeviceInput{
		InstallationID: "ni5BJjBEB8flsp3sh_95VIcYA52LdZHWeupCksV0y-I", InstallationSecret: "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8",
		SchemaVersion: 1, UserID: 7, SessionID: "AAECAwQFBgcICQoLDA0ODxAR", OperationID: "c0a00000-0000-4000-8000-000000000001",
		Platform: "android", ProviderToken: "test-address", NotificationsEnabled: true,
	}
}
