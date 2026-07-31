package dto

type ManagedGroupItemDto struct {
	ID               uint     `json:"id"`
	Name             string   `json:"name"`
	Categories       []string `json:"categories"`
	SmallDescription string   `json:"smallDescription"`
	MemberCount      int      `json:"memberCount"`
	Image            string   `json:"image"`
}

type ManagedGroupsDto struct {
	Admin     []ManagedGroupItemDto `json:"admin"`
	Moderator []ManagedGroupItemDto `json:"moderator"`
}
