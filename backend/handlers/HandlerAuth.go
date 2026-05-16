package handlers

import (
	"friendship/services"
	"friendship/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type authHandler struct {
	srv services.AuthService
}

func NewAuthHandler(srv services.AuthService) AuthHandler {
	return &authHandler{
		srv: srv,
	}
}

type UserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string                        `json:"access_token"`
	RefreshToken string                        `json:"refresh_token"`
	AdminGroups  []services.AdminGroupResponse `json:"admin_groups"`
}

// RefreshToken godoc
// @Summary      Обновить токены авторизации
// @Description  Принимает токен обновления и возвращает новую пару токенов доступа и обновления.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refreshRequest  body      RefreshRequest    true  "Данные refresh-токена"
// @Success      200             {object}  dto.AuthResponse  "Токены обновлены"
// @Failure      400             {object}  dto.ErrorResponse "Отсутствует или некорректный refresh-токен"
// @Failure      401             {object}  dto.ErrorResponse "Невалидный или истекший refresh-токен"
// @Router       /api/v2/auth/refresh [post]
func (h *authHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid_request", utils.WithMessage("Refresh токен обязателен"))
		return
	}

	authRes, err := h.srv.RefreshTokens(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, "invalid_refresh_token", utils.WithMessage("Невалидный или истекший refresh токен"))
		return
	}

	c.JSON(http.StatusOK, authRes)
}

// Login godoc
// @Summary      Вход пользователя
// @Description  Проверяет адрес электронной почты и пароль, затем возвращает токены доступа и обновления.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      UserRequest       true  "Email и пароль"
// @Success      200   {object}  dto.AuthResponse  "Токены созданы"
// @Failure      400   {object}  dto.ErrorResponse "Некорректный JSON или ошибка валидации"
// @Failure      401   {object}  dto.ErrorResponse "Ошибка аутентификации"
// @Router       /api/v2/auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid_request", utils.WithMessage("Некорректный формат данных. Проверьте email и пароль"))
		return
	}
	authRes, err := h.srv.Login(req.Email, req.Password)

	if err != nil {
		utils.Unauthorized(c, "authentication_failed", utils.WithMessage(err.Error()))
		return
	}

	c.JSON(http.StatusOK, authRes)
}

// RequestPasswordReset godoc
// Неактивный устаревший маршрут: /api/users/request-reset [post]
// @Summary      Запросить сброс пароля
// @Description  Запускает устаревший сценарий сброса пароля и возвращает метаданные сессии.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      services.ResetPasswordRequest  true  "Email пользователя"
// @Success      200    {object}  models.SessionRegResponse
// @Failure      400    {object}  dto.ErrorResponse
func RequestPasswordReset(c *gin.Context) {
	var input services.ResetPasswordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := services.CreateSessionReset(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ConfirmPasswordReset godoc
// Неактивный устаревший маршрут: /api/users/confirm-reset [post]
// @Summary      Подтвердить сброс пароля
// @Description  Завершает устаревший сценарий сброса пароля.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      services.ConfirmResetPasswordInput  true  "Данные подтверждения сброса"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  dto.ErrorResponse
func ConfirmPasswordReset(c *gin.Context) {
	var input services.ConfirmResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.ResetPassword(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменён"})
}

// GettingUserId godoc
// Неактивный устаревший маршрут: /api/users/{us} [get]
// @Summary      Получить пользователя по us
// @Description  Исторический обработчик поиска пользователя оставлен только как закомментированная справка.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        us   path      string  true  "Значение поля us"
// @Success      200  {object}  models.User
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// func GettingUserId(c *gin.Context) {
// 	us := c.Param("us")
//
// 	user, err := services.FindUserByUs(us)
// 	if err != nil {
// 		c.JSON(404, gin.H{"error": "User not found"})
// 		return
// 	}
//
// 	c.JSON(200, user)
// }
