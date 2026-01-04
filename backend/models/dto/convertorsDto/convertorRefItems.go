package convertorsdto

import (
	"friendship/models"
	"friendship/models/dto"
	"friendship/models/events"
)

func ConvertToReferenceItems(categories []models.Category) []dto.ReferenceItemDto {
	result := make([]dto.ReferenceItemDto, 0, len(categories))
	for _, cat := range categories {
		result = append(result, dto.ReferenceItemDto{
			ID:   cat.ID,
			Name: cat.Name,
		})
	}
	return result
}

func ConvertLocationsToReferenceItems(locations []events.EventLocation) []dto.ReferenceItemDto {
	result := make([]dto.ReferenceItemDto, 0, len(locations))
	for _, loc := range locations {
		result = append(result, dto.ReferenceItemDto{
			ID:   loc.ID,
			Name: loc.Name,
		})
	}
	return result
}

func ConvertAgeLimitsToReferenceItems(ageLimits []events.AgeLimit) []dto.ReferenceItemDto {
	result := make([]dto.ReferenceItemDto, 0, len(ageLimits))
	for _, limit := range ageLimits {
		result = append(result, dto.ReferenceItemDto{
			ID:   limit.ID,
			Name: limit.Name,
		})
	}
	return result
}

func ConvertStatusesToReferenceItems(statuses []events.Status) []dto.ReferenceItemDto {
	result := make([]dto.ReferenceItemDto, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, dto.ReferenceItemDto{
			ID:   status.ID,
			Name: status.Name,
		})
	}
	return result
}

func ConvertGenresToReferenceItems(genres []events.Genre) []dto.ReferenceItemDto {
	result := make([]dto.ReferenceItemDto, 0, len(genres))
	for _, genre := range genres {
		result = append(result, dto.ReferenceItemDto{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}
	return result
}
