# Infiniour Staging Deployment

**Architecture**

```
Internet
   ↓ :80 / :443
Host Nginx (Ubuntu service)
   ↓ 127.0.0.1:8080
Docker: infiniour-api  (ghcr.io/saurabh9678/dms-api-server:staging)
   ↓ postgres:5432 (Docker network only)
Docker: infiniour-postgres  (postgres:17-alpine)
```

Images are built in GitHub Actions and pushed to GHCR. The VM only pulls and runs — no build toolchain required.

---

## Folder Structure

The repository lives on the VM at `/opt/dms-api-server`. All deployment files are under `deploy/staging/`:

```
/opt/dms-api-server/
└── deploy/staging/
    ├── README.md
    ├── docker-compose.yml
    ├── .env                 # created on server (not committed)
    ├── .env.example
    ├── nginx.conf
    ├── backups/             # created on server
    └── scripts/
        ├── backup.sh
        ├── deploy.sh
        ├── migrate.sh
        ├── restore.sh
        ├── setup-nginx.sh
        └── setup-ssl.sh
```

---

## Prerequisites

- Ubuntu 24.04 Azure VM
- Docker Engine installed
- Nginx installed
- DNS: `stag-api.infiniour.com` → VM public IP

### Install Docker

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER
newgrp docker
```

### Install Nginx

```bash
sudo apt update && sudo apt install -y nginx
sudo systemctl enable nginx
```

---

## Step 1 — Authenticate with GHCR

GHCR requires a GitHub Personal Access Token (PAT) to pull images. This is a one-time setup per VM.

**Create a PAT:** GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) → New token  
Required scope: `read:packages` only.

```bash
# Store token securely
echo "YOUR_PAT_HERE" > ~/.ghcr_token
chmod 600 ~/.ghcr_token

# Authenticate Docker
cat ~/.ghcr_token | docker login ghcr.io -u saurabh9678 --password-stdin
# Expected: "Login Succeeded"

# Verify
cat ~/.docker/config.json | grep ghcr.io

# Test pull
docker pull ghcr.io/saurabh9678/dms-api-server:staging
```

**Troubleshooting:**

| Error | Fix |
|---|---|
| `401 Unauthorized` | Regenerate PAT, re-run `docker login` |
| `403 Forbidden` | PAT is missing `read:packages` scope |
| `denied: permission_denied` | Make the GHCR package public, or add your user as a collaborator |
| Network error on `https://ghcr.io` | Check VM firewall/NSG allows outbound 443 |

---

## Step 2 — Clone Repository on VM

```bash
sudo mkdir -p /opt
sudo git clone https://github.com/saurabh9678/dms-api-server.git /opt/dms-api-server
cd /opt/dms-api-server
git checkout staging

sudo mkdir -p deploy/staging/backups
sudo chmod +x deploy/staging/scripts/*.sh
```

If the repo is already present, pull the latest changes instead:

```bash
cd /opt/dms-api-server
git pull
```

---

## Step 3 — Configure Environment

Create `.env` next to `docker-compose.yml`:

```bash
cd /opt/dms-api-server/deploy/staging
sudo cp .env.example .env
sudo nano .env   # fill in real passwords and secrets
sudo chmod 600 .env
```

Generate a strong secret for `AUTH_ACCESS_TOKEN_SECRET`:

```bash
openssl rand -hex 32
```

---

## Bootstrap Scripts

Two scripts automate the manual Nginx and SSL steps below.

**Step 4 automated:**
```bash
# Installs nginx if missing, copies nginx.conf, enables site, validates, reloads
/opt/dms-api-server/deploy/staging/scripts/setup-nginx.sh
```

**Step 6 automated:**
```bash
# Installs certbot, obtains Let's Encrypt cert, tests renewal
/opt/dms-api-server/deploy/staging/scripts/setup-ssl.sh
```

Run `setup-nginx.sh` first (Step 4), then `setup-ssl.sh` after DNS is live (Step 6). The manual steps below remain for reference.

---

## Step 4 — Configure Host Nginx

### Main config

Replace `/etc/nginx/nginx.conf`:

```nginx
user www-data;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log  /var/log/nginx/error.log warn;

    sendfile       on;
    tcp_nopush     on;
    tcp_nodelay    on;
    keepalive_timeout 65;
    server_tokens off;

    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 1000;
    gzip_types
        text/plain
        text/css
        text/xml
        application/json
        application/javascript
        application/xml+rss;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
```

### Site config

Create `/etc/nginx/sites-available/stag-api.infiniour.com`:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name stag-api.infiniour.com;

    access_log /var/log/nginx/stag-api-access.log main;
    error_log  /var/log/nginx/stag-api-error.log warn;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Certbot will insert ACME challenge location and SSL redirect here automatically.
    # Run: sudo certbot --nginx -d stag-api.infiniour.com

    location / {
        proxy_pass          http://127.0.0.1:8080;
        proxy_http_version  1.1;

        proxy_set_header    Host              $host;
        proxy_set_header    X-Real-IP         $remote_addr;
        proxy_set_header    X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header    X-Forwarded-Proto $scheme;

        proxy_connect_timeout 10s;
        proxy_read_timeout    30s;
        proxy_send_timeout    30s;

        proxy_buffering     on;
        proxy_buffer_size   4k;
        proxy_buffers       8 4k;
    }
}

# HTTPS server block — Certbot populates this automatically after SSL setup.
# Reference only; do not uncomment manually.
#
# server {
#     listen 443 ssl;
#     listen [::]:443 ssl;
#     server_name stag-api.infiniour.com;
#
#     ssl_certificate     /etc/letsencrypt/live/stag-api.infiniour.com/fullchain.pem;
#     ssl_certificate_key /etc/letsencrypt/live/stag-api.infiniour.com/privkey.pem;
#     include             /etc/letsencrypt/options-ssl-nginx.conf;
#     ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;
#
#     add_header Strict-Transport-Security "max-age=63072000; includeSubDomains" always;
#     add_header X-Frame-Options "SAMEORIGIN" always;
#     add_header X-Content-Type-Options "nosniff" always;
#     add_header Referrer-Policy "strict-origin-when-cross-origin" always;
#
#     location / {
#         proxy_pass          http://127.0.0.1:8080;
#         proxy_http_version  1.1;
#         proxy_set_header    Host              $host;
#         proxy_set_header    X-Real-IP         $remote_addr;
#         proxy_set_header    X-Forwarded-For   $proxy_add_x_forwarded_for;
#         proxy_set_header    X-Forwarded-Proto $scheme;
#     }
# }
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/stag-api.infiniour.com \
           /etc/nginx/sites-enabled/stag-api.infiniour.com
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

---

## Step 5 — Initial Deployment

Start Postgres, apply database migrations, then bring up the full stack:

```bash
cd /opt/dms-api-server/deploy/staging
docker compose up -d postgres
./scripts/migrate.sh
docker compose up -d

# Verify
docker compose ps
curl http://127.0.0.1:8080/health
curl http://stag-api.infiniour.com/health
```

`migrate.sh` runs pending SQL migrations from `migrations/` using the `migrate/migrate` Docker image on the Compose network. Re-run it after schema changes are merged — it only applies new migrations.

Check current migration version:

```bash
./scripts/migrate.sh version
```

---

## Step 6 — Configure SSL (After DNS is Live)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d stag-api.infiniour.com

# Test renewal
sudo certbot renew --dry-run

# Check cert status
sudo certbot certificates

# Renewal runs automatically via systemd timer
sudo systemctl status certbot.timer
```

---

## Operational Runbook

### Deploy new API image

Every push to `staging` in GitHub Actions builds and pushes a new `staging` tag to GHCR. To deploy:

```bash
cd /opt/dms-api-server/deploy/staging
./scripts/deploy.sh
```

This runs `docker compose pull`, `docker compose up -d`, and `docker image prune -f`.

### Apply database migrations

After pulling code that adds new migrations, or on a fresh VM after Postgres is running:

```bash
cd /opt/dms-api-server/deploy/staging
./scripts/migrate.sh
```

### Update API image only (no Postgres restart)

```bash
cd /opt/dms-api-server/deploy/staging
docker compose pull api
docker compose up -d --no-deps api
```

### Restart a single service

```bash
cd /opt/dms-api-server/deploy/staging
docker compose restart api
docker compose restart postgres
```

### View logs

```bash
cd /opt/dms-api-server/deploy/staging

# API logs
docker compose logs -f api

# Postgres logs
docker compose logs -f postgres

# Nginx access log
tail -f /var/log/nginx/stag-api-access.log

# Nginx error log
tail -f /var/log/nginx/stag-api-error.log

# Nginx systemd journal
journalctl -u nginx -f
```

### Manage Nginx

```bash
sudo nginx -t                    # validate config
sudo systemctl reload nginx      # reload without dropping connections
sudo systemctl restart nginx     # full restart
sudo systemctl status nginx      # check status
```

### Database backup

```bash
cd /opt/dms-api-server/deploy/staging
./scripts/backup.sh
# Backups saved to /opt/dms-api-server/deploy/staging/backups/
# Files older than 14 days are pruned automatically
```

Add to cron for nightly automated backups:

```bash
sudo crontab -e
# Add:
0 2 * * * /opt/dms-api-server/deploy/staging/scripts/backup.sh >> /var/log/infiniour-backup.log 2>&1
```

### Database restore

```bash
cd /opt/dms-api-server/deploy/staging
./scripts/restore.sh /opt/dms-api-server/deploy/staging/backups/dms_20260613_020000.sql.gz
```

### Health check

```bash
curl http://stag-api.infiniour.com/health   # via Nginx
curl http://127.0.0.1:8080/health           # direct to API (from VM only)
cd /opt/dms-api-server/deploy/staging && docker compose ps   # service status
docker exec infiniour-postgres pg_isready -U infiniour -d dms
```

---

## Security: Exposed Ports

| Port | Accessible From | Notes |
|---|---|---|
| 22 | Internet (restrict to your IP in Azure NSG) | SSH |
| 80 | Internet | HTTP, ACME challenge |
| 443 | Internet | HTTPS (after SSL setup) |
| 8080 | VM loopback only (`127.0.0.1`) | API — not publicly reachable |
| 5432 | Docker network only | Postgres — not reachable from host or internet |

**Azure NSG inbound rules:**
- TCP 22 from your office/VPN IP
- TCP 80 from Any
- TCP 443 from Any
- Deny all other inbound

**Verify Postgres and API are not exposed:**

```bash
# From an external machine:
nmap -p 5432,8080 <VM_PUBLIC_IP>
# Both should show: filtered or closed
```

---

## Disaster Recovery After VM Recreation

1. Provision a new Ubuntu 24.04 VM on Azure
2. Install Docker: `curl -fsSL https://get.docker.com | sudo sh`
3. Install Nginx: `sudo apt install -y nginx`
4. Clone or copy `/opt/dms-api-server/` to the new VM (include `.env` and `deploy/staging/backups/`)
5. Authenticate with GHCR: `cat ~/.ghcr_token | docker login ghcr.io -u saurabh9678 --password-stdin`
6. Copy and enable Nginx site config:
   ```bash
   sudo ln -s /etc/nginx/sites-available/stag-api.infiniour.com /etc/nginx/sites-enabled/
   sudo systemctl enable --now nginx
   ```
7. Start services: `cd /opt/dms-api-server/deploy/staging && docker compose up -d postgres && ./scripts/migrate.sh && docker compose up -d`
8. Restore database: `cd /opt/dms-api-server/deploy/staging && ./scripts/restore.sh backups/<latest>.sql.gz`
9. Re-issue SSL cert: `sudo certbot --nginx -d stag-api.infiniour.com`
10. Verify: `curl https://stag-api.infiniour.com/health`
