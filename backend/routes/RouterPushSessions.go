package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterInternalPushSessionRoutes(router *gin.Engine, handler *handlers.PushSessionHandler, internalAuth gin.HandlerFunc) {
	router.POST("/internal/v1/push-sessions/check", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	}, internalAuth, handler.Check)
}
