# 后端部署指南

## 1. 编译（在开发机上交叉编译）

```bash
# 在 backend/ 目录下
cd backend

# 编译 Linux amd64 二进制
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o code-exec .
```

> Windows PowerShell 中需要先设置环境变量：
> ```powershell
> $env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
> go build -o code-exec .
> ```

## 2. 上传到服务器

```bash
# 上传二进制、配置文件和语言配置
scp code-exec root@your-server:/opt/code-exec/
scp .env root@your-server:/opt/code-exec/
scp languages.yaml root@your-server:/opt/code-exec/
```

## 3. 配置 systemd 守护进程

```bash
# 复制 service 文件
sudo cp code-exec.service /etc/systemd/system/

# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start code-exec

# 设置开机自启
sudo systemctl enable code-exec
```

## 4. 常用管理命令

```bash
# 查看状态
sudo systemctl status code-exec

# 查看日志（实时跟踪）
sudo journalctl -u code-exec -f

# 查看最近 100 行日志
sudo journalctl -u code-exec -n 100

# 重启服务
sudo systemctl restart code-exec

# 停止服务
sudo systemctl stop code-exec
```

## 5. 通过管理员面板查看日志

部署完成后，也可以通过**管理员面板 → 运行日志**页面底部的「服务器日志终端」实时查看后端日志，无需 SSH。

该功能基于内存环形缓冲区（最近 500 条日志），支持：
- 🔄 刷新：拉取最新日志
- 🗑️ 清空：清空缓冲区
- 🎨 语法高亮：错误红色、警告黄色、池操作青色

## 6. 注意事项

- 后端需要 **Docker 权限**，所以 service 文件中使用 `User=root`
- `.env` 文件中需要配置 MySQL DSN、Redis 地址、JWT 密钥等
- 确保服务器上已安装 Docker 并运行 `docker-compose up -d`（MySQL + Redis）
- 前端静态文件通过 Nginx 托管，Nginx 反向代理 `/api` 到后端端口
