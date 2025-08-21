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

test: ## Run tests
	@echo "Running tests..."
	go test ./... -v

# Documentation
docs: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/farmers-service/main.go
	@echo "✅ Swagger documentation generated"

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

# Default target
.DEFAULT_GOAL := help
