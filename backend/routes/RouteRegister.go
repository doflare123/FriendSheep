package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RoutesUsers(r *gin.Engine) {
	RegisretGroup := r.Group("api/users")
	{
		RegisretGroup.POST("", handlers.CreateUserHandler)
	}
	SessionGroup := r.Group("api/sessions")
	{
		SessionGroup.POST("/register", handlers.CreateSessionRegisterHandler)
		SessionGroup.PATCH("/verify", handlers.VerifySessionHandler)
	}
}
