# Farmers Module - Deployment Guide

This guide explains how to update your existing CloudFormation stack with all the fixes and deploy the Farmers Module to AWS.

## Table of Contents
- [Prerequisites](#prerequisites)
- [What Was Fixed](#what-was-fixed)
- [Deployment Strategy](#deployment-strategy)
- [Step-by-Step Deployment](#step-by-step-deployment)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before deploying, ensure you have:

1. **AWS CLI** configured with appropriate credentials
   ```bash
   aws configure
   # or use AWS_PROFILE environment variable
   export AWS_PROFILE=your-profile-name
   ```

2. **Required Secrets** - Update the following in your parameter files:
   - Database password (minimum 8 characters)
   - AAA API key
   - JWT secret (minimum 32 characters)
   - Application secret key (minimum 32 characters)
   - GitHub personal access token
   - ACM certificate ARN for HTTPS

3. **GitHub Repository** - Ensure your code is pushed to the repository

4. **AWS Account** - Note your AWS Account ID (needed for parameter files)

---

## What Was Fixed

### 1. CodeBuild Issues
- ✅ Fixed deprecated ECR login command
- ✅ Corrected Dockerfile path from `.infra/` to `infra/`
- ✅ Fixed artifact paths for CodeDeploy
- ✅ Added dynamic taskdef.json updates

### 2. CloudFormation Template
- ✅ Added CodeDeploy resources for Blue/Green deployment (production)
- ✅ Added green target group for production
- ✅ Fixed buildspec path reference
- ✅ Conditional deployment controller (CODE_DEPLOY for prod, ECS for others)
- ✅ Updated CodePipeline with CodeDeploy integration
- ✅ Added CodeDeploy permissions

### 3. Dockerfile Issues
- ✅ Fixed user/group GID/UID conflicts (65534 → 10001)
- ✅ Added ARG re-declaration in runtime stage
- ✅ Fixed swagger generation with proper fallback
- ✅ Successfully builds 122MB optimized image

### 4. Deployment Strategy
- **Production:** CodeDeploy Blue/Green with zero-downtime
- **Dev/Beta/Staging:** ECS Rolling with circuit breaker

---

## Deployment Strategy

### Production Environment (Blue/Green)
- Uses AWS CodeDeploy for zero-downtime deployments
- Automated traffic shifting between blue and green task sets
- Automatic rollback on deployment failure or alarms
- 5-minute termination wait time for old tasks
- Requires `appspec.yaml` and `taskdef.json`

### Non-Production Environments (Rolling)
- Uses ECS native rolling deployments
- Circuit breaker with automatic rollback
- Faster deployments (no traffic shifting delay)
- Uses `imagedefinitions.json` only

---

## Step-by-Step Deployment

### Step 1: Update Parameter Files

Update the secrets in your parameter files (replace `REPLACE_WITH_SECRET`):

```bash
# For production
vim infra/params-production.json

# For dev/beta/staging
vim infra/params-dev.json
vim infra/params-beta.json
vim infra/params-staging.json
```

**Required parameters to update:**
- `DBPassword` - Strong database password
- `AAAApiKey` - Your AAA service API key
- `JWTSecret` - 32+ character random string
- `SecretKey` - 32+ character random string
- `GitHubToken` - GitHub personal access token with repo access
- `CertificateArn` - Your ACM certificate ARN

**Example for generating secrets:**
```bash
# Generate random secrets (macOS/Linux)
openssl rand -base64 32  # For JWT_SECRET
openssl rand -base64 32  # For SECRET_KEY
openssl rand -base64 24  # For DB_PASSWORD
```

### Step 2: Validate CloudFormation Template

```bash
aws cloudformation validate-template \
  --template-body file://infra/cloudformation.yaml \
  --region us-east-1
```

### Step 3: Check Existing Stack Status

```bash
# List existing stacks
aws cloudformation list-stacks \
  --stack-status-filter CREATE_COMPLETE UPDATE_COMPLETE \
  --query 'StackSummaries[?contains(StackName, `farmers-module`)].[StackName,StackStatus]' \
  --output table

# Get details of existing stack (if any)
aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --region us-east-1
```

### Step 4: Update Existing Stack (Recommended)

**For Production:**
```bash
aws cloudformation update-stack \
  --stack-name farmers-module-production \
  --template-body file://infra/cloudformation.yaml \
  --parameters file://infra/params-production.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

**For Dev/Beta/Staging:**
```bash
# Dev
aws cloudformation update-stack \
  --stack-name farmers-module-dev \
  --template-body file://infra/cloudformation.yaml \
  --parameters file://infra/params-dev.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1

# Beta
aws cloudformation update-stack \
  --stack-name farmers-module-beta \
  --template-body file://infra/cloudformation.yaml \
  --parameters file://infra/params-beta.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1

# Staging
aws cloudformation update-stack \
  --stack-name farmers-module-staging \
  --template-body file://infra/cloudformation.yaml \
  --parameters file://infra/params-staging.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

### Step 5: Monitor Stack Update

```bash
# Watch stack events in real-time
aws cloudformation describe-stack-events \
  --stack-name farmers-module-production \
  --max-items 20 \
  --query 'StackEvents[*].[Timestamp,ResourceStatus,ResourceType,LogicalResourceId]' \
  --output table

# Or use AWS Console:
# https://console.aws.amazon.com/cloudformation/
```

**Stack update typically takes 15-30 minutes** depending on:
- VPC and networking resources
- RDS instance provisioning
- ECS service updates
- CodeDeploy configuration

### Step 6: Wait for Stack Update Complete

```bash
aws cloudformation wait stack-update-complete \
  --stack-name farmers-module-production \
  --region us-east-1

echo "Stack update completed!"
```

### Step 7: Verify Resources Created

```bash
# Get stack outputs
aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --query 'Stacks[0].Outputs' \
  --output table

# Important outputs:
# - ALBDNSName: Load balancer endpoint
# - ALBURL: Full HTTPS URL
# - ECRRepositoryURI: Container registry
# - CodePipelineName: Pipeline name
# - HealthCheckURL: Health check endpoint
```

### Step 8: Trigger First Deployment

The CloudFormation stack creates the pipeline, but you need to trigger the first build:

**Option A: Push code to trigger webhook**
```bash
git add .
git commit -m "deploy: update infrastructure configuration"
git push origin main
```

**Option B: Manually start pipeline**
```bash
aws codepipeline start-pipeline-execution \
  --name farmers-module-production-pipeline \
  --region us-east-1
```

**Option C: Use AWS Console**
1. Go to CodePipeline console
2. Find `farmers-module-production-pipeline`
3. Click "Release change"

### Step 9: Monitor Deployment

```bash
# Watch CodePipeline execution
aws codepipeline get-pipeline-state \
  --name farmers-module-production-pipeline \
  --query 'stageStates[*].[stageName,latestExecution.status]' \
  --output table

# Watch CodeBuild logs
aws logs tail /aws/codebuild/farmers-module-production-build --follow

# For production (CodeDeploy Blue/Green)
aws deploy get-deployment \
  --deployment-id <DEPLOYMENT_ID> \
  --query 'deploymentInfo.{status:status,description:description}'

# Watch ECS service status (all environments)
aws ecs describe-services \
  --cluster farmers-module-production-cluster \
  --services farmers-module-production-service \
  --query 'services[0].{desired:desiredCount,running:runningCount,status:status}'
```

### Step 10: Verify Application

```bash
# Get ALB DNS name
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \
  --output text)

# Test health endpoint
curl -k https://${ALB_DNS}/health

# Expected response:
# {"status":"healthy","timestamp":"2025-11-11T...","components":{...}}

# Test API docs
curl -k https://${ALB_DNS}/docs
```

---

## Important Notes for Production

### ECS Service Update Considerations

⚠️ **WARNING:** When updating from ECS rolling to CodeDeploy Blue/Green deployment controller:

The `DeploymentController` property **CANNOT be changed** on an existing ECS service. You have two options:

#### Option 1: Recreate ECS Service (Recommended for Production)
```bash
# 1. Scale down to 0 tasks first
aws ecs update-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --desired-count 0

# 2. Delete the service
aws ecs delete-service \
  --cluster farmers-module-production-cluster \
  --service farmers-module-production-service \
  --force

# 3. Update CloudFormation stack (will create new service with CODE_DEPLOY)
aws cloudformation update-stack \
  --stack-name farmers-module-production \
  --template-body file://infra/cloudformation.yaml \
  --parameters file://infra/params-production.json \
  --capabilities CAPABILITY_NAMED_IAM
```

#### Option 2: Create New Stack with Different Name
```bash
# Update ServiceName parameter in params-production.json to "farmers-module-v2"
aws cloudformation create-stack \
  --stack-name farmers-module-production-v2 \
  --template-body file://infra/cloudformation.yaml \
  --parameters file://infra/params-production.json \
  --capabilities CAPABILITY_NAMED_IAM

# After verification, delete old stack
aws cloudformation delete-stack --stack-name farmers-module-production
```

#### Option 3: Keep ECS Rolling for Production (Simplest)

If you want to avoid service recreation, you can:

1. Remove the `UseCodeDeploy` condition from the CloudFormation template
2. Keep ECS rolling deployment for all environments
3. Still benefit from circuit breaker and auto-rollback

**To do this:**
```bash
# Edit cloudformation.yaml and change line 235:
# FROM: UseCodeDeploy: !Equals [!Ref Environment, production]
# TO:   UseCodeDeploy: !Ref AWS::NoValue

# This disables CodeDeploy entirely and uses ECS rolling for all environments
```

---

## Troubleshooting

### Common Issues

#### 1. Stack Update Failed - "Resource already exists"
```bash
# Check what resources exist
aws cloudformation describe-stack-resources \
  --stack-name farmers-module-production

# Solution: Delete conflicting resources or use different names
```

#### 2. CodeBuild fails with ECR login error
```bash
# Verify ECR repository exists
aws ecr describe-repositories --repository-names farmers-module-production

# If not, create it manually:
aws ecr create-repository --repository-name farmers-module-production
```

#### 3. Docker build fails in CodeBuild
```bash
# Check CodeBuild logs
aws logs tail /aws/codebuild/farmers-module-production-build --follow

# Common issues:
# - Missing go.mod/go.sum (ensure they're committed)
# - Swagger generation errors (check main.go imports)
# - Network issues (check VPC endpoints if using private build)
```

#### 4. ECS Tasks fail to start
```bash
# Check task logs
aws logs tail /ecs/farmers-module-production --follow

# Check task definition
aws ecs describe-task-definition \
  --task-definition farmers-module-production

# Common issues:
# - Secrets Manager permissions (check task execution role)
# - Environment variables misconfigured
# - Database connection issues (check security groups)
```

#### 5. CodeDeploy deployment fails (Production)
```bash
# Get deployment details
aws deploy get-deployment --deployment-id <ID>

# Common issues:
# - Missing appspec.yaml in artifacts
# - Missing taskdef.json in artifacts
# - Target group configuration mismatch
# - Health check failures
```

#### 6. Health check failures
```bash
# Check container logs
aws logs tail /ecs/farmers-module-production --follow

# Test health endpoint from within VPC
aws ec2-instance-connect ssh --instance-id <bastion-id>
curl http://<private-ip>:8000/health

# Common issues:
# - Database not accessible (check security groups)
# - AAA service not reachable
# - Application startup errors
```

---

## Rollback Procedures

### Rolling Back CloudFormation Stack
```bash
# Option 1: Use previous template
aws cloudformation update-stack \
  --stack-name farmers-module-production \
  --use-previous-template \
  --parameters file://infra/params-production.json.backup

# Option 2: Cancel update in progress
aws cloudformation cancel-update-stack \
  --stack-name farmers-module-production
```

### Rolling Back Application Deployment

**For Production (CodeDeploy):**
```bash
# CodeDeploy automatically rolls back on failure
# Manual rollback:
aws deploy stop-deployment \
  --deployment-id <DEPLOYMENT_ID> \
  --auto-rollback-enabled
```

**For Dev/Beta/Staging (ECS Rolling):**
```bash
# Update to previous task definition revision
aws ecs update-service \
  --cluster farmers-module-dev-cluster \
  --service farmers-module-dev-service \
  --task-definition farmers-module-dev:PREVIOUS_REVISION
```

---

## Post-Deployment Verification

### 1. Check Application Health
```bash
ALB_URL=$(aws cloudformation describe-stacks \
  --stack-name farmers-module-production \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBURL`].OutputValue' \
  --output text)

curl -s ${ALB_URL}/health | jq
```

### 2. Verify Database Connectivity
```bash
# Check ECS logs for database connection
aws logs filter-log-events \
  --log-group-name /ecs/farmers-module-production \
  --filter-pattern "database" \
  --max-items 10
```

### 3. Test API Endpoints
```bash
# View API documentation
open ${ALB_URL}/docs

# Test a simple endpoint
curl -X GET ${ALB_URL}/api/v1/health
```

### 4. Monitor Metrics
```bash
# View CloudWatch dashboard
aws cloudwatch get-dashboard \
  --dashboard-name farmers-module-production

# Or use AWS Console:
# https://console.aws.amazon.com/cloudwatch/
```

---

## Clean Up (If Needed)

To delete the entire stack:

```bash
# WARNING: This will delete ALL resources including database!
aws cloudformation delete-stack \
  --stack-name farmers-module-production

# Wait for deletion
aws cloudformation wait stack-delete-complete \
  --stack-name farmers-module-production
```

---

## Next Steps

After successful deployment:

1. **Set up monitoring alerts** - Configure SNS topics for CloudWatch alarms
2. **Configure custom domain** - Add Route53 alias record to ALB
3. **Set up backup strategy** - RDS automated backups are enabled, verify restore procedures
4. **Enable AWS WAF** (optional) - Add web application firewall rules to ALB
5. **Set up CI/CD notifications** - Configure SNS/Slack notifications for pipeline events
6. **Review security** - Run AWS Security Hub and Trusted Advisor checks

---

## Support

For issues or questions:
- Check CloudFormation events for detailed error messages
- Review CodeBuild/CodePipeline logs
- Consult AWS documentation for specific services
- Check application logs in CloudWatch

---

**Last Updated:** 2025-11-11
**Template Version:** 1.0.0 (with CodeDeploy Blue/Green support)
