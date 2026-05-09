# Code Switch Web 应用改造 - 技术设计文档

## 1. 架构概览

### 1.1 当前架构（Wails 桌面应用）

```
┌─────────────────────────────────────────────┐
│              Wails 桌面应用                    │
│  ┌─────────────┐    ┌────────────────────┐  │
│  │  Vue 3 前端  │◄──►│  Go 后端 (RPC绑定)  │  │
│  │  (WebView)  │    │  25+ 服务          │  │
│  └─────────────┘    └────────────────────┘  │
│                            │                 │
│                     ┌──────▼──────┐          │
│                     │ Gin 代理服务  │          │
│                     │  (:18100)   │          │
│                     └─────────────┘          │
└─────────────────────────────────────────────┘
```

### 1.2 目标架构（前后端分离 Web 应用）

```
┌──────────────────┐         ┌──────────────────────────────────┐
│   浏览器 (前端)    │  HTTP   │        Linux 服务器 (后端)         │
│                  │◄───────►│                                  │
│  Vue 3 + Vite   │  /api/* │  ┌────────────────────────────┐  │
│  (SPA 静态部署)   │         │  │   Go HTTP Server (Gin)     │  │
│                  │  WS     │  │   - REST API (:8080)       │  │
│                  │◄───────►│  │   - WebSocket (事件推送)     │  │
└──────────────────┘         │  │   - 静态文件服务 (前端)      │  │
                             │  └────────────────────────────┘  │
┌──────────────────┐         │              │                   │
│  第三方应用        │  HTTP   │  ┌───────────▼────────────────┐  │
│  (API 网关客户端)  │◄───────►│  │   Proxy 代理服务 (:18100)   │  │
└──────────────────┘         │  │   - 智能降级/故障转移        │  │
                             │  │   - 模型映射                 │  │
                             │  │   - 用量统计                 │  │
                             │  └────────────────────────────┘  │
                             │              │                   │
                             │  ┌───────────▼────────────────┐  │
                             │  │   SQLite 数据库              │  │
                             │  │   ~/.code-switch/app.db     │  │
                             │  └────────────────────────────┘  │
                             └──────────────────────────────────┘
```

## 2. 后端设计

### 2.1 服务层改造策略

**核心原则**：最大化复用现有 Go 服务代码，仅替换 Wails RPC 绑定层为 REST API。

现有服务通过 Wails 的 `application.NewService()` 注册，前端通过 `$Call.ByID()` 调用。改造后：
- 移除 `github.com/wailsapp/wails/v3` 依赖
- 为每个服务创建对应的 HTTP Handler（Gin router group）
- 事件推送从 `app.EmitEvent()` 改为 WebSocket

### 2.2 API 路由设计

```
/api/v1/
├── auth/
│   ├── POST   /login              # 用户登录
│   ├── POST   /logout             # 用户登出
│   └── GET    /me                 # 获取当前用户信息
│
├── providers/
│   ├── GET    /:kind              # 获取供应商列表 (kind: claude/codex/gemini)
│   ├── POST   /:kind              # 保存供应商列表
│   ├── POST   /:kind/duplicate/:id # 复制供应商
│   ├── PUT    /:kind/:id/rename   # 重命名供应商
│   └── POST   /:kind/reorder     # 调整排序
│
├── proxy/
│   ├── GET    /status             # 代理状态
│   ├── POST   /start              # 启动代理
│   ├── POST   /stop               # 停止代理
│   └── GET    /last-used          # 最后使用的供应商
│
├── settings/
│   ├── GET    /                   # 获取所有设置
│   ├── PUT    /                   # 更新设置
│   ├── GET    /app                # 应用设置
│   └── PUT    /app                # 更新应用设置
│
├── blacklist/
│   ├── GET    /                   # 获取黑名单
│   ├── POST   /recover/:id       # 手动恢复
│   └── GET    /settings           # 黑名单设置
│
├── logs/
│   ├── GET    /                   # 获取日志列表 (支持分页/过滤)
│   ├── GET    /stats              # 日志统计
│   ├── GET    /heatmap            # 热力图数据
│   └── DELETE /                   # 清空日志
│
├── mcp/
│   ├── GET    /                   # 获取 MCP 服务器列表
│   ├── POST   /                   # 添加 MCP 服务器
│   ├── PUT    /:id                # 更新 MCP 服务器
│   └── DELETE /:id                # 删除 MCP 服务器
│
├── cli-config/
│   ├── GET    /:platform          # 获取 CLI 配置
│   └── PUT    /:platform          # 更新 CLI 配置
│
├── prompts/
│   ├── GET    /                   # 获取提示词列表
│   ├── POST   /                   # 添加提示词
│   ├── PUT    /:id                # 更新提示词
│   └── DELETE /:id                # 删除提示词
│
├── skills/
│   ├── GET    /                   # 获取技能列表
│   ├── POST   /install            # 安装技能
│   └── DELETE /:id                # 删除技能
│
├── health/
│   ├── GET    /                   # 系统健康检查
│   ├── POST   /check              # 执行供应商健康检查
│   └── GET    /history            # 健康检查历史
│
├── speedtest/
│   └── POST   /run                # 执行速度测试
│
├── import/
│   ├── POST   /config             # 导入配置
│   └── GET    /export             # 导出配置
│
├── gateway/                        # 新增：API 网关
│   ├── POST   /keys               # 创建 API Key
│   ├── GET    /keys               # 获取 API Key 列表
│   ├── DELETE /keys/:id           # 删除 API Key
│   ├── GET    /stats              # 网关使用统计
│   └── PUT    /rate-limit         # 设置速率限制
│
├── assistant/                      # 新增：AI 助手
│   ├── POST   /chat               # 发送消息
│   ├── GET    /history            # 获取对话历史
│   ├── DELETE /history            # 清空对话历史
│   └── POST   /execute            # 执行配置操作
│
├── update/
│   ├── GET    /check              # 检查更新
│   └── POST   /install            # 安装更新
│
└── ws/                            # WebSocket
    └── /events                    # 实时事件推送
```

### 2.3 API 网关端点（对外暴露）

```
# 代理端口 :18100 保持不变，新增 API Key 认证
# 第三方应用通过此端口访问 AI 模型

POST /v1/messages          # Anthropic 格式
POST /v1/chat/completions  # OpenAI 格式
POST /v1/responses         # Codex 格式

# 请求头：
# Authorization: Bearer <gateway-api-key>
# 或
# X-API-Key: <gateway-api-key>
```

### 2.4 认证设计

```go
// JWT 认证中间件
type AuthConfig struct {
    SecretKey     string        // JWT 签名密钥（环境变量 JWT_SECRET）
    TokenExpiry   time.Duration // Token 有效期（默认 24h）
    RefreshExpiry time.Duration // 刷新 Token 有效期（默认 7d）
}

// 用户存储（SQLite）
// 初始化时创建默认管理员账户（用户名/密码通过环境变量配置）
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'admin',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 2.5 WebSocket 事件推送

替代 Wails 的 `app.EmitEvent()`，用于实时推送：
- 代理状态变化
- 请求日志实时更新
- 健康检查结果
- AI 助手流式响应

```go
// WebSocket 消息格式
type WSMessage struct {
    Type    string      `json:"type"`    // event_type
    Payload interface{} `json:"payload"` // 事件数据
}

// 事件类型
const (
    WSEventProxyStatus    = "proxy:status"
    WSEventRequestLog     = "request:log"
    WSEventHealthCheck    = "health:result"
    WSEventAssistantReply = "assistant:reply"
)
```

### 2.6 数据库改造

**保持 SQLite**，适合单服务器部署场景。新增表：

```sql
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'admin',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- API 网关密钥表
CREATE TABLE IF NOT EXISTS gateway_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    key_hash TEXT UNIQUE NOT NULL,
    key_prefix TEXT NOT NULL,        -- 前8位，用于显示
    rate_limit INTEGER DEFAULT 60,   -- 每分钟请求数
    enabled BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME
);

-- AI 助手对话历史表
CREATE TABLE IF NOT EXISTS assistant_conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL,               -- user/assistant/system
    content TEXT NOT NULL,
    model TEXT,
    tokens_used INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 2.7 配置管理

```go
// 应用配置（环境变量 + 配置文件）
type AppConfig struct {
    // 服务器
    Port           int    `env:"PORT" default:"8080"`
    ProxyPort      int    `env:"PROXY_PORT" default:"18100"`
    ProxyListenAddr string `env:"PROXY_LISTEN_ADDR" default:"0.0.0.0"` // Web 版改为监听所有接口
    
    // 数据库
    DatabasePath   string `env:"DATABASE_PATH" default:"~/.code-switch/app.db"`
    
    // 认证
    JWTSecret      string `env:"JWT_SECRET" required:"true"`
    AdminUsername  string `env:"ADMIN_USERNAME" default:"admin"`
    AdminPassword  string `env:"ADMIN_PASSWORD" required:"true"` // 首次启动时创建
    
    // CORS
    CORSOrigins    string `env:"CORS_ORIGINS" default:"*"` // 生产环境应限制
    
    // 日志
    LogLevel       string `env:"LOG_LEVEL" default:"info"`
    LogDir         string `env:"LOG_DIR" default:"~/.code-switch/logs"`
    
    // AI 助手
    AssistantModel string `env:"ASSISTANT_MODEL" default:"claude-sonnet-4-20250514"`
}
```

## 3. 前端设计

### 3.1 改造策略

**核心变更**：将 Wails RPC 调用替换为 HTTP API 调用。

```
当前：前端 → @wailsio/runtime Call.ByID() → Go 方法
目标：前端 → axios/fetch → REST API → Go Handler → Go 服务方法
```

### 3.2 通信层替换

创建统一的 API 客户端层，替代 Wails bindings：

```typescript
// src/api/client.ts
import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
})

// 请求拦截器：自动附加 JWT Token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：处理 401 自动跳转登录
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api
```

```typescript
// src/api/providers.ts — 替代 bindings/codeswitch/services/providerservice.ts
import api from './client'
import type { Provider } from '@/types'

export const providerApi = {
  load: (kind: string) => api.get<Provider[]>(`/providers/${kind}`),
  save: (kind: string, providers: Provider[]) => api.post(`/providers/${kind}`, providers),
  duplicate: (kind: string, id: number) => api.post<Provider>(`/providers/${kind}/duplicate/${id}`),
  rename: (kind: string, id: number, newName: string) => api.put(`/providers/${kind}/${id}/rename`, { newName }),
}
```

### 3.3 WebSocket 集成

```typescript
// src/api/websocket.ts
class WSClient {
  private ws: WebSocket | null = null
  private listeners: Map<string, Set<Function>> = new Map()

  connect() {
    const token = localStorage.getItem('token')
    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/api/v1/ws/events?token=${token}`
    this.ws = new WebSocket(wsUrl)
    
    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      this.emit(msg.type, msg.payload)
    }
    
    this.ws.onclose = () => {
      setTimeout(() => this.connect(), 3000) // 自动重连
    }
  }

  on(event: string, callback: Function) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  private emit(event: string, data: any) {
    this.listeners.get(event)?.forEach(cb => cb(data))
  }
}

export const wsClient = new WSClient()
```

### 3.4 新增页面

| 页面 | 路由 | 说明 |
|------|------|------|
| 登录页 | `/login` | JWT 认证登录 |
| API 网关管理 | `/gateway` | API Key 管理、速率限制、使用统计 |
| AI 助手 | `/assistant` | 聊天界面、配置操作 |

### 3.5 路由改造

```typescript
// 从 Hash 模式改为 History 模式（更适合 Web 应用）
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/login', component: LoginPage, meta: { public: true } },
  { path: '/', component: MainPage },
  { path: '/logs', component: LogsPage },
  { path: '/availability', component: AvailabilityPage },
  { path: '/settings', component: GeneralPage },
  { path: '/mcp', component: McpPage },
  { path: '/prompts', component: PromptsPage },
  { path: '/skill', component: SkillPage },
  { path: '/speedtest', component: SpeedTestPage },
  { path: '/env', component: EnvCheckPage },
  { path: '/console', component: ConsolePage },
  { path: '/gateway', component: GatewayPage },      // 新增
  { path: '/assistant', component: AssistantPage },   // 新增
]

// 路由守卫：未登录跳转到登录页
router.beforeEach((to) => {
  if (!to.meta.public && !localStorage.getItem('token')) {
    return '/login'
  }
})
```

### 3.6 移除桌面相关功能

| 移除项 | 原因 |
|--------|------|
| `@wailsio/runtime` 依赖 | Wails RPC 不再需要 |
| Tray 页面 (`/tray`) | Web 应用无系统托盘 |
| 窗口操作 (SetSize, Center) | 浏览器无窗口控制 |
| 深度链接 (`ccswitch://`) | 改为 URL 参数导入 |
| 自动更新 UI | 服务端更新，无需前端参与 |

## 4. AI 助手设计（需求 27）

### 4.1 Function Calling 架构

```
用户输入 → Backend → AI 模型 (with tools) → Function Call → 执行操作 → 返回结果 → AI 总结 → 用户
```

### 4.2 可用工具定义

```json
{
  "tools": [
    {
      "name": "list_providers",
      "description": "列出指定平台的所有供应商配置",
      "parameters": { "kind": "string (claude/codex/gemini)" }
    },
    {
      "name": "add_provider",
      "description": "添加新的供应商",
      "parameters": { "kind": "string", "name": "string", "apiUrl": "string", "apiKey": "string", "level": "number" }
    },
    {
      "name": "toggle_provider",
      "description": "启用或禁用供应商",
      "parameters": { "kind": "string", "id": "number", "enabled": "boolean" }
    },
    {
      "name": "get_proxy_status",
      "description": "获取代理服务状态",
      "parameters": {}
    },
    {
      "name": "start_proxy",
      "description": "启动代理服务",
      "parameters": {}
    },
    {
      "name": "stop_proxy",
      "description": "停止代理服务",
      "parameters": {}
    },
    {
      "name": "get_stats",
      "description": "获取用量统计",
      "parameters": { "days": "number (默认7)" }
    },
    {
      "name": "get_settings",
      "description": "获取系统设置",
      "parameters": {}
    },
    {
      "name": "update_settings",
      "description": "更新系统设置",
      "parameters": { "key": "string", "value": "any" }
    }
  ]
}
```

### 4.3 安全控制

敏感操作（修改 API Key、删除供应商等）需要二次确认：
1. AI 模型返回 function call
2. Backend 检测到敏感操作
3. 返回确认请求给前端
4. 用户确认后执行

## 5. 部署设计

### 5.1 Docker 部署

```dockerfile
# 多阶段构建
# Stage 1: 构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install
COPY frontend/ ./
RUN pnpm build

# Stage 2: 构建后端
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/codeswitch ./cmd/server

# Stage 3: 运行
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/codeswitch .
EXPOSE 8080 18100
ENTRYPOINT ["./codeswitch"]
```

### 5.2 Docker Compose

```yaml
version: '3.8'
services:
  codeswitch:
    build: .
    ports:
      - "8080:8080"    # Web UI + API
      - "18100:18100"  # 代理服务
    environment:
      - JWT_SECRET=your-secret-key-here
      - ADMIN_USERNAME=admin
      - ADMIN_PASSWORD=changeme
      - CORS_ORIGINS=http://localhost:8080
      - LOG_LEVEL=info
      - PROXY_LISTEN_ADDR=0.0.0.0
    volumes:
      - codeswitch-data:/root/.code-switch
    restart: unless-stopped

volumes:
  codeswitch-data:
```

### 5.3 项目目录结构（改造后）

```
code-switch-web/
├── cmd/
│   └── server/
│       └── main.go              # 新入口（无 Wails 依赖）
├── internal/
│   ├── api/
│   │   ├── router.go           # Gin 路由注册
│   │   ├── middleware/
│   │   │   ├── auth.go         # JWT 认证中间件
│   │   │   ├── cors.go         # CORS 中间件
│   │   │   └── ratelimit.go    # 速率限制中间件
│   │   └── handlers/
│   │       ├── auth.go         # 认证 Handler
│   │       ├── providers.go    # 供应商 Handler
│   │       ├── proxy.go        # 代理管理 Handler
│   │       ├── logs.go         # 日志 Handler
│   │       ├── mcp.go          # MCP Handler
│   │       ├── gateway.go      # API 网关 Handler
│   │       ├── assistant.go    # AI 助手 Handler
│   │       └── ...
│   ├── config/
│   │   └── config.go           # 环境变量配置
│   └── ws/
│       └── hub.go              # WebSocket Hub
├── services/                    # 复用现有服务代码（最小改动）
│   ├── providerrelay.go
│   ├── providerservice.go
│   ├── database.go
│   └── ...
├── frontend/                    # Vue 3 前端（改造后）
│   ├── src/
│   │   ├── api/               # HTTP API 客户端（替代 bindings）
│   │   ├── components/
│   │   ├── router/
│   │   └── ...
│   └── ...
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── Taskfile.yml
```

## 6. 迁移步骤

### Phase 1: 后端 API 化（核心）
1. 创建 `cmd/server/main.go` 新入口，移除 Wails 依赖
2. 创建 `internal/api/` 路由层，将服务方法暴露为 REST API
3. 实现 JWT 认证中间件
4. 实现 WebSocket 事件推送
5. 修改 `ProviderRelayService` 监听地址为 `0.0.0.0:18100`

### Phase 2: 前端改造
6. 移除 `@wailsio/runtime` 依赖
7. 创建 `src/api/` HTTP 客户端层
8. 替换所有 Wails binding 调用为 HTTP API 调用
9. 添加登录页面和路由守卫
10. 路由从 Hash 模式改为 History 模式

### Phase 3: 新功能
11. 实现 API 网关功能（Key 管理、速率限制）
12. 实现 AI 助手功能（Function Calling）
13. 添加 Gateway 和 Assistant 页面

### Phase 4: 部署
14. 编写 Dockerfile 和 docker-compose.yml
15. 配置 Nginx 反向代理（可选）
16. 编写部署文档

## 7. 关键技术决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 数据库 | 保持 SQLite | 单服务器部署，无需 PostgreSQL 的复杂性 |
| HTTP 框架 | 保持 Gin | 现有代理服务已使用 Gin，统一框架 |
| 前端构建 | 保持 Vite | 现有配置可复用 |
| 认证 | JWT | 无状态，适合前后端分离 |
| 实时通信 | WebSocket | 替代 Wails 事件系统 |
| 部署 | Docker | 简化 Linux 服务器部署 |
| 前端部署 | 后端静态文件服务 | 单容器部署，简化运维 |
| API 网关认证 | 独立 API Key | 与用户 JWT 分离，适合第三方调用 |

## 8. 安全考虑

1. **代理监听地址**：桌面版监听 `127.0.0.1:18100`（仅本地），Web 版改为 `0.0.0.0:18100`（需要 API Key 认证）
2. **API Key 加密**：使用 AES-256-GCM 加密存储，密钥从环境变量读取
3. **HTTPS**：生产环境通过 Nginx/Caddy 反向代理提供 TLS
4. **速率限制**：API 网关和管理 API 均需速率限制
5. **CORS**：生产环境严格限制允许的 Origin

## 9. 与需求的映射关系

| 需求 | 设计对应 |
|------|---------|
| 需求 1: 用户认证 | §2.4 认证设计 |
| 需求 2-4: 供应商/降级/映射 | §2.2 providers/proxy API |
| 需求 5: 用量统计 | §2.2 logs API |
| 需求 6-7: MCP/CLI 配置 | §2.2 mcp/cli-config API |
| 需求 8: 日志 | §2.2 logs API + §2.5 WebSocket |
| 需求 9-10: 代理/设置 | §2.2 proxy/settings API |
| 需求 11: CORS | §2.7 配置管理 |
| 需求 12: 备份 | §2.2 import API |
| 需求 13-14: 健康/速度 | §2.2 health/speedtest API |
| 需求 15: 提示词 | §2.2 prompts API |
| 需求 16: 深度链接 | §3.6 改为 URL 参数 |
| 需求 17-20: UX | §3 前端设计 |
| 需求 21: 安全 | §8 安全考虑 |
| 需求 22-23: 部署 | §5 部署设计 |
| 需求 24: API 文档 | Swagger 集成（Phase 2） |
| 需求 25: 版本管理 | §2.2 update API |
| 需求 26: API 网关 | §2.3 + §2.2 gateway API |
| 需求 27: AI 助手 | §4 AI 助手设计 |
