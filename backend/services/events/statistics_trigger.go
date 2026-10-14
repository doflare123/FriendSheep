package events

import (
	"context"
	"errors"
	"time"
)

const (
	StatisticsTriggerSchemaVersion = 1
	DefaultStatisticsTriggerLimit  = 100
	MaxStatisticsTriggerLimit      = 500
)

var (
	ErrInvalidStatisticsTriggerCursor = errors.New("некорректный cursor statistics trigger")
	ErrInvalidStatisticsTriggerLimit  = errors.New("некорректный limit statistics trigger")
	errStatisticsTriggerUnavailable  = errors.New("хранилище statistics trigger не настроено")
)

type StatisticsTriggerMessage struct {
	Sequence      int64     `json:"sequence"`
	SchemaVersion int       `json:"schemaVersion"`
	EventID       uint      `json:"eventId"`
	CompletedAt   time.Time `json:"completedAt"`
}

type StatisticsTriggersPage struct {
	Messages   []StatisticsTriggerMessage `json:"messages"`
	NextCursor int64                      `json:"nextCursor"`
	HasMore    bool                       `json:"hasMore"`
}

type StatisticsBackfillResult struct {
	NextEventID uint `json:"nextEventId"`
	Scanned     int  `json:"scanned"`
	Published   int  `json:"published"`
	HasMore     bool `json:"hasMore"`
}

type StatisticsTriggerReader interface {
	ListStatisticsTriggers(ctx context.Context, after int64, limit int) ([]StatisticsTriggerMessage, error)
}

type StatisticsTriggerBackfiller interface {
	BackfillStatisticsTriggers(ctx context.Context, afterEventID uint, limit int) (StatisticsBackfillResult, error)
}

type StatisticsTriggerService interface {
	List(ctx context.Context, after int64, limit int) (StatisticsTriggersPage, error)
	Backfill(ctx context.Context, afterEventID uint, limit int) (StatisticsBackfillResult, error)
}

type statisticsTriggerService struct {
	reader StatisticsTriggerReader
	backfiller StatisticsTriggerBackfiller
}

func NewStatisticsTriggerService(reader StatisticsTriggerReader, backfiller StatisticsTriggerBackfiller) StatisticsTriggerService {
	return &statisticsTriggerService{reader: reader, backfiller: backfiller}
}

func statisticsTriggerLimit(limit int) (int, error) {
	if limit == 0 { return DefaultStatisticsTriggerLimit, nil }
	if limit < 1 || limit > MaxStatisticsTriggerLimit { return 0, ErrInvalidStatisticsTriggerLimit }
	return limit, nil
}

func (s *statisticsTriggerService) List(ctx context.Context, after int64, limit int) (StatisticsTriggersPage, error) {
	if after < 0 { return StatisticsTriggersPage{}, ErrInvalidStatisticsTriggerCursor }
	limit, err := statisticsTriggerLimit(limit)
	if err != nil { return StatisticsTriggersPage{}, err }
	if s.reader == nil { return StatisticsTriggersPage{}, errStatisticsTriggerUnavailable }
	messages, err := s.reader.ListStatisticsTriggers(ctx, after, limit+1)
	if err != nil { return StatisticsTriggersPage{}, err }
	hasMore := len(messages) > limit
	if hasMore { messages = messages[:limit] }
	nextCursor := after
	if len(messages) > 0 { nextCursor = messages[len(messages)-1].Sequence }
	if messages == nil { messages = []StatisticsTriggerMessage{} }
	return StatisticsTriggersPage{Messages: messages, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s *statisticsTriggerService) Backfill(ctx context.Context, afterEventID uint, limit int) (StatisticsBackfillResult, error) {
	limit, err := statisticsTriggerLimit(limit)
	if err != nil { return StatisticsBackfillResult{}, err }
	if s.backfiller == nil { return StatisticsBackfillResult{}, errStatisticsTriggerUnavailable }
	return s.backfiller.BackfillStatisticsTriggers(ctx, afterEventID, limit)
}
