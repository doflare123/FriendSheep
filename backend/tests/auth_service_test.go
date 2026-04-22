package tests

import (
	"errors"
	"strings"
	"testing"

	"friendship/models/dto"
	"friendship/services"
	"friendship/utils"
)

type fakeAuthRepository struct {
	user        services.AuthUser
	userErr     error
	adminGroups []dto.AdminGroupResponse
	adminErr    error
}

func (r *fakeAuthRepository) FindAuthUserByEmail(email string) (services.AuthUser, error) {
	if r.userErr != nil {
		return services.AuthUser{}, r.userErr
	}
	return r.user, nil
}

func (r *fakeAuthRepository) FindAuthUserByID(id uint) (services.AuthUser, error) {
	if r.userErr != nil {
		return services.AuthUser{}, r.userErr
	}
	return r.user, nil
}

func (r *fakeAuthRepository) GetAuthAdminGroups(userID uint) ([]dto.AdminGroupResponse, error) {
	if r.adminErr != nil {
		return nil, r.adminErr
	}
	return r.adminGroups, nil
}

func TestAuthServiceLoginUsesRepositoryPort(t *testing.T) {
	passwordHash, err := utils.HashPassword("Password123!")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	service := services.NewAuthServiceWithRepository(
		&testLogger{},
		utils.NewJWTUtils("test-secret"),
		&fakeAuthRepository{
			user: services.AuthUser{
				ID:       10,
				Name:     "Valid User",
				Password: passwordHash,
				Us:       "valid_user",
				Image:    "avatar.png",
			},
			adminGroups: []dto.AdminGroupResponse{},
		},
	)

	res, err := service.Login("user@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("Login returned empty tokens")
	}
}

func TestAuthServiceLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := utils.HashPassword("Password123!")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	service := services.NewAuthServiceWithRepository(
		&testLogger{},
		utils.NewJWTUtils("test-secret"),
		&fakeAuthRepository{
			user: services.AuthUser{
				ID:       10,
				Name:     "Valid User",
				Password: passwordHash,
				Us:       "valid_user",
				Image:    "avatar.png",
			},
		},
	)

	_, err = service.Login("user@example.com", "wrong")
	if err == nil {
		t.Fatal("Login returned nil error for wrong password")
	}
}

func TestAuthServiceRefreshUsesRepositoryPort(t *testing.T) {
	jwtService := utils.NewJWTUtils("test-secret")
	tokenPair, err := jwtService.GenerateTokenPair(10, "Valid User", "valid_user", "avatar.png")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	service := services.NewAuthServiceWithRepository(
		&testLogger{},
		jwtService,
		&fakeAuthRepository{
			user: services.AuthUser{
				ID:    10,
				Name:  "Valid User",
				Us:    "valid_user",
				Image: "avatar.png",
			},
			adminGroups: []dto.AdminGroupResponse{},
		},
	)

	res, err := service.RefreshTokens(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshTokens returned error: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("RefreshTokens returned empty tokens")
	}
}

func TestAuthServiceRefreshWrapsUserLookupError(t *testing.T) {
	jwtService := utils.NewJWTUtils("test-secret")
	tokenPair, err := jwtService.GenerateTokenPair(10, "Valid User", "valid_user", "avatar.png")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	service := services.NewAuthServiceWithRepository(
		&testLogger{},
		jwtService,
		&fakeAuthRepository{userErr: errors.New("not found")},
	)

	_, err = service.RefreshTokens(tokenPair.RefreshToken)
	if err == nil {
		t.Fatal("RefreshTokens returned nil error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("RefreshTokens error = %q, want wrapped lookup error", err.Error())
	}
}
