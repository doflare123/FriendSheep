package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

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

func RegisterAuthRoutes(r *gin.Engine, authH handlers.AuthHandler, authMiddleware ...*middlewares.AuthMiddleware) {
	auth := r.Group("api/v2/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.RefreshToken)
	}

	if len(authMiddleware) == 0 || authMiddleware[0] == nil {
		return
	}

	protectedAuth := r.Group("api/v2/auth")
	protectedAuth.Use(authMiddleware[0].RequireAuth())
	{
		protectedAuth.POST("/logout", authH.Logout)
		protectedAuth.POST("/logout-all", authH.LogoutAll)
	}

	user := r.Group("api/v2/user")
	user.Use(authMiddleware[0].RequireAuth())
	{
		user.GET("/me", authH.Me)
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
