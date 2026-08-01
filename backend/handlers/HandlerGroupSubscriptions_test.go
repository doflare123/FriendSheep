package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/models/dto"
	group "friendship/services/groups"

	"github.com/gin-gonic/gin"
)

type subscribedGroupsHandlerContextKey struct{}

type subscribedGroupsHandlerServiceStub struct {
	group.GroupsService
	result *dto.SubscribedGroupsResponseDto
	err    error
	calls  int
	ctx    context.Context
	userID uint
	page   int
	limit  int
}

func (s *subscribedGroupsHandlerServiceStub) GetSubscribedGroups(ctx context.Context, userID uint, page int, limit int) (*dto.SubscribedGroupsResponseDto, error) {
	s.calls++
	s.ctx = ctx
	s.userID = userID
	s.page = page
	s.limit = limit
	return s.result, s.err
}

func TestGroupHandlerGetSubscribedGroupsForwardsCurrentUserPaginationAndExactJSON(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantPage  int
		wantLimit int
	}{
		{name: "defaults", target: "/api/v2/users/me/groups/subscriptions", wantPage: group.DefaultSubscribedGroupsPage, wantLimit: group.DefaultSubscribedGroupsLimit},
		{name: "explicit page", target: "/api/v2/users/me/groups/subscriptions?page=2&limit=10", wantPage: 2, wantLimit: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			service := &subscribedGroupsHandlerServiceStub{result: &dto.SubscribedGroupsResponseDto{
				Items: []dto.ManagedGroupItemDto{{
					ID:               11,
					Name:             "Member group",
					Categories:       []string{"Board games", "Travel"},
					SmallDescription: "Subscription description",
					MemberCount:      14,
					Image:            "https://example.com/group.png",
				}},
				Total:   21,
				Page:    tt.wantPage,
				Limit:   tt.wantLimit,
				HasMore: true,
			}}
			handler := NewGroupHandler(service)
			recorder, ctx := newSubscribedGroupsHandlerContext(t, tt.target)
			requestCtx := context.WithValue(ctx.Request.Context(), subscribedGroupsHandlerContextKey{}, "subscription request")
			ctx.Request = ctx.Request.WithContext(requestCtx)
			ctx.Set("userID", uint(73))

			handler.GetSubscribedGroups(ctx)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if service.calls != 1 || service.userID != 73 || service.page != tt.wantPage || service.limit != tt.wantLimit {
				t.Fatalf("service call = count:%d user:%d page:%d limit:%d, want once with user:73 page:%d limit:%d", service.calls, service.userID, service.page, service.limit, tt.wantPage, tt.wantLimit)
			}
			if service.ctx.Value(subscribedGroupsHandlerContextKey{}) != "subscription request" {
				t.Fatal("request context was not passed to group service")
			}
			assertSubscribedGroupsJSONContract(t, recorder.Body.Bytes(), tt.wantPage, tt.wantLimit)
		})
	}
}

func TestGroupHandlerGetSubscribedGroupsSerializesEmptyItemsAsArray(t *testing.T) {
	service := &subscribedGroupsHandlerServiceStub{result: &dto.SubscribedGroupsResponseDto{
		Items: make([]dto.ManagedGroupItemDto, 0), Page: 1, Limit: 20,
	}}
	handler := NewGroupHandler(service)
	recorder, ctx := newSubscribedGroupsHandlerContext(t, "/api/v2/users/me/groups/subscriptions")
	ctx.Set("userID", uint(73))

	handler.GetSubscribedGroups(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if string(payload["items"]) != "[]" {
		t.Fatalf("items = %s, want []", payload["items"])
	}
}

func TestGroupHandlerGetSubscribedGroupsRejectsInvalidPaginationWithoutServiceCall(t *testing.T) {
	for _, query := range []string{"page=0", "page=-1", "page=bad", "limit=0", "limit=101", "limit=bad"} {
		t.Run(query, func(t *testing.T) {
			service := &subscribedGroupsHandlerServiceStub{}
			handler := NewGroupHandler(service)
			recorder, ctx := newSubscribedGroupsHandlerContext(t, "/api/v2/users/me/groups/subscriptions?"+query)
			ctx.Set("userID", uint(73))

			handler.GetSubscribedGroups(ctx)

			assertSubscribedGroupsHandlerError(t, recorder, http.StatusBadRequest, "")
			if service.calls != 0 {
				t.Fatalf("service calls = %d, want 0", service.calls)
			}
		})
	}
}

func TestGroupHandlerGetSubscribedGroupsRejectsMissingCurrentUser(t *testing.T) {
	service := &subscribedGroupsHandlerServiceStub{}
	handler := NewGroupHandler(service)
	recorder, ctx := newSubscribedGroupsHandlerContext(t, "/api/v2/users/me/groups/subscriptions")

	handler.GetSubscribedGroups(ctx)

	assertSubscribedGroupsHandlerError(t, recorder, http.StatusUnauthorized, "")
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
}

func TestGroupHandlerGetSubscribedGroupsMapsServiceFailureToInternalError(t *testing.T) {
	service := &subscribedGroupsHandlerServiceStub{err: errors.New("subscriptions unavailable")}
	handler := NewGroupHandler(service)
	recorder, ctx := newSubscribedGroupsHandlerContext(t, "/api/v2/users/me/groups/subscriptions")
	ctx.Set("userID", uint(73))

	handler.GetSubscribedGroups(ctx)

	assertSubscribedGroupsHandlerError(t, recorder, http.StatusInternalServerError, "subscriptions unavailable")
}

func newSubscribedGroupsHandlerContext(t *testing.T, target string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return recorder, ctx
}

func assertSubscribedGroupsJSONContract(t *testing.T, body []byte, wantPage, wantLimit int) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	wantKeys := []string{"items", "total", "page", "limit", "hasMore"}
	if len(payload) != len(wantKeys) {
		t.Fatalf("top-level keys = %#v, want exactly %v", payload, wantKeys)
	}
	for _, key := range wantKeys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("response is missing %q: %s", key, body)
		}
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(payload["items"], &items); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %#v, want one item", items)
	}
	item := items[0]
	wantItemKeys := []string{"id", "name", "categories", "smallDescription", "memberCount", "image"}
	if len(item) != len(wantItemKeys) {
		t.Fatalf("item keys = %#v, want exactly %v", item, wantItemKeys)
	}
	for _, key := range wantItemKeys {
		if _, ok := item[key]; !ok {
			t.Fatalf("item is missing %q: %#v", key, item)
		}
	}
	if item["id"] != float64(11) || item["name"] != "Member group" || item["smallDescription"] != "Subscription description" || item["memberCount"] != float64(14) || item["image"] != "https://example.com/group.png" {
		t.Fatalf("item = %#v, want exact subscription group fields", item)
	}
	if got, ok := item["categories"].([]interface{}); !ok || len(got) != 2 || got[0] != "Board games" || got[1] != "Travel" {
		t.Fatalf("categories = %#v, want [Board games Travel]", item["categories"])
	}

	var page int
	var limit int
	var total int64
	var hasMore bool
	if err := json.Unmarshal(payload["page"], &page); err != nil {
		t.Fatalf("decode page: %v", err)
	}
	if err := json.Unmarshal(payload["limit"], &limit); err != nil {
		t.Fatalf("decode limit: %v", err)
	}
	if err := json.Unmarshal(payload["total"], &total); err != nil {
		t.Fatalf("decode total: %v", err)
	}
	if err := json.Unmarshal(payload["hasMore"], &hasMore); err != nil {
		t.Fatalf("decode hasMore: %v", err)
	}
	if page != wantPage || limit != wantLimit || total != 21 || !hasMore {
		t.Fatalf("pagination = total:%d page:%d limit:%d hasMore:%t, want total:21 page:%d limit:%d hasMore:true", total, page, limit, hasMore, wantPage, wantLimit)
	}
}

func assertSubscribedGroupsHandlerError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}
	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error == "" || response.Message == "" || response.Error != response.Message {
		t.Fatalf("error response = %#v, want mirrored non-empty error and message", response)
	}
	if wantMessage != "" && response.Error != wantMessage {
		t.Fatalf("error = %q, want %q", response.Error, wantMessage)
	}
}
