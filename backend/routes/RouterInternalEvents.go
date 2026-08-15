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
