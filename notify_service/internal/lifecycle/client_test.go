package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPClientAdvanceLifecycleRejectsZeroEventIDLocally(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "token", time.Second, nil)
	_, err := client.AdvanceLifecycle(context.Background(), 0)

	var clientErr *ClientError
	if !errors.As(err, &clientErr) || clientErr.Code != ErrorCodeInvalidEventID || !clientErr.Terminal {
		t.Fatalf("err = %v, want terminal invalid event id", err)
	}
	if calls != 0 {
		t.Fatalf("HTTP calls = %d, want 0", calls)
	}
}

func TestHTTPClientClassifiesRetryableAndTerminalHTTPErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		body      string
		wantCode  string
		retryable bool
		terminal  bool
	}{
		{"not found", http.StatusNotFound, `{"error":"event_not_found"}`, ErrorCodeEventNotFound, false, true},
		{"server error", http.StatusBadGateway, `{"error":"upstream_failed"}`, "upstream_failed", true, false},
		{"unauthorized", http.StatusUnauthorized, `{"error":"invalid_internal_token"}`, ErrorCodeInvalidInternalToken, true, false},
		{"forbidden", http.StatusForbidden, `{"error":"invalid_internal_token"}`, ErrorCodeInvalidInternalToken, true, false},
		{"rate limited", http.StatusTooManyRequests, `{"error":"rate_limited"}`, "rate_limited", true, false},
		{"bad request", http.StatusBadRequest, `{"error":"invalid_event_id"}`, ErrorCodeInvalidEventID, false, true},
		{"conflict", http.StatusConflict, `{"error":"invalid_event_lifecycle_state"}`, ErrorCodeInvalidLifecycle, false, true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			client := NewHTTPClient(server.URL, "token", time.Second, nil)
			_, err := client.AdvanceLifecycle(context.Background(), 42)

			var clientErr *ClientError
			if !errors.As(err, &clientErr) {
				t.Fatalf("err = %v, want ClientError", err)
			}
			if clientErr.Code != test.wantCode || clientErr.Retryable != test.retryable || clientErr.Terminal != test.terminal {
				t.Fatalf("clientErr = %#v", clientErr)
			}
		})
	}
}

func TestHTTPClientUsesAuthenticatedLifecycleContracts(t *testing.T) {
	t.Parallel()

	const token = "test-internal-token"
	start := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Internal-Token"); got != token {
			t.Errorf("internal token = %q", got)
		}
		switch r.URL.Path {
		case "/internal/v1/event-lifecycle/schedule-events":
			if r.Method != http.MethodGet || r.URL.Query().Get("after") != "12" || r.URL.Query().Get("limit") != "50" {
				t.Errorf("schedule request = %s %s", r.Method, r.URL.String())
			}
			_ = json.NewEncoder(w).Encode(SchedulePage{Items: []ScheduleEvent{}, NextCursor: 12})
		case "/internal/v1/events/42/lifecycle/advance":
			if r.Method != http.MethodPost || r.ContentLength > 0 {
				t.Errorf("advance request method/content length = %s/%d", r.Method, r.ContentLength)
			}
			_ = json.NewEncoder(w).Encode(AdvanceResult{
				EventID: 42, Outcome: OutcomeStarted, Applied: true,
				StartTime: start, EndTime: end, ProcessedAt: start,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, token, time.Second, nil)
	if _, err := client.ListScheduleEvents(context.Background(), 12, 50); err != nil {
		t.Fatalf("ListScheduleEvents() error = %v", err)
	}
	result, err := client.AdvanceLifecycle(context.Background(), 42)
	if err != nil || result.Outcome != OutcomeStarted {
		t.Fatalf("AdvanceLifecycle() = %#v, %v", result, err)
	}
}

func TestHTTPClientNeverForwardsInternalTokenAcrossRedirect(t *testing.T) {
	t.Parallel()

	var redirectedCalls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirectedCalls.Add(1)
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL)
		w.WriteHeader(http.StatusFound)
	}))
	defer source.Close()

	client := NewHTTPClient(source.URL, "must-not-leak", time.Second, nil)
	_, err := client.AdvanceLifecycle(context.Background(), 42)
	var clientErr *ClientError
	if !errors.As(err, &clientErr) || !clientErr.Retryable {
		t.Fatalf("redirect error = %v, want retryable ClientError", err)
	}
	if redirectedCalls.Load() != 0 {
		t.Fatalf("redirect target calls = %d, want 0", redirectedCalls.Load())
	}
}

func TestHTTPClientRejectsMalformedSuccessfulLifecycleResponse(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	tests := []AdvanceResult{
		{EventID: 7, Outcome: OutcomeStarted, Applied: true, StartTime: start, EndTime: start.Add(time.Hour)},
		{EventID: 42, Outcome: OutcomeStarted, Applied: true, StartTime: start, EndTime: start},
		{EventID: 42, Outcome: OutcomeStarted, Applied: false, StartTime: start, EndTime: start.Add(time.Hour)},
	}
	for index, response := range tests {
		response := response
		t.Run(fmt.Sprintf("case_%d", index), func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := NewHTTPClient(server.URL, "token", time.Second, nil)
			_, err := client.AdvanceLifecycle(context.Background(), 42)
			var clientErr *ClientError
			if !errors.As(err, &clientErr) || clientErr.Code != ErrorCodeMalformedResponse || !clientErr.Retryable {
				t.Fatalf("error = %v, want retryable malformed response", err)
			}
		})
	}
}
