# Internal event lifecycle schedule outbox

Монолит остаётся владельцем мероприятий и атомарно публикует только минимальные
изменения их расписания. `notify_service` получает эти записи по внутреннему HTTP
API и не читает базу данных монолита напрямую.

## Transactional outbox

Таблица `event_lifecycle_schedule_outbox` создаётся versioned migration
`000004_event_lifecycle_schedule_outbox`. Запись добавляется в ту же транзакцию,
что и изменение мероприятия:

- create — `schedule_upsert`;
- изменение `start_time` или вычисляемого `end_time` — `schedule_upsert`;
- delete — `schedule_cancel`.

`sequence` назначается PostgreSQL identity и задаёт единственный порядок
применения. `message_id` — уникальный UUID для дедупликации. Текущая
`schema_version` равна `1`. Upsert содержит `start_time` и `end_time`; cancel
сохраняет `event_id` и не имеет внешнего ключа на Event, поэтому tombstone
остаётся доступным после удаления мероприятия. Ошибка записи outbox откатывает
event mutation, а откат event transaction откатывает outbox.

## Read API

```http
GET /internal/v1/event-lifecycle/schedule-events?after=123&limit=100
X-Internal-Token: <service-token>
```

Используется тот же `NOTIFY_SERVICE_TOKEN` и constant-time middleware, что и для
lifecycle command. Пользовательский JWT не даёт доступ. Значение токена, полные
headers и payload нельзя журналировать.

`after` — exclusive unsigned sequence cursor. Записи всегда возвращаются по
`sequence ASC`. Default limit — `100`, maximum — `500`; нулевой, отрицательный,
нечисловой или превышающий maximum limit возвращает `400 invalid_limit`.
Некорректный cursor возвращает `400 invalid_after`.

```json
{
  "items": [
    {
      "sequence": 124,
      "messageId": "f69d7f57-a785-456c-b08e-518035e77cba",
      "schemaVersion": 1,
      "operation": "schedule_upsert",
      "eventId": 42,
      "startTime": "2026-08-27T18:00:00Z",
      "endTime": "2026-08-27T21:00:00Z",
      "occurredAt": "2026-08-27T12:00:00Z"
    }
  ],
  "nextCursor": 124,
  "hasMore": false
}
```

Для пустой страницы `items` — пустой массив, а `nextCursor` равен переданному
`after`. `hasMore` означает, что сразу после последнего возвращённого sequence
есть ещё запись. Повторный запрос с тем же cursor возвращает тот же ordered batch,
пока outbox не очищен.

## Delivery, replay и retention

Доставка source messages имеет семантику at-least-once. Effectively-once
обработка достигается transactionally persisted cursor и сравнением source
sequence/message ID в `notify_service`; authority перехода статуса остаётся у
идемпотентной lifecycle command монолита.

Автоматическая очистка outbox в P0.2b отсутствует. Источником подтверждённого
durable ingestion является cursor в отдельной БД `notify_service`, а безопасного
ack/retention watermark API пока нет. До появления такого протокола записи нельзя
удалять. Операционный follow-up: определить authenticated acknowledgement,
наблюдаемый lag и retention window с запасом для восстановления consumer из
backup.

Пользовательские уведомления, reminders, inbox, внешние каналы и статистика не
входят в этот контракт.
