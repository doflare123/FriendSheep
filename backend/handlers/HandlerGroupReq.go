package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetJoinRequestsResponseDoc struct {
	Requests []GroupJoinRequestDoc `json:"requests"`
}

type GroupJoinRequestDoc struct {
	ID      uint            `json:"id" example:"3"`
	UserID  uint            `json:"userId" example:"42"`
	GroupID uint            `json:"groupId" example:"5"`
	Status  string          `json:"status" example:"pending"`
	User    UserPreviewDoc  `json:"user"`
	Group   GroupPreviewDoc `json:"group"`
}

type UserPreviewDoc struct {
	ID    uint   `json:"id" example:"42"`
	Name  string `json:"name" example:"Иван"`
	Email string `json:"email" example:"ivan@example.com"`
	Image string `json:"image" example:"https://img.com/avatar.png"`
}

type GroupPreviewDoc struct {
	ID    uint   `json:"id" example:"5"`
	Name  string `json:"name" example:"Закрытая группа"`
	Image string `json:"image" example:"https://img.com/group.png"`
}

// GetJoinRequests godoc
// @Summary Получить заявки на вступление
// @Description Получение всех заявок на вступление в закрытые группы, где текущий пользователь является администратором
// @Tags groups
// @Security BearerAuth
// @Param groupId path int true "ID группы"
// @Accept       json
// @Produce json
// @Success 200 {object} GetJoinRequestsResponseDoc "Список заявок"
// @Failure 401 {object} map[string]string "Ошибка авторизации"
// @Failure 403 {object} map[string]string "Нет доступа или пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /api/groups/requests/{groupId} [get]
func GetJoinRequests(c *gin.Context) {
	email := c.MustGet("email").(string)

	groupIDParam := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID группы"})
		return
	}

	requests, err := services.GetPendingJoinRequestsForAdmin(email, uint(groupID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// ApproveJoinRequest godoc
// @Summary Одобрить заявку
// @Description Одобрение заявки на вступление в закрытую группу. Только админ может это сделать.
// @Tags groups_admin
// @Security BearerAuth
// @Param requestId path int true "ID заявки"
// @Produce json
// @Success 200 {object} map[string]string "Заявка одобрена"
// @Failure 400 {object} map[string]string "Ошибка: заявка не найдена или уже обработана"
// @Failure 401 {object} map[string]string "Ошибка авторизации"
// @Failure 403 {object} map[string]string "Вы не админ"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/admin/groups/requests/{requestId}/approve [post]
func ApproveJoinRequest(c *gin.Context) {
	email, _ := c.Get("email")
	requestId := c.Param("requestId")

	if err := services.ApproveJoinRequestByID(email.(string), requestId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка одобрена, пользователь добавлен в группу"})
}

// RejectJoinRequest godoc
// @Summary Отклонить заявку
// @Description Отклонение заявки на вступление в закрытую группу. Только админ может это сделать.
// @Tags groups_admin
// @Security BearerAuth
// @Param requestId path int true "ID заявки"
// @Produce json
// @Success 200 {object} map[string]string "Заявка отклонена"
// @Failure 400 {object} map[string]string "Ошибка: заявка не найдена или уже обработана"
// @Failure 401 {object} map[string]string "Ошибка авторизации"
// @Failure 403 {object} map[string]string "Вы не админ"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/admin/groups/requests/{requestId}/reject [post]
func RejectJoinRequest(c *gin.Context) {
	email, _ := c.Get("email")
	requestId := c.Param("requestId")

	if err := services.RejectJoinRequestByID(email.(string), requestId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка отклонена"})
}
