-- 003_crop_master_data.sql
-- Migration to add crop master data tables

-- Crops master table
CREATE TABLE crops (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  category VARCHAR(50) NOT NULL CHECK (category IN ('CEREAL', 'LEGUME', 'VEGETABLE', 'OIL_SEEDS', 'FRUIT', 'SPICE')),
  crop_duration_days INTEGER,
  typical_units JSONB DEFAULT '[]',
  seasons JSONB DEFAULT '[]',
  image_url VARCHAR(500),
  document_id VARCHAR(255),
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Crop varieties table
CREATE TABLE crop_varieties (
  id VARCHAR(255) PRIMARY KEY,
  crop_id VARCHAR(255) NOT NULL REFERENCES crops(id) ON DELETE CASCADE,
  variety_name VARCHAR(100) NOT NULL,
  duration_days INTEGER,
  characteristics TEXT,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (crop_id, variety_name)
);

-- Crop stages table
CREATE TABLE crop_stages (
  id VARCHAR(255) PRIMARY KEY,
  crop_id VARCHAR(255) NOT NULL REFERENCES crops(id) ON DELETE CASCADE,
  stage_name VARCHAR(100) NOT NULL,
  stage_order INTEGER NOT NULL,
  typical_duration_days INTEGER,
  description TEXT,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (crop_id, stage_order)
);

-- Create indexes for performance
CREATE INDEX idx_crops_category ON crops (category);
CREATE INDEX idx_crops_name ON crops (name);
CREATE INDEX idx_crop_varieties_crop_id ON crop_varieties (crop_id);
CREATE INDEX idx_crop_varieties_name ON crop_varieties (variety_name);
CREATE INDEX idx_crop_stages_crop_id ON crop_stages (crop_id);
CREATE INDEX idx_crop_stages_order ON crop_stages (crop_id, stage_order);

-- Add comments for documentation
COMMENT ON TABLE crops IS 'Master data for crops with categories, duration, units, and seasons';
COMMENT ON TABLE crop_varieties IS 'Varieties of specific crops with characteristics and duration';
COMMENT ON TABLE crop_stages IS 'Growth stages for crops with order and typical duration';
COMMENT ON COLUMN crops.typical_units IS 'Array of typical units for this crop (KG, QUINTAL, TONNES, PIECES)';
COMMENT ON COLUMN crops.seasons IS 'Array of seasons when this crop is grown (KHARIF, RABI, SUMMER, PERENNIAL)';
COMMENT ON COLUMN crop_varieties.characteristics IS 'Description of variety characteristics and features';
COMMENT ON COLUMN crop_stages.stage_order IS 'Order of the stage in the crop growth cycle (0-based)';
COMMENT ON COLUMN crop_stages.typical_duration_days IS 'Typical duration of this stage in days';
