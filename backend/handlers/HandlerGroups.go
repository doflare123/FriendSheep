package handlers

import (
	"errors"
	"net/http"
	"strconv"

	group "friendship/services/groups"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type GroupHandler interface {
	CreateGroup(c *gin.Context)
	UpdateGroup(c *gin.Context)
	DeleteGroup(c *gin.Context)
	JoinGroup(c *gin.Context)
	LeaveGroup(c *gin.Context)

	// Управление заявками
	ApproveAllJoinRequests(c *gin.Context)
	RejectAllJoinRequests(c *gin.Context)
	ApproveJoinRequest(c *gin.Context)
	RejectJoinRequest(c *gin.Context)
	GetJoinRequests(c *gin.Context)

	// Управление правами
	AddPermissions(c *gin.Context)
	RemovePermissions(c *gin.Context)
	GetGroupBlacklist(c *gin.Context)

	// Управление участниками
	DeleteUserFromGroup(c *gin.Context)
	RemoveFromBlacklist(c *gin.Context)

	// Приглашения
	CreateJoinInvite(c *gin.Context)
	AcceptJoinInvite(c *gin.Context)
	RejectJoinInvite(c *gin.Context)

	// История действий
	WatchRecentActions(c *gin.Context)
}

type groupHandler struct {
	srv group.GroupsService
}

func NewGroupHandler(srv group.GroupsService) GroupHandler {
	return &groupHandler{
		srv: srv,
	}
}

type GroupUpdateRequest struct {
	GroupID          uint    `json:"groupId" binding:"required"`
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	SmallDescription *string `json:"smallDescription"`
	Image            *string `json:"image"`
	IsPrivate        *bool   `json:"isPrivate"`
	City             *string `json:"city"`
	Categories       []*uint `json:"categories"`
	Contacts         *string `json:"contacts"`
}

// CreateGroup godoc
// @Summary      Создание группы
// @Description  Создает новую группу. Контакты передаются строкой в формате "название:ссылка, название:ссылка"
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.CreateGroupInput true "Данные для создания группы"
// @Success      201 {object} dto.GroupDto "Группа успешно создана"
// @Failure      400 {object} map[string]interface{} "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/v2/groups [post]
func (h *groupHandler) CreateGroup(c *gin.Context) {
	idValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Не найден userID в контексте",
		})
		return
	}
	id := idValue.(uint)

	var input group.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	if input.Image == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Изображение группы обязательно",
		})
		return
	}

	groupDto, err := h.srv.CreateGroup(id, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь не найден",
			})
		case errors.Is(err, group.ErrCategoriesNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некоторые категории не найдены",
			})
		case errors.Is(err, group.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, group.ErrGroupCreation):
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Ошибка при создании группы",
				"details": err.Error(),
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, groupDto)
}

// UpdateGroup godoc
// @Summary      Обновление группы
// @Description  Обновляет информацию о группе. Все поля кроме groupId опциональны
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body GroupUpdateRequest true "Данные для обновления группы"
// @Success      200 {object} dto.GroupDto "Группа успешно обновлена"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Группа не найдена"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/v2/groups [put]
func (h *groupHandler) UpdateGroup(c *gin.Context) {
	actorID := c.GetUint("userID")

	var request GroupUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Некорректные данные",
			"details": err.Error(),
		})
		return
	}

	// Преобразуем request в input для сервиса
	input := group.GroupUpdateInput{
		GroupID:          request.GroupID,
		Name:             request.Name,
		Description:      request.Description,
		SmallDescription: request.SmallDescription,
		Image:            request.Image,
		IsPrivate:        request.IsPrivate,
		City:             request.City,
		Categories:       request.Categories,
		Contacts:         request.Contacts,
	}

	groupDto, err := h.srv.UpdateGroup(actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Группа не найдена"})
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, groupDto)
}

// DeleteGroup godoc
// @Summary      Удаление группы
// @Description  Удаляет группу. Доступно только для админа группы
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Группа успешно удалена"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Группа не найдена"
// @Router       /api/v2/groups/{groupId} [delete]
func (h *groupHandler) DeleteGroup(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	success, err := h.srv.DeleteGroup(actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Группа не найдена"})
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Только админ может удалить группу"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Группа успешно удалена",
	})
}

// JoinGroup godoc
// @Summary      Вступление в группу
// @Description  Для открытых групп - сразу добавляет пользователя. Для приватных - создает заявку на вступление
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} group.GroupResult "Результат вступления в группу"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Группа не найдена"
// @Failure      409 {object} map[string]string "Уже в группе или заявка уже отправлена"
// @Router       /api/v2/groups/{groupId}/join [post]
func (h *groupHandler) JoinGroup(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	result, err := h.srv.JoinGroup(userID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Группа не найдена"})
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, group.ErrAlreadyInGroup):
			c.JSON(http.StatusConflict, gin.H{"error": "Вы уже состоите в этой группе"})
		case errors.Is(err, group.ErrRequestAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Ваша заявка уже отправлена"})
		case errors.Is(err, group.ErrUserInBlacklist):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы в черном списке этой группы"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// LeaveGroup godoc
// @Summary      Выход из группы
// @Description  Пользователь покидает группу. Админ не может покинуть группу
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Вы покинули группу"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Админ не может покинуть группу"
// @Failure      404 {object} map[string]string "Вы не состоите в группе"
// @Router       /api/v2/groups/{groupId}/leave [post]
func (h *groupHandler) LeaveGroup(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	success, err := h.srv.LeaveGroup(userID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusNotFound, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Вы успешно покинули группу",
	})
}

// ApproveAllJoinRequests godoc
// @Summary      Одобрить все заявки на вступление
// @Description  Одобряет все ожидающие заявки на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Количество одобренных заявок"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/requests/approve-all [post]
func (h *groupHandler) ApproveAllJoinRequests(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	count, err := h.srv.ApproveAllJoinRequests(actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Заявки одобрены",
		"count":   count,
	})
}

// RejectAllJoinRequests godoc
// @Summary      Отклонить все заявки на вступление
// @Description  Отклоняет все ожидающие заявки на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Количество отклоненных заявок"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/requests/reject-all [post]
func (h *groupHandler) RejectAllJoinRequests(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	count, err := h.srv.RejectAllJoinRequests(actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Заявки отклонены",
		"count":   count,
	})
}

// ApproveJoinRequest godoc
// @Summary      Одобрить заявку на вступление
// @Description  Одобряет конкретную заявку на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        requestId path int true "ID заявки"
// @Success      200 {object} map[string]interface{} "Заявка одобрена"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Заявка не найдена"
// @Router       /api/v2/groups/requests/{requestId}/approve [post]
func (h *groupHandler) ApproveJoinRequest(c *gin.Context) {
	actorID := c.GetUint("userID")

	requestIDStr := c.Param("requestId")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заявки"})
		return
	}

	success, err := h.srv.ApproveJoinRequest(actorID, uint(requestID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrUserInBlacklist):
			c.JSON(http.StatusForbidden, gin.H{"error": "Пользователь в черном списке"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Заявка одобрена",
	})
}

// RejectJoinRequest godoc
// @Summary      Отклонить заявку на вступление
// @Description  Отклоняет конкретную заявку на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        requestId path int true "ID заявки"
// @Success      200 {object} map[string]interface{} "Заявка отклонена"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Заявка не найдена"
// @Router       /api/v2/groups/requests/{requestId}/reject [post]
func (h *groupHandler) RejectJoinRequest(c *gin.Context) {
	actorID := c.GetUint("userID")

	requestIDStr := c.Param("requestId")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заявки"})
		return
	}

	success, err := h.srv.RejectJoinRequest(actorID, uint(requestID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Заявка отклонена",
	})
}

// AddPermissions godoc
// @Summary      Назначить оператора
// @Description  Назначает пользователя оператором группы. Доступно только для админа
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.PermissionInput true "ID группы и пользователя"
// @Success      200 {object} map[string]interface{} "Права успешно назначены"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Router       /api/v2/groups/permissions/add [post]
func (h *groupHandler) AddPermissions(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input group.PermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	success, err := h.srv.AddPermissions(actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Только админ может назначать операторов"})
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь не состоит в группе"})
		case errors.Is(err, group.ErrCannotChangeOwnRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя изменить собственную роль"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь назначен оператором группы",
	})
}

// RemovePermissions godoc
// @Summary      Снять права оператора
// @Description  Снимает с пользователя права оператора группы. Доступно только для админа
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.PermissionInput true "ID группы и пользователя"
// @Success      200 {object} map[string]interface{} "Права успешно сняты"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Router       /api/v2/groups/permissions/remove [post]
func (h *groupHandler) RemovePermissions(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input group.PermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	success, err := h.srv.RemovePermissions(actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Только админ может снимать права оператора"})
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь не состоит в группе"})
		case errors.Is(err, group.ErrCannotChangeOwnRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя изменить собственную роль"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Права оператора сняты",
	})
}

// DeleteUserFromGroup godoc
// @Summary      Удалить пользователя из группы
// @Description  Удаляет пользователя из группы и добавляет в черный список. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        userId path int true "ID пользователя"
// @Success      200 {object} map[string]interface{} "Пользователь удален из группы"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Router       /api/v2/groups/{groupId}/members/{userId} [delete]
func (h *groupHandler) DeleteUserFromGroup(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	success, err := h.srv.DeleteUserFromGroup(actorID, uint(groupID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не состоит в группе"})
		case errors.Is(err, group.ErrCannotRemoveSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя удалить самого себя"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь удален из группы и добавлен в черный список",
	})
}

// RemoveFromBlacklist godoc
// @Summary      Убрать из черного списка
// @Description  Удаляет пользователя из черного списка группы. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        userId path int true "ID пользователя"
// @Success      200 {object} map[string]interface{} "Пользователь удален из черного списка"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Пользователь не найден в черном списке"
// @Router       /api/v2/groups/{groupId}/blacklist/{userId} [delete]
func (h *groupHandler) RemoveFromBlacklist(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	success, err := h.srv.RemoveFromBlacklist(actorID, uint(groupID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь удален из черного списка",
	})
}

// CreateJoinInvite godoc
// @Summary      Создать приглашение в группу
// @Description  Отправляет приглашение пользователю вступить в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.JoinInviteInput true "ID группы и пользователя"
// @Success      200 {object} map[string]interface{} "Приглашение отправлено"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Failure      409 {object} map[string]string "Приглашение уже существует или пользователь уже в группе"
// @Router       /api/v2/groups/invites [post]
func (h *groupHandler) CreateJoinInvite(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input group.JoinInviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	success, err := h.srv.CreateJoinInvite(actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, group.ErrAlreadyInGroup):
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже в группе"})
		case errors.Is(err, group.ErrInviteAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Приглашение уже отправлено"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Приглашение отправлено",
	})
}

// AcceptJoinInvite godoc
// @Summary      Принять приглашение в группу
// @Description  Принимает приглашение и добавляет пользователя в группу
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения"
// @Success      200 {object} group.GroupResult "Приглашение принято"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Пользователь в черном списке"
// @Failure      404 {object} map[string]string "Приглашение не найдено"
// @Router       /api/v2/groups/invites/{inviteId}/accept [post]
func (h *groupHandler) AcceptJoinInvite(c *gin.Context) {
	userID := c.GetUint("userID")

	inviteIDStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(inviteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID приглашения"})
		return
	}

	result, err := h.srv.AcceptJoinInvite(userID, uint(inviteID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrUserInBlacklist):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы в черном списке этой группы"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RejectJoinInvite godoc
// @Summary      Отклонить приглашение в группу
// @Description  Отклоняет приглашение вступить в группу
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения"
// @Success      200 {object} map[string]interface{} "Приглашение отклонено"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Приглашение не найдено"
// @Router       /api/v2/groups/invites/{inviteId}/reject [post]
func (h *groupHandler) RejectJoinInvite(c *gin.Context) {
	userID := c.GetUint("userID")

	inviteIDStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(inviteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID приглашения"})
		return
	}

	success, err := h.srv.RejectJoinInvite(userID, uint(inviteID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Приглашение отклонено",
	})
}

// WatchRecentActions godoc
// @Summary      Получить историю действий в группе
// @Description  Возвращает список последних действий операторов и админов в группе
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        limit query int false "Количество записей (по умолчанию 50, максимум 100)"
// @Success      200 {array} group.GroupAction "История действий"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/actions [get]
func (h *groupHandler) WatchRecentActions(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	actions, err := h.srv.WatchRecentActions(userID, uint(groupID), limit)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, actions)
}

// GetGroupBlacklist godoc
// @Summary      Получить черный список группы
// @Description  Возвращает список забаненных пользователей. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        limit query int false "Количество записей (по умолчанию 50, максимум 100)"
// @Success      200 {array} group.BlacklistUser "Черный список"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/blacklist [get]
func (h *groupHandler) GetGroupBlacklist(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	blacklist, err := h.srv.GetGroupBlacklist(actorID, uint(groupID), limit)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, blacklist)
}

// GetJoinRequests godoc
// @Summary      Получить заявки на вступление
// @Description  Возвращает список заявок на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        status query string false "Статус заявки (pending, approved, rejected)"
// @Param        limit query int false "Количество записей (по умолчанию 50, максимум 100)"
// @Success      200 {array} group.JoinRequestInfo "Список заявок"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/requests [get]
func (h *groupHandler) GetJoinRequests(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	requests, err := h.srv.GetJoinRequests(actorID, uint(groupID), status, limit)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, group.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в этой группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, requests)
}
