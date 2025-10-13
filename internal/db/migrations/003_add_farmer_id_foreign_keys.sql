-- 003_add_farmer_id_foreign_keys.sql
-- Migration to add farmer_id foreign key relationships to user-related tables

-- Add farmer_id to farms table
-- This establishes a direct link from farm to the farmer who owns it
ALTER TABLE farms
ADD COLUMN IF NOT EXISTS farmer_id VARCHAR(255);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_farms_farmer_id ON farms(farmer_id);

-- Add foreign key constraint to link farms to farmers
-- Using ON DELETE CASCADE to ensure data consistency when a farmer is deleted
ALTER TABLE farms
ADD CONSTRAINT fk_farms_farmer_id
FOREIGN KEY (farmer_id)
REFERENCES farmers(id)
ON DELETE CASCADE;

-- Update crop_cycles.farmer_id column type and add foreign key constraint
-- First, alter the column type from UUID to VARCHAR(255) for consistency
ALTER TABLE crop_cycles
ALTER COLUMN farmer_id TYPE VARCHAR(255);

-- Make farmer_id NOT NULL as every crop cycle must belong to a farmer
ALTER TABLE crop_cycles
ALTER COLUMN farmer_id SET NOT NULL;

-- Add foreign key constraint to link crop cycles to farmers
ALTER TABLE crop_cycles
ADD CONSTRAINT fk_crop_cycles_farmer_id
FOREIGN KEY (farmer_id)
REFERENCES farmers(id)
ON DELETE CASCADE;

-- Add farmer_id to farm_activities table
-- This allows direct queries for activities by farmer without joining through crop_cycle
ALTER TABLE farm_activities
ADD COLUMN IF NOT EXISTS farmer_id VARCHAR(255);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_farm_activities_farmer_id ON farm_activities(farmer_id);

-- Make farmer_id NOT NULL as every activity must belong to a farmer
ALTER TABLE farm_activities
ALTER COLUMN farmer_id SET NOT NULL;

-- Add foreign key constraint to link farm activities to farmers
ALTER TABLE farm_activities
ADD CONSTRAINT fk_farm_activities_farmer_id
FOREIGN KEY (farmer_id)
REFERENCES farmers(id)
ON DELETE CASCADE;

-- Add comments for documentation
COMMENT ON COLUMN farms.farmer_id IS 'Foreign key reference to the farmer who owns this farm';
COMMENT ON COLUMN crop_cycles.farmer_id IS 'Foreign key reference to the farmer managing this crop cycle';
COMMENT ON COLUMN farm_activities.farmer_id IS 'Foreign key reference to the farmer who performed/planned this activity';

-- Update existing farm records to set farmer_id based on aaa_user_id
-- This is a data migration to populate the new farmer_id column
-- You may need to adjust this based on your actual data mapping logic
UPDATE farms f
SET farmer_id = fr.id
FROM farmers fr
WHERE f.aaa_user_id = fr.aaa_user_id
  AND f.aaa_org_id = fr.aaa_org_id
  AND f.farmer_id IS NULL;

-- Update existing crop_cycles records to ensure farmer_id is populated
-- If farmer_id is already populated but needs to be linked properly
UPDATE crop_cycles cc
SET farmer_id = fr.id
FROM farms f
JOIN farmers fr ON f.farmer_id = fr.id
WHERE cc.farm_id = f.id
  AND cc.farmer_id IS NULL;

-- Update existing farm_activities records to set farmer_id from crop_cycle
UPDATE farm_activities fa
SET farmer_id = cc.farmer_id
FROM crop_cycles cc
WHERE fa.crop_cycle_id = cc.id
  AND fa.farmer_id IS NULL;
