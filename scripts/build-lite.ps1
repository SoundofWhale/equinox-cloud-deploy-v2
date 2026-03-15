Write-Host "🍃 Building Equinox 2.0 (Lite Version)..."

# Ensure docker is running
docker info >$null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error: Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Build and run using the lite docker-compose.lite.yml (No Ollama)
docker-compose -f docker/docker-compose.lite.yml up -d --build

Write-Host "✅ Equinox Lite successfully built and deployed!" -ForegroundColor Green
Write-Host "🌐 Frontend: http://localhost:5173"
Write-Host "⚙️ Backend:  http://localhost:8080"
