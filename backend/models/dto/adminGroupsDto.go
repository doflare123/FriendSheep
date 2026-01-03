package dto

import "time"

type AdminGroupResponse struct {
	ID               *uint     `json:"id"`
	Name             *string   `json:"name"`
	Image            *string   `json:"image"`
	Type             *string   `json:"type"`
	SmallDescription *string   `json:"small_description"`
	Category         []*string `json:"category"`
	MemberCount      *int64    `json:"member_count"`
	Role             string    `json:"role_in_group"`
}

type EventAdminDto struct {
	EventFullDto
	AllParticipants []EventAdminParticipantDto `json:"allParticipants"`
}

type EventAdminParticipantDto struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Image     string    `json:"image"`
	IsCreator bool      `json:"isCreator"`
	JoinedAt  time.Time `json:"joinedAt"`
}
