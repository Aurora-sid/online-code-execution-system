# 编译阶段
FROM golang:1.20 as builder

WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译成二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行阶段
FROM alpine:3.18

WORKDIR /app

# 从编译阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY configs/ ./configs/

# 暴露端口
EXPOSE 8080

# 运行二进制文件
CMD ["./main"]