package repository

import (
	"context"
	"friendship/config"
	"friendship/logger"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	Client() *redis.Client

	// Базовые операции
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// Hash операции
	HSet(ctx context.Context, key string, values map[string]interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HMGet(ctx context.Context, key string, fields ...string) (map[string]string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(logger logger.Logger, conf *config.Config) RedisRepository {
	logger.Info("Try Redis connection")

	client, err := initRedis(conf.Redis.Addr, conf.Redis.Password, conf.Redis.DB, logger)
	if err != nil {
		logger.Error("Failure Redis connection", "error", err)
		os.Exit(1)
	}

	logger.Info("Success Redis connection")
	return &redisRepository{client: client}
}

func initRedis(addr, pass string, db int, logger logger.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Info("Redis connected to", "addr", addr)
	return client, nil
}

func (r *redisRepository) Client() *redis.Client {
	return r.client
}

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisRepository) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisRepository) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

func (r *redisRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *redisRepository) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	return r.client.HSet(ctx, key, values).Err()
}

func (r *redisRepository) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *redisRepository) HMGet(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	values, err := r.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for i, field := range fields {
		if values[i] != nil {
			result[field] = values[i].(string)
		}
	}
	return result, nil
}

func (r *redisRepository) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}
