package tests

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"friendship/config"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type denyNamedRateLimitStore struct {
	policyName string
}

func (s denyNamedRateLimitStore) Take(
	_ context.Context,
	key string,
	_ int,
	window time.Duration,
) (middlewares.RateLimitDecision, error) {
	return middlewares.RateLimitDecision{
		Allowed:    !strings.Contains(key, ":"+s.policyName+":"),
		Count:      1,
		ResetAfter: window,
	}, nil
}

func TestImportantAuthenticatedRoutesHaveStrictRateLimitPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		path       string
		policyName string
	}{
		{name: "image upload", method: http.MethodPost, path: "/api/v2/sub/UploadImg", policyName: "image-upload-user"},
		{name: "create group", method: http.MethodPost, path: "/api/v2/groups", policyName: "group-create-user"},
		{name: "update group", method: http.MethodPut, path: "/api/v2/groups", policyName: "group-update-user"},
		{name: "delete group", method: http.MethodDelete, path: "/api/v2/groups/1", policyName: "destructive-admin-user"},
		{name: "join group", method: http.MethodPost, path: "/api/v2/groups/1/join", policyName: "membership-change-user"},
		{name: "leave group", method: http.MethodPost, path: "/api/v2/groups/1/leave", policyName: "membership-change-user"},
		{name: "accept invite", method: http.MethodPost, path: "/api/v2/groups/invites/1/accept", policyName: "group-invite-response-user"},
		{name: "reject invite", method: http.MethodPost, path: "/api/v2/groups/invites/1/reject", policyName: "group-invite-response-user"},
		{name: "approve request", method: http.MethodPost, path: "/api/v2/groups/requests/1/approve", policyName: "join-request-review-user"},
		{name: "reject request", method: http.MethodPost, path: "/api/v2/groups/requests/1/reject", policyName: "join-request-review-user"},
		{name: "approve all requests", method: http.MethodPost, path: "/api/v2/groups/1/requests/approve-all", policyName: "join-request-bulk-user"},
		{name: "reject all requests", method: http.MethodPost, path: "/api/v2/groups/1/requests/reject-all", policyName: "join-request-bulk-user"},
		{name: "add permissions", method: http.MethodPost, path: "/api/v2/groups/permissions/add", policyName: "group-permission-user"},
		{name: "remove permissions", method: http.MethodPost, path: "/api/v2/groups/permissions/remove", policyName: "group-permission-user"},
		{name: "remove member", method: http.MethodDelete, path: "/api/v2/groups/1/members/2", policyName: "member-moderation-user"},
		{name: "remove blacklist entry", method: http.MethodDelete, path: "/api/v2/groups/1/blacklist/2", policyName: "member-moderation-user"},
		{name: "create invite", method: http.MethodPost, path: "/api/v2/groups/invites", policyName: "group-invite-user"},
		{name: "create event", method: http.MethodPost, path: "/api/v2/admin/events", policyName: "event-create-user"},
		{name: "update event", method: http.MethodPut, path: "/api/v2/admin/events/1", policyName: "event-update-user"},
		{name: "delete event", method: http.MethodDelete, path: "/api/v2/admin/events/1", policyName: "destructive-admin-user"},
		{name: "kick event member", method: http.MethodDelete, path: "/api/v2/admin/events/1/kick/2", policyName: "member-moderation-user"},
		{name: "join event", method: http.MethodPost, path: "/api/v2/events/1/join", policyName: "membership-change-user"},
		{name: "leave event", method: http.MethodPost, path: "/api/v2/events/1/leave", policyName: "membership-change-user"},
	}

	jwtUtils := utils.NewJWTUtils("rate-limit-route-test-secret")
	tokenPair, err := jwtUtils.GenerateTokenPair(42, "Alice", "alice", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			rateLimiter := middlewares.NewRateLimitMiddlewareWithConfig(
				&testLogger{},
				denyNamedRateLimitStore{policyName: tt.policyName},
				"rate-limit-key-secret",
				config.DefaultRateLimitConfig(),
			)
			authMiddleware := middlewares.NewAuthMiddleware(jwtUtils)
			authMiddleware.SetRateLimiter(rateLimiter)
			groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(nil)

			routes.RegisterSubRoutes(router, &subRouteHandlerStub{}, authMiddleware)
			routes.RegisterGroupsRoutes(router, &groupRouteHandlerStub{}, authMiddleware, groupRoleMiddleware)
			routes.RegisterEventsRoutes(
				router,
				&eventsRouteHandlerStub{},
				&popularEventsRouteHandlerStub{},
				authMiddleware,
				groupRoleMiddleware,
			)

			recorder := performRouteRequest(router, tt.method, tt.path, tokenPair.AccessToken)
			if recorder.Code != http.StatusTooManyRequests {
				t.Fatalf(
					"%s %s status = %d, want %d for policy %s; body=%s",
					tt.method,
					tt.path,
					recorder.Code,
					http.StatusTooManyRequests,
					tt.policyName,
					recorder.Body.String(),
				)
			}
		})
	}
}
