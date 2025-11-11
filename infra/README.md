# Farmers Module - AWS Infrastructure

Production-ready AWS infrastructure for the Farmers Module microservice, deployed on ECS Fargate with RDS PostgreSQL (PostGIS), Application Load Balancer, and complete CI/CD pipeline.

## Directory Contents

```
.infra/
├── README.md                          # This file
├── infrastructure_analysis.md         # Comprehensive infrastructure analysis
├── cloudformation.yaml                # Complete AWS infrastructure template
├── buildspec.yml                      # CodeBuild specification for Docker builds
├── appspec.yaml                       # CodeDeploy specification (reference)
├── taskdef.json                       # ECS task definition template
├── .env.production.example            # Production environment variables
└── DEPLOYMENT_GUIDE.md                # Complete deployment guide
```

## Quick Links

- **Start Here:** [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md)
- **Infrastructure Details:** [infrastructure_analysis.md](./infrastructure_analysis.md)
- **CloudFormation Template:** [cloudformation.yaml](./cloudformation.yaml)

## Architecture Overview

### Components

**Compute:**
- ECS Fargate (0.5 vCPU, 1 GB RAM per task)
- Auto-scaling (2-10 tasks based on CPU, memory, request count)
- Blue/Green deployment with circuit breaker

**Database:**
- RDS PostgreSQL 16 with PostGIS extension
- Multi-AZ for production
- Automated backups (7-35 days retention)
- gp3 storage with auto-scaling (100-500 GB)

**Networking:**
- VPC with public/private/database subnets
- Application Load Balancer (HTTPS with ACM certificate)
- NAT Gateways for private subnet internet access
- Security groups with least-privilege access

**CI/CD:**
- GitHub source integration
- CodeBuild for Docker image builds
- CodePipeline for automated deployments
- ECR for container image storage

**Observability:**
- CloudWatch Logs (7-30 day retention)
- CloudWatch Metrics and Alarms
- Container Insights enabled
- OpenTelemetry support (optional)

**Security:**
- AWS Secrets Manager for sensitive data
- IAM roles with least-privilege permissions
- Encryption at rest (RDS, S3, Secrets Manager)
- TLS in transit (ALB, RDS)

### Key Features

- **Zero-Downtime Deployments:** Rolling updates with health checks
- **Auto-Scaling:** CPU, memory, and request-based scaling policies
- **High Availability:** Multi-AZ RDS, 2+ ECS tasks across AZs
- **Cost Optimized:** ~$340/month for production baseline
- **Production Ready:** Comprehensive monitoring, alerting, and rollback procedures

## Prerequisites

Before deploying, ensure you have:

1. AWS Account with appropriate permissions
2. AWS CLI v2 installed and configured
3. SSL/TLS certificate in AWS Certificate Manager
4. AAA service endpoint and API key
5. GitHub personal access token (for CodePipeline)

See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md#prerequisites) for detailed requirements.

## Quick Start

### 1. Prepare Secrets

```bash
# Generate strong random secrets
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-20)
AAA_API_KEY=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)
SECRET_KEY=$(openssl rand -hex 32)

# Store securely (password manager, not version control)
```

### 2. Deploy CloudFormation Stack

```bash
aws cloudformation create-stack \
  --stack-name farmers-module-production \
  --template-body file://.infra/cloudformation.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=production \
    ParameterKey=DBPassword,ParameterValue="$DB_PASSWORD" \
    ParameterKey=AAAApiKey,ParameterValue="$AAA_API_KEY" \
    ParameterKey=JWTSecret,ParameterValue="$JWT_SECRET" \
    ParameterKey=SecretKey,ParameterValue="$SECRET_KEY" \
    ParameterKey=CertificateArn,ParameterValue="arn:aws:acm:..." \
    ParameterKey=GitHubToken,ParameterValue="ghp_..." \
    ParameterKey=AAAServiceEndpoint,ParameterValue="aaa-service.internal:50052" \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1

# Wait for completion (15-20 minutes)
aws cloudformation wait stack-create-complete \
  --stack-name farmers-module-production \
  --region us-east-1
```

### 3. Get Stack Outputs

```bash
aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs' \
  --output table

# Note the ALB DNS name for access
```

### 4. Build and Push Initial Docker Image

```bash
# Get ECR URI from outputs
ECR_URI=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' \
  --output text)

# Authenticate to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin $ECR_URI

# Build and push
docker build -t $ECR_URI:latest -f deployment/docker/Dockerfile .
docker push $ECR_URI:latest
```

### 5. Trigger ECS Deployment

```bash
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --force-new-deployment \
  --region us-east-1
```

### 6. Verify Deployment

```bash
# Get ALB DNS
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \
  --output text)

# Test health endpoint
curl https://$ALB_DNS/health
# Expected: {"status":"ok","service":"farmers-module"}

# Test API docs
curl https://$ALB_DNS/docs
```

## Infrastructure Details

### Resource Sizing

**ECS Tasks:**
- CPU: 512 (0.5 vCPU)
- Memory: 1024 MB (1 GB)
- Startup Time: 25-35 seconds (cold start with migrations)

**RDS PostgreSQL:**
- Instance: db.t4g.medium (2 vCPU, 4 GB RAM)
- Storage: 100 GB gp3 (auto-scaling to 500 GB)
- Multi-AZ: Enabled (production)
- PostGIS: Enabled

**Auto-Scaling:**
- Min Tasks: 2
- Max Tasks: 10
- Target CPU: 60%
- Target Memory: 70%
- Target Requests: 1000/target/minute

### Cost Estimation

**Monthly Cost (Production):** ~$340/month
- ECS Fargate: $36
- RDS PostgreSQL: $231
- ALB: $46
- Data Transfer: $5
- CloudWatch: $21
- Secrets Manager: $2

See [infrastructure_analysis.md](./infrastructure_analysis.md#cost-estimation) for detailed breakdown.

### Security Configuration

**Secrets Management:**
- Database password: AWS Secrets Manager
- AAA API key: AWS Secrets Manager
- JWT secret: AWS Secrets Manager
- Application secret: AWS Secrets Manager

**Network Security:**
- ECS tasks in private subnets (no public IPs)
- Security groups with least-privilege rules
- RDS accessible only from ECS security group
- ALB in public subnets with HTTPS termination

**IAM Roles:**
- Task Execution Role: Pull images, fetch secrets
- Task Role: Application permissions (logs, secrets)

## Health Checks

**ALB Target Group:**
- Path: `/health`
- Interval: 30 seconds
- Timeout: 5 seconds
- Healthy Threshold: 2
- Unhealthy Threshold: 3
- Initial Delay: 45 seconds (accounts for migrations)

**Container Health Check:**
- Disabled (migrations run on startup; rely on ALB only)

## Monitoring & Alerts

### CloudWatch Alarms

**Critical (Immediate Action):**
- ALB Unhealthy Targets > 0 for 5 minutes
- RDS CPU > 80% for 10 minutes
- RDS Free Storage < 10 GB
- ECS Desired ≠ Running for 10 minutes
- ALB 5xx Error Rate > 5% for 5 minutes

**Warning (Investigation):**
- ALB 4xx Error Rate > 10% for 10 minutes
- RDS Connection Count > 80 (80% of max)
- ECS Memory Utilization > 80%
- ALB Target Response Time > 1s (p99)

### Logs

**CloudWatch Log Group:** `/ecs/farmers-module-production`
**Retention:** 30 days (production), 7 days (staging)

```bash
# Tail logs in real-time
aws logs tail /ecs/farmers-module-production --follow

# Filter by error level
aws logs tail /ecs/farmers-module-production --follow --filter-pattern "ERROR"
```

## CI/CD Pipeline

### Pipeline Stages

1. **Source:** GitHub repository (webhook on push)
2. **Build:** CodeBuild builds Docker image, pushes to ECR
3. **Deploy:** CodePipeline updates ECS service with new image

### Build Process

- Build time: 3-5 minutes
- Docker multi-stage build (Go 1.24.4)
- Image scanning enabled (ECR scan on push)
- Artifacts: Docker image in ECR, imagedefinitions.json

### Deployment Strategy

- Type: Rolling update
- Circuit Breaker: Enabled (auto-rollback on failures)
- Minimum Healthy: 100%
- Maximum Percent: 200%

## Troubleshooting

### Common Issues

**ECS Tasks Not Starting:**
- Check CloudWatch Logs for startup errors
- Verify Secrets Manager permissions
- Ensure RDS endpoint is reachable
- Check security group rules (ECS → RDS on port 5432)

**ALB 502/503 Errors:**
- Check target health in ALB console
- Verify health check path returns 200 OK
- Increase health check grace period if migrations are slow
- Check task logs for errors

**Database Migration Failures:**
- Verify PostGIS extension is enabled: `CREATE EXTENSION IF NOT EXISTS postgis;`
- Check RDS parameter group has `postgis` in `shared_preload_libraries`
- Ensure database user has CREATE EXTENSION privilege

**AAA Service Connection Failures:**
- Verify AAA service endpoint is correct
- Check security group allows ECS → AAA on port 50052
- Ensure AAA API key is valid
- AAA seeding failure is non-fatal; service will start but auth may fail

See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md#troubleshooting) for detailed troubleshooting.

## Rollback Procedures

### Rollback ECS Deployment

```bash
# Get previous task definition
PREVIOUS_TASK_DEF=$(aws ecs describe-services \
  --cluster farmers-module-production-cluster \
  --services farmers-module-production-service \
  --region us-east-1 \
  --query 'services[0].deployments[1].taskDefinition' \
  --output text)

# Rollback
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --task-definition $PREVIOUS_TASK_DEF \
  --force-new-deployment \
  --region us-east-1
```

### Rollback Database

```bash
# Restore from automated backup (point-in-time recovery)
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier farmers-module-production-postgres \
  --target-db-instance-identifier farmers-module-production-postgres-rollback \
  --restore-time 2025-11-10T12:00:00Z \
  --region us-east-1
```

## Maintenance

### Secret Rotation (Every 30-90 Days)

```bash
# Rotate database password
NEW_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-20)

# Update in Secrets Manager
aws secretsmanager update-secret \
  --secret-id farmers-module/production/db-password \
  --secret-string "$NEW_PASSWORD" \
  --region us-east-1

# Update RDS
aws rds modify-db-instance \
  --db-instance-identifier farmers-module-production-postgres \
  --master-user-password "$NEW_PASSWORD" \
  --apply-immediately \
  --region us-east-1

# Restart ECS tasks
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --force-new-deployment \
  --region us-east-1
```

### Database Backups

**Automated Backups:**
- Enabled by default
- Retention: 7 days (staging), 35 days (production)
- Backup Window: 03:00-04:00 UTC

**Manual Snapshots:**
```bash
aws rds create-db-snapshot \
  --db-instance-identifier farmers-module-production-postgres \
  --db-snapshot-identifier farmers-manual-snapshot-$(date +%Y%m%d) \
  --region us-east-1
```

## Support

For deployment issues or questions:

1. Check [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md)
2. Review [infrastructure_analysis.md](./infrastructure_analysis.md)
3. Check CloudWatch Logs: `/ecs/farmers-module-production`
4. Contact DevOps team: devops@kisanlink.com

## License

Proprietary - Kisanlink

---

**Version:** 1.0.0
**Last Updated:** 2025-11-10
**Maintained by:** Kisanlink DevOps Team
