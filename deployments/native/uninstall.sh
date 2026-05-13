#!/usr/bin/env bash
set -euo pipefail

INSTALL_PATH="/usr/local/bin/conductor"
DATA_DIR="/var/lib/conductor"
DEFAULTS_FILE="/etc/default/conductor"
SERVICE_FILE="/etc/systemd/system/conductor.service"
SERVICE_USER="conductor"

# ---- prereqs ----------------------------------------------------------------

if [ "$(id -u)" -ne 0 ]; then
    echo "error: must be run as root (use sudo)" >&2
    exit 1
fi

# ---- stop and disable service -----------------------------------------------

if systemctl is-active --quiet conductor 2>/dev/null; then
    echo "Stopping conductor service..."
    systemctl stop conductor
fi

if systemctl is-enabled --quiet conductor 2>/dev/null; then
    systemctl disable conductor
fi

# ---- remove service file ----------------------------------------------------

if [ -f "$SERVICE_FILE" ]; then
    rm "$SERVICE_FILE"
    systemctl daemon-reload
fi

# ---- remove binary ----------------------------------------------------------

if [ -f "$INSTALL_PATH" ]; then
    rm "$INSTALL_PATH"
    echo "Removed $INSTALL_PATH"
fi

# ---- defaults file ----------------------------------------------------------

if [ -f "$DEFAULTS_FILE" ]; then
    rm "$DEFAULTS_FILE"
    echo "Removed $DEFAULTS_FILE"
fi

# ---- data directory ---------------------------------------------------------

if [ -d "$DATA_DIR" ]; then
    rm -rf "$DATA_DIR"
    echo "Removed $DATA_DIR"
fi

# ---- system user ------------------------------------------------------------

if id -u "$SERVICE_USER" &>/dev/null; then
    userdel "$SERVICE_USER"
    echo "Removed user $SERVICE_USER"
fi

echo ""
echo "conductor has been uninstalled."
