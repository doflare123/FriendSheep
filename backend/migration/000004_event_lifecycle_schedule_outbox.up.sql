CREATE TABLE IF NOT EXISTS event_lifecycle_schedule_outbox (
    sequence BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    message_id UUID NOT NULL UNIQUE,
    event_id BIGINT NOT NULL CHECK (event_id > 0),
    operation VARCHAR(32) NOT NULL CHECK (operation IN ('schedule_upsert', 'schedule_cancel')),
    start_time TIMESTAMPTZ NULL,
    end_time TIMESTAMPTZ NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    schema_version SMALLINT NOT NULL DEFAULT 1 CHECK (schema_version = 1),
    CONSTRAINT event_lifecycle_schedule_payload_check CHECK (
        (operation = 'schedule_upsert' AND start_time IS NOT NULL AND end_time IS NOT NULL AND end_time > start_time)
        OR
        (operation = 'schedule_cancel' AND start_time IS NULL AND end_time IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_event_lifecycle_schedule_outbox_event_sequence
    ON event_lifecycle_schedule_outbox (event_id, sequence DESC);

CREATE INDEX IF NOT EXISTS idx_event_lifecycle_schedule_outbox_occurred_at
    ON event_lifecycle_schedule_outbox (occurred_at);
