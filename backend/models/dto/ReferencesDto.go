package dto

type ReferencesDto struct {
	EventTypes       []ReferenceItemDto       `json:"eventTypes"`
	Locations        []ReferenceItemDto       `json:"locations"`
	AgeLimits        []ReferenceItemDto       `json:"ageLimits"`
	Statuses         []ReferenceItemDto       `json:"statuses"`
	Genres           []ReferenceItemDto       `json:"genres"`
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
