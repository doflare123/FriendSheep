package events_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"friendship/services/events"
)

type lifecycleScheduleReaderStub struct {
	items []events.EventLifecycleScheduleEvent
	err   error
	ctx   context.Context
	after int64
	limit int
	calls int
}

func (s *lifecycleScheduleReaderStub) ListLifecycleScheduleEvents(
	ctx context.Context,
	after int64,
	limit int,
) ([]events.EventLifecycleScheduleEvent, error) {
	s.calls++
	s.ctx = ctx
	s.after = after
	s.limit = limit
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]events.EventLifecycleScheduleEvent(nil), s.items...), s.err
}

func TestLifecycleScheduleOutboxListUsesExclusiveCursorAndLookAheadPagination(t *testing.T) {
	t.Parallel()

	reader := &lifecycleScheduleReaderStub{items: lifecycleScheduleEvents(41, 42, 43)}
	service := events.NewEventLifecycleScheduleOutboxService(reader)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "preserved")

	page, err := service.List(ctx, 40, 2)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if reader.calls != 1 || reader.ctx != ctx || reader.after != 40 || reader.limit != 3 {
		t.Fatalf("reader call = count:%d context:%t after:%d limit:%d, want 1/true/40/3",
			reader.calls, reader.ctx == ctx, reader.after, reader.limit)
	}
	if !page.HasMore || page.NextCursor != 42 {
		t.Fatalf("page metadata = %#v, want hasMore=true nextCursor=42", page)
	}
	if want := lifecycleScheduleEvents(41, 42); !reflect.DeepEqual(page.Items, want) {
		t.Fatalf("page items = %#v, want %#v", page.Items, want)
	}
}

func TestLifecycleScheduleOutboxListUsesDefaultAndMaximumLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		limit           int
		wantReaderLimit int
		wantErr         error
	}{
		{name: "default", limit: 0, wantReaderLimit: events.DefaultLifecycleScheduleEventsLimit + 1},
		{name: "maximum", limit: events.MaxLifecycleScheduleEventsLimit, wantReaderLimit: events.MaxLifecycleScheduleEventsLimit + 1},
		{name: "negative", limit: -1, wantErr: events.ErrInvalidScheduleEventsLimit},
		{name: "above maximum", limit: events.MaxLifecycleScheduleEventsLimit + 1, wantErr: events.ErrInvalidScheduleEventsLimit},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reader := &lifecycleScheduleReaderStub{}
			service := events.NewEventLifecycleScheduleOutboxService(reader)
			_, err := service.List(context.Background(), 0, test.limit)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("List() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				if reader.calls != 0 {
					t.Fatalf("reader called %d times after invalid limit", reader.calls)
				}
				return
			}
			if reader.calls != 1 || reader.limit != test.wantReaderLimit {
				t.Fatalf("reader call = count:%d limit:%d, want 1/%d", reader.calls, reader.limit, test.wantReaderLimit)
			}
		})
	}
}

func TestLifecycleScheduleOutboxListRejectsNegativeCursorBeforeStorage(t *testing.T) {
	t.Parallel()

	reader := &lifecycleScheduleReaderStub{}
	service := events.NewEventLifecycleScheduleOutboxService(reader)
	if _, err := service.List(context.Background(), -1, 10); !errors.Is(err, events.ErrInvalidScheduleEventsCursor) {
		t.Fatalf("List() error = %v, want ErrInvalidScheduleEventsCursor", err)
	}
	if reader.calls != 0 {
		t.Fatalf("reader called %d times for invalid cursor", reader.calls)
	}
}

func TestLifecycleScheduleOutboxListReturnsStableEmptyPage(t *testing.T) {
	t.Parallel()

	service := events.NewEventLifecycleScheduleOutboxService(&lifecycleScheduleReaderStub{})
	page, err := service.List(context.Background(), 987, 10)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if page.Items == nil || len(page.Items) != 0 || page.NextCursor != 987 || page.HasMore {
		t.Fatalf("empty page = %#v, want non-nil items, unchanged cursor, hasMore=false", page)
	}
}

func TestLifecycleScheduleOutboxListPreservesCancellationAndReaderErrors(t *testing.T) {
	t.Parallel()

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	reader := &lifecycleScheduleReaderStub{}
	service := events.NewEventLifecycleScheduleOutboxService(reader)
	if _, err := service.List(cancelled, 0, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled List() error = %v, want context.Canceled", err)
	}

	injected := errors.New("injected outbox read failure")
	reader = &lifecycleScheduleReaderStub{err: injected}
	service = events.NewEventLifecycleScheduleOutboxService(reader)
	if _, err := service.List(context.Background(), 0, 10); !errors.Is(err, injected) {
		t.Fatalf("List() error = %v, want injected error", err)
	}
}

func lifecycleScheduleEvents(sequences ...int64) []events.EventLifecycleScheduleEvent {
	start := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	result := make([]events.EventLifecycleScheduleEvent, 0, len(sequences))
	for _, sequence := range sequences {
		result = append(result, events.EventLifecycleScheduleEvent{
			Sequence:      sequence,
			MessageID:     "00000000-0000-0000-0000-000000000042",
			SchemaVersion: events.LifecycleScheduleSchemaVersion,
			Operation:     events.LifecycleScheduleOperationUpsert,
			EventID:       42,
			StartTime:     &start,
			EndTime:       eventScheduleTimePointer(start.Add(2 * time.Hour)),
			OccurredAt:    start.Add(-time.Hour),
		})
	}
	return result
}

func eventScheduleTimePointer(value time.Time) *time.Time {
	return &value
}

var _ events.EventLifecycleScheduleOutboxReader = (*lifecycleScheduleReaderStub)(nil)
