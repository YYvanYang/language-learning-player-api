# Makefile for the Language Learning Player API

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
# PostgreSQL Docker settings (for local 'make deps-run')
PG_CONTAINER_NAME ?= language-learner-postgres
PG_USER ?= user
PG_PASSWORD ?= password
PG_DB ?= language_learner_db
PG_PORT ?= 5432
PG_VERSION ?= 16
PG_READY_TIMEOUT ?= 30 # Seconds to wait for PostgreSQL to be ready
# MinIO Docker settings (for local 'make deps-run')
MINIO_CONTAINER_NAME ?= language-learner-minio
MINIO_ROOT_USER ?= minioadmin
MINIO_ROOT_PASSWORD ?= minioadmin
MINIO_API_PORT ?= 9000
MINIO_CONSOLE_PORT ?= 9001
MINIO_BUCKET_NAME ?= language-audio # Ensure this matches config.development.yaml AND docker-compose.yml .env
MINIO_READY_TIMEOUT ?= 30 # Seconds to wait for MinIO to be ready
# Migrate CLI path relative to project root
MIGRATIONS_PATH=migrations
# Swag CLI variables (if using swaggo/swag)
SWAG_ENTRY_POINT=${CMD_PATH}/main.go
SWAG_OUTPUT_DIR=./docs
# Docker image settings (IMPORTANT: Customize for your Docker Hub/Registry)
DOCKER_IMAGE_NAME ?= your-dockerhub-username/language-player-api # <<< CHANGE THIS
DOCKER_IMAGE_TAG ?= latest
DOCKER_PLATFORM ?= linux/arm64 # Target platform for Raspberry Pi
DOCKER_BUILDX_PUSH_FLAGS=--platform $(DOCKER_PLATFORM) --push
DOCKER_BUILDX_BUILD_FLAGS=--platform $(DOCKER_PLATFORM) --load # Use --load to build locally for the target platform

# --- Go Tools Installation ---
GOPATH := $(shell go env GOPATH)
GOBIN ?= $(firstword $(shell go env GOBIN) $(GOPATH)/bin $(HOME)/go/bin)

# Tool binaries
MIGRATE := $(GOBIN)/migrate
SWAG := $(GOBIN)/swag
GOLANGCILINT := $(GOBIN)/golangci-lint
GOVULNCHECK := $(GOBIN)/govulncheck
MOCKERY := $(GOBIN)/mockery

.PHONY: all build clean \
	tools install-migrate install-swag install-lint install-vulncheck install-mockery \
	generate generate-swag generate-mocks swagger \
	run deps-run deps-stop \
	migrate-create migrate-up migrate-down migrate-force check-db-url \
	test test-unit test-integration test-cover lint fmt check-vuln \
	docker-build docker-build-arm64 docker-build-push-arm64 docker-run docker-stop docker-push \
	docker-postgres-run docker-postgres-stop docker-minio-run docker-minio-stop \
	help

# Default target (often set to 'all' or 'help')
all: help

# Target to install all necessary Go tools
tools: install-migrate install-swag install-lint install-vulncheck install-mockery

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

# Check if swag is installed, if not, install it
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

# Install latest mockery v3+
install-mockery:
	@if ! command -v mockery &> /dev/null; then \
		echo ">>> Installing mockery CLI (latest v3+)..."; \
		if go install github.com/vektra/mockery/v3@v3.0.0; then \
			echo ">>> mockery installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install mockery. Please check Go version (needs 1.18+) and network connectivity."; \
			exit 1; \
		fi; \
	else \
		echo ">>> mockery is already installed."; \
	fi

# --- Build ---
# Build the Go binary for linux/amd64 (Standard build for CI/Cloud)
build: clean tools
	@echo ">>> Building binary for linux/amd64..."
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo ">>> Binary built at $(OUTPUT_DIR)/$(BINARY_NAME)"

# Remove build artifacts
clean:
	@echo ">>> Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f coverage.out

# --- Run ---
# Run the application locally using go run (requires dependencies like DB running via 'make deps-run')
run: tools
	@echo ">>> Running application locally (using go run - requires deps-run)..."
	@APP_ENV=development go run $(CMD_PATH)/main.go

# --- Database Migrations ---
# Internal target to check if DATABASE_URL is set
check-db-url:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo ">>> ERROR: DATABASE_URL environment variable is not set."; \
		echo ">>> Please set it before running migrations, e.g.: export DATABASE_URL='postgresql://user:password@host:port/db?sslmode=disable'"; \
		exit 1; \
	fi

# Create a new migration file
migrate-create: tools
	@echo ">>> Creating migration file: $(name)..."
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo ">>> Migration file created."

# Apply all up migrations (requires DB running and DATABASE_URL set/exported)
migrate-up: tools check-db-url
	@echo ">>> Applying database migrations..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) up
	@echo ">>> Migrations applied."

# Roll back the last migration (requires DB running and DATABASE_URL set/exported)
migrate-down: tools check-db-url
	@echo ">>> Reverting last database migration..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) down 1
	@echo ">>> Last migration reverted."

# Force set migration version (Use with caution!) (requires DB running and DATABASE_URL set/exported)
migrate-force: tools check-db-url
	@echo ">>> Forcing migration version to $(version)..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) force $(version)
	@echo ">>> Migration version forced."

# --- Code Generation ---
generate: generate-swag generate-mocks

# Generate OpenAPI docs using swag
generate-swag: tools
	@echo ">>> Generating OpenAPI docs using swag..."
	@$(SWAG) init -g $(SWAG_ENTRY_POINT) --output $(SWAG_OUTPUT_DIR)
	@echo ">>> OpenAPI docs generated in $(SWAG_OUTPUT_DIR)."

# Alias for generating OpenAPI docs
swagger: generate-swag

# Generate mocks using mockery (reads .mockery.yaml)
generate-mocks: tools
	@echo ">>> Generating mocks using mockery (reading .mockery.yaml)..."
	@if [ ! -f .mockery.yaml ]; then \
		echo ">>> ERROR: .mockery.yaml config file not found. Please create it."; \
		exit 1; \
	fi
	@echo ">>> Cleaning existing mocks in internal/mocks/port/..."
	@rm -rf ./internal/mocks/port
	@mkdir -p ./internal/mocks/port
	@echo ">>> Running mockery..."
	@$(MOCKERY)
	@echo ">>> Mocks generation complete (check output above for errors)."

# --- Testing ---
# Run all tests (unit + integration, integration might require Docker for deps)
test: tools
	@echo ">>> Running all tests (unit + integration)..."
	@go test $(GO_TEST_FLAGS)
	@echo ">>> Tests complete. Coverage report generated at coverage.out"

# Run only unit tests (placeholder, assumes tests use -short)
test-unit: tools
	@echo ">>> Running unit tests (placeholder)..."
	@go test $(GO_TEST_FLAGS) -short

# Run only integration tests (requires Docker for deps)
test-integration: tools deps-run # Start dependencies for integration tests
	@echo ">>> Running integration tests (requires Docker)..."
	@echo ">>> Waiting for dependencies to be fully ready..."
	@sleep 5 # Give extra time after deps-run checks pass
	@# Specify integration test package(s) or use build tags
	@# Example: go test ./internal/adapter/repository/postgres/... -v
	@# Or: go test ./... -tags=integration -v
	@echo ">>> Running integration tests (replace with your specific command)..."
	@go test ./internal/adapter/repository/postgres/... -v
	@echo ">>> Integration tests finished."
	@make deps-stop # Stop dependencies after tests

# Show test coverage in browser
test-cover: test
	@echo ">>> Opening test coverage report..."
	@go tool cover -html=coverage.out

# --- Linting & Formatting ---
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
# Build Docker image for the host architecture (e.g., amd64 on Surface)
docker-build:
	@echo ">>> Building Docker image for host architecture [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo ">>> Docker image built."

# Build Docker image for ARM64 (for Pi), loading it into local Docker daemon (for testing build process)
docker-build-arm64:
	@echo ">>> Building Docker image for ARM64 (local load) [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@echo ">>> Note: Requires Docker Buildx setup (usually available with Docker Desktop)."
	@docker buildx build $(DOCKER_BUILDX_BUILD_FLAGS) -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo ">>> ARM64 Docker image built and loaded locally."

# Build Docker image for ARM64 AND push to registry (for Pi deployment)
docker-build-push-arm64:
	@echo ">>> Building AND PUSHING Docker image for ARM64 [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@echo ">>> Note: Ensure DOCKER_IMAGE_NAME is set correctly for your registry (e.g., your Docker Hub username)."
	@echo ">>> Note: Requires Docker Buildx setup and 'docker login'."
	@docker buildx build $(DOCKER_BUILDX_PUSH_FLAGS) -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo ">>> ARM64 Docker image built and pushed."

# Run Docker container locally (uses image built for host via 'docker-build' or loaded via 'docker-build-arm64')
# Uses environment variables from a local .env file if present. Requires deps running.
docker-run:
	@echo ">>> Running Docker container [$(BINARY_NAME)] using image [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@echo ">>> Note: Assumes image for host arch exists. Run 'make docker-build' first."
	@echo ">>> Note: Requires dependencies (Postgres/MinIO) to be running. Use 'make deps-run'."
	@docker run -d --name $(BINARY_NAME) \
		-p 8080:8080 \
		--network=host \
		--env-file .env \
		$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo ">>> Container started. Use 'make docker-stop' to stop."

# Stop and remove the running container started by 'docker-run'
docker-stop:
	@echo ">>> Stopping and removing Docker container [$(BINARY_NAME)]..."
	@docker stop $(BINARY_NAME) || true
	@docker rm $(BINARY_NAME) || true
	@echo ">>> Container stopped and removed."

# Push Docker image (built for host architecture) to registry
docker-push: docker-build
	@echo ">>> Pushing HOST architecture Docker image [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@echo ">>> Note: Ensure DOCKER_IMAGE_NAME variable is set correctly for your registry."
	@echo ">>> Note: Use 'make docker-build-push-arm64' to push the ARM64 image for Raspberry Pi."
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo ">>> Image pushed."

# --- PostgreSQL Docker (Local Dev Dependency) ---
# Run PostgreSQL in Docker container
docker-postgres-run:
	@echo ">>> Ensuring PostgreSQL container [$(PG_CONTAINER_NAME)] is running..."
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
	@echo ">>> Connection string for migrations: $(DATABASE_URL)"

# Stop and remove PostgreSQL container
docker-postgres-stop:
	@echo ">>> Stopping and removing PostgreSQL container [$(PG_CONTAINER_NAME)]..."
	@docker stop $(PG_CONTAINER_NAME) || true
	@docker rm $(PG_CONTAINER_NAME) || true
	@echo ">>> PostgreSQL container stopped and removed."

# --- MinIO Docker (Local Dev Dependency) ---
# Run MinIO in Docker container
docker-minio-run:
	@echo ">>> Ensuring MinIO container [$(MINIO_CONTAINER_NAME)] is running..."
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
	until curl -s --max-time 1 --output /dev/null --fail "http://localhost:$(MINIO_API_PORT)/minio/health/live"; do \
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
# Show help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Local Development (on your PC/Mac/Surface):"
	@echo "  run               Run the application using 'go run' (requires dependencies running)"
	@echo "  deps-run          Start local PostgreSQL and MinIO containers"
	@echo "  deps-stop         Stop local PostgreSQL and MinIO containers"
	@echo "  tools             Install necessary Go CLI tools (migrate, swag, lint, vulncheck, mockery)"
	@echo "  fmt               Format Go code using go fmt and goimports"
	@echo "  lint              Run golangci-lint"
	@echo "  check-vuln        Check for known vulnerabilities"
	@echo ""
	@echo "Code Generation:"
	@echo "  generate          Run all code generators (swag, mockery)"
	@echo "  generate-swag     Generate OpenAPI docs using swag"
	@echo "  generate-mocks    Generate mocks for interfaces using mockery (reads .mockery.yaml)"
	@echo ""
	@echo "Database Migrations (Local or Remote DB):"
	@echo "  migrate-create name=<name> Create a new migration file"
	@echo "  migrate-up        Apply database migrations (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-down      Revert the last database migration (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-force version=<ver> Force migration version (requires DB running and DATABASE_URL set/exported)"
	@echo ""
	@echo "Testing (Local):"
	@echo "  test              Run all tests and generate coverage"
	@echo "  test-unit         Run unit tests (placeholder)"
	@echo "  test-integration  Start deps, run integration tests, stop deps"
	@echo "  test-cover        Show test coverage report in browser"
	@echo ""
	@echo "Building:"
	@echo "  build             Build the Go binary for linux/amd64 (typical CI/Cloud build)"
	@echo "  clean             Remove build artifacts"
	@echo ""
	@echo "Docker (Application Image):"
	@echo "  docker-build           Build the application Docker image for your current machine's architecture"
	@echo "  docker-build-arm64     Build the application Docker image for ARM64 (Pi) locally (requires buildx)"
	@echo "  docker-build-push-arm64 Build the ARM64 image AND PUSH it to Docker Hub/Registry (for Pi deployment)"
	@echo "  docker-run             Build (host arch) and run the application container locally (uses .env, requires deps-run)"
	@echo "  docker-stop            Stop and remove the running application container started by 'make docker-run'"
	@echo "  docker-push            Push the HOST architecture image to Docker Hub/Registry"
	@echo ""
	@echo "Docker - Dependencies (Local Dev):"
	@echo "  docker-postgres-run    Start PostgreSQL in Docker container locally"
	@echo "  docker-postgres-stop   Stop and remove PostgreSQL Docker container locally"
	@echo "  docker-minio-run       Start MinIO in Docker container locally"
	@echo "  docker-minio-stop      Stop and remove MinIO Docker container locally"
	@echo ""
	@echo "Raspberry Pi Deployment Workflow:"
	@echo "  1. (On your Surface) Customize DOCKER_IMAGE_NAME in Makefile (e.g., your Docker Hub username)."
	@echo "  2. (On your Surface) Run 'docker login' to log in to Docker Hub."
	@echo "  3. (On your Surface) Run 'make docker-build-push-arm64' to build and push the ARM64 image."
	@echo "  4. (On your Pi) SSH into your Raspberry Pi."
	@echo "  5. (On your Pi) Create a deployment directory (e.g., '~/language-player-deployment')."
	@echo "  6. (On your Pi) Create 'docker-compose.yml' and '.env' files in that directory (see examples/docs)."
	@echo "     -> IMPORTANT: Update image name in docker-compose.yml to match Step 1."
	@echo "     -> IMPORTANT: Configure volumes in docker-compose.yml to use your SSD path."
	@echo "     -> IMPORTANT: Set strong passwords and secrets in .env."
	@echo "  7. (On your Pi) Run 'cd ~/language-player-deployment && docker compose up -d' to start services."
	@echo "  8. (On your Pi/Surface) Run database migrations (e.g., 'export DATABASE_URL=... && make migrate-up' from Surface)."
	@echo "  9. (On your Pi/Surface) Create the MinIO bucket ('language-audio') via console or 'mc'."
	@echo "  10. (On your Pi) Check logs: 'docker compose logs -f'"
	@echo "  11. (On your Pi) Stop services: 'docker compose down'"
	@echo ""
	@echo "Other:"
	@echo "  help              Show this help message"

# Default target executed when 'make' is run without arguments
.DEFAULT_GOAL := help

# --- Convenience Targets ---
# Start local development dependencies (PostgreSQL + MinIO)
deps-run: docker-postgres-run docker-minio-run

# Stop local development dependencies
deps-stop: docker-postgres-stop docker-minio-stop