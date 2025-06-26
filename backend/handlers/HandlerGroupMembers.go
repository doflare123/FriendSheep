package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RemoveUserHandler godoc
// @Summary Удалить участника из группы
// @Description Удаляет участника из группы, если текущий пользователь — админ или оператор
// @Tags groups
// @Security BearerAuth
// @Param groupId path int true "ID группы"
// @Param userId path int true "ID пользователя, которого нужно удалить"
// @Produce json
// @Success 200 {object} map[string]string "Пользователь удалён из группы"
// @Failure 400 {object} map[string]string "Некорректные параметры"
// @Failure 403 {object} map[string]string "Нет прав на удаление"
// @Failure 404 {object} map[string]string "Пользователь не найден или не в группе"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /api/groups/{groupId}/members/{userId} [delete]
func RemoveUserHandler(c *gin.Context) {
	email := c.MustGet("email").(string)
	groupID, _ := strconv.ParseUint(c.Param("groupId"), 10, 64)
	userID, _ := strconv.ParseUint(c.Param("userId"), 10, 64)

	err := services.RemoveUserFromGroup(email, uint(groupID), uint(userID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удалён из группы"})
}

// LeaveGroupHandler godoc
// @Summary Покинуть группу
// @Description Удаляет текущего пользователя из группы. Админ не может покинуть группу, если он единственный админ.
// @Tags groups
// @Security BearerAuth
// @Param groupId path int true "ID группы"
// @Produce json
// @Success 200 {object} map[string]string "Вы покинули группу"
// @Failure 400 {object} map[string]string "Ошибка — пользователь не найден, не состоит в группе или единственный админ"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /api/groups/{groupId}/leave [delete]
func LeaveGroupHandler(c *gin.Context) {
	email := c.MustGet("email").(string)
	groupID, _ := strconv.ParseUint(c.Param("groupId"), 10, 64)

	err := services.LeaveGroup(email, uint(groupID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вы покинули группу"})
}
