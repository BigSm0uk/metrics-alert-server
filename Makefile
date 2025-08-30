# Makefile for macOS/Linux Go project

# Go commands and tools
GO := go
GOFMT := go fmt
GOPATH := $(shell go env GOPATH)
GOIMPORTS := $(GOPATH)/bin/goimports
GOLANGCI_LINT := $(GOPATH)/bin/golangci-lint

# Directories 
SERVER_DIR := cmd/server
AGENT_DIR := cmd/agent
DEPLOY_DIR := deploy
MIGRATIONS_DIR := migrations
DOCS_DIR := docs
# Binary names
SERVER_BIN := $(SERVER_DIR)/server
AGENT_BIN := $(AGENT_DIR)/agent

# Database configuration
DB_HOST := localhost
DB_PORT := 5432
DB_USER := metrics_user
DB_PASSWORD := metrics_password
DB_NAME := metrics_dev
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: all fmt lint vet test build clean install-tools help docker-up docker-down migrate-up migrate-down migrate-status migrate-create generate-docs

# Default target
all: build

# Format code
fmt:
	@echo "[+] Formatting Go code..."
	$(GO) fmt ./...
	@if [ -f "$(GOIMPORTS)" ]; then \
		echo "[+] Running goimports..." && \
		"$(GOIMPORTS)" -w . ; \
	else \
		echo "[!] goimports not found. Run 'make install-tools' first" ; \
	fi

# Run linter
lint:
	@if [ -f "$(GOLANGCI_LINT)" ]; then \
		echo "[+] Running linter..." && \
		"$(GOLANGCI_LINT)" run ./... ; \
	else \
		echo "[!] golangci-lint not found. Run 'make install-tools' first" ; \
	fi

# Run tests
test:
	@echo "[+] Running tests..."
	$(GO) test -v ./...

# Run go vet
vet:
	@echo "[+] Running go vet..."
	$(GO) vet ./...

# Build server binary
build-server:
	@echo "[+] Building server..."
	$(GO) build -o $(SERVER_BIN) ./cmd/server

# Build agent binary
build-agent:
	@echo "[+] Building agent..."
	$(GO) build -o $(AGENT_BIN) ./cmd/agent

# Build all binaries
build: build-server build-agent
	@echo "[+] Build complete"

# Run server
run-server: build-server
	@echo "[+] Starting server..."
	./$(SERVER_BIN) --config=config/config.local.yaml

# Run agent
run-agent: build-agent
	@echo "[+] Starting agent..."
	./$(AGENT_BIN)

# Clean build artifacts
clean:
	@echo "[+] Cleaning build artifacts..."
	@rm -f $(SERVER_BIN)
	@rm -f $(AGENT_BIN)
	$(GO) clean -cache

# Install required tools
install-tools:
	@echo "[+] Installing required tools..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "[+] Tools installation complete."
	@echo "[+] Please ensure that $(GOPATH)/bin is in your PATH"
	@echo "[+] Current GOPATH: $(GOPATH)"

# Docker Compose commands
docker-up:
	@echo "[+] Starting development environment..."
	@cd $(DEPLOY_DIR) && docker-compose -f docker-compose.dev.yaml up -d
	@echo "[+] Development environment started"

docker-down:
	@echo "[+] Stopping development environment..."
	@cd $(DEPLOY_DIR) && docker-compose -f docker-compose.dev.yaml down
	@echo "[+] Development environment stopped"

# Database migrations with Goose
migrate-up:
	@echo "[+] Running database migrations up..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	@echo "[+] Rolling back database migrations..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	@echo "[+] Checking migration status..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "[+] Creating new migration: $(NAME)..."
	@goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

migrate-reset:
	@echo "[+] Resetting database (DOWN all + UP all)..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" reset

generate-docs:
	@echo "[+] Generating docs..."
	@npx @redocly/cli build-docs api/openapi.yaml -o $(DOCS_DIR)/redoc.html      
	@echo "[+] Docs generated"
# Show help
help:
	@echo "Available commands:"
	@echo "  make fmt           - Format Go code"
	@echo "  make lint          - Run golangci-lint"
	@echo "  make vet          - Run go vet"
	@echo "  make test         - Run tests with race detection"
	@echo "  make build        - Build all binaries"
	@echo "  make build-server - Build server binary"
	@echo "  make build-agent  - Build agent binary"
	@echo "  make run-server   - Build and run server"
	@echo "  make run-agent    - Build and run agent"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install-tools - Install required development tools"
	@echo "  make docker-up    - Start development environment (PostgreSQL)"
	@echo "  make docker-down  - Stop development environment"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make migrate-status - Check migration status"
	@echo "  make migrate-create NAME=name - Create new migration"
	@echo "  make migrate-reset - Reset database (down all + up all)"
	@echo "  make generate-docs - Generate docs"