package dto

import "time"

type GroupSearchItemDto struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	Categories       []string  `json:"categories"`
	MemberCount      int       `json:"memberCount"`
	IsSubscribed     bool      `json:"isSubscribed"`
	Image            string    `json:"image"`
	CreatedAt        time.Time `json:"createdAt"`
	SmallDescription string    `json:"smallDescription"`
	IsPrivate        bool      `json:"isPrivate"`
	Enterprise       bool      `json:"enterprise"`
}

type GroupSearchResponseDto struct {
	Items      []GroupSearchItemDto `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"totalPages"`
	HasMore    bool                 `json:"hasMore"`
}
