# AGENTS.md

This file provides guidance to Codex when working with code in this repository.

## 项目概述

**Daytona Lite** 是 [Daytona](https://github.com/daytonaio/daytona) 的精简私有化部署版本，保留核心 Sandbox 管理能力，移除了 SaaS 相关组件，并将认证方式改为 **Admin Password + API Key** 双模式。

与上游 Daytona 的主要差异：

- 移除了 OIDC / Dex，改为管理员密码登录和 API Key 鉴权
- 移除了计费、审计日志、Webhook、邮件、实时通知、CLI、文档站点等 SaaS 能力
- 保留 Dashboard、API、Runner、Proxy、SSH Gateway、Snapshot Manager 等核心执行链路

## 仓库结构

项目使用 **Nx monorepo** 管理，包管理器为 **yarn**。

```text
apps/
  api/               NestJS REST API
  dashboard/         React + Vite 管理界面
  proxy/             Sandbox 端口代理（Go）
  runner/            Sandbox 生命周期管理（Go）
  daemon/            容器运行时（Go）
  snapshot-manager/  镜像快照管理（Go）
  ssh-gateway/       SSH 接入网关
libs/
  api-client/                     TypeScript API Client（生成产物）
  api-client-python/              Python API Client（生成产物）
  api-client-python-async/        Python Async API Client（生成产物）
  sdk-typescript/                 TypeScript SDK
  sdk-python/                     Python SDK
  runner-api-client/              Runner API Client
  toolbox-api-client/             Toolbox API Client（生成产物）
  toolbox-api-client-python/      Python Toolbox API Client（生成产物）
  toolbox-api-client-python-async/ Python Async Toolbox API Client（生成产物）
  common-go/                      Go 公共库
docker/            Docker Compose 与部署配置
docs/              项目文档
scripts/           本地开发辅助脚本
```

## 推荐工作方式

优先使用仓库内封装好的脚本，而不是手写长命令：

```bash
yarn dev:doctor       # 检查本地开发环境
yarn dev:start        # 启动开发依赖（Postgres / Redis / MinIO / Registry）
yarn dev:api          # 本机启动 API（热重载）
yarn dev:dashboard    # 本机启动 Dashboard
yarn dev:full         # 一键启动开发环境
yarn dev:stop         # 停止开发依赖
```

补充说明：

- 开发依赖默认来自 `docker/docker-compose.dev.yml`
- `scripts/dev.sh` 会在缺少 `apps/api/.env` 时自动从 `.env.example` 生成
- 本机开发 API 时，脚本会自动将容器环境里的 `db`、`redis`、`minio` 等地址改写为 `localhost`
- Dashboard 默认运行在 `http://localhost:3000`
- API 本机开发端口来自 `apps/api/.env` 的 `PORT`，默认是 `3001`

## GitHub 提交安全规则

- 当前维护仓库是 `origin = https://github.com/shabooboo006/daytona-lite.git`
- `upstream = https://github.com/daytonaio/daytona.git` 仅作为上游参考仓库，**绝对不要向 upstream push**
- 当前 `origin` 仓库在 GitHub 上是 **fork**；在 GitHub 网页打开 PR 时，base repository 可能会默认切到父仓库 `daytonaio/daytona`
- 任何 push 或 PR 操作前都必须先确认目标仓库仍然是 `shabooboo006/daytona-lite`
- push 前必须执行并核对：

```bash
git remote get-url origin
git remote get-url --push origin
```

- 只有输出仍为 `https://github.com/shabooboo006/daytona-lite.git` 时，才允许执行 `git push`
- 如果使用 GitHub CLI，所有 PR / repo 操作都应显式指定：

```bash
--repo shabooboo006/daytona-lite
```

- 如果在 GitHub 页面看到 PR base 指向 `daytonaio/daytona`，必须先改回 `shabooboo006/daytona-lite`，再继续提交

## 常用命令

### 安装依赖

```bash
yarn install
```

### 构建

```bash
yarn build
yarn build:production
npx nx build api
```

### 代码检查

```bash
yarn lint:ts
yarn lint:py
yarn lint
yarn lint:fix
```

### 测试

```bash
npx nx run-many --target=test --all
npx nx test api
npx nx test api --testFile=apps/api/src/sandbox/services/sandbox.service.spec.ts
```

### 数据库迁移

```bash
yarn migration:generate
yarn migration:run:pre-deploy
yarn migration:run:post-deploy
yarn migration:revert
```

### 生成客户端

```bash
yarn generate:api-client
```

## 架构说明

### API（`apps/api`）

基于 **NestJS**，核心模块包括：

- `AuthModule`：管理员 JWT + API Key 双模式认证
- `SandboxModule`：Sandbox 生命周期管理
- `OrganizationModule`：组织与配额管理
- `AdminModule`：管理员接口
- `RegionModule`：区域管理
- `ApiKeyModule`：API Key CRUD 与校验
- `ObjectStorageModule`：S3 兼容对象存储
- `UsageModule`：资源使用统计

基础设施依赖：

- PostgreSQL：TypeORM 持久化
- Redis：缓存与部分鉴权结果缓存
- MinIO：S3 兼容对象存储

迁移目录位于：

- `apps/api/src/migrations/pre-deploy/`
- `apps/api/src/migrations/post-deploy/`

### Dashboard（`apps/dashboard`）

基于 **React + Vite**，主要技术栈：

- `react-router-dom` 6
- `@tanstack/react-query`
- `shadcn/ui` + `Radix UI` + `Tailwind CSS`
- `@daytonaio/api-client`
- `msw`（开发环境 mock）

认证流程：

- Dashboard 调用 `POST /api/admin/login` 获取 JWT
- 后续请求走 `CombinedAuthGuard`
- 启动时通过 `/api/config` 加载运行时配置

### Lite 版本认证机制

- 管理员登录：`POST /api/admin/login`
- 管理员密码来自 `ADMIN_PASSWORD`
- JWT secret 使用 `JWT_SECRET` 或 `ENCRYPTION_KEY`
- API 调用使用 `Authorization: Bearer <api-key>`
- `CombinedAuthGuard` 同时支持 `admin-jwt` 与 `api-key`

## API 与代码约定

### HTTP / 方法命名

- HTTP `POST` 对应 `Create` / `Update`
- HTTP `DELETE` 对应 `Delete`
- HTTP `PUT` 对应 `Save`
- HTTP `GET` 对应 `Find` / `List`
- Service 方法名里不要重复模型名，例如用 `Create`，不要用 `CreateSandbox`

### 修改范围建议

- 业务逻辑优先在 `apps/api/src` 或 `apps/dashboard/src` 中修改
- 生成客户端优先通过生成流程更新，不要直接手改生成文件
- 当 OpenAPI 变化时，优先执行 `yarn generate:api-client`

通常不要直接手改以下目录中的生成产物，除非任务明确要求：

- `libs/api-client/`
- `libs/api-client-python/`
- `libs/api-client-python-async/`
- `libs/toolbox-api-client/`
- `libs/toolbox-api-client-python/`
- `libs/toolbox-api-client-python-async/`

## 关键环境变量

优先查看 `apps/api/src/config/configuration.ts` 作为权威配置来源。

重点变量：

- `ADMIN_PASSWORD`：管理员登录密码，必须修改
- `ENCRYPTION_KEY`：敏感数据加密密钥，同时可作为 JWT secret
- `ENCRYPTION_SALT`：加密盐值
- `SERVER_IP`：影响 Dashboard 与 Proxy 对外访问地址
- `PROXY_API_KEY`：Proxy 内部通信密钥
- `DEFAULT_SNAPSHOT`：默认 Sandbox 镜像
- `DEFAULT_RUNNER_*`：Runner 连接配置

生产部署配置主要看：

- `docker/docker-compose.yaml`

本地开发依赖配置主要看：

- `docker/docker-compose.dev.yml`

## 文档规范

项目文档统一放在 `docs/` 目录。

编写要求：

- 使用 ATX 风格标题
- 代码块必须标注语言
- 多字段对比优先用表格
- 内部链接使用相对路径
- 技术文档应尽量包含 Mermaid 图表辅助理解

新增文档时：

1. 在 `docs/` 合适子目录创建文档
2. 尽量包含至少一个 Mermaid 图表
3. 视情况更新 `README.md`
4. 如果开发流程或项目约束发生变化，同步更新 `CLAUDE.md` 与本文件

## 给 Codex 的额外建议

- 搜索优先使用 `rg`
- 优先运行最小范围的 `nx` 构建、测试、lint，而不是全量任务
- 修改 API 相关代码后，优先验证受影响模块的测试或 lint
- 修改 Dashboard 相关代码后，至少检查 TypeScript / lint 是否通过
- 如果发现仓库里同时存在用户的未提交改动，不要覆盖或回退它们
- 涉及鉴权、配置、端口时，优先核对脚本与配置文件的实际值，不要只依赖 README 描述
