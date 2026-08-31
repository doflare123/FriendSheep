package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"friendship/services/notifications"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type EventReminderRecipientsHandler interface {
	Resolve(c *gin.Context)
}

type eventReminderRecipientsHandler struct {
	service notifications.EventReminderRecipientService
}

func NewEventReminderRecipientsHandler(service notifications.EventReminderRecipientService) EventReminderRecipientsHandler {
	return &eventReminderRecipientsHandler{service: service}
}

func (h *eventReminderRecipientsHandler) Resolve(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("eventId"), 10, 64)
	if err != nil || eventID == 0 {
		utils.BadRequest(c, "invalid_event_id", utils.WithMessage("Некорректный ID мероприятия"))
		return
	}
	offset, err := strconv.Atoi(c.Query("reminderOffsetMinutes"))
	if err != nil || offset <= 0 {
		utils.BadRequest(c, "invalid_reminder_offset", utils.WithMessage("Некорректный интервал напоминания"))
		return
	}

	response, err := h.service.Resolve(c.Request.Context(), uint(eventID), offset)
	if err != nil {
		switch {
		case errors.Is(err, notifications.ErrEventNotFound):
			utils.NotFound(c, "event_not_found", utils.WithMessage("Мероприятие не найдено"))
		case errors.Is(err, notifications.ErrInvalidReminderOffset):
			utils.BadRequest(c, "invalid_reminder_offset", utils.WithMessage("Некорректный интервал напоминания"))
		default:
			utils.InternalError(c, "reminder_recipient_resolution_failed", utils.WithMessage("Не удалось определить получателей напоминания"))
		}
		return
	}
	c.JSON(http.StatusOK, response)
}
