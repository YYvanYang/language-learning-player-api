好的，我们可以创建一个名为 `Makefile.js` 的文件，使用 `shelljs/make` 来模拟 Makefile 的功能。这需要你安装 `shelljs` 作为项目的开发依赖。

**1. 安装 ShellJS:**

如果你的 Go 项目还没有 `package.json`，先创建一个：

```bash
npm init -y
```

然后安装 `shelljs`:

```bash
npm install --save-dev shelljs
```

**2. 创建 `Makefile.js` 文件:**

在你的项目根目录下创建 `Makefile.js` 文件，并粘贴以下内容。请仔细阅读注释，并**修改**需要定制的部分（特别是 Docker Hub 用户名和环境变量的处理）。

```javascript
/**
 * @fileoverview Build file for the Language Learning Player API (Node.js version)
 * Mimics the functionality of the original Makefile using shelljs.
 */
/* eslint no-use-before-define: "off", no-console: "off" -- CLI */
"use strict";

//------------------------------------------------------------------------------
// Requirements
//------------------------------------------------------------------------------

const path = require("node:path");
const os = require("node:os");

require("shelljs/make");
/* global target -- global.target is declared in `shelljs/make.js` */

// shelljs functions: https://github.com/shelljs/shelljs#command-reference
const {
	cat, // eslint-disable-line no-unused-vars -- used implicitly by .to()
	cd,
	cp, // eslint-disable-line no-unused-vars
	echo,
	exec,
	exit,
	find, // eslint-disable-line no-unused-vars
	ls, // eslint-disable-line no-unused-vars
	mkdir,
	pwd, // eslint-disable-line no-unused-vars
	rm,
	test,
	which,
	sleep, // Added for waiting
} = require("shelljs");

//------------------------------------------------------------------------------
// Settings & Configuration
//------------------------------------------------------------------------------

// --- Project Configuration ---
const BINARY_NAME = "language-player-api";
const CMD_PATH = "./cmd/api";
const OUTPUT_DIR = "./build";
const MIGRATIONS_DIR = "migrations";
const SWAG_OUTPUT_DIR = "./docs";
const COVERAGE_FILE = "coverage.out";
const MOCKERY_CONFIG = ".mockery.yaml";
const ENV_FILE = ".env"; // For local docker run

// --- Go Configuration ---
const GO_BUILD_FLAGS = "-ldflags='-w -s' -trimpath";
const GO_TEST_FLAGS = `./... -coverprofile=${COVERAGE_FILE}`;
const GO_ENV_LINUX_AMD64 = "CGO_ENABLED=0 GOOS=linux GOARCH=amd64";
const GO_ENV_LINUX_ARM64 = "CGO_ENABLED=0 GOOS=linux GOARCH=arm64"; // For Raspberry Pi

// --- Tools Configuration ---
// Assumes tools are installed in GOBIN or PATH
// You might need to adjust if using a specific GOBIN path
const GOBIN = process.env.GOBIN || path.join(os.homedir(), "go", "bin"); // Basic guess for GOBIN
const MIGRATE_CMD = process.env.MIGRATE_CMD || "migrate";
const SWAG_CMD = process.env.SWAG_CMD || "swag";
const LINT_CMD = process.env.LINT_CMD || "golangci-lint";
const VULN_CMD = process.env.VULN_CMD || "govulncheck";
const MOCKERY_CMD = process.env.MOCKERY_CMD || "mockery";
const GOIMPORTS_CMD = process.env.GOIMPORTS_CMD || "goimports"; // Optional for fmt

// --- Docker Configuration ---
// !!! IMPORTANT: SET YOUR DOCKER HUB USERNAME/REGISTRY HERE !!!
const DOCKER_IMAGE_REPO =
	process.env.DOCKER_IMAGE_REPO || "your-dockerhub-username"; // <<< CHANGE THIS
const DOCKER_IMAGE_NAME = `${DOCKER_IMAGE_REPO}/${BINARY_NAME}`;
const DOCKER_IMAGE_TAG = process.env.DOCKER_IMAGE_TAG || "latest";
const DOCKER_TARGET_PLATFORM =
	process.env.DOCKER_TARGET_PLATFORM || "linux/arm64"; // Target platform for Pi
const DOCKER_BUILDX_PUSH_FLAGS = `--platform ${DOCKER_TARGET_PLATFORM} --push`;
const DOCKER_BUILDX_BUILD_FLAGS = `--platform ${DOCKER_TARGET_PLATFORM} --load`;
const DOCKER_CONTAINER_NAME = BINARY_NAME; // For 'make docker-run'

// --- Local Dependencies Configuration (for deps-run/stop) ---
const PG_CONTAINER_NAME = "language-learner-postgres";
const PG_USER = "user";
const PG_PASSWORD = "password";
const PG_DB = "language_learner_db";
const PG_PORT = 5432;
const PG_VERSION = 16;
const PG_READY_TIMEOUT_S = 30;
const LOCAL_DB_URL = `postgresql://${PG_USER}:${PG_PASSWORD}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable`;

const MINIO_CONTAINER_NAME = "language-learner-minio";
const MINIO_ROOT_USER = "minioadmin";
const MINIO_ROOT_PASSWORD = "minioadmin";
const MINIO_API_PORT = 9000;
const MINIO_CONSOLE_PORT = 9001;
const MINIO_BUCKET_NAME = "language-audio";
const MINIO_READY_TIMEOUT_S = 30;

// --- Environment Variables ---
// Read DATABASE_URL for migrations, default to local one if not set externally
const DATABASE_URL = process.env.DATABASE_URL || LOCAL_DB_URL;

// Helper function to check command execution result
function run(command, errorMessage) {
	errorMessage = errorMessage || `Command failed: ${command}`;
	echo(`---> Executing: ${command}`);
	const result = exec(command);
	if (result.code !== 0) {
		echo(`Error: ${errorMessage}`);
		echo(`Stderr: ${result.stderr}`);
		exit(1);
	}
	return result;
}

// Helper function to check if a command exists
function checkCommand(commandName) {
	if (!which(commandName)) {
		echo(
			`Error: Command "${commandName}" not found. Please ensure it's installed and in your PATH.`,
		);
		exit(1);
	}
}

// Helper function to check if a Go tool exists, installs if not
function ensureGoTool(toolName, installCommand) {
	if (!which(toolName)) {
		echo(`>>> Tool "${toolName}" not found. Installing...`);
		run(
			installCommand,
			`Failed to install ${toolName}. Please check network connectivity and Go proxy settings.`,
		);
		echo(`>>> ${toolName} installed successfully.`);
	} else {
		echo(`>>> Tool "${toolName}" is already installed.`);
	}
}

// Helper function to wait for a Docker container's health check (simplified)
function waitForContainer(containerName, checkCommand, timeoutSeconds) {
	echo(`>>> Waiting for ${containerName} to be ready (max ${timeoutSeconds}s)...`);
	let elapsed = 0;
	while (elapsed < timeoutSeconds) {
		// Use silent:true to avoid spamming output on checks
		if (exec(checkCommand, { silent: true }).code === 0) {
			echo(`>>> ${containerName} is ready.`);
			return true;
		}
		sleep(1); // Wait 1 second
		elapsed++;
	}
	echo(`>>> ERROR: ${containerName} did not become ready in time.`);
	// Print logs on failure
	exec(`docker logs ${containerName}`);
	return false;
}

//------------------------------------------------------------------------------
// Tasks (Targets)
//------------------------------------------------------------------------------

/**
 * Default target: Show help message.
 */
target.all = function () {
	target.help();
};

/**
 * Install necessary Go tools.
 */
target.tools = function () {
	echo(">>> Ensuring Go tools are installed...");
	ensureGoTool(
		MIGRATE_CMD,
		"go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest",
	);
	ensureGoTool(SWAG_CMD, "go install github.com/swaggo/swag/cmd/swag@latest");
	ensureGoTool(
		LINT_CMD,
		"go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
	);
	ensureGoTool(
		VULN_CMD,
		"go install golang.org/x/vuln/cmd/govulncheck@latest",
	);
	ensureGoTool(
		MOCKERY_CMD,
		"go install github.com/vektra/mockery/v3@v3.0.0",
	);
	// Optional: Check for goimports
	if (!which(GOIMPORTS_CMD)) {
		echo(">>> Optional: goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest");
	}
	echo(">>> Go tools check complete.");
};

/**
 * Clean build artifacts.
 */
target.clean = function () {
	echo(">>> Cleaning build artifacts...");
	if (test("-d", OUTPUT_DIR)) {
		rm("-rf", OUTPUT_DIR);
	}
	if (test("-f", COVERAGE_FILE)) {
		rm(COVERAGE_FILE);
	}
	echo(">>> Clean finished.");
};

/**
 * Build the Go binary for linux/amd64 (standard).
 */
target.build = function () {
	target.clean();
	target.tools(); // Ensure tools are present (though not strictly needed for build)
	echo(">>> Building binary for linux/amd64...");
	mkdir("-p", OUTPUT_DIR);
	const buildCmd = `${GO_ENV_LINUX_AMD64} go build ${GO_BUILD_FLAGS} -o ${path.join(OUTPUT_DIR, BINARY_NAME)} ${CMD_PATH}`;
	run(buildCmd);
	echo(`>>> Binary built at ${path.join(OUTPUT_DIR, BINARY_NAME)}`);
};

/**
 * Run the application locally using 'go run'. Requires deps running.
 */
target.run = function () {
	target.tools(); // Ensure tools (like migrate for potential future auto-migration)
	echo(">>> Running application locally (using go run)...");
	echo(">>> Ensure dependencies (Postgres/MinIO) are running (use 'node Makefile.js deps-run').");
	// Use shelljs exec which will run it in the foreground
	// Set APP_ENV=development using cross-env or direct assignment depending on OS
	const runCmd = `APP_ENV=development go run ${path.join(CMD_PATH, "main.go")}`;
	// Note: This exec call will block until the server is stopped (Ctrl+C).
	// For background execution, consider using Node's child_process directly.
	exec(runCmd);
};

/**
 * Check if DATABASE_URL environment variable is set.
 */
function checkDbUrl() {
	if (!DATABASE_URL || DATABASE_URL === LOCAL_DB_URL) {
		if (!process.env.DATABASE_URL) {
			// Check the external env var specifically
			echo(">>> ERROR: DATABASE_URL environment variable is not set.");
			echo(">>> Please set it before running migrations, e.g.:");
			echo(">>> export DATABASE_URL='postgresql://user:password@host:port/db?sslmode=disable'");
			exit(1);
		}
	}
	echo(`>>> Using DATABASE_URL: ${DATABASE_URL}`); // Show the URL being used
}

/**
 * Create a new migration file. Use `NAME=... node Makefile.js migrate-create`
 */
target["migrate-create"] = function () {
	const migrationName = process.env.NAME;
	if (!migrationName) {
		echo(">>> ERROR: Please provide migration name via NAME environment variable.");
		echo(">>> Example: NAME=my_new_migration node Makefile.js migrate-create");
		exit(1);
	}
	target.tools(); // Ensure migrate CLI is installed
	echo(`>>> Creating migration file: ${migrationName}...`);
	const migrateCmd = `${MIGRATE_CMD} create -ext sql -dir ${MIGRATIONS_DIR} -seq ${migrationName}`;
	run(migrateCmd);
	echo(">>> Migration file created.");
};

/**
 * Apply all up migrations. Reads DATABASE_URL from environment.
 */
target["migrate-up"] = function () {
	target.tools();
	checkDbUrl();
	echo(">>> Applying database migrations...");
	const migrateCmd = `${MIGRATE_CMD} -database "${DATABASE_URL}" -path ${MIGRATIONS_DIR} up`;
	run(migrateCmd);
	echo(">>> Migrations applied.");
};

/**
 * Roll back the last migration. Reads DATABASE_URL from environment.
 */
target["migrate-down"] = function () {
	target.tools();
	checkDbUrl();
	echo(">>> Reverting last database migration...");
	const migrateCmd = `${MIGRATE_CMD} -database "${DATABASE_URL}" -path ${MIGRATIONS_DIR} down 1`;
	run(migrateCmd);
	echo(">>> Last migration reverted.");
};

/**
 * Force migration version. Use `VERSION=... node Makefile.js migrate-force`
 */
target["migrate-force"] = function () {
	const version = process.env.VERSION;
	if (!version) {
		echo(">>> ERROR: Please provide migration version via VERSION environment variable.");
		echo(">>> Example: VERSION=000001 node Makefile.js migrate-force");
		exit(1);
	}
	target.tools();
	checkDbUrl();
	echo(`>>> Forcing migration version to ${version}...`);
	const migrateCmd = `${MIGRATE_CMD} -database "${DATABASE_URL}" -path ${MIGRATIONS_DIR} force ${version}`;
	run(migrateCmd);
	echo(">>> Migration version forced.");
};

/**
 * Generate OpenAPI docs using swag.
 */
target["generate-swag"] = function () {
	target.tools(); // Ensure swag is installed
	echo(">>> Generating OpenAPI docs using swag...");
	const swagCmd = `${SWAG_CMD} init -g ${path.join(CMD_PATH, "main.go")} --output ${SWAG_OUTPUT_DIR}`;
	run(swagCmd);
	echo(`>>> OpenAPI docs generated in ${SWAG_OUTPUT_DIR}.`);
};

/**
 * Alias for generate-swag.
 */
target.swagger = target["generate-swag"];

/**
 * Generate mocks using mockery.
 */
target["generate-mocks"] = function () {
	target.tools(); // Ensure mockery is installed
	echo(">>> Generating mocks using mockery (reading .mockery.yaml)...");
	if (!test("-f", MOCKERY_CONFIG)) {
		echo(`>>> ERROR: ${MOCKERY_CONFIG} config file not found. Please create it.`);
		exit(1);
	}
	const mocksDir = path.join("internal", "mocks", "port");
	echo(`>>> Cleaning existing mocks in ${mocksDir}...`);
	rm("-rf", mocksDir);
	mkdir("-p", mocksDir);
	echo(">>> Running mockery...");
	run(MOCKERY_CMD);
	echo(">>> Mocks generation complete (check output above for errors).");
};

/**
 * Run all code generators.
 */
target.generate = function () {
	target["generate-swag"]();
	target["generate-mocks"]();
};

/**
 * Run all tests.
 */
target.test = function () {
	target.tools();
	echo(">>> Running all tests (unit + integration)...");
	const testCmd = `go test ${GO_TEST_FLAGS}`;
	run(testCmd);
	echo(`>>> Tests complete. Coverage report generated at ${COVERAGE_FILE}`);
};

/**
 * Run unit tests only.
 */
target["test-unit"] = function () {
	target.tools();
	echo(">>> Running unit tests (placeholder)...");
	const testCmd = `go test ${GO_TEST_FLAGS} -short`;
	run(testCmd);
};

/**
 * Run integration tests (requires Docker dependencies).
 */
target["test-integration"] = function () {
	target.tools();
	echo(">>> Starting dependencies for integration tests...");
	target["deps-run"]();
	echo(">>> Waiting for dependencies to be fully ready...");
	sleep(5); // Give extra time
	echo(">>> Running integration tests (replace with your specific command)...");
	// Specify integration test package(s) or use build tags
	// Example: run(`go test ./internal/adapter/repository/postgres/... -v`);
	// Or: run(`go test ./... -tags=integration -v`);
	run(`go test ./internal/adapter/repository/postgres/... -v`); // Example command
	echo(">>> Integration tests finished.");
	target["deps-stop"]();
};

/**
 * Show test coverage report in browser (prints command).
 */
target["test-cover"] = function () {
	target.test();
	echo(">>> To view coverage report, run:");
	echo(`>>> go tool cover -html=${COVERAGE_FILE}`);
	// Note: Automatically opening a browser is complex and platform-dependent.
	// run(`go tool cover -html=${COVERAGE_FILE}`); // This might work on some systems
};

/**
 * Run linter.
 */
target.lint = function () {
	target.tools(); // Ensure linter is installed
	echo(">>> Running linter...");
	run(`${LINT_CMD} run ./...`);
};

/**
 * Format Go code.
 */
target.fmt = function () {
	echo(">>> Formatting Go code...");
	run("go fmt ./...");
	if (which(GOIMPORTS_CMD)) {
		run(`${GOIMPORTS_CMD} -w .`);
	} else {
		echo(">>> Skipping goimports (not found).");
	}
};

/**
 * Check for known vulnerabilities.
 */
target["check-vuln"] = function () {
	target.tools(); // Ensure vulncheck is installed
	echo(">>> Checking for vulnerabilities...");
	run(`${VULN_CMD} ./...`);
};

/**
 * Build Docker image for the host architecture.
 */
target["docker-build"] = function () {
	echo(`>>> Building Docker image for host architecture [${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}]...`);
	run(`docker build -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .`);
	echo(">>> Docker image built.");
};

/**
 * Build Docker image for ARM64 (for Pi), loading locally.
 */
target["docker-build-arm64"] = function () {
	echo(`>>> Building Docker image for ARM64 (local load) [${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}]...`);
	echo(">>> Note: Requires Docker Buildx setup (usually available with Docker Desktop).");
	run(
		`docker buildx build ${DOCKER_BUILDX_BUILD_FLAGS} -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .`,
	);
	echo(">>> ARM64 Docker image built and loaded locally.");
};

/**
 * Build Docker image for ARM64 AND push to registry.
 */
target["docker-build-push-arm64"] = function () {
	echo(`>>> Building AND PUSHING Docker image for ARM64 [${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}]...`);
	if (DOCKER_IMAGE_REPO === "your-dockerhub-username") {
		echo(">>> WARNING: DOCKER_IMAGE_REPO is set to the default placeholder.");
		echo(">>> Please edit Makefile.js and set your Docker Hub username or registry.");
		// exit(1); // Optional: make it a hard failure
	}
	echo(`>>> Pushing to: ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}`);
	echo(">>> Note: Requires Docker Buildx setup and 'docker login'.");
	run(
		`docker buildx build ${DOCKER_BUILDX_PUSH_FLAGS} -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .`,
	);
	echo(">>> ARM64 Docker image built and pushed.");
};

/**
 * Run Docker container locally (uses host image, requires deps).
 */
target["docker-run"] = function () {
	echo(`>>> Running Docker container [${DOCKER_CONTAINER_NAME}] using image [${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}]...`);
	echo(">>> Note: Assumes image for host arch exists. Run 'node Makefile.js docker-build' first.");
	echo(">>> Note: Requires dependencies (Postgres/MinIO) to be running. Use 'node Makefile.js deps-run'.");
	if (!test("-f", ENV_FILE)) {
		echo(`>>> Warning: ${ENV_FILE} not found. Container might not have necessary environment variables.`);
	}
	// Use --network=host for simple local access to DB/MinIO on localhost
	const runCmd = `docker run -d --name ${DOCKER_CONTAINER_NAME} -p 8080:8080 --network=host --env-file ${ENV_FILE} ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}`;
	run(runCmd);
	echo(`>>> Container ${DOCKER_CONTAINER_NAME} started. Use 'node Makefile.js docker-stop' to stop.`);
};

/**
 * Stop and remove the running application container.
 */
target["docker-stop"] = function () {
	echo(`>>> Stopping and removing Docker container [${DOCKER_CONTAINER_NAME}]...`);
	// Use || true equivalent by ignoring non-zero exit code for stop/rm
	exec(`docker stop ${DOCKER_CONTAINER_NAME}`);
	exec(`docker rm ${DOCKER_CONTAINER_NAME}`);
	echo(">>> Container stopped and removed (if it existed).");
};

/**
 * Push the host architecture Docker image to registry.
 */
target["docker-push"] = function () {
	target["docker-build"](); // Ensure host image is built
	echo(`>>> Pushing HOST architecture Docker image [${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}]...`);
	if (DOCKER_IMAGE_REPO === "your-dockerhub-username") {
		echo(">>> WARNING: DOCKER_IMAGE_REPO is set to the default placeholder.");
		echo(">>> Please edit Makefile.js and set your Docker Hub username or registry.");
		// exit(1); // Optional: make it a hard failure
	}
	echo(`>>> Pushing to: ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}`);
	echo(">>> Note: Use 'node Makefile.js docker-build-push-arm64' to push the ARM64 image for Raspberry Pi.");
	run(`docker push ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}`);
	echo(">>> Image pushed.");
};

/**
 * Start local PostgreSQL container.
 */
target["docker-postgres-run"] = function () {
	echo(`>>> Ensuring PostgreSQL container [${PG_CONTAINER_NAME}] is running...`);
	exec(`docker stop ${PG_CONTAINER_NAME}`); // Ignore errors if not running
	exec(`docker rm ${PG_CONTAINER_NAME}`); // Ignore errors if not exists
	echo(">>> Starting PostgreSQL container...");
	const runCmd = `docker run --name ${PG_CONTAINER_NAME} -e POSTGRES_USER=${PG_USER} -e POSTGRES_PASSWORD=${PG_PASSWORD} -e POSTGRES_DB=${PG_DB} -p ${PG_PORT}:5432 -d postgres:${PG_VERSION}-alpine`;
	run(runCmd);
	const checkCmd = `docker exec ${PG_CONTAINER_NAME} pg_isready -U ${PG_USER} -d ${PG_DB} -q`;
	if (!waitForContainer(PG_CONTAINER_NAME, checkCmd, PG_READY_TIMEOUT_S)) {
		exit(1);
	}
	echo(`>>> PostgreSQL container [${PG_CONTAINER_NAME}] started successfully.`);
	echo(`>>> Connection string for migrations: ${LOCAL_DB_URL}`);
};

/**
 * Stop and remove local PostgreSQL container.
 */
target["docker-postgres-stop"] = function () {
	echo(`>>> Stopping and removing PostgreSQL container [${PG_CONTAINER_NAME}]...`);
	exec(`docker stop ${PG_CONTAINER_NAME}`);
	exec(`docker rm ${PG_CONTAINER_NAME}`);
	echo(">>> PostgreSQL container stopped and removed (if it existed).");
};

/**
 * Start local MinIO container.
 */
target["docker-minio-run"] = function () {
	echo(`>>> Ensuring MinIO container [${MINIO_CONTAINER_NAME}] is running...`);
	exec(`docker stop ${MINIO_CONTAINER_NAME}`);
	exec(`docker rm ${MINIO_CONTAINER_NAME}`);
	echo(">>> Starting MinIO container...");
	const runCmd = `docker run --name ${MINIO_CONTAINER_NAME} -p ${MINIO_API_PORT}:9000 -p ${MINIO_CONSOLE_PORT}:9001 -e MINIO_ROOT_USER=${MINIO_ROOT_USER} -e MINIO_ROOT_PASSWORD=${MINIO_ROOT_PASSWORD} -d minio/minio server /data --console-address ":9001"`;
	run(runCmd);
	// Use curl within the container or host if available
	const checkUrl = `http://localhost:${MINIO_API_PORT}/minio/health/live`;
	const checkCmd = `curl -s --max-time 1 --fail ${checkUrl}`;
	if (!waitForContainer(MINIO_CONTAINER_NAME, checkCmd, MINIO_READY_TIMEOUT_S)) {
		exit(1);
	}
	echo(`>>> MinIO container [${MINIO_CONTAINER_NAME}] started successfully.`);
	echo(`>>> MinIO API: http://localhost:${MINIO_API_PORT}`);
	echo(`>>> MinIO Console: http://localhost:${MINIO_CONSOLE_PORT} (Login: ${MINIO_ROOT_USER}/${MINIO_ROOT_PASSWORD})`);

	echo(`>>> Ensuring bucket '${MINIO_BUCKET_NAME}' exists...`);
	sleep(5); // Give MinIO extra time
	const mcAliasCmd = `docker exec ${MINIO_CONTAINER_NAME} mc alias set local http://localhost:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD}`;
	const mcMbCmd = `docker exec ${MINIO_CONTAINER_NAME} mc mb local/${MINIO_BUCKET_NAME}`;
	exec(mcAliasCmd); // Ignore errors, might already exist
	if (exec(mcMbCmd).code !== 0) {
		echo(`>>> Bucket '${MINIO_BUCKET_NAME}' likely already exists.`);
	}
};

/**
 * Stop and remove local MinIO container.
 */
target["docker-minio-stop"] = function () {
	echo(`>>> Stopping and removing MinIO container [${MINIO_CONTAINER_NAME}]...`);
	exec(`docker stop ${MINIO_CONTAINER_NAME}`);
	exec(`docker rm ${MINIO_CONTAINER_NAME}`);
	echo(">>> MinIO container stopped and removed (if it existed).");
};

/**
 * Start local development dependencies (PostgreSQL + MinIO).
 */
target["deps-run"] = function () {
	target["docker-postgres-run"]();
	target["docker-minio-run"]();
};

/**
 * Stop local development dependencies.
 */
target["deps-stop"] = function () {
	target["docker-postgres-stop"]();
	target["docker-minio-stop"]();
};

/**
 * Show help message.
 */
target.help = function () {
	echo("Usage: node Makefile.js [target]");
	echo("");
	echo("Local Development (on your PC/Mac/Surface):");
	echo("  run               Run the application using 'go run' (requires dependencies running)");
	echo("  deps-run          Start local PostgreSQL and MinIO containers");
	echo("  deps-stop         Stop local PostgreSQL and MinIO containers");
	echo("  tools             Install necessary Go CLI tools (migrate, swag, lint, vulncheck, mockery)");
	echo("  fmt               Format Go code using go fmt and goimports");
	echo("  lint              Run golangci-lint");
	echo("  check-vuln        Check for known vulnerabilities");
	echo("");
	echo("Code Generation:");
	echo("  generate          Run all code generators (swag, mockery)");
	echo("  generate-swag     Generate OpenAPI docs using swag");
	echo("  generate-mocks    Generate mocks for interfaces using mockery (reads .mockery.yaml)");
	echo("");
	echo("Database Migrations (Local or Remote DB):");
	echo("  migrate-create    Create a new migration file (use NAME=... env var)");
	echo("  migrate-up        Apply database migrations (uses DATABASE_URL env var)");
	echo("  migrate-down      Revert the last database migration (uses DATABASE_URL env var)");
	echo("  migrate-force     Force migration version (use VERSION=... env var, uses DATABASE_URL)");
	echo("");
	echo("Testing (Local):");
	echo("  test              Run all tests and generate coverage");
	echo("  test-unit         Run unit tests (placeholder)");
	echo("  test-integration  Start deps, run integration tests, stop deps");
	echo("  test-cover        Show command to view test coverage report in browser");
	echo("");
	echo("Building:");
	echo("  build             Build the Go binary for linux/amd64 (typical CI/Cloud build)");
	echo("  clean             Remove build artifacts");
	echo("");
	echo("Docker (Application Image):");
	echo("  docker-build           Build the application Docker image for your current machine's architecture");
	echo("  docker-build-arm64     Build the application Docker image for ARM64 (Pi) locally (requires buildx)");
	echo("  docker-build-push-arm64 Build the ARM64 image AND PUSH it to Docker Hub/Registry (for Pi deployment)");
	echo("  docker-run             Run the application container locally (uses host image, .env file, requires deps-run)");
	echo("  docker-stop            Stop and remove the running application container started by 'docker-run'");
	echo("  docker-push            Push the HOST architecture image to Docker Hub/Registry");
	echo("");
	echo("Docker - Dependencies (Local Dev):");
	echo("  docker-postgres-run    Start PostgreSQL in Docker container locally");
	echo("  docker-postgres-stop   Stop and remove PostgreSQL Docker container locally");
	echo("  docker-minio-run       Start MinIO in Docker container locally");
	echo("  docker-minio-stop      Stop and remove MinIO Docker container locally");
	echo("");
	echo("Raspberry Pi Deployment Workflow:");
	echo("  1. (On your Surface) Edit DOCKER_IMAGE_REPO in Makefile.js.");
	echo("  2. (On your Surface) Run 'docker login'.");
	echo("  3. (On your Surface) Run 'node Makefile.js docker-build-push-arm64'.");
	echo("  4. (On your Pi) Follow steps in deployment documentation (create dir, docker-compose.yml, .env).");
	echo("  5. (On your Pi) Run 'docker compose up -d'.");
	echo("  6. (On Pi/Surface) Run database migrations ('migrate-up').");
	echo("  7. (On Pi/Surface) Create MinIO bucket.");
	echo("");
	echo("Other:");
	echo("  help              Show this help message");
};
```

**如何使用 `Makefile.js`:**

1.  **保存文件:** 将上面的代码保存为项目根目录下的 `Makefile.js`。
2.  **安装依赖:** 运行 `npm install --save-dev shelljs`。
3.  **运行目标:** 在终端中，使用 `node Makefile.js <target_name>` 来执行任务，例如：
    *   `node Makefile.js tools` (安装 Go 工具)
    *   `node Makefile.js build` (构建 Go 二进制文件)
    *   `node Makefile.js deps-run` (启动本地依赖)
    *   `node Makefile.js run` (本地运行 Go 程序)
    *   `NAME=add_new_feature node Makefile.js migrate-create` (创建迁移)
    *   `node Makefile.js migrate-up` (应用迁移, 需要设置 `DATABASE_URL` 环境变量)
    *   `node Makefile.js docker-build-push-arm64` (为 Pi 构建并推送 Docker 镜像)
    *   `node Makefile.js help` (显示帮助信息)

**与原 Makefile 的主要差异和注意事项:**

*   **执行方式:** 使用 `node Makefile.js <target>` 而不是 `make <target>`。
*   **环境变量传递:** 对于需要参数的目标（如 `migrate-create`, `migrate-force`），此脚本使用环境变量 (`NAME=...`, `VERSION=...`)。在命令行设置环境变量的方式取决于你的操作系统（Linux/macOS: `VAR=value command`, Windows CMD: `set VAR=value && command`, PowerShell: `$env:VAR="value"; command`）。
*   **依赖管理:** 这个脚本假设你已经通过 `go install` 安装了 Go 工具，或者它会尝试为你安装。
*   **Docker Hub 用户名:** **必须**修改 `Makefile.js` 文件顶部的 `DOCKER_IMAGE_REPO` 常量。
*   **错误处理:** 脚本使用 `exec` 的返回码来判断命令是否成功，并在失败时退出。
*   **异步操作:** `shelljs.exec` 默认是同步的。对于长时间运行的命令（如 `docker buildx push` 或 `go run`），脚本会等待其完成。
*   **平台差异:** `shelljs` 旨在提供跨平台兼容性，但复杂的 shell 逻辑（比如 `Makefile` 中的 `while` 循环和退出逻辑）在 JavaScript 中需要更明确地编写。
*   **依赖关系:** `shelljs/make` 不像 `make` 那样内置文件依赖检查（例如，如果源文件没有改变就不重新构建）。这个脚本每次都会执行命令。

这个 `Makefile.js` 文件提供了一个功能上与原 `Makefile` 非常接近的替代方案，让你可以在 Node.js 环境中管理 Go 项目的构建、测试和部署流程。