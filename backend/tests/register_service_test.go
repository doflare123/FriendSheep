package tests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

type fakeRegistrationStore struct {
	createCalls         []register.UserBootstrap
	createResult        register.RegisteredUser
	createErr           error
	changePasswordCalls []changePasswordCall
	changePasswordID    uint
	changePasswordErr   error
}

type changePasswordCall struct {
	email          string
	hashedPassword string
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

func (f *fakeRegistrationStore) CreateUser(_ context.Context, bootstrap register.UserBootstrap) (register.RegisteredUser, error) {
	f.createCalls = append(f.createCalls, bootstrap)
	return f.createResult, f.createErr
}

func (f *fakeRegistrationStore) ChangePasswordByEmail(_ context.Context, email, hashedPassword string) (uint, error) {
	f.changePasswordCalls = append(f.changePasswordCalls, changePasswordCall{
		email:          email,
		hashedPassword: hashedPassword,
	})
	return f.changePasswordID, f.changePasswordErr
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

	service := newRegisterServiceForTest(t, store, register.NewGORMRegistrationStore(repo), notifier)
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

func TestRegisterServiceCreateUserDelegatesAtomicBootstrapToStore(t *testing.T) {
	sessions := newFakeSessionStore()
	sessions.sessions["verified-session"] = &session.Session{
		Code:       "123456",
		IsVerified: true,
		Type:       models.SessionTypeRegister,
		Extra: map[string]string{
			"email": "user@example.com",
		},
	}
	registrationStore := &fakeRegistrationStore{
		createResult: register.RegisteredUser{
			ID:       42,
			Name:     "Valid User",
			Email:    "user@example.com",
			Username: "generated-user",
		},
	}
	notifier := &fakeRegistrationEmailSender{}
	service := newRegisterServiceForTest(t, sessions, registrationStore, notifier)

	authResponse, err := service.CreateUser(context.Background(), register.CreateUserInput{
		Name:      "Valid User",
		Password:  "Password123!",
		Email:     "USER@example.com",
		SessionID: "verified-session",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if authResponse == nil || authResponse.AccessToken == "" || authResponse.RefreshToken == "" {
		t.Fatalf("authResponse = %#v, want generated token pair", authResponse)
	}
	if len(registrationStore.createCalls) != 1 {
		t.Fatalf("CreateUser store calls = %d, want 1", len(registrationStore.createCalls))
	}

	bootstrap := registrationStore.createCalls[0]
	if bootstrap.Name != "Valid User" || bootstrap.Email != "user@example.com" {
		t.Fatalf("bootstrap identity = %#v, want normalized verified identity", bootstrap)
	}
	if bootstrap.Username == "" {
		t.Fatal("bootstrap username is empty")
	}
	if !utils.VerifyPassword(bootstrap.HashedPassword, "Password123!") {
		t.Fatal("bootstrap password is not a hash of the submitted password")
	}
	if sessions.deleted["verified-session"] != true {
		t.Fatal("verified registration session was not deleted after store success")
	}
	if len(notifier.welcomeCalls) != 1 || notifier.welcomeCalls[0].email != "user@example.com" {
		t.Fatalf("welcome calls = %#v, want normalized registered user email", notifier.welcomeCalls)
	}
}

func TestRegisterServiceCreateUserPreservesSessionWhenStoreRejectsDuplicate(t *testing.T) {
	sessions := newFakeSessionStore()
	sessions.sessions["verified-session"] = &session.Session{
		IsVerified: true,
		Type:       models.SessionTypeRegister,
		Extra: map[string]string{
			"email": "duplicate@example.com",
		},
	}
	registrationStore := &fakeRegistrationStore{createErr: register.ErrUserAlreadyExists}
	notifier := &fakeRegistrationEmailSender{}
	service := newRegisterServiceForTest(t, sessions, registrationStore, notifier)

	authResponse, err := service.CreateUser(context.Background(), register.CreateUserInput{
		Name:      "Valid User",
		Password:  "Password123!",
		Email:     "duplicate@example.com",
		SessionID: "verified-session",
	})

	if authResponse != nil {
		t.Fatalf("authResponse = %#v, want nil", authResponse)
	}
	if !errors.Is(err, register.ErrUserAlreadyExists) {
		t.Fatalf("err = %v, want ErrUserAlreadyExists", err)
	}
	if len(registrationStore.createCalls) != 1 {
		t.Fatalf("CreateUser store calls = %d, want 1", len(registrationStore.createCalls))
	}
	if sessions.deleted["verified-session"] {
		t.Fatal("verified registration session was deleted after rejected bootstrap")
	}
	if len(notifier.welcomeCalls) != 0 {
		t.Fatalf("welcome emails = %d, want 0", len(notifier.welcomeCalls))
	}
}

func TestGORMRegistrationStoreMapsDuplicateAndKeepsBootstrapAtomic(t *testing.T) {
	db := newRegisterDB(t)
	registrationStore := register.NewGORMRegistrationStore(&testPostgresRepository{db: db})
	first := register.UserBootstrap{
		Name:           "First User",
		HashedPassword: mustHashPassword(t, "Password123!"),
		Email:          "duplicate@example.com",
		Username:       "first-user",
	}
	if _, err := registrationStore.CreateUser(context.Background(), first); err != nil {
		t.Fatalf("first CreateUser returned error: %v", err)
	}

	duplicate := first
	duplicate.Name = "Duplicate User"
	duplicate.Username = "second-user"
	if _, err := registrationStore.CreateUser(context.Background(), duplicate); !errors.Is(err, register.ErrUserAlreadyExists) {
		t.Fatalf("duplicate CreateUser err = %v, want ErrUserAlreadyExists", err)
	}

	assertCount(t, db, &models.User{}, 1)
	assertCount(t, db, &statsusers.SettingTile{}, 1)
	assertCount(t, db, &statsusers.SessionStats_users{}, 1)
	assertCount(t, db, &statsusers.SideStats_users{}, 1)
}

func TestGORMRegistrationStoreRollsBackWhenDefaultCreationFails(t *testing.T) {
	db := newRegisterDB(t)
	injectedErr := errors.New("injected session stats failure")
	const callbackName = "test:fail_registration_session_stats"

	if err := db.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if _, ok := tx.Statement.Dest.(*statsusers.SessionStats_users); ok {
			tx.AddError(injectedErr)
		}
	}); err != nil {
		t.Fatalf("register failure callback: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Create().Remove(callbackName)
	})

	registrationStore := register.NewGORMRegistrationStore(&testPostgresRepository{db: db})
	_, err := registrationStore.CreateUser(context.Background(), register.UserBootstrap{
		Name:           "Rollback User",
		HashedPassword: mustHashPassword(t, "Password123!"),
		Email:          "rollback@example.com",
		Username:       "rollback-user",
	})
	if !errors.Is(err, injectedErr) {
		t.Fatalf("CreateUser err = %v, want injected failure", err)
	}

	assertCount(t, db, &models.User{}, 0)
	assertCount(t, db, &statsusers.SettingTile{}, 0)
	assertCount(t, db, &statsusers.SessionStats_users{}, 0)
	assertCount(t, db, &statsusers.SideStats_users{}, 0)
}

func TestGORMRegistrationStoreRollsBackWhenContextIsCanceledAfterLastInsert(t *testing.T) {
	db := newRegisterDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	const callbackName = "test:cancel_registration_after_side_stats"
	if err := db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if _, ok := tx.Statement.Dest.(*statsusers.SideStats_users); ok && tx.Error == nil {
			cancel()
		}
	}); err != nil {
		t.Fatalf("register cancellation callback: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Create().Remove(callbackName)
	})

	registrationStore := register.NewGORMRegistrationStore(&testPostgresRepository{db: db})
	_, err := registrationStore.CreateUser(ctx, register.UserBootstrap{
		Name:           "Canceled User",
		HashedPassword: mustHashPassword(t, "Password123!"),
		Email:          "canceled@example.com",
		Username:       "canceled-user",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateUser err = %v, want context.Canceled", err)
	}

	assertCount(t, db, &models.User{}, 0)
	assertCount(t, db, &statsusers.SettingTile{}, 0)
	assertCount(t, db, &statsusers.SessionStats_users{}, 0)
	assertCount(t, db, &statsusers.SideStats_users{}, 0)
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

	service := newRegisterServiceForTest(t, store, register.NewGORMRegistrationStore(repo), notifier)
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

	service := newRegisterServiceForTest(t, store, register.NewGORMRegistrationStore(repo), notifier)
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

	service := newRegisterServiceForTest(t, store, register.NewGORMRegistrationStore(repo), nil)
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

	service := newRegisterServiceForTest(t, store, register.NewGORMRegistrationStore(repo), nil)
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

func TestRegisterServiceChangePasswordDelegatesToStoreAndDeletesSession(t *testing.T) {
	sessions := newFakeSessionStore()
	sessions.sessions["verified-reset-session"] = &session.Session{
		IsVerified: true,
		Type:       models.SessionTypeResetPassword,
		Extra: map[string]string{
			"email": "user@example.com",
		},
	}
	registrationStore := &fakeRegistrationStore{changePasswordID: 51}
	service := newRegisterServiceForTest(t, sessions, registrationStore, nil)

	err := service.ChangePassword(context.Background(), register.ChangePasswordInput{
		NewPassword: "NewPassword123!",
		Email:       "USER@example.com",
		SessionID:   "verified-reset-session",
	})
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	if len(registrationStore.changePasswordCalls) != 1 {
		t.Fatalf("ChangePasswordByEmail calls = %d, want 1", len(registrationStore.changePasswordCalls))
	}
	call := registrationStore.changePasswordCalls[0]
	if call.email != "user@example.com" {
		t.Fatalf("ChangePasswordByEmail email = %q, want normalized bound email", call.email)
	}
	if !utils.VerifyPassword(call.hashedPassword, "NewPassword123!") {
		t.Fatal("ChangePasswordByEmail received an invalid password hash")
	}
	if !sessions.deleted["verified-reset-session"] {
		t.Fatal("reset-password session was not deleted after store success")
	}
}

func TestRegisterServiceChangePasswordPreservesSessionWhenUserIsMissing(t *testing.T) {
	sessions := newFakeSessionStore()
	sessions.sessions["verified-reset-session"] = &session.Session{
		IsVerified: true,
		Type:       models.SessionTypeResetPassword,
		Extra: map[string]string{
			"email": "missing@example.com",
		},
	}
	registrationStore := &fakeRegistrationStore{changePasswordErr: register.ErrUserNotFound}
	service := newRegisterServiceForTest(t, sessions, registrationStore, nil)

	err := service.ChangePassword(context.Background(), register.ChangePasswordInput{
		NewPassword: "NewPassword123!",
		Email:       "missing@example.com",
		SessionID:   "verified-reset-session",
	})
	if !errors.Is(err, register.ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
	if len(registrationStore.changePasswordCalls) != 1 {
		t.Fatalf("ChangePasswordByEmail calls = %d, want 1", len(registrationStore.changePasswordCalls))
	}
	if sessions.deleted["verified-reset-session"] {
		t.Fatal("reset-password session was deleted after store failure")
	}
}

func TestGORMRegistrationStoreChangePasswordMapsMissingUser(t *testing.T) {
	db := newRegisterDB(t)
	registrationStore := register.NewGORMRegistrationStore(&testPostgresRepository{db: db})

	_, err := registrationStore.ChangePasswordByEmail(context.Background(), "missing@example.com", mustHashPassword(t, "NewPassword123!"))
	if !errors.Is(err, register.ErrUserNotFound) {
		t.Fatalf("ChangePasswordByEmail err = %v, want ErrUserNotFound", err)
	}
}

func TestGORMRegistrationStoreChangesPasswordAndReturnsUserID(t *testing.T) {
	db := newRegisterDB(t)
	oldHash := mustHashPassword(t, "OldPassword123!")
	user := models.User{
		Name:     "Existing User",
		Password: oldHash,
		Us:       "password-user",
		Email:    "password@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	newHash := mustHashPassword(t, "NewPassword123!")
	registrationStore := register.NewGORMRegistrationStore(&testPostgresRepository{db: db})
	userID, err := registrationStore.ChangePasswordByEmail(context.Background(), user.Email, newHash)
	if err != nil {
		t.Fatalf("ChangePasswordByEmail returned error: %v", err)
	}
	if userID != user.ID {
		t.Fatalf("ChangePasswordByEmail userID = %d, want %d", userID, user.ID)
	}
	assertUserPasswordHash(t, db, user.Email, newHash)
}

func newRegisterServiceForTest(t *testing.T, store *fakeSessionStore, registrationStore register.RegistrationStore, notifier *fakeRegistrationEmailSender) register.RegService {
	t.Helper()

	if registrationStore == nil {
		registrationStore = &fakeRegistrationStore{}
	}
	if notifier == nil {
		notifier = &fakeRegistrationEmailSender{}
	}

	service := register.NewRegisterSrvWithEmailSender(
		&testLogger{},
		store,
		registrationStore,
		&config.Config{},
		utils.NewJWTUtils("test-secret"),
		notifier,
	)
	return service
}

func newRegisterDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "-" +
		strconv.FormatInt(time.Now().UnixNano(), 10) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
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
	if err := db.FirstOrCreate(&models.DaysWeek{ID: 1, Name: "Monday"}).Error; err != nil {
		t.Fatalf("seed default day: %v", err)
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
