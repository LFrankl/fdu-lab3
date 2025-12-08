# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 设置Go模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o express-server ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装时区包
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

# 复制编译后的二进制文件
COPY --from=builder /app/express-server .
COPY --from=builder /app/config/app.yaml ./config/

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./express-server"]