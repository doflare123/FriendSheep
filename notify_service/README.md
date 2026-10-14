# friendSheep notify_service

`notify_service` развёртывается как отдельный сервер и подключается только к
собственной PostgreSQL. Он владеет тремя независимыми durable-процессами:

- lifecycle scheduling P0.2b: получает lifecycle outbox монолита и вызывает
  идемпотентную lifecycle-команду;
- event reminders P0.3: получает отдельный notification-intent outbox,
  разворачивает schedule revision в jobs `T−24h`, `T−6h`, `T−1h`, разрешает
  актуальных получателей и сохраняет authoritative inbox/read state;
- event statistics P0.4: получает отдельный statistics trigger outbox и вызывает
  идемпотентную команду расчёта монолита. Исходные факты, пары и агрегаты
  пользователей хранятся только в монолите.

Монолит остаётся источником истины для Event/User, участников, пользовательской
авторизации и preferences. Сервисы не импортируют реализацию друг друга и не
обращаются к чужой БД. В P0.3 зарегистрирован только канал `in_app`: успешная
доставка означает атомарное сохранение inbox record и delivery result. Telegram,
FCM/Firebase, email, device registration и пользовательский JWT отсутствуют.

## Запуск

Задайте обязательные `NOTIFY_DATABASE_URL`, `NOTIFY_MONOLITH_BASE_URL` и
`NOTIFY_SERVICE_TOKEN` по образцу `.env.example`, затем выполните:

```text
go run .
```

Процесс не загружает `.env`. Локальные инструменты разработки могут экспортировать
значения из игнорируемого файла до запуска процесса.

Для команды статистики предусмотрены отдельные необязательные
`NOTIFY_STATISTICS_HTTP_TIMEOUT=30s` и `NOTIFY_STATISTICS_JOB_LEASE_DURATION=2m`;
lease должен превышать HTTP timeout. Source polling сохраняет общий короткий
`NOTIFY_MONOLITH_HTTP_TIMEOUT`. Перед production deployment измерьте обработку
максимального мероприятия на целевой инфраструктуре с запасом к timeout.

## HTTP API

- `GET /health` сообщает, что HTTP-процесс работает.
- `GET /ready` проверяет собственную БД сервиса. Временная недоступность
  монолита видна в retry logs, но не делает локальную БД неготовой.
- `/internal/v1/users/{userId}/notifications`, `/unread-count` и
  `PATCH /internal/v1/users/{userId}/notifications/{notificationId}/read`
  доступны только с `X-Internal-Token`. Их вызывает монолит после проверки
  пользовательского JWT; сам `notify_service` JWT не принимает.

Стабильная cursor pagination использует `(created_at DESC, id DESC)`. Отметка
прочитанным проверяет ownership, идемпотентна и не заменяет исходный `readAt`.

Подробности ingestion, окон доставки, state machines, retry, lease, каналов и
восстановления после простоя описаны в `docs/architecture.md`. Межсервисные
контракты и общие versioned fixtures находятся в `../backend/docs/internal/`.
