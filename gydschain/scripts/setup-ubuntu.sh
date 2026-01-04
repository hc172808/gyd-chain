#!/bin/bash

# GYDS Chain Ubuntu 22.04 Setup Script
# This script sets up a complete GYDS Chain node environment

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                  GYDS Chain Setup Script                      ║"
echo "║                     Ubuntu 22.04 LTS                          ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}Note: Some commands may require sudo access${NC}"
fi

# Update system
echo -e "${BLUE}[1/8] Updating system packages...${NC}"
sudo apt-get update -y
sudo apt-get upgrade -y

# Install essential packages
echo -e "${BLUE}[2/8] Installing essential packages...${NC}"
sudo apt-get install -y \
    build-essential \
    git \
    curl \
    wget \
    jq \
    make \
    gcc \
    g++ \
    pkg-config \
    libssl-dev \
    libleveldb-dev \
    liblz4-dev \
    libsnappy-dev \
    libzstd-dev \
    screen \
    htop \
    ufw \
    fail2ban

# Install Go
echo -e "${BLUE}[3/8] Installing Go 1.21...${NC}"
GO_VERSION="1.21.6"
if ! command -v go &> /dev/null || [[ $(go version | awk '{print $3}') != "go$GO_VERSION" ]]; then
    wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
    echo 'export GOPATH=$HOME/go' | sudo tee -a /etc/profile.d/go.sh
    echo 'export PATH=$PATH:$GOPATH/bin' | sudo tee -a /etc/profile.d/go.sh
    source /etc/profile.d/go.sh
fi
echo -e "${GREEN}Go version: $(go version)${NC}"

# Install PostgreSQL (for indexer)
echo -e "${BLUE}[4/8] Installing PostgreSQL 15...${NC}"
if ! command -v psql &> /dev/null; then
    sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
    sudo apt-get update
    sudo apt-get install -y postgresql-15 postgresql-client-15
fi

# Start PostgreSQL
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Create database for indexer
echo -e "${BLUE}[5/8] Setting up PostgreSQL database...${NC}"
sudo -u postgres psql -c "CREATE USER gydschain WITH PASSWORD 'gydschain123';" 2>/dev/null || true
sudo -u postgres psql -c "CREATE DATABASE gydschain_indexer OWNER gydschain;" 2>/dev/null || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gydschain_indexer TO gydschain;" 2>/dev/null || true
echo -e "${GREEN}Database created: gydschain_indexer${NC}"

# Create gydschain user and directories
echo -e "${BLUE}[6/8] Creating gydschain user and directories...${NC}"
if ! id "gydschain" &>/dev/null; then
    sudo useradd -m -s /bin/bash gydschain
fi

sudo mkdir -p /opt/gydschain
sudo mkdir -p /opt/gydschain/bin
sudo mkdir -p /opt/gydschain/data
sudo mkdir -p /opt/gydschain/config
sudo mkdir -p /opt/gydschain/logs
sudo mkdir -p /var/log/gydschain
sudo chown -R gydschain:gydschain /opt/gydschain
sudo chown -R gydschain:gydschain /var/log/gydschain

# Configure firewall
echo -e "${BLUE}[7/8] Configuring firewall...${NC}"
sudo ufw allow ssh
sudo ufw allow 30303/tcp  # P2P
sudo ufw allow 30303/udp  # P2P discovery
sudo ufw allow 8545/tcp   # RPC
sudo ufw allow 8546/tcp   # WebSocket
sudo ufw allow 8080/tcp   # Indexer API
echo "y" | sudo ufw enable || true
echo -e "${GREEN}Firewall configured${NC}"

# Create systemd service files
echo -e "${BLUE}[8/8] Creating systemd service files...${NC}"

# Node service
sudo tee /etc/systemd/system/gydschain-node.service > /dev/null <<EOF
[Unit]
Description=GYDS Chain Node
After=network.target

[Service]
Type=simple
User=gydschain
Group=gydschain
WorkingDirectory=/opt/gydschain
ExecStart=/opt/gydschain/bin/gydschain \\
    --datadir /opt/gydschain/data \\
    --config /opt/gydschain/config/config.json \\
    --genesis /opt/gydschain/config/genesis.json \\
    --rpcport 8545 \\
    --wsport 8546 \\
    --listen 0.0.0.0:30303 \\
    --loglevel info
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
StandardOutput=append:/var/log/gydschain/node.log
StandardError=append:/var/log/gydschain/node-error.log

[Install]
WantedBy=multi-user.target
EOF

# Indexer service
sudo tee /etc/systemd/system/gydschain-indexer.service > /dev/null <<EOF
[Unit]
Description=GYDS Chain Indexer
After=network.target postgresql.service gydschain-node.service

[Service]
Type=simple
User=gydschain
Group=gydschain
WorkingDirectory=/opt/gydschain
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=gydschain_indexer"
Environment="DB_USER=gydschain"
Environment="DB_PASSWORD=gydschain123"
Environment="RPC_URL=http://localhost:8545"
ExecStart=/opt/gydschain/bin/gydschain-indexer \\
    --api-port 8080 \\
    --log-level info
Restart=on-failure
RestartSec=10
StandardOutput=append:/var/log/gydschain/indexer.log
StandardError=append:/var/log/gydschain/indexer-error.log

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Setup Complete!                            ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo ""
echo "1. Build the project:"
echo "   cd /opt/gydschain"
echo "   git clone https://github.com/gydschain/gydschain.git ."
echo "   go build -o bin/gydschain ./cmd/node"
echo "   go build -o bin/gydschain-indexer ./cmd/indexer"
echo "   go build -o bin/gydschain-cli ./cmd/cli"
echo ""
echo "2. Copy configuration files:"
echo "   cp scripts/genesis.json /opt/gydschain/config/"
echo "   # Edit config.json as needed"
echo ""
echo "3. Initialize the database:"
echo "   psql -h localhost -U gydschain -d gydschain_indexer -f indexer/db/schema.sql"
echo ""
echo "4. Start the services:"
echo "   sudo systemctl enable gydschain-node"
echo "   sudo systemctl start gydschain-node"
echo "   sudo systemctl enable gydschain-indexer"
echo "   sudo systemctl start gydschain-indexer"
echo ""
echo "5. Check status:"
echo "   sudo systemctl status gydschain-node"
echo "   sudo systemctl status gydschain-indexer"
echo ""
echo "6. View logs:"
echo "   sudo tail -f /var/log/gydschain/node.log"
echo "   sudo tail -f /var/log/gydschain/indexer.log"
echo ""
echo -e "${BLUE}Ports:${NC}"
echo "  - P2P:       30303"
echo "  - RPC:       8545"
echo "  - WebSocket: 8546"
echo "  - Indexer:   8080"
echo ""
echo -e "${GREEN}Happy staking! 🚀${NC}"
