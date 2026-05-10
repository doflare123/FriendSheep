# Backend Agent Context

## Project

This repository is the Go backend for `friendSheep`, a social/group/event service.

Main stack:

- Go + Gin HTTP API.
- GORM-backed Postgres repository facade in `repository/postgresRepository.go`.
- Redis sessions/cache in `repository/redisRepository.go` and `sessions/`.
- Swagger docs generated under `docs/`.
- Tests live mainly in `tests/`, with some package-local tests in `db/` and `sessions/`.

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

Known architecture debt:

- `repository/postgresRepository.go` leaks GORM API/types across services.
- Redis and Mongo repository layers likely also leak library-specific behavior.
- Role names now have constants in `models/groups/groupRoles.go`, but authorization is still mostly role-string based rather than capability based.
- Groups/events services are still large and storage-coupled; repo-port extraction should be vertical and test-backed.

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
- For git commands in sandbox, use:
  `git -c safe.directory=D:/friendSheep ...`

## Current State

Recent green checks:

- `go test ./tests -run '^(TestRegisterGroupsRoutesRequestID|TestGroupService|TestEventsService)' -count=2`
- `go test ./... -count=1`

Recent completed work:

- Registration bootstrap/schema issue fixed.
- Email verification path hardened for SMTP sender/auth/header issues.
- Auth service decoupled behind narrow auth repository.
- Group-role middleware extracted behind reader seam and tested.
- Event admin routes using `eventId` use event-aware role middleware instead of plain `groupId` middleware.
- Group approve/reject request routes now stay outside the public block and use request-id authorization: `requestId -> groupId -> role`.
- `CreateJoinInvite` fixed: it loads target user, returns `ErrUserNotFound` for missing targets, and logs real target identity.
- `CreateGroup` now fails if not all requested categories exist.
- Event references now include genres.
- Event create/update now treat missing age limit as typed validation error.
- Swagger drift for active group/event DTOs was fixed and docs regenerated.
- Production role literals were centralized behind constants in `models/groups/groupRoles.go`.
- `JoinGroup` now creates group membership/request inside the transaction.
- `JoinEvent` now uses an atomic `current_users < max_users` counter update.
- Group/event membership tables now have composite uniqueness for duplicate membership prevention.

Important test files:

- Shared GORM test adapter/logger: `tests/test_support_test.go`.
- Groups service coverage: `tests/group_service_test.go`.
- Events service coverage: `tests/events_service_test.go`.
- Group-role middleware coverage: `tests/group_role_middleware_test.go`.
- Route/middleware regression coverage: `tests/v2_routes_test.go`.
- Session store coverage: `sessions/sessionStore_test.go`.

SQLite test rules:

- Use shared in-memory SQLite only when service code mixes transaction repo and root repo.
- Make DSNs unique per test invocation to support `-count=2`.
- Enable `PRAGMA foreign_keys = ON`.
- Seed reference rows explicitly or idempotently when FK-dependent models are inserted.

## Priority Vulnerabilities And Risks

### P0: Migration / Data Integrity Rollout

What:

- New composite unique indexes for group and event membership protect against duplicate memberships, but production rollout can fail if duplicate rows already exist.
- The project still has no mature SQL migration layer; relying on GORM auto-migration for constraints is unsafe for production data.

Why P0:

- This can break startup/deploy or silently leave constraints unapplied.
- It affects data integrity and all group/event participation behavior.

Required next work:

- Add real SQL migration support.
- Add pre-migration cleanup/query for duplicate `group_users(user_id, group_id)` and `events_users(event_id, user_id)`.
- Verify on Postgres, not only SQLite.

### P1: Pending Join Request Race

What:

- `JoinGroup` checks duplicate pending requests inside a transaction, but there is no DB-level partial unique constraint for only `status = 'pending'`.
- A normal GORM composite unique on `(user_id, group_id, status)` was intentionally not added because it would also block valid repeat requests after reject/approve for the same status.

Why P1:

- Concurrent requests can still create duplicate pending join requests.
- This is a real integrity issue, but narrower than P0 because it affects only private group request flow.

Required next work:

- Add a Postgres partial unique index, for example on `(user_id, group_id) WHERE status = 'pending'`.
- Add a Postgres integration test for concurrent duplicate request attempts.

### P1: Repository Port Coupling

What:

- `repository/postgresRepository.go` still exposes GORM methods/types directly to services.
- Groups/events services still depend on broad repository behavior instead of narrow feature ports.

Why P1:

- This keeps architecture brittle and makes future storage/library changes expensive.
- It also makes testing harder and encourages service code to mix persistence details with business logic.

Required next work:

- Extract narrow ports feature-by-feature after current tests are stable.
- Start with a small group/event slice, not a full repository rewrite.

### P2: Capability Model / Role Authorization

What:

- Role names are centralized as constants, but permission checks still use role names directly.
- If roles or permissions grow, string-role checks will spread again.

Why P2:

- Current behavior is safer than before, but long-term authorization logic can become inconsistent.
- It is less urgent than data integrity because current route/service tests cover the active role paths.

Required next work:

- Introduce role capabilities or permission predicates when authorization rules grow.
- Keep route middleware and service authorization using the same capability source.

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

### P3: Popular Events / Cache Paths

What:

- Popular/cache scheduler paths remain thinly tested.
- Redis/cache seams are not yet extracted enough for clean unit tests.

Why P3:

- Lower direct registration/group/event correctness risk right now.
- Should wait until cache/notifier/scheduler ports are extracted.

Required next work:

- Extract cache/popular-event behavior ports.
- Add tests after seams exist.

## Recommended Next Chat Start

Start with one focused task:

- Postgres SQL migrations for uniqueness constraints and production rollout cleanup, or
- common middleware error DTO/documentation cleanup, or
- first narrow repo-port extraction for group/event services.

Do not combine all of these in one branch unless explicitly requested.
