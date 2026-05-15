-- Read-only preflight for migration 000002_group_join_request_pending_uniqueness.up.sql.
-- This script uses only temporary session objects and fails on blockers.

CREATE TEMP TABLE group_join_request_pending_uniqueness_preflight_checks (
    check_name TEXT NOT NULL,
    status TEXT NOT NULL,
    details TEXT NOT NULL
) ON COMMIT DROP;

DO $$
DECLARE
    current_is_superuser BOOLEAN := FALSE;
    index_compatible BOOLEAN := FALSE;
    table_owner TEXT := NULL;
    can_select_table BOOLEAN := FALSE;
    pending_duplicate_keys BIGINT := 0;
    pending_duplicate_rows BIGINT := 0;
    null_key_rows BIGINT := 0;
BEGIN
    SELECT r.rolsuper
    INTO current_is_superuser
    FROM pg_roles r
    WHERE r.rolname = current_user;

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
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

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
    VALUES (
        'contract.group_join_requests_columns',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.group_join_requests'::regclass
                  AND att.attname IN ('id', 'user_id', 'group_id', 'status')
                  AND att.attnum > 0
                  AND NOT att.attisdropped
            ) = 4 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            ELSE 'expected columns: id, user_id, group_id, status'
        END
    );

    IF to_regclass('public.group_join_requests') IS NOT NULL THEN
        can_select_table := has_table_privilege(current_user, 'public.group_join_requests', 'SELECT');
    END IF;

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
    VALUES (
        'readiness.group_join_requests_select',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN can_select_table THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            WHEN can_select_table THEN 'current_user can SELECT from public.group_join_requests'
            ELSE 'current_user lacks SELECT on public.group_join_requests'
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

        SELECT pg_get_userbyid(c.relowner)
        INTO table_owner
        FROM pg_class c
        WHERE c.oid = 'public.group_join_requests'::regclass;
    END IF;

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
    VALUES (
        'index_contract.idx_group_join_request_pending_unique',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN to_regclass('public.idx_group_join_request_pending_unique') IS NULL THEN 'PASS'
            WHEN index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            WHEN to_regclass('public.idx_group_join_request_pending_unique') IS NULL THEN
                'index is absent; migration should create UNIQUE (user_id, group_id) WHERE status = ''pending'''
            WHEN index_compatible THEN
                'existing index matches UNIQUE public.group_join_requests(user_id, group_id) WHERE status = ''pending'''
            ELSE
                'existing public.idx_group_join_request_pending_unique is incompatible with migration contract'
        END
    );

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
    VALUES (
        'readiness.group_join_requests_owner_or_ready_index',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN current_is_superuser OR table_owner = current_user OR index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            WHEN current_is_superuser THEN 'current_user is superuser'
            WHEN table_owner = current_user THEN 'current_user owns public.group_join_requests'
            WHEN index_compatible THEN 'compatible partial unique index already exists'
            ELSE format('current_user does not own public.group_join_requests; owner=%s', table_owner)
        END
    );

    IF can_select_table THEN
        EXECUTE '
            SELECT COUNT(*)
            FROM public.group_join_requests
            WHERE user_id IS NULL OR group_id IS NULL OR status IS NULL
        '
        INTO null_key_rows;

        EXECUTE '
            SELECT
                COUNT(*)::bigint,
                COALESCE(SUM(cnt - 1), 0)::bigint
            FROM (
                SELECT COUNT(*) AS cnt
                FROM public.group_join_requests
                WHERE status = ''pending''
                GROUP BY user_id, group_id
                HAVING COUNT(*) > 1
            ) AS duplicates
        '
        INTO pending_duplicate_keys, pending_duplicate_rows;
    END IF;

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
    VALUES (
        'contract.group_join_requests_not_null_keys',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN can_select_table
             AND (
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
            WHEN NOT can_select_table THEN 'current_user lacks SELECT on public.group_join_requests'
            ELSE format('NOT NULL columns required; current NULL-key rows=%s', null_key_rows)
        END
    );

    INSERT INTO group_join_request_pending_uniqueness_preflight_checks
    VALUES (
        'data.pending_join_request_duplicates',
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'FAIL'
            WHEN NOT can_select_table THEN 'FAIL'
            WHEN pending_duplicate_keys = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_join_requests') IS NULL THEN 'table public.group_join_requests is missing'
            WHEN NOT can_select_table THEN 'current_user lacks SELECT on public.group_join_requests'
            ELSE format(
                'blocking pending duplicate keys=%s; extra pending rows=%s',
                pending_duplicate_keys,
                pending_duplicate_rows
            )
        END
    );
END $$;

SELECT check_name, status, details
FROM group_join_request_pending_uniqueness_preflight_checks
ORDER BY check_name;

DO $$
DECLARE
    blocker_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO blocker_count
    FROM group_join_request_pending_uniqueness_preflight_checks
    WHERE status = 'FAIL';

    IF blocker_count > 0 THEN
        RAISE EXCEPTION
            'group join request pending uniqueness preflight failed: % blocking checks. Resolve FAIL rows before rollout.',
            blocker_count;
    END IF;
END $$;
