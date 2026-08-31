package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterInternalEventLifecycleRoutes(
	router *gin.Engine,
	handler handlers.EventLifecycleHandler,
	internalAuth gin.HandlerFunc,
) {
	internal := router.Group("/internal/v1")
	internal.Use(internalAuth)
	internal.POST("/events/:eventId/lifecycle/advance", handler.Advance)
}

func RegisterInternalEventLifecycleScheduleRoutes(
	router *gin.Engine,
	handler handlers.EventLifecycleScheduleOutboxHandler,
	internalAuth gin.HandlerFunc,
) {
	internal := router.Group("/internal/v1")
	internal.Use(internalAuth)
	internal.GET("/event-lifecycle/schedule-events", handler.List)
}

func RegisterInternalEventReminderIntentRoutes(
	router *gin.Engine,
	handler handlers.EventReminderIntentOutboxHandler,
	internalAuth gin.HandlerFunc,
) {
	internal := router.Group("/internal/v1")
	internal.Use(internalAuth)
	internal.GET("/notification-intents/event-reminders", handler.List)
}

func RegisterInternalEventReminderRecipientRoutes(
	router *gin.Engine,
	handler handlers.EventReminderRecipientsHandler,
	internalAuth gin.HandlerFunc,
) {
	internal := router.Group("/internal/v1")
	internal.Use(internalAuth)
	internal.GET("/event-reminders/:eventId/recipients", handler.Resolve)
}
