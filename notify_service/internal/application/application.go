package application

import (
	"context"
	"time"
)

// Notification не зависит от транспорта. Следующий срез доставки сможет расширить
// контракт без связи application-кода с SDK Telegram, FCM или электронной почты.
type Notification struct {
	ID          string
	RecipientID string
	Title       string
	Body        string
	ImageURL    string
	CreatedAt   time.Time
}

// NotificationChannel — исходящий порт доставки. Базовая версия не регистрирует каналы
// и никогда не вызывает Send.
type NotificationChannel interface {
	Send(context.Context, Notification) error
}

type Application struct {
	channels []NotificationChannel
}

func New(channels ...NotificationChannel) *Application {
	return &Application{channels: append([]NotificationChannel(nil), channels...)}
}

func (a *Application) ChannelCount() int {
	if a == nil {
		return 0
	}
	return len(a.channels)
}
