package routes

import (
	"friendship/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterStatisticsRoutes(router *gin.Engine, handler *handlers.StatisticsHandler, jwtAuth gin.HandlerFunc) {
	group := router.Group("/api/v2/users/me")
	group.Use(jwtAuth)
	group.GET("/statistics", handler.Read)
}
func RegisterInternalStatisticsRoutes(router *gin.Engine, handler *handlers.StatisticsHandler, internalAuth gin.HandlerFunc) {
	group := router.Group("/internal/v1")
	group.Use(internalAuth)
	group.GET("/statistics/triggers", handler.ListTriggers)
	group.POST("/statistics/backfill", handler.Backfill)
	group.POST("/events/:eventId/statistics/process", handler.Process)
}
