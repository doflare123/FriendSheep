package events

import (
	"context"
	"fmt"
	eventmodels "friendship/models/events"
	"friendship/repository"
	"time"
)

type gormPopularEventsStore struct {
	postgresRepo repository.PostgresRepository
}

func NewGORMPopularEventsStore(postgresRepo repository.PostgresRepository) PopularEventsStore {
	return &gormPopularEventsStore{postgresRepo: postgresRepo}
}

func (s *gormPopularEventsStore) ListTopPopularEvents(ctx context.Context, now time.Time, limit int) ([]PopularEventRecord, error) {
	var recruitmentStatus eventmodels.Status
	if err := s.postgresRepo.Model(&eventmodels.Status{}).
		WithContext(ctx).
		Where("name = ?", "Набор").
		First(&recruitmentStatus).Error; err != nil {
		return nil, fmt.Errorf("recruitment status not found: %w", err)
	}

	var eventModels []eventmodels.Event
	if err := s.postgresRepo.Model(&eventmodels.Event{}).
		WithContext(ctx).
		Preload("EventType").
		Preload("EventLocation").
		Preload("AgeLimit").
		Preload("Status").
		Preload("Group").
		Preload("Genres.Genre").
		Joins("JOIN groups ON events.group_id = groups.id").
		Where("groups.is_private = ?", false).
		Where("events.start_time > ?", now).
		Where("events.max_users > 0").
		Where("events.current_users > 1").
		Where("events.status_id = ?", recruitmentStatus.ID).
		Order("(CAST(events.current_users AS FLOAT) / CAST(events.max_users AS FLOAT)) DESC, events.current_users DESC").
		Limit(limit).
		Find(&eventModels).Error; err != nil {
		return nil, fmt.Errorf("query popular events: %w", err)
	}

	records := make([]PopularEventRecord, 0, len(eventModels))
	for i := range eventModels {
		records = append(records, mapPopularEventRecord(eventModels[i]))
	}

	if len(records) == 0 {
		return []PopularEventRecord{}, nil
	}

	type ownerRow struct {
		GroupID   uint
		CreatorID uint
		Email     string
		GroupName string
	}

	groupIDs := make([]uint, 0, len(records))
	for _, record := range records {
		groupIDs = append(groupIDs, record.Group.ID)
	}

	var ownerRows []ownerRow
	if err := s.postgresRepo.Model(&eventmodels.Event{}).
		WithContext(ctx).
		Table("groups").
		Select("groups.id AS group_id, groups.creater_id AS creator_id, groups.name AS group_name, users.email").
		Joins("JOIN users ON users.id = groups.creater_id").
		Where("groups.id IN ?", groupIDs).
		Scan(&ownerRows).Error; err != nil {
		return nil, fmt.Errorf("load popular event owners: %w", err)
	}

	ownersByGroupID := make(map[uint]ownerRow, len(ownerRows))
	for _, row := range ownerRows {
		ownersByGroupID[row.GroupID] = row
	}

	for i := range records {
		if owner, ok := ownersByGroupID[records[i].Group.ID]; ok {
			records[i].OwnerEmail = owner.Email
			records[i].OwnerUserID = owner.CreatorID
			records[i].GroupName = owner.GroupName
		}
	}

	return records, nil
}

func mapPopularEventRecord(eventModel eventmodels.Event) PopularEventRecord {
	genres := make([]string, 0, len(eventModel.Genres))
	for _, genre := range eventModel.Genres {
		genres = append(genres, genre.Genre.Name)
	}

	popularityRate := 0.0
	if eventModel.MaxUsers > 0 {
		popularityRate = float64(eventModel.CurrentUsers) / float64(eventModel.MaxUsers)
	}

	return PopularEventRecord{
		PopularEventView: PopularEventView{
			ID:    eventModel.ID,
			Title: eventModel.Title,
			Group: PopularEventGroupView{
				ID:         eventModel.Group.ID,
				Name:       eventModel.Group.Name,
				Image:      eventModel.Group.Image,
				Enterprise: eventModel.Group.Enterprise,
			},
			Image:        eventModel.ImageURL,
			CurrentUsers: eventModel.CurrentUsers,
			MaxUsers:     eventModel.MaxUsers,
			Duration:     eventModel.Duration,
			StartTime:    eventModel.StartTime,
			EventType:    eventModel.EventType.Name,
			LocationType: eventModel.EventLocation.Name,
			AgeLimit:     eventModel.AgeLimit.Name,
			Status:       eventModel.Status.Name,
			City:         eventModel.Group.City,
			Genres:       genres,
			Subscribed:   false,
		},
		GroupName:      eventModel.Group.Name,
		PopularityRate: popularityRate,
		StartTime:      eventModel.StartTime,
	}
}
