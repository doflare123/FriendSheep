package tests

import (
	"errors"
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

func TestGroupHandlerAcceptJoinInviteErrorMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", group.ErrInviteNotFound, http.StatusNotFound},
		{"not owned", group.ErrInviteNotOwned, http.StatusForbidden},
		{"already handled", group.ErrInviteAlreadyHandled, http.StatusBadRequest},
		{"blacklisted", group.ErrUserInBlacklist, http.StatusForbidden},
		{"internal", errors.New("storage failed"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			groupHandler := handlers.NewGroupHandler(&inviteErrorGroupService{acceptErr: tt.err})
			router := gin.New()
			router.POST("/api/v2/groups/invites/:inviteId/accept", withTestUserID(42), groupHandler.AcceptJoinInvite)

			req := httptest.NewRequest(http.MethodPost, "/api/v2/groups/invites/7/accept", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
			payload := assertCommonErrorShape(t, rec)
			if tt.want == http.StatusInternalServerError && payload["details"] == "" {
				t.Fatalf("details is empty for internal error; payload=%#v", payload)
			}
		})
	}
}

func TestGroupHandlerRejectJoinInviteErrorMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", group.ErrInviteNotFound, http.StatusNotFound},
		{"not owned", group.ErrInviteNotOwned, http.StatusForbidden},
		{"already handled", group.ErrInviteAlreadyHandled, http.StatusBadRequest},
		{"internal", errors.New("storage failed"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			groupHandler := handlers.NewGroupHandler(&inviteErrorGroupService{rejectErr: tt.err})
			router := gin.New()
			router.POST("/api/v2/groups/invites/:inviteId/reject", withTestUserID(42), groupHandler.RejectJoinInvite)

			req := httptest.NewRequest(http.MethodPost, "/api/v2/groups/invites/7/reject", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
			payload := assertCommonErrorShape(t, rec)
			if tt.want == http.StatusInternalServerError && payload["details"] == "" {
				t.Fatalf("details is empty for internal error; payload=%#v", payload)
			}
		})
	}
}

type inviteErrorGroupService struct {
	group.GroupsService
	acceptErr error
	rejectErr error
}

func (s *inviteErrorGroupService) AcceptJoinInvite(uint, uint) (*group.GroupResult, error) {
	if s.acceptErr != nil {
		return nil, s.acceptErr
	}
	return &group.GroupResult{Joined: true}, nil
}

func (s *inviteErrorGroupService) RejectJoinInvite(uint, uint) (bool, error) {
	if s.rejectErr != nil {
		return false, s.rejectErr
	}
	return true, nil
}

func withTestUserID(userID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}
