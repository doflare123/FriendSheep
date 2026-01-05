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
	router.GET("/api/v2/events/genres", eventsHandler.GetAllGenres)
	router.GET("/api/v2/references", eventsHandler.GetAllReferences)
	router.GET("/api/v2/events/popular", popularHandler.GetPopularEvents)

	events := router.Group("/api/v2/events")
	events.Use(authMiddleware.RequireAuth())
	{
		events.GET("/:eventId", eventsHandler.GetEventDetails)
		events.POST("/:eventId/join", eventsHandler.JoinEvent)
		events.POST("/:eventId/leave", eventsHandler.LeaveEvent)
	}
	eventsGroup := router.Group("/api/v2/groups/events")
	eventsGroup.Use(authMiddleware.RequireAuth(), groupRoleMiddleware.RequireOperatorOrAdmin())
	{
		eventsGroup.GET("/:groupId/events", eventsHandler.GetGroupEvents)
	}
	eventsAdmin := router.Group("/api/v2/admin/events")
	eventsAdmin.Use(authMiddleware.RequireAuth(), groupRoleMiddleware.RequireOperatorOrAdmin())
	{
		eventsAdmin.GET("/:eventId", eventsHandler.GetEventDetailsForAdmin)
		eventsAdmin.POST("", eventsHandler.CreateEvent)
		eventsAdmin.PUT("/:eventId", eventsHandler.UpdateEvent)
		eventsAdmin.DELETE("/:eventId", eventsHandler.DeleteEvent)
		eventsAdmin.DELETE("/:eventId/kick/:userId", eventsHandler.KickUserFromEvent)
	}
}
