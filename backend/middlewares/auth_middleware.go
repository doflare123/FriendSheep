package middlewares

import (
	"friendship/utils"
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
			utils.AbortJSONError(c, 401, "отсутствует токен авторизации")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.AbortJSONError(c, 401, "неверный формат токена")
			return
		}

		tokenString := parts[1]

		claims, err := m.jwtUtils.ParseAccessToken(tokenString)
		if err != nil {
			utils.AbortJSONError(c, 401, "невалидный токен")
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
