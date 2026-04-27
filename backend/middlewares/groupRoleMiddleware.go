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
		userID, ok := contextUserID(c)
		if !ok {
			abortUnauthorized(c)
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID группы"})
			c.Abort()
			return
		}

		groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID группы"})
			c.Abort()
			return
		}

		roleName, err := m.reader.FindUserGroupRole(userID, uint(groupID))
		if err != nil {
			abortRoleLookupError(c, err)
			return
		}

		if !roleAllowed(roleName, allowedRoles) {
			abortForbiddenRole(c, roleName, allowedRoles)
			return
		}

		c.Set("groupID", uint(groupID))
		c.Set("groupRole", roleName)

		c.Next()
	}
}

func (m *GroupRoleMiddleware) RequireEventGroupRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := contextUserID(c)
		if !ok {
			abortUnauthorized(c)
			return
		}

		eventIDStr := c.Param("eventId")
		if eventIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID события"})
			c.Abort()
			return
		}

		eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID события"})
			c.Abort()
			return
		}

		groupID, roleName, err := m.reader.FindUserEventGroupRole(userID, uint(eventID))
		if err != nil {
			abortRoleLookupError(c, err)
			return
		}

		if !roleAllowed(roleName, allowedRoles) {
			abortForbiddenRole(c, roleName, allowedRoles)
			return
		}

		c.Set("eventID", uint(eventID))
		c.Set("groupID", groupID)
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

func (m *GroupRoleMiddleware) RequireEventOperatorOrAdmin() gin.HandlerFunc {
	return m.RequireEventGroupRole("Админ", "Модератор")
}

func (m *GroupRoleMiddleware) RequireMember() gin.HandlerFunc {
	return m.RequireGroupRole("Админ", "Модератор", "Участник")
}

func contextUserID(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	return contextUint(userIDValue)
}

func abortUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
	c.Abort()
}

func abortRoleLookupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errEventNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Событие не найдено"})
	case errors.Is(err, errGroupMembershipNotFound):
		c.JSON(http.StatusForbidden, gin.H{"error": "Вы не являетесь участником этой группы"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки доступа"})
	}
	c.Abort()
}

func abortForbiddenRole(c *gin.Context, roleName string, allowedRoles []string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error":         "Недостаточно прав для выполнения этого действия",
		"required_role": allowedRoles,
		"your_role":     roleName,
	})
	c.Abort()
}

func roleAllowed(roleName string, allowedRoles []string) bool {
	for _, allowedRole := range allowedRoles {
		if roleName == allowedRole {
			return true
		}
	}
	return false
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
