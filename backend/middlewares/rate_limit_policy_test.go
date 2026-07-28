package middlewares

import (
	"net/http"
	"testing"
	"time"

	"friendship/config"
)

func TestPublicRateLimitPoliciesCoverActiveSensitiveAndPublicRoutes(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		wantNames   []string
		wantKeyKind []rateLimitKeyKind
		wantFailure []rateLimitFailureMode
	}{
		{
			name:        "login",
			method:      http.MethodPost,
			path:        "/api/v2/auth/login",
			wantNames:   []string{"auth-login-ip", "auth-login-email"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP, rateLimitByJSONField},
			wantFailure: []rateLimitFailureMode{rateLimitFailClosed, rateLimitFailClosed},
		},
		{
			name:        "refresh",
			method:      http.MethodPost,
			path:        "/api/v2/auth/refresh",
			wantNames:   []string{"auth-refresh-ip"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP},
			wantFailure: []rateLimitFailureMode{rateLimitFailClosed},
		},
		{
			name:        "registration session",
			method:      http.MethodPost,
			path:        "/api/v2/register/session/register",
			wantNames:   []string{"register-session-ip", "register-session-email"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP, rateLimitByJSONField},
			wantFailure: []rateLimitFailureMode{rateLimitFailClosed, rateLimitFailClosed},
		},
		{
			name:        "registration verification",
			method:      http.MethodPatch,
			path:        "/api/v2/register/session/verify",
			wantNames:   []string{"register-verify-ip", "register-verify-session"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP, rateLimitByJSONField},
			wantFailure: []rateLimitFailureMode{rateLimitFailClosed, rateLimitFailClosed},
		},
		{
			name:        "create user",
			method:      http.MethodPost,
			path:        "/api/v2/register/",
			wantNames:   []string{"register-create-ip", "register-create-session"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP, rateLimitByJSONField},
			wantFailure: []rateLimitFailureMode{rateLimitFailClosed, rateLimitFailClosed},
		},
		{
			name:        "change password",
			method:      http.MethodPost,
			path:        "/api/v2/register/password/change",
			wantNames:   []string{"password-change-ip", "password-change-session"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP, rateLimitByJSONField},
			wantFailure: []rateLimitFailureMode{rateLimitFailClosed, rateLimitFailClosed},
		},
		{
			name:        "search events",
			method:      http.MethodGet,
			path:        "/api/v2/events/search",
			wantNames:   []string{"events-search-ip"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP},
			wantFailure: []rateLimitFailureMode{rateLimitFailOpen},
		},
		{
			name:        "popular events",
			method:      http.MethodGet,
			path:        "/api/v2/events/popular",
			wantNames:   []string{"events-popular-ip"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP},
			wantFailure: []rateLimitFailureMode{rateLimitFailOpen},
		},
		{
			name:        "references",
			method:      http.MethodGet,
			path:        "/api/v2/references",
			wantNames:   []string{"references-read-ip"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP},
			wantFailure: []rateLimitFailureMode{rateLimitFailOpen},
		},
		{
			name:        "genre references",
			method:      http.MethodGet,
			path:        "/api/v2/references/genres",
			wantNames:   []string{"references-read-ip"},
			wantKeyKind: []rateLimitKeyKind{rateLimitByIP},
			wantFailure: []rateLimitFailureMode{rateLimitFailOpen},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies := publicRateLimitPolicies(tt.method, tt.path)
			assertRateLimitPolicySet(t, policies, tt.wantNames, tt.wantKeyKind, tt.wantFailure)
		})
	}
}

func TestAuthenticatedRateLimitPoliciesCoverImportantActiveRoutes(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantName string
	}{
		{name: "image upload", method: http.MethodPost, path: "/api/v2/sub/UploadImg", wantName: "image-upload-user"},
		{name: "group create", method: http.MethodPost, path: "/api/v2/groups", wantName: "group-create-user"},
		{name: "event create", method: http.MethodPost, path: "/api/v2/admin/events", wantName: "event-create-user"},
		{name: "group join", method: http.MethodPost, path: "/api/v2/groups/:groupId/join", wantName: "membership-change-user"},
		{name: "group leave", method: http.MethodPost, path: "/api/v2/groups/:groupId/leave", wantName: "membership-change-user"},
		{name: "event join", method: http.MethodPost, path: "/api/v2/events/:eventId/join", wantName: "membership-change-user"},
		{name: "event leave", method: http.MethodPost, path: "/api/v2/events/:eventId/leave", wantName: "membership-change-user"},
		{name: "group invite", method: http.MethodPost, path: "/api/v2/groups/invites", wantName: "group-invite-user"},
		{name: "accept group invite", method: http.MethodPost, path: "/api/v2/groups/invites/:inviteId/accept", wantName: "group-invite-response-user"},
		{name: "reject group invite", method: http.MethodPost, path: "/api/v2/groups/invites/:inviteId/reject", wantName: "group-invite-response-user"},
		{name: "approve all join requests", method: http.MethodPost, path: "/api/v2/groups/:groupId/requests/approve-all", wantName: "join-request-bulk-user"},
		{name: "reject all join requests", method: http.MethodPost, path: "/api/v2/groups/:groupId/requests/reject-all", wantName: "join-request-bulk-user"},
		{name: "approve join request", method: http.MethodPost, path: "/api/v2/groups/requests/:requestId/approve", wantName: "join-request-review-user"},
		{name: "reject join request", method: http.MethodPost, path: "/api/v2/groups/requests/:requestId/reject", wantName: "join-request-review-user"},
		{name: "add permissions", method: http.MethodPost, path: "/api/v2/groups/permissions/add", wantName: "group-permission-user"},
		{name: "remove permissions", method: http.MethodPost, path: "/api/v2/groups/permissions/remove", wantName: "group-permission-user"},
		{name: "update group", method: http.MethodPut, path: "/api/v2/groups", wantName: "group-update-user"},
		{name: "update event", method: http.MethodPut, path: "/api/v2/admin/events/:eventId", wantName: "event-update-user"},
		{name: "delete group", method: http.MethodDelete, path: "/api/v2/groups/:groupId", wantName: "destructive-admin-user"},
		{name: "delete event", method: http.MethodDelete, path: "/api/v2/admin/events/:eventId", wantName: "destructive-admin-user"},
		{name: "remove group member", method: http.MethodDelete, path: "/api/v2/groups/:groupId/members/:userId", wantName: "member-moderation-user"},
		{name: "remove from blacklist", method: http.MethodDelete, path: "/api/v2/groups/:groupId/blacklist/:userId", wantName: "member-moderation-user"},
		{name: "kick event member", method: http.MethodDelete, path: "/api/v2/admin/events/:eventId/kick/:userId", wantName: "member-moderation-user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies := authenticatedRateLimitPolicies(tt.method, tt.path)
			if len(policies) != 1 {
				t.Fatalf("policies = %#v, want exactly one policy", policies)
			}
			policy := policies[0]
			if policy.name != tt.wantName {
				t.Fatalf("policy name = %q, want %q", policy.name, tt.wantName)
			}
			if policy.keyKind != rateLimitByUser {
				t.Fatalf("key kind = %d, want rateLimitByUser", policy.keyKind)
			}
			if policy.onFailure != rateLimitFailClosed {
				t.Fatalf("failure mode = %d, want fail closed", policy.onFailure)
			}
			if policy.limit <= 0 || policy.window <= 0 {
				t.Fatalf("invalid policy quota: %#v", policy)
			}
		})
	}
}

func TestRateLimitPoliciesIgnoreWrongMethodsAndUnlistedRoutes(t *testing.T) {
	if got := publicRateLimitPolicies(http.MethodGet, "/api/v2/auth/login"); len(got) != 0 {
		t.Fatalf("GET login policies = %#v, want none", got)
	}
	if got := authenticatedRateLimitPolicies(http.MethodGet, "/api/v2/groups"); len(got) != 0 {
		t.Fatalf("GET groups policies = %#v, want none", got)
	}
}

func TestRateLimitPoliciesUseProvidedCatalogQuotas(t *testing.T) {
	rateLimits := config.DefaultRateLimitConfig()
	rateLimits.AuthLoginIPLimit = 13
	rateLimits.AuthLoginIPWindow = 37 * time.Second
	rateLimits.MembershipChangeUserLimit = 9
	rateLimits.MembershipChangeUserWindow = 3 * time.Minute
	catalog := newRateLimitCatalog(rateLimits)

	publicPolicies := publicRateLimitPoliciesWithCatalog(
		http.MethodPost,
		"/api/v2/auth/login",
		catalog,
	)
	if len(publicPolicies) != 2 {
		t.Fatalf("login policies = %#v, want two policies", publicPolicies)
	}
	if got := publicPolicies[0]; got.limit != 13 || got.window != 37*time.Second {
		t.Fatalf("auth-login-ip quota = {limit:%d window:%s}, want {limit:13 window:37s}", got.limit, got.window)
	}

	authenticatedPolicies := authenticatedRateLimitPoliciesWithCatalog(
		http.MethodPost,
		"/api/v2/events/:eventId/join",
		catalog,
	)
	if len(authenticatedPolicies) != 1 {
		t.Fatalf("membership policies = %#v, want one policy", authenticatedPolicies)
	}
	if got := authenticatedPolicies[0]; got.limit != 9 || got.window != 3*time.Minute {
		t.Fatalf("membership quota = {limit:%d window:%s}, want {limit:9 window:3m}", got.limit, got.window)
	}
}

func assertRateLimitPolicySet(
	t *testing.T,
	policies []rateLimitPolicy,
	wantNames []string,
	wantKeyKinds []rateLimitKeyKind,
	wantFailureModes []rateLimitFailureMode,
) {
	t.Helper()

	if len(policies) != len(wantNames) {
		t.Fatalf("policies = %#v, want %d policies", policies, len(wantNames))
	}
	for i, policy := range policies {
		if policy.name != wantNames[i] {
			t.Errorf("policy[%d].name = %q, want %q", i, policy.name, wantNames[i])
		}
		if policy.keyKind != wantKeyKinds[i] {
			t.Errorf("policy[%d].keyKind = %d, want %d", i, policy.keyKind, wantKeyKinds[i])
		}
		if policy.onFailure != wantFailureModes[i] {
			t.Errorf("policy[%d].onFailure = %d, want %d", i, policy.onFailure, wantFailureModes[i])
		}
		if policy.limit <= 0 || policy.window <= 0 {
			t.Errorf("policy[%d] has invalid quota: %#v", i, policy)
		}
	}
}
