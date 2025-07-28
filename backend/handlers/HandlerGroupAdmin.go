package handlers

import (
	"friendship/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAdminGroups godoc
// @Summary      Получить группы, созданные пользователем
// @Description  Возвращает список всех групп, где текущий авторизованный пользователь является создателем. В ответе также будет количество участников.
// @Tags         groups_admin
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   services.AdminGroupResponse "Список групп администратора"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      404  {object}  map[string]string "Пользователь не найден"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/admin/groups [get]
func GetAdminGroups(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}

	groups, err := services.GetAdminGroups(email)
	if err != nil {
		if strings.Contains(err.Error(), "пользователь не найден") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить группы"})
		return
	}
	c.JSON(http.StatusOK, groups)
}
