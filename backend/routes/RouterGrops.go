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

	}
	GroupAdminGroups := r.Group("api/admin/groups")
	{
		GroupAdminGroups.POST("/requests/:requestId/approve", middlewares.JWTAuthMiddleware(), handlers.ApproveJoinRequest)
		GroupAdminGroups.POST("/requests/:requestId/reject", middlewares.JWTAuthMiddleware(), handlers.RejectJoinRequest)
		GroupAdminGroups.DELETE("/:groupId/members/:userId", middlewares.JWTAuthMiddleware(), handlers.RemoveUserHandler)
		GroupAdminGroups.DELETE("/:groupId", middlewares.JWTAuthMiddleware(), handlers.DeleteGroups)
		GroupAdminGroups.PATCH("/:groupId", middlewares.JWTAuthMiddleware(), handlers.UpdateGroupHandler)
	}
}
