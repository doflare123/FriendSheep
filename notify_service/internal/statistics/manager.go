package statistics

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"notify_service/internal/config"
)

type Manager struct {
	poller  *Poller
	worker  *Worker
	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	started bool
	stopped bool
}

func NewManager(cfg config.Config, logger *slog.Logger, db *sql.DB, transport http.RoundTripper) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	store := NewStore(db)
	sourceClient := NewHTTPClient(cfg.MonolithBaseURL, cfg.InternalToken, cfg.MonolithHTTPTimeout, transport)
	commandClient := NewHTTPClient(cfg.MonolithBaseURL, cfg.InternalToken, cfg.StatisticsHTTPTimeout, transport)
	clock := SystemClock{}
	return &Manager{
		poller: NewPoller(logger, store, sourceClient, clock, Backoff{Min: cfg.SourceRetryMinBackoff, Max: cfg.SourceRetryMaxBackoff}, cfg.SourceBatchLimit, cfg.SourcePollInterval),
		worker: NewWorker(logger, store, commandClient, clock, Backoff{Min: cfg.WorkerRetryMinBackoff, Max: cfg.WorkerRetryMaxBackoff}, cfg.StatisticsJobLeaseDuration, cfg.WorkerScanInterval),
		done:   make(chan struct{}),
	}
}

func (m *Manager) Start(parent context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if parent == nil || m.poller == nil || m.worker == nil {
		return errors.New("statistics manager не инициализирован")
	}
	if m.stopped {
		return errors.New("statistics manager уже остановлен")
	}
	if m.started {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	m.cancel = cancel
	m.started = true
	go func() {
		defer close(m.done)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); m.poller.Run(ctx) }()
		go func() { defer wg.Done(); m.worker.Run(ctx) }()
		wg.Wait()
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
		return errors.New("истекло время остановки statistics manager")
	}
}
