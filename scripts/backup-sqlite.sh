#!/usr/bin/env bash
set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
TARGET_FILE="${BACKUP_DIR}/cards_backup_${TIMESTAMP}.db"
CONTAINER_NAME="cards-api"
DB_PATH="/data/cards.db"

mkdir -p "${BACKUP_DIR}"

echo "Starting SQLite safe online backup..."
if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    # Use sqlite3 .backup API for zero-lock, fully consistent backup
    docker exec "${CONTAINER_NAME}" sqlite3 "${DB_PATH}" ".backup '/tmp/backup.db'"
    docker cp "${CONTAINER_NAME}:/tmp/backup.db" "${TARGET_FILE}"
    docker exec "${CONTAINER_NAME}" rm -f /tmp/backup.db
    gzip -f "${TARGET_FILE}"
    echo "Backup completed successfully: ${TARGET_FILE}.gz"
else
    echo "Error: Container ${CONTAINER_NAME} is not running."
    exit 1
fi
