# Aurora Code - 在线代码执行平台

**Aurora Code** 是一个基于 MicroVM 和容器技术的现代化多语言在线编程平台。它旨在提供一个安全、快速、美观的在线代码执行环境，特别针对数据可视化和交互式编程进行了深度优化。

## ✨ 核心特性

### 1. 🚀 多语言支持与极速执行
*   **原生支持**: 目前完美支持 **Go, Python, Java, C++, C, JavaScript (Node.js)**。
*   **并发队列**: 基于 Redis 的任务队列系统，有效应对高并发提交。
*   **预热池技术**: 独创容器预热池（Container Pool），将代码执行启动时间降低至 **毫秒级**。

### 2. 🛡️ 企业级 Docker 沙箱
*   **资源隔离**: 每个任务运行在独立的 Docker 容器中，严格限制 CPU (Quota) 和内存 (Memory Limit)。
*   **攻击防御**:
    *   **Anti-Fork Bomb**: 实时监控并拦截进程无限繁衍攻击。
    *   **Memory Guard**: 智能识别并终止内存耗尽攻击（OOM）。
    *   **Sensitive IO Block**: 自动拦截对敏感系统文件（如 `/etc/passwd`）的非法读取。

### 3. 🖼️ 数据可视化 (Data Visualization)
*   **无头环境渲染**: 专为 Matplotlib 等库优化。检测到 GUI 调用（如 `plt.show()`）时自动提示用户。
*   **智能图片捕获**: 自动捕获容器生成的 `output.png` 图片文件。
*   **实时预览**: 后端自动转码传输，前端终端可直接点击缩略图查看高清大图，体验接近 Jupyter Notebook。

### 4. ⌨️ 实时交互 (Interactive Mode)
*   不同于传统的批处理 OJ，Aurora Code 支持**标准输入 (Stdin) 流式交互**。
*   用户可以在程序运行过程中，通过终端实时发送据，实现真正的“交互式编程”。

### 5. 🎨  UI 设计
*   **沉浸式体验**: 采用极光深色主题，长时间编程不刺眼。
*   **专业编辑器**: 集成 Microsoft Monaco Editor（VS Code 核心），支持智能提示、代码高亮。
*   **超级终端**: 定制版 Xterm.js，支持字体缩放、自适应布局。

---

## 🛠️ 技术栈

| 领域 | 技术组件 | 说明 |
| :--- | :--- | :--- |
| **Backend** | **Go (Golang)** | 高性能核心服务 |
| | **Gin Gonic** | Web 框架 & WebSocket 处理 |
| | **Docker Client** | 容器编排与控制 API |
| | **Redis** | 任务队列 & 缓存 |
| | **Gorm / MySQL** | 数据持久化 |
| **Frontend** | **Vue 3** | 渐进式 JavaScript 框架 |
| | **Vite** | 极速构建工具 |
| | **Tailwind CSS** | 实用主义 CSS 框架 |
| | **Monaco Editor** | 代码编辑器核心 |
| | **Xterm.js** | 网页终端模拟器 |

---

## 🚀 快速开始

### 1. 环境依赖
*   **Docker Desktop** (必须，用于沙箱环境)
*   **Node.js 18+** (前端开发)
*   **Go 1.21+** (后端开发)

### 2. 启动基础设施
启动 MySQL 和 Redis 服务：
```bash
cd deploy
docker-compose up -d
```

### 3. 构建执行环境镜像
**注意**: 项目首次运行必须构建沙箱使用的镜像。
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
(更多语言请参考 `deploy/images` 目录)

### 4. 启动后端服务
您可以在本地直接运行 Go 后端，它会自动连接到 Docker Daemon。
```bash
cd backend

# 安装依赖
go mod tidy

# 启动服务
go run main.go
```
*后端服务将监听 :8080*

### 5. 启动前端服务
```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```
*访问 **http://localhost:5173** 开始使用*

---

## 🔧 常见问题 (FAQ)

### Q: 程序里无法显示图片？(plt.show 报错)
**A:** 由于沙箱运行在服务器端（无显示器），请勿使用 `plt.show()`。
正确做法是保存为文件：
```python
import matplotlib.pyplot as plt
# ... 画图代码 ...
plt.savefig('output.png') # <--- 关键：保存为 output.png
```
系统会自动检测到该文件并显示在您的终端里。

### Q: 运行超时？
**A:** 每个请求有默认的 **25秒** 超时限制。如果您的代码涉及大量计算或 `sleep`，可能会被强制终止。

---

## 📂 项目结构

```
├── backend/          # Go 后端核心
│   ├── config/       # 配置管理
│   ├── internal/     
│   │   ├── api/      # HTTP & WebSocket 接口
│   │   ├── docker/   # Docker 沙箱池与执行逻辑
│   │   ├── queue/    # Redis 任务队列消费者
│   │   └── auth/     # 认证模块
│   └── languages.yaml # 语言编译/运行配置
├── frontend/         # Vue 3 前端
│   ├── src/
│   │   ├── components/  # 核心组件 (Terminal, CodeEditor)
│   │   └── assets/      # 静态资源
├── deploy/           # 部署与基础设施
│   ├── images/       # 语言镜像 Dockerfile
│   └── docker-compose.yml
├── jmeter/           # JMeter 压力测试套件
```

---
150: 
151: ## 🌍 如何添加新语言 (How to Add a New Language)
152: 
153: 想要为平台添加一门新语言（例如 `Ruby`）？只需按照以下三步操作：
154: 
155: ### 第一步：创建 Docker 执行环境
156: 
157: 1. 在 `deploy/images` 下创建新目录，例如 `ruby`。
158: 2. 在该目录中创建 `Dockerfile`，确保环境可以编译/运行该语言。
159:    ```dockerfile
160:    # deploy/images/ruby/Dockerfile
161:    FROM ruby:3.2-alpine
162:    WORKDIR /app
163:    CMD ["sleep", "infinity"] # 必须添加：保持容器常驻运行，等待调度
164:    ```
165: 3. 构建并标记镜像：
166:    ```bash
167:    docker build -t code-exec/ruby deploy/images/ruby
168:    ```
169: 
170: ### 第二步：配置后端
171: 
172: 1. 修改 `backend/languages.yaml`，注册语言信息和执行命令：
173:    ```yaml
174:    - id: ruby             # 唯一标识符
175:      image: code-exec/ruby # 对应的Docker镜像名
176:      filename: main.rb    # 用户代码保存的文件名
177:      run_cmd: ["ruby", "main.rb"] # 运行命令
178:    ```
179: 2. (可选) 如果希望初始化数据库时自动添加：
180:    修改 `backend/main.go` 中的 `seedLanguages` 函数，将新语言加入列表。
181: 
182: ### 第三步：配置前端
183: 
184: 1. **图标**: 将语言图标（如 `ruby.png`）放入 `frontend/src/assets/icons/`。
185: 2. **选择器**: 修改 `frontend/src/components/LanguageSelector.vue`，在 `defaultLanguages` 数组中添加新项。
186: 3. **编辑器支持**: 修改 `frontend/src/views/EditorView.vue`，在 `fileExtensions`, `mimeTypes`, `snippets` 等对象中添加对应配置。
187: 4. **初始化数据**: 修改 `frontend/src/stores/editor.js`，添加默认的 Hello World 代码片段。
188: 
189: 完成后重启后端服务即可生效！
190: 
191: ---

## 📊 在线用户追踪机制

### 当前实现（活跃用户计数）

**文件位置**: `backend/internal/api/middleware.go`

```go
// 在线用户存储（内存Map）
var onlineUsers = make(map[uint]bool)

// 用户访问受保护API时标记上线
func TrackUserOnline(userID uint) {
    onlineUsers[userID] = true
}

// 获取在线人数
func GetOnlineUserCount() int {
    return len(onlineUsers)
}
```

**触发时机**: 用户通过 JWT 认证访问任何受保护接口时自动标记。

### 机制特点

| 特性 | 说明 |
|------|------|
| **计数方式** | 唯一用户ID去重 |
| **存储位置** | 内存（服务重启清零） |
| **追踪范围** | 仅统计访问过API的用户 |

### ⚠️ 当前局限

1. **无超时机制**: 用户一旦上线，除非服务重启否则不会自动下线
2. **无实时性**: 不是基于WebSocket的实时连接追踪
3. **不持久化**: 服务重启后计数归零

### 🔧 优化建议（可选实现）

#### 方案一：心跳超时机制
```go
type OnlineUser struct {
    LastActive time.Time
}
var onlineUsers = make(map[uint]OnlineUser)

// 定时清理5分钟无活动的用户
func cleanupInactiveUsers() {
    for userID, user := range onlineUsers {
        if time.Since(user.LastActive) > 5*time.Minute {
            delete(onlineUsers, userID)
        }
    }
}
```

#### 方案二：WebSocket连接追踪
在 `websocket.go` 中追踪连接：
```go
// 连接建立时
connectedUsers[userID] = wsConn

// 连接断开时
delete(connectedUsers, userID)

// 实时在线 = len(connectedUsers)
```

---

## 🧪 压力测试

详见 `jmeter/README.md`，包含完整的JMeter测试计划和使用说明。