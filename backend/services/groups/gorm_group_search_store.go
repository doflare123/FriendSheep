package group

import (
	"context"
	"fmt"

	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

type gormGroupSearchStore struct {
	store repository.PostgresRepository
}

func newGORMGroupSearchStore(store repository.PostgresRepository) groupSearchStore {
	return gormGroupSearchStore{store: store}
}

func (s gormGroupSearchStore) SearchGroups(ctx context.Context, query groupSearchQuery) (groupSearchPage, error) {
	if ctx == nil {
		return groupSearchPage{}, errGroupOperationContextMissing
	}

	base := s.buildQuery(ctx, query)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return groupSearchPage{}, fmt.Errorf("count group search results: %w", err)
	}

	page := groupSearchPage{Items: make([]groupSearchItemView, 0), Total: total}
	if total == 0 {
		return page, nil
	}

	var found []groups.Group
	if err := s.applyOrder(base, query).
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("categories.name ASC, categories.id ASC")
		}).
		Offset((query.Page - 1) * query.Limit).
		Limit(query.Limit).
		Find(&found).Error; err != nil {
		return groupSearchPage{}, fmt.Errorf("search groups: %w", err)
	}

	groupIDs := make([]uint, 0, len(found))
	for _, group := range found {
		groupIDs = append(groupIDs, group.ID)
	}
	memberCounts, err := s.countMembers(ctx, groupIDs)
	if err != nil {
		return groupSearchPage{}, err
	}
	viewerSubscriptions, err := s.countViewerSubscriptions(ctx, query.ViewerID, groupIDs)
	if err != nil {
		return groupSearchPage{}, err
	}

	page.Items = make([]groupSearchItemView, 0, len(found))
	for _, group := range found {
		categories := make([]string, 0, len(group.Categories))
		for _, category := range group.Categories {
			categories = append(categories, category.Name)
		}
		page.Items = append(page.Items, groupSearchItemView{
			ID:               group.ID,
			Name:             group.Name,
			Categories:       categories,
			MemberCount:      memberCounts[group.ID],
			ViewerSubscribed: viewerSubscriptions[group.ID],
			Image:            group.Image,
			CreatedAt:        group.CreatedAt,
			SmallDescription: group.SmallDescription,
			IsPrivate:        group.IsPrivate,
			Enterprise:       group.Enterprise,
		})
	}

	return page, nil
}

func (s gormGroupSearchStore) buildQuery(ctx context.Context, input groupSearchQuery) *gorm.DB {
	query := s.store.Model(&groups.Group{}).WithContext(ctx)
	if input.Query != "" {
		like := "%" + input.Query + "%"
		query = query.Where(
			"LOWER(groups.name) LIKE LOWER(?) OR LOWER(groups.small_description) LIKE LOWER(?) OR LOWER(groups.description) LIKE LOWER(?)",
			like,
			like,
			like,
		)
	}
	if len(input.CategoryIDs) > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM group_group_categories ggc WHERE ggc.group_id = groups.id AND ggc.group_category_id IN ?)",
			input.CategoryIDs,
		)
	}
	if input.IsPrivate != nil {
		query = query.Where("groups.is_private = ?", *input.IsPrivate)
	}
	if input.City != "" {
		query = query.Where("LOWER(groups.city) LIKE LOWER(?)", "%"+input.City+"%")
	}
	return query
}

func (s gormGroupSearchStore) applyOrder(query *gorm.DB, input groupSearchQuery) *gorm.DB {
	direction := "DESC"
	if input.SortOrder == GroupSearchOrderAscending {
		direction = "ASC"
	}

	switch input.SortBy {
	case GroupSearchSortMemberCount:
		query = query.Order("(SELECT COUNT(*) FROM group_users gu WHERE gu.group_id = groups.id) " + direction)
	case GroupSearchSortName:
		query = query.Order("LOWER(groups.name) " + direction)
	default:
		query = query.Order("groups.created_at " + direction)
	}
	return query.Order("groups.id " + direction)
}

func (s gormGroupSearchStore) countMembers(ctx context.Context, groupIDs []uint) (map[uint]int64, error) {
	counts := make(map[uint]int64, len(groupIDs))
	if len(groupIDs) == 0 {
		return counts, nil
	}

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
		return nil, fmt.Errorf("count group search members: %w", err)
	}

	for _, row := range rows {
		counts[row.GroupID] = row.MemberCount
	}
	return counts, nil
}

func (s gormGroupSearchStore) countViewerSubscriptions(ctx context.Context, viewerID uint, groupIDs []uint) (map[uint]bool, error) {
	subscriptions := make(map[uint]bool, len(groupIDs))
	if viewerID == 0 || len(groupIDs) == 0 {
		return subscriptions, nil
	}

	type viewerGroupRow struct {
		GroupID uint
	}
	rows := make([]viewerGroupRow, 0, len(groupIDs))
	if err := s.store.Model(&groups.GroupUsers{}).
		WithContext(ctx).
		Select("group_id").
		Where("user_id = ? AND group_id IN ?", viewerID, groupIDs).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("load group search viewer subscriptions: %w", err)
	}

	for _, row := range rows {
		subscriptions[row.GroupID] = true
	}
	return subscriptions, nil
}
