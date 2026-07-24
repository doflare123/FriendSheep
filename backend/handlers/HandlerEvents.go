package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"friendship/services/events"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type EventsHandler interface {
	CreateEvent(c *gin.Context)
	UpdateEvent(c *gin.Context)
	DeleteEvent(c *gin.Context)
	SearchEvents(c *gin.Context)
	GetGroupEvents(c *gin.Context)
	GetEventDetails(c *gin.Context)
	GetEventDetailsForAdmin(c *gin.Context)
	JoinEvent(c *gin.Context)
	LeaveEvent(c *gin.Context)
	KickUserFromEvent(c *gin.Context)
}

type eventsHandler struct {
	membership events.EventMembershipService
	commands   events.EventCommandService
	reads      events.EventReadService
	admin      events.EventAdminService
}

type EventsHandlerDependencies struct {
	Membership events.EventMembershipService
	Commands   events.EventCommandService
	Reads      events.EventReadService
	Admin      events.EventAdminService
}

func NewEventsHandler(dependencies EventsHandlerDependencies) EventsHandler {
	return &eventsHandler{
		membership: dependencies.Membership,
		commands:   dependencies.Commands,
		reads:      dependencies.Reads,
		admin:      dependencies.Admin,
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

	eventDto, err := h.commands.CreateEvent(c.Request.Context(), actorID, input)
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

	eventDto, err := h.commands.UpdateEvent(c.Request.Context(), actorID, uint(eventID), input)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
		case errors.Is(err, events.ErrEventAlreadyStarted):
			utils.BadRequest(c, "Событие уже началось, изменение невозможно")
		case errors.Is(err, events.ErrInvalidGenres):
			utils.BadRequest(c, err.Error())
		case errors.Is(err, events.ErrAgeLimitNotFound):
			utils.BadRequest(c, "Возрастное ограничение не найдено")
		case errors.Is(err, events.ErrMaxUsersBelowCurrent):
			utils.BadRequest(c, err.Error())
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

	success, err := h.commands.DeleteEvent(c.Request.Context(), actorID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
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

	eventDto, err := h.admin.GetEventDetailsForAdmin(c.Request.Context(), actorID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotGroupMember):
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

	success, err := h.admin.KickUserFromEvent(c.Request.Context(), actorID, uint(eventID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, events.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
		case errors.Is(err, events.ErrCreatorCantLeave):
			utils.BadRequest(c, "Нельзя исключить создателя события")
		case errors.Is(err, events.ErrActorCantKickSelf):
			utils.BadRequest(c, "Используйте выход из события")
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

	event, err := h.reads.GetGroupEvents(c.Request.Context(), actorID, uint(groupID))
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

// SearchEvents godoc
// @Summary      Поиск событий
// @Description  Публичный поиск событий с фильтрами, исключениями и пагинацией. События приватных групп видны только участникам этих групп. Если передать JWT, дополнительно доступны события приватных групп пользователя и режим feed=subscription_news.
// @Description  Поддерживаются алиасы query: q/query, categoryIds/categoryId/categories, genreIds/genreId/genres, pageSize/limit, dateFrom/from, dateTo/to, locationType/locationTypes/type, excludeLocationType/excludeLocationTypes/excludeType, feed=subscription_news/subscription-news/новинки подписок.
// @Tags         events
// @Produce      json
// @Param        q query string false "Поиск по названию/описанию события и названию группы"
// @Param        groupId query int false "ID группы"
// @Param        categoryIds query []int false "ID категорий групп, повтором или CSV"
// @Param        excludeCategoryIds query []int false "ID категорий групп для исключения, повтором или CSV"
// @Param        genreIds query []int false "ID жанров, повтором или CSV"
// @Param        excludeGenreIds query []int false "ID жанров для исключения, повтором или CSV"
// @Param        eventTypeIds query []int false "ID типов событий, повтором или CSV"
// @Param        excludeEventTypeIds query []int false "ID типов событий для исключения, повтором или CSV"
// @Param        locationType query string false "online/offline"
// @Param        excludeLocationType query string false "online/offline для исключения"
// @Param        city query string false "Город проведения"
// @Param        dateFrom query string false "Дата/время начала от, RFC3339 или YYYY-MM-DD"
// @Param        dateTo query string false "Дата/время начала до, RFC3339 или YYYY-MM-DD"
// @Param        hasFreeSlots query bool false "true - только со свободными местами, false - только заполненные"
// @Param        page query int false "Страница, по умолчанию 1"
// @Param        limit query int false "Размер страницы, по умолчанию 20, максимум 100"
// @Param        feed query string false "subscription_news для новинок подписок"
// @Success      200 {object} dto.EventSearchResponse "Страница событий"
// @Failure      400 {object} dto.ErrorResponse "Некорректные параметры поиска"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/events/search [get]
func (h *eventsHandler) SearchEvents(c *gin.Context) {
	userID := c.GetUint("userID")

	input, err := bindEventSearchQuery(c)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	result, err := h.reads.SearchEvents(c.Request.Context(), userID, input)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
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

	eventDto, err := h.reads.GetEventDetails(c.Request.Context(), userID, uint(eventID))
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
// @Failure      400      {object}  dto.ErrorResponse "Событие уже началось, заполнено или пользователь уже участвует"
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

	success, err := h.membership.JoinEvent(c.Request.Context(), userID, uint(eventID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "Событие не найдено")
		case errors.Is(err, events.ErrNotGroupMember):
			utils.Forbidden(c, "Вы не состоите в группе")
		case errors.Is(err, events.ErrEventAlreadyStarted):
			utils.BadRequest(c, err.Error())
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

	success, err := h.membership.LeaveEvent(c.Request.Context(), userID, uint(eventID))
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

func bindEventSearchQuery(c *gin.Context) (events.EventSearchInput, error) {
	var input events.EventSearchInput

	input.Query = firstQuery(c, "q", "query")

	groupID, err := parseOptionalUintQuery(c, "groupId")
	if err != nil {
		return input, err
	}
	input.GroupID = groupID

	input.CategoryIDs, err = parseUintListQuery(c, "categoryIds", "categoryId", "categories")
	if err != nil {
		return input, err
	}
	input.ExcludeCategoryIDs, err = parseUintListQuery(c, "excludeCategoryIds", "excludeCategoryId", "excludeCategories")
	if err != nil {
		return input, err
	}
	input.GenreIDs, err = parseUintListQuery(c, "genreIds", "genreId", "genres")
	if err != nil {
		return input, err
	}
	input.ExcludeGenreIDs, err = parseUintListQuery(c, "excludeGenreIds", "excludeGenreId", "excludeGenres")
	if err != nil {
		return input, err
	}
	input.EventTypeIDs, err = parseUintListQuery(c, "eventTypeIds", "eventTypeId")
	if err != nil {
		return input, err
	}
	input.ExcludeEventTypeIDs, err = parseUintListQuery(c, "excludeEventTypeIds", "excludeEventTypeId")
	if err != nil {
		return input, err
	}
	if hasQueryParam(c, "locationId") {
		return input, errors.New("Параметр locationId не используется в поиске событий, используйте locationType")
	}
	input.LocationTypes = collectQueryValues(c, "locationTypes", "locationType", "type")
	input.ExcludeLocationTypes = collectQueryValues(c, "excludeLocationTypes", "excludeLocationType", "excludeType")
	input.City = firstQuery(c, "city")

	input.DateFrom, err = parseOptionalTimeQuery(c, false, "dateFrom", "from")
	if err != nil {
		return input, err
	}
	input.DateTo, err = parseOptionalTimeQuery(c, true, "dateTo", "to")
	if err != nil {
		return input, err
	}
	input.HasFreeSlots, err = parseOptionalBoolQuery(c, "hasFreeSlots")
	if err != nil {
		return input, err
	}

	input.Page, err = parseOptionalIntQuery(c, 1, "page")
	if err != nil {
		return input, err
	}
	input.Limit, err = parseOptionalIntQuery(c, 20, "limit", "pageSize")
	if err != nil {
		return input, err
	}

	feed := strings.ToLower(strings.TrimSpace(firstQuery(c, "feed", "category")))
	if feed != "" {
		switch feed {
		case "subscription_news", "subscription-news", "новинки подписок":
			input.OnlySubscriptionNews = true
		default:
			return input, errors.New("Некорректный параметр feed")
		}
	}

	if err := validateEventSearchInput(input); err != nil {
		return input, err
	}

	return input, nil
}

func firstQuery(c *gin.Context, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(c.Query(name)); value != "" {
			return value
		}
	}
	return ""
}

func hasQueryParam(c *gin.Context, name string) bool {
	_, ok := c.Request.URL.Query()[name]
	return ok
}

func collectQueryValues(c *gin.Context, names ...string) []string {
	result := make([]string, 0)
	for _, name := range names {
		for _, raw := range c.QueryArray(name) {
			for _, part := range strings.Split(raw, ",") {
				if value := strings.TrimSpace(part); value != "" {
					result = append(result, value)
				}
			}
		}
	}
	return result
}

func parseUintListQuery(c *gin.Context, names ...string) ([]uint, error) {
	values, err := collectStrictCSVQueryValues(c, names...)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}

	result := make([]uint, 0, len(values))
	seen := make(map[uint]struct{}, len(values))
	for _, value := range values {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil || parsed == 0 {
			return nil, errors.New("Некорректный параметр " + names[0])
		}
		id := uint(parsed)
		if _, ok := seen[id]; ok {
			return nil, errors.New("Параметр " + names[0] + " содержит повторяющееся значение " + value)
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	if len(result) > 100 {
		return nil, errors.New("Параметр " + names[0] + " содержит слишком много значений")
	}

	return result, nil
}

func collectStrictCSVQueryValues(c *gin.Context, names ...string) ([]string, error) {
	result := make([]string, 0)
	for _, name := range names {
		for _, raw := range c.QueryArray(name) {
			if raw == "" {
				continue
			}
			for _, part := range strings.Split(raw, ",") {
				value := strings.TrimSpace(part)
				if value == "" {
					return nil, errors.New("Некорректный параметр " + name)
				}
				result = append(result, value)
			}
		}
	}
	return result, nil
}

func parseOptionalUintQuery(c *gin.Context, name string) (*uint, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return nil, errors.New("Некорректный параметр " + name)
	}

	result := uint(parsed)
	return &result, nil
}

func parseOptionalIntQuery(c *gin.Context, defaultValue int, names ...string) (int, error) {
	value := firstQuery(c, names...)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("Некорректный параметр " + names[0])
	}

	return parsed, nil
}

func parseOptionalBoolQuery(c *gin.Context, name string) (*bool, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, errors.New("Некорректный параметр " + name)
	}

	return &parsed, nil
}

func parseOptionalTimeQuery(c *gin.Context, endOfDay bool, names ...string) (*time.Time, error) {
	value := firstQuery(c, names...)
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, errors.New("Некорректный параметр " + names[0])
	}
	if endOfDay {
		parsed = parsed.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	return &parsed, nil
}

func validateEventSearchInput(input events.EventSearchInput) error {
	if input.Page < 1 {
		return errors.New("Параметр page должен быть больше 0")
	}
	if input.Page > 10000 {
		return errors.New("Параметр page слишком большой")
	}
	if input.Limit < 1 || input.Limit > 100 {
		return errors.New("Параметр limit должен быть от 1 до 100")
	}
	if len(input.Query) > 200 {
		return errors.New("Параметр q слишком длинный")
	}
	if len(input.City) > 100 {
		return errors.New("Параметр city слишком длинный")
	}
	if input.DateFrom != nil && input.DateTo != nil && input.DateFrom.After(*input.DateTo) {
		return errors.New("Параметр dateFrom не может быть позже dateTo")
	}
	if err := rejectUintIntersections("categoryIds", input.CategoryIDs, "excludeCategoryIds", input.ExcludeCategoryIDs); err != nil {
		return err
	}
	if err := rejectUintIntersections("genreIds", input.GenreIDs, "excludeGenreIds", input.ExcludeGenreIDs); err != nil {
		return err
	}
	if err := rejectUintIntersections("eventTypeIds", input.EventTypeIDs, "excludeEventTypeIds", input.ExcludeEventTypeIDs); err != nil {
		return err
	}
	if err := validateLocationTypeValues("locationType", input.LocationTypes); err != nil {
		return err
	}
	if err := validateLocationTypeValues("excludeLocationType", input.ExcludeLocationTypes); err != nil {
		return err
	}
	if err := rejectStringIntersections("locationType", input.LocationTypes, "excludeLocationType", input.ExcludeLocationTypes); err != nil {
		return err
	}
	return nil
}

func rejectUintIntersections(leftName string, left []uint, rightName string, right []uint) error {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}

	seen := make(map[uint]struct{}, len(left))
	for _, value := range left {
		seen[value] = struct{}{}
	}
	for _, value := range right {
		if _, ok := seen[value]; ok {
			return errors.New("Параметры " + leftName + " и " + rightName + " конфликтуют: " + strconv.FormatUint(uint64(value), 10))
		}
	}
	return nil
}

func validateLocationTypeValues(name string, values []string) error {
	for _, value := range values {
		switch normalizeLocationTypeKey(value) {
		case "", "online", "offline":
		default:
			return errors.New("Некорректный параметр " + name)
		}
	}
	return nil
}

func rejectStringIntersections(leftName string, left []string, rightName string, right []string) error {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(left))
	for _, value := range left {
		key := normalizeLocationTypeKey(value)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	for _, value := range right {
		key := normalizeLocationTypeKey(value)
		if _, ok := seen[key]; ok {
			return errors.New("Параметры " + leftName + " и " + rightName + " конфликтуют: " + key)
		}
	}
	return nil
}

func normalizeLocationTypeKey(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "online", "онлайн":
		return "online"
	case "offline", "off-line", "офлайн", "оффлайн":
		return "offline"
	case "":
		return ""
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}
