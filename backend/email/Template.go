package email

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

const (
	// Основные цвета сервиса FriendShip
	ColorWhite    = "#FFFFFF"
	ColorBlue     = "#37A2E6"
	ColorDarkBlue = "#316BC2"
	LogoURL       = "https://5f6eebb9-e236-4aa6-b0ab-329c75a0d00b.selstorage.ru/logo-app-no-back.png"
	ServiceName   = "FriendShip"
)

type TemplateType string

const (
	TemplateVerificationCode TemplateType = "verification_code"
	TemplateResetPassword    TemplateType = "reset_password"
	TemplatePopularEvent     TemplateType = "popular_event"
	TemplateEventReminder    TemplateType = "event_reminder"
	TemplateWelcome          TemplateType = "welcome"
	TemplateGroupInvitation  TemplateType = "group_invitation"
)

type EmailData struct {
	ServiceName string
	LogoURL     string
	Year        int

	Code           string
	Title          string
	Message        string
	ActionURL      string
	ActionText     string
	EventName      string
	EventDate      string
	GroupName      string
	UserName       string
	AdditionalInfo string
}

// Содержит общую структуру для всех email
const baseTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background-color: #f4f6f9;
			margin: 0;
			padding: 20px 0;
			line-height: 1.6;
		}
		.email-wrapper {
			max-width: 600px;
			margin: 0 auto;
			background-color: #FFFFFF;
		}
		.header {
			background: linear-gradient(135deg, #37A2E6 0%, #316BC2 100%);
			padding: 30px 20px;
			text-align: center;
		}
		.header img {
			max-width: 180px;
			height: auto;
		}
		.header h1 {
			color: #FFFFFF;
			font-size: 24px;
			margin-top: 15px;
			font-weight: 600;
		}
		.content {
			padding: 40px 30px;
		}
		.greeting {
			font-size: 18px;
			color: #333;
			margin-bottom: 20px;
			font-weight: 500;
		}
		.message {
			font-size: 15px;
			color: #555;
			line-height: 1.8;
			margin-bottom: 25px;
		}
		.code-container {
			background: linear-gradient(135deg, #E8F4FD 0%, #D6EBFA 100%);
			border: 2px solid #37A2E6;
			border-radius: 12px;
			padding: 25px;
			text-align: center;
			margin: 30px 0;
		}
		.code {
			font-size: 32px;
			font-weight: bold;
			color: #316BC2;
			letter-spacing: 8px;
			font-family: 'Courier New', monospace;
			user-select: all;
			display: inline-block;
			padding: 10px 20px;
			background: #FFFFFF;
			border-radius: 8px;
			box-shadow: 0 2px 8px rgba(49, 107, 194, 0.15);
		}
		.action-button {
			display: inline-block;
			background: linear-gradient(135deg, #37A2E6 0%, #316BC2 100%);
			color: #FFFFFF !important;
			text-decoration: none;
			padding: 14px 35px;
			border-radius: 8px;
			font-weight: 600;
			font-size: 16px;
			margin: 20px 0;
			transition: transform 0.2s;
			box-shadow: 0 4px 12px rgba(49, 107, 194, 0.3);
		}
		.action-button:hover {
			transform: translateY(-2px);
			box-shadow: 0 6px 16px rgba(49, 107, 194, 0.4);
		}
		.info-box {
			background: #F8FAFB;
			border-left: 4px solid #37A2E6;
			padding: 15px 20px;
			margin: 20px 0;
			border-radius: 6px;
		}
		.info-box p {
			margin: 8px 0;
			font-size: 14px;
			color: #555;
		}
		.info-box strong {
			color: #316BC2;
		}
		.footer {
			background: #F8FAFB;
			padding: 30px;
			text-align: center;
			border-top: 1px solid #E5E9F0;
		}
		.footer-text {
			font-size: 13px;
			color: #8B95A5;
			margin: 8px 0;
		}
		.footer-links {
			margin: 15px 0;
		}
		.footer-link {
			color: #37A2E6;
			text-decoration: none;
			margin: 0 10px;
			font-size: 13px;
		}
		.footer-link:hover {
			color: #316BC2;
			text-decoration: underline;
		}
		.social-links {
			margin: 15px 0;
		}
		.divider {
			height: 1px;
			background: #E5E9F0;
			margin: 25px 0;
		}
		.warning {
			background: #FFF8E1;
			border-left: 4px solid #FFA726;
			padding: 12px 15px;
			margin: 20px 0;
			border-radius: 6px;
			font-size: 13px;
			color: #666;
		}
	</style>
</head>
<body>
	<div class="email-wrapper">
		<div class="header">
			<img src="{{.LogoURL}}" alt="{{.ServiceName}}">
			<h1>{{.ServiceName}}</h1>
		</div>
		
		<div class="content">
			{{template "body" .}}
		</div>
		
		<div class="footer">
			<p class="footer-text">© {{.Year}} {{.ServiceName}}. Все права защищены.</p>
			<p class="footer-text">Сервис для поиска друзей и организации мероприятий</p>
			<div class="footer-links">
				<a href="#" class="footer-link">О нас</a>
				<a href="#" class="footer-link">Помощь</a>
				<a href="#" class="footer-link">Политика конфиденциальности</a>
			</div>
			<p class="footer-text" style="margin-top: 15px; font-size: 12px;">
				Это автоматическое письмо. Пожалуйста, не отвечайте на него.
			</p>
		</div>
	</div>
</body>
</html>
`

// Тело письма для кода подтверждения
const verificationCodeBody = `
{{define "body"}}
	<h2 class="greeting">Здравствуйте! 👋</h2>
	<p class="message">{{.Message}}</p>
	
	<div class="code-container">
		<p style="margin-bottom: 10px; color: #555; font-size: 14px;">Ваш код подтверждения:</p>
		<div class="code">{{.Code}}</div>
	</div>
	
	<div class="warning">
		⏰ Код действителен в течение <strong>10 минут</strong>. Если вы не запрашивали это действие, просто проигнорируйте это письмо.
	</div>
	
	<div class="divider"></div>
	
	<p style="font-size: 13px; color: #8B95A5;">
		Для безопасности вашего аккаунта никогда не сообщайте этот код другим людям.
	</p>
{{end}}
`

// Тело письма о попадании события в топ
const popularEventBody = `
{{define "body"}}
	<h2 class="greeting">Поздравляем! 🎉</h2>
	<p class="message">
		Отличные новости! Ваше событие <strong>{{.EventName}}</strong> попало в подборку самых популярных мероприятий на {{.ServiceName}}!
	</p>
	
	<div class="info-box">
		<p><strong>Событие:</strong> {{.EventName}}</p>
		<p><strong>Группа:</strong> {{.GroupName}}</p>
		<p><strong>Дата проведения:</strong> {{.EventDate}}</p>
	</div>
	
	<p class="message">
		Это значит, что ваше событие увидят тысячи пользователей! Популярные события отображаются на главной странице и получают приоритет в поиске.
	</p>
	
	<center>
		<a href="{{.ActionURL}}" class="action-button">{{.ActionText}}</a>
	</center>
	
	<div class="divider"></div>
	
	<p style="font-size: 14px; color: #555;">
		<strong>Советы для организаторов:</strong>
	</p>
	<ul style="font-size: 14px; color: #555; line-height: 1.8; margin: 10px 0 0 20px;">
		<li>Своевременно отвечайте на вопросы участников</li>
		<li>Обновляйте информацию о событии при необходимости</li>
		<li>Поддерживайте активность в группе</li>
	</ul>
{{end}}
`

// Напоминание о событии
const eventReminderBody = `
{{define "body"}}
	<h2 class="greeting">Напоминаем! ⏰</h2>
	<p class="message">
		Не забудьте о событии <strong>{{.EventName}}</strong>, которое скоро начнётся!
	</p>
	
	<div class="info-box">
		<p><strong>Событие:</strong> {{.EventName}}</p>
		<p><strong>Группа:</strong> {{.GroupName}}</p>
		<p><strong>Дата и время:</strong> {{.EventDate}}</p>
		{{if .AdditionalInfo}}
		<p><strong>Место:</strong> {{.AdditionalInfo}}</p>
		{{end}}
	</div>
	
	<center>
		<a href="{{.ActionURL}}" class="action-button">{{.ActionText}}</a>
	</center>
	
	<p class="message" style="margin-top: 25px;">
		Будем рады видеть вас! Если планы изменились, вы можете отменить участие в личном кабинете.
	</p>
{{end}}
`

// Приветственное письмо
const welcomeBody = `
{{define "body"}}
	<h2 class="greeting">Добро пожаловать в {{.ServiceName}}! 🎊</h2>
	<p class="message">
		Привет, <strong>{{.UserName}}</strong>! Мы рады приветствовать вас в нашем сообществе.
	</p>
	
	<p class="message">
		{{.ServiceName}} — это платформа для поиска друзей по интересам и организации совместных мероприятий. 
		Здесь вы можете создавать события, присоединяться к группам и находить единомышленников.
	</p>
	
	<div class="info-box">
		<p><strong>Что вы можете сделать прямо сейчас:</strong></p>
		<ul style="margin: 10px 0 0 20px; line-height: 1.8;">
			<li>Найти интересные события и группы</li>
			<li>Создать своё первое событие</li>
			<li>Настроить профиль</li>
			<li>Пригласить друзей</li>
		</ul>
	</div>
	
	<center>
		<a href="{{.ActionURL}}" class="action-button">{{.ActionText}}</a>
	</center>
	
	<p class="message" style="margin-top: 25px;">
		Если у вас возникнут вопросы, наша команда поддержки всегда готова помочь!
	</p>
{{end}}
`

// Приглашение в группу
const groupInvitationBody = `
{{define "body"}}
	<h2 class="greeting">Вас пригласили в группу! 📨</h2>
	<p class="message">
		У вас новое приглашение в группу <strong>{{.GroupName}}</strong> на {{.ServiceName}}!
	</p>
	
	<div class="info-box">
		<p><strong>Группа:</strong> {{.GroupName}}</p>
		{{if .AdditionalInfo}}
		<p><strong>Описание:</strong> {{.AdditionalInfo}}</p>
		{{end}}
	</div>
	
	<p class="message">
		Присоединяйтесь к группе, чтобы участвовать в мероприятиях и общаться с единомышленниками!
	</p>
	
	<center>
		<a href="{{.ActionURL}}" class="action-button">{{.ActionText}}</a>
	</center>
{{end}}
`

// Управляет шаблонами email
type EmailTemplateManager struct {
	templates map[TemplateType]*template.Template
}

// Создаёт новый менеджер шаблонов
func NewEmailTemplateManager() (*EmailTemplateManager, error) {
	manager := &EmailTemplateManager{
		templates: make(map[TemplateType]*template.Template),
	}

	if err := manager.registerTemplate(TemplateVerificationCode, verificationCodeBody); err != nil {
		return nil, err
	}
	if err := manager.registerTemplate(TemplateResetPassword, verificationCodeBody); err != nil {
		return nil, err
	}
	if err := manager.registerTemplate(TemplatePopularEvent, popularEventBody); err != nil {
		return nil, err
	}
	if err := manager.registerTemplate(TemplateEventReminder, eventReminderBody); err != nil {
		return nil, err
	}
	if err := manager.registerTemplate(TemplateWelcome, welcomeBody); err != nil {
		return nil, err
	}
	if err := manager.registerTemplate(TemplateGroupInvitation, groupInvitationBody); err != nil {
		return nil, err
	}

	return manager, nil
}

// Регистрирует шаблон
func (m *EmailTemplateManager) registerTemplate(templateType TemplateType, bodyTemplate string) error {
	tmpl, err := template.New(string(templateType)).Parse(baseTemplate + bodyTemplate)
	if err != nil {
		return fmt.Errorf("ошибка парсинга шаблона %s: %w", templateType, err)
	}
	m.templates[templateType] = tmpl
	return nil
}

// Рендерит шаблон с данными
func (m *EmailTemplateManager) RenderTemplate(templateType TemplateType, data EmailData) (string, error) {
	tmpl, exists := m.templates[templateType]
	if !exists {
		return "", fmt.Errorf("шаблон %s не найден", templateType)
	}

	if data.ServiceName == "" {
		data.ServiceName = ServiceName
	}
	if data.LogoURL == "" {
		data.LogoURL = LogoURL
	}
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("ошибка рендеринга шаблона: %w", err)
	}

	return buf.String(), nil
}

// Возвращает тему письма для типа шаблона
func GetSubject(templateType TemplateType, customSubject string) string {
	if customSubject != "" {
		return customSubject
	}

	subjects := map[TemplateType]string{
		TemplateVerificationCode: "Код подтверждения — " + ServiceName,
		TemplateResetPassword:    "Восстановление пароля — " + ServiceName,
		TemplatePopularEvent:     "🎉 Ваше событие в топе! — " + ServiceName,
		TemplateEventReminder:    "⏰ Напоминание о событии — " + ServiceName,
		TemplateWelcome:          "Добро пожаловать в " + ServiceName + "!",
		TemplateGroupInvitation:  "📨 Приглашение в группу — " + ServiceName,
	}

	return subjects[templateType]
}
