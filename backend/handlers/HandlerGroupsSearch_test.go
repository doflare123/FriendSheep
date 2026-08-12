package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"friendship/models/dto"
	group "friendship/services/groups"

	"github.com/gin-gonic/gin"
)

type groupSearchHandlerContextKey struct{}

type groupSearchHandlerServiceStub struct {
	group.GroupsService
	result *dto.GroupSearchResponseDto
	err    error
	calls  int
	ctx    context.Context
	userID uint
	input  group.GroupSearchInput
}

func (s *groupSearchHandlerServiceStub) SearchGroups(ctx context.Context, userID uint, input group.GroupSearchInput) (*dto.GroupSearchResponseDto, error) {
	s.calls++
	s.ctx = ctx
	s.userID = userID
	s.input = input
	return s.result, s.err
}

func TestGroupHandlerSearchGroupsParsesFiltersAndReturnsExactJSON(t *testing.T) {
	createdAt := time.Date(2028, 1, 2, 15, 4, 5, 0, time.UTC)
	stub := &groupSearchHandlerServiceStub{result: &dto.GroupSearchResponseDto{
		Items: []dto.GroupSearchItemDto{{
			ID:               17,
			Name:             "Board Club",
			Categories:       []string{"Board Games", "Travel"},
			MemberCount:      9,
			Image:            "https://example.com/group.png",
			CreatedAt:        createdAt,
			SmallDescription: "Play together",
			IsPrivate:        false,
			Enterprise:       true,
			IsSubscribed:     true,
		}},
		Total: 5, Page: 2, Limit: 2, TotalPages: 3, HasMore: true,
	}}
	handler := NewGroupHandler(stub)
	recorder, ctx := newGroupSearchHandlerContext(t, "/api/v2/groups/search?q=%20board%20&categoryIds=1,2&categoryIds=3&isPrivate=false&city=%20Moscow%20&sortBy=memberCount&sortOrder=ASC&page=2&limit=2")
	requestCtx := context.WithValue(ctx.Request.Context(), groupSearchHandlerContextKey{}, "search-request")
	ctx.Request = ctx.Request.WithContext(requestCtx)
	ctx.Set("userID", uint(73))

	handler.SearchGroups(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.calls != 1 || stub.ctx != requestCtx || stub.ctx.Value(groupSearchHandlerContextKey{}) != "search-request" || stub.userID != 73 {
		t.Fatalf("service call = count:%d context:%v userID:%d, want exact request context and authenticated user once", stub.calls, stub.ctx, stub.userID)
	}
	if stub.input.Query != "board" || !reflect.DeepEqual(stub.input.CategoryIDs, []uint{1, 2, 3}) ||
		stub.input.IsPrivate == nil || *stub.input.IsPrivate || stub.input.City != "Moscow" ||
		stub.input.SortBy != "memberCount" || stub.input.SortOrder != "asc" || stub.input.Page != 2 || stub.input.Limit != 2 {
		t.Fatalf("service input = %#v, want parsed exact query", stub.input)
	}
	assertExactGroupSearchJSON(t, recorder.Body.Bytes(), createdAt, true)
}

func TestGroupHandlerSearchGroupsUsesDefaults(t *testing.T) {
	stub := &groupSearchHandlerServiceStub{result: &dto.GroupSearchResponseDto{
		Items: make([]dto.GroupSearchItemDto, 0),
		Page:  group.DefaultGroupSearchPage, Limit: group.DefaultGroupSearchLimit,
	}}
	handler := NewGroupHandler(stub)
	recorder, ctx := newGroupSearchHandlerContext(t, "/api/v2/groups/search")

	handler.SearchGroups(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.calls != 1 || stub.input.Page != group.DefaultGroupSearchPage || stub.input.Limit != group.DefaultGroupSearchLimit ||
		stub.input.SortBy != "" || stub.input.SortOrder != "" || stub.input.IsPrivate != nil || stub.userID != 0 {
		t.Fatalf("default input = %#v, want pagination defaults, service-owned sort defaults, and no privacy filter", stub.input)
	}
}

func TestGroupHandlerSearchGroupsRejectsMalformedQueriesWithoutServiceCall(t *testing.T) {
	manyIDs := make([]string, 101)
	for i := range manyIDs {
		manyIDs[i] = strconv.Itoa(i + 1)
	}
	tests := []string{
		"categoryIds=abc",
		"categoryIds=0",
		"categoryIds=1&categoryIds=1",
		"categoryIds=1,,2",
		"categoryIds=" + strings.Join(manyIDs, ","),
		"isPrivate=maybe",
		"isPrivate=1",
		"isPrivate=true&isPrivate=false",
		"page=bad",
		"page=0",
		"page=-1",
		"limit=bad",
		"limit=0",
		"limit=101",
		"sortBy=popular",
		"sortOrder=sideways",
	}

	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			stub := &groupSearchHandlerServiceStub{}
			handler := NewGroupHandler(stub)
			recorder, ctx := newGroupSearchHandlerContext(t, "/api/v2/groups/search?"+query)

			handler.SearchGroups(ctx)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if stub.calls != 0 {
				t.Fatalf("service calls = %d, want 0", stub.calls)
			}
			assertGroupSearchErrorShape(t, recorder)
		})
	}
}

func TestGroupHandlerSearchGroupsMapsServiceValidationToBadRequest(t *testing.T) {
	stub := &groupSearchHandlerServiceStub{err: group.ErrInvalidGroupSearchInput}
	handler := NewGroupHandler(stub)
	recorder, ctx := newGroupSearchHandlerContext(t, "/api/v2/groups/search")

	handler.SearchGroups(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	assertGroupSearchErrorShape(t, recorder)
}

func TestGroupHandlerSearchGroupsDoesNotExposeInternalError(t *testing.T) {
	const internalDetails = "search groups: SQL logic error: no such table"
	stub := &groupSearchHandlerServiceStub{err: errors.New(internalDetails)}
	handler := NewGroupHandler(stub)
	recorder, ctx := newGroupSearchHandlerContext(t, "/api/v2/groups/search")

	handler.SearchGroups(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), internalDetails) || strings.Contains(recorder.Body.String(), "SQL") {
		t.Fatalf("public error response exposes internal details: %s", recorder.Body.String())
	}
	assertGroupSearchErrorShape(t, recorder)
}

func newGroupSearchHandlerContext(t *testing.T, target string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return recorder, ctx
}

func assertExactGroupSearchJSON(t *testing.T, body []byte, createdAt time.Time, isSubscribed bool) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	wantTopKeys := []string{"items", "total", "page", "limit", "totalPages", "hasMore"}
	if len(payload) != len(wantTopKeys) {
		t.Fatalf("top-level keys = %#v, want exactly %v", payload, wantTopKeys)
	}
	for _, key := range wantTopKeys {
		if _, exists := payload[key]; !exists {
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
	wantItem := map[string]interface{}{
		"id": float64(17), "name": "Board Club", "categories": []interface{}{"Board Games", "Travel"},
		"memberCount": float64(9), "image": "https://example.com/group.png", "createdAt": createdAt.Format(time.RFC3339),
		"smallDescription": "Play together", "isPrivate": false, "enterprise": true, "isSubscribed": isSubscribed,
	}
	if len(items[0]) != len(wantItem) {
		t.Fatalf("item fields = %#v, want exactly %#v", items[0], wantItem)
	}
	if !reflect.DeepEqual(items[0], wantItem) {
		t.Fatalf("item = %#v, want %#v", items[0], wantItem)
	}
}

func assertGroupSearchErrorShape(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error == "" || response.Message == "" || response.Error != response.Message {
		t.Fatalf("error response = %#v, want mirrored non-empty error and message", response)
	}
}
