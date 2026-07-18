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
	"strings"
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
	ErrAgeLimitNotFound    = errors.New("возрастное ограничение не найдено")
	ErrEventAlreadyStarted = errors.New("событие уже началось")
	ErrNotInGroup          = errors.New("событие не принадлежит группе")
)

type EventsService interface {
	// Получение информации
	SearchEvents(userID uint, input EventSearchInput) (*dto.EventSearchResponse, error)
	GetGroupEvents(actorID uint, groupID uint) ([]dto.EventShortDto, error)
	GetEventDetails(userID uint, eventID uint) (*dto.EventFullDto, error)
	GetEventDetailsForAdmin(actorID uint, eventID uint) (*dto.EventAdminDto, error)

	KickUserFromEvent(actorID uint, eventID uint, targetUserID uint) (bool, error)

	// Справочники
	GetAllGenres() ([]dto.ReferenceItemDto, error)
	GetAllReferences() (*dto.ReferencesDto, error)
}

type GenreDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type EventSearchInput struct {
	Query                string
	GroupID              *uint
	CategoryIDs          []uint
	ExcludeCategoryIDs   []uint
	GenreIDs             []uint
	ExcludeGenreIDs      []uint
	EventTypeIDs         []uint
	ExcludeEventTypeIDs  []uint
	LocationTypes        []string
	ExcludeLocationTypes []string
	City                 string
	DateFrom             *time.Time
	DateTo               *time.Time
	HasFreeSlots         *bool
	OnlySubscriptionNews bool
	Page                 int
	Limit                int
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

// Получает список событий группы
func (s *eventsService) SearchEvents(userID uint, input EventSearchInput) (*dto.EventSearchResponse, error) {
	input = normalizeEventSearchInput(input)

	if input.OnlySubscriptionNews && userID == 0 {
		return emptyEventSearchResponse(input.Page, input.Limit), nil
	}

	query := s.repo.
		Model(&events.Event{}).
		Joins("JOIN groups ON groups.id = events.group_id").
		Joins("LEFT JOIN event_locations ON event_locations.id = events.event_location_id")

	query = applyEventSearchVisibility(query, userID)
	query = applyEventSearchFilters(query, userID, input)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Не удалось посчитать события для поиска", "error", err)
		return nil, fmt.Errorf("ошибка поиска событий: %w", err)
	}

	totalPages := calculateTotalPages(total, input.Limit)
	offset64 := int64(input.Page-1) * int64(input.Limit)
	if offset64 > int64(maxIntValue()) {
		return nil, fmt.Errorf("ошибка поиска событий: слишком большое смещение страницы")
	}
	offset := int(offset64)

	var found []events.Event
	eventQuery := query.
		Preload("Group").
		Preload("EventType").
		Preload("EventLocation").
		Preload("Genres.Genre")
	if userID != 0 {
		eventQuery = eventQuery.Preload("Users", "user_id = ?", userID)
	}

	err := eventQuery.
		Order(eventSearchOrder(input)).
		Limit(input.Limit).
		Offset(offset).
		Find(&found).Error
	if err != nil {
		s.logger.Error("Не удалось выполнить поиск событий", "error", err)
		return nil, fmt.Errorf("ошибка поиска событий: %w", err)
	}

	return &dto.EventSearchResponse{
		Items:       convertorsdto.ConvertManyToSearchItemDtoForUser(found, userID),
		Total:       total,
		Limit:       input.Limit,
		CurrentPage: input.Page,
		TotalPages:  totalPages,
		HasMore:     input.Page < totalPages,
	}, nil
}

func (s *eventsService) GetGroupEvents(actorID uint, groupID uint) ([]dto.EventShortDto, error) {
	privateGroup, err := s.isPrivateGroup(groupID)
	if err != nil {
		return nil, err
	}

	if privateGroup {
		isMember, err := s.isGroupMember(actorID, groupID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrNotGroupMember
		}
	}

	var events []events.Event
	err = s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("AgeLimit").
		Preload("Status").
		Preload("Genres.Genre").
		Preload("Users", "user_id = ?", actorID).
		Where("group_id = ?", groupID).
		Order("start_time DESC").
		Find(&events).Error

	if err != nil {
		s.logger.Error("Не удалось получить события группы", "groupID", groupID, "error", err)
		return nil, fmt.Errorf("ошибка получения событий: %w", err)
	}

	return convertorsdto.ConvertManyToShortDtoForUser(events, actorID), nil
}

// Получает полную информацию о событии
func (s *eventsService) GetEventDetails(userID uint, eventID uint) (*dto.EventFullDto, error) {
	var event events.Event

	err := s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Group").
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

	if event.Group.IsPrivate {
		isMember, err := s.isGroupMember(userID, event.GroupID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrNotGroupMember
		}
	}

	return convertorsdto.ConvertToFullDto(&event, userID, true), nil
}

// Получает все доступные жанры
func (s *eventsService) GetAllGenres() ([]dto.ReferenceItemDto, error) {
	var genres []events.Genre

	if err := s.repo.Order("name ASC").Find(&genres).Error; err != nil {
		s.logger.Error("Не удалось получить жанры", "error", err)
		return nil, fmt.Errorf("ошибка получения жанров: %w", err)
	}
	result := convertorsdto.ConvertGenresToReferenceItems(genres)

	return result, nil
}

// GetAllReferences получает все справочники для событий
func (s *eventsService) GetAllReferences() (*dto.ReferencesDto, error) {
	// Получаем типы событий
	var eventTypes []models.Category
	if err := s.repo.Order("name ASC").Find(&eventTypes).Error; err != nil {
		s.logger.Error("Не удалось получить типы событий", "error", err)
		return nil, fmt.Errorf("ошибка получения типов событий: %w", err)
	}

	// Получаем места проведения
	var locations []events.EventLocation
	if err := s.repo.Order("name ASC").Find(&locations).Error; err != nil {
		s.logger.Error("Не удалось получить места проведения", "error", err)
		return nil, fmt.Errorf("ошибка получения мест проведения: %w", err)
	}

	// Получаем возрастные ограничения
	var ageLimits []events.AgeLimit
	if err := s.repo.Order("id ASC").Find(&ageLimits).Error; err != nil {
		s.logger.Error("Не удалось получить возрастные ограничения", "error", err)
		return nil, fmt.Errorf("ошибка получения возрастных ограничений: %w", err)
	}

	// Получаем статусы
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

	// Получаем категории групп (одно и тоже что и типы, но мб что-то сломается)
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

	// Конвертируем в DTO
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

// Вспомогательные функции

func normalizeEventSearchInput(input EventSearchInput) EventSearchInput {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Page > 10000 {
		input.Page = 10000
	}
	if input.Limit < 1 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	input.Query = strings.TrimSpace(input.Query)
	input.City = strings.TrimSpace(input.City)
	input.LocationTypes = normalizeLocationTypes(input.LocationTypes)
	input.ExcludeLocationTypes = normalizeLocationTypes(input.ExcludeLocationTypes)
	return input
}

func maxIntValue() int {
	return int(^uint(0) >> 1)
}

func emptyEventSearchResponse(page, limit int) *dto.EventSearchResponse {
	return &dto.EventSearchResponse{
		Items:       []dto.EventSearchItemDto{},
		Total:       0,
		Limit:       limit,
		CurrentPage: page,
		TotalPages:  0,
		HasMore:     false,
	}
}

func calculateTotalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

func applyEventSearchVisibility(query *gorm.DB, userID uint) *gorm.DB {
	if userID == 0 {
		return query.Where("groups.is_private = ?", false)
	}

	return query.Where(
		"groups.is_private = ? OR EXISTS (SELECT 1 FROM group_users gu_visibility WHERE gu_visibility.group_id = groups.id AND gu_visibility.user_id = ?)",
		false,
		userID,
	)
}

func applyEventSearchFilters(query *gorm.DB, userID uint, input EventSearchInput) *gorm.DB {
	if input.Query != "" {
		like := "%" + strings.ToLower(input.Query) + "%"
		query = query.Where(
			"LOWER(events.title) LIKE ? OR LOWER(events.description) LIKE ? OR LOWER(groups.name) LIKE ?",
			like,
			like,
			like,
		)
	}

	if input.GroupID != nil {
		query = query.Where("events.group_id = ?", *input.GroupID)
	}
	if len(input.CategoryIDs) > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM group_group_categories ggc WHERE ggc.group_id = groups.id AND ggc.group_category_id IN ?)",
			input.CategoryIDs,
		)
	}
	if len(input.ExcludeCategoryIDs) > 0 {
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM group_group_categories ggc_ex WHERE ggc_ex.group_id = groups.id AND ggc_ex.group_category_id IN ?)",
			input.ExcludeCategoryIDs,
		)
	}
	if len(input.GenreIDs) > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM event_genres eg WHERE eg.event_id = events.id AND eg.genre_id IN ?)",
			input.GenreIDs,
		)
	}
	if len(input.ExcludeGenreIDs) > 0 {
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM event_genres eg_ex WHERE eg_ex.event_id = events.id AND eg_ex.genre_id IN ?)",
			input.ExcludeGenreIDs,
		)
	}
	if len(input.EventTypeIDs) > 0 {
		query = query.Where("events.event_type_id IN ?", input.EventTypeIDs)
	}
	if len(input.ExcludeEventTypeIDs) > 0 {
		query = query.Where("events.event_type_id NOT IN ?", input.ExcludeEventTypeIDs)
	}
	if len(input.LocationTypes) > 0 {
		query = query.Where("LOWER(event_locations.name) IN ?", input.LocationTypes)
	}
	if len(input.ExcludeLocationTypes) > 0 {
		query = query.Where("LOWER(event_locations.name) NOT IN ?", input.ExcludeLocationTypes)
	}
	if input.City != "" {
		query = query.Where("LOWER(groups.city) LIKE ?", "%"+strings.ToLower(input.City)+"%")
	}
	if input.DateFrom != nil {
		query = query.Where("events.start_time >= ?", *input.DateFrom)
	}
	if input.DateTo != nil {
		query = query.Where("events.start_time <= ?", *input.DateTo)
	}
	if input.HasFreeSlots != nil {
		if *input.HasFreeSlots {
			query = query.Where("events.current_users < events.max_users")
		} else {
			query = query.Where("events.current_users >= events.max_users")
		}
	}
	if input.OnlySubscriptionNews {
		query = query.Where(
			"EXISTS (SELECT 1 FROM group_users gu_subscription WHERE gu_subscription.group_id = events.group_id AND gu_subscription.user_id = ?)",
			userID,
		)
		query = query.Where(
			"NOT EXISTS (SELECT 1 FROM events_users eu_subscription WHERE eu_subscription.event_id = events.id AND eu_subscription.user_id = ?)",
			userID,
		)
	}

	return query
}

func eventSearchOrder(input EventSearchInput) string {
	if input.OnlySubscriptionNews {
		return "events.created_at DESC, events.id DESC"
	}
	return "events.start_time ASC, events.id DESC"
}

func normalizeLocationTypes(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "online", "онлайн":
			normalized = append(normalized, "online", "онлайн")
		case "offline", "off-line", "офлайн", "оффлайн":
			normalized = append(normalized, "offline", "off-line", "офлайн", "оффлайн")
		case "":
		default:
			normalized = append(normalized, strings.ToLower(strings.TrimSpace(value)))
		}
	}

	return normalized
}

// Проверяет права доступа к группе
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

func (s *eventsService) isPrivateGroup(groupID uint) (bool, error) {
	var group groups.Group
	err := s.repo.Select("id", "is_private").First(&group, groupID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("ошибка получения группы: %w", err)
	}

	return group.IsPrivate, nil
}

func (s *eventsService) isGroupMember(userID uint, groupID uint) (bool, error) {
	var count int64
	if err := s.repo.Model(&groups.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки членства в группе: %w", err)
	}

	return count > 0, nil
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

// Записывает действие в журнал группы
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
