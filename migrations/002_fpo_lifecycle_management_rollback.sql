-- Rollback: Remove FPO Lifecycle Management Fields
-- Purpose: Revert changes from 002_fpo_lifecycle_management.sql
-- Date: 2025-11-18

BEGIN;

-- ========================================
-- Step 1: Restore original status values
-- ========================================

-- Restore PENDING_SETUP status for FPOs marked as SETUP_FAILED
UPDATE fpo_refs
SET status = 'PENDING_SETUP',
    status_reason = NULL
WHERE status = 'SETUP_FAILED';

-- ========================================
-- Step 2: Drop audit logs table
-- ========================================

DROP TABLE IF EXISTS fpo_audit_logs;

-- ========================================
-- Step 3: Drop indexes from fpo_refs
-- ========================================

DROP INDEX IF EXISTS idx_fpo_refs_ceo_user_id;
DROP INDEX IF EXISTS idx_fpo_refs_registration_no;
DROP INDEX IF EXISTS idx_fpo_refs_aaa_org_id;
DROP INDEX IF EXISTS idx_fpo_refs_status;

-- ========================================
-- Step 4: Remove added columns from fpo_refs
-- ========================================

ALTER TABLE fpo_refs DROP COLUMN IF EXISTS parent_fpo_id;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS ceo_user_id;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS setup_progress;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS last_setup_at;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS setup_attempts;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS verification_notes;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS verified_by;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS verified_at;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS verification_status;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS status_changed_by;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS status_changed_at;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS status_reason;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS previous_status;

COMMIT;
