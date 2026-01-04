#!/bin/bash

# GYDS Chain Node Startup Script

set -e

# Configuration
DATA_DIR="${DATA_DIR:-./data}"
CONFIG_FILE="${CONFIG_FILE:-./config.json}"
GENESIS_FILE="${GENESIS_FILE:-./genesis.json}"
LOG_LEVEL="${LOG_LEVEL:-info}"
RPC_PORT="${RPC_PORT:-8545}"
WS_PORT="${WS_PORT:-8546}"
P2P_PORT="${P2P_PORT:-30303}"
VALIDATOR="${VALIDATOR:-false}"
MINER="${MINER:-false}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         GYDS Chain Node Launcher         ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}"

# Check if binary exists
if [ ! -f "./gydschain" ]; then
    echo -e "${RED}Error: gydschain binary not found${NC}"
    echo "Please build the project first: go build -o gydschain ./cmd/node"
    exit 1
fi

# Create data directory
mkdir -p "$DATA_DIR"
mkdir -p "$DATA_DIR/logs"

# Check genesis file
if [ ! -f "$GENESIS_FILE" ]; then
    echo -e "${YELLOW}Warning: Genesis file not found at $GENESIS_FILE${NC}"
    echo "Using default genesis configuration..."
fi

# Build command
CMD="./gydschain"
CMD="$CMD --datadir $DATA_DIR"
CMD="$CMD --loglevel $LOG_LEVEL"
CMD="$CMD --rpcport $RPC_PORT"
CMD="$CMD --wsport $WS_PORT"
CMD="$CMD --listen 0.0.0.0:$P2P_PORT"

if [ -f "$CONFIG_FILE" ]; then
    CMD="$CMD --config $CONFIG_FILE"
fi

if [ -f "$GENESIS_FILE" ]; then
    CMD="$CMD --genesis $GENESIS_FILE"
fi

if [ "$VALIDATOR" = "true" ]; then
    echo -e "${YELLOW}Running as validator node${NC}"
    CMD="$CMD --validator"
    
    if [ -n "$VALIDATOR_KEY" ]; then
        CMD="$CMD --validatorkey $VALIDATOR_KEY"
    fi
fi

if [ "$MINER" = "true" ]; then
    echo -e "${YELLOW}Mining enabled${NC}"
    CMD="$CMD --mine"
    
    if [ -n "$MINER_ADDRESS" ]; then
        CMD="$CMD --miner $MINER_ADDRESS"
    fi
    
    if [ -n "$MINING_THREADS" ]; then
        CMD="$CMD --threads $MINING_THREADS"
    fi
fi

# Add bootstrap peers if specified
if [ -n "$BOOTSTRAP_PEERS" ]; then
    CMD="$CMD --bootstrap $BOOTSTRAP_PEERS"
fi

echo -e "${GREEN}Starting node...${NC}"
echo "Command: $CMD"
echo ""

# Run the node
exec $CMD 2>&1 | tee "$DATA_DIR/logs/node.log"
