package register

import (
	"context"
	"errors"
	"fmt"
	"friendship/models"
	statsusers "friendship/models/stats_users"
	"friendship/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormRegistrationStore struct {
	db repository.PostgresRepository
}

func NewGORMRegistrationStore(db repository.PostgresRepository) RegistrationStore {
	return &gormRegistrationStore{db: db}
}

func (s *gormRegistrationStore) CreateUser(ctx context.Context, input UserBootstrap) (RegisteredUser, error) {
	user := models.User{
		Name:     input.Name,
		Password: input.HashedPassword,
		Email:    input.Email,
		Us:       input.Username,
		Birthday: input.Birthday,
		CityID:   input.CityID,
	}

	err := s.db.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		if input.CityID != nil {
			var city models.City
			if err := tx.Model(&models.City{}).WithContext(ctx).Select("id").First(&city, *input.CityID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrCityNotFound
				}
				return fmt.Errorf("ошибка проверки города: %w", err)
			}
		}

		if err := tx.Model(&models.User{}).WithContext(ctx).Create(&user).Error; err != nil {
			if isRegistrationUniqueViolation(err) {
				return ErrUserAlreadyExists
			}
			return fmt.Errorf("ошибка создания пользователя: %w", err)
		}

		// Выбор плиток создаётся вместе с пользователем: ошибка любого этапа
		// регистрации откатывает и профиль, и его начальное оформление.
		defaultTiles := models.UserStatisticTiles{UserID: user.ID}
		if err := tx.Model(&models.UserStatisticTiles{}).WithContext(ctx).Create(&defaultTiles).Error; err != nil {
			return fmt.Errorf("не удалось создать настройки плиток: %w", err)
		}

		statsUser := statsusers.SessionStats_users{UserID: user.ID}
		if err := tx.Model(&statsusers.SessionStats_users{}).WithContext(ctx).Create(&statsUser).Error; err != nil {
			return fmt.Errorf("не удалось создать статистику пользователя: %w", err)
		}

		defaultDay := uint16(1)
		sideStats := statsusers.SideStats_users{
			UserID:     &user.ID,
			MostPopDay: &defaultDay,
		}
		if err := tx.Model(&statsusers.SideStats_users{}).WithContext(ctx).Create(&sideStats).Error; err != nil {
			return fmt.Errorf("не удалось создать статистику сессии: %w", err)
		}

		return nil
	})
	if err != nil {
		return RegisteredUser{}, err
	}

	return RegisteredUser{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Username: user.Us,
		Image:    user.Image,
	}, nil
}

func (s *gormRegistrationStore) ChangePasswordByEmail(ctx context.Context, email, hashedPassword string) (uint, error) {
	var user models.User
	result := s.db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		WithContext(ctx).
		Model(&user).
		Where("email = ?", email).
		Update("password", hashedPassword)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, ErrUserNotFound
	}

	return user.ID, nil
}

func (s *gormRegistrationStore) FindUserIDByEmail(ctx context.Context, email string) (uint, error) {
	var user models.User
	if err := s.db.Model(&models.User{}).WithContext(ctx).Select("id").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return user.ID, nil
}

func isRegistrationUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
