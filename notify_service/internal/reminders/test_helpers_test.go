package reminders

import (
	"io"
	"log/slog"
	"time"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func discardReminderLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
