package reminders

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
	client := NewHTTPClient(cfg.MonolithBaseURL, cfg.InternalToken, cfg.MonolithHTTPTimeout, transport)
	registry, err := NewChannelRegistry(InAppChannel{})
	if err != nil {
		panic(err)
	}
	store := NewStore(db, registry)
	clock := SystemClock{}
	delaySource := RealDelaySource{}
	return &Manager{
		logger: logger,
		store:  store,
		poller: NewPoller(
			logger,
			store,
			client,
			clock,
			delaySource,
			ExponentialBackoff{Min: cfg.SourceRetryMinBackoff, Max: cfg.SourceRetryMaxBackoff},
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
			ExponentialBackoff{Min: cfg.WorkerRetryMinBackoff, Max: cfg.WorkerRetryMaxBackoff},
			cfg.JobLeaseDuration,
			cfg.WorkerScanInterval,
		),
		done: make(chan struct{}),
	}
}

func (m *Manager) Start(parent context.Context) error {
	m.startOnce.Do(func() {
		if parent == nil {
			m.startErr = errors.New("reminder manager parent context не задан")
			return
		}
		if m.store == nil {
			m.startErr = errors.New("reminder manager store не инициализирован")
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
		return errors.New("истекло время остановки reminder manager")
	}
}

func (m *Manager) PingContext(ctx context.Context) error {
	if m == nil || m.store == nil {
		return errors.New("reminder manager не инициализирован")
	}
	return m.store.PingContext(ctx)
}

func (m *Manager) ListNotifications(ctx context.Context, userID uint64, after string, limit int, unreadOnly bool) (NotificationPage, error) {
	return m.store.ListNotifications(ctx, userID, after, limit, unreadOnly)
}

func (m *Manager) CountUnread(ctx context.Context, userID uint64) (int, error) {
	return m.store.CountUnread(ctx, userID)
}

func (m *Manager) MarkAsRead(ctx context.Context, userID uint64, notificationID string) (NotificationRecord, error) {
	return m.store.MarkAsRead(ctx, userID, notificationID, time.Now().UTC())
}
