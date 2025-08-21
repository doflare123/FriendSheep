package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterGroups(r *gin.Engine) {
	GroupGroup := r.Group("/api/groups")
	{
		GroupGroup.POST("/createGroup", middlewares.JWTAuthMiddleware(), handlers.CreateGroup)
		GroupGroup.POST("/joinToGroup", middlewares.JWTAuthMiddleware(), handlers.JoinGroup)
		GroupGroup.GET("/requests/:groupId", middlewares.JWTAuthMiddleware(), handlers.GetJoinRequests)
		GroupGroup.DELETE("/:groupId/leave", middlewares.JWTAuthMiddleware(), handlers.LeaveGroupHandler)
		GroupGroup.GET("/:groupId", middlewares.JWTAuthMiddleware(), handlers.GetGroupInf)
		GroupGroup.GET("/search", middlewares.JWTAuthMiddleware(), handlers.SearchGroups)
	}

	GroupAdminGroups := r.Group("api/admin/groups")
	{
		GroupAdminGroups.GET("", middlewares.JWTAuthMiddleware(), handlers.GetAdminGroups)
		GroupAdminGroups.POST("/requestsForUser", middlewares.JWTAuthMiddleware(), handlers.SentJoinRequests)
		GroupAdminGroups.POST("/requests/:requestId/approve", middlewares.JWTAuthMiddleware(), handlers.ApproveJoinRequest)
		GroupAdminGroups.POST("/requests/:requestId/reject", middlewares.JWTAuthMiddleware(), handlers.RejectJoinRequest)
		GroupAdminGroups.POST("/UploadPhoto", middlewares.JWTAuthMiddleware(), handlers.ChangePhoto)

		// Роуты, требующие прав администратора или оператора
		moderatorRequired := GroupAdminGroups.Group("/:groupId", middlewares.JWTAuthMiddleware(), middlewares.GroupRoleMiddleware("admin", "operator"))
		{
			moderatorRequired.DELETE("/members/:userId", handlers.RemoveUserHandler)
			moderatorRequired.GET("/infGroup", handlers.GetInfAdminGroup)
		}

		// Роуты, требующие прав только администратора
		adminRequired := GroupAdminGroups.Group("/:groupId", middlewares.JWTAuthMiddleware(), middlewares.GroupRoleMiddleware("admin"))
		{
			adminRequired.DELETE("/", handlers.DeleteGroups)
			adminRequired.PATCH("/", handlers.UpdateGroupHandler)
		}
	}
}
