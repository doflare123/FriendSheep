package group

import (
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	"friendship/models/groups"
	"friendship/repository"
	"friendship/services"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound         = errors.New("пользователь не найден")
	ErrCategoriesNotFound   = errors.New("категории не найдены")
	ErrInvalidInput         = errors.New("невалидная структура данных")
	ErrGroupCreation        = errors.New("ошибка создания группы")
	ErrGroupNotFound        = errors.New("группа не найдена")
	ErrPermissionDenied     = errors.New("недостаточно прав")
	ErrAlreadyInGroup       = errors.New("пользователь уже в группе")
	ErrNotInGroup           = errors.New("пользователь не в группе")
	ErrRequestAlreadyExists = errors.New("заявка уже существует")
	ErrInviteAlreadyExists  = errors.New("приглашение уже существует")
	ErrUserInBlacklist      = errors.New("пользователь в черном списке")
	ErrCannotRemoveSelf     = errors.New("нельзя удалить самого себя")
	ErrCannotChangeOwnRole  = errors.New("нельзя изменить собственную роль")
)

type CreateGroupInput struct {
	Name             string  `json:"name" form:"name" binding:"required,min=5,max=40" example:"Любители настольных игр"`
	Description      string  `json:"description" form:"description" binding:"required,min=5,max=300" example:"Группа для любителей игр"`
	SmallDescription string  `json:"smallDescription" form:"smallDescription" binding:"required,min=5,max=50" example:"Играем вместе!"`
	Image            string  `json:"image" form:"image" binding:"required,url" example:"https://cdn.example.com/images/board-games.jpg"`
	IsPrivate        *bool   `json:"isPrivate" form:"isPrivate" binding:"required" example:"false"`
	City             string  `json:"city,omitempty" form:"city" example:"Москва"`
	Categories       []*uint `json:"categories" form:"categories" binding:"required,min=1" example:"[1,3,5]"`
	Contacts         string  `json:"contacts,omitempty" form:"contacts" example:"vk:https://vk.com/mygroup, tg:https://t.me/mygroup"`
}

type GroupUpdateInput struct {
	GroupID          uint
	Name             *string
	Description      *string
	SmallDescription *string
	Image            *string
	IsPrivate        *bool
	City             *string
	Categories       []*uint
	Contacts         *string
}

type PermissionInput struct {
	GroupID uint `json:"groupId" binding:"required"`
	UserID  uint `json:"userId" binding:"required"`
}

type JoinInviteInput struct {
	GroupID uint `json:"groupId" binding:"required"`
	UserID  uint `json:"userId" binding:"required"`
}

type GroupResult struct {
	Message string `json:"message"`
	Joined  bool   `json:"joined"`
}

type GroupAction struct {
	ID          uint      `json:"id"`
	GroupID     uint      `json:"groupId"`
	UserID      uint      `json:"userId"`
	Username    string    `json:"username"`
	Us          string    `json:"us"`
	Role        string    `json:"role"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type BlacklistUser struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"userId"`
	Name         string    `json:"name"`
	Us           string    `json:"us"`
	Image        string    `json:"image"`
	BannedBy     uint      `json:"bannedBy"`
	BannedByName string    `json:"bannedByName"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

type JoinRequestInfo struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Name      string    `json:"name"`
	Us        string    `json:"us"`
	Image     string    `json:"image"`
	GroupID   uint      `json:"groupId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type GroupsService interface {
	// Базовые операции с группами
	CreateGroup(id uint, inf CreateGroupInput) (*dto.GroupDto, error)
	UpdateGroup(actorID uint, inf GroupUpdateInput) (*dto.GroupDto, error)
	DeleteGroup(actorID uint, groupID uint) (bool, error)

	// Управление заявками
	ApproveAllJoinRequests(actorID uint, groupID uint) (int, error)
	RejectAllJoinRequests(actorID uint, groupID uint) (int, error)
	ApproveJoinRequest(actorID uint, requestID uint) (bool, error)
	RejectJoinRequest(actorID uint, requestID uint) (bool, error)
	GetJoinRequests(actorID uint, groupID uint, status string, limit int) ([]JoinRequestInfo, error)

	// Вступление/выход
	JoinGroup(userID uint, groupID uint) (*GroupResult, error)
	LeaveGroup(userID uint, groupID uint) (bool, error)

	// Управление правами
	AddPermissions(actorID uint, input PermissionInput) (bool, error)
	RemovePermissions(actorID uint, input PermissionInput) (bool, error)

	// Управление участниками
	DeleteUserFromGroup(actorID uint, groupID uint, targetUserID uint) (bool, error)
	RemoveFromBlacklist(actorID uint, groupID uint, targetUserID uint) (bool, error)
	GetGroupBlacklist(actorID uint, groupID uint, limit int) ([]BlacklistUser, error)

	// Приглашения
	CreateJoinInvite(actorID uint, input JoinInviteInput) (bool, error)
	AcceptJoinInvite(userID uint, inviteID uint) (*GroupResult, error)
	RejectJoinInvite(userID uint, inviteID uint) (bool, error)

	// История действий
	WatchRecentActions(userID uint, groupID uint, limit int) ([]GroupAction, error)
}

type groupService struct {
	logger logger.Logger
	post   repository.PostgresRepository
}

func NewGroupService(logger logger.Logger, rep repository.PostgresRepository) GroupsService {
	return &groupService{
		logger: logger,
		post:   rep,
	}
}

// CreateGroup создает новую группу
func (s *groupService) CreateGroup(id uint, input CreateGroupInput) (*dto.GroupDto, error) {
	if err := services.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	var contacts map[string]string
	if input.Contacts != "" {
		contacts = parseContacts(input.Contacts)
	}

	var creator models.User
	if _, err := creator.FindUserByID(id, s.post); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("Creator not found", "Id", id)
			return nil, fmt.Errorf("%w: %d", ErrUserNotFound, id)
		}
		s.logger.Error("Database error while finding creator", "Id", id, "error", err)
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	var categories []models.Category
	if len(input.Categories) > 0 {
		if err := s.post.Where("id IN ?", input.Categories).Find(&categories).Error; err != nil {
			s.logger.Error("Error loading categories", "categoryIDs", input.Categories, "error", err)
			return nil, fmt.Errorf("%w: %v", ErrCategoriesNotFound, err)
		}

		if len(categories) != len(input.Categories) {
			s.logger.Warn("Not all categories found", "requested", len(input.Categories), "found", len(categories))
		}
	}

	var newGroup *groups.Group
	err := s.post.Transaction(func(tx repository.PostgresRepository) error {
		newGroup = &groups.Group{
			Name:             input.Name,
			Description:      input.Description,
			SmallDescription: input.SmallDescription,
			Image:            input.Image,
			CreaterID:        id,
			IsPrivate:        *input.IsPrivate,
			City:             input.City,
			Categories:       categories,
		}

		if err := tx.Create(newGroup).Error; err != nil {
			return fmt.Errorf("ошибка создания группы: %w", err)
		}

		roleID := new(groups.Role_in_group).GetIdRole("Админ", s.post)
		if roleID == 0 {
			return fmt.Errorf("роль Админ не найдена")
		}

		groupUser := groups.GroupUsers{
			UserID:        id,
			GroupID:       newGroup.ID,
			RoleInGroupID: roleID,
		}

		if err := tx.Create(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
		}

		if len(contacts) > 0 {
			groupContacts := make([]groups.GroupContact, 0, len(contacts))
			for name, link := range contacts {
				if name != "" && link != "" {
					groupContacts = append(groupContacts, groups.GroupContact{
						GroupID: newGroup.ID,
						Name:    name,
						Link:    link,
					})
				}
			}

			if len(groupContacts) > 0 {
				if err := tx.Create(&groupContacts).Error; err != nil {
					return fmt.Errorf("ошибка сохранения контактов группы: %w", err)
				}
			}
		}

		// Логируем действие
		action := fmt.Sprintf("Создал группу '%s'", newGroup.Name)
		if err := s.logAction(tx, newGroup.ID, id, creator.Name, creator.Us, "Админ", "create_group", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed during group creation", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrGroupCreation, err)
	}

	if err := s.post.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		First(newGroup, newGroup.ID).Error; err != nil {
		s.logger.Warn("Failed to reload group with associations", "groupID", newGroup.ID, "error", err)
	}

	s.logger.Info("Group created successfully", "groupID", newGroup.ID, "name", newGroup.Name, "creatorID", creator.ID)

	groupDto := convertGroupToDto(newGroup)
	return groupDto, nil
}

// UpdateGroup обновляет группу
func (s *groupService) UpdateGroup(actorID uint, input GroupUpdateInput) (*dto.GroupDto, error) {
	// Проверяем права доступа
	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, []string{"Админ", "Модератор"})
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	var group groups.Group
	var actor models.User

	err = s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&group, input.GroupID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("ошибка поиска группы: %w", err)
		}

		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		// Обновляем только переданные поля
		updates := make(map[string]interface{})
		if input.Name != nil {
			updates["name"] = *input.Name
		}
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if input.SmallDescription != nil {
			updates["small_description"] = *input.SmallDescription
		}
		if input.Image != nil {
			updates["image"] = *input.Image
		}
		if input.IsPrivate != nil {
			updates["is_private"] = *input.IsPrivate
		}
		if input.City != nil {
			updates["city"] = *input.City
		}

		if len(updates) > 0 {
			if err := tx.Model(&group).Updates(updates).Error; err != nil {
				return fmt.Errorf("не удалось сохранить изменения группы: %w", err)
			}
		}

		// Обновляем категории только если они переданы
		if input.Categories != nil {
			// Очищаем старые категории
			if err := tx.Model(&group).Association("Categories").Clear(); err != nil {
				return fmt.Errorf("не удалось очистить старые категории: %w", err)
			}

			// Добавляем новые категории (если массив не пустой)
			if len(input.Categories) > 0 {
				var newCategories []models.Category
				if err := tx.Where("id IN ?", input.Categories).Find(&newCategories).Error; err != nil {
					return fmt.Errorf("не удалось найти переданные категории: %w", err)
				}
				if err := tx.Model(&group).Association("Categories").Replace(&newCategories); err != nil {
					return fmt.Errorf("не удалось назначить новые категории: %w", err)
				}
			}
		}

		// Обновляем контакты только если они переданы
		if input.Contacts != nil {
			newContacts := parseContacts(*input.Contacts)
			if err := updateContactsInTx(tx, &group.ID, newContacts); err != nil {
				return fmt.Errorf("ошибка обновления контактов: %w", err)
			}
		}

		// Логируем действие
		action := fmt.Sprintf("Обновил информацию группы '%s'", group.Name)
		if err := s.logAction(tx, group.ID, actorID, actor.Name, actor.Us, role, "update_group", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to update group", "groupID", input.GroupID, "error", err)
		return nil, err
	}

	if err := s.post.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		First(&group, group.ID).Error; err != nil {
		s.logger.Warn("Failed to reload group with associations", "groupID", group.ID, "error", err)
	}

	s.logger.Info("Group updated successfully", "groupID", group.ID, "actorID", actorID)

	groupDto := convertGroupToDto(&group)
	return groupDto, nil
}

// DeleteGroup удаляет группу (только админ)
func (s *groupService) DeleteGroup(actorID uint, groupID uint) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, []string{"Админ"})
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var actor models.User
	var group groups.Group

	err = s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&group, groupID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("ошибка поиска группы: %w", err)
		}

		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		// Удаляем группу (каскадно удалятся связи)
		if err := tx.Delete(&group).Error; err != nil {
			return fmt.Errorf("ошибка удаления группы: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to delete group", "groupID", groupID, "error", err)
		return false, err
	}

	s.logger.Info("Group deleted successfully", "groupID", groupID, "actorID", actorID, "role", role)
	return true, nil
}

// AddPermissions добавляет права оператора (только админ)
func (s *groupService) AddPermissions(actorID uint, input PermissionInput) (bool, error) {
	if actorID == input.UserID {
		return false, ErrCannotChangeOwnRole
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, []string{"Админ"})
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User
	var groupUser groups.GroupUsers

	err = s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.First(&targetUser, input.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		err := tx.Where("user_id = ? AND group_id = ?", input.UserID, input.GroupID).
			First(&groupUser).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotInGroup
			}
			return fmt.Errorf("ошибка поиска участника: %w", err)
		}

		operatorRoleID := new(groups.Role_in_group).GetIdRole("Модератор", s.post)
		if operatorRoleID == 0 {
			return fmt.Errorf("роль Модератор не найдена")
		}

		if err := tx.Model(&groupUser).Update("role_in_group_id", operatorRoleID).Error; err != nil {
			return fmt.Errorf("ошибка обновления роли: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Назначил пользователя '%s' (@%s) оператором группы", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, input.GroupID, actorID, actor.Name, actor.Us, role, "add_operator", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to add permissions", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	s.logger.Info("Added operator permissions", "actorID", actorID, "targetUserID", input.UserID, "groupID", input.GroupID)
	return true, nil
}

// RemovePermissions убирает права оператора (только админ)
func (s *groupService) RemovePermissions(actorID uint, input PermissionInput) (bool, error) {
	if actorID == input.UserID {
		return false, ErrCannotChangeOwnRole
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, []string{"Админ"})
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User
	var groupUser groups.GroupUsers

	err = s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.First(&targetUser, input.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		err := tx.Where("user_id = ? AND group_id = ?", input.UserID, input.GroupID).
			First(&groupUser).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotInGroup
			}
			return fmt.Errorf("ошибка поиска участника: %w", err)
		}

		memberRoleID := new(groups.Role_in_group).GetIdRole("Участник", s.post)
		if memberRoleID == 0 {
			return fmt.Errorf("роль Участник не найдена")
		}

		if err := tx.Model(&groupUser).Update("role_in_group_id", memberRoleID).Error; err != nil {
			return fmt.Errorf("ошибка обновления роли: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Снял с пользователя '%s' (@%s) права оператора", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, input.GroupID, actorID, actor.Name, actor.Us, role, "remove_operator", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to remove permissions", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	s.logger.Info("Removed operator permissions", "actorID", actorID, "targetUserID", input.UserID, "groupID", input.GroupID)
	return true, nil
}

// DeleteUserFromGroup удаляет пользователя из группы и добавляет в черный список
func (s *groupService) DeleteUserFromGroup(actorID uint, groupID uint, targetUserID uint) (bool, error) {
	if actorID == targetUserID {
		return false, ErrCannotRemoveSelf
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, []string{"Админ", "Модератор"})
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User
	var groupUser groups.GroupUsers

	err = s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if _, err := targetUser.FindUserByID(targetUserID, tx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		err := tx.Where("user_id = ? AND group_id = ?", targetUserID, groupID).
			First(&groupUser).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotInGroup
			}
			return fmt.Errorf("ошибка поиска участника: %w", err)
		}

		// Проверяем, что целевой пользователь не админ
		var targetRole groups.Role_in_group
		if err := tx.First(&targetRole, groupUser.RoleInGroupID).Error; err == nil {
			if targetRole.Name == "Админ" {
				return fmt.Errorf("нельзя удалить администратора группы")
			}
		}

		// Удаляем из группы
		if err := tx.Delete(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка удаления участника: %w", err)
		}

		// Добавляем в черный список
		blacklist := groups.GroupBlacklist{
			GroupID:   groupID,
			UserID:    targetUserID,
			BannedBy:  actorID,
			Reason:    "Удален из группы",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&blacklist).Error; err != nil {
			return fmt.Errorf("ошибка добавления в черный список: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Удалил пользователя '%s' (@%s) из группы и добавил в черный список", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, groupID, actorID, actor.Name, actor.Us, role, "ban_user", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to delete user from group", "actorID", actorID, "targetUserID", targetUserID, "error", err)
		return false, err
	}

	s.logger.Info("Deleted user from group and added to blacklist", "actorID", actorID, "targetUserID", targetUserID, "groupID", groupID)
	return true, nil
}

// RemoveFromBlacklist убирает пользователя из черного списка
func (s *groupService) RemoveFromBlacklist(actorID uint, groupID uint, targetUserID uint) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, []string{"Админ", "Модератор"})
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User

	err = s.post.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if _, err := targetUser.FindUserByID(targetUserID, tx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		result := tx.Where("group_id = ? AND user_id = ?", groupID, targetUserID).
			Delete(&groups.GroupBlacklist{})

		if result.Error != nil {
			return fmt.Errorf("ошибка удаления из черного списка: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("пользователь не найден в черном списке")
		}

		// Логируем действие
		action := fmt.Sprintf("Убрал пользователя '%s' (@%s) из черного списка", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, groupID, actorID, actor.Name, actor.Us, role, "unban_user", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to remove from blacklist", "actorID", actorID, "targetUserID", targetUserID, "error", err)
		return false, err
	}

	s.logger.Info("Removed from blacklist", "actorID", actorID, "targetUserID", targetUserID, "groupID", groupID)
	return true, nil
}

// WatchRecentActions получает историю действий в группе
func (s *groupService) WatchRecentActions(userID uint, groupID uint, limit int) ([]GroupAction, error) {
	// Проверяем, что пользователь в группе
	hasAccess, _, err := s.checkGroupAccess(userID, groupID, []string{"Админ", "Модератор", "Участник"})
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

	var actions []GroupAction
	err = s.post.Model(&groups.GroupActionLog{}).
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).
		Find(&actions).Error

	if err != nil {
		s.logger.Error("Failed to fetch group actions", "groupID", groupID, "error", err)
		return nil, fmt.Errorf("ошибка получения истории действий: %w", err)
	}

	return actions, nil
}

// Вспомогательные функции

// checkGroupAccess проверяет, имеет ли пользователь доступ к группе с нужной ролью
func (s *groupService) checkGroupAccess(userID uint, groupID uint, allowedRoles []string) (bool, string, error) {
	var groupUser groups.GroupUsers
	if userID != 0 || groupID != 0 {
		s.logger.Info("Тут прикол?", "Пользователь", userID, "Группа", groupID)
	}
	err := s.post.
		Preload("RoleInGroup").
		Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&groupUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", ErrNotInGroup
		}
		return false, "", fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	var role groups.Role_in_group
	if err := s.post.First(&role, groupUser.RoleInGroupID).Error; err != nil {
		return false, "", fmt.Errorf("ошибка получения роли: %w", err)
	}

	for _, allowedRole := range allowedRoles {
		if strings.EqualFold(role.Name, allowedRole) {
			return true, role.Name, nil
		}
	}

	return false, role.Name, nil
}

// GetGroupBlacklist получает черный список группы
func (s *groupService) GetGroupBlacklist(actorID uint, groupID uint, limit int) ([]BlacklistUser, error) {
	// Проверяем права доступа (admin или operator)
	hasAccess, _, err := s.checkGroupAccess(actorID, groupID, []string{"Админ", "Модератор"})
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

	var blacklist []groups.GroupBlacklist
	err = s.post.
		Preload("User").
		Preload("Banner").
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).
		Find(&blacklist).Error

	if err != nil {
		s.logger.Error("Failed to fetch blacklist", "groupID", groupID, "error", err)
		return nil, fmt.Errorf("ошибка получения черного списка: %w", err)
	}

	result := make([]BlacklistUser, 0, len(blacklist))
	for _, bl := range blacklist {
		result = append(result, BlacklistUser{
			ID:           bl.ID,
			UserID:       bl.UserID,
			Name:         bl.User.Name,
			Us:           bl.User.Us,
			Image:        bl.User.Image,
			BannedBy:     bl.BannedBy,
			BannedByName: bl.Banner.Name,
			Reason:       bl.Reason,
			CreatedAt:    bl.CreatedAt,
		})
	}

	return result, nil
}

// updateContactsInTx обновляет контакты в транзакции
func updateContactsInTx(tx repository.PostgresRepository, groupID *uint, newContacts map[string]string) error {
	var existingContacts []groups.GroupContact
	if err := tx.Where("group_id = ?", groupID).Find(&existingContacts).Error; err != nil {
		return fmt.Errorf("ошибка получения существующих контактов: %v", err)
	}

	existingMap := make(map[string]groups.GroupContact)
	for _, c := range existingContacts {
		if c.Name != "" {
			existingMap[c.Name] = c
		}
	}

	for name, link := range newContacts {
		if existingContact, ok := existingMap[name]; ok {
			if existingContact.Link != "" && existingContact.Link != link {
				if err := tx.Model(&existingContact).Update("link", link).Error; err != nil {
					return fmt.Errorf("ошибка обновления контакта '%s': %v", name, err)
				}
			}
			delete(existingMap, name)
		} else {
			newContact := groups.GroupContact{
				GroupID: *groupID,
				Name:    name,
				Link:    link,
			}
			if err := tx.Create(&newContact).Error; err != nil {
				return fmt.Errorf("ошибка добавления контакта '%s': %v", name, err)
			}
		}
	}

	for _, contactToDelete := range existingMap {
		if err := tx.Delete(&contactToDelete).Error; err != nil {
			contactName := "unknown"
			if contactToDelete.Name != "" {
				contactName = contactToDelete.Name
			}
			return fmt.Errorf("ошибка удаления старого контакта '%s': %v", contactName, err)
		}
	}

	return nil
}

// logAction записывает действие в лог
func (s *groupService) logAction(tx repository.PostgresRepository, groupID uint, userID uint, username, us, role, actionType, description string) error {
	action := groups.GroupActionLog{
		GroupID:     groupID,
		UserID:      userID,
		Username:    username,
		Us:          us,
		Role:        role,
		Action:      actionType,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return tx.Create(&action).Error
}

// parseContacts парсит строку контактов
func parseContacts(contactsStr string) map[string]string {
	contacts := make(map[string]string)
	if contactsStr == "" {
		return contacts
	}

	pairs := strings.Split(contactsStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			link := strings.TrimSpace(parts[1])
			if name != "" && link != "" {
				contacts[name] = link
			}
		}
	}

	return contacts
}

// convertGroupToDto конвертирует группу в DTO
func convertGroupToDto(group *groups.Group) *dto.GroupDto {
	if group == nil {
		return nil
	}

	categoryNames := make([]string, 0, len(group.Categories))
	for _, cat := range group.Categories {
		categoryNames = append(categoryNames, cat.Name)
	}

	contacts := make([]dto.ContactDto, 0, len(group.Contacts))
	for _, contact := range group.Contacts {
		contacts = append(contacts, dto.ContactDto{
			Name: contact.Name,
			Link: contact.Link,
		})
	}

	return &dto.GroupDto{
		ID:               group.ID,
		Name:             group.Name,
		Description:      group.Description,
		SmallDescription: group.SmallDescription,
		Image:            group.Image,
		CreatorID:        group.CreaterID,
		CreatorName:      group.Creater.Name,
		CreatorUsername:  group.Creater.Us,
		IsPrivate:        group.IsPrivate,
		City:             group.City,
		Categories:       categoryNames,
		Contacts:         contacts,
		CreatedAt:        group.CreatedAt,
		UpdatedAt:        group.UpdatedAt,
	}
}
