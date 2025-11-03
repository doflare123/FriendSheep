package routes

import (
	"friendship/handlers"
	"friendship/middlewares"
	"friendship/transport"

	"github.com/gin-gonic/gin"
)

func RouterUsersInfo(r *gin.Engine) {
	GetSessionGroup := r.Group("api/users")
	GetSessionGroup.Use()
	{
		GetSessionGroup.GET("/search", middlewares.JWTAuthMiddleware(), handlers.SearchUsers)
		// GetSessionGroup.GET("/sessions/new", handlers.GetNewSessions)
		GetSessionGroup.GET("/sessions/popular", handlers.GetPopularSessions)
		// GetSessionGroup.GET("/sessions/category", handlers.GetCategorySessions)
		GetSessionGroup.GET("/sessions/:sessionId", middlewares.JWTAuthMiddleware(), handlers.GetDetailedInfo)
		GetSessionGroup.GET("/sessions/search", middlewares.JWTAuthMiddleware(), handlers.SearchSessions)
		GetSessionGroup.GET("/sessions/user-groups", middlewares.JWTAuthMiddleware(), handlers.GetSessionsUserGroups)
		GetSessionGroup.GET("/subscriptions", middlewares.JWTAuthMiddleware(), handlers.GetGroupsUserSub)
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
	UserInfGroup := r.Group("api/users")
	UserInfGroup.Use(middlewares.JWTAuthMiddleware())
	{
		UserInfGroup.GET("/inf", handlers.GetInfAboutUser)
		UserInfGroup.GET("/inf/:id", handlers.GetInfAboutUserByID)
		UserInfGroup.GET("/notify", handlers.GetNotify)
		UserInfGroup.GET("/notify/inf", handlers.GetNotifyInf)
		UserInfGroup.POST("/notifications/viewed", handlers.MarkNotificationViewed)
		UserInfGroup.PUT("/invites/:id/approve", handlers.ApproveInvite)
		UserInfGroup.PUT("/invites/:id/reject", handlers.RejectInvite)
		UserInfGroup.PATCH("/user/profile", handlers.UpdateUserProfile)
		UserInfGroup.PATCH("/password", handlers.ChangePassword)
		UserInfGroup.PATCH("/tiles", handlers.ChangeTilesPattern)
		UserInfGroup.DELETE("/delete", handlers.DeleteAccount)
		UserInfGroup.GET("/:us", handlers.GettingUserId)
	}
}
