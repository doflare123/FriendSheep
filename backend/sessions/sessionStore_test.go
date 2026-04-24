package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"friendship/models"
)

type fakeSessionValueStore struct {
	hashes     map[string]map[string]string
	expiration map[string]time.Duration
	hsetErr    error
	expireErr  error
	hmgetErr   error
	hgetallErr error
	delErr     error
}

func newFakeSessionValueStore() *fakeSessionValueStore {
	return &fakeSessionValueStore{
		hashes:     make(map[string]map[string]string),
		expiration: make(map[string]time.Duration),
	}
}

func (f *fakeSessionValueStore) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	if f.hsetErr != nil {
		return f.hsetErr
	}
	if _, ok := f.hashes[key]; !ok {
		f.hashes[key] = make(map[string]string)
	}
	for field, value := range values {
		f.hashes[key][field] = toString(value)
	}
	return nil
}

func (f *fakeSessionValueStore) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if f.expireErr != nil {
		return f.expireErr
	}
	f.expiration[key] = expiration
	return nil
}

func (f *fakeSessionValueStore) HMGet(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	if f.hmgetErr != nil {
		return nil, f.hmgetErr
	}
	source := f.hashes[key]
	result := make(map[string]string)
	for _, field := range fields {
		if value, ok := source[field]; ok {
			result[field] = value
		}
	}
	return result, nil
}

func (f *fakeSessionValueStore) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if f.hgetallErr != nil {
		return nil, f.hgetallErr
	}
	source := f.hashes[key]
	result := make(map[string]string, len(source))
	for field, value := range source {
		result[field] = value
	}
	return result, nil
}

func (f *fakeSessionValueStore) Del(ctx context.Context, key string) error {
	if f.delErr != nil {
		return f.delErr
	}
	delete(f.hashes, key)
	return nil
}

func TestSessionStoreCreateAndReadSession(t *testing.T) {
	store := newFakeSessionValueStore()
	sessionStore := NewSessionStore(store)

	err := sessionStore.CreateSession(
		context.Background(),
		"session-1",
		"123456",
		models.SessionTypeRegister,
		10*time.Minute,
		map[string]string{"email": "user@example.com"},
	)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	session, err := sessionStore.GetSession(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}

	if session.Code != "123456" {
		t.Fatalf("code = %q, want 123456", session.Code)
	}
	if session.Type != models.SessionTypeRegister {
		t.Fatalf("type = %q, want %q", session.Type, models.SessionTypeRegister)
	}
	if session.IsVerified {
		t.Fatal("IsVerified = true, want false")
	}
	if session.Extra["email"] != "user@example.com" {
		t.Fatalf("email = %q, want user@example.com", session.Extra["email"])
	}
	if store.expiration["session-1"] != 10*time.Minute {
		t.Fatalf("expiration = %v, want 10m", store.expiration["session-1"])
	}
}

func TestSessionStoreIncrementAttemptsAndMarkVerified(t *testing.T) {
	store := newFakeSessionValueStore()
	store.hashes["session-1"] = map[string]string{
		"code":        "123456",
		"is_verified": "0",
		"type":        string(models.SessionTypeRegister),
		"attempts":    "1",
	}
	sessionStore := NewSessionStore(store)

	attempts, err := sessionStore.IncrementAttempts(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("IncrementAttempts returned error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}

	if err := sessionStore.MarkAsVerified(context.Background(), "session-1"); err != nil {
		t.Fatalf("MarkAsVerified returned error: %v", err)
	}

	session, err := sessionStore.GetSession(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if !session.IsVerified {
		t.Fatal("IsVerified = false, want true")
	}
}

func TestSessionStoreMissingSessionReturnsNotFound(t *testing.T) {
	sessionStore := NewSessionStore(newFakeSessionValueStore())

	_, err := sessionStore.GetSession(context.Background(), "missing")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}
