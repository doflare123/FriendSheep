package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"time"

	"gorm.io/gorm"
)

// GetJoinRequests получает все заявки на вступление в группу
func (s *groupService) GetJoinRequests(actorID uint, groupID uint, status string, limit int) ([]JoinRequestInfo, error) {
	// Проверяем права доступа (admin или operator)
	hasAccess, _, err := s.checkGroupAccess(actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := s.post.
		Preload("User").
		Where("group_id = ?", groupID)

	// Фильтр по статусу (опционально)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var requests []groups.GroupJoinRequest
	err = query.
		Order("created_at DESC").
		Limit(limit).
		Find(&requests).Error

	if err != nil {
		s.logger.Error("Не удалось получить заявки на вступление", "groupID", groupID, "error", err)
		return nil, fmt.Errorf("ошибка получения заявок: %w", err)
	}

	result := make([]JoinRequestInfo, 0, len(requests))
	for _, req := range requests {
		result = append(result, JoinRequestInfo{
			ID:        req.ID,
			UserID:    req.UserID,
			Name:      req.User.Name,
			Us:        req.User.Us,
			Image:     req.User.Image,
			GroupID:   req.GroupID,
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
		})
	}

	return result, nil
}

// CreateJoinInvite создает приглашение в группу
func (s *groupService) CreateJoinInvite(actorID uint, input JoinInviteInput) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, groups.CapabilityModerate)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User

	err = s.runInTx(func(tx groupTx) error {
		if err := tx.First(&targetUser, input.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}
		// Проверяем, состоит ли уже в группе
		var existingCount int64
		if err := tx.Model(&groups.GroupUsers{}).
			Where("user_id = ? AND group_id = ?", input.UserID, input.GroupID).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("ошибка проверки членства: %w", err)
		}
		if existingCount > 0 {
			return ErrAlreadyInGroup
		}

		// Проверяем существующие приглашения
		var existingInvite groups.GroupJoinInvite
		err := tx.Where("user_id = ? AND group_id = ? AND status = ?", input.UserID, input.GroupID, "pending").
			First(&existingInvite).Error
		if err == nil {
			return ErrInviteAlreadyExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ошибка проверки приглашений: %w", err)
		}

		// Создаем приглашение
		invite := groups.GroupJoinInvite{
			UserID:    input.UserID,
			GroupID:   input.GroupID,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&invite).Error; err != nil {
			return fmt.Errorf("ошибка создания приглашения: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Отправил приглашение пользователю '%s' (@%s)", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, input.GroupID, actorID, actor.Name, actor.Us, role, "send_invite", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось создать приглашение", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	s.logger.Info("Приглашение в группу создано", "actorID", actorID, "targetUserID", input.UserID, "groupID", input.GroupID)
	return true, nil
}

// ApproveAllJoinRequests одобряет все заявки
func (s *groupService) ApproveAllJoinRequests(actorID uint, groupID uint) (int, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return 0, err
	}
	if !hasAccess {
		return 0, ErrPermissionDenied
	}

	var requests []groups.GroupJoinRequest
	var actor models.User
	count := 0

	err = s.runInTx(func(tx groupTx) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.Where("group_id = ? AND status = ?", groupID, "pending").
			Preload("User").
			Find(&requests).Error; err != nil {
			return fmt.Errorf("ошибка получения заявок: %w", err)
		}

		memberRoleID, err := findGroupRoleID(tx, groups.RoleMember)
		if err != nil {
			return ErrRoleMemberNotFound
		}

		for _, req := range requests {
			// Проверяем черный список
			var blacklistCount int64
			if err := tx.Model(&groups.GroupBlacklist{}).
				Where("group_id = ? AND user_id = ?", groupID, req.UserID).
				Count(&blacklistCount).Error; err != nil {
				return fmt.Errorf("ошибка проверки черного списка: %w", err)
			}
			if blacklistCount > 0 {
				continue
			}

			// Добавляем в группу
			groupUser := groups.GroupUsers{
				UserID:        req.UserID,
				GroupID:       groupID,
				RoleInGroupID: memberRoleID,
			}
			if err := tx.Create(&groupUser).Error; err != nil {
				s.logger.Warn("Не удалось добавить пользователя в группу", "userID", req.UserID, "error", err)
				continue
			}

			// Обновляем статус заявки
			if err := tx.Model(&req).Update("status", "approved").Error; err != nil {
				s.logger.Warn("Не удалось обновить статус заявки", "requestID", req.ID, "error", err)
			}

			count++

			// Логируем действие
			action := fmt.Sprintf("Одобрил заявку пользователя '%s' (@%s)", req.User.Name, req.User.Us)
			if err := s.logAction(tx, groupID, actorID, actor.Name, actor.Us, role, "approve_request", action); err != nil {
				s.logger.Warn("Не удалось записать действие в журнал", "error", err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось одобрить все заявки", "groupID", groupID, "error", err)
		return 0, err
	}

	s.logger.Info("Все заявки на вступление одобрены", "groupID", groupID, "count", count, "actorID", actorID)
	return count, nil
}

// RejectAllJoinRequests отклоняет все заявки
func (s *groupService) RejectAllJoinRequests(actorID uint, groupID uint) (int, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return 0, err
	}
	if !hasAccess {
		return 0, ErrPermissionDenied
	}

	var actor models.User
	count := 0

	err = s.runInTx(func(tx groupTx) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		result := tx.Model(&groups.GroupJoinRequest{}).
			Where("group_id = ? AND status = ?", groupID, "pending").
			Update("status", "rejected")

		if result.Error != nil {
			return fmt.Errorf("ошибка отклонения заявок: %w", result.Error)
		}

		count = int(result.RowsAffected)

		// Логируем действие
		action := fmt.Sprintf("Отклонил все заявки на вступление (%d шт.)", count)
		if err := s.logAction(tx, groupID, actorID, actor.Name, actor.Us, role, "reject_all_requests", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось отклонить все заявки", "groupID", groupID, "error", err)
		return 0, err
	}

	s.logger.Info("Все заявки на вступление отклонены", "groupID", groupID, "count", count, "actorID", actorID)
	return count, nil
}

// ApproveJoinRequest одобряет конкретную заявку
func (s *groupService) ApproveJoinRequest(actorID uint, requestID uint) (bool, error) {
	var request groups.GroupJoinRequest
	var actor models.User

	err := s.runInTx(func(tx groupTx) error {
		if err := tx.Preload("User").First(&request, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJoinRequestNotFound
			}
			return fmt.Errorf("ошибка поиска заявки: %w", err)
		}

		hasAccess, role, err := s.checkGroupAccess(actorID, request.GroupID, groups.CapabilityModerate)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrPermissionDenied
		}

		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if request.Status != "pending" {
			return ErrJoinRequestHandled
		}

		// Проверяем черный список
		var blacklistCount int64
		if err := tx.Model(&groups.GroupBlacklist{}).
			Where("group_id = ? AND user_id = ?", request.GroupID, request.UserID).
			Count(&blacklistCount).Error; err != nil {
			return fmt.Errorf("ошибка проверки черного списка: %w", err)
		}
		if blacklistCount > 0 {
			return ErrUserInBlacklist
		}

		memberRoleID, err := findGroupRoleID(tx, groups.RoleMember)
		if err != nil {
			return ErrRoleMemberNotFound
		}

		groupUser := groups.GroupUsers{
			UserID:        request.UserID,
			GroupID:       request.GroupID,
			RoleInGroupID: memberRoleID,
		}
		if err := tx.Create(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
		}

		if err := tx.Model(&request).Update("status", "approved").Error; err != nil {
			return fmt.Errorf("ошибка обновления статуса заявки: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Одобрил заявку пользователя '%s' (@%s)", request.User.Name, request.User.Us)
		if err := s.logAction(tx, request.GroupID, actorID, actor.Name, actor.Us, role, "approve_request", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось одобрить заявку", "requestID", requestID, "error", err)
		return false, err
	}

	s.logger.Info("Заявка на вступление одобрена", "requestID", requestID, "actorID", actorID)
	return true, nil
}

// RejectJoinRequest отклоняет конкретную заявку
func (s *groupService) RejectJoinRequest(actorID uint, requestID uint) (bool, error) {
	var request groups.GroupJoinRequest
	var actor models.User

	err := s.runInTx(func(tx groupTx) error {
		if err := tx.Preload("User").First(&request, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJoinRequestNotFound
			}
			return fmt.Errorf("ошибка поиска заявки: %w", err)
		}

		hasAccess, role, err := s.checkGroupAccess(actorID, request.GroupID, groups.CapabilityModerate)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrPermissionDenied
		}

		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if request.Status != "pending" {
			return ErrJoinRequestHandled
		}

		if err := tx.Model(&request).Update("status", "rejected").Error; err != nil {
			return fmt.Errorf("ошибка обновления статуса заявки: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Отклонил заявку пользователя '%s' (@%s)", request.User.Name, request.User.Us)
		if err := s.logAction(tx, request.GroupID, actorID, actor.Name, actor.Us, role, "reject_request", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось отклонить заявку", "requestID", requestID, "error", err)
		return false, err
	}

	s.logger.Info("Заявка на вступление отклонена", "requestID", requestID, "actorID", actorID)
	return true, nil
}
