CREATE TABLE IF NOT EXISTS notify_service.push_device_installations (
    installation_id VARCHAR(43) PRIMARY KEY,
    state VARCHAR(32) NOT NULL,
    version UUID NOT NULL,
    user_id BIGINT NULL,
    session_id VARCHAR(24) NULL,
    binding_id UUID NULL,
    platform VARCHAR(16) NULL,
    provider VARCHAR(16) NULL,
    provider_project VARCHAR(64) NULL,
    environment VARCHAR(64) NULL,
    provider_token_ciphertext BYTEA NULL,
    provider_token_nonce BYTEA NULL,
    provider_token_key_id VARCHAR(64) NULL,
    token_fingerprint BYTEA NULL,
    last_operation_id UUID NULL,
    last_operation_fingerprint BYTEA NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT push_device_installations_id_check CHECK (installation_id ~ '^[A-Za-z0-9_-]{42}[AEIMQUYcgkosw048]$'),
    CONSTRAINT push_device_installations_state_check CHECK (state IN ('active', 'invalid_address', 'detached')),
    CONSTRAINT push_device_installations_user_check CHECK (user_id IS NULL OR user_id > 0),
    CONSTRAINT push_device_installations_session_check CHECK (session_id IS NULL OR session_id ~ '^[A-Za-z0-9_-]{24}$'),
    CONSTRAINT push_device_installations_platform_check CHECK (platform IS NULL OR platform IN ('android', 'ios')),
    CONSTRAINT push_device_installations_provider_check CHECK (provider IS NULL OR provider = 'fcm'),
    CONSTRAINT push_device_installations_project_check CHECK (provider_project IS NULL OR provider_project ~ '^[a-z][a-z0-9-]{0,63}$'),
    CONSTRAINT push_device_installations_environment_check CHECK (environment IS NULL OR environment ~ '^[a-z][a-z0-9-]{0,63}$'),
    CONSTRAINT push_device_installations_key_id_check CHECK (provider_token_key_id IS NULL OR provider_token_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'),
    CONSTRAINT push_device_installations_token_fingerprint_check CHECK (token_fingerprint IS NULL OR octet_length(token_fingerprint) = 32),
    CONSTRAINT push_device_installations_operation_fields_check CHECK (
        (last_operation_id IS NULL AND last_operation_fingerprint IS NULL)
        OR (last_operation_id IS NOT NULL AND last_operation_fingerprint IS NOT NULL AND octet_length(last_operation_fingerprint) = 32)
    ),
    CONSTRAINT push_device_installations_state_fields_check CHECK (
        (
            state = 'active'
            AND user_id IS NOT NULL
            AND session_id IS NOT NULL
            AND binding_id IS NOT NULL
            AND platform IS NOT NULL
            AND provider IS NOT NULL
            AND provider_project IS NOT NULL
            AND environment IS NOT NULL
            AND provider_token_ciphertext IS NOT NULL
            AND octet_length(provider_token_ciphertext) BETWEEN 17 AND 4112
            AND provider_token_nonce IS NOT NULL
            AND octet_length(provider_token_nonce) = 12
            AND provider_token_key_id IS NOT NULL
            AND token_fingerprint IS NOT NULL
        )
        OR (
            state = 'invalid_address'
            AND user_id IS NOT NULL
            AND session_id IS NOT NULL
            AND binding_id IS NOT NULL
            AND platform IS NOT NULL
            AND provider IS NOT NULL
            AND provider_project IS NOT NULL
            AND environment IS NOT NULL
            AND provider_token_ciphertext IS NULL
            AND provider_token_nonce IS NULL
            AND provider_token_key_id IS NULL
            AND token_fingerprint IS NULL
        )
        OR (
            state = 'detached'
            AND user_id IS NULL
            AND session_id IS NULL
            AND binding_id IS NULL
            AND provider_token_ciphertext IS NULL
            AND provider_token_nonce IS NULL
            AND provider_token_key_id IS NULL
            AND token_fingerprint IS NULL
            AND (
                (platform IS NULL AND provider IS NULL AND provider_project IS NULL AND environment IS NULL)
                OR (platform IS NOT NULL AND provider IS NOT NULL AND provider_project IS NOT NULL AND environment IS NOT NULL)
            )
        )
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS push_device_installations_unique_active_token
    ON notify_service.push_device_installations (provider, provider_project, environment, token_fingerprint)
    WHERE state = 'active';

CREATE INDEX IF NOT EXISTS idx_push_device_installations_active_user
    ON notify_service.push_device_installations (user_id, installation_id)
    WHERE state = 'active';
