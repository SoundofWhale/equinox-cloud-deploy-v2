FROM golang:1.22-bookworm AS builder

# Install build dependencies for SQLCipher
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run tests inside container (CGO available here)
RUN CGO_ENABLED=1 go test ./internal/ai/... -v -run TestAnonymize
RUN CGO_ENABLED=1 go test ./internal/arbiter/... -v

# Build all binaries
RUN CGO_ENABLED=1 go build -o /equinox-server ./cmd/server
RUN CGO_ENABLED=1 go build -o /seed-slots ./cmd/seed_slots
RUN CGO_ENABLED=1 go build -o /check-cipher ./cmd/check_cipher

# ─── Final minimal image ───────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /equinox-server .
COPY --from=builder /seed-slots .
COPY --from=builder /check-cipher .

# Vault volume — mount at runtime
VOLUME ["/vault"]

EXPOSE 8080

CMD ["./equinox-server"]
