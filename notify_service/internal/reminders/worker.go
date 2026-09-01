package reminders

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type Worker struct {
	logger        *slog.Logger
	store         workerStore
	client        recipientClient
	clock         Clock
	delaySource   DelaySource
	retryBackoff  ExponentialBackoff
	leaseDuration time.Duration
	scanInterval  time.Duration
}

type workerStore interface {
	ClaimDueJob(context.Context, time.Time, time.Duration) (*Job, error)
	MaterializeJob(context.Context, Job, RecipientSnapshot, time.Time) (int, error)
	RescheduleRetry(context.Context, Job, time.Time, string, time.Time) error
	MarkFinal(context.Context, Job, string, string, time.Time) error
}

func NewWorker(logger *slog.Logger, store workerStore, client recipientClient, clock Clock, delaySource DelaySource, retryBackoff ExponentialBackoff, leaseDuration time.Duration, scanInterval time.Duration) *Worker {
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
			w.logger.Warn("reminder worker step failed", "error_code", errorCode(err))
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

	windowState := evaluateJobWindow(*job, now)
	switch windowState {
	case StateExpired:
		if err := w.store.MarkFinal(ctx, *job, StateExpired, "", w.clock.Now()); err != nil {
			return true, err
		}
		w.logger.Info("reminder job expired", "job_id", job.ID, "event_id", job.EventID, "source_sequence", job.SourceRevision, "reminder_offset_minutes", job.ReminderOffsetMinutes, "attempt", job.AttemptCount+1)
		return true, nil
	case StateSkippedMissedWindow:
		if err := w.store.MarkFinal(ctx, *job, StateSkippedMissedWindow, "", w.clock.Now()); err != nil {
			return true, err
		}
		w.logger.Info("reminder job missed delivery window", "job_id", job.ID, "event_id", job.EventID, "source_sequence", job.SourceRevision, "reminder_offset_minutes", job.ReminderOffsetMinutes, "attempt", job.AttemptCount+1)
		return true, nil
	case StateScheduled:
		if err := w.store.MarkFinal(ctx, *job, StateTerminalFailed, ErrorCodeMalformedResponse, w.clock.Now()); err != nil {
			return true, err
		}
		return true, nil
	}

	started := w.clock.Now()
	snapshot, err := w.client.ResolveRecipients(ctx, job.EventID, job.ReminderOffsetMinutes)
	duration := w.clock.Now().Sub(started)
	if err == nil {
		if job.StartTime == nil || !snapshot.StartTime.Equal(*job.StartTime) {
			if markErr := w.store.MarkFinal(ctx, *job, StateSuperseded, ErrorCodeStaleSchedule, w.clock.Now()); markErr != nil {
				return true, markErr
			}
			w.logger.Info(
				"reminder job superseded by fresher schedule snapshot",
				"job_id", job.ID,
				"event_id", job.EventID,
				"source_sequence", job.SourceRevision,
				"reminder_offset_minutes", job.ReminderOffsetMinutes,
				"duration_ms", duration.Milliseconds(),
			)
			return true, nil
		}
		materializedUsers, completeErr := w.store.MaterializeJob(ctx, *job, snapshot, w.clock.Now())
		if completeErr != nil {
			return true, w.handleFailure(ctx, *job, completeErr, duration)
		}
		w.logger.Info(
			"reminder job applied",
			"job_id", job.ID,
			"event_id", job.EventID,
			"source_sequence", job.SourceRevision,
			"reminder_offset_minutes", job.ReminderOffsetMinutes,
			"recipient_count", materializedUsers,
			"outcome", StateCompleted,
			"attempt", job.AttemptCount+1,
			"duration_ms", duration.Milliseconds(),
		)
		return true, nil
	}
	if errors.Is(err, context.Canceled) {
		return true, err
	}

	return true, w.handleFailure(ctx, *job, err, duration)
}

func (w *Worker) handleFailure(ctx context.Context, job Job, err error, duration time.Duration) error {
	if errors.Is(err, context.Canceled) {
		return err
	}

	clientErr := &ClientError{Code: ErrorCodeUnexpectedStatus, Retryable: true, Cause: err}
	var typedClientErr *ClientError
	switch {
	case errors.As(err, &typedClientErr):
		clientErr = typedClientErr
	}
	if clientErr.Terminal {
		if markErr := w.store.MarkFinal(ctx, job, StateTerminalFailed, clientErr.Code, w.clock.Now()); markErr != nil {
			return markErr
		}
		w.logger.Warn(
			"reminder job terminal failure",
			"job_id", job.ID,
			"event_id", job.EventID,
			"source_sequence", job.SourceRevision,
			"reminder_offset_minutes", job.ReminderOffsetMinutes,
			"error_code", clientErr.Code,
			"attempt", job.AttemptCount+1,
			"duration_ms", duration.Milliseconds(),
		)
		return nil
	}

	nextAttempt := w.clock.Now().Add(w.retryBackoff.Delay(job.AttemptCount + 1))
	if retryErr := w.store.RescheduleRetry(ctx, job, nextAttempt, clientErr.Code, w.clock.Now()); retryErr != nil {
		return retryErr
	}
	w.logger.Warn(
		"reminder job scheduled for retry",
		"job_id", job.ID,
		"event_id", job.EventID,
		"source_sequence", job.SourceRevision,
		"reminder_offset_minutes", job.ReminderOffsetMinutes,
		"error_code", clientErr.Code,
		"attempt", job.AttemptCount+1,
		"duration_ms", duration.Milliseconds(),
		"retry_at", nextAttempt,
	)
	return nil
}
