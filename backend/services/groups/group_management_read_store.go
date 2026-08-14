package group

import (
	"context"
	"errors"
	"fmt"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	"friendship/models/groups"
	"friendship/repository"
	"time"

	"gorm.io/gorm"
)

type groupManagementReadStore interface {
	GetGroupDetails(ctx context.Context, userID uint, groupID uint) (*dto.GroupFullDto, error)
	GetManagedGroups(ctx context.Context, userID uint) (*dto.ManagedGroupsDto, error)
	ListJoinRequests(ctx context.Context, groupID uint, status string, limit int) ([]JoinRequestInfo, error)
	ListGroupBlacklist(ctx context.Context, groupID uint, limit int) ([]BlacklistUser, error)
	ListGroupActions(ctx context.Context, groupID uint, filter GroupActionFilter) ([]GroupAction, error)
}

type gormGroupManagementReadStore struct {
	store repository.PostgresRepository
}

func newGroupManagementReadStore(store repository.PostgresRepository) groupManagementReadStore {
	return gormGroupManagementReadStore{store: store}
}

func (s gormGroupManagementReadStore) GetGroupDetails(ctx context.Context, userID uint, groupID uint) (*dto.GroupFullDto, error) {
	if ctx == nil {
		return nil, errGroupOperationContextMissing
	}

	var group groups.Group
	err := s.store.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		WithContext(ctx).
		First(&group, groupID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("ошибка получения группы: %w", err)
	}

	if group.IsPrivate {
		isMember, err := s.isGroupMember(ctx, groupID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrPermissionDenied
		}
	}

	totalMembers, err := s.countGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	members, err := s.listGroupMembers(ctx, groupID, 10)
	if err != nil {
		return nil, err
	}

	activeEvents, err := s.listActiveGroupEvents(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	isSubscribed, userRole, err := s.findUserGroupSubscription(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	return convertorsdto.ConvertToGroupFullDto(group, totalMembers, members, activeEvents, isSubscribed, userRole), nil
}

func (s gormGroupManagementReadStore) GetManagedGroups(ctx context.Context, userID uint) (*dto.ManagedGroupsDto, error) {
	if ctx == nil {
		return nil, errGroupOperationContextMissing
	}

	var memberships []groups.GroupUsers
	err := s.store.
		Preload("Group.Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("categories.name ASC, categories.id ASC")
		}).
		Preload("RoleInGroup").
		WithContext(ctx).
		Joins("JOIN role_in_groups ON role_in_groups.id = group_users.role_in_group_id").
		Where("group_users.user_id = ?", userID).
		Where("role_in_groups.name IN ?", []string{groups.RoleAdmin, groups.RoleModerator}).
		Order("group_users.group_id DESC").
		Find(&memberships).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка получения управляемых групп: %w", err)
	}

	result := &dto.ManagedGroupsDto{
		Admin:     make([]dto.ManagedGroupItemDto, 0),
		Moderator: make([]dto.ManagedGroupItemDto, 0),
	}
	if len(memberships) == 0 {
		return result, nil
	}

	groupIDs := make([]uint, 0, len(memberships))
	for _, membership := range memberships {
		groupIDs = append(groupIDs, membership.GroupID)
	}

	memberCounts, err := s.countGroupMembersByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	for _, membership := range memberships {
		item := convertorsdto.ConvertToManagedGroupItemDto(membership.Group, memberCounts[membership.GroupID])

		switch groups.NormalizeRoleName(membership.RoleInGroup.Name) {
		case groups.RoleAdmin:
			result.Admin = append(result.Admin, item)
		case groups.RoleModerator:
			result.Moderator = append(result.Moderator, item)
		}
	}

	return result, nil
}

func (s gormGroupManagementReadStore) ListJoinRequests(ctx context.Context, groupID uint, status string, limit int) ([]JoinRequestInfo, error) {
	if ctx == nil {
		return nil, errGroupOperationContextMissing
	}

	query := s.store.
		Preload("User").
		WithContext(ctx).
		Where("group_id = ?", groupID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var requests []groups.GroupJoinRequest
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&requests).Error
	if err != nil {
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

func (s gormGroupManagementReadStore) ListGroupBlacklist(ctx context.Context, groupID uint, limit int) ([]BlacklistUser, error) {
	if ctx == nil {
		return nil, errGroupOperationContextMissing
	}

	var blacklist []groups.GroupBlacklist
	err := s.store.
		Preload("User").
		Preload("Banned").
		WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).
		Find(&blacklist).Error
	if err != nil {
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
			BannedByName: bl.Banned.Name,
			Reason:       bl.Reason,
			CreatedAt:    bl.CreatedAt,
		})
	}

	return result, nil
}

func (s gormGroupManagementReadStore) isGroupMember(ctx context.Context, groupID uint, userID uint) (bool, error) {
	var memberCount int64
	result := s.store.Model(&groups.GroupUsers{}).
		WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&memberCount)
	if result.Error != nil {
		return false, fmt.Errorf("ошибка проверки доступа к группе: %w", result.Error)
	}

	return memberCount > 0, nil
}

func (s gormGroupManagementReadStore) countGroupMembers(ctx context.Context, groupID uint) (int64, error) {
	var totalMembers int64
	result := s.store.Model(&groups.GroupUsers{}).
		WithContext(ctx).
		Where("group_id = ?", groupID).
		Count(&totalMembers)
	if result.Error != nil {
		return 0, fmt.Errorf("ошибка подсчета участников группы: %w", result.Error)
	}

	return totalMembers, nil
}

func (s gormGroupManagementReadStore) countGroupMembersByGroupIDs(ctx context.Context, groupIDs []uint) (map[uint]int64, error) {
	type memberCountRow struct {
		GroupID     uint
		MemberCount int64
	}

	rows := make([]memberCountRow, 0, len(groupIDs))
	if err := s.store.Model(&groups.GroupUsers{}).
		WithContext(ctx).
		Select("group_id, COUNT(*) AS member_count").
		Where("group_id IN ?", groupIDs).
		Group("group_id").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("ошибка подсчета участников управляемых групп: %w", err)
	}

	counts := make(map[uint]int64, len(groupIDs))
	for _, row := range rows {
		counts[row.GroupID] = row.MemberCount
	}
	for _, groupID := range groupIDs {
		if _, exists := counts[groupID]; !exists {
			counts[groupID] = 0
		}
	}

	return counts, nil
}

func (s gormGroupManagementReadStore) listGroupMembers(ctx context.Context, groupID uint, limit int) ([]dto.GroupMemberDto, error) {
	var groupUsers []groups.GroupUsers
	err := s.store.
		Preload("User").
		Preload("RoleInGroup").
		WithContext(ctx).
		Where("group_id = ?", groupID).
		Limit(limit).
		Find(&groupUsers).Error
	if err != nil {
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

	return members, nil
}

func (s gormGroupManagementReadStore) listActiveGroupEvents(ctx context.Context, groupID uint, userID uint) ([]dto.EventSearchItemDto, error) {
	var activeEvents []events.Event
	err := s.store.
		Preload("Group").
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Preload("Users", "user_id = ?", userID).
		WithContext(ctx).
		Where("group_id = ? AND status_id IN (?)", groupID, []uint{1, 2}).
		Where("start_time > ?", time.Now()).
		Order("start_time ASC").
		Find(&activeEvents).Error
	if err != nil {
		return nil, fmt.Errorf("list active group events: %w", err)
	}

	return convertorsdto.ConvertManyToSearchItemDtoForUser(activeEvents, userID), nil
}

func (s gormGroupManagementReadStore) findUserGroupSubscription(ctx context.Context, groupID uint, userID uint) (bool, string, error) {
	var userGroupMembership groups.GroupUsers
	err := s.store.
		Preload("RoleInGroup").
		WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&userGroupMembership).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("find user group subscription: %w", err)
	}

	return true, userGroupMembership.RoleInGroup.Name, nil
}

func (s gormGroupManagementReadStore) ListGroupActions(ctx context.Context, groupID uint, filter GroupActionFilter) ([]GroupAction, error) {
	if ctx == nil {
		return nil, errGroupOperationContextMissing
	}

	var logs []groups.GroupActionLog
	query := s.store.
		Preload("ActionType").
		Preload("User").
		Preload("TargetUser").
		WithContext(ctx).
		Where("group_id = ?", groupID)

	if filter.ActionTypeID != 0 {
		query = query.Where("action_type_id = ?", filter.ActionTypeID)
	}
	if filter.Action != "" {
		actionTypeSubquery := s.store.Model(&groups.GroupActionType{}).
			WithContext(ctx).
			Select("id").
			Where("code = ?", filter.Action)
		query = query.Where("action_type_id = (?)", actionTypeSubquery)
	}

	order := "created_at DESC"
	if filter.Order == "asc" {
		order = "created_at ASC"
	}

	err := query.
		Order(order).
		Limit(filter.Limit).
		Find(&logs).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории действий: %w", err)
	}

	actions := make([]GroupAction, 0, len(logs))
	for _, log := range logs {
		actions = append(actions, groupActionFromLog(log))
	}

	return actions, nil
}

func groupActionFromLog(log groups.GroupActionLog) GroupAction {
	actorName := log.Username
	actorUs := log.Us
	if log.User.ID != 0 {
		actorName = log.User.Name
		actorUs = log.User.Us
	}

	targetName := ""
	targetUs := ""
	if log.TargetUser.ID != 0 {
		targetName = log.TargetUser.Name
		targetUs = log.TargetUser.Us
	}

	return GroupAction{
		ID:           log.ID,
		GroupID:      log.GroupID,
		UserID:       log.UserID,
		Username:     actorName,
		Us:           actorUs,
		Role:         log.Role,
		ActionTypeID: log.ActionTypeID,
		Action:       log.ActionType.Code,
		ActionName:   log.ActionType.Name,
		Description:  formatGroupActionDescription(log, actorName, actorUs, targetName, targetUs),
		TargetUserID: log.TargetUserID,
		TargetName:   targetName,
		TargetUs:     targetUs,
		EntityID:     log.EntityID,
		EntityName:   log.EntityName,
		CreatedAt:    log.CreatedAt,
	}
}

func formatGroupActionDescription(log groups.GroupActionLog, actorName string, actorUs string, targetName string, targetUs string) string {
	actor := actionUserLabel(actorName, actorUs, log.UserID)
	target := ""
	if log.TargetUserID != nil {
		target = actionUserLabel(targetName, targetUs, *log.TargetUserID)
	}

	switch log.ActionType.Code {
	case groups.ActionCreateGroup:
		return fmt.Sprintf("%s создал группу", actor)
	case groups.ActionUpdateGroup:
		return fmt.Sprintf("%s изменил информацию о группе", actor)
	case groups.ActionJoinGroup:
		return fmt.Sprintf("%s вступил в группу", actor)
	case groups.ActionCreateJoinRequest:
		return fmt.Sprintf("%s отправил заявку на вступление в группу", actor)
	case groups.ActionLeaveGroup:
		return fmt.Sprintf("%s вышел из группы", actor)
	case groups.ActionSendInvite:
		return fmt.Sprintf("%s отправил приглашение пользователю %s", actor, target)
	case groups.ActionAcceptInvite:
		return fmt.Sprintf("%s принял приглашение в группу", actor)
	case groups.ActionRejectInvite:
		return fmt.Sprintf("%s отклонил приглашение в группу", actor)
	case groups.ActionApproveRequest:
		return fmt.Sprintf("%s одобрил заявку пользователя %s", actor, target)
	case groups.ActionRejectRequest:
		return fmt.Sprintf("%s отклонил заявку пользователя %s", actor, target)
	case groups.ActionApproveAllRequests:
		if log.Description != "" {
			return fmt.Sprintf("%s: %s", actor, log.Description)
		}
		return fmt.Sprintf("%s одобрил все ожидающие заявки", actor)
	case groups.ActionRejectAllRequests:
		if log.Description != "" {
			return fmt.Sprintf("%s: %s", actor, log.Description)
		}
		return fmt.Sprintf("%s отклонил все ожидающие заявки", actor)
	case groups.ActionAddOperator:
		return fmt.Sprintf("%s назначил пользователя %s оператором группы", actor, target)
	case groups.ActionRemoveOperator:
		return fmt.Sprintf("%s снял с пользователя %s права оператора", actor, target)
	case groups.ActionBanUser:
		return fmt.Sprintf("%s удалил пользователя %s из группы и добавил в черный список", actor, target)
	case groups.ActionUnbanUser:
		return fmt.Sprintf("%s убрал пользователя %s из черного списка", actor, target)
	case groups.ActionCreateEvent:
		return fmt.Sprintf("%s создал событие %s", actor, actionEntityLabel(log))
	case groups.ActionUpdateEvent:
		return fmt.Sprintf("%s изменил событие %s", actor, actionEntityLabel(log))
	case groups.ActionDeleteEvent:
		return fmt.Sprintf("%s удалил событие %s", actor, actionEntityLabel(log))
	case groups.ActionJoinEvent:
		return fmt.Sprintf("%s вступил в событие %s", actor, actionEntityLabel(log))
	case groups.ActionLeaveEvent:
		return fmt.Sprintf("%s вышел из события %s", actor, actionEntityLabel(log))
	case groups.ActionKickFromEvent:
		return fmt.Sprintf("%s исключил пользователя %s из события %s", actor, target, actionEntityLabel(log))
	default:
		if log.ActionType.Name != "" {
			return fmt.Sprintf("%s: %s", actor, log.ActionType.Name)
		}
		return actor
	}
}

func actionEntityLabel(log groups.GroupActionLog) string {
	if log.EntityName != "" && log.EntityID != nil {
		return fmt.Sprintf("'%s' (ID: %d)", log.EntityName, *log.EntityID)
	}
	if log.EntityName != "" {
		return fmt.Sprintf("'%s'", log.EntityName)
	}
	if log.EntityID != nil {
		return fmt.Sprintf("ID: %d", *log.EntityID)
	}
	return ""
}

func actionUserLabel(name string, us string, userID uint) string {
	if name != "" && us != "" {
		return fmt.Sprintf("'%s' (@%s)", name, us)
	}
	if name != "" {
		return fmt.Sprintf("'%s' (#%d)", name, userID)
	}
	if us != "" {
		return fmt.Sprintf("@%s (#%d)", us, userID)
	}
	return fmt.Sprintf("пользователь #%d", userID)
}
