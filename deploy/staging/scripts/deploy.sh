#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$DEPLOY_DIR"

echo "[deploy] Pulling latest images..."
docker compose pull

echo "[deploy] Starting services..."
docker compose up -d

echo "[deploy] Pruning dangling images..."
docker image prune -f

echo "[deploy] Verifying API health..."
sleep 5
if ! curl -sf http://127.0.0.1:8080/health > /dev/null; then
    echo "[deploy] ERROR: Health check failed. Showing recent logs:"
    docker compose logs --tail=50 api
    exit 1
fi
echo "[deploy] Health check passed."

echo "[deploy] Service status:"
docker compose ps
