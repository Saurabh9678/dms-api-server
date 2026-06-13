#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$DEPLOY_DIR/../.." && pwd)"
MIGRATIONS_DIR="$REPO_ROOT/migrations"
MIGRATE_IMAGE="migrate/migrate:v4.18.1"
ACTION="${1:-up}"

set -a; source "$DEPLOY_DIR/.env"; set +a

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "[migrate] ERROR: migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

if ! docker inspect infiniour-postgres &>/dev/null; then
    echo "[migrate] ERROR: infiniour-postgres is not running."
    echo "[migrate] Start Postgres first: cd $DEPLOY_DIR && docker compose up -d postgres"
    exit 1
fi

NETWORK="$(docker inspect infiniour-postgres --format '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}')"
ENCODED_PASSWORD="$(python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$POSTGRES_PASSWORD")"
DATABASE_URL="postgres://${POSTGRES_USER}:${ENCODED_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable"

run_migrate() {
    docker run --rm \
        --network "$NETWORK" \
        -v "$MIGRATIONS_DIR:/migrations" \
        "$MIGRATE_IMAGE" \
        -path /migrations \
        -database "$DATABASE_URL" \
        "$@"
}

echo "[migrate] Running migrations ($ACTION) on network $NETWORK..."
run_migrate "$ACTION"

if [ "$ACTION" = "up" ]; then
    echo "[migrate] Current version:"
    run_migrate version
fi

echo "[migrate] Done."
