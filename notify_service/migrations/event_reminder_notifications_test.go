package migrations_test

import (
	"io/fs"
	"strings"
	"testing"

	"notify_service/migrations"
)

func TestEventReminderMigrationDefinesDurableInboxAndDeliveryState(t *testing.T) {
	t.Parallel()

	body, err := fs.ReadFile(migrations.Files, "000003_event_reminder_notifications.sql")
	if err != nil {
		t.Fatalf("не удалось прочитать reminder migration: %v", err)
	}
	sql := strings.ToLower(string(body))
	required := []string{
		"notify_service.event_reminder_source_cursors",
		"notify_service.event_reminder_source_messages",
		"message_id uuid not null",
		"primary key (source_name, message_id)",
		"notify_service.event_reminder_jobs",
		"source_message_id uuid not null",
		"unique (event_id, source_revision, reminder_offset_minutes)",
		"skipped_missed_window",
		"superseded",
		"expired",
		"idx_event_reminder_jobs_due",
		"idx_event_reminder_jobs_retry",
		"idx_event_reminder_jobs_lease",
		"notify_service.notifications",
		"idempotency_key text not null unique",
		"idx_notifications_unread",
		"notify_service.notification_delivery_attempts",
		"unique (notification_id, channel_code, attempt_number)",
		"idx_notification_delivery_attempts_delivered_once",
	}
	for _, marker := range required {
		if !strings.Contains(sql, marker) {
			t.Errorf("reminder migration не содержит %q", marker)
		}
	}
	for _, forbidden := range []string{"public.events", "public.users", "telegram", "firebase", "fcm_token", "email"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("reminder migration содержит внешний или чужой marker %q", forbidden)
		}
	}
}
