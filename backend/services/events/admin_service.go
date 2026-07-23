package events

import (
	"context"
	"errors"
	"reflect"
	"time"

	"friendship/logger"
	"friendship/models/dto"
	groupmodels "friendship/models/groups"
)

var (
	errEventAdminReaderUnavailable = errors.New("хранилище административного чтения событий не настроено")
	errEventAdminStoreUnavailable  = errors.New("хранилище административных операций событий не настроено")
	ErrUserNotFound                = errors.New("пользователь не найден")
	ErrActorCantKickSelf           = errors.New("используйте выход из события вместо исключения самого себя")
)

type EventAdminService interface {
	GetEventDetailsForAdmin(ctx context.Context, actorID uint, eventID uint) (*dto.EventAdminDto, error)
	KickUserFromEvent(ctx context.Context, actorID uint, eventID uint, targetUserID uint) (bool, error)
}

type EventAdminReader interface {
	InspectAccess(ctx context.Context, actorID uint, eventID uint) (EventAdminAccessView, error)
	LoadDetails(ctx context.Context, actorID uint, eventID uint) (EventAdminDetailsView, error)
}

type EventAdminStore interface {
	FindEvent(eventID uint) (EventAdminEventSnapshot, error)
	FindGroupRole(actorID uint, groupID uint) (string, error)
	UserExists(userID uint) (bool, error)
	RemoveParticipant(userID uint, eventID uint) error
	DecrementParticipants(eventID uint) error
}

type EventAdminAccessView struct {
	EventFound         bool
	ActorIsGroupMember bool
	ActorRole          string
}

type EventAdminEventSnapshot struct {
	ID        uint
	GroupID   uint
	CreatorID uint
	Title     string
}

type EventAdminDetailsView struct {
	Found        bool
	Access       EventAdminAccessView
	Event        EventFullView
	Participants []EventAdminParticipantView
}

type EventAdminParticipantView struct {
	ID       uint
	UserID   uint
	Name     string
	Username string
	Image    string
	JoinedAt time.Time
}

type eventAdminService struct {
	logger     logger.Logger
	reader     EventAdminReader
	unitOfWork EventUnitOfWork
}

type unsupportedEventAdminReader struct{}

func NewEventAdminService(logger logger.Logger, reader EventAdminReader, unitOfWork EventUnitOfWork) EventAdminService {
	if isNilEventAdminReader(reader) {
		reader = unsupportedEventAdminReader{}
	}
	if unitOfWork == nil {
		unitOfWork = unsupportedEventUnitOfWork{}
	}

	return &eventAdminService{
		logger:     logger,
		reader:     reader,
		unitOfWork: unitOfWork,
	}
}

func isNilEventAdminReader(reader EventAdminReader) bool {
	if reader == nil {
		return true
	}

	value := reflect.ValueOf(reader)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (unsupportedEventAdminReader) InspectAccess(context.Context, uint, uint) (EventAdminAccessView, error) {
	return EventAdminAccessView{}, errEventAdminReaderUnavailable
}

func (unsupportedEventAdminReader) LoadDetails(context.Context, uint, uint) (EventAdminDetailsView, error) {
	return EventAdminDetailsView{}, errEventAdminReaderUnavailable
}

func (s *eventAdminService) GetEventDetailsForAdmin(
	ctx context.Context,
	actorID uint,
	eventID uint,
) (*dto.EventAdminDto, error) {
	access, err := s.reader.InspectAccess(ctx, actorID, eventID)
	if err != nil {
		s.logger.Error("Не удалось проверить доступ к событию", "eventID", eventID, "actorID", actorID, "error", err)
		return nil, err
	}
	if !access.EventFound {
		return nil, ErrEventNotFound
	}
	if !access.ActorIsGroupMember {
		return nil, ErrNotGroupMember
	}
	if !groupmodels.HasCapability(access.ActorRole, groupmodels.CapabilityModerate) {
		return nil, ErrPermissionDenied
	}

	details, err := s.reader.LoadDetails(ctx, actorID, eventID)
	if err != nil {
		s.logger.Error("Не удалось получить событие для администратора", "eventID", eventID, "actorID", actorID, "error", err)
		return nil, err
	}
	if !details.Access.EventFound {
		return nil, ErrEventNotFound
	}
	if !details.Access.ActorIsGroupMember {
		return nil, ErrNotGroupMember
	}
	if !groupmodels.HasCapability(details.Access.ActorRole, groupmodels.CapabilityModerate) {
		return nil, ErrPermissionDenied
	}
	if !details.Found {
		return nil, ErrEventNotFound
	}

	return eventAdminDetailsToDTO(details), nil
}

func (s *eventAdminService) KickUserFromEvent(
	ctx context.Context,
	actorID uint,
	eventID uint,
	targetUserID uint,
) (bool, error) {
	err := s.unitOfWork.WithinTransaction(ctx, func(tx EventTransaction) error {
		store := tx.Admin()
		if store == nil {
			return errEventAdminStoreUnavailable
		}

		event, err := store.FindEvent(eventID)
		if err != nil {
			return err
		}

		role, err := store.FindGroupRole(actorID, event.GroupID)
		if err != nil {
			return err
		}
		if !groupmodels.HasCapability(role, groupmodels.CapabilityModerate) {
			return ErrPermissionDenied
		}
		if event.CreatorID == targetUserID {
			return ErrCreatorCantLeave
		}
		if actorID == targetUserID {
			return ErrActorCantKickSelf
		}

		targetExists, err := store.UserExists(targetUserID)
		if err != nil {
			return err
		}
		if !targetExists {
			return ErrUserNotFound
		}

		if err := store.RemoveParticipant(targetUserID, eventID); err != nil {
			return err
		}
		if err := store.DecrementParticipants(eventID); err != nil {
			return err
		}

		s.recordAdminAudit(tx.Audit(), event, actorID, targetUserID)
		return nil
	})
	if err != nil {
		s.logger.Error(
			"Не удалось исключить пользователя из события",
			"eventID", eventID,
			"userID", targetUserID,
			"actorID", actorID,
			"error", err,
		)
		return false, err
	}

	s.logger.Info(
		"Пользователь исключен из события",
		"eventID", eventID,
		"kickedUserID", targetUserID,
		"actorID", actorID,
	)
	return true, nil
}

func (s *eventAdminService) recordAdminAudit(
	audit EventAuditStore,
	event EventAdminEventSnapshot,
	actorID uint,
	targetUserID uint,
) {
	if audit == nil {
		return
	}

	entityID := event.ID
	if err := audit.RecordBestEffort(EventAuditInput{
		GroupID:      event.GroupID,
		ActorID:      actorID,
		Action:       groupmodels.ActionKickFromEvent,
		TargetUserID: &targetUserID,
		EntityID:     &entityID,
		EntityName:   event.Title,
		CreatedAt:    time.Now(),
	}); err != nil {
		s.logger.Warn("Не удалось записать действие в журнал", "error", err)
	}
}

func eventAdminDetailsToDTO(details EventAdminDetailsView) *dto.EventAdminDto {
	eventDTO := eventFullViewToDTO(details.Event)
	participants := make([]dto.EventAdminParticipantDto, 0, len(details.Participants))
	for _, participant := range details.Participants {
		participants = append(participants, dto.EventAdminParticipantDto{
			ID:        participant.ID,
			UserID:    participant.UserID,
			Name:      participant.Name,
			Username:  participant.Username,
			Image:     participant.Image,
			IsCreator: participant.UserID == details.Event.Creator.ID,
			JoinedAt:  participant.JoinedAt,
		})
	}

	return &dto.EventAdminDto{
		EventFullDto:    *eventDTO,
		AllParticipants: participants,
	}
}
