#!/bin/bash

set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "Usage: $0 <path-to-backup.sql.gz>"
  exit 1
fi

BACKUP_FILE="$1"

[ -f "$BACKUP_FILE" ] || { echo "Error: file not found: $BACKUP_FILE"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

set -a; source "$DEPLOY_DIR/.env"; set +a

echo "Restoring from: $BACKUP_FILE"
echo "WARNING: This will overwrite the current database. Ctrl+C to cancel. Proceeding in 5s..."
sleep 5

gunzip -c "$BACKUP_FILE" | docker exec -i infiniour-postgres \
  psql -U "$POSTGRES_USER" "$POSTGRES_DB"

echo "Restore complete."
