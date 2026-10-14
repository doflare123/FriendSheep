package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"friendship/services/events"
	"friendship/services/statistics"
	"friendship/utils"
	"github.com/gin-gonic/gin"
)

type StatisticsApplication interface {
	Read(context.Context, uint) (statistics.View, error)
	Process(context.Context, uint) (statistics.ProcessResult, error)
}
type StatisticsHandler struct {
	service  StatisticsApplication
	triggers events.StatisticsTriggerService
}

func NewStatisticsHandler(service StatisticsApplication, triggers events.StatisticsTriggerService) *StatisticsHandler {
	return &StatisticsHandler{service: service, triggers: triggers}
}

// Read возвращает готовую статистику текущего пользователя.
// @Summary Получить собственную статистику участия
// @Description Учитывает состав завершённых мероприятий, а не физическое посещение. plannedMinutes — сумма запланированных минут; календарь UTC. Агрегаты обновляются асинхронно. lastUpdatedAt не гарантирует отсутствие backlog. При равных максимумах выбирается минимальный ID или ISO день недели. Сумма жанров может превышать число мероприятий.
// @Tags Statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} statistics.View
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/me/statistics [get]
func (h *StatisticsHandler) Read(c *gin.Context) {
	id, ok := authenticatedUserID(c)
	if !ok {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
		return
	}
	view, err := h.service.Read(c.Request.Context(), id)
	if err != nil {
		utils.InternalError(c, "statistics_read_failed", utils.WithMessage("Не удалось получить статистику"))
		return
	}
	c.JSON(http.StatusOK, view)
}
func (h *StatisticsHandler) Process(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("eventId"), 10, 64)
	if err != nil || id == 0 || uint64(uint(id)) != id {
		utils.BadRequest(c, "invalid_event_id", utils.WithMessage("Некорректный ID мероприятия"))
		return
	}
	result, err := h.service.Process(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, statistics.ErrInvalidInput) {
			utils.BadRequest(c, "invalid_statistics_event", utils.WithMessage("Некорректные исходные данные статистики"))
			return
		}
		utils.InternalError(c, "statistics_process_failed", utils.WithMessage("Не удалось обработать статистику"))
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *StatisticsHandler) ListTriggers(c *gin.Context) {
	after, limit, ok := statisticsPagination(c, "afterSequence")
	if !ok {
		return
	}
	page, err := h.triggers.List(c.Request.Context(), after, limit)
	if err != nil {
		utils.InternalError(c, "statistics_triggers_read_failed", utils.WithMessage("Не удалось прочитать задания статистики"))
		return
	}
	c.JSON(http.StatusOK, page)
}
func (h *StatisticsHandler) Backfill(c *gin.Context) {
	after, limit, ok := statisticsPagination(c, "afterEventId")
	if !ok {
		return
	}
	if uint64(uint(after)) != uint64(after) {
		utils.BadRequest(c, "invalid_cursor")
		return
	}
	result, err := h.triggers.Backfill(c.Request.Context(), uint(after), limit)
	if err != nil {
		utils.InternalError(c, "statistics_backfill_failed", utils.WithMessage("Не удалось опубликовать исторические задания статистики"))
		return
	}
	c.JSON(http.StatusOK, result)
}
func statisticsPagination(c *gin.Context, key string) (int64, int, bool) {
	after := int64(0)
	limit := events.DefaultStatisticsTriggerLimit
	var err error
	if raw := c.Query(key); raw != "" {
		after, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || after < 0 {
			utils.BadRequest(c, "invalid_cursor", utils.WithMessage("Некорректный курсор статистики"))
			return 0, 0, false
		}
	}
	if raw := c.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > events.MaxStatisticsTriggerLimit {
			utils.BadRequest(c, "invalid_limit", utils.WithMessage("Некорректный limit статистики"))
			return 0, 0, false
		}
	}
	return after, limit, true
}
