package dto

type AdminGroupResponse struct {
	ID               *uint     `json:"id"`
	Name             *string   `json:"name"`
	Image            *string   `json:"image"`
	Type             *string   `json:"type"`
	SmallDescription *string   `json:"small_description"`
	Category         []*string `json:"category"`
	MemberCount      *int64    `json:"member_count"`
}
