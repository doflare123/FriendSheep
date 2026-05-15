DO $$
DECLARE
    index_exists BOOLEAN := FALSE;
    index_is_compatible BOOLEAN := FALSE;
    duplicate_pending_keys BIGINT := 0;
    duplicate_pending_rows BIGINT := 0;
BEGIN
    IF to_regclass('public.group_join_requests') IS NULL THEN
        RAISE EXCEPTION 'table public.group_join_requests is missing';
    END IF;

    SELECT to_regclass('public.idx_group_join_request_pending_unique') IS NOT NULL
    INTO index_exists;

    IF index_exists THEN
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
        INTO index_is_compatible;

        IF index_is_compatible THEN
            RETURN;
        END IF;

        RAISE EXCEPTION
            'Existing index public.idx_group_join_request_pending_unique is not a valid UNIQUE partial index on public.group_join_requests(user_id, group_id) WHERE status = ''pending''';
    END IF;

    EXECUTE 'LOCK TABLE public.group_join_requests IN SHARE ROW EXCLUSIVE MODE';

    SELECT
        COUNT(*)::BIGINT,
        COALESCE(SUM(cnt - 1), 0)::BIGINT
    INTO duplicate_pending_keys, duplicate_pending_rows
    FROM (
        SELECT COUNT(*) AS cnt
        FROM public.group_join_requests
        WHERE status = 'pending'
        GROUP BY user_id, group_id
        HAVING COUNT(*) > 1
    ) AS duplicates;

    IF duplicate_pending_keys > 0 THEN
        RAISE EXCEPTION
            'Blocking pending join request duplicates detected in public.group_join_requests: duplicate keys=% and duplicate rows=%',
            duplicate_pending_keys,
            duplicate_pending_rows;
    END IF;

    CREATE UNIQUE INDEX idx_group_join_request_pending_unique
        ON public.group_join_requests (user_id, group_id)
        WHERE status = 'pending';
END $$;
