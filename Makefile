# ============================================================
# Makefile for Pulzifi Backend
# ============================================================

.PHONY: help dev dev-web down logs build swagger clean migrate cms-migrate test test-integration test-billing-integration test-db-reset check-arch backup

.DEFAULT_GOAL := help

ENV_FILE := .env

GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m

help: ## Show this help message
	@echo "============================================================"
	@echo " Pulzifi Backend"
	@echo "============================================================"
	@echo ""
	@echo "$(GREEN)DEVELOPMENT:$(NC)"
	@echo "  $(YELLOW)make dev$(NC)           - Start local dev environment, run migrations, attach logs"
	@echo "  $(YELLOW)make dev-with-prod$(NC) - Start dev with production DB (no local postgres)"
	@echo "  $(YELLOW)make dev-web$(NC)  - Start Next.js (Go proxies unmatched routes)"
	@echo "  $(YELLOW)make down$(NC)     - Stop local dev environment"
	@echo "  $(YELLOW)make logs$(NC)     - View logs (use: make logs service=monolith)"
	@echo ""
	@echo "$(GREEN)DATABASE:$(NC)"
	@echo "  $(YELLOW)make migrate$(NC)      - Run Go API migrations from .env (make migrate cmd=up|down|version)"
	@echo "  $(YELLOW)make cms-migrate$(NC)  - Push Payload CMS schema + seed data (requires dev server running for seed)"
	@echo ""
	@echo "$(GREEN)TESTS:$(NC)"
	@echo "  $(YELLOW)make test$(NC)                      - Run unit tests (no DB required)"
	@echo "  $(YELLOW)make test-integration$(NC)          - Run ALL integration tests against an isolated pulzifi_test DB"
	@echo "  $(YELLOW)make test-billing-integration$(NC)  - Run billing-only integration tests (stripe webhook + plan assigner)"
	@echo "  $(YELLOW)make test-db-reset$(NC)             - Drop and recreate the pulzifi_test DB from scratch"
	@echo ""
	@echo "$(GREEN)BUILD:$(NC)"
	@echo "  $(YELLOW)make build$(NC)    - Build API binary locally"
	@echo "  $(YELLOW)make swagger$(NC)  - Regenerate Swagger docs"
	@echo "  $(YELLOW)make clean$(NC)    - Stop containers and prune Docker resources"
	@echo ""
	@echo "$(GREEN)OPS:$(NC)"
	@echo "  $(YELLOW)make backup$(NC)   - Run a manual PostgreSQL backup (pg_dump → gzip → MinIO/S3)"

check-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(YELLOW)⚠️  .env file not found$(NC)"; \
		cp .env.example $(ENV_FILE) 2>/dev/null && echo "$(GREEN)✓ .env created from .env.example — please review it$(NC)"; \
	fi

# ============================================================
# LOCAL DEVELOPMENT
# ============================================================

dev: check-env ## Start local dev (postgres + scraper + API + worker with hot reload)
	@./tools/scripts/assign-dev-ports.sh $(ENV_FILE) docker
	@echo "$(GREEN)Starting local dev environment...$(NC)"
	@docker-compose up -d --remove-orphans
	@echo "$(GREEN)Waiting for postgres...$(NC)"
	@until docker-compose exec -T postgres pg_isready > /dev/null 2>&1; do sleep 1; done
	@echo "$(GREEN)Running migrations...$(NC)"
	@export $(shell grep -v '^#' $(ENV_FILE) | xargs) && \
		go run ./cmd/migrate \
			-db "postgres://$${DB_USER}:$${DB_PASSWORD}@localhost:$${DB_PORT}/$${DB_NAME}?sslmode=disable" \
			-cmd up
	@echo "$(GREEN)Ready. Attaching to logs (Ctrl+C detaches — containers keep running)...$(NC)"
	@docker-compose logs -f

dev-with-prod: check-env ## Start dev environment using production DB (no local postgres)
	@./tools/scripts/assign-dev-ports.sh $(ENV_FILE) prod-docker
	@echo "$(GREEN)Starting dev environment with production DB...$(NC)"
	@docker-compose -f docker-compose.yml -f docker-compose.prod-db.yml up -d --remove-orphans redis localstack scraper monolith worker
	@echo "$(GREEN)Ready. Attaching to logs (Ctrl+C detaches — containers keep running)...$(NC)"
	@docker-compose -f docker-compose.yml -f docker-compose.prod-db.yml logs -f

dev-web: check-env ## Start Next.js on configured DEV_WEB_PORT (Go proxies unmatched routes)
	@./tools/scripts/assign-dev-ports.sh $(ENV_FILE) web
	@set -a; . $(PWD)/$(ENV_FILE); set +a; \
		echo "$(GREEN)Starting Next.js on :$${DEV_WEB_PORT}...$(NC)"; \
		echo "$(YELLOW)Access the app at http://<tenant>.localhost:$${HTTP_PORT} (Go serves as entry point)$(NC)"; \
		cd frontend/apps/web && PORT=$${DEV_WEB_PORT} bun dev

dev-web-with-prod: check-env ## Start Next.js + CMS connected to production DB
	@./tools/scripts/assign-dev-ports.sh $(ENV_FILE) prod-web
	@set -a; . $(PWD)/$(ENV_FILE); set +a; \
		echo "$(GREEN)Starting Next.js on :$${DEV_WEB_PORT} (prod DB)...$(NC)"; \
		echo "$(YELLOW)Access the app at http://<tenant>.localhost:$${HTTP_PORT}$(NC)"; \
		cd frontend/apps/web && DB_PORT=5432 PORT=$${DEV_WEB_PORT} bun dev

down: ## Stop local dev environment
	@docker-compose down -v

logs: ## View logs (use: make logs service=monolith)
	@docker-compose logs -f $(service)

# ============================================================
# DATABASE
# ============================================================

cms-migrate: check-env ## Push Payload CMS schema and seed default content
	@echo "$(GREEN)Running CMS migrations + seed...$(NC)"
	@set -a; . $(ENV_FILE); set +a; \
		cd frontend/apps/web && bun run cms:migrate
	@echo "$(GREEN)✓ CMS migrate + seed complete$(NC)"

migrate: check-env ## Run database migrations from .env (use: make migrate cmd=up|down|version)
	@export $(shell grep -v '^#' $(ENV_FILE) | xargs) && \
	go run ./cmd/migrate \
		-db "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" \
		-cmd $(or $(cmd),up)

# ============================================================
# ARCHITECTURE
# ============================================================

# Set ARCH_WARN_ONLY=1 to emit violations as warnings (non-zero exit still emitted by script).
# After the one-week rollout window, flip this to 0 to enforce strict mode.
ARCH_WARN_ONLY ?= 1

check-arch: ## Check hexagonal architecture rules (includes test files); warn-only during rollout
	@ARCH_WARN_ONLY=$(ARCH_WARN_ONLY) ./tools/scripts/check-architecture.sh --include-tests

# ============================================================
# TESTS
# ============================================================

# Test DB lives in the same Postgres container as dev, just a different database.
# This isolates integration test data from the dev DB so failed/interrupted runs
# can't pollute jcsoftdev-inc, etc.
TEST_DB_NAME := pulzifi_test
# Tenant schema used by integration tests as a fixture.
# See modules/integration/infrastructure/persistence/destination_postgres_test.go
TEST_TENANT := jcsoftdev_inc

test: ## Run unit tests (no DB required)
	@go test ./...

test-integration: check-env ## Run integration tests against an isolated pulzifi_test DB
	@export $$(grep -v '^#' $(ENV_FILE) | xargs) && \
		TEST_DB_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$(TEST_DB_NAME)?sslmode=disable" && \
		echo "$(GREEN)Ensuring $(TEST_DB_NAME) database exists...$(NC)" && \
		( docker exec pulzifi-postgres psql -U $$DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$(TEST_DB_NAME)'" | grep -q 1 || \
			docker exec pulzifi-postgres psql -U $$DB_USER -d postgres -c "CREATE DATABASE $(TEST_DB_NAME)" ) && \
		echo "$(GREEN)Applying public migrations to $(TEST_DB_NAME)...$(NC)" && \
		go run ./cmd/migrate -db "$$TEST_DB_URL" -scope public -cmd up && \
		echo "$(GREEN)Provisioning tenant schema $(TEST_TENANT)...$(NC)" && \
		go run ./cmd/migrate -db "$$TEST_DB_URL" -scope tenant -tenant $(TEST_TENANT) -cmd up && \
		echo "$(GREEN)Running integration tests...$(NC)" && \
		DATABASE_URL="$$TEST_DB_URL" DB_NAME=$(TEST_DB_NAME) \
			go test -tags=integration -count=1 ./...

test-billing-integration: check-env ## Run billing integration tests only (webhook idempotency + plan assigner SQL)
	@export $$(grep -v '^#' $(ENV_FILE) | xargs) && \
		TEST_DB_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$(TEST_DB_NAME)?sslmode=disable" && \
		echo "$(GREEN)Ensuring $(TEST_DB_NAME) database exists...$(NC)" && \
		( docker exec pulzifi-postgres psql -U $$DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$(TEST_DB_NAME)'" | grep -q 1 || \
			docker exec pulzifi-postgres psql -U $$DB_USER -d postgres -c "CREATE DATABASE $(TEST_DB_NAME)" ) && \
		echo "$(GREEN)Applying public migrations to $(TEST_DB_NAME)...$(NC)" && \
		go run ./cmd/migrate -db "$$TEST_DB_URL" -scope public -cmd up && \
		echo "$(GREEN)Running billing integration tests...$(NC)" && \
		DATABASE_URL="$$TEST_DB_URL" DB_NAME=$(TEST_DB_NAME) \
			go test -tags=integration -count=1 -v \
				./modules/billing/infrastructure/persistence/postgres/... \
				./modules/billing/infrastructure/http/... \
				./cmd/wiring/billing/...

test-db-reset: check-env ## Drop and recreate pulzifi_test from scratch
	@export $$(grep -v '^#' $(ENV_FILE) | xargs) && \
		TEST_DB_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$(TEST_DB_NAME)?sslmode=disable" && \
		echo "$(YELLOW)Dropping $(TEST_DB_NAME)...$(NC)" && \
		docker exec pulzifi-postgres psql -U $$DB_USER -d postgres -c "DROP DATABASE IF EXISTS $(TEST_DB_NAME)" && \
		docker exec pulzifi-postgres psql -U $$DB_USER -d postgres -c "CREATE DATABASE $(TEST_DB_NAME)" && \
		go run ./cmd/migrate -db "$$TEST_DB_URL" -scope public -cmd up && \
		go run ./cmd/migrate -db "$$TEST_DB_URL" -scope tenant -tenant $(TEST_TENANT) -cmd up && \
		echo "$(GREEN)✓ $(TEST_DB_NAME) recreated and migrated$(NC)"

# ============================================================
# BUILD
# ============================================================

build: check-env ## Build API binary locally
	@mkdir -p ./bin
	@go build -o ./bin/api ./cmd/server/
	@echo "$(GREEN)✓ Binary built at ./bin/api$(NC)"

swagger: ## Regenerate Swagger docs
	@swag init -g cmd/server/main.go --output docs
	@echo "$(GREEN)✓ Swagger docs generated$(NC)"

# ============================================================
# CLEANUP
# ============================================================

clean: ## Stop all containers and prune Docker resources
	@docker-compose down -v 2>/dev/null || true
	@docker system prune -f --volumes 2>/dev/null || true
	@echo "$(GREEN)✓ Cleanup completed$(NC)"

# ============================================================
# OPS
# ============================================================

backup: check-env ## Run a manual PostgreSQL backup (pg_dump → gzip → MinIO/S3)
	@export $$(grep -v '^#' $(ENV_FILE) | xargs) && ./tools/scripts/pg-backup.sh
