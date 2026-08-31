# Notification inbox/read contract v1

`notify_service` является единственным владельцем notification records,
delivery attempts и `readAt`. Пользовательский JWT проверяет только монолит;
между серверами используется `X-Internal-Token`. Внутренние endpoints не
предназначены для браузера или мобильного клиента.

## Internal API notify_service

Все маршруты требуют корректный `X-Internal-Token`:

- `GET /internal/v1/users/{userId}/notifications?after=&limit=&unread=`;
- `GET /internal/v1/users/{userId}/notifications/unread-count`;
- `PATCH /internal/v1/users/{userId}/notifications/{notificationId}/read`.

List сортирует записи по `(createdAt DESC, id DESC)`. `after` — непрозрачный ID
последней записи предыдущей страницы; он валиден только для того же владельца.
`limit` имеет диапазон 1..100. `unread=true` оставляет записи с `readAt=null`.
Ответ имеет форму `{items,nextCursor,hasMore}`.

`event_reminder` содержит channel-independent envelope:

```json
{
  "id": "stable-random-id",
  "kind": "event_reminder",
  "resourceType": "event",
  "resourceId": 42,
  "reminderOffsetMinutes": 360,
  "scheduleRevision": 17,
  "payload": {
    "schemaVersion": 1,
    "eventId": 42,
    "title": "Вечерняя встреча",
    "startTime": "2036-09-01T18:00:00Z"
  },
  "createdAt": "2036-09-01T12:00:00Z",
  "readAt": null
}
```

Mark-read обновляет строку только при совпадении `id + userId`. Чужой или
несуществующий ID возвращает `404 notification_not_found`. Повторный запрос
успешен и сохраняет первоначальный `readAt`.

## Public API монолита

- `GET /api/v2/users/me/notifications?cursor=&limit=&unread=`;
- `GET /api/v2/users/me/notifications/unread-count`;
- `PATCH /api/v2/users/me/notifications/{notificationId}/read`.

Монолит вычисляет user ID из проверенного JWT и никогда не читает его из query,
body или headers клиента. `cursor` публичного API передаётся как internal
`after`. Upstream errors преобразуются в общий безопасный error DTO без token,
headers, DSN или внутреннего response body.

## Идемпотентность

Логический ключ notification включает
`kind + eventId + scheduleRevision + reminderOffsetMinutes + userId`.
Delivery boundary дополнительно включает channel code. Поэтому разные offsets и
новая revision создают самостоятельные записи, а replay/crash одной occurrence
не создаёт второй inbox record или второй успешный delivery result.
