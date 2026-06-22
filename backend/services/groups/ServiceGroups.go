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

type GroupUserInput struct {
	GroupID uint `json:"groupId" binding:"required"`
	UserID  uint `json:"userId" binding:"required"`
}

type GroupResult struct {
	Message string `json:"message"`
	Joined  bool   `json:"joined"`
}

type GroupAction struct {
	ID           uint      `json:"id"`
	GroupID      uint      `json:"groupId"`
	UserID       uint      `json:"userId"`
	Username     string    `json:"username"`
	Us           string    `json:"us"`
	Role         string    `json:"role"`
	ActionTypeID uint      `json:"actionTypeId"`
	Action       string    `json:"action"`
	ActionName   string    `json:"actionName"`
	Description  string    `json:"description"`
	TargetUserID *uint     `json:"targetUserId,omitempty"`
	TargetName   string    `json:"targetName,omitempty"`
	TargetUs     string    `json:"targetUs,omitempty"`
	EntityID     *uint     `json:"entityId,omitempty"`
	EntityName   string    `json:"entityName,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type GroupActionFilter struct {
	Limit        int
	Action       string
	ActionTypeID uint
	Order        string
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
	AddPermissions(actorID uint, input GroupUserInput) (bool, error)
	RemovePermissions(actorID uint, input GroupUserInput) (bool, error)

	// Управление участниками
	DeleteUserFromGroup(actorID uint, groupID uint, targetUserID uint) (bool, error)
	RemoveFromBlacklist(actorID uint, groupID uint, targetUserID uint) (bool, error)
	GetGroupBlacklist(actorID uint, groupID uint, limit int) ([]BlacklistUser, error)

	// Приглашения
	CreateJoinInvite(actorID uint, input GroupUserInput) (bool, error)
	AcceptJoinInvite(userID uint, inviteID uint) (*GroupResult, error)
	RejectJoinInvite(userID uint, inviteID uint) (bool, error)

	// История действий
	WatchRecentActions(userID uint, groupID uint, filter GroupActionFilter) ([]GroupAction, error)
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

	var result groupCreateResult
	err := s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		var createErr error
		result, createErr = store.CreateGroup(id, input, contacts)
		return createErr
	})

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.logger.Error("Создатель группы не найден", "Id", id)
			return nil, err
		}
		if errors.Is(err, ErrCategoriesNotFound) {
			s.logger.Error("Ошибка загрузки категорий", "categoryIDs", input.Categories, "error", err)
			return nil, err
		}
		s.logger.Error("Транзакция создания группы завершилась с ошибкой", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrGroupCreation, err)
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Группа успешно создана", "groupID", result.GroupID, "name", result.GroupName, "creatorID", result.CreatorID)

	groupDto, err := s.GetGroupDetails(id, result.GroupID)
	if err != nil {
		s.logger.Error("Не удалось сформировать полный DTO группы после создания", "groupID", result.GroupID, "error", err)
		return nil, fmt.Errorf("ошибка формирования данных группы: %w", err)
	}

	return groupDto, nil
}

// UpdateGroup обновляет группу
func (s *groupService) UpdateGroup(actorID uint, input GroupUpdateInput) (*dto.GroupFullDto, error) {
	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, groups.CapabilityModerate)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	var result groupAdminResult

	err = s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var updateErr error
		result, updateErr = store.UpdateGroup(input, actor, role)
		return updateErr
	})

	if err != nil {
		s.logger.Error("Не удалось обновить группу", "groupID", input.GroupID, "error", err)
		return nil, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Группа успешно обновлена", "groupID", input.GroupID, "actorID", actorID)

	groupDto, err := s.GetGroupDetails(actorID, input.GroupID)
	if err != nil {
		s.logger.Error("Не удалось сформировать полный DTO группы после обновления", "groupID", input.GroupID, "error", err)
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

	err = s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		if _, err := store.FindActor(actorID); err != nil {
			return err
		}

		return store.DeleteGroup(groupID, role)
	})

	if err != nil {
		s.logger.Error("Не удалось удалить группу", "groupID", groupID, "error", err)
		return false, err
	}

	s.logger.Info("Группа успешно удалена", "groupID", groupID, "actorID", actorID, "role", role)
	return true, nil
}

// AddPermissions добавляет права оператора (только админ)
func (s *groupService) AddPermissions(actorID uint, input GroupUserInput) (bool, error) {
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

	var result groupAdminResult

	err = s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var changeErr error
		result, changeErr = store.ChangeMemberRole(input, groups.RoleModerator, actor, role)
		return changeErr
	})

	if err != nil {
		s.logger.Error("Не удалось добавить права", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Права модератора выданы", "actorID", actorID, "targetUserID", input.UserID, "groupID", input.GroupID)
	return true, nil
}

// RemovePermissions убирает права оператора (только админ)
func (s *groupService) RemovePermissions(actorID uint, input GroupUserInput) (bool, error) {
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

	var result groupAdminResult

	err = s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var changeErr error
		result, changeErr = store.ChangeMemberRole(input, groups.RoleMember, actor, role)
		return changeErr
	})

	if err != nil {
		s.logger.Error("Не удалось снять права", "actorID", actorID, "targetUserID", input.UserID, "error", err)
		return false, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
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

	var result groupAdminResult

	err = s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var banErr error
		result, banErr = store.BanMember(groupID, targetUserID, actor, role)
		return banErr
	})

	if err != nil {
		s.logger.Error("Не удалось удалить пользователя из группы", "actorID", actorID, "targetUserID", targetUserID, "error", err)
		return false, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
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

	var result groupAdminResult

	err = s.runInTx(func(tx groupTx) error {
		store := newGroupAdminStore(tx)
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var removeErr error
		result, removeErr = store.RemoveFromBlacklist(groupID, targetUserID, actor, role)
		return removeErr
	})

	if err != nil {
		s.logger.Error("Не удалось убрать пользователя из черного списка", "actorID", actorID, "targetUserID", targetUserID, "error", err)
		return false, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Пользователь убран из черного списка", "actorID", actorID, "targetUserID", targetUserID, "groupID", groupID)
	return true, nil
}

// WatchRecentActions получает историю действий в группе
func (s *groupService) WatchRecentActions(userID uint, groupID uint, filter GroupActionFilter) ([]GroupAction, error) {
	// Проверяем, что пользователь в группе
	hasAccess, _, err := s.checkGroupAccess(userID, groupID, groups.CapabilityMember)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Order != "asc" {
		filter.Order = "desc"
	}

	actions, err := s.reads.ListGroupActions(groupID, filter)
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

// logAction записывает действие в лог
func (s *groupService) logAction(tx groupTx, groupID uint, userID uint, username, us, role, actionType, description string) error {
	return s.logActionRecord(tx, groupID, userID, username, us, role, actionType, description, nil)
}

func (s *groupService) logActionWithTargetUser(tx groupTx, groupID uint, userID uint, username, us, role, actionType, description string, targetUserID uint) error {
	return s.logActionRecord(tx, groupID, userID, username, us, role, actionType, description, &targetUserID)
}

func (s *groupService) logActionRecord(tx groupTx, groupID uint, userID uint, username, us, role, actionType, description string, targetUserID *uint) error {
	actionTypeID, err := groups.FindGroupActionTypeID(tx, actionType)
	if err != nil {
		return fmt.Errorf("тип действия группы %q не найден: %w", actionType, err)
	}

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
		GroupID:      groupID,
		UserID:       userID,
		Username:     username,
		Us:           us,
		Role:         role,
		ActionTypeID: actionTypeID,
		Description:  description,
		TargetUserID: targetUserID,
		CreatedAt:    time.Now(),
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
