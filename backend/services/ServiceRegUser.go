package services

import (
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/utils"
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

	rand.Seed(time.Now().UnixNano())
	us := fmt.Sprintf("%s%d", input.Name, rand.Intn(100000)+1)

	hashPass, salt := utils.HashPassword(input.Password)

	user := models.User{
		Name:     input.Name,
		Password: hashPass,
		Salt:     salt,
		Email:    input.Email,
		Us:       us,
	}

	if err := db.GetDB().Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
