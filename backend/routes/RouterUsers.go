package routes

import (
	"friendship/handlers"
	"friendship/middlewares"
	"friendship/transport"

	"github.com/gin-gonic/gin"
)

func RouterUsersInfo(r *gin.Engine) {
	GetSessionGroup := r.Group("api/users")
	GetSessionGroup.Use(middlewares.JWTAuthMiddleware())
	{
		GetSessionGroup.GET("/sessions/new", handlers.GetNewSessions)
		GetSessionGroup.GET("/sessions/popular", handlers.GetPopularSessions)
		GetSessionGroup.GET("/sessions/category", handlers.GetCategorySessions)
		GetSessionGroup.GET("/sessions/:sessionId", handlers.GetDetailedInfo)
		GetSessionGroup.GET("/sessions/search", handlers.SearchSessions)
		GetSessionGroup.GET("/sessions/user-groups", handlers.GetSessionsUserGroups)
		GetSessionGroup.GET("/subscriptions", handlers.GetGroupsUserSub)
	}
	BotGroup := r.Group("api/bot")
	BotGroup.Use(middlewares.BotAuthMiddleware())
	{
		BotGroup.POST("/subscriptions", handlers.SubscriptionsTelegramNotify)
		BotGroup.POST("/unsubscriptions", handlers.UnSubscriptionsTelegramNotify)
	}
	InternalStatsGroup := r.Group("/internal")
	InternalStatsGroup.Use()
	{
		InternalStatsGroup.POST("/update-statistics", transport.UpdateStatisticsHandler)
	}
}
