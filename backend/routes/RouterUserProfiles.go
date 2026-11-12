package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterUserProfileRoutes(router *gin.Engine, handler *handlers.UserProfileHandler, jwtAuth gin.HandlerFunc) {
	group := router.Group("/api/v2/users")
	group.Use(jwtAuth)
	group.GET("", handler.GetOwnProfile)
	group.GET("/:identifier", handler.GetProfile)
	group.GET("/me/privacy", handler.GetPrivacy)
	group.PATCH("/me/privacy", handler.PatchPrivacy)
	group.GET("/me/statistics/tiles", handler.GetTileSettings)
	group.PUT("/me/statistics/tiles", handler.ReplaceTileSettings)
}
