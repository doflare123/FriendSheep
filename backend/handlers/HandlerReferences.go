package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"friendship/services/references"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type ReferencesHandler interface {
	GetReferences(c *gin.Context)
	SearchGenres(c *gin.Context)
}

type referencesHandler struct {
	service references.ReferenceService
}

func NewReferencesHandler(service references.ReferenceService) ReferencesHandler {
	return &referencesHandler{service: service}
}

// GetReferences возвращает общие справочники.
// @Summary      Получить справочники
// @Description  Возвращает общие справочники групп и событий. Жанры доступны отдельно через GET /api/v2/references/genres.
// @Tags         references
// @Produce      json
// @Success      200 {object} dto.ReferencesDto "Справочники"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/references [get]
func (h *referencesHandler) GetReferences(c *gin.Context) {
	result, err := h.service.GetReferences(c.Request.Context())
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchGenres возвращает страницу жанров.
// @Summary      Получить жанры
// @Description  Возвращает жанры с регистронезависимым поиском по подстроке и пагинацией.
// @Tags         references
// @Produce      json
// @Param        q query string false "Подстрока названия жанра"
// @Param        page query int false "Номер страницы" default(1) minimum(1)
// @Param        limit query int false "Размер страницы" default(50) minimum(1) maximum(100)
// @Success      200 {object} dto.GenreSearchResponseDto "Страница жанров"
// @Failure      400 {object} dto.ErrorResponse "Некорректные параметры пагинации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/references/genres [get]
func (h *referencesHandler) SearchGenres(c *gin.Context) {
	page, err := parseReferencePositiveInt(c.Query("page"), references.DefaultGenrePage, references.MaxGenreLimit, false)
	if err != nil {
		utils.BadRequest(c, references.ErrInvalidGenrePage.Error())
		return
	}

	limit, err := parseReferencePositiveInt(c.Query("limit"), references.DefaultGenreLimit, references.MaxGenreLimit, true)
	if err != nil {
		utils.BadRequest(c, references.ErrInvalidGenreLimit.Error())
		return
	}

	result, err := h.service.SearchGenres(c.Request.Context(), references.GenreSearchInput{
		Query: strings.TrimSpace(c.Query("q")),
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, references.ErrInvalidGenrePage),
			errors.Is(err, references.ErrInvalidGenreLimit):
			utils.BadRequest(c, err.Error())
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseReferencePositiveInt(raw string, defaultValue, maxValue int, enforceMax bool) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || enforceMax && value > maxValue {
		return 0, errors.New("некорректный положительный параметр")
	}
	return value, nil
}
