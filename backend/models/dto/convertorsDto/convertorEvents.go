package convertorsdto

import (
	"friendship/models/dto"
	"friendship/models/events"
)

func ConvertToShortDto(event *events.Event) *dto.EventShortDto {
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

	subscribed := false
	for _, u := range event.Users {
		if u.UserID == userID {
			subscribed = true
			break
		}
	}

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
		GroupID:      event.GroupID,

		EventType:    event.EventType.ID,
		LocationType: event.EventLocation.ID,
		Status:       event.Status.ID,
		Genres:       genres,

		Creator: dto.EventCreatorDto{
			ID:       event.Creator.ID,
			Name:     event.Creator.Name,
			Username: event.Creator.Us,
			Image:    event.Creator.Image,
		},

		Address:      event.Address,
		Country:      event.Country,
		AgeLimit:     event.AgeLimit.Name,
		Year:         event.Year,
		Notes:        event.Notes,
		CustomFields: event.CustomFields,

		Subscribed: subscribed,
		IsCreator:  isCreator,

		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}

	return fullDto
}

func ConvertManyToShortDto(events []events.Event) []dto.EventShortDto {
	result := make([]dto.EventShortDto, 0, len(events))
	for i := range events {
		if shortDto := ConvertToShortDto(&events[i]); shortDto != nil {
			result = append(result, *shortDto)
		}
	}
	return result
}
