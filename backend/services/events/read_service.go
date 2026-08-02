package events

import (
	"context"
	"errors"
	"reflect"

	"friendship/logger"
	"friendship/models/dto"
)

var errEventReadStoreUnavailable = errors.New("хранилище чтения событий не настроено")

type EventReadService interface {
	SearchEvents(ctx context.Context, userID uint, input EventSearchInput) (*dto.EventSearchResponse, error)
	GetGroupEvents(ctx context.Context, actorID uint, groupID uint) ([]dto.EventShortDto, error)
	GetEventDetails(ctx context.Context, userID uint, eventID uint) (*dto.EventFullDto, error)
}

type EventReadStore interface {
	SearchEvents(ctx context.Context, query EventSearchQuery) (EventSearchPageView, error)
	GetGroupEvents(ctx context.Context, query EventGroupEventsQuery) (EventGroupEventsView, error)
	GetEventDetails(ctx context.Context, query EventDetailsQuery) (EventDetailsView, error)
}

type eventReadService struct {
	logger logger.Logger
	store  EventReadStore
}

type unsupportedEventReadStore struct{}

func NewEventReadService(logger logger.Logger, store EventReadStore) EventReadService {
	if isNilEventReadStore(store) {
		store = unsupportedEventReadStore{}
	}

	return &eventReadService{
		logger: logger,
		store:  store,
	}
}

func isNilEventReadStore(store EventReadStore) bool {
	if store == nil {
		return true
	}

	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (unsupportedEventReadStore) SearchEvents(context.Context, EventSearchQuery) (EventSearchPageView, error) {
	return EventSearchPageView{}, errEventReadStoreUnavailable
}

func (unsupportedEventReadStore) GetGroupEvents(context.Context, EventGroupEventsQuery) (EventGroupEventsView, error) {
	return EventGroupEventsView{}, errEventReadStoreUnavailable
}

func (unsupportedEventReadStore) GetEventDetails(context.Context, EventDetailsQuery) (EventDetailsView, error) {
	return EventDetailsView{}, errEventReadStoreUnavailable
}

func (s *eventReadService) SearchEvents(ctx context.Context, userID uint, input EventSearchInput) (*dto.EventSearchResponse, error) {
	input = normalizeEventSearchInput(input)

	if input.OnlySubscriptionNews && userID == 0 {
		return emptyEventSearchResponse(input.Page, input.Limit), nil
	}

	page, err := s.store.SearchEvents(ctx, EventSearchQuery{
		ViewerID:             userID,
		Query:                input.Query,
		GroupID:              input.GroupID,
		CategoryIDs:          append([]uint(nil), input.CategoryIDs...),
		ExcludeCategoryIDs:   append([]uint(nil), input.ExcludeCategoryIDs...),
		GenreIDs:             append([]uint(nil), input.GenreIDs...),
		ExcludeGenreIDs:      append([]uint(nil), input.ExcludeGenreIDs...),
		EventTypeIDs:         append([]uint(nil), input.EventTypeIDs...),
		ExcludeEventTypeIDs:  append([]uint(nil), input.ExcludeEventTypeIDs...),
		LocationTypes:        append([]string(nil), input.LocationTypes...),
		ExcludeLocationTypes: append([]string(nil), input.ExcludeLocationTypes...),
		City:                 input.City,
		DateFrom:             input.DateFrom,
		DateTo:               input.DateTo,
		HasFreeSlots:         input.HasFreeSlots,
		OnlySubscriptionNews: input.OnlySubscriptionNews,
		Page:                 input.Page,
		Limit:                input.Limit,
	})
	if err != nil {
		s.logger.Error("Не удалось выполнить поиск событий", "userID", userID, "error", err)
		return nil, err
	}

	totalPages := calculateTotalPages(page.Total, input.Limit)
	return &dto.EventSearchResponse{
		Items:       eventSearchItemsToDTO(page.Items),
		Total:       page.Total,
		Limit:       input.Limit,
		CurrentPage: input.Page,
		TotalPages:  totalPages,
		HasMore:     input.Page < totalPages,
	}, nil
}

func (s *eventReadService) GetGroupEvents(ctx context.Context, actorID uint, groupID uint) ([]dto.EventShortDto, error) {
	view, err := s.store.GetGroupEvents(ctx, EventGroupEventsQuery{
		ViewerID: actorID,
		GroupID:  groupID,
	})
	if err != nil {
		s.logger.Error("Не удалось получить события группы", "groupID", groupID, "actorID", actorID, "error", err)
		return nil, err
	}

	if !view.GroupFound {
		return []dto.EventShortDto{}, nil
	}
	if view.GroupPrivate && !view.ViewerIsGroupMember {
		return nil, ErrNotGroupMember
	}

	return eventShortViewsToDTO(view.Items), nil
}

func (s *eventReadService) GetEventDetails(ctx context.Context, userID uint, eventID uint) (*dto.EventFullDto, error) {
	view, err := s.store.GetEventDetails(ctx, EventDetailsQuery{
		ViewerID: userID,
		EventID:  eventID,
	})
	if err != nil {
		s.logger.Error("Не удалось получить событие", "eventID", eventID, "userID", userID, "error", err)
		return nil, err
	}

	if !view.Found {
		return nil, ErrEventNotFound
	}
	if view.GroupPrivate && !view.ViewerIsGroupMember {
		return nil, ErrNotGroupMember
	}

	return eventFullViewToDTO(view.Event), nil
}

func eventSearchItemsToDTO(items []EventSearchItemView) []dto.EventSearchItemDto {
	result := make([]dto.EventSearchItemDto, 0, len(items))
	for _, item := range items {
		result = append(result, dto.EventSearchItemDto{
			ID:    item.ID,
			Title: item.Title,
			Group: dto.EventSearchGroupDto{
				ID:         item.Group.ID,
				Name:       item.Group.Name,
				Enterprise: item.Group.Enterprise,
			},
			Image:             item.ImageURL,
			ParticipantsCount: item.ParticipantsCount,
			MaxUsers:          item.MaxUsers,
			Duration:          item.Duration,
			StartDate:         item.StartTime.Format("2006-01-02"),
			EventType:         item.EventType,
			LocationType:      item.LocationType,
			City:              item.City,
			Genres:            append([]string(nil), item.Genres...),
			Subscribed:        item.ViewerSubscribed,
		})
	}
	return result
}

func eventShortViewsToDTO(items []EventShortView) []dto.EventShortDto {
	result := make([]dto.EventShortDto, 0, len(items))
	for _, item := range items {
		result = append(result, dto.EventShortDto{
			ID:           item.ID,
			Title:        item.Title,
			ImageURL:     item.ImageURL,
			MaxUsers:     item.MaxUsers,
			CurrentUsers: item.CurrentUsers,
			EventType:    item.EventTypeID,
			LocationType: item.LocationTypeID,
			AgeLimit:     item.AgeLimit,
			Genres:       append([]string(nil), item.Genres...),
			StartTime:    item.StartTime,
			Duration:     item.Duration,
			EventID:      item.ID,
			GroupID:      item.GroupID,
			Status:       item.Status,
			Subscribed:   item.ViewerSubscribed,
		})
	}
	return result
}

func eventFullViewToDTO(view EventFullView) *dto.EventFullDto {
	return &dto.EventFullDto{
		ID:           view.ID,
		Title:        view.Title,
		Description:  view.Description,
		ImageURL:     view.ImageURL,
		MaxUsers:     view.MaxUsers,
		CurrentUsers: view.CurrentUsers,
		StartTime:    view.StartTime,
		EndTime:      view.EndTime,
		Duration:     view.Duration,
		Group: dto.EventGroupDto{
			ID:         view.Group.ID,
			Name:       view.Group.Name,
			Image:      view.Group.Image,
			Enterprise: view.Group.Enterprise,
		},
		EventType:    view.EventTypeID,
		LocationType: view.LocationTypeID,
		Status:       view.StatusID,
		Genres:       append([]string(nil), view.Genres...),
		Creator: dto.EventCreatorDto{
			ID:       view.Creator.ID,
			Name:     view.Creator.Name,
			Us:       view.Creator.Us,
			Image:    view.Creator.Image,
			Verified: view.Creator.Verified,
		},
		Address:      view.Address,
		Country:      view.Country,
		AgeLimit:     view.AgeLimit,
		Year:         view.Year,
		Notes:        view.Notes,
		CustomFields: cloneReadCustomFields(view.CustomFields),
		Subscribed:   view.ViewerSubscribed,
		IsCreator:    view.ViewerIsCreator,
		CreatedAt:    view.CreatedAt,
		UpdatedAt:    view.UpdatedAt,
	}
}

func cloneReadCustomFields(source map[string]interface{}) map[string]interface{} {
	if source == nil {
		return nil
	}

	result := make(map[string]interface{}, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
