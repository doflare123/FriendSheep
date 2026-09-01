# Durable lifecycle scheduling P0.2b и event reminders P0.3a

Последовательность запуска намеренно линейна:

1. загрузить и проверить конфигурацию окружения;
2. создать JSON-logger на `slog`;
3. подключиться к PostgreSQL, принадлежащей `notify_service`;
4. под advisory lock применить встроенные версионированные миграции;
5. собрать application-слои lifecycle и reminders, зарегистрировав для
   reminders только `in_app` channel adapter;
6. создать независимые source clients, pollers и workers;
7. запустить оба manager и HTTP server;
8. при `SIGINT`/`SIGTERM` отменить общий context, дождаться HTTP и активной
   worker-операции в пределах shutdown timeout, затем закрыть БД.

## Source consumer

Poller вызывает только authenticated endpoint монолита:

`GET /internal/v1/event-lifecycle/schedule-events?after=<sequence>&limit=<limit>`

Заголовок `X-Internal-Token` содержит `NOTIFY_SERVICE_TOKEN`; redirects запрещены,
токен и полные headers не журналируются. Batch обязан быть строго упорядочен по
возрастающему sequence, начинаться после exclusive cursor и иметь согласованный
`nextCursor`. Пустой `hasMore=true` считается malformed response и уходит в
bounded exponential backoff, поэтому busy loop невозможен.

Применение batch, запись дедупликационного receipt по `message_id`, upsert/cancel
job и продвижение cursor выполняются одной транзакцией. Cursor row блокируется
`FOR UPDATE`, поэтому несколько poller-реплик безопасно переигрывают один batch.
Sequence старее job не меняет состояние; cancel отсутствующего event создаёт
tombstone; новый upsert с большим sequence может перепланировать job.

Delivery source имеет семантику at-least-once. Обработка effectively-once
обеспечивается durable cursor, глобальным receipt `message_id`, source sequence и
идемпотентными командами монолита. Полная Event-модель не копируется.

## Job state machine

- `scheduled` ждёт `start_time`;
- `processing` означает короткий lease и HTTP-вызов вне DB transaction;
- `active_wait` ждёт `end_time` после `started`/`already_active`;
- `retry_wait` хранит attempt count и `next_attempt_at`;
- `completed`, `cancelled` и `terminal_failed` не claim-ятся.

Claim использует короткую транзакцию с `FOR UPDATE SKIP LOCKED`, затем lease token.
DB transaction не удерживается во время HTTP. Истёкший `processing` lease снова
доступен worker, поэтому аварийный процесс или restart не теряет job. Ack/retry
обновляют строку только при совпадении lease token; stale worker не перезаписывает
новое расписание или cancel. Lease duration обязан превышать HTTP timeout.

Worker вызывает только:

`POST /internal/v1/events/{eventId}/lifecycle/advance`

Он не передаёт target status или клиентское `now`. Outcomes `started` и
`already_active` планируют `end_time`; `completed`, `completed_catch_up` и
`already_completed` завершают job; `not_due` возвращает job к известному времени.
HTTP 404 и стабильные 400/409 ошибки terminal. Network error, timeout, 429 и 5xx
получают bounded exponential backoff с jitter. 401/403 также используют capped
retry без быстрого цикла и безопасный error code; secret не попадает в logs.
Маленького max-attempts нет: временная недоступность не превращается в потерянный
transition.

## Recovery и наблюдаемость

После простоя poller продолжает с persisted cursor, worker подхватывает due или
expired-lease jobs. Если уже наступил `end_time`, один lifecycle call позволяет
монолиту выполнить `completed_catch_up`. Attempts и `next_attempt_at` переживают
restart.

Structured logs содержат event/job/source sequence, operation или outcome,
attempt, безопасный error code и duration. `/ready` проверяет собственную DB и
тем самым собранную lifecycle dependency. Недоступный монолит виден в retry logs,
но не делает локальную DB неготовой.

Монолит остаётся единственным authority статусов и переходов; `notify_service` —
authority расписания и retry. Outbox пока не очищается: authenticated ack/retention
watermark остаётся follow-up, и до него source records нельзя удалять.

## Граница каналов уведомлений

`application.NotificationChannel` остаётся независимым исходящим port. P0.2b не
регистрирует реализации и никогда его не вызывает.

Для будущего Telegram-среза выбрана единственная стратегия: внешний bot gateway,
доступный через аутентифицированный HTTP-адаптер с ограниченным временем ожидания.
Сервис не будет содержать конкурирующий прямой клиент Telegram Bot API.
Конфигурация и учётные данные gateway появятся только вместе с тестируемым
вертикальным срезом.

FCM остаётся полностью неинициализированным, пока отдельный срез device identity
и FCM-доставки не определит владение, конфигурацию, хранение, retry и тесты.

## Event reminder source consumer P0.3

Reminder poller использует отдельный endpoint и отдельные cursor/receipts:

`GET /internal/v1/notification-intents/event-reminders?after=<sequence>&limit=<limit>`

Versioned `event_reminder` intent не содержит получателей. `schedule_upsert`
содержит `eventId`, `startTime` и числовые offsets `[1440,360,60]`; sequence
является schedule revision. `schedule_cancel` создаёт tombstone без Event FK.
Batch и cursor применяются одной транзакцией, сообщения дедуплицируются по
`source_name + message_id`, а более старый sequence не может воскресить
перенесённое или отменённое расписание.

Один upsert создаёт три строки с уникальностью
`event_id + source_revision + reminder_offset_minutes`:

- 1440: `due_at=T−24h`, окно до `T−6h`;
- 360: `due_at=T−6h`, окно до `T−1h`;
- 60: `due_at=T−1h`, окно до `T`.

Jobs имеют состояния `scheduled`, `processing`, `retry_wait`, `completed`,
`cancelled`, `superseded`, `skipped_missed_window`, `expired` и
`terminal_failed`. Поздний worker пропускает завершившиеся окна, доставляет
только текущее и не создаёт reminders после начала Event. Claim использует
`FOR UPDATE SKIP LOCKED` и lease; HTTP recipient resolution выполняется вне DB
transaction. Reschedule снимает старые незавершённые jobs, а cancel делает все
предыдущие occurrences недоступными для claim.

## Recipient resolution и inbox

Перед delivery worker запрашивает текущий snapshot:

`GET /internal/v1/event-reminders/{eventId}/recipients?reminderOffsetMinutes=<offset>`

Монолит заново читает участников и применяет единый preference port. Поэтому
вышедший участник исключается, а присоединившийся может получить оставшиеся
occurrences. Default policy включает все три offset и только `in_app`.

Логическая `Notification` не зависит от канала. Её idempotency key включает
kind/event/revision/offset/user. После recipient resolution reminder worker в
одной короткой транзакции создаёт notification, уникальные delivery targets по
`notification_id + channel_code` и переводит job в `completed`. Это состояние
означает durable materialization и больше не зависит от результатов внешней
доставки. `in_app` target сразу получает `delivered` в этой же транзакции, потому
что inbox уже является самой доставкой и не требует сетевого вызова.

Остальные targets проходят независимо через `pending`, `processing`,
`retry_wait`, `delivered` или `terminal_failed`. Отдельный dispatcher делает
короткий claim через `FOR UPDATE SKIP LOCKED`, фиксирует lease и только после
commit вызывает `DeliveryChannel.Deliver`. Результат сохраняет application-слой
отдельной lease-token-операцией. Поэтому сбой одного канала не откатывает inbox,
materialized reminder job или другой канал. Истёкший lease восстанавливает
at-least-once dispatch с тем же сохранённым idempotency key; effectively-once
внешняя доставка возможна только при поддержке этого ключа провайдером.

Authoritative inbox доступен только через internal-token API. Публичный JWT
проверяет монолит, вычисляет user ID и проксирует list/unread-count/mark-read.
Ownership проверяется SQL-условием `notification.id + user_id`; повторный
mark-read сохраняет первоначальный `read_at`.

Retryable ошибки delivery получают persisted bounded exponential backoff на
конкретном target. Неизвестный или терминально отказавший channel завершает
только свой target безопасным error code. Удалённый Event и invalid recipient
response по-прежнему терминально завершают reminder job до materialization. Logs
содержат только технические идентификаторы, outcome, attempt, duration и error
code.

## Открытые release gates

- PostgreSQL-only cursor serialization, `SKIP LOCKED`, lease recovery,
  constraints и полный cross-process black-box подтверждены локально на
  disposable database через `NOTIFY_SERVICE_TEST_POSTGRES_DSN` и
  `FRIENDSHEEP_TEST_POSTGRES_DSN`; те же проверки должны стать обязательными в
  CI/release job.
- Notification outbox не очищается до появления authenticated ack/retention
  watermark.
- Первый production bootstrap монолита остаётся release gate до schema freeze и
  создания clean PostgreSQL baseline.
