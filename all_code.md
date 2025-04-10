# language-learning-player-backend Codebase

## `go.mod`

```
module github.com/yvanyang/language-learning-player-api

go 1.23.3

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/go-playground/validator/v10 v10.26.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v7 v7.0.89
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.4
	golang.org/x/crypto v0.36.0
	golang.org/x/time v0.11.0
	google.golang.org/api v0.228.0
)

require (
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/minio/crc64nvme v1.0.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

## `Dockerfile`

```
# Dockerfile

# ---- Build Stage ----
    FROM golang:1.21-alpine AS builder

    # Set working directory inside the container
    WORKDIR /app
    
    # Install build tools if needed (e.g., git for private repos, gcc for cgo if CGO_ENABLED=1)
    # RUN apk add --no-cache git build-base
    
    # Download Go modules to leverage Docker layer caching
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the entire application source code
    COPY . .
    
    # Build the Go application
    # - Static linking (optional but recommended for alpine/distroless)
    # - Linker flags to reduce binary size (optional)
    # - Trimpath to remove absolute file paths
    # - Set output path
    ARG TARGETOS=linux
    ARG TARGETARCH=amd64
    RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
        -ldflags='-w -s' \
        -trimpath \
        -o /app/language-player-api \
        ./cmd/api
    
    # ---- Final Stage ----
    # Use a minimal base image
    # Option 1: Alpine (small, includes shell)
    FROM alpine:3.18
    
    # Option 2: Distroless static (even smaller, no shell, requires fully static binary from build stage)
    # FROM gcr.io/distroless/static-debian11 AS final
    
    # Set working directory
    WORKDIR /app
    
    # Copy necessary files from the builder stage
    # Copy only the compiled binary
    COPY --from=builder /app/language-player-api /app/language-player-api
    # Copy configuration templates or default configs (if needed at runtime)
    COPY config.example.yaml /app/config.example.yaml
    # Copy migration files (needed if running migrations from container, or for reference)
    COPY migrations /app/migrations
    
    # Set permissions (Optional, good practice if not running as root)
    RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    RUN chown -R appuser:appgroup /app
    USER appuser
    
    # Expose the port the application listens on (from config, default 8080)
    EXPOSE 8080
    
    # Define the entry point for the container
    # This command will run when the container starts
    ENTRYPOINT ["/app/language-player-api"]
    
    # Optional: Define a default command (can be overridden)
    # CMD [""]
```

## `Makefile`

```
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
GOPATH := $(shell go env GOPATH)
GOBIN ?= $(firstword $(shell go env GOBIN) $(GOPATH)/bin $(HOME)/go/bin)

# Tool binaries
MIGRATE := $(GOBIN)/migrate
SWAG := $(GOBIN)/swag
GOLANGCILINT := $(GOBIN)/golangci-lint
GOVULNCHECK := $(GOBIN)/govulncheck
MOCKERY := $(GOBIN)/mockery # ADDED: Variable for mockery path

# MODIFIED .PHONY: Added install-mockery, generate-mocks. Removed install-sqlc.
.PHONY: tools install-migrate install-swag install-lint install-vulncheck install-mockery generate generate-swag generate-mocks swagger test test-unit test-integration test-cover lint fmt check-vuln docker-build docker-run docker-stop docker-push docker-postgres-run docker-postgres-stop docker-minio-run docker-minio-stop deps-run deps-stop help clean migrate-create migrate-up migrate-down migrate-force check-db-url build run

# Target to install all necessary Go tools
# MODIFIED: Added install-mockery dependency
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

# MODIFIED: Install latest mockery v3+
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
# Run the application locally (requires dependencies like DB running)
run: tools
	@echo ">>> Running application locally (using go run)..."
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
migrate-force: tools check-db-url
	@echo ">>> Forcing migration version to $(version)..."
	@$(MIGRATE) -database "$(DATABASE_URL)" -path $(MIGRATIONS_PATH) force $(version)
	@echo ">>> Migration version forced."


# --- Code Generation ---
# MODIFIED: Added generate-mocks to generate target
generate: generate-swag generate-mocks

# Generate OpenAPI docs using swag (Optional)
generate-swag: tools
	@echo ">>> Generating OpenAPI docs using swag..."
	@echo ">>> Using swag command: $(SWAG)"
	@$(SWAG) init -g $(SWAG_ENTRY_POINT) --output $(SWAG_OUTPUT_DIR)
	@echo ">>> OpenAPI docs generated in $(SWAG_OUTPUT_DIR)."

# Alias for generating OpenAPI docs
swagger: generate-swag

# MODIFIED: Generate mocks using mockery (reads .mockery.yaml)
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
# Run all tests (unit + integration, requires Docker for integration)
test: tools
	@echo ">>> Running all tests (unit + integration)..."
	@go test $(GO_TEST_FLAGS)
	@echo ">>> Tests complete. Coverage report generated at coverage.out"

# Run only unit tests (placeholder)
test-unit: tools
	@echo ">>> Running unit tests (placeholder)..."
	@go test $(GO_TEST_FLAGS) -short # -short flag might skip long-running tests if tests use it

# Run only integration tests (requires Docker)
test-integration: tools
	@echo ">>> Running integration tests (requires Docker)..."
	@go test ./internal/adapter/repository/postgres/... -v # Run tests specifically in the repo package
	# Or use build tags: @go test ./... -tags=integration -v

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
# Build Docker image
docker-build:
	@echo ">>> Building Docker image [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo ">>> Docker image built."

# Run Docker container locally (using env vars from .env file if present)
docker-run: docker-build
	@echo ">>> Running Docker container [$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)]..."
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
	@echo "Local Development:"
	@echo "  run               Run the application locally (requires dependencies: use 'make deps-run')"
	@echo "  deps-run          Start local PostgreSQL and MinIO containers"
	@echo "  deps-stop         Stop local PostgreSQL and MinIO containers"
	@echo "  tools             Install necessary Go CLI tools (migrate, swag, lint, vulncheck, mockery)" # MODIFIED
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-create name=<name> Create a new migration file"
	@echo "  migrate-up        Apply database migrations (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-down      Revert the last database migration (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-force version=<ver> Force migration version (requires DB running and DATABASE_URL set/exported)"
	@echo ""
	@echo "Code Generation & Formatting:"
	@echo "  generate          Run all code generators (swag, mockery)" # MODIFIED
	@echo "  generate-swag     Generate OpenAPI docs using swag"
	@echo "  generate-mocks    Generate mocks for interfaces using mockery (reads .mockery.yaml)" # ADDED
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
# Start local development dependencies (PostgreSQL + MinIO)
deps-run: docker-postgres-run docker-minio-run

# Stop local development dependencies
deps-stop: docker-postgres-stop docker-minio-stop
```

## `.mockery.yaml`

```yaml
# .mockery.yaml - Mockery v3+ Configuration
# See: https://vektra.github.io/mockery/latest/configuration/
# See: https://vektra.github.io/mockery/latest/template/

# --- Global Output Configuration ---

# The Go package name for all generated mock files.
# This is a valid top-level key.
pkgname: mocks

# MOVED dir to per-package config as it uses package-specific template variables.
# dir: ./internal/mocks/{{base .PackagePath}}

# --- Global Naming Conventions ---

# Customize the generated filename (Default: mock_{{.InterfaceName}}.go)
# This is a valid top-level key.
filename: "mock_{{.InterfaceName}}.go"

# Customize the generated mock struct name (Default: Mock{{.InterfaceName}})
# This is a valid top-level key.
structname: "Mock{{.InterfaceName}}"

# REMOVED: Invalid top-level key for v3.0.0: header-template
# Header customization might require a custom template.

# --- Interface Discovery ---

# Specify the packages containing the interfaces to mock.
packages:
  # Target the 'port' package where your repository, service, and use case interfaces live.
  github.com/yvanyang/language-learning-player-api/internal/port:
    config:
      # Explicitly tell mockery to generate mocks for ALL interfaces found in this package.
      all: true
      # CORRECTED: Specify output dir using the documented `.SrcPackagePath` variable
      # and the `base` function to get the desired output structure.
      # Example: For src 'internal/port', output is './internal/mocks/port/'
      dir: ./internal/mocks/{{base .SrcPackagePath}}
    # You could override other global settings per-package here if needed, e.g.:
    # config:
    #   all: true
    #   dir: ./internal/mocks/{{base .SrcPackagePath}}
    #   pkgname: specificmocks
    #   filename: "iface_{{.InterfaceName}}_mock.go"

# --- Optional Settings ---

# quiet: false # Set to true to suppress informational messages during generation.
log-level: info # CHANGED BACK: Set log level back to info after successful generation
# recursive: false # Usually false when using the 'packages' directive.
# all: false # Usually false, specify packages explicitly.
# testonly: false # Keep false for shared mocks.
#keeptree: false # Keep false when using `dir` with templates for central directory.
#inpackage: false # MUST be false (or omitted) to use 'pkgname' and 'dir' for a central mocks directory.
```

## `go.sum`

```
cloud.google.com/go/auth v0.15.0 h1:Ly0u4aA5vG/fsSsxu98qCQBemXtAtJf+95z9HK+cxps=
cloud.google.com/go/auth v0.15.0/go.mod h1:WJDGqZ1o9E9wKIL+IwStfyn/+s59zl4Bi+1KQNVXLZ8=
cloud.google.com/go/auth/oauth2adapt v0.2.8 h1:keo8NaayQZ6wimpNSmW5OPc283g65QNIiLpZnkHRbnc=
cloud.google.com/go/auth/oauth2adapt v0.2.8/go.mod h1:XQ9y31RkqZCcwJWNSx2Xvric3RrU88hAYYbjDWYDL+c=
cloud.google.com/go/compute/metadata v0.6.0 h1:A6hENjEsCDtC1k8byVsgwvVcioamEHvZ4j01OwKxG9I=
cloud.google.com/go/compute/metadata v0.6.0/go.mod h1:FjyFAW1MW0C203CEOMDTu3Dk1FlqW3Rga40jzHL4hfg=
github.com/KyleBanks/depth v1.2.1 h1:5h8fQADFrWtarTdtDudMmGsC7GPbOAu6RVB3ffsVFHc=
github.com/KyleBanks/depth v1.2.1/go.mod h1:jzSb9d0L43HxTQfT+oSA1EEp2q+ne2uh6XgeJcm8brE=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/dustin/go-humanize v1.0.1 h1:GzkhY7T5VNhEkwH0PVJgjz+fX1rhBrR7pRT3mDkpeCY=
github.com/dustin/go-humanize v1.0.1/go.mod h1:Mu1zIs6XwVuF/gI1OepvI0qD18qycQx+mFykh5fBlto=
github.com/felixge/httpsnoop v1.0.4 h1:NFTV2Zj1bL4mc9sqWACXbQFVBBg2W3GPvqp8/ESS2Wg=
github.com/felixge/httpsnoop v1.0.4/go.mod h1:m8KPJKqk1gH5J9DgRY2ASl2lWCfGKXixSwevea8zH2U=
github.com/frankban/quicktest v1.14.6 h1:7Xjx+VpznH+oBnejlPUj8oUpdxnVs4f8XU8WnHkI4W8=
github.com/frankban/quicktest v1.14.6/go.mod h1:4ptaffx2x8+WTWXmUCuVU6aPUX1/Mz7zb5vbUoiM6w0=
github.com/fsnotify/fsnotify v1.8.0 h1:dAwr6QBTBZIkG8roQaJjGof0pp0EeF+tNV7YBP3F/8M=
github.com/fsnotify/fsnotify v1.8.0/go.mod h1:8jBTzvmWwFyi3Pb8djgCCO5IBqzKJ/Jwo8TRcHyHii0=
github.com/gabriel-vasile/mimetype v1.4.8 h1:FfZ3gj38NjllZIeJAmMhr+qKL8Wu+nOoI3GqacKw1NM=
github.com/gabriel-vasile/mimetype v1.4.8/go.mod h1:ByKUIKGjh1ODkGM1asKUbQZOLGrPjydw3hYPU2YU9t8=
github.com/go-chi/chi/v5 v5.2.1 h1:KOIHODQj58PmL80G2Eak4WdvUzjSJSm0vG72crDCqb8=
github.com/go-chi/chi/v5 v5.2.1/go.mod h1:L2yAIGWB3H+phAw1NxKwWM+7eUH/lU8pOMm5hHcoops=
github.com/go-chi/cors v1.2.1 h1:xEC8UT3Rlp2QuWNEr4Fs/c2EAGVKBwy/1vHx3bppil4=
github.com/go-chi/cors v1.2.1/go.mod h1:sSbTewc+6wYHBBCW7ytsFSn836hqM7JxpglAy2Vzc58=
github.com/go-ini/ini v1.67.0 h1:z6ZrTEZqSWOTyH2FlglNbNgARyHG8oLW9gMELqKr06A=
github.com/go-ini/ini v1.67.0/go.mod h1:ByCAeIL28uOIIG0E3PJtZPDL8WnHpFKFOtgjp+3Ies8=
github.com/go-logr/logr v1.2.2/go.mod h1:jdQByPbusPIv2/zmleS9BjJVeZ6kBagPoEUsqbVz/1A=
github.com/go-logr/logr v1.4.2 h1:6pFjapn8bFcIbiKo3XT4j/BhANplGihG6tvd+8rYgrY=
github.com/go-logr/logr v1.4.2/go.mod h1:9T104GzyrTigFIr8wt5mBrctHMim0Nb2HLGrmQ40KvY=
github.com/go-logr/stdr v1.2.2 h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=
github.com/go-logr/stdr v1.2.2/go.mod h1:mMo/vtBO5dYbehREoey6XUKy/eSumjCCveDpRre4VKE=
github.com/go-openapi/jsonpointer v0.21.1 h1:whnzv/pNXtK2FbX/W9yJfRmE2gsmkfahjMKB0fZvcic=
github.com/go-openapi/jsonpointer v0.21.1/go.mod h1:50I1STOfbY1ycR8jGz8DaMeLCdXiI6aDteEdRNNzpdk=
github.com/go-openapi/jsonreference v0.21.0 h1:Rs+Y7hSXT83Jacb7kFyjn4ijOuVGSvOdF2+tg1TRrwQ=
github.com/go-openapi/jsonreference v0.21.0/go.mod h1:LmZmgsrTkVg9LG4EaHeY8cBDslNPMo06cago5JNLkm4=
github.com/go-openapi/spec v0.21.0 h1:LTVzPc3p/RzRnkQqLRndbAzjY0d0BCL72A6j3CdL9ZY=
github.com/go-openapi/spec v0.21.0/go.mod h1:78u6VdPw81XU44qEWGhtr982gJ5BWg2c0I5XwVMotYk=
github.com/go-openapi/swag v0.23.1 h1:lpsStH0n2ittzTnbaSloVZLuB5+fvSY/+hnagBjSNZU=
github.com/go-openapi/swag v0.23.1/go.mod h1:STZs8TbRvEQQKUA+JZNAm3EWlgaOBGpyFDqQnDHMef0=
github.com/go-playground/assert/v2 v2.2.0 h1:JvknZsQTYeFEAhQwI4qEt9cyV5ONwRHC+lYKSsYSR8s=
github.com/go-playground/assert/v2 v2.2.0/go.mod h1:VDjEfimB/XKnb+ZQfWdccd7VUvScMdVu0Titje2rxJ4=
github.com/go-playground/locales v0.14.1 h1:EWaQ/wswjilfKLTECiXz7Rh+3BjFhfDFKv/oXslEjJA=
github.com/go-playground/locales v0.14.1/go.mod h1:hxrqLVvrK65+Rwrd5Fc6F2O76J/NuW9t0sjnWqG1slY=
github.com/go-playground/universal-translator v0.18.1 h1:Bcnm0ZwsGyWbCzImXv+pAJnYK9S473LQFuzCbDbfSFY=
github.com/go-playground/universal-translator v0.18.1/go.mod h1:xekY+UJKNuX9WP91TpwSH2VMlDf28Uj24BCp08ZFTUY=
github.com/go-playground/validator/v10 v10.26.0 h1:SP05Nqhjcvz81uJaRfEV0YBSSSGMc/iMaVtFbr3Sw2k=
github.com/go-playground/validator/v10 v10.26.0/go.mod h1:I5QpIEbmr8On7W0TktmJAumgzX4CA1XNl4ZmDuVHKKo=
github.com/go-viper/mapstructure/v2 v2.2.1 h1:ZAaOCxANMuZx5RCeg0mBdEZk7DZasvvZIxtHqx8aGss=
github.com/go-viper/mapstructure/v2 v2.2.1/go.mod h1:oJDH3BJKyqBA2TXFhDsKDGDTlndYOZ6rGS0BRZIxGhM=
github.com/goccy/go-json v0.10.5 h1:Fq85nIqj+gXn/S5ahsiTlK3TmC85qgirsdTP/+DeaC4=
github.com/goccy/go-json v0.10.5/go.mod h1:oq7eo15ShAhp70Anwd5lgX2pLfOS3QCiwU/PULtXL6M=
github.com/golang-jwt/jwt/v5 v5.2.2 h1:Rl4B7itRWVtYIHFrSNd7vhTiz9UpLdi6gZhZ3wEeDy8=
github.com/golang-jwt/jwt/v5 v5.2.2/go.mod h1:pqrtFR0X4osieyHYxtmOUWsAWrfe1Q5UVIyoH402zdk=
github.com/golang/protobuf v1.5.4 h1:i7eJL8qZTpSEXOPTxNKhASYpMn+8e5Q6AdndVa1dWek=
github.com/golang/protobuf v1.5.4/go.mod h1:lnTiLA8Wa4RWRcIUkrtSVa5nRhsEGBg48fD6rSs7xps=
github.com/google/go-cmp v0.7.0 h1:wk8382ETsv4JYUZwIsn6YpYiWiBsYLSJiTsyBybVuN8=
github.com/google/go-cmp v0.7.0/go.mod h1:pXiqmnSA92OHEEa9HXL2W4E7lf9JzCmGVUdgjX3N/iU=
github.com/google/s2a-go v0.1.9 h1:LGD7gtMgezd8a/Xak7mEWL0PjoTQFvpRudN895yqKW0=
github.com/google/s2a-go v0.1.9/go.mod h1:YA0Ei2ZQL3acow2O62kdp9UlnvMmU7kA6Eutn0dXayM=
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/googleapis/enterprise-certificate-proxy v0.3.6 h1:GW/XbdyBFQ8Qe+YAmFU9uHLo7OnF5tL52HFAgMmyrf4=
github.com/googleapis/enterprise-certificate-proxy v0.3.6/go.mod h1:MkHOF77EYAE7qfSuSS9PU6g4Nt4e11cnsDUowfwewLA=
github.com/googleapis/gax-go/v2 v2.14.1 h1:hb0FFeiPaQskmvakKu5EbCbpntQn48jyHuvrkurSS/Q=
github.com/googleapis/gax-go/v2 v2.14.1/go.mod h1:Hb/NubMaVM88SrNkvl8X/o8XWwDJEPqouaLeN2IUxoA=
github.com/jackc/pgpassfile v1.0.0 h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=
github.com/jackc/pgpassfile v1.0.0/go.mod h1:CEx0iS5ambNFdcRtxPj5JhEz+xB6uRky5eyVu/W2HEg=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 h1:iCEnooe7UlwOQYpKFhBabPMi4aNAfoODPEFNiAnClxo=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761/go.mod h1:5TJZWKEWniPve33vlWYSoGYefn3gLQRzjfDlhSJ9ZKM=
github.com/jackc/pgx/v5 v5.7.4 h1:9wKznZrhWa2QiHL+NjTSPP6yjl3451BX3imWDnokYlg=
github.com/jackc/pgx/v5 v5.7.4/go.mod h1:ncY89UGWxg82EykZUwSpUKEfccBGGYq1xjrOpsbsfGQ=
github.com/jackc/puddle/v2 v2.2.2 h1:PR8nw+E/1w0GLuRFSmiioY6UooMp6KJv0/61nB7icHo=
github.com/jackc/puddle/v2 v2.2.2/go.mod h1:vriiEXHvEE654aYKXXjOvZM39qJ0q+azkZFrfEOc3H4=
github.com/josharian/intern v1.0.0 h1:vlS4z54oSdjm0bgjRigI+G1HpF+tI+9rE5LLzOg8HmY=
github.com/josharian/intern v1.0.0/go.mod h1:5DoeVV0s6jJacbCEi61lwdGj/aVlrQvzHFFd8Hwg//Y=
github.com/klauspost/compress v1.18.0 h1:c/Cqfb0r+Yi+JtIEq73FWXVkRonBlf0CRNYc8Zttxdo=
github.com/klauspost/compress v1.18.0/go.mod h1:2Pp+KzxcywXVXMr50+X0Q/Lsb43OQHYWRCY2AiWywWQ=
github.com/klauspost/cpuid/v2 v2.0.1/go.mod h1:FInQzS24/EEf25PyTYn52gqo7WaD8xa0213Md/qVLRg=
github.com/klauspost/cpuid/v2 v2.2.10 h1:tBs3QSyvjDyFTq3uoc/9xFpCuOsJQFNPiAhYdw2skhE=
github.com/klauspost/cpuid/v2 v2.2.10/go.mod h1:hqwkgyIinND0mEev00jJYCxPNVRVXFQeu1XKlok6oO0=
github.com/kr/pretty v0.3.1 h1:flRD4NNwYAUpkphVc1HcthR4KEIFJ65n8Mw5qdRn3LE=
github.com/kr/pretty v0.3.1/go.mod h1:hoEshYVHaxMs3cyo3Yncou5ZscifuDolrwPKZanG3xk=
github.com/kr/text v0.2.0 h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=
github.com/kr/text v0.2.0/go.mod h1:eLer722TekiGuMkidMxC/pM04lWEeraHUUmBw8l2grE=
github.com/leodido/go-urn v1.4.0 h1:WT9HwE9SGECu3lg4d/dIA+jxlljEa1/ffXKmRjqdmIQ=
github.com/leodido/go-urn v1.4.0/go.mod h1:bvxc+MVxLKB4z00jd1z+Dvzr47oO32F/QSNjSBOlFxI=
github.com/lib/pq v1.10.9 h1:YXG7RB+JIjhP29X+OtkiDnYaXQwpS4JEWq7dtCCRUEw=
github.com/lib/pq v1.10.9/go.mod h1:AlVN5x4E4T544tWzH6hKfbfQvm3HdbOxrmggDNAPY9o=
github.com/mailru/easyjson v0.9.0 h1:PrnmzHw7262yW8sTBwxi1PdJA3Iw/EKBa8psRf7d9a4=
github.com/mailru/easyjson v0.9.0/go.mod h1:1+xMtQp2MRNVL/V1bOzuP3aP8VNwRW55fQUto+XFtTU=
github.com/minio/crc64nvme v1.0.1 h1:DHQPrYPdqK7jQG/Ls5CTBZWeex/2FMS3G5XGkycuFrY=
github.com/minio/crc64nvme v1.0.1/go.mod h1:eVfm2fAzLlxMdUGc0EEBGSMmPwmXD5XiNRpnu9J3bvg=
github.com/minio/md5-simd v1.1.2 h1:Gdi1DZK69+ZVMoNHRXJyNcxrMA4dSxoYHZSQbirFg34=
github.com/minio/md5-simd v1.1.2/go.mod h1:MzdKDxYpY2BT9XQFocsiZf/NKVtR7nkE4RoEpN+20RM=
github.com/minio/minio-go/v7 v7.0.89 h1:hx4xV5wwTUfyv8LarhJAwNecnXpoTsj9v3f3q/ZkiJU=
github.com/minio/minio-go/v7 v7.0.89/go.mod h1:2rFnGAp02p7Dddo1Fq4S2wYOfpF0MUTSeLTRC90I204=
github.com/pelletier/go-toml/v2 v2.2.3 h1:YmeHyLY8mFWbdkNWwpr+qIL2bEqT0o95WSdkNHvL12M=
github.com/pelletier/go-toml/v2 v2.2.3/go.mod h1:MfCQTFTvCcUyyvvwm1+G6H/jORL20Xlb6rzQu9GuUkc=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/rogpeppe/go-internal v1.13.1 h1:KvO1DLK/DRN07sQ1LQKScxyZJuNnedQ5/wKSR38lUII=
github.com/rogpeppe/go-internal v1.13.1/go.mod h1:uMEvuHeurkdAXX61udpOXGD/AzZDWNMNyH2VO9fmH0o=
github.com/rs/xid v1.6.0 h1:fV591PaemRlL6JfRxGDEPl69wICngIQ3shQtzfy2gxU=
github.com/rs/xid v1.6.0/go.mod h1:7XoLgs4eV+QndskICGsho+ADou8ySMSjJKDIan90Nz0=
github.com/sagikazarmark/locafero v0.7.0 h1:5MqpDsTGNDhY8sGp0Aowyf0qKsPrhewaLSsFaodPcyo=
github.com/sagikazarmark/locafero v0.7.0/go.mod h1:2za3Cg5rMaTMoG/2Ulr9AwtFaIppKXTRYnozin4aB5k=
github.com/sourcegraph/conc v0.3.0 h1:OQTbbt6P72L20UqAkXXuLOj79LfEanQ+YQFNpLA9ySo=
github.com/sourcegraph/conc v0.3.0/go.mod h1:Sdozi7LEKbFPqYX2/J+iBAM6HpqSLTASQIKqDmF7Mt0=
github.com/spf13/afero v1.12.0 h1:UcOPyRBYczmFn6yvphxkn9ZEOY65cpwGKb5mL36mrqs=
github.com/spf13/afero v1.12.0/go.mod h1:ZTlWwG4/ahT8W7T0WQ5uYmjI9duaLQGy3Q2OAl4sk/4=
github.com/spf13/cast v1.7.1 h1:cuNEagBQEHWN1FnbGEjCXL2szYEXqfJPbP2HNUaca9Y=
github.com/spf13/cast v1.7.1/go.mod h1:ancEpBxwJDODSW/UG4rDrAqiKolqNNh2DX3mk86cAdo=
github.com/spf13/pflag v1.0.6 h1:jFzHGLGAlb3ruxLB8MhbI6A8+AQX/2eW4qeyNZXNp2o=
github.com/spf13/pflag v1.0.6/go.mod h1:McXfInJRrz4CZXVZOBLb0bTZqETkiAhM9Iw0y3An2Bg=
github.com/spf13/viper v1.20.1 h1:ZMi+z/lvLyPSCoNtFCpqjy0S4kPbirhpTMwl8BkW9X4=
github.com/spf13/viper v1.20.1/go.mod h1:P9Mdzt1zoHIG8m2eZQinpiBjo6kCmZSKBClNNqjJvu4=
github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
github.com/stretchr/objx v0.5.2 h1:xuMeJ0Sdp5ZMRXx/aWO6RZxdr3beISkG5/G/aIRr3pY=
github.com/stretchr/objx v0.5.2/go.mod h1:FRsXN1f5AsAjCGJKqEizvkpNtU+EGNCLh3NxZ/8L+MA=
github.com/stretchr/testify v1.3.0/go.mod h1:M5WIy9Dh21IEIfnGCwXGc5bZfKNJtfHm1UVUgZn+9EI=
github.com/stretchr/testify v1.7.0/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
github.com/stretchr/testify v1.10.0 h1:Xv5erBjTwe/5IxqUQTdXv5kgmIvbHo3QQyRwhJsOfJA=
github.com/stretchr/testify v1.10.0/go.mod h1:r2ic/lqez/lEtzL7wO/rwa5dbSLXVDPFyf8C91i36aY=
github.com/subosito/gotenv v1.6.0 h1:9NlTDc1FTs4qu0DDq7AEtTPNw6SVm7uBMsUCUjABIf8=
github.com/subosito/gotenv v1.6.0/go.mod h1:Dk4QP5c2W3ibzajGcXpNraDfq2IrhjMIvMSWPKKo0FU=
github.com/swaggo/files v1.0.1 h1:J1bVJ4XHZNq0I46UU90611i9/YzdrF7x92oX1ig5IdE=
github.com/swaggo/files v1.0.1/go.mod h1:0qXmMNH6sXNf+73t65aKeB+ApmgxdnkQzVTAj2uaMUg=
github.com/swaggo/http-swagger v1.3.4 h1:q7t/XLx0n15H1Q9/tk3Y9L4n210XzJF5WtnDX64a5ww=
github.com/swaggo/http-swagger v1.3.4/go.mod h1:9dAh0unqMBAlbp1uE2Uc2mQTxNMU/ha4UbucIg1MFkQ=
github.com/swaggo/swag v1.16.4 h1:clWJtd9LStiG3VeijiCfOVODP6VpHtKdQy9ELFG3s1A=
github.com/swaggo/swag v1.16.4/go.mod h1:VBsHJRsDvfYvqoiMKnsdwhNV9LEMHgEDZcyVYX0sxPg=
github.com/yuin/goldmark v1.4.13/go.mod h1:6yULJ656Px+3vBD8DxQVa3kxgyrAnzto9xy5taEt/CY=
go.opentelemetry.io/auto/sdk v1.1.0 h1:cH53jehLUN6UFLY71z+NDOiNJqDdPRaXzTel0sJySYA=
go.opentelemetry.io/auto/sdk v1.1.0/go.mod h1:3wSPjt5PWp2RhlCcmmOial7AvC4DQqZb7a7wCow3W8A=
go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0 h1:rgMkmiGfix9vFJDcDi1PK8WEQP4FLQwLDfhp5ZLpFeE=
go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0/go.mod h1:ijPqXp5P6IRRByFVVg9DY8P5HkxkHE5ARIa+86aXPf4=
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 h1:CV7UdSGJt/Ao6Gp4CXckLxVRRsRgDHoI8XjbL3PDl8s=
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0/go.mod h1:FRmFuRJfag1IZ2dPkHnEoSFVgTVPUd2qf5Vi69hLb8I=
go.opentelemetry.io/otel v1.34.0 h1:zRLXxLCgL1WyKsPVrgbSdMN4c0FMkDAskSTQP+0hdUY=
go.opentelemetry.io/otel v1.34.0/go.mod h1:OWFPOQ+h4G8xpyjgqo4SxJYdDQ/qmRH+wivy7zzx9oI=
go.opentelemetry.io/otel/metric v1.34.0 h1:+eTR3U0MyfWjRDhmFMxe2SsW64QrZ84AOhvqS7Y+PoQ=
go.opentelemetry.io/otel/metric v1.34.0/go.mod h1:CEDrp0fy2D0MvkXE+dPV7cMi8tWZwX3dmaIhwPOaqHE=
go.opentelemetry.io/otel/sdk v1.34.0 h1:95zS4k/2GOy069d321O8jWgYsW3MzVV+KuSPKp7Wr1A=
go.opentelemetry.io/otel/sdk v1.34.0/go.mod h1:0e/pNiaMAqaykJGKbi+tSjWfNNHMTxoC9qANsCzbyxU=
go.opentelemetry.io/otel/sdk/metric v1.34.0 h1:5CeK9ujjbFVL5c1PhLuStg1wxA7vQv7ce1EK0Gyvahk=
go.opentelemetry.io/otel/sdk/metric v1.34.0/go.mod h1:jQ/r8Ze28zRKoNRdkjCZxfs6YvBTG1+YIqyFVFYec5w=
go.opentelemetry.io/otel/trace v1.34.0 h1:+ouXS2V8Rd4hp4580a8q23bg0azF2nI8cqLYnC8mh/k=
go.opentelemetry.io/otel/trace v1.34.0/go.mod h1:Svm7lSjQD7kG7KJ/MUHPVXSDGz2OX4h0M2jHBhmSfRE=
go.uber.org/atomic v1.9.0 h1:ECmE8Bn/WFTYwEW/bpKD3M8VtR/zQVbavAoalC1PYyE=
go.uber.org/atomic v1.9.0/go.mod h1:fEN4uk6kAWBTFdckzkM89CLk9XfWZrxpCo0nPH17wJc=
go.uber.org/multierr v1.9.0 h1:7fIwc/ZtS0q++VgcfqFDxSBZVv/Xo49/SYnDFupUwlI=
go.uber.org/multierr v1.9.0/go.mod h1:X2jQV1h+kxSjClGpnseKVIxpmcjrj7MNnI0bnlfKTVQ=
golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2/go.mod h1:djNgcEr1/C05ACkg1iLfiJU5Ep61QUkGW8qpdssI0+w=
golang.org/x/crypto v0.0.0-20210921155107-089bfa567519/go.mod h1:GvvjBRRGRdwPK5ydBHafDWAxML/pGHZbMvKqRZ5+Abc=
golang.org/x/crypto v0.36.0 h1:AnAEvhDddvBdpY+uR+MyHmuZzzNqXSe/GvuDeob5L34=
golang.org/x/crypto v0.36.0/go.mod h1:Y4J0ReaxCR1IMaabaSMugxJES1EpwhBHhv2bDHklZvc=
golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4/go.mod h1:jJ57K6gSWd91VN4djpZkiMVwK6gcyfeH4XE8wZrZaV4=
golang.org/x/mod v0.24.0 h1:ZfthKaKaT4NrhGVZHO1/WDTwGES4De8KtWO0SIbNJMU=
golang.org/x/mod v0.24.0/go.mod h1:IXM97Txy2VM4PJ3gI61r1YEk/gAj6zAHN3AdZt6S9Ww=
golang.org/x/net v0.0.0-20190620200207-3b0461eec859/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
golang.org/x/net v0.0.0-20210226172049-e18ecbb05110/go.mod h1:m0MpNAwzfU5UDzcl9v0D8zg8gWTRqZa9RBIspLL5mdg=
golang.org/x/net v0.0.0-20220722155237-a158d28d115b/go.mod h1:XRhObCWvk6IyKnWLug+ECip1KBveYUHfp+8e9klMJ9c=
golang.org/x/net v0.7.0/go.mod h1:2Tu9+aMcznHK/AK1HMvgo6xiTLG5rD5rZLDS+rp2Bjs=
golang.org/x/net v0.37.0 h1:1zLorHbz+LYj7MQlSf1+2tPIIgibq2eL5xkrGk6f+2c=
golang.org/x/net v0.37.0/go.mod h1:ivrbrMbzFq5J41QOQh0siUuly180yBYtLp+CKbEaFx8=
golang.org/x/oauth2 v0.28.0 h1:CrgCKl8PPAVtLnU3c+EDw6x11699EWlsDeWNWKdIOkc=
golang.org/x/oauth2 v0.28.0/go.mod h1:onh5ek6nERTohokkhCD/y2cV4Do3fxFHFuAejCkRWT8=
golang.org/x/sync v0.0.0-20190423024810-112230192c58/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.12.0 h1:MHc5BpPuC30uJk597Ri8TV3CNZcTLu6B6z4lJy+g6Jw=
golang.org/x/sync v0.12.0/go.mod h1:1dzgHSNfp02xaA81J2MS99Qcpr2w7fw1gpm99rleRqA=
golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
golang.org/x/sys v0.0.0-20201119102817-f84b799fce68/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.5.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.31.0 h1:ioabZlmFYtWhL+TRYpcnNlLwhyxaM9kWTDEmfnprqik=
golang.org/x/sys v0.31.0/go.mod h1:BJP2sWEmIv4KK5OTEluFJCKSidICx8ciO85XgH3Ak8k=
golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1/go.mod h1:bj7SfCRtBDWHUb9snDiAeCFNEtKQo2Wmx5Cou7ajbmo=
golang.org/x/term v0.0.0-20210927222741-03fcf44c2211/go.mod h1:jbD1KX2456YbFQfuXm/mYQcufACuNUgVhRMnK/tPxf8=
golang.org/x/term v0.5.0/go.mod h1:jMB1sMXY+tzblOD4FWmEbocvup2/aLOaQEp7JmGp78k=
golang.org/x/text v0.3.0/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
golang.org/x/text v0.3.3/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
golang.org/x/text v0.3.7/go.mod h1:u+2+/6zg+i71rQMx5EYifcz6MCKuco9NR6JIITiCfzQ=
golang.org/x/text v0.7.0/go.mod h1:mrYo+phRRbMaCq/xk9113O4dZlRixOauAjOtrjsXDZ8=
golang.org/x/text v0.23.0 h1:D71I7dUrlY+VX0gQShAThNGHFxZ13dGLBHQLVl1mJlY=
golang.org/x/text v0.23.0/go.mod h1:/BLNzu4aZCJ1+kcD0DNRotWKage4q2rGVAg4o22unh4=
golang.org/x/time v0.11.0 h1:/bpjEDfN9tkoN/ryeYHnv5hcMlc8ncjMcM4XBk5NWV0=
golang.org/x/time v0.11.0/go.mod h1:CDIdPxbZBQxdj6cxyCIdrNogrJKMJ7pr37NYpMcMDSg=
golang.org/x/tools v0.0.0-20180917221912-90fa682c2a6e/go.mod h1:n7NCudcB/nEzxVGmLbDWY5pfWTLqBcC2KZ6jyYvM4mQ=
golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
golang.org/x/tools v0.1.12/go.mod h1:hNGJHUnrk76NpqgfD5Aqm5Crs+Hm0VOH/i9J2+nxYbc=
golang.org/x/tools v0.31.0 h1:0EedkvKDbh+qistFTd0Bcwe/YLh4vHwWEkiI0toFIBU=
golang.org/x/tools v0.31.0/go.mod h1:naFTU+Cev749tSJRXJlna0T3WxKvb1kWEx15xA4SdmQ=
golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
google.golang.org/api v0.228.0 h1:X2DJ/uoWGnY5obVjewbp8icSL5U4FzuCfy9OjbLSnLs=
google.golang.org/api v0.228.0/go.mod h1:wNvRS1Pbe8r4+IfBIniV8fwCpGwTrYa+kMUDiC5z5a4=
google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 h1:iK2jbkWL86DXjEx0qiHcRE9dE4/Ahua5k6V8OWFb//c=
google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4/go.mod h1:LuRYeWDFV6WOn90g357N17oMCaxpgCnbi/44qJvDn2I=
google.golang.org/grpc v1.71.0 h1:kF77BGdPTQ4/JZWMlb9VpJ5pa25aqvVqogsxNHHdeBg=
google.golang.org/grpc v1.71.0/go.mod h1:H0GRtasmQOh9LkFoCPDu3ZrwUtD1YGE+b2vYBYd/8Ec=
google.golang.org/protobuf v1.36.6 h1:z1NpPI8ku2WgiWnf+t9wTPsn6eP1L7ksHUlkfLvd9xY=
google.golang.org/protobuf v1.36.6/go.mod h1:jduwjTPXsFjZGTmRluh+L6NjiWu7pchiJ2/5YcXBHnY=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=
gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c/go.mod h1:JHkPIbrfpd72SG/EVd6muEfDQjcINNoR0C8j2r3qZ4Q=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
```

## `config.development.yaml`

```yaml
# config.development.yaml
# 开发环境配置文件
# 值可以通过环境变量覆盖（例如，SERVER_PORT=9000）。

server:
  port: "8080"
  readTimeout: 5s
  writeTimeout: 10s
  idleTimeout: 120s

database:
  # 开发环境PostgreSQL连接字符串
  # 使用与 Docker 容器匹配的凭据 (user/password)
  dsn: "postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable"
  maxOpenConns: 25
  maxIdleConns: 25
  connMaxLifetime: 5m
  connMaxIdleTime: 5m

jwt:
  # 开发环境JWT密钥 - 仅用于开发，生产环境请使用环境变量
  secretKey: "development-jwt-secret-key-for-testing-purposes-only"
  accessTokenExpiry: 1h
  # refreshTokenExpiry: 720h # ~30 days

minio:
  # MinIO对象存储配置
  endpoint: "localhost:9000"
  accessKeyId: "minioadmin"
  secretAccessKey: "minioadmin"
  useSsl: false
  bucketName: "language-audio"
  presignExpiry: 1h

google:
  # 开发环境Google OAuth配置 - 仅用于开发
  clientId: "development-google-client-id.apps.googleusercontent.com"
  clientSecret: "DEVELOPMENT_GOOGLE_CLIENT_SECRET"

log:
  level: "debug" # 开发环境使用debug级别
  json: false    # 开发环境使用易读的非JSON格式

cors:
  # 开发环境CORS配置，允许本地前端服务器
  allowedOrigins: ["http://localhost:3000", "http://127.0.0.1:3000"]
  allowedMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowedHeaders: ["Accept", "Authorization", "Content-Type", "X-CSRF-Token"]
  allowCredentials: true
  maxAge: 300
```

## `.gitignore`

```
*.o
*.a
*.so
*.exe
*.out
*.test
*.log
config.*.yaml
!config.example.yaml
!config.development.yaml
.env
vendor/
build/
```

## `config.example.yaml`

```yaml
# config.example.yaml
# Example configuration file. Copy this to config.<env>.yaml (e.g., config.dev.yaml)
# and fill in the actual values. DO NOT COMMIT sensitive data.
# Values can also be overridden by environment variables (e.g., SERVER_PORT=9000).

server:
  port: "8080"
  readTimeout: 5s
  writeTimeout: 10s
  idleTimeout: 120s

database:
  # Example DSN for PostgreSQL. Replace with your actual connection string.
  # Use environment variable DATABASE_DSN for production.
  dsn: "postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable"
  maxOpenConns: 25
  maxIdleConns: 25
  connMaxLifetime: 5m
  connMaxIdleTime: 5m

jwt:
  # Use environment variable JWT_SECRETKEY for production.
  # Generate a strong secret key (e.g., using openssl rand -base64 32)
  secretKey: "your-very-strong-and-secret-jwt-key" # CHANGE THIS!
  accessTokenExpiry: 1h
  # refreshTokenExpiry: 720h # ~30 days

minio:
  # Use environment variables MINIO_ENDPOINT, MINIO_ACCESSKEYID, MINIO_SECRETACCESSKEY for production.
  endpoint: "localhost:9000" # Your MinIO server endpoint
  accessKeyId: "minioadmin" # Your MinIO access key
  secretAccessKey: "minioadmin" # Your MinIO secret key
  useSsl: false # Set to true if MinIO uses HTTPS
  bucketName: "language-audio" # Name of the bucket to store audio files
  presignExpiry: 1h # Default expiry for presigned URLs

google:
  # Use environment variables GOOGLE_CLIENTID, GOOGLE_CLIENTSECRET for production.
  clientId: "your-google-client-id.apps.googleusercontent.com" # CHANGE THIS!
  clientSecret: "YOUR_GOOGLE_CLIENT_SECRET" # CHANGE THIS!

log:
  level: "debug" # Set to "info" or "warn" for production
  json: false   # Set to true for production logging

cors:
  # For development, allowing localhost is common. Adjust for your frontend URL.
  # Use environment variable CORS_ALLOWEDORIGINS="http://your-frontend.com,https://your-frontend.com" for production.
  allowedOrigins: ["http://localhost:3000", "http://127.0.0.1:3000"] # Example for local React dev server
  allowedMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowedHeaders: ["Accept", "Authorization", "Content-Type", "X-CSRF-Token"]
  allowCredentials: true
  maxAge: 300
```

## `.env.example`

```
# .env (Example - DO NOT COMMIT SENSITIVE DATA)
DATABASE_DSN=postgresql://user:password@host.docker.internal:5432/language_learner_db?sslmode=disable
JWT_SECRETKEY=your-very-strong-and-secret-jwt-key-from-make-run-env
MINIO_ENDPOINT=host.docker.internal:9000
MINIO_ACCESSKEYID=minioadmin
MINIO_SECRETACCESSKEY=minioadmin
MINIO_BUCKETNAME=language-audio
GOOGLE_CLIENTID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENTSECRET=YOUR_GOOGLE_CLIENT_SECRET
LOG_LEVEL=debug
LOG_JSON=false
# Add other necessary env vars like CORS_ALLOWEDORIGINS etc.
```

## `cmd/api/main.go`

```go
// ============================================
// FILE: cmd/api/main.go (MODIFIED)
// ============================================
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/time/rate"

	_ "github.com/yvanyang/language-learning-player-api/docs"                                    // Keep this - Import generated docs
	httpadapter "github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http" // Alias for http handler package
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"  // Adjust import path
	repo "github.com/yvanyang/language-learning-player-api/internal/adapter/repository/postgres" // Alias for postgres repo package
	googleauthadapter "github.com/yvanyang/language-learning-player-api/internal/adapter/service/google_auth"
	minioadapter "github.com/yvanyang/language-learning-player-api/internal/adapter/service/minio"
	"github.com/yvanyang/language-learning-player-api/internal/config"     // Adjust import path
	uc "github.com/yvanyang/language-learning-player-api/internal/usecase" // Alias usecase package if needed elsewhere
	"github.com/yvanyang/language-learning-player-api/pkg/logger"          // Adjust import path
	"github.com/yvanyang/language-learning-player-api/pkg/security"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// @title Language Learning Audio Player API
// @version 1.0.0
// @description API specification for the backend of the Language Learning Audio Player application. Provides endpoints for user authentication, audio content management, and user activity tracking.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @tag.name Authentication
// @tag.description Operations related to user signup, login, and external authentication (e.g., Google).
// @tag.name Users
// @tag.description Operations related to user profiles.
// @tag.name Audio Tracks
// @tag.description Operations related to individual audio tracks, including retrieval and listing.
// @tag.name Audio Collections
// @tag.description Operations related to managing audio collections (playlists, courses).
// @tag.name User Activity
// @tag.description Operations related to tracking user interactions like playback progress and bookmarks.
// @tag.name Uploads
// @tag.description Operations related to requesting upload URLs and finalizing uploads.
// @tag.name Health
// @tag.description API health checks.

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Configuration ---
	cfg, err := config.LoadConfig(".") // Pass full config struct
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// --- Logger ---
	appLogger := logger.NewLogger(cfg.Log)
	slog.SetDefault(appLogger)
	appLogger.Info("Configuration loaded")

	// --- Dependency Initialization ---
	appLogger.Info("Initializing dependencies...")

	// Database Pool
	dbPool, err := repo.NewPgxPool(ctx, cfg.Database, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize database connection pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close() // Ensure pool is closed on shutdown

	// Initialize Transaction Manager
	txManager := repo.NewTxManager(dbPool, appLogger)

	// Repositories
	userRepo := repo.NewUserRepository(dbPool, appLogger)
	trackRepo := repo.NewAudioTrackRepository(dbPool, appLogger)
	collectionRepo := repo.NewAudioCollectionRepository(dbPool, appLogger)
	progressRepo := repo.NewPlaybackProgressRepository(dbPool, appLogger)
	bookmarkRepo := repo.NewBookmarkRepository(dbPool, appLogger)
	refreshTokenRepo := repo.NewRefreshTokenRepository(dbPool, appLogger) // ADDED

	// Services / Helpers
	secHelper, err := security.NewSecurity(cfg.JWT.SecretKey, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize security helper", "error", err)
		os.Exit(1)
	}
	storageService, err := minioadapter.NewMinioStorageService(cfg.Minio, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize MinIO storage service", "error", err)
		os.Exit(1)
	}
	googleAuthService, err := googleauthadapter.NewGoogleAuthService(cfg.Google.ClientID, appLogger)
	if err != nil {
		// Optional: Log warning instead of exiting if Google Auth isn't critical for basic functionality
		appLogger.Warn("Failed to initialize Google Auth service (Google login disabled?)", "error", err)
	}

	// Validator
	validator := validation.New()

	// Inject dependencies into Use Cases
	authUseCase := uc.NewAuthUseCase(cfg.JWT, userRepo, refreshTokenRepo, secHelper, googleAuthService, appLogger)
	audioUseCase := uc.NewAudioContentUseCase(
		cfg, trackRepo, collectionRepo, storageService, txManager, progressRepo, bookmarkRepo, appLogger,
	)
	activityUseCase := uc.NewUserActivityUseCase(progressRepo, bookmarkRepo, trackRepo, appLogger)
	uploadUseCase := uc.NewUploadUseCase(cfg.Minio, trackRepo, storageService, txManager, appLogger)
	userUseCase := uc.NewUserUseCase(userRepo, appLogger)

	// HTTP Handlers
	authHandler := httpadapter.NewAuthHandler(authUseCase, validator)
	audioHandler := httpadapter.NewAudioHandler(audioUseCase, validator)
	activityHandler := httpadapter.NewUserActivityHandler(activityUseCase, validator)
	uploadHandler := httpadapter.NewUploadHandler(uploadUseCase, validator)
	userHandler := httpadapter.NewUserHandler(userUseCase)

	appLogger.Info("Dependencies initialized successfully")

	// --- HTTP Router Setup ---
	appLogger.Info("Setting up HTTP router...")
	router := chi.NewRouter()

	// --- Global Middleware Setup (Applied to ALL routes) ---
	router.Use(middleware.RequestID)
	router.Use(middleware.RequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(chimiddleware.RealIP)
	ipLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20) // TODO: Make configurable
	router.Use(middleware.RateLimit(ipLimiter))
	router.Use(chimiddleware.StripSlashes)
	router.Use(chimiddleware.Timeout(60 * time.Second))

	// --- CORS Middleware (Applied globally before specific route groups) ---
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))

	// --- Routes ---

	// Health Check (Outside specific security groups)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Swagger Docs Group - Apply relaxed security headers
	router.Group(func(r chi.Router) {
		r.Use(middleware.SwaggerSecurityHeaders)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
		})
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})

	// API v1 Routes Group - Apply strict security headers and authentication
	router.Route("/api/v1", func(r chi.Router) {
		// Apply API-specific security headers to all /api/v1 routes
		r.Use(middleware.ApiSecurityHeaders)

		// Public API routes (Still get ApiSecurityHeaders)
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/google/callback", authHandler.GoogleCallback)
			r.Post("/auth/refresh", authHandler.Refresh) // ADDED Refresh route
			r.Post("/auth/logout", authHandler.Logout)   // ADDED Logout route

			r.Get("/audio/tracks", audioHandler.ListTracks)
			r.Get("/audio/tracks/{trackId}", audioHandler.GetTrackDetails)
			r.Get("/audio/collections/{collectionId}", audioHandler.GetCollectionDetails)
		})

		// Protected API routes (Apply Authenticator middleware + ApiSecurityHeaders)
		r.Group(func(r chi.Router) {
			// IMPORTANT: Authenticator middleware ONLY checks the ACCESS token.
			// Refresh/Logout routes should NOT be in this group.
			r.Use(middleware.Authenticator(secHelper))

			// User Profile
			r.Get("/users/me", userHandler.GetMyProfile)

			// Audio Collections
			r.Post("/audio/collections", audioHandler.CreateCollection)
			r.Put("/audio/collections/{collectionId}", audioHandler.UpdateCollectionMetadata)
			r.Delete("/audio/collections/{collectionId}", audioHandler.DeleteCollection)
			r.Put("/audio/collections/{collectionId}/tracks", audioHandler.UpdateCollectionTracks)

			// User Activity
			r.Post("/users/me/progress", activityHandler.RecordProgress)
			r.Get("/users/me/progress", activityHandler.ListProgress)
			r.Get("/users/me/progress/{trackId}", activityHandler.GetProgress)
			r.Post("/users/me/bookmarks", activityHandler.CreateBookmark)
			r.Get("/users/me/bookmarks", activityHandler.ListBookmarks)
			r.Delete("/users/me/bookmarks/{bookmarkId}", activityHandler.DeleteBookmark)

			// Upload Routes (Require Auth)
			// Single File
			r.Post("/uploads/audio/request", uploadHandler.RequestUpload)
			r.Post("/audio/tracks", uploadHandler.CompleteUploadAndCreateTrack)
			// Batch Files
			r.Post("/uploads/audio/batch/request", uploadHandler.RequestBatchUpload)                 // ADDED
			r.Post("/audio/tracks/batch/complete", uploadHandler.CompleteBatchUploadAndCreateTracks) // ADDED
		})
	})

	appLogger.Info("HTTP router setup complete")

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError),
	}

	// --- Start Server & Graceful Shutdown ---
	go func() {
		appLogger.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	appLogger.Info("Closing database connection pool...")
	dbPool.Close()
	appLogger.Info("Database connection pool closed.")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server shutdown complete.")
}
```

## `migrations/000004_create_refresh_tokens.down.sql`

```sql
-- migrations/000004_create_refresh_tokens.down.sql

DROP TABLE IF EXISTS refresh_tokens;
```

## `migrations/000003_create_activity_tables.down.sql`

```sql
-- migrations/000003_create_activity_tables.down.sql

DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS playback_progress;
```

## `migrations/000001_create_users_table.down.sql`

```sql
-- migrations/000001_create_users_table.down.sql

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;

-- Optional: Drop the extension if it's certain no other table needs it.
-- Usually, it's safe to leave it enabled.
-- DROP EXTENSION IF EXISTS "pgcrypto";
```

## `migrations/000002_create_audio_tables.down.sql`

```sql
-- migrations/000002_create_audio_tables.down.sql

DROP TRIGGER IF EXISTS update_audio_collections_updated_at ON audio_collections;
DROP TRIGGER IF EXISTS update_audio_tracks_updated_at ON audio_tracks;

DROP TABLE IF EXISTS collection_tracks;
DROP TABLE IF EXISTS audio_collections;
DROP TABLE IF EXISTS audio_tracks;
```

## `migrations/000003_create_activity_tables.up.sql`

```sql
-- migrations/000003_create_activity_tables.up.sql

-- Playback Progress Table
CREATE TABLE playback_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    -- Progress stored in MILLISECONDS as a BIGINT
    progress_ms BIGINT NOT NULL DEFAULT 0 CHECK (progress_ms >= 0),
    last_listened_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Composite primary key ensures one progress record per user/track pair
    PRIMARY KEY (user_id, track_id)
);

-- Index for quickly finding recent progress for a user
CREATE INDEX idx_playbackprogress_user_lastlistened ON playback_progress(user_id, last_listened_at DESC);


-- Bookmarks Table
CREATE TABLE bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    -- Timestamp stored in MILLISECONDS as a BIGINT
    timestamp_ms BIGINT NOT NULL CHECK (timestamp_ms >= 0),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- No updated_at needed if bookmarks are immutable once created (except for deletion)
);

-- Index for efficient listing of bookmarks for a user on a specific track, ordered by time
CREATE INDEX idx_bookmarks_user_track_time ON bookmarks(user_id, track_id, timestamp_ms ASC);
-- Index for listing recent bookmarks for a user across all tracks
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC);

-- No triggers needed here assuming upsert handles last_listened_at
-- and created_at uses default.
```

## `migrations/000002_create_audio_tables.up.sql`

```sql
-- migrations/000002_create_audio_tables.up.sql

-- Audio Tracks Table
CREATE TABLE audio_tracks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    language_code VARCHAR(10) NOT NULL, -- e.g., 'en-US', 'zh-CN'
    level VARCHAR(50),                  -- e.g., 'A1', 'B2', 'NATIVE' (Matches domain.AudioLevel)
    duration_ms BIGINT NOT NULL DEFAULT 0 CHECK (duration_ms >= 0), -- Store duration in BIGINT MILLISECONDS
    minio_bucket VARCHAR(100) NOT NULL,
    minio_object_key VARCHAR(1024) NOT NULL UNIQUE, -- Object key should be unique within bucket
    cover_image_url VARCHAR(1024),
    uploader_id UUID NULL REFERENCES users(id) ON DELETE SET NULL, -- Optional link to user
    is_public BOOLEAN NOT NULL DEFAULT true,
    tags TEXT[] NULL,                   -- Array of text tags
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for audio_tracks
CREATE INDEX idx_audiotracks_language ON audio_tracks(language_code);
CREATE INDEX idx_audiotracks_level ON audio_tracks(level);
CREATE INDEX idx_audiotracks_uploader ON audio_tracks(uploader_id);
CREATE INDEX idx_audiotracks_is_public ON audio_tracks(is_public);
CREATE INDEX idx_audiotracks_tags ON audio_tracks USING GIN (tags); -- GIN index for array searching
CREATE INDEX idx_audiotracks_created_at ON audio_tracks(created_at DESC);

-- Audio Collections Table
CREATE TABLE audio_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Collections deleted if owner is deleted
    type VARCHAR(50) NOT NULL CHECK (type IN ('COURSE', 'PLAYLIST')), -- Matches domain.CollectionType
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for audio_collections
CREATE INDEX idx_audiocollections_owner ON audio_collections(owner_id);
CREATE INDEX idx_audiocollections_type ON audio_collections(type);

-- Collection Tracks Association Table (Many-to-Many with Ordering)
CREATE TABLE collection_tracks (
    collection_id UUID NOT NULL REFERENCES audio_collections(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE, -- If track deleted, remove from collection
    position INTEGER NOT NULL DEFAULT 0 CHECK (position >= 0), -- Order within the collection
    PRIMARY KEY (collection_id, track_id) -- Ensure a track is only added once per collection
);

-- Index for finding collections a track belongs to
CREATE INDEX idx_collectiontracks_track_id ON collection_tracks(track_id);
-- Index for retrieving tracks in order for a collection
CREATE INDEX idx_collectiontracks_order ON collection_tracks(collection_id, position);


-- Add triggers to automatically update updated_at timestamps for new tables
-- Ensure the function update_updated_at_column exists from migration 000001
CREATE TRIGGER update_audio_tracks_updated_at
BEFORE UPDATE ON audio_tracks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_audio_collections_updated_at
BEFORE UPDATE ON audio_collections
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
```

## `migrations/000004_create_refresh_tokens.up.sql`

```sql
-- migrations/000004_create_refresh_tokens.up.sql

CREATE TABLE refresh_tokens (
    -- id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Use hash as PK? No, better to have separate ID
    -- Store the SHA-256 hash of the refresh token value
    -- Use TEXT as hash length is fixed (64 hex chars) but TEXT is simpler. Or VARCHAR(64).
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- revoked_at TIMESTAMPTZ NULL -- Optional: Add if explicit revocation tracking is needed besides deletion
);

-- Index for faster lookup by user_id (e.g., for deleting all tokens for a user)
CREATE INDEX idx_refreshtokens_user_id ON refresh_tokens(user_id);

-- Index on expires_at for potential cleanup jobs
CREATE INDEX idx_refreshtokens_expires_at ON refresh_tokens(expires_at);
```

## `migrations/000001_create_users_table.up.sql`

```sql
-- migrations/000001_create_users_table.up.sql

-- Enable UUID generation if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the users table based on domain model and Section 7 DDL
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    password_hash VARCHAR(255) NULL, -- Nullable for external auth providers
    name VARCHAR(100),
    google_id VARCHAR(255) UNIQUE NULL,
    auth_provider VARCHAR(50) NOT NULL DEFAULT 'local', -- 'local', 'google' etc.
    profile_image_url VARCHAR(1024) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add indexes for faster lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- Optional: Add a trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
```

## `internal/mocks/port/mock_FileStorageService.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"
	"time"

	mock "github.com/stretchr/testify/mock"
)

// NewMockFileStorageService creates a new instance of MockFileStorageService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFileStorageService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFileStorageService {
	mock := &MockFileStorageService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockFileStorageService is an autogenerated mock type for the FileStorageService type
type MockFileStorageService struct {
	mock.Mock
}

type MockFileStorageService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFileStorageService) EXPECT() *MockFileStorageService_Expecter {
	return &MockFileStorageService_Expecter{mock: &_m.Mock}
}

// DeleteObject provides a mock function for the type MockFileStorageService
func (_mock *MockFileStorageService) DeleteObject(ctx context.Context, bucket string, objectKey string) error {
	ret := _mock.Called(ctx, bucket, objectKey)

	if len(ret) == 0 {
		panic("no return value specified for DeleteObject")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = returnFunc(ctx, bucket, objectKey)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockFileStorageService_DeleteObject_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteObject'
type MockFileStorageService_DeleteObject_Call struct {
	*mock.Call
}

// DeleteObject is a helper method to define mock.On call
//   - ctx
//   - bucket
//   - objectKey
func (_e *MockFileStorageService_Expecter) DeleteObject(ctx interface{}, bucket interface{}, objectKey interface{}) *MockFileStorageService_DeleteObject_Call {
	return &MockFileStorageService_DeleteObject_Call{Call: _e.mock.On("DeleteObject", ctx, bucket, objectKey)}
}

func (_c *MockFileStorageService_DeleteObject_Call) Run(run func(ctx context.Context, bucket string, objectKey string)) *MockFileStorageService_DeleteObject_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockFileStorageService_DeleteObject_Call) Return(err error) *MockFileStorageService_DeleteObject_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockFileStorageService_DeleteObject_Call) RunAndReturn(run func(ctx context.Context, bucket string, objectKey string) error) *MockFileStorageService_DeleteObject_Call {
	_c.Call.Return(run)
	return _c
}

// GetPresignedGetURL provides a mock function for the type MockFileStorageService
func (_mock *MockFileStorageService) GetPresignedGetURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error) {
	ret := _mock.Called(ctx, bucket, objectKey, expiry)

	if len(ret) == 0 {
		panic("no return value specified for GetPresignedGetURL")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) (string, error)); ok {
		return returnFunc(ctx, bucket, objectKey, expiry)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) string); ok {
		r0 = returnFunc(ctx, bucket, objectKey, expiry)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, time.Duration) error); ok {
		r1 = returnFunc(ctx, bucket, objectKey, expiry)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockFileStorageService_GetPresignedGetURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPresignedGetURL'
type MockFileStorageService_GetPresignedGetURL_Call struct {
	*mock.Call
}

// GetPresignedGetURL is a helper method to define mock.On call
//   - ctx
//   - bucket
//   - objectKey
//   - expiry
func (_e *MockFileStorageService_Expecter) GetPresignedGetURL(ctx interface{}, bucket interface{}, objectKey interface{}, expiry interface{}) *MockFileStorageService_GetPresignedGetURL_Call {
	return &MockFileStorageService_GetPresignedGetURL_Call{Call: _e.mock.On("GetPresignedGetURL", ctx, bucket, objectKey, expiry)}
}

func (_c *MockFileStorageService_GetPresignedGetURL_Call) Run(run func(ctx context.Context, bucket string, objectKey string, expiry time.Duration)) *MockFileStorageService_GetPresignedGetURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(time.Duration))
	})
	return _c
}

func (_c *MockFileStorageService_GetPresignedGetURL_Call) Return(s string, err error) *MockFileStorageService_GetPresignedGetURL_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockFileStorageService_GetPresignedGetURL_Call) RunAndReturn(run func(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error)) *MockFileStorageService_GetPresignedGetURL_Call {
	_c.Call.Return(run)
	return _c
}

// GetPresignedPutURL provides a mock function for the type MockFileStorageService
func (_mock *MockFileStorageService) GetPresignedPutURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error) {
	ret := _mock.Called(ctx, bucket, objectKey, expiry)

	if len(ret) == 0 {
		panic("no return value specified for GetPresignedPutURL")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) (string, error)); ok {
		return returnFunc(ctx, bucket, objectKey, expiry)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) string); ok {
		r0 = returnFunc(ctx, bucket, objectKey, expiry)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, time.Duration) error); ok {
		r1 = returnFunc(ctx, bucket, objectKey, expiry)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockFileStorageService_GetPresignedPutURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPresignedPutURL'
type MockFileStorageService_GetPresignedPutURL_Call struct {
	*mock.Call
}

// GetPresignedPutURL is a helper method to define mock.On call
//   - ctx
//   - bucket
//   - objectKey
//   - expiry
func (_e *MockFileStorageService_Expecter) GetPresignedPutURL(ctx interface{}, bucket interface{}, objectKey interface{}, expiry interface{}) *MockFileStorageService_GetPresignedPutURL_Call {
	return &MockFileStorageService_GetPresignedPutURL_Call{Call: _e.mock.On("GetPresignedPutURL", ctx, bucket, objectKey, expiry)}
}

func (_c *MockFileStorageService_GetPresignedPutURL_Call) Run(run func(ctx context.Context, bucket string, objectKey string, expiry time.Duration)) *MockFileStorageService_GetPresignedPutURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(time.Duration))
	})
	return _c
}

func (_c *MockFileStorageService_GetPresignedPutURL_Call) Return(s string, err error) *MockFileStorageService_GetPresignedPutURL_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockFileStorageService_GetPresignedPutURL_Call) RunAndReturn(run func(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error)) *MockFileStorageService_GetPresignedPutURL_Call {
	_c.Call.Return(run)
	return _c
}

// ObjectExists provides a mock function for the type MockFileStorageService
func (_mock *MockFileStorageService) ObjectExists(ctx context.Context, bucket string, objectKey string) (bool, error) {
	ret := _mock.Called(ctx, bucket, objectKey)

	if len(ret) == 0 {
		panic("no return value specified for ObjectExists")
	}

	var r0 bool
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return returnFunc(ctx, bucket, objectKey)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = returnFunc(ctx, bucket, objectKey)
	} else {
		r0 = ret.Get(0).(bool)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = returnFunc(ctx, bucket, objectKey)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockFileStorageService_ObjectExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ObjectExists'
type MockFileStorageService_ObjectExists_Call struct {
	*mock.Call
}

// ObjectExists is a helper method to define mock.On call
//   - ctx
//   - bucket
//   - objectKey
func (_e *MockFileStorageService_Expecter) ObjectExists(ctx interface{}, bucket interface{}, objectKey interface{}) *MockFileStorageService_ObjectExists_Call {
	return &MockFileStorageService_ObjectExists_Call{Call: _e.mock.On("ObjectExists", ctx, bucket, objectKey)}
}

func (_c *MockFileStorageService_ObjectExists_Call) Run(run func(ctx context.Context, bucket string, objectKey string)) *MockFileStorageService_ObjectExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockFileStorageService_ObjectExists_Call) Return(b bool, err error) *MockFileStorageService_ObjectExists_Call {
	_c.Call.Return(b, err)
	return _c
}

func (_c *MockFileStorageService_ObjectExists_Call) RunAndReturn(run func(ctx context.Context, bucket string, objectKey string) (bool, error)) *MockFileStorageService_ObjectExists_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_UploadUseCase.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// NewMockUploadUseCase creates a new instance of MockUploadUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUploadUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUploadUseCase {
	mock := &MockUploadUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockUploadUseCase is an autogenerated mock type for the UploadUseCase type
type MockUploadUseCase struct {
	mock.Mock
}

type MockUploadUseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUploadUseCase) EXPECT() *MockUploadUseCase_Expecter {
	return &MockUploadUseCase_Expecter{mock: &_m.Mock}
}

// CompleteBatchUpload provides a mock function for the type MockUploadUseCase
func (_mock *MockUploadUseCase) CompleteBatchUpload(ctx context.Context, userID domain.UserID, req port.BatchCompleteInput) ([]port.BatchCompleteResultItem, error) {
	ret := _mock.Called(ctx, userID, req)

	if len(ret) == 0 {
		panic("no return value specified for CompleteBatchUpload")
	}

	var r0 []port.BatchCompleteResultItem
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, port.BatchCompleteInput) ([]port.BatchCompleteResultItem, error)); ok {
		return returnFunc(ctx, userID, req)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, port.BatchCompleteInput) []port.BatchCompleteResultItem); ok {
		r0 = returnFunc(ctx, userID, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]port.BatchCompleteResultItem)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, port.BatchCompleteInput) error); ok {
		r1 = returnFunc(ctx, userID, req)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUploadUseCase_CompleteBatchUpload_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompleteBatchUpload'
type MockUploadUseCase_CompleteBatchUpload_Call struct {
	*mock.Call
}

// CompleteBatchUpload is a helper method to define mock.On call
//   - ctx
//   - userID
//   - req
func (_e *MockUploadUseCase_Expecter) CompleteBatchUpload(ctx interface{}, userID interface{}, req interface{}) *MockUploadUseCase_CompleteBatchUpload_Call {
	return &MockUploadUseCase_CompleteBatchUpload_Call{Call: _e.mock.On("CompleteBatchUpload", ctx, userID, req)}
}

func (_c *MockUploadUseCase_CompleteBatchUpload_Call) Run(run func(ctx context.Context, userID domain.UserID, req port.BatchCompleteInput)) *MockUploadUseCase_CompleteBatchUpload_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(port.BatchCompleteInput))
	})
	return _c
}

func (_c *MockUploadUseCase_CompleteBatchUpload_Call) Return(batchCompleteResultItems []port.BatchCompleteResultItem, err error) *MockUploadUseCase_CompleteBatchUpload_Call {
	_c.Call.Return(batchCompleteResultItems, err)
	return _c
}

func (_c *MockUploadUseCase_CompleteBatchUpload_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, req port.BatchCompleteInput) ([]port.BatchCompleteResultItem, error)) *MockUploadUseCase_CompleteBatchUpload_Call {
	_c.Call.Return(run)
	return _c
}

// CompleteUpload provides a mock function for the type MockUploadUseCase
func (_mock *MockUploadUseCase) CompleteUpload(ctx context.Context, userID domain.UserID, req port.CompleteUploadInput) (*domain.AudioTrack, error) {
	ret := _mock.Called(ctx, userID, req)

	if len(ret) == 0 {
		panic("no return value specified for CompleteUpload")
	}

	var r0 *domain.AudioTrack
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, port.CompleteUploadInput) (*domain.AudioTrack, error)); ok {
		return returnFunc(ctx, userID, req)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, port.CompleteUploadInput) *domain.AudioTrack); ok {
		r0 = returnFunc(ctx, userID, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.AudioTrack)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, port.CompleteUploadInput) error); ok {
		r1 = returnFunc(ctx, userID, req)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUploadUseCase_CompleteUpload_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompleteUpload'
type MockUploadUseCase_CompleteUpload_Call struct {
	*mock.Call
}

// CompleteUpload is a helper method to define mock.On call
//   - ctx
//   - userID
//   - req
func (_e *MockUploadUseCase_Expecter) CompleteUpload(ctx interface{}, userID interface{}, req interface{}) *MockUploadUseCase_CompleteUpload_Call {
	return &MockUploadUseCase_CompleteUpload_Call{Call: _e.mock.On("CompleteUpload", ctx, userID, req)}
}

func (_c *MockUploadUseCase_CompleteUpload_Call) Run(run func(ctx context.Context, userID domain.UserID, req port.CompleteUploadInput)) *MockUploadUseCase_CompleteUpload_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(port.CompleteUploadInput))
	})
	return _c
}

func (_c *MockUploadUseCase_CompleteUpload_Call) Return(audioTrack *domain.AudioTrack, err error) *MockUploadUseCase_CompleteUpload_Call {
	_c.Call.Return(audioTrack, err)
	return _c
}

func (_c *MockUploadUseCase_CompleteUpload_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, req port.CompleteUploadInput) (*domain.AudioTrack, error)) *MockUploadUseCase_CompleteUpload_Call {
	_c.Call.Return(run)
	return _c
}

// RequestBatchUpload provides a mock function for the type MockUploadUseCase
func (_mock *MockUploadUseCase) RequestBatchUpload(ctx context.Context, userID domain.UserID, req port.BatchRequestUploadInput) ([]port.BatchURLResultItem, error) {
	ret := _mock.Called(ctx, userID, req)

	if len(ret) == 0 {
		panic("no return value specified for RequestBatchUpload")
	}

	var r0 []port.BatchURLResultItem
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, port.BatchRequestUploadInput) ([]port.BatchURLResultItem, error)); ok {
		return returnFunc(ctx, userID, req)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, port.BatchRequestUploadInput) []port.BatchURLResultItem); ok {
		r0 = returnFunc(ctx, userID, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]port.BatchURLResultItem)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, port.BatchRequestUploadInput) error); ok {
		r1 = returnFunc(ctx, userID, req)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUploadUseCase_RequestBatchUpload_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequestBatchUpload'
type MockUploadUseCase_RequestBatchUpload_Call struct {
	*mock.Call
}

// RequestBatchUpload is a helper method to define mock.On call
//   - ctx
//   - userID
//   - req
func (_e *MockUploadUseCase_Expecter) RequestBatchUpload(ctx interface{}, userID interface{}, req interface{}) *MockUploadUseCase_RequestBatchUpload_Call {
	return &MockUploadUseCase_RequestBatchUpload_Call{Call: _e.mock.On("RequestBatchUpload", ctx, userID, req)}
}

func (_c *MockUploadUseCase_RequestBatchUpload_Call) Run(run func(ctx context.Context, userID domain.UserID, req port.BatchRequestUploadInput)) *MockUploadUseCase_RequestBatchUpload_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(port.BatchRequestUploadInput))
	})
	return _c
}

func (_c *MockUploadUseCase_RequestBatchUpload_Call) Return(batchURLResultItems []port.BatchURLResultItem, err error) *MockUploadUseCase_RequestBatchUpload_Call {
	_c.Call.Return(batchURLResultItems, err)
	return _c
}

func (_c *MockUploadUseCase_RequestBatchUpload_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, req port.BatchRequestUploadInput) ([]port.BatchURLResultItem, error)) *MockUploadUseCase_RequestBatchUpload_Call {
	_c.Call.Return(run)
	return _c
}

// RequestUpload provides a mock function for the type MockUploadUseCase
func (_mock *MockUploadUseCase) RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*port.RequestUploadResult, error) {
	ret := _mock.Called(ctx, userID, filename, contentType)

	if len(ret) == 0 {
		panic("no return value specified for RequestUpload")
	}

	var r0 *port.RequestUploadResult
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, string, string) (*port.RequestUploadResult, error)); ok {
		return returnFunc(ctx, userID, filename, contentType)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, string, string) *port.RequestUploadResult); ok {
		r0 = returnFunc(ctx, userID, filename, contentType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*port.RequestUploadResult)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, string, string) error); ok {
		r1 = returnFunc(ctx, userID, filename, contentType)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUploadUseCase_RequestUpload_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequestUpload'
type MockUploadUseCase_RequestUpload_Call struct {
	*mock.Call
}

// RequestUpload is a helper method to define mock.On call
//   - ctx
//   - userID
//   - filename
//   - contentType
func (_e *MockUploadUseCase_Expecter) RequestUpload(ctx interface{}, userID interface{}, filename interface{}, contentType interface{}) *MockUploadUseCase_RequestUpload_Call {
	return &MockUploadUseCase_RequestUpload_Call{Call: _e.mock.On("RequestUpload", ctx, userID, filename, contentType)}
}

func (_c *MockUploadUseCase_RequestUpload_Call) Run(run func(ctx context.Context, userID domain.UserID, filename string, contentType string)) *MockUploadUseCase_RequestUpload_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *MockUploadUseCase_RequestUpload_Call) Return(requestUploadResult *port.RequestUploadResult, err error) *MockUploadUseCase_RequestUpload_Call {
	_c.Call.Return(requestUploadResult, err)
	return _c
}

func (_c *MockUploadUseCase_RequestUpload_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, filename string, contentType string) (*port.RequestUploadResult, error)) *MockUploadUseCase_RequestUpload_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_RefreshTokenRepository.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// NewMockRefreshTokenRepository creates a new instance of MockRefreshTokenRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRefreshTokenRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRefreshTokenRepository {
	mock := &MockRefreshTokenRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockRefreshTokenRepository is an autogenerated mock type for the RefreshTokenRepository type
type MockRefreshTokenRepository struct {
	mock.Mock
}

type MockRefreshTokenRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRefreshTokenRepository) EXPECT() *MockRefreshTokenRepository_Expecter {
	return &MockRefreshTokenRepository_Expecter{mock: &_m.Mock}
}

// DeleteByTokenHash provides a mock function for the type MockRefreshTokenRepository
func (_mock *MockRefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	ret := _mock.Called(ctx, tokenHash)

	if len(ret) == 0 {
		panic("no return value specified for DeleteByTokenHash")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, tokenHash)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockRefreshTokenRepository_DeleteByTokenHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteByTokenHash'
type MockRefreshTokenRepository_DeleteByTokenHash_Call struct {
	*mock.Call
}

// DeleteByTokenHash is a helper method to define mock.On call
//   - ctx
//   - tokenHash
func (_e *MockRefreshTokenRepository_Expecter) DeleteByTokenHash(ctx interface{}, tokenHash interface{}) *MockRefreshTokenRepository_DeleteByTokenHash_Call {
	return &MockRefreshTokenRepository_DeleteByTokenHash_Call{Call: _e.mock.On("DeleteByTokenHash", ctx, tokenHash)}
}

func (_c *MockRefreshTokenRepository_DeleteByTokenHash_Call) Run(run func(ctx context.Context, tokenHash string)) *MockRefreshTokenRepository_DeleteByTokenHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockRefreshTokenRepository_DeleteByTokenHash_Call) Return(err error) *MockRefreshTokenRepository_DeleteByTokenHash_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockRefreshTokenRepository_DeleteByTokenHash_Call) RunAndReturn(run func(ctx context.Context, tokenHash string) error) *MockRefreshTokenRepository_DeleteByTokenHash_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteByUser provides a mock function for the type MockRefreshTokenRepository
func (_mock *MockRefreshTokenRepository) DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error) {
	ret := _mock.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteByUser")
	}

	var r0 int64
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID) (int64, error)); ok {
		return returnFunc(ctx, userID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID) int64); ok {
		r0 = returnFunc(ctx, userID)
	} else {
		r0 = ret.Get(0).(int64)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID) error); ok {
		r1 = returnFunc(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockRefreshTokenRepository_DeleteByUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteByUser'
type MockRefreshTokenRepository_DeleteByUser_Call struct {
	*mock.Call
}

// DeleteByUser is a helper method to define mock.On call
//   - ctx
//   - userID
func (_e *MockRefreshTokenRepository_Expecter) DeleteByUser(ctx interface{}, userID interface{}) *MockRefreshTokenRepository_DeleteByUser_Call {
	return &MockRefreshTokenRepository_DeleteByUser_Call{Call: _e.mock.On("DeleteByUser", ctx, userID)}
}

func (_c *MockRefreshTokenRepository_DeleteByUser_Call) Run(run func(ctx context.Context, userID domain.UserID)) *MockRefreshTokenRepository_DeleteByUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID))
	})
	return _c
}

func (_c *MockRefreshTokenRepository_DeleteByUser_Call) Return(n int64, err error) *MockRefreshTokenRepository_DeleteByUser_Call {
	_c.Call.Return(n, err)
	return _c
}

func (_c *MockRefreshTokenRepository_DeleteByUser_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID) (int64, error)) *MockRefreshTokenRepository_DeleteByUser_Call {
	_c.Call.Return(run)
	return _c
}

// FindByTokenHash provides a mock function for the type MockRefreshTokenRepository
func (_mock *MockRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*port.RefreshTokenData, error) {
	ret := _mock.Called(ctx, tokenHash)

	if len(ret) == 0 {
		panic("no return value specified for FindByTokenHash")
	}

	var r0 *port.RefreshTokenData
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (*port.RefreshTokenData, error)); ok {
		return returnFunc(ctx, tokenHash)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) *port.RefreshTokenData); ok {
		r0 = returnFunc(ctx, tokenHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*port.RefreshTokenData)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, tokenHash)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockRefreshTokenRepository_FindByTokenHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByTokenHash'
type MockRefreshTokenRepository_FindByTokenHash_Call struct {
	*mock.Call
}

// FindByTokenHash is a helper method to define mock.On call
//   - ctx
//   - tokenHash
func (_e *MockRefreshTokenRepository_Expecter) FindByTokenHash(ctx interface{}, tokenHash interface{}) *MockRefreshTokenRepository_FindByTokenHash_Call {
	return &MockRefreshTokenRepository_FindByTokenHash_Call{Call: _e.mock.On("FindByTokenHash", ctx, tokenHash)}
}

func (_c *MockRefreshTokenRepository_FindByTokenHash_Call) Run(run func(ctx context.Context, tokenHash string)) *MockRefreshTokenRepository_FindByTokenHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockRefreshTokenRepository_FindByTokenHash_Call) Return(refreshTokenData *port.RefreshTokenData, err error) *MockRefreshTokenRepository_FindByTokenHash_Call {
	_c.Call.Return(refreshTokenData, err)
	return _c
}

func (_c *MockRefreshTokenRepository_FindByTokenHash_Call) RunAndReturn(run func(ctx context.Context, tokenHash string) (*port.RefreshTokenData, error)) *MockRefreshTokenRepository_FindByTokenHash_Call {
	_c.Call.Return(run)
	return _c
}

// Save provides a mock function for the type MockRefreshTokenRepository
func (_mock *MockRefreshTokenRepository) Save(ctx context.Context, tokenData *port.RefreshTokenData) error {
	ret := _mock.Called(ctx, tokenData)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *port.RefreshTokenData) error); ok {
		r0 = returnFunc(ctx, tokenData)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockRefreshTokenRepository_Save_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Save'
type MockRefreshTokenRepository_Save_Call struct {
	*mock.Call
}

// Save is a helper method to define mock.On call
//   - ctx
//   - tokenData
func (_e *MockRefreshTokenRepository_Expecter) Save(ctx interface{}, tokenData interface{}) *MockRefreshTokenRepository_Save_Call {
	return &MockRefreshTokenRepository_Save_Call{Call: _e.mock.On("Save", ctx, tokenData)}
}

func (_c *MockRefreshTokenRepository_Save_Call) Run(run func(ctx context.Context, tokenData *port.RefreshTokenData)) *MockRefreshTokenRepository_Save_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*port.RefreshTokenData))
	})
	return _c
}

func (_c *MockRefreshTokenRepository_Save_Call) Return(err error) *MockRefreshTokenRepository_Save_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockRefreshTokenRepository_Save_Call) RunAndReturn(run func(ctx context.Context, tokenData *port.RefreshTokenData) error) *MockRefreshTokenRepository_Save_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_AudioCollectionRepository.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// NewMockAudioCollectionRepository creates a new instance of MockAudioCollectionRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAudioCollectionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAudioCollectionRepository {
	mock := &MockAudioCollectionRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockAudioCollectionRepository is an autogenerated mock type for the AudioCollectionRepository type
type MockAudioCollectionRepository struct {
	mock.Mock
}

type MockAudioCollectionRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAudioCollectionRepository) EXPECT() *MockAudioCollectionRepository_Expecter {
	return &MockAudioCollectionRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) Create(ctx context.Context, collection *domain.AudioCollection) error {
	ret := _mock.Called(ctx, collection)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.AudioCollection) error); ok {
		r0 = returnFunc(ctx, collection)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioCollectionRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockAudioCollectionRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx
//   - collection
func (_e *MockAudioCollectionRepository_Expecter) Create(ctx interface{}, collection interface{}) *MockAudioCollectionRepository_Create_Call {
	return &MockAudioCollectionRepository_Create_Call{Call: _e.mock.On("Create", ctx, collection)}
}

func (_c *MockAudioCollectionRepository_Create_Call) Run(run func(ctx context.Context, collection *domain.AudioCollection)) *MockAudioCollectionRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.AudioCollection))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_Create_Call) Return(err error) *MockAudioCollectionRepository_Create_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioCollectionRepository_Create_Call) RunAndReturn(run func(ctx context.Context, collection *domain.AudioCollection) error) *MockAudioCollectionRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) Delete(ctx context.Context, id domain.CollectionID) error {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) error); ok {
		r0 = returnFunc(ctx, id)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioCollectionRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockAudioCollectionRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockAudioCollectionRepository_Expecter) Delete(ctx interface{}, id interface{}) *MockAudioCollectionRepository_Delete_Call {
	return &MockAudioCollectionRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *MockAudioCollectionRepository_Delete_Call) Run(run func(ctx context.Context, id domain.CollectionID)) *MockAudioCollectionRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_Delete_Call) Return(err error) *MockAudioCollectionRepository_Delete_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioCollectionRepository_Delete_Call) RunAndReturn(run func(ctx context.Context, id domain.CollectionID) error) *MockAudioCollectionRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// FindByID provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for FindByID")
	}

	var r0 *domain.AudioCollection
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) (*domain.AudioCollection, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) *domain.AudioCollection); ok {
		r0 = returnFunc(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.AudioCollection)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.CollectionID) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioCollectionRepository_FindByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByID'
type MockAudioCollectionRepository_FindByID_Call struct {
	*mock.Call
}

// FindByID is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockAudioCollectionRepository_Expecter) FindByID(ctx interface{}, id interface{}) *MockAudioCollectionRepository_FindByID_Call {
	return &MockAudioCollectionRepository_FindByID_Call{Call: _e.mock.On("FindByID", ctx, id)}
}

func (_c *MockAudioCollectionRepository_FindByID_Call) Run(run func(ctx context.Context, id domain.CollectionID)) *MockAudioCollectionRepository_FindByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_FindByID_Call) Return(audioCollection *domain.AudioCollection, err error) *MockAudioCollectionRepository_FindByID_Call {
	_c.Call.Return(audioCollection, err)
	return _c
}

func (_c *MockAudioCollectionRepository_FindByID_Call) RunAndReturn(run func(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)) *MockAudioCollectionRepository_FindByID_Call {
	_c.Call.Return(run)
	return _c
}

// FindWithTracks provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for FindWithTracks")
	}

	var r0 *domain.AudioCollection
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) (*domain.AudioCollection, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) *domain.AudioCollection); ok {
		r0 = returnFunc(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.AudioCollection)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.CollectionID) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioCollectionRepository_FindWithTracks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindWithTracks'
type MockAudioCollectionRepository_FindWithTracks_Call struct {
	*mock.Call
}

// FindWithTracks is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockAudioCollectionRepository_Expecter) FindWithTracks(ctx interface{}, id interface{}) *MockAudioCollectionRepository_FindWithTracks_Call {
	return &MockAudioCollectionRepository_FindWithTracks_Call{Call: _e.mock.On("FindWithTracks", ctx, id)}
}

func (_c *MockAudioCollectionRepository_FindWithTracks_Call) Run(run func(ctx context.Context, id domain.CollectionID)) *MockAudioCollectionRepository_FindWithTracks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_FindWithTracks_Call) Return(audioCollection *domain.AudioCollection, err error) *MockAudioCollectionRepository_FindWithTracks_Call {
	_c.Call.Return(audioCollection, err)
	return _c
}

func (_c *MockAudioCollectionRepository_FindWithTracks_Call) RunAndReturn(run func(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)) *MockAudioCollectionRepository_FindWithTracks_Call {
	_c.Call.Return(run)
	return _c
}

// ListByOwner provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) ([]*domain.AudioCollection, int, error) {
	ret := _mock.Called(ctx, ownerID, page)

	if len(ret) == 0 {
		panic("no return value specified for ListByOwner")
	}

	var r0 []*domain.AudioCollection
	var r1 int
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, pagination.Page) ([]*domain.AudioCollection, int, error)); ok {
		return returnFunc(ctx, ownerID, page)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, pagination.Page) []*domain.AudioCollection); ok {
		r0 = returnFunc(ctx, ownerID, page)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.AudioCollection)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, pagination.Page) int); ok {
		r1 = returnFunc(ctx, ownerID, page)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, domain.UserID, pagination.Page) error); ok {
		r2 = returnFunc(ctx, ownerID, page)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockAudioCollectionRepository_ListByOwner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByOwner'
type MockAudioCollectionRepository_ListByOwner_Call struct {
	*mock.Call
}

// ListByOwner is a helper method to define mock.On call
//   - ctx
//   - ownerID
//   - page
func (_e *MockAudioCollectionRepository_Expecter) ListByOwner(ctx interface{}, ownerID interface{}, page interface{}) *MockAudioCollectionRepository_ListByOwner_Call {
	return &MockAudioCollectionRepository_ListByOwner_Call{Call: _e.mock.On("ListByOwner", ctx, ownerID, page)}
}

func (_c *MockAudioCollectionRepository_ListByOwner_Call) Run(run func(ctx context.Context, ownerID domain.UserID, page pagination.Page)) *MockAudioCollectionRepository_ListByOwner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(pagination.Page))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_ListByOwner_Call) Return(collections []*domain.AudioCollection, total int, err error) *MockAudioCollectionRepository_ListByOwner_Call {
	_c.Call.Return(collections, total, err)
	return _c
}

func (_c *MockAudioCollectionRepository_ListByOwner_Call) RunAndReturn(run func(ctx context.Context, ownerID domain.UserID, page pagination.Page) ([]*domain.AudioCollection, int, error)) *MockAudioCollectionRepository_ListByOwner_Call {
	_c.Call.Return(run)
	return _c
}

// ManageTracks provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	ret := _mock.Called(ctx, collectionID, orderedTrackIDs)

	if len(ret) == 0 {
		panic("no return value specified for ManageTracks")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID, []domain.TrackID) error); ok {
		r0 = returnFunc(ctx, collectionID, orderedTrackIDs)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioCollectionRepository_ManageTracks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ManageTracks'
type MockAudioCollectionRepository_ManageTracks_Call struct {
	*mock.Call
}

// ManageTracks is a helper method to define mock.On call
//   - ctx
//   - collectionID
//   - orderedTrackIDs
func (_e *MockAudioCollectionRepository_Expecter) ManageTracks(ctx interface{}, collectionID interface{}, orderedTrackIDs interface{}) *MockAudioCollectionRepository_ManageTracks_Call {
	return &MockAudioCollectionRepository_ManageTracks_Call{Call: _e.mock.On("ManageTracks", ctx, collectionID, orderedTrackIDs)}
}

func (_c *MockAudioCollectionRepository_ManageTracks_Call) Run(run func(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID)) *MockAudioCollectionRepository_ManageTracks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID), args[2].([]domain.TrackID))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_ManageTracks_Call) Return(err error) *MockAudioCollectionRepository_ManageTracks_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioCollectionRepository_ManageTracks_Call) RunAndReturn(run func(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error) *MockAudioCollectionRepository_ManageTracks_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateMetadata provides a mock function for the type MockAudioCollectionRepository
func (_mock *MockAudioCollectionRepository) UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error {
	ret := _mock.Called(ctx, collection)

	if len(ret) == 0 {
		panic("no return value specified for UpdateMetadata")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.AudioCollection) error); ok {
		r0 = returnFunc(ctx, collection)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioCollectionRepository_UpdateMetadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateMetadata'
type MockAudioCollectionRepository_UpdateMetadata_Call struct {
	*mock.Call
}

// UpdateMetadata is a helper method to define mock.On call
//   - ctx
//   - collection
func (_e *MockAudioCollectionRepository_Expecter) UpdateMetadata(ctx interface{}, collection interface{}) *MockAudioCollectionRepository_UpdateMetadata_Call {
	return &MockAudioCollectionRepository_UpdateMetadata_Call{Call: _e.mock.On("UpdateMetadata", ctx, collection)}
}

func (_c *MockAudioCollectionRepository_UpdateMetadata_Call) Run(run func(ctx context.Context, collection *domain.AudioCollection)) *MockAudioCollectionRepository_UpdateMetadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.AudioCollection))
	})
	return _c
}

func (_c *MockAudioCollectionRepository_UpdateMetadata_Call) Return(err error) *MockAudioCollectionRepository_UpdateMetadata_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioCollectionRepository_UpdateMetadata_Call) RunAndReturn(run func(ctx context.Context, collection *domain.AudioCollection) error) *MockAudioCollectionRepository_UpdateMetadata_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_TransactionManager.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
)

// NewMockTransactionManager creates a new instance of MockTransactionManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTransactionManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTransactionManager {
	mock := &MockTransactionManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockTransactionManager is an autogenerated mock type for the TransactionManager type
type MockTransactionManager struct {
	mock.Mock
}

type MockTransactionManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTransactionManager) EXPECT() *MockTransactionManager_Expecter {
	return &MockTransactionManager_Expecter{mock: &_m.Mock}
}

// Begin provides a mock function for the type MockTransactionManager
func (_mock *MockTransactionManager) Begin(ctx context.Context) (context.Context, error) {
	ret := _mock.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Begin")
	}

	var r0 context.Context
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) (context.Context, error)); ok {
		return returnFunc(ctx)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context) context.Context); ok {
		r0 = returnFunc(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = returnFunc(ctx)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockTransactionManager_Begin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Begin'
type MockTransactionManager_Begin_Call struct {
	*mock.Call
}

// Begin is a helper method to define mock.On call
//   - ctx
func (_e *MockTransactionManager_Expecter) Begin(ctx interface{}) *MockTransactionManager_Begin_Call {
	return &MockTransactionManager_Begin_Call{Call: _e.mock.On("Begin", ctx)}
}

func (_c *MockTransactionManager_Begin_Call) Run(run func(ctx context.Context)) *MockTransactionManager_Begin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockTransactionManager_Begin_Call) Return(TxContext context.Context, err error) *MockTransactionManager_Begin_Call {
	_c.Call.Return(TxContext, err)
	return _c
}

func (_c *MockTransactionManager_Begin_Call) RunAndReturn(run func(ctx context.Context) (context.Context, error)) *MockTransactionManager_Begin_Call {
	_c.Call.Return(run)
	return _c
}

// Commit provides a mock function for the type MockTransactionManager
func (_mock *MockTransactionManager) Commit(ctx context.Context) error {
	ret := _mock.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = returnFunc(ctx)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockTransactionManager_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type MockTransactionManager_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
//   - ctx
func (_e *MockTransactionManager_Expecter) Commit(ctx interface{}) *MockTransactionManager_Commit_Call {
	return &MockTransactionManager_Commit_Call{Call: _e.mock.On("Commit", ctx)}
}

func (_c *MockTransactionManager_Commit_Call) Run(run func(ctx context.Context)) *MockTransactionManager_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockTransactionManager_Commit_Call) Return(err error) *MockTransactionManager_Commit_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockTransactionManager_Commit_Call) RunAndReturn(run func(ctx context.Context) error) *MockTransactionManager_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// Execute provides a mock function for the type MockTransactionManager
func (_mock *MockTransactionManager) Execute(ctx context.Context, fn func(txCtx context.Context) error) error {
	ret := _mock.Called(ctx, fn)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, func(txCtx context.Context) error) error); ok {
		r0 = returnFunc(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockTransactionManager_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockTransactionManager_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx
//   - fn
func (_e *MockTransactionManager_Expecter) Execute(ctx interface{}, fn interface{}) *MockTransactionManager_Execute_Call {
	return &MockTransactionManager_Execute_Call{Call: _e.mock.On("Execute", ctx, fn)}
}

func (_c *MockTransactionManager_Execute_Call) Run(run func(ctx context.Context, fn func(txCtx context.Context) error)) *MockTransactionManager_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(func(txCtx context.Context) error))
	})
	return _c
}

func (_c *MockTransactionManager_Execute_Call) Return(err error) *MockTransactionManager_Execute_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockTransactionManager_Execute_Call) RunAndReturn(run func(ctx context.Context, fn func(txCtx context.Context) error) error) *MockTransactionManager_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// Rollback provides a mock function for the type MockTransactionManager
func (_mock *MockTransactionManager) Rollback(ctx context.Context) error {
	ret := _mock.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = returnFunc(ctx)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockTransactionManager_Rollback_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rollback'
type MockTransactionManager_Rollback_Call struct {
	*mock.Call
}

// Rollback is a helper method to define mock.On call
//   - ctx
func (_e *MockTransactionManager_Expecter) Rollback(ctx interface{}) *MockTransactionManager_Rollback_Call {
	return &MockTransactionManager_Rollback_Call{Call: _e.mock.On("Rollback", ctx)}
}

func (_c *MockTransactionManager_Rollback_Call) Run(run func(ctx context.Context)) *MockTransactionManager_Rollback_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockTransactionManager_Rollback_Call) Return(err error) *MockTransactionManager_Rollback_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockTransactionManager_Rollback_Call) RunAndReturn(run func(ctx context.Context) error) *MockTransactionManager_Rollback_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_SecurityHelper.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"
	"time"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
)

// NewMockSecurityHelper creates a new instance of MockSecurityHelper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockSecurityHelper(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSecurityHelper {
	mock := &MockSecurityHelper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockSecurityHelper is an autogenerated mock type for the SecurityHelper type
type MockSecurityHelper struct {
	mock.Mock
}

type MockSecurityHelper_Expecter struct {
	mock *mock.Mock
}

func (_m *MockSecurityHelper) EXPECT() *MockSecurityHelper_Expecter {
	return &MockSecurityHelper_Expecter{mock: &_m.Mock}
}

// CheckPasswordHash provides a mock function for the type MockSecurityHelper
func (_mock *MockSecurityHelper) CheckPasswordHash(ctx context.Context, password string, hash string) bool {
	ret := _mock.Called(ctx, password, hash)

	if len(ret) == 0 {
		panic("no return value specified for CheckPasswordHash")
	}

	var r0 bool
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = returnFunc(ctx, password, hash)
	} else {
		r0 = ret.Get(0).(bool)
	}
	return r0
}

// MockSecurityHelper_CheckPasswordHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckPasswordHash'
type MockSecurityHelper_CheckPasswordHash_Call struct {
	*mock.Call
}

// CheckPasswordHash is a helper method to define mock.On call
//   - ctx
//   - password
//   - hash
func (_e *MockSecurityHelper_Expecter) CheckPasswordHash(ctx interface{}, password interface{}, hash interface{}) *MockSecurityHelper_CheckPasswordHash_Call {
	return &MockSecurityHelper_CheckPasswordHash_Call{Call: _e.mock.On("CheckPasswordHash", ctx, password, hash)}
}

func (_c *MockSecurityHelper_CheckPasswordHash_Call) Run(run func(ctx context.Context, password string, hash string)) *MockSecurityHelper_CheckPasswordHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockSecurityHelper_CheckPasswordHash_Call) Return(b bool) *MockSecurityHelper_CheckPasswordHash_Call {
	_c.Call.Return(b)
	return _c
}

func (_c *MockSecurityHelper_CheckPasswordHash_Call) RunAndReturn(run func(ctx context.Context, password string, hash string) bool) *MockSecurityHelper_CheckPasswordHash_Call {
	_c.Call.Return(run)
	return _c
}

// GenerateJWT provides a mock function for the type MockSecurityHelper
func (_mock *MockSecurityHelper) GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error) {
	ret := _mock.Called(ctx, userID, duration)

	if len(ret) == 0 {
		panic("no return value specified for GenerateJWT")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, time.Duration) (string, error)); ok {
		return returnFunc(ctx, userID, duration)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, time.Duration) string); ok {
		r0 = returnFunc(ctx, userID, duration)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, time.Duration) error); ok {
		r1 = returnFunc(ctx, userID, duration)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockSecurityHelper_GenerateJWT_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenerateJWT'
type MockSecurityHelper_GenerateJWT_Call struct {
	*mock.Call
}

// GenerateJWT is a helper method to define mock.On call
//   - ctx
//   - userID
//   - duration
func (_e *MockSecurityHelper_Expecter) GenerateJWT(ctx interface{}, userID interface{}, duration interface{}) *MockSecurityHelper_GenerateJWT_Call {
	return &MockSecurityHelper_GenerateJWT_Call{Call: _e.mock.On("GenerateJWT", ctx, userID, duration)}
}

func (_c *MockSecurityHelper_GenerateJWT_Call) Run(run func(ctx context.Context, userID domain.UserID, duration time.Duration)) *MockSecurityHelper_GenerateJWT_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(time.Duration))
	})
	return _c
}

func (_c *MockSecurityHelper_GenerateJWT_Call) Return(s string, err error) *MockSecurityHelper_GenerateJWT_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockSecurityHelper_GenerateJWT_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error)) *MockSecurityHelper_GenerateJWT_Call {
	_c.Call.Return(run)
	return _c
}

// GenerateRefreshTokenValue provides a mock function for the type MockSecurityHelper
func (_mock *MockSecurityHelper) GenerateRefreshTokenValue() (string, error) {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GenerateRefreshTokenValue")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func() (string, error)); ok {
		return returnFunc()
	}
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func() error); ok {
		r1 = returnFunc()
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockSecurityHelper_GenerateRefreshTokenValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenerateRefreshTokenValue'
type MockSecurityHelper_GenerateRefreshTokenValue_Call struct {
	*mock.Call
}

// GenerateRefreshTokenValue is a helper method to define mock.On call
func (_e *MockSecurityHelper_Expecter) GenerateRefreshTokenValue() *MockSecurityHelper_GenerateRefreshTokenValue_Call {
	return &MockSecurityHelper_GenerateRefreshTokenValue_Call{Call: _e.mock.On("GenerateRefreshTokenValue")}
}

func (_c *MockSecurityHelper_GenerateRefreshTokenValue_Call) Run(run func()) *MockSecurityHelper_GenerateRefreshTokenValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSecurityHelper_GenerateRefreshTokenValue_Call) Return(s string, err error) *MockSecurityHelper_GenerateRefreshTokenValue_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockSecurityHelper_GenerateRefreshTokenValue_Call) RunAndReturn(run func() (string, error)) *MockSecurityHelper_GenerateRefreshTokenValue_Call {
	_c.Call.Return(run)
	return _c
}

// HashPassword provides a mock function for the type MockSecurityHelper
func (_mock *MockSecurityHelper) HashPassword(ctx context.Context, password string) (string, error) {
	ret := _mock.Called(ctx, password)

	if len(ret) == 0 {
		panic("no return value specified for HashPassword")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return returnFunc(ctx, password)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = returnFunc(ctx, password)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, password)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockSecurityHelper_HashPassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HashPassword'
type MockSecurityHelper_HashPassword_Call struct {
	*mock.Call
}

// HashPassword is a helper method to define mock.On call
//   - ctx
//   - password
func (_e *MockSecurityHelper_Expecter) HashPassword(ctx interface{}, password interface{}) *MockSecurityHelper_HashPassword_Call {
	return &MockSecurityHelper_HashPassword_Call{Call: _e.mock.On("HashPassword", ctx, password)}
}

func (_c *MockSecurityHelper_HashPassword_Call) Run(run func(ctx context.Context, password string)) *MockSecurityHelper_HashPassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockSecurityHelper_HashPassword_Call) Return(s string, err error) *MockSecurityHelper_HashPassword_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockSecurityHelper_HashPassword_Call) RunAndReturn(run func(ctx context.Context, password string) (string, error)) *MockSecurityHelper_HashPassword_Call {
	_c.Call.Return(run)
	return _c
}

// HashRefreshTokenValue provides a mock function for the type MockSecurityHelper
func (_mock *MockSecurityHelper) HashRefreshTokenValue(tokenValue string) string {
	ret := _mock.Called(tokenValue)

	if len(ret) == 0 {
		panic("no return value specified for HashRefreshTokenValue")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func(string) string); ok {
		r0 = returnFunc(tokenValue)
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// MockSecurityHelper_HashRefreshTokenValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HashRefreshTokenValue'
type MockSecurityHelper_HashRefreshTokenValue_Call struct {
	*mock.Call
}

// HashRefreshTokenValue is a helper method to define mock.On call
//   - tokenValue
func (_e *MockSecurityHelper_Expecter) HashRefreshTokenValue(tokenValue interface{}) *MockSecurityHelper_HashRefreshTokenValue_Call {
	return &MockSecurityHelper_HashRefreshTokenValue_Call{Call: _e.mock.On("HashRefreshTokenValue", tokenValue)}
}

func (_c *MockSecurityHelper_HashRefreshTokenValue_Call) Run(run func(tokenValue string)) *MockSecurityHelper_HashRefreshTokenValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockSecurityHelper_HashRefreshTokenValue_Call) Return(s string) *MockSecurityHelper_HashRefreshTokenValue_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *MockSecurityHelper_HashRefreshTokenValue_Call) RunAndReturn(run func(tokenValue string) string) *MockSecurityHelper_HashRefreshTokenValue_Call {
	_c.Call.Return(run)
	return _c
}

// VerifyJWT provides a mock function for the type MockSecurityHelper
func (_mock *MockSecurityHelper) VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error) {
	ret := _mock.Called(ctx, tokenString)

	if len(ret) == 0 {
		panic("no return value specified for VerifyJWT")
	}

	var r0 domain.UserID
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (domain.UserID, error)); ok {
		return returnFunc(ctx, tokenString)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) domain.UserID); ok {
		r0 = returnFunc(ctx, tokenString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(domain.UserID)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, tokenString)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockSecurityHelper_VerifyJWT_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VerifyJWT'
type MockSecurityHelper_VerifyJWT_Call struct {
	*mock.Call
}

// VerifyJWT is a helper method to define mock.On call
//   - ctx
//   - tokenString
func (_e *MockSecurityHelper_Expecter) VerifyJWT(ctx interface{}, tokenString interface{}) *MockSecurityHelper_VerifyJWT_Call {
	return &MockSecurityHelper_VerifyJWT_Call{Call: _e.mock.On("VerifyJWT", ctx, tokenString)}
}

func (_c *MockSecurityHelper_VerifyJWT_Call) Run(run func(ctx context.Context, tokenString string)) *MockSecurityHelper_VerifyJWT_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockSecurityHelper_VerifyJWT_Call) Return(userID domain.UserID, err error) *MockSecurityHelper_VerifyJWT_Call {
	_c.Call.Return(userID, err)
	return _c
}

func (_c *MockSecurityHelper_VerifyJWT_Call) RunAndReturn(run func(ctx context.Context, tokenString string) (domain.UserID, error)) *MockSecurityHelper_VerifyJWT_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_PlaybackProgressRepository.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// NewMockPlaybackProgressRepository creates a new instance of MockPlaybackProgressRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPlaybackProgressRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlaybackProgressRepository {
	mock := &MockPlaybackProgressRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockPlaybackProgressRepository is an autogenerated mock type for the PlaybackProgressRepository type
type MockPlaybackProgressRepository struct {
	mock.Mock
}

type MockPlaybackProgressRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPlaybackProgressRepository) EXPECT() *MockPlaybackProgressRepository_Expecter {
	return &MockPlaybackProgressRepository_Expecter{mock: &_m.Mock}
}

// Find provides a mock function for the type MockPlaybackProgressRepository
func (_mock *MockPlaybackProgressRepository) Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	ret := _mock.Called(ctx, userID, trackID)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 *domain.PlaybackProgress
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID) (*domain.PlaybackProgress, error)); ok {
		return returnFunc(ctx, userID, trackID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID) *domain.PlaybackProgress); ok {
		r0 = returnFunc(ctx, userID, trackID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.PlaybackProgress)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, domain.TrackID) error); ok {
		r1 = returnFunc(ctx, userID, trackID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockPlaybackProgressRepository_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type MockPlaybackProgressRepository_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//   - ctx
//   - userID
//   - trackID
func (_e *MockPlaybackProgressRepository_Expecter) Find(ctx interface{}, userID interface{}, trackID interface{}) *MockPlaybackProgressRepository_Find_Call {
	return &MockPlaybackProgressRepository_Find_Call{Call: _e.mock.On("Find", ctx, userID, trackID)}
}

func (_c *MockPlaybackProgressRepository_Find_Call) Run(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID)) *MockPlaybackProgressRepository_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(domain.TrackID))
	})
	return _c
}

func (_c *MockPlaybackProgressRepository_Find_Call) Return(playbackProgress *domain.PlaybackProgress, err error) *MockPlaybackProgressRepository_Find_Call {
	_c.Call.Return(playbackProgress, err)
	return _c
}

func (_c *MockPlaybackProgressRepository_Find_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)) *MockPlaybackProgressRepository_Find_Call {
	_c.Call.Return(run)
	return _c
}

// ListByUser provides a mock function for the type MockPlaybackProgressRepository
func (_mock *MockPlaybackProgressRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.PlaybackProgress, int, error) {
	ret := _mock.Called(ctx, userID, page)

	if len(ret) == 0 {
		panic("no return value specified for ListByUser")
	}

	var r0 []*domain.PlaybackProgress
	var r1 int
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, pagination.Page) ([]*domain.PlaybackProgress, int, error)); ok {
		return returnFunc(ctx, userID, page)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, pagination.Page) []*domain.PlaybackProgress); ok {
		r0 = returnFunc(ctx, userID, page)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.PlaybackProgress)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, pagination.Page) int); ok {
		r1 = returnFunc(ctx, userID, page)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, domain.UserID, pagination.Page) error); ok {
		r2 = returnFunc(ctx, userID, page)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockPlaybackProgressRepository_ListByUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByUser'
type MockPlaybackProgressRepository_ListByUser_Call struct {
	*mock.Call
}

// ListByUser is a helper method to define mock.On call
//   - ctx
//   - userID
//   - page
func (_e *MockPlaybackProgressRepository_Expecter) ListByUser(ctx interface{}, userID interface{}, page interface{}) *MockPlaybackProgressRepository_ListByUser_Call {
	return &MockPlaybackProgressRepository_ListByUser_Call{Call: _e.mock.On("ListByUser", ctx, userID, page)}
}

func (_c *MockPlaybackProgressRepository_ListByUser_Call) Run(run func(ctx context.Context, userID domain.UserID, page pagination.Page)) *MockPlaybackProgressRepository_ListByUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(pagination.Page))
	})
	return _c
}

func (_c *MockPlaybackProgressRepository_ListByUser_Call) Return(progressList []*domain.PlaybackProgress, total int, err error) *MockPlaybackProgressRepository_ListByUser_Call {
	_c.Call.Return(progressList, total, err)
	return _c
}

func (_c *MockPlaybackProgressRepository_ListByUser_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.PlaybackProgress, int, error)) *MockPlaybackProgressRepository_ListByUser_Call {
	_c.Call.Return(run)
	return _c
}

// Upsert provides a mock function for the type MockPlaybackProgressRepository
func (_mock *MockPlaybackProgressRepository) Upsert(ctx context.Context, progress *domain.PlaybackProgress) error {
	ret := _mock.Called(ctx, progress)

	if len(ret) == 0 {
		panic("no return value specified for Upsert")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.PlaybackProgress) error); ok {
		r0 = returnFunc(ctx, progress)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockPlaybackProgressRepository_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type MockPlaybackProgressRepository_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//   - ctx
//   - progress
func (_e *MockPlaybackProgressRepository_Expecter) Upsert(ctx interface{}, progress interface{}) *MockPlaybackProgressRepository_Upsert_Call {
	return &MockPlaybackProgressRepository_Upsert_Call{Call: _e.mock.On("Upsert", ctx, progress)}
}

func (_c *MockPlaybackProgressRepository_Upsert_Call) Run(run func(ctx context.Context, progress *domain.PlaybackProgress)) *MockPlaybackProgressRepository_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.PlaybackProgress))
	})
	return _c
}

func (_c *MockPlaybackProgressRepository_Upsert_Call) Return(err error) *MockPlaybackProgressRepository_Upsert_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockPlaybackProgressRepository_Upsert_Call) RunAndReturn(run func(ctx context.Context, progress *domain.PlaybackProgress) error) *MockPlaybackProgressRepository_Upsert_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_UserRepository.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
)

// NewMockUserRepository creates a new instance of MockUserRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserRepository {
	mock := &MockUserRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockUserRepository is an autogenerated mock type for the UserRepository type
type MockUserRepository struct {
	mock.Mock
}

type MockUserRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserRepository) EXPECT() *MockUserRepository_Expecter {
	return &MockUserRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function for the type MockUserRepository
func (_mock *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	ret := _mock.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.User) error); ok {
		r0 = returnFunc(ctx, user)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockUserRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockUserRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx
//   - user
func (_e *MockUserRepository_Expecter) Create(ctx interface{}, user interface{}) *MockUserRepository_Create_Call {
	return &MockUserRepository_Create_Call{Call: _e.mock.On("Create", ctx, user)}
}

func (_c *MockUserRepository_Create_Call) Run(run func(ctx context.Context, user *domain.User)) *MockUserRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.User))
	})
	return _c
}

func (_c *MockUserRepository_Create_Call) Return(err error) *MockUserRepository_Create_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockUserRepository_Create_Call) RunAndReturn(run func(ctx context.Context, user *domain.User) error) *MockUserRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// EmailExists provides a mock function for the type MockUserRepository
func (_mock *MockUserRepository) EmailExists(ctx context.Context, email domain.Email) (bool, error) {
	ret := _mock.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for EmailExists")
	}

	var r0 bool
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.Email) (bool, error)); ok {
		return returnFunc(ctx, email)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.Email) bool); ok {
		r0 = returnFunc(ctx, email)
	} else {
		r0 = ret.Get(0).(bool)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.Email) error); ok {
		r1 = returnFunc(ctx, email)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserRepository_EmailExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EmailExists'
type MockUserRepository_EmailExists_Call struct {
	*mock.Call
}

// EmailExists is a helper method to define mock.On call
//   - ctx
//   - email
func (_e *MockUserRepository_Expecter) EmailExists(ctx interface{}, email interface{}) *MockUserRepository_EmailExists_Call {
	return &MockUserRepository_EmailExists_Call{Call: _e.mock.On("EmailExists", ctx, email)}
}

func (_c *MockUserRepository_EmailExists_Call) Run(run func(ctx context.Context, email domain.Email)) *MockUserRepository_EmailExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.Email))
	})
	return _c
}

func (_c *MockUserRepository_EmailExists_Call) Return(b bool, err error) *MockUserRepository_EmailExists_Call {
	_c.Call.Return(b, err)
	return _c
}

func (_c *MockUserRepository_EmailExists_Call) RunAndReturn(run func(ctx context.Context, email domain.Email) (bool, error)) *MockUserRepository_EmailExists_Call {
	_c.Call.Return(run)
	return _c
}

// FindByEmail provides a mock function for the type MockUserRepository
func (_mock *MockUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	ret := _mock.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for FindByEmail")
	}

	var r0 *domain.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.Email) (*domain.User, error)); ok {
		return returnFunc(ctx, email)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.Email) *domain.User); ok {
		r0 = returnFunc(ctx, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.Email) error); ok {
		r1 = returnFunc(ctx, email)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserRepository_FindByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByEmail'
type MockUserRepository_FindByEmail_Call struct {
	*mock.Call
}

// FindByEmail is a helper method to define mock.On call
//   - ctx
//   - email
func (_e *MockUserRepository_Expecter) FindByEmail(ctx interface{}, email interface{}) *MockUserRepository_FindByEmail_Call {
	return &MockUserRepository_FindByEmail_Call{Call: _e.mock.On("FindByEmail", ctx, email)}
}

func (_c *MockUserRepository_FindByEmail_Call) Run(run func(ctx context.Context, email domain.Email)) *MockUserRepository_FindByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.Email))
	})
	return _c
}

func (_c *MockUserRepository_FindByEmail_Call) Return(user *domain.User, err error) *MockUserRepository_FindByEmail_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *MockUserRepository_FindByEmail_Call) RunAndReturn(run func(ctx context.Context, email domain.Email) (*domain.User, error)) *MockUserRepository_FindByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// FindByID provides a mock function for the type MockUserRepository
func (_mock *MockUserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for FindByID")
	}

	var r0 *domain.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID) (*domain.User, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID) *domain.User); ok {
		r0 = returnFunc(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserRepository_FindByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByID'
type MockUserRepository_FindByID_Call struct {
	*mock.Call
}

// FindByID is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockUserRepository_Expecter) FindByID(ctx interface{}, id interface{}) *MockUserRepository_FindByID_Call {
	return &MockUserRepository_FindByID_Call{Call: _e.mock.On("FindByID", ctx, id)}
}

func (_c *MockUserRepository_FindByID_Call) Run(run func(ctx context.Context, id domain.UserID)) *MockUserRepository_FindByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID))
	})
	return _c
}

func (_c *MockUserRepository_FindByID_Call) Return(user *domain.User, err error) *MockUserRepository_FindByID_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *MockUserRepository_FindByID_Call) RunAndReturn(run func(ctx context.Context, id domain.UserID) (*domain.User, error)) *MockUserRepository_FindByID_Call {
	_c.Call.Return(run)
	return _c
}

// FindByProviderID provides a mock function for the type MockUserRepository
func (_mock *MockUserRepository) FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error) {
	ret := _mock.Called(ctx, provider, providerUserID)

	if len(ret) == 0 {
		panic("no return value specified for FindByProviderID")
	}

	var r0 *domain.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.AuthProvider, string) (*domain.User, error)); ok {
		return returnFunc(ctx, provider, providerUserID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.AuthProvider, string) *domain.User); ok {
		r0 = returnFunc(ctx, provider, providerUserID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.AuthProvider, string) error); ok {
		r1 = returnFunc(ctx, provider, providerUserID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserRepository_FindByProviderID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByProviderID'
type MockUserRepository_FindByProviderID_Call struct {
	*mock.Call
}

// FindByProviderID is a helper method to define mock.On call
//   - ctx
//   - provider
//   - providerUserID
func (_e *MockUserRepository_Expecter) FindByProviderID(ctx interface{}, provider interface{}, providerUserID interface{}) *MockUserRepository_FindByProviderID_Call {
	return &MockUserRepository_FindByProviderID_Call{Call: _e.mock.On("FindByProviderID", ctx, provider, providerUserID)}
}

func (_c *MockUserRepository_FindByProviderID_Call) Run(run func(ctx context.Context, provider domain.AuthProvider, providerUserID string)) *MockUserRepository_FindByProviderID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.AuthProvider), args[2].(string))
	})
	return _c
}

func (_c *MockUserRepository_FindByProviderID_Call) Return(user *domain.User, err error) *MockUserRepository_FindByProviderID_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *MockUserRepository_FindByProviderID_Call) RunAndReturn(run func(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error)) *MockUserRepository_FindByProviderID_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function for the type MockUserRepository
func (_mock *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	ret := _mock.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.User) error); ok {
		r0 = returnFunc(ctx, user)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockUserRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockUserRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx
//   - user
func (_e *MockUserRepository_Expecter) Update(ctx interface{}, user interface{}) *MockUserRepository_Update_Call {
	return &MockUserRepository_Update_Call{Call: _e.mock.On("Update", ctx, user)}
}

func (_c *MockUserRepository_Update_Call) Run(run func(ctx context.Context, user *domain.User)) *MockUserRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.User))
	})
	return _c
}

func (_c *MockUserRepository_Update_Call) Return(err error) *MockUserRepository_Update_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockUserRepository_Update_Call) RunAndReturn(run func(ctx context.Context, user *domain.User) error) *MockUserRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_AuthUseCase.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// NewMockAuthUseCase creates a new instance of MockAuthUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAuthUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuthUseCase {
	mock := &MockAuthUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockAuthUseCase is an autogenerated mock type for the AuthUseCase type
type MockAuthUseCase struct {
	mock.Mock
}

type MockAuthUseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAuthUseCase) EXPECT() *MockAuthUseCase_Expecter {
	return &MockAuthUseCase_Expecter{mock: &_m.Mock}
}

// AuthenticateWithGoogle provides a mock function for the type MockAuthUseCase
func (_mock *MockAuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (port.AuthResult, error) {
	ret := _mock.Called(ctx, googleIdToken)

	if len(ret) == 0 {
		panic("no return value specified for AuthenticateWithGoogle")
	}

	var r0 port.AuthResult
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (port.AuthResult, error)); ok {
		return returnFunc(ctx, googleIdToken)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) port.AuthResult); ok {
		r0 = returnFunc(ctx, googleIdToken)
	} else {
		r0 = ret.Get(0).(port.AuthResult)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, googleIdToken)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAuthUseCase_AuthenticateWithGoogle_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AuthenticateWithGoogle'
type MockAuthUseCase_AuthenticateWithGoogle_Call struct {
	*mock.Call
}

// AuthenticateWithGoogle is a helper method to define mock.On call
//   - ctx
//   - googleIdToken
func (_e *MockAuthUseCase_Expecter) AuthenticateWithGoogle(ctx interface{}, googleIdToken interface{}) *MockAuthUseCase_AuthenticateWithGoogle_Call {
	return &MockAuthUseCase_AuthenticateWithGoogle_Call{Call: _e.mock.On("AuthenticateWithGoogle", ctx, googleIdToken)}
}

func (_c *MockAuthUseCase_AuthenticateWithGoogle_Call) Run(run func(ctx context.Context, googleIdToken string)) *MockAuthUseCase_AuthenticateWithGoogle_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockAuthUseCase_AuthenticateWithGoogle_Call) Return(authResult port.AuthResult, err error) *MockAuthUseCase_AuthenticateWithGoogle_Call {
	_c.Call.Return(authResult, err)
	return _c
}

func (_c *MockAuthUseCase_AuthenticateWithGoogle_Call) RunAndReturn(run func(ctx context.Context, googleIdToken string) (port.AuthResult, error)) *MockAuthUseCase_AuthenticateWithGoogle_Call {
	_c.Call.Return(run)
	return _c
}

// LoginWithPassword provides a mock function for the type MockAuthUseCase
func (_mock *MockAuthUseCase) LoginWithPassword(ctx context.Context, emailStr string, password string) (port.AuthResult, error) {
	ret := _mock.Called(ctx, emailStr, password)

	if len(ret) == 0 {
		panic("no return value specified for LoginWithPassword")
	}

	var r0 port.AuthResult
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) (port.AuthResult, error)); ok {
		return returnFunc(ctx, emailStr, password)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) port.AuthResult); ok {
		r0 = returnFunc(ctx, emailStr, password)
	} else {
		r0 = ret.Get(0).(port.AuthResult)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = returnFunc(ctx, emailStr, password)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAuthUseCase_LoginWithPassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoginWithPassword'
type MockAuthUseCase_LoginWithPassword_Call struct {
	*mock.Call
}

// LoginWithPassword is a helper method to define mock.On call
//   - ctx
//   - emailStr
//   - password
func (_e *MockAuthUseCase_Expecter) LoginWithPassword(ctx interface{}, emailStr interface{}, password interface{}) *MockAuthUseCase_LoginWithPassword_Call {
	return &MockAuthUseCase_LoginWithPassword_Call{Call: _e.mock.On("LoginWithPassword", ctx, emailStr, password)}
}

func (_c *MockAuthUseCase_LoginWithPassword_Call) Run(run func(ctx context.Context, emailStr string, password string)) *MockAuthUseCase_LoginWithPassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockAuthUseCase_LoginWithPassword_Call) Return(authResult port.AuthResult, err error) *MockAuthUseCase_LoginWithPassword_Call {
	_c.Call.Return(authResult, err)
	return _c
}

func (_c *MockAuthUseCase_LoginWithPassword_Call) RunAndReturn(run func(ctx context.Context, emailStr string, password string) (port.AuthResult, error)) *MockAuthUseCase_LoginWithPassword_Call {
	_c.Call.Return(run)
	return _c
}

// Logout provides a mock function for the type MockAuthUseCase
func (_mock *MockAuthUseCase) Logout(ctx context.Context, refreshTokenValue string) error {
	ret := _mock.Called(ctx, refreshTokenValue)

	if len(ret) == 0 {
		panic("no return value specified for Logout")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, refreshTokenValue)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAuthUseCase_Logout_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Logout'
type MockAuthUseCase_Logout_Call struct {
	*mock.Call
}

// Logout is a helper method to define mock.On call
//   - ctx
//   - refreshTokenValue
func (_e *MockAuthUseCase_Expecter) Logout(ctx interface{}, refreshTokenValue interface{}) *MockAuthUseCase_Logout_Call {
	return &MockAuthUseCase_Logout_Call{Call: _e.mock.On("Logout", ctx, refreshTokenValue)}
}

func (_c *MockAuthUseCase_Logout_Call) Run(run func(ctx context.Context, refreshTokenValue string)) *MockAuthUseCase_Logout_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockAuthUseCase_Logout_Call) Return(err error) *MockAuthUseCase_Logout_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAuthUseCase_Logout_Call) RunAndReturn(run func(ctx context.Context, refreshTokenValue string) error) *MockAuthUseCase_Logout_Call {
	_c.Call.Return(run)
	return _c
}

// RefreshAccessToken provides a mock function for the type MockAuthUseCase
func (_mock *MockAuthUseCase) RefreshAccessToken(ctx context.Context, refreshTokenValue string) (port.AuthResult, error) {
	ret := _mock.Called(ctx, refreshTokenValue)

	if len(ret) == 0 {
		panic("no return value specified for RefreshAccessToken")
	}

	var r0 port.AuthResult
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (port.AuthResult, error)); ok {
		return returnFunc(ctx, refreshTokenValue)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) port.AuthResult); ok {
		r0 = returnFunc(ctx, refreshTokenValue)
	} else {
		r0 = ret.Get(0).(port.AuthResult)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, refreshTokenValue)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAuthUseCase_RefreshAccessToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RefreshAccessToken'
type MockAuthUseCase_RefreshAccessToken_Call struct {
	*mock.Call
}

// RefreshAccessToken is a helper method to define mock.On call
//   - ctx
//   - refreshTokenValue
func (_e *MockAuthUseCase_Expecter) RefreshAccessToken(ctx interface{}, refreshTokenValue interface{}) *MockAuthUseCase_RefreshAccessToken_Call {
	return &MockAuthUseCase_RefreshAccessToken_Call{Call: _e.mock.On("RefreshAccessToken", ctx, refreshTokenValue)}
}

func (_c *MockAuthUseCase_RefreshAccessToken_Call) Run(run func(ctx context.Context, refreshTokenValue string)) *MockAuthUseCase_RefreshAccessToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockAuthUseCase_RefreshAccessToken_Call) Return(authResult port.AuthResult, err error) *MockAuthUseCase_RefreshAccessToken_Call {
	_c.Call.Return(authResult, err)
	return _c
}

func (_c *MockAuthUseCase_RefreshAccessToken_Call) RunAndReturn(run func(ctx context.Context, refreshTokenValue string) (port.AuthResult, error)) *MockAuthUseCase_RefreshAccessToken_Call {
	_c.Call.Return(run)
	return _c
}

// RegisterWithPassword provides a mock function for the type MockAuthUseCase
func (_mock *MockAuthUseCase) RegisterWithPassword(ctx context.Context, emailStr string, password string, name string) (*domain.User, port.AuthResult, error) {
	ret := _mock.Called(ctx, emailStr, password, name)

	if len(ret) == 0 {
		panic("no return value specified for RegisterWithPassword")
	}

	var r0 *domain.User
	var r1 port.AuthResult
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, string) (*domain.User, port.AuthResult, error)); ok {
		return returnFunc(ctx, emailStr, password, name)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, string) *domain.User); ok {
		r0 = returnFunc(ctx, emailStr, password, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, string) port.AuthResult); ok {
		r1 = returnFunc(ctx, emailStr, password, name)
	} else {
		r1 = ret.Get(1).(port.AuthResult)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string, string, string) error); ok {
		r2 = returnFunc(ctx, emailStr, password, name)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockAuthUseCase_RegisterWithPassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegisterWithPassword'
type MockAuthUseCase_RegisterWithPassword_Call struct {
	*mock.Call
}

// RegisterWithPassword is a helper method to define mock.On call
//   - ctx
//   - emailStr
//   - password
//   - name
func (_e *MockAuthUseCase_Expecter) RegisterWithPassword(ctx interface{}, emailStr interface{}, password interface{}, name interface{}) *MockAuthUseCase_RegisterWithPassword_Call {
	return &MockAuthUseCase_RegisterWithPassword_Call{Call: _e.mock.On("RegisterWithPassword", ctx, emailStr, password, name)}
}

func (_c *MockAuthUseCase_RegisterWithPassword_Call) Run(run func(ctx context.Context, emailStr string, password string, name string)) *MockAuthUseCase_RegisterWithPassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *MockAuthUseCase_RegisterWithPassword_Call) Return(user *domain.User, authResult port.AuthResult, err error) *MockAuthUseCase_RegisterWithPassword_Call {
	_c.Call.Return(user, authResult, err)
	return _c
}

func (_c *MockAuthUseCase_RegisterWithPassword_Call) RunAndReturn(run func(ctx context.Context, emailStr string, password string, name string) (*domain.User, port.AuthResult, error)) *MockAuthUseCase_RegisterWithPassword_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_AudioContentUseCase.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// NewMockAudioContentUseCase creates a new instance of MockAudioContentUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAudioContentUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAudioContentUseCase {
	mock := &MockAudioContentUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockAudioContentUseCase is an autogenerated mock type for the AudioContentUseCase type
type MockAudioContentUseCase struct {
	mock.Mock
}

type MockAudioContentUseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAudioContentUseCase) EXPECT() *MockAudioContentUseCase_Expecter {
	return &MockAudioContentUseCase_Expecter{mock: &_m.Mock}
}

// CreateCollection provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) CreateCollection(ctx context.Context, title string, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	ret := _mock.Called(ctx, title, description, colType, initialTrackIDs)

	if len(ret) == 0 {
		panic("no return value specified for CreateCollection")
	}

	var r0 *domain.AudioCollection
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, domain.CollectionType, []domain.TrackID) (*domain.AudioCollection, error)); ok {
		return returnFunc(ctx, title, description, colType, initialTrackIDs)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, domain.CollectionType, []domain.TrackID) *domain.AudioCollection); ok {
		r0 = returnFunc(ctx, title, description, colType, initialTrackIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.AudioCollection)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, domain.CollectionType, []domain.TrackID) error); ok {
		r1 = returnFunc(ctx, title, description, colType, initialTrackIDs)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioContentUseCase_CreateCollection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateCollection'
type MockAudioContentUseCase_CreateCollection_Call struct {
	*mock.Call
}

// CreateCollection is a helper method to define mock.On call
//   - ctx
//   - title
//   - description
//   - colType
//   - initialTrackIDs
func (_e *MockAudioContentUseCase_Expecter) CreateCollection(ctx interface{}, title interface{}, description interface{}, colType interface{}, initialTrackIDs interface{}) *MockAudioContentUseCase_CreateCollection_Call {
	return &MockAudioContentUseCase_CreateCollection_Call{Call: _e.mock.On("CreateCollection", ctx, title, description, colType, initialTrackIDs)}
}

func (_c *MockAudioContentUseCase_CreateCollection_Call) Run(run func(ctx context.Context, title string, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID)) *MockAudioContentUseCase_CreateCollection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(domain.CollectionType), args[4].([]domain.TrackID))
	})
	return _c
}

func (_c *MockAudioContentUseCase_CreateCollection_Call) Return(audioCollection *domain.AudioCollection, err error) *MockAudioContentUseCase_CreateCollection_Call {
	_c.Call.Return(audioCollection, err)
	return _c
}

func (_c *MockAudioContentUseCase_CreateCollection_Call) RunAndReturn(run func(ctx context.Context, title string, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error)) *MockAudioContentUseCase_CreateCollection_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteCollection provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	ret := _mock.Called(ctx, collectionID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteCollection")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) error); ok {
		r0 = returnFunc(ctx, collectionID)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioContentUseCase_DeleteCollection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteCollection'
type MockAudioContentUseCase_DeleteCollection_Call struct {
	*mock.Call
}

// DeleteCollection is a helper method to define mock.On call
//   - ctx
//   - collectionID
func (_e *MockAudioContentUseCase_Expecter) DeleteCollection(ctx interface{}, collectionID interface{}) *MockAudioContentUseCase_DeleteCollection_Call {
	return &MockAudioContentUseCase_DeleteCollection_Call{Call: _e.mock.On("DeleteCollection", ctx, collectionID)}
}

func (_c *MockAudioContentUseCase_DeleteCollection_Call) Run(run func(ctx context.Context, collectionID domain.CollectionID)) *MockAudioContentUseCase_DeleteCollection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID))
	})
	return _c
}

func (_c *MockAudioContentUseCase_DeleteCollection_Call) Return(err error) *MockAudioContentUseCase_DeleteCollection_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioContentUseCase_DeleteCollection_Call) RunAndReturn(run func(ctx context.Context, collectionID domain.CollectionID) error) *MockAudioContentUseCase_DeleteCollection_Call {
	_c.Call.Return(run)
	return _c
}

// GetAudioTrackDetails provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*port.GetAudioTrackDetailsResult, error) {
	ret := _mock.Called(ctx, trackID)

	if len(ret) == 0 {
		panic("no return value specified for GetAudioTrackDetails")
	}

	var r0 *port.GetAudioTrackDetailsResult
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) (*port.GetAudioTrackDetailsResult, error)); ok {
		return returnFunc(ctx, trackID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) *port.GetAudioTrackDetailsResult); ok {
		r0 = returnFunc(ctx, trackID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*port.GetAudioTrackDetailsResult)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.TrackID) error); ok {
		r1 = returnFunc(ctx, trackID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioContentUseCase_GetAudioTrackDetails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAudioTrackDetails'
type MockAudioContentUseCase_GetAudioTrackDetails_Call struct {
	*mock.Call
}

// GetAudioTrackDetails is a helper method to define mock.On call
//   - ctx
//   - trackID
func (_e *MockAudioContentUseCase_Expecter) GetAudioTrackDetails(ctx interface{}, trackID interface{}) *MockAudioContentUseCase_GetAudioTrackDetails_Call {
	return &MockAudioContentUseCase_GetAudioTrackDetails_Call{Call: _e.mock.On("GetAudioTrackDetails", ctx, trackID)}
}

func (_c *MockAudioContentUseCase_GetAudioTrackDetails_Call) Run(run func(ctx context.Context, trackID domain.TrackID)) *MockAudioContentUseCase_GetAudioTrackDetails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.TrackID))
	})
	return _c
}

func (_c *MockAudioContentUseCase_GetAudioTrackDetails_Call) Return(getAudioTrackDetailsResult *port.GetAudioTrackDetailsResult, err error) *MockAudioContentUseCase_GetAudioTrackDetails_Call {
	_c.Call.Return(getAudioTrackDetailsResult, err)
	return _c
}

func (_c *MockAudioContentUseCase_GetAudioTrackDetails_Call) RunAndReturn(run func(ctx context.Context, trackID domain.TrackID) (*port.GetAudioTrackDetailsResult, error)) *MockAudioContentUseCase_GetAudioTrackDetails_Call {
	_c.Call.Return(run)
	return _c
}

// GetCollectionDetails provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	ret := _mock.Called(ctx, collectionID)

	if len(ret) == 0 {
		panic("no return value specified for GetCollectionDetails")
	}

	var r0 *domain.AudioCollection
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) (*domain.AudioCollection, error)); ok {
		return returnFunc(ctx, collectionID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) *domain.AudioCollection); ok {
		r0 = returnFunc(ctx, collectionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.AudioCollection)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.CollectionID) error); ok {
		r1 = returnFunc(ctx, collectionID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioContentUseCase_GetCollectionDetails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCollectionDetails'
type MockAudioContentUseCase_GetCollectionDetails_Call struct {
	*mock.Call
}

// GetCollectionDetails is a helper method to define mock.On call
//   - ctx
//   - collectionID
func (_e *MockAudioContentUseCase_Expecter) GetCollectionDetails(ctx interface{}, collectionID interface{}) *MockAudioContentUseCase_GetCollectionDetails_Call {
	return &MockAudioContentUseCase_GetCollectionDetails_Call{Call: _e.mock.On("GetCollectionDetails", ctx, collectionID)}
}

func (_c *MockAudioContentUseCase_GetCollectionDetails_Call) Run(run func(ctx context.Context, collectionID domain.CollectionID)) *MockAudioContentUseCase_GetCollectionDetails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID))
	})
	return _c
}

func (_c *MockAudioContentUseCase_GetCollectionDetails_Call) Return(audioCollection *domain.AudioCollection, err error) *MockAudioContentUseCase_GetCollectionDetails_Call {
	_c.Call.Return(audioCollection, err)
	return _c
}

func (_c *MockAudioContentUseCase_GetCollectionDetails_Call) RunAndReturn(run func(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error)) *MockAudioContentUseCase_GetCollectionDetails_Call {
	_c.Call.Return(run)
	return _c
}

// GetCollectionTracks provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	ret := _mock.Called(ctx, collectionID)

	if len(ret) == 0 {
		panic("no return value specified for GetCollectionTracks")
	}

	var r0 []*domain.AudioTrack
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) ([]*domain.AudioTrack, error)); ok {
		return returnFunc(ctx, collectionID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID) []*domain.AudioTrack); ok {
		r0 = returnFunc(ctx, collectionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.AudioTrack)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.CollectionID) error); ok {
		r1 = returnFunc(ctx, collectionID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioContentUseCase_GetCollectionTracks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCollectionTracks'
type MockAudioContentUseCase_GetCollectionTracks_Call struct {
	*mock.Call
}

// GetCollectionTracks is a helper method to define mock.On call
//   - ctx
//   - collectionID
func (_e *MockAudioContentUseCase_Expecter) GetCollectionTracks(ctx interface{}, collectionID interface{}) *MockAudioContentUseCase_GetCollectionTracks_Call {
	return &MockAudioContentUseCase_GetCollectionTracks_Call{Call: _e.mock.On("GetCollectionTracks", ctx, collectionID)}
}

func (_c *MockAudioContentUseCase_GetCollectionTracks_Call) Run(run func(ctx context.Context, collectionID domain.CollectionID)) *MockAudioContentUseCase_GetCollectionTracks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID))
	})
	return _c
}

func (_c *MockAudioContentUseCase_GetCollectionTracks_Call) Return(audioTracks []*domain.AudioTrack, err error) *MockAudioContentUseCase_GetCollectionTracks_Call {
	_c.Call.Return(audioTracks, err)
	return _c
}

func (_c *MockAudioContentUseCase_GetCollectionTracks_Call) RunAndReturn(run func(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error)) *MockAudioContentUseCase_GetCollectionTracks_Call {
	_c.Call.Return(run)
	return _c
}

// ListTracks provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) ListTracks(ctx context.Context, input port.ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error) {
	ret := _mock.Called(ctx, input)

	if len(ret) == 0 {
		panic("no return value specified for ListTracks")
	}

	var r0 []*domain.AudioTrack
	var r1 int
	var r2 pagination.Page
	var r3 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error)); ok {
		return returnFunc(ctx, input)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListTracksInput) []*domain.AudioTrack); ok {
		r0 = returnFunc(ctx, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.AudioTrack)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, port.ListTracksInput) int); ok {
		r1 = returnFunc(ctx, input)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, port.ListTracksInput) pagination.Page); ok {
		r2 = returnFunc(ctx, input)
	} else {
		r2 = ret.Get(2).(pagination.Page)
	}
	if returnFunc, ok := ret.Get(3).(func(context.Context, port.ListTracksInput) error); ok {
		r3 = returnFunc(ctx, input)
	} else {
		r3 = ret.Error(3)
	}
	return r0, r1, r2, r3
}

// MockAudioContentUseCase_ListTracks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListTracks'
type MockAudioContentUseCase_ListTracks_Call struct {
	*mock.Call
}

// ListTracks is a helper method to define mock.On call
//   - ctx
//   - input
func (_e *MockAudioContentUseCase_Expecter) ListTracks(ctx interface{}, input interface{}) *MockAudioContentUseCase_ListTracks_Call {
	return &MockAudioContentUseCase_ListTracks_Call{Call: _e.mock.On("ListTracks", ctx, input)}
}

func (_c *MockAudioContentUseCase_ListTracks_Call) Run(run func(ctx context.Context, input port.ListTracksInput)) *MockAudioContentUseCase_ListTracks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(port.ListTracksInput))
	})
	return _c
}

func (_c *MockAudioContentUseCase_ListTracks_Call) Return(audioTracks []*domain.AudioTrack, n int, page pagination.Page, err error) *MockAudioContentUseCase_ListTracks_Call {
	_c.Call.Return(audioTracks, n, page, err)
	return _c
}

func (_c *MockAudioContentUseCase_ListTracks_Call) RunAndReturn(run func(ctx context.Context, input port.ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error)) *MockAudioContentUseCase_ListTracks_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateCollectionMetadata provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title string, description string) error {
	ret := _mock.Called(ctx, collectionID, title, description)

	if len(ret) == 0 {
		panic("no return value specified for UpdateCollectionMetadata")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID, string, string) error); ok {
		r0 = returnFunc(ctx, collectionID, title, description)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioContentUseCase_UpdateCollectionMetadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateCollectionMetadata'
type MockAudioContentUseCase_UpdateCollectionMetadata_Call struct {
	*mock.Call
}

// UpdateCollectionMetadata is a helper method to define mock.On call
//   - ctx
//   - collectionID
//   - title
//   - description
func (_e *MockAudioContentUseCase_Expecter) UpdateCollectionMetadata(ctx interface{}, collectionID interface{}, title interface{}, description interface{}) *MockAudioContentUseCase_UpdateCollectionMetadata_Call {
	return &MockAudioContentUseCase_UpdateCollectionMetadata_Call{Call: _e.mock.On("UpdateCollectionMetadata", ctx, collectionID, title, description)}
}

func (_c *MockAudioContentUseCase_UpdateCollectionMetadata_Call) Run(run func(ctx context.Context, collectionID domain.CollectionID, title string, description string)) *MockAudioContentUseCase_UpdateCollectionMetadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *MockAudioContentUseCase_UpdateCollectionMetadata_Call) Return(err error) *MockAudioContentUseCase_UpdateCollectionMetadata_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioContentUseCase_UpdateCollectionMetadata_Call) RunAndReturn(run func(ctx context.Context, collectionID domain.CollectionID, title string, description string) error) *MockAudioContentUseCase_UpdateCollectionMetadata_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateCollectionTracks provides a mock function for the type MockAudioContentUseCase
func (_mock *MockAudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	ret := _mock.Called(ctx, collectionID, orderedTrackIDs)

	if len(ret) == 0 {
		panic("no return value specified for UpdateCollectionTracks")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.CollectionID, []domain.TrackID) error); ok {
		r0 = returnFunc(ctx, collectionID, orderedTrackIDs)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioContentUseCase_UpdateCollectionTracks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateCollectionTracks'
type MockAudioContentUseCase_UpdateCollectionTracks_Call struct {
	*mock.Call
}

// UpdateCollectionTracks is a helper method to define mock.On call
//   - ctx
//   - collectionID
//   - orderedTrackIDs
func (_e *MockAudioContentUseCase_Expecter) UpdateCollectionTracks(ctx interface{}, collectionID interface{}, orderedTrackIDs interface{}) *MockAudioContentUseCase_UpdateCollectionTracks_Call {
	return &MockAudioContentUseCase_UpdateCollectionTracks_Call{Call: _e.mock.On("UpdateCollectionTracks", ctx, collectionID, orderedTrackIDs)}
}

func (_c *MockAudioContentUseCase_UpdateCollectionTracks_Call) Run(run func(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID)) *MockAudioContentUseCase_UpdateCollectionTracks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.CollectionID), args[2].([]domain.TrackID))
	})
	return _c
}

func (_c *MockAudioContentUseCase_UpdateCollectionTracks_Call) Return(err error) *MockAudioContentUseCase_UpdateCollectionTracks_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioContentUseCase_UpdateCollectionTracks_Call) RunAndReturn(run func(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error) *MockAudioContentUseCase_UpdateCollectionTracks_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_UserActivityUseCase.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"
	"time"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// NewMockUserActivityUseCase creates a new instance of MockUserActivityUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserActivityUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserActivityUseCase {
	mock := &MockUserActivityUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockUserActivityUseCase is an autogenerated mock type for the UserActivityUseCase type
type MockUserActivityUseCase struct {
	mock.Mock
}

type MockUserActivityUseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserActivityUseCase) EXPECT() *MockUserActivityUseCase_Expecter {
	return &MockUserActivityUseCase_Expecter{mock: &_m.Mock}
}

// CreateBookmark provides a mock function for the type MockUserActivityUseCase
func (_mock *MockUserActivityUseCase) CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error) {
	ret := _mock.Called(ctx, userID, trackID, timestamp, note)

	if len(ret) == 0 {
		panic("no return value specified for CreateBookmark")
	}

	var r0 *domain.Bookmark
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID, time.Duration, string) (*domain.Bookmark, error)); ok {
		return returnFunc(ctx, userID, trackID, timestamp, note)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID, time.Duration, string) *domain.Bookmark); ok {
		r0 = returnFunc(ctx, userID, trackID, timestamp, note)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Bookmark)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, domain.TrackID, time.Duration, string) error); ok {
		r1 = returnFunc(ctx, userID, trackID, timestamp, note)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserActivityUseCase_CreateBookmark_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateBookmark'
type MockUserActivityUseCase_CreateBookmark_Call struct {
	*mock.Call
}

// CreateBookmark is a helper method to define mock.On call
//   - ctx
//   - userID
//   - trackID
//   - timestamp
//   - note
func (_e *MockUserActivityUseCase_Expecter) CreateBookmark(ctx interface{}, userID interface{}, trackID interface{}, timestamp interface{}, note interface{}) *MockUserActivityUseCase_CreateBookmark_Call {
	return &MockUserActivityUseCase_CreateBookmark_Call{Call: _e.mock.On("CreateBookmark", ctx, userID, trackID, timestamp, note)}
}

func (_c *MockUserActivityUseCase_CreateBookmark_Call) Run(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string)) *MockUserActivityUseCase_CreateBookmark_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(domain.TrackID), args[3].(time.Duration), args[4].(string))
	})
	return _c
}

func (_c *MockUserActivityUseCase_CreateBookmark_Call) Return(bookmark *domain.Bookmark, err error) *MockUserActivityUseCase_CreateBookmark_Call {
	_c.Call.Return(bookmark, err)
	return _c
}

func (_c *MockUserActivityUseCase_CreateBookmark_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)) *MockUserActivityUseCase_CreateBookmark_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteBookmark provides a mock function for the type MockUserActivityUseCase
func (_mock *MockUserActivityUseCase) DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error {
	ret := _mock.Called(ctx, userID, bookmarkID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteBookmark")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.BookmarkID) error); ok {
		r0 = returnFunc(ctx, userID, bookmarkID)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockUserActivityUseCase_DeleteBookmark_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteBookmark'
type MockUserActivityUseCase_DeleteBookmark_Call struct {
	*mock.Call
}

// DeleteBookmark is a helper method to define mock.On call
//   - ctx
//   - userID
//   - bookmarkID
func (_e *MockUserActivityUseCase_Expecter) DeleteBookmark(ctx interface{}, userID interface{}, bookmarkID interface{}) *MockUserActivityUseCase_DeleteBookmark_Call {
	return &MockUserActivityUseCase_DeleteBookmark_Call{Call: _e.mock.On("DeleteBookmark", ctx, userID, bookmarkID)}
}

func (_c *MockUserActivityUseCase_DeleteBookmark_Call) Run(run func(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID)) *MockUserActivityUseCase_DeleteBookmark_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(domain.BookmarkID))
	})
	return _c
}

func (_c *MockUserActivityUseCase_DeleteBookmark_Call) Return(err error) *MockUserActivityUseCase_DeleteBookmark_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockUserActivityUseCase_DeleteBookmark_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error) *MockUserActivityUseCase_DeleteBookmark_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlaybackProgress provides a mock function for the type MockUserActivityUseCase
func (_mock *MockUserActivityUseCase) GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	ret := _mock.Called(ctx, userID, trackID)

	if len(ret) == 0 {
		panic("no return value specified for GetPlaybackProgress")
	}

	var r0 *domain.PlaybackProgress
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID) (*domain.PlaybackProgress, error)); ok {
		return returnFunc(ctx, userID, trackID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID) *domain.PlaybackProgress); ok {
		r0 = returnFunc(ctx, userID, trackID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.PlaybackProgress)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, domain.TrackID) error); ok {
		r1 = returnFunc(ctx, userID, trackID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserActivityUseCase_GetPlaybackProgress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlaybackProgress'
type MockUserActivityUseCase_GetPlaybackProgress_Call struct {
	*mock.Call
}

// GetPlaybackProgress is a helper method to define mock.On call
//   - ctx
//   - userID
//   - trackID
func (_e *MockUserActivityUseCase_Expecter) GetPlaybackProgress(ctx interface{}, userID interface{}, trackID interface{}) *MockUserActivityUseCase_GetPlaybackProgress_Call {
	return &MockUserActivityUseCase_GetPlaybackProgress_Call{Call: _e.mock.On("GetPlaybackProgress", ctx, userID, trackID)}
}

func (_c *MockUserActivityUseCase_GetPlaybackProgress_Call) Run(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID)) *MockUserActivityUseCase_GetPlaybackProgress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(domain.TrackID))
	})
	return _c
}

func (_c *MockUserActivityUseCase_GetPlaybackProgress_Call) Return(playbackProgress *domain.PlaybackProgress, err error) *MockUserActivityUseCase_GetPlaybackProgress_Call {
	_c.Call.Return(playbackProgress, err)
	return _c
}

func (_c *MockUserActivityUseCase_GetPlaybackProgress_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)) *MockUserActivityUseCase_GetPlaybackProgress_Call {
	_c.Call.Return(run)
	return _c
}

// ListBookmarks provides a mock function for the type MockUserActivityUseCase
func (_mock *MockUserActivityUseCase) ListBookmarks(ctx context.Context, params port.ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error) {
	ret := _mock.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for ListBookmarks")
	}

	var r0 []*domain.Bookmark
	var r1 int
	var r2 pagination.Page
	var r3 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error)); ok {
		return returnFunc(ctx, params)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListBookmarksInput) []*domain.Bookmark); ok {
		r0 = returnFunc(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Bookmark)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, port.ListBookmarksInput) int); ok {
		r1 = returnFunc(ctx, params)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, port.ListBookmarksInput) pagination.Page); ok {
		r2 = returnFunc(ctx, params)
	} else {
		r2 = ret.Get(2).(pagination.Page)
	}
	if returnFunc, ok := ret.Get(3).(func(context.Context, port.ListBookmarksInput) error); ok {
		r3 = returnFunc(ctx, params)
	} else {
		r3 = ret.Error(3)
	}
	return r0, r1, r2, r3
}

// MockUserActivityUseCase_ListBookmarks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListBookmarks'
type MockUserActivityUseCase_ListBookmarks_Call struct {
	*mock.Call
}

// ListBookmarks is a helper method to define mock.On call
//   - ctx
//   - params
func (_e *MockUserActivityUseCase_Expecter) ListBookmarks(ctx interface{}, params interface{}) *MockUserActivityUseCase_ListBookmarks_Call {
	return &MockUserActivityUseCase_ListBookmarks_Call{Call: _e.mock.On("ListBookmarks", ctx, params)}
}

func (_c *MockUserActivityUseCase_ListBookmarks_Call) Run(run func(ctx context.Context, params port.ListBookmarksInput)) *MockUserActivityUseCase_ListBookmarks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(port.ListBookmarksInput))
	})
	return _c
}

func (_c *MockUserActivityUseCase_ListBookmarks_Call) Return(bookmarks []*domain.Bookmark, n int, page pagination.Page, err error) *MockUserActivityUseCase_ListBookmarks_Call {
	_c.Call.Return(bookmarks, n, page, err)
	return _c
}

func (_c *MockUserActivityUseCase_ListBookmarks_Call) RunAndReturn(run func(ctx context.Context, params port.ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error)) *MockUserActivityUseCase_ListBookmarks_Call {
	_c.Call.Return(run)
	return _c
}

// ListUserProgress provides a mock function for the type MockUserActivityUseCase
func (_mock *MockUserActivityUseCase) ListUserProgress(ctx context.Context, params port.ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error) {
	ret := _mock.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for ListUserProgress")
	}

	var r0 []*domain.PlaybackProgress
	var r1 int
	var r2 pagination.Page
	var r3 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error)); ok {
		return returnFunc(ctx, params)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListProgressInput) []*domain.PlaybackProgress); ok {
		r0 = returnFunc(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.PlaybackProgress)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, port.ListProgressInput) int); ok {
		r1 = returnFunc(ctx, params)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, port.ListProgressInput) pagination.Page); ok {
		r2 = returnFunc(ctx, params)
	} else {
		r2 = ret.Get(2).(pagination.Page)
	}
	if returnFunc, ok := ret.Get(3).(func(context.Context, port.ListProgressInput) error); ok {
		r3 = returnFunc(ctx, params)
	} else {
		r3 = ret.Error(3)
	}
	return r0, r1, r2, r3
}

// MockUserActivityUseCase_ListUserProgress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListUserProgress'
type MockUserActivityUseCase_ListUserProgress_Call struct {
	*mock.Call
}

// ListUserProgress is a helper method to define mock.On call
//   - ctx
//   - params
func (_e *MockUserActivityUseCase_Expecter) ListUserProgress(ctx interface{}, params interface{}) *MockUserActivityUseCase_ListUserProgress_Call {
	return &MockUserActivityUseCase_ListUserProgress_Call{Call: _e.mock.On("ListUserProgress", ctx, params)}
}

func (_c *MockUserActivityUseCase_ListUserProgress_Call) Run(run func(ctx context.Context, params port.ListProgressInput)) *MockUserActivityUseCase_ListUserProgress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(port.ListProgressInput))
	})
	return _c
}

func (_c *MockUserActivityUseCase_ListUserProgress_Call) Return(playbackProgresss []*domain.PlaybackProgress, n int, page pagination.Page, err error) *MockUserActivityUseCase_ListUserProgress_Call {
	_c.Call.Return(playbackProgresss, n, page, err)
	return _c
}

func (_c *MockUserActivityUseCase_ListUserProgress_Call) RunAndReturn(run func(ctx context.Context, params port.ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error)) *MockUserActivityUseCase_ListUserProgress_Call {
	_c.Call.Return(run)
	return _c
}

// RecordPlaybackProgress provides a mock function for the type MockUserActivityUseCase
func (_mock *MockUserActivityUseCase) RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error {
	ret := _mock.Called(ctx, userID, trackID, progress)

	if len(ret) == 0 {
		panic("no return value specified for RecordPlaybackProgress")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID, time.Duration) error); ok {
		r0 = returnFunc(ctx, userID, trackID, progress)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockUserActivityUseCase_RecordPlaybackProgress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecordPlaybackProgress'
type MockUserActivityUseCase_RecordPlaybackProgress_Call struct {
	*mock.Call
}

// RecordPlaybackProgress is a helper method to define mock.On call
//   - ctx
//   - userID
//   - trackID
//   - progress
func (_e *MockUserActivityUseCase_Expecter) RecordPlaybackProgress(ctx interface{}, userID interface{}, trackID interface{}, progress interface{}) *MockUserActivityUseCase_RecordPlaybackProgress_Call {
	return &MockUserActivityUseCase_RecordPlaybackProgress_Call{Call: _e.mock.On("RecordPlaybackProgress", ctx, userID, trackID, progress)}
}

func (_c *MockUserActivityUseCase_RecordPlaybackProgress_Call) Run(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration)) *MockUserActivityUseCase_RecordPlaybackProgress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(domain.TrackID), args[3].(time.Duration))
	})
	return _c
}

func (_c *MockUserActivityUseCase_RecordPlaybackProgress_Call) Return(err error) *MockUserActivityUseCase_RecordPlaybackProgress_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockUserActivityUseCase_RecordPlaybackProgress_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error) *MockUserActivityUseCase_RecordPlaybackProgress_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_Tx.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	mock "github.com/stretchr/testify/mock"
)

// NewMockTx creates a new instance of MockTx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTx(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTx {
	mock := &MockTx{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockTx is an autogenerated mock type for the Tx type
type MockTx struct {
	mock.Mock
}

type MockTx_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTx) EXPECT() *MockTx_Expecter {
	return &MockTx_Expecter{mock: &_m.Mock}
}
```

## `internal/mocks/port/mock_UserUseCase.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
)

// NewMockUserUseCase creates a new instance of MockUserUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserUseCase {
	mock := &MockUserUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockUserUseCase is an autogenerated mock type for the UserUseCase type
type MockUserUseCase struct {
	mock.Mock
}

type MockUserUseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserUseCase) EXPECT() *MockUserUseCase_Expecter {
	return &MockUserUseCase_Expecter{mock: &_m.Mock}
}

// GetUserProfile provides a mock function for the type MockUserUseCase
func (_mock *MockUserUseCase) GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	ret := _mock.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetUserProfile")
	}

	var r0 *domain.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID) (*domain.User, error)); ok {
		return returnFunc(ctx, userID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID) *domain.User); ok {
		r0 = returnFunc(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID) error); ok {
		r1 = returnFunc(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserUseCase_GetUserProfile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUserProfile'
type MockUserUseCase_GetUserProfile_Call struct {
	*mock.Call
}

// GetUserProfile is a helper method to define mock.On call
//   - ctx
//   - userID
func (_e *MockUserUseCase_Expecter) GetUserProfile(ctx interface{}, userID interface{}) *MockUserUseCase_GetUserProfile_Call {
	return &MockUserUseCase_GetUserProfile_Call{Call: _e.mock.On("GetUserProfile", ctx, userID)}
}

func (_c *MockUserUseCase_GetUserProfile_Call) Run(run func(ctx context.Context, userID domain.UserID)) *MockUserUseCase_GetUserProfile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID))
	})
	return _c
}

func (_c *MockUserUseCase_GetUserProfile_Call) Return(user *domain.User, err error) *MockUserUseCase_GetUserProfile_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *MockUserUseCase_GetUserProfile_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID) (*domain.User, error)) *MockUserUseCase_GetUserProfile_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_ExternalAuthService.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// NewMockExternalAuthService creates a new instance of MockExternalAuthService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockExternalAuthService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockExternalAuthService {
	mock := &MockExternalAuthService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockExternalAuthService is an autogenerated mock type for the ExternalAuthService type
type MockExternalAuthService struct {
	mock.Mock
}

type MockExternalAuthService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockExternalAuthService) EXPECT() *MockExternalAuthService_Expecter {
	return &MockExternalAuthService_Expecter{mock: &_m.Mock}
}

// VerifyGoogleToken provides a mock function for the type MockExternalAuthService
func (_mock *MockExternalAuthService) VerifyGoogleToken(ctx context.Context, idToken string) (*port.ExternalUserInfo, error) {
	ret := _mock.Called(ctx, idToken)

	if len(ret) == 0 {
		panic("no return value specified for VerifyGoogleToken")
	}

	var r0 *port.ExternalUserInfo
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (*port.ExternalUserInfo, error)); ok {
		return returnFunc(ctx, idToken)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) *port.ExternalUserInfo); ok {
		r0 = returnFunc(ctx, idToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*port.ExternalUserInfo)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, idToken)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockExternalAuthService_VerifyGoogleToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VerifyGoogleToken'
type MockExternalAuthService_VerifyGoogleToken_Call struct {
	*mock.Call
}

// VerifyGoogleToken is a helper method to define mock.On call
//   - ctx
//   - idToken
func (_e *MockExternalAuthService_Expecter) VerifyGoogleToken(ctx interface{}, idToken interface{}) *MockExternalAuthService_VerifyGoogleToken_Call {
	return &MockExternalAuthService_VerifyGoogleToken_Call{Call: _e.mock.On("VerifyGoogleToken", ctx, idToken)}
}

func (_c *MockExternalAuthService_VerifyGoogleToken_Call) Run(run func(ctx context.Context, idToken string)) *MockExternalAuthService_VerifyGoogleToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockExternalAuthService_VerifyGoogleToken_Call) Return(externalUserInfo *port.ExternalUserInfo, err error) *MockExternalAuthService_VerifyGoogleToken_Call {
	_c.Call.Return(externalUserInfo, err)
	return _c
}

func (_c *MockExternalAuthService_VerifyGoogleToken_Call) RunAndReturn(run func(ctx context.Context, idToken string) (*port.ExternalUserInfo, error)) *MockExternalAuthService_VerifyGoogleToken_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_AudioTrackRepository.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// NewMockAudioTrackRepository creates a new instance of MockAudioTrackRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAudioTrackRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAudioTrackRepository {
	mock := &MockAudioTrackRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockAudioTrackRepository is an autogenerated mock type for the AudioTrackRepository type
type MockAudioTrackRepository struct {
	mock.Mock
}

type MockAudioTrackRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAudioTrackRepository) EXPECT() *MockAudioTrackRepository_Expecter {
	return &MockAudioTrackRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) Create(ctx context.Context, track *domain.AudioTrack) error {
	ret := _mock.Called(ctx, track)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.AudioTrack) error); ok {
		r0 = returnFunc(ctx, track)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioTrackRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockAudioTrackRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx
//   - track
func (_e *MockAudioTrackRepository_Expecter) Create(ctx interface{}, track interface{}) *MockAudioTrackRepository_Create_Call {
	return &MockAudioTrackRepository_Create_Call{Call: _e.mock.On("Create", ctx, track)}
}

func (_c *MockAudioTrackRepository_Create_Call) Run(run func(ctx context.Context, track *domain.AudioTrack)) *MockAudioTrackRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.AudioTrack))
	})
	return _c
}

func (_c *MockAudioTrackRepository_Create_Call) Return(err error) *MockAudioTrackRepository_Create_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioTrackRepository_Create_Call) RunAndReturn(run func(ctx context.Context, track *domain.AudioTrack) error) *MockAudioTrackRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) Delete(ctx context.Context, id domain.TrackID) error {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) error); ok {
		r0 = returnFunc(ctx, id)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioTrackRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockAudioTrackRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockAudioTrackRepository_Expecter) Delete(ctx interface{}, id interface{}) *MockAudioTrackRepository_Delete_Call {
	return &MockAudioTrackRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *MockAudioTrackRepository_Delete_Call) Run(run func(ctx context.Context, id domain.TrackID)) *MockAudioTrackRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.TrackID))
	})
	return _c
}

func (_c *MockAudioTrackRepository_Delete_Call) Return(err error) *MockAudioTrackRepository_Delete_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioTrackRepository_Delete_Call) RunAndReturn(run func(ctx context.Context, id domain.TrackID) error) *MockAudioTrackRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) Exists(ctx context.Context, id domain.TrackID) (bool, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) (bool, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) bool); ok {
		r0 = returnFunc(ctx, id)
	} else {
		r0 = ret.Get(0).(bool)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.TrackID) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioTrackRepository_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type MockAudioTrackRepository_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockAudioTrackRepository_Expecter) Exists(ctx interface{}, id interface{}) *MockAudioTrackRepository_Exists_Call {
	return &MockAudioTrackRepository_Exists_Call{Call: _e.mock.On("Exists", ctx, id)}
}

func (_c *MockAudioTrackRepository_Exists_Call) Run(run func(ctx context.Context, id domain.TrackID)) *MockAudioTrackRepository_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.TrackID))
	})
	return _c
}

func (_c *MockAudioTrackRepository_Exists_Call) Return(b bool, err error) *MockAudioTrackRepository_Exists_Call {
	_c.Call.Return(b, err)
	return _c
}

func (_c *MockAudioTrackRepository_Exists_Call) RunAndReturn(run func(ctx context.Context, id domain.TrackID) (bool, error)) *MockAudioTrackRepository_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// FindByID provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for FindByID")
	}

	var r0 *domain.AudioTrack
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) (*domain.AudioTrack, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.TrackID) *domain.AudioTrack); ok {
		r0 = returnFunc(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.AudioTrack)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.TrackID) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioTrackRepository_FindByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByID'
type MockAudioTrackRepository_FindByID_Call struct {
	*mock.Call
}

// FindByID is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockAudioTrackRepository_Expecter) FindByID(ctx interface{}, id interface{}) *MockAudioTrackRepository_FindByID_Call {
	return &MockAudioTrackRepository_FindByID_Call{Call: _e.mock.On("FindByID", ctx, id)}
}

func (_c *MockAudioTrackRepository_FindByID_Call) Run(run func(ctx context.Context, id domain.TrackID)) *MockAudioTrackRepository_FindByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.TrackID))
	})
	return _c
}

func (_c *MockAudioTrackRepository_FindByID_Call) Return(audioTrack *domain.AudioTrack, err error) *MockAudioTrackRepository_FindByID_Call {
	_c.Call.Return(audioTrack, err)
	return _c
}

func (_c *MockAudioTrackRepository_FindByID_Call) RunAndReturn(run func(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error)) *MockAudioTrackRepository_FindByID_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) List(ctx context.Context, filters port.ListTracksFilters, page pagination.Page) ([]*domain.AudioTrack, int, error) {
	ret := _mock.Called(ctx, filters, page)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*domain.AudioTrack
	var r1 int
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListTracksFilters, pagination.Page) ([]*domain.AudioTrack, int, error)); ok {
		return returnFunc(ctx, filters, page)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, port.ListTracksFilters, pagination.Page) []*domain.AudioTrack); ok {
		r0 = returnFunc(ctx, filters, page)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.AudioTrack)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, port.ListTracksFilters, pagination.Page) int); ok {
		r1 = returnFunc(ctx, filters, page)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, port.ListTracksFilters, pagination.Page) error); ok {
		r2 = returnFunc(ctx, filters, page)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockAudioTrackRepository_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type MockAudioTrackRepository_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx
//   - filters
//   - page
func (_e *MockAudioTrackRepository_Expecter) List(ctx interface{}, filters interface{}, page interface{}) *MockAudioTrackRepository_List_Call {
	return &MockAudioTrackRepository_List_Call{Call: _e.mock.On("List", ctx, filters, page)}
}

func (_c *MockAudioTrackRepository_List_Call) Run(run func(ctx context.Context, filters port.ListTracksFilters, page pagination.Page)) *MockAudioTrackRepository_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(port.ListTracksFilters), args[2].(pagination.Page))
	})
	return _c
}

func (_c *MockAudioTrackRepository_List_Call) Return(tracks []*domain.AudioTrack, total int, err error) *MockAudioTrackRepository_List_Call {
	_c.Call.Return(tracks, total, err)
	return _c
}

func (_c *MockAudioTrackRepository_List_Call) RunAndReturn(run func(ctx context.Context, filters port.ListTracksFilters, page pagination.Page) ([]*domain.AudioTrack, int, error)) *MockAudioTrackRepository_List_Call {
	_c.Call.Return(run)
	return _c
}

// ListByIDs provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error) {
	ret := _mock.Called(ctx, ids)

	if len(ret) == 0 {
		panic("no return value specified for ListByIDs")
	}

	var r0 []*domain.AudioTrack
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, []domain.TrackID) ([]*domain.AudioTrack, error)); ok {
		return returnFunc(ctx, ids)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, []domain.TrackID) []*domain.AudioTrack); ok {
		r0 = returnFunc(ctx, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.AudioTrack)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, []domain.TrackID) error); ok {
		r1 = returnFunc(ctx, ids)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockAudioTrackRepository_ListByIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByIDs'
type MockAudioTrackRepository_ListByIDs_Call struct {
	*mock.Call
}

// ListByIDs is a helper method to define mock.On call
//   - ctx
//   - ids
func (_e *MockAudioTrackRepository_Expecter) ListByIDs(ctx interface{}, ids interface{}) *MockAudioTrackRepository_ListByIDs_Call {
	return &MockAudioTrackRepository_ListByIDs_Call{Call: _e.mock.On("ListByIDs", ctx, ids)}
}

func (_c *MockAudioTrackRepository_ListByIDs_Call) Run(run func(ctx context.Context, ids []domain.TrackID)) *MockAudioTrackRepository_ListByIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]domain.TrackID))
	})
	return _c
}

func (_c *MockAudioTrackRepository_ListByIDs_Call) Return(audioTracks []*domain.AudioTrack, err error) *MockAudioTrackRepository_ListByIDs_Call {
	_c.Call.Return(audioTracks, err)
	return _c
}

func (_c *MockAudioTrackRepository_ListByIDs_Call) RunAndReturn(run func(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error)) *MockAudioTrackRepository_ListByIDs_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function for the type MockAudioTrackRepository
func (_mock *MockAudioTrackRepository) Update(ctx context.Context, track *domain.AudioTrack) error {
	ret := _mock.Called(ctx, track)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.AudioTrack) error); ok {
		r0 = returnFunc(ctx, track)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockAudioTrackRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockAudioTrackRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx
//   - track
func (_e *MockAudioTrackRepository_Expecter) Update(ctx interface{}, track interface{}) *MockAudioTrackRepository_Update_Call {
	return &MockAudioTrackRepository_Update_Call{Call: _e.mock.On("Update", ctx, track)}
}

func (_c *MockAudioTrackRepository_Update_Call) Run(run func(ctx context.Context, track *domain.AudioTrack)) *MockAudioTrackRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.AudioTrack))
	})
	return _c
}

func (_c *MockAudioTrackRepository_Update_Call) Return(err error) *MockAudioTrackRepository_Update_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockAudioTrackRepository_Update_Call) RunAndReturn(run func(ctx context.Context, track *domain.AudioTrack) error) *MockAudioTrackRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/mocks/port/mock_BookmarkRepository.go`

```go
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// NewMockBookmarkRepository creates a new instance of MockBookmarkRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockBookmarkRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBookmarkRepository {
	mock := &MockBookmarkRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockBookmarkRepository is an autogenerated mock type for the BookmarkRepository type
type MockBookmarkRepository struct {
	mock.Mock
}

type MockBookmarkRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockBookmarkRepository) EXPECT() *MockBookmarkRepository_Expecter {
	return &MockBookmarkRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function for the type MockBookmarkRepository
func (_mock *MockBookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	ret := _mock.Called(ctx, bookmark)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, *domain.Bookmark) error); ok {
		r0 = returnFunc(ctx, bookmark)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockBookmarkRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockBookmarkRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx
//   - bookmark
func (_e *MockBookmarkRepository_Expecter) Create(ctx interface{}, bookmark interface{}) *MockBookmarkRepository_Create_Call {
	return &MockBookmarkRepository_Create_Call{Call: _e.mock.On("Create", ctx, bookmark)}
}

func (_c *MockBookmarkRepository_Create_Call) Run(run func(ctx context.Context, bookmark *domain.Bookmark)) *MockBookmarkRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.Bookmark))
	})
	return _c
}

func (_c *MockBookmarkRepository_Create_Call) Return(err error) *MockBookmarkRepository_Create_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockBookmarkRepository_Create_Call) RunAndReturn(run func(ctx context.Context, bookmark *domain.Bookmark) error) *MockBookmarkRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function for the type MockBookmarkRepository
func (_mock *MockBookmarkRepository) Delete(ctx context.Context, id domain.BookmarkID) error {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.BookmarkID) error); ok {
		r0 = returnFunc(ctx, id)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockBookmarkRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockBookmarkRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockBookmarkRepository_Expecter) Delete(ctx interface{}, id interface{}) *MockBookmarkRepository_Delete_Call {
	return &MockBookmarkRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *MockBookmarkRepository_Delete_Call) Run(run func(ctx context.Context, id domain.BookmarkID)) *MockBookmarkRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.BookmarkID))
	})
	return _c
}

func (_c *MockBookmarkRepository_Delete_Call) Return(err error) *MockBookmarkRepository_Delete_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockBookmarkRepository_Delete_Call) RunAndReturn(run func(ctx context.Context, id domain.BookmarkID) error) *MockBookmarkRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// FindByID provides a mock function for the type MockBookmarkRepository
func (_mock *MockBookmarkRepository) FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for FindByID")
	}

	var r0 *domain.Bookmark
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.BookmarkID) (*domain.Bookmark, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.BookmarkID) *domain.Bookmark); ok {
		r0 = returnFunc(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Bookmark)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.BookmarkID) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockBookmarkRepository_FindByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByID'
type MockBookmarkRepository_FindByID_Call struct {
	*mock.Call
}

// FindByID is a helper method to define mock.On call
//   - ctx
//   - id
func (_e *MockBookmarkRepository_Expecter) FindByID(ctx interface{}, id interface{}) *MockBookmarkRepository_FindByID_Call {
	return &MockBookmarkRepository_FindByID_Call{Call: _e.mock.On("FindByID", ctx, id)}
}

func (_c *MockBookmarkRepository_FindByID_Call) Run(run func(ctx context.Context, id domain.BookmarkID)) *MockBookmarkRepository_FindByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.BookmarkID))
	})
	return _c
}

func (_c *MockBookmarkRepository_FindByID_Call) Return(bookmark *domain.Bookmark, err error) *MockBookmarkRepository_FindByID_Call {
	_c.Call.Return(bookmark, err)
	return _c
}

func (_c *MockBookmarkRepository_FindByID_Call) RunAndReturn(run func(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error)) *MockBookmarkRepository_FindByID_Call {
	_c.Call.Return(run)
	return _c
}

// ListByUser provides a mock function for the type MockBookmarkRepository
func (_mock *MockBookmarkRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error) {
	ret := _mock.Called(ctx, userID, page)

	if len(ret) == 0 {
		panic("no return value specified for ListByUser")
	}

	var r0 []*domain.Bookmark
	var r1 int
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, pagination.Page) ([]*domain.Bookmark, int, error)); ok {
		return returnFunc(ctx, userID, page)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, pagination.Page) []*domain.Bookmark); ok {
		r0 = returnFunc(ctx, userID, page)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Bookmark)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, pagination.Page) int); ok {
		r1 = returnFunc(ctx, userID, page)
	} else {
		r1 = ret.Get(1).(int)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, domain.UserID, pagination.Page) error); ok {
		r2 = returnFunc(ctx, userID, page)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockBookmarkRepository_ListByUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByUser'
type MockBookmarkRepository_ListByUser_Call struct {
	*mock.Call
}

// ListByUser is a helper method to define mock.On call
//   - ctx
//   - userID
//   - page
func (_e *MockBookmarkRepository_Expecter) ListByUser(ctx interface{}, userID interface{}, page interface{}) *MockBookmarkRepository_ListByUser_Call {
	return &MockBookmarkRepository_ListByUser_Call{Call: _e.mock.On("ListByUser", ctx, userID, page)}
}

func (_c *MockBookmarkRepository_ListByUser_Call) Run(run func(ctx context.Context, userID domain.UserID, page pagination.Page)) *MockBookmarkRepository_ListByUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(pagination.Page))
	})
	return _c
}

func (_c *MockBookmarkRepository_ListByUser_Call) Return(bookmarks []*domain.Bookmark, total int, err error) *MockBookmarkRepository_ListByUser_Call {
	_c.Call.Return(bookmarks, total, err)
	return _c
}

func (_c *MockBookmarkRepository_ListByUser_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error)) *MockBookmarkRepository_ListByUser_Call {
	_c.Call.Return(run)
	return _c
}

// ListByUserAndTrack provides a mock function for the type MockBookmarkRepository
func (_mock *MockBookmarkRepository) ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error) {
	ret := _mock.Called(ctx, userID, trackID)

	if len(ret) == 0 {
		panic("no return value specified for ListByUserAndTrack")
	}

	var r0 []*domain.Bookmark
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID) ([]*domain.Bookmark, error)); ok {
		return returnFunc(ctx, userID, trackID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, domain.UserID, domain.TrackID) []*domain.Bookmark); ok {
		r0 = returnFunc(ctx, userID, trackID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Bookmark)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, domain.UserID, domain.TrackID) error); ok {
		r1 = returnFunc(ctx, userID, trackID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockBookmarkRepository_ListByUserAndTrack_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByUserAndTrack'
type MockBookmarkRepository_ListByUserAndTrack_Call struct {
	*mock.Call
}

// ListByUserAndTrack is a helper method to define mock.On call
//   - ctx
//   - userID
//   - trackID
func (_e *MockBookmarkRepository_Expecter) ListByUserAndTrack(ctx interface{}, userID interface{}, trackID interface{}) *MockBookmarkRepository_ListByUserAndTrack_Call {
	return &MockBookmarkRepository_ListByUserAndTrack_Call{Call: _e.mock.On("ListByUserAndTrack", ctx, userID, trackID)}
}

func (_c *MockBookmarkRepository_ListByUserAndTrack_Call) Run(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID)) *MockBookmarkRepository_ListByUserAndTrack_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(domain.UserID), args[2].(domain.TrackID))
	})
	return _c
}

func (_c *MockBookmarkRepository_ListByUserAndTrack_Call) Return(bookmarks []*domain.Bookmark, err error) *MockBookmarkRepository_ListByUserAndTrack_Call {
	_c.Call.Return(bookmarks, err)
	return _c
}

func (_c *MockBookmarkRepository_ListByUserAndTrack_Call) RunAndReturn(run func(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error)) *MockBookmarkRepository_ListByUserAndTrack_Call {
	_c.Call.Return(run)
	return _c
}
```

## `internal/config/config.go`

```go
// ============================================
// FILE: internal/config/config.go (MODIFIED)
// ============================================
package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Minio    MinioConfig    `mapstructure:"minio"`
	Google   GoogleConfig   `mapstructure:"google"`
	Log      LogConfig      `mapstructure:"log"`
	Cors     CorsConfig     `mapstructure:"cors"`
	CDN      CDNConfig      `mapstructure:"cdn"`
}

// ServerConfig holds server specific configuration.
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
}

// JWTConfig holds JWT related configuration.
type JWTConfig struct {
	SecretKey          string        `mapstructure:"secretKey"`
	AccessTokenExpiry  time.Duration `mapstructure:"accessTokenExpiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refreshTokenExpiry"` // ADDED
}

// MinioConfig holds MinIO connection configuration.
type MinioConfig struct {
	Endpoint        string        `mapstructure:"endpoint"`
	AccessKeyID     string        `mapstructure:"accessKeyId"`
	SecretAccessKey string        `mapstructure:"secretAccessKey"`
	UseSSL          bool          `mapstructure:"useSsl"`
	BucketName      string        `mapstructure:"bucketName"`
	PresignExpiry   time.Duration `mapstructure:"presignExpiry"`
}

// GoogleConfig holds Google OAuth configuration.
type GoogleConfig struct {
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
	JSON  bool   `mapstructure:"json"`
}

// CorsConfig holds CORS configuration.
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins"`
	AllowedMethods   []string `mapstructure:"allowedMethods"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"`
}

// CDNConfig holds optional CDN configuration.
type CDNConfig struct {
	BaseURL string `mapstructure:"baseUrl"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	setDefaultValues(v) // Set defaults first

	v.AddConfigPath(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigName("config")
	if errReadBase := v.ReadInConfig(); errReadBase != nil {
		if _, ok := errReadBase.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("error reading base config file 'config.yaml': %w", errReadBase)
		}
		fmt.Fprintln(os.Stderr, "Info: Base config file 'config.yaml' not found. Skipping.")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	if errReadEnv := v.MergeInConfig(); errReadEnv != nil {
		if _, ok := errReadEnv.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "Info: Environment config file 'config.%s.yaml' not found. Relying on base/defaults/env.\n", env)
		} else {
			return config, fmt.Errorf("error merging config file 'config.%s.yaml': %w", env, errReadEnv)
		}
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if config.Database.DSN == "" {
		envDSN := os.Getenv("DATABASE_URL")
		if envDSN != "" {
			fmt.Fprintln(os.Stderr, "Info: Database DSN not found in config, using DATABASE_URL environment variable.")
			config.Database.DSN = envDSN
		} else {
			return config, fmt.Errorf("database DSN is required but not found in config or DATABASE_URL environment variable")
		}
	}

	if config.CDN.BaseURL != "" {
		_, parseErr := url.ParseRequestURI(config.CDN.BaseURL)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid CDN BaseURL configured ('%s'): %v. CDN rewriting will be disabled.\n", config.CDN.BaseURL, parseErr)
			config.CDN.BaseURL = ""
		}
	}

	// Validate token expirations
	if config.JWT.AccessTokenExpiry <= 0 {
		return config, fmt.Errorf("jwt.accessTokenExpiry must be a positive duration")
	}
	if config.JWT.RefreshTokenExpiry <= 0 {
		return config, fmt.Errorf("jwt.refreshTokenExpiry must be a positive duration")
	}
	if config.JWT.RefreshTokenExpiry <= config.JWT.AccessTokenExpiry {
		return config, fmt.Errorf("jwt.refreshTokenExpiry must be longer than jwt.accessTokenExpiry")
	}

	return config, nil
}

func setDefaultValues(v *viper.Viper) {
	// Server Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.readTimeout", "5s")
	v.SetDefault("server.writeTimeout", "10s")
	v.SetDefault("server.idleTimeout", "120s")

	// Database Defaults
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 25)
	v.SetDefault("database.connMaxLifetime", "5m")
	v.SetDefault("database.connMaxIdleTime", "5m")

	// JWT Defaults
	v.SetDefault("jwt.secretKey", "default-insecure-secret-key-please-override")
	v.SetDefault("jwt.accessTokenExpiry", "1h")
	v.SetDefault("jwt.refreshTokenExpiry", "720h") // Default: 30 days (ADDED)

	// MinIO Defaults
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.accessKeyId", "minioadmin")
	v.SetDefault("minio.secretAccessKey", "minioadmin")
	v.SetDefault("minio.useSsl", false)
	v.SetDefault("minio.bucketName", "language-audio")
	v.SetDefault("minio.presignExpiry", "1h")

	// Google Defaults
	v.SetDefault("google.clientId", "")
	v.SetDefault("google.clientSecret", "")

	// Log Defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.json", false)

	// CORS Defaults
	v.SetDefault("cors.allowedOrigins", []string{"http://localhost:3000", "http://127.0.0.1:3000"})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	v.SetDefault("cors.allowCredentials", true)
	v.SetDefault("cors.maxAge", 300)

	// CDN Default
	v.SetDefault("cdn.baseUrl", "")
}

func GetConfig() (Config, error) {
	return LoadConfig(".")
}
```

## `internal/adapter/handler/http/user_activity_handler.go`

```go
// internal/adapter/handler/http/user_activity_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// UserActivityHandler handles HTTP requests related to user progress and bookmarks.
type UserActivityHandler struct {
	activityUseCase port.UserActivityUseCase // 使用port包中定义的接口
	validator       *validation.Validator
}

// NewUserActivityHandler creates a new UserActivityHandler.
func NewUserActivityHandler(uc port.UserActivityUseCase, v *validation.Validator) *UserActivityHandler {
	return &UserActivityHandler{
		activityUseCase: uc,
		validator:       v,
	}
}

// --- Progress Handlers ---

// RecordProgress handles POST /api/v1/users/me/progress
// @Summary Record playback progress
// @Description Records or updates the playback progress for a specific audio track for the authenticated user.
// @ID record-playback-progress
// @Tags User Activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param progress body dto.RecordProgressRequestDTO true "Playback progress details (progressMs in milliseconds)"
// @Success 204 "Progress recorded successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 404 {object} httputil.ErrorResponseDTO "Track Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/progress [post]
func (h *UserActivityHandler) RecordProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.RecordProgressRequestDTO // DTO now uses ProgressMs
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	trackID, err := domain.TrackIDFromString(req.TrackID)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	// Convert milliseconds (int64) to duration
	progressDuration := time.Duration(req.ProgressMs) * time.Millisecond

	err = h.activityUseCase.RecordPlaybackProgress(r.Context(), userID, trackID, progressDuration)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound (track), internal errors
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content is suitable for successful upsert
}

// GetProgress handles GET /api/v1/users/me/progress/{trackId}
// @Summary Get playback progress for a track
// @Description Retrieves the playback progress for a specific audio track for the authenticated user.
// @ID get-playback-progress
// @Tags User Activity
// @Produce json
// @Security BearerAuth
// @Param trackId path string true "Audio Track UUID" Format(uuid)
// @Success 200 {object} dto.PlaybackProgressResponseDTO "Playback progress found (progressMs in milliseconds)"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 404 {object} httputil.ErrorResponseDTO "Progress Not Found (or Track Not Found)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/progress/{trackId} [get]
func (h *UserActivityHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	trackIDStr := chi.URLParam(r, "trackId")
	trackID, err := domain.TrackIDFromString(trackIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	progress, err := h.activityUseCase.GetPlaybackProgress(r.Context(), userID, trackID)
	if err != nil {
		// Handles ErrNotFound correctly via RespondError
		httputil.RespondError(w, r, err)
		return
	}

	resp := dto.MapDomainProgressToResponseDTO(progress) // DTO mapping updated for milliseconds
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// ListProgress handles GET /api/v1/users/me/progress
// @Summary List user's playback progress
// @Description Retrieves a paginated list of playback progress records for the authenticated user.
// @ID list-playback-progress
// @Tags User Activity
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Pagination limit" default(50) minimum(1) maximum(100)
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Success 200 {object} dto.PaginatedResponseDTO{data=[]dto.PlaybackProgressResponseDTO} "Paginated list of playback progress (progressMs in milliseconds)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/progress [get]
func (h *UserActivityHandler) ListProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))   // Use 0 if parsing fails
	offset, _ := strconv.Atoi(q.Get("offset")) // Use 0 if parsing fails

	// Create pagination parameters and apply defaults/constraints
	pageParams := pagination.NewPageFromOffset(limit, offset)

	// Create use case parameters struct
	ucParams := port.ListProgressInput{
		UserID: userID,
		Page:   pageParams,
	}

	// Call use case with the params struct
	progressList, total, actualPageInfo, err := h.activityUseCase.ListUserProgress(r.Context(), ucParams)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.PlaybackProgressResponseDTO, len(progressList))
	for i, p := range progressList {
		respData[i] = dto.MapDomainProgressToResponseDTO(p) // DTO mapping updated for milliseconds
	}

	// Use the returned actualPageInfo for the response
	paginatedResult := pagination.NewPaginatedResponse(respData, total, actualPageInfo)

	// Use the generic PaginatedResponseDTO from the DTO package
	resp := dto.PaginatedResponseDTO{
		Data:       paginatedResult.Data,
		Total:      paginatedResult.Total,
		Limit:      paginatedResult.Limit,
		Offset:     paginatedResult.Offset,
		Page:       paginatedResult.Page,
		TotalPages: paginatedResult.TotalPages,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// --- Bookmark Handlers ---

// CreateBookmark handles POST /api/v1/users/me/bookmarks
// @Summary Create a bookmark
// @Description Creates a new bookmark at a specific timestamp (in milliseconds) in an audio track for the authenticated user.
// @ID create-bookmark
// @Tags User Activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bookmark body dto.CreateBookmarkRequestDTO true "Bookmark details (timestampMs in milliseconds)"
// @Success 201 {object} dto.BookmarkResponseDTO "Bookmark created successfully (timestampMs in milliseconds)"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 404 {object} httputil.ErrorResponseDTO "Track Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/bookmarks [post]
func (h *UserActivityHandler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.CreateBookmarkRequestDTO // DTO now uses TimestampMs
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	trackID, err := domain.TrackIDFromString(req.TrackID)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	// Convert milliseconds (int64) to duration
	timestampDuration := time.Duration(req.TimestampMs) * time.Millisecond

	bookmark, err := h.activityUseCase.CreateBookmark(r.Context(), userID, trackID, timestampDuration, req.Note)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound (track), internal errors
		return
	}

	resp := dto.MapDomainBookmarkToResponseDTO(bookmark) // DTO mapping updated for milliseconds
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // 201 Created
}

// ListBookmarks handles GET /api/v1/users/me/bookmarks
// @Summary List user's bookmarks
// @Description Retrieves a paginated list of bookmarks for the authenticated user, optionally filtered by track ID.
// @ID list-bookmarks
// @Tags User Activity
// @Produce json
// @Security BearerAuth
// @Param trackId query string false "Filter by Audio Track UUID" Format(uuid)
// @Param limit query int false "Pagination limit" default(50) minimum(1) maximum(100)
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Success 200 {object} dto.PaginatedResponseDTO{data=[]dto.BookmarkResponseDTO} "Paginated list of bookmarks (timestampMs in milliseconds)"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format (if provided)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/bookmarks [get]
func (h *UserActivityHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	// Create pagination parameters
	pageParams := pagination.NewPageFromOffset(limit, offset)

	// Check for optional trackId filter
	var trackIDFilter *domain.TrackID
	if trackIDStr := q.Get("trackId"); trackIDStr != "" {
		tid, err := domain.TrackIDFromString(trackIDStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid trackId query parameter format", domain.ErrInvalidArgument))
			return
		}
		trackIDFilter = &tid
	}

	// Create use case parameters struct
	ucParams := port.ListBookmarksInput{
		UserID:        userID,
		TrackIDFilter: trackIDFilter,
		Page:          pageParams,
	}

	// Call use case with the params struct
	bookmarks, total, actualPageInfo, err := h.activityUseCase.ListBookmarks(r.Context(), ucParams)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.BookmarkResponseDTO, len(bookmarks))
	for i, b := range bookmarks {
		respData[i] = dto.MapDomainBookmarkToResponseDTO(b) // DTO mapping updated for milliseconds
	}

	// Use the returned actualPageInfo for the response
	paginatedResult := pagination.NewPaginatedResponse(respData, total, actualPageInfo)

	// Use the generic PaginatedResponseDTO from the DTO package
	resp := dto.PaginatedResponseDTO{
		Data:       paginatedResult.Data,
		Total:      paginatedResult.Total,
		Limit:      paginatedResult.Limit,
		Offset:     paginatedResult.Offset,
		Page:       paginatedResult.Page,
		TotalPages: paginatedResult.TotalPages,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// DeleteBookmark handles DELETE /api/v1/users/me/bookmarks/{bookmarkId}
// @Summary Delete a bookmark
// @Description Deletes a specific bookmark owned by the current user.
// @ID delete-bookmark
// @Tags User Activity
// @Produce json
// @Security BearerAuth
// @Param bookmarkId path string true "Bookmark UUID" Format(uuid)
// @Success 204 "Bookmark deleted successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Bookmark Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/bookmarks/{bookmarkId} [delete]
func (h *UserActivityHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	bookmarkIDStr := chi.URLParam(r, "bookmarkId")
	bookmarkID, err := domain.BookmarkIDFromString(bookmarkIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid bookmark ID format", domain.ErrInvalidArgument))
		return
	}

	err = h.activityUseCase.DeleteBookmark(r.Context(), userID, bookmarkID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
```

## `internal/adapter/handler/http/auth_handler.go`

```go
// ==========================================================
// FILE: internal/adapter/handler/http/auth_handler.go (MODIFIED)
// ==========================================================
package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	authUseCase port.AuthUseCase
	validator   *validation.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc port.AuthUseCase, v *validation.Validator) *AuthHandler {
	return &AuthHandler{
		authUseCase: uc,
		validator:   v,
	}
}

// mapAuthResultToDTO maps the use case AuthResult to the response DTO.
func mapAuthResultToDTO(result port.AuthResult) dto.AuthResponseDTO {
	resp := dto.AuthResponseDTO{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
	if result.IsNewUser { // Only include if true
		isNewPtr := true
		resp.IsNewUser = &isNewPtr
	}
	return resp
}

// Register handles user registration requests.
// @Summary Register a new user
// @Description Registers a new user account using email and password.
// @ID register-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequestDTO true "User Registration Info"
// @Success 201 {object} dto.AuthResponseDTO "Registration successful, returns access and refresh tokens" // MODIFIED DESCRIPTION
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict - Email Exists"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// MODIFIED: Use case returns AuthResult
	_, authResult, err := h.authUseCase.RegisterWithPassword(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// MODIFIED: Map AuthResult to DTO
	resp := mapAuthResultToDTO(authResult)
	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}

// Login handles user login requests.
// @Summary Login a user
// @Description Authenticates a user with email and password, returns access and refresh tokens. // MODIFIED DESCRIPTION
// @ID login-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body dto.LoginRequestDTO true "User Login Credentials"
// @Success 200 {object} dto.AuthResponseDTO "Login successful, returns access and refresh tokens" // MODIFIED DESCRIPTION
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"
// @Failure 401 {object} httputil.ErrorResponseDTO "Authentication Failed"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// MODIFIED: Use case returns AuthResult
	authResult, err := h.authUseCase.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// MODIFIED: Map AuthResult to DTO
	resp := mapAuthResultToDTO(authResult)
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// GoogleCallback handles the callback from Google OAuth flow.
// @Summary Handle Google OAuth callback
// @Description Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning access and refresh tokens. // MODIFIED DESCRIPTION
// @ID google-callback
// @Tags Authentication
// @Accept json
// @Produce json
// @Param googleCallback body dto.GoogleCallbackRequestDTO true "Google ID Token"
// @Success 200 {object} dto.AuthResponseDTO "Authentication successful, returns access/refresh tokens. isNewUser indicates new account creation." // MODIFIED DESCRIPTION
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (Missing or Invalid ID Token)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Authentication Failed (Invalid Google Token)"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict - Email already exists with a different login method"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/google/callback [post]
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req dto.GoogleCallbackRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// MODIFIED: Use case returns AuthResult
	authResult, err := h.authUseCase.AuthenticateWithGoogle(r.Context(), req.IDToken)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// MODIFIED: Map AuthResult to DTO
	resp := mapAuthResultToDTO(authResult)
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// Refresh handles token refresh requests.
// @Summary Refresh access token
// @Description Provides a valid refresh token to get a new pair of access and refresh tokens.
// @ID refresh-token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh body dto.RefreshRequestDTO true "Refresh Token"
// @Success 200 {object} dto.AuthResponseDTO "Tokens refreshed successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (Missing Refresh Token)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Authentication Failed (Invalid or Expired Refresh Token)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/refresh [post]
// ADDED METHOD
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	authResult, err := h.authUseCase.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		// Use case should return domain.ErrAuthenticationFailed for invalid/expired tokens
		httputil.RespondError(w, r, err)
		return
	}

	resp := mapAuthResultToDTO(authResult) // Reuse mapping
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// Logout handles user logout requests by invalidating the refresh token.
// @Summary Logout user
// @Description Invalidates the provided refresh token, effectively logging the user out of that session/device.
// @ID logout-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param logout body dto.LogoutRequestDTO true "Refresh Token to invalidate"
// @Success 204 "Logout successful"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (Missing Refresh Token)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/logout [post]
// ADDED METHOD
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	err := h.authUseCase.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		// Logout use case should generally not return client errors unless input is bad,
		// but handle potential internal errors.
		httputil.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful logout
}
```

## `internal/adapter/handler/http/audio_handler.go`

```go
// internal/adapter/handler/http/audio_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings" // Import strings

	"github.com/go-chi/chi/v5"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto" // Import for GetUserIDFromContext
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// AudioHandler handles HTTP requests related to audio tracks and collections.
type AudioHandler struct {
	audioUseCase port.AudioContentUseCase
	validator    *validation.Validator
	// No longer needs activityUseCase here, logic moved into audioUseCase
}

// NewAudioHandler creates a new AudioHandler.
func NewAudioHandler(uc port.AudioContentUseCase, v *validation.Validator) *AudioHandler {
	return &AudioHandler{
		audioUseCase: uc,
		validator:    v,
	}
}

// --- Track Handlers ---

// GetTrackDetails handles GET /api/v1/audio/tracks/{trackId}
// @Summary Get audio track details
// @Description Retrieves details for a specific audio track, including metadata, playback URL, and user-specific progress/bookmarks if authenticated.
// @ID get-track-details
// @Tags Audio Tracks
// @Produce json
// @Param trackId path string true "Audio Track UUID" Format(uuid)
// @Security BearerAuth // Optional: Indicate that auth affects the response (user data)
// @Success 200 {object} dto.AudioTrackDetailsResponseDTO "Audio track details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized (if accessing private track without auth)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Track Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks/{trackId} [get]
func (h *AudioHandler) GetTrackDetails(w http.ResponseWriter, r *http.Request) {
	trackIDStr := chi.URLParam(r, "trackId")
	trackID, err := domain.TrackIDFromString(trackIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	// Usecase now returns a result struct containing track, URL, and optional user data
	result, err := h.audioUseCase.GetAudioTrackDetails(r.Context(), trackID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated, internal errors
		return
	}

	// Point 4: Mapping moved here, using a mapper function in dto package
	resp := dto.MapDomainTrackToDetailsResponseDTO(result) // Pass the result struct to the mapper

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// Point 6: Refactored parameter parsing into helper function parseListTracksInput
func parseListTracksInput(r *http.Request) (port.ListTracksInput, error) {
	q := r.URL.Query()
	input := port.ListTracksInput{} // Use usecase input struct

	// Parse simple string filters
	if query := q.Get("q"); query != "" {
		input.Query = &query
	}
	if lang := q.Get("lang"); lang != "" {
		input.LanguageCode = &lang
	}
	input.Tags = q["tags"]

	// Parse Level (and validate)
	if levelStr := q.Get("level"); levelStr != "" {
		level := domain.AudioLevel(strings.ToUpper(levelStr)) // Normalize case
		if level.IsValid() {
			input.Level = &level
		} else {
			return input, fmt.Errorf("%w: invalid level query parameter '%s'", domain.ErrInvalidArgument, levelStr)
		}
	}

	// Parse isPublic (boolean)
	if isPublicStr := q.Get("isPublic"); isPublicStr != "" {
		isPublic, err := strconv.ParseBool(isPublicStr)
		if err == nil {
			input.IsPublic = &isPublic
		} else {
			return input, fmt.Errorf("%w: invalid isPublic query parameter (must be true or false)", domain.ErrInvalidArgument)
		}
	}

	// Parse Sort parameters
	input.SortBy = q.Get("sortBy")
	input.SortDirection = q.Get("sortDir")

	// Parse Pagination parameters
	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")
	limit, errLimit := strconv.Atoi(limitStr)
	offset, errOffset := strconv.Atoi(offsetStr)
	if (limitStr != "" && errLimit != nil) || (offsetStr != "" && errOffset != nil) {
		return input, fmt.Errorf("%w: invalid limit or offset query parameter", domain.ErrInvalidArgument)
	}
	// Let the usecase/pagination package handle defaults/constraints for limit/offset
	input.Page = pagination.Page{Limit: limit, Offset: offset}

	return input, nil
}

// ListTracks handles GET /api/v1/audio/tracks
// @Summary List audio tracks
// @Description Retrieves a paginated list of audio tracks, supporting filtering and sorting.
// @ID list-audio-tracks
// @Tags Audio Tracks
// @Produce json
// @Param q query string false "Search query (searches title, description)"
// @Param lang query string false "Filter by language code (e.g., en-US)"
// @Param level query string false "Filter by audio level (e.g., A1, B2)" Enums(A1, A2, B1, B2, C1, C2, NATIVE)
// @Param isPublic query boolean false "Filter by public status (true or false)"
// @Param tags query []string false "Filter by tags (e.g., ?tags=news&tags=podcast)" collectionFormat(multi)
// @Param sortBy query string false "Sort field (e.g., createdAt, title, durationMs)" default(createdAt)
// @Param sortDir query string false "Sort direction (asc or desc)" default(desc) Enums(asc, desc)
// @Param limit query int false "Pagination limit" default(20) minimum(1) maximum(100)
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Success 200 {object} dto.PaginatedResponseDTO{data=[]dto.AudioTrackResponseDTO} "Paginated list of audio tracks"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Query Parameter Format"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [get]
func (h *AudioHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	// Point 6: Use helper function for parsing
	ucInput, err := parseListTracksInput(r)
	if err != nil {
		httputil.RespondError(w, r, err) // Return 400 Bad Request if parsing fails
		return
	}

	// Point 5: Call use case with the ListTracksInput struct
	tracks, total, actualPageInfo, err := h.audioUseCase.ListTracks(r.Context(), ucInput)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.AudioTrackResponseDTO, len(tracks))
	for i, track := range tracks {
		respData[i] = dto.MapDomainTrackToResponseDTO(track) // Point 1: DTO mapping uses ms
	}

	// Use the pagination info returned by the usecase (which applied constraints)
	paginatedResult := pagination.NewPaginatedResponse(respData, total, actualPageInfo)
	// Map to the common DTO structure
	resp := dto.PaginatedResponseDTO{
		Data:       paginatedResult.Data,
		Total:      paginatedResult.Total,
		Limit:      paginatedResult.Limit,
		Offset:     paginatedResult.Offset,
		Page:       paginatedResult.Page,
		TotalPages: paginatedResult.TotalPages,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// --- Collection Handlers ---

// CreateCollection handles POST /api/v1/audio/collections
// @Summary Create an audio collection
// @Description Creates a new audio collection (playlist or course) for the authenticated user.
// @ID create-audio-collection
// @Tags Audio Collections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param collection body dto.CreateCollectionRequestDTO true "Collection details"
// @Success 201 {object} dto.AudioCollectionResponseDTO "Collection created successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Track ID Format / Collection Type"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections [post]
func (h *AudioHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCollectionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	initialTrackIDs := make([]domain.TrackID, 0, len(req.InitialTrackIDs))
	for _, idStr := range req.InitialTrackIDs {
		id, err := domain.TrackIDFromString(idStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid initial track ID format '%s'", domain.ErrInvalidArgument, idStr))
			return
		}
		initialTrackIDs = append(initialTrackIDs, id)
	}

	collectionType := domain.CollectionType(req.Type)

	collection, err := h.audioUseCase.CreateCollection(r.Context(), req.Title, req.Description, collectionType, initialTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	resp := dto.MapDomainCollectionToResponseDTO(collection, nil)
	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}

// GetCollectionDetails handles GET /api/v1/audio/collections/{collectionId}
// @Summary Get audio collection details
// @Description Retrieves details for a specific audio collection, including its metadata and ordered list of tracks.
// @ID get-collection-details
// @Tags Audio Collections
// @Produce json
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Success 200 {object} dto.AudioCollectionResponseDTO "Audio collection details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Collection ID Format"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error (e.g., failed to fetch tracks)"
// @Router /audio/collections/{collectionId} [get]
func (h *AudioHandler) GetCollectionDetails(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}

	collection, err := h.audioUseCase.GetCollectionDetails(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Fetch tracks separately (Usecase GetCollectionDetails returns the collection metadata + IDs)
	// If the usecase returned full tracks, this call wouldn't be needed.
	tracks, err := h.audioUseCase.GetCollectionTracks(r.Context(), collectionID)
	if err != nil {
		slog.Default().ErrorContext(r.Context(), "Failed to fetch tracks for collection details", "error", err, "collectionID", collectionID)
		httputil.RespondError(w, r, fmt.Errorf("failed to retrieve collection tracks: %w", err))
		return
	}

	resp := dto.MapDomainCollectionToResponseDTO(collection, tracks)
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// UpdateCollectionMetadata handles PUT /api/v1/audio/collections/{collectionId}
// @Summary Update collection metadata
// @Description Updates the title and description of an audio collection owned by the authenticated user.
// @ID update-collection-metadata
// @Tags Audio Collections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Param collection body dto.UpdateCollectionRequestDTO true "Updated collection metadata"
// @Success 204 "Collection metadata updated successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Collection ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections/{collectionId} [put]
func (h *AudioHandler) UpdateCollectionMetadata(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}
	var req dto.UpdateCollectionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}
	err = h.audioUseCase.UpdateCollectionMetadata(r.Context(), collectionID, req.Title, req.Description)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateCollectionTracks handles PUT /api/v1/audio/collections/{collectionId}/tracks
// @Summary Update collection tracks
// @Description Updates the ordered list of tracks within a specific collection owned by the authenticated user. Replaces the entire list.
// @ID update-collection-tracks
// @Tags Audio Collections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Param tracks body dto.UpdateCollectionTracksRequestDTO true "Ordered list of track UUIDs"
// @Success 204 "Collection tracks updated successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Collection or Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found / Track Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections/{collectionId}/tracks [put]
func (h *AudioHandler) UpdateCollectionTracks(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}
	var req dto.UpdateCollectionTracksRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}
	orderedTrackIDs := make([]domain.TrackID, 0, len(req.OrderedTrackIDs))
	for _, idStr := range req.OrderedTrackIDs {
		id, err := domain.TrackIDFromString(idStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format '%s' in ordered list", domain.ErrInvalidArgument, idStr))
			return
		}
		orderedTrackIDs = append(orderedTrackIDs, id)
	}
	err = h.audioUseCase.UpdateCollectionTracks(r.Context(), collectionID, orderedTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteCollection handles DELETE /api/v1/audio/collections/{collectionId}
// @Summary Delete an audio collection
// @Description Deletes an audio collection owned by the authenticated user.
// @ID delete-audio-collection
// @Tags Audio Collections
// @Produce json
// @Security BearerAuth
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Success 204 "Collection deleted successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections/{collectionId} [delete]
func (h *AudioHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}
	err = h.audioUseCase.DeleteCollection(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

## `internal/adapter/handler/http/user_handler.go`

```go
// internal/adapter/handler/http/user_handler.go
package http

import (
	// REMOVED: "context" - Not needed directly here
	"net/http"

	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto" // Import dto package
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port" // Import port package for UserUseCase interface
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
)

// UserHandler handles HTTP requests related to user profiles.
type UserHandler struct {
	userUseCase port.UserUseCase // Use interface from port package
	// validator *validation.Validator // Add validator if needed for PUT/PATCH later
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(uc port.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: uc,
	}
}

// GetMyProfile handles GET /api/v1/users/me
// @Summary Get current user's profile
// @Description Retrieves the profile information for the currently authenticated user.
// @ID get-my-profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponseDTO "User profile retrieved successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 404 {object} httputil.ErrorResponseDTO "User Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me [get]
func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	user, err := h.userUseCase.GetUserProfile(r.Context(), userID)
	if err != nil {
		// Handles domain.ErrNotFound, etc.
		httputil.RespondError(w, r, err)
		return
	}

	resp := dto.MapDomainUserToResponseDTO(user) // Use mapping function from DTO package
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}
```

## `internal/adapter/handler/http/upload_handler.go`

```go
// internal/adapter/handler/http/upload_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	// Assuming module path is updated
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port" // Import port package
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// UploadHandler handles HTTP requests related to file uploads.
type UploadHandler struct {
	uploadUseCase port.UploadUseCase // Use interface from port package
	validator     *validation.Validator
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uc port.UploadUseCase, v *validation.Validator) *UploadHandler {
	return &UploadHandler{
		uploadUseCase: uc,
		validator:     v,
	}
}

// RequestUpload handles POST /api/v1/uploads/audio/request
// @Summary Request presigned URL for audio upload
// @Description Requests a presigned URL from the object storage (MinIO/S3) that can be used by the client to directly upload an audio file.
// @ID request-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uploadRequest body dto.RequestUploadRequestDTO true "Upload Request Info (filename, content type)"
// @Success 200 {object} dto.RequestUploadResponseDTO "Presigned URL and object key generated"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error (e.g., failed to generate URL)"
// @Router /uploads/audio/request [post]
func (h *UploadHandler) RequestUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.RequestUploadRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Call use case - returns port.RequestUploadResult
	result, err := h.uploadUseCase.RequestUpload(r.Context(), userID, req.Filename, req.ContentType)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map port result to response DTO
	resp := dto.RequestUploadResponseDTO{
		UploadURL: result.UploadURL,
		ObjectKey: result.ObjectKey,
	}
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// CompleteUploadAndCreateTrack handles POST /api/v1/audio/tracks
// @Summary Complete audio upload and create track metadata (Single File)
// @Description After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database. Use `/audio/tracks/batch/complete` for batch uploads.
// @ID complete-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param completeUpload body dto.CompleteUploadInputDTO true "Track metadata and object key"
// @Success 201 {object} dto.AudioTrackResponseDTO "Track metadata created successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., validation errors, object key not found, file not in storage)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Object key mismatch)"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict (e.g., object key already used in DB)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [post]
func (h *UploadHandler) CompleteUploadAndCreateTrack(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.CompleteUploadInputDTO // DTO from adapter layer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO to port.CompleteUploadInput
	portReq := port.CompleteUploadInput{
		ObjectKey:     req.ObjectKey,
		Title:         req.Title,
		Description:   req.Description,
		LanguageCode:  req.LanguageCode,
		Level:         req.Level,
		DurationMs:    req.DurationMs,
		IsPublic:      req.IsPublic,
		Tags:          req.Tags,
		CoverImageURL: req.CoverImageURL,
	}

	// Call use case with port params
	track, err := h.uploadUseCase.CompleteUpload(r.Context(), userID, portReq)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map domain.AudioTrack to response DTO
	resp := dto.MapDomainTrackToResponseDTO(track)
	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}

// --- Batch Upload Handlers ---

// RequestBatchUpload handles POST /api/v1/uploads/audio/batch/request
// @Summary Request presigned URLs for batch audio upload
// @Description Requests multiple presigned URLs for uploading several audio files in parallel.
// @ID request-batch-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param batchUploadRequest body dto.BatchRequestUploadInputRequestDTO true "List of files to request URLs for"
// @Success 200 {object} dto.BatchRequestUploadInputResponseDTO "List of generated presigned URLs and object keys, including potential errors per item."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., empty file list)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /uploads/audio/batch/request [post]
func (h *UploadHandler) RequestBatchUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.BatchRequestUploadInputRequestDTO // Use batch DTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO request to port request
	portReq := port.BatchRequestUploadInput{Files: make([]port.BatchRequestUploadInputItem, len(req.Files))}
	for i, f := range req.Files {
		portReq.Files[i] = port.BatchRequestUploadInputItem{
			Filename:    f.Filename,
			ContentType: f.ContentType,
		}
	}

	// Call use case - returns []port.BatchURLResultItem
	results, err := h.uploadUseCase.RequestBatchUpload(r.Context(), userID, portReq)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map port results to response DTO results
	respItems := make([]dto.BatchRequestUploadInputResponseItemDTO, len(results))
	for i, res := range results {
		respItems[i] = dto.BatchRequestUploadInputResponseItemDTO{
			OriginalFilename: res.OriginalFilename,
			ObjectKey:        res.ObjectKey,
			UploadURL:        res.UploadURL,
			Error:            res.Error,
		}
	}
	resp := dto.BatchRequestUploadInputResponseDTO{Results: respItems}
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// CompleteBatchUploadAndCreateTracks handles POST /api/v1/audio/tracks/batch/complete
// @Summary Complete batch audio upload and create track metadata
// @Description After clients successfully upload multiple files using presigned URLs, this endpoint is called to create the corresponding audio track metadata records in the database within a single transaction.
// @ID complete-batch-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param batchCompleteUpload body dto.BatchCompleteUploadInputDTO true "List of track metadata and object keys for uploaded files"
// @Success 201 {object} dto.BatchCompleteUploadResponseDTO "Batch processing attempted. Results indicate success/failure per item. If overall transaction succeeded, status is 201."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., validation errors in items, files not in storage)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Object key mismatch)"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict (e.g., duplicate object key during processing)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error (e.g., transaction failure)"
// @Router /audio/tracks/batch/complete [post]
func (h *UploadHandler) CompleteBatchUploadAndCreateTracks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.BatchCompleteUploadInputDTO // Use batch DTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO request to port request
	portReq := port.BatchCompleteInput{Tracks: make([]port.BatchCompleteItem, len(req.Tracks))}
	for i, t := range req.Tracks {
		portReq.Tracks[i] = port.BatchCompleteItem{
			ObjectKey:     t.ObjectKey,
			Title:         t.Title,
			Description:   t.Description,
			LanguageCode:  t.LanguageCode,
			Level:         t.Level,
			DurationMs:    t.DurationMs,
			IsPublic:      t.IsPublic,
			Tags:          t.Tags,
			CoverImageURL: t.CoverImageURL,
		}
	}

	// Call use case - returns []port.BatchCompleteResultItem
	results, err := h.uploadUseCase.CompleteBatchUpload(r.Context(), userID, portReq)
	// `err` represents an overall transactional or fundamental failure

	// Map port results to response DTO results
	respItems := make([]dto.BatchCompleteUploadResponseItemDTO, len(results))
	for i, res := range results {
		respItems[i] = dto.BatchCompleteUploadResponseItemDTO{
			ObjectKey: res.ObjectKey,
			Success:   res.Success,
			TrackID:   res.TrackID,
			Error:     res.Error,
		}
	}
	resp := dto.BatchCompleteUploadResponseDTO{Results: respItems}

	if err != nil {
		// Overall transaction or critical pre-check failed
		httputil.RespondError(w, r, err) // Map the overall error
		return
	}

	// Pre-checks passed and transaction committed (even if some DB inserts failed within the rolled-back tx)
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // Use 201 Created
}
```

## `internal/adapter/handler/http/dto/activity_dto.go`

```go
// internal/adapter/handler/http/dto/activity_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
)

// --- Request DTOs ---

// RecordProgressRequestDTO defines the JSON body for recording playback progress.
type RecordProgressRequestDTO struct {
	TrackID    string `json:"trackId" validate:"required,uuid"`
	ProgressMs int64  `json:"progressMs" validate:"required,gte=0"` // Point 1: Already uses ms
}

// CreateBookmarkRequestDTO defines the JSON body for creating a bookmark.
type CreateBookmarkRequestDTO struct {
	TrackID     string `json:"trackId" validate:"required,uuid"`
	TimestampMs int64  `json:"timestampMs" validate:"required,gte=0"` // Point 1: Already uses ms
	Note        string `json:"note"`
}

// --- Response DTOs ---

// PlaybackProgressResponseDTO defines the JSON representation of playback progress.
type PlaybackProgressResponseDTO struct {
	UserID         string    `json:"userId"`
	TrackID        string    `json:"trackId"`
	ProgressMs     int64     `json:"progressMs"` // Point 1: Already uses ms
	LastListenedAt time.Time `json:"lastListenedAt"`
}

// Point 1: MapDomainProgressToResponseDTO converts domain progress (with time.Duration) to DTO (with int64 ms).
func MapDomainProgressToResponseDTO(p *domain.PlaybackProgress) PlaybackProgressResponseDTO {
	if p == nil {
		return PlaybackProgressResponseDTO{}
	} // Handle nil gracefully
	return PlaybackProgressResponseDTO{
		UserID:         p.UserID.String(),
		TrackID:        p.TrackID.String(),
		ProgressMs:     p.Progress.Milliseconds(), // Convert duration to ms
		LastListenedAt: p.LastListenedAt,
	}
}

// BookmarkResponseDTO defines the JSON representation of a bookmark.
// THIS IS NOW THE SINGLE SOURCE OF TRUTH FOR BOOKMARK RESPONSE DTO.
type BookmarkResponseDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	TrackID     string    `json:"trackId"`
	TimestampMs int64     `json:"timestampMs"`
	Note        string    `json:"note,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// MapDomainBookmarkToResponseDTO converts domain bookmark (with time.Duration) to DTO (with int64 ms).
func MapDomainBookmarkToResponseDTO(b *domain.Bookmark) BookmarkResponseDTO {
	if b == nil {
		return BookmarkResponseDTO{}
	} // Handle nil gracefully
	return BookmarkResponseDTO{
		ID:          b.ID.String(),
		UserID:      b.UserID.String(), // Keep UserID here for now
		TrackID:     b.TrackID.String(),
		TimestampMs: b.Timestamp.Milliseconds(), // Convert duration to ms
		Note:        b.Note,
		CreatedAt:   b.CreatedAt,
	}
}
```

## `internal/adapter/handler/http/dto/common_dto.go`

```go
// internal/adapter/handler/http/dto/common_dto.go
package dto

// PaginatedResponseDTO defines the standard structure for paginated API responses.
type PaginatedResponseDTO struct {
	Data       interface{} `json:"data"`       // The slice of items for the current page (e.g., []AudioTrackResponseDTO)
	Total      int         `json:"total"`      // Total number of items matching the query
	Limit      int         `json:"limit"`      // The limit used for this page
	Offset     int         `json:"offset"`     // The offset used for this page
	Page       int         `json:"page"`       // Current page number (1-based)
	TotalPages int         `json:"totalPages"` // Total number of pages
}
```

## `internal/adapter/handler/http/dto/user_dto.go`

```go
// internal/adapter/handler/http/dto/user_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

// UserResponseDTO defines the JSON representation of a user profile.
type UserResponseDTO struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	Name            string  `json:"name"`
	AuthProvider    string  `json:"authProvider"`
	ProfileImageURL *string `json:"profileImageUrl,omitempty"`
	CreatedAt       string  `json:"createdAt"` // Use string format like RFC3339
	UpdatedAt       string  `json:"updatedAt"`
}

// MapDomainUserToResponseDTO converts a domain user to its DTO representation.
func MapDomainUserToResponseDTO(user *domain.User) UserResponseDTO {
	return UserResponseDTO{
		ID:              user.ID.String(),
		Email:           user.Email.String(),
		Name:            user.Name,
		AuthProvider:    string(user.AuthProvider),
		ProfileImageURL: user.ProfileImageURL,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339), // Format time
		UpdatedAt:       user.UpdatedAt.Format(time.RFC3339),
	}
}

// Note: Auth Request/Response DTOs remain in auth_dto.go
```

## `internal/adapter/handler/http/dto/audio_dto.go`

```go
// internal/adapter/handler/http/dto/audio_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port" // Import port for result struct
)

// --- Request DTOs ---

// ListTracksRequestDTO (Documents query parameters, not used for binding)
type ListTracksRequestDTO struct {
	Query         *string  `query:"q"`
	LanguageCode  *string  `query:"lang"`
	Level         *string  `query:"level"`
	IsPublic      *bool    `query:"isPublic"`
	Tags          []string `query:"tags"`
	SortBy        string   `query:"sortBy"`
	SortDirection string   `query:"sortDir"`
	Limit         int      `query:"limit"`
	Offset        int      `query:"offset"`
}

// CreateCollectionRequestDTO defines the JSON body for creating a collection.
type CreateCollectionRequestDTO struct {
	Title           string   `json:"title" validate:"required,max=255"`
	Description     string   `json:"description"`
	Type            string   `json:"type" validate:"required,oneof=COURSE PLAYLIST"`
	InitialTrackIDs []string `json:"initialTrackIds" validate:"omitempty,dive,uuid"`
}

// UpdateCollectionRequestDTO defines the JSON body for updating collection metadata.
type UpdateCollectionRequestDTO struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// UpdateCollectionTracksRequestDTO defines the JSON body for updating tracks in a collection.
type UpdateCollectionTracksRequestDTO struct {
	OrderedTrackIDs []string `json:"orderedTrackIds" validate:"omitempty,dive,uuid"`
}

// --- Response DTOs ---

// AudioTrackResponseDTO defines the JSON representation of a single audio track's basic info.
type AudioTrackResponseDTO struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	LanguageCode  string    `json:"languageCode"`
	Level         string    `json:"level,omitempty"`
	DurationMs    int64     `json:"durationMs"` // Point 1: Use milliseconds (int64)
	CoverImageURL *string   `json:"coverImageUrl,omitempty"`
	UploaderID    *string   `json:"uploaderId,omitempty"`
	IsPublic      bool      `json:"isPublic"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AudioTrackDetailsResponseDTO includes the track metadata, playback URL, and user-specific info.
type AudioTrackDetailsResponseDTO struct {
	AudioTrackResponseDTO                       // Embed basic track info
	PlayURL               string                `json:"playUrl"`                  // Presigned URL
	UserProgressMs        *int64                `json:"userProgressMs,omitempty"` // Point 1: User progress in ms
	UserBookmarks         []BookmarkResponseDTO `json:"userBookmarks,omitempty"`  // Array of user bookmarks for this track
}

// UploaderInfoDTO - embedded within AudioTrackDetailsResponseDTO if needed
type UploaderInfoDTO struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// MapDomainTrackToResponseDTO converts a domain track to its basic response DTO.
func MapDomainTrackToResponseDTO(track *domain.AudioTrack) AudioTrackResponseDTO {
	if track == nil {
		return AudioTrackResponseDTO{}
	} // Handle nil track gracefully

	var uploaderIDStr *string
	if track.UploaderID != nil {
		s := track.UploaderID.String()
		uploaderIDStr = &s
	}
	return AudioTrackResponseDTO{
		ID:            track.ID.String(),
		Title:         track.Title,
		Description:   track.Description,
		LanguageCode:  track.Language.Code(),
		Level:         string(track.Level),
		DurationMs:    track.Duration.Milliseconds(), // Convert Duration to ms
		CoverImageURL: track.CoverImageURL,
		UploaderID:    uploaderIDStr,
		IsPublic:      track.IsPublic,
		Tags:          track.Tags,
		CreatedAt:     track.CreatedAt,
		UpdatedAt:     track.UpdatedAt,
	}
}

// MapDomainTrackToDetailsResponseDTO converts the result from the usecase to the detailed response DTO.
func MapDomainTrackToDetailsResponseDTO(result *port.GetAudioTrackDetailsResult) AudioTrackDetailsResponseDTO {
	if result == nil || result.Track == nil {
		return AudioTrackDetailsResponseDTO{} // Return empty DTO if track is nil
	}

	baseDTO := MapDomainTrackToResponseDTO(result.Track)
	detailsDTO := AudioTrackDetailsResponseDTO{
		AudioTrackResponseDTO: baseDTO,
		PlayURL:               result.PlayURL,
		UserProgressMs:        nil,
		UserBookmarks:         make([]BookmarkResponseDTO, 0), // Initialize with correct type
	}

	if result.UserProgress != nil {
		progressMs := result.UserProgress.Progress.Milliseconds()
		detailsDTO.UserProgressMs = &progressMs
	}

	if len(result.UserBookmarks) > 0 {
		// UPDATED: Ensure we create BookmarkResponseDTO instances using the function from activity_dto.go scope
		detailsDTO.UserBookmarks = make([]BookmarkResponseDTO, len(result.UserBookmarks))
		for i, b := range result.UserBookmarks {
			// Call the mapping function (which is accessible within the dto package)
			detailsDTO.UserBookmarks[i] = MapDomainBookmarkToResponseDTO(b)
		}
	}

	return detailsDTO
}

// AudioCollectionResponseDTO defines the JSON representation of a collection.
type AudioCollectionResponseDTO struct {
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description,omitempty"`
	OwnerID     string                  `json:"ownerId"`
	Type        string                  `json:"type"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	Tracks      []AudioTrackResponseDTO `json:"tracks,omitempty"`
}

// MapDomainCollectionToResponseDTO converts a domain collection to its response DTO.
func MapDomainCollectionToResponseDTO(collection *domain.AudioCollection, tracks []*domain.AudioTrack) AudioCollectionResponseDTO {
	if collection == nil {
		return AudioCollectionResponseDTO{}
	} // Handle nil gracefully

	dto := AudioCollectionResponseDTO{
		ID:          collection.ID.String(),
		Title:       collection.Title,
		Description: collection.Description,
		OwnerID:     collection.OwnerID.String(),
		Type:        string(collection.Type),
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
		Tracks:      make([]AudioTrackResponseDTO, 0), // Initialize empty
	}
	if tracks != nil {
		dto.Tracks = make([]AudioTrackResponseDTO, len(tracks))
		for i, t := range tracks {
			dto.Tracks[i] = MapDomainTrackToResponseDTO(t) // Use the basic track mapper
		}
	}
	return dto
}

// PaginatedTracksResponseDTO - Using common PaginatedResponseDTO instead
// type PaginatedTracksResponseDTO struct { ... }

// PaginatedCollectionsResponseDTO - Using common PaginatedResponseDTO instead
// type PaginatedCollectionsResponseDTO struct { ... }
```

## `internal/adapter/handler/http/dto/auth_dto.go`

```go
// internal/adapter/handler/http/dto/auth_dto.go
package dto

// --- Request DTOs ---

// RegisterRequestDTO defines the expected JSON body for user registration.
type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`                    // Add example tag
	Password string `json:"password" validate:"required,min=8" format:"password" example:"Str0ngP@ssw0rd"` // Add format tag
	Name     string `json:"name" validate:"required,max=100" example:"John Doe"`
}

// LoginRequestDTO defines the expected JSON body for user login.
type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// GoogleCallbackRequestDTO defines the expected JSON body for Google OAuth callback.
type GoogleCallbackRequestDTO struct {
	IDToken string `json:"idToken" validate:"required"`
}

// RefreshRequestDTO defines the expected JSON body for token refresh.
// ADDED
type RefreshRequestDTO struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// LogoutRequestDTO defines the expected JSON body for logout.
// ADDED (Optional, can also just take token from body or header)
type LogoutRequestDTO struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// --- Response DTOs ---

// AuthResponseDTO defines the JSON response body for successful authentication/refresh.
// MODIFIED: Added refreshToken
type AuthResponseDTO struct {
	AccessToken  string `json:"accessToken"`         // The JWT access token
	RefreshToken string `json:"refreshToken"`        // The refresh token value
	IsNewUser    *bool  `json:"isNewUser,omitempty"` // Pointer, only included for Google callback if user is new
}
```

## `internal/adapter/handler/http/dto/upload_dto.go`

```go
// internal/adapter/handler/http/dto/upload_dto.go
package dto

// === Single Upload DTOs ===

// RequestUploadRequestDTO defines the JSON body for requesting an upload URL.
type RequestUploadRequestDTO struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"contentType" validate:"required"` // e.g., "audio/mpeg"
}

// RequestUploadResponseDTO defines the JSON response after requesting an upload URL.
type RequestUploadResponseDTO struct {
	UploadURL string `json:"uploadUrl"` // The presigned PUT URL
	ObjectKey string `json:"objectKey"` // The key the client should use/report back
}

// CompleteUploadInputDTO defines the JSON body for finalizing an upload
// and creating the audio track metadata record.
type CompleteUploadInputDTO struct {
	ObjectKey     string   `json:"objectKey" validate:"required"`
	Title         string   `json:"title" validate:"required,max=255"`
	Description   string   `json:"description"`
	LanguageCode  string   `json:"languageCode" validate:"required"`
	Level         string   `json:"level" validate:"omitempty,oneof=A1 A2 B1 B2 C1 C2 NATIVE"` // Allow empty or valid level
	DurationMs    int64    `json:"durationMs" validate:"required,gt=0"`                       // Duration in Milliseconds, must be positive
	IsPublic      bool     `json:"isPublic"`                                                  // Defaults to false if omitted? Define behavior.
	Tags          []string `json:"tags"`
	CoverImageURL *string  `json:"coverImageUrl" validate:"omitempty,url"`
}

// === Batch Upload DTOs ===

// BatchRequestUploadInputItemDTO represents a single file in the batch request for URLs.
type BatchRequestUploadInputItemDTO struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"contentType" validate:"required"` // e.g., "audio/mpeg"
}

// BatchRequestUploadInputRequestDTO is the request body for requesting multiple upload URLs.
type BatchRequestUploadInputRequestDTO struct {
	Files []BatchRequestUploadInputItemDTO `json:"files" validate:"required,min=1,dive"` // Ensure at least one file, validate each item
}

// BatchRequestUploadInputResponseItemDTO represents the response for a single file URL request.
type BatchRequestUploadInputResponseItemDTO struct {
	OriginalFilename string `json:"originalFilename"` // Helps client match response to request
	ObjectKey        string `json:"objectKey"`        // The generated object key for this file
	UploadURL        string `json:"uploadUrl"`        // The presigned PUT URL for this file
	Error            string `json:"error,omitempty"`  // Error message if URL generation failed for this item
}

// BatchRequestUploadInputResponseDTO is the response body containing results for multiple URL requests.
type BatchRequestUploadInputResponseDTO struct {
	Results []BatchRequestUploadInputResponseItemDTO `json:"results"`
}

// BatchCompleteUploadItemDTO represents metadata for one successfully uploaded file in the batch completion request.
type BatchCompleteUploadItemDTO struct {
	ObjectKey     string   `json:"objectKey" validate:"required"`
	Title         string   `json:"title" validate:"required,max=255"`
	Description   string   `json:"description"`
	LanguageCode  string   `json:"languageCode" validate:"required"`
	Level         string   `json:"level" validate:"omitempty,oneof=A1 A2 B1 B2 C1 C2 NATIVE"`
	DurationMs    int64    `json:"durationMs" validate:"required,gt=0"`
	IsPublic      bool     `json:"isPublic"`
	Tags          []string `json:"tags"`
	CoverImageURL *string  `json:"coverImageUrl" validate:"omitempty,url"`
}

// BatchCompleteUploadInputDTO is the request body for finalizing multiple uploads.
type BatchCompleteUploadInputDTO struct {
	Tracks []BatchCompleteUploadItemDTO `json:"tracks" validate:"required,min=1,dive"` // Ensure at least one track, validate each item
}

// BatchCompleteUploadResponseItemDTO represents the processing result for a single item in the batch completion.
type BatchCompleteUploadResponseItemDTO struct {
	ObjectKey string `json:"objectKey"`         // Identifies the item
	Success   bool   `json:"success"`           // Whether processing this item succeeded
	TrackID   string `json:"trackId,omitempty"` // The ID of the created track if successful
	Error     string `json:"error,omitempty"`   // Error message if processing failed for this item
}

// BatchCompleteUploadResponseDTO is the overall response for the batch completion request.
// Note: This will likely be a 2xx status even if some items failed validation before DB commit,
// or a 4xx/5xx if the overall transaction failed. The details are in the items.
// Consider using 207 Multi-Status if you want to represent partial success explicitly at the HTTP level,
// but for atomicity, usually it's all-or-nothing for the DB part.
type BatchCompleteUploadResponseDTO struct {
	Results []BatchCompleteUploadResponseItemDTO `json:"results"`
}
```

## `internal/adapter/handler/http/middleware/recovery.go`

```go
// internal/adapter/handler/http/middleware/recovery.go
package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
)

// Recoverer is a middleware that recovers from panics, logs the panic,
// and returns a 500 Internal Server Error.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Log the panic
				logger := slog.Default() // Get default logger (set in main)
				// Get request ID if available (assuming RequestID middleware runs before)
				reqID := GetReqID(r.Context())
				logger.ErrorContext(r.Context(), "Panic recovered",
					"error", rvr,
					"request_id", reqID,
					"stack", string(debug.Stack()),
				)

				// Use httputil.RespondError for consistent JSON error response
				err := errors.New("internal server error: recovered from panic")
				httputil.RespondError(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
```

## `internal/adapter/handler/http/middleware/auth.go`

```go
// internal/adapter/handler/http/middleware/auth.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/internal/port"   // Adjust import path (for SecurityHelper interface)
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"    // Adjust import path
)

const UserIDKey httputil.ContextKey = "userID" // Use httputil.ContextKey type

// Authenticator creates a middleware that verifies the JWT token.
func Authenticator(secHelper port.SecurityHelper) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				err := fmt.Errorf("%w: Authorization header missing", domain.ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				err := fmt.Errorf("%w: Authorization header format must be Bearer {token}", domain.ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			tokenString := headerParts[1]
			if tokenString == "" {
				err := fmt.Errorf("%w: Authorization token missing", domain.ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			// Verify the token using the SecurityHelper
			userID, err := secHelper.VerifyJWT(r.Context(), tokenString)
			if err != nil {
				// VerifyJWT should return domain errors (ErrAuthenticationFailed, ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)

			// Token is valid, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext retrieves the UserID from the context.
// Returns domain.UserID zero value and false if not found or type is wrong.
func GetUserIDFromContext(ctx context.Context) (domain.UserID, bool) {
	userID, ok := ctx.Value(UserIDKey).(domain.UserID)
	return userID, ok
}
```

## `internal/adapter/handler/http/middleware/logger.go`

```go
// internal/adapter/handler/http/middleware/logger.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// ResponseWriterWrapper wraps http.ResponseWriter to capture status code.
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	// Default to 200 OK if WriteHeader is not called
	return &ResponseWriterWrapper{w, http.StatusOK}
}

func (rw *ResponseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs incoming requests and their processing time.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := slog.Default().With(
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"request_id", GetReqID(r.Context()), // Get request ID from context
		)

		logger.InfoContext(r.Context(), "Request started")

		// Wrap response writer to capture status code
		wrappedWriter := NewResponseWriterWrapper(w)

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		logger.InfoContext(r.Context(), "Request finished",
			"status_code", wrappedWriter.statusCode,
			"duration_ms", duration.Milliseconds(),
		)
	})
}
```

## `internal/adapter/handler/http/middleware/api_security_headers.go`

```go
// internal/adapter/handler/http/middleware/api_security_headers.go
package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// ApiSecurityHeaders adds common security-related HTTP headers tailored for API responses.
// It uses a strict Content-Security-Policy.
func ApiSecurityHeaders(next http.Handler) http.Handler {
	// Use a shorter HSTS max-age during testing/initial deployment (e.g., 5 minutes = 300)
	// Remove 'preload' until ready to submit to HSTS preload list.
	hstsMaxAgeSeconds := int(5 * time.Minute / time.Second) // 300 seconds
	hstsHeader := "max-age=" + strconv.Itoa(hstsMaxAgeSeconds) + "; includeSubDomains"

	// Strict Content-Security-Policy for APIs (prevents loading external resources, inline scripts/styles)
	// Adjust 'connect-src' if your API needs to make requests to other origins from client-side JS (unlikely for pure API)
	cspHeader := "default-src 'self'; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()

		// HTTP Strict Transport Security (HSTS)
		headers.Set("Strict-Transport-Security", hstsHeader)

		// X-Content-Type-Options
		headers.Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options (Good practice even for APIs, belt-and-suspenders)
		headers.Set("X-Frame-Options", "DENY")

		// Content-Security-Policy (CSP) - Strict version
		headers.Set("Content-Security-Policy", cspHeader)

		// Referrer-Policy
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (Example: disable features not needed by API)
		headers.Set("Permissions-Policy", "microphone=(), geolocation=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
```

## `internal/adapter/handler/http/middleware/ratelimit.go`

```go
// internal/adapter/handler/http/middleware/ratelimit.go
package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"golang.org/x/time/rate"
)

// limiterEntry stores the limiter and the last seen time.
// Point 7: Added struct to track lastSeen
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter stores rate limiters for IP addresses.
type IPRateLimiter struct {
	ips map[string]*limiterEntry // Point 7: Changed map value type
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new IPRateLimiter.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips: make(map[string]*limiterEntry),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
	// Point 7: Start cleanup goroutine
	go limiter.CleanUpOldLimiters(10*time.Minute, 30*time.Minute) // Example: cleanup every 10min, remove if idle > 30min

	return limiter
}

// Point 7: Modified AddIP to use limiterEntry
func (i *IPRateLimiter) addIP(ip string) *limiterEntry {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Double check existence inside lock
	entry, exists := i.ips[ip]
	if !exists {
		entry = &limiterEntry{
			limiter:  rate.NewLimiter(i.r, i.b),
			lastSeen: time.Now(),
		}
		i.ips[ip] = entry
		slog.Debug("Added new rate limiter entry", "ip_address", ip) // Added debug log
	}
	return entry
}

// Point 7: Modified GetLimiter to return limiter and update lastSeen
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	entry, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		entry = i.addIP(ip) // addIP handles locking
	} else {
		// Update lastSeen without full lock if possible, but requires locking for write.
		// Simple approach: Lock for update.
		i.mu.Lock()
		entry.lastSeen = time.Now()
		i.mu.Unlock()
	}

	return entry.limiter
}

// RateLimit is the middleware handler.
func RateLimit(limiter *IPRateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get IP logic remains the same
			ip := r.RemoteAddr
			if realIP := r.Context().Value(http.CanonicalHeaderKey("X-Real-IP")); realIP != nil {
				if ripStr, ok := realIP.(string); ok {
					ip = ripStr
				}
			} else if forwardedFor := r.Context().Value(http.CanonicalHeaderKey("X-Forwarded-For")); forwardedFor != nil {
				if fwdStr, ok := forwardedFor.(string); ok {
					parts := strings.Split(fwdStr, ",")
					if len(parts) > 0 {
						ip = strings.TrimSpace(parts[0])
					}
				}
			} else {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					ip = host
				}
			}

			// Get the rate limiter for the current IP address.
			l := limiter.GetLimiter(ip) // This now also updates lastSeen

			if !l.Allow() {
				logger := slog.Default()
				logger.WarnContext(r.Context(), "Rate limit exceeded", "ip_address", ip, "request_id", httputil.GetReqID(r.Context()))

				// Use specific rate limit error message
				err := fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied)
				httputil.RespondError(w, r, err) // httputil maps this to 429
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Point 7: Implemented CleanUpOldLimiters logic
func (i *IPRateLimiter) CleanUpOldLimiters(interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		count := 0
		now := time.Now()
		for ip, entry := range i.ips {
			if now.Sub(entry.lastSeen) > maxAge {
				delete(i.ips, ip)
				count++
			}
		}
		if count > 0 {
			slog.Debug("Rate limiter cleanup removed entries", "removed_count", count, "current_size", len(i.ips))
		}
		i.mu.Unlock()
	}
}
```

## `internal/adapter/handler/http/middleware/swagger_security_headers.go`

```go
// internal/adapter/handler/http/middleware/swagger_security_headers.go
package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// SwaggerSecurityHeaders adds security headers tailored for Swagger UI compatibility.
// It uses a relaxed Content-Security-Policy allowing 'unsafe-inline'.
func SwaggerSecurityHeaders(next http.Handler) http.Handler {
	// HSTS: Optional for Swagger docs, especially in dev. If used, keep short and no 'preload'.
	hstsMaxAgeSeconds := int(5 * time.Minute / time.Second) // 300 seconds
	hstsHeader := "max-age=" + strconv.Itoa(hstsMaxAgeSeconds) + "; includeSubDomains"

	// Relaxed Content-Security-Policy for Swagger UI
	// Allows inline styles and scripts, and data URIs for images (for icons).
	cspHeader := "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()

		// HTTP Strict Transport Security (HSTS) - Optional for Swagger
		headers.Set("Strict-Transport-Security", hstsHeader)

		// X-Content-Type-Options
		headers.Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options (Swagger UI might need SAMEORIGIN if embedded, DENY is safer)
		headers.Set("X-Frame-Options", "DENY")

		// Content-Security-Policy (CSP) - Relaxed version
		headers.Set("Content-Security-Policy", cspHeader)

		// Referrer-Policy
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (Generally not critical for Swagger docs)
		// headers.Set("Permissions-Policy", "...")

		next.ServeHTTP(w, r)
	})
}
```

## `internal/adapter/handler/http/middleware/request_id.go`

```go
// internal/adapter/handler/http/middleware/request_id.go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid" // Using google/uuid for request IDs
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
)

// RequestID is a middleware that injects a unique request ID into the context
// and sets it in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get existing ID from header or generate a new one
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// Set ID in response header
		w.Header().Set("X-Request-ID", reqID)

		// Add ID to request context
		ctx := context.WithValue(r.Context(), httputil.RequestIDKey, reqID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetReqID retrieves the request ID from the context.
// Returns an empty string if not found.
func GetReqID(ctx context.Context) string {
	return httputil.GetReqID(ctx)
}
```

## `internal/adapter/repository/postgres/audiotrack_repo.go`

```go
// internal/adapter/repository/postgres/audiotrack_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq" // Using lib/pq for array handling with pgx and error codes

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

type AudioTrackRepository struct {
	db         *pgxpool.Pool
	logger     *slog.Logger
	getQuerier func(ctx context.Context) Querier
}

func NewAudioTrackRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioTrackRepository {
	repo := &AudioTrackRepository{
		db:     db,
		logger: logger.With("repository", "AudioTrackRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db)
	}
	return repo
}

// --- Interface Implementation ---

func (r *AudioTrackRepository) Create(ctx context.Context, track *domain.AudioTrack) error {
	q := r.getQuerier(ctx)
	query := `
		INSERT INTO audio_tracks
			(id, title, description, language_code, level, duration_ms,
			 minio_bucket, minio_object_key, cover_image_url, uploader_id,
			 is_public, tags, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := q.Exec(ctx, query,
		track.ID,
		track.Title,
		track.Description,
		track.Language.Code(),
		track.Level,
		track.Duration.Milliseconds(), // Point 1: Convert domain Duration to int64 ms
		track.MinioBucket,
		track.MinioObjectKey,
		track.CoverImageURL,
		track.UploaderID,
		track.IsPublic,
		pq.Array(track.Tags),
		track.CreatedAt,
		track.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "audio_tracks_minio_object_key_key") {
				return fmt.Errorf("creating audio track: %w: object key '%s' already exists", domain.ErrConflict, track.MinioObjectKey)
			}
			r.logger.WarnContext(ctx, "Unique constraint violation on audio track creation", "constraint", pgErr.ConstraintName, "trackID", track.ID)
			return fmt.Errorf("creating audio track: %w: resource conflict on unique field", domain.ErrConflict)
		}
		if errors.As(err, &pgErr) && pgErr.Code == ForeignKeyViolation {
			return fmt.Errorf("creating audio track: %w: referenced resource not found", domain.ErrInvalidArgument)
		}
		r.logger.ErrorContext(ctx, "Error creating audio track", "error", err, "trackID", track.ID)
		return fmt.Errorf("creating audio track: %w", err)
	}
	r.logger.InfoContext(ctx, "Audio track created successfully", "trackID", track.ID, "title", track.Title)
	return nil
}

func (r *AudioTrackRepository) FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = $1
    `
	track, err := r.scanTrack(ctx, q.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding audio track by ID", "error", err, "trackID", id)
		return nil, fmt.Errorf("finding audio track by ID: %w", err)
	}
	return track, nil
}

func (r *AudioTrackRepository) ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error) {
	q := r.getQuerier(ctx)
	if len(ids) == 0 {
		return []*domain.AudioTrack{}, nil
	}
	uuidStrs := make([]string, len(ids))
	for i, id := range ids {
		uuidStrs[i] = id.String()
	}
	querySimple := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = ANY($1)
    `
	rows, err := q.Query(ctx, querySimple, uuidStrs)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks by IDs", "error", err)
		return nil, fmt.Errorf("listing audio tracks by IDs: %w", err)
	}
	defer rows.Close()
	trackMap := make(map[domain.TrackID]*domain.AudioTrack, len(ids))
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track in ListByIDs", "error", err)
			continue
		}
		trackMap[track.ID] = track
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track rows in ListByIDs", "error", err)
		return nil, fmt.Errorf("iterating track rows: %w", err)
	}
	orderedTracks := make([]*domain.AudioTrack, 0, len(ids))
	for _, id := range ids {
		if t, ok := trackMap[id]; ok {
			orderedTracks = append(orderedTracks, t)
		}
	}
	return orderedTracks, nil
}

// Point 5: Updated to use filters port.ListTracksFilters
func (r *AudioTrackRepository) List(ctx context.Context, filters port.ListTracksFilters, page pagination.Page) ([]*domain.AudioTrack, int, error) {
	q := r.getQuerier(ctx)
	var args []interface{}
	argID := 1
	baseQuery := ` FROM audio_tracks `
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, title, description, language_code, level, duration_ms, minio_bucket, minio_object_key, cover_image_url, uploader_id, is_public, tags, created_at, updated_at ` + baseQuery
	whereClause := " WHERE 1=1"

	// Apply filters from ListTracksFilters
	if filters.Query != nil && *filters.Query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID)
		args = append(args, "%"+*filters.Query+"%")
		argID++
	}
	if filters.LanguageCode != nil && *filters.LanguageCode != "" {
		whereClause += fmt.Sprintf(" AND language_code = $%d", argID)
		args = append(args, *filters.LanguageCode)
		argID++
	}
	if filters.Level != nil && *filters.Level != "" {
		whereClause += fmt.Sprintf(" AND level = $%d", argID)
		args = append(args, *filters.Level)
		argID++
	}
	if filters.IsPublic != nil {
		whereClause += fmt.Sprintf(" AND is_public = $%d", argID)
		args = append(args, *filters.IsPublic)
		argID++
	}
	if filters.UploaderID != nil {
		whereClause += fmt.Sprintf(" AND uploader_id = $%d", argID)
		args = append(args, *filters.UploaderID)
		argID++
	}
	if len(filters.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND tags @> $%d", argID)
		args = append(args, pq.Array(filters.Tags))
		argID++
	}

	var total int
	err := q.QueryRow(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting audio tracks", "error", err, "filters", filters)
		return nil, 0, fmt.Errorf("counting audio tracks: %w", err)
	}
	if total == 0 {
		return []*domain.AudioTrack{}, 0, nil
	}

	orderByClause := " ORDER BY created_at DESC"
	if filters.SortBy != "" {
		// Point 1 & 5: Use duration_ms for sorting if specified
		allowedSorts := map[string]string{"createdAt": "created_at", "title": "title", "durationMs": "duration_ms", "level": "level"}
		dbColumn, ok := allowedSorts[filters.SortBy]
		if ok {
			direction := " ASC"
			if strings.ToLower(filters.SortDirection) == "desc" {
				direction = " DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY %s%s", dbColumn, direction)
		} else {
			r.logger.WarnContext(ctx, "Invalid sort field requested", "sortBy", filters.SortBy)
		}
	}
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)
	finalQuery := selectQuery + whereClause + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks", "error", err, "filters", filters, "page", page)
		return nil, 0, fmt.Errorf("listing audio tracks: %w", err)
	}
	defer rows.Close()
	tracks := make([]*domain.AudioTrack, 0, page.Limit)
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track in List", "error", err)
			continue
		}
		tracks = append(tracks, track)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track rows in List", "error", err)
		return nil, 0, fmt.Errorf("iterating track rows: %w", err)
	}
	return tracks, total, nil
}

func (r *AudioTrackRepository) Update(ctx context.Context, track *domain.AudioTrack) error {
	q := r.getQuerier(ctx)
	track.UpdatedAt = time.Now()
	query := `
		UPDATE audio_tracks SET
			title = $2, description = $3, language_code = $4, level = $5, duration_ms = $6,
			minio_bucket = $7, minio_object_key = $8, cover_image_url = $9, uploader_id = $10,
			is_public = $11, tags = $12, updated_at = $13
		WHERE id = $1
	`
	cmdTag, err := q.Exec(ctx, query,
		track.ID, track.Title, track.Description,
		track.Language.Code(),
		track.Level,
		track.Duration.Milliseconds(), // Point 1: Convert domain Duration to int64 ms
		track.MinioBucket, track.MinioObjectKey, track.CoverImageURL, track.UploaderID,
		track.IsPublic, pq.Array(track.Tags), track.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "audio_tracks_minio_object_key_key") {
				return fmt.Errorf("updating audio track: %w: object key '%s' already exists", domain.ErrConflict, track.MinioObjectKey)
			}
			r.logger.WarnContext(ctx, "Unique constraint violation on audio track update", "constraint", pgErr.ConstraintName, "trackID", track.ID)
			return fmt.Errorf("updating audio track: %w: resource conflict on unique field", domain.ErrConflict)
		}
		r.logger.ErrorContext(ctx, "Error updating audio track", "error", err, "trackID", track.ID)
		return fmt.Errorf("updating audio track: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	r.logger.InfoContext(ctx, "Audio track updated successfully", "trackID", track.ID)
	return nil
}

func (r *AudioTrackRepository) Delete(ctx context.Context, id domain.TrackID) error {
	q := r.getQuerier(ctx)
	query := `DELETE FROM audio_tracks WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting audio track", "error", err, "trackID", id)
		return fmt.Errorf("deleting audio track: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	r.logger.InfoContext(ctx, "Audio track deleted successfully", "trackID", id)
	return nil
}

func (r *AudioTrackRepository) Exists(ctx context.Context, id domain.TrackID) (bool, error) {
	q := r.getQuerier(ctx)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM audio_tracks WHERE id = $1)`
	err := q.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error checking audio track existence", "error", err, "trackID", id)
		return false, fmt.Errorf("checking track existence: %w", err)
	}
	return exists, nil
}

// Point 1: Updated scanTrack
func (r *AudioTrackRepository) scanTrack(ctx context.Context, row RowScanner) (*domain.AudioTrack, error) {
	var track domain.AudioTrack
	var langCode string
	var levelStr string
	var durationMs int64 // Scan into int64
	var tags pq.StringArray
	var uploaderID uuid.NullUUID

	err := row.Scan(
		&track.ID, &track.Title, &track.Description,
		&langCode,
		&levelStr,
		&durationMs, // Scan duration_ms column
		&track.MinioBucket, &track.MinioObjectKey, &track.CoverImageURL,
		&uploaderID,
		&track.IsPublic, &tags, &track.CreatedAt, &track.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	langVO, langErr := domain.NewLanguage(langCode, "")
	if langErr != nil {
		r.logger.ErrorContext(ctx, "Invalid language code found in database", "error", langErr, "langCode", langCode, "trackID", track.ID)
		return nil, fmt.Errorf("invalid language code %s in DB for track %s: %w", langCode, track.ID, langErr)
	}
	track.Language = langVO
	track.Level = domain.AudioLevel(levelStr)
	if !track.Level.IsValid() {
		r.logger.WarnContext(ctx, "Invalid audio level found in database", "level", levelStr, "trackID", track.ID)
		track.Level = domain.LevelUnknown
	}
	// Convert scanned milliseconds back to time.Duration
	track.Duration = time.Duration(durationMs) * time.Millisecond
	track.Tags = tags

	if uploaderID.Valid {
		uid := domain.UserID(uploaderID.UUID)
		track.UploaderID = &uid
	} else {
		track.UploaderID = nil
	}

	return &track, nil
}

var _ port.AudioTrackRepository = (*AudioTrackRepository)(nil)
```

## `internal/adapter/repository/postgres/playbackprogress_repo.go`

```go
// internal/adapter/repository/postgres/playbackprogress_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

type PlaybackProgressRepository struct {
	db         *pgxpool.Pool
	logger     *slog.Logger
	getQuerier func(ctx context.Context) Querier
}

func NewPlaybackProgressRepository(db *pgxpool.Pool, logger *slog.Logger) *PlaybackProgressRepository {
	repo := &PlaybackProgressRepository{
		db:     db,
		logger: logger.With("repository", "PlaybackProgressRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db)
	}
	return repo
}

// --- Interface Implementation ---

func (r *PlaybackProgressRepository) Upsert(ctx context.Context, progress *domain.PlaybackProgress) error {
	q := r.getQuerier(ctx)
	progress.LastListenedAt = time.Now() // Update timestamp before saving
	query := `
        INSERT INTO playback_progress (user_id, track_id, progress_ms, last_listened_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, track_id) DO UPDATE SET
            progress_ms = EXCLUDED.progress_ms,
            last_listened_at = EXCLUDED.last_listened_at
    `
	_, err := q.Exec(ctx, query,
		progress.UserID,
		progress.TrackID,
		progress.Progress.Milliseconds(), // Point 1: Convert domain Duration to int64 ms
		progress.LastListenedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error upserting playback progress", "error", err, "userID", progress.UserID, "trackID", progress.TrackID)
		return fmt.Errorf("upserting playback progress: %w", err)
	}
	r.logger.DebugContext(ctx, "Playback progress upserted", "userID", progress.UserID, "trackID", progress.TrackID, "progressMs", progress.Progress.Milliseconds())
	return nil
}

func (r *PlaybackProgressRepository) Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT user_id, track_id, progress_ms, last_listened_at
        FROM playback_progress
        WHERE user_id = $1 AND track_id = $2
    `
	progress, err := r.scanProgress(ctx, q.QueryRow(ctx, query, userID, trackID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding playback progress", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("finding playback progress: %w", err)
	}
	return progress, nil
}

func (r *PlaybackProgressRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.PlaybackProgress, int, error) {
	q := r.getQuerier(ctx)
	baseQuery := `FROM playback_progress WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT user_id, track_id, progress_ms, last_listened_at ` + baseQuery

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting progress by user: %w", err)
	}
	if total == 0 {
		return []*domain.PlaybackProgress{}, 0, nil
	}

	orderByClause := " ORDER BY last_listened_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3)
	args := []interface{}{userID, page.Limit, page.Offset}
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("listing progress by user: %w", err)
	}
	defer rows.Close()

	progressList := make([]*domain.PlaybackProgress, 0, page.Limit)
	for rows.Next() {
		progress, err := r.scanProgress(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning progress in ListByUser", "error", err)
			continue
		}
		progressList = append(progressList, progress)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating progress rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating progress rows: %w", err)
	}
	return progressList, total, nil
}

// Point 1: Updated scanProgress
func (r *PlaybackProgressRepository) scanProgress(ctx context.Context, row RowScanner) (*domain.PlaybackProgress, error) {
	var p domain.PlaybackProgress
	var progressMs int64 // Scan into int64

	err := row.Scan(
		&p.UserID,
		&p.TrackID,
		&progressMs, // Scan progress_ms column
		&p.LastListenedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert scanned milliseconds back to time.Duration
	p.Progress = time.Duration(progressMs) * time.Millisecond

	return &p, nil
}

var _ port.PlaybackProgressRepository = (*PlaybackProgressRepository)(nil)
```

## `internal/adapter/repository/postgres/db.go`

```go
// internal/adapter/repository/postgres/db.go
package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yvanyang/language-learning-player-api/internal/config" // Adjust import path
)

// NewPgxPool creates a new PostgreSQL connection pool.
func NewPgxPool(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	// Apply pool size settings from config
	pgxConfig.MaxConns = int32(cfg.MaxOpenConns)
	pgxConfig.MinConns = int32(cfg.MaxIdleConns) // pgx uses MinConns roughly like MaxIdleConns
	pgxConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	pgxConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	// Optional: Add logging hook for pgx
	// pgxConfig.ConnConfig.Logger = pgxLogger // Requires implementing pgx logging interface

	logger.Info("Connecting to PostgreSQL", "host", pgxConfig.ConnConfig.Host, "port", pgxConfig.ConnConfig.Port, "db", pgxConfig.ConnConfig.Database)

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err = pool.Ping(ctx); err != nil {
		pool.Close() // Close pool if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return pool, nil
}
```

## `internal/adapter/repository/postgres/bookmark_repo.go`

```go
// internal/adapter/repository/postgres/bookmark_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// BookmarkRepository implements port.BookmarkRepository using PostgreSQL.
type BookmarkRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// getQuerier retrieves the correct Querier (Pool or Tx) from context.
	getQuerier func(ctx context.Context) Querier
}

// NewBookmarkRepository creates a new BookmarkRepository.
func NewBookmarkRepository(db *pgxpool.Pool, logger *slog.Logger) *BookmarkRepository {
	repo := &BookmarkRepository{
		db:     db,
		logger: logger.With("repository", "BookmarkRepository"),
	}
	// Initialize the helper function to get the Querier from context or pool.
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Uses the helper from tx_manager.go
	}
	return repo
}

// --- Interface Implementation ---

// Create inserts a new bookmark record into the database.
func (r *BookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	q := r.getQuerier(ctx) // Get Pool or Transaction Querier

	// Ensure CreatedAt is set (usually done by domain constructor, but good fallback)
	if bookmark.CreatedAt.IsZero() {
		bookmark.CreatedAt = time.Now()
	}
	// Ensure ID is set (should be done by domain constructor)
	if bookmark.ID == (domain.BookmarkID{}) {
		bookmark.ID = domain.NewBookmarkID()
	}

	// SQL uses timestamp_ms column (BIGINT)
	query := `
        INSERT INTO bookmarks (id, user_id, track_id, timestamp_ms, note, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := q.Exec(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.TrackID,
		bookmark.Timestamp.Milliseconds(), // Convert domain's time.Duration to int64 milliseconds
		bookmark.Note,
		bookmark.CreatedAt,
	)

	if err != nil {
		// Consider checking for FK violation on user_id or track_id
		r.logger.ErrorContext(ctx, "Error creating bookmark", "error", err, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
		return fmt.Errorf("creating bookmark: %w", err)
	}

	r.logger.InfoContext(ctx, "Bookmark created successfully", "bookmarkID", bookmark.ID, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
	return nil
}

// FindByID retrieves a bookmark by its unique ID.
func (r *BookmarkRepository) FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	// SQL selects timestamp_ms column
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at
        FROM bookmarks
        WHERE id = $1
    `
	// Use the scanBookmark helper which handles the ms -> duration conversion
	bookmark, err := r.scanBookmark(ctx, q.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map DB error to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding bookmark by ID", "error", err, "bookmarkID", id)
		return nil, fmt.Errorf("finding bookmark by ID: %w", err)
	}
	return bookmark, nil
}

// ListByUserAndTrack retrieves all bookmarks for a specific user on a specific track, ordered by timestamp.
func (r *BookmarkRepository) ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	// SQL selects timestamp_ms and orders by it
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at
        FROM bookmarks
        WHERE user_id = $1 AND track_id = $2
        ORDER BY timestamp_ms ASC
    `
	rows, err := q.Query(ctx, query, userID, trackID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user and track", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("listing bookmarks by user and track: %w", err)
	}
	defer rows.Close()

	bookmarks := make([]*domain.Bookmark, 0)
	for rows.Next() {
		// Use the scanBookmark helper
		bookmark, err := r.scanBookmark(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUserAndTrack", "error", err)
			continue // Skip faulty row? Or return error? Let's skip for now.
		}
		bookmarks = append(bookmarks, bookmark)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUserAndTrack", "error", err)
		return nil, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, nil
}

// ListByUser retrieves a paginated list of all bookmarks for a user, ordered by creation time descending.
func (r *BookmarkRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error) {
	q := r.getQuerier(ctx)
	baseQuery := `FROM bookmarks WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	// SQL selects timestamp_ms column
	selectQuery := `SELECT id, user_id, track_id, timestamp_ms, note, created_at ` + baseQuery

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting bookmarks by user: %w", err)
	}
	if total == 0 {
		return []*domain.Bookmark{}, 0, nil
	}

	// Use page.Limit and page.Offset directly as they are validated/defaulted by the pagination package
	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3) // Arguments start from $2
	args := []interface{}{userID, page.Limit, page.Offset}
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user", "error", err, "userID", userID, "page", page)
		return nil, 0, fmt.Errorf("listing bookmarks by user: %w", err)
	}
	defer rows.Close()

	bookmarks := make([]*domain.Bookmark, 0, page.Limit)
	for rows.Next() {
		// Use the scanBookmark helper
		bookmark, err := r.scanBookmark(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUser", "error", err)
			continue // Skip faulty row
		}
		bookmarks = append(bookmarks, bookmark)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, total, nil
}

// Delete removes a bookmark by its ID. Ownership check is expected to be done in the Usecase layer before calling this.
func (r *BookmarkRepository) Delete(ctx context.Context, id domain.BookmarkID) error {
	q := r.getQuerier(ctx)
	query := `DELETE FROM bookmarks WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting bookmark", "error", err, "bookmarkID", id)
		return fmt.Errorf("deleting bookmark: %w", err)
	}
	// Check if any row was actually deleted
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound // Bookmark ID did not exist
	}
	r.logger.InfoContext(ctx, "Bookmark deleted successfully", "bookmarkID", id)
	return nil
}

// scanBookmark is a helper function to scan a single row into a domain.Bookmark.
// It handles the conversion from the database's timestamp_ms (BIGINT) to domain's Timestamp (time.Duration).
func (r *BookmarkRepository) scanBookmark(ctx context.Context, row RowScanner) (*domain.Bookmark, error) {
	var b domain.Bookmark
	var timestampMs int64 // Variable to scan the BIGINT milliseconds value into

	err := row.Scan(
		&b.ID,
		&b.UserID,
		&b.TrackID,
		&timestampMs, // Scan the timestamp_ms column into int64
		&b.Note,
		&b.CreatedAt,
	)
	if err != nil {
		return nil, err // Propagate scan errors (including pgx.ErrNoRows)
	}

	// Convert the scanned milliseconds back into time.Duration for the domain object
	b.Timestamp = time.Duration(timestampMs) * time.Millisecond

	return &b, nil
}

// Compile-time check to ensure BookmarkRepository satisfies the port.BookmarkRepository interface.
var _ port.BookmarkRepository = (*BookmarkRepository)(nil)
```

## `internal/adapter/repository/postgres/audiocollection_repo.go`

```go
// internal/adapter/repository/postgres/audiocollection_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	// "github.com/jackc/pgx/v5/pgconn" // Needed only if checking specific pg error codes

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"  // Import pagination
)

type AudioCollectionRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewAudioCollectionRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioCollectionRepository {
	repo := &AudioCollectionRepository{
		db:     db,
		logger: logger.With("repository", "AudioCollectionRepository"),
	}
	// Initialize the getQuerier helper function
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper defined in tx_manager.go or here
	}
	return repo
}

// --- Interface Implementation ---

func (r *AudioCollectionRepository) Create(ctx context.Context, collection *domain.AudioCollection) error {
	q := r.getQuerier(ctx) // Get appropriate querier (pool or tx)
	query := `
        INSERT INTO audio_collections (id, title, description, owner_id, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := q.Exec(ctx, query,
		collection.ID, collection.Title, collection.Description, collection.OwnerID,
		collection.Type, collection.CreatedAt, collection.UpdatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error creating audio collection", "error", err, "collectionID", collection.ID, "ownerID", collection.OwnerID)
		// Consider mapping specific DB errors (like FK violation) to domain errors if needed
		return fmt.Errorf("creating audio collection: %w", err)
	}
	r.logger.InfoContext(ctx, "Audio collection created successfully", "collectionID", collection.ID, "title", collection.Title)
	return nil
}

func (r *AudioCollectionRepository) FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, title, description, owner_id, type, created_at, updated_at
        FROM audio_collections
        WHERE id = $1
    `
	collection, err := r.scanCollection(ctx, q.QueryRow(ctx, query, id)) // Pass QueryRow
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding collection by ID", "error", err, "collectionID", id)
		return nil, fmt.Errorf("finding collection by ID: %w", err)
	}
	collection.TrackIDs = make([]domain.TrackID, 0) // Ensure slice is initialized
	return collection, nil
}

func (r *AudioCollectionRepository) FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	q := r.getQuerier(ctx) // Use querier from context for potential transaction

	// 1. Get collection metadata
	collection, err := r.FindByID(ctx, id) // Reuse FindByID which uses getQuerier
	if err != nil {
		return nil, err // Handles ErrNotFound already
	}

	// 2. Get ordered track IDs using the same querier (pool or tx)
	queryTracks := `
        SELECT track_id
        FROM collection_tracks
        WHERE collection_id = $1
        ORDER BY position ASC
    `
	rows, err := q.Query(ctx, queryTracks, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error fetching track IDs for collection", "error", err, "collectionID", id)
		return nil, fmt.Errorf("fetching tracks for collection %s: %w", id, err)
	}
	defer rows.Close()

	trackIDs := make([]domain.TrackID, 0)
	for rows.Next() {
		var trackID domain.TrackID
		if err := rows.Scan(&trackID); err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track ID for collection", "error", err, "collectionID", id)
			return nil, fmt.Errorf("scanning track ID for collection %s: %w", id, err)
		}
		trackIDs = append(trackIDs, trackID)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track IDs for collection", "error", err, "collectionID", id)
		return nil, fmt.Errorf("iterating track IDs for collection %s: %w", id, err)
	}

	collection.TrackIDs = trackIDs
	return collection, nil
}

func (r *AudioCollectionRepository) ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) ([]*domain.AudioCollection, int, error) {
	q := r.getQuerier(ctx)
	args := []interface{}{ownerID}
	argID := 2 // Start arg numbering after ownerID ($1)
	baseQuery := ` FROM audio_collections WHERE owner_id = $1 `
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, title, description, owner_id, type, created_at, updated_at ` + baseQuery

	var total int
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting collections by owner", "error", err, "ownerID", ownerID)
		return nil, 0, fmt.Errorf("counting collections by owner: %w", err)
	}
	if total == 0 {
		return []*domain.AudioCollection{}, 0, nil
	}

	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing collections by owner", "error", err, "ownerID", ownerID)
		return nil, 0, fmt.Errorf("listing collections by owner: %w", err)
	}
	defer rows.Close()
	collections := make([]*domain.AudioCollection, 0, page.Limit)
	for rows.Next() {
		collection, err := r.scanCollection(ctx, rows) // Use RowScanner compatible scan
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning collection in ListByOwner", "error", err)
			continue
		}
		collection.TrackIDs = make([]domain.TrackID, 0)
		collections = append(collections, collection)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating collection rows in ListByOwner", "error", err)
		return nil, 0, fmt.Errorf("iterating collection rows: %w", err)
	}
	return collections, total, nil
}

// UpdateMetadata updates only title and description. Ownership check via WHERE clause.
func (r *AudioCollectionRepository) UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error {
	q := r.getQuerier(ctx)
	collection.UpdatedAt = time.Now()
	query := `
        UPDATE audio_collections SET
            title = $2, description = $3, updated_at = $4
        WHERE id = $1 AND owner_id = $5 -- Ensure owner matches
    `
	cmdTag, err := q.Exec(ctx, query,
		collection.ID, collection.Title, collection.Description, collection.UpdatedAt, collection.OwnerID,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error updating collection metadata", "error", err, "collectionID", collection.ID)
		return fmt.Errorf("updating collection metadata: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		exists, _ := r.exists(ctx, collection.ID) // Check existence
		if !exists {
			return domain.ErrNotFound
		}
		return domain.ErrPermissionDenied // Assume owner mismatch if exists but not updated
	}
	r.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collection.ID)
	return nil
}

// ManageTracks replaces the entire set of tracks associated with a collection.
// This method now expects to run within a transaction context provided by the Usecase.
func (r *AudioCollectionRepository) ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	q := r.getQuerier(ctx) // Gets Tx if called within TxManager.Execute, otherwise Pool

	// 1. Delete existing tracks for the collection
	deleteQuery := `DELETE FROM collection_tracks WHERE collection_id = $1`
	if _, err := q.Exec(ctx, deleteQuery, collectionID); err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete old tracks in ManageTracks", "error", err, "collectionID", collectionID)
		return fmt.Errorf("deleting old tracks: %w", err)
	}

	// 2. Insert new tracks with correct positions
	if len(orderedTrackIDs) > 0 {
		insertQuery := `
            INSERT INTO collection_tracks (collection_id, track_id, position)
            VALUES ($1, $2, $3)
        `
		// Use pgx's Batch for potentially better performance than looping Exec
		batch := &pgx.Batch{}
		for i, trackID := range orderedTrackIDs {
			batch.Queue(insertQuery, collectionID, trackID, i)
		}

		br := q.SendBatch(ctx, batch)
		defer br.Close() // Ensure batch results are closed

		// Check results for each queued insert
		for i := 0; i < len(orderedTrackIDs); i++ {
			_, err := br.Exec()
			if err != nil {
				r.logger.ErrorContext(ctx, "Failed to insert track in ManageTracks batch", "error", err, "collectionID", collectionID, "position", i)
				// Consider checking for FK violation on track_id here
				return fmt.Errorf("inserting track at position %d: %w", i, err)
			}
		}
	}

	// 3. Update the collection's updated_at timestamp
	updateTsQuery := `UPDATE audio_collections SET updated_at = $1 WHERE id = $2`
	if _, err := q.Exec(ctx, updateTsQuery, time.Now(), collectionID); err != nil {
		r.logger.ErrorContext(ctx, "Failed to update collection timestamp in ManageTracks", "error", err, "collectionID", collectionID)
		return fmt.Errorf("updating collection timestamp: %w", err)
	}

	r.logger.DebugContext(ctx, "ManageTracks repo operations completed (within transaction)", "collectionID", collectionID, "trackCount", len(orderedTrackIDs))
	return nil // Usecase layer handles commit/rollback
}

func (r *AudioCollectionRepository) Delete(ctx context.Context, id domain.CollectionID) error {
	q := r.getQuerier(ctx)
	// Ownership check is done in Usecase layer
	// ON DELETE CASCADE handles collection_tracks entries automatically
	query := `DELETE FROM audio_collections WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting audio collection", "error", err, "collectionID", id)
		return fmt.Errorf("deleting audio collection: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	r.logger.InfoContext(ctx, "Audio collection deleted successfully", "collectionID", id)
	return nil
}

// --- Helper Methods ---

// exists checks if a collection with the given ID exists.
func (r *AudioCollectionRepository) exists(ctx context.Context, id domain.CollectionID) (bool, error) {
	q := r.getQuerier(ctx)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM audio_collections WHERE id = $1)`
	err := q.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error checking collection existence", "error", err, "collectionID", id)
		return false, fmt.Errorf("checking collection existence: %w", err)
	}
	return exists, nil
}

// scanCollection scans a single row into a domain.AudioCollection.
// CHANGED: Accepts RowScanner interface
func (r *AudioCollectionRepository) scanCollection(ctx context.Context, row RowScanner) (*domain.AudioCollection, error) {
	var collection domain.AudioCollection
	err := row.Scan(
		&collection.ID, &collection.Title, &collection.Description, &collection.OwnerID,
		&collection.Type, &collection.CreatedAt, &collection.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// Ensure implementation satisfies the interface
var _ port.AudioCollectionRepository = (*AudioCollectionRepository)(nil)
```

## `internal/adapter/repository/postgres/refreshtoken_repo.go`

```go
// ==========================================================
// FILE: internal/adapter/repository/postgres/refreshtoken_repo.go (NEW FILE)
// ==========================================================
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

type RefreshTokenRepository struct {
	db         *pgxpool.Pool
	logger     *slog.Logger
	getQuerier func(ctx context.Context) Querier
}

func NewRefreshTokenRepository(db *pgxpool.Pool, logger *slog.Logger) *RefreshTokenRepository {
	repo := &RefreshTokenRepository{
		db:     db,
		logger: logger.With("repository", "RefreshTokenRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db)
	}
	return repo
}

func (r *RefreshTokenRepository) Save(ctx context.Context, tokenData *port.RefreshTokenData) error {
	q := r.getQuerier(ctx)
	query := `
        INSERT INTO refresh_tokens (token_hash, user_id, expires_at, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (token_hash) DO UPDATE SET -- Should not happen if hashes are unique
            expires_at = EXCLUDED.expires_at,
            user_id = EXCLUDED.user_id -- Update user_id just in case (though unlikely to change)
    `
	if tokenData.CreatedAt.IsZero() {
		tokenData.CreatedAt = time.Now() // Ensure created_at is set
	}

	_, err := q.Exec(ctx, query,
		tokenData.TokenHash,
		tokenData.UserID,
		tokenData.ExpiresAt,
		tokenData.CreatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error saving refresh token", "error", err, "userID", tokenData.UserID)
		// Consider checking for FK violation on user_id -> domain.ErrInvalidArgument ?
		return fmt.Errorf("saving refresh token: %w", err)
	}
	r.logger.DebugContext(ctx, "Refresh token saved", "userID", tokenData.UserID, "expiresAt", tokenData.ExpiresAt)
	return nil
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*port.RefreshTokenData, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT token_hash, user_id, expires_at, created_at
        FROM refresh_tokens
        WHERE token_hash = $1
    `
	row := q.QueryRow(ctx, query, tokenHash)
	tokenData, err := r.scanTokenData(ctx, row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map DB error to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding refresh token by hash", "error", err)
		return nil, fmt.Errorf("finding refresh token by hash: %w", err)
	}
	return tokenData, nil
}

func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	q := r.getQuerier(ctx)
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	cmdTag, err := q.Exec(ctx, query, tokenHash)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting refresh token by hash", "error", err)
		return fmt.Errorf("deleting refresh token by hash: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		// Not finding the token to delete is not necessarily an error in a logout/refresh flow
		r.logger.DebugContext(ctx, "Refresh token hash not found for deletion", "tokenHash", tokenHash) // Log hash prefix? No.
		return domain.ErrNotFound                                                                       // Or return nil? ErrNotFound seems slightly more informative.
	}
	r.logger.DebugContext(ctx, "Refresh token deleted by hash")
	return nil
}

func (r *RefreshTokenRepository) DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error) {
	q := r.getQuerier(ctx)
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	cmdTag, err := q.Exec(ctx, query, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting refresh tokens by user ID", "error", err, "userID", userID)
		return 0, fmt.Errorf("deleting refresh tokens by user ID: %w", err)
	}
	deletedCount := cmdTag.RowsAffected()
	r.logger.InfoContext(ctx, "Refresh tokens deleted by user ID", "userID", userID, "count", deletedCount)
	return deletedCount, nil
}

func (r *RefreshTokenRepository) scanTokenData(ctx context.Context, row RowScanner) (*port.RefreshTokenData, error) {
	var data port.RefreshTokenData
	err := row.Scan(
		&data.TokenHash,
		&data.UserID,
		&data.ExpiresAt,
		&data.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

var _ port.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
```

## `internal/adapter/repository/postgres/user_repo.go`

```go
// internal/adapter/repository/postgres/user_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings" // Import strings package
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn" // Import pgconn for PgError
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/internal/port"   // Adjust import path
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *pgxpool.Pool, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.With("repository", "UserRepository"), // Add context to logger
	}
}

// --- Interface Implementation ---

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email.String(), // Use string representation of value object
		user.Name,
		user.HashedPassword,
		user.GoogleID,
		user.AuthProvider,
		user.ProfileImageURL,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		// Check if the error is a PostgreSQL error and specifically a unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			// Check which constraint was violated based on its name (defined in migration)
			// Common constraint names are like <table_name>_<column_name>_key
			if strings.Contains(pgErr.ConstraintName, "users_email_key") {
				// Specific error message for email conflict
				return fmt.Errorf("creating user: %w: email already exists", domain.ErrConflict)
			}
			if strings.Contains(pgErr.ConstraintName, "users_google_id_key") {
				// Specific error message for Google ID conflict
				return fmt.Errorf("creating user: %w: google ID already exists", domain.ErrConflict)
			}
			// Generic conflict if constraint name is unknown or not specifically handled
			r.logger.WarnContext(ctx, "Unique constraint violation on user creation", "constraint", pgErr.ConstraintName, "userID", user.ID)
			return fmt.Errorf("creating user: %w: resource conflict on unique field", domain.ErrConflict)
		}
		// If it's not a unique violation, log as internal error
		r.logger.ErrorContext(ctx, "Error creating user", "error", err, "userID", user.ID)
		return fmt.Errorf("creating user: %w", err)
	}
	r.logger.InfoContext(ctx, "User created successfully", "userID", user.ID, "email", user.Email.String())
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
        SELECT id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at
        FROM users
        WHERE id = $1
    `
	user, err := r.scanUser(ctx, r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding user by ID", "error", err, "userID", id)
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	query := `
        SELECT id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at
        FROM users
        WHERE email = $1
    `
	user, err := r.scanUser(ctx, r.db.QueryRow(ctx, query, email.String()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding user by Email", "error", err, "email", email.String())
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error) {
	// For now, only handling Google ID. Extend if more providers are added.
	if provider != domain.AuthProviderGoogle {
		return nil, fmt.Errorf("finding user by provider ID: provider '%s' not supported", provider)
	}

	query := `
        SELECT id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at
        FROM users
        WHERE google_id = $1 AND auth_provider = $2
    `
	user, err := r.scanUser(ctx, r.db.QueryRow(ctx, query, providerUserID, provider))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding user by Provider ID", "error", err, "provider", provider, "providerUserID", providerUserID)
		return nil, fmt.Errorf("finding user by provider ID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	// Ensure updated_at is current
	user.UpdatedAt = time.Now()

	query := `
        UPDATE users
        SET email = $2, name = $3, password_hash = $4, google_id = $5, auth_provider = $6, profile_image_url = $7, updated_at = $8
        WHERE id = $1
    `
	cmdTag, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email.String(),
		user.Name,
		user.HashedPassword,
		user.GoogleID,
		user.AuthProvider,
		user.ProfileImageURL,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		// Check for unique constraint violation on update (e.g., changing email to an existing one)
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "users_email_key") {
				return fmt.Errorf("updating user: %w: email already exists", domain.ErrConflict)
			}
			if strings.Contains(pgErr.ConstraintName, "users_google_id_key") {
				return fmt.Errorf("updating user: %w: google ID already exists", domain.ErrConflict)
			}
			r.logger.WarnContext(ctx, "Unique constraint violation on user update", "constraint", pgErr.ConstraintName, "userID", user.ID)
			return fmt.Errorf("updating user: %w: resource conflict on unique field", domain.ErrConflict)
		}
		r.logger.ErrorContext(ctx, "Error updating user", "error", err, "userID", user.ID)
		return fmt.Errorf("updating user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		// If no rows were affected, the user ID likely didn't exist
		return domain.ErrNotFound // Or return a specific "update failed" error? ErrNotFound seems reasonable.
	}

	r.logger.InfoContext(ctx, "User updated successfully", "userID", user.ID)
	return nil
}

// EmailExists checks if a user with the given email already exists.
// ADDED: Implementation for EmailExists
func (r *UserRepository) EmailExists(ctx context.Context, email domain.Email) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email.String()).Scan(&exists)
	if err != nil {
		// Don't treat pgx.ErrNoRows as an error here, EXISTS correctly returns false
		// Log other potential errors
		r.logger.ErrorContext(ctx, "Error checking email existence", "error", err, "email", email.String())
		return false, fmt.Errorf("checking email existence: %w", err)
	}
	return exists, nil
}

// --- Helper Methods ---

// scanUser is a helper function to scan a row into a domain.User object.
// It handles the conversion from SQL null types to Go pointers/value objects.
func (r *UserRepository) scanUser(ctx context.Context, row pgx.Row) (*domain.User, error) {
	var user domain.User
	var emailStr string // Scan email into a simple string first

	err := row.Scan(
		&user.ID,
		&emailStr,
		&user.Name,
		&user.HashedPassword, // Directly scans into *string (handles NULL)
		&user.GoogleID,       // Directly scans into *string (handles NULL)
		&user.AuthProvider,
		&user.ProfileImageURL, // Directly scans into *string (handles NULL)
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err // Let caller handle pgx.ErrNoRows or other errors
	}

	// Convert scanned email string to domain.Email value object
	emailVO, voErr := domain.NewEmail(emailStr)
	if voErr != nil {
		// This should ideally not happen if DB constraint is correct, but handle defensively
		r.logger.ErrorContext(ctx, "Invalid email format found in database", "error", voErr, "email", emailStr, "userID", user.ID)
		return nil, fmt.Errorf("invalid email format in DB for user %s: %w", user.ID, voErr)
	}
	user.Email = emailVO

	return &user, nil
}

// Compile-time check to ensure UserRepository satisfies the port.UserRepository interface
var _ port.UserRepository = (*UserRepository)(nil)
```

## `internal/adapter/repository/postgres/tx_manager.go`

```go
// internal/adapter/repository/postgres/tx_manager.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn" // ADDED: Import pgconn
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/port" // Adjust import path
)

type txKey struct{} // Private key type for context value

// Querier defines the common methods needed from pgxpool.Pool and pgx.Tx
// Used by repositories to work transparently with or without a transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	// Add CopyFrom if needed
}

// RowScanner defines an interface for scanning database rows, compatible with pgx.Row and pgx.Rows.
// ADDED: Interface definition
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// getQuerier attempts to retrieve a pgx.Tx from the context.
// If not found, it returns the original *pgxpool.Pool.
// Both satisfy the Querier interface.
func getQuerier(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx // Use transaction if available
	}
	return pool // Fallback to pool
}

// TxManager implements the port.TransactionManager interface for PostgreSQL using pgx.
type TxManager struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewTxManager creates a new TxManager.
func NewTxManager(pool *pgxpool.Pool, logger *slog.Logger) *TxManager {
	return &TxManager{
		pool:   pool,
		logger: logger.With("component", "TxManager"),
	}
}

// Begin starts a new transaction and stores the pgx.Tx handle in the context.
func (tm *TxManager) Begin(ctx context.Context) (context.Context, error) {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		tm.logger.WarnContext(ctx, "Attempted to begin transaction within an existing transaction")
		return ctx, fmt.Errorf("transaction already in progress")
	}

	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		tm.logger.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	tm.logger.DebugContext(ctx, "Transaction begun")
	txCtx := context.WithValue(ctx, txKey{}, tx)
	return txCtx, nil
}

// Commit commits the transaction stored in the context.
func (tm *TxManager) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		tm.logger.WarnContext(ctx, "Commit called without an active transaction in context")
		return fmt.Errorf("no transaction found in context to commit")
	}

	err := tx.Commit(ctx)
	if err != nil {
		tm.logger.ErrorContext(ctx, "Failed to commit transaction", "error", err)
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	tm.logger.DebugContext(ctx, "Transaction committed")
	return nil
}

// Rollback rolls back the transaction stored in the context.
func (tm *TxManager) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		tm.logger.WarnContext(ctx, "Rollback called without an active transaction in context")
		return fmt.Errorf("no transaction found in context to rollback")
	}

	err := tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) { // Don't error if already closed
		tm.logger.ErrorContext(ctx, "Failed to rollback transaction", "error", err)
		return fmt.Errorf("transaction rollback failed: %w", err)
	}
	if err == nil {
		tm.logger.DebugContext(ctx, "Transaction rolled back")
	}
	return nil // Return nil even if ErrTxClosed
}

// Execute runs the given function within a transaction.
func (tm *TxManager) Execute(ctx context.Context, fn func(txCtx context.Context) error) (err error) {
	txCtx, beginErr := tm.Begin(ctx)
	if beginErr != nil {
		return beginErr
	}

	defer func() {
		if p := recover(); p != nil {
			rbErr := tm.Rollback(txCtx)
			err = fmt.Errorf("panic occurred during transaction: %v", p)
			if rbErr != nil {
				tm.logger.ErrorContext(ctx, "Failed to rollback after panic", "rollbackError", rbErr, "panic", p)
				err = fmt.Errorf("panic (%v) occurred and rollback failed: %w", p, rbErr)
			} else {
				tm.logger.ErrorContext(ctx, "Transaction rolled back due to panic", "panic", p)
			}
		}
	}()

	if fnErr := fn(txCtx); fnErr != nil {
		if rbErr := tm.Rollback(txCtx); rbErr != nil {
			tm.logger.ErrorContext(ctx, "Failed to rollback after function error", "rollbackError", rbErr, "functionError", fnErr)
			return fmt.Errorf("function failed (%w) and rollback failed: %w", fnErr, rbErr)
		}
		tm.logger.WarnContext(ctx, "Transaction rolled back due to function error", "error", fnErr)
		return fnErr // Return the original error from the function
	}

	if commitErr := tm.Commit(txCtx); commitErr != nil {
		tm.logger.ErrorContext(ctx, "Failed to commit transaction, attempting rollback", "commitError", commitErr)
		// Rollback might fail if the connection is broken after commit failure
		if rbErr := tm.Rollback(txCtx); rbErr != nil {
			tm.logger.ErrorContext(ctx, "Failed to rollback after commit failure", "rollbackError", rbErr, "commitError", commitErr)
			// Return both errors? Or just the commit error? Commit error is primary.
			return fmt.Errorf("commit failed (%w) and subsequent rollback failed: %w", commitErr, rbErr)
		}
		return commitErr // Return the commit error
	}

	tm.logger.DebugContext(ctx, "Transaction executed and committed successfully")
	return nil
}

// Compile-time check
var _ port.TransactionManager = (*TxManager)(nil)
```

## `internal/adapter/repository/postgres/errors.go`

```go
// internal/adapter/repository/postgres/errors.go
package postgres

// PostgreSQL错误代码常量
const (
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
)
```

## `internal/adapter/service/google_auth/google_adapter.go`

```go
// internal/adapter/service/google_auth/google_adapter.go
package googleauthadapter

import (
	"context"
	"fmt"
	"log/slog"

	// Use the official Google idtoken verifier
	"google.golang.org/api/idtoken"

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/internal/port"   // Adjust import path
)

// GoogleAuthService implements the port.ExternalAuthService interface for Google.
type GoogleAuthService struct {
	googleClientID string
	logger         *slog.Logger
	// http.Client is implicitly used by idtoken.Validate, but could be injected for testing/customization
}

// NewGoogleAuthService creates a new GoogleAuthService.
func NewGoogleAuthService(clientID string, logger *slog.Logger) (*GoogleAuthService, error) {
	if clientID == "" {
		return nil, fmt.Errorf("google client ID cannot be empty")
	}
	return &GoogleAuthService{
		googleClientID: clientID,
		logger:         logger.With("service", "GoogleAuthService"),
	}, nil
}

// VerifyGoogleToken verifies a Google ID token and returns standardized user info.
func (s *GoogleAuthService) VerifyGoogleToken(ctx context.Context, idToken string) (*port.ExternalUserInfo, error) {
	if idToken == "" {
		return nil, fmt.Errorf("%w: google ID token cannot be empty", domain.ErrAuthenticationFailed)
	}

	// Validate the token using Google's library.
	// It checks signature, expiration, issuer ('accounts.google.com' or 'https://accounts.google.com'),
	// and audience (must match our client ID).
	payload, err := idtoken.Validate(ctx, idToken, s.googleClientID)
	if err != nil {
		s.logger.WarnContext(ctx, "Google ID token validation failed", "error", err)
		// Map validation errors to our domain error
		// The library might return specific error types, but a general failure is often sufficient here.
		return nil, fmt.Errorf("%w: invalid google token: %v", domain.ErrAuthenticationFailed, err)
	}

	// Token is valid, extract claims.
	// Standard claims reference: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	userID, ok := payload.Claims["sub"].(string) // Subject - Google's unique user ID
	if !ok || userID == "" {
		s.logger.ErrorContext(ctx, "Google ID token missing 'sub' (subject) claim", "claims", payload.Claims)
		return nil, fmt.Errorf("%w: missing required user identifier in token", domain.ErrAuthenticationFailed)
	}

	email, _ := payload.Claims["email"].(string)                // Email claim
	emailVerified, _ := payload.Claims["email_verified"].(bool) // Email verified claim
	name, _ := payload.Claims["name"].(string)                  // Name claim
	picture, _ := payload.Claims["picture"].(string)            // Picture URL claim

	// Basic check: ensure we have at least an email if email_verified is true
	if emailVerified && email == "" {
		s.logger.WarnContext(ctx, "Google ID token claims email_verified=true but email is missing", "subject", userID)
		// Decide policy: reject or proceed without email? Rejecting is safer.
		// return nil, fmt.Errorf("%w: verified email missing in token", domain.ErrAuthenticationFailed)
	}
	// If email is not verified by Google, maybe we shouldn't trust it for login/registration?
	// Current policy: Use the email regardless, but store the verification status.
	// Consider enforcing emailVerified == true if needed.

	s.logger.InfoContext(ctx, "Google ID token verified successfully", "subject", userID, "email", email)

	// Map to our standardized ExternalUserInfo struct
	userInfo := &port.ExternalUserInfo{
		Provider:        domain.AuthProviderGoogle,
		ProviderUserID:  userID,
		Email:           email,
		IsEmailVerified: emailVerified,
		Name:            name,
		// PictureURL:      &picture, // Assign if picture is not empty
	}
	if picture != "" {
		userInfo.PictureURL = &picture
	}

	return userInfo, nil
}

// Compile-time check
var _ port.ExternalAuthService = (*GoogleAuthService)(nil)
```

## `internal/adapter/service/minio/minio_adapter.go`

```go
// internal/adapter/service/minio/minio_adapter.go
package minioadapter // Use a distinct name like minioadapter or minioservice

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/yvanyang/language-learning-player-api/internal/config" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/internal/port"   // Adjust import path
)

// MinioStorageService implements the port.FileStorageService interface using MinIO.
type MinioStorageService struct {
	client        *minio.Client
	defaultBucket string
	defaultExpiry time.Duration
	logger        *slog.Logger
}

// NewMinioStorageService creates a new MinioStorageService.
func NewMinioStorageService(cfg config.MinioConfig, logger *slog.Logger) (*MinioStorageService, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("minio configuration (Endpoint, AccessKeyID, SecretAccessKey, BucketName) cannot be empty")
	}

	log := logger.With("service", "MinioStorageService")
	log.Info("Initializing MinIO client", "endpoint", cfg.Endpoint, "ssl", cfg.UseSSL, "bucket", cfg.BucketName)

	// Initialize minio client object.
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Error("Failed to initialize MinIO client", "error", err)
		return nil, fmt.Errorf("minio client initialization failed: %w", err)
	}

	// Optional: Ping MinIO to check connectivity (MinIO server needs to be running)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		log.Error("Failed to check if MinIO bucket exists", "error", err, "bucket", cfg.BucketName)
		// Decide if this is fatal or not. Log warning and proceed.
	}
	if !exists {
		log.Warn("Default MinIO bucket does not exist. Consider creating it.", "bucket", cfg.BucketName)
	} else {
		log.Info("MinIO bucket found", "bucket", cfg.BucketName)
	}

	return &MinioStorageService{
		client:        minioClient,
		defaultBucket: cfg.BucketName,
		defaultExpiry: cfg.PresignExpiry, // Use expiry from config
		logger:        log,
	}, nil
}

// GetPresignedGetURL generates a temporary URL for downloading an object.
func (s *MinioStorageService) GetPresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	if expiry <= 0 {
		expiry = s.defaultExpiry
	}
	reqParams := make(url.Values)

	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, reqParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate presigned GET URL", "error", err, "bucket", bucket, "key", objectKey)
		return "", fmt.Errorf("failed to get presigned URL for %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.DebugContext(ctx, "Generated presigned GET URL", "bucket", bucket, "key", objectKey, "expiry", expiry)
	return presignedURL.String(), nil
}

// DeleteObject removes an object from MinIO storage.
func (s *MinioStorageService) DeleteObject(ctx context.Context, bucket, objectKey string) error {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	opts := minio.RemoveObjectOptions{}

	err := s.client.RemoveObject(ctx, bucket, objectKey, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete object from MinIO", "error", err, "bucket", bucket, "key", objectKey)
		return fmt.Errorf("failed to delete object %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.InfoContext(ctx, "Deleted object from MinIO", "bucket", bucket, "key", objectKey)
	return nil
}

// GetPresignedPutURL generates a temporary URL for uploading an object.
func (s *MinioStorageService) GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	if expiry <= 0 {
		expiry = s.defaultExpiry
	}

	// policy := minio.NewPostPolicy() // Note: Using PostPolicy for Put? Check SDK if PresignedPutObject directly is better. Let's stick to PresignedPutObject for simplicity.
	// policy.SetBucket(bucket)
	// policy.SetKey(objectKey)
	// policy.SetExpires(time.Now().UTC().Add(expiry))

	presignedURL, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate presigned PUT URL", "error", err, "bucket", bucket, "key", objectKey)
		return "", fmt.Errorf("failed to get presigned PUT URL for %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.DebugContext(ctx, "Generated presigned PUT URL", "bucket", bucket, "key", objectKey, "expiry", expiry)
	return presignedURL.String(), nil
}

// ObjectExists checks if an object exists in the specified bucket using StatObject.
// ADDED: Implementation for ObjectExists
func (s *MinioStorageService) ObjectExists(ctx context.Context, bucket, objectKey string) (bool, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}

	_, err := s.client.StatObject(ctx, bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		// Check if the error code indicates that the object was not found.
		if errResponse.Code == "NoSuchKey" {
			s.logger.DebugContext(ctx, "Object check: Object not found", "bucket", bucket, "key", objectKey)
			return false, nil // Object does not exist, but this is not an error in checking
		}
		// For any other error (network issue, permissions, etc.), log and return the error.
		s.logger.ErrorContext(ctx, "Failed to stat object", "error", err, "bucket", bucket, "key", objectKey)
		return false, fmt.Errorf("failed to check object existence for %s/%s: %w", bucket, objectKey, err)
	}
	// If StatObject returns no error, the object exists.
	s.logger.DebugContext(ctx, "Object check: Object found", "bucket", bucket, "key", objectKey)
	return true, nil
}

// Compile-time check to ensure MinioStorageService satisfies the port.FileStorageService interface
var _ port.FileStorageService = (*MinioStorageService)(nil)
```

## `internal/port/params.go`

```go
// internal/port/params.go
package port

import (
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// === Use Case Layer Input/Result Structs ===

// ListTracksInput defines parameters for listing/searching tracks at the use case layer.
// It embeds pagination.Page.
type ListTracksInput struct {
	Query         *string            // Search query (title, description, maybe tags)
	LanguageCode  *string            // Filter by language code
	Level         *domain.AudioLevel // Filter by level
	IsPublic      *bool              // Filter by public status
	UploaderID    *domain.UserID     // Filter by uploader
	Tags          []string           // Filter by tags (match any)
	SortBy        string             // e.g., "createdAt", "title", "durationMs"
	SortDirection string             // "asc" or "desc"
	Page          pagination.Page    // Embed pagination parameters
}

// ListProgressInput defines parameters for listing user progress at the use case layer.
type ListProgressInput struct {
	UserID domain.UserID
	Page   pagination.Page
}

// ListBookmarksInput defines parameters for listing user bookmarks at the use case layer.
type ListBookmarksInput struct {
	UserID        domain.UserID
	TrackIDFilter *domain.TrackID // Optional filter by track
	Page          pagination.Page
}

// RequestUploadResult holds the result of requesting an upload URL.
type RequestUploadResult struct {
	UploadURL string
	ObjectKey string
}

// CompleteUploadInput holds the data needed to finalize an upload and create a track record.
type CompleteUploadInput struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string
	DurationMs    int64
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}

// --- Batch Request Input ---
type BatchRequestUploadInputItem struct {
	Filename    string
	ContentType string
}
type BatchRequestUploadInput struct {
	Files []BatchRequestUploadInputItem
}

// --- Batch URL Result ---
type BatchURLResultItem struct {
	OriginalFilename string
	ObjectKey        string
	UploadURL        string
	// Using string for Error is simpler for JSON marshalling in batch results,
	// though less type-safe internally. Acknowledge this trade-off.
	Error string
}

// --- Batch Complete Input ---
type BatchCompleteItem struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string
	DurationMs    int64
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}
type BatchCompleteInput struct {
	Tracks []BatchCompleteItem
}

// --- Batch Complete Result ---
type BatchCompleteResultItem struct {
	ObjectKey string
	Success   bool
	TrackID   string // Use string for UUID here as it's just data
	// Using string for Error is simpler for JSON marshalling in batch results,
	// though less type-safe internally. Acknowledge this trade-off.
	Error string
}

// === Audio Track Details Result (Used by AudioContentUseCase) ===

// GetAudioTrackDetailsResult holds the combined result for getting track details.
type GetAudioTrackDetailsResult struct {
	Track         *domain.AudioTrack
	PlayURL       string
	UserProgress  *domain.PlaybackProgress // Nil if user not logged in or no progress
	UserBookmarks []*domain.Bookmark       // Empty slice if user not logged in or no bookmarks
}
```

## `internal/port/service.go`

```go
// internal/port/service.go
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

// --- External Service Interfaces ---

// FileStorageService defines the contract for interacting with object storage.
type FileStorageService interface {
	// GetPresignedGetURL returns a temporary, signed URL for reading a private object.
	// expiry defines how long the URL should be valid.
	GetPresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)

	// GetPresignedPutURL returns a temporary, signed URL for uploading/overwriting an object.
	// The client MUST use the HTTP PUT method with this URL.
	GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)

	// DeleteObject removes an object from storage.
	DeleteObject(ctx context.Context, bucket, objectKey string) error

	// ObjectExists checks if an object exists in the specified bucket.
	// ADDED: ObjectExists method
	ObjectExists(ctx context.Context, bucket, objectKey string) (bool, error)

	// TODO: Consider adding methods for getting metadata, etc. if needed.
}

// ExternalUserInfo contains standardized user info retrieved from an external identity provider.
type ExternalUserInfo struct {
	Provider        domain.AuthProvider // e.g., "google"
	ProviderUserID  string              // Unique ID from the provider (e.g., Google subject ID)
	Email           string              // Email address provided by the provider
	IsEmailVerified bool                // Whether the provider claims the email is verified
	Name            string
	PictureURL      *string // Optional profile picture URL
}

// ExternalAuthService defines the contract for verifying external authentication credentials.
type ExternalAuthService interface {
	// VerifyGoogleToken verifies a Google ID token and returns standardized user info.
	// Returns domain.ErrAuthenticationFailed if the token is invalid or verification fails.
	VerifyGoogleToken(ctx context.Context, idToken string) (*ExternalUserInfo, error)

	// Add methods for other providers if needed (e.g., VerifyFacebookToken, VerifyAppleToken)
}

// --- Internal Helper Service Interfaces ---

// SecurityHelper defines cryptographic operations needed by use cases.
type SecurityHelper interface {
	// HashPassword generates a secure hash (e.g., bcrypt) of the password.
	HashPassword(ctx context.Context, password string) (string, error)
	// CheckPasswordHash compares a plain password with a stored hash.
	CheckPasswordHash(ctx context.Context, password, hash string) bool
	// GenerateJWT creates a signed JWT (Access Token) for the given user ID.
	GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error)
	// VerifyJWT validates a JWT string and returns the UserID contained within.
	// Returns domain.ErrUnauthenticated or domain.ErrAuthenticationFailed on failure.
	VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error)

	GenerateRefreshTokenValue() (string, error)     // ADDED
	HashRefreshTokenValue(tokenValue string) string // ADDED
}

// REMOVED UserUseCase interface from here
```

## `internal/port/repository.go`

```go
// internal/port/repository.go
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// --- Repository Interfaces ---

// RefreshTokenData holds the details of a stored refresh token.
type RefreshTokenData struct {
	TokenHash string // SHA-256 hash
	UserID    domain.UserID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Save(ctx context.Context, tokenData *RefreshTokenData) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshTokenData, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error) // Returns number of tokens deleted
}

// UserRepository defines the persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	// ADDED: EmailExists method
	EmailExists(ctx context.Context, email domain.Email) (bool, error)
}

// ListTracksFilters defines parameters for filtering/searching tracks at the repository layer.
// RENAMED from ListTracksParams
type ListTracksFilters struct {
	Query         *string            // Search query (title, description, maybe tags)
	LanguageCode  *string            // Filter by language code
	Level         *domain.AudioLevel // Filter by level
	IsPublic      *bool              // Filter by public status
	UploaderID    *domain.UserID     // Filter by uploader
	Tags          []string           // Filter by tags (match any)
	SortBy        string             // e.g., "createdAt", "title", "durationMs" (DB column name might differ)
	SortDirection string             // "asc" or "desc"
}

// AudioTrackRepository defines the persistence operations for AudioTrack entities.
type AudioTrackRepository interface {
	FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error)
	ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error)
	// List retrieves a paginated list of tracks based on filter and sort parameters.
	// RENAMED params type to ListTracksFilters
	List(ctx context.Context, filters ListTracksFilters, page pagination.Page) (tracks []*domain.AudioTrack, total int, err error)
	Create(ctx context.Context, track *domain.AudioTrack) error
	Update(ctx context.Context, track *domain.AudioTrack) error
	Delete(ctx context.Context, id domain.TrackID) error
	Exists(ctx context.Context, id domain.TrackID) (bool, error)
}

// AudioCollectionRepository defines the persistence operations for AudioCollection entities.
type AudioCollectionRepository interface {
	FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) (collections []*domain.AudioCollection, total int, err error)
	Create(ctx context.Context, collection *domain.AudioCollection) error
	UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error
	ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	Delete(ctx context.Context, id domain.CollectionID) error
}

// PlaybackProgressRepository defines the persistence operations for PlaybackProgress entities.
type PlaybackProgressRepository interface {
	Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	Upsert(ctx context.Context, progress *domain.PlaybackProgress) error
	ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) (progressList []*domain.PlaybackProgress, total int, err error)
}

// BookmarkRepository defines the persistence operations for Bookmark entities.
type BookmarkRepository interface {
	FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error)
	ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error)
	ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) (bookmarks []*domain.Bookmark, total int, err error)
	Create(ctx context.Context, bookmark *domain.Bookmark) error
	Delete(ctx context.Context, id domain.BookmarkID) error
}

// --- Transaction Management ---

type Tx interface{}

type TransactionManager interface {
	Begin(ctx context.Context) (TxContext context.Context, err error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error
}
```

## `internal/port/usecase.go`

```go
// ============================================
// FILE: internal/port/usecase.go (MODIFIED)
// ============================================
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// AuthResult holds both access and refresh tokens, and registration status.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	IsNewUser    bool // Only relevant for external auth methods like Google sign-in
}

// AuthUseCase defines the methods for the Auth use case layer.
type AuthUseCase interface {
	// RegisterWithPassword registers a new user with email/password.
	// Returns the created user, auth tokens, and error.
	RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, AuthResult, error)

	// LoginWithPassword authenticates a user with email/password.
	// Returns auth tokens and error.
	LoginWithPassword(ctx context.Context, emailStr, password string) (AuthResult, error)

	// AuthenticateWithGoogle handles login or registration via Google ID Token.
	// Returns auth tokens and error. The IsNewUser field in AuthResult indicates
	// if a new account was created during this process.
	AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (AuthResult, error)

	// RefreshAccessToken validates a refresh token and issues a new pair of access/refresh tokens.
	RefreshAccessToken(ctx context.Context, refreshTokenValue string) (AuthResult, error)

	// Logout invalidates the provided refresh token.
	Logout(ctx context.Context, refreshTokenValue string) error
}

// AudioContentUseCase defines the methods for the Audio Content use case layer.
type AudioContentUseCase interface {
	GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*GetAudioTrackDetailsResult, error)
	ListTracks(ctx context.Context, input ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error)
	CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error)
	GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error)
	GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error)
	UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error
	UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error
}

// UserActivityUseCase defines the methods for the User Activity use case layer.
type UserActivityUseCase interface {
	RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error
	GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	ListUserProgress(ctx context.Context, params ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error)
	CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)
	ListBookmarks(ctx context.Context, params ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error)
	DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error
}

// UserUseCase defines the interface for user-related operations (e.g., profile)
type UserUseCase interface {
	GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error)
}

// UploadUseCase defines the methods for the Upload use case layer.
type UploadUseCase interface {
	RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*RequestUploadResult, error)
	CompleteUpload(ctx context.Context, userID domain.UserID, req CompleteUploadInput) (*domain.AudioTrack, error)
	RequestBatchUpload(ctx context.Context, userID domain.UserID, req BatchRequestUploadInput) ([]BatchURLResultItem, error)
	CompleteBatchUpload(ctx context.Context, userID domain.UserID, req BatchCompleteInput) ([]BatchCompleteResultItem, error)
}
```

## `internal/usecase/audio_content_uc.go`

```go
// internal/usecase/audio_content_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/config"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// AudioContentUseCase handles business logic related to audio tracks and collections.
type AudioContentUseCase struct {
	trackRepo      port.AudioTrackRepository
	collectionRepo port.AudioCollectionRepository
	storageService port.FileStorageService
	txManager      port.TransactionManager
	// ADDED: Inject activity repo to fetch user specific data in GetAudioTrackDetails
	progressRepo  port.PlaybackProgressRepository
	bookmarkRepo  port.BookmarkRepository
	presignExpiry time.Duration
	cdnBaseURL    *url.URL
	logger        *slog.Logger
}

// NewAudioContentUseCase creates a new AudioContentUseCase.
// ADDED: progressRepo and bookmarkRepo dependencies
func NewAudioContentUseCase(
	cfg config.Config,
	tr port.AudioTrackRepository,
	cr port.AudioCollectionRepository,
	ss port.FileStorageService,
	tm port.TransactionManager,
	pr port.PlaybackProgressRepository, // Added
	br port.BookmarkRepository, // Added
	log *slog.Logger,
) *AudioContentUseCase {
	if tm == nil {
		log.Warn("AudioContentUseCase created without TransactionManager implementation. Transactional operations will fail.")
	}
	var parsedCdnBaseURL *url.URL
	var parseErr error
	if cfg.CDN.BaseURL != "" {
		parsedCdnBaseURL, parseErr = url.Parse(cfg.CDN.BaseURL)
		if parseErr != nil {
			log.Warn("Invalid CDN BaseURL in config, CDN rewriting disabled", "url", cfg.CDN.BaseURL, "error", parseErr)
			parsedCdnBaseURL = nil
		} else {
			log.Info("CDN Rewriting Enabled", "baseUrl", parsedCdnBaseURL.String())
		}
	}

	return &AudioContentUseCase{
		trackRepo:      tr,
		collectionRepo: cr,
		storageService: ss,
		txManager:      tm,
		progressRepo:   pr, // Added
		bookmarkRepo:   br, // Added
		presignExpiry:  cfg.Minio.PresignExpiry,
		cdnBaseURL:     parsedCdnBaseURL,
		logger:         log.With("usecase", "AudioContentUseCase"),
	}
}

// --- Track Use Cases ---

// Point 4: GetAudioTrackDetails retrieves details and user-specific info, returns result struct.
func (uc *AudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*port.GetAudioTrackDetailsResult, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx) // Check if user is logged in

	track, err := uc.trackRepo.FindByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Audio track not found", "trackID", trackID)
		} else {
			uc.logger.ErrorContext(ctx, "Failed to get audio track from repository", "error", err, "trackID", trackID)
		}
		return nil, err // Propagate error (NotFound or Internal)
	}

	// Authorization check (Example: only public tracks accessible anonymously)
	if !track.IsPublic && !userAuthenticated {
		uc.logger.WarnContext(ctx, "Anonymous user attempted to access private track", "trackID", trackID)
		return nil, domain.ErrUnauthenticated // Require login for private tracks
	}
	// Further checks (ownership, subscription) could go here if needed

	// Generate Presigned URL
	presignedURLStr, err := uc.storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, uc.presignExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate presigned URL for track", "error", err, "trackID", trackID)
		presignedURLStr = "" // Continue but log the error, URL will be empty
	}

	// Rewrite URL if CDN is configured
	finalPlayURL := uc.rewriteURLForCDN(ctx, presignedURLStr)

	result := &port.GetAudioTrackDetailsResult{
		Track:   track,
		PlayURL: finalPlayURL,
		// UserProgress and UserBookmarks will be filled below if user is authenticated
	}

	// Fetch user-specific data if authenticated
	if userAuthenticated {
		// Fetch Progress
		progress, errProg := uc.progressRepo.Find(ctx, userID, trackID)
		if errProg != nil && !errors.Is(errProg, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to get user progress for track details", "error", errProg, "trackID", trackID, "userID", userID)
			// Continue without progress, error is logged
		} else if errProg == nil {
			result.UserProgress = progress
		}

		// Fetch Bookmarks
		bookmarks, errBook := uc.bookmarkRepo.ListByUserAndTrack(ctx, userID, trackID)
		if errBook != nil {
			uc.logger.ErrorContext(ctx, "Failed to get user bookmarks for track details", "error", errBook, "trackID", trackID, "userID", userID)
			// Continue without bookmarks, error is logged
		} else {
			result.UserBookmarks = bookmarks // Assign even if empty slice
		}
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved audio track details", "trackID", trackID, "authenticated", userAuthenticated)
	return result, nil
}

// rewriteURLForCDN is a helper to rewrite presigned URL if CDN is configured.
func (uc *AudioContentUseCase) rewriteURLForCDN(ctx context.Context, originalURL string) string {
	if uc.cdnBaseURL == nil || originalURL == "" {
		return originalURL // No CDN configured or no original URL
	}

	parsedOriginalURL, parseErr := url.Parse(originalURL)
	if parseErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to parse original presigned URL for CDN rewriting", "url", originalURL, "error", parseErr)
		return originalURL // Fallback to original URL on parsing error
	}

	// Construct rewritten URL using CDN base and original path + query
	rewrittenURL := &url.URL{
		Scheme:   uc.cdnBaseURL.Scheme,
		Host:     uc.cdnBaseURL.Host,
		Path:     parsedOriginalURL.Path,
		RawQuery: parsedOriginalURL.RawQuery, // Preserve signature etc.
	}

	finalURL := rewrittenURL.String()
	uc.logger.DebugContext(ctx, "Rewrote presigned URL for CDN", "original", originalURL, "rewritten", finalURL)
	return finalURL
}

// Point 5: ListTracks now takes input port.ListTracksInput
func (uc *AudioContentUseCase) ListTracks(ctx context.Context, input port.ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error) {
	pageParams := pagination.NewPageFromOffset(input.Page.Limit, input.Page.Offset)

	// Map Usecase input to Repository filters
	repoFilters := port.ListTracksFilters{
		Query:         input.Query,
		LanguageCode:  input.LanguageCode,
		Level:         input.Level,
		IsPublic:      input.IsPublic,
		UploaderID:    input.UploaderID,
		Tags:          input.Tags,
		SortBy:        input.SortBy,
		SortDirection: input.SortDirection,
	}

	tracks, total, err := uc.trackRepo.List(ctx, repoFilters, pageParams)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list audio tracks from repository", "error", err, "filters", repoFilters, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve track list: %w", err)
	}
	uc.logger.InfoContext(ctx, "Successfully listed audio tracks", "count", len(tracks), "total", total, "input", input)
	return tracks, total, pageParams, nil
}

// --- Collection Use Cases --- (No changes needed for the requested points in these methods)

func (uc *AudioContentUseCase) CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	if uc.txManager == nil {
		return nil, fmt.Errorf("internal configuration error: transaction manager not available")
	}

	collection, err := domain.NewAudioCollection(title, description, userID, colType)
	if err != nil {
		return nil, err
	}

	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		if err := uc.collectionRepo.Create(txCtx, collection); err != nil {
			return fmt.Errorf("saving collection metadata: %w", err)
		}
		if len(initialTrackIDs) > 0 {
			exists, validateErr := uc.validateTrackIDsExist(txCtx, initialTrackIDs)
			if validateErr != nil {
				return fmt.Errorf("validating initial tracks: %w", validateErr)
			}
			if !exists {
				return fmt.Errorf("%w: one or more initial track IDs do not exist", domain.ErrInvalidArgument)
			}
			if err := uc.collectionRepo.ManageTracks(txCtx, collection.ID, initialTrackIDs); err != nil {
				return fmt.Errorf("adding initial tracks: %w", err)
			}
			collection.TrackIDs = initialTrackIDs
		}
		return nil
	})

	if finalErr != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed during collection creation", "error", finalErr, "collectionID", collection.ID, "userID", userID)
		return nil, fmt.Errorf("failed to create collection: %w", finalErr)
	}
	uc.logger.InfoContext(ctx, "Audio collection created", "collectionID", collection.ID, "userID", userID)
	return collection, nil
}

func (uc *AudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx)
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil { /* Handle NotFound, log other errors */
		return nil, err
	}
	if !userAuthenticated || collection.OwnerID != userID { /* Log, return permission denied */
		return nil, domain.ErrPermissionDenied
	}
	uc.logger.InfoContext(ctx, "Successfully retrieved collection details", "collectionID", collectionID, "trackCount", len(collection.TrackIDs))
	return collection, nil
}

func (uc *AudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	if len(collection.TrackIDs) == 0 {
		return []*domain.AudioTrack{}, nil
	}
	tracks, err := uc.trackRepo.ListByIDs(ctx, collection.TrackIDs)
	if err != nil { /* Log, return wrapped error */
		return nil, fmt.Errorf("failed to retrieve track details for collection: %w", err)
	}
	uc.logger.InfoContext(ctx, "Successfully retrieved tracks for collection", "collectionID", collectionID, "trackCount", len(tracks))
	return tracks, nil
}

func (uc *AudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return domain.ErrUnauthenticated
	}
	if title == "" {
		return fmt.Errorf("%w: collection title cannot be empty", domain.ErrInvalidArgument)
	}
	tempCollection := &domain.AudioCollection{ID: collectionID, OwnerID: userID, Title: title, Description: description}
	err := uc.collectionRepo.UpdateMetadata(ctx, tempCollection)
	if err != nil { /* Log if not NotFound/PermissionDenied, return err */
		return err
	}
	uc.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collectionID, "userID", userID)
	return nil
}

func (uc *AudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return domain.ErrUnauthenticated
	}
	if uc.txManager == nil {
		return fmt.Errorf("internal configuration error: transaction manager not available")
	}

	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		collection, err := uc.collectionRepo.FindByID(txCtx, collectionID)
		if err != nil {
			return err
		}
		if collection.OwnerID != userID {
			return domain.ErrPermissionDenied
		}
		if len(orderedTrackIDs) > 0 {
			exists, validateErr := uc.validateTrackIDsExist(txCtx, orderedTrackIDs)
			if validateErr != nil {
				return fmt.Errorf("validating tracks: %w", validateErr)
			}
			if !exists {
				return fmt.Errorf("%w: one or more track IDs do not exist", domain.ErrInvalidArgument)
			}
		}
		if err := uc.collectionRepo.ManageTracks(txCtx, collectionID, orderedTrackIDs); err != nil {
			return fmt.Errorf("updating collection tracks in repository: %w", err)
		}
		return nil
	})

	if finalErr != nil { /* Log, return wrapped error */
		return fmt.Errorf("failed to update collection tracks: %w", finalErr)
	}
	uc.logger.InfoContext(ctx, "Collection tracks updated", "collectionID", collectionID, "userID", userID, "trackCount", len(orderedTrackIDs))
	return nil
}

func (uc *AudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return domain.ErrUnauthenticated
	}
	collection, err := uc.collectionRepo.FindByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection.OwnerID != userID {
		return domain.ErrPermissionDenied
	}
	err = uc.collectionRepo.Delete(ctx, collectionID)
	if err != nil { /* Log if not NotFound, return err */
		return err
	}
	uc.logger.InfoContext(ctx, "Collection deleted", "collectionID", collectionID, "userID", userID)
	return nil
}

// Helper remains the same
func (uc *AudioContentUseCase) validateTrackIDsExist(ctx context.Context, trackIDs []domain.TrackID) (bool, error) {
	if len(trackIDs) == 0 {
		return true, nil
	}
	existingTracks, err := uc.trackRepo.ListByIDs(ctx, trackIDs)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to validate track IDs existence", "error", err)
		return false, fmt.Errorf("failed to verify tracks: %w", err)
	}
	if len(existingTracks) != len(trackIDs) { /* Log missing IDs */
		return false, nil
	}
	return true, nil
}

var _ port.AudioContentUseCase = (*AudioContentUseCase)(nil)
```

## `internal/usecase/user_uc.go`

```go
// internal/usecase/user_uc.go
package usecase

import (
	"context"
	"log/slog"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// userUseCase implements the port.UserUseCase interface.
type userUseCase struct {
	userRepo port.UserRepository
	logger   *slog.Logger
}

// NewUserUseCase creates a new UserUseCase.
func NewUserUseCase(ur port.UserRepository, log *slog.Logger) port.UserUseCase {
	return &userUseCase{
		userRepo: ur,
		logger:   log,
	}
}

// GetUserProfile retrieves a user's profile by their ID.
func (uc *userUseCase) GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		// FindByID should return domain.ErrNotFound if not found
		uc.logger.WarnContext(ctx, "Failed to get user profile", "userID", userID, "error", err)
		return nil, err
	}
	// Note: We might want to selectively return fields or use a specific DTO
	// instead of the full domain.User if there's sensitive info like password hash.
	// However, the repository FindByID should already exclude the hash if necessary.
	return user, nil
}
```

## `internal/usecase/auth_uc.go`

```go
// ============================================
// FILE: internal/usecase/auth_uc.go (MODIFIED)
// ============================================
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/config"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

type AuthUseCase struct {
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository // Dependency for refresh token storage
	secHelper        port.SecurityHelper
	extAuthService   port.ExternalAuthService
	cfg              config.JWTConfig // Store the whole JWT config for expiries
	logger           *slog.Logger
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(
	cfg config.JWTConfig, // Pass whole JWTConfig
	ur port.UserRepository,
	rtr port.RefreshTokenRepository, // Inject RefreshTokenRepository
	sh port.SecurityHelper,
	eas port.ExternalAuthService,
	log *slog.Logger,
) *AuthUseCase {
	if eas == nil {
		log.Warn("AuthUseCase created without ExternalAuthService implementation.")
	}
	// Add check for RefreshTokenRepository
	if rtr == nil {
		// Log as error because refresh token functionality will be broken
		log.Error("AuthUseCase created without RefreshTokenRepository implementation. Refresh tokens cannot be stored or validated.")
	}
	return &AuthUseCase{
		userRepo:         ur,
		refreshTokenRepo: rtr, // Assign injected repo
		secHelper:        sh,
		extAuthService:   eas,
		cfg:              cfg, // Store config
		logger:           log.With("usecase", "AuthUseCase"),
	}
}

// generateAndStoreTokens is a helper to create access/refresh tokens and store the refresh token hash.
func (uc *AuthUseCase) generateAndStoreTokens(ctx context.Context, userID domain.UserID) (accessToken, refreshTokenValue string, err error) {
	// Ensure repo dependency is available before proceeding
	if uc.refreshTokenRepo == nil {
		uc.logger.ErrorContext(ctx, "RefreshTokenRepository is nil, cannot generate/store tokens", "userID", userID)
		return "", "", fmt.Errorf("internal server error: authentication system misconfigured")
	}

	accessToken, err = uc.secHelper.GenerateJWT(ctx, userID, uc.cfg.AccessTokenExpiry)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenValue, err = uc.secHelper.GenerateRefreshTokenValue()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token value: %w", err)
	}

	refreshTokenHash := uc.secHelper.HashRefreshTokenValue(refreshTokenValue)
	expiresAt := time.Now().Add(uc.cfg.RefreshTokenExpiry)

	tokenData := &port.RefreshTokenData{
		TokenHash: refreshTokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(), // Set creation time here
	}

	// Save the refresh token data to the repository
	if err = uc.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshTokenValue, nil
}

// RegisterWithPassword handles user registration with email and password.
func (uc *AuthUseCase) RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, port.AuthResult, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid email provided during registration", "email", emailStr, "error", err)
		return nil, port.AuthResult{}, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	exists, err := uc.userRepo.EmailExists(ctx, emailVO)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error checking email existence", "error", err, "email", emailStr)
		return nil, port.AuthResult{}, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		uc.logger.WarnContext(ctx, "Registration attempt with existing email", "email", emailStr)
		return nil, port.AuthResult{}, fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}

	if len(password) < 8 {
		return nil, port.AuthResult{}, fmt.Errorf("%w: password must be at least 8 characters long", domain.ErrInvalidArgument)
	}

	hashedPassword, err := uc.secHelper.HashPassword(ctx, password)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to hash password during registration", "error", err)
		return nil, port.AuthResult{}, fmt.Errorf("failed to process password: %w", err)
	}

	user, err := domain.NewLocalUser(emailStr, name, hashedPassword)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create domain user object", "error", err)
		return nil, port.AuthResult{}, fmt.Errorf("failed to create user data: %w", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new user to repository", "error", err, "userID", user.ID)
		// Create should map unique constraints to ErrConflict
		return nil, port.AuthResult{}, fmt.Errorf("failed to register user: %w", err)
	}

	// Generate and store tokens
	accessToken, refreshToken, tokenErr := uc.generateAndStoreTokens(ctx, user.ID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store tokens after registration", "error", tokenErr, "userID", user.ID)
		return nil, port.AuthResult{}, fmt.Errorf("failed to finalize registration session: %w", tokenErr)
	}

	uc.logger.InfoContext(ctx, "User registered successfully via password", "userID", user.ID, "email", emailStr)
	return user, port.AuthResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// LoginWithPassword handles user login with email and password.
func (uc *AuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (port.AuthResult, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		return port.AuthResult{}, domain.ErrAuthenticationFailed
	}
	user, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Login attempt for non-existent email", "email", emailStr)
			return port.AuthResult{}, domain.ErrAuthenticationFailed
		}
		uc.logger.ErrorContext(ctx, "Error finding user by email during login", "error", err, "email", emailStr)
		return port.AuthResult{}, fmt.Errorf("failed during login process: %w", err)
	}
	if user.AuthProvider != domain.AuthProviderLocal || user.HashedPassword == nil {
		uc.logger.WarnContext(ctx, "Login attempt for user with non-local provider or no password", "email", emailStr, "userID", user.ID, "provider", user.AuthProvider)
		return port.AuthResult{}, domain.ErrAuthenticationFailed
	}
	if !uc.secHelper.CheckPasswordHash(ctx, password, *user.HashedPassword) {
		uc.logger.WarnContext(ctx, "Incorrect password provided for user", "email", emailStr, "userID", user.ID)
		return port.AuthResult{}, domain.ErrAuthenticationFailed
	}

	// Generate and store tokens
	accessToken, refreshToken, tokenErr := uc.generateAndStoreTokens(ctx, user.ID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store tokens during login", "error", tokenErr, "userID", user.ID)
		return port.AuthResult{}, fmt.Errorf("failed to finalize login session: %w", tokenErr)
	}

	uc.logger.InfoContext(ctx, "User logged in successfully via password", "userID", user.ID)
	return port.AuthResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// AuthenticateWithGoogle handles login or registration via Google ID Token.
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (port.AuthResult, error) {
	if uc.extAuthService == nil {
		uc.logger.ErrorContext(ctx, "ExternalAuthService not configured for Google authentication")
		return port.AuthResult{}, fmt.Errorf("google authentication is not enabled")
	}

	extInfo, err := uc.extAuthService.VerifyGoogleToken(ctx, googleIdToken)
	if err != nil {
		return port.AuthResult{}, err // Propagate verification error
	}

	var targetUser *domain.User
	var isNewUser bool

	// Check if user exists by Provider ID first
	user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
	if err == nil {
		// Case 1: User found by Google ID -> Login success
		uc.logger.InfoContext(ctx, "User authenticated via existing Google ID", "userID", user.ID, "googleID", extInfo.ProviderUserID)
		targetUser = user
		isNewUser = false
	} else if errors.Is(err, domain.ErrNotFound) {
		// User not found by Google ID, proceed to check by email (if available)
		if extInfo.Email != "" {
			emailVO, emailErr := domain.NewEmail(extInfo.Email)
			if emailErr != nil {
				uc.logger.WarnContext(ctx, "Invalid email format received from Google token", "error", emailErr, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
				return port.AuthResult{}, fmt.Errorf("%w: invalid email format from provider", domain.ErrAuthenticationFailed)
			}

			userByEmail, errEmail := uc.userRepo.FindByEmail(ctx, emailVO)
			if errEmail == nil {
				// Case 2: User found by email -> Conflict (Strategy C)
				uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but Google ID did not match or was not linked", "email", extInfo.Email, "existingUserID", userByEmail.ID, "existingProvider", userByEmail.AuthProvider)
				return port.AuthResult{}, fmt.Errorf("%w: email is already associated with a different account", domain.ErrConflict)
			} else if !errors.Is(errEmail, domain.ErrNotFound) {
				uc.logger.ErrorContext(ctx, "Error finding user by email", "error", errEmail, "email", extInfo.Email)
				return port.AuthResult{}, fmt.Errorf("database error during authentication: %w", errEmail)
			}
			uc.logger.DebugContext(ctx, "Email not found, proceeding to create new Google user", "email", extInfo.Email)
		} else {
			uc.logger.InfoContext(ctx, "Google token verified, but no email provided. Proceeding to create new user based on Google ID only.", "googleID", extInfo.ProviderUserID)
		}

		// Case 3: Create new user
		uc.logger.InfoContext(ctx, "Creating new user via Google authentication", "googleID", extInfo.ProviderUserID, "email", extInfo.Email)
		newUser, errCreate := domain.NewGoogleUser(extInfo.Email, extInfo.Name, extInfo.ProviderUserID, extInfo.PictureURL)
		if errCreate != nil {
			uc.logger.ErrorContext(ctx, "Failed to create new Google user domain object", "error", errCreate, "extInfo", extInfo)
			return port.AuthResult{}, fmt.Errorf("failed to process user data from Google: %w", errCreate)
		}
		if errDb := uc.userRepo.Create(ctx, newUser); errDb != nil {
			uc.logger.ErrorContext(ctx, "Failed to save new Google user to repository", "error", errDb, "googleID", newUser.GoogleID, "email", newUser.Email.String())
			// Create maps unique constraints to ErrConflict
			return port.AuthResult{}, fmt.Errorf("failed to create new user account: %w", errDb)
		}
		uc.logger.InfoContext(ctx, "New user created successfully via Google", "userID", newUser.ID, "email", newUser.Email.String())
		targetUser = newUser
		isNewUser = true

	} else {
		// Handle unexpected database errors during provider ID lookup
		uc.logger.ErrorContext(ctx, "Error finding user by provider ID", "error", err, "provider", extInfo.Provider, "providerUserID", extInfo.ProviderUserID)
		return port.AuthResult{}, fmt.Errorf("database error during authentication: %w", err)
	}

	// Generate and store tokens for the targetUser (either found or newly created)
	accessToken, refreshToken, tokenErr := uc.generateAndStoreTokens(ctx, targetUser.ID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store tokens for Google auth", "error", tokenErr, "userID", targetUser.ID)
		return port.AuthResult{}, fmt.Errorf("failed to finalize authentication session: %w", tokenErr)
	}

	return port.AuthResult{AccessToken: accessToken, RefreshToken: refreshToken, IsNewUser: isNewUser}, nil
}

// RefreshAccessToken validates a refresh token, revokes it, and issues new access/refresh tokens.
func (uc *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshTokenValue string) (port.AuthResult, error) {
	// Ensure repo dependency is available before proceeding
	if uc.refreshTokenRepo == nil {
		uc.logger.ErrorContext(ctx, "RefreshTokenRepository is nil, cannot refresh token")
		return port.AuthResult{}, fmt.Errorf("internal server error: authentication system misconfigured")
	}

	if refreshTokenValue == "" {
		return port.AuthResult{}, fmt.Errorf("%w: refresh token is required", domain.ErrAuthenticationFailed)
	}

	tokenHash := uc.secHelper.HashRefreshTokenValue(refreshTokenValue)

	// Find the token data by hash
	tokenData, err := uc.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Refresh token hash not found in storage during refresh attempt")
			// It's crucial to return an authentication error here to prevent probing attacks
			return port.AuthResult{}, domain.ErrAuthenticationFailed
		}
		uc.logger.ErrorContext(ctx, "Error finding refresh token by hash", "error", err)
		return port.AuthResult{}, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check expiry
	if time.Now().After(tokenData.ExpiresAt) {
		uc.logger.WarnContext(ctx, "Expired refresh token presented", "userID", tokenData.UserID, "expiresAt", tokenData.ExpiresAt)
		// Delete the expired token (best effort)
		_ = uc.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
		return port.AuthResult{}, fmt.Errorf("%w: refresh token expired", domain.ErrAuthenticationFailed)
	}

	// --- Rotation: Invalidate the old token ---
	delErr := uc.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
	if delErr != nil && !errors.Is(delErr, domain.ErrNotFound) {
		// Log the failure but proceed to issue new tokens.
		// The user presented a valid token, failure to delete shouldn't block refresh.
		// A background job could clean up orphaned tokens later if needed.
		uc.logger.ErrorContext(ctx, "Failed to delete used refresh token during rotation, proceeding anyway", "error", delErr, "userID", tokenData.UserID)
	}

	// --- Issue new tokens ---
	newAccessToken, newRefreshTokenValue, tokenErr := uc.generateAndStoreTokens(ctx, tokenData.UserID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store new tokens during refresh", "error", tokenErr, "userID", tokenData.UserID)
		// This is a more critical failure. User might be left logged out.
		return port.AuthResult{}, fmt.Errorf("failed to issue new tokens after validating refresh token: %w", tokenErr)
	}

	uc.logger.InfoContext(ctx, "Access token refreshed successfully", "userID", tokenData.UserID)
	return port.AuthResult{AccessToken: newAccessToken, RefreshToken: newRefreshTokenValue}, nil
}

// Logout invalidates a specific refresh token.
func (uc *AuthUseCase) Logout(ctx context.Context, refreshTokenValue string) error {
	// Ensure repo dependency is available before proceeding
	if uc.refreshTokenRepo == nil {
		uc.logger.ErrorContext(ctx, "RefreshTokenRepository is nil, cannot logout")
		return fmt.Errorf("internal server error: authentication system misconfigured")
	}

	if refreshTokenValue == "" {
		// Nothing to invalidate if no token is provided.
		uc.logger.DebugContext(ctx, "Logout called with empty refresh token, nothing to do.")
		return nil
	}

	tokenHash := uc.secHelper.HashRefreshTokenValue(refreshTokenValue)

	// Delete the token by hash. ErrNotFound is acceptable.
	err := uc.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		// Log actual errors during deletion
		uc.logger.ErrorContext(ctx, "Failed to delete refresh token during logout", "error", err)
		// Return a generic internal error to the client
		return fmt.Errorf("failed to process logout request: %w", err)
	}

	if errors.Is(err, domain.ErrNotFound) {
		uc.logger.InfoContext(ctx, "Logout attempt for a refresh token that was already invalid or deleted.")
	} else {
		uc.logger.InfoContext(ctx, "User session logged out (refresh token invalidated).")
	}
	return nil
}

// Compile-time check to ensure AuthUseCase satisfies the port.AuthUseCase interface
var _ port.AuthUseCase = (*AuthUseCase)(nil)
```

## `internal/usecase/upload_uc.go`

```go
// ============================================
// FILE: internal/usecase/upload_uc.go (MODIFIED)
// ============================================
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/yvanyang/language-learning-player-api/internal/config"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// UploadUseCase handles the business logic for file uploads.
type UploadUseCase struct {
	trackRepo      port.AudioTrackRepository
	storageService port.FileStorageService
	txManager      port.TransactionManager
	logger         *slog.Logger
	minioBucket    string
}

// NewUploadUseCase creates a new UploadUseCase.
func NewUploadUseCase(
	cfg config.MinioConfig,
	tr port.AudioTrackRepository,
	ss port.FileStorageService,
	tm port.TransactionManager,
	log *slog.Logger,
) *UploadUseCase {
	if tm == nil {
		log.Warn("UploadUseCase created without TransactionManager implementation. Batch completion will not be transactional.")
	}
	if ss == nil {
		log.Error("UploadUseCase created without FileStorageService implementation. Uploads will fail.")
	}
	return &UploadUseCase{
		trackRepo:      tr,
		storageService: ss,
		txManager:      tm,
		logger:         log.With("usecase", "UploadUseCase"),
		minioBucket:    cfg.BucketName,
	}
}

// RequestUpload generates a presigned PUT URL for the client to upload a single file.
func (uc *UploadUseCase) RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*port.RequestUploadResult, error) {
	log := uc.logger.With("userID", userID.String(), "filename", filename, "contentType", contentType)

	if filename == "" {
		return nil, fmt.Errorf("%w: filename cannot be empty", domain.ErrInvalidArgument)
	}
	_ = uc.validateContentType(contentType)

	objectKey := uc.generateObjectKey(userID, filename)
	log = log.With("objectKey", objectKey)

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}
	uploadURLExpiry := 15 * time.Minute
	uploadURL, err := uc.storageService.GetPresignedPutURL(ctx, uc.minioBucket, objectKey, uploadURLExpiry)
	if err != nil {
		log.Error("Failed to get presigned PUT URL", "error", err)
		return nil, fmt.Errorf("failed to prepare upload: %w", err)
	}

	log.Info("Generated presigned URL for file upload")
	result := &port.RequestUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}
	return result, nil
}

// CompleteUpload finalizes the upload process by creating an AudioTrack record in the database.
// CHANGED: Parameter type to port.CompleteUploadInput
func (uc *UploadUseCase) CompleteUpload(ctx context.Context, userID domain.UserID, input port.CompleteUploadInput) (*domain.AudioTrack, error) {
	log := uc.logger.With("userID", userID.String(), "objectKey", input.ObjectKey)

	// CHANGED: Use fields from input
	if err := uc.validateCompleteUploadRequest(ctx, userID, input.ObjectKey, input.Title, input.LanguageCode, input.DurationMs, input.Level); err != nil {
		return nil, err
	}

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}
	exists, checkErr := uc.storageService.ObjectExists(ctx, uc.minioBucket, input.ObjectKey)
	if checkErr != nil {
		log.Error("Failed to check object existence in storage", "error", checkErr)
		return nil, fmt.Errorf("failed to verify upload status: %w", checkErr)
	}
	if !exists {
		log.Warn("Attempted to complete upload for a non-existent object in storage")
		return nil, fmt.Errorf("%w: uploaded file not found in storage for the given key", domain.ErrInvalidArgument)
	}

	// CHANGED: Pass input struct
	track, err := uc.createDomainTrack(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	err = uc.trackRepo.Create(ctx, track)
	if err != nil {
		log.Error("Failed to create audio track record in repository", "error", err, "trackID", track.ID)
		if errors.Is(err, domain.ErrConflict) {
			log.Warn("Conflict during track creation, potentially duplicate object key", "objectKey", input.ObjectKey)
			return nil, fmt.Errorf("%w: track identifier conflict, possibly duplicate object key", domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to save track information: %w", err) // Internal error
	}

	log.Info("Upload completed and track record created", "trackID", track.ID)
	return track, nil
}

// --- Batch Upload Methods ---

// RequestBatchUpload generates presigned PUT URLs for multiple files.
// CHANGED: Parameter type to port.BatchRequestUploadInput
func (uc *UploadUseCase) RequestBatchUpload(ctx context.Context, userID domain.UserID, input port.BatchRequestUploadInput) ([]port.BatchURLResultItem, error) {
	log := uc.logger.With("userID", userID.String(), "batchSize", len(input.Files))
	log.Info("Requesting batch upload URLs")

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}

	results := make([]port.BatchURLResultItem, len(input.Files))
	uploadURLExpiry := 15 * time.Minute

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, fileReq := range input.Files { // Iterate over input.Files
		wg.Add(1)
		go func(index int, f port.BatchRequestUploadInputItem) {
			defer wg.Done()
			itemLog := log.With("originalFilename", f.Filename, "contentType", f.ContentType)
			responseItem := port.BatchURLResultItem{
				OriginalFilename: f.Filename,
			}

			if f.Filename == "" {
				responseItem.Error = "filename cannot be empty"
			}
			_ = uc.validateContentType(f.ContentType)

			objectKey := uc.generateObjectKey(userID, f.Filename)
			responseItem.ObjectKey = objectKey
			itemLog = itemLog.With("objectKey", objectKey)

			if responseItem.Error == "" {
				uploadURL, err := uc.storageService.GetPresignedPutURL(ctx, uc.minioBucket, objectKey, uploadURLExpiry)
				if err != nil {
					itemLog.Error("Failed to get presigned PUT URL for batch item", "error", err)
					responseItem.Error = "failed to prepare upload URL"
				} else {
					responseItem.UploadURL = uploadURL
				}
			}

			mu.Lock()
			results[index] = responseItem
			mu.Unlock()
		}(i, fileReq)
	}

	wg.Wait()
	log.Info("Finished generating batch upload URLs")
	return results, nil
}

// CompleteBatchUpload finalizes multiple uploads within a single database transaction.
// CHANGED: Parameter type to port.BatchCompleteInput
func (uc *UploadUseCase) CompleteBatchUpload(ctx context.Context, userID domain.UserID, input port.BatchCompleteInput) ([]port.BatchCompleteResultItem, error) {
	log := uc.logger.With("userID", userID.String(), "batchSize", len(input.Tracks))
	log.Info("Attempting to complete batch upload")

	if uc.txManager == nil {
		return nil, fmt.Errorf("internal server error: batch processing misconfigured")
	}
	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}

	var processingErr error

	preCheckFailed := false
	validatedItems := make([]port.BatchCompleteItem, 0, len(input.Tracks))
	tempResults := make([]port.BatchCompleteResultItem, len(input.Tracks))

	for i, trackReq := range input.Tracks { // Iterate over input.Tracks
		itemLog := log.With("objectKey", trackReq.ObjectKey, "title", trackReq.Title)
		resultItem := port.BatchCompleteResultItem{
			ObjectKey: trackReq.ObjectKey,
			Success:   false,
		}

		validationErr := uc.validateCompleteUploadRequest(ctx, userID, trackReq.ObjectKey, trackReq.Title, trackReq.LanguageCode, trackReq.DurationMs, trackReq.Level)
		if validationErr != nil {
			itemLog.Warn("Pre-validation failed for batch item", "error", validationErr)
			resultItem.Error = validationErr.Error()
			preCheckFailed = true
		} else {
			exists, checkErr := uc.storageService.ObjectExists(ctx, uc.minioBucket, trackReq.ObjectKey)
			if checkErr != nil {
				itemLog.Error("Failed to check object existence pre-transaction", "error", checkErr)
				resultItem.Error = "failed to verify upload status"
				preCheckFailed = true
			} else if !exists {
				itemLog.Warn("Object not found in storage pre-transaction")
				resultItem.Error = "uploaded file not found in storage"
				preCheckFailed = true
			} else {
				validatedItems = append(validatedItems, trackReq)
				resultItem.Success = true
			}
		}
		tempResults[i] = resultItem
	}

	if preCheckFailed {
		log.Warn("Batch completion aborted due to pre-transaction validation/existence check failures")
		return tempResults, fmt.Errorf("%w: one or more items failed validation or were not found in storage", domain.ErrInvalidArgument)
	}

	log.Info("All items passed pre-checks, proceeding with database transaction.")
	finalDbResults := make(map[string]*port.BatchCompleteResultItem)
	for i := range tempResults {
		finalDbResults[tempResults[i].ObjectKey] = &tempResults[i]
	}

	txErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		var firstDbErr error
		for _, trackReq := range validatedItems {
			itemLog := log.With("objectKey", trackReq.ObjectKey, "title", trackReq.Title)
			resultItemPtr := finalDbResults[trackReq.ObjectKey]

			// CHANGED: Pass trackReq (port.BatchCompleteItem)
			track, domainErr := uc.createDomainTrack(txCtx, userID, trackReq)
			if domainErr != nil {
				itemLog.Error("Failed to create domain object for batch item (unexpected)", "error", domainErr)
				resultItemPtr.Success = false
				resultItemPtr.Error = "failed to process track data"
				if firstDbErr == nil {
					firstDbErr = fmt.Errorf("item %s failed: %w", trackReq.ObjectKey, domainErr)
				}
				continue
			}

			dbErr := uc.trackRepo.Create(txCtx, track)
			if dbErr != nil {
				itemLog.Error("Failed to create track record for batch item", "error", dbErr, "trackID", track.ID)
				resultItemPtr.Success = false
				resultItemPtr.Error = "failed to save track information"
				if errors.Is(dbErr, domain.ErrConflict) {
					resultItemPtr.Error = "track identifier conflict"
					log.Warn("Conflict during batch track creation, potentially duplicate object key", "objectKey", trackReq.ObjectKey)
				}
				if firstDbErr == nil {
					firstDbErr = fmt.Errorf("item %s failed: %w", trackReq.ObjectKey, dbErr)
				}
			} else {
				resultItemPtr.Success = true
				resultItemPtr.TrackID = track.ID.String()
				resultItemPtr.Error = ""
				itemLog.Info("Batch item processed and track created successfully in transaction", "trackID", track.ID)
			}
		}
		return firstDbErr
	})

	finalResults := make([]port.BatchCompleteResultItem, len(input.Tracks))
	for i := range tempResults {
		finalResults[i] = tempResults[i]
		if txErr != nil && finalDbResults[finalResults[i].ObjectKey] != nil && finalDbResults[finalResults[i].ObjectKey].Success {
			finalResults[i].Success = false
			if finalResults[i].Error == "" {
				finalResults[i].Error = "database transaction failed"
			}
		}
	}

	if txErr != nil {
		log.Error("Batch completion failed and transaction rolled back", "error", txErr)
		processingErr = fmt.Errorf("batch processing failed: %w", txErr)
	} else {
		log.Info("Batch completion finished and transaction committed")
	}

	return finalResults, processingErr
}

// --- Helper Methods ---

func (uc *UploadUseCase) validateContentType(contentType string) error {
	if contentType == "" {
		return fmt.Errorf("%w: contentType cannot be empty", domain.ErrInvalidArgument)
	}
	// Add more specific validation if needed (e.g., allow only audio/*)
	return nil
}

func (uc *UploadUseCase) generateObjectKey(userID domain.UserID, filename string) string {
	extension := filepath.Ext(filename)
	randomUUID := uuid.NewString()
	// Ensure consistent path separator
	return fmt.Sprintf("user-uploads/%s/%s%s", userID.String(), randomUUID, extension)
}

func (uc *UploadUseCase) validateCompleteUploadRequest(ctx context.Context, userID domain.UserID, objectKey, title, langCode string, durationMs int64, level string) error {
	log := uc.logger.With("userID", userID.String(), "objectKey", objectKey)
	if objectKey == "" {
		return fmt.Errorf("%w: objectKey is required", domain.ErrInvalidArgument)
	}
	if title == "" {
		return fmt.Errorf("%w: title is required", domain.ErrInvalidArgument)
	}
	if langCode == "" {
		return fmt.Errorf("%w: languageCode is required", domain.ErrInvalidArgument)
	}
	if durationMs <= 0 {
		return fmt.Errorf("%w: valid durationMs is required", domain.ErrInvalidArgument)
	}
	expectedPrefix := fmt.Sprintf("user-uploads/%s/", userID.String())
	if !strings.HasPrefix(objectKey, expectedPrefix) {
		log.Warn("Attempt to complete upload for object key not belonging to user", "expectedPrefix", expectedPrefix)
		return fmt.Errorf("%w: invalid object key provided", domain.ErrPermissionDenied)
	}
	levelVO := domain.AudioLevel(level)
	if level != "" && !levelVO.IsValid() {
		return fmt.Errorf("%w: invalid audio level '%s'", domain.ErrInvalidArgument, level)
	}
	_, err := domain.NewLanguage(langCode, "")
	if err != nil {
		return err
	}
	return nil
}

// createDomainTrack creates an AudioTrack domain object from the request data.
func (uc *UploadUseCase) createDomainTrack(ctx context.Context, userID domain.UserID, reqData interface{}) (*domain.AudioTrack, error) {
	var title, description, objectKey, langCode, levelStr string
	var durationMs int64
	var isPublic bool
	var tags []string
	var coverURL *string

	switch r := reqData.(type) {
	case port.CompleteUploadInput: // CHANGED: Use Input type
		title = r.Title
		description = r.Description
		objectKey = r.ObjectKey
		langCode = r.LanguageCode
		levelStr = r.Level
		durationMs = r.DurationMs
		isPublic = r.IsPublic
		tags = r.Tags
		coverURL = r.CoverImageURL
	case port.BatchCompleteItem:
		title = r.Title
		description = r.Description
		objectKey = r.ObjectKey
		langCode = r.LanguageCode
		levelStr = r.Level
		durationMs = r.DurationMs
		isPublic = r.IsPublic
		tags = r.Tags
		coverURL = r.CoverImageURL
	default:
		return nil, fmt.Errorf("internal error: unsupported type for createDomainTrack: %T", reqData)
	}

	langVO, err := domain.NewLanguage(langCode, "")
	if err != nil {
		return nil, err
	}
	levelVO := domain.AudioLevel(levelStr)
	duration := time.Duration(durationMs) * time.Millisecond
	uploaderID := userID

	track, err := domain.NewAudioTrack(title, description, uc.minioBucket, objectKey, langVO, levelVO, duration, &uploaderID, isPublic, tags, coverURL)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create AudioTrack domain object", "error", err)
		return nil, fmt.Errorf("failed to process track data: %w", err)
	}
	return track, nil
}

var _ port.UploadUseCase = (*UploadUseCase)(nil)
```

## `internal/usecase/user_activity_uc.go`

```go
// ============================================
// FILE: internal/usecase/user_activity_uc.go (MODIFIED)
// ============================================
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

type UserActivityUseCase struct {
	progressRepo port.PlaybackProgressRepository
	bookmarkRepo port.BookmarkRepository
	trackRepo    port.AudioTrackRepository
	logger       *slog.Logger
}

func NewUserActivityUseCase(
	pr port.PlaybackProgressRepository,
	br port.BookmarkRepository,
	tr port.AudioTrackRepository,
	log *slog.Logger,
) *UserActivityUseCase {
	return &UserActivityUseCase{
		progressRepo: pr,
		bookmarkRepo: br,
		trackRepo:    tr,
		logger:       log.With("usecase", "UserActivityUseCase"),
	}
}

// --- Playback Progress Use Cases ---

func (uc *UserActivityUseCase) RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error {
	exists, err := uc.trackRepo.Exists(ctx, trackID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to check track existence for progress update", "error", err, "trackID", trackID, "userID", userID)
		return fmt.Errorf("failed to validate track: %w", err)
	}
	if !exists {
		uc.logger.WarnContext(ctx, "Attempt to record progress for non-existent track", "trackID", trackID, "userID", userID)
		return domain.ErrNotFound
	}
	prog, err := domain.NewOrUpdatePlaybackProgress(userID, trackID, progress)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid progress value provided", "error", err, "userID", userID, "trackID", trackID, "progress", progress)
		return err
	}
	if err := uc.progressRepo.Upsert(ctx, prog); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to upsert playback progress", "error", err, "userID", userID, "trackID", trackID)
		return fmt.Errorf("failed to save progress: %w", err)
	}
	uc.logger.InfoContext(ctx, "Playback progress recorded", "userID", userID, "trackID", trackID, "progressMs", progress.Milliseconds())
	return nil
}

func (uc *UserActivityUseCase) GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	progress, err := uc.progressRepo.Find(ctx, userID, trackID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to get playback progress", "error", err, "userID", userID, "trackID", trackID)
		}
		return nil, err
	}
	return progress, nil
}

// ListUserProgress retrieves a paginated list of all progress records for a user.
// CHANGED: Parameter type to port.ListProgressInput
func (uc *UserActivityUseCase) ListUserProgress(ctx context.Context, input port.ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error) {
	// CHANGED: Use input.Page
	pageParams := pagination.NewPageFromOffset(input.Page.Limit, input.Page.Offset)
	// CHANGED: Use input.UserID
	progressList, total, err := uc.progressRepo.ListByUser(ctx, input.UserID, pageParams)
	if err != nil {
		// CHANGED: Use input.UserID
		uc.logger.ErrorContext(ctx, "Failed to list user progress", "error", err, "userID", input.UserID, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve progress list: %w", err)
	}
	return progressList, total, pageParams, nil
}

// --- Bookmark Use Cases ---

func (uc *UserActivityUseCase) CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error) {
	exists, err := uc.trackRepo.Exists(ctx, trackID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to check track existence for bookmark creation", "error", err, "trackID", trackID, "userID", userID)
		return nil, fmt.Errorf("failed to validate track: %w", err)
	}
	if !exists {
		uc.logger.WarnContext(ctx, "Attempt to create bookmark for non-existent track", "trackID", trackID, "userID", userID)
		return nil, fmt.Errorf("%w: track not found", domain.ErrNotFound)
	}
	bookmark, err := domain.NewBookmark(userID, trackID, timestamp, note)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid bookmark timestamp provided", "error", err, "userID", userID, "trackID", trackID, "timestampMs", timestamp.Milliseconds())
		return nil, err
	}
	if err := uc.bookmarkRepo.Create(ctx, bookmark); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save bookmark", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}
	uc.logger.InfoContext(ctx, "Bookmark created", "bookmarkID", bookmark.ID, "userID", userID, "trackID", trackID)
	return bookmark, nil
}

// ListBookmarks retrieves bookmarks for a user, optionally filtered by track.
// CHANGED: Parameter type to port.ListBookmarksInput
func (uc *UserActivityUseCase) ListBookmarks(ctx context.Context, input port.ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error) {
	var bookmarks []*domain.Bookmark
	var total int
	var err error
	// CHANGED: Use input.Page
	pageParams := pagination.NewPageFromOffset(input.Page.Limit, input.Page.Offset)

	// CHANGED: Use input.TrackIDFilter and input.UserID
	if input.TrackIDFilter != nil {
		bookmarks, err = uc.bookmarkRepo.ListByUserAndTrack(ctx, input.UserID, *input.TrackIDFilter)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user and track", "error", err, "userID", input.UserID, "trackID", *input.TrackIDFilter)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks for track: %w", err)
		}
		total = len(bookmarks)
		pageParams = pagination.Page{Limit: total, Offset: 0}
		if total == 0 {
			pageParams.Limit = pagination.DefaultLimit
		}
	} else {
		// CHANGED: Use input.UserID
		bookmarks, total, err = uc.bookmarkRepo.ListByUser(ctx, input.UserID, pageParams)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user", "error", err, "userID", input.UserID, "page", pageParams)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks: %w", err)
		}
	}
	return bookmarks, total, pageParams, nil
}

func (uc *UserActivityUseCase) DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error {
	bookmark, err := uc.bookmarkRepo.FindByID(ctx, bookmarkID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to find bookmark for deletion", "error", err, "bookmarkID", bookmarkID, "userID", userID)
		}
		return err
	}
	if bookmark.UserID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete bookmark not owned by user", "bookmarkID", bookmarkID, "ownerID", bookmark.UserID, "userID", userID)
		return domain.ErrPermissionDenied
	}
	// CHANGED: Pass only bookmarkID to Delete, ownership check is done above
	if err := uc.bookmarkRepo.Delete(ctx, bookmarkID); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to delete bookmark from repository", "error", err, "bookmarkID", bookmarkID, "userID", userID)
		}
		return err // Return NotFound if repo returns it (double deletion attempt)
	}
	uc.logger.InfoContext(ctx, "Bookmark deleted", "bookmarkID", bookmarkID, "userID", userID)
	return nil
}

var _ port.UserActivityUseCase = (*UserActivityUseCase)(nil)
```

## `internal/domain/user.go`

```go
// internal/domain/user.go
package domain

import (
	"time"
	"fmt"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserID is the unique identifier for a User.
type UserID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.New())
}

func UserIDFromString(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("invalid UserID format: %w", err)
	}
	return UserID(id), nil
}

func (uid UserID) String() string {
	return uuid.UUID(uid).String()
}

// AuthProvider represents the method used for authentication.
type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGoogle AuthProvider = "google"
	// Add other providers like Facebook, Apple etc. here
)

// User represents a user in the system.
type User struct {
	ID              UserID
	Email           Email // Using validated Email value object
	Name            string
	HashedPassword  *string // Pointer allows null for external auth users
	GoogleID        *string // Unique ID from Google (subject claim)
	AuthProvider    AuthProvider
	ProfileImageURL *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewLocalUser creates a new user who registered with email and password.
func NewLocalUser(emailAddr, name, hashedPassword string) (*User, error) {
	emailVO, err := NewEmail(emailAddr)
	if err != nil {
		return nil, err
	}
	if hashedPassword == "" {
		return nil, fmt.Errorf("%w: password hash cannot be empty for local user", ErrInvalidArgument)
	}

	now := time.Now()
	return &User{
		ID:             NewUserID(),
		Email:          emailVO,
		Name:           name,
		HashedPassword: &hashedPassword, // Store the already hashed password
		GoogleID:       nil,
		AuthProvider:   AuthProviderLocal,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// NewGoogleUser creates a new user authenticated via Google.
func NewGoogleUser(emailAddr, name, googleID string, profileImageURL *string) (*User, error) {
	emailVO, err := NewEmail(emailAddr)
	if err != nil {
		return nil, err
	}
	if googleID == "" {
		return nil, fmt.Errorf("%w: google ID cannot be empty for google user", ErrInvalidArgument)
	}

	now := time.Now()
	return &User{
		ID:              NewUserID(),
		Email:           emailVO,
		Name:            name,
		HashedPassword:  nil, // No password for Google users initially
		GoogleID:        &googleID,
		AuthProvider:    AuthProviderGoogle,
		ProfileImageURL: profileImageURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// ValidatePassword checks if the provided plain password matches the user's hashed password.
// Returns true if the password matches or if the user is not a local user (no password set).
// Returns false if the password does not match or an error occurred during comparison.
func (u *User) ValidatePassword(plainPassword string) (bool, error) {
	if u.AuthProvider != AuthProviderLocal || u.HashedPassword == nil {
		// Cannot validate password for non-local users or if hash is missing
		// Consider returning an error or specific bool value based on desired behavior.
		// Returning false might be misleading if the intent is "password auth not applicable".
		// Let's return an error for clarity.
		return false, fmt.Errorf("password validation not applicable for provider %s", u.AuthProvider)
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.HashedPassword), []byte(plainPassword))
	if err == nil {
		return true, nil // Passwords match
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil // Passwords don't match
	}
	// Some other error occurred (e.g., invalid hash format)
	return false, fmt.Errorf("error comparing password hash: %w", err)
}

// UpdateProfile updates mutable profile fields.
func (u *User) UpdateProfile(name string, profileImageURL *string) {
	u.Name = name
	u.ProfileImageURL = profileImageURL
	u.UpdatedAt = time.Now()
}

// LinkGoogleID links a Google account ID to an existing local user.
// Use this cautiously, ensure proper verification beforehand in usecase.
func (u *User) LinkGoogleID(googleID string) error {
	if u.GoogleID != nil && *u.GoogleID != "" {
		return fmt.Errorf("%w: user already linked to a google account", ErrConflict)
	}
	if googleID == "" {
		return fmt.Errorf("%w: google ID cannot be empty", ErrInvalidArgument)
	}
	u.GoogleID = &googleID
	// Potentially change auth provider or keep it? Depends on login flow.
	// If user can now login via EITHER method, maybe keep 'local' or add a list?
	// For simplicity, let's assume linking doesn't change the primary AuthProvider if it was 'local'.
	// If the user was created via Google first, AuthProvider would already be 'google'.
	u.UpdatedAt = time.Now()
	return nil
}
```

## `internal/domain/user_test.go`

```go
package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNewLocalUser(t *testing.T) {
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	validHash := string(hashedPwd)

	tests := []struct {
		name           string
		email          string
		userName       string
		hashedPassword string
		wantErr        bool
		errType        error // Expected error type (e.g., ErrInvalidArgument)
	}{
		{"Valid local user", "local@example.com", "Local User", validHash, false, nil},
		{"Invalid email", "local@", "Local User", validHash, true, ErrInvalidArgument},
		{"Empty name", "local@example.com", "", validHash, false, nil}, // Name can be empty
		{"Empty hash", "local@example.com", "Local User", "", true, ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewLocalUser(tt.email, tt.userName, tt.hashedPassword)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, UserID{}, got.ID)
				assert.Equal(t, tt.email, got.Email.String())
				assert.Equal(t, tt.userName, got.Name)
				assert.NotNil(t, got.HashedPassword)
				assert.Equal(t, tt.hashedPassword, *got.HashedPassword)
				assert.Nil(t, got.GoogleID)
				assert.Equal(t, AuthProviderLocal, got.AuthProvider)
				assert.Nil(t, got.ProfileImageURL)
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}

func TestNewGoogleUser(t *testing.T) {
	profileURL := "http://example.com/pic.jpg"

	tests := []struct {
		name       string
		email      string
		userName   string
		googleID   string
		profileURL *string
		wantErr    bool
		errType    error
	}{
		{"Valid google user", "google@example.com", "Google User", "google123", &profileURL, false, nil},
		{"Valid google user no pic", "google2@example.com", "Google User 2", "google456", nil, false, nil},
		{"Invalid email", "google@", "Google User", "google123", nil, true, ErrInvalidArgument},
		{"Empty name", "google@example.com", "", "google123", nil, false, nil}, // Name can be empty
		{"Empty google ID", "google@example.com", "Google User", "", nil, true, ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewGoogleUser(tt.email, tt.userName, tt.googleID, tt.profileURL)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, UserID{}, got.ID)
				assert.Equal(t, tt.email, got.Email.String())
				assert.Equal(t, tt.userName, got.Name)
				assert.Nil(t, got.HashedPassword)
				assert.NotNil(t, got.GoogleID)
				assert.Equal(t, tt.googleID, *got.GoogleID)
				assert.Equal(t, AuthProviderGoogle, got.AuthProvider)
				if tt.profileURL != nil {
					assert.NotNil(t, got.ProfileImageURL)
					assert.Equal(t, *tt.profileURL, *got.ProfileImageURL)
				} else {
					assert.Nil(t, got.ProfileImageURL)
				}
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	plainPassword := "password123"
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	validHash := string(hashedPwd)

	localUser, _ := NewLocalUser("local@example.com", "Test", validHash)
	googleUser, _ := NewGoogleUser("google@example.com", "Test", "gid1", nil)
	localUserNoHash, _ := NewLocalUser("localnohash@example.com", "Test", "dummy")
	localUserNoHash.HashedPassword = nil // Simulate case where hash is somehow nil

	tests := []struct {
		name        string
		user        *User
		password    string
		wantMatch   bool
		wantErr     bool
		errContains string
	}{
		{"Correct password local user", localUser, plainPassword, true, false, ""},
		{"Incorrect password local user", localUser, "wrongpassword", false, false, ""},
		{"Google user", googleUser, plainPassword, false, true, "password validation not applicable"},
		{"Local user with nil hash", localUserNoHash, plainPassword, false, true, "password validation not applicable"},
		{"Invalid hash format check", localUser, plainPassword, false, true, "error comparing password hash"}, // Need to provide an invalid hash to trigger this
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.user
			password := tt.password
			hash := ""
			if user.HashedPassword != nil {
				hash = *user.HashedPassword
			}

			// Special case for invalid hash format test
			if tt.name == "Invalid hash format check" {
				hash = "this-is-not-a-bcrypt-hash"
				// Recreate user with this invalid hash temporarily for the test
				user = &User{
					ID: localUser.ID, Email: localUser.Email, Name: localUser.Name, AuthProvider: localUser.AuthProvider,
					HashedPassword: &hash, CreatedAt: localUser.CreatedAt, UpdatedAt: localUser.UpdatedAt,
				}
			}

			gotMatch, err := user.ValidatePassword(password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMatch, gotMatch)
			}
		})
	}
}

// Add tests for UpdateProfile, LinkGoogleID if needed
```

## `internal/domain/value_objects_test.go`

```go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLanguage(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		langName string
		wantCode string
		wantErr  bool
	}{
		{"Valid language", "en-US", "English (US)", "EN-US", false},
		{"Valid language lowercase", "fr-ca", "French (Canada)", "FR-CA", false},
		{"Empty code", "", "No Language", "", true},
		{"Only name", "en", "", "EN", false}, // Allow empty name
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLanguage(tt.code, tt.langName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidArgument) // Check specific error type
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCode, got.Code())
				assert.Equal(t, tt.langName, got.Name())
			}
		})
	}
}

func TestAudioLevel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		level AudioLevel
		want  bool
	}{
		{"Valid A1", LevelA1, true},
		{"Valid B2", LevelB2, true},
		{"Valid Native", LevelNative, true},
		{"Valid Unknown", LevelUnknown, true},
		{"Invalid Lowercase", AudioLevel("a1"), false},
		{"Invalid Other", AudioLevel("XYZ"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.IsValid())
			assert.Equal(t, string(tt.level), tt.level.String()) // Test String() method
		})
	}
}

func TestCollectionType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		typ  CollectionType
		want bool
	}{
		{"Valid Course", TypeCourse, true},
		{"Valid Playlist", TypePlaylist, true},
		{"Valid Unknown", TypeUnknown, true},
		{"Invalid Lowercase", CollectionType("course"), false},
		{"Invalid Other", CollectionType("XYZ"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.typ.IsValid())
			assert.Equal(t, string(tt.typ), tt.typ.String()) // Test String() method
		})
	}
}

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
		wantErr bool
	}{
		{"Valid email", "test@example.com", "test@example.com", false},
		{"Valid email with name", "Test User <test.user+alias@example.co.uk>", "test.user+alias@example.co.uk", false},
		{"Invalid format", "test@", "", true},
		{"Empty string", "", "", true},
		{"Only domain", "@example.com", "", true},
		{"Accept No TLD (mail.ParseAddress allows)", "test@example", "test@example", false}, // mail.ParseAddress requires valid TLD usually
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidArgument)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.String())
			}
		})
	}
}
```

## `internal/domain/playbackprogress.go`

```go
// internal/domain/playbackprogress.go
package domain

import (
	"fmt"
	"time"
)

// PlaybackProgress represents a user's listening progress on a specific audio track.
// UserID and TrackID together form the primary key.
type PlaybackProgress struct {
	UserID         UserID
	TrackID        TrackID
	Progress       time.Duration // Current listening position
	LastListenedAt time.Time     // When the progress was last updated
}

// NewOrUpdatePlaybackProgress creates or updates a playback progress record.
func NewOrUpdatePlaybackProgress(userID UserID, trackID TrackID, progress time.Duration) (*PlaybackProgress, error) {
	if progress < 0 {
		return nil, fmt.Errorf("%w: progress duration cannot be negative", ErrInvalidArgument)
	}
	// UserID and TrackID validity assumed to be checked elsewhere (e.g., foreign keys or usecase)

	return &PlaybackProgress{
		UserID:         userID,
		TrackID:        trackID,
		Progress:       progress,
		LastListenedAt: time.Now(),
	}, nil
}
```

## `internal/domain/audiotrack.go`

```go
// internal/domain/audiotrack.go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TrackID is the unique identifier for an AudioTrack.
type TrackID uuid.UUID

func NewTrackID() TrackID {
	return TrackID(uuid.New())
}

func TrackIDFromString(s string) (TrackID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TrackID{}, fmt.Errorf("invalid TrackID format: %w", err)
	}
	return TrackID(id), nil
}

func (tid TrackID) String() string {
	return uuid.UUID(tid).String()
}

// AudioTrack represents a single audio file with metadata.
type AudioTrack struct {
	ID              TrackID
	Title           string
	Description     string
	Language        Language     // CORRECTED TYPE: Use Language value object
	Level           AudioLevel   // CORRECTED TYPE: Use AudioLevel value object
	Duration        time.Duration // Store as duration for easier use
	MinioBucket     string
	MinioObjectKey  string
	CoverImageURL   *string
	UploaderID      *UserID // Optional link to the user who uploaded it
	IsPublic        bool
	Tags            []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// TranscriptionID *TranscriptionID // Optional if transcriptions are separate entities
}

// NewAudioTrack creates a new audio track instance.
func NewAudioTrack(
	title, description, bucket, objectKey string,
	lang Language, // CORRECTED TYPE: Accept Language VO
	level AudioLevel, // CORRECTED TYPE: Accept AudioLevel VO
	duration time.Duration,
	uploaderID *UserID, isPublic bool, tags []string, coverURL *string,
) (*AudioTrack, error) {
	if title == "" {
		return nil, fmt.Errorf("%w: track title cannot be empty", ErrInvalidArgument)
	}
	if bucket == "" || objectKey == "" {
		return nil, fmt.Errorf("%w: minio bucket and key cannot be empty", ErrInvalidArgument)
	}
	if duration < 0 {
		return nil, fmt.Errorf("%w: duration cannot be negative", ErrInvalidArgument)
	}
	// Allow LevelUnknown, but validate others
	if level != LevelUnknown && !level.IsValid() {
		return nil, fmt.Errorf("%w: invalid audio level '%s'", ErrInvalidArgument, level)
	}
	// Language VO creation handles its own validation if called externally before this

	now := time.Now()
	return &AudioTrack{
		ID:             NewTrackID(),
		Title:          title,
		Description:    description,
		Language:       lang, // Assign VO directly
		Level:          level, // Assign VO directly
		Duration:       duration,
		MinioBucket:    bucket,
		MinioObjectKey: objectKey,
		UploaderID:     uploaderID,
		IsPublic:       isPublic,
		Tags:           tags,
		CoverImageURL:  coverURL,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
```

## `internal/domain/value_objects.go`

```go
// internal/domain/value_objects.go
package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

// --- Basic Value Objects ---

// Language represents a language with its code and name. Immutable.
type Language struct {
	code string // e.g., "en-US"
	name string // e.g., "English (US)"
}

func NewLanguage(code, name string) (Language, error) {
	if code == "" {
		return Language{}, fmt.Errorf("%w: language code cannot be empty", ErrInvalidArgument)
	}
	// Add more validation if needed (e.g., format check)
	return Language{code: strings.ToUpper(code), name: name}, nil
}
func (l Language) Code() string { return l.code }
func (l Language) Name() string { return l.name }

// AudioLevel represents the difficulty level of an audio track. Immutable.
type AudioLevel string // e.g., "A1", "B2", "NATIVE"

const (
	LevelA1 AudioLevel = "A1"
	LevelA2 AudioLevel = "A2"
	LevelB1 AudioLevel = "B1"
	LevelB2 AudioLevel = "B2"
	LevelC1 AudioLevel = "C1"
	LevelC2 AudioLevel = "C2"
	LevelNative AudioLevel = "NATIVE"
	LevelUnknown AudioLevel = "" // Or "UNKNOWN"
)

func (l AudioLevel) IsValid() bool {
	switch l {
	case LevelA1, LevelA2, LevelB1, LevelB2, LevelC1, LevelC2, LevelNative, LevelUnknown:
		return true
	default:
		return false
	}
}
func (l AudioLevel) String() string { return string(l) }

// CollectionType represents the type of an audio collection. Immutable.
type CollectionType string // e.g., "COURSE", "PLAYLIST"

const (
	TypeCourse   CollectionType = "COURSE"
	TypePlaylist CollectionType = "PLAYLIST"
	TypeUnknown  CollectionType = "" // Or "UNKNOWN"
)

func (t CollectionType) IsValid() bool {
	switch t {
	case TypeCourse, TypePlaylist, TypeUnknown:
		return true
	default:
		return false
	}
}
func (t CollectionType) String() string { return string(t) }

// --- Email Value Object (Example with Validation) ---

// Email represents a validated email address. Immutable.
type Email struct {
	address string
}

func NewEmail(address string) (Email, error) {
	if address == "" {
		return Email{}, fmt.Errorf("%w: email address cannot be empty", ErrInvalidArgument)
	}
	parsed, err := mail.ParseAddress(address)
	if err != nil {
		return Email{}, fmt.Errorf("%w: invalid email address format: %v", ErrInvalidArgument, err)
	}
	// Potentially add more domain-specific validation (e.g., allowed domains)
	return Email{address: parsed.Address}, nil // Use the parsed address for canonical form
}

func (e Email) String() string { return e.address }

// --- Time-related Value Objects (Go's time types often suffice) ---
// time.Duration can represent playback progress or track duration.
// time.Time can represent timestamps like CreatedAt, UpdatedAt, LastListenedAt.
```

## `internal/domain/audiocollection_test.go`

```go
package domain

import (
	"testing"
	"time"

	// 确保导入 slices 包
	"github.com/stretchr/testify/assert"
)

func TestNewCollectionID(t *testing.T) {
	id1 := NewCollectionID()
	id2 := NewCollectionID()
	assert.NotEqual(t, CollectionID{}, id1) // Not zero value
	assert.NotEqual(t, id1, id2)            // IDs should be unique
}

func TestCollectionIDFromString(t *testing.T) {
	validUUID := "a4a5b418-2150-4d1b-9c0a-4b8f8e7a8e21"
	invalidUUID := "not-a-uuid"

	id, err := CollectionIDFromString(validUUID)
	assert.NoError(t, err)
	assert.Equal(t, validUUID, id.String())

	_, err = CollectionIDFromString(invalidUUID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CollectionID format")
}

func TestNewCollection(t *testing.T) {
	ownerID := NewUserID()

	tests := []struct {
		name        string
		title       string
		description string
		ownerID     UserID
		colType     CollectionType
		wantErr     bool
		errType     error
	}{
		{"Valid Course", "Intro Course", "Description", ownerID, TypeCourse, false, nil},
		{"Valid Playlist", "My Playlist", "", ownerID, TypePlaylist, false, nil}, // Empty description allowed
		{"Empty Title", "", "Description", ownerID, TypeCourse, true, ErrInvalidArgument},
		{"Invalid Type", "Title", "Desc", ownerID, CollectionType("INVALID"), true, ErrInvalidArgument},
		{"Unknown Type", "Title", "Desc", ownerID, TypeUnknown, true, ErrInvalidArgument}, // Unknown type not allowed by constructor
		// OwnerID validation usually done by DB FK constraint, not here
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewAudioCollection(tt.title, tt.description, tt.ownerID, tt.colType)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, CollectionID{}, got.ID)
				assert.Equal(t, tt.title, got.Title)
				assert.Equal(t, tt.description, got.Description)
				assert.Equal(t, tt.ownerID, got.OwnerID)
				assert.Equal(t, tt.colType, got.Type)
				assert.NotNil(t, got.TrackIDs) // Should be initialized
				assert.Empty(t, got.TrackIDs)  // Should be empty initially
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}

func TestAudioCollection_TrackManagement(t *testing.T) {
	ownerID := NewUserID()
	collection, _ := NewAudioCollection("Test Collection", "", ownerID, TypePlaylist)
	track1 := NewTrackID()
	track2 := NewTrackID()
	track3 := NewTrackID()
	initialTime := collection.UpdatedAt

	// --- AddTrack ---
	t.Run("AddTrack", func(t *testing.T) {
		// Add first track
		err := collection.AddTrack(track1, 0)
		assert.NoError(t, err)
		assert.Equal(t, []TrackID{track1}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(initialTime))
		timeAfterAdd1 := collection.UpdatedAt

		// Add second track at beginning
		err = collection.AddTrack(track2, 0)
		assert.NoError(t, err)
		assert.Equal(t, []TrackID{track2, track1}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterAdd1))
		timeAfterAdd2 := collection.UpdatedAt

		// Add third track at end (position out of bounds)
		err = collection.AddTrack(track3, 10) // Position 10 is > len(2)
		assert.NoError(t, err)
		assert.Equal(t, []TrackID{track2, track1, track3}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterAdd2))
		timeAfterAdd3 := collection.UpdatedAt

		// Add existing track (should error)
		err = collection.AddTrack(track1, 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrConflict)
		assert.Equal(t, []TrackID{track2, track1, track3}, collection.TrackIDs) // No change
		assert.Equal(t, timeAfterAdd3, collection.UpdatedAt)                    // Timestamp unchanged
	})

	// Reset state for next tests if needed, or use subtests carefully
	collection.TrackIDs = []TrackID{track2, track1, track3} // Reset state
	timeBeforeRemove := collection.UpdatedAt

	// --- RemoveTrack ---
	t.Run("RemoveTrack", func(t *testing.T) {
		// Remove middle track
		removed := collection.RemoveTrack(track1)
		assert.True(t, removed)
		assert.Equal(t, []TrackID{track2, track3}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeBeforeRemove))
		timeAfterRemove1 := collection.UpdatedAt

		// Remove non-existent track
		removed = collection.RemoveTrack(NewTrackID()) // New random ID
		assert.False(t, removed)
		assert.Equal(t, []TrackID{track2, track3}, collection.TrackIDs) // No change
		assert.Equal(t, timeAfterRemove1, collection.UpdatedAt)         // Timestamp unchanged

		// Remove last track
		removed = collection.RemoveTrack(track3)
		assert.True(t, removed)
		assert.Equal(t, []TrackID{track2}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterRemove1))
		timeAfterRemove2 := collection.UpdatedAt

		// Remove first track (last remaining)
		removed = collection.RemoveTrack(track2)
		assert.True(t, removed)
		assert.Empty(t, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterRemove2))
	})

	// --- ReorderTracks ---
	t.Run("ReorderTracks", func(t *testing.T) {
		// Reset state
		collection.TrackIDs = []TrackID{track1, track2, track3}
		timeBeforeReorder := time.Now()
		collection.UpdatedAt = timeBeforeReorder // Set known time

		// Valid reorder
		newOrder := []TrackID{track3, track1, track2}
		err := collection.ReorderTracks(newOrder)
		assert.NoError(t, err)
		assert.Equal(t, newOrder, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeBeforeReorder))
		timeAfterReorder1 := collection.UpdatedAt

		// Error: Incorrect number of tracks
		wrongCountOrder := []TrackID{track1, track3}
		err = collection.ReorderTracks(wrongCountOrder)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "number of provided track IDs")
		assert.Equal(t, newOrder, collection.TrackIDs)           // Order unchanged
		assert.Equal(t, timeAfterReorder1, collection.UpdatedAt) // Timestamp unchanged

		// Error: Track not originally present
		track4 := NewTrackID()
		notPresentOrder := []TrackID{track1, track2, track4}
		err = collection.ReorderTracks(notPresentOrder)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "is not part of the original collection")
		assert.Equal(t, newOrder, collection.TrackIDs)           // Order unchanged
		assert.Equal(t, timeAfterReorder1, collection.UpdatedAt) // Timestamp unchanged

		// Error: Duplicate track in new order
		duplicateOrder := []TrackID{track1, track1, track2}
		err = collection.ReorderTracks(duplicateOrder)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "appears multiple times")
		assert.Equal(t, newOrder, collection.TrackIDs)           // Order unchanged
		assert.Equal(t, timeAfterReorder1, collection.UpdatedAt) // Timestamp unchanged

		// Reorder to empty list (should fail if original list not empty)
		collection.TrackIDs = []TrackID{track1} // Set to non-empty
		err = collection.ReorderTracks([]TrackID{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "number of provided track IDs")

		// Reorder empty list to empty list (should succeed)
		collection.TrackIDs = []TrackID{} // Set to empty
		timeBeforeEmptyReorder := time.Now()
		collection.UpdatedAt = timeBeforeEmptyReorder
		err = collection.ReorderTracks([]TrackID{})
		assert.NoError(t, err)
		assert.Empty(t, collection.TrackIDs)
		// assert.True(t, collection.UpdatedAt.After(timeBeforeEmptyReorder)) // debatable if timestamp should change here
	})
}
```

## `internal/domain/bookmark_test.go`

```go
package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBookmarkID(t *testing.T) {
	id1 := NewBookmarkID()
	id2 := NewBookmarkID()
	assert.NotEqual(t, BookmarkID{}, id1)
	assert.NotEqual(t, id1, id2)
}

func TestBookmarkIDFromString(t *testing.T) {
	validUUID := "d4e5f418-3150-4d1b-9c0a-4b8f8e7a8e21"
	invalidUUID := "not-a-uuid"

	id, err := BookmarkIDFromString(validUUID)
	assert.NoError(t, err)
	assert.Equal(t, validUUID, id.String())

	_, err = BookmarkIDFromString(invalidUUID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid BookmarkID format")
}

func TestNewBookmark(t *testing.T) {
	userID := NewUserID()
	trackID := NewTrackID()

	tests := []struct {
		name      string
		userID    UserID
		trackID   TrackID
		timestamp time.Duration
		note      string
		wantErr   bool
		errType   error
	}{
		{"Valid bookmark", userID, trackID, 15 * time.Second, "Interesting point", false, nil},
		{"Valid bookmark zero timestamp", userID, trackID, 0, "Start", false, nil},
		{"Valid bookmark empty note", userID, trackID, 1 * time.Minute, "", false, nil},
		{"Invalid negative timestamp", userID, trackID, -5 * time.Second, "Should fail", true, ErrInvalidArgument},
		// UserID and TrackID validity assumed to be handled elsewhere
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewBookmark(tt.userID, tt.trackID, tt.timestamp, tt.note)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, BookmarkID{}, got.ID)
				assert.Equal(t, tt.userID, got.UserID)
				assert.Equal(t, tt.trackID, got.TrackID)
				assert.Equal(t, tt.timestamp, got.Timestamp)
				assert.Equal(t, tt.note, got.Note)
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}
```

## `internal/domain/bookmark.go`

```go
// internal/domain/bookmark.go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BookmarkID is the unique identifier for a Bookmark.
type BookmarkID uuid.UUID

func NewBookmarkID() BookmarkID {
	return BookmarkID(uuid.New())
}

func BookmarkIDFromString(s string) (BookmarkID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return BookmarkID{}, fmt.Errorf("invalid BookmarkID format: %w", err)
	}
	return BookmarkID(id), nil
}

func (bid BookmarkID) String() string {
	return uuid.UUID(bid).String()
}

// Bookmark represents a specific point in an audio track saved by a user.
type Bookmark struct {
	ID        BookmarkID
	UserID    UserID
	TrackID   TrackID
	Timestamp time.Duration // Position in the audio track where the bookmark is placed
	Note      string        // Optional user note about the bookmark
	CreatedAt time.Time
}

// NewBookmark creates a new bookmark instance.
func NewBookmark(userID UserID, trackID TrackID, timestamp time.Duration, note string) (*Bookmark, error) {
	if timestamp < 0 {
		return nil, fmt.Errorf("%w: bookmark timestamp cannot be negative", ErrInvalidArgument)
	}
	// UserID and TrackID validity assumed

	return &Bookmark{
		ID:        NewBookmarkID(),
		UserID:    userID,
		TrackID:   trackID,
		Timestamp: timestamp,
		Note:      note,
		CreatedAt: time.Now(),
	}, nil
}
```

## `internal/domain/playbackprogress_test.go`

```go
package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOrUpdatePlaybackProgress(t *testing.T) {
	userID := NewUserID()
	trackID := NewTrackID()

	tests := []struct {
		name     string
		userID   UserID
		trackID  TrackID
		progress time.Duration
		wantErr  bool
		errType  error
	}{
		{"Valid progress", userID, trackID, 30 * time.Second, false, nil},
		{"Zero progress", userID, trackID, 0, false, nil},
		{"Negative progress", userID, trackID, -10 * time.Millisecond, true, ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewOrUpdatePlaybackProgress(tt.userID, tt.trackID, tt.progress)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.userID, got.UserID)
				assert.Equal(t, tt.trackID, got.TrackID)
				assert.Equal(t, tt.progress, got.Progress)
				assert.WithinDuration(t, start, got.LastListenedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}
```

## `internal/domain/audiocollection.go`

```go
// internal/domain/audiocollection.go
package domain

import (
	"fmt"
	"time"
	"slices"

	"github.com/google/uuid"
)

// CollectionID is the unique identifier for an AudioCollection.
type CollectionID uuid.UUID

func NewCollectionID() CollectionID {
	return CollectionID(uuid.New())
}

func CollectionIDFromString(s string) (CollectionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return CollectionID{}, fmt.Errorf("invalid CollectionID format: %w", err)
	}
	return CollectionID(id), nil
}

func (cid CollectionID) String() string {
	return uuid.UUID(cid).String()
}


// AudioCollection represents a curated list of audio tracks (e.g., a course or playlist).
type AudioCollection struct {
	ID          CollectionID
	Title       string
	Description string
	OwnerID     UserID // The user who owns/created the collection
	Type        CollectionType // Value object (COURSE or PLAYLIST)
	TrackIDs    []TrackID // Ordered list of TrackIDs in the collection
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAudioCollection creates a new audio collection.
func NewAudioCollection(title, description string, ownerID UserID, colType CollectionType) (*AudioCollection, error) {
	if title == "" {
		return nil, fmt.Errorf("%w: collection title cannot be empty", ErrInvalidArgument)
	}
	if !colType.IsValid() || colType == TypeUnknown {
		return nil, fmt.Errorf("%w: invalid collection type '%s'", ErrInvalidArgument, colType)
	}
	// OwnerID validation happens implicitly via foreign key constraint usually

	now := time.Now()
	return &AudioCollection{
		ID:          NewCollectionID(),
		Title:       title,
		Description: description,
		OwnerID:     ownerID,
		Type:        colType,
		TrackIDs:    make([]TrackID, 0), // Initialize empty slice
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddTrack adds a track ID to the collection at a specific position.
// Position is 0-based. If position is out of bounds, it appends to the end.
func (c *AudioCollection) AddTrack(trackID TrackID, position int) error {
	// Check if track already exists
	for _, existingID := range c.TrackIDs {
		if existingID == trackID {
			return fmt.Errorf("%w: track %s already exists in collection %s", ErrConflict, trackID, c.ID)
		}
	}

	if position < 0 || position > len(c.TrackIDs) {
		position = len(c.TrackIDs) // Append to end if out of bounds
	}

	// Insert element at position
	c.TrackIDs = slices.Insert(c.TrackIDs, position, trackID)
	c.UpdatedAt = time.Now()
	return nil
}

// RemoveTrack removes a track ID from the collection.
func (c *AudioCollection) RemoveTrack(trackID TrackID) bool {
	initialLen := len(c.TrackIDs)
	c.TrackIDs = slices.DeleteFunc(c.TrackIDs, func(id TrackID) bool {
		return id == trackID
	})
	removed := len(c.TrackIDs) < initialLen
	if removed {
		c.UpdatedAt = time.Now()
	}
	return removed
}

// ReorderTracks sets the track order to the provided list of IDs.
// It ensures all provided IDs were already present in the collection.
func (c *AudioCollection) ReorderTracks(orderedTrackIDs []TrackID) error {
	if len(orderedTrackIDs) != len(c.TrackIDs) {
		return fmt.Errorf("%w: number of provided track IDs (%d) does not match current number (%d)", ErrInvalidArgument, len(orderedTrackIDs), len(c.TrackIDs))
	}

	// Check if all original tracks are present in the new order exactly once
	currentSet := make(map[TrackID]struct{}, len(c.TrackIDs))
	for _, id := range c.TrackIDs {
		currentSet[id] = struct{}{}
	}
	newSet := make(map[TrackID]struct{}, len(orderedTrackIDs))
	for _, id := range orderedTrackIDs {
		if _, exists := currentSet[id]; !exists {
			return fmt.Errorf("%w: track ID %s is not part of the original collection", ErrInvalidArgument, id)
		}
		if _, duplicate := newSet[id]; duplicate {
			return fmt.Errorf("%w: track ID %s appears multiple times in the new order", ErrInvalidArgument, id)
		}
		newSet[id] = struct{}{}
	}

	c.TrackIDs = orderedTrackIDs
	c.UpdatedAt = time.Now()
	return nil
}
```

## `internal/domain/audiotrack_test.go`

```go
package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAudioTrack(t *testing.T) {
	langEn, _ := NewLanguage("en-US", "English (US)")
	langFr, _ := NewLanguage("fr-FR", "French (France)")
	uploaderID := NewUserID()
	coverURL := "http://example.com/cover.jpg"

	tests := []struct {
		name       string
		title      string
		desc       string
		bucket     string
		objectKey  string
		lang       Language
		level      AudioLevel
		duration   time.Duration
		uploaderID *UserID
		isPublic   bool
		tags       []string
		coverURL   *string
		wantErr    bool
		errType    error
	}{
		{"Valid public track", "Track 1", "Desc 1", "bucket", "key1", langEn, LevelB1, 120 * time.Second, &uploaderID, true, []string{"news", "politics"}, &coverURL, false, nil},
		{"Valid private track no uploader", "Track 2", "", "bucket", "key2", langFr, LevelA2, 10 * time.Millisecond, nil, false, nil, nil, false, nil},
		{"Empty title", "", "Desc", "bucket", "key3", langEn, LevelC1, time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Empty bucket", "Track 3", "Desc", "", "key4", langEn, LevelA1, time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Empty object key", "Track 4", "Desc", "bucket", "", langEn, LevelA1, time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Negative duration", "Track 5", "Desc", "bucket", "key5", langEn, LevelNative, -time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Invalid Level", "Track 6", "Desc", "bucket", "key6", langEn, AudioLevel("XYZ"), time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Unknown Level", "Track 7", "Desc", "bucket", "key7", langEn, LevelUnknown, time.Second, nil, true, nil, nil, false, nil}, // Unknown level is allowed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewAudioTrack(tt.title, tt.desc, tt.bucket, tt.objectKey, tt.lang, tt.level, tt.duration, tt.uploaderID, tt.isPublic, tt.tags, tt.coverURL)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, TrackID{}, got.ID)
				assert.Equal(t, tt.title, got.Title)
				assert.Equal(t, tt.desc, got.Description)
				assert.Equal(t, tt.bucket, got.MinioBucket)
				assert.Equal(t, tt.objectKey, got.MinioObjectKey)
				assert.Equal(t, tt.lang, got.Language)
				assert.Equal(t, tt.level, got.Level)
				assert.Equal(t, tt.duration, got.Duration)
				assert.Equal(t, tt.uploaderID, got.UploaderID)
				assert.Equal(t, tt.isPublic, got.IsPublic)
				assert.Equal(t, tt.tags, got.Tags) // Check slice equality
				assert.Equal(t, tt.coverURL, got.CoverImageURL)
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}
```

## `internal/domain/errors.go`

```go
// internal/domain/errors.go
package domain

import "errors"

// Standard business logic errors
var (
	// ErrNotFound indicates that a requested entity could not be found.
	ErrNotFound = errors.New("entity not found")
	// ErrConflict indicates a conflict, e.g., resource already exists.
	ErrConflict = errors.New("resource conflict")
	// ErrInvalidArgument indicates that the input provided was invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrPermissionDenied indicates that the user does not have permission for the action.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrAuthenticationFailed indicates that user authentication failed.
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrUnauthenticated indicates the user needs to be authenticated.
	ErrUnauthenticated = errors.New("unauthenticated") // Could be used by middleware later
)
```

## `docs/swagger.yaml`

```yaml
basePath: /api/v1
definitions:
  dto.AudioCollectionResponseDTO:
    properties:
      createdAt:
        type: string
      description:
        type: string
      id:
        type: string
      ownerId:
        type: string
      title:
        type: string
      tracks:
        items:
          $ref: '#/definitions/dto.AudioTrackResponseDTO'
        type: array
      type:
        type: string
      updatedAt:
        type: string
    type: object
  dto.AudioTrackDetailsResponseDTO:
    properties:
      coverImageUrl:
        type: string
      createdAt:
        type: string
      description:
        type: string
      durationMs:
        description: 'Point 1: Use milliseconds (int64)'
        type: integer
      id:
        type: string
      isPublic:
        type: boolean
      languageCode:
        type: string
      level:
        type: string
      playUrl:
        description: Presigned URL
        type: string
      tags:
        items:
          type: string
        type: array
      title:
        type: string
      updatedAt:
        type: string
      uploaderId:
        type: string
      userBookmarks:
        description: Array of user bookmarks for this track
        items:
          $ref: '#/definitions/dto.BookmarkResponseDTO'
        type: array
      userProgressMs:
        description: 'Point 1: User progress in ms'
        type: integer
    type: object
  dto.AudioTrackResponseDTO:
    properties:
      coverImageUrl:
        type: string
      createdAt:
        type: string
      description:
        type: string
      durationMs:
        description: 'Point 1: Use milliseconds (int64)'
        type: integer
      id:
        type: string
      isPublic:
        type: boolean
      languageCode:
        type: string
      level:
        type: string
      tags:
        items:
          type: string
        type: array
      title:
        type: string
      updatedAt:
        type: string
      uploaderId:
        type: string
    type: object
  dto.AuthResponseDTO:
    properties:
      accessToken:
        description: The JWT access token
        type: string
      isNewUser:
        description: Pointer, only included for Google callback if user is new
        type: boolean
      refreshToken:
        description: The refresh token value
        type: string
    type: object
  dto.BatchCompleteUploadInputDTO:
    properties:
      tracks:
        description: Ensure at least one track, validate each item
        items:
          $ref: '#/definitions/dto.BatchCompleteUploadItemDTO'
        minItems: 1
        type: array
    required:
    - tracks
    type: object
  dto.BatchCompleteUploadItemDTO:
    properties:
      coverImageUrl:
        type: string
      description:
        type: string
      durationMs:
        type: integer
      isPublic:
        type: boolean
      languageCode:
        type: string
      level:
        enum:
        - A1
        - A2
        - B1
        - B2
        - C1
        - C2
        - NATIVE
        type: string
      objectKey:
        type: string
      tags:
        items:
          type: string
        type: array
      title:
        maxLength: 255
        type: string
    required:
    - durationMs
    - languageCode
    - objectKey
    - title
    type: object
  dto.BatchCompleteUploadResponseDTO:
    properties:
      results:
        items:
          $ref: '#/definitions/dto.BatchCompleteUploadResponseItemDTO'
        type: array
    type: object
  dto.BatchCompleteUploadResponseItemDTO:
    properties:
      error:
        description: Error message if processing failed for this item
        type: string
      objectKey:
        description: Identifies the item
        type: string
      success:
        description: Whether processing this item succeeded
        type: boolean
      trackId:
        description: The ID of the created track if successful
        type: string
    type: object
  dto.BatchRequestUploadInputItemDTO:
    properties:
      contentType:
        description: e.g., "audio/mpeg"
        type: string
      filename:
        type: string
    required:
    - contentType
    - filename
    type: object
  dto.BatchRequestUploadInputRequestDTO:
    properties:
      files:
        description: Ensure at least one file, validate each item
        items:
          $ref: '#/definitions/dto.BatchRequestUploadInputItemDTO'
        minItems: 1
        type: array
    required:
    - files
    type: object
  dto.BatchRequestUploadInputResponseDTO:
    properties:
      results:
        items:
          $ref: '#/definitions/dto.BatchRequestUploadInputResponseItemDTO'
        type: array
    type: object
  dto.BatchRequestUploadInputResponseItemDTO:
    properties:
      error:
        description: Error message if URL generation failed for this item
        type: string
      objectKey:
        description: The generated object key for this file
        type: string
      originalFilename:
        description: Helps client match response to request
        type: string
      uploadUrl:
        description: The presigned PUT URL for this file
        type: string
    type: object
  dto.BookmarkResponseDTO:
    properties:
      createdAt:
        type: string
      id:
        type: string
      note:
        type: string
      timestampMs:
        type: integer
      trackId:
        type: string
      userId:
        type: string
    type: object
  dto.CompleteUploadInputDTO:
    properties:
      coverImageUrl:
        type: string
      description:
        type: string
      durationMs:
        description: Duration in Milliseconds, must be positive
        type: integer
      isPublic:
        description: Defaults to false if omitted? Define behavior.
        type: boolean
      languageCode:
        type: string
      level:
        description: Allow empty or valid level
        enum:
        - A1
        - A2
        - B1
        - B2
        - C1
        - C2
        - NATIVE
        type: string
      objectKey:
        type: string
      tags:
        items:
          type: string
        type: array
      title:
        maxLength: 255
        type: string
    required:
    - durationMs
    - languageCode
    - objectKey
    - title
    type: object
  dto.CreateBookmarkRequestDTO:
    properties:
      note:
        type: string
      timestampMs:
        description: 'Point 1: Already uses ms'
        minimum: 0
        type: integer
      trackId:
        type: string
    required:
    - timestampMs
    - trackId
    type: object
  dto.CreateCollectionRequestDTO:
    properties:
      description:
        type: string
      initialTrackIds:
        items:
          type: string
        type: array
      title:
        maxLength: 255
        type: string
      type:
        enum:
        - COURSE
        - PLAYLIST
        type: string
    required:
    - title
    - type
    type: object
  dto.GoogleCallbackRequestDTO:
    properties:
      idToken:
        type: string
    required:
    - idToken
    type: object
  dto.LoginRequestDTO:
    properties:
      email:
        type: string
      password:
        type: string
    required:
    - email
    - password
    type: object
  dto.LogoutRequestDTO:
    properties:
      refreshToken:
        type: string
    required:
    - refreshToken
    type: object
  dto.PaginatedResponseDTO:
    properties:
      data:
        description: The slice of items for the current page (e.g., []AudioTrackResponseDTO)
      limit:
        description: The limit used for this page
        type: integer
      offset:
        description: The offset used for this page
        type: integer
      page:
        description: Current page number (1-based)
        type: integer
      total:
        description: Total number of items matching the query
        type: integer
      totalPages:
        description: Total number of pages
        type: integer
    type: object
  dto.PlaybackProgressResponseDTO:
    properties:
      lastListenedAt:
        type: string
      progressMs:
        description: 'Point 1: Already uses ms'
        type: integer
      trackId:
        type: string
      userId:
        type: string
    type: object
  dto.RecordProgressRequestDTO:
    properties:
      progressMs:
        description: 'Point 1: Already uses ms'
        minimum: 0
        type: integer
      trackId:
        type: string
    required:
    - progressMs
    - trackId
    type: object
  dto.RefreshRequestDTO:
    properties:
      refreshToken:
        type: string
    required:
    - refreshToken
    type: object
  dto.RegisterRequestDTO:
    properties:
      email:
        description: Add example tag
        example: user@example.com
        type: string
      name:
        example: John Doe
        maxLength: 100
        type: string
      password:
        description: Add format tag
        example: Str0ngP@ssw0rd
        format: password
        minLength: 8
        type: string
    required:
    - email
    - name
    - password
    type: object
  dto.RequestUploadRequestDTO:
    properties:
      contentType:
        description: e.g., "audio/mpeg"
        type: string
      filename:
        type: string
    required:
    - contentType
    - filename
    type: object
  dto.RequestUploadResponseDTO:
    properties:
      objectKey:
        description: The key the client should use/report back
        type: string
      uploadUrl:
        description: The presigned PUT URL
        type: string
    type: object
  dto.UpdateCollectionRequestDTO:
    properties:
      description:
        type: string
      title:
        maxLength: 255
        type: string
    required:
    - title
    type: object
  dto.UpdateCollectionTracksRequestDTO:
    properties:
      orderedTrackIds:
        items:
          type: string
        type: array
    type: object
  dto.UserResponseDTO:
    properties:
      authProvider:
        type: string
      createdAt:
        description: Use string format like RFC3339
        type: string
      email:
        type: string
      id:
        type: string
      name:
        type: string
      profileImageUrl:
        type: string
      updatedAt:
        type: string
    type: object
  httputil.ErrorResponseDTO:
    properties:
      code:
        description: Application-specific error code (e.g., "INVALID_INPUT", "NOT_FOUND")
        type: string
      message:
        description: User-friendly error message
        type: string
      requestId:
        description: Include request ID for tracing
        type: string
    type: object
host: localhost:8080
info:
  contact:
    email: support@example.com
    name: API Support Team
    url: http://www.example.com/support
  description: API specification for the backend of the Language Learning Audio Player
    application. Provides endpoints for user authentication, audio content management,
    and user activity tracking.
  license:
    name: Apache 2.0
    url: http://www.apache.org/licenses/LICENSE-2.0.html
  termsOfService: http://swagger.io/terms/
  title: Language Learning Audio Player API
  version: 1.0.0
paths:
  /audio/collections:
    post:
      consumes:
      - application/json
      description: Creates a new audio collection (playlist or course) for the authenticated
        user.
      operationId: create-audio-collection
      parameters:
      - description: Collection details
        in: body
        name: collection
        required: true
        schema:
          $ref: '#/definitions/dto.CreateCollectionRequestDTO'
      produces:
      - application/json
      responses:
        "201":
          description: Collection created successfully
          schema:
            $ref: '#/definitions/dto.AudioCollectionResponseDTO'
        "400":
          description: Invalid Input / Track ID Format / Collection Type
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Create an audio collection
      tags:
      - Audio Collections
  /audio/collections/{collectionId}:
    delete:
      description: Deletes an audio collection owned by the authenticated user.
      operationId: delete-audio-collection
      parameters:
      - description: Audio Collection UUID
        format: uuid
        in: path
        name: collectionId
        required: true
        type: string
      produces:
      - application/json
      responses:
        "204":
          description: Collection deleted successfully
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "403":
          description: Forbidden (Not Owner)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Collection Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Delete an audio collection
      tags:
      - Audio Collections
    get:
      description: Retrieves details for a specific audio collection, including its
        metadata and ordered list of tracks.
      operationId: get-collection-details
      parameters:
      - description: Audio Collection UUID
        format: uuid
        in: path
        name: collectionId
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: Audio collection details found
          schema:
            $ref: '#/definitions/dto.AudioCollectionResponseDTO'
        "400":
          description: Invalid Collection ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Collection Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error (e.g., failed to fetch tracks)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Get audio collection details
      tags:
      - Audio Collections
    put:
      consumes:
      - application/json
      description: Updates the title and description of an audio collection owned
        by the authenticated user.
      operationId: update-collection-metadata
      parameters:
      - description: Audio Collection UUID
        format: uuid
        in: path
        name: collectionId
        required: true
        type: string
      - description: Updated collection metadata
        in: body
        name: collection
        required: true
        schema:
          $ref: '#/definitions/dto.UpdateCollectionRequestDTO'
      produces:
      - application/json
      responses:
        "204":
          description: Collection metadata updated successfully
        "400":
          description: Invalid Input / Collection ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "403":
          description: Forbidden (Not Owner)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Collection Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Update collection metadata
      tags:
      - Audio Collections
  /audio/collections/{collectionId}/tracks:
    put:
      consumes:
      - application/json
      description: Updates the ordered list of tracks within a specific collection
        owned by the authenticated user. Replaces the entire list.
      operationId: update-collection-tracks
      parameters:
      - description: Audio Collection UUID
        format: uuid
        in: path
        name: collectionId
        required: true
        type: string
      - description: Ordered list of track UUIDs
        in: body
        name: tracks
        required: true
        schema:
          $ref: '#/definitions/dto.UpdateCollectionTracksRequestDTO'
      produces:
      - application/json
      responses:
        "204":
          description: Collection tracks updated successfully
        "400":
          description: Invalid Input / Collection or Track ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "403":
          description: Forbidden (Not Owner)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Collection Not Found / Track Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Update collection tracks
      tags:
      - Audio Collections
  /audio/tracks:
    get:
      description: Retrieves a paginated list of audio tracks, supporting filtering
        and sorting.
      operationId: list-audio-tracks
      parameters:
      - description: Search query (searches title, description)
        in: query
        name: q
        type: string
      - description: Filter by language code (e.g., en-US)
        in: query
        name: lang
        type: string
      - description: Filter by audio level (e.g., A1, B2)
        enum:
        - A1
        - A2
        - B1
        - B2
        - C1
        - C2
        - NATIVE
        in: query
        name: level
        type: string
      - description: Filter by public status (true or false)
        in: query
        name: isPublic
        type: boolean
      - collectionFormat: multi
        description: Filter by tags (e.g., ?tags=news&tags=podcast)
        in: query
        items:
          type: string
        name: tags
        type: array
      - default: createdAt
        description: Sort field (e.g., createdAt, title, durationMs)
        in: query
        name: sortBy
        type: string
      - default: desc
        description: Sort direction (asc or desc)
        enum:
        - asc
        - desc
        in: query
        name: sortDir
        type: string
      - default: 20
        description: Pagination limit
        in: query
        maximum: 100
        minimum: 1
        name: limit
        type: integer
      - default: 0
        description: Pagination offset
        in: query
        minimum: 0
        name: offset
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: Paginated list of audio tracks
          schema:
            allOf:
            - $ref: '#/definitions/dto.PaginatedResponseDTO'
            - properties:
                data:
                  items:
                    $ref: '#/definitions/dto.AudioTrackResponseDTO'
                  type: array
              type: object
        "400":
          description: Invalid Query Parameter Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: List audio tracks
      tags:
      - Audio Tracks
    post:
      consumes:
      - application/json
      description: After the client successfully uploads a file using the presigned
        URL, this endpoint is called to create the corresponding audio track metadata
        record in the database. Use `/audio/tracks/batch/complete` for batch uploads.
      operationId: complete-audio-upload
      parameters:
      - description: Track metadata and object key
        in: body
        name: completeUpload
        required: true
        schema:
          $ref: '#/definitions/dto.CompleteUploadInputDTO'
      produces:
      - application/json
      responses:
        "201":
          description: Track metadata created successfully
          schema:
            $ref: '#/definitions/dto.AudioTrackResponseDTO'
        "400":
          description: Invalid Input (e.g., validation errors, object key not found,
            file not in storage)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "403":
          description: Forbidden (Object key mismatch)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "409":
          description: Conflict (e.g., object key already used in DB)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Complete audio upload and create track metadata (Single File)
      tags:
      - Uploads
  /audio/tracks/{trackId}:
    get:
      description: Retrieves details for a specific audio track, including metadata,
        playback URL, and user-specific progress/bookmarks if authenticated.
      operationId: get-track-details
      parameters:
      - description: Audio Track UUID
        format: uuid
        in: path
        name: trackId
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: Audio track details found
          schema:
            $ref: '#/definitions/dto.AudioTrackDetailsResponseDTO'
        "400":
          description: Invalid Track ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized (if accessing private track without auth)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Track Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - 'BearerAuth // Optional: Indicate that auth affects the response (user data)': []
      summary: Get audio track details
      tags:
      - Audio Tracks
  /audio/tracks/batch/complete:
    post:
      consumes:
      - application/json
      description: After clients successfully upload multiple files using presigned
        URLs, this endpoint is called to create the corresponding audio track metadata
        records in the database within a single transaction.
      operationId: complete-batch-audio-upload
      parameters:
      - description: List of track metadata and object keys for uploaded files
        in: body
        name: batchCompleteUpload
        required: true
        schema:
          $ref: '#/definitions/dto.BatchCompleteUploadInputDTO'
      produces:
      - application/json
      responses:
        "201":
          description: Batch processing attempted. Results indicate success/failure
            per item. If overall transaction succeeded, status is 201.
          schema:
            $ref: '#/definitions/dto.BatchCompleteUploadResponseDTO'
        "400":
          description: Invalid Input (e.g., validation errors in items, files not
            in storage)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "403":
          description: Forbidden (Object key mismatch)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "409":
          description: Conflict (e.g., duplicate object key during processing)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error (e.g., transaction failure)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Complete batch audio upload and create track metadata
      tags:
      - Uploads
  /auth/google/callback:
    post:
      consumes:
      - application/json
      description: Receives the ID token from the frontend after Google sign-in, verifies
        it, and performs user registration or login, returning access and refresh
        tokens. // MODIFIED DESCRIPTION
      operationId: google-callback
      parameters:
      - description: Google ID Token
        in: body
        name: googleCallback
        required: true
        schema:
          $ref: '#/definitions/dto.GoogleCallbackRequestDTO'
      produces:
      - application/json
      responses:
        "200":
          description: Authentication successful, returns access/refresh tokens. isNewUser
            indicates new account creation." // MODIFIED DESCRIPTION
          schema:
            $ref: '#/definitions/dto.AuthResponseDTO'
        "400":
          description: Invalid Input (Missing or Invalid ID Token)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Authentication Failed (Invalid Google Token)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "409":
          description: Conflict - Email already exists with a different login method
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Handle Google OAuth callback
      tags:
      - Authentication
  /auth/login:
    post:
      consumes:
      - application/json
      description: Authenticates a user with email and password, returns access and
        refresh tokens. // MODIFIED DESCRIPTION
      operationId: login-user
      parameters:
      - description: User Login Credentials
        in: body
        name: login
        required: true
        schema:
          $ref: '#/definitions/dto.LoginRequestDTO'
      produces:
      - application/json
      responses:
        "200":
          description: Login successful, returns access and refresh tokens" // MODIFIED
            DESCRIPTION
          schema:
            $ref: '#/definitions/dto.AuthResponseDTO'
        "400":
          description: Invalid Input
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Authentication Failed
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Login a user
      tags:
      - Authentication
  /auth/logout:
    post:
      consumes:
      - application/json
      description: Invalidates the provided refresh token, effectively logging the
        user out of that session/device.
      operationId: logout-user
      parameters:
      - description: Refresh Token to invalidate
        in: body
        name: logout
        required: true
        schema:
          $ref: '#/definitions/dto.LogoutRequestDTO'
      produces:
      - application/json
      responses:
        "204":
          description: Logout successful
        "400":
          description: Invalid Input (Missing Refresh Token)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Logout user
      tags:
      - Authentication
  /auth/refresh:
    post:
      consumes:
      - application/json
      description: Provides a valid refresh token to get a new pair of access and
        refresh tokens.
      operationId: refresh-token
      parameters:
      - description: Refresh Token
        in: body
        name: refresh
        required: true
        schema:
          $ref: '#/definitions/dto.RefreshRequestDTO'
      produces:
      - application/json
      responses:
        "200":
          description: Tokens refreshed successfully
          schema:
            $ref: '#/definitions/dto.AuthResponseDTO'
        "400":
          description: Invalid Input (Missing Refresh Token)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Authentication Failed (Invalid or Expired Refresh Token)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Refresh access token
      tags:
      - Authentication
  /auth/register:
    post:
      consumes:
      - application/json
      description: Registers a new user account using email and password.
      operationId: register-user
      parameters:
      - description: User Registration Info
        in: body
        name: register
        required: true
        schema:
          $ref: '#/definitions/dto.RegisterRequestDTO'
      produces:
      - application/json
      responses:
        "201":
          description: Registration successful, returns access and refresh tokens"
            // MODIFIED DESCRIPTION
          schema:
            $ref: '#/definitions/dto.AuthResponseDTO'
        "400":
          description: Invalid Input
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "409":
          description: Conflict - Email Exists
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Register a new user
      tags:
      - Authentication
  /uploads/audio/batch/request:
    post:
      consumes:
      - application/json
      description: Requests multiple presigned URLs for uploading several audio files
        in parallel.
      operationId: request-batch-audio-upload
      parameters:
      - description: List of files to request URLs for
        in: body
        name: batchUploadRequest
        required: true
        schema:
          $ref: '#/definitions/dto.BatchRequestUploadInputRequestDTO'
      produces:
      - application/json
      responses:
        "200":
          description: List of generated presigned URLs and object keys, including
            potential errors per item.
          schema:
            $ref: '#/definitions/dto.BatchRequestUploadInputResponseDTO'
        "400":
          description: Invalid Input (e.g., empty file list)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Request presigned URLs for batch audio upload
      tags:
      - Uploads
  /uploads/audio/request:
    post:
      consumes:
      - application/json
      description: Requests a presigned URL from the object storage (MinIO/S3) that
        can be used by the client to directly upload an audio file.
      operationId: request-audio-upload
      parameters:
      - description: Upload Request Info (filename, content type)
        in: body
        name: uploadRequest
        required: true
        schema:
          $ref: '#/definitions/dto.RequestUploadRequestDTO'
      produces:
      - application/json
      responses:
        "200":
          description: Presigned URL and object key generated
          schema:
            $ref: '#/definitions/dto.RequestUploadResponseDTO'
        "400":
          description: Invalid Input
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error (e.g., failed to generate URL)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Request presigned URL for audio upload
      tags:
      - Uploads
  /users/me:
    get:
      description: Retrieves the profile information for the currently authenticated
        user.
      operationId: get-my-profile
      produces:
      - application/json
      responses:
        "200":
          description: User profile retrieved successfully
          schema:
            $ref: '#/definitions/dto.UserResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: User Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Get current user's profile
      tags:
      - Users
  /users/me/bookmarks:
    get:
      description: Retrieves a paginated list of bookmarks for the authenticated user,
        optionally filtered by track ID.
      operationId: list-bookmarks
      parameters:
      - description: Filter by Audio Track UUID
        format: uuid
        in: query
        name: trackId
        type: string
      - default: 50
        description: Pagination limit
        in: query
        maximum: 100
        minimum: 1
        name: limit
        type: integer
      - default: 0
        description: Pagination offset
        in: query
        minimum: 0
        name: offset
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: Paginated list of bookmarks (timestampMs in milliseconds)
          schema:
            allOf:
            - $ref: '#/definitions/dto.PaginatedResponseDTO'
            - properties:
                data:
                  items:
                    $ref: '#/definitions/dto.BookmarkResponseDTO'
                  type: array
              type: object
        "400":
          description: Invalid Track ID Format (if provided)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: List user's bookmarks
      tags:
      - User Activity
    post:
      consumes:
      - application/json
      description: Creates a new bookmark at a specific timestamp (in milliseconds)
        in an audio track for the authenticated user.
      operationId: create-bookmark
      parameters:
      - description: Bookmark details (timestampMs in milliseconds)
        in: body
        name: bookmark
        required: true
        schema:
          $ref: '#/definitions/dto.CreateBookmarkRequestDTO'
      produces:
      - application/json
      responses:
        "201":
          description: Bookmark created successfully (timestampMs in milliseconds)
          schema:
            $ref: '#/definitions/dto.BookmarkResponseDTO'
        "400":
          description: Invalid Input / Track ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Track Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Create a bookmark
      tags:
      - User Activity
  /users/me/bookmarks/{bookmarkId}:
    delete:
      description: Deletes a specific bookmark owned by the current user.
      operationId: delete-bookmark
      parameters:
      - description: Bookmark UUID
        format: uuid
        in: path
        name: bookmarkId
        required: true
        type: string
      produces:
      - application/json
      responses:
        "204":
          description: Bookmark deleted successfully
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "403":
          description: Forbidden (Not Owner)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Bookmark Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Delete a bookmark
      tags:
      - User Activity
  /users/me/progress:
    get:
      description: Retrieves a paginated list of playback progress records for the
        authenticated user.
      operationId: list-playback-progress
      parameters:
      - default: 50
        description: Pagination limit
        in: query
        maximum: 100
        minimum: 1
        name: limit
        type: integer
      - default: 0
        description: Pagination offset
        in: query
        minimum: 0
        name: offset
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: Paginated list of playback progress (progressMs in milliseconds)
          schema:
            allOf:
            - $ref: '#/definitions/dto.PaginatedResponseDTO'
            - properties:
                data:
                  items:
                    $ref: '#/definitions/dto.PlaybackProgressResponseDTO'
                  type: array
              type: object
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: List user's playback progress
      tags:
      - User Activity
    post:
      consumes:
      - application/json
      description: Records or updates the playback progress for a specific audio track
        for the authenticated user.
      operationId: record-playback-progress
      parameters:
      - description: Playback progress details (progressMs in milliseconds)
        in: body
        name: progress
        required: true
        schema:
          $ref: '#/definitions/dto.RecordProgressRequestDTO'
      produces:
      - application/json
      responses:
        "204":
          description: Progress recorded successfully
        "400":
          description: Invalid Input / Track ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Track Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Record playback progress
      tags:
      - User Activity
  /users/me/progress/{trackId}:
    get:
      description: Retrieves the playback progress for a specific audio track for
        the authenticated user.
      operationId: get-playback-progress
      parameters:
      - description: Audio Track UUID
        format: uuid
        in: path
        name: trackId
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: Playback progress found (progressMs in milliseconds)
          schema:
            $ref: '#/definitions/dto.PlaybackProgressResponseDTO'
        "400":
          description: Invalid Track ID Format
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "404":
          description: Progress Not Found (or Track Not Found)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Get playback progress for a track
      tags:
      - User Activity
schemes:
- http
- https
securityDefinitions:
  BearerAuth:
    description: 'Type "Bearer" followed by a space and JWT token. Example: "Bearer
      eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."'
    in: header
    name: Authorization
    type: apiKey
swagger: "2.0"
tags:
- description: Operations related to user signup, login, and external authentication
    (e.g., Google).
  name: Authentication
- description: Operations related to user profiles.
  name: Users
- description: Operations related to individual audio tracks, including retrieval
    and listing.
  name: Audio Tracks
- description: Operations related to managing audio collections (playlists, courses).
  name: Audio Collections
- description: Operations related to tracking user interactions like playback progress
    and bookmarks.
  name: User Activity
- description: Operations related to requesting upload URLs and finalizing uploads.
  name: Uploads
- description: API health checks.
  name: Health
```

## `docs/docs.go`

```go
// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support Team",
            "url": "http://www.example.com/support",
            "email": "support@example.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/audio/collections": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Creates a new audio collection (playlist or course) for the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Create an audio collection",
                "operationId": "create-audio-collection",
                "parameters": [
                    {
                        "description": "Collection details",
                        "name": "collection",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CreateCollectionRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Collection created successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioCollectionResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input / Track ID Format / Collection Type",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/collections/{collectionId}": {
            "get": {
                "description": "Retrieves details for a specific audio collection, including its metadata and ordered list of tracks.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Get audio collection details",
                "operationId": "get-collection-details",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Audio collection details found",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioCollectionResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Collection ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error (e.g., failed to fetch tracks)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Updates the title and description of an audio collection owned by the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Update collection metadata",
                "operationId": "update-collection-metadata",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated collection metadata",
                        "name": "collection",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.UpdateCollectionRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Collection metadata updated successfully"
                    },
                    "400": {
                        "description": "Invalid Input / Collection ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Deletes an audio collection owned by the authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Delete an audio collection",
                "operationId": "delete-audio-collection",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Collection deleted successfully"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/collections/{collectionId}/tracks": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Updates the ordered list of tracks within a specific collection owned by the authenticated user. Replaces the entire list.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Update collection tracks",
                "operationId": "update-collection-tracks",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Ordered list of track UUIDs",
                        "name": "tracks",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.UpdateCollectionTracksRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Collection tracks updated successfully"
                    },
                    "400": {
                        "description": "Invalid Input / Collection or Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found / Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/tracks": {
            "get": {
                "description": "Retrieves a paginated list of audio tracks, supporting filtering and sorting.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Tracks"
                ],
                "summary": "List audio tracks",
                "operationId": "list-audio-tracks",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Search query (searches title, description)",
                        "name": "q",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by language code (e.g., en-US)",
                        "name": "lang",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "A1",
                            "A2",
                            "B1",
                            "B2",
                            "C1",
                            "C2",
                            "NATIVE"
                        ],
                        "type": "string",
                        "description": "Filter by audio level (e.g., A1, B2)",
                        "name": "level",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Filter by public status (true or false)",
                        "name": "isPublic",
                        "in": "query"
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "collectionFormat": "multi",
                        "description": "Filter by tags (e.g., ?tags=news\u0026tags=podcast)",
                        "name": "tags",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "createdAt",
                        "description": "Sort field (e.g., createdAt, title, durationMs)",
                        "name": "sortBy",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "asc",
                            "desc"
                        ],
                        "type": "string",
                        "default": "desc",
                        "description": "Sort direction (asc or desc)",
                        "name": "sortDir",
                        "in": "query"
                    },
                    {
                        "maximum": 100,
                        "minimum": 1,
                        "type": "integer",
                        "default": 20,
                        "description": "Pagination limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Paginated list of audio tracks",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/dto.PaginatedResponseDTO"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Invalid Query Parameter Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database. Use ` + "`" + `/audio/tracks/batch/complete` + "`" + ` for batch uploads.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Complete audio upload and create track metadata (Single File)",
                "operationId": "complete-audio-upload",
                "parameters": [
                    {
                        "description": "Track metadata and object key",
                        "name": "completeUpload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CompleteUploadInputDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Track metadata created successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (e.g., validation errors, object key not found, file not in storage)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Object key mismatch)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict (e.g., object key already used in DB)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/tracks/batch/complete": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "After clients successfully upload multiple files using presigned URLs, this endpoint is called to create the corresponding audio track metadata records in the database within a single transaction.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Complete batch audio upload and create track metadata",
                "operationId": "complete-batch-audio-upload",
                "parameters": [
                    {
                        "description": "List of track metadata and object keys for uploaded files",
                        "name": "batchCompleteUpload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.BatchCompleteUploadInputDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Batch processing attempted. Results indicate success/failure per item. If overall transaction succeeded, status is 201.",
                        "schema": {
                            "$ref": "#/definitions/dto.BatchCompleteUploadResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (e.g., validation errors in items, files not in storage)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Object key mismatch)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict (e.g., duplicate object key during processing)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error (e.g., transaction failure)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/tracks/{trackId}": {
            "get": {
                "security": [
                    {
                        "BearerAuth // Optional: Indicate that auth affects the response (user data)": []
                    }
                ],
                "description": "Retrieves details for a specific audio track, including metadata, playback URL, and user-specific progress/bookmarks if authenticated.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Tracks"
                ],
                "summary": "Get audio track details",
                "operationId": "get-track-details",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Track UUID",
                        "name": "trackId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Audio track details found",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioTrackDetailsResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized (if accessing private track without auth)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/google/callback": {
            "post": {
                "description": "Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning access and refresh tokens. // MODIFIED DESCRIPTION",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Handle Google OAuth callback",
                "operationId": "google-callback",
                "parameters": [
                    {
                        "description": "Google ID Token",
                        "name": "googleCallback",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.GoogleCallbackRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Authentication successful, returns access/refresh tokens. isNewUser indicates new account creation.\" // MODIFIED DESCRIPTION",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (Missing or Invalid ID Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Authentication Failed (Invalid Google Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict - Email already exists with a different login method",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/login": {
            "post": {
                "description": "Authenticates a user with email and password, returns access and refresh tokens. // MODIFIED DESCRIPTION",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Login a user",
                "operationId": "login-user",
                "parameters": [
                    {
                        "description": "User Login Credentials",
                        "name": "login",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.LoginRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Login successful, returns access and refresh tokens\" // MODIFIED DESCRIPTION",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Authentication Failed",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/logout": {
            "post": {
                "description": "Invalidates the provided refresh token, effectively logging the user out of that session/device.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Logout user",
                "operationId": "logout-user",
                "parameters": [
                    {
                        "description": "Refresh Token to invalidate",
                        "name": "logout",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.LogoutRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Logout successful"
                    },
                    "400": {
                        "description": "Invalid Input (Missing Refresh Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Provides a valid refresh token to get a new pair of access and refresh tokens.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Refresh access token",
                "operationId": "refresh-token",
                "parameters": [
                    {
                        "description": "Refresh Token",
                        "name": "refresh",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RefreshRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Tokens refreshed successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (Missing Refresh Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Authentication Failed (Invalid or Expired Refresh Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Registers a new user account using email and password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Register a new user",
                "operationId": "register-user",
                "parameters": [
                    {
                        "description": "User Registration Info",
                        "name": "register",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RegisterRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Registration successful, returns access and refresh tokens\" // MODIFIED DESCRIPTION",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict - Email Exists",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/uploads/audio/batch/request": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Requests multiple presigned URLs for uploading several audio files in parallel.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Request presigned URLs for batch audio upload",
                "operationId": "request-batch-audio-upload",
                "parameters": [
                    {
                        "description": "List of files to request URLs for",
                        "name": "batchUploadRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.BatchRequestUploadInputRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of generated presigned URLs and object keys, including potential errors per item.",
                        "schema": {
                            "$ref": "#/definitions/dto.BatchRequestUploadInputResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (e.g., empty file list)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/uploads/audio/request": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Requests a presigned URL from the object storage (MinIO/S3) that can be used by the client to directly upload an audio file.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Request presigned URL for audio upload",
                "operationId": "request-audio-upload",
                "parameters": [
                    {
                        "description": "Upload Request Info (filename, content type)",
                        "name": "uploadRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RequestUploadRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Presigned URL and object key generated",
                        "schema": {
                            "$ref": "#/definitions/dto.RequestUploadResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error (e.g., failed to generate URL)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves the profile information for the currently authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Get current user's profile",
                "operationId": "get-my-profile",
                "responses": {
                    "200": {
                        "description": "User profile retrieved successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.UserResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "User Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/bookmarks": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves a paginated list of bookmarks for the authenticated user, optionally filtered by track ID.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "List user's bookmarks",
                "operationId": "list-bookmarks",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Filter by Audio Track UUID",
                        "name": "trackId",
                        "in": "query"
                    },
                    {
                        "maximum": 100,
                        "minimum": 1,
                        "type": "integer",
                        "default": 50,
                        "description": "Pagination limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Paginated list of bookmarks (timestampMs in milliseconds)",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/dto.PaginatedResponseDTO"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/dto.BookmarkResponseDTO"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Invalid Track ID Format (if provided)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Creates a new bookmark at a specific timestamp (in milliseconds) in an audio track for the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Create a bookmark",
                "operationId": "create-bookmark",
                "parameters": [
                    {
                        "description": "Bookmark details (timestampMs in milliseconds)",
                        "name": "bookmark",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CreateBookmarkRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Bookmark created successfully (timestampMs in milliseconds)",
                        "schema": {
                            "$ref": "#/definitions/dto.BookmarkResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input / Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/bookmarks/{bookmarkId}": {
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Deletes a specific bookmark owned by the current user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Delete a bookmark",
                "operationId": "delete-bookmark",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Bookmark UUID",
                        "name": "bookmarkId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Bookmark deleted successfully"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Bookmark Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/progress": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves a paginated list of playback progress records for the authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "List user's playback progress",
                "operationId": "list-playback-progress",
                "parameters": [
                    {
                        "maximum": 100,
                        "minimum": 1,
                        "type": "integer",
                        "default": 50,
                        "description": "Pagination limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Paginated list of playback progress (progressMs in milliseconds)",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/dto.PaginatedResponseDTO"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/dto.PlaybackProgressResponseDTO"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Records or updates the playback progress for a specific audio track for the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Record playback progress",
                "operationId": "record-playback-progress",
                "parameters": [
                    {
                        "description": "Playback progress details (progressMs in milliseconds)",
                        "name": "progress",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RecordProgressRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Progress recorded successfully"
                    },
                    "400": {
                        "description": "Invalid Input / Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/progress/{trackId}": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves the playback progress for a specific audio track for the authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Get playback progress for a track",
                "operationId": "get-playback-progress",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Track UUID",
                        "name": "trackId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Playback progress found (progressMs in milliseconds)",
                        "schema": {
                            "$ref": "#/definitions/dto.PlaybackProgressResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Progress Not Found (or Track Not Found)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "dto.AudioCollectionResponseDTO": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "ownerId": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "tracks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                    }
                },
                "type": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "dto.AudioTrackDetailsResponseDTO": {
            "type": "object",
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "Point 1: Use milliseconds (int64)",
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "isPublic": {
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "type": "string"
                },
                "playUrl": {
                    "description": "Presigned URL",
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "uploaderId": {
                    "type": "string"
                },
                "userBookmarks": {
                    "description": "Array of user bookmarks for this track",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BookmarkResponseDTO"
                    }
                },
                "userProgressMs": {
                    "description": "Point 1: User progress in ms",
                    "type": "integer"
                }
            }
        },
        "dto.AudioTrackResponseDTO": {
            "type": "object",
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "Point 1: Use milliseconds (int64)",
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "isPublic": {
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "uploaderId": {
                    "type": "string"
                }
            }
        },
        "dto.AuthResponseDTO": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "description": "The JWT access token",
                    "type": "string"
                },
                "isNewUser": {
                    "description": "Pointer, only included for Google callback if user is new",
                    "type": "boolean"
                },
                "refreshToken": {
                    "description": "The refresh token value",
                    "type": "string"
                }
            }
        },
        "dto.BatchCompleteUploadInputDTO": {
            "type": "object",
            "required": [
                "tracks"
            ],
            "properties": {
                "tracks": {
                    "description": "Ensure at least one track, validate each item",
                    "type": "array",
                    "minItems": 1,
                    "items": {
                        "$ref": "#/definitions/dto.BatchCompleteUploadItemDTO"
                    }
                }
            }
        },
        "dto.BatchCompleteUploadItemDTO": {
            "type": "object",
            "required": [
                "durationMs",
                "languageCode",
                "objectKey",
                "title"
            ],
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "type": "integer"
                },
                "isPublic": {
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "type": "string",
                    "enum": [
                        "A1",
                        "A2",
                        "B1",
                        "B2",
                        "C1",
                        "C2",
                        "NATIVE"
                    ]
                },
                "objectKey": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "dto.BatchCompleteUploadResponseDTO": {
            "type": "object",
            "properties": {
                "results": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BatchCompleteUploadResponseItemDTO"
                    }
                }
            }
        },
        "dto.BatchCompleteUploadResponseItemDTO": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "Error message if processing failed for this item",
                    "type": "string"
                },
                "objectKey": {
                    "description": "Identifies the item",
                    "type": "string"
                },
                "success": {
                    "description": "Whether processing this item succeeded",
                    "type": "boolean"
                },
                "trackId": {
                    "description": "The ID of the created track if successful",
                    "type": "string"
                }
            }
        },
        "dto.BatchRequestUploadInputItemDTO": {
            "type": "object",
            "required": [
                "contentType",
                "filename"
            ],
            "properties": {
                "contentType": {
                    "description": "e.g., \"audio/mpeg\"",
                    "type": "string"
                },
                "filename": {
                    "type": "string"
                }
            }
        },
        "dto.BatchRequestUploadInputRequestDTO": {
            "type": "object",
            "required": [
                "files"
            ],
            "properties": {
                "files": {
                    "description": "Ensure at least one file, validate each item",
                    "type": "array",
                    "minItems": 1,
                    "items": {
                        "$ref": "#/definitions/dto.BatchRequestUploadInputItemDTO"
                    }
                }
            }
        },
        "dto.BatchRequestUploadInputResponseDTO": {
            "type": "object",
            "properties": {
                "results": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BatchRequestUploadInputResponseItemDTO"
                    }
                }
            }
        },
        "dto.BatchRequestUploadInputResponseItemDTO": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "Error message if URL generation failed for this item",
                    "type": "string"
                },
                "objectKey": {
                    "description": "The generated object key for this file",
                    "type": "string"
                },
                "originalFilename": {
                    "description": "Helps client match response to request",
                    "type": "string"
                },
                "uploadUrl": {
                    "description": "The presigned PUT URL for this file",
                    "type": "string"
                }
            }
        },
        "dto.BookmarkResponseDTO": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "note": {
                    "type": "string"
                },
                "timestampMs": {
                    "type": "integer"
                },
                "trackId": {
                    "type": "string"
                },
                "userId": {
                    "type": "string"
                }
            }
        },
        "dto.CompleteUploadInputDTO": {
            "type": "object",
            "required": [
                "durationMs",
                "languageCode",
                "objectKey",
                "title"
            ],
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "Duration in Milliseconds, must be positive",
                    "type": "integer"
                },
                "isPublic": {
                    "description": "Defaults to false if omitted? Define behavior.",
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "description": "Allow empty or valid level",
                    "type": "string",
                    "enum": [
                        "A1",
                        "A2",
                        "B1",
                        "B2",
                        "C1",
                        "C2",
                        "NATIVE"
                    ]
                },
                "objectKey": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "dto.CreateBookmarkRequestDTO": {
            "type": "object",
            "required": [
                "timestampMs",
                "trackId"
            ],
            "properties": {
                "note": {
                    "type": "string"
                },
                "timestampMs": {
                    "description": "Point 1: Already uses ms",
                    "type": "integer",
                    "minimum": 0
                },
                "trackId": {
                    "type": "string"
                }
            }
        },
        "dto.CreateCollectionRequestDTO": {
            "type": "object",
            "required": [
                "title",
                "type"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "initialTrackIds": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                },
                "type": {
                    "type": "string",
                    "enum": [
                        "COURSE",
                        "PLAYLIST"
                    ]
                }
            }
        },
        "dto.GoogleCallbackRequestDTO": {
            "type": "object",
            "required": [
                "idToken"
            ],
            "properties": {
                "idToken": {
                    "type": "string"
                }
            }
        },
        "dto.LoginRequestDTO": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "dto.LogoutRequestDTO": {
            "type": "object",
            "required": [
                "refreshToken"
            ],
            "properties": {
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "dto.PaginatedResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "The slice of items for the current page (e.g., []AudioTrackResponseDTO)"
                },
                "limit": {
                    "description": "The limit used for this page",
                    "type": "integer"
                },
                "offset": {
                    "description": "The offset used for this page",
                    "type": "integer"
                },
                "page": {
                    "description": "Current page number (1-based)",
                    "type": "integer"
                },
                "total": {
                    "description": "Total number of items matching the query",
                    "type": "integer"
                },
                "totalPages": {
                    "description": "Total number of pages",
                    "type": "integer"
                }
            }
        },
        "dto.PlaybackProgressResponseDTO": {
            "type": "object",
            "properties": {
                "lastListenedAt": {
                    "type": "string"
                },
                "progressMs": {
                    "description": "Point 1: Already uses ms",
                    "type": "integer"
                },
                "trackId": {
                    "type": "string"
                },
                "userId": {
                    "type": "string"
                }
            }
        },
        "dto.RecordProgressRequestDTO": {
            "type": "object",
            "required": [
                "progressMs",
                "trackId"
            ],
            "properties": {
                "progressMs": {
                    "description": "Point 1: Already uses ms",
                    "type": "integer",
                    "minimum": 0
                },
                "trackId": {
                    "type": "string"
                }
            }
        },
        "dto.RefreshRequestDTO": {
            "type": "object",
            "required": [
                "refreshToken"
            ],
            "properties": {
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "dto.RegisterRequestDTO": {
            "type": "object",
            "required": [
                "email",
                "name",
                "password"
            ],
            "properties": {
                "email": {
                    "description": "Add example tag",
                    "type": "string",
                    "example": "user@example.com"
                },
                "name": {
                    "type": "string",
                    "maxLength": 100,
                    "example": "John Doe"
                },
                "password": {
                    "description": "Add format tag",
                    "type": "string",
                    "format": "password",
                    "minLength": 8,
                    "example": "Str0ngP@ssw0rd"
                }
            }
        },
        "dto.RequestUploadRequestDTO": {
            "type": "object",
            "required": [
                "contentType",
                "filename"
            ],
            "properties": {
                "contentType": {
                    "description": "e.g., \"audio/mpeg\"",
                    "type": "string"
                },
                "filename": {
                    "type": "string"
                }
            }
        },
        "dto.RequestUploadResponseDTO": {
            "type": "object",
            "properties": {
                "objectKey": {
                    "description": "The key the client should use/report back",
                    "type": "string"
                },
                "uploadUrl": {
                    "description": "The presigned PUT URL",
                    "type": "string"
                }
            }
        },
        "dto.UpdateCollectionRequestDTO": {
            "type": "object",
            "required": [
                "title"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "dto.UpdateCollectionTracksRequestDTO": {
            "type": "object",
            "properties": {
                "orderedTrackIds": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "dto.UserResponseDTO": {
            "type": "object",
            "properties": {
                "authProvider": {
                    "type": "string"
                },
                "createdAt": {
                    "description": "Use string format like RFC3339",
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "profileImageUrl": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "httputil.ErrorResponseDTO": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "Application-specific error code (e.g., \"INVALID_INPUT\", \"NOT_FOUND\")",
                    "type": "string"
                },
                "message": {
                    "description": "User-friendly error message",
                    "type": "string"
                },
                "requestId": {
                    "description": "Include request ID for tracing",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Type \"Bearer\" followed by a space and JWT token. Example: \"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...\"",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    },
    "tags": [
        {
            "description": "Operations related to user signup, login, and external authentication (e.g., Google).",
            "name": "Authentication"
        },
        {
            "description": "Operations related to user profiles.",
            "name": "Users"
        },
        {
            "description": "Operations related to individual audio tracks, including retrieval and listing.",
            "name": "Audio Tracks"
        },
        {
            "description": "Operations related to managing audio collections (playlists, courses).",
            "name": "Audio Collections"
        },
        {
            "description": "Operations related to tracking user interactions like playback progress and bookmarks.",
            "name": "User Activity"
        },
        {
            "description": "Operations related to requesting upload URLs and finalizing uploads.",
            "name": "Uploads"
        },
        {
            "description": "API health checks.",
            "name": "Health"
        }
    ]
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "Language Learning Audio Player API",
	Description:      "API specification for the backend of the Language Learning Audio Player application. Provides endpoints for user authentication, audio content management, and user activity tracking.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
```

## `docs/swagger.json`

```json
{
    "schemes": [
        "http",
        "https"
    ],
    "swagger": "2.0",
    "info": {
        "description": "API specification for the backend of the Language Learning Audio Player application. Provides endpoints for user authentication, audio content management, and user activity tracking.",
        "title": "Language Learning Audio Player API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support Team",
            "url": "http://www.example.com/support",
            "email": "support@example.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0.0"
    },
    "host": "localhost:8080",
    "basePath": "/api/v1",
    "paths": {
        "/audio/collections": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Creates a new audio collection (playlist or course) for the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Create an audio collection",
                "operationId": "create-audio-collection",
                "parameters": [
                    {
                        "description": "Collection details",
                        "name": "collection",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CreateCollectionRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Collection created successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioCollectionResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input / Track ID Format / Collection Type",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/collections/{collectionId}": {
            "get": {
                "description": "Retrieves details for a specific audio collection, including its metadata and ordered list of tracks.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Get audio collection details",
                "operationId": "get-collection-details",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Audio collection details found",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioCollectionResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Collection ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error (e.g., failed to fetch tracks)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Updates the title and description of an audio collection owned by the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Update collection metadata",
                "operationId": "update-collection-metadata",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated collection metadata",
                        "name": "collection",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.UpdateCollectionRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Collection metadata updated successfully"
                    },
                    "400": {
                        "description": "Invalid Input / Collection ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Deletes an audio collection owned by the authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Delete an audio collection",
                "operationId": "delete-audio-collection",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Collection deleted successfully"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/collections/{collectionId}/tracks": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Updates the ordered list of tracks within a specific collection owned by the authenticated user. Replaces the entire list.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Collections"
                ],
                "summary": "Update collection tracks",
                "operationId": "update-collection-tracks",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Collection UUID",
                        "name": "collectionId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Ordered list of track UUIDs",
                        "name": "tracks",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.UpdateCollectionTracksRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Collection tracks updated successfully"
                    },
                    "400": {
                        "description": "Invalid Input / Collection or Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Collection Not Found / Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/tracks": {
            "get": {
                "description": "Retrieves a paginated list of audio tracks, supporting filtering and sorting.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Tracks"
                ],
                "summary": "List audio tracks",
                "operationId": "list-audio-tracks",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Search query (searches title, description)",
                        "name": "q",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by language code (e.g., en-US)",
                        "name": "lang",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "A1",
                            "A2",
                            "B1",
                            "B2",
                            "C1",
                            "C2",
                            "NATIVE"
                        ],
                        "type": "string",
                        "description": "Filter by audio level (e.g., A1, B2)",
                        "name": "level",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Filter by public status (true or false)",
                        "name": "isPublic",
                        "in": "query"
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "collectionFormat": "multi",
                        "description": "Filter by tags (e.g., ?tags=news\u0026tags=podcast)",
                        "name": "tags",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "createdAt",
                        "description": "Sort field (e.g., createdAt, title, durationMs)",
                        "name": "sortBy",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "asc",
                            "desc"
                        ],
                        "type": "string",
                        "default": "desc",
                        "description": "Sort direction (asc or desc)",
                        "name": "sortDir",
                        "in": "query"
                    },
                    {
                        "maximum": 100,
                        "minimum": 1,
                        "type": "integer",
                        "default": 20,
                        "description": "Pagination limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Paginated list of audio tracks",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/dto.PaginatedResponseDTO"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Invalid Query Parameter Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database. Use `/audio/tracks/batch/complete` for batch uploads.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Complete audio upload and create track metadata (Single File)",
                "operationId": "complete-audio-upload",
                "parameters": [
                    {
                        "description": "Track metadata and object key",
                        "name": "completeUpload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CompleteUploadInputDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Track metadata created successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (e.g., validation errors, object key not found, file not in storage)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Object key mismatch)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict (e.g., object key already used in DB)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/tracks/batch/complete": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "After clients successfully upload multiple files using presigned URLs, this endpoint is called to create the corresponding audio track metadata records in the database within a single transaction.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Complete batch audio upload and create track metadata",
                "operationId": "complete-batch-audio-upload",
                "parameters": [
                    {
                        "description": "List of track metadata and object keys for uploaded files",
                        "name": "batchCompleteUpload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.BatchCompleteUploadInputDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Batch processing attempted. Results indicate success/failure per item. If overall transaction succeeded, status is 201.",
                        "schema": {
                            "$ref": "#/definitions/dto.BatchCompleteUploadResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (e.g., validation errors in items, files not in storage)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Object key mismatch)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict (e.g., duplicate object key during processing)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error (e.g., transaction failure)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/audio/tracks/{trackId}": {
            "get": {
                "security": [
                    {
                        "BearerAuth // Optional: Indicate that auth affects the response (user data)": []
                    }
                ],
                "description": "Retrieves details for a specific audio track, including metadata, playback URL, and user-specific progress/bookmarks if authenticated.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Audio Tracks"
                ],
                "summary": "Get audio track details",
                "operationId": "get-track-details",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Track UUID",
                        "name": "trackId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Audio track details found",
                        "schema": {
                            "$ref": "#/definitions/dto.AudioTrackDetailsResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized (if accessing private track without auth)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/google/callback": {
            "post": {
                "description": "Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning access and refresh tokens. // MODIFIED DESCRIPTION",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Handle Google OAuth callback",
                "operationId": "google-callback",
                "parameters": [
                    {
                        "description": "Google ID Token",
                        "name": "googleCallback",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.GoogleCallbackRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Authentication successful, returns access/refresh tokens. isNewUser indicates new account creation.\" // MODIFIED DESCRIPTION",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (Missing or Invalid ID Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Authentication Failed (Invalid Google Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict - Email already exists with a different login method",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/login": {
            "post": {
                "description": "Authenticates a user with email and password, returns access and refresh tokens. // MODIFIED DESCRIPTION",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Login a user",
                "operationId": "login-user",
                "parameters": [
                    {
                        "description": "User Login Credentials",
                        "name": "login",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.LoginRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Login successful, returns access and refresh tokens\" // MODIFIED DESCRIPTION",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Authentication Failed",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/logout": {
            "post": {
                "description": "Invalidates the provided refresh token, effectively logging the user out of that session/device.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Logout user",
                "operationId": "logout-user",
                "parameters": [
                    {
                        "description": "Refresh Token to invalidate",
                        "name": "logout",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.LogoutRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Logout successful"
                    },
                    "400": {
                        "description": "Invalid Input (Missing Refresh Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Provides a valid refresh token to get a new pair of access and refresh tokens.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Refresh access token",
                "operationId": "refresh-token",
                "parameters": [
                    {
                        "description": "Refresh Token",
                        "name": "refresh",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RefreshRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Tokens refreshed successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (Missing Refresh Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Authentication Failed (Invalid or Expired Refresh Token)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Registers a new user account using email and password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authentication"
                ],
                "summary": "Register a new user",
                "operationId": "register-user",
                "parameters": [
                    {
                        "description": "User Registration Info",
                        "name": "register",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RegisterRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Registration successful, returns access and refresh tokens\" // MODIFIED DESCRIPTION",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "409": {
                        "description": "Conflict - Email Exists",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/uploads/audio/batch/request": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Requests multiple presigned URLs for uploading several audio files in parallel.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Request presigned URLs for batch audio upload",
                "operationId": "request-batch-audio-upload",
                "parameters": [
                    {
                        "description": "List of files to request URLs for",
                        "name": "batchUploadRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.BatchRequestUploadInputRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of generated presigned URLs and object keys, including potential errors per item.",
                        "schema": {
                            "$ref": "#/definitions/dto.BatchRequestUploadInputResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input (e.g., empty file list)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/uploads/audio/request": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Requests a presigned URL from the object storage (MinIO/S3) that can be used by the client to directly upload an audio file.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Request presigned URL for audio upload",
                "operationId": "request-audio-upload",
                "parameters": [
                    {
                        "description": "Upload Request Info (filename, content type)",
                        "name": "uploadRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RequestUploadRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Presigned URL and object key generated",
                        "schema": {
                            "$ref": "#/definitions/dto.RequestUploadResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error (e.g., failed to generate URL)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves the profile information for the currently authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Get current user's profile",
                "operationId": "get-my-profile",
                "responses": {
                    "200": {
                        "description": "User profile retrieved successfully",
                        "schema": {
                            "$ref": "#/definitions/dto.UserResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "User Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/bookmarks": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves a paginated list of bookmarks for the authenticated user, optionally filtered by track ID.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "List user's bookmarks",
                "operationId": "list-bookmarks",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Filter by Audio Track UUID",
                        "name": "trackId",
                        "in": "query"
                    },
                    {
                        "maximum": 100,
                        "minimum": 1,
                        "type": "integer",
                        "default": 50,
                        "description": "Pagination limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Paginated list of bookmarks (timestampMs in milliseconds)",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/dto.PaginatedResponseDTO"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/dto.BookmarkResponseDTO"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Invalid Track ID Format (if provided)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Creates a new bookmark at a specific timestamp (in milliseconds) in an audio track for the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Create a bookmark",
                "operationId": "create-bookmark",
                "parameters": [
                    {
                        "description": "Bookmark details (timestampMs in milliseconds)",
                        "name": "bookmark",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CreateBookmarkRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Bookmark created successfully (timestampMs in milliseconds)",
                        "schema": {
                            "$ref": "#/definitions/dto.BookmarkResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input / Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/bookmarks/{bookmarkId}": {
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Deletes a specific bookmark owned by the current user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Delete a bookmark",
                "operationId": "delete-bookmark",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Bookmark UUID",
                        "name": "bookmarkId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Bookmark deleted successfully"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "403": {
                        "description": "Forbidden (Not Owner)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Bookmark Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/progress": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves a paginated list of playback progress records for the authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "List user's playback progress",
                "operationId": "list-playback-progress",
                "parameters": [
                    {
                        "maximum": 100,
                        "minimum": 1,
                        "type": "integer",
                        "default": 50,
                        "description": "Pagination limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Paginated list of playback progress (progressMs in milliseconds)",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/dto.PaginatedResponseDTO"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/dto.PlaybackProgressResponseDTO"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Records or updates the playback progress for a specific audio track for the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Record playback progress",
                "operationId": "record-playback-progress",
                "parameters": [
                    {
                        "description": "Playback progress details (progressMs in milliseconds)",
                        "name": "progress",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.RecordProgressRequestDTO"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Progress recorded successfully"
                    },
                    "400": {
                        "description": "Invalid Input / Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Track Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        },
        "/users/me/progress/{trackId}": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Retrieves the playback progress for a specific audio track for the authenticated user.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User Activity"
                ],
                "summary": "Get playback progress for a track",
                "operationId": "get-playback-progress",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Audio Track UUID",
                        "name": "trackId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Playback progress found (progressMs in milliseconds)",
                        "schema": {
                            "$ref": "#/definitions/dto.PlaybackProgressResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Track ID Format",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "404": {
                        "description": "Progress Not Found (or Track Not Found)",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.ErrorResponseDTO"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "dto.AudioCollectionResponseDTO": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "ownerId": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "tracks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                    }
                },
                "type": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "dto.AudioTrackDetailsResponseDTO": {
            "type": "object",
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "Point 1: Use milliseconds (int64)",
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "isPublic": {
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "type": "string"
                },
                "playUrl": {
                    "description": "Presigned URL",
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "uploaderId": {
                    "type": "string"
                },
                "userBookmarks": {
                    "description": "Array of user bookmarks for this track",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BookmarkResponseDTO"
                    }
                },
                "userProgressMs": {
                    "description": "Point 1: User progress in ms",
                    "type": "integer"
                }
            }
        },
        "dto.AudioTrackResponseDTO": {
            "type": "object",
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "Point 1: Use milliseconds (int64)",
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "isPublic": {
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "uploaderId": {
                    "type": "string"
                }
            }
        },
        "dto.AuthResponseDTO": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "description": "The JWT access token",
                    "type": "string"
                },
                "isNewUser": {
                    "description": "Pointer, only included for Google callback if user is new",
                    "type": "boolean"
                },
                "refreshToken": {
                    "description": "The refresh token value",
                    "type": "string"
                }
            }
        },
        "dto.BatchCompleteUploadInputDTO": {
            "type": "object",
            "required": [
                "tracks"
            ],
            "properties": {
                "tracks": {
                    "description": "Ensure at least one track, validate each item",
                    "type": "array",
                    "minItems": 1,
                    "items": {
                        "$ref": "#/definitions/dto.BatchCompleteUploadItemDTO"
                    }
                }
            }
        },
        "dto.BatchCompleteUploadItemDTO": {
            "type": "object",
            "required": [
                "durationMs",
                "languageCode",
                "objectKey",
                "title"
            ],
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "type": "integer"
                },
                "isPublic": {
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "type": "string",
                    "enum": [
                        "A1",
                        "A2",
                        "B1",
                        "B2",
                        "C1",
                        "C2",
                        "NATIVE"
                    ]
                },
                "objectKey": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "dto.BatchCompleteUploadResponseDTO": {
            "type": "object",
            "properties": {
                "results": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BatchCompleteUploadResponseItemDTO"
                    }
                }
            }
        },
        "dto.BatchCompleteUploadResponseItemDTO": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "Error message if processing failed for this item",
                    "type": "string"
                },
                "objectKey": {
                    "description": "Identifies the item",
                    "type": "string"
                },
                "success": {
                    "description": "Whether processing this item succeeded",
                    "type": "boolean"
                },
                "trackId": {
                    "description": "The ID of the created track if successful",
                    "type": "string"
                }
            }
        },
        "dto.BatchRequestUploadInputItemDTO": {
            "type": "object",
            "required": [
                "contentType",
                "filename"
            ],
            "properties": {
                "contentType": {
                    "description": "e.g., \"audio/mpeg\"",
                    "type": "string"
                },
                "filename": {
                    "type": "string"
                }
            }
        },
        "dto.BatchRequestUploadInputRequestDTO": {
            "type": "object",
            "required": [
                "files"
            ],
            "properties": {
                "files": {
                    "description": "Ensure at least one file, validate each item",
                    "type": "array",
                    "minItems": 1,
                    "items": {
                        "$ref": "#/definitions/dto.BatchRequestUploadInputItemDTO"
                    }
                }
            }
        },
        "dto.BatchRequestUploadInputResponseDTO": {
            "type": "object",
            "properties": {
                "results": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BatchRequestUploadInputResponseItemDTO"
                    }
                }
            }
        },
        "dto.BatchRequestUploadInputResponseItemDTO": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "Error message if URL generation failed for this item",
                    "type": "string"
                },
                "objectKey": {
                    "description": "The generated object key for this file",
                    "type": "string"
                },
                "originalFilename": {
                    "description": "Helps client match response to request",
                    "type": "string"
                },
                "uploadUrl": {
                    "description": "The presigned PUT URL for this file",
                    "type": "string"
                }
            }
        },
        "dto.BookmarkResponseDTO": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "note": {
                    "type": "string"
                },
                "timestampMs": {
                    "type": "integer"
                },
                "trackId": {
                    "type": "string"
                },
                "userId": {
                    "type": "string"
                }
            }
        },
        "dto.CompleteUploadInputDTO": {
            "type": "object",
            "required": [
                "durationMs",
                "languageCode",
                "objectKey",
                "title"
            ],
            "properties": {
                "coverImageUrl": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "Duration in Milliseconds, must be positive",
                    "type": "integer"
                },
                "isPublic": {
                    "description": "Defaults to false if omitted? Define behavior.",
                    "type": "boolean"
                },
                "languageCode": {
                    "type": "string"
                },
                "level": {
                    "description": "Allow empty or valid level",
                    "type": "string",
                    "enum": [
                        "A1",
                        "A2",
                        "B1",
                        "B2",
                        "C1",
                        "C2",
                        "NATIVE"
                    ]
                },
                "objectKey": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "dto.CreateBookmarkRequestDTO": {
            "type": "object",
            "required": [
                "timestampMs",
                "trackId"
            ],
            "properties": {
                "note": {
                    "type": "string"
                },
                "timestampMs": {
                    "description": "Point 1: Already uses ms",
                    "type": "integer",
                    "minimum": 0
                },
                "trackId": {
                    "type": "string"
                }
            }
        },
        "dto.CreateCollectionRequestDTO": {
            "type": "object",
            "required": [
                "title",
                "type"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "initialTrackIds": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                },
                "type": {
                    "type": "string",
                    "enum": [
                        "COURSE",
                        "PLAYLIST"
                    ]
                }
            }
        },
        "dto.GoogleCallbackRequestDTO": {
            "type": "object",
            "required": [
                "idToken"
            ],
            "properties": {
                "idToken": {
                    "type": "string"
                }
            }
        },
        "dto.LoginRequestDTO": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "dto.LogoutRequestDTO": {
            "type": "object",
            "required": [
                "refreshToken"
            ],
            "properties": {
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "dto.PaginatedResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "The slice of items for the current page (e.g., []AudioTrackResponseDTO)"
                },
                "limit": {
                    "description": "The limit used for this page",
                    "type": "integer"
                },
                "offset": {
                    "description": "The offset used for this page",
                    "type": "integer"
                },
                "page": {
                    "description": "Current page number (1-based)",
                    "type": "integer"
                },
                "total": {
                    "description": "Total number of items matching the query",
                    "type": "integer"
                },
                "totalPages": {
                    "description": "Total number of pages",
                    "type": "integer"
                }
            }
        },
        "dto.PlaybackProgressResponseDTO": {
            "type": "object",
            "properties": {
                "lastListenedAt": {
                    "type": "string"
                },
                "progressMs": {
                    "description": "Point 1: Already uses ms",
                    "type": "integer"
                },
                "trackId": {
                    "type": "string"
                },
                "userId": {
                    "type": "string"
                }
            }
        },
        "dto.RecordProgressRequestDTO": {
            "type": "object",
            "required": [
                "progressMs",
                "trackId"
            ],
            "properties": {
                "progressMs": {
                    "description": "Point 1: Already uses ms",
                    "type": "integer",
                    "minimum": 0
                },
                "trackId": {
                    "type": "string"
                }
            }
        },
        "dto.RefreshRequestDTO": {
            "type": "object",
            "required": [
                "refreshToken"
            ],
            "properties": {
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "dto.RegisterRequestDTO": {
            "type": "object",
            "required": [
                "email",
                "name",
                "password"
            ],
            "properties": {
                "email": {
                    "description": "Add example tag",
                    "type": "string",
                    "example": "user@example.com"
                },
                "name": {
                    "type": "string",
                    "maxLength": 100,
                    "example": "John Doe"
                },
                "password": {
                    "description": "Add format tag",
                    "type": "string",
                    "format": "password",
                    "minLength": 8,
                    "example": "Str0ngP@ssw0rd"
                }
            }
        },
        "dto.RequestUploadRequestDTO": {
            "type": "object",
            "required": [
                "contentType",
                "filename"
            ],
            "properties": {
                "contentType": {
                    "description": "e.g., \"audio/mpeg\"",
                    "type": "string"
                },
                "filename": {
                    "type": "string"
                }
            }
        },
        "dto.RequestUploadResponseDTO": {
            "type": "object",
            "properties": {
                "objectKey": {
                    "description": "The key the client should use/report back",
                    "type": "string"
                },
                "uploadUrl": {
                    "description": "The presigned PUT URL",
                    "type": "string"
                }
            }
        },
        "dto.UpdateCollectionRequestDTO": {
            "type": "object",
            "required": [
                "title"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "title": {
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "dto.UpdateCollectionTracksRequestDTO": {
            "type": "object",
            "properties": {
                "orderedTrackIds": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "dto.UserResponseDTO": {
            "type": "object",
            "properties": {
                "authProvider": {
                    "type": "string"
                },
                "createdAt": {
                    "description": "Use string format like RFC3339",
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "profileImageUrl": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "httputil.ErrorResponseDTO": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "Application-specific error code (e.g., \"INVALID_INPUT\", \"NOT_FOUND\")",
                    "type": "string"
                },
                "message": {
                    "description": "User-friendly error message",
                    "type": "string"
                },
                "requestId": {
                    "description": "Include request ID for tracing",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Type \"Bearer\" followed by a space and JWT token. Example: \"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...\"",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    },
    "tags": [
        {
            "description": "Operations related to user signup, login, and external authentication (e.g., Google).",
            "name": "Authentication"
        },
        {
            "description": "Operations related to user profiles.",
            "name": "Users"
        },
        {
            "description": "Operations related to individual audio tracks, including retrieval and listing.",
            "name": "Audio Tracks"
        },
        {
            "description": "Operations related to managing audio collections (playlists, courses).",
            "name": "Audio Collections"
        },
        {
            "description": "Operations related to tracking user interactions like playback progress and bookmarks.",
            "name": "User Activity"
        },
        {
            "description": "Operations related to requesting upload URLs and finalizing uploads.",
            "name": "Uploads"
        },
        {
            "description": "API health checks.",
            "name": "Health"
        }
    ]
}
```

## `pkg/httputil/response.go`

```go
// ============================================
// FILE: pkg/httputil/response.go (MODIFIED)
// ============================================
package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/pkg/apierrors"   // Import the new package
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const RequestIDKey ContextKey = "requestID"

// GetReqID retrieves the request ID from the context.
// Returns an empty string if not found.
func GetReqID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// ErrorResponseDTO defines the standard JSON error response body.
type ErrorResponseDTO struct {
	Code      string `json:"code"`                // Application-specific error code (e.g., "INVALID_INPUT", "NOT_FOUND")
	Message   string `json:"message"`             // User-friendly error message
	RequestID string `json:"requestId,omitempty"` // Include request ID for tracing
	// Details   interface{} `json:"details,omitempty"` // Optional: For more detailed error info (e.g., validation fields)
}

// RespondJSON writes a JSON response with the given status code and payload.
func RespondJSON(w http.ResponseWriter, r *http.Request, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload != nil {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			// Log error, but can't write header again
			logger := slog.Default()
			reqID := GetReqID(r.Context()) // Get request ID from context
			logger.ErrorContext(r.Context(), "Failed to encode JSON response", "error", err, "status", status, "request_id", reqID)
			// Attempt to write a plain text error if JSON encoding fails *after* header was written
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// MapDomainErrorToHTTP maps domain errors to HTTP status codes and error codes.
// This is a central place to define the mapping.
func MapDomainErrorToHTTP(err error) (status int, code string, message string) {
	if err == nil {
		return http.StatusOK, "", "" // No error, return OK status
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		// Use constant from apierrors package
		return http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found."
	case errors.Is(err, domain.ErrConflict):
		// Use constant from apierrors package
		return http.StatusConflict, apierrors.CodeConflict, "A conflict occurred with the current state of the resource."
	case errors.Is(err, domain.ErrInvalidArgument):
		// Use constant from apierrors package
		// Use the specific error message for validation details
		return http.StatusBadRequest, apierrors.CodeInvalidInput, err.Error()
	case errors.Is(err, domain.ErrPermissionDenied):
		// Use constant from apierrors package
		// Check for specific rate limit error message if needed, otherwise use generic forbidden
		if strings.Contains(err.Error(), "rate limit exceeded") {
			return http.StatusTooManyRequests, apierrors.CodeRateLimitExceeded, "Too many requests. Please try again later."
		}
		return http.StatusForbidden, apierrors.CodeForbidden, "You do not have permission to perform this action."
	case errors.Is(err, domain.ErrAuthenticationFailed):
		// Use constant from apierrors package
		return http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication failed. Please check your credentials."
	case errors.Is(err, domain.ErrUnauthenticated):
		// Use constant from apierrors package
		return http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication required. Please log in."
	default:
		// Any other error is treated as an internal server error
		// Use constant from apierrors package
		return http.StatusInternalServerError, apierrors.CodeInternalError, "An unexpected internal error occurred."
	}
}

// RespondError maps a domain error to an HTTP status code and JSON error response.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := MapDomainErrorToHTTP(err)

	reqID := GetReqID(r.Context()) // Get request ID from context
	logger := slog.Default()

	// Log internal server errors with more detail
	if status >= 500 {
		logger.ErrorContext(r.Context(), "Internal server error occurred", "error", err, "status", status, "code", code, "request_id", reqID)
		// Avoid leaking internal error details in the response message for 500 errors
		message = "An unexpected internal error occurred."
	} else {
		// Log client errors (4xx) at a lower level, e.g., Warn or Info
		logger.WarnContext(r.Context(), "Client error occurred", "error", err, "status", status, "code", code, "request_id", reqID)
	}

	errorResponse := ErrorResponseDTO{
		Code:      code,
		Message:   message,
		RequestID: reqID,
	}

	RespondJSON(w, r, status, errorResponse)
}
```

## `pkg/httputil/response_test.go`

```go
package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/pkg/apierrors"   // Adjust import path
)

func TestGetReqID(t *testing.T) {
	reqID := "test-request-123"
	ctxWithID := context.WithValue(context.Background(), RequestIDKey, reqID)
	ctxWithoutID := context.Background()

	// Case 1: Context has ID
	retrievedID := GetReqID(ctxWithID)
	assert.Equal(t, reqID, retrievedID)

	// Case 2: Context does not have ID
	retrievedID = GetReqID(ctxWithoutID)
	assert.Empty(t, retrievedID)

	// Case 3: Context has wrong type
	ctxWithWrongType := context.WithValue(context.Background(), RequestIDKey, 123)
	retrievedID = GetReqID(ctxWithWrongType)
	assert.Empty(t, retrievedID)
}

func TestRespondJSON(t *testing.T) {
	type samplePayload struct {
		Message string `json:"message"`
		Value   int    `json:"value"`
	}
	payload := samplePayload{Message: "Success", Value: 100}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	RespondJSON(rr, req, http.StatusOK, payload)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var responsePayload samplePayload
	err := json.Unmarshal(rr.Body.Bytes(), &responsePayload)
	assert.NoError(t, err)
	assert.Equal(t, payload, responsePayload)
}

func TestRespondJSON_NilPayload(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	RespondJSON(rr, req, http.StatusNoContent, nil)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type")) // Header is still set
	assert.Empty(t, rr.Body.String())
}

// TestRespondJSON_EncodeError requires a type that cannot be marshaled
type unmarshalable struct {
	FuncField func()
}

func TestRespondJSON_EncodeError(t *testing.T) {
	// This test might be flaky depending on internal logging, but tests the principle
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	payload := unmarshalable{FuncField: func() {}}

	RespondJSON(rr, req, http.StatusOK, payload)

	// Expecting 500 because encoding fails after header is written
	assert.Equal(t, http.StatusOK, rr.Code) // Header was already written
	// Body might contain the plain text error or be empty depending on http internals
	assert.Contains(t, rr.Body.String(), "Internal Server Error")
}

func TestMapDomainErrorToHTTP(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string // Base message before potential wrapping
	}{
		{"Not Found", domain.ErrNotFound, http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found."},
		{"Conflict", domain.ErrConflict, http.StatusConflict, apierrors.CodeConflict, "A conflict occurred with the current state of the resource."},
		{"Invalid Argument", fmt.Errorf("%w: email is required", domain.ErrInvalidArgument), http.StatusBadRequest, apierrors.CodeInvalidInput, "email is required"}, // Specific message used
		{"Permission Denied", domain.ErrPermissionDenied, http.StatusForbidden, apierrors.CodeForbidden, "You do not have permission to perform this action."},
		{"Rate Limit", fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied), http.StatusTooManyRequests, apierrors.CodeRateLimitExceeded, "Too many requests. Please try again later."},
		{"Authentication Failed", domain.ErrAuthenticationFailed, http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication failed. Please check your credentials."},
		{"Unauthenticated", domain.ErrUnauthenticated, http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication required. Please log in."},
		{"Wrapped Not Found", fmt.Errorf("specific item not found: %w", domain.ErrNotFound), http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found."},
		{"Generic Error", errors.New("some generic error"), http.StatusInternalServerError, apierrors.CodeInternalError, "An unexpected internal error occurred."},
		{"Nil Error", nil, http.StatusOK, "", ""}, // Should maybe not be called with nil, but handle gracefully
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// REMOVED the unused 'message' variable declaration and assignment here

			gotStatus, gotCode, gotMessage := MapDomainErrorToHTTP(tt.err)
			assert.Equal(t, tt.wantStatus, gotStatus)
			assert.Equal(t, tt.wantCode, gotCode)
			// Special check for InvalidArgument message
			if errors.Is(tt.err, domain.ErrInvalidArgument) {
				// The MapDomainErrorToHTTP function should return the full error string for ErrInvalidArgument
				assert.Equal(t, tt.err.Error(), gotMessage)
			} else {
				assert.Equal(t, tt.wantMessage, gotMessage)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedCode    string
		expectedMessage string // Expected message IN THE RESPONSE BODY
		includeReqId    bool
	}{
		{"Not Found Error", domain.ErrNotFound, http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found.", true},
		{"Invalid Argument Error", fmt.Errorf("%w: missing field 'name'", domain.ErrInvalidArgument), http.StatusBadRequest, apierrors.CodeInvalidInput, "invalid argument: missing field 'name'", true},
		{"Internal Server Error", errors.New("database connection failed"), http.StatusInternalServerError, apierrors.CodeInternalError, "An unexpected internal error occurred.", true}, // Message overwritten for 500
		{"Permission Denied Error", domain.ErrPermissionDenied, http.StatusForbidden, apierrors.CodeForbidden, "You do not have permission to perform this action.", false},              // Test without reqId in context
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/error", nil)
			reqId := ""
			if tt.includeReqId {
				reqId = "req-" + tt.name
				ctx := context.WithValue(req.Context(), RequestIDKey, reqId)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			RespondError(rr, req, tt.err)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var errResp ErrorResponseDTO
			err := json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCode, errResp.Code)
			assert.Equal(t, tt.expectedMessage, errResp.Message)
			if tt.includeReqId {
				assert.Equal(t, reqId, errResp.RequestID)
			} else {
				assert.Empty(t, errResp.RequestID)
			}
		})
	}
}
```

## `pkg/pagination/pagination_test.go`

```go
// ============================================
// FILE: pkg/pagination/pagination_test.go
// ============================================

package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPage(t *testing.T) {
	tests := []struct {
		name         string
		pageNum      int
		pageSize     int
		expectedPage Page
	}{
		{"Valid page 1", 1, 15, Page{Limit: 15, Offset: 0}},
		{"Valid page 3", 3, 10, Page{Limit: 10, Offset: 20}},
		{"Zero page size", 2, 0, Page{Limit: DefaultLimit, Offset: DefaultLimit}},      // Uses default limit for offset calculation
		{"Negative page size", 2, -5, Page{Limit: DefaultLimit, Offset: DefaultLimit}}, // Uses default limit
		{"Page size exceeds max", 1, MaxLimit + 10, Page{Limit: MaxLimit, Offset: 0}},
		{"Page number 0", 0, 20, Page{Limit: 20, Offset: 0}},         // Treated as page 1
		{"Page number negative", -1, 20, Page{Limit: 20, Offset: 0}}, // Treated as page 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPage(tt.pageNum, tt.pageSize)
			assert.Equal(t, tt.expectedPage, p)
		})
	}
}

func TestNewPageFromOffset(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		offset       int
		expectedPage Page
	}{
		{"Valid limit and offset", 15, 30, Page{Limit: 15, Offset: 30}},
		{"Zero limit", 0, 50, Page{Limit: DefaultLimit, Offset: 50}},
		{"Negative limit", -10, 50, Page{Limit: DefaultLimit, Offset: 50}},
		{"Limit exceeds max", MaxLimit + 5, 10, Page{Limit: MaxLimit, Offset: 10}},
		{"Zero offset", 25, 0, Page{Limit: 25, Offset: 0}},
		{"Negative offset", 25, -10, Page{Limit: 25, Offset: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPageFromOffset(tt.limit, tt.offset)
			assert.Equal(t, tt.expectedPage, p)
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	type sampleData struct {
		ID int
	}
	dataPage1 := []sampleData{{ID: 1}, {ID: 2}}
	dataPageLast := []sampleData{{ID: 5}}

	tests := []struct {
		name        string
		data        interface{}
		total       int
		pageParams  Page
		expectedRes PaginatedResponse
	}{
		{
			name:       "First page",
			data:       dataPage1,
			total:      5,
			pageParams: Page{Limit: 2, Offset: 0},
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      5,
				Limit:      2,
				Offset:     0,
				Page:       1, // (0 / 2) + 1
				TotalPages: 3, // ceil(5 / 2)
			},
		},
		{
			name:       "Middle page",
			data:       dataPage1, // Assuming data for page 2 is []sampleData{{ID: 3}, {ID: 4}}
			total:      5,
			pageParams: Page{Limit: 2, Offset: 2},
			expectedRes: PaginatedResponse{
				Data:       dataPage1, // Pass the actual data for the page being tested
				Total:      5,
				Limit:      2,
				Offset:     2,
				Page:       2, // (2 / 2) + 1
				TotalPages: 3,
			},
		},
		{
			name:       "Last page",
			data:       dataPageLast,
			total:      5,
			pageParams: Page{Limit: 2, Offset: 4},
			expectedRes: PaginatedResponse{
				Data:       dataPageLast,
				Total:      5,
				Limit:      2,
				Offset:     4,
				Page:       3, // (4 / 2) + 1
				TotalPages: 3,
			},
		},
		{
			name:       "Total items less than limit",
			data:       dataPage1,
			total:      2,
			pageParams: Page{Limit: 10, Offset: 0},
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      2,
				Limit:      10,
				Offset:     0,
				Page:       1,
				TotalPages: 1, // ceil(2 / 10)
			},
		},
		{
			name:       "Zero total items",
			data:       []sampleData{},
			total:      0,
			pageParams: Page{Limit: 10, Offset: 0},
			expectedRes: PaginatedResponse{
				Data:       []sampleData{},
				Total:      0,
				Limit:      10,
				Offset:     0,
				Page:       1,
				TotalPages: 0,
			},
		},
		{
			name:       "Invalid page params (default limit)",
			data:       dataPage1,
			total:      5,
			pageParams: Page{Limit: 0, Offset: 0}, // Should use DefaultLimit
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      5,
				Limit:      DefaultLimit,
				Offset:     0,
				Page:       1,
				TotalPages: 1, // ceil(5 / DefaultLimit)
			},
		},
		{
			name:       "Invalid page params (max limit)",
			data:       dataPage1,
			total:      200,
			pageParams: Page{Limit: MaxLimit + 10, Offset: 0}, // Should use MaxLimit
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      200,
				Limit:      MaxLimit,
				Offset:     0,
				Page:       1,
				TotalPages: 2, // ceil(200 / MaxLimit)
			},
		},
		{
			name:       "Invalid page params (negative offset)",
			data:       dataPage1,
			total:      5,
			pageParams: Page{Limit: 10, Offset: -10}, // Should use Offset 0
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      5,
				Limit:      10,
				Offset:     0,
				Page:       1, // (0 / 10) + 1
				TotalPages: 1, // ceil(5 / 10)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewPaginatedResponse(tt.data, tt.total, tt.pageParams)
			assert.Equal(t, tt.expectedRes, res)
		})
	}
}
```

## `pkg/pagination/pagination.go`

```go
// pkg/pagination/pagination.go
package pagination

import (
	"math"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Page represents pagination parameters from a request.
type Page struct {
	Limit  int // Number of items per page
	Offset int // Number of items to skip
}

// NewPage creates a Page struct from limit and offset inputs, applying defaults and constraints.
// pageNum is 1-based, pageSize is the number of items per page.
func NewPage(pageNum, pageSize int) Page {
	limit := pageSize
	if limit <= 0 {
		limit = DefaultLimit // Apply default limit if invalid or zero
	}
	if limit > MaxLimit {
		limit = MaxLimit // Apply max limit constraint
	}

	offset := 0
	if pageNum > 1 {
		offset = (pageNum - 1) * limit // Calculate offset based on 1-based page number
	}
	// Ensure offset is not negative (though calculation above shouldn't make it negative if pageNum >= 1)
	if offset < 0 {
		offset = 0
	}


	return Page{
		Limit:  limit,
		Offset: offset,
	}
}

// NewPageFromOffset creates a Page struct directly from limit and offset, applying defaults and constraints.
func NewPageFromOffset(limit, offset int) Page {
    pLimit := limit
	if pLimit <= 0 {
		pLimit = DefaultLimit
	}
	if pLimit > MaxLimit {
		pLimit = MaxLimit
	}

    pOffset := offset
    if pOffset < 0 {
        pOffset = 0
    }

	return Page{
		Limit:  pLimit,
		Offset: pOffset,
	}
}


// PaginatedResponse defines the standard structure for paginated API responses.
// The Data field should hold a slice of the actual response DTOs (e.g., []dto.AudioTrackResponseDTO).
type PaginatedResponse struct {
	Data       interface{} `json:"data"`        // The slice of items for the current page
	Total      int         `json:"total"`       // Total number of items matching the query
	Limit      int         `json:"limit"`       // The limit used for this page
	Offset     int         `json:"offset"`      // The offset used for this page
	Page       int         `json:"page"`        // Current page number (1-based)
	TotalPages int         `json:"totalPages"` // Total number of pages
}

// NewPaginatedResponse creates a PaginatedResponse struct.
// It calculates Page and TotalPages based on the provided data.
// - data: The slice of items for the current page (e.g., []*domain.Track).
// - total: The total number of items available across all pages.
// - pageParams: The Page struct (Limit, Offset) used to fetch this data.
func NewPaginatedResponse(data interface{}, total int, pageParams Page) PaginatedResponse {
	// Ensure limit and offset from pageParams are valid (apply constraints again just in case)
	limit := pageParams.Limit
	if limit <= 0 {
        limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
    offset := pageParams.Offset
    if offset < 0 {
        offset = 0
    }

	totalPages := 0
	if total > 0 && limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

    currentPage := 1
    if limit > 0 {
        currentPage = (offset / limit) + 1
    }


	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Limit:      limit, // Use the constrained limit
		Offset:     offset, // Use the constrained offset
		Page:       currentPage,
		TotalPages: totalPages,
	}
}
```

## `pkg/logger/logger.go`

```go
// pkg/logger/logger.go
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/yvanyang/language-learning-player-api/internal/config" // Adjust import path
)

// NewLogger creates and returns a new slog.Logger based on the provided configuration.
func NewLogger(cfg config.LogConfig) *slog.Logger {
	var logWriter io.Writer = os.Stdout // Default to standard output

	// Determine log level
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo // Default to Info level if unspecified or invalid
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// AddSource: true, // Uncomment this to include source file and line number in logs (can affect performance)
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(logWriter, opts)
	} else {
		handler = slog.NewTextHandler(logWriter, opts)
	}

	logger := slog.New(handler)
	return logger
}
```

## `pkg/security/security_test.go`

```go
package security

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

// Use the same test secret as jwt_test.go for consistency
const testSecurityJwtSecret = "test-super-secret-key-for-unit-testing"

func TestSecurity_Integration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sec, err := NewSecurity(testSecurityJwtSecret, logger)
	assert.NoError(t, err)
	assert.NotNil(t, sec)

	ctx := context.Background()
	password := "verySecureP@ssw0rd"
	userID := domain.NewUserID()
	duration := 5 * time.Minute

	// 1. Hash Password
	hashedPassword, err := sec.HashPassword(ctx, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// 2. Check Correct Password
	match := sec.CheckPasswordHash(ctx, password, hashedPassword)
	assert.True(t, match)

	// 3. Check Incorrect Password
	match = sec.CheckPasswordHash(ctx, "wrongPassword", hashedPassword)
	assert.False(t, match)

	// 4. Generate JWT
	tokenString, err := sec.GenerateJWT(ctx, userID, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 5. Verify JWT
	parsedUserID, err := sec.VerifyJWT(ctx, tokenString)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)

	// 6. Generate Refresh Token Value
	refreshTokenVal, err := sec.GenerateRefreshTokenValue()
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshTokenVal)
	assert.Len(t, refreshTokenVal, 44) // 32 bytes -> 44 base64 chars

	// 7. Hash Refresh Token Value
	refreshTokenHash := sec.HashRefreshTokenValue(refreshTokenVal)
	assert.NotEmpty(t, refreshTokenHash)
	assert.Len(t, refreshTokenHash, 64) // SHA-256 hash is 64 hex chars

	// Ensure hashing is deterministic
	refreshTokenHash2 := sec.HashRefreshTokenValue(refreshTokenVal)
	assert.Equal(t, refreshTokenHash, refreshTokenHash2)

	// Ensure different values produce different hashes
	refreshTokenVal2, _ := sec.GenerateRefreshTokenValue()
	refreshTokenHash3 := sec.HashRefreshTokenValue(refreshTokenVal2)
	assert.NotEqual(t, refreshTokenHash, refreshTokenHash3)
}

func TestNewSecurity_EmptySecret(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewSecurity("", logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}
```

## `pkg/security/hasher_test.go`

```go
// ============================================
// FILE: pkg/security/hasher_test.go
// ============================================

package security

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher_HashPassword(t *testing.T) {
	hasher := NewBcryptHasher(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	password := "mysecretpassword"

	hash, err := hasher.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify the hash cost matches the default
	cost, err := bcrypt.Cost([]byte(hash))
	assert.NoError(t, err)
	assert.Equal(t, defaultPasswordCost, cost)

	// Hashing the same password again should produce a different hash (due to salt)
	hash2, err2 := hasher.HashPassword(password)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash, hash2)
}

func TestBcryptHasher_CheckPasswordHash(t *testing.T) {
	hasher := NewBcryptHasher(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	password := "mysecretpassword"
	wrongPassword := "wrongpassword"

	hash, err := hasher.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Check correct password
	match := hasher.CheckPasswordHash(password, hash)
	assert.True(t, match)

	// Check incorrect password
	match = hasher.CheckPasswordHash(wrongPassword, hash)
	assert.False(t, match)

	// Check with invalid hash format (should return false and log warning)
	match = hasher.CheckPasswordHash(password, "invalid-hash-format")
	assert.False(t, match)
}

func TestSha256Hash(t *testing.T) {
	input := "this is a test string"
	// 使用标准库直接计算预期的哈希值，确保正确性
	h := sha256.New()
	h.Write([]byte(input))
	expectedHashBytes := h.Sum(nil)
	expectedHash := hex.EncodeToString(expectedHashBytes)
	// 正确的 SHA-256 哈希值应该是：c7be1ed902fb8dd4d48997c6452f5d7e509fbcdbe2808b16bcf4edce4c07d14e

	// 调用被测试的函数
	hash := Sha256Hash(input)
	// 断言生成的哈希值与预期值相等
	assert.Equal(t, expectedHash, hash, "Input: '%s'", input) // 添加消息以便调试

	// 测试空字符串
	emptyInput := ""
	hEmpty := sha256.New()
	hEmpty.Write([]byte(emptyInput))
	expectedEmptyHashBytes := hEmpty.Sum(nil)
	expectedEmptyHash := hex.EncodeToString(expectedEmptyHashBytes)

	emptyHash := Sha256Hash(emptyInput)
	assert.Equal(t, expectedEmptyHash, emptyHash, "Input: '%s'", emptyInput)

	// 测试再次哈希相同值得到相同结果 (确定性)
	hashAgain := Sha256Hash(input)
	assert.Equal(t, hash, hashAgain, "Hashing same input again should yield same hash")

	// 测试不同输入产生不同哈希
	differentInput := "this is another test string"
	differentHash := Sha256Hash(differentInput)
	assert.NotEqual(t, hash, differentHash, "Hashing different inputs should yield different hashes")
}
```

## `pkg/security/security.go`

```go
// ============================================
// FILE: pkg/security/security.go (MODIFIED)
// ============================================
package security

import (
	"context"
	"crypto/rand"     // ADDED
	"encoding/base64" // ADDED
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// Security implements the port.SecurityHelper interface.
type Security struct {
	hasher *BcryptHasher
	jwt    *JWTHelper
	logger *slog.Logger
}

// NewSecurity creates a new Security instance.
func NewSecurity(jwtSecretKey string, logger *slog.Logger) (*Security, error) {
	hasher := NewBcryptHasher(logger)
	jwtHelper, err := NewJWTHelper(jwtSecretKey, logger)
	if err != nil {
		return nil, err
	}
	return &Security{
		hasher: hasher,
		jwt:    jwtHelper,
		logger: logger.With("service", "SecurityHelper"),
	}, nil
}

// HashPassword generates a secure hash of the password.
func (s *Security) HashPassword(ctx context.Context, password string) (string, error) {
	return s.hasher.HashPassword(password)
}

// CheckPasswordHash compares a plain password with a stored hash.
func (s *Security) CheckPasswordHash(ctx context.Context, password, hash string) bool {
	return s.hasher.CheckPasswordHash(password, hash)
}

// GenerateJWT creates a signed JWT (Access Token) for the given user ID.
func (s *Security) GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error) {
	return s.jwt.GenerateJWT(userID, duration)
}

// VerifyJWT validates a JWT string and returns the UserID contained within.
func (s *Security) VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error) {
	return s.jwt.VerifyJWT(tokenString)
}

// GenerateRefreshTokenValue creates a cryptographically secure random string for the refresh token.
// ADDED METHOD
func (s *Security) GenerateRefreshTokenValue() (string, error) {
	numBytes := 32 // Generate a 32-byte random token -> 44 Base64 chars
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		s.logger.Error("Failed to generate random bytes for refresh token", "error", err)
		return "", fmt.Errorf("failed to generate refresh token value: %w", err)
	}
	// Use URL-safe base64 encoding
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashRefreshTokenValue generates a SHA-256 hash of the refresh token value for storage.
// ADDED METHOD
func (s *Security) HashRefreshTokenValue(tokenValue string) string {
	return Sha256Hash(tokenValue) // Use the helper from hasher.go
}

// Compile-time check to ensure Security satisfies the port.SecurityHelper interface
// ADDED: New methods to SecurityHelper interface in port/service.go
var _ port.SecurityHelper = (*Security)(nil)
```

## `pkg/security/jwt_test.go`

```go
package security

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

const testJwtSecret = "test-super-secret-key-for-unit-testing"

func TestJWTHelper_GenerateAndVerifyJWT(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	helper, err := NewJWTHelper(testJwtSecret, logger)
	assert.NoError(t, err)

	userID := domain.NewUserID()
	duration := 15 * time.Minute

	tokenString, err := helper.GenerateJWT(userID, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify the generated token
	parsedUserID, err := helper.VerifyJWT(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)

	// Try verifying with a tampered token (invalid signature)
	tamperedToken := tokenString + "tamper"
	_, err = helper.VerifyJWT(tamperedToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed) // VerifyJWT maps errors

	// Try verifying an expired token
	shortDuration := -5 * time.Minute // Expired 5 minutes ago
	expiredTokenString, err := helper.GenerateJWT(userID, shortDuration)
	assert.NoError(t, err)

	// Wait a tiny bit to ensure expiry check works reliably
	time.Sleep(50 * time.Millisecond)

	_, err = helper.VerifyJWT(expiredTokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed) // VerifyJWT maps errors
	assert.Contains(t, err.Error(), "token expired")       // Check underlying reason if needed

	// Try verifying with a different secret key
	wrongHelper, _ := NewJWTHelper("wrong-secret", logger)
	_, err = wrongHelper.VerifyJWT(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)

	// Try verifying a malformed token
	malformedToken := "this.is.not.a.jwt"
	_, err = helper.VerifyJWT(malformedToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	assert.Contains(t, err.Error(), "malformed token")
}

func TestJWTHelper_VerifyJWT_InvalidUserIDFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	helper, err := NewJWTHelper(testJwtSecret, logger)
	assert.NoError(t, err)

	// Generate a token with a non-UUID string in the UserID claim
	claims := &Claims{
		UserID: "not-a-valid-uuid",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "language-learning-player",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(helper.secretKey)
	assert.NoError(t, err)

	// Try to verify it
	_, err = helper.VerifyJWT(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	assert.Contains(t, err.Error(), "invalid user ID format in token")
}

func TestNewJWTHelper_EmptySecret(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewJWTHelper("", logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}
```

## `pkg/security/jwt.go`

```go
// pkg/security/jwt.go
package security

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

// JWTHelper provides JWT generation and verification functionality.
type JWTHelper struct {
	secretKey []byte // Store as byte slice for JWT library
	logger    *slog.Logger
}

// Claims defines the structure of the JWT claims used in this application.
type Claims struct {
	UserID string `json:"uid"` // Store UserID as string in JWT
	jwt.RegisteredClaims
}

// NewJWTHelper creates a new JWTHelper.
func NewJWTHelper(secretKey string, logger *slog.Logger) (*JWTHelper, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}
	return &JWTHelper{
		secretKey: []byte(secretKey),
		logger:    logger.With("component", "JWTHelper"),
	}, nil
}

// GenerateJWT creates a new JWT token for the given user ID and duration.
func (h *JWTHelper) GenerateJWT(userID domain.UserID, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &Claims{
		UserID: userID.String(), // Convert UserID (UUID) to string
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "language-learning-player", // Optional: Identify issuer
			// Subject: userID.String(), // Can also use Subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secretKey)
	if err != nil {
		h.logger.Error("Error signing JWT token", "error", err, "userID", userID.String())
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// VerifyJWT validates the token string and returns the UserID.
func (h *JWTHelper) VerifyJWT(tokenString string) (domain.UserID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			h.logger.Warn("JWT token has expired", "error", err)
			return domain.UserID{}, fmt.Errorf("%w: token expired", domain.ErrAuthenticationFailed)
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			h.logger.Warn("Malformed JWT token received", "error", err)
			return domain.UserID{}, fmt.Errorf("%w: malformed token", domain.ErrAuthenticationFailed)
		}
		// Handle other errors like ErrSignatureInvalid, ErrTokenNotValidYet etc.
		h.logger.Warn("JWT token validation failed", "error", err)
		return domain.UserID{}, fmt.Errorf("%w: %v", domain.ErrAuthenticationFailed, err)
	}

	if !token.Valid {
		h.logger.Warn("Invalid JWT token received")
		return domain.UserID{}, domain.ErrAuthenticationFailed // General invalid token
	}

	// Convert UserID string from claim back to domain.UserID (UUID)
	userID, parseErr := domain.UserIDFromString(claims.UserID)
	if parseErr != nil {
		h.logger.Error("Error parsing UserID from valid JWT claims", "error", parseErr, "claimUserID", claims.UserID)
		return domain.UserID{}, fmt.Errorf("%w: invalid user ID format in token", domain.ErrAuthenticationFailed)
	}

	return userID, nil
}
```

## `pkg/security/hasher.go`

```go
// ============================================
// FILE: pkg/security/hasher.go (MODIFIED)
// ============================================
package security

import (
	"crypto/sha256" // ADDED
	"encoding/hex"  // ADDED
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

const defaultPasswordCost = 12 // Adjust cost based on your security needs and performance

// BcryptHasher provides password hashing using bcrypt.
type BcryptHasher struct {
	cost   int
	logger *slog.Logger
}

// NewBcryptHasher creates a new BcryptHasher.
func NewBcryptHasher(logger *slog.Logger) *BcryptHasher {
	return &BcryptHasher{
		cost:   defaultPasswordCost,
		logger: logger.With("component", "BcryptHasher"),
	}
}

// HashPassword generates a bcrypt hash for the given password.
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		h.logger.Error("Error generating password hash", "error", err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash.
func (h *BcryptHasher) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			h.logger.Debug("Password hash mismatch", "error", err)
		} else {
			h.logger.Warn("Error comparing password hash", "error", err)
		}
		return false
	}
	return true
}

// Sha256Hash generates a SHA-256 hash (hex encoded) for non-password secrets like refresh tokens.
// ADDED FUNCTION
func Sha256Hash(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value)) // Hash the string value
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes) // Return hex string representation
}
```

## `pkg/apierrors/codes.go`

```go
// ============================================
// FILE: pkg/apierrors/codes.go (NEW FILE)
// ============================================
package apierrors

// Standard API Error Codes
const (
	CodeNotFound          = "NOT_FOUND"
	CodeConflict          = "RESOURCE_CONFLICT"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeForbidden         = "FORBIDDEN"
	CodeUnauthenticated   = "UNAUTHENTICATED"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED" // Added for potential use
)
```

## `pkg/validation/validator.go`

```go
// pkg/validation/validator.go
package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator.v10 instance.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	validate := validator.New()

	// Register a function to get the 'json' tag name for field names in errors.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// TODO: Register custom validation rules here if needed
	// Example: validate.RegisterValidation("customRule", customValidationFunc)

	return &Validator{validate: validate}
}

// ValidateStruct validates the given struct based on 'validate' tags.
// Returns a formatted error string if validation fails, otherwise nil.
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		// Translate validation errors into a user-friendly message
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			// Use the 'json' tag name if available, otherwise use the field name
			fieldName := err.Field()
			// Construct a message based on the validation tag
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation on the '%s' rule", fieldName, err.Tag()))
			// Alternative: More user-friendly messages based on tag
			// msg := fmt.Sprintf("Invalid value for field %s (%s rule)", fieldName, err.Tag())
			// errorMessages = append(errorMessages, msg)
		}
		// Return a single error string joining all messages
		// Prepend with domain.ErrInvalidArgument? Maybe not, let handler decide mapping.
		return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
		// Alternative: return the original validator.ValidationErrors if handler needs more detail
		// return err
	}
	return nil
}

// TODO: Add ValidateVariable for single field validation if needed.
// func (v *Validator) ValidateVariable(i interface{}, tag string) error { ... }
```

## `pkg/validation/validator_test.go`

```go
package validation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	RequiredField string    `json:"reqField" validate:"required"`
	EmailField    string    `json:"emailField" validate:"required,email"`
	UUIDField     string    `json:"uuidField" validate:"required,uuid"`
	MinLenField   string    `json:"minLenField" validate:"required,min=5"`
	MaxLenField   string    `json:"maxLenField" validate:"required,max=10"`
	OneOfField    string    `json:"oneOfField" validate:"required,oneof=A B C"`
	OptionalField *string   `json:"optField" validate:"omitempty,min=3"` // Optional, but min length if present
	DiveField     []SubTest `json:"diveField" validate:"required,dive"`
}

type SubTest struct {
	SubField string `json:"subField" validate:"required,alpha"`
}

func TestValidator_ValidateStruct(t *testing.T) {
	v := New()
	validUUID := uuid.NewString()

	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
		errSubstr []string // Substrings expected in the error message
	}{
		{
			name: "Valid struct",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: false,
		},
		{
			name: "Missing required field",
			input: &TestStruct{
				// RequiredField missing
				EmailField:  "test@example.com",
				UUIDField:   validUUID,
				MinLenField: "12345",
				MaxLenField: "1234567890",
				OneOfField:  "B",
				DiveField:   []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'reqField' failed validation on the 'required' rule"},
		},
		{
			name: "Invalid email format",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "not-an-email", // Invalid
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "C",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'emailField' failed validation on the 'email' rule"},
		},
		{
			name: "Invalid UUID format",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     "not-a-uuid", // Invalid
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'uuidField' failed validation on the 'uuid' rule"},
		},
		{
			name: "Min length failure",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "123", // Too short
				MaxLenField:   "1234567890",
				OneOfField:    "B",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'minLenField' failed validation on the 'min' rule"},
		},
		{
			name: "Max length failure",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "12345678901", // Too long
				OneOfField:    "C",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'maxLenField' failed validation on the 'max' rule"},
		},
		{
			name: "OneOf failure",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "D", // Not in A, B, C
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'oneOfField' failed validation on the 'oneof' rule"},
		},
		{
			name: "Valid optional field (present)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				OptionalField: func() *string { s := "abc"; return &s }(), // Valid length
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: false,
		},
		{
			name: "Valid optional field (absent)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "B",
				OptionalField: nil, // Absent, should pass omitempty
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: false,
		},
		{
			name: "Invalid optional field (present but too short)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "C",
				OptionalField: func() *string { s := "ab"; return &s }(), // Too short
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'optField' failed validation on the 'min' rule"},
		},
		{
			name: "Dive validation failure (sub-struct)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				DiveField:     []SubTest{{SubField: "Not1Alpha"}}, // Invalid sub-field
			},
			expectErr: true,
			errSubstr: []string{"'subField' failed validation on the 'alpha' rule"}, // Note: Field name comes from sub-struct
		},
		{
			name: "Multiple errors",
			input: &TestStruct{
				// RequiredField missing
				EmailField:  "not-an-email",
				UUIDField:   "not-uuid",
				MinLenField: "123",
				MaxLenField: "12345678901",
				OneOfField:  "D",
				DiveField:   []SubTest{{SubField: "Not1Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{
				"'reqField' failed validation on the 'required' rule",
				"'emailField' failed validation on the 'email' rule",
				"'uuidField' failed validation on the 'uuid' rule",
				"'minLenField' failed validation on the 'min' rule",
				"'maxLenField' failed validation on the 'max' rule",
				"'oneOfField' failed validation on the 'oneof' rule",
				"'subField' failed validation on the 'alpha' rule",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				if err != nil {
					for _, sub := range tt.errSubstr {
						assert.Contains(t, err.Error(), sub)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

