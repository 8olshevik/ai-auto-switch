# Implementation Plan: Code Switch Web 应用改造

## Overview

将现有 Wails 桌面应用改造为前后端分离的 Web 应用。后端使用 Go + Gin 暴露 REST API，前端使用 Vue 3 + TypeScript 通过 HTTP/WebSocket 与后端通信。按四个阶段递进实施：后端 API 化 → 前端改造 → 新功能 → 部署。

## Tasks

- [x] 1. Phase 1: 后端 API 化 — 项目结构与核心基础设施
  - [x] 1.1 创建新的服务入口 `cmd/server/main.go`
    - 创建 `cmd/server/` 目录
    - 实现 `main.go`：初始化配置、数据库、服务实例，启动 Gin HTTP Server
    - 移除对 `github.com/wailsapp/wails/v3` 的依赖引用
    - 加载环境变量配置（PORT、JWT_SECRET、ADMIN_USERNAME、ADMIN_PASSWORD 等）
    - _Requirements: 23.1, 23.2_

  - [x] 1.2 实现环境变量配置模块 `internal/config/config.go`
    - 定义 `AppConfig` 结构体（Port、ProxyPort、DatabasePath、JWTSecret、CORSOrigins、LogLevel 等）
    - 实现从环境变量读取配置，未设置时使用默认值
    - 验证必填字段（JWT_SECRET、ADMIN_PASSWORD）
    - _Requirements: 23.1, 23.2, 23.3, 23.4_

  - [x] 1.3 创建数据库迁移 — 新增 users、gateway_keys、assistant_conversations 表
    - 在现有数据库初始化逻辑中添加新表创建
    - 实现 `users` 表（id, username, password_hash, role, created_at）
    - 实现 `gateway_keys` 表（id, name, key_hash, key_prefix, rate_limit, enabled, created_at, last_used_at）
    - 实现 `assistant_conversations` 表（id, user_id, role, content, model, tokens_used, created_at）
    - 首次启动时创建默认管理员账户（从环境变量读取用户名密码）
    - _Requirements: 1.2, 1.9, 26.1, 27.12_

  - [x] 1.4 实现 JWT 认证中间件 `internal/api/middleware/auth.go`
    - 实现 JWT Token 生成（包含 user_id、username、role、过期时间）
    - 实现 JWT Token 验证中间件（从 Authorization header 提取并验证）
    - Token 过期或无效时返回 401 Unauthorized
    - 支持 Token 刷新机制（RefreshExpiry 7 天）
    - _Requirements: 1.5, 1.6, 1.8, 21.4_

  - [x] 1.5 实现 CORS 中间件 `internal/api/middleware/cors.go`
    - 从配置读取允许的 Origins 列表
    - 设置 Access-Control-Allow-Origin、Allow-Methods、Allow-Headers 响应头
    - 不在允许列表中的 Origin 返回 403 Forbidden
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [x] 1.6 实现 Gin 路由注册 `internal/api/router.go`
    - 创建 Gin Engine 实例，注册全局中间件（CORS、Logger、Recovery）
    - 注册 `/api/v1/` 路由组，应用 JWT 认证中间件
    - 注册公开路由（/auth/login、/health）不需要认证
    - 注册静态文件服务（前端 dist 目录）
    - _Requirements: 1.8, 11.1_

- [x] 2. Phase 1: 后端 API 化 — 认证与核心 Handler
  - [x] 2.1 实现认证 Handler `internal/api/handlers/auth.go`
    - POST `/auth/login`：验证用户名密码，返回 JWT Token
    - POST `/auth/logout`：清除会话数据
    - GET `/auth/me`：返回当前用户信息
    - 密码使用 bcrypt 哈希验证
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.7_

  - [x] 2.2 实现供应商管理 Handler `internal/api/handlers/providers.go`
    - GET `/providers/:kind`：获取供应商列表
    - POST `/providers/:kind`：保存供应商列表（加密 API Key）
    - POST `/providers/:kind/duplicate/:id`：复制供应商
    - PUT `/providers/:kind/:id/rename`：重命名供应商
    - POST `/providers/:kind/reorder`：调整排序
    - 验证必填字段，无效时返回 400 Bad Request
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 2.10_

  - [x] 2.3 实现代理管理 Handler `internal/api/handlers/proxy.go`
    - GET `/proxy/status`：获取代理状态
    - POST `/proxy/start`：启动代理服务
    - POST `/proxy/stop`：停止代理服务
    - GET `/proxy/last-used`：获取最后使用的供应商
    - 修改 `ProviderRelayService` 监听地址为 `0.0.0.0:18100`
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

  - [x] 2.4 实现设置 Handler `internal/api/handlers/settings.go`
    - GET `/settings/`：获取所有设置
    - PUT `/settings/`：更新设置（验证端口范围 1-65535）
    - GET `/settings/app`：获取应用设置
    - PUT `/settings/app`：更新应用设置
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

  - [x] 2.5 实现日志 Handler `internal/api/handlers/logs.go`
    - GET `/logs/`：获取日志列表（支持分页、关键词过滤、时间范围过滤）
    - GET `/logs/stats`：获取日志统计（请求总数、成功率、平均响应时间）
    - GET `/logs/heatmap`：获取热力图数据
    - DELETE `/logs/`：清空日志
    - _Requirements: 5.1, 5.3, 5.4, 5.5, 5.6, 5.7, 8.1, 8.2, 8.3, 8.4, 8.6_

  - [x] 2.6 实现 MCP 服务器管理 Handler `internal/api/handlers/mcp.go`
    - GET `/mcp/`：获取 MCP 服务器列表
    - POST `/mcp/`：添加 MCP 服务器
    - PUT `/mcp/:id`：更新 MCP 服务器
    - DELETE `/mcp/:id`：删除 MCP 服务器
    - 保存后同步配置到 Claude Code 和 Codex 配置文件
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

  - [x] 2.7 实现其他 Handler（CLI 配置、提示词、技能、健康检查、速度测试、导入导出、更新）
    - `internal/api/handlers/cli_config.go`：GET/PUT `/cli-config/:platform`
    - `internal/api/handlers/prompts.go`：CRUD `/prompts/`
    - `internal/api/handlers/skills.go`：GET/POST/DELETE `/skills/`
    - `internal/api/handlers/health.go`：GET/POST `/health/`
    - `internal/api/handlers/speedtest.go`：POST `/speedtest/run`
    - `internal/api/handlers/import.go`：POST `/import/config`、GET `/import/export`
    - `internal/api/handlers/update.go`：GET/POST `/update/`
    - `internal/api/handlers/blacklist.go`：GET/POST `/blacklist/`
    - _Requirements: 7.1-7.6, 12.1-12.6, 13.1-13.5, 14.1-14.3, 15.1-15.6, 25.1-25.4_

- [x] 3. Phase 1: 后端 API 化 — WebSocket 与事件推送
  - [x] 3.1 实现 WebSocket Hub `internal/ws/hub.go`
    - 实现 WebSocket 连接管理（注册、注销、广播）
    - 定义消息格式 `WSMessage{Type, Payload}`
    - 支持按事件类型订阅
    - 实现心跳检测和自动断开
    - _Requirements: 8.2, 9.6_

  - [x] 3.2 实现 WebSocket Handler 和事件集成
    - 注册 `/api/v1/ws/events` WebSocket 端点（需要 JWT 认证）
    - 替换所有 `app.EmitEvent()` 调用为 WebSocket 广播
    - 实现事件类型：proxy:status、request:log、health:result
    - _Requirements: 8.2, 9.4, 9.6, 13.3_

  - [x]* 3.3 编写后端 API 单元测试
    - 测试 JWT 认证中间件（有效/无效/过期 Token）
    - 测试认证 Handler（登录成功/失败）
    - 测试供应商 Handler（CRUD 操作、验证逻辑）
    - 测试 CORS 中间件（允许/拒绝 Origin）
    - _Requirements: 1.4, 1.5, 1.6, 11.2, 11.3_

- [x] 4. Checkpoint - Phase 1 完成验证
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Phase 2: 前端改造 — 通信层替换
  - [x] 5.1 创建 HTTP API 客户端 `src/api/client.ts`
    - 使用 axios 创建 API 实例，baseURL 从环境变量读取
    - 实现请求拦截器：自动附加 JWT Token 到 Authorization header
    - 实现响应拦截器：401 时清除 Token 并跳转登录页
    - 设置请求超时（30s）
    - _Requirements: 1.5, 19.4_

  - [x] 5.2 创建各模块 API 服务文件（替代 Wails bindings）
    - `src/api/auth.ts`：login、logout、getMe
    - `src/api/providers.ts`：load、save、duplicate、rename、reorder
    - `src/api/proxy.ts`：getStatus、start、stop、getLastUsed
    - `src/api/settings.ts`：get、update、getApp、updateApp
    - `src/api/logs.ts`：list、stats、heatmap、clear
    - `src/api/mcp.ts`：list、create、update、delete
    - `src/api/prompts.ts`：list、create、update、delete
    - `src/api/health.ts`：check、history
    - `src/api/import.ts`：importConfig、exportConfig
    - _Requirements: 2.1, 5.1, 6.1, 8.1, 9.1, 10.1, 12.1, 13.1_

  - [x] 5.3 创建 WebSocket 客户端 `src/api/websocket.ts`
    - 实现 WSClient 类（connect、on、off、reconnect）
    - 连接时附加 JWT Token 作为查询参数
    - 实现自动重连机制（3 秒间隔）
    - 实现事件监听和分发
    - _Requirements: 8.2, 9.6_

  - [x] 5.4 移除 `@wailsio/runtime` 依赖和 Wails bindings
    - 从 `package.json` 移除 `@wailsio/runtime`
    - 删除 `frontend/bindings/` 目录
    - 移除所有 `import` 中对 `@wailsio/runtime` 和 bindings 的引用
    - 移除窗口操作相关代码（SetSize、Center 等）
    - _Requirements: 设计 §3.6_

- [x] 6. Phase 2: 前端改造 — 页面与路由
  - [x] 6.1 实现登录页面 `src/pages/LoginPage.vue`
    - 用户名/密码输入表单
    - 调用 auth API 进行登录
    - 登录成功保存 Token 到 localStorage 并跳转主页
    - 登录失败显示错误提示
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 6.2 改造路由配置 — Hash 模式改为 History 模式
    - 使用 `createWebHistory` 替代 `createWebHashHistory`
    - 添加路由守卫：未登录时重定向到 `/login`
    - 添加 `/login` 路由（meta: { public: true }）
    - 添加 `/gateway` 和 `/assistant` 路由占位
    - 移除 `/tray` 路由
    - _Requirements: 1.1, 1.8_

  - [x] 6.3 替换所有页面组件中的 Wails binding 调用为 HTTP API 调用
    - 逐个修改现有页面组件，将 `$Call.ByID()` 替换为对应的 API 服务调用
    - 替换 Wails 事件监听为 WebSocket 事件监听
    - 确保所有数据加载、保存、删除操作使用新的 API 客户端
    - 处理异步错误和加载状态
    - _Requirements: 2.1, 5.1, 6.1, 8.1, 9.1, 10.1, 19.1, 19.2, 19.3_

  - [x] 6.4 实现前端错误处理与用户反馈机制
    - 操作成功显示 Toast 提示
    - 操作失败显示错误信息和具体原因
    - 长时间操作（>1s）显示加载指示器
    - 网络断开时显示连接错误提示和重试选项
    - 表单验证错误高亮显示
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5_

  - [ ]* 6.5 编写前端 API 客户端单元测试
    - 测试请求拦截器（Token 附加）
    - 测试响应拦截器（401 处理）
    - 测试 WebSocket 重连逻辑
    - _Requirements: 1.5, 1.6_

- [ ] 7. Checkpoint - Phase 2 完成验证
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Phase 3: 新功能 — API 网关
  - [x] 8.1 实现 API 网关后端 — Key 管理与速率限制
    - `internal/api/handlers/gateway.go`：
    - POST `/gateway/keys`：创建 API Key（生成随机 Key，存储哈希值和前缀）
    - GET `/gateway/keys`：获取 API Key 列表（显示前缀和名称）
    - DELETE `/gateway/keys/:id`：删除 API Key
    - GET `/gateway/stats`：获取网关使用统计（按来源应用过滤）
    - PUT `/gateway/rate-limit`：设置速率限制
    - _Requirements: 26.1, 26.7, 26.8, 26.9, 26.10_

  - [x] 8.2 实现速率限制中间件 `internal/api/middleware/ratelimit.go`
    - 基于 API Key 的令牌桶算法
    - 从 gateway_keys 表读取每个 Key 的速率限制配置
    - 超过限制返回 429 Too Many Requests
    - _Requirements: 26.9, 26.10_

  - [x] 8.3 实现代理端口 API Key 认证
    - 在代理服务 `:18100` 端口添加 API Key 验证
    - 支持 `Authorization: Bearer <key>` 和 `X-API-Key: <key>` 两种方式
    - 验证 Key 有效性和启用状态
    - 记录请求来源应用信息
    - _Requirements: 26.1, 26.2, 26.3, 26.4, 26.5, 26.6_

  - [x] 8.4 实现 API 网关前端页面 `src/pages/GatewayPage.vue`
    - API Key 列表展示（名称、前缀、创建时间、状态）
    - 创建/删除 API Key 操作
    - 速率限制配置界面
    - 使用统计图表（按来源应用分组）
    - _Requirements: 26.1, 26.8, 26.9_

  - [ ]* 8.5 编写 API 网关集成测试
    - 测试 API Key 创建和验证流程
    - 测试速率限制触发（429 响应）
    - 测试无效 Key 拒绝访问
    - _Requirements: 26.1, 26.9, 26.10_

- [x] 9. Phase 3: 新功能 — AI 助手
  - [x] 9.1 实现 AI 助手后端 `internal/api/handlers/assistant.go`
    - POST `/assistant/chat`：接收用户消息，转发到 AI 模型（通过 Proxy）
    - GET `/assistant/history`：获取对话历史
    - DELETE `/assistant/history`：清空对话历史
    - 实现流式响应（通过 WebSocket 推送 assistant:reply 事件）
    - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.12, 27.13_

  - [x] 9.2 实现 Function Calling 工具定义与执行引擎
    - 定义工具列表（list_providers、add_provider、toggle_provider、get_proxy_status、start_proxy、stop_proxy、get_stats、get_settings、update_settings）
    - 实现工具调用分发：解析 AI 模型返回的 function_call，执行对应服务方法
    - 敏感操作（修改 API Key、删除供应商）返回确认请求
    - POST `/assistant/execute`：用户确认后执行敏感操作
    - _Requirements: 27.5, 27.6, 27.7, 27.8, 27.9, 27.10, 27.11_

  - [x] 9.3 实现 AI 助手前端页面 `src/pages/AssistantPage.vue`
    - 聊天界面（消息列表、输入框、发送按钮）
    - 流式响应展示（逐字显示 AI 回复）
    - 敏感操作确认对话框
    - 对话历史查看和清空
    - _Requirements: 27.1, 27.4, 27.9, 27.10, 27.11, 27.12, 27.13_

  - [ ]* 9.4 编写 AI 助手单元测试
    - 测试 Function Calling 工具分发逻辑
    - 测试敏感操作检测
    - 测试对话历史存储和检索
    - _Requirements: 27.5, 27.6, 27.10_

- [x] 10. Checkpoint - Phase 3 完成验证
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 11. Phase 4: 部署与文档
  - [x] 11.1 编写 Dockerfile（多阶段构建）
    - Stage 1：Node.js 构建前端（pnpm install + pnpm build）
    - Stage 2：Go 构建后端（go mod download + go build）
    - Stage 3：Alpine 运行镜像（复制二进制文件，暴露端口 8080、18100）
    - _Requirements: 22.1, 22.2, 22.3_

  - [x] 11.2 编写 docker-compose.yml
    - 定义 codeswitch 服务（build、ports、environment、volumes）
    - 配置环境变量（JWT_SECRET、ADMIN_USERNAME、ADMIN_PASSWORD、CORS_ORIGINS 等）
    - 配置数据持久化卷
    - 设置 restart: unless-stopped
    - _Requirements: 22.2, 22.3, 22.4, 23.4_

  - [x] 11.3 实现后端静态文件服务和 History 模式支持
    - Gin 配置静态文件服务（serve frontend/dist）
    - 所有非 `/api/` 路径 fallback 到 `index.html`（支持 Vue Router History 模式）
    - 容器启动时自动创建必要目录（数据库、日志、备份）
    - _Requirements: 22.2, 22.3_

  - [x]* 11.4 编写部署文档 `doc/deployment.md`
    - Docker 部署步骤
    - 环境变量说明
    - Nginx 反向代理配置示例（可选）
    - 数据备份与恢复说明
    - _Requirements: 22.1, 23.4_

- [x] 12. Final Checkpoint - 全部完成验证
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- 后端使用 Go + Gin，前端使用 TypeScript + Vue 3
- 现有服务代码（services/ 目录）最大化复用，仅替换 Wails 绑定层
- 数据库保持 SQLite，新增表通过迁移脚本创建
- Phase 1-2 为核心改造，Phase 3-4 为功能扩展和部署

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["1.3", "1.4", "1.5"] },
    { "id": 2, "tasks": ["1.6"] },
    { "id": 3, "tasks": ["2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "3.1"] },
    { "id": 4, "tasks": ["3.2", "3.3"] },
    { "id": 5, "tasks": ["5.1", "5.4"] },
    { "id": 6, "tasks": ["5.2", "5.3"] },
    { "id": 7, "tasks": ["6.1", "6.2"] },
    { "id": 8, "tasks": ["6.3", "6.4", "6.5"] },
    { "id": 9, "tasks": ["8.1", "8.2"] },
    { "id": 10, "tasks": ["8.3", "8.4", "8.5"] },
    { "id": 11, "tasks": ["9.1", "9.2"] },
    { "id": 12, "tasks": ["9.3", "9.4"] },
    { "id": 13, "tasks": ["11.1", "11.2", "11.3"] },
    { "id": 14, "tasks": ["11.4"] }
  ]
}
```
