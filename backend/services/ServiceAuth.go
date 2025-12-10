package services

import (
	"fmt"
	"friendship/config"
	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	"friendship/models/groups"
	"friendship/repository"
	"friendship/utils"
)

type AuthService interface {
	Login(email, password string) (dto.AuthResponse, error)
	RefreshTokens(refreshToken string) (dto.AuthResponse, error)
}

type authService struct {
	logger     logger.Logger
	cfg        config.Config
	rep        repository.PostgresRepository
	jwtService *utils.JWTUtils
}

func NewAuthService(logger logger.Logger, cfg config.Config, jwtService *utils.JWTUtils, rep repository.PostgresRepository) AuthService {
	return &authService{
		logger:     logger,
		cfg:        cfg,
		rep:        rep,
		jwtService: jwtService,
	}
}

func (a *authService) Login(email, password string) (dto.AuthResponse, error) {
	if email == "" || password == "" {
		return dto.AuthResponse{}, fmt.Errorf("email и пароль обязательны")
	}

	user, err := new(models.User).FindUserByEmailPassword(email, a.rep)
	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("неверный логин или пароль")
	}

	if !utils.VerifyPassword(user.Password, password) {
		return dto.AuthResponse{}, fmt.Errorf("неверный логин или пароль")
	}

	tokenPair, err := a.jwtService.GenerateTokenPair(user.ID, user.Name, user.Us, user.Image)
	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	adminGroups, err := new(groups.Group).GetAdminGroups(user.ID, a.rep)
	if err != nil {
		a.logger.Warn("Предупреждение: не удалось загрузить группы администратора: %v\n", err)
		adminGroups = []dto.AdminGroupResponse{}
	}

	return dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		AdminGroups:  adminGroups,
	}, nil
}

func (a *authService) RefreshTokens(refreshToken string) (dto.AuthResponse, error) {
	if refreshToken == "" {
		return dto.AuthResponse{}, fmt.Errorf("refresh токен обязателен")
	}

	userID, err := a.jwtService.ParseRefreshToken(refreshToken)
	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("невалидный refresh токен: %w", err)
	}

	user, err := new(models.User).FindUserByID(userID, a.rep)
	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("пользователь не найден: %w", err)
	}

	tokenPair, err := a.jwtService.GenerateTokenPair(uint(user.Id), user.Username, user.Us, user.Image)
	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	adminGroups, err := new(groups.Group).GetAdminGroups(uint(user.Id), a.rep)
	if err != nil {
		adminGroups = []dto.AdminGroupResponse{}
	}

	return dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		AdminGroups:  adminGroups,
	}, nil
}
