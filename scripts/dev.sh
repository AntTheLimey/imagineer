#!/usr/bin/env bash
set -euo pipefail

# Imagineer Development Startup Script
# Starts all services needed for local development

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

cleanup() {
    log_info "Stopping services..."
    # Kill background processes
    if [ -n "${SERVER_PID:-}" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
    fi
    if [ -n "${CLIENT_PID:-}" ] && kill -0 "$CLIENT_PID" 2>/dev/null; then
        kill "$CLIENT_PID" 2>/dev/null || true
    fi
    log_info "Done."
}

trap cleanup EXIT INT TERM

cd "$PROJECT_DIR"

echo ""
echo "=========================================="
echo "  Imagineer Development Environment"
echo "=========================================="
echo ""

# Step 1: Check for .env file
log_step "Checking environment configuration..."
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        log_warn ".env file not found, copying from .env.example"
        cp .env.example .env
        log_warn "Please edit .env with your Google OAuth credentials"
        exit 1
    else
        log_error ".env file not found and no .env.example available"
        exit 1
    fi
fi

# Check for required OAuth variables
if ! grep -q "GOOGLE_CLIENT_ID=.\+" .env 2>/dev/null; then
    log_error "GOOGLE_CLIENT_ID not configured in .env"
    log_error "Please add your Google OAuth credentials to .env"
    exit 1
fi

# Step 2: Start Docker services
log_step "Starting Docker services..."
if ! docker info >/dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker Desktop."
    exit 1
fi

docker-compose up -d
log_info "Waiting for database to be ready..."
sleep 5

# Wait for postgres to be healthy
for i in {1..30}; do
    if docker exec imagineer-postgres pg_isready -U imagineer -d imagineer >/dev/null 2>&1; then
        log_info "Database is ready!"
        break
    fi
    if [ "$i" -eq 30 ]; then
        log_error "Database failed to start"
        exit 1
    fi
    sleep 1
done

# Step 3: Run migrations
log_step "Running database migrations..."
./scripts/migrate.sh

# Step 4: Load environment variables and start Go server
log_step "Starting API server..."
export $(grep -v '^#' .env | xargs)

# Kill any existing process on port 3001
lsof -ti:3001 2>/dev/null | xargs -r kill -9 2>/dev/null || true

go run ./cmd/server &
SERVER_PID=$!
sleep 3

if ! curl -s http://localhost:3001/health >/dev/null 2>&1; then
    log_error "API server failed to start"
    exit 1
fi
log_info "API server running at http://localhost:3001"

# Step 5: Start React client
log_step "Starting React client..."
cd client

# Kill any existing process on port 5173
lsof -ti:5173 2>/dev/null | xargs -r kill -9 2>/dev/null || true

npm run dev &
CLIENT_PID=$!
sleep 5

if ! curl -s http://localhost:5173 >/dev/null 2>&1; then
    log_error "React client failed to start"
    exit 1
fi
log_info "React client running at http://localhost:5173"

cd "$PROJECT_DIR"

echo ""
echo "=========================================="
echo "  All services started successfully!"
echo "=========================================="
echo ""
echo "  API Server:  http://localhost:3001"
echo "  Web Client:  http://localhost:5173"
echo "  MCP Server:  http://localhost:8080"
echo ""
echo "  Press Ctrl+C to stop all services"
echo ""

# Wait for user to stop
wait
