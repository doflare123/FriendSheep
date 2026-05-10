-- duplicate group_users with chosen keeper
WITH ranked AS (
    SELECT
        gu.id,
        gu.user_id,
        gu.group_id,
        COALESCE(rig.name, '<NULL>') AS role_name,
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
),
keepers AS (
    SELECT user_id, group_id, id AS kept_id, role_name AS kept_role
    FROM ranked
    WHERE rn = 1
),
losers AS (
    SELECT
        user_id,
        group_id,
        COUNT(*) AS removed_rows,
        ARRAY_AGG(id ORDER BY id) AS removed_ids,
        ARRAY_AGG(role_name ORDER BY id) AS removed_roles
    FROM ranked
    WHERE rn > 1
    GROUP BY user_id, group_id
)
SELECT l.user_id, l.group_id, k.kept_id, k.kept_role, l.removed_rows, l.removed_ids, l.removed_roles
FROM losers l
JOIN keepers k ON k.user_id = l.user_id AND k.group_id = l.group_id
ORDER BY l.removed_rows DESC, l.group_id, l.user_id;

-- unexpected roles in duplicate group memberships
WITH duplicate_pairs AS (
    SELECT user_id, group_id
    FROM public.group_users
    GROUP BY user_id, group_id
    HAVING COUNT(*) > 1
)
SELECT COALESCE(rig.name, '<NULL>') AS role_name, COUNT(*) AS duplicate_rows
FROM public.group_users gu
LEFT JOIN public.role_in_groups rig ON rig.id = gu.role_in_group_id
JOIN duplicate_pairs dp ON dp.user_id = gu.user_id AND dp.group_id = gu.group_id
WHERE rig.name IS NULL OR rig.name NOT IN ('Админ', 'Модератор', 'Участник')
GROUP BY COALESCE(rig.name, '<NULL>')
ORDER BY duplicate_rows DESC, role_name;

-- duplicate events_users
SELECT event_id, user_id, COUNT(*) AS duplicate_rows, ARRAY_AGG(id ORDER BY id) AS row_ids
FROM public.events_users
GROUP BY event_id, user_id
HAVING COUNT(*) > 1
ORDER BY duplicate_rows DESC, event_id, user_id;

-- current_users drift
WITH expected AS (
    SELECT e.id AS event_id, COALESCE(COUNT(eu.user_id), 0)::BIGINT AS expected_current_users
    FROM public.events e
    LEFT JOIN public.events_users eu ON eu.event_id = e.id
    GROUP BY e.id
)
SELECT e.event_id, events.current_users AS stored_current_users, e.expected_current_users
FROM expected e
JOIN public.events ON events.id = e.event_id
WHERE events.current_users IS DISTINCT FROM e.expected_current_users
ORDER BY e.event_id;

-- existing membership indexes
SELECT schemaname, tablename, indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'public'
  AND tablename IN ('group_users', 'events_users')
ORDER BY tablename, indexname;

-- postflight: duplicate keys must be zero
SELECT 'group_users' AS table_name, COUNT(*) AS duplicate_keys
FROM (
    SELECT user_id, group_id
    FROM public.group_users
    GROUP BY user_id, group_id
    HAVING COUNT(*) > 1
) d
UNION ALL
SELECT 'events_users' AS table_name, COUNT(*) AS duplicate_keys
FROM (
    SELECT event_id, user_id
    FROM public.events_users
    GROUP BY event_id, user_id
    HAVING COUNT(*) > 1
) d;
