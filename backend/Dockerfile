# Stage 1: Build
FROM golang:1.23-bookworm AS builder

# Install build dependencies for sqlcipher/cgo
RUN apt-get update && apt-get install -y \
    build-essential \
    libsqlite3-dev \
    tcl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod and go.sum from the backend directory
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for sqlcipher
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server/main.go

# Stage 2: Runtime
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Default environment variables
ENV EQUINOX_DB_PATH=/data/vault.db
ENV EQUINOX_HEX_KEY=0000000000000000000000000000000000000000000000000000000000000000
ENV JWT_SECRET=equinox-default-secret-change-me

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./main"]
