package events

import (
	"context"
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models/dto"
	groupmodels "friendship/models/groups"
	"time"
)

var ErrMaxUsersBelowCurrent = errors.New("нельзя установить максимум меньше текущего количества участников")

type EventCommandService interface {
	CreateEvent(ctx context.Context, actorID uint, input CreateEventInput) (*dto.EventFullDto, error)
	UpdateEvent(ctx context.Context, actorID uint, eventID uint, input UpdateEventInput) (*dto.EventFullDto, error)
	DeleteEvent(ctx context.Context, actorID uint, eventID uint) (bool, error)
}

type EventCommandSnapshot struct {
	ID           uint
	GroupID      uint
	Title        string
	StartTime    time.Time
	Duration     uint16
	CurrentUsers uint16
}

type EventCreateRecord struct {
	Title           string
	Description     string
	GroupID         uint
	EventTypeID     uint
	LocationID      uint
	CreatorID       uint
	StartTime       time.Time
	EndTime         time.Time
	Duration        uint16
	MaxUsers        uint16
	CurrentUsers    uint16
	ImageURL        string
	StatusID        uint
	Address         string
	Country         string
	AgeLimitID      uint
	Year            *int
	Notes           string
	CustomFields    map[string]interface{}
	GenreIDs        []uint
	CreatorJoinedAt time.Time
}

type EventUpdateRecord struct {
	EventID       uint
	Title         *string
	Description   *string
	EventTypeID   *uint
	LocationID    *uint
	ImageURL      *string
	StartTime     *time.Time
	EndTime       *time.Time
	Duration      *uint16
	MaxUsers      *uint16
	Address       *string
	Country       *string
	AgeLimitID    *uint
	Year          *int
	Notes         *string
	CustomFields  *map[string]interface{}
	ReplaceGenres bool
	GenreIDs      []uint
}

type EventCommandGroupView struct {
	ID         uint
	Name       string
	Image      string
	Enterprise bool
}

type EventCommandCreatorView struct {
	ID       uint
	Name     string
	Us       string
	Image    string
	Verified bool
}

type EventCommandView struct {
	ID           uint
	Title        string
	Description  string
	ImageURL     string
	MaxUsers     uint16
	CurrentUsers uint16
	StartTime    time.Time
	EndTime      time.Time
	Duration     uint16
	Group        EventCommandGroupView
	EventTypeID  uint
	LocationID   uint
	StatusID     uint
	Genres       []string
	Creator      EventCommandCreatorView
	Address      string
	Country      string
	AgeLimit     string
	Year         *int
	Notes        string
	CustomFields map[string]interface{}
	Subscribed   bool
	IsCreator    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EventCommandStore interface {
	FindEventForUpdate(eventID uint) (EventCommandSnapshot, error)
	FindGroupRole(actorID uint, groupID uint) (string, error)
	AgeLimitExists(ageLimitID uint) (bool, error)
	CountGenres(genreIDs []uint) (int, error)
	CreateEvent(record EventCreateRecord) (uint, error)
	UpdateEvent(record EventUpdateRecord) error
	DeleteEventAggregate(eventID uint) error
	LoadEventResult(eventID uint, actorID uint) (EventCommandView, error)
}

type eventCommandService struct {
	logger     logger.Logger
	unitOfWork EventUnitOfWork
}

func NewEventCommandService(logger logger.Logger, unitOfWork EventUnitOfWork) EventCommandService {
	if unitOfWork == nil {
		unitOfWork = unsupportedEventUnitOfWork{}
	}

	return &eventCommandService{
		logger:     logger,
		unitOfWork: unitOfWork,
	}
}

func (s *eventCommandService) CreateEvent(ctx context.Context, actorID uint, input CreateEventInput) (*dto.EventFullDto, error) {
	var view EventCommandView
	err := s.unitOfWork.WithinTransaction(ctx, func(tx EventTransaction) error {
		store := tx.Commands()
		if store == nil {
			return errEventCommandStoreUnavailable
		}

		role, err := store.FindGroupRole(actorID, input.GroupID)
		if err != nil {
			return err
		}
		if !groupmodels.HasCapability(role, groupmodels.CapabilityModerate) {
			return ErrPermissionDenied
		}

		if err := validateEventGenreCount(input.Genres); err != nil {
			return err
		}
		ageLimitExists, err := store.AgeLimitExists(input.AgeLimitID)
		if err != nil {
			return err
		}
		if !ageLimitExists {
			return ErrAgeLimitNotFound
		}
		genreCount, err := store.CountGenres(input.Genres)
		if err != nil {
			return err
		}
		if genreCount != len(input.Genres) {
			return fmt.Errorf("%w: некоторые жанры не найдены", ErrInvalidGenres)
		}

		now := time.Now()
		eventID, err := store.CreateEvent(EventCreateRecord{
			Title:           input.Title,
			Description:     input.Description,
			GroupID:         input.GroupID,
			EventTypeID:     input.EventTypeID,
			LocationID:      input.LocationID,
			CreatorID:       actorID,
			StartTime:       input.StartTime,
			EndTime:         input.StartTime.Add(time.Duration(input.Duration) * time.Minute),
			Duration:        input.Duration,
			MaxUsers:        input.MaxUsers,
			CurrentUsers:    1,
			ImageURL:        input.ImageURL,
			StatusID:        1,
			Address:         input.Address,
			Country:         input.Country,
			AgeLimitID:      input.AgeLimitID,
			Year:            input.Year,
			Notes:           input.Notes,
			CustomFields:    input.CustomFields,
			GenreIDs:        append([]uint(nil), input.Genres...),
			CreatorJoinedAt: now,
		})
		if err != nil {
			return err
		}

		s.recordCommandAudit(tx.Audit(), EventAuditInput{
			GroupID:    input.GroupID,
			ActorID:    actorID,
			Action:     groupmodels.ActionCreateEvent,
			EntityID:   uintPointer(eventID),
			EntityName: input.Title,
			CreatedAt:  now,
		})

		view, err = store.LoadEventResult(eventID, actorID)
		return err
	})
	if err != nil {
		s.logger.Error("Не удалось создать событие", "error", err)
		return nil, err
	}

	s.logger.Info("Событие успешно создано", "eventID", view.ID, "groupID", input.GroupID)
	return eventCommandViewToDTO(view), nil
}

func (s *eventCommandService) UpdateEvent(ctx context.Context, actorID uint, eventID uint, input UpdateEventInput) (*dto.EventFullDto, error) {
	var view EventCommandView
	err := s.unitOfWork.WithinTransaction(ctx, func(tx EventTransaction) error {
		store := tx.Commands()
		if store == nil {
			return errEventCommandStoreUnavailable
		}

		event, err := store.FindEventForUpdate(eventID)
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

		now := time.Now()
		if now.After(event.StartTime) {
			return ErrEventAlreadyStarted
		}
		if input.MaxUsers != nil && *input.MaxUsers < event.CurrentUsers {
			return fmt.Errorf("%w (%d)", ErrMaxUsersBelowCurrent, event.CurrentUsers)
		}
		if input.AgeLimit != nil {
			exists, err := store.AgeLimitExists(*input.AgeLimit)
			if err != nil {
				return err
			}
			if !exists {
				return ErrAgeLimitNotFound
			}
		}
		if input.Genres != nil {
			if err := validateEventGenreCount(input.Genres); err != nil {
				return err
			}
			genreCount, err := store.CountGenres(input.Genres)
			if err != nil {
				return err
			}
			if genreCount != len(input.Genres) {
				return fmt.Errorf("%w: некоторые жанры не найдены", ErrInvalidGenres)
			}
		}

		update := EventUpdateRecord{
			EventID:       eventID,
			Title:         input.Title,
			Description:   input.Description,
			EventTypeID:   input.EventTypeID,
			LocationID:    input.LocationID,
			ImageURL:      input.ImageURL,
			StartTime:     input.StartTime,
			Duration:      input.Duration,
			MaxUsers:      input.MaxUsers,
			Address:       input.Address,
			Country:       input.Country,
			AgeLimitID:    input.AgeLimit,
			Year:          input.Year,
			Notes:         input.Notes,
			CustomFields:  input.CustomFields,
			ReplaceGenres: input.Genres != nil,
			GenreIDs:      append([]uint(nil), input.Genres...),
		}
		update.EndTime = calculateUpdatedEventEndTime(event, input)
		if err := store.UpdateEvent(update); err != nil {
			return err
		}

		entityName := event.Title
		if input.Title != nil {
			entityName = *input.Title
		}
		s.recordCommandAudit(tx.Audit(), EventAuditInput{
			GroupID:    event.GroupID,
			ActorID:    actorID,
			Action:     groupmodels.ActionUpdateEvent,
			EntityID:   uintPointer(event.ID),
			EntityName: entityName,
			CreatedAt:  now,
		})

		view, err = store.LoadEventResult(eventID, actorID)
		return err
	})
	if err != nil {
		s.logger.Error("Не удалось обновить событие", "eventID", eventID, "error", err)
		return nil, err
	}

	s.logger.Info("Событие успешно обновлено", "eventID", eventID)
	return eventCommandViewToDTO(view), nil
}

func (s *eventCommandService) DeleteEvent(ctx context.Context, actorID uint, eventID uint) (bool, error) {
	err := s.unitOfWork.WithinTransaction(ctx, func(tx EventTransaction) error {
		store := tx.Commands()
		if store == nil {
			return errEventCommandStoreUnavailable
		}

		event, err := store.FindEventForUpdate(eventID)
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
		if err := store.DeleteEventAggregate(eventID); err != nil {
			return err
		}

		s.recordCommandAudit(tx.Audit(), EventAuditInput{
			GroupID:    event.GroupID,
			ActorID:    actorID,
			Action:     groupmodels.ActionDeleteEvent,
			EntityID:   uintPointer(event.ID),
			EntityName: event.Title,
			CreatedAt:  time.Now(),
		})
		return nil
	})
	if err != nil {
		s.logger.Error("Не удалось удалить событие", "eventID", eventID, "error", err)
		return false, err
	}

	s.logger.Info("Событие успешно удалено", "eventID", eventID)
	return true, nil
}

func (s *eventCommandService) recordCommandAudit(audit EventAuditStore, input EventAuditInput) {
	if audit == nil {
		return
	}
	if err := audit.RecordBestEffort(input); err != nil {
		s.logger.Warn("Не удалось записать действие в журнал", "error", err)
	}
}

func validateEventGenreCount(genreIDs []uint) error {
	if len(genreIDs) == 0 || len(genreIDs) > 9 {
		return fmt.Errorf("%w: необходимо от 1 до 9 жанров", ErrInvalidGenres)
	}
	return nil
}

func calculateUpdatedEventEndTime(event EventCommandSnapshot, input UpdateEventInput) *time.Time {
	if input.StartTime == nil && input.Duration == nil {
		return nil
	}

	startTime := event.StartTime
	if input.StartTime != nil {
		startTime = *input.StartTime
	}
	duration := event.Duration
	if input.Duration != nil {
		duration = *input.Duration
	}
	endTime := startTime.Add(time.Duration(duration) * time.Minute)
	return &endTime
}

func uintPointer(value uint) *uint {
	result := value
	return &result
}

func eventCommandViewToDTO(view EventCommandView) *dto.EventFullDto {
	return &dto.EventFullDto{
		ID:           view.ID,
		Title:        view.Title,
		Description:  view.Description,
		ImageURL:     view.ImageURL,
		MaxUsers:     view.MaxUsers,
		CurrentUsers: view.CurrentUsers,
		StartTime:    view.StartTime,
		EndTime:      view.EndTime,
		Duration:     view.Duration,
		Group: dto.EventGroupDto{
			ID:         view.Group.ID,
			Name:       view.Group.Name,
			Image:      view.Group.Image,
			Enterprise: view.Group.Enterprise,
		},
		EventType:    view.EventTypeID,
		LocationType: view.LocationID,
		Status:       view.StatusID,
		Genres:       append([]string(nil), view.Genres...),
		Creator: dto.EventCreatorDto{
			ID:       view.Creator.ID,
			Name:     view.Creator.Name,
			Us:       view.Creator.Us,
			Image:    view.Creator.Image,
			Verified: view.Creator.Verified,
		},
		Address:      view.Address,
		Country:      view.Country,
		AgeLimit:     view.AgeLimit,
		Year:         view.Year,
		Notes:        view.Notes,
		CustomFields: cloneCommandCustomFields(view.CustomFields),
		Subscribed:   view.Subscribed,
		IsCreator:    view.IsCreator,
		CreatedAt:    view.CreatedAt,
		UpdatedAt:    view.UpdatedAt,
	}
}

func cloneCommandCustomFields(source map[string]interface{}) map[string]interface{} {
	if source == nil {
		return nil
	}
	result := make(map[string]interface{}, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
