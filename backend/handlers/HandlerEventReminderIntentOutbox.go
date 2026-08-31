package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"friendship/services/events"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type EventReminderIntentOutboxHandler interface {
	List(c *gin.Context)
}

type eventReminderIntentOutboxHandler struct {
	service events.EventReminderIntentService
}

func NewEventReminderIntentOutboxHandler(service events.EventReminderIntentService) EventReminderIntentOutboxHandler {
	return &eventReminderIntentOutboxHandler{service: service}
}

func (h *eventReminderIntentOutboxHandler) List(c *gin.Context) {
	after, err := parseEventReminderAfter(c.Query("after"))
	if err != nil {
		utils.BadRequest(c, "invalid_after", utils.WithMessage("Некорректный cursor notification intents"))
		return
	}
	limit, err := parseEventReminderLimit(c.Query("limit"))
	if err != nil {
		utils.BadRequest(c, "invalid_limit", utils.WithMessage("Некорректный limit notification intents"))
		return
	}

	page, err := h.service.List(c.Request.Context(), after, limit)
	if err != nil {
		if errors.Is(err, events.ErrInvalidEventReminderIntentLimit) {
			utils.BadRequest(c, "invalid_limit", utils.WithMessage("Некорректный limit notification intents"))
			return
		}
		utils.InternalError(c, "notification_intent_read_failed", utils.WithMessage("Не удалось прочитать notification intents"))
		return
	}
	c.JSON(http.StatusOK, page)
}

func parseEventReminderAfter(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, events.ErrInvalidEventReminderIntentCursor
	}
	return value, nil
}

func parseEventReminderLimit(raw string) (int, error) {
	if raw == "" {
		return events.DefaultEventReminderIntentLimit, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > events.MaxEventReminderIntentLimit {
		return 0, events.ErrInvalidEventReminderIntentLimit
	}
	return value, nil
}
