package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	"friendship/repository"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupRoleMiddleware struct {
	reader groupRoleReader
}

func NewGroupRoleMiddleware(repo repository.PostgresRepository) *GroupRoleMiddleware {
	return &GroupRoleMiddleware{reader: newRepositoryGroupRoleReader(repo)}
}

func (m *GroupRoleMiddleware) RequireGroupRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Пользователь не авторизован",
			})
			c.Abort()
			return
		}

		userID, ok := contextUint(userIDValue)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Пользователь не авторизован",
			})
			c.Abort()
			return
		}

		groupIDStr := c.Param("groupId")
		if groupIDStr == "" {
			groupIDStr = c.Query("groupId")
		}

		if groupIDStr == "" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var body struct {
					GroupID uint `json:"groupId"`
				}
				if err := json.Unmarshal(bodyBytes, &body); err == nil && body.GroupID > 0 {
					groupIDStr = strconv.Itoa(int(body.GroupID))
				}
			}
		}

		if groupIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Не указан ID группы",
			})
			c.Abort()
			return
		}

		groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректный ID группы",
			})
			c.Abort()
			return
		}

		roleName, err := m.reader.FindUserGroupRole(userID, uint(groupID))
		if err != nil {
			if errors.Is(err, errGroupMembershipNotFound) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Вы не являетесь участником этой группы",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка проверки доступа",
				})
			}
			c.Abort()
			return
		}

		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if roleName == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Недостаточно прав для выполнения этого действия",
				"required_role": allowedRoles,
				"your_role":     roleName,
			})
			c.Abort()
			return
		}

		c.Set("groupID", uint(groupID))
		c.Set("groupRole", roleName)

		c.Next()
	}
}

func (m *GroupRoleMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireGroupRole("Админ")
}

func (m *GroupRoleMiddleware) RequireOperatorOrAdmin() gin.HandlerFunc {
	return m.RequireGroupRole("Админ", "Модератор")
}

func (m *GroupRoleMiddleware) RequireMember() gin.HandlerFunc {
	return m.RequireGroupRole("Админ", "Модератор", "Участник")
}

func contextUint(value interface{}) (uint, bool) {
	switch typed := value.(type) {
	case uint:
		return typed, true
	case uint8:
		return uint(typed), true
	case uint16:
		return uint(typed), true
	case uint32:
		return uint(typed), true
	case uint64:
		return uint(typed), true
	case int:
		if typed < 0 {
			return 0, false
		}
		return uint(typed), true
	case int8:
		if typed < 0 {
			return 0, false
		}
		return uint(typed), true
	case int16:
		if typed < 0 {
			return 0, false
		}
		return uint(typed), true
	case int32:
		if typed < 0 {
			return 0, false
		}
		return uint(typed), true
	case int64:
		if typed < 0 {
			return 0, false
		}
		return uint(typed), true
	default:
		return 0, false
	}
}
