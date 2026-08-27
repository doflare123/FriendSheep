CREATE TABLE IF NOT EXISTS notify_service.event_lifecycle_source_cursors (
    source_name TEXT PRIMARY KEY,
    last_sequence BIGINT NOT NULL DEFAULT 0 CHECK (last_sequence >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notify_service.event_lifecycle_source_messages (
    source_name TEXT NOT NULL,
    message_id UUID NOT NULL,
    sequence BIGINT NOT NULL CHECK (sequence >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (source_name, message_id)
);

CREATE TABLE IF NOT EXISTS notify_service.event_lifecycle_jobs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id BIGINT NOT NULL UNIQUE CHECK (event_id > 0),
    source_sequence BIGINT NOT NULL CHECK (source_sequence >= 0),
    source_message_id UUID NOT NULL,
    start_time TIMESTAMPTZ NULL,
    end_time TIMESTAMPTZ NULL,
    state VARCHAR(32) NOT NULL,
    next_action_at TIMESTAMPTZ NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at TIMESTAMPTZ NULL,
    lease_until TIMESTAMPTZ NULL,
    lease_token UUID NULL,
    last_error_code VARCHAR(64) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT event_lifecycle_jobs_state_check CHECK (
        state IN (
            'scheduled',
            'active_wait',
            'processing',
            'completed',
            'cancelled',
            'retry_wait',
            'terminal_failed'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_event_lifecycle_jobs_due_actions
    ON notify_service.event_lifecycle_jobs (state, next_action_at, id)
    WHERE state IN ('scheduled', 'active_wait');

CREATE INDEX IF NOT EXISTS idx_event_lifecycle_jobs_due_retries
    ON notify_service.event_lifecycle_jobs (state, next_attempt_at, id)
    WHERE state = 'retry_wait';

CREATE INDEX IF NOT EXISTS idx_event_lifecycle_jobs_lease
    ON notify_service.event_lifecycle_jobs (lease_until, id)
    WHERE lease_until IS NOT NULL;
