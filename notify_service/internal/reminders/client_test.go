package reminders

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClientUsesAuthenticatedVersionedReminderEndpoints(t *testing.T) {
	t.Parallel()

	const token = "internal-test-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != token {
			t.Fatalf("X-Internal-Token = %q", r.Header.Get("X-Internal-Token"))
		}
		switch r.URL.Path {
		case "/internal/v1/notification-intents/event-reminders":
			if r.URL.Query().Get("after") != "12" || r.URL.Query().Get("limit") != "25" {
				t.Fatalf("source query = %q", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items":[],"nextCursor":12,"hasMore":false}`))
		case "/internal/v1/event-reminders/42/recipients":
			if r.URL.Query().Get("reminderOffsetMinutes") != "360" {
				t.Fatalf("recipient query = %q", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"eventId":42,"title":"Встреча","startTime":"2036-09-01T18:00:00Z","reminderOffsetMinutes":360,"recipients":[{"userId":7,"channels":["in_app"]}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, token, time.Second, nil)
	page, err := client.ListReminderIntents(context.Background(), 12, 25)
	if err != nil {
		t.Fatalf("ListReminderIntents(): %v", err)
	}
	if page.NextCursor != 12 || page.Items == nil {
		t.Fatalf("page = %#v", page)
	}
	snapshot, err := client.ResolveRecipients(context.Background(), 42, 360)
	if err != nil {
		t.Fatalf("ResolveRecipients(): %v", err)
	}
	if snapshot.EventID != 42 || len(snapshot.Recipients) != 1 || snapshot.Recipients[0].UserID != 7 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestHTTPClientClassifiesReminderFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		body      string
		retryable bool
		terminal  bool
	}{
		{name: "auth", status: http.StatusUnauthorized, body: `{"error":"invalid_internal_token"}`, retryable: true},
		{name: "not found", status: http.StatusNotFound, body: `{"error":"event_not_found"}`, terminal: true},
		{name: "invalid", status: http.StatusBadRequest, body: `{"error":"invalid_reminder_offset"}`, terminal: true},
		{name: "rate limited", status: http.StatusTooManyRequests, body: `{"error":"rate_limited"}`, retryable: true},
		{name: "server", status: http.StatusServiceUnavailable, body: `{"error":"temporarily_unavailable"}`, retryable: true},
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
			_, err := client.ResolveRecipients(context.Background(), 42, 60)
			var clientErr *ClientError
			if !errors.As(err, &clientErr) {
				t.Fatalf("error = %v, want ClientError", err)
			}
			if clientErr.Retryable != test.retryable || clientErr.Terminal != test.terminal {
				t.Fatalf("classification = retryable:%t terminal:%t", clientErr.Retryable, clientErr.Terminal)
			}
		})
	}
}

func TestHTTPClientRejectsMalformedRecipientSnapshotAsTerminal(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"eventId":42,"title":"","startTime":"2036-09-01T18:00:00Z","reminderOffsetMinutes":60,"recipients":[]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "token", time.Second, nil)
	_, err := client.ResolveRecipients(context.Background(), 42, 60)
	var clientErr *ClientError
	if !errors.As(err, &clientErr) || !clientErr.Terminal || clientErr.Code != ErrorCodeMalformedResponse {
		t.Fatalf("error = %#v", err)
	}
}
