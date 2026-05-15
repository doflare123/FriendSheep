-- blocker rows: duplicate pending join requests
SELECT
    user_id,
    group_id,
    COUNT(*) AS duplicate_rows,
    ARRAY_AGG(id ORDER BY created_at, id) AS request_ids,
    ARRAY_AGG(created_at ORDER BY created_at, id) AS created_at_values
FROM public.group_join_requests
WHERE status = 'pending'
GROUP BY user_id, group_id
HAVING COUNT(*) > 1
ORDER BY duplicate_rows DESC, group_id, user_id;

-- same user/group timeline with statuses
SELECT
    user_id,
    group_id,
    ARRAY_AGG(format('id=%s status=%s created_at=%s', id, status, created_at) ORDER BY created_at, id) AS request_timeline
FROM public.group_join_requests
GROUP BY user_id, group_id
HAVING COUNT(*) FILTER (WHERE status = 'pending') > 0
ORDER BY group_id, user_id;

-- existing indexes on group_join_requests
SELECT schemaname, tablename, indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'public'
  AND tablename = 'group_join_requests'
ORDER BY indexname;

-- postflight: pending duplicate keys must be zero
SELECT COUNT(*) AS pending_duplicate_keys
FROM (
    SELECT user_id, group_id
    FROM public.group_join_requests
    WHERE status = 'pending'
    GROUP BY user_id, group_id
    HAVING COUNT(*) > 1
) AS duplicates;
