package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/models/dto"
	"friendship/repository"
	"friendship/routes"
	"friendship/services"
	"friendship/services/notifications"
	"friendship/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type pushBlackBoxRedis struct {
	repository.RedisRepository
	client *redis.Client
}

func (r pushBlackBoxRedis) Client() *redis.Client { return r.client }

func TestPushDevicesCrossServiceSessionRevocation(t *testing.T) {
	backendDSN := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	notifyDSN := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	redisURL := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_REDIS_URL"))
	if backendDSN == "" || notifyDSN == "" || redisURL == "" {
		t.Skip("для межсервисной проверки push нужны две временные PostgreSQL и Redis")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatal("некорректный тестовый Redis URL")
	}
	redisClient := redis.NewClient(redisOptions)
	t.Cleanup(func() { _ = redisClient.Close() })
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatal("тестовая Redis недоступна")
	}
	prefix := fmt.Sprintf("push-blackbox:%d:", time.Now().UnixNano())
	sessionConfig := services.AuthSessionStoreConfig{RefreshTTL: time.Minute, SessionKeyPrefix: prefix + "session:", UserSessionsPrefix: prefix + "owner:"}
	sessions, err := services.NewRedisAuthSessionStore(pushBlackBoxRedis{client: redisClient}, sessionConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, id := range []uint{1, 2} {
			_ = sessions.RevokeAllUserSessions(context.Background(), id)
		}
	})
	backendDB := newPostgresEventCommandDB(t, backendDSN)
	seedEventUser(t, backendDB, 1)
	seedEventUser(t, backendDB, 2)
	jwt := utils.NewJWTUtils("push-blackbox-jwt-key-at-least-32-bytes")
	auth := services.NewAuthService(&testLogger{}, jwt, &testPostgresRepository{db: backendDB}, sessions)
	authMiddleware := middlewares.NewAuthMiddleware(jwt, sessions)
	const internalToken = "push-blackbox-service-token"
	gin.SetMode(gin.TestMode)
	internal := gin.New()
	routes.RegisterInternalPushSessionRoutes(internal, handlers.NewPushSessionHandler(sessions), middlewares.NewInternalTokenMiddleware(internalToken))
	monolith := httptest.NewServer(internal)
	t.Cleanup(monolith.Close)

	newLogin := func(id uint) dto.AuthResponse {
		t.Helper()
		login, err := auth.IssueTokens(ctx, id)
		if err != nil {
			t.Fatalf("issue tokens: %v", err)
		}
		return login
	}
	sessionID := func(login dto.AuthResponse) string {
		t.Helper()
		claims, err := jwt.ParseAccessToken(login.AccessToken)
		if err != nil {
			t.Fatal("не удалось прочитать тестовую auth-сессию")
		}
		return claims.SessionID
	}
	public := func(notifyURL string) *gin.Engine {
		t.Helper()
		client, err := notifications.NewHTTPPushDeviceRegistryClient(notifyURL, internalToken, time.Second, "friendsheep-test", "test")
		if err != nil {
			t.Fatal(err)
		}
		router := gin.New()
		routes.RegisterAuthRoutes(router, handlers.NewAuthHandler(auth), authMiddleware)
		routes.RegisterPushDeviceRoutes(router, handlers.NewPushDevicesHandler(client), authMiddleware.RequireAuth())
		return router
	}
	pushEnv := []string{"NOTIFY_PUSH_ALLOWED_SCOPES=friendsheep-test:test", "NOTIFY_PUSH_ENCRYPTION_KEY_ID=blackbox-1"}
	for _, name := range []string{"NOTIFY_PUSH_ENCRYPTION_KEY", "NOTIFY_PUSH_TOKEN_FINGERPRINT_KEY", "NOTIFY_PUSH_OPERATION_FINGERPRINT_KEY"} {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			t.Fatal(err)
		}
		pushEnv = append(pushEnv, name+"="+base64.RawURLEncoding.EncodeToString(key))
	}

	notifyURL, stopFirst := startNotifyServiceWithStateForBlackBox(t, notifyDSN, monolith.URL, internalToken, true, pushEnv...)
	stoppedFirst := false
	t.Cleanup(func() {
		if !stoppedFirst {
			stopFirst()
		}
	})
	waitForBlackBoxReadiness(t, notifyURL)
	router := public(notifyURL)
	phoneA, secondA, phoneB := newLogin(1), newLogin(1), newLogin(2)
	firstDevice, secondDevice, switchedDevice := newPushBlackBoxInstallation(t), newPushBlackBoxInstallation(t), newPushBlackBoxInstallation(t)
	first := requirePushBlackBoxMutation(t, firstDevice.put(router, phoneA.AccessToken, 1, nil, "android"), http.StatusOK)
	second := requirePushBlackBoxMutation(t, secondDevice.put(router, secondA.AccessToken, 2, nil, "ios"), http.StatusOK)
	if replay := requirePushBlackBoxMutation(t, firstDevice.put(router, phoneA.AccessToken, 1, nil, "android"), http.StatusOK); replay != first {
		t.Fatal("повтор регистрации изменил результат")
	}

	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": secondA.RefreshToken})
	refresh := pushBlackBoxRequest(router, http.MethodPost, "/api/v2/auth/refresh", "", refreshBody, nil)
	var refreshed dto.AuthResponse
	if refresh.Code != http.StatusOK || json.Unmarshal(refresh.Body.Bytes(), &refreshed) != nil || sessionID(refreshed) != sessionID(secondA) {
		t.Fatalf("refresh должен сохранить auth-сессию, код %d", refresh.Code)
	}
	if replay := requirePushBlackBoxMutation(t, secondDevice.put(router, refreshed.AccessToken, 2, nil, "ios"), http.StatusOK); replay != second {
		t.Fatal("refresh изменил привязку устройства")
	}
	explicitDevice := newPushBlackBoxInstallation(t)
	explicit := requirePushBlackBoxMutation(t, explicitDevice.put(router, refreshed.AccessToken, 3, nil, "android"), http.StatusOK)
	detached := requirePushBlackBoxMutation(t, explicitDevice.detach(router, refreshed.AccessToken, 4, explicit.Version), http.StatusOK)
	if replay := requirePushBlackBoxMutation(t, explicitDevice.detach(router, refreshed.AccessToken, 4, explicit.Version), http.StatusOK); replay != detached || detached.State != "detached" {
		t.Fatal("повтор явной отвязки изменил результат")
	}

	if response := pushBlackBoxRequest(router, http.MethodPost, "/api/v2/auth/logout", phoneA.AccessToken, nil, nil); response.Code != http.StatusNoContent {
		t.Fatalf("logout вернул %d", response.Code)
	}
	assertPushBlackBoxSession(t, monolith.URL, internalToken, 1, sessionID(phoneA), false)
	assertPushBlackBoxSession(t, monolith.URL, internalToken, 1, sessionID(refreshed), true)

	switched := requirePushBlackBoxMutation(t, switchedDevice.put(router, refreshed.AccessToken, 5, nil, "android"), http.StatusOK)
	withoutProof := switchedDevice
	withoutProof.secret = ""
	assertPushBlackBoxError(t, withoutProof.put(router, phoneB.AccessToken, 6, &switched.Version, "android"), 404, "push_device_not_found")
	rebound := requirePushBlackBoxMutation(t, switchedDevice.put(router, phoneB.AccessToken, 6, &switched.Version, "android"), http.StatusOK)
	assertPushBlackBoxError(t, switchedDevice.put(router, refreshed.AccessToken, 7, &switched.Version, "android"), 409, "push_version_conflict")
	assertPushBlackBoxError(t, switchedDevice.detach(router, refreshed.AccessToken, 8, rebound.Version), 404, "push_device_not_found")

	expired := newLogin(2)
	expiredDevice := newPushBlackBoxInstallation(t)
	requirePushBlackBoxMutation(t, expiredDevice.put(router, expired.AccessToken, 9, nil, "ios"), http.StatusOK)
	if err := redisClient.PExpire(ctx, sessionConfig.SessionKeyPrefix+sessionID(expired), time.Millisecond).Err(); err != nil {
		t.Fatal(err)
	}
	for {
		remaining, err := redisClient.Exists(ctx, sessionConfig.SessionKeyPrefix+sessionID(expired)).Result()
		if err != nil {
			t.Fatal(err)
		}
		if remaining == 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	assertPushBlackBoxSession(t, monolith.URL, internalToken, 2, sessionID(expired), false)

	replayed := newLogin(2)
	replayedDevice := newPushBlackBoxInstallation(t)
	requirePushBlackBoxMutation(t, replayedDevice.put(router, replayed.AccessToken, 11, nil, "android"), http.StatusOK)
	replayedBody, _ := json.Marshal(map[string]string{"refresh_token": replayed.RefreshToken})
	if response := pushBlackBoxRequest(router, http.MethodPost, "/api/v2/auth/refresh", "", replayedBody, nil); response.Code != http.StatusOK {
		t.Fatalf("первая ротация refresh вернула %d", response.Code)
	}
	assertPushBlackBoxError(t, pushBlackBoxRequest(router, http.MethodPost, "/api/v2/auth/refresh", "", replayedBody, nil), 401, "invalid_refresh_token")
	assertPushBlackBoxSession(t, monolith.URL, internalToken, 2, sessionID(replayed), false)

	stopFirst()
	stoppedFirst = true
	assertPushBlackBoxError(t, secondDevice.put(router, refreshed.AccessToken, 10, &second.Version, "ios"), 503, "push_registry_unavailable")
	if response := pushBlackBoxRequest(router, http.MethodPost, "/api/v2/auth/logout-all", refreshed.AccessToken, nil, nil); response.Code != http.StatusNoContent {
		t.Fatalf("logout-all при недоступном notify вернул %d", response.Code)
	}
	assertPushBlackBoxSession(t, monolith.URL, internalToken, 1, sessionID(refreshed), false)
	assertPushBlackBoxSession(t, monolith.URL, internalToken, 2, sessionID(phoneB), true)

	restartedURL, stopRestarted := startNotifyServiceWithStateForBlackBox(t, notifyDSN, monolith.URL, internalToken, false, pushEnv...)
	t.Cleanup(stopRestarted)
	waitForBlackBoxReadiness(t, restartedURL)
	db, err := sql.Open("postgres", notifyDSN)
	if err != nil {
		t.Fatal("не удалось открыть тестовую БД notify")
	}
	t.Cleanup(func() { _ = db.Close() })
	sweepCtx, stopSweep := context.WithTimeout(ctx, 8*time.Second)
	defer stopSweep()
	for _, installation := range []pushBlackBoxInstallation{firstDevice, secondDevice, expiredDevice, replayedDevice} {
		waitForPushRegistryState(t, sweepCtx, db, installation.id, "detached")
	}
	var survivingUser int64
	if err := db.QueryRowContext(ctx, `SELECT user_id FROM notify_service.push_device_installations WHERE installation_id = $1 AND state = 'active'`, switchedDevice.id).Scan(&survivingUser); err != nil || survivingUser != 2 {
		t.Fatalf("перепривязка B не пережила отзыв сессий A: user=%d, err=%v", survivingUser, err)
	}
}

type pushBlackBoxInstallation struct{ id, secret string }

func newPushBlackBoxInstallation(t *testing.T) pushBlackBoxInstallation {
	t.Helper()
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatal(err)
	}
	checksum := sha256.Sum256(append([]byte("friendSheep:push-installation:v1:"), raw...))
	return pushBlackBoxInstallation{base64.RawURLEncoding.EncodeToString(checksum[:]), base64.RawURLEncoding.EncodeToString(raw)}
}

func (d pushBlackBoxInstallation) put(router http.Handler, token string, operation int, version *string, platform string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{"schemaVersion": 1, "operationId": pushBlackBoxOperation(operation), "expectedVersion": version,
		"platform": platform, "providerToken": "test-address-" + d.id, "notificationsEnabled": true})
	return pushBlackBoxRequest(router, http.MethodPut, "/api/v2/users/me/push-devices/"+d.id, token, body, map[string]string{"X-Installation-Secret": d.secret})
}

func (d pushBlackBoxInstallation) detach(router http.Handler, token string, operation int, version string) *httptest.ResponseRecorder {
	return pushBlackBoxRequest(router, http.MethodDelete, "/api/v2/users/me/push-devices/"+d.id, token, nil, map[string]string{
		"X-Installation-Secret": d.secret, "X-Push-Operation-ID": pushBlackBoxOperation(operation), "X-Push-Expected-Version": version})
}

func pushBlackBoxOperation(index int) string {
	return fmt.Sprintf("a0000000-0000-4000-8000-%012x", index)
}

func pushBlackBoxRequest(router http.Handler, method, path, token string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if len(body) != 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func requirePushBlackBoxMutation(t *testing.T, response *httptest.ResponseRecorder, status int) notifications.PushDeviceMutationResult {
	t.Helper()
	var result notifications.PushDeviceMutationResult
	if response.Code != status || json.Unmarshal(response.Body.Bytes(), &result) != nil || result.SchemaVersion != 1 || result.Version == "" {
		t.Fatalf("операция устройства вернула %d (ожидалось %d)", response.Code, status)
	}
	return result
}

func assertPushBlackBoxError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	var result struct{ Error string }
	if response.Code != status || json.Unmarshal(response.Body.Bytes(), &result) != nil || result.Error != code {
		t.Fatalf("код операции %d/%q, ожидалось %d/%q", response.Code, result.Error, status, code)
	}
}

func assertPushBlackBoxSession(t *testing.T, monolithURL, token string, userID int, sessionID string, active bool) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"schemaVersion": 1, "userId": userID, "sessionId": sessionID})
	request, err := http.NewRequest(http.MethodPost, monolithURL+"/internal/v1/push-sessions/check", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Token", token)
	client := http.Client{Timeout: time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal("проверка сессии по HTTP недоступна")
	}
	defer response.Body.Close()
	var result struct{ Active *bool }
	if response.StatusCode != http.StatusOK || json.NewDecoder(response.Body).Decode(&result) != nil || result.Active == nil || *result.Active != active {
		t.Fatalf("проверка сессии вернула статус %d, ожидалось active=%t", response.StatusCode, active)
	}
}

func waitForPushRegistryState(t *testing.T, ctx context.Context, db *sql.DB, installationID, expected string) {
	t.Helper()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		var state string
		if err := db.QueryRowContext(ctx, `SELECT state FROM notify_service.push_device_installations WHERE installation_id = $1`, installationID).Scan(&state); err != nil {
			t.Fatal(err)
		}
		if state == expected {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("после перезапуска ожидалось %q, состояние %q", expected, state)
		case <-ticker.C:
		}
	}
}
