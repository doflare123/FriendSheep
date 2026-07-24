package references

import (
	"context"
	"fmt"
	"strings"

	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	"friendship/repository"
)

type gormReferenceStore struct {
	repo repository.PostgresRepository
}

func NewGORMReferenceStore(repo repository.PostgresRepository) ReferenceStore {
	return &gormReferenceStore{repo: repo}
}

func (s *gormReferenceStore) LoadReferences(ctx context.Context) (ReferenceSnapshot, error) {
	ctx = normalizeReferenceContext(ctx)

	var categories []models.Category
	if err := s.repo.Model(&models.Category{}).WithContext(ctx).Order("name ASC, id ASC").Find(&categories).Error; err != nil {
		return ReferenceSnapshot{}, fmt.Errorf("ошибка получения категорий: %w", err)
	}

	var locations []eventmodels.EventLocation
	if err := s.repo.Model(&eventmodels.EventLocation{}).WithContext(ctx).Order("name ASC, id ASC").Find(&locations).Error; err != nil {
		return ReferenceSnapshot{}, fmt.Errorf("ошибка получения мест проведения: %w", err)
	}

	var ageLimits []eventmodels.AgeLimit
	if err := s.repo.Model(&eventmodels.AgeLimit{}).WithContext(ctx).Order("id ASC").Find(&ageLimits).Error; err != nil {
		return ReferenceSnapshot{}, fmt.Errorf("ошибка получения возрастных ограничений: %w", err)
	}

	var statuses []eventmodels.Status
	if err := s.repo.
		Model(&eventmodels.Status{}).
		WithContext(ctx).
		Select("MIN(id) AS id, name").
		Group("name").
		Order("MIN(id) ASC").
		Find(&statuses).Error; err != nil {
		return ReferenceSnapshot{}, fmt.Errorf("ошибка получения статусов: %w", err)
	}

	var actionTypes []groupmodels.GroupActionType
	if err := s.repo.Model(&groupmodels.GroupActionType{}).WithContext(ctx).Order("id ASC").Find(&actionTypes).Error; err != nil {
		return ReferenceSnapshot{}, fmt.Errorf("ошибка получения типов действий группы: %w", err)
	}

	categoryItems := referenceItemsFromCategories(categories)
	return ReferenceSnapshot{
		EventTypes:       categoryItems,
		Locations:        referenceItemsFromLocations(locations),
		AgeLimits:        referenceItemsFromAgeLimits(ageLimits),
		Statuses:         referenceItemsFromStatuses(statuses),
		GroupCategories:  append([]ReferenceItem(nil), categoryItems...),
		GroupActionTypes: actionReferenceItemsFromModels(actionTypes),
	}, nil
}

func (s *gormReferenceStore) SearchGenres(ctx context.Context, query GenreSearchQuery) (GenreSearchPage, error) {
	ctx = normalizeReferenceContext(ctx)
	db := s.repo.Model(&eventmodels.Genre{}).WithContext(ctx)

	if query.Query != "" {
		db = db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(query.Query)+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return GenreSearchPage{}, fmt.Errorf("ошибка подсчёта жанров: %w", err)
	}

	var genres []eventmodels.Genre
	offset := (query.Page - 1) * query.Limit
	if err := db.
		Order("name ASC, id ASC").
		Offset(offset).
		Limit(query.Limit).
		Find(&genres).Error; err != nil {
		return GenreSearchPage{}, fmt.Errorf("ошибка получения жанров: %w", err)
	}

	return GenreSearchPage{
		Items: referenceItemsFromGenres(genres),
		Total: total,
	}, nil
}

func normalizeReferenceContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func referenceItemsFromCategories(items []models.Category) []ReferenceItem {
	result := make([]ReferenceItem, 0, len(items))
	for _, item := range items {
		result = append(result, ReferenceItem{ID: item.ID, Name: item.Name})
	}
	return result
}

func referenceItemsFromLocations(items []eventmodels.EventLocation) []ReferenceItem {
	result := make([]ReferenceItem, 0, len(items))
	for _, item := range items {
		result = append(result, ReferenceItem{ID: item.ID, Name: item.Name})
	}
	return result
}

func referenceItemsFromAgeLimits(items []eventmodels.AgeLimit) []ReferenceItem {
	result := make([]ReferenceItem, 0, len(items))
	for _, item := range items {
		result = append(result, ReferenceItem{ID: item.ID, Name: item.Name})
	}
	return result
}

func referenceItemsFromStatuses(items []eventmodels.Status) []ReferenceItem {
	result := make([]ReferenceItem, 0, len(items))
	for _, item := range items {
		result = append(result, ReferenceItem{ID: item.ID, Name: item.Name})
	}
	return result
}

func referenceItemsFromGenres(items []eventmodels.Genre) []ReferenceItem {
	result := make([]ReferenceItem, 0, len(items))
	for _, item := range items {
		result = append(result, ReferenceItem{ID: item.ID, Name: item.Name})
	}
	return result
}

func actionReferenceItemsFromModels(items []groupmodels.GroupActionType) []ActionReferenceItem {
	result := make([]ActionReferenceItem, 0, len(items))
	for _, item := range items {
		result = append(result, ActionReferenceItem{
			ID:   item.ID,
			Code: item.Code,
			Name: item.Name,
		})
	}
	return result
}
