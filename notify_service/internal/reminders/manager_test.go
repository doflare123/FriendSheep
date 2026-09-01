package reminders

import (
	"context"
	"testing"
	"time"
)

type idleReminderStore struct{}

func (idleReminderStore) LoadCursor(context.Context, string) (uint64, error) {
	return 0, nil
}

func (idleReminderStore) ApplySourceEvents(_ context.Context, _ string, _ []ReminderIntent, _ time.Time) (ApplyResult, error) {
	return ApplyResult{}, nil
}

func (idleReminderStore) ClaimDueJob(context.Context, time.Time, time.Duration) (*Job, error) {
	return nil, nil
}

func (idleReminderStore) MaterializeJob(context.Context, Job, RecipientSnapshot, time.Time) (int, error) {
	return 0, nil
}

func (idleReminderStore) RescheduleRetry(context.Context, Job, time.Time, string, time.Time) error {
	return nil
}

func (idleReminderStore) MarkFinal(context.Context, Job, string, string, time.Time) error {
	return nil
}

func (idleReminderStore) ClaimDeliveryTarget(context.Context, time.Time, time.Duration) (*DeliveryTarget, error) {
	return nil, nil
}

func (idleReminderStore) MarkDeliveryDelivered(context.Context, DeliveryTarget, DeliveryResult, time.Time) error {
	return nil
}

func (idleReminderStore) RescheduleDelivery(context.Context, DeliveryTarget, time.Time, string, time.Time) error {
	return nil
}

func (idleReminderStore) MarkDeliveryTerminal(context.Context, DeliveryTarget, string, time.Time) error {
	return nil
}

type emptyReminderSource struct{}

func (emptyReminderSource) ListReminderIntents(_ context.Context, after uint64, _ int) (IntentPage, error) {
	return IntentPage{Items: []ReminderIntent{}, NextCursor: after}, nil
}

func TestReminderManagerStopsPollerWorkerAndDispatcherAfterCancellation(t *testing.T) {
	t.Parallel()

	store := idleReminderStore{}
	clock := fixedClock{now: time.Date(2036, 8, 31, 12, 0, 0, 0, time.UTC)}
	registry, err := NewChannelRegistry(InAppChannel{})
	if err != nil {
		t.Fatalf("NewChannelRegistry(): %v", err)
	}
	manager := &Manager{
		store:      &Store{},
		poller:     NewPoller(discardReminderLogger(), store, emptyReminderSource{}, clock, RealDelaySource{}, ExponentialBackoff{Min: time.Second, Max: time.Minute}, SourceName, 10, time.Hour),
		worker:     NewWorker(discardReminderLogger(), store, &recipientClientStub{}, clock, RealDelaySource{}, ExponentialBackoff{Min: time.Second, Max: time.Minute}, time.Minute, time.Hour),
		dispatcher: NewDispatcher(discardReminderLogger(), store, registry, clock, RealDelaySource{}, ExponentialBackoff{Min: time.Second, Max: time.Minute}, time.Minute, time.Hour),
		done:       make(chan struct{}),
	}

	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := manager.Stop(time.Second); err != nil {
		t.Fatalf("Stop(): %v", err)
	}
	select {
	case <-manager.done:
	default:
		t.Fatal("manager.done не закрыт после Stop")
	}
}
