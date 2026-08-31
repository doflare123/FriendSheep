# Event reminder notification contract v1

P0.3 uses a notification flow separate from event lifecycle scheduling. The
monolith owns events, participants, authentication and reminder preferences;
`notify_service` owns ingestion, reminder jobs, logical notifications, delivery
attempts and read state. Neither service imports implementation code from the
other or connects to the other's database.

## Notification intent source

`GET /internal/v1/notification-intents/event-reminders?after=<sequence>&limit=<limit>`
requires `X-Internal-Token`. `after` is an exclusive cursor, results are ordered
by PostgreSQL-assigned `sequence ASC`, default limit is 100 and maximum is 500.
The response uses `{items,nextCursor,hasMore}` with stable replay semantics.

The single source of truth for the extensible v1 JSON item is
`contracts/event_reminder_intent_v1.schema.json`; example provider/consumer
fixtures are alongside it. The current default provider publishes the three
canonical numeric offsets `[1440,360,60]`, while the field accepts future
positive numeric offsets without a database or application-contract redesign.
`schedule_cancel` omits schedule fields.
The source sequence is also the schedule revision. No recipients or personal
data are published in the outbox.

Create publishes an upsert in the event transaction. Only a `startTime` change
publishes a new reminder revision; duration-only updates affect lifecycle
scheduling but not reminder scheduling. Delete publishes a cancellation
tombstone before deleting the event aggregate. The outbox has no Event foreign
key and is not cleaned without a future acknowledgement/retention protocol.

## Recipient resolution

`GET /internal/v1/event-reminders/{eventId}/recipients?reminderOffsetMinutes=360`
requires `X-Internal-Token`. It loads the current event and distinct current
participants and applies `EventReminderPreferenceReader` for that exact numeric
offset. The current default policy enables 1440, 360 and 60 minutes and returns
only `in_app`; replacing the preference adapter does not change scheduling or
inbox contracts.

```json
{
  "eventId": 42,
  "title": "Evening meetup",
  "startTime": "2026-09-01T18:00:00Z",
  "reminderOffsetMinutes": 360,
  "recipients": [{"userId": 7, "channels": ["in_app"]}]
}
```

Missing/deleted events return `404 event_not_found`. Invalid offsets return
`400 invalid_reminder_offset`; infrastructure failures return a sanitized 500.

## Inbox proxy

Public endpoints authenticate only in the monolith and never accept `userId`:

- `GET /api/v2/users/me/notifications?cursor=&limit=&unread=`;
- `GET /api/v2/users/me/notifications/unread-count`;
- `PATCH /api/v2/users/me/notifications/{notificationId}/read`.

The monolith injects the JWT-derived user ID into authenticated internal calls
to `notify_service`. Mark-read is ownership checked and idempotent; an existing
`readAt` is never replaced. Internal details and secrets are not returned.

The authoritative logical notification is channel-independent and records the
event resource, versioned v1 payload, offset and schedule revision. Delivery attempts add
the channel code; only `in_app` is registered in P0.3. Logical idempotency covers
kind/event/revision/offset/user, and delivery idempotency additionally covers
the channel.

The reminder scheduler deliberately has no `NOTIFY_EVENT_REMINDER_LEAD_TIME`:
the mandatory multi-trigger correction replaced the earlier single-lead design.
Canonical provider values are numeric 1440/360/60 minute offsets, while the v1
contract and job schema remain extensible to other positive offsets.

## Occurrences and states

One upsert expands into independent jobs for `T-24h`, `T-6h`, and `T-1h`.
Their delivery windows end at `T-6h`, `T-1h`, and `T`. The application clock
classifies an occurrence as `scheduled`, `processing`, `retry_wait`,
`completed`, `superseded`, `cancelled`, `skipped_missed_window`, `expired`, or
`terminal_failed`. Late recovery skips prior windows and can deliver only the
currently active occurrence; nothing is delivered after `startTime`.

Reschedule supersedes unfinished jobs from older revisions. Cancel supersedes
all earlier revisions and persists a non-claimable tombstone. Notifications
already delivered by an old revision remain in inbox.

Notification and successful `in_app` delivery creation share the same DB
transaction as the job completion. Recipient resolution occurs before that
transaction, so no database transaction is held during HTTP.

The detailed inbox/read transport is documented in
`notification_inbox_contract.md`.
