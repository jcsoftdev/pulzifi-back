# ============================================================
# Makefile for Pulzifi Backend
# ============================================================

.PHONY: help dev dev-web down logs build swagger clean migrate

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
	@echo "  $(YELLOW)make dev$(NC)      - Start local dev environment (postgres + extractor + hot reload)"
	@echo "  $(YELLOW)make dev-web$(NC)  - Start Caddy proxy + Next.js (direct API calls, no proxy overhead)"
	@echo "  $(YELLOW)make down$(NC)     - Stop local dev environment"
	@echo "  $(YELLOW)make logs$(NC)     - View logs (use: make logs service=monolith)"
	@echo ""
	@echo "$(GREEN)DATABASE:$(NC)"
	@echo "  $(YELLOW)make migrate$(NC)  - Run migrations from .env (make migrate cmd=up|down|version)"
	@echo ""
	@echo "$(GREEN)BUILD:$(NC)"
	@echo "  $(YELLOW)make build$(NC)    - Build API binary locally"
	@echo "  $(YELLOW)make swagger$(NC)  - Regenerate Swagger docs"
	@echo "  $(YELLOW)make clean$(NC)    - Stop containers and prune Docker resources"

check-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(YELLOW)⚠️  .env file not found$(NC)"; \
		cp .env.example $(ENV_FILE) 2>/dev/null && echo "$(GREEN)✓ .env created from .env.example — please review it$(NC)"; \
	fi

# ============================================================
# LOCAL DEVELOPMENT
# ============================================================

dev: check-env ## Start local dev (postgres + extractor + API + worker with hot reload)
	@echo "$(GREEN)Starting local dev environment...$(NC)"
	@docker-compose -f docker-compose.monolith.yml up

dev-web: ## Start Caddy proxy (:3000) + Next.js (:3001) for local frontend dev
	@command -v caddy >/dev/null 2>&1 || { echo "$(YELLOW)caddy not found — install with: brew install caddy$(NC)"; exit 1; }
	@echo "$(GREEN)Starting Caddy on :3000 and Next.js on :3001...$(NC)"
	@echo "$(YELLOW)Access the app at http://<tenant>.localhost:3000$(NC)"
	@trap 'kill 0' INT; \
		caddy run --config Caddyfile & \
		(cd frontend/apps/web && PORT=3001 bun dev) & \
		wait

down: ## Stop local dev environment
	@docker-compose -f docker-compose.monolith.yml down -v

logs: ## View logs (use: make logs service=monolith)
	@docker-compose -f docker-compose.monolith.yml logs -f $(service)

# ============================================================
# DATABASE
# ============================================================

migrate: check-env ## Run database migrations from .env (use: make migrate cmd=up|down|version)
	@export $(shell grep -v '^#' $(ENV_FILE) | xargs) && \
	go run ./cmd/migrate \
		-db "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" \
		-cmd $(or $(cmd),up)

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
	@docker-compose -f docker-compose.monolith.yml down -v 2>/dev/null || true
	@docker system prune -f --volumes 2>/dev/null || true
	@echo "$(GREEN)✓ Cleanup completed$(NC)"
