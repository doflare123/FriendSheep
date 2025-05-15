package services

import (
	"fmt"
	"friendship/db"
	"friendship/models"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type CreateUserInput struct {
	Name     string `json:"name"     validate:"required"`
	Password string `json:"password" validate:"required,password"`
	Email    string `json:"email"    validate:"required,email"`
}

var validate = binding.Validator.Engine().(*validator.Validate)

func CreateUser(input CreateUserInput) (*models.User, error) {
	// 1) валидация
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	// 2) генерируем Us = Name + случайное число 1-10000
	rand.Seed(time.Now().UnixNano())
	us := fmt.Sprintf("%s%d", input.Name, rand.Intn(10000)+1)

	// 3) создаём модель
	user := models.User{
		Name:     input.Name,
		Password: input.Password,
		Email:    input.Email,
		Us:       us,
	}

	if err := db.GetDB().Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
