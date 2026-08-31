package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(router *gin.Engine, handler *handlers.NotificationsHandler, jwtAuth gin.HandlerFunc) {
	group := router.Group("/api/v2/users/me/notifications")
	group.Use(jwtAuth)
	group.GET("", handler.List)
	group.GET("/unread-count", handler.UnreadCount)
	group.PATCH("/:notificationId/read", handler.MarkRead)
}
