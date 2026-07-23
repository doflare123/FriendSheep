package events

import (
	"context"
	"errors"
	"fmt"

	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormEventAdminStore struct {
	repo repository.PostgresRepository
}

var _ EventAdminReader = (*gormEventAdminStore)(nil)
var _ EventAdminStore = (*gormEventAdminStore)(nil)

func NewGORMEventAdminReader(repo repository.PostgresRepository) EventAdminReader {
	if repo == nil {
		return nil
	}
	return &gormEventAdminStore{repo: repo}
}

func (s *gormEventAdminStore) InspectAccess(
	ctx context.Context,
	actorID uint,
	eventID uint,
) (EventAdminAccessView, error) {
	var event eventmodels.Event
	err := s.repo.
		Model(&eventmodels.Event{}).
		WithContext(normalizeAdminContext(ctx)).
		Select("id", "group_id").
		First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventAdminAccessView{}, nil
		}
		return EventAdminAccessView{}, fmt.Errorf("ошибка поиска события: %w", err)
	}

	view := EventAdminAccessView{EventFound: true}
	var membership groupmodels.GroupUsers
	err = s.repo.
		Model(&groupmodels.GroupUsers{}).
		WithContext(normalizeAdminContext(ctx)).
		Where("user_id = ? AND group_id = ?", actorID, event.GroupID).
		First(&membership).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return view, nil
		}
		return EventAdminAccessView{}, fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	var role groupmodels.Role_in_group
	if err := s.repo.
		Model(&groupmodels.Role_in_group{}).
		WithContext(normalizeAdminContext(ctx)).
		First(&role, membership.RoleInGroupID).Error; err != nil {
		return EventAdminAccessView{}, fmt.Errorf("ошибка получения роли: %w", err)
	}

	view.ActorIsGroupMember = true
	view.ActorRole = groupmodels.NormalizeRoleName(role.Name)
	return view, nil
}

func (s *gormEventAdminStore) LoadDetails(
	ctx context.Context,
	actorID uint,
	eventID uint,
) (EventAdminDetailsView, error) {
	access, err := s.InspectAccess(ctx, actorID, eventID)
	if err != nil {
		return EventAdminDetailsView{}, err
	}
	if !access.EventFound ||
		!access.ActorIsGroupMember ||
		!groupmodels.HasCapability(access.ActorRole, groupmodels.CapabilityModerate) {
		return EventAdminDetailsView{Access: access}, nil
	}

	var event eventmodels.Event
	err = s.repo.
		Model(&eventmodels.Event{}).
		WithContext(normalizeAdminContext(ctx)).
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Group").
		Preload("Creator").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Preload("Users.User").
		First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventAdminDetailsView{}, nil
		}
		return EventAdminDetailsView{}, fmt.Errorf("ошибка получения события: %w", err)
	}

	participants := make([]EventAdminParticipantView, 0, len(event.Users))
	for _, participant := range event.Users {
		participants = append(participants, EventAdminParticipantView{
			ID:       participant.ID,
			UserID:   participant.UserID,
			Name:     participant.User.Name,
			Username: participant.User.Us,
			Image:    participant.User.Image,
			JoinedAt: participant.JoinedAt,
		})
	}

	return EventAdminDetailsView{
		Found:        true,
		Access:       access,
		Event:        mapFullEvent(event, actorID),
		Participants: participants,
	}, nil
}

func (s *gormEventAdminStore) FindEvent(eventID uint) (EventAdminEventSnapshot, error) {
	var event eventmodels.Event
	err := s.repo.
		Model(&eventmodels.Event{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "group_id", "creator_id", "title").
		First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventAdminEventSnapshot{}, ErrEventNotFound
		}
		return EventAdminEventSnapshot{}, fmt.Errorf("ошибка поиска события: %w", err)
	}

	return EventAdminEventSnapshot{
		ID:        event.ID,
		GroupID:   event.GroupID,
		CreatorID: event.CreatorID,
		Title:     event.Title,
	}, nil
}

func (s *gormEventAdminStore) FindGroupRole(actorID uint, groupID uint) (string, error) {
	var membership groupmodels.GroupUsers
	if err := s.repo.
		Where("user_id = ? AND group_id = ?", actorID, groupID).
		First(&membership).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotGroupMember
		}
		return "", fmt.Errorf("ошибка проверки доступа: %w", err)
	}

	var role groupmodels.Role_in_group
	if err := s.repo.First(&role, membership.RoleInGroupID).Error; err != nil {
		return "", fmt.Errorf("ошибка получения роли: %w", err)
	}
	return groupmodels.NormalizeRoleName(role.Name), nil
}

func (s *gormEventAdminStore) UserExists(userID uint) (bool, error) {
	var count int64
	if err := s.repo.
		Model(&models.User{}).
		Where("id = ?", userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}
	return count > 0, nil
}

func (s *gormEventAdminStore) RemoveParticipant(userID uint, eventID uint) error {
	result := s.repo.
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Delete(&eventmodels.EventsUser{})
	if result.Error != nil {
		return fmt.Errorf("ошибка удаления из события: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotJoined
	}
	return nil
}

func (s *gormEventAdminStore) DecrementParticipants(eventID uint) error {
	result := s.repo.
		Model(&eventmodels.Event{}).
		Where("id = ?", eventID).
		UpdateColumn("current_users", gorm.Expr("CASE WHEN current_users > 0 THEN current_users - 1 ELSE 0 END"))
	if result.Error != nil {
		return fmt.Errorf("ошибка обновления счетчика участников: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrEventNotFound
	}
	return nil
}

func normalizeAdminContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
