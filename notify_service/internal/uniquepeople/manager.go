package uniquepeople

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"notify_service/internal/config"
)

type Manager struct {
	worker  *Worker
	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	started bool
	stopped bool
}

func NewManager(cfg config.Config, logger *slog.Logger, transport http.RoundTripper) *Manager {
	clock := SystemClock{}
	client := NewHTTPClient(cfg.MonolithBaseURL, cfg.InternalToken, cfg.UniquePeopleHTTPTimeout, transport, clock)
	return &Manager{
		worker: NewWorker(logger, client, clock, cfg.UniquePeopleConcurrency, cfg.UniquePeoplePollInterval),
		done:   make(chan struct{}),
	}
}

func (m *Manager) Start(parent context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if parent == nil || m.worker == nil {
		return errors.New("unique people manager не инициализирован")
	}
	if m.stopped {
		return errors.New("unique people manager уже остановлен")
	}
	if m.started {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	m.cancel = cancel
	m.started = true
	go func() {
		defer close(m.done)
		m.worker.Run(ctx)
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
		return errors.New("истекло время остановки unique people manager")
	}
}
