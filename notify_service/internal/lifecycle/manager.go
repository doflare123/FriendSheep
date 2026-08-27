package lifecycle

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
	logger *slog.Logger
	store  *Store
	poller *Poller
	worker *Worker

	runCtx    context.Context
	cancelRun context.CancelFunc
	done      chan struct{}

	startOnce sync.Once
	stopOnce  sync.Once
	startErr  error
}

func NewManager(cfg config.Config, logger *slog.Logger, db *sql.DB, transport http.RoundTripper) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	clock := SystemClock{}
	delaySource := RealDelaySource{}
	client := NewHTTPClient(cfg.MonolithBaseURL, cfg.InternalToken, cfg.MonolithHTTPTimeout, transport)
	store := NewStore(db)
	return &Manager{
		logger: logger,
		store:  store,
		poller: NewPoller(
			logger,
			store,
			client,
			clock,
			delaySource,
			ExponentialBackoff{Min: cfg.SourceRetryMinBackoff, Max: cfg.SourceRetryMaxBackoff, Jitter: NewRandomJitter(time.Now().UnixNano())},
			SourceName,
			cfg.SourceBatchLimit,
			cfg.SourcePollInterval,
		),
		worker: NewWorker(
			logger,
			store,
			client,
			clock,
			delaySource,
			ExponentialBackoff{Min: cfg.WorkerRetryMinBackoff, Max: cfg.WorkerRetryMaxBackoff, Jitter: NewRandomJitter(time.Now().UnixNano() + 1)},
			cfg.JobLeaseDuration,
			cfg.WorkerScanInterval,
		),
		done: make(chan struct{}),
	}
}

func (m *Manager) Start(parent context.Context) error {
	m.startOnce.Do(func() {
		if parent == nil {
			m.startErr = errors.New("lifecycle manager parent context не задан")
			return
		}
		if m.store == nil {
			m.startErr = errors.New("lifecycle manager store не инициализирован")
			return
		}
		runCtx, cancel := context.WithCancel(parent)
		m.runCtx = runCtx
		m.cancelRun = cancel

		go func() {
			defer close(m.done)
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				m.poller.Run(runCtx)
			}()
			go func() {
				defer wg.Done()
				m.worker.Run(runCtx)
			}()
			wg.Wait()
		}()
	})
	return m.startErr
}

func (m *Manager) Stop(timeout time.Duration) error {
	m.stopOnce.Do(func() {
		if m.cancelRun != nil {
			m.cancelRun()
		}
	})
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	select {
	case <-m.done:
		return nil
	case <-time.After(timeout):
		return errors.New("истекло время остановки lifecycle manager")
	}
}

func (m *Manager) PingContext(ctx context.Context) error {
	if m == nil || m.store == nil {
		return errors.New("lifecycle manager не инициализирован")
	}
	return m.store.PingContext(ctx)
}
