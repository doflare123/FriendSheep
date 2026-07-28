package handlers

import (
	"context"
	"friendship/models/dto"
	"friendship/services/events"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PopularEventsHandler interface {
	GetPopularEvents(c *gin.Context)
}

type popularEventsReader interface {
	GetPopularEvents(ctx context.Context) (*events.PopularEventsSnapshot, error)
}

type popularEventsHandler struct {
	srv popularEventsReader
}

func NewPopularEventsHandler(srv popularEventsReader) PopularEventsHandler {
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
// @Failure      500   {object}  dto.ErrorResponse    "Ошибка сервера"
// @Router       /api/v2/events/popular [get]
func (h *popularEventsHandler) GetPopularEvents(c *gin.Context) {
	snapshot, err := h.srv.GetPopularEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Ошибка получения популярных событий",
		})
		return
	}

	c.JSON(http.StatusOK, popularEventsSnapshotToDTO(snapshot))
}

func popularEventsSnapshotToDTO(snapshot *events.PopularEventsSnapshot) dto.CachedPopularEvents {
	if snapshot == nil {
		return dto.CachedPopularEvents{}
	}

	result := dto.CachedPopularEvents{
		UpdatedAt: snapshot.UpdatedAt,
		Count:     snapshot.Count,
	}
	if len(snapshot.Events) == 0 {
		return result
	}

	result.Events = make([]dto.EventShortDto, len(snapshot.Events))
	for i, item := range snapshot.Events {
		result.Events[i] = dto.EventShortDto{
			ID:           item.ID,
			Title:        item.Title,
			ImageURL:     item.ImageURL,
			MaxUsers:     item.MaxUsers,
			CurrentUsers: item.CurrentUsers,
			EventType:    item.EventType,
			LocationType: item.LocationType,
			AgeLimit:     item.AgeLimit,
			Genres:       append([]string(nil), item.Genres...),
			StartTime:    item.StartTime,
			Duration:     item.Duration,
			EventID:      item.EventID,
			GroupID:      item.GroupID,
			Status:       item.Status,
			Subscribed:   item.Subscribed,
		}
	}

	return result
}
