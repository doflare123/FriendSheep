package convertorsdto

import (
	"friendship/models/dto"
	"friendship/models/events"
)

// Конвертирует Event в административную DTO
func ConvertToAdminDto(event *events.Event, actorID uint) *dto.EventAdminDto {
	if event == nil {
		return nil
	}

	fullDto := ConvertToFullDto(event, actorID, false)

	// Формируем список всех участников
	allParticipants := make([]dto.EventAdminParticipantDto, 0, len(event.Users))
	for _, u := range event.Users {
		allParticipants = append(allParticipants, dto.EventAdminParticipantDto{
			ID:        u.ID,
			UserID:    u.UserID,
			Name:      u.User.Name,
			Username:  u.User.Us,
			Image:     u.User.Image,
			IsCreator: u.UserID == event.CreatorID,
			JoinedAt:  u.JoinedAt,
		})
	}

	return &dto.EventAdminDto{
		EventFullDto:    *fullDto,
		AllParticipants: allParticipants,
	}
}
