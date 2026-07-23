package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/models/dto"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

type eventReadHandlerContextKey struct{}

type eventReadHandlerStub struct {
	searchResult  *dto.EventSearchResponse
	searchErr     error
	groupResult   []dto.EventShortDto
	groupErr      error
	detailsResult *dto.EventFullDto
	detailsErr    error

	searchCalls  int
	groupCalls   int
	detailsCalls int

	lastContext context.Context
	lastUserID  uint
	lastGroupID uint
	lastEventID uint
	lastSearch  events.EventSearchInput
}

func (s *eventReadHandlerStub) SearchEvents(ctx context.Context, userID uint, input events.EventSearchInput) (*dto.EventSearchResponse, error) {
	s.searchCalls++
	s.lastContext = ctx
	s.lastUserID = userID
	s.lastSearch = input
	return s.searchResult, s.searchErr
}

func (s *eventReadHandlerStub) GetGroupEvents(ctx context.Context, userID uint, groupID uint) ([]dto.EventShortDto, error) {
	s.groupCalls++
	s.lastContext = ctx
	s.lastUserID = userID
	s.lastGroupID = groupID
	return s.groupResult, s.groupErr
}

func (s *eventReadHandlerStub) GetEventDetails(ctx context.Context, userID uint, eventID uint) (*dto.EventFullDto, error) {
	s.detailsCalls++
	s.lastContext = ctx
	s.lastUserID = userID
	s.lastEventID = eventID
	return s.detailsResult, s.detailsErr
}

func TestSearchEventsUsesReadServiceAndReturnsResult(t *testing.T) {
	stub := &eventReadHandlerStub{
		searchResult: &dto.EventSearchResponse{
			Items: []dto.EventSearchItemDto{
				{ID: 17, Title: "Настольные игры", Subscribed: true},
			},
			Total:       1,
			Limit:       5,
			CurrentPage: 2,
			TotalPages:  1,
			HasMore:     false,
		},
	}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(
		t,
		http.MethodGet,
		"/api/v2/events/search?q=board&page=2&limit=5&hasFreeSlots=true",
	)

	handler.SearchEvents(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.searchCalls != 1 {
		t.Fatalf("search calls = %d, want 1", stub.searchCalls)
	}
	if stub.lastUserID != 9 {
		t.Fatalf("user ID = %d, want 9", stub.lastUserID)
	}
	if stub.lastContext.Value(eventReadHandlerContextKey{}) != "request-context" {
		t.Fatal("контекст запроса не передан в read service")
	}
	if stub.lastSearch.Query != "board" || stub.lastSearch.Page != 2 || stub.lastSearch.Limit != 5 {
		t.Fatalf("search input = %#v, want query=board page=2 limit=5", stub.lastSearch)
	}
	if stub.lastSearch.HasFreeSlots == nil || !*stub.lastSearch.HasFreeSlots {
		t.Fatalf("HasFreeSlots = %#v, want true", stub.lastSearch.HasFreeSlots)
	}

	var response dto.EventSearchResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Total != 1 || len(response.Items) != 1 || response.Items[0].ID != 17 || !response.Items[0].Subscribed {
		t.Fatalf("response = %#v, want configured search result", response)
	}
}

func TestSearchEventsRejectsInvalidQueryWithoutReadCall(t *testing.T) {
	stub := &eventReadHandlerStub{}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/events/search?page=0")

	handler.SearchEvents(ctx)

	assertEventReadHandlerError(t, recorder, http.StatusBadRequest, "Параметр page должен быть больше 0")
	if stub.searchCalls != 0 {
		t.Fatalf("search calls = %d, want 0", stub.searchCalls)
	}
}

func TestSearchEventsMapsReadErrorToInternalServerError(t *testing.T) {
	stub := &eventReadHandlerStub{searchErr: errors.New("ошибка поиска")}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/events/search")

	handler.SearchEvents(ctx)

	assertEventReadHandlerError(t, recorder, http.StatusInternalServerError, "ошибка поиска")
}

func TestGetGroupEventsUsesReadServiceAndReturnsResult(t *testing.T) {
	stub := &eventReadHandlerStub{
		groupResult: []dto.EventShortDto{
			{ID: 23, EventID: 23, GroupID: 17, Title: "Встреча", Subscribed: true},
		},
	}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/groups/events/17/events")
	ctx.Params = gin.Params{{Key: "groupId", Value: "17"}}

	handler.GetGroupEvents(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.groupCalls != 1 || stub.lastUserID != 9 || stub.lastGroupID != 17 {
		t.Fatalf("delegation = calls:%d user:%d group:%d, want 1/9/17", stub.groupCalls, stub.lastUserID, stub.lastGroupID)
	}
	if stub.lastContext.Value(eventReadHandlerContextKey{}) != "request-context" {
		t.Fatal("контекст запроса не передан в read service")
	}

	var response []dto.EventShortDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response) != 1 || response[0].ID != 23 || !response[0].Subscribed {
		t.Fatalf("response = %#v, want configured group event", response)
	}
}

func TestGetGroupEventsRejectsInvalidIDWithoutReadCall(t *testing.T) {
	stub := &eventReadHandlerStub{}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/groups/events/not-a-number/events")
	ctx.Params = gin.Params{{Key: "groupId", Value: "not-a-number"}}

	handler.GetGroupEvents(ctx)

	assertEventReadHandlerError(t, recorder, http.StatusBadRequest, "Некорректный ID группы")
	if stub.groupCalls != 0 {
		t.Fatalf("group calls = %d, want 0", stub.groupCalls)
	}
}

func TestGetGroupEventsMapsReadErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "недостаточно прав",
			err:         events.ErrPermissionDenied,
			wantStatus:  http.StatusForbidden,
			wantMessage: "Недостаточно прав",
		},
		{
			name:        "не участник приватной группы",
			err:         events.ErrNotGroupMember,
			wantStatus:  http.StatusForbidden,
			wantMessage: "Вы не состоите в группе",
		},
		{
			name:        "ошибка хранилища",
			err:         errors.New("ошибка чтения группы"),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "ошибка чтения группы",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &eventReadHandlerStub{groupErr: tt.err}
			handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
			recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/groups/events/17/events")
			ctx.Params = gin.Params{{Key: "groupId", Value: "17"}}

			handler.GetGroupEvents(ctx)

			assertEventReadHandlerError(t, recorder, tt.wantStatus, tt.wantMessage)
		})
	}
}

func TestGetEventDetailsUsesReadServiceAndReturnsResult(t *testing.T) {
	stub := &eventReadHandlerStub{
		detailsResult: &dto.EventFullDto{
			ID:           31,
			Title:        "Подробности события",
			AgeLimit:     "18+",
			Subscribed:   true,
			IsCreator:    false,
			CustomFields: map[string]interface{}{"format": "table"},
		},
	}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/events/31")
	ctx.Params = gin.Params{{Key: "eventId", Value: "31"}}

	handler.GetEventDetails(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.detailsCalls != 1 || stub.lastUserID != 9 || stub.lastEventID != 31 {
		t.Fatalf("delegation = calls:%d user:%d event:%d, want 1/9/31", stub.detailsCalls, stub.lastUserID, stub.lastEventID)
	}
	if stub.lastContext.Value(eventReadHandlerContextKey{}) != "request-context" {
		t.Fatal("контекст запроса не передан в read service")
	}

	var response dto.EventFullDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != 31 || response.AgeLimit != "18+" || !response.Subscribed || response.IsCreator {
		t.Fatalf("response = %#v, want configured details", response)
	}
}

func TestGetEventDetailsRejectsInvalidIDWithoutReadCall(t *testing.T) {
	stub := &eventReadHandlerStub{}
	handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
	recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/events/not-a-number")
	ctx.Params = gin.Params{{Key: "eventId", Value: "not-a-number"}}

	handler.GetEventDetails(ctx)

	assertEventReadHandlerError(t, recorder, http.StatusBadRequest, "Некорректный ID события")
	if stub.detailsCalls != 0 {
		t.Fatalf("details calls = %d, want 0", stub.detailsCalls)
	}
}

func TestGetEventDetailsMapsReadErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "событие не найдено",
			err:         events.ErrEventNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "Событие не найдено",
		},
		{
			name:        "не участник приватной группы",
			err:         events.ErrNotGroupMember,
			wantStatus:  http.StatusForbidden,
			wantMessage: "Вы не состоите в группе этого события",
		},
		{
			name:        "ошибка хранилища",
			err:         errors.New("ошибка чтения события"),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "ошибка чтения события",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &eventReadHandlerStub{detailsErr: tt.err}
			handler := NewEventsHandler(EventsHandlerDependencies{Reads: stub})
			recorder, ctx := newEventReadHandlerContext(t, http.MethodGet, "/api/v2/events/31")
			ctx.Params = gin.Params{{Key: "eventId", Value: "31"}}

			handler.GetEventDetails(ctx)

			assertEventReadHandlerError(t, recorder, tt.wantStatus, tt.wantMessage)
		})
	}
}

func newEventReadHandlerContext(t *testing.T, method string, target string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	request := httptest.NewRequest(method, target, nil)
	request = request.WithContext(context.WithValue(request.Context(), eventReadHandlerContextKey{}, "request-context"))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	ctx.Set("userID", uint(9))
	return recorder, ctx
}

func assertEventReadHandlerError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error != wantMessage || response.Message != wantMessage {
		t.Fatalf("response = %#v, want error and message %q", response, wantMessage)
	}
}

var _ events.EventReadService = (*eventReadHandlerStub)(nil)
