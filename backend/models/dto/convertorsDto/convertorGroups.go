package convertorsdto

import (
	"friendship/models/dto"
	"friendship/models/groups"
)

func ConvertToGroupFullDto(group groups.Group, totalMembers int64, members []dto.GroupMemberDto, activeEvents []dto.EventShortDto, isSubscribed bool, userRole string) *dto.GroupFullDto {
	categories := make([]string, 0, len(group.Categories))
	for _, cat := range group.Categories {
		categories = append(categories, cat.Name)
	}

	contacts := make([]dto.ContactDto, 0, len(group.Contacts))
	for _, contact := range group.Contacts {
		contacts = append(contacts, dto.ContactDto{
			Name: contact.Name,
			Link: contact.Link,
		})
	}

	return &dto.GroupFullDto{
		ID:               group.ID,
		Name:             group.Name,
		Description:      group.Description,
		SmallDescription: group.SmallDescription,
		Image:            group.Image,
		IsPrivate:        group.IsPrivate,
		Enterprise:       group.Enterprise,
		City:             group.City,
		Categories:       categories,
		Contacts:         contacts,
		MemberCount:      int(totalMembers),
		Creator: dto.GroupCreatorDto{
			ID:       group.Creater.ID,
			Name:     group.Creater.Name,
			Username: group.Creater.Us,
			Image:    group.Creater.Image,
			Verified: group.Creater.VerifiedUser,
		},
		Members:      members,
		ActiveEvents: activeEvents,
		IsSubscribed: isSubscribed,
		UserRole:     userRole,
		CreatedAt:    group.CreatedAt,
		UpdatedAt:    group.UpdatedAt,
	}
}
