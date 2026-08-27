package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"friendship/services/events"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type EventLifecycleScheduleOutboxHandler interface {
	List(c *gin.Context)
}

type eventLifecycleScheduleOutboxHandler struct {
	service events.EventLifecycleScheduleOutboxService
}

func NewEventLifecycleScheduleOutboxHandler(service events.EventLifecycleScheduleOutboxService) EventLifecycleScheduleOutboxHandler {
	return &eventLifecycleScheduleOutboxHandler{service: service}
}

func (h *eventLifecycleScheduleOutboxHandler) List(c *gin.Context) {
	after, err := parseScheduleEventsAfter(c.Query("after"))
	if err != nil {
		utils.BadRequest(c, "invalid_after", utils.WithMessage("Некорректный cursor расписания мероприятий"))
		return
	}
	limit, err := parseScheduleEventsLimit(c.Query("limit"))
	if err != nil {
		utils.BadRequest(c, "invalid_limit", utils.WithMessage("Некорректный limit расписания мероприятий"))
		return
	}

	page, err := h.service.List(c.Request.Context(), after, limit)
	if err != nil {
		if errors.Is(err, events.ErrInvalidScheduleEventsLimit) {
			utils.BadRequest(c, "invalid_limit", utils.WithMessage("Некорректный limit расписания мероприятий"))
			return
		}
		utils.InternalError(c, "event_lifecycle_schedule_read_failed", utils.WithMessage("Не удалось прочитать расписание мероприятий"))
		return
	}
	c.JSON(http.StatusOK, page)
}

func parseScheduleEventsAfter(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		return 0, events.ErrInvalidScheduleEventsCursor
	}
	return parsed, nil
}

func parseScheduleEventsLimit(raw string) (int, error) {
	if raw == "" {
		return events.DefaultLifecycleScheduleEventsLimit, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 1 || parsed > events.MaxLifecycleScheduleEventsLimit {
		return 0, events.ErrInvalidScheduleEventsLimit
	}
	return parsed, nil
}
