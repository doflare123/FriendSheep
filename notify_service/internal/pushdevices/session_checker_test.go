package pushdevices

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPMonolithSessionCheckerSendsAuthenticatedBoundedRequest(t *testing.T) {
	sessionID := testSessionID(71)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/internal/v1/push-sessions/check" {
			t.Errorf("запрос = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("X-Internal-Token") != "internal-secret" {
			t.Error("внутренний токен не передан")
		}
		var body struct {
			SchemaVersion int    `json:"schemaVersion"`
			UserID        uint64 `json:"userId"`
			SessionID     string `json:"sessionId"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("тело запроса: %v", err)
		}
		if body.SchemaVersion != SchemaVersion || body.UserID != 77 || body.SessionID != sessionID {
			t.Errorf("тело запроса = %#v", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":true}`))
	}))
	defer server.Close()

	checker, err := NewHTTPMonolithSessionChecker(server.URL, "internal-secret", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	active, err := checker.CheckSession(context.Background(), 77, sessionID)
	if err != nil || !active {
		t.Fatalf("CheckSession() = %t, %v", active, err)
	}
}

func TestHTTPMonolithSessionCheckerReturnsConfirmedInactive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":false}`))
	}))
	defer server.Close()
	checker, err := NewHTTPMonolithSessionChecker(server.URL, "internal-secret", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	active, err := checker.CheckSession(context.Background(), 77, testSessionID(76))
	if err != nil || active {
		t.Fatalf("CheckSession() = %t, %v", active, err)
	}
}

func TestHTTPMonolithSessionCheckerDoesNotFollowRedirect(t *testing.T) {
	var redirected atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirected.Add(1)
	}))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Location", destination.URL)
		writer.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer source.Close()

	checker, err := NewHTTPMonolithSessionChecker(source.URL, "internal-secret", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	active, err := checker.CheckSession(context.Background(), 77, testSessionID(72))
	if active || SessionCheckErrorCode(err) != "unexpected_status_307" || redirected.Load() != 0 {
		t.Fatalf("redirect: active=%t code=%q redirected=%d", active, SessionCheckErrorCode(err), redirected.Load())
	}
}

func TestHTTPMonolithSessionCheckerTreatsTimeoutAndUnavailableAsErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		select {
		case <-request.Context().Done():
		case <-time.After(time.Second):
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	checker, err := NewHTTPMonolithSessionChecker(server.URL, "internal-secret", 10*time.Millisecond, nil)
	if err != nil {
		t.Fatal(err)
	}
	active, err := checker.CheckSession(context.Background(), 77, testSessionID(73))
	if active || SessionCheckErrorCode(err) != SessionCheckErrorTimeout {
		t.Fatalf("timeout: active=%t code=%q error=%v", active, SessionCheckErrorCode(err), err)
	}
}

func TestHTTPMonolithSessionCheckerRejectsMalformedAndOversizedResponses(t *testing.T) {
	tests := []struct {
		name string
		body func(http.ResponseWriter)
		code string
	}{
		{name: "unknown field", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":false,"unexpected":true}`))
		}, code: SessionCheckErrorMalformedResponse},
		{name: "duplicate active", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":true,"active":false}`))
		}, code: SessionCheckErrorMalformedResponse},
		{name: "alias active", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":false,"Active":true}`))
		}, code: SessionCheckErrorMalformedResponse},
		{name: "null active", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":null}`))
		}, code: SessionCheckErrorMalformedResponse},
		{name: "missing active", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write([]byte(`{"schemaVersion":1}`))
		}, code: SessionCheckErrorMalformedResponse},
		{name: "wrong active type", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write([]byte(`{"schemaVersion":1,"active":"false"}`))
		}, code: SessionCheckErrorMalformedResponse},
		{name: "oversized", body: func(writer http.ResponseWriter) {
			_, _ = writer.Write(make([]byte, maxSessionCheckResponseBytes+1))
		}, code: SessionCheckErrorResponseTooLarge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { test.body(writer) }))
			defer server.Close()
			checker, err := NewHTTPMonolithSessionChecker(server.URL, "internal-secret", time.Second, nil)
			if err != nil {
				t.Fatal(err)
			}
			active, err := checker.CheckSession(context.Background(), 77, testSessionID(74))
			if active || SessionCheckErrorCode(err) != test.code {
				t.Fatalf("CheckSession() = %t, %v, code=%q", active, err, SessionCheckErrorCode(err))
			}
		})
	}
}

func TestHTTPMonolithSessionCheckerPreservesCancellation(t *testing.T) {
	checker, err := NewHTTPMonolithSessionChecker("http://127.0.0.1:1", "internal-secret", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = checker.CheckSession(ctx, 77, testSessionID(75))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ошибка = %v, ожидалась отмена контекста", err)
	}
}
