-- Postflight verification for migration 000002_group_join_request_pending_uniqueness.up.sql.
-- This script uses only temporary session objects and raises on any failed invariant.

CREATE TEMP TABLE group_join_request_pending_uniqueness_postflight_checks (
    check_name TEXT NOT NULL,
    status TEXT NOT NULL,
    details TEXT NOT NULL
) ON COMMIT DROP;

DO $$
DECLARE
    index_compatible BOOLEAN := FALSE;
    pending_duplicate_keys BIGINT := 0;
    null_key_rows BIGINT := 0;
BEGIN
    INSERT INTO group_join_request_pending_uniqueness_postflight_checks
    VALUES (
        'readiness.group_join_requests_table',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NOT NULL THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NOT NULL THEN 'table public.group_join_requests exists'
            ELSE 'table public.group_join_requests is missing'
        END
    );

    IF to_regclass('public.group_join_requests') IS NOT NULL THEN
        SELECT EXISTS (
            SELECT 1
            FROM pg_index i
            JOIN pg_class idx ON idx.oid = i.indexrelid
            JOIN pg_class tbl ON tbl.oid = i.indrelid
            JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
            WHERE idx.oid = to_regclass('public.idx_group_join_request_pending_unique')
              AND ns.nspname = 'public'
              AND tbl.relname = 'group_join_requests'
              AND i.indisunique
              AND i.indisvalid
              AND pg_get_expr(i.indpred, tbl.oid) = '(status = ''pending''::text)'
              AND ARRAY(
                    SELECT att.attname::TEXT
                    FROM unnest(i.indkey) WITH ORDINALITY AS ord(attnum, position)
                    JOIN pg_attribute att
                      ON att.attrelid = tbl.oid
                     AND att.attnum = ord.attnum
                    ORDER BY ord.position
                ) = ARRAY['user_id', 'group_id']
        )
        INTO index_compatible;

        EXECUTE '
            SELECT COUNT(*)
            FROM public.group_join_requests
            WHERE user_id IS NULL OR group_id IS NULL OR status IS NULL
        '
        INTO null_key_rows;

        EXECUTE '
            SELECT COUNT(*)::bigint
            FROM (
                SELECT 1
                FROM public.group_join_requests
                WHERE status = ''pending''
                GROUP BY user_id, group_id
                HAVING COUNT(*) > 1
            ) AS duplicates
        '
        INTO pending_duplicate_keys;
    END IF;

    INSERT INTO group_join_request_pending_uniqueness_postflight_checks
    VALUES (
        'contract.group_join_requests_not_null_keys',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.group_join_requests'::regclass
                  AND att.attname IN ('user_id', 'group_id', 'status')
                  AND att.attnotnull
            ) = 3
             AND null_key_rows = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            ELSE format('NULL-key rows=%s', null_key_rows)
        END
    );

    INSERT INTO group_join_request_pending_uniqueness_postflight_checks
    VALUES (
        'index_contract.idx_group_join_request_pending_unique',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            WHEN index_compatible THEN
                'partial unique index matches public.group_join_requests(user_id, group_id) WHERE status = ''pending'''
            ELSE
                'missing or incompatible public.idx_group_join_request_pending_unique'
        END
    );

    INSERT INTO group_join_request_pending_uniqueness_postflight_checks
    VALUES (
        'data.pending_join_request_duplicates_removed',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN pending_duplicate_keys = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            ELSE format('remaining pending duplicate keys=%s', pending_duplicate_keys)
        END
    );
END $$;

SELECT check_name, status, details
FROM group_join_request_pending_uniqueness_postflight_checks
ORDER BY check_name;

DO $$
DECLARE
    failed_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO failed_count
    FROM group_join_request_pending_uniqueness_postflight_checks
    WHERE status = 'FAIL';

    IF failed_count > 0 THEN
        RAISE EXCEPTION
            'group join request pending uniqueness postflight failed: % invariant checks failed. Keep rollout open and investigate.',
            failed_count;
    END IF;
END $$;
