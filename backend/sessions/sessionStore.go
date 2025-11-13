package session

import (
	"context"
	"errors"
	"fmt"
	"friendship/models"
	"friendship/repository"
	"strconv"
	"time"
)

var (
	ErrSessionNotFound = errors.New("сессия не найдена")
	ErrInvalidSession  = errors.New("проблема с данными сессии")
)

type SessionStore interface {
	CreateSession(ctx context.Context, sessionID, code string, sessionType models.SessionTypeReg, expiration time.Duration, extra map[string]string) error
	GetSessionFields(ctx context.Context, sessionID string, fields ...string) (map[string]string, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
	IncrementAttempts(ctx context.Context, sessionID string) (int, error)
	MarkAsVerified(ctx context.Context, sessionID string) error
}

type Session struct {
	Code       string
	IsVerified bool
	Type       models.SessionTypeReg
	Attempts   int
	Extra      map[string]string
}

type sessionStore struct {
	redis repository.RedisRepository
}

func NewSessionStore(redis repository.RedisRepository) SessionStore {
	return &sessionStore{
		redis: redis,
	}
}

func (s *sessionStore) CreateSession(
	ctx context.Context,
	sessionID, code string,
	sessionType models.SessionTypeReg,
	expiration time.Duration,
	extra map[string]string,
) error {
	fields := map[string]interface{}{
		"code":        code,
		"is_verified": "0",
		"type":        string(sessionType),
		"attempts":    "0",
	}

	for k, v := range extra {
		fields[k] = v
	}

	if err := s.redis.HSet(ctx, sessionID, fields); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.redis.Expire(ctx, sessionID, expiration); err != nil {
		return fmt.Errorf("failed to set session expiration: %w", err)
	}

	return nil
}

func (s *sessionStore) GetSessionFields(ctx context.Context, sessionID string, fields ...string) (map[string]string, error) {
	result, err := s.redis.HMGet(ctx, sessionID, fields...)
	if err != nil {
		return nil, fmt.Errorf("failed to get session fields: %w", err)
	}

	if len(result) == 0 {
		return nil, ErrSessionNotFound
	}

	return result, nil
}

func (s *sessionStore) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	data, err := s.redis.HGetAll(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if len(data) == 0 {
		return nil, ErrSessionNotFound
	}

	session := &Session{
		Code:  data["code"],
		Type:  models.SessionTypeReg(data["type"]),
		Extra: make(map[string]string),
	}

	if isVerified, ok := data["is_verified"]; ok {
		session.IsVerified = isVerified == "1"
	}

	if attempts, ok := data["attempts"]; ok {
		if attemptsInt, err := strconv.Atoi(attempts); err == nil {
			session.Attempts = attemptsInt
		}
	}

	standardFields := map[string]bool{
		"code": true, "is_verified": true, "type": true, "attempts": true,
	}
	for k, v := range data {
		if !standardFields[k] {
			session.Extra[k] = v
		}
	}

	return session, nil
}

func (s *sessionStore) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	fields := map[string]interface{}{field: value}
	if err := s.redis.HSet(ctx, sessionID, fields); err != nil {
		return fmt.Errorf("failed to update session field: %w", err)
	}
	return nil
}

func (s *sessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.redis.Del(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *sessionStore) IncrementAttempts(ctx context.Context, sessionID string) (int, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	newAttempts := session.Attempts + 1
	if err := s.UpdateSessionField(ctx, sessionID, "attempts", strconv.Itoa(newAttempts)); err != nil {
		return 0, err
	}

	return newAttempts, nil
}

func (s *sessionStore) MarkAsVerified(ctx context.Context, sessionID string) error {
	return s.UpdateSessionField(ctx, sessionID, "is_verified", "1")
}
