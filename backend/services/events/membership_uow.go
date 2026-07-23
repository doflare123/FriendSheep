package events

import (
	"context"
	"errors"
	"time"
)

var (
	errEventUnitOfWorkUnavailable      = errors.New("хранилище событий не поддерживает транзакции")
	errEventTransactionCallbackMissing = errors.New("обработчик транзакции события не задан")
	errEventCommandStoreUnavailable    = errors.New("хранилище команд событий не настроено")
)

type EventMembershipSnapshot struct {
	ID        uint
	GroupID   uint
	CreatorID uint
	Title     string
}

type EventAuditInput struct {
	GroupID      uint
	ActorID      uint
	Action       string
	TargetUserID *uint
	EntityID     *uint
	EntityName   string
	CreatedAt    time.Time
}

type EventMembershipStore interface {
	FindEvent(eventID uint) (EventMembershipSnapshot, error)
	IsGroupMember(userID uint, groupID uint) (bool, error)
	IsParticipant(userID uint, eventID uint) (bool, error)
	AddParticipant(userID uint, eventID uint, joinedAt time.Time) error
	IncrementParticipantsIfSpace(eventID uint) (bool, error)
	RemoveParticipant(userID uint, eventID uint) error
	DecrementParticipants(eventID uint) error
}

type EventAuditStore interface {
	// RecordBestEffort обязана изолировать свою ошибку от основной транзакции события.
	RecordBestEffort(input EventAuditInput) error
}

type EventTransaction struct {
	membership EventMembershipStore
	commands   EventCommandStore
	admin      EventAdminStore
	audit      EventAuditStore
}

type EventTransactionStores struct {
	MembershipStore EventMembershipStore
	CommandStore    EventCommandStore
	AdminStore      EventAdminStore
	AuditStore      EventAuditStore
}

func NewEventTransaction(stores EventTransactionStores) EventTransaction {
	return EventTransaction{
		membership: stores.MembershipStore,
		commands:   stores.CommandStore,
		admin:      stores.AdminStore,
		audit:      stores.AuditStore,
	}
}

func (tx EventTransaction) Membership() EventMembershipStore {
	return tx.membership
}

func (tx EventTransaction) Commands() EventCommandStore {
	return tx.commands
}

func (tx EventTransaction) Admin() EventAdminStore {
	return tx.admin
}

func (tx EventTransaction) Audit() EventAuditStore {
	return tx.audit
}

type EventUnitOfWork interface {
	// WithinTransaction привязывает ctx ко всем хранилищам EventTransaction на время выполнения функции; методы хранилищ намеренно не принимают отдельный контекст.
	WithinTransaction(ctx context.Context, fn func(EventTransaction) error) error
}

type unsupportedEventUnitOfWork struct{}

func (unsupportedEventUnitOfWork) WithinTransaction(context.Context, func(EventTransaction) error) error {
	return errEventUnitOfWorkUnavailable
}
