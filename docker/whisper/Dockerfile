# Stage 1: Build Whisper.cpp
FROM alpine:3.19 AS builder

RUN apk add --no-cache \
    build-base \
    cmake \
    git \
    linux-headers

WORKDIR /app
RUN git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git .
RUN cmake -B build -DWHISPER_BUILD_EXAMPLES=ON -DBUILD_SHARED_LIBS=OFF && \
    cmake --build build --config Release -j $(nproc)

# Build the Go wrapper
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY main.go .
RUN go build -o wrapper main.go

# Stage 2: Runtime Environment
FROM alpine:3.19

RUN apk add --no-cache \
    ca-certificates \
    libstdc++ \
    libgcc \
    libgomp \
    ffmpeg

WORKDIR /app
COPY --from=builder /app/build/bin/whisper-cli ./whisper
COPY --from=go-builder /app/wrapper ./wrapper

# Ensure we have a place for the model
RUN mkdir /models

# Performance optimizations for CPU
ENV OMP_NUM_THREADS=2

EXPOSE 8080

# Clean startup
CMD ["./wrapper"]
