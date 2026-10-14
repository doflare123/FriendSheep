package statistics

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"
)

type Backoff struct{ Min, Max time.Duration }

func (b Backoff) Delay(attempt int) time.Duration {
	base := b.Min
	if base <= 0 {
		base = time.Second
	}
	maximum := b.Max
	if maximum < base {
		maximum = base
	}
	for i := 1; i < attempt && base < maximum; i++ {
		if base > maximum/2 {
			base = maximum
			break
		}
		base *= 2
	}
	if base > maximum {
		base = maximum
	}
	// Jitter в пределах [base/2, base] не превышает настроенный максимум.
	return base/2 + time.Duration(rand.Int64N(int64(base-base/2)+1))
}

type processClient interface {
	ProcessEvent(context.Context, uint64) (ProcessResult, error)
}
type workerStore interface {
	ClaimDueJob(context.Context, time.Time, time.Duration) (*Job, error)
	Complete(context.Context, Job, ProcessResult, time.Time) error
	Retry(context.Context, Job, time.Time, string, time.Time) error
}

type Worker struct {
	logger          *slog.Logger
	store           workerStore
	client          processClient
	clock           Clock
	backoff         Backoff
	lease, interval time.Duration
}

func NewWorker(logger *slog.Logger, store workerStore, client processClient, clock Clock, backoff Backoff, lease, interval time.Duration) *Worker {
	return &Worker{logger: logger, store: store, client: client, clock: clock, backoff: backoff, lease: lease, interval: interval}
}

func (w *Worker) ProcessNext(ctx context.Context) (bool, error) {
	job, err := w.store.ClaimDueJob(ctx, w.clock.Now(), w.lease)
	if err != nil || job == nil {
		return false, err
	}
	result, err := w.client.ProcessEvent(ctx, job.EventID)
	if errors.Is(err, context.Canceled) {
		return true, err
	}
	if err == nil && result.Outcome != "not_completed" {
		if err = w.store.Complete(ctx, *job, result, w.clock.Now()); err != nil {
			return true, err
		}
		w.logger.Info("statistics job завершён", "job_id", job.ID, "event_id", job.EventID, "source_sequence", job.SourceSequence, "outcome", result.Outcome)
		return true, nil
	}
	errorCode := "not_completed"
	if err != nil {
		errorCode = code(err)
	}
	now := w.clock.Now()
	next := now.Add(w.backoff.Delay(job.AttemptCount + 1))
	if err = w.store.Retry(ctx, *job, next, errorCode, now); err != nil {
		return true, err
	}
	w.logger.Warn("statistics job ожидает повтор", "job_id", job.ID, "event_id", job.EventID, "error_code", errorCode, "attempt", job.AttemptCount+1, "retry_at", next)
	return true, nil
}

func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		processed, err := w.ProcessNext(ctx)
		delay := w.interval
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			w.logger.Warn("statistics worker step failed", "error_code", code(err))
			delay = w.backoff.Delay(1)
		} else if processed {
			delay = 0
		}
		if !wait(ctx, delay) {
			return
		}
	}
}
