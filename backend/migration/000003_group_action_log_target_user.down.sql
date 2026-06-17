ALTER TABLE IF EXISTS public.group_action_logs
    ADD COLUMN IF NOT EXISTS action text NOT NULL DEFAULT '';

UPDATE public.group_action_logs log
SET action = action_type.code
FROM public.group_action_types action_type
WHERE log.action_type_id = action_type.id;

ALTER TABLE IF EXISTS public.group_action_logs
    DROP CONSTRAINT IF EXISTS fk_group_action_logs_target_user;

ALTER TABLE IF EXISTS public.group_action_logs
    DROP CONSTRAINT IF EXISTS fk_group_action_logs_action_type;

ALTER TABLE IF EXISTS public.group_action_logs
    DROP COLUMN IF EXISTS target_user_id,
    DROP COLUMN IF EXISTS action_type_id,
    DROP COLUMN IF EXISTS entity_id,
    DROP COLUMN IF EXISTS entity_name;

DROP TABLE IF EXISTS public.group_action_types;
