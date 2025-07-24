package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// func RateLimitMiddleware() gin.HandlerFunc {
// 	// Создаем лимитер: 10 запросов в секунду с буфером до 20 запросов
// 	limiter := rate.NewLimiter(rate.Limit(10), 20)

// 	return func(c *gin.Context) {
// 		if !limiter.Allow() {
// 			c.JSON(http.StatusTooManyRequests, gin.H{
// 				"error": "слишком много запросов, повторите позже",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }

func ValidatePageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if pageStr := c.Query("page"); pageStr != "" {
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 || page > 1000 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "неверный номер страницы (должен быть от 1 до 1000)",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
