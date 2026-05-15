#!/usr/bin/env bash
set -euo pipefail

GITHUB_REPO="agiannif/conductor"
INSTALL_PATH="/usr/local/bin/conductor"
DATA_DIR="/var/lib/conductor"
DEFAULTS_FILE="/etc/default/conductor"
CADDYFILE="/etc/caddy/Caddyfile"
SERVICE_FILE="/etc/systemd/system/conductor.service"
SERVICE_USER="conductor"

# ---- prereqs ----------------------------------------------------------------

if [ "$(id -u)" -ne 0 ]; then
    echo "error: must be run as root (use sudo)" >&2
    exit 1
fi

for cmd in curl systemctl; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "error: required command not found: $cmd" >&2
        exit 1
    fi
done

# ---- arch detection ---------------------------------------------------------

case "$(uname -m)" in
    x86_64)  ARCH="linux-amd64"  ;;
    aarch64) ARCH="linux-arm64"  ;;
    armv7l)  ARCH="linux-armhf"  ;;
    *)
        echo "error: unsupported architecture: $(uname -m)" >&2
        exit 1
        ;;
esac

# ---- version resolution -----------------------------------------------------

if [ -z "${VERSION:-}" ]; then
    echo "Fetching latest release version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/$GITHUB_REPO/releases/latest" \
        | grep '"tag_name"' \
        | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
fi

if [ -z "$VERSION" ]; then
    echo "error: could not determine version — set VERSION=vX.Y.Z to pin one" >&2
    exit 1
fi

# ---- mode selection ---------------------------------------------------------
# Non-interactive override: set DOMAIN=yourdomain.com to skip prompts and
# go straight to HTTPS mode.

DOMAIN=${DOMAIN:-}
MODE=""

if [ -n "$DOMAIN" ]; then
    MODE=3
else
    echo ""
    echo "How should users access conductor?"
    echo ""
    echo "  1) http://hostname:8080   — simple, no extra config  [default]"
    echo "  2) http://hostname        — port 80, no port needed in the URL"
    echo "  3) https://yourdomain.com — HTTPS via Caddy (requires a public domain)"
    echo ""
    read -r -p "Choice [1]: " MODE </dev/tty
    MODE="${MODE:-1}"

    if [ "$MODE" = "3" ]; then
        echo ""
        read -r -p "Domain name (e.g. tasks.example.com): " DOMAIN </dev/tty
        if [ -z "$DOMAIN" ]; then
            echo "error: domain name is required for HTTPS mode" >&2
            exit 1
        fi
    fi
fi

case "$MODE" in
    1|2|3) ;;
    *)
        echo "error: invalid choice: $MODE" >&2
        exit 1
        ;;
esac

# ---- stop existing service (upgrade path) -----------------------------------

if systemctl is-active --quiet conductor 2>/dev/null; then
    echo "Stopping conductor service for upgrade..."
    systemctl stop conductor
fi

# ---- download binary --------------------------------------------------------

echo ""
echo "Installing conductor $VERSION ($ARCH)..."

DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/conductor-$ARCH"
echo "Downloading $DOWNLOAD_URL..."
curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

# ---- port 80: grant capability to bind privileged ports ---------------------
# setcap allows the binary to bind port 80 when executed directly (e.g. admin
# CLI). The systemd unit uses AmbientCapabilities for the same effect within
# the service — the two mechanisms cover each execution context.

if [ "$MODE" = "2" ]; then
    if ! command -v setcap &>/dev/null; then
        echo "error: setcap not found — install libcap2-bin (apt) or libcap (rpm)" >&2
        exit 1
    fi
    setcap cap_net_bind_service=+ep "$INSTALL_PATH"
fi

# ---- system user ------------------------------------------------------------

if ! id -u "$SERVICE_USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
fi

# ---- data directory ---------------------------------------------------------

mkdir -p "$DATA_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"

# ---- defaults file ----------------------------------------------------------

case "$MODE" in
    1)
        cat > "$DEFAULTS_FILE" << 'EOF'
CONDUCTOR_DB_PATH=/var/lib/conductor/conductor.db
CONDUCTOR_SECURE_COOKIE=false
CONDUCTOR_LISTEN_ADDR=:8080
EOF
        ;;
    2)
        cat > "$DEFAULTS_FILE" << 'EOF'
CONDUCTOR_DB_PATH=/var/lib/conductor/conductor.db
CONDUCTOR_SECURE_COOKIE=false
CONDUCTOR_LISTEN_ADDR=:80
EOF
        ;;
    3)
        cat > "$DEFAULTS_FILE" << 'EOF'
CONDUCTOR_DB_PATH=/var/lib/conductor/conductor.db
CONDUCTOR_SECURE_COOKIE=true
CONDUCTOR_LISTEN_ADDR=127.0.0.1:8080
EOF
        ;;
esac

# ---- systemd unit -----------------------------------------------------------
# Mode 2 (port 80) uses AmbientCapabilities instead of NoNewPrivileges —
# the two are incompatible. All other modes keep NoNewPrivileges for tighter
# sandboxing.

if [ "$MODE" = "2" ]; then
    cat > "$SERVICE_FILE" << 'EOF'
[Unit]
Description=Conductor task tracker
After=network.target

[Service]
Type=simple
User=conductor
Group=conductor
ExecStart=/usr/local/bin/conductor
Restart=on-failure
RestartSec=5

EnvironmentFile=/etc/default/conductor

AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

ProtectSystem=strict
ReadWritePaths=/var/lib/conductor
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
else
    cat > "$SERVICE_FILE" << 'EOF'
[Unit]
Description=Conductor task tracker
After=network.target

[Service]
Type=simple
User=conductor
Group=conductor
ExecStart=/usr/local/bin/conductor
Restart=on-failure
RestartSec=5

EnvironmentFile=/etc/default/conductor

NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/var/lib/conductor
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
fi

systemctl daemon-reload
systemctl enable conductor
systemctl restart conductor

# ---- caddy (mode 3 only) ----------------------------------------------------

if [ "$MODE" = "3" ]; then
    echo ""
    echo "Installing Caddy..."

    if ! command -v caddy &>/dev/null; then
        if ! command -v apt-get &>/dev/null; then
            echo "error: apt-get not found — install Caddy manually: https://caddyserver.com/docs/install" >&2
            exit 1
        fi
        apt-get install -y debian-keyring debian-archive-keyring apt-transport-https
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
            | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
            | tee /etc/apt/sources.list.d/caddy-stable.list
        apt-get update
        apt-get install -y caddy
    fi

    cat > "$CADDYFILE" << EOF
$DOMAIN {
    reverse_proxy 127.0.0.1:8080
}
EOF

    systemctl enable caddy
    systemctl reload caddy 2>/dev/null || systemctl restart caddy
fi

# ---- done -------------------------------------------------------------------

echo ""
echo "conductor $VERSION installed and running."
echo ""
echo "  Status:   systemctl status conductor"
echo "  Logs:     journalctl -u conductor -f"
echo "  Data:     $DATA_DIR"
echo ""

case "$MODE" in
    1) echo "  Access:   http://$(hostname):8080" ;;
    2) echo "  Access:   http://$(hostname)" ;;
    3) echo "  Access:   https://$DOMAIN  (certificate provisioned automatically)" ;;
esac

echo ""
echo "Create your first user:"
echo "  conductor admin add-user <username> <password>"
