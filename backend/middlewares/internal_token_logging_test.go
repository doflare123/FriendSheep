package middlewares

import "testing"

func TestSanitizeHeadersFullyRedactsInternalToken(t *testing.T) {
	const token = "notify-service-secret-value"
	headers := sanitizeHeaders(map[string][]string{
		"X-Internal-Token": {token},
	})

	values := headers["X-Internal-Token"]
	if len(values) != 1 || values[0] != "[REDACTED]" {
		t.Fatalf("sanitized internal token = %#v, want full redaction", values)
	}
}
