package services

import (
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/utils"
	"log"
	"os"
	"strconv"
	"time"
)

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmResetPasswordInput struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,password"`
}

func CreateSessionReset(email string) (*models.SessionRegResponse, error) {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("пользователь с таким email не найден")
	}

	sessionID := utils.GenerateSessioID(12)
	code := utils.GenerationSessionCode(6)

	store := db.NewSessionStore(os.Getenv("REDIS_URI"))
	err := store.CreateSession(sessionID, code, models.SessionTypeResetPassword, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	go func(email, code string) {
		subject := "Сброс пароля"

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
            <h2>Сброс пароля</h2>
            <p>Здравствуйте! 👋</p>
            <p>Мы получили запрос на сброс пароля. Используйте следующий код для подтверждения:</p>
            <div class="code">%s</div>
            <p>Код действителен <b>10 минут</b>. Если вы не запрашивали сброс, просто игнорируйте это письмо.</p>
            <div class="footer">© %d Ваш сервис</div>
        </div>
    </body>
    </html>
    `, code, time.Now().Year())

		if err := utils.SendEmail(email, subject, body); err != nil {
			log.Printf("Ошибка отправки email: %v", err)
		}
	}(email, code)

	return &models.SessionRegResponse{SessionID: sessionID}, nil
}

func ResetPassword(input ConfirmResetPasswordInput) error {
	store := db.NewSessionStore(os.Getenv("REDIS_URI"))

	fields, err := store.GetSessionFields(input.SessionID, "code", "type", "attempts")
	if err != nil {
		return err
	}

	if fields["type"] != string(models.SessionTypeResetPassword) {
		return fmt.Errorf("неверный тип сессии")
	}

	if fields["code"] != input.Code {
		attempts, _ := strconv.Atoi(fields["attempts"])
		attempts++
		if attempts >= 3 {
			_ = store.DeleteSession(input.SessionID)
			return fmt.Errorf("превышено количество попыток, сессия удалена")
		}
		_ = store.UpdateSessionField(input.SessionID, "attempts", attempts)
		return fmt.Errorf("неверный код")
	}

	var user models.User
	if err := db.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		return fmt.Errorf("пользователь не найден")
	}

	hashPass, salt := utils.HashPassword(input.Password)
	user.Password = hashPass
	user.Salt = salt

	if err := db.GetDB().Save(&user).Error; err != nil {
		return err
	}

	return store.DeleteSession(input.SessionID)
}
