-- Migration: Fix area_ha_computed column to use geography cast instead of geometry
-- This ensures accurate area calculation in hectares for geographic coordinates (SRID 4326)
--
-- Issue: The original formula used geometry::geometry which calculates area in square degrees
-- Fix: Using geometry::geography which calculates actual area in square meters
--
-- Date: 2025-10-14
-- Related: Farm entity area_ha field

-- Drop the existing generated column if it exists
ALTER TABLE farms DROP COLUMN IF EXISTS area_ha_computed;

-- Re-create the column with the correct formula using geography cast
-- ST_Area(geometry::geography) returns area in square meters
-- Dividing by 10000 converts square meters to hectares
ALTER TABLE farms ADD COLUMN area_ha_computed NUMERIC(12,4)
    GENERATED ALWAYS AS (ST_Area(geometry::geography)/10000.0) STORED;

-- Add comment to document the column
COMMENT ON COLUMN farms.area_ha_computed IS 'Automatically calculated farm area in hectares using PostGIS geography calculations. Generated from geometry field.';
