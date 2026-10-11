package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type sourcePageCursorStore struct {
	cursor uint64
	calls  int
}

func (s *sourcePageCursorStore) LoadCursor(context.Context, string) (uint64, error) {
	return s.cursor, nil
}
func (s *sourcePageCursorStore) ApplySourceEvents(_ context.Context, _ string, items []ScheduleEvent, _ time.Time) (ApplyResult, error) {
	s.calls++
	if len(items) > 0 {
		s.cursor = items[len(items)-1].Sequence
	}
	return ApplyResult{LastSequence: s.cursor, AppliedCount: len(items)}, nil
}
func TestHTTPSourceMaximumPageAdvancesCursor(t *testing.T) {
	now := time.Date(9998, 12, 31, 23, 59, 59, 999999999, time.FixedZone("max", 14*3600))
	end := now.Add(time.Hour)
	_ = end
	const after uint64 = 9223372036854775000
	page := SchedulePage{Items: make([]ScheduleEvent, 500), NextCursor: after + 500, HasMore: true}
	for i := range page.Items {
		page.Items[i] = ScheduleEvent{Sequence: after + uint64(i) + 1, MessageID: fmt.Sprintf("00000000-0000-4000-8000-%012d", i), SchemaVersion: 1, Operation: OperationScheduleUpsert, EventID: 9223372036854775807, StartTime: &now, OccurredAt: now, EndTime: &end}
	}
	body, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) <= 64<<10 {
		t.Fatalf("fixture too small: %d", len(body))
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		expected := after
		if calls > 1 {
			expected = page.NextCursor
		}
		if r.URL.Query().Get("after") != fmt.Sprint(expected) || r.URL.Query().Get("limit") != "500" {
			t.Errorf("query=%s", r.URL.RawQuery)
		}
		if calls == 1 {
			_, _ = w.Write(body)
			return
		}
		_ = json.NewEncoder(w).Encode(SchedulePage{Items: []ScheduleEvent{}, NextCursor: page.NextCursor})
	}))
	defer server.Close()
	store := &sourcePageCursorStore{cursor: after}
	poller := NewPoller(discardLogger(), store, NewHTTPClient(server.URL, "token", time.Second, nil), fixedClock{now: now}, nil, ExponentialBackoff{}, SourceName, 500, time.Second)
	more, err := poller.PollOnce(context.Background())
	if err != nil || !more || store.cursor != page.NextCursor {
		t.Fatalf("first poll more=%t cursor=%d err=%v", more, store.cursor, err)
	}
	more, err = poller.PollOnce(context.Background())
	if err != nil || more || store.cursor != page.NextCursor || calls != 2 {
		t.Fatalf("second poll more=%t cursor=%d calls=%d err=%v", more, store.cursor, calls, err)
	}
}
func TestHTTPSourceRejectsOversizeAndMalformedWithoutCursorAdvance(t *testing.T) {
	for _, tc := range []struct{ name, body, code string }{
		{"oversize", strings.Repeat(" ", int(maxResponseBodyBytes)) + "{}", ErrorCodeResponseTooLarge},
		{"malformed", `{"items":[`, ErrorCodeMalformedResponse},
		{"trailing json", `{"items":[],"nextCursor":17,"hasMore":false}{}`, ErrorCodeMalformedResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(tc.body)) }))
			defer server.Close()
			store := &sourcePageCursorStore{cursor: 17}
			poller := NewPoller(discardLogger(), store, NewHTTPClient(server.URL, "token", time.Second, nil), fixedClock{now: time.Now()}, nil, ExponentialBackoff{}, SourceName, 500, time.Second)
			_, err := poller.PollOnce(context.Background())
			var ce *ClientError
			if !errors.As(err, &ce) || ce.Code != tc.code || !ce.Retryable {
				t.Fatalf("error=%#v", err)
			}
			if store.cursor != 17 || store.calls != 0 {
				t.Fatalf("cursor=%d apply calls=%d", store.cursor, store.calls)
			}
		})
	}
}
