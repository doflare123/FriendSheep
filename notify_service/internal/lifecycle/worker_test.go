package lifecycle

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type workerStoreStub struct {
	job           *Job
	claimErr      error
	successErr    error
	retryErr      error
	terminalErr   error
	claimCtx      context.Context
	claimNow      time.Time
	leaseDuration time.Duration
	claimCalls    int
	successCalls  int
	retryCalls    int
	terminalCalls int
	successJob    Job
	successResult AdvanceResult
	successNow    time.Time
	retryJob      Job
	nextAttempt   time.Time
	retryCode     string
	retryNow      time.Time
	terminalJob   Job
	terminalState string
	terminalCode  string
	terminalNow   time.Time
}

func (store *workerStoreStub) ClaimDueJob(ctx context.Context, now time.Time, leaseDuration time.Duration) (*Job, error) {
	store.claimCalls++
	store.claimCtx = ctx
	store.claimNow = now
	store.leaseDuration = leaseDuration
	return store.job, store.claimErr
}

func (store *workerStoreStub) RescheduleAfterSuccess(ctx context.Context, job Job, result AdvanceResult, now time.Time) error {
	store.successCalls++
	store.successJob = job
	store.successResult = result
	store.successNow = now
	return store.successErr
}

func (store *workerStoreStub) RescheduleRetry(ctx context.Context, job Job, nextAttempt time.Time, code string, now time.Time) error {
	store.retryCalls++
	store.retryJob = job
	store.nextAttempt = nextAttempt
	store.retryCode = code
	store.retryNow = now
	return store.retryErr
}

func (store *workerStoreStub) MarkTerminal(ctx context.Context, job Job, state string, code string, now time.Time) error {
	store.terminalCalls++
	store.terminalJob = job
	store.terminalState = state
	store.terminalCode = code
	store.terminalNow = now
	return store.terminalErr
}

type advanceClientStub struct {
	result     AdvanceResult
	err        error
	ctx        context.Context
	eventID    uint64
	calls      int
	started    chan struct{}
	waitForCtx bool
}

func (client *advanceClientStub) AdvanceLifecycle(ctx context.Context, eventID uint64) (AdvanceResult, error) {
	client.calls++
	client.ctx = ctx
	client.eventID = eventID
	if client.started != nil {
		close(client.started)
	}
	if client.waitForCtx {
		<-ctx.Done()
		return AdvanceResult{}, ctx.Err()
	}
	return client.result, client.err
}

func TestWorkerProcessNextClaimsOutsideHTTPAndPersistsSuccessfulOutcome(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	job := testClaimedJob(now)
	result := testAdvanceResult(OutcomeStarted, now)
	store := &workerStoreStub{job: &job}
	client := &advanceClientStub{result: result}
	worker := NewWorker(discardLifecycleLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{}, 30*time.Second, time.Second)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "preserved")

	processed, err := worker.ProcessNext(ctx)

	if err != nil || !processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v, want true/nil", processed, err)
	}
	if store.claimCalls != 1 || store.claimCtx != ctx || !store.claimNow.Equal(now) || store.leaseDuration != 30*time.Second {
		t.Fatalf("claim = count:%d context:%t now:%s lease:%s", store.claimCalls, store.claimCtx == ctx, store.claimNow, store.leaseDuration)
	}
	if client.calls != 1 || client.ctx != ctx || client.eventID != job.EventID {
		t.Fatalf("client = count:%d context:%t event:%d", client.calls, client.ctx == ctx, client.eventID)
	}
	if store.successCalls != 1 || store.successJob != job || !reflect.DeepEqual(store.successResult, result) || !store.successNow.Equal(now) {
		t.Fatalf("success ack = count:%d job:%#v result:%#v now:%s", store.successCalls, store.successJob, store.successResult, store.successNow)
	}
	if store.retryCalls != 0 || store.terminalCalls != 0 {
		t.Fatalf("unexpected retry/terminal calls = %d/%d", store.retryCalls, store.terminalCalls)
	}
}

func TestWorkerProcessNextDoesNothingWhenNoJobIsDue(t *testing.T) {
	t.Parallel()

	store := &workerStoreStub{}
	client := &advanceClientStub{}
	worker := NewWorker(discardLifecycleLogger(), store, client, fixedClock{}, nil, ExponentialBackoff{}, 30*time.Second, time.Second)

	processed, err := worker.ProcessNext(context.Background())

	if err != nil || processed {
		t.Fatalf("ProcessNext() = processed:%t error:%v, want false/nil", processed, err)
	}
	if client.calls != 0 || store.successCalls != 0 || store.retryCalls != 0 || store.terminalCalls != 0 {
		t.Fatalf("side effects with no job: client=%d success=%d retry=%d terminal=%d",
			client.calls, store.successCalls, store.retryCalls, store.terminalCalls)
	}
}

func TestWorkerProcessNextPersistsRetryWithBoundedExponentialBackoff(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "network", err: &ClientError{Code: ErrorCodeNetwork, Retryable: true}, code: ErrorCodeNetwork},
		{name: "timeout", err: &ClientError{Code: ErrorCodeTimeout, Retryable: true}, code: ErrorCodeTimeout},
		{name: "rate limited", err: &ClientError{Code: "rate_limited", Retryable: true}, code: "rate_limited"},
		{name: "server error", err: &ClientError{Code: "temporary_failure", Retryable: true}, code: "temporary_failure"},
		{name: "unauthorized", err: &ClientError{Code: ErrorCodeInvalidInternalToken, Retryable: true}, code: ErrorCodeInvalidInternalToken},
		{name: "malformed response", err: &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true}, code: ErrorCodeMalformedResponse},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			job := testClaimedJob(now)
			job.AttemptCount = 2
			store := &workerStoreStub{job: &job}
			client := &advanceClientStub{err: test.err}
			worker := NewWorker(
				discardLifecycleLogger(), store, client, fixedClock{now: now}, nil,
				ExponentialBackoff{Min: time.Second, Max: 10 * time.Second},
				30*time.Second, time.Second,
			)

			processed, err := worker.ProcessNext(context.Background())

			if err != nil || !processed {
				t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
			}
			if store.retryCalls != 1 || store.retryJob != job || store.retryCode != test.code ||
				!store.nextAttempt.Equal(now.Add(4*time.Second)) || !store.retryNow.Equal(now) {
				t.Fatalf("retry ack = calls:%d job:%#v code:%q next:%s now:%s",
					store.retryCalls, store.retryJob, store.retryCode, store.nextAttempt, store.retryNow)
			}
			if store.successCalls != 0 || store.terminalCalls != 0 {
				t.Fatalf("unexpected success/terminal calls = %d/%d", store.successCalls, store.terminalCalls)
			}
		})
	}
}

func TestWorkerProcessNextMarksPermanentFailuresTerminal(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	for _, code := range []string{ErrorCodeEventNotFound, ErrorCodeInvalidEventID, ErrorCodeInvalidLifecycle} {
		code := code
		t.Run(code, func(t *testing.T) {
			job := testClaimedJob(now)
			store := &workerStoreStub{job: &job}
			client := &advanceClientStub{err: &ClientError{Code: code, Terminal: true}}
			worker := NewWorker(discardLifecycleLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{}, 30*time.Second, time.Second)

			processed, err := worker.ProcessNext(context.Background())

			if err != nil || !processed {
				t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
			}
			if store.terminalCalls != 1 || store.terminalJob != job || store.terminalState != StateTerminalFailed ||
				store.terminalCode != code || !store.terminalNow.Equal(now) {
				t.Fatalf("terminal ack = calls:%d job:%#v state:%q code:%q now:%s",
					store.terminalCalls, store.terminalJob, store.terminalState, store.terminalCode, store.terminalNow)
			}
			if store.successCalls != 0 || store.retryCalls != 0 {
				t.Fatalf("unexpected success/retry calls = %d/%d", store.successCalls, store.retryCalls)
			}
		})
	}
}

func TestWorkerRunCancelsActiveRequestAndWaitsForItToReturn(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	job := testClaimedJob(now)
	store := &workerStoreStub{job: &job}
	client := &advanceClientStub{started: make(chan struct{}), waitForCtx: true}
	worker := NewWorker(discardLifecycleLogger(), store, client, fixedClock{now: now}, newControlledDelaySource(), ExponentialBackoff{}, 30*time.Second, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	select {
	case <-client.started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start active lifecycle request")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not wait for cancelled active request to return")
	}
	if store.successCalls != 0 || store.retryCalls != 0 || store.terminalCalls != 0 {
		t.Fatalf("cancelled request was acknowledged: success=%d retry=%d terminal=%d",
			store.successCalls, store.retryCalls, store.terminalCalls)
	}
}

func TestTransitionAfterSuccessCoversStartEndCatchUpAndIdempotentOutcomes(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	result := testAdvanceResult("", now)
	tests := []struct {
		outcome   string
		wantState string
		wantNext  *time.Time
	}{
		{outcome: OutcomeStarted, wantState: StateActiveWait, wantNext: &result.EndTime},
		{outcome: OutcomeAlreadyActive, wantState: StateActiveWait, wantNext: &result.EndTime},
		{outcome: OutcomeCompleted, wantState: StateCompleted},
		{outcome: OutcomeCompletedCatchUp, wantState: StateCompleted},
		{outcome: OutcomeAlreadyCompleted, wantState: StateCompleted},
		{outcome: OutcomeNotDue, wantState: StateScheduled, wantNext: &result.StartTime},
	}
	for _, test := range tests {
		t.Run(test.outcome, func(t *testing.T) {
			result := result
			result.Outcome = test.outcome
			state, next, err := transitionAfterSuccess(testClaimedJob(now), result, now)
			if err != nil || state != test.wantState || !equalOptionalTime(next, test.wantNext) {
				t.Fatalf("transition = state:%q next:%v error:%v, want %q/%v/nil", state, next, err, test.wantState, test.wantNext)
			}
		})
	}

	pastResult := result
	pastResult.Outcome = OutcomeNotDue
	pastResult.StartTime = now.Add(-time.Minute)
	state, next, err := transitionAfterSuccess(testClaimedJob(now), pastResult, now)
	wantNext := now.Add(time.Second)
	if err != nil || state != StateScheduled || next == nil || !next.Equal(wantNext) {
		t.Fatalf("past not_due transition = state:%q next:%v error:%v, want scheduled/%s/nil", state, next, err, wantNext)
	}

	result.Outcome = "invented"
	if _, _, err := transitionAfterSuccess(testClaimedJob(now), result, now); err == nil {
		t.Fatal("unknown outcome was accepted")
	}
}

func TestExponentialBackoffIsPersistentAttemptBasedAndBounded(t *testing.T) {
	t.Parallel()

	backoff := ExponentialBackoff{Min: time.Second, Max: 5 * time.Second}
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second}
	for index, expected := range want {
		if got := backoff.Delay(index + 1); got != expected {
			t.Errorf("Delay(%d) = %s, want %s", index+1, got, expected)
		}
	}
}

func testClaimedJob(now time.Time) Job {
	leaseUntil := now.Add(30 * time.Second)
	return Job{
		ID:              7,
		EventID:         42,
		SourceSequence:  11,
		SourceMessageID: "00000000-0000-0000-0000-000000000011",
		State:           StateProcessing,
		AttemptCount:    0,
		LeaseUntil:      &leaseUntil,
		LeaseToken:      "00000000-0000-0000-0000-000000000007",
	}
}

func testAdvanceResult(outcome string, now time.Time) AdvanceResult {
	return AdvanceResult{
		EventID:     42,
		Outcome:     outcome,
		StartTime:   now.Add(time.Hour),
		EndTime:     now.Add(3 * time.Hour),
		ProcessedAt: now,
	}
}

func equalOptionalTime(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

var _ workerStore = (*workerStoreStub)(nil)
var _ advanceClient = (*advanceClientStub)(nil)
var _ Clock = fixedClock{}
