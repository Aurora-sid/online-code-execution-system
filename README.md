# 在线代码执行平台 - 运行指南

## 📋 项目简介
基于 Docker 的多用户在线编程执行系统，支持 Python、Java、C++、Go 四种编程语言的在线编写和执行。

## 🛠️ 环境要求
- **Docker Desktop** (必须)
- **Node.js** 22.12+ (前端开发)
- **Go** 1.21+ (可选，后端开发)

---

## 🚀 快速启动

### 1. 启动基础设施 (MySQL + Redis)
```bash
cd deploy
docker-compose up -d
```

### 2. 构建代码执行环境镜像
```bash
# Python 环境
docker build -t code-exec/python deploy/images/python

# Java 环境
docker build -t code-exec/java deploy/images/java

# C++ 环境
docker build -t code-exec/cpp deploy/images/cpp

# Go 环境
docker build -t code-exec/go deploy/images/go
```

### 3. 启动后端服务
```bash
# 使用 Docker 运行后端（推荐）
docker run -d --name backend_server \
  --network deploy_code_exec_net \
  -p 8080:8080 \
  -v $(pwd)/backend:/app \
  -w /app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e MYSQL_HOST=code_exec_mysql \
  -e MYSQL_PORT=3306 \
  -e REDIS_ADDR=code_exec_redis:6379 \
  code-exec/go go run main.go
```

### 4. 启动前端服务
```bash
cd frontend
npm install
npm run dev
```

### 5. 访问应用
打开浏览器访问：**http://localhost:5173**

---

## � 再次启动（日常使用）

镜像构建是**一次性**的，下次启动只需执行以下步骤：

```bash
# 1. 启动基础设施（如果已停止）
cd deploy
docker-compose up -d

# 2. 启动后端服务（如果已停止）
docker start backend_server

# 3. 启动前端
cd frontend
npm run dev
```

> **提示**：如果 `backend_server` 容器不存在，需要重新运行第3步的 `docker run` 命令创建。

---

## �📁 项目结构
```
├── backend/          # Go 后端服务
│   ├── config/       # 配置加载
│   ├── internal/     # 核心业务逻辑
│   │   ├── api/      # HTTP/WebSocket 接口
│   │   ├── auth/     # JWT 认证
│   │   ├── docker/   # Docker 沙箱执行
│   │   ├── model/    # 数据模型
│   │   └── queue/    # Redis 任务队列
│   └── main.go       # 入口文件
├── frontend/         # Vue 3 前端
│   └── src/
│       ├── components/  # 编辑器和终端组件
│       └── App.vue      # 主应用
└── deploy/           # 部署配置
    ├── docker-compose.yml  # MySQL + Redis
    └── images/             # 代码执行环境镜像
```

---

## 🔧 常见问题

### 后端启动失败
```bash
# 检查日志
docker logs backend_server

# 重启服务
docker restart backend_server
```

### 前端连接错误
确保后端服务已启动并监听 8080 端口。

### Docker 镜像构建慢
Java 镜像构建较慢（需下载 OpenJDK），请耐心等待或先使用其他语言测试。

---

## 📞 端口说明
| 服务 | 端口 |
|------|------|
| 前端 | 5173 |
| 后端 | 8080 |
| MySQL | 13306 |
| Redis | 16379 |
