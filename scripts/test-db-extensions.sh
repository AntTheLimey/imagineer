#!/bin/bash
# Test that all required PostgreSQL extensions are installed
# Run this after docker compose up to verify database configuration

set -e

echo "Testing PostgreSQL extensions..."

# Required extensions
REQUIRED_EXTENSIONS=(
    "pg_trgm"
    "uuid-ossp"
    "vector"
    "vchord_bm25"
    "pgmq"
    "pg_cron"
    "pg_tokenizer"
    "pgedge_vectorizer"
    "vectorize"
)

# Get installed extensions
INSTALLED=$(docker exec imagineer-postgres psql -U imagineer -t -c \
    "SELECT extname FROM pg_extension ORDER BY extname;")

FAILED=0

for ext in "${REQUIRED_EXTENSIONS[@]}"; do
    if echo "$INSTALLED" | grep -q "^[[:space:]]*${ext}[[:space:]]*$"; then
        echo "  ✓ $ext"
    else
        echo "  ✗ $ext - NOT INSTALLED"
        FAILED=1
    fi
done

if [ $FAILED -eq 1 ]; then
    echo ""
    echo "FAIL: Some required extensions are missing."
    echo "Run 'docker compose down && rm -rf data/postgres && docker compose up -d' to reinitialize."
    exit 1
fi

echo ""
echo "All required extensions are installed."

# Verify PostgreSQL version
VERSION=$(docker exec imagineer-postgres psql -U imagineer -t -c "SHOW server_version;")
echo "PostgreSQL version: $VERSION"

# Verify shared_preload_libraries
PRELOAD=$(docker exec imagineer-postgres psql -U imagineer -t -c "SHOW shared_preload_libraries;")
echo "shared_preload_libraries: $PRELOAD"

echo ""
echo "Database configuration test passed."
