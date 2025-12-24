package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	"friendship/models/groups"
	"friendship/repository"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GroupRoleMiddleware struct {
	repo repository.PostgresRepository
}

func NewGroupRoleMiddleware(repo repository.PostgresRepository) *GroupRoleMiddleware {
	return &GroupRoleMiddleware{repo: repo}
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
		userID := userIDValue.(uint)

		// Получаем groupID из параметров запроса
		groupIDStr := c.Param("groupId")
		if groupIDStr == "" {
			groupIDStr = c.Query("groupId")
		}

		// Если не нашли в параметрах, читаем из body
		if groupIDStr == "" {
			// ✅ ПРАВИЛЬНО: Читаем body и восстанавливаем его
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Восстанавливаем body для последующего использования
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Парсим JSON для получения groupId
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

		// Проверяем членство пользователя в группе
		var groupUser groups.GroupUsers
		err = m.repo.Where("user_id = ? AND group_id = ?", userID, uint(groupID)).
			First(&groupUser).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
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

		// Получаем роль пользователя
		var role groups.Role_in_group
		err = m.repo.First(&role, groupUser.RoleInGroupID).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения роли пользователя",
			})
			c.Abort()
			return
		}

		// Проверяем, входит ли роль в список разрешенных
		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if role.Name == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Недостаточно прав для выполнения этого действия",
				"required_role": allowedRoles,
				"your_role":     role.Name,
			})
			c.Abort()
			return
		}

		// Сохраняем информацию о группе и роли в контекст
		c.Set("groupID", uint(groupID))
		c.Set("groupRole", role.Name)

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
