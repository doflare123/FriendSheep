package events

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	"friendship/repository"
	"time"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("пользователь не найден")
)

// Получает полную информацию о событии для администратора
func (s *eventsService) GetEventDetailsForAdmin(actorID uint, eventID uint) (*dto.EventAdminDto, error) {
	var event events.Event

	err := s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Creator").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Preload("Users.User").
		First(&event, eventID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("ошибка получения события: %w", err)
	}

	hasAccess, _, err := s.checkGroupAccess(actorID, event.GroupID, []string{"Админ", "Модератор"})
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	return convertorsdto.ConvertToAdminDto(&event, actorID), nil
}

// Удаляет пользователя из события
func (s *eventsService) KickUserFromEvent(actorID uint, eventID uint, targetUserID uint) (bool, error) {
	var event events.Event
	var eventUser events.EventsUser
	var actor models.User
	var targetUser models.User

	err := s.repo.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&event, eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}
			return fmt.Errorf("ошибка поиска события: %w", err)
		}

		hasAccess, role, err := s.checkGroupAccess(actorID, event.GroupID, []string{"Админ", "Модератор"})
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrPermissionDenied
		}

		if event.CreatorID == targetUserID {
			return ErrCreatorCantLeave
		}

		if actorID == targetUserID {
			return fmt.Errorf("используйте метод покинуть событие для выхода")
		}

		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска актора: %w", err)
		}

		if err := tx.First(&targetUser, targetUserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.Where("event_id = ? AND user_id = ?", eventID, targetUserID).
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

		action := fmt.Sprintf("Исключил пользователя '%s' (@%s) из события '%s' (ID: %d)",
			targetUser.Name, targetUser.Us, event.Title, event.ID)
		if err := s.logGroupAction(tx, event.GroupID, actorID, actor.Name, actor.Us, role, "kick_from_event", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to kick user from event", "eventID", eventID, "userID", targetUserID, "error", err)
		return false, err
	}

	s.logger.Info("User kicked from event", "eventID", eventID, "kickedUserID", targetUserID, "actorID", actorID)
	return true, nil
}

// Создает новое событие
func (s *eventsService) CreateEvent(actorID uint, input CreateEventInput) (*dto.EventFullDto, error) {
	// Проверяем права доступа (admin или operator)
	hasAccess, role, err := s.checkGroupAccess(actorID, input.GroupID, []string{"Админ", "Модератор"})
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	if len(input.Genres) == 0 || len(input.Genres) > 9 {
		return nil, fmt.Errorf("%w: необходимо от 1 до 9 жанров", ErrInvalidGenres)
	}

	var event events.Event
	var actor models.User

	err = s.repo.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		var ageLimit events.AgeLimit
		if err := tx.First(&ageLimit, input.AgeLimitID).Error; err != nil {
			return fmt.Errorf("возрастное ограничение не найдено: %w", err)
		}

		endTime := input.StartTime.Add(time.Duration(input.Duration) * time.Minute)

		event = events.Event{
			Title:           input.Title,
			Description:     input.Description,
			GroupID:         input.GroupID,
			EventTypeID:     input.EventTypeID,
			EventLocationID: input.LocationID,
			CreatorID:       actorID,
			StartTime:       input.StartTime,
			EndTime:         endTime,
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
		}

		if err := tx.Create(&event).Error; err != nil {
			return fmt.Errorf("ошибка создания события: %w", err)
		}

		var existingGenres []events.Genre
		if err := tx.Where("id IN ?", input.Genres).Find(&existingGenres).Error; err != nil {
			return fmt.Errorf("ошибка проверки жанров: %w", err)
		}

		if len(existingGenres) != len(input.Genres) {
			return fmt.Errorf("%w: некоторые жанры не найдены", ErrInvalidGenres)
		}

		for _, genreID := range input.Genres {
			eventGenre := events.EventGenre{
				EventID: event.ID,
				GenreID: genreID,
			}
			if err := tx.Create(&eventGenre).Error; err != nil {
				return fmt.Errorf("ошибка связывания жанра: %w", err)
			}
		}

		creatorParticipant := events.EventsUser{
			EventID:  event.ID,
			UserID:   actorID,
			JoinedAt: time.Now(),
		}
		if err := tx.Create(&creatorParticipant).Error; err != nil {
			return fmt.Errorf("ошибка добавления создателя в участники: %w", err)
		}

		action := fmt.Sprintf("Создал событие '%s' (ID: %d)", event.Title, event.ID)
		if err := s.logGroupAction(tx, input.GroupID, actorID, actor.Name, actor.Us, role, "create_event", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create event", "error", err)
		return nil, err
	}

	if err := s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Creator").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Preload("Users.User").
		First(&event, event.ID).Error; err != nil {
		return nil, fmt.Errorf("ошибка загрузки события: %w", err)
	}

	s.logger.Info("Event created successfully", "eventID", event.ID, "groupID", input.GroupID)

	return convertorsdto.ConvertToFullDto(&event, actorID, false), nil
}

// Обновляет событие
func (s *eventsService) UpdateEvent(actorID uint, eventID uint, input UpdateEventInput) (*dto.EventFullDto, error) {
	var event events.Event
	var actor models.User

	if err := s.repo.First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("ошибка поиска события: %w", err)
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, event.GroupID, []string{"Админ", "Модератор"})
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrPermissionDenied
	}

	if time.Now().After(event.StartTime) {
		return nil, ErrEventAlreadyStarted
	}

	err = s.repo.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		updates := make(map[string]interface{})
		if input.Title != nil {
			updates["title"] = *input.Title
		}
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if input.EventTypeID != nil {
			updates["event_type_id"] = *input.EventTypeID
		}
		if input.LocationID != nil {
			updates["event_location_id"] = *input.LocationID
		}
		if input.ImageURL != nil {
			updates["image_url"] = *input.ImageURL
		}
		if input.Duration != nil {
			updates["duration"] = *input.Duration
			if input.StartTime != nil {
				updates["end_time"] = input.StartTime.Add(time.Duration(*input.Duration) * time.Minute)
			} else {
				updates["end_time"] = event.StartTime.Add(time.Duration(*input.Duration) * time.Minute)
			}
		}
		if input.StartTime != nil {
			updates["start_time"] = *input.StartTime
			if input.Duration != nil {
				updates["end_time"] = input.StartTime.Add(time.Duration(*input.Duration) * time.Minute)
			} else {
				updates["end_time"] = input.StartTime.Add(time.Duration(event.Duration) * time.Minute)
			}
		}
		if input.MaxUsers != nil {
			if *input.MaxUsers < event.CurrentUsers {
				return fmt.Errorf("нельзя установить максимум меньше текущего количества участников (%d)", event.CurrentUsers)
			}
			updates["max_users"] = *input.MaxUsers
		}
		if input.Address != nil {
			updates["address"] = *input.Address
		}
		if input.Country != nil {
			updates["country"] = *input.Country
		}
		if input.AgeLimit != nil {
			var ageLimit events.AgeLimit
			if err := tx.First(&ageLimit, *input.AgeLimit).Error; err != nil {
				return fmt.Errorf("возрастное ограничение не найдено: %w", err)
			}
			updates["age_limit_id"] = *input.AgeLimit
		}
		if input.Year != nil {
			updates["year"] = *input.Year
		}
		if input.Notes != nil {
			updates["notes"] = *input.Notes
		}
		if input.CustomFields != nil {
			updates["custom_fields"] = *input.CustomFields
		}

		if len(updates) > 0 {
			if err := tx.Model(&event).Updates(updates).Error; err != nil {
				return fmt.Errorf("ошибка обновления события: %w", err)
			}
		}

		if input.Genres != nil {
			if len(input.Genres) == 0 || len(input.Genres) > 9 {
				return fmt.Errorf("%w: необходимо от 1 до 9 жанров", ErrInvalidGenres)
			}

			if err := tx.Where("event_id = ?", eventID).Delete(&events.EventGenre{}).Error; err != nil {
				return fmt.Errorf("ошибка очистки жанров: %w", err)
			}

			var existingGenres []events.Genre
			if err := tx.Where("id IN ?", input.Genres).Find(&existingGenres).Error; err != nil {
				return fmt.Errorf("ошибка проверки жанров: %w", err)
			}

			if len(existingGenres) != len(input.Genres) {
				return fmt.Errorf("%w: некоторые жанры не найдены", ErrInvalidGenres)
			}

			for _, genreID := range input.Genres {
				eventGenre := events.EventGenre{
					EventID: eventID,
					GenreID: genreID,
				}
				if err := tx.Create(&eventGenre).Error; err != nil {
					return fmt.Errorf("ошибка связывания жанра: %w", err)
				}
			}
		}

		action := fmt.Sprintf("Обновил событие '%s' (ID: %d)", event.Title, event.ID)
		if err := s.logGroupAction(tx, event.GroupID, actorID, actor.Name, actor.Us, role, "update_event", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to update event", "eventID", eventID, "error", err)
		return nil, err
	}

	if err := s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Creator").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Preload("Users.User").
		First(&event, eventID).Error; err != nil {
		return nil, fmt.Errorf("ошибка загрузки события: %w", err)
	}

	s.logger.Info("Event updated successfully", "eventID", eventID)

	return convertorsdto.ConvertToFullDto(&event, actorID, false), nil
}

// Удаляет событие
func (s *eventsService) DeleteEvent(actorID uint, eventID uint) (bool, error) {
	var event events.Event
	var actor models.User

	if err := s.repo.First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrEventNotFound
		}
		return false, fmt.Errorf("ошибка поиска события: %w", err)
	}

	hasAccess, role, err := s.checkGroupAccess(actorID, event.GroupID, []string{"Админ", "Модератор"})
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrPermissionDenied
	}

	err = s.repo.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.First(&actor, actorID).Error; err != nil {
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		if err := tx.Where("event_id = ?", eventID).Delete(&events.EventGenre{}).Error; err != nil {
			return fmt.Errorf("ошибка удаления связей с жанрами: %w", err)
		}

		if err := tx.Where("event_id = ?", eventID).Delete(&events.EventsUser{}).Error; err != nil {
			return fmt.Errorf("ошибка удаления участников: %w", err)
		}

		if err := tx.Delete(&event).Error; err != nil {
			return fmt.Errorf("ошибка удаления события: %w", err)
		}

		action := fmt.Sprintf("Удалил событие '%s' (ID: %d)", event.Title, event.ID)
		if err := s.logGroupAction(tx, event.GroupID, actorID, actor.Name, actor.Us, role, "delete_event", action); err != nil {
			s.logger.Warn("Failed to log action", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to delete event", "eventID", eventID, "error", err)
		return false, err
	}

	s.logger.Info("Event deleted successfully", "eventID", eventID)
	return true, nil
}
