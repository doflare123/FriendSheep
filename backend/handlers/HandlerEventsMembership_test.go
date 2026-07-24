package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/models/dto"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

type eventMembershipHandlerStub struct {
	joinResult  bool
	joinErr     error
	joinCalls   int
	lastUserID  uint
	lastEventID uint
}

func (s *eventMembershipHandlerStub) JoinEvent(_ context.Context, userID uint, eventID uint) (bool, error) {
	s.joinCalls++
	s.lastUserID = userID
	s.lastEventID = eventID
	return s.joinResult, s.joinErr
}

func (s *eventMembershipHandlerStub) LeaveEvent(context.Context, uint, uint) (bool, error) {
	return false, nil
}

func TestJoinEventMapsAlreadyStartedErrorToBadRequest(t *testing.T) {
	stub := &eventMembershipHandlerStub{joinErr: events.ErrEventAlreadyStarted}
	handler := NewEventsHandler(EventsHandlerDependencies{Membership: stub})
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v2/events/17/join", nil)
	ctx.Params = gin.Params{{Key: "eventId", Value: "17"}}
	ctx.Set("userID", uint(9))

	handler.JoinEvent(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if stub.joinCalls != 1 || stub.lastUserID != 9 || stub.lastEventID != 17 {
		t.Fatalf(
			"delegation = calls:%d user:%d event:%d, want 1/9/17",
			stub.joinCalls,
			stub.lastUserID,
			stub.lastEventID,
		)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error != events.ErrEventAlreadyStarted.Error() ||
		response.Message != events.ErrEventAlreadyStarted.Error() {
		t.Fatalf("response = %#v, want ErrEventAlreadyStarted message", response)
	}
}

var _ events.EventMembershipService = (*eventMembershipHandlerStub)(nil)
