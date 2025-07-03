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
		// GroupSession.POST("/joinToGroup", middlewares.JWTAuthMiddleware(), handlers.JoinGroup)
		// GroupSession.GET("/requests", middlewares.JWTAuthMiddleware(), handlers.GetJoinRequests)
		// GroupSession.POST("/requests/:requestId/approve", middlewares.JWTAuthMiddleware(), handlers.ApproveJoinRequest)
		// GroupSession.POST("/requests/:requestId/reject", middlewares.JWTAuthMiddleware(), handlers.RejectJoinRequest)
		// GroupSession.DELETE("/:groupId/members/:userId", middlewares.JWTAuthMiddleware(), handlers.RemoveUserHandler)
		// GroupSession.DELETE("/:groupId/leave", middlewares.JWTAuthMiddleware(), handlers.LeaveGroupHandler)
		// GroupSession.DELETE("/:groupId", middlewares.JWTAuthMiddleware(), handlers.DeleteGroups)
	}
}
