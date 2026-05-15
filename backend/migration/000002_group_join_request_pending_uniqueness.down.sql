-- WARNING:
-- This migration drops only the partial unique index for pending join requests.
-- It does NOT recreate or merge duplicate application rows.
DROP INDEX IF EXISTS public.idx_group_join_request_pending_unique;
