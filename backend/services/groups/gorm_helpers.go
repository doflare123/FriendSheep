package group

import (
	"errors"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

func isGroupRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func findGroupRoleID(store repository.PostgresRepository, roleName string) (uint, error) {
	var role groups.Role_in_group
	if err := store.Where("name = ?", roleName).First(&role).Error; err != nil {
		return 0, err
	}

	return role.Id, nil
}

func findGroupActionTypeID(store repository.PostgresRepository, code string) (uint, error) {
	var actionType groups.GroupActionType
	if err := store.Where("code = ?", code).First(&actionType).Error; err != nil {
		return 0, err
	}

	return actionType.ID, nil
}
