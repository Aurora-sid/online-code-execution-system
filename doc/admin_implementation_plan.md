# 管理员后台系统实现方案 (Admin Dashboard Implementation Plan)

本方案基于现有架构设计，旨在添加一个安全的管理后台，用于系统监控、用户管理和运行日志审计。

## 1. 数据库变更 (Database Changes)

首先需要标识用户权限。

### `User` 表更新
新增 `Role` 字段，默认值为 `user`。

```go
type User struct {
    ID        uint   `gorm:"primaryKey"`
    Username  string `gorm:"uniqueIndex;size:50"`
    Password  string `gorm:"size:255"`
    Role      string `gorm:"size:20;default:'user'"` // 新增: "admin" 或 "user"
    CreatedAt time.Time
}
```

> **注意**: 首次启动时，需要通过 seed 或手动方式创建一个默认的 `superadmin` 账号。

---

## 2. 后端 API 开发 (Backend)

### 2.1 中间件 (Middleware)
新增 `AdminRequired` 中间件，在 `AuthMiddleware` 之后运行，检查 `user.role == 'admin'`。

### 2.2 管理员接口 (`/api/admin`)

| 请求方法 | 路径 | 功能描述 | 核心逻辑 |
| :--- | :--- | :--- | :--- |
| **GET** | `/stats` | **系统概览** | 1. **在线人数**: 读取 Redis 或 WebSocket 连接池计数。<br>2. **总用户数**: `Count(User)`。<br>3. **今日提交**: `Count(Submission where CreatedAt > today)`。 |
| **GET** | `/users` | **用户列表** | 分页列出用户 (ID, Username, Role, CreatedAt)。 |
| **POST** | `/users` | **创建用户** | 管理员手动创建新用户 (User/Admin)。 |
| **DELETE** | `/users/:id` | **删除用户** | 删除指定用户及相关提交记录 (需软删除或级联删除)。 |
| **GET** | `/submissions` | **详细日志** | 查阅所有提交记录。支持筛选：`status` (Failed/Timeout), `user_id`。<br>**重点**: 返回完整的 `Output` 和 `Error` 日志。 |

### 2.3 实时在线人数实现
由于后端可能是多副本部署 (虽然目前是单机 Docker)，建议将**在线连接**状态存储在 Redis 中。
*   **连接时**: `INCR online_users`
*   **断开时**: `DECR online_users`
*   或者维护一个 `Set` 集合 `SADD online_users {user_id}`，这样能去重（同一用户多端登录算1人）。

---

## 3. 前端开发 (Frontend)

新建 `src/views/Admin` 目录，包含以下页面：

### 3.1 路由结构
```javascript
{
  path: '/admin',
  component: AdminLayout,
  meta: { requiresAdmin: true },
  children: [
    { path: '', component: Dashboard },        // 仪表盘
    { path: 'users', component: UserManage },  // 用户管理
    { path: 'logs', component: SubmissionLogs } // 运行日志
  ]
}
```

### 3.2 页面设计

#### 📊 仪表盘 (Dashboard)
*   **卡片展示**: 在线人数 (实时)、总用户数、今日运行次数、错误率。
*   **图表** (可选): 使用 `ECharts` 或 `Chart.js` 展示最近 7 天的提交量趋势。

#### 👥 用户管理 (User Management)
*   **表格**: 包含 ID, 用户名, 注册时间, 操作列(删除/重置密码)。
*   **添加按钮**: 弹窗表单，输入用户名密码直接通过 Admin API 创建。

#### 📜 运行日志 (Submission Logs)
这是**核心排错功能**。
*   **表格**: 时间 | 用户 | 语言 | 状态(彩色Badge) | 耗时
*   **详情弹窗**: 点击某一行，弹出模态框，完整显示：
    *   **Source Code**: 高亮显示代码。
    *   **Input**: 标准输入内容。
    *   **Output/Error**: 完整的执行结果，包括沙箱拦截的系统报错信息。

---

## 4. 实施步骤 (Implementation Steps)

1.  **Phase 1: 基础建设**
    *   修改 Go Struct，执行数据库迁移。
    *   编写 `AdminAuthMiddleware`。
    *   手动在此次更新中注入一个管理员账号 (`admin/admin123`)。

2.  **Phase 2: 后端逻辑**
    *   实现 `/api/admin/stats` (先做伪实时的在线人数，读取 WebSocket 内存 Map 长度)。
    *   实现日志查询和用户管理 CRUD。

3.  **Phase 3: 前端开发**
    *   创建 Admin 布局和路由守卫。
    *   对接 Dashboard 数据。
    *   对接日志详情页。

4.  **Phase 4: 测试与部署**
    *   验证普通用户无法访问 `/admin`。
    *   验证删除用户功能是否安全。

---

## 5. 预期效果预览

**仪表盘:**
> [ 在线: 12人 ] [ 总用户: 156 ] [ 今日提交: 43 ] [ 报错: 2 ]

**日志详情:**
> **Status**: <span style="color:red">Failed</span>
> **Error**: `panic: runtime error: index out of range [3] with length 3`
