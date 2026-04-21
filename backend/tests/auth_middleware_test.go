package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/middlewares"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtUtils := utils.NewJWTUtils("test-secret")
	middleware := middlewares.NewAuthMiddleware(jwtUtils)

	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{name: "missing header", wantStatus: http.StatusUnauthorized},
		{name: "bad format", header: "bad-token", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer invalid-token", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
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
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthMiddlewareSetsClaimsForValidAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtUtils := utils.NewJWTUtils("test-secret")
	tokenPair, err := jwtUtils.GenerateTokenPair(7, "Alice", "alice", "avatar.png")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	router := gin.New()
	router.GET("/protected", middlewares.NewAuthMiddleware(jwtUtils).RequireAuth(), func(c *gin.Context) {
		if got := c.GetUint("userID"); got != 7 {
			t.Fatalf("userID = %d, want 7", got)
		}
		if got := c.GetString("username"); got != "Alice" {
			t.Fatalf("username = %q, want Alice", got)
		}
		if got := c.GetString("us"); got != "alice" {
			t.Fatalf("us = %q, want alice", got)
		}
		if got := c.GetString("image"); got != "avatar.png" {
			t.Fatalf("image = %q, want avatar.png", got)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestAuthMiddlewareRejectsRefreshTokenAsAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtUtils := utils.NewJWTUtils("test-secret")
	tokenPair, err := jwtUtils.GenerateTokenPair(7, "Alice", "alice", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	router := gin.New()
	router.GET("/protected", middlewares.NewAuthMiddleware(jwtUtils).RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.RefreshToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
