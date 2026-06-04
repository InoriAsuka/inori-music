# 阶段 22：中文文档与 README 拆分（v0.22.0）

## 需求快照

按照最新要求，Markdown 文档统一使用中文维护，并将 README 拆分为 `README.md` 入口和 `README_zh.md` 中文完整说明。

## 任务清单

- [x] 新增 `README_zh.md` 中文完整说明。
- [x] 将 `README.md` 调整为中文项目入口与导航。
- [x] 将 `requirement.md` 历史内容中文化。
- [x] 将 `docs/architecture` 与 `docs/adr` 下的 Markdown 文档中文化。
- [x] 将 `.plan/` 历史阶段计划中文化，并新增本阶段计划。
- [x] 更新版本到 `v0.22.0`。
- [x] 运行格式化、测试和差异检查。

## 非目标

- 不改变 Go API 运行时行为。
- 不修改 OpenAPI 路由语义。

## 后续候选

- 后续可增加英文 README 或多语言站点，但当前以中文文档为准。
