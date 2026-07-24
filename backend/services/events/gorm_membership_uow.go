package events

import (
	"context"
	"errors"
	"fmt"
	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	"friendship/repository"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type eventRepositoryTransactor interface {
	TransactionWithContext(ctx context.Context, fn func(repository.PostgresRepository) error) error
}

type gormEventUnitOfWork struct {
	transactor eventRepositoryTransactor
}

type gormEventMembershipStore struct {
	repo repository.PostgresRepository
}

type gormEventAuditStore struct {
	repo repository.PostgresRepository
	ctx  context.Context
}

func NewGORMEventUnitOfWork(transactor eventRepositoryTransactor) EventUnitOfWork {
	if isNilEventRepositoryTransactor(transactor) {
		return unsupportedEventUnitOfWork{}
	}

	return &gormEventUnitOfWork{transactor: transactor}
}

func (uow *gormEventUnitOfWork) WithinTransaction(ctx context.Context, fn func(EventTransaction) error) error {
	if ctx == nil {
		return errors.New("контекст транзакции события не задан")
	}
	if fn == nil {
		return errEventTransactionCallbackMissing
	}

	return uow.transactor.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		return fn(NewEventTransaction(EventTransactionStores{
			MembershipStore: &gormEventMembershipStore{repo: tx},
			CommandStore:    &gormEventCommandStore{repo: tx},
			AdminStore:      &gormEventAdminStore{repo: tx},
			AuditStore:      &gormEventAuditStore{repo: tx, ctx: ctx},
		}))
	})
}

func isNilEventRepositoryTransactor(transactor eventRepositoryTransactor) bool {
	if transactor == nil {
		return true
	}

	value := reflect.ValueOf(transactor)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (s *gormEventMembershipStore) FindEvent(eventID uint) (EventMembershipSnapshot, error) {
	var event eventmodels.Event
	if err := s.repo.Select("id", "group_id", "creator_id", "title", "start_time").First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventMembershipSnapshot{}, ErrEventNotFound
		}
		return EventMembershipSnapshot{}, fmt.Errorf("ошибка поиска события: %w", err)
	}

	return EventMembershipSnapshot{
		ID:        event.ID,
		GroupID:   event.GroupID,
		CreatorID: event.CreatorID,
		Title:     event.Title,
		StartTime: event.StartTime,
	}, nil
}

func (s *gormEventMembershipStore) IsGroupMember(userID uint, groupID uint) (bool, error) {
	var count int64
	if err := s.repo.Model(&groupmodels.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства в группе: %w", err)
	}

	return count > 0, nil
}

func (s *gormEventMembershipStore) IsParticipant(userID uint, eventID uint) (bool, error) {
	var count int64
	if err := s.repo.Model(&eventmodels.EventsUser{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки участия в событии: %w", err)
	}

	return count > 0, nil
}

func (s *gormEventMembershipStore) AddParticipant(userID uint, eventID uint, joinedAt time.Time) error {
	participant := eventmodels.EventsUser{
		EventID:  eventID,
		UserID:   userID,
		JoinedAt: joinedAt,
	}
	if err := s.repo.Create(&participant).Error; err != nil {
		if isEventMembershipUniqueViolation(err) {
			return ErrAlreadyJoined
		}
		return fmt.Errorf("ошибка добавления к событию: %w", err)
	}

	return nil
}

func (s *gormEventMembershipStore) IncrementParticipantsIfSpace(eventID uint) (bool, error) {
	result := s.repo.Model(&eventmodels.Event{}).
		Where("id = ? AND current_users < max_users", eventID).
		UpdateColumn("current_users", gorm.Expr("current_users + ?", 1))
	if result.Error != nil {
		return false, fmt.Errorf("ошибка обновления счетчика участников: %w", result.Error)
	}

	return result.RowsAffected > 0, nil
}

func (s *gormEventMembershipStore) RemoveParticipant(userID uint, eventID uint) error {
	result := s.repo.
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Delete(&eventmodels.EventsUser{})
	if result.Error != nil {
		return fmt.Errorf("ошибка удаления из события: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotJoined
	}

	return nil
}

func (s *gormEventMembershipStore) DecrementParticipants(eventID uint) error {
	result := s.repo.Model(&eventmodels.Event{}).
		Where("id = ?", eventID).
		UpdateColumn("current_users", gorm.Expr("CASE WHEN current_users > 0 THEN current_users - 1 ELSE 0 END"))
	if result.Error != nil {
		return fmt.Errorf("ошибка обновления счетчика участников: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (s *gormEventAuditStore) RecordBestEffort(input EventAuditInput) error {
	// Вложенная транзакция создает точку сохранения: ошибка необязательного аудита не ломает основную операцию.
	return s.repo.TransactionWithContext(s.ctx, func(tx repository.PostgresRepository) error {
		var actor models.User
		if err := tx.First(&actor, input.ActorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		var membership groupmodels.GroupUsers
		if err := tx.Where("user_id = ? AND group_id = ?", input.ActorID, input.GroupID).First(&membership).Error; err != nil {
			return fmt.Errorf("ошибка поиска членства в группе: %w", err)
		}

		var role groupmodels.Role_in_group
		if err := tx.First(&role, membership.RoleInGroupID).Error; err != nil {
			return fmt.Errorf("ошибка получения роли: %w", err)
		}

		var actionType groupmodels.GroupActionType
		if err := tx.Where("code = ?", input.Action).First(&actionType).Error; err != nil {
			return fmt.Errorf("тип действия группы %q не найден: %w", input.Action, err)
		}

		action := groupmodels.GroupActionLog{
			GroupID:      input.GroupID,
			UserID:       input.ActorID,
			Username:     actor.Name,
			Us:           actor.Us,
			Role:         groupmodels.NormalizeRoleName(role.Name),
			ActionTypeID: actionType.ID,
			TargetUserID: input.TargetUserID,
			EntityID:     input.EntityID,
			EntityName:   input.EntityName,
			CreatedAt:    input.CreatedAt,
		}

		return tx.Create(&action).Error
	})
}

func isEventMembershipUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code == "23505" && postgresError.ConstraintName == "idx_event_user_membership"
	}

	errorText := err.Error()
	return strings.Contains(errorText, "UNIQUE constraint failed: events_users.event_id, events_users.user_id") ||
		strings.Contains(errorText, "UNIQUE constraint failed: events_users.user_id, events_users.event_id")
}
