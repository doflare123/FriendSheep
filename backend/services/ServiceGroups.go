package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"
	"log"
	"strings"

	"gorm.io/gorm"
)

type CreateGroupInput struct {
	Name             string  `form:"name" binding:"required,min=5,max=40" example:"Моя группа"`
	Description      string  `form:"description" binding:"required,min=5,max=300" example:"Полное описание группы"`
	SmallDescription string  `form:"smallDescription" binding:"required,min=5,max=50" example:"Короткое описание"`
	Image            string  `form:"-"`
	IsPrivate        *bool   `form:"isPrivate" binding:"required" example:"true/1/0/false"`
	City             string  `form:"city,omitempty" example:"Москва"`
	Categories       []*uint `form:"categories" binding:"required" example:"[1, 2, 3]"`
	Contacts         string  `form:"contacts,omitempty" example:"vk:https://vk.com/mygroup, tg:https://t.me/mygroup, inst:https://instagram.com/mygroup"`
}

type JoinGroupResult struct {
	Message string `json:"message"`
	Joined  bool   `json:"joined"`
}

// func (input *CreateGroupInput) IsPrivateBool() (bool, error) {
// 	switch input.IsPrivate {
// 	case "1", "true", "True", "TRUE":
// 		return true, nil
// 	case "0", "false", "False", "FALSE":
// 		return false, nil
// 	default:
// 		return false, fmt.Errorf("некорректное значение isPrivate: %s", input.IsPrivate)
// 	}
// }

func parseContacts(contactsStr string) map[string]string {
	contacts := make(map[string]string)
	if contactsStr == "" {
		return contacts
	}

	// Разделяем по запятым
	pairs := strings.Split(contactsStr, ",")
	for _, pair := range pairs {
		// Убираем пробелы
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Разделяем по двоеточию
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			link := strings.TrimSpace(parts[1])
			if name != "" && link != "" {
				contacts[name] = link
			}
		}
	}

	return contacts
}

type JoinGroupInput struct {
	GroupID uint `json:"groupId" binding:"required"`
}

func CreateGroup(email *string, input CreateGroupInput) (group *groups.Group, err error) {
	if err := ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	// Парсим контакты из строки
	var contacts map[string]string
	if input.Contacts != "" {
		contacts = parseContacts(input.Contacts)
	}

	var creator models.User
	if err := db.GetDB().Where("email = ?", email).First(&creator).Error; err != nil {
		return nil, fmt.Errorf("создатель не найден по email (%s): %v", *email, err)
	}

	// Загрузка категорий
	var categories []models.Category
	if len(input.Categories) > 0 {
		if err := db.GetDB().Where("id IN ?", input.Categories).Find(&categories).Error; err != nil {
			return nil, fmt.Errorf("ошибка загрузки категорий: %v", err)
		}
	}

	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("паника в CreateGroup: %v", r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	newGroup := &groups.Group{
		Name:             input.Name,
		Description:      input.Description,
		SmallDescription: input.SmallDescription,
		Image:            input.Image,
		CreaterID:        creator.ID,
		IsPrivate:        *input.IsPrivate,
		City:             input.City,
		Categories:       categories,
	}

	if err = tx.Create(newGroup).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания группы: %v", err)
	}

	groupUser := groups.GroupUsers{
		UserID:      creator.ID,
		GroupID:     newGroup.ID,
		RoleInGroup: "admin",
	}
	if err = tx.Create(&groupUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
	}

	if len(contacts) > 0 {
		groupContacts := make([]groups.GroupContact, 0, len(contacts))
		for name, link := range contacts {
			if name != "" && link != "" {
				groupContacts = append(groupContacts, groups.GroupContact{
					GroupID: newGroup.ID,
					Name:    name,
					Link:    link,
				})
			}
		}

		if len(groupContacts) > 0 {
			if err = tx.Create(&groupContacts).Error; err != nil {
				return nil, fmt.Errorf("ошибка сохранения контактов группы: %v", err)
			}
		}
	}

	if err = tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("не удалось закоммитить транзакцию: %v", err)
	}

	if errReload := db.GetDB().Preload("Categories").Preload("Contacts").Preload("Creater").First(newGroup, newGroup.ID).Error; errReload != nil {
		log.Printf("Не удалось перезагрузить группу с ассоциациями: %v", errReload)
	}

	return newGroup, nil
}

func JoinGroup(email string, input JoinGroupInput) (*JoinGroupResult, error) {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	var existing groups.GroupUsers
	if err := db.GetDB().
		Where("user_id = ? AND group_id = ?", user.ID, input.GroupID).
		First(&existing).Error; err == nil {
		return nil, fmt.Errorf("пользователь уже в группе")
	}

	var existingRequest groups.GroupJoinRequest
	if err := db.GetDB().
		Where("user_id = ? AND group_id = ?", user.ID, input.GroupID).
		First(&existingRequest).Error; err == nil {
		if existingRequest.Status == "pending" {
			return nil, fmt.Errorf("заявка уже отправлена и ожидает подтверждения")
		}
	}

	var group groups.Group
	if err := db.GetDB().Where("id = ?", input.GroupID).First(&group).Error; err != nil {
		return nil, fmt.Errorf("группа не найдена")
	}

	if group.IsPrivate {
		request := groups.GroupJoinRequest{
			UserID:  user.ID,
			GroupID: group.ID,
			Status:  "pending",
		}
		if err := db.GetDB().Create(&request).Error; err != nil {
			return nil, fmt.Errorf("ошибка создания заявки: %v", err)
		}
		return &JoinGroupResult{
			Message: "Заявка на вступление отправлена, ожидайте подтверждения от администратора группы",
			Joined:  false,
		}, nil
	} else {
		member := groups.GroupUsers{
			UserID:      user.ID,
			GroupID:     group.ID,
			RoleInGroup: "member",
		}
		if err := db.GetDB().Create(&member).Error; err != nil {
			return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
		}
		return &JoinGroupResult{
			Message: "Пользователь успешно присоединился к группе",
			Joined:  true,
		}, nil
	}
}

func DeleteGroup(groupID uint) error {
	group := groups.Group{ID: groupID}
	result := db.GetDB().Select("Categories").Delete(&group)

	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении группы: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("группа не найдена или уже удалена")
	}

	return nil
}

type GroupInf struct {
	ID           uint                    `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Image        string                  `json:"image"`
	City         string                  `json:"city"`
	Creater      string                  `json:"creater"`
	CountMembers int64                   `json:"count_members"`
	Subscription bool                    `json:"subscription"`
	Users        []UsersGroups           `json:"users"`
	Categories   []*string               `json:"categories"`
	Contacts     []*Contacts             `json:"contacts"`
	Sessions     []SessionDetailResponse `json:"sessions"`
}

type UsersGroups struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Contacts struct {
	Name *string `json:"name"`
	Link *string `json:"link"`
}

func GetGroupInf(groupID *uint64, email *string) (*GroupInf, error) {
	if *groupID == 0 {
		return nil, errors.New("некорректный ID группы")
	}
	user, err := FindUserByEmail(*email)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %v", err)
	}

	var group groups.Group
	var information GroupInf

	err = db.GetDB().Preload("Categories").
		Preload("Contacts").
		Where("id = ?", groupID).
		First(&group).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("группа не найдена")
		}
		return nil, err
	}

	if group.IsPrivate {
		isMember, err := checkGroupMembership(groupID, &user.ID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.New("доступ к приватной группе запрещен")
		}
	}

	information.ID = group.ID
	information.Name = group.Name
	information.Description = group.Description
	information.Image = group.Image
	information.City = group.City
	information.Creater = group.Creater.Name

	users, err := getGroupUsers(*groupID)
	if err != nil {
		return nil, err
	}
	information.Users = users

	var subscriptionCount int64
	err = db.GetDB().Model(&groups.GroupUsers{}).
		Where("group_id = ? AND user_id = ?", groupID, user.ID).
		Count(&subscriptionCount).Error
	if err != nil {
		return nil, err
	}
	information.Subscription = subscriptionCount > 0

	var totalMembers int64
	err = db.GetDB().Model(&groups.GroupUsers{}).
		Where("group_id = ?", groupID).
		Count(&totalMembers).Error
	if err != nil {
		return nil, err
	}
	information.CountMembers = totalMembers

	categories := make([]*string, len(group.Categories))
	for i, category := range group.Categories {
		if category.Name != "" {
			categoryName := category.Name
			categories[i] = &categoryName
		}
	}
	information.Categories = categories

	contacts := make([]*Contacts, 0, len(group.Contacts))
	for _, contact := range group.Contacts {
		if contact.Name != "" && contact.Link != "" {
			contactName := &contact.Name
			contactLink := &contact.Link

			contacts = append(contacts, &Contacts{
				Name: contactName,
				Link: contactLink,
			})
		}
	}
	information.Contacts = contacts

	sessions, err := getGroupSessions(*groupID)
	if err != nil {
		return nil, err
	}
	information.Sessions = sessions

	return &information, nil
}

//Admin часть групп

type GroupUpdateInput struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	SmallDescription *string `json:"small_description"`
	Image            *string `json:"image"`
	IsPrivate        *bool   `json:"is_private"`
	City             *string `json:"city"`
	Categories       []*uint `json:"categories"`
	Contacts         *string `json:"contacts"`
}

func UpdateGroup(groupID uint, input GroupUpdateInput) (err error) {
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("паника в UpdateGroup: %v", r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	var group groups.Group
	if err = tx.First(&group, groupID).Error; err != nil {
		return fmt.Errorf("группа не найдена")
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.SmallDescription != nil {
		updates["small_description"] = *input.SmallDescription
	}
	if input.Image != nil {
		updates["image"] = *input.Image
	}
	if input.IsPrivate != nil {
		updates["is_private"] = *input.IsPrivate
	}
	if input.City != nil {
		updates["city"] = *input.City
	}

	if len(updates) > 0 {
		if err = tx.Model(&group).Updates(updates).Error; err != nil {
			return fmt.Errorf("не удалось сохранить изменения группы: %v", err)
		}
	}

	if input.Categories != nil {
		if err = tx.Model(&group).Association("Categories").Clear(); err != nil {
			return fmt.Errorf("не удалось очистить старые категории: %v", err)
		}

		if len(input.Categories) > 0 {
			var newCategories []models.Category
			if err = tx.Where("id IN ?", input.Categories).Find(&newCategories).Error; err != nil {
				return fmt.Errorf("не удалось найти переданные категории: %v", err)
			}
			if err = tx.Model(&group).Association("Categories").Replace(&newCategories); err != nil {
				return fmt.Errorf("не удалось назначить новые категории: %v", err)
			}
		}
	}

	if input.Contacts != nil {
		newContacts := parseContacts(*input.Contacts)
		if err = updateContactsInTx(tx, &groupID, newContacts); err != nil {
			return fmt.Errorf("ошибка обновления контактов: %v", err)
		}
	}

	return tx.Commit().Error
}

func updateContactsInTx(tx *gorm.DB, groupID *uint, newContacts map[string]string) error {
	var existingContacts []groups.GroupContact
	if err := tx.Where("group_id = ?", groupID).Find(&existingContacts).Error; err != nil {
		return fmt.Errorf("ошибка получения существующих контактов: %v", err)
	}

	existingMap := make(map[string]groups.GroupContact)
	for _, c := range existingContacts {
		if c.Name != "" {
			existingMap[c.Name] = c
		}
	}

	// Обрабатываем новые контакты
	for name, link := range newContacts {
		if existingContact, ok := existingMap[name]; ok {
			if existingContact.Link != "" && existingContact.Link != link {
				if err := tx.Model(&existingContact).Update("link", link).Error; err != nil {
					return fmt.Errorf("ошибка обновления контакта '%s': %v", name, err)
				}
			}
			delete(existingMap, name)
		} else {
			newContact := groups.GroupContact{
				GroupID: *groupID,
				Name:    name,
				Link:    link,
			}
			if err := tx.Create(&newContact).Error; err != nil {
				return fmt.Errorf("ошибка добавления контакта '%s': %v", name, err)
			}
		}
	}

	for _, contactToDelete := range existingMap {
		if err := tx.Delete(&contactToDelete).Error; err != nil {
			contactName := "unknown"
			if contactToDelete.Name != "" {
				contactName = contactToDelete.Name
			}
			return fmt.Errorf("ошибка удаления старого контакта '%s': %v", contactName, err)
		}
	}

	return nil
}

// вспомогательные для получения данных о группе
func checkGroupMembership(groupID *uint64, userID *uint) (bool, error) {
	var count int64
	err := db.GetDB().Model(&groups.GroupUsers{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func getGroupUsers(groupID uint64) ([]UsersGroups, error) {
	var groupUsers []groups.GroupUsers

	err := db.GetDB().
		Preload("User").
		Where("group_id = ?", groupID).
		Limit(6).
		Find(&groupUsers).Error

	if err != nil {
		return nil, err
	}

	users := make([]UsersGroups, 0, len(groupUsers))
	for _, groupUser := range groupUsers {
		if groupUser.User.Name != "" {
			userName := groupUser.User.Name
			var userImage *string
			if groupUser.User.Image != "" {
				userImageCopy := groupUser.User.Image
				userImage = &userImageCopy
			}

			users = append(users, UsersGroups{
				Name:  userName,
				Image: *userImage,
			})
		}
	}

	return users, nil
}

func getGroupSessions(groupID uint64) ([]SessionDetailResponse, error) {
	var Gsessions []sessions.Session

	err := db.GetDB().
		Preload("SessionType").
		Preload("SessionPlace").
		Where("group_id = ?", groupID).
		Order("start_time ASC").
		Find(&Gsessions).Error

	if err != nil {
		return nil, err
	}

	if len(Gsessions) == 0 {
		return []SessionDetailResponse{}, nil
	}

	sessionIDs := make([]uint, len(Gsessions))
	for i, session := range Gsessions {
		sessionIDs[i] = session.ID
	}

	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
	if err != nil {
		log.Printf("Ошибка получения метаданных: %v", err)
		metadataMap = make(map[uint]*sessions.SessionMetadata)
	}

	sessionResponses := make([]SessionDetailResponse, 0, len(Gsessions))
	for _, session := range Gsessions {
		if session.ID == 0 {
			continue
		}

		var sessionTypeName *string
		if session.SessionType.Name != "" {
			typeName := session.SessionType.Name
			sessionTypeName = &typeName
		}

		var sessionPlaceName *string
		if session.SessionPlace.Title != "" {
			placeName := session.SessionPlace.Title
			sessionPlaceName = &placeName
		}

		var title, imageURL *string
		if session.Title != "" {
			titleCopy := session.Title
			title = &titleCopy
		}
		if session.ImageURL != "" {
			imageURLCopy := session.ImageURL
			imageURL = &imageURLCopy
		}

		var duration, currentUsers, countUsersMax *uint16
		if session.Duration != 0 {
			durationCopy := session.Duration
			duration = &durationCopy
		}
		if session.CurrentUsers != 0 {
			currentUsersCopy := session.CurrentUsers
			currentUsers = &currentUsersCopy
		}
		if session.CountUsersMax != 0 {
			countUsersMaxCopy := session.CountUsersMax
			countUsersMax = &countUsersMaxCopy
		}

		var groupIDCopy *uint
		if session.GroupID != 0 {
			groupID := session.GroupID
			groupIDCopy = &groupID
		}

		var sessionIDCopy *uint
		if session.ID != 0 {
			sessionID := session.ID
			sessionIDCopy = &sessionID
		}

		subSession := SubSessionDetail{
			ID:            *sessionIDCopy,
			Title:         *title,
			SessionType:   *sessionTypeName,
			SessionPlace:  *sessionPlaceName,
			GroupID:       *groupIDCopy,
			StartTime:     session.StartTime,
			EndTime:       session.EndTime,
			Duration:      *duration,
			CurrantUsers:  *currentUsers,
			CountUsersMax: *countUsersMax,
			ImageURL:      *imageURL,
		}

		// Получаем метаданные для текущей сессии
		var metadata *sessions.SessionMetadata
		if metaData, exists := metadataMap[session.ID]; exists {
			metadata = metaData
		}

		sessionResponse := SessionDetailResponse{
			Session:  subSession,
			Metadata: metadata,
		}

		sessionResponses = append(sessionResponses, sessionResponse)
	}

	return sessionResponses, nil
}
