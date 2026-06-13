#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF_SRC="$SCRIPT_DIR/../nginx.conf"
SITE_NAME="stag-api.infiniour.com"
SITES_AVAIL="/etc/nginx/sites-available/$SITE_NAME"
SITES_ENABLED="/etc/nginx/sites-enabled/$SITE_NAME"

if ! command -v nginx &>/dev/null; then
    echo "[nginx] Installing nginx..."
    sudo apt-get update -qq
    sudo apt-get install -y nginx
fi

sudo systemctl enable nginx

echo "[nginx] Copying site config..."
sudo cp "$CONF_SRC" "$SITES_AVAIL"

echo "[nginx] Enabling site..."
sudo ln -sf "$SITES_AVAIL" "$SITES_ENABLED"
sudo rm -f /etc/nginx/sites-enabled/default

echo "[nginx] Validating config..."
sudo nginx -t

echo "[nginx] Reloading nginx..."
sudo systemctl reload nginx

echo "[nginx] Done. Site $SITE_NAME is active."
