# P0.3b — надёжность lifecycle/reminder pipeline

Проверено локально 11 октября 2026 на disposable PostgreSQL 18; базы двух
серверов раздельные. Исходные пользовательские изменения сохранены.

## Устранённые сценарии отказа

1. Поздний commit A с меньшим sequence больше не остаётся позади cursor B.
   Общий PostgreSQL transaction advisory lock `(1179862864, 1)` берётся в event
   UOW до domain locks и выдачи sequence обоих outbox, удерживается до
   commit/rollback. Это работает между экземплярами сервера и сохраняет request
   context. Rollback gaps не задерживают следующий batch. Sequence требуют
   `CACHE 1`; новые издатели обязаны соблюдать тот же UOW и порядок блокировок.
   Компромисс — сериализация текущих event UOW, включая membership/admin.
2. Reminder, актуальный до HTTP или ожидания job lock, больше не материализуется
   после закрытия окна. Worker повторяет проверку после HTTP, store читает
   `Clock.Now()` после `FOR UPDATE` и guards lease/state/revision. При равенстве
   T−6h/T−1h завершается прежнее окно как `skipped_missed_window`, при равенстве
   T — `expired`, атомарно без notifications/targets. Решение предшествует
   первой вставке; завершение commit до временной границы не гарантируется.
3. Корректная source страница из 500 записей больше 64 КиБ теперь принимается
   обоими HTTP clients. Лимит 1 MiB покрывает максимальную фиксированную форму v1
   с запасом (менее 256 KiB), чтение ограничено limit+1. Превышение явно
   распознаётся как retryable `response_too_large`, повреждённый JSON как
   `malformed_response`; частичный batch и продвижение cursor исключены.

## Подтверждённые проверки

- Backend focused: `go test ./services/events ./handlers ./tests -run
  'Test(EventOutboxPublication|EventCommand|LifecycleScheduleOutbox|EventReminderIntent)' -count=2`.
- Notify focused: `go test -buildvcs=false ./internal/reminders ./internal/lifecycle -count=2`.
- Отдельный verbose notify regression с `-count=2` подтвердил выполнение
  PostgreSQL ingestion/idempotency, duplicate pollers, SKIP LOCKED, lease recovery,
  HTTP pages/ошибок, controlled HTTP/lock-wait часов и exact boundaries. SKIP отсутствует.
- Backend commit-order regression с `-count=2 -v`: commit A, rollback A с gaps,
  два читателя/replay и context cancellation при ожидании advisory lock прошли.
  Управляемые каналы и наблюдаемые PostgreSQL locks заменяют случайные sleep.
- Полные `go test ./... -count=1`, `go vet ./...` монолита прошли.
- Notify `go test -buildvcs=false ./... -count=1`, `go vet -buildvcs=false ./...`
  и `go build -buildvcs=false ./...` прошли.
- `go test ./tests -run '^TestP03EventReminderCrossServiceBlackBoxPostgres$'
  -count=2 -v` прошёл дважды с реальным процессом notify_service и authenticated HTTP.
- Отдельный diff review не нашёл блокирующих дефектов; UTF-8 и mojibake проверены.

Оба PostgreSQL test DSN были заданы. Требуемые PostgreSQL и cross-process проверки
не пропущены. Первый параллельный полный прогон монолита временно исключал
black-box, чтобы не пересечь очистку notify schema с notify suites; после них
black-box выполнен отдельно дважды и монолит повторно проверен целиком.

## Сохранённые границы и ограничения

Монолит остаётся владельцем событий, notify_service — scheduling/inbox/targets.
Интеграция — authenticated versioned HTTP; materialization атомарна, HTTP и
DeliveryChannel выполняются вне транзакций. Зарегистрирован только in_app.
Dispatch остаётся at-least-once; effectively-once внешняя доставка требует
provider idempotency. Абсолютный exactly-once не заявляется.

Локальные результаты не заменяют обязательный CI/release прогон. Clean PostgreSQL
baseline монолита остаётся отдельным release gate. Статистика, новые каналы,
пользовательские настройки и новости не входят в P0.3b.
