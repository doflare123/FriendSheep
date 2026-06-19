package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	groupmodels "friendship/models/groups"
	"friendship/repository"
	"friendship/utils"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupRoleMiddleware struct {
	reader groupRoleReader
}

type groupRoleRequirement struct {
	allows        groupmodels.RolePredicate
	requiredRoles []string
}

func NewGroupRoleMiddleware(repo repository.PostgresRepository) *GroupRoleMiddleware {
	return &GroupRoleMiddleware{reader: newRepositoryGroupRoleReader(repo)}
}

// Deprecated: use RequireGroupCapability or one of the capability-specific wrappers.
func (m *GroupRoleMiddleware) RequireGroupRole(allowedRoles ...string) gin.HandlerFunc {
	return m.requireGroupAccess(groupRoleRequirementForRoles(allowedRoles...))
}

func (m *GroupRoleMiddleware) requireGroupAccess(requirement groupRoleRequirement) gin.HandlerFunc {
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
			utils.AbortJSONError(c, http.StatusBadRequest, "Не указан ID группы")
			return
		}

		groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			utils.AbortJSONError(c, http.StatusBadRequest, "Некорректный ID группы")
			return
		}

		roleName, err := m.reader.FindUserGroupRole(userID, uint(groupID))
		if err != nil {
			abortRoleLookupError(c, err)
			return
		}

		if !requirement.allows(roleName) {
			abortForbiddenRole(c, roleName, requirement.requiredRoles)
			return
		}

		c.Set("groupID", uint(groupID))
		c.Set("groupRole", groupmodels.NormalizeRoleName(roleName))

		c.Next()
	}
}

// Deprecated: use RequireEventGroupCapability or one of the capability-specific wrappers.
func (m *GroupRoleMiddleware) RequireEventGroupRole(allowedRoles ...string) gin.HandlerFunc {
	return m.requireEventGroupAccess(groupRoleRequirementForRoles(allowedRoles...))
}

func (m *GroupRoleMiddleware) requireEventGroupAccess(requirement groupRoleRequirement) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := contextUserID(c)
		if !ok {
			abortUnauthorized(c)
			return
		}

		eventIDStr := c.Param("eventId")
		if eventIDStr == "" {
			utils.AbortJSONError(c, http.StatusBadRequest, "Не указан ID события")
			return
		}

		eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
		if err != nil {
			utils.AbortJSONError(c, http.StatusBadRequest, "Некорректный ID события")
			return
		}

		groupID, roleName, err := m.reader.FindUserEventGroupRole(userID, uint(eventID))
		if err != nil {
			abortRoleLookupError(c, err)
			return
		}

		if !requirement.allows(roleName) {
			abortForbiddenRole(c, roleName, requirement.requiredRoles)
			return
		}

		c.Set("eventID", uint(eventID))
		c.Set("groupID", groupID)
		c.Set("groupRole", groupmodels.NormalizeRoleName(roleName))

		c.Next()
	}
}

// Deprecated: use RequireJoinRequestGroupCapability or one of the capability-specific wrappers.
func (m *GroupRoleMiddleware) RequireJoinRequestGroupRole(allowedRoles ...string) gin.HandlerFunc {
	return m.requireJoinRequestGroupAccess(groupRoleRequirementForRoles(allowedRoles...))
}

func (m *GroupRoleMiddleware) requireJoinRequestGroupAccess(requirement groupRoleRequirement) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := contextUserID(c)
		if !ok {
			abortUnauthorized(c)
			return
		}

		requestIDStr := c.Param("requestId")
		if requestIDStr == "" {
			utils.AbortJSONError(c, http.StatusBadRequest, "Не указан ID заявки")
			return
		}

		requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
		if err != nil {
			utils.AbortJSONError(c, http.StatusBadRequest, "Некорректный ID заявки")
			return
		}

		groupID, roleName, err := m.reader.FindUserJoinRequestGroupRole(userID, uint(requestID))
		if err != nil {
			abortRoleLookupError(c, err)
			return
		}

		if !requirement.allows(roleName) {
			abortForbiddenRole(c, roleName, requirement.requiredRoles)
			return
		}

		c.Set("requestID", uint(requestID))
		c.Set("groupID", groupID)
		c.Set("groupRole", groupmodels.NormalizeRoleName(roleName))

		c.Next()
	}
}

func (m *GroupRoleMiddleware) RequireGroupCapability(required groupmodels.Capability) gin.HandlerFunc {
	return m.requireGroupAccess(groupRoleRequirementForCapability(required))
}

func (m *GroupRoleMiddleware) RequireEventGroupCapability(required groupmodels.Capability) gin.HandlerFunc {
	return m.requireEventGroupAccess(groupRoleRequirementForCapability(required))
}

func (m *GroupRoleMiddleware) RequireJoinRequestGroupCapability(required groupmodels.Capability) gin.HandlerFunc {
	return m.requireJoinRequestGroupAccess(groupRoleRequirementForCapability(required))
}

func (m *GroupRoleMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireGroupCapability(groupmodels.CapabilityAdmin)
}

func (m *GroupRoleMiddleware) RequireOperatorOrAdmin() gin.HandlerFunc {
	return m.RequireGroupCapability(groupmodels.CapabilityModerate)
}

func (m *GroupRoleMiddleware) RequireEventOperatorOrAdmin() gin.HandlerFunc {
	return m.RequireEventGroupCapability(groupmodels.CapabilityModerate)
}

func (m *GroupRoleMiddleware) RequireJoinRequestOperatorOrAdmin() gin.HandlerFunc {
	return m.RequireJoinRequestGroupCapability(groupmodels.CapabilityModerate)
}

func (m *GroupRoleMiddleware) RequireMember() gin.HandlerFunc {
	return m.RequireGroupCapability(groupmodels.CapabilityMember)
}

func groupRoleRequirementForRoles(allowedRoles ...string) groupRoleRequirement {
	requiredRoles := append([]string(nil), allowedRoles...)
	return groupRoleRequirement{
		allows: func(roleName string) bool {
			return groupmodels.HasAnyRole(roleName, requiredRoles...)
		},
		requiredRoles: requiredRoles,
	}
}

func groupRoleRequirementForCapability(required groupmodels.Capability) groupRoleRequirement {
	requiredRoles := groupmodels.RolesWithCapability(required)
	return groupRoleRequirement{
		allows:        groupmodels.PredicateForCapability(required),
		requiredRoles: requiredRoles,
	}
}

func contextUserID(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	return contextUint(userIDValue)
}

func abortUnauthorized(c *gin.Context) {
	utils.AbortJSONError(c, http.StatusUnauthorized, "Пользователь не авторизован")
}

func abortRoleLookupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errEventNotFound):
		utils.AbortJSONError(c, http.StatusNotFound, "Событие не найдено")
	case errors.Is(err, errJoinRequestNotFound):
		utils.AbortJSONError(c, http.StatusNotFound, "Заявка не найдена")
	case errors.Is(err, errGroupMembershipNotFound):
		utils.AbortJSONError(c, http.StatusForbidden, "Вы не являетесь участником этой группы")
	default:
		utils.AbortJSONError(c, http.StatusInternalServerError, "Ошибка проверки доступа")
	}
}

func abortForbiddenRole(c *gin.Context, roleName string, allowedRoles []string) {
	utils.AbortJSONError(
		c,
		http.StatusForbidden,
		"Недостаточно прав для выполнения этого действия",
		utils.WithRequiredRoles(allowedRoles),
		utils.WithYourRole(groupmodels.NormalizeRoleName(roleName)),
	)
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
