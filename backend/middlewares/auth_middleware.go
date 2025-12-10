package middlewares

import (
	"friendship/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtUtils *utils.JWTUtils
}

func NewAuthMiddleware(jwtService *utils.JWTUtils) *AuthMiddleware {
	return &AuthMiddleware{jwtUtils: jwtService}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "отсутствует токен авторизации",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "неверный формат токена",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := m.jwtUtils.ParseAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "невалидный токен",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("us", claims.Us)
		c.Set("image", claims.Image)

		c.Next()
	}
}

// Вроде должно работать. Когда можно давать смотреть что-то челу, который не авторизован
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := m.jwtUtils.ParseAccessToken(parts[1])
		if err == nil {
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("us", claims.Us)
			c.Set("image", claims.Image)
		}

		c.Next()
	}
}
