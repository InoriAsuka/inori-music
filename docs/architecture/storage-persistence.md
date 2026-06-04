# 存储持久化策略

## 当前阶段

0.x 早期使用内存仓储和可选 JSON 文件仓储。文件仓储面向开发与单节点自托管，采用临时文件、同步和原子重命名保存状态。

## 环境变量

- `INORI_STORAGE_REPOSITORY_FILE`：启用存储后端 JSON 文件仓储。
- `INORI_MEDIA_OBJECT_REPOSITORY_FILE`：启用媒体对象 JSON 文件仓储。

## 后续方向

生产级元数据持久化将迁移到 PostgreSQL。领域服务应避免与具体 SQL 实现强耦合，以便后续补充迁移、索引、事务和审计日志。
