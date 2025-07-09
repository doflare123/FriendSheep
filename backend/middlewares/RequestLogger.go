package middlewares

import (
	"friendship/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		utils.Log.Printf("INFO: %s %s [%d] %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}
