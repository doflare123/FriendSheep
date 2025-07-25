package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RoutesAuth(r *gin.Engine) {
	AuthGroup := r.Group("api/users")
	{
		AuthGroup.POST("/login", handlers.AuthUser)
		AuthGroup.POST("/refresh", handlers.RefreshTokenHandler)
	}
}
