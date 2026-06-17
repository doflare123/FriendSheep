CREATE TABLE IF NOT EXISTS public.group_action_types (
    id bigserial PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL
);

INSERT INTO public.group_action_types (code, name)
VALUES
    ('create_group', 'Создание группы'),
    ('update_group', 'Изменение группы'),
    ('join_group', 'Вступление в группу'),
    ('create_join_request', 'Создание заявки на вступление'),
    ('leave_group', 'Выход из группы'),
    ('send_invite', 'Создание приглашения'),
    ('accept_invite', 'Принятие приглашения'),
    ('reject_invite', 'Отклонение приглашения'),
    ('approve_request', 'Одобрение заявки'),
    ('reject_request', 'Отклонение заявки'),
    ('approve_all_requests', 'Одобрение всех заявок'),
    ('reject_all_requests', 'Отклонение всех заявок'),
    ('add_operator', 'Назначение оператора'),
    ('remove_operator', 'Снятие оператора'),
    ('change_role', 'Изменение роли'),
    ('ban_user', 'Удаление участника'),
    ('unban_user', 'Удаление из черного списка'),
    ('create_event', 'Создание события'),
    ('update_event', 'Изменение события'),
    ('delete_event', 'Удаление события'),
    ('join_event', 'Вступление в событие'),
    ('leave_event', 'Выход из события'),
    ('kick_from_event', 'Удаление из события')
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name;

DO $$
BEGIN
    IF to_regclass('public.group_action_logs') IS NULL THEN
        RAISE EXCEPTION 'table public.group_action_logs is missing';
    END IF;

    ALTER TABLE public.group_action_logs
        ADD COLUMN IF NOT EXISTS action_type_id bigint,
        ADD COLUMN IF NOT EXISTS target_user_id bigint,
        ADD COLUMN IF NOT EXISTS entity_id bigint,
        ADD COLUMN IF NOT EXISTS entity_name text NOT NULL DEFAULT '';

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'group_action_logs'
          AND column_name = 'action'
    ) THEN
        UPDATE public.group_action_logs log
        SET action_type_id = action_type.id
        FROM public.group_action_types action_type
        WHERE log.action = action_type.code
          AND log.action_type_id IS NULL;
    END IF;

    DELETE FROM public.group_action_logs
    WHERE action_type_id IS NULL;

    ALTER TABLE public.group_action_logs
        ALTER COLUMN action_type_id SET NOT NULL;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_group_action_logs_action_type'
          AND conrelid = 'public.group_action_logs'::regclass
    ) THEN
        ALTER TABLE public.group_action_logs
            ADD CONSTRAINT fk_group_action_logs_action_type
            FOREIGN KEY (action_type_id)
            REFERENCES public.group_action_types(id)
            ON UPDATE CASCADE
            ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_group_action_logs_target_user'
          AND conrelid = 'public.group_action_logs'::regclass
    ) THEN
        ALTER TABLE public.group_action_logs
            ADD CONSTRAINT fk_group_action_logs_target_user
            FOREIGN KEY (target_user_id)
            REFERENCES public.users(id)
            ON UPDATE CASCADE
            ON DELETE SET NULL;
    END IF;

    ALTER TABLE public.group_action_logs
        DROP COLUMN IF EXISTS action;
END $$;
