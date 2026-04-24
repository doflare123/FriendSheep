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

type fakeRegistrationEmailSender struct {
	verificationCalls []verificationEmailCall
	welcomeCalls      []welcomeEmailCall
	verificationErr   error
	welcomeErr        error
}

type verificationEmailCall struct {
	email      string
	code       string
	actionType string
}

type welcomeEmailCall struct {
	email string
	name  string
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{
		sessions: make(map[string]*session.Session),
		deleted:  make(map[string]bool),
	}
}

func (f *fakeRegistrationEmailSender) SendVerificationEmail(userEmail, code, actionType string) error {
	f.verificationCalls = append(f.verificationCalls, verificationEmailCall{
		email:      userEmail,
		code:       code,
		actionType: actionType,
	})
	return f.verificationErr
}

func (f *fakeRegistrationEmailSender) SendWelcomeEmail(userEmail, userName string) error {
	f.welcomeCalls = append(f.welcomeCalls, welcomeEmailCall{
		email: userEmail,
		name:  userName,
	})
	return f.welcomeErr
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

	service := newRegisterServiceForTest(t, store, nil, nil)
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

	service := newRegisterServiceForTest(t, store, nil, nil)
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

func TestRegisterServiceCreateSessionRejectsInvalidEmail(t *testing.T) {
	store := newFakeSessionStore()
	service := newRegisterServiceForTest(t, store, nil, nil)

	session, err := service.CreateSessionRegister(context.Background(), "roma.sakovich2gmail.com", string(models.SessionTypeRegister))

	if err == nil {
		t.Fatal("CreateSessionRegister returned nil error")
	}
	if session != nil {
		t.Fatalf("session = %#v, want nil", session)
	}
	if len(store.sessions) != 0 {
		t.Fatalf("sessions len = %d, want 0", len(store.sessions))
	}
}

func TestRegisterServiceCreateUserCreatesDefaultsAndDeletesSession(t *testing.T) {
	db := newRegisterDB(t)
	repo := &testPostgresRepository{db: db}
	store := newFakeSessionStore()
	notifier := &fakeRegistrationEmailSender{}
	store.sessions["verified-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeRegister,
		Extra: map[string]string{
			"email": "user@example.com",
		},
	}

	service := newRegisterServiceForTest(t, store, repo, notifier)
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
	if len(notifier.welcomeCalls) != 1 {
		t.Fatalf("welcome emails = %d, want 1", len(notifier.welcomeCalls))
	}
}

func TestRegisterServiceCreateUserRejectsResetPasswordSessionWithoutSideEffects(t *testing.T) {
	db := newRegisterDB(t)
	repo := &testPostgresRepository{db: db}
	store := newFakeSessionStore()
	notifier := &fakeRegistrationEmailSender{}
	store.sessions["verified-reset-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeResetPassword,
		Extra: map[string]string{
			"email": "user@example.com",
		},
	}

	service := newRegisterServiceForTest(t, store, repo, notifier)
	authResponse, err := service.CreateUser(context.Background(), register.CreateUserInput{
		Name:      "Valid User",
		Password:  "Password123!",
		Email:     "user@example.com",
		SessionID: "verified-reset-session",
	})

	if authResponse != nil {
		t.Fatalf("authResponse = %#v, want nil", authResponse)
	}
	if !errors.Is(err, register.ErrSessionTypeMismatch) {
		t.Fatalf("err = %v, want ErrSessionTypeMismatch", err)
	}
	if store.deleted["verified-reset-session"] {
		t.Fatal("session was deleted on rejected create user")
	}

	assertCount(t, db, &models.User{}, 0)
	assertCount(t, db, &statsusers.SettingTile{}, 0)
	assertCount(t, db, &statsusers.SessionStats_users{}, 0)
	assertCount(t, db, &statsusers.SideStats_users{}, 0)
	if len(notifier.welcomeCalls) != 0 {
		t.Fatalf("welcome emails = %d, want 0", len(notifier.welcomeCalls))
	}
}

func TestRegisterServiceCreateUserRejectsEmailMismatchWithoutSideEffects(t *testing.T) {
	db := newRegisterDB(t)
	repo := &testPostgresRepository{db: db}
	store := newFakeSessionStore()
	notifier := &fakeRegistrationEmailSender{}
	store.sessions["verified-register-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeRegister,
		Extra: map[string]string{
			"email": "bound@example.com",
		},
	}

	service := newRegisterServiceForTest(t, store, repo, notifier)
	authResponse, err := service.CreateUser(context.Background(), register.CreateUserInput{
		Name:      "Valid User",
		Password:  "Password123!",
		Email:     "other@example.com",
		SessionID: "verified-register-session",
	})

	// Assumption: active flow must bind CreateUser email to the verified session email
	// and reject mismatches before any durable side effects are committed.
	if authResponse != nil {
		t.Fatalf("authResponse = %#v, want nil", authResponse)
	}
	if !errors.Is(err, register.ErrSessionEmailMismatch) {
		t.Fatalf("err = %v, want ErrSessionEmailMismatch", err)
	}
	if store.deleted["verified-register-session"] {
		t.Fatal("session was deleted on rejected create user")
	}

	assertCount(t, db, &models.User{}, 0)
	assertCount(t, db, &statsusers.SettingTile{}, 0)
	assertCount(t, db, &statsusers.SessionStats_users{}, 0)
	assertCount(t, db, &statsusers.SideStats_users{}, 0)
	if len(notifier.welcomeCalls) != 0 {
		t.Fatalf("welcome emails = %d, want 0", len(notifier.welcomeCalls))
	}
}

func TestRegisterServiceChangePasswordRejectsWrongSessionTypeAndPreservesPassword(t *testing.T) {
	db := newRegisterDB(t)
	repo := &testPostgresRepository{db: db}
	store := newFakeSessionStore()
	store.sessions["verified-register-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeRegister,
		Extra: map[string]string{
			"email": "user@example.com",
		},
	}

	originalPassword := mustHashPassword(t, "OldPassword123!")
	user := models.User{
		Name:     "Existing User",
		Password: originalPassword,
		Us:       "existing-user",
		Email:    "user@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := newRegisterServiceForTest(t, store, repo, nil)
	err := service.ChangePassword(context.Background(), register.ChangePasswordInput{
		NewPassword: "NewPassword123!",
		Email:       "user@example.com",
		SessionID:   "verified-register-session",
	})

	if !errors.Is(err, register.ErrSessionTypeMismatch) {
		t.Fatalf("err = %v, want ErrSessionTypeMismatch", err)
	}
	if store.deleted["verified-register-session"] {
		t.Fatal("session was deleted on rejected password change")
	}

	assertUserPasswordHash(t, db, "user@example.com", originalPassword)
}

func TestRegisterServiceChangePasswordRejectsEmailMismatchAndPreservesPassword(t *testing.T) {
	db := newRegisterDB(t)
	repo := &testPostgresRepository{db: db}
	store := newFakeSessionStore()
	store.sessions["verified-reset-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeResetPassword,
		Extra: map[string]string{
			"email": "bound@example.com",
		},
	}

	originalPassword := mustHashPassword(t, "OldPassword123!")
	user := models.User{
		Name:     "Existing User",
		Password: originalPassword,
		Us:       "existing-user-2",
		Email:    "bound@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := newRegisterServiceForTest(t, store, repo, nil)
	err := service.ChangePassword(context.Background(), register.ChangePasswordInput{
		NewPassword: "NewPassword123!",
		Email:       "other@example.com",
		SessionID:   "verified-reset-session",
	})

	// Assumption: reset-password flow must bind ChangePassword email to the verified
	// session email and reject mismatches before any password write occurs.
	if !errors.Is(err, register.ErrSessionEmailMismatch) {
		t.Fatalf("err = %v, want ErrSessionEmailMismatch", err)
	}
	if store.deleted["verified-reset-session"] {
		t.Fatal("session was deleted on rejected password change")
	}

	assertUserPasswordHash(t, db, "bound@example.com", originalPassword)
}

func newRegisterServiceForTest(t *testing.T, store *fakeSessionStore, repo *testPostgresRepository, notifier *fakeRegistrationEmailSender) register.RegService {
	t.Helper()

	if repo == nil {
		repo = &testPostgresRepository{db: newRegisterDB(t)}
	}
	if notifier == nil {
		notifier = &fakeRegistrationEmailSender{}
	}

	service := register.NewRegisterSrvWithEmailSender(
		&testLogger{},
		store,
		repo,
		&config.Config{},
		utils.NewJWTUtils("test-secret"),
		notifier,
	)
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

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword(%q): %v", password, err)
	}
	return hash
}

func assertUserPasswordHash(t *testing.T, db *gorm.DB, email, wantHash string) {
	t.Helper()

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		t.Fatalf("find user by email %q: %v", email, err)
	}
	if user.Password != wantHash {
		t.Fatalf("password hash = %q, want %q", user.Password, wantHash)
	}
}
