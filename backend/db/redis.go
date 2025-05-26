package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type SessionStore struct {
	redisClient *redis.Client
}

func NewSessionStore(addr string) *SessionStore {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &SessionStore{redisClient: rdb}
}
