package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"

	"gorm.io/gorm"
)

type joinRequestReviewStore interface {
	groupActorRoleFinder

	FindJoinRequest(requestID uint) (joinRequestReview, error)
	FindActor(actorID uint) (joinRequestActor, error)
	IsUserBlacklisted(groupID uint, userID uint) (bool, error)
	CreateRequestUserMembership(groupID uint, userID uint) error
	UpdateJoinRequestStatus(requestID uint, status string) error
	CreateActionLog(groupID uint, actor joinRequestActor, role string, action string, description string) error
}

type joinRequestReview struct {
	ID       uint
	UserID   uint
	GroupID  uint
	Status   string
	UserName string
	UserUs   string
}

type joinRequestActor struct {
	ID   uint
	Name string
	Us   string
}

type gormJoinRequestReviewStore struct {
	tx groupTx
	txGroupAccessStore
}

func newJoinRequestReviewStore(tx groupTx) joinRequestReviewStore {
	return gormJoinRequestReviewStore{
		tx:                 tx,
		txGroupAccessStore: newTxGroupAccessStore(tx),
	}
}

func (s gormJoinRequestReviewStore) FindJoinRequest(requestID uint) (joinRequestReview, error) {
	var request groups.GroupJoinRequest
	if err := s.tx.Preload("User").First(&request, requestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return joinRequestReview{}, ErrJoinRequestNotFound
		}
		return joinRequestReview{}, fmt.Errorf("ошибка поиска заявки: %w", err)
	}

	return joinRequestReview{
		ID:       request.ID,
		UserID:   request.UserID,
		GroupID:  request.GroupID,
		Status:   request.Status,
		UserName: request.User.Name,
		UserUs:   request.User.Us,
	}, nil
}

func (s gormJoinRequestReviewStore) FindActor(actorID uint) (joinRequestActor, error) {
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

func (s gormJoinRequestReviewStore) IsUserBlacklisted(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupBlacklist{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки черного списка: %w", err)
	}

	return count > 0, nil
}

func (s gormJoinRequestReviewStore) CreateRequestUserMembership(groupID uint, userID uint) error {
	memberRoleID, err := findGroupRoleID(s.tx, groups.RoleMember)
	if err != nil {
		return ErrRoleMemberNotFound
	}

	groupUser := groups.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: memberRoleID,
	}
	if err := s.tx.Create(&groupUser).Error; err != nil {
		if isGroupMembershipUniqueViolation(err) {
			return ErrAlreadyInGroup
		}
		return fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
	}

	return nil
}

func (s gormJoinRequestReviewStore) UpdateJoinRequestStatus(requestID uint, status string) error {
	result := s.tx.Model(&groups.GroupJoinRequest{}).
		Where("id = ?", requestID).
		Where("status = ?", "pending").
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("ошибка обновления статуса заявки: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrJoinRequestHandled
	}

	return nil
}

func (s gormJoinRequestReviewStore) CreateActionLog(groupID uint, actor joinRequestActor, role string, action string, description string) error {
	return createGroupActionLog(s.tx, groupID, actor, role, action, description)
}
