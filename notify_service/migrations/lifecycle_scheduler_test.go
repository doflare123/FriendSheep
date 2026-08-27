package migrations_test

import (
	"io/fs"
	"strings"
	"testing"

	"notify_service/migrations"
)

func TestLifecycleSchedulerMigrationDefinesDurableServiceOwnedState(t *testing.T) {
	t.Parallel()

	body, err := fs.ReadFile(migrations.Files, "000002_event_lifecycle_scheduler.sql")
	if err != nil {
		t.Fatalf("read lifecycle scheduler migration: %v", err)
	}
	sql := strings.ToLower(string(body))

	required := []string{
		"notify_service.event_lifecycle_source_cursors",
		"source_name text primary key",
		"last_sequence bigint not null",
		"notify_service.event_lifecycle_source_messages",
		"primary key (source_name, message_id)",
		"notify_service.event_lifecycle_jobs",
		"event_id bigint not null unique",
		"source_sequence bigint not null",
		"source_message_id uuid not null",
		"start_time timestamptz",
		"end_time timestamptz",
		"state varchar(32) not null",
		"next_action_at timestamptz",
		"attempt_count integer not null",
		"next_attempt_at timestamptz",
		"lease_until timestamptz",
		"last_error_code varchar(64)",
		"scheduled",
		"active_wait",
		"processing",
		"completed",
		"cancelled",
		"retry_wait",
		"terminal_failed",
		"idx_event_lifecycle_jobs_due_actions",
		"idx_event_lifecycle_jobs_due_retries",
		"idx_event_lifecycle_jobs_lease",
	}
	for _, marker := range required {
		if !strings.Contains(sql, marker) {
			t.Errorf("lifecycle migration is missing %q", marker)
		}
	}
	for _, forbidden := range []string{
		"public.events",
		"public.users",
		"public.groups",
		"event_title",
		"event_description",
		"recipient",
		"telegram",
		"firebase",
	} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("lifecycle migration contains out-of-scope/monolith marker %q", forbidden)
		}
	}
}

func TestLifecycleSchedulerMigrationHasGlobalMessageDeduplication(t *testing.T) {
	t.Parallel()

	body, err := fs.ReadFile(migrations.Files, "000002_event_lifecycle_scheduler.sql")
	if err != nil {
		t.Fatalf("read lifecycle scheduler migration: %v", err)
	}
	sql := strings.ToLower(string(body))
	jobColumnUnique := strings.Contains(sql, "source_message_id uuid not null unique")
	receiptKey := strings.Contains(sql, "notify_service.event_lifecycle_source_messages") &&
		strings.Contains(sql, "primary key (source_name, message_id)")
	if !jobColumnUnique && !receiptKey {
		t.Fatal("lifecycle source message IDs are not deduplicated by a unique job column or source receipt key")
	}
}
