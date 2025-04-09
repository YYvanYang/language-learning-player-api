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
# PostgreSQL Docker settings
PG_CONTAINER_NAME ?= language-learner-postgres
PG_USER ?= user
PG_PASSWORD ?= password
PG_DB ?= language_learner_db
PG_PORT ?= 5432
PG_VERSION ?= 16
PG_READY_TIMEOUT ?= 30 # Seconds to wait for PostgreSQL to be ready
# MinIO Docker settings
MINIO_CONTAINER_NAME ?= language-learner-minio
MINIO_ROOT_USER ?= minioadmin
MINIO_ROOT_PASSWORD ?= minioadmin
MINIO_API_PORT ?= 9000
MINIO_CONSOLE_PORT ?= 9001
MINIO_BUCKET_NAME ?= language-audio # Ensure this matches config.development.yaml
MINIO_READY_TIMEOUT ?= 30 # Seconds to wait for MinIO to be ready
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
# Explicitly use full path to go for shell commands to avoid PATH issues during variable parsing
GOPATH := $(shell /usr/local/go/bin/go env GOPATH)
# Define GOBIN, prioritize go env GOBIN, fallback to GOPATH/bin, then $HOME/go/bin
GOBIN ?= $(firstword $(shell /usr/local/go/bin/go env GOBIN) $(GOPATH)/bin $(HOME)/go/bin)

# Tool binaries
MIGRATE := $(GOBIN)/migrate
# SQLC variable removed
# SWAG := $(shell go env GOPATH)/bin/swag # Temporarily comment out the dynamic path
# SWAG := /home/yvan/go/bin/swag # Temporarily hardcode the path - CHANGE IF YOURS IS DIFFERENT!
SWAG := $(GOBIN)/swag
GOLANGCILINT := $(GOBIN)/golangci-lint
GOVULNCHECK := $(GOBIN)/govulncheck

.PHONY: tools install-migrate install-swag install-lint install-vulncheck
# install-sqlc removed from .PHONY

# Target to install all necessary Go tools
tools: install-migrate install-swag install-lint install-vulncheck # install-sqlc removed

# Check if migrate is installed, if not, install it
install-migrate:
	@if ! command -v migrate &> /dev/null; then \
		echo ">>> Installing migrate CLI..."; \
		if go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; then \
			echo ">>> migrate installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install migrate. Please check network connectivity and Go proxy settings."; \
			exit 1; \
		fi; \
	else \
		echo ">>> migrate is already installed."; \
	fi

# install-sqlc target removed

# Check if swag is installed, if not, install it (Optional, if using swaggo/swag)
install-swag:
	@if ! command -v swag &> /dev/null; then \
		echo ">>> Installing swag CLI..."; \
		if go install github.com/swaggo/swag/cmd/swag@latest; then \
			echo ">>> swag installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install swag. Please check network connectivity and Go proxy settings."; \
			exit 1; \
		fi; \
	else \
		echo ">>> swag is already installed."; \
	fi

# Check if golangci-lint is installed, if not, install it
install-lint:
	@if ! command -v golangci-lint &> /dev/null; then \
		echo ">>> Installing golangci-lint..."; \
		if go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; then \
			echo ">>> golangci-lint installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install golangci-lint. Please check network connectivity and Go proxy settings."; \
			exit 1; \
		fi; \
	else \
		echo ">>> golangci-lint is already installed."; \
	fi

# Check if govulncheck is installed, if not, install it
install-vulncheck:
	@if ! command -v govulncheck &> /dev/null; then \
		echo ">>> Installing govulncheck..."; \
		if go install golang.org/x/vuln/cmd/govulncheck@latest; then \
			echo ">>> govulncheck installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install govulncheck. Please check network connectivity and Go proxy settings."; \
			exit 1; \
		fi; \
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
.PHONY: migrate-create migrate-up migrate-down migrate-force check-db-url

# Internal target to check if DATABASE_URL is set
check-db-url:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo ">>> ERROR: DATABASE_URL environment variable is not set."; \
		echo ">>> Please set it before running migrations, e.g.: export DATABASE_URL='postgresql://user:password@host:port/db?sslmode=disable'"; \
		exit 1; \
	fi

# Create a new migration file
# Usage: make migrate-create name=your_migration_name
migrate-create: tools
	@echo ">>> Creating migration file: $(name)..."
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo ">>> Migration file created."

# Apply all up migrations
migrate-up: tools check-db-url
	@echo ">>> Applying database migrations..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) up
	@echo ">>> Migrations applied."

# Roll back the last migration
migrate-down: tools check-db-url
	@echo ">>> Reverting last database migration..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) down 1
	@echo ">>> Last migration reverted."

# Force set migration version (Use with caution!)
# Usage: make migrate-force version=YYYYMMDDHHMMSS
migrate-force: tools check-db-url
	@echo ">>> Forcing migration version to $(version)..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) force $(version)
	@echo ">>> Migration version forced."


# --- Code Generation ---
.PHONY: generate generate-swag swagger # generate-sqlc removed

# Target to run all generators
generate: generate-swag # generate-sqlc removed

# generate-sqlc target removed

# Generate OpenAPI docs using swag (Optional)
generate-swag: tools
	@echo ">>> Generating OpenAPI docs using swag..."
	@echo ">>> Using swag command: $(SWAG)"
	@$(SWAG) init -g $(SWAG_ENTRY_POINT) --output $(SWAG_OUTPUT_DIR)
	@echo ">>> OpenAPI docs generated in $(SWAG_OUTPUT_DIR)."

# Alias for generating OpenAPI docs
swagger: generate-swag

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
# Update .PHONY to include new MinIO targets
.PHONY: docker-build docker-run docker-stop docker-push docker-postgres-run docker-postgres-stop docker-minio-run docker-minio-stop

# Build Docker image
docker-build:
	@echo ">>> Building Docker image [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo ">>> Docker image built."

# Run Docker container locally (using env vars from .env file if present)
docker-run: docker-build
	@echo ">>> Running Docker container [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@# Docker will read variables directly from .env file using --env-file
	@# Ensure .env file exists if you rely on it, or pass env vars directly with -e
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
	@echo ">>> Note: Ensure DOCKER_IMAGE_NAME variable is set correctly for your registry."
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo ">>> Image pushed."

# --- PostgreSQL Docker ---
# Run PostgreSQL in Docker container (Simplified)
docker-postgres-run:
	@echo ">>> Ensuring PostgreSQL container [$(PG_CONTAINER_NAME)] is running..."
	@# Stop and remove existing container if it exists
	@docker stop $(PG_CONTAINER_NAME) > /dev/null 2>&1 || true
	@docker rm $(PG_CONTAINER_NAME) > /dev/null 2>&1 || true
	@echo ">>> Starting PostgreSQL container..."
	@docker run --name $(PG_CONTAINER_NAME) \
		-e POSTGRES_USER=$(PG_USER) \
		-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
		-e POSTGRES_DB=$(PG_DB) \
		-p $(PG_PORT):5432 \
		-d postgres:$(PG_VERSION)-alpine > /dev/null
	@echo ">>> Waiting for PostgreSQL to be ready (max $(PG_READY_TIMEOUT)s)..."
	@timeout=$(PG_READY_TIMEOUT); \
	while ! docker exec $(PG_CONTAINER_NAME) pg_isready -U $(PG_USER) -d $(PG_DB) -q; do \
		timeout=$$((timeout-1)); \
		if [ $$timeout -eq 0 ]; then \
			echo ">>> ERROR: PostgreSQL did not become ready in time."; \
			docker logs $(PG_CONTAINER_NAME); \
			exit 1; \
		fi; \
		sleep 1; \
	done
	@echo ">>> PostgreSQL container [$(PG_CONTAINER_NAME)] started successfully."
	@echo ">>> Connection string: $(DATABASE_URL)"


# Stop and remove PostgreSQL container
docker-postgres-stop:
	@echo ">>> Stopping and removing PostgreSQL container [$(PG_CONTAINER_NAME)]..."
	@docker stop $(PG_CONTAINER_NAME) || true
	@docker rm $(PG_CONTAINER_NAME) || true
	@echo ">>> PostgreSQL container stopped and removed."

# --- MinIO Docker ---
# Run MinIO in Docker container
docker-minio-run:
	@echo ">>> Ensuring MinIO container [$(MINIO_CONTAINER_NAME)] is running..."
	@# Stop and remove existing container if it exists
	@docker stop $(MINIO_CONTAINER_NAME) > /dev/null 2>&1 || true
	@docker rm $(MINIO_CONTAINER_NAME) > /dev/null 2>&1 || true
	@echo ">>> Starting MinIO container..."
	@docker run --name $(MINIO_CONTAINER_NAME) \
		-p $(MINIO_API_PORT):9000 \
		-p $(MINIO_CONSOLE_PORT):9001 \
		-e MINIO_ROOT_USER=$(MINIO_ROOT_USER) \
		-e MINIO_ROOT_PASSWORD=$(MINIO_ROOT_PASSWORD) \
		-d minio/minio server /data --console-address ":9001" > /dev/null
	@echo ">>> Waiting for MinIO to be ready (max $(MINIO_READY_TIMEOUT)s)..."
	@timeout=$(MINIO_READY_TIMEOUT); \
	# MODIFIED: Use curl --fail to check for 2xx status code instead of grepping body
	@until curl -s --max-time 1 --output /dev/null --fail "http://localhost:$(MINIO_API_PORT)/minio/health/live"; do \
		timeout=$$((timeout-1)); \
		if [ $$timeout -eq 0 ]; then \
			echo ">>> ERROR: MinIO did not become ready in time."; \
			docker logs $(MINIO_CONTAINER_NAME); \
			exit 1; \
		fi; \
		sleep 1; \
	done
	@echo ">>> MinIO container [$(MINIO_CONTAINER_NAME)] started successfully."
	@echo ">>> MinIO API: http://localhost:$(MINIO_API_PORT)"
	@echo ">>> MinIO Console: http://localhost:$(MINIO_CONSOLE_PORT) (Login: $(MINIO_ROOT_USER)/$(MINIO_ROOT_PASSWORD))"
	@# Automatically create the bucket if it doesn't exist
	@echo ">>> Ensuring bucket '$(MINIO_BUCKET_NAME)' exists..."
	@sleep 5 # Give MinIO extra time before creating bucket
	@docker exec $(MINIO_CONTAINER_NAME) mc alias set local http://localhost:9000 $(MINIO_ROOT_USER) $(MINIO_ROOT_PASSWORD) > /dev/null || true
	@docker exec $(MINIO_CONTAINER_NAME) mc mb local/$(MINIO_BUCKET_NAME) > /dev/null || echo ">>> Bucket '$(MINIO_BUCKET_NAME)' likely already exists."


# Stop and remove MinIO container
docker-minio-stop:
	@echo ">>> Stopping and removing MinIO container [$(MINIO_CONTAINER_NAME)]..."
	@docker stop $(MINIO_CONTAINER_NAME) || true
	@docker rm $(MINIO_CONTAINER_NAME) || true
	@echo ">>> MinIO container stopped and removed."

# --- Help ---
.PHONY: help

# Show help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Local Development:"
	@echo "  run               Run the application locally (requires dependencies: use 'make deps-run')"
	@echo "  deps-run          Start local PostgreSQL and MinIO containers"
	@echo "  deps-stop         Stop local PostgreSQL and MinIO containers"
	@echo "  tools             Install necessary Go CLI tools (migrate, swag, lint, vulncheck)" # Removed sqlc
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-create name=<name> Create a new migration file"
	@echo "  migrate-up        Apply database migrations (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-down      Revert the last database migration (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-force version=<ver> Force migration version (requires DB running and DATABASE_URL set/exported)"
	@echo ""
	@echo "Code Generation & Formatting:"
	@echo "  generate          Run all code generators (swag)" # Removed sqlc
	# Removed generate-sqlc help line
	@echo "  generate-swag     Generate OpenAPI docs using swag"
	@echo "  fmt               Format Go code using go fmt and goimports"
	@echo ""
	@echo "Testing & Linting:"
	@echo "  test              Run all tests and generate coverage"
	@echo "  test-unit         Run unit tests (placeholder)"
	@echo "  test-integration  Run integration tests (requires Docker)"
	@echo "  test-cover        Show test coverage report in browser"
	@echo "  lint              Run golangci-lint"
	@echo "  check-vuln        Check for known vulnerabilities"
	@echo ""
	@echo "Docker - Application:"
	@echo "  docker-build      Build the application Docker image"
	@echo "  docker-run        Build and run the application container locally (uses .env file)"
	@echo "  docker-stop       Stop and remove the running application container"
	@echo "  docker-push       Push the application Docker image to registry (Note: Customize DOCKER_IMAGE_NAME)"
	@echo ""
	@echo "Docker - Dependencies:"
	@echo "  docker-postgres-run Start PostgreSQL in Docker container"
	@echo "  docker-postgres-stop Stop and remove PostgreSQL Docker container"
	@echo "  docker-minio-run    Start MinIO in Docker container"
	@echo "  docker-minio-stop   Stop and remove MinIO Docker container"
	@echo ""
	@echo "Other:"
	@echo "  build             Build the Go binary for linux/amd64"
	@echo "  clean             Remove build artifacts"
	@echo "  help              Show this help message"

# Default target
.DEFAULT_GOAL := help

# --- Convenience Targets ---
.PHONY: deps-run deps-stop

# Start local development dependencies (PostgreSQL + MinIO)
deps-run: docker-postgres-run docker-minio-run

# Stop local development dependencies
deps-stop: docker-postgres-stop docker-minio-stop