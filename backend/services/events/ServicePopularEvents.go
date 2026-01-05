package events

import (
	"context"
	"encoding/json"
	"fmt"
	"friendship/config"
	"friendship/email"
	"friendship/logger"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	"friendship/repository"
	"friendship/utils"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	PopularEventsCacheKey = "popular_events:top10"
	CacheExpiration       = 6 * time.Hour
	TopEventsLimit        = 10
)

type PopularEventsService interface {
	GetPopularEvents() (*dto.CachedPopularEvents, error)
	UpdateCache() error
	Start() error
	Stop()
}

type popularEventsService struct {
	logger       logger.Logger
	postgresRepo repository.PostgresRepository
	redisRepo    repository.RedisRepository
	emailManager *email.EmailTemplateManager
	config       *config.Config
	cron         *cron.Cron
	ctx          context.Context
}

// NewPopularEventsService создаёт новый сервис популярных событий
func NewPopularEventsService(
	logger logger.Logger,
	postgresRepo repository.PostgresRepository,
	redisRepo repository.RedisRepository,
	config *config.Config,
) (PopularEventsService, error) {
	emailManager, err := email.NewEmailTemplateManager()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации email менеджера: %w", err)
	}

	return &popularEventsService{
		logger:       logger,
		postgresRepo: postgresRepo,
		redisRepo:    redisRepo,
		emailManager: emailManager,
		config:       config,
		ctx:          context.Background(),
	}, nil
}

// Start запускает cron задачу для обновления кэша
func (s *popularEventsService) Start() error {
	s.cron = cron.New()

	_, err := s.cron.AddFunc("0 */6 * * *", func() {
		s.logger.Info("Обновление кэша популярных событий...")
		if err := s.UpdateCache(); err != nil {
			s.logger.Error("Ошибка обновления кэша популярных событий", "error", err)
		} else {
			s.logger.Info("Кэш популярных событий успешно обновлен")
		}
	})

	if err != nil {
		return fmt.Errorf("ошибка создания cron задачи: %w", err)
	}

	s.cron.Start()
	s.logger.Info("Cron задача для популярных событий запущена")

	go func() {
		s.logger.Info("Первоначальное заполнение кэша популярных событий...")
		if err := s.UpdateCache(); err != nil {
			s.logger.Error("Ошибка первоначального заполнения кэша", "error", err)
		} else {
			s.logger.Info("Кэш популярных событий изначально заполнен")
		}
	}()

	return nil
}

// Stop останавливает cron задачу
func (s *popularEventsService) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		s.logger.Info("Cron задача для популярных событий остановлена")
	}
}

// GetPopularEvents получает популярные события из кэша
func (s *popularEventsService) GetPopularEvents() (*dto.CachedPopularEvents, error) {
	cachedData, err := s.redisRepo.Get(s.ctx, PopularEventsCacheKey)
	if err != nil {
		s.logger.Info("Кэш популярных событий пуст, обновляем...")
		if updateErr := s.UpdateCache(); updateErr != nil {
			return nil, fmt.Errorf("кэш пуст и не удалось обновить: %w", updateErr)
		}

		cachedData, err = s.redisRepo.Get(s.ctx, PopularEventsCacheKey)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения данных из кэша после обновления: %w", err)
		}
	}

	var result dto.CachedPopularEvents
	if err := json.Unmarshal([]byte(cachedData), &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации данных из кэша: %w", err)
	}

	if time.Since(result.UpdatedAt) > CacheExpiration {
		go func() {
			s.logger.Info("Кэш устарел, асинхронное обновление...")
			if err := s.UpdateCache(); err != nil {
				s.logger.Error("Ошибка асинхронного обновления кэша", "error", err)
			}
		}()
	}

	return &result, nil
}

// UpdateCache обновляет кэш популярных событий
func (s *popularEventsService) UpdateCache() error {
	events, newEventIDs, err := s.fetchPopularEventsFromDB()
	if err != nil {
		return fmt.Errorf("ошибка получения данных из БД: %w", err)
	}

	cachedData := dto.CachedPopularEvents{
		Events:    events,
		UpdatedAt: time.Now(),
		Count:     len(events),
	}

	jsonData, err := json.Marshal(cachedData)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	err = s.redisRepo.Set(s.ctx, PopularEventsCacheKey, jsonData, CacheExpiration+time.Hour)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в Redis: %w", err)
	}

	s.logger.Info("Кэш обновлен", "count", len(events))

	if len(newEventIDs) > 0 {
		go s.sendPopularEventNotifications(newEventIDs, events)
	}

	return nil
}

// fetchPopularEventsFromDB получает популярные события из базы данных
func (s *popularEventsService) fetchPopularEventsFromDB() ([]dto.EventShortDto, []uint, error) {
	var eventsData []events.Event

	var recruitmentStatus events.Status
	if err := s.postgresRepo.Where("name = ?", "Набор").First(&recruitmentStatus).Error; err != nil {
		return nil, nil, fmt.Errorf("статус 'Набор' не найден: %w", err)
	}

	query := s.postgresRepo.
		Model(&events.Event{}).
		Preload("EventType").
		Preload("EventLocation").
		Preload("AgeLimit").
		Preload("Status").
		Preload("Group").
		Preload("Genres.Genre").
		Joins("JOIN groups ON events.group_id = groups.id").
		Where("groups.is_private = ?", false).
		Where("events.start_time > ?", time.Now()).
		Where("events.max_users > 0").
		Where("events.current_users > 1").
		Where("events.status_id = ?", recruitmentStatus.ID).
		Order("(CAST(events.current_users AS FLOAT) / CAST(events.max_users AS FLOAT)) DESC, events.current_users DESC").
		Limit(TopEventsLimit)

	if err := query.Find(&eventsData).Error; err != nil {
		return nil, nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	if len(eventsData) == 0 {
		return []dto.EventShortDto{}, []uint{}, nil
	}

	previousEventIDs := s.getPreviousPopularEventIDs()

	result := make([]dto.EventShortDto, 0, len(eventsData))
	newEventIDs := make([]uint, 0)

	for _, event := range eventsData {
		eventDto := convertorsdto.ConvertToShortDto(&event)
		result = append(result, *eventDto)

		if !contains(previousEventIDs, event.ID) {
			newEventIDs = append(newEventIDs, event.ID)
		}
	}

	return result, newEventIDs, nil
}

// getPreviousPopularEventIDs получает ID событий из предыдущего кэша
func (s *popularEventsService) getPreviousPopularEventIDs() []uint {
	cachedData, err := s.redisRepo.Get(s.ctx, PopularEventsCacheKey)
	if err != nil {
		return []uint{}
	}

	var previousCache dto.CachedPopularEvents
	if err := json.Unmarshal([]byte(cachedData), &previousCache); err != nil {
		return []uint{}
	}

	ids := make([]uint, len(previousCache.Events))
	for i, event := range previousCache.Events {
		ids[i] = event.ID
	}
	return ids
}

// sendPopularEventNotifications отправляет уведомления владельцам групп
func (s *popularEventsService) sendPopularEventNotifications(eventIDs []uint, events []dto.EventShortDto) {
	s.logger.Info("Отправка уведомлений о популярных событиях", "count", len(eventIDs))

	eventMap := make(map[uint]dto.EventShortDto)
	for _, event := range events {
		eventMap[event.ID] = event
	}

	for _, eventID := range eventIDs {
		event, exists := eventMap[eventID]
		if !exists {
			continue
		}

		var group struct {
			CreaterID uint
			Email     string
			Name      string
		}

		err := s.postgresRepo.
			Model("groups").
			Select("groups.creater_id, users.email, groups.name").
			Joins("JOIN users ON users.id = groups.creater_id").
			Where("groups.id = ?", event.GroupID).
			Scan(&group).Error

		if err != nil {
			s.logger.Error("Ошибка получения владельца группы", "eventID", eventID, "error", err)
			continue
		}

		emailData := email.EmailData{
			EventName:  event.Title,
			GroupName:  group.Name,
			EventDate:  event.StartTime.Format("02.01.2006 15:04"),
			ActionURL:  fmt.Sprintf("https://friendsheep.ru/events/%d", event.ID),
			ActionText: "Посмотреть событие",
		}

		body, err := s.emailManager.RenderTemplate(email.TemplatePopularEvent, emailData)
		if err != nil {
			s.logger.Error("Ошибка рендеринга email шаблона", "eventID", eventID, "error", err)
			continue
		}

		subject := email.GetSubject(email.TemplatePopularEvent, "")

		if err := utils.SendEmail(group.Email, subject, body, s.config); err != nil {
			s.logger.Error("Ошибка отправки email", "email", group.Email, "error", err)
		} else {
			s.logger.Info("Email отправлен", "email", group.Email, "eventID", eventID)
		}
	}
}

func contains(slice []uint, item uint) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
