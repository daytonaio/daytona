# Daytona Self-Hosted Deployment Guide

## Cloudflare Tunnel + Nginx Proxy Manager + Docker Compose

This guide walks through deploying Daytona on your own server with HTTPS using Cloudflare Tunnels for ingress and Nginx Proxy Manager (NPM) for internal routing. By the end, you will have:

- Daytona dashboard and API accessible at `https://your-domain.com`
- OIDC authentication via Dex at `https://dex.your-domain.com`
- Sandbox proxy (terminals, port previews) at `https://*.proxy.your-domain.com`
- SSH gateway for direct sandbox access

> **Note**: This guide assumes Cloudflare manages your DNS. If you use a different DNS/reverse proxy setup, the concepts are the same — you need three HTTP routes and one TCP route.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Prerequisites](#prerequisites)
3. [DNS & Domain Planning](#dns--domain-planning)
4. [Step 1: Configure the `.env` File](#step-1-configure-the-env-file)
5. [Step 2: Configure Dex (OIDC Provider)](#step-2-configure-dex-oidc-provider)
6. [Step 3: Set Up Nginx Proxy Manager](#step-3-set-up-nginx-proxy-manager)
7. [Step 4: Set Up Cloudflare Tunnels](#step-4-set-up-cloudflare-tunnels)
8. [Step 5: Start Daytona](#step-5-start-daytona)
9. [Step 6: Verify the Deployment](#step-6-verify-the-deployment)
10. [SSH Gateway Setup](#ssh-gateway-setup)
11. [Troubleshooting](#troubleshooting)
12. [Security Hardening](#security-hardening)
13. [Reference: Environment Variables](#reference-environment-variables)

---

## Architecture Overview

```
                         Internet
                            │
              ┌─────────────┼─────────────┐
              │       Cloudflare CDN       │
              │    (SSL termination)       │
              └─────────────┬─────────────┘
                            │
              ┌─────────────┴─────────────┐
              │    Cloudflare Tunnel(s)    │
              │       (cloudflared)        │
              └──┬──────────┬─────────┬───┘
                 │          │         │
    ┌────────────┤          │         ├────────────┐
    │ your-domain.com       │    *.proxy.your-domain.com
    │            │  dex.your-domain.com        │
    │            │          │                  │
    └────────┐   └──────┐   └──────┐   ┌──────┘
             │          │          │   │
       ┌─────▼──────────▼──────────▼───▼─────┐
       │        Nginx Proxy Manager          │
       │          (internal routing)          │
       └───┬────────────┬──────────────┬─────┘
           │            │              │
     ┌─────▼─────┐ ┌────▼────┐ ┌──────▼──────┐
     │  API:3000  │ │Dex:5556 │ │ Proxy:4000  │
     │ (dashboard │ │ (OIDC)  │ │ (sandboxes) │
     │  + API)    │ │         │ │             │
     └───────────┘ └─────────┘ └─────────────┘
```

**How it works:**

1. Cloudflare handles DNS and SSL termination for all domains
2. `cloudflared` tunnels forward traffic to Nginx Proxy Manager on your server
3. NPM routes requests to the correct Docker service based on hostname
4. The API serves both the dashboard UI and the REST API on port 3000
5. Dex handles OIDC authentication on port 5556
6. The proxy handles sandbox port previews (terminals, web apps) on port 4000

**SSH gateway** is a special case — it uses raw TCP, so it goes direct (not through Cloudflare/NPM).

---

## Prerequisites

- A Linux server (VPS, bare metal, or home server) with Docker and Docker Compose installed
- A domain name with DNS managed by Cloudflare (free tier is fine)
- Cloudflare account with Tunnels enabled (free with any plan)
- Nginx Proxy Manager running on the same server (or install it now — see Step 3)
- Basic familiarity with terminal/SSH

### Minimum Server Specs

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Disk | 20 GB | 50+ GB (sandbox images) |
| OS | Any Linux with Docker | Ubuntu 22.04+ / Debian 12+ |

---

## DNS & Domain Planning

You need **three hostnames** and optionally a **TCP endpoint** for SSH:

| Purpose | Hostname | Example | Cloudflare Proxy |
|---------|----------|---------|-----------------|
| Dashboard + API | `your-domain.com` | `daytona.example.com` | Yes (orange cloud) |
| OIDC (Dex) | `dex.your-domain.com` | `dex.example.com` | Yes (orange cloud) |
| Sandbox Proxy | `proxy.your-domain.com` | `proxy.daytona.example.com` | Yes (orange cloud) |
| Sandbox Proxy wildcard | `*.proxy.your-domain.com` | `*.proxy.daytona.example.com` | Yes (orange cloud) |
| SSH Gateway (optional) | Any hostname + port | `ssh.example.com:65022` | **No** (TCP, not HTTP) |

**Important**: The proxy requires **wildcard DNS** because each sandbox gets its own subdomain (e.g., `22222-abc123.proxy.daytona.example.com`). Cloudflare Tunnels support wildcard routes on paid plans, or you can use a wildcard DNS record pointing to your server's IP.

---

## Step 1: Configure the `.env` File

The `docker/.env` file contains all configurable values. Copy the template and edit it:

```bash
cd docker
cp .env .env.backup  # optional: keep the original
```

Open `.env` in your editor and update these values. Replace `example.com` with your actual domain throughout:

### Required Domain Settings

```bash
# ── Main Domain ──
DAYTONA_DOMAIN=daytona.example.com
DAYTONA_URL=https://daytona.example.com

# ── Dashboard ──
DASHBOARD_URL=https://daytona.example.com/dashboard
DASHBOARD_BASE_API_URL=https://daytona.example.com

# ── OIDC / Dex ──
DEX_PUBLIC_URL=https://dex.example.com/dex
# Keep this as-is — it's the Docker-internal URL:
DEX_INTERNAL_URL=http://dex:5556/dex

# ── Proxy ──
PROXY_DOMAIN=proxy.daytona.example.com
PROXY_PROTOCOL=https
PROXY_TEMPLATE_URL=https://{{PORT}}-{{sandboxId}}.proxy.daytona.example.com
PROXY_TOOLBOX_BASE_URL=https://proxy.daytona.example.com
```

### Required Security Secrets

Generate strong random values for these. **Do not use the defaults in production.**

```bash
# Generate random secrets (run these commands and paste the output):
openssl rand -hex 32   # → use for ENCRYPTION_KEY
openssl rand -hex 16   # → use for ENCRYPTION_SALT
openssl rand -hex 32   # → use for PROXY_API_KEY
openssl rand -hex 32   # → use for DEFAULT_RUNNER_API_KEY
openssl rand -hex 32   # → use for SSH_GATEWAY_API_KEY
openssl rand -hex 32   # → use for OTEL_COLLECTOR_API_KEY
openssl rand -hex 32   # → use for HEALTH_CHECK_API_KEY
openssl rand -base64 24  # → use for DB_PASSWORD
openssl rand -base64 24  # → use for S3_ACCESS_KEY
openssl rand -base64 24  # → use for S3_SECRET_KEY
openssl rand -base64 24  # → use for REGISTRY_PASSWORD
```

Update the values in `.env`:

```bash
ENCRYPTION_KEY=<your-generated-value>
ENCRYPTION_SALT=<your-generated-value>
PROXY_API_KEY=<your-generated-value>
DEFAULT_RUNNER_API_KEY=<your-generated-value>
SSH_GATEWAY_API_KEY=<your-generated-value>
DB_PASSWORD=<your-generated-value>
# ... etc.
```

### SSH Gateway Settings (optional)

If you want SSH access to sandboxes, configure the SSH gateway hostname and port. This must be a hostname/port that is **directly reachable** from the internet (no Cloudflare proxy — it's raw TCP):

```bash
SSH_GATEWAY_HOST=ssh.example.com       # or your server's public IP
SSH_GATEWAY_PORT=2222                   # or any available port (e.g., 65022)
SSH_GATEWAY_URL=ssh.example.com:2222
SSH_GATEWAY_COMMAND=ssh -p 2222 {{TOKEN}}@ssh.example.com
```

> **Tip**: If your server is behind NAT, you'll need to port-forward this port on your router. If using a VPS with a public IP, the port is directly accessible. You can also use a dynamic DNS service (like DynDNS or Dynu) if your IP changes.

### Full `.env` Reference

See [Reference: Environment Variables](#reference-environment-variables) at the bottom of this guide for a complete listing of every variable with descriptions.

---

## Step 2: Configure Dex (OIDC Provider)

Dex is the identity provider that handles login. Its config is at `docker/dex/config.yaml`.

Open `docker/dex/config.yaml` and update:

### 2a. Set the Issuer URL

Change the `issuer` to your public Dex URL:

```yaml
issuer: https://dex.example.com/dex
```

This **must** match the `DEX_PUBLIC_URL` in your `.env` file exactly.

### 2b. Update Redirect URIs

Under `staticClients`, update the `redirectURIs` to include your external domains:

```yaml
staticClients:
  - id: daytona
    redirectURIs:
      # External domain URLs (required for production)
      - https://daytona.example.com
      - https://daytona.example.com/api/oauth2-redirect.html
      - https://proxy.daytona.example.com/callback
      # Localhost URLs (keep for local development/debugging)
      - http://localhost:3000
      - http://localhost:3000/api/oauth2-redirect.html
      - http://proxy.localhost:4000/callback
    name: "Daytona"
    public: true
```

### 2c. Change the Default Admin Password

The default config has a static user `dev@daytona.io` with password `password`. **Change this for production.**

Generate a bcrypt hash for your new password:

```bash
# Using htpasswd (install: apt-get install apache2-utils)
htpasswd -nbBC 10 "" "YourSecurePassword123!" | tr -d ':\n' | sed 's/$2y/$2a/'

# Or using Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'YourSecurePassword123!', bcrypt.gensalt(10)).decode())"

# Or using Docker
docker run --rm -it alpine sh -c "apk add --no-cache apache2-utils && htpasswd -nbBC 10 '' 'YourSecurePassword123!' | tr -d ':\n' | sed 's/\$2y/\$2a/'"
```

Update `dex/config.yaml`:

```yaml
staticPasswords:
  - email: "admin@example.com"
    hash: "$2a$10$your-generated-bcrypt-hash-here"
    username: "admin"
    userID: "1234"
```

---

## Step 3: Set Up Nginx Proxy Manager

Nginx Proxy Manager routes incoming HTTP(S) traffic to the correct Docker service. If you already have NPM running, skip to 3b.

### 3a. Install Nginx Proxy Manager (if needed)

Create a `docker-compose.npm.yaml` file somewhere on your server (e.g., `/opt/npm/`):

```yaml
services:
  npm:
    image: jc21/nginx-proxy-manager:latest
    restart: unless-stopped
    ports:
      - "80:80"       # HTTP (used for Cloudflare tunnel health checks)
      - "443:443"     # HTTPS (not strictly needed with Cloudflare tunnels)
      - "81:81"       # NPM Admin UI
    volumes:
      - npm_data:/data
      - npm_letsencrypt:/etc/letsencrypt
    networks:
      - daytona_default   # Must be on the same network as Daytona services

networks:
  daytona_default:
    external: true      # Created by Daytona's docker-compose

volumes:
  npm_data:
  npm_letsencrypt:
```

**Important**: NPM must be on the **same Docker network** as the Daytona services. The Daytona compose file creates a default network. Start Daytona first (or create the network manually) so NPM can attach to it:

```bash
# Create the network first if Daytona isn't started yet
docker network create docker_default

# Start NPM
docker compose -f docker-compose.npm.yaml up -d
```

Access the NPM admin UI at `http://your-server-ip:81` (default login: `admin@example.com` / `changeme`).

> **Alternative**: If NPM and Daytona are on different Docker networks, you can use the host's IP address instead of container names in the proxy host settings (e.g., `172.17.0.1:3000` instead of `api:3000`).

### 3b. Create Proxy Hosts in NPM

You need **three proxy hosts** in Nginx Proxy Manager:

#### Proxy Host 1: Dashboard + API

| Field | Value |
|-------|-------|
| Domain Names | `daytona.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `api` (Docker service name) or server IP |
| Forward Port | `3000` |
| Websockets Support | **Yes** (required for real-time notifications) |
| Block Common Exploits | Yes |
| Cache Assets | No |

> The WebSocket support is needed because the dashboard uses Socket.IO at `/api/socket.io/` for real-time notifications.

Under the **Custom Nginx Configuration** tab (Advanced), add:

```nginx
# Increase body size for file uploads to sandboxes
client_max_body_size 500m;

# Increase timeouts for long-running operations
proxy_read_timeout 300s;
proxy_connect_timeout 75s;
proxy_send_timeout 300s;
```

#### Proxy Host 2: Dex (OIDC)

| Field | Value |
|-------|-------|
| Domain Names | `dex.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `dex` (Docker service name) or server IP |
| Forward Port | `5556` |
| Websockets Support | No |
| Block Common Exploits | Yes |

No special configuration needed.

#### Proxy Host 3: Sandbox Proxy (Wildcard)

| Field | Value |
|-------|-------|
| Domain Names | `proxy.daytona.example.com` and `*.proxy.daytona.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `proxy` (Docker service name) or server IP |
| Forward Port | `4000` |
| Websockets Support | **Yes** (required for terminal WebSocket connections) |
| Block Common Exploits | No (sandboxes may trigger false positives) |

Under the **Custom Nginx Configuration** tab (Advanced), add:

```nginx
# Increase body size for file uploads to sandboxes
client_max_body_size 500m;

# Increase timeouts for long-running sandbox connections
proxy_read_timeout 3600s;
proxy_connect_timeout 75s;
proxy_send_timeout 3600s;

# Required for terminal WebSocket connections
proxy_buffering off;
```

> **Why WebSocket support for proxy?** The terminal runs inside the sandbox on port 22222 and is accessed via `https://22222-{sandboxId}.proxy.daytona.example.com`. The browser loads an iframe that opens a WebSocket connection for the interactive terminal.

### 3c. SSL Certificates in NPM

Since Cloudflare handles SSL termination, you have two options:

**Option A: No SSL in NPM (simpler)** — If Cloudflare tunnel connects to NPM over HTTP on the same machine, you don't need SSL certificates in NPM. Just set the scheme to `http` for all proxy hosts.

**Option B: SSL in NPM** — If you want end-to-end encryption, you can add Let's Encrypt certificates in NPM. However, this is usually unnecessary when Cloudflare tunnel runs on the same machine.

---

## Step 4: Set Up Cloudflare Tunnels

Cloudflare Tunnels create secure outbound connections from your server to Cloudflare's edge, eliminating the need to open inbound ports or expose your server's IP.

### 4a. Install cloudflared

```bash
# Debian/Ubuntu
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb

# Or via Docker
docker pull cloudflare/cloudflared:latest
```

### 4b. Authenticate and Create a Tunnel

```bash
# Login to Cloudflare (opens browser)
cloudflared tunnel login

# Create a tunnel
cloudflared tunnel create daytona

# Note the tunnel ID and credentials file path from the output
# Example: Created tunnel daytona with id a1b2c3d4-e5f6-...
```

### 4c. Configure Tunnel Routes

Create (or edit) `~/.cloudflared/config.yml`:

```yaml
tunnel: a1b2c3d4-e5f6-...   # Your tunnel ID
credentials-file: /home/youruser/.cloudflared/a1b2c3d4-e5f6-....json

ingress:
  # Route 1: Dashboard + API
  - hostname: daytona.example.com
    service: http://localhost:80   # Points to NPM's HTTP port
    originRequest:
      noTLSVerify: true

  # Route 2: Dex (OIDC)
  - hostname: dex.example.com
    service: http://localhost:80   # Points to NPM's HTTP port
    originRequest:
      noTLSVerify: true

  # Route 3: Proxy (exact domain)
  - hostname: proxy.daytona.example.com
    service: http://localhost:80   # Points to NPM's HTTP port
    originRequest:
      noTLSVerify: true

  # Route 4: Proxy (wildcard for sandbox subdomains)
  - hostname: "*.proxy.daytona.example.com"
    service: http://localhost:80   # Points to NPM's HTTP port
    originRequest:
      noTLSVerify: true

  # Catch-all (required by cloudflared)
  - service: http_status:404
```

> **Wildcard routes**: Cloudflare Tunnels support wildcard hostnames. The `*.proxy.daytona.example.com` route catches all sandbox-specific subdomains like `22222-abc123.proxy.daytona.example.com`.

### 4d. Create DNS Records

For each hostname, create a CNAME record pointing to your tunnel:

```bash
cloudflared tunnel route dns daytona daytona.example.com
cloudflared tunnel route dns daytona dex.example.com
cloudflared tunnel route dns daytona proxy.daytona.example.com
```

For the **wildcard**, you need to create it manually in the Cloudflare dashboard:

1. Go to your domain in the Cloudflare dashboard
2. Navigate to **DNS** > **Records**
3. Add a new **CNAME** record:
   - Name: `*.proxy.daytona` (or `*.proxy` if using `proxy.example.com`)
   - Target: `a1b2c3d4-e5f6-....cfargotunnel.com` (your tunnel's CNAME target)
   - Proxy status: **Proxied** (orange cloud)

> **Note on wildcard DNS with Cloudflare free plan**: Wildcard DNS records (`*`) work on the free plan. Wildcard SSL certificates (which Cloudflare issues automatically for proxied records) also work on the free plan when using Cloudflare's proxy (orange cloud). The traffic flows through Cloudflare's edge which handles TLS.

### 4e. Start the Tunnel

```bash
# Test first
cloudflared tunnel run daytona

# Once working, install as a system service
sudo cloudflared service install
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
```

**Alternative: Cloudflare Dashboard Tunnel Setup**

If you prefer a GUI, you can create tunnels directly in the Cloudflare Zero Trust dashboard:

1. Go to [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) > **Networks** > **Tunnels**
2. Click **Create a tunnel** > **Cloudflared**
3. Name it `daytona` and install the connector on your server
4. Add public hostname entries for each of the four routes above
5. For each entry, set the service to `http://localhost:80` (NPM's port)

---

## Step 5: Start Daytona

With everything configured, start the services:

```bash
cd docker

# Start all services
docker compose up -d

# Watch the logs for errors
docker compose logs -f api proxy dex
```

The first startup will:
- Initialize the PostgreSQL database
- Create MinIO buckets
- Build/pull the default sandbox snapshot (this takes a few minutes)

Wait for the API to be healthy:

```bash
# Check service health
docker compose ps

# The API should show "healthy" after ~30 seconds
docker compose logs api | grep "Nest application successfully started"
```

---

## Step 6: Verify the Deployment

Run through these checks in order:

### 6a. API Config Endpoint

```bash
curl -s https://daytona.example.com/api/config | jq .
```

Verify the response contains your external URLs:

```json
{
  "proxyTemplateUrl": "https://{{PORT}}-{{sandboxId}}.proxy.daytona.example.com",
  "proxyToolboxUrl": "https://proxy.daytona.example.com/toolbox",
  "dashboardUrl": "https://daytona.example.com/dashboard",
  "oidc": {
    "issuer": "https://dex.example.com/dex",
    "clientId": "daytona",
    "audience": "daytona"
  },
  "sshGatewayCommand": "ssh -p 2222 {{TOKEN}}@ssh.example.com"
}
```

If you see `localhost` in any of these values, your `.env` is not configured correctly.

### 6b. Dex Discovery

```bash
curl -s https://dex.example.com/dex/.well-known/openid-configuration | jq .issuer
```

Should return: `"https://dex.example.com/dex"`

### 6c. Dashboard Login

1. Open `https://daytona.example.com` in your browser
2. You should be redirected to `https://dex.example.com/dex/auth/...`
3. Log in with your Dex credentials (default: `dev@daytona.io` / `password`)
4. You should be redirected back to the dashboard

### 6d. Create a Sandbox

1. In the dashboard, make sure the default snapshot is active (check **Snapshots** page)
2. Create a new sandbox
3. Once running, open the terminal tab — it should load in an iframe
4. Check the terminal iframe URL — it should be `https://22222-*.proxy.daytona.example.com/...`

### 6e. SSH Access (if configured)

```bash
# Get the SSH command from a sandbox's details page, or construct it:
ssh -p 2222 <sandbox-token>@ssh.example.com
```

---

## SSH Gateway Setup

The SSH gateway allows direct SSH access to sandboxes. It uses raw TCP, so it **cannot go through Cloudflare Tunnels** (which only support HTTP/HTTPS/WebSocket).

### Options for Exposing SSH

**Option A: Direct Port (simplest)**

If your server has a public IP, just map the port in `docker-compose.yaml` (already done by default):

```yaml
ssh-gateway:
  ports:
    - "2222:2222"
```

Set in `.env`:
```bash
SSH_GATEWAY_HOST=your-server-public-ip
SSH_GATEWAY_PORT=2222
```

**Option B: Non-Standard Port (behind NAT/firewall)**

If port 2222 is blocked, use a high port:

1. Change the port mapping in `docker-compose.yaml`:
   ```yaml
   ssh-gateway:
     ports:
       - "65022:2222"   # External port 65022 → internal 2222
   ```

2. Update `.env`:
   ```bash
   SSH_GATEWAY_PORT=65022
   SSH_GATEWAY_URL=your-server.example.com:65022
   SSH_GATEWAY_COMMAND=ssh -p 65022 {{TOKEN}}@your-server.example.com
   ```

3. Port-forward 65022 on your router/firewall to the server.

**Option C: Dynamic DNS**

If your public IP changes, use a dynamic DNS service (e.g., Dynu, No-IP, DuckDNS):

```bash
SSH_GATEWAY_HOST=myserver.dynu.net
SSH_GATEWAY_PORT=65022
SSH_GATEWAY_URL=myserver.dynu.net:65022
SSH_GATEWAY_COMMAND=ssh -p 65022 {{TOKEN}}@myserver.dynu.net
```

**Option D: Disable SSH**

If you don't need SSH access, leave the SSH variables empty in `.env`:

```bash
SSH_GATEWAY_HOST=
SSH_GATEWAY_PORT=
SSH_GATEWAY_URL=
SSH_GATEWAY_COMMAND=
```

---

## Troubleshooting

### "OIDC issuer mismatch" / JWT validation fails

**Symptom**: Login redirects work, but after authentication you get an error.

**Cause**: The `issuer` in `dex/config.yaml` doesn't match `DEX_PUBLIC_URL` in `.env`.

**Fix**: Ensure both are identical (including trailing path):
```
dex/config.yaml:  issuer: https://dex.example.com/dex
.env:             DEX_PUBLIC_URL=https://dex.example.com/dex
```

Also ensure `DEX_INTERNAL_URL` remains as the Docker-internal URL (`http://dex:5556/dex`). The API uses `DEX_INTERNAL_URL` to fetch JWKS keys internally, while presenting `DEX_PUBLIC_URL` to browsers.

### Terminal/sandbox iframe shows "connection refused" or blank

**Symptom**: Dashboard loads, sandboxes start, but the terminal tab is empty.

**Possible causes**:

1. **Wildcard DNS not configured**: Ensure `*.proxy.daytona.example.com` resolves to your server/tunnel.
   ```bash
   dig +short 22222-test.proxy.daytona.example.com
   ```

2. **NPM wildcard proxy host missing**: The proxy host in NPM must include `*.proxy.daytona.example.com` in its domain names.

3. **WebSocket not enabled in NPM**: The proxy host must have "Websockets Support" turned on.

4. **Cloudflare WebSocket setting**: In your Cloudflare domain settings, go to **Network** and ensure **WebSockets** is enabled (it is by default).

### Dashboard shows localhost URLs / API requests go to localhost

**Symptom**: After login, the dashboard tries to reach `http://localhost:3000/api`.

**Cause**: `DASHBOARD_BASE_API_URL` in `.env` is wrong or empty.

**Fix**: Set it to your external URL (without trailing slash):
```bash
DASHBOARD_BASE_API_URL=https://daytona.example.com
```

Then restart the API service:
```bash
docker compose restart api
```

### Dex login page not loading

**Symptom**: Clicking "Login" redirects to `https://dex.example.com/dex/auth/...` but the page doesn't load.

**Check**:
1. Cloudflare tunnel route for `dex.example.com` exists
2. NPM proxy host for `dex.example.com` → `dex:5556` exists
3. Dex container is running: `docker compose logs dex`

### "Invalid redirect URI" error from Dex

**Cause**: The redirect URI in the browser doesn't match any URI in `dex/config.yaml`.

**Fix**: Add the missing redirect URI to `staticClients[0].redirectURIs` in `dex/config.yaml`:
```yaml
redirectURIs:
  - https://daytona.example.com
  - https://daytona.example.com/api/oauth2-redirect.html
  - https://proxy.daytona.example.com/callback
```

### Sandbox proxy returns 502 Bad Gateway

**Symptom**: Terminal URL loads but shows a 502 error.

**Possible causes**:

1. **Sandbox is not running**: Check sandbox status in the dashboard
2. **Proxy can't reach the runner**: Check proxy logs:
   ```bash
   docker compose logs proxy
   ```
3. **Proxy API key mismatch**: Ensure `PROXY_API_KEY` in `.env` is the same value used by both the API and proxy services

### Cloudflare 1033 error (Argo Tunnel not found)

**Cause**: DNS CNAME record points to a tunnel that doesn't exist or isn't running.

**Fix**: Verify your tunnel is running:
```bash
cloudflared tunnel list
cloudflared tunnel info daytona
```

### Cloudflare 522/524 timeout errors

**Cause**: Cloudflare can reach your tunnel, but NPM or the backend service isn't responding.

**Check**:
1. NPM is running and listening on port 80
2. The target Docker service is healthy
3. NPM proxy host exists for the requested hostname
4. Docker network connectivity: NPM can reach the service container

---

## Security Hardening

### Change Default Credentials

Before going to production, change **all** of these:

- [ ] Dex admin password in `dex/config.yaml` (see Step 2c)
- [ ] Database password (`DB_PASSWORD` in `.env`)
- [ ] MinIO credentials (`S3_ACCESS_KEY`, `S3_SECRET_KEY`)
- [ ] Registry credentials (`REGISTRY_ADMIN`, `REGISTRY_PASSWORD`)
- [ ] All API keys (`PROXY_API_KEY`, `DEFAULT_RUNNER_API_KEY`, `SSH_GATEWAY_API_KEY`, etc.)
- [ ] Encryption key and salt (`ENCRYPTION_KEY`, `ENCRYPTION_SALT`)

### Restrict Service Ports

The default `docker-compose.yaml` exposes some ports for development convenience. In production, you should only expose what's needed:

- **Keep exposed**: SSH gateway port (e.g., 2222) if using SSH access
- **Remove/restrict**: PgAdmin (5050), MinIO console (9001), Registry UI (5100), MailDev (1080), Jaeger (16686)

You can comment out or remove port mappings for development tools in `docker-compose.yaml`, or restrict them to localhost:

```yaml
# Instead of:
ports:
  - "5050:80"
# Use:
ports:
  - "127.0.0.1:5050:80"
```

### Firewall Rules

On your server, only allow inbound traffic on:
- Port 80 (HTTP — for Cloudflare tunnel health checks, if needed)
- Port 81 (NPM admin — restrict to your IP only)
- SSH gateway port (e.g., 2222 or 65022)
- Your server's SSH port (e.g., 22)

Block everything else with `ufw` or `iptables`:

```bash
sudo ufw default deny incoming
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 2222/tcp       # SSH gateway
sudo ufw allow from YOUR_IP to any port 81  # NPM admin
sudo ufw enable
```

### SMTP Configuration

The default setup uses MailDev (a testing email server). For production, configure a real SMTP provider:

```bash
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your-smtp-api-key
SMTP_FROM=no-reply@example.com
SMTP_SECURE=true
```

---

## Reference: Environment Variables

### Main Domain

| Variable | Description | Example |
|----------|-------------|---------|
| `DAYTONA_DOMAIN` | Primary domain (used as label) | `daytona.example.com` |
| `DAYTONA_URL` | Full URL with protocol | `https://daytona.example.com` |

### Dashboard

| Variable | Description | Example |
|----------|-------------|---------|
| `DASHBOARD_URL` | Public dashboard URL | `https://daytona.example.com/dashboard` |
| `DASHBOARD_BASE_API_URL` | Base URL the dashboard uses for API calls | `https://daytona.example.com` |

### OIDC / Dex

| Variable | Description | Example |
|----------|-------------|---------|
| `DEX_PUBLIC_URL` | Public-facing Dex issuer URL (browser-reachable) | `https://dex.example.com/dex` |
| `DEX_INTERNAL_URL` | Docker-internal Dex URL (for backend JWKS validation) | `http://dex:5556/dex` |
| `OIDC_CLIENT_ID` | OIDC client ID (must match Dex config) | `daytona` |
| `OIDC_AUDIENCE` | OIDC audience string | `daytona` |

### Proxy

| Variable | Description | Example |
|----------|-------------|---------|
| `PROXY_DOMAIN` | Proxy base domain (sandbox subdomains go under this) | `proxy.daytona.example.com` |
| `PROXY_PROTOCOL` | `http` or `https` | `https` |
| `PROXY_TEMPLATE_URL` | URL template with `{{PORT}}` and `{{sandboxId}}` placeholders | `https://{{PORT}}-{{sandboxId}}.proxy.daytona.example.com` |
| `PROXY_TOOLBOX_BASE_URL` | Base URL for the toolbox API | `https://proxy.daytona.example.com` |
| `PROXY_API_KEY` | Shared secret between API and proxy services | (random hex string) |

### SSH Gateway

| Variable | Description | Example |
|----------|-------------|---------|
| `SSH_GATEWAY_HOST` | Public hostname for SSH access | `ssh.example.com` |
| `SSH_GATEWAY_PORT` | Public port for SSH access | `2222` |
| `SSH_GATEWAY_URL` | Combined host:port | `ssh.example.com:2222` |
| `SSH_GATEWAY_COMMAND` | SSH command template (shown to users) | `ssh -p 2222 {{TOKEN}}@ssh.example.com` |
| `SSH_GATEWAY_API_KEY` | Auth key for SSH gateway service | (random hex string) |

### Database (PostgreSQL)

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database hostname | `db` (Docker service) |
| `DB_PORT` | Database port | `5432` |
| `DB_USERNAME` | Database user | `daytona` |
| `DB_PASSWORD` | Database password | **Change this!** |
| `DB_DATABASE` | Database name | `daytona` |

### Redis

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis hostname | `redis` (Docker service) |
| `REDIS_PORT` | Redis port | `6379` |

### Encryption

| Variable | Description | Default |
|----------|-------------|---------|
| `ENCRYPTION_KEY` | Key for encrypting sensitive data at rest | **Change this!** |
| `ENCRYPTION_SALT` | Salt for encryption key derivation | **Change this!** |

### S3 / MinIO

| Variable | Description | Default |
|----------|-------------|---------|
| `S3_ENDPOINT` | S3-compatible endpoint | `http://minio:9000` |
| `S3_STS_ENDPOINT` | S3 STS endpoint | `http://minio:9000` |
| `S3_REGION` | S3 region | `us-east-1` |
| `S3_ACCESS_KEY` | S3 access key | `minioadmin` (**Change!**) |
| `S3_SECRET_KEY` | S3 secret key | `minioadmin` (**Change!**) |
| `S3_DEFAULT_BUCKET` | Default bucket name | `daytona` |

### Registry

| Variable | Description | Default |
|----------|-------------|---------|
| `REGISTRY_URL` | Docker registry URL | `http://registry:6000` |
| `REGISTRY_ADMIN` | Registry admin user | `admin` |
| `REGISTRY_PASSWORD` | Registry admin password | **Change this!** |
| `REGISTRY_PROJECT_ID` | Registry project/namespace | `daytona` |

### SMTP

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server | `maildev` (dev only) |
| `SMTP_PORT` | SMTP port | `1025` (dev only) |
| `SMTP_USER` | SMTP username | (empty) |
| `SMTP_PASSWORD` | SMTP password | (empty) |
| `SMTP_FROM` | From email address | `no-reply@daytona.io` |
| `SMTP_SECURE` | Use TLS | (empty = no) |

### Runner

| Variable | Description | Default |
|----------|-------------|---------|
| `DEFAULT_RUNNER_API_KEY` | API key for the default runner | **Change this!** |

### Telemetry

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_ENABLED` | Enable OpenTelemetry | `true` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP exporter endpoint | `http://otel-collector:4318` |
| `OTEL_COLLECTOR_API_KEY` | OTel collector auth key | **Change this!** |

### Health Check

| Variable | Description | Default |
|----------|-------------|---------|
| `HEALTH_CHECK_API_KEY` | Health check endpoint auth | **Change this!** |
