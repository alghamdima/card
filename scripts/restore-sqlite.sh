#!/usr/bin/env bash
# Restore the SQLite database from a backup created by backup-sqlite.sh.
# The API is stopped while the file is replaced: overwriting a database that a running process
# holds open (with its -wal/-shm files) can corrupt it.
# Usage: ./scripts/restore-sqlite.sh <backup.db.gz | backup.db>
set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <path-to-backup.db.gz or path-to-backup.db>" >&2
    exit 1
fi

BACKUP_SRC="$1"
CONTAINER_NAME="${CONTAINER_NAME:-cards-api}"
VOLUME_NAME="${VOLUME_NAME:-cards_db_data}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"

if [ ! -f "${BACKUP_SRC}" ]; then
    echo "Error: backup file not found: ${BACKUP_SRC}" >&2
    exit 1
fi

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT
RESTORE_DB="${WORK_DIR}/restore.db"

if [[ "${BACKUP_SRC}" == *.gz ]]; then
    gunzip -c "${BACKUP_SRC}" > "${RESTORE_DB}"
else
    cp "${BACKUP_SRC}" "${RESTORE_DB}"
fi

# Refuse to restore a file that is not a healthy SQLite database.
CHECK="$(docker run --rm -v "${WORK_DIR}:/restore:ro" alpine:3.20 sh -c 'apk add --no-cache sqlite >/dev/null && sqlite3 /restore/restore.db "PRAGMA integrity_check;"')"
if [ "${CHECK}" != "ok" ]; then
    echo "Error: the backup failed the integrity check: ${CHECK}" >&2
    exit 1
fi

mkdir -p "${BACKUP_DIR}"
SAFETY_COPY="pre_restore_$(date +%Y%m%d_%H%M%S).db"

echo "Stopping ${CONTAINER_NAME}..."
docker stop "${CONTAINER_NAME}" >/dev/null

echo "Replacing the database (a copy of the current one is kept as ${BACKUP_DIR}/${SAFETY_COPY})..."
docker run --rm \
    -v "${VOLUME_NAME}:/data" \
    -v "${WORK_DIR}:/restore:ro" \
    -v "$(cd "${BACKUP_DIR}" && pwd):/backups" \
    -e SAFETY_COPY="${SAFETY_COPY}" \
    alpine:3.20 sh -c '
        set -e
        [ -f /data/cards.db ] && cp /data/cards.db "/backups/${SAFETY_COPY}" || true
        rm -f /data/cards.db-wal /data/cards.db-shm
        cp /restore/restore.db /data/cards.db
    '

echo "Starting ${CONTAINER_NAME}..."
docker start "${CONTAINER_NAME}" >/dev/null

echo "Database restored successfully."
