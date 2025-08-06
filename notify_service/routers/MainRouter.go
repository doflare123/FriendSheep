package routers

import (
	"notify_service/handlers"

	"github.com/gin-gonic/gin"
)

func RoutesPush(r *gin.Engine) {
	PushGroup := r.Group("api/push")
	{
		PushGroup.POST("/register", handlers.RegisterDevice)
		// PushGroup.POST("/send", handlers.SendNotification)
		// PushGroup.POST("/sendMulticast", handlers.SendMulticastNotification)
	}

}
