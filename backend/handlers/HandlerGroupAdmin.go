package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"
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

	groups, err := services.GetAdminGroups(&email)
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

// GetInfAdminGroup godoc
// @Summary      Получить детальную информацию о группе для администратора
// @Description  Возвращает детальную информацию о группе, созданной пользователем, включая заявки на вступление и сессии в статусе 'Набор'. Доступно только создателю группы.
// @Tags         groups_admin
// @Security     BearerAuth
// @Produce      json
// @Param        groupId path int true "ID группы"
// @Success      200  {object}  services.AdminGroupInfResponse "Детальная информация о группе"
// @Failure      400  {object}  map[string]string "Некорректный ID группы"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      403  {object}  map[string]string "Доступ запрещен (пользователь не является создателем)"
// @Failure      404  {object}  map[string]string "Группа не найдена"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/admin/groups/{groupId}/infGroup [get]
func GetInfAdminGroup(c *gin.Context) {
	email := c.MustGet("email").(string)

	groupIDStr := c.Param("groupId")
	groupID64, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil || groupID64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID группы"})
		return
	}
	groupID := uint(groupID64)
	group, err := services.GetAdminGroupInfo(&email, &groupID)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "invalid group ID") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
		} else if strings.Contains(errStr, "group not found") || strings.Contains(errStr, "пользователь не найден") {
			c.JSON(http.StatusNotFound, gin.H{"error": errStr})
		} else if strings.Contains(errStr, "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": errStr})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errStr})
		}
		return
	}
	c.JSON(http.StatusOK, group)
}
