package tests

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	servicesevents "friendship/services/events"
)

type popularEventsServiceContextKey struct{}

type popularEventsStoreStub struct {
	mu      sync.Mutex
	records []servicesevents.PopularEventRecord
	err     error
	calls   int
	ctx     context.Context
	now     time.Time
	limit   int
	callCh  chan struct{}
	release <-chan struct{}
}

func (s *popularEventsStoreStub) ListTopPopularEvents(ctx context.Context, now time.Time, limit int) ([]servicesevents.PopularEventRecord, error) {
	s.mu.Lock()
	s.calls++
	s.ctx = ctx
	s.now = now
	s.limit = limit
	records := append([]servicesevents.PopularEventRecord(nil), s.records...)
	err := s.err
	callCh := s.callCh
	s.mu.Unlock()

	if callCh != nil {
		callCh <- struct{}{}
	}
	if release := s.release; release != nil {
		<-release
	}
	return records, err
}

func (s *popularEventsStoreStub) snapshot() (int, context.Context, time.Time, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls, s.ctx, s.now, s.limit
}

type popularEventsCacheSetCall struct {
	ctx        context.Context
	key        string
	snapshot   servicesevents.PopularEventsSnapshot
	expiration time.Duration
}

type popularEventsCacheStub struct {
	mu             sync.Mutex
	snapshot       *servicesevents.PopularEventsSnapshot
	snapshotsByKey map[string]*servicesevents.PopularEventsSnapshot
	getErr         error
	getErrs        []error
	setErr         error
	getCalls       int
	getCtxs        []context.Context
	getKeys        []string
	setCalls       []popularEventsCacheSetCall
	setCh          chan struct{}
}

func (c *popularEventsCacheStub) Get(ctx context.Context, key string) (*servicesevents.PopularEventsSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.getCalls++
	c.getCtxs = append(c.getCtxs, ctx)
	c.getKeys = append(c.getKeys, key)
	if len(c.getErrs) > 0 {
		err := c.getErrs[0]
		c.getErrs = c.getErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	if c.getErr != nil {
		return nil, c.getErr
	}
	if c.snapshotsByKey != nil {
		snapshot := c.snapshotsByKey[key]
		if snapshot == nil {
			return nil, servicesevents.ErrPopularEventsCacheMiss
		}

		cloned := clonePopularEventsTestSnapshot(*snapshot)
		return &cloned, nil
	}
	if c.snapshot == nil {
		return nil, servicesevents.ErrPopularEventsCacheMiss
	}

	snapshot := clonePopularEventsTestSnapshot(*c.snapshot)
	return &snapshot, nil
}

func (c *popularEventsCacheStub) Set(
	ctx context.Context,
	key string,
	snapshot servicesevents.PopularEventsSnapshot,
	expiration time.Duration,
) error {
	c.mu.Lock()
	call := popularEventsCacheSetCall{
		ctx:        ctx,
		key:        key,
		snapshot:   clonePopularEventsTestSnapshot(snapshot),
		expiration: expiration,
	}
	c.setCalls = append(c.setCalls, call)
	err := c.setErr
	if err == nil {
		stored := clonePopularEventsTestSnapshot(snapshot)
		c.snapshot = &stored
	}
	setCh := c.setCh
	c.mu.Unlock()

	if setCh != nil {
		setCh <- struct{}{}
	}
	return err
}

func (c *popularEventsCacheStub) callsSnapshot() (int, []context.Context, []string, []popularEventsCacheSetCall) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getCalls,
		append([]context.Context(nil), c.getCtxs...),
		append([]string(nil), c.getKeys...),
		append([]popularEventsCacheSetCall(nil), c.setCalls...)
}

type popularEventsNotifierStub struct {
	mu            sync.Mutex
	err           error
	calls         int
	ctx           context.Context
	notifications []servicesevents.PopularEventNotification
	callCh        chan struct{}
}

func (n *popularEventsNotifierStub) Notify(ctx context.Context, notifications []servicesevents.PopularEventNotification) error {
	n.mu.Lock()
	n.calls++
	n.ctx = ctx
	n.notifications = append([]servicesevents.PopularEventNotification(nil), notifications...)
	err := n.err
	callCh := n.callCh
	n.mu.Unlock()

	if callCh != nil {
		callCh <- struct{}{}
	}
	return err
}

func (n *popularEventsNotifierStub) snapshot() (int, context.Context, []servicesevents.PopularEventNotification) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.calls, n.ctx, append([]servicesevents.PopularEventNotification(nil), n.notifications...)
}

type popularEventsSchedulerStub struct {
	scheduleErr error
	spec        string
	job         func()
	startCalls  int
	stopCalls   int
}

func (s *popularEventsSchedulerStub) Schedule(spec string, job func()) error {
	s.spec = spec
	s.job = job
	return s.scheduleErr
}

func (s *popularEventsSchedulerStub) Start() {
	s.startCalls++
}

func (s *popularEventsSchedulerStub) Stop() {
	s.stopCalls++
}

func TestPopularEventsServiceCacheHitReturnsDefensiveCopyAndPropagatesContext(t *testing.T) {
	now := time.Date(2031, 3, 4, 5, 6, 7, 0, time.UTC)
	cached := servicesevents.PopularEventsSnapshot{
		Events: []servicesevents.PopularEventView{{
			ID:     17,
			Title:  "Board games",
			Genres: []string{"Strategy", "Party"},
		}},
		UpdatedAt: now.Add(-time.Hour),
		Count:     1,
	}
	cache := &popularEventsCacheStub{snapshot: &cached}
	store := &popularEventsStoreStub{}
	service := newPopularEventsTestService(store, cache, nil, nil, now)
	defer service.Stop()
	ctx := context.WithValue(context.Background(), popularEventsServiceContextKey{}, "cache-hit")

	result, err := service.GetPopularEvents(ctx)

	if err != nil {
		t.Fatalf("GetPopularEvents returned error: %v", err)
	}
	if !reflect.DeepEqual(*result, cached) {
		t.Fatalf("result = %#v, want %#v", *result, cached)
	}
	result.Events[0].Title = "changed"
	result.Events[0].Genres[0] = "changed"
	if cached.Events[0].Title != "Board games" || cached.Events[0].Genres[0] != "Strategy" {
		t.Fatalf("service returned cache-owned memory: cached = %#v", cached)
	}

	storeCalls, _, _, _ := store.snapshot()
	if storeCalls != 0 {
		t.Fatalf("ranking store calls = %d, want 0 on cache hit", storeCalls)
	}
	getCalls, getCtxs, getKeys, setCalls := cache.callsSnapshot()
	if getCalls != 1 || len(setCalls) != 0 {
		t.Fatalf("cache get/set calls = %d/%d, want 1/0", getCalls, len(setCalls))
	}
	if getCtxs[0] != ctx || getCtxs[0].Value(popularEventsServiceContextKey{}) != "cache-hit" {
		t.Fatal("request context was not propagated to popular-events cache")
	}
	if getKeys[0] != servicesevents.PopularEventsCacheKey {
		t.Fatalf("cache key = %q, want %q", getKeys[0], servicesevents.PopularEventsCacheKey)
	}
}

func TestPopularEventsServiceCacheMissRefreshesSynchronouslyAndReturnsStoredSnapshot(t *testing.T) {
	now := time.Date(2032, 4, 5, 6, 7, 8, 0, time.UTC)
	store := &popularEventsStoreStub{records: []servicesevents.PopularEventRecord{{
		PopularEventView: servicesevents.PopularEventView{
			ID:        21,
			Title:     "Fresh event",
			StartTime: now.Add(24 * time.Hour),
			Genres:    []string{"Quest"},
		},
	}}}
	cache := &popularEventsCacheStub{}
	service := newPopularEventsTestService(store, cache, nil, nil, now)
	defer service.Stop()
	ctx := context.WithValue(context.Background(), popularEventsServiceContextKey{}, "cache-miss")

	result, err := service.GetPopularEvents(ctx)

	if err != nil {
		t.Fatalf("GetPopularEvents returned error: %v", err)
	}
	if result.Count != 1 || len(result.Events) != 1 || result.Events[0].ID != 21 {
		t.Fatalf("result = %#v, want freshly ranked event", result)
	}
	if !result.UpdatedAt.Equal(now) {
		t.Fatalf("UpdatedAt = %v, want %v", result.UpdatedAt, now)
	}

	storeCalls, storeCtx, storeNow, limit := store.snapshot()
	if storeCalls != 1 || storeCtx != ctx || !storeNow.Equal(now) || limit != servicesevents.TopEventsLimit {
		t.Fatalf("ranking call = calls:%d ctx:%v now:%v limit:%d", storeCalls, storeCtx, storeNow, limit)
	}
	getCalls, getCtxs, getKeys, setCalls := cache.callsSnapshot()
	if getCalls != 6 {
		t.Fatalf("cache get calls = %d, want 6 (miss, locked recheck, v3/v2/legacy previous IDs, post-refresh read)", getCalls)
	}
	if len(setCalls) != 1 {
		t.Fatalf("cache set calls = %d, want 1", len(setCalls))
	}
	for index, getCtx := range getCtxs {
		if getCtx != ctx {
			t.Fatalf("cache get context %d was not the request context", index)
		}
		wantKey := servicesevents.PopularEventsCacheKey
		if index == 3 {
			wantKey = "popular_events:v2:top10"
		}
		if index == 4 {
			wantKey = "popular_events:top10"
		}
		if getKeys[index] != wantKey {
			t.Fatalf("cache get key %d = %q", index, getKeys[index])
		}
	}
	if setCalls[0].ctx != ctx {
		t.Fatal("cache set did not receive request context")
	}
}

func TestPopularEventsServiceConcurrentCacheMissesShareOneRefresh(t *testing.T) {
	now := time.Date(2032, 5, 6, 7, 8, 9, 0, time.UTC)
	releaseStore := make(chan struct{})
	store := &popularEventsStoreStub{
		records: []servicesevents.PopularEventRecord{{
			PopularEventView: servicesevents.PopularEventView{
				ID:    22,
				Title: "Shared refresh",
			},
		}},
		callCh:  make(chan struct{}, 1),
		release: releaseStore,
	}
	cache := &popularEventsCacheStub{}
	service := newPopularEventsTestService(store, cache, nil, nil, now)
	defer service.Stop()

	type result struct {
		snapshot *servicesevents.PopularEventsSnapshot
		err      error
	}
	results := make(chan result, 2)
	load := func() {
		snapshot, err := service.GetPopularEvents(context.Background())
		results <- result{snapshot: snapshot, err: err}
	}

	go load()
	waitPopularEventsSignal(t, store.callCh, "first cold-cache refresh")
	go load()
	waitPopularEventsCacheGets(t, cache, 3)
	close(releaseStore)

	for range 2 {
		got := <-results
		if got.err != nil {
			t.Fatalf("GetPopularEvents returned error: %v", got.err)
		}
		if got.snapshot == nil || got.snapshot.Count != 1 ||
			len(got.snapshot.Events) != 1 || got.snapshot.Events[0].ID != 22 {
			t.Fatalf("result = %#v, want shared refreshed snapshot", got.snapshot)
		}
	}

	storeCalls, _, _, _ := store.snapshot()
	if storeCalls != 1 {
		t.Fatalf("ranking store calls = %d, want one shared refresh", storeCalls)
	}
	_, _, _, setCalls := cache.callsSnapshot()
	if len(setCalls) != 1 {
		t.Fatalf("cache set calls = %d, want one shared refresh write", len(setCalls))
	}
}

func TestPopularEventsServiceStaleCacheReturnsImmediatelyAndRefreshesAsynchronously(t *testing.T) {
	now := time.Date(2033, 5, 6, 7, 8, 9, 0, time.UTC)
	stale := servicesevents.PopularEventsSnapshot{
		Events:    []servicesevents.PopularEventView{{ID: 31, Title: "Stale"}},
		UpdatedAt: now.Add(-servicesevents.CacheExpiration - time.Minute),
		Count:     1,
	}
	store := &popularEventsStoreStub{
		records: []servicesevents.PopularEventRecord{{
			PopularEventView: servicesevents.PopularEventView{ID: 32, Title: "Refreshed"},
		}},
		callCh: make(chan struct{}, 1),
	}
	cache := &popularEventsCacheStub{
		snapshot: &stale,
		setCh:    make(chan struct{}, 1),
	}
	service := newPopularEventsTestService(store, cache, nil, nil, now)
	defer service.Stop()

	result, err := service.GetPopularEvents(context.Background())

	if err != nil {
		t.Fatalf("GetPopularEvents returned error: %v", err)
	}
	if result.Events[0].ID != 31 {
		t.Fatalf("result event ID = %d, want stale cached event 31", result.Events[0].ID)
	}
	waitPopularEventsSignal(t, store.callCh, "asynchronous ranking refresh")
	waitPopularEventsSignal(t, cache.setCh, "asynchronous cache write")

	_, _, _, setCalls := cache.callsSnapshot()
	if len(setCalls) != 1 || len(setCalls[0].snapshot.Events) != 1 || setCalls[0].snapshot.Events[0].ID != 32 {
		t.Fatalf("async cache writes = %#v, want refreshed event 32", setCalls)
	}
	if !setCalls[0].snapshot.UpdatedAt.Equal(now) {
		t.Fatalf("async UpdatedAt = %v, want %v", setCalls[0].snapshot.UpdatedAt, now)
	}
}

func TestPopularEventsServiceUpdateCacheWritesExactSnapshotTTLAndNotifications(t *testing.T) {
	now := time.Date(2034, 6, 7, 8, 9, 10, 0, time.UTC)
	previous := servicesevents.PopularEventsSnapshot{
		Events:    []servicesevents.PopularEventView{{ID: 40}},
		UpdatedAt: now.Add(-time.Hour),
		Count:     1,
	}
	records := []servicesevents.PopularEventRecord{
		{
			PopularEventView: servicesevents.PopularEventView{
				ID:           40,
				Title:        "Already popular",
				Group:        servicesevents.PopularEventGroupView{ID: 4, Name: "Existing group"},
				CurrentUsers: 8,
				MaxUsers:     10,
				StartTime:    now.Add(24 * time.Hour),
				Genres:       []string{"Existing"},
			},
			GroupName:   "Existing group",
			OwnerEmail:  "existing@example.com",
			OwnerUserID: 400,
			StartTime:   now.Add(24 * time.Hour),
		},
		{
			PopularEventView: servicesevents.PopularEventView{
				ID:           41,
				Title:        "Newly popular",
				Group:        servicesevents.PopularEventGroupView{ID: 5, Name: "New group"},
				CurrentUsers: 9,
				MaxUsers:     10,
				StartTime:    now.Add(48 * time.Hour),
				Genres:       []string{"New"},
				Subscribed:   true,
			},
			GroupName:   "New group",
			OwnerEmail:  "owner@example.com",
			OwnerUserID: 500,
			StartTime:   now.Add(48 * time.Hour),
		},
		{
			PopularEventView: servicesevents.PopularEventView{
				ID:        42,
				Title:     "No recipient",
				Group:     servicesevents.PopularEventGroupView{ID: 6, Name: "No-mail group"},
				StartTime: now.Add(72 * time.Hour),
			},
			GroupName: "No-mail group",
			StartTime: now.Add(72 * time.Hour),
		},
	}
	store := &popularEventsStoreStub{records: records}
	cache := &popularEventsCacheStub{snapshot: &previous}
	notifier := &popularEventsNotifierStub{callCh: make(chan struct{}, 1)}
	service := newPopularEventsTestService(store, cache, notifier, nil, now)
	defer service.Stop()
	ctx := context.WithValue(context.Background(), popularEventsServiceContextKey{}, "refresh")

	if err := service.UpdateCache(ctx); err != nil {
		t.Fatalf("UpdateCache returned error: %v", err)
	}
	waitPopularEventsSignal(t, notifier.callCh, "popular-event notification")

	storeCalls, storeCtx, storeNow, limit := store.snapshot()
	if storeCalls != 1 || storeCtx != ctx || !storeNow.Equal(now) || limit != servicesevents.TopEventsLimit {
		t.Fatalf("ranking call = calls:%d ctx:%v now:%v limit:%d", storeCalls, storeCtx, storeNow, limit)
	}
	_, _, _, setCalls := cache.callsSnapshot()
	if len(setCalls) != 1 {
		t.Fatalf("cache set calls = %d, want 1", len(setCalls))
	}
	setCall := setCalls[0]
	if setCall.ctx != ctx ||
		setCall.key != servicesevents.PopularEventsCacheKey ||
		setCall.expiration != servicesevents.CacheExpiration+time.Hour {
		t.Fatalf("cache set call = %#v, want request ctx, canonical key and seven-hour TTL", setCall)
	}
	wantViews := []servicesevents.PopularEventView{
		records[0].PopularEventView,
		records[1].PopularEventView,
		records[2].PopularEventView,
	}
	if setCall.snapshot.Count != len(records) ||
		!setCall.snapshot.UpdatedAt.Equal(now) ||
		!reflect.DeepEqual(setCall.snapshot.Events, wantViews) {
		t.Fatalf("cached snapshot = %#v, want exact records/count/time", setCall.snapshot)
	}

	notifyCalls, notifyCtx, notifications := notifier.snapshot()
	if notifyCalls != 1 {
		t.Fatalf("notifier calls = %d, want 1", notifyCalls)
	}
	if notifyCtx == nil || notifyCtx == ctx {
		t.Fatal("asynchronous notifier should receive the service lifecycle context")
	}
	wantNotifications := []servicesevents.PopularEventNotification{{
		EventID:        41,
		EventName:      "Newly popular",
		GroupName:      "New group",
		RecipientEmail: "owner@example.com",
		StartTime:      now.Add(48 * time.Hour),
	}}
	if !reflect.DeepEqual(notifications, wantNotifications) {
		t.Fatalf("notifications = %#v, want %#v", notifications, wantNotifications)
	}
}

func TestPopularEventsServiceUsesPreviousCacheIDsDuringV3Cutover(t *testing.T) {
	now := time.Date(2034, 6, 7, 8, 9, 10, 0, time.UTC)
	previous := servicesevents.PopularEventsSnapshot{
		Events: []servicesevents.PopularEventView{{ID: 81}},
		Count:  1,
	}

	tests := []struct {
		name        string
		fallbackKey string
		wantGetKeys []string
	}{
		{
			name:        "v2 snapshot",
			fallbackKey: "popular_events:v2:top10",
			wantGetKeys: []string{"popular_events:v3:top10", "popular_events:v2:top10"},
		},
		{
			name:        "legacy snapshot",
			fallbackKey: "popular_events:top10",
			wantGetKeys: []string{"popular_events:v3:top10", "popular_events:v2:top10", "popular_events:top10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &popularEventsStoreStub{records: []servicesevents.PopularEventRecord{{
				PopularEventView: servicesevents.PopularEventView{ID: 81, Title: "Already popular"},
				OwnerEmail:       "owner@example.com",
			}}}
			cache := &popularEventsCacheStub{snapshotsByKey: map[string]*servicesevents.PopularEventsSnapshot{
				tt.fallbackKey: &previous,
			}}
			notifier := &popularEventsNotifierStub{callCh: make(chan struct{}, 1)}
			service := newPopularEventsTestService(store, cache, notifier, nil, now)
			defer service.Stop()

			if err := service.UpdateCache(context.Background()); err != nil {
				t.Fatalf("UpdateCache returned error: %v", err)
			}

			select {
			case <-notifier.callCh:
				t.Fatalf("event from %s was notified again during v3 cache cutover", tt.fallbackKey)
			case <-time.After(50 * time.Millisecond):
			}

			_, _, getKeys, setCalls := cache.callsSnapshot()
			if !reflect.DeepEqual(getKeys, tt.wantGetKeys) {
				t.Fatalf("cache get keys = %#v, want %#v", getKeys, tt.wantGetKeys)
			}
			if len(setCalls) != 1 || setCalls[0].key != servicesevents.PopularEventsCacheKey {
				t.Fatalf("cache set calls = %#v, want one v3 write", setCalls)
			}
		})
	}
}

func TestPopularEventsServicePropagatesCacheAndRankingErrors(t *testing.T) {
	cacheReadErr := errors.New("cache unavailable")
	rankingErr := errors.New("ranking unavailable")
	cacheWriteErr := errors.New("cache write unavailable")

	tests := []struct {
		name      string
		store     *popularEventsStoreStub
		cache     *popularEventsCacheStub
		call      func(servicesevents.PopularEventsService) error
		wantErr   error
		wantStore int
		wantSet   int
	}{
		{
			name:  "cache read error is not treated as a miss",
			store: &popularEventsStoreStub{},
			cache: &popularEventsCacheStub{getErr: cacheReadErr},
			call: func(service servicesevents.PopularEventsService) error {
				_, err := service.GetPopularEvents(context.Background())
				return err
			},
			wantErr: cacheReadErr,
		},
		{
			name:  "ranking error aborts miss refresh",
			store: &popularEventsStoreStub{err: rankingErr},
			cache: &popularEventsCacheStub{},
			call: func(service servicesevents.PopularEventsService) error {
				_, err := service.GetPopularEvents(context.Background())
				return err
			},
			wantErr:   rankingErr,
			wantStore: 1,
		},
		{
			name:  "cache write error is propagated",
			store: &popularEventsStoreStub{records: []servicesevents.PopularEventRecord{{PopularEventView: servicesevents.PopularEventView{ID: 51}}}},
			cache: &popularEventsCacheStub{setErr: cacheWriteErr},
			call: func(service servicesevents.PopularEventsService) error {
				return service.UpdateCache(context.Background())
			},
			wantErr:   cacheWriteErr,
			wantStore: 1,
			wantSet:   1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newPopularEventsTestService(test.store, test.cache, nil, nil, time.Now())
			defer service.Stop()

			err := test.call(service)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err = %v, want errors.Is(..., %v)", err, test.wantErr)
			}
			storeCalls, _, _, _ := test.store.snapshot()
			if storeCalls != test.wantStore {
				t.Fatalf("ranking store calls = %d, want %d", storeCalls, test.wantStore)
			}
			_, _, _, setCalls := test.cache.callsSnapshot()
			if len(setCalls) != test.wantSet {
				t.Fatalf("cache set calls = %d, want %d", len(setCalls), test.wantSet)
			}
		})
	}
}

func TestPopularEventsServiceDelegatesSchedulerLifecycleWithoutRealCron(t *testing.T) {
	now := time.Date(2035, 7, 8, 9, 10, 11, 0, time.UTC)
	store := &popularEventsStoreStub{
		records: []servicesevents.PopularEventRecord{},
		callCh:  make(chan struct{}, 2),
	}
	cache := &popularEventsCacheStub{setCh: make(chan struct{}, 2)}
	scheduler := &popularEventsSchedulerStub{}
	service := newPopularEventsTestService(store, cache, nil, scheduler, now)

	if err := service.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if scheduler.spec != "0 */6 * * *" || scheduler.job == nil {
		t.Fatalf("scheduled spec/job = %q/%v", scheduler.spec, scheduler.job != nil)
	}
	if scheduler.startCalls != 1 {
		t.Fatalf("scheduler start calls = %d, want 1", scheduler.startCalls)
	}
	waitPopularEventsSignal(t, store.callCh, "initial warmup ranking")
	waitPopularEventsSignal(t, cache.setCh, "initial warmup cache write")

	triggerPopularEventsJobUntilSignal(t, scheduler.job, store.callCh, "scheduled ranking")
	waitPopularEventsSignal(t, cache.setCh, "scheduled cache write")

	service.Stop()
	if scheduler.stopCalls != 1 {
		t.Fatalf("scheduler stop calls = %d, want 1", scheduler.stopCalls)
	}
}

func TestPopularEventsServiceDoesNotStartSchedulerWhenScheduleFails(t *testing.T) {
	scheduleErr := errors.New("invalid schedule")
	store := &popularEventsStoreStub{}
	cache := &popularEventsCacheStub{}
	scheduler := &popularEventsSchedulerStub{scheduleErr: scheduleErr}
	service := newPopularEventsTestService(store, cache, nil, scheduler, time.Now())
	defer service.Stop()

	err := service.Start()

	if !errors.Is(err, scheduleErr) {
		t.Fatalf("Start error = %v, want schedule error", err)
	}
	if scheduler.startCalls != 0 {
		t.Fatalf("scheduler start calls = %d, want 0", scheduler.startCalls)
	}
	storeCalls, _, _, _ := store.snapshot()
	if storeCalls != 0 {
		t.Fatalf("ranking store calls = %d, want no warmup after scheduling failure", storeCalls)
	}
}

func newPopularEventsTestService(
	store servicesevents.PopularEventsStore,
	cache servicesevents.PopularEventsCache,
	notifier servicesevents.PopularEventsNotifier,
	scheduler servicesevents.PopularEventsScheduler,
	now time.Time,
) servicesevents.PopularEventsService {
	return servicesevents.NewPopularEventsService(
		&testLogger{},
		store,
		cache,
		notifier,
		scheduler,
		func() time.Time { return now },
	)
}

func clonePopularEventsTestSnapshot(snapshot servicesevents.PopularEventsSnapshot) servicesevents.PopularEventsSnapshot {
	result := snapshot
	result.Events = append([]servicesevents.PopularEventView(nil), snapshot.Events...)
	for index := range result.Events {
		result.Events[index].Genres = append([]string(nil), snapshot.Events[index].Genres...)
	}
	return result
}

func waitPopularEventsSignal(t *testing.T, signal <-chan struct{}, operation string) {
	t.Helper()

	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", operation)
	}
}

func waitPopularEventsCacheGets(t *testing.T, cache *popularEventsCacheStub, minimum int) {
	t.Helper()

	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	retry := time.NewTicker(5 * time.Millisecond)
	defer retry.Stop()

	for {
		getCalls, _, _, _ := cache.callsSnapshot()
		if getCalls >= minimum {
			return
		}

		select {
		case <-deadline.C:
			t.Fatalf("timed out waiting for %d cache reads; got %d", minimum, getCalls)
		case <-retry.C:
		}
	}
}

func triggerPopularEventsJobUntilSignal(t *testing.T, job func(), signal <-chan struct{}, operation string) {
	t.Helper()

	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	retry := time.NewTicker(5 * time.Millisecond)
	defer retry.Stop()

	for {
		job()
		select {
		case <-signal:
			return
		case <-retry.C:
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s", operation)
		}
	}
}
