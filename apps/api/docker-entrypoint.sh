#!/bin/sh
# Runs as root only long enough to make the data volume writable, then drops privileges.
# Volumes created by older (root-running) images keep their root ownership, so this fix-up is required for upgrades.
set -e

if [ "$(id -u)" = "0" ]; then
    chown -R app:app /data
    exec su-exec app "$@"
fi

exec "$@"
