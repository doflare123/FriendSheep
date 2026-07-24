package dto

type ReferencesDto struct {
	EventTypes       []ReferenceItemDto       `json:"eventTypes"`
	Locations        []ReferenceItemDto       `json:"locations"`
	AgeLimits        []ReferenceItemDto       `json:"ageLimits"`
	Statuses         []ReferenceItemDto       `json:"statuses"`
	GroupCategories  []ReferenceItemDto       `json:"groupCategories"`
	GroupActionTypes []ActionReferenceItemDto `json:"groupActionTypes"`
}

type ReferenceItemDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ActionReferenceItemDto struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type GenreSearchResponseDto struct {
	Items   []ReferenceItemDto `json:"items"`
	Total   int64              `json:"total"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	HasMore bool               `json:"hasMore"`
}
