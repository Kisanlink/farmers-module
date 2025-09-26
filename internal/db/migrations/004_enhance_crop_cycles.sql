-- 004_enhance_crop_cycles.sql
-- Migration to enhance crop_cycles table with additional fields

-- Add new columns to crop_cycles table
ALTER TABLE crop_cycles 
ADD COLUMN crop_id VARCHAR(255),
ADD COLUMN variety_id VARCHAR(255),
ADD COLUMN crop_type VARCHAR(20) NOT NULL DEFAULT 'ANNUAL' CHECK (crop_type IN ('ANNUAL', 'PERENNIAL')),
ADD COLUMN crop_name VARCHAR(100),
ADD COLUMN variety_name VARCHAR(100),
ADD COLUMN acreage DECIMAL(12,4),
ADD COLUMN number_of_trees INTEGER,
ADD COLUMN sowing_transplanting_date DATE,
ADD COLUMN harvest_date DATE,
ADD COLUMN yield_per_acre DECIMAL(12,4),
ADD COLUMN yield_per_tree DECIMAL(12,4),
ADD COLUMN total_yield DECIMAL(12,4),
ADD COLUMN yield_unit VARCHAR(20),
ADD COLUMN tree_age_range_min INTEGER,
ADD COLUMN tree_age_range_max INTEGER,
ADD COLUMN image_url VARCHAR(500),
ADD COLUMN document_id VARCHAR(255),
ADD COLUMN report_data JSONB DEFAULT '{}',
ADD COLUMN metadata JSONB DEFAULT '{}';

-- Rename start_date to sowing_transplanting_date if it exists and is different
-- Note: This is handled in the application layer by using both fields

-- Add foreign key constraints
ALTER TABLE crop_cycles 
ADD CONSTRAINT fk_crop_cycles_crop_id 
FOREIGN KEY (crop_id) REFERENCES crops(id) ON DELETE SET NULL;

ALTER TABLE crop_cycles 
ADD CONSTRAINT fk_crop_cycles_variety_id 
FOREIGN KEY (variety_id) REFERENCES crop_varieties(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX idx_crop_cycles_crop_id ON crop_cycles (crop_id);
CREATE INDEX idx_crop_cycles_variety_id ON crop_cycles (variety_id);
CREATE INDEX idx_crop_cycles_crop_type ON crop_cycles (crop_type);
CREATE INDEX idx_crop_cycles_harvest_date ON crop_cycles (harvest_date);
CREATE INDEX idx_crop_cycles_sowing_date ON crop_cycles (sowing_transplanting_date);

-- Add comments for documentation
COMMENT ON COLUMN crop_cycles.crop_id IS 'Reference to the crop master data';
COMMENT ON COLUMN crop_cycles.variety_id IS 'Reference to the specific crop variety';
COMMENT ON COLUMN crop_cycles.crop_type IS 'Type of crop: ANNUAL or PERENNIAL';
COMMENT ON COLUMN crop_cycles.crop_name IS 'Name of the crop (denormalized for performance)';
COMMENT ON COLUMN crop_cycles.variety_name IS 'Name of the variety (denormalized for performance)';
COMMENT ON COLUMN crop_cycles.acreage IS 'Area under cultivation in acres';
COMMENT ON COLUMN crop_cycles.number_of_trees IS 'Number of trees for perennial crops';
COMMENT ON COLUMN crop_cycles.sowing_transplanting_date IS 'Date of sowing or transplanting';
COMMENT ON COLUMN crop_cycles.harvest_date IS 'Date of harvest';
COMMENT ON COLUMN crop_cycles.yield_per_acre IS 'Yield per acre';
COMMENT ON COLUMN crop_cycles.yield_per_tree IS 'Yield per tree (for perennial crops)';
COMMENT ON COLUMN crop_cycles.total_yield IS 'Total yield for the cycle';
COMMENT ON COLUMN crop_cycles.yield_unit IS 'Unit of measurement for yield';
COMMENT ON COLUMN crop_cycles.tree_age_range_min IS 'Minimum age of trees in years';
COMMENT ON COLUMN crop_cycles.tree_age_range_max IS 'Maximum age of trees in years';
COMMENT ON COLUMN crop_cycles.image_url IS 'URL of crop cycle image';
COMMENT ON COLUMN crop_cycles.document_id IS 'Reference to uploaded document';
COMMENT ON COLUMN crop_cycles.report_data IS 'Additional report data in JSON format';
COMMENT ON COLUMN crop_cycles.metadata IS 'Additional metadata in JSON format';
