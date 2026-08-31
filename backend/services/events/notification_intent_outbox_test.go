package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"friendship/services/events"
)

type reminderIntentReaderStub struct {
	items []events.EventReminderIntent
	err   error
	ctx   context.Context
	after int64
	limit int
	calls int
}

func (s *reminderIntentReaderStub) ListEventReminderIntents(ctx context.Context, after int64, limit int) ([]events.EventReminderIntent, error) {
	s.calls++
	s.ctx = ctx
	s.after = after
	s.limit = limit
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]events.EventReminderIntent(nil), s.items...), s.err
}

func TestEventReminderIntentListUsesExclusiveCursorLookAheadAndStableReplay(t *testing.T) {
	reader := &reminderIntentReaderStub{items: reminderIntents(41, 42, 43)}
	service := events.NewEventReminderIntentService(reader)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "preserved")

	page, err := service.List(ctx, 40, 2)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	replay, err := service.List(ctx, 40, 2)
	if err != nil {
		t.Fatalf("replayed List() returned error: %v", err)
	}
	if reader.calls != 2 || reader.ctx != ctx || reader.after != 40 || reader.limit != 3 {
		t.Fatalf("reader call = count:%d context:%t after:%d limit:%d, want 2/true/40/3", reader.calls, reader.ctx == ctx, reader.after, reader.limit)
	}
	if !page.HasMore || page.NextCursor != 42 || !reflect.DeepEqual(page.Items, reminderIntents(41, 42)) {
		t.Fatalf("page = %#v, want sequences 41,42 and look-ahead", page)
	}
	if !reflect.DeepEqual(page, replay) {
		t.Fatalf("replay = %#v, want stable %#v", replay, page)
	}
}

func TestEventReminderIntentListValidatesPaginationAndStableEmptyPage(t *testing.T) {
	tests := []struct {
		name            string
		after           int64
		limit           int
		wantReaderLimit int
		wantErr         error
	}{
		{name: "default", limit: 0, wantReaderLimit: events.DefaultEventReminderIntentLimit + 1},
		{name: "maximum", limit: events.MaxEventReminderIntentLimit, wantReaderLimit: events.MaxEventReminderIntentLimit + 1},
		{name: "negative cursor", after: -1, limit: 10, wantErr: events.ErrInvalidEventReminderIntentCursor},
		{name: "negative limit", limit: -1, wantErr: events.ErrInvalidEventReminderIntentLimit},
		{name: "above maximum", limit: events.MaxEventReminderIntentLimit + 1, wantErr: events.ErrInvalidEventReminderIntentLimit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := &reminderIntentReaderStub{}
			page, err := events.NewEventReminderIntentService(reader).List(context.Background(), test.after, test.limit)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("List() err = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				if reader.calls != 0 {
					t.Fatalf("reader called %d times for invalid input", reader.calls)
				}
				return
			}
			if reader.calls != 1 || reader.limit != test.wantReaderLimit {
				t.Fatalf("reader call = count:%d limit:%d, want 1/%d", reader.calls, reader.limit, test.wantReaderLimit)
			}
			if page.Items == nil || len(page.Items) != 0 || page.NextCursor != test.after || page.HasMore {
				t.Fatalf("empty page = %#v, want non-nil items and unchanged cursor", page)
			}
		})
	}
}

func TestEventReminderIntentListPreservesCancellationAndReaderErrors(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	reader := &reminderIntentReaderStub{}
	if _, err := events.NewEventReminderIntentService(reader).List(cancelled, 0, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled List() err = %v, want context.Canceled", err)
	}

	injected := errors.New("injected outbox read failure")
	reader = &reminderIntentReaderStub{err: injected}
	if _, err := events.NewEventReminderIntentService(reader).List(context.Background(), 0, 10); !errors.Is(err, injected) {
		t.Fatalf("List() err = %v, want injected failure", err)
	}
}

func TestEventReminderIntentV1ContractFixturesMatchProviderConstants(t *testing.T) {
	contractDir := filepath.Join("..", "..", "docs", "internal", "contracts")
	schemaBytes, err := os.ReadFile(filepath.Join(contractDir, "event_reminder_intent_v1.schema.json"))
	if err != nil {
		t.Fatalf("read contract schema: %v", err)
	}
	var schema struct {
		Required   []string `json:"required"`
		Properties map[string]struct {
			Const any   `json:"const"`
			Enum  []any `json:"enum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatalf("decode contract schema: %v", err)
	}
	if got := schema.Properties["schemaVersion"].Const; got != float64(events.EventReminderIntentSchemaVersion) {
		t.Fatalf("schemaVersion const = %#v, want %d", got, events.EventReminderIntentSchemaVersion)
	}
	if got := schema.Properties["intentType"].Const; got != events.EventReminderIntentType {
		t.Fatalf("intentType const = %#v, want %q", got, events.EventReminderIntentType)
	}
	wantOperations := []any{events.EventReminderOperationUpsert, events.EventReminderOperationCancel}
	if !reflect.DeepEqual(schema.Properties["operation"].Enum, wantOperations) {
		t.Fatalf("operation enum = %#v, want %#v", schema.Properties["operation"].Enum, wantOperations)
	}

	for _, fixture := range []struct {
		name          string
		operation     string
		wantStartTime bool
		wantOffsets   []int
	}{
		{name: "event_reminder_intent_v1_upsert.json", operation: events.EventReminderOperationUpsert, wantStartTime: true, wantOffsets: []int{1440, 360, 60}},
		{name: "event_reminder_intent_v1_cancel.json", operation: events.EventReminderOperationCancel},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			fixtureBytes, err := os.ReadFile(filepath.Join(contractDir, fixture.name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			var intent events.EventReminderIntent
			if err := json.Unmarshal(fixtureBytes, &intent); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			if intent.Sequence <= 0 || intent.MessageID == "" || intent.EventID == 0 || intent.OccurredAt.IsZero() ||
				intent.SchemaVersion != events.EventReminderIntentSchemaVersion || intent.IntentType != events.EventReminderIntentType ||
				intent.Operation != fixture.operation {
				t.Fatalf("fixture decoded to invalid contract: %#v", intent)
			}
			if (intent.StartTime != nil) != fixture.wantStartTime || !reflect.DeepEqual(intent.ReminderOffsetMinutes, fixture.wantOffsets) {
				t.Fatalf("fixture schedule fields = start:%v offsets:%v, want start:%t offsets:%v", intent.StartTime, intent.ReminderOffsetMinutes, fixture.wantStartTime, fixture.wantOffsets)
			}
		})
	}
}

func reminderIntents(sequences ...int64) []events.EventReminderIntent {
	start := time.Date(2036, 8, 30, 18, 0, 0, 0, time.UTC)
	result := make([]events.EventReminderIntent, 0, len(sequences))
	for _, sequence := range sequences {
		startCopy := start
		result = append(result, events.EventReminderIntent{
			Sequence: sequence, MessageID: "00000000-0000-0000-0000-000000000042",
			SchemaVersion: events.EventReminderIntentSchemaVersion, IntentType: events.EventReminderIntentType,
			Operation: events.EventReminderOperationUpsert, EventID: 42, StartTime: &startCopy,
			ReminderOffsetMinutes: []int{1440, 360, 60}, OccurredAt: start.Add(-time.Hour),
		})
	}
	return result
}

var _ events.EventReminderIntentOutboxReader = (*reminderIntentReaderStub)(nil)
