# 存储管理 HTTP API

## 范围

HTTP API 提供受认证保护的管理接口，用于管理存储后端、执行探测、刷新健康状态、读取容量、登记媒体对象、执行只读完整性校验、更新生命周期、查询分页列表和查看元数据统计。

## 认证

`/healthz` 保持公开；`/api/v1/admin/*` 需要 `Authorization: Bearer <INORI_ADMIN_TOKEN>`。未配置管理令牌时，管理接口以 `503 admin_auth_not_configured` 失败关闭。

## 主要端点

- `GET /healthz`：进程健康检查。
- `GET/POST /api/v1/admin/storage/backends`：列出或注册存储后端。
- `POST /api/v1/admin/storage/backends/validate`：仅验证候选后端配置。
- `POST /api/v1/admin/storage/backends/refresh`：批量刷新后端健康与容量。
- `POST /api/v1/admin/storage/backends/{id}/probe`：执行安全探测。
- `GET /api/v1/admin/media/objects`：按单一元数据条件分页查询媒体对象。
- `POST /api/v1/admin/media/objects`：登记媒体对象元数据。
- `GET /api/v1/admin/media/objects/stats`：读取只基于元数据的统计。
- `POST /api/v1/admin/media/objects/{id}/lifecycle`：更新生命周期元数据。
- `POST /api/v1/admin/media/objects/{id}/verify`：只读校验单个对象。
- `POST /api/v1/admin/media/objects/verify`：按后端或内容哈希批量校验。

## OpenAPI

接口合同位于 `packages/api-contract/openapi/storage-admin.v1.json`，测试会校验路由、参数、认证、schema 和错误码覆盖。
