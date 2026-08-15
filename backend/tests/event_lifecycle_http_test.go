package tests

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
	servicesevents "friendship/services/events"

	"github.com/gin-gonic/gin"
)

type lifecycleHandlerServiceStub struct {
	result  servicesevents.EventLifecycleResult
	err     error
	ctx     context.Context
	eventID uint
	calls   int
}

func (s *lifecycleHandlerServiceStub) Advance(ctx context.Context, eventID uint) (servicesevents.EventLifecycleResult, error) {
	s.ctx = ctx
	s.eventID = eventID
	s.calls++
	return s.result, s.err
}

func TestInternalEventLifecycleRouteRequiresOnlyValidServiceToken(t *testing.T) {
	const token = "notify-service-secret-value"
	tests := []struct {
		name          string
		header        string
		authorization string
		configured    string
		wantStatus    int
		wantCalls     int
	}{
		{"valid internal token", token, "", token, http.StatusOK, 1},
		{"missing token", "", "", token, http.StatusUnauthorized, 0},
		{"wrong token", "wrong", "", token, http.StatusUnauthorized, 0},
		{"user JWT without internal token", "", "Bearer otherwise-valid-user-jwt", token, http.StatusUnauthorized, 0},
		{"empty configuration fails closed", "anything", "", "", http.StatusUnauthorized, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &lifecycleHandlerServiceStub{result: lifecycleHTTPResult()}
			router := newLifecycleTestRouter(stub, tt.configured)
			request := httptest.NewRequest(http.MethodPost, "/internal/v1/events/123/lifecycle/advance", nil)
			if tt.header != "" {
				request.Header.Set(middlewares.InternalTokenHeader, tt.header)
			}
			if tt.authorization != "" {
				request.Header.Set("Authorization", tt.authorization)
			}
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if stub.calls != tt.wantCalls {
				t.Fatalf("service calls = %d, want %d", stub.calls, tt.wantCalls)
			}
			if tt.wantStatus == http.StatusUnauthorized {
				assertLifecycleErrorDTO(t, recorder, "invalid_internal_token")
			}
		})
	}
}

func TestInternalEventLifecycleHandlerResponseAndErrorContract(t *testing.T) {
	const token = "notify-service-secret-value"
	tests := []struct {
		name       string
		path       string
		err        error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{"invalid event ID text", "/internal/v1/events/nope/lifecycle/advance", nil, http.StatusBadRequest, "invalid_event_id", 0},
		{"zero event ID", "/internal/v1/events/0/lifecycle/advance", nil, http.StatusBadRequest, "invalid_event_id", 0},
		{"not found", "/internal/v1/events/123/lifecycle/advance", servicesevents.ErrEventNotFound, http.StatusNotFound, "event_not_found", 1},
		{"invalid lifecycle state", "/internal/v1/events/123/lifecycle/advance", servicesevents.ErrInvalidEventLifecycleState, http.StatusConflict, "invalid_event_lifecycle_state", 1},
		{"internal failure is stable and sanitized", "/internal/v1/events/123/lifecycle/advance", errors.New("database password leaked"), http.StatusInternalServerError, "event_lifecycle_failed", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &lifecycleHandlerServiceStub{result: lifecycleHTTPResult(), err: tt.err}
			router := newLifecycleTestRouter(stub, token)
			request := httptest.NewRequest(http.MethodPost, tt.path, nil)
			request.Header.Set(middlewares.InternalTokenHeader, token)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if stub.calls != tt.wantCalls {
				t.Fatalf("service calls = %d, want %d", stub.calls, tt.wantCalls)
			}
			assertLifecycleErrorDTO(t, recorder, tt.wantCode)
			if tt.wantStatus == http.StatusInternalServerError && strings.Contains(recorder.Body.String(), "database password leaked") {
				t.Fatal("internal error response leaked implementation details")
			}
		})
	}
}

func TestInternalEventLifecycleHandlerPreservesContextAndReturnsDTO(t *testing.T) {
	const token = "notify-service-secret-value"
	type contextKey struct{}
	stub := &lifecycleHandlerServiceStub{result: lifecycleHTTPResult()}
	router := newLifecycleTestRouter(stub, token)
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/events/123/lifecycle/advance", nil)
	request = request.WithContext(context.WithValue(request.Context(), contextKey{}, "request-value"))
	request.Header.Set(middlewares.InternalTokenHeader, token)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}
	if stub.calls != 1 || stub.eventID != 123 || stub.ctx.Value(contextKey{}) != "request-value" {
		t.Fatalf("service call = count:%d event:%d context:%v", stub.calls, stub.eventID, stub.ctx.Value(contextKey{}))
	}
	var response servicesevents.EventLifecycleResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode lifecycle response: %v", err)
	}
	if response != stub.result {
		t.Fatalf("response = %#v, want %#v", response, stub.result)
	}
}

func lifecycleHTTPResult() servicesevents.EventLifecycleResult {
	return servicesevents.EventLifecycleResult{
		EventID:        123,
		PreviousStatus: "Набор",
		CurrentStatus:  "В процессе",
		Outcome:        servicesevents.LifecycleOutcomeStarted,
		Applied:        true,
		StartTime:      time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2036, 8, 15, 14, 0, 0, 0, time.UTC),
		ProcessedAt:    time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC),
	}
}

func newLifecycleTestRouter(service servicesevents.EventLifecycleService, token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.RegisterInternalEventLifecycleRoutes(
		router,
		handlers.NewEventLifecycleHandler(service),
		middlewares.NewInternalTokenMiddleware(token),
	)
	return router
}

func assertLifecycleErrorDTO(t *testing.T, recorder *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error DTO: %v; body = %s", err, recorder.Body.String())
	}
	if response.Error != wantCode || response.Message == "" {
		t.Fatalf("error response = %#v, want code %q and non-empty message", response, wantCode)
	}
}

var _ servicesevents.EventLifecycleService = (*lifecycleHandlerServiceStub)(nil)
