-- Initialize PostGIS extension for Farmers Module
-- This script runs automatically when the PostgreSQL container is first created

\echo 'Creating PostGIS extension...'

-- Enable PostGIS extension (includes geometry, geography, and raster support)
CREATE EXTENSION IF NOT EXISTS postgis;

-- Enable PostGIS topology support (optional, but useful for advanced spatial operations)
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- Enable PostGIS raster support (optional, for raster data)
-- Uncomment if needed:
-- CREATE EXTENSION IF NOT EXISTS postgis_raster;

-- Verify PostGIS installation
SELECT PostGIS_Version();

\echo 'PostGIS extension created successfully!'
