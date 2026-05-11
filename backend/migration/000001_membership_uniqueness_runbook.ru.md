# Ранбук: rollout `000001_membership_uniqueness`

## Для кого

Этот документ для:
- администратора/DevOps, который запускает миграцию;
- бэкенд-разработчика, который сопровождает rollout;
- тестировщика, который проверяет preflight/postflight результаты.

## Назначение

Миграция `migration/000001_membership_uniqueness.up.sql` делает три вещи:

1. удаляет дубликаты из `public.group_users` и `public.events_users`, записывая метаданные удалений в `public.membership_dedupe_audit`;
2. пересчитывает `public.events.current_users` по фактическому составу `public.events_users`;
3. приводит индексы к контракту:
   - `idx_group_user_membership` на `(user_id, group_id)`
   - `idx_event_user_membership` на `(event_id, user_id)`

`down.sql` удаляет только схемные артефакты (индексы и audit-объекты). Удаленные дубликаты и прежние счетчики он не восстанавливает.

## Порядок выполнения

Перед запуском команд убедитесь, что заданы DB-переменные окружения (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).

Рекомендуемый способ запуска через отдельный CLI:
- полный цикл: `go run ./cmd/migrate-rollout all`
- только preflight: `go run ./cmd/migrate-rollout preflight`
- только migration up: `go run ./cmd/migrate-rollout up`
- только postflight: `go run ./cmd/migrate-rollout postflight`

0. Для non-DEV startup-пути включайте `ENABLE_STARTUP_SQL_MIGRATIONS=true` только в окно планового rollout.
1. Запустите `migration/000001_membership_uniqueness_preflight.sql` под тем же DB-пользователем, который будет запускать миграцию.
2. Если в preflight есть `FAIL` — rollout останавливается.
3. Если есть только `WARN` — сохраните вывод и продолжайте.
4. При необходимости детализации запустите `migration/000001_membership_uniqueness_rollout_checks.sql`.
5. Запланируйте low-write окно: миграция берет блокировки таблиц membership, делает `DELETE/UPDATE` и обычный `CREATE UNIQUE INDEX`.
6. Выполните `migration/000001_membership_uniqueness.up.sql` один раз.
7. Сразу выполните `migration/000001_membership_uniqueness_postflight.sql`.
8. Если в postflight есть `FAIL` — rollout не закрывать, разбирать инварианты.
9. Если postflight полностью `PASS` — заархивировать вывод preflight/postflight и количество строк в `membership_dedupe_audit`.

## Блокеры

- Любой `FAIL` в preflight/postflight.
- Отсутствие обязательных таблиц/колонок.
- Несовместимые уже существующие индексы с целевыми именами.
- Отсутствие ролей `Админ`, `Модератор`, `Участник` в `role_in_groups`.
- Наличие неожиданных role values внутри дублей `group_users(user_id, group_id)`.
- Недостаточные права пользователя миграции (`SELECT/DELETE/UPDATE/INSERT/CREATE`).

## Smoke-check startup-пути

1. `есть SQL миграции + ENABLE_STARTUP_SQL_MIGRATIONS=false + схема полная`:
   сервис должен стартовать и логировать, что startup SQL migration path пропущен.
2. `есть SQL миграции + ENABLE_STARTUP_SQL_MIGRATIONS=false + схема неполная`:
   сервис должен завершиться с явной ошибкой о необходимости включить флаг для планового rollout.
3. `есть SQL миграции + ENABLE_STARTUP_SQL_MIGRATIONS=true`:
   сервис должен пройти migration path и продолжить обычный старт.
