-- Migration: Add FPO Lifecycle Management Fields
-- Purpose: Extend fpo_refs table with lifecycle management and audit tracking
-- Date: 2025-11-18

BEGIN;

-- ========================================
-- Step 1: Add lifecycle fields to fpo_refs
-- ========================================

-- Lifecycle state tracking
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS previous_status VARCHAR(50);
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS status_reason TEXT;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS status_changed_at TIMESTAMP;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS status_changed_by VARCHAR(255);

-- Verification tracking
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS verification_status VARCHAR(50);
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS verified_by VARCHAR(255);
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS verification_notes TEXT;

-- Setup retry tracking
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS setup_attempts INT DEFAULT 0;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS last_setup_at TIMESTAMP;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS setup_progress JSONB DEFAULT '{}';

-- Relationships
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS ceo_user_id VARCHAR(255);
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS parent_fpo_id VARCHAR(255);

-- Metadata field if not exists
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- ========================================
-- Step 2: Create FPO Audit Logs Table
-- ========================================

CREATE TABLE IF NOT EXISTS fpo_audit_logs (
    id VARCHAR(255) PRIMARY KEY,
    fpo_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    previous_state VARCHAR(50),
    new_state VARCHAR(50),
    reason TEXT,
    performed_by VARCHAR(255) NOT NULL,
    performed_at TIMESTAMP NOT NULL,
    details JSONB,
    request_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_fpo_audit_logs_fpo_id FOREIGN KEY (fpo_id) REFERENCES fpo_refs(id) ON DELETE CASCADE
);

-- ========================================
-- Step 3: Add Indexes
-- ========================================

-- Performance indexes for fpo_refs
CREATE INDEX IF NOT EXISTS idx_fpo_refs_status ON fpo_refs(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fpo_refs_aaa_org_id ON fpo_refs(aaa_org_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fpo_refs_registration_no ON fpo_refs(registration_number) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fpo_refs_ceo_user_id ON fpo_refs(ceo_user_id) WHERE deleted_at IS NULL;

-- Audit log indexes
CREATE INDEX IF NOT EXISTS idx_fpo_audit_logs_fpo_id ON fpo_audit_logs(fpo_id);
CREATE INDEX IF NOT EXISTS idx_fpo_audit_logs_performed_at ON fpo_audit_logs(performed_at DESC);
CREATE INDEX IF NOT EXISTS idx_fpo_audit_logs_action ON fpo_audit_logs(action);

-- ========================================
-- Step 4: Update existing records
-- ========================================

-- Mark existing ACTIVE FPOs with no previous status
UPDATE fpo_refs
SET status = 'ACTIVE',
    previous_status = NULL
WHERE status = 'ACTIVE' AND previous_status IS NULL;

-- Convert PENDING_SETUP to SETUP_FAILED for retry capability
UPDATE fpo_refs
SET status = 'SETUP_FAILED',
    previous_status = 'PENDING_SETUP',
    status_reason = 'Legacy pending setup - requires retry via CompleteFPOSetup or new lifecycle retry endpoint'
WHERE status = 'PENDING_SETUP';

-- ========================================
-- Step 5: Add comments for documentation
-- ========================================

COMMENT ON COLUMN fpo_refs.status IS 'Current FPO lifecycle status: DRAFT, PENDING_VERIFICATION, VERIFIED, REJECTED, PENDING_SETUP, SETUP_FAILED, ACTIVE, SUSPENDED, INACTIVE, ARCHIVED';
COMMENT ON COLUMN fpo_refs.previous_status IS 'Previous status before last transition';
COMMENT ON COLUMN fpo_refs.status_reason IS 'Reason for current status';
COMMENT ON COLUMN fpo_refs.status_changed_at IS 'Timestamp of last status change';
COMMENT ON COLUMN fpo_refs.status_changed_by IS 'User ID who performed the status change';
COMMENT ON COLUMN fpo_refs.setup_attempts IS 'Number of setup retry attempts (max 3)';
COMMENT ON COLUMN fpo_refs.ceo_user_id IS 'AAA user ID of FPO CEO';

COMMENT ON TABLE fpo_audit_logs IS 'Audit trail for FPO lifecycle state transitions';

COMMIT;
