package group

import (
	"errors"
	"fmt"
	"friendship/models/groups"

	"gorm.io/gorm"
)

type joinRequestReviewStore interface {
	groupActorRoleFinder
	groupActorFinder
	groupRelationChecks

	FindJoinRequest(requestID uint) (joinRequestReview, error)
	CreateRequestUserMembership(groupID uint, userID uint) error
	UpdateJoinRequestStatus(requestID uint, status string) error
	CreateActionLog(input groupActionLogInput) error
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
	tx groupPersistence
	txGroupAccessStore
	txGroupActorStore
	txGroupRelationStore
}

func newJoinRequestReviewStore(tx groupPersistence) joinRequestReviewStore {
	return gormJoinRequestReviewStore{
		tx:                   tx,
		txGroupAccessStore:   newTxGroupAccessStore(tx),
		txGroupActorStore:    newTxGroupActorStore(tx),
		txGroupRelationStore: newTxGroupRelationStore(tx),
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
		Where("status = ?", groups.JoinStatusPending).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("ошибка обновления статуса заявки: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrJoinRequestHandled
	}

	return nil
}

func (s gormJoinRequestReviewStore) CreateActionLog(input groupActionLogInput) error {
	return createGroupActionLog(s.tx, input)
}

type bulkJoinRequestStore interface {
	groupActorFinder

	ApproveAllPending(groupID uint, actor joinRequestActor, role string) (bulkJoinRequestResult, error)
	RejectAllPending(groupID uint, actor joinRequestActor, role string) (bulkJoinRequestResult, error)
}

type bulkJoinRequestResult struct {
	Count    int
	Warnings []bulkJoinRequestLogWarning
}

type bulkJoinRequestLogWarning struct {
	Message string
	Args    []interface{}
}

type gormBulkJoinRequestStore struct {
	tx groupPersistence
	txGroupActorStore
}

func newBulkJoinRequestStore(tx groupPersistence) bulkJoinRequestStore {
	return gormBulkJoinRequestStore{
		tx:                tx,
		txGroupActorStore: newTxGroupActorStore(tx),
	}
}

func (s gormBulkJoinRequestStore) ApproveAllPending(groupID uint, actor joinRequestActor, role string) (bulkJoinRequestResult, error) {
	var requests []groups.GroupJoinRequest
	if err := s.tx.Where("group_id = ? AND status = ?", groupID, groups.JoinStatusPending).
		Preload("User").
		Find(&requests).Error; err != nil {
		return bulkJoinRequestResult{}, fmt.Errorf("ошибка получения заявок: %w", err)
	}

	memberRoleID, err := findGroupRoleID(s.tx, groups.RoleMember)
	if err != nil {
		return bulkJoinRequestResult{}, ErrRoleMemberNotFound
	}

	result := bulkJoinRequestResult{}
	for _, req := range requests {
		isBlacklisted, err := s.isUserBlacklisted(groupID, req.UserID)
		if err != nil {
			return bulkJoinRequestResult{}, err
		}
		if isBlacklisted {
			continue
		}

		groupUser := groups.GroupUsers{
			UserID:        req.UserID,
			GroupID:       groupID,
			RoleInGroupID: memberRoleID,
		}
		if err := s.tx.Create(&groupUser).Error; err != nil {
			result.Warnings = append(result.Warnings, bulkJoinRequestLogWarning{
				Message: "Заявка не одобрена: не удалось добавить пользователя в группу",
				Args:    []interface{}{"userID", req.UserID, "error", err},
			})
			continue
		}

		if err := s.tx.Model(&req).Update("status", groups.JoinStatusApproved).Error; err != nil {
			result.Warnings = append(result.Warnings, bulkJoinRequestLogWarning{
				Message: "Пользователь добавлен в группу, но статус заявки не обновлен",
				Args:    []interface{}{"requestID", req.ID, "error", err},
			})
		}

		result.Count++

		targetUserID := req.UserID
		if err := createGroupActionLog(s.tx, groupActionLogInput{
			GroupID:      groupID,
			Actor:        actor,
			Role:         role,
			Action:       groups.ActionApproveRequest,
			TargetUserID: &targetUserID,
		}); err != nil {
			result.Warnings = append(result.Warnings, bulkJoinRequestLogWarning{
				Message: "Заявка обработана, но действие не записано в журнал группы",
				Args:    []interface{}{"error", err},
			})
		}
	}

	action := fmt.Sprintf("Одобрил все ожидающие заявки (%d шт.)", result.Count)
	if err := createGroupActionLog(s.tx, groupActionLogInput{
		GroupID:     groupID,
		Actor:       actor,
		Role:        role,
		Action:      groups.ActionApproveAllRequests,
		Description: action,
	}); err != nil {
		result.Warnings = append(result.Warnings, bulkJoinRequestLogWarning{
			Message: "Заявки обработаны, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return result, nil
}

func (s gormBulkJoinRequestStore) RejectAllPending(groupID uint, actor joinRequestActor, role string) (bulkJoinRequestResult, error) {
	updateResult := s.tx.Model(&groups.GroupJoinRequest{}).
		Where("group_id = ? AND status = ?", groupID, groups.JoinStatusPending).
		Update("status", groups.JoinStatusRejected)
	if updateResult.Error != nil {
		return bulkJoinRequestResult{}, fmt.Errorf("ошибка отклонения заявок: %w", updateResult.Error)
	}

	result := bulkJoinRequestResult{Count: int(updateResult.RowsAffected)}
	action := fmt.Sprintf("Отклонил все заявки на вступление (%d шт.)", result.Count)
	if err := createGroupActionLog(s.tx, groupActionLogInput{
		GroupID:     groupID,
		Actor:       actor,
		Role:        role,
		Action:      groups.ActionRejectAllRequests,
		Description: action,
	}); err != nil {
		result.Warnings = append(result.Warnings, bulkJoinRequestLogWarning{
			Message: "Заявки отклонены, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return result, nil
}

func (s gormBulkJoinRequestStore) isUserBlacklisted(groupID uint, userID uint) (bool, error) {
	var blacklistCount int64
	if err := s.tx.Model(&groups.GroupBlacklist{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&blacklistCount).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки черного списка: %w", err)
	}

	return blacklistCount > 0, nil
}
