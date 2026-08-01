package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"friendship/models/dto"
	group "friendship/services/groups"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type managedGroupsHandlerServiceStub struct {
	group.GroupsService
	result     *dto.ManagedGroupsDto
	err        error
	lastUserID uint
	lastCtx    context.Context
	calls      int
}

func (s *managedGroupsHandlerServiceStub) GetManagedGroups(ctx context.Context, userID uint) (*dto.ManagedGroupsDto, error) {
	s.calls++
	s.lastCtx = ctx
	s.lastUserID = userID
	return s.result, s.err
}

func TestGroupHandlerGetManagedGroupsPassesExactRequestContext(t *testing.T) {
	service := &managedGroupsHandlerServiceStub{result: &dto.ManagedGroupsDto{
		Admin:     make([]dto.ManagedGroupItemDto, 0),
		Moderator: make([]dto.ManagedGroupItemDto, 0),
	}}
	handler := NewGroupHandler(service)
	recorder, ctx := newManagedGroupsHandlerContext(t)
	ctx.Set("userID", uint(73))

	type contextKey string
	const key contextKey = "managed-groups-request-context"
	requestContext := context.WithValue(context.Background(), key, "request-value")
	ctx.Request = ctx.Request.WithContext(requestContext)

	handler.GetManagedGroups(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.lastCtx != requestContext {
		t.Fatal("service did not receive the exact request context")
	}
	if got := service.lastCtx.Value(key); got != "request-value" {
		t.Fatalf("service context value = %v, want request-value", got)
	}
}

func TestGroupHandlerGetManagedGroupsUsesCurrentUserAndExactJSONContract(t *testing.T) {
	service := &managedGroupsHandlerServiceStub{result: &dto.ManagedGroupsDto{
		Admin: []dto.ManagedGroupItemDto{{
			ID:               11,
			Name:             "Admin group",
			Categories:       []string{"Board games", "Travel"},
			SmallDescription: "Admin description",
			MemberCount:      14,
			Image:            "https://example.com/admin.png",
		}},
		Moderator: []dto.ManagedGroupItemDto{{
			ID:               22,
			Name:             "Moderator group",
			Categories:       []string{"Sport"},
			SmallDescription: "Moderator description",
			MemberCount:      27,
			Image:            "https://example.com/moderator.png",
		}},
	}}
	handler := NewGroupHandler(service)
	recorder, ctx := newManagedGroupsHandlerContext(t)
	ctx.Set("userID", uint(73))

	handler.GetManagedGroups(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.calls != 1 || service.lastUserID != 73 {
		t.Fatalf("service call = count:%d userID:%d, want current user 73 once", service.calls, service.lastUserID)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload) != 2 || payload["admin"] == nil || payload["moderator"] == nil {
		t.Fatalf("top-level keys = %#v, want exactly admin and moderator", payload)
	}
	assertManagedGroupItemJSONContract(t, payload["admin"], map[string]interface{}{
		"id":               float64(11),
		"name":             "Admin group",
		"categories":       []interface{}{"Board games", "Travel"},
		"smallDescription": "Admin description",
		"memberCount":      float64(14),
		"image":            "https://example.com/admin.png",
	})
	assertManagedGroupItemJSONContract(t, payload["moderator"], map[string]interface{}{
		"id":               float64(22),
		"name":             "Moderator group",
		"categories":       []interface{}{"Sport"},
		"smallDescription": "Moderator description",
		"memberCount":      float64(27),
		"image":            "https://example.com/moderator.png",
	})
}

func TestGroupHandlerGetManagedGroupsSerializesEmptySectionsAsArrays(t *testing.T) {
	service := &managedGroupsHandlerServiceStub{result: &dto.ManagedGroupsDto{
		Admin:     make([]dto.ManagedGroupItemDto, 0),
		Moderator: make([]dto.ManagedGroupItemDto, 0),
	}}
	handler := NewGroupHandler(service)
	recorder, ctx := newManagedGroupsHandlerContext(t)
	ctx.Set("userID", uint(73))

	handler.GetManagedGroups(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if string(payload["admin"]) != "[]" || string(payload["moderator"]) != "[]" {
		t.Fatalf("response = %s, want admin:[] and moderator:[]", recorder.Body.String())
	}
}

func TestGroupHandlerGetManagedGroupsRejectsMissingCurrentUser(t *testing.T) {
	service := &managedGroupsHandlerServiceStub{}
	handler := NewGroupHandler(service)
	recorder, ctx := newManagedGroupsHandlerContext(t)

	handler.GetManagedGroups(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
}

func TestGroupHandlerGetManagedGroupsMapsServiceFailureToInternalError(t *testing.T) {
	service := &managedGroupsHandlerServiceStub{err: errors.New("managed groups unavailable")}
	handler := NewGroupHandler(service)
	recorder, ctx := newManagedGroupsHandlerContext(t)
	ctx.Set("userID", uint(73))

	handler.GetManagedGroups(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
}

func newManagedGroupsHandlerContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v2/users/me/groups/managed", nil)
	return recorder, ctx
}

func assertManagedGroupItemJSONContract(t *testing.T, raw json.RawMessage, want map[string]interface{}) {
	t.Helper()

	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("decode managed group items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %#v, want one item", items)
	}
	item := items[0]
	wantKeys := []string{"id", "name", "categories", "smallDescription", "memberCount", "image"}
	if len(item) != len(wantKeys) {
		t.Fatalf("item keys = %#v, want exactly %v", item, wantKeys)
	}
	for _, key := range wantKeys {
		if _, ok := item[key]; !ok {
			t.Fatalf("item is missing %q: %#v", key, item)
		}
	}
	if _, ok := item["small_description"]; ok {
		t.Fatalf("legacy small_description key leaked: %#v", item)
	}
	if _, ok := item["category"]; ok {
		t.Fatalf("legacy category key leaked: %#v", item)
	}
	if _, ok := item["member_count"]; ok {
		t.Fatalf("legacy member_count key leaked: %#v", item)
	}
	for key, wantValue := range want {
		gotBytes, err := json.Marshal(item[key])
		if err != nil {
			t.Fatalf("marshal item[%q]: %v", key, err)
		}
		wantBytes, err := json.Marshal(wantValue)
		if err != nil {
			t.Fatalf("marshal expected item[%q]: %v", key, err)
		}
		if string(gotBytes) != string(wantBytes) {
			t.Fatalf("item[%q] = %s, want %s", key, gotBytes, wantBytes)
		}
	}
}
