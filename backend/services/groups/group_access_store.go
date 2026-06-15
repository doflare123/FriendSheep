package group

import (
	"errors"
	"fmt"
	"friendship/models/groups"
	"time"

	"gorm.io/gorm"
)

type txGroupAccessStore struct {
	lookup groupRoleLookup
}

type groupRoleLookup interface {
	FindMembershipRoleID(actorID uint, groupID uint) (uint, error)
	FindRoleName(roleID uint) (string, error)
}

type gormGroupRoleLookup struct {
	store groupRoleStore
}

type groupActorRoleFinder interface {
	FindActorRole(actorID uint, groupID uint, required groups.Capability) (bool, string, error)
}

func newTxGroupAccessStore(tx groupTx) txGroupAccessStore {
	return newGroupAccessStore(tx)
}

func newGroupAccessStore(store groupRoleStore) txGroupAccessStore {
	return txGroupAccessStore{lookup: gormGroupRoleLookup{store: store}}
}

func (s txGroupAccessStore) FindActorRole(actorID uint, groupID uint, required groups.Capability) (bool, string, error) {
	roleName, err := s.findActorRoleName(actorID, groupID)
	if err != nil {
		return false, "", err
	}

	normalizedRole := groups.NormalizeRoleName(roleName)
	if groups.HasCapability(roleName, required) {
		return true, normalizedRole, nil
	}

	return false, normalizedRole, nil
}

func (s txGroupAccessStore) findActorRoleName(actorID uint, groupID uint) (string, error) {
	roleID, err := s.lookup.FindMembershipRoleID(actorID, groupID)
	if err != nil {
		return "", err
	}

	return s.lookup.FindRoleName(roleID)
}

func (s gormGroupRoleLookup) FindMembershipRoleID(actorID uint, groupID uint) (uint, error) {
	var groupUser groups.GroupUsers
	err := s.store.Where(&groups.GroupUsers{
		UserID:  actorID,
		GroupID: groupID,
	}).Take(&groupUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrNotInGroup
		}
		return 0, fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	return groupUser.RoleInGroupID, nil
}

func (s gormGroupRoleLookup) FindRoleName(roleID uint) (string, error) {
	var role groups.Role_in_group
	if err := s.store.Where(&groups.Role_in_group{Id: roleID}).Take(&role).Error; err != nil {
		return "", fmt.Errorf("ошибка получения роли: %w", err)
	}

	return role.Name, nil
}

func createGroupActionLog(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, description string) error {
	actionLog := groups.GroupActionLog{
		GroupID:     groupID,
		UserID:      actor.ID,
		Username:    actor.Name,
		Us:          actor.Us,
		Role:        role,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return tx.Create(&actionLog).Error
}
