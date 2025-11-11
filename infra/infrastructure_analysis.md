# Farmers Module Infrastructure Analysis
## AWS ECS Fargate Production Deployment

**Generated:** 2025-11-10
**Service:** farmers-module
**Version:** 1.0.0
**Analysis Scope:** Production-ready AWS infrastructure with ECS Fargate, RDS PostgreSQL with PostGIS, and CodePipeline CI/CD

---

## 1. Application Analysis

### 1.1 Runtime & Framework
- **Language:** Go 1.24.4 (latest stable)
- **Framework:** Gin web framework
- **Build System:** Standard Go build with multi-stage Docker
- **Binary Output:** Single statically-linked executable (`/app/farmers-server`)
- **Startup Command:** `/app/farmers-server` (no arguments required)

### 1.2 Service Type
- **Primary Protocol:** HTTP REST API on port 8000
- **Documentation:** Swagger/OpenAPI available at `/docs`
- **Health Check:** `GET /health` returns JSON status
- **Root Endpoint:** `GET /` returns service metadata

### 1.3 Dependencies & External Services

#### Critical External Dependencies:
1. **PostgreSQL Database (REQUIRED)**
   - Version: PostgreSQL 16+
   - Extension: **PostGIS** (spatial operations for farm boundaries)
   - SSL Mode: Required in production
   - Max Connections: 50 recommended for production

2. **AAA Service gRPC (REQUIRED)**
   - Authentication, Authorization, and Auditing microservice
   - Protocol: gRPC over port 50052
   - Endpoint: Configurable via `AAA_GRPC_ADDR`
   - Startup Behavior: Role seeding on startup (30s timeout, non-fatal if fails)
   - Impact: Service starts even if AAA unavailable, but auth will fail

#### No Cache/Queue Dependencies:
- No Redis requirement (confirmed)
- No message queue (SQS, RabbitMQ, etc.)
- No S3 direct access from application

### 1.4 Configuration Requirements

**Required Environment Variables:**
```bash
# Database (PostgreSQL with PostGIS)
DB_POSTGRES_HOST=<rds-endpoint>
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=farmers_service
DB_POSTGRES_PASSWORD=<from-secrets-manager>
DB_POSTGRES_DBNAME=farmers_production
DB_POSTGRES_SSLMODE=require
DB_POSTGRES_MAX_CONNS=50

# AAA Service Integration
AAA_GRPC_ADDR=<aaa-service-endpoint>:50052
AAA_API_KEY=<from-secrets-manager>
AAA_ENABLED=true
AAA_RETRY_ATTEMPTS=3
AAA_RETRY_BACKOFF=100ms
AAA_REQUEST_TIMEOUT=5s

# Server Configuration
SERVICE_PORT=8000
SERVICE_NAME=farmers-module
ENVIRONMENT=production
HOST=0.0.0.0

# JWT Configuration (if local validation needed)
JWT_SECRET=<from-secrets-manager>
SECRET_KEY=<from-secrets-manager>

# Observability
LOG_LEVEL=warn
ENABLE_TRACING=true
ENABLE_METRICS=true
OTEL_EXPORTER_OTLP_ENDPOINT=<optional-collector>:4317
```

**Secrets to Store in AWS Secrets Manager:**
1. `DB_POSTGRES_PASSWORD` - RDS master password
2. `AAA_API_KEY` - AAA service authentication key
3. `JWT_SECRET` - JWT signing key
4. `SECRET_KEY` - Application secret key

---

## 2. Database Migration Analysis

### 2.1 Migration Strategy
**Type:** Auto-migration via GORM on application startup
**Location:** `internal/db/db.go::SetupDatabase()`
**Execution:** Runs BEFORE server starts listening

### 2.2 Migration Complexity
- **Total Migration Code:** ~582 lines
- **Migration Types:**
  - Schema creation (AutoMigrate for 15+ entities)
  - PostGIS extension enablement
  - Custom ENUM creation (6 enums: season, crop_category, cycle_status, etc.)
  - Spatial indexes (GIST indexes for geometry columns)
  - Computed columns (area_ha_computed using ST_Area)
  - Database triggers (farmer acreage rollup)
  - Multi-table foreign key constraints

### 2.3 Migration Duration Estimate
**Startup Time Analysis:**
- Empty database (first deployment): **15-25 seconds**
  - PostGIS extension: 2-3s
  - ENUM creation: 1-2s
  - Table creation (15+ tables): 5-8s
  - Indexes creation: 3-5s
  - Trigger setup: 2-3s
  - ID counter initialization: 2-4s

- Existing database (rolling update): **3-5 seconds**
  - Schema comparison and validation
  - Index verification

- AAA role seeding: **2-5 seconds** (30s timeout, non-fatal)

**Total Cold Start:** 20-30 seconds
**Total Warm Start (existing DB):** 8-10 seconds

### 2.4 Migration Failure Handling
- Fatal: Database connection failure
- Fatal: Migration errors (schema conflicts, permission issues)
- Non-fatal: AAA role seeding failure (logs warning, continues)
- Fallback: Graceful degradation without PostGIS (spatial features disabled)

---

## 3. Health Check Configuration

### 3.1 Application Health Endpoints

**Primary Health Check:**
- **Path:** `/health`
- **Method:** GET
- **Expected Response:** 200 OK with JSON `{"status":"ok","service":"farmers-module"}`
- **No Authentication Required**

**Additional Monitoring Endpoints:**
- `/` - Root endpoint (200 OK, service metadata)
- `/api/v1/admin/health` - Admin health check (may require auth)

### 3.2 Health Check Timing Requirements

**Container Startup Sequence:**
1. Container starts: 0s
2. Database migration begins: +1s
3. Migration completes: +20-30s (cold) or +5-10s (warm)
4. AAA role seeding: +2-5s (non-blocking)
5. Gin server starts listening: +1s
6. **READY TO ACCEPT TRAFFIC:** 25-35s (cold) or 10-15s (warm)

**Recommended Health Check Configuration:**

**ALB Target Group Health Check:**
- **Health Check Path:** `/health`
- **Health Check Protocol:** HTTP
- **Success Codes:** 200
- **Interval:** 30 seconds
- **Timeout:** 5 seconds
- **Healthy Threshold:** 2 consecutive successes
- **Unhealthy Threshold:** 3 consecutive failures
- **Initial Delay:** 45 seconds (accounts for migration time)

**Container Health Check:**
- **DISABLE container-level health checks** (rely on ALB only)
- **Reason:** Migration runs on startup; container health check would fail during migration

### 3.3 Graceful Shutdown
- Listens for SIGINT/SIGTERM
- Closes database connections cleanly
- No explicit timeout (instant shutdown after connection close)

---

## 4. Resource Sizing Analysis

### 4.1 Memory Requirements

**Base Application:**
- Go runtime: ~20-30 MB
- Gin framework: ~10-15 MB
- GORM ORM + connection pool (50 conns): ~30-40 MB
- gRPC client connections: ~5-10 MB
- Application code + data structures: ~20-30 MB

**Peak Usage Scenarios:**
- Bulk farmer import (1000 records): +50-100 MB
- Concurrent API requests (50 req/s): +30-50 MB
- PostGIS spatial operations: +20-40 MB

**Recommended Memory:**
- **Minimum:** 512 MB (tight, for staging)
- **Production:** **1024 MB (1 GB)** ✅
- **Justification:** Allows headroom for bulk operations and concurrent requests

### 4.2 CPU Requirements

**Workload Characteristics:**
- REST API (I/O bound): Light CPU
- Database queries (ORM): Moderate CPU
- Spatial operations (PostGIS): Offloaded to database
- JSON serialization: Light-moderate CPU
- gRPC calls: Light CPU

**Recommended CPU:**
- **Minimum:** 0.25 vCPU (256 CPU units)
- **Production:** **0.5 vCPU (512 CPU units)** ✅
- **Justification:** Go is efficient; 0.5 vCPU handles 100+ req/s for typical CRUD

### 4.3 Storage Requirements
- Container ephemeral storage: 20 GB (ECS default, sufficient)
- No persistent volume needed (stateless service)
- Logs: CloudWatch Logs (no local disk usage)

---

## 5. Network & Security Configuration

### 5.1 Port Exposure
- **Container Port:** 8000 (HTTP)
- **ALB Listener:** 443 (HTTPS) → Target Group Port 8000
- **Internal Communication:** AAA service on port 50052 (gRPC)

### 5.2 Security Group Rules

**ECS Task Security Group (Ingress):**
- Port 8000 from ALB Security Group only
- No other ingress (deny all)

**ECS Task Security Group (Egress):**
- Port 5432 to RDS Security Group (PostgreSQL)
- Port 50052 to AAA Service (gRPC) - via Service Discovery or VPC endpoint
- Port 443 to Internet Gateway (for AAA service if public, or VPC endpoints)
- Port 443 to AWS services (Secrets Manager, CloudWatch via VPC endpoints)

**ALB Security Group (Ingress):**
- Port 443 from 0.0.0.0/0 (public HTTPS)
- Port 80 from 0.0.0.0/0 (redirect to 443)

**RDS Security Group (Ingress):**
- Port 5432 from ECS Task Security Group only

### 5.3 IAM Permissions

**ECS Task Role (Application Permissions):**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:region:account:secret:farmers-module/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:region:account:log-group:/ecs/farmers-module:*"
    }
  ]
}
```

**ECS Task Execution Role (Container Launch Permissions):**
- AmazonECSTaskExecutionRolePolicy (managed policy)
- Secrets Manager access for environment injection

---

## 6. Database Configuration (RDS PostgreSQL with PostGIS)

### 6.1 RDS Instance Specifications

**Engine:**
- PostgreSQL 16.x
- **PostGIS Extension:** REQUIRED (must be enabled via `CREATE EXTENSION IF NOT EXISTS postgis`)

**Instance Class:**
- **Staging:** db.t4g.small (2 vCPU, 2 GB RAM)
- **Production:** db.t4g.medium (2 vCPU, 4 GB RAM) or db.r6g.large (2 vCPU, 16 GB RAM) for high load

**Storage:**
- Type: gp3 (General Purpose SSD)
- Allocated: 100 GB (auto-scaling enabled up to 500 GB)
- IOPS: 3000 (baseline for gp3)

**Multi-AZ:** Enabled for production (automatic failover)

### 6.2 Database Configuration Parameters

**Custom Parameter Group:**
```
max_connections = 100
shared_buffers = 1GB (for db.t4g.medium with 4GB RAM)
effective_cache_size = 3GB
maintenance_work_mem = 256MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 10MB
min_wal_size = 1GB
max_wal_size = 4GB
```

**PostGIS-Specific:**
- Install `postgis` extension in RDS via Parameter Group or manual SQL
- SRID 4326 (WGS 84) for spatial data

### 6.3 Connection Pooling
- Application pool: 50 connections (`DB_POSTGRES_MAX_CONNS=50`)
- RDS max_connections: 100 (allows 2 task instances or admin connections)

### 6.4 Backup & Recovery
- Automated backups: Enabled (7-35 day retention)
- Backup window: 03:00-04:00 UTC
- Maintenance window: Sunday 04:00-05:00 UTC
- Point-in-time recovery: Enabled

---

## 7. Autoscaling Strategy

### 7.1 ECS Service Autoscaling

**Target Tracking Policies:**

1. **CPU Utilization:**
   - Target: 60%
   - Scale-out: Add 1 task when CPU > 60% for 3 min
   - Scale-in: Remove 1 task when CPU < 50% for 15 min

2. **Memory Utilization:**
   - Target: 70%
   - Scale-out: Add 1 task when Memory > 70% for 3 min
   - Scale-in: Remove 1 task when Memory < 60% for 15 min

3. **ALB Request Count per Target:**
   - Target: 1000 requests/target/minute
   - Scale-out: Add task if request count > 1000/min
   - Scale-in: Remove task if request count < 500/min for 15 min

**Task Limits:**
- **Minimum Tasks:** 2 (high availability)
- **Maximum Tasks:** 10 (cost control, adjust based on load)
- **Desired Tasks:** 2 (initial state)

### 7.2 Database Autoscaling
- RDS storage autoscaling: Enabled (100 GB → 500 GB max)
- No CPU-based RDS autoscaling (t4g instances don't support Read Replicas easily)
- Consider Read Replica if read-heavy workload emerges

---

## 8. Deployment Strategy

### 8.1 CodeDeploy Deployment Type
- **Type:** Blue/Green (ECS)
- **Traffic Shift:** Linear 10% every 1 minute (10 minutes total)
- **Rollback:** Automatic on CloudWatch alarms (5xx errors, unhealthy targets)

### 8.2 Deployment Safety
- **Pre-traffic Hook:** None (health checks sufficient)
- **Post-traffic Hook:** None (validation via monitoring)
- **Rollback Triggers:**
  - ALB 5xx error rate > 5%
  - Unhealthy target count > 0 for 5 minutes
  - Custom CloudWatch metrics (optional)

### 8.3 Zero-Downtime Requirements
- Minimum 2 tasks always running
- Blue/Green deployment ensures old tasks run until new tasks healthy
- Database migrations are idempotent (GORM AutoMigrate + `IF NOT EXISTS`)

---

## 9. Observability & Monitoring

### 9.1 CloudWatch Logs
- **Log Group:** `/ecs/farmers-module`
- **Retention:** 7 days (staging), 30 days (production)
- **Structured Logging:** Yes (Uber Zap with JSON format)

### 9.2 CloudWatch Metrics

**Application Metrics:**
- HTTP request count
- HTTP error rates (4xx, 5xx)
- Response time (p50, p95, p99)
- Database connection pool usage

**Infrastructure Metrics:**
- ECS CPU/Memory utilization
- ALB target health
- RDS CPU, memory, storage, IOPS
- RDS connection count

### 9.3 CloudWatch Alarms

**Critical Alarms (Immediate Action):**
1. ALB Unhealthy Target Count > 0 for 5 min
2. RDS CPU > 80% for 10 min
3. RDS Storage < 10 GB free
4. ECS Service Desired ≠ Running for 10 min
5. ALB 5xx error rate > 5% for 5 min

**Warning Alarms (Investigation):**
1. ALB 4xx error rate > 10% for 10 min
2. RDS Connection count > 80 (80% of max_connections)
3. ECS Memory utilization > 80%
4. ALB Target Response Time > 1s (p99)

### 9.4 OpenTelemetry Integration
- Application supports OTEL (optional)
- Export to AWS X-Ray or third-party collector
- Configure via `OTEL_EXPORTER_OTLP_ENDPOINT`

---

## 10. Cost Estimation

### 10.1 Monthly Cost Breakdown (Production, us-east-1)

**ECS Fargate (2 tasks, 0.5 vCPU, 1 GB each):**
- vCPU: 2 tasks × 0.5 vCPU × $0.04048/vCPU/hour × 730 hours = $29.55
- Memory: 2 tasks × 1 GB × $0.004445/GB/hour × 730 hours = $6.49
- **Subtotal:** $36.04/month

**RDS PostgreSQL (db.t4g.medium, Multi-AZ):**
- Instance: $0.144/hour × 730 hours × 2 AZ = $210.24
- Storage: 100 GB × $0.115/GB = $11.50
- Backup: 100 GB × $0.095/GB = $9.50 (if > 100 GB total backups)
- **Subtotal:** $231.24/month

**Application Load Balancer:**
- ALB hour: $0.0225/hour × 730 hours = $16.43
- LCU usage: ~5 LCU × $0.008/LCU × 730 hours = $29.20
- **Subtotal:** $45.63/month

**Data Transfer:**
- Inter-AZ (ECS ↔ RDS): ~10 GB × $0.01/GB = $0.10
- Internet out: ~50 GB × $0.09/GB = $4.50
- **Subtotal:** $4.60/month

**CloudWatch:**
- Logs: 10 GB ingested × $0.50/GB = $5.00
- Metrics: 50 custom metrics × $0.30/metric = $15.00
- Alarms: 10 alarms × $0.10/alarm = $1.00
- **Subtotal:** $21.00/month

**Secrets Manager:**
- 4 secrets × $0.40/secret = $1.60
- API calls: ~10,000 × $0.05/10k calls = $0.05
- **Subtotal:** $1.65/month

**TOTAL ESTIMATED MONTHLY COST:** **~$340/month**

### 10.2 Cost Optimization Recommendations
1. Use Reserved Instances for RDS (save ~40%)
2. Use Savings Plan for Fargate (save ~20%)
3. Right-size instances based on actual usage
4. Enable S3 for long-term log archival (cheaper than CloudWatch)

---

## 11. Critical Deployment Considerations

### 11.1 Health Check Configuration (CRITICAL)
- **DO NOT** enable container-level health checks in ECS task definition
- **REASON:** Migrations run on startup; container would be killed before healthy
- **SOLUTION:** Rely solely on ALB target group health checks with 45s delay

### 11.2 ALB Target Group Success Codes
- **Current Implementation:** `/health` returns 200 for unauthenticated requests
- **Configuration:** `200` only (no 401 needed)
- **Note:** If authenticated endpoints used for health, add `200,401`

### 11.3 Database Migration Idempotency
- GORM AutoMigrate is idempotent (safe for concurrent deployments)
- PostGIS and ENUM creation use `IF NOT EXISTS`
- Triggers use `CREATE OR REPLACE`
- **Safe for rolling deployments:** ✅

### 11.4 AAA Service Dependency
- Application starts even if AAA service unavailable
- Role seeding has 30s timeout (non-fatal)
- **Impact:** Authentication will fail if AAA down, but service won't crash
- **Recommendation:** Ensure AAA service highly available or use circuit breaker

### 11.5 Secret Management
- Never hardcode secrets in task definition
- Use `valueFrom` with Secrets Manager ARNs
- Rotate secrets regularly (30-90 days)

### 11.6 Network Architecture
- **Private Subnets:** ECS tasks in private subnets (no public IP)
- **Public Subnets:** ALB in public subnets
- **NAT Gateway:** Required for ECS tasks to reach internet (AWS APIs, AAA service if public)
- **VPC Endpoints:** Recommended for Secrets Manager, CloudWatch (save NAT costs)

---

## 12. Deployment Checklist

### Pre-Deployment:
- [ ] RDS PostgreSQL 16 instance created with PostGIS support
- [ ] VPC with public/private subnets and NAT Gateway configured
- [ ] Security groups created (ALB, ECS, RDS)
- [ ] Secrets stored in AWS Secrets Manager
- [ ] ECR repository created for Docker images
- [ ] IAM roles created (ECS Task Role, Task Execution Role)
- [ ] CloudWatch Log Group created
- [ ] SSL/TLS certificate imported to ACM (for ALB HTTPS)

### Deployment:
- [ ] CloudFormation stack deployed
- [ ] ECS cluster and service created
- [ ] ALB target group health checks passing
- [ ] Database migrations completed successfully
- [ ] AAA service connectivity verified
- [ ] CloudWatch alarms configured

### Post-Deployment Validation:
- [ ] Health check endpoint responding: `curl https://<alb-dns>/health`
- [ ] API documentation accessible: `https://<alb-dns>/docs`
- [ ] Database tables created (verify via psql)
- [ ] PostGIS extension enabled: `SELECT PostGIS_version();`
- [ ] CloudWatch logs streaming
- [ ] Autoscaling policies active

---

## 13. Appendix: Migration SQL Analysis

### Tables Created (15 total):
1. addresses
2. farmers (with total_acreage_ha computed column)
3. farmer_links
4. fpo_refs
5. farms (with PostGIS geometry, area_ha_computed)
6. farm_soil_types (junction)
7. farm_irrigation_sources (junction)
8. soil_types
9. irrigation_sources
10. crops
11. crop_varieties
12. stages
13. crop_stages
14. crop_cycles
15. farm_activities
16. bulk_operations
17. bulk_processing_details

### Indexes Created (~30 total):
- GIST spatial index on farms.geometry
- Unique indexes on farmer identifiers
- Foreign key indexes for joins
- Status/category filter indexes

### Custom ENUMs (6 total):
- season, crop_category, cycle_status, activity_status, link_status, farmer_status

### Database Triggers (2 total):
- Farmer acreage rollup on farm insert/update/delete

---

**End of Infrastructure Analysis**
