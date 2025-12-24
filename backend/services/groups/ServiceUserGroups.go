package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

// JoinGroup вступление в группу
func (s *groupService) JoinGroup(userID uint, groupID uint) (*GroupResult, error) {
	var user models.User
	var group groups.Group

	err := s.post.Transaction(func(tx repository.PostgresRepository) error {
		if _, err := user.FindUserByID(userID, tx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.First(&group, groupID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("ошибка поиска группы: %w", err)
		}

		// Проверяем черный список
		var blacklistCount int64
		if err := tx.Model(&groups.GroupBlacklist{}).
			Where("group_id = ? AND user_id = ?", groupID, userID).
			Count(&blacklistCount).Error; err != nil {
			return fmt.Errorf("ошибка проверки черного списка: %w", err)
		}
		if blacklistCount > 0 {
			return ErrUserInBlacklist
		}

		// Проверяем, состоит ли уже в группе
		var existingCount int64
		if err := tx.Model(&groups.GroupUsers{}).
			Where("user_id = ? AND group_id = ?", userID, groupID).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("ошибка проверки членства: %w", err)
		}
		if existingCount > 0 {
			return ErrAlreadyInGroup
		}

		// Проверяем существующие заявки
		var existingRequest groups.GroupJoinRequest
		err := tx.Where("user_id = ? AND group_id = ? AND status = ?", userID, groupID, "pending").
			First(&existingRequest).Error
		if err == nil {
			return ErrRequestAlreadyExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ошибка проверки заявок: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to join group", "userID", userID, "groupID", groupID, "error", err)
		return nil, err
	}

	// Приватная группа - создаем заявку
	if group.IsPrivate {
		request := groups.GroupJoinRequest{
			UserID:  userID,
			GroupID: groupID,
			Status:  "pending",
		}
		if err := s.post.Create(&request).Error; err != nil {
			s.logger.Error("Failed to create join request", "error", err)
			return nil, fmt.Errorf("ошибка создания заявки: %w", err)
		}

		s.logger.Info("Join request created", "userID", userID, "groupID", groupID)
		return &GroupResult{
			Message: "Заявка на вступление отправлена, ожидайте подтверждения от администратора группы",
			Joined:  false,
		}, nil
	}

	// Открытая группа - сразу добавляем
	memberRoleID := new(groups.Role_in_group).GetIdRole("Участник", s.post)
	if memberRoleID == 0 {
		return nil, fmt.Errorf("роль member не найдена")
	}

	member := groups.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: memberRoleID,
	}
	if err := s.post.Create(&member).Error; err != nil {
		s.logger.Error("Failed to add user to group", "error", err)
		return nil, fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
	}

	s.logger.Info("User joined group", "userID", userID, "groupID", groupID)
	return &GroupResult{
		Message: "Вы успешно присоединились к группе",
		Joined:  true,
	}, nil
}

// LeaveGroup выход из группы
func (s *groupService) LeaveGroup(userID uint, groupID uint) (bool, error) {
	var groupUser groups.GroupUsers

	err := s.post.Transaction(func(tx repository.PostgresRepository) error {
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
			if role.Name == "Админ" {
				return fmt.Errorf("администратор не может покинуть группу. Передайте права другому участнику или удалите группу")
			}
		}

		if err := tx.Delete(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка удаления участника: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to leave group", "userID", userID, "groupID", groupID, "error", err)
		return false, err
	}

	s.logger.Info("User left group", "userID", userID, "groupID", groupID)
	return true, nil
}

// AcceptJoinInvite принимает приглашение в группу
func (s *groupService) AcceptJoinInvite(userID uint, inviteID uint) (*GroupResult, error) {
	var invite groups.GroupJoinInvite

	err := s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&invite, inviteID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("приглашение не найдено")
			}
			return fmt.Errorf("ошибка поиска приглашения: %w", err)
		}

		if invite.UserID != userID {
			return fmt.Errorf("это не ваше приглашение")
		}

		if invite.Status != "pending" {
			return fmt.Errorf("приглашение уже обработано")
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

		memberRoleID := new(groups.Role_in_group).GetIdRole("Участник", s.post)
		if memberRoleID == 0 {
			return fmt.Errorf("роль member не найдена")
		}

		// Добавляем в группу
		groupUser := groups.GroupUsers{
			UserID:        userID,
			GroupID:       invite.GroupID,
			RoleInGroupID: memberRoleID,
		}
		if err := tx.Create(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
		}

		// Обновляем статус приглашения
		if err := tx.Model(&invite).Update("status", "accepted").Error; err != nil {
			return fmt.Errorf("ошибка обновления статуса приглашения: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to accept invite", "userID", userID, "inviteID", inviteID, "error", err)
		return nil, err
	}

	s.logger.Info("Accepted join invite", "userID", userID, "inviteID", inviteID)
	return &GroupResult{
		Message: "Вы успешно приняли приглашение и присоединились к группе",
		Joined:  true,
	}, nil
}

// RejectJoinInvite отклоняет приглашение в группу
func (s *groupService) RejectJoinInvite(userID uint, inviteID uint) (bool, error) {
	var invite groups.GroupJoinInvite

	err := s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&invite, inviteID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("приглашение не найдено")
			}
			return fmt.Errorf("ошибка поиска приглашения: %w", err)
		}

		if invite.UserID != userID {
			return fmt.Errorf("это не ваше приглашение")
		}

		if invite.Status != "pending" {
			return fmt.Errorf("приглашение уже обработано")
		}

		if err := tx.Model(&invite).Update("status", "rejected").Error; err != nil {
			return fmt.Errorf("ошибка обновления статуса приглашения: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to reject invite", "userID", userID, "inviteID", inviteID, "error", err)
		return false, err
	}

	s.logger.Info("Rejected join invite", "userID", userID, "inviteID", inviteID)
	return true, nil
}
