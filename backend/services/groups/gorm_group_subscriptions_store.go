package group

import (
	"context"
	"fmt"

	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

type gormGroupSubscriptionsStore struct {
	store repository.PostgresRepository
}

func newGORMGroupSubscriptionsStore(store repository.PostgresRepository) groupSubscriptionsStore {
	return gormGroupSubscriptionsStore{store: store}
}

func (s gormGroupSubscriptionsStore) ListSubscribedGroups(
	ctx context.Context,
	query groupSubscriptionsQuery,
) (groupSubscriptionsPage, error) {
	if ctx == nil {
		return groupSubscriptionsPage{}, errGroupOperationContextMissing
	}

	var total int64
	if err := s.membershipsQuery(ctx, query.UserID).Count(&total).Error; err != nil {
		return groupSubscriptionsPage{}, fmt.Errorf("count subscribed groups: %w", err)
	}

	page := groupSubscriptionsPage{
		Items: make([]dto.ManagedGroupItemDto, 0),
		Total: total,
	}
	if total == 0 {
		return page, nil
	}

	var memberships []groups.GroupUsers
	err := s.membershipsQuery(ctx, query.UserID).
		Preload("Group.Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("categories.name ASC, categories.id ASC")
		}).
		Order("group_users.group_id DESC").
		Offset((query.Page - 1) * query.Limit).
		Limit(query.Limit).
		Find(&memberships).Error
	if err != nil {
		return groupSubscriptionsPage{}, fmt.Errorf("list subscribed groups: %w", err)
	}
	if len(memberships) == 0 {
		return page, nil
	}

	groupIDs := make([]uint, 0, len(memberships))
	for _, membership := range memberships {
		groupIDs = append(groupIDs, membership.GroupID)
	}

	memberCounts, err := s.countMembers(ctx, groupIDs)
	if err != nil {
		return groupSubscriptionsPage{}, err
	}

	page.Items = make([]dto.ManagedGroupItemDto, 0, len(memberships))
	for _, membership := range memberships {
		page.Items = append(
			page.Items,
			convertorsdto.ConvertToManagedGroupItemDto(
				membership.Group,
				memberCounts[membership.GroupID],
			),
		)
	}

	return page, nil
}

func (s gormGroupSubscriptionsStore) membershipsQuery(ctx context.Context, userID uint) *gorm.DB {
	return s.store.
		Model(&groups.GroupUsers{}).
		WithContext(ctx).
		Joins("JOIN role_in_groups ON role_in_groups.id = group_users.role_in_group_id").
		Where("group_users.user_id = ?", userID).
		Where("role_in_groups.name = ?", groups.RoleMember)
}

func (s gormGroupSubscriptionsStore) countMembers(ctx context.Context, groupIDs []uint) (map[uint]int64, error) {
	type memberCountRow struct {
		GroupID     uint
		MemberCount int64
	}

	rows := make([]memberCountRow, 0, len(groupIDs))
	err := s.store.
		Model(&groups.GroupUsers{}).
		WithContext(ctx).
		Select("group_id, COUNT(*) AS member_count").
		Where("group_id IN ?", groupIDs).
		Group("group_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("count subscribed group members: %w", err)
	}

	counts := make(map[uint]int64, len(groupIDs))
	for _, row := range rows {
		counts[row.GroupID] = row.MemberCount
	}

	return counts, nil
}
