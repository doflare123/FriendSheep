package services

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidRefreshToken  = errors.New("невалидный refresh токен")
	ErrRefreshTokenReplay   = errors.New("обнаружен повторный refresh токен")
	ErrAuthSessionNotFound  = errors.New("auth сессия не найдена")
	ErrAuthSessionRevoked   = errors.New("auth сессия отозвана")
	errAuthSessionStoreNil  = errors.New("auth session store is not configured")
	errAuthSessionConfigNil = errors.New("auth session store config is invalid")
)

type AuthSession struct {
	SessionID        string
	UserID           uint
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type CreateAuthSessionInput struct {
	UserID uint
}

type RotateAuthSessionInput struct {
	RefreshToken string
}

type AuthSessionStore interface {
	CreateSession(ctx context.Context, input CreateAuthSessionInput) (AuthSession, error)
	RotateSession(ctx context.Context, input RotateAuthSessionInput) (AuthSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID uint) error
	HasActiveSession(ctx context.Context, sessionID string) (bool, error)
}

type AuthSessionStoreConfig struct {
	RefreshTTL         time.Duration
	SessionKeyPrefix   string
	UserSessionsPrefix string
}

func (c AuthSessionStoreConfig) normalized() (AuthSessionStoreConfig, error) {
	if c.RefreshTTL <= 0 {
		return AuthSessionStoreConfig{}, fmt.Errorf("%w: refresh ttl must be positive", errAuthSessionConfigNil)
	}
	if c.SessionKeyPrefix == "" {
		c.SessionKeyPrefix = "auth:sessions:"
	}
	if c.UserSessionsPrefix == "" {
		c.UserSessionsPrefix = "auth:user_sessions:"
	}
	return c, nil
}
