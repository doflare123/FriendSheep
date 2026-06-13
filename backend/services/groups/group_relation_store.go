package group

import (
	"fmt"
	"friendship/models/groups"
)

type groupRelationChecks interface {
	IsUserBlacklisted(groupID uint, userID uint) (bool, error)
	IsGroupMember(groupID uint, userID uint) (bool, error)
}

type txGroupRelationStore struct {
	tx groupTx
}

func newTxGroupRelationStore(tx groupTx) txGroupRelationStore {
	return txGroupRelationStore{tx: tx}
}

func (s txGroupRelationStore) IsUserBlacklisted(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupBlacklist{}).
		Where(&groups.GroupBlacklist{GroupID: groupID, UserID: userID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки черного списка: %w", err)
	}

	return count > 0, nil
}

func (s txGroupRelationStore) IsGroupMember(groupID uint, userID uint) (bool, error) {
	var count int64
	if err := s.tx.Model(&groups.GroupUsers{}).
		Where(&groups.GroupUsers{UserID: userID, GroupID: groupID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства: %w", err)
	}

	return count > 0, nil
}
