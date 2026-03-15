FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 先复制依赖文件，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建静态链接的可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build  -o server ./cmd/server/

FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从 builder 复制编译好的二进制文件
COPY --from=builder /app/server .

# 使用非 root 用户运行（安全最佳实践）
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080 9090

CMD ["./server"]
