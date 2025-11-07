package services

import (
	"friendship/config"
	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	"friendship/repository"
	"friendship/utils"
)

type AuthService interface {
	Login() (dto.AuthResponse, error)
	RefreshTokens(refreshToken string) (dto.AuthResponse, error)
}

type authService struct {
	logger logger.Logger
	cfg    config.Config
	rep    repository.PostgresRepository
}

func NewAuthService(logger logger.Logger, cfg config.Config, rep repository.PostgresRepository) AuthService {
	return &authService{
		logger: logger,
		cfg:    cfg,
		rep:    rep,
	}
}

func (a *authService) Login() (tokenPair dto.AuthResponse, err error) {
	return tokenPair, err
}

func (a *authService) RefreshTokens(refreshToken string) (tokenPair dto.AuthResponse, err error) {
	userId, err := utils.ParseJWT(refreshToken)
	if err != nil {
		return tokenPair, err
	}
	user, err := new(models.User).FindUserByID(*userId, a.rep)
	tokenPair, err = utils.RefreshTokens(refreshToken, user, a.rep, a.cfg.JWTSecretKey)
	return tokenPair, err
}
