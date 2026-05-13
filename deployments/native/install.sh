#!/usr/bin/env bash
set -euo pipefail

GITHUB_REPO="agiannif/conductor"
INSTALL_PATH="/usr/local/bin/conductor"
DATA_DIR="/var/lib/conductor"
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

echo "Installing conductor $VERSION ($ARCH)..."

# ---- download binary --------------------------------------------------------

DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/conductor-$ARCH"
echo "Downloading $DOWNLOAD_URL"
curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

# ---- system user ------------------------------------------------------------

if ! id -u "$SERVICE_USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
fi

# ---- data directory ---------------------------------------------------------

mkdir -p "$DATA_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"

# ---- defaults file ----------------------------------------------------------
# Written to /etc/default/conductor so that the binary reads the correct DB
# path when run directly (e.g. `sudo conductor admin add-user ...`), not only
# when started by systemd.

cat > /etc/default/conductor << 'EOF'
CONDUCTOR_DB_PATH=/var/lib/conductor/conductor.db
CONDUCTOR_SECURE_COOKIE=false
EOF

# ---- systemd unit -----------------------------------------------------------

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

systemctl daemon-reload
systemctl enable conductor
systemctl restart conductor

# ---- done -------------------------------------------------------------------

echo ""
echo "conductor $VERSION installed and running."
echo ""
echo "  Status:   systemctl status conductor"
echo "  Logs:     journalctl -u conductor -f"
echo "  Data:     $DATA_DIR"
echo ""
echo "Create your first user:"
echo "  conductor admin add-user <username> <password>"
