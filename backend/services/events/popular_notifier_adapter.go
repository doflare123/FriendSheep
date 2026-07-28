package events

import (
	"context"
	"fmt"
	"friendship/config"
	"friendship/email"
	"friendship/logger"
	"friendship/utils"
)

type emailPopularEventsNotifier struct {
	logger       logger.Logger
	emailManager *email.EmailTemplateManager
	config       *config.Config
}

func NewEmailPopularEventsNotifier(logger logger.Logger, cfg *config.Config) (PopularEventsNotifier, error) {
	emailManager, err := email.NewEmailTemplateManager()
	if err != nil {
		return nil, fmt.Errorf("initialize popular events email templates: %w", err)
	}

	return &emailPopularEventsNotifier{
		logger:       logger,
		emailManager: emailManager,
		config:       cfg,
	}, nil
}

func (n *emailPopularEventsNotifier) Notify(ctx context.Context, notifications []PopularEventNotification) error {
	for _, notification := range notifications {
		if err := ctx.Err(); err != nil {
			return err
		}

		body, err := n.emailManager.RenderTemplate(email.TemplatePopularEvent, email.EmailData{
			EventName:  notification.EventName,
			GroupName:  notification.GroupName,
			EventDate:  notification.StartTime.Format("02.01.2006 15:04"),
			ActionURL:  fmt.Sprintf("https://friendsheep.ru/events/%d", notification.EventID),
			ActionText: "Посмотреть событие",
		})
		if err != nil {
			n.logger.Error("Render popular event email failed", "eventID", notification.EventID, "error", err)
			continue
		}

		subject := email.GetSubject(email.TemplatePopularEvent, "")
		if err := utils.SendEmail(notification.RecipientEmail, subject, body, n.config); err != nil {
			n.logger.Error("Send popular event email failed", "eventID", notification.EventID, "email", notification.RecipientEmail, "error", err)
			continue
		}

		n.logger.Info("Popular event email sent", "eventID", notification.EventID, "email", notification.RecipientEmail)
	}

	return nil
}
