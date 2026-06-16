---
name: phase-release-process
description: Every phase must end with a full release: VERSION bump, requirement.md history entry, OpenAPI version sync, commit, push, git tag.
metadata:
  type: feedback
---

每个 phase 完成后必须执行完整发布流程，缺少任何一步都不算完成：

1. `VERSION` 文件更新为新语义版本
2. `requirement.md` 的 `## Current Version` 更新，并在 `## Requirement History` 末尾追加新版本条目
3. `packages/api-contract/openapi/storage-admin.v1.json` 的 `info.version` 同步更新
4. `git commit`（包含所有变更）
5. `git tag vX.Y.Z`
6. `git push origin main --tags`

**Why:** 用户明确要求每个 phase 都要走完整发布流程，版本号、文档、contract 必须三者同步，tag 必须推到远端。

**How to apply:** phase 代码完成、测试全绿后，立即执行上述 6 步，不要等用户提醒。
