**语言学习音频播放器 - 后端 API 架构设计方案 (完整版)**

**版本:** 1.3
**日期:** 2023-10-27 (最终修订)

**目录**

1.  [引言](#1-引言)
    *   [1.1 文档目的](#11-文档目的)
    *   [1.2 项目背景](#12-项目背景)
    *   [1.3 设计目标](#13-设计目标)
    *   [1.4 范围](#14-范围)
2.  [架构指导原则](#2-架构指导原则)
3.  [架构概览](#3-架构概览)
    *   [3.1 分层架构图](#31-分层架构图)
    *   [3.2 核心组件职责](#32-核心组件职责)
    *   [3.3 关于音频流式传输与 MinIO](#33-关于音频流式传输与-minio)
4.  [详细分层设计](#4-详细分层设计)
    *   [4.1 `cmd`](#41-cmd)
    *   [4.2 `internal/domain`](#42-internaldomain)
    *   [4.3 `internal/port`](#43-internalport)
    *   [4.4 `internal/usecase`](#44-internalusecase)
    *   [4.5 `internal/adapter`](#45-internaladapter)
    *   [4.6 `pkg`](#46-pkg)
5.  [关键技术决策](#5-关键技术决策)
6.  [核心业务流程示例](#6-核心业务流程示例)
    *   [6.1 用户获取音频详情（包含播放地址）](#61-用户获取音频详情包含播放地址)
    *   [6.2 用户记录播放进度](#62-用户记录播放进度)
    *   [6.3 用户创建书签](#63-用户创建书签)
    *   [6.4 用户使用 Google 账号认证](#64-用户使用-google-账号认证)
7.  [数据模型（详细）](#7-数据模型详细)
8.  [API 设计](#8-api-设计)
    *   [8.1 设计原则](#81-设计原则)
    *   [8.2 资源与端点示例](#82-资源与端点示例)
    *   [8.3 数据传输对象 (DTO)](#83-数据传输对象-dto)
    *   [8.4 认证与授权](#84-认证与授权)
    *   [8.5 分页、过滤与排序](#85-分页过滤与排序)
    *   [8.6 API 文档](#86-api-文档)
9.  [非功能性需求考量](#9-非功能性需求考量)
    *   [9.1 安全性](#91-安全性)
    *   [9.2 性能与可伸缩性](#92-性能与可伸缩性)
    *   [9.3 可观测性](#93-可观测性)
    *   [9.4 可测试性](#94-可测试性)
    *   [9.5 配置管理](#95-配置管理)
    *   [9.6 错误处理策略](#96-错误处理策略)
10. [代码规范与质量](#10-代码规范与质量)
11. [部署策略（概要）](#11-部署策略概要)
12. [未来考虑](#12-未来考虑)

---

## 1. 引言

### 1.1 文档目的

本文档旨在提供一份全面且详细的“语言学习音频播放器”后端 API 架构设计方案。它将作为开发团队构建、测试、部署、维护和未来演进后端系统的核心技术指南，确保架构的清晰性、一致性和可维护性。

### 1.2 项目背景

为满足语言学习者日益增长的移动化、个性化学习需求，本项目旨在打造一个功能完善、体验流畅的音频学习平台。后端 API 作为核心驱动，需支撑用户管理（包括本地及第三方登录）、内容分发（高效安全的音频访问）、学习活动记录（进度、书签）等关键功能。

### 1.3 设计目标

*   **高内聚、低耦合:** 实现清晰的模块化和分层，降低系统复杂度，提高代码可复用性。
*   **可维护性:** 代码结构清晰、逻辑分离，易于理解、修改和功能扩展。
*   **可测试性:** 组件设计易于进行单元、集成和端到端测试，保障代码质量。
*   **可扩展性:** 单体架构内部易于添加新功能领域，并为未来可能的微服务拆分奠定良好基础。
*   **健壮性:** 具备完善的错误处理机制和系统韧性，能优雅地处理异常情况。
*   **安全性:** 严格保护用户隐私、数据安全及内容版权，采用行业标准认证机制。
*   **性能:** 保证 API 响应速度，优化资源利用，支持流畅的音频访问体验。

### 1.4 范围

*   **包含:**
    *   后端 API 的整体架构设计（分层、模块划分）。
    *   核心业务逻辑的实现方式。
    *   用户认证机制（支持本地邮箱/密码及外部 Google OAuth 2.0）。
    *   数据持久化方案（使用 PostgreSQL）。
    *   音频文件存储与访问策略（使用 MinIO 和预签名 URL）。
    *   与外部服务（Google Auth, MinIO）的集成方式。
    *   API 接口定义原则与示例。
    *   关键的非功能性需求（安全、性能、可观测性、可测试性、配置管理、错误处理）的设计考量。
*   **不包含:**
    *   前端（Web/Mobile App）的具体 UI/UX 实现细节。
    *   基础设施的详细配置（如 Kubernetes YAML、网络配置）。
    *   CI/CD 流水线的具体脚本实现。

## 2. 架构指导原则

*   **关注点分离 (Separation of Concerns, SoC):** 严格划分各层职责，如 API 路由与请求解析、业务流程编排、核心领域逻辑、数据持久化、外部服务交互等。
*   **依赖倒置原则 (Dependency Inversion Principle, DIP):** 高层模块（业务策略）不依赖于低层模块（实现细节），两者均依赖于抽象（接口）。具体实现为 Usecase 依赖 Port 层定义的接口，而非具体的 Repository 或 Service 实现类。
*   **显式依赖 (Explicit Dependencies) 与依赖注入 (Dependency Injection, DI):** 所有组件的依赖项都应通过其构造函数显式传入，而不是在内部创建或从全局获取。这提高了代码的透明度、可测试性和可维护性。
*   **领域驱动设计 (Domain-Driven Design, DDD) 轻量应用:** 以 `internal/domain` 目录为核心，封装业务实体（Entities）、值对象（Value Objects）和核心业务规则，使其独立于具体的技术实现和基础设施。
*   **面向接口编程 (Programming to Interfaces):** 优先依赖接口而非具体类型，定义组件间的契约，从而降低耦合度，方便替换实现（如更换数据库、外部服务）和进行 Mock 测试。
*   **上下文传播 (`context.Context`):** 在处理请求的整个生命周期中，始终传递 `context.Context`。用于控制请求超时、传递取消信号，以及携带请求范围的值（如 Request ID, User ID, Trace ID）。

## 3. 架构概览

### 3.1 分层架构图

```mermaid
graph TD
    subgraph External ["外部交互 / 基础设施"]
        CLIENT([客户端<br/>Web/Mobile App]) -- HTTP/S --> API_GW{API Gateway<br/>(Optional)}
        API_GW -- HTTP/S --> HTTP_HANDLER

        DB[(数据库<br/>PostgreSQL)]
        MINIO[(对象存储<br/>MinIO / S3 Compatible)]
        GOOGLE_AUTH[Google OAuth 2.0<br/>Identity Platform]

        MONITORING[可观测性平台<br/>(Prometheus, Grafana, Jaeger/Tempo)]
        LOGGING[日志聚合系统<br/>(ELK Stack, Loki+Grafana)]
    end

    subgraph Application ["应用后端 (Go Monolith)"]
        subgraph Adapters ["适配器层 (internal/adapter)"]
            style Adapters fill:#cfc,stroke:#333,stroke-width:1px
            HTTP_HANDLER[HTTP Handler<br/>(handler/http)<br/>Gin / Chi]
            POSTGRES_REPO[PostgreSQL Repository<br/>(repository/postgres)<br/>pgx / sqlc / GORM]
            MINIO_SERVICE[MinIO Service Adapter<br/>(service/minio)<br/>minio-go SDK]
            GOOGLE_AUTH_ADAPTER[Google Auth Adapter<br/>(service/google_auth)<br/>Google API Client Lib]
        end

        subgraph Ports ["端口层 (internal/port)"]
            style Ports fill:#ccf,stroke:#333,stroke-width:1px
            REPO_INTERFACE{Repository<br/>Interfaces<br/>(UserRepository, AudioTrackRepository, ...)}
            STORAGE_INTERFACE{FileStorageService<br/>Interface}
            EXT_AUTH_INTERFACE{ExternalAuthService<br/>Interface}
        end

        subgraph ApplicationCore ["应用核心"]
            style ApplicationCore fill:#f9f,stroke:#333,stroke-width:2px
            USECASE[Use Cases / Application Logic<br/>(internal/usecase)]
            DOMAIN[Domain Model<br/>(Entities, Value Objects, Rules)<br/>(internal/domain)]
        end

        subgraph SharedKernel ["共享内核 / 工具库 (pkg)"]
           style SharedKernel fill:#eee,stroke:#666,stroke-width:1px
            LOGGER[Logger]
            ERRORS[Error Handling]
            VALIDATION[Validation Utils]
            SECURITY[Security Helpers<br/>(Hashing, JWT)]
            HTTPUTIL[HTTP Utils]
            PAGINATION[Pagination Helpers]
        end
    end

    %% Dependencies
    CLIENT -- HTTP Request (API Calls, Auth Redirects) --> HTTP_HANDLER
    HTTP_HANDLER -- Calls --> USECASE

    USECASE -- Uses --> DOMAIN
    USECASE -- Depends on --> REPO_INTERFACE
    USECASE -- Depends on --> STORAGE_INTERFACE
    USECASE -- Depends on --> EXT_AUTH_INTERFACE

    POSTGRES_REPO -- Implements --> REPO_INTERFACE
    POSTGRES_REPO -- Interacts --> DB

    MINIO_SERVICE -- Implements --> STORAGE_INTERFACE
    MINIO_SERVICE -- Interacts --> MINIO

    GOOGLE_AUTH_ADAPTER -- Implements --> EXT_AUTH_INTERFACE
    GOOGLE_AUTH_ADAPTER -- Interacts --> GOOGLE_AUTH

    %% Shared Kernel Usage
    HTTP_HANDLER -- Uses --> SharedKernel
    USECASE -- Uses --> SharedKernel
    POSTGRES_REPO -- Uses --> SharedKernel
    MINIO_SERVICE -- Uses --> SharedKernel
    GOOGLE_AUTH_ADAPTER -- Uses --> SharedKernel

    %% Observability & Logging
    HTTP_HANDLER -- Reports Metrics/Logs/Traces --> MONITORING & LOGGING
    USECASE -- Reports Metrics/Logs/Traces --> MONITORING & LOGGING
    POSTGRES_REPO -- Reports Metrics/Logs/Traces --> MONITORING & LOGGING
    MINIO_SERVICE -- Reports Metrics/Logs/Traces --> MONITORING & LOGGING
    GOOGLE_AUTH_ADAPTER -- Reports Metrics/Logs/Traces --> MONITORING & LOGGING

    HTTP_HANDLER -- Returns HTTP Response (Data, Errors, JWT) --> CLIENT
```

### 3.2 核心组件职责

*   **Domain (`internal/domain`):**
    *   定义业务的核心概念：实体（如 `User`, `AudioTrack`, `AudioCollection`, `Bookmark`）、值对象（如 `Language`, `AudioLevel`, `TimeRange`）和领域事件（可选）。
    *   封装不变的业务规则和逻辑，确保业务约束得到满足。
    *   是应用中最稳定、最纯粹的部分，与具体技术实现（数据库、API 框架）完全解耦。
*   **Port (`internal/port`):**
    *   定义应用核心与外部世界交互的边界和契约（Go 接口）。
    *   **输出端口 (Driven Ports):** 定义应用核心 *需要* 外部适配器实现的接口，用于访问基础设施。例如 `UserRepository`, `AudioTrackRepository`, `FileStorageService`, `ExternalAuthService`。
    *   **输入端口 (Driving Ports):** （可选）定义 Usecase 本身的接口，供输入适配器（如 HTTP Handler）调用。通常直接使用 Usecase 结构体及其方法即可。
*   **Usecase (`internal/usecase`):**
    *   实现具体的应用业务流程（用户故事或用例）。
    *   编排 Domain 实体和值对象，协调完成特定任务。
    *   调用 Port 层定义的输出接口来读取数据、持久化状态变更、与外部服务交互。
    *   处理事务边界、授权逻辑和应用层错误。
*   **Adapter (`internal/adapter`):**
    *   负责将应用核心与具体的技术和外部系统连接起来。
    *   **输入适配器 (Driving Adapters):** 如 `handler/http`，负责接收外部输入（如 HTTP 请求），进行协议转换、数据反序列化（到 DTO）、初步验证，然后调用相应的 Usecase。最后将 Usecase 的结果格式化为外部响应（如 JSON）。
    *   **输出适配器 (Driven Adapters):** 如 `repository/postgres`, `service/minio`, `service/google_auth`，负责实现 Port 层定义的输出接口，封装与具体数据库（SQL 语句、ORM）、对象存储（SDK 调用）、第三方认证服务（API 调用）的交互细节。
*   **Pkg (`pkg`):**
    *   存放项目内可共享的、与特定业务领域无关的通用工具或基础库代码。
    *   例如：结构化日志记录器配置、通用错误处理工具、请求验证助手、安全相关函数（密码哈希、JWT 操作）、HTTP 响应封装、分页计算等。

### 3.3 关于音频流式传输与 MinIO

*   **MinIO/S3 对流式传输的支持:** MinIO（及 S3 兼容存储）本身通过支持 **HTTP Range Requests** (请求头 `Range: bytes=start-end`) 来实现高效的流式传输。客户端（如浏览器 `<audio>` 标签、Web Audio API、移动端播放器库）可以请求文件的任意部分，从而实现“边下边播”的效果，以及断点续传。
*   **后端 API 的角色:** 在本架构设计中，后端 API **不直接代理**音频数据流。这样做会消耗 API 服务器大量的带宽和计算资源，违背了将大文件处理卸载到专用存储服务的初衷。
*   **推荐方案：预签名 URL (Presigned URL):**
    1.  客户端（前端）向后端 API 请求获取某个音频轨道的详细信息。
    2.  后端 API 在验证用户权限后，除了返回音频的元数据（标题、描述等），还会实时调用 MinIO SDK，生成一个指向 MinIO 中该音频文件的 **预签名 URL**。
    3.  这个 URL 包含了临时的访问凭证（签名），并有**较短的有效期**（例如 15 分钟到 1 小时）。
    4.  后端 API 将此预签名 URL 返回给客户端。
    5.  客户端直接使用这个预签名 URL 向 MinIO（或配置的 CDN 地址）发起 GET 请求来获取音频数据。客户端的播放器库会自动利用 Range Requests 与 MinIO/CDN 进行流式交互。
*   **优势:**
    *   **负载卸载:** 将大文件传输的带宽和 CPU 压力从后端 API 转移到专用的、高可扩展的对象存储或 CDN。
    *   **安全性:** 音频文件在 MinIO 中可以设置为私有，只有持有有效预签名 URL 的用户才能在有效期内访问。
    *   **性能:** 用户直接从最近的 CDN 节点或 MinIO 获取数据，延迟更低，速度更快。

## 4. 详细分层设计

### 4.1 `cmd`

*   **目录结构:** `cmd/api/main.go`
*   **职责:** 应用程序的启动入口，负责初始化和组装 (Wiring) 所有组件。
*   **核心步骤:**
    1.  **配置加载:** 使用 `internal/config` 包和 `spf13/viper` 加载配置（从文件、环境变量）。配置项包括数据库连接信息、MinIO 服务器地址和凭证、JWT 密钥、Google OAuth Client ID/Secret、服务监听端口、日志级别等。
    2.  **日志初始化:** 配置并初始化结构化日志记录器 (如 `slog`, `zap`)，设定输出格式和级别。
    3.  **基础设施连接初始化:**
        *   创建 PostgreSQL 数据库连接池 (如 `pgxpool.Pool`)，并执行 Ping 测试连通性。
        *   创建 MinIO 客户端 (`minio.New`)，传入 Endpoint, AccessKey, SecretKey, UseSSL 等配置。
        *   创建用于调用 Google API 的 `http.Client` (可能配置了特定的 Transport 或 Timeout)。
    4.  **依赖注入 (Dependency Injection):** 按照依赖关系，逐层实例化组件并将依赖项传入构造函数。
        *   **实例化 Repository (Adapter):** `userRepo := postgres.NewUserRepository(dbPool)` ...
        *   **实例化 Service (Adapter):** `storageService := minioadapter.NewMinioStorageService(minioClient, config.MinioBucket)`，`googleAuthSvc := googleauthadapter.NewGoogleAuthService(googleHttpClient, config.GoogleClientID, config.GoogleClientSecret)`。
        *   **实例化 Usecase:** `authUseCase := usecase.NewAuthUseCase(userRepo, securityHelper, googleAuthSvc)`，`audioContentUseCase := usecase.NewAudioContentUseCase(trackRepo, collectionRepo, storageService)` ...
        *   **实例化 Handler (Adapter):** `authHandler := httpadapter.NewAuthHandler(authUseCase, validator)`，`audioHandler := httpadapter.NewAudioHandler(audioContentUseCase, validator)` ...
    5.  **HTTP Router 与中间件设置:**
        *   初始化 HTTP Router (如 `chi.NewRouter()` 或 `gin.Default()`)。
        *   **注册全局中间件 (按顺序):**
            *   `Recovery`: 捕获 panic 并返回 500 错误，记录堆栈。
            *   `RequestID`: 生成唯一请求 ID 并注入 Context 和响应头。
            *   `Logger`: 记录每个请求的入口和出口信息（方法、路径、状态码、耗时、RequestID 等）。
            *   `CORS`: 配置跨域资源共享策略。
            *   `Timeout`: （可选）设置请求处理的全局超时。
            *   `Auth Middleware`: 验证 `Authorization: Bearer` 头中的 JWT，解析 `userID` 并注入 Context (对于需要认证的路由组)。
        *   **注册路由:** 将 Handler 的方法绑定到具体的 HTTP 路径和方法上，可按资源分组。
    6.  **启动 HTTP Server:** 创建 `http.Server` 实例，配置监听地址、端口、Read/Write Timeout，并使用设置好的 Router 作为 Handler 启动服务 (`server.ListenAndServe()`)。
    7.  **实现优雅停机 (Graceful Shutdown):**
        *   监听操作系统信号 `syscall.SIGINT` 和 `syscall.SIGTERM`。
        *   收到信号后，调用 `server.Shutdown(ctx)`，给正在处理的请求一段宽限时间完成。
        *   在 Shutdown 过程中或之后，关闭数据库连接池、MinIO 客户端等资源。

### 4.2 `internal/domain`

*   **目录结构:** `internal/domain/user.go`, `audiotrack.go`, `audiocollection.go`, `bookmark.go`, `playbackprogress.go`, `value_objects.go`, `errors.go`...
*   **核心职责:** 封装业务核心概念、状态和规则，保持纯净，无基础设施依赖。
*   **实体 (Entities):** （具有唯一标识符和生命周期）
    *   `User`: `ID` (UUID), `Email` (Value Object?), `Name` (string), `HashedPassword` (*string, 可为空), `GoogleID` (*string, 可为空, unique), `AuthProvider` (string, 'local'/'google'), `ProfileImageURL` (*string), `CreatedAt`, `UpdatedAt`. 方法如 `ValidatePassword(plain string) bool`, `UpdateProfile(...) error`。
    *   `AudioTrack`: `ID` (UUID), `Title` (string), `Description` (string), `Language` (Value Object `Language`), `Level` (Value Object `AudioLevel`), `Duration` (`time.Duration`), `MinioBucket` (string), `MinioObjectKey` (string), `CoverImageURL` (*string), `UploaderID` (*UUID), `IsPublic` (bool), `Tags` ([]string), `CreatedAt`, `UpdatedAt`.
    *   `AudioCollection`: `ID` (UUID), `Title` (string), `Description` (string), `OwnerID` (UUID), `Type` (Value Object `CollectionType`), `TrackIDs` ([]UUID, 有序), `CreatedAt`, `UpdatedAt`. 方法如 `AddTrack(trackID UUID, position int) error`, `ReorderTracks(orderedIDs []UUID) error`.
    *   `Bookmark`: `ID` (UUID), `UserID` (UUID), `TrackID` (UUID), `Timestamp` (`time.Duration`), `Note` (string), `CreatedAt`.
    *   `PlaybackProgress`: `UserID` (UUID), `TrackID` (UUID), (作为复合主键), `Progress` (`time.Duration`), `LastListenedAt` (`time.Time`).
*   **值对象 (Value Objects):** （不可变，由属性定义身份）
    *   `Language`: `Code` (string, e.g., "en-US"), `Name` (string, e.g., "English (US)").
    *   `AudioLevel`: (e.g., "A1", "B2", "C1").
    *   `CollectionType`: (e.g., "COURSE", "PLAYLIST").
    *   `Email`: (包含验证逻辑)。
*   **领域错误 (`errors.go`):**
    *   定义标准业务错误常量: `var ErrNotFound = errors.New("entity not found")`, `ErrInvalidArgument`, `ErrPermissionDenied`, `ErrConflict`, `ErrAuthenticationFailed`.
    *   提供错误检查辅助函数或类型。

### 4.3 `internal/port`

*   **目录结构:** `internal/port/repository.go`, `service.go`
*   **职责:** 定义应用核心与外部适配器交互的接口契约。
*   **输出端口 - 仓库接口 (`repository.go`):**
    ```go
    package port

    import (
        "context"
        "your_project/internal/domain"
        "your_project/pkg/pagination" // Assuming a common pagination package
    )

    type UserRepository interface {
        FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
        FindByEmail(ctx context.Context, email string) (*domain.User, error)
        FindByProviderID(ctx context.Context, provider string, providerUserID string) (*domain.User, error) // New
        Create(ctx context.Context, user *domain.User) error
        Update(ctx context.Context, user *domain.User) error
        LinkProviderID(ctx context.Context, userID domain.UserID, provider string, providerUserID string) error // Optional, New
    }

    type AudioTrackRepository interface {
        FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error)
        ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error) // For collections
        List(ctx context.Context, params ListTracksParams, page pagination.Page) ([]*domain.AudioTrack, int, error) // Returns tracks and total count
        Create(ctx context.Context, track *domain.AudioTrack) error
        Update(ctx context.Context, track *domain.AudioTrack) error
        Delete(ctx context.Context, id domain.TrackID) error
        Exists(ctx context.Context, id domain.TrackID) (bool, error) // Helper
    }
    // Define ListTracksParams struct with filter options (language, level, query, etc.)

    type AudioCollectionRepository interface {
        FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
        ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) ([]*domain.AudioCollection, int, error)
        Create(ctx context.Context, collection *domain.AudioCollection) error
        Update(ctx context.Context, collection *domain.AudioCollection) error // Update metadata
        UpdateTracks(ctx context.Context, id domain.CollectionID, orderedTrackIDs []domain.TrackID) error // Update track list/order
        Delete(ctx context.Context, id domain.CollectionID, ownerID domain.UserID) error // Ensure owner deletes
    }

    type PlaybackProgressRepository interface {
        Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
        Upsert(ctx context.Context, progress *domain.PlaybackProgress) error
        ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.PlaybackProgress, int, error)
    }

    type BookmarkRepository interface {
        FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error)
        ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error)
        ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error)
        Create(ctx context.Context, bookmark *domain.Bookmark) error
        Delete(ctx context.Context, id domain.BookmarkID, userID domain.UserID) error // Ensure owner deletes
    }

    type TranscriptionRepository interface { // If storing transcriptions
        FindByTrackID(ctx context.Context, trackID domain.TrackID) (*domain.Transcription, error) // Assuming Transcription domain object
        Upsert(ctx context.Context, transcription *domain.Transcription) error
    }

    // TransactionManager (Optional, for complex use cases needing transactions)
    type TransactionManager interface {
        Begin(ctx context.Context) (context.Context, error) // Returns context with Tx
        Commit(ctx context.Context) error
        Rollback(ctx context.Context) error
    }
    ```
*   **输出端口 - 服务接口 (`service.go`):**
    ```go
    package port

    import (
        "context"
        "time"
    )

    // FileStorageService defines the contract for interacting with object storage
    type FileStorageService interface {
        // Returns a temporary, signed URL for reading a private object
        GetPresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
        // Returns a temporary, signed URL for uploading/overwriting an object
        GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
        // Deletes an object from storage
        DeleteObject(ctx context.Context, bucket, objectKey string) error
    }

    // ExternalUserInfo contains standardized user info from an external provider
    type ExternalUserInfo struct {
        Provider      string // e.g., "google"
        ProviderUserID string // Unique ID from the provider (e.g., Google subject ID)
        Email         string // Email provided (should be verified by provider)
        IsEmailVerified bool   // Whether the provider claims the email is verified
        Name          string
        PictureURL    string // Optional profile picture URL
    }

    // ExternalAuthService defines the contract for verifying external auth credentials
    type ExternalAuthService interface {
        // Verifies a Google ID token and returns standardized user info.
        VerifyGoogleToken(ctx context.Context, idToken string) (*ExternalUserInfo, error)
        // Add methods for other providers if needed (e.g., VerifyFacebookToken, VerifyAppleToken)
    }

    // SecurityHelper defines crypto operations needed by use cases
    type SecurityHelper interface {
         HashPassword(ctx context.Context, password string) (string, error)
         CheckPasswordHash(ctx context.Context, password, hash string) bool
         GenerateJWT(ctx context.Context, userID domain.UserID) (string, error)
    }
    ```

### 4.4 `internal/usecase`

*   **目录结构:** `internal/usecase/auth_uc.go`, `audio_content_uc.go`, `user_activity_uc.go`...
*   **职责:** 实现应用的核心业务流程，协调领域对象和基础设施（通过 Port 接口）。
*   **结构:** 通常为每个主要业务领域或聚合根创建一个 Usecase 结构体，包含其依赖的 Port 接口。
    ```go
    package usecase

    import (
        "context"
        "errors"
        "time"
        "your_project/internal/domain"
        "your_project/internal/port"
        "your_project/pkg/pagination"
        // ... other imports
    )

    type AuthUseCase struct {
        userRepo       port.UserRepository
        secHelper      port.SecurityHelper
        extAuthService port.ExternalAuthService
        logger         port.Logger // Assuming a logger interface
    }

    func NewAuthUseCase(ur port.UserRepository, sh port.SecurityHelper, eas port.ExternalAuthService, log port.Logger) *AuthUseCase {
        return &AuthUseCase{userRepo: ur, secHelper: sh, extAuthService: eas, logger: log}
    }

    // --- AuthUseCase Methods ---

    func (uc *AuthUseCase) RegisterWithPassword(ctx context.Context, email, password, name string) (*domain.User, string, error) {
        // 1. Validate input (e.g., password complexity - maybe in domain or here)
        if _, err := uc.userRepo.FindByEmail(ctx, email); err == nil {
            return nil, "", domain.ErrConflict // Email already exists
        } else if !errors.Is(err, domain.ErrNotFound) {
            return nil, "", err // Other repo error
        }

        // 2. Hash password
        hashedPassword, err := uc.secHelper.HashPassword(ctx, password)
        if err != nil {
            return nil, "", err
        }

        // 3. Create domain user
        user := domain.NewLocalUser(email, hashedPassword, name) // Domain constructor

        // 4. Save user
        if err := uc.userRepo.Create(ctx, user); err != nil {
            return nil, "", err
        }

        // 5. Generate JWT
        token, err := uc.secHelper.GenerateJWT(ctx, user.ID)
        if err != nil {
            // Log error, but maybe still return user? Or handle differently.
            return user, "", err // Or just log and return ("", err)
        }

        uc.logger.Info(ctx, "User registered successfully", "userID", user.ID, "email", email)
        return user, token, nil
    }

    func (uc *AuthUseCase) LoginWithPassword(ctx context.Context, email, password string) (string, error) {
        user, err := uc.userRepo.FindByEmail(ctx, email)
        if err != nil {
            if errors.Is(err, domain.ErrNotFound) {
                return "", domain.ErrAuthenticationFailed // Use generic auth failed error
            }
            return "", err
        }

        if user.AuthProvider != "local" || user.HashedPassword == nil {
            return "", domain.ErrAuthenticationFailed // Account exists but not via password
        }

        if !uc.secHelper.CheckPasswordHash(ctx, password, *user.HashedPassword) {
            return "", domain.ErrAuthenticationFailed
        }

        token, err := uc.secHelper.GenerateJWT(ctx, user.ID)
        if err != nil {
            return "", err
        }
        uc.logger.Info(ctx, "User logged in successfully", "userID", user.ID)
        return token, nil
    }

    func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error) {
        extInfo, err := uc.extAuthService.VerifyGoogleToken(ctx, googleIdToken)
        if err != nil {
            uc.logger.Warn(ctx, "Google token verification failed", "error", err)
            return "", false, domain.ErrAuthenticationFailed // External validation failed
        }

        // Try finding user by Google ID first
        user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
        if err == nil {
            // User found via Google ID - Login success
            token, err := uc.secHelper.GenerateJWT(ctx, user.ID)
            if err != nil { return "", false, err }
             uc.logger.Info(ctx, "User authenticated via existing Google ID", "userID", user.ID)
            return token, false, nil
        } else if !errors.Is(err, domain.ErrNotFound) {
            return "", false, err // Unexpected repository error
        }

        // User not found by Google ID, try finding by email
        user, err = uc.userRepo.FindByEmail(ctx, extInfo.Email)
        if err == nil {
            // User found by email, but Google ID is not linked
            // **Strategy C: Conflict - Email already exists with different provider/method**
            uc.logger.Warn(ctx, "Google auth attempt failed: Email exists with different provider", "email", extInfo.Email, "existingProvider", user.AuthProvider)
            return "", false, domain.ErrConflict // Inform client email is taken

            // **Alternative Strategy A: Auto-link (Use with caution!)**
            // if user.AuthProvider == "local" && user.GoogleID == nil { // Only link if local and not already linked
            //     err = uc.userRepo.LinkProviderID(ctx, user.ID, extInfo.Provider, extInfo.ProviderUserID)
            //     if err != nil { return "", false, err }
            //     token, err := uc.secHelper.GenerateJWT(ctx, user.ID)
            //     if err != nil { return "", false, err }
            //     uc.logger.Info(ctx, "Existing local user linked to Google ID", "userID", user.ID)
            //     return token, false, nil
            // } else {
            //     // User exists with email but already linked to Google or another provider - Conflict
            //     return "", false, domain.ErrConflict
            // }

        } else if !errors.Is(err, domain.ErrNotFound) {
            return "", false, err // Unexpected repository error
        }

        // User not found by Google ID or Email - Create new user from Google info
        newUser := domain.NewGoogleUser(
            extInfo.Email,
            extInfo.Name,
            extInfo.PictureURL,
            extInfo.ProviderUserID,
        ) // Domain constructor for external user

        if err := uc.userRepo.Create(ctx, newUser); err != nil {
            return "", false, err
        }

        token, err := uc.secHelper.GenerateJWT(ctx, newUser.ID)
        if err != nil {
            // Hard to recover here, maybe delete the created user? Log critical error.
            uc.logger.Error(ctx, "Failed to generate JWT for newly created Google user", "error", err, "userID", newUser.ID)
            return "", true, err
        }

        uc.logger.Info(ctx, "New user created via Google authentication", "userID", newUser.ID)
        return token, true, nil
    }


    // --- AudioContentUseCase ---
    type AudioContentUseCase struct {
        trackRepo      port.AudioTrackRepository
        collectionRepo port.AudioCollectionRepository
        storageService port.FileStorageService
        // ... other dependencies like logger
    }
    // Methods like GetAudioTrackDetails, ListTracks, etc.
    func (uc *AudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*AudioTrackDetailsDTO, error) { // Assuming DTO defined elsewhere
        track, err := uc.trackRepo.FindByID(ctx, trackID)
        if err != nil {
            return nil, err // Handles ErrNotFound from repo
        }

        // Potentially check authorization if track is not public
        // if !track.IsPublic { /* check if user in ctx has access */ }

        presignedURL, err := uc.storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, 1*time.Hour) // Configurable expiry
        if err != nil {
            // Log error but maybe don't fail the whole request? Or return DTO without URL? Decision needed.
             uc.logger.Error(ctx, "Failed to get presigned URL", "error", err, "trackID", trackID)
             // return nil, err // Or handle gracefully
             presignedURL = "" // Example: return empty URL on failure
        }

         // Fetch other related data if needed (e.g., user progress, bookmarks for this track)
         // progress, _ := progressRepo.Find(ctx, userIDFromCtx, trackID)

        // Map domain.AudioTrack (+ presignedURL, progress, etc.) to AudioTrackDetailsDTO
        dto := MapTrackToDetailsDTO(track, presignedURL /*, progress, bookmarks */)
        return dto, nil
    }


    // --- UserActivityUseCase ---
    type UserActivityUseCase struct {
        progressRepo port.PlaybackProgressRepository
        bookmarkRepo port.BookmarkRepository
        trackRepo    port.AudioTrackRepository // To validate track existence
        // ... logger
    }
    // Methods like RecordPlaybackProgress, CreateBookmark, ListBookmarks etc.
     func (uc *UserActivityUseCase) RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error {
         // 1. Validate track exists (optional but good practice)
         exists, err := uc.trackRepo.Exists(ctx, trackID)
         if err != nil { return err }
         if !exists { return domain.ErrNotFound }

         // 2. Create or update domain object
         prog := &domain.PlaybackProgress{
             UserID:         userID,
             TrackID:        trackID,
             Progress:       progress,
             LastListenedAt: time.Now(),
         }

         // 3. Call repo Upsert
         return uc.progressRepo.Upsert(ctx, prog)
     }
    ```
*   **DTOs:** Usecase 方法的输入和输出通常使用简单的结构体 (DTOs - Data Transfer Objects) 定义，与 Handler 层的 API DTO 保持一致或进行简单映射，避免直接暴露 Domain 实体。

### 4.5 `internal/adapter`

*   **`handler/http`:** (`internal/adapter/handler/http/`)
    *   **`router.go`:** 设置路由、分组、应用中间件。
    *   **`middleware/`:** `auth.go` (JWT验证), `logger.go`, `recovery.go` 等。
    *   **`auth_handler.go`, `audio_handler.go`, ...:** 实现具体的 Handler 方法。
        *   职责：解析 HTTP 请求（路径参数、查询参数、请求体 JSON），绑定到 Request DTO。
        *   调用 `pkg/validation` 验证 Request DTO，不通过则返回 400 错误。
        *   从 `context.Context` 获取 `userID`（由 Auth 中间件注入）。
        *   调用对应的 Usecase 方法，传入 `ctx` 和输入 DTO（或转换后的参数）。
        *   处理 Usecase 返回的结果（DTO 或 error）。
        *   使用 `pkg/httputil` 将结果或错误封装成统一格式的 JSON 响应，并设置正确的 HTTP 状态码。
    *   **`dto/`:** 定义 API 请求体和响应体的结构体，使用 `json` tags 进行序列化/反序列化，使用 `validate` tags 进行输入验证。
*   **`repository/postgres`:** (`internal/adapter/repository/postgres/`)
    *   **`db.go`:** 初始化 `pgxpool.Pool`，包含连接池配置。
    *   **`user_repo.go`, `audiotrack_repo.go`, ...:** 实现 `internal/port` 中定义的 `Repository` 接口。
        *   **技术选型:**
            *   **`pgx` (推荐):** 直接编写 SQL 语句，性能好，控制力强。使用 `pool.QueryRow(...).Scan(...)`, `pool.Query(...)`。需要手动处理 `pgx.ErrNoRows` 到 `domain.ErrNotFound` 的转换。
            *   **`sqlc`:** 编写 `.sql` 文件定义查询，`sqlc` 生成类型安全的 Go 代码。Repository 实现调用生成的代码，简洁高效。
            *   **`GORM`:** ORM，开发速度快，但可能隐藏 SQL 复杂性，需要注意性能和 N+1 问题。如果使用，建议定义单独的 GORM 模型，与 Domain 实体进行映射。
    *   **`tx_manager.go` (可选):** 实现 `port.TransactionManager` 接口，封装 `pgx` 的事务处理 (`pool.Begin`, `tx.Commit`, `tx.Rollback`)。
*   **`service/minio`:** (`internal/adapter/service/minio/`)
    *   **`minio_adapter.go`:** 实现 `port.FileStorageService` 接口。
    *   使用官方 `minio-go/v7` SDK。
    *   `New(...)`: 接收 MinIO Client 和配置 (bucket name) 作为依赖。
    *   实现 `GetPresignedGetURL`, `GetPresignedPutURL`, `DeleteObject` 方法，调用 MinIO SDK 的相应函数。处理 SDK 可能返回的错误。
*   **`service/google_auth`:** (`internal/adapter/service/google_auth/`)
    *   **`google_adapter.go`:** 实现 `port.ExternalAuthService` 接口。
    *   使用 Google 官方 Go 库 (`google.golang.org/api/idtoken` 或 `golang.org/x/oauth2`)。
    *   `New(...)`: 接收 `http.Client`, Google Client ID 等配置。
    *   `VerifyGoogleToken(...)`: 实现调用 Google API 验证 ID Token 的逻辑，包括检查 audience, issuer, expiry，并解析用户信息，最后返回 `port.ExternalUserInfo` 或错误。

### 4.6 `pkg`

*   **`logger`:** 提供配置好的、全局可用的 `slog.Logger` 实例或接口。可能包含从 Context 提取/注入 RequestID 的辅助函数。
*   **`errors`:** 通用错误处理工具，如错误包装 `Wrap(err, msg)`, `Wrapf(...)`，以及错误类型检查 `IsNotFound(err)` 等。
*   **`validation`:** 基于 `go-playground/validator` 的验证器实例和注册自定义验证规则的函数。提供 `ValidateStruct(s interface{}) error` 帮助函数。
*   **`httputil`:** `RespondJSON(w, status, payload)`, `RespondError(w, apiError)` 等统一 API 响应的帮助函数。定义标准错误响应体结构 `APIError { Code string `json:"code"`, Message string `json:"message"` }`。
*   **`security`:** 实现 `port.SecurityHelper` 接口。包含 `BcryptHasher` (密码哈希) 和 `JWTHelper` (JWT 生成与验证，配置密钥和有效期)。
*   **`pagination`:** 定义分页请求参数结构体 `Page { Limit int, Offset int }` 和响应结构体 `PaginatedResponse { Data []interface{}, Total int, Limit int, Offset int }`，以及计算 Offset 的帮助函数。

## 5. 关键技术决策

| 领域                 | 技术选型                                                  | 理由                                                                      |
| :------------------- | :-------------------------------------------------------- | :------------------------------------------------------------------------ |
| **语言**             | Go (最新稳定版, e.g., 1.21+)                              | 高性能，并发模型优秀 (Goroutines)，静态类型安全，部署简单，生态系统完善。      |
| **Web 框架/路由**   | `chi`                                                     | 轻量级，符合 Go 习惯 (net/http compatible)，中间件组合灵活，性能良好。       |
| *(备选)*             | `gin-gonic`                                               | 功能更全，更 opinionated，学习曲线稍陡，性能也很好。                       |
| **数据库**           | PostgreSQL (最新稳定版)                                   | 功能强大 (JSONB, GIS)， ACID 兼容，稳定可靠，社区活跃，扩展性好。            |
| **数据库访问 (首选)** | `pgx` (原生驱动) + `sqlc` (代码生成)                      | `pgx` 性能最佳，`sqlc` 从 SQL 生成类型安全的 Go 代码，减少模板代码，提高安全性。 |
| *(备选)*             | `GORM` (ORM)                                              | 开发效率高，适用于简单 CRUD，但可能隐藏 SQL 复杂性，需谨慎用于复杂查询。        |
| **对象存储**         | MinIO                                                     | S3 兼容，开源，可自托管，易于本地开发和测试，成本可控。                      |
| **MinIO SDK**        | `minio-go/v7`                                             | 官方 Go SDK，功能完整，维护良好。                                         |
| **配置管理**         | `spf13/viper`                                             | 支持多种格式 (YAML, JSON, Env, Flags)，优先级管理，功能强大灵活。           |
| **日志记录**         | `slog` (Go 1.21+ 标准库)                                  | 标准库支持，结构化日志，性能好，易于集成。                                  |
| *(备选)*             | `uber-go/zap` / `rs/zerolog`                              | 性能极高，功能更丰富，社区广泛使用。                                      |
| **内部认证**         | JWT (JSON Web Tokens)                                     | 无状态 API 的事实标准，实现相对简单，适用于分布式系统。                      |
| **外部认证协议**     | OAuth 2.0 / OpenID Connect (OIDC)                         | 行业标准，安全可靠，主流 IdP (Google, Facebook, Apple) 均支持。            |
| **Google 认证库**    | `google.golang.org/api/idtoken` 或 `golang.org/x/oauth2`    | Google 官方/推荐库，封装了与 Google API 交互的细节。                     |
| **密码哈希**         | `golang.org/x/crypto/bcrypt`                              | 行业标准，加盐，自适应强度，安全性高。                                    |
| **数据验证**         | `go-playground/validator`                                   | 功能强大，支持结构体验证和自定义规则，与 Web 框架集成良好。                |
| **数据库迁移**       | `golang-migrate` 或 `goose`                                 | 管理数据库 Schema 版本，确保开发、测试、生产环境数据库结构一致性。          |
| **依赖注入 (DI)**    | 手动注入 (Manual DI in `main.go`)                         | 对于中小型项目足够清晰简单，易于理解。                                    |
| *(备选, 复杂项目)*   | `google/wire` (编译时 DI)                                  | 编译时进行依赖检查和代码生成，类型安全，无运行时反射开销。                  |
| **测试 Mock**        | `stretchr/testify/mock` 或 `golang/mock/gomock`           | 主流 Mock 框架，用于在单元测试中模拟接口依赖。                            |
| **集成测试容器**     | `ory/dockertest`                                          | 方便在测试中启动临时的 Docker 容器 (PostgreSQL, MinIO) 进行集成测试。        |

## 6. 核心业务流程示例

### 6.1 用户获取音频详情（包含播放地址）

1.  **Client:** 发送 `GET /api/v1/audio/tracks/{trackId}` 请求，在 `Authorization` Header 中携带有效的应用 JWT。
2.  **Middleware (Auth):** 验证 JWT，解析出 `userID`，将其注入 `context.Context`。
3.  **HTTP Handler (`AudioHandler.GetDetails`):**
    *   从 URL 路径参数中解析 `trackId`。
    *   调用 `audioContentUseCase.GetAudioTrackDetails(ctx, trackId)`。
    *   接收 Usecase 返回的 `AudioTrackDetailsDTO` 或 `error`。
    *   若成功，使用 `httputil.RespondJSON(w, http.StatusOK, dto)` 返回 200 和 DTO 数据。
    *   若失败（如 `domain.ErrNotFound`），使用 `httputil.RespondError(w, apiError)` 返回相应的 4xx/5xx 错误。
4.  **Usecase (`AudioContentUseCase.GetAudioTrackDetails`):**
    *   调用 `trackRepo.FindByID(ctx, trackId)` 获取 `domain.AudioTrack` 实体。若返回 `ErrNotFound`，则向上层返回该错误。
    *   (可选权限检查：如果 `track.IsPublic` 为 false，检查 `ctx` 中的 `userID` 是否有权访问)。
    *   调用 `storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, presignedUrlExpiry)` 获取临时的 MinIO 播放 URL。
    *   若获取 URL 失败，记录错误日志，可以选择向上层返回错误，或在 DTO 中将播放 URL 设为空字符串或特定错误标记。
    *   (可选) 调用其他 Repository 获取关联信息，如用户对此 Track 的播放进度、书签列表。
    *   将 `AudioTrack` 实体、生成的 `presignedURL` 以及其他获取到的信息，映射到一个 `AudioTrackDetailsDTO` 结构体中。
    *   返回 `dto` 和 `nil` 错误。
5.  **Adapter (`MinioService.GetPresignedGetURL`):**
    *   使用 `minio-go` SDK 的 `PresignedGetObject` 方法，传入 bucket, object key, expiry duration 和请求方法 ("GET")。
    *   返回生成的 URL 字符串或 MinIO SDK 返回的错误。
6.  **Adapter (`PostgresAudioTrackRepository.FindByID`):**
    *   执行 SQL 查询 `SELECT ... FROM audio_tracks WHERE id = $1`。
    *   将查询结果扫描 (Scan) 到 `domain.AudioTrack` 结构体中。
    *   如果查询无结果 (`pgx.ErrNoRows`)，返回 `domain.ErrNotFound`。否则返回实体或 SQL 执行错误。
7.  **Client:** 收到包含音频元数据和 `playUrl` (预签名 URL) 的 JSON 响应。客户端的音频播放器使用此 `playUrl` 直接向 MinIO/CDN 发起请求，利用 HTTP Range Requests 实现流式播放。

### 6.2 用户记录播放进度

1.  **Client:** 发送 `POST /api/v1/users/me/progress` 请求，携带 JWT，请求体为 `{"trackId": "uuid-of-track", "progressSeconds": 123}`。
2.  **Middleware (Auth):** 验证 JWT，注入 `userID` 到 Context。
3.  **HTTP Handler (`ProgressHandler.RecordProgress`):**
    *   绑定 JSON 请求体到 `RecordProgressRequestDTO`。
    *   使用 `pkg/validation` 验证 DTO (e.g., `trackId` 必填, `progressSeconds >= 0`)。若失败返回 400 Bad Request。
    *   从 Context 获取 `userID`。
    *   将 `dto.ProgressSeconds` 转换为 `time.Duration`。
    *   调用 `userActivityUseCase.RecordPlaybackProgress(ctx, userID, dto.TrackID, progressDuration)`。
    *   若 Usecase 返回错误 (如 `domain.ErrNotFound` for track)，返回相应错误码 (404)。
    *   若成功，返回 204 No Content 或 200 OK。
4.  **Usecase (`UserActivityUseCase.RecordPlaybackProgress`):**
    *   (推荐) 调用 `trackRepo.Exists(ctx, trackID)` 验证音频轨道是否存在，不存在则返回 `domain.ErrNotFound`。
    *   创建 `domain.PlaybackProgress` 实体实例，填充 `UserID`, `TrackID`, `Progress`, `LastListenedAt`。
    *   调用 `progressRepo.Upsert(ctx, progress)` 将数据写入数据库（插入或更新）。
    *   处理并返回 Repository 可能出现的错误。
5.  **Adapter (`PostgresPlaybackProgressRepository.Upsert`):**
    *   执行 SQL `INSERT INTO playback_progress (...) VALUES (...) ON CONFLICT (user_id, track_id) DO UPDATE SET progress_seconds = EXCLUDED.progress_seconds, last_listened_at = EXCLUDED.last_listened_at`。
    *   返回 SQL 执行的错误（或 `nil`）。

### 6.3 用户创建书签

1.  **Client:** 发送 `POST /api/v1/bookmarks` 请求，携带 JWT，请求体 `{"trackId": "uuid-track", "timestampSeconds": 65, "note": "Important point"}`。
2.  **Middleware (Auth):** 验证 JWT，注入 `userID`。
3.  **HTTP Handler (`BookmarkHandler.CreateBookmark`):**
    *   绑定请求体到 `CreateBookmarkRequestDTO`。
    *   验证 DTO (`trackId` 必填, `timestampSeconds >= 0`)。
    *   从 Context 获取 `userID`。
    *   调用 `userActivityUseCase.CreateBookmark(ctx, userID, dto.TrackID, time.Duration(dto.TimestampSeconds)*time.Second, dto.Note)`。
    *   若成功，Usecase 返回创建的 `domain.Bookmark` 实体。Handler 将其映射为 `BookmarkResponseDTO`，返回 201 Created 和 DTO。
    *   若失败 (e.g., Track 不存在)，返回相应错误码。
4.  **Usecase (`UserActivityUseCase.CreateBookmark`):**
    *   (推荐) 验证 `trackID` 是否存在。
    *   创建 `domain.Bookmark` 实体实例。
    *   调用 `bookmarkRepo.Create(ctx, bookmark)`。
    *   若 Repository 返回错误，则向上层返回。
    *   若成功，返回创建的 `Bookmark` 实体（已包含由数据库生成的 ID 和 `CreatedAt`）。
5.  **Adapter (`PostgresBookmarkRepository.Create`):**
    *   执行 SQL `INSERT INTO bookmarks (...) VALUES (...) RETURNING id, created_at`。
    *   将返回的 `id` 和 `created_at` 扫描回传入的 `bookmark` 实体指针。
    *   返回 SQL 执行错误（或 `nil`）。

### 6.4 用户使用 Google 账号认证

1.  **Client (Pre-step):** 用户在前端点击“使用 Google 登录”。前端启动 Google OAuth 2.0 流程（通常使用 Google 的 JS 库或平台 SDK）。
2.  **Google Auth Flow:** 用户被重定向到 Google 登录和授权页面。用户同意授权后，Google 将用户重定向回在 Google Cloud Console 中配置的 **前端应用的回调 URI**，并附带凭证（推荐使用 **ID Token**，因为它直接包含了用户信息且已签名验证；或者使用 Authorization Code，需要后端再用 Code 去换 Token）。
3.  **Client:** 前端应用在其回调处理逻辑中，从 URL 或其他方式获取到 Google 返回的 `id_token`。
4.  **Client:** 向后端 API 发送请求：`POST /api/v1/auth/google/callback`，请求体为 `{"idToken": "eyJhbGciOiJSUzI1NiIs..."}`。
5.  **Backend - HTTP Handler (`AuthHandler.GoogleCallback`):**
    *   绑定请求体到 `GoogleCallbackRequestDTO`。
    *   验证 `idToken` 字段非空。
    *   调用 `authUseCase.AuthenticateWithGoogle(r.Context(), dto.IDToken)`。
    *   处理 Usecase 返回的结果：
        *   成功：得到应用 JWT (`authToken`) 和 `isNewUser` 标志。返回 200 OK，响应体 `{"token": authToken, "isNewUser": isNewUser}`。
        *   失败 (`domain.ErrAuthenticationFailed`): 返回 401 Unauthorized。
        *   失败 (`domain.ErrConflict`): 返回 409 Conflict，响应体可包含错误码和信息 `{"code": "EMAIL_EXISTS", "message": "Email is already registered with a different method."}`。
        *   其他内部错误：返回 500 Internal Server Error。
6.  **Backend - Usecase (`AuthUseCase.AuthenticateWithGoogle`):** (详细逻辑见 4.4 节)
    *   调用 `extAuthService.VerifyGoogleToken(ctx, idToken)` 验证 Google Token 并获取 `ExternalUserInfo`。
    *   根据 `ProviderUserID` (Google sub) 查找用户 (`userRepo.FindByProviderID`)。
    *   若找到，生成并返回应用 JWT。
    *   若未找到，根据 `Email` 查找 (`userRepo.FindByEmail`)。
    *   若找到 Email 但 `AuthProvider` 不是 'google' 或 `GoogleID` 已有值，返回 `domain.ErrConflict` (根据策略)。
    *   若 Email 也未找到，创建新用户 (`domain.NewGoogleUser`, `userRepo.Create`)，设置 `AuthProvider='google'`, `GoogleID=sub`，**不设置密码**。然后生成并返回应用 JWT 和 `isNewUser=true`。
7.  **Backend - Adapter (`GoogleAuthService.VerifyGoogleToken`):**
    *   使用 Google 客户端库调用 Google 的 token 验证端点。
    *   **执行严格验证:** 检查签名、`aud` (Audience 必须是后端应用的 Client ID)、`iss` (Issuer 必须是 `accounts.google.com` 或 `https://accounts.google.com`)、Token 是否过期。
    *   从验证通过的 Token Payload 中提取 `sub`, `email`, `email_verified`, `name`, `picture` 等信息。
    *   组装并返回 `port.ExternalUserInfo` 结构体。如果验证失败，返回包装后的错误。
8.  **Backend - Adapter (`PostgresUserRepository.FindByProviderID`, `FindByEmail`, `Create`):** 执行相应的 SQL 查询或插入操作。
9.  **Client:** 接收到后端的响应。
    *   若成功 (200 OK)，保存返回的应用 JWT。如果 `isNewUser` 为 true，可能需要引导用户完成额外的 Profile 设置。用户现在已通过应用自身 JWT 认证。
    *   若失败 (4xx/5xx)，向用户显示相应的错误信息。

## 7. 数据模型（详细）

以下是使用类 SQL DDL 定义的表结构，重点关注字段、类型、约束和索引。

```sql
-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'), -- Basic email format check
    password_hash VARCHAR(255) NULL, -- Can be NULL for external auth users
    name VARCHAR(100),
    google_id VARCHAR(255) UNIQUE NULL, -- Google User ID (subject claim)
    auth_provider VARCHAR(50) NOT NULL DEFAULT 'local', -- 'local', 'google', etc.
    profile_image_url VARCHAR(1024) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for users table
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL; -- Index only non-null values
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- 音频轨道表
CREATE TABLE audio_tracks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    language_code VARCHAR(10) NOT NULL, -- e.g., 'en-US', 'zh-CN'
    level VARCHAR(50), -- e.g., 'A1', 'B2', 'Native'
    duration_ms INTEGER NOT NULL DEFAULT 0 CHECK (duration_ms >= 0),
    minio_bucket VARCHAR(100) NOT NULL,
    minio_object_key VARCHAR(1024) NOT NULL,
    cover_image_url VARCHAR(1024),
    uploader_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    is_public BOOLEAN NOT NULL DEFAULT true,
    tags TEXT[] NULL, -- Array of tags
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for audio_tracks table
CREATE INDEX idx_audiotracks_language ON audio_tracks(language_code);
CREATE INDEX idx_audiotracks_level ON audio_tracks(level);
CREATE INDEX idx_audiotracks_uploader ON audio_tracks(uploader_id);
CREATE INDEX idx_audiotracks_tags ON audio_tracks USING GIN (tags); -- Index for array searching

-- 音频合集表 (课程、播放列表)
CREATE TABLE audio_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('COURSE', 'PLAYLIST')), -- Collection type
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for audio_collections table
CREATE INDEX idx_audiocollections_owner ON audio_collections(owner_id);
CREATE INDEX idx_audiocollections_type ON audio_collections(type);

-- 合集与轨道关联表 (多对多，带顺序)
CREATE TABLE collection_tracks (
    collection_id UUID NOT NULL REFERENCES audio_collections(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0 CHECK (position >= 0), -- Order within the collection
    PRIMARY KEY (collection_id, track_id)
);

-- Indexes for collection_tracks table
CREATE INDEX idx_collectiontracks_track_id ON collection_tracks(track_id); -- To quickly find which collections a track belongs to

-- 播放进度表
CREATE TABLE playback_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    progress_seconds INTEGER NOT NULL DEFAULT 0 CHECK (progress_seconds >= 0),
    last_listened_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, track_id)
);

-- Indexes for playback_progress table
CREATE INDEX idx_playbackprogress_lastlistened ON playback_progress(user_id, last_listened_at DESC); -- To quickly find recent progress

-- 书签表
CREATE TABLE bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    timestamp_seconds INTEGER NOT NULL CHECK (timestamp_seconds >= 0),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for bookmarks table
CREATE INDEX idx_bookmarks_user_track ON bookmarks(user_id, track_id, timestamp_seconds); -- Efficiently list bookmarks for a user on a track
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC); -- List recent bookmarks for a user

-- 字幕表 (如果选择存储在数据库中)
CREATE TABLE transcriptions (
    track_id UUID PRIMARY KEY REFERENCES audio_tracks(id) ON DELETE CASCADE,
    -- segments JSONB structure: [{"start": float, "end": float, "text": string}, ...]
    segments JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for transcriptions table (if complex queries on segments are needed, consider specific JSONB indexing)
-- CREATE INDEX idx_transcriptions_segments ON transcriptions USING GIN (segments);
```

**注意:**
*   `gen_random_uuid()` 需要启用 `pgcrypto` 扩展。
*   索引的设计需要根据实际的查询模式进行调整和优化。
*   时间戳统一使用 `TIMESTAMPTZ` (Timestamp with time zone) 以避免时区问题。
*   添加了 `CHECK` 约束来保证数据有效性。

## 8. API 设计

### 8.1 设计原则

*   **RESTful:** 基于资源进行设计，使用标准的 HTTP 方法 (GET, POST, PUT, DELETE, PATCH) 表达操作意图。
*   **JSON:** API 请求体和响应体主要使用 `application/json` 格式。
*   **版本化:** 在 URL 路径中包含主版本号，例如 `/api/v1/...`，以支持未来 API 的演进。
*   **一致性:**
    *   **命名:** API 端点路径使用小写和连字符（kebab-case）或下划线（snake_case）。JSON 字段建议使用驼峰式（camelCase）或下划线（snake_case），需在项目内保持统一（本文档示例偏向 camelCase）。
    *   **响应格式:** 成功响应包含数据，错误响应包含标准化的错误代码和消息。
    *   **状态码:** 正确使用 HTTP 状态码 (200, 201, 204, 400, 401, 403, 404, 409, 500等)。
*   **无状态:** 服务器不维护客户端的会话状态，每个请求都应包含所有必要的信息（主要通过 JWT 进行认证）。
*   **幂等性:** 对于 PUT 和 DELETE 操作，应设计为幂等的（多次执行相同请求效果一致）。GET 请求天然幂等。POST 通常不幂等（用于创建资源）。

### 8.2 资源与端点示例

*   **认证 (Authentication): `/api/v1/auth`**
    *   `POST /register`: 注册新用户 (邮箱+密码)。
    *   `POST /login`: 用户登录 (邮箱+密码)，返回 JWT。
    *   `POST /google/callback`: 处理 Google 登录回调，验证 ID Token，返回 JWT。
    *   `POST /refresh`: (可选) 使用 Refresh Token 获取新的 Access Token (JWT)。
*   **用户 (Current User): `/api/v1/users/me`** (需要认证)
    *   `GET /`: 获取当前登录用户的个人资料。
    *   `PUT /`: 更新当前登录用户的个人资料。
    *   `GET /progress`: 获取当前用户的播放进度列表 (分页)。
    *   `POST /progress`: 记录/更新当前用户的播放进度。
    *   `GET /progress/{trackId}`: 获取当前用户对特定音轨的播放进度。
    *   `GET /bookmarks`: 获取当前用户的书签列表 (可按 trackId 过滤, 分页)。
    *   `GET /collections`: 获取当前用户拥有的音频合集列表 (分页)。
*   **音频轨道 (Audio Tracks): `/api/v1/audio/tracks`**
    *   `GET /`: 列出/搜索音频轨道 (支持按语言、级别、标签等过滤，支持排序，分页)。
    *   `GET /{trackId}`: 获取单个音频轨道详情 (包含元数据和预签名播放 URL)。
*   **音频合集 (Audio Collections): `/api/v1/audio/collections`**
    *   `POST /`: 创建新的音频合集 (需要认证)。
    *   `GET /{collectionId}`: 获取合集详情 (包含元数据和轨道列表及顺序)。
    *   `PUT /{collectionId}`: 更新合集元数据 (标题、描述) (需要认证和所有权)。
    *   `DELETE /{collectionId}`: 删除合集 (需要认证和所有权)。
    *   `POST /{collectionId}/tracks`: 向合集添加音轨 (需要认证和所有权)。请求体可能包含 `{"trackId": "...", "position": 0}`。
    *   `PUT /{collectionId}/tracks`: 更新合集中音轨的顺序 (需要认证和所有权)。请求体 `{"orderedTrackIds": ["id1", "id2", ...]}`。
    *   `DELETE /{collectionId}/tracks/{trackId}`: 从合集中移除音轨 (需要认证和所有权)。
*   **书签 (Bookmarks): `/api/v1/bookmarks`** (需要认证)
    *   `POST /`: 创建新书签。
    *   `DELETE /{bookmarkId}`: 删除书签 (需要认证和所有权)。
*   **字幕 (Transcriptions): `/api/v1/transcriptions`**
    *   `GET /{trackId}`: 获取指定音轨的字幕数据。

### 8.3 数据传输对象 (DTO)

位于 `internal/adapter/handler/http/dto/` 目录下，为每个 API 端点的请求和响应定义 Go 结构体。

*   **命名:** 建议使用 `[Action][Resource]RequestDTO` 和 `[Resource]ResponseDTO` 或 `[Action][Resource]ResponseDTO` 的模式。
*   **Tags:** 使用 `json:"fieldName"` 控制 JSON 序列化/反序列化。使用 `validate:"rule"` (配合 `go-playground/validator`) 定义输入验证规则。
*   **示例:**
    ```go
    // POST /api/v1/auth/register
    type RegisterRequestDTO struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"` // Add complexity rules if needed
        Name     string `json:"name" validate:"omitempty,max=100"`
    }

    // POST /api/v1/auth/login & /api/v1/auth/google/callback Response
    type AuthResponseDTO struct {
        Token     string `json:"token"`
        IsNewUser *bool  `json:"isNewUser,omitempty"` // Only for Google Callback response
    }

     // POST /api/v1/auth/google/callback Request
     type GoogleCallbackRequestDTO struct {
         IDToken string `json:"idToken" validate:"required"`
     }

    // GET /api/v1/audio/tracks/{trackId} Response
    type AudioTrackDetailsResponseDTO struct {
        ID            string    `json:"id"`
        Title         string    `json:"title"`
        Description   string    `json:"description,omitempty"`
        LanguageCode  string    `json:"languageCode"`
        Level         string    `json:"level,omitempty"`
        DurationMs    int       `json:"durationMs"`
        CoverImageURL *string   `json:"coverImageUrl,omitempty"`
        PlayURL       string    `json:"playUrl"` // Presigned URL
        IsPublic      bool      `json:"isPublic"`
        Tags          []string  `json:"tags,omitempty"`
        // Optional related data:
        // UserProgressSeconds *int      `json:"userProgressSeconds,omitempty"`
        // Bookmarks           []BookmarkResponseDTO `json:"bookmarks,omitempty"`
        // TranscriptionAvailable bool `json:"transcriptionAvailable"`
    }

    // Common Error Response
    type ErrorResponseDTO struct {
        Code    string `json:"code"`    // e.g., "INVALID_INPUT", "NOT_FOUND", "UNAUTHENTICATED"
        Message string `json:"message"` // User-friendly error message
    }

    // Common Paginated List Response
    type PaginatedResponseDTO struct {
        Data  interface{} `json:"data"` // Array of response DTOs
        Total int         `json:"total"`
        Limit int         `json:"limit"`
        Offset int        `json:"offset"`
    }
    ```

### 8.4 认证与授权

*   **认证 (Authentication):**
    *   **主要机制:** 使用 **JWT (JSON Web Tokens)** 作为无状态认证凭证。
    *   **Token 获取:**
        *   本地登录 (`/auth/login`) 成功后颁发 JWT。
        *   外部认证 (`/auth/google/callback`) 成功后颁发 JWT。
    *   **Token 传递:** 客户端在后续需要认证的请求中，将 JWT 放入 `Authorization: Bearer <jwt_token>` HTTP Header 中。
    *   **Token 验证:** 后端通过 Auth 中间件验证 Token 的签名、有效期，并解析出 `userID` 注入请求的 `context.Context`。
    *   **(可选) Refresh Tokens:** 为提高安全性，可以实现 Refresh Token 机制。Access Token (JWT) 有效期短（如 15-60 分钟），Refresh Token 有效期长（如几天或几周）。当 Access Token 过期时，客户端使用 Refresh Token 调用 `/auth/refresh` 端点获取新的 Access Token。Refresh Token 需要安全存储（如数据库）并具备撤销机制。
*   **授权 (Authorization):**
    *   **基本授权:** Auth 中间件确保了访问受保护路由的用户必须是已认证的。
    *   **基于所有权的授权:** 在 Usecase 或 Handler 层面，对于修改或删除资源的操作，必须检查当前登录用户 (`userID` from context) 是否是该资源的所有者（例如，只能删除自己的书签、修改自己的合集）。
    *   **(可选) 基于角色的访问控制 (RBAC):** 如果未来引入管理员角色或其他角色，需要在 User 领域模型中添加角色信息，并在 Usecase/Handler 中检查用户角色是否满足操作所需的权限。

### 8.5 分页、过滤与排序

*   **分页:**
    *   **机制:** 使用 `limit` (每页数量) 和 `offset` (起始偏移量) 查询参数。
    *   **默认值:** 应设定合理的默认值（如 `limit=20, offset=0`）。
    *   **最大值:** `limit` 应有上限，防止客户端请求过多数据。
    *   **响应:** 列表接口的响应体应包含分页信息 (`total`, `limit`, `offset`) 和当前页的数据列表 (`data`)，使用 `PaginatedResponseDTO` 结构。
*   **过滤:**
    *   使用查询参数进行过滤，参数名应清晰表达过滤字段。
    *   示例: `GET /api/v1/audio/tracks?languageCode=en-US&level=B1&isPublic=true`
    *   后端 Usecase 和 Repository 需要解析这些参数并构建相应的查询条件。
*   **排序:**
    *   使用 `sort` 查询参数指定排序字段和方向。
    *   格式约定：`sort=fieldName` (升序), `sort=fieldName:asc`, `sort=fieldName:desc`。可支持多字段排序 `sort=level:asc,createdAt:desc`。
    *   后端需要解析排序参数，并转换为数据库查询的 `ORDER BY` 子句。需要限制可排序的字段以防止滥用。

### 8.6 API 文档

*   **规范:** 使用 **OpenAPI Specification v3 (OAS3)** 编写 API 文档。
*   **工具:**
    *   可以使用 Go 代码注释生成 (`swaggo/swag`)。
    *   可以手写 `openapi.yaml` 或 `openapi.json` 文件 (更灵活，推荐)。
    *   可以使用可视化编辑器 (Stoplight Studio, Swagger Editor)。
*   **内容:** 文档必须包含：
    *   API 基本信息（版本、标题、描述）。
    *   服务器 URL (开发、生产)。
    *   认证方式定义 (JWT Bearer Auth)。
    *   所有端点的路径、HTTP 方法、描述、标签 (Tags)。
    *   每个端点的参数定义（路径、查询、请求体）、必填项、数据类型、格式、示例。
    *   每个端点的响应定义（HTTP 状态码、描述、响应体 Schema、示例）。
    *   可重用的 Schema 定义（用于 DTOs）。
*   **用途:** 作为前后端开发的契约，方便测试、生成客户端代码、提供交互式文档 (Swagger UI, Redoc)。

## 9. 非功能性需求考量

### 9.1 安全性

*   **认证与授权:**
    *   强制使用 HTTPS 防止数据在传输过程中被窃听或篡改。
    *   JWT 安全：使用强密钥 (`HS256` 至少 256 位随机密钥，`RS256` 使用 RSA 密钥对)，设置合理的有效期，考虑使用 Refresh Token 机制。服务端验证签名和 `exp` 声明。
    *   OAuth 2.0 安全：严格验证外部 IdP 返回的 Token (ID Token 的 `aud`, `iss`, `exp`)，使用 `state` 参数防 CSRF（前端责任），安全存储 Client Secret。
    *   密码安全：使用 `bcrypt` (cost factor >= 10) 存储密码哈希。实现密码强度策略和密码重置流程（通过邮件发送有时效的重置链接）。
    *   权限检查：确保用户只能访问和修改其自身的数据（所有权检查）。
*   **输入验证与输出编码:**
    *   在 Handler 层对所有外部输入（URL 参数、Query 参数、请求体）进行严格验证（类型、长度、格式、范围）。
    *   对所有输出到客户端（尤其是未来可能渲染到 HTML）的数据进行适当编码，防止 XSS 攻击。API 响应 JSON 时，Go 的 `encoding/json` 默认会进行 HTML 转义。
*   **依赖安全:**
    *   定期使用 `govulncheck` 扫描项目依赖，及时更新有漏洞的库。
*   **基础设施安全:**
    *   数据库访问凭证、MinIO 凭证、JWT 密钥、Google Client Secret 等敏感信息绝不能硬编码或提交到代码库。使用环境变量、配置文件（注意权限）或 Secret Management 系统（如 HashiCorp Vault, AWS Secrets Manager）进行管理和注入。
    *   配置 MinIO Bucket 为私有访问，仅允许通过预签名 URL 或特定 IAM 策略访问。
    *   数据库网络访问控制。
*   **速率限制 (Rate Limiting):**
    *   在 API Gateway 或应用中间件层实施速率限制，防止暴力破解和 DoS 攻击。可以基于 IP 地址或用户 ID 进行限制。
*   **防止注入:**
    *   使用参数化查询或 ORM 提供的安全方法，严防 SQL 注入。

### 9.2 性能与可伸缩性

*   **数据库优化:**
    *   **索引:** 根据查询模式（WHERE 子句、JOIN 条件、ORDER BY 字段）创建合适的数据库索引。定期分析慢查询 (`EXPLAIN ANALYZE`)。
    *   **连接池:** 配置合理的数据库连接池大小 (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`)，避免连接耗尽或过多空闲连接。
    *   **查询优化:** 避免在循环中执行数据库查询 (N+1 问题)，使用 JOIN 或单次查询获取所需数据。编写高效的 SQL 语句。
*   **对象存储与 CDN:**
    *   **CDN:** **强烈建议**在 MinIO 前端部署 CDN (Content Delivery Network)。客户端通过 CDN 的边缘节点获取音频，极大降低延迟，提高全球访问速度和可用性，并分担 MinIO 负载。预签名 URL 应指向 CDN 地址（配置 CDN 回源到 MinIO）。
    *   **MinIO 优化:** 根据负载情况调整 MinIO 的部署模式（分布式）、硬件资源和网络配置。
*   **API 性能:**
    *   **代码优化:** 减少不必要的计算和内存分配。优化算法复杂度。
    *   **并发:** 合理利用 Go 的 Goroutine 处理并发请求，但要注意控制并发度，避免耗尽下游资源（如数据库连接）。
    *   **缓存:** 对于频繁访问且不经常变化的数据（如热门音频列表、用户配置），引入内存缓存 (in-memory cache, e.g., `ristretto`) 或分布式缓存 (Redis, Memcached)。需要仔细设计缓存键、缓存粒度、过期策略和缓存失效机制。
*   **可伸缩性:**
    *   **无状态服务:** 后端 API 设计为无状态，使得可以轻松地水平扩展（运行多个相同的服务实例）。
    *   **负载均衡:** 在多个 API 实例前使用负载均衡器（Nginx, HAProxy, 云提供商 LB）分发流量。
    *   **数据库扩展:** 考虑数据库读写分离、分片（Sharding）等策略（复杂，按需引入）。

### 9.3 可观测性 (Observability) - Logging, Metrics, Tracing (LMT)

*   **结构化日志 (Logging):**
    *   **工具:** `slog` (标准库) 或 `zap`/`zerolog`。
    *   **格式:** JSON 或 Logfmt，便于机器解析。
    *   **内容:** 每条日志包含 `timestamp`, `level` (INFO, WARN, ERROR), `message`, `service_name`, `request_id`, `user_id` (若有), 以及与事件相关的键值对。错误日志应包含 `error` 字段和 `stack_trace` (对于 panic 或严重错误)。
    *   **级别:** 生产环境通常设置为 INFO 或 WARN 级别。
    *   **聚合:** 将日志发送到集中式日志系统 (ELK Stack, Loki+Grafana, Datadog Logs)。
*   **指标 (Metrics):**
    *   **工具:** 使用 `prometheus/client_golang` 库在应用内暴露 `/metrics` 端点。
    *   **采集:** 使用 Prometheus Server 定期抓取 `/metrics` 端点的数据。
    *   **可视化:** 使用 Grafana 创建仪表盘展示关键指标。
    *   **关键指标 (RED):**
        *   **Rate:** 每秒请求数 (QPS)，按端点、HTTP 方法、状态码区分。
        *   **Errors:** 请求错误率，按端点、状态码区分。数据库/外部服务调用错误计数。
        *   **Duration:** 请求处理延迟（平均值、P95、P99 分位数），按端点区分。数据库/外部服务调用延迟。
    *   **其他指标:** Goroutine 数量、内存/CPU 使用率（通过 node-exporter 或 cAdvisor 获取）、数据库连接池状态（活跃、空闲、等待）。
*   **分布式追踪 (Tracing):**
    *   **标准/工具:** 使用 OpenTelemetry (OTel) Go SDK。
    *   **实现:**
        *   在请求入口（中间件）启动 Trace，生成 Trace ID 和根 Span ID，通过 `context.Context` 向下传递。
        *   在关键操作（Handler 处理、Usecase 执行、Repository 查询、外部服务调用）开始时创建子 Span，结束时记录耗时和状态（成功/失败及错误信息）、添加相关属性 (Attributes)。
        *   确保 Trace Context (Trace ID, Span ID) 跨网络调用（如调用 MinIO、Google API）传播（如果这些服务也支持追踪）。
    *   **后端:** 将 Trace 数据导出到追踪后端系统 (Jaeger, Tempo, Datadog APM)。
    *   **用途:** 分析请求的完整链路，定位性能瓶颈和错误源头。

### 9.4 可测试性

*   **单元测试 (Unit Testing):**
    *   **目标:** 测试单个函数或方法的逻辑，隔离其依赖。
    *   **`domain` 层:** 直接测试其业务逻辑，因为无外部依赖。
    *   **`usecase` 层:** 使用 Mock 框架 (`testify/mock`, `gomock`) 模拟 `port.Repository` 和 `port.Service` 接口，断言 Usecase 是否正确调用了接口方法以及是否处理了不同的返回值。
    *   **`handler` 层:** 使用 `net/http/httptest` 创建模拟的 HTTP 请求和响应记录器。Mock Usecase 接口依赖，断言 Handler 是否正确解析请求、调用 Usecase、处理结果并生成了预期的 HTTP 响应。
    *   **`pkg` 工具库:** 直接测试其功能。
    *   **覆盖率:** 目标单元测试覆盖率 > 80%。
*   **集成测试 (Integration Testing):**
    *   **目标:** 测试多个组件之间的协作是否符合预期，特别是与外部依赖（数据库、MinIO）的交互。
    *   **Repository 层:** 编写测试用例，使用 **测试容器 (`ory/dockertest`)** 在测试执行期间启动临时的 PostgreSQL 数据库实例。连接到该实例，执行 Repository 的方法，并验证数据库状态的变化或查询结果是否正确。测试结束后销毁容器。
    *   **Service 层 (MinIO/Google):** 类似地，可以启动临时的 MinIO 容器进行测试。对于 Google Auth，可以 Mock `http.Client` 来模拟 Google API 的响应，或者依赖 Mock `ExternalAuthService` 接口进行更高层次的集成测试。
    *   **Handler -> Usecase -> Repository 流程:** 可以编写覆盖这个流程的集成测试，使用真实的 Repository 实现（连接到测试容器数据库），但 Mock 其他外部服务（如 MinIO）。
*   **契约测试 (Contract Testing - 可选):**
    *   使用 OpenAPI 规范作为契约。可以使用工具（如 `dredd`) 验证 API 实现是否符合规范定义。确保前后端对 API 的理解一致。
*   **测试组织:** 测试文件 (`_test.go`) 放在被测代码的同一目录下。集成测试可以放在单独的 `test/integration` 目录下。

### 9.5 配置管理

*   **工具:** `spf13/viper`。
*   **来源:** 支持从配置文件 (YAML, JSON, TOML)、环境变量、命令行标志、远程配置中心 (etcd, Consul) 加载。
*   **优先级:** 明确配置加载的优先级顺序（通常：命令行标志 > 环境变量 > 配置文件 > 默认值）。
*   **结构化:** 定义 `Config` 结构体 (`internal/config/config.go`) 映射所有配置项，便于类型安全访问。
    ```go
    type Config struct {
        Server   ServerConfig   `mapstructure:"server"`
        Database DatabaseConfig `mapstructure:"database"`
        JWT      JWTConfig      `mapstructure:"jwt"`
        Minio    MinioConfig    `mapstructure:"minio"`
        Google   GoogleConfig   `mapstructure:"google"`
        Log      LogConfig      `mapstructure:"log"`
        // ... other sections
    }
    // Define nested structs like ServerConfig { Port string `mapstructure:"port"` ... }
    ```
*   **环境区分:** 通过不同的配置文件名 (`config.dev.yaml`, `config.prod.yaml`) 或环境变量 (`APP_ENV=production`) 来加载对应环境的配置。
*   **热加载 (可选):** Viper 支持监听配置文件变化并热加载配置（需要应用代码配合处理配置更新）。
*   **敏感信息:** 严格遵守安全原则，通过环境变量或 Secret Management 注入。

### 9.6 错误处理策略

*   **错误类型与包装:**
    *   在 `domain` 层定义业务相关的标准错误（`ErrNotFound`, `ErrConflict`等）。
    *   在 `repository` 或 `service` (Adapter) 层，将底层库的错误（如 `pgx.ErrNoRows`, `minio.ErrorResponse`）转换为 `domain` 层定义的标准错误或包装后返回，使用 `fmt.Errorf("operation failed: %w", err)`。
    *   在 `usecase` 层，处理 Repository/Service 返回的错误，或直接向上传递包装后的错误。
*   **错误传递:** 错误应在调用链中清晰传递，保留原始错误信息和上下文。
*   **Handler 层处理:**
    *   统一处理 Usecase 返回的错误。
    *   使用 `errors.Is()` 或 `errors.As()` 判断错误类型。
    *   将业务错误映射为合适的 HTTP 状态码和标准化的错误响应体 (`ErrorResponseDTO`)。
        *   `domain.ErrNotFound` -> `404 Not Found` (`{"code": "NOT_FOUND", ...}`)
        *   `domain.ErrInvalidArgument` / Validation Error -> `400 Bad Request` (`{"code": "INVALID_INPUT", ...}`)
        *   `domain.ErrPermissionDenied` / `domain.ErrAuthenticationFailed` -> `401 Unauthorized` or `403 Forbidden` (`{"code": "UNAUTHENTICATED/FORBIDDEN", ...}`)
        *   `domain.ErrConflict` -> `409 Conflict` (`{"code": "RESOURCE_CONFLICT", ...}`)
        *   其他未预期的错误 -> `500 Internal Server Error` (`{"code": "INTERNAL_ERROR", "message": "An unexpected error occurred."}`)
    *   **日志记录:** 对于 5xx 错误，必须记录详细的错误信息和堆栈跟踪。对于 4xx 错误，可以记录级别较低的日志（如 WARN 或 INFO）。
    *   **客户端信息:** 不向客户端暴露敏感的内部错误细节（如 SQL 语句、堆栈跟踪）。返回标准化的错误代码和用户友好的消息。
*   **Panic 处理:** Recovery 中间件负责捕获 `panic`，记录详细错误和堆栈，并返回 500 错误响应。

## 10. 代码规范与质量

*   **Go 标准:** 遵循 Go 官方的代码风格指南。
*   **格式化:** 强制使用 `gofmt` 或 `goimports` 统一代码格式。提交前自动格式化。
*   **Linting:** 在 CI/CD 流程中集成 `golangci-lint`，使用配置好的规则集进行静态代码分析，检查潜在错误、风格问题、复杂度等。
*   **命名约定:**
    *   包名：小写、简洁、有意义。
    *   变量/函数/类型：驼峰式 (`camelCase`)。
    *   接口：通常以 `er` 结尾 (如 `Reader`, `Writer`) 或描述其能力。
    *   常量：驼峰式或全大写（根据上下文）。
*   **注释:**
    *   为所有公共的类型、函数、方法、常量编写清晰的 Go Doc 注释。
    *   对复杂的代码逻辑块添加必要的行内注释解释意图。
*   **简洁性与可读性:** 编写简洁、清晰、易于理解的代码。避免过度复杂的抽象和嵌套。函数/方法应尽可能短小，单一职责。
*   **错误处理:** 显式处理所有可能返回错误的函数调用。避免忽略错误 (`_ = someFunc()`)，除非明确知道错误无关紧要。
*   **资源管理:** 确保及时关闭需要关闭的资源（如 `*sql.Rows`, 文件句柄），善用 `defer`。
*   **并发安全:** 谨慎处理共享状态的并发访问，使用 Mutex, RWMutex, Channel 或原子操作保证线程安全。
*   **代码审查 (Code Review):** 所有代码变更（特别是新功能和重构）必须经过至少一位其他团队成员的审查才能合并到主分支。关注逻辑正确性、代码风格、可维护性、测试覆盖率等。

## 11. 部署策略（概要）

*   **构建:**
    *   使用 Go 的交叉编译能力生成目标平台（通常是 Linux amd64）的静态链接二进制文件 (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ...`)。
*   **容器化:**
    *   编写 `Dockerfile`，采用 **多阶段构建 (Multi-stage build)**：
        *   第一阶段 (`builder`): 使用包含 Go 工具链的基础镜像，下载依赖 (`go mod download`)，编译应用，生成二进制文件。
        *   第二阶段 (final): 使用一个极简的基础镜像（如 `alpine` 或 `gcr.io/distroless/static-debian11`），仅复制第一阶段生成的二进制文件和必要的配置文件/静态资源。这可以显著减小最终镜像的体积和攻击面。
    *   在 Dockerfile 中设置非 root 用户运行应用。
*   **镜像管理:** 将构建好的 Docker 镜像推送到容器镜像仓库（Docker Hub, Google Container Registry (GCR), AWS Elastic Container Registry (ECR), Harbor 等）。使用 Git commit hash 或语义化版本号作为镜像标签。
*   **配置注入:** 通过 **环境变量** 将配置（特别是敏感信息）注入到运行的容器中。Kubernetes 可以使用 ConfigMaps 和 Secrets。
*   **数据库迁移:**
    *   使用 `golang-migrate` 或 `goose` 管理数据库 Schema 变更脚本（SQL 文件）。
    *   将数据库迁移作为部署流程的一部分：
        *   **方式一 (应用启动时迁移):** 在应用启动逻辑 (`main.go`) 中调用迁移库执行迁移。简单但不推荐用于生产环境（可能导致多实例并发迁移冲突，启动时间变长）。
        *   **方式二 (独立迁移任务 - 推荐):** 在 CI/CD 流水线中，部署新版本应用 *之前*，运行一个单独的 Job/Task 来执行数据库迁移命令。
*   **部署环境:**
    *   **开发环境:** 可以使用 Docker Compose 在本地运行应用及其依赖（PostgreSQL, MinIO）。
    *   **测试/预发环境:** 与生产环境尽量一致，使用独立的数据库和配置。
    *   **生产环境:** 推荐使用容器编排平台（Kubernetes, Docker Swarm）或云平台的 PaaS 服务 (AWS ECS/EKS, Google Cloud Run/GKE, Azure AKS) 进行部署和管理。
*   **部署策略:**
    *   **滚动更新 (Rolling Update):** 逐步替换旧版本的实例，保证服务不中断，是 K8s 等平台的默认策略。
    *   **蓝绿部署 (Blue-Green Deployment):** 同时部署新旧两个版本的环境，测试通过后将流量切换到新版本。回滚快速。
    *   **金丝雀发布 (Canary Release):** 先将少量流量导入新版本，观察稳定后再逐步扩大流量比例。
*   **健康检查:** 实现 HTTP 健康检查端点 (`/healthz` 或 `/livez`)，供负载均衡器或容器编排平台检查应用实例是否存活和准备好接收流量。

## 12. 未来考虑

*   **缓存策略细化:** 针对具体的热点数据（如课程列表、用户 Profile）设计详细的缓存方案，包括缓存工具选型 (Redis/Memcached)、缓存粒度、过期时间、失效策略（如 TTL, LRU, 写操作时失效）。
*   **全文搜索:** 如果对音频标题、描述、甚至字幕内容的搜索需求变得复杂，考虑引入专门的搜索引擎如 Elasticsearch 或 OpenSearch，将相关数据索引进去。
*   **异步任务处理:** 对于耗时较长或可以后台执行的操作（如上传音频后的转码、格式转换、自动生成字幕、发送批量通知邮件），引入异步任务队列（如 Asynq, Machinery）和后台 Worker 服务来处理，避免阻塞 API 请求。
*   **实时功能:** 如果需要实现实时通知（如新课程上线推送、协作功能），可以引入 WebSocket 或 Server-Sent Events (SSE) 技术。
*   **内容管理后台 (Admin Panel):** 开发一个独立的 Web 应用或 API，供运营人员管理用户、上传/编辑音频内容、管理课程、查看统计数据等。可能需要更细粒度的权限控制。
*   **多租户支持:** 如果未来需要为不同的组织或学校提供隔离的服务实例，需要从数据模型、API 设计、认证授权等多个层面进行改造以支持多租户。
*   **国际化 (i18n) 与本地化 (l10n):** 如果应用需要支持多种界面语言，API 层面可能需要支持根据 `Accept-Language` Header 返回不同语言的错误消息或元数据。
*   **微服务演进:** 随着业务复杂度和团队规模的增长，如果单体应用遇到瓶颈（开发效率、部署耦合、技术栈限制、扩展性需求差异），可以考虑按照清晰的业务边界（如用户服务、内容服务、学习活动服务、支付服务）逐步将单体拆分为独立的微服务。当前的分层和模块化设计是未来进行拆分的基础。
