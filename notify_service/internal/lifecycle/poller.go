package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type Poller struct {
	logger      *slog.Logger
	store       pollerStore
	client      scheduleClient
	clock       Clock
	delaySource DelaySource
	backoff     ExponentialBackoff
	source      string
	limit       int
	interval    time.Duration
}

type pollerStore interface {
	LoadCursor(context.Context, string) (uint64, error)
	ApplySourceEvents(context.Context, string, []ScheduleEvent, time.Time) (ApplyResult, error)
}

func NewPoller(logger *slog.Logger, store pollerStore, client scheduleClient, clock Clock, delaySource DelaySource, backoff ExponentialBackoff, source string, limit int, interval time.Duration) *Poller {
	return &Poller{
		logger:      logger,
		store:       store,
		client:      client,
		clock:       clock,
		delaySource: delaySource,
		backoff:     backoff,
		source:      source,
		limit:       limit,
		interval:    interval,
	}
}

func (p *Poller) Run(ctx context.Context) {
	failures := 0
	for {
		hasMore, err := p.PollOnce(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			failures++
			delay := p.backoff.Delay(failures)
			p.logger.Warn("poll lifecycle source failed", "error_code", errorCode(err), "attempt", failures, "retry_in_ms", delay.Milliseconds())
			if !waitForDelay(ctx, p.delaySource, delay) {
				return
			}
			continue
		}
		failures = 0
		delay := p.interval
		if hasMore {
			delay = 0
		}
		if !waitForDelay(ctx, p.delaySource, delay) {
			return
		}
	}
}

func (p *Poller) PollOnce(ctx context.Context) (bool, error) {
	started := p.clock.Now()
	cursor, err := p.store.LoadCursor(ctx, p.source)
	if err != nil {
		return false, err
	}
	page, err := p.client.ListScheduleEvents(ctx, cursor, p.limit)
	if err != nil {
		return false, err
	}
	if err := validateSchedulePage(page, cursor); err != nil {
		return false, err
	}
	result, err := p.store.ApplySourceEvents(ctx, p.source, page.Items, p.clock.Now())
	if err != nil {
		return false, err
	}
	p.logger.Info(
		"lifecycle source batch applied",
		"source_sequence", result.LastSequence,
		"items", len(page.Items),
		"applied", result.AppliedCount,
		"has_more", page.HasMore,
		"duration_ms", p.clock.Now().Sub(started).Milliseconds(),
	)
	for _, item := range page.Items {
		p.logger.Info(
			"lifecycle source event durably ingested",
			"event_id", item.EventID,
			"source_sequence", item.Sequence,
			"operation", item.Operation,
			"outcome", "applied_or_replayed",
		)
	}
	return page.HasMore, nil
}

func validateSchedulePage(page SchedulePage, after uint64) error {
	if page.Items == nil {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("items отсутствует в schedule response")}
	}
	if page.HasMore && len(page.Items) == 0 {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("hasMore=true при пустом batch")}
	}
	if len(page.Items) == 0 {
		if page.NextCursor != after {
			return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("пустой batch вернул неожиданный nextCursor")}
		}
		return nil
	}
	if page.Items[0].Sequence <= after {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("batch начинается не после after cursor")}
	}
	if page.NextCursor != page.Items[len(page.Items)-1].Sequence {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("nextCursor не совпадает с последней sequence")}
	}
	return nil
}

func waitForDelay(ctx context.Context, delaySource DelaySource, delay time.Duration) bool {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}
	if delaySource == nil {
		delaySource = RealDelaySource{}
	}
	select {
	case <-ctx.Done():
		return false
	case <-delaySource.After(delay):
		return true
	}
}

func errorCode(err error) string {
	var clientErr *ClientError
	if errors.As(err, &clientErr) && clientErr.Code != "" {
		return clientErr.Code
	}
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%T", err)
}
