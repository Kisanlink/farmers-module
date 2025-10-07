-- 002_crop_master_data.sql
-- Migration to add crop master data tables and enhance crop cycles

-- Create crop category enum
CREATE TYPE crop_category AS ENUM (
  'CEREALS',
  'PULSES',
  'VEGETABLES',
  'FRUITS',
  'OIL_SEEDS',
  'SPICES',
  'CASH_CROPS',
  'FODDER',
  'MEDICINAL',
  'OTHER'
);

-- Crops master table
CREATE TABLE crops (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  scientific_name VARCHAR(255),
  category crop_category NOT NULL,
  duration_days INTEGER, -- typical growing duration in days
  seasons JSONB NOT NULL DEFAULT '[]', -- ["RABI", "KHARIF", "ZAID"]
  unit VARCHAR(50) NOT NULL DEFAULT 'kg', -- measurement unit for yield
  properties JSONB NOT NULL DEFAULT '{}', -- additional metadata like water requirements, soil type preferences
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index for faster searches
CREATE INDEX idx_crops_category ON crops(category);
CREATE INDEX idx_crops_name ON crops(name);
CREATE INDEX idx_crops_active ON crops(is_active);

-- Crop varieties table
CREATE TABLE crop_varieties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  crop_id UUID NOT NULL REFERENCES crops(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  duration_days INTEGER, -- variety-specific duration, overrides crop default
  yield_potential_kg_per_ha INTEGER, -- expected yield per hectare
  properties JSONB NOT NULL DEFAULT '{}', -- variety-specific properties
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(crop_id, name)
);

-- Create index for faster variety lookups
CREATE INDEX idx_crop_varieties_crop_id ON crop_varieties(crop_id);
CREATE INDEX idx_crop_varieties_active ON crop_varieties(is_active);

-- Add new columns to crop_cycles table for structured crop data
ALTER TABLE crop_cycles
ADD COLUMN crop_id UUID REFERENCES crops(id),
ADD COLUMN variety_id UUID REFERENCES crop_varieties(id),
ADD COLUMN farmer_id UUID; -- This was missing from the original schema

-- Create indexes for the new foreign keys
CREATE INDEX idx_crop_cycles_crop_id ON crop_cycles(crop_id);
CREATE INDEX idx_crop_cycles_variety_id ON crop_cycles(variety_id);
CREATE INDEX idx_crop_cycles_farmer_id ON crop_cycles(farmer_id);

-- Add constraint to ensure either crop_id or planned_crops is provided
ALTER TABLE crop_cycles
ADD CONSTRAINT check_crop_data CHECK (
  crop_id IS NOT NULL OR
  (planned_crops IS NOT NULL AND jsonb_array_length(planned_crops) > 0)
);

-- Add comments for documentation
COMMENT ON TABLE crops IS 'Master data for crops with categories, duration, units, and seasons';
COMMENT ON TABLE crop_varieties IS 'Varieties of crops with specific characteristics and yield potential';
COMMENT ON COLUMN crop_cycles.crop_id IS 'Reference to the crop master data (new structured approach)';
COMMENT ON COLUMN crop_cycles.variety_id IS 'Reference to specific crop variety';
COMMENT ON COLUMN crop_cycles.planned_crops IS 'Legacy field for backward compatibility - stores crop names as strings';