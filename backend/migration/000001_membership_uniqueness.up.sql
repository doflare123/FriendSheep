CREATE TABLE IF NOT EXISTS membership_dedupe_audit (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    duplicate_row_id BIGINT NOT NULL,
    key_a BIGINT NOT NULL,
    key_b BIGINT NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF to_regclass('public.group_users') IS NOT NULL
       AND to_regclass('public.role_in_groups') IS NOT NULL THEN
        EXECUTE 'LOCK TABLE public.group_users IN SHARE ROW EXCLUSIVE MODE';

        EXECUTE $sql$
            WITH ranked AS (
                SELECT
                    gu.id,
                    gu.user_id,
                    gu.group_id,
                    ROW_NUMBER() OVER (
                        PARTITION BY gu.user_id, gu.group_id
                        ORDER BY
                            CASE rig.name
                                WHEN 'Админ' THEN 1
                                WHEN 'Модератор' THEN 2
                                WHEN 'Участник' THEN 3
                                ELSE 4
                            END ASC,
                            gu.id ASC
                    ) AS rn
                FROM public.group_users gu
                LEFT JOIN public.role_in_groups rig ON rig.id = gu.role_in_group_id
            )
            INSERT INTO membership_dedupe_audit (table_name, duplicate_row_id, key_a, key_b)
            SELECT 'group_users', id, user_id, group_id
            FROM ranked
            WHERE rn > 1
        $sql$;

        EXECUTE $sql$
            WITH ranked AS (
                SELECT
                    gu.id,
                    ROW_NUMBER() OVER (
                        PARTITION BY gu.user_id, gu.group_id
                        ORDER BY
                            CASE rig.name
                                WHEN 'Админ' THEN 1
                                WHEN 'Модератор' THEN 2
                                WHEN 'Участник' THEN 3
                                ELSE 4
                            END ASC,
                            gu.id ASC
                    ) AS rn
                FROM public.group_users gu
                LEFT JOIN public.role_in_groups rig ON rig.id = gu.role_in_group_id
            )
            DELETE FROM public.group_users gu
            USING ranked
            WHERE gu.id = ranked.id
              AND ranked.rn > 1
        $sql$;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('public.events_users') IS NOT NULL THEN
        EXECUTE 'LOCK TABLE public.events_users IN SHARE ROW EXCLUSIVE MODE';

        EXECUTE $sql$
            WITH ranked AS (
                SELECT
                    eu.id,
                    eu.event_id,
                    eu.user_id,
                    ROW_NUMBER() OVER (
                        PARTITION BY eu.event_id, eu.user_id
                        ORDER BY eu.id ASC
                    ) AS rn
                FROM public.events_users eu
            )
            INSERT INTO membership_dedupe_audit (table_name, duplicate_row_id, key_a, key_b)
            SELECT 'events_users', id, event_id, user_id
            FROM ranked
            WHERE rn > 1
        $sql$;

        EXECUTE $sql$
            WITH ranked AS (
                SELECT
                    eu.id,
                    ROW_NUMBER() OVER (
                        PARTITION BY eu.event_id, eu.user_id
                        ORDER BY eu.id ASC
                    ) AS rn
                FROM public.events_users eu
            )
            DELETE FROM public.events_users eu
            USING ranked
            WHERE eu.id = ranked.id
              AND ranked.rn > 1
        $sql$;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('public.events') IS NOT NULL
       AND to_regclass('public.events_users') IS NOT NULL THEN
        EXECUTE $sql$
            WITH stats AS (
                SELECT
                    e.id AS event_id,
                    COALESCE(COUNT(eu.user_id), 0)::BIGINT AS cnt
                FROM public.events e
                LEFT JOIN public.events_users eu ON eu.event_id = e.id
                GROUP BY e.id
            )
            UPDATE public.events e
            SET current_users = stats.cnt
            FROM stats
            WHERE e.id = stats.event_id
              AND e.current_users IS DISTINCT FROM stats.cnt
        $sql$;
    END IF;
END $$;

DO $$
DECLARE
    index_exists BOOLEAN;
    index_is_compatible BOOLEAN;
BEGIN
    IF to_regclass('public.group_users') IS NULL
       OR to_regclass('public.role_in_groups') IS NULL THEN
        RETURN;
    END IF;

    SELECT to_regclass('public.idx_group_user_membership') IS NOT NULL
    INTO index_exists;

    IF index_exists THEN
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
        INTO index_is_compatible;

        IF NOT index_is_compatible THEN
            RAISE EXCEPTION
                'Existing index public.idx_group_user_membership is not a valid UNIQUE index on public.group_users(user_id, group_id)';
        END IF;
    ELSE
        CREATE UNIQUE INDEX idx_group_user_membership
            ON public.group_users (user_id, group_id);
    END IF;
END $$;

DO $$
DECLARE
    index_exists BOOLEAN;
    index_is_compatible BOOLEAN;
BEGIN
    IF to_regclass('public.events_users') IS NULL THEN
        RETURN;
    END IF;

    SELECT to_regclass('public.idx_event_user_membership') IS NOT NULL
    INTO index_exists;

    IF index_exists THEN
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
        INTO index_is_compatible;

        IF NOT index_is_compatible THEN
            RAISE EXCEPTION
                'Existing index public.idx_event_user_membership is not a valid UNIQUE index on public.events_users(event_id, user_id)';
        END IF;
    ELSE
        CREATE UNIQUE INDEX idx_event_user_membership
            ON public.events_users (event_id, user_id);
    END IF;
END $$;
