package register

import (
	"context"
	"errors"
	"fmt"
	"friendship/config"
	"friendship/email"
	"friendship/logger"
	"friendship/models"
	"friendship/models/dto"
	statsusers "friendship/models/stats_users"
	"friendship/repository"
	session "friendship/sessions"
	"friendship/utils"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type CreateUserInput struct {
	Name      string `json:"name"       binding:"required,username,min=5,max=40"`
	Password  string `json:"password"   binding:"required,password"`
	Email     string `json:"email"      binding:"required,email"`
	SessionID string `json:"session_id" binding:"required"`
}

type VerifySessionInput struct {
	SessionID string `json:"session_id" binding:"required"`
	Type      string `json:"type"       binding:"required"`
	Code      string `json:"code"       binding:"required"`
}

type ChangePasswordInput struct {
	NewPassword string `json:"new_password" binding:"required,password"`
	Email       string `json:"email" binding:"required,email"`
	SessionID   string `json:"session_id" binding:"required"`
}

var (
	ErrSessionNotVerified  = errors.New("сессия не подтверждена")
	ErrSessionNotFound     = errors.New("сессия не найдена или удалена")
	ErrSessionTypeMismatch = errors.New("тип сессии не совпадает")
	ErrInvalidCode         = errors.New("неверный код")
	ErrTooManyAttempts     = errors.New("превышено количество попыток ввода кода, сессия удалена")
	ErrUserAlreadyExists   = errors.New("пользователь с таким email уже существует")
	ErrUserNotFound        = errors.New("пользователь не найден")
)

type RegService interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*dto.AuthResponse, error)
	CreateSessionRegister(ctx context.Context, email string, type_ses string) (*models.SessionRegResponse, error)
	VerifySession(ctx context.Context, input VerifySessionInput) (bool, error)
	ChangePassword(ctx context.Context, input ChangePasswordInput) error
}

type regService struct {
	logger       logger.Logger
	redis        session.SessionStore
	cfg          *config.Config
	postgres     repository.PostgresRepository
	jwtService   *utils.JWTUtils
	emailManager *email.EmailTemplateManager
}

func NewRegisterSrv(
	logger logger.Logger,
	redis session.SessionStore,
	postgres repository.PostgresRepository,
	cfg *config.Config,
	jwtService *utils.JWTUtils,
) (RegService, error) {
	emailManager, err := email.NewEmailTemplateManager()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации email менеджера: %w", err)
	}

	return &regService{
		logger:       logger,
		redis:        redis,
		cfg:          cfg,
		postgres:     postgres,
		jwtService:   jwtService,
		emailManager: emailManager,
	}, nil
}

func (s *regService) CreateUser(ctx context.Context, input CreateUserInput) (*dto.AuthResponse, error) {
	sess, err := s.redis.GetSession(ctx, input.SessionID)
	if err != nil {
		s.logger.Error("Failed to get session", "sessionID", input.SessionID, "error", err)
		return nil, ErrSessionNotFound
	}

	if !sess.IsVerified {
		s.logger.Warn("Attempt to create user with unverified session", "sessionID", input.SessionID)
		return nil, ErrSessionNotVerified
	}

	us := generateUsername()

	hashPass, err := utils.HashPassword(input.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	user := models.User{
		Name:     input.Name,
		Password: hashPass,
		Email:    input.Email,
		Us:       us,
	}

	err = s.postgres.Transaction(func(tx repository.PostgresRepository) error {
		if err := tx.Create(&user).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrUserAlreadyExists
			}
			return fmt.Errorf("ошибка создания пользователя: %w", err)
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

		if err := tx.Create(&defaultTiles).Error; err != nil {
			return fmt.Errorf("не удалось создать настройки тайлов: %w", err)
		}

		statsUser := statsusers.SessionStats_users{
			UserID: user.ID,
		}

		if err := tx.Create(&statsUser).Error; err != nil {
			return fmt.Errorf("не удалось создать статистику пользователя: %w", err)
		}

		defaultDay := uint16(1)
		sideInf := statsusers.SideStats_users{
			UserID:     &user.ID,
			MostPopDay: &defaultDay,
		}

		if err := tx.Create(&sideInf).Error; err != nil {
			return fmt.Errorf("не удалось создать статистику сессии: %w", err)
		}

		go s.sendWelcomeEmail(user.Email, user.Name)

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, err
	}

	if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
		s.logger.Warn("Failed to delete session after user creation", "sessionID", input.SessionID, "error", err)
	}

	s.logger.Info("User created successfully", "userID", user.ID, "email", user.Email)

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Name, user.Us, user.Image)
	if err != nil {
		s.logger.Error("Failed to generate tokens after registration", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("пользователь создан, но не удалось сгенерировать токены: %w", err)
	}

	authResponse := &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		AdminGroups:  []dto.AdminGroupResponse{},
	}

	s.logger.Info("User auto-logged in after registration", "userID", user.ID)

	return authResponse, nil
}

func (s *regService) CreateSessionRegister(ctx context.Context, email, type_ses string) (*models.SessionRegResponse, error) {
	sessionID := utils.GenerateSessioID(12)
	code := utils.GenerationSessionCode(6)

	if type_ses == "register" {
		err := s.redis.CreateSession(
			ctx,
			sessionID,
			code,
			models.SessionTypeRegister,
			10*time.Minute,
			map[string]string{
				"email": email,
			},
		)
		if err != nil {
			s.logger.Error("Failed to create session", "email", email, "error", err)
			return nil, fmt.Errorf("не удалось создать сессию: %w", err)
		}
	} else {
		err := s.redis.CreateSession(
			ctx,
			sessionID,
			code,
			models.SessionTypeResetPassword,
			10*time.Minute,
			map[string]string{
				"email": email,
			},
		)
		if err != nil {
			s.logger.Error("Failed to create session", "email", email, "error", err)
			return nil, fmt.Errorf("не удалось создать сессию: %w", err)
		}
	}

	go s.sendVerificationEmail(email, code, type_ses)

	s.logger.Info("Registration session created", "sessionID", sessionID, "email", email)
	return &models.SessionRegResponse{SessionID: sessionID}, nil
}

func (s *regService) sendVerificationEmail(userEmail, code, actionType string) error {
	messageMap := map[string]string{
		"reset_password": "Чтобы завершить смену пароля, введите следующий код подтверждения:",
		"register":       "Чтобы завершить регистрацию, введите следующий код подтверждения:",
	}

	message, exists := messageMap[actionType]
	if !exists {
		message = "Введите следующий код подтверждения:"
	}

	// Определяем тип шаблона
	var templateType email.TemplateType
	if actionType == "reset_password" {
		templateType = email.TemplateResetPassword
	} else {
		templateType = email.TemplateVerificationCode
	}

	// Формируем данные для email
	emailData := email.EmailData{
		Code:    code,
		Message: message,
	}

	// Рендерим шаблон
	body, err := s.emailManager.RenderTemplate(templateType, emailData)
	if err != nil {
		return fmt.Errorf("ошибка рендеринга email шаблона: %w", err)
	}

	subject := email.GetSubject(templateType, "")

	// Отправляем email
	if err := utils.SendEmail(userEmail, subject, body, s.cfg); err != nil {
		s.logger.Error("Failed to send verification email", "email", userEmail, "error", err)
		return err
	}

	s.logger.Info("Verification email sent", "email", userEmail, "type", actionType)
	return nil
}

// sendWelcomeEmail отправляет приветственное письмо
func (s *regService) sendWelcomeEmail(userEmail, userName string) error {
	emailData := email.EmailData{
		UserName:   userName,
		ActionURL:  "https://friendsheep.ru/",
		ActionText: "Перейти на главную",
	}

	body, err := s.emailManager.RenderTemplate(email.TemplateWelcome, emailData)
	if err != nil {
		return fmt.Errorf("ошибка рендеринга welcome шаблона: %w", err)
	}

	subject := email.GetSubject(email.TemplateWelcome, "")

	if err := utils.SendEmail(userEmail, subject, body, s.cfg); err != nil {
		s.logger.Error("Failed to send welcome email", "email", userEmail, "error", err)
		return err
	}

	s.logger.Info("Welcome email sent", "email", userEmail)
	return nil
}

func (s *regService) VerifySession(ctx context.Context, input VerifySessionInput) (bool, error) {
	sess, err := s.redis.GetSession(ctx, input.SessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			s.logger.Warn("Session not found", "sessionID", input.SessionID)
			return false, ErrSessionNotFound
		}
		s.logger.Error("Failed to get session", "sessionID", input.SessionID, "error", err)
		return false, err
	}

	if string(sess.Type) != input.Type {
		s.logger.Warn("Session type mismatch", "sessionID", input.SessionID, "expected", input.Type, "actual", sess.Type)
		return false, ErrSessionTypeMismatch
	}

	if sess.Attempts >= 3 {
		s.logger.Warn("Too many attempts, deleting session", "sessionID", input.SessionID)
		if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
			s.logger.Error("Failed to delete session after too many attempts", "sessionID", input.SessionID, "error", err)
		}
		return false, ErrTooManyAttempts
	}

	if sess.Code != input.Code {
		attempts, err := s.redis.IncrementAttempts(ctx, input.SessionID)
		if err != nil {
			s.logger.Error("Failed to increment attempts", "sessionID", input.SessionID, "error", err)
			return false, err
		}

		s.logger.Warn("Invalid code", "sessionID", input.SessionID, "attempts", attempts)

		if attempts >= 3 {
			if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
				s.logger.Error("Failed to delete session after last attempt", "sessionID", input.SessionID, "error", err)
			}
			return false, ErrTooManyAttempts
		}

		return false, ErrInvalidCode
	}

	if err := s.redis.MarkAsVerified(ctx, input.SessionID); err != nil {
		s.logger.Error("Failed to mark session as verified", "sessionID", input.SessionID, "error", err)
		return false, err
	}

	s.logger.Info("Session verified successfully", "sessionID", input.SessionID)
	return true, nil
}

func (s *regService) ChangePassword(ctx context.Context, input ChangePasswordInput) error {

	sess, err := s.redis.GetSession(ctx, input.SessionID)
	if err != nil {
		s.logger.Error("Failed to get session", "sessionID", input.SessionID, "error", err)
		return ErrSessionNotFound
	}

	if !sess.IsVerified {
		s.logger.Warn("Attempt to create user with unverified session", "sessionID", input.SessionID)
		return ErrSessionNotVerified
	}

	var user models.User

	if err := s.postgres.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found for password change", "email", input.Email)
			return ErrUserNotFound
		}
		s.logger.Error("Failed to find user", "email", input.Email, "error", err)
		return err
	}

	hashPass, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		s.logger.Error("Failed to hash new password", "error", err)
		return fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	user.Password = hashPass

	if err := s.postgres.Save(&user).Error; err != nil {
		s.logger.Error("Failed to update user password", "userID", user.ID, "error", err)
		return err
	}

	s.logger.Info("Password changed successfully", "userID", user.ID, "email", input.Email)
	return nil
}

func generateUsername() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("user%d", r.Intn(1000000)+1)
}
