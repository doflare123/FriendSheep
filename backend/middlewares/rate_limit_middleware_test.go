package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"friendship/models/dto"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type rateLimitStoreCall struct {
	key    string
	limit  int
	window time.Duration
}

type alwaysActiveAuthSessionReader struct{}

func (alwaysActiveAuthSessionReader) HasActiveSession(context.Context, string) (bool, error) {
	return true, nil
}

type fakeRateLimitStore struct {
	mu       sync.Mutex
	calls    []rateLimitStoreCall
	counts   map[string]int
	err      error
	decision func(key string, limit int, count int) RateLimitDecision
}

func newFakeRateLimitStore() *fakeRateLimitStore {
	return &fakeRateLimitStore{counts: make(map[string]int)}
}

func (s *fakeRateLimitStore) Take(_ context.Context, key string, limit int, window time.Duration) (RateLimitDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, rateLimitStoreCall{
		key:    key,
		limit:  limit,
		window: window,
	})
	if s.err != nil {
		return RateLimitDecision{}, s.err
	}

	s.counts[key]++
	count := s.counts[key]
	if s.decision != nil {
		return s.decision(key, limit, count), nil
	}

	return RateLimitDecision{
		Allowed:    count <= limit,
		Count:      count,
		ResetAfter: 2 * time.Second,
	}, nil
}

func (s *fakeRateLimitStore) snapshotCalls() []rateLimitStoreCall {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]rateLimitStoreCall(nil), s.calls...)
}

type rateLimitTestLogger struct{}

func (rateLimitTestLogger) Info(string, ...interface{})  {}
func (rateLimitTestLogger) Error(string, ...interface{}) {}
func (rateLimitTestLogger) Debug(string, ...interface{}) {}
func (rateLimitTestLogger) Warn(string, ...interface{})  {}
func (rateLimitTestLogger) Fatal(string, ...interface{}) {}
func (rateLimitTestLogger) Panic(string, ...interface{}) {}

func TestRateLimitMiddlewareAllowsWithinLimitAndReturnsCommon429(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	store.decision = func(key string, _ int, count int) RateLimitDecision {
		return RateLimitDecision{
			Allowed:    !strings.Contains(key, ":references-read-ip:") || count == 1,
			Count:      count,
			ResetAfter: 1500 * time.Millisecond,
		}
	}
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	router.GET("/api/v2/references", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	first := performRateLimitRequest(router, http.MethodGet, "/api/v2/references", nil, "192.0.2.10:1000", "")
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d; body=%s", first.Code, http.StatusNoContent, first.Body.String())
	}

	second := performRateLimitRequest(router, http.MethodGet, "/api/v2/references", nil, "192.0.2.10:1001", "")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d; body=%s", second.Code, http.StatusTooManyRequests, second.Body.String())
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(second.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal 429 response: %v; body=%s", err, second.Body.String())
	}
	if response.Error == "" || response.Message == "" {
		t.Fatalf("429 response is not dto.ErrorResponse-compatible: %#v", response)
	}

	calls := store.snapshotCalls()
	if len(calls) != 4 {
		t.Fatalf("store calls = %d, want 4 (global and references policy per request): %#v", len(calls), calls)
	}
	assertRateLimitHeaders(t, first, calls[1].limit, false)
	assertRateLimitHeaders(t, second, calls[3].limit, true)
}

func TestRateLimitMiddlewareSkipsNonAPIV2Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	store.decision = func(_ string, _ int, count int) RateLimitDecision {
		return RateLimitDecision{Allowed: false, Count: count, ResetAfter: time.Minute}
	}
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	router.GET("/docs", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := performRateLimitRequest(router, http.MethodGet, "/docs", nil, "192.0.2.11:1000", "")
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
	if calls := store.snapshotCalls(); len(calls) != 0 {
		t.Fatalf("non-API request consumed rate limit: %#v", calls)
	}
}

func TestRateLimitIdentifierPoliciesRestoreJSONBodyAndIsolateIdentifiers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	router.POST("/api/v2/auth/login", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			t.Errorf("downstream ShouldBindJSON failed after limiter inspected body: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}
		if input.Password != "secret" {
			t.Errorf("downstream password = %q, want secret", input.Password)
		}
		c.Status(http.StatusNoContent)
	})

	emails := []string{"first@example.com", "second@example.com"}
	var keys []string
	for i, email := range emails {
		body := []byte(fmt.Sprintf(`{"email":%q,"password":"secret"}`, email))
		recorder := performRateLimitRequest(router, http.MethodPost, "/api/v2/auth/login", body, "192.0.2.12:1000", "application/json")
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d; body=%s", i, recorder.Code, http.StatusNoContent, recorder.Body.String())
		}

		calls := store.snapshotCalls()
		wantCalls := (i + 1) * 3
		if len(calls) != wantCalls {
			t.Fatalf("after request %d store calls = %d, want %d: %#v", i, len(calls), wantCalls, calls)
		}
		keys = append(keys, calls[len(calls)-1].key)
	}

	if keys[0] == keys[1] {
		t.Fatalf("different login identifiers share key %q", keys[0])
	}
	if bytes.Contains([]byte(keys[0]), []byte(emails[0])) || bytes.Contains([]byte(keys[1]), []byte(emails[1])) {
		t.Fatalf("rate-limit keys expose raw identifiers: %#v", keys)
	}
}

func TestRateLimitIdentifierPolicyRejectsOversizedValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	handlerCalled := false
	router.POST("/api/v2/auth/login", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})

	body := append(
		[]byte(`{"email":"user@example.com","password":"secret"}`),
		bytes.Repeat([]byte(" "), maxRateLimitBodySize)...,
	)
	recorder := performRateLimitRequest(
		router,
		http.MethodPost,
		"/api/v2/auth/login",
		body,
		"192.0.2.18:1000",
		"application/json",
	)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"status = %d, want %d; body=%s",
			recorder.Code,
			http.StatusRequestEntityTooLarge,
			recorder.Body.String(),
		)
	}
	if handlerCalled {
		t.Fatal("oversized request reached login handler")
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Error != "request_too_large" {
		t.Fatalf("error = %q, want request_too_large", response.Error)
	}
}

func TestRateLimitStrictIdentifierAllowsExactlyConfiguredLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	var handlerCalls atomic.Int32
	router.POST("/api/v2/auth/login", func(c *gin.Context) {
		handlerCalls.Add(1)
		c.Status(http.StatusNoContent)
	})

	body := []byte(`{"email":"limited@example.com","password":"secret"}`)
	const emailPolicyLimit = 10
	for i := 0; i < emailPolicyLimit; i++ {
		recorder := performRateLimitRequest(router, http.MethodPost, "/api/v2/auth/login", body, "192.0.2.17:1000", "application/json")
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d; body=%s", i+1, recorder.Code, http.StatusNoContent, recorder.Body.String())
		}
	}

	limited := performRateLimitRequest(router, http.MethodPost, "/api/v2/auth/login", body, "192.0.2.17:1000", "application/json")
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("request after limit status = %d, want %d; body=%s", limited.Code, http.StatusTooManyRequests, limited.Body.String())
	}
	if got := handlerCalls.Load(); got != emailPolicyLimit {
		t.Fatalf("handler calls = %d, want %d", got, emailPolicyLimit)
	}

	calls := store.snapshotCalls()
	last := calls[len(calls)-1]
	if !strings.Contains(last.key, ":auth-login-email:") {
		t.Fatalf("limiting policy key = %q, want auth-login-email", last.key)
	}
	if last.limit != emailPolicyLimit {
		t.Fatalf("identifier policy limit = %d, want %d", last.limit, emailPolicyLimit)
	}
	assertRateLimitHeaders(t, limited, emailPolicyLimit, true)
}

func TestRateLimitIPKeysIgnoreSourcePortAndIsolateAddresses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	router.GET("/api/v2/other", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, remoteAddr := range []string{
		"192.0.2.21:1000",
		"192.0.2.21:2000",
		"192.0.2.22:1000",
	} {
		recorder := performRateLimitRequest(router, http.MethodGet, "/api/v2/other", nil, remoteAddr, "")
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("remote %s status = %d, want %d", remoteAddr, recorder.Code, http.StatusNoContent)
		}
	}

	calls := store.snapshotCalls()
	if len(calls) != 3 {
		t.Fatalf("store calls = %d, want 3: %#v", len(calls), calls)
	}
	if calls[0].key != calls[1].key {
		t.Fatalf("same IP with different source ports produced different keys: %q != %q", calls[0].key, calls[1].key)
	}
	if calls[1].key == calls[2].key {
		t.Fatalf("different IPs share key %q", calls[1].key)
	}
}

func TestRateLimitStoreErrorsUsePolicyFailureMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		path       string
		body       []byte
		wantStatus int
	}{
		{
			name:       "ordinary API read fails open",
			method:     http.MethodGet,
			path:       "/api/v2/references",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "login fails closed",
			method:     http.MethodPost,
			path:       "/api/v2/auth/login",
			body:       []byte(`{"email":"user@example.com","password":"secret"}`),
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "registration verification fails closed",
			method:     http.MethodPatch,
			path:       "/api/v2/register/session/verify",
			body:       []byte(`{"session_id":"session-1","type":"register","code":"123456"}`),
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeRateLimitStore()
			store.err = errors.New("rate limit store unavailable")
			router := gin.New()
			router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			recorder := performRateLimitRequest(router, tt.method, tt.path, tt.body, "192.0.2.13:1000", "application/json")
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if tt.wantStatus != http.StatusNoContent {
				var response dto.ErrorResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("unmarshal fail-closed response: %v; body=%s", err, recorder.Body.String())
				}
				if response.Error == "" || response.Message == "" {
					t.Fatalf("fail-closed response is not dto.ErrorResponse-compatible: %#v", response)
				}
			}
		})
	}
}

func TestAuthMiddlewareAppliesLimitOnlyAfterSuccessfulRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	store.decision = func(_ string, _ int, count int) RateLimitDecision {
		return RateLimitDecision{Allowed: false, Count: count, ResetAfter: time.Second}
	}
	rateLimiter := NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret")
	jwtUtils := utils.NewJWTUtils("test-secret")
	auth := NewAuthMiddleware(jwtUtils, alwaysActiveAuthSessionReader{})
	auth.SetRateLimiter(rateLimiter)

	router := gin.New()
	router.GET("/protected", auth.RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	invalid := performRateLimitRequest(router, http.MethodGet, "/protected", nil, "192.0.2.14:1000", "application/json")
	if invalid.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d; body=%s", invalid.Code, http.StatusUnauthorized, invalid.Body.String())
	}
	if calls := store.snapshotCalls(); len(calls) != 0 {
		t.Fatalf("unauthenticated request consumed user limit: %#v", calls)
	}

	accessToken, _, _, err := jwtUtils.GenerateAccessToken(42, "session-42")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	authenticated := performRateLimitRequest(router, http.MethodGet, "/protected", nil, "192.0.2.14:1001", "application/json", accessToken)
	if authenticated.Code != http.StatusTooManyRequests {
		t.Fatalf("authenticated status = %d, want %d; body=%s", authenticated.Code, http.StatusTooManyRequests, authenticated.Body.String())
	}
	if calls := store.snapshotCalls(); len(calls) != 1 {
		t.Fatalf("authenticated store calls = %d, want 1: %#v", len(calls), calls)
	}
}

func TestAuthMiddlewareOptionalAuthUsesPerUserKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	rateLimiter := NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret")
	jwtUtils := utils.NewJWTUtils("test-secret")
	auth := NewAuthMiddleware(jwtUtils, alwaysActiveAuthSessionReader{})
	auth.SetRateLimiter(rateLimiter)

	router := gin.New()
	router.GET("/optional", auth.OptionalAuth(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	anonymous := performRateLimitRequest(router, http.MethodGet, "/optional", nil, "192.0.2.15:1000", "")
	if anonymous.Code != http.StatusNoContent {
		t.Fatalf("anonymous status = %d, want %d", anonymous.Code, http.StatusNoContent)
	}
	if calls := store.snapshotCalls(); len(calls) != 0 {
		t.Fatalf("anonymous optional-auth request consumed user limit: %#v", calls)
	}

	var keys []string
	for _, userID := range []uint{51, 52} {
		accessToken, _, _, err := jwtUtils.GenerateAccessToken(userID, "session-"+strconv.FormatUint(uint64(userID), 10))
		if err != nil {
			t.Fatalf("GenerateAccessToken(%d): %v", userID, err)
		}
		recorder := performRateLimitRequest(router, http.MethodGet, "/optional", nil, "192.0.2.15:1000", "", accessToken)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("user %d status = %d, want %d; body=%s", userID, recorder.Code, http.StatusNoContent, recorder.Body.String())
		}
		calls := store.snapshotCalls()
		keys = append(keys, calls[len(calls)-1].key)
	}

	if keys[0] == keys[1] {
		t.Fatalf("different authenticated users share key %q", keys[0])
	}
}

func TestAuthenticatedRouteUsesNormalizedFullPathPolicyAfterRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeRateLimitStore()
	store.decision = func(key string, _ int, count int) RateLimitDecision {
		return RateLimitDecision{
			Allowed:    !strings.Contains(key, ":membership-change-user:"),
			Count:      count,
			ResetAfter: time.Minute,
		}
	}
	rateLimiter := NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret")
	jwtUtils := utils.NewJWTUtils("test-secret")
	auth := NewAuthMiddleware(jwtUtils, alwaysActiveAuthSessionReader{})
	auth.SetRateLimiter(rateLimiter)

	router := gin.New()
	router.Use(rateLimiter.APIBaseLimit())
	var handlerCalled atomic.Bool
	router.POST("/api/v2/events/:eventId/join", auth.RequireAuth(), func(c *gin.Context) {
		handlerCalled.Store(true)
		c.Status(http.StatusNoContent)
	})

	accessToken, _, _, err := jwtUtils.GenerateAccessToken(61, "session-61")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	recorder := performRateLimitRequest(
		router,
		http.MethodPost,
		"/api/v2/events/17/join",
		nil,
		"192.0.2.23:1000",
		"",
		accessToken,
	)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusTooManyRequests, recorder.Body.String())
	}
	if handlerCalled.Load() {
		t.Fatal("limited request reached handler")
	}

	calls := store.snapshotCalls()
	if len(calls) != 3 {
		t.Fatalf("store calls = %d, want global, authenticated write, and membership policies: %#v", len(calls), calls)
	}
	wantPolicyNames := []string{"api-global-ip", "authenticated-write-user", "membership-change-user"}
	for i, wantName := range wantPolicyNames {
		if !strings.Contains(calls[i].key, ":"+wantName+":") {
			t.Errorf("call[%d] key = %q, want policy %q", i, calls[i].key, wantName)
		}
	}
}

func TestRateLimitMiddlewareConcurrentTakeAllowsExactlyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const requestCount = 40
	const allowedLimit = 7
	store := newFakeRateLimitStore()
	store.decision = func(_ string, _ int, count int) RateLimitDecision {
		return RateLimitDecision{
			Allowed:    count <= allowedLimit,
			Count:      count,
			ResetAfter: time.Second,
		}
	}
	router := gin.New()
	router.Use(NewRateLimitMiddleware(rateLimitTestLogger{}, store, "test-hash-secret").APIBaseLimit())
	router.GET("/api/v2/other", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	var allowed atomic.Int32
	var denied atomic.Int32
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(requestCount)
	for i := 0; i < requestCount; i++ {
		go func() {
			defer wg.Done()
			<-start
			recorder := performRateLimitRequest(router, http.MethodGet, "/api/v2/other", nil, "192.0.2.16:1000", "")
			switch recorder.Code {
			case http.StatusNoContent:
				allowed.Add(1)
			case http.StatusTooManyRequests:
				denied.Add(1)
			default:
				t.Errorf("unexpected status = %d; body=%s", recorder.Code, recorder.Body.String())
			}
		}()
	}
	close(start)
	wg.Wait()

	calls := store.snapshotCalls()
	if len(calls) != requestCount {
		t.Fatalf("store calls = %d, want %d", len(calls), requestCount)
	}
	if got := int(allowed.Load()); got != allowedLimit {
		t.Fatalf("allowed = %d, want %d (denied=%d)", got, allowedLimit, denied.Load())
	}
	if got := int(denied.Load()); got != requestCount-allowedLimit {
		t.Fatalf("denied = %d, want %d", got, requestCount-allowedLimit)
	}
}

func performRateLimitRequest(
	router http.Handler,
	method string,
	path string,
	body []byte,
	remoteAddr string,
	contentType string,
	accessToken ...string,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.RemoteAddr = remoteAddr
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if len(accessToken) > 0 && accessToken[0] != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken[0])
	}
	router.ServeHTTP(recorder, request)
	return recorder
}

func assertRateLimitHeaders(t *testing.T, recorder *httptest.ResponseRecorder, wantLimit int, wantRetryAfter bool) {
	t.Helper()

	if got := recorder.Header().Get("X-RateLimit-Limit"); got != strconv.Itoa(wantLimit) {
		t.Fatalf("X-RateLimit-Limit = %q, want %q", got, strconv.Itoa(wantLimit))
	}
	if got := recorder.Header().Get("X-RateLimit-Remaining"); got == "" {
		t.Fatal("X-RateLimit-Remaining is empty")
	}
	if got := recorder.Header().Get("X-RateLimit-Reset"); got == "" {
		t.Fatal("X-RateLimit-Reset is empty")
	}
	if retryAfter := recorder.Header().Get("Retry-After"); wantRetryAfter && retryAfter == "" {
		t.Fatal("Retry-After is empty for 429 response")
	} else if !wantRetryAfter && retryAfter != "" {
		t.Fatalf("Retry-After = %q for allowed response, want empty", retryAfter)
	}
}
