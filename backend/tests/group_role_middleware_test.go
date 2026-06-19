package tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/middlewares"
	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestGroupRoleMiddlewareRejectsUnauthorizedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/groups/:groupId", middlewares.NewGroupRoleMiddleware(nil).RequireGroupRole("allowed"), func(c *gin.Context) {
		t.Fatal("handler should not be reached")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/groups/1", nil)

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertCommonErrorShape(t, rec)
}

func TestGroupRoleCapabilitySourcePreservesAccessMatrix(t *testing.T) {
	tests := []struct {
		role          string
		adminAccess   bool
		operateAccess bool
		memberAccess  bool
	}{
		{role: groupmodels.RoleAdmin, adminAccess: true, operateAccess: true, memberAccess: true},
		{role: groupmodels.RoleModerator, operateAccess: true, memberAccess: true},
		{role: groupmodels.RoleMember, memberAccess: true},
		{role: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			capability, knownRole := groupmodels.CapabilityOf(tt.role)
			if knownRole != (tt.role != "unknown") {
				t.Fatalf("known role = %v, want %v", knownRole, tt.role != "unknown")
			}
			if tt.adminAccess && capability != groupmodels.CapabilityAdmin {
				t.Fatalf("capability = %v, want admin", capability)
			}
			if !tt.adminAccess && tt.operateAccess && capability != groupmodels.CapabilityModerate {
				t.Fatalf("capability = %v, want moderate", capability)
			}
			if !tt.operateAccess && tt.memberAccess && capability != groupmodels.CapabilityMember {
				t.Fatalf("capability = %v, want member", capability)
			}
			if got := groupmodels.HasCapability(tt.role, groupmodels.CapabilityAdmin); got != tt.adminAccess {
				t.Fatalf("admin capability = %v, want %v", got, tt.adminAccess)
			}
			if got := groupmodels.HasCapability(tt.role, groupmodels.CapabilityModerate); got != tt.operateAccess {
				t.Fatalf("moderate capability = %v, want %v", got, tt.operateAccess)
			}
			if got := groupmodels.HasCapability(tt.role, groupmodels.CapabilityMember); got != tt.memberAccess {
				t.Fatalf("member capability = %v, want %v", got, tt.memberAccess)
			}
		})
	}

	assertSameStringSlice(t, groupmodels.RolesWithCapability(groupmodels.CapabilityAdmin), []string{groupmodels.RoleAdmin})
	assertSameStringSlice(t, groupmodels.RolesWithCapability(groupmodels.CapabilityModerate), []string{groupmodels.RoleAdmin, groupmodels.RoleModerator})
	assertSameStringSlice(t, groupmodels.RolesWithCapability(groupmodels.CapabilityMember), []string{groupmodels.RoleAdmin, groupmodels.RoleModerator, groupmodels.RoleMember})
	assertSameStringSlice(t, groupmodels.RolesWithCapability(groupmodels.Capability(99)), nil)
}

func TestGroupRoleMiddlewareCapabilityWrappersPreserveRouteAccess(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		middleware func(*middlewares.GroupRoleMiddleware) gin.HandlerFunc
		wantStatus int
	}{
		{"admin allows admin", groupmodels.RoleAdmin, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireAdmin() }, http.StatusNoContent},
		{"admin rejects moderator", groupmodels.RoleModerator, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireAdmin() }, http.StatusForbidden},
		{"admin rejects member", groupmodels.RoleMember, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireAdmin() }, http.StatusForbidden},
		{"operator allows admin", groupmodels.RoleAdmin, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireOperatorOrAdmin() }, http.StatusNoContent},
		{"operator allows moderator", groupmodels.RoleModerator, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireOperatorOrAdmin() }, http.StatusNoContent},
		{"operator rejects member", groupmodels.RoleMember, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireOperatorOrAdmin() }, http.StatusForbidden},
		{"member allows admin", groupmodels.RoleAdmin, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireMember() }, http.StatusNoContent},
		{"member allows moderator", groupmodels.RoleModerator, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireMember() }, http.StatusNoContent},
		{"member allows member", groupmodels.RoleMember, func(m *middlewares.GroupRoleMiddleware) gin.HandlerFunc { return m.RequireMember() }, http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newGroupRoleRepo(t)
			seedGroupRoleMembership(t, repo.db, 71, 88, tt.role)

			router := gin.New()
			router.GET("/groups/:groupId", withUserID(71), tt.middleware(middlewares.NewGroupRoleMiddleware(repo)), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/groups/88", nil)

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus >= 400 {
				assertCommonErrorShape(t, rec)
			}
		})
	}
}

func TestGroupRoleMiddlewareEventCapabilityWrapperUsesEventGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	seedGroupRoleMembership(t, repo.db, 72, 89, groupmodels.RoleModerator)
	seedGroupRoleEvent(t, repo.db, 93, 89, 72)

	router := gin.New()
	router.GET("/events/:eventId", withUserID(72), middlewares.NewGroupRoleMiddleware(repo).RequireEventOperatorOrAdmin(), func(c *gin.Context) {
		if got := c.GetUint("eventID"); got != 93 {
			t.Fatalf("eventID = %d, want 93", got)
		}
		if got := c.GetUint("groupID"); got != 89 {
			t.Fatalf("groupID = %d, want 89", got)
		}
		if got := c.GetString("groupRole"); got != groupmodels.RoleModerator {
			t.Fatalf("groupRole = %q, want %q", got, groupmodels.RoleModerator)
		}
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/93", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestGroupRoleMiddlewareJoinRequestCapabilityWrapperUsesRequestGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	seedGroupRoleMembership(t, repo.db, 73, 90, groupmodels.RoleMember)
	seedGroupRoleJoinRequest(t, repo.db, 94, 90, 101)

	router := gin.New()
	router.POST("/requests/:requestId", withUserID(73), middlewares.NewGroupRoleMiddleware(repo).RequireJoinRequestOperatorOrAdmin(), func(c *gin.Context) {
		t.Fatal("handler should not be reached")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/requests/94", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	payload := assertCommonErrorShape(t, rec)
	requiredRoles, ok := payload["required_role"].([]any)
	if !ok {
		t.Fatalf("required_role = %#v, want list", payload["required_role"])
	}
	if len(requiredRoles) != 2 || requiredRoles[0] != groupmodels.RoleAdmin || requiredRoles[1] != groupmodels.RoleModerator {
		t.Fatalf("required_role = %#v, want admin and moderator roles", requiredRoles)
	}
	if got := payload["your_role"]; got != groupmodels.RoleMember {
		t.Fatalf("your_role = %v, want %q", got, groupmodels.RoleMember)
	}
}

func TestGroupRoleMiddlewareRequireMemberGroupEventsRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	seedGroupRoleMembership(t, repo.db, 81, 91, groupmodels.RoleMember)

	router := gin.New()
	router.GET("/api/v2/groups/events/:groupId/events", withUserID(81), middlewares.NewGroupRoleMiddleware(repo).RequireMember(), func(c *gin.Context) {
		if got := c.GetUint("groupID"); got != 91 {
			t.Fatalf("groupID = %d, want 91", got)
		}
		if got := c.GetString("groupRole"); got != groupmodels.RoleMember {
			t.Fatalf("groupRole = %q, want %q", got, groupmodels.RoleMember)
		}
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/groups/events/91/events", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestGroupRoleMiddlewareRequireMemberGroupEventsRouteRejectsNonMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	seedGroupRoleMembership(t, repo.db, 81, 91, groupmodels.RoleMember)

	router := gin.New()
	router.GET("/api/v2/groups/events/:groupId/events", withUserID(82), middlewares.NewGroupRoleMiddleware(repo).RequireMember(), func(c *gin.Context) {
		t.Fatal("handler should not be reached")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/groups/events/91/events", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	assertCommonErrorShape(t, rec)
}

func TestGroupRoleMiddlewareRejectsMissingAndBadGroupID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
	}{
		{
			name:       "missing group id",
			method:     http.MethodPost,
			target:     "/groups",
			body:       `{"name":"no-group"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad path group id",
			method:     http.MethodGet,
			target:     "/groups/not-a-number",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad query group id",
			method:     http.MethodGet,
			target:     "/groups?groupId=bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			middleware := middlewares.NewGroupRoleMiddleware(nil).RequireGroupRole("allowed")

			router.GET("/groups/:groupId", withUserID(7), middleware, func(c *gin.Context) {
				t.Fatal("handler should not be reached")
			})
			router.GET("/groups", withUserID(7), middleware, func(c *gin.Context) {
				t.Fatal("handler should not be reached")
			})
			router.POST("/groups", withUserID(7), middleware, func(c *gin.Context) {
				t.Fatal("handler should not be reached")
			})

			var body io.Reader
			if tt.body != "" {
				body = bytes.NewBufferString(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.target, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestGroupRoleMiddlewareRejectsNonMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	seedGroupRoleUserAndGroup(t, repo.db, 11, 33)

	router := gin.New()
	router.GET("/groups/:groupId", withUserID(11), middlewares.NewGroupRoleMiddleware(repo).RequireGroupRole("allowed"), func(c *gin.Context) {
		t.Fatal("handler should not be reached")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/groups/33", nil)

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGroupRoleMiddlewareRejectsForbiddenRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	roleID := seedGroupRoleMembership(t, repo.db, 12, 44, "member")
	if roleID == 0 {
		t.Fatal("roleID = 0")
	}

	router := gin.New()
	router.GET("/groups/:groupId", withUserID(12), middlewares.NewGroupRoleMiddleware(repo).RequireGroupRole("allowed"), func(c *gin.Context) {
		t.Fatal("handler should not be reached")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/groups/44", nil)

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	payload := assertCommonErrorShape(t, rec)
	if got := payload["your_role"]; got != "member" {
		t.Fatalf("your_role = %v, want member", got)
	}
}

func assertSameStringSlice(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice length = %d, want %d; got=%#v want=%#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice[%d] = %q, want %q; got=%#v want=%#v", i, got[i], want[i], got, want)
		}
	}
}

func TestGroupRoleMiddlewareAllowsPathAndQueryGroupID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		target string
	}{
		{name: "path group id", target: "/groups/55"},
		{name: "query group id", target: "/groups?groupId=55"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newGroupRoleRepo(t)
			seedGroupRoleMembership(t, repo.db, 21, 55, "allowed")

			router := gin.New()
			middleware := middlewares.NewGroupRoleMiddleware(repo).RequireGroupRole("allowed")
			reached := false

			router.GET("/groups/:groupId", withUserID(21), middleware, func(c *gin.Context) {
				reached = true
				if got := c.GetUint("groupID"); got != 55 {
					t.Fatalf("groupID = %d, want 55", got)
				}
				if got := c.GetString("groupRole"); got != "allowed" {
					t.Fatalf("groupRole = %q, want allowed", got)
				}
				c.Status(http.StatusNoContent)
			})
			router.GET("/groups", withUserID(21), middleware, func(c *gin.Context) {
				reached = true
				if got := c.GetUint("groupID"); got != 55 {
					t.Fatalf("groupID = %d, want 55", got)
				}
				if got := c.GetString("groupRole"); got != "allowed" {
					t.Fatalf("groupRole = %q, want allowed", got)
				}
				c.Status(http.StatusNoContent)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
			}
			if !reached {
				t.Fatal("handler was not reached")
			}
		})
	}
}

func TestGroupRoleMiddlewareReadsGroupIDFromJSONBodyAndPreservesBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGroupRoleRepo(t)
	seedGroupRoleMembership(t, repo.db, 31, 66, "allowed")

	router := gin.New()
	router.POST("/groups/check", withUserID(31), middlewares.NewGroupRoleMiddleware(repo).RequireGroupRole("allowed"), func(c *gin.Context) {
		if got := c.GetUint("groupID"); got != 66 {
			t.Fatalf("groupID = %d, want 66", got)
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if got := string(bodyBytes); got != `{"groupId":66,"action":"ping"}` {
			t.Fatalf("body = %q, want original JSON", got)
		}

		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/groups/check", bytes.NewBufferString(`{"groupId":66,"action":"ping"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func withUserID(userID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

func newGroupRoleRepo(t *testing.T) *testPostgresRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&groupmodels.Group{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&eventmodels.Event{},
		&groupmodels.GroupJoinRequest{},
	); err != nil {
		t.Fatalf("auto migrate group role models: %v", err)
	}

	return &testPostgresRepository{db: db}
}

func seedGroupRoleMembership(t *testing.T, db *gorm.DB, userID, groupID uint, roleName string) uint {
	t.Helper()

	role := groupmodels.Role_in_group{Name: roleName}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	seedGroupRoleUserAndGroup(t, db, userID, groupID)

	membership := groupmodels.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: role.Id,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create group membership: %v", err)
	}

	return role.Id
}

func seedGroupRoleEvent(t *testing.T, db *gorm.DB, eventID, groupID, creatorID uint) {
	t.Helper()

	event := eventmodels.Event{
		ID:          eventID,
		Title:       "Test Event",
		GroupID:     groupID,
		CreatorID:   creatorID,
		EventTypeID: 1,
		MaxUsers:    10,
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatalf("create event: %v", err)
	}
}

func seedGroupRoleJoinRequest(t *testing.T, db *gorm.DB, requestID, groupID, userID uint) {
	t.Helper()

	user := models.User{
		ID:       userID,
		Name:     "Join Request User",
		Password: "Password123!",
		Us:       "join-request-user",
		Email:    "join-request-user@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create join request user: %v", err)
	}

	request := groupmodels.GroupJoinRequest{
		ID:      requestID,
		UserID:  userID,
		GroupID: groupID,
		Status:  groupmodels.JoinStatusPending,
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatalf("create join request: %v", err)
	}
}

func seedGroupRoleUserAndGroup(t *testing.T, db *gorm.DB, userID, groupID uint) {
	t.Helper()

	user := models.User{
		ID:       userID,
		Name:     "Test User",
		Password: "Password123!",
		Us:       "user-role-test",
		Email:    "user-role-test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	group := groupmodels.Group{
		ID:               groupID,
		Name:             "Test Group",
		Description:      "Description",
		SmallDescription: "Small",
		Image:            "image.png",
		CreaterID:        userID,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
}
