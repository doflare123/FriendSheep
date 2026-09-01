package reminders

import (
	"context"
	"fmt"
	"regexp"
	"sort"
)

var channelCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

type DeliveryResult struct {
	ProviderMessageID string
}

type DeliveryRequest struct {
	TargetID       int64
	Notification   NotificationRecord
	IdempotencyKey string
}

type DeliveryError struct {
	Code      string
	Retryable bool
	Terminal  bool
	Cause     error
}

func (e *DeliveryError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Code + ": " + e.Cause.Error()
	}
	return e.Code
}

func (e *DeliveryError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type DeliveryChannel interface {
	Code() string
	Deliver(context.Context, DeliveryRequest) (DeliveryResult, error)
}

type ChannelRegistry struct {
	byCode map[string]DeliveryChannel
}

func NewChannelRegistry(channels ...DeliveryChannel) (*ChannelRegistry, error) {
	registry := &ChannelRegistry{byCode: make(map[string]DeliveryChannel, len(channels))}
	for _, channel := range channels {
		if channel == nil {
			return nil, fmt.Errorf("delivery channel не инициализирован")
		}
		code := channel.Code()
		if !channelCodePattern.MatchString(code) {
			return nil, fmt.Errorf("delivery channel вернул некорректный code %q", code)
		}
		if _, exists := registry.byCode[code]; exists {
			return nil, fmt.Errorf("delivery channel %q зарегистрирован повторно", code)
		}
		registry.byCode[code] = channel
	}
	return registry, nil
}

func (r *ChannelRegistry) Resolve(codes []string) ([]DeliveryChannel, error) {
	unique := make(map[string]struct{}, len(codes))
	resolved := make([]DeliveryChannel, 0, len(codes))
	for _, code := range codes {
		if _, seen := unique[code]; seen {
			continue
		}
		channel, ok := r.byCode[code]
		if !ok {
			return nil, &ClientError{Code: ErrorCodeInvalidChannel, Terminal: true}
		}
		unique[code] = struct{}{}
		resolved = append(resolved, channel)
	}
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Code() < resolved[j].Code()
	})
	return resolved, nil
}

func (r *ChannelRegistry) Channel(code string) (DeliveryChannel, bool) {
	if r == nil {
		return nil, false
	}
	channel, ok := r.byCode[code]
	return channel, ok
}

type InAppChannel struct{}

func (InAppChannel) Code() string {
	return ChannelCodeInApp
}

func (InAppChannel) Deliver(context.Context, DeliveryRequest) (DeliveryResult, error) {
	return DeliveryResult{}, nil
}
