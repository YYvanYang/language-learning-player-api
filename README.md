# Language Learning Player Backend

This is the backend API for the Language Learning Player application.

## Development Setup (开发环境设置)

### Go Proxy (可选)

如果你在下载 Go 模块时遇到困难或速度缓慢（尤其是在中国大陆地区），建议设置 `GOPROXY` 环境变量来使用国内的代理。

在你的终端或配置文件 (`.bashrc`, `.zshrc`, etc.) 中添加以下行：

```bash
export GOPROXY=https://goproxy.cn,direct
```

这会将 `goproxy.cn` 设置为首选代理，如果代理不可用，则回退到直接下载 (`direct`)。

## Configuration (配置)

本项目的配置管理采用了 [Viper](https://github.com/spf13/viper)，它支持从多种来源加载配置，并遵循以下优先级顺序（从高到低）：

1.  **环境变量**: 优先级最高。任何设置的环境变量都会覆盖配置文件中的同名设置。
2.  **环境特定配置文件**: 根据 `APP_ENV` 环境变量的值加载，例如 `config.development.yaml` 或 `config.production.yaml`。如果 `APP_ENV` 未设置，则默认为 `development`。
3.  **基础配置文件**: `config.yaml` (如果存在)。
4.  **代码中设置的默认值**: 在 `internal/config/config.go` 中定义，优先级最低。

### 环境变量命名约定

Viper 会自动将配置文件中的键名（例如 `server.port`）映射到环境变量名（例如 `SERVER_PORT`）。它会自动将 `.` 替换为 `_` 并转换为大写。

### 本地开发配置

对于本地开发，你有两种主要方式来设置配置（特别是敏感信息）：

1.  **使用 `.env` 文件 (推荐)**:
    *   复制项目根目录下的 `.env.example` 文件为 `.env`。
    *   编辑 `.env` 文件，填入你的本地开发环境所需的值（例如数据库连接字符串、JWT 密钥等）。**切勿将 `.env` 文件提交到 Git 仓库！**
    *   如果你使用 `make docker-run` 启动服务，Makefile 会自动加载 `.env` 文件中的变量。
2.  **直接导出环境变量**:
    *   在运行 `go run` 或 `make run` 的终端中直接导出环境变量。
    *   示例：
        ```bash
        export JWT_SECRETKEY="your-local-dev-secret"
        export DATABASE_URL="postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable"
        make run
        ```

### 关键环境变量

以下是一些关键的环境变量，你很可能需要在 `.env` 文件或部署环境中进行设置：

*   `APP_ENV`: 指定运行环境（例如 `development`, `production`, `staging`），用于加载对应的 `config.<env>.yaml` 文件。默认为 `development`。
*   `SERVER_PORT`: API 服务器监听的端口 (默认: `8080`)。
*   `DATABASE_DSN` 或 `DATABASE_URL`: PostgreSQL 数据库的连接字符串。代码会优先使用 `DATABASE_DSN`，如果未设置，则会尝试使用 `DATABASE_URL`。格式示例：`postgresql://user:password@host:port/dbname?sslmode=disable`。
*   `JWT_SECRETKEY`: 用于签发和验证 JWT 的密钥。**必须设置为一个强随机字符串，并且在生产环境中保密！**
*   `MINIO_ENDPOINT`: MinIO 服务的地址和端口 (例如 `localhost:9000` 或 `minio.example.com`)。
*   `MINIO_ACCESSKEYID`: MinIO 的 Access Key ID。
*   `MINIO_SECRETACCESSKEY`: MinIO 的 Secret Access Key。
*   `MINIO_BUCKETNAME`: 用于存储音频文件的 MinIO 存储桶名称 (默认: `language-audio`)。
*   `GOOGLE_CLIENTID`: Google OAuth 2.0 Client ID。
*   `GOOGLE_CLIENTSECRET`: Google OAuth 2.0 Client Secret。
*   `CORS_ALLOWEDORIGINS`: 允许访问 API 的前端源地址列表，以逗号分隔 (例如 `"http://localhost:3000,https://your-frontend.com"`)。
*   `LOG_LEVEL`: 日志级别 (例如 `debug`, `info`, `warn`, `error`)。
*   `LOG_JSON`: 是否以 JSON 格式输出日志 (布尔值 `true` 或 `false`)。

请参考 `config.example.yaml` 和 `internal/config/config.go` 中的 `setDefaultValues` 函数以获取完整的配置项列表及其默认值。

## API 文档 (API Documentation)

本项目的 API 使用 [OpenAPI Specification (OAS3)](https://swagger.io/specification/) 进行描述。

### 访问文档

当后端服务在本地运行时 (例如通过 `make run` 或 `make docker-run` 启动)，你可以通过浏览器访问以下地址来查看交互式的 Swagger UI 文档：

[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### 生成/更新文档

项目使用 Go 代码中的注释和 [swaggo/swag](https://github.com/swaggo/swag) 工具来生成 OpenAPI 文档文件 (`docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`)。

如果你修改了 API Handler 中的注释，你需要重新生成文档文件。可以使用以下 Make 命令：

```bash
make swagger
```

**注意:** 建议在提交涉及 API 更改的代码前运行 `make swagger`，以确保文档与代码保持同步。

## Running the API Locally (本地运行 API)

以下步骤将指导你在本地计算机上运行后端 API 服务。

### 1. 先决条件 (Prerequisites)

确保你的开发环境中已安装以下软件：

*   **Go**: 建议使用最新稳定版本 (项目应能正常编译)。
*   **Docker** 和 **Docker Compose**: 用于运行依赖服务 (PostgreSQL, MinIO)。
*   **Make**: 用于执行 `Makefile` 中定义的命令。
*   **Git**: 用于克隆代码仓库。

### 2. 克隆代码仓库 (Clone Repository)

```bash
git clone <your-repository-url>
cd language-learning-player-backend
```

### 3. 安装开发工具 (Install Tools)

运行以下命令安装项目所需的 Go CLI 工具 (如 `migrate`, `swag` 等)：

```bash
make tools
```

### 4. 配置环境变量 (Configure Environment Variables)

API 运行需要一些配置，特别是数据库连接信息和各种服务的密钥。

*   **复制示例文件**:
    ```bash
    cp .env.example .env
    ```
*   **编辑 `.env` 文件**: 打开 `.env` 文件，根据你的本地环境修改以下关键变量：
    *   `DATABASE_URL`: 确保它指向你本地运行的 PostgreSQL 实例 (如果使用 `make deps-run` 启动，默认值通常可用)。格式：`postgresql://user:password@host:port/dbname?sslmode=disable`。
    *   `JWT_SECRETKEY`: 设置一个 **强随机字符串** 作为 JWT 密钥。**不要使用默认值！**
    *   `MINIO_ENDPOINT`: 确保它指向你本地运行的 MinIO 实例 (默认 `localhost:9000`)。
    *   `MINIO_ACCESSKEYID`, `MINIO_SECRETACCESSKEY`: 确保与你本地 MinIO 设置匹配 (默认 `minioadmin`/`minioadmin`)。
    *   `MINIO_BUCKETNAME`: MinIO 存储桶名称 (默认 `language-audio`)。
    *   `GOOGLE_CLIENTID`, `GOOGLE_CLIENTSECRET`: 如果需要测试 Google 登录，请填入你的 Google Cloud OAuth 凭证。
    *   `CORS_ALLOWEDORIGINS`: 添加你本地前端应用的访问地址 (例如 `http://localhost:3000`)。
    *   其他变量按需修改。

**重要提示:** `.env` 文件已被添加到 `.gitignore` 中，确保不会将你的本地敏感配置提交到 Git 仓库。

### 5. 启动依赖服务 (Start Dependencies)

使用 Docker 启动 PostgreSQL 和 MinIO 服务：

```bash
make deps-run
```

这个命令会：
*   在后台启动名为 `language-learner-postgres` 的 PostgreSQL 容器 (端口 5432)。
*   在后台启动名为 `language-learner-minio` 的 MinIO 容器 (API 端口 9000, Console 端口 9001)。
*   等待服务可用。
*   自动在 MinIO 中创建名为 `language-audio` (或你在 `.env` 中配置的名称) 的存储桶。

你可以使用 `docker ps` 查看容器状态，或使用 `docker logs <container_name>` 查看日志。

### 6. 运行数据库迁移 (Run Database Migrations)

在依赖服务（特别是 PostgreSQL）启动并运行后，应用数据库结构变更：

```bash
make migrate-up
```

这个命令会读取 `migrations/` 目录下的 SQL 文件，并将更改应用到 `.env` 文件中 `DATABASE_URL` 指定的数据库。

### 7. 运行 API 服务 (Run the API Service)

你有两种主要方式运行 API 服务：

**方式 A: 直接运行 Go 代码 (推荐用于开发调试)**

```bash
make run
```

*   此命令使用 `go run` 直接编译和运行 `cmd/api/main.go`。
*   服务将在前台运行，日志会直接输出到终端。
*   你可以使用 `Ctrl+C` 来停止服务。
*   它会读取 `config.development.yaml` 和 `.env` 文件中的配置。

**方式 B: 在 Docker 容器中运行**

```bash
make docker-run
```

*   此命令首先会使用 `Dockerfile` 构建一个 Docker 镜像 (如果镜像不存在或代码有更新)。
*   然后，它会在后台启动一个名为 `language-player-api` 的 Docker 容器来运行 API。
*   容器会从 `.env` 文件加载环境变量进行配置。
*   使用 `make docker-stop` 来停止并移除这个容器。

### 8. 验证服务是否运行 (Verify Service)

服务启动后，你可以通过以下方式验证：

*   **健康检查**: 访问 `http://localhost:8080/healthz` (或你在配置中指定的端口)。如果看到 "OK"，表示服务基本运行正常。
*   **API 文档**: 访问 `http://localhost:8080/`，它应该会自动重定向到 `http://localhost:8080/swagger/index.html`，显示 Swagger UI。

### 9. 停止服务 (Stopping the Service)

*   **如果使用 `make run`**: 在运行命令的终端按 `Ctrl+C`。
*   **如果使用 `make docker-run`**: 运行 `make docker-stop`。
*   **停止依赖服务 (PostgreSQL, MinIO)**: 运行 `make deps-stop`。

现在，你可以按照这些步骤在本地成功运行后端 API 了。 