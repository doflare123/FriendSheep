# Durable lifecycle scheduling P0.2b

Последовательность запуска намеренно линейна:

1. загрузить и проверить конфигурацию окружения;
2. создать JSON-logger на `slog`;
3. подключиться к PostgreSQL, принадлежащей `notify_service`;
4. под advisory lock применить встроенные версионированные миграции;
5. собрать application-слой без пользовательских delivery channels;
6. создать lifecycle source client, poller и worker;
7. запустить workers и HTTP server;
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
