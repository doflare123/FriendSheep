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

// AuthUser godoc
// @Summary      Аутентификация пользователя
// @Description  Проверяет email и пароль, возвращает access и refresh токены
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      UserRequest  true  "Данные пользователя"
// @Success      200   {object}  map[string]string  "Токены успешно созданы"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
		return
	}

	token, err := utils.GenerateTokenPair(user.Email, user.Us)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
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

	newTokens, err := utils.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный или просроченный refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newTokens.AccessToken,
		"refresh_token": newTokens.RefreshToken,
	})
}
