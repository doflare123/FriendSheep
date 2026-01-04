package dto

type ReferencesDto struct {
	EventTypes      []ReferenceItemDto `json:"eventTypes"`
	Locations       []ReferenceItemDto `json:"locations"`
	AgeLimits       []ReferenceItemDto `json:"ageLimits"`
	Statuses        []ReferenceItemDto `json:"statuses"`
	GroupCategories []ReferenceItemDto `json:"groupCategories"`
}

type ReferenceItemDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
