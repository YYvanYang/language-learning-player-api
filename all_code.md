# language-learning-player-backend Codebase

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
    # RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    # RUN chown -R appuser:appgroup /app
    # USER appuser
    
    # Expose the port the application listens on (from config, default 8080)
    EXPOSE 8080
    
    # Define the entry point for the container
    # This command will run when the container starts
    ENTRYPOINT ["/app/language-player-api"]
    
    # Optional: Define a default command (can be overridden)
    # CMD [""]
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

## `Makefile`

```
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
SQLC := $(GOBIN)/sqlc
# SWAG := $(shell go env GOPATH)/bin/swag # Temporarily comment out the dynamic path
# SWAG := /home/yvan/go/bin/swag # Temporarily hardcode the path - CHANGE IF YOURS IS DIFFERENT!
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
		if go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; then \
			echo ">>> migrate installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install migrate. Please check network connectivity and Go proxy settings."; \
			exit 1; \
		fi; \
	else \
		echo ">>> migrate is already installed."; \
	fi

# Check if sqlc is installed, if not, install it (Optional, if using sqlc)
install-sqlc:
	@if ! command -v sqlc &> /dev/null; then \
		echo ">>> Installing sqlc CLI..."; \
		if go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; then \
			echo ">>> sqlc installed successfully."; \
		else \
			echo ">>> ERROR: Failed to install sqlc. Please check network connectivity and Go proxy settings."; \
			exit 1; \
		fi; \
	else \
		echo ">>> sqlc is already installed."; \
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
.PHONY: generate generate-sqlc generate-swag swagger

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
	until curl -s --max-time 1 "http://localhost:$(MINIO_API_PORT)/minio/health/live" | grep -q 'OK'; do \
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
	@echo "  tools             Install necessary Go CLI tools"
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-create name=<name> Create a new migration file"
	@echo "  migrate-up        Apply database migrations (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-down      Revert the last database migration (requires DB running and DATABASE_URL set/exported)"
	@echo "  migrate-force version=<ver> Force migration version (requires DB running and DATABASE_URL set/exported)"
	@echo ""
	@echo "Code Generation & Formatting:"
	@echo "  generate          Run all code generators (sqlc, swag)"
	@echo "  generate-sqlc     Generate Go code from SQL using sqlc"
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
```

## `go.sum`

```
cloud.google.com/go/auth v0.15.0 h1:Ly0u4aA5vG/fsSsxu98qCQBemXtAtJf+95z9HK+cxps=
cloud.google.com/go/auth v0.15.0/go.mod h1:WJDGqZ1o9E9wKIL+IwStfyn/+s59zl4Bi+1KQNVXLZ8=
cloud.google.com/go/auth/oauth2adapt v0.2.8 h1:keo8NaayQZ6wimpNSmW5OPc283g65QNIiLpZnkHRbnc=
cloud.google.com/go/auth/oauth2adapt v0.2.8/go.mod h1:XQ9y31RkqZCcwJWNSx2Xvric3RrU88hAYYbjDWYDL+c=
cloud.google.com/go/compute/metadata v0.6.0 h1:A6hENjEsCDtC1k8byVsgwvVcioamEHvZ4j01OwKxG9I=
cloud.google.com/go/compute/metadata v0.6.0/go.mod h1:FjyFAW1MW0C203CEOMDTu3Dk1FlqW3Rga40jzHL4hfg=
dario.cat/mergo v1.0.0 h1:AGCNq9Evsj31mOgNPcLyXc+4PNABt905YmuqPYYpBWk=
dario.cat/mergo v1.0.0/go.mod h1:uNxQE+84aUszobStD9th8a29P2fMDhsBdgRYvZOxGmk=
filippo.io/edwards25519 v1.1.0 h1:FNf4tywRC1HmFuKW5xopWpigGjJKiJSV0Cqo0cJWDaA=
filippo.io/edwards25519 v1.1.0/go.mod h1:BxyFTGdWcka3PhytdK4V28tE5sGfRvvvRV7EaN4VDT4=
github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 h1:L/gRVlceqvL25UVaW/CKtUDjefjrs0SPonmDGUVOYP0=
github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161/go.mod h1:xomTg63KZ2rFqZQzSB4Vz2SUXa1BpHTVz9L5PTmPC4E=
github.com/KyleBanks/depth v1.2.1 h1:5h8fQADFrWtarTdtDudMmGsC7GPbOAu6RVB3ffsVFHc=
github.com/KyleBanks/depth v1.2.1/go.mod h1:jzSb9d0L43HxTQfT+oSA1EEp2q+ne2uh6XgeJcm8brE=
github.com/Microsoft/go-winio v0.6.2 h1:F2VQgta7ecxGYO8k3ZZz3RS8fVIXVxONVUPlNERoyfY=
github.com/Microsoft/go-winio v0.6.2/go.mod h1:yd8OoFMLzJbo9gZq8j5qaps8bJ9aShtEA8Ipt1oGCvU=
github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 h1:TngWCqHvy9oXAN6lEVMRuU21PR1EtLVZJmdB18Gu3Rw=
github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5/go.mod h1:lmUJ/7eu/Q8D7ML55dXQrVaamCz2vxCfdQBasLZfHKk=
github.com/cenkalti/backoff/v4 v4.3.0 h1:MyRJ/UdXutAwSAT+s3wNd7MfTIcy71VQueUuFK343L8=
github.com/cenkalti/backoff/v4 v4.3.0/go.mod h1:Y3VNntkOUPxTVeUxJ/G5vcM//AlwfmyYozVcomhLiZE=
github.com/containerd/continuity v0.4.5 h1:ZRoN1sXq9u7V6QoHMcVWGhOwDFqZ4B9i5H6un1Wh0x4=
github.com/containerd/continuity v0.4.5/go.mod h1:/lNJvtJKUQStBzpVQ1+rasXO1LAWtUQssk28EZvJ3nE=
github.com/creack/pty v1.1.18 h1:n56/Zwd5o6whRC5PMGretI4IdRLlmBXYNjScPaBgsbY=
github.com/creack/pty v1.1.18/go.mod h1:MOBLtS5ELjhRRrroQr9kyvTxUAFNvYEK993ew/Vr4O4=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/dhui/dktest v0.4.4 h1:+I4s6JRE1yGuqflzwqG+aIaMdgXIorCf5P98JnaAWa8=
github.com/dhui/dktest v0.4.4/go.mod h1:4+22R4lgsdAXrDyaH4Nqx2JEz2hLp49MqQmm9HLCQhM=
github.com/distribution/reference v0.6.0 h1:0IXCQ5g4/QMHHkarYzh5l+u8T3t73zM5QvfrDyIgxBk=
github.com/distribution/reference v0.6.0/go.mod h1:BbU0aIcezP1/5jX/8MP0YiH4SdvB5Y4f/wlDRiLyi3E=
github.com/docker/cli v27.4.1+incompatible h1:VzPiUlRJ/xh+otB75gva3r05isHMo5wXDfPRi5/b4hI=
github.com/docker/cli v27.4.1+incompatible/go.mod h1:JLrzqnKDaYBop7H2jaqPtU4hHvMKP+vjCwu2uszcLI8=
github.com/docker/docker v27.2.0+incompatible h1:Rk9nIVdfH3+Vz4cyI/uhbINhEZ/oLmc+CBXmH6fbNk4=
github.com/docker/docker v27.2.0+incompatible/go.mod h1:eEKB0N0r5NX/I1kEveEz05bcu8tLC/8azJZsviup8Sk=
github.com/docker/go-connections v0.5.0 h1:USnMq7hx7gwdVZq1L49hLXaFtUdTADjXGp+uj1Br63c=
github.com/docker/go-connections v0.5.0/go.mod h1:ov60Kzw0kKElRwhNs9UlUHAE/F9Fe6GLaXnqyDdmEXc=
github.com/docker/go-units v0.5.0 h1:69rxXcBk27SvSaaxTtLh/8llcHD8vYHT7WSdRZ/jvr4=
github.com/docker/go-units v0.5.0/go.mod h1:fgPhTUdO+D/Jk86RDLlptpiXQzgHJF7gydDDbaIK4Dk=
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
github.com/go-sql-driver/mysql v1.8.1 h1:LedoTUt/eveggdHS9qUFC1EFSa8bU2+1pZjSRpvNJ1Y=
github.com/go-sql-driver/mysql v1.8.1/go.mod h1:wEBSXgmK//2ZFJyE+qWnIsVGmvmEKlqwuVSjsCm7DZg=
github.com/go-viper/mapstructure/v2 v2.2.1 h1:ZAaOCxANMuZx5RCeg0mBdEZk7DZasvvZIxtHqx8aGss=
github.com/go-viper/mapstructure/v2 v2.2.1/go.mod h1:oJDH3BJKyqBA2TXFhDsKDGDTlndYOZ6rGS0BRZIxGhM=
github.com/goccy/go-json v0.10.5 h1:Fq85nIqj+gXn/S5ahsiTlK3TmC85qgirsdTP/+DeaC4=
github.com/goccy/go-json v0.10.5/go.mod h1:oq7eo15ShAhp70Anwd5lgX2pLfOS3QCiwU/PULtXL6M=
github.com/gogo/protobuf v1.3.2 h1:Ov1cvc58UF3b5XjBnZv7+opcTcQFZebYjWzi34vdm4Q=
github.com/gogo/protobuf v1.3.2/go.mod h1:P1XiOD3dCwIKUDQYPy72D8LYyHL2YPYrpS2s69NZV8Q=
github.com/golang-jwt/jwt/v5 v5.2.2 h1:Rl4B7itRWVtYIHFrSNd7vhTiz9UpLdi6gZhZ3wEeDy8=
github.com/golang-jwt/jwt/v5 v5.2.2/go.mod h1:pqrtFR0X4osieyHYxtmOUWsAWrfe1Q5UVIyoH402zdk=
github.com/golang-migrate/migrate/v4 v4.18.2 h1:2VSCMz7x7mjyTXx3m2zPokOY82LTRgxK1yQYKo6wWQ8=
github.com/golang-migrate/migrate/v4 v4.18.2/go.mod h1:2CM6tJvn2kqPXwnXO/d3rAQYiyoIm180VsO8PRX6Rpk=
github.com/golang/protobuf v1.5.4 h1:i7eJL8qZTpSEXOPTxNKhASYpMn+8e5Q6AdndVa1dWek=
github.com/golang/protobuf v1.5.4/go.mod h1:lnTiLA8Wa4RWRcIUkrtSVa5nRhsEGBg48fD6rSs7xps=
github.com/google/go-cmp v0.7.0 h1:wk8382ETsv4JYUZwIsn6YpYiWiBsYLSJiTsyBybVuN8=
github.com/google/go-cmp v0.7.0/go.mod h1:pXiqmnSA92OHEEa9HXL2W4E7lf9JzCmGVUdgjX3N/iU=
github.com/google/s2a-go v0.1.9 h1:LGD7gtMgezd8a/Xak7mEWL0PjoTQFvpRudN895yqKW0=
github.com/google/s2a-go v0.1.9/go.mod h1:YA0Ei2ZQL3acow2O62kdp9UlnvMmU7kA6Eutn0dXayM=
github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 h1:El6M4kTTCOh6aBiKaUGG7oYTSPP8MxqL4YI3kZKwcP4=
github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510/go.mod h1:pupxD2MaaD3pAXIBCelhxNneeOaAeabZDe5s4K6zSpQ=
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/googleapis/enterprise-certificate-proxy v0.3.6 h1:GW/XbdyBFQ8Qe+YAmFU9uHLo7OnF5tL52HFAgMmyrf4=
github.com/googleapis/enterprise-certificate-proxy v0.3.6/go.mod h1:MkHOF77EYAE7qfSuSS9PU6g4Nt4e11cnsDUowfwewLA=
github.com/googleapis/gax-go/v2 v2.14.1 h1:hb0FFeiPaQskmvakKu5EbCbpntQn48jyHuvrkurSS/Q=
github.com/googleapis/gax-go/v2 v2.14.1/go.mod h1:Hb/NubMaVM88SrNkvl8X/o8XWwDJEPqouaLeN2IUxoA=
github.com/hashicorp/errwrap v1.0.0/go.mod h1:YH+1FKiLXxHSkmPseP+kNlulaMuP3n2brvKWEqk/Jc4=
github.com/hashicorp/errwrap v1.1.0 h1:OxrOeh75EUXMY8TBjag2fzXGZ40LB6IKw45YeGUDY2I=
github.com/hashicorp/errwrap v1.1.0/go.mod h1:YH+1FKiLXxHSkmPseP+kNlulaMuP3n2brvKWEqk/Jc4=
github.com/hashicorp/go-multierror v1.1.1 h1:H5DkEtf6CXdFp0N0Em5UCwQpXMWke8IA0+lD48awMYo=
github.com/hashicorp/go-multierror v1.1.1/go.mod h1:iw975J/qwKPdAO1clOe2L8331t/9/fmwbPZ6JB6eMoM=
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
github.com/kisielk/errcheck v1.5.0/go.mod h1:pFxgyoBC7bSaBwPgfKdkLd5X25qrDl4LWUI2bnpBCr8=
github.com/kisielk/gotool v1.0.0/go.mod h1:XhKaO+MFFWcvkIS/tQcRk01m1F5IRFswLeQ+oQHNcck=
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
github.com/moby/docker-image-spec v1.3.1 h1:jMKff3w6PgbfSa69GfNg+zN/XLhfXJGnEx3Nl2EsFP0=
github.com/moby/docker-image-spec v1.3.1/go.mod h1:eKmb5VW8vQEh/BAr2yvVNvuiJuY6UIocYsFu/DxxRpo=
github.com/moby/sys/user v0.3.0 h1:9ni5DlcW5an3SvRSx4MouotOygvzaXbaSrc/wGDFWPo=
github.com/moby/sys/user v0.3.0/go.mod h1:bG+tYYYJgaMtRKgEmuueC0hJEAZWwtIbZTB+85uoHjs=
github.com/moby/term v0.5.0 h1:xt8Q1nalod/v7BqbG21f8mQPqH+xAaC9C3N3wfWbVP0=
github.com/moby/term v0.5.0/go.mod h1:8FzsFHVUBGZdbDsJw/ot+X+d5HLUbvklYLJ9uGfcI3Y=
github.com/morikuni/aec v1.0.0 h1:nP9CBfwrvYnBRgY6qfDQkygYDmYwOilePFkwzv4dU8A=
github.com/morikuni/aec v1.0.0/go.mod h1:BbKIizmSmc5MMPqRYbxO4ZU0S0+P200+tUnFx7PXmsc=
github.com/opencontainers/go-digest v1.0.0 h1:apOUWs51W5PlhuyGyz9FCeeBIOUDA/6nW8Oi/yOhh5U=
github.com/opencontainers/go-digest v1.0.0/go.mod h1:0JzlMkj0TRzQZfJkVvzbP0HBR3IKzErnv2BNG4W4MAM=
github.com/opencontainers/image-spec v1.1.0 h1:8SG7/vwALn54lVB/0yZ/MMwhFrPYtpEHQb2IpWsCzug=
github.com/opencontainers/image-spec v1.1.0/go.mod h1:W4s4sFTMaBeK1BQLXbG4AdM2szdn85PY75RI83NrTrM=
github.com/opencontainers/runc v1.2.3 h1:fxE7amCzfZflJO2lHXf4y/y8M1BoAqp+FVmG19oYB80=
github.com/opencontainers/runc v1.2.3/go.mod h1:nSxcWUydXrsBZVYNSkTjoQ/N6rcyTtn+1SD5D4+kRIM=
github.com/ory/dockertest/v3 v3.12.0 h1:3oV9d0sDzlSQfHtIaB5k6ghUCVMVLpAY8hwrqoCyRCw=
github.com/ory/dockertest/v3 v3.12.0/go.mod h1:aKNDTva3cp8dwOWwb9cWuX84aH5akkxXRvO7KCwWVjE=
github.com/pelletier/go-toml/v2 v2.2.3 h1:YmeHyLY8mFWbdkNWwpr+qIL2bEqT0o95WSdkNHvL12M=
github.com/pelletier/go-toml/v2 v2.2.3/go.mod h1:MfCQTFTvCcUyyvvwm1+G6H/jORL20Xlb6rzQu9GuUkc=
github.com/pkg/errors v0.9.1 h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=
github.com/pkg/errors v0.9.1/go.mod h1:bwawxfHBFNV+L2hUp1rHADufV3IMtnDRdf1r5NINEl0=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/rogpeppe/go-internal v1.13.1 h1:KvO1DLK/DRN07sQ1LQKScxyZJuNnedQ5/wKSR38lUII=
github.com/rogpeppe/go-internal v1.13.1/go.mod h1:uMEvuHeurkdAXX61udpOXGD/AzZDWNMNyH2VO9fmH0o=
github.com/rs/xid v1.6.0 h1:fV591PaemRlL6JfRxGDEPl69wICngIQ3shQtzfy2gxU=
github.com/rs/xid v1.6.0/go.mod h1:7XoLgs4eV+QndskICGsho+ADou8ySMSjJKDIan90Nz0=
github.com/sagikazarmark/locafero v0.7.0 h1:5MqpDsTGNDhY8sGp0Aowyf0qKsPrhewaLSsFaodPcyo=
github.com/sagikazarmark/locafero v0.7.0/go.mod h1:2za3Cg5rMaTMoG/2Ulr9AwtFaIppKXTRYnozin4aB5k=
github.com/sirupsen/logrus v1.9.3 h1:dueUQJ1C2q9oE3F7wvmSGAaVtTmUizReu6fjN8uqzbQ=
github.com/sirupsen/logrus v1.9.3/go.mod h1:naHLuLoDiP4jHNo9R0sCBMtWGeIprob74mVsIT4qYEQ=
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
github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f/go.mod h1:N2zxlSyiKSe5eX1tZViRH5QA0qijqEDrYZiPEAiq3wU=
github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb h1:zGWFAtiMcyryUHoUjUJX0/lt1H2+i2Ka2n+D3DImSNo=
github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb/go.mod h1:N2zxlSyiKSe5eX1tZViRH5QA0qijqEDrYZiPEAiq3wU=
github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 h1:EzJWgHovont7NscjpAxXsDA8S8BMYve8Y5+7cuRE7R0=
github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415/go.mod h1:GwrjFmJcFw6At/Gs6z4yjiIwzuJ1/+UwLxMQDVQXShQ=
github.com/xeipuuv/gojsonschema v1.2.0 h1:LhYJRs+L4fBtjZUfuSZIKGeVu0QRy8e5Xi7D17UxZ74=
github.com/xeipuuv/gojsonschema v1.2.0/go.mod h1:anYRn/JVcOK2ZgGU+IjEV4nwlhoK5sQluxsYJ78Id3Y=
github.com/yuin/goldmark v1.1.27/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
github.com/yuin/goldmark v1.2.1/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
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
golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550/go.mod h1:yigFU9vqHzYiE8UmvKecakEJjdnWj3jj499lnFckfCI=
golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9/go.mod h1:LzIPMQfyMNhhGPhUkYOs5KpL4U8rLKemX1yGLhDgUto=
golang.org/x/crypto v0.0.0-20210921155107-089bfa567519/go.mod h1:GvvjBRRGRdwPK5ydBHafDWAxML/pGHZbMvKqRZ5+Abc=
golang.org/x/crypto v0.36.0 h1:AnAEvhDddvBdpY+uR+MyHmuZzzNqXSe/GvuDeob5L34=
golang.org/x/crypto v0.36.0/go.mod h1:Y4J0ReaxCR1IMaabaSMugxJES1EpwhBHhv2bDHklZvc=
golang.org/x/mod v0.2.0/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
golang.org/x/mod v0.3.0/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4/go.mod h1:jJ57K6gSWd91VN4djpZkiMVwK6gcyfeH4XE8wZrZaV4=
golang.org/x/mod v0.24.0 h1:ZfthKaKaT4NrhGVZHO1/WDTwGES4De8KtWO0SIbNJMU=
golang.org/x/mod v0.24.0/go.mod h1:IXM97Txy2VM4PJ3gI61r1YEk/gAj6zAHN3AdZt6S9Ww=
golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
golang.org/x/net v0.0.0-20190620200207-3b0461eec859/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
golang.org/x/net v0.0.0-20200226121028-0de0cce0169b/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
golang.org/x/net v0.0.0-20201021035429-f5854403a974/go.mod h1:sp8m0HH+o8qH0wwXwYZr8TS3Oi6o0r6Gce1SSxlDquU=
golang.org/x/net v0.0.0-20210226172049-e18ecbb05110/go.mod h1:m0MpNAwzfU5UDzcl9v0D8zg8gWTRqZa9RBIspLL5mdg=
golang.org/x/net v0.0.0-20220722155237-a158d28d115b/go.mod h1:XRhObCWvk6IyKnWLug+ECip1KBveYUHfp+8e9klMJ9c=
golang.org/x/net v0.7.0/go.mod h1:2Tu9+aMcznHK/AK1HMvgo6xiTLG5rD5rZLDS+rp2Bjs=
golang.org/x/net v0.37.0 h1:1zLorHbz+LYj7MQlSf1+2tPIIgibq2eL5xkrGk6f+2c=
golang.org/x/net v0.37.0/go.mod h1:ivrbrMbzFq5J41QOQh0siUuly180yBYtLp+CKbEaFx8=
golang.org/x/oauth2 v0.28.0 h1:CrgCKl8PPAVtLnU3c+EDw6x11699EWlsDeWNWKdIOkc=
golang.org/x/oauth2 v0.28.0/go.mod h1:onh5ek6nERTohokkhCD/y2cV4Do3fxFHFuAejCkRWT8=
golang.org/x/sync v0.0.0-20190423024810-112230192c58/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.12.0 h1:MHc5BpPuC30uJk597Ri8TV3CNZcTLu6B6z4lJy+g6Jw=
golang.org/x/sync v0.12.0/go.mod h1:1dzgHSNfp02xaA81J2MS99Qcpr2w7fw1gpm99rleRqA=
golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
golang.org/x/sys v0.0.0-20190412213103-97732733099d/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20201119102817-f84b799fce68/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20210616094352-59db8d763f22/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
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
golang.org/x/tools v0.0.0-20200619180055-7c47624df98f/go.mod h1:EkVYQZoAsY45+roYkvgYkIh4xh/qjgUK9TdY2XT94GE=
golang.org/x/tools v0.0.0-20210106214847-113979e3529a/go.mod h1:emZCQorbCU4vsT4fOWvOPXz4eW1wZW4PmDk9uLelYpA=
golang.org/x/tools v0.1.12/go.mod h1:hNGJHUnrk76NpqgfD5Aqm5Crs+Hm0VOH/i9J2+nxYbc=
golang.org/x/tools v0.31.0 h1:0EedkvKDbh+qistFTd0Bcwe/YLh4vHwWEkiI0toFIBU=
golang.org/x/tools v0.31.0/go.mod h1:naFTU+Cev749tSJRXJlna0T3WxKvb1kWEx15xA4SdmQ=
golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
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
gopkg.in/yaml.v2 v2.4.0 h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=
gopkg.in/yaml.v2 v2.4.0/go.mod h1:RDklbk79AGWmwhnvt/jBztapEOGDOx6ZbXqjP6csGnQ=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gotest.tools/v3 v3.5.1 h1:EENdUnS3pdur5nybKYIh2Vfgc8IUNBjxDPSjtiJcOzU=
gotest.tools/v3 v3.5.1/go.mod h1:isy3WKz7GK6uNw/sbHzfKBLvlvXwUyV06n6brMxxopU=
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

## `go.mod`

```
module github.com/yvanyang/language-learning-player-backend

go 1.23.3

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/go-playground/validator/v10 v10.26.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v7 v7.0.89
	github.com/ory/dockertest/v3 v3.12.0
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
	dario.cat/mergo v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v27.4.1+incompatible // indirect
	github.com/docker/docker v27.2.0+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
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
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
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
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.2.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
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
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
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

## `cmd/api/main.go`

```go
// cmd/api/main.go
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

	_ "github.com/yvanyang/language-learning-player-backend/docs" // Keep this - Import generated docs
	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
	httpadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias for http handler package
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // Adjust import path for our custom middleware
	repo "github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres" // Alias for postgres repo package
	googleauthadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/google_auth"
	minioadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/minio"
	uc "github.com/yvanyang/language-learning-player-backend/internal/usecase" // Alias usecase package if needed elsewhere
	"github.com/yvanyang/language-learning-player-backend/pkg/logger"      // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/security"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
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
	cfg, err := config.LoadConfig(".")
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

	// --- ADDED: Initialize Transaction Manager ---
	txManager := repo.NewTxManager(dbPool, appLogger)

	// Repositories
	userRepo := repo.NewUserRepository(dbPool, appLogger)
	trackRepo := repo.NewAudioTrackRepository(dbPool, appLogger)
	collectionRepo := repo.NewAudioCollectionRepository(dbPool, appLogger)
	progressRepo := repo.NewPlaybackProgressRepository(dbPool, appLogger)
	bookmarkRepo := repo.NewBookmarkRepository(dbPool, appLogger)

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
		appLogger.Error("Failed to initialize Google Auth service", "error", err)
		os.Exit(1) // Assuming Google Auth is critical
	}

	// Validator
	validator := validation.New()

	// --- Inject TransactionManager into Use Cases that need it ---
	authUseCase := uc.NewAuthUseCase(cfg.JWT, userRepo, secHelper, googleAuthService, appLogger)
	audioUseCase := uc.NewAudioContentUseCase(cfg.Minio, trackRepo, collectionRepo, storageService, txManager, appLogger) // Pass txManager
	activityUseCase := uc.NewUserActivityUseCase(progressRepo, bookmarkRepo, trackRepo, appLogger)
	uploadUseCase := uc.NewUploadUseCase(cfg.Minio, trackRepo, storageService, appLogger)
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

	// --- Middleware Setup ---
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RequestLogger)
	router.Use(chimiddleware.RealIP)
	ipLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20) // Example: 10 req/sec, burst 20
	router.Use(middleware.RateLimit(ipLimiter))
	router.Use(chimiddleware.StripSlashes)
	router.Use(chimiddleware.Timeout(60 * time.Second)) // Example timeout
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		ExposedHeaders:   []string{"Link", "X-Request-ID"}, // Expose Request ID if needed by client
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))

	// --- Routes ---
	// Health Check
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add deeper health checks (e.g., DB ping) if needed
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Swagger Docs
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	// API v1 Routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/google/callback", authHandler.GoogleCallback)

			// Public Audio Routes
			r.Get("/audio/tracks", audioHandler.ListTracks)
			r.Get("/audio/tracks/{trackId}", audioHandler.GetTrackDetails) // Maybe requires auth if track is private? Handled in usecase/handler for now.
			// Public Collection Routes (If any? e.g., listing public courses)
			// r.Get("/audio/collections", audioHandler.ListPublicCollections) // Example
			r.Get("/audio/collections/{collectionId}", audioHandler.GetCollectionDetails) // Requires auth if collection is private
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticator(secHelper))

			// User Profile
			r.Get("/users/me", userHandler.GetMyProfile)
			// r.Put("/users/me", userHandler.UpdateMyProfile) // TODO

			// Authenticated Collection Actions
			r.Post("/audio/collections", audioHandler.CreateCollection)
			// r.Get("/users/me/collections", audioHandler.ListMyCollections) // Endpoint to list *only* user's collections
			r.Put("/audio/collections/{collectionId}", audioHandler.UpdateCollectionMetadata)
			r.Delete("/audio/collections/{collectionId}", audioHandler.DeleteCollection)
			r.Put("/audio/collections/{collectionId}/tracks", audioHandler.UpdateCollectionTracks)
			// Add/Remove single track endpoints?
			// r.Post("/audio/collections/{collectionId}/tracks", audioHandler.AddTrackToCollection)
			// r.Delete("/audio/collections/{collectionId}/tracks/{trackId}", audioHandler.RemoveTrackFromCollection)

			// User Activity Routes
			r.Post("/users/me/progress", activityHandler.RecordProgress)
			r.Get("/users/me/progress", activityHandler.ListProgress)
			r.Get("/users/me/progress/{trackId}", activityHandler.GetProgress)

			r.Post("/bookmarks", activityHandler.CreateBookmark)
			r.Get("/bookmarks", activityHandler.ListBookmarks)
			r.Delete("/bookmarks/{bookmarkId}", activityHandler.DeleteBookmark)

			// Upload Routes (Require Auth)
			r.Post("/uploads/audio/request", uploadHandler.RequestUpload)
			r.Post("/audio/tracks", uploadHandler.CompleteUploadAndCreateTrack) // Create track metadata after upload
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
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError), // Route server errors to slog
	}

	// --- Start Server & Graceful Shutdown ---
	go func() {
		appLogger.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("Server failed to start", "error", err)
			os.Exit(1) // Exit if server cannot start
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 10 seconds to finish
	// the requests it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close database pool first
	appLogger.Info("Closing database connection pool...")
	dbPool.Close()
	appLogger.Info("Database connection pool closed.")

	// Attempt graceful server shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server shutdown complete.")
}
```

## `docs/openapi.yaml`

```yaml
# docs/openapi.yaml
openapi: 3.0.3
info:
  title: Language Learning Audio Player API
  description: API specification for the backend of the Language Learning Audio Player application. Provides endpoints for user authentication, audio content management, and user activity tracking.
  version: 1.0.0 # API Version (distinct from app version)
  contact:
    name: API Support Team
    email: support@example.com # Replace with actual contact
  license:
    name: Apache 2.0 # Or your chosen license
    url: http://www.apache.org/licenses/LICENSE-2.0.html

servers:
  - url: http://localhost:8080/api/v1 # Local development server base path
    description: Local Development Server
  - url: https://api.yourdomain.com/api/v1 # Production server base path (Replace!)
    description: Production Server

tags: # Group related endpoints
  - name: Authentication
    description: User registration, login, and third-party authentication.
  - name: User Activity
    description: Managing user playback progress and bookmarks.
  - name: Audio Tracks
    description: Accessing audio track metadata and playback URLs.
  - name: Audio Collections
    description: Managing user-created audio collections (playlists, courses).
  - name: Health
    description: API health checks.

paths:
  /healthz:
    get:
      tags:
        - Health
      summary: Check API Health
      description: Returns OK status if the API server is running. May include checks for database connectivity in the future.
      operationId: getHealthCheck
      responses:
        '200':
          description: API is healthy.
          content:
            text/plain:
              schema:
                type: string
                example: OK
        '503':
          description: Service unavailable (e.g., cannot connect to database).
          $ref: '#/components/responses/ErrorResponse' # Reference shared error response

  /auth/register:
    post:
      tags:
        - Authentication
      summary: Register a new user
      description: Registers a new user account using email and password.
      operationId: registerUser
      requestBody:
        description: User registration details.
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegisterRequestDTO' # Reference DTO schema
      responses:
        '201':
          description: User registered successfully. Returns an authentication token.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AuthResponseDTO'
        '400':
          description: Invalid input data (e.g., invalid email format, weak password).
          $ref: '#/components/responses/ValidationErrorResponse'
        '409':
          description: Conflict - Email already exists.
          $ref: '#/components/responses/ConflictErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'

  /auth/login:
    post:
      tags:
        - Authentication
      summary: Login a user
      description: Authenticates a user with email and password, returns a JWT token.
      operationId: loginUser
      requestBody:
        description: User login credentials.
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequestDTO'
      responses:
        '200':
          description: Login successful. Returns an authentication token.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AuthResponseDTO'
        '400':
          description: Invalid input data (e.g., missing fields).
          $ref: '#/components/responses/ValidationErrorResponse'
        '401':
          description: Authentication failed (invalid email or password).
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'

  /auth/google/callback:
    post:
      tags:
        - Authentication
      summary: Handle Google Sign-In callback
      description: Verifies a Google ID Token obtained by the client, logs in or registers the user, and returns an application JWT token.
      operationId: handleGoogleCallback
      requestBody:
        description: Google ID Token.
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GoogleCallbackRequestDTO'
      responses:
        '200':
          description: Google authentication successful. Returns an authentication token and optionally indicates if it was a new user registration.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AuthResponseDTO' # Includes optional isNewUser field
        '400':
          description: Invalid input (e.g., missing ID token).
          $ref: '#/components/responses/ValidationErrorResponse'
        '401':
          description: Authentication failed (invalid Google token).
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '409':
          description: Conflict - Email associated with Google account already exists with a different login method or different linked Google account.
          content:
            application/json:
              schema:
                 $ref: '#/components/schemas/ErrorResponseDTO' # Use standard error DTO
                 example:
                   code: "EMAIL_EXISTS"
                   message: "Email is already registered with a different method."
                   requestId: "req_123abc"
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'

  /users/me:
    get:
      tags:
        - User Activity # Or a dedicated "User Profile" tag
      summary: Get current user profile
      description: Retrieves the profile information for the currently authenticated user.
      operationId: getCurrentUserProfile
      security:
        - BearerAuth: [] # Indicates this endpoint requires Bearer token authentication
      responses:
        '200':
          description: Successfully retrieved user profile.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponseDTO' # Define this DTO
        '401':
          description: Unauthorized - No or invalid token provided.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'
    # TODO: Add PUT /users/me endpoint description

  /users/me/progress:
    post:
      tags:
        - User Activity
      summary: Record playback progress
      description: Records or updates the listening progress for a specific track for the current user.
      operationId: recordUserProgress
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RecordProgressRequestDTO'
      responses:
        '204':
          description: Progress successfully recorded/updated. No content returned.
        '400':
          description: Invalid input data (e.g., invalid track ID, negative progress).
          $ref: '#/components/responses/ValidationErrorResponse'
        '401':
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '404':
          description: Track not found.
          $ref: '#/components/responses/NotFoundErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'
    get:
      tags:
        - User Activity
      summary: List user playback progress
      description: Retrieves a paginated list of all playback progress records for the current user, ordered by last listened time.
      operationId: listUserProgress
      security:
        - BearerAuth: []
      parameters:
        - $ref: '#/components/parameters/LimitParam'
        - $ref: '#/components/parameters/OffsetParam'
      responses:
        '200':
          description: A paginated list of playback progress records.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PaginatedProgressResponseDTO'
        '401':
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'

  /users/me/progress/{trackId}:
    get:
      tags:
        - User Activity
      summary: Get progress for a specific track
      description: Retrieves the playback progress for a specific track for the current user.
      operationId: getUserTrackProgress
      security:
        - BearerAuth: []
      parameters:
        - name: trackId
          in: path
          required: true
          description: The UUID of the audio track.
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Playback progress found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PlaybackProgressResponseDTO'
        '401':
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '404':
          description: Progress record not found for this user and track (or track itself not found).
          $ref: '#/components/responses/NotFoundErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'


  /bookmarks:
    post:
      tags:
        - User Activity
      summary: Create a bookmark
      description: Creates a new bookmark at a specific timestamp in an audio track for the current user.
      operationId: createBookmark
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateBookmarkRequestDTO'
      responses:
        '201':
          description: Bookmark created successfully. Returns the created bookmark details.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BookmarkResponseDTO'
        '400':
          description: Invalid input data.
          $ref: '#/components/responses/ValidationErrorResponse'
        '401':
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '404':
          description: Track not found.
          $ref: '#/components/responses/NotFoundErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'
    get:
      tags:
        - User Activity
      summary: List user bookmarks
      description: Retrieves a paginated list of bookmarks for the current user, optionally filtered by track ID. Ordered by creation date descending.
      operationId: listUserBookmarks
      security:
        - BearerAuth: []
      parameters:
        - $ref: '#/components/parameters/LimitParam'
        - $ref: '#/components/parameters/OffsetParam'
        - name: trackId
          in: query
          required: false
          description: Filter bookmarks by a specific track UUID.
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: A paginated list of bookmarks.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PaginatedBookmarksResponseDTO'
        '401':
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'

  /bookmarks/{bookmarkId}:
    delete:
      tags:
        - User Activity
      summary: Delete a bookmark
      description: Deletes a specific bookmark owned by the current user.
      operationId: deleteBookmark
      security:
        - BearerAuth: []
      parameters:
        - name: bookmarkId
          in: path
          required: true
          description: The UUID of the bookmark to delete.
          schema:
            type: string
            format: uuid
      responses:
        '204':
          description: Bookmark deleted successfully. No content returned.
        '401':
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '403':
          description: Forbidden - User does not own this bookmark.
          $ref: '#/components/responses/ForbiddenErrorResponse'
        '404':
          description: Bookmark not found.
          $ref: '#/components/responses/NotFoundErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'


  /audio/tracks:
    get:
      tags:
        - Audio Tracks
      summary: List audio tracks
      description: Retrieves a list of audio tracks, supports filtering, sorting, and pagination. Primarily lists public tracks unless specific permissions grant access to private ones (handled by implementation).
      operationId: listAudioTracks
      # No security needed for basic public listing, but add BearerAuth if filtering/sorting requires auth
      parameters:
        - $ref: '#/components/parameters/LimitParam'
        - $ref: '#/components/parameters/OffsetParam'
        - name: q
          in: query
          description: Search query string (searches title and description).
          required: false
          schema:
            type: string
        - name: lang
          in: query
          description: Filter by language code (e.g., 'en-US').
          required: false
          schema:
            type: string
        - name: level
          in: query
          description: Filter by difficulty level (e.g., 'A1', 'B2', 'NATIVE').
          required: false
          schema:
            type: string
            enum: [A1, A2, B1, B2, C1, C2, NATIVE] # Reference domain.AudioLevel
        - name: isPublic
          in: query
          description: Filter by public visibility (true or false).
          required: false
          schema:
            type: boolean
        - name: tags
          in: query
          description: Filter by tags (provide multiple times, e.g., ?tags=news&tags=easy). Matches tracks containing ALL specified tags (or adjust logic/repo query for ANY).
          required: false
          style: form # How array parameters are formatted
          explode: true # Use separate parameters for each value (?tags=a&tags=b)
          schema:
            type: array
            items:
              type: string
        - name: sortBy
          in: query
          description: Field to sort by (e.g., 'createdAt', 'title', 'duration', 'level').
          required: false
          schema:
            type: string
            enum: [createdAt, title, duration, level] # Whitelist sortable fields
        - name: sortDir
          in: query
          description: Sort direction ('asc' or 'desc'). Default is 'desc' for 'createdAt', 'asc' otherwise? (Define default behavior).
          required: false
          schema:
            type: string
            enum: [asc, desc]
      responses:
        '200':
          description: A paginated list of audio tracks.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PaginatedTracksResponseDTO'
        '400':
          description: Invalid query parameter value (e.g., invalid level, invalid boolean for isPublic).
          $ref: '#/components/responses/ValidationErrorResponse'
        '500':
          description: Internal server error.
          $ref: '#/components/responses/ErrorResponse'

  /audio/tracks/{trackId}:
    get:
      tags:
        - Audio Tracks
      summary: Get audio track details
      description: Retrieves details for a specific audio track, including metadata and a temporary playback URL. Access to private tracks requires authentication.
      operationId: getAudioTrackDetails
      # Add security if private tracks exist and require auth to view *any* detail
      # security:
      #  - BearerAuth: [] # Optional based on whether private tracks return 401/403 or just exclude playUrl
      parameters:
        - name: trackId
          in: path
          required: true
          description: The UUID of the audio track.
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Audio track details found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AudioTrackDetailsResponseDTO'
        '404':
          description: Audio track not found.
          $ref: '#/components/responses/NotFoundErrorResponse'
        '401': # If security is applied and token is invalid for private track access
          description: Unauthorized.
          $ref: '#/components/responses/UnauthorizedErrorResponse'
        '403': # If security is applied and user doesn't have permission for private track
          description: Forbidden.
          $ref: '#/components/responses/ForbiddenErrorResponse'
        '500':
          description: Internal server error (e.g., failed to generate presigned URL).
          $ref: '#/components/responses/ErrorResponse'

  # --- TODO: Add paths for Audio Collections ---
  # POST /audio/collections
  # GET /audio/collections/{collectionId}
  # PUT /audio/collections/{collectionId}
  # DELETE /audio/collections/{collectionId}
  # PUT /audio/collections/{collectionId}/tracks

components:
  schemas:
    # --- Request DTOs ---
    RegisterRequestDTO:
      type: object
      required: [email, password, name]
      properties:
        email:
          type: string
          format: email
          description: User's email address. Must be unique.
          example: user@example.com
        password:
          type: string
          format: password
          minLength: 8
          description: User's password (at least 8 characters).
          example: Str0ngP@ssw0rd
        name:
          type: string
          maxLength: 100
          description: User's display name.
          example: John Doe
    LoginRequestDTO:
      type: object
      required: [email, password]
      properties:
        email:
          type: string
          format: email
          example: user@example.com
        password:
          type: string
          format: password
          example: Str0ngP@ssw0rd
    GoogleCallbackRequestDTO:
      type: object
      required: [idToken]
      properties:
        idToken:
          type: string
          description: The Google ID Token obtained from the client-side Google Sign-In flow.
          example: "eyJhbGciOiJSUzI1NiIsImtpZCI6I..."
    RecordProgressRequestDTO:
      type: object
      required: [trackId, progressSeconds]
      properties:
        trackId:
          type: string
          format: uuid
          description: The UUID of the audio track.
          example: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
        progressSeconds:
          type: number
          format: double
          minimum: 0
          description: The playback progress in seconds.
          example: 123.45
    CreateBookmarkRequestDTO:
      type: object
      required: [trackId, timestampSeconds]
      properties:
        trackId:
          type: string
          format: uuid
          description: The UUID of the audio track.
          example: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
        timestampSeconds:
          type: number
          format: double
          minimum: 0
          description: The timestamp in the audio track (in seconds) where the bookmark is placed.
          example: 65.0
        note:
          type: string
          description: An optional note associated with the bookmark.
          example: "Remember this vocabulary."
    CreateCollectionRequestDTO:
      # TODO: Define schema based on DTO
      type: object
      # ...
    UpdateCollectionRequestDTO:
      # TODO: Define schema based on DTO
      type: object
      # ...
    UpdateCollectionTracksRequestDTO:
      # TODO: Define schema based on DTO
      type: object
      # ...

    # --- Response DTOs ---
    AuthResponseDTO:
      type: object
      required: [token]
      properties:
        token:
          type: string
          description: The JWT authentication token for subsequent API requests.
          example: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        isNewUser:
          type: boolean
          description: Included and set to true only if the authentication resulted in a new user account being created (typically via third-party login like Google).
          example: true
    UserResponseDTO:
      type: object
      properties:
        id:
          type: string
          format: uuid
          example: "a1b2c3d4-e5f6-7890-1234-567890abcdef"
        email:
          type: string
          format: email
          example: "user@example.com"
        name:
          type: string
          example: "John Doe"
        authProvider:
          type: string
          enum: [local, google] # Reference domain.AuthProvider
          example: "local"
        profileImageUrl:
          type: string
          format: url
          nullable: true
          example: "https://lh3.googleusercontent.com/a-/..."
        createdAt:
          type: string
          format: date-time # RFC3339 format
          example: "2023-10-27T10:00:00Z"
        updatedAt:
          type: string
          format: date-time
          example: "2023-10-27T10:05:00Z"
    PlaybackProgressResponseDTO:
      type: object
      properties:
         userId:
           type: string
           format: uuid
         trackId:
           type: string
           format: uuid
         progressSeconds:
           type: number
           format: double
         lastListenedAt:
           type: string
           format: date-time
    BookmarkResponseDTO:
      type: object
      properties:
        id:
          type: string
          format: uuid
        userId:
           type: string
           format: uuid
        trackId:
           type: string
           format: uuid
        timestampSeconds:
           type: number
           format: double
        note:
          type: string
          nullable: true
        createdAt:
           type: string
           format: date-time
    AudioTrackResponseDTO:
      type: object
      properties:
        id: { type: string, format: uuid }
        title: { type: string }
        description: { type: string }
        languageCode: { type: string }
        level: { type: string, enum: [A1, A2, B1, B2, C1, C2, NATIVE] }
        durationMs: { type: integer, format: int64 }
        coverImageUrl: { type: string, format: url, nullable: true }
        uploaderId: { type: string, format: uuid, nullable: true }
        isPublic: { type: boolean }
        tags: { type: array, items: { type: string } }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
    AudioTrackDetailsResponseDTO:
      allOf: # Combine Track DTO with playUrl
        - $ref: '#/components/schemas/AudioTrackResponseDTO'
        - type: object
          properties:
            playUrl:
              type: string
              format: url
              description: A temporary, pre-signed URL to stream the audio content. Valid for a limited time.
    AudioCollectionResponseDTO:
      # TODO: Define schema based on DTO
      type: object
      # ...
    PaginatedProgressResponseDTO:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/PlaybackProgressResponseDTO'
        total: { type: integer, format: int32 }
        limit: { type: integer, format: int32 }
        offset: { type: integer, format: int32 }
    PaginatedBookmarksResponseDTO:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/BookmarkResponseDTO'
        total: { type: integer, format: int32 }
        limit: { type: integer, format: int32 }
        offset: { type: integer, format: int32 }
    PaginatedTracksResponseDTO:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/AudioTrackResponseDTO'
        total: { type: integer, format: int32 }
        limit: { type: integer, format: int32 }
        offset: { type: integer, format: int32 }
    PaginatedCollectionsResponseDTO:
      # TODO: Define schema
      type: object
      # ...
    ErrorResponseDTO:
      type: object
      required: [code, message]
      properties:
        code:
          type: string
          description: An application-specific error code.
          example: "INVALID_INPUT"
        message:
          type: string
          description: A human-readable error message.
          example: "Field 'email' failed validation on the 'required' rule"
        requestId:
          type: string
          description: The unique ID of the request, useful for tracing logs.
          example: "req_zX8f..."

  parameters:
    LimitParam:
      name: limit
      in: query
      description: Maximum number of items to return per page.
      required: false
      schema:
        type: integer
        format: int32
        minimum: 1
        maximum: 100 # Set a reasonable max limit
        default: 20
    OffsetParam:
      name: offset
      in: query
      description: Number of items to skip for pagination.
      required: false
      schema:
        type: integer
        format: int32
        minimum: 0
        default: 0

  responses: # Define reusable responses
    ErrorResponse:
      description: Generic error response. Status code will vary (e.g., 500).
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponseDTO'
    ValidationErrorResponse:
      description: Invalid input provided by the client.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponseDTO'
            example:
              code: "INVALID_INPUT"
              message: "validation failed: Field 'email' failed validation on the 'required' rule"
              requestId: "req_123abc"
    UnauthorizedErrorResponse:
      description: Authentication failed or is required. Missing or invalid credentials (e.g., JWT).
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponseDTO'
            example:
              code: "UNAUTHENTICATED"
              message: "Authentication required. Please log in."
              requestId: "req_123abc"
    ForbiddenErrorResponse:
      description: The authenticated user does not have permission to perform the requested action.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponseDTO'
            example:
              code: "FORBIDDEN"
              message: "You do not have permission to perform this action."
              requestId: "req_123abc"
    NotFoundErrorResponse:
      description: The requested resource was not found.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponseDTO'
            example:
              code: "NOT_FOUND"
              message: "The requested resource was not found."
              requestId: "req_123abc"
    ConflictErrorResponse:
      description: The request could not be completed due to a conflict with the current state of the resource (e.g., resource already exists).
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponseDTO'
            example:
              code: "RESOURCE_CONFLICT"
              message: "Email already exists."
              requestId: "req_123abc"

  securitySchemes:
    BearerAuth: # Define the security scheme used
      type: http
      scheme: bearer
      bearerFormat: JWT # Optional, just documentation
      description: "Enter JWT Bearer token **_only_**"
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
                            "$ref": "#/definitions/dto.PaginatedTracksResponseDTO"
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
                "description": "After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Complete audio upload and create track metadata",
                "operationId": "complete-audio-upload",
                "parameters": [
                    {
                        "description": "Track metadata and object key",
                        "name": "completeUpload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CompleteUploadRequestDTO"
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
                        "description": "Invalid Input (e.g., validation errors, object key not found)",
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
                    "409": {
                        "description": "Conflict (e.g., object key already used)\" // Depending on use case logic",
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
        "/audio/tracks/{trackId}": {
            "get": {
                "description": "Retrieves details for a specific audio track, including metadata and a temporary playback URL.",
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
                "description": "Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning a JWT.",
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
                        "description": "Authentication successful, returns JWT. isNewUser field indicates if a new account was created.",
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
                "description": "Authenticates a user with email and password, returns a JWT token.",
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
                        "description": "Login successful, returns JWT",
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
                        "description": "Registration successful, returns JWT\" // Success response: code, type, description",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input\"                // Failure response: code, type, description",
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
        "/bookmarks": {
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
                            "$ref": "#/definitions/dto.PaginatedBookmarksResponseDTO"
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
        "/bookmarks/{bookmarkId}": {
            "delete": {
                "security": [
                    {
                        "BearerAuth // Apply the security definition defined in main.go": []
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
                            "$ref": "#/definitions/dto.PaginatedProgressResponseDTO"
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
                    "description": "Include full track details if needed by frontend",
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
                    "description": "Use time.Time, will marshal to RFC3339",
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "CORRECTED: Use milliseconds",
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
                    "description": "Domain type maps to string here",
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
                    "description": "Use string UUID",
                    "type": "string"
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
                    "description": "Use time.Time, will marshal to RFC3339",
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "CORRECTED: Use milliseconds",
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
                    "description": "Domain type maps to string here",
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
                    "description": "Use string UUID",
                    "type": "string"
                }
            }
        },
        "dto.AuthResponseDTO": {
            "type": "object",
            "properties": {
                "isNewUser": {
                    "description": "Pointer, only included for Google callback if user is new",
                    "type": "boolean"
                },
                "token": {
                    "description": "The JWT access token",
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
                    "description": "CORRECTED: Use milliseconds",
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
        "dto.CompleteUploadRequestDTO": {
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
                    "description": "Optional note",
                    "type": "string"
                },
                "timestampMs": {
                    "description": "CORRECTED: Use milliseconds (int64)",
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
                    "description": "Add validation for slice elements",
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
                    "description": "Matches domain.CollectionType",
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
        "dto.PaginatedBookmarksResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BookmarkResponseDTO"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "totalPages": {
                    "type": "integer"
                }
            }
        },
        "dto.PaginatedProgressResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.PlaybackProgressResponseDTO"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "totalPages": {
                    "type": "integer"
                }
            }
        },
        "dto.PaginatedTracksResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "totalPages": {
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
                    "description": "CORRECTED: Use milliseconds",
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
                    "description": "CORRECTED: Use milliseconds (int64)",
                    "type": "integer",
                    "minimum": 0
                },
                "trackId": {
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
                    "description": "Add validation",
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
                            "$ref": "#/definitions/dto.PaginatedTracksResponseDTO"
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
                "description": "After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Uploads"
                ],
                "summary": "Complete audio upload and create track metadata",
                "operationId": "complete-audio-upload",
                "parameters": [
                    {
                        "description": "Track metadata and object key",
                        "name": "completeUpload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CompleteUploadRequestDTO"
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
                        "description": "Invalid Input (e.g., validation errors, object key not found)",
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
                    "409": {
                        "description": "Conflict (e.g., object key already used)\" // Depending on use case logic",
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
        "/audio/tracks/{trackId}": {
            "get": {
                "description": "Retrieves details for a specific audio track, including metadata and a temporary playback URL.",
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
                "description": "Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning a JWT.",
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
                        "description": "Authentication successful, returns JWT. isNewUser field indicates if a new account was created.",
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
                "description": "Authenticates a user with email and password, returns a JWT token.",
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
                        "description": "Login successful, returns JWT",
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
                        "description": "Registration successful, returns JWT\" // Success response: code, type, description",
                        "schema": {
                            "$ref": "#/definitions/dto.AuthResponseDTO"
                        }
                    },
                    "400": {
                        "description": "Invalid Input\"                // Failure response: code, type, description",
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
        "/bookmarks": {
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
                            "$ref": "#/definitions/dto.PaginatedBookmarksResponseDTO"
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
        "/bookmarks/{bookmarkId}": {
            "delete": {
                "security": [
                    {
                        "BearerAuth // Apply the security definition defined in main.go": []
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
                            "$ref": "#/definitions/dto.PaginatedProgressResponseDTO"
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
                    "description": "Include full track details if needed by frontend",
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
                    "description": "Use time.Time, will marshal to RFC3339",
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "CORRECTED: Use milliseconds",
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
                    "description": "Domain type maps to string here",
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
                    "description": "Use string UUID",
                    "type": "string"
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
                    "description": "Use time.Time, will marshal to RFC3339",
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "durationMs": {
                    "description": "CORRECTED: Use milliseconds",
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
                    "description": "Domain type maps to string here",
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
                    "description": "Use string UUID",
                    "type": "string"
                }
            }
        },
        "dto.AuthResponseDTO": {
            "type": "object",
            "properties": {
                "isNewUser": {
                    "description": "Pointer, only included for Google callback if user is new",
                    "type": "boolean"
                },
                "token": {
                    "description": "The JWT access token",
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
                    "description": "CORRECTED: Use milliseconds",
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
        "dto.CompleteUploadRequestDTO": {
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
                    "description": "Optional note",
                    "type": "string"
                },
                "timestampMs": {
                    "description": "CORRECTED: Use milliseconds (int64)",
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
                    "description": "Add validation for slice elements",
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
                    "description": "Matches domain.CollectionType",
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
        "dto.PaginatedBookmarksResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.BookmarkResponseDTO"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "totalPages": {
                    "type": "integer"
                }
            }
        },
        "dto.PaginatedProgressResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.PlaybackProgressResponseDTO"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "totalPages": {
                    "type": "integer"
                }
            }
        },
        "dto.PaginatedTracksResponseDTO": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.AudioTrackResponseDTO"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "totalPages": {
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
                    "description": "CORRECTED: Use milliseconds",
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
                    "description": "CORRECTED: Use milliseconds (int64)",
                    "type": "integer",
                    "minimum": 0
                },
                "trackId": {
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
                    "description": "Add validation",
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
        description: Include full track details if needed by frontend
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
        description: Use time.Time, will marshal to RFC3339
        type: string
      description:
        type: string
      durationMs:
        description: 'CORRECTED: Use milliseconds'
        type: integer
      id:
        type: string
      isPublic:
        type: boolean
      languageCode:
        type: string
      level:
        description: Domain type maps to string here
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
        description: Use string UUID
        type: string
    type: object
  dto.AudioTrackResponseDTO:
    properties:
      coverImageUrl:
        type: string
      createdAt:
        description: Use time.Time, will marshal to RFC3339
        type: string
      description:
        type: string
      durationMs:
        description: 'CORRECTED: Use milliseconds'
        type: integer
      id:
        type: string
      isPublic:
        type: boolean
      languageCode:
        type: string
      level:
        description: Domain type maps to string here
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
        description: Use string UUID
        type: string
    type: object
  dto.AuthResponseDTO:
    properties:
      isNewUser:
        description: Pointer, only included for Google callback if user is new
        type: boolean
      token:
        description: The JWT access token
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
        description: 'CORRECTED: Use milliseconds'
        type: integer
      trackId:
        type: string
      userId:
        type: string
    type: object
  dto.CompleteUploadRequestDTO:
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
        description: Optional note
        type: string
      timestampMs:
        description: 'CORRECTED: Use milliseconds (int64)'
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
        description: Add validation for slice elements
        items:
          type: string
        type: array
      title:
        maxLength: 255
        type: string
      type:
        description: Matches domain.CollectionType
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
  dto.PaginatedBookmarksResponseDTO:
    properties:
      data:
        items:
          $ref: '#/definitions/dto.BookmarkResponseDTO'
        type: array
      limit:
        type: integer
      offset:
        type: integer
      page:
        type: integer
      total:
        type: integer
      totalPages:
        type: integer
    type: object
  dto.PaginatedProgressResponseDTO:
    properties:
      data:
        items:
          $ref: '#/definitions/dto.PlaybackProgressResponseDTO'
        type: array
      limit:
        type: integer
      offset:
        type: integer
      page:
        type: integer
      total:
        type: integer
      totalPages:
        type: integer
    type: object
  dto.PaginatedTracksResponseDTO:
    properties:
      data:
        items:
          $ref: '#/definitions/dto.AudioTrackResponseDTO'
        type: array
      limit:
        type: integer
      offset:
        type: integer
      page:
        type: integer
      total:
        type: integer
      totalPages:
        type: integer
    type: object
  dto.PlaybackProgressResponseDTO:
    properties:
      lastListenedAt:
        type: string
      progressMs:
        description: 'CORRECTED: Use milliseconds'
        type: integer
      trackId:
        type: string
      userId:
        type: string
    type: object
  dto.RecordProgressRequestDTO:
    properties:
      progressMs:
        description: 'CORRECTED: Use milliseconds (int64)'
        minimum: 0
        type: integer
      trackId:
        type: string
    required:
    - progressMs
    - trackId
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
        description: Add validation
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
            $ref: '#/definitions/dto.PaginatedTracksResponseDTO'
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
        record in the database.
      operationId: complete-audio-upload
      parameters:
      - description: Track metadata and object key
        in: body
        name: completeUpload
        required: true
        schema:
          $ref: '#/definitions/dto.CompleteUploadRequestDTO'
      produces:
      - application/json
      responses:
        "201":
          description: Track metadata created successfully
          schema:
            $ref: '#/definitions/dto.AudioTrackResponseDTO'
        "400":
          description: Invalid Input (e.g., validation errors, object key not found)
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "401":
          description: Unauthorized
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "409":
          description: Conflict (e.g., object key already used)" // Depending on use
            case logic
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      security:
      - BearerAuth: []
      summary: Complete audio upload and create track metadata
      tags:
      - Uploads
  /audio/tracks/{trackId}:
    get:
      description: Retrieves details for a specific audio track, including metadata
        and a temporary playback URL.
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
        "404":
          description: Track Not Found
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
        "500":
          description: Internal Server Error
          schema:
            $ref: '#/definitions/httputil.ErrorResponseDTO'
      summary: Get audio track details
      tags:
      - Audio Tracks
  /auth/google/callback:
    post:
      consumes:
      - application/json
      description: Receives the ID token from the frontend after Google sign-in, verifies
        it, and performs user registration or login, returning a JWT.
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
          description: Authentication successful, returns JWT. isNewUser field indicates
            if a new account was created.
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
      description: Authenticates a user with email and password, returns a JWT token.
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
          description: Login successful, returns JWT
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
          description: 'Registration successful, returns JWT" // Success response:
            code, type, description'
          schema:
            $ref: '#/definitions/dto.AuthResponseDTO'
        "400":
          description: 'Invalid Input"                // Failure response: code, type,
            description'
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
  /bookmarks:
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
            $ref: '#/definitions/dto.PaginatedBookmarksResponseDTO'
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
  /bookmarks/{bookmarkId}:
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
      - BearerAuth // Apply the security definition defined in main.go: []
      summary: Delete a bookmark
      tags:
      - User Activity
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
            $ref: '#/definitions/dto.PaginatedProgressResponseDTO'
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

## `migrations/000002_create_audio_tables.down.sql`

```sql
-- migrations/000002_create_audio_tables.down.sql

DROP TRIGGER IF EXISTS update_audio_collections_updated_at ON audio_collections;
DROP TRIGGER IF EXISTS update_audio_tracks_updated_at ON audio_tracks;

DROP TABLE IF EXISTS collection_tracks;
DROP TABLE IF EXISTS audio_collections;
DROP TABLE IF EXISTS audio_tracks;
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

## `migrations/000003_create_activity_tables.down.sql`

```sql
-- migrations/000003_create_activity_tables.down.sql

DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS playback_progress;
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

## `internal/config/config.go`

```go
// internal/config/config.go
package config

import (
	"fmt"
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
	DSN             string        `mapstructure:"dsn"` // Data Source Name (e.g., postgresql://user:password@host:port/dbname?sslmode=disable)
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
}

// JWTConfig holds JWT related configuration.
type JWTConfig struct {
	SecretKey         string        `mapstructure:"secretKey"`
	AccessTokenExpiry time.Duration `mapstructure:"accessTokenExpiry"`
	// RefreshTokenExpiry time.Duration `mapstructure:"refreshTokenExpiry"` // Add if implementing refresh tokens
}

// MinioConfig holds MinIO connection configuration.
type MinioConfig struct {
	Endpoint        string        `mapstructure:"endpoint"`
	AccessKeyID     string        `mapstructure:"accessKeyId"`
	SecretAccessKey string        `mapstructure:"secretAccessKey"`
	UseSSL          bool          `mapstructure:"useSsl"`
	BucketName      string        `mapstructure:"bucketName"`
	PresignExpiry   time.Duration `mapstructure:"presignExpiry"` // Default expiry for presigned URLs
}

// GoogleConfig holds Google OAuth configuration.
type GoogleConfig struct {
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
	// RedirectURL string `mapstructure:"redirectUrl"` // Usually handled by frontend, but maybe needed later
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"` // e.g., "debug", "info", "warn", "error"
	JSON  bool   `mapstructure:"json"`  // Output logs in JSON format
}

// CorsConfig holds CORS configuration.
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins"`
	AllowedMethods   []string `mapstructure:"allowedMethods"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"`
}


// LoadConfig reads configuration from file or environment variables.
// It looks for config files in the specified path (e.g., "." for current directory).
func LoadConfig(path string) (config Config, err error) {
	v := viper.New()

	// Set default values first
	setDefaultValues(v)

	// Configure Viper
	v.AddConfigPath(path)         // Path to look for the config file in
	v.SetConfigType("yaml")
	v.AutomaticEnv()              // Read in environment variables that match
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores for env var names (e.g., server.port -> SERVER_PORT)

	// 1. Load base configuration (config.yaml)
	v.SetConfigName("config")
	if errReadBase := v.ReadInConfig(); errReadBase != nil {
		if _, ok := errReadBase.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a 'file not found' error for the base file
			return config, fmt.Errorf("error reading base config file 'config.yaml': %w", errReadBase)
		}
		// Base file not found is okay, we can proceed without it.
		fmt.Fprintln(os.Stderr, "Info: Base config file 'config.yaml' not found. Skipping.")
	}

	// 2. Load and merge environment-specific configuration (e.g., config.development.yaml)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // Default environment
	}
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	if errReadEnv := v.MergeInConfig(); errReadEnv != nil {
		if _, ok := errReadEnv.(viper.ConfigFileNotFoundError); ok {
			// Env specific file not found, rely on base/env vars/defaults
			fmt.Fprintf(os.Stderr, "Info: Environment config file 'config.%s.yaml' not found. Relying on base/defaults/env.\n", env)
		} else {
			// Other error reading env config file
			return config, fmt.Errorf("error merging config file 'config.%s.yaml': %w", env, errReadEnv)
		}
	}

	// Unmarshal the final merged configuration
	err = v.Unmarshal(&config)
	if err != nil {
		 return config, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Optional: Add validation logic here if needed
    // For example, check if required fields like database DSN are set:
    if config.Database.DSN == "" {
        // Attempt to get from DATABASE_URL env var as a fallback
        envDSN := os.Getenv("DATABASE_URL")
        if envDSN != "" {
            fmt.Fprintln(os.Stderr, "Info: Database DSN not found in config, using DATABASE_URL environment variable.")
            config.Database.DSN = envDSN
        } else {
            // If still empty after checking env var, return error
            return config, fmt.Errorf("database DSN is required but not found in config or DATABASE_URL environment variable")
        }
    }

	return config, nil
}

// Use a dedicated viper instance for defaults to avoid conflicts
func setDefaultValues(v *viper.Viper) {
    // Server Defaults
    v.SetDefault("server.port", "8080")
    v.SetDefault("server.readTimeout", "5s")
    v.SetDefault("server.writeTimeout", "10s")
    v.SetDefault("server.idleTimeout", "120s")

    // Database Defaults (Note: DSN has no sensible default)
    v.SetDefault("database.maxOpenConns", 25)
    v.SetDefault("database.maxIdleConns", 25)
    v.SetDefault("database.connMaxLifetime", "5m")
	v.SetDefault("database.connMaxIdleTime", "5m") // Usually same as lifetime

    // JWT Defaults
	v.SetDefault("jwt.secretKey", "default-insecure-secret-key-please-override") // Provide a default but stress it's insecure
    v.SetDefault("jwt.accessTokenExpiry", "1h")

    // MinIO Defaults
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.accessKeyId", "minioadmin")
	v.SetDefault("minio.secretAccessKey", "minioadmin")
    v.SetDefault("minio.useSsl", false)
    v.SetDefault("minio.presignExpiry", "1h") // Default URL expiry

	// Google Defaults
	v.SetDefault("google.clientId", "")
	v.SetDefault("google.clientSecret", "")

    // Log Defaults
    v.SetDefault("log.level", "info")
    v.SetDefault("log.json", false) // Easier to read logs for dev default

	// CORS Defaults (Example: Allow local dev server)
	v.SetDefault("cors.allowedOrigins", []string{"http://localhost:3000", "http://127.0.0.1:3000"})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	v.SetDefault("cors.allowCredentials", true)
	v.SetDefault("cors.maxAge", 300) // 5 minutes
}

// GetConfig is a helper function to load config, often called from main.
// Assumes config files are in the current working directory (".")
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
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
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
// @Success 200 {object} dto.PaginatedProgressResponseDTO "Paginated list of playback progress (progressMs in milliseconds)"
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
	limit, _ := strconv.Atoi(q.Get("limit")) // Use 0 if parsing fails
	offset, _ := strconv.Atoi(q.Get("offset")) // Use 0 if parsing fails

	// Create pagination parameters and apply defaults/constraints
	pageParams := pagination.NewPageFromOffset(limit, offset)

	// Create use case parameters struct
	ucParams := port.ListProgressParams{
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

// CreateBookmark handles POST /api/v1/bookmarks
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
// @Router /bookmarks [post]
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

// ListBookmarks handles GET /api/v1/bookmarks
// @Summary List user's bookmarks
// @Description Retrieves a paginated list of bookmarks for the authenticated user, optionally filtered by track ID.
// @ID list-bookmarks
// @Tags User Activity
// @Produce json
// @Security BearerAuth
// @Param trackId query string false "Filter by Audio Track UUID" Format(uuid)
// @Param limit query int false "Pagination limit" default(50) minimum(1) maximum(100)
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Success 200 {object} dto.PaginatedBookmarksResponseDTO "Paginated list of bookmarks (timestampMs in milliseconds)"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format (if provided)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /bookmarks [get]
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
	ucParams := port.ListBookmarksParams{
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

// DeleteBookmark handles DELETE /api/v1/bookmarks/{bookmarkId}
// @Summary Delete a bookmark
// @Description Deletes a specific bookmark owned by the current user.
// @ID delete-bookmark
// @Tags User Activity
// @Produce json
// @Security BearerAuth // Apply the security definition defined in main.go
// @Param bookmarkId path string true "Bookmark UUID" Format(uuid)
// @Success 204 "Bookmark deleted successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Bookmark Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /bookmarks/{bookmarkId} [delete]
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

## `internal/adapter/handler/http/upload_handler.go`

```go
// internal/adapter/handler/http/upload_handler.go
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
)

// UploadHandler handles HTTP requests related to file uploads.
type UploadHandler struct {
	uploadUseCase UploadUseCase // Use interface defined below
	validator     *validation.Validator
}

// UploadUseCase defines the methods expected from the use case layer.
// Input Port for this handler.
type UploadUseCase interface {
	RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*usecase.RequestUploadResult, error)
	CompleteUpload(ctx context.Context, userID domain.UserID, req usecase.CompleteUploadRequest) (*domain.AudioTrack, error)
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uc UploadUseCase, v *validation.Validator) *UploadHandler {
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

	var req dto.RequestUploadRequestDTO // Define this DTO below
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	result, err := h.uploadUseCase.RequestUpload(r.Context(), userID, req.Filename, req.ContentType)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles InvalidArgument, internal errors
		return
	}

	// Map result to response DTO
	resp := dto.RequestUploadResponseDTO{ // Define this DTO below
		UploadURL: result.UploadURL,
		ObjectKey: result.ObjectKey,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// CompleteUploadAndCreateTrack handles POST /api/v1/audio/tracks
// @Summary Complete audio upload and create track metadata
// @Description After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database.
// @ID complete-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param completeUpload body dto.CompleteUploadRequestDTO true "Track metadata and object key"
// @Success 201 {object} dto.AudioTrackResponseDTO "Track metadata created successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., validation errors, object key not found)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict (e.g., object key already used)" // Depending on use case logic
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [post] // Reuses POST on /audio/tracks conceptually
func (h *UploadHandler) CompleteUploadAndCreateTrack(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.CompleteUploadRequestDTO // Define this DTO below
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO to Usecase request struct
	ucReq := usecase.CompleteUploadRequest{
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

	track, err := h.uploadUseCase.CompleteUpload(r.Context(), userID, ucReq)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles PermissionDenied, InvalidArgument, Conflict, internal errors
		return
	}

	// Return the newly created track details
	resp := dto.MapDomainTrackToResponseDTO(track) // Reuse existing mapping
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // 201 Created
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
	// REMOVED: "time" - No longer needed directly in this file after changes
	"github.com/go-chi/chi/v5"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	// REMOVED: "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" - Not used directly
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
)

// AudioHandler handles HTTP requests related to audio tracks and collections.
type AudioHandler struct {
	audioUseCase port.AudioContentUseCase // 使用port包中定义的接口
	validator    *validation.Validator
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
// @Description Retrieves details for a specific audio track, including metadata and a temporary playback URL.
// @ID get-track-details
// @Tags Audio Tracks
// @Produce json
// @Param trackId path string true "Audio Track UUID" Format(uuid)
// @Success 200 {object} dto.AudioTrackDetailsResponseDTO "Audio track details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format"
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

	track, playURL, err := h.audioUseCase.GetAudioTrackDetails(r.Context(), trackID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, internal errors
		return
	}

	// Map to DTO using the function (which now correctly handles domain types)
	resp := dto.AudioTrackDetailsResponseDTO{
		AudioTrackResponseDTO: dto.MapDomainTrackToResponseDTO(track),
		PlayURL:               playURL,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
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
// @Success 200 {object} dto.PaginatedTracksResponseDTO "Paginated list of audio tracks"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Query Parameter Format"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [get]
func (h *AudioHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ucParams := port.UseCaseListTracksParams{} // Use UseCaseListTracksParams

	// Parse simple string filters
	if query := q.Get("q"); query != "" { ucParams.Query = &query }
	if lang := q.Get("lang"); lang != "" { ucParams.LanguageCode = &lang }
	ucParams.Tags = q["tags"] // Get array of tags

	// Parse Level (and validate)
	if levelStr := q.Get("level"); levelStr != "" {
		level := domain.AudioLevel(levelStr)
		if level.IsValid() {
			ucParams.Level = &level
		} else {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid level query parameter '%s'", domain.ErrInvalidArgument, levelStr))
			return
		}
	}

	// Parse isPublic (boolean)
	if isPublicStr := q.Get("isPublic"); isPublicStr != "" {
		isPublic, err := strconv.ParseBool(isPublicStr)
		if err == nil {
			ucParams.IsPublic = &isPublic
		} else {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid isPublic query parameter (must be true or false)", domain.ErrInvalidArgument))
			return
		}
	}

	// Parse Sort parameters
	ucParams.SortBy = q.Get("sortBy")
	ucParams.SortDirection = q.Get("sortDir")

	// Parse Pagination parameters
	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")
	limit, errLimit := strconv.Atoi(limitStr)
	offset, errOffset := strconv.Atoi(offsetStr)
	if (limitStr != "" && errLimit != nil) || (offsetStr != "" && errOffset != nil) {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid limit or offset query parameter", domain.ErrInvalidArgument))
		return
	}
	ucParams.Page = pagination.NewPageFromOffset(limit, offset)


	// Call use case with the single params struct
	tracks, total, actualPageInfo, err := h.audioUseCase.ListTracks(r.Context(), ucParams)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.AudioTrackResponseDTO, len(tracks))
	for i, track := range tracks {
		respData[i] = dto.MapDomainTrackToResponseDTO(track)
	}

	paginatedResult := pagination.NewPaginatedResponse(respData, total, actualPageInfo)
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

// --- Collection Handlers --- (Rest of the file remains the same)

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

	collectionType := domain.CollectionType(req.Type) // Already validated by tag

	collection, err := h.audioUseCase.CreateCollection(r.Context(), req.Title, req.Description, collectionType, initialTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles Unauthenticated, internal errors
		return
	}

	// Map response without tracks initially (as CreateCollection might return before tracks are fully associated if error occurred)
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

	// Get collection metadata
	collection, err := h.audioUseCase.GetCollectionDetails(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied
		return
	}

	// Get associated track details (ordered)
	tracks, err := h.audioUseCase.GetCollectionTracks(r.Context(), collectionID)
	if err != nil {
		slog.Default().ErrorContext(r.Context(), "Failed to fetch tracks for collection details", "error", err, "collectionID", collectionID)
		httputil.RespondError(w, r, fmt.Errorf("failed to retrieve collection tracks: %w", err))
		return
	}

	// Map domain object and tracks to response DTO
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
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful update with no body
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

	// Validate the request DTO (e.g., check if track IDs are valid UUIDs)
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Convert string IDs to domain IDs
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
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated, InvalidArgument (bad track id)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
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
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
```

## `internal/adapter/handler/http/auth_handler.go`

```go
// internal/adapter/handler/http/auth_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" 
	"github.com/yvanyang/language-learning-player-backend/internal/port"   
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto" 
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"    
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"  
)

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	authUseCase port.AuthUseCase // Use interface defined in port package
	validator   *validation.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc port.AuthUseCase, v *validation.Validator) *AuthHandler {
	return &AuthHandler{
		authUseCase: uc,
		validator:   v,
	}
}

// Register handles user registration requests.
// @Summary Register a new user
// @Description Registers a new user account using email and password.
// @ID register-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequestDTO true "User Registration Info" // Input parameter: name, in, type, required, description
// @Success 201 {object} dto.AuthResponseDTO "Registration successful, returns JWT" // Success response: code, type, description
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"                // Failure response: code, type, description
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict - Email Exists"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/register [post]  // Route path and method
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	// Validate input DTO
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Call use case
	_, token, err := h.authUseCase.RegisterWithPassword(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		// UseCase should return domain errors, let RespondError map them
		httputil.RespondError(w, r, err)
		return
	}

	// Return JWT token
	resp := dto.AuthResponseDTO{Token: token}
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // 201 Created for successful registration
}

// Login handles user login requests.
// @Summary Login a user
// @Description Authenticates a user with email and password, returns a JWT token.
// @ID login-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body dto.LoginRequestDTO true "User Login Credentials"
// @Success 200 {object} dto.AuthResponseDTO "Login successful, returns JWT"
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

	// Validate input DTO
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Call use case
	token, err := h.authUseCase.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		// Handles domain.ErrAuthenticationFailed, domain.ErrNotFound (mapped to auth failed), etc.
		httputil.RespondError(w, r, err)
		return
	}

	// Return JWT token
	resp := dto.AuthResponseDTO{Token: token}
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// GoogleCallback handles the callback from Google OAuth flow.
// @Summary Handle Google OAuth callback
// @Description Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning a JWT.
// @ID google-callback
// @Tags Authentication
// @Accept json
// @Produce json
// @Param googleCallback body dto.GoogleCallbackRequestDTO true "Google ID Token"
// @Success 200 {object} dto.AuthResponseDTO "Authentication successful, returns JWT. isNewUser field indicates if a new account was created."
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

	token, isNew, err := h.authUseCase.AuthenticateWithGoogle(r.Context(), req.IDToken)
	if err != nil {
		// Handles domain.ErrAuthenticationFailed, domain.ErrConflict, etc.
		httputil.RespondError(w, r, err)
		return
	}

	resp := dto.AuthResponseDTO{Token: token}
	// Only include isNewUser field if it's true, otherwise omit it
	if isNew {
		isNewPtr := true
		resp.IsNewUser = &isNewPtr
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}
```

## `internal/adapter/handler/http/auth_handler_test.go`

```go
// internal/adapter/handler/http/auth_handler_test.go
package http_test // Use _test package

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5" // Need chi for routing context if using URL params
	adapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias for handler package
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	ucmocks "github.com/yvanyang/language-learning-player-backend/internal/usecase/mocks" // Mocks for usecase interfaces
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper to create a request and response recorder for handler tests
func setupHandlerTest(method, path string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
    var reqBody *bytes.Buffer = nil
    if body != nil {
        b, _ := json.Marshal(body)
        reqBody = bytes.NewBuffer(b)
    }

    req := httptest.NewRequest(method, path, reqBody)
    if reqBody != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    // Add context if needed (e.g., for middleware values like UserID)
    // ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID)
    // req = req.WithContext(ctx)

    rr := httptest.NewRecorder()
    return req, rr
}


func TestAuthHandler_Register_Success(t *testing.T) {
    mockAuthUC := ucmocks.NewMockAuthUseCase(t) // Using mock for the usecase INPUT PORT interface
    validator := validation.New()
    handler := adapter.NewAuthHandler(mockAuthUC, validator) // Instantiate the actual handler

    reqBody := dto.RegisterRequestDTO{
        Email:    "new@example.com",
        Password: "password123",
        Name:     "New User",
    }
    expectedToken := "new_jwt_token"
    dummyUser := &domain.User{ID: domain.NewUserID(), Email: domain.Email{}} // Dummy user for return

    // Expect the usecase method to be called
    mockAuthUC.On("RegisterWithPassword", mock.Anything, reqBody.Email, reqBody.Password, reqBody.Name).
        Return(dummyUser, expectedToken, nil). // Return success
        Once()

    req, rr := setupHandlerTest(http.MethodPost, "/api/v1/auth/register", reqBody)

    // Serve the request
    handler.Register(rr, req)

    // Assert the response
    require.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")

    var respBody dto.AuthResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &respBody)
    require.NoError(t, err, "Failed to unmarshal response body")
    assert.Equal(t, expectedToken, respBody.Token)

    mockAuthUC.AssertExpectations(t) // Verify usecase was called
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
    mockAuthUC := ucmocks.NewMockAuthUseCase(t)
    validator := validation.New()
    handler := adapter.NewAuthHandler(mockAuthUC, validator)

    reqBody := dto.RegisterRequestDTO{ // Invalid email
        Email:    "invalid-email",
        Password: "password123",
        Name:     "New User",
    }

    req, rr := setupHandlerTest(http.MethodPost, "/api/v1/auth/register", reqBody)
    handler.Register(rr, req)

    // Assert validation error response
    require.Equal(t, http.StatusBadRequest, rr.Code)
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "INVALID_INPUT", errResp.Code)
    assert.Contains(t, errResp.Message, "'email' failed validation")

    // Ensure usecase was NOT called
    mockAuthUC.AssertNotCalled(t, "RegisterWithPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_Register_UseCaseConflictError(t *testing.T) {
    mockAuthUC := ucmocks.NewMockAuthUseCase(t)
    validator := validation.New()
    handler := adapter.NewAuthHandler(mockAuthUC, validator)

    reqBody := dto.RegisterRequestDTO{
        Email:    "exists@example.com",
        Password: "password123",
        Name:     "Existing User",
    }
    conflictError := fmt.Errorf("%w: email already exists", domain.ErrConflict)

    // Expect usecase to be called and return conflict error
    mockAuthUC.On("RegisterWithPassword", mock.Anything, reqBody.Email, reqBody.Password, reqBody.Name).
        Return(nil, "", conflictError).
        Once()

    req, rr := setupHandlerTest(http.MethodPost, "/api/v1/auth/register", reqBody)
    handler.Register(rr, req)

    // Assert conflict response
    require.Equal(t, http.StatusConflict, rr.Code) // 409 Conflict
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "RESOURCE_CONFLICT", errResp.Code)

    mockAuthUC.AssertExpectations(t)
}

// TODO: Add tests for Login handler (success, auth failed, validation error)
// TODO: Add tests for GoogleCallback handler (success new user, success existing, failure)
```

## `internal/adapter/handler/http/audio_handler_test.go`

```go
package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	httpadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port" // Import port for interfaces
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
)

// --- Mock UseCase ---
// Use Mockery or manual mock based on port interface
type MockAudioContentUseCase struct {
	mock.Mock
	port.AudioContentUseCase // Embed the interface
}

// Implement only the methods needed for the test, forwarding calls to the mock object
func (m *MockAudioContentUseCase) ListTracks(ctx context.Context, params port.ListTracksParams, page port.Page) ([]*domain.AudioTrack, int, error) {
	args := m.Called(ctx, params, page)
	// Handle nil return for the slice pointer correctly
	if args.Get(0) == nil {
		// Return empty slice, total count, and error
		return []*domain.AudioTrack{}, args.Int(1), args.Error(2)
	}
	// Return tracks, total count, and error
	return args.Get(0).([]*domain.AudioTrack), args.Int(1), args.Error(2)
}

// Correct signatures based on audio_handler.go usage
func (m *MockAudioContentUseCase) GetTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error) {
	args := m.Called(ctx, trackID)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.AudioTrack), args.String(1), args.Error(2)
}

func (m *MockAudioContentUseCase) CreateCollection(ctx context.Context, title string, description string, collectionType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	args := m.Called(ctx, title, description, collectionType, initialTrackIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AudioCollection), args.Error(1)
}

func (m *MockAudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	args := m.Called(ctx, collectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AudioCollection), args.Error(1)
}

func (m *MockAudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	args := m.Called(ctx, collectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AudioTrack), args.Error(1)
}

func (m *MockAudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title string, description string) error {
	args := m.Called(ctx, collectionID, title, description)
	return args.Error(0)
}

func (m *MockAudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	args := m.Called(ctx, collectionID)
	return args.Error(0)
}

func (m *MockAudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	args := m.Called(ctx, collectionID, orderedTrackIDs)
	return args.Error(0)
}

// --- Test Function ---
func TestListTracks(t *testing.T) {
	validator := validation.New() // Use real validator

	// Define a sample domain object for use case response
	trackID, _ := domain.TrackIDFromString("uuid-track-1") // Use TrackIDFromString, ignore error for test setup
	sampleTrackDomain := &domain.AudioTrack{
		ID:            trackID,
		Title:         "Sample Track",
		Language:      "en-US",        // CHANGE: Correct field name
		Level:         "B1",
		Duration:      120 * time.Second, // CHANGE: Correct field name and type
		CoverImageURL: nil,
		IsPublic:      true,
		// Need MinioBucket and MinioObjectKey if MapDomainTrackToResponseDTO requires them
	}
	// Define the corresponding response DTO
	sampleTrackDTO := dto.AudioTrackResponseDTO{
		ID:            "uuid-track-1",
		Title:         "Sample Track",
		LanguageCode:  "en-US", // DTO field might still be LanguageCode
		Level:         "B1",
		DurationMs:    120000,  // DTO field might still be DurationMs
		CoverImageURL: nil,
		IsPublic:      true,
	}

	t.Run("Success - List tracks with pagination and filter", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase)
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)


		// Prepare expected response from use case (domain objects and total count)
		expectedTracks := []*domain.AudioTrack{sampleTrackDomain}
		expectedTotal := 1

		// Set expectations on the mock
		isPublicValue := true
		// Ensure pointers match the types in port.ListTracksParams
		langParam := string(sampleTrackDomain.Language) // Convert if necessary
		levelParam := domain.AudioLevel(sampleTrackDomain.Level)
		expectedParams := port.ListTracksParams{
			LanguageCode: &langParam, // CHANGE: Use pointer to string or correct type
			Level:        &levelParam,
			IsPublic:     &isPublicValue,
		}
		expectedPage := port.Page{Limit: 10, Offset: 0} // CHANGE: Use port.Page
		// Use mock.Anything for context, specific types for others
		mockUseCase.On("ListTracks", mock.Anything, expectedParams, expectedPage).Return(expectedTracks, expectedTotal, nil).Once()

		// Prepare HTTP request
		// Ensure query params match expectedParams fields
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks?lang=en-US&level=B1&limit=10&offset=0&isPublic=true", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusOK, rr.Code)

		var actualResponse dto.PaginatedTracksResponseDTO
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)

		// Assertions on paginated fields
		assert.Equal(t, expectedTotal, actualResponse.Total)
		assert.Equal(t, expectedPage.Limit, actualResponse.Limit)
		assert.Equal(t, expectedPage.Offset, actualResponse.Offset)
		assert.Len(t, actualResponse.Data, 1) // Check number of items

		// Optional: Deeper check of the actual data requires unmarshalling the interface{} slice items
		if len(actualResponse.Data) > 0 {
				// Need to handle the case where data is map[string]interface{} after json.Unmarshal
				var firstItemMap map[string]interface{}
				firstItemBytes, _ := json.Marshal(actualResponse.Data[0])
				err = json.Unmarshal(firstItemBytes, &firstItemMap)
				assert.NoError(t, err)

				// Compare map fields or remarshal to the expected DTO type
				var actualTrackDTO dto.AudioTrackResponseDTO
				err = json.Unmarshal(firstItemBytes, &actualTrackDTO)
				assert.NoError(t, err)
				assert.Equal(t, sampleTrackDTO, actualTrackDTO)
		}


		mockUseCase.AssertExpectations(t)
	})

	t.Run("Failure - UseCase returns error", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase)
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)

		expectedError := errors.New("internal database error")
		// Use mock.AnythingOfType for struct parameters if precise matching is difficult or not needed
		mockUseCase.On("ListTracks", mock.Anything, mock.AnythingOfType("port.ListTracksParams"), mock.AnythingOfType("port.Page")).Return(nil, 0, expectedError).Once() // CHANGE: Use port.Page

		// Prepare HTTP request
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var errResponse httputil.ErrorResponseDTO // CHANGE: Use correct DTO type
		err := json.Unmarshal(rr.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", errResponse.Code)
		// assert.Contains(t, errResponse.Message, "Failed to list audio tracks") // Check message if needed

		mockUseCase.AssertExpectations(t)
	})

	t.Run("Failure - Invalid query parameter value", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase) // UseCase should not be called
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)

		// Prepare HTTP request with invalid limit
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks?limit=abc", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect 400 due to bad input

		var errResponse httputil.ErrorResponseDTO // CHANGE: Use correct DTO type
		err := json.Unmarshal(rr.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_INPUT", errResponse.Code)
		// assert.Contains(t, errResponse.Message, "Invalid value for parameter 'limit'") // Check message if needed

		// Assert that the use case method was NOT called
		mockUseCase.AssertNotCalled(t, "ListTracks", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Success - No query parameters (uses defaults)", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase)
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)

		// Prepare expected response for default parameters
		expectedTracks := []*domain.AudioTrack{}
		expectedTotal := 0

		// Set expectations on the mock with default params
		// Adjust default values based on actual implementation in handler
		expectedParams := port.ListTracksParams{
			// Default IsPublic might be nil depending on handler logic
			IsPublic: nil, // Assuming nil means no filter by default
		}
		// Assuming default limit=20, offset=0 based on handler code
		expectedPage := port.Page{Limit: 20, Offset: 0} // CHANGE: Use port.Page
		mockUseCase.On("ListTracks", mock.Anything, expectedParams, expectedPage).Return(expectedTracks, expectedTotal, nil).Once()

		// Prepare HTTP request with no query params
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusOK, rr.Code)

		var actualResponse dto.PaginatedTracksResponseDTO
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)

		assert.Equal(t, expectedTotal, actualResponse.Total)
		assert.Equal(t, expectedPage.Limit, actualResponse.Limit)
		assert.Equal(t, expectedPage.Offset, actualResponse.Offset)
		assert.Len(t, actualResponse.Data, 0)

		mockUseCase.AssertExpectations(t)
	})

	// Add more test cases as needed.
}

// Ensure necessary DTOs and Param structs are defined:
// - internal/adapter/handler/http/dto/audio_dto.go needs AudioTrackResponseDTO and PaginatedTracksResponseDTO
// - internal/port/usecase.go needs AudioContentUseCase interface definition
// - internal/port/usecase.go or internal/port/repository.go needs ListTracksParams struct definition
// - internal/port/port.go (or similar) needs Page struct definition
```

## `internal/adapter/handler/http/user_handler.go`

```go
// internal/adapter/handler/http/user_handler.go
package http

import (
	// REMOVED: "context" - Not needed directly here
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto" // Import dto package
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/internal/port" // Import port package for UserUseCase interface
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

## `internal/adapter/handler/http/middleware/.keep`

```

```

## `internal/adapter/handler/http/middleware/ratelimit.go`

```go
package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil" // Adjust import path
    "github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
)

// IPRateLimiter stores rate limiters for IP addresses.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit // Allowed requests per second
	b   int        // Burst size
}

// NewIPRateLimiter creates a new IPRateLimiter.
// r: requests per second, b: burst size.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// AddIP creates a new rate limiter for the given IP address if it doesn't exist.
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}
	return limiter
}

// GetLimiter returns the rate limiter for the given IP address.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip) // Add if not exists (lazy initialization)
	}
	return limiter
}

// RateLimit is the middleware handler.
func RateLimit(limiter *IPRateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the client's real IP address.
			// Use chi's RealIP middleware result if available, otherwise fallback.
			ip := r.RemoteAddr
			if realIP := r.Context().Value(http.CanonicalHeaderKey("X-Real-IP")); realIP != nil {
                if ripStr, ok := realIP.(string); ok {
				    ip = ripStr
                }
			} else if forwardedFor := r.Context().Value(http.CanonicalHeaderKey("X-Forwarded-For")); forwardedFor != nil {
                if fwdStr, ok := forwardedFor.(string); ok {
                    // X-Forwarded-For can contain multiple IPs, get the first one
                    parts := strings.Split(fwdStr, ",")
                    if len(parts) > 0 {
                        ip = strings.TrimSpace(parts[0])
                    }
                }
			} else {
                // Fallback if headers are not present (e.g. direct connection)
                host, _, err := net.SplitHostPort(r.RemoteAddr)
                if err == nil {
                    ip = host
                }
            }


			// Get the rate limiter for the current IP address.
			l := limiter.GetLimiter(ip)

			// Check if the request is allowed.
			if !l.Allow() {
				// Log the rate limit event
				logger := slog.Default()
				logger.WarnContext(r.Context(), "Rate limit exceeded", "ip_address", ip)

				// Respond with 429 Too Many Requests
				// Use a custom error type or map directly
                err := fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied) // Or a new domain error type
                // Map this specific error text or type in RespondError if needed
                httputil.RespondError(w, r, err) // This might return 403, adjust mapping if needed
                // Or respond directly:
                // w.Header().Set("Retry-After", "60") // Optional: suggest retry time
                // http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CleanUpOldLimiters periodically removes limiters for IPs not seen recently.
// This prevents the map from growing indefinitely. Should be run in a separate goroutine.
// Note: This is a basic example; more sophisticated cleanup might be needed.
func (i *IPRateLimiter) CleanUpOldLimiters(interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		for ip, limiter := range i.ips {
			// This is tricky: rate.Limiter doesn't expose last access time.
			// A more complex solution would wrap the limiter and track last access.
			// Simple approach (less accurate): remove if unused for a while? Needs wrapping.
			// Placeholder: For now, this cleanup isn't effective without tracking last access.
            _ = limiter // Avoid unused variable error
            _ = ip
		}
		// Implement actual cleanup logic here based on tracking last access time.
		slog.Debug("Running rate limiter cleanup (placeholder)...", "current_size", len(i.ips))
		i.mu.Unlock()
	}
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
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
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

## `internal/adapter/handler/http/middleware/recovery.go`

```go
// internal/adapter/handler/http/middleware/recovery.go
package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
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

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path (for SecurityHelper interface)
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"    // Adjust import path
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

## `internal/adapter/handler/http/dto/activity_dto.go`

```go
// internal/adapter/handler/http/dto/activity_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
)

// --- Request DTOs ---

// RecordProgressRequestDTO defines the JSON body for recording playback progress.
type RecordProgressRequestDTO struct {
	TrackID    string `json:"trackId" validate:"required,uuid"`
	ProgressMs int64  `json:"progressMs" validate:"required,gte=0"` // CORRECTED: Use milliseconds (int64)
}

// CreateBookmarkRequestDTO defines the JSON body for creating a bookmark.
type CreateBookmarkRequestDTO struct {
	TrackID     string `json:"trackId" validate:"required,uuid"`
	TimestampMs int64  `json:"timestampMs" validate:"required,gte=0"` // CORRECTED: Use milliseconds (int64)
	Note        string `json:"note"`                                  // Optional note
}

// --- Response DTOs ---

// PlaybackProgressResponseDTO defines the JSON representation of playback progress.
type PlaybackProgressResponseDTO struct {
	UserID         string    `json:"userId"`
	TrackID        string    `json:"trackId"`
	ProgressMs     int64     `json:"progressMs"` // CORRECTED: Use milliseconds
	LastListenedAt time.Time `json:"lastListenedAt"`
}

// MapDomainProgressToResponseDTO converts domain progress to DTO.
func MapDomainProgressToResponseDTO(p *domain.PlaybackProgress) PlaybackProgressResponseDTO {
	return PlaybackProgressResponseDTO{
		UserID:         p.UserID.String(),
		TrackID:        p.TrackID.String(),
		ProgressMs:     p.Progress.Milliseconds(), // CORRECTED: Get milliseconds
		LastListenedAt: p.LastListenedAt,
	}
}

// BookmarkResponseDTO defines the JSON representation of a bookmark.
type BookmarkResponseDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	TrackID     string    `json:"trackId"`
	TimestampMs int64     `json:"timestampMs"` // CORRECTED: Use milliseconds
	Note        string    `json:"note,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// MapDomainBookmarkToResponseDTO converts domain bookmark to DTO.
func MapDomainBookmarkToResponseDTO(b *domain.Bookmark) BookmarkResponseDTO {
	return BookmarkResponseDTO{
		ID:          b.ID.String(),
		UserID:      b.UserID.String(),
		TrackID:     b.TrackID.String(),
		TimestampMs: b.Timestamp.Milliseconds(), // CORRECTED: Get milliseconds
		Note:        b.Note,
		CreatedAt:   b.CreatedAt,
	}
}

// --- Paginated Response DTOs (using generic one now) ---
// Keep these structs defined if you intend to use them in Swagger or for type safety,
// even if the actual response uses the generic common_dto.PaginatedResponseDTO.

// PaginatedProgressResponseDTO defines the paginated response for progress list.
type PaginatedProgressResponseDTO struct {
	Data       []PlaybackProgressResponseDTO `json:"data"`
	Total      int                         `json:"total"`
	Limit      int                         `json:"limit"`
	Offset     int                         `json:"offset"`
	Page       int                         `json:"page"`
	TotalPages int                         `json:"totalPages"`
}

// PaginatedBookmarksResponseDTO defines the paginated response for bookmark list.
type PaginatedBookmarksResponseDTO struct {
	Data       []BookmarkResponseDTO `json:"data"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	Page       int                   `json:"page"`
	TotalPages int                   `json:"totalPages"`
}
```

## `internal/adapter/handler/http/dto/.keep`

```

```

## `internal/adapter/handler/http/dto/upload_dto.go`

```go
// internal/adapter/handler/http/dto/upload_dto.go
package dto

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

// CompleteUploadRequestDTO defines the JSON body for finalizing an upload
// and creating the audio track metadata record.
type CompleteUploadRequestDTO struct {
	ObjectKey     string   `json:"objectKey" validate:"required"`
	Title         string   `json:"title" validate:"required,max=255"`
	Description   string   `json:"description"`
	LanguageCode  string   `json:"languageCode" validate:"required"`
	Level         string   `json:"level" validate:"omitempty,oneof=A1 A2 B1 B2 C1 C2 NATIVE"` // Allow empty or valid level
	DurationMs    int64    `json:"durationMs" validate:"required,gt=0"` // Duration in Milliseconds, must be positive
	IsPublic      bool     `json:"isPublic"` // Defaults to false if omitted? Define behavior.
	Tags          []string `json:"tags"`
	CoverImageURL *string  `json:"coverImageUrl" validate:"omitempty,url"`
}
```

## `internal/adapter/handler/http/dto/auth_dto.go`

```go
// internal/adapter/handler/http/dto/auth_dto.go
package dto

// --- Request DTOs ---

// RegisterRequestDTO defines the expected JSON body for user registration.
type RegisterRequestDTO struct {
    Email    string `json:"email" validate:"required,email" example:"user@example.com"` // Add example tag
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


// --- Response DTOs ---

// AuthResponseDTO defines the JSON response body for successful authentication.
type AuthResponseDTO struct {
	Token     string `json:"token"`                // The JWT access token
	IsNewUser *bool  `json:"isNewUser,omitempty"` // Pointer, only included for Google callback if user is new
}

// REMOVED UserResponseDTO from here. It now resides in user_dto.go
```

## `internal/adapter/handler/http/dto/user_dto.go`

```go
// internal/adapter/handler/http/dto/user_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
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

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
)

// --- Request DTOs ---

// ListTracksRequestDTO holds query parameters for listing tracks.
// We'll typically bind these from r.URL.Query() in the handler, not from request body.
// Validation tags aren't standard for query params via validator, often handled manually.
// This struct isn't directly used for request binding but documents parameters.
type ListTracksRequestDTO struct {
	Query         *string   `query:"q"`           // Search query
	LanguageCode  *string   `query:"lang"`      // Filter by language
	Level         *string   `query:"level"`     // Filter by level
	IsPublic      *bool     `query:"isPublic"`  // Filter by public status
	Tags          []string  `query:"tags"`      // Filter by tags (e.g., ?tags=news&tags=podcast)
	SortBy        string    `query:"sortBy"`    // e.g., "createdAt", "title"
	SortDirection string    `query:"sortDir"`   // "asc" or "desc"
	Limit         int       `query:"limit"`     // Pagination limit
	Offset        int       `query:"offset"`    // Pagination offset
}

// CreateCollectionRequestDTO defines the JSON body for creating a collection.
type CreateCollectionRequestDTO struct {
	Title           string   `json:"title" validate:"required,max=255"`
	Description     string   `json:"description"`
	Type            string   `json:"type" validate:"required,oneof=COURSE PLAYLIST"` // Matches domain.CollectionType
	InitialTrackIDs []string `json:"initialTrackIds" validate:"omitempty,dive,uuid"` // Add validation for slice elements
}

// UpdateCollectionRequestDTO defines the JSON body for updating collection metadata.
type UpdateCollectionRequestDTO struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// UpdateCollectionTracksRequestDTO defines the JSON body for updating tracks in a collection.
type UpdateCollectionTracksRequestDTO struct {
	OrderedTrackIDs []string `json:"orderedTrackIds" validate:"omitempty,dive,uuid"` // Add validation
}


// --- Response DTOs ---

// AudioTrackResponseDTO defines the JSON representation of a single audio track.
type AudioTrackResponseDTO struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	LanguageCode  string    `json:"languageCode"`
	Level         string    `json:"level,omitempty"` // Domain type maps to string here
	DurationMs    int64     `json:"durationMs"`   // CORRECTED: Use milliseconds
	CoverImageURL *string   `json:"coverImageUrl,omitempty"`
	UploaderID    *string   `json:"uploaderId,omitempty"` // Use string UUID
	IsPublic      bool      `json:"isPublic"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedAt     time.Time `json:"createdAt"` // Use time.Time, will marshal to RFC3339
	UpdatedAt     time.Time `json:"updatedAt"`
}


// AudioTrackDetailsResponseDTO includes the track metadata and playback URL.
type AudioTrackDetailsResponseDTO struct {
	AudioTrackResponseDTO        // Embed basic track info
	PlayURL               string `json:"playUrl"` // Presigned URL
}

// MapDomainTrackToResponseDTO converts a domain track to its response DTO.
func MapDomainTrackToResponseDTO(track *domain.AudioTrack) AudioTrackResponseDTO {
	var uploaderIDStr *string
	if track.UploaderID != nil {
		s := track.UploaderID.String()
		uploaderIDStr = &s
	}
	return AudioTrackResponseDTO{
		ID:            track.ID.String(),
		Title:         track.Title,
		Description:   track.Description,
		LanguageCode:  track.Language.Code(), // Use Code() method
		Level:         string(track.Level),   // Convert domain level to string
		DurationMs:    track.Duration.Milliseconds(), // CORRECTED: Get milliseconds
		CoverImageURL: track.CoverImageURL,
		UploaderID:    uploaderIDStr,
		IsPublic:      track.IsPublic,
		Tags:          track.Tags,
		CreatedAt:     track.CreatedAt,
		UpdatedAt:     track.UpdatedAt,
	}
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
	Tracks      []AudioTrackResponseDTO `json:"tracks,omitempty"` // Include full track details if needed by frontend
}

// MapDomainCollectionToResponseDTO converts a domain collection to its response DTO.
func MapDomainCollectionToResponseDTO(collection *domain.AudioCollection, tracks []*domain.AudioTrack) AudioCollectionResponseDTO {
	dto := AudioCollectionResponseDTO{
		ID:          collection.ID.String(),
		Title:       collection.Title,
		Description: collection.Description,
		OwnerID:     collection.OwnerID.String(),
		Type:        string(collection.Type),
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
		Tracks:      make([]AudioTrackResponseDTO, 0),
	}
	if tracks != nil {
		dto.Tracks = make([]AudioTrackResponseDTO, len(tracks))
		for i, t := range tracks {
			dto.Tracks[i] = MapDomainTrackToResponseDTO(t)
		}
	}
	return dto
}


// PaginatedTracksResponseDTO defines the paginated response for track list.
// Use the generic one from common_dto.go if preferred
type PaginatedTracksResponseDTO struct {
	Data       []AudioTrackResponseDTO `json:"data"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	Page       int                   `json:"page"`
	TotalPages int                   `json:"totalPages"`
}

// PaginatedCollectionsResponseDTO defines the paginated response for collection list.
// Use the generic one from common_dto.go if preferred
type PaginatedCollectionsResponseDTO struct {
    Data       []AudioCollectionResponseDTO `json:"data"`
    Total      int                        `json:"total"`
    Limit      int                        `json:"limit"`
    Offset     int                        `json:"offset"`
    Page       int                        `json:"page"`
    TotalPages int                        `json:"totalPages"`
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

## `internal/adapter/service/minio/.keep`

```

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

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
)

// MinioStorageService implements the port.FileStorageService interface using MinIO.
type MinioStorageService struct {
	client         *minio.Client
	defaultBucket  string
	defaultExpiry  time.Duration
	logger         *slog.Logger
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
	// Using a short context timeout for the check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Check if the default bucket exists. Create if not? (Consider permissions)
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		log.Error("Failed to check if MinIO bucket exists", "error", err, "bucket", cfg.BucketName)
		// Decide if this is fatal or not. Maybe log warning and proceed?
		// return nil, fmt.Errorf("failed to check minio bucket %s: %w", cfg.BucketName, err)
	}
	if !exists {
		log.Warn("Default MinIO bucket does not exist. Consider creating it.", "bucket", cfg.BucketName)
		// Optionally create the bucket here if desired and permissions allow:
		// err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		// if err != nil { ... handle error ... }
	} else {
		log.Info("MinIO bucket found", "bucket", cfg.BucketName)
	}


	return &MinioStorageService{
		client:         minioClient,
		defaultBucket:  cfg.BucketName,
		defaultExpiry:  cfg.PresignExpiry, // Use expiry from config
		logger:         log,
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

	// Set request parameters (empty for GET)
	reqParams := make(url.Values)
	// Example: Force download with filename
	// reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectKey))

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

	 opts := minio.RemoveObjectOptions{
		 // GovernanceBypass: true, // Set to true to bypass object retention locks if configured
	 }

	 err := s.client.RemoveObject(ctx, bucket, objectKey, opts)
	 if err != nil {
		 s.logger.ErrorContext(ctx, "Failed to delete object from MinIO", "error", err, "bucket", bucket, "key", objectKey)
		 // Check if the error indicates the object wasn't found - might not be a fatal error depending on use case
		 // errResp := minio.ToErrorResponse(err)
		 // if errResp.Code == "NoSuchKey" { return nil } // Example: treat not found as success
		 return fmt.Errorf("failed to delete object %s/%s: %w", bucket, objectKey, err)
	 }

	 s.logger.InfoContext(ctx, "Deleted object from MinIO", "bucket", bucket, "key", objectKey)
	 return nil
}


// GetPresignedPutURL generates a temporary URL for uploading an object.
func (s *MinioStorageService) GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration /*, opts port.PutObjectOptions? */) (string, error) {
    if bucket == "" {
        bucket = s.defaultBucket
    }
    if expiry <= 0 {
        expiry = s.defaultExpiry
    }

    // Removed unused putOpts variable
    // putOpts := minio.PutObjectOptions{}
    // if opts != nil && opts.ContentType != "" {
    //     putOpts.ContentType = opts.ContentType
    // }

    // Note: PresignedPutObject might require url.Values for certain headers like content-type constraints
    // Check minio-go SDK documentation for the exact way to enforce Content-Type if needed.
    // Example (conceptual, check SDK):
     policy := minio.NewPostPolicy()
     policy.SetBucket(bucket)
     policy.SetKey(objectKey)
     policy.SetExpires(time.Now().UTC().Add(expiry))
     // if opts != nil && opts.ContentType != "" {
     //    policy.SetContentType(opts.ContentType)
     // }
     // presignedURL, err := s.client.PresignedPostPolicy(ctx, policy) // For POST uploads
     // OR use PresignedPutObject directly, potentially setting headers via request parameters

    // Simpler version using PresignedPutObject without strict header enforcement in signature itself:
     presignedURL, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
    if err != nil {
        s.logger.ErrorContext(ctx, "Failed to generate presigned PUT URL", "error", err, "bucket", bucket, "key", objectKey)
        return "", fmt.Errorf("failed to get presigned PUT URL for %s/%s: %w", bucket, objectKey, err)
    }

    s.logger.DebugContext(ctx, "Generated presigned PUT URL", "bucket", bucket, "key", objectKey, "expiry", expiry)
    return presignedURL.String(), nil
}

// Compile-time check to ensure MinioStorageService satisfies the port.FileStorageService interface
var _ port.FileStorageService = (*MinioStorageService)(nil)
```

## `internal/adapter/service/google_auth/.keep`

```

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

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
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

	email, _ := payload.Claims["email"].(string) // Email claim
	emailVerified, _ := payload.Claims["email_verified"].(bool) // Email verified claim
	name, _ := payload.Claims["name"].(string) // Name claim
	picture, _ := payload.Claims["picture"].(string) // Picture URL claim

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

## `internal/adapter/repository/postgres/audiotrack_repo_integration_test.go`

```go
// internal/adapter/repository/postgres/audiotrack_repo_integration_test.go
package postgres_test

import (
	"context"
	"testing"
	"time"
    "slices" // For checking list order

	"github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
    "github.com/yvanyang/language-learning-player-backend/internal/port" // For ListParams, Page

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAudioTrackRepo(t *testing.T) *postgres.AudioTrackRepository {
    require.NotNil(t, testDBPool, "Database pool not initialized")
    require.NotNil(t, testLogger, "Logger not initialized")
	return postgres.NewAudioTrackRepository(testDBPool, testLogger)
}

// Helper to clear audio related tables (add collections etc. if needed)
func clearAudioTables(t *testing.T, ctx context.Context) {
    // Delete in reverse order of FK dependencies if needed
    _, err := testDBPool.Exec(ctx, "DELETE FROM collection_tracks")
    require.NoError(t, err, "Failed to clear collection_tracks table")
    _, err = testDBPool.Exec(ctx, "DELETE FROM audio_tracks")
    require.NoError(t, err, "Failed to clear audio_tracks table")
    // clear users if necessary or handle FK on track creation
}

// Helper to create a dummy track for testing
func createTestTrack(t *testing.T, ctx context.Context, repo *postgres.AudioTrackRepository, suffix string) *domain.AudioTrack {
    lang, _ := domain.NewLanguage("en-US", "English")
    track, err := domain.NewAudioTrack(
        "Test Track "+suffix,
        "Description "+suffix,
        "test-bucket",
        fmt.Sprintf("test/object/key_%s.mp3", suffix),
        lang,
        domain.LevelB1,
        120*time.Second,
        nil, // No uploader
        true,
        []string{"test", suffix},
        nil, // No cover image
    )
    require.NoError(t, err)
    err = repo.Create(ctx, track)
    require.NoError(t, err)
    return track
}


func TestAudioTrackRepository_Integration_CreateAndFind(t *testing.T) {
    ctx := context.Background()
    repo := setupAudioTrackRepo(t)
    clearAudioTables(t, ctx)

    track1 := createTestTrack(t, ctx, repo, "1")

    // Find by ID
    found, err := repo.FindByID(ctx, track1.ID)
    require.NoError(t, err)
    require.NotNil(t, found)
    assert.Equal(t, track1.ID, found.ID)
    assert.Equal(t, track1.Title, found.Title)
    assert.Equal(t, track1.MinioObjectKey, found.MinioObjectKey)
    assert.Equal(t, track1.Language.Code(), found.Language.Code())
    assert.Equal(t, track1.Level, found.Level)
    assert.Equal(t, track1.Duration, found.Duration)
    assert.Equal(t, track1.Tags, found.Tags)
    assert.WithinDuration(t, track1.CreatedAt, found.CreatedAt, time.Second)

    // Find non-existent
     _, err = repo.FindByID(ctx, domain.NewTrackID())
     require.Error(t, err)
     assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAudioTrackRepository_Integration_ListByIDs(t *testing.T) {
    ctx := context.Background()
    repo := setupAudioTrackRepo(t)
    clearAudioTables(t, ctx)

    track1 := createTestTrack(t, ctx, repo, "T1")
    track2 := createTestTrack(t, ctx, repo, "T2")
    track3 := createTestTrack(t, ctx, repo, "T3")

    // Test case 1: Get specific tracks in order
    idsToFetch := []domain.TrackID{track3.ID, track1.ID}
    tracks, err := repo.ListByIDs(ctx, idsToFetch)
    require.NoError(t, err)
    require.Len(t, tracks, 2)
    assert.Equal(t, track3.ID, tracks[0].ID) // Check order if repo guarantees it
    assert.Equal(t, track1.ID, tracks[1].ID)

    // Test case 2: Include non-existent ID
    nonExistentID := domain.NewTrackID()
    idsWithNonExistent := []domain.TrackID{track2.ID, nonExistentID}
    tracks, err = repo.ListByIDs(ctx, idsWithNonExistent)
    require.NoError(t, err)
    require.Len(t, tracks, 1) // Should only return existing track
    assert.Equal(t, track2.ID, tracks[0].ID)

    // Test case 3: Empty list
     tracks, err = repo.ListByIDs(ctx, []domain.TrackID{})
     require.NoError(t, err)
     require.Len(t, tracks, 0)
}


func TestAudioTrackRepository_Integration_List_PaginationAndFilter(t *testing.T) {
    ctx := context.Background()
    repo := setupAudioTrackRepo(t)
    clearAudioTables(t, ctx)

    // Create test data
    langEN, _ := domain.NewLanguage("en-US", "")
    langES, _ := domain.NewLanguage("es-ES", "")
	createTrackWithDetails := func(suffix string, lang domain.Language, level domain.AudioLevel, tags []string) {
        track, _ := domain.NewAudioTrack("Title "+suffix, "Desc "+suffix, "bucket", fmt.Sprintf("key_%s", suffix), lang, level, 10*time.Second, nil, true, tags, nil)
        err := repo.Create(ctx, track)
        require.NoError(t, err)
    }
    createTrackWithDetails("EN_A1", langEN, domain.LevelA1, []string{"news", "easy"})
    createTrackWithDetails("EN_A2", langEN, domain.LevelA2, []string{"story", "easy"})
    createTrackWithDetails("EN_B1", langEN, domain.LevelB1, []string{"news", "intermediate"})
    createTrackWithDetails("ES_A2", langES, domain.LevelA2, []string{"story", "easy", "spanish"})
    createTrackWithDetails("ES_B2", langES, domain.LevelB2, []string{"culture", "intermediate", "spanish"})

    // Test Case 1: List all (default sort: createdAt desc), first page
    page1 := port.Page{Limit: 3, Offset: 0}
    tracks, total, err := repo.List(ctx, port.ListTracksParams{}, page1)
    require.NoError(t, err)
    assert.Equal(t, 5, total)
    require.Len(t, tracks, 3)
    // Assert default order (newest first - ES_B2, ES_A2, EN_B1)
    assert.Equal(t, "Title ES_B2", tracks[0].Title)
    assert.Equal(t, "Title ES_A2", tracks[1].Title)
    assert.Equal(t, "Title EN_B1", tracks[2].Title)

    // Test Case 2: Second page
    page2 := port.Page{Limit: 3, Offset: 3}
    tracks, total, err = repo.List(ctx, port.ListTracksParams{}, page2)
    require.NoError(t, err)
    assert.Equal(t, 5, total)
    require.Len(t, tracks, 2)
     // Assert order (EN_A2, EN_A1)
    assert.Equal(t, "Title EN_A2", tracks[0].Title)
    assert.Equal(t, "Title EN_A1", tracks[1].Title)

    // Test Case 3: Filter by language 'es-ES'
    langFilter := "es-ES"
    paramsLang := port.ListTracksParams{LanguageCode: &langFilter}
    tracks, total, err = repo.List(ctx, paramsLang, port.Page{Limit: 10, Offset: 0})
    require.NoError(t, err)
    assert.Equal(t, 2, total)
    require.Len(t, tracks, 2)
    assert.Equal(t, "Title ES_B2", tracks[0].Title) // Default sort still applies
    assert.Equal(t, "Title ES_A2", tracks[1].Title)

    // Test Case 4: Filter by level A2
    levelFilter := domain.LevelA2
    paramsLevel := port.ListTracksParams{Level: &levelFilter}
    tracks, total, err = repo.List(ctx, paramsLevel, port.Page{Limit: 10, Offset: 0})
    require.NoError(t, err)
    assert.Equal(t, 2, total)
    require.Len(t, tracks, 2)
     // Check titles, order might depend on creation time or default sort
    titles := []string{tracks[0].Title, tracks[1].Title}
    assert.Contains(t, titles, "Title EN_A2")
    assert.Contains(t, titles, "Title ES_A2")

    // Test Case 5: Filter by tag 'news' and sort by title asc
    tagFilter := []string{"news"}
    sortByTitle := "title"
    sortDirAsc := "asc"
    paramsTagSort := port.ListTracksParams{Tags: tagFilter, SortBy: sortByTitle, SortDirection: sortDirAsc}
    tracks, total, err = repo.List(ctx, paramsTagSort, port.Page{Limit: 10, Offset: 0})
    require.NoError(t, err)
    assert.Equal(t, 2, total)
    require.Len(t, tracks, 2)
    assert.Equal(t, "Title EN_A1", tracks[0].Title) // Sorted A-Z
    assert.Equal(t, "Title EN_B1", tracks[1].Title)

    // Test Case 6: Filter by query string
    query := "InteRmedIate" // Case-insensitive search in title/desc
    paramsQuery := port.ListTracksParams{Query: &query}
    tracks, total, err = repo.List(ctx, paramsQuery, port.Page{Limit: 10, Offset: 0})
    require.NoError(t, err)
    assert.Equal(t, 2, total) // EN_B1, ES_B2 contain 'intermediate' tag, but query searches title/desc. Let's assume desc has level name.
    // This test might fail if desc doesn't contain the level string - adjust test data or query logic.
    // For now, let's assume the search finds EN_B1 and ES_B2 based on Description content (adjust if needed).
     titles = []string{tracks[0].Title, tracks[1].Title}
     assert.Contains(t, titles, "Title EN_B1")
     assert.Contains(t, titles, "Title ES_B2")

}


// TODO: Add tests for Update (success, conflict on object key)
// TODO: Add tests for Delete
// TODO: Add tests for Exists
```

## `internal/adapter/repository/postgres/.keep`

```

```

## `internal/adapter/repository/postgres/user_repo_integration_test.go`

```go
// internal/adapter/repository/postgres/user_repo_integration_test.go
package postgres_test // Use _test package

import (
	"context"
	"testing"

	"github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres" // Import actual repo package
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a user repo instance for tests
func setupUserRepo(t *testing.T) *postgres.UserRepository {
    require.NotNil(t, testDBPool, "Database pool not initialized")
    require.NotNil(t, testLogger, "Logger not initialized")
    return postgres.NewUserRepository(testDBPool, testLogger)
}

// Helper to clean the users table before/after tests
func clearUsersTable(t *testing.T, ctx context.Context) {
    _, err := testDBPool.Exec(ctx, "DELETE FROM users")
    require.NoError(t, err, "Failed to clear users table")
}

func TestUserRepository_Integration_CreateAndFind(t *testing.T) {
    ctx := context.Background()
    repo := setupUserRepo(t)
    clearUsersTable(t, ctx) // Clean before test

    email := "integration@example.com"
    name := "Integration User"
    hashedPassword := "$2a$12$..." // Use a real (but maybe known) bcrypt hash for testing if needed, or generate one

    // Create user
    user, err := domain.NewLocalUser(email, name, hashedPassword)
    require.NoError(t, err)
    err = repo.Create(ctx, user)
    require.NoError(t, err, "Failed to create user")

    // Find by ID
    foundByID, err := repo.FindByID(ctx, user.ID)
    require.NoError(t, err, "Failed to find user by ID")
    require.NotNil(t, foundByID)
    assert.Equal(t, user.ID, foundByID.ID)
    assert.Equal(t, email, foundByID.Email.String())
    assert.Equal(t, name, foundByID.Name)
    require.NotNil(t, foundByID.HashedPassword)
    assert.Equal(t, hashedPassword, *foundByID.HashedPassword)
    assert.Equal(t, domain.AuthProviderLocal, foundByID.AuthProvider)

    // Find by Email
    emailVO, _ := domain.NewEmail(email)
    foundByEmail, err := repo.FindByEmail(ctx, emailVO)
    require.NoError(t, err, "Failed to find user by Email")
    require.NotNil(t, foundByEmail)
    assert.Equal(t, user.ID, foundByEmail.ID)

    // Find non-existent ID
    _, err = repo.FindByID(ctx, domain.NewUserID()) // Generate random ID
    require.Error(t, err, "Expected error for non-existent ID")
    assert.ErrorIs(t, err, domain.ErrNotFound, "Expected ErrNotFound")
}


func TestUserRepository_Integration_Create_ConflictEmail(t *testing.T) {
    ctx := context.Background()
    repo := setupUserRepo(t)
    clearUsersTable(t, ctx)

    email := "conflict@example.com"
    name1 := "User One"
    name2 := "User Two"
    hash := "somehash"

    // Create first user
    user1, _ := domain.NewLocalUser(email, name1, hash)
    err := repo.Create(ctx, user1)
    require.NoError(t, err)

    // Try creating second user with same email
    user2, _ := domain.NewLocalUser(email, name2, hash)
    err = repo.Create(ctx, user2)
    require.Error(t, err, "Expected error when creating user with duplicate email")
    assert.ErrorIs(t, err, domain.ErrConflict, "Expected ErrConflict for duplicate email")
    assert.Contains(t, err.Error(), "email already exists", "Error message should mention email conflict")
}

func TestUserRepository_Integration_Update(t *testing.T) {
    ctx := context.Background()
    repo := setupUserRepo(t)
    clearUsersTable(t, ctx)

    user, _ := domain.NewLocalUser("update@example.com", "Initial Name", "hash1")
    err := repo.Create(ctx, user)
    require.NoError(t, err)

    // Update fields
    user.Name = "Updated Name"
    newEmailStr := "updated@example.com"
    newEmailVO, _ := domain.NewEmail(newEmailStr)
    user.Email = newEmailVO
    user.GoogleID = nil // Ensure it updates NULL correctly

    err = repo.Update(ctx, user)
    require.NoError(t, err, "Failed to update user")

    // Verify update
    updatedUser, err := repo.FindByID(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, "Updated Name", updatedUser.Name)
    assert.Equal(t, newEmailStr, updatedUser.Email.String())
    assert.Nil(t, updatedUser.GoogleID)
    assert.True(t, updatedUser.UpdatedAt.After(user.CreatedAt), "UpdatedAt should be newer")
}

// TODO: Add tests for FindByProviderID
// TODO: Add tests for Update causing email conflict
// TODO: Add tests for deleting (if applicable)
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

	"github.com/yvanyang/language-learning-player-backend/internal/port" // Adjust import path
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

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
)

type PlaybackProgressRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewPlaybackProgressRepository(db *pgxpool.Pool, logger *slog.Logger) *PlaybackProgressRepository {
	repo := &PlaybackProgressRepository{
		db:     db,
		logger: logger.With("repository", "PlaybackProgressRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper
	}
	return repo
}

// --- Interface Implementation ---

func (r *PlaybackProgressRepository) Upsert(ctx context.Context, progress *domain.PlaybackProgress) error {
	q := r.getQuerier(ctx)
	progress.LastListenedAt = time.Now()
	query := `
        INSERT INTO playback_progress (user_id, track_id, progress_ms, last_listened_at) -- CHANGED column name
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, track_id) DO UPDATE SET
            progress_ms = EXCLUDED.progress_ms, -- CHANGED column name
            last_listened_at = EXCLUDED.last_listened_at
    `
	_, err := q.Exec(ctx, query,
		progress.UserID,
		progress.TrackID,
		progress.Progress.Milliseconds(), // CORRECTED: Save milliseconds
		progress.LastListenedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error upserting playback progress", "error", err, "userID", progress.UserID, "trackID", progress.TrackID)
		// Consider checking for FK violation errors here too
		return fmt.Errorf("upserting playback progress: %w", err)
	}
	r.logger.DebugContext(ctx, "Playback progress upserted", "userID", progress.UserID, "trackID", progress.TrackID, "progressMs", progress.Progress.Milliseconds())
	return nil
}

func (r *PlaybackProgressRepository) Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT user_id, track_id, progress_ms, last_listened_at -- CHANGED column name
        FROM playback_progress
        WHERE user_id = $1 AND track_id = $2
    `
	progress, err := r.scanProgress(ctx, q.QueryRow(ctx, query, userID, trackID)) // Pass QueryRow
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
	selectQuery := `SELECT user_id, track_id, progress_ms, last_listened_at ` + baseQuery // CHANGED column name

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting progress by user: %w", err)
	}
	if total == 0 { return []*domain.PlaybackProgress{}, 0, nil }

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
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning progress in ListByUser", "error", err); continue }
		progressList = append(progressList, progress)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating progress rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating progress rows: %w", err)
	}
	return progressList, total, nil
}

// scanProgress scans a single row into a domain.PlaybackProgress.
func (r *PlaybackProgressRepository) scanProgress(ctx context.Context, row RowScanner) (*domain.PlaybackProgress, error) {
	var p domain.PlaybackProgress
	var progressMs int64 // CORRECTED: Scan milliseconds into int64

	err := row.Scan(
		&p.UserID,
		&p.TrackID,
		&progressMs, // Scan progress_ms
		&p.LastListenedAt,
	)
	if err != nil { return nil, err }

	// CORRECTED: Convert milliseconds back to duration
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
	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
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

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
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
		if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
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
	if total == 0 { return []*domain.AudioCollection{}, 0, nil }

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
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning collection in ListByOwner", "error", err); continue }
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
		if !exists { return domain.ErrNotFound }
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
	if cmdTag.RowsAffected() == 0 { return domain.ErrNotFound }
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
	if err != nil { return nil, err }
	return &collection, nil
}

// Ensure implementation satisfies the interface
var _ port.AudioCollectionRepository = (*AudioCollectionRepository)(nil)
```

## `internal/adapter/repository/postgres/main_test.go`

```go
// internal/adapter/repository/postgres/main_test.go
package postgres_test // Use _test package

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres" // DB driver for migrate
    _ "github.com/golang-migrate/migrate/v4/source/file"      // File source driver for migrate
    "log/slog"
    "io"
)

var testDBPool *pgxpool.Pool
var testLogger *slog.Logger

// TestMain sets up the test database container and runs migrations.
func TestMain(m *testing.M) {
    testLogger = slog.New(slog.NewTextHandler(io.Discard, nil)) // Discard logs during tests
    pool, err := dockertest.NewPool("") // Connect to Docker daemon
    if err != nil {
        log.Fatalf("Could not construct docker pool: %s", err)
    }

    err = pool.Client.Ping() // Check Docker connection
    if err != nil {
        log.Fatalf("Could not connect to Docker: %s", err)
    }

    // Pull image, run container, and wait for it to be ready
    resource, err := pool.RunWithOptions(&dockertest.RunOptions{
        Repository: "postgres",
        Tag:        "15-alpine", // Use a specific version
        Env: []string{
            "POSTGRES_PASSWORD=secret",
            "POSTGRES_USER=testuser",
            "POSTGRES_DB=testdb",
            "listen_addresses = '*'",
        },
    }, func(config *docker.HostConfig) {
        // Set AutoRemove to true so that stopped container is automatically removed
        config.AutoRemove = true
        config.RestartPolicy = docker.RestartPolicy{Name: "no"}
    })
    if err != nil {
        log.Fatalf("Could not start resource: %s", err)
    }

    // Set expiration timer for the container (optional)
    // resource.Expire(120) // Tell docker to kill the container after 120 seconds

    hostAndPort := resource.GetHostPort("5432/tcp")
    databaseUrl := fmt.Sprintf("postgresql://testuser:secret@%s/testdb?sslmode=disable", hostAndPort)

    log.Println("Connecting to database on url: ", databaseUrl)

    // Wait for the database container to be ready
    if err := pool.Retry(func() error {
        var err error
        // Use context with timeout for connection attempt
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        testDBPool, err = pgxpool.New(ctx, databaseUrl)
        if err != nil {
            return err
        }
        return testDBPool.Ping(ctx)
    }); err != nil {
        log.Fatalf("Could not connect to docker database: %s", err)
    }

    log.Println("Database connection successful.")

    // --- Run Migrations ---
    // Construct the migration source URL relative to the test file's location
    // This assumes your migrations folder is ../../../migrations relative to this file
    // Adjust the path accordingly if your structure is different.
    migrationURL := "file://../../../migrations"
    log.Println("Running migrations from:", migrationURL)

    mig, err := migrate.New(migrationURL, databaseUrl)
    if err != nil {
        log.Fatalf("Could not create migrate instance: %s", err)
    }
    if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
         log.Fatalf("Could not run database migrations: %s", err)
    }
    log.Println("Database migrations applied successfully.")
    // ----------------------

    // Run the tests
    code := m.Run()

    // --- Teardown ---
    // Close pool before purging container
    if testDBPool != nil {
        testDBPool.Close()
    }

    // You can't defer this because os.Exit doesn't care for defer
    if err := pool.Purge(resource); err != nil {
        log.Fatalf("Could not purge resource: %s", err)
    }
    log.Println("Database container purged.")

    os.Exit(code)
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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq" // Using lib/pq for array handling with pgx and error codes
	"github.com/google/uuid" // For uuid.NullUUID

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                       // Import pagination
)

type AudioTrackRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewAudioTrackRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioTrackRepository {
	repo := &AudioTrackRepository{
		db:     db,
		logger: logger.With("repository", "AudioTrackRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper defined in tx_manager.go or here
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
		track.Language.Code(),      // Use code from Language VO
		track.Level,                // Use AudioLevel directly (string)
		track.Duration.Milliseconds(), // CORRECTED: Store as milliseconds BIGINT
		track.MinioBucket,
		track.MinioObjectKey,
		track.CoverImageURL,
		track.UploaderID,
		track.IsPublic,
		pq.Array(track.Tags), // Use pq.Array to handle []string -> TEXT[]
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
	track, err := r.scanTrack(ctx, q.QueryRow(ctx, query, id)) // Pass QueryRow directly
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
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
	// Convert domain.TrackID slice to primitive UUID slice for the query parameter
	uuidStrs := make([]string, len(ids))
	for i, id := range ids {
		uuidStrs[i] = id.String()
	}
	// REMOVED unused 'query' variable
	querySimple := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = ANY($1)
    `
	rows, err := q.Query(ctx, querySimple, uuidStrs) // Use simpler query
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
	// Re-order in Go code
	orderedTracks := make([]*domain.AudioTrack, 0, len(ids))
	for _, id := range ids {
		if t, ok := trackMap[id]; ok {
			orderedTracks = append(orderedTracks, t)
		}
	}
	return orderedTracks, nil
}

func (r *AudioTrackRepository) List(ctx context.Context, params port.ListTracksParams, page pagination.Page) ([]*domain.AudioTrack, int, error) {
	q := r.getQuerier(ctx)
	var args []interface{}
	argID := 1
	baseQuery := ` FROM audio_tracks `
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, title, description, language_code, level, duration_ms, minio_bucket, minio_object_key, cover_image_url, uploader_id, is_public, tags, created_at, updated_at ` + baseQuery
	whereClause := " WHERE 1=1"

	if params.Query != nil && *params.Query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID)
		args = append(args, "%"+*params.Query+"%")
		argID++
	}
	if params.LanguageCode != nil && *params.LanguageCode != "" {
		whereClause += fmt.Sprintf(" AND language_code = $%d", argID)
		args = append(args, *params.LanguageCode)
		argID++
	}
	if params.Level != nil && *params.Level != "" {
		whereClause += fmt.Sprintf(" AND level = $%d", argID)
		args = append(args, *params.Level)
		argID++
	}
	if params.IsPublic != nil {
		whereClause += fmt.Sprintf(" AND is_public = $%d", argID)
		args = append(args, *params.IsPublic)
		argID++
	}
	if params.UploaderID != nil {
		whereClause += fmt.Sprintf(" AND uploader_id = $%d", argID)
		args = append(args, *params.UploaderID)
		argID++
	}
	if len(params.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND tags @> $%d", argID) // Check if array contains elements
		args = append(args, pq.Array(params.Tags))
		argID++
	}

	var total int
	err := q.QueryRow(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting audio tracks", "error", err, "filters", params)
		return nil, 0, fmt.Errorf("counting audio tracks: %w", err)
	}
	if total == 0 { return []*domain.AudioTrack{}, 0, nil }

	orderByClause := " ORDER BY created_at DESC" // Default sort
	if params.SortBy != "" {
		allowedSorts := map[string]string{"createdAt": "created_at", "title": "title", "duration": "duration_ms", "level": "level"}
		dbColumn, ok := allowedSorts[params.SortBy]
		if ok {
			direction := " ASC"; if strings.ToLower(params.SortDirection) == "desc" { direction = " DESC" }
			orderByClause = fmt.Sprintf(" ORDER BY %s%s", dbColumn, direction)
		} else { r.logger.WarnContext(ctx, "Invalid sort field requested", "sortBy", params.SortBy) }
	}
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)
	finalQuery := selectQuery + whereClause + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks", "error", err, "filters", params, "page", page)
		return nil, 0, fmt.Errorf("listing audio tracks: %w", err)
	}
	defer rows.Close()
	tracks := make([]*domain.AudioTrack, 0, page.Limit)
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning track in List", "error", err); continue }
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
		track.Language.Code(),      // CORRECTED: Use VO code
		track.Level,                // CORRECTED: Use domain type directly
		track.Duration.Milliseconds(), // CORRECTED: Save milliseconds
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
	if cmdTag.RowsAffected() == 0 { return domain.ErrNotFound }
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
	if cmdTag.RowsAffected() == 0 { return domain.ErrNotFound }
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

// scanTrack scans a single row into a domain.AudioTrack.
// CHANGED: Accepts RowScanner interface
func (r *AudioTrackRepository) scanTrack(ctx context.Context, row RowScanner) (*domain.AudioTrack, error) {
	var track domain.AudioTrack
	var langCode string
	var levelStr string
	var durationMs int64
	var tags pq.StringArray
	var uploaderID uuid.NullUUID

	err := row.Scan(
		&track.ID, &track.Title, &track.Description,
		&langCode,
		&levelStr,
		&durationMs,
		&track.MinioBucket, &track.MinioObjectKey, &track.CoverImageURL,
		&uploaderID,
		&track.IsPublic, &tags, &track.CreatedAt, &track.UpdatedAt,
	)
	if err != nil { return nil, err }

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
	track.Duration = time.Duration(durationMs) * time.Millisecond
	track.Tags = tags

	if uploaderID.Valid {
		uid := domain.UserID(uploaderID.UUID)
		track.UploaderID = &uid
	} else { track.UploaderID = nil }

	return &track, nil
}

var _ port.AudioTrackRepository = (*AudioTrackRepository)(nil)
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
	"github.com/jackc/pgx/v5/pgconn"   // Import pgconn for PgError
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
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

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
)

type BookmarkRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewBookmarkRepository(db *pgxpool.Pool, logger *slog.Logger) *BookmarkRepository {
	repo := &BookmarkRepository{
		db:     db,
		logger: logger.With("repository", "BookmarkRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper
	}
	return repo
}

// --- Interface Implementation ---

func (r *BookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	q := r.getQuerier(ctx)
	query := `
        INSERT INTO bookmarks (id, user_id, track_id, timestamp_ms, note, created_at) -- CHANGED column name
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := q.Exec(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.TrackID,
		bookmark.Timestamp.Milliseconds(), // CORRECTED: Save milliseconds
		bookmark.Note,
		bookmark.CreatedAt,
	)
	if err != nil {
		// Consider checking FK violations here
		r.logger.ErrorContext(ctx, "Error creating bookmark", "error", err, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
		return fmt.Errorf("creating bookmark: %w", err)
	}
	r.logger.InfoContext(ctx, "Bookmark created successfully", "bookmarkID", bookmark.ID, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
	return nil
}

func (r *BookmarkRepository) FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at -- CHANGED column name
        FROM bookmarks
        WHERE id = $1
    `
	bookmark, err := r.scanBookmark(ctx, q.QueryRow(ctx, query, id)) // Pass QueryRow
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding bookmark by ID", "error", err, "bookmarkID", id)
		return nil, fmt.Errorf("finding bookmark by ID: %w", err)
	}
	return bookmark, nil
}

func (r *BookmarkRepository) ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at -- CHANGED column name
        FROM bookmarks
        WHERE user_id = $1 AND track_id = $2
        ORDER BY timestamp_ms ASC -- CHANGED column name
    `
	rows, err := q.Query(ctx, query, userID, trackID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user and track", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("listing bookmarks by user and track: %w", err)
	}
	defer rows.Close()
	bookmarks := make([]*domain.Bookmark, 0)
	for rows.Next() {
		bookmark, err := r.scanBookmark(ctx, rows) // Use RowScanner compatible scan
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUserAndTrack", "error", err); continue }
		bookmarks = append(bookmarks, bookmark)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUserAndTrack", "error", err)
		return nil, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, nil
}

func (r *BookmarkRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error) {
	q := r.getQuerier(ctx)
	baseQuery := `FROM bookmarks WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, user_id, track_id, timestamp_ms, note, created_at ` + baseQuery // CHANGED column name

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting bookmarks by user: %w", err)
	}
	if total == 0 { return []*domain.Bookmark{}, 0, nil }

	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3)
	args := []interface{}{userID, page.Limit, page.Offset}
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("listing bookmarks by user: %w", err)
	}
	defer rows.Close()
	bookmarks := make([]*domain.Bookmark, 0, page.Limit)
	for rows.Next() {
		bookmark, err := r.scanBookmark(ctx, rows) // Use RowScanner compatible scan
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUser", "error", err); continue }
		bookmarks = append(bookmarks, bookmark)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, total, nil
}

func (r *BookmarkRepository) Delete(ctx context.Context, id domain.BookmarkID) error {
	// Ownership check done in Usecase
	q := r.getQuerier(ctx)
	query := `DELETE FROM bookmarks WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting bookmark", "error", err, "bookmarkID", id)
		return fmt.Errorf("deleting bookmark: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	r.logger.InfoContext(ctx, "Bookmark deleted successfully", "bookmarkID", id)
	return nil
}

// scanBookmark scans a single row into a domain.Bookmark.
// Accepts RowScanner interface
func (r *BookmarkRepository) scanBookmark(ctx context.Context, row RowScanner) (*domain.Bookmark, error) {
	var b domain.Bookmark
	var timestampMs int64 // CORRECTED: Scan milliseconds into int64

	err := row.Scan(
		&b.ID,
		&b.UserID,
		&b.TrackID,
		&timestampMs, // **FIXED:** Scan timestamp_ms (Changed '×' to 't')
		&b.Note,
		&b.CreatedAt,
	)
	if err != nil { return nil, err }

	// CORRECTED: Convert milliseconds back to duration
	b.Timestamp = time.Duration(timestampMs) * time.Millisecond

	return &b, nil
}

var _ port.BookmarkRepository = (*BookmarkRepository)(nil)
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

## `internal/domain/.keep`

```

```

## `internal/domain/value_objects_test.go`

```go
// internal/domain/value_objects_test.go
package domain_test // Use _test package to test only exported identifiers

import (
	"testing"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Import the actual package

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedVal domain.Email
		expectError bool
		errorType   error // Optional: Check for specific error type
	}{
		{
			name:        "Valid Email",
			input:       "test@example.com",
			expectedVal: domain.Email{}, // We'll check string representation
			expectError: false,
		},
		{
			name:        "Valid Email with Subdomain",
			input:       "test.sub@example.co.uk",
			expectedVal: domain.Email{},
			expectError: false,
		},
		{
			name:        "Valid Email with Display Name (ignored)",
			input:       "Test User <test@example.com>",
			expectedVal: domain.Email{}, // Should parse out just the address
			expectError: false,
		},
		{
			name:        "Invalid Email - Missing @",
			input:       "testexample.com",
			expectError: true,
			errorType:   domain.ErrInvalidArgument,
		},
		{
			name:        "Invalid Email - Missing Domain",
			input:       "test@",
			expectError: true,
			errorType:   domain.ErrInvalidArgument,
		},
		{
			name:        "Empty Email",
			input:       "",
			expectError: true,
			errorType:   domain.ErrInvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			email, err := domain.NewEmail(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error but got none")
				if tc.errorType != nil {
					assert.ErrorIs(t, err, tc.errorType, "Expected error type mismatch")
				}
			} else {
				require.NoError(t, err, "Expected no error but got one: %v", err)
				// Check the string representation
				expectedEmailStr := tc.input
				if tc.name == "Valid Email with Display Name (ignored)" {
					expectedEmailStr = "test@example.com" // Check the parsed value
				}
				assert.Equal(t, expectedEmailStr, email.String(), "Email string representation mismatch")
			}
		})
	}
}

func TestAudioLevel_IsValid(t *testing.T) {
	assert.True(t, domain.LevelA1.IsValid())
	assert.True(t, domain.LevelB2.IsValid())
	assert.True(t, domain.LevelNative.IsValid())
	assert.True(t, domain.LevelUnknown.IsValid())
	assert.False(t, domain.AudioLevel("D1").IsValid())
	assert.False(t, domain.AudioLevel("").IsValid()) // Should use LevelUnknown
}

func TestCollectionType_IsValid(t *testing.T) {
    assert.True(t, domain.TypeCourse.IsValid())
    assert.True(t, domain.TypePlaylist.IsValid())
    assert.True(t, domain.TypeUnknown.IsValid())
    assert.False(t, domain.CollectionType("BOOK").IsValid())
    assert.False(t, domain.CollectionType("").IsValid())
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

## `internal/domain/user_test.go`

```go
// internal/domain/user_test.go
package domain_test

import (
	"testing"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewLocalUser(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	plainPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	user, err := domain.NewLocalUser(email, name, string(hashedPassword))

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEqual(t, domain.UserID{}, user.ID) // Check ID is generated
	assert.Equal(t, email, user.Email.String())
	assert.Equal(t, name, user.Name)
	require.NotNil(t, user.HashedPassword)
	assert.Equal(t, string(hashedPassword), *user.HashedPassword)
	assert.Nil(t, user.GoogleID)
	assert.Equal(t, domain.AuthProviderLocal, user.AuthProvider)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestNewLocalUser_InvalidEmail(t *testing.T) {
	_, err := domain.NewLocalUser("invalid-email", "Test", "hash")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func TestNewLocalUser_EmptyHash(t *testing.T) {
	_, err := domain.NewLocalUser("test@example.com", "Test", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}


func TestUser_ValidatePassword(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	plainPassword := "password123"
	wrongPassword := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	// Create a local user
	localUser, err := domain.NewLocalUser(email, name, string(hashedPassword))
	require.NoError(t, err)

	// Create a Google user
	googleUser, err := domain.NewGoogleUser(email, name, "google123", nil)
	require.NoError(t, err)


	// Test cases
	match, err := localUser.ValidatePassword(plainPassword)
	assert.NoError(t, err)
	assert.True(t, match, "Correct password should match")

	match, err = localUser.ValidatePassword(wrongPassword)
	assert.NoError(t, err)
	assert.False(t, match, "Incorrect password should not match")

	match, err = localUser.ValidatePassword("")
	assert.NoError(t, err)
	assert.False(t, match, "Empty password should not match")

	// Test on Google user (should fail)
	match, err = googleUser.ValidatePassword(plainPassword)
	assert.Error(t, err, "Should return error for non-local user")
	assert.False(t, match, "Match should be false on error")

    // Test with nil hash (should not happen with NewLocalUser, but test defensively)
    localUser.HashedPassword = nil
    match, err = localUser.ValidatePassword(plainPassword)
	assert.Error(t, err, "Should return error if hash is nil")
	assert.False(t, match, "Match should be false on error")

}

func TestUser_LinkGoogleID(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	plainPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
    googleID := "google12345"
    googleID2 := "google67890"

	localUser, err := domain.NewLocalUser(email, name, string(hashedPassword))
	require.NoError(t, err)

    // Initial state: No Google ID
    assert.Nil(t, localUser.GoogleID)

    // Link valid ID
    err = localUser.LinkGoogleID(googleID)
    assert.NoError(t, err)
    require.NotNil(t, localUser.GoogleID)
    assert.Equal(t, googleID, *localUser.GoogleID)
    // AuthProvider should remain local (as per current implementation)
    assert.Equal(t, domain.AuthProviderLocal, localUser.AuthProvider)

    // Try linking again (should fail)
    err = localUser.LinkGoogleID(googleID2)
    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict)
    assert.Equal(t, googleID, *localUser.GoogleID) // ID should not change

    // Try linking empty ID (should fail)
    localUser.GoogleID = nil // Reset for test
    err = localUser.LinkGoogleID("")
    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidArgument)
    assert.Nil(t, localUser.GoogleID)

}

// TODO: Add tests for other entities like AudioCollection (AddTrack, RemoveTrack, ReorderTracks)
```

## `internal/port/repository.go`

```go
// internal/port/repository.go
package port

import (
	"context"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	// Assuming a common pagination package exists or will be created in pkg/pagination
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import the new package
)

// --- Remove Pagination Placeholder ---
/* Remove this section
// Replace with actual implementation in pkg/pagination later
type Page struct {
	Limit  int
	Offset int
}
*/
// --- End Pagination Placeholder ---


// --- Repository Interfaces ---

// UserRepository defines the persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) // Use Email value object
	FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	// LinkProviderID(ctx context.Context, userID domain.UserID, provider domain.AuthProvider, providerUserID string) error // Maybe Update is sufficient? Depends on implementation. Let's rely on Update for now.
}

// ListTracksParams defines parameters for listing/searching tracks.
type ListTracksParams struct {
	Query         *string           // Search query (title, description, maybe tags)
	LanguageCode  *string           // Filter by language code
	Level         *domain.AudioLevel// Filter by level
	IsPublic      *bool             // Filter by public status
	UploaderID    *domain.UserID    // Filter by uploader
	Tags          []string          // Filter by tags (match any)
	SortBy        string            // e.g., "createdAt", "title", "duration"
	SortDirection string            // "asc" or "desc"
}

// AudioTrackRepository defines the persistence operations for AudioTrack entities.
type AudioTrackRepository interface {
	FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error)
	// ListByIDs retrieves multiple tracks efficiently, preserving order if possible.
	ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error)
	// List retrieves a paginated list of tracks based on filter and sort parameters.
	// Returns the list of tracks for the current page and the total count matching the filters.
	List(ctx context.Context, params ListTracksParams, page pagination.Page) (tracks []*domain.AudioTrack, total int, err error)
	Create(ctx context.Context, track *domain.AudioTrack) error
	Update(ctx context.Context, track *domain.AudioTrack) error
	Delete(ctx context.Context, id domain.TrackID) error
	Exists(ctx context.Context, id domain.TrackID) (bool, error) // Helper to check existence efficiently
}

// AudioCollectionRepository defines the persistence operations for AudioCollection entities.
type AudioCollectionRepository interface {
	FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	// FindWithTracks retrieves collection metadata and its ordered track list.
	// This might require a JOIN or separate queries in implementation.
	FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) (collections []*domain.AudioCollection, total int, err error)
	Create(ctx context.Context, collection *domain.AudioCollection) error // Creates collection metadata only
	UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error // Updates title, description
	// ManageTracks persists the full ordered list of track IDs for a collection.
	// Implementations will likely clear existing associations and insert the new ones.
	ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	Delete(ctx context.Context, id domain.CollectionID) error // Assumes ownership check happens in Usecase
}

// PlaybackProgressRepository defines the persistence operations for PlaybackProgress entities.
type PlaybackProgressRepository interface {
	// Find retrieves progress for a specific user and track. Returns domain.ErrNotFound if none exists.
	Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	// Upsert creates or updates the progress record.
	Upsert(ctx context.Context, progress *domain.PlaybackProgress) error
	// ListByUser retrieves all progress records for a user, paginated, ordered by LastListenedAt descending.
	ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) (progressList []*domain.PlaybackProgress, total int, err error)
}

// BookmarkRepository defines the persistence operations for Bookmark entities.
type BookmarkRepository interface {
	FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error)
	// ListByUserAndTrack retrieves all bookmarks for a specific user on a specific track, ordered by timestamp.
	ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error)
	// ListByUser retrieves all bookmarks for a user, paginated, ordered by CreatedAt descending.
	ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) (bookmarks []*domain.Bookmark, total int, err error)
	Create(ctx context.Context, bookmark *domain.Bookmark) error
	Delete(ctx context.Context, id domain.BookmarkID) error // Assumes ownership check happens in Usecase
}


// --- Transaction Management (Optional but Recommended for Complex Usecases) ---

// Tx represents a database transaction. Specific implementation depends on the driver (e.g., pgx.Tx).
// Using interface{} allows flexibility but requires type assertions in the implementation.
// A more type-safe approach might involve generics (Go 1.18+) or specific interfaces per DB type.
type Tx interface{}

// TransactionManager defines an interface for managing database transactions.
type TransactionManager interface {
	// Begin starts a new transaction and returns a context containing the transaction handle.
	Begin(ctx context.Context) (TxContext context.Context, err error)
	// Commit commits the transaction stored in the context.
	Commit(ctx context.Context) error
	// Rollback aborts the transaction stored in the context.
	Rollback(ctx context.Context) error
	// Execute runs the given function within a transaction.
	// It automatically handles Begin, Commit, and Rollback based on the function's return error.
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error
}

// --- End Transaction Management ---
```

## `internal/port/.keep`

```

```

## `internal/port/usecase.go`

```go
// internal/port/usecase.go
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
)

// AuthUseCase defines the methods for the Auth use case layer.
type AuthUseCase interface {
	RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error)
	LoginWithPassword(ctx context.Context, emailStr, password string) (string, error)
	AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error)
}

// AudioContentUseCase defines the methods for the Audio Content use case layer.
type AudioContentUseCase interface {
	GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error)
	// CHANGED: ListTracks now takes UseCaseListTracksParams
	ListTracks(ctx context.Context, params UseCaseListTracksParams) ([]*domain.AudioTrack, int, pagination.Page, error)
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
	// CHANGED: ListUserProgress now takes ListProgressParams
	ListUserProgress(ctx context.Context, params ListProgressParams) ([]*domain.PlaybackProgress, int, pagination.Page, error)
	CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)
	// CHANGED: ListBookmarks now takes ListBookmarksParams
	ListBookmarks(ctx context.Context, params ListBookmarksParams) ([]*domain.Bookmark, int, pagination.Page, error)
	DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error
}

// UserUseCase defines the interface for user-related operations (e.g., profile)
type UserUseCase interface {
	GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error)
	// Add UpdateUserProfile, etc. here later
}
```

## `internal/port/params.go`

```go
// internal/port/params.go
package port

import (
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
)

// === Use Case Layer Parameter Structs ===

// UseCaseListTracksParams defines parameters for listing/searching tracks at the use case layer.
// It embeds pagination.Page.
// RENAMED from ListTracksParams to avoid conflict with repository params.
type UseCaseListTracksParams struct {
	Query         *string           // Search query (title, description, maybe tags)
	LanguageCode  *string           // Filter by language code
	Level         *domain.AudioLevel// Filter by level
	IsPublic      *bool             // Filter by public status
	UploaderID    *domain.UserID    // Filter by uploader
	Tags          []string          // Filter by tags (match any)
	SortBy        string            // e.g., "createdAt", "title", "duration"
	SortDirection string            // "asc" or "desc"
	Page          pagination.Page   // Embed pagination parameters
}

// ListProgressParams defines parameters for listing user progress at the use case layer.
type ListProgressParams struct {
	UserID domain.UserID
	Page   pagination.Page
}

// ListBookmarksParams defines parameters for listing user bookmarks at the use case layer.
type ListBookmarksParams struct {
	UserID        domain.UserID
	TrackIDFilter *domain.TrackID // Optional filter by track
	Page          pagination.Page
}
```

## `internal/port/service.go`

```go
// internal/port/service.go
package port

import (
	"context"
	"time"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
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

	// TODO: Consider adding methods for checking object existence, getting metadata, etc. if needed.
}

// ExternalUserInfo contains standardized user info retrieved from an external identity provider.
type ExternalUserInfo struct {
	Provider        domain.AuthProvider // e.g., "google"
	ProviderUserID  string              // Unique ID from the provider (e.g., Google subject ID)
	Email           string              // Email address provided by the provider
	IsEmailVerified bool                // Whether the provider claims the email is verified
	Name            string
	PictureURL      *string             // Optional profile picture URL
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

	 // TODO: Add methods for Refresh Token generation/validation if implementing that flow.
}

// REMOVED UserUseCase interface from here
```

## `internal/port/mocks/user_repository_mock.go`

```go
// internal/port/mocks/user_repository_mock.go
package mocks

import (
	"context"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock type for the UserRepository type
type MockUserRepository struct {
	mock.Mock
}

// FindByID provides a mock function with given fields: ctx, id
func (_m *MockUserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	ret := _m.Called(ctx, id)
	var r0 *domain.User
	if rf, ok := ret.Get(0).(func(context.Context, domain.UserID) *domain.User); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.UserID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// FindByEmail provides a mock function with given fields: ctx, email
func (_m *MockUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	ret := _m.Called(ctx, email)
	// ... (similar pattern for return values) ...
	// Simplified:
	var r0 *domain.User
	if ret.Get(0) != nil { r0 = ret.Get(0).(*domain.User)}
	var r1 error = ret.Error(1)
	return r0, r1
}

// FindByProviderID provides a mock function with given fields: ctx, provider, providerUserID
func (_m *MockUserRepository) FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error) {
	ret := _m.Called(ctx, provider, providerUserID)
	// ... (similar pattern) ...
	var r0 *domain.User
	if ret.Get(0) != nil { r0 = ret.Get(0).(*domain.User)}
	var r1 error = ret.Error(1)
	return r0, r1
}

// Create provides a mock function with given fields: ctx, user
func (_m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	ret := _m.Called(ctx, user)
	var r0 error = ret.Error(0)
	return r0
}

// Update provides a mock function with given fields: ctx, user
func (_m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	ret := _m.Called(ctx, user)
	var r0 error = ret.Error(0)
	return r0
}

// Ensure mock satisfies the interface (optional but good practice)
// var _ port.UserRepository = (*MockUserRepository)(nil) // Cannot do this outside the original package
```

## `internal/usecase/auth_uc.go`

```go
// internal/usecase/auth_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
)

type AuthUseCase struct {
	userRepo       port.UserRepository
	secHelper      port.SecurityHelper
	extAuthService port.ExternalAuthService
	jwtExpiry      time.Duration
	logger         *slog.Logger
}

func NewAuthUseCase(
	cfg config.JWTConfig,
	ur port.UserRepository,
	sh port.SecurityHelper,
	eas port.ExternalAuthService,
	log *slog.Logger,
) *AuthUseCase {
	if eas == nil {
		log.Warn("AuthUseCase created without ExternalAuthService implementation.")
	}
	return &AuthUseCase{
		userRepo:       ur,
		secHelper:      sh,
		extAuthService: eas,
		jwtExpiry:      cfg.AccessTokenExpiry,
		logger:         log.With("usecase", "AuthUseCase"),
	}
}

// RegisterWithPassword handles user registration with email and password.
func (uc *AuthUseCase) RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid email provided during registration", "email", emailStr, "error", err)
		return nil, "", fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	existingUser, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err == nil && existingUser != nil {
		uc.logger.WarnContext(ctx, "Registration attempt with existing email", "email", emailStr)
		return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking for existing email", "error", err, "email", emailStr)
		return nil, "", fmt.Errorf("failed to check email existence: %w", err)
	}

	if len(password) < 8 {
		return nil, "", fmt.Errorf("%w: password must be at least 8 characters long", domain.ErrInvalidArgument)
	}

	hashedPassword, err := uc.secHelper.HashPassword(ctx, password)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to hash password during registration", "error", err)
		return nil, "", fmt.Errorf("failed to process password: %w", err)
	}

	user, err := domain.NewLocalUser(emailStr, name, hashedPassword)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create domain user object", "error", err)
		return nil, "", fmt.Errorf("failed to create user data: %w", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new user to repository", "error", err, "userID", user.ID)
		if errors.Is(err, domain.ErrConflict) {
			return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
		}
		return nil, "", fmt.Errorf("failed to register user: %w", err)
	}

	token, err := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT after registration", "error", err, "userID", user.ID)
		return nil, "", fmt.Errorf("failed to finalize registration: %w", err)
	}

	uc.logger.InfoContext(ctx, "User registered successfully via password", "userID", user.ID, "email", emailStr)
	return user, token, nil
}

// LoginWithPassword handles user login with email and password.
func (uc *AuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (string, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		return "", domain.ErrAuthenticationFailed
	}
	user, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Login attempt for non-existent email", "email", emailStr)
			return "", domain.ErrAuthenticationFailed
		}
		uc.logger.ErrorContext(ctx, "Error finding user by email during login", "error", err, "email", emailStr)
		return "", fmt.Errorf("failed during login process: %w", err)
	}
	if user.AuthProvider != domain.AuthProviderLocal || user.HashedPassword == nil {
		uc.logger.WarnContext(ctx, "Login attempt for user with non-local provider or no password", "email", emailStr, "userID", user.ID, "provider", user.AuthProvider)
		return "", domain.ErrAuthenticationFailed
	}
	if !uc.secHelper.CheckPasswordHash(ctx, password, *user.HashedPassword) {
		uc.logger.WarnContext(ctx, "Incorrect password provided for user", "email", emailStr, "userID", user.ID)
		return "", domain.ErrAuthenticationFailed
	}
	token, err := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT during login", "error", err, "userID", user.ID)
		return "", fmt.Errorf("failed to finalize login: %w", err)
	}
	uc.logger.InfoContext(ctx, "User logged in successfully via password", "userID", user.ID)
	return token, nil
}

// AuthenticateWithGoogle handles login or registration via Google ID Token.
// Implements Strategy C: Conflict if email exists with a different provider/link.
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error) {
	if uc.extAuthService == nil {
		uc.logger.ErrorContext(ctx, "ExternalAuthService not configured for Google authentication")
		return "", false, fmt.Errorf("google authentication is not enabled")
	}

	extInfo, err := uc.extAuthService.VerifyGoogleToken(ctx, googleIdToken)
	if err != nil { return "", false, err }

	user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
	if err == nil {
		// Case 1: User found by Google ID -> Login success
		uc.logger.InfoContext(ctx, "User authenticated via existing Google ID", "userID", user.ID, "googleID", extInfo.ProviderUserID)
		token, tokenErr := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
		if tokenErr != nil { return "", false, fmt.Errorf("failed to finalize login: %w", tokenErr) }
		return token, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error finding user by provider ID", "error", err, "provider", extInfo.Provider, "providerUserID", extInfo.ProviderUserID)
		return "", false, fmt.Errorf("database error during authentication: %w", err)
	}

	// User not found by Google ID, check email if available
	if extInfo.Email != "" {
		emailVO, emailErr := domain.NewEmail(extInfo.Email)
		if emailErr != nil {
			uc.logger.WarnContext(ctx, "Invalid email format received from Google token", "error", emailErr, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
			return "", false, fmt.Errorf("%w: invalid email format from provider", domain.ErrAuthenticationFailed)
		}

		userByEmail, errEmail := uc.userRepo.FindByEmail(ctx, emailVO)
		if errEmail == nil {
			// Case 2: User found by email -> Conflict (Strategy C)
			uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but Google ID did not match or was not linked", "email", extInfo.Email, "existingUserID", userByEmail.ID, "existingProvider", userByEmail.AuthProvider)
			return "", false, fmt.Errorf("%w: email is already associated with a different account", domain.ErrConflict)
		} else if !errors.Is(errEmail, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Error finding user by email", "error", errEmail, "email", extInfo.Email)
			return "", false, fmt.Errorf("database error during authentication: %w", errEmail)
		}
		// If errEmail is ErrNotFound, fall through to user creation.
	} else {
		uc.logger.InfoContext(ctx, "Google token verified, but no email provided. Proceeding to create new user based on Google ID only.", "googleID", extInfo.ProviderUserID)
	}

	// Case 3: User not found by Google ID or conflicting Email -> Create new user
	uc.logger.InfoContext(ctx, "No existing user found by Google ID or conflicting email. Creating new user.", "googleID", extInfo.ProviderUserID, "email", extInfo.Email)

	newUser, err := domain.NewGoogleUser(
		extInfo.Email, extInfo.Name, extInfo.ProviderUserID, extInfo.PictureURL,
	)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create new Google user domain object", "error", err, "extInfo", extInfo)
		return "", false, fmt.Errorf("failed to process user data from Google: %w", err)
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new Google user to repository", "error", err, "googleID", newUser.GoogleID, "email", newUser.Email.String())
		if errors.Is(err, domain.ErrConflict) {
             return "", false, fmt.Errorf("%w: account conflict during creation", domain.ErrConflict)
        }
		return "", false, fmt.Errorf("failed to create new user account: %w", err)
	}

	uc.logger.InfoContext(ctx, "New user created via Google authentication", "userID", newUser.ID, "email", newUser.Email.String())
	token, tokenErr := uc.secHelper.GenerateJWT(ctx, newUser.ID, uc.jwtExpiry)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT for newly created Google user", "error", tokenErr, "userID", newUser.ID)
		return "", true, fmt.Errorf("failed to finalize registration: %w", tokenErr)
	}
	return token, true, nil
}

// Compile-time check to ensure AuthUseCase satisfies the port.AuthUseCase interface
var _ port.AuthUseCase = (*AuthUseCase)(nil)
```

## `internal/usecase/.keep`

```

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
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path (for presign expiry)
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // GetUserID
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                      // Import pagination
)

// AudioContentUseCase handles business logic related to audio tracks and collections.
type AudioContentUseCase struct {
	trackRepo      port.AudioTrackRepository
	collectionRepo port.AudioCollectionRepository
	storageService port.FileStorageService
	txManager      port.TransactionManager
	presignExpiry  time.Duration
	logger         *slog.Logger
}

// NewAudioContentUseCase creates a new AudioContentUseCase.
func NewAudioContentUseCase(
	minioCfg config.MinioConfig,
	tr port.AudioTrackRepository,
	cr port.AudioCollectionRepository,
	ss port.FileStorageService,
	tm port.TransactionManager,
	log *slog.Logger,
) *AudioContentUseCase {
	if tm == nil {
		log.Warn("AudioContentUseCase created without TransactionManager implementation. Transactional operations will fail.")
	}
	return &AudioContentUseCase{
		trackRepo:      tr,
		collectionRepo: cr,
		storageService: ss,
		txManager:      tm,
		presignExpiry:  minioCfg.PresignExpiry,
		logger:         log.With("usecase", "AudioContentUseCase"),
	}
}

// --- Track Use Cases ---

// GetAudioTrackDetails retrieves details for a single audio track, including a presigned URL.
func (uc *AudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error) {
	track, err := uc.trackRepo.FindByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Audio track not found", "trackID", trackID)
		} else {
			uc.logger.ErrorContext(ctx, "Failed to get audio track from repository", "error", err, "trackID", trackID)
		}
		return nil, "", err
	}

	playURL, err := uc.storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, uc.presignExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate presigned URL for track", "error", err, "trackID", trackID)
		return nil, "", fmt.Errorf("could not retrieve playback URL: %w", err)
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved audio track details", "trackID", trackID)
	return track, playURL, nil
}

// ListTracks retrieves a paginated list of tracks based on filtering and sorting criteria.
// CHANGED: Signature now accepts params port.UseCaseListTracksParams
func (uc *AudioContentUseCase) ListTracks(ctx context.Context, params port.UseCaseListTracksParams) ([]*domain.AudioTrack, int, pagination.Page, error) {
	// Apply defaults/constraints to pagination within the params struct
	pageParams := pagination.NewPageFromOffset(params.Page.Limit, params.Page.Offset)
	// Update the params struct with the constrained page info, so repo gets the correct values
	params.Page = pageParams

	// Convert UseCaseListTracksParams to the format expected by the repository List method
	// (This assumes the repository's ListTracksParams is the same or a subset)
	// If repository needs different params, perform mapping here.
	// For now, assume they are compatible enough (or repo uses the same struct name).
	repoParams := port.ListTracksParams{ // This refers to the repo's ListTracksParams struct in port/repository.go
		Query:         params.Query,
		LanguageCode:  params.LanguageCode,
		Level:         params.Level,
		IsPublic:      params.IsPublic,
		UploaderID:    params.UploaderID,
		Tags:          params.Tags,
		SortBy:        params.SortBy,
		SortDirection: params.SortDirection,
	}


	tracks, total, err := uc.trackRepo.List(ctx, repoParams, pageParams) // Pass repoParams and pageParams
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list audio tracks from repository", "error", err, "params", params, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve track list: %w", err)
	}
	uc.logger.InfoContext(ctx, "Successfully listed audio tracks", "count", len(tracks), "total", total, "params", params, "page", pageParams)
	// Return the actual page params used (after applying defaults/constraints)
	return tracks, total, pageParams, nil
}


// --- Collection Use Cases --- (Rest of the file remains the same)

// CreateCollection creates a new audio collection, potentially adding initial tracks atomically.
func (uc *AudioContentUseCase) CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return nil, domain.ErrUnauthenticated }

	if uc.txManager == nil {
		uc.logger.ErrorContext(ctx, "TransactionManager is nil, cannot create collection atomically")
		return nil, fmt.Errorf("internal configuration error: transaction manager not available")
	}


	collection, err := domain.NewAudioCollection(title, description, userID, colType)
	if err != nil {
		uc.logger.WarnContext(ctx, "Failed to create collection domain object", "error", err, "title", title, "type", colType, "userID", userID)
		return nil, err
	}

	// Execute repository operations within a transaction
	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		// 1. Save collection metadata within the transaction
		if err := uc.collectionRepo.Create(txCtx, collection); err != nil {
			return fmt.Errorf("saving collection metadata: %w", err)
		}

		// 2. Add initial tracks if provided, also within the transaction
		if len(initialTrackIDs) > 0 {
			// Optional: Validate track IDs exist before attempting insertion
			exists, validateErr := uc.validateTrackIDsExist(txCtx, initialTrackIDs)
			if validateErr != nil {
				// Logged inside validateTrackIDsExist
				return fmt.Errorf("validating initial tracks: %w", validateErr) // Propagate validation error
			}
			if !exists {
				// This condition implies specific IDs were missing, error logged inside validate func
				// The specific error message should come from validateTrackIDsExist ideally
				return fmt.Errorf("%w: one or more initial track IDs do not exist", domain.ErrInvalidArgument)
			}

			// Call ManageTracks repo method within the transaction context
			// Note: The repo method might implicitly use txCtx via getQuerier
			if err := uc.collectionRepo.ManageTracks(txCtx, collection.ID, initialTrackIDs); err != nil {
				return fmt.Errorf("adding initial tracks: %w", err)
			}
			// Update in-memory object only on success within transaction scope
			collection.TrackIDs = initialTrackIDs
		}
		return nil // Commit transaction
	})

	if finalErr != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed during collection creation", "error", finalErr, "collectionID", collection.ID, "userID", userID)
		return nil, fmt.Errorf("failed to create collection: %w", finalErr)
	}

	uc.logger.InfoContext(ctx, "Audio collection created", "collectionID", collection.ID, "userID", userID)
	return collection, nil
}

// GetCollectionDetails retrieves details for a single collection, including its ordered track list.
func (uc *AudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx)

	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) { uc.logger.WarnContext(ctx, "Audio collection not found", "collectionID", collectionID) } else { uc.logger.ErrorContext(ctx, "Failed to get audio collection from repository", "error", err, "collectionID", collectionID) }
		return nil, err
	}

	// Authorization Check: Only owner can view (example rule)
	if !userAuthenticated || collection.OwnerID != userID {
		uc.logger.WarnContext(ctx, "Permission denied for user on collection", "collectionID", collectionID, "ownerID", collection.OwnerID, "requestingUserID", userID)
		return nil, domain.ErrPermissionDenied
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved collection details", "collectionID", collectionID, "trackCount", len(collection.TrackIDs))
	return collection, nil
}

// GetCollectionTracks retrieves the ordered list of tracks for a specific collection.
func (uc *AudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	// NOTE: Assumes authorization check happened before (e.g., in GetCollectionDetails or middleware)
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
    if err != nil { return nil, err } // Handles NotFound

    if len(collection.TrackIDs) == 0 { return []*domain.AudioTrack{}, nil }

    tracks, err := uc.trackRepo.ListByIDs(ctx, collection.TrackIDs)
    if err != nil {
        uc.logger.ErrorContext(ctx, "Failed to list tracks by IDs for collection", "error", err, "collectionID", collectionID)
        return nil, fmt.Errorf("failed to retrieve track details for collection: %w", err)
    }
    uc.logger.InfoContext(ctx, "Successfully retrieved tracks for collection", "collectionID", collectionID, "trackCount", len(tracks))
    return tracks, nil
}


// UpdateCollectionMetadata updates the title and description of a collection owned by the user.
func (uc *AudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	if title == "" { return fmt.Errorf("%w: collection title cannot be empty", domain.ErrInvalidArgument) }

	// Repo update includes owner check in WHERE clause
	tempCollection := &domain.AudioCollection{ ID: collectionID, OwnerID: userID, Title: title, Description: description }
	err := uc.collectionRepo.UpdateMetadata(ctx, tempCollection)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrPermissionDenied) {
			uc.logger.ErrorContext(ctx, "Failed to update collection metadata", "error", err, "collectionID", collectionID, "userID", userID)
		}
		return err
	}
	uc.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collectionID, "userID", userID)
	return nil
}

// UpdateCollectionTracks updates the list and order of tracks atomically using TransactionManager.
func (uc *AudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	if uc.txManager == nil {
		uc.logger.ErrorContext(ctx, "TransactionManager is nil, cannot update collection tracks atomically")
		return fmt.Errorf("internal configuration error: transaction manager not available")
	}


	// Wrap the logic in a transaction
	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		// 1. Verify collection exists and user owns it (within Tx)
		collection, err := uc.collectionRepo.FindByID(txCtx, collectionID) // Use txCtx
		if err != nil { return err } // Propagate error (NotFound or DB error)
		if collection.OwnerID != userID {
			uc.logger.WarnContext(txCtx, "Attempt to modify tracks of collection not owned by user", "collectionID", collectionID, "ownerID", collection.OwnerID, "userID", userID)
			return domain.ErrPermissionDenied
		}

		// 2. Validate that all provided track IDs exist (within Tx)
		if len(orderedTrackIDs) > 0 {
			exists, validateErr := uc.validateTrackIDsExist(txCtx, orderedTrackIDs) // Use txCtx
			if validateErr != nil { return fmt.Errorf("validating tracks: %w", validateErr) }
			if !exists { return fmt.Errorf("%w: one or more track IDs do not exist", domain.ErrInvalidArgument) }
		}

		// 3. Call repository method to manage tracks (passes txCtx implicitly via getQuerier)
		if err := uc.collectionRepo.ManageTracks(txCtx, collectionID, orderedTrackIDs); err != nil {
			return fmt.Errorf("updating collection tracks in repository: %w", err)
		}
		return nil // Commit transaction
	})

	if finalErr != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed during collection track update", "error", finalErr, "collectionID", collectionID, "userID", userID)
		return fmt.Errorf("failed to update collection tracks: %w", finalErr)
	}

	uc.logger.InfoContext(ctx, "Collection tracks updated", "collectionID", collectionID, "userID", userID, "trackCount", len(orderedTrackIDs))
	return nil
}


// DeleteCollection deletes a collection owned by the user.
func (uc *AudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	// 1. Verify ownership first (read operation, likely no Tx needed)
	collection, err := uc.collectionRepo.FindByID(ctx, collectionID)
	if err != nil { return err } // Handles ErrNotFound
	if collection.OwnerID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete collection not owned by user", "collectionID", collectionID, "ownerID", collection.OwnerID, "userID", userID)
		return domain.ErrPermissionDenied
	}

	// 2. Delete from repository (simple delete, likely no Tx needed unless linked hooks exist)
	err = uc.collectionRepo.Delete(ctx, collectionID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to delete collection from repository", "error", err, "collectionID", collectionID, "userID", userID) }
		return err
	}
	uc.logger.InfoContext(ctx, "Collection deleted", "collectionID", collectionID, "userID", userID)
	return nil
}


// Helper function to validate track IDs exist
func (uc *AudioContentUseCase) validateTrackIDsExist(ctx context.Context, trackIDs []domain.TrackID) (bool, error) {
	if len(trackIDs) == 0 {
		return true, nil // Nothing to validate
	}
	existingTracks, err := uc.trackRepo.ListByIDs(ctx, trackIDs)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to validate track IDs existence", "error", err)
		return false, fmt.Errorf("failed to verify tracks: %w", err)
	}
	if len(existingTracks) != len(trackIDs) {
		// Find which IDs are missing (optional logging detail)
		providedSet := make(map[domain.TrackID]struct{}, len(trackIDs)); for _, id := range trackIDs { providedSet[id] = struct{}{} }; for _, track := range existingTracks { delete(providedSet, track.ID) }; missingIDs := make([]string, 0, len(providedSet)); for id := range providedSet { missingIDs = append(missingIDs, id.String()) }
		uc.logger.WarnContext(ctx, "Attempt to use non-existent track IDs", "missingTrackIDs", missingIDs)
		return false, nil // Return false, no error (usecase handles returning ErrInvalidArgument)
	}
	return true, nil
}

// Compile-time check to ensure AudioContentUseCase satisfies the port.AudioContentUseCase interface
var _ port.AudioContentUseCase = (*AudioContentUseCase)(nil)
```

## `internal/usecase/user_activity_uc.go`

```go
// internal/usecase/user_activity_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
)

// UserActivityUseCase handles business logic for user interactions like progress and bookmarks.
type UserActivityUseCase struct {
	progressRepo port.PlaybackProgressRepository
	bookmarkRepo port.BookmarkRepository
	trackRepo    port.AudioTrackRepository // Needed to validate track existence
	logger       *slog.Logger
}

// NewUserActivityUseCase creates a new UserActivityUseCase.
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

// RecordPlaybackProgress saves or updates the user's listening progress for a track.
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

// GetPlaybackProgress retrieves the user's progress for a specific track.
func (uc *UserActivityUseCase) GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
    progress, err := uc.progressRepo.Find(ctx, userID, trackID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to get playback progress", "error", err, "userID", userID, "trackID", trackID) }
		return nil, err
	}
	return progress, nil
}

// ListUserProgress retrieves a paginated list of all progress records for a user.
// CHANGED: Signature now accepts params port.ListProgressParams
func (uc *UserActivityUseCase) ListUserProgress(ctx context.Context, params port.ListProgressParams) ([]*domain.PlaybackProgress, int, pagination.Page, error) {
	// Apply defaults/constraints to pagination within the params struct
	pageParams := pagination.NewPageFromOffset(params.Page.Limit, params.Page.Offset)

	progressList, total, err := uc.progressRepo.ListByUser(ctx, params.UserID, pageParams)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list user progress", "error", err, "userID", params.UserID, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve progress list: %w", err)
	}
	// Return the actual page params used (after applying defaults/constraints)
	return progressList, total, pageParams, nil
}

// --- Bookmark Use Cases ---

// CreateBookmark creates a new bookmark for the user on a specific track.
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
// CHANGED: Signature now accepts params port.ListBookmarksParams
func (uc *UserActivityUseCase) ListBookmarks(ctx context.Context, params port.ListBookmarksParams) ([]*domain.Bookmark, int, pagination.Page, error) {
	var bookmarks []*domain.Bookmark
	var total int
	var err error
	// Apply defaults/constraints to pagination within the params struct
	pageParams := pagination.NewPageFromOffset(params.Page.Limit, params.Page.Offset)


	if params.TrackIDFilter != nil {
		// Listing for a specific track: Fetch ALL, NO standard pagination applied here
		// The repo method ListByUserAndTrack fetches all for that track.
		bookmarks, err = uc.bookmarkRepo.ListByUserAndTrack(ctx, params.UserID, *params.TrackIDFilter)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user and track", "error", err, "userID", params.UserID, "trackID", *params.TrackIDFilter)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks for track: %w", err)
		}
		total = len(bookmarks)
		// Adjust pageParams to reflect all results were returned for this filter
		pageParams = pagination.Page{Limit: total, Offset: 0}
		if total == 0 { pageParams.Limit = pagination.DefaultLimit }

	} else {
		// Listing all bookmarks for the user (PAGINATED)
		bookmarks, total, err = uc.bookmarkRepo.ListByUser(ctx, params.UserID, pageParams)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user", "error", err, "userID", params.UserID, "page", pageParams)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks: %w", err)
		}
		// pageParams already holds the constrained page info used for the query
	}

	return bookmarks, total, pageParams, nil
}

// DeleteBookmark deletes a bookmark owned by the user.
func (uc *UserActivityUseCase) DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error {
	bookmark, err := uc.bookmarkRepo.FindByID(ctx, bookmarkID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to find bookmark for deletion", "error", err, "bookmarkID", bookmarkID, "userID", userID) }
		return err
	}
	if bookmark.UserID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete bookmark not owned by user", "bookmarkID", bookmarkID, "ownerID", bookmark.UserID, "userID", userID)
		return domain.ErrPermissionDenied
	}
	// Use bookmark ID directly for deletion, ownership already checked
	if err := uc.bookmarkRepo.Delete(ctx, bookmarkID); err != nil {
		// Handles ErrNotFound race condition gracefully
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to delete bookmark from repository", "error", err, "bookmarkID", bookmarkID, "userID", userID) }
		// Return the error from the Delete operation
		return err
	}
	uc.logger.InfoContext(ctx, "Bookmark deleted", "bookmarkID", bookmarkID, "userID", userID)
	return nil
}

// Compile-time check to ensure UserActivityUseCase satisfies the port.UserActivityUseCase interface
var _ port.UserActivityUseCase = (*UserActivityUseCase)(nil)
```

## `internal/usecase/user_uc.go`

```go
// internal/usecase/user_uc.go
package usecase

import (
	"context"
	"log/slog"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
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

## `internal/usecase/user_activity_uc_test.go`

```go
// internal/usecase/user_activity_uc_test.go
package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks"
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
    "io" // For discarding logs
	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use test helpers if defined in another _test file in the same package,
// otherwise, redefine or move to a shared test utility package.
func newTestLogger() *slog.Logger {
     return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestUserActivityUseCase_RecordPlaybackProgress_Success(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t) // Needed for constructor
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLogger()

	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progress := 30 * time.Second

	// Expect TrackRepo.Exists to be called and return true
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(true, nil).Once()

	// Expect ProgressRepo.Upsert to be called
	mockProgressRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(p *domain.PlaybackProgress) bool {
		return p.UserID == userID && p.TrackID == trackID && p.Progress == progress
	})).Return(nil).Once()

	// Execute
	err := uc.RecordPlaybackProgress(context.Background(), userID, trackID, progress)

	// Assert
	require.NoError(t, err)
	mockTrackRepo.AssertExpectations(t)
	mockProgressRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_RecordPlaybackProgress_TrackNotFound(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLogger()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progress := 30 * time.Second

	// Expect TrackRepo.Exists to return false
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(false, nil).Once()

	// Execute
	err := uc.RecordPlaybackProgress(context.Background(), userID, trackID, progress)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound) // Expect NotFound because track doesn't exist

	mockTrackRepo.AssertExpectations(t)
	// Ensure Upsert was not called
	mockProgressRepo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
}

func TestUserActivityUseCase_CreateBookmark_Success(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLogger()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	timestamp := 60 * time.Second
	note := "Important point"

	// Expect TrackRepo.Exists to return true
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(true, nil).Once()

	// Expect BookmarkRepo.Create to be called
	// We capture the bookmark to check its details
	var createdBookmark *domain.Bookmark
	mockBookmarkRepo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Bookmark) bool {
		createdBookmark = b
		return b.UserID == userID && b.TrackID == trackID && b.Timestamp == timestamp && b.Note == note
	})).Return(nil).Once()


	// Execute
	bookmark, err := uc.CreateBookmark(context.Background(), userID, trackID, timestamp, note)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, bookmark)
	// Check if the returned bookmark matches the one captured (or has expected fields)
	assert.Equal(t, createdBookmark.ID, bookmark.ID) // ID is generated by domain/repo
	assert.Equal(t, userID, bookmark.UserID)
	assert.Equal(t, trackID, bookmark.TrackID)
	assert.Equal(t, timestamp, bookmark.Timestamp)
	assert.Equal(t, note, bookmark.Note)

	mockTrackRepo.AssertExpectations(t)
	mockBookmarkRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_DeleteBookmark_SuccessOwned(t *testing.T) {
    mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
    mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    logger := newTestLogger()
    uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

    userID := domain.NewUserID()    // The user performing the action
    bookmarkID := domain.NewBookmarkID()
    trackID := domain.NewTrackID()

    // The bookmark to be deleted, owned by the current user
    ownedBookmark := &domain.Bookmark{
        ID:        bookmarkID,
        UserID:    userID, // Belongs to the user
        TrackID:   trackID,
        Timestamp: 10 * time.Second,
    }

    // Expect FindByID to find the owned bookmark
    mockBookmarkRepo.On("FindByID", mock.Anything, bookmarkID).Return(ownedBookmark, nil).Once()
    // Expect Delete to be called with the correct ID
    mockBookmarkRepo.On("Delete", mock.Anything, bookmarkID).Return(nil).Once()

    // Execute
    err := uc.DeleteBookmark(context.Background(), userID, bookmarkID)

    // Assert
    require.NoError(t, err)
    mockBookmarkRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_DeleteBookmark_PermissionDenied(t *testing.T) {
    mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
    mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    logger := newTestLogger()
    uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

    requestingUserID := domain.NewUserID() // The user trying to delete
    ownerUserID := domain.NewUserID()      // The actual owner (different)
    bookmarkID := domain.NewBookmarkID()
    trackID := domain.NewTrackID()

    // Bookmark owned by someone else
    otherUsersBookmark := &domain.Bookmark{
        ID:        bookmarkID,
        UserID:    ownerUserID, // Belongs to a different user
        TrackID:   trackID,
        Timestamp: 10 * time.Second,
    }

    // Expect FindByID to find the bookmark
    mockBookmarkRepo.On("FindByID", mock.Anything, bookmarkID).Return(otherUsersBookmark, nil).Once()

    // Execute
    err := uc.DeleteBookmark(context.Background(), requestingUserID, bookmarkID)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrPermissionDenied) // Expect permission denied

    mockBookmarkRepo.AssertExpectations(t) // FindByID was called
    // Ensure Delete was NOT called
    mockBookmarkRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// TODO: Add tests for ListBookmarks (with and without track filter)
// TODO: Add tests for GetPlaybackProgress, ListUserProgress
```

## `internal/usecase/upload_uc.go`

```go
// internal/usecase/upload_uc.go
package usecase

import (
	"context"
	"errors" // Import errors package
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
)

// UploadUseCase handles the business logic for file uploads.
type UploadUseCase struct {
	trackRepo      port.AudioTrackRepository
	storageService port.FileStorageService
	logger         *slog.Logger
	minioBucket    string // Store default bucket name
}

// NewUploadUseCase creates a new UploadUseCase.
func NewUploadUseCase(
	cfg config.MinioConfig, // Need minio config for bucket name
	tr port.AudioTrackRepository,
	ss port.FileStorageService,
	log *slog.Logger,
) *UploadUseCase {
	return &UploadUseCase{
		trackRepo:      tr,
		storageService: ss,
		logger:         log.With("usecase", "UploadUseCase"),
		minioBucket:    cfg.BucketName,
	}
}

// RequestUploadResult holds the result of requesting an upload URL.
type RequestUploadResult struct {
	UploadURL string `json:"uploadUrl"`
	ObjectKey string `json:"objectKey"`
}

// RequestUpload generates a presigned PUT URL for the client to upload a file.
func (uc *UploadUseCase) RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*RequestUploadResult, error) {
	log := uc.logger.With("userID", userID.String(), "filename", filename, "contentType", contentType)

	// Basic validation (can add more, e.g., allowed content types, size limits if known)
	if filename == "" {
		return nil, fmt.Errorf("%w: filename cannot be empty", domain.ErrInvalidArgument)
	}
	// Optional: Validate contentType against allowed types
	allowedTypes := []string{"audio/mpeg", "audio/ogg", "audio/wav", "audio/aac", "audio/mp4"} // Example allowed types
	isAllowedType := false
	for _, t := range allowedTypes {
		if contentType == t {
			isAllowedType = true
			break
		}
	}
	if !isAllowedType {
		log.Warn("Upload requested with disallowed content type")
		// return nil, fmt.Errorf("%w: content type '%s' is not allowed", domain.ErrInvalidArgument, contentType)
		// Or allow it, but log warning. Let's allow for now.
	}


	// Generate a unique object key
	extension := filepath.Ext(filename)                    // Get extension (e.g., ".mp3")
	randomUUID := uuid.NewString()
	objectKey := fmt.Sprintf("user-uploads/%s/%s%s", userID.String(), randomUUID, extension) // Structure: prefix/userID/uuid.ext

	log = log.With("objectKey", objectKey) // Add objectKey to logger context

	// Get the presigned PUT URL from the storage service
	// Use a reasonable expiry time for the upload URL
	uploadURLExpiry := 15 * time.Minute // Configurable?
	uploadURL, err := uc.storageService.GetPresignedPutURL(ctx, uc.minioBucket, objectKey, uploadURLExpiry)
	if err != nil {
		log.Error("Failed to get presigned PUT URL", "error", err)
		return nil, fmt.Errorf("failed to prepare upload: %w", err) // Internal error
	}

	log.Info("Generated presigned URL for file upload")

	result := &RequestUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}
	return result, nil
}

// CompleteUploadRequest holds the data needed to finalize an upload and create a track record.
type CompleteUploadRequest struct {
	ObjectKey    string
	Title        string
	Description  string
	LanguageCode string
	Level        string // e.g., "A1", "B2"
	DurationMs   int64  // Duration in milliseconds (client needs to provide this, e.g., from HTML5 audio duration property)
	IsPublic     bool
	Tags         []string
	CoverImageURL *string
}

// CompleteUpload finalizes the upload process by creating an AudioTrack record in the database.
func (uc *UploadUseCase) CompleteUpload(ctx context.Context, userID domain.UserID, req CompleteUploadRequest) (*domain.AudioTrack, error) {
	log := uc.logger.With("userID", userID.String(), "objectKey", req.ObjectKey)

	// --- Validation ---
	if req.ObjectKey == "" { return nil, fmt.Errorf("%w: objectKey is required", domain.ErrInvalidArgument)}
	if req.Title == "" { return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidArgument)}
	if req.LanguageCode == "" { return nil, fmt.Errorf("%w: languageCode is required", domain.ErrInvalidArgument)}
	if req.DurationMs <= 0 { return nil, fmt.Errorf("%w: valid durationMs is required", domain.ErrInvalidArgument)}

	// Validate object key prefix belongs to the user (basic security check)
	expectedPrefix := fmt.Sprintf("user-uploads/%s/", userID.String())
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		log.Warn("Attempt to complete upload for object key not belonging to user", "expectedPrefix", expectedPrefix)
		// Return NotFound or PermissionDenied? NotFound might obscure the real issue.
		return nil, fmt.Errorf("%w: invalid object key provided", domain.ErrPermissionDenied)
	}

	// Optional: Check if object actually exists in MinIO?
	// This adds an extra call to MinIO but ensures the client *did* upload successfully.
	// _, err := uc.storageService.GetObjectInfo(ctx, uc.minioBucket, req.ObjectKey) // Assumes GetObjectInfo exists
	// if err != nil {
	//     log.Error("Failed to verify object existence in MinIO or access denied", "error", err)
	//     if errors.Is(err, domain.ErrNotFound) { // Assuming GetObjectInfo returns ErrNotFound
	//         return nil, fmt.Errorf("%w: uploaded file not found in storage", domain.ErrInvalidArgument)
	//     }
	//     return nil, fmt.Errorf("failed to verify upload: %w", err)
	// }

	// Validate Language and Level
	langVO, err := domain.NewLanguage(req.LanguageCode, "") // Name not stored
	if err != nil { return nil, err}
	levelVO := domain.AudioLevel(req.Level)
	if req.Level != "" && !levelVO.IsValid() { // Allow empty level
		return nil, fmt.Errorf("%w: invalid audio level '%s'", domain.ErrInvalidArgument, req.Level)
	}
	// --- End Validation ---


	// --- Create Domain Object ---
	duration := time.Duration(req.DurationMs) * time.Millisecond
	uploaderID := userID // Link track to the uploading user

	track, err := domain.NewAudioTrack(
		req.Title,
		req.Description,
		uc.minioBucket, // Use the configured bucket
		req.ObjectKey,
		langVO,
		levelVO,
		duration,
		&uploaderID,
		req.IsPublic,
		req.Tags,
		req.CoverImageURL,
	)
	if err != nil {
		log.Error("Failed to create AudioTrack domain object", "error", err, "request", req)
		return nil, fmt.Errorf("failed to process track data: %w", err) // Likely internal validation mismatch
	}
	// --- End Create Domain Object ---


	// --- Save to Repository ---
	err = uc.trackRepo.Create(ctx, track)
	if err != nil {
		log.Error("Failed to create audio track record in repository", "error", err, "trackID", track.ID)
		// Check for conflict (e.g., duplicate object key if somehow possible despite UUID)
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("%w: track with this identifier already exists", domain.ErrConflict)
		}
		// Handle other potential errors (e.g., database connection issues)
		return nil, fmt.Errorf("failed to save track information: %w", err) // Internal error
	}
	// --- End Save to Repository ---

	log.Info("Upload completed and track record created", "trackID", track.ID)
	return track, nil
}
```

## `internal/usecase/auth_uc_test.go`

```go
// internal/usecase/auth_uc_test.go
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Import the generated mocks
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
	"log/slog"
    "io" // For discarding logs

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper to create a dummy logger that discards output
func newTestLogger() *slog.Logger {
     return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to create a dummy JWT config
func newTestJWTConfig() config.JWTConfig {
    return config.JWTConfig{AccessTokenExpiry: 1 * time.Hour}
}


func TestAuthUseCase_RegisterWithPassword_Success(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository(t) // testify v1.9+ style
	mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t) // Needed for constructor, but not used here
    logger := newTestLogger()
    cfg := newTestJWTConfig()

	uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

	emailStr := "test@example.com"
	password := "password123"
	name := "Test User"
	emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "hashed_password"
    expectedToken := "jwt_token"

	// Define mock expectations
    // 1. FindByEmail should return NotFound
	mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()
    // 2. HashPassword should succeed
    mockSecHelper.On("HashPassword", mock.Anything, password).Return(hashedPassword, nil).Once()
    // 3. Create should succeed (match any user with correct email and hash)
    mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        return user.Email == emailVO && user.HashedPassword != nil && *user.HashedPassword == hashedPassword
    })).Return(nil).Once()
    // 4. GenerateJWT should succeed
    mockSecHelper.On("GenerateJWT", mock.Anything, mock.AnythingOfType("domain.UserID"), cfg.AccessTokenExpiry).Return(expectedToken, nil).Once()


	// Execute the use case method
	user, token, err := uc.RegisterWithPassword(context.Background(), emailStr, password, name)

	// Assert results
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, emailStr, user.Email.String())

	// Verify that mock expectations were met
	mockUserRepo.AssertExpectations(t)
	mockSecHelper.AssertExpectations(t)
}


func TestAuthUseCase_RegisterWithPassword_EmailExists(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()

    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    emailStr := "exists@example.com"
    emailVO, _ := domain.NewEmail(emailStr)
    existingUser := &domain.User{Email: emailVO} // Dummy existing user

    // Expect FindByEmail to find the user
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(existingUser, nil).Once()

    // Execute
    user, token, err := uc.RegisterWithPassword(context.Background(), emailStr, "password", "name")

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict)
    assert.Nil(t, user)
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertNotCalled(t, "HashPassword", mock.Anything, mock.Anything) // Should not be called
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)         // Should not be called
}


func TestAuthUseCase_LoginWithPassword_Success(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()

    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    emailStr := "test@example.com"
    password := "password123"
    emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "hashed_password" // Assume this is the correct hash
    userID := domain.NewUserID()
    foundUser := &domain.User{
        ID:             userID,
        Email:          emailVO,
        HashedPassword: &hashedPassword,
        AuthProvider:   domain.AuthProviderLocal,
    }
    expectedToken := "jwt_token"

    // Expectations
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundUser, nil).Once()
    mockSecHelper.On("CheckPasswordHash", mock.Anything, password, hashedPassword).Return(true).Once()
    mockSecHelper.On("GenerateJWT", mock.Anything, userID, cfg.AccessTokenExpiry).Return(expectedToken, nil).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, password)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedToken, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
}


func TestAuthUseCase_LoginWithPassword_NotFound(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()

    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)
    emailStr := "notfound@example.com"
    emailVO, _ := domain.NewEmail(emailStr)

    // Expect FindByEmail to return NotFound
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, "password")

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrAuthenticationFailed) // Should map NotFound to AuthFailed
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertNotCalled(t, "CheckPasswordHash", mock.Anything, mock.Anything, mock.Anything)
    mockSecHelper.AssertNotCalled(t, "GenerateJWT", mock.Anything, mock.Anything, mock.Anything)
}

 func TestAuthUseCase_LoginWithPassword_WrongPassword(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    // ... setup uc ...
    emailStr := "test@example.com"
    password := "wrongpassword"
    emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "correct_hashed_password"
    foundUser := &domain.User{ /* ... setup user ... */ HashedPassword: &hashedPassword, AuthProvider: domain.AuthProviderLocal}

    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundUser, nil).Once()
    // Expect CheckPasswordHash to return false
    mockSecHelper.On("CheckPasswordHash", mock.Anything, password, hashedPassword).Return(false).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, password)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t) // CheckPasswordHash was called
    mockSecHelper.AssertNotCalled(t, "GenerateJWT", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthUseCase_AuthenticateWithGoogle_NewUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google123"
    googleEmail := "new.google.user@example.com"
    googleName := "Google User"
    expectedJWT := "new_user_jwt"

    // Mock ExternalAuthService verification
    extInfo := &port.ExternalUserInfo{
        Provider:        domain.AuthProviderGoogle,
        ProviderUserID:  googleUserID,
        Email:           googleEmail,
        IsEmailVerified: true,
        Name:            googleName,
    }
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()

    // Mock User Repo: FindByProviderID should return NotFound
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(nil, domain.ErrNotFound).Once()

    // Mock User Repo: FindByEmail should return NotFound
    emailVO, _ := domain.NewEmail(googleEmail)
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()

    // Mock User Repo: Create should succeed
    // We capture the user being created to use its ID for JWT generation mock
    var createdUser *domain.User
    mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        createdUser = user // Capture the created user
        return user.GoogleID != nil && *user.GoogleID == googleUserID &&
               user.Email.String() == googleEmail &&
               user.AuthProvider == domain.AuthProviderGoogle
    })).Return(nil).Once()

    // Mock Sec Helper: GenerateJWT should succeed (using captured user ID)
    mockSecHelper.On("GenerateJWT", mock.Anything, mock.AnythingOfType("domain.UserID"), cfg.AccessTokenExpiry).
        Run(func(args mock.Arguments) {
            // Ensure the correct user ID is passed before returning
            passedUserID := args.Get(1).(domain.UserID)
            assert.Equal(t, createdUser.ID, passedUserID, "GenerateJWT called with wrong user ID")
        }).
        Return(expectedJWT, nil).
        Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.NoError(t, err)
    assert.True(t, isNew, "Expected isNewUser to be true")
    assert.Equal(t, expectedJWT, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
}


func TestAuthUseCase_AuthenticateWithGoogle_ExistingGoogleUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google123"
    googleEmail := "exists.google.user@example.com"
    expectedJWT := "existing_user_jwt"
    existingUserID := domain.NewUserID()

    extInfo := &port.ExternalUserInfo{ /* ... setup extInfo ... */ ProviderUserID: googleUserID, Email: googleEmail}
    foundUser := &domain.User{ID: existingUserID, GoogleID: &googleUserID, AuthProvider: domain.AuthProviderGoogle}

    // Mock ExternalAuthService
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()

    // Mock User Repo: FindByProviderID finds the user
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(foundUser, nil).Once()

    // Mock Sec Helper: GenerateJWT for the existing user
    mockSecHelper.On("GenerateJWT", mock.Anything, existingUserID, cfg.AccessTokenExpiry).Return(expectedJWT, nil).Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.NoError(t, err)
    assert.False(t, isNew, "Expected isNewUser to be false")
    assert.Equal(t, expectedJWT, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
    // Ensure FindByEmail and Create were not called
    mockUserRepo.AssertNotCalled(t, "FindByEmail", mock.Anything, mock.Anything)
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAuthUseCase_AuthenticateWithGoogle_LinkToLocalUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google456"
    googleEmail := "local.user@example.com" // Email matches existing local user
    expectedJWT := "linked_user_jwt"
    existingUserID := domain.NewUserID()
    emailVO, _ := domain.NewEmail(googleEmail)

    extInfo := &port.ExternalUserInfo{ ProviderUserID: googleUserID, Email: googleEmail, IsEmailVerified: true} // Email verified
    foundLocalUser := &domain.User{ID: existingUserID, Email: emailVO, AuthProvider: domain.AuthProviderLocal, GoogleID: nil} // Local user, no Google ID yet

    // Mock ExternalAuthService
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()

    // Mock User Repo: FindByProviderID returns NotFound
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(nil, domain.ErrNotFound).Once()

    // Mock User Repo: FindByEmail finds the local user
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundLocalUser, nil).Once()

    // Mock User Repo: Update should be called to link the Google ID
    mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        return user.ID == existingUserID && user.GoogleID != nil && *user.GoogleID == googleUserID
    })).Return(nil).Once()

    // Mock Sec Helper: GenerateJWT for the existing (now linked) user
    mockSecHelper.On("GenerateJWT", mock.Anything, existingUserID, cfg.AccessTokenExpiry).Return(expectedJWT, nil).Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.NoError(t, err)
    assert.False(t, isNew, "Expected isNewUser to be false for linking")
    assert.Equal(t, expectedJWT, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything) // Ensure Create is not called
}


func TestAuthUseCase_AuthenticateWithGoogle_EmailConflictDifferentGoogleID(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    attemptedGoogleUserID := "google-attempt-123"
    existingGoogleUserID := "google-exists-456"
    conflictEmail := "conflict.email@example.com"
    emailVO, _ := domain.NewEmail(conflictEmail)

    extInfo := &port.ExternalUserInfo{ ProviderUserID: attemptedGoogleUserID, Email: conflictEmail }
    // User exists with the same email but is already linked to a *different* Google ID
    existingUserWithDifferentGoogleID := &domain.User{
        ID: domain.NewUserID(),
        Email: emailVO,
        GoogleID: &existingGoogleUserID, // Linked to a different Google ID
        AuthProvider: domain.AuthProviderGoogle,
    }

    // Mocks
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, attemptedGoogleUserID).Return(nil, domain.ErrNotFound).Once()
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(existingUserWithDifferentGoogleID, nil).Once()

     // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict) // Expect conflict
    assert.False(t, isNew)
    assert.Empty(t, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    // Ensure Update and Create were not called
    mockUserRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
    mockSecHelper.AssertNotCalled(t, "GenerateJWT", mock.Anything, mock.Anything, mock.Anything)
}

// TODO: Add more AuthenticateWithGoogle tests:
// - Verification fails from ExternalAuthService
// - Email exists, but IsEmailVerified is false (if policy enforces verification)
// - Email exists, but provider is different and GoogleID is nil -> Link (covered by LinkToLocalUser)
// - Email exists, provider is different and GoogleID is NOT nil -> Conflict
// - Repository errors (other than NotFound) during find/create/update
```

## `internal/usecase/mocks/auth_uc_mock.go`

```go
// internal/usecase/mocks/auth_uc_mock.go
package mocks

import (
	"context"
	
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockAuthUseCase is a mock implementation of the AuthUseCase interface
type MockAuthUseCase struct {
	mock.Mock
}

// NewMockAuthUseCase creates a new MockAuthUseCase
func NewMockAuthUseCase(t interface{}) *MockAuthUseCase {
	return &MockAuthUseCase{}
}

// RegisterWithPassword is a mock implementation of the RegisterWithPassword method
func (m *MockAuthUseCase) RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error) {
	args := m.Called(ctx, emailStr, password, name)
	
	var user *domain.User
	if args.Get(0) != nil {
		user = args.Get(0).(*domain.User)
	}
	
	return user, args.String(1), args.Error(2)
}

// LoginWithPassword is a mock implementation of the LoginWithPassword method
func (m *MockAuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (string, error) {
	args := m.Called(ctx, emailStr, password)
	return args.String(0), args.Error(1)
}

// AuthenticateWithGoogle is a mock implementation of the AuthenticateWithGoogle method
func (m *MockAuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIDToken string) (string, bool, error) {
	args := m.Called(ctx, googleIDToken)
	return args.String(0), args.Bool(1), args.Error(2)
}
```

## `pkg/httputil/.keep`

```

```

## `pkg/httputil/response.go`

```go
// pkg/httputil/response.go
package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
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
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "The requested resource was not found."
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "RESOURCE_CONFLICT", "A conflict occurred with the current state of the resource." // Generic message
		// Consider more specific messages based on context if err wraps more info
		// if strings.Contains(err.Error(), "email") { message = "Email already exists."} ...
	case errors.Is(err, domain.ErrInvalidArgument):
		return http.StatusBadRequest, "INVALID_INPUT", err.Error() // Use error message directly for validation details
	case errors.Is(err, domain.ErrPermissionDenied):
		return http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action."
	case errors.Is(err, domain.ErrAuthenticationFailed):
		return http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication failed. Please check your credentials." // Use 401 for auth failure
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required. Please log in." // Also 401
	default:
		// Any other error is treated as an internal server error
		return http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred."
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

## `pkg/logger/logger.go`

```go
// pkg/logger/logger.go
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
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

## `pkg/logger/.keep`

```

```

## `pkg/security/hasher_test.go`

```go
// pkg/security/hasher_test.go
package security_test

import (
	"testing"
    "log/slog"
    "io"
	"github.com/yvanyang/language-learning-player-backend/pkg/security" // Adjust

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a dummy logger
func newTestLogger() *slog.Logger {
     return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestBcryptHasher(t *testing.T) {
	hasher := security.NewBcryptHasher(newTestLogger())
	password := "mysecretpassword"

	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, password, hash) // Ensure it's hashed

	// Check correct password
	match := hasher.CheckPasswordHash(password, hash)
	assert.True(t, match, "Correct password should match")

	// Check incorrect password
	match = hasher.CheckPasswordHash("wrongpassword", hash)
	assert.False(t, match, "Incorrect password should not match")

    // Check empty password
	match = hasher.CheckPasswordHash("", hash)
	assert.False(t, match, "Empty password should not match")

    // Check with invalid hash (should return false, log warning)
    match = hasher.CheckPasswordHash(password, "invalidhashformat")
    assert.False(t, match, "Invalid hash format should result in mismatch")
}
```

## `pkg/security/hasher.go`

```go
// pkg/security/hasher.go
package security

import (
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
		// Log mismatch errors only at debug level if desired, other errors as warnings/errors
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			h.logger.Debug("Password hash mismatch", "error", err) // Optional: Log mismatch at debug level
		} else {
			h.logger.Warn("Error comparing password hash", "error", err)
		}
		return false
	}
	return true
}
```

## `pkg/security/.keep`

```

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
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
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

## `pkg/security/security.go`

```go
// pkg/security/security.go
package security

import (
	"context"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
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
	// Context is not used by bcrypt, but included for interface consistency
	return s.hasher.HashPassword(password)
}

// CheckPasswordHash compares a plain password with a stored hash.
func (s *Security) CheckPasswordHash(ctx context.Context, password, hash string) bool {
	// Context is not used by bcrypt, but included for interface consistency
	return s.hasher.CheckPasswordHash(password, hash)
}

// GenerateJWT creates a signed JWT (Access Token) for the given user ID.
func (s *Security) GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error) {
	// Context could potentially be used for tracing in the future
	return s.jwt.GenerateJWT(userID, duration)
}

// VerifyJWT validates a JWT string and returns the UserID contained within.
func (s *Security) VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error) {
	// Context could potentially be used for tracing in the future
	return s.jwt.VerifyJWT(tokenString)
}

// Compile-time check to ensure Security satisfies the port.SecurityHelper interface
var _ port.SecurityHelper = (*Security)(nil)
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

## `pkg/validation/.keep`

```

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

## `sqlc_queries/.keep`

```

```

