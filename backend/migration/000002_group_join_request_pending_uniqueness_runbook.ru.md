# Ранбук: rollout `000002_group_join_request_pending_uniqueness`

## Для кого

Этот документ для:
- администратора/DevOps, который запускает rollout SQL-миграции;
- backend-разработчика, который сопровождает rollout по join requests;
- тестировщика, который фиксирует preflight/postflight результаты.

## Назначение

Миграция `migration/000002_group_join_request_pending_uniqueness.up.sql` добавляет Postgres partial unique index:

- `idx_group_join_request_pending_unique`
- на `public.group_join_requests(user_id, group_id)`
- с предикатом `WHERE status = 'pending'`

Цель: закрыть race в `JoinGroup` для приватных групп, где две конкурентные транзакции могли создать две `pending`-заявки для одной пары `(user_id, group_id)`.

Важно:
- миграция не удаляет и не сливает blocker-строки автоматически;
- если уже существуют дубли `pending`, rollout должен остановиться до `up`;
- `down.sql` снимает только индекс и не восстанавливает проблемные данные.

## Порядок выполнения

Перед запуском убедитесь, что заданы DB-переменные окружения (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).

Рекомендуемый CLI-путь:
- полный цикл: `go run ./cmd/migrate-rollout all 000002_group_join_request_pending_uniqueness`
- только preflight: `go run ./cmd/migrate-rollout preflight 000002_group_join_request_pending_uniqueness`
- только migration up: `go run ./cmd/migrate-rollout up 000002_group_join_request_pending_uniqueness`
- только postflight: `go run ./cmd/migrate-rollout postflight 000002_group_join_request_pending_uniqueness`

CLI поддерживает и режим без второго аргумента, тогда будет выбран последний `.up.sql` rollout. Для production rollout этого индекса рекомендуется указывать base явно, чтобы не уехать на более позднюю миграцию.

1. Запустите `migration/000002_group_join_request_pending_uniqueness_preflight.sql` под тем же DB-пользователем, который будет выполнять `up`.
2. Если в preflight есть `FAIL`, rollout останавливается.
3. Для детализации blocker-строк запустите `migration/000002_group_join_request_pending_uniqueness_rollout_checks.sql`.
4. Если preflight зелёный, запланируйте low-write окно: `up` берёт `LOCK TABLE ... IN SHARE ROW EXCLUSIVE MODE` на `public.group_join_requests` и создаёт обычный `CREATE UNIQUE INDEX`.
5. Выполните `migration/000002_group_join_request_pending_uniqueness.up.sql` один раз.
6. Сразу выполните `migration/000002_group_join_request_pending_uniqueness_postflight.sql`.
7. Если в postflight есть `FAIL`, rollout не закрывать, разбирать инварианты и фактическое состояние индекса.
8. Если postflight полностью `PASS`, заархивировать вывод preflight/postflight и вывод `rollout_checks.sql`.

## Blocker checks

- Любой `FAIL` в preflight/postflight.
- Отсутствие таблицы `public.group_join_requests` или обязательных колонок `id`, `user_id`, `group_id`, `status`.
- Несовместимый уже существующий индекс с именем `idx_group_join_request_pending_unique`.
- Отсутствие прав на чтение таблицы или отсутствие владения таблицей для создания индекса.
- Наличие строк-дублей среди `status = 'pending'`.

## Как читать blocker duplicates

`rollout_checks.sql` выводит:
- пары `(user_id, group_id)` с количеством дублирующих `pending`-строк;
- список `request_ids` и `created_at_values` для точечной ручной разборки;
- timeline по всем заявкам пользователя в группу, чтобы отделить допустимые исторические `approved/rejected` записи от blocker `pending`.

Решение по cleanup должно быть ручным и осознанным: индекс запрещает только повторные `pending`, но не запрещает хранить историю `approved/rejected` заявок для той же пары.
