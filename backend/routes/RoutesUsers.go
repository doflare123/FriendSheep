package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RoutesUsers(r *gin.Engine) {
	UserGroup := r.Group("api/users")
	{
		UserGroup.GET("/", handlers.GetAllUsers)
		UserGroup.GET("/:id", handlers.GetUserByID)
		UserGroup.POST("/register", handlers.CreateUserHandler)
		UserGroup.PUT("/:id", handlers.UpdateUser)
	}
}
