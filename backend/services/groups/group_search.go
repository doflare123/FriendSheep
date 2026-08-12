package group

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"friendship/models/dto"
)

const (
	DefaultGroupSearchPage  = 1
	DefaultGroupSearchLimit = 20
	MaxGroupSearchLimit     = 100
	MaxGroupSearchPage      = 10000

	GroupSearchSortCreatedAt   = "createdAt"
	GroupSearchSortMemberCount = "memberCount"
	GroupSearchSortName        = "name"

	GroupSearchOrderAscending  = "asc"
	GroupSearchOrderDescending = "desc"

	maxGroupSearchQueryLength = 200
	maxGroupSearchCityLength  = 100
	maxGroupSearchCategories  = 100
)

var ErrInvalidGroupSearchInput = fmt.Errorf("%w: параметры поиска групп", ErrInvalidInput)

type GroupSearchInput struct {
	Query       string
	CategoryIDs []uint
	IsPrivate   *bool
	City        string
	SortBy      string
	SortOrder   string
	Page        int
	Limit       int
}

type groupSearchQuery struct {
	ViewerID    uint
	Query       string
	CategoryIDs []uint
	IsPrivate   *bool
	City        string
	SortBy      string
	SortOrder   string
	Page        int
	Limit       int
}

type groupSearchItemView struct {
	ID               uint
	Name             string
	Categories       []string
	MemberCount      int64
	ViewerSubscribed bool
	Image            string
	CreatedAt        time.Time
	SmallDescription string
	IsPrivate        bool
	Enterprise       bool
}

type groupSearchPage struct {
	Items []groupSearchItemView
	Total int64
}

type groupSearchStore interface {
	SearchGroups(ctx context.Context, query groupSearchQuery) (groupSearchPage, error)
}

func (s *groupService) SearchGroups(ctx context.Context, userID uint, input GroupSearchInput) (*dto.GroupSearchResponseDto, error) {
	query, err := normalizeGroupSearchInput(input)
	if err != nil {
		return nil, err
	}
	query.ViewerID = userID

	page, err := s.search.SearchGroups(ctx, query)
	if err != nil {
		s.logger.Error("Не удалось выполнить поиск групп", "error", err)
		return nil, err
	}

	items := make([]dto.GroupSearchItemDto, 0, len(page.Items))
	for _, item := range page.Items {
		categories := item.Categories
		if categories == nil {
			categories = make([]string, 0)
		}
		items = append(items, dto.GroupSearchItemDto{
			ID:               item.ID,
			Name:             item.Name,
			Categories:       categories,
			MemberCount:      int(item.MemberCount),
			IsSubscribed:     item.ViewerSubscribed,
			Image:            item.Image,
			CreatedAt:        item.CreatedAt,
			SmallDescription: item.SmallDescription,
			IsPrivate:        item.IsPrivate,
			Enterprise:       item.Enterprise,
		})
	}

	totalPages := 0
	if page.Total > 0 {
		totalPages = int((page.Total + int64(query.Limit) - 1) / int64(query.Limit))
	}
	return &dto.GroupSearchResponseDto{
		Items:      items,
		Total:      page.Total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
		HasMore:    int64(query.Page)*int64(query.Limit) < page.Total,
	}, nil
}

func normalizeGroupSearchInput(input GroupSearchInput) (groupSearchQuery, error) {
	input.Query = strings.TrimSpace(input.Query)
	input.City = strings.TrimSpace(input.City)
	if input.SortBy == "" {
		input.SortBy = GroupSearchSortCreatedAt
	}
	if input.SortOrder == "" {
		input.SortOrder = GroupSearchOrderDescending
	}
	if input.Page == 0 {
		input.Page = DefaultGroupSearchPage
	}
	if input.Limit == 0 {
		input.Limit = DefaultGroupSearchLimit
	}

	if input.Page < 1 || input.Page > MaxGroupSearchPage || input.Limit < 1 || input.Limit > MaxGroupSearchLimit {
		return groupSearchQuery{}, ErrInvalidGroupSearchInput
	}
	if utf8.RuneCountInString(input.Query) > maxGroupSearchQueryLength || utf8.RuneCountInString(input.City) > maxGroupSearchCityLength {
		return groupSearchQuery{}, ErrInvalidGroupSearchInput
	}
	switch input.SortBy {
	case GroupSearchSortCreatedAt, GroupSearchSortMemberCount, GroupSearchSortName:
	default:
		return groupSearchQuery{}, ErrInvalidGroupSearchInput
	}
	switch input.SortOrder {
	case GroupSearchOrderAscending, GroupSearchOrderDescending:
	default:
		return groupSearchQuery{}, ErrInvalidGroupSearchInput
	}
	if len(input.CategoryIDs) > maxGroupSearchCategories {
		return groupSearchQuery{}, ErrInvalidGroupSearchInput
	}
	seenCategoryIDs := make(map[uint]struct{}, len(input.CategoryIDs))
	for _, categoryID := range input.CategoryIDs {
		if categoryID == 0 {
			return groupSearchQuery{}, ErrInvalidGroupSearchInput
		}
		if _, exists := seenCategoryIDs[categoryID]; exists {
			return groupSearchQuery{}, ErrInvalidGroupSearchInput
		}
		seenCategoryIDs[categoryID] = struct{}{}
	}

	return groupSearchQuery{
		Query:       input.Query,
		CategoryIDs: append([]uint(nil), input.CategoryIDs...),
		IsPrivate:   input.IsPrivate,
		City:        input.City,
		SortBy:      input.SortBy,
		SortOrder:   input.SortOrder,
		Page:        input.Page,
		Limit:       input.Limit,
	}, nil
}
