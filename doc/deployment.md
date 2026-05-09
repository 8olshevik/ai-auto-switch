# Code Switch 部署文档

本文档介绍如何在 Linux 服务器上通过 Docker 部署 Code Switch Web 应用。

## 目录

- [前置要求](#前置要求)
- [Docker 部署步骤](#docker-部署步骤)
- [环境变量说明](#环境变量说明)
- [Nginx 反向代理配置](#nginx-反向代理配置)
- [数据备份与恢复](#数据备份与恢复)
- [常见问题](#常见问题)

---

## 前置要求

- Linux 服务器（Ubuntu 20.04+ / Debian 11+ / CentOS 8+）
- Docker 20.10+
- Docker Compose 2.0+（可选）
- 端口 8080（Web 服务）和 18100（代理服务）可用

---

## Docker 部署步骤

### 1. 构建 Docker 镜像

```bash
# 克隆项目
git clone https://github.com/Rogers-F/ai-auto-switch.git
cd ai-auto-switch

# 构建后端镜像
docker build -t codeswitch-backend:latest -f Dockerfile.backend .

# 构建前端镜像
docker build -t codeswitch-frontend:latest -f Dockerfile.frontend ./frontend
```

### 2. 创建数据目录

```bash
# 创建用于持久化存储的目录
mkdir -p ~/.code-switch/{data,logs}

# 设置适当的权限
chmod 755 ~/.code-switch
```

### 3. 运行容器

#### 使用 Docker Run（基础方式）

```bash
# 运行后端服务
docker run -d \
  --name codeswitch-backend \
  -p 8080:8080 \
  -p 18100:18100 \
  -v ~/.code-switch/data:/data \
  -v ~/.code-switch/logs:/logs \
  -e JWT_SECRET=your-secure-secret-key \
  -e ADMIN_PASSWORD=your-admin-password \
  codeswitch-backend:latest

# 运行前端服务（可选，前端也可以直接由后端服务静态托管）
docker run -d \
  --name codeswitch-frontend \
  -p 3000:80 \
  codeswitch-frontend:latest
```

#### 使用 Docker Compose（推荐方式）

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  backend:
    image: codeswitch-backend:latest
    container_name: codeswitch-backend
    ports:
      - "8080:8080"
      - "18100:18100"
    volumes:
      - ./data:/data
      - ./logs:/logs
    environment:
      - PORT=8080
      - PROXY_PORT=18100
      - PROXY_LISTEN_ADDR=0.0.0.0
      - DATABASE_PATH=/data/app.db
      - JWT_SECRET=${JWT_SECRET}
      - ADMIN_USERNAME=admin
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
      - CORS_ORIGINS=*
      - LOG_LEVEL=info
      - LOG_DIR=/logs
      - ASSISTANT_MODEL=claude-sonnet-4-20250514
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    image: codeswitch-frontend:latest
    container_name: codeswitch-frontend
    ports:
      - "3000:80"
    environment:
      - VITE_API_BASE_URL=http://localhost:8080
    depends_on:
      - backend
    restart: unless-stopped
```

创建 `.env` 文件：

```bash
# 生成安全的随机密钥
# 使用 openssl rand -base64 32 生成
JWT_SECRET=your-secure-jwt-secret-min-32-chars
ADMIN_PASSWORD=your-secure-admin-password
```

启动服务：

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 4. 验证部署

```bash
# 检查后端健康状态
curl http://localhost:8080/api/v1/health

# 检查代理服务状态
curl http://localhost:18100/health

# 访问 Web 界面
# 浏览器打开 http://<server-ip>:3000
```

---

## 环境变量说明

### 必需环境变量

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | string | - | JWT 签名密钥，**必须设置**。建议使用 32 位以上的随机字符串 |
| `ADMIN_PASSWORD` | string | - | 管理员账户密码，**必须设置** |

### 可选环境变量

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `PORT` | int | 8080 | Web 服务监听端口 |
| `PROXY_PORT` | int | 18100 | 代理服务监听端口（第三方应用调用 AI API 的端口） |
| `PROXY_LISTEN_ADDR` | string | "0.0.0.0" | 代理服务绑定地址 |
| `DATABASE_PATH` | string | "~/.code-switch/app.db" | SQLite 数据库文件路径 |
| `ADMIN_USERNAME` | string | "admin" | 管理员用户名 |
| `CORS_ORIGINS` | string | "*" | CORS 允许的来源，多个用逗号分隔 |
| `LOG_LEVEL` | string | "info" | 日志级别：debug, info, warn, error |
| `LOG_DIR` | string | "~/.code-switch/logs" | 日志目录路径 |
| `ASSISTANT_MODEL` | string | "claude-sonnet-4-20250514" | AI 助手默认使用的模型 |

### 环境变量示例

```bash
# 最小配置
JWT_SECRET=my-secure-secret-key-12345678901234567890
ADMIN_PASSWORD=admin123

# 完整配置
PORT=8080
PROXY_PORT=18100
PROXY_LISTEN_ADDR=0.0.0.0
DATABASE_PATH=/data/app.db
JWT_SECRET=my-secure-secret-key-12345678901234567890
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
CORS_ORIGINS=https://example.com,http://192.168.1.100:3000
LOG_LEVEL=debug
LOG_DIR=/logs
ASSISTANT_MODEL=claude-sonnet-4-20250514
```

---

## Nginx 反向代理配置

### 方案一：后端托管前端静态文件

配置 Nginx 同时代理 Web 服务和代理服务：

```nginx
upstream codeswitch_backend {
    server 127.0.0.1:8080;
}

upstream codeswitch_proxy {
    server 127.0.0.1:18100;
}

server {
    listen 80;
    server_name codeswitch.example.com;

    # Web 界面（前端）
    location / {
        proxy_pass http://codeswitch_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API 代理（第三方应用调用）
    location /api/ {
        proxy_pass http://codeswitch_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # AI 代理服务（第三方应用调用）
    location /v1/ {
        proxy_pass http://codeswitch_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket 支持
    location /api/v1/ws/ {
        proxy_pass http://codeswitch_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 方案二：独立前端部署

如果前端单独部署在 3000 端口：

```nginx
upstream codeswitch_frontend {
    server 127.0.0.1:3000;
}

upstream codeswitch_backend {
    server 127.0.0.1:8080;
}

upstream codeswitch_proxy {
    server 127.0.0.1:18100;
}

server {
    listen 80;
    server_name codeswitch.example.com;

    # 前端静态文件
    location / {
        proxy_pass http://codeswitch_frontend;
        proxy_set_header Host $host;
    }

    # 后端 API
    location /api/ {
        proxy_pass http://codeswitch_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # AI 代理服务
    location /v1/ {
        proxy_pass http://codeswitch_proxy;
        proxy_set_header Host $host;
    }

    # WebSocket
    location /api/v1/ws/ {
        proxy_pass http://codeswitch_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### HTTPS 配置（推荐）

```nginx
server {
    listen 443 ssl http2;
    server_name codeswitch.example.com;

    ssl_certificate /path/to/ssl/certificate.crt;
    ssl_certificate_key /path/to/ssl/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # ... 其他配置同上 ...
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name codeswitch.example.com;
    return 301 https://$server_name$request_uri;
}
```

---

## 数据备份与恢复

### 备份数据

Code Switch 的所有数据都存储在 SQLite 数据库中，位于配置的 `DATABASE_PATH` 路径。

#### 自动备份脚本

创建备份脚本 `backup.sh`：

```bash
#!/bin/bash

# 配置
BACKUP_DIR="/backup/codeswitch"
DATABASE_PATH="${DATABASE_PATH:-~/.code-switch/app.db}"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份数据库
cp "$DATABASE_PATH" "$BACKUP_DIR/app_$DATE.db"

# 备份日志（可选）
if [ -d ~/.code-switch/logs ]; then
    tar -czf "$BACKUP_DIR/logs_$DATE.tar.gz" ~/.code-switch/logs/
fi

# 清理旧备份
find "$BACKUP_DIR" -name "app_*.db" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "logs_*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "备份完成: $DATE"
```

添加定时任务：

```bash
# 编辑 crontab
crontab -e

# 每天凌晨 3 点执行备份
0 3 * * * /path/to/backup.sh
```

#### Docker 环境备份

```bash
# 备份数据库文件
docker cp codeswitch-backend:/data/app.db ./backup_app_$(date +%Y%m%d).db

# 或使用 volumes 备份
docker run --rm -v codeswitch_data:/data -v $(pwd)/backup:/backup alpine \
    cp /data/app.db /backup/app_$(date +%Y%m%d).db
```

### 恢复数据

#### 从备份恢复

```bash
# 停止服务
docker-compose down

# 恢复数据库
cp ./backup_app_20240101.db ~/.code-switch/app.db

# 启动服务
docker-compose up -d
```

#### Docker 环境恢复

```bash
# 停止服务
docker-compose down

# 恢复数据库到容器
docker cp ./backup_app_20240101.db codeswitch-backend:/data/app.db

# 启动服务
docker-compose up -d
```

### 数据迁移

如需将数据从一台服务器迁移到另一台：

```bash
# 源服务器：导出数据库
docker cp codeswitch-backend:/data/app.db ./app.db

# 复制到目标服务器
scp app.db user@target-server:/path/

# 目标服务器：导入数据库
docker cp ./app.db codeswitch-backend:/data/app.db

# 重启服务
docker-compose restart
```

---

## 常见问题

### 1. 容器无法启动

检查日志：
```bash
docker-compose logs backend
```

常见问题：
- 端口被占用：修改 `PORT` 或 `PROXY_PORT` 环境变量
- 权限问题：确保数据目录权限正确
- 环境变量未设置：确保 `JWT_SECRET` 和 `ADMIN_PASSWORD` 已设置

### 2. 无法访问 Web 界面

1. 检查防火墙：
   ```bash
   sudo ufw allow 8080
   sudo ufw allow 3000
   ```

2. 检查容器状态：
   ```bash
   docker ps
   docker-compose ps
   ```

### 3. 第三方应用无法调用 AI API

确认：
1. 代理服务端口 18100 已开放
2. 使用正确的 API Key 进行认证
3. 查看代理服务日志：
   ```bash
   docker-compose logs backend | grep proxy
   ```

### 4. 数据丢失

确保定期备份数据库。参考「数据备份与恢复」部分。

### 5. 性能问题

- 调整日志级别：`LOG_LEVEL=warn`
- 增加容器资源限制
- 使用 SSD 存储数据库

---

## 相关文档

- [快速开始指南](../QUICK_START.md)
- [API 端点文档](../API_ENDPOINT_PLAN_v2.2.0.md)
- [发布说明](../RELEASE_NOTES.md)