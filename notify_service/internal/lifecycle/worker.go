package lifecycle

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type Worker struct {
	logger        *slog.Logger
	store         workerStore
	client        advanceClient
	clock         Clock
	delaySource   DelaySource
	retryBackoff  ExponentialBackoff
	leaseDuration time.Duration
	scanInterval  time.Duration
}

type workerStore interface {
	ClaimDueJob(context.Context, time.Time, time.Duration) (*Job, error)
	RescheduleAfterSuccess(context.Context, Job, AdvanceResult, time.Time) error
	RescheduleRetry(context.Context, Job, time.Time, string, time.Time) error
	MarkTerminal(context.Context, Job, string, string, time.Time) error
}

func NewWorker(logger *slog.Logger, store workerStore, client advanceClient, clock Clock, delaySource DelaySource, retryBackoff ExponentialBackoff, leaseDuration time.Duration, scanInterval time.Duration) *Worker {
	return &Worker{
		logger:        logger,
		store:         store,
		client:        client,
		clock:         clock,
		delaySource:   delaySource,
		retryBackoff:  retryBackoff,
		leaseDuration: leaseDuration,
		scanInterval:  scanInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		processed, err := w.ProcessNext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			w.logger.Warn("lifecycle worker step failed", "error_code", errorCode(err))
			if !waitForDelay(ctx, w.delaySource, w.retryBackoff.Delay(1)) {
				return
			}
			continue
		}
		delay := w.scanInterval
		if processed {
			delay = 0
		}
		if !waitForDelay(ctx, w.delaySource, delay) {
			return
		}
	}
}

func (w *Worker) ProcessNext(ctx context.Context) (bool, error) {
	now := w.clock.Now()
	job, err := w.store.ClaimDueJob(ctx, now, w.leaseDuration)
	if err != nil || job == nil {
		return false, err
	}

	started := w.clock.Now()
	result, err := w.client.AdvanceLifecycle(ctx, job.EventID)
	duration := w.clock.Now().Sub(started)
	if err == nil {
		if updateErr := w.store.RescheduleAfterSuccess(ctx, *job, result, w.clock.Now()); updateErr != nil {
			return true, updateErr
		}
		w.logger.Info(
			"lifecycle job applied",
			"job_id", job.ID,
			"event_id", job.EventID,
			"source_sequence", job.SourceSequence,
			"outcome", result.Outcome,
			"attempt", job.AttemptCount+1,
			"duration_ms", duration.Milliseconds(),
		)
		return true, nil
	}
	if errors.Is(err, context.Canceled) {
		return true, err
	}

	clientErr := &ClientError{Code: ErrorCodeUnexpectedStatus, Retryable: true, Cause: err}
	var typedErr *ClientError
	if errors.As(err, &typedErr) {
		clientErr = typedErr
	}
	if clientErr.Terminal {
		if markErr := w.store.MarkTerminal(ctx, *job, StateTerminalFailed, clientErr.Code, w.clock.Now()); markErr != nil {
			return true, markErr
		}
		w.logger.Warn(
			"lifecycle job terminal failure",
			"job_id", job.ID,
			"event_id", job.EventID,
			"source_sequence", job.SourceSequence,
			"error_code", clientErr.Code,
			"attempt", job.AttemptCount+1,
			"duration_ms", duration.Milliseconds(),
		)
		return true, nil
	}

	nextAttempt := w.clock.Now().Add(w.retryBackoff.Delay(job.AttemptCount + 1))
	if retryErr := w.store.RescheduleRetry(ctx, *job, nextAttempt, clientErr.Code, w.clock.Now()); retryErr != nil {
		return true, retryErr
	}
	w.logger.Warn(
		"lifecycle job scheduled for retry",
		"job_id", job.ID,
		"event_id", job.EventID,
		"source_sequence", job.SourceSequence,
		"error_code", clientErr.Code,
		"attempt", job.AttemptCount+1,
		"duration_ms", duration.Milliseconds(),
		"retry_at", nextAttempt,
	)
	return true, nil
}
