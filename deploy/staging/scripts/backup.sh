#!/bin/bash

set -euo pipefail

BACKUP_DIR="/opt/infiniour/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/dms_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

set -a; source /opt/infiniour/.env; set +a

docker exec infiniour-postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" | gzip > "$BACKUP_FILE"

echo "Backup written: $BACKUP_FILE"

find "$BACKUP_DIR" -name "*.sql.gz" -mtime +14 -delete
echo "Old backups pruned (>14 days)."
