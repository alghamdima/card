#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <path-to-backup.db.gz or path-to-backup.db>"
    exit 1
fi

BACKUP_SRC="$1"
CONTAINER_NAME="cards-api"
DB_PATH="/data/cards.db"

if [ ! -f "${BACKUP_SRC}" ]; then
    echo "Error: Backup file not found: ${BACKUP_SRC}"
    exit 1
fi

TEMP_DB="/tmp/restore_$$.db"

if [[ "${BACKUP_SRC}" == *.gz ]]; then
    echo "Decompressing backup..."
    gunzip -c "${BACKUP_SRC}" > "${TEMP_DB}"
else
    cp "${BACKUP_SRC}" "${TEMP_DB}"
fi

echo "Restoring SQLite database to container ${CONTAINER_NAME}..."
docker cp "${TEMP_DB}" "${CONTAINER_NAME}:/tmp/restore.db"
docker exec "${CONTAINER_NAME}" sh -c "
    sqlite3 ${DB_PATH} '.restore /tmp/restore.db' &&
    rm -f /tmp/restore.db
"
rm -f "${TEMP_DB}"

echo "Database restored successfully."
