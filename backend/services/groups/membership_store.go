package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

type joinGroupStore interface {
	groupRelationChecks

	EnsureUserExists(userID uint) error
	FindJoinGroupTarget(groupID uint) (joinGroupTarget, error)
	HasPendingJoinRequest(groupID uint, userID uint) (bool, error)
	CreatePendingJoinRequest(groupID uint, userID uint) error
	CreateGroupMember(groupID uint, userID uint) error
}

type joinGroupTarget struct {
	ID        uint
	IsPrivate bool
}

type gormJoinGroupStore struct {
	tx repository.PostgresRepository
	txGroupRelationStore
}

func newJoinGroupStore(tx repository.PostgresRepository) joinGroupStore {
	return gormJoinGroupStore{
		tx:                   tx,
		txGroupRelationStore: newTxGroupRelationStore(tx),
	}
}

func (s gormJoinGroupStore) EnsureUserExists(userID uint) error {
	var user models.User
	if err := s.tx.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return nil
}

func (s gormJoinGroupStore) FindJoinGroupTarget(groupID uint) (joinGroupTarget, error) {
	var group groups.Group
	if err := s.tx.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return joinGroupTarget{}, ErrGroupNotFound
		}
		return joinGroupTarget{}, fmt.Errorf("ошибка поиска группы: %w", err)
	}

	return joinGroupTarget{
		ID:        group.ID,
		IsPrivate: group.IsPrivate,
	}, nil
}

func (s gormJoinGroupStore) HasPendingJoinRequest(groupID uint, userID uint) (bool, error) {
	var request groups.GroupJoinRequest
	err := s.tx.Where("user_id = ? AND group_id = ? AND status = ?", userID, groupID, groups.JoinStatusPending).
		First(&request).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, fmt.Errorf("ошибка проверки заявок: %w", err)
}

func (s gormJoinGroupStore) CreatePendingJoinRequest(groupID uint, userID uint) error {
	request := groups.GroupJoinRequest{
		UserID:  userID,
		GroupID: groupID,
		Status:  groups.JoinStatusPending,
	}
	if err := s.tx.Create(&request).Error; err != nil {
		if isPendingJoinRequestUniqueViolation(err) {
			return ErrRequestAlreadyExists
		}
		return fmt.Errorf("ошибка создания заявки: %w", err)
	}

	return nil
}

func (s gormJoinGroupStore) CreateGroupMember(groupID uint, userID uint) error {
	roleID, err := findGroupRoleID(s.tx, groups.RoleMember)
	if err != nil {
		return ErrRoleMemberNotFound
	}

	member := groups.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: roleID,
	}
	if err := s.tx.Create(&member).Error; err != nil {
		if isGroupMembershipUniqueViolation(err) {
			return ErrAlreadyInGroup
		}
		return fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
	}

	return nil
}

type leaveGroupStore interface {
	LeaveGroup(userID uint, groupID uint) (string, error)
}

type gormLeaveGroupStore struct {
	tx repository.PostgresRepository
}

func newLeaveGroupStore(tx repository.PostgresRepository) leaveGroupStore {
	return gormLeaveGroupStore{tx: tx}
}

func (s gormLeaveGroupStore) LeaveGroup(userID uint, groupID uint) (string, error) {
	var groupUser groups.GroupUsers
	err := s.tx.Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&groupUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotInGroup
		}
		return "", fmt.Errorf("ошибка поиска участника: %w", err)
	}

	roleName := groups.RoleMember
	var role groups.Role_in_group
	if err := s.tx.First(&role, groupUser.RoleInGroupID).Error; err == nil {
		roleName = groups.NormalizeRoleName(role.Name)
		if groups.HasCapability(role.Name, groups.CapabilityAdmin) {
			return "", fmt.Errorf("администратор не может покинуть группу. Передайте права другому участнику или удалите группу")
		}
	}

	if err := s.tx.Delete(&groupUser).Error; err != nil {
		return "", fmt.Errorf("ошибка удаления участника: %w", err)
	}

	return roleName, nil
}
