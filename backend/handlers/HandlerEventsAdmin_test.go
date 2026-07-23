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

type eventAdminHandlerContextKey struct{}

type eventAdminHandlerStub struct {
	detailsResult *dto.EventAdminDto
	detailsErr    error
	kickResult    bool
	kickErr       error

	detailsCalls int
	kickCalls    int

	lastContext      context.Context
	lastActorID      uint
	lastEventID      uint
	lastTargetUserID uint
}

func (s *eventAdminHandlerStub) GetEventDetailsForAdmin(ctx context.Context, actorID uint, eventID uint) (*dto.EventAdminDto, error) {
	s.detailsCalls++
	s.lastContext = ctx
	s.lastActorID = actorID
	s.lastEventID = eventID
	return s.detailsResult, s.detailsErr
}

func (s *eventAdminHandlerStub) KickUserFromEvent(ctx context.Context, actorID uint, eventID uint, targetUserID uint) (bool, error) {
	s.kickCalls++
	s.lastContext = ctx
	s.lastActorID = actorID
	s.lastEventID = eventID
	s.lastTargetUserID = targetUserID
	return s.kickResult, s.kickErr
}

func TestGetEventDetailsForAdminUsesAdminServiceAndReturnsResult(t *testing.T) {
	stub := &eventAdminHandlerStub{
		detailsResult: &dto.EventAdminDto{
			EventFullDto: dto.EventFullDto{
				ID:         31,
				Title:      "Admin Event",
				Subscribed: true,
			},
			AllParticipants: []dto.EventAdminParticipantDto{
				{UserID: 9, IsCreator: true},
				{UserID: 14, IsCreator: false},
			},
		},
	}
	handler := NewEventsHandler(EventsHandlerDependencies{Admin: stub})
	recorder, ctx := newEventAdminHandlerContext(http.MethodGet, "/api/v2/admin/events/31")
	ctx.Params = gin.Params{{Key: "eventId", Value: "31"}}

	handler.GetEventDetailsForAdmin(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.detailsCalls != 1 || stub.lastActorID != 9 || stub.lastEventID != 31 {
		t.Fatalf("delegation = calls:%d actor:%d event:%d, want 1/9/31", stub.detailsCalls, stub.lastActorID, stub.lastEventID)
	}
	if stub.lastContext.Value(eventAdminHandlerContextKey{}) != "admin-request-context" {
		t.Fatal("request context was not forwarded to admin service")
	}

	var response dto.EventAdminDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != 31 || !response.Subscribed || len(response.AllParticipants) != 2 {
		t.Fatalf("response = %#v, want configured admin details", response)
	}
}

func TestGetEventDetailsForAdminRejectsInvalidIDWithoutAdminCall(t *testing.T) {
	stub := &eventAdminHandlerStub{}
	handler := NewEventsHandler(EventsHandlerDependencies{Admin: stub})
	recorder, ctx := newEventAdminHandlerContext(http.MethodGet, "/api/v2/admin/events/not-a-number")
	ctx.Params = gin.Params{{Key: "eventId", Value: "not-a-number"}}

	handler.GetEventDetailsForAdmin(ctx)

	assertEventAdminHandlerError(t, recorder, http.StatusBadRequest, false, "")
	if stub.detailsCalls != 0 {
		t.Fatalf("details calls = %d, want 0", stub.detailsCalls)
	}
}

func TestGetEventDetailsForAdminMapsErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantExactError bool
		wantMessage    string
	}{
		{
			name:       "event not found",
			err:        events.ErrEventNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "permission denied",
			err:        events.ErrPermissionDenied,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "not group member",
			err:        events.ErrNotGroupMember,
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "storage error",
			err:            errors.New("admin details failure"),
			wantStatus:     http.StatusInternalServerError,
			wantExactError: true,
			wantMessage:    "admin details failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &eventAdminHandlerStub{detailsErr: tt.err}
			handler := NewEventsHandler(EventsHandlerDependencies{Admin: stub})
			recorder, ctx := newEventAdminHandlerContext(http.MethodGet, "/api/v2/admin/events/31")
			ctx.Params = gin.Params{{Key: "eventId", Value: "31"}}

			handler.GetEventDetailsForAdmin(ctx)

			assertEventAdminHandlerError(t, recorder, tt.wantStatus, tt.wantExactError, tt.wantMessage)
		})
	}
}

func TestKickUserFromEventUsesAdminServiceAndReturnsResult(t *testing.T) {
	stub := &eventAdminHandlerStub{kickResult: true}
	handler := NewEventsHandler(EventsHandlerDependencies{Admin: stub})
	recorder, ctx := newEventAdminHandlerContext(http.MethodDelete, "/api/v2/admin/events/31/kick/14")
	ctx.Params = gin.Params{
		{Key: "eventId", Value: "31"},
		{Key: "userId", Value: "14"},
	}

	handler.KickUserFromEvent(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.kickCalls != 1 || stub.lastActorID != 9 || stub.lastEventID != 31 || stub.lastTargetUserID != 14 {
		t.Fatalf(
			"delegation = calls:%d actor:%d event:%d target:%d, want 1/9/31/14",
			stub.kickCalls,
			stub.lastActorID,
			stub.lastEventID,
			stub.lastTargetUserID,
		)
	}
	if stub.lastContext.Value(eventAdminHandlerContextKey{}) != "admin-request-context" {
		t.Fatal("request context was not forwarded to admin service")
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if success, ok := response["success"].(bool); !ok || !success {
		t.Fatalf("response = %#v, want success=true", response)
	}
}

func TestKickUserFromEventRejectsInvalidIDsWithoutAdminCall(t *testing.T) {
	tests := []struct {
		name    string
		eventID string
		userID  string
	}{
		{
			name:    "invalid event id",
			eventID: "not-a-number",
			userID:  "14",
		},
		{
			name:    "invalid user id",
			eventID: "31",
			userID:  "not-a-number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &eventAdminHandlerStub{}
			handler := NewEventsHandler(EventsHandlerDependencies{Admin: stub})
			recorder, ctx := newEventAdminHandlerContext(http.MethodDelete, "/api/v2/admin/events/"+tt.eventID+"/kick/"+tt.userID)
			ctx.Params = gin.Params{
				{Key: "eventId", Value: tt.eventID},
				{Key: "userId", Value: tt.userID},
			}

			handler.KickUserFromEvent(ctx)

			assertEventAdminHandlerError(t, recorder, http.StatusBadRequest, false, "")
			if stub.kickCalls != 0 {
				t.Fatalf("kick calls = %d, want 0", stub.kickCalls)
			}
		})
	}
}

func TestKickUserFromEventMapsErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantExactError bool
		wantMessage    string
	}{
		{
			name:       "event not found",
			err:        events.ErrEventNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "user not found",
			err:        events.ErrUserNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "permission denied",
			err:        events.ErrPermissionDenied,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "not group member",
			err:        events.ErrNotGroupMember,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "creator cannot be kicked",
			err:        events.ErrCreatorCantLeave,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "self kick",
			err:        events.ErrActorCantKickSelf,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not joined",
			err:        events.ErrNotJoined,
			wantStatus: http.StatusNotFound,
		},
		{
			name:           "storage error",
			err:            errors.New("admin kick failure"),
			wantStatus:     http.StatusInternalServerError,
			wantExactError: true,
			wantMessage:    "admin kick failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &eventAdminHandlerStub{kickErr: tt.err}
			handler := NewEventsHandler(EventsHandlerDependencies{Admin: stub})
			recorder, ctx := newEventAdminHandlerContext(http.MethodDelete, "/api/v2/admin/events/31/kick/14")
			ctx.Params = gin.Params{
				{Key: "eventId", Value: "31"},
				{Key: "userId", Value: "14"},
			}

			handler.KickUserFromEvent(ctx)

			assertEventAdminHandlerError(t, recorder, tt.wantStatus, tt.wantExactError, tt.wantMessage)
		})
	}
}

func newEventAdminHandlerContext(method string, target string) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)

	request := httptest.NewRequest(method, target, nil)
	request = request.WithContext(context.WithValue(request.Context(), eventAdminHandlerContextKey{}, "admin-request-context"))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	ctx.Set("userID", uint(9))
	return recorder, ctx
}

func assertEventAdminHandlerError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantExactError bool, wantMessage string) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error == "" || response.Message == "" {
		t.Fatalf("response = %#v, want non-empty error payload", response)
	}
	if response.Error != response.Message {
		t.Fatalf("response = %#v, want mirrored error/message fields", response)
	}
	if wantExactError && response.Error != wantMessage {
		t.Fatalf("response = %#v, want error %q", response, wantMessage)
	}
}

var _ events.EventAdminService = (*eventAdminHandlerStub)(nil)
