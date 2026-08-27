package lifecycle

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"
)

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time { return clock.now }

type pollerStoreStub struct {
	cursor      uint64
	loadErr     error
	applyResult ApplyResult
	applyErr    error
	loadCtx     context.Context
	loadSource  string
	applyCtx    context.Context
	applySource string
	applyItems  []ScheduleEvent
	applyNow    time.Time
	loadCalls   int
	applyCalls  int
}

func (store *pollerStoreStub) LoadCursor(ctx context.Context, source string) (uint64, error) {
	store.loadCalls++
	store.loadCtx = ctx
	store.loadSource = source
	return store.cursor, store.loadErr
}

func (store *pollerStoreStub) ApplySourceEvents(ctx context.Context, source string, items []ScheduleEvent, now time.Time) (ApplyResult, error) {
	store.applyCalls++
	store.applyCtx = ctx
	store.applySource = source
	store.applyItems = append([]ScheduleEvent(nil), items...)
	store.applyNow = now
	return store.applyResult, store.applyErr
}

type scheduleClientStub struct {
	page       SchedulePage
	err        error
	ctx        context.Context
	after      uint64
	limit      int
	calls      int
	callSignal chan struct{}
}

func (client *scheduleClientStub) ListScheduleEvents(ctx context.Context, after uint64, limit int) (SchedulePage, error) {
	client.calls++
	client.ctx = ctx
	client.after = after
	client.limit = limit
	if client.callSignal != nil {
		select {
		case client.callSignal <- struct{}{}:
		default:
		}
	}
	return client.page, client.err
}

type controlledDelaySource struct {
	mu       sync.Mutex
	delays   []time.Duration
	requests chan time.Duration
	channel  chan time.Time
}

func newControlledDelaySource() *controlledDelaySource {
	return &controlledDelaySource{
		requests: make(chan time.Duration, 10),
		channel:  make(chan time.Time),
	}
}

func (source *controlledDelaySource) After(delay time.Duration) <-chan time.Time {
	source.mu.Lock()
	source.delays = append(source.delays, delay)
	source.mu.Unlock()
	source.requests <- delay
	return source.channel
}

func TestPollOnceLoadsCursorFetchesAndAtomicallyAppliesBatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 12, 0, 0, 0, time.UTC)
	item := validScheduleEvent(43, 42, OperationScheduleUpsert, now)
	store := &pollerStoreStub{cursor: 42, applyResult: ApplyResult{LastSequence: 43, AppliedCount: 1}}
	client := &scheduleClientStub{page: SchedulePage{Items: []ScheduleEvent{item}, NextCursor: 43, HasMore: true}}
	poller := NewPoller(discardLifecycleLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{}, SourceName, 100, time.Second)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "preserved")

	hasMore, err := poller.PollOnce(ctx)

	if err != nil || !hasMore {
		t.Fatalf("PollOnce() = hasMore:%t error:%v, want true/nil", hasMore, err)
	}
	if store.loadCalls != 1 || store.loadCtx != ctx || store.loadSource != SourceName {
		t.Fatalf("LoadCursor call = count:%d context:%t source:%q", store.loadCalls, store.loadCtx == ctx, store.loadSource)
	}
	if client.calls != 1 || client.ctx != ctx || client.after != 42 || client.limit != 100 {
		t.Fatalf("client call = count:%d context:%t after:%d limit:%d", client.calls, client.ctx == ctx, client.after, client.limit)
	}
	if store.applyCalls != 1 || store.applyCtx != ctx || store.applySource != SourceName ||
		!reflect.DeepEqual(store.applyItems, []ScheduleEvent{item}) || !store.applyNow.Equal(now) {
		t.Fatalf("ApplySourceEvents call = count:%d context:%t source:%q items:%#v now:%s",
			store.applyCalls, store.applyCtx == ctx, store.applySource, store.applyItems, store.applyNow)
	}
}

func TestPollOnceStopsBeforeApplyOnSourceOrContractFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		loadErr   error
		clientErr error
		page      SchedulePage
	}{
		{name: "cursor failure", loadErr: errors.New("cursor failure")},
		{name: "network failure", clientErr: &ClientError{Code: ErrorCodeNetwork, Retryable: true}},
		{name: "missing items field", page: SchedulePage{NextCursor: 10}},
		{name: "empty page cannot claim more", page: SchedulePage{NextCursor: 0, HasMore: true}},
		{name: "page begins before cursor", page: SchedulePage{Items: []ScheduleEvent{validScheduleEvent(9, 1, OperationScheduleCancel, now)}, NextCursor: 9}},
		{name: "next cursor mismatch", page: SchedulePage{Items: []ScheduleEvent{validScheduleEvent(11, 1, OperationScheduleCancel, now)}, NextCursor: 12}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &pollerStoreStub{cursor: 10, loadErr: test.loadErr}
			client := &scheduleClientStub{page: test.page, err: test.clientErr}
			poller := NewPoller(discardLifecycleLogger(), store, client, fixedClock{now: now}, nil, ExponentialBackoff{}, SourceName, 100, time.Second)

			if _, err := poller.PollOnce(context.Background()); err == nil {
				t.Fatal("PollOnce() returned nil error")
			}
			if store.applyCalls != 0 {
				t.Fatalf("ApplySourceEvents called %d times after failure", store.applyCalls)
			}
			if test.loadErr != nil && client.calls != 0 {
				t.Fatalf("source client called %d times after cursor failure", client.calls)
			}
		})
	}
}

func TestPollerRunUsesControlledBackoffAndStopsOnCancellation(t *testing.T) {
	t.Parallel()

	delaySource := newControlledDelaySource()
	client := &scheduleClientStub{err: &ClientError{Code: ErrorCodeNetwork, Retryable: true}}
	poller := NewPoller(
		discardLifecycleLogger(),
		&pollerStoreStub{},
		client,
		fixedClock{},
		delaySource,
		ExponentialBackoff{Min: 2 * time.Second, Max: 10 * time.Second},
		SourceName,
		100,
		5*time.Second,
	)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	select {
	case delay := <-delaySource.requests:
		if delay != 2*time.Second {
			t.Fatalf("first retry delay = %s, want 2s", delay)
		}
	case <-time.After(time.Second):
		t.Fatal("poller did not request controlled retry delay")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop after cancellation")
	}
}

func TestPollerRunDoesNotBusyLoopAfterSuccessfulEmptyPage(t *testing.T) {
	t.Parallel()

	delaySource := newControlledDelaySource()
	client := &scheduleClientStub{page: SchedulePage{Items: []ScheduleEvent{}, NextCursor: 0}}
	poller := NewPoller(
		discardLifecycleLogger(),
		&pollerStoreStub{},
		client,
		fixedClock{},
		delaySource,
		ExponentialBackoff{},
		SourceName,
		100,
		7*time.Second,
	)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	select {
	case delay := <-delaySource.requests:
		if delay != 7*time.Second {
			t.Fatalf("poll interval = %s, want 7s", delay)
		}
	case <-time.After(time.Second):
		t.Fatal("poller did not request its poll interval")
	}
	if client.calls != 1 {
		t.Fatalf("client calls before releasing timer = %d, want 1", client.calls)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop after cancellation")
	}
}

func TestValidateSourceBatchRejectsOrderingAndMalformedPayloads(t *testing.T) {
	t.Parallel()

	now := time.Date(2036, 8, 27, 12, 0, 0, 0, time.UTC)
	upsert := validScheduleEvent(1, 42, OperationScheduleUpsert, now)
	cancel := validScheduleEvent(2, 42, OperationScheduleCancel, now)
	tests := []struct {
		name  string
		items []ScheduleEvent
	}{
		{name: "zero sequence", items: []ScheduleEvent{{Sequence: 0, EventID: 42, SchemaVersion: 1, Operation: OperationScheduleCancel}}},
		{name: "sequence over bigint", items: []ScheduleEvent{{Sequence: uint64(1 << 63), MessageID: "id", EventID: 42, OccurredAt: now, SchemaVersion: 1, Operation: OperationScheduleCancel}}},
		{name: "duplicate sequence", items: []ScheduleEvent{upsert, upsert}},
		{name: "descending sequence", items: []ScheduleEvent{cancel, upsert}},
		{name: "unsupported schema", items: []ScheduleEvent{{Sequence: 1, EventID: 42, SchemaVersion: 2, Operation: OperationScheduleCancel}}},
		{name: "zero event", items: []ScheduleEvent{{Sequence: 1, EventID: 0, SchemaVersion: 1, Operation: OperationScheduleCancel}}},
		{name: "event over bigint", items: []ScheduleEvent{{Sequence: 1, MessageID: "id", EventID: uint64(1 << 63), OccurredAt: now, SchemaVersion: 1, Operation: OperationScheduleCancel}}},
		{name: "empty message id", items: []ScheduleEvent{{Sequence: 1, EventID: 42, OccurredAt: now, SchemaVersion: 1, Operation: OperationScheduleCancel}}},
		{name: "empty occurred at", items: []ScheduleEvent{{Sequence: 1, MessageID: "id", EventID: 42, SchemaVersion: 1, Operation: OperationScheduleCancel}}},
		{name: "upsert missing schedule", items: []ScheduleEvent{{Sequence: 1, EventID: 42, SchemaVersion: 1, Operation: OperationScheduleUpsert}}},
		{name: "cancel has schedule", items: []ScheduleEvent{func() ScheduleEvent { item := cancel; item.StartTime = &now; return item }()}},
		{name: "unknown operation", items: []ScheduleEvent{{Sequence: 1, EventID: 42, SchemaVersion: 1, Operation: "unknown"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateSourceBatch(test.items); err == nil {
				t.Fatal("validateSourceBatch() returned nil error")
			}
		})
	}
	if err := validateSourceBatch([]ScheduleEvent{upsert, cancel}); err != nil {
		t.Fatalf("valid ordered batch rejected: %v", err)
	}
}

func validScheduleEvent(sequence uint64, eventID uint64, operation string, now time.Time) ScheduleEvent {
	item := ScheduleEvent{
		Sequence:      sequence,
		MessageID:     "00000000-0000-0000-0000-000000000042",
		SchemaVersion: 1,
		Operation:     operation,
		EventID:       eventID,
		OccurredAt:    now,
	}
	if operation == OperationScheduleUpsert {
		start := now.Add(time.Hour)
		end := start.Add(2 * time.Hour)
		item.StartTime = &start
		item.EndTime = &end
	}
	return item
}

func discardLifecycleLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var _ pollerStore = (*pollerStoreStub)(nil)
var _ scheduleClient = (*scheduleClientStub)(nil)
var _ DelaySource = (*controlledDelaySource)(nil)
