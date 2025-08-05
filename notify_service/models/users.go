package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type User struct {
	ID           uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"name" gorm:"not null" validate:"required"`
	Password     string    `json:"password" gorm:"not null" validate:"required,password"`
	Salt         string    `json:"-" gorm:"not null" validate:"required"`
	Us           string    `json:"us" gorm:"uniqueIndex;not null" validate:"required"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Image        string    `json:"image" gorm:"default:https://cdn-icons-png.flaticon.com/512/149/149071.png"`
	DataRegister time.Time `json:"data_register" gorm:"autoCreateTime"`
	Enterprise   bool      `json:"enterprise" gorm:"default:false"`
	VerifiedUser bool      `json:"-" gorm:"default:false"`
	Role         string    `json:"-" gorm:"default:user"`
	Status       string    `json:"-"`
	TelegramID   *string   `gorm:"default:null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ValidateUser(user *User) error {
	validate := validator.New()
	return validate.Struct(user)
}
