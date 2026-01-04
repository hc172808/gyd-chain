#!/bin/bash
set -e

# ================= CONFIG =================
APP_USER="gydschain"
APP_DIR="/opt/gydschain"
BIN_DIR="$APP_DIR/bin"
LOG_DIR="/var/log/gydschain"
ENV_FILE="$APP_DIR/.env"
VPN_DIR="$APP_DIR/vpn"

GO_VERSION="1.21.6"
GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"
GO_LOCAL_PATH="/tmp/$GO_ARCHIVE"
GO_URL="https://go.dev/dl/$GO_ARCHIVE"
FULL_NODES_LIST="$APP_DIR/fullnodes.txt"
# ==========================================

# ---------- Root check ----------
if [ "$EUID" -ne 0 ]; then
    echo "❌ Run as root"
    exit 1
fi

# ---------- Helper functions ----------
install_go() {
    echo "⬇️ Installing Go $GO_VERSION..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "$GO_LOCAL_PATH"
}

ensure_go() {
    if command -v go >/dev/null 2>&1 && go version | grep -q "go$GO_VERSION"; then
        echo "✅ Go $GO_VERSION already installed"
    else
        if [ ! -f "$GO_LOCAL_PATH" ]; then
            echo "📥 Go archive not found locally, downloading..."
            wget -O "$GO_LOCAL_PATH" "$GO_URL"
        else
            echo "📦 Using local Go archive: $GO_LOCAL_PATH"
        fi
        install_go
    fi
    export PATH=$PATH:/usr/local/go/bin
}

ensure_dirs() {
    mkdir -p "$APP_DIR" "$BIN_DIR" "$LOG_DIR" "$VPN_DIR"
    chown -R $APP_USER:$APP_USER "$APP_DIR" "$LOG_DIR" "$VPN_DIR"
}

update_repo() {
    echo "📦 Pulling latest repo..."
    read -p "Enter branch/tag/commit to pull (default: main): " GIT_REF
    GIT_REF=${GIT_REF:-main}

    cd "$APP_DIR"
    if [ ! -d ".git" ]; then
        echo "⚠️ Repo not initialized, cloning..."
        git clone https://github.com/hc172808/gyd-chain.git "$APP_DIR"
        cd "$APP_DIR"
    fi

    git fetch --all
    git reset --hard "origin/$GIT_REF"
    git clean -fd
    echo "✅ Repo updated to $GIT_REF"
}

build_binaries() {
    echo "🔨 Building node & indexer..."
    ensure_go

    for module_dir in gydschain/cmd/node gydschain/cmd/indexer; do
        if [ ! -f "$module_dir/go.mod" ]; then
            echo "⚙️ Initializing Go module in $module_dir"
            sudo -u $APP_USER bash -c "cd $module_dir && go mod init github.com/hc172808/gyd-chain/$module_dir && go mod tidy"
        fi
    done

    sudo -u $APP_USER bash <<EOF
export PATH=$PATH
cd $APP_DIR
go build -o $BIN_DIR/gydschain-node ./gydschain/cmd/node
go build -o $BIN_DIR/gydschain-indexer ./gydschain/cmd/indexer
EOF
    echo "✅ Build complete"
}

configure_node_type() {
    echo "📝 Select node type:"
    echo "1) Full node (private, admin only)"
    echo "2) Lite node (public, connects to full nodes via VPN)"
    read -p "Enter choice [1-2]: " NODE_TYPE

    if [[ "$NODE_TYPE" == "1" ]]; then
        NODE_TYPE_NAME="full"
        PRIVATE_NODE=true
    else
        NODE_TYPE_NAME="lite"
        PRIVATE_NODE=false
    fi

    # Update .env
    cat >"$ENV_FILE" <<EOF
# GYDS Chain Environment
RPC_PORT=8545
P2P_PORT=30303
INDEXER_PORT=8080
PRIVATE_NODE=$PRIVATE_NODE
EOF

    chown $APP_USER:$APP_USER "$ENV_FILE"
    chmod 600 "$ENV_FILE"

    echo "✅ Node type configured: $NODE_TYPE_NAME"
}

setup_vpn_lite() {
    if [[ "$PRIVATE_NODE" == "false" ]]; then
        echo "🌐 Setting up WireGuard VPN for lite node..."

        if [ ! -f "$FULL_NODES_LIST" ]; then
            echo "⚠️ Full nodes list not found. Please create $FULL_NODES_LIST with one IP/domain per line."
            return
        fi

        WG_CONFIG="$VPN_DIR/wg0.conf"
        cat > "$WG_CONFIG" <<EOF
[Interface]
PrivateKey = $(wg genkey)
Address = 10.0.0.2/24
DNS = 10.0.0.1
EOF

        echo "[Peers]" >> "$WG_CONFIG"

        while read -r FULL_NODE; do
            echo "[Peer]" >> "$WG_CONFIG"
            echo "PublicKey = <FULL_NODE_PUBLIC_KEY>" >> "$WG_CONFIG"
            echo "Endpoint = $FULL_NODE:51820" >> "$WG_CONFIG"
            echo "AllowedIPs = 0.0.0.0/0" >> "$WG_CONFIG"
            echo "PersistentKeepalive = 25" >> "$WG_CONFIG"
            echo "" >> "$WG_CONFIG"
        done < "$FULL_NODES_LIST"

        chown $APP_USER:$APP_USER "$WG_CONFIG"
        chmod 600 "$WG_CONFIG"

        echo "✅ WireGuard config created at $WG_CONFIG"
        echo "Use: sudo wg-quick up $WG_CONFIG to start the VPN"
    fi
}

restart_services() {
    echo "🔁 Restarting systemd services..."
    for svc in gydschain-node gydschain-indexer; do
        if systemctl list-unit-files | grep -q "$svc"; then
            systemctl restart $svc
        fi
    done
    echo "✅ Services restarted"
}

# ==================== MAIN MENU ====================
while true; do
    echo ""
    echo "======================================="
    echo "     GYDS CHAIN GLOBAL DEPLOY MENU      "
    echo "======================================="
    echo "1) Pull latest code"
    echo "2) Build node & indexer"
    echo "3) Configure node type (full/lite)"
    echo "4) Setup WireGuard VPN for lite nodes"
    echo "5) Restart services"
    echo "6) Exit"
    echo "======================================="
    read -p "Enter choice [1-6]: " CHOICE

    case "$CHOICE" in
        1) update_repo ;;
        2) build_binaries ;;
        3) configure_node_type ;;
        4) setup_vpn_lite ;;
        5) restart_services ;;
        6) echo "Exiting..."; break ;;
        *) echo "⚠️ Invalid choice";;
    esac
done
