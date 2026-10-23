package uniquepeople

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type Client interface {
	Claim(context.Context, int) (ClaimResponse, error)
	Execute(context.Context, Claim) (ExecuteResult, error)
}

type Worker struct {
	logger      *slog.Logger
	client      Client
	clock       Clock
	concurrency int
	interval    time.Duration
}

func NewWorker(logger *slog.Logger, client Client, clock Clock, concurrency int, interval time.Duration) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	if clock == nil {
		clock = SystemClock{}
	}
	return &Worker{logger: logger, client: client, clock: clock, concurrency: concurrency, interval: interval}
}

func (w *Worker) ProcessBatch(ctx context.Context) (int, error) {
	response, err := w.client.Claim(ctx, w.concurrency)
	if err != nil {
		return 0, err
	}
	workerCount := min(w.concurrency, len(response.Jobs))
	jobs := make(chan Claim)
	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				started := w.clock.Now()
				result, executeErr := w.client.Execute(ctx, job)
				duration := w.clock.Now().Sub(started)
				if executeErr != nil {
					if !errors.Is(executeErr, context.Canceled) {
						w.logger.Warn("unique people execute failed", "user_id", job.UserID, "generation", job.Generation, "error_code", errorCode(executeErr), "duration", duration)
					}
					continue
				}
				w.logger.Info("unique people job завершён", "user_id", job.UserID, "claimed_generation", job.Generation, "applied_generation", result.Generation, "outcome", result.Outcome, "duration", duration)
			}
		}()
	}
	for _, job := range response.Jobs {
		select {
		case jobs <- job:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return len(response.Jobs), ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return len(response.Jobs), err
	}
	return len(response.Jobs), nil
}

func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		claimed, err := w.ProcessBatch(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			w.logger.Warn("unique people claim failed", "error_code", errorCode(err))
		} else if claimed > 0 {
			w.logger.Info("unique people batch обработан", "claimed", claimed)
		}
		if !wait(ctx, w.interval) {
			return
		}
	}
}

func wait(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
