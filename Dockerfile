# Stage 1: Build SvelteKit Frontend
FROM node:20-slim AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.22-bookworm AS backend-builder
WORKDIR /app/backend
# Install SQLCipher dependencies
RUN apt-get update && apt-get install -y libsqlite3-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 go build -o /equinox-backend ./cmd/server/main.go

# Stage 3: Build Whisper.cpp
FROM debian:bookworm-slim AS whisper-builder
RUN apt-get update && apt-get install -y build-essential cmake git
WORKDIR /whisper
RUN git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git .
RUN cmake -B build -DWHISPER_BUILD_EXAMPLES=ON -DBUILD_SHARED_LIBS=OFF && \
    cmake --build build --config Release -j $(nproc)

# Stage 4: Build Whisper Go Wrapper
FROM golang:1.22-bookworm AS whisper-wrapper-builder
WORKDIR /app/whisper
COPY docker/whisper/main.go .
RUN go build -o /whisper-wrapper main.go

# Stage 5: Final Runtime Image
FROM node:20-bookworm-slim
RUN apt-get update && apt-get install -y ffmpeg libsqlite3-0 ca-certificates procps wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy all binaries
COPY --from=backend-builder /equinox-backend ./backend-server
COPY --from=whisper-builder /whisper/build/bin/whisper-cli ./whisper
COPY --from=whisper-wrapper-builder /whisper-wrapper ./whisper-wrapper

# Copy frontend
COPY --from=frontend-builder /app/frontend/build ./frontend/build
COPY --from=frontend-builder /app/frontend/package.json ./frontend/package.json
COPY --from=frontend-builder /app/frontend/node_modules ./frontend/node_modules

# Download tiny model
RUN mkdir -p /models && \
    wget -O /models/model.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin

# Copy startup script
COPY entrypoint.sh ./
RUN chmod +x ./entrypoint.sh

# Environment defaults
ENV EQUINOX_WHISPER_URL="http://localhost:8081/inference"
ENV EQUINOX_FRONTEND_URL="http://localhost:3000"
ENV NODE_ENV=production
ENV PORT=8080

EXPOSE 8080

CMD ["./entrypoint.sh"]
