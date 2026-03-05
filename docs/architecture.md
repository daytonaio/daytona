# 系统架构

Daytona Lite 是一个精简的 AI Agent 执行环境，提供安全隔离的 Sandbox 容器运行能力。

## 整体服务拓扑

```mermaid
graph TD
  subgraph external[外部访问]
    user[用户 / SDK / API Client]
  end

  subgraph core[核心服务层]
    api[API Server\nNestJS :3000]
    proxy[Proxy\n:4000]
    dashboard[Dashboard\nReact SPA]
  end

  subgraph execution[执行层]
    runner[Runner\n:3003]
    daemon[Daemon\n容器内运行时]
    ssh[SSH Gateway\n:2222]
    snapshot[Snapshot Manager]
  end

  subgraph storage[存储层]
    pg[(PostgreSQL\n元数据)]
    redis[(Redis\n缓存 / 队列)]
    minio[(MinIO\nS3 兼容存储)]
    registry[Container Registry\n:6000]
  end

  user -->|HTTP REST| api
  user -->|SSH| ssh
  user -->|端口代理| proxy
  user -->|浏览器| dashboard
  dashboard -->|HTTP REST| api
  api --> pg & redis & minio & registry
  api -->|创建 / 管理 Sandbox| runner
  runner -->|启动容器| daemon
  runner --> registry
  proxy -->|转发端口| runner
  ssh -->|SSH 隧道| runner
  snapshot -->|推送镜像| registry
  snapshot --> minio
```

## 认证流程

Daytona Lite 采用双模式认证，无需 OIDC/Dex 容器。

```mermaid
sequenceDiagram
  participant client as 客户端
  participant guard as CombinedAuthGuard
  participant admin as AdminAuthStrategy
  participant apikey as ApiKeyStrategy
  participant redis as Redis 缓存

  alt Dashboard 登录场景
    client->>guard: POST /api/admin/login\n{password: "..."}
    guard->>admin: 验证密码
    admin-->>guard: 校验 ADMIN_PASSWORD 环境变量
    guard-->>client: 颁发 JWT Token（64h 有效期）
    client->>guard: 后续请求 Authorization: Bearer <jwt>
    guard->>admin: 验证 JWT 签名
    admin-->>guard: 解码用户信息
    guard-->>client: 通过鉴权
  end

  alt SDK / 程序化调用场景
    client->>guard: 请求 Authorization: Bearer <api-key>
    guard->>apikey: 验证 API Key
    apikey->>redis: 查询缓存
    alt 缓存命中
      redis-->>apikey: 返回组织信息
    else 缓存未命中
      apikey->>apikey: 查询数据库
      apikey->>redis: 写入缓存
    end
    apikey-->>guard: 验证通过
    guard-->>client: 通过鉴权
  end
```

## Sandbox 生命周期

```mermaid
stateDiagram-v2
  [*] --> Creating : POST /api/sandboxes
  Creating --> Started : Runner 启动容器成功
  Creating --> Error : 启动失败

  Started --> Stopping : POST /sandboxes/:id/stop
  Stopping --> Stopped : 容器停止成功

  Stopped --> Started : POST /sandboxes/:id/start
  Stopped --> Archiving : 闲置超时触发归档

  Archiving --> Archived : 快照保存完成
  Archived --> Started : POST /sandboxes/:id/start（从快照恢复）

  Started --> Destroying : DELETE /sandboxes/:id
  Stopped --> Destroying : DELETE /sandboxes/:id
  Archived --> Destroying : DELETE /sandboxes/:id
  Destroying --> [*] : 容器和数据清理完成

  Error --> Destroying : 手动清理
```

## 各服务职责

### API Server（`apps/api`）

基于 **NestJS** 构建的 REST API 服务，是整个系统的控制面。

| 模块 | 职责 |
|------|------|
| `AuthModule` | 双模式认证（Admin JWT + API Key），`CombinedAuthGuard` |
| `SandboxModule` | Sandbox 完整生命周期管理，Manager/Action/Runner 分层架构 |
| `OrganizationModule` | 组织管理、资源配额、权限隔离 |
| `AdminModule` | 管理员接口（Runner 注册、全局 Sandbox 管理） |
| `ApiKeyModule` | API Key CRUD，Redis 缓存加速验证 |
| `ObjectStorageModule` | MinIO S3 兼容存储操作 |
| `RegionModule` | 多区域部署支持 |

### Runner（`apps/runner`）

Sandbox 生命周期的执行层，运行在 Linux 宿主机上（需要特权模式）。

- 接收 API 下发的 Sandbox 创建/启动/停止指令
- 通过 Docker API 管理 Sandbox 容器
- 启动 Daemon 进程作为容器内运行时

### Daemon（`apps/daemon`）

运行在 Sandbox 容器内部的代理进程，提供 Toolbox API：

- **文件系统**：读写文件、目录列表
- **进程执行**：命令执行、代码运行、PTY 终端
- **Git 操作**：Clone、Commit、Branch
- **LSP 支持**：代码补全、定义跳转

### Proxy（`apps/proxy`）

将外部请求转发到 Sandbox 内部端口，支持 HTTP 和 WebSocket。访问格式：`http://<sandbox-id>-<port>.proxy.<server-ip>.nip.io`

### SSH Gateway（`apps/ssh-gateway`）

提供标准 SSH 协议接入，将 SSH 连接路由到对应的 Sandbox 容器。

## 数据流示意

```mermaid
flowchart LR
  sdk[SDK / API Client]

  sdk -->|1. 认证| api[API Server]
  api -->|2. 写入 Sandbox 记录| pg[(PostgreSQL)]
  api -->|3. 下发创建指令| runner[Runner]
  runner -->|4. 拉取镜像| registry[Registry]
  runner -->|5. 启动容器| daemon[Daemon 容器]
  daemon -->|6. 就绪回调| runner
  runner -->|7. 更新状态| api
  api -->|8. 返回 Sandbox 信息| sdk
  sdk -->|9. 执行代码| api
  api -->|10. 转发 Toolbox 请求| daemon
  daemon -->|11. 返回执行结果| api
  api -->|12. 返回结果| sdk
```

## 部署拓扑

### 单机部署（推荐入门）

所有服务运行在同一台 Linux 服务器上，通过 Docker Compose 编排。适合评估和小规模使用。

### 分离部署（生产推荐）

```mermaid
graph LR
  subgraph server1[控制面服务器]
    api[API Server]
    dashboard[Dashboard]
    pg[(PostgreSQL)]
    redis[(Redis)]
    minio[(MinIO)]
  end

  subgraph server2[执行面服务器]
    runner[Runner]
    registry[Registry]
    proxy[Proxy]
    ssh[SSH Gateway]
  end

  api -->|gRPC / HTTP| runner
  runner --> registry
```

将控制面（API、数据库）与执行面（Runner）分离，便于独立扩容执行节点。
