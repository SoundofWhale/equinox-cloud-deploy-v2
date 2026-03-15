Write-Host "🧠 Building Equinox 2.0 (Full Version)..."

# Ensure docker is running
docker info >$null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error: Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Build and run using the primary docker-compose.yml (includes Ollama)
docker-compose -f docker/docker-compose.yml up -d --build

Write-Host "⬇️ Pulling Mistral AI model (this may take a while)..."
docker exec -it equinox-ollama ollama pull mistral:7b-instruct

Write-Host "✅ Equinox Full successfully built and deployed!" -ForegroundColor Green
Write-Host "🌐 Frontend: http://localhost:5173"
Write-Host "⚙️ Backend:  http://localhost:8080"
