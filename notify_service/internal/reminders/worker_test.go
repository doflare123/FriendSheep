package reminders

import (
	"context"
	"errors"
	"testing"
	"time"
)

type workerStoreStub struct {
	job          *Job
	claimErr     error
	claimCalls   int
	completeErr  error
	completeCall int
	completeJob  Job
	completeSnap RecipientSnapshot
	finalErr     error
	finalCalls   int
	finalState   string
	finalCode    string
	retryErr     error
	retryCalls   int
	retryCode    string
	retryAt      time.Time
}

func (s *workerStoreStub) ClaimDueJob(_ context.Context, _ time.Time, _ time.Duration) (*Job, error) {
	s.claimCalls++
	return s.job, s.claimErr
}

func (s *workerStoreStub) MaterializeJob(_ context.Context, job Job, snapshot RecipientSnapshot, _ time.Time) (int, error) {
	s.completeCall++
	s.completeJob = job
	s.completeSnap = snapshot
	return len(snapshot.Recipients), s.completeErr
}

func (s *workerStoreStub) RescheduleRetry(_ context.Context, _ Job, nextAttempt time.Time, code string, _ time.Time) error {
	s.retryCalls++
	s.retryCode = code
	s.retryAt = nextAttempt
	return s.retryErr
}

func (s *workerStoreStub) MarkFinal(_ context.Context, _ Job, state string, code string, _ time.Time) error {
	s.finalCalls++
	s.finalState = state
	s.finalCode = code
	return s.finalErr
}

type recipientClientStub struct {
	snapshot RecipientSnapshot
	err      error
}

func (c *recipientClientStub) ResolveRecipients(context.Context, uint64, int) (RecipientSnapshot, error) {
	return c.snapshot, c.err
}

func TestBuildOccurrencesKeepsCanonicalWindowBoundaries(t *testing.T) {
	t.Parallel()

	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	occurrences, err := buildOccurrences(start, []int{60, 1440, 360})
	if err != nil {
		t.Fatalf("buildOccurrences(): %v", err)
	}
	if len(occurrences) != 3 {
		t.Fatalf("len(occurrences) = %d, want 3", len(occurrences))
	}
	if occurrences[0].OffsetMinutes != 1440 || !occurrences[0].DueAt.Equal(start.Add(-24*time.Hour)) || !occurrences[0].WindowEndAt.Equal(start.Add(-6*time.Hour)) {
		t.Fatalf("24h occurrence = %#v", occurrences[0])
	}
	if occurrences[1].OffsetMinutes != 360 || !occurrences[1].DueAt.Equal(start.Add(-6*time.Hour)) || !occurrences[1].WindowEndAt.Equal(start.Add(-time.Hour)) {
		t.Fatalf("6h occurrence = %#v", occurrences[1])
	}
	if occurrences[2].OffsetMinutes != 60 || !occurrences[2].DueAt.Equal(start.Add(-time.Hour)) || !occurrences[2].WindowEndAt.Equal(start) {
		t.Fatalf("1h occurrence = %#v", occurrences[2])
	}
}

func TestEvaluateJobWindowSkipsMissedOccurrencesAndExpiresAfterStart(t *testing.T) {
	t.Parallel()

	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-24 * time.Hour)
	windowEnd := start.Add(-6 * time.Hour)
	job := Job{
		StartTime:   &start,
		DueAt:       &due,
		WindowEndAt: &windowEnd,
	}

	if got := evaluateJobWindow(job, start.Add(-12*time.Hour)); got != StateProcessing {
		t.Fatalf("evaluateJobWindow(within window) = %q, want %q", got, StateProcessing)
	}
	if got := evaluateJobWindow(job, start.Add(-6*time.Hour)); got != StateSkippedMissedWindow {
		t.Fatalf("evaluateJobWindow(at next window boundary) = %q, want %q", got, StateSkippedMissedWindow)
	}
	if got := evaluateJobWindow(job, start); got != StateExpired {
		t.Fatalf("evaluateJobWindow(at start) = %q, want %q", got, StateExpired)
	}
}

func TestWorkerCompletesReminderWithinActiveWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 14, 0, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-6 * time.Hour)
	windowEnd := start.Add(-time.Hour)
	job := Job{ID: 7, EventID: 42, SourceRevision: 11, ReminderOffsetMinutes: 360, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease", State: StateProcessing}
	snapshot := RecipientSnapshot{
		EventID:               42,
		Title:                 "Night Ride",
		StartTime:             start,
		ReminderOffsetMinutes: 360,
		Recipients:            []Recipient{{UserID: 9, Channels: []string{ChannelCodeInApp}}},
	}
	store := &workerStoreStub{job: &job}
	client := &recipientClientStub{snapshot: snapshot}
	worker := NewWorker(discardReminderLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.completeCall != 1 || store.finalCalls != 0 || store.retryCalls != 0 {
		t.Fatalf("complete/final/retry = %d/%d/%d", store.completeCall, store.finalCalls, store.retryCalls)
	}
	if store.completeSnap.ReminderOffsetMinutes != 360 {
		t.Fatalf("snapshot offset = %d, want 360", store.completeSnap.ReminderOffsetMinutes)
	}
}

func TestWorkerMarksMissedWindowWithoutCallingMonolith(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 19, 30, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-24 * time.Hour)
	windowEnd := start.Add(-6 * time.Hour)
	job := Job{ID: 7, EventID: 42, SourceRevision: 11, ReminderOffsetMinutes: 1440, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease", State: StateProcessing}
	store := &workerStoreStub{job: &job}
	client := &recipientClientStub{}
	worker := NewWorker(discardReminderLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.finalCalls != 1 || store.finalState != StateSkippedMissedWindow {
		t.Fatalf("final = calls:%d state:%q", store.finalCalls, store.finalState)
	}
	if store.completeCall != 0 || store.retryCalls != 0 {
		t.Fatalf("complete/retry = %d/%d", store.completeCall, store.retryCalls)
	}
}

func TestWorkerSchedulesRetryForRetryableRecipientFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 14, 0, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-6 * time.Hour)
	windowEnd := start.Add(-time.Hour)
	job := Job{ID: 7, EventID: 42, SourceRevision: 11, ReminderOffsetMinutes: 360, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease", State: StateProcessing, AttemptCount: 1}
	store := &workerStoreStub{job: &job}
	client := &recipientClientStub{err: &ClientError{Code: ErrorCodeTimeout, Retryable: true}}
	worker := NewWorker(discardReminderLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.retryCalls != 1 || store.retryCode != ErrorCodeTimeout || !store.retryAt.Equal(now.Add(2*time.Second)) {
		t.Fatalf("retry = calls:%d code:%q at:%s", store.retryCalls, store.retryCode, store.retryAt)
	}
}

func TestWorkerPersistsRetryForMaterializationFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 14, 0, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-6 * time.Hour)
	windowEnd := start.Add(-time.Hour)
	job := Job{ID: 8, EventID: 43, SourceRevision: 12, ReminderOffsetMinutes: 360, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease-delivery", State: StateProcessing}
	snapshot := RecipientSnapshot{
		EventID: 43, Title: "Delivery retry", StartTime: start, ReminderOffsetMinutes: 360,
		Recipients: []Recipient{{UserID: 10, Channels: []string{ChannelCodeInApp}}},
	}
	store := &workerStoreStub{
		job:         &job,
		completeErr: &ClientError{Code: "materialization_failed", Retryable: true},
	}
	worker := NewWorker(
		discardReminderLogger(), store, &recipientClientStub{snapshot: snapshot}, fixedClock{now: now}, nil,
		ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second,
	)

	processed, err := worker.ProcessNext(context.Background())
	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.completeCall != 1 || store.retryCalls != 1 || store.retryCode != "materialization_failed" {
		t.Fatalf("complete/retry/code = %d/%d/%q", store.completeCall, store.retryCalls, store.retryCode)
	}
	if store.finalCalls != 0 {
		t.Fatalf("terminal calls = %d, want 0", store.finalCalls)
	}
}

func TestWorkerPersistsTerminalMaterializationFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 14, 0, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-6 * time.Hour)
	windowEnd := start.Add(-time.Hour)
	job := Job{ID: 9, EventID: 44, SourceRevision: 13, ReminderOffsetMinutes: 360, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease-terminal", State: StateProcessing}
	snapshot := RecipientSnapshot{
		EventID: 44, Title: "Delivery terminal", StartTime: start, ReminderOffsetMinutes: 360,
		Recipients: []Recipient{{UserID: 11, Channels: []string{"test_terminal"}}},
	}
	store := &workerStoreStub{
		job:         &job,
		completeErr: &ClientError{Code: "invalid_materialization", Terminal: true},
	}
	worker := NewWorker(
		discardReminderLogger(), store, &recipientClientStub{snapshot: snapshot}, fixedClock{now: now}, nil,
		ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second,
	)

	processed, err := worker.ProcessNext(context.Background())
	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.completeCall != 1 || store.finalCalls != 1 || store.finalState != StateTerminalFailed || store.finalCode != "invalid_materialization" {
		t.Fatalf("complete/final/state/code = %d/%d/%q/%q", store.completeCall, store.finalCalls, store.finalState, store.finalCode)
	}
	if store.retryCalls != 0 {
		t.Fatalf("retry calls = %d, want 0", store.retryCalls)
	}
}

func TestWorkerMarksStaleSnapshotAsSuperseded(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 14, 0, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-6 * time.Hour)
	windowEnd := start.Add(-time.Hour)
	job := Job{ID: 7, EventID: 42, SourceRevision: 11, ReminderOffsetMinutes: 360, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease", State: StateProcessing}
	store := &workerStoreStub{job: &job}
	client := &recipientClientStub{snapshot: RecipientSnapshot{
		EventID:               42,
		Title:                 "Rescheduled",
		StartTime:             start.Add(time.Hour),
		ReminderOffsetMinutes: 360,
		Recipients:            []Recipient{},
	}}
	worker := NewWorker(discardReminderLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.finalCalls != 1 || store.finalState != StateSuperseded || store.finalCode != ErrorCodeStaleSchedule {
		t.Fatalf("final = calls:%d state:%q code:%q", store.finalCalls, store.finalState, store.finalCode)
	}
}

func TestWorkerMarksTerminalRecipientFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 19, 30, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-time.Hour)
	windowEnd := start
	job := Job{ID: 7, EventID: 42, SourceRevision: 11, ReminderOffsetMinutes: 60, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease", State: StateProcessing}
	store := &workerStoreStub{job: &job}
	client := &recipientClientStub{err: &ClientError{Code: ErrorCodeEventNotFound, Terminal: true}}
	worker := NewWorker(discardReminderLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if store.finalCalls != 1 || store.finalState != StateTerminalFailed || store.finalCode != ErrorCodeEventNotFound {
		t.Fatalf("final = calls:%d state:%q code:%q", store.finalCalls, store.finalState, store.finalCode)
	}
}

func TestWorkerPropagatesContextCancellation(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 30, 14, 0, 0, 0, time.UTC)
	start := time.Date(2036, 8, 30, 20, 0, 0, 0, time.UTC)
	due := start.Add(-6 * time.Hour)
	windowEnd := start.Add(-time.Hour)
	job := Job{ID: 7, EventID: 42, SourceRevision: 11, ReminderOffsetMinutes: 360, StartTime: &start, DueAt: &due, WindowEndAt: &windowEnd, LeaseToken: "lease", State: StateProcessing}
	store := &workerStoreStub{job: &job}
	client := &recipientClientStub{err: context.Canceled}
	worker := NewWorker(discardReminderLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{Min: time.Second, Max: time.Minute}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if !processed || !errors.Is(err, context.Canceled) {
		t.Fatalf("ProcessNext() = processed:%t error:%v, want true/context.Canceled", processed, err)
	}
	if store.finalCalls != 0 || store.completeCall != 0 || store.retryCalls != 0 {
		t.Fatalf("unexpected acknowledgements = final:%d complete:%d retry:%d", store.finalCalls, store.completeCall, store.retryCalls)
	}
}
