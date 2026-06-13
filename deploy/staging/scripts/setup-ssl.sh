#!/bin/bash
set -euo pipefail

DOMAIN="stag-api.infiniour.com"

echo "[ssl] Installing certbot..."
sudo apt-get update -qq
sudo apt-get install -y certbot python3-certbot-nginx

echo "[ssl] Requesting certificate for $DOMAIN..."
sudo certbot --nginx -d "$DOMAIN"

echo "[ssl] Testing auto-renewal..."
sudo certbot renew --dry-run

echo "[ssl] Certificate issued. Auto-renewal timer status:"
sudo systemctl status certbot.timer --no-pager || true

echo "[ssl] Done. HTTPS is now active for $DOMAIN."
