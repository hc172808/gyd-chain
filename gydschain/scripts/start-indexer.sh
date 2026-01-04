#!/bin/bash

# GYDS Chain Indexer Startup Script

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-gydschain_indexer}"
DB_USER="${DB_USER:-gydschain}"
DB_PASSWORD="${DB_PASSWORD:-}"
RPC_URL="${RPC_URL:-http://localhost:8545}"
API_PORT="${API_PORT:-8080}"
LOG_LEVEL="${LOG_LEVEL:-info}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║       GYDS Chain Indexer Launcher        ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}"

# Check if binary exists
if [ ! -f "./gydschain-indexer" ]; then
    echo -e "${RED}Error: gydschain-indexer binary not found${NC}"
    echo "Please build the project first: go build -o gydschain-indexer ./cmd/indexer"
    exit 1
fi

# Build database connection string
if [ -n "$DB_PASSWORD" ]; then
    DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
else
    DB_URL="postgres://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
fi

# Check database connection
echo -e "${YELLOW}Checking database connection...${NC}"
if command -v pg_isready &> /dev/null; then
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME" -U "$DB_USER" &> /dev/null; then
        echo -e "${RED}Error: Cannot connect to database${NC}"
        echo "Please ensure PostgreSQL is running and the database exists."
        exit 1
    fi
    echo -e "${GREEN}Database connection OK${NC}"
fi

# Run migrations if schema file exists
if [ -f "./indexer/db/schema.sql" ]; then
    echo -e "${YELLOW}Running database migrations...${NC}"
    if command -v psql &> /dev/null; then
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "./indexer/db/schema.sql" 2>/dev/null || true
        echo -e "${GREEN}Migrations complete${NC}"
    else
        echo -e "${YELLOW}Warning: psql not found, skipping migrations${NC}"
    fi
fi

# Build command
CMD="./gydschain-indexer"
CMD="$CMD --db-url $DB_URL"
CMD="$CMD --rpc-url $RPC_URL"
CMD="$CMD --api-port $API_PORT"
CMD="$CMD --log-level $LOG_LEVEL"

echo -e "${GREEN}Starting indexer...${NC}"
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "RPC URL: $RPC_URL"
echo "API Port: $API_PORT"
echo ""

# Create logs directory
mkdir -p ./logs

# Run the indexer
exec $CMD 2>&1 | tee "./logs/indexer.log"
