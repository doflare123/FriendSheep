package events

import (
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	"friendship/models/groups"
	"friendship/repository"
	"time"

	"gorm.io/gorm"
)

var (
	ErrEventNotFound       = errors.New("событие не найдено")
	ErrPermissionDenied    = errors.New("недостаточно прав")
	ErrNotGroupMember      = errors.New("вы не состоите в группе")
	ErrEventFull           = errors.New("событие заполнено")
	ErrAlreadyJoined       = errors.New("вы уже присоединились к событию")
	ErrNotJoined           = errors.New("вы не присоединялись к событию")
	ErrCreatorCantLeave    = errors.New("создатель не может покинуть событие")
	ErrInvalidGenres       = errors.New("некорректные жанры")
	ErrEventAlreadyStarted = errors.New("событие уже началось")
	ErrNotInGroup          = errors.New("событие не принадлежит группе")
)

type EventsService interface {
	// Управление событиями
	CreateEvent(actorID uint, input CreateEventInput) (*dto.EventFullDto, error)
	UpdateEvent(actorID uint, eventID uint, input UpdateEventInput) (*dto.EventFullDto, error)
	DeleteEvent(actorID uint, eventID uint) (bool, error)

	// Получение информации
	GetGroupEvents(actorID uint, groupID uint) ([]dto.EventShortDto, error)
	GetEventDetails(userID uint, eventID uint) (*dto.EventFullDto, error)
	GetEventDetailsForAdmin(actorID uint, eventID uint) (*dto.EventAdminDto, error)

	// Управление участием
	JoinEvent(userID uint, eventID uint) (bool, error)
	LeaveEvent(userID uint, eventID uint) (bool, error)
	KickUserFromEvent(actorID uint, eventID uint, targetUserID uint) (bool, error)

	// Справочники
	GetAllGenres() ([]GenreDto, error)
}

type GenreDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type eventsService struct {
	logger logger.Logger
	repo   repository.PostgresRepository
}

func NewEventsService(logger logger.Logger, repo repository.PostgresRepository) EventsService {
	return &eventsService{
		logger: logger,
		repo:   repo,
	}
}

// Получает список событий группы
func (s *eventsService) GetGroupEvents(actorID uint, groupID uint) ([]dto.EventShortDto, error) {
	hasAccess, _, err := s.checkGroupAccess(actorID, groupID, []string{"Админ", "Модератор"})
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	var events []events.Event
	err = s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Genres.Genre").
		Where("group_id = ?", groupID).
		Order("start_time DESC").
		Find(&events).Error

	if err != nil {
		s.logger.Error("Failed to fetch group events", "groupID", groupID, "error", err)
		return nil, fmt.Errorf("ошибка получения событий: %w", err)
	}

	return convertorsdto.ConvertManyToShortDto(events), nil
}

// Получает полную информацию о событии
func (s *eventsService) GetEventDetails(userID uint, eventID uint) (*dto.EventFullDto, error) {
	var event events.Event

	err := s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Creator").
		Preload("Genres.Genre").
		Preload("Users.User").
		First(&event, eventID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("ошибка получения события: %w", err)
	}

	var count int64
	s.repo.Model(&groups.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, event.GroupID).
		Count(&count)

	if count == 0 {
		return nil, ErrNotGroupMember
	}

	return convertorsdto.ConvertToFullDto(&event, userID, true), nil
}

// Присоединение к событию
func (s *eventsService) JoinEvent(userID uint, eventID uint) (bool, error) {
	var event events.Event

	err := s.repo.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.Model(&events.Event{}).Where("id = ?", eventID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}
			return fmt.Errorf("ошибка поиска события: %w", err)
		}

		var groupUserCount int64
		if err := tx.Model(&groups.GroupUsers{}).
			Where("user_id = ? AND group_id = ?", userID, event.GroupID).
			Count(&groupUserCount).Error; err != nil {
			return fmt.Errorf("ошибка проверки членства в группе: %w", err)
		}

		if groupUserCount == 0 {
			return ErrNotGroupMember
		}

		var existingCount int64
		if err := tx.Model(&events.EventsUser{}).
			Where("event_id = ? AND user_id = ?", eventID, userID).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("ошибка проверки участия: %w", err)
		}

		if existingCount > 0 {
			return ErrAlreadyJoined
		}

		if event.CurrentUsers >= event.MaxUsers {
			return ErrEventFull
		}

		eventUser := events.EventsUser{
			EventID:  eventID,
			UserID:   userID,
			JoinedAt: time.Now(),
		}

		if err := tx.Create(&eventUser).Error; err != nil {
			return fmt.Errorf("ошибка добавления к событию: %w", err)
		}

		if err := tx.Model(&event).Update("current_users", event.CurrentUsers+1).Error; err != nil {
			return fmt.Errorf("ошибка обновления счетчика: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to join event", "eventID", eventID, "userID", userID, "error", err)
		return false, err
	}

	s.logger.Info("User joined event", "eventID", eventID, "userID", userID)
	return true, nil
}

// Покинуть событие
func (s *eventsService) LeaveEvent(userID uint, eventID uint) (bool, error) {
	var event events.Event
	var eventUser events.EventsUser

	err := s.repo.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&event, eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}
			return fmt.Errorf("ошибка поиска события: %w", err)
		}

		if event.CreatorID == userID {
			return ErrCreatorCantLeave
		}

		if err := tx.Where("event_id = ? AND user_id = ?", eventID, userID).
			First(&eventUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotJoined
			}
			return fmt.Errorf("ошибка поиска участия: %w", err)
		}

		if err := tx.Delete(&eventUser).Error; err != nil {
			return fmt.Errorf("ошибка удаления из события: %w", err)
		}

		if event.CurrentUsers > 0 {
			if err := tx.Model(&event).Update("current_users", event.CurrentUsers-1).Error; err != nil {
				return fmt.Errorf("ошибка обновления счетчика: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to leave event", "eventID", eventID, "userID", userID, "error", err)
		return false, err
	}

	s.logger.Info("User left event", "eventID", eventID, "userID", userID)
	return true, nil
}

// Получает все доступные жанры
func (s *eventsService) GetAllGenres() ([]GenreDto, error) {
	var genres []events.Genre

	if err := s.repo.Order("name ASC").Find(&genres).Error; err != nil {
		s.logger.Error("Failed to fetch genres", "error", err)
		return nil, fmt.Errorf("ошибка получения жанров: %w", err)
	}

	result := make([]GenreDto, 0, len(genres))
	for _, g := range genres {
		result = append(result, GenreDto{
			ID:   g.ID,
			Name: g.Name,
		})
	}

	return result, nil
}

// Вспомогательные функции

// Проверяет права доступа к группе
func (s *eventsService) checkGroupAccess(userID uint, groupID uint, allowedRoles []string) (bool, string, error) {
	var groupUser groups.GroupUsers
	err := s.repo.
		Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&groupUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", ErrNotGroupMember
		}
		return false, "", fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	var role groups.Role_in_group
	if err := s.repo.First(&role, groupUser.RoleInGroupID).Error; err != nil {
		return false, "", fmt.Errorf("ошибка получения роли: %w", err)
	}

	for _, allowedRole := range allowedRoles {
		if role.Name == allowedRole {
			return true, role.Name, nil
		}
	}

	return false, role.Name, nil
}

// Записывает действие в журнал группы
func (s *eventsService) logGroupAction(tx repository.PostgresRepository, groupID uint, userID uint, username, us, role, actionType, description string) error {
	action := groups.GroupActionLog{
		GroupID:     groupID,
		UserID:      userID,
		Username:    username,
		Us:          us,
		Role:        role,
		Action:      actionType,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return tx.Create(&action).Error
}
