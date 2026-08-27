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
	"friendship/models/dto"
	"friendship/routes"
	"friendship/services/events"

	"github.com/gin-gonic/gin"
)

type lifecycleScheduleServiceStub struct {
	page  events.EventLifecycleScheduleEventsPage
	err   error
	ctx   context.Context
	after int64
	limit int
	calls int
}

func (s *lifecycleScheduleServiceStub) List(
	ctx context.Context,
	after int64,
	limit int,
) (events.EventLifecycleScheduleEventsPage, error) {
	s.calls++
	s.ctx = ctx
	s.after = after
	s.limit = limit
	return s.page, s.err
}

func TestLifecycleScheduleOutboxHTTPRequiresInternalToken(t *testing.T) {
	t.Parallel()

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
		{name: "user JWT is insufficient", authorization: "Bearer valid-user-jwt", configured: token, wantStatus: http.StatusUnauthorized},
		{name: "empty configuration fails closed", internalToken: token, wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &lifecycleScheduleServiceStub{page: events.EventLifecycleScheduleEventsPage{Items: []events.EventLifecycleScheduleEvent{}}}
			router := newLifecycleScheduleTestRouter(stub, test.configured)
			request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-lifecycle/schedule-events", nil)
			request.Header.Set(middlewares.InternalTokenHeader, test.internalToken)
			request.Header.Set("Authorization", test.authorization)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

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

func TestLifecycleScheduleOutboxHTTPPreservesContextCursorAndStableDTO(t *testing.T) {
	t.Parallel()

	const token = "notify-service-secret-value"
	type contextKey struct{}
	start := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	stub := &lifecycleScheduleServiceStub{page: events.EventLifecycleScheduleEventsPage{
		Items: []events.EventLifecycleScheduleEvent{{
			Sequence:      123,
			MessageID:     "00000000-0000-0000-0000-000000000123",
			SchemaVersion: 1,
			Operation:     events.LifecycleScheduleOperationUpsert,
			EventID:       42,
			StartTime:     &start,
			EndTime:       scheduleHandlerTimePointer(start.Add(3 * time.Hour)),
			OccurredAt:    start.Add(-time.Hour),
		}},
		NextCursor: 123,
		HasMore:    false,
	}}
	router := newLifecycleScheduleTestRouter(stub, token)
	request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-lifecycle/schedule-events?after=122&limit=50", nil)
	request = request.WithContext(context.WithValue(request.Context(), contextKey{}, "preserved"))
	request.Header.Set(middlewares.InternalTokenHeader, token)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if stub.calls != 1 || stub.after != 122 || stub.limit != 50 || stub.ctx.Value(contextKey{}) != "preserved" {
		t.Fatalf("service call = count:%d after:%d limit:%d context:%v",
			stub.calls, stub.after, stub.limit, stub.ctx.Value(contextKey{}))
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(raw) != 3 || raw["items"] == nil || raw["nextCursor"] == nil || raw["hasMore"] == nil {
		t.Fatalf("top-level response shape = %s", response.Body.String())
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw["items"], &items); err != nil || len(items) != 1 {
		t.Fatalf("decode items: len=%d err=%v", len(items), err)
	}
	wantFields := []string{"sequence", "messageId", "schemaVersion", "operation", "eventId", "startTime", "endTime", "occurredAt"}
	if len(items[0]) != len(wantFields) {
		t.Fatalf("item fields = %v, want exactly %v", scheduleMapKeys(items[0]), wantFields)
	}
	for _, field := range wantFields {
		if items[0][field] == nil {
			t.Errorf("stable DTO is missing field %q: %s", field, response.Body.String())
		}
	}
}

func TestLifecycleScheduleOutboxHTTPCancelOmitsScheduleTimes(t *testing.T) {
	t.Parallel()

	const token = "notify-service-secret-value"
	stub := &lifecycleScheduleServiceStub{page: events.EventLifecycleScheduleEventsPage{
		Items: []events.EventLifecycleScheduleEvent{{
			Sequence:      124,
			MessageID:     "00000000-0000-0000-0000-000000000124",
			SchemaVersion: 1,
			Operation:     events.LifecycleScheduleOperationCancel,
			EventID:       42,
			OccurredAt:    time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC),
		}},
		NextCursor: 124,
	}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-lifecycle/schedule-events", nil)
	request.Header.Set(middlewares.InternalTokenHeader, token)
	newLifecycleScheduleTestRouter(stub, token).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "startTime") || strings.Contains(response.Body.String(), "endTime") {
		t.Fatalf("cancel DTO contains schedule times: %s", response.Body.String())
	}
}

func TestLifecycleScheduleOutboxHTTPRejectsInvalidPaginationAndSanitizesErrors(t *testing.T) {
	t.Parallel()

	const token = "notify-service-secret-value"
	tests := []struct {
		name       string
		query      string
		serviceErr error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{name: "negative after", query: "?after=-1", wantStatus: http.StatusBadRequest, wantCode: "invalid_after"},
		{name: "non-numeric after", query: "?after=nope", wantStatus: http.StatusBadRequest, wantCode: "invalid_after"},
		{name: "zero limit", query: "?limit=0", wantStatus: http.StatusBadRequest, wantCode: "invalid_limit"},
		{name: "limit over maximum", query: "?limit=501", wantStatus: http.StatusBadRequest, wantCode: "invalid_limit"},
		{name: "internal error", serviceErr: errors.New("database password leaked"), wantStatus: http.StatusInternalServerError, wantCode: "event_lifecycle_schedule_read_failed", wantCalls: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &lifecycleScheduleServiceStub{err: test.serviceErr}
			request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-lifecycle/schedule-events"+test.query, nil)
			request.Header.Set(middlewares.InternalTokenHeader, token)
			response := httptest.NewRecorder()

			newLifecycleScheduleTestRouter(stub, token).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if stub.calls != test.wantCalls {
				t.Fatalf("service calls = %d, want %d", stub.calls, test.wantCalls)
			}
			assertScheduleErrorDTO(t, response, test.wantCode)
			if strings.Contains(response.Body.String(), "database password leaked") {
				t.Fatalf("response leaked implementation details: %s", response.Body.String())
			}
		})
	}
}

func newLifecycleScheduleTestRouter(service events.EventLifecycleScheduleOutboxService, token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.RegisterInternalEventLifecycleScheduleRoutes(
		router,
		handlers.NewEventLifecycleScheduleOutboxHandler(service),
		middlewares.NewInternalTokenMiddleware(token),
	)
	return router
}

func assertScheduleErrorDTO(t *testing.T, response *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var body dto.ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error DTO: %v; body = %s", err, response.Body.String())
	}
	if body.Error != wantCode || body.Message == "" {
		t.Fatalf("error DTO = %#v, want code %q and non-empty message", body, wantCode)
	}
}

func scheduleHandlerTimePointer(value time.Time) *time.Time {
	return &value
}

func scheduleMapKeys(values map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

var _ events.EventLifecycleScheduleOutboxService = (*lifecycleScheduleServiceStub)(nil)
