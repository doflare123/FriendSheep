package services

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestOpaqueRefreshTokenRoundTrip(t *testing.T) {
	sessionID, err := randomSessionToken(18)
	if err != nil {
		t.Fatalf("randomSessionToken: %v", err)
	}

	token, tokenHash, err := newOpaqueRefreshToken(sessionID)
	if err != nil {
		t.Fatalf("newOpaqueRefreshToken: %v", err)
	}

	parsedSessionID, secret, err := parseOpaqueRefreshToken(token)
	if err != nil {
		t.Fatalf("parseOpaqueRefreshToken: %v", err)
	}
	if parsedSessionID != sessionID {
		t.Fatalf("session ID = %q, want %q", parsedSessionID, sessionID)
	}
	if tokenHash != hashRefreshSecret(secret) {
		t.Fatal("stored hash does not match the opaque refresh secret")
	}
	if strings.Contains(tokenHash, secret) {
		t.Fatal("refresh secret is present in its stored representation")
	}
}

func TestParseOpaqueRefreshTokenRejectsMalformedValues(t *testing.T) {
	tests := []string{
		"",
		"missing-dot",
		"too.many.parts",
		".secret",
		"session.",
		"invalid-base64.invalid-base64",
	}

	for _, token := range tests {
		t.Run(token, func(t *testing.T) {
			if _, _, err := parseOpaqueRefreshToken(token); err == nil {
				t.Fatalf("parseOpaqueRefreshToken(%q) returned nil error", token)
			}
		})
	}
}

func TestRedisAuthSessionStoreRotationAndRevocation(t *testing.T) {
	redisURL := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_REDIS_URL"))
	if redisURL == "" {
		t.Skip("FRIENDSHEEP_TEST_REDIS_URL is not set")
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse FRIENDSHEEP_TEST_REDIS_URL: %v", err)
	}
	client := redis.NewClient(options)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping test Redis: %v", err)
	}

	prefix := "test:friendship:auth:" + time.Now().UTC().Format("20060102150405.000000000") + ":"
	store := &redisAuthSessionStore{
		client:             client,
		refreshTTL:         5 * time.Minute,
		sessionKeyPrefix:   prefix + "session:",
		userSessionsPrefix: prefix + "user:",
	}

	var createdSessionIDs []string
	t.Cleanup(func() {
		keys := make([]string, 0, len(createdSessionIDs)+2)
		for _, sessionID := range createdSessionIDs {
			keys = append(keys, store.sessionKey(sessionID))
		}
		keys = append(keys, store.userSessionsKey(77), store.userSessionsKey(88))
		_ = client.Del(context.Background(), keys...).Err()
	})

	created, err := store.CreateSession(ctx, CreateAuthSessionInput{UserID: 77})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	createdSessionIDs = append(createdSessionIDs, created.SessionID)

	stored, err := client.HGetAll(ctx, store.sessionKey(created.SessionID)).Result()
	if err != nil {
		t.Fatalf("read stored session: %v", err)
	}
	_, refreshSecret, err := parseOpaqueRefreshToken(created.RefreshToken)
	if err != nil {
		t.Fatalf("parse created refresh token: %v", err)
	}
	if stored[authSessionRefreshHashField] != hashRefreshSecret(refreshSecret) {
		t.Fatal("Redis does not contain the expected refresh-token hash")
	}
	for _, value := range stored {
		if strings.Contains(value, refreshSecret) || strings.Contains(value, created.RefreshToken) {
			t.Fatal("Redis contains raw refresh-token material")
		}
	}

	type rotateResult struct {
		session AuthSession
		err     error
	}
	results := make(chan rotateResult, 2)
	var start sync.WaitGroup
	start.Add(1)
	for range 2 {
		go func() {
			start.Wait()
			session, rotateErr := store.RotateSession(ctx, RotateAuthSessionInput{RefreshToken: created.RefreshToken})
			results <- rotateResult{session: session, err: rotateErr}
		}()
	}
	start.Done()

	var successfulRotation AuthSession
	var successCount int
	var replayCount int
	for range 2 {
		result := <-results
		switch {
		case result.err == nil:
			successfulRotation = result.session
			successCount++
		case errors.Is(result.err, ErrRefreshTokenReplay):
			replayCount++
		default:
			t.Fatalf("concurrent RotateSession returned unexpected error: %v", result.err)
		}
	}
	if successCount != 1 || replayCount != 1 {
		t.Fatalf("rotation results: successes=%d replays=%d, want 1 and 1", successCount, replayCount)
	}

	active, err := store.HasActiveSession(ctx, created.SessionID)
	if err != nil {
		t.Fatalf("HasActiveSession after replay: %v", err)
	}
	if active {
		t.Fatal("session remained active after refresh-token replay")
	}
	if _, err := store.RotateSession(ctx, RotateAuthSessionInput{RefreshToken: successfulRotation.RefreshToken}); !errors.Is(err, ErrAuthSessionRevoked) {
		t.Fatalf("RotateSession with the winning token error = %v, want ErrAuthSessionRevoked", err)
	}

	first, err := store.CreateSession(ctx, CreateAuthSessionInput{UserID: 88})
	if err != nil {
		t.Fatalf("create first user session: %v", err)
	}
	second, err := store.CreateSession(ctx, CreateAuthSessionInput{UserID: 88})
	if err != nil {
		t.Fatalf("create second user session: %v", err)
	}
	createdSessionIDs = append(createdSessionIDs, first.SessionID, second.SessionID)

	if err := store.RevokeAllUserSessions(ctx, 88); err != nil {
		t.Fatalf("RevokeAllUserSessions: %v", err)
	}
	for _, sessionID := range []string{first.SessionID, second.SessionID} {
		active, err := store.HasActiveSession(ctx, sessionID)
		if err != nil {
			t.Fatalf("HasActiveSession(%q): %v", sessionID, err)
		}
		if active {
			t.Fatalf("session %q remained active after revoke-all", sessionID)
		}
	}
}
