package application_test

import (
	"context"
	"testing"

	"notify_service/internal/application"
)

type fakeChannel struct {
	sendCalls int
}

func (channel *fakeChannel) Send(context.Context, application.Notification) error {
	channel.sendCalls++
	return nil
}

var _ application.NotificationChannel = (*fakeChannel)(nil)

func TestBaselineApplicationHasNoNotificationChannels(t *testing.T) {
	t.Parallel()

	app := application.New()
	if got := app.ChannelCount(); got != 0 {
		t.Fatalf("ChannelCount() = %d, ожидалось 0", got)
	}
}

func TestApplicationCompositionDoesNotSendNotifications(t *testing.T) {
	t.Parallel()

	first := &fakeChannel{}
	second := &fakeChannel{}
	app := application.New(first, second)

	if got := app.ChannelCount(); got != 2 {
		t.Fatalf("ChannelCount() = %d, ожидалось 2", got)
	}
	if first.sendCalls != 0 || second.sendCalls != 0 {
		t.Fatalf("запуск приложения вызвал каналы уведомлений: первый=%d второй=%d", first.sendCalls, second.sendCalls)
	}
}

func TestNilApplicationReportsNoChannels(t *testing.T) {
	t.Parallel()

	var app *application.Application
	if got := app.ChannelCount(); got != 0 {
		t.Fatalf("ChannelCount() для nil-приложения = %d, ожидалось 0", got)
	}
}
