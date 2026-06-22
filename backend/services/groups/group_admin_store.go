package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"time"

	"gorm.io/gorm"
)

type groupAdminStore interface {
	groupActorFinder

	DeleteGroup(groupID uint) error
	UpdateGroup(input GroupUpdateInput, actor joinRequestActor, actorRole string) (groupAdminResult, error)
	ChangeMemberRole(input GroupUserInput, roleName string, actor joinRequestActor, actorRole string) (groupAdminResult, error)
	BanMember(groupID uint, targetUserID uint, actor joinRequestActor, actorRole string) (groupAdminResult, error)
	RemoveFromBlacklist(groupID uint, targetUserID uint, actor joinRequestActor, actorRole string) (groupAdminResult, error)
}

type groupAdminResult struct {
	Warnings []groupAdminLogWarning
}

type groupAdminLogWarning struct {
	Message string
	Args    []interface{}
}

type gormGroupAdminStore struct {
	tx groupTx
	txGroupActorStore
}

func newGroupAdminStore(tx groupTx) groupAdminStore {
	return gormGroupAdminStore{
		tx:                tx,
		txGroupActorStore: newTxGroupActorStore(tx),
	}
}

func (s gormGroupAdminStore) DeleteGroup(groupID uint) error {
	var group groups.Group
	if err := s.tx.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupNotFound
		}
		return fmt.Errorf("ошибка поиска группы: %w", err)
	}

	if err := s.tx.Delete(&group).Error; err != nil {
		return fmt.Errorf("ошибка удаления группы: %w", err)
	}

	return nil
}

func (s gormGroupAdminStore) UpdateGroup(input GroupUpdateInput, actor joinRequestActor, actorRole string) (groupAdminResult, error) {
	if !groups.HasCapability(actorRole, groups.CapabilityModerate) {
		return groupAdminResult{}, ErrPermissionDenied
	}

	var group groups.Group
	if err := s.tx.First(&group, input.GroupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return groupAdminResult{}, ErrGroupNotFound
		}
		return groupAdminResult{}, fmt.Errorf("ошибка поиска группы: %w", err)
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
		if err := s.tx.Model(&group).Updates(updates).Error; err != nil {
			return groupAdminResult{}, fmt.Errorf("не удалось сохранить изменения группы: %w", err)
		}
	}

	if input.Categories != nil {
		if err := s.tx.Model(&group).Association("Categories").Clear(); err != nil {
			return groupAdminResult{}, fmt.Errorf("не удалось очистить старые категории: %w", err)
		}

		if len(input.Categories) > 0 {
			var newCategories []models.Category
			if err := s.tx.Where("id IN ?", input.Categories).Find(&newCategories).Error; err != nil {
				return groupAdminResult{}, fmt.Errorf("не удалось найти переданные категории: %w", err)
			}
			if err := s.tx.Model(&group).Association("Categories").Replace(&newCategories); err != nil {
				return groupAdminResult{}, fmt.Errorf("не удалось назначить новые категории: %w", err)
			}
		}
	}

	if input.Contacts != nil {
		newContacts := parseContacts(*input.Contacts)
		if err := updateContactsInTx(s.tx, group.ID, newContacts); err != nil {
			return groupAdminResult{}, fmt.Errorf("ошибка обновления контактов: %w", err)
		}
	}

	result := groupAdminResult{}
	if err := createGroupActionLog(s.tx, groupActionLogInput{
		GroupID: group.ID,
		Actor:   actor,
		Role:    actorRole,
		Action:  groups.ActionUpdateGroup,
	}); err != nil {
		result.Warnings = append(result.Warnings, groupAdminLogWarning{
			Message: "Группа обновлена, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return result, nil
}

func (s gormGroupAdminStore) ChangeMemberRole(input GroupUserInput, roleName string, actor joinRequestActor, actorRole string) (groupAdminResult, error) {
	targetUser, err := s.findTargetUser(input.UserID)
	if err != nil {
		return groupAdminResult{}, err
	}

	groupUser, err := s.findGroupMember(input.GroupID, input.UserID)
	if err != nil {
		return groupAdminResult{}, err
	}

	roleID, err := findGroupRoleID(s.tx, roleName)
	if err != nil {
		return groupAdminResult{}, roleNotFoundError(roleName)
	}

	if err := s.tx.Model(&groupUser).Update("role_in_group_id", roleID).Error; err != nil {
		return groupAdminResult{}, fmt.Errorf("ошибка обновления роли: %w", err)
	}

	result := groupAdminResult{}
	actionType := roleChangeAction(roleName)
	targetUserID := targetUser.ID
	if err := createGroupActionLog(s.tx, groupActionLogInput{
		GroupID:      input.GroupID,
		Actor:        actor,
		Role:         actorRole,
		Action:       actionType,
		TargetUserID: &targetUserID,
	}); err != nil {
		result.Warnings = append(result.Warnings, groupAdminLogWarning{
			Message: "Роль участника изменена, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return result, nil
}

func (s gormGroupAdminStore) BanMember(groupID uint, targetUserID uint, actor joinRequestActor, actorRole string) (groupAdminResult, error) {
	if _, err := s.findTargetUser(targetUserID); err != nil {
		return groupAdminResult{}, err
	}

	groupUser, err := s.findGroupMember(groupID, targetUserID)
	if err != nil {
		return groupAdminResult{}, err
	}

	var targetRole groups.Role_in_group
	if err := s.tx.First(&targetRole, groupUser.RoleInGroupID).Error; err == nil {
		if groups.HasCapability(targetRole.Name, groups.CapabilityAdmin) {
			return groupAdminResult{}, fmt.Errorf("нельзя удалить администратора группы")
		}
	}

	if err := s.tx.Delete(&groupUser).Error; err != nil {
		return groupAdminResult{}, fmt.Errorf("ошибка удаления участника: %w", err)
	}

	blacklist := groups.GroupBlacklist{
		GroupID:   groupID,
		UserID:    targetUserID,
		BannedBy:  actor.ID,
		Reason:    "Удален из группы",
		CreatedAt: time.Now(),
	}
	if err := s.tx.Create(&blacklist).Error; err != nil {
		return groupAdminResult{}, fmt.Errorf("ошибка добавления в черный список: %w", err)
	}

	result := groupAdminResult{}
	if err := createGroupActionLog(s.tx, groupActionLogInput{
		GroupID:      groupID,
		Actor:        actor,
		Role:         actorRole,
		Action:       groups.ActionBanUser,
		TargetUserID: &targetUserID,
	}); err != nil {
		result.Warnings = append(result.Warnings, groupAdminLogWarning{
			Message: "Пользователь удален из группы и добавлен в черный список, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return result, nil
}

func (s gormGroupAdminStore) RemoveFromBlacklist(groupID uint, targetUserID uint, actor joinRequestActor, actorRole string) (groupAdminResult, error) {
	if _, err := s.findTargetUser(targetUserID); err != nil {
		return groupAdminResult{}, err
	}

	result := s.tx.Where("group_id = ? AND user_id = ?", groupID, targetUserID).
		Delete(&groups.GroupBlacklist{})
	if result.Error != nil {
		return groupAdminResult{}, fmt.Errorf("ошибка удаления из черного списка: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return groupAdminResult{}, fmt.Errorf("пользователь не найден в черном списке")
	}

	adminResult := groupAdminResult{}
	if err := createGroupActionLog(s.tx, groupActionLogInput{
		GroupID:      groupID,
		Actor:        actor,
		Role:         actorRole,
		Action:       groups.ActionUnbanUser,
		TargetUserID: &targetUserID,
	}); err != nil {
		adminResult.Warnings = append(adminResult.Warnings, groupAdminLogWarning{
			Message: "Пользователь убран из черного списка, но действие не записано в журнал группы",
			Args:    []interface{}{"error", err},
		})
	}

	return adminResult, nil
}

func (s gormGroupAdminStore) findTargetUser(userID uint) (models.User, error) {
	var targetUser models.User
	if err := s.tx.First(&targetUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return targetUser, nil
}

func (s gormGroupAdminStore) findGroupMember(groupID uint, userID uint) (groups.GroupUsers, error) {
	var groupUser groups.GroupUsers
	err := s.tx.Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&groupUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return groups.GroupUsers{}, ErrNotInGroup
		}
		return groups.GroupUsers{}, fmt.Errorf("ошибка поиска участника: %w", err)
	}

	return groupUser, nil
}

func updateContactsInTx(tx groupTx, groupID uint, newContacts map[string]string) error {
	var existingContacts []groups.GroupContact
	if err := tx.Where("group_id = ?", groupID).Find(&existingContacts).Error; err != nil {
		return fmt.Errorf("ошибка получения существующих контактов: %w", err)
	}

	existingMap := make(map[string]groups.GroupContact)
	for _, c := range existingContacts {
		if c.Name != "" {
			existingMap[c.Name] = c
		}
	}

	for name, link := range newContacts {
		if existingContact, ok := existingMap[name]; ok {
			if existingContact.Link != "" && existingContact.Link != link {
				if err := tx.Model(&existingContact).Update("link", link).Error; err != nil {
					return fmt.Errorf("ошибка обновления контакта '%s': %w", name, err)
				}
			}
			delete(existingMap, name)
		} else {
			newContact := groups.GroupContact{
				GroupID: groupID,
				Name:    name,
				Link:    link,
			}
			if err := tx.Create(&newContact).Error; err != nil {
				return fmt.Errorf("ошибка добавления контакта '%s': %w", name, err)
			}
		}
	}

	for _, contactToDelete := range existingMap {
		if err := tx.Delete(&contactToDelete).Error; err != nil {
			contactName := "unknown"
			if contactToDelete.Name != "" {
				contactName = contactToDelete.Name
			}
			return fmt.Errorf("ошибка удаления старого контакта '%s': %w", contactName, err)
		}
	}

	return nil
}

func roleNotFoundError(roleName string) error {
	switch roleName {
	case groups.RoleModerator:
		return ErrRoleModeratorNotFound
	case groups.RoleMember:
		return ErrRoleMemberNotFound
	case groups.RoleAdmin:
		return ErrRoleAdminNotFound
	default:
		return fmt.Errorf("роль %q не найдена", roleName)
	}
}

func roleChangeAction(roleName string) string {
	switch roleName {
	case groups.RoleModerator:
		return groups.ActionAddOperator
	case groups.RoleMember:
		return groups.ActionRemoveOperator
	default:
		return groups.ActionChangeRole
	}
}
