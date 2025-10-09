# Docker Quick Start Guide

## TL;DR - Get Started in 3 Commands

```bash
# 1. Copy environment file
cp deployment/docker/.env.example deployment/docker/.env

# 2. Start development environment
make docker-dev

# 3. Access the API
curl http://localhost:8000/health
```

## Services Available

| Service | URL | Credentials |
|---------|-----|-------------|
| Farmers API | http://localhost:8000 | N/A |
| API Docs | http://localhost:8000/docs | N/A |
| pgAdmin | http://localhost:5050 | admin@farmers.local / admin |
| PostgreSQL | localhost:5432 | postgres / postgres |

## Common Commands

### Start/Stop

```bash
# Start development
make docker-dev

# Start in background
make docker-up

# Stop everything
make docker-down
```

### Monitoring

```bash
# View logs
make docker-logs

# View app logs only
make docker-logs-app

# Check service health
make docker-health
```

### Development

```bash
# Restart after code changes
make docker-restart-app

# Access container shell
make docker-shell

# Access database
make docker-shell-db
```

### Testing

```bash
# Run tests in Docker
make docker-test
```

### Cleanup

```bash
# Remove containers
make docker-down

# Remove containers + volumes (deletes data)
make docker-down-volumes
```

## Troubleshooting

### Container won't start?

```bash
# Check logs
make docker-logs-app

# Check health
make docker-health
```

### Port already in use?

Edit `deployment/docker/.env`:
```bash
SERVICE_PORT=8001  # Change from 8000
```

### Database connection errors?

```bash
# Restart database
cd deployment/docker
docker-compose restart postgres

# Check PostgreSQL logs
make docker-logs-db
```

### Need to reset everything?

```bash
# Warning: This deletes all data!
make docker-down-volumes
make docker-dev
```

## Environment Configurations

### Development (default)
```bash
make docker-dev
```
- Debug logging
- PostgreSQL exposed on port 5432
- pgAdmin available
- Fast iteration

### Staging
```bash
make docker-staging
```
- Production-like settings
- Resource limits enforced
- Metrics enabled

### Production (reference only)
```bash
make docker-prod
```
- Use Kubernetes instead!
- Managed database recommended

## Help & Documentation

```bash
# Show all commands
make help

# Show quick start
make docker-quickstart

# Read full documentation
cat deployment/docker/README.md
```

## Need More Help?

- Full documentation: `deployment/docker/README.md`
- Implementation summary: `DOCKER_IMPLEMENTATION_SUMMARY.md`
- Specification: `.kiro/specs/docker-containerization.md`
