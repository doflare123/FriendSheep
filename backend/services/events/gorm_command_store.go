package events

import (
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/events"
	groupmodels "friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormEventCommandStore struct {
	repo repository.PostgresRepository
}

func (s *gormEventCommandStore) FindEventForUpdate(eventID uint) (EventCommandSnapshot, error) {
	var event events.Event
	err := s.repo.
		Model(&events.Event{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventCommandSnapshot{}, ErrEventNotFound
		}
		return EventCommandSnapshot{}, fmt.Errorf("ошибка поиска события: %w", err)
	}

	return EventCommandSnapshot{
		ID:           event.ID,
		GroupID:      event.GroupID,
		Title:        event.Title,
		StartTime:    event.StartTime,
		Duration:     event.Duration,
		CurrentUsers: event.CurrentUsers,
	}, nil
}

func (s *gormEventCommandStore) FindGroupRole(actorID uint, groupID uint) (string, error) {
	var membership groupmodels.GroupUsers
	if err := s.repo.Where("user_id = ? AND group_id = ?", actorID, groupID).First(&membership).Error; err != nil {
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

func (s *gormEventCommandStore) EventTypeExists(eventTypeID uint) (bool, error) {
	var category models.Category
	if err := s.repo.First(&category, eventTypeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("ошибка проверки типа события: %w", err)
	}
	return true, nil
}

func (s *gormEventCommandStore) EventLocationExists(locationID uint) (bool, error) {
	var location events.EventLocation
	if err := s.repo.First(&location, locationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("ошибка проверки формата события: %w", err)
	}
	return true, nil
}

func (s *gormEventCommandStore) AgeLimitExists(ageLimitID uint) (bool, error) {
	var ageLimit events.AgeLimit
	if err := s.repo.First(&ageLimit, ageLimitID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("ошибка проверки возрастного ограничения: %w", err)
	}
	return true, nil
}

func (s *gormEventCommandStore) CountGenres(genreIDs []uint) (int, error) {
	if len(genreIDs) == 0 {
		return 0, nil
	}

	var count int64
	if err := s.repo.Model(&events.Genre{}).Where("id IN ?", genreIDs).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("ошибка проверки жанров: %w", err)
	}
	return int(count), nil
}

func (s *gormEventCommandStore) CreateEvent(record EventCreateRecord) (uint, error) {
	event := events.Event{
		Title:           record.Title,
		Description:     record.Description,
		GroupID:         record.GroupID,
		EventTypeID:     record.EventTypeID,
		EventLocationID: record.LocationID,
		CreatorID:       record.CreatorID,
		StartTime:       record.StartTime,
		EndTime:         record.EndTime,
		Duration:        record.Duration,
		MaxUsers:        record.MaxUsers,
		CurrentUsers:    record.CurrentUsers,
		ImageURL:        record.ImageURL,
		StatusID:        record.StatusID,
		Address:         record.Address,
		Country:         record.Country,
		AgeLimitID:      record.AgeLimitID,
		Year:            record.Year,
		Notes:           record.Notes,
		CustomFields:    events.CustomFields(record.CustomFields),
	}
	if err := s.repo.Create(&event).Error; err != nil {
		return 0, fmt.Errorf("ошибка создания события: %w", err)
	}

	for _, genreID := range record.GenreIDs {
		relation := events.EventGenre{EventID: event.ID, GenreID: genreID}
		if err := s.repo.Create(&relation).Error; err != nil {
			return 0, fmt.Errorf("ошибка связывания жанра: %w", err)
		}
	}

	participant := events.EventsUser{
		EventID:  event.ID,
		UserID:   record.CreatorID,
		JoinedAt: record.CreatorJoinedAt,
	}
	if err := s.repo.Create(&participant).Error; err != nil {
		return 0, fmt.Errorf("ошибка добавления создателя в участники: %w", err)
	}

	return event.ID, nil
}

func (s *gormEventCommandStore) UpdateEvent(record EventUpdateRecord) error {
	updates := make(map[string]interface{})
	if record.Title != nil {
		updates["title"] = *record.Title
	}
	if record.Description != nil {
		updates["description"] = *record.Description
	}
	if record.EventTypeID != nil {
		updates["event_type_id"] = *record.EventTypeID
	}
	if record.LocationID != nil {
		updates["event_location_id"] = *record.LocationID
	}
	if record.ImageURL != nil {
		updates["image_url"] = *record.ImageURL
	}
	if record.StartTime != nil {
		updates["start_time"] = *record.StartTime
	}
	if record.EndTime != nil {
		updates["end_time"] = *record.EndTime
	}
	if record.Duration != nil {
		updates["duration"] = *record.Duration
	}
	if record.MaxUsers != nil {
		updates["max_users"] = *record.MaxUsers
	}
	if record.Address != nil {
		updates["address"] = *record.Address
	}
	if record.Country != nil {
		updates["country"] = *record.Country
	}
	if record.AgeLimitID != nil {
		updates["age_limit_id"] = *record.AgeLimitID
	}
	if record.Year != nil {
		updates["year"] = *record.Year
	}
	if record.Notes != nil {
		updates["notes"] = *record.Notes
	}
	if record.CustomFields != nil {
		updates["custom_fields"] = events.CustomFields(*record.CustomFields)
	}

	if len(updates) > 0 {
		if err := s.repo.Model(&events.Event{ID: record.EventID}).Updates(updates).Error; err != nil {
			return fmt.Errorf("ошибка обновления события: %w", err)
		}
	}

	if record.ReplaceGenres {
		if err := s.repo.Where("event_id = ?", record.EventID).Delete(&events.EventGenre{}).Error; err != nil {
			return fmt.Errorf("ошибка очистки жанров: %w", err)
		}
		for _, genreID := range record.GenreIDs {
			relation := events.EventGenre{EventID: record.EventID, GenreID: genreID}
			if err := s.repo.Create(&relation).Error; err != nil {
				return fmt.Errorf("ошибка связывания жанра: %w", err)
			}
		}
	}

	return nil
}

func (s *gormEventCommandStore) DeleteEventAggregate(eventID uint) error {
	if err := s.repo.Where("event_id = ?", eventID).Delete(&events.EventGenre{}).Error; err != nil {
		return fmt.Errorf("ошибка удаления связей с жанрами: %w", err)
	}
	if err := s.repo.Where("event_id = ?", eventID).Delete(&events.EventsUser{}).Error; err != nil {
		return fmt.Errorf("ошибка удаления участников: %w", err)
	}
	if err := s.repo.Delete(&events.Event{ID: eventID}).Error; err != nil {
		return fmt.Errorf("ошибка удаления события: %w", err)
	}
	return nil
}

func (s *gormEventCommandStore) LoadEventResult(eventID uint, actorID uint) (EventCommandView, error) {
	var event events.Event
	err := s.repo.
		Preload("EventType").
		Preload("EventLocation").
		Preload("Status").
		Preload("Group").
		Preload("Creator").
		Preload("AgeLimit").
		Preload("Genres.Genre").
		Preload("Users", "user_id = ?", actorID).
		First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventCommandView{}, ErrEventNotFound
		}
		return EventCommandView{}, fmt.Errorf("ошибка загрузки события: %w", err)
	}

	genres := make([]string, 0, len(event.Genres))
	for _, relation := range event.Genres {
		genres = append(genres, relation.Genre.Name)
	}
	subscribed := false
	for _, participant := range event.Users {
		if participant.UserID == actorID {
			subscribed = true
			break
		}
	}

	return EventCommandView{
		ID:           event.ID,
		Title:        event.Title,
		Description:  event.Description,
		ImageURL:     event.ImageURL,
		MaxUsers:     event.MaxUsers,
		CurrentUsers: event.CurrentUsers,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Duration:     event.Duration,
		Group: EventCommandGroupView{
			ID:         event.Group.ID,
			Name:       event.Group.Name,
			Image:      event.Group.Image,
			Enterprise: event.Group.Enterprise,
		},
		EventTypeID: event.EventType.ID,
		LocationID:  event.EventLocation.ID,
		StatusID:    event.Status.ID,
		Genres:      genres,
		Creator: EventCommandCreatorView{
			ID:       event.Creator.ID,
			Name:     event.Creator.Name,
			Us:       event.Creator.Us,
			Image:    event.Creator.Image,
			Verified: event.Creator.VerifiedUser,
		},
		Address:      event.Address,
		Country:      event.Country,
		AgeLimit:     event.AgeLimit.Name,
		Year:         event.Year,
		Notes:        event.Notes,
		CustomFields: map[string]interface{}(event.CustomFields),
		Subscribed:   subscribed,
		IsCreator:    event.CreatorID == actorID,
		CreatedAt:    event.CreatedAt,
		UpdatedAt:    event.UpdatedAt,
	}, nil
}
