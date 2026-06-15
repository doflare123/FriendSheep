package group

import (
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	"friendship/models/groups"
	"friendship/services"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound          = errors.New("пользователь не найден")
	ErrCategoriesNotFound    = errors.New("категории не найдены")
	ErrInvalidInput          = errors.New("невалидная структура данных")
	ErrGroupCreation         = errors.New("ошибка создания группы")
	ErrGroupNotFound         = errors.New("группа не найдена")
	ErrPermissionDenied      = errors.New("недостаточно прав")
	ErrAlreadyInGroup        = errors.New("пользователь уже в группе")
	ErrNotInGroup            = errors.New("пользователь не в группе")
	ErrRequestAlreadyExists  = errors.New("заявка уже существует")
	ErrJoinRequestNotFound   = errors.New("заявка не найдена")
	ErrInviteAlreadyExists   = errors.New("приглашение уже существует")
	ErrUserInBlacklist       = errors.New("пользователь в черном списке")
	ErrInviteNotFound        = errors.New("приглашение не найдено")
	ErrInviteNotOwned        = errors.New("приглашение не принадлежит пользователю")
	ErrInviteAlreadyHandled  = errors.New("приглашение уже обработано")
	ErrCannotRemoveSelf      = errors.New("нельзя удалить самого себя")
	ErrCannotChangeOwnRole   = errors.New("нельзя изменить собственную роль")
	ErrRoleAdminNotFound     = errors.New("роль администратора не найдена")
	ErrRoleModeratorNotFound = errors.New("роль модератора не найдена")
	ErrRoleMemberNotFound    = errors.New("роль участника не найдена")
	ErrJoinRequestHandled    = errors.New("заявка уже обработана")
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
	CreateGroup(id uint, inf CreateGroupInput) (*dto.GroupFullDto, error)
	UpdateGroup(actorID uint, inf GroupUpdateInput) (*dto.GroupFullDto, error)
	DeleteGroup(actorID uint, groupID uint) (bool, error)
	GetGroupDetails(userID uint, groupID uint) (*dto.GroupFullDto, error)

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
	post   groupStore
	tx     groupTransactionRunner
	access groupActorRoleFinder
	reads  groupManagementReadStore
}

func NewGroupService(logger logger.Logger, rep groupStore) GroupsService {
	return &groupService{
		logger: logger,
		post:   rep,
		tx:     newGroupTransactionRunner(rep),
		access: newGroupAccessStore(rep),
		reads:  newGroupManagementReadStore(rep),
	}
}

// GetGroupDetails получает полную информацию о группе
func (s *groupService) GetGroupDetails(userID uint, groupID uint) (*dto.GroupFullDto, error) {
	var group groups.Group

	err := s.post.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		First(&group, groupID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("ошибка получения группы: %w", err)
	}

	if group.IsPrivate {
		var memberCount int64
		s.post.Model(&groups.GroupUsers{}).
			Where("group_id = ? AND user_id = ?", groupID, userID).
			Count(&memberCount)

		if memberCount == 0 {
			return nil, ErrPermissionDenied
		}
	}

	var totalMembers int64
	s.post.Model(&groups.GroupUsers{}).
		Where("group_id = ?", groupID).
		Count(&totalMembers)

	var groupUsers []groups.GroupUsers
	err = s.post.
		Preload("User").
		Preload("RoleInGroup").
		Where("group_id = ?", groupID).
		Limit(10).
		Find(&groupUsers).Error

	if err != nil {
		s.logger.Error("Не удалось получить участников группы", "groupID", groupID, "error", err)
		return nil, fmt.Errorf("ошибка получения участников: %w", err)
	}

	members := make([]dto.GroupMemberDto, 0, len(groupUsers))
	for _, gu := range groupUsers {
		members = append(members, dto.GroupMemberDto{
			ID:       gu.User.ID,
			Name:     gu.User.Name,
			Username: gu.User.Us,
			Image:    gu.User.Image,
			Role:     gu.RoleInGroup.Name,
		})
	}

	var activeEvents []events.Event
	err = s.post.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Where("group_id = ? AND status_id IN (?)", groupID, []uint{1, 2}).
		Where("start_time > ?", time.Now()).
		Order("start_time ASC").
		Find(&activeEvents).Error

	if err != nil {
		s.logger.Error("Не удалось получить события группы", "groupID", groupID, "error", err)
		activeEvents = []events.Event{}
	}

	groupEvents := convertorsdto.ConvertManyToShortDto(activeEvents)

	var userGroupMembership groups.GroupUsers
	isSubscribed := false
	userRole := ""

	err = s.post.
		Preload("RoleInGroup").
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&userGroupMembership).Error

	if err == nil {
		isSubscribed = true
		userRole = userGroupMembership.RoleInGroup.Name
	}

	// Формируем категории
	categories := make([]string, 0, len(group.Categories))
	for _, cat := range group.Categories {
		categories = append(categories, cat.Name)
	}

	// Формируем контакты
	contacts := make([]dto.ContactDto, 0, len(group.Contacts))
	for _, contact := range group.Contacts {
		contacts = append(contacts, dto.ContactDto{
			Name: contact.Name,
			Link: contact.Link,
		})
	}

	// Формируем итоговую DTO
	groupDto := &dto.GroupFullDto{
		ID:               group.ID,
		Name:             group.Name,
		Description:      group.Description,
		SmallDescription: group.SmallDescription,
		Image:            group.Image,
		IsPrivate:        group.IsPrivate,
		Enterprise:       group.Enterprise,
		City:             group.City,
		Categories:       categories,
		Contacts:         contacts,
		MemberCount:      int(totalMembers),

		Creator: dto.GroupCreatorDto{
			ID:       group.Creater.ID,
			Name:     group.Creater.Name,
			Username: group.Creater.Us,
			Image:    group.Creater.Image,
			Verified: group.Creater.VerifiedUser,
		},

		Members:      members,
		ActiveEvents: groupEvents,

		IsSubscribed: isSubscribed,
		UserRole:     userRole,

		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}

	return groupDto, nil
}

// CreateGroup создает новую группу
func (s *groupService) CreateGroup(id uint, input CreateGroupInput) (*dto.GroupFullDto, error) {
	if err := services.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %w", err)
	}

	var contacts map[string]string
	if input.Contacts != "" {
		contacts = parseContacts(input.Contacts)
	}

	var creator models.User
	if err := s.post.First(&creator, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("Создатель группы не найден", "Id", id)
			return nil, fmt.Errorf("%w: %d", ErrUserNotFound, id)
		}
		s.logger.Error("Ошибка БД при поиске создателя", "Id", id, "error", err)
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	var categories []models.Category
	if len(input.Categories) > 0 {
		if err := s.post.Where("id IN ?", input.Categories).Find(&categories).Error; err != nil {
			s.logger.Error("Ошибка загрузки категорий", "categoryIDs", input.Categories, "error", err)
			return nil, fmt.Errorf("%w: %v", ErrCategoriesNotFound, err)
		}

		if len(categories) != len(input.Categories) {
			s.logger.Warn("Найдены не все запрошенные категории", "requested", len(input.Categories), "found", len(categories))
			return nil, ErrCategoriesNotFound
		}
	}

	var newGroup *groups.Group
	err := s.runInTx(func(tx groupTx) error {
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

		roleID, err := findGroupRoleID(tx, groups.RoleAdmin)
		if err != nil {
			return ErrRoleAdminNotFound
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
		if err := s.logAction(tx, newGroup.ID, id, creator.Name, creator.Us, groups.RoleAdmin, "create_group", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Транзакция создания группы завершилась с ошибкой", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrGroupCreation, err)
	}

	if err := s.post.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		First(newGroup, newGroup.ID).Error; err != nil {
		s.logger.Warn("Не удалось перезагрузить группу с ассоциациями", "groupID", newGroup.ID, "error", err)
	}

	s.logger.Info("Группа успешно создана", "groupID", newGroup.ID, "name", newGroup.Name, "creatorID", creator.ID)

	groupDto, err := s.GetGroupDetails(id, newGroup.ID)
	if err != nil {
		s.logger.Error("Не удалось сформировать полный DTO группы после создания", "groupID", newGroup.ID, "error", err)
		return nil, fmt.Errorf("ошибка формирования данных группы: %w", err)
	}

	return groupDto, nil
}

// UpdateGroup обновляет группу
func (s *groupService) UpdateGroup(actorID uint, input GroupUpdateInput) (*dto.GroupFullDto, error) {
	// Проверяем права доступа
	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, groups.CapabilityModerate)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	var group groups.Group
	var actor models.User

	err = s.runInTx(func(tx groupTx) error {
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
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось обновить группу", "groupID", input.GroupID, "error", err)
		return nil, err
	}

	if err := s.post.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		First(&group, group.ID).Error; err != nil {
		s.logger.Warn("Не удалось перезагрузить группу с ассоциациями", "groupID", group.ID, "error", err)
	}

	s.logger.Info("Группа успешно обновлена", "groupID", group.ID, "actorID", actorID)

	groupDto, err := s.GetGroupDetails(actorID, group.ID)
	if err != nil {
		s.logger.Error("Не удалось сформировать полный DTO группы после обновления", "groupID", group.ID, "error", err)
		return nil, fmt.Errorf("ошибка формирования данных группы: %w", err)
	}

	return groupDto, nil
}

// DeleteGroup удаляет группу (только админ)
func (s *groupService) DeleteGroup(actorID uint, groupID uint) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, groups.CapabilityAdmin)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var actor models.User
	var group groups.Group

	err = s.runInTx(func(tx groupTx) error {
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
		s.logger.Error("Не удалось удалить группу", "groupID", groupID, "error", err)
		return false, err
	}

	s.logger.Info("Группа успешно удалена", "groupID", groupID, "actorID", actorID, "role", role)
	return true, nil
}

// AddPermissions добавляет права оператора (только админ)
func (s *groupService) AddPermissions(actorID uint, input PermissionInput) (bool, error) {
	if actorID == input.UserID {
		return false, ErrCannotChangeOwnRole
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, groups.CapabilityAdmin)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User
	var groupUser groups.GroupUsers

	err = s.runInTx(func(tx groupTx) error {
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

		operatorRoleID, err := findGroupRoleID(tx, groups.RoleModerator)
		if err != nil {
			return ErrRoleModeratorNotFound
		}

		if err := tx.Model(&groupUser).Update("role_in_group_id", operatorRoleID).Error; err != nil {
			return fmt.Errorf("ошибка обновления роли: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Назначил пользователя '%s' (@%s) оператором группы", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, input.GroupID, actorID, actor.Name, actor.Us, role, "add_operator", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось добавить права", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	s.logger.Info("Права модератора выданы", "actorID", actorID, "targetUserID", input.UserID, "groupID", input.GroupID)
	return true, nil
}

// RemovePermissions убирает права оператора (только админ)
func (s *groupService) RemovePermissions(actorID uint, input PermissionInput) (bool, error) {
	if actorID == input.UserID {
		return false, ErrCannotChangeOwnRole
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, groups.CapabilityAdmin)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User
	var groupUser groups.GroupUsers

	err = s.runInTx(func(tx groupTx) error {
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

		memberRoleID, err := findGroupRoleID(tx, groups.RoleMember)
		if err != nil {
			return ErrRoleMemberNotFound
		}

		if err := tx.Model(&groupUser).Update("role_in_group_id", memberRoleID).Error; err != nil {
			return fmt.Errorf("ошибка обновления роли: %w", err)
		}

		// Логируем действие
		action := fmt.Sprintf("Снял с пользователя '%s' (@%s) права оператора", targetUser.Name, targetUser.Us)
		if err := s.logAction(tx, input.GroupID, actorID, actor.Name, actor.Us, role, "remove_operator", action); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось снять права", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	s.logger.Info("Права модератора сняты", "actorID", actorID, "targetUserID", input.UserID, "groupID", input.GroupID)
	return true, nil
}

// DeleteUserFromGroup удаляет пользователя из группы и добавляет в черный список
func (s *groupService) DeleteUserFromGroup(actorID uint, groupID uint, targetUserID uint) (bool, error) {
	if actorID == targetUserID {
		return false, ErrCannotRemoveSelf
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User
	var groupUser groups.GroupUsers

	err = s.runInTx(func(tx groupTx) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.First(&targetUser, targetUserID).Error; err != nil {
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
			if groups.HasCapability(targetRole.Name, groups.CapabilityAdmin) {
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
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось удалить пользователя из группы", "actorID", actorID, "targetUserID", targetUserID, "error", err)
		return false, err
	}

	s.logger.Info("Пользователь удален из группы и добавлен в черный список", "actorID", actorID, "targetUserID", targetUserID, "groupID", groupID)
	return true, nil
}

// RemoveFromBlacklist убирает пользователя из черного списка
func (s *groupService) RemoveFromBlacklist(actorID uint, groupID uint, targetUserID uint) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var targetUser models.User
	var actor models.User

	err = s.runInTx(func(tx groupTx) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.First(&targetUser, targetUserID).Error; err != nil {
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
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось убрать пользователя из черного списка", "actorID", actorID, "targetUserID", targetUserID, "error", err)
		return false, err
	}

	s.logger.Info("Пользователь убран из черного списка", "actorID", actorID, "targetUserID", targetUserID, "groupID", groupID)
	return true, nil
}

// WatchRecentActions получает историю действий в группе
func (s *groupService) WatchRecentActions(userID uint, groupID uint, limit int) ([]GroupAction, error) {
	// Проверяем, что пользователь в группе
	hasAccess, _, err := s.checkGroupAccess(userID, groupID, groups.CapabilityMember)
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

	actions, err := s.reads.ListGroupActions(groupID, limit)
	if err != nil {
		s.logger.Error("Не удалось получить историю действий группы", "groupID", groupID, "error", err)
		return nil, err
	}

	return actions, nil
}

// Вспомогательные функции

// checkGroupAccess проверяет, имеет ли пользователь доступ к группе с нужной ролью
func (s *groupService) checkGroupAccess(userID uint, groupID uint, required groups.Capability) (bool, string, error) {
	return s.access.FindActorRole(userID, groupID, required)
}

// GetGroupBlacklist получает черный список группы
func (s *groupService) GetGroupBlacklist(actorID uint, groupID uint, limit int) ([]BlacklistUser, error) {
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

	blacklist, err := s.reads.ListGroupBlacklist(groupID, limit)
	if err != nil {
		s.logger.Error("Не удалось получить черный список", "groupID", groupID, "error", err)
		return nil, err
	}

	return blacklist, nil
}

// updateContactsInTx обновляет контакты в транзакции
func updateContactsInTx(tx groupTx, groupID *uint, newContacts map[string]string) error {
	var existingContacts []groups.GroupContact
	if err := tx.Where("group_id = ?", groupID).Find(&existingContacts).Error; err != nil {
		return fmt.Errorf("ошибка получения существующих контактов: %w", err)
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
					return fmt.Errorf("ошибка обновления контакта '%s': %w", name, err)
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
				return fmt.Errorf("ошибка добавления контакта '%s': %w", name, err)
			}
		}
	}

	for _, contactToDelete := range existingMap {
		if err := tx.Delete(&contactToDelete).Error; err != nil {
			contactName := "unknown"
			if contactToDelete.Name != "" {
				contactName = contactToDelete.Name
			}
			return fmt.Errorf("ошибка удаления старого контакта '%s': %w", contactName, err)
		}
	}

	return nil
}

// logAction записывает действие в лог
func (s *groupService) logAction(tx groupTx, groupID uint, userID uint, username, us, role, actionType, description string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(us) == "" {
		var actor models.User
		if err := tx.Select("id", "name", "us").First(&actor, userID).Error; err == nil {
			if strings.TrimSpace(username) == "" {
				username = actor.Name
			}
			if strings.TrimSpace(us) == "" {
				us = actor.Us
			}
		}
	}

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
