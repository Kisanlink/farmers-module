# Farmers Module Product Overview

## Product Summary

The Farmers Module is a Go-based microservice that manages agricultural operations within the KisanLink ecosystem. It handles farmer profiles, farm management, crop cycles, and farm activities with integrated AAA (Authentication, Authorization, Audit) enforcement.

## Core Domain

- **Farmers**: Individual agricultural practitioners linked to FPO organizations
- **FPOs**: Farmer Producer Organizations that group and manage farmers
- **Farms**: Geographic agricultural land parcels with PostGIS spatial data
- **Crop Cycles**: Seasonal agricultural cycles (Rabi, Kharif, Zaid)
- **Farm Activities**: Individual tasks within crop cycles (planting, harvesting, etc.)

## Key Features

- Workflow-based architecture with 19 defined workflows (W1-W19)
- AAA service integration for delegated authentication and authorization
- PostGIS spatial operations for farm boundary management
- Multi-protocol support (HTTP REST + gRPC)
- Structured audit logging and error handling

## Business Value

- Streamlines farm management operations
- Ensures secure access control through AAA integration
- Provides spatial analytics for agricultural planning
- Enables scalable farmer onboarding and management
