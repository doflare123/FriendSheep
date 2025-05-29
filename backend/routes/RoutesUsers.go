package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RoutesUsers(r *gin.Engine) {
	RegisretGroup := r.Group("api/users")
	{
		RegisretGroup.POST("/register", handlers.CreateUserHandler)
		RegisretGroup.GET("/sessionsReg", handlers.CreateSessionRegisterHandler)
		RegisretGroup.PATCH("/verifySessionsReg", handlers.VerifySessionHandler)
	}
}
