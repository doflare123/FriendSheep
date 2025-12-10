package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

// func RoutesAuth(r *gin.Engine) {
// 	AuthGroup := r.Group("api/users")
// 	{
// 		AuthGroup.POST("/login", handlers.AuthUser)
// 		AuthGroup.POST("/refresh", handlers.RefreshTokenHandler)
// 		AuthGroup.POST("/request-reset", handlers.RequestPasswordReset)
// 		AuthGroup.POST("/confirm-reset", handlers.ConfirmPasswordReset)
// 	}
// }

func RegisterAuthRoutes(r *gin.Engine, authH handlers.AuthHandler) {
	auth := r.Group("api/v2/auth")
	{
		auth.POST("/login", authH.Login)
		auth.GET("/refresh", authH.RefreshToken)
	}
}

func RegisterRegRoutes(r *gin.Engine, regH handlers.RegHandler) {
	reg := r.Group("api/v2/register")
	{
		reg.POST("/session/register", regH.CreateSessionRegister)
		reg.PATCH("/session/verify", regH.VerifySession)
		reg.POST("/", regH.CreateUser)
		reg.POST("/password/change", regH.ChangePassword)
	}
}
