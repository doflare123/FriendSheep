package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/models/dto"
	"friendship/routes"
	"friendship/services"

	"github.com/gin-gonic/gin"
)

type authServiceStub struct {
	loginResponse    dto.AuthResponse
	loginErr         error
	refreshResponse  dto.AuthResponse
	refreshErr       error
	meResponse       dto.AuthMeResponse
	meErr            error
	revokeCurrentErr error
	revokeAllErr     error

	loginEmail         string
	loginPassword      string
	refreshToken       string
	meUserID           uint
	revokeSessionID    string
	revokeAllUserID    uint
	loginContextSeen   bool
	refreshContextSeen bool
}

func (s *authServiceStub) Login(ctx context.Context, email, password string) (dto.AuthResponse, error) {
	s.loginContextSeen = ctx != nil
	s.loginEmail = email
	s.loginPassword = password
	return s.loginResponse, s.loginErr
}

func (s *authServiceStub) RefreshTokens(ctx context.Context, refreshToken string) (dto.AuthResponse, error) {
	s.refreshContextSeen = ctx != nil
	s.refreshToken = refreshToken
	return s.refreshResponse, s.refreshErr
}

func (s *authServiceStub) IssueTokens(context.Context, uint) (dto.AuthResponse, error) {
	return dto.AuthResponse{}, nil
}

func (s *authServiceStub) GetMe(_ context.Context, userID uint) (dto.AuthMeResponse, error) {
	s.meUserID = userID
	return s.meResponse, s.meErr
}

func (s *authServiceStub) RevokeCurrent(_ context.Context, sessionID string) error {
	s.revokeSessionID = sessionID
	return s.revokeCurrentErr
}

func (s *authServiceStub) RevokeAll(_ context.Context, userID uint) error {
	s.revokeAllUserID = userID
	return s.revokeAllErr
}

func TestAuthHandlerLoginAndRefreshReturnTokensWithMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		path     string
		body     string
		response dto.AuthResponse
		assert   func(*testing.T, *authServiceStub)
	}{
		{
			name: "login",
			path: "/api/v2/auth/login",
			body: `{"email":"user@example.com","password":"secret"}`,
			response: dto.AuthResponse{
				AccessToken:  "access-token",
				RefreshToken: "session.opaque-secret",
				Me:           dto.AuthMeResponse{ID: 7, Name: "User", Us: "user", Image: "avatar.png"},
				AdminGroups:  []dto.AdminGroupResponse{},
			},
			assert: func(t *testing.T, stub *authServiceStub) {
				if !stub.loginContextSeen || stub.loginEmail != "user@example.com" || stub.loginPassword != "secret" {
					t.Fatalf("login call = ctx:%v email:%q password:%q", stub.loginContextSeen, stub.loginEmail, stub.loginPassword)
				}
			},
		},
		{
			name: "refresh",
			path: "/api/v2/auth/refresh",
			body: `{"refresh_token":"session.old-secret"}`,
			response: dto.AuthResponse{
				AccessToken:  "new-access",
				RefreshToken: "session.new-secret",
				Me:           dto.AuthMeResponse{ID: 7, Name: "Updated User", Us: "updated", Image: "new.png"},
				AdminGroups:  []dto.AdminGroupResponse{},
			},
			assert: func(t *testing.T, stub *authServiceStub) {
				if !stub.refreshContextSeen || stub.refreshToken != "session.old-secret" {
					t.Fatalf("refresh call = ctx:%v token:%q", stub.refreshContextSeen, stub.refreshToken)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &authServiceStub{loginResponse: tt.response, refreshResponse: tt.response}
			router := newAuthHandlerRouter(stub, nil)
			rec := performJSONRouteRequest(router, http.MethodPost, tt.path, tt.body, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var response dto.AuthResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.AccessToken != tt.response.AccessToken || response.RefreshToken != tt.response.RefreshToken || response.Me != tt.response.Me {
				t.Fatalf("response = %#v, want %#v", response, tt.response)
			}
			tt.assert(t, stub)
		})
	}
}

func TestAuthHandlerMapsCredentialRefreshAndValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		body       string
		stub       *authServiceStub
		wantStatus int
	}{
		{
			name: "invalid login request", path: "/api/v2/auth/login",
			body: `{"email":"not-email","password":""}`, stub: &authServiceStub{}, wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad credentials", path: "/api/v2/auth/login",
			body: `{"email":"user@example.com","password":"wrong"}`, stub: &authServiceStub{loginErr: services.ErrInvalidCredentials}, wantStatus: http.StatusUnauthorized,
		},
		{
			name: "login repository unavailable", path: "/api/v2/auth/login",
			body: `{"email":"user@example.com","password":"secret"}`, stub: &authServiceStub{loginErr: errors.New("database unavailable")}, wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "missing refresh", path: "/api/v2/auth/refresh",
			body: `{}`, stub: &authServiceStub{}, wantStatus: http.StatusBadRequest,
		},
		{
			name: "replayed refresh", path: "/api/v2/auth/refresh",
			body: `{"refresh_token":"session.replayed"}`, stub: &authServiceStub{refreshErr: services.ErrRefreshTokenReplay}, wantStatus: http.StatusUnauthorized,
		},
		{
			name: "refresh store unavailable", path: "/api/v2/auth/refresh",
			body: `{"refresh_token":"session.secret"}`, stub: &authServiceStub{refreshErr: errors.New("store down")}, wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performJSONRouteRequest(newAuthHandlerRouter(tt.stub, nil), http.MethodPost, tt.path, tt.body, "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			assertCommonErrorShape(t, rec)
		})
	}
}

func TestAuthHandlerMapsMeLookupErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "missing user", err: services.ErrAuthUserNotFound, wantStatus: http.StatusNotFound},
		{name: "repository unavailable", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &authServiceStub{meErr: tt.err}
			router := newAuthHandlerRouter(stub, &authSessionReaderStub{active: true})
			rec := performJSONRouteRequest(router, http.MethodGet, "/api/v2/user/me", "", mustAccessToken(t, 27, "session-27"))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			assertCommonErrorShape(t, rec)
		})
	}
}

func TestAuthRoutesMeAndLogoutEndpointsUseAuthenticatedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		assert     func(*testing.T, *authServiceStub)
	}{
		{
			name: "me", method: http.MethodGet, path: "/api/v2/user/me", wantStatus: http.StatusOK,
			assert: func(t *testing.T, stub *authServiceStub) {
				if stub.meUserID != 27 {
					t.Fatalf("GetMe userID = %d, want 27", stub.meUserID)
				}
			},
		},
		{
			name: "logout current", method: http.MethodPost, path: "/api/v2/auth/logout", wantStatus: http.StatusNoContent,
			assert: func(t *testing.T, stub *authServiceStub) {
				if stub.revokeSessionID != "session-27" {
					t.Fatalf("RevokeCurrent sessionID = %q, want session-27", stub.revokeSessionID)
				}
			},
		},
		{
			name: "logout all", method: http.MethodPost, path: "/api/v2/auth/logout-all", wantStatus: http.StatusNoContent,
			assert: func(t *testing.T, stub *authServiceStub) {
				if stub.revokeAllUserID != 27 {
					t.Fatalf("RevokeAll userID = %d, want 27", stub.revokeAllUserID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &authServiceStub{meResponse: dto.AuthMeResponse{ID: 27, Name: "Me", Us: "me", Image: "me.png"}}
			sessions := &authSessionReaderStub{active: true}
			router := newAuthHandlerRouter(stub, sessions)
			rec := performJSONRouteRequest(router, tt.method, tt.path, "", mustAccessToken(t, 27, "session-27"))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == http.StatusNoContent && rec.Body.Len() != 0 {
				t.Fatalf("204 response body = %q, want empty", rec.Body.String())
			}
			if tt.path == "/api/v2/user/me" {
				var me dto.AuthMeResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
					t.Fatalf("decode me response: %v", err)
				}
				if me != stub.meResponse {
					t.Fatalf("me = %#v, want %#v", me, stub.meResponse)
				}
			}
			tt.assert(t, stub)
		})
	}
}

func TestAuthRoutesProtectedEndpointsRejectMissingAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAuthHandlerRouter(&authServiceStub{}, &authSessionReaderStub{active: true})

	for _, request := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/user/me"},
		{http.MethodPost, "/api/v2/auth/logout"},
		{http.MethodPost, "/api/v2/auth/logout-all"},
	} {
		rec := performJSONRouteRequest(router, request.method, request.path, "", "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want %d", request.method, request.path, rec.Code, http.StatusUnauthorized)
		}
		assertCommonErrorShape(t, rec)
	}
}

func newAuthHandlerRouter(service services.AuthService, sessions middlewares.AuthSessionReader) *gin.Engine {
	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(newConfiguredJWTUtils(), sessions)
	routes.RegisterAuthRoutes(router, handlers.NewAuthHandler(service), authMiddleware)
	return router
}

func performJSONRouteRequest(router *gin.Engine, method, path, body, accessToken string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	router.ServeHTTP(rec, req)
	return rec
}

var _ services.AuthService = (*authServiceStub)(nil)
