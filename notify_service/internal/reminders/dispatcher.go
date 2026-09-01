package reminders

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type deliveryDispatcherStore interface {
	ClaimDeliveryTarget(context.Context, time.Time, time.Duration) (*DeliveryTarget, error)
	MarkDeliveryDelivered(context.Context, DeliveryTarget, DeliveryResult, time.Time) error
	RescheduleDelivery(context.Context, DeliveryTarget, time.Time, string, time.Time) error
	MarkDeliveryTerminal(context.Context, DeliveryTarget, string, time.Time) error
}

type Dispatcher struct {
	logger        *slog.Logger
	store         deliveryDispatcherStore
	channels      *ChannelRegistry
	clock         Clock
	delaySource   DelaySource
	retryBackoff  ExponentialBackoff
	leaseDuration time.Duration
	scanInterval  time.Duration
}

func NewDispatcher(logger *slog.Logger, store deliveryDispatcherStore, channels *ChannelRegistry, clock Clock, delaySource DelaySource, retryBackoff ExponentialBackoff, leaseDuration time.Duration, scanInterval time.Duration) *Dispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{
		logger:        logger,
		store:         store,
		channels:      channels,
		clock:         clock,
		delaySource:   delaySource,
		retryBackoff:  retryBackoff,
		leaseDuration: leaseDuration,
		scanInterval:  scanInterval,
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	for {
		processed, err := d.ProcessNext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			d.logger.Warn("delivery dispatcher step failed", "error_code", errorCode(err))
			if !waitForDelay(ctx, d.delaySource, d.retryBackoff.Delay(1)) {
				return
			}
			continue
		}
		delay := d.scanInterval
		if processed {
			delay = 0
		}
		if !waitForDelay(ctx, d.delaySource, delay) {
			return
		}
	}
}

func (d *Dispatcher) ProcessNext(ctx context.Context) (bool, error) {
	target, err := d.store.ClaimDeliveryTarget(ctx, d.clock.Now(), d.leaseDuration)
	if err != nil || target == nil {
		return false, err
	}

	channel, found := d.channels.Channel(target.ChannelCode)
	if !found {
		if err := d.store.MarkDeliveryTerminal(ctx, *target, ErrorCodeInvalidChannel, d.clock.Now()); err != nil {
			return true, err
		}
		return true, nil
	}

	result, deliveryErr := channel.Deliver(ctx, DeliveryRequest{
		TargetID:       target.ID,
		Notification:   target.Notification,
		IdempotencyKey: target.IdempotencyKey,
	})
	if deliveryErr == nil {
		if err := d.store.MarkDeliveryDelivered(ctx, *target, result, d.clock.Now()); err != nil {
			return true, err
		}
		d.logger.Info(
			"delivery target delivered",
			"delivery_target_id", target.ID,
			"channel", target.ChannelCode,
			"attempt", target.AttemptCount,
		)
		return true, nil
	}
	if errors.Is(deliveryErr, context.Canceled) {
		return true, deliveryErr
	}

	errorCode, retryable := classifyDeliveryError(deliveryErr)
	if !retryable {
		if err := d.store.MarkDeliveryTerminal(ctx, *target, errorCode, d.clock.Now()); err != nil {
			return true, err
		}
		d.logger.Warn(
			"delivery target terminal failure",
			"delivery_target_id", target.ID,
			"channel", target.ChannelCode,
			"attempt", target.AttemptCount,
			"error_code", errorCode,
		)
		return true, nil
	}

	now := d.clock.Now()
	nextAttempt := now.Add(d.retryBackoff.Delay(target.AttemptCount))
	if err := d.store.RescheduleDelivery(ctx, *target, nextAttempt, errorCode, now); err != nil {
		return true, err
	}
	d.logger.Warn(
		"delivery target scheduled for retry",
		"delivery_target_id", target.ID,
		"channel", target.ChannelCode,
		"attempt", target.AttemptCount,
		"error_code", errorCode,
		"retry_at", nextAttempt,
	)
	return true, nil
}

func classifyDeliveryError(err error) (string, bool) {
	var deliveryErr *DeliveryError
	if errors.As(err, &deliveryErr) {
		code := deliveryErr.Code
		if code == "" {
			code = ErrorCodeUnexpectedStatus
		}
		return code, deliveryErr.Retryable && !deliveryErr.Terminal
	}
	return ErrorCodeUnexpectedStatus, true
}
