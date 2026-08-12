package convertorsdto

import (
	"friendship/models/dto"
	"friendship/models/events"
)

func ConvertToShortDto(event *events.Event) *dto.EventShortDto {
	return ConvertToShortDtoForUser(event, 0)
}

func ConvertToShortDtoForUser(event *events.Event, userID uint) *dto.EventShortDto {
	if event == nil {
		return nil
	}

	genres := make([]string, 0, len(event.Genres))
	for _, g := range event.Genres {
		genres = append(genres, g.Genre.Name)
	}

	return &dto.EventShortDto{
		ID:           event.ID,
		Title:        event.Title,
		ImageURL:     event.ImageURL,
		MaxUsers:     event.MaxUsers,
		CurrentUsers: event.CurrentUsers,
		EventType:    event.EventType.ID,
		LocationType: event.EventLocation.ID,
		AgeLimit:     event.AgeLimit.Name,
		Genres:       genres,
		StartTime:    event.StartTime,
		Duration:     event.Duration,
		EventID:      event.ID,
		GroupID:      event.GroupID,
		Status:       event.Status.Name,
		Subscribed:   eventSubscribedByUser(event, userID),
	}
}

func ConvertToFullDto(event *events.Event, userID uint, includeParticipants bool) *dto.EventFullDto {
	if event == nil {
		return nil
	}

	genres := make([]string, 0, len(event.Genres))
	for _, g := range event.Genres {
		genres = append(genres, g.Genre.Name)
	}

	isCreator := event.CreatorID == userID

	fullDto := &dto.EventFullDto{
		ID:           event.ID,
		Title:        event.Title,
		Description:  event.Description,
		ImageURL:     event.ImageURL,
		MaxUsers:     event.MaxUsers,
		CurrentUsers: event.CurrentUsers,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Duration:     event.Duration,
		Group: dto.EventGroupDto{
			ID:         event.Group.ID,
			Name:       event.Group.Name,
			Image:      event.Group.Image,
			Enterprise: event.Group.Enterprise,
		},

		EventType:    event.EventType.ID,
		LocationType: event.EventLocation.ID,
		Status:       event.Status.ID,
		Genres:       genres,

		Creator: dto.EventCreatorDto{
			ID:       event.Creator.ID,
			Name:     event.Creator.Name,
			Us:       event.Creator.Us,
			Image:    event.Creator.Image,
			Verified: event.Creator.VerifiedUser,
		},

		Address:      event.Address,
		Country:      event.Country,
		AgeLimit:     event.AgeLimit.Name,
		Year:         event.Year,
		Notes:        event.Notes,
		CustomFields: event.CustomFields,

		Subscribed: eventSubscribedByUser(event, userID),
		IsCreator:  isCreator,

		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}

	return fullDto
}

func ConvertManyToShortDto(events []events.Event) []dto.EventShortDto {
	return ConvertManyToShortDtoForUser(events, 0)
}

func ConvertManyToShortDtoForUser(events []events.Event, userID uint) []dto.EventShortDto {
	result := make([]dto.EventShortDto, 0, len(events))
	for i := range events {
		if shortDto := ConvertToShortDtoForUser(&events[i], userID); shortDto != nil {
			result = append(result, *shortDto)
		}
	}
	return result
}

func ConvertToSearchItemDto(event *events.Event) *dto.EventSearchItemDto {
	return ConvertToSearchItemDtoForUser(event, 0)
}

func ConvertToSearchItemDtoForUser(event *events.Event, userID uint) *dto.EventSearchItemDto {
	if event == nil {
		return nil
	}

	genres := make([]string, 0, len(event.Genres))
	for _, g := range event.Genres {
		genres = append(genres, g.Genre.Name)
	}

	return &dto.EventSearchItemDto{
		ID:    event.ID,
		Title: event.Title,
		Group: dto.EventSearchGroupDto{
			ID:         event.Group.ID,
			Name:       event.Group.Name,
			Enterprise: event.Group.Enterprise,
		},
		Image:        event.ImageURL,
		CurrentUsers: event.CurrentUsers,
		MaxUsers:     event.MaxUsers,
		Duration:     event.Duration,
		StartDate:    event.StartTime.Format("2006-01-02"),
		EventType:    event.EventType.Name,
		LocationType: event.EventLocation.Name,
		City:         event.Group.City,
		Genres:       genres,
		Subscribed:   eventSubscribedByUser(event, userID),
	}
}

func ConvertManyToSearchItemDto(events []events.Event) []dto.EventSearchItemDto {
	return ConvertManyToSearchItemDtoForUser(events, 0)
}

func ConvertManyToSearchItemDtoForUser(events []events.Event, userID uint) []dto.EventSearchItemDto {
	result := make([]dto.EventSearchItemDto, 0, len(events))
	for i := range events {
		if searchDto := ConvertToSearchItemDtoForUser(&events[i], userID); searchDto != nil {
			result = append(result, *searchDto)
		}
	}
	return result
}

func eventSubscribedByUser(event *events.Event, userID uint) bool {
	if event == nil || userID == 0 {
		return false
	}

	for _, u := range event.Users {
		if u.UserID == userID {
			return true
		}
	}

	return false
}
