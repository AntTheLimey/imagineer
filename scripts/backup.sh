#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/imagineer_${TIMESTAMP}.sql.gz"

mkdir -p "${BACKUP_DIR}"

echo "Creating backup: ${BACKUP_FILE}"
docker exec imagineer-postgres pg_dump -U imagineer imagineer | gzip > "${BACKUP_FILE}"

echo "Backup complete: ${BACKUP_FILE}"
echo "Size: $(du -h "${BACKUP_FILE}" | cut -f1)"

# Keep only last 10 backups
ls -t "${BACKUP_DIR}"/imagineer_*.sql.gz 2>/dev/null | tail -n +11 | xargs -r rm -f
echo "Old backups cleaned up (keeping last 10)"
