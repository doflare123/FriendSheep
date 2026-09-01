CREATE TABLE IF NOT EXISTS notify_service.event_reminder_source_cursors (
    source_name TEXT PRIMARY KEY,
    last_sequence BIGINT NOT NULL DEFAULT 0 CHECK (last_sequence >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notify_service.event_reminder_source_messages (
    source_name TEXT NOT NULL,
    message_id UUID NOT NULL,
    sequence BIGINT NOT NULL CHECK (sequence >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (source_name, message_id)
);

CREATE TABLE IF NOT EXISTS notify_service.event_reminder_jobs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id BIGINT NOT NULL CHECK (event_id > 0),
    source_revision BIGINT NOT NULL CHECK (source_revision >= 0),
    source_message_id UUID NOT NULL,
    reminder_offset_minutes INTEGER NOT NULL CHECK (reminder_offset_minutes >= 0),
    start_time TIMESTAMPTZ NULL,
    due_at TIMESTAMPTZ NULL,
    window_end_at TIMESTAMPTZ NULL,
    state VARCHAR(32) NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at TIMESTAMPTZ NULL,
    lease_until TIMESTAMPTZ NULL,
    lease_token TEXT NULL,
    last_error_code VARCHAR(64) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT event_reminder_jobs_unique_revision_offset UNIQUE (event_id, source_revision, reminder_offset_minutes),
    CONSTRAINT event_reminder_jobs_state_check CHECK (
        state IN (
            'scheduled',
            'processing',
            'retry_wait',
            'completed',
            'cancelled',
            'superseded',
            'skipped_missed_window',
            'expired',
            'terminal_failed'
        )
    ),
    CONSTRAINT event_reminder_jobs_window_check CHECK (
        (reminder_offset_minutes = 0 AND start_time IS NULL AND due_at IS NULL AND window_end_at IS NULL)
        OR (reminder_offset_minutes > 0 AND start_time IS NOT NULL AND due_at IS NOT NULL AND window_end_at IS NOT NULL AND due_at <= window_end_at AND window_end_at <= start_time)
    )
);

CREATE INDEX IF NOT EXISTS idx_event_reminder_jobs_due
    ON notify_service.event_reminder_jobs (state, due_at, id)
    WHERE state = 'scheduled';

CREATE INDEX IF NOT EXISTS idx_event_reminder_jobs_retry
    ON notify_service.event_reminder_jobs (state, next_attempt_at, id)
    WHERE state = 'retry_wait';

CREATE INDEX IF NOT EXISTS idx_event_reminder_jobs_lease
    ON notify_service.event_reminder_jobs (lease_until, id)
    WHERE lease_until IS NOT NULL;

CREATE TABLE IF NOT EXISTS notify_service.notifications (
    id TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL CHECK (user_id > 0),
    kind VARCHAR(64) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id BIGINT NOT NULL CHECK (resource_id > 0),
    payload JSONB NOT NULL,
    reminder_offset_minutes INTEGER NOT NULL CHECK (reminder_offset_minutes > 0),
    schedule_revision BIGINT NOT NULL CHECK (schedule_revision >= 0),
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    read_at TIMESTAMPTZ NULL,
    CONSTRAINT notifications_kind_check CHECK (kind IN ('event_reminder')),
    CONSTRAINT notifications_resource_type_check CHECK (resource_type IN ('event'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
    ON notify_service.notifications (user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_unread
    ON notify_service.notifications (user_id, read_at, created_at DESC, id DESC)
    WHERE read_at IS NULL;

CREATE TABLE IF NOT EXISTS notify_service.notification_delivery_targets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    notification_id TEXT NOT NULL REFERENCES notify_service.notifications (id) ON DELETE CASCADE,
    channel_code VARCHAR(64) NOT NULL,
    state VARCHAR(32) NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at TIMESTAMPTZ NULL,
    lease_until TIMESTAMPTZ NULL,
    lease_token TEXT NULL,
    last_error_code VARCHAR(64) NULL,
    provider_message_id TEXT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    delivered_at TIMESTAMPTZ NULL,
    CONSTRAINT notification_delivery_targets_unique_channel UNIQUE (notification_id, channel_code),
    CONSTRAINT notification_delivery_targets_channel_check CHECK (channel_code ~ '^[a-z][a-z0-9_]{1,63}$'),
    CONSTRAINT notification_delivery_targets_state_check CHECK (
        state IN ('pending', 'processing', 'retry_wait', 'delivered', 'terminal_failed')
    ),
    CONSTRAINT notification_delivery_targets_state_fields_check CHECK (
        (state = 'pending' AND next_attempt_at IS NOT NULL AND lease_until IS NULL AND lease_token IS NULL AND delivered_at IS NULL)
        OR (state = 'processing' AND next_attempt_at IS NULL AND lease_until IS NOT NULL AND lease_token IS NOT NULL AND delivered_at IS NULL)
        OR (state = 'retry_wait' AND next_attempt_at IS NOT NULL AND lease_until IS NULL AND lease_token IS NULL AND delivered_at IS NULL)
        OR (state = 'delivered' AND next_attempt_at IS NULL AND lease_until IS NULL AND lease_token IS NULL AND delivered_at IS NOT NULL)
        OR (state = 'terminal_failed' AND next_attempt_at IS NULL AND lease_until IS NULL AND lease_token IS NULL AND delivered_at IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_notification_delivery_targets_due
    ON notify_service.notification_delivery_targets (state, next_attempt_at, id)
    WHERE state IN ('pending', 'retry_wait');

CREATE INDEX IF NOT EXISTS idx_notification_delivery_targets_lease
    ON notify_service.notification_delivery_targets (lease_until, id)
    WHERE state = 'processing';
