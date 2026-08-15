package tests

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	servicesevents "friendship/services/events"
)

type lifecycleClockStub struct{ now time.Time }

func (c lifecycleClockStub) Now() time.Time { return c.now }

type lifecycleTransactionStub struct {
	snapshot       servicesevents.EventLifecycleSnapshot
	loadErr        error
	resolveErr     error
	updateErr      error
	statusIDs      map[string]uint
	loadCtx        context.Context
	resolveCtx     context.Context
	updateCtx      context.Context
	loadedEventID  uint
	resolvedStatus string
	updatedEventID uint
	updatedStatus  uint
	updateCalls    int
}

func (tx *lifecycleTransactionStub) LoadEventForUpdate(ctx context.Context, eventID uint) (servicesevents.EventLifecycleSnapshot, error) {
	tx.loadCtx = ctx
	tx.loadedEventID = eventID
	return tx.snapshot, tx.loadErr
}

func (tx *lifecycleTransactionStub) ResolveStatusID(ctx context.Context, statusName string) (uint, error) {
	tx.resolveCtx = ctx
	tx.resolvedStatus = statusName
	return tx.statusIDs[statusName], tx.resolveErr
}

func (tx *lifecycleTransactionStub) UpdateEventStatus(ctx context.Context, eventID uint, statusID uint) error {
	tx.updateCtx = ctx
	tx.updatedEventID = eventID
	tx.updatedStatus = statusID
	tx.updateCalls++
	return tx.updateErr
}

type lifecycleStoreStub struct {
	tx            servicesevents.EventLifecycleTransaction
	err           error
	ctx           context.Context
	eventID       uint
	calls         int
	callbackCalls int
}

func (s *lifecycleStoreStub) WithinLifecycleTransaction(ctx context.Context, eventID uint, fn func(servicesevents.EventLifecycleTransaction) error) error {
	s.calls++
	s.ctx = ctx
	s.eventID = eventID
	if s.err != nil {
		return s.err
	}
	s.callbackCalls++
	return fn(s.tx)
}

func newLifecycleTransaction(now time.Time, status string) *lifecycleTransactionStub {
	return &lifecycleTransactionStub{
		snapshot: servicesevents.EventLifecycleSnapshot{
			ID:         123,
			StatusID:   41,
			StatusName: status,
			StartTime:  now.Add(-time.Hour),
			EndTime:    now.Add(time.Hour),
		},
		statusIDs: map[string]uint{
			eventmodels.StatusActive:    52,
			eventmodels.StatusCompleted: 63,
		},
	}
}

func TestEventLifecycleServiceDecisions(t *testing.T) {
	now := time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		status        string
		startTime     time.Time
		endTime       time.Time
		wantOutcome   string
		wantStatus    string
		wantApplied   bool
		wantUpdateID  uint
		wantUpdateNum int
	}{
		{"before start", eventmodels.StatusRecruitment, now.Add(time.Second), now.Add(time.Hour), servicesevents.LifecycleOutcomeNotDue, eventmodels.StatusRecruitment, false, 0, 0},
		{"exactly at start", eventmodels.StatusRecruitment, now, now.Add(time.Hour), servicesevents.LifecycleOutcomeStarted, eventmodels.StatusActive, true, 52, 1},
		{"between start and end", eventmodels.StatusRecruitment, now.Add(-time.Second), now.Add(time.Second), servicesevents.LifecycleOutcomeStarted, eventmodels.StatusActive, true, 52, 1},
		{"exactly at end", eventmodels.StatusActive, now.Add(-time.Hour), now, servicesevents.LifecycleOutcomeCompleted, eventmodels.StatusCompleted, true, 63, 1},
		{"after end", eventmodels.StatusActive, now.Add(-2 * time.Hour), now.Add(-time.Second), servicesevents.LifecycleOutcomeCompleted, eventmodels.StatusCompleted, true, 63, 1},
		{"catch up recruitment exactly at end", eventmodels.StatusRecruitment, now.Add(-time.Hour), now, servicesevents.LifecycleOutcomeCompletedCatchUp, eventmodels.StatusCompleted, true, 63, 1},
		{"catch up recruitment after end", eventmodels.StatusRecruitment, now.Add(-2 * time.Hour), now.Add(-time.Second), servicesevents.LifecycleOutcomeCompletedCatchUp, eventmodels.StatusCompleted, true, 63, 1},
		{"repeated after start", eventmodels.StatusActive, now.Add(-time.Hour), now.Add(time.Hour), servicesevents.LifecycleOutcomeAlreadyActive, eventmodels.StatusActive, false, 0, 0},
		{"repeated after completion", eventmodels.StatusCompleted, now.Add(-2 * time.Hour), now.Add(-time.Hour), servicesevents.LifecycleOutcomeAlreadyCompleted, eventmodels.StatusCompleted, false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := newLifecycleTransaction(now, tt.status)
			tx.snapshot.StartTime = tt.startTime
			tx.snapshot.EndTime = tt.endTime
			store := &lifecycleStoreStub{tx: tx}
			service := servicesevents.NewEventLifecycleService(store, lifecycleClockStub{now: now})
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "lifecycle")

			result, err := service.Advance(ctx, 123)
			if err != nil {
				t.Fatalf("Advance() error = %v", err)
			}
			if result.EventID != 123 || result.PreviousStatus != tt.status || result.CurrentStatus != tt.wantStatus ||
				result.Outcome != tt.wantOutcome || result.Applied != tt.wantApplied || !result.ProcessedAt.Equal(now) ||
				!result.StartTime.Equal(tt.startTime) || !result.EndTime.Equal(tt.endTime) {
				t.Fatalf("Advance() result = %#v", result)
			}
			if store.ctx != ctx || tx.loadCtx != ctx {
				t.Fatal("request context was not preserved through store and transaction")
			}
			if tx.updateCalls != tt.wantUpdateNum || tx.updatedStatus != tt.wantUpdateID {
				t.Fatalf("status updates = %d with status ID %d, want %d with status ID %d", tx.updateCalls, tx.updatedStatus, tt.wantUpdateNum, tt.wantUpdateID)
			}
			if tt.wantApplied && (tx.resolveCtx != ctx || tx.updateCtx != ctx || tx.updatedEventID != 123) {
				t.Fatal("transition lookup/update did not preserve context or event ID")
			}
		})
	}
}

func TestEventLifecycleServiceErrorsCauseNoSuccessfulTransition(t *testing.T) {
	now := time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC)
	loadFailure := errors.New("load failed")
	resolveFailure := errors.New("status lookup failed")
	updateFailure := errors.New("save failed")
	storeFailure := errors.New("transaction failed")
	tests := []struct {
		name       string
		configure  func(*lifecycleStoreStub, *lifecycleTransactionStub)
		wantErr    error
		wantUpdate int
	}{
		{"event not found", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) {
			tx.loadErr = servicesevents.ErrEventNotFound
		}, servicesevents.ErrEventNotFound, 0},
		{"event load failure", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) { tx.loadErr = loadFailure }, loadFailure, 0},
		{"unknown current status", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) {
			tx.snapshot.StatusName = "Неизвестен"
		}, servicesevents.ErrInvalidEventLifecycleState, 0},
		{"invalid time range", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) { tx.snapshot.EndTime = tx.snapshot.StartTime }, servicesevents.ErrInvalidEventLifecycleState, 0},
		{"target status lookup failure", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) { tx.resolveErr = resolveFailure }, resolveFailure, 0},
		{"target status missing", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) { tx.statusIDs[eventmodels.StatusActive] = 0 }, servicesevents.ErrInvalidEventLifecycleState, 0},
		{"status save failure rolls back result", func(_ *lifecycleStoreStub, tx *lifecycleTransactionStub) { tx.updateErr = updateFailure }, updateFailure, 1},
		{"transaction failure", func(store *lifecycleStoreStub, _ *lifecycleTransactionStub) { store.err = storeFailure }, storeFailure, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := newLifecycleTransaction(now, eventmodels.StatusRecruitment)
			store := &lifecycleStoreStub{tx: tx}
			tt.configure(store, tx)
			service := servicesevents.NewEventLifecycleService(store, lifecycleClockStub{now: now})

			result, err := service.Advance(context.Background(), 123)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Advance() error = %v, want errors.Is(_, %v)", err, tt.wantErr)
			}
			if !reflect.DeepEqual(result, servicesevents.EventLifecycleResult{}) {
				t.Fatalf("Advance() result on error = %#v, want zero result", result)
			}
			if tx.updateCalls != tt.wantUpdate {
				t.Fatalf("UpdateEventStatus calls = %d, want %d", tx.updateCalls, tt.wantUpdate)
			}
		})
	}
}

// This serializing fake proves application-level retry behavior without claiming
// SQLite or an in-memory mutex proves PostgreSQL row-lock semantics. The separate
// PostgreSQL integration test below owns that database-specific evidence.
func TestEventLifecycleServiceConcurrentCallsApplyTransitionOnceWithSerializingStore(t *testing.T) {
	now := time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC)
	store := &serialLifecycleStore{
		snapshot: servicesevents.EventLifecycleSnapshot{
			ID: 123, StatusID: 41, StatusName: eventmodels.StatusRecruitment,
			StartTime: now, EndTime: now.Add(time.Hour),
		},
		statusIDs: map[string]uint{eventmodels.StatusActive: 52, eventmodels.StatusCompleted: 63},
	}
	service := servicesevents.NewEventLifecycleService(store, lifecycleClockStub{now: now})
	results := make(chan servicesevents.EventLifecycleResult, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := service.Advance(context.Background(), 123)
			results <- result
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Advance() error = %v", err)
		}
	}
	applied := 0
	for result := range results {
		if result.Applied {
			applied++
		}
	}
	if applied != 1 || store.updateCalls != 1 || store.snapshot.StatusName != eventmodels.StatusActive {
		t.Fatalf("applied results/updates/final status = %d/%d/%q, want 1/1/%q", applied, store.updateCalls, store.snapshot.StatusName, eventmodels.StatusActive)
	}
}

type serialLifecycleStore struct {
	mu          sync.Mutex
	snapshot    servicesevents.EventLifecycleSnapshot
	statusIDs   map[string]uint
	updateCalls int
}

func (s *serialLifecycleStore) WithinLifecycleTransaction(ctx context.Context, _ uint, fn func(servicesevents.EventLifecycleTransaction) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn((*serialLifecycleTx)(s))
}

type serialLifecycleTx serialLifecycleStore

func (tx *serialLifecycleTx) LoadEventForUpdate(context.Context, uint) (servicesevents.EventLifecycleSnapshot, error) {
	return tx.snapshot, nil
}

func (tx *serialLifecycleTx) ResolveStatusID(_ context.Context, status string) (uint, error) {
	return tx.statusIDs[status], nil
}

func (tx *serialLifecycleTx) UpdateEventStatus(_ context.Context, _ uint, statusID uint) error {
	tx.updateCalls++
	tx.snapshot.StatusID = statusID
	for name, id := range tx.statusIDs {
		if id == statusID {
			tx.snapshot.StatusName = name
			break
		}
	}
	return nil
}

var _ servicesevents.EventLifecycleStore = (*lifecycleStoreStub)(nil)
var _ servicesevents.EventLifecycleStore = (*serialLifecycleStore)(nil)
