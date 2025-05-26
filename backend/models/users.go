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
	DataRegister time.Time `json:"data_register" gorm:"autoCreateTime"`
	Enterprise   bool      `json:"enterprise" gorm:"default:false"`
	VerifiedUser bool      `json:"-" gorm:"default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ValidateUser(user *User) error {
	validate := validator.New()
	return validate.Struct(user)
}
