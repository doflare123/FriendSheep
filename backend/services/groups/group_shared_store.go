package group

import (
	"errors"
	"fmt"
	"friendship/models"
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

type groupActorFinder interface {
	FindActor(actorID uint) (joinRequestActor, error)
}

type txGroupActorStore struct {
	tx groupTx
}

type groupRelationChecks interface {
	IsUserBlacklisted(groupID uint, userID uint) (bool, error)
	IsGroupMember(groupID uint, userID uint) (bool, error)
}

type txGroupRelationStore struct {
	tx groupTx
}

func newTxGroupAccessStore(tx groupTx) txGroupAccessStore {
	return newGroupAccessStore(tx)
}

func newGroupAccessStore(store groupRoleStore) txGroupAccessStore {
	return txGroupAccessStore{lookup: gormGroupRoleLookup{store: store}}
}

func newTxGroupActorStore(tx groupTx) txGroupActorStore {
	return txGroupActorStore{tx: tx}
}

func newTxGroupRelationStore(tx groupTx) txGroupRelationStore {
	return txGroupRelationStore{tx: tx}
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

func (s txGroupActorStore) FindActor(actorID uint) (joinRequestActor, error) {
	var actor models.User
	if err := s.tx.First(&actor, actorID).Error; err != nil {
		return joinRequestActor{}, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return joinRequestActor{
		ID:   actor.ID,
		Name: actor.Name,
		Us:   actor.Us,
	}, nil
}

func (s txGroupRelationStore) IsUserBlacklisted(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupBlacklist{}).
		Where(&groups.GroupBlacklist{GroupID: groupID, UserID: userID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки черного списка: %w", err)
	}

	return count > 0, nil
}

func (s txGroupRelationStore) IsGroupMember(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupUsers{}).
		Where(&groups.GroupUsers{UserID: userID, GroupID: groupID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства: %w", err)
	}

	return count > 0, nil
}

func createGroupActionLog(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, description string) error {
	return createGroupActionLogRecord(tx, groupID, actor, role, action, description, nil, nil, "")
}

func createGroupTargetUserActionLog(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, targetUserID uint, description string) error {
	return createGroupActionLogRecord(tx, groupID, actor, role, action, description, &targetUserID, nil, "")
}

func createGroupEntityActionLog(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, entityID uint, entityName string, description string) error {
	return createGroupActionLogRecord(tx, groupID, actor, role, action, description, nil, &entityID, entityName)
}

func createGroupTargetUserEntityActionLog(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, targetUserID uint, entityID uint, entityName string, description string) error {
	return createGroupActionLogRecord(tx, groupID, actor, role, action, description, &targetUserID, &entityID, entityName)
}

func createGroupActionLogRecord(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, description string, targetUserID *uint, entityID *uint, entityName string) error {
	actionTypeID, err := groups.FindGroupActionTypeID(tx, action)
	if err != nil {
		return fmt.Errorf("тип действия группы %q не найден: %w", action, err)
	}

	actionLog := groups.GroupActionLog{
		GroupID:      groupID,
		UserID:       actor.ID,
		Username:     actor.Name,
		Us:           actor.Us,
		Role:         role,
		ActionTypeID: actionTypeID,
		Description:  description,
		TargetUserID: targetUserID,
		EntityID:     entityID,
		EntityName:   entityName,
		CreatedAt:    time.Now(),
	}

	return tx.Create(&actionLog).Error
}
