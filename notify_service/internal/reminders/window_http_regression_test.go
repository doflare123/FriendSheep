package reminders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

type httpWindowClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *httpWindowClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *httpWindowClock) set(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
}

type httpWindowStore struct {
	workerStoreStub
	state string
}

func (s *httpWindowStore) MaterializeJob(_ context.Context, job Job, _ RecipientSnapshot, clock Clock) (int, error) {
	s.completeCall++
	s.state = evaluateJobWindow(job, clock.Now())
	return 0, nil
}

func TestWorkerRechecksWindowAfterRecipientHTTP(t *testing.T) {
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	for _, offset := range []int{1440, 360, 60} {
		occurrences, err := buildOccurrences(start, []int{1440, 360, 60})
		if err != nil {
			t.Fatal(err)
		}
		var due, end time.Time
		for _, occurrence := range occurrences {
			if occurrence.OffsetMinutes == offset {
				due, end = occurrence.DueAt, occurrence.WindowEndAt
			}
		}
		for _, delta := range []time.Duration{0, time.Nanosecond} {
			name := strconv.Itoa(offset) + "/" + delta.String()
			t.Run(name, func(t *testing.T) {
				clock := &httpWindowClock{now: end.Add(-time.Second)}
				entered, release := make(chan struct{}), make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					close(entered)
					select {
					case <-release:
					case <-r.Context().Done():
						return
					}
					_ = json.NewEncoder(w).Encode(RecipientSnapshot{
						EventID: 1, Title: "Напоминание", StartTime: start,
						ReminderOffsetMinutes: offset,
						Recipients:            []Recipient{{UserID: 1, Channels: []string{ChannelCodeInApp}}},
					})
				}))
				defer server.Close()
				store := &httpWindowStore{workerStoreStub: workerStoreStub{job: &Job{
					ID: 1, EventID: 1, ReminderOffsetMinutes: offset,
					StartTime: &start, DueAt: &due, WindowEndAt: &end,
				}}}
				worker := NewWorker(discardReminderLogger(), store,
					NewHTTPClient(server.URL, "test", 5*time.Second, nil), clock, nil,
					ExponentialBackoff{Min: time.Second, Max: time.Minute}, time.Minute, time.Second)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				done := make(chan error, 1)
				go func() { _, err := worker.ProcessNext(ctx); done <- err }()
				select {
				case <-entered:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				clock.set(end.Add(delta))
				close(release)
				if err := <-done; err != nil {
					t.Fatal(err)
				}
				want := StateSkippedMissedWindow
				if offset == 60 {
					want = StateExpired
				}
				if store.completeCall != 1 || store.state != want || store.finalCalls != 0 {
					t.Fatalf("materialization calls=%d state=%s final calls=%d; want 1/%s/0", store.completeCall, store.state, store.finalCalls, want)
				}
			})
		}
	}
}
