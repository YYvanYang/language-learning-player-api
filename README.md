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