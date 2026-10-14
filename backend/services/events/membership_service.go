package events

import (
	"context"
	"errors"
	"friendship/logger"
	groupmodels "friendship/models/groups"
)

var errEventMembershipStoreUnavailable = errors.New("хранилище участия в событиях не настроено")

type EventMembershipService interface {
	JoinEvent(ctx context.Context, userID uint, eventID uint) (bool, error)
	LeaveEvent(ctx context.Context, userID uint, eventID uint) (bool, error)
}

type eventMembershipService struct {
	logger     logger.Logger
	unitOfWork EventUnitOfWork
	clock      Clock
}

func NewEventMembershipService(logger logger.Logger, unitOfWork EventUnitOfWork) EventMembershipService {
	return NewEventMembershipServiceWithClock(logger, unitOfWork, SystemClock{})
}

func NewEventMembershipServiceWithClock(logger logger.Logger, unitOfWork EventUnitOfWork, clock Clock) EventMembershipService {
	if unitOfWork == nil {
		unitOfWork = unsupportedEventUnitOfWork{}
	}
	if clock == nil {
		clock = SystemClock{}
	}

	return &eventMembershipService{
		logger:     logger,
		unitOfWork: unitOfWork,
		clock:      clock,
	}
}

func (s *eventMembershipService) JoinEvent(ctx context.Context, userID uint, eventID uint) (bool, error) {
	err := s.unitOfWork.WithinTransaction(ctx, func(tx EventTransaction) error {
		membership := tx.Membership()
		if membership == nil {
			return errEventMembershipStoreUnavailable
		}

		event, err := membership.FindEvent(eventID)
		if err != nil {
			return err
		}

		isGroupMember, err := membership.IsGroupMember(userID, event.GroupID)
		if err != nil {
			return err
		}
		if !isGroupMember {
			return ErrNotGroupMember
		}

		isParticipant, err := membership.IsParticipant(userID, eventID)
		if err != nil {
			return err
		}
		if isParticipant {
			return ErrAlreadyJoined
		}
		if !s.clock.Now().Before(event.StartTime) {
			return ErrEventAlreadyStarted
		}

		if err := membership.AddParticipant(userID, eventID, s.clock.Now()); err != nil {
			return err
		}

		hasSpace, err := membership.IncrementParticipantsIfSpace(eventID)
		if err != nil {
			return err
		}
		if !hasSpace {
			return ErrEventFull
		}

		s.recordMembershipAudit(tx.Audit(), event, userID, groupmodels.ActionJoinEvent)

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось вступить в событие", "eventID", eventID, "userID", userID, "error", err)
		return false, err
	}

	s.logger.Info("Пользователь вступил в событие", "eventID", eventID, "userID", userID)
	return true, nil
}

func (s *eventMembershipService) LeaveEvent(ctx context.Context, userID uint, eventID uint) (bool, error) {
	err := s.unitOfWork.WithinTransaction(ctx, func(tx EventTransaction) error {
		membership := tx.Membership()
		if membership == nil {
			return errEventMembershipStoreUnavailable
		}

		event, err := membership.FindEvent(eventID)
		if err != nil {
			return err
		}
		if event.CreatorID == userID {
			return ErrCreatorCantLeave
		}
		isParticipant, err := membership.IsParticipant(userID, eventID)
		if err != nil {
			return err
		}
		if !isParticipant {
			return ErrNotJoined
		}
		if eventHasEnded(event.EndTime, s.clock.Now()) {
			return ErrEventAlreadyEnded
		}

		if err := membership.RemoveParticipant(userID, eventID); err != nil {
			return err
		}
		if err := membership.DecrementParticipants(eventID); err != nil {
			return err
		}

		s.recordMembershipAudit(tx.Audit(), event, userID, groupmodels.ActionLeaveEvent)

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось покинуть событие", "eventID", eventID, "userID", userID, "error", err)
		return false, err
	}

	s.logger.Info("Пользователь покинул событие", "eventID", eventID, "userID", userID)
	return true, nil
}

func (s *eventMembershipService) recordMembershipAudit(audit EventAuditStore, event EventMembershipSnapshot, userID uint, action string) {
	if audit == nil {
		return
	}

	targetUserID := userID
	entityID := event.ID
	if err := audit.RecordBestEffort(EventAuditInput{
		GroupID:      event.GroupID,
		ActorID:      userID,
		Action:       action,
		TargetUserID: &targetUserID,
		EntityID:     &entityID,
		EntityName:   event.Title,
		CreatedAt:    s.clock.Now(),
	}); err != nil {
		s.logger.Warn("Не удалось записать действие в журнал", "error", err)
	}
}
