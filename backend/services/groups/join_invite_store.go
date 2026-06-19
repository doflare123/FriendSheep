package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type joinInviteCreationStore interface {
	groupActorRoleFinder
	groupActorFinder
	groupRelationChecks

	FindInviteUser(userID uint) (joinInviteUser, error)
	HasPendingInvite(groupID uint, userID uint) (bool, error)
	CreatePendingInvite(groupID uint, userID uint) error
	CreateActionLog(input groupActionLogInput) error
}

type joinInviteUser struct {
	ID   uint
	Name string
	Us   string
}

type gormJoinInviteCreationStore struct {
	tx groupTx
	txGroupAccessStore
	txGroupActorStore
	txGroupRelationStore
}

func newJoinInviteCreationStore(tx groupTx) joinInviteCreationStore {
	return gormJoinInviteCreationStore{
		tx:                   tx,
		txGroupAccessStore:   newTxGroupAccessStore(tx),
		txGroupActorStore:    newTxGroupActorStore(tx),
		txGroupRelationStore: newTxGroupRelationStore(tx),
	}
}

func (s gormJoinInviteCreationStore) FindInviteUser(userID uint) (joinInviteUser, error) {
	var user models.User
	if err := s.tx.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return joinInviteUser{}, ErrUserNotFound
		}
		return joinInviteUser{}, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return joinInviteUser{
		ID:   user.ID,
		Name: user.Name,
		Us:   user.Us,
	}, nil
}

func (s gormJoinInviteCreationStore) HasPendingInvite(groupID uint, userID uint) (bool, error) {
	var invite groups.GroupJoinInvite
	err := s.tx.Where("user_id = ? AND group_id = ? AND status = ?", userID, groupID, groups.JoinStatusPending).
		First(&invite).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, fmt.Errorf("ошибка проверки приглашений: %w", err)
}

func (s gormJoinInviteCreationStore) CreatePendingInvite(groupID uint, userID uint) error {
	invite := groups.GroupJoinInvite{
		UserID:    userID,
		GroupID:   groupID,
		Status:    groups.JoinStatusPending,
		CreatedAt: time.Now(),
	}
	if err := s.tx.Create(&invite).Error; err != nil {
		return fmt.Errorf("ошибка создания приглашения: %w", err)
	}

	return nil
}

func (s gormJoinInviteCreationStore) CreateActionLog(input groupActionLogInput) error {
	return createGroupActionLog(s.tx, input)
}

type joinInviteResponseStore interface {
	groupRelationChecks

	FindJoinInvite(inviteID uint) (joinInviteResponse, error)
	CreateGroupMemberIfMissing(groupID uint, userID uint) error
	UpdateInviteStatus(inviteID uint, status string) error
}

type joinInviteResponse struct {
	ID      uint
	UserID  uint
	GroupID uint
	Status  string
}

type gormJoinInviteResponseStore struct {
	tx groupTx
	txGroupRelationStore
}

func newJoinInviteResponseStore(tx groupTx) joinInviteResponseStore {
	return gormJoinInviteResponseStore{
		tx:                   tx,
		txGroupRelationStore: newTxGroupRelationStore(tx),
	}
}

func (s gormJoinInviteResponseStore) FindJoinInvite(inviteID uint) (joinInviteResponse, error) {
	var invite groups.GroupJoinInvite
	if err := s.tx.First(&invite, inviteID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return joinInviteResponse{}, ErrInviteNotFound
		}
		return joinInviteResponse{}, fmt.Errorf("ошибка поиска приглашения: %w", err)
	}

	return joinInviteResponse{
		ID:      invite.ID,
		UserID:  invite.UserID,
		GroupID: invite.GroupID,
		Status:  invite.Status,
	}, nil
}

func (s gormJoinInviteResponseStore) CreateGroupMemberIfMissing(groupID uint, userID uint) error {
	memberRoleID, err := findGroupRoleID(s.tx, groups.RoleMember)
	if err != nil {
		return ErrRoleMemberNotFound
	}

	groupUser := groups.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: memberRoleID,
	}
	createMembership := s.tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&groupUser)
	if createMembership.Error != nil {
		return fmt.Errorf("ошибка добавления пользователя в группу: %w", createMembership.Error)
	}
	if createMembership.RowsAffected > 0 {
		return nil
	}

	isMember, err := s.IsGroupMember(groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("членство пользователя в группе не подтверждено после принятия приглашения")
	}

	return nil
}

func (s gormJoinInviteResponseStore) UpdateInviteStatus(inviteID uint, status string) error {
	result := s.tx.Model(&groups.GroupJoinInvite{}).
		Where(&groups.GroupJoinInvite{ID: inviteID}).
		Where("status = ?", groups.JoinStatusPending).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("ошибка обновления статуса приглашения: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInviteAlreadyHandled
	}

	return nil
}
