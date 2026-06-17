package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"

	"gorm.io/gorm"
)

type groupAdminStore interface {
	groupActorFinder

	DeleteGroup(groupID uint) error
	ChangeMemberRole(input GroupUserInput, roleName string, actor joinRequestActor, actorRole string) (groupAdminResult, error)
}

type groupAdminResult struct {
	Warnings []groupAdminLogWarning
}

type groupAdminLogWarning struct {
	Message string
	Args    []interface{}
}

type gormGroupAdminStore struct {
	tx groupTx
	txGroupActorStore
}

func newGroupAdminStore(tx groupTx) groupAdminStore {
	return gormGroupAdminStore{
		tx:                tx,
		txGroupActorStore: newTxGroupActorStore(tx),
	}
}

func (s gormGroupAdminStore) DeleteGroup(groupID uint) error {
	var group groups.Group
	if err := s.tx.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupNotFound
		}
		return fmt.Errorf("ошибка поиска группы: %w", err)
	}

	if err := s.tx.Delete(&group).Error; err != nil {
		return fmt.Errorf("ошибка удаления группы: %w", err)
	}

	return nil
}

func (s gormGroupAdminStore) ChangeMemberRole(input GroupUserInput, roleName string, actor joinRequestActor, actorRole string) (groupAdminResult, error) {
	targetUser, err := s.findTargetUser(input.UserID)
	if err != nil {
		return groupAdminResult{}, err
	}

	groupUser, err := s.findGroupMember(input.GroupID, input.UserID)
	if err != nil {
		return groupAdminResult{}, err
	}

	roleID, err := findGroupRoleID(s.tx, roleName)
	if err != nil {
		return groupAdminResult{}, roleNotFoundError(roleName)
	}

	if err := s.tx.Model(&groupUser).Update("role_in_group_id", roleID).Error; err != nil {
		return groupAdminResult{}, fmt.Errorf("ошибка обновления роли: %w", err)
	}

	result := groupAdminResult{}
	actionType := roleChangeAction(roleName)
	if err := createGroupTargetUserActionLog(s.tx, input.GroupID, actor, actorRole, actionType, targetUser.ID, ""); err != nil {
		result.Warnings = append(result.Warnings, groupAdminLogWarning{
			Message: "Роль участника изменена, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return result, nil
}

func (s gormGroupAdminStore) findTargetUser(userID uint) (models.User, error) {
	var targetUser models.User
	if err := s.tx.First(&targetUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return targetUser, nil
}

func (s gormGroupAdminStore) findGroupMember(groupID uint, userID uint) (groups.GroupUsers, error) {
	var groupUser groups.GroupUsers
	err := s.tx.Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&groupUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return groups.GroupUsers{}, ErrNotInGroup
		}
		return groups.GroupUsers{}, fmt.Errorf("ошибка поиска участника: %w", err)
	}

	return groupUser, nil
}

func roleNotFoundError(roleName string) error {
	switch roleName {
	case groups.RoleModerator:
		return ErrRoleModeratorNotFound
	case groups.RoleMember:
		return ErrRoleMemberNotFound
	case groups.RoleAdmin:
		return ErrRoleAdminNotFound
	default:
		return fmt.Errorf("роль %q не найдена", roleName)
	}
}

func roleChangeAction(roleName string) string {
	switch roleName {
	case groups.RoleModerator:
		return groups.ActionAddOperator
	case groups.RoleMember:
		return groups.ActionRemoveOperator
	default:
		return groups.ActionChangeRole
	}
}
