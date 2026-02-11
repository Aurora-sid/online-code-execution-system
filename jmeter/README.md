# 在线代码执行系统 - JMeter 压力测试

## 📁 文件清单

| 文件 | 说明 |
|------|------|
| `code_execution_test.jmx` | JMeter 测试计划 |
| `users.csv` | 测试用户凭证 |
| `test_codes.csv` | 多语言测试代码 |
| `create_test_users.bat` | 批量注册用户脚本 |

---

## 🚀 测试步骤

### 步骤 1: 确保后端运行
```powershell
# 检查 Docker 容器
docker ps | findstr backend
```

### 步骤 2: 创建测试用户
```powershell
cd g:\Go\jmeter
.\create_test_users.bat
```

### 步骤 3: 运行 JMeter 测试

**GUI模式（调试）：**
```powershell
D:\apache-jmeter\apache-jmeter-5.6.3\bin\jmeter.bat -t "g:\Go\jmeter\code_execution_test.jmx"
```

**命令行模式（推荐）：**
```powershell
D:\apache-jmeter\apache-jmeter-5.6.3\bin\jmeter -n -t "g:\Go\jmeter\code_execution_test.jmx" -l "g:\Go\jmeter\results.jtl" -e -o "g:\Go\jmeter\report"
```

### 步骤 4: 查看报告
打开 `g:\Go\jmeter\report\index.html`

---

## ⚙️ 测试配置

| 参数 | 值 | 说明 |
|------|-----|------|
| 并发用户数 | 20 | 模拟用户 |
| 循环次数 | 5 | 每用户执行次数 |
| 总请求数 | 300 | 20×5×3接口 |

---

## 📊 关键指标

| 指标 | 健康阈值 |
|------|----------|
| 错误率 | < 1% |
| 平均响应时间 | < 2秒 |
| 90%响应时间 | < 5秒 |
