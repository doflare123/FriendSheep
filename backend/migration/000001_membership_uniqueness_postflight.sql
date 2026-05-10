-- Postflight verification for migration 000001_membership_uniqueness.up.sql.
-- This script uses only temporary session objects and raises on any failed invariant.

CREATE TEMP TABLE membership_uniqueness_postflight_checks (
    check_name TEXT NOT NULL,
    status TEXT NOT NULL,
    details TEXT NOT NULL
) ON COMMIT DROP;

DO $$
DECLARE
    group_index_compatible BOOLEAN := FALSE;
    event_index_compatible BOOLEAN := FALSE;
    audit_table_compatible BOOLEAN := FALSE;
    duplicate_group_keys BIGINT := 0;
    duplicate_event_keys BIGINT := 0;
    drifted_events BIGINT := 0;
    group_null_key_rows BIGINT := 0;
    event_null_key_rows BIGINT := 0;
    audit_rows BIGINT := 0;
BEGIN
    IF to_regclass('public.membership_dedupe_audit') IS NOT NULL THEN
        SELECT EXISTS (
            SELECT 1
            FROM pg_class tbl
            JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
            JOIN LATERAL (
                SELECT
                    COUNT(*) FILTER (
                        WHERE att.attname = 'id'
                          AND format_type(att.atttypid, att.atttypmod) = 'bigint'
                          AND (att.attidentity IN ('a', 'd') OR def.adbin IS NOT NULL)
                    ) AS id_ok,
                    COUNT(*) FILTER (
                        WHERE att.attname = 'table_name'
                          AND format_type(att.atttypid, att.atttypmod) = 'text'
                    ) AS table_name_ok,
                    COUNT(*) FILTER (
                        WHERE att.attname = 'duplicate_row_id'
                          AND format_type(att.atttypid, att.atttypmod) = 'bigint'
                    ) AS duplicate_row_id_ok,
                    COUNT(*) FILTER (
                        WHERE att.attname = 'key_a'
                          AND format_type(att.atttypid, att.atttypmod) = 'bigint'
                    ) AS key_a_ok,
                    COUNT(*) FILTER (
                        WHERE att.attname = 'key_b'
                          AND format_type(att.atttypid, att.atttypmod) = 'bigint'
                    ) AS key_b_ok,
                    COUNT(*) FILTER (
                        WHERE att.attname = 'captured_at'
                          AND format_type(att.atttypid, att.atttypmod) = 'timestamp with time zone'
                          AND def.adbin IS NOT NULL
                    ) AS captured_at_ok
                FROM pg_attribute att
                LEFT JOIN pg_attrdef def
                  ON def.adrelid = att.attrelid
                 AND def.adnum = att.attnum
                WHERE att.attrelid = tbl.oid
                  AND att.attnum > 0
                  AND NOT att.attisdropped
            ) AS contract ON TRUE
            WHERE ns.nspname = 'public'
              AND tbl.relname = 'membership_dedupe_audit'
              AND contract.id_ok = 1
              AND contract.table_name_ok = 1
              AND contract.duplicate_row_id_ok = 1
              AND contract.key_a_ok = 1
              AND contract.key_b_ok = 1
              AND contract.captured_at_ok = 1
        )
        INTO audit_table_compatible;

        EXECUTE 'SELECT COUNT(*) FROM public.membership_dedupe_audit'
        INTO audit_rows;
    END IF;

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'contract.membership_dedupe_audit',
        CASE
            WHEN to_regclass('public.membership_dedupe_audit') IS NOT NULL
             AND audit_table_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.membership_dedupe_audit') IS NULL THEN
                'table public.membership_dedupe_audit is missing'
            WHEN NOT audit_table_compatible THEN
                'public.membership_dedupe_audit exists but does not match expected contract'
            ELSE
                format('audit table present and compatible; audit rows=%s', audit_rows)
        END
    );

    IF to_regclass('public.group_users') IS NOT NULL THEN
        SELECT EXISTS (
            SELECT 1
            FROM pg_index i
            JOIN pg_class idx ON idx.oid = i.indexrelid
            JOIN pg_class tbl ON tbl.oid = i.indrelid
            JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
            WHERE idx.oid = to_regclass('public.idx_group_user_membership')
              AND ns.nspname = 'public'
              AND tbl.relname = 'group_users'
              AND i.indisunique
              AND i.indisvalid
              AND i.indpred IS NULL
              AND ARRAY(
                    SELECT att.attname::TEXT
                    FROM unnest(i.indkey) WITH ORDINALITY AS ord(attnum, position)
                    JOIN pg_attribute att
                      ON att.attrelid = tbl.oid
                     AND att.attnum = ord.attnum
                    ORDER BY ord.position
                ) = ARRAY['user_id', 'group_id']
        )
        INTO group_index_compatible;

        EXECUTE '
            SELECT COUNT(*)
            FROM public.group_users
            WHERE user_id IS NULL OR group_id IS NULL
        '
        INTO group_null_key_rows;

        EXECUTE '
            SELECT COUNT(*)::bigint
            FROM (
                SELECT 1
                FROM public.group_users
                GROUP BY user_id, group_id
                HAVING COUNT(*) > 1
            ) AS d
        '
        INTO duplicate_group_keys;
    END IF;

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'contract.group_users_not_null_keys',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.group_users'::regclass
                  AND att.attname IN ('user_id', 'group_id')
                  AND att.attnotnull
            ) = 2
             AND group_null_key_rows = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            ELSE format('NULL-key rows=%s', group_null_key_rows)
        END
    );

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'index_contract.idx_group_user_membership',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN group_index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            WHEN group_index_compatible THEN 'unique index matches public.group_users(user_id, group_id)'
            ELSE 'missing or incompatible public.idx_group_user_membership'
        END
    );

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'data.group_users_duplicates_removed',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN duplicate_group_keys = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            ELSE format('remaining duplicate keys=%s', duplicate_group_keys)
        END
    );

    IF to_regclass('public.events_users') IS NOT NULL THEN
        SELECT EXISTS (
            SELECT 1
            FROM pg_index i
            JOIN pg_class idx ON idx.oid = i.indexrelid
            JOIN pg_class tbl ON tbl.oid = i.indrelid
            JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
            WHERE idx.oid = to_regclass('public.idx_event_user_membership')
              AND ns.nspname = 'public'
              AND tbl.relname = 'events_users'
              AND i.indisunique
              AND i.indisvalid
              AND i.indpred IS NULL
              AND ARRAY(
                    SELECT att.attname::TEXT
                    FROM unnest(i.indkey) WITH ORDINALITY AS ord(attnum, position)
                    JOIN pg_attribute att
                      ON att.attrelid = tbl.oid
                     AND att.attnum = ord.attnum
                    ORDER BY ord.position
                ) = ARRAY['event_id', 'user_id']
        )
        INTO event_index_compatible;

        EXECUTE '
            SELECT COUNT(*)
            FROM public.events_users
            WHERE event_id IS NULL OR user_id IS NULL
        '
        INTO event_null_key_rows;

        EXECUTE '
            SELECT COUNT(*)::bigint
            FROM (
                SELECT 1
                FROM public.events_users
                GROUP BY event_id, user_id
                HAVING COUNT(*) > 1
            ) AS d
        '
        INTO duplicate_event_keys;
    END IF;

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'contract.events_users_not_null_keys',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.events_users'::regclass
                  AND att.attname IN ('event_id', 'user_id')
                  AND att.attnotnull
            ) = 2
             AND event_null_key_rows = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            ELSE format('NULL-key rows=%s', event_null_key_rows)
        END
    );

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'index_contract.idx_event_user_membership',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN event_index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            WHEN event_index_compatible THEN 'unique index matches public.events_users(event_id, user_id)'
            ELSE 'missing or incompatible public.idx_event_user_membership'
        END
    );

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'data.events_users_duplicates_removed',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN duplicate_event_keys = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            ELSE format('remaining duplicate keys=%s', duplicate_event_keys)
        END
    );

    IF to_regclass('public.events') IS NOT NULL
       AND to_regclass('public.events_users') IS NOT NULL THEN
        EXECUTE '
            WITH expected AS (
                SELECT
                    e.id AS event_id,
                    COALESCE(COUNT(eu.user_id), 0)::bigint AS expected_current_users
                FROM public.events e
                LEFT JOIN public.events_users eu ON eu.event_id = e.id
                GROUP BY e.id
            )
            SELECT COUNT(*)::bigint
            FROM expected
            JOIN public.events ON events.id = expected.event_id
            WHERE events.current_users IS DISTINCT FROM expected.expected_current_users
        '
        INTO drifted_events;
    END IF;

    INSERT INTO membership_uniqueness_postflight_checks
    VALUES (
        'data.events_current_users_reconciled',
        CASE
            WHEN to_regclass('public.events') IS NULL OR to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN drifted_events = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events') IS NULL OR to_regclass('public.events_users') IS NULL THEN
                'public.events and/or public.events_users is missing'
            ELSE
                format('events still drifting=%s', drifted_events)
        END
    );
END $$;

SELECT check_name, status, details
FROM membership_uniqueness_postflight_checks
ORDER BY
    CASE status
        WHEN 'FAIL' THEN 1
        ELSE 2
    END,
    check_name;

DO $$
DECLARE
    failed_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO failed_count
    FROM membership_uniqueness_postflight_checks
    WHERE status = 'FAIL';

    IF failed_count > 0 THEN
        RAISE EXCEPTION
            'membership uniqueness postflight failed: % invariant checks failed. Keep rollout open and investigate.',
            failed_count;
    END IF;
END $$;
