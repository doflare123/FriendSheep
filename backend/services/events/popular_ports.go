package events

import (
	"context"
	"errors"
	"time"
)

var ErrPopularEventsCacheMiss = errors.New("popular events cache miss")

type PopularEventsService interface {
	GetPopularEvents(ctx context.Context) (*PopularEventsSnapshot, error)
	UpdateCache(ctx context.Context) error
	Start() error
	Stop()
}

type PopularEventsStore interface {
	ListTopPopularEvents(ctx context.Context, now time.Time, limit int) ([]PopularEventRecord, error)
}

type PopularEventsCache interface {
	Get(ctx context.Context, key string) (*PopularEventsSnapshot, error)
	Set(ctx context.Context, key string, snapshot PopularEventsSnapshot, expiration time.Duration) error
}

type PopularEventsNotifier interface {
	Notify(ctx context.Context, notifications []PopularEventNotification) error
}

type PopularEventsScheduler interface {
	Schedule(spec string, job func()) error
	Start()
	Stop()
}

type PopularEventsNow func() time.Time
