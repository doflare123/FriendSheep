package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"

	"gorm.io/gorm"
)

type AdminGroupResponse struct {
	ID               *uint     `json:"id"`
	Name             *string   `json:"name"`
	Image            *string   `json:"image"`
	Type             *string   `json:"type"`
	SmallDescription *string   `json:"small_description"`
	Category         []*string `json:"category"`
	MemberCount      *int64    `json:"member_count"`
}

func GetAdminGroups(email *string) ([]AdminGroupResponse, error) {
	db := db.GetDB()
	user, err := FindUserByEmail(*email)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %v", err)
	}

	var adminGroups []AdminGroupResponse

	if err := db.Table("groups").
		Select("groups.id, groups.name, groups.image, groups.small_description, CASE WHEN groups.is_private THEN 'приватная группа' ELSE 'открытая группа' END as type, COUNT(group_users.user_id) as member_count").
		Joins("LEFT JOIN group_users ON group_users.group_id = groups.id").
		Where("groups.creater_id = ?", user.ID).
		Group("groups.id, groups.name, groups.image, groups.small_description, groups.is_private").
		Order("groups.id DESC").
		Scan(&adminGroups).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении групп: %v", err)
	}

	for i := range adminGroups {
		if adminGroups[i].ID != nil {
			var categories []string
			if err := db.Table("categories").
				Select("categories.name").
				Joins("JOIN group_group_categories ON group_group_categories.group_category_id = categories.id").
				Where("group_group_categories.group_id = ?", *adminGroups[i].ID).
				Pluck("name", &categories).Error; err != nil {
				return nil, fmt.Errorf("ошибка при получении категорий для группы %d: %v", *adminGroups[i].ID, err)
			}

			adminGroups[i].Category = make([]*string, len(categories))
			for j, cat := range categories {
				adminGroups[i].Category[j] = &cat
			}
		}
	}

	return adminGroups, nil
}

type AdminGroupInfResponse struct {
	ID               uint                  `json:"id"`
	Name             string                `json:"name"`
	SmallDescription string                `json:"small_description"`
	Description      string                `json:"description"`
	Contacts         []Contacts            `json:"contacts"`
	Image            string                `json:"image"`
	City             string                `json:"city"`
	IsPrivete        bool                  `json:"private"`
	Categories       []string              `json:"categories"`
	Sessions         []SessionResponse     `json:"sessions"`
	Applications     []GroupJoinRequestRes `json:"applications"`
}

type GroupJoinRequestRes struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"userId"`
	Name   string `json:"name"`
	Us     string `json:"us"`
	Image  string `json:"image"`
}

func GetAdminGroupInfo(email string, groupID *uint) (*AdminGroupInfResponse, error) {
	if groupID == nil {
		return nil, errors.New("invalid group ID")
	}

	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("пользователь не найден: %v", err)
	}

	var group groups.Group
	err := db.GetDB().Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		Where("id = ?", groupID).
		First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	if group.CreaterID == 0 || user.ID == 0 {
		return nil, errors.New("access denied: invalid user or group data")
	}
	if group.CreaterID != user.ID {
		return nil, fmt.Errorf("access denied: user %d is not the group creator %d", &user.ID, &group.CreaterID)
	}

	type applicationResult struct {
		applications []GroupJoinRequestRes
		err          error
	}
	appChan := make(chan applicationResult, 1)

	go func() {
		applications, err := GetPendingJoinRequestsForAdmin(email, *groupID)
		appChan <- applicationResult{applications: applications, err: err}
	}()

	var Gsessions []sessions.Session
	var recruitmentStatusID *uint
	var recruitmentStatus sessions.Status
	err = db.GetDB().Where("status = ?", "Набор").First(&recruitmentStatus).Error
	if err == nil && recruitmentStatus.ID != 0 {
		recruitmentStatusID = &recruitmentStatus.ID
		err = db.GetDB().Preload("SessionType").
			Preload("SessionPlace").
			Preload("Group").
			Where("group_id = ? AND status_id = ?", groupID, recruitmentStatusID).
			Find(&Gsessions).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get sessions: %w", err)
		}
	}

	var sessionIDs []uint
	for _, session := range Gsessions {
		if session.ID != 0 {
			sessionIDs = append(sessionIDs, session.ID)
		}
	}

	var metadataMap map[uint]*sessions.SessionMetadata
	if len(sessionIDs) > 0 {
		metadataMap, err = db.GetSessionsMetadata(sessionIDs)
		if err != nil {
			fmt.Printf("Warning: failed to get session metadata: %v\n", err)
			metadataMap = make(map[uint]*sessions.SessionMetadata)
		}
	} else {
		metadataMap = make(map[uint]*sessions.SessionMetadata)
	}

	appResult := <-appChan
	if appResult.err != nil {
		return nil, fmt.Errorf("failed to get pending join requests: %w", appResult.err)
	}

	response := &AdminGroupInfResponse{
		ID:               group.ID,
		Name:             group.Name,
		SmallDescription: group.SmallDescription,
		Description:      group.Description,
		Image:            group.Image,
		City:             group.City,
		IsPrivete:        group.IsPrivate,
		Applications:     appResult.applications,
	}

	var contacts []Contacts
	for _, contact := range group.Contacts {
		if contact.Name != "" && contact.Link != "" {
			contacts = append(contacts, Contacts{
				Name: &contact.Name,
				Link: &contact.Link,
			})
		}
	}
	response.Contacts = contacts

	var categories []string
	for _, category := range group.Categories {
		if category.Name != "" {
			categories = append(categories, category.Name)
		}
	}
	response.Categories = categories

	var sessionResponses []SessionResponse
	for _, session := range Gsessions {
		sessionResponse := SessionResponse{
			ID:            session.ID,
			Title:         session.Title,
			StartTime:     session.StartTime,
			ImageURL:      session.ImageURL,
			Duration:      session.Duration,
			CurrentUsers:  session.CurrentUsers,
			CountUsersMax: session.CountUsersMax,
			GroupID:       session.GroupID,
		}
		if session.SessionType.Name != "" {
			sessionResponse.SessionType = session.SessionType.Name
		}
		if session.SessionPlace.Title != "" {
			sessionResponse.SessionPlace = session.SessionPlace.Title
		}
		if session.Group.Name != "" {
			sessionResponse.GroupName = session.Group.Name
		}
		if session.ID != 0 {
			if metadata, exists := metadataMap[session.ID]; exists && metadata != nil {
				sessionResponse.Genres = metadata.Genres
			}
		}
		sessionResponses = append(sessionResponses, sessionResponse)
	}
	response.Sessions = sessionResponses

	return response, nil
}
