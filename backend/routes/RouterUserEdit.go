package routes

import (
	"friendship/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterUserEditRoutes(router *gin.Engine, handler *handlers.UserEditHandler, jwtAuth gin.HandlerFunc) {
	group := router.Group("/api/v2/users/me/profile")
	group.Use(jwtAuth)
	group.GET("", handler.GetOwn)
	group.PATCH("", handler.Patch)
	group.PUT("/image", handler.UploadImage)
	group.DELETE("/image", handler.DeleteImage)
}
