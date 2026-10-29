# --- Build Stage ---
# Use a specific version for reproducibility and a smaller base
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy dependency definition files to leverage Docker's layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
# It is recommended to add a .dockerignore file to exclude unnecessary files
COPY . .

# Build smaller, statically-linked binaries
# -ldflags "-w -s" strips debug info, reducing binary size
# CGO_ENABLED=0 helps create static binaries
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/bin/lkctl ./cmd/lkctl

# --- Final Stage ---
# Use a specific, minimal version of Alpine for the final image
FROM alpine:3.20

# Set timezone
ENV TZ=Asia/Shanghai

# 证书工具容器需要 ca-certificates 以正确解析系统时间
RUN apk --no-cache add tzdata ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# Copy the compiled binaries from the builder stage into a standard location
COPY --from=builder /app/bin/lkctl /usr/local/bin/

# Switch to the non-root user
USER appuser

# Set working directory for the running container
WORKDIR /app

# Set the entrypoint for the container
ENTRYPOINT ["lkctl"]

# Default command to run when the container starts
CMD ["--help"]

# --- 生产部署时用宿主机指纹绑定：
# docker run \
#   --mount type=bind,source=/etc/machine-id,target=/host/etc/machine-id,readonly \
#   --mount type=bind,source=/sys/class/dmi/id,target=/host/sys/class/dmi/id,readonly \
#   --network host \
#   -v $(pwd)/license.lic:/app/license.lic:ro \
#   your-image
#
# 这样容器内指纹 = 宿主裸机指纹，docker save 到新宿主机后验证必然失败。
# 如果不加上述 mount，容器内指纹依赖容器自身 machine-id + veth MAC，
# 每次 docker run 可能不同——不适合生产授权。 