# Docker Containerization Specification

## Overview
Implement Docker containerization for the Farmers Module to ensure consistent builds and deployments across all environments, eliminating "works on my machine" issues.

## Objectives
1. Create reproducible build artifacts through Docker multi-stage builds
2. Provide environment-specific configurations (dev, staging, production)
3. Integrate PostgreSQL 16 with PostGIS extension
4. Optimize image size and build times
5. Enable local development with docker-compose

## Technical Requirements

### Directory Structure
```
deployment/
└── docker/
    ├── Dockerfile
    ├── docker-compose.yml
    ├── docker-compose.dev.yml
    ├── docker-compose.staging.yml
    ├── docker-compose.prod.yml
    └── .env.example
```

### Dockerfile Requirements
- **Base Image**: Official Go 1.24+ Alpine image for small footprint
- **Multi-stage build**:
  - Stage 1 (builder): Compile Go application with all dependencies
  - Stage 2 (runtime): Minimal Alpine image with only the binary
- **Build optimizations**:
  - Layer caching for Go modules
  - CGO disabled for static binaries
  - Strip debug symbols for smaller binaries
- **Security**:
  - Run as non-root user
  - No sensitive data in image layers
  - Minimal attack surface

### Docker Compose Configuration

#### Services
1. **farmers-service**
   - Built from local Dockerfile
   - Exposed ports: 8000 (HTTP), 8081 (gRPC)
   - Depends on: postgres, (optional: aaa-service)
   - Health checks enabled
   - Volume mounts for development

2. **postgres**
   - Image: postgres:16-alpine with PostGIS
   - Environment:
     - POSTGRES_DB=farmers_module
     - POSTGRES_USER=postgres
     - POSTGRES_PASSWORD=postgres
   - PostGIS extension auto-enabled
   - Persistent volume for data
   - Health checks enabled

3. **aaa-service** (optional, for local testing)
   - External service connection or mock
   - Port: 50051 (gRPC)

#### Environment-Specific Overrides

**docker-compose.dev.yml**
- Hot-reload enabled with volume mounts
- Debug logging
- Local database exposed on host port
- Development-friendly settings

**docker-compose.staging.yml**
- Production-like configuration
- Metrics and monitoring enabled
- Resource limits applied
- External database connection

**docker-compose.prod.yml**
- Optimized for production
- Strict resource limits
- Read-only root filesystem
- Production security settings
- External database and services

### Build Specifications

#### Dockerfile Build Args
- `GO_VERSION`: Go toolchain version (default: 1.24.4)
- `BUILD_DATE`: Build timestamp
- `VERSION`: Application version from VERSION file
- `GIT_COMMIT`: Git commit hash

#### Image Tagging Strategy
- Development: `farmers-module:dev`
- Staging: `farmers-module:staging-${GIT_SHA}`
- Production: `farmers-module:${VERSION}`

### .dockerignore
Exclude unnecessary files from build context:
- `.git/`
- `*.md` (except required docs)
- `coverage.out`, `coverage.html`
- `farmers-server` binary
- `.env`, `.env.*`
- `tmp/`, `temp/`
- Test files in development

### Integration with Existing Build System

#### Makefile Additions
```makefile
# Docker commands
docker-build          # Build Docker image
docker-up             # Start services with docker-compose
docker-down           # Stop and remove containers
docker-dev            # Start development environment
docker-test           # Run tests in Docker
docker-logs           # View service logs
docker-shell          # Open shell in running container
```

### Database Initialization
- Automatic PostGIS extension installation
- GORM auto-migration on service startup
- Seed data support via environment variable
- Migration rollback capability

### Health Checks
- **farmers-service**: HTTP GET /health endpoint
- **postgres**: pg_isready command
- Startup probes, liveness probes, readiness probes

### Networking
- Custom bridge network: `farmers-network`
- Service discovery via DNS
- Internal communication on dedicated network
- Only required ports exposed to host

### Volume Management
- **postgres-data**: Persistent PostgreSQL data
- **app-logs**: Application logs (optional)
- **dev-workspace**: Source code mount for development

## Security Considerations
1. No secrets in Dockerfile or docker-compose
2. Use .env files for sensitive configuration
3. Non-root user in container
4. Read-only filesystem where possible
5. Minimal base images (Alpine)
6. Regular security updates via base image updates

## Testing Strategy
1. Test multi-stage build produces working binary
2. Verify PostgreSQL + PostGIS connectivity
3. Confirm auto-migration success
4. Test health check endpoints
5. Validate environment variable substitution
6. Test all docker-compose configurations

## Documentation Requirements
1. Quick start guide for developers
2. Build and deployment instructions
3. Environment variable reference
4. Troubleshooting guide
5. Production deployment checklist

## Success Criteria
- [ ] Docker build completes in under 5 minutes
- [ ] Image size under 50MB (runtime stage)
- [ ] All tests pass in Docker environment
- [ ] Services start successfully with docker-compose up
- [ ] Database migrations apply correctly
- [ ] Health checks pass for all services
- [ ] Development hot-reload works
- [ ] Documentation complete and tested

## Implementation Notes
- Follow Docker best practices for layer caching
- Use BuildKit for faster builds
- Implement graceful shutdown handling
- Support both ARM64 and AMD64 architectures
- Version pin all base images
- Use explicit dependency versions in go.mod

## Future Enhancements
- Kubernetes deployment manifests (Helm charts)
- CI/CD pipeline integration
- Container image scanning
- Multi-architecture builds
- Docker Swarm orchestration (if needed)
