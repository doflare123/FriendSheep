package groups

import (
	"errors"
	"friendship/repository"
	"strings"

	"gorm.io/gorm"
)

type Role_in_group struct {
	Id   uint
	Name string `gorm:"uniqueIndex;not null"`
}

const (
	RoleAdmin     = "Админ"
	RoleModerator = "Модератор"
	RoleMember    = "Участник"
)

type Capability uint8

const (
	CapabilityMember Capability = iota + 1
	CapabilityModerate
	CapabilityAdmin
)

var roleCapabilities = map[string]map[Capability]struct{}{
	RoleAdmin: {
		CapabilityAdmin:    {},
		CapabilityModerate: {},
		CapabilityMember:   {},
	},
	RoleModerator: {
		CapabilityModerate: {},
		CapabilityMember:   {},
	},
	RoleMember: {
		CapabilityMember: {},
	},
}

var rolesByAccessLevel = []string{
	RoleAdmin,
	RoleModerator,
	RoleMember,
}

type RolePredicate func(roleName string) bool

func NormalizeRoleName(roleName string) string {
	trimmed := strings.TrimSpace(roleName)

	switch {
	case strings.EqualFold(trimmed, RoleAdmin):
		return RoleAdmin
	case strings.EqualFold(trimmed, RoleModerator):
		return RoleModerator
	case strings.EqualFold(trimmed, RoleMember):
		return RoleMember
	default:
		return trimmed
	}
}

func CapabilityOf(roleName string) (Capability, bool) {
	capabilities, ok := roleCapabilities[NormalizeRoleName(roleName)]
	if !ok {
		return 0, false
	}

	for _, capability := range []Capability{CapabilityAdmin, CapabilityModerate, CapabilityMember} {
		if _, ok := capabilities[capability]; ok {
			return capability, true
		}
	}

	return 0, false
}

func HasCapability(roleName string, required Capability) bool {
	capabilities, ok := roleCapabilities[NormalizeRoleName(roleName)]
	if !ok {
		return false
	}

	_, ok = capabilities[required]
	return ok
}

func RolesWithCapability(required Capability) []string {
	roles := make([]string, 0, len(rolesByAccessLevel))
	for _, roleName := range rolesByAccessLevel {
		if HasCapability(roleName, required) {
			roles = append(roles, roleName)
		}
	}
	return roles
}

func HasAnyRole(roleName string, allowedRoles ...string) bool {
	normalizedRole := NormalizeRoleName(roleName)

	for _, allowedRole := range allowedRoles {
		if normalizedRole == NormalizeRoleName(allowedRole) {
			return true
		}
	}

	return false
}

func PredicateForCapability(required Capability) RolePredicate {
	return func(roleName string) bool {
		return HasCapability(roleName, required)
	}
}

func (r *Role_in_group) GetIdRole(str string, post repository.PostgresRepository) uint {
	var role Role_in_group

	err := post.Where("name = ?", str).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0
		}
		return 0
	}

	return role.Id
}
