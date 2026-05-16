package events

import (
	"errors"
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

type eventsRepositoryTransactor interface {
	Transaction(func(tx repository.PostgresRepository) error) error
}

type eventsRepositoryTransactionRunner struct {
	transactor eventsRepositoryTransactor
}

type eventsUnsupportedTransactionRunner struct {
	err error
}

func newEventsTransactionRunner(store interface{}) eventsTransactionRunner {
	transactor, ok := store.(eventsRepositoryTransactor)
	if !ok {
		return eventsUnsupportedTransactionRunner{
			err: errors.New("events service store does not support transactions"),
		}
	}

	return eventsRepositoryTransactionRunner{transactor: transactor}
}

func (r eventsRepositoryTransactionRunner) WithinTransaction(fn func(eventsTxPort) error) error {
	return r.transactor.Transaction(func(tx repository.PostgresRepository) error {
		return fn(tx)
	})
}

func (r eventsUnsupportedTransactionRunner) WithinTransaction(func(eventsTxPort) error) error {
	return r.err
}

func (s *eventsService) runInTx(fn func(eventsTxPort) error) error {
	return s.tx.WithinTransaction(fn)
}
