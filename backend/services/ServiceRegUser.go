package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/utils"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type CreateUserInput struct {
	Name      string `json:"name"     validate:"required"`
	Password  string `json:"password" validate:"required,password"`
	Email     string `json:"email"    validate:"required,email"`
	SessionID string `json:"session_id" validate:"required"`
}

type VerifySessionInput struct {
	SessionID string `json:"session_id" validate:"required"`
	Type      string `json:"type"       validate:"required"`
	Code      string `json:"code"       validate:"required"`
}

var validate = binding.Validator.Engine().(*validator.Validate)

func CreateUser(input CreateUserInput) (*models.User, error) {
	store := db.NewSessionStore(os.Getenv("REDIS_URI"))
	// 1) валидация
	if err := validate.Struct(input); err != nil {
		return nil, err
	}
	field, err := store.GetSessionFields(input.SessionID, "is_verified")
	if err != nil {
		return nil, err
	}

	isVerifiedStr := field["is_verified"]
	isVerified, err := strconv.ParseBool(isVerifiedStr)
	if err != nil {
		return nil, fmt.Errorf("невалидное значение is_verified: %v", err)
	}

	if !isVerified {
		return nil, errors.New("сессия не подтверждена")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	us := fmt.Sprintf("%s%d", input.Name, r.Intn(100000)+1)

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

	return nil, nil
}

func CreateSessionRegister(email string) (*models.SessionRegResponse, error) {
	Id := utils.GenerateSessioID(12)
	codeSession := utils.GenerationSessionCode(6)

	store := db.NewSessionStore(os.Getenv("REDIS_URI"))

	err := store.CreateSession(Id, codeSession, models.SessionTypeRegister, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	// Отправка письма в отдельной горутине (не блокирует HTTP)
	go func(email, code string) {
		subject := "Код подтверждения"
		body := fmt.Sprintf("Ваш код подтверждения: %s\nОн действителен 5 минут.", code)
		err := utils.SendEmail(email, subject, body)
		if err != nil {
			log.Printf("Ошибка отправки email: %v", err)
		}
	}(email, codeSession)

	return &models.SessionRegResponse{SessionID: Id}, nil
}

func VerifySession(input VerifySessionInput) (bool, error) {
	store := db.NewSessionStore(os.Getenv("REDIS_URI"))

	fields, err := store.GetSessionFields(input.SessionID, "code", "type")
	if err != nil {
		return false, err
	}

	// Проверка: сессия не найдена (нет поля code или type)
	if fields["code"] == "" || fields["type"] == "" {
		return false, fmt.Errorf("session not found or incomplete")
	}

	// Проверка кода и типа
	if fields["code"] != input.Code || fields["type"] != input.Type {
		return false, nil
	}

	// Обновляем статус
	err = store.UpdateSessionField(input.SessionID, "is_verified", "1")
	if err != nil {
		return false, err
	}

	return true, nil
}
