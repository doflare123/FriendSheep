package events

import (
	"errors"
	"fmt"
	"time"

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
	ErrNotInGroup          = errors.New("событие не принадлежит группе")
)

type EventsService interface {
	GetEventDetailsForAdmin(actorID uint, eventID uint) (*dto.EventAdminDto, error)
	KickUserFromEvent(actorID uint, eventID uint, targetUserID uint) (bool, error)
	GetAllGenres() ([]dto.ReferenceItemDto, error)
	GetAllReferences() (*dto.ReferencesDto, error)
}

type eventsService struct {
	logger logger.Logger
	repo   eventsRepoPort
	tx     eventsTransactionRunner
}

func NewEventsService(logger logger.Logger, repo eventsRepoPort) EventsService {
	return &eventsService{
		logger: logger,
		repo:   repo,
		tx:     newEventsTransactionRunner(repo),
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

func (s *eventsService) checkGroupAccess(userID uint, groupID uint, required groups.Capability) (bool, string, error) {
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

	if groups.HasCapability(role.Name, required) {
		return true, groups.NormalizeRoleName(role.Name), nil
	}

	return false, groups.NormalizeRoleName(role.Name), nil
}

type eventGroupActionLogInput struct {
	GroupID      uint
	UserID       uint
	Username     string
	Us           string
	Role         string
	Action       string
	Description  string
	TargetUserID *uint
	EntityID     *uint
	EntityName   string
}

func (s *eventsService) logGroupAction(tx eventsTxPort, input eventGroupActionLogInput) error {
	actionTypeID, err := groups.FindGroupActionTypeID(tx, input.Action)
	if err != nil {
		return fmt.Errorf("тип действия группы %q не найден: %w", input.Action, err)
	}

	action := groups.GroupActionLog{
		GroupID:      input.GroupID,
		UserID:       input.UserID,
		Username:     input.Username,
		Us:           input.Us,
		Role:         input.Role,
		ActionTypeID: actionTypeID,
		Description:  input.Description,
		TargetUserID: input.TargetUserID,
		EntityID:     input.EntityID,
		EntityName:   input.EntityName,
		CreatedAt:    time.Now(),
	}

	return tx.Create(&action).Error
}
