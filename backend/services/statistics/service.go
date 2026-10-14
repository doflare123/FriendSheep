package statistics

import (
	"context"
	"errors"
	"strconv"
	"time"
)

var ErrInvalidInput = errors.New("некорректный запрос статистики")
var ErrEventMissing = errors.New("мероприятие не найдено")

type CountItem struct {
	ID    uint  `json:"id"`
	Count int64 `json:"count"`
}
type MonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}
type WeekdayCount struct {
	Weekday int   `json:"weekday"`
	Count   int64 `json:"count"`
}
type View struct {
	CompletedEvents        int64          `json:"completedEvents"`
	PlannedMinutes         int64          `json:"plannedMinutes"`
	CreatedCompletedEvents int64          `json:"createdCompletedEvents"`
	UniquePeopleCount      int64          `json:"uniquePeopleCount"`
	Types                  []CountItem    `json:"types"`
	Genres                 []CountItem    `json:"genres"`
	Months                 []MonthCount   `json:"months"`
	Weekdays               []WeekdayCount `json:"weekdays"`
	MostFrequentTypeID     *uint          `json:"mostFrequentTypeId"`
	MostFrequentGenreID    *uint          `json:"mostFrequentGenreId"`
	MostFrequentWeekday    *int           `json:"mostFrequentWeekday"`
	Timezone               string         `json:"timezone"`
	LastUpdatedAt          *time.Time     `json:"lastUpdatedAt"`
}
type ProcessResult struct {
	EventID     uint       `json:"eventId"`
	Outcome     string     `json:"outcome"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
}
type Snapshot struct {
	EventID, CreatorID, TypeID uint
	StartTime, EndTime         time.Time
	DurationMinutes            int64
	Completed                  bool
	Participants, Genres       []uint
	ProcessedAt                *time.Time
}
type Contribution struct {
	Snapshot
	Month   string
	Weekday int
}
type Transaction interface {
	LoadEvent(context.Context, uint) (Snapshot, error)
	Apply(context.Context, Contribution, time.Time) (bool, error)
}
type Store interface {
	WithinTransaction(context.Context, func(Transaction) error) error
	Read(context.Context, uint) (View, error)
}
type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service { return &Service{store: store, now: time.Now} }
func (s *Service) Read(ctx context.Context, userID uint) (View, error) {
	if userID == 0 {
		return View{}, ErrInvalidInput
	}
	return s.store.Read(ctx, userID)
}
func (s *Service) Process(ctx context.Context, eventID uint) (ProcessResult, error) {
	if eventID == 0 {
		return ProcessResult{}, ErrInvalidInput
	}
	result := ProcessResult{EventID: eventID}
	err := s.store.WithinTransaction(ctx, func(tx Transaction) error {
		snapshot, err := tx.LoadEvent(ctx, eventID)
		if errors.Is(err, ErrEventMissing) {
			result.Outcome = "missing"
			return nil
		}
		if err != nil {
			return err
		}
		if snapshot.ProcessedAt != nil {
			result.Outcome = "already_processed"
			result.ProcessedAt = snapshot.ProcessedAt
			return nil
		}
		now := s.now().UTC()
		if !snapshot.Completed || now.Before(snapshot.EndTime) {
			result.Outcome = "not_completed"
			return nil
		}
		if snapshot.DurationMinutes <= 0 || !snapshot.EndTime.After(snapshot.StartTime) || snapshot.CreatorID == 0 || snapshot.TypeID == 0 {
			return ErrInvalidInput
		}
		start := snapshot.StartTime.UTC()
		weekday := int(start.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		applied, err := tx.Apply(ctx, Contribution{Snapshot: snapshot, Month: start.Format("2006-01"), Weekday: weekday}, now)
		if err != nil {
			return err
		}
		result.Outcome = "processed"
		if !applied {
			result.Outcome = "already_processed"
		}
		result.ProcessedAt = &now
		return nil
	})
	if err != nil {
		return ProcessResult{}, err
	}
	return result, nil
}
func emptyView() View {
	return View{Types: []CountItem{}, Genres: []CountItem{}, Months: []MonthCount{}, Weekdays: []WeekdayCount{}, Timezone: "UTC"}
}
func numericKey(id uint) string { return strconv.FormatUint(uint64(id), 10) }
