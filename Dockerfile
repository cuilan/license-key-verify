# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git make

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/lkctl /usr/local/bin/lkctl
COPY --from=builder /app/bin/lkverify /usr/local/bin/lkverify

# 设置可执行权限
RUN chmod +x /usr/local/bin/lkctl /usr/local/bin/lkverify

# 切换到非 root 用户
USER appuser

# 设置入口点
ENTRYPOINT ["lkctl"]

# 默认命令
CMD ["--help"] 