package routers

import (
	"github.com/gin-gonic/gin"
)

func RoutesPush(r *gin.Engine) {
	PushGroup := r.Group("api/push")
	{
		PushGroup.POST("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
		// PushGroup.POST("/register", handlers.RegisterDevice)
		// PushGroup.POST("/send", handlers.SendNotification)
		// PushGroup.POST("/sendMulticast", handlers.SendMulticastNotification)
	}

}
