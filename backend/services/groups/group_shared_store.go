package group

import (
	"context"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"friendship/repository"
	"time"
)

type rootGroupActorRoleFinder interface {
	FindActorRole(ctx context.Context, actorID uint, groupID uint, required groups.Capability) (bool, string, error)
}

type rootGroupAccessStore struct {
	lookup rootGroupRoleLookup
}

type rootGroupRoleLookup interface {
	FindMembershipRoleID(ctx context.Context, actorID uint, groupID uint) (uint, error)
	FindRoleName(ctx context.Context, roleID uint) (string, error)
}

type gormRootGroupRoleLookup struct {
	store repository.PostgresRepository
}

type txGroupAccessStore struct {
	lookup groupRoleLookup
}

type groupRoleLookup interface {
	FindMembershipRoleID(actorID uint, groupID uint) (uint, error)
	FindRoleName(roleID uint) (string, error)
}

type gormGroupRoleLookup struct {
	store repository.PostgresRepository
}

type groupActorRoleFinder interface {
	FindActorRole(actorID uint, groupID uint, required groups.Capability) (bool, string, error)
}

type groupActorFinder interface {
	FindActor(actorID uint) (joinRequestActor, error)
}

type txGroupActorStore struct {
	tx repository.PostgresRepository
}

type groupRelationChecks interface {
	IsUserBlacklisted(groupID uint, userID uint) (bool, error)
	IsGroupMember(groupID uint, userID uint) (bool, error)
}

type txGroupRelationStore struct {
	tx repository.PostgresRepository
}

func newTxGroupAccessStore(tx repository.PostgresRepository) txGroupAccessStore {
	return newGroupAccessStore(tx)
}

func newGroupAccessStore(store repository.PostgresRepository) txGroupAccessStore {
	return txGroupAccessStore{lookup: gormGroupRoleLookup{store: store}}
}

func newRootGroupAccessStore(store repository.PostgresRepository) rootGroupAccessStore {
	return rootGroupAccessStore{lookup: gormRootGroupRoleLookup{store: store}}
}

func newTxGroupActorStore(tx repository.PostgresRepository) txGroupActorStore {
	return txGroupActorStore{tx: tx}
}

func newTxGroupRelationStore(tx repository.PostgresRepository) txGroupRelationStore {
	return txGroupRelationStore{tx: tx}
}

func (s rootGroupAccessStore) FindActorRole(ctx context.Context, actorID uint, groupID uint, required groups.Capability) (bool, string, error) {
	if ctx == nil {
		return false, "", errGroupOperationContextMissing
	}

	roleID, err := s.lookup.FindMembershipRoleID(ctx, actorID, groupID)
	if err != nil {
		return false, "", err
	}

	roleName, err := s.lookup.FindRoleName(ctx, roleID)
	if err != nil {
		return false, "", err
	}

	normalizedRole := groups.NormalizeRoleName(roleName)
	if groups.HasCapability(roleName, required) {
		return true, normalizedRole, nil
	}

	return false, normalizedRole, nil
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

func (s gormRootGroupRoleLookup) FindMembershipRoleID(ctx context.Context, actorID uint, groupID uint) (uint, error) {
	var groupUser groups.GroupUsers
	err := s.store.Where(&groups.GroupUsers{
		UserID:  actorID,
		GroupID: groupID,
	}).WithContext(ctx).Take(&groupUser).Error
	if err != nil {
		if isGroupRecordNotFound(err) {
			return 0, ErrNotInGroup
		}
		return 0, fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	return groupUser.RoleInGroupID, nil
}

func (s gormRootGroupRoleLookup) FindRoleName(ctx context.Context, roleID uint) (string, error) {
	var role groups.Role_in_group
	if err := s.store.Where(&groups.Role_in_group{Id: roleID}).WithContext(ctx).Take(&role).Error; err != nil {
		return "", fmt.Errorf("ошибка получения роли: %w", err)
	}

	return role.Name, nil
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
		if isGroupRecordNotFound(err) {
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

type groupActionLogInput struct {
	GroupID      uint
	Actor        joinRequestActor
	Role         string
	Action       string
	Description  string
	TargetUserID *uint
	EntityID     *uint
	EntityName   string
}

func createGroupActionLog(tx repository.PostgresRepository, input groupActionLogInput) error {
	actionTypeID, err := findGroupActionTypeID(tx, input.Action)
	if err != nil {
		return fmt.Errorf("тип действия группы %q не найден: %w", input.Action, err)
	}

	actionLog := groups.GroupActionLog{
		GroupID:      input.GroupID,
		UserID:       input.Actor.ID,
		Username:     input.Actor.Name,
		Us:           input.Actor.Us,
		Role:         input.Role,
		ActionTypeID: actionTypeID,
		Description:  input.Description,
		TargetUserID: input.TargetUserID,
		EntityID:     input.EntityID,
		EntityName:   input.EntityName,
		CreatedAt:    time.Now(),
	}

	return tx.Create(&actionLog).Error
}
