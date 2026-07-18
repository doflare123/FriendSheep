package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/models/dto"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

type eventCommandHandlerStub struct {
	createErr error
	updateErr error
	deleteErr error
}

func (s *eventCommandHandlerStub) CreateEvent(context.Context, uint, events.CreateEventInput) (*dto.EventFullDto, error) {
	return nil, s.createErr
}

func (s *eventCommandHandlerStub) UpdateEvent(context.Context, uint, uint, events.UpdateEventInput) (*dto.EventFullDto, error) {
	return nil, s.updateErr
}

func (s *eventCommandHandlerStub) DeleteEvent(context.Context, uint, uint) (bool, error) {
	return false, s.deleteErr
}

func TestCreateEventMapsCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "пользователь не состоит в группе",
			err:         events.ErrNotGroupMember,
			wantStatus:  http.StatusForbidden,
			wantMessage: "Вы не состоите в группе",
		},
		{
			name:        "неверные жанры",
			err:         fmt.Errorf("%w: жанр не найден", events.ErrInvalidGenres),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "некорректные жанры: жанр не найден",
		},
		{
			name:        "возрастное ограничение не найдено",
			err:         events.ErrAgeLimitNotFound,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Возрастное ограничение не найдено",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEventsHandler(EventsHandlerDependencies{
				Commands: &eventCommandHandlerStub{createErr: tt.err},
			})
			recorder, ctx := newEventCommandHandlerContext(t, http.MethodPost, "/api/v2/admin/events", validCreateEventHandlerBody())

			handler.CreateEvent(ctx)

			assertEventCommandHandlerError(t, recorder, tt.wantStatus, tt.wantMessage)
		})
	}
}

func TestUpdateEventMapsCommandErrors(t *testing.T) {
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
			name:        "пользователь не состоит в группе",
			err:         events.ErrNotGroupMember,
			wantStatus:  http.StatusForbidden,
			wantMessage: "Вы не состоите в группе",
		},
		{
			name:        "неверные жанры",
			err:         fmt.Errorf("%w: жанр не найден", events.ErrInvalidGenres),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "некорректные жанры: жанр не найден",
		},
		{
			name:        "возрастное ограничение не найдено",
			err:         events.ErrAgeLimitNotFound,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Возрастное ограничение не найдено",
		},
		{
			name:        "максимум меньше числа участников",
			err:         fmt.Errorf("%w (4)", events.ErrMaxUsersBelowCurrent),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "нельзя установить максимум меньше текущего количества участников (4)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEventsHandler(EventsHandlerDependencies{
				Commands: &eventCommandHandlerStub{updateErr: tt.err},
			})
			recorder, ctx := newEventCommandHandlerContext(t, http.MethodPut, "/api/v2/admin/events/17", []byte(`{}`))
			ctx.Params = gin.Params{{Key: "eventId", Value: "17"}}

			handler.UpdateEvent(ctx)

			assertEventCommandHandlerError(t, recorder, tt.wantStatus, tt.wantMessage)
		})
	}
}

func TestDeleteEventMapsCommandErrors(t *testing.T) {
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
			name:        "пользователь не состоит в группе",
			err:         events.ErrNotGroupMember,
			wantStatus:  http.StatusForbidden,
			wantMessage: "Вы не состоите в группе",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEventsHandler(EventsHandlerDependencies{
				Commands: &eventCommandHandlerStub{deleteErr: tt.err},
			})
			recorder, ctx := newEventCommandHandlerContext(t, http.MethodDelete, "/api/v2/admin/events/17", nil)
			ctx.Params = gin.Params{{Key: "eventId", Value: "17"}}

			handler.DeleteEvent(ctx)

			assertEventCommandHandlerError(t, recorder, tt.wantStatus, tt.wantMessage)
		})
	}
}

func newEventCommandHandlerContext(t *testing.T, method string, target string, body []byte) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("userID", uint(9))
	return recorder, ctx
}

func validCreateEventHandlerBody() []byte {
	return []byte(`{
		"title":"Новая встреча",
		"description":"Подробное описание нового события",
		"groupId":7,
		"eventTypeId":1,
		"locationId":1,
		"imageUrl":"https://example.com/event.png",
		"startTime":"2030-01-01T12:00:00Z",
		"duration":60,
		"maxUsers":10,
		"genres":[1],
		"ageLimit":1
	}`)
}

func assertEventCommandHandlerError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
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

var _ events.EventCommandService = (*eventCommandHandlerStub)(nil)
