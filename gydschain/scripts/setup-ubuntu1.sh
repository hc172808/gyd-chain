#!/bin/bash
set -e

#############################################
# GYDS CHAIN – UNIVERSAL NODE SETUP
# Ubuntu 22.04 LTS
#
# Node modes:
#   full  – full blockchain + rpc + p2p
#   lite  – rpc only (no mining)
#   sync  – sync only, no rpc
#
# Includes:
#   - Go 1.21.6 (local tarball supported)
#   - PostgreSQL 15 (indexer)
#   - WireGuard client
#   - SSH → HTTPS → ZIP repo fallback
#############################################

### COLORS
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

### GLOBALS
INSTALL_DIR="/opt/gydschain"
GO_BIN="/usr/local/go/bin/go"
GO_VERSION="1.21.6"
GO_TARBALL="go1.21.6.linux-amd64.tar.gz"

SSH_REPO="git@github.com:gydschain/gydschain.git"
HTTPS_REPO="https://github.com/gydschain/gydschain.git"
ZIP_URL="https://github.com/hc172808/gyd-chain/archive/refs/heads/main.zip"

#############################################
# BANNER
#############################################
echo -e "${GREEN}"
echo "╔══════════════════════════════════════════════════════════╗"
echo "║              GYDS CHAIN NODE INSTALLER                   ║"
echo "║              Ubuntu 22.04 LTS                            ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

#############################################
# NODE MODE
#############################################
echo -e "${BLUE}Select node type:${NC}"
echo "1) Full node"
echo "2) Lite node (RPC only)"
echo "3) Sync node (no RPC)"
read -p "Enter choice [1-3]: " NODE_CHOICE

case "$NODE_CHOICE" in
  1) NODE_MODE="full" ;;
  2) NODE_MODE="lite" ;;
  3) NODE_MODE="sync" ;;
  *) echo -e "${RED}Invalid choice${NC}"; exit 1 ;;
esac

#############################################
# SYSTEM PACKAGES
#############################################
echo -e "${BLUE}Updating system...${NC}"
apt update -y && apt upgrade -y

echo -e "${BLUE}Installing dependencies...${NC}"
apt install -y \
  build-essential git curl wget unzip jq \
  libssl-dev libleveldb-dev liblz4-dev \
  libsnappy-dev libzstd-dev \
  ufw fail2ban screen htop \
  postgresql postgresql-client \
  wireguard resolvconf

#############################################
# GO INSTALL
#############################################
echo -e "${BLUE}Installing Go ${GO_VERSION}...${NC}"

if [ -f "$GO_TARBALL" ]; then
  echo -e "${GREEN}Using local Go tarball${NC}"
else
  wget https://go.dev/dl/$GO_TARBALL
fi

rm -rf /usr/local/go
tar -C /usr/local -xzf $GO_TARBALL

echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' > /etc/profile.d/go.sh
chmod +x /etc/profile.d/go.sh

$GO_BIN version

#############################################
# USER & DIRS
#############################################
id gydschain &>/dev/null || useradd -m -s /bin/bash gydschain

mkdir -p $INSTALL_DIR/{bin,data,config,logs}
mkdir -p /var/log/gydschain
chown -R gydschain:gydschain $INSTALL_DIR /var/log/gydschain

#############################################
# POSTGRES
#############################################
echo -e "${BLUE}Configuring PostgreSQL...${NC}"
systemctl enable postgresql
systemctl start postgresql

sudo -u postgres psql <<EOF || true
CREATE USER gydschain WITH PASSWORD 'gydschain123';
CREATE DATABASE gydschain_indexer OWNER gydschain;
GRANT ALL PRIVILEGES ON DATABASE gydschain_indexer TO gydschain;
EOF

#############################################
# FIREWALL
#############################################
ufw allow ssh
ufw allow 30303/tcp
ufw allow 30303/udp
ufw allow 8545/tcp
ufw allow 8546/tcp
ufw --force enable

#############################################
# CLONE / UPDATE CODE
#############################################
echo -e "${BLUE}Installing GYDS Chain source...${NC}"

clone_repo() {
  git clone "$1" $INSTALL_DIR
}

download_zip() {
  rm -rf $INSTALL_DIR
  wget -O /tmp/gyds.zip "$ZIP_URL"
  unzip /tmp/gyds.zip -d /opt
  mv /opt/gyd-chain-main $INSTALL_DIR
}

if [ -d "$INSTALL_DIR/.git" ]; then
  echo -e "${YELLOW}Repo exists. Pulling updates...${NC}"
  git -C $INSTALL_DIR pull || download_zip
else
  clone_repo "$SSH_REPO" || clone_repo "$HTTPS_REPO" || download_zip
fi

chown -R gydschain:gydschain $INSTALL_DIR

#############################################
# BUILD BINARIES
#############################################
echo -e "${BLUE}Building binaries...${NC}"

sudo -u gydschain bash <<EOF
cd $INSTALL_DIR
$GO_BIN build -o bin/gydschain ./cmd/node
$GO_BIN build -o bin/gydschain-indexer ./cmd/indexer
$GO_BIN build -o bin/gydschain-cli ./cmd/cli
EOF

#############################################
# WIREGUARD CLIENT
#############################################
read -p "Install WireGuard client config? (y/N): " WG
if [[ "$WG" =~ ^[Yy]$ ]]; then
  mkdir -p /etc/wireguard
  read -p "Paste WireGuard config filename (wg0.conf): " WGCONF
  nano /etc/wireguard/$WGCONF
  systemctl enable wg-quick@$WGCONF
  systemctl start wg-quick@$WGCONF
fi

#############################################
# SYSTEMD SERVICE
#############################################
echo -e "${BLUE}Creating systemd service...${NC}"

cat >/etc/systemd/system/gydschain.service <<EOF
[Unit]
Description=GYDS Chain Node ($NODE_MODE)
After=network.target

[Service]
User=gydschain
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/gydschain \\
  --datadir $INSTALL_DIR/data \\
  --config $INSTALL_DIR/config/config.json \\
  --mode $NODE_MODE
Restart=always
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable gydschain
systemctl restart gydschain

#############################################
# DONE
#############################################
echo -e "${GREEN}"
echo "╔══════════════════════════════════════════════════════════╗"
echo "║          GYDS CHAIN NODE INSTALL COMPLETE               ║"
echo "║          MODE: $NODE_MODE                               ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo "Check status:"
echo "  systemctl status gydschain"
echo ""
echo "View logs:"
echo "  journalctl -u gydschain -f"
