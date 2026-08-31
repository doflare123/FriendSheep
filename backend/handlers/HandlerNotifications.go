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
