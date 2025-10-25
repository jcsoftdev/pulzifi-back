# ============================================================
# Makefile for Pulzifi Backend Docker Operations
# Simplifies common Docker tasks
# ============================================================

.PHONY: help build start stop restart clean logs status health dev prod

# Default target
.DEFAULT_GOAL := help

# Variables
DOCKER_COMPOSE_FILE := docker-compose.yml
DOCKER_COMPOSE_DEV := docker-compose.dev.yml
ENV_FILE := .env

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m

# Help target
help: ## Show this help message
	@echo "============================================================"
	@echo "🐳 Pulzifi Backend Docker Management"
	@echo "Using Go 1.25 and PostgreSQL 17"
	@echo "============================================================"
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build          # Build all services"
	@echo "  make dev            # Start development environment"
	@echo "  make logs service=alert  # View specific service logs"

# Check if required files exist
check-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(YELLOW)Creating .env from .env.example...$(NC)"; \
		cp .env.example $(ENV_FILE); \
		echo "$(GREEN)Please review and update .env file$(NC)"; \
	fi

# Build all Docker images
build: check-env ## Build all Docker images with latest dependencies
	@echo "$(GREEN)Building all services with Go 1.25...$(NC)"
	@docker-compose build --parallel --no-cache
	@echo "$(GREEN)Build completed!$(NC)"

# Start all services in production mode
start: check-env ## Start all services in production mode
	@echo "$(GREEN)Starting all services...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)Services started! Use 'make status' to check health$(NC)"

# Start development environment
dev: check-env ## Start development environment (infrastructure only)
	@echo "$(GREEN)Starting development environment...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE_DEV) up -d
	@echo "$(GREEN)Development environment ready!$(NC)"
	@echo "Infrastructure services available:"
	@echo "  • PostgreSQL: localhost:5434"
	@echo "  • Redis: localhost:6379"
	@echo "  • Kafka: localhost:9092"

# Start only infrastructure services
infra: check-env ## Start only infrastructure services (PostgreSQL, Redis, Kafka)
	@echo "$(GREEN)Starting infrastructure services...$(NC)"
	@docker-compose up -d postgres redis zookeeper kafka
	@echo "$(GREEN)Infrastructure services started!$(NC)"

# Stop all services
stop: ## Stop all services
	@echo "$(YELLOW)Stopping all services...$(NC)"
	@docker-compose down
	@docker-compose -f $(DOCKER_COMPOSE_DEV) down 2>/dev/null || true
	@echo "$(GREEN)All services stopped!$(NC)"

# Restart all services
restart: ## Restart all services
	@echo "$(YELLOW)Restarting all services...$(NC)"
	@docker-compose restart
	@echo "$(GREEN)Services restarted!$(NC)"

# Clean up Docker resources
clean: ## Stop services and clean up Docker resources
	@echo "$(YELLOW)Cleaning up Docker resources...$(NC)"
	@docker-compose down -v --remove-orphans
	@docker-compose -f $(DOCKER_COMPOSE_DEV) down -v --remove-orphans 2>/dev/null || true
	@docker system prune -f
	@echo "$(GREEN)Cleanup completed!$(NC)"

# Show service status
status: ## Show status of all services
	@echo "$(GREEN)Service Status:$(NC)"
	@docker-compose ps
	@echo ""
	@echo "$(GREEN)Container Health:$(NC)"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep pulzifi || echo "No Pulzifi containers running"

# Show logs for all services or specific service
logs: ## Show logs (use 'make logs service=SERVICE_NAME' for specific service)
ifdef service
	@docker-compose logs -f $(service)
else
	@docker-compose logs --tail=50 -f
endif

# Run health check
health: ## Run comprehensive health check on all services
	@echo "$(GREEN)Running health check...$(NC)"
	@./scripts/docker-health-check.sh

# Update dependencies and rebuild
update: ## Update Go dependencies and rebuild all services
	@echo "$(GREEN)Updating Go dependencies...$(NC)"
	@go mod tidy
	@go mod download
	@echo "$(GREEN)Rebuilding services with updated dependencies...$(NC)"
	@make clean build start
	@echo "$(GREEN)Update completed!$(NC)"

# Database operations
db-setup: ## Initialize database with setup script
	@echo "$(GREEN)Setting up database...$(NC)"
	@docker-compose exec postgres psql -U pulzifi_user -d pulzifi -f /docker-entrypoint-initdb.d/01-init.sql
	@echo "$(GREEN)Database setup completed!$(NC)"

db-shell: ## Open PostgreSQL shell
	@docker-compose exec postgres psql -U pulzifi_user -d pulzifi

db-backup: ## Backup database to sql file
	@mkdir -p backups
	@docker-compose exec postgres pg_dump -U pulzifi_user pulzifi > backups/pulzifi_backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Database backup created in backups/ directory$(NC)"

# Redis operations
redis-shell: ## Open Redis CLI
	@docker-compose exec redis redis-cli -a $${REDIS_PASSWORD:-redis_password}

# Kafka operations
kafka-topics: ## List Kafka topics
	@docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-shell: ## Open Kafka shell
	@docker-compose exec kafka bash

# View specific service logs
logs-auth: ## View auth service logs
	@make logs service=auth-service

logs-alert: ## View alert service logs
	@make logs service=alert-service

logs-postgres: ## View PostgreSQL logs
	@make logs service=postgres

logs-nginx: ## View Nginx logs
	@make logs service=nginx

# Production deployment helpers
prod-build: ## Build production images with optimizations
	@echo "$(GREEN)Building production images...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) build --no-cache --compress
	@echo "$(GREEN)Production build completed!$(NC)"

prod-start: ## Start production environment with all optimizations
	@echo "$(GREEN)Starting production environment...$(NC)"
	@cp .env.docker .env
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d --scale auth-service=2 --scale alert-service=2
	@echo "$(GREEN)Production environment started with load balancing!$(NC)"

# Development helpers
test-service: ## Build and test a specific service (use 'make test-service service=SERVICE_NAME')
ifdef service
	@echo "$(GREEN)Testing $(service) service...$(NC)"
	@docker build --build-arg MODULE_NAME=$(service) -t pulzifi-$(service):test .
	@docker run --rm -e MODULE_NAME=$(service) pulzifi-$(service):test go test ./modules/$(service)/...
else
	@echo "$(YELLOW)Usage: make test-service service=SERVICE_NAME$(NC)"
endif

# Show Docker system information
docker-info: ## Show Docker system information and resource usage
	@echo "$(GREEN)Docker System Information:$(NC)"
	@docker system df
	@echo ""
	@echo "$(GREEN)Running Containers:$(NC)"
	@docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"