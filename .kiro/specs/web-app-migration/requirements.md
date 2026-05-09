# Code Switch Web 应用改造 - 需求文档

## 介绍

Code Switch 是一个 AI API 供应商管理和智能降级系统。当前以 Wails 桌面应用形式运行，需要改造为前后端分离的 Web 应用，以支持在 Linux 服务器上部署，通过浏览器访问。改造后保留所有现有功能，并适配 Web 环境的特性。

## 术语表

- **System**: Code Switch Web 应用系统，包括后端服务和前端应用
- **Backend**: Go 后端服务，运行在 Linux 服务器上，提供 REST API 接口
- **Frontend**: Vue 3 前端应用，通过浏览器访问，与后端通过 HTTP/HTTPS 通信
- **Provider**: AI API 供应商（Claude Code / Codex / Gemini CLI 等）
- **Proxy**: 本地代理服务，转发请求到配置的供应商，监听指定端口
- **MCP_Server**: Model Context Protocol 服务器，支持 URL 和命令两种类型
- **Token**: API 调用中的 Token 单位，包括输入 Token 和输出 Token
- **Fallback**: 智能降级机制，当主供应商失败时自动切换到备选供应商
- **User**: 使用 Code Switch 的开发者或系统管理员
- **Session**: 用户登录会话，由 JWT Token 标识
- **API_Key**: 供应商的认证密钥，在数据库中加密存储
- **JWT_Token**: JSON Web Token，用于用户认证和授权
- **CORS**: 跨域资源共享，允许前端跨域访问后端 API
- **Level**: 供应商优先级分组，范围 1-10，数字越小优先级越高
- **Configuration**: 系统配置信息，包括代理端口、日志级别、数据库路径等

## 需求

### 需求 1: 用户认证与授权

**用户故事**: 作为系统管理员，我想要对 Web 应用进行用户认证，以便只有授权用户才能访问和管理供应商配置。

#### 接受标准

1. WHEN 用户首次访问 Web 应用，THE Frontend SHALL 显示登录页面
2. WHEN 用户输入用户名和密码后点击登录按钮，THE Backend SHALL 验证凭证并返回 JWT_Token
3. IF 凭证验证成功，THEN THE Frontend SHALL 保存 JWT_Token 并重定向到主页面
4. IF 凭证验证失败，THEN THE Backend SHALL 返回 401 Unauthorized 错误和错误信息
5. WHEN 用户发送 API 请求时，THE Backend SHALL 从请求头中提取 JWT_Token 并验证其有效性
6. IF JWT_Token 过期或无效，THEN THE Backend SHALL 返回 401 Unauthorized 错误
7. WHEN 用户点击登出按钮，THE Frontend SHALL 删除本地存储的 JWT_Token，THE Backend SHALL 清除会话数据
8. WHILE 用户未登录，THE Backend SHALL 拒绝所有非公开 API 请求并返回 401 错误
9. WHERE 多用户场景，THE Backend SHALL 支持多个用户同时登录，每个用户拥有独立的 JWT_Token

### 需求 2: 供应商管理

**用户故事**: 作为开发者，我想要通过 Web 界面管理多个 AI API 供应商，以便灵活切换和配置不同的服务。

#### 接受标准

1. WHEN 用户访问供应商管理页面，THE Frontend SHALL 从 Backend 获取所有已配置的供应商列表并显示
2. WHEN 用户点击"添加供应商"按钮，THE Frontend SHALL 显示供应商配置表单
3. WHEN 用户填写供应商信息（名称、API URL、API Key、优先级）并提交，THE Backend SHALL 验证所有必填字段非空
4. IF 供应商名称为空或 API URL 格式无效，THEN THE Backend SHALL 返回 400 Bad Request 错误和具体错误信息
5. WHEN 验证通过，THE Backend SHALL 加密 API Key 后保存供应商配置到数据库
6. WHEN 用户编辑现有供应商，THE Backend SHALL 更新数据库中的配置并返回成功响应
7. WHEN 用户删除供应商，THE Backend SHALL 从数据库中移除该供应商并返回成功响应
8. WHEN 用户拖拽供应商卡片调整顺序，THE Frontend SHALL 发送更新请求，THE Backend SHALL 更新供应商的优先级顺序
9. WHEN 用户点击供应商的启用/禁用开关，THE Backend SHALL 更新供应商的状态字段
10. WHEN 用户点击"复制供应商"按钮，THE Backend SHALL 创建一个新的供应商副本（除 API Key 外的所有字段相同）

### 需求 3: 智能降级与故障转移

**用户故事**: 作为开发者，我想要配置多个供应商并设置优先级，以便在主供应商失败时自动切换到备选供应商。

#### 接受标准

1. WHEN 用户配置多个供应商并设置不同的优先级（Level 1-10），THE Backend SHALL 按优先级顺序尝试供应商
2. WHEN 请求发送到 Proxy 服务，THE Proxy SHALL 首先尝试 Level 1 的所有启用供应商
3. IF Level 1 的所有启用供应商都失败（连接超时或返回错误），THEN THE Proxy SHALL 尝试 Level 2 的启用供应商
4. WHEN 某个供应商成功响应（HTTP 状态码 2xx），THE Proxy SHALL 返回结果并记录日志
5. IF 所有供应商都失败，THEN THE Proxy SHALL 返回 502 Bad Gateway 错误并记录所有失败原因
6. WHILE 供应商被禁用，THE Proxy SHALL 跳过该供应商，不尝试转发请求
7. WHEN 用户配置了模型映射规则，THE Proxy SHALL 在转发请求前查找并应用模型名称转换

### 需求 4: 模型映射

**用户故事**: 作为开发者，我想要配置模型名称映射，以便不同供应商的模型名称能够自动转换。

#### 接受标准

1. WHEN 用户访问模型映射页面，THE Frontend SHALL 从 Backend 获取所有已配置的模型映射规则并显示
2. WHEN 用户添加新的模型映射规则（源模型名称、目标模型名称），THE Backend SHALL 验证两个字段都非空
3. IF 验证通过，THEN THE Backend SHALL 保存映射规则到数据库
4. WHEN 请求包含需要映射的模型名称，THE Proxy SHALL 查找映射规则并替换模型名称
5. IF 模型名称没有映射规则，THEN THE Proxy SHALL 使用原始模型名称转发请求
6. WHEN 用户编辑映射规则，THE Backend SHALL 更新数据库中的规则
7. WHEN 用户删除映射规则，THE Backend SHALL 从数据库中移除该规则

### 需求 5: 用量统计与成本追踪

**用户故事**: 作为开发者，我想要查看 API 使用统计和成本分析，以便了解使用情况和控制成本。

#### 接受标准

1. WHEN 用户访问统计页面，THE Frontend SHALL 从 Backend 获取统计数据并显示仪表板
2. WHEN 请求通过 Proxy 转发，THE Backend SHALL 记录请求信息（时间戳、供应商名称、模型名称、输入 Token 数、输出 Token 数、成本）到数据库
3. WHEN 用户查看热力图，THE Frontend SHALL 显示每日使用量的可视化图表
4. WHEN 用户查看请求统计，THE Frontend SHALL 显示请求总数、成功率、平均响应时间
5. WHEN 用户查看 Token 统计，THE Frontend SHALL 显示输入 Token 总数、输出 Token 总数、总 Token 数
6. WHEN 用户查看成本统计，THE Frontend SHALL 基于官方定价计算费用并显示总成本
7. WHERE 用户选择时间范围，THE Backend SHALL 过滤统计数据并返回该时间范围内的统计结果

### 需求 6: MCP 服务器管理

**用户故事**: 作为开发者，我想要通过 Web 界面集中管理 MCP 服务器配置，以便在多个平台间同步配置。

#### 接受标准

1. WHEN 用户访问 MCP 服务器管理页面，THE Frontend SHALL 从 Backend 获取所有已配置的 MCP_Server 并显示
2. WHEN 用户添加新的 MCP_Server，THE Frontend SHALL 显示配置表单，支持 URL 和命令两种类型
3. WHEN 用户填写 MCP_Server 配置信息并提交，THE Backend SHALL 验证所有必填字段非空
4. IF 验证通过，THEN THE Backend SHALL 保存 MCP_Server 配置到数据库
5. WHEN 用户保存 MCP_Server 配置，THE Backend SHALL 同步配置到 Claude Code 和 Codex 的配置文件
6. WHEN 用户编辑或删除 MCP_Server，THE Backend SHALL 更新数据库和配置文件
7. WHEN 用户查看 MCP_Server 列表，THE Frontend SHALL 显示服务器名称、类型（URL 或命令）、状态

### 需求 7: CLI 配置编辑

**用户故事**: 作为开发者，我想要通过 Web 界面编辑 CLI 配置文件，以便管理 Claude Code 和 Codex 的配置。

#### 接受标准

1. WHEN 用户访问 CLI 配置编辑页面，THE Frontend SHALL 从 Backend 获取当前的 CLI 配置并显示
2. WHEN 用户修改可编辑字段（模型、插件等），THE Frontend SHALL 验证输入数据
3. WHEN 用户保存配置，THE Backend SHALL 验证配置格式并更新配置文件
4. IF 配置格式无效，THEN THE Backend SHALL 返回 400 Bad Request 错误和错误信息
5. WHEN 用户点击"解锁编辑"按钮，THE Frontend SHALL 允许用户直接编辑原始配置文件（JSON/YAML 格式）
6. WHEN 用户编辑原始配置，THE Backend SHALL 验证 JSON/YAML 格式的有效性

### 需求 8: 日志查看与调试

**用户故事**: 作为开发者，我想要查看代理的详细日志，以便调试问题和监控系统运行状态。

#### 接受标准

1. WHEN 用户访问日志页面，THE Frontend SHALL 从 Backend 获取最近的代理日志并显示
2. WHEN 请求通过 Proxy 转发，THE Backend SHALL 记录详细的日志信息（时间戳、请求 URL、供应商名称、响应状态码、错误信息）到日志文件
3. WHEN 用户在搜索框输入关键词，THE Backend SHALL 按关键词过滤日志并返回匹配结果
4. WHEN 用户选择时间范围，THE Backend SHALL 返回该时间范围内的日志
5. WHEN 用户点击日志条目，THE Frontend SHALL 显示详细信息（完整请求、完整响应、错误堆栈等）
6. WHEN 用户点击"清空日志"按钮，THE Backend SHALL 删除所有日志记录

### 需求 9: 代理服务管理

**用户故事**: 作为系统管理员，我想要管理代理服务的启动和停止，以便控制系统的运行状态。

#### 接受标准

1. WHEN 用户访问系统设置页面，THE Frontend SHALL 从 Backend 获取代理服务的状态并显示
2. WHEN 用户点击"启动代理"按钮，THE Backend SHALL 启动 Proxy 服务
3. WHEN 用户点击"停止代理"按钮，THE Backend SHALL 停止 Proxy 服务
4. WHEN 代理服务启动成功，THE Backend SHALL 返回成功响应，THE Frontend SHALL 显示成功提示和监听端口号
5. IF 代理服务启动失败，THEN THE Backend SHALL 返回错误响应和失败原因，THE Frontend SHALL 显示错误提示
6. WHILE 代理服务运行中，THE Frontend SHALL 显示"运行中"状态和绿色指示灯

### 需求 10: 系统设置与配置

**用户故事**: 作为系统管理员，我想要配置系统参数，以便自定义系统行为。

#### 接受标准

1. WHEN 用户访问系统设置页面，THE Frontend SHALL 从 Backend 获取所有可配置的参数并显示
2. WHEN 用户修改代理监听端口，THE Backend SHALL 验证端口号在 1-65535 范围内
3. IF 端口号已被占用或无效，THEN THE Backend SHALL 返回错误响应
4. WHEN 用户修改日志级别（DEBUG、INFO、WARN、ERROR），THE Backend SHALL 更新日志输出级别
5. WHEN 用户修改数据库路径，THE Backend SHALL 验证路径的有效性和可写权限
6. WHEN 用户保存设置，THE Backend SHALL 持久化配置到配置文件并返回成功响应

### 需求 11: 跨域资源共享 (CORS)

**用户故事**: 作为系统架构师，我想要配置 CORS 策略，以便前端能够安全地访问后端 API。

#### 接受标准

1. WHEN 前端发送跨域请求，THE Backend SHALL 检查请求的 Origin 是否在 CORS 允许列表中
2. IF 请求来源在允许列表中，THEN THE Backend SHALL 返回 CORS 响应头（Access-Control-Allow-Origin、Access-Control-Allow-Methods、Access-Control-Allow-Headers）
3. IF 请求来源不在允许列表中，THEN THE Backend SHALL 拒绝请求并返回 403 Forbidden 错误
4. WHEN 系统部署到生产环境，THE Backend SHALL 配置正确的 CORS 允许列表（仅包含前端域名）

### 需求 12: 数据持久化与备份

**用户故事**: 作为系统管理员，我想要备份和恢复系统数据，以便防止数据丢失。

#### 接受标准

1. WHEN 用户点击"导出配置"按钮，THE Backend SHALL 导出所有配置数据（供应商、模型映射、MCP 服务器等）为 JSON 文件并返回下载链接
2. WHEN 用户点击"导入配置"按钮，THE Frontend SHALL 显示文件选择对话框
3. WHEN 用户选择配置文件，THE Backend SHALL 验证文件格式是否为有效的 JSON
4. IF 文件格式无效，THEN THE Backend SHALL 返回错误响应，THE Frontend SHALL 显示错误提示
5. WHEN 用户确认导入，THE Backend SHALL 将配置数据导入到数据库
6. WHEN 系统运行时，THE Backend SHALL 每天定期备份数据库到备份目录

### 需求 13: 健康检查与监控

**用户故事**: 作为系统管理员，我想要监控系统和供应商的健康状态，以便及时发现问题。

#### 接受标准

1. WHEN 用户访问仪表板，THE Frontend SHALL 从 Backend 获取系统健康状态并显示
2. WHEN 用户点击"健康检查"按钮，THE Backend SHALL 向所有启用的供应商发送测试请求
3. WHEN 供应商连接成功（返回 HTTP 2xx 状态码），THE Backend SHALL 记录"正常"状态
4. IF 供应商连接失败，THEN THE Backend SHALL 记录"异常"状态和错误信息
5. WHEN 用户查看供应商详情，THE Frontend SHALL 显示最后一次健康检查的时间戳和结果

### 需求 14: 速度测试

**用户故事**: 作为开发者，我想要测试供应商的响应速度，以便选择最快的供应商。

#### 接受标准

1. WHEN 用户点击"速度测试"按钮，THE Backend SHALL 向所有启用的供应商发送测试请求并记录响应时间
2. WHEN 测试完成，THE Backend SHALL 返回每个供应商的响应时间（毫秒）
3. WHEN 用户查看测试结果，THE Frontend SHALL 按响应时间从快到慢排序显示供应商

### 需求 15: 自定义提示词管理

**用户故事**: 作为开发者，我想要管理自定义系统提示词，以便为不同的场景配置不同的提示词。

#### 接受标准

1. WHEN 用户访问提示词管理页面，THE Frontend SHALL 从 Backend 获取所有已保存的提示词并显示
2. WHEN 用户点击"添加提示词"按钮，THE Frontend SHALL 显示提示词编辑表单
3. WHEN 用户填写提示词名称和内容并保存，THE Backend SHALL 验证两个字段都非空
4. IF 验证通过，THEN THE Backend SHALL 保存提示词到数据库
5. WHEN 用户编辑或删除提示词，THE Backend SHALL 更新或删除数据库中的记录
6. WHEN 用户在 CLI 中使用提示词，THE Backend SHALL 从数据库中检索并应用

### 需求 16: 深度链接支持

**用户故事**: 作为开发者，我想要通过深度链接导入配置，以便快速分享和导入配置。

#### 接受标准

1. WHEN 用户点击 `ccswitch://` 深度链接，THE Frontend SHALL 解析链接参数
2. WHEN 链接包含有效的配置数据（Base64 编码），THE Frontend SHALL 显示导入确认对话框
3. WHEN 用户确认导入，THE Backend SHALL 验证配置数据格式并导入到系统
4. IF 链接格式无效或配置数据损坏，THEN THE Frontend SHALL 显示错误提示

### 需求 17: 前端响应式设计

**用户故事**: 作为用户，我想要在不同设备上访问 Web 应用，以便在桌面、平板和手机上都能正常使用。

#### 接受标准

1. WHEN 用户在桌面浏览器（宽度 > 1024px）上访问应用，THE Frontend SHALL 显示完整的桌面布局
2. WHEN 用户在平板设备（宽度 768-1024px）上访问应用，THE Frontend SHALL 显示适配的平板布局
3. WHEN 用户在手机设备（宽度 < 768px）上访问应用，THE Frontend SHALL 显示适配的移动布局
4. WHEN 用户调整浏览器窗口大小，THE Frontend SHALL 自动调整布局以适应新的窗口大小

### 需求 18: 国际化支持

**用户故事**: 作为全球用户，我想要使用不同语言的界面，以便更好地理解应用功能。

#### 接受标准

1. WHEN 用户首次访问应用，THE Frontend SHALL 根据浏览器的 Accept-Language 请求头显示对应语言的界面
2. WHEN 用户手动切换语言，THE Frontend SHALL 保存语言偏好到本地存储并更新界面
3. WHEN 用户刷新页面，THE Frontend SHALL 使用保存的语言偏好显示界面
4. WHEN 系统支持多种语言（中文、英文等），THE Frontend SHALL 提供语言切换菜单

### 需求 19: 错误处理与用户反馈

**用户故事**: 作为用户，我想要获得清晰的错误提示和操作反馈，以便了解系统状态和问题原因。

#### 接受标准

1. WHEN 操作成功，THE Frontend SHALL 显示成功提示信息（Toast 或 Notification）
2. WHEN 操作失败，THE Frontend SHALL 显示错误提示信息和具体的失败原因
3. WHEN 用户执行长时间操作（> 1 秒），THE Frontend SHALL 显示加载指示器或进度条
4. WHEN 网络连接中断，THE Frontend SHALL 显示连接错误提示并提供重试选项
5. WHEN 用户输入无效数据，THE Frontend SHALL 显示验证错误提示并指出错误字段

### 需求 20: 性能优化

**用户故事**: 作为用户，我想要应用响应迅速，以便获得良好的使用体验。

#### 接受标准

1. WHEN 用户首次加载应用，THE Frontend SHALL 在 3 秒内完成初始加载（包括 HTML、CSS、JavaScript）
2. WHEN 用户切换页面，THE Frontend SHALL 在 500ms 内完成页面切换
3. WHEN 用户执行 API 请求，THE Backend SHALL 在 1 秒内返回响应（不包括网络延迟）
4. WHEN 用户查看大量数据（> 100 条），THE Frontend SHALL 实现分页或虚拟滚动以优化性能

### 需求 21: 安全性

**用户故事**: 作为系统管理员，我想要确保系统的安全性，以便保护用户数据和 API 密钥。

#### 接受标准

1. WHEN 用户输入 API_Key，THE Backend SHALL 使用 AES-256 加密算法在数据库中加密存储
2. WHEN 用户查看 API_Key，THE Frontend SHALL 只显示最后 4 位字符，其余用星号（*）隐藏
3. WHEN 用户通过 HTTP 访问应用，THE Backend SHALL 强制重定向到 HTTPS
4. WHEN 用户发送 API 请求，THE Backend SHALL 验证 JWT_Token 的签名和过期时间
5. WHEN 用户登出，THE Backend SHALL 清除所有会话数据和 JWT_Token

### 需求 22: 容器化部署

**用户故事**: 作为系统管理员，我想要使用 Docker 容器部署应用，以便简化部署流程。

#### 接受标准

1. WHEN 系统提供 Dockerfile，THE Dockerfile SHALL 包含后端和前端的构建步骤
2. WHEN 用户运行 Docker 容器，THE System SHALL 正确初始化并启动所有服务
3. WHEN 容器启动，THE Backend SHALL 自动创建必要的目录（数据库目录、日志目录、备份目录）和配置文件
4. WHEN 用户停止容器，THE System SHALL 正确清理资源并保存数据

### 需求 23: 环境变量配置

**用户故事**: 作为系统管理员，我想要通过环境变量配置系统参数，以便在不同环境中灵活部署。

#### 接受标准

1. WHEN 系统启动，THE Backend SHALL 读取环境变量（如 PORT、LOG_LEVEL、DATABASE_PATH、CORS_ORIGINS）并应用配置
2. WHEN 环境变量未设置，THE Backend SHALL 使用预定义的默认值
3. WHEN 用户修改环境变量，THE Backend SHALL 在重启后应用新配置
4. WHEN 系统运行在 Docker 容器中，THE Backend SHALL 支持通过 .env 文件或 docker-compose 环境变量配置

### 需求 24: API 文档

**用户故事**: 作为开发者，我想要查看完整的 API 文档，以便集成和扩展系统。

#### 接受标准

1. WHEN 用户访问 `/api/docs` 端点，THE Backend SHALL 返回 Swagger/OpenAPI 3.0 格式的 API 文档
2. WHEN 用户查看 API 文档，THE Documentation SHALL 包含所有端点、请求参数、响应格式、错误码
3. WHEN 用户在文档中测试 API，THE Documentation SHALL 允许直接调用 API 并显示响应结果

### 需求 25: 版本管理与更新

**用户故事**: 作为系统管理员，我想要管理应用版本和更新，以便保持系统最新。

#### 接受标准

1. WHEN 用户访问关于页面，THE Frontend SHALL 显示当前应用版本号
2. WHEN 新版本发布，THE Backend SHALL 定期检查更新并通知用户
3. WHEN 用户点击"更新"按钮，THE Backend SHALL 下载新版本并安装
4. WHEN 更新完成，THE System SHALL 提示用户重启应用以应用新版本

### 需求 26: API 网关功能

**用户故事**: 作为第三方开发者，我想要通过 Code Switch 提供的 API 网关访问 AI 模型，以便复用其智能降级与故障转移功能。

#### 接受标准

1. WHEN 第三方应用向 Code Switch 发送 API 请求（带有有效的 API Key），THE Backend SHALL 验证 API Key 的有效性和权限
2. IF API Key 验证成功，THEN THE Backend SHALL 根据配置的供应商和优先级转发请求到对应的 AI 模型
3. WHEN 请求转发到 Proxy 服务，THE Proxy SHALL 按照智能降级规则（Level 1-10 优先级）尝试供应商
4. IF 主供应商失败，THEN THE Proxy SHALL 自动切换到备选供应商并重试
5. WHEN 某个供应商成功响应，THE Proxy SHALL 返回结果给第三方应用
6. IF 所有供应商都失败，THEN THE Backend SHALL 返回 502 Bad Gateway 错误和失败原因
7. WHEN 第三方应用发送请求，THE Backend SHALL 记录请求信息（来源应用、模型名称、Token 数、成本）到数据库
8. WHEN 用户查看统计数据，THE Frontend SHALL 支持按来源应用过滤和分析 API 网关的使用情况
9. WHEN 用户配置 API 网关，THE Backend SHALL 支持设置速率限制（Rate Limiting）以防止滥用
10. WHEN 第三方应用超过速率限制，THE Backend SHALL 返回 429 Too Many Requests 错误

### 需求 27: AI 助手与配置管理

**用户故事**: 作为用户，我想要通过 AI 助手与大模型直接沟通，并让 AI 助手帮我管理 Code Switch 的配置。

#### 接受标准

1. WHEN 用户访问 AI 助手页面，THE Frontend SHALL 显示聊天界面
2. WHEN 用户输入问题或指令，THE Frontend SHALL 发送消息到 Backend
3. WHEN Backend 收到消息，THE Backend SHALL 将消息转发到配置的 AI 模型（通过 Proxy 服务）
4. WHEN AI 模型返回响应，THE Backend SHALL 解析响应内容并返回给 Frontend
5. WHEN 用户的指令涉及配置操作（如"添加新供应商"、"修改代理端口"），THE Backend SHALL 解析指令并执行对应的操作
6. IF 指令解析成功，THEN THE Backend SHALL 执行操作并返回执行结果给 AI 模型
7. WHEN AI 模型收到执行结果，THE Backend SHALL 将结果转发给 Frontend 显示给用户
8. WHEN 用户要求 AI 助手查询配置信息（如"列出所有供应商"、"显示当前代理状态"），THE Backend SHALL 从数据库查询信息并返回给 AI 模型
9. WHEN AI 模型返回查询结果，THE Frontend SHALL 以友好的格式显示信息
10. WHEN 用户与 AI 助手的对话涉及敏感操作（如修改 API Key、删除供应商），THE Backend SHALL 要求用户确认
11. WHEN 用户确认操作，THE Backend SHALL 执行操作并记录操作日志
12. WHEN 用户查看对话历史，THE Frontend SHALL 显示与 AI 助手的所有对话记录
13. WHEN 用户清空对话历史，THE Backend SHALL 删除对话记录

