package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterPushDeviceRoutes(router *gin.Engine, handler *handlers.PushDevicesHandler, jwtAuth gin.HandlerFunc) {
	group := router.Group("/api/v2/users/me/push-devices")
	group.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	})
	group.Use(jwtAuth)
	group.PUT("/:installationId", handler.Register)
	group.DELETE("/:installationId", handler.Detach)
}
