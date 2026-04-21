package tests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"friendship/config"
	"friendship/models"
	statsusers "friendship/models/stats_users"
	"friendship/services/register"
	session "friendship/sessions"
	"friendship/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeSessionStore struct {
	sessions map[string]*session.Session
	deleted  map[string]bool
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{
		sessions: make(map[string]*session.Session),
		deleted:  make(map[string]bool),
	}
}

func (s *fakeSessionStore) CreateSession(ctx context.Context, sessionID, code string, sessionType models.SessionTypeReg, expiration time.Duration, extra map[string]string) error {
	s.sessions[sessionID] = &session.Session{
		Code:  code,
		Type:  sessionType,
		Extra: extra,
	}
	return nil
}

func (s *fakeSessionStore) GetSessionFields(ctx context.Context, sessionID string, fields ...string) (map[string]string, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, field := range fields {
		switch field {
		case "code":
			result[field] = sess.Code
		case "is_verified":
			if sess.IsVerified {
				result[field] = "1"
			} else {
				result[field] = "0"
			}
		case "type":
			result[field] = string(sess.Type)
		case "attempts":
			result[field] = strconv.Itoa(sess.Attempts)
		default:
			result[field] = sess.Extra[field]
		}
	}
	return result, nil
}

func (s *fakeSessionStore) GetSession(ctx context.Context, sessionID string) (*session.Session, error) {
	if s.deleted[sessionID] {
		return nil, session.ErrSessionNotFound
	}
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, session.ErrSessionNotFound
	}
	return sess, nil
}

func (s *fakeSessionStore) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	switch field {
	case "is_verified":
		sess.IsVerified = value == "1" || value == 1 || value == true
	case "attempts":
		switch v := value.(type) {
		case string:
			attempts, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			sess.Attempts = attempts
		case int:
			sess.Attempts = v
		}
	default:
		if sess.Extra == nil {
			sess.Extra = make(map[string]string)
		}
		sess.Extra[field] = fmt.Sprint(value)
	}
	return nil
}

func (s *fakeSessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	s.deleted[sessionID] = true
	delete(s.sessions, sessionID)
	return nil
}

func (s *fakeSessionStore) IncrementAttempts(ctx context.Context, sessionID string) (int, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	sess.Attempts++
	return sess.Attempts, nil
}

func (s *fakeSessionStore) MarkAsVerified(ctx context.Context, sessionID string) error {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	sess.IsVerified = true
	return nil
}

func TestRegisterServiceVerifySessionSuccess(t *testing.T) {
	store := newFakeSessionStore()
	store.sessions["session-1"] = &session.Session{
		Code: "123456",
		Type: models.SessionTypeRegister,
	}

	service := newRegisterServiceForTest(t, store, nil)
	verified, err := service.VerifySession(context.Background(), register.VerifySessionInput{
		SessionID: "session-1",
		Type:      string(models.SessionTypeRegister),
		Code:      "123456",
	})

	if err != nil {
		t.Fatalf("VerifySession returned error: %v", err)
	}
	if !verified {
		t.Fatal("VerifySession returned false")
	}
	if !store.sessions["session-1"].IsVerified {
		t.Fatal("session was not marked as verified")
	}
}

func TestRegisterServiceVerifySessionTooManyAttemptsDeletesSession(t *testing.T) {
	store := newFakeSessionStore()
	store.sessions["session-1"] = &session.Session{
		Code:     "123456",
		Type:     models.SessionTypeRegister,
		Attempts: 2,
	}

	service := newRegisterServiceForTest(t, store, nil)
	verified, err := service.VerifySession(context.Background(), register.VerifySessionInput{
		SessionID: "session-1",
		Type:      string(models.SessionTypeRegister),
		Code:      "000000",
	})

	if verified {
		t.Fatal("VerifySession returned true for invalid code")
	}
	if !errors.Is(err, register.ErrTooManyAttempts) {
		t.Fatalf("err = %v, want ErrTooManyAttempts", err)
	}
	if !store.deleted["session-1"] {
		t.Fatal("session was not deleted after too many attempts")
	}
}

func TestRegisterServiceCreateUserCreatesDefaultsAndDeletesSession(t *testing.T) {
	db := newRegisterDB(t)
	repo := &testPostgresRepository{db: db}
	store := newFakeSessionStore()
	store.sessions["verified-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeRegister,
	}

	service := newRegisterServiceForTest(t, store, repo)
	authResponse, err := service.CreateUser(context.Background(), register.CreateUserInput{
		Name:      "Valid User",
		Password:  "Password123!",
		Email:     "user@example.com",
		SessionID: "verified-session",
	})

	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if authResponse.AccessToken == "" || authResponse.RefreshToken == "" {
		t.Fatal("CreateUser returned empty token pair")
	}
	if !store.deleted["verified-session"] {
		t.Fatal("verified session was not deleted")
	}

	var usersCount int64
	if err := db.Model(&models.User{}).Where("email = ?", "user@example.com").Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 1 {
		t.Fatalf("users count = %d, want 1", usersCount)
	}

	assertCount(t, db, &statsusers.SettingTile{}, 1)
	assertCount(t, db, &statsusers.SessionStats_users{}, 1)
	assertCount(t, db, &statsusers.SideStats_users{}, 1)
}

func newRegisterServiceForTest(t *testing.T, store *fakeSessionStore, repo *testPostgresRepository) register.RegService {
	t.Helper()

	if repo == nil {
		repo = &testPostgresRepository{db: newRegisterDB(t)}
	}

	service, err := register.NewRegisterSrv(
		&testLogger{},
		store,
		repo,
		&config.Config{},
		utils.NewJWTUtils("test-secret"),
	)
	if err != nil {
		t.Fatalf("NewRegisterSrv returned error: %v", err)
	}
	return service
}

func newRegisterDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.DaysWeek{},
		&statsusers.SettingTile{},
		&statsusers.SessionStats_users{},
		&statsusers.SideStats_users{},
	); err != nil {
		t.Fatalf("auto migrate register models: %v", err)
	}

	return db
}

func assertCount(t *testing.T, db *gorm.DB, model interface{}, want int64) {
	t.Helper()

	var got int64
	if err := db.Model(model).Count(&got).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if got != want {
		t.Fatalf("count %T = %d, want %d", model, got, want)
	}
}
