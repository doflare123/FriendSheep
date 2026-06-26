package group

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"friendship/repository"
	"strings"
	"time"
)

type groupRepositoryTransactor interface {
	Transaction(func(tx repository.PostgresRepository) error) error
}

type groupRepositoryTransactionRunner struct {
	transactor groupRepositoryTransactor
}

type groupUnsupportedTransactionRunner struct {
	err error
}

type gormGroupRepository struct {
	store repository.PostgresRepository
	tx    groupRepositoryTransactionRunner
}

type gormGroupTxAdapter struct {
	tx repository.PostgresRepository
}

// NewGORMGroupRepository wraps the temporary shared Postgres repository behind
// the pure group service storage port.
func NewGORMGroupRepository(store repository.PostgresRepository) groupUnitOfWork {
	return gormGroupRepository{
		store: store,
		tx:    newGroupTransactionRunner(store),
	}
}

func newGroupTransactionRunner(store interface{}) groupRepositoryTransactionRunner {
	transactor, ok := store.(groupRepositoryTransactor)
	if !ok {
		transactor = groupUnsupportedTransactionRunner{
			err: errors.New("group service store does not support transactions"),
		}
	}

	return groupRepositoryTransactionRunner{transactor: transactor}
}

func (r groupRepositoryTransactionRunner) WithinTransaction(fn func(groupTx) error) error {
	return r.transactor.Transaction(func(tx repository.PostgresRepository) error {
		return fn(gormGroupTxAdapter{tx: tx})
	})
}

func (r groupUnsupportedTransactionRunner) Transaction(func(repository.PostgresRepository) error) error {
	return r.err
}

func (r gormGroupRepository) WithinTransaction(fn func(groupTx) error) error {
	return r.tx.WithinTransaction(fn)
}

func (r gormGroupRepository) Access() groupActorRoleFinder {
	return newGroupAccessStore(r.store)
}

func (r gormGroupRepository) Reads() groupManagementReadStore {
	return newGroupManagementReadStore(r.store)
}

func (tx gormGroupTxAdapter) Admin() groupAdminStore {
	return newGroupAdminStore(tx.tx)
}

func (tx gormGroupTxAdapter) JoinGroup() joinGroupStore {
	return newJoinGroupStore(tx.tx)
}

func (tx gormGroupTxAdapter) LeaveGroup() leaveGroupStore {
	return newLeaveGroupStore(tx.tx)
}

func (tx gormGroupTxAdapter) JoinInviteCreation() joinInviteCreationStore {
	return newJoinInviteCreationStore(tx.tx)
}

func (tx gormGroupTxAdapter) JoinInviteResponse() joinInviteResponseStore {
	return newJoinInviteResponseStore(tx.tx)
}

func (tx gormGroupTxAdapter) JoinRequestReview() joinRequestReviewStore {
	return newJoinRequestReviewStore(tx.tx)
}

func (tx gormGroupTxAdapter) BulkJoinRequest() bulkJoinRequestStore {
	return newBulkJoinRequestStore(tx.tx)
}

func (tx gormGroupTxAdapter) LogActorAction(input groupActorActionLogInput) error {
	return createActorGroupActionLog(tx.tx, input)
}

func findGroupRoleID(store groupRoleStore, roleName string) (uint, error) {
	var role groups.Role_in_group
	if err := store.Where("name = ?", roleName).First(&role).Error; err != nil {
		return 0, err
	}

	return role.Id, nil
}

func createActorGroupActionLog(tx groupPersistence, input groupActorActionLogInput) error {
	actionTypeID, err := groups.FindGroupActionTypeID(tx, input.Action)
	if err != nil {
		return fmt.Errorf("group action type %q not found: %w", input.Action, err)
	}

	username := input.Username
	us := input.Us
	if strings.TrimSpace(username) == "" || strings.TrimSpace(us) == "" {
		var actor models.User
		if err := tx.Select("id", "name", "us").First(&actor, input.UserID).Error; err == nil {
			if strings.TrimSpace(username) == "" {
				username = actor.Name
			}
			if strings.TrimSpace(us) == "" {
				us = actor.Us
			}
		}
	}

	action := groups.GroupActionLog{
		GroupID:      input.GroupID,
		UserID:       input.UserID,
		Username:     username,
		Us:           us,
		Role:         input.Role,
		ActionTypeID: actionTypeID,
		Description:  input.Description,
		TargetUserID: input.TargetUserID,
		CreatedAt:    time.Now(),
	}

	return tx.Create(&action).Error
}
