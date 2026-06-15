package group

import (
	"fmt"
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

	requests, err := s.reads.ListJoinRequests(groupID, status, limit)
	if err != nil {
		s.logger.Error("Не удалось получить заявки на вступление", "groupID", groupID, "error", err)
		return nil, err
	}

	return requests, nil
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

		actor, err := store.FindActor(actorID)
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

	var result bulkJoinRequestResult

	err = s.runInTx(func(tx groupTx) error {
		store := newBulkJoinRequestStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var approveErr error
		result, approveErr = store.ApproveAllPending(groupID, actor, role)
		return approveErr
	})

	if err != nil {
		s.logger.Error("Не удалось одобрить все заявки", "groupID", groupID, "error", err)
		return 0, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Все заявки на вступление одобрены", "groupID", groupID, "count", result.Count, "actorID", actorID)
	return result.Count, nil
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

	var result bulkJoinRequestResult

	err = s.runInTx(func(tx groupTx) error {
		store := newBulkJoinRequestStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var rejectErr error
		result, rejectErr = store.RejectAllPending(groupID, actor, role)
		return rejectErr
	})

	if err != nil {
		s.logger.Error("Не удалось отклонить все заявки", "groupID", groupID, "error", err)
		return 0, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Все заявки на вступление отклонены", "groupID", groupID, "count", result.Count, "actorID", actorID)
	return result.Count, nil
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

		if request.Status != groups.JoinStatusPending {
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

		if err := store.UpdateJoinRequestStatus(request.ID, groups.JoinStatusApproved); err != nil {
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

		if request.Status != groups.JoinStatusPending {
			return ErrJoinRequestHandled
		}

		if err := store.UpdateJoinRequestStatus(request.ID, groups.JoinStatusRejected); err != nil {
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
