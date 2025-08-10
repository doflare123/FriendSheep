package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func BotAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" || apiKey != os.Getenv("TELEGRAM_BOT_API_KEY") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ты че тут забыл? Это линия для ботов"})
			c.Abort()
			return
		}
		c.Next()
	}
}
