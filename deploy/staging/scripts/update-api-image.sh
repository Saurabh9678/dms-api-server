#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$DEPLOY_DIR"

echo "[update-api-image] Pulling latest API image..."
docker compose pull api

echo "[update-api-image] Recreating API container (Postgres untouched)..."
docker compose up -d --no-deps api

echo "[update-api-image] Pruning dangling images..."
docker image prune -f

echo "[update-api-image] Verifying API health..."
sleep 5
if ! curl -sf http://127.0.0.1:8080/health > /dev/null; then
    echo "[update-api-image] ERROR: Health check failed. Showing recent logs:"
    docker compose logs --tail=50 api
    exit 1
fi
echo "[update-api-image] Health check passed."

echo "[update-api-image] Service status:"
docker compose ps api
