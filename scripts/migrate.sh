#!/usr/bin/env bash
set -euo pipefail

# Imagineer Database Migration Runner
# Applies SQL migrations in order, tracking applied versions

MIGRATIONS_DIR="./migrations"
CONTAINER_NAME="imagineer-postgres"
DB_USER="imagineer"
DB_NAME="imagineer"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    log_error "Container ${CONTAINER_NAME} is not running. Start it with 'make up'"
    exit 1
fi

# Ensure schema_migrations table exists
docker exec "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" -q -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);
" 2>/dev/null

# Get list of applied migrations
APPLIED=$(docker exec "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" -t -A -c "
    SELECT version FROM schema_migrations ORDER BY version;
" 2>/dev/null || echo "")

log_info "Checking for pending migrations..."

# Track if any migrations were applied
MIGRATIONS_APPLIED=0

# Process each migration file in order
for migration_file in $(ls -1 "${MIGRATIONS_DIR}"/*.sql 2>/dev/null | sort); do
    filename=$(basename "$migration_file")
    version="${filename%.sql}"

    # Check if already applied
    if echo "$APPLIED" | grep -q "^${version}$"; then
        echo "  [✓] ${version} (already applied)"
        continue
    fi

    log_info "Applying migration: ${version}"

    # Apply the migration
    if docker exec -i "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" < "$migration_file" 2>&1; then
        # Record the migration (if not already recorded by the migration itself)
        docker exec "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" -q -c "
            INSERT INTO schema_migrations (version) VALUES ('${version}')
            ON CONFLICT (version) DO NOTHING;
        " 2>/dev/null
        log_info "  [✓] ${version} applied successfully"
        MIGRATIONS_APPLIED=$((MIGRATIONS_APPLIED + 1))
    else
        log_error "  [✗] Failed to apply ${version}"
        exit 1
    fi
done

if [ $MIGRATIONS_APPLIED -eq 0 ]; then
    log_info "Database is up to date. No migrations to apply."
else
    log_info "Applied ${MIGRATIONS_APPLIED} migration(s) successfully."
fi

# Show current migration status
echo ""
log_info "Migration status:"
docker exec "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" -c "
    SELECT version, applied_at FROM schema_migrations ORDER BY version;
"
