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
	GetPopularEventsForViewer(ctx context.Context, viewerID uint) (*events.PopularEventsSnapshot, error)
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
// @Description  Возвращает публичный топ-10 популярных событий с маркировками 0+, 6+ и 12+. Персонализированный поиск доступен через /api/v2/events/search.
// @Description  При JWT пользователя младше 12 лет исключается также 12+; результат может содержать меньше 10 событий.
// @Tags         events
// @Accept       json
// @Produce      json
// @Success      200   {object}  dto.CachedPopularEvents  "Список популярных событий"
// @Failure      500   {object}  dto.ErrorResponse    "Ошибка сервера"
// @Router       /api/v2/events/popular [get]
func (h *popularEventsHandler) GetPopularEvents(c *gin.Context) {
	viewerID, _ := c.Get("userID")
	id, _ := viewerID.(uint)
	snapshot, err := h.srv.GetPopularEventsForViewer(c.Request.Context(), id)
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

	result.Events = make([]dto.EventSearchItemDto, len(snapshot.Events))
	for i, item := range snapshot.Events {
		result.Events[i] = dto.EventSearchItemDto{
			ID:    item.ID,
			Title: item.Title,
			Group: dto.EventSearchGroupDto{
				ID:         item.Group.ID,
				Name:       item.Group.Name,
				Image:      item.Group.Image,
				Enterprise: item.Group.Enterprise,
			},
			Image:        item.Image,
			CurrentUsers: item.CurrentUsers,
			MaxUsers:     item.MaxUsers,
			Duration:     item.Duration,
			StartTime:    item.StartTime,
			EventType:    item.EventType,
			LocationType: item.LocationType,
			AgeLimit:     item.AgeLimit,
			Status:       item.Status,
			City:         item.City,
			Genres:       append([]string(nil), item.Genres...),
			Subscribed:   item.Subscribed,
		}
	}

	return result
}
