# Makefile for the Language Learning Player Backend

# --- Variables ---
# Go related variables
BINARY_NAME=language-player-api
CMD_PATH=./cmd/api
OUTPUT_DIR=./build
GO_BUILD_FLAGS=-ldflags='-w -s' -trimpath
GO_TEST_FLAGS=./... -coverprofile=coverage.out
# DSN for local database operations (Can be overridden by environment variable)
# Example: export DATABASE_URL="postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable"
DATABASE_URL ?= postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable
# Migrate CLI path relative to project root
MIGRATIONS_PATH=migrations
# Swag CLI variables (if using swaggo/swag)
SWAG_ENTRY_POINT=${CMD_PATH}/main.go
SWAG_OUTPUT_DIR=./docs
# Docker image name
DOCKER_IMAGE_NAME ?= your-dockerhub-username/language-player-api
DOCKER_IMAGE_TAG ?= latest

# --- Go Tools Installation ---
# Define paths for Go tools
GOPATH := $(shell go env GOPATH)
GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
# Ensure GOBIN is in PATH for Make to find the tools
export PATH := $(GOBIN):$(PATH)

# Tool binaries
MIGRATE := $(GOBIN)/migrate
SQLC := $(GOBIN)/sqlc
SWAG := $(GOBIN)/swag
GOLANGCILINT := $(GOBIN)/golangci-lint
GOVULNCHECK := $(GOBIN)/govulncheck

.PHONY: tools install-migrate install-sqlc install-swag install-lint install-vulncheck

# Target to install all necessary Go tools
tools: install-migrate install-sqlc install-swag install-lint install-vulncheck

# Check if migrate is installed, if not, install it
install-migrate:
	@if ! command -v migrate &> /dev/null; then \
		echo ">>> Installing migrate CLI..."; \
		go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		echo ">>> migrate installed."; \
	else \
		echo ">>> migrate is already installed."; \
	fi

# Check if sqlc is installed, if not, install it (Optional, if using sqlc)
install-sqlc:
	@if ! command -v sqlc &> /dev/null; then \
		echo ">>> Installing sqlc CLI..."; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
		echo ">>> sqlc installed."; \
	else \
		echo ">>> sqlc is already installed."; \
	fi

# Check if swag is installed, if not, install it (Optional, if using swaggo/swag)
install-swag:
	@if ! command -v swag &> /dev/null; then \
		echo ">>> Installing swag CLI..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		echo ">>> swag installed."; \
	else \
		echo ">>> swag is already installed."; \
	fi

# Check if golangci-lint is installed, if not, install it
install-lint:
	@if ! command -v golangci-lint &> /dev/null; then \
		echo ">>> Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo ">>> golangci-lint installed."; \
	else \
		echo ">>> golangci-lint is already installed."; \
	fi

# Check if govulncheck is installed, if not, install it
install-vulncheck:
	@if ! command -v govulncheck &> /dev/null; then \
		echo ">>> Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		echo ">>> govulncheck installed."; \
	else \
		echo ">>> govulncheck is already installed."; \
	fi

# --- Build ---
.PHONY: build clean

# Build the Go binary
build: clean tools
	@echo ">>> Building binary..."
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo ">>> Binary built at $(OUTPUT_DIR)/$(BINARY_NAME)"

# Remove build artifacts
clean:
	@echo ">>> Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f coverage.out

# --- Run ---
.PHONY: run

# Run the application locally (requires dependencies like DB running)
# Uses local configuration (e.g., config.development.yaml)
# Ensure required env vars (like secrets) are set or use tools like direnv
run: tools
	@echo ">>> Running application locally (using go run)..."
	@APP_ENV=development go run $(CMD_PATH)/main.go

# --- Database Migrations ---
.PHONY: migrate-create migrate-up migrate-down migrate-force

# Create a new migration file
# Usage: make migrate-create name=your_migration_name
migrate-create: tools
	@echo ">>> Creating migration file: $(name)..."
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo ">>> Migration file created."

# Apply all up migrations
migrate-up: tools
	@echo ">>> Applying database migrations..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) up
	@echo ">>> Migrations applied."

# Roll back the last migration
migrate-down: tools
	@echo ">>> Reverting last database migration..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) down 1
	@echo ">>> Last migration reverted."

# Force set migration version (Use with caution!)
# Usage: make migrate-force version=YYYYMMDDHHMMSS
migrate-force: tools
	@echo ">>> Forcing migration version to $(version)..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) force $(version)
	@echo ">>> Migration version forced."


# --- Code Generation ---
.PHONY: generate generate-sqlc generate-swag

# Target to run all generators
generate: generate-sqlc generate-swag

# Generate Go code from SQL queries using sqlc (Optional)
generate-sqlc: tools
	@echo ">>> Generating Go code from SQL queries using sqlc..."
	@$(SQLC) generate
	@echo ">>> sqlc generation complete."

# Generate OpenAPI docs using swag (Optional)
generate-swag: tools
	@echo ">>> Generating OpenAPI docs using swag..."
	@$(SWAG) init -g $(SWAG_ENTRY_POINT) --output $(SWAG_OUTPUT_DIR)
	@echo ">>> OpenAPI docs generated in $(SWAG_OUTPUT_DIR)."


# --- Testing ---
.PHONY: test test-unit test-integration test-cover

# Run all tests (unit + integration, requires Docker for integration)
test: tools
	@echo ">>> Running all tests (unit + integration)..."
	@go test $(GO_TEST_FLAGS)
	@echo ">>> Tests complete. Coverage report generated at coverage.out"

# Run only unit tests (usually tests not ending in _integration_test.go or in specific packages)
# This might require better test organization or build tags later.
# For now, a simple placeholder assuming non-integration tests are faster.
test-unit: tools
	@echo ">>> Running unit tests (placeholder)..."
	@go test $(GO_TEST_FLAGS) -short # -short flag might skip long-running tests if tests use it

# Run only integration tests (requires Docker)
# Assuming integration tests are marked with _integration_test.go suffix or specific build tag
# Requires proper test file naming. Example: go test ./... -tags=integration
test-integration: tools
	@echo ">>> Running integration tests (requires Docker)..."
	@go test ./internal/adapter/repository/postgres/... -v # Run tests specifically in the repo package
	# Or use build tags: @go test ./... -tags=integration -v

# Show test coverage in browser
test-cover: test
	@echo ">>> Opening test coverage report..."
	@go tool cover -html=coverage.out

# --- Linting & Formatting ---
.PHONY: lint fmt check-vuln

# Run golangci-lint
lint: tools
	@echo ">>> Running linter..."
	@$(GOLANGCILINT) run ./...

# Format Go code
fmt:
	@echo ">>> Formatting Go code..."
	@go fmt ./...
	@goimports -w . # Optional: run goimports if installed and preferred

# Check for known vulnerabilities
check-vuln: tools
	@echo ">>> Checking for vulnerabilities..."
	@$(GOVULNCHECK) ./...


# --- Docker ---
.PHONY: docker-build docker-run docker-stop docker-push

# Build Docker image
docker-build:
	@echo ">>> Building Docker image [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo ">>> Docker image built."

# Run Docker container locally (using env vars from .env file if present)
docker-run: docker-build
	@echo ">>> Running Docker container [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	# Load .env file if it exists (requires .env file with variable assignments)
	@$(if $(wildcard .env), $(eval include .env) $(eval export))
	@docker run -d --name $(BINARY_NAME) \
		-p 8080:8080 \
		--env-file .env \
		$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo ">>> Container started. Use 'make docker-stop' to stop."

# Stop and remove the running container
docker-stop:
	@echo ">>> Stopping and removing Docker container [$(BINARY_NAME)]..."
	@docker stop $(BINARY_NAME) || true
	@docker rm $(BINARY_NAME) || true
	@echo ">>> Container stopped and removed."

# Push Docker image to registry (requires docker login)
docker-push:
	@echo ">>> Pushing Docker image [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo ">>> Image pushed."

# --- Help ---
.PHONY: help

# Show help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build             Build the Go binary for linux/amd64"
	@echo "  clean             Remove build artifacts"
	@echo "  run               Run the application locally using go run (needs dependencies)"
	@echo "  tools             Install necessary Go CLI tools (migrate, sqlc, swag, lint, vulncheck)"
	@echo "  migrate-create name=<name> Create a new migration file"
	@echo "  migrate-up        Apply database migrations"
	@echo "  migrate-down      Revert the last database migration"
	@echo "  migrate-force version=<ver> Force migration version (use with caution)"
	@echo "  generate          Run all code generators (sqlc, swag)"
	@echo "  generate-sqlc     Generate Go code from SQL using sqlc"
	@echo "  generate-swag     Generate OpenAPI docs using swag"
	@echo "  test              Run all tests (unit + integration) and generate coverage"
	@echo "  test-unit         Run unit tests (placeholder)"
	@echo "  test-integration  Run integration tests (requires Docker)"
	@echo "  test-cover        Show test coverage report in browser"
	@echo "  lint              Run golangci-lint"
	@echo "  fmt               Format Go code using go fmt and goimports"
	@echo "  check-vuln        Check for known vulnerabilities using govulncheck"
	@echo "  docker-build      Build the Docker image"
	@echo "  docker-run        Build and run the Docker container locally (uses .env file)"
	@echo "  docker-stop       Stop and remove the running Docker container"
	@echo "  docker-push       Push the Docker image to registry (requires login)"
	@echo "  help              Show this help message"

# Default target
.DEFAULT_GOAL := help