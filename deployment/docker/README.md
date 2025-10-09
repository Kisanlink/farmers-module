# Farmers Module - Docker Deployment Guide

This directory contains Docker configuration files for the Farmers Module microservice, enabling consistent deployment across development, staging, and production environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [File Structure](#file-structure)
- [Environment Configuration](#environment-configuration)
- [Development Workflow](#development-workflow)
- [Staging Deployment](#staging-deployment)
- [Production Deployment](#production-deployment)
- [Makefile Commands](#makefile-commands)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)

## Quick Start

### Prerequisites

- Docker Engine 20.10+ or Docker Desktop
- Docker Compose 2.0+
- Git (for version metadata)
- Make (optional, for convenience commands)

### Initial Setup

1. **Copy environment file:**
   ```bash
   cp deployment/docker/.env.example deployment/docker/.env
   ```

2. **Edit environment variables:**
   ```bash
   # Edit .env file with your configuration
   nano deployment/docker/.env
   ```

3. **Start development environment:**
   ```bash
   make docker-dev
   # or
   cd deployment/docker && docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
   ```

4. **Access services:**
   - API: http://localhost:8000
   - API Documentation: http://localhost:8000/docs
   - pgAdmin: http://localhost:5050 (admin@farmers.local / admin)
   - PostgreSQL: localhost:5432

5. **Stop environment:**
   ```bash
   make docker-dev-down
   # or
   cd deployment/docker && docker-compose -f docker-compose.yml -f docker-compose.dev.yml down
   ```

## Architecture

### Multi-Stage Docker Build

The Dockerfile uses a multi-stage build approach:

1. **Builder Stage (golang:1.24.4-alpine)**
   - Compiles Go application
   - Downloads dependencies
   - Generates Swagger documentation
   - Produces static binary with build metadata

2. **Runtime Stage (alpine:3.20)**
   - Minimal Alpine Linux base
   - Non-root user (appuser, UID 65534)
   - Only the compiled binary
   - Health check support
   - Final image size: ~40-50MB

### Service Stack

```
┌─────────────────────────────────────────┐
│         farmers-service                  │
│   (Go Application Container)             │
│   Ports: 8000 (HTTP), 8081 (gRPC)       │
└───────────────┬─────────────────────────┘
                │
                │ DB Connection
                ▼
┌─────────────────────────────────────────┐
│          postgres                        │
│   (PostgreSQL 16 + PostGIS)              │
│   Port: 5432 (internal only)             │
└─────────────────────────────────────────┘
```

## File Structure

```
deployment/docker/
├── Dockerfile                    # Multi-stage build configuration
├── docker-compose.yml            # Base service definitions
├── docker-compose.dev.yml        # Development overrides
├── docker-compose.staging.yml    # Staging overrides
├── docker-compose.prod.yml       # Production overrides
├── .env.example                  # Environment variables template
├── init-scripts/                 # Database initialization scripts
│   └── 01-init-postgis.sql      # PostGIS extension setup
└── README.md                     # This file
```

## Environment Configuration

### Environment Files

Create a `.env` file in `deployment/docker/` with your configuration:

```bash
# Service Configuration
SERVICE_NAME=farmers-module
SERVICE_PORT=8000
ENVIRONMENT=development

# Database Configuration
DB_POSTGRES_HOST=postgres
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=postgres
DB_POSTGRES_DBNAME=farmers_module
DB_POSTGRES_SSLMODE=disable

# AAA Service (optional)
AAA_ENABLED=false

# Observability
LOG_LEVEL=info
ENABLE_TRACING=false
ENABLE_METRICS=false
```

### Environment-Specific Configurations

#### Development (docker-compose.dev.yml)
- Debug logging enabled
- PostgreSQL exposed on host port 5432
- pgAdmin for database management
- Hot-reload support (future enhancement)
- Debugging tools enabled

#### Staging (docker-compose.staging.yml)
- Production-like configuration
- Resource limits enforced
- PostgreSQL tuned for performance
- Structured logging
- Metrics and tracing enabled

#### Production (docker-compose.prod.yml)
- Strict security settings
- Read-only filesystem
- Maximum resource limits
- External database recommended
- No exposed ports (except service endpoints)

## Development Workflow

### Starting Development Environment

```bash
# Using Makefile (recommended)
make docker-dev

# Or using docker-compose directly
cd deployment/docker
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

### Building from Scratch

```bash
# Rebuild and start
make docker-dev-build

# Or
cd deployment/docker
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

### Viewing Logs

```bash
# All services
make docker-logs

# Farmers service only
make docker-logs-app

# PostgreSQL only
make docker-logs-db
```

### Accessing Containers

```bash
# Shell in farmers-service
make docker-shell

# PostgreSQL shell
make docker-shell-db

# Or using docker-compose
cd deployment/docker
docker-compose exec farmers-service /bin/sh
docker-compose exec postgres psql -U postgres -d farmers_module
```

### Running Tests

```bash
# Run tests in Docker
make docker-test

# Or manually
docker run --rm -v $(PWD):/app -w /app golang:1.24.4-alpine sh -c \
  "apk add --no-cache git make && go test ./... -v"
```

### Database Management

Access pgAdmin at http://localhost:5050:
- Email: admin@farmers.local
- Password: admin

**Add PostgreSQL Server:**
1. Right-click "Servers" → "Register" → "Server"
2. General Tab: Name = "Farmers DB"
3. Connection Tab:
   - Host: postgres
   - Port: 5432
   - Database: farmers_module
   - Username: postgres
   - Password: postgres

## Staging Deployment

### Prerequisites

1. Configure environment variables for staging:
   ```bash
   cp deployment/docker/.env.example deployment/docker/.env.staging
   # Edit .env.staging with staging values
   ```

2. Set Git commit hash:
   ```bash
   export GIT_COMMIT=$(git rev-parse --short HEAD)
   ```

### Deploy to Staging

```bash
# Using Makefile
make docker-staging

# Or using docker-compose
cd deployment/docker
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d
```

### Monitor Staging

```bash
# Check service health
make docker-health

# View logs
cd deployment/docker
docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs -f
```

## Production Deployment

**IMPORTANT:** The production docker-compose configuration is for reference only. In production, use orchestration platforms like:
- Kubernetes with Helm charts
- AWS ECS/Fargate
- Google Cloud Run
- Azure Container Instances

### Kubernetes Deployment (Recommended)

For production, migrate to Kubernetes:

```yaml
# Example Kubernetes Deployment (reference)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: farmers-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: farmers-service
        image: ghcr.io/kisanlink/farmers-module:1.0.0
        ports:
        - containerPort: 8000
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
```

### Managed Database

Use managed PostgreSQL services:
- **AWS**: RDS for PostgreSQL with PostGIS
- **GCP**: Cloud SQL for PostgreSQL
- **Azure**: Azure Database for PostgreSQL
- **DigitalOcean**: Managed Databases

## Makefile Commands

### Build Commands

```bash
make docker-build          # Build Docker image with metadata
```

### Environment Commands

```bash
make docker-dev            # Start development environment
make docker-dev-build      # Build and start development
make docker-dev-down       # Stop development environment
make docker-staging        # Start staging environment
make docker-prod           # Start production environment (reference)
```

### Management Commands

```bash
make docker-up             # Start services (base config)
make docker-down           # Stop and remove containers
make docker-down-volumes   # Stop and remove volumes (deletes data)
make docker-restart        # Restart all services
make docker-restart-app    # Restart farmers-service only
```

### Monitoring Commands

```bash
make docker-logs           # View all logs
make docker-logs-app       # View farmers-service logs
make docker-logs-db        # View PostgreSQL logs
make docker-ps             # List running containers
make docker-health         # Check health status
```

### Container Access

```bash
make docker-shell          # Shell in farmers-service
make docker-shell-db       # psql shell in PostgreSQL
```

### Testing Commands

```bash
make docker-test           # Run tests in Docker
```

### Cleanup Commands

```bash
make docker-clean          # Remove unused Docker resources
make docker-clean-all      # Remove ALL Docker resources (nuclear)
```

### Help Commands

```bash
make help                  # Show all available commands
make docker-quickstart     # Show quick start guide
```

## Troubleshooting

### Container Won't Start

**Check logs:**
```bash
make docker-logs-app
```

**Common issues:**
- Database not ready: Wait for PostgreSQL health check
- Port already in use: Change ports in .env file
- Missing environment variables: Verify .env file

### Database Connection Errors

**Verify PostgreSQL is healthy:**
```bash
cd deployment/docker
docker-compose ps
```

**Check PostGIS extension:**
```bash
make docker-shell-db
# In psql:
SELECT PostGIS_Version();
```

### Permission Denied Errors

**On macOS/Linux:**
```bash
# Ensure Docker has access to current directory
# Check Docker Desktop settings → Resources → File Sharing
```

### Image Size Too Large

**Build without cache:**
```bash
cd deployment/docker
docker-compose build --no-cache --pull
```

**Verify final image size:**
```bash
docker images farmers-module:latest
# Should be ~40-50MB for runtime stage
```

### Health Check Failing

**Check health endpoint manually:**
```bash
docker-compose exec farmers-service curl -f http://localhost:8000/health
```

**Increase health check timeout:**
Edit `docker-compose.yml`:
```yaml
healthcheck:
  start_period: 60s  # Increase if app takes longer to start
```

## Security Best Practices

### Development

- ✅ Use `.env` files, never commit secrets
- ✅ Use default credentials for local development
- ✅ Keep development environment isolated

### Staging

- ✅ Use strong passwords
- ✅ Enable SSL for database connections
- ✅ Restrict network access
- ✅ Enable AAA service
- ✅ Use secrets management

### Production

- ✅ Use managed database services
- ✅ Enable SSL/TLS everywhere
- ✅ Use secrets managers (AWS Secrets Manager, Vault)
- ✅ Implement network policies
- ✅ Enable read-only filesystem
- ✅ Run as non-root user
- ✅ Regular security updates
- ✅ Container image scanning
- ✅ Audit logging enabled

### Docker Image Security

```bash
# Scan image for vulnerabilities
docker scan farmers-module:latest

# Use Trivy for detailed scanning
trivy image farmers-module:latest
```

## Performance Tuning

### PostgreSQL Configuration

For production workloads, tune PostgreSQL settings in `docker-compose.prod.yml`:

- `shared_buffers`: 25% of RAM
- `effective_cache_size`: 50-75% of RAM
- `max_connections`: Based on workload
- `work_mem`: RAM / max_connections / 2

### Application Configuration

- Set `DB_POSTGRES_MAX_CONNS` based on expected load
- Enable connection pooling
- Configure timeouts appropriately
- Use caching (Redis) for frequently accessed data

## Next Steps

1. **Set up CI/CD pipeline** to automate builds and deployments
2. **Implement container scanning** in CI/CD (Trivy, Snyk)
3. **Create Kubernetes manifests** for production deployment
4. **Set up monitoring** with Prometheus and Grafana
5. **Implement backup strategy** for PostgreSQL data
6. **Configure log aggregation** (ELK, Loki)

## Support

For issues or questions:
- Check existing [GitHub Issues](https://github.com/Kisanlink/farmers-module/issues)
- Review project documentation in `/docs`
- Contact the development team

## License

See LICENSE file in project root.
