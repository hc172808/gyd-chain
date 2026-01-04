#!/bin/bash

# GYDS Chain Ubuntu 22.04 Setup Script
# Fully automated: installs dependencies, Go, PostgreSQL, builds binaries, sets up systemd

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
echo -e "${BLUE}[1/9] Updating system packages...${NC}"
sudo apt-get update -y
sudo apt-get upgrade -y

# Install essential packages
echo -e "${BLUE}[2/9] Installing essential packages...${NC}"
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
    fail2ban \
    postgresql-15 \
    postgresql-client-15

# Install Go
echo -e "${BLUE}[3/9] Installing Go 1.21...${NC}"
GO_VERSION="1.21.6"
GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"

if ! command -v go &> /dev/null || [[ $(go version | awk '{print $3}') != "go$GO_VERSION" ]]; then
    if [ ! -f "$GO_TARBALL" ]; then
        echo -e "${YELLOW}Go tarball not found locally, downloading...${NC}"
        wget "https://go.dev/dl/${GO_TARBALL}"
    else
        echo -e "${GREEN}Found local Go tarball: $GO_TARBALL${NC}"
    fi

    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_TARBALL"

    # Add Go to PATH system-wide
    echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' | sudo tee /etc/profile.d/go.sh
    echo 'export GOPATH=$HOME/go' | sudo tee -a /etc/profile.d/go.sh
    source /etc/profile.d/go.sh
fi
echo -e "${GREEN}Go version: $(go version)${NC}"

# Create gydschain user and directories
echo -e "${BLUE}[4/9] Creating gydschain user and directories...${NC}"
if ! id "gydschain" &>/dev/null; then
    sudo useradd -m -s /bin/bash gydschain
fi

sudo mkdir -p /opt/gydschain/{bin,data,config,logs}
sudo mkdir -p /var/log/gydschain
sudo chown -R gydschain:gydschain /opt/gydschain
sudo chown -R gydschain:gydschain /var/log/gydschain

# Configure PostgreSQL database
echo -e "${BLUE}[5/9] Configuring PostgreSQL...${NC}"
sudo systemctl enable postgresql
sudo systemctl start postgresql

sudo -u postgres psql -c "CREATE USER gydschain WITH PASSWORD 'gydschain123';" 2>/dev/null || true
sudo -u postgres psql -c "CREATE DATABASE gydschain_indexer OWNER gydschain;" 2>/dev/null || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gydschain_indexer TO gydschain;" 2>/dev/null || true
echo -e "${GREEN}Database created: gydschain_indexer${NC}"

# Configure firewall
echo -e "${BLUE}[6/9] Configuring firewall...${NC}"
sudo ufw allow ssh
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp
sudo ufw allow 8545/tcp
sudo ufw allow 8546/tcp
sudo ufw allow 8080/tcp
sudo ufw --force enable
echo -e "${GREEN}Firewall configured${NC}"

# Clone and build GYDS Chain (if not already cloned)
echo -e "${BLUE}[7/9] Cloning and building GYDS Chain...${NC}"
if [ ! -d "/opt/gydschain/.git" ]; then
    sudo -u gydschain git clone https://github.com/gydschain/gydschain.git /opt/gydschain
else
    echo -e "${YELLOW}Repository already exists, pulling latest changes...${NC}"
    sudo -u gydschain git -C /opt/gydschain pull
fi

cd /opt/gydschain
sudo -u gydschain go build -o bin/gydschain ./cmd/node
sudo -u gydschain go build -o bin/gydschain-indexer ./cmd/indexer
sudo -u gydschain go build -o bin/gydschain-cli ./cmd/cli
echo -e "${GREEN}Binaries built successfully${NC}"

# Create systemd service files
echo -e "${BLUE}[8/9] Creating systemd service files...${NC}"

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
StandardOutput=file:/var/log/gydschain/node.log
StandardError=file:/var/log/gydschain/node-error.log

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
StandardOutput=file:/var/log/gydschain/indexer.log
StandardError=file:/var/log/gydschain/indexer-error.log

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable services
sudo systemctl daemon-reload
sudo systemctl enable gydschain-node
sudo systemctl start gydschain-node
sudo systemctl enable gydschain-indexer
sudo systemctl start gydschain-indexer

echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Setup Complete!                            ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"

echo -e "${YELLOW}Next steps:${NC}"
echo ""
echo "1. Copy configuration files if needed:"
echo "   cp scripts/genesis.json /opt/gydschain/config/"
echo "   # Edit config.json as needed"
echo ""
echo "2. Initialize the database (if not already done):"
echo "   psql -h localhost -U gydschain -d gydschain_indexer -f indexer/db/schema.sql"
echo ""
echo "3. Check services status:"
echo "   sudo systemctl status gydschain-node"
echo "   sudo systemctl status gydschain-indexer"
echo ""
echo "4. View logs:"
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
