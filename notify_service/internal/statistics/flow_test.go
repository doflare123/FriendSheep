package statistics

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"notify_service/internal/config"
)

func testLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestClientContractsAndBounds(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	page := Page{Messages: []Trigger{{Sequence: 2, SchemaVersion: 1, EventID: 7, CompletedAt: now}}, NextCursor: 2}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "token" {
			t.Error("internal token отсутствует")
		}
		if r.Method == http.MethodGet {
			if r.URL.Path != "/internal/v1/statistics/triggers" || r.URL.Query().Get("afterSequence") != "1" || r.URL.Query().Get("limit") != "500" {
				t.Errorf("неверный URL: %s", r.URL)
			}
			_ = json.NewEncoder(w).Encode(page)
		} else {
			_ = json.NewEncoder(w).Encode(ProcessResult{EventID: 7, Outcome: "processed", ProcessedAt: &now})
		}
	}))
	defer server.Close()
	client := NewHTTPClient(server.URL, "token", time.Second, nil)
	if result, err := client.ListTriggers(context.Background(), 1, 500); err != nil || result.NextCursor != 2 {
		t.Fatalf("page=%+v err=%v", result, err)
	}
	if result, err := client.ProcessEvent(context.Background(), 7); err != nil || result.Outcome != "processed" {
		t.Fatalf("process=%+v err=%v", result, err)
	}
	for _, body := range []string{`{"messages":null,"nextCursor":1}`, `{"messages":[],"nextCursor":2}`, `{"messages":[],"nextCursor":1,"hasMore":true}`, `{"messages":[{"sequence":2,"schemaVersion":2,"eventId":7,"completedAt":"2026-10-14T12:00:00Z"}],"nextCursor":2}`, strings.Repeat(" ", int(maxResponseBodyBytes)+1)} {
		t.Run("malformed", func(t *testing.T) {
			bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }))
			defer bad.Close()
			if _, err := NewHTTPClient(bad.URL, "token", time.Second, nil).ListTriggers(context.Background(), 1, 500); err == nil {
				t.Fatal("неверная страница принята")
			}
		})
	}
}

func TestClientRejectsInconsistentResultAndDoesNotFollowRedirect(t *testing.T) {
	for _, body := range []string{`{"eventId":8,"outcome":"missing"}`, `{"eventId":7,"outcome":"processed"}`, `{"eventId":7,"outcome":"unknown"}`, `{"eventId":7,"outcome":"not_completed","processedAt":"2026-10-14T12:00:00Z"}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }))
		if _, err := NewHTTPClient(server.URL, "token", time.Second, nil).ProcessEvent(context.Background(), 7); err == nil {
			t.Errorf("неверный result принят: %s", body)
		}
		server.Close()
	}
	destinationCalled := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { destinationCalled = true }))
	defer destination.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	if _, err := NewHTTPClient(server.URL, "token", time.Second, nil).ProcessEvent(context.Background(), 7); err == nil || destinationCalled {
		t.Fatal("redirect должен быть запрещён")
	}
}

type fakeSource struct {
	page Page
	err  error
}

func (s fakeSource) ListTriggers(context.Context, int64, int) (Page, error) { return s.page, s.err }

type fakeCursorStore struct {
	applied bool
	cursor  int64
}

func (s *fakeCursorStore) LoadCursor(context.Context, string) (int64, error) { return s.cursor, nil }
func (s *fakeCursorStore) ApplyTriggers(_ context.Context, _ string, messages []Trigger, _ time.Time) (ApplyResult, error) {
	s.applied = true
	s.cursor = messages[len(messages)-1].Sequence
	return ApplyResult{LastSequence: s.cursor}, nil
}
func TestMalformedPollDoesNotAdvanceCursor(t *testing.T) {
	store := &fakeCursorStore{cursor: 2}
	clock := fixedClock{time.Now().UTC()}
	poller := NewPoller(testLogger(), store, fakeSource{page: Page{Messages: []Trigger{}, NextCursor: 3}}, clock, Backoff{}, 10, time.Second)
	if _, err := poller.PollOnce(context.Background()); err == nil || store.applied || store.cursor != 2 {
		t.Fatal("неверная страница изменила durable cursor")
	}
}

type fakeProcess struct {
	result ProcessResult
	err    error
}

func (f fakeProcess) ProcessEvent(context.Context, uint64) (ProcessResult, error) {
	return f.result, f.err
}

type fakeWorkerStore struct {
	job         *Job
	completed   bool
	retried     bool
	code        string
	next        time.Time
	completeErr error
}

func (s *fakeWorkerStore) ClaimDueJob(context.Context, time.Time, time.Duration) (*Job, error) {
	return s.job, nil
}
func (s *fakeWorkerStore) Complete(context.Context, Job, ProcessResult, time.Time) error {
	s.completed = true
	return s.completeErr
}
func (s *fakeWorkerStore) Retry(_ context.Context, _ Job, next time.Time, code string, _ time.Time) error {
	s.retried = true
	s.next = next
	s.code = code
	return nil
}
func TestWorkerOutcomesRetriesAndAckFailure(t *testing.T) {
	now := time.Now().UTC()
	for _, outcome := range []string{"processed", "already_processed", "missing", "not_completed", "network_error", "cancelled", "ack_failure"} {
		t.Run(outcome, func(t *testing.T) {
			store := &fakeWorkerStore{job: &Job{ID: 1, EventID: 7, LeaseToken: "lease", AttemptCount: 2}}
			client := fakeProcess{result: ProcessResult{EventID: 7, Outcome: outcome, ProcessedAt: &now}}
			if outcome == "network_error" {
				client.err = &ClientError{Code: outcome}
			}
			if outcome == "cancelled" {
				client.err = context.Canceled
			}
			if outcome == "ack_failure" {
				client.result.Outcome = "processed"
				store.completeErr = errors.New("ack DB недоступна")
			}
			worker := NewWorker(testLogger(), store, client, fixedClock{now}, Backoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)
			processed, err := worker.ProcessNext(context.Background())
			if !processed {
				t.Fatal("job не принят")
			}
			switch outcome {
			case "not_completed", "network_error":
				if err != nil || !store.retried || store.completed || !store.next.After(now) {
					t.Fatalf("retry=%+v err=%v", store, err)
				}
			case "cancelled":
				if !errors.Is(err, context.Canceled) || store.completed || store.retried {
					t.Fatal("shutdown изменил job вместо lease recovery")
				}
			case "ack_failure":
				if err == nil || store.retried {
					t.Fatal("потеря ack должна восстановиться по lease")
				}
			default:
				if err != nil || !store.completed || store.retried {
					t.Fatalf("complete=%+v err=%v", store, err)
				}
			}
		})
	}
}

func TestManagerStopsBeforeStartAndRejectsRestart(t *testing.T) {
	manager := &Manager{done: make(chan struct{})}
	if err := manager.Stop(time.Second); err != nil {
		t.Fatal(err)
	}
	if err := manager.Start(context.Background()); err == nil {
		t.Fatal("неинициализированный manager запущен")
	}
}

func TestManagerSeparatesStatisticsCommandBudgetFromSourcePolling(t *testing.T) {
	manager := NewManager(config.Config{
		MonolithBaseURL: "http://monolith", InternalToken: "token", MonolithHTTPTimeout: 5 * time.Second,
		StatisticsHTTPTimeout: 30 * time.Second, StatisticsJobLeaseDuration: 2 * time.Minute, JobLeaseDuration: 30 * time.Second,
	}, testLogger(), nil, nil)
	source := manager.poller.client.(*HTTPClient)
	command := manager.worker.client.(*HTTPClient)
	if source.client.Timeout != 5*time.Second || command.client.Timeout != 30*time.Second || manager.worker.lease != 2*time.Minute {
		t.Fatalf("source timeout=%s command timeout=%s statistics lease=%s", source.client.Timeout, command.client.Timeout, manager.worker.lease)
	}
}

type cancellingSource struct{ started chan struct{} }

func (s cancellingSource) ListTriggers(ctx context.Context, _ int64, _ int) (Page, error) {
	close(s.started)
	<-ctx.Done()
	return Page{}, ctx.Err()
}

type cancellingWorkerStore struct{ started chan struct{} }

func (s cancellingWorkerStore) ClaimDueJob(ctx context.Context, _ time.Time, _ time.Duration) (*Job, error) {
	close(s.started)
	<-ctx.Done()
	return nil, ctx.Err()
}
func (cancellingWorkerStore) Complete(context.Context, Job, ProcessResult, time.Time) error {
	return nil
}
func (cancellingWorkerStore) Retry(context.Context, Job, time.Time, string, time.Time) error {
	return nil
}

func TestManagerShutdownCancelsPollAndClaimAndWaits(t *testing.T) {
	sourceStarted, workerStarted := make(chan struct{}), make(chan struct{})
	clock := fixedClock{time.Now().UTC()}
	manager := &Manager{
		poller: NewPoller(testLogger(), &fakeCursorStore{}, cancellingSource{sourceStarted}, clock, Backoff{}, 10, time.Second),
		worker: NewWorker(testLogger(), cancellingWorkerStore{workerStarted}, fakeProcess{}, clock, Backoff{}, time.Minute, time.Second),
		done:   make(chan struct{}),
	}
	if err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, started := range []chan struct{}{sourceStarted, workerStarted} {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("worker не запущен")
		}
	}
	if err := manager.Stop(time.Second); err != nil {
		t.Fatal(err)
	}
	select {
	case <-manager.done:
	default:
		t.Fatal("shutdown не дождался workers")
	}
	if err := manager.Start(context.Background()); err == nil {
		t.Fatal("manager перезапущен после shutdown")
	}
}
