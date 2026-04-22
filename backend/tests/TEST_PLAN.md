# Backend Test Plan

Goal: all active v2 backend behavior covered in `tests/`. Old commented/disabled handlers are out of scope until migrated.

## Current Coverage

- Auth handler:
  - login success
  - login service error
  - refresh success
  - invalid login payload
- Auth routes:
  - `POST /api/v2/auth/refresh` is active
  - `GET /api/v2/auth/refresh` is not active
- Register routes:
  - session create route
  - session verify route
  - user create route
  - password change route
- Sub routes:
  - upload requires Bearer access token
  - upload reaches handler with valid token
- Events routes:
  - public genres, references, popular endpoints
  - protected event details requires auth
- Groups routes:
  - representative public/operator/admin groups routes require auth
- Middleware:
  - auth token format, invalid token, refresh-token rejection
- JWT utils:
  - access/refresh token round trip
  - wrong token type rejection
  - empty secret rejection
- Services/register:
  - verification success
  - too many attempts deletes session
  - user creation creates auth tokens, user, default tiles, stats rows, deletes session
- Infrastructure:
  - seeder fills current reference tables

## Next Coverage

- Services/register:
  - session creation writes Redis fields
  - verification handles bad code before final attempt and type mismatch
  - password change requires verified reset session
- Services/events:
  - create event access check, genre validation, creator auto-join
  - update event cannot mutate started event
  - join/leave event counters and duplicate protection
  - admin detail/kick/delete permission paths
- Services/groups:
  - create group creates admin membership and contacts
  - private group access rules
  - join open/private group behavior
  - approve/reject requests and invites
  - role changes and blacklist flows
- Middleware:
  - group-role path/query/body groupId extraction
  - forbidden role response
- Infrastructure:
  - config loads env defaults/errors
  - S3 image validation and upload error mapping
