#!/bin/bash
# scripts/build-lite.sh

echo "🛠️  Building Equinox 2.0 (Lite Version)..."

# Ensure docker is running
docker info >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ Error: Docker is not running. Please start Docker Desktop first."
  exit 1
fi

# We build and run using docker-compose.lite.yml
docker-compose -f docker/docker-compose.lite.yml up -d --build

echo "✅ Equinox Lite successfully built and deployed!"
echo "🌐 Frontend: http://localhost:5173"
echo "⚙️  Backend:  http://localhost:8080"
