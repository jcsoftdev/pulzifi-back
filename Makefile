# ============================================================
# Makefile for Pulzifi Backend Docker Operations
# Simplifies common Docker tasks
# ============================================================

.PHONY: help microservices-dev microservices-prod monolith-dev \
        microservices-down monolith-down \
        microservices-logs monolith-logs \
        clean build swagger

# Default target
.DEFAULT_GOAL := help

# Variables
ENV_FILE := .env

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

# Help target
help: ## Show this help message
	@echo "============================================================"
	@echo "🐳 Pulzifi Backend - THREE DEPLOYMENT MODES"
	@echo "============================================================"
	@echo ""
	@echo "$(GREEN)DEVELOPMENT MODES:$(NC)"
	@echo "  $(YELLOW)make microservices-dev$(NC)   - Microservices with hot reload"
	@echo "  $(YELLOW)make monolith-dev$(NC)         - Monolith all-in-one with hot reload"
	@echo ""
	@echo "$(GREEN)PRODUCTION MODE:$(NC)"
	@echo "  $(YELLOW)make microservices-prod$(NC)  - Microservices production build"
	@echo ""
	@echo "$(GREEN)COMMON COMMANDS:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v "^help" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)EXAMPLES:$(NC)"
	@echo "  make microservices-dev              # Start dev microservices with hot reload"
	@echo "  make monolith-dev                   # Start monolith with hot reload"
	@echo "  make microservices-logs             # View microservices logs"
	@echo "  make clean                          # Stop and cleanup everything"

# Check if required files exist
check-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(YELLOW)⚠️  .env file not found$(NC)"; \
		echo "$(YELLOW)Creating from .env.example...$(NC)"; \
		cp .env.example $(ENV_FILE) 2>/dev/null || echo "{}"; \
		echo "$(GREEN)✓ .env created (please review and update)$(NC)"; \
	fi

# ============================================================
# MODE 1: MICROSERVICES DEVELOPMENT
# Each module in separate container with hot reload
# ============================================================

microservices-dev: check-env ## Start all microservices with hot reload (docker-compose.dev.yml)
	@echo "$(GREEN)🚀 Starting Microservices Development Environment$(NC)"
	@echo "$(YELLOW)Each module runs in its own container with hot reload$(NC)"
	@echo ""
	@echo "$(GREEN)Services:$(NC)"
	@echo "  - PostgreSQL 17 on port 5434"
	@echo "  - Redis 7 on port 6379"
	@echo "  - Kafka on port 9092"
	@echo "  - Auth on port 8080"
	@echo "  - Organization on port 8081"
	@echo "  - Workspace on port 8082"
	@echo "  - Page on port 8083"
	@echo "  - Alert on port 8084"
	@echo "  - Monitoring on port 8085"
	@echo "  - And more..."
	@echo ""
	@docker-compose -f docker-compose.dev.yml up

microservices-down: ## Stop and remove microservices containers
	@echo "$(YELLOW)Stopping microservices...$(NC)"
	@docker-compose -f docker-compose.dev.yml down -v
	@docker-compose -f docker-compose.yml down -v 2>/dev/null || true
	@echo "$(GREEN)✓ Microservices stopped$(NC)"

microservices-logs: ## View microservices logs (use 'make microservices-logs service=SERVICE_NAME')
	@docker-compose -f docker-compose.dev.yml logs -f $(service)

# ============================================================
# MODE 2: MICROSERVICES PRODUCTION
# Production build without hot reload
# ============================================================

microservices-prod: check-env ## Start microservices production build (docker-compose.yml)
	@echo "$(GREEN)🚀 Starting Microservices Production$(NC)"
	@echo "$(YELLOW)Production optimized containers$(NC)"
	@docker-compose -f docker-compose.yml up -d
	@echo "$(GREEN)✓ Microservices started in background$(NC)"

# ============================================================
# MODE 3: MONOLITH DEVELOPMENT
# Monolith with hot reload connected to postgres + redis
# ============================================================

monolith-dev: check-env ## Start monolith with hot reload (postgres + redis + app with hot reload)
	@echo "$(GREEN)🚀 Starting Monolith Development$(NC)"
	@echo "$(YELLOW)Includes: PostgreSQL, Redis, Monolith with hot reload$(NC)"
	@echo ""
	@echo "$(GREEN)Services:$(NC)"
	@echo "  - PostgreSQL on port 5432"
	@echo "  - Redis on port 6379"
	@echo "  - Monolith API on port 8080 (with hot reload)"
	@echo ""
	@docker-compose -f docker-compose.monolith.yml up

monolith-down: ## Stop monolith and services
	@echo "$(YELLOW)Stopping monolith...$(NC)"
	@docker-compose -f docker-compose.monolith.yml down -v
	@echo "$(GREEN)✓ Monolith stopped$(NC)"

monolith-logs: ## View monolith logs
	@docker-compose -f docker-compose.monolith.yml logs -f monolith

# ============================================================
# BUILD TARGETS
# ============================================================

build: check-env ## Build monolith binary locally
	@echo "$(GREEN)Building monolith binary...$(NC)"
	@mkdir -p ./bin
	@go build -o ./bin/pulzifi-monolith ./cmd/server/main.go
	@echo "$(GREEN)✓ Binary built at ./bin/pulzifi-monolith$(NC)"

swagger: ## Generate Swagger docs for monolith (run in cmd/server)
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@cd cmd/server && swag init -g main.go
	@echo "$(GREEN)✓ Swagger docs generated$(NC)"

# ============================================================
# CLEANUP
# ============================================================

clean: ## Stop all containers and cleanup Docker resources
	@echo "$(YELLOW)⚠️  Cleaning up Docker resources...$(NC)"
	@docker-compose -f docker-compose.dev.yml down -v 2>/dev/null || true
	@docker-compose -f docker-compose.yml down -v 2>/dev/null || true
	@docker stop pulzifi-monolith 2>/dev/null || true
	@docker system prune -f --volumes 2>/dev/null || true
	@echo "$(GREEN)✓ Cleanup completed$(NC)"

# ============================================================
# REFERENCE
# ============================================================
