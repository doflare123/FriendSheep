package events

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/dto"
	convertorsdto "friendship/models/dto/convertorsDto"
	"friendship/models/events"
	groupmodels "friendship/models/groups"

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

	hasAccess, _, err := s.checkGroupAccess(actorID, event.GroupID, groupmodels.CapabilityModerate)
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

	err := s.runInTx(func(tx eventsTxPort) error {
		if err := tx.First(&event, eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}
			return fmt.Errorf("ошибка поиска события: %w", err)
		}

		hasAccess, role, err := s.checkGroupAccess(actorID, event.GroupID, groupmodels.CapabilityModerate)
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

		targetUserIDForLog := targetUserID
		entityID := event.ID
		if err := s.logGroupAction(tx, eventGroupActionLogInput{
			GroupID:      event.GroupID,
			UserID:       actorID,
			Username:     actor.Name,
			Us:           actor.Us,
			Role:         role,
			Action:       groupmodels.ActionKickFromEvent,
			TargetUserID: &targetUserIDForLog,
			EntityID:     &entityID,
			EntityName:   event.Title,
		}); err != nil {
			s.logger.Warn("Не удалось записать действие в журнал", "error", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Не удалось исключить пользователя из события", "eventID", eventID, "userID", targetUserID, "error", err)
		return false, err
	}

	s.logger.Info("Пользователь исключен из события", "eventID", eventID, "kickedUserID", targetUserID, "actorID", actorID)
	return true, nil
}
