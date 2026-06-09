package group

import (
	"errors"
	"fmt"
	"friendship/models/groups"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
			return store.CreatePendingJoinRequest(groupID, userID)
		}

		return store.CreateGroupMember(groupID, userID)
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
	var groupUser groups.GroupUsers

	err := s.runInTx(func(tx groupTx) error {
		err := tx.Where("user_id = ? AND group_id = ?", userID, groupID).
			First(&groupUser).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotInGroup
			}
			return fmt.Errorf("ошибка поиска участника: %w", err)
		}

		// Проверяем, не админ ли это (админ не может выйти)
		var role groups.Role_in_group
		if err := tx.First(&role, groupUser.RoleInGroupID).Error; err == nil {
			if groups.HasCapability(role.Name, groups.CapabilityAdmin) {
				return fmt.Errorf("администратор не может покинуть группу. Передайте права другому участнику или удалите группу")
			}
		}

		if err := tx.Delete(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка удаления участника: %w", err)
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
	var invite groups.GroupJoinInvite

	err := s.runInTx(func(tx groupTx) error {
		if err := tx.First(&invite, inviteID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInviteNotFound
			}
			return fmt.Errorf("ошибка поиска приглашения: %w", err)
		}

		if invite.UserID != userID {
			return ErrInviteNotOwned
		}

		if invite.Status == "accepted" {
			isMember, err := groupMembershipExists(tx, userID, invite.GroupID)
			if err != nil {
				return err
			}
			if isMember {
				return nil
			}
			return ErrInviteAlreadyHandled
		}

		if invite.Status != "pending" {
			return ErrInviteAlreadyHandled
		}

		// Проверяем черный список
		var blacklistCount int64
		if err := tx.Model(&groups.GroupBlacklist{}).
			Where("group_id = ? AND user_id = ?", invite.GroupID, userID).
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

		isMember, err := groupMembershipExists(tx, userID, invite.GroupID)
		if err != nil {
			return err
		}
		if isMember {
			if err := tx.Model(&invite).Update("status", "accepted").Error; err != nil {
				return fmt.Errorf("ошибка обновления статуса приглашения: %w", err)
			}
			return nil
		}

		// Добавляем в группу
		groupUser := groups.GroupUsers{
			UserID:        userID,
			GroupID:       invite.GroupID,
			RoleInGroupID: memberRoleID,
		}
		createMembership := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&groupUser)
		if createMembership.Error != nil {
			return fmt.Errorf("ошибка добавления пользователя в группу: %w", createMembership.Error)
		}
		if createMembership.RowsAffected == 0 {
			isMember, err = groupMembershipExists(tx, userID, invite.GroupID)
			if err != nil {
				return err
			}
			if !isMember {
				return fmt.Errorf("членство пользователя в группе не подтверждено после принятия приглашения")
			}
		}

		// Обновляем статус приглашения
		if err := tx.Model(&invite).Update("status", "accepted").Error; err != nil {
			return fmt.Errorf("ошибка обновления статуса приглашения: %w", err)
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

func groupMembershipExists(tx groupTx, userID uint, groupID uint) (bool, error) {
	var existingCount int64
	if err := tx.Model(&groups.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&existingCount).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства: %w", err)
	}
	return existingCount > 0, nil
}

// RejectJoinInvite отклоняет приглашение в группу
func (s *groupService) RejectJoinInvite(userID uint, inviteID uint) (bool, error) {
	var invite groups.GroupJoinInvite

	err := s.runInTx(func(tx groupTx) error {
		if err := tx.First(&invite, inviteID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInviteNotFound
			}
			return fmt.Errorf("ошибка поиска приглашения: %w", err)
		}

		if invite.UserID != userID {
			return ErrInviteNotOwned
		}

		if invite.Status != "pending" {
			return ErrInviteAlreadyHandled
		}

		if err := tx.Model(&invite).Update("status", "rejected").Error; err != nil {
			return fmt.Errorf("ошибка обновления статуса приглашения: %w", err)
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
