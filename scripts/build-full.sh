#!/bin/bash
# scripts/build-full.sh

echo "🧠 Building Equinox 2.0 (Full Version)..."

# Ensure docker is running
docker info >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ Error: Docker is not running. Please start Docker Desktop first."
  exit 1
fi

# Build and run using the primary docker-compose.yml (includes Ollama)
docker-compose -f docker/docker-compose.yml up -d --build

echo "⬇️  Pulling Mistral AI model (this may take a while)..."
docker exec -it equinox-ollama ollama pull mistral:7b-instruct

echo "✅ Equinox Full successfully built and deployed!"
echo "🌐 Frontend: http://localhost:5173"
echo "⚙️  Backend:  http://localhost:8080"
