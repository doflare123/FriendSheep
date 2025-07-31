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
