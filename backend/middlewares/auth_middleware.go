package middlewares

import (
	"context"
	"friendship/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthSessionReader interface {
	HasActiveSession(ctx context.Context, sessionID string) (bool, error)
}

type AuthMiddleware struct {
	jwtUtils    *utils.JWTUtils
	sessions    AuthSessionReader
	rateLimiter *RateLimitMiddleware
}

func NewAuthMiddleware(jwtService *utils.JWTUtils, sessionReaders ...AuthSessionReader) *AuthMiddleware {
	var sessions AuthSessionReader
	if len(sessionReaders) > 0 {
		sessions = sessionReaders[0]
	}
	return &AuthMiddleware{jwtUtils: jwtService, sessions: sessions}
}

func (m *AuthMiddleware) SetRateLimiter(rateLimiter *RateLimitMiddleware) {
	if m == nil {
		return
	}
	m.rateLimiter = rateLimiter
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.AbortJSONError(c, http.StatusUnauthorized, "missing_access_token", utils.WithMessage("Отсутствует токен авторизации"))
			return
		}

		claims, ok := m.authenticate(c, authHeader)
		if !ok {
			return
		}

		setAuthContext(c, claims)
		if m.rateLimiter != nil && !m.rateLimiter.ApplyAuthenticatedLimit(c, claims.UserID) {
			return
		}

		c.Next()
	}
}

// OptionalAuth keeps requests without credentials anonymous. If a client sends
// an Authorization header, invalid or revoked credentials are rejected.
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		claims, ok := m.authenticate(c, authHeader)
		if !ok {
			return
		}

		setAuthContext(c, claims)
		if m.rateLimiter != nil && !m.rateLimiter.ApplyAuthenticatedLimit(c, claims.UserID) {
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) authenticate(c *gin.Context, authHeader string) (*utils.Claims, bool) {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		utils.AbortJSONError(c, http.StatusUnauthorized, "invalid_authorization_header", utils.WithMessage("Неверный формат токена авторизации"))
		return nil, false
	}

	claims, err := m.jwtUtils.ParseAccessToken(parts[1])
	if err != nil {
		utils.AbortJSONError(c, http.StatusUnauthorized, "invalid_access_token", utils.WithMessage("Невалидный или истекший токен доступа"))
		return nil, false
	}

	if m.sessions == nil {
		utils.AbortJSONError(c, http.StatusServiceUnavailable, "authentication_unavailable", utils.WithMessage("Проверка сессии авторизации не настроена"))
		return nil, false
	}
	active, err := m.sessions.HasActiveSession(c.Request.Context(), claims.SessionID)
	if err != nil {
		utils.AbortJSONError(c, http.StatusServiceUnavailable, "authentication_unavailable", utils.WithMessage("Сервис авторизации временно недоступен"))
		return nil, false
	}
	if !active {
		utils.AbortJSONError(c, http.StatusUnauthorized, "revoked_access_token", utils.WithMessage("Сессия авторизации завершена"))
		return nil, false
	}

	return claims, true
}

func setAuthContext(c *gin.Context, claims *utils.Claims) {
	c.Set("userID", claims.UserID)
	c.Set("authSessionID", claims.SessionID)
	c.Set("authTokenID", claims.ID)
}
