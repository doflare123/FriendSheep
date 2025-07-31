package middlewares

import (
	"fmt"
	"friendship/db"
	"friendship/models/groups"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GroupRoleMiddleware проверяет, имеет ли пользователь одну из указанных ролей в группе.
// Он извлекает groupID из URL и email из JWT.
func GroupRoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("email")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
			return
		}

		groupIDStr := c.Param("groupId")
		if groupIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ID группы не указан в URL"})
			return
		}

		groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "некорректный ID группы"})
			return
		}

		var userID uint
		if err := db.GetDB().Table("users").Select("id").Where("email = ?", email).Scan(&userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "пользователь не найден"})
			return
		}

		var groupUser groups.GroupUsers
		if err := db.GetDB().Where("user_id = ? AND group_id = ? AND role_in_group IN ?", userID, groupID, roles).First(&groupUser).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("у вас нет необходимых прав в этой группе (%v)", roles)})
			return
		}

		c.Next()
	}
}
