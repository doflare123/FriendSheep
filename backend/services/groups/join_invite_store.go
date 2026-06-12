package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"time"

	"gorm.io/gorm"
)

type joinInviteCreationStore interface {
	groupActorRoleFinder

	FindInviteUser(userID uint) (joinInviteUser, error)
	FindInviteActor(actorID uint) (joinRequestActor, error)
	IsGroupMember(groupID uint, userID uint) (bool, error)
	HasPendingInvite(groupID uint, userID uint) (bool, error)
	CreatePendingInvite(groupID uint, userID uint) error
	CreateActionLog(groupID uint, actor joinRequestActor, role string, action string, description string) error
}

type joinInviteUser struct {
	ID   uint
	Name string
	Us   string
}

type gormJoinInviteCreationStore struct {
	tx groupTx
	txGroupAccessStore
}

func newJoinInviteCreationStore(tx groupTx) joinInviteCreationStore {
	return gormJoinInviteCreationStore{
		tx:                 tx,
		txGroupAccessStore: newTxGroupAccessStore(tx),
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

func (s gormJoinInviteCreationStore) FindInviteActor(actorID uint) (joinRequestActor, error) {
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

func (s gormJoinInviteCreationStore) IsGroupMember(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства: %w", err)
	}

	return count > 0, nil
}

func (s gormJoinInviteCreationStore) HasPendingInvite(groupID uint, userID uint) (bool, error) {
	var invite groups.GroupJoinInvite
	err := s.tx.Where("user_id = ? AND group_id = ? AND status = ?", userID, groupID, "pending").
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
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	if err := s.tx.Create(&invite).Error; err != nil {
		return fmt.Errorf("ошибка создания приглашения: %w", err)
	}

	return nil
}

func (s gormJoinInviteCreationStore) CreateActionLog(groupID uint, actor joinRequestActor, role string, action string, description string) error {
	return createGroupActionLog(s.tx, groupID, actor, role, action, description)
}
