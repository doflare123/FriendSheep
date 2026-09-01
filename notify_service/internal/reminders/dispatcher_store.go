package reminders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) ClaimDeliveryTarget(ctx context.Context, now time.Time, leaseDuration time.Duration) (*DeliveryTarget, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("delivery store не инициализирован")
	}
	if leaseDuration <= 0 {
		return nil, errors.New("delivery lease duration должен быть положительным")
	}
	leaseToken, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать delivery lease token: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию claim delivery target: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id
			FROM notify_service.notification_delivery_targets
			WHERE state IN ($1, $2, $3)
			  AND (
			       (state IN ($1, $2) AND next_attempt_at IS NOT NULL AND next_attempt_at <= $4)
			    OR (state = $3 AND lease_until IS NOT NULL AND lease_until <= $4)
			  )
			ORDER BY COALESCE(next_attempt_at, lease_until), id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE notify_service.notification_delivery_targets target
		SET state = $3,
		    attempt_count = target.attempt_count + 1,
		    next_attempt_at = NULL,
		    lease_until = $5,
		    lease_token = $6,
		    updated_at = $4
		FROM candidate
		WHERE target.id = candidate.id
		RETURNING target.id, target.notification_id, target.channel_code, target.state, target.attempt_count,
		          target.next_attempt_at, target.lease_until, COALESCE(target.lease_token, ''),
		          COALESCE(target.last_error_code, ''), COALESCE(target.provider_message_id, ''), target.idempotency_key
	`, DeliveryStatePending, DeliveryStateRetryWait, DeliveryStateProcessing, now, now.Add(leaseDuration), leaseToken)

	target, found, err := scanDeliveryTargetRow(row)
	if err != nil {
		return nil, err
	}
	if !found {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("не удалось завершить пустой claim delivery target: %w", err)
		}
		return nil, nil
	}

	notificationRow := tx.QueryRowContext(
		ctx,
		`SELECT id, user_id, kind, resource_type, resource_id, reminder_offset_minutes, schedule_revision, payload, created_at, read_at
		 FROM notify_service.notifications
		 WHERE id = $1`,
		target.NotificationID,
	)
	notification, found, err := scanNotificationRecordRow(notificationRow)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("notification %q для delivery target не найдена", target.NotificationID)
	}
	target.Notification = notification

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось зафиксировать claim delivery target: %w", err)
	}
	return &target, nil
}

func (s *Store) MarkDeliveryDelivered(ctx context.Context, target DeliveryTarget, result DeliveryResult, now time.Time) error {
	execResult, err := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.notification_delivery_targets
		 SET state = $4,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULL,
		     provider_message_id = NULLIF($5, ''),
		     delivered_at = $6,
		     updated_at = $6
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		target.ID,
		target.LeaseToken,
		DeliveryStateProcessing,
		DeliveryStateDelivered,
		result.ProviderMessageID,
		now,
	)
	return checkDeliveryRowsAffected(execResult, err, "подтвердить доставку")
}

func (s *Store) RescheduleDelivery(ctx context.Context, target DeliveryTarget, nextAttempt time.Time, errorCode string, now time.Time) error {
	execResult, err := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.notification_delivery_targets
		 SET state = $4,
		     next_attempt_at = $5,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULLIF($6, ''),
		     updated_at = $7
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		target.ID,
		target.LeaseToken,
		DeliveryStateProcessing,
		DeliveryStateRetryWait,
		nextAttempt,
		errorCode,
		now,
	)
	return checkDeliveryRowsAffected(execResult, err, "перенести delivery target в retry_wait")
}

func (s *Store) MarkDeliveryTerminal(ctx context.Context, target DeliveryTarget, errorCode string, now time.Time) error {
	execResult, err := s.db.ExecContext(
		ctx,
		`UPDATE notify_service.notification_delivery_targets
		 SET state = $4,
		     next_attempt_at = NULL,
		     lease_until = NULL,
		     lease_token = NULL,
		     last_error_code = NULLIF($5, ''),
		     updated_at = $6
		 WHERE id = $1 AND lease_token = $2 AND state = $3`,
		target.ID,
		target.LeaseToken,
		DeliveryStateProcessing,
		DeliveryStateTerminalFailed,
		errorCode,
		now,
	)
	return checkDeliveryRowsAffected(execResult, err, "завершить delivery target терминально")
}

func scanDeliveryTargetRow(scanner interface{ Scan(...any) error }) (DeliveryTarget, bool, error) {
	var (
		target        DeliveryTarget
		nextAttemptAt sql.NullTime
		leaseUntil    sql.NullTime
	)
	err := scanner.Scan(
		&target.ID,
		&target.NotificationID,
		&target.ChannelCode,
		&target.State,
		&target.AttemptCount,
		&nextAttemptAt,
		&leaseUntil,
		&target.LeaseToken,
		&target.LastErrorCode,
		&target.ProviderMessageID,
		&target.IdempotencyKey,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return DeliveryTarget{}, false, nil
	}
	if err != nil {
		return DeliveryTarget{}, false, fmt.Errorf("не удалось прочитать delivery target: %w", err)
	}
	target.NextAttemptAt = nullTimePointer(nextAttemptAt)
	target.LeaseUntil = nullTimePointer(leaseUntil)
	return target, true, nil
}

func checkDeliveryRowsAffected(result sql.Result, err error, action string) error {
	if err != nil {
		return fmt.Errorf("не удалось %s: %w", action, err)
	}
	rows, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return fmt.Errorf("не удалось подтвердить %s: %w", action, rowsErr)
	}
	if rows != 1 {
		return fmt.Errorf("lease delivery target потерян во время операции %q", action)
	}
	return nil
}
