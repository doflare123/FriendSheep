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
		GroupGroup.POST("/requests/:requestId/approve", middlewares.JWTAuthMiddleware(), handlers.ApproveJoinRequest)
		GroupGroup.POST("/requests/:requestId/reject", middlewares.JWTAuthMiddleware(), handlers.RejectJoinRequest)
		GroupGroup.DELETE("/:groupId/members/:userId", middlewares.JWTAuthMiddleware(), handlers.RemoveUserHandler)
		GroupGroup.DELETE("/:groupId/leave", middlewares.JWTAuthMiddleware(), handlers.LeaveGroupHandler)
		GroupGroup.DELETE("/:groupId", middlewares.JWTAuthMiddleware(), handlers.DeleteGroups)
	}
}
