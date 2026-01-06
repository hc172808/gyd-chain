#!/bin/bash

# GYDS Chain Ubuntu 22.04 Setup Script
# Supports: Full Node, Lite Node, Frontend, and Admin Panel

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
GITHUB_REPO="https://github.com/gydschain/gydschain.git"
INSTALL_DIR="/opt/gydschain"
FRONTEND_DIR="/var/www/gydschain"
CONFIG_DIR="/opt/gydschain/config"
LOG_DIR="/var/log/gydschain"
NODE_REGISTRY_FILE="/opt/gydschain/config/node_registry.json"
VPN_CONFIG_DIR="/etc/wireguard"

print_banner() {
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                  GYDS Chain Setup Script                      ║"
    echo "║                     Ubuntu 22.04 LTS                          ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

show_menu() {
    echo -e "${CYAN}Select installation type:${NC}"
    echo ""
    echo "  1) Full Node (Validator/Mining Node)"
    echo "  2) Lite Node (Light Client - Syncs with network)"
    echo "  3) Admin Panel (Full setup with dashboard)"
    echo "  4) Frontend Only (Web interface)"
    echo "  5) Update/Rebuild from GitHub"
    echo "  6) Exit"
    echo ""
    read -p "Enter choice [1-6]: " choice
}

# Common system updates
install_base_packages() {
    echo -e "${BLUE}[1/8] Updating system packages...${NC}"
    sudo apt-get update -y
    sudo apt-get upgrade -y

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
        fail2ban \
        nginx \
        certbot \
        python3-certbot-nginx
}

install_go() {
    echo -e "${BLUE}Installing Go 1.21...${NC}"
    GO_VERSION="1.21.6"
    if ! command -v go &> /dev/null || [[ $(go version | awk '{print $3}') != "go$GO_VERSION" ]]; then
        wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
        rm "go${GO_VERSION}.linux-amd64.tar.gz"
        
        echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
        echo 'export GOPATH=$HOME/go' | sudo tee -a /etc/profile.d/go.sh
        echo 'export PATH=$PATH:$GOPATH/bin' | sudo tee -a /etc/profile.d/go.sh
        source /etc/profile.d/go.sh
    fi
    echo -e "${GREEN}Go version: $(go version)${NC}"
}

install_nodejs() {
    echo -e "${BLUE}Installing Node.js 20 LTS...${NC}"
    if ! command -v node &> /dev/null; then
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
    echo -e "${GREEN}Node.js version: $(node --version)${NC}"
    echo -e "${GREEN}NPM version: $(npm --version)${NC}"
}

install_postgresql() {
    echo -e "${BLUE}Installing PostgreSQL 15...${NC}"
    if ! command -v psql &> /dev/null; then
        sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
        wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
        sudo apt-get update
        sudo apt-get install -y postgresql-15 postgresql-client-15
    fi
    sudo systemctl enable postgresql
    sudo systemctl start postgresql

    sudo -u postgres psql -c "CREATE USER gydschain WITH PASSWORD 'gydschain123';" 2>/dev/null || true
    sudo -u postgres psql -c "CREATE DATABASE gydschain_indexer OWNER gydschain;" 2>/dev/null || true
    sudo -u postgres psql -c "CREATE DATABASE gydschain_admin OWNER gydschain;" 2>/dev/null || true
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gydschain_indexer TO gydschain;" 2>/dev/null || true
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gydschain_admin TO gydschain;" 2>/dev/null || true
}

install_wireguard() {
    echo -e "${BLUE}Installing WireGuard VPN...${NC}"
    sudo apt-get install -y wireguard wireguard-tools
    
    # Generate server keys if not exists
    if [ ! -f "$VPN_CONFIG_DIR/server_private.key" ]; then
        sudo mkdir -p $VPN_CONFIG_DIR
        wg genkey | sudo tee $VPN_CONFIG_DIR/server_private.key | wg pubkey | sudo tee $VPN_CONFIG_DIR/server_public.key
        sudo chmod 600 $VPN_CONFIG_DIR/server_private.key
    fi
}

setup_directories() {
    echo -e "${BLUE}Creating directories...${NC}"
    
    if ! id "gydschain" &>/dev/null; then
        sudo useradd -m -s /bin/bash gydschain
    fi

    sudo mkdir -p $INSTALL_DIR/{bin,data,config,logs,nodes}
    sudo mkdir -p $FRONTEND_DIR
    sudo mkdir -p $LOG_DIR
    sudo mkdir -p /opt/gydschain/nodes/pending
    sudo mkdir -p /opt/gydschain/nodes/approved
    
    # Initialize node registry
    if [ ! -f "$NODE_REGISTRY_FILE" ]; then
        echo '{"pending":[],"approved":[],"rejected":[]}' | sudo tee $NODE_REGISTRY_FILE > /dev/null
    fi
    
    sudo chown -R gydschain:gydschain $INSTALL_DIR
    sudo chown -R www-data:www-data $FRONTEND_DIR
    sudo chown -R gydschain:gydschain $LOG_DIR
}

configure_firewall() {
    echo -e "${BLUE}Configuring firewall...${NC}"
    sudo ufw allow ssh
    sudo ufw allow 80/tcp      # HTTP
    sudo ufw allow 443/tcp     # HTTPS
    sudo ufw allow 30303/tcp   # P2P
    sudo ufw allow 30303/udp   # P2P discovery
    sudo ufw allow 8545/tcp    # RPC
    sudo ufw allow 8546/tcp    # WebSocket
    sudo ufw allow 8080/tcp    # Indexer API
    sudo ufw allow 51820/udp   # WireGuard VPN
    echo "y" | sudo ufw enable || true
}

clone_and_build_backend() {
    echo -e "${BLUE}Cloning and building backend...${NC}"
    
    cd $INSTALL_DIR
    if [ -d ".git" ]; then
        sudo -u gydschain git pull origin main
    else
        sudo -u gydschain git clone $GITHUB_REPO .
    fi
    
    # Build Go binaries
    sudo -u gydschain go build -o bin/gydschain ./cmd/node
    sudo -u gydschain go build -o bin/gydschain-indexer ./cmd/indexer
    sudo -u gydschain go build -o bin/gydschain-cli ./cmd/cli
    sudo -u gydschain go build -o bin/gydschain-miner ./cmd/miner
    sudo -u gydschain go build -o bin/gydschain-litenode ./cmd/litenode
    
    echo -e "${GREEN}Backend built successfully!${NC}"
}

build_frontend() {
    echo -e "${BLUE}Building frontend...${NC}"
    
    cd $INSTALL_DIR
    
    # Install dependencies and build
    npm install
    npm run build
    
    # Copy to web directory
    sudo rm -rf $FRONTEND_DIR/*
    sudo cp -r dist/* $FRONTEND_DIR/
    sudo chown -R www-data:www-data $FRONTEND_DIR
    
    echo -e "${GREEN}Frontend built and deployed!${NC}"
}

setup_nginx() {
    echo -e "${BLUE}Configuring Nginx...${NC}"
    
    sudo tee /etc/nginx/sites-available/gydschain > /dev/null <<EOF
server {
    listen 80;
    server_name _;
    root $FRONTEND_DIR;
    index index.html;

    # Frontend
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://127.0.0.1:8080/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_cache_bypass \$http_upgrade;
    }

    # RPC proxy
    location /rpc/ {
        proxy_pass http://127.0.0.1:8545/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    # WebSocket proxy
    location /ws/ {
        proxy_pass http://127.0.0.1:8546/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
    }

    # Admin API
    location /admin-api/ {
        proxy_pass http://127.0.0.1:9000/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF

    sudo ln -sf /etc/nginx/sites-available/gydschain /etc/nginx/sites-enabled/
    sudo rm -f /etc/nginx/sites-enabled/default
    sudo nginx -t && sudo systemctl reload nginx
}

create_systemd_services() {
    echo -e "${BLUE}Creating systemd services...${NC}"

    # Full Node service
    sudo tee /etc/systemd/system/gydschain-node.service > /dev/null <<EOF
[Unit]
Description=GYDS Chain Full Node
After=network.target

[Service]
Type=simple
User=gydschain
Group=gydschain
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/gydschain \\
    --datadir $INSTALL_DIR/data \\
    --config $CONFIG_DIR/config.json \\
    --genesis $CONFIG_DIR/genesis.json \\
    --rpcport 8545 \\
    --wsport 8546 \\
    --listen 0.0.0.0:30303 \\
    --loglevel info
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
StandardOutput=append:$LOG_DIR/node.log
StandardError=append:$LOG_DIR/node-error.log

[Install]
WantedBy=multi-user.target
EOF

    # Lite Node service
    sudo tee /etc/systemd/system/gydschain-litenode.service > /dev/null <<EOF
[Unit]
Description=GYDS Chain Lite Node
After=network.target

[Service]
Type=simple
User=gydschain
Group=gydschain
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/gydschain-litenode \\
    --datadir $INSTALL_DIR/data/lite \\
    --config $CONFIG_DIR/litenode.json \\
    --sync-mode light \\
    --bootstrap-nodes $CONFIG_DIR/bootstrap.json
Restart=on-failure
RestartSec=10
StandardOutput=append:$LOG_DIR/litenode.log
StandardError=append:$LOG_DIR/litenode-error.log

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
WorkingDirectory=$INSTALL_DIR
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=gydschain_indexer"
Environment="DB_USER=gydschain"
Environment="DB_PASSWORD=gydschain123"
Environment="RPC_URL=http://localhost:8545"
ExecStart=$INSTALL_DIR/bin/gydschain-indexer \\
    --api-port 8080 \\
    --log-level info
Restart=on-failure
RestartSec=10
StandardOutput=append:$LOG_DIR/indexer.log
StandardError=append:$LOG_DIR/indexer-error.log

[Install]
WantedBy=multi-user.target
EOF

    # Admin API service
    sudo tee /etc/systemd/system/gydschain-admin.service > /dev/null <<EOF
[Unit]
Description=GYDS Chain Admin API
After=network.target postgresql.service

[Service]
Type=simple
User=gydschain
Group=gydschain
WorkingDirectory=$INSTALL_DIR
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=gydschain_admin"
Environment="DB_USER=gydschain"
Environment="DB_PASSWORD=gydschain123"
Environment="NODE_REGISTRY=$NODE_REGISTRY_FILE"
Environment="VPN_CONFIG_DIR=$VPN_CONFIG_DIR"
ExecStart=$INSTALL_DIR/bin/gydschain-admin \\
    --port 9000 \\
    --log-level info
Restart=on-failure
RestartSec=10
StandardOutput=append:$LOG_DIR/admin.log
StandardError=append:$LOG_DIR/admin-error.log

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
}

# ============ LITE NODE REGISTRATION ============
register_litenode() {
    echo -e "${BLUE}Registering Lite Node with Admin...${NC}"
    
    # Generate node keypair
    NODE_ID=$(cat /dev/urandom | tr -dc 'a-f0-9' | fold -w 64 | head -n 1)
    HOSTNAME=$(hostname)
    PUBLIC_IP=$(curl -s ifconfig.me)
    
    # Generate WireGuard keys for this node
    PRIVATE_KEY=$(wg genkey)
    PUBLIC_KEY=$(echo $PRIVATE_KEY | wg pubkey)
    
    # Create node info file
    NODE_INFO=$(cat <<EOF
{
    "node_id": "$NODE_ID",
    "hostname": "$HOSTNAME",
    "public_ip": "$PUBLIC_IP",
    "wireguard_public_key": "$PUBLIC_KEY",
    "registered_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "status": "pending",
    "type": "litenode"
}
EOF
)
    
    echo "$NODE_INFO" | sudo tee $CONFIG_DIR/node_info.json > /dev/null
    echo "$PRIVATE_KEY" | sudo tee $CONFIG_DIR/wireguard_private.key > /dev/null
    sudo chmod 600 $CONFIG_DIR/wireguard_private.key
    
    # Send registration to admin server
    read -p "Enter Admin Server URL (e.g., https://admin.gydschain.io): " ADMIN_URL
    
    RESPONSE=$(curl -s -X POST "$ADMIN_URL/admin-api/nodes/register" \
        -H "Content-Type: application/json" \
        -d "$NODE_INFO")
    
    if echo "$RESPONSE" | grep -q "success"; then
        echo -e "${GREEN}✅ Node registered successfully! Waiting for admin approval...${NC}"
        echo -e "${YELLOW}Your Node ID: $NODE_ID${NC}"
        echo -e "${YELLOW}Once approved, run: sudo ./setup-ubuntu.sh and select 'Activate Lite Node'${NC}"
    else
        echo -e "${RED}❌ Registration failed: $RESPONSE${NC}"
    fi
}

activate_litenode() {
    echo -e "${BLUE}Activating Lite Node...${NC}"
    
    if [ ! -f "$CONFIG_DIR/node_info.json" ]; then
        echo -e "${RED}Node not registered. Run registration first.${NC}"
        return 1
    fi
    
    NODE_ID=$(jq -r '.node_id' $CONFIG_DIR/node_info.json)
    
    read -p "Enter Admin Server URL: " ADMIN_URL
    
    # Check approval status and get VPN config
    RESPONSE=$(curl -s "$ADMIN_URL/admin-api/nodes/$NODE_ID/config")
    
    if echo "$RESPONSE" | grep -q "approved"; then
        # Extract VPN config
        VPN_CONFIG=$(echo "$RESPONSE" | jq -r '.vpn_config')
        BOOTSTRAP_NODES=$(echo "$RESPONSE" | jq -r '.bootstrap_nodes')
        
        # Setup WireGuard
        echo "$VPN_CONFIG" | sudo tee /etc/wireguard/wg0.conf > /dev/null
        sudo chmod 600 /etc/wireguard/wg0.conf
        
        # Save bootstrap nodes
        echo "$BOOTSTRAP_NODES" | sudo tee $CONFIG_DIR/bootstrap.json > /dev/null
        
        # Start WireGuard and Lite Node
        sudo systemctl enable wg-quick@wg0
        sudo systemctl start wg-quick@wg0
        sudo systemctl enable gydschain-litenode
        sudo systemctl start gydschain-litenode
        
        echo -e "${GREEN}✅ Lite Node activated and syncing!${NC}"
    else
        echo -e "${YELLOW}Node not yet approved. Contact admin.${NC}"
    fi
}

# ============ INSTALLATION FUNCTIONS ============
install_full_node() {
    echo -e "${GREEN}Installing Full Node...${NC}"
    install_base_packages
    install_go
    install_postgresql
    setup_directories
    configure_firewall
    clone_and_build_backend
    create_systemd_services
    
    # Copy config files
    sudo cp $INSTALL_DIR/scripts/genesis.json $CONFIG_DIR/
    
    # Initialize database
    psql -h localhost -U gydschain -d gydschain_indexer -f $INSTALL_DIR/indexer/db/schema.sql
    
    echo -e "${GREEN}Full Node installed! Start with: sudo systemctl start gydschain-node${NC}"
}

install_litenode() {
    echo -e "${GREEN}Installing Lite Node...${NC}"
    install_base_packages
    install_go
    install_wireguard
    setup_directories
    configure_firewall
    
    # Build only litenode binary
    cd $INSTALL_DIR
    if [ -d ".git" ]; then
        sudo -u gydschain git pull origin main
    else
        sudo -u gydschain git clone $GITHUB_REPO .
    fi
    sudo -u gydschain go build -o bin/gydschain-litenode ./cmd/litenode
    
    create_systemd_services
    
    echo ""
    echo -e "${YELLOW}Lite Node installed. Now register with admin:${NC}"
    register_litenode
}

install_admin_panel() {
    echo -e "${GREEN}Installing Admin Panel with Full Stack...${NC}"
    install_base_packages
    install_go
    install_nodejs
    install_postgresql
    install_wireguard
    setup_directories
    configure_firewall
    clone_and_build_backend
    build_frontend
    setup_nginx
    create_systemd_services
    
    # Initialize databases
    psql -h localhost -U gydschain -d gydschain_indexer -f $INSTALL_DIR/indexer/db/schema.sql
    psql -h localhost -U gydschain -d gydschain_admin -f $INSTALL_DIR/admin/db/schema.sql
    
    # Start all services
    sudo systemctl enable gydschain-node gydschain-indexer gydschain-admin nginx
    sudo systemctl start gydschain-node gydschain-indexer gydschain-admin nginx
    
    echo -e "${GREEN}Admin Panel installed and running!${NC}"
}

install_frontend_only() {
    echo -e "${GREEN}Installing Frontend Only...${NC}"
    install_base_packages
    install_nodejs
    setup_directories
    
    cd $INSTALL_DIR
    if [ -d ".git" ]; then
        sudo -u gydschain git pull origin main
    else
        sudo -u gydschain git clone $GITHUB_REPO .
    fi
    
    build_frontend
    setup_nginx
    
    sudo systemctl enable nginx
    sudo systemctl start nginx
    
    echo -e "${GREEN}Frontend deployed!${NC}"
}

update_and_rebuild() {
    echo -e "${GREEN}Updating from GitHub and rebuilding...${NC}"
    
    cd $INSTALL_DIR
    
    # Stop services
    sudo systemctl stop gydschain-node gydschain-indexer gydschain-admin 2>/dev/null || true
    
    # Pull latest
    sudo -u gydschain git fetch origin
    sudo -u gydschain git reset --hard origin/main
    
    # Rebuild backend
    sudo -u gydschain go build -o bin/gydschain ./cmd/node
    sudo -u gydschain go build -o bin/gydschain-indexer ./cmd/indexer
    sudo -u gydschain go build -o bin/gydschain-cli ./cmd/cli
    sudo -u gydschain go build -o bin/gydschain-miner ./cmd/miner
    sudo -u gydschain go build -o bin/gydschain-litenode ./cmd/litenode
    sudo -u gydschain go build -o bin/gydschain-admin ./cmd/admin
    
    # Rebuild frontend
    npm install
    npm run build
    sudo rm -rf $FRONTEND_DIR/*
    sudo cp -r dist/* $FRONTEND_DIR/
    sudo chown -R www-data:www-data $FRONTEND_DIR
    
    # Restart services
    sudo systemctl start gydschain-node gydschain-indexer gydschain-admin
    
    echo -e "${GREEN}Update complete! All services restarted.${NC}"
}

# ============ MAIN ============
print_banner

if [ "$1" == "--register-litenode" ]; then
    register_litenode
    exit 0
fi

if [ "$1" == "--activate-litenode" ]; then
    activate_litenode
    exit 0
fi

if [ "$1" == "--update" ]; then
    update_and_rebuild
    exit 0
fi

show_menu

case $choice in
    1)
        install_full_node
        ;;
    2)
        install_litenode
        ;;
    3)
        install_admin_panel
        ;;
    4)
        install_frontend_only
        ;;
    5)
        update_and_rebuild
        ;;
    6)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Setup Complete!                            ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Service Commands:${NC}"
echo "  sudo systemctl status gydschain-node"
echo "  sudo systemctl status gydschain-litenode"
echo "  sudo systemctl status gydschain-indexer"
echo "  sudo systemctl status gydschain-admin"
echo ""
echo -e "${BLUE}Log Files:${NC}"
echo "  sudo tail -f $LOG_DIR/node.log"
echo "  sudo tail -f $LOG_DIR/litenode.log"
echo ""
echo -e "${GREEN}Happy staking! 🚀${NC}"
