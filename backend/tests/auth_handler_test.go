package tests

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	"friendship/models/dto"

	"github.com/gin-gonic/gin"
)

type authServiceStub struct {
	loginResponse   dto.AuthResponse
	loginErr        error
	refreshResponse dto.AuthResponse
	refreshErr      error
}

func (s *authServiceStub) Login(email, password string) (dto.AuthResponse, error) {
	if s.loginErr != nil {
		return dto.AuthResponse{}, s.loginErr
	}
	return s.loginResponse, nil
}

func (s *authServiceStub) RefreshTokens(refreshToken string) (dto.AuthResponse, error) {
	if s.refreshErr != nil {
		return dto.AuthResponse{}, s.refreshErr
	}
	return s.refreshResponse, nil
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := handlers.NewAuthHandler(&authServiceStub{
		loginResponse: dto.AuthResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			AdminGroups:  []dto.AdminGroupResponse{},
		},
	})

	rec := performAuthRequest(handler.Login, `{"email":"user@example.com","password":"secret"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"access_token":"access-token"`) {
		t.Fatalf("response does not contain access token: %s", rec.Body.String())
	}
}

func TestAuthHandlerLoginServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := handlers.NewAuthHandler(&authServiceStub{loginErr: errors.New("bad credentials")})
	rec := performAuthRequest(handler.Login, `{"email":"user@example.com","password":"secret"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertCommonErrorShape(t, rec)
}

func TestAuthHandlerRefreshSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := handlers.NewAuthHandler(&authServiceStub{
		refreshResponse: dto.AuthResponse{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			AdminGroups:  []dto.AdminGroupResponse{},
		},
	})

	rec := performAuthRequest(handler.RefreshToken, `{"refresh_token":"refresh-token"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"refresh_token":"new-refresh"`) {
		t.Fatalf("response does not contain refresh token: %s", rec.Body.String())
	}
}

func TestAuthHandlerRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := handlers.NewAuthHandler(&authServiceStub{})
	rec := performAuthRequest(handler.Login, `{"email":"not-email","password":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	assertCommonErrorShape(t, rec)
}

func performAuthRequest(handler gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler(ctx)

	return rec
}
