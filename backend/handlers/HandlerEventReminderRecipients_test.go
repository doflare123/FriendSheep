package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services/notifications"

	"github.com/gin-gonic/gin"
)

type reminderRecipientServiceStub struct {
	response notifications.EventReminderRecipients
	err      error
	eventID  uint
	offset   int
	calls    int
}

func (s *reminderRecipientServiceStub) Resolve(_ context.Context, eventID uint, offset int) (notifications.EventReminderRecipients, error) {
	s.calls++
	s.eventID = eventID
	s.offset = offset
	return s.response, s.err
}

func TestEventReminderRecipientsHTTPRequiresInternalTokenAndExactOffset(t *testing.T) {
	const token = "notify-service-secret-value"
	tests := []struct {
		name          string
		internalToken string
		authorization string
		query         string
		wantStatus    int
		wantCalls     int
	}{
		{name: "valid internal request", internalToken: token, query: "?reminderOffsetMinutes=360", wantStatus: http.StatusOK, wantCalls: 1},
		{name: "missing internal token", query: "?reminderOffsetMinutes=360", wantStatus: http.StatusUnauthorized},
		{name: "JWT alone is insufficient", authorization: "Bearer valid-user-jwt", query: "?reminderOffsetMinutes=360", wantStatus: http.StatusUnauthorized},
		{name: "missing offset", internalToken: token, wantStatus: http.StatusBadRequest},
		{name: "non-positive offset", internalToken: token, query: "?reminderOffsetMinutes=0", wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &reminderRecipientServiceStub{response: notifications.EventReminderRecipients{Recipients: []notifications.EventReminderRecipient{}}}
			request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-reminders/42/recipients"+test.query, nil)
			request.Header.Set(middlewares.InternalTokenHeader, test.internalToken)
			request.Header.Set("Authorization", test.authorization)
			response := httptest.NewRecorder()
			newReminderRecipientTestRouter(stub, token).ServeHTTP(response, request)
			if response.Code != test.wantStatus || stub.calls != test.wantCalls {
				t.Fatalf("response = status:%d calls:%d body:%s, want status:%d calls:%d", response.Code, stub.calls, response.Body.String(), test.wantStatus, test.wantCalls)
			}
			if stub.calls == 1 && (stub.eventID != 42 || stub.offset != 360) {
				t.Fatalf("Resolve() inputs = event:%d offset:%d, want 42/360", stub.eventID, stub.offset)
			}
		})
	}
}

func TestEventReminderRecipientsHTTPSanitizesStoreErrors(t *testing.T) {
	const token = "notify-service-secret-value"
	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "event missing", err: notifications.ErrEventNotFound, wantStatus: http.StatusNotFound, wantCode: "event_not_found"},
		{name: "invalid offset", err: notifications.ErrInvalidReminderOffset, wantStatus: http.StatusBadRequest, wantCode: "invalid_reminder_offset"},
		{name: "internal failure", err: errors.New("postgres DSN and password leaked"), wantStatus: http.StatusInternalServerError, wantCode: "reminder_recipient_resolution_failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &reminderRecipientServiceStub{err: test.err}
			request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-reminders/42/recipients?reminderOffsetMinutes=60", nil)
			request.Header.Set(middlewares.InternalTokenHeader, token)
			response := httptest.NewRecorder()
			newReminderRecipientTestRouter(stub, token).ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			assertScheduleErrorDTO(t, response, test.wantCode)
			if strings.Contains(response.Body.String(), "password") || strings.Contains(response.Body.String(), "DSN") {
				t.Fatalf("response leaked internal details: %s", response.Body.String())
			}
		})
	}
}

func TestEventReminderRecipientsHTTPReturnsMinimalDisplaySnapshot(t *testing.T) {
	const token = "notify-service-secret-value"
	start := time.Date(2036, 8, 30, 18, 0, 0, 0, time.UTC)
	stub := &reminderRecipientServiceStub{response: notifications.EventReminderRecipients{
		EventID: 42, Title: "Reminder event", StartTime: start,
		Recipients: []notifications.EventReminderRecipient{{UserID: 7, Channels: []string{notifications.ChannelInApp}}},
	}}
	request := httptest.NewRequest(http.MethodGet, "/internal/v1/event-reminders/42/recipients?reminderOffsetMinutes=1440", nil)
	request.Header.Set(middlewares.InternalTokenHeader, token)
	response := httptest.NewRecorder()
	newReminderRecipientTestRouter(stub, token).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"email", "telegram", "token", "authorization"} {
		if strings.Contains(strings.ToLower(response.Body.String()), forbidden) {
			t.Fatalf("minimal response contains forbidden field %q: %s", forbidden, response.Body.String())
		}
	}
}

func newReminderRecipientTestRouter(service notifications.EventReminderRecipientService, token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.RegisterInternalEventReminderRecipientRoutes(router, handlers.NewEventReminderRecipientsHandler(service), middlewares.NewInternalTokenMiddleware(token))
	return router
}

var _ notifications.EventReminderRecipientService = (*reminderRecipientServiceStub)(nil)
