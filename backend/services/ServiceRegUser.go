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
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateUserInput struct {
	Name      string `json:"name"     binding:"required"`
	Password  string `json:"password" binding:"required,password"`
	Email     string `json:"email"    binding:"required,email"`
	SessionID string `json:"session_id" binding:"required"`
}

type VerifySessionInput struct {
	SessionID string `json:"session_id" validate:"required"`
	Type      string `json:"type"       validate:"required"`
	Code      string `json:"code"       validate:"required"`
}

var validate = binding.Validator.Engine().(*validator.Validate)

func CreateUser(input CreateUserInput) (*models.User, error) {
	store := db.NewSessionStore(os.Getenv("REDIS_URI"))

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

	err = db.GetDB().Create(&user).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("пользователь с таким email уже существует")
		}
		return nil, err
	}

	if err := store.DeleteSession(input.SessionID); err != nil {
		fmt.Printf("Не удалось удалить сессию %s: %v\n", input.SessionID, err)
	}

	return nil, nil
}

func CreateSessionRegister(email string) (*models.SessionRegResponse, error) {
	Id := utils.GenerateSessioID(12)
	codeSession := utils.GenerationSessionCode(6)

	store := db.NewSessionStore(os.Getenv("REDIS_URI"))

	err := store.CreateSession(Id, codeSession, models.SessionTypeRegister, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	go func(email, code string) {
		subject := "Код подтверждения"
		body := fmt.Sprintf("Ваш код подтверждения: %s\nОн действителен 10 минут.", code)
		err := utils.SendEmail(email, subject, body)
		if err != nil {
			log.Printf("Ошибка отправки email: %v", err)
		}
	}(email, codeSession)

	return &models.SessionRegResponse{SessionID: Id}, nil
}

func VerifySession(input VerifySessionInput) (bool, error) {
	store := db.NewSessionStore(os.Getenv("REDIS_URI"))

	fields, err := store.GetSessionFields(input.SessionID, "code", "type", "attempts")
	if err != nil {
		return false, err
	}

	if fields["code"] == "" || fields["type"] == "" {
		return false, fmt.Errorf("сессия не найдена или удалена")
	}

	if fields["type"] != input.Type {
		return false, nil
	}

	attemptsStr := fields["attempts"]
	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		attempts = 0
	}

	if fields["code"] != input.Code {
		attempts++
		if attempts >= 3 {
			err := store.DeleteSession(input.SessionID)
			if err != nil {
				return false, err
			}
			return false, fmt.Errorf("превышено количество попыток ввода кода, сессия удалена")
		}

		err = store.UpdateSessionField(input.SessionID, "attempts", attempts)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	err = store.UpdateSessionField(input.SessionID, "is_verified", "1")
	if err != nil {
		return false, err
	}

	_ = store.UpdateSessionField(input.SessionID, "attempts", 0)

	return true, nil
}
