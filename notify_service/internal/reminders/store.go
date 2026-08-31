package reminders

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const inboxPageLimitMax = 100

type Store struct {
	db       *sql.DB
	channels *ChannelRegistry
}

type txDeliveryWriter struct {
	tx *sql.Tx
}

func NewStore(db *sql.DB, channels *ChannelRegistry) *Store {
	return &Store{db: db, channels: channels}
}

func (s *Store) PingContext(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("reminder store не инициализирован")
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
		`SELECT last_sequence FROM notify_service.event_reminder_source_cursors WHERE source_name = $1`,
		source,
	).Scan(&lastSequence); err != nil {
		return 0, fmt.Errorf("не удалось прочитать reminder cursor: %w", err)
	}
	return uint64(lastSequence), nil
}

func (s *Store) ApplySourceEvents(ctx context.Context, source string, items []ReminderIntent, now time.Time) (ApplyResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось начать транзакцию reminder ingestion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := ensureCursorTx(ctx, tx, source); err != nil {
		return ApplyResult{}, err
	}

	var lastSequence int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT last_sequence FROM notify_service.event_reminder_source_cursors WHERE source_name = $1 FOR UPDATE`,
		source,
	).Scan(&lastSequence); err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось заблокировать reminder cursor: %w", err)
	}

	if err := validateSourceBatch(items); err != nil {
		return ApplyResult{}, err
	}

	current := uint64(lastSequence)
	applied := 0
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
		`UPDATE notify_service.event_reminder_source_cursors
		 SET last_sequence = $2, updated_at = $3
		 WHERE source_name = $1`,
		source,
		int64(current),
		now,
	); err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось обновить reminder cursor: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ApplyResult{}, fmt.Errorf("не удалось зафиксировать reminder ingestion: %w", err)
	}
	return ApplyResult{LastSequence: current, AppliedCount: applied}, nil
}

func (s *Store) ClaimDueJob(ctx context.Context, now time.Time, leaseDuration time.Duration) (*Job, error) {
	if leaseDuration <= 0 {
		return nil, errors.New("lease duration должен быть положительным")
	}
	leaseUntil := now.Add(leaseDuration)
	leaseToken, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать reminder lease token: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию claim reminder job: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id
			FROM notify_service.event_reminder_jobs
			WHERE state IN ($1, $2, $3)
			  AND (lease_until IS NULL OR lease_until <= $4)
			  AND (
			       (state = $1 AND due_at IS NOT NULL AND due_at <= $4)
			    OR (state = $2 AND next_attempt_at IS NOT NULL AND next_attempt_at <= $4)
			    OR (state = $3 AND lease_until IS NOT NULL AND lease_until <= $4)
			  )
			ORDER BY COALESCE(next_attempt_at, due_at), id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE notify_service.event_reminder_jobs job
		SET state = $3,
		    lease_until = $5,
		    lease_token = $6,
		    updated_at = $4
		FROM candidate
		WHERE job.id = candidate.id
		RETURNING job.id, job.event_id, job.source_revision, job.source_message_id, job.reminder_offset_minutes,
		          job.start_time, job.due_at, job.window_end_at, job.state, job.attempt_count, job.next_attempt_at,
		          job.lease_until, COALESCE(job.lease_token, ''), COALESCE(job.last_error_code, '')
	`, StateScheduled, StateRetryWait, StateProcessing, now, leaseUntil, leaseToken)

	job, found, err := scanClaimedJobRow(row)
	if err != nil {
		return nil, err
	}
	if !found {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("не удалось завершить пустой claim reminder job: %w", err)
		}
		return nil, nil
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось зафиксировать claim reminder job: %w", err)
	}
	return job, nil
}

func (s *Store) CompleteJob(ctx context.Context, job Job, snapshot RecipientSnapshot, now time.Time) (int, error) {
	if s == nil || s.channels == nil {
		return 0, errors.New("reminder store channels не инициализирован")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("не удалось начать транзакцию reminder delivery: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	lockedJob, found, err := loadJobByLeaseForUpdate(ctx, tx, job.ID, job.LeaseToken)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("lease reminder job потерян до доставки")
	}
	if lockedJob.SourceRevision != job.SourceRevision || lockedJob.EventID != job.EventID || lockedJob.ReminderOffsetMinutes != job.ReminderOffsetMinutes {
		return 0, fmt.Errorf("reminder job изменился до доставки")
	}

	seenUsers := make(map[uint64]struct{}, len(snapshot.Recipients))
	deliveredUsers := 0
	for _, recipient := range snapshot.Recipients {
		if recipient.UserID == 0 {
			return 0, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: errors.New("recipient.userId=0")}
		}
		if _, exists := seenUsers[recipient.UserID]; exists {
			continue
		}
		resolvedChannels, err := s.channels.Resolve(recipient.Channels)
		if err != nil {
			return 0, err
		}
		if len(resolvedChannels) == 0 {
			continue
		}

		notification, err := insertNotificationTx(ctx, tx, job, snapshot, recipient.UserID, now)
		if err != nil {
			return 0, err
		}
		writer := txDeliveryWriter{tx: tx}
		for _, channel := range resolvedChannels {
			result, err := channel.Deliver(ctx, notification, writer, now)
			if err != nil {
				return 0, err
			}
			if result.Status != DeliveryStatusDelivered {
				return 0, &ClientError{Code: ErrorCodeUnexpectedStatus, Retryable: true, Cause: fmt.Errorf("channel %s вернул неподдерживаемый status %q", channel.Code(), result.Status)}
			}
		}
		seenUsers[recipient.UserID] = struct{}{}
		deliveredUsers++
	}

	result, execErr := tx.ExecContext(
		ctx,
		`UPDATE notify_service.event_reminder_jobs
		 SET state = $4,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULL,
		     updated_at = $5
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		job.ID,
		job.LeaseToken,
		StateProcessing,
		StateCompleted,
		now,
	)
	if err := checkRowsAffected(result, execErr, "завершить reminder job после доставки"); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("не удалось зафиксировать reminder delivery: %w", err)
	}
	return deliveredUsers, nil
}

func (s *Store) RescheduleRetry(ctx context.Context, job Job, nextAttempt time.Time, errorCode string, now time.Time) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.event_reminder_jobs
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
	return checkRowsAffected(result, err, "перевести reminder job в retry_wait")
}

func (s *Store) MarkFinal(ctx context.Context, job Job, state string, errorCode string, now time.Time) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.event_reminder_jobs
		 SET state = $4,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULLIF($5, ''),
		     updated_at = $6
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		job.ID,
		job.LeaseToken,
		StateProcessing,
		state,
		errorCode,
		now,
	)
	return checkRowsAffected(result, err, "завершить reminder job терминально")
}

func (s *Store) ListNotifications(ctx context.Context, userID uint64, after string, limit int, unreadOnly bool) (NotificationPage, error) {
	if userID == 0 {
		return NotificationPage{}, &ClientError{Code: ErrorCodeInvalidUserID, Terminal: true}
	}
	if limit < 1 || limit > inboxPageLimitMax {
		return NotificationPage{}, &ClientError{Code: ErrorCodeInvalidLimit, Terminal: true}
	}

	var (
		anchorCreatedAt sql.NullTime
		anchorID        string
	)
	if after != "" {
		err := s.db.QueryRowContext(
			ctx,
			`SELECT created_at, id
			 FROM notify_service.notifications
			 WHERE id = $1 AND user_id = $2`,
			after,
			int64(userID),
		).Scan(&anchorCreatedAt, &anchorID)
		if errors.Is(err, sql.ErrNoRows) {
			return NotificationPage{}, &ClientError{Code: ErrorCodeInvalidCursor, Terminal: true}
		}
		if err != nil {
			return NotificationPage{}, fmt.Errorf("не удалось загрузить cursor inbox: %w", err)
		}
	}

	query := `
		SELECT id, user_id, kind, resource_type, resource_id, reminder_offset_minutes, schedule_revision, payload, created_at, read_at
		FROM notify_service.notifications
		WHERE user_id = $1`
	args := []any{int64(userID)}
	nextArg := 2
	if unreadOnly {
		query += ` AND read_at IS NULL`
	}
	if after != "" {
		query += fmt.Sprintf(` AND (created_at, id) < ($%d, $%d)`, nextArg, nextArg+1)
		args = append(args, anchorCreatedAt.Time, anchorID)
		nextArg += 2
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, nextArg)
	args = append(args, limit+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return NotificationPage{}, fmt.Errorf("не удалось загрузить inbox notifications: %w", err)
	}
	defer rows.Close()

	records := make([]NotificationRecord, 0, limit+1)
	for rows.Next() {
		record, err := scanNotificationRecord(rows)
		if err != nil {
			return NotificationPage{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return NotificationPage{}, fmt.Errorf("не удалось прочитать inbox notifications: %w", err)
	}

	page := NotificationPage{}
	if len(records) > limit {
		page.HasMore = true
		page.NextCursor = records[limit-1].ID
		records = records[:limit]
	}
	page.Items = records
	return page, nil
}

func (s *Store) CountUnread(ctx context.Context, userID uint64) (int, error) {
	if userID == 0 {
		return 0, &ClientError{Code: ErrorCodeInvalidUserID, Terminal: true}
	}
	var count int
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		 FROM notify_service.notifications
		 WHERE user_id = $1 AND read_at IS NULL`,
		int64(userID),
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("не удалось посчитать unread notifications: %w", err)
	}
	return count, nil
}

func (s *Store) MarkAsRead(ctx context.Context, userID uint64, notificationID string, now time.Time) (NotificationRecord, error) {
	if userID == 0 {
		return NotificationRecord{}, &ClientError{Code: ErrorCodeInvalidUserID, Terminal: true}
	}
	if notificationID == "" {
		return NotificationRecord{}, &ClientError{Code: ErrorCodeNotificationNotFound, Terminal: true}
	}
	row := s.db.QueryRowContext(
		ctx,
		`UPDATE notify_service.notifications
		 SET read_at = COALESCE(read_at, $3),
		     updated_at = CASE WHEN read_at IS NULL THEN $3 ELSE updated_at END
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, kind, resource_type, resource_id, reminder_offset_minutes, schedule_revision, payload, created_at, read_at`,
		notificationID,
		int64(userID),
		now,
	)
	record, found, err := scanNotificationRecordRow(row)
	if err != nil {
		return NotificationRecord{}, err
	}
	if !found {
		return NotificationRecord{}, &ClientError{Code: ErrorCodeNotificationNotFound, Terminal: true}
	}
	return record, nil
}

func applySourceEventTx(ctx context.Context, tx *sql.Tx, item ReminderIntent, now time.Time) error {
	switch item.Operation {
	case OperationScheduleUpsert:
		if item.StartTime == nil {
			return fmt.Errorf("schedule_upsert без startTime для event %d", item.EventID)
		}
		occurrences, err := buildOccurrences(*item.StartTime, item.ReminderOffsetMinutes)
		if err != nil {
			return err
		}
		if err := supersedeOlderJobsTx(ctx, tx, item.EventID, item.Sequence, StateSuperseded, now); err != nil {
			return err
		}
		for _, occurrence := range occurrences {
			if err := upsertJobTx(ctx, tx, item, occurrence, now); err != nil {
				return err
			}
		}
		return nil
	case OperationScheduleCancel:
		if err := supersedeOlderJobsTx(ctx, tx, item.EventID, item.Sequence+1, StateCancelled, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO notify_service.event_reminder_jobs (
				event_id, source_revision, source_message_id, reminder_offset_minutes, start_time, due_at, window_end_at,
				state, attempt_count, created_at, updated_at
			) VALUES ($1, $2, $3, 0, NULL, NULL, NULL, $4, 0, $5, $5)
			ON CONFLICT (event_id, source_revision, reminder_offset_minutes)
			DO UPDATE SET state = EXCLUDED.state, source_message_id = EXCLUDED.source_message_id, updated_at = EXCLUDED.updated_at,
			             next_attempt_at = NULL, lease_until = NULL, lease_token = NULL, last_error_code = NULL`,
			int64(item.EventID),
			int64(item.Sequence),
			item.MessageID,
			StateCancelled,
			now,
		)
		if err != nil {
			return fmt.Errorf("не удалось сохранить reminder cancel tombstone: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("неизвестная reminder операция %q", item.Operation)
	}
}

func upsertJobTx(ctx context.Context, tx *sql.Tx, item ReminderIntent, occurrence occurrence, now time.Time) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_reminder_jobs (
			event_id, source_revision, source_message_id, reminder_offset_minutes, start_time, due_at, window_end_at,
			state, attempt_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, $9, $9)
		ON CONFLICT (event_id, source_revision, reminder_offset_minutes)
		DO UPDATE SET source_message_id = EXCLUDED.source_message_id,
		             start_time = EXCLUDED.start_time,
		             due_at = EXCLUDED.due_at,
		             window_end_at = EXCLUDED.window_end_at,
		             state = EXCLUDED.state,
		             attempt_count = 0,
		             next_attempt_at = NULL,
		             lease_until = NULL,
		             lease_token = NULL,
		             last_error_code = NULL,
		             updated_at = EXCLUDED.updated_at`,
		int64(item.EventID),
		int64(item.Sequence),
		item.MessageID,
		occurrence.OffsetMinutes,
		*item.StartTime,
		occurrence.DueAt,
		occurrence.WindowEndAt,
		StateScheduled,
		now,
	)
	if err != nil {
		return fmt.Errorf("не удалось сохранить reminder job: %w", err)
	}
	return nil
}

func supersedeOlderJobsTx(ctx context.Context, tx *sql.Tx, eventID, sequence uint64, state string, now time.Time) error {
	_, err := tx.ExecContext(
		ctx,
		`UPDATE notify_service.event_reminder_jobs
		 SET state = $4,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULL,
		     updated_at = $5
		 WHERE event_id = $1
		   AND source_revision < $2
		   AND state IN ($3, $6, $7)`,
		int64(eventID),
		int64(sequence),
		StateScheduled,
		state,
		now,
		StateRetryWait,
		StateProcessing,
	)
	if err != nil {
		return fmt.Errorf("не удалось обновить старые reminder jobs: %w", err)
	}
	return nil
}

func insertNotificationTx(ctx context.Context, tx *sql.Tx, job Job, snapshot RecipientSnapshot, userID uint64, now time.Time) (NotificationRecord, error) {
	payload, err := marshalPayload(NotificationPayload{
		SchemaVersion: 1,
		EventID:       snapshot.EventID,
		Title:         snapshot.Title,
		StartTime:     snapshot.StartTime,
	})
	if err != nil {
		return NotificationRecord{}, err
	}
	idempotencyKey := notificationIDempotencyKey(job, userID)
	notificationID, err := randomToken()
	if err != nil {
		return NotificationRecord{}, fmt.Errorf("не удалось сгенерировать notification id: %w", err)
	}

	record := NotificationRecord{
		UserID:                userID,
		Kind:                  NotificationKindEventReminder,
		ResourceType:          ResourceTypeEvent,
		ResourceID:            snapshot.EventID,
		ReminderOffsetMinutes: job.ReminderOffsetMinutes,
		ScheduleRevision:      job.SourceRevision,
		Payload: NotificationPayload{
			SchemaVersion: 1,
			EventID:       snapshot.EventID,
			Title:         snapshot.Title,
			StartTime:     snapshot.StartTime,
		},
		CreatedAt: now,
	}
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO notify_service.notifications (
			id, user_id, kind, resource_type, resource_id, payload, reminder_offset_minutes, schedule_revision,
			idempotency_key, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $10)
		ON CONFLICT (idempotency_key)
		DO UPDATE SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING id`,
		notificationID,
		int64(userID),
		NotificationKindEventReminder,
		ResourceTypeEvent,
		int64(snapshot.EventID),
		string(payload),
		job.ReminderOffsetMinutes,
		int64(job.SourceRevision),
		idempotencyKey,
		now,
	).Scan(&record.ID)
	if err != nil {
		return NotificationRecord{}, fmt.Errorf("не удалось сохранить notification inbox record: %w", err)
	}
	return record, nil
}

func loadJobByLeaseForUpdate(ctx context.Context, tx *sql.Tx, jobID int64, leaseToken string) (Job, bool, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, event_id, source_revision, source_message_id, reminder_offset_minutes, start_time, due_at, window_end_at,
		        state, attempt_count, next_attempt_at, lease_until, COALESCE(lease_token, ''), COALESCE(last_error_code, '')
		 FROM notify_service.event_reminder_jobs
		 WHERE id = $1 AND lease_token = $2 AND state = $3
		 FOR UPDATE`,
		jobID,
		leaseToken,
		StateProcessing,
	)
	job, found, err := scanJobRow(row)
	if err != nil {
		return Job{}, false, err
	}
	return job, found, nil
}

func scanJobRow(scanner interface{ Scan(...any) error }) (Job, bool, error) {
	var (
		job            Job
		eventID        int64
		sourceRevision int64
		startTime      sql.NullTime
		dueAt          sql.NullTime
		windowEndAt    sql.NullTime
		nextAttemptAt  sql.NullTime
		leaseUntil     sql.NullTime
	)
	err := scanner.Scan(
		&job.ID,
		&eventID,
		&sourceRevision,
		&job.SourceMessageID,
		&job.ReminderOffsetMinutes,
		&startTime,
		&dueAt,
		&windowEndAt,
		&job.State,
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
		return Job{}, false, fmt.Errorf("не удалось загрузить reminder job: %w", err)
	}
	job.EventID = uint64(eventID)
	job.SourceRevision = uint64(sourceRevision)
	job.StartTime = nullTimePointer(startTime)
	job.DueAt = nullTimePointer(dueAt)
	job.WindowEndAt = nullTimePointer(windowEndAt)
	job.NextAttemptAt = nullTimePointer(nextAttemptAt)
	job.LeaseUntil = nullTimePointer(leaseUntil)
	return job, true, nil
}

func scanClaimedJobRow(scanner interface{ Scan(...any) error }) (*Job, bool, error) {
	job, found, err := scanJobRow(scanner)
	if err != nil || !found {
		return nil, found, err
	}
	return &job, true, nil
}

func validateSourceBatch(items []ReminderIntent) error {
	const maxDatabaseBigInt = uint64(1<<63 - 1)
	var previous uint64
	for _, item := range items {
		if item.Sequence == 0 {
			return errors.New("reminder batch содержит sequence=0")
		}
		if item.Sequence > maxDatabaseBigInt {
			return errors.New("reminder batch содержит sequence вне диапазона PostgreSQL BIGINT")
		}
		if item.Sequence <= previous {
			return fmt.Errorf("reminder batch нарушает строгий порядок sequence: %d после %d", item.Sequence, previous)
		}
		if item.SchemaVersion != 1 {
			return fmt.Errorf("неподдерживаемая reminder schemaVersion=%d", item.SchemaVersion)
		}
		if item.IntentType != IntentTypeEventReminder {
			return fmt.Errorf("неподдерживаемый intentType=%q", item.IntentType)
		}
		if item.EventID == 0 || item.EventID > maxDatabaseBigInt {
			return errors.New("reminder batch содержит некорректный eventId")
		}
		if !reminderMessageIDPattern.MatchString(item.MessageID) {
			return errors.New("reminder batch содержит некорректный messageId")
		}
		if item.OccurredAt.IsZero() {
			return errors.New("reminder batch содержит пустой occurredAt")
		}
		switch item.Operation {
		case OperationScheduleUpsert:
			if item.StartTime == nil {
				return fmt.Errorf("reminder upsert для event %d не содержит startTime", item.EventID)
			}
			if _, err := canonicalOffsets(item.ReminderOffsetMinutes); err != nil {
				return fmt.Errorf("reminder upsert для event %d невалиден: %w", item.EventID, err)
			}
		case OperationScheduleCancel:
			if item.StartTime != nil {
				return fmt.Errorf("reminder cancel для event %d не должен содержать startTime", item.EventID)
			}
			if len(item.ReminderOffsetMinutes) != 0 {
				return fmt.Errorf("reminder cancel для event %d не должен содержать reminderOffsetMinutes", item.EventID)
			}
		default:
			return fmt.Errorf("неизвестная reminder операция %q", item.Operation)
		}
		previous = item.Sequence
	}
	return nil
}

func recordSourceMessageTx(ctx context.Context, tx *sql.Tx, source string, item ReminderIntent) (bool, error) {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_reminder_source_messages (source_name, message_id, sequence)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (source_name, message_id) DO NOTHING`,
		source,
		item.MessageID,
		int64(item.Sequence),
	)
	if err != nil {
		return false, fmt.Errorf("не удалось сохранить reminder source receipt: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("не удалось проверить reminder source receipt: %w", err)
	}
	if rows > 0 {
		return false, nil
	}

	var existingSequence int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT sequence FROM notify_service.event_reminder_source_messages
		 WHERE source_name = $1 AND message_id = $2`,
		source,
		item.MessageID,
	).Scan(&existingSequence); err != nil {
		return false, fmt.Errorf("не удалось загрузить reminder source receipt: %w", err)
	}
	if uint64(existingSequence) != item.Sequence {
		return false, fmt.Errorf("messageId %s повторно использован с другой sequence", item.MessageID)
	}
	return true, nil
}

func ensureCursorTx(ctx context.Context, tx *sql.Tx, source string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_reminder_source_cursors (source_name, last_sequence)
		 VALUES ($1, 0)
		 ON CONFLICT (source_name) DO NOTHING`,
		source,
	)
	if err != nil {
		return fmt.Errorf("не удалось подготовить reminder cursor: %w", err)
	}
	return nil
}

func (s *Store) ensureCursor(ctx context.Context, source string) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO notify_service.event_reminder_source_cursors (source_name, last_sequence)
		 VALUES ($1, 0)
		 ON CONFLICT (source_name) DO NOTHING`,
		source,
	)
	if err != nil {
		return fmt.Errorf("не удалось подготовить reminder cursor: %w", err)
	}
	return nil
}

func scanNotificationRecord(scanner interface{ Scan(...any) error }) (NotificationRecord, error) {
	record, _, err := scanNotificationRecordRow(scanner)
	return record, err
}

func scanNotificationRecordRow(scanner interface{ Scan(...any) error }) (NotificationRecord, bool, error) {
	var (
		record           NotificationRecord
		userID           int64
		resourceID       int64
		scheduleRevision int64
		payloadBody      []byte
		readAt           sql.NullTime
	)
	err := scanner.Scan(
		&record.ID,
		&userID,
		&record.Kind,
		&record.ResourceType,
		&resourceID,
		&record.ReminderOffsetMinutes,
		&scheduleRevision,
		&payloadBody,
		&record.CreatedAt,
		&readAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return NotificationRecord{}, false, nil
	}
	if err != nil {
		return NotificationRecord{}, false, fmt.Errorf("не удалось прочитать notification record: %w", err)
	}
	record.UserID = uint64(userID)
	record.ResourceID = uint64(resourceID)
	record.ScheduleRevision = uint64(scheduleRevision)
	record.ReadAt = nullTimePointer(readAt)
	if err := json.Unmarshal(payloadBody, &record.Payload); err != nil {
		return NotificationRecord{}, false, fmt.Errorf("не удалось декодировать notification payload: %w", err)
	}
	return record, true, nil
}

func (w txDeliveryWriter) RecordDeliveredAttempt(ctx context.Context, notification NotificationRecord, channelCode string, result DeliveryResult, now time.Time) error {
	if result.AttemptNumber <= 0 {
		return errors.New("delivery attempt number должен быть положительным")
	}
	execResult, err := w.tx.ExecContext(
		ctx,
		`INSERT INTO notify_service.notification_delivery_attempts (
			notification_id, channel_code, status, attempt_number, provider_message_id, error_code, created_at, updated_at, delivered_at, next_attempt_at
		) VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $7, $7, NULL)
		ON CONFLICT (notification_id, channel_code, attempt_number) DO NOTHING`,
		notification.ID,
		channelCode,
		result.Status,
		result.AttemptNumber,
		result.ProviderMessageID,
		result.ErrorCode,
		now,
	)
	if err != nil {
		return fmt.Errorf("не удалось сохранить delivery attempt: %w", err)
	}
	if _, err := execResult.RowsAffected(); err != nil {
		return fmt.Errorf("не удалось подтвердить delivery attempt: %w", err)
	}
	return nil
}
