好的，我们来梳理一下当前项目的状态，总结一下基于我们之前的讨论和架构设计文档，还有哪些功能需要补充或可以进一步完善的地方。

**语言学习音频播放器 - 后端项目状态与待办事项总结**

**版本:** 基于架构设计 v1.3 的实现进度
**日期:** 2023-10-27 (总结时点)

**A. 已完成的核心功能和基础架构:**

1.  **项目结构:** 遵循分层架构 (cmd, internal/domain, internal/port, internal/usecase, internal/adapter, pkg)。
2.  **基础设置:** Go Modules, Git, 配置管理 (Viper), 结构化日志 (slog), 优雅停机。
3.  **领域模型:** 定义了核心实体 (User, AudioTrack, AudioCollection, PlaybackProgress, Bookmark) 和值对象。
4.  **端口定义:** 定义了 Repository, Service, Helper 的接口。
5.  **核心认证:**
    *   本地用户注册 (邮箱+密码)。
    *   本地用户登录 (返回 JWT)。
    *   JWT 签发与验证中间件。
    *   Google OAuth 2.0 登录/注册 (通过 ID Token 验证流程)。
    *   密码哈希 (bcrypt)。
6.  **音频内容管理:**
    *   创建/获取/更新/删除音频合集 (Audio Collections - Metadata & Track list management)。
    *   获取音频轨道详情 (Audio Tracks)，包括预签名 GET URL 用于播放。
    *   列表查询音频轨道 (支持过滤、排序、分页)。
7.  **用户活动:**
    *   记录/获取播放进度 (Playback Progress)。
    *   创建/获取/删除书签 (Bookmarks)。
8.  **文件上传支持:**
    *   实现了后端生成预签名 PUT URL 的逻辑 (`FileStorageService.GetPresignedPutURL`)。
    *   实现了请求上传 URL (`/uploads/audio/request`) 和完成上传并创建 Track 记录 (`POST /audio/tracks`) 的 Usecase 和 Handler 逻辑。
9.  **数据库:**
    *   使用 PostgreSQL。
    *   使用 pgx 连接池。
    *   使用 golang-migrate 管理数据库迁移 (已包含 users, audio\*, activity\* 表)。
    *   Repository 实现基本 CRUD 操作。
10. **对象存储:**
    *   使用 MinIO (通过 S3 兼容接口)。
    *   实现了获取预签名 GET/PUT URL 和删除对象的 Service Adapter。
11. **基础 Web 服务:**
    *   使用 Chi Router。
    *   实现了基础中间件 (Recovery, RequestID, Logger, RealIP, CORS, Timeout)。
12. **初步 NFR 实现:**
    *   **错误处理:** 定义了领域错误，实现了 Handler 层的统一错误响应映射，部分 Repository 添加了错误映射 (ErrNotFound, ErrConflict)。
    *   **配置管理:** 支持文件和环境变量，区分环境加载。
    *   **安全性:** 使用了 bcrypt, JWT, 参数化查询，添加了基础速率限制示例。
13. **初步测试:**
    *   提供了 Domain, Usecase (单元, Mock), Repository (集成, dockertest), Handler (单元, Mock) 的测试示例代码。
14. **初步文档与部署:**
    *   提供了 OpenAPI (Swagger) 文档的骨架和主要端点示例 (`openapi.yaml`)。
    *   提供了 `Dockerfile` 用于容器化构建。
    *   提供了 `Makefile` 用于简化开发流程（工具安装、构建、测试、迁移、生成文档等）。

**B. 需要补充或完善的功能与细节:**

1.  **核心功能 & 逻辑:**
    *   **音频时长获取:** `CompleteUpload` Usecase 目前需要客户端提供 `DurationMs`。实际应用中，更可靠的方式是上传成功后，**触发一个异步任务**（当前未实现）去 MinIO 下载文件或使用 `ffprobe` 等工具分析获取准确时长，然后更新 `audio_tracks` 记录。
    *   **用户 Profile 管理:** `GET /users/me` 仅返回了占位符/用户ID。需要实现完整的 Handler 和 Usecase 逻辑来从 `UserRepository` 获取并返回 `UserResponseDTO` 定义的用户信息。同时需要实现 `PUT /users/me` 端点来允许用户更新他们的个人资料（如姓名、头像 URL - 头像本身上传是另一回事）。
    *   **Track 记录与 MinIO 对象生命周期同步:** 当通过 API 删除 `audio_tracks` 记录时，对应的 MinIO 对象**目前不会被删除**。需要扩展 `AudioContentUseCase` 中的删除逻辑，在删除数据库记录**之后**（或之前，根据策略）调用 `FileStorageService.DeleteObject` 来清理存储桶中的文件。反之，如果 MinIO 对象因故丢失，数据库记录不会自动更新（这通常需要额外的校验或后台任务）。
    *   **文件上传 - 错误处理:** `CompleteUpload` 中对 MinIO 对象存在的校验是可选的，可以考虑添加（调用 MinIO 的 `StatObject`）以确保客户端确实上传了文件，再创建数据库记录。
    *   **事务管理标准化:** `AudioCollectionRepository.ManageTracks` 直接使用了 `db.Begin`。虽然 `TransactionManager` 接口已在 `port` 中定义，但尚未在 `postgres` 适配器中实现，也未在需要跨 Repository 操作的 Usecase (如 `CreateCollection` 带初始 Tracks，如果需要原子性的话) 中使用。应实现 `TransactionManager` 并统一使用其 `Execute` 方法来包装事务逻辑。

2.  **测试 (Phase 7):**
    *   **覆盖率提升:** 当前测试覆盖率较低。需要系统性地为所有 Usecase、Handler、Repository 和 Pkg 编写单元测试和集成测试，目标覆盖率应 > 80%。
    *   **Usecase 单元测试:** 重点测试所有业务逻辑分支、错误处理路径、与 Mock 交互的正确性。
    *   **Repository 集成测试:** 覆盖所有方法，包括各种过滤、排序、分页场景，以及事务（如果使用 `TransactionManager`）。测试外键约束和级联删除行为。
    *   **Handler 单元测试:** 测试请求解析、参数绑定、DTO 验证、Usecase 调用、认证/授权上下文处理、响应格式化。
    *   **并发测试 (可选):** 对于可能存在竞态条件的逻辑（虽然当前设计中较少），可以编写并发测试。

3.  **API 文档 (Phase 9.1):**
    *   **完成度:** `openapi.yaml` 需要补充所有缺失的端点（特别是 Collections 和 Uploads）、请求/响应 Schema 定义 (DTOs)。
    *   **准确性:** 确保文档中的参数、响应码、安全要求、数据类型与实际 Handler 实现完全一致。
    *   **描述:** 完善各端点和 Schema 的 `description` 和 `summary`，使其清晰易懂。
    *   **验证与 UI:** 使用验证工具检查规范，并强烈建议集成 Swagger UI 或 Redoc 提供交互式文档。

4.  **非功能性需求 (Phase 8):**
    *   **错误处理:**
        *   **系统性映射:** 在所有 Repository 的 `Create`/`Update` 方法中，系统地检查并映射 `pgerrcode.UniqueViolation` 到 `domain.ErrConflict`，并映射 `pgerrcode.ForeignKeyViolation` 到 `domain.ErrInvalidArgument` 或 `domain.ErrNotFound` (取决于具体情况)。
        *   **Service 错误映射:** 在 MinIO/Google Service Adapter 中映射更具体的错误到领域错误。
        *   **响应细节:** 考虑在 `ErrorResponseDTO` 中添加可选的 `details` 字段，用于返回更具体的验证错误信息（例如哪个字段失败了）。
    *   **可观测性 (Observability):**
        *   **Metrics:** 实现 Prometheus 指标暴露 (`/metrics` 端点)，收集 RED 指标和其他关键业务指标。
        *   **Tracing:** 实现 OpenTelemetry 分布式追踪，跨请求和服务调用。
    *   **安全性:**
        *   **生产密钥管理:** 必须使用环境变量、Secrets Manager 或 Vault 等工具管理生产环境的敏感配置（JWT Secret, DB 密码, Google Secret, MinIO Keys）。
        *   **权限细化 (可选):** 当前主要是基于所有权授权。如果未来需要更复杂的角色（如管理员、编辑），需要引入 RBAC (Role-Based Access Control)。
        *   **速率限制器增强:** 当前内存限流器适用于单实例，生产环境（多实例）需要基于 Redis 或类似存储的分布式限流器。
    *   **性能与伸缩性:**
        *   **缓存:** 尚未实现任何缓存策略。对于热点数据（如热门 Track 列表、用户配置等）应考虑引入内存缓存 (Ristretto) 或分布式缓存 (Redis)。
        *   **CDN:** 部署时应在 MinIO 前配置 CDN 以优化音频分发。
        *   **数据库优化:** 定期审查慢查询，确保索引合理。检查 N+1 查询问题（尤其是在加载关联数据时，如 Collection 的 Tracks）。

5.  **部署 (Phase 9.3):**
    *   **CI/CD 流水线:** 需要完整设计和实现，包括自动化测试、构建、推送、迁移和部署步骤。
    *   **生产环境配置:** 需要为生产环境准备好基础设施（数据库、MinIO、负载均衡器、容器编排平台等）和相应的配置。

6.  **代码质量:**
    *   **TODO 清理:** 查找并处理代码中遗留的 `// TODO:` 注释。
    *   **日志审查:** 确保日志在所有关键路径都有记录，且信息足够用于调试，同时避免记录过多敏感信息。

**优先级建议:**

1.  **测试完善 (High):** 确保核心功能稳定可靠，是后续开发的基础。
2.  **OpenAPI 文档完善 (High):** 清晰的 API 契约对前后端协作至关重要。
3.  **错误处理细化 (Medium):** 提升应用的健壮性和用户体验。
4.  **核心功能完善 (Medium):** 补全 Profile 管理、Track 删除联动 MinIO 删除、音频时长处理等逻辑。
5.  **可观测性 (Medium/High for Prod):** Metrics 和 Tracing 对生产环境监控和排错非常重要。
6.  **CI/CD (High for Prod):** 实现自动化部署流程。
7.  **其他 NFRs 和功能增强 (Low/Medium - As Needed):** 缓存、RBAC、异步任务等根据实际需求引入。

这份总结应该能给你一个清晰的视图，了解当前项目的完成度以及接下来的工作重点。