package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/middlewares"
	"friendship/models"
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

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := payload["your_role"]; got != "member" {
		t.Fatalf("your_role = %v, want member", got)
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
