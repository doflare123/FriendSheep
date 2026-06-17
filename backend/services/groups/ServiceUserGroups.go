package group

import (
	"errors"
	"friendship/models/groups"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// JoinGroup вступление в группу
func (s *groupService) JoinGroup(userID uint, groupID uint) (*GroupResult, error) {
	var target joinGroupTarget

	err := s.runInTx(func(tx groupTx) error {
		store := newJoinGroupStore(tx)

		if err := store.EnsureUserExists(userID); err != nil {
			return err
		}

		foundTarget, err := store.FindJoinGroupTarget(groupID)
		if err != nil {
			return err
		}
		target = foundTarget

		isBlacklisted, err := store.IsUserBlacklisted(groupID, userID)
		if err != nil {
			return err
		}
		if isBlacklisted {
			return ErrUserInBlacklist
		}

		isMember, err := store.IsGroupMember(groupID, userID)
		if err != nil {
			return err
		}
		if isMember {
			return ErrAlreadyInGroup
		}

		hasPendingRequest, err := store.HasPendingJoinRequest(groupID, userID)
		if err != nil {
			return err
		}
		if hasPendingRequest {
			return ErrRequestAlreadyExists
		}

		if target.IsPrivate {
			if err := store.CreatePendingJoinRequest(groupID, userID); err != nil {
				return err
			}
			if err := s.logActionWithTargetUser(tx, groupID, userID, "", "", groups.RoleMember, groups.ActionCreateJoinRequest, "", userID); err != nil {
				s.logger.Warn("Не удалось записать действие в журнал", "error", err)
			}
			return nil
		}

		if err := store.CreateGroupMember(groupID, userID); err != nil {
			return err
		}
		if err := s.logActionWithTargetUser(tx, groupID, userID, "", "", groups.RoleMember, groups.ActionJoinGroup, "", userID); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось вступить в группу", "userID", userID, "groupID", groupID, "error", err)
		return nil, err
	}

	// Приватная группа - создаем заявку
	if target.IsPrivate {
		s.logger.Info("Создана заявка на вступление", "userID", userID, "groupID", groupID)
		return &GroupResult{
			Message: "Заявка на вступление отправлена, ожидайте подтверждения от администратора группы",
			Joined:  false,
		}, nil
	}

	s.logger.Info("Пользователь вступил в группу", "userID", userID, "groupID", groupID)
	return &GroupResult{
		Message: "Вы успешно присоединились к группе",
		Joined:  true,
	}, nil
}

func isPendingJoinRequestUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "idx_group_join_request_pending_unique"
	}

	errText := err.Error()
	return strings.Contains(errText, "idx_group_join_request_pending_unique") ||
		strings.Contains(errText, "UNIQUE constraint failed: group_join_requests.user_id, group_join_requests.group_id")
}

func isGroupMembershipUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "idx_group_user_membership"
	}

	errText := err.Error()
	return strings.Contains(errText, "idx_group_user_membership") ||
		strings.Contains(errText, "UNIQUE constraint failed: group_users.user_id, group_users.group_id")
}

// LeaveGroup выход из группы
func (s *groupService) LeaveGroup(userID uint, groupID uint) (bool, error) {
	err := s.runInTx(func(tx groupTx) error {
		role, err := newLeaveGroupStore(tx).LeaveGroup(userID, groupID)
		if err != nil {
			return err
		}
		if err := s.logActionWithTargetUser(tx, groupID, userID, "", "", role, groups.ActionLeaveGroup, "", userID); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось выйти из группы", "userID", userID, "groupID", groupID, "error", err)
		return false, err
	}

	s.logger.Info("Пользователь вышел из группы", "userID", userID, "groupID", groupID)
	return true, nil
}

// AcceptJoinInvite принимает приглашение в группу
func (s *groupService) AcceptJoinInvite(userID uint, inviteID uint) (*GroupResult, error) {
	var invite joinInviteResponse

	err := s.runInTx(func(tx groupTx) error {
		store := newJoinInviteResponseStore(tx)

		foundInvite, err := store.FindJoinInvite(inviteID)
		if err != nil {
			return err
		}
		invite = foundInvite

		if invite.UserID != userID {
			return ErrInviteNotOwned
		}

		if invite.Status == groups.JoinStatusAccepted {
			isMember, err := store.IsGroupMember(invite.GroupID, userID)
			if err != nil {
				return err
			}
			if isMember {
				return nil
			}
			return ErrInviteAlreadyHandled
		}

		if invite.Status != groups.JoinStatusPending {
			return ErrInviteAlreadyHandled
		}

		isBlacklisted, err := store.IsUserBlacklisted(invite.GroupID, userID)
		if err != nil {
			return err
		}
		if isBlacklisted {
			return ErrUserInBlacklist
		}

		if err := store.CreateGroupMemberIfMissing(invite.GroupID, userID); err != nil {
			return err
		}

		if err := store.UpdateInviteStatus(invite.ID, groups.JoinStatusAccepted); err != nil {
			return err
		}

		if err := s.logActionWithTargetUser(tx, invite.GroupID, userID, "", "", groups.RoleMember, groups.ActionAcceptInvite, "", userID); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось принять приглашение", "userID", userID, "inviteID", inviteID, "error", err)
		return nil, err
	}

	s.logger.Info("Приглашение в группу принято", "userID", userID, "inviteID", inviteID)
	return &GroupResult{
		Message: "Вы успешно приняли приглашение и присоединились к группе",
		Joined:  true,
	}, nil
}

// RejectJoinInvite отклоняет приглашение в группу
func (s *groupService) RejectJoinInvite(userID uint, inviteID uint) (bool, error) {
	err := s.runInTx(func(tx groupTx) error {
		store := newJoinInviteResponseStore(tx)

		invite, err := store.FindJoinInvite(inviteID)
		if err != nil {
			return err
		}

		if invite.UserID != userID {
			return ErrInviteNotOwned
		}

		if invite.Status != groups.JoinStatusPending {
			return ErrInviteAlreadyHandled
		}

		if err := store.UpdateInviteStatus(invite.ID, groups.JoinStatusRejected); err != nil {
			return err
		}

		if err := s.logActionWithTargetUser(tx, invite.GroupID, userID, "", "", groups.RoleMember, groups.ActionRejectInvite, "", userID); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось отклонить приглашение", "userID", userID, "inviteID", inviteID, "error", err)
		return false, err
	}

	s.logger.Info("Приглашение в группу отклонено", "userID", userID, "inviteID", inviteID)
	return true, nil
}
