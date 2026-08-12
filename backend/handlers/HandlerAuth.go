package handlers

import (
	"errors"
	"friendship/services"
	"friendship/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	LogoutAll(c *gin.Context)
	Me(c *gin.Context)
}

type authHandler struct {
	srv services.AuthService
}

func NewAuthHandler(srv services.AuthService) AuthHandler {
	return &authHandler{srv: srv}
}

type UserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=256"`
}

// RefreshToken godoc
// @Summary      Обновить токены авторизации
// @Description  Одноразово ротирует opaque refresh-токен и возвращает новую пару токенов с актуальными данными me.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refreshRequest  body      RefreshRequest    true  "Opaque refresh-токен"
// @Success      200             {object}  dto.AuthResponse
// @Failure      400             {object}  dto.ErrorResponse
// @Failure      401             {object}  dto.ErrorResponse
// @Failure      503             {object}  dto.ErrorResponse
// @Router       /api/v2/auth/refresh [post]
func (h *authHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid_request", utils.WithMessage("Refresh-токен обязателен"))
		return
	}

	authRes, err := h.srv.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidRefreshToken),
			errors.Is(err, services.ErrRefreshTokenReplay),
			errors.Is(err, services.ErrAuthSessionRevoked),
			errors.Is(err, services.ErrAuthSessionNotFound):
			utils.Unauthorized(c, "invalid_refresh_token", utils.WithMessage("Невалидный, истекший или уже использованный refresh-токен"))
		default:
			utils.JSONError(c, http.StatusServiceUnavailable, "authentication_unavailable", utils.WithMessage("Сервис авторизации временно недоступен"))
		}
		return
	}

	writeAuthResponse(c, authRes)
}

// Login godoc
// @Summary      Вход пользователя
// @Description  Проверяет email и пароль, создаёт серверную auth-сессию и возвращает access/refresh-токены с данными me.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      UserRequest       true  "Email и пароль"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      503   {object}  dto.ErrorResponse
// @Router       /api/v2/auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid_request", utils.WithMessage("Некорректный формат email или пароля"))
		return
	}

	authRes, err := h.srv.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.Unauthorized(c, "authentication_failed", utils.WithMessage("Неверный email или пароль"))
			return
		}
		utils.JSONError(c, http.StatusServiceUnavailable, "authentication_unavailable", utils.WithMessage("Сервис авторизации временно недоступен"))
		return
	}

	writeAuthResponse(c, authRes)
}

// Logout godoc
// @Summary      Завершить текущую auth-сессию
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      204
// @Failure      401 {object} dto.ErrorResponse
// @Failure      503 {object} dto.ErrorResponse
// @Router       /api/v2/auth/logout [post]
func (h *authHandler) Logout(c *gin.Context) {
	sessionID := c.GetString("authSessionID")
	if sessionID == "" {
		utils.Unauthorized(c, "missing_auth_session", utils.WithMessage("Сессия авторизации не найдена"))
		return
	}
	if err := h.srv.RevokeCurrent(c.Request.Context(), sessionID); err != nil {
		utils.JSONError(c, http.StatusServiceUnavailable, "authentication_unavailable", utils.WithMessage("Не удалось завершить сессию"))
		return
	}
	c.Status(http.StatusNoContent)
}

// LogoutAll godoc
// @Summary      Завершить все auth-сессии пользователя
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      204
// @Failure      401 {object} dto.ErrorResponse
// @Failure      503 {object} dto.ErrorResponse
// @Router       /api/v2/auth/logout-all [post]
func (h *authHandler) LogoutAll(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
		return
	}
	if err := h.srv.RevokeAll(c.Request.Context(), userID); err != nil {
		utils.JSONError(c, http.StatusServiceUnavailable, "authentication_unavailable", utils.WithMessage("Не удалось завершить сессии"))
		return
	}
	c.Status(http.StatusNoContent)
}

// Me godoc
// @Summary      Получить текущего пользователя
// @Description  Возвращает актуальные данные профиля, которые не хранятся в access-токене.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} dto.AuthMeResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/v2/user/me [get]
func (h *authHandler) Me(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
		return
	}

	me, err := h.srv.GetMe(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrAuthUserNotFound) {
			utils.NotFound(c, "user_not_found", utils.WithMessage("Пользователь не найден"))
			return
		}
		utils.InternalError(c, "user_lookup_failed", utils.WithMessage("Не удалось получить данные пользователя"))
		return
	}

	c.JSON(http.StatusOK, me)
}

func writeAuthResponse(c *gin.Context, response any) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, response)
}
