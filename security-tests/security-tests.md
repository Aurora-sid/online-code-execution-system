# 安全测试用例集

## 自动化测试指南

本项目包含一个自动化安全测试套件，可自动运行多语言的恶意代码测试并生成报告。

### 运行测试

前提：确保后端服务（Go）和前端服务（可选）正在运行。

```bash
# 安装依赖
pip install requests

# 运行完整测试套件
python security-test-runner.py

# 指定 API 地址（如果不是默认的 localhost:8080）
python security-test-runner.py --url http://localhost:8080/execute
```

测试完成后，会在当前目录生成 `security-test-report.html`。

---

## 1. Fork 炸弹（进程资源耗尽）

### C 语言版本
详见 `security-tests/01-fork-bomb/test.c`

### Python 版本
详见 `security-tests/01-fork-bomb/test.py`

**预期结果**：PidsLimit=50 应阻止无限创建进程，超时后容器被停止。

## 2. 内存炸弹（内存资源耗尽）

详见 `security-tests/02-memory-bomb/` 目录下的测试代码。

**预期结果**：Memory=512MB 限制应阻止申请超过限制的内存，容器被 OOM 杀死。

## 3. CPU 密集型任务

详见 `security-tests/03-cpu-exhaustion/`。

**预期结果**：NanoCPUs=0.5 限制 CPU 使用，超时机制应强制终止。

## 4. 文件系统访问（权限隔离测试）

详见 `security-tests/04-filesystem-attacks/`。
- **读取测试**：尝试读取系统敏感文件。
- **写入测试**：尝试写入系统目录。
- **磁盘炸弹**：尝试填满磁盘空间。

**预期结果**：权限拒绝或只能访问容器内隔离的文件系统。

## 5. 网络访问测试

详见 `security-tests/05-network-access/`。

**预期结果**：NetworkMode=none 应阻止所有网络访问，报错 "Network is unreachable"。

## 6. 信息泄露测试

详见 `security-tests/06-info-leakage/`。

**预期结果**：不应泄露宿主机的环境变量或进程信息。

---

## 测试检查清单

- [x] PidsLimit: 50 (Fork Bomb)
- [x] Memory: 512MB (Memory Bomb)
- [x] NanoCPUs: 0.5 (CPU Limits)
- [x] NetworkMode: none (Network Isolation)
- [x] Execution Timeout: 5s
- [x] Filesystem Isolation
- [ ] Read-only Root FS (Optional)

