package events

import (
	"context"
	"encoding/json"
	"errors"
	"friendship/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisPopularEventsCache struct {
	redisRepo repository.RedisRepository
}

func NewRedisPopularEventsCache(redisRepo repository.RedisRepository) PopularEventsCache {
	return &redisPopularEventsCache{redisRepo: redisRepo}
}

func (c *redisPopularEventsCache) Get(ctx context.Context, key string) (*PopularEventsSnapshot, error) {
	payload, err := c.redisRepo.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPopularEventsCacheMiss
		}
		return nil, err
	}

	var snapshot PopularEventsSnapshot
	if err := json.Unmarshal([]byte(payload), &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (c *redisPopularEventsCache) Set(ctx context.Context, key string, snapshot PopularEventsSnapshot, expiration time.Duration) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	return c.redisRepo.Set(ctx, key, payload, expiration)
}
