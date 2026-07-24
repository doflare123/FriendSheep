package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterEventsRoutes(
	router *gin.Engine,
	eventsHandler handlers.EventsHandler,
	popularHandler handlers.PopularEventsHandler,
	authMiddleware *middlewares.AuthMiddleware,
	groupRoleMiddleware *middlewares.GroupRoleMiddleware,
) {
	router.GET("/api/v2/events/popular", popularHandler.GetPopularEvents)
	router.GET("/api/v2/events/search", authMiddleware.OptionalAuth(), eventsHandler.SearchEvents)

	events := router.Group("/api/v2/events")
	events.Use(authMiddleware.RequireAuth())
	{
		events.GET("/:eventId", eventsHandler.GetEventDetails)
		events.POST("/:eventId/join", eventsHandler.JoinEvent)
		events.POST("/:eventId/leave", eventsHandler.LeaveEvent)
	}

	eventsGroup := router.Group("/api/v2/groups/events")
	eventsGroup.Use(authMiddleware.RequireAuth())
	{
		eventsGroup.GET("/:groupId/events", eventsHandler.GetGroupEvents)
	}

	eventsAdmin := router.Group("/api/v2/admin/events")
	eventsAdmin.Use(authMiddleware.RequireAuth())
	{
		eventsAdmin.GET("/:eventId", groupRoleMiddleware.RequireEventOperatorOrAdmin(), eventsHandler.GetEventDetailsForAdmin)
		eventsAdmin.POST("", groupRoleMiddleware.RequireOperatorOrAdmin(), eventsHandler.CreateEvent)
		eventsAdmin.PUT("/:eventId", groupRoleMiddleware.RequireEventOperatorOrAdmin(), eventsHandler.UpdateEvent)
		eventsAdmin.DELETE("/:eventId", groupRoleMiddleware.RequireEventOperatorOrAdmin(), eventsHandler.DeleteEvent)
		eventsAdmin.DELETE("/:eventId/kick/:userId", groupRoleMiddleware.RequireEventOperatorOrAdmin(), eventsHandler.KickUserFromEvent)
	}
}
