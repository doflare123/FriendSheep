package models

import (
	"friendship/models/dto"
	"friendship/repository"
	"time"

	"github.com/go-playground/validator/v10"
)

type User struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"not null" validate:"required"`
	Password     string `gorm:"not null" validate:"required,password"`
	Salt         string `gorm:"not null" validate:"required"`
	Us           string `gorm:"uniqueIndex;not null" validate:"required"`
	Email        string `gorm:"uniqueIndex;not null" validate:"required,email"`
	Image        string `gorm:"default:https://cdn-icons-png.flaticon.com/512/149/149071.png"`
	Enterprise   bool   `gorm:"default:false"`
	VerifiedUser bool   `gorm:"default:false"`
	Role         string `gorm:"default:user"`
	Status       string
	TelegramID   *string `gorm:"default:null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ValidateUser(user *User) error {
	validate := validator.New()
	return validate.Struct(user)
}

func (u *User) FindUserByEmail(email string, rep repository.PostgresRepository) (dto.UserDto, error) {
	var user User
	if err := rep.Where("email = ?", email).First(&user).Error; err != nil {
		return *convertUser(user), err
	}
	return *convertUser(user), nil
}

func (u *User) FindUserByID(id uint, rep repository.PostgresRepository) (dto.UserDto, error) {
	var user User
	if err := rep.Where("id = ?", id).First(&user).Error; err != nil {
		return *convertUser(user), err
	}
	return *convertUser(user), nil
}

func (u *User) FindUserByUs(us string, rep repository.PostgresRepository) (*uint, error) {
	var user User
	if err := rep.Where("us = ?", us).First(&user).Error; err != nil {
		return nil, err
	}
	return &user.ID, nil
}

func convertUser(user User) *dto.UserDto {
	return &dto.UserDto{
		Id:           int64(user.ID),
		Username:     user.Name,
		Us:           user.Us,
		Image:        user.Image,
		DataRegister: user.CreatedAt,
		Enterprise:   user.Enterprise,
		VerifiedUser: user.VerifiedUser,
		Role:         user.Role,
		Status:       user.Status,
		Telegram:     user.TelegramID != nil && *user.TelegramID != "",
	}
}
