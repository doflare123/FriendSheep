package tests

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"friendship/models/dto"
	"friendship/services"
	"friendship/utils"
)

type fakeAuthRepository struct {
	user             services.AuthUser
	userErr          error
	adminGroups      []dto.AdminGroupResponse
	adminErr         error
	findByEmailCalls int
	findByIDCalls    int
}

func (r *fakeAuthRepository) FindAuthUserByEmail(_ context.Context, email string) (services.AuthUser, error) {
	r.findByEmailCalls++
	if r.userErr != nil {
		return services.AuthUser{}, r.userErr
	}
	return r.user, nil
}

func (r *fakeAuthRepository) FindAuthUserByID(_ context.Context, id uint) (services.AuthUser, error) {
	r.findByIDCalls++
	if r.userErr != nil {
		return services.AuthUser{}, r.userErr
	}
	return r.user, nil
}

func (r *fakeAuthRepository) GetAuthAdminGroups(_ context.Context, userID uint) ([]dto.AdminGroupResponse, error) {
	if r.adminErr != nil {
		return nil, r.adminErr
	}
	return r.adminGroups, nil
}

type fakeAuthSessionStore struct {
	createResult services.AuthSession
	createErr    error
	rotateResult services.AuthSession
	rotateErr    error
	active       bool
	activeErr    error

	createCalls    []services.CreateAuthSessionInput
	rotateCalls    []services.RotateAuthSessionInput
	revokeCalls    []string
	revokeAllCalls []uint
}

func (s *fakeAuthSessionStore) CreateSession(_ context.Context, input services.CreateAuthSessionInput) (services.AuthSession, error) {
	s.createCalls = append(s.createCalls, input)
	return s.createResult, s.createErr
}

func (s *fakeAuthSessionStore) RotateSession(_ context.Context, input services.RotateAuthSessionInput) (services.AuthSession, error) {
	s.rotateCalls = append(s.rotateCalls, input)
	return s.rotateResult, s.rotateErr
}

func (s *fakeAuthSessionStore) RevokeSession(_ context.Context, sessionID string) error {
	s.revokeCalls = append(s.revokeCalls, sessionID)
	return nil
}

func (s *fakeAuthSessionStore) RevokeAllUserSessions(_ context.Context, userID uint) error {
	s.revokeAllCalls = append(s.revokeAllCalls, userID)
	return nil
}

func (s *fakeAuthSessionStore) HasActiveSession(_ context.Context, _ string) (bool, error) {
	return s.active, s.activeErr
}

func TestAuthServiceLoginCreatesServerSideSessionAndReturnsMe(t *testing.T) {
	passwordHash := mustAuthPasswordHash(t, "Password123!")
	repository := &fakeAuthRepository{
		user: services.AuthUser{
			ID:       10,
			Name:     "Valid User",
			Password: passwordHash,
			Us:       "valid_user",
			Image:    "avatar.png",
		},
		adminGroups: []dto.AdminGroupResponse{},
	}
	sessions := &fakeAuthSessionStore{createResult: services.AuthSession{
		SessionID:        "session-10",
		UserID:           10,
		RefreshToken:     "session-10.opaque-secret",
		RefreshExpiresAt: time.Now().Add(time.Hour),
	}}
	service := newAuthServiceForTest(repository, sessions)

	res, err := service.Login(context.Background(), "user@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if len(sessions.createCalls) != 1 || sessions.createCalls[0].UserID != 10 {
		t.Fatalf("CreateSession calls = %#v, want one call for user 10", sessions.createCalls)
	}
	if res.RefreshToken != "session-10.opaque-secret" {
		t.Fatalf("refresh token = %q, want opaque store token", res.RefreshToken)
	}
	assertOpaqueRefreshToken(t, res.RefreshToken)
	assertAuthResponseIdentity(t, res, repository.user)

	claims, err := newConfiguredJWTUtils().ParseAccessToken(res.AccessToken)
	if err != nil {
		t.Fatalf("parse login access token: %v", err)
	}
	if claims.UserID != 10 || claims.SessionID != "session-10" {
		t.Fatalf("access identity = user %d session %q", claims.UserID, claims.SessionID)
	}
}

func TestAuthServiceLoginRejectsWrongPasswordWithoutCreatingSession(t *testing.T) {
	repository := &fakeAuthRepository{user: services.AuthUser{
		ID:       10,
		Password: mustAuthPasswordHash(t, "Password123!"),
	}}
	sessions := &fakeAuthSessionStore{}
	service := newAuthServiceForTest(repository, sessions)

	if _, err := service.Login(context.Background(), "user@example.com", "wrong"); err == nil {
		t.Fatal("Login returned nil error for wrong password")
	}
	if len(sessions.createCalls) != 0 {
		t.Fatalf("CreateSession calls = %d, want 0", len(sessions.createCalls))
	}
}

func TestAuthServiceLoginDistinguishesMissingUserFromRepositoryFailure(t *testing.T) {
	tests := []struct {
		name      string
		repoError error
		wantError error
	}{
		{name: "missing user", repoError: services.ErrAuthUserNotFound, wantError: services.ErrInvalidCredentials},
		{name: "repository unavailable", repoError: errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newAuthServiceForTest(&fakeAuthRepository{userErr: tt.repoError}, &fakeAuthSessionStore{})
			_, err := service.Login(context.Background(), "user@example.com", "Password123!")
			if err == nil {
				t.Fatal("Login returned nil error")
			}
			if tt.wantError != nil && !errors.Is(err, tt.wantError) {
				t.Fatalf("Login error = %v, want %v", err, tt.wantError)
			}
			if tt.wantError == nil && errors.Is(err, services.ErrInvalidCredentials) {
				t.Fatalf("repository failure was collapsed into ErrInvalidCredentials: %v", err)
			}
		})
	}
}

func TestAuthServiceRefreshUsesAtomicRotationAndReturnsMe(t *testing.T) {
	repository := &fakeAuthRepository{user: services.AuthUser{
		ID: 10, Name: "Valid User", Us: "valid_user", Image: "avatar.png",
	}}
	sessions := &fakeAuthSessionStore{rotateResult: services.AuthSession{
		SessionID:        "session-10",
		UserID:           10,
		RefreshToken:     "session-10.new-opaque-secret",
		RefreshExpiresAt: time.Now().Add(time.Hour),
	}}
	service := newAuthServiceForTest(repository, sessions)

	res, err := service.RefreshTokens(context.Background(), "session-10.old-opaque-secret")
	if err != nil {
		t.Fatalf("RefreshTokens returned error: %v", err)
	}
	if len(sessions.rotateCalls) != 1 || sessions.rotateCalls[0].RefreshToken != "session-10.old-opaque-secret" {
		t.Fatalf("RotateSession calls = %#v", sessions.rotateCalls)
	}
	if len(sessions.createCalls) != 0 {
		t.Fatalf("CreateSession calls = %d, want rotation in existing session", len(sessions.createCalls))
	}
	if res.RefreshToken != "session-10.new-opaque-secret" {
		t.Fatalf("rotated refresh token = %q", res.RefreshToken)
	}
	assertOpaqueRefreshToken(t, res.RefreshToken)
	assertAuthResponseIdentity(t, res, repository.user)
}

func TestAuthServiceRefreshReplayIsRejectedWithoutUserLookup(t *testing.T) {
	repository := &fakeAuthRepository{user: services.AuthUser{ID: 10}}
	sessions := &fakeAuthSessionStore{rotateErr: services.ErrRefreshTokenReplay}
	service := newAuthServiceForTest(repository, sessions)

	_, err := service.RefreshTokens(context.Background(), "session-10.replayed-secret")
	if !errors.Is(err, services.ErrRefreshTokenReplay) {
		t.Fatalf("RefreshTokens err = %v, want ErrRefreshTokenReplay", err)
	}
	if repository.findByIDCalls != 0 {
		t.Fatalf("FindAuthUserByID calls = %d, want 0 after replay", repository.findByIDCalls)
	}
}

func TestAuthServiceRefreshRevokesSessionWhenUserLookupFails(t *testing.T) {
	lookupErr := errors.New("not found")
	repository := &fakeAuthRepository{userErr: lookupErr}
	sessions := &fakeAuthSessionStore{rotateResult: services.AuthSession{
		SessionID:    "session-10",
		UserID:       10,
		RefreshToken: "session-10.new-secret",
	}}
	service := newAuthServiceForTest(repository, sessions)

	_, err := service.RefreshTokens(context.Background(), "session-10.old-secret")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("RefreshTokens err = %v, want wrapped lookup error", err)
	}
	if len(sessions.revokeCalls) != 1 || sessions.revokeCalls[0] != "session-10" {
		t.Fatalf("RevokeSession calls = %#v, want compromised session revoked", sessions.revokeCalls)
	}
}

func TestAuthServiceLogoutCurrentAndAll(t *testing.T) {
	sessions := &fakeAuthSessionStore{}
	service := newAuthServiceForTest(&fakeAuthRepository{}, sessions)

	if err := service.RevokeCurrent(context.Background(), "session-current"); err != nil {
		t.Fatalf("RevokeCurrent returned error: %v", err)
	}
	if err := service.RevokeAll(context.Background(), 42); err != nil {
		t.Fatalf("RevokeAll returned error: %v", err)
	}
	if len(sessions.revokeCalls) != 1 || sessions.revokeCalls[0] != "session-current" {
		t.Fatalf("RevokeSession calls = %#v", sessions.revokeCalls)
	}
	if len(sessions.revokeAllCalls) != 1 || sessions.revokeAllCalls[0] != 42 {
		t.Fatalf("RevokeAllUserSessions calls = %#v", sessions.revokeAllCalls)
	}
}

func TestAuthServiceGetMeLoadsCurrentProfile(t *testing.T) {
	repository := &fakeAuthRepository{user: services.AuthUser{
		ID: 17, Name: "Current User", Us: "current", Image: "current.png",
	}}
	service := newAuthServiceForTest(repository, &fakeAuthSessionStore{})

	me, err := service.GetMe(context.Background(), 17)
	if err != nil {
		t.Fatalf("GetMe returned error: %v", err)
	}
	if me.ID != 17 || me.Name != "Current User" || me.Us != "current" || me.Image != "current.png" {
		t.Fatalf("me = %#v, want current profile", me)
	}
}

func TestAuthServiceGetMeDistinguishesMissingUserFromRepositoryFailure(t *testing.T) {
	missingService := newAuthServiceForTest(&fakeAuthRepository{userErr: services.ErrAuthUserNotFound}, &fakeAuthSessionStore{})
	if _, err := missingService.GetMe(context.Background(), 17); !errors.Is(err, services.ErrAuthUserNotFound) {
		t.Fatalf("GetMe missing user error = %v, want ErrAuthUserNotFound", err)
	}

	unavailableService := newAuthServiceForTest(&fakeAuthRepository{userErr: errors.New("database unavailable")}, &fakeAuthSessionStore{})
	if _, err := unavailableService.GetMe(context.Background(), 17); err == nil || errors.Is(err, services.ErrAuthUserNotFound) {
		t.Fatalf("GetMe repository failure = %v, must not be ErrAuthUserNotFound", err)
	}
}

func newAuthServiceForTest(repository services.AuthUserAdminGroupRepository, sessions services.AuthSessionStore) services.AuthService {
	return services.NewAuthServiceWithRepository(&testLogger{}, newConfiguredJWTUtils(), repository, sessions)
}

func mustAuthPasswordHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	return hash
}

func assertOpaqueRefreshToken(t *testing.T, refreshToken string) {
	t.Helper()
	if refreshToken == "" {
		t.Fatal("refresh token is empty")
	}
	if strings.Count(refreshToken, ".") != 1 {
		t.Fatalf("refresh token %q is not opaque <sid>.<secret> format", refreshToken)
	}
	if _, _, err := decodeJWTPartsWithoutFailure(refreshToken); err == nil {
		t.Fatalf("refresh token %q unexpectedly decodes as JWT", refreshToken)
	}
}

func assertAuthResponseIdentity(t *testing.T, response dto.AuthResponse, want services.AuthUser) {
	t.Helper()
	if response.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	if response.Me.ID != want.ID || response.Me.Name != want.Name || response.Me.Us != want.Us || response.Me.Image != want.Image {
		t.Fatalf("response me = %#v, want identity %#v", response.Me, want)
	}
}

func decodeJWTPartsWithoutFailure(token string) (string, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", errors.New("not a compact JWT")
	}
	return parts[0], parts[1], nil
}
