package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"friendship/services/statistics"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type UniquePeopleApplication interface {
	Claim(context.Context, int) (statistics.UniquePeopleClaims, error)
	Execute(context.Context, statistics.UniquePeopleClaim) (statistics.UniquePeopleExecution, error)
	Backfill(context.Context, uint, int) (statistics.UniquePeopleBackfillResult, error)
	Metrics(context.Context) (statistics.UniquePeopleMetrics, error)
}

type UniquePeopleHandler struct{ service UniquePeopleApplication }

func NewUniquePeopleHandler(service UniquePeopleApplication) *UniquePeopleHandler {
	return &UniquePeopleHandler{service: service}
}

func (h *UniquePeopleHandler) Claim(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	limit := 1
	if raw := c.Query("limit"); raw != "" {
		var err error
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 8 {
			utils.BadRequest(c, "invalid_unique_people_limit")
			return
		}
	}
	result, err := h.service.Claim(ctx, limit)
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UniquePeopleHandler) Execute(c *gin.Context) {
	var request struct {
		statistics.UniquePeopleClaim
		SchemaVersion *int `json:"schemaVersion,omitempty"`
	}
	// Ограничение тела и запрет второго JSON не дают частично принять команду.
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 4096))
	if err := decoder.Decode(&request); err != nil {
		utils.BadRequest(c, "invalid_unique_people_claim")
		return
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		utils.BadRequest(c, "invalid_unique_people_claim")
		return
	}
	if request.SchemaVersion != nil && *request.SchemaVersion != statistics.UniquePeopleSchemaVersion {
		utils.BadRequest(c, "invalid_unique_people_version")
		return
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 35*time.Second)
	defer cancel()
	result, err := h.service.Execute(ctx, request.UniquePeopleClaim)
	if err != nil {
		slog.Warn("ошибка пересчёта уникальных соучастников", "duration_ms", time.Since(started).Milliseconds(), "timeout", errors.Is(err, context.DeadlineExceeded))
		h.failure(c, err)
		return
	}
	slog.Info("пересчёт уникальных соучастников", "outcome", result.Outcome, "duration_ms", time.Since(started).Milliseconds())
	c.JSON(http.StatusOK, result)
}

func (h *UniquePeopleHandler) Backfill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	after, limit, ok := statisticsPagination(c, "afterUserId")
	if !ok {
		return
	}
	if uint64(uint(after)) != uint64(after) {
		utils.BadRequest(c, "invalid_cursor")
		return
	}
	result, err := h.service.Backfill(ctx, uint(after), limit)
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UniquePeopleHandler) Metrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	result, err := h.service.Metrics(ctx)
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UniquePeopleHandler) failure(c *gin.Context, err error) {
	if errors.Is(err, statistics.ErrInvalidInput) {
		utils.BadRequest(c, "invalid_unique_people_request")
		return
	}
	utils.InternalError(c, "unique_people_failed", utils.WithMessage("Не удалось обработать пересчёт статистики"))
}
