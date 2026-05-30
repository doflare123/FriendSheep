package group

import (
	"errors"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type groupStore interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Limit(limit int) *gorm.DB
}

type groupTx interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Clauses(conds ...clause.Expression) *gorm.DB
}

type groupRoleStore interface {
	Where(query interface{}, args ...interface{}) *gorm.DB
}

type groupTransactionRunner interface {
	WithinTransaction(func(groupTx) error) error
}

type groupRepositoryTransactor interface {
	Transaction(func(tx repository.PostgresRepository) error) error
}

type groupRepositoryTransactionRunner struct {
	transactor groupRepositoryTransactor
}

type groupUnsupportedTransactionRunner struct {
	err error
}

func newGroupTransactionRunner(store interface{}) groupTransactionRunner {
	transactor, ok := store.(groupRepositoryTransactor)
	if !ok {
		return groupUnsupportedTransactionRunner{
			err: errors.New("хранилище сервиса групп не поддерживает транзакции"),
		}
	}

	return groupRepositoryTransactionRunner{transactor: transactor}
}

func (r groupRepositoryTransactionRunner) WithinTransaction(fn func(groupTx) error) error {
	return r.transactor.Transaction(func(tx repository.PostgresRepository) error {
		return fn(tx)
	})
}

func (r groupUnsupportedTransactionRunner) WithinTransaction(func(groupTx) error) error {
	return r.err
}

func (s *groupService) runInTx(fn func(groupTx) error) error {
	return s.tx.WithinTransaction(fn)
}

func findGroupRoleID(store groupRoleStore, roleName string) (uint, error) {
	var role groups.Role_in_group
	if err := store.Where("name = ?", roleName).First(&role).Error; err != nil {
		return 0, err
	}

	return role.Id, nil
}
