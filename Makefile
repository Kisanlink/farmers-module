# Farmers Module Makefile
# Provides common development and deployment commands

.PHONY: help build test clean run docs

# Default target
help: ## Show this help message
	@echo "Farmers Module - Farm Management Service"
	@echo "========================================"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
build: ## Build the application
	@echo "Building Farmers Module..."
	go build -o farmers-server cmd/farmers-service/main.go
	@echo "✅ Build complete: farmers-server"

test: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	go test ./... -v

test-short: ## Run only unit tests (skip integration tests)
	@echo "Running unit tests..."
	go test ./... -v -short

test-integration: ## Run integration tests with TestContainers
	@echo "Running integration tests with TestContainers..."
	@echo "Note: Requires Docker to be running"
	go test ./... -v -run Integration

test-contract: ## Run contract tests for mock-real service parity
	@echo "Running contract tests..."
	go test ./internal/services -v -run Contract

test-security: ## Run security validation tests
	@echo "Running security tests..."
	go test ./internal/services -v -run Security

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test ./... -bench=. -benchmem -run=^$$

# Documentation
docs: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/farmers-service/main.go --parseVendor --parseDependency
	@echo "✅ Swagger documentation generated"

# Protocol Buffers
proto-gen: ## Regenerate protocol buffer files
	@echo "Regenerating proto files..."
	@# Generate all proto files with dependencies resolved
	cd pkg/proto && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--proto_path=. \
		*.proto
	@echo "✅ Proto files regenerated"

# Development server
run: ## Run the farmers module server
	@echo "Starting Farmers Module server..."
	@echo "HTTP server will be available at: http://localhost:8080"
	@echo "API documentation at: http://localhost:8080/docs"
	@echo ""
	go run cmd/farmers-service/main.go

dev: run ## Alias for run command

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f farmers-server
	@echo "✅ Cleanup complete"

# Quick development workflow
dev-setup: build docs ## Setup development environment with docs
	@echo "✅ Development setup complete"

# Helpers
version: ## Show version information
	@echo "Farmers Module"
	@echo "Version: 1.0.0"
	@echo "Date: $(shell date)"

# ==============================================================================
# Docker Commands
# ==============================================================================

# Docker variables
DOCKER_COMPOSE_DIR := deployment/docker
DOCKER_COMPOSE := docker-compose -f $(DOCKER_COMPOSE_DIR)/docker-compose.yml
DOCKER_COMPOSE_DEV := $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_DIR)/docker-compose.dev.yml
DOCKER_COMPOSE_STAGING := $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_DIR)/docker-compose.staging.yml
DOCKER_COMPOSE_PROD := $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_DIR)/docker-compose.prod.yml

# Build metadata
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")

# Docker build
docker-build: ## Build Docker image (local platform)
	@echo "Building Docker image..."
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	cd $(DOCKER_COMPOSE_DIR) && docker-compose build \
		--build-arg GO_VERSION=1.24.4 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE)
	@echo "✅ Docker image built successfully"

# ==============================================================================
# AWS ECS Deployment Commands
# ==============================================================================

# AWS/ECR variables
AWS_REGION ?= ap-south-1
AWS_ACCOUNT_ID ?= $(shell aws sts get-caller-identity --query Account --output text 2>/dev/null)
ECR_REPO_STAGING := farmers-module-staging
ECR_REPO_PRODUCTION := farmers-module-production
REPOSITORY_URI_STAGING := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(ECR_REPO_STAGING)
REPOSITORY_URI_PRODUCTION := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(ECR_REPO_PRODUCTION)
IMAGE_TAG := $(GIT_COMMIT)

docker-build-ecs: ## Build Docker image for AWS ECS Fargate (linux/amd64)
	@echo "Building Docker image for AWS ECS Fargate (linux/amd64)..."
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo ""
	docker buildx build \
		--platform linux/amd64 \
		--build-arg GO_VERSION=1.24.4 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t farmers-module:$(IMAGE_TAG) \
		-t farmers-module:latest \
		-f deployment/docker/Dockerfile \
		--load \
		.
	@echo "✅ Docker image built successfully for ECS"

docker-build-staging: ## Build and tag for ECR staging repository
	@echo "Building Docker image for ECR staging..."
	@if [ -z "$(AWS_ACCOUNT_ID)" ]; then \
		echo "❌ Error: AWS_ACCOUNT_ID not set. Please configure AWS CLI."; \
		exit 1; \
	fi
	@echo "Repository URI: $(REPOSITORY_URI_STAGING)"
	@echo "Image Tag: $(IMAGE_TAG)"
	@echo ""
	docker buildx build \
		--platform linux/amd64 \
		--build-arg GO_VERSION=1.24.4 \
		--build-arg VERSION=$(IMAGE_TAG) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(REPOSITORY_URI_STAGING):latest \
		-t $(REPOSITORY_URI_STAGING):$(IMAGE_TAG) \
		-f deployment/docker/Dockerfile \
		--load \
		.
	@echo "✅ Docker image built and tagged for staging"

docker-build-production: ## Build and tag for ECR production repository
	@echo "Building Docker image for ECR production..."
	@if [ -z "$(AWS_ACCOUNT_ID)" ]; then \
		echo "❌ Error: AWS_ACCOUNT_ID not set. Please configure AWS CLI."; \
		exit 1; \
	fi
	@echo "Repository URI: $(REPOSITORY_URI_PRODUCTION)"
	@echo "Image Tag: $(VERSION)"
	@echo ""
	docker buildx build \
		--platform linux/amd64 \
		--build-arg GO_VERSION=1.24.4 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(REPOSITORY_URI_PRODUCTION):latest \
		-t $(REPOSITORY_URI_PRODUCTION):$(VERSION) \
		-f deployment/docker/Dockerfile \
		--load \
		.
	@echo "✅ Docker image built and tagged for production"

docker-ecr-login: ## Login to AWS ECR
	@echo "Logging in to ECR..."
	@if [ -z "$(AWS_ACCOUNT_ID)" ]; then \
		echo "❌ Error: AWS_ACCOUNT_ID not set. Please configure AWS CLI."; \
		exit 1; \
	fi
	@aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
	@echo "✅ Logged in to ECR"

docker-push-staging: docker-ecr-login ## Push Docker image to ECR staging
	@echo "Pushing image to ECR staging..."
	docker push $(REPOSITORY_URI_STAGING):latest
	docker push $(REPOSITORY_URI_STAGING):$(IMAGE_TAG)
	@echo "✅ Image pushed to staging"
	@echo ""
	@echo "Image URIs:"
	@echo "  - $(REPOSITORY_URI_STAGING):latest"
	@echo "  - $(REPOSITORY_URI_STAGING):$(IMAGE_TAG)"

docker-push-production: docker-ecr-login ## Push Docker image to ECR production
	@echo "Pushing image to ECR production..."
	docker push $(REPOSITORY_URI_PRODUCTION):latest
	docker push $(REPOSITORY_URI_PRODUCTION):$(VERSION)
	@echo "✅ Image pushed to production"
	@echo ""
	@echo "Image URIs:"
	@echo "  - $(REPOSITORY_URI_PRODUCTION):latest"
	@echo "  - $(REPOSITORY_URI_PRODUCTION):$(VERSION)"

docker-deploy-staging: docker-build-staging docker-push-staging ## Build, push, and deploy to ECS staging
	@echo "Triggering ECS service update..."
	@aws ecs update-service \
		--cluster farmers-module-staging-cluster \
		--service farmers-module-staging-service \
		--force-new-deployment \
		--region $(AWS_REGION) \
		--output json > /dev/null
	@echo "✅ ECS staging deployment triggered"
	@echo ""
	@echo "Monitor deployment:"
	@echo "  aws ecs describe-services --cluster farmers-module-staging-cluster --services farmers-module-staging-service --region $(AWS_REGION)"

docker-deploy-production: docker-build-production docker-push-production ## Build, push, and deploy to ECS production
	@echo "⚠️  WARNING: Deploying to PRODUCTION"
	@read -p "Are you sure you want to deploy to production? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Triggering ECS service update..."; \
		aws ecs update-service \
			--cluster farmers-module-production-cluster \
			--service farmers-module-production-service \
			--force-new-deployment \
			--region $(AWS_REGION) \
			--output json > /dev/null; \
		echo "✅ ECS production deployment triggered"; \
	else \
		echo "Deployment cancelled"; \
	fi

# Development environment
docker-dev: ## Start development environment with hot-reload
	@echo "Starting development environment..."
	@echo "Services will be available at:"
	@echo "  - Farmers API: http://localhost:8000"
	@echo "  - API Docs: http://localhost:8000/docs"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - pgAdmin: http://localhost:5050"
	@echo ""
	cd $(DOCKER_COMPOSE_DIR) && docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

docker-dev-build: ## Build and start development environment
	@echo "Building and starting development environment..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

docker-dev-down: ## Stop development environment
	@echo "Stopping development environment..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose -f docker-compose.yml -f docker-compose.dev.yml down
	@echo "✅ Development environment stopped"

# Production-like environments
docker-up: ## Start services (base configuration)
	@echo "Starting services..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose up -d
	@echo "✅ Services started"

docker-staging: ## Start staging environment
	@echo "Starting staging environment..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d
	@echo "✅ Staging environment started"

docker-prod: ## Start production environment (reference only)
	@echo "⚠️  WARNING: Production environment should use orchestration platforms (Kubernetes)"
	@echo "Starting production environment..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "✅ Production environment started"

# Stop and cleanup
docker-down: ## Stop and remove all containers
	@echo "Stopping and removing containers..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose down
	@echo "✅ Containers stopped and removed"

docker-down-volumes: ## Stop containers and remove volumes (WARNING: deletes data)
	@echo "⚠️  WARNING: This will delete all data in Docker volumes"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd $(DOCKER_COMPOSE_DIR) && docker-compose down -v; \
		echo "✅ Containers and volumes removed"; \
	else \
		echo "Cancelled"; \
	fi

# Logs and monitoring
docker-logs: ## View logs from all services
	@echo "Viewing logs (Ctrl+C to exit)..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose logs -f

docker-logs-app: ## View logs from farmers-service only
	@echo "Viewing farmers-service logs (Ctrl+C to exit)..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose logs -f farmers-service

docker-logs-db: ## View logs from PostgreSQL only
	@echo "Viewing PostgreSQL logs (Ctrl+C to exit)..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose logs -f postgres

# Container management
docker-ps: ## List running containers
	@cd $(DOCKER_COMPOSE_DIR) && docker-compose ps

docker-shell: ## Open shell in running farmers-service container
	@echo "Opening shell in farmers-service container..."
	@cd $(DOCKER_COMPOSE_DIR) && docker-compose exec farmers-service /bin/sh

docker-shell-db: ## Open psql shell in PostgreSQL container
	@echo "Opening psql shell in PostgreSQL container..."
	@cd $(DOCKER_COMPOSE_DIR) && docker-compose exec postgres psql -U postgres -d farmers_module

# Testing in Docker
docker-test: ## Run tests in Docker container
	@echo "Running tests in Docker..."
	docker run --rm \
		-v $(PWD):/app \
		-w /app \
		golang:1.24.4-alpine \
		sh -c "apk add --no-cache git make && go test ./... -v"
	@echo "✅ Tests completed"

# Restart services
docker-restart: ## Restart all services
	@echo "Restarting services..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose restart
	@echo "✅ Services restarted"

docker-restart-app: ## Restart farmers-service only
	@echo "Restarting farmers-service..."
	cd $(DOCKER_COMPOSE_DIR) && docker-compose restart farmers-service
	@echo "✅ farmers-service restarted"

# Health checks
docker-health: ## Check health status of all services
	@echo "Checking health status..."
	@cd $(DOCKER_COMPOSE_DIR) && docker-compose ps

# Clean Docker resources
docker-clean: ## Remove unused Docker resources
	@echo "Cleaning unused Docker resources..."
	docker system prune -f
	@echo "✅ Docker cleanup complete"

docker-clean-all: ## Remove all Docker resources (WARNING: nuclear option)
	@echo "⚠️  WARNING: This will remove ALL Docker resources (images, containers, volumes, networks)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker system prune -a --volumes -f; \
		echo "✅ All Docker resources removed"; \
	else \
		echo "Cancelled"; \
	fi

# Quick start guide
docker-quickstart: ## Quick start guide for Docker setup
	@echo "========================================"
	@echo "Farmers Module - Docker Quick Start"
	@echo "========================================"
	@echo ""
	@echo "1. Copy environment file:"
	@echo "   cp deployment/docker/.env.example deployment/docker/.env"
	@echo ""
	@echo "2. Start development environment:"
	@echo "   make docker-dev"
	@echo ""
	@echo "3. Access services:"
	@echo "   - API: http://localhost:8000"
	@echo "   - Docs: http://localhost:8000/docs"
	@echo "   - pgAdmin: http://localhost:5050"
	@echo ""
	@echo "4. Stop environment:"
	@echo "   make docker-dev-down"
	@echo ""
	@echo "For more commands, run: make help"
	@echo ""

# Default target
.DEFAULT_GOAL := help
