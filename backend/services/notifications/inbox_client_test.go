package notifications_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"friendship/services/notifications"
)

func TestHTTPNotificationInboxClientUsesOwnedInternalPathsAndToken(t *testing.T) {
	const token = "notify-service-secret-value"
	requests := make(chan *http.Request, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/internal/v1/users/42/notifications":
			_, _ = w.Write([]byte(`{"items":[],"nextCursor":"cursor-2","hasMore":false}`))
		case "/internal/v1/users/42/notifications/unread-count":
			_, _ = w.Write([]byte(`{"unreadCount":3}`))
		case "/internal/v1/users/42/notifications/notification-1/read":
			_, _ = w.Write([]byte(`{"id":"notification-1","readAt":"2036-08-30T12:00:00Z"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := notifications.NewHTTPNotificationInboxClient(server.URL, token, 2*time.Second)
	if err != nil {
		t.Fatalf("NewHTTPNotificationInboxClient() error: %v", err)
	}

	page, err := client.List(context.Background(), 42, "cursor-1", 11, true)
	if err != nil || page.Items == nil || page.NextCursor != "cursor-2" {
		t.Fatalf("List() = page:%#v err:%v", page, err)
	}
	count, err := client.UnreadCount(context.Background(), 42)
	if err != nil || count.UnreadCount != 3 {
		t.Fatalf("UnreadCount() = result:%#v err:%v", count, err)
	}
	marked, err := client.MarkRead(context.Background(), 42, "notification-1")
	if err != nil || marked.ID != "notification-1" || marked.ReadAt.IsZero() {
		t.Fatalf("MarkRead() = result:%#v err:%v", marked, err)
	}

	listRequest := <-requests
	countRequest := <-requests
	markRequest := <-requests
	if listRequest.Method != http.MethodGet || listRequest.URL.Path != "/internal/v1/users/42/notifications" ||
		listRequest.URL.Query().Get("after") != "cursor-1" || listRequest.URL.Query().Get("limit") != "11" ||
		listRequest.URL.Query().Get("unread") != "true" {
		t.Fatalf("list request = %s %s", listRequest.Method, listRequest.URL.String())
	}
	if countRequest.Method != http.MethodGet || countRequest.URL.Path != "/internal/v1/users/42/notifications/unread-count" {
		t.Fatalf("count request = %s %s", countRequest.Method, countRequest.URL.String())
	}
	if markRequest.Method != http.MethodPatch || markRequest.URL.Path != "/internal/v1/users/42/notifications/notification-1/read" {
		t.Fatalf("mark request = %s %s", markRequest.Method, markRequest.URL.String())
	}
	for _, request := range []*http.Request{listRequest, countRequest, markRequest} {
		if request.Header.Get("X-Internal-Token") != token || request.Header.Get("Authorization") != "" {
			t.Fatalf("internal auth headers = X-Internal-Token:%q Authorization:%q", request.Header.Get("X-Internal-Token"), request.Header.Get("Authorization"))
		}
	}
}

func TestHTTPNotificationInboxClientMapsOwnershipAndSanitizesUpstreamErrors(t *testing.T) {
	for _, test := range []struct {
		name       string
		status     int
		wantErr    error
		response   string
		wantMethod string
	}{
		{name: "ownership not found", status: http.StatusNotFound, wantErr: notifications.ErrInboxNotFound, response: `{"error":"not yours"}`, wantMethod: http.MethodPatch},
		{name: "bad request", status: http.StatusBadRequest, wantErr: notifications.ErrInboxInvalidInput, response: `{"error":"bad"}`, wantMethod: http.MethodPatch},
		{name: "internal auth", status: http.StatusUnauthorized, wantErr: notifications.ErrInboxUnauthorized, response: `{"token":"secret"}`, wantMethod: http.MethodPatch},
		{name: "upstream failure", status: http.StatusInternalServerError, wantErr: notifications.ErrInboxUnavailable, response: `postgres DSN password=secret`, wantMethod: http.MethodPatch},
	} {
		t.Run(test.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.response))
			}))
			defer server.Close()
			client, err := notifications.NewHTTPNotificationInboxClient(server.URL, "internal-token", time.Second)
			if err != nil {
				t.Fatalf("create client: %v", err)
			}
			_, err = client.MarkRead(context.Background(), 7, "notification-owned-by-someone-else")
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("MarkRead() err = %v, want %v", err, test.wantErr)
			}
			if gotPath != "/internal/v1/users/7/notifications/notification-owned-by-someone-else/read" {
				t.Fatalf("ownership path = %q", gotPath)
			}
			if strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "not yours") {
				t.Fatalf("client error leaked upstream body: %v", err)
			}
		})
	}
}

func TestHTTPNotificationInboxClientRejectsUnsafeConfigurationAndNotificationIDs(t *testing.T) {
	for _, test := range []struct {
		name    string
		baseURL string
		token   string
		timeout time.Duration
	}{
		{name: "relative URL", baseURL: "/notify", token: "token", timeout: time.Second},
		{name: "unsupported URL scheme", baseURL: "ftp://notify.test", token: "token", timeout: time.Second},
		{name: "empty token", baseURL: "http://notify.test", timeout: time.Second},
		{name: "zero timeout", baseURL: "http://notify.test", token: "token"},
		{name: "excessive timeout", baseURL: "http://notify.test", token: "token", timeout: 31 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			if client, err := notifications.NewHTTPNotificationInboxClient(test.baseURL, test.token, test.timeout); err == nil || client != nil {
				t.Fatalf("constructor = client:%#v err:%v, want nil/error", client, err)
			}
		})
	}

	client, err := notifications.NewHTTPNotificationInboxClient("http://notify.test", "token", time.Second)
	if err != nil {
		t.Fatalf("create valid client: %v", err)
	}
	for _, id := range []string{"", "   ", "a/b"} {
		if _, err := client.MarkRead(context.Background(), 7, id); !errors.Is(err, notifications.ErrInboxInvalidInput) {
			t.Fatalf("MarkRead(%q) err = %v, want ErrInboxInvalidInput", id, err)
		}
	}
}
