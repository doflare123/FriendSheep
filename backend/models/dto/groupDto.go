package dto

import "time"

// GroupFullDto - полная информация о группе
type GroupFullDto struct {
	ID               uint         `json:"id"`
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	SmallDescription string       `json:"smallDescription"`
	Image            string       `json:"image"`
	IsPrivate        bool         `json:"isPrivate"`
	Enterprise       bool         `json:"enterprise"`
	City             string       `json:"city,omitempty"`
	Categories       []string     `json:"categories"`
	Contacts         []ContactDto `json:"contacts"`
	MemberCount      int          `json:"memberCount"`

	// Информация о создателе
	Creator GroupCreatorDto `json:"creator"`

	// Первые 10 участников
	Members []GroupMemberDto `json:"members"`

	// События в стадии набора
	ActiveEvents []EventShortDto `json:"activeEvents"`

	// Флаги для пользователя
	IsSubscribed bool   `json:"isSubscribed"`
	UserRole     string `json:"userRole"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GroupCreatorDto - информация о создателе группы
type GroupCreatorDto struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Verified bool   `json:"verified"`
}

// GroupMemberDto - информация об участнике группы
type GroupMemberDto struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Role     string `json:"role"`
}

type ContactDto struct {
	Name string `json:"name"`
	Link string `json:"link"`
}
