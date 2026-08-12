package services

import (
	"context"
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models/dto"
	"friendship/utils"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (dto.AuthResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (dto.AuthResponse, error)
	IssueTokens(ctx context.Context, userID uint) (dto.AuthResponse, error)
	GetMe(ctx context.Context, userID uint) (dto.AuthMeResponse, error)
	RevokeCurrent(ctx context.Context, sessionID string) error
	RevokeAll(ctx context.Context, userID uint) error
}

var (
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrAuthUserNotFound   = errors.New("пользователь не найден")
)

type authService struct {
	logger       logger.Logger
	rep          AuthUserAdminGroupRepository
	jwtService   *utils.JWTUtils
	sessionStore AuthSessionStore
}

func NewAuthService(logger logger.Logger, jwtService *utils.JWTUtils, rep AuthGORMStore, sessionStore AuthSessionStore) AuthService {
	return NewAuthServiceWithRepository(logger, jwtService, NewGORMAuthRepository(rep), sessionStore)
}

func NewAuthServiceWithRepository(logger logger.Logger, jwtService *utils.JWTUtils, rep AuthUserAdminGroupRepository, sessionStore AuthSessionStore) AuthService {
	return &authService{
		logger:       logger,
		rep:          rep,
		jwtService:   jwtService,
		sessionStore: sessionStore,
	}
}

func (a *authService) Login(ctx context.Context, email, password string) (dto.AuthResponse, error) {
	if email == "" || password == "" {
		return dto.AuthResponse{}, ErrInvalidCredentials
	}

	user, err := a.rep.FindAuthUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrAuthUserNotFound) {
			return dto.AuthResponse{}, ErrInvalidCredentials
		}
		return dto.AuthResponse{}, fmt.Errorf("ошибка поиска пользователя для входа: %w", err)
	}

	if !utils.VerifyPassword(user.Password, password) {
		return dto.AuthResponse{}, ErrInvalidCredentials
	}

	return a.issueTokensForUser(ctx, user)
}

func (a *authService) RefreshTokens(ctx context.Context, refreshToken string) (dto.AuthResponse, error) {
	if refreshToken == "" {
		return dto.AuthResponse{}, fmt.Errorf("refresh токен обязателен")
	}

	session, err := a.sessionStore.RotateSession(ctx, RotateAuthSessionInput{RefreshToken: refreshToken})
	if err != nil {
		if errors.Is(err, ErrRefreshTokenReplay) {
			return dto.AuthResponse{}, ErrRefreshTokenReplay
		}
		if errors.Is(err, ErrAuthSessionNotFound) || errors.Is(err, ErrAuthSessionRevoked) || errors.Is(err, ErrInvalidRefreshToken) {
			return dto.AuthResponse{}, ErrInvalidRefreshToken
		}
		return dto.AuthResponse{}, fmt.Errorf("ошибка обновления auth сессии: %w", err)
	}

	user, err := a.rep.FindAuthUserByID(ctx, session.UserID)
	if err != nil {
		a.revokeSessionOnFailure(ctx, session.SessionID, "refresh user lookup failed")
		if errors.Is(err, ErrAuthUserNotFound) {
			return dto.AuthResponse{}, ErrAuthUserNotFound
		}
		return dto.AuthResponse{}, fmt.Errorf("ошибка поиска пользователя при обновлении токенов: %w", err)
	}

	return a.authResponseForSession(ctx, user, session, "refresh access token generation failed")
}

func (a *authService) IssueTokens(ctx context.Context, userID uint) (dto.AuthResponse, error) {
	if userID == 0 {
		return dto.AuthResponse{}, fmt.Errorf("некорректный идентификатор пользователя")
	}

	user, err := a.rep.FindAuthUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrAuthUserNotFound) {
			return dto.AuthResponse{}, ErrAuthUserNotFound
		}
		return dto.AuthResponse{}, fmt.Errorf("ошибка поиска пользователя для выдачи токенов: %w", err)
	}

	return a.issueTokensForUser(ctx, user)
}

func (a *authService) GetMe(ctx context.Context, userID uint) (dto.AuthMeResponse, error) {
	if userID == 0 {
		return dto.AuthMeResponse{}, fmt.Errorf("некорректный идентификатор пользователя")
	}

	user, err := a.rep.FindAuthUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrAuthUserNotFound) {
			return dto.AuthMeResponse{}, ErrAuthUserNotFound
		}
		return dto.AuthMeResponse{}, fmt.Errorf("ошибка получения текущего пользователя: %w", err)
	}

	return dto.AuthMeResponse{
		ID:    user.ID,
		Name:  user.Name,
		Us:    user.Us,
		Image: user.Image,
	}, nil
}

func (a *authService) RevokeCurrent(ctx context.Context, sessionID string) error {
	if err := a.sessionStore.RevokeSession(ctx, sessionID); err != nil {
		if errors.Is(err, ErrAuthSessionNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (a *authService) RevokeAll(ctx context.Context, userID uint) error {
	if userID == 0 {
		return fmt.Errorf("некорректный идентификатор пользователя")
	}
	return a.sessionStore.RevokeAllUserSessions(ctx, userID)
}

func (a *authService) issueTokensForUser(ctx context.Context, user AuthUser) (dto.AuthResponse, error) {
	session, err := a.sessionStore.CreateSession(ctx, CreateAuthSessionInput{UserID: user.ID})
	if err != nil {
		return dto.AuthResponse{}, fmt.Errorf("ошибка создания auth сессии: %w", err)
	}

	return a.authResponseForSession(ctx, user, session, "access token generation failed")
}

func (a *authService) authResponseForSession(ctx context.Context, user AuthUser, session AuthSession, revokeReason string) (dto.AuthResponse, error) {
	accessToken, _, _, err := a.jwtService.GenerateAccessToken(user.ID, session.SessionID)
	if err != nil {
		a.revokeSessionOnFailure(ctx, session.SessionID, revokeReason)
		return dto.AuthResponse{}, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	adminGroups, err := a.rep.GetAuthAdminGroups(ctx, user.ID)
	if err != nil {
		a.logger.Warn("Не удалось загрузить группы администратора", "userID", user.ID, "error", err)
		adminGroups = []dto.AdminGroupResponse{}
	}

	return dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(a.jwtService.AccessTTL() / time.Second),
		Me: dto.AuthMeResponse{
			ID:    user.ID,
			Name:  user.Name,
			Us:    user.Us,
			Image: user.Image,
		},
		AdminGroups: adminGroups,
	}, nil
}

func (a *authService) revokeSessionOnFailure(ctx context.Context, sessionID string, reason string) {
	if a.sessionStore == nil || sessionID == "" {
		return
	}
	if err := a.sessionStore.RevokeSession(ctx, sessionID); err != nil {
		a.logger.Warn("Не удалось отозвать auth сессию", "sessionID", sessionID, "reason", reason, "error", err)
	}
}
