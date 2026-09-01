package reminders

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type dispatcherStoreStub struct {
	mu            sync.Mutex
	targets       []*DeliveryTarget
	claimActive   bool
	deliveredIDs  []int64
	retryIDs      []int64
	terminalIDs   []int64
	deliveredErrs []error
	retryAt       time.Time
}

func (s *dispatcherStoreStub) ClaimDeliveryTarget(_ context.Context, _ time.Time, _ time.Duration) (*DeliveryTarget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claimActive = true
	defer func() { s.claimActive = false }()
	if len(s.targets) == 0 {
		return nil, nil
	}
	target := s.targets[0]
	s.targets = s.targets[1:]
	return target, nil
}

func (s *dispatcherStoreStub) MarkDeliveryDelivered(_ context.Context, target DeliveryTarget, _ DeliveryResult, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveredIDs = append(s.deliveredIDs, target.ID)
	if len(s.deliveredErrs) == 0 {
		return nil
	}
	err := s.deliveredErrs[0]
	s.deliveredErrs = s.deliveredErrs[1:]
	return err
}

func (s *dispatcherStoreStub) RescheduleDelivery(_ context.Context, target DeliveryTarget, nextAttempt time.Time, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.retryIDs = append(s.retryIDs, target.ID)
	s.retryAt = nextAttempt
	return nil
}

func (s *dispatcherStoreStub) MarkDeliveryTerminal(_ context.Context, target DeliveryTarget, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.terminalIDs = append(s.terminalIDs, target.ID)
	return nil
}

type dispatcherChannelStub struct {
	code    string
	deliver func(context.Context, DeliveryRequest) (DeliveryResult, error)
}

func (c dispatcherChannelStub) Code() string { return c.code }

func (c dispatcherChannelStub) Deliver(ctx context.Context, request DeliveryRequest) (DeliveryResult, error) {
	return c.deliver(ctx, request)
}

func TestDispatcherCallsChannelOnlyAfterClaimReturns(t *testing.T) {
	store := &dispatcherStoreStub{targets: []*DeliveryTarget{deliveryTargetFixture(1, "external", "stable-key", "lease-1")}}
	called := false
	channel := dispatcherChannelStub{code: "external", deliver: func(_ context.Context, _ DeliveryRequest) (DeliveryResult, error) {
		store.mu.Lock()
		claimActive := store.claimActive
		store.mu.Unlock()
		if claimActive {
			t.Fatal("channel adapter вызван внутри claim-транзакции")
		}
		called = true
		return DeliveryResult{ProviderMessageID: "provider-1"}, nil
	}}
	dispatcher := newTestDispatcher(t, store, channel)

	processed, err := dispatcher.ProcessNext(context.Background())
	if err != nil || !processed || !called {
		t.Fatalf("ProcessNext() = processed:%t called:%t error:%v", processed, called, err)
	}
}

func TestDispatcherReusesIdempotencyKeyAfterProviderSuccessAndAckFailure(t *testing.T) {
	first := deliveryTargetFixture(2, "external", "stable-provider-key", "lease-1")
	second := deliveryTargetFixture(2, "external", "stable-provider-key", "lease-2")
	store := &dispatcherStoreStub{
		targets:       []*DeliveryTarget{first, second},
		deliveredErrs: []error{errors.New("injected ack failure"), nil},
	}
	var keys []string
	channel := dispatcherChannelStub{code: "external", deliver: func(_ context.Context, request DeliveryRequest) (DeliveryResult, error) {
		keys = append(keys, request.IdempotencyKey)
		return DeliveryResult{ProviderMessageID: "provider-accepted"}, nil
	}}
	dispatcher := newTestDispatcher(t, store, channel)

	if processed, err := dispatcher.ProcessNext(context.Background()); !processed || err == nil {
		t.Fatalf("first ProcessNext() = processed:%t error:%v, want ack failure", processed, err)
	}
	if processed, err := dispatcher.ProcessNext(context.Background()); !processed || err != nil {
		t.Fatalf("second ProcessNext() = processed:%t error:%v", processed, err)
	}
	if len(keys) != 2 || keys[0] != "stable-provider-key" || keys[1] != keys[0] {
		t.Fatalf("delivery idempotency keys = %#v", keys)
	}
}

func TestDispatcherRetryForOneChannelDoesNotBlockAnother(t *testing.T) {
	now := time.Date(2036, 9, 1, 12, 0, 0, 0, time.UTC)
	store := &dispatcherStoreStub{targets: []*DeliveryTarget{
		deliveryTargetFixture(3, "retrying", "retry-key", "lease-1"),
		deliveryTargetFixture(4, "healthy", "healthy-key", "lease-2"),
	}}
	retrying := dispatcherChannelStub{code: "retrying", deliver: func(context.Context, DeliveryRequest) (DeliveryResult, error) {
		return DeliveryResult{}, &DeliveryError{Code: ErrorCodeTimeout, Retryable: true}
	}}
	healthy := dispatcherChannelStub{code: "healthy", deliver: func(context.Context, DeliveryRequest) (DeliveryResult, error) {
		return DeliveryResult{ProviderMessageID: "ok"}, nil
	}}
	dispatcher := newTestDispatcherAt(t, store, now, retrying, healthy)

	if processed, err := dispatcher.ProcessNext(context.Background()); !processed || err != nil {
		t.Fatalf("retrying ProcessNext() = processed:%t error:%v", processed, err)
	}
	if processed, err := dispatcher.ProcessNext(context.Background()); !processed || err != nil {
		t.Fatalf("healthy ProcessNext() = processed:%t error:%v", processed, err)
	}
	if len(store.retryIDs) != 1 || store.retryIDs[0] != 3 || len(store.deliveredIDs) != 1 || store.deliveredIDs[0] != 4 {
		t.Fatalf("retry/delivered IDs = %#v/%#v", store.retryIDs, store.deliveredIDs)
	}
	if !store.retryAt.Equal(now.Add(2 * time.Second)) {
		t.Fatalf("retryAt = %s, want %s", store.retryAt, now.Add(2*time.Second))
	}
}

func TestDispatcherTerminalFailureKeepsMaterializedNotification(t *testing.T) {
	target := deliveryTargetFixture(5, "terminal", "terminal-key", "lease-1")
	target.Notification.ID = "inbox-still-present"
	store := &dispatcherStoreStub{targets: []*DeliveryTarget{target}}
	channel := dispatcherChannelStub{code: "terminal", deliver: func(context.Context, DeliveryRequest) (DeliveryResult, error) {
		return DeliveryResult{}, &DeliveryError{Code: "rejected", Terminal: true}
	}}
	dispatcher := newTestDispatcher(t, store, channel)

	if processed, err := dispatcher.ProcessNext(context.Background()); !processed || err != nil {
		t.Fatalf("ProcessNext() = processed:%t error:%v", processed, err)
	}
	if len(store.terminalIDs) != 1 || target.Notification.ID != "inbox-still-present" {
		t.Fatalf("terminal IDs/notification = %#v/%q", store.terminalIDs, target.Notification.ID)
	}
}

func TestDispatcherStopsActiveDeliveryAfterCancellation(t *testing.T) {
	store := &dispatcherStoreStub{targets: []*DeliveryTarget{deliveryTargetFixture(6, "blocking", "blocking-key", "lease-1")}}
	started := make(chan struct{})
	channel := dispatcherChannelStub{code: "blocking", deliver: func(ctx context.Context, _ DeliveryRequest) (DeliveryResult, error) {
		close(started)
		<-ctx.Done()
		return DeliveryResult{}, ctx.Err()
	}}
	dispatcher := newTestDispatcher(t, store, channel)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		dispatcher.Run(ctx)
	}()
	<-started
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("dispatcher не остановился после отмены context")
	}
	if len(store.deliveredIDs) != 0 || len(store.retryIDs) != 0 || len(store.terminalIDs) != 0 {
		t.Fatalf("отменённая доставка была подтверждена: delivered=%v retry=%v terminal=%v", store.deliveredIDs, store.retryIDs, store.terminalIDs)
	}
}

func deliveryTargetFixture(id int64, channelCode, idempotencyKey, leaseToken string) *DeliveryTarget {
	return &DeliveryTarget{
		ID:             id,
		NotificationID: "notification-1",
		ChannelCode:    channelCode,
		State:          DeliveryStateProcessing,
		AttemptCount:   2,
		LeaseToken:     leaseToken,
		IdempotencyKey: idempotencyKey,
		Notification:   NotificationRecord{ID: "notification-1", UserID: 7},
	}
}

func newTestDispatcher(t *testing.T, store deliveryDispatcherStore, channels ...DeliveryChannel) *Dispatcher {
	t.Helper()
	return newTestDispatcherAt(t, store, time.Date(2036, 9, 1, 12, 0, 0, 0, time.UTC), channels...)
}

func newTestDispatcherAt(t *testing.T, store deliveryDispatcherStore, now time.Time, channels ...DeliveryChannel) *Dispatcher {
	t.Helper()
	registry, err := NewChannelRegistry(channels...)
	if err != nil {
		t.Fatalf("NewChannelRegistry(): %v", err)
	}
	return NewDispatcher(
		discardReminderLogger(),
		store,
		registry,
		fixedClock{now: now},
		RealDelaySource{},
		ExponentialBackoff{Min: time.Second, Max: time.Minute},
		30*time.Second,
		time.Hour,
	)
}
