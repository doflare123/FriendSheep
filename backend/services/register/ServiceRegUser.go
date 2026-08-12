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
	session "friendship/sessions"
	"friendship/utils"
	"math/rand"
	"time"
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
	ErrSessionNotVerified   = errors.New("сессия не подтверждена")
	ErrSessionNotFound      = errors.New("сессия не найдена или удалена")
	ErrSessionTypeMismatch  = errors.New("тип сессии не совпадает")
	ErrSessionEmailMismatch = errors.New("email сессии не совпадает")
	ErrInvalidCode          = errors.New("неверный код")
	ErrTooManyAttempts      = errors.New("превышено количество попыток ввода кода, сессия удалена")
	ErrUserAlreadyExists    = errors.New("пользователь с таким email уже существует")
	ErrUserNotFound         = errors.New("пользователь не найден")
)

type RegService interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*dto.AuthResponse, error)
	CreateSessionRegister(ctx context.Context, email string, type_ses string) (*models.SessionRegResponse, error)
	VerifySession(ctx context.Context, input VerifySessionInput) (bool, error)
	ChangePassword(ctx context.Context, input ChangePasswordInput) error
}

type RegistrationAuthService interface {
	IssueTokens(ctx context.Context, userID uint) (dto.AuthResponse, error)
	RevokeAll(ctx context.Context, userID uint) error
}

type registrationEmailSender interface {
	SendVerificationEmail(userEmail, code, actionType string) error
	SendWelcomeEmail(userEmail, userName string) error
}

type regService struct {
	logger   logger.Logger
	redis    session.SessionStore
	cfg      *config.Config
	store    RegistrationStore
	auth     RegistrationAuthService
	notifier registrationEmailSender
}

func NewRegisterSrv(
	logger logger.Logger,
	redis session.SessionStore,
	store RegistrationStore,
	cfg *config.Config,
	auth RegistrationAuthService,
) (RegService, error) {
	emailManager, err := email.NewEmailTemplateManager()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации email менеджера: %w", err)
	}

	return NewRegisterSrvWithEmailSender(
		logger,
		redis,
		store,
		cfg,
		auth,
		&templateRegistrationEmailSender{
			logger:       logger,
			cfg:          cfg,
			emailManager: emailManager,
		},
	), nil
}

func NewRegisterSrvWithEmailSender(
	logger logger.Logger,
	redis session.SessionStore,
	store RegistrationStore,
	cfg *config.Config,
	auth RegistrationAuthService,
	notifier registrationEmailSender,
) RegService {
	return &regService{
		logger:   logger,
		redis:    redis,
		cfg:      cfg,
		store:    store,
		auth:     auth,
		notifier: notifier,
	}
}

func (s *regService) CreateUser(ctx context.Context, input CreateUserInput) (*dto.AuthResponse, error) {
	sess, err := s.redis.GetSession(ctx, input.SessionID)
	if err != nil {
		s.logger.Error("Не удалось получить сессию", "sessionID", input.SessionID, "error", err)
		return nil, mapRegisterSessionLookupError(err)
	}

	boundEmail, err := validateVerifiedSessionBinding(sess, models.SessionTypeRegister, input.Email)
	if err != nil {
		s.logger.Warn("Несовпадение привязки сессии при регистрации", "sessionID", input.SessionID, "email", input.Email, "error", err)
		return nil, err
	}

	us := generateUsername()

	hashPass, err := utils.HashPassword(input.Password)
	if err != nil {
		s.logger.Error("Не удалось захешировать пароль", "error", err)
		return nil, fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	user, err := s.store.CreateUser(ctx, UserBootstrap{
		Name:           input.Name,
		HashedPassword: hashPass,
		Email:          boundEmail,
		Username:       us,
	})

	if err != nil {
		s.logger.Error("Не удалось создать пользователя", "error", err)
		return nil, err
	}

	if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
		s.logger.Warn("Не удалось удалить сессию после создания пользователя", "sessionID", input.SessionID, "error", err)
	}
	if err := s.notifier.SendWelcomeEmail(user.Email, user.Name); err != nil {
		s.logger.Warn("Не удалось отправить приветственное письмо после создания пользователя", "userID", user.ID, "email", user.Email, "error", err)
	}

	s.logger.Info("Пользователь успешно создан", "userID", user.ID, "email", user.Email)

	if s.auth == nil {
		return nil, fmt.Errorf("пользователь создан, но сервис авторизации не настроен")
	}
	authResponse, err := s.auth.IssueTokens(ctx, user.ID)
	if err != nil {
		s.logger.Error("Не удалось сгенерировать токены после регистрации", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("пользователь создан, но не удалось сгенерировать токены: %w", err)
	}

	s.logger.Info("Пользователь автоматически авторизован после регистрации", "userID", user.ID)

	return &authResponse, nil
}

func (s *regService) CreateSessionRegister(ctx context.Context, email, type_ses string) (*models.SessionRegResponse, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	if type_ses != string(models.SessionTypeRegister) && type_ses != string(models.SessionTypeResetPassword) {
		return nil, fmt.Errorf("некорректный тип сессии: %s", type_ses)
	}

	sessionID := utils.GenerateSessioID(12)
	code := utils.GenerationSessionCode(6)

	if type_ses == string(models.SessionTypeRegister) {
		err := s.redis.CreateSession(
			ctx,
			sessionID,
			code,
			models.SessionTypeRegister,
			10*time.Minute,
			map[string]string{
				"email": normalizedEmail,
			},
		)
		if err != nil {
			s.logger.Error("Не удалось создать сессию", "email", normalizedEmail, "error", err)
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
				"email": normalizedEmail,
			},
		)
		if err != nil {
			s.logger.Error("Не удалось создать сессию", "email", normalizedEmail, "error", err)
			return nil, fmt.Errorf("не удалось создать сессию: %w", err)
		}
	}

	if err := s.notifier.SendVerificationEmail(normalizedEmail, code, type_ses); err != nil {
		if deleteErr := s.redis.DeleteSession(ctx, sessionID); deleteErr != nil {
			s.logger.Warn("Не удалось удалить сессию после ошибки отправки письма", "sessionID", sessionID, "error", deleteErr)
		}
		return nil, fmt.Errorf("не удалось отправить письмо подтверждения: %w", err)
	}

	s.logger.Info("Сессия регистрации создана", "sessionID", sessionID, "email", normalizedEmail)
	return &models.SessionRegResponse{SessionID: sessionID}, nil
}

type templateRegistrationEmailSender struct {
	logger       logger.Logger
	cfg          *config.Config
	emailManager *email.EmailTemplateManager
}

func (s *templateRegistrationEmailSender) SendVerificationEmail(userEmail, code, actionType string) error {
	messageMap := map[string]string{
		"reset_password": "Чтобы завершить смену пароля, введите следующий код подтверждения:",
		"register":       "Чтобы завершить регистрацию, введите следующий код подтверждения:",
	}

	message, exists := messageMap[actionType]
	if !exists {
		message = "Введите следующий код подтверждения:"
	}

	var templateType email.TemplateType
	if actionType == "reset_password" {
		templateType = email.TemplateResetPassword
	} else {
		templateType = email.TemplateVerificationCode
	}

	emailData := email.EmailData{
		Code:    code,
		Message: message,
	}

	body, err := s.emailManager.RenderTemplate(templateType, emailData)
	if err != nil {
		return fmt.Errorf("ошибка рендеринга email шаблона: %w", err)
	}

	subject := email.GetSubject(templateType, "")
	if err := utils.SendEmail(userEmail, subject, body, s.cfg); err != nil {
		s.logger.Error("Не удалось отправить письмо с кодом подтверждения", "email", userEmail, "error", err)
		return err
	}

	s.logger.Info("Письмо с кодом подтверждения отправлено", "email", userEmail, "type", actionType)
	return nil
}

func (s *templateRegistrationEmailSender) SendWelcomeEmail(userEmail, userName string) error {
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
		s.logger.Error("Не удалось отправить приветственное письмо", "email", userEmail, "error", err)
		return err
	}

	s.logger.Info("Приветственное письмо отправлено", "email", userEmail)
	return nil
}

func (s *regService) VerifySession(ctx context.Context, input VerifySessionInput) (bool, error) {
	sess, err := s.redis.GetSession(ctx, input.SessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			s.logger.Warn("Сессия не найдена", "sessionID", input.SessionID)
			return false, ErrSessionNotFound
		}
		s.logger.Error("Не удалось получить сессию", "sessionID", input.SessionID, "error", err)
		return false, err
	}

	if string(sess.Type) != input.Type {
		s.logger.Warn("Несовпадение типа сессии", "sessionID", input.SessionID, "expected", input.Type, "actual", sess.Type)
		return false, ErrSessionTypeMismatch
	}

	if sess.Attempts >= 3 {
		s.logger.Warn("Превышено число попыток, удаляем сессию", "sessionID", input.SessionID)
		if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
			s.logger.Error("Не удалось удалить сессию после превышения числа попыток", "sessionID", input.SessionID, "error", err)
		}
		return false, ErrTooManyAttempts
	}

	if sess.Code != input.Code {
		attempts, err := s.redis.IncrementAttempts(ctx, input.SessionID)
		if err != nil {
			s.logger.Error("Не удалось увеличить счетчик попыток", "sessionID", input.SessionID, "error", err)
			return false, err
		}

		s.logger.Warn("Неверный код", "sessionID", input.SessionID, "attempts", attempts)

		if attempts >= 3 {
			if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
				s.logger.Error("Не удалось удалить сессию после последней попытки", "sessionID", input.SessionID, "error", err)
			}
			return false, ErrTooManyAttempts
		}

		return false, ErrInvalidCode
	}

	if err := s.redis.MarkAsVerified(ctx, input.SessionID); err != nil {
		s.logger.Error("Не удалось пометить сессию как подтвержденную", "sessionID", input.SessionID, "error", err)
		return false, err
	}

	s.logger.Info("Сессия успешно подтверждена", "sessionID", input.SessionID)
	return true, nil
}

func (s *regService) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	sess, err := s.redis.GetSession(ctx, input.SessionID)
	if err != nil {
		s.logger.Error("Не удалось получить сессию", "sessionID", input.SessionID, "error", err)
		return mapRegisterSessionLookupError(err)
	}

	boundEmail, err := validateVerifiedSessionBinding(sess, models.SessionTypeResetPassword, input.Email)
	if err != nil {
		s.logger.Warn("Несовпадение привязки сессии при смене пароля", "sessionID", input.SessionID, "email", input.Email, "error", err)
		return err
	}

	hashPass, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		s.logger.Error("Не удалось захешировать новый пароль", "error", err)
		return fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	userID, err := s.store.FindUserIDByEmail(ctx, boundEmail)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if s.auth == nil {
		return fmt.Errorf("сервис отзыва авторизации не настроен")
	}

	if err := s.auth.RevokeAll(ctx, userID); err != nil {
		s.logger.Error("Не удалось отозвать auth сессии перед сменой пароля", "userID", userID, "error", err)
		return fmt.Errorf("не удалось завершить активные сессии перед сменой пароля: %w", err)
	}

	if err := s.redis.DeleteSession(ctx, input.SessionID); err != nil {
		s.logger.Error("Не удалось поглотить reset-сессию перед сменой пароля", "sessionID", input.SessionID, "error", err)
		return fmt.Errorf("не удалось завершить reset-сессию: %w", err)
	}

	changedUserID, err := s.store.ChangePasswordByEmail(ctx, boundEmail, hashPass)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.logger.Warn("Пользователь для смены пароля не найден", "email", boundEmail)
			return ErrUserNotFound
		}
		s.logger.Error("Не удалось обновить пароль пользователя", "email", boundEmail, "error", err)
		return err
	}
	if changedUserID != userID {
		return fmt.Errorf("идентификатор пользователя изменился во время смены пароля")
	}

	if err := s.auth.RevokeAll(ctx, userID); err != nil {
		s.logger.Error("Не удалось отозвать auth сессии после смены пароля", "userID", userID, "error", err)
		return fmt.Errorf("пароль изменён, но не удалось завершить активные сессии: %w", err)
	}

	s.logger.Info("Пароль успешно изменен", "userID", userID, "email", boundEmail)
	return nil
}

func generateUsername() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("user%d", r.Intn(1000000)+1)
}
