package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RoutesNews(r *gin.Engine) {
	NewsGroup := r.Group("api/news")
	{
		NewsGroup.POST("", middlewares.JWTAuthMiddleware(), handlers.CreateNews)
		NewsGroup.GET("", handlers.GetNews)
		NewsGroup.GET("/:id", handlers.GetNewsByID)
		NewsGroup.POST("/:id/comments", middlewares.JWTAuthMiddleware(), handlers.AddComment)
		NewsGroup.DELETE("/:newsId/comments/:commentId", middlewares.JWTAuthMiddleware(), handlers.DeleteComment)
	}
}
