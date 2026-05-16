package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/middlewares"
	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	"friendship/routes"
	"friendship/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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
func (h *eventsRouteHandlerStub) UpdateEvent(c *gin.Context) {
	h.called = "update-event"
	c.Status(http.StatusNoContent)
}
func (h *eventsRouteHandlerStub) DeleteEvent(c *gin.Context) {
	h.called = "delete-event"
	c.Status(http.StatusNoContent)
}
func (h *eventsRouteHandlerStub) GetGroupEvents(c *gin.Context) {
	h.called = "group-events"
	c.Status(http.StatusNoContent)
}
func (h *eventsRouteHandlerStub) GetEventDetails(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) GetEventDetailsForAdmin(c *gin.Context) {
	h.called = "admin-details"
	c.Status(http.StatusNoContent)
}
func (h *eventsRouteHandlerStub) JoinEvent(c *gin.Context)  { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) LeaveEvent(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *eventsRouteHandlerStub) KickUserFromEvent(c *gin.Context) {
	h.called = "kick-user"
	c.Status(http.StatusNoContent)
}

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

type groupRouteHandlerStub struct {
	called string
}

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
func (h *groupRouteHandlerStub) ApproveJoinRequest(c *gin.Context) {
	h.called = "approve-request"
	c.Status(http.StatusNoContent)
}
func (h *groupRouteHandlerStub) RejectJoinRequest(c *gin.Context) {
	h.called = "reject-request"
	c.Status(http.StatusNoContent)
}
func (h *groupRouteHandlerStub) GetJoinRequests(c *gin.Context)     { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) AddPermissions(c *gin.Context)      { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RemovePermissions(c *gin.Context)   { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) GetGroupBlacklist(c *gin.Context)   { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) DeleteUserFromGroup(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RemoveFromBlacklist(c *gin.Context) { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) CreateJoinInvite(c *gin.Context)    { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) AcceptJoinInvite(c *gin.Context)    { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) RejectJoinInvite(c *gin.Context)    { c.Status(http.StatusNoContent) }
func (h *groupRouteHandlerStub) WatchRecentActions(c *gin.Context)  { c.Status(http.StatusNoContent) }

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
	assertCommonErrorShape(t, rec)
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

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/events/genres", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status for removed genres route = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertCommonErrorShape(t, rec)
}

func TestRegisterEventsRoutesProtectedEndpointRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newEventsTestRouter(&eventsRouteHandlerStub{}, &popularEventsRouteHandlerStub{})
	rec := performRouteRequest(router, http.MethodGet, "/api/v2/events/123", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertCommonErrorShape(t, rec)
}

func TestRegisterEventsRoutesGroupEventsAllowsMemberRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, groupID, memberID := newGroupEventsRouteRepo(t, groupmodels.RoleMember, false)

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterEventsRoutes(router, eventsHandler, &popularEventsRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(memberID, "Member", "member", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/groups/events/"+testUintString(groupID)+"/events", tokenPair.AccessToken)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if eventsHandler.called != "group-events" {
		t.Fatalf("called = %q, want %q", eventsHandler.called, "group-events")
	}
}

func TestRegisterEventsRoutesGroupEventsAllowsNonMemberForPublicGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, groupID, _ := newGroupEventsRouteRepo(t, groupmodels.RoleMember, false)

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterEventsRoutes(router, eventsHandler, &popularEventsRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(9999, "Outsider", "outsider", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/groups/events/"+testUintString(groupID)+"/events", tokenPair.AccessToken)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if eventsHandler.called != "group-events" {
		t.Fatalf("called = %q, want %q", eventsHandler.called, "group-events")
	}
}

func TestRegisterEventsRoutesGroupEventsPassesThroughToHandlerForPrivateGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, groupID, _ := newGroupEventsRouteRepo(t, groupmodels.RoleMember, true)

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterEventsRoutes(router, eventsHandler, &popularEventsRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(9999, "Outsider", "outsider", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/groups/events/"+testUintString(groupID)+"/events", tokenPair.AccessToken)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if eventsHandler.called != "group-events" {
		t.Fatalf("called = %q, want %q", eventsHandler.called, "group-events")
	}
}

func TestRegisterEventsRoutesGroupEventsRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	_, groupID, _ := newGroupEventsRouteRepo(t, groupmodels.RoleMember, false)
	router := newEventsTestRouter(eventsHandler, &popularEventsRouteHandlerStub{})

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/groups/events/"+testUintString(groupID)+"/events", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	assertCommonErrorShape(t, rec)
	if eventsHandler.called != "" {
		t.Fatalf("handler was called: %q", eventsHandler.called)
	}
}

func TestRegisterEventsRoutesEventIDAdminEndpointsReachHandlerAfterAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, eventID := newEventAdminRouteRepo(t, 42, "Админ")

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterEventsRoutes(router, eventsHandler, &popularEventsRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alex", "alex", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		wantCalled string
	}{
		{"admin details", http.MethodGet, "/api/v2/admin/events/" + testUintString(eventID), "admin-details"},
		{"admin update", http.MethodPut, "/api/v2/admin/events/" + testUintString(eventID), "update-event"},
		{"admin delete", http.MethodDelete, "/api/v2/admin/events/" + testUintString(eventID), "delete-event"},
		{"admin kick", http.MethodDelete, "/api/v2/admin/events/" + testUintString(eventID) + "/kick/99", "kick-user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventsHandler.called = ""
			rec := performRouteRequest(router, tt.method, tt.path, tokenPair.AccessToken)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, rec.Code, http.StatusNoContent)
			}
			if eventsHandler.called != tt.wantCalled {
				t.Fatalf("called = %q, want %q", eventsHandler.called, tt.wantCalled)
			}
		})
	}
}

func TestRegisterEventsRoutesEventIDAdminEndpointsRejectNonAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventsHandler := &eventsRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, eventID := newEventAdminRouteRepo(t, 42, "Участник")

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterEventsRoutes(router, eventsHandler, &popularEventsRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alex", "alex", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodGet, "/api/v2/admin/events/"+testUintString(eventID), tokenPair.AccessToken)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	assertCommonErrorShape(t, rec)
	if eventsHandler.called != "" {
		t.Fatalf("handler was called: %q", eventsHandler.called)
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
			assertCommonErrorShape(t, rec)
		})
	}
}

func TestRegisterGroupsRoutesRequestIDActionsRequireOperatorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupHandler := &groupRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, requestID := newGroupJoinRequestRouteRepo(t, 42, "Админ")

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterGroupsRoutes(router, groupHandler, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alex", "alex", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodPost, "/api/v2/groups/requests/"+testUintString(requestID)+"/approve", tokenPair.AccessToken)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if groupHandler.called != "approve-request" {
		t.Fatalf("called = %q, want approve-request", groupHandler.called)
	}
}

func TestRegisterGroupsRoutesRequestIDActionsRejectPlainMemberBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupHandler := &groupRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, requestID := newGroupJoinRequestRouteRepo(t, 42, "Участник")

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterGroupsRoutes(router, groupHandler, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alex", "alex", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodPost, "/api/v2/groups/requests/"+testUintString(requestID)+"/reject", tokenPair.AccessToken)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	assertCommonErrorShape(t, rec)
	if groupHandler.called != "" {
		t.Fatalf("handler was called: %q", groupHandler.called)
	}
}

func TestRegisterGroupsRoutesRequestIDActionsReturnNotFoundBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupHandler := &groupRouteHandlerStub{}
	jwtUtils := utils.NewJWTUtils("test-secret")
	repo, _ := newGroupJoinRequestRouteRepo(t, 42, "Админ")

	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(repo)
	routes.RegisterGroupsRoutes(router, groupHandler, authMiddleware, groupRoleMiddleware)

	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alex", "alex", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	rec := performRouteRequest(router, http.MethodPost, "/api/v2/groups/requests/999/reject", tokenPair.AccessToken)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	assertCommonErrorShape(t, rec)
	if groupHandler.called != "" {
		t.Fatalf("handler was called: %q", groupHandler.called)
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

func newGroupJoinRequestRouteRepo(t *testing.T, actorID uint, actorRole string) (*testPostgresRepository, uint) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&groupmodels.Group{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&groupmodels.GroupJoinRequest{},
	); err != nil {
		t.Fatalf("auto migrate group request route models: %v", err)
	}

	actor := models.User{
		ID:       actorID,
		Name:     "Route Actor",
		Password: "Password123!",
		Us:       "route-actor",
		Email:    "route-actor@example.com",
	}
	if err := db.Create(&actor).Error; err != nil {
		t.Fatalf("create route actor: %v", err)
	}

	target := models.User{
		ID:       actorID + 1,
		Name:     "Route Target",
		Password: "Password123!",
		Us:       "route-target",
		Email:    "route-target@example.com",
	}
	if err := db.Create(&target).Error; err != nil {
		t.Fatalf("create route target: %v", err)
	}

	group := groupmodels.Group{
		Name:             "Route Group",
		Description:      "Route Group Description",
		SmallDescription: "Route Group",
		Image:            "https://example.com/group.png",
		CreaterID:        actorID,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create route group: %v", err)
	}

	role := groupmodels.Role_in_group{Name: actorRole}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create route role: %v", err)
	}

	membership := groupmodels.GroupUsers{
		UserID:        actorID,
		GroupID:       group.ID,
		RoleInGroupID: role.Id,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create route group membership: %v", err)
	}

	request := groupmodels.GroupJoinRequest{
		UserID:  target.ID,
		GroupID: group.ID,
		Status:  "pending",
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatalf("create route join request: %v", err)
	}

	return &testPostgresRepository{db: db}, request.ID
}

func newGroupEventsRouteRepo(t *testing.T, roleName string, isPrivate bool) (*testPostgresRepository, uint, uint) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&groupmodels.Group{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
	); err != nil {
		t.Fatalf("auto migrate group events route models: %v", err)
	}

	memberID := uint(77)
	member := models.User{
		ID:       memberID,
		Name:     "Route Member",
		Password: "Password123!",
		Us:       "route-member",
		Email:    "route-member@example.com",
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create route member: %v", err)
	}

	group := groupmodels.Group{
		Name:             "Route Group Events",
		Description:      "Route Group Events Description",
		SmallDescription: "Route Group Events",
		Image:            "https://example.com/group-events.png",
		CreaterID:        memberID,
		IsPrivate:        isPrivate,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create route group events: %v", err)
	}

	role := groupmodels.Role_in_group{Name: roleName}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create route group events role: %v", err)
	}

	membership := groupmodels.GroupUsers{
		UserID:        memberID,
		GroupID:       group.ID,
		RoleInGroupID: role.Id,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create route group events membership: %v", err)
	}

	return &testPostgresRepository{db: db}, group.ID, memberID
}

func newEventAdminRouteRepo(t *testing.T, userID uint, roleName string) (*testPostgresRepository, uint) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&groupmodels.Group{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&eventmodels.Event{},
		&eventmodels.EventLocation{},
		&eventmodels.Status{},
		&eventmodels.AgeLimit{},
	); err != nil {
		t.Fatalf("auto migrate admin event route models: %v", err)
	}

	user := models.User{
		ID:       userID,
		Name:     "Route User",
		Password: "Password123!",
		Us:       "route-user",
		Email:    "route-user@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create route user: %v", err)
	}

	group := groupmodels.Group{
		Name:             "Route Group",
		Description:      "Route Group Description",
		SmallDescription: "Route Group",
		Image:            "https://example.com/group.png",
		CreaterID:        userID,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create route group: %v", err)
	}

	role := groupmodels.Role_in_group{Name: roleName}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create route role: %v", err)
	}

	membership := groupmodels.GroupUsers{
		UserID:        userID,
		GroupID:       group.ID,
		RoleInGroupID: role.Id,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create route group membership: %v", err)
	}

	event := eventmodels.Event{
		Title:           "Route Event",
		Description:     "Route Event Description",
		GroupID:         group.ID,
		EventTypeID:     1,
		EventLocationID: 1,
		CreatorID:       userID,
		StatusID:        1,
		AgeLimitID:      1,
		MaxUsers:        10,
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatalf("create route event: %v", err)
	}

	return &testPostgresRepository{db: db}, event.ID
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
