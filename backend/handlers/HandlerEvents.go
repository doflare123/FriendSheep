package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"friendship/services/events"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type EventsHandler interface {
	CreateEvent(c *gin.Context)
	UpdateEvent(c *gin.Context)
	DeleteEvent(c *gin.Context)
	GetGroupEvents(c *gin.Context)
	GetEventDetails(c *gin.Context)
	GetEventDetailsForAdmin(c *gin.Context)
	JoinEvent(c *gin.Context)
	LeaveEvent(c *gin.Context)
	KickUserFromEvent(c *gin.Context)
	GetAllGenres(c *gin.Context)
	GetAllReferences(c *gin.Context)
}

type eventsHandler struct {
	srv events.EventsService
}

func NewEventsHandler(srv events.EventsService) EventsHandler {
	return &eventsHandler{
		srv: srv,
	}
}

// CreateEvent godoc
// @Summary      Создать событие
// @Description  Создает новое событие в группе. Доступно для админов и операторов
// @Tags         events_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body events.CreateEventInput true "Данные события"
// @Success      201 {object} dto.EventFullDto "Событие создано"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/admin/events [post]
func (h *eventsHandler) CreateEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input events.CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	eventDto, err := h.srv.CreateEvent(actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, events.ErrNotGroupMember):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в группе"})
		case errors.Is(err, events.ErrInvalidGenres):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, eventDto)
}

// UpdateEvent godoc
// @Summary      Обновить событие
// @Description  Обновляет информацию о событии. Доступно для админов и операторов
// @Tags         events_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Param        request body events.UpdateEventInput true "Данные для обновления"
// @Success      200 {object} dto.EventFullDto "Событие обновлено"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Событие не найдено"
// @Router       /api/v2/admin/events/{eventId} [put]
func (h *eventsHandler) UpdateEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	var input events.UpdateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	eventDto, err := h.srv.UpdateEvent(actorID, uint(eventID), input)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, events.ErrEventAlreadyStarted):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Событие уже началось, изменение невозможно"})
		case errors.Is(err, events.ErrInvalidGenres):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, eventDto)
}

// DeleteEvent godoc
// @Summary      Удалить событие
// @Description  Удаляет событие. Доступно для админов и операторов
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Success      200 {object} map[string]interface{} "Событие удалено"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Событие не найдено"
// @Router       /api/v2/admin/events/{eventId} [delete]
func (h *eventsHandler) DeleteEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	success, err := h.srv.DeleteEvent(actorID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Событие успешно удалено",
	})
}

// GetEventDetailsForAdmin godoc
// @Summary      Получить детали события для администратора
// @Description  Возвращает полную информацию о событии со списком всех участников. Доступно для админов и операторов группы
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Success      200 {object} dto.EventAdminDto "Детали события для админа"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Событие не найдено"
// @Router       /api/v2/admin/events/{eventId} [get]
func (h *eventsHandler) GetEventDetailsForAdmin(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	eventDto, err := h.srv.GetEventDetailsForAdmin(actorID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, events.ErrNotInGroup):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, eventDto)
}

// KickUserFromEvent godoc
// @Summary      Исключить участника из события
// @Description  Удаляет участника из события. Доступно для админов и операторов группы. Создателя исключить нельзя
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Param        userId path int true "ID пользователя"
// @Success      200 {object} map[string]interface{} "Пользователь исключен"
// @Failure      400 {object} map[string]string "Некорректные данные или нельзя исключить создателя"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Событие или пользователь не найден"
// @Router       /api/v2/admin/events/{eventId}/kick/{userId} [delete]
func (h *eventsHandler) KickUserFromEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	success, err := h.srv.KickUserFromEvent(actorID, uint(eventID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, events.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, events.ErrCreatorCantLeave):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя исключить создателя события"})
		case errors.Is(err, events.ErrNotJoined):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не участвует в событии"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь исключен из события",
	})
}

// GetGroupEvents godoc
// @Summary      Получить события группы
// @Description  Возвращает список всех событий группы. Доступно для админов и операторов
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {array} dto.EventShortDto "Список событий"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Router       /api/v2/groups/events/{groupId}/events [get]
func (h *eventsHandler) GetGroupEvents(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
		return
	}

	event, err := h.srv.GetGroupEvents(actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		case errors.Is(err, events.ErrNotGroupMember):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в группе"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetEventDetails godoc
// @Summary      Получить детали события
// @Description  Возвращает полную информацию о событии
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Success      200 {object} dto.EventFullDto "Детали события"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Недостаточно прав"
// @Failure      404 {object} map[string]string "Событие не найдено"
// @Router       /api/v2/events/{eventId} [get]
func (h *eventsHandler) GetEventDetails(c *gin.Context) {
	userID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	eventDto, err := h.srv.GetEventDetails(userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrNotGroupMember):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в группе этого события"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, eventDto)
}

// JoinEvent godoc
// @Summary      Присоединиться к событию
// @Description  Пользователь присоединяется к событию (только для участников группы)
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Success      200 {object} map[string]interface{} "Присоединение успешно"
// @Failure      400 {object} map[string]string "Событие заполнено или уже присоединились"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Не состоите в группе"
// @Failure      404 {object} map[string]string "Событие не найдено"
// @Router       /api/v2/events/{eventId}/join [post]
func (h *eventsHandler) JoinEvent(c *gin.Context) {
	userID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	success, err := h.srv.JoinEvent(userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrNotGroupMember):
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не состоите в группе"})
		case errors.Is(err, events.ErrEventFull):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Событие заполнено"})
		case errors.Is(err, events.ErrAlreadyJoined):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Вы уже присоединились к этому событию"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Вы успешно присоединились к событию",
	})
}

// LeaveEvent godoc
// @Summary      Покинуть событие
// @Description  Пользователь покидает событие (создатель не может покинуть)
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        eventId path int true "ID события"
// @Success      200 {object} map[string]interface{} "Успешно покинули событие"
// @Failure      400 {object} map[string]string "Создатель не может покинуть событие"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Вы не присоединялись к событию"
// @Router       /api/v2/events/{eventId}/leave [post]
func (h *eventsHandler) LeaveEvent(c *gin.Context) {
	userID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
		return
	}

	success, err := h.srv.LeaveEvent(userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
		case errors.Is(err, events.ErrCreatorCantLeave):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Создатель не может покинуть событие"})
		case errors.Is(err, events.ErrNotJoined):
			c.JSON(http.StatusNotFound, gin.H{"error": "Вы не присоединялись к этому событию"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Вы покинули событие",
	})
}

// GetAllGenres godoc
// @Summary      Получить все жанры
// @Description  Возвращает список всех доступных жанров событий
// @Tags         events
// @Produce      json
// @Success      200 {array} dto.ReferenceItemDto "Список жанров"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/v2/events/genres [get]
func (h *eventsHandler) GetAllGenres(c *gin.Context) {
	genres, err := h.srv.GetAllGenres()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, genres)
}

// GetAllReferences godoc
// @Summary      Получить все справочники
// @Description  Возвращает все справочники для событий и групп (типы, локации, возрастные ограничения, статусы, жанры, категории)
// @Tags         references
// @Produce      json
// @Success      200 {object} dto.ReferencesDto "Все справочники"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/v2/references [get]
func (h *eventsHandler) GetAllReferences(c *gin.Context) {
	references, err := h.srv.GetAllReferences()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, references)
}
