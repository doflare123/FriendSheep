-- WARNING:
-- This migration is logically one-way for data.
-- The matching .up migration removes duplicate membership rows and recalculates counters.
-- Running .down only removes schema artifacts (indexes/audit objects) and does NOT restore deleted data.
DROP INDEX IF EXISTS public.idx_event_user_membership;
DROP INDEX IF EXISTS public.idx_group_user_membership;
DROP TABLE IF EXISTS public.membership_dedupe_audit;
DROP SEQUENCE IF EXISTS public.membership_dedupe_audit_id_seq;
