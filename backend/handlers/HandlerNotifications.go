package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"friendship/services/notifications"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

const (
	defaultInboxLimit = 20
	maxInboxLimit     = 100
)

type NotificationsHandler struct {
	inbox notifications.NotificationInbox
}

func NewNotificationsHandler(inbox notifications.NotificationInbox) *NotificationsHandler {
	return &NotificationsHandler{inbox: inbox}
}

// List возвращает страницу уведомлений текущего пользователя.
// @Summary      Получить уведомления текущего пользователя
// @Description  Возвращает уведомления только из inbox авторизованного пользователя с курсорной пагинацией и необязательным фильтром непрочитанных.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  query     string  false  "Курсор следующей страницы"
// @Param        limit   query     int     false  "Размер страницы (по умолчанию 20, максимум 100)"  minimum(1)  maximum(100)
// @Param        unread  query     bool    false  "Вернуть только непрочитанные уведомления"
// @Success      200     {object}  notifications.InboxPage
// @Failure      400     {object}  dto.ErrorResponse
// @Failure      401     {object}  dto.ErrorResponse
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /api/v2/users/me/notifications [get]
func (h *NotificationsHandler) List(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
		return
	}
	limit := defaultInboxLimit
	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maxInboxLimit {
			utils.BadRequest(c, "invalid_limit", utils.WithMessage("Некорректный limit уведомлений"))
			return
		}
		limit = value
	}
	unreadOnly := false
	if raw := c.Query("unread"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			utils.BadRequest(c, "invalid_unread_filter", utils.WithMessage("Некорректный параметр unread"))
			return
		}
		unreadOnly = value
	}
	page, err := h.inbox.List(c.Request.Context(), userID, c.Query("cursor"), limit, unreadOnly)
	if err != nil {
		handleInboxError(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// UnreadCount возвращает число непрочитанных уведомлений текущего пользователя.
// @Summary      Получить число непрочитанных уведомлений
// @Description  Возвращает число непрочитанных уведомлений только для авторизованного пользователя.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  notifications.UnreadCount
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v2/users/me/notifications/unread-count [get]
func (h *NotificationsHandler) UnreadCount(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
		return
	}
	result, err := h.inbox.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		handleInboxError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// MarkRead отмечает уведомление текущего пользователя как прочитанное.
// @Summary      Отметить уведомление прочитанным
// @Description  Идемпотентно отмечает принадлежащее авторизованному пользователю уведомление как прочитанное.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        notificationId  path      string  true  "ID уведомления"
// @Success      200             {object}  notifications.MarkReadResult
// @Failure      400             {object}  dto.ErrorResponse
// @Failure      401             {object}  dto.ErrorResponse
// @Failure      404             {object}  dto.ErrorResponse
// @Failure      500             {object}  dto.ErrorResponse
// @Router       /api/v2/users/me/notifications/{notificationId}/read [patch]
func (h *NotificationsHandler) MarkRead(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		utils.Unauthorized(c, "missing_authenticated_user", utils.WithMessage("Пользователь не авторизован"))
		return
	}
	notificationID := strings.TrimSpace(c.Param("notificationId"))
	if notificationID == "" {
		utils.BadRequest(c, "invalid_notification_id", utils.WithMessage("Некорректный ID уведомления"))
		return
	}
	result, err := h.inbox.MarkRead(c.Request.Context(), userID, notificationID)
	if err != nil {
		handleInboxError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func authenticatedUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	switch typed := value.(type) {
	case uint:
		return typed, typed > 0
	case uint64:
		return uint(typed), typed > 0
	case int:
		return uint(typed), typed > 0
	default:
		return 0, false
	}
}

func handleInboxError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, notifications.ErrInboxInvalidInput):
		utils.BadRequest(c, "invalid_notification_request", utils.WithMessage("Некорректный запрос уведомлений"))
	case errors.Is(err, notifications.ErrInboxNotFound):
		utils.NotFound(c, "notification_not_found", utils.WithMessage("Уведомление не найдено"))
	default:
		utils.InternalError(c, "notification_service_unavailable", utils.WithMessage("Сервис уведомлений временно недоступен"))
	}
}
