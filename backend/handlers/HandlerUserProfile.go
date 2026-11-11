package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"friendship/services/users"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type UserProfileApplication interface {
	GetProfile(context.Context, uint, users.Target) (users.Profile, error)
	GetPrivacy(context.Context, uint) (users.Privacy, error)
	PatchPrivacy(context.Context, uint, users.PrivacyPatch) (users.Privacy, error)
}

type UserProfileHandler struct {
	service UserProfileApplication
}

func NewUserProfileHandler(service UserProfileApplication) *UserProfileHandler {
	return &UserProfileHandler{service: service}
}

// GetOwnProfile возвращает профиль владельца без загрузки истории и планов.
// @Summary Получить собственный профиль
// @Description Личность определяется проверенным JWT. Приватность не ограничивает владельцу доступ к компактной статистике; полная статистика доступна отдельно через /api/v2/users/me/statistics.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} users.Profile
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users [get]
func (h *UserProfileHandler) GetOwnProfile(c *gin.Context) {
	h.readProfile(c, users.OwnTarget())
}

// GetProfile разрешает адрес пользователя и применяет его настройки приватности.
// @Summary Получить профиль по ID или уникальному us
// @Description Положительное число означает ID, строка с @ — уникальный us, включая числовой. Ноль и остальные форматы запрещены. Владелец определяется по устойчивому ID. Чужой closed-профиль возвращает 403; open показывает компактную статистику, custom — только выбранные поля. Планы и история не возвращаются.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param identifier path string true "Положительный ID или @us"
// @Success 200 {object} users.Profile
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/{identifier} [get]
func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	if _, ok := profileUserID(c); !ok {
		return
	}
	target, err := users.ParseIdentifier(c.Param("identifier"))
	if err != nil {
		handleUserProfileError(c, err)
		return
	}
	h.readProfile(c, target)
}

func (h *UserProfileHandler) readProfile(c *gin.Context, target users.Target) {
	viewerID, ok := profileUserID(c)
	if !ok {
		return
	}
	profile, err := h.service.GetProfile(c.Request.Context(), viewerID, target)
	if err != nil {
		handleUserProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

// GetPrivacy возвращает настройки исключительно текущего пользователя.
// @Summary Получить свои настройки приватности
// @Description Отсутствующая запись означает open с полным набором компактных показателей.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} users.Privacy
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/me/privacy [get]
func (h *UserProfileHandler) GetPrivacy(c *gin.Context) {
	userID, ok := profileUserID(c)
	if !ok {
		return
	}
	privacy, err := h.service.GetPrivacy(c.Request.Context(), userID)
	if err != nil {
		handleUserProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, privacy)
}

// PatchPrivacy изменяет только переданные настройки текущего пользователя.
// @Summary Изменить свои настройки приватности
// @Description Допустимы mode (open/closed/custom) и visibleStatistics. Массив полностью заменяет выбор для custom; [] скрывает все показатели. Пропущенные поля сохраняются. null, неизвестные поля, повторы показателей и пустой запрос запрещены. open показывает все показатели независимо от сохранённого выбора; closed закрывает профиль для других.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param privacy body users.PrivacyPatch true "Изменяемые настройки"
// @Success 200 {object} users.Privacy
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/me/privacy [patch]
func (h *UserProfileHandler) PatchPrivacy(c *gin.Context) {
	userID, ok := profileUserID(c)
	if !ok {
		return
	}
	patch, err := decodePrivacyPatch(c)
	if err != nil {
		handleUserProfileError(c, users.ErrInvalidPrivacy)
		return
	}
	privacy, err := h.service.PatchPrivacy(c.Request.Context(), userID, patch)
	if err != nil {
		handleUserProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, privacy)
}

func decodePrivacyPatch(c *gin.Context) (users.PrivacyPatch, error) {
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 8192))
	var fields map[string]json.RawMessage
	if err := decoder.Decode(&fields); err != nil || len(fields) == 0 {
		return users.PrivacyPatch{}, users.ErrInvalidPrivacy
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return users.PrivacyPatch{}, users.ErrInvalidPrivacy
	}
	var patch users.PrivacyPatch
	for field, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return users.PrivacyPatch{}, users.ErrInvalidPrivacy
		}
		switch field {
		case "mode":
			var mode users.PrivacyMode
			if err := json.Unmarshal(raw, &mode); err != nil {
				return users.PrivacyPatch{}, users.ErrInvalidPrivacy
			}
			patch.Mode = &mode
		case "visibleStatistics":
			var selection []users.StatisticField
			if err := json.Unmarshal(raw, &selection); err != nil {
				return users.PrivacyPatch{}, users.ErrInvalidPrivacy
			}
			patch.VisibleStatistics = &selection
		default:
			return users.PrivacyPatch{}, users.ErrInvalidPrivacy
		}
	}
	return patch, nil
}

func profileUserID(c *gin.Context) (uint, bool) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
	}
	return userID, ok
}

func handleUserProfileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, users.ErrUnauthenticated):
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
	case errors.Is(err, users.ErrInvalidTarget):
		utils.BadRequest(c, "invalid_user_identifier", utils.WithMessage("Укажите положительный ID или @us"))
	case errors.Is(err, users.ErrInvalidPrivacy):
		utils.BadRequest(c, "invalid_privacy", utils.WithMessage("Некорректные настройки приватности"))
	case errors.Is(err, users.ErrUserNotFound):
		utils.NotFound(c, "user_not_found", utils.WithMessage("Пользователь не найден"))
	case errors.Is(err, users.ErrProfileClosed):
		utils.Forbidden(c, "profile_closed", utils.WithMessage("Профиль закрыт"))
	default:
		utils.InternalError(c, "user_profile_failed", utils.WithMessage("Не удалось обработать профиль"))
	}
}
