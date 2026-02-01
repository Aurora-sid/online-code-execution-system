#!/usr/bin/env python3
"""
快速安全测试脚本 - 通过前端界面手动测试
将下面的代码复制到前端代码编辑器中运行
"""

print("""
=== 在线代码执行平台 - 安全测试用例 ===

请依次在前端界面测试以下代码：

-------------------
测试 1: Fork 炸弹（C语言）
-------------------
#include <unistd.h>
int main() {
    while(1) {
        fork();
    }
    return 0;
}

预期结果：
✓ 错误信息："执行超时：进程数量达到限制(50)，可能触发了 fork 炸弹攻击"
✓ 容器被自动清理

-------------------
测试 2: 内存炸弹（Python）
-------------------
data = []
while True:
    data.append('A' * 10**8)  # 每次100MB

预期结果：
✓ 错误信息："执行超时：内存使用过高(XX%)，可能触发了内存炸弹攻击"
   或 "执行超时：内存耗尽被系统终止(OOM Killed)"

-------------------
测试 3: CPU 死循环（C语言）
-------------------
int main() {
    while(1) {}
    return 0;
}

预期结果：
✓ 错误信息："执行超时：代码运行时间超过5秒限制，可能是死循环或计算量过大"
✓ CPU 限制在 50%

-------------------
测试 4: 读取系统文件（Python）
-------------------
try:
    with open('/etc/passwd', 'r') as f:
        content = f.read()
        print(f'File length: {len(content)} bytes')
        print(content[:200])  # 只打印前200字符
except Exception as e:
    print(f'Error: {e}')

预期结果：
✓ 能读取文件（容器内的 passwd，与宿主机隔离）
✓ 或 Permission denied

-------------------
测试 5: 网络访问（Python）
-------------------
import socket
try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect(('example.com', 80))
    print('Network accessible - SECURITY ISSUE!')
    s.close()
except Exception as e:
    print(f'Network blocked (Expected): {e}')

预期结果：
✓ 错误信息包含 "Network is unreachable" 或类似
✓ 网络访问被阻止

-------------------
测试 6: 环境变量检查（Python）
-------------------
import os
print('Environment variables:')
for key, value in os.environ.items():
    print(f'{key}={value}')

预期结果：
✓ 只显示容器内的环境变量
✓ 不应泄露宿主机敏感信息

-------------------
安全配置检查清单
-------------------
✓ PidsLimit: 50 - 防止 fork 炸弹
✓ Memory: 512MB - 防止内存耗尽
✓ NanoCPUs: 0.5 - CPU 使用限制
✓ NetworkMode: none - 禁止网络访问
✓ 执行超时: 5秒 - 防止无限循环
✓ 容器隔离 - 文件系统隔离
✓ 详细错误反馈 - 安全事件

测试方法：
1. 打开前端界面 http://localhost:5173
2. 选择对应的编程语言
3. 复制测试代码到编辑器
4. 点击"运行"按钮
5. 观察输出和错误信息
6. 检查 Docker 容器是否被正确清理：
   docker ps  # 不应有遗留的 code-exec 容器
""")
