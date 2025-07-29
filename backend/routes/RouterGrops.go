package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterGroups(r *gin.Engine) {
	GroupGroup := r.Group("api/groups")
	{
		GroupGroup.POST("/createGroup", middlewares.JWTAuthMiddleware(), handlers.CreateGroup)
		GroupGroup.POST("/joinToGroup", middlewares.JWTAuthMiddleware(), handlers.JoinGroup)
		GroupGroup.GET("/requests", middlewares.JWTAuthMiddleware(), handlers.GetJoinRequests)
		GroupGroup.DELETE("/:groupId/leave", middlewares.JWTAuthMiddleware(), handlers.LeaveGroupHandler)
		GroupGroup.GET("/:groupId", middlewares.JWTAuthMiddleware(), handlers.GetGroupInf)
	}
	GroupAdminGroups := r.Group("api/admin/groups")
	{
		GroupAdminGroups.GET("/", middlewares.JWTAuthMiddleware(), handlers.GetAdminGroups)
		GroupAdminGroups.POST("/requests/:requestId/approve", middlewares.JWTAuthMiddleware(), handlers.ApproveJoinRequest)
		GroupAdminGroups.POST("/requests/:requestId/reject", middlewares.JWTAuthMiddleware(), handlers.RejectJoinRequest)

		// Роуты, требующие прав администратора или оператора
		moderatorRequired := GroupAdminGroups.Group("/:groupId", middlewares.JWTAuthMiddleware(), middlewares.GroupRoleMiddleware("admin", "operator"))
		{
			moderatorRequired.DELETE("/members/:userId", handlers.RemoveUserHandler)
		}

		// Роуты, требующие прав только администратора
		adminRequired := GroupAdminGroups.Group("/:groupId", middlewares.JWTAuthMiddleware(), middlewares.GroupRoleMiddleware("admin"))
		{
			adminRequired.DELETE("/", handlers.DeleteGroups)
			adminRequired.PATCH("/", handlers.UpdateGroupHandler)
		}
	}
}
