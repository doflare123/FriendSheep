package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"friendship/models/dto"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

type popularEventsHandlerContextKey struct{}

type popularEventsHandlerServiceStub struct {
	snapshot *events.PopularEventsSnapshot
	err      error
	ctx      context.Context
	calls    int
}

func (s *popularEventsHandlerServiceStub) GetPopularEvents(ctx context.Context) (*events.PopularEventsSnapshot, error) {
	s.calls++
	s.ctx = ctx
	return s.snapshot, s.err
}

func (s *popularEventsHandlerServiceStub) UpdateCache(context.Context) error {
	return nil
}

func (s *popularEventsHandlerServiceStub) Start() error {
	return nil
}

func (s *popularEventsHandlerServiceStub) Stop() {}

func TestPopularEventsHandlerPropagatesRequestContextAndReturnsSnapshot(t *testing.T) {
	updatedAt := time.Date(2036, 8, 9, 10, 11, 12, 0, time.UTC)
	stub := &popularEventsHandlerServiceStub{snapshot: &events.PopularEventsSnapshot{
		Events: []events.PopularEventView{{
			ID:           61,
			EventID:      61,
			Title:        "Popular event",
			CurrentUsers: 9,
			MaxUsers:     10,
			Genres:       []string{"Strategy"},
		}},
		UpdatedAt: updatedAt,
		Count:     1,
	}}
	handler := NewPopularEventsHandler(stub)
	recorder, ctx := newPopularEventsHandlerContext(t)

	handler.GetPopularEvents(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.calls != 1 {
		t.Fatalf("service calls = %d, want 1", stub.calls)
	}
	if stub.ctx != ctx.Request.Context() ||
		stub.ctx.Value(popularEventsHandlerContextKey{}) != "popular-handler" {
		t.Fatal("handler did not propagate the HTTP request context")
	}

	var response events.PopularEventsSnapshot
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode popular events response: %v", err)
	}
	if response.Count != 1 ||
		len(response.Events) != 1 ||
		response.Events[0].ID != 61 ||
		!response.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response = %#v, want service snapshot", response)
	}
}

func TestPopularEventsHandlerMapsServiceFailureToInternalError(t *testing.T) {
	stub := &popularEventsHandlerServiceStub{err: errors.New("cache failed")}
	handler := NewPopularEventsHandler(stub)
	recorder, ctx := newPopularEventsHandlerContext(t)

	handler.GetPopularEvents(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
	if stub.ctx != ctx.Request.Context() {
		t.Fatal("handler did not propagate request context on service failure")
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error != "internal_error" {
		t.Fatalf("error code = %q, want internal_error", response.Error)
	}
}

func newPopularEventsHandlerContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	request := httptest.NewRequest(http.MethodGet, "/api/v2/events/popular", nil)
	request = request.WithContext(context.WithValue(
		request.Context(),
		popularEventsHandlerContextKey{},
		"popular-handler",
	))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	return recorder, ctx
}

var _ events.PopularEventsService = (*popularEventsHandlerServiceStub)(nil)
