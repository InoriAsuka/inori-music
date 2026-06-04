# inori-music

这是项目的主入口文档。当前仓库文档采用中文维护，完整说明请阅读 [`README_zh.md`](README_zh.md)。

## 当前版本

当前架构基线版本：`0.22.0`。

## 快速导航

- [`README_zh.md`](README_zh.md)：中文完整项目说明。
- [`requirement.md`](requirement.md)：中文需求基线与版本历史。
- [`.plan/`](.plan/)：中文阶段计划。
- [`docs/architecture/`](docs/architecture/)：中文架构文档。
- [`docs/adr/`](docs/adr/)：中文架构决策记录。
- [`packages/api-contract/openapi/storage-admin.v1.json`](packages/api-contract/openapi/storage-admin.v1.json)：OpenAPI 3.1 管理接口合同。

## 运行示例

```bash
INORI_ADMIN_TOKEN=change-me-development-token go run ./services/api/cmd/server
```
