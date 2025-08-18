package handlers

import (
	"friendship/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetGroupsUserSub godoc
// @Summary Получить подписки пользователя на группы
// @Description Возвращает список групп, на которые подписан пользователь, исключая группы, созданные им.
// @Tags Users inf
// @Security BearerAuth
// @Produce json
// @Success 200 {array} services.GroupResponse "Список групп, на которые подписан пользователь"
// @Failure 400 {object} map[string]string "Не передан jwt"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/subscriptions [get]
func GetGroupsUserSub(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	groups, err := services.GetGroupsUserSub(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func SubscriptionsTelegramNotify(c *gin.Context) {
	var input services.TelegramNotifyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := input.SubscriptionsTelegramNotify()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "details": "Привязка телеграм аккаунта не удалась"})
	}

}

func UnSubscriptionsTelegramNotify(c *gin.Context) {
	var input services.TelegramNotifyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := input.UnSubscriptionsTelegramNotify()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "details": "Отвязка телеграм аккаунта не удалась"})
	}

}

// GetInfAboutUser godoc
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полную информацию о текущем авторизованном пользователе, включая его сессии и статистику.
// @Tags         Users inf
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  services.InformationAboutUser "Полная информация о пользователе"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      404  {object}  map[string]string "Пользователь не найден"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/users/inf [get]
func GetInfAboutUser(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}

	email, ok := emailValue.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email в контексте"})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email не может быть пустым"})
		return
	}

	userInf, err := services.GetInfAboutUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if userInf == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, userInf)
}
