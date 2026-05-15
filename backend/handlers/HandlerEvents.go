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
// @Description  Справочные значения берите из GET /api/v2/references: eventTypeId -> eventTypes[].id, locationId -> locations[].id, ageLimit -> ageLimits[].id. Поле status передавать не нужно, при создании ставится "Набор".
// @Description  Создает новое событие в группе. Доступно модераторам и администраторам группы.
// @Tags         events_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      events.CreateEventInput  true  "Данные события"
// @Success      201      {object}  dto.EventFullDto         "Событие создано"
// @Failure      400      {object}  dto.ErrorResponse        "Ошибка валидации или справочников"
// @Failure      401      {object}  dto.ErrorResponse        "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse        "Недостаточно прав"
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
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
		case errors.Is(err, events.ErrInvalidGenres):
			utils.BadRequest(c, err.Error())
		case errors.Is(err, events.ErrAgeLimitNotFound):
			utils.BadRequest(c, "Возрастное ограничение не найдено")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, eventDto)
}

// UpdateEvent godoc
// @Summary      Обновить событие
// @Description  Обновляет событие. Доступно модераторам и администраторам группы.
// @Tags         events_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int                      true  "ID события"
// @Param        request  body      events.UpdateEventInput  true  "Данные для обновления"
// @Success      200      {object}  dto.EventFullDto         "Событие обновлено"
// @Failure      400      {object}  dto.ErrorResponse        "Ошибка валидации или состояния"
// @Failure      401      {object}  dto.ErrorResponse        "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse        "Недостаточно прав"
// @Failure      404      {object}  dto.ErrorResponse        "Событие не найдено"
// @Router       /api/v2/admin/events/{eventId} [put]
func (h *eventsHandler) UpdateEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
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
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrEventAlreadyStarted):
			utils.BadRequest(c, "Событие уже началось, изменение невозможно")
		case errors.Is(err, events.ErrInvalidGenres):
			utils.BadRequest(c, err.Error())
		case errors.Is(err, events.ErrAgeLimitNotFound):
			utils.BadRequest(c, "Возрастное ограничение не найдено")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, eventDto)
}

// DeleteEvent godoc
// @Summary      Удалить событие
// @Description  Удаляет событие. Доступно модераторам и администраторам группы.
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int               true  "ID события"
// @Success      200      {object}  map[string]any    "Результат удаления"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse "Недостаточно прав"
// @Failure      404      {object}  dto.ErrorResponse "Событие не найдено"
// @Router       /api/v2/admin/events/{eventId} [delete]
func (h *eventsHandler) DeleteEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
		return
	}

	success, err := h.srv.DeleteEvent(actorID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Событие успешно удалено",
	})
}

// GetEventDetailsForAdmin godoc
// @Summary      Получить событие для администратора
// @Description  Возвращает детали события со списком участников для модераторов и администраторов группы.
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int               true  "ID события"
// @Success      200      {object}  dto.EventAdminDto "Детали события для администратора"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse "Недостаточно прав"
// @Failure      404      {object}  dto.ErrorResponse "Событие не найдено"
// @Router       /api/v2/admin/events/{eventId} [get]
func (h *eventsHandler) GetEventDetailsForAdmin(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
		return
	}

	eventDto, err := h.srv.GetEventDetailsForAdmin(actorID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, eventDto)
}

// KickUserFromEvent godoc
// @Summary      Удалить участника события
// @Description  Удаляет участника из события. Доступно модераторам и администраторам группы, кроме удаления создателя события.
// @Tags         events_admin
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int               true  "ID события"
// @Param        userId   path      int               true  "ID пользователя"
// @Success      200      {object}  map[string]any    "Участник удален"
// @Failure      400      {object}  dto.ErrorResponse "Некорректные ID или попытка удалить создателя"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse "Недостаточно прав"
// @Failure      404      {object}  dto.ErrorResponse "Событие или пользователь не найдены"
// @Router       /api/v2/admin/events/{eventId}/kick/{userId} [delete]
func (h *eventsHandler) KickUserFromEvent(c *gin.Context) {
	actorID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID пользователя")
		return
	}

	success, err := h.srv.KickUserFromEvent(actorID, uint(eventID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrCreatorCantLeave):
			utils.BadRequest(c, "Нельзя исключить создателя события")
		case errors.Is(err, events.ErrNotJoined):
			utils.NotFound(c, "Пользователь не участвует в событии")
		default:
			utils.InternalError(c, err.Error())
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
// @Description  Возвращает список событий группы. Доступно любому участнику группы.
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        groupId  path      int               true  "ID группы"
// @Success      200      {array}   dto.EventShortDto "События группы"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse "Требуется участие в группе"
// @Router       /api/v2/groups/events/{groupId}/events [get]
func (h *eventsHandler) GetGroupEvents(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	event, err := h.srv.GetGroupEvents(actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetEventDetails godoc
// @Summary      Получить детали события
// @Description  Возвращает детали события для участника группы.
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int               true  "ID события"
// @Success      200      {object}  dto.EventFullDto  "Детали события"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse "Требуется участие в группе"
// @Failure      404      {object}  dto.ErrorResponse "Событие не найдено"
// @Router       /api/v2/events/{eventId} [get]
func (h *eventsHandler) GetEventDetails(c *gin.Context) {
	userID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
		return
	}

	eventDto, err := h.srv.GetEventDetails(userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе этого события")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, eventDto)
}

// JoinEvent godoc
// @Summary      Присоединиться к событию
// @Description  Добавляет текущего пользователя в событие. Доступно только участникам группы.
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int               true  "ID события"
// @Success      200      {object}  map[string]any    "Результат вступления"
// @Failure      400      {object}  dto.ErrorResponse "Событие заполнено или пользователь уже участвует"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      403      {object}  dto.ErrorResponse "Требуется участие в группе"
// @Failure      404      {object}  dto.ErrorResponse "Событие не найдено"
// @Router       /api/v2/events/{eventId}/join [post]
func (h *eventsHandler) JoinEvent(c *gin.Context) {
	userID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
		return
	}

	success, err := h.srv.JoinEvent(userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
		case errors.Is(err, events.ErrEventFull):
			utils.BadRequest(c, "Событие заполнено")
		case errors.Is(err, events.ErrAlreadyJoined):
			utils.BadRequest(c, "Вы уже присоединились к этому событию")
		default:
			utils.InternalError(c, err.Error())
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
// @Description  Удаляет текущего пользователя из события. Создатель события не может его покинуть.
// @Tags         events
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  path      int               true  "ID события"
// @Success      200      {object}  map[string]any    "Результат выхода"
// @Failure      400      {object}  dto.ErrorResponse "Создатель не может покинуть событие"
// @Failure      401      {object}  dto.ErrorResponse "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse "Участие не найдено"
// @Router       /api/v2/events/{eventId}/leave [post]
func (h *eventsHandler) LeaveEvent(c *gin.Context) {
	userID := c.GetUint("userID")

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID события")
		return
	}

	success, err := h.srv.LeaveEvent(userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrCreatorCantLeave):
			utils.BadRequest(c, "Создатель не может покинуть событие")
		case errors.Is(err, events.ErrNotJoined):
			utils.NotFound(c, "Вы не присоединялись к этому событию")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Вы покинули событие",
	})
}

// GetAllGenres godoc
// @Summary      Получить жанры
// @Description  Возвращает все жанры событий.
// @Tags         events
// @Produce      json
// @Success      200 {array} dto.ReferenceItemDto "Список жанров"
// @Failure      500 {object} dto.ErrorResponse    "Внутренняя ошибка сервера"
// @Router       /api/v2/events/genres [get]
func (h *eventsHandler) GetAllGenres(c *gin.Context) {
	genres, err := h.srv.GetAllGenres()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, genres)
}

// GetAllReferences godoc
// @Summary      Получить справочники
// @Description  Возвращает все справочники, используемые группами и событиями.
// @Tags         references
// @Produce      json
// @Success      200 {object} dto.ReferencesDto "Справочники"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/references [get]
func (h *eventsHandler) GetAllReferences(c *gin.Context) {
	references, err := h.srv.GetAllReferences()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, references)
}
