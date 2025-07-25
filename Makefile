# Makefile for Windows 11 Go project

# Go commands and tools
GO := go
GOFMT := go fmt
GOPATH := $(shell go env GOPATH)
GOIMPORTS := $(GOPATH)\bin\goimports.exe
GOLANGCI_LINT := $(GOPATH)\bin\golangci-lint.exe

# Directories 
SERVER_DIR := cmd\server
AGENT_DIR := cmd\agent

# Binary names
SERVER_BIN := $(SERVER_DIR)\server.exe
AGENT_BIN := $(AGENT_DIR)\agent.exe

.PHONY: all fmt lint vet test build clean install-tools help

# Default target
all: build

# Format code
fmt:
	@echo [+] Formatting Go code...
	$(GO) fmt ./...
	@if exist "$(GOIMPORTS)" ( \
		echo [+] Running goimports... && \
		"$(GOIMPORTS)" -w . \
	) else ( \
		echo [!] goimports not found. Run 'make install-tools' first \
	)

# Run linter
lint:
	@if exist "$(GOLANGCI_LINT)" ( \
		echo [+] Running linter... && \
		"$(GOLANGCI_LINT)" run ./... \
	) else ( \
		echo [!] golangci-lint not found. Run 'make install-tools' first \
	)

# Run go vet
vet:
	@echo [+] Running go vet...
	$(GO) vet ./...

# Run tests
test:
	@echo [+] Running tests...
	$(GO) test -v -race ./...

# Build server binary
build-server:
	@echo [+] Building server...
	$(GO) build -o $(SERVER_BIN) .\cmd\server

# Build agent binary
build-agent:
	@echo [+] Building agent...
	$(GO) build -o $(AGENT_BIN) .\cmd\agent

# Build all binaries
build: build-server build-agent
	@echo [+] Build complete

# Run server
run-server: build-server
	@echo [+] Starting server...
	$(SERVER_BIN)

# Run agent
run-agent: build-agent
	@echo [+] Starting agent...
	$(AGENT_BIN)

# Clean build artifacts
clean:
	@echo [+] Cleaning build artifacts...
	@if exist "$(SERVER_BIN)" del /F /Q "$(SERVER_BIN)"
	@if exist "$(AGENT_BIN)" del /F /Q "$(AGENT_BIN)"
	$(GO) clean -cache

# Install required tools
install-tools:
	@echo [+] Installing required tools...
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo [+] Tools installation complete.
	@echo [+] Please ensure that $(GOPATH)\bin is in your PATH
	@echo [+] Current GOPATH: $(GOPATH)

# Show help
help:
	@echo Available commands:
	@echo   make fmt           - Format Go code
	@echo   make lint          - Run golangci-lint
	@echo   make vet          - Run go vet
	@echo   make test         - Run tests with race detection
	@echo   make build        - Build all binaries
	@echo   make build-server - Build server binary
	@echo   make build-agent  - Build agent binary
	@echo   make run-server   - Build and run server
	@echo   make run-agent    - Build and run agent
	@echo   make clean        - Remove build artifacts
	@echo   make install-tools - Install required development tools