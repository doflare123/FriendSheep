package group

import (
	"fmt"
	"friendship/models/groups"
)

type groupManagementReadStore interface {
	ListJoinRequests(groupID uint, status string, limit int) ([]JoinRequestInfo, error)
	ListGroupBlacklist(groupID uint, limit int) ([]BlacklistUser, error)
	ListGroupActions(groupID uint, filter GroupActionFilter) ([]GroupAction, error)
}

type gormGroupManagementReadStore struct {
	store groupStore
}

func newGroupManagementReadStore(store groupStore) groupManagementReadStore {
	return gormGroupManagementReadStore{store: store}
}

func (s gormGroupManagementReadStore) ListJoinRequests(groupID uint, status string, limit int) ([]JoinRequestInfo, error) {
	query := s.store.
		Preload("User").
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

func (s gormGroupManagementReadStore) ListGroupBlacklist(groupID uint, limit int) ([]BlacklistUser, error) {
	var blacklist []groups.GroupBlacklist
	err := s.store.
		Preload("User").
		Preload("Banned").
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

func (s gormGroupManagementReadStore) ListGroupActions(groupID uint, filter GroupActionFilter) ([]GroupAction, error) {
	var logs []groups.GroupActionLog
	query := s.store.
		Preload("ActionType").
		Preload("User").
		Preload("TargetUser").
		Where("group_id = ?", groupID)

	if filter.ActionTypeID != 0 {
		query = query.Where("action_type_id = ?", filter.ActionTypeID)
	}
	if filter.Action != "" {
		actionTypeSubquery := s.store.Model(&groups.GroupActionType{}).
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
