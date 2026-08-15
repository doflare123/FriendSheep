package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"friendship/services/events"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type EventLifecycleHandler interface {
	Advance(c *gin.Context)
}

type eventLifecycleHandler struct {
	service events.EventLifecycleService
}

func NewEventLifecycleHandler(service events.EventLifecycleService) EventLifecycleHandler {
	return &eventLifecycleHandler{service: service}
}

func (h *eventLifecycleHandler) Advance(c *gin.Context) {
	parsedID, err := strconv.ParseUint(c.Param("eventId"), 10, 64)
	if err != nil || parsedID == 0 || parsedID > uint64(^uint(0)) {
		utils.BadRequest(c, "invalid_event_id", utils.WithMessage("Некорректный ID мероприятия"))
		return
	}

	result, err := h.service.Advance(c.Request.Context(), uint(parsedID))
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			utils.NotFound(c, "event_not_found", utils.WithMessage("Мероприятие не найдено"))
		case errors.Is(err, events.ErrInvalidEventLifecycleState):
			utils.JSONError(c, http.StatusConflict, "invalid_event_lifecycle_state", utils.WithMessage("Некорректное состояние жизненного цикла мероприятия"))
		default:
			utils.InternalError(c, "event_lifecycle_failed", utils.WithMessage("Не удалось обработать жизненный цикл мероприятия"))
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
