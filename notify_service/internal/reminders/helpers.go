package reminders

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"
)

var reminderMessageIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type DelaySource interface {
	After(time.Duration) <-chan time.Time
}

type RealDelaySource struct{}

func (RealDelaySource) After(delay time.Duration) <-chan time.Time {
	return time.After(delay)
}

type Jitter interface {
	Apply(time.Duration) time.Duration
}

type ExponentialBackoff struct {
	Min    time.Duration
	Max    time.Duration
	Jitter Jitter
}

func (b ExponentialBackoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	base := b.Min
	if base <= 0 {
		base = time.Second
	}
	max := b.Max
	if max < base {
		max = base
	}
	for i := 1; i < attempt && base < max; i++ {
		if base > max/2 {
			base = max
			break
		}
		base *= 2
	}
	if base > max {
		base = max
	}
	if b.Jitter != nil {
		return b.Jitter.Apply(base)
	}
	return base
}

type occurrence struct {
	OffsetMinutes int
	DueAt         time.Time
	WindowEndAt   time.Time
}

func canonicalOffsets(offsets []int) ([]int, error) {
	if len(offsets) == 0 {
		return nil, errors.New("reminderOffsetMinutes обязателен для schedule_upsert")
	}
	seen := make(map[int]struct{}, len(offsets))
	result := make([]int, 0, len(offsets))
	for _, offset := range offsets {
		if offset <= 0 {
			return nil, fmt.Errorf("reminderOffsetMinutes=%d должен быть положительным", offset)
		}
		if !isSupportedReminderOffset(offset) {
			return nil, fmt.Errorf("reminderOffsetMinutes=%d не поддерживается текущей версией контракта", offset)
		}
		if _, exists := seen[offset]; exists {
			return nil, fmt.Errorf("reminderOffsetMinutes=%d повторяется", offset)
		}
		seen[offset] = struct{}{}
		result = append(result, offset)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result, nil
}

func isSupportedReminderOffset(offset int) bool {
	switch offset {
	case 1440, 360, 60:
		return true
	default:
		return false
	}
}

func buildOccurrences(start time.Time, offsets []int) ([]occurrence, error) {
	canonical, err := canonicalOffsets(offsets)
	if err != nil {
		return nil, err
	}
	items := make([]occurrence, 0, len(canonical))
	for index, offset := range canonical {
		dueAt := start.Add(-time.Duration(offset) * time.Minute)
		windowEndAt := start
		if index < len(canonical)-1 {
			windowEndAt = start.Add(-time.Duration(canonical[index+1]) * time.Minute)
		}
		items = append(items, occurrence{
			OffsetMinutes: offset,
			DueAt:         dueAt,
			WindowEndAt:   windowEndAt,
		})
	}
	return items, nil
}

func evaluateJobWindow(job Job, now time.Time) string {
	if job.StartTime == nil || job.DueAt == nil || job.WindowEndAt == nil {
		return StateTerminalFailed
	}
	if !now.Before(*job.StartTime) {
		return StateExpired
	}
	if now.Before(*job.DueAt) {
		return StateScheduled
	}
	if !now.Before(*job.WindowEndAt) {
		return StateSkippedMissedWindow
	}
	return StateProcessing
}

func notificationIDempotencyKey(job Job, userID uint64) string {
	return fmt.Sprintf("%s:%d:%d:%d:%d", NotificationKindEventReminder, job.EventID, job.SourceRevision, job.ReminderOffsetMinutes, userID)
}

func deliveryIDempotencyKey(notificationID string, channelCode string) string {
	return fmt.Sprintf("delivery:%s:%s", notificationID, channelCode)
}

func randomToken() (string, error) {
	var buffer [16]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer[:]), nil
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func marshalPayload(payload NotificationPayload) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать notification payload: %w", err)
	}
	return body, nil
}

func checkRowsAffected(result sql.Result, err error, action string) error {
	if err != nil {
		return fmt.Errorf("не удалось %s: %w", action, err)
	}
	rows, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return fmt.Errorf("не удалось подтвердить %s: %w", action, rowsErr)
	}
	if rows != 1 {
		return fmt.Errorf("lease reminder job потерян во время операции %q", action)
	}
	return nil
}
