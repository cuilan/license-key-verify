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
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/bin/lkctl ./cmd/lkctl && \
    CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/bin/lkverify ./cmd/lkverify

# --- Final Stage ---
# Use a specific, minimal version of Alpine for the final image
FROM alpine:3.20

# Set timezone
ENV TZ=Asia/Shanghai

# Install dependencies and create user in a single RUN instruction to reduce layers
RUN apk --no-cache add tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# Copy the compiled binaries from the builder stage into a standard location
COPY --from=builder /app/bin/lkctl /app/bin/lkverify /usr/local/bin/

# Switch to the non-root user
USER appuser

# Set working directory for the running container
WORKDIR /app

# Set the entrypoint for the container
ENTRYPOINT ["lkctl"]

# Default command to run when the container starts
CMD ["--help"] 