package statistics

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type sourceClient interface {
	ListTriggers(context.Context, int64, int) (Page, error)
}
type pollerStore interface {
	LoadCursor(context.Context, string) (int64, error)
	ApplyTriggers(context.Context, string, []Trigger, time.Time) (ApplyResult, error)
}

type Poller struct {
	logger   *slog.Logger
	store    pollerStore
	client   sourceClient
	clock    Clock
	backoff  Backoff
	limit    int
	interval time.Duration
}

func NewPoller(logger *slog.Logger, store pollerStore, client sourceClient, clock Clock, backoff Backoff, limit int, interval time.Duration) *Poller {
	return &Poller{logger: logger, store: store, client: client, clock: clock, backoff: backoff, limit: limit, interval: interval}
}

func (p *Poller) PollOnce(ctx context.Context) (bool, error) {
	cursor, err := p.store.LoadCursor(ctx, SourceName)
	if err != nil {
		return false, err
	}
	page, err := p.client.ListTriggers(ctx, cursor, p.limit)
	if err != nil {
		return false, err
	}
	if err = validatePage(page, cursor, p.limit); err != nil {
		return false, err
	}
	result, err := p.store.ApplyTriggers(ctx, SourceName, page.Messages, p.clock.Now())
	if err != nil {
		return false, err
	}
	if len(page.Messages) > 0 {
		p.logger.Info("statistics batch сохранён", "source_sequence", result.LastSequence, "messages", len(page.Messages), "applied", result.AppliedCount, "has_more", page.HasMore)
	}
	return page.HasMore, nil
}

func (p *Poller) Run(ctx context.Context) {
	failures := 0
	for ctx.Err() == nil {
		more, err := p.PollOnce(ctx)
		delay := p.interval
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			failures++
			delay = p.backoff.Delay(failures)
			p.logger.Warn("statistics source poll failed", "error_code", code(err), "attempt", failures)
		} else {
			failures = 0
			if more {
				delay = 0
			}
		}
		if !wait(ctx, delay) {
			return
		}
	}
}
