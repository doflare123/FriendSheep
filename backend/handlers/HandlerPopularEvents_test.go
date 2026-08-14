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
			ID:    61,
			Title: "Popular event",
			Group: events.PopularEventGroupView{
				ID:         7,
				Name:       "Strategy Club",
				Image:      "https://example.com/group.png",
				Enterprise: true,
			},
			Image:        "https://example.com/popular.png",
			CurrentUsers: 9,
			MaxUsers:     10,
			Duration:     120,
			StartTime:    time.Date(2036, 8, 10, 21, 22, 23, 0, time.UTC),
			EventType:    "Game",
			LocationType: "Offline",
			City:         "Kaliningrad",
			Genres:       []string{"Strategy"},
			AgeLimit:     "18+",
			Status:       "Recruitment",
			Subscribed:   true,
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

	var response dto.CachedPopularEvents
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode popular events response: %v", err)
	}
	if response.Count != 1 ||
		len(response.Events) != 1 ||
		response.Events[0].ID != 61 ||
		!response.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response = %#v, want service snapshot", response)
	}
	item := response.Events[0]
	if item.Group.ID != 7 || item.Group.Name != "Strategy Club" ||
		item.Group.Image != "https://example.com/group.png" || !item.Group.Enterprise ||
		item.Image != "https://example.com/popular.png" ||
		!item.StartTime.Equal(time.Date(2036, 8, 10, 21, 22, 23, 0, time.UTC)) ||
		item.EventType != "Game" || item.LocationType != "Offline" ||
		item.City != "Kaliningrad" || item.AgeLimit != "18+" ||
		item.Status != "Recruitment" || !item.Subscribed {
		t.Fatalf("response item = %#v, want EventSearchItemDto fields including subscribed", item)
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
