-- Независимый durable поток оркестрации статистики; факты и агрегаты остаются в монолите.
CREATE TABLE notify_service.event_statistics_source_cursors (
    source_name text PRIMARY KEY,
    last_sequence bigint NOT NULL DEFAULT 0 CHECK (last_sequence >= 0),
    updated_at timestamptz NOT NULL
);
CREATE TABLE notify_service.event_statistics_source_receipts (
    source_name text NOT NULL REFERENCES notify_service.event_statistics_source_cursors(source_name),
    sequence bigint NOT NULL CHECK (sequence > 0),
    event_id bigint NOT NULL CHECK (event_id > 0),
    schema_version integer NOT NULL CHECK (schema_version = 1),
    completed_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    PRIMARY KEY (source_name, sequence)
);
CREATE TABLE notify_service.event_statistics_jobs (
    id bigserial PRIMARY KEY,
    event_id bigint NOT NULL UNIQUE CHECK (event_id > 0),
    source_sequence bigint NOT NULL CHECK (source_sequence > 0),
    completed_at timestamptz NOT NULL,
    state text NOT NULL CHECK (state IN ('scheduled','processing','retry_wait','completed','missing')),
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at timestamptz NOT NULL,
    lease_until timestamptz,
    lease_token text,
    last_error_code text,
    outcome text,
    processed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK ((lease_until IS NULL) = (lease_token IS NULL))
);
CREATE INDEX event_statistics_jobs_due_idx ON notify_service.event_statistics_jobs(next_attempt_at, id)
    WHERE state IN ('scheduled','processing','retry_wait');
