package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterUsersInfo(r *gin.Engine) {
	UserGroup := r.Group("api/users")
	{
		UserGroup.GET("/new", middlewares.JWTAuthMiddleware(), handlers.GetNewSessions)
	}
}
