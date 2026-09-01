package reminders

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestReminderIntentFixturesDecodeAgainstConsumerValidation(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"event_reminder_intent_v1_upsert.json",
		"event_reminder_intent_v1_cancel.json",
	} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body, err := os.ReadFile(filepath.Join(contractFixturesRoot(t), name))
			if err != nil {
				t.Fatalf("не удалось прочитать fixture %s: %v", name, err)
			}
			var item ReminderIntent
			if err := json.Unmarshal(body, &item); err != nil {
				t.Fatalf("не удалось декодировать fixture %s: %v", name, err)
			}
			if err := validateSourceBatch([]ReminderIntent{item}); err != nil {
				t.Fatalf("fixture %s не проходит consumer validation: %v", name, err)
			}
		})
	}
}

func TestReminderIntentSchemaExistsAsSharedContractSource(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join(contractFixturesRoot(t), "event_reminder_intent_v1.schema.json")
	info, err := os.Stat(schemaPath)
	if err != nil {
		t.Fatalf("не удалось прочитать schema file %s: %v", schemaPath, err)
	}
	if info.Size() == 0 {
		t.Fatalf("schema file %s пустой", schemaPath)
	}
}

func TestReminderIntentConsumerRejectsMalformedMessageID(t *testing.T) {
	t.Parallel()

	start := time.Date(2036, 9, 1, 18, 0, 0, 0, time.UTC)
	item := ReminderIntent{
		Sequence: 1, MessageID: "not-a-uuid", SchemaVersion: 1,
		IntentType: IntentTypeEventReminder, Operation: OperationScheduleUpsert,
		EventID: 42, StartTime: &start, ReminderOffsetMinutes: []int{1440, 360, 60}, OccurredAt: start.Add(-time.Hour),
	}
	if err := validateSourceBatch([]ReminderIntent{item}); err == nil {
		t.Fatal("некорректный messageId неожиданно прошёл consumer validation")
	}
}

func TestReminderIntentConsumerRejectsUnsupportedOffset(t *testing.T) {
	t.Parallel()

	start := time.Date(2036, 9, 1, 18, 0, 0, 0, time.UTC)
	item := ReminderIntent{
		Sequence: 1, MessageID: "00000000-0000-0000-0000-000000000042", SchemaVersion: 1,
		IntentType: IntentTypeEventReminder, Operation: OperationScheduleUpsert,
		EventID: 42, StartTime: &start, ReminderOffsetMinutes: []int{1440, 360, 90}, OccurredAt: start.Add(-time.Hour),
	}
	if err := validateSourceBatch([]ReminderIntent{item}); err == nil {
		t.Fatal("неподдерживаемый reminder offset неожиданно прошёл consumer validation")
	}
}

func contractFixturesRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("не удалось определить путь к тестовому файлу")
	}
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "backend", "docs", "internal", "contracts")
}
