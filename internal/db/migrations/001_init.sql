-- 001_init.sql
-- Initial migration for farmers module

-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create custom types
CREATE TYPE season AS ENUM ('RABI', 'KHARIF', 'ZAID', 'OTHER');
CREATE TYPE cycle_status AS ENUM ('PLANNED', 'ACTIVE', 'COMPLETED', 'CANCELLED');

-- FPO references table (local cache/config only)
CREATE TABLE fpo_refs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  aaa_org_id TEXT UNIQUE NOT NULL,
  business_config JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Farmer links to FPOs
CREATE TABLE farmer_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  aaa_user_id TEXT NOT NULL,
  aaa_org_id TEXT NOT NULL,
  kisan_sathi_user_id TEXT,
  status TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (aaa_user_id, aaa_org_id)
);

-- Farms table with PostGIS geometry
CREATE TABLE farms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  aaa_farmer_user_id TEXT NOT NULL,
  aaa_org_id TEXT NOT NULL,
  geom geometry(Polygon, 4326) NOT NULL,
  area_ha NUMERIC(12,4) GENERATED ALWAYS AS (ST_Area(geom::geography)/10000.0) STORED,
  metadata JSONB NOT NULL DEFAULT '{}',
  created_by TEXT NOT NULL, -- aaa user id
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Crop cycles table
CREATE TABLE crop_cycles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  farm_id UUID NOT NULL REFERENCES farms(id) ON DELETE CASCADE,
  season season NOT NULL,
  status cycle_status NOT NULL DEFAULT 'PLANNED',
  start_date DATE NOT NULL,
  end_date DATE,
  planned_crops JSONB NOT NULL DEFAULT '[]',
  outcome JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Farm activities table
CREATE TABLE farm_activities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cycle_id UUID NOT NULL REFERENCES crop_cycles(id) ON DELETE CASCADE,
  activity_type TEXT NOT NULL,
  planned_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  metadata JSONB NOT NULL DEFAULT '{}',
  output JSONB,
  created_by TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create indexes for performance
CREATE INDEX idx_farms_geom ON farms USING GIST (geom);
CREATE INDEX idx_farms_farmer_org ON farms (aaa_farmer_user_id, aaa_org_id);
CREATE INDEX idx_crop_cycles_farm_id ON crop_cycles (farm_id);
CREATE INDEX idx_farm_activities_cycle_id ON farm_activities (cycle_id);
CREATE INDEX idx_farmer_links_org ON farmer_links (aaa_org_id);
CREATE INDEX idx_farmer_links_kisan_sathi ON farmer_links (kisan_sathi_user_id);

-- Add comments for documentation
COMMENT ON TABLE fpo_refs IS 'Local cache of FPO organization references from AAA service';
COMMENT ON TABLE farmer_links IS 'Links between farmers (AAA users) and FPOs (AAA organizations)';
COMMENT ON TABLE farms IS 'Farm boundaries and metadata, linked to AAA users and organizations';
COMMENT ON TABLE crop_cycles IS 'Agricultural cycles within farms, including seasons and outcomes';
COMMENT ON TABLE farm_activities IS 'Individual activities within crop cycles';
COMMENT ON COLUMN farms.area_ha IS 'Farm area in hectares, computed from PostGIS geometry';
COMMENT ON COLUMN farms.geom IS 'Farm boundary as PostGIS Polygon in WGS84 (SRID 4326)';
