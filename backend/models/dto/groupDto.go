package dto

import "time"

type GroupDto struct {
	ID               uint         `json:"id"`
	Name             string       `json:"nameGroup"`
	Description      string       `json:"description"`
	SmallDescription string       `json:"smallDescription"`
	Image            string       `json:"image"`
	CreatorID        uint         `json:"creatorId"`
	CreatorName      string       `json:"creatorName"`
	CreatorUsername  string       `json:"creatorUsername"`
	IsPrivate        bool         `json:"isPrivate"`
	City             string       `json:"city,omitempty"`
	Categories       []string     `json:"categories"`
	Contacts         []ContactDto `json:"contacts"`
	MemberCount      *int         `json:"memberCount,omitempty"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

type ContactDto struct {
	Name string `json:"name"`
	Link string `json:"link"`
}
