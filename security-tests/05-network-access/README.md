# 网络访问测试

## 攻击原理
代码尝试访问外部网络（连接公网IP、解析域名），如果沙盒未禁用网络，攻击者可能利用容器作为跳板攻击内网其他服务，或下载恶意脚本并在本地执行（即 C2 通信）。

## 测试用例

### test.py (Python)
- 尝试 Socket 连接
- 尝试 DNS 解析

### test.java (Java)
- 尝试 HttpURLConnection 发起 HTTP 请求

### test.js (Node.js)
- 尝试 http.get 发起请求

## 预期安全响应

✅ **正确行为**：
- 错误信息：`Network is unreachable` 或 `Name or service not known`
- 连接超时（如果防火墙丢包）
- 无法解析域名

❌ **失败情况**：
- 成功建立连接
- 成功获取 HTTP 响应状态码（200, 301等）

## 安全配置验证

- `NetworkMode: "none"` -  Docker 容器完全禁用网络栈
- 容器内只应有 `lo` 回环接口，且无法路由到外部
