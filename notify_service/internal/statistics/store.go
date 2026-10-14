package statistics

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) LoadCursor(ctx context.Context, source string) (int64, error) {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO notify_service.event_statistics_source_cursors(source_name,last_sequence,updated_at) VALUES($1,0,$2) ON CONFLICT DO NOTHING`, source, time.Now().UTC()); err != nil {
		return 0, err
	}
	var cursor int64
	err := s.db.QueryRowContext(ctx, `SELECT last_sequence FROM notify_service.event_statistics_source_cursors WHERE source_name=$1`, source).Scan(&cursor)
	return cursor, err
}

func (s *Store) ApplyTriggers(ctx context.Context, source string, messages []Trigger, now time.Time) (ApplyResult, error) {
	if err := validateBatch(messages); err != nil {
		return ApplyResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplyResult{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO notify_service.event_statistics_source_cursors(source_name,last_sequence,updated_at) VALUES($1,0,$2) ON CONFLICT DO NOTHING`, source, now); err != nil {
		return ApplyResult{}, err
	}
	var result ApplyResult
	if err = tx.QueryRowContext(ctx, `SELECT last_sequence FROM notify_service.event_statistics_source_cursors WHERE source_name=$1 FOR UPDATE`, source).Scan(&result.LastSequence); err != nil {
		return ApplyResult{}, err
	}
	for _, message := range messages {
		if message.Sequence <= result.LastSequence {
			continue
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO notify_service.event_statistics_source_receipts(source_name,sequence,event_id,schema_version,completed_at,received_at) VALUES($1,$2,$3,$4,$5,$6)`, source, message.Sequence, int64(message.EventID), message.SchemaVersion, message.CompletedAt, now); err != nil {
			return ApplyResult{}, err
		}
		insert, insertErr := tx.ExecContext(ctx, `INSERT INTO notify_service.event_statistics_jobs(event_id,source_sequence,completed_at,state,next_attempt_at,created_at,updated_at) VALUES($1,$2,$3,'scheduled',$4,$4,$4) ON CONFLICT(event_id) DO NOTHING`, int64(message.EventID), message.Sequence, message.CompletedAt, now)
		if insertErr != nil {
			return ApplyResult{}, insertErr
		}
		count, err := insert.RowsAffected()
		if err != nil {
			return ApplyResult{}, err
		}
		result.AppliedCount += int(count)
		result.LastSequence = message.Sequence
	}
	if _, err = tx.ExecContext(ctx, `UPDATE notify_service.event_statistics_source_cursors SET last_sequence=$2,updated_at=$3 WHERE source_name=$1`, source, result.LastSequence, now); err != nil {
		return ApplyResult{}, err
	}
	if err = tx.Commit(); err != nil {
		return ApplyResult{}, err
	}
	return result, nil
}

func (s *Store) ClaimDueJob(ctx context.Context, now time.Time, lease time.Duration) (*Job, error) {
	if lease <= 0 {
		return nil, errors.New("lease должен быть положительным")
	}
	var tokenBytes [16]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes[:])
	// Один SQL statement завершает claim до любого межсерверного HTTP-вызова.
	var job Job
	var eventID int64
	err := s.db.QueryRowContext(ctx, `WITH candidate AS (
	 SELECT id FROM notify_service.event_statistics_jobs
	 WHERE state IN ('scheduled','retry_wait','processing') AND next_attempt_at <= $1
	 AND (lease_until IS NULL OR lease_until <= $1)
	 ORDER BY next_attempt_at,id FOR UPDATE SKIP LOCKED LIMIT 1
	) UPDATE notify_service.event_statistics_jobs job SET state='processing',lease_until=$2,lease_token=$3,updated_at=$1
	 FROM candidate WHERE job.id=candidate.id
	 RETURNING job.id,job.event_id,job.source_sequence,job.attempt_count,job.lease_token`, now, now.Add(lease), token).Scan(&job.ID, &eventID, &job.SourceSequence, &job.AttemptCount, &job.LeaseToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job.EventID = uint64(eventID)
	return &job, nil
}

func (s *Store) Complete(ctx context.Context, job Job, result ProcessResult, now time.Time) error {
	state := StateCompleted
	if result.Outcome == "missing" {
		state = StateMissing
	}
	return changed(s.db.ExecContext(ctx, `UPDATE notify_service.event_statistics_jobs SET state=$3,outcome=$4,processed_at=$5,lease_until=NULL,lease_token=NULL,last_error_code=NULL,updated_at=$6 WHERE id=$1 AND lease_token=$2 AND state='processing'`, job.ID, job.LeaseToken, state, result.Outcome, result.ProcessedAt, now))
}

func (s *Store) Retry(ctx context.Context, job Job, next time.Time, errorCode string, now time.Time) error {
	return changed(s.db.ExecContext(ctx, `UPDATE notify_service.event_statistics_jobs SET state='retry_wait',attempt_count=attempt_count+1,next_attempt_at=$3,last_error_code=$4,lease_until=NULL,lease_token=NULL,updated_at=$5 WHERE id=$1 AND lease_token=$2 AND state='processing'`, job.ID, job.LeaseToken, next, errorCode, now))
}

func changed(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("statistics lease потерян")
	}
	return nil
}
