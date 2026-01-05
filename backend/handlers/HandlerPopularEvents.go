package handlers

import (
	"friendship/models/dto"
	"friendship/services/events"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PopularEventsHandler interface {
	GetPopularEvents(c *gin.Context)
}

type popularEventsHandler struct {
	srv events.PopularEventsService
}

func NewPopularEventsHandler(srv events.PopularEventsService) PopularEventsHandler {
	return &popularEventsHandler{
		srv: srv,
	}
}

// GetPopularEvents godoc
// @Summary      Получить популярные события
// @Description  Возвращает топ-10 самых популярных событий
// @Tags         events
// @Accept       json
// @Produce      json
// @Success      200   {object}  dto.CachedPopularEvents  "Список популярных событий"
// @Failure      500   {object}  map[string]string    "Ошибка сервера"
// @Router       /api/v2/events/popular [get]
func (h *popularEventsHandler) GetPopularEvents(c *gin.Context) {
	events, err := h.srv.GetPopularEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Ошибка получения популярных событий",
		})
		return
	}

	c.JSON(http.StatusOK, events)
}
