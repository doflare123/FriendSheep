package events

import (
	"errors"
	"fmt"

	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	"friendship/models/groups"

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
	ErrAgeLimitNotFound    = errors.New("возрастное ограничение не найдено")
	ErrEventAlreadyStarted = errors.New("событие уже началось")
)

type EventsService interface {
	GetAllGenres() ([]dto.ReferenceItemDto, error)
	GetAllReferences() (*dto.ReferencesDto, error)
}

type eventsRepoPort interface {
	Model(value interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
}

type eventsService struct {
	logger logger.Logger
	repo   eventsRepoPort
}

func NewEventsService(logger logger.Logger, repo eventsRepoPort) EventsService {
	return &eventsService{
		logger: logger,
		repo:   repo,
	}
}

func (s *eventsService) GetAllGenres() ([]dto.ReferenceItemDto, error) {
	var genres []events.Genre

	if err := s.repo.Order("name ASC").Find(&genres).Error; err != nil {
		s.logger.Error("Не удалось получить жанры", "error", err)
		return nil, fmt.Errorf("ошибка получения жанров: %w", err)
	}
	result := convertorsdto.ConvertGenresToReferenceItems(genres)

	return result, nil
}

func (s *eventsService) GetAllReferences() (*dto.ReferencesDto, error) {
	var eventTypes []models.Category
	if err := s.repo.Order("name ASC").Find(&eventTypes).Error; err != nil {
		s.logger.Error("Не удалось получить типы событий", "error", err)
		return nil, fmt.Errorf("ошибка получения типов событий: %w", err)
	}

	var locations []events.EventLocation
	if err := s.repo.Order("name ASC").Find(&locations).Error; err != nil {
		s.logger.Error("Не удалось получить места проведения", "error", err)
		return nil, fmt.Errorf("ошибка получения мест проведения: %w", err)
	}

	var ageLimits []events.AgeLimit
	if err := s.repo.Order("id ASC").Find(&ageLimits).Error; err != nil {
		s.logger.Error("Не удалось получить возрастные ограничения", "error", err)
		return nil, fmt.Errorf("ошибка получения возрастных ограничений: %w", err)
	}

	var statuses []events.Status
	if err := s.repo.
		Model(&events.Status{}).
		Select("MIN(id) AS id, name").
		Group("name").
		Order("MIN(id) ASC").
		Find(&statuses).Error; err != nil {
		s.logger.Error("Не удалось получить статусы", "error", err)
		return nil, fmt.Errorf("ошибка получения статусов: %w", err)
	}

	var genres []events.Genre
	if err := s.repo.Order("name ASC").Find(&genres).Error; err != nil {
		s.logger.Error("Не удалось получить жанры", "error", err)
		return nil, fmt.Errorf("ошибка получения жанров: %w", err)
	}

	var groupCategories []models.Category
	if err := s.repo.Order("name ASC").Find(&groupCategories).Error; err != nil {
		s.logger.Error("Не удалось получить категории групп", "error", err)
		return nil, fmt.Errorf("ошибка получения категорий групп: %w", err)
	}

	var groupActionTypes []groups.GroupActionType
	if err := s.repo.Order("id ASC").Find(&groupActionTypes).Error; err != nil {
		s.logger.Error("Не удалось получить типы действий группы", "error", err)
		return nil, fmt.Errorf("ошибка получения типов действий группы: %w", err)
	}

	references := &dto.ReferencesDto{
		EventTypes:       convertorsdto.ConvertToReferenceItems(eventTypes),
		Locations:        convertorsdto.ConvertLocationsToReferenceItems(locations),
		AgeLimits:        convertorsdto.ConvertAgeLimitsToReferenceItems(ageLimits),
		Statuses:         convertorsdto.ConvertStatusesToReferenceItems(statuses),
		Genres:           convertorsdto.ConvertGenresToReferenceItems(genres),
		GroupCategories:  convertorsdto.ConvertToReferenceItems(groupCategories),
		GroupActionTypes: convertorsdto.ConvertGroupActionTypesToReferenceItems(groupActionTypes),
	}

	return references, nil
}
