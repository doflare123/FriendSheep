package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	statsusers "friendship/models/stats_users"
	"friendship/utils"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateUserInput struct {
	Name      string `json:"name"     binding:"required,username"`
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

func generateUsername(name string) string {
	username := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "_"))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s%d", username, r.Intn(100000)+1)
}

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

	us := generateUsername(input.Name)

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

	defaultTiles := statsusers.SettingTile{
		UserID:      user.ID,
		Count_films: true,
		Count_games: true,
		Count_table: true,
		Count_other: false,
		Count_all:   true,
		Spent_time:  false,
	}

	if err := db.GetDB().Create(&defaultTiles).Error; err != nil {
		return nil, fmt.Errorf("не удалось создать настройки тайлов: %v", err)
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

		body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f6f9;
				margin: 0;
				padding: 0;
			}
			.container {
				max-width: 480px;
				margin: 30px auto;
				background: #fff;
				border-radius: 12px;
				padding: 24px;
				box-shadow: 0 4px 12px rgba(0,0,0,0.1);
			}
			h2 {
				color: #333;
				text-align: center;
			}
			p {
				font-size: 15px;
				color: #555;
				line-height: 1.6;
			}
			.code {
				display: block;
				text-align: center;
				font-size: 24px;
				font-weight: bold;
				margin: 20px 0;
				padding: 12px;
				background: #f0f4ff;
				border: 1px dashed #4a6cf7;
				border-radius: 8px;
				color: #4a6cf7;
				cursor: pointer;
				user-select: all;
			}
			.footer {
				font-size: 12px;
				text-align: center;
				color: #aaa;
				margin-top: 16px;
			}
		</style>
		</head>
		<body>
			<div class="container">
				<h2>Подтверждение регистрации</h2>
				<p>Здравствуйте! 👋</p>
				<p>Чтобы завершить регистрацию, введите следующий код подтверждения:</p>
				<div class="code">%s</div>
				<p>Код действителен <b>10 минут</b>. Если вы не запрашивали регистрацию, просто игнорируйте это письмо.</p>
				<div class="footer">© %d Ваш сервис</div>
			</div>
		</body>
		</html>
		`, code, time.Now().Year())

		if err := utils.SendEmail(email, subject, body); err != nil {
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

func ChangePassword(email, newPassword string) error {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	hashPass, salt := utils.HashPassword(newPassword)

	user.Password = hashPass
	user.Salt = salt

	if err := db.GetDB().Save(&user).Error; err != nil {
		return err
	}

	return nil
}
