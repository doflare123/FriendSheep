package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterUsersInfo(r *gin.Engine) {
	GetSessionGroup := r.Group("api/users")
	{
		GetSessionGroup.GET("/sessions/new", middlewares.JWTAuthMiddleware(), handlers.GetNewSessions)
		GetSessionGroup.GET("/sessions/popular", middlewares.JWTAuthMiddleware(), handlers.GetPopularSessions)
		GetSessionGroup.GET("/sessions/category", middlewares.JWTAuthMiddleware(), handlers.GetCategorySessions)
		GetSessionGroup.GET("/sessions/:sessionId", middlewares.JWTAuthMiddleware(), handlers.GetDetailedInfo)
		GetSessionGroup.GET("/sessions/search", middlewares.JWTAuthMiddleware(), handlers.SearchSessions)
		GetSessionGroup.GET("/sessions/user-groups", middlewares.JWTAuthMiddleware(), handlers.GetSessionsUserGroups)
		GetSessionGroup.GET("/subscriptions", middlewares.JWTAuthMiddleware(), handlers.GetGroupsUserSub)
	}
}
