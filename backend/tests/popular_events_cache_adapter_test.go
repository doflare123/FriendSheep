package tests

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"friendship/repository"
	servicesevents "friendship/services/events"

	"github.com/redis/go-redis/v9"
)

type popularEventsRedisRepositoryStub struct {
	getPayload string
	getErr     error
	getCtx     context.Context
	getKey     string

	setErr        error
	setCtx        context.Context
	setKey        string
	setValue      interface{}
	setExpiration time.Duration
}

func (r *popularEventsRedisRepositoryStub) Client() *redis.Client {
	return nil
}

func (r *popularEventsRedisRepositoryStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	r.setCtx = ctx
	r.setKey = key
	r.setValue = value
	r.setExpiration = expiration
	return r.setErr
}

func (r *popularEventsRedisRepositoryStub) Get(ctx context.Context, key string) (string, error) {
	r.getCtx = ctx
	r.getKey = key
	return r.getPayload, r.getErr
}

func (r *popularEventsRedisRepositoryStub) Del(context.Context, string) error {
	return nil
}

func (r *popularEventsRedisRepositoryStub) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func (r *popularEventsRedisRepositoryStub) Expire(context.Context, string, time.Duration) error {
	return nil
}

func (r *popularEventsRedisRepositoryStub) HSet(context.Context, string, map[string]interface{}) error {
	return nil
}

func (r *popularEventsRedisRepositoryStub) HGet(context.Context, string, string) (string, error) {
	return "", nil
}

func (r *popularEventsRedisRepositoryStub) HMGet(context.Context, string, ...string) (map[string]string, error) {
	return nil, nil
}

func (r *popularEventsRedisRepositoryStub) HGetAll(context.Context, string) (map[string]string, error) {
	return nil, nil
}

func TestRedisPopularEventsCacheSerializesSnapshotAndPropagatesContextKeyTTL(t *testing.T) {
	ctx := context.WithValue(context.Background(), popularEventsServiceContextKey{}, "redis-set")
	snapshot := servicesevents.PopularEventsSnapshot{
		Events: []servicesevents.PopularEventView{{
			ID:     71,
			Title:  "Cached adapter event",
			Genres: []string{"Co-op"},
		}},
		UpdatedAt: time.Date(2037, 9, 10, 11, 12, 13, 0, time.UTC),
		Count:     1,
	}
	repo := &popularEventsRedisRepositoryStub{}
	cache := servicesevents.NewRedisPopularEventsCache(repo)
	expiration := 7 * time.Hour

	err := cache.Set(ctx, "popular:test", snapshot, expiration)

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if repo.setCtx != ctx || repo.setKey != "popular:test" || repo.setExpiration != expiration {
		t.Fatalf("repository set call = ctx:%v key:%q ttl:%v", repo.setCtx, repo.setKey, repo.setExpiration)
	}

	var payload []byte
	switch value := repo.setValue.(type) {
	case []byte:
		payload = value
	case string:
		payload = []byte(value)
	default:
		t.Fatalf("Redis adapter leaked unencoded value %T to repository", repo.setValue)
	}
	var stored servicesevents.PopularEventsSnapshot
	if err := json.Unmarshal(payload, &stored); err != nil {
		t.Fatalf("decode stored snapshot: %v", err)
	}
	if !reflect.DeepEqual(stored, snapshot) {
		t.Fatalf("stored snapshot = %#v, want %#v", stored, snapshot)
	}
}

func TestRedisPopularEventsCacheDeserializesSnapshotAndMapsRedisMiss(t *testing.T) {
	want := servicesevents.PopularEventsSnapshot{
		Events:    []servicesevents.PopularEventView{{ID: 72, Title: "Redis hit"}},
		UpdatedAt: time.Date(2038, 10, 11, 12, 13, 14, 0, time.UTC),
		Count:     1,
	}
	payload, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	ctx := context.WithValue(context.Background(), popularEventsServiceContextKey{}, "redis-get")
	repo := &popularEventsRedisRepositoryStub{getPayload: string(payload)}
	cache := servicesevents.NewRedisPopularEventsCache(repo)

	got, err := cache.Get(ctx, "popular:test")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("snapshot = %#v, want %#v", got, want)
	}
	if repo.getCtx != ctx || repo.getKey != "popular:test" {
		t.Fatalf("repository get call = ctx:%v key:%q", repo.getCtx, repo.getKey)
	}

	repo.getErr = redis.Nil
	_, err = cache.Get(ctx, "popular:missing")
	if !errors.Is(err, servicesevents.ErrPopularEventsCacheMiss) {
		t.Fatalf("miss error = %v, want ErrPopularEventsCacheMiss", err)
	}
}

func TestRedisPopularEventsCachePropagatesBackendAndDecodeErrors(t *testing.T) {
	backendErr := errors.New("redis unavailable")
	repo := &popularEventsRedisRepositoryStub{getErr: backendErr}
	cache := servicesevents.NewRedisPopularEventsCache(repo)

	if _, err := cache.Get(context.Background(), "popular:test"); !errors.Is(err, backendErr) {
		t.Fatalf("backend error = %v, want %v", err, backendErr)
	}

	repo.getErr = nil
	repo.getPayload = "{not-json"
	if _, err := cache.Get(context.Background(), "popular:test"); err == nil ||
		errors.Is(err, servicesevents.ErrPopularEventsCacheMiss) {
		t.Fatalf("decode error = %v, want non-miss error", err)
	}

	repo.setErr = backendErr
	if err := cache.Set(
		context.Background(),
		"popular:test",
		servicesevents.PopularEventsSnapshot{},
		time.Hour,
	); !errors.Is(err, backendErr) {
		t.Fatalf("set error = %v, want %v", err, backendErr)
	}
}

var _ repository.RedisRepository = (*popularEventsRedisRepositoryStub)(nil)
