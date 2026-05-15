package tests

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func assertCommonErrorShape(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error response: %v; body=%s", err, rec.Body.String())
	}
	if payload["error"] == "" {
		t.Fatalf("error response missing non-empty error field: %#v", payload)
	}
	if payload["message"] == "" {
		t.Fatalf("error response missing non-empty message field: %#v", payload)
	}
	return payload
}
