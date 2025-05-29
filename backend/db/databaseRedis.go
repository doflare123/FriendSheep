package db

import (
	"context"
	"friendship/models"
	"time"

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

func (s *SessionStore) CreateSession(sessionID, code string, sessionType models.SessionTypeReg, expiration time.Duration) error {
	fields := map[string]interface{}{
		"code":        code,
		"is_verified": "0",
		"type":        string(sessionType), // <- обязательно привести к string
	}
	err := s.redisClient.HSet(ctx, sessionID, fields).Err()
	if err != nil {
		return err
	}

	// Можно задать TTL сессии, если нужно:
	return s.redisClient.Expire(ctx, sessionID, expiration).Err()
}

func (s *SessionStore) GetSessionFields(sessionID string, fields ...string) (map[string]string, error) {
	values, err := s.redisClient.HMGet(ctx, sessionID, fields...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for i, field := range fields {
		// Redis может вернуть nil если поля нет, проверим это
		if values[i] != nil {
			result[field] = values[i].(string)
		}
	}
	return result, nil
}

func (s *SessionStore) UpdateSessionField(sessionID, field string, value interface{}) error {
	return s.redisClient.HSet(ctx, sessionID, field, value).Err()
}
