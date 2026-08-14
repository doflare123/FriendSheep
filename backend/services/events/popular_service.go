package events

import (
	"context"
	"errors"
	"fmt"
	"friendship/logger"
	"sync"
	"sync/atomic"
	"time"
)

const (
	PopularEventsCacheKey         = "popular_events:v3:top10"
	previousPopularEventsCacheKey = "popular_events:v2:top10"
	legacyPopularEventsCacheKey   = "popular_events:top10"
	CacheExpiration               = 6 * time.Hour
	TopEventsLimit                = 10
	popularEventsCronSpec         = "0 */6 * * *"
)

type popularEventsService struct {
	logger    logger.Logger
	store     PopularEventsStore
	cache     PopularEventsCache
	notifier  PopularEventsNotifier
	scheduler PopularEventsScheduler
	now       PopularEventsNow

	backgroundCtx    context.Context
	backgroundCancel context.CancelFunc

	updateMu        sync.Mutex
	refreshInFlight atomic.Bool
}

func NewPopularEventsService(
	logger logger.Logger,
	store PopularEventsStore,
	cache PopularEventsCache,
	notifier PopularEventsNotifier,
	scheduler PopularEventsScheduler,
	now PopularEventsNow,
) PopularEventsService {
	if now == nil {
		now = time.Now
	}

	backgroundCtx, cancel := context.WithCancel(context.Background())

	return &popularEventsService{
		logger:           logger,
		store:            store,
		cache:            cache,
		notifier:         notifier,
		scheduler:        scheduler,
		now:              now,
		backgroundCtx:    backgroundCtx,
		backgroundCancel: cancel,
	}
}

func (s *popularEventsService) Start() error {
	if s.scheduler == nil {
		return nil
	}

	if err := s.scheduler.Schedule(popularEventsCronSpec, func() {
		s.runRefreshAsync("scheduled refresh")
	}); err != nil {
		return fmt.Errorf("schedule popular events refresh: %w", err)
	}

	s.scheduler.Start()
	s.logger.Info("Popular events scheduler started")
	s.runRefreshAsync("initial warmup")
	return nil
}

func (s *popularEventsService) Stop() {
	s.backgroundCancel()
	if s.scheduler != nil {
		s.scheduler.Stop()
		s.logger.Info("Popular events scheduler stopped")
	}
}

func (s *popularEventsService) GetPopularEvents(ctx context.Context) (*PopularEventsSnapshot, error) {
	snapshot, err := s.cache.Get(ctx, PopularEventsCacheKey)
	if err != nil {
		if !errors.Is(err, ErrPopularEventsCacheMiss) {
			return nil, fmt.Errorf("load popular events cache: %w", err)
		}

		s.logger.Info("Popular events cache is empty, refreshing")
		if updateErr := s.refreshCacheAfterMiss(ctx); updateErr != nil {
			return nil, fmt.Errorf("popular events cache miss and refresh failed: %w", updateErr)
		}

		snapshot, err = s.cache.Get(ctx, PopularEventsCacheKey)
		if err != nil {
			return nil, fmt.Errorf("load popular events cache after refresh: %w", err)
		}
	}

	if s.now().Sub(snapshot.UpdatedAt) > CacheExpiration {
		s.runRefreshAsync("stale cache")
	}

	result := clonePopularEventsSnapshot(snapshot)
	return &result, nil
}

func (s *popularEventsService) UpdateCache(ctx context.Context) error {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	return s.updateCache(ctx)
}

func (s *popularEventsService) refreshCacheAfterMiss(ctx context.Context) error {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	_, err := s.cache.Get(ctx, PopularEventsCacheKey)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrPopularEventsCacheMiss) {
		return fmt.Errorf("recheck popular events cache: %w", err)
	}

	return s.updateCache(ctx)
}

func (s *popularEventsService) updateCache(ctx context.Context) error {
	records, err := s.store.ListTopPopularEvents(ctx, s.now(), TopEventsLimit)
	if err != nil {
		return fmt.Errorf("load popular events ranking: %w", err)
	}

	previousIDs := s.getPreviousPopularEventIDs(ctx)
	snapshot := PopularEventsSnapshot{
		Events:    popularRecordsToViews(records),
		UpdatedAt: s.now(),
		Count:     len(records),
	}

	if err := s.cache.Set(ctx, PopularEventsCacheKey, snapshot, CacheExpiration+time.Hour); err != nil {
		return fmt.Errorf("save popular events cache: %w", err)
	}

	s.logger.Info("Popular events cache updated", "count", len(snapshot.Events))

	if notifications := buildPopularEventNotifications(records, previousIDs); len(notifications) > 0 {
		s.notifyAsync(notifications)
	}

	return nil
}

func (s *popularEventsService) getPreviousPopularEventIDs(ctx context.Context) []uint {
	snapshot, err := s.cache.Get(ctx, PopularEventsCacheKey)
	if errors.Is(err, ErrPopularEventsCacheMiss) {
		snapshot, err = s.cache.Get(ctx, previousPopularEventsCacheKey)
	}
	if errors.Is(err, ErrPopularEventsCacheMiss) {
		snapshot, err = s.cache.Get(ctx, legacyPopularEventsCacheKey)
	}
	if err != nil || snapshot == nil {
		return []uint{}
	}

	ids := make([]uint, 0, len(snapshot.Events))
	for _, event := range snapshot.Events {
		ids = append(ids, event.ID)
	}
	return ids
}

func (s *popularEventsService) runRefreshAsync(reason string) {
	if !s.refreshInFlight.CompareAndSwap(false, true) {
		return
	}

	go func() {
		defer s.refreshInFlight.Store(false)

		if err := s.backgroundCtx.Err(); err != nil {
			return
		}

		s.logger.Info("Refreshing popular events asynchronously", "reason", reason)
		if err := s.UpdateCache(s.backgroundCtx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("Popular events async refresh failed", "reason", reason, "error", err)
		}
	}()
}

func (s *popularEventsService) notifyAsync(notifications []PopularEventNotification) {
	if s.notifier == nil || len(notifications) == 0 {
		return
	}

	go func() {
		if err := s.backgroundCtx.Err(); err != nil {
			return
		}

		if err := s.notifier.Notify(s.backgroundCtx, notifications); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("Popular events notification dispatch failed", "count", len(notifications), "error", err)
		}
	}()
}

func popularRecordsToViews(records []PopularEventRecord) []PopularEventView {
	views := make([]PopularEventView, 0, len(records))
	for _, record := range records {
		views = append(views, record.PopularEventView)
	}
	return views
}

func buildPopularEventNotifications(records []PopularEventRecord, previousIDs []uint) []PopularEventNotification {
	if len(records) == 0 {
		return nil
	}

	previous := make(map[uint]struct{}, len(previousIDs))
	for _, id := range previousIDs {
		previous[id] = struct{}{}
	}

	notifications := make([]PopularEventNotification, 0)
	for _, record := range records {
		if _, exists := previous[record.ID]; exists {
			continue
		}
		if record.OwnerEmail == "" {
			continue
		}

		notifications = append(notifications, PopularEventNotification{
			EventID:        record.ID,
			EventName:      record.Title,
			GroupName:      record.GroupName,
			RecipientEmail: record.OwnerEmail,
			StartTime:      record.StartTime,
		})
	}

	return notifications
}

func clonePopularEventsSnapshot(snapshot *PopularEventsSnapshot) PopularEventsSnapshot {
	if snapshot == nil {
		return PopularEventsSnapshot{}
	}

	result := PopularEventsSnapshot{
		UpdatedAt: snapshot.UpdatedAt,
		Count:     snapshot.Count,
	}
	if len(snapshot.Events) == 0 {
		return result
	}

	result.Events = make([]PopularEventView, len(snapshot.Events))
	copy(result.Events, snapshot.Events)
	for i := range result.Events {
		if len(snapshot.Events[i].Genres) == 0 {
			continue
		}
		result.Events[i].Genres = append([]string(nil), snapshot.Events[i].Genres...)
	}

	return result
}
