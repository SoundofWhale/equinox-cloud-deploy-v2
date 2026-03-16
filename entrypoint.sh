#!/bin/sh

# EQUINOX Combo Startup Script
echo "🌿 EQUINOX All-in-One system starting..."

# 1. Start Whisper Engine (Local port 8081)
# We override the port via a wrapper if needed, but here we just run it.
# The wrapper defaults to 8080, so we might need to modify the PORT env.
export PORT=8081
/app/whisper-wrapper &
WHISPER_PID=$!

# 2. Start Frontend (Local port 3000)
# Node.js SvelteKit server
export PORT=3000
node /app/frontend/build &
FRONTEND_PID=$!

# 3. Start Backend (Public port 8080)
# Go binary acting as gateway
export EQUINOX_WHISPER_URL="http://localhost:8081/inference"
export EQUINOX_FRONTEND_URL="http://localhost:3000"
/app/backend-server
BACKEND_PID=$!

# Monitoring loop
while true; do
  ps -p $WHISPER_PID > /dev/null || { echo "❌ Whisper died"; exit 1; }
  ps -p $FRONTEND_PID > /dev/null || { echo "❌ Frontend died"; exit 1; }
  ps -p $BACKEND_PID > /dev/null || { echo "❌ Backend died"; exit 1; }
  sleep 10
done
