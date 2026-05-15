package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	group "friendship/services/groups"

	"github.com/gin-gonic/gin"
)

func TestGroupHandlerUpdateGroupBindErrorUsesCommonErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var srv group.GroupsService
	groupHandler := handlers.NewGroupHandler(srv)

	router := gin.New()
	router.PUT("/api/v2/groups", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	}, groupHandler.UpdateGroup)

	req := httptest.NewRequest(http.MethodPut, "/api/v2/groups", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	payload := assertCommonErrorShape(t, rec)
	if payload["details"] == "" {
		t.Fatalf("details is empty, expected bind error details; payload=%#v", payload)
	}
}
