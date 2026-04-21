package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/middlewares"
	"friendship/routes"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type authRouteHandlerStub struct {
	loginCalled   bool
	refreshCalled bool
}

func (h *authRouteHandlerStub) Login(c *gin.Context) {
	h.loginCalled = true
	c.Status(http.StatusNoContent)
}

func (h *authRouteHandlerStub) RefreshToken(c *gin.Context) {
	h.refreshCalled = true
	c.Status(http.StatusNoContent)
}

type regRouteHandlerStub struct {
	called string
}

func (h *regRouteHandlerStub) CreateSessionRegister(c *gin.Context) {
	h.called = "create-session"
	c.Status(http.StatusNoContent)
}

func (h *regRouteHandlerStub) VerifySession(c *gin.Context) {
	h.called = "verify-session"
	c.Status(http.StatusNoContent)
}

func (h *regRouteHandlerStub) CreateUser(c *gin.Context) {
	h.called = "create-user"
	c.Status(http.StatusNoContent)
}

func (h *regRouteHandlerStub) ChangePassword(c *gin.Context) {
	h.called = "change-password"
	c.Status(http.StatusNoContent)
}

type subRouteHandlerStub struct {
	called bool
}

func (h *subRouteHandlerStub) ChangePhoto(c *gin.Context) {
	h.called = true
	c.Status(http.StatusNoContent)
}

type eventsRouteHandlerStub struct {
	called string
}

func (h *eventsRouteHandlerStub) CreateEvent(c *gin.Context) {
	h.called = "create-event"
	c.Status(http.StatusNoContent)
}
func (h *eventsRouteHandlerStub) UpdateEvent(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) DeleteEvent(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) GetGroupEvents(c *gin.Context)  { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) GetEventDetails(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) GetEventDetailsForAdmin(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
func (h *eventsRouteHandlerStub) JoinEvent(c *gin.Context)         { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) LeaveEvent(c *gin.Context)        { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) KickUserFromEvent(c *gin.Context) { c.Status(http.StatusNoContent) }

func (h *eventsRouteHandlerStub) GetAllGenres(c *gin.Context) {
	h.called = "genres"
	c.Status(http.StatusNoContent)
}

func (h *eventsRouteHandlerStub) GetAllReferences(c *gin.Context) {
	h.called = "references"
	c.Status(http.StatusNoContent)
}

type popularEventsRouteHandlerStub struct {
	called bool
}

func (h *popularEventsRouteHandlerStub) GetPopularEvents(c *gin.Context) {
	h.called = true
	c.Status(http.StatusNoContent)
}

type groupRouteHandlerStub struct{}

func (h *groupRouteHandlerStub) CreateGroup(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) UpdateGroup(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) DeleteGroup(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) JoinGroup(c *gin.Context)       { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) LeaveGroup(c *gin.Context)      { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) GetGroupDetails(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) ApproveAllJoinRequests(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
func (h *groupRouteHandlerStub) RejectAllJoinRequests(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) ApproveJoinRequest(c *gin.Context)    { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RejectJoinRequest(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) GetJoinRequests(c *gin.Context)       { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) AddPermissions(c *gin.Context)        { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RemovePermissions(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) GetGroupBlacklist(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) DeleteUserFromGroup(c *gin.Context)   { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RemoveFromBlacklist(c *gin.Context)   { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) CreateJoinInvite(c *gin.Context)      { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) AcceptJoinInvite(c *gin.Context)      { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RejectJoinInvite(c *gin.Context)      { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) WatchRecentActions(c *gin.Context)    { c.Status(http.StatusNoContent) }

func TestRegisterAuthRoutesRefreshUsesPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &authRouteHandlerStub{}
	router := gin.New()
	routes.RegisterAuthRoutes(router, handler)

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/auth/refresh", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/v2/auth/refresh status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if handler.refreshCalled {
		t.Fatal("GET /api/v2/auth/refresh unexpectedly reached refresh handler")
	}

	rec = performRouteRequest(router, http.MethodPost, "/api/v2/auth/refresh", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/v2/auth/refresh status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !handler.refreshCalled {
		t.Fatal("POST /api/v2/auth/refresh did not reach refresh handler")
	}
}

func TestRegisterRegRoutesExposeCurrentEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &regRouteHandlerStub{}
	router := gin.New()
	routes.RegisterRegRoutes(router, handler)

	tests := []struct {
		name   string
		method string
		path   string
		called string
	}{
		{"create session", http.MethodPost, "/api/v2/register/session/register", "create-session"},
		{"verify session", http.MethodPatch, "/api/v2/register/session/verify", "verify-session"},
		{"create user", http.MethodPost, "/api/v2/register/", "create-user"},
		{"change password", http.MethodPost, "/api/v2/register/password/change", "change-password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRouteRequest(router, tt.method, tt.path, "")
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
			}
			if handler.called != tt.called {
				t.Fatalf("called = %q, want %q", handler.called, tt.called)
			}
		})
	}
}

func TestRegisterSubRoutesRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &subRouteHandlerStub{}
	router := gin.New()
	jwtUtils := utils.NewJWTUtils("test-secret")
	routes.RegisterSubRoutes(router, handler, middlewares.NewAuthMiddleware(jwtUtils))

	rec := performRouteRequest(router, http.MethodPost, "/api/v2/sub/UploadImg", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated upload status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if handler.called {
		t.Fatal("unauthenticated upload reached handler")
	}

	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alex", "alex", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec = performRouteRequest(router, http.MethodPost, "/api/v2/sub/UploadImg", tokenPair.AccessToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("authenticated upload status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !handler.called {
		t.Fatal("authenticated upload did not reach handler")
	}
}

func TestRegisterEventsRoutesPublicEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	popularHandler := &popularEventsRouteHandlerStub{}
	router := newEventsTestRouter(eventsHandler, popularHandler)

	tests := []struct {
		name       string
		path       string
		wantCalled func() bool
	}{
		{"genres", "/api/v2/events/genres", func() bool { return eventsHandler.called == "genres" }},
		{"references", "/api/v2/references", func() bool { return eventsHandler.called == "references" }},
		{"popular", "/api/v2/events/popular", func() bool { return popularHandler.called }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRouteRequest(router, http.MethodGet, tt.path, "")
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
			}
			if !tt.wantCalled() {
				t.Fatal("handler was not called")
			}
		})
	}
}

func TestRegisterEventsRoutesProtectedEndpointRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newEventsTestRouter(&eventsRouteHandlerStub{}, &popularEventsRouteHandlerStub{})
	rec := performRouteRequest(router, http.MethodGet, "/api/v2/events/123", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterGroupsRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	jwtUtils := utils.NewJWTUtils("test-secret")
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(nil)
	routes.RegisterGroupsRoutes(router, &groupRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"details", http.MethodGet, "/api/v2/groups/1"},
		{"create", http.MethodPost, "/api/v2/groups"},
		{"join", http.MethodPost, "/api/v2/groups/1/join"},
		{"operator-list", http.MethodGet, "/api/v2/groups/1/requests"},
		{"admin-actions", http.MethodGet, "/api/v2/groups/1/actions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRouteRequest(router, tt.method, tt.path, "")
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func newEventsTestRouter(eventsHandler *eventsRouteHandlerStub, popularHandler *popularEventsRouteHandlerStub) *gin.Engine {
	router := gin.New()
	jwtUtils := utils.NewJWTUtils("test-secret")
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(nil)
	routes.RegisterEventsRoutes(router, eventsHandler, popularHandler, authMiddleware, groupRoleMiddleware)
	return router
}

func performRouteRequest(router *gin.Engine, method, path, accessToken string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	router.ServeHTTP(rec, req)
	return rec
}
