package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	"friendship/routes"
	"friendship/services/notifications"

	"github.com/gin-gonic/gin"
)

type pushRegistryStub struct {
	register notifications.RegisterPushDeviceInput
	detach   notifications.DetachPushDeviceInput
	err      error
}

func (s *pushRegistryStub) Register(ctx context.Context, input notifications.RegisterPushDeviceInput) (notifications.PushDeviceMutationResult, error) {
	if ctx == nil {
		return notifications.PushDeviceMutationResult{}, errors.New("контекст отсутствует")
	}
	s.register = input
	if s.err != nil {
		return notifications.PushDeviceMutationResult{}, s.err
	}
	return notifications.PushDeviceMutationResult{SchemaVersion: 1, InstallationID: input.InstallationID, Version: "c0000000-0000-4000-8000-000000000001", State: notifications.PushDeviceStateActive}, nil
}
func (s *pushRegistryStub) Detach(_ context.Context, input notifications.DetachPushDeviceInput) (notifications.PushDeviceMutationResult, error) {
	s.detach = input
	if s.err != nil {
		return notifications.PushDeviceMutationResult{}, s.err
	}
	return notifications.PushDeviceMutationResult{SchemaVersion: 1, InstallationID: input.InstallationID, Version: "c0000000-0000-4000-8000-000000000002", State: notifications.PushDeviceStateDetached}, nil
}

func TestPushDevicesHandlerUsesTrustedSessionIdentityAndMapsRegistryFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const installationID = "ni5BJjBEB8flsp3sh_95VIcYA52LdZHWeupCksV0y-I"
	const secret = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"
	const sessionID = "AAECAwQFBgcICQoLDA0ODxAR"

	t.Run("регистрация получает только аутентифицированную пару", func(t *testing.T) {
		registry := &pushRegistryStub{}
		router := pushDeviceRouter(registry, 7, sessionID)
		body := `{"schemaVersion":1,"operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null,"platform":"android","providerToken":"token-A","notificationsEnabled":true}`
		request := httptest.NewRequest(http.MethodPut, "/api/v2/users/me/push-devices/"+installationID+"?userId=99", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Installation-Secret", secret)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || registry.register.UserID != 0 {
			t.Fatalf("query/status/register=%d/%#v", response.Code, registry.register)
		}

		request = httptest.NewRequest(http.MethodPut, "/api/v2/users/me/push-devices/"+installationID, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Installation-Secret", secret)
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK || registry.register.UserID != 7 || registry.register.SessionID != sessionID {
			t.Fatalf("status/input=%d/%#v", response.Code, registry.register)
		}
		var result notifications.PushDeviceMutationResult
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.State != notifications.PushDeviceStateActive {
			t.Fatalf("result/error=%#v/%v", result, err)
		}
	})

	t.Run("недоступный реестр не превращается в успех", func(t *testing.T) {
		registry := &pushRegistryStub{err: errors.New("notify down")}
		router := pushDeviceRouter(registry, 7, sessionID)
		request := httptest.NewRequest(http.MethodPut, "/api/v2/users/me/push-devices/"+installationID, strings.NewReader(`{"schemaVersion":1,"operationId":"c0000000-0000-4000-8000-000000000001","expectedVersion":null,"platform":"android","providerToken":"token-A","notificationsEnabled":true}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Installation-Secret", secret)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "push_registry_unavailable") {
			t.Fatalf("status/body=%d/%s", response.Code, response.Body.String())
		}
	})
}

func pushDeviceRouter(registry notifications.PushDeviceRegistry, userID uint, sessionID string) *gin.Engine {
	router := gin.New()
	routes.RegisterPushDeviceRoutes(router, handlers.NewPushDevicesHandler(registry), func(c *gin.Context) { c.Set("userID", userID); c.Set("authSessionID", sessionID); c.Next() })
	return router
}
