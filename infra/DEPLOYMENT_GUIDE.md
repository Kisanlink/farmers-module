# Farmers Module - AWS Deployment Guide
## Production Infrastructure on ECS Fargate with RDS PostgreSQL (PostGIS)

**Version:** 1.0.0
**Last Updated:** 2025-11-10
**Target Environment:** AWS (us-east-1 or any region)

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Pre-Deployment Checklist](#pre-deployment-checklist)
3. [Quick Start (Automated)](#quick-start-automated)
4. [Step-by-Step Deployment](#step-by-step-deployment)
5. [Post-Deployment Validation](#post-deployment-validation)
6. [Operational Procedures](#operational-procedures)
7. [Troubleshooting](#troubleshooting)
8. [Rollback Procedures](#rollback-procedures)
9. [Cost Optimization](#cost-optimization)
10. [Security Best Practices](#security-best-practices)

---

## Prerequisites

### Required Tools

Install the following tools on your deployment machine:

```bash
# AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Verify installation
aws --version  # Should show v2.x.x

# Docker (for local testing)
sudo yum install -y docker
sudo systemctl start docker
sudo usermod -aG docker $USER

# jq (JSON processor)
sudo yum install -y jq

# PostgreSQL client (for database verification)
sudo yum install -y postgresql16
```

### AWS Account Setup

1. **IAM User/Role with Required Permissions:**
   - CloudFormation full access
   - ECS/Fargate full access
   - RDS full access
   - ECR full access
   - VPC management
   - IAM role creation
   - Secrets Manager access
   - S3 access
   - CloudWatch Logs/Alarms

2. **AWS CLI Configuration:**
   ```bash
   aws configure
   # Enter: Access Key ID, Secret Access Key, Region (e.g., us-east-1), Output format (json)

   # Verify credentials
   aws sts get-caller-identity
   ```

3. **Service Quotas:**
   - VPC Elastic IPs: At least 2 (for NAT Gateways)
   - VPC count: At least 1 available
   - RDS instances: At least 1 available
   - ECS tasks: At least 10 (for max autoscaling)

### External Dependencies

1. **SSL/TLS Certificate:**
   - Import or request certificate in AWS Certificate Manager (ACM)
   - Domain: Your production domain (e.g., `farmers.kisanlink.com`)
   - Validation: DNS or Email

   ```bash
   # Request certificate
   aws acm request-certificate \
     --domain-name farmers.kisanlink.com \
     --validation-method DNS \
     --region us-east-1

   # Note the CertificateArn for CloudFormation parameters
   ```

2. **AAA Service Endpoint:**
   - Ensure AAA service is deployed and accessible
   - Note the gRPC endpoint (e.g., `aaa-service.internal:50052`)
   - Obtain API key for authentication

3. **GitHub Repository:**
   - Create GitHub Personal Access Token with `repo` scope
   - Store token securely (needed for CodePipeline)

---

## Pre-Deployment Checklist

Use this checklist before running CloudFormation deployment:

```
Infrastructure:
[ ] AWS account credentials configured
[ ] Target AWS region selected (recommend: us-east-1)
[ ] VPC quota available (or using existing VPC)
[ ] Elastic IP quota available (need 2 for NAT Gateways)

Secrets & Configuration:
[ ] Database master password generated (min 8 chars, alphanumeric)
[ ] AAA API key obtained
[ ] JWT secret generated (min 32 chars, random)
[ ] Application secret key generated (min 32 chars, random)
[ ] SSL/TLS certificate ARN from ACM
[ ] AAA service endpoint confirmed

Repository:
[ ] GitHub personal access token created
[ ] Repository accessible (Kisanlink/farmers-module)
[ ] Target branch identified (main or production)

Domain & DNS:
[ ] Production domain decided (e.g., farmers.kisanlink.com)
[ ] Route 53 hosted zone available (or external DNS provider)
[ ] SSL certificate validated in ACM

Cost Awareness:
[ ] Estimated monthly cost reviewed (~$340/month for production)
[ ] Budget alerts configured (optional)
```

---

## Quick Start (Automated)

For experienced operators, use this automated deployment script:

```bash
#!/bin/bash
# deploy-farmers-module.sh - Automated deployment script

set -e  # Exit on error

# Configuration
STACK_NAME="farmers-module-production"
REGION="us-east-1"
ENVIRONMENT="production"

# Prompt for required parameters
read -sp "Enter RDS master password (min 8 chars): " DB_PASSWORD
echo
read -sp "Enter AAA API key: " AAA_API_KEY
echo
read -sp "Enter JWT secret (min 32 chars): " JWT_SECRET
echo
read -sp "Enter Application secret key (min 32 chars): " SECRET_KEY
echo
read -p "Enter ACM Certificate ARN: " CERT_ARN
read -sp "Enter GitHub Personal Access Token: " GITHUB_TOKEN
echo
read -p "Enter AAA Service Endpoint (e.g., aaa-service.internal:50052): " AAA_ENDPOINT

# Deploy CloudFormation stack
echo "Deploying CloudFormation stack: $STACK_NAME..."
aws cloudformation create-stack \
  --stack-name $STACK_NAME \
  --template-body file://.infra/cloudformation.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=$ENVIRONMENT \
    ParameterKey=DBPassword,ParameterValue="$DB_PASSWORD" \
    ParameterKey=AAAApiKey,ParameterValue="$AAA_API_KEY" \
    ParameterKey=JWTSecret,ParameterValue="$JWT_SECRET" \
    ParameterKey=SecretKey,ParameterValue="$SECRET_KEY" \
    ParameterKey=CertificateArn,ParameterValue="$CERT_ARN" \
    ParameterKey=GitHubToken,ParameterValue="$GITHUB_TOKEN" \
    ParameterKey=AAAServiceEndpoint,ParameterValue="$AAA_ENDPOINT" \
  --capabilities CAPABILITY_NAMED_IAM \
  --region $REGION

# Wait for stack creation
echo "Waiting for stack creation to complete (this may take 15-20 minutes)..."
aws cloudformation wait stack-create-complete \
  --stack-name $STACK_NAME \
  --region $REGION

# Get outputs
echo "Stack created successfully!"
echo "Fetching outputs..."
aws cloudformation describe-stacks \
  --stack-name $STACK_NAME \
  --region $REGION \
  --query 'Stacks[0].Outputs' \
  --output table

echo "Deployment complete! See outputs above for ALB URL and other resources."
```

Save as `deploy-farmers-module.sh`, make executable, and run:

```bash
chmod +x deploy-farmers-module.sh
./deploy-farmers-module.sh
```

---

## Step-by-Step Deployment

For first-time deployments or troubleshooting, follow these detailed steps:

### Step 1: Prepare Secrets

Generate strong secrets before deployment:

```bash
# Generate random secrets
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-20)
AAA_API_KEY=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)
SECRET_KEY=$(openssl rand -hex 32)

echo "Generated secrets (STORE SECURELY):"
echo "DB_PASSWORD: $DB_PASSWORD"
echo "AAA_API_KEY: $AAA_API_KEY"
echo "JWT_SECRET: $JWT_SECRET"
echo "SECRET_KEY: $SECRET_KEY"

# Save to secure location (e.g., password manager, not version control)
```

### Step 2: Validate CloudFormation Template

```bash
aws cloudformation validate-template \
  --template-body file://.infra/cloudformation.yaml \
  --region us-east-1

# Expected output: TemplateDescription and Parameters list
```

### Step 3: Create CloudFormation Stack

Create a parameters file for reusability:

```bash
cat > /tmp/cfn-parameters.json <<EOF
[
  {
    "ParameterKey": "Environment",
    "ParameterValue": "production"
  },
  {
    "ParameterKey": "DBPassword",
    "ParameterValue": "$DB_PASSWORD"
  },
  {
    "ParameterKey": "AAAApiKey",
    "ParameterValue": "$AAA_API_KEY"
  },
  {
    "ParameterKey": "JWTSecret",
    "ParameterValue": "$JWT_SECRET"
  },
  {
    "ParameterKey": "SecretKey",
    "ParameterValue": "$SECRET_KEY"
  },
  {
    "ParameterKey": "CertificateArn",
    "ParameterValue": "arn:aws:acm:us-east-1:123456789012:certificate/xxxxxx"
  },
  {
    "ParameterKey": "GitHubToken",
    "ParameterValue": "ghp_xxxxxxxxxxxx"
  },
  {
    "ParameterKey": "AAAServiceEndpoint",
    "ParameterValue": "aaa-service.internal:50052"
  },
  {
    "ParameterKey": "GitHubBranch",
    "ParameterValue": "main"
  }
]
EOF

# Deploy stack
aws cloudformation create-stack \
  --stack-name farmers-module-production \
  --template-body file://.infra/cloudformation.yaml \
  --parameters file:///tmp/cfn-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --tags Key=Environment,Value=production Key=Service,Value=farmers-module \
  --region us-east-1

# Clean up parameters file (contains secrets)
rm /tmp/cfn-parameters.json
```

### Step 4: Monitor Stack Creation

```bash
# Watch stack events in real-time
aws cloudformation describe-stack-events \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'StackEvents[*].[Timestamp,ResourceStatus,ResourceType,LogicalResourceId]' \
  --output table

# Or use wait command
aws cloudformation wait stack-create-complete \
  --stack-name farmers-module-production \
  --region us-east-1

# Estimated time: 15-20 minutes
# - VPC and networking: 2-3 min
# - NAT Gateways: 2-3 min
# - RDS instance: 8-12 min (Multi-AZ takes longer)
# - ECS cluster and service: 2-3 min
# - ALB and target groups: 1-2 min
```

### Step 5: Retrieve Stack Outputs

```bash
aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs' \
  --output table

# Save important outputs:
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \
  --output text)

RDS_ENDPOINT=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs[?OutputKey==`RDSEndpoint`].OutputValue' \
  --output text)

ECR_URI=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1 \
  --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' \
  --output text)

echo "ALB DNS: $ALB_DNS"
echo "RDS Endpoint: $RDS_ENDPOINT"
echo "ECR URI: $ECR_URI"
```

### Step 6: Build and Push Initial Docker Image

CodePipeline will handle subsequent builds, but you need an initial image:

```bash
# Authenticate to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin $ECR_URI

# Build Docker image
cd /path/to/farmers-module
docker build \
  --build-arg GO_VERSION=1.24.4 \
  --build-arg VERSION=1.0.0 \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  -t $ECR_URI:latest \
  -t $ECR_URI:1.0.0 \
  -f deployment/docker/Dockerfile \
  .

# Push to ECR
docker push $ECR_URI:latest
docker push $ECR_URI:1.0.0

echo "Docker image pushed successfully!"
```

### Step 7: Trigger Initial ECS Deployment

```bash
# Update ECS service to pull new image
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --force-new-deployment \
  --region us-east-1

# Monitor deployment
aws ecs wait services-stable \
  --cluster farmers-module-production-cluster \
  --services farmers-module-production-service \
  --region us-east-1

echo "ECS service deployment complete!"
```

### Step 8: Configure DNS (Optional)

Point your domain to the ALB:

```bash
# If using Route 53
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch '{
    "Changes": [{
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "farmers.kisanlink.com",
        "Type": "CNAME",
        "TTL": 300,
        "ResourceRecords": [{"Value": "'$ALB_DNS'"}]
      }
    }]
  }'

# Or create Route 53 Alias record (recommended for ALB)
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch '{
    "Changes": [{
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "farmers.kisanlink.com",
        "Type": "A",
        "AliasTarget": {
          "HostedZoneId": "<ALB-HOSTED-ZONE-ID>",
          "DNSName": "'$ALB_DNS'",
          "EvaluateTargetHealth": true
        }
      }
    }]
  }'
```

---

## Post-Deployment Validation

### Validation Checklist

```bash
# 1. Verify ALB health
echo "Testing ALB endpoint..."
curl -k https://$ALB_DNS/health
# Expected: {"status":"ok","service":"farmers-module"}

# 2. Check ECS tasks are running
aws ecs describe-services \
  --cluster farmers-module-production-cluster \
  --services farmers-module-production-service \
  --region us-east-1 \
  --query 'services[0].[runningCount,desiredCount]' \
  --output text
# Expected: 2 2 (or your desired count)

# 3. Verify database connection
# Connect to RDS via bastion or ECS Exec
aws ecs execute-command \
  --cluster farmers-module-production-cluster \
  --task <task-id> \
  --container farmers-module \
  --interactive \
  --command "psql -h $RDS_ENDPOINT -U farmers_service -d farmers_production -c 'SELECT PostGIS_version();'"
# Expected: PostGIS version output

# 4. Check CloudWatch Logs
aws logs tail /ecs/farmers-module-production --follow

# 5. Verify autoscaling policies
aws application-autoscaling describe-scaling-policies \
  --service-namespace ecs \
  --resource-id service/farmers-module-production-cluster/farmers-module-production-service \
  --region us-east-1

# 6. Test API documentation
curl -I https://$ALB_DNS/docs
# Expected: HTTP/2 200

# 7. Verify secrets in Secrets Manager
aws secretsmanager list-secrets \
  --filters Key=name,Values=farmers-module/production \
  --region us-east-1

# 8. Check CodePipeline status
aws codepipeline get-pipeline-state \
  --name farmers-module-production-pipeline \
  --region us-east-1
```

### Smoke Tests

Run these API tests to ensure functionality:

```bash
# Health check
curl https://$ALB_DNS/health

# Root endpoint
curl https://$ALB_DNS/

# API docs
curl https://$ALB_DNS/docs

# Sample authenticated endpoint (requires valid JWT)
curl -H "Authorization: Bearer <token>" https://$ALB_DNS/api/v1/farmers

# Expected: Varies by endpoint, but should not return 500 errors
```

### Database Verification

```bash
# Connect to RDS (from ECS task or bastion)
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U farmers_service -d farmers_production

# Run verification queries
SELECT PostGIS_version();
\dt  -- List all tables
SELECT COUNT(*) FROM farmers;
SELECT COUNT(*) FROM farms;
\q
```

---

## Operational Procedures

### Viewing Logs

```bash
# Tail logs in real-time
aws logs tail /ecs/farmers-module-production --follow

# Filter by specific string
aws logs tail /ecs/farmers-module-production --follow --filter-pattern "ERROR"

# View logs from specific time range
aws logs tail /ecs/farmers-module-production \
  --since 1h \
  --format short
```

### Scaling ECS Service

```bash
# Manual scale (overrides autoscaling temporarily)
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --desired-count 5 \
  --region us-east-1

# Update autoscaling limits
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --resource-id service/farmers-module-production-cluster/farmers-module-production-service \
  --scalable-dimension ecs:service:DesiredCount \
  --min-capacity 3 \
  --max-capacity 15 \
  --region us-east-1
```

### Database Maintenance

```bash
# Create manual snapshot
aws rds create-db-snapshot \
  --db-instance-identifier farmers-module-production-postgres \
  --db-snapshot-identifier farmers-manual-snapshot-$(date +%Y%m%d-%H%M%S) \
  --region us-east-1

# List snapshots
aws rds describe-db-snapshots \
  --db-instance-identifier farmers-module-production-postgres \
  --region us-east-1

# Restore from snapshot (creates new instance)
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier farmers-module-production-postgres-restore \
  --db-snapshot-identifier farmers-manual-snapshot-20251110-120000 \
  --region us-east-1
```

### Secret Rotation

```bash
# Rotate database password
NEW_DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-20)

# Update secret in Secrets Manager
aws secretsmanager update-secret \
  --secret-id farmers-module/production/db-password \
  --secret-string "$NEW_DB_PASSWORD" \
  --region us-east-1

# Update RDS master password
aws rds modify-db-instance \
  --db-instance-identifier farmers-module-production-postgres \
  --master-user-password "$NEW_DB_PASSWORD" \
  --apply-immediately \
  --region us-east-1

# Restart ECS tasks to pick up new secret
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --force-new-deployment \
  --region us-east-1
```

---

## Troubleshooting

### Common Issues

#### 1. ECS Tasks Failing to Start

**Symptoms:** Tasks start but immediately stop, status shows "STOPPED"

**Diagnosis:**
```bash
# Get task details
aws ecs describe-tasks \
  --cluster farmers-module-production-cluster \
  --tasks <task-id> \
  --region us-east-1

# Check stopped reason
aws ecs describe-tasks \
  --cluster farmers-module-production-cluster \
  --tasks <task-id> \
  --region us-east-1 \
  --query 'tasks[0].stopCode'
```

**Solutions:**
- Check CloudWatch Logs for startup errors
- Verify Secrets Manager permissions (executionRoleArn)
- Ensure RDS endpoint is reachable from private subnets
- Check security group rules (ECS → RDS on port 5432)

#### 2. Database Migration Failures

**Symptoms:** Logs show "failed to setup database" or migration errors

**Diagnosis:**
```bash
# Check logs for migration errors
aws logs tail /ecs/farmers-module-production --filter-pattern "migration"

# Connect to database and check schema
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U farmers_service -d farmers_production -c '\dt'
```

**Solutions:**
- Verify PostGIS extension: `CREATE EXTENSION IF NOT EXISTS postgis;`
- Check RDS parameter group has `postgis` in `shared_preload_libraries`
- Manually run migrations if needed (see db.go)
- Ensure database user has CREATE EXTENSION privilege

#### 3. ALB Returns 502/503 Errors

**Symptoms:** ALB health checks failing, targets unhealthy

**Diagnosis:**
```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn> \
  --region us-east-1

# Check ECS task health
aws ecs describe-tasks \
  --cluster farmers-module-production-cluster \
  --tasks <task-id> \
  --region us-east-1 \
  --query 'tasks[0].containers[0].healthStatus'
```

**Solutions:**
- Increase health check grace period (45s minimum for migrations)
- Check task logs for startup errors
- Verify security group allows ALB → ECS on port 8000
- Ensure /health endpoint returns 200 OK

#### 4. AAA Service Connection Failures

**Symptoms:** Logs show "failed to seed AAA roles" or gRPC errors

**Diagnosis:**
```bash
# Check AAA service connectivity from ECS task
aws ecs execute-command \
  --cluster farmers-module-production-cluster \
  --task <task-id> \
  --container farmers-module \
  --interactive \
  --command "nc -zv aaa-service.internal 50052"
```

**Solutions:**
- Verify AAA service endpoint is correct
- Check security group allows ECS → AAA on port 50052
- Ensure AAA API key is valid
- AAA seeding failure is non-fatal; service will start but auth may fail

#### 5. CodePipeline Build Failures

**Symptoms:** CodeBuild stage fails, no Docker image pushed

**Diagnosis:**
```bash
# Get build logs
aws codebuild batch-get-builds \
  --ids <build-id> \
  --region us-east-1

# Check ECR permissions
aws ecr get-login-password --region us-east-1
```

**Solutions:**
- Verify CodeBuild service role has ECR permissions
- Check buildspec.yml syntax
- Ensure Dockerfile path is correct: `deployment/docker/Dockerfile`
- Verify Go version in buildspec matches go.mod

---

## Rollback Procedures

### Rollback ECS Deployment

```bash
# Option 1: Rollback to previous task definition
PREVIOUS_TASK_DEF=$(aws ecs describe-services \
  --cluster farmers-module-production-cluster \
  --services farmers-module-production-service \
  --region us-east-1 \
  --query 'services[0].deployments[1].taskDefinition' \
  --output text)

aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --task-definition $PREVIOUS_TASK_DEF \
  --force-new-deployment \
  --region us-east-1

# Option 2: Rollback to specific image tag
# Update task definition with previous image
aws ecs register-task-definition \
  --cli-input-json file:///tmp/previous-taskdef.json

# Update service
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --task-definition farmers-module-production:<revision> \
  --region us-east-1
```

### Rollback Database Changes

```bash
# Restore from automated backup (point-in-time recovery)
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier farmers-module-production-postgres \
  --target-db-instance-identifier farmers-module-production-postgres-rollback \
  --restore-time 2025-11-10T12:00:00Z \
  --region us-east-1

# After verification, update ECS service to use new endpoint
# (requires task definition update)
```

### Rollback CloudFormation Stack

```bash
# Rollback to previous stack version
aws cloudformation cancel-update-stack \
  --stack-name farmers-module-production \
  --region us-east-1

# Or delete and recreate (DESTRUCTIVE)
aws cloudformation delete-stack \
  --stack-name farmers-module-production \
  --region us-east-1
```

---

## Cost Optimization

### Monthly Cost Breakdown

**Production Baseline (~$340/month):**
- ECS Fargate: $36/month (2 tasks, 0.5 vCPU, 1 GB)
- RDS PostgreSQL: $231/month (db.t4g.medium, Multi-AZ, 100 GB)
- ALB: $46/month
- Data Transfer: $5/month
- CloudWatch: $21/month
- Secrets Manager: $2/month

### Optimization Strategies

1. **Use Reserved Instances/Savings Plans:**
   ```bash
   # RDS Reserved Instance (1-year, no upfront)
   # Saves ~40% on RDS costs = $92/month savings

   # Fargate Savings Plan (1-year)
   # Saves ~20% on Fargate = $7/month savings
   ```

2. **Right-Size Resources:**
   ```bash
   # Staging environment: Use smaller instances
   # - db.t4g.small instead of db.t4g.medium ($105/month savings)
   # - Single-AZ RDS ($115/month savings for staging)
   # - 0.25 vCPU Fargate tasks ($18/month savings)
   ```

3. **Reduce Data Transfer:**
   ```bash
   # Use VPC Endpoints for AWS services (avoid NAT Gateway costs)
   aws ec2 create-vpc-endpoint \
     --vpc-id <vpc-id> \
     --service-name com.amazonaws.us-east-1.secretsmanager \
     --route-table-ids <route-table-id>
   ```

4. **Optimize CloudWatch Costs:**
   ```bash
   # Reduce log retention for non-production
   aws logs put-retention-policy \
     --log-group-name /ecs/farmers-module-staging \
     --retention-in-days 3

   # Archive old logs to S3 (cheaper long-term storage)
   ```

---

## Security Best Practices

### 1. Secrets Management

- Store all secrets in AWS Secrets Manager
- Enable automatic secret rotation (30-90 days)
- Use IAM roles for AWS service access (no hardcoded credentials)
- Never log sensitive data (passwords, API keys, tokens)

### 2. Network Security

- Deploy ECS tasks in private subnets (no public IPs)
- Use security groups with least-privilege rules
- Enable VPC Flow Logs for network monitoring
- Use AWS WAF on ALB for DDoS protection

### 3. Database Security

- Enable RDS encryption at rest (enabled by default in template)
- Use SSL/TLS for database connections (`sslmode=require`)
- Restrict database access to ECS security group only
- Enable automated backups with retention
- Enable Multi-AZ for high availability (production)

### 4. Container Security

- Scan Docker images for vulnerabilities (ECR scan on push)
- Run containers as non-root user (already done in Dockerfile)
- Use minimal base images (Alpine Linux)
- Regularly update base images and dependencies

### 5. Access Control

- Use IAM roles with least-privilege permissions
- Enable CloudTrail for audit logging
- Enable MFA for AWS console access
- Rotate access keys and tokens regularly

### 6. Monitoring & Alerting

- Set up CloudWatch Alarms for critical metrics
- Enable AWS GuardDuty for threat detection
- Configure SNS notifications for alarms
- Review CloudWatch Logs regularly for anomalies

---

## Appendix: Reference Commands

### ECS Commands

```bash
# List clusters
aws ecs list-clusters

# Describe service
aws ecs describe-services --cluster <cluster> --services <service>

# List tasks
aws ecs list-tasks --cluster <cluster> --service-name <service>

# Describe task
aws ecs describe-tasks --cluster <cluster> --tasks <task-id>

# Stop task (force restart)
aws ecs stop-task --cluster <cluster> --task <task-id>

# Update service
aws ecs update-service --cluster <cluster> --service <service> --desired-count 3
```

### RDS Commands

```bash
# Describe instance
aws rds describe-db-instances --db-instance-identifier <instance-id>

# Create snapshot
aws rds create-db-snapshot --db-instance-identifier <instance-id> --db-snapshot-identifier <snapshot-id>

# Modify instance
aws rds modify-db-instance --db-instance-identifier <instance-id> --db-instance-class db.t4g.large --apply-immediately

# Reboot instance
aws rds reboot-db-instance --db-instance-identifier <instance-id>
```

### CloudWatch Commands

```bash
# Tail logs
aws logs tail <log-group> --follow

# Get metric statistics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=<service> \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average

# List alarms
aws cloudwatch describe-alarms --alarm-name-prefix farmers-module
```

---

## Support & Escalation

For deployment issues:

1. Check CloudWatch Logs: `/ecs/farmers-module-production`
2. Review CloudFormation Events: Stack → Events tab
3. Consult this guide: [Troubleshooting](#troubleshooting)
4. Contact DevOps team: devops@kisanlink.com

For production incidents:

1. Check [CloudWatch Alarms](#cloudwatch-alarms)
2. Review [Rollback Procedures](#rollback-procedures)
3. Escalate to on-call engineer if critical

---

**End of Deployment Guide**

**Version:** 1.0.0
**Maintained by:** Kisanlink DevOps Team
**Last Reviewed:** 2025-11-10
