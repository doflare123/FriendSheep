package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterUserGroupsRoutes(
	router *gin.Engine,
	groupHandler handlers.GroupHandler,
	authMiddleware *middlewares.AuthMiddleware,
) {
	userGroups := router.Group("/api/v2/users/me/groups")
	userGroups.Use(authMiddleware.RequireAuth())
	{
		userGroups.GET("/managed", groupHandler.GetManagedGroups)
	}
}
