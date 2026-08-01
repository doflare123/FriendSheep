package group

import (
	"context"
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models/dto"
	"friendship/models/groups"
	"friendship/services"
	"strings"
	"time"
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
	CreateGroup(ctx context.Context, id uint, inf CreateGroupInput) (*dto.GroupFullDto, error)
	UpdateGroup(ctx context.Context, actorID uint, inf GroupUpdateInput) (*dto.GroupFullDto, error)
	DeleteGroup(ctx context.Context, actorID uint, groupID uint) (bool, error)
	GetGroupDetails(ctx context.Context, userID uint, groupID uint) (*dto.GroupFullDto, error)
	GetManagedGroups(ctx context.Context, userID uint) (*dto.ManagedGroupsDto, error)
	GetSubscribedGroups(ctx context.Context, userID uint, page int, limit int) (*dto.SubscribedGroupsResponseDto, error)

	// Управление заявками
	ApproveAllJoinRequests(ctx context.Context, actorID uint, groupID uint) (int, error)
	RejectAllJoinRequests(ctx context.Context, actorID uint, groupID uint) (int, error)
	ApproveJoinRequest(ctx context.Context, actorID uint, requestID uint) (bool, error)
	RejectJoinRequest(ctx context.Context, actorID uint, requestID uint) (bool, error)
	GetJoinRequests(ctx context.Context, actorID uint, groupID uint, status string, limit int) ([]JoinRequestInfo, error)

	// Вступление/выход
	JoinGroup(ctx context.Context, userID uint, groupID uint) (*GroupResult, error)
	LeaveGroup(ctx context.Context, userID uint, groupID uint) (bool, error)

	// Управление правами
	AddPermissions(ctx context.Context, actorID uint, input GroupUserInput) (bool, error)
	RemovePermissions(ctx context.Context, actorID uint, input GroupUserInput) (bool, error)

	// Управление участниками
	DeleteUserFromGroup(ctx context.Context, actorID uint, groupID uint, targetUserID uint) (bool, error)
	RemoveFromBlacklist(ctx context.Context, actorID uint, groupID uint, targetUserID uint) (bool, error)
	GetGroupBlacklist(ctx context.Context, actorID uint, groupID uint, limit int) ([]BlacklistUser, error)

	// Приглашения
	CreateJoinInvite(ctx context.Context, actorID uint, input GroupUserInput) (bool, error)
	AcceptJoinInvite(ctx context.Context, userID uint, inviteID uint) (*GroupResult, error)
	RejectJoinInvite(ctx context.Context, userID uint, inviteID uint) (bool, error)

	// История действий
	WatchRecentActions(ctx context.Context, userID uint, groupID uint, filter GroupActionFilter) ([]GroupAction, error)
}

type groupService struct {
	logger        logger.Logger
	uow           groupUnitOfWork
	access        rootGroupActorRoleFinder
	reads         groupManagementReadStore
	subscriptions groupSubscriptionsStore
}

func NewGroupService(logger logger.Logger, uow groupUnitOfWork) GroupsService {
	return &groupService{
		logger:        logger,
		uow:           uow,
		access:        uow.Access(),
		reads:         uow.Reads(),
		subscriptions: uow.Subscriptions(),
	}
}

// GetGroupDetails получает полную информацию о группе
func (s *groupService) GetGroupDetails(ctx context.Context, userID uint, groupID uint) (*dto.GroupFullDto, error) {
	if userID == 0 || groupID == 0 {
		return nil, ErrInvalidInput
	}

	return s.reads.GetGroupDetails(ctx, userID, groupID)
}

func (s *groupService) GetManagedGroups(ctx context.Context, userID uint) (*dto.ManagedGroupsDto, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	managedGroups, err := s.reads.GetManagedGroups(ctx, userID)
	if err != nil {
		s.logger.Error("Не удалось получить управляемые группы пользователя", "userID", userID, "error", err)
		return nil, err
	}

	return managedGroups, nil
}

// CreateGroup создает новую группу
func (s *groupService) CreateGroup(ctx context.Context, id uint, input CreateGroupInput) (*dto.GroupFullDto, error) {
	if err := services.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %w", err)
	}

	var contacts map[string]string
	if input.Contacts != "" {
		contacts = parseContacts(input.Contacts)
	}

	var result groupCreateResult
	var groupDto *dto.GroupFullDto
	var detailsErr error
	err := s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
		var createErr error
		result, createErr = store.CreateGroup(id, input, contacts)
		if createErr != nil {
			return createErr
		}

		groupDto, detailsErr = tx.Reads().GetGroupDetails(ctx, id, result.GroupID)
		return detailsErr
	})

	if err != nil {
		if detailsErr != nil {
			s.logger.Error("Не удалось сформировать полный DTO группы после создания", "groupID", result.GroupID, "error", detailsErr)
			return nil, fmt.Errorf("ошибка формирования данных группы: %w", detailsErr)
		}
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

	return groupDto, nil
}

// UpdateGroup обновляет группу
func (s *groupService) UpdateGroup(ctx context.Context, actorID uint, input GroupUpdateInput) (*dto.GroupFullDto, error) {
	hasAccess, role, err := s.checkGroupAccess(ctx, actorID, input.GroupID, groups.CapabilityModerate)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	var result groupAdminResult
	var groupDto *dto.GroupFullDto
	var detailsErr error

	err = s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
		actor, err := store.FindActor(actorID)
		if err != nil {
			return err
		}

		var updateErr error
		result, updateErr = store.UpdateGroup(input, actor, role)
		if updateErr != nil {
			return updateErr
		}

		groupDto, detailsErr = tx.Reads().GetGroupDetails(ctx, actorID, input.GroupID)
		return detailsErr
	})

	if err != nil {
		if detailsErr != nil {
			s.logger.Error("Не удалось сформировать полный DTO группы после обновления", "groupID", input.GroupID, "error", detailsErr)
			return nil, fmt.Errorf("ошибка формирования данных группы: %w", detailsErr)
		}
		s.logger.Error("Не удалось обновить группу", "groupID", input.GroupID, "error", err)
		return nil, err
	}

	for _, warning := range result.Warnings {
		s.logger.Warn(warning.Message, warning.Args...)
	}

	s.logger.Info("Группа успешно обновлена", "groupID", input.GroupID, "actorID", actorID)

	return groupDto, nil
}

// DeleteGroup удаляет группу (только админ)
func (s *groupService) DeleteGroup(ctx context.Context, actorID uint, groupID uint) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(ctx, actorID, groupID, groups.CapabilityAdmin)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	err = s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
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
func (s *groupService) AddPermissions(ctx context.Context, actorID uint, input GroupUserInput) (bool, error) {
	if actorID == input.UserID {
		return false, ErrCannotChangeOwnRole
	}

	hasAccess, role, err := s.checkGroupAccess(ctx, actorID, input.GroupID, groups.CapabilityAdmin)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var result groupAdminResult

	err = s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
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
func (s *groupService) RemovePermissions(ctx context.Context, actorID uint, input GroupUserInput) (bool, error) {
	if actorID == input.UserID {
		return false, ErrCannotChangeOwnRole
	}

	hasAccess, role, err := s.checkGroupAccess(ctx, actorID, input.GroupID, groups.CapabilityAdmin)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var result groupAdminResult

	err = s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
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
func (s *groupService) DeleteUserFromGroup(ctx context.Context, actorID uint, groupID uint, targetUserID uint) (bool, error) {
	if actorID == targetUserID {
		return false, ErrCannotRemoveSelf
	}

	hasAccess, role, err := s.checkGroupAccess(ctx, actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var result groupAdminResult

	err = s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
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
func (s *groupService) RemoveFromBlacklist(ctx context.Context, actorID uint, groupID uint, targetUserID uint) (bool, error) {
	hasAccess, role, err := s.checkGroupAccess(ctx, actorID, groupID, groups.CapabilityModerate)
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	var result groupAdminResult

	err = s.runInTx(ctx, func(tx groupTx) error {
		store := tx.Admin()
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
func (s *groupService) WatchRecentActions(ctx context.Context, userID uint, groupID uint, filter GroupActionFilter) ([]GroupAction, error) {
	// Проверяем, что пользователь в группе
	hasAccess, _, err := s.checkGroupAccess(ctx, userID, groupID, groups.CapabilityMember)
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

	actions, err := s.reads.ListGroupActions(ctx, groupID, filter)
	if err != nil {
		s.logger.Error("Не удалось получить историю действий группы", "groupID", groupID, "error", err)
		return nil, err
	}

	return actions, nil
}

// Вспомогательные функции

// checkGroupAccess проверяет, имеет ли пользователь доступ к группе с нужной ролью
func (s *groupService) checkGroupAccess(ctx context.Context, userID uint, groupID uint, required groups.Capability) (bool, string, error) {
	return s.access.FindActorRole(ctx, userID, groupID, required)
}

// GetGroupBlacklist получает черный список группы
func (s *groupService) GetGroupBlacklist(ctx context.Context, actorID uint, groupID uint, limit int) ([]BlacklistUser, error) {
	// Проверяем права доступа (admin или operator)
	hasAccess, _, err := s.checkGroupAccess(ctx, actorID, groupID, groups.CapabilityModerate)
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

	blacklist, err := s.reads.ListGroupBlacklist(ctx, groupID, limit)
	if err != nil {
		s.logger.Error("Не удалось получить черный список", "groupID", groupID, "error", err)
		return nil, err
	}

	return blacklist, nil
}

func (s *groupService) logActionWithTargetUser(tx groupTx, groupID uint, userID uint, username, us, role, actionType, description string, targetUserID uint) error {
	return tx.LogActorAction(groupActorActionLogInput{
		GroupID:      groupID,
		UserID:       userID,
		Username:     username,
		Us:           us,
		Role:         role,
		Action:       actionType,
		Description:  description,
		TargetUserID: &targetUserID,
	})
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
