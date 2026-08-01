package group

import (
	"context"

	"friendship/models/dto"
)

const (
	DefaultSubscribedGroupsPage  = 1
	DefaultSubscribedGroupsLimit = 20
	MaxSubscribedGroupsLimit     = 100
)

type groupSubscriptionsQuery struct {
	UserID uint
	Page   int
	Limit  int
}

type groupSubscriptionsPage struct {
	Items []dto.ManagedGroupItemDto
	Total int64
}

type groupSubscriptionsStore interface {
	ListSubscribedGroups(ctx context.Context, query groupSubscriptionsQuery) (groupSubscriptionsPage, error)
}

func (s *groupService) GetSubscribedGroups(
	ctx context.Context,
	userID uint,
	page int,
	limit int,
) (*dto.SubscribedGroupsResponseDto, error) {
	maxInt := int(^uint(0) >> 1)
	if userID == 0 || page < 1 || limit < 1 || limit > MaxSubscribedGroupsLimit || page-1 > maxInt/limit {
		return nil, ErrInvalidInput
	}

	result, err := s.subscriptions.ListSubscribedGroups(ctx, groupSubscriptionsQuery{
		UserID: userID,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		s.logger.Error(
			"Не удалось получить подписки пользователя на группы",
			"userID", userID,
			"page", page,
			"limit", limit,
			"error", err,
		)
		return nil, err
	}

	items := result.Items
	if items == nil {
		items = make([]dto.ManagedGroupItemDto, 0)
	}

	hasMore := result.Total > 0 && int64(page-1) < (result.Total-1)/int64(limit)
	return &dto.SubscribedGroupsResponseDto{
		Items:   items,
		Total:   result.Total,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}, nil
}
