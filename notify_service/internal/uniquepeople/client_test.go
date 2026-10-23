package uniquepeople

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestHTTPClientUsesAuthenticatedUniquePeopleContracts(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	leaseUntil := now.Add(2 * time.Minute)
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("X-Internal-Token") != "secret" {
			t.Fatalf("отсутствует internal token")
		}
		switch r.URL.Path {
		case "/internal/v1/statistics/unique-people/claim":
			if r.Method != http.MethodPost || r.URL.Query().Get("limit") != "2" || r.ContentLength != 0 {
				t.Fatalf("некорректный claim запрос: %s %s content-length=%d", r.Method, r.URL.String(), r.ContentLength)
			}
			_ = json.NewEncoder(w).Encode(ClaimResponse{SchemaVersion: 1, Jobs: []Claim{{UserID: 7, Generation: 3, LeaseToken: "lease", LeaseUntil: leaseUntil}}})
		case "/internal/v1/statistics/unique-people/execute":
			if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("некорректный execute запрос")
			}
			var request map[string]any
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request["schemaVersion"] != float64(1) || request["userId"] != float64(7) || request["generation"] != float64(3) || request["leaseToken"] != "lease" {
				t.Fatalf("некорректное тело execute: %#v", request)
			}
			count := int64(4)
			calculated := now.Add(time.Second)
			_ = json.NewEncoder(w).Encode(ExecuteResult{SchemaVersion: 1, UserID: 7, Generation: 4, Outcome: OutcomeApplied, Count: &count, CalculatedAt: &calculated})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "secret", time.Second, nil, fixedClock{now})
	page, err := client.Claim(context.Background(), 2)
	if err != nil || len(page.Jobs) != 1 {
		t.Fatalf("Claim() = %#v, %v", page, err)
	}
	result, err := client.Execute(context.Background(), page.Jobs[0])
	if err != nil || result.Generation != 4 || result.Outcome != OutcomeApplied {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestHTTPClientRejectsMalformedResponses(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		body string
	}{
		{name: "unsupported schema", body: `{"schemaVersion":2,"jobs":[]}`},
		{name: "nil jobs", body: `{"schemaVersion":1,"jobs":null}`},
		{name: "expired lease", body: `{"schemaVersion":1,"jobs":[{"userId":1,"generation":1,"leaseToken":"x","leaseUntil":"2026-10-23T09:59:59Z"}]}`},
		{name: "empty token", body: `{"schemaVersion":1,"jobs":[{"userId":1,"generation":1,"leaseToken":" ","leaseUntil":"2026-10-23T10:02:00Z"}]}`},
		{name: "duplicate user", body: `{"schemaVersion":1,"jobs":[{"userId":1,"generation":1,"leaseToken":"a","leaseUntil":"2026-10-23T10:02:00Z"},{"userId":1,"generation":2,"leaseToken":"b","leaseUntil":"2026-10-23T10:02:00Z"}]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, test.body)
			}))
			defer server.Close()
			client := NewHTTPClient(server.URL, "secret", time.Second, nil, fixedClock{now})
			if _, err := client.Claim(context.Background(), 1); err == nil || !strings.Contains(err.Error(), "malformed_response") {
				t.Fatalf("malformed response принят: %v", err)
			}
		})
	}
}

func TestHTTPClientRequiresSuccessfulExecutePayload(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	claim := Claim{UserID: 7, Generation: 3, LeaseToken: "lease", LeaseUntil: now.Add(time.Minute)}
	for _, body := range []string{
		`{"schemaVersion":1,"userId":7,"generation":3,"outcome":"applied","calculatedAt":"2026-10-23T10:00:01Z"}`,
		`{"schemaVersion":1,"userId":7,"generation":3,"outcome":"already_applied","count":4}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, body)
		}))
		client := NewHTTPClient(server.URL, "secret", time.Second, nil, fixedClock{now})
		_, err := client.Execute(context.Background(), claim)
		server.Close()
		if err == nil || !strings.Contains(err.Error(), "malformed_response") {
			t.Fatalf("неполный successful payload принят: %v", err)
		}
	}
}

func TestHTTPClientBoundsResponseAndValidatesExecuteGeneration(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	claim := Claim{UserID: 7, Generation: 3, LeaseToken: "lease", LeaseUntil: now.Add(time.Minute)}
	t.Run("response too large", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, strings.Repeat("x", int(maxResponseBodyBytes)+1))
		}))
		defer server.Close()
		client := NewHTTPClient(server.URL, "secret", time.Second, nil, fixedClock{now})
		if _, err := client.Claim(context.Background(), 1); err == nil || !strings.Contains(err.Error(), "response_too_large") {
			t.Fatalf("oversized response принят: %v", err)
		}
	})
	t.Run("stale generation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"schemaVersion":1,"userId":7,"generation":2,"outcome":"applied"}`)
		}))
		defer server.Close()
		client := NewHTTPClient(server.URL, "secret", time.Second, nil, fixedClock{now})
		if _, err := client.Execute(context.Background(), claim); err == nil || !strings.Contains(err.Error(), "malformed_response") {
			t.Fatalf("stale generation принят: %v", err)
		}
	})
	t.Run("stale lease generation zero", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"schemaVersion":1,"userId":7,"generation":0,"outcome":"stale_lease"}`)
		}))
		defer server.Close()
		client := NewHTTPClient(server.URL, "secret", time.Second, nil, fixedClock{now})
		if _, err := client.Execute(context.Background(), claim); err != nil {
			t.Fatalf("stale_lease отклонён: %v", err)
		}
	})
}
