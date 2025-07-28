package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"log"
	"strings"

	"gorm.io/gorm"
)

type CreateGroupInput struct {
	Name             string  `form:"name" binding:"required" example:"Моя группа"`
	Description      string  `form:"description" binding:"required" example:"Полное описание группы"`
	SmallDescription string  `form:"smallDescription" binding:"required" example:"Короткое описание"`
	Image            string  `form:"-"`
	IsPrivate        *bool   `form:"isPrivate" binding:"required" example:"true/1/0/false"`
	City             string  `form:"city,omitempty" example:"Москва"`
	Categories       []*uint `form:"categories" binding:"required" example:"[1, 2, 3]"`
	Contacts         string  `form:"contacts,omitempty" example:"vk:https://vk.com/mygroup, tg:https://t.me/mygroup, inst:https://instagram.com/mygroup"`
}

type JoinGroupResult struct {
	Message string
	Joined  bool
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

	// Сохраняем контакты, если они есть
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

	// Перезагружаем группу с ассоциациями
	if errReload := db.GetDB().Preload("Categories").Preload("Contacts").Preload("Creater").First(newGroup, newGroup.ID).Error; errReload != nil {
		log.Printf("Не удалось перезагрузить группу с ассоциациями: %v", errReload)
	}

	return newGroup, nil
}

func JoinGroup(email string, input JoinGroupInput) (*JoinGroupResult, error) {
	// Найти пользователя
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	// Проверить, есть ли уже заявка или участие
	var existing groups.GroupUsers
	if err := db.GetDB().
		Where("user_id = ? AND group_id = ?", user.ID, input.GroupID).
		First(&existing).Error; err == nil {
		return nil, fmt.Errorf("пользователь уже в группе")
	}

	// Проверка на уже поданную заявку
	var existingRequest groups.GroupJoinRequest
	if err := db.GetDB().
		Where("user_id = ? AND group_id = ?", user.ID, input.GroupID).
		First(&existingRequest).Error; err == nil {
		if existingRequest.Status == "pending" {
			return nil, fmt.Errorf("заявка уже отправлена и ожидает подтверждения")
		}
	}

	// Получить группу
	var group groups.Group
	if err := db.GetDB().Where("id = ?", input.GroupID).First(&group).Error; err != nil {
		return nil, fmt.Errorf("группа не найдена")
	}

	if group.IsPrivate {
		// Группа закрыта — создаём заявку
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
		}, nil // возвращаем группу с инфой, но не добавляем в participants
	} else {
		// Открытая группа — добавляем напрямую
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
	// Проверка прав администратора выполняется в middleware.
	// GORM при удалении записи автоматически удалит связанные данные
	// из many-to-many таблицы (group_group_categories).
	// Для has-many связей (GroupUsers, GroupJoinRequest, GroupContact)
	// мы полагаемся на `ON DELETE CASCADE`, настроенный в БД через теги `constraint`.
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
	CountMembers int64                   `json:"count_members"`
	Users        []UsersGroups           `json:"users"`
	Categories   []string                `json:"categories"`
	Contacts     []Contacts              `json:"contacts"`
	Sessions     []SessionDetailResponse `json:"sessions"`
}

type UsersGroups struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Contacts struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// func GetGroupInf(groupID uint) (*groups.Group, error) {

// }

//Admin часть групп

type GroupUpdateInput struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	SmallDescription *string `json:"small_description"`
	Image            *string `json:"image"`
	IsPrivate        *bool   `json:"is_private"`
	City             *string `json:"city"`
	Contacts         *string `json:"contacts"` // Изменено: теперь строка вместо map
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

	// Обрабатываем контакты, если они переданы
	if input.Contacts != nil {
		newContacts := parseContacts(*input.Contacts)
		if err = updateContactsInTx(tx, &groupID, newContacts); err != nil {
			return fmt.Errorf("ошибка обновления контактов: %v", err)
		}
	}

	return tx.Commit().Error
}

// обрабатывает логику обновления контактов группы в рамках транзакции.
func updateContactsInTx(tx *gorm.DB, groupID *uint, newContacts map[string]string) error {
	// Получаем все существующие контакты группы
	var existingContacts []groups.GroupContact
	if err := tx.Where("group_id = ?", groupID).Find(&existingContacts).Error; err != nil {
		return fmt.Errorf("ошибка получения существующих контактов: %v", err)
	}

	// Создаем map для быстрого поиска существующих контактов
	existingMap := make(map[string]groups.GroupContact)
	for _, c := range existingContacts {
		if c.Name != "" {
			existingMap[c.Name] = c
		}
	}

	// Обрабатываем новые контакты
	for name, link := range newContacts {
		if existingContact, ok := existingMap[name]; ok {
			// Контакт существует - обновляем, если ссылка изменилась
			if existingContact.Link != "" && existingContact.Link != link {
				if err := tx.Model(&existingContact).Update("link", link).Error; err != nil {
					return fmt.Errorf("ошибка обновления контакта '%s': %v", name, err)
				}
			}
			// Убираем из map, чтобы потом не удалить
			delete(existingMap, name)
		} else {
			// Новый контакт - создаем
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

	// Удаляем контакты, которые не были переданы в новом списке
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
