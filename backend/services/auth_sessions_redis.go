package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"friendship/repository"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	authSessionUserIDField      = "user_id"
	authSessionRefreshHashField = "refresh_hash"
	authSessionRevokedField     = "revoked"
	authSessionCreatedAtField   = "created_at"
	authSessionRotatedAtField   = "rotated_at"
	authSessionRevokedAtField   = "revoked_at"
)

var rotateAuthSessionScript = redis.NewScript(`
local values = redis.call("HMGET", KEYS[1], "user_id", "refresh_hash", "revoked")
if not values[1] then
	return {"missing"}
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl <= 0 then
	redis.call("DEL", KEYS[1])
	return {"missing"}
end
if values[3] == "1" then
	return {"revoked", values[1], tostring(ttl)}
end
if values[2] ~= ARGV[1] then
	redis.call("HSET", KEYS[1], "revoked", "1", "revoked_at", ARGV[3])
	return {"replay", values[1], tostring(ttl)}
end
redis.call("HSET", KEYS[1], "refresh_hash", ARGV[2], "rotated_at", ARGV[3], "revoked", "0")
return {"ok", values[1], tostring(ttl)}
`)

var revokeAllAuthSessionsScript = redis.NewScript(`
local session_ids = redis.call("SMEMBERS", KEYS[1])
for _, session_id in ipairs(session_ids) do
	redis.call("DEL", ARGV[1] .. session_id)
end
redis.call("DEL", KEYS[1])
return #session_ids
`)

type redisAuthSessionStore struct {
	client             *redis.Client
	refreshTTL         time.Duration
	sessionKeyPrefix   string
	userSessionsPrefix string
}

func NewRedisAuthSessionStore(redisRepo repository.RedisRepository, cfg AuthSessionStoreConfig) (AuthSessionStore, error) {
	if redisRepo == nil {
		return nil, errAuthSessionStoreNil
	}

	normalized, err := cfg.normalized()
	if err != nil {
		return nil, err
	}
	if redisRepo.Client() == nil {
		return nil, errAuthSessionStoreNil
	}

	return &redisAuthSessionStore{
		client:             redisRepo.Client(),
		refreshTTL:         normalized.RefreshTTL,
		sessionKeyPrefix:   normalized.SessionKeyPrefix,
		userSessionsPrefix: normalized.UserSessionsPrefix,
	}, nil
}

func (s *redisAuthSessionStore) CreateSession(ctx context.Context, input CreateAuthSessionInput) (AuthSession, error) {
	if s == nil || s.client == nil {
		return AuthSession{}, errAuthSessionStoreNil
	}
	if input.UserID == 0 {
		return AuthSession{}, fmt.Errorf("%w: user id must be positive", ErrInvalidRefreshToken)
	}

	sessionID, err := randomSessionToken(18)
	if err != nil {
		return AuthSession{}, fmt.Errorf("ошибка генерации идентификатора auth сессии: %w", err)
	}

	refreshToken, refreshHash, err := newOpaqueRefreshToken(sessionID)
	if err != nil {
		return AuthSession{}, fmt.Errorf("ошибка генерации refresh токена: %w", err)
	}

	now := time.Now().UTC()
	key := s.sessionKey(sessionID)
	userSessionsKey := s.userSessionsKey(input.UserID)

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		authSessionUserIDField:      strconv.FormatUint(uint64(input.UserID), 10),
		authSessionRefreshHashField: refreshHash,
		authSessionRevokedField:     "0",
		authSessionCreatedAtField:   strconv.FormatInt(now.Unix(), 10),
	})
	pipe.PExpire(ctx, key, s.refreshTTL)
	pipe.SAdd(ctx, userSessionsKey, sessionID)
	pipe.PExpire(ctx, userSessionsKey, s.refreshTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return AuthSession{}, fmt.Errorf("ошибка сохранения auth сессии: %w", err)
	}

	return AuthSession{
		SessionID:        sessionID,
		UserID:           input.UserID,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: now.Add(s.refreshTTL),
	}, nil
}

func (s *redisAuthSessionStore) RotateSession(ctx context.Context, input RotateAuthSessionInput) (AuthSession, error) {
	if s == nil || s.client == nil {
		return AuthSession{}, errAuthSessionStoreNil
	}

	sessionID, secret, err := parseOpaqueRefreshToken(input.RefreshToken)
	if err != nil {
		return AuthSession{}, ErrInvalidRefreshToken
	}

	newRefreshToken, newRefreshHash, err := newOpaqueRefreshToken(sessionID)
	if err != nil {
		return AuthSession{}, fmt.Errorf("ошибка генерации refresh токена: %w", err)
	}

	expectedHash := hashRefreshSecret(secret)
	now := time.Now().UTC()
	result, err := rotateAuthSessionScript.Run(
		ctx,
		s.client,
		[]string{s.sessionKey(sessionID)},
		expectedHash,
		newRefreshHash,
		strconv.FormatInt(now.Unix(), 10),
	).Result()
	if err != nil {
		return AuthSession{}, fmt.Errorf("ошибка ротации auth сессии: %w", err)
	}

	values, err := toStringSlice(result)
	if err != nil || len(values) == 0 {
		return AuthSession{}, fmt.Errorf("ошибка ротации auth сессии: неожиданный ответ redis")
	}

	status := values[0]
	switch status {
	case "missing":
		return AuthSession{}, ErrAuthSessionNotFound
	case "revoked":
		return AuthSession{}, ErrAuthSessionRevoked
	case "replay":
		return AuthSession{}, ErrRefreshTokenReplay
	case "ok":
	default:
		return AuthSession{}, fmt.Errorf("ошибка ротации auth сессии: неизвестный статус %q", status)
	}

	if len(values) < 2 {
		return AuthSession{}, fmt.Errorf("ошибка ротации auth сессии: отсутствует user_id")
	}

	userIDValue, err := strconv.ParseUint(values[1], 10, 64)
	if err != nil || userIDValue == 0 {
		return AuthSession{}, fmt.Errorf("ошибка ротации auth сессии: некорректный user_id")
	}

	refreshExpiresAt := now
	if len(values) >= 3 {
		ttlMilliseconds, ttlErr := strconv.ParseInt(values[2], 10, 64)
		if ttlErr == nil && ttlMilliseconds > 0 {
			refreshExpiresAt = now.Add(time.Duration(ttlMilliseconds) * time.Millisecond)
		}
	}

	return AuthSession{
		SessionID:        sessionID,
		UserID:           uint(userIDValue),
		RefreshToken:     newRefreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *redisAuthSessionStore) RevokeSession(ctx context.Context, sessionID string) error {
	if s == nil || s.client == nil {
		return errAuthSessionStoreNil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ErrAuthSessionNotFound
	}

	userID, err := s.lookupSessionUserID(ctx, sessionID)
	if err != nil && !errorsIsSessionMissing(err) {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.sessionKey(sessionID))
	if userID > 0 {
		pipe.SRem(ctx, s.userSessionsKey(userID), sessionID)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("ошибка отзыва auth сессии: %w", err)
	}

	return nil
}

func (s *redisAuthSessionStore) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	if s == nil || s.client == nil {
		return errAuthSessionStoreNil
	}
	if userID == 0 {
		return ErrAuthSessionNotFound
	}

	if _, err := revokeAllAuthSessionsScript.Run(
		ctx,
		s.client,
		[]string{s.userSessionsKey(userID)},
		s.sessionKeyPrefix,
	).Result(); err != nil {
		return fmt.Errorf("ошибка отзыва auth сессий пользователя: %w", err)
	}

	return nil
}

func (s *redisAuthSessionStore) HasActiveSession(ctx context.Context, sessionID string) (bool, error) {
	if s == nil || s.client == nil {
		return false, errAuthSessionStoreNil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false, nil
	}

	values, err := s.client.HMGet(ctx, s.sessionKey(sessionID), authSessionUserIDField, authSessionRevokedField).Result()
	if err != nil {
		return false, fmt.Errorf("ошибка чтения auth сессии: %w", err)
	}
	if len(values) == 0 || values[0] == nil {
		return false, nil
	}
	if len(values) > 1 && values[1] != nil && fmt.Sprint(values[1]) == "1" {
		return false, nil
	}

	return true, nil
}

func (s *redisAuthSessionStore) lookupSessionUserID(ctx context.Context, sessionID string) (uint, error) {
	value, err := s.client.HGet(ctx, s.sessionKey(sessionID), authSessionUserIDField).Result()
	if err == redis.Nil {
		return 0, ErrAuthSessionNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения user_id auth сессии: %w", err)
	}

	userID, parseErr := strconv.ParseUint(value, 10, 64)
	if parseErr != nil || userID == 0 {
		return 0, fmt.Errorf("ошибка чтения user_id auth сессии: некорректное значение")
	}

	return uint(userID), nil
}

func (s *redisAuthSessionStore) sessionKey(sessionID string) string {
	return s.sessionKeyPrefix + sessionID
}

func (s *redisAuthSessionStore) userSessionsKey(userID uint) string {
	return s.userSessionsPrefix + strconv.FormatUint(uint64(userID), 10)
}

func newOpaqueRefreshToken(sessionID string) (string, string, error) {
	secret, err := randomSessionToken(32)
	if err != nil {
		return "", "", err
	}
	return sessionID + "." + secret, hashRefreshSecret(secret), nil
}

func parseOpaqueRefreshToken(token string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid refresh token format")
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid refresh token format")
	}
	sessionBytes, sessionErr := base64.RawURLEncoding.DecodeString(parts[0])
	secretBytes, secretErr := base64.RawURLEncoding.DecodeString(parts[1])
	if sessionErr != nil || secretErr != nil || len(sessionBytes) != 18 || len(secretBytes) != 32 {
		return "", "", fmt.Errorf("invalid refresh token encoding")
	}
	return parts[0], parts[1], nil
}

func hashRefreshSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func randomSessionToken(bytesCount int) (string, error) {
	if bytesCount <= 0 {
		return "", fmt.Errorf("bytesCount must be positive")
	}

	buffer := make([]byte, bytesCount)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func toStringSlice(value interface{}) ([]string, error) {
	items, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected redis response type %T", value)
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			result = append(result, "")
			continue
		}
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected redis response item type %T", item)
		}
		result = append(result, text)
	}

	return result, nil
}

func errorsIsSessionMissing(err error) bool {
	return err == nil || err == ErrAuthSessionNotFound
}
