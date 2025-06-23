package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterGroups(r *gin.Engine) {
	GroupGroup := r.Group("api/groups")
	{
		GroupGroup.POST("/createGroup", middlewares.JWTAuthMiddleware(), handlers.CreateGroup)
		// GroupGroup.POST("/refresh", handlers.RefreshTokenHandler)
	}
}
