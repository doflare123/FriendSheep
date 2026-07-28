package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"friendship/config"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddlewareUsesConfiguredPolicyQuotas(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimits := config.DefaultRateLimitConfig()
	rateLimits.APIGlobalIPLimit = 11
	rateLimits.APIGlobalIPWindow = 2 * time.Minute
	rateLimits.ReferencesReadIPLimit = 7
	rateLimits.ReferencesReadIPWindow = 45 * time.Second

	store := newFakeRateLimitStore()
	limiter := NewRateLimitMiddlewareWithConfig(
		rateLimitTestLogger{},
		store,
		"test-hash-secret",
		rateLimits,
	)
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		t.Fatalf("SetTrustedProxies(nil): %v", err)
	}
	router.Use(limiter.APIBaseLimit())
	router.GET("/api/v2/references", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := performRateLimitRequest(
		router,
		http.MethodGet,
		"/api/v2/references",
		nil,
		"192.0.2.70:1234",
		"",
	)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}

	calls := store.snapshotCalls()
	if got, want := len(calls), 2; got != want {
		t.Fatalf("store calls = %d, want %d: %#v", got, want, calls)
	}
	assertRateLimitStoreCallQuota(t, calls[0], 11, 2*time.Minute)
	assertRateLimitStoreCallQuota(t, calls[1], 7, 45*time.Second)
	if got, want := recorder.Header().Get("X-RateLimit-Limit"), "7"; got != want {
		t.Fatalf("X-RateLimit-Limit = %q, want %q from configured route policy", got, want)
	}
}

func TestRateLimitMiddlewareOnlyAcceptsForwardedIPFromTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		trustedProxies []string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		wantClientIP   string
	}{
		{
			name:           "untrusted peer headers ignored",
			trustedProxies: []string{"10.0.0.0/8"},
			remoteAddr:     "192.0.2.71:1234",
			forwardedFor:   "203.0.113.80",
			realIP:         "203.0.113.81",
			wantClientIP:   "192.0.2.71",
		},
		{
			name:           "configured trusted proxy forwards client",
			trustedProxies: []string{"10.0.0.0/8"},
			remoteAddr:     "10.1.2.3:1234",
			forwardedFor:   "203.0.113.80",
			realIP:         "203.0.113.81",
			wantClientIP:   "203.0.113.80",
		},
		{
			name:           "no trusted proxies means headers ignored",
			trustedProxies: nil,
			remoteAddr:     "10.1.2.3:1234",
			forwardedFor:   "203.0.113.80",
			realIP:         "203.0.113.81",
			wantClientIP:   "10.1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeRateLimitStore()
			limiter := NewRateLimitMiddlewareWithConfig(
				rateLimitTestLogger{},
				store,
				"test-hash-secret",
				config.DefaultRateLimitConfig(),
			)
			router := gin.New()
			if err := router.SetTrustedProxies(tt.trustedProxies); err != nil {
				t.Fatalf("SetTrustedProxies(%#v): %v", tt.trustedProxies, err)
			}
			router.Use(limiter.APIBaseLimit())
			router.GET("/api/v2/other", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/api/v2/other", nil)
			request.RemoteAddr = tt.remoteAddr
			request.Header.Set("X-Forwarded-For", tt.forwardedFor)
			request.Header.Set("X-Real-IP", tt.realIP)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
			}
			calls := store.snapshotCalls()
			if got, want := len(calls), 1; got != want {
				t.Fatalf("store calls = %d, want %d: %#v", got, want, calls)
			}
			wantKey := limiter.rateLimitKey("api-global-ip", "ip:"+tt.wantClientIP)
			if calls[0].key != wantKey {
				t.Fatalf(
					"rate-limit key = %q, want key for Gin client IP %q (%q)",
					calls[0].key,
					tt.wantClientIP,
					wantKey,
				)
			}
		})
	}
}

func assertRateLimitStoreCallQuota(
	t *testing.T,
	call rateLimitStoreCall,
	wantLimit int,
	wantWindow time.Duration,
) {
	t.Helper()

	if call.limit != wantLimit || call.window != wantWindow {
		t.Fatalf(
			"store quota = {limit:%d window:%s}, want {limit:%d window:%s}",
			call.limit,
			call.window,
			wantLimit,
			wantWindow,
		)
	}
}
