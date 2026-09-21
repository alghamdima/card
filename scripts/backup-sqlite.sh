#!/usr/bin/env bash
# Online, consistent backup of the SQLite database using SQLite's own .backup API.
# Usage: ./scripts/backup-sqlite.sh        (env: BACKUP_DIR=./backups KEEP_DAYS=30)
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
KEEP_DAYS="${KEEP_DAYS:-30}"
CONTAINER_NAME="${CONTAINER_NAME:-cards-api}"
DB_PATH="${DB_PATH:-/data/cards.db}"

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
TARGET_FILE="${BACKUP_DIR}/cards_backup_${TIMESTAMP}.db"
CONTAINER_TMP="/tmp/backup_${TIMESTAMP}.db"

if ! docker ps --format '{{.Names}}' | grep -qx "${CONTAINER_NAME}"; then
    echo "Error: container ${CONTAINER_NAME} is not running." >&2
    exit 1
fi

mkdir -p "${BACKUP_DIR}"

# Always remove the temporary file inside the container, even when a step fails.
cleanup() { docker exec "${CONTAINER_NAME}" rm -f "${CONTAINER_TMP}" >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo "Creating online backup..."
docker exec "${CONTAINER_NAME}" sqlite3 "${DB_PATH}" ".backup '${CONTAINER_TMP}'"

# A backup that cannot be opened is worse than none: verify it before keeping it.
CHECK="$(docker exec "${CONTAINER_NAME}" sqlite3 "${CONTAINER_TMP}" 'PRAGMA integrity_check;')"
if [ "${CHECK}" != "ok" ]; then
    echo "Error: backup failed the integrity check: ${CHECK}" >&2
    exit 1
fi

docker cp "${CONTAINER_NAME}:${CONTAINER_TMP}" "${TARGET_FILE}"
gzip -f "${TARGET_FILE}"
echo "Backup completed successfully: ${TARGET_FILE}.gz"

# Retention: drop backups older than KEEP_DAYS (set KEEP_DAYS=0 to keep everything).
if [ "${KEEP_DAYS}" -gt 0 ]; then
    find "${BACKUP_DIR}" -name 'cards_backup_*.db.gz' -mtime "+${KEEP_DAYS}" -print -delete
fi
