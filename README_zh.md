# inori-music

音乐集中平台，目标是构建支持 Web、Android、iOS、PC 等平台的全平台音乐播放系统，同时兼容 B/S 与 C/S 架构。

## 版本

当前架构基线版本：`0.22.0`。

## 文档入口

- [`README.md`](README.md)：项目入口与文档导航。
- [`README_zh.md`](README_zh.md)：中文完整说明。
- [`requirement.md`](requirement.md)：中文需求基线与版本历史。
- [`.plan/`](.plan/)：中文阶段计划与完成状态。
- [`docs/architecture/`](docs/architecture/)：中文架构说明。
- [`docs/adr/`](docs/adr/)：中文架构决策记录。

## 0.x 技术方向

- 跨平台客户端：优先采用 Flutter，覆盖 Web、Android、iOS 与桌面端。
- 服务端：优先采用 Go 模块化单体，后续按边界拆分服务。
- 服务端元数据数据库：PostgreSQL 优先。
- 客户端本地存储：SQLite，用于离线队列、缓存索引和本地检索。
- 搜索：0.x 从 PostgreSQL 全文检索开始，后续可接入外部搜索引擎。
- 媒体存储：服务端统一管理多后端存储配置，大文件不进入关系型数据库。

## 已完成阶段

### 阶段 1：存储架构

确立由服务端统一管理的多后端媒体存储方向，覆盖 local、nfs、smb、s3 与 distributed。

### 阶段 2：存储领域脚手架

建立 Go API 与存储领域模型，提供验证、能力推导、默认后端与内存仓储。

### 阶段 3：存储管理 HTTP API

暴露基础管理 HTTP API，支持校验、注册、列表、默认选择与禁用。

### 阶段 4：管理端认证

使用 INORI_ADMIN_TOKEN 保护 /api/v1/admin/* 管理路由。

### 阶段 5：文件系统健康探测

为本地、NFS、SMB 与挂载式分布式后端增加安全文件系统探测。

### 阶段 6：S3 兼容对象探测

为 S3 兼容后端增加受控对象写读删探测与环境变量凭据解析。

### 阶段 7：健康刷新与容量报告

增加批量刷新、后台刷新调度与文件系统容量报告。

### 阶段 8：OpenAPI 合同

发布 OpenAPI 3.1 合同并通过测试约束路由与安全声明。

### 阶段 9：持久化文件仓储

增加可选 JSON 文件仓储，供开发与单节点自托管保留后端状态。

### 阶段 10：媒体对象登记脚手架

增加媒体对象登记领域，记录二进制资产引用而不保存媒体字节。

### 阶段 11：媒体对象 HTTP API

暴露媒体对象注册、查询与过滤 HTTP 管理接口。

### 阶段 12：媒体对象文件仓储

增加可选 JSON 文件仓储以持久化媒体对象元数据。

### 阶段 13：媒体对象完整性校验

增加只读完整性校验，验证文件存在、大小与 sha256。

### 阶段 14：批量媒体对象校验

增加按后端或内容哈希批量校验，并在单项失败后继续执行。

### 阶段 15：最新校验状态持久化

将最新校验结果持久化到媒体对象元数据。

### 阶段 16：校验状态过滤

支持按 verified、failed、unknown 查询最新校验状态。

### 阶段 17：媒体对象分页列表

为媒体对象列表增加 limit、offset 与分页元数据。

### 阶段 18：媒体对象元数据统计

增加只读元数据统计，供管理看板统计对象数量、大小和状态桶。

### 阶段 19：生命周期管理

增加生命周期元数据更新，deleted 作为终态且不删除真实字节。

### 阶段 20：生命周期过滤

支持按 lifecycleState 过滤媒体对象列表。

### 阶段 21：资产类型过滤

支持按 assetKind 过滤媒体对象列表。

### 阶段 22：中文文档与 README 拆分

将 Markdown 文档中文化，并拆分 README.md 与 README_zh.md。

## 运行 API 脚手架

```bash
INORI_ADMIN_TOKEN=change-me-development-token INORI_STORAGE_REPOSITORY_FILE=./var/storage-backends.json INORI_MEDIA_OBJECT_REPOSITORY_FILE=./var/media-objects.json INORI_STORAGE_REFRESH_INTERVAL=15m go run ./services/api/cmd/server
```

默认监听 `127.0.0.1:8080`。管理接口需要 `Authorization: Bearer <INORI_ADMIN_TOKEN>`。未配置 `INORI_STORAGE_REPOSITORY_FILE` 或 `INORI_MEDIA_OBJECT_REPOSITORY_FILE` 时，服务使用内存仓储。

## 后续展望

- 引入 PostgreSQL 迁移与索引，替换当前开发/自托管用 JSON 仓储。
- 增加导入任务、审计事件、批量生命周期变更与管理端 UI。
- 在播放器侧补齐流式播放、缓存、离线队列和跨端同步。
