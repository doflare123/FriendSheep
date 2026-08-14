package events

import (
	"context"
	"errors"
	"fmt"
	"strings"

	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

type gormEventReadStore struct {
	repo repository.PostgresRepository
}

func NewGORMEventReadStore(repo repository.PostgresRepository) EventReadStore {
	if repo == nil {
		return nil
	}
	return &gormEventReadStore{repo: repo}
}

func (s *gormEventReadStore) SearchEvents(ctx context.Context, query EventSearchQuery) (EventSearchPageView, error) {
	countQuery := s.buildSearchQuery(ctx, query)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return EventSearchPageView{}, fmt.Errorf("ошибка поиска событий: %w", err)
	}

	offset64 := int64(query.Page-1) * int64(query.Limit)
	if offset64 > int64(maxIntValue()) {
		return EventSearchPageView{}, fmt.Errorf("ошибка поиска событий: слишком большое смещение страницы")
	}

	itemsQuery := s.buildSearchQuery(ctx, query).
		Preload("Group").
		Preload("EventType").
		Preload("EventLocation").
		Preload("AgeLimit").
		Preload("Status").
		Preload("Genres.Genre")
	if query.ViewerID != 0 {
		itemsQuery = itemsQuery.Preload("Users", "user_id = ?", query.ViewerID)
	}

	var found []eventmodels.Event
	if err := itemsQuery.
		Order(eventSearchOrder(query)).
		Limit(query.Limit).
		Offset(int(offset64)).
		Find(&found).Error; err != nil {
		return EventSearchPageView{}, fmt.Errorf("ошибка поиска событий: %w", err)
	}

	return EventSearchPageView{
		Total: total,
		Items: mapSearchEvents(found, query.ViewerID),
	}, nil
}

func (s *gormEventReadStore) GetGroupEvents(ctx context.Context, query EventGroupEventsQuery) (EventGroupEventsView, error) {
	group, found, err := s.findGroupAccessView(ctx, query.GroupID)
	if err != nil {
		return EventGroupEventsView{}, err
	}
	if !found {
		return EventGroupEventsView{
			Items: []EventSearchItemView{},
		}, nil
	}

	view := EventGroupEventsView{
		GroupFound:   true,
		GroupPrivate: group.IsPrivate,
		Items:        []EventSearchItemView{},
	}

	if group.IsPrivate {
		isMember, err := s.isGroupMember(ctx, query.GroupID, query.ViewerID)
		if err != nil {
			return EventGroupEventsView{}, err
		}
		view.ViewerIsGroupMember = isMember
		if !isMember {
			return view, nil
		}
	}

	eventQuery := s.repo.
		Model(&eventmodels.Event{}).
		WithContext(normalizeReadContext(ctx)).
		Preload("Group").
		Preload("EventType").
		Preload("EventLocation").
		Preload("AgeLimit").
		Preload("Status").
		Preload("Genres.Genre").
		Where("group_id = ?", query.GroupID).
		Order("start_time DESC")
	if query.ViewerID != 0 {
		eventQuery = eventQuery.Preload("Users", "user_id = ?", query.ViewerID)
	}

	var foundEvents []eventmodels.Event
	if err := eventQuery.Find(&foundEvents).Error; err != nil {
		return EventGroupEventsView{}, fmt.Errorf("ошибка получения событий: %w", err)
	}

	view.Items = mapSearchEvents(foundEvents, query.ViewerID)
	return view, nil
}

func (s *gormEventReadStore) GetEventDetails(ctx context.Context, query EventDetailsQuery) (EventDetailsView, error) {
	eventQuery := s.repo.
		Model(&eventmodels.Event{}).
		WithContext(normalizeReadContext(ctx)).
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Group").
		Preload("Creator").
		Preload("AgeLimit").
		Preload("Genres.Genre")
	if query.ViewerID != 0 {
		eventQuery = eventQuery.Preload("Users", "user_id = ?", query.ViewerID)
	}

	var event eventmodels.Event
	if err := eventQuery.First(&event, query.EventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventDetailsView{}, nil
		}
		return EventDetailsView{}, fmt.Errorf("ошибка получения события: %w", err)
	}

	view := EventDetailsView{
		Found:        true,
		GroupPrivate: event.Group.IsPrivate,
		Event:        mapFullEvent(event, query.ViewerID),
	}

	if event.Group.IsPrivate {
		isMember, err := s.isGroupMember(ctx, event.GroupID, query.ViewerID)
		if err != nil {
			return EventDetailsView{}, err
		}
		view.ViewerIsGroupMember = isMember
	}

	return view, nil
}

func (s *gormEventReadStore) buildSearchQuery(ctx context.Context, query EventSearchQuery) *gorm.DB {
	result := s.repo.
		Model(&eventmodels.Event{}).
		WithContext(normalizeReadContext(ctx)).
		Joins("JOIN groups ON groups.id = events.group_id").
		Joins("LEFT JOIN event_locations ON event_locations.id = events.event_location_id")

	result = applyEventSearchVisibility(result, query.ViewerID)
	result = applyEventSearchFilters(result, query)
	return result
}

func applyEventSearchVisibility(query *gorm.DB, viewerID uint) *gorm.DB {
	if viewerID == 0 {
		return query.Where("groups.is_private = ?", false)
	}

	return query.Where(
		"groups.is_private = ? OR EXISTS (SELECT 1 FROM group_users gu_visibility WHERE gu_visibility.group_id = groups.id AND gu_visibility.user_id = ?)",
		false,
		viewerID,
	)
}

func applyEventSearchFilters(query *gorm.DB, input EventSearchQuery) *gorm.DB {
	if input.Query != "" {
		like := "%" + strings.ToLower(input.Query) + "%"
		query = query.Where(
			"LOWER(events.title) LIKE ? OR LOWER(events.description) LIKE ? OR LOWER(groups.name) LIKE ?",
			like,
			like,
			like,
		)
	}

	if input.GroupID != nil {
		query = query.Where("events.group_id = ?", *input.GroupID)
	}
	if len(input.CategoryIDs) > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM group_group_categories ggc WHERE ggc.group_id = groups.id AND ggc.group_category_id IN ?)",
			input.CategoryIDs,
		)
	}
	if len(input.ExcludeCategoryIDs) > 0 {
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM group_group_categories ggc_ex WHERE ggc_ex.group_id = groups.id AND ggc_ex.group_category_id IN ?)",
			input.ExcludeCategoryIDs,
		)
	}
	if len(input.GenreIDs) > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM event_genres eg WHERE eg.event_id = events.id AND eg.genre_id IN ?)",
			input.GenreIDs,
		)
	}
	if len(input.ExcludeGenreIDs) > 0 {
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM event_genres eg_ex WHERE eg_ex.event_id = events.id AND eg_ex.genre_id IN ?)",
			input.ExcludeGenreIDs,
		)
	}
	if len(input.EventTypeIDs) > 0 {
		query = query.Where("events.event_type_id IN ?", input.EventTypeIDs)
	}
	if len(input.ExcludeEventTypeIDs) > 0 {
		query = query.Where("events.event_type_id NOT IN ?", input.ExcludeEventTypeIDs)
	}
	if len(input.LocationTypes) > 0 {
		query = query.Where("LOWER(event_locations.name) IN ?", input.LocationTypes)
	}
	if len(input.ExcludeLocationTypes) > 0 {
		query = query.Where("LOWER(event_locations.name) NOT IN ?", input.ExcludeLocationTypes)
	}
	if input.City != "" {
		query = query.Where("LOWER(groups.city) LIKE ?", "%"+strings.ToLower(input.City)+"%")
	}
	if input.DateFrom != nil {
		query = query.Where("events.start_time >= ?", *input.DateFrom)
	}
	if input.DateTo != nil {
		query = query.Where("events.start_time <= ?", *input.DateTo)
	}
	if input.HasFreeSlots != nil {
		if *input.HasFreeSlots {
			query = query.Where("events.current_users < events.max_users")
		} else {
			query = query.Where("events.current_users >= events.max_users")
		}
	}
	if input.OnlySubscriptionNews {
		query = query.Where(
			"EXISTS (SELECT 1 FROM group_users gu_subscription WHERE gu_subscription.group_id = events.group_id AND gu_subscription.user_id = ?)",
			input.ViewerID,
		)
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM events_users eu_subscription WHERE eu_subscription.event_id = events.id AND eu_subscription.user_id = ?)",
			input.ViewerID,
		)
	}

	return query
}

func eventSearchOrder(input EventSearchQuery) string {
	if input.OnlySubscriptionNews {
		return "events.created_at DESC, events.id DESC"
	}
	return "events.start_time ASC, events.id DESC"
}

func (s *gormEventReadStore) findGroupAccessView(ctx context.Context, groupID uint) (groupmodels.Group, bool, error) {
	var group groupmodels.Group
	err := s.repo.
		Model(&groupmodels.Group{}).
		WithContext(normalizeReadContext(ctx)).
		Select("id", "is_private").
		First(&group, groupID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return groupmodels.Group{}, false, nil
		}
		return groupmodels.Group{}, false, fmt.Errorf("ошибка получения группы: %w", err)
	}

	return group, true, nil
}

func (s *gormEventReadStore) isGroupMember(ctx context.Context, groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.repo.
		Model(&groupmodels.GroupUsers{}).
		WithContext(normalizeReadContext(ctx)).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства в группе: %w", err)
	}

	return count > 0, nil
}

func normalizeReadContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func mapSearchEvents(items []eventmodels.Event, viewerID uint) []EventSearchItemView {
	result := make([]EventSearchItemView, 0, len(items))
	for _, item := range items {
		result = append(result, EventSearchItemView{
			ID:    item.ID,
			Title: item.Title,
			Group: EventReadGroupView{
				ID:         item.Group.ID,
				Name:       item.Group.Name,
				Image:      item.Group.Image,
				Enterprise: item.Group.Enterprise,
				City:       item.Group.City,
			},
			ImageURL:         item.ImageURL,
			CurrentUsers:     item.CurrentUsers,
			MaxUsers:         item.MaxUsers,
			Duration:         item.Duration,
			StartTime:        item.StartTime,
			EventType:        item.EventType.Name,
			LocationType:     item.EventLocation.Name,
			AgeLimit:         item.AgeLimit.Name,
			Status:           item.Status.Name,
			City:             item.Group.City,
			Genres:           eventGenresToNames(item.Genres),
			ViewerSubscribed: eventHasViewerSubscription(item.Users, viewerID),
		})
	}
	return result
}

func mapFullEvent(item eventmodels.Event, viewerID uint) EventFullView {
	return EventFullView{
		ID:           item.ID,
		Title:        item.Title,
		Description:  item.Description,
		ImageURL:     item.ImageURL,
		MaxUsers:     item.MaxUsers,
		CurrentUsers: item.CurrentUsers,
		StartTime:    item.StartTime,
		EndTime:      item.EndTime,
		Duration:     item.Duration,
		Group: EventReadGroupView{
			ID:         item.Group.ID,
			Name:       item.Group.Name,
			Image:      item.Group.Image,
			Enterprise: item.Group.Enterprise,
			City:       item.Group.City,
		},
		EventTypeID:    item.EventType.ID,
		LocationTypeID: item.EventLocation.ID,
		StatusID:       item.Status.ID,
		Genres:         eventGenresToNames(item.Genres),
		Creator: EventReadCreatorView{
			ID:       item.Creator.ID,
			Name:     item.Creator.Name,
			Us:       item.Creator.Us,
			Image:    item.Creator.Image,
			Verified: item.Creator.VerifiedUser,
		},
		Address:          item.Address,
		Country:          item.Country,
		AgeLimit:         item.AgeLimit.Name,
		Year:             item.Year,
		Notes:            item.Notes,
		CustomFields:     mapCustomFields(item.CustomFields),
		ViewerSubscribed: eventHasViewerSubscription(item.Users, viewerID),
		ViewerIsCreator:  item.CreatorID == viewerID,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func eventGenresToNames(genres []eventmodels.EventGenre) []string {
	result := make([]string, 0, len(genres))
	for _, genre := range genres {
		result = append(result, genre.Genre.Name)
	}
	return result
}

func eventHasViewerSubscription(users []eventmodels.EventsUser, viewerID uint) bool {
	if viewerID == 0 {
		return false
	}

	for _, user := range users {
		if user.UserID == viewerID {
			return true
		}
	}
	return false
}

func mapCustomFields(fields eventmodels.CustomFields) map[string]interface{} {
	if fields == nil {
		return nil
	}

	result := make(map[string]interface{}, len(fields))
	for key, value := range fields {
		result[key] = value
	}
	return result
}
