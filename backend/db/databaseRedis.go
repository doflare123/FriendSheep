package db

import (
	"context"
	"fmt"
	"friendship/models"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type SessionStore struct {
	redisClient *redis.Client
}

var GlobalSessionStore *SessionStore

func NewSessionStore(addr string) *SessionStore {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &SessionStore{redisClient: rdb}
}

func InitRedis() error {
	redisAddr := os.Getenv("REDIS_URI")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	sessionStore := NewSessionStore(redisAddr)

	ctx := context.Background()
	_, err := sessionStore.redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("не удалось подключиться к Redis: %v", err)
	}

	SetGlobalSessionStore(sessionStore)
	return nil
}

func (s *SessionStore) CreateSession(sessionID, code string, sessionType models.SessionTypeReg, expiration time.Duration, extra ...map[string]string) error {
	fields := map[string]interface{}{
		"code":        code,
		"is_verified": "0",
		"type":        string(sessionType),
		"attempts":    "0",
	}

	// сохраняем дополнительные поля (например email)
	if len(extra) > 0 {
		for k, v := range extra[0] {
			fields[k] = v
		}
	}

	if err := s.redisClient.HSet(ctx, sessionID, fields).Err(); err != nil {
		return err
	}

	return s.redisClient.Expire(ctx, sessionID, expiration).Err()
}

func (s *SessionStore) GetSessionFields(sessionID string, fields ...string) (map[string]string, error) {
	values, err := s.redisClient.HMGet(ctx, sessionID, fields...).Result()
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

func (s *SessionStore) UpdateSessionField(sessionID, field string, value interface{}) error {
	return s.redisClient.HSet(ctx, sessionID, field, value).Err()
}

func (s *SessionStore) DeleteSession(sessionID string) error {
	return s.redisClient.Del(ctx, sessionID).Err()
}

func (s *SessionStore) GetRedisClient() *redis.Client {
	return s.redisClient
}

func SetGlobalSessionStore(store *SessionStore) {
	GlobalSessionStore = store
}

func GetRedis() *redis.Client {
	if GlobalSessionStore == nil {
		return nil
	}
	return GlobalSessionStore.GetRedisClient()
}
