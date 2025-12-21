package handlers

import (
	"friendship/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterDeviceToken godoc
// @Summary Регистрация FCM токена устройства
// @Description Регистрирует или обновляет FCM токен для получения push-уведомлений
// @Tags Device Tokens
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.RegisterDeviceTokenRequest true "Данные токена устройства"
// @Success 200 {object} services.DeviceTokenResponse "Токен успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/device-tokens/register [post]
func RegisterDeviceToken(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	user, err := services.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
		return
	}

	var req services.RegisterDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные: " + err.Error()})
		return
	}

	response, err := services.RegisterOrUpdateDeviceToken(user.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "токен успешно зарегистрирован",
		"data":    response,
	})
}

// GetMyDevices godoc
// @Summary Получить список устройств пользователя
// @Description Возвращает все зарегистрированные устройства текущего пользователя
// @Tags Device Tokens
// @Security BearerAuth
// @Produce json
// @Success 200 {array} services.DeviceTokenResponse "Список устройств"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/device-tokens [get]
func GetMyDevices(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	user, err := services.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
		return
	}

	devices, err := services.GetUserDevices(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// DeactivateDeviceToken godoc
// @Summary Деактивировать токен устройства
// @Description Помечает токен как неактивный (перестанут приходить уведомления)
// @Tags Device Tokens
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Токен устройства" example({"device_token": "fcm_token_here"})
// @Success 200 {object} map[string]string "Токен деактивирован"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 404 {object} map[string]string "Токен не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/device-tokens/deactivate [post]
func DeactivateDeviceToken(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	user, err := services.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
		return
	}

	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные"})
		return
	}

	if err := services.DeactivateDeviceToken(user.ID, req.DeviceToken); err != nil {
		if err.Error() == "токен не найден" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "токен успешно деактивирован"})
}

// DeleteDeviceToken godoc
// @Summary Удалить токен устройства
// @Description Полностью удаляет токен из системы
// @Tags Device Tokens
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param device_token query string true "Токен устройства"
// @Success 200 {object} map[string]string "Токен удален"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 404 {object} map[string]string "Токен не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/device-tokens [delete]
func DeleteDeviceToken(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	user, err := services.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
		return
	}

	deviceToken := c.Query("device_token")
	if deviceToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не указан device_token"})
		return
	}

	if err := services.DeleteDeviceToken(user.ID, deviceToken); err != nil {
		if err.Error() == "токен не найден" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "токен успешно удален"})
}
