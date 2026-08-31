# Backend Agent Context

## Project

This repository is the Go backend for `friendSheep`, a social/group/event service.

Main stack:

- Go + Gin HTTP API.
- GORM-backed Postgres repository facade in `repository/postgresRepository.go`.
- Redis sessions/cache in `repository/redisRepository.go` and `sessions/`.
- Swagger docs generated under `docs/`.
- Tests live mainly in `tests/`, with some package-local tests in `db/` and `sessions/`.

System topology:

- The project also contains `../notify_service`, a separate Go server with its own process, deployment lifecycle, configuration, and failure boundary.
- `notify_service` works with this monolith but is not a package inside it. Never import implementation code across the two servers or treat an in-process call as their integration contract.
- The notification server is intended to own notification scheduling, notification-inbox/delivery state (including read state and external channels), and collection/orchestration of statistics jobs. The monolith remains the source of truth for users, groups, events, authorization, profile/privacy settings, and transactional domain invariants.
- Cross-server behavior must use explicit, authenticated, versioned, idempotent contracts (initially internal HTTP is acceptable; a broker/outbox can be introduced when delivery guarantees require it). Do not create an implicit distributed transaction.
- Prefer the monolith to execute event status mutations and statistics aggregation behind internal application commands, while `notify_service` schedules and invokes those commands and handles delivery. Do not let the notification server become a second owner of the monolith schema.

## Current Architecture Direction

The project is moving from older global/shared repository usage toward narrower feature-level ports.

Do not rewrite the shared repositories in one big change. Treat them as temporary low-level adapters and migrate vertically by feature:

- auth/register/session first,
- then group-role authorization,
- then groups/events,
- then cache/popular-event paths.

Existing good examples:

- `services/auth_repository.go` is a narrow auth repository adapter.
- `sessions/sessionStore.go` uses a local behavior port for session storage.
- `middlewares/group_role_reader.go` is a narrow authorization reader seam.
- `services/groups/group_shared_store.go` contains shared group actor/access/relation seams.
- Group join request/invite/admin slices use narrower local stores instead of direct service-level GORM calls.
- Event command, membership, read, and admin slices use clean service/store or unit-of-work contracts under `services/events/`, with GORM isolated in adapter files.
- Popular events use `PopularEventsStore`, `PopularEventsCache`, `PopularEventsNotifier`, and `PopularEventsScheduler`; GORM, Redis, email, and cron are isolated in adapters.
- Event update validation is owned by the event command service, while reference existence checks are exposed through the clean `EventCommandStore`.
- `services/references/reference_service.go` owns clean reference contracts, while `services/references/gorm_reference_store.go` is the only storage adapter for that slice.

Known architecture debt:

- `repository/postgresRepository.go` leaks GORM API/types across services.
- Redis and Mongo repository layers likely also leak library-specific behavior.
- Role names now have constants in `models/groups/groupRoles.go`, and group access increasingly uses capabilities, but some service paths still pass role strings around.
- Some group storage contracts, especially `services/groups/store.go`, are still GORM-shaped and should be migrated vertically rather than rewritten globally.
- Popular-event application and handler contracts no longer expose repository or library types, but their GORM/Redis adapters still receive the temporary broad repository facades.
- Event and group adapters still write group audit records through local helpers; the audit module boundary and action-type lookup remain architecture debt.
- A final clean-database SQL baseline does not exist yet. Current SQL migrations are draft rollout artifacts from the intermediate schema and are not a complete empty-database bootstrap.

Explicit architecture decision:

- Automatic event status transitions are intentionally not implemented in this monolith. They are planned for a separate service because status changes will be coupled to notification delivery. Do not add a monolith scheduler for event status changes unless this decision is explicitly revised.
- The separate service for that orchestration is `../notify_service`. The scheduler belongs there, but the actual event mutation should go through an authenticated, idempotent monolith application boundary rather than direct cross-service GORM writes.

## Current Product Direction

- Group functionality is complete for the current product scope. Do not reopen it unless notification integration exposes a concrete missing domain event/contract or a new product requirement is supplied.
- Core event functionality, automatic lifecycle transitions and the first
  `in_app` reminder delivery are complete for the current product scope.
- The user area is not restored as a coherent vertical slice. Account deletion,
  own/public profile reads, user search, profile editing, privacy settings,
  profile statistics selection/display and notification preferences still need
  contract and implementation review. Inbox list/unread/mark-read is complete.
- News is also a legacy/inactive slice. Creation, editing, listing/details, comments, comment interactions/authorization, and notification integration need to be rebuilt against the current architecture.
- These product areas should be restored vertically behind narrow feature ports. Do not reactivate old routes and broad GORM-driven services wholesale.

### Current `notify_service` Condition

P0.2b durable lifecycle scheduling в `../notify_service` завершён:

- Legacy scheduler и копии моделей монолита удалены; сервис не читает и не изменяет таблицы монолита.
- Монолит атомарно пишет ordered lifecycle schedule outbox для create/reschedule/delete и предоставляет authenticated pull API; автоматическая очистка outbox отложена до безопасного ack/retention протокола.
- `notify_service` хранит собственные cursor, message receipts и lifecycle jobs, применяет batch и cursor одной транзакцией, claim-ит due jobs через PostgreSQL `FOR UPDATE SKIP LOCKED`/lease и вызывает monolith lifecycle command с persisted retry/backoff.
- Startup разделён на environment config -> собственная PostgreSQL -> версионированные миграции -> application/lifecycle workers -> HTTP server.
- Реализованы `/health`, DB-backed `/ready` с ограниченным временем ожидания, structured JSON logs и graceful shutdown.
- Пользовательские JWT, device registration, FCM и обе старые Telegram-реализации удалены. Будущая Telegram-стратегия — единственный внешний bot gateway за чистым channel port; в P0.3 зарегистрирован только `in_app`, внешняя отправка отсутствует.
- Локальные `.env` и Firebase credential file удалены и игнорируются; legacy compose больше не подключает notify_service к общей БД/Telegram и требует отдельный `NOTIFY_DATABASE_URL`.
- Ротация ранее раскрытых внешних credentials остаётся обязательным operator action до реального deployment; удаление файлов и переписывание Git history не заменяют отзыв у провайдера. Статус зафиксирован в `../notify_service/SECURITY_ROTATION.md`.

P0.3 event reminder intent/inbox slice в `../notify_service` завершён:

- Монолит атомарно пишет отдельный transactional notification-intent outbox для reminder schedule upsert/cancel и отдаёт его через authenticated pull API.
- Reminder scheduling использует versioned numeric offsets `[1440,360,60]`, отдельные cursor/receipts/jobs, late-window semantics и recipient resolution в момент firing, а не при публикации intent.
- `notify_service` хранит authoritative inbox/read state, logical notifications и delivery attempts; идемпотентность logical notification покрывает `kind/event/revision/offset/user`, а delivery attempt дополнительно включает channel.
- Монолит проксирует только JWT-authenticated public inbox APIs (`list`, `unread-count`, `mark-read`) и никогда не принимает клиентский `userId`; внутренние notify endpoints защищены `X-Internal-Token`.
- В P0.3 зарегистрирован только `in_app` channel adapter. Telegram/FCM/email/external delivery по-прежнему отсутствуют и должны приходить отдельными вертикальными срезами за чистым channel port.

## Working Rules

- Preserve user changes. Never run destructive git commands unless explicitly requested.
- Use `apply_patch` for manual edits.
- Run `gofmt` after Go edits.
- Use a local Go cache inside repo when testing:
  `PowerShell: $env:GOCACHE='D:\friendSheep\backend\.cache\go-build'; go test ./...`
- Always clean `.cache` after work:
  `PowerShell: $target = Resolve-Path .\.cache -ErrorAction SilentlyContinue; if ($target -and $target.Path.StartsWith((Resolve-Path .).Path)) { Remove-Item -LiteralPath $target.Path -Recurse -Force }`
- If PowerShell displays Russian as mojibake, verify real bytes with `rg`. Do not introduce mojibake into files.
- After every review/check pass, explicitly verify Russian and other non-ASCII strings for mojibake/encoding corruption before finalizing changes.
- Все новые и изменяемые агентом комментарии в коде должны быть написаны на русском языке. Исключения допустимы только для точных внешних терминов, идентификаторов и цитат из спецификаций, которые нельзя корректно перевести.
- For implementation work, the main agent must orchestrate subagents by default:
  - `golang-pro`/`backend` for backend code changes,
  - dedicated test subagent for writing/updating tests,
  - `reviewer` subagent for final diff review (bugs/regressions/missed coverage).
  If a preferred subagent is unavailable in the current environment, document the fallback and continue with the closest available role.
- For git commands in sandbox, use:
  `git -c safe.directory=D:/friendSheep ...`

## Current State

Deployment and migration context:

- The service is under active, large-scale refactoring and there is no persistent development, staging, or production database whose data or migration history must be preserved.
- Existing SQL migrations have not been applied to a persistent database and may be squashed or replaced before the first deployment.
- During active refactoring, keep GORM models, database invariants, PostgreSQL-specific tests, and schema intent current; do not create a final SQL patch for every transient model change.
- At schema freeze, before the first persistent environment, replace/squash the draft chain into a clean initial-schema baseline plus required reference seed migrations, then verify first and repeated startup on empty PostgreSQL.
- After the first persistent database exists, migration history becomes append-only and applied migrations must not be rewritten.

Recent green checks:

- `go test ./services/groups ./services/events ./handlers -count=1`
- `go test ./tests -run '^(TestGroupService|TestEventsService|TestSeeder)' -count=1`
- `go test ./... -count=1`
- `go test ./... -count=1` was re-run during the June 19, 2026 context check.
- `go test ./tests -run "^(TestEventsServiceGetGroupEvents|TestEventsServiceSearchEvents|TestGroupServiceGetGroupDetailsMarksActiveEventSubscription)$" -count=1`
- `go test ./services/events ./services/groups ./models/dto/... -count=1`
- `go test ./... -count=1` was re-run after the June 22, 2026 group-details/event-subscription DTO work.
- `go test ./... -count=1` was re-run on July 24, 2026 after the event vertical-slice and reference extraction work.
- `go test ./handlers ./tests -run '^(TestReference|TestReferences|TestGORMReference|TestRegisterReferencesRoutes)' -count=2`
- `go test ./services/events ./handlers ./tests -run '^(TestEventCommandService|TestUpdateEvent|TestCreateEvent|Test.*EventInput)' -count=2`
- `go test ./... -count=1` was re-run on July 28, 2026 after the popular-event extraction and event-update validation work.
- Lifecycle/config/HTTP/store/graceful-shutdown focused tests were run twice with `-count=2` on August 15, 2026.
- `go test ./services/events ./handlers ./tests -count=1` and `go test ./... -count=1` were re-run on August 15, 2026 after the internal event lifecycle boundary was added.
- В `notify_service` focused baseline suite выполнен с `-count=2`, затем повторно выполнены `go test -buildvcs=false ./... -count=1`, `go build -buildvcs=false ./...` и `go vet -buildvcs=false ./...` после P0.2a.
- `go vet ./...`
- PostgreSQL-only tests were not executed in the August 15 pass because `FRIENDSHEEP_TEST_POSTGRES_DSN` was not set; skipped tests do not fail `go test ./...`.
- P0.2b focused monolith lifecycle/outbox/handler tests and focused `notify_service` lifecycle/config/bootstrap/HTTP/database/migration tests were run with `-count=2` on August 27, 2026.
- After P0.2b, `go test ./services/events ./handlers ./tests -count=1`, backend `go test ./... -count=1`, backend `go vet ./...`, notify `go test -buildvcs=false ./... -count=1`, notify `go build -buildvcs=false ./...`, and notify `go vet -buildvcs=false ./...` passed.
- PostgreSQL-only P0.2b cursor-lock/`SKIP LOCKED`/lease tests were not executed because neither `FRIENDSHEEP_TEST_POSTGRES_DSN` nor `NOTIFY_SERVICE_TEST_POSTGRES_DSN` was set; the tests explicitly skipped and SQLite was not treated as locking evidence.
- После P0.3 focused backend/notify suites прошли с `-count=2`, затем прошли
  полные `go test ./... -count=1`, `go vet ./...`, а для `notify_service` также
  `go build -buildvcs=false ./...`. PostgreSQL-only reminder/lifecycle/migration
  tests были запущены отдельно и явно пропущены, потому что оба test DSN не были
  заданы; PostgreSQL concurrency и полный cross-process black-box остаются
  release-gate evidence.

Recent completed work:

- P0.2a Notify Service Baseline завершён: legacy scheduler/модели/JWT/FCM удалены, добавлены fail-fast environment config, собственная migration boundary, health/readiness, structured logging, graceful shutdown и пустой notification-channel port без отправок.
- P0.2b Durable Lifecycle Scheduling завершён: event create/reschedule/delete atomically publish ordered schedule outbox records; `notify_service` ingests them through authenticated HTTP into durable cursor/message receipts/jobs and invokes the idempotent lifecycle advance command with lease-based multi-replica claiming, retry, catch-up and graceful shutdown. Notification delivery side effects remain absent.
- P0.3 Event Reminder Intent/Inbox Slice завершён: event create/start-time reschedule/delete atomically publish reminder intent outbox records; `notify_service` ingests them into durable reminder cursors/receipts/jobs, expands canonical `T-24h`/`T-6h`/`T-1h` windows, resolves current recipients on demand, persists authoritative inbox/read state plus delivery attempts, and exposes internal inbox APIs that the monolith proxies behind JWT-authenticated `users/me/notifications` routes.
- Monolith graceful shutdown is complete: SIGINT/SIGTERM stop background admission, drain HTTP requests, cancel request contexts on deadline, and close PostgreSQL, Redis, and MongoDB resources with regression coverage.
- The authenticated internal event lifecycle boundary is implemented at `POST /internal/v1/events/{eventId}/lifecycle/advance`: `EventLifecycleService` owns time-based transitions and catch-up, while a GORM adapter owns the transaction and PostgreSQL row lock.
- Event lifecycle calls use the fail-fast `NOTIFY_SERVICE_TOKEN` configuration and constant-time `X-Internal-Token` middleware; retries and concurrent calls are idempotent, request context is preserved, and no scheduler or notification delivery side effects were added to the monolith.
- The internal lifecycle contract is documented in `docs/internal/event_lifecycle_contract.md`; lifecycle service, handler, GORM, concurrency, and optional PostgreSQL locking tests cover the boundary.
- SQL migration support was added through `golang-migrate` in `db/migrations.go`.
- Draft SQL migrations currently exist under `migration/` for group/event membership uniqueness, pending join-request uniqueness, and group action log action types/target users.
- Migration rollout helper SQL/runbooks exist for membership uniqueness and pending join-request uniqueness.
- Registration bootstrap/schema issue fixed.
- Email verification path hardened for SMTP sender/auth/header issues.
- Auth service decoupled behind narrow auth repository.
- Group-role middleware extracted behind reader seam and tested.
- Event admin routes using `eventId` use event-aware role middleware instead of plain `groupId` middleware.
- Group approve/reject request routes now stay outside the public block and use request-id authorization: `requestId -> groupId -> role`.
- `CreateJoinInvite` fixed: it loads target user, returns `ErrUserNotFound` for missing targets, and logs real target identity.
- `DeleteGroup`, `AddPermissions`, and `RemovePermissions` were moved behind group admin stores instead of doing direct service-level GORM work.
- `LeaveGroup`, `ApproveAllJoinRequests`, and `RejectAllJoinRequests` were moved behind narrower group stores.
- `DeleteUserFromGroup`, `RemoveFromBlacklist`, `UpdateGroup`, `CreateGroup`, and `GetGroupDetails` were moved behind narrower group stores/read stores.
- Group admin/update paths now perform additional capability checks in stores, including defensive actor-role validation for `UpdateGroup`.
- `GetGroupDetails` now performs basic service-level ID validation and maps invalid IDs to a validation error/HTTP 400.
- `GroupFullDto` mapping now lives in `models/dto/convertorsDto` via `ConvertToGroupFullDto`, keeping the read store focused on loading/access rules.
- `CreateGroup` now fails if not all requested categories exist.
- Event create/update/delete paths now use `EventCommandService`, a clean `EventCommandStore`, and the event unit of work.
- Event join/leave paths now use `EventMembershipService` and `EventUnitOfWork`; membership and audit writes remain in one transaction.
- `JoinEvent` rejects events whose start time has been reached, without membership, counter, or audit side effects; the typed error is mapped at the HTTP boundary and documented in Swagger.
- Event search/group-event/details reads now use `EventReadService` and `EventReadStore`.
- Event admin details and kick paths now use `EventAdminService`, `EventAdminReader`, and transactional `EventAdminStore`.
- The legacy broad `EventsService`, `NewEventsService`, and event handler reference dependency were removed; shared event errors now live in `services/events/errors.go`.
- References now use the dedicated `HandlerReferences -> ReferenceService -> ReferenceStore -> GORMReferenceStore` vertical slice with request context propagation.
- `GetAllGenres` and `/api/v2/events/genres` were removed. Genres are available only from public `GET /api/v2/references/genres` with `q`, `page`, and `limit`, case-insensitive substring search, stable `name ASC, id ASC` order, default limit 50, and maximum limit 100.
- `GET /api/v2/references` no longer contains or loads genres; Swagger reflects the separate genre contract.
- Event create/update now treat missing age limit as typed validation error.
- Swagger drift for active group/event DTOs was fixed and docs regenerated.
- Event short/search DTO conversion is now user-aware for `subscribed`; `GetGroupDetails`, `GetGroupEvents`, and `SearchEvents` preload current-user event membership and use the shared converters.
- `EventSearchItemDto` now includes `subscribed`, and Swagger docs were regenerated for that contract.
- The legacy `ServicePopularEvents.go` was removed. Popular-event ranking, cache, notifications, and scheduling now depend on clean application ports, while GORM, Redis, email templates/sending, and cron remain adapter details.
- The popular-events handler depends only on a local reader contract, propagates request context, maps application views to DTOs, and uses the common error DTO.
- Popular-event cache-miss fill is serialized, stale refreshes are asynchronous and coalesced, startup warmup and scheduler shutdown are explicit, returned snapshots are cloned, and notifications are sent only for newly popular events with owner email.
- `UpdateEvent` now rejects an empty update and validates Unicode-aware title/description lengths, positive reference IDs, absolute image URLs with scheme and host, future start time, duration, capacity, genre count/uniqueness, address, and country limits.
- Event type, location, age-limit, and genre existence checks run through the clean event command store before mutation. Missing references and validation failures are typed errors mapped to HTTP 400.
- Update validation runs before mutation/audit/result loading, preserves authorization and already-started-event error precedence, and keeps `maxUsers >= currentUsers`.
- Update-event transport tags and Swagger now match the application contract, including positive IDs, genre item/count/uniqueness constraints, absolute image URL semantics, and future start-time documentation.
- Production role literals were centralized behind constants in `models/groups/groupRoles.go`.
- `JoinGroup` now creates group membership/request inside the transaction.
- `JoinEvent` now uses an atomic `current_users < max_users` counter update.
- Group/event membership tables now have composite uniqueness for duplicate membership prevention.
- Pending private-group join requests now have a Postgres partial unique migration for `(user_id, group_id) WHERE status = 'pending'`.
- Group action log action codes moved to `group_action_types`; API references return `groupActionTypes`.
- Group action logs now store stable actor/target IDs and entity IDs/names where relevant; display names are assembled on read.
- `WatchRecentActions` supports action filtering and default newest-first sorting, with optional oldest-first ordering.
- Group/event audit logging was expanded for join/leave group, invites, request approve/reject/all, role changes, event create/update/delete, event join/leave, and event kick.
- Group audit log helpers were simplified to one `groupActionLogInput` writer for group stores and one `eventGroupActionLogInput` writer for event service logging.

Important test files:

- Shared GORM test adapter/logger: `tests/test_support_test.go`.
- Groups service coverage: `tests/group_service_test.go`.
- Events service coverage: `tests/events_service_test.go`.
- Postgres group service integration coverage: `tests/group_service_postgres_integration_test.go`; it is skipped unless `FRIENDSHEEP_TEST_POSTGRES_DSN` is set.
- Event command service/update validation coverage: `tests/event_command_service_test.go`.
- Event command PostgreSQL locking/savepoint coverage: `tests/event_command_postgres_integration_test.go`; it is skipped unless `FRIENDSHEEP_TEST_POSTGRES_DSN` is set.
- P0.3 cross-process event reminder black-box coverage:
  `tests/p03_cross_service_blackbox_postgres_test.go`; it builds and starts the
  real `notify_service` and is skipped unless both PostgreSQL test DSNs are set.
- Event membership service coverage: `tests/event_membership_service_test.go`.
- Event command/membership HTTP coverage: `handlers/HandlerEventsCommand_test.go` and `handlers/HandlerEventsMembership_test.go`.
- Popular-event service/contract coverage: `tests/popular_events_service_test.go` and `tests/popular_events_contract_test.go`.
- Popular-event GORM/Redis adapter coverage: `tests/popular_events_store_adapter_test.go` and `tests/popular_events_cache_adapter_test.go`.
- Popular-event HTTP coverage: `handlers/HandlerPopularEvents_test.go`.
- Group-role middleware coverage: `tests/group_role_middleware_test.go`.
- Route/middleware regression coverage: `tests/v2_routes_test.go`.
- Seeder/reference coverage: `tests/seeder_test.go`.
- Reference service and contract coverage: `tests/reference_service_test.go` and `tests/reference_contract_test.go`.
- Reference GORM/runtime coverage: `tests/reference_runtime_test.go`.
- Reference HTTP handler coverage: `handlers/HandlerReferences_test.go`.
- Migration coverage: `db/migrations_test.go`.
- Session store coverage: `sessions/sessionStore_test.go`.

SQLite test rules:

- Use shared in-memory SQLite only when service code mixes transaction repo and root repo.
- Make DSNs unique per test invocation to support `-count=2`.
- Enable `PRAGMA foreign_keys = ON`.
- Seed reference rows explicitly or idempotently when FK-dependent models are inserted.

## Definitively Closed In The Current Code

The following slices are complete for the current architecture and should not be reopened without a new product requirement:

- Joining an already-started event: the service rejects it before any write, preserves existing membership-error precedence, and has service, GORM, handler, and Swagger coverage.
- Popular-events boundary extraction: application/handler code is independent of GORM, Redis, cron, and email library types; adapters and focused tests now own those details.
- Popular-events refresh lifecycle: cache-hit/miss/stale behavior, refresh coalescing, scheduled warmup, shutdown cancellation, immutable returned snapshots, and newly-popular notification selection are covered.
- Event-update validation: empty updates, field boundaries, reference existence, capacity consistency, error mapping, rollback/no-side-effect behavior, and Swagger contract are covered.
- Internal event lifecycle advance: authenticated routing, time-boundary decisions, direct completion catch-up, transactional row locking, idempotent retries, and stable HTTP outcomes are covered. PostgreSQL-specific locking evidence still requires the optional integration environment.

These closures do not claim that PostgreSQL-specific concurrency tests have run in the current environment or that the final production SQL baseline exists; those are separate release-verification items below.

## Priority Vulnerabilities And Risks

### P0 Next Product Dependency: Notification Statistics And Additional Delivery Channels

What:

- Durable event lifecycle scheduling, reminder intents, recipient resolution, authoritative inbox/read state and the first `in_app` reminder delivery slice are complete, but statistics orchestration and additional external delivery channels do not exist yet.
- Future news notifications and richer user-notification UX must reuse the new authoritative notification identity, persistence, delivery-attempt and retry model in `notify_service` rather than duplicating state in the monolith.
- The lifecycle schedule outbox and reminder intent outbox are deliberately separate; future notification domains must not overload either contract with unrelated payloads or recreate fire-and-forget delivery inside monolith handlers.

Why P0:

- Product-visible reminders and a reusable inbox now exist, so the next smallest block is statistics. A new delivery channel requires a separate product requirement.
- This is still required by the user area and later news/comment interactions.
- The new reminder/inbox backbone should now be reused rather than bypassed.

Required next work:

- Define versioned statistics trigger/result contracts and keep them separate from lifecycle scheduling and reminder delivery contracts.
- Add another delivery adapter only when explicitly required, behind the clean `DeliveryChannel` port and without changing reminder scheduling, inbox identity or the monolith JWT/public API boundary.
- Reuse the reliable monolith outbox/consumer handoff for future notification intents; do not add fire-and-forget network calls inside monolith transactions.
- Extend cross-service contract tests and disposable-infrastructure integration coverage, including downtime recovery and PostgreSQL-only locking evidence.

### Release Gate: Clean PostgreSQL Baseline / First Production Bootstrap

What:

- There is no existing database or production data to migrate, so duplicate cleanup and old-data rollout are not active risks.
- The current SQL migration chain is not a full baseline: non-DEV startup runs SQL migrations before the DEV-only GORM bootstrap, while migrations `000002` and later assume core tables already exist.
- This work is intentionally deferred while the schema is changing substantially.

Why it is a release gate:

- It does not block ongoing feature refactoring with disposable DEV/test databases.
- It will block the first non-DEV deployment because a completely empty PostgreSQL database is not yet guaranteed to bootstrap successfully.

Required work at schema freeze:

- Inventory the final models, foreign keys, indexes, partial indexes, and required reference data.
- Squash or replace the unapplied draft history with a clean initial-schema baseline and idempotent required seed migrations.
- Verify migration from version zero on empty PostgreSQL, verify the resulting schema contract, start the server, and verify a second startup produces no changes.
- Make the clean-bootstrap and migration-contract checks required in CI/release verification.

### P2 During Development / Release Gate: PostgreSQL-Specific Concurrency Verification

What:

- A draft Postgres partial unique migration defines the required `(user_id, group_id) WHERE status = 'pending'` constraint; it has not been applied because no persistent database exists.
- PostgreSQL integration tests cover concurrent duplicate join requests, event-command row-lock/savepoint behavior, lifecycle cursor locking and multi-worker `SKIP LOCKED`/lease recovery. They are skipped unless `FRIENDSHEEP_TEST_POSTGRES_DSN` or `NOTIFY_SERVICE_TEST_POSTGRES_DSN` is provided for the owning service.
- SQLite/unit coverage cannot prove PostgreSQL row locks, partial unique indexes, context cancellation while waiting for a lock, savepoint recovery, cursor serialization or `SKIP LOCKED` claiming.

Why it remains open:

- The implementation and tests exist, but the PostgreSQL-only evidence was not produced in the current environment.
- No production database is needed; these tests should run against a disposable PostgreSQL instance before the first release.

Required next work:

- Run the existing PostgreSQL integration tests in CI or a release verification job with `FRIENDSHEEP_TEST_POSTGRES_DSN` and `NOTIFY_SERVICE_TEST_POSTGRES_DSN`.
- Keep the unique-violation mapping to stable user-facing errors covered when the join/request path changes.
- Keep event command locking and best-effort audit savepoint behavior covered when transaction adapters change.

### P1: Repository Port Coupling

What:

- `repository/postgresRepository.go` still exposes GORM methods/types directly to services.
- Core event command/membership/read/admin/popular paths and the reference slice are migrated to narrow ports with library details isolated in adapters.
- Some group paths still use GORM-shaped local store contracts.
- Popular-event GORM/Redis adapter constructors still accept the broad legacy repository facades, but these types no longer cross the application or handler boundary.
- Group action log lookup still uses a GORM-shaped lookup interface for action type IDs.

Why P1:

- This keeps architecture brittle and makes future storage/library changes expensive.
- It also makes testing harder and encourages service code to mix persistence details with business logic.

Required next work:

- Continue extracting narrow ports feature-by-feature after current tests are stable.
- Prioritize remaining GORM-shaped group contracts and audit action-type lookup.
- Keep `services/references` and the core event services storage-agnostic; do not reintroduce the removed broad `EventsService` or genres in the general references response.
- Do not collapse the completed popular-event ports back into a repository-driven service.
- Keep avoiding a full shared repository rewrite.

### P2: Capability Model / Role Authorization

What:

- Role names are centralized as constants, and capability checks exist, but some service and logging flows still pass normalized role strings.
- If roles or permissions grow, string-role checks will spread again.

Why P2:

- Current behavior is safer than before, but long-term authorization logic can become inconsistent.
- It is less urgent than data integrity because current route/service tests cover the active role paths.

Required next work:

- Introduce role capabilities or permission predicates when authorization rules grow.
- Keep route middleware and service authorization using the same capability source.
- Avoid adding new raw role-string checks in services.

### P2: Group Audit Log Product Semantics

What:

- Audit logs now store action type IDs and stable user/entity IDs, and descriptions are generated on read.
- The action type table is still group-focused even though it includes event-related group actions.
- Logging coverage is broader, but it is not yet a full cross-module audit subsystem.

Why P2:

- Admin-facing logs are now more reliable, but action taxonomy and filtering will keep growing.
- Without a clearer audit module boundary, event/group/business actions can become tangled again.

Required next work:

- Decide whether audit logging remains group-scoped or becomes a general audit module.
- Add tests for action filtering/order and for display name generation after user profile changes.
- Keep new action types seeded through `DefaultGroupActionTypes()` and SQL migrations.

### P2: Mixed Error Response Shapes

What:

- Middleware and handlers can return different JSON shapes for forbidden/not-found/validation errors.
- Swagger now matches active DTOs better, but common error DTO policy is still incomplete.

Why P2:

- This affects API consistency and clients, not core data safety.
- It is still worth fixing before frontend/API clients rely heavily on response contracts.

Required next work:

- Define common error DTOs for auth/middleware/domain errors.
- Update Swagger comments and route tests around response shape.

## Recommended Next Chat Start

Recommended product order:

1. Add statistics trigger/result orchestration on top of the completed reminder/inbox backbone in `notify_service`, keeping contracts separate from lifecycle scheduling and reminder delivery.
2. Restore the user domain in smaller slices: public/own profile and privacy contract; search and profile editing; account deletion; notification preferences and profile statistics selection/display. Reuse the notification backbone instead of embedding delivery in user handlers.
3. Rebuild news/comments after identity, privacy, and notifications are stable: read model first, then authoring/editing permissions, comments/interactions, moderation, and notification intents.

The next implementation chat should stay focused on statistics orchestration or one user-domain slice, not on rebuilding reminder scheduling/inbox from scratch. Do not restore the removed general legacy scheduler, JWT/device/Telegram/FCM implementations, the full user rebuild, or news in the same slice.

Architecture-debt tasks such as remaining GORM-shaped group ports, audit-module cleanup, and common error DTOs remain valid, but they should be handled opportunistically inside the selected vertical slice or in separate maintenance branches rather than displacing the P0 product dependency.

Do not prioritize the final SQL baseline until schema freeze unless the next objective becomes the first staging/non-DEV deployment. PostgreSQL integration verification can be done earlier against a disposable database.

Do not combine all of these in one branch unless explicitly requested.
