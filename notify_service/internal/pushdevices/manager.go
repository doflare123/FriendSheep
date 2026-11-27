package pushdevices

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type managedStore interface {
	admissionStore
	reconciliationStore
}

// Manager владеет фоновым отзывом и предоставляет точку допуска для будущего
// push-диспетчера. Он не создаёт цели и не выполняет внешнюю отправку.
type Manager struct {
	authorizer *AddressAuthorizer
	reconciler *Reconciler

	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	started bool
	stopped bool
}

func NewManager(logger *slog.Logger, store managedStore, checker SessionChecker, interval time.Duration, pageSize int) (*Manager, error) {
	authorizer, err := NewAddressAuthorizer(store, checker)
	if err != nil {
		return nil, err
	}
	reconciler, err := NewReconciler(logger, store, checker, interval, pageSize)
	if err != nil {
		return nil, err
	}
	return &Manager{authorizer: authorizer, reconciler: reconciler, done: make(chan struct{})}, nil
}

func (m *Manager) Start(parent context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if parent == nil || m.reconciler == nil || m.authorizer == nil {
		return errors.New("push device manager не инициализирован")
	}
	if m.stopped {
		return errors.New("push device manager уже остановлен")
	}
	if m.started {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	m.cancel = cancel
	m.started = true
	go func() {
		defer close(m.done)
		m.reconciler.Run(ctx)
	}()
	return nil
}

func (m *Manager) Stop(timeout time.Duration) error {
	m.mu.Lock()
	m.stopped = true
	started := m.started
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Unlock()
	if !started {
		return nil
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-m.done:
		return nil
	case <-timer.C:
		return errors.New("истекло время остановки push device manager")
	}
}

func (m *Manager) Authorize(ctx context.Context, installationID, bindingID string, userID uint64) (ResolvedAddress, bool, error) {
	if m == nil || m.authorizer == nil {
		return ResolvedAddress{}, false, errors.New("push device manager не инициализирован")
	}
	return m.authorizer.Authorize(ctx, installationID, bindingID, userID)
}
