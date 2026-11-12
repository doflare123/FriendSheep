package users

import (
	"context"
	"encoding/json"
	"errors"

	"friendship/models"
	"friendship/repository"
	"friendship/services/references"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormProfileStore struct {
	repo repository.PostgresRepository
}

func NewGORMProfileStore(repo repository.PostgresRepository) ProfileStore {
	return &gormProfileStore{repo: repo}
}

func (s *gormProfileStore) FindUser(ctx context.Context, target Target) (UserRecord, error) {
	var user models.User
	query := s.repo.Model(&models.User{}).WithContext(ctx)
	switch target.Kind {
	case TargetByID:
		query = query.Where("id = ?", target.ID)
	case TargetByUs:
		query = query.Where("us = ?", target.Us)
	default:
		return UserRecord{}, ErrInvalidTarget
	}
	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserRecord{}, ErrUserNotFound
		}
		return UserRecord{}, err
	}
	privacy, err := readPrivacy(ctx, s.repo, user.ID)
	if err != nil {
		return UserRecord{}, err
	}
	return UserRecord{
		ID:           user.ID,
		Name:         user.Name,
		Us:           user.Us,
		Image:        user.Image,
		Status:       user.Status,
		CreatedAt:    user.CreatedAt,
		VerifiedUser: user.VerifiedUser,
		Privacy:      privacy,
	}, nil
}

func (s *gormProfileStore) GetPrivacy(ctx context.Context, userID uint) (Privacy, error) {
	if err := ensureUser(ctx, s.repo, userID); err != nil {
		return Privacy{}, err
	}
	return readPrivacy(ctx, s.repo, userID)
}

func (s *gormProfileStore) PatchPrivacy(ctx context.Context, userID uint, patch PrivacyPatch) (Privacy, error) {
	normalized, err := normalizePatch(patch)
	if err != nil {
		return Privacy{}, err
	}
	var result Privacy
	err = s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		if err := ensureUser(ctx, tx, userID); err != nil {
			return err
		}
		effective := DefaultPrivacy()
		updates := make(map[string]any, 2)
		if normalized.Mode != nil {
			effective.Mode = *normalized.Mode
			updates["mode"] = string(*normalized.Mode)
		}
		if normalized.VisibleStatistics != nil {
			effective.VisibleStatistics = append([]StatisticField(nil), (*normalized.VisibleStatistics)...)
			encoded, err := json.Marshal(keysToStrings(*normalized.VisibleStatistics))
			if err != nil {
				return err
			}
			updates["visible_statistics"] = string(encoded)
		}
		row := models.UserPrivacy{
			UserID:            userID,
			Mode:              string(effective.Mode),
			VisibleStatistics: keysToStrings(effective.VisibleStatistics),
		}
		upsert := tx.Model(&models.UserPrivacy{}).WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(updates),
		}).Create(&row)
		if upsert.Error != nil {
			return upsert.Error
		}
		result, err = readPrivacy(ctx, tx, userID)
		return err
	})
	if err != nil {
		return Privacy{}, err
	}
	return result, nil
}

func (s *gormProfileStore) GetTileSettings(ctx context.Context, userID uint) (TileSettings, error) {
	if err := ensureUser(ctx, s.repo, userID); err != nil {
		return TileSettings{}, err
	}
	return readTileSettings(ctx, s.repo, userID)
}

func (s *gormProfileStore) ReplaceTileSettings(ctx context.Context, userID uint, settings TileSettings) (TileSettings, error) {
	normalized, err := normalizeTileSettings(settings)
	if err != nil {
		return TileSettings{}, err
	}
	var result TileSettings
	err = s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		if err := ensureUser(ctx, tx, userID); err != nil {
			return err
		}
		encoded, err := json.Marshal(keysToStrings(normalized.Tiles))
		if err != nil {
			return err
		}
		row := models.UserStatisticTiles{UserID: userID, SelectedKeys: keysToStrings(normalized.Tiles)}
		// Upsert и чтение результата держатся в одной транзакции, чтобы PUT вернул своё состояние.
		upsert := tx.Model(&models.UserStatisticTiles{}).WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{"selected_keys": string(encoded)}),
		}).Create(&row)
		if upsert.Error != nil {
			return upsert.Error
		}
		result, err = readTileSettings(ctx, tx, userID)
		return err
	})
	if err != nil {
		return TileSettings{}, err
	}
	return result, nil
}

func (s *gormProfileStore) FindEventTypesByNames(ctx context.Context, names []string) ([]references.ReferenceItem, error) {
	if len(names) == 0 {
		return []references.ReferenceItem{}, nil
	}
	var categories []models.Category
	if err := s.repo.Model(&models.Category{}).WithContext(ctx).
		Select("id", "name").
		Where("name IN ?", names).
		Find(&categories).Error; err != nil {
		return nil, err
	}
	result := make([]references.ReferenceItem, len(categories))
	for i, category := range categories {
		result[i] = references.ReferenceItem{ID: category.ID, Name: category.Name}
	}
	return result, nil
}

func ensureUser(ctx context.Context, repo repository.PostgresRepository, userID uint) error {
	var count int64
	if err := repo.Model(&models.User{}).WithContext(ctx).Where("id = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrUserNotFound
	}
	return nil
}

func readPrivacy(ctx context.Context, repo repository.PostgresRepository, userID uint) (Privacy, error) {
	var row models.UserPrivacy
	err := repo.Model(&models.UserPrivacy{}).WithContext(ctx).First(&row, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultPrivacy(), nil
	}
	if err != nil {
		return Privacy{}, err
	}
	visible := make([]StatisticField, len(row.VisibleStatistics))
	for i, field := range row.VisibleStatistics {
		visible[i] = StatisticField(field)
	}
	privacy := Privacy{Mode: PrivacyMode(row.Mode), VisibleStatistics: visible}
	if !validMode(privacy.Mode) {
		return Privacy{}, ErrInvalidPrivacy
	}
	seen := make(map[StatisticField]struct{}, len(visible))
	for _, field := range visible {
		if !validStatisticField(field) {
			return Privacy{}, ErrInvalidPrivacy
		}
		if _, duplicate := seen[field]; duplicate {
			return Privacy{}, ErrInvalidPrivacy
		}
		seen[field] = struct{}{}
	}
	return privacy, nil
}

func readTileSettings(ctx context.Context, repo repository.PostgresRepository, userID uint) (TileSettings, error) {
	var row models.UserStatisticTiles
	err := repo.Model(&models.UserStatisticTiles{}).WithContext(ctx).First(&row, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultTileSettings(), nil
	}
	if err != nil {
		return TileSettings{}, err
	}
	if row.SelectedKeys == nil {
		return TileSettings{}, ErrInvalidTileSettings
	}
	settings := TileSettings{Tiles: make([]TileKey, len(row.SelectedKeys))}
	for i, key := range row.SelectedKeys {
		settings.Tiles[i] = TileKey(key)
	}
	return normalizeTileSettings(settings)
}

func keysToStrings[K ~string](keys []K) []string {
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = string(key)
	}
	return result
}
