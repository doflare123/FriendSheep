package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterSubRoutes(r *gin.Engine, subH handlers.SubHandler, authMiddleware *middlewares.AuthMiddleware) {
	sub := r.Group("api/v2/sub")
	sub.Use(authMiddleware.RequireAuth())
	{
		sub.POST("/UploadImg", subH.ChangePhoto)
	}
}
