package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterUsersInfo(r *gin.Engine) {
	UserGroup := r.Group("api/users")
	{
		UserGroup.GET("/sessions/new", middlewares.JWTAuthMiddleware(), handlers.GetNewSessions)
		UserGroup.GET("/sessions/popular", middlewares.JWTAuthMiddleware(), handlers.GetPopularSessions)
		UserGroup.GET("/sessions/category", middlewares.JWTAuthMiddleware(), handlers.GetCategorySessions)
		UserGroup.GET("/sessions/:sessionId", middlewares.JWTAuthMiddleware(), handlers.GetDetailedInfo)
		UserGroup.GET("/sessions/search", middlewares.JWTAuthMiddleware(), handlers.SearchSessions)
	}
}
