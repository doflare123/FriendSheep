package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterReferencesRoutes(router *gin.Engine, referencesHandler handlers.ReferencesHandler) {
	referencesGroup := router.Group("/api/v2/references")
	{
		referencesGroup.GET("", referencesHandler.GetReferences)
		referencesGroup.GET("/genres", referencesHandler.SearchGenres)
	}
}
