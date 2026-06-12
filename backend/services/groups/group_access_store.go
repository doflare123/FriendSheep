package group

import (
	"errors"
	"fmt"
	"friendship/models/groups"
	"time"

	"gorm.io/gorm"
)

type txGroupAccessStore struct {
	tx groupTx
}

type groupActorRoleFinder interface {
	FindActorRole(actorID uint, groupID uint, required groups.Capability) (bool, string, error)
}

func newTxGroupAccessStore(tx groupTx) txGroupAccessStore {
	return txGroupAccessStore{tx: tx}
}

func (s txGroupAccessStore) FindActorRole(actorID uint, groupID uint, required groups.Capability) (bool, string, error) {
	var groupUser groups.GroupUsers
	err := s.tx.
		Where("user_id = ? AND group_id = ?", actorID, groupID).
		First(&groupUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", ErrNotInGroup
		}
		return false, "", fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	var role groups.Role_in_group
	if err := s.tx.First(&role, groupUser.RoleInGroupID).Error; err != nil {
		return false, "", fmt.Errorf("ошибка получения роли: %w", err)
	}

	normalizedRole := groups.NormalizeRoleName(role.Name)
	if groups.HasCapability(role.Name, required) {
		return true, normalizedRole, nil
	}

	return false, normalizedRole, nil
}

func createGroupActionLog(tx groupTx, groupID uint, actor joinRequestActor, role string, action string, description string) error {
	actionLog := groups.GroupActionLog{
		GroupID:     groupID,
		UserID:      actor.ID,
		Username:    actor.Name,
		Us:          actor.Us,
		Role:        role,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return tx.Create(&actionLog).Error
}
