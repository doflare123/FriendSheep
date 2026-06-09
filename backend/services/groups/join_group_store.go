package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"

	"gorm.io/gorm"
)

type joinGroupStore interface {
	EnsureUserExists(userID uint) error
	FindJoinGroupTarget(groupID uint) (joinGroupTarget, error)
	IsUserBlacklisted(groupID uint, userID uint) (bool, error)
	IsGroupMember(groupID uint, userID uint) (bool, error)
	HasPendingJoinRequest(groupID uint, userID uint) (bool, error)
	CreatePendingJoinRequest(groupID uint, userID uint) error
	CreateGroupMember(groupID uint, userID uint) error
}

type joinGroupTarget struct {
	ID        uint
	IsPrivate bool
}

type gormJoinGroupStore struct {
	tx groupTx
}

func newJoinGroupStore(tx groupTx) joinGroupStore {
	return gormJoinGroupStore{tx: tx}
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

func (s gormJoinGroupStore) IsUserBlacklisted(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupBlacklist{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки черного списка: %w", err)
	}

	return count > 0, nil
}

func (s gormJoinGroupStore) IsGroupMember(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства: %w", err)
	}

	return count > 0, nil
}

func (s gormJoinGroupStore) HasPendingJoinRequest(groupID uint, userID uint) (bool, error) {
	var request groups.GroupJoinRequest
	err := s.tx.Where("user_id = ? AND group_id = ? AND status = ?", userID, groupID, "pending").
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
		Status:  "pending",
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
