package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RoutesUsers(r *gin.Engine) {
	UserGroup := r.Group("api/users")
	{
		UserGroup.POST("/register", handlers.CreateUserHandler)
		UserGroup.POST("/sessionsReg")
	}
}
