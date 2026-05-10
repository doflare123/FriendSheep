-- Read-only preflight for migration 000001_membership_uniqueness.up.sql.
-- This script uses only temporary session objects and fails fast on blocking drift.

CREATE TEMP TABLE membership_uniqueness_preflight_checks (
    check_name TEXT NOT NULL,
    status TEXT NOT NULL,
    details TEXT NOT NULL
) ON COMMIT DROP;

DO $$
DECLARE
    current_is_superuser BOOLEAN := FALSE;
    group_index_compatible BOOLEAN := FALSE;
    event_index_compatible BOOLEAN := FALSE;
    audit_table_compatible BOOLEAN := FALSE;
    duplicate_group_keys BIGINT := 0;
    duplicate_group_rows BIGINT := 0;
    duplicate_event_keys BIGINT := 0;
    duplicate_event_rows BIGINT := 0;
    drifted_events BIGINT := 0;
    drift_abs_delta BIGINT := 0;
    group_null_key_rows BIGINT := 0;
    event_null_key_rows BIGINT := 0;
    group_users_owner TEXT;
    events_users_owner TEXT;
    audit_rows BIGINT := 0;
    admin_roles BIGINT := 0;
    moderator_roles BIGINT := 0;
    member_roles BIGINT := 0;
    unexpected_role_duplicate_rows BIGINT := 0;
BEGIN
    SELECT r.rolsuper
    INTO current_is_superuser
    FROM pg_roles r
    WHERE r.rolname = current_user;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.public_schema_usage',
        CASE
            WHEN has_schema_privilege(current_user, 'public', 'USAGE') THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN has_schema_privilege(current_user, 'public', 'USAGE') THEN 'schema public is accessible'
            ELSE 'schema public is not accessible for current_user'
        END
    );

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
    ELSE
        audit_table_compatible := FALSE;
        audit_rows := 0;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.membership_dedupe_audit',
        CASE
            WHEN to_regclass('public.membership_dedupe_audit') IS NULL THEN
                CASE
                    WHEN has_schema_privilege(current_user, 'public', 'CREATE') THEN 'PASS'
                    ELSE 'FAIL'
                END
            WHEN audit_table_compatible
             AND has_table_privilege(current_user, 'public.membership_dedupe_audit', 'INSERT') THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.membership_dedupe_audit') IS NULL
             AND has_schema_privilege(current_user, 'public', 'CREATE') THEN
                'audit table is absent; current_user can create it'
            WHEN to_regclass('public.membership_dedupe_audit') IS NULL THEN
                'audit table is absent; current_user lacks CREATE on schema public'
            WHEN NOT audit_table_compatible THEN
                'existing membership_dedupe_audit does not match expected insert contract'
            WHEN NOT has_table_privilege(current_user, 'public.membership_dedupe_audit', 'INSERT') THEN
                'existing membership_dedupe_audit is compatible, but current_user lacks INSERT'
            ELSE
                format('audit table is compatible; existing audit rows=%s', audit_rows)
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.group_users_table',
        CASE
            WHEN to_regclass('public.group_users') IS NOT NULL THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NOT NULL THEN 'table public.group_users exists'
            ELSE 'table public.group_users is missing'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_users_table',
        CASE
            WHEN to_regclass('public.events_users') IS NOT NULL THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NOT NULL THEN 'table public.events_users exists'
            ELSE 'table public.events_users is missing'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_table',
        CASE
            WHEN to_regclass('public.events') IS NOT NULL THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events') IS NOT NULL THEN 'table public.events exists'
            ELSE 'table public.events is missing'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.role_in_groups_table',
        CASE
            WHEN to_regclass('public.role_in_groups') IS NOT NULL THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.role_in_groups') IS NOT NULL THEN 'table public.role_in_groups exists'
            ELSE 'table public.role_in_groups is missing'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.group_users_columns',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.group_users'::regclass
                  AND att.attname IN ('id', 'user_id', 'group_id', 'role_in_group_id')
                  AND att.attnum > 0
                  AND NOT att.attisdropped
            ) = 4 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            ELSE 'expected columns: id, user_id, group_id, role_in_group_id'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.events_users_columns',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.events_users'::regclass
                  AND att.attname IN ('id', 'event_id', 'user_id')
                  AND att.attnum > 0
                  AND NOT att.attisdropped
            ) = 3 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            ELSE 'expected columns: id, event_id, user_id'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.events_columns',
        CASE
            WHEN to_regclass('public.events') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.events'::regclass
                  AND att.attname IN ('id', 'current_users')
                  AND att.attnum > 0
                  AND NOT att.attisdropped
            ) = 2 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events') IS NULL THEN 'table public.events is missing'
            ELSE 'expected columns: id, current_users'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.role_in_groups_columns',
        CASE
            WHEN to_regclass('public.role_in_groups') IS NULL THEN 'FAIL'
            WHEN (
                SELECT COUNT(*)
                FROM pg_attribute att
                WHERE att.attrelid = 'public.role_in_groups'::regclass
                  AND att.attname IN ('id', 'name')
                  AND att.attnum > 0
                  AND NOT att.attisdropped
            ) = 2 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.role_in_groups') IS NULL THEN 'table public.role_in_groups is missing'
            ELSE 'expected columns: id, name'
        END
    );

    IF to_regclass('public.role_in_groups') IS NOT NULL THEN
        EXECUTE $$SELECT COUNT(*) FROM public.role_in_groups WHERE name = 'Админ'$$
        INTO admin_roles;
        EXECUTE $$SELECT COUNT(*) FROM public.role_in_groups WHERE name = 'Модератор'$$
        INTO moderator_roles;
        EXECUTE $$SELECT COUNT(*) FROM public.role_in_groups WHERE name = 'Участник'$$
        INTO member_roles;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.group_role_priority_values',
        CASE
            WHEN to_regclass('public.role_in_groups') IS NULL THEN 'FAIL'
            WHEN admin_roles = 1 AND moderator_roles = 1 AND member_roles = 1 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.role_in_groups') IS NULL THEN 'table public.role_in_groups is missing'
            ELSE format(
                'required canonical roles: Админ=%s, Модератор=%s, Участник=%s',
                admin_roles,
                moderator_roles,
                member_roles
            )
        END
    );

    IF to_regclass('public.group_users') IS NOT NULL
       AND to_regclass('public.role_in_groups') IS NOT NULL THEN
        EXECUTE $$
            WITH duplicate_pairs AS (
                SELECT user_id, group_id
                FROM public.group_users
                GROUP BY user_id, group_id
                HAVING COUNT(*) > 1
            )
            SELECT COUNT(*)::bigint
            FROM public.group_users gu
            LEFT JOIN public.role_in_groups rig ON rig.id = gu.role_in_group_id
            JOIN duplicate_pairs dp ON dp.user_id = gu.user_id AND dp.group_id = gu.group_id
            WHERE rig.name IS NULL OR rig.name NOT IN ('Админ', 'Модератор', 'Участник')
        $$
        INTO unexpected_role_duplicate_rows;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'contract.group_users_unexpected_roles_in_duplicates',
        CASE
            WHEN to_regclass('public.group_users') IS NULL OR to_regclass('public.role_in_groups') IS NULL THEN 'FAIL'
            WHEN unexpected_role_duplicate_rows = 0 THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL OR to_regclass('public.role_in_groups') IS NULL THEN
                'public.group_users and/or public.role_in_groups is missing'
            ELSE
                format('unexpected role rows inside duplicate group memberships=%s', unexpected_role_duplicate_rows)
        END
    );

    IF to_regclass('public.group_users') IS NOT NULL THEN
        SELECT pg_get_userbyid(c.relowner)
        INTO group_users_owner
        FROM pg_class c
        WHERE c.oid = 'public.group_users'::regclass;
    END IF;

    IF to_regclass('public.events_users') IS NOT NULL THEN
        SELECT pg_get_userbyid(c.relowner)
        INTO events_users_owner
        FROM pg_class c
        WHERE c.oid = 'public.events_users'::regclass;
    END IF;

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
    END IF;

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
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.group_users_owner_or_ready_index',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN current_is_superuser OR group_users_owner = current_user OR group_index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            WHEN current_is_superuser THEN 'current_user is superuser'
            WHEN group_users_owner = current_user THEN 'current_user owns public.group_users'
            WHEN group_index_compatible THEN 'compatible group_users unique index already exists'
            ELSE format('current_user does not own public.group_users; owner=%s', group_users_owner)
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_users_owner_or_ready_index',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN current_is_superuser OR events_users_owner = current_user OR event_index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            WHEN current_is_superuser THEN 'current_user is superuser'
            WHEN events_users_owner = current_user THEN 'current_user owns public.events_users'
            WHEN event_index_compatible THEN 'compatible events_users unique index already exists'
            ELSE format('current_user does not own public.events_users; owner=%s', events_users_owner)
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.group_users_dml_privileges',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN has_table_privilege(current_user, 'public.group_users', 'SELECT,DELETE') THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            WHEN has_table_privilege(current_user, 'public.group_users', 'SELECT,DELETE') THEN
                'current_user can SELECT and DELETE from public.group_users'
            ELSE
                'current_user lacks SELECT and/or DELETE on public.group_users'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_users_dml_privileges',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN has_table_privilege(current_user, 'public.events_users', 'SELECT,DELETE') THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            WHEN has_table_privilege(current_user, 'public.events_users', 'SELECT,DELETE') THEN
                'current_user can SELECT and DELETE from public.events_users'
            ELSE
                'current_user lacks SELECT and/or DELETE on public.events_users'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_update_privileges',
        CASE
            WHEN to_regclass('public.events') IS NULL THEN 'FAIL'
            WHEN has_table_privilege(current_user, 'public.events', 'SELECT,UPDATE') THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events') IS NULL THEN 'table public.events is missing'
            WHEN has_table_privilege(current_user, 'public.events', 'SELECT,UPDATE') THEN
                'current_user can SELECT and UPDATE public.events.current_users'
            ELSE
                'current_user lacks SELECT and/or UPDATE on public.events'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.role_in_groups_select',
        CASE
            WHEN to_regclass('public.role_in_groups') IS NULL THEN 'FAIL'
            WHEN has_table_privilege(current_user, 'public.role_in_groups', 'SELECT') THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.role_in_groups') IS NULL THEN 'table public.role_in_groups is missing'
            WHEN has_table_privilege(current_user, 'public.role_in_groups', 'SELECT') THEN
                'current_user can read canonical group roles'
            ELSE
                'current_user lacks SELECT on public.role_in_groups'
        END
    );

    IF to_regclass('public.group_users') IS NOT NULL THEN
        EXECUTE '
            SELECT COUNT(*)
            FROM public.group_users
            WHERE user_id IS NULL OR group_id IS NULL
        '
        INTO group_null_key_rows;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
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
            ELSE format('NOT NULL columns required; current NULL-key rows=%s', group_null_key_rows)
        END
    );

    IF to_regclass('public.events_users') IS NOT NULL THEN
        EXECUTE '
            SELECT COUNT(*)
            FROM public.events_users
            WHERE event_id IS NULL OR user_id IS NULL
        '
        INTO event_null_key_rows;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
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
            ELSE format('NOT NULL columns required; current NULL-key rows=%s', event_null_key_rows)
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'index_contract.idx_group_user_membership',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN to_regclass('public.idx_group_user_membership') IS NULL THEN 'PASS'
            WHEN group_index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            WHEN to_regclass('public.idx_group_user_membership') IS NULL THEN
                'index is absent; migration should create UNIQUE (user_id, group_id)'
            WHEN group_index_compatible THEN
                'existing index matches UNIQUE public.group_users(user_id, group_id)'
            ELSE
                'existing public.idx_group_user_membership is incompatible with migration contract'
        END
    );

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'index_contract.idx_event_user_membership',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN to_regclass('public.idx_event_user_membership') IS NULL THEN 'PASS'
            WHEN event_index_compatible THEN 'PASS'
            ELSE 'FAIL'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            WHEN to_regclass('public.idx_event_user_membership') IS NULL THEN
                'index is absent; migration should create UNIQUE (event_id, user_id)'
            WHEN event_index_compatible THEN
                'existing index matches UNIQUE public.events_users(event_id, user_id)'
            ELSE
                'existing public.idx_event_user_membership is incompatible with migration contract'
        END
    );

    IF to_regclass('public.group_users') IS NOT NULL THEN
        EXECUTE '
            SELECT
                COUNT(*)::bigint,
                COALESCE(SUM(cnt - 1), 0)::bigint
            FROM (
                SELECT COUNT(*) AS cnt
                FROM public.group_users
                GROUP BY user_id, group_id
                HAVING COUNT(*) > 1
            ) AS d
        '
        INTO duplicate_group_keys, duplicate_group_rows;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.group_users_duplicates',
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'FAIL'
            WHEN duplicate_group_keys = 0 THEN 'PASS'
            ELSE 'WARN'
        END,
        CASE
            WHEN to_regclass('public.group_users') IS NULL THEN 'table public.group_users is missing'
            ELSE format(
                'duplicate keys=%s; rows to delete=%s',
                duplicate_group_keys,
                duplicate_group_rows
            )
        END
    );

    IF to_regclass('public.events_users') IS NOT NULL THEN
        EXECUTE '
            SELECT
                COUNT(*)::bigint,
                COALESCE(SUM(cnt - 1), 0)::bigint
            FROM (
                SELECT COUNT(*) AS cnt
                FROM public.events_users
                GROUP BY event_id, user_id
                HAVING COUNT(*) > 1
            ) AS d
        '
        INTO duplicate_event_keys, duplicate_event_rows;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_users_duplicates',
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN duplicate_event_keys = 0 THEN 'PASS'
            ELSE 'WARN'
        END,
        CASE
            WHEN to_regclass('public.events_users') IS NULL THEN 'table public.events_users is missing'
            ELSE format(
                'duplicate keys=%s; rows to delete=%s',
                duplicate_event_keys,
                duplicate_event_rows
            )
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
            SELECT
                COUNT(*)::bigint,
                COALESCE(SUM(ABS(events.current_users - expected.expected_current_users)), 0)::bigint
            FROM expected
            JOIN public.events ON events.id = expected.event_id
            WHERE events.current_users IS DISTINCT FROM expected.expected_current_users
        '
        INTO drifted_events, drift_abs_delta;
    END IF;

    INSERT INTO membership_uniqueness_preflight_checks
    VALUES (
        'readiness.events_current_users_drift',
        CASE
            WHEN to_regclass('public.events') IS NULL OR to_regclass('public.events_users') IS NULL THEN 'FAIL'
            WHEN drifted_events = 0 THEN 'PASS'
            ELSE 'WARN'
        END,
        CASE
            WHEN to_regclass('public.events') IS NULL OR to_regclass('public.events_users') IS NULL THEN
                'public.events and/or public.events_users is missing'
            ELSE
                format(
                    'events with drift=%s; summed absolute delta=%s',
                    drifted_events,
                    drift_abs_delta
                )
        END
    );
END $$;

SELECT check_name, status, details
FROM membership_uniqueness_preflight_checks
ORDER BY
    CASE status
        WHEN 'FAIL' THEN 1
        WHEN 'WARN' THEN 2
        ELSE 3
    END,
    check_name;

DO $$
DECLARE
    blocker_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO blocker_count
    FROM membership_uniqueness_preflight_checks
    WHERE status = 'FAIL';

    IF blocker_count > 0 THEN
        RAISE EXCEPTION
            'membership uniqueness preflight failed: % blocking checks. Resolve FAIL rows before rollout.',
            blocker_count;
    END IF;
END $$;
