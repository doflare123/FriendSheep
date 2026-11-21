package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"friendship/models"
	"friendship/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormEditStore struct {
	repo repository.PostgresRepository
}

func NewGORMEditStore(repo repository.PostgresRepository) EditStore {
	return &gormEditStore{repo: repo}
}

func (s *gormEditStore) GetOwn(ctx context.Context, userID uint) (EditableProfile, error) {
	var user models.User
	if err := s.repo.Model(&models.User{}).WithContext(ctx).Preload("City").First(&user, userID).Error; err != nil {
		return EditableProfile{}, mapEditUserError(err)
	}
	return editableProfileFromModel(user), nil
}

func (s *gormEditStore) Patch(ctx context.Context, userID uint, patch ProfilePatch) (EditableProfile, error) {
	var result EditableProfile
	err := s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).First(&user, userID).Error; err != nil {
			return mapEditUserError(err)
		}
		if user.Birthday != nil && patch.Birthday.Set {
			if !patch.hasChangesBesidesBirthday() {
				return ErrBirthdayChangeDenied
			}
			patch.Birthday = OptionalDate{}
		}
		if patch.CityID.Set && patch.CityID.Value != nil {
			var city models.City
			if err := tx.Model(&models.City{}).WithContext(ctx).Select("id").First(&city, *patch.CityID.Value).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrCityNotFound
				}
				return fmt.Errorf("проверка города: %w", err)
			}
		}
		updates := make(map[string]any, 5)
		if patch.Name != nil {
			updates["name"] = *patch.Name
		}
		if patch.Us != nil {
			updates["us"] = *patch.Us
		}
		if patch.Status != nil {
			updates["status"] = *patch.Status
		}
		if patch.Birthday.Set {
			if patch.Birthday.Value == nil {
				updates["birthday"] = nil
			} else {
				updates["birthday"] = patch.Birthday.Value.Time
			}
		}
		if patch.CityID.Set {
			updates["city_id"] = patch.CityID.Value
		}
		if err := tx.Model(&models.User{}).WithContext(ctx).Where("id = ?", userID).Updates(updates).Error; err != nil {
			if isUsUniqueViolation(err) {
				return ErrUsTaken
			}
			return fmt.Errorf("обновление профиля: %w", err)
		}
		profile, err := loadEditableProfile(ctx, tx, userID)
		result = profile
		return err
	})
	return result, err
}

func (s *gormEditStore) ReplaceImage(ctx context.Context, userID uint, imageURL, objectKey string) (ImageChange, error) {
	if imageURL == "" || objectKey == "" {
		return ImageChange{}, ErrInvalidImage
	}
	return s.changeImage(ctx, userID, imageURL, objectKey)
}

func (s *gormEditStore) DeleteImage(ctx context.Context, userID uint) (ImageChange, error) {
	return s.changeImage(ctx, userID, models.DefaultUserImage, "")
}

func (s *gormEditStore) changeImage(ctx context.Context, userID uint, imageURL, objectKey string) (ImageChange, error) {
	var result ImageChange
	err := s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Select("id", "image_object_key").First(&user, userID).Error; err != nil {
			return mapEditUserError(err)
		}
		result.PreviousObjectKey = user.ImageObjectKey
		if err := tx.Model(&models.User{}).WithContext(ctx).Where("id = ?", userID).Updates(map[string]any{
			"image": imageURL, "image_object_key": objectKey,
		}).Error; err != nil {
			return fmt.Errorf("обновление изображения профиля: %w", err)
		}
		profile, err := loadEditableProfile(ctx, tx, userID)
		result.Profile = profile
		return err
	})
	return result, err
}

func loadEditableProfile(ctx context.Context, repo repository.PostgresRepository, userID uint) (EditableProfile, error) {
	var user models.User
	if err := repo.Model(&models.User{}).WithContext(ctx).Preload("City").First(&user, userID).Error; err != nil {
		return EditableProfile{}, mapEditUserError(err)
	}
	return editableProfileFromModel(user), nil
}

func editableProfileFromModel(user models.User) EditableProfile {
	profile := EditableProfile{
		ID: user.ID, Name: user.Name, Us: user.Us, Status: user.Status, Image: user.Image,
	}
	if user.Birthday != nil {
		date := Date{Time: *user.Birthday}
		profile.Birthday = &date
	}
	if user.City != nil {
		profile.City = &EditableCity{ID: user.City.ID, Name: user.City.Name, Region: user.City.Region}
	}
	return profile
}

func mapEditUserError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	return err
}

func isUsUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code == "23505" && (postgresError.ConstraintName == "idx_users_us_lower" || strings.Contains(postgresError.ConstraintName, "us"))
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") && (strings.Contains(message, "users.us") || strings.Contains(message, "idx_users_us_lower"))
}
