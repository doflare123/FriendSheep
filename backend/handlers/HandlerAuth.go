package handlers

import (
	"friendship/services"
	"friendship/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

// AuthUser godoc
// @Summary      Аутентификация пользователя
// @Description  Проверяет email и пароль, возвращает access и refresh токены
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      UserRequest  true  "Данные пользователя"
// @Success      200   {object}  AuthResponse  "Токены успешно созданы"
// @Failure      400   {object}  map[string]string  "Некорректный JSON или параметры"
// @Failure      401   {object}  map[string]string  "Неверный пароль"
// @Failure      404   {object}  map[string]string  "Пользователь не найден"
// @Failure      500   {object}  map[string]string  "Ошибка сервера"
// @Router       /api/users/login [post]
func AuthUser(c *gin.Context) {
	var input UserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON"})
		return
	}

	user, err := services.FindUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	if !utils.ComparePasswords(input.Password, user.Password, user.Salt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	token, err := utils.GenerateTokenPair(user.Email, user.Name, user.Us, user.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	adminGroups, err := services.GetAdminGroups(&user.Email)
	if err != nil {
		adminGroups = []services.AdminGroupResponse{}
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		AdminGroups:  adminGroups,
	})
}

// RefreshTokenHandler godoc
// @Summary      Обновление токенов
// @Description  По refresh токену выдает новые access и refresh токены
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refreshRequest  body      RefreshRequest  true  "Refresh токен"
// @Success      200             {object}  map[string]string  "Новые токены успешно созданы"
// @Failure      400             {object}  map[string]string  "Отсутствует или неверный refresh_token"
// @Failure      401             {object}  map[string]string  "Невалидный или просроченный refresh token"
// @Router       /api/users/refresh [post]
func RefreshTokenHandler(c *gin.Context) {
	var req RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Отсутствует refresh_token в запросе"})
		return
	}

	newTokens, err := utils.RefreshTokens(req.RefreshToken, services.FindUserByEmail)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный или просроченный refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newTokens.AccessToken,
		"refresh_token": newTokens.RefreshToken,
	})
}

// RequestPasswordReset godoc
// @Summary      Запрос на сброс пароля
// @Description  Пользователь указывает email, на него отправляется код подтверждения для смены пароля.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  services.ResetPasswordRequest  true  "Email пользователя"
// @Success      200    {object} models.SessionRegResponse
// @Failure      400    {object} map[string]string
// @Router       /api/users/request-reset [post]
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
// @Summary      Подтверждение сброса пароля
// @Description  Пользователь вводит session_id, код из email и новый пароль. При успешной верификации пароль меняется.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  services.ConfirmResetPasswordInput  true  "Данные для подтверждения и новый пароль"
// @Success      200    {object} map[string]string
// @Failure      400    {object} map[string]string
// @Router       /api/users/confirm-reset [post]
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
