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
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(logger logger.Logger, conf config.Config) RedisRepository {
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

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisRepository) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
