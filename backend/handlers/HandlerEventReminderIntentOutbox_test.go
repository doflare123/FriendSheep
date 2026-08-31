package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

type reminderIntentServiceStub struct {
	page  events.EventReminderIntentPage
	err   error
	ctx   context.Context
	after int64
	limit int
	calls int
}

func (s *reminderIntentServiceStub) List(ctx context.Context, after int64, limit int) (events.EventReminderIntentPage, error) {
	s.calls++
	s.ctx = ctx
	s.after = after
	s.limit = limit
	return s.page, s.err
}

func TestEventReminderIntentHTTPRequiresInternalTokenAndRejectsJWTAlone(t *testing.T) {
	const token = "notify-service-secret-value"
	tests := []struct {
		name          string
		internalToken string
		authorization string
		configured    string
		wantStatus    int
		wantCalls     int
	}{
		{name: "valid internal token", internalToken: token, configured: token, wantStatus: http.StatusOK, wantCalls: 1},
		{name: "missing token", configured: token, wantStatus: http.StatusUnauthorized},
		{name: "wrong token", internalToken: "wrong", configured: token, wantStatus: http.StatusUnauthorized},
		{name: "JWT alone is insufficient", authorization: "Bearer valid-user-jwt", configured: token, wantStatus: http.StatusUnauthorized},
		{name: "empty configuration fails closed", internalToken: token, wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &reminderIntentServiceStub{page: events.EventReminderIntentPage{Items: []events.EventReminderIntent{}}}
			request := httptest.NewRequest(http.MethodGet, "/internal/v1/notification-intents/event-reminders", nil)
			request.Header.Set(middlewares.InternalTokenHeader, test.internalToken)
			request.Header.Set("Authorization", test.authorization)
			response := httptest.NewRecorder()

			newReminderIntentTestRouter(stub, test.configured).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if stub.calls != test.wantCalls {
				t.Fatalf("service calls = %d, want %d", stub.calls, test.wantCalls)
			}
			if test.wantStatus == http.StatusUnauthorized {
				assertScheduleErrorDTO(t, response, "invalid_internal_token")
			}
		})
	}
}

func TestEventReminderIntentHTTPPreservesCursorContextAndStableDTO(t *testing.T) {
	const token = "notify-service-secret-value"
	type contextKey struct{}
	start := time.Date(2036, 8, 30, 18, 0, 0, 0, time.UTC)
	stub := &reminderIntentServiceStub{page: events.EventReminderIntentPage{
		Items: []events.EventReminderIntent{{
			Sequence: 125, MessageID: "7556d607-1df0-41df-9682-8baba85e37e7",
			SchemaVersion: 1, IntentType: events.EventReminderIntentType,
			Operation: events.EventReminderOperationUpsert, EventID: 42, StartTime: &start,
			ReminderOffsetMinutes: []int{1440, 360, 60}, OccurredAt: start.Add(-time.Hour),
		}},
		NextCursor: 125,
	}}
	request := httptest.NewRequest(http.MethodGet, "/internal/v1/notification-intents/event-reminders?after=124&limit=50", nil)
	request = request.WithContext(context.WithValue(request.Context(), contextKey{}, "preserved"))
	request.Header.Set(middlewares.InternalTokenHeader, token)
	response := httptest.NewRecorder()

	newReminderIntentTestRouter(stub, token).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if stub.calls != 1 || stub.after != 124 || stub.limit != 50 || stub.ctx.Value(contextKey{}) != "preserved" {
		t.Fatalf("service call = count:%d after:%d limit:%d context:%v", stub.calls, stub.after, stub.limit, stub.ctx.Value(contextKey{}))
	}
	var body struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || len(body.Items) != 1 {
		t.Fatalf("decode response: items=%d err=%v body=%s", len(body.Items), err, response.Body.String())
	}
	wantFields := []string{"sequence", "messageId", "schemaVersion", "intentType", "operation", "eventId", "startTime", "reminderOffsetMinutes", "occurredAt"}
	if len(body.Items[0]) != len(wantFields) {
		t.Fatalf("item fields = %v, want exactly %v", scheduleMapKeys(body.Items[0]), wantFields)
	}
	for _, field := range wantFields {
		if body.Items[0][field] == nil {
			t.Errorf("DTO missing field %q: %s", field, response.Body.String())
		}
	}
}

func TestEventReminderIntentHTTPCancelOmitsScheduleAndSanitizesErrors(t *testing.T) {
	const token = "notify-service-secret-value"
	t.Run("cancel tombstone", func(t *testing.T) {
		stub := &reminderIntentServiceStub{page: events.EventReminderIntentPage{Items: []events.EventReminderIntent{{
			Sequence: 126, MessageID: "84bea6b7-ea88-4ece-bcd6-49538e9d0e58", SchemaVersion: 1,
			IntentType: events.EventReminderIntentType, Operation: events.EventReminderOperationCancel,
			EventID: 42, OccurredAt: time.Date(2036, 8, 30, 13, 0, 0, 0, time.UTC),
		}}}}
		request := httptest.NewRequest(http.MethodGet, "/internal/v1/notification-intents/event-reminders", nil)
		request.Header.Set(middlewares.InternalTokenHeader, token)
		response := httptest.NewRecorder()
		newReminderIntentTestRouter(stub, token).ServeHTTP(response, request)
		if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "startTime") || strings.Contains(response.Body.String(), "reminderOffsetMinutes") {
			t.Fatalf("cancel response = status:%d body:%s", response.Code, response.Body.String())
		}
	})

	for _, test := range []struct {
		name       string
		query      string
		serviceErr error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{name: "negative after", query: "?after=-1", wantStatus: http.StatusBadRequest, wantCode: "invalid_after"},
		{name: "invalid limit", query: "?limit=501", wantStatus: http.StatusBadRequest, wantCode: "invalid_limit"},
		{name: "service failure", serviceErr: errors.New("database password leaked"), wantStatus: http.StatusInternalServerError, wantCode: "notification_intent_read_failed", wantCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &reminderIntentServiceStub{err: test.serviceErr}
			request := httptest.NewRequest(http.MethodGet, "/internal/v1/notification-intents/event-reminders"+test.query, nil)
			request.Header.Set(middlewares.InternalTokenHeader, token)
			response := httptest.NewRecorder()
			newReminderIntentTestRouter(stub, token).ServeHTTP(response, request)
			if response.Code != test.wantStatus || stub.calls != test.wantCalls {
				t.Fatalf("response = status:%d calls:%d body:%s, want status:%d calls:%d", response.Code, stub.calls, response.Body.String(), test.wantStatus, test.wantCalls)
			}
			assertScheduleErrorDTO(t, response, test.wantCode)
			if strings.Contains(response.Body.String(), "database password leaked") {
				t.Fatalf("response leaked internal error: %s", response.Body.String())
			}
		})
	}
}

func newReminderIntentTestRouter(service events.EventReminderIntentService, token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.RegisterInternalEventReminderIntentRoutes(router, handlers.NewEventReminderIntentOutboxHandler(service), middlewares.NewInternalTokenMiddleware(token))
	return router
}

var _ events.EventReminderIntentService = (*reminderIntentServiceStub)(nil)
