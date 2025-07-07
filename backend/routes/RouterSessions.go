package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterSessions(r *gin.Engine) {
	GroupSession := r.Group("api/sessions")
	{
		GroupSession.POST("/createSession", middlewares.JWTAuthMiddleware(), handlers.CreateSession)
		GroupSession.DELETE("/sessions/:id", middlewares.JWTAuthMiddleware(), handlers.DeleteSession)
	}
}
