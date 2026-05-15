package events

import (
	"friendship/repository"

	"gorm.io/gorm"
)

type eventsRepoPort interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Transaction(func(tx repository.PostgresRepository) error) error
}

type eventsTxPort interface {
	Model(value interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
}

type eventsTransactionRunner interface {
	WithinTransaction(func(eventsTxPort) error) error
}

type eventsRepositoryTransactionRunner struct {
	transactor eventsRepoPort
}

func newEventsTransactionRunner(store eventsRepoPort) eventsTransactionRunner {
	return eventsRepositoryTransactionRunner{transactor: store}
}

func (r eventsRepositoryTransactionRunner) WithinTransaction(fn func(eventsTxPort) error) error {
	return r.transactor.Transaction(func(tx repository.PostgresRepository) error {
		return fn(tx)
	})
}

func (s *eventsService) runInTx(fn func(eventsTxPort) error) error {
	return s.tx.WithinTransaction(fn)
}
