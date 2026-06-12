package group

import (
	"fmt"
	"friendship/models"
	"friendship/models/groups"
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
	err := s.runInTx(func(tx groupTx) error {
		store := newJoinInviteCreationStore(tx)

		hasAccess, role, err := store.FindActorRole(actorID, input.GroupID, groups.CapabilityModerate)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrPermissionDenied
		}

		targetUser, err := store.FindInviteUser(input.UserID)
		if err != nil {
			return err
		}

		actor, err := store.FindInviteActor(actorID)
		if err != nil {
			return err
		}

		isMember, err := store.IsGroupMember(input.GroupID, input.UserID)
		if err != nil {
			return err
		}
		if isMember {
			return ErrAlreadyInGroup
		}

		hasPendingInvite, err := store.HasPendingInvite(input.GroupID, input.UserID)
		if err != nil {
			return err
		}
		if hasPendingInvite {
			return ErrInviteAlreadyExists
		}

		if err := store.CreatePendingInvite(input.GroupID, input.UserID); err != nil {
			return err
		}

		// Логируем действие
		action := fmt.Sprintf("Отправил приглашение пользователю '%s' (@%s)", targetUser.Name, targetUser.Us)
		if err := store.CreateActionLog(input.GroupID, actor, role, "send_invite", action); err != nil {
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
	err := s.runInTx(func(tx groupTx) error {
		store := newJoinRequestReviewStore(tx)

		request, err := store.FindJoinRequest(requestID)
		if err != nil {
			return err
		}

		hasAccess, role, err := store.FindActorRole(actorID, request.GroupID, groups.CapabilityModerate)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrPermissionDenied
		}

		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		if request.Status != "pending" {
			return ErrJoinRequestHandled
		}

		isBlacklisted, err := store.IsUserBlacklisted(request.GroupID, request.UserID)
		if err != nil {
			return err
		}
		if isBlacklisted {
			return ErrUserInBlacklist
		}

		if err := store.CreateRequestUserMembership(request.GroupID, request.UserID); err != nil {
			return err
		}

		if err := store.UpdateJoinRequestStatus(request.ID, "approved"); err != nil {
			return err
		}

		// Логируем действие
		action := fmt.Sprintf("Одобрил заявку пользователя '%s' (@%s)", request.UserName, request.UserUs)
		if err := store.CreateActionLog(request.GroupID, actor, role, "approve_request", action); err != nil {
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
	err := s.runInTx(func(tx groupTx) error {
		store := newJoinRequestReviewStore(tx)

		request, err := store.FindJoinRequest(requestID)
		if err != nil {
			return err
		}

		hasAccess, role, err := store.FindActorRole(actorID, request.GroupID, groups.CapabilityModerate)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrPermissionDenied
		}

		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		if request.Status != "pending" {
			return ErrJoinRequestHandled
		}

		if err := store.UpdateJoinRequestStatus(request.ID, "rejected"); err != nil {
			return err
		}

		// Логируем действие
		action := fmt.Sprintf("Отклонил заявку пользователя '%s' (@%s)", request.UserName, request.UserUs)
		if err := store.CreateActionLog(request.GroupID, actor, role, "reject_request", action); err != nil {
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
