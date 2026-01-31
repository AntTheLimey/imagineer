.PHONY: help up down reset logs shell psql build test test-server test-client test-all test-db lint coverage migrate migrate-status backup restore status client-dev client-build

# Default target
help:
	@echo "Imagineer - TTRPG Campaign Intelligence Platform"
	@echo ""
	@echo "Usage:"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make status         - Show service status"
	@echo "  make reset          - Stop services and remove data (DESTRUCTIVE)"
	@echo "  make logs           - Follow service logs"
	@echo "  make shell          - Open shell in postgres container"
	@echo "  make psql           - Open psql session"
	@echo "  make build          - Build Go binaries"
	@echo "  make test           - Run all tests"
	@echo "  make test-server    - Run Go server tests"
	@echo "  make test-client    - Run React client tests"
	@echo "  make test-all       - Run all tests with coverage and linting"
	@echo "  make test-db        - Test database extensions are installed"
	@echo "  make lint           - Run linters"
	@echo "  make coverage       - Run tests with coverage"
	@echo "  make migrate        - Run pending database migrations"
	@echo "  make migrate-status - Show migration status"
	@echo "  make backup         - Create database backup"
	@echo "  make restore        - Restore from backup"
	@echo "  make client-dev     - Start React dev server"
	@echo "  make client-build   - Build React client"

# Docker commands
up:
	@echo "Starting Imagineer services..."
	@cp -n .env.example .env 2>/dev/null || true
	docker-compose up -d
	@echo "Services started. MCP Server: http://localhost:8080"

down:
	@echo "Stopping Imagineer services..."
	docker-compose down

reset:
	@echo "WARNING: This will delete all data!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	docker-compose down -v
	rm -rf ./data/postgres/*
	@echo "Reset complete."

logs:
	docker-compose logs -f

shell:
	docker exec -it imagineer-postgres /bin/bash

psql:
	docker exec -it imagineer-postgres psql -U imagineer -d imagineer

# Build commands
build:
	go build -o bin/imagineer ./cmd/server
	go build -o bin/imagineer-cli ./cmd/cli

# Test commands
test: test-server test-client

test-server:
	@echo "=== Running Go tests ==="
	go test -v ./...

test-client:
	@echo "=== Running React tests ==="
	@if [ -d "client" ] && [ -f "client/package.json" ]; then \
		cd client && npm test; \
	else \
		echo "Client not set up yet"; \
	fi

test-all: lint coverage
	@echo "=== All tests passed ==="

lint:
	@echo "=== Running linters ==="
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@GOBIN_PATH=$$(go env GOPATH)/bin; \
	if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	elif [ -x "$$GOBIN_PATH/golangci-lint" ]; then \
		$$GOBIN_PATH/golangci-lint run ./...; \
	else \
		echo "ERROR: golangci-lint not found. Add $$GOBIN_PATH to your PATH or install manually."; \
		exit 1; \
	fi
	@gofmt -l . | (! grep .) || (echo "Some files need gofmt" && exit 1)
	@if [ -d "client" ] && [ -f "client/package.json" ]; then \
		cd client && npm run lint; \
	fi

coverage:
	@echo "=== Running tests with coverage ==="
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@if [ -d "client" ] && [ -f "client/package.json" ]; then \
		cd client && npm run test:coverage; \
	fi

test-db:
	@echo "=== Testing database extensions ==="
	@./scripts/test-db-extensions.sh

# Client commands
client-dev:
	@if [ -d "client" ]; then \
		cd client && npm run dev; \
	else \
		echo "Client not set up. Run setup first."; \
	fi

client-build:
	@if [ -d "client" ]; then \
		cd client && npm run build; \
	else \
		echo "Client not set up. Run setup first."; \
	fi

# Database commands
migrate:
	@./scripts/migrate.sh

migrate-status:
	@docker exec imagineer-postgres psql -U imagineer -d imagineer -c "SELECT version, applied_at FROM schema_migrations ORDER BY version;"

backup:
	@./scripts/backup.sh

restore:
	@./scripts/restore.sh

# Status commands
status:
	@echo "=== Imagineer Service Status ==="
	@docker ps --filter "name=imagineer" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "=== Database Tables ==="
	@docker exec imagineer-postgres psql -U imagineer -d imagineer -c "\dt" 2>/dev/null || echo "Database not available"
	@echo ""
	@echo "=== Game Systems ==="
	@docker exec imagineer-postgres psql -U imagineer -d imagineer -t -c "SELECT name FROM game_systems;" 2>/dev/null || echo "Not available"
