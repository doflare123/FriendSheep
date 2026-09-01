package lifecycle

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) PingContext(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("lifecycle store не инициализирован")
	}
	return s.db.PingContext(ctx)
}

func (s *Store) LoadCursor(ctx context.Context, source string) (uint64, error) {
	if err := s.ensureCursor(ctx, source); err != nil {
		return 0, err
	}
	var lastSequence int64
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT last_sequence FROM notify_service.event_lifecycle_source_cursors WHERE source_name = $1`,
		source,
	).Scan(&lastSequence); err != nil {
		return 0, fmt.Errorf("не удалось прочитать lifecycle cursor: %w", err)
	}
	return uint64(lastSequence), nil
}

func (s *Store) ApplySourceEvents(ctx context.Context, source string, items []ScheduleEvent, now time.Time) (ApplyResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось начать транзакцию lifecycle ingestion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := ensureCursorTx(ctx, tx, source); err != nil {
		return ApplyResult{}, err
	}

	var lastSequence int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT last_sequence FROM notify_service.event_lifecycle_source_cursors WHERE source_name = $1 FOR UPDATE`,
		source,
	).Scan(&lastSequence); err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось заблокировать lifecycle cursor: %w", err)
	}

	applied := 0
	current := uint64(lastSequence)
	if err := validateSourceBatch(items); err != nil {
		return ApplyResult{}, err
	}
	for _, item := range items {
		if item.Sequence <= current {
			continue
		}
		seen, err := recordSourceMessageTx(ctx, tx, source, item)
		if err != nil {
			return ApplyResult{}, err
		}
		if seen {
			current = item.Sequence
			continue
		}
		if err := applySourceEventTx(ctx, tx, item, now); err != nil {
			return ApplyResult{}, err
		}
		current = item.Sequence
		applied++
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE notify_service.event_lifecycle_source_cursors
		 SET last_sequence = $2, updated_at = $3
		 WHERE source_name = $1`,
		source,
		int64(current),
		now,
	); err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось обновить lifecycle cursor: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось зафиксировать lifecycle ingestion: %w", err)
	}
	return ApplyResult{LastSequence: current, AppliedCount: applied}, nil
}

func applySourceEventTx(ctx context.Context, tx *sql.Tx, item ScheduleEvent, now time.Time) error {
	existing, found, err := loadJobForUpdate(ctx, tx, item.EventID)
	if err != nil {
		return err
	}

	if !found {
		return insertJobTx(ctx, tx, item, now)
	}
	if item.Sequence <= existing.SourceSequence {
		return nil
	}

	switch item.Operation {
	case OperationScheduleUpsert:
		if item.StartTime == nil || item.EndTime == nil {
			return fmt.Errorf("schedule_upsert без времени для event %d", item.EventID)
		}
		result, execErr := tx.ExecContext(
			ctx,
			`UPDATE notify_service.event_lifecycle_jobs
			 SET source_sequence = $2,
			     source_message_id = $3,
			     start_time = $4,
			     end_time = $5,
			     state = $6,
			     next_action_at = $4,
			     attempt_count = 0,
			     next_attempt_at = NULL,
			     lease_until = NULL,
			     lease_token = NULL,
			     last_error_code = NULL,
			     updated_at = $7
			 WHERE id = $1`,
			existing.ID,
			int64(item.Sequence),
			item.MessageID,
			*item.StartTime,
			*item.EndTime,
			StateScheduled,
			now,
		)
		err = checkRowsAffected(result, execErr, "применить lifecycle upsert")
	case OperationScheduleCancel:
		result, execErr := tx.ExecContext(
			ctx,
			`UPDATE notify_service.event_lifecycle_jobs
			 SET source_sequence = $2,
			     source_message_id = $3,
			     start_time = NULL,
			     end_time = NULL,
			     state = $4,
			     next_action_at = NULL,
			     attempt_count = 0,
			     next_attempt_at = NULL,
			     lease_until = NULL,
			     lease_token = NULL,
			     last_error_code = NULL,
			     updated_at = $5
			 WHERE id = $1`,
			existing.ID,
			int64(item.Sequence),
			item.MessageID,
			StateCancelled,
			now,
		)
		err = checkRowsAffected(result, execErr, "применить lifecycle cancel")
	default:
		return fmt.Errorf("неизвестная lifecycle операция %q", item.Operation)
	}
	if err != nil {
		return fmt.Errorf("не удалось применить lifecycle source event: %w", err)
	}
	return nil
}

func insertJobTx(ctx context.Context, tx *sql.Tx, item ScheduleEvent, now time.Time) error {
	switch item.Operation {
	case OperationScheduleUpsert:
		if item.StartTime == nil || item.EndTime == nil {
			return fmt.Errorf("schedule_upsert без времени для event %d", item.EventID)
		}
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO notify_service.event_lifecycle_jobs (
				event_id, source_sequence, source_message_id, start_time, end_time, state,
				next_action_at, attempt_count, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $4, 0, $7, $7)`,
			int64(item.EventID),
			int64(item.Sequence),
			item.MessageID,
			*item.StartTime,
			*item.EndTime,
			StateScheduled,
			now,
		)
		if err != nil {
			return fmt.Errorf("не удалось вставить lifecycle job: %w", err)
		}
	case OperationScheduleCancel:
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO notify_service.event_lifecycle_jobs (
				event_id, source_sequence, source_message_id, state, attempt_count, created_at, updated_at
			) VALUES ($1, $2, $3, $4, 0, $5, $5)`,
			int64(item.EventID),
			int64(item.Sequence),
			item.MessageID,
			StateCancelled,
			now,
		)
		if err != nil {
			return fmt.Errorf("не удалось вставить lifecycle tombstone: %w", err)
		}
	default:
		return fmt.Errorf("неизвестная lifecycle операция %q", item.Operation)
	}
	return nil
}

func (s *Store) ClaimDueJob(ctx context.Context, now time.Time, leaseDuration time.Duration) (*Job, error) {
	if leaseDuration <= 0 {
		return nil, errors.New("lease duration должен быть положительным")
	}
	leaseUntil := now.Add(leaseDuration)
	leaseToken, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать lease token: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию claim lifecycle job: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id
			FROM notify_service.event_lifecycle_jobs
			WHERE state IN ($1, $2, $3, $4)
			  AND (lease_until IS NULL OR lease_until <= $5)
			  AND (
			       (state IN ($1, $2, $4) AND next_action_at IS NOT NULL AND next_action_at <= $5)
			    OR (state = $3 AND next_attempt_at IS NOT NULL AND next_attempt_at <= $5)
			  )
			ORDER BY COALESCE(next_attempt_at, next_action_at), id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE notify_service.event_lifecycle_jobs job
		SET state = $6,
		    lease_until = $7,
		    lease_token = $8,
		    updated_at = $5
		FROM candidate
		WHERE job.id = candidate.id
		RETURNING job.id, job.event_id, job.source_sequence, job.source_message_id, job.start_time, job.end_time,
		          job.state, job.next_action_at, job.attempt_count, job.next_attempt_at, job.lease_until,
		          job.lease_token, COALESCE(job.last_error_code, '')
	`, StateScheduled, StateActiveWait, StateRetryWait, StateProcessing, now, StateProcessing, leaseUntil, leaseToken)

	job, found, err := scanClaimedJob(row)
	if err != nil {
		return nil, err
	}
	if !found {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("не удалось завершить пустой claim lifecycle job: %w", err)
		}
		return nil, nil
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось зафиксировать claim lifecycle job: %w", err)
	}
	return job, nil
}

func scanClaimedJob(scanner interface{ Scan(...any) error }) (*Job, bool, error) {
	var (
		job            Job
		eventID        int64
		sourceSequence int64
		startTime      sql.NullTime
		endTime        sql.NullTime
		nextActionAt   sql.NullTime
		nextAttemptAt  sql.NullTime
		leaseUntil     sql.NullTime
		lastErrorCode  string
	)
	err := scanner.Scan(
		&job.ID,
		&eventID,
		&sourceSequence,
		&job.SourceMessageID,
		&startTime,
		&endTime,
		&job.State,
		&nextActionAt,
		&job.AttemptCount,
		&nextAttemptAt,
		&leaseUntil,
		&job.LeaseToken,
		&lastErrorCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("не удалось claim lifecycle job: %w", err)
	}
	job.EventID = uint64(eventID)
	job.SourceSequence = uint64(sourceSequence)
	job.StartTime = nullTimePointer(startTime)
	job.EndTime = nullTimePointer(endTime)
	job.NextActionAt = nullTimePointer(nextActionAt)
	job.NextAttemptAt = nullTimePointer(nextAttemptAt)
	job.LeaseUntil = nullTimePointer(leaseUntil)
	job.LastErrorCode = lastErrorCode
	return &job, true, nil
}

func (s *Store) RescheduleAfterSuccess(ctx context.Context, job Job, result AdvanceResult, now time.Time) error {
	state, nextActionAt, err := transitionAfterSuccess(job, result, now)
	if err != nil {
		return err
	}
	execResult, execErr := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.event_lifecycle_jobs
		 SET state = $4,
		     start_time = $5,
		     end_time = $6,
		     next_action_at = $7,
		     attempt_count = 0,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULL,
		     updated_at = $8
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		job.ID,
		job.LeaseToken,
		StateProcessing,
		state,
		result.StartTime,
		result.EndTime,
		nextActionAt,
		now,
	)
	return checkRowsAffected(execResult, execErr, "завершить lifecycle job после успеха")
}

func (s *Store) RescheduleRetry(ctx context.Context, job Job, nextAttempt time.Time, errorCode string, now time.Time) error {
	execResult, execErr := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.event_lifecycle_jobs
		 SET state = $4,
		     attempt_count = attempt_count + 1,
		     next_attempt_at = $5,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = $6,
		     updated_at = $7
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		job.ID,
		job.LeaseToken,
		StateProcessing,
		StateRetryWait,
		nextAttempt,
		errorCode,
		now,
	)
	return checkRowsAffected(execResult, execErr, "перевести lifecycle job в retry_wait")
}

func (s *Store) MarkTerminal(ctx context.Context, job Job, state string, errorCode string, now time.Time) error {
	execResult, execErr := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.event_lifecycle_jobs
		 SET state = $4,
		     next_action_at = NULL,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = $5,
		     updated_at = $6
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		job.ID,
		job.LeaseToken,
		StateProcessing,
		state,
		errorCode,
		now,
	)
	return checkRowsAffected(execResult, execErr, "завершить lifecycle job терминально")
}

func transitionAfterSuccess(job Job, result AdvanceResult, now time.Time) (string, *time.Time, error) {
	switch result.Outcome {
	case OutcomeStarted, OutcomeAlreadyActive:
		next := result.EndTime
		return StateActiveWait, &next, nil
	case OutcomeCompleted, OutcomeCompletedCatchUp, OutcomeAlreadyCompleted:
		return StateCompleted, nil, nil
	case OutcomeNotDue:
		next := result.StartTime
		if !next.After(now) {
			next = now.Add(time.Second)
		}
		return StateScheduled, &next, nil
	default:
		return "", nil, fmt.Errorf("неизвестный lifecycle outcome %q", result.Outcome)
	}
}

func loadJobForUpdate(ctx context.Context, tx *sql.Tx, eventID uint64) (Job, bool, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, event_id, source_sequence, source_message_id, start_time, end_time, state, next_action_at,
		        attempt_count, next_attempt_at, lease_until, COALESCE(lease_token::text, ''), COALESCE(last_error_code, '')
		 FROM notify_service.event_lifecycle_jobs
		 WHERE event_id = $1
		 FOR UPDATE`,
		int64(eventID),
	)
	var (
		job           Job
		rawEventID    int64
		rawSequence   int64
		startTime     sql.NullTime
		endTime       sql.NullTime
		nextActionAt  sql.NullTime
		nextAttemptAt sql.NullTime
		leaseUntil    sql.NullTime
	)
	err := row.Scan(
		&job.ID,
		&rawEventID,
		&rawSequence,
		&job.SourceMessageID,
		&startTime,
		&endTime,
		&job.State,
		&nextActionAt,
		&job.AttemptCount,
		&nextAttemptAt,
		&leaseUntil,
		&job.LeaseToken,
		&job.LastErrorCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, fmt.Errorf("не удалось заблокировать lifecycle job: %w", err)
	}
	job.EventID = uint64(rawEventID)
	job.SourceSequence = uint64(rawSequence)
	job.StartTime = nullTimePointer(startTime)
	job.EndTime = nullTimePointer(endTime)
	job.NextActionAt = nullTimePointer(nextActionAt)
	job.NextAttemptAt = nullTimePointer(nextAttemptAt)
	job.LeaseUntil = nullTimePointer(leaseUntil)
	return job, true, nil
}

func ensureCursorTx(ctx context.Context, tx *sql.Tx, source string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_lifecycle_source_cursors (source_name, last_sequence)
		 VALUES ($1, 0)
		 ON CONFLICT (source_name) DO NOTHING`,
		source,
	)
	if err != nil {
		return fmt.Errorf("не удалось подготовить lifecycle cursor: %w", err)
	}
	return nil
}

func recordSourceMessageTx(ctx context.Context, tx *sql.Tx, source string, item ScheduleEvent) (bool, error) {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_lifecycle_source_messages (source_name, message_id, sequence)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (source_name, message_id) DO NOTHING`,
		source,
		item.MessageID,
		int64(item.Sequence),
	)
	if err != nil {
		return false, fmt.Errorf("не удалось сохранить lifecycle source receipt: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("не удалось проверить lifecycle source receipt: %w", err)
	}
	if rows > 0 {
		return false, nil
	}

	var existingSequence int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT sequence FROM notify_service.event_lifecycle_source_messages
		 WHERE source_name = $1 AND message_id = $2`,
		source,
		item.MessageID,
	).Scan(&existingSequence); err != nil {
		return false, fmt.Errorf("не удалось загрузить lifecycle source receipt: %w", err)
	}
	if uint64(existingSequence) != item.Sequence {
		return false, fmt.Errorf("message_id %s повторно использован с другой sequence", item.MessageID)
	}
	return true, nil
}

func validateSourceBatch(items []ScheduleEvent) error {
	const maxDatabaseBigInt = uint64(1<<63 - 1)
	var previous uint64
	for _, item := range items {
		if item.Sequence == 0 {
			return errors.New("lifecycle batch содержит sequence=0")
		}
		if item.Sequence > maxDatabaseBigInt {
			return errors.New("lifecycle batch содержит sequence вне диапазона PostgreSQL BIGINT")
		}
		if item.Sequence <= previous {
			return fmt.Errorf("lifecycle batch нарушает строгий порядок sequence: %d после %d", item.Sequence, previous)
		}
		if item.SchemaVersion != 1 {
			return fmt.Errorf("неподдерживаемая schemaVersion=%d", item.SchemaVersion)
		}
		if item.EventID == 0 {
			return errors.New("lifecycle batch содержит event_id=0")
		}
		if item.EventID > maxDatabaseBigInt {
			return errors.New("lifecycle batch содержит event_id вне диапазона PostgreSQL BIGINT")
		}
		if item.MessageID == "" {
			return errors.New("lifecycle batch содержит пустой message_id")
		}
		if item.OccurredAt.IsZero() {
			return errors.New("lifecycle batch содержит пустой occurred_at")
		}
		switch item.Operation {
		case OperationScheduleUpsert:
			if item.StartTime == nil || item.EndTime == nil || !item.EndTime.After(*item.StartTime) {
				return fmt.Errorf("некорректный lifecycle upsert для event %d", item.EventID)
			}
		case OperationScheduleCancel:
			if item.StartTime != nil || item.EndTime != nil {
				return fmt.Errorf("cancel lifecycle сообщения для event %d не должно содержать время", item.EventID)
			}
		default:
			return fmt.Errorf("неизвестная lifecycle операция %q", item.Operation)
		}
		previous = item.Sequence
	}
	return nil
}

func (s *Store) ensureCursor(ctx context.Context, source string) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_lifecycle_source_cursors (source_name, last_sequence)
		 VALUES ($1, 0)
		 ON CONFLICT (source_name) DO NOTHING`,
		source,
	)
	if err != nil {
		return fmt.Errorf("не удалось подготовить lifecycle cursor: %w", err)
	}
	return nil
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func randomToken() (string, error) {
	var buffer [16]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer[:]), nil
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
		return fmt.Errorf("lease lifecycle job потерян во время операции %q", action)
	}
	return nil
}
