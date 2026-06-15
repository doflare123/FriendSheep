package group

import (
	"fmt"
	"friendship/models/groups"
)

type groupManagementReadStore interface {
	ListJoinRequests(groupID uint, status string, limit int) ([]JoinRequestInfo, error)
	ListGroupBlacklist(groupID uint, limit int) ([]BlacklistUser, error)
	ListGroupActions(groupID uint, limit int) ([]GroupAction, error)
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

func (s gormGroupManagementReadStore) ListGroupActions(groupID uint, limit int) ([]GroupAction, error) {
	var actions []GroupAction
	err := s.store.Model(&groups.GroupActionLog{}).
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).
		Find(&actions).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории действий: %w", err)
	}

	return actions, nil
}
