package group

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	"friendship/models/groups"
	"friendship/repository"
	"friendship/services"
	"log"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrCategoriesNotFound = errors.New("категории не найдены")
	ErrInvalidInput       = errors.New("невалидная структура данных")
	ErrGroupCreation      = errors.New("ошибка создания группы")
)

type CreateGroupInput struct {
	Name             string  `json:"name" form:"name" binding:"required,min=5,max=40" example:"Любители настольных игр"`
	Description      string  `json:"description" form:"description" binding:"required,min=5,max=300" example:"Группа для любителей игр"`
	SmallDescription string  `json:"smallDescription" form:"smallDescription" binding:"required,min=5,max=50" example:"Играем вместе!"`
	Image            string  `json:"image" form:"image" binding:"required,url" example:"https://cdn.example.com/images/board-games.jpg"`
	IsPrivate        *bool   `json:"isPrivate" form:"isPrivate" binding:"required" example:"false"`
	City             string  `json:"city,omitempty" form:"city" example:"Москва"`
	Categories       []*uint `json:"categories" form:"categories" binding:"required,min=1" example:"[1,3,5]"`
	Contacts         string  `json:"contacts,omitempty" form:"contacts" example:"vk:https://vk.com/mygroup, tg:https://t.me/mygroup"`
}

type GroupsService interface {
	CreateGroup(id uint, inf CreateGroupInput) (*dto.GroupDto, error)
	UpdateGroup(inf GroupUpdateInput) (dto.GroupDto, error)
	DeleteGroup(idGroup int) (bool, error)
	ApproveAllJoinRequests(idGroup int) (bool, error)
	RejectAllJoinRequests(idGroup int) (bool, error)
	ApproveJoinRequests(id int) (bool, error)
	RejectJoinRequests(id int) (bool, error)
	JoinGroup(idGroup int, idUser int) (GroupResult, error)
	LeaveGroup(idGroup int, idUser int) (GroupResult, error)
}

type groupService struct {
	logger logger.Logger
	post   repository.PostgresRepository
}

func (g *groupService) JoinGroup(idGroup int, idUser int) (GroupResult, error) {
	panic("unimplemented")
}

func (g *groupService) LeaveGroup(idGroup int, idUser int) (GroupResult, error) {
	panic("unimplemented")
}

func (g *groupService) ApproveAllJoinRequests(idGroup int) (bool, error) {
	panic("unimplemented")
}

func (g *groupService) ApproveJoinRequests(id int) (bool, error) {
	panic("unimplemented")
}

func (g *groupService) DeleteGroup(idGroup int) (bool, error) {
	panic("unimplemented")
}

func (g *groupService) RejectAllJoinRequests(idGroup int) (bool, error) {
	panic("unimplemented")
}

func (g *groupService) RejectJoinRequests(id int) (bool, error) {
	panic("unimplemented")
}

func (g *groupService) UpdateGroup(inf GroupUpdateInput) (dto.GroupDto, error) {
	panic("unimplemented")
}

func NewGroupService(logger logger.Logger, rep repository.PostgresRepository) GroupsService {
	return &groupService{
		logger: logger,
		post:   rep,
	}
}

func (s *groupService) CreateGroup(id uint, input CreateGroupInput) (*dto.GroupDto, error) {
	if err := services.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	var contacts map[string]string
	if input.Contacts != "" {
		contacts = parseContacts(input.Contacts)
	}

	var creator models.User
	if _, err := creator.FindUserByID(id, s.post); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("Creator not found", "Id", id)
			return nil, fmt.Errorf("%w по email: %s", ErrUserNotFound, id)
		}
		s.logger.Error("Database error while finding creator", "Id", id, "error", err)
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	var categories []models.Category
	if len(input.Categories) > 0 {
		if err := s.post.Where("id IN ?", input.Categories).Find(&categories).Error; err != nil {
			s.logger.Error("Error loading categories", "categoryIDs", input.Categories, "error", err)
			return nil, fmt.Errorf("%w: %v", ErrCategoriesNotFound, err)
		}

		if len(categories) != len(input.Categories) {
			s.logger.Warn("Not all categories found", "requested", len(input.Categories), "found", len(categories))
		}
	}

	var newGroup *groups.Group
	err := s.post.Transaction(func(tx repository.PostgresRepository) error {
		newGroup = &groups.Group{
			Name:             input.Name,
			Description:      input.Description,
			SmallDescription: input.SmallDescription,
			Image:            input.Image,
			CreaterID:        id,
			IsPrivate:        *input.IsPrivate,
			City:             input.City,
			Categories:       categories,
		}

		if err := tx.Create(newGroup).Error; err != nil {
			return fmt.Errorf("ошибка создания группы: %w", err)
		}

		groupUser := groups.GroupUsers{
			UserID:        id,
			GroupID:       newGroup.ID,
			RoleInGroupID: new(groups.Role_in_group).GetIdRole("Админ", s.post),
		}

		if err := tx.Create(&groupUser).Error; err != nil {
			return fmt.Errorf("ошибка добавления пользователя в группу: %w", err)
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
				if err := tx.Create(&groupContacts).Error; err != nil {
					return fmt.Errorf("ошибка сохранения контактов группы: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed during group creation", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrGroupCreation, err)
	}

	if err := s.post.
		Preload("Categories").
		Preload("Contacts").
		Preload("Creater").
		First(newGroup, newGroup.ID).Error; err != nil {
		s.logger.Warn("Failed to reload group with associations", "groupID", newGroup.ID, "error", err)
	}

	s.logger.Info("Group created successfully", "groupID", newGroup.ID, "name", newGroup.Name, "creatorID", creator.ID)

	groupDto := convertGroupToDto(newGroup)
	return groupDto, nil
}

// парсинга контактов
func parseContacts(contactsStr string) map[string]string {
	contacts := make(map[string]string)
	if contactsStr == "" {
		return contacts
	}

	pairs := strings.Split(contactsStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
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

// Конвертация в GroupDto
func convertGroupToDto(group *groups.Group) *dto.GroupDto {
	if group == nil {
		return nil
	}

	categoryNames := make([]string, 0, len(group.Categories))
	for _, cat := range group.Categories {
		categoryNames = append(categoryNames, cat.Name)
	}

	contacts := make([]dto.ContactDto, 0, len(group.Contacts))
	for _, contact := range group.Contacts {
		contacts = append(contacts, dto.ContactDto{
			Name: contact.Name,
			Link: contact.Link,
		})
	}

	return &dto.GroupDto{
		ID:               group.ID,
		Name:             group.Name,
		Description:      group.Description,
		SmallDescription: group.SmallDescription,
		Image:            group.Image,
		CreatorID:        group.CreaterID,
		CreatorName:      group.Creater.Name,
		CreatorUsername:  group.Creater.Us,
		IsPrivate:        group.IsPrivate,
		City:             group.City,
		Categories:       categoryNames,
		Contacts:         contacts,
		CreatedAt:        group.CreatedAt,
		UpdatedAt:        group.UpdatedAt,
	}
}

type GroupResult struct {
	Message string `json:"message"`
	Joined  bool   `json:"joined"`
}

type JoinGroupInput struct {
	GroupID uint `json:"groupId" binding:"required"`
}

func CreateGroup(email *string, input CreateGroupInput) (group *groups.Group, err error) {

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
		UserID:  creator.ID,
		GroupID: newGroup.ID,
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

func JoinGroup(email string, input JoinGroupInput) (*GroupResult, error) {
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
		return &GroupResult{
			Message: "Заявка на вступление отправлена, ожидайте подтверждения от администратора группы",
			Joined:  false,
		}, nil
	} else {
		member := groups.GroupUsers{
			UserID:  user.ID,
			GroupID: group.ID,
		}
		if err := db.GetDB().Create(&member).Error; err != nil {
			return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
		}
		return &GroupResult{
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
	ID           uint          `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Image        string        `json:"image"`
	City         string        `json:"city"`
	Creater      string        `json:"creater"`
	CountMembers int64         `json:"count_members"`
	Subscription bool          `json:"subscription"`
	Users        []UsersGroups `json:"users"`
	Categories   []*string     `json:"categories"`
	Contacts     []*Contacts   `json:"contacts"`
	// Sessions     []SessionDetailResponse `json:"sessions"`
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
	user := models.User{}
	// if err != nil {
	// 	return nil, fmt.Errorf("пользователь не найден: %v", err)
	// }

	var group groups.Group
	var information GroupInf

	err := db.GetDB().Preload("Categories").
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

	// sessions, err := getGroupSessions(*groupID)
	// if err != nil {
	// 	return nil, err
	// }
	// information.Sessions = sessions

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

// func getGroupSessions(groupID uint64) ([]SessionDetailResponse, error) {
// 	var Gsessions []sessions.Session

// 	err := db.GetDB().
// 		Preload("SessionType").
// 		Preload("SessionPlace").
// 		Where("group_id = ?", groupID).
// 		Order("start_time ASC").
// 		Find(&Gsessions).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(Gsessions) == 0 {
// 		return []SessionDetailResponse{}, nil
// 	}

// 	sessionIDs := make([]uint, len(Gsessions))
// 	for i, session := range Gsessions {
// 		sessionIDs[i] = session.ID
// 	}

// 	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
// 	if err != nil {
// 		log.Printf("Ошибка получения метаданных: %v", err)
// 		metadataMap = make(map[uint]*sessions.SessionMetadata)
// 	}

// 	sessionResponses := make([]SessionDetailResponse, 0, len(Gsessions))
// 	for _, session := range Gsessions {
// 		if session.ID == 0 {
// 			continue
// 		}

// 		var sessionTypeName *string
// 		if session.SessionType.Name != "" {
// 			typeName := session.SessionType.Name
// 			sessionTypeName = &typeName
// 		}

// 		var sessionPlaceName *string
// 		if session.SessionPlace.Title != "" {
// 			placeName := session.SessionPlace.Title
// 			sessionPlaceName = &placeName
// 		}

// 		var title, imageURL *string
// 		if session.Title != "" {
// 			titleCopy := session.Title
// 			title = &titleCopy
// 		}
// 		if session.ImageURL != "" {
// 			imageURLCopy := session.ImageURL
// 			imageURL = &imageURLCopy
// 		}

// 		var duration, currentUsers, countUsersMax *uint16
// 		if session.Duration != 0 {
// 			durationCopy := session.Duration
// 			duration = &durationCopy
// 		}
// 		if session.CurrentUsers != 0 {
// 			currentUsersCopy := session.CurrentUsers
// 			currentUsers = &currentUsersCopy
// 		}
// 		if session.CountUsersMax != 0 {
// 			countUsersMaxCopy := session.CountUsersMax
// 			countUsersMax = &countUsersMaxCopy
// 		}

// 		var groupIDCopy *uint
// 		if session.GroupID != 0 {
// 			groupID := session.GroupID
// 			groupIDCopy = &groupID
// 		}

// 		var sessionIDCopy *uint
// 		if session.ID != 0 {
// 			sessionID := session.ID
// 			sessionIDCopy = &sessionID
// 		}

// 		subSession := SubSessionDetail{
// 			ID:            *sessionIDCopy,
// 			Title:         *title,
// 			SessionType:   *sessionTypeName,
// 			SessionPlace:  *sessionPlaceName,
// 			GroupID:       *groupIDCopy,
// 			StartTime:     session.StartTime,
// 			EndTime:       session.EndTime,
// 			Duration:      *duration,
// 			CurrantUsers:  *currentUsers,
// 			CountUsersMax: *countUsersMax,
// 			ImageURL:      *imageURL,
// 		}

// 		// Получаем метаданные для текущей сессии
// 		var metadata *sessions.SessionMetadata
// 		if metaData, exists := metadataMap[session.ID]; exists {
// 			metadata = metaData
// 		}

// 		sessionResponse := SessionDetailResponse{
// 			Session:  subSession,
// 			Metadata: metadata,
// 		}

// 		sessionResponses = append(sessionResponses, sessionResponse)
// 	}

// 	return sessionResponses, nil
// }
