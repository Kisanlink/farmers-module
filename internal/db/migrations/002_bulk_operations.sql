-- 002_bulk_operations.sql
-- Migration to add bulk operations tables

-- Bulk operations table
CREATE TABLE bulk_operations (
  id VARCHAR(255) PRIMARY KEY,
  fpo_org_id VARCHAR(255) NOT NULL,
  initiated_by VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
  input_format VARCHAR(20) NOT NULL,
  processing_mode VARCHAR(20) NOT NULL,
  total_records INTEGER NOT NULL DEFAULT 0,
  processed_records INTEGER NOT NULL DEFAULT 0,
  successful_records INTEGER NOT NULL DEFAULT 0,
  failed_records INTEGER NOT NULL DEFAULT 0,
  skipped_records INTEGER NOT NULL DEFAULT 0,
  start_time TIMESTAMPTZ,
  end_time TIMESTAMPTZ,
  processing_time BIGINT, -- in milliseconds
  result_file_url TEXT,
  error_summary JSONB DEFAULT '{}',
  options JSONB DEFAULT '{}',
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Processing details table
CREATE TABLE processing_details (
  id VARCHAR(255) PRIMARY KEY,
  operation_id VARCHAR(255) NOT NULL REFERENCES bulk_operations(id) ON DELETE CASCADE,
  record_index INTEGER NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
  farmer_id VARCHAR(255),
  aaa_user_id VARCHAR(255),
  error_message TEXT,
  error_code VARCHAR(100),
  processing_time BIGINT, -- in milliseconds
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (operation_id, record_index)
);

-- Create indexes for performance
CREATE INDEX idx_bulk_operations_fpo_org_id ON bulk_operations (fpo_org_id);
CREATE INDEX idx_bulk_operations_status ON bulk_operations (status);
CREATE INDEX idx_bulk_operations_initiated_by ON bulk_operations (initiated_by);
CREATE INDEX idx_processing_details_operation_id ON processing_details (operation_id);
CREATE INDEX idx_processing_details_status ON processing_details (status);

-- Add comments for documentation
COMMENT ON TABLE bulk_operations IS 'Bulk farmer addition operations with progress tracking';
COMMENT ON TABLE processing_details IS 'Individual record processing details within bulk operations';
COMMENT ON COLUMN bulk_operations.processing_time IS 'Total processing time in milliseconds';
COMMENT ON COLUMN processing_details.processing_time IS 'Individual record processing time in milliseconds';
