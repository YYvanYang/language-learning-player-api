# Language Learning Audio Player - Backend API

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
<!-- Add more badges after setting up CI/CD -->
<!-- [![Build Status](https://img.shields.io/github/actions/workflow/status/your-username/your-repo/.github/workflows/ci.yml?branch=main)](https://github.com/your-username/your-repo/actions) -->
<!-- [![Coverage Status](https://coveralls.io/repos/github/your-username/your-repo/badge.svg?branch=main)](https://coveralls.io/github/your-username/your-repo?branch=main) -->

This repository contains the backend API service for the Language Learning Audio Player application. It provides functionalities for user authentication (including Google Sign-In), managing audio tracks and collections (playlists/courses), tracking user playback progress, and managing bookmarks.

## Overview

The backend is a monolithic Go application built following principles inspired by Clean Architecture / Hexagonal Architecture. It emphasizes separation of concerns, testability, and maintainability. Key features include:

*   **User Authentication:** Secure user registration (email/password), login, and Google OAuth 2.0 integration. Uses JWT for session management.
*   **Audio Content Management:** API endpoints to list, search, and retrieve details of audio tracks and collections.
*   **User Activity Tracking:** Recording user playback progress and managing bookmarks at specific timestamps.
*   **Audio File Handling:** Uses object storage (MinIO / S3-compatible) for storing audio files. Provides secure, temporary access via **presigned URLs**.
*   **API Documentation:** OpenAPI (Swagger) specification for clear API contracts.
*   **Configuration Management:** Flexible configuration using YAML files and environment variables.
*   **Database Migrations:** Managed database schema changes using `golang-migrate`.
*   **Containerized:** Includes a `Dockerfile` for easy containerization and deployment.
*   **Development Tooling:** `Makefile` provides convenient commands for building, testing, running dependencies, and more.

## Architecture

The application follows a layered architecture:

*   **Domain:** Contains core business entities, value objects, and business rules. Pure Go, no external dependencies. (`internal/domain`)
*   **Usecase:** Orchestrates application-specific business logic, coordinating domain objects and repository/service interactions. (`internal/usecase`)
*   **Port:** Defines interfaces (contracts) between layers, especially between use cases and adapters. (`internal/port`)
*   **Adapter:** Connects the application core to the outside world.
    *   **Handlers:** Handle incoming requests (e.g., HTTP) and drive the use cases. (`internal/adapter/handler/http`)
    *   **Repositories:** Implement data persistence logic for specific databases (e.g., PostgreSQL). (`internal/adapter/repository/postgres`)
    *   **Services:** Implement interactions with external services (e.g., MinIO, Google Auth). (`internal/adapter/service/*`)
*   **Pkg:** Shared, non-domain-specific utility code (logging, validation, security). (`pkg`)

*(Refer to the architecture design document for a more detailed explanation).*

## Directory Structure

```
.
├── cmd/api/                # Application entry point (main.go)
├── config/                 # Example configuration files
├── docs/                   # Generated Swagger/OpenAPI documentation
├── internal/               # Internal application code (private)
│   ├── adapter/            # Adapters (handlers, repositories, services)
│   ├── config/             # Configuration loading logic
│   ├── domain/             # Core domain entities and logic
│   ├── port/               # Interfaces (ports) defining layer boundaries
│   └── usecase/            # Application business logic / use cases
├── migrations/             # Database migration files (.sql)
├── pkg/                    # Shared utility libraries (public)
├── sqlc_queries/           # (If using sqlc) SQL queries for code generation
├── test/                   # (Optional) End-to-end / cross-component tests
├── .env.example            # Example environment variables
├── .gitignore
├── config.example.yaml     # Example base configuration
├── config.development.yaml # Default development configuration
├── Dockerfile              # Container build instructions
├── go.mod                  # Go module dependencies
├── go.sum
└── Makefile                # Development task runner
```

## Prerequisites

*   **Go:** Version 1.21 or later ([Installation Guide](https://golang.org/doc/install))
*   **Docker:** For running dependencies locally and building the container ([Installation Guide](https://docs.docker.com/engine/install/))
*   **Make:** For using the Makefile commands. Often pre-installed on Linux/macOS.
*   **`migrate` CLI:** For database migrations ([Installation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)). Can be installed via `make tools`.
*   **(Optional) `swag` CLI:** For generating Swagger docs ([Installation](https://github.com/swaggo/swag#install)). Can be installed via `make tools`.
*   **(Optional) `sqlc` CLI:** If using `sqlc` for database code generation ([Installation](https://docs.sqlc.dev/en/latest/overview/install.html)). Can be installed via `make tools`.

## Local Development Setup

1.  **Clone the Repository:**
    ```bash
    git clone <repository-url>
    cd language-learning-player-backend
    ```

2.  **Install Go Tools:**
    ```bash
    make tools
    ```
    This will install `migrate`, `swag` (if needed), `sqlc` (if needed), `golangci-lint`, etc. using `go install`. Ensure `$GOPATH/bin` or `$GOBIN` is in your system's `PATH`.

3.  **Configuration:**
    *   **Environment Variables:** Copy `.env.example` to `.env` and fill in any necessary secrets or overrides (like `DATABASE_URL`, `JWT_SECRETKEY`). **DO NOT commit the `.env` file.**
      ```bash
      cp .env.example .env
      # Edit .env file
      ```
    *   **YAML Configuration:** The application loads `config.development.yaml` by default when `APP_ENV` is not set or set to `development`. You can modify this file or create `config.<your_env>.yaml` and set `APP_ENV`. Values in `.env` (if mapped correctly, e.g., `DATABASE_DSN`) or environment variables (e.g., `SERVER_PORT`) override YAML values.

4.  **Start Dependencies (Database & MinIO):**
    This command uses Docker to start PostgreSQL and MinIO containers.
    ```bash
    make deps-run
    ```
    *   PostgreSQL will be accessible at `localhost:5432` (or the port specified in `Makefile`) with credentials `user`/`password` and database `language_learner_db`.
    *   MinIO API will be at `http://localhost:9000` and the console at `http://localhost:9001` (credentials: `minioadmin`/`minioadmin`). The required bucket (`language-audio`) will be created automatically.

5.  **Run Database Migrations:**
    Ensure the `DATABASE_URL` environment variable is set correctly (either exported in your shell or defined in the `.env` file if your shell loads it).
    ```bash
    # Option 1: Export directly (replace with your actual DSN from step 4 if different)
    export DATABASE_URL="postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable"

    # Option 2: Rely on .env file if your environment/tool loads it automatically

    # Run migrations
    make migrate-up
    ```

## Running the Application Locally

After completing the setup:

```bash
make run
```

This command uses `go run` to build and start the application. It will load `config.development.yaml` by default. The API server will typically be available at `http://localhost:8080` (or the port specified in the configuration).

## Running Tests

*   **Run all tests (Unit + Integration):** Requires Docker for integration tests.
    ```bash
    make test
    ```
    This also generates a `coverage.out` file.

*   **Run only unit tests (placeholder):**
    ```bash
    make test-unit
    ```
    *(Note: The effectiveness depends on how tests are organized/tagged).*

*   **Run only integration tests:** Requires Docker.
    ```bash
    make test-integration
    ```
    *(Note: Runs tests specifically in the repository package or requires tests tagged with `integration`)*.

*   **View Test Coverage:**
    ```bash
    make test-cover
    ```
    This will run all tests and open the HTML coverage report in your browser.

*   **Run Linter:**
    ```bash
    make lint
    ```

*   **Check for Vulnerabilities:**
    ```bash
    make check-vuln
    ```

## Building the Application

To create a production-ready binary (statically linked for Linux amd64 by default):

```bash
make build
```

The binary will be placed in the `./build/` directory.

## Configuration

The application uses [Viper](https://github.com/spf13/viper) for configuration management.

*   **Files:** Looks for `config.yaml` (base) and `config.<APP_ENV>.yaml` (environment specific) in the current directory. Defaults to `config.development.yaml` if `APP_ENV` is not set.
*   **Environment Variables:** Variables matching config keys (with `.` replaced by `_`, e.g., `SERVER_PORT`) override file values. Sensitive data (DB passwords, JWT secrets, API keys) **should** be configured via environment variables.
*   **Defaults:** Default values are defined in `internal/config/config.go`.

Refer to `config.example.yaml` and `.env.example` for available configuration options.

## Database Migrations

Database schema changes are managed using `golang-migrate/migrate`. Migration files (`.up.sql`, `.down.sql`) are located in the `/migrations` directory.

*   **Requirements:** `migrate` CLI installed (`make tools`) and `DATABASE_URL` environment variable set.
*   **Create a new migration:**
    ```bash
    make migrate-create name=describe_your_change
    ```
*   **Apply migrations:**
    ```bash
    make migrate-up
    ```
*   **Rollback the last migration:**
    ```bash
    make migrate-down
    ```
*   **Force migration version (use with caution):**
    ```bash
    make migrate-force version=<version_number>
    ```

## API Documentation (Swagger)

This project uses `swaggo/swag` to generate OpenAPI (Swagger) documentation from code annotations.

*   **Generate/Update Docs:**
    ```bash
    make swagger
    ```
    This updates the files in the `/docs` directory. Commit these generated files.
*   **Access Docs:** When the application is running locally, access the Swagger UI at:
    [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) (Adjust port if necessary).

## Docker Usage

*   **Build the Docker Image:**
    ```bash
    # Customize DOCKER_IMAGE_NAME in Makefile or pass it as an argument
    make docker-build DOCKER_IMAGE_NAME=your-repo/language-player-api DOCKER_IMAGE_TAG=latest
    ```
*   **Run the Docker Container:** Uses the `.env` file for configuration by default.
    ```bash
    make docker-run
    ```
    The container will run in the background, exposing port 8080.
*   **Stop the Docker Container:**
    ```bash
    make docker-stop
    ```
*   **Push the Image (Requires Login):**
    ```bash
    # Ensure DOCKER_IMAGE_NAME is set correctly
    make docker-push
    ```

## Deployment (Recommended Low-Cost Strategy)

For personal projects or cost-sensitive deployments, the following strategy is recommended:

1.  **Compute:** Deploy the container image to a **Serverless Container Platform** like **Google Cloud Run**. This provides scale-to-zero capability (no cost when idle) and pay-per-use pricing.
2.  **Database:** Use a **Managed PostgreSQL Service** like **Google Cloud SQL** or **AWS RDS**. Start with the smallest available instance size covered by the free tier. Ensure the database is in the *same region* as your compute service.
3.  **Object Storage:** Use **Cloudflare R2**. Its **free egress** is a major cost saver for audio streaming. It's S3-compatible. Alternatively, use AWS S3 or GCS but heavily rely on a CDN.
4.  **CDN:** Use **Cloudflare (Free Plan)** in front of your object storage (R2/S3/GCS). This caches audio files globally, improves user experience, and drastically reduces egress costs from your origin storage.
5.  **Configuration:** Inject secrets (DB password, JWT key, Google credentials) via environment variables provided by the cloud platform (e.g., Cloud Run secret management).
6.  **Deployment:** Build the Docker image in a CI/CD pipeline (e.g., GitHub Actions) and deploy directly to Cloud Run using `gcloud` CLI commands.

*(This is a high-level overview. Refer to the specific cloud provider documentation for detailed setup instructions.)*

## Contributing

Contributions are welcome! Please follow standard Go practices, ensure tests pass (`make test`), code is linted (`make lint`), and documentation is updated if necessary. Open an issue to discuss significant changes before submitting a pull request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details (You need to add a LICENSE file).