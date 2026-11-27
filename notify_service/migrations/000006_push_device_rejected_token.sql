ALTER TABLE notify_service.push_device_installations
    ADD COLUMN last_rejected_token_fingerprint BYTEA NULL;

ALTER TABLE notify_service.push_device_installations
    ADD CONSTRAINT push_device_installations_rejected_token_fingerprint_check
    CHECK (
        last_rejected_token_fingerprint IS NULL
        OR octet_length(last_rejected_token_fingerprint) = 32
    );
