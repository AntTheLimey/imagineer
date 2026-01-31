#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="./backups"

# List available backups
echo "Available backups:"
ls -lt "${BACKUP_DIR}"/imagineer_*.sql.gz 2>/dev/null | head -10

echo ""
read -p "Enter backup filename to restore: " BACKUP_FILE

if [[ ! -f "${BACKUP_DIR}/${BACKUP_FILE}" ]]; then
    echo "Error: Backup file not found: ${BACKUP_DIR}/${BACKUP_FILE}"
    exit 1
fi

echo "WARNING: This will replace all current data!"
read -p "Are you sure? [y/N] " confirm
if [[ "$confirm" != "y" ]]; then
    echo "Restore cancelled."
    exit 0
fi

echo "Restoring from: ${BACKUP_FILE}"

# Drop and recreate database
docker exec imagineer-postgres psql -U imagineer -d postgres -c "DROP DATABASE IF EXISTS imagineer;"
docker exec imagineer-postgres psql -U imagineer -d postgres -c "CREATE DATABASE imagineer;"

# Restore backup
gunzip -c "${BACKUP_DIR}/${BACKUP_FILE}" | docker exec -i imagineer-postgres psql -U imagineer -d imagineer

echo "Restore complete!"
