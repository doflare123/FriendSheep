package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/middlewares"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type authSessionReaderStub struct {
	active bool
	err    error
	calls  []string
}

func (s *authSessionReaderStub) HasActiveSession(_ context.Context, sessionID string) (bool, error) {
	s.calls = append(s.calls, sessionID)
	return s.active, s.err
}

func TestAuthMiddlewareRequireAuthRejectsInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		header     string
		sessions   *authSessionReaderStub
		wantStatus int
		wantCalls  int
	}{
		{name: "missing header", wantStatus: http.StatusUnauthorized},
		{name: "bad format", header: "bad-token", wantStatus: http.StatusUnauthorized},
		{name: "empty bearer", header: "Bearer ", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer invalid-token", wantStatus: http.StatusUnauthorized},
		{
			name:       "revoked session",
			header:     "Bearer " + mustAccessToken(t, 7, "revoked-session"),
			sessions:   &authSessionReaderStub{active: false},
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
		{
			name:       "session check unavailable",
			header:     "Bearer " + mustAccessToken(t, 7, "session-7"),
			sessions:   &authSessionReaderStub{err: errors.New("store down")},
			wantStatus: http.StatusServiceUnavailable,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			middleware := middlewares.NewAuthMiddleware(newConfiguredJWTUtils(), tt.sessions)
			router.GET("/protected", middleware.RequireAuth(), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			assertCommonErrorShape(t, rec)
			if tt.sessions != nil && len(tt.sessions.calls) != tt.wantCalls {
				t.Fatalf("HasActiveSession calls = %d, want %d", len(tt.sessions.calls), tt.wantCalls)
			}
		})
	}
}

func TestAuthMiddlewareSetsTrustedAccessIdentityAndSessionContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sessions := &authSessionReaderStub{active: true}
	token := mustAccessToken(t, 7, "session-7")
	router := gin.New()
	router.GET("/protected", middlewares.NewAuthMiddleware(newConfiguredJWTUtils(), sessions).RequireAuth(), func(c *gin.Context) {
		if got := c.GetUint("userID"); got != 7 {
			t.Fatalf("userID = %d, want 7", got)
		}
		if got := c.GetString("authSessionID"); got != "session-7" {
			t.Fatalf("authSessionID = %q, want session-7", got)
		}
		if got := c.GetString("authTokenID"); got == "" {
			t.Fatal("authTokenID is empty")
		}
		for _, legacyKey := range []string{"username", "us", "image"} {
			if _, exists := c.Get(legacyKey); exists {
				t.Fatalf("legacy profile context key %q was set", legacyKey)
			}
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if len(sessions.calls) != 1 || sessions.calls[0] != "session-7" {
		t.Fatalf("HasActiveSession calls = %#v, want session-7", sessions.calls)
	}
}

func TestAuthMiddlewareOptionalAuthKeepsMissingCredentialsAnonymousButRejectsRevokedSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing credentials", func(t *testing.T) {
		router := gin.New()
		router.GET("/optional", middlewares.NewAuthMiddleware(newConfiguredJWTUtils()).OptionalAuth(), func(c *gin.Context) {
			if _, exists := c.Get("userID"); exists {
				t.Fatal("anonymous request unexpectedly has userID")
			}
			c.Status(http.StatusNoContent)
		})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/optional", nil))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("revoked credentials", func(t *testing.T) {
		router := gin.New()
		sessions := &authSessionReaderStub{active: false}
		router.GET("/optional", middlewares.NewAuthMiddleware(newConfiguredJWTUtils(), sessions).OptionalAuth(), func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodGet, "/optional", nil)
		req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, 8, "revoked-session"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		assertCommonErrorShape(t, rec)
	})
}

func mustAccessToken(t *testing.T, userID uint, sessionID string, jwtServices ...*utils.JWTUtils) string {
	t.Helper()
	jwtService := newConfiguredJWTUtils()
	if len(jwtServices) > 0 && jwtServices[0] != nil {
		jwtService = jwtServices[0]
	}
	token, _, _, err := jwtService.GenerateAccessToken(userID, sessionID)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	return token
}

var _ middlewares.AuthSessionReader = (*authSessionReaderStub)(nil)
