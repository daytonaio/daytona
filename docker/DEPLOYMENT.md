# Daytona Self-Hosted Deployment Guide

Deploy Daytona on your own server with HTTPS using Cloudflare Tunnel and Nginx Proxy Manager.

This guide assumes:
- A Linux server (e.g., Ubuntu/Debian VPS, home server, Proxmox VM)
- Docker and Docker Compose installed
- A domain managed by Cloudflare (free plan is fine)
- You want HTTPS without opening ports 80/443 on your firewall

---

## Architecture Overview

```
Browser
  |
  | HTTPS
  v
Cloudflare (SSL termination, CDN, DDoS protection)
  |
  | Cloudflare Tunnel (encrypted, no open ports needed)
  v
cloudflared container (on your server)
  |
  | HTTP (internal)
  v
Nginx Proxy Manager (routing by hostname)
  |
  |-- daytona.example.com          --> api:3000       (dashboard + API)
  |-- dex.example.com              --> dex:5556       (OIDC auth)
  |-- proxy.example.com            --> proxy:4000     (sandbox proxy)
  |-- *.proxy.example.com          --> proxy:4000     (sandbox port preview)
  v
Docker containers (daytona-network)
```

SSH access bypasses Cloudflare entirely (raw TCP):

```
SSH client --> your-server:65022 --> ssh-gateway:2222
```

---

## Step 1: Plan Your Domains

You need **four** DNS entries. Choose your domain structure:

| Purpose | Example | Routes to |
|---|---|---|
| Dashboard + API | `daytona.example.com` | `api:3000` |
| Dex (OIDC auth) | `dex.example.com` | `dex:5556` |
| Proxy (base) | `proxy.daytona.example.com` | `proxy:4000` |
| Proxy (wildcard) | `*.proxy.daytona.example.com` | `proxy:4000` |

The wildcard entry is critical -- each sandbox port gets its own subdomain like `22222-abc123.proxy.daytona.example.com`.

> **Note**: Cloudflare's free plan supports proxied wildcard DNS records, but only at one subdomain level. Using `*.proxy.daytona.example.com` works. Deeper nesting like `*.sandbox.proxy.daytona.example.com` would require Cloudflare Advanced Certificate Manager.

For SSH access, you need a **separate hostname with direct TCP** (not through Cloudflare). Options:
- A dynamic DNS hostname (e.g., `myserver.dynu.net`)
- A direct A record pointing to your server's public IP
- The same domain with a non-standard port forwarded through your router

---

## Step 2: Install Nginx Proxy Manager

If you don't already have Nginx Proxy Manager (NPM) running, create a separate compose file for it. NPM will be the central reverse proxy that routes traffic to your Daytona services.

Create `~/nginx-proxy-manager/docker-compose.yaml`:

```yaml
name: nginx-proxy-manager
services:
  npm:
    image: jc21/nginx-proxy-manager:latest
    restart: always
    ports:
      - 80:80       # HTTP (for Let's Encrypt challenges if needed)
      - 443:443     # HTTPS (not used with Cloudflare tunnel, but NPM needs it)
      - 81:81       # NPM admin panel
    volumes:
      - npm_data:/data
      - npm_letsencrypt:/etc/letsencrypt
    networks:
      - daytona-network

volumes:
  npm_data: {}
  npm_letsencrypt: {}

networks:
  daytona-network:
    external: true
    name: daytona_daytona-network
```

> **Important**: The `daytona-network` must be the same Docker network that Daytona's services use. The Daytona compose file creates a network called `daytona_daytona-network` (the compose project name `daytona` is prefixed automatically). We reference it as an external network here so NPM can reach the Daytona containers by service name.

Start NPM:

```bash
# First, start Daytona once to create the network (or create it manually)
docker network create daytona_daytona-network

# Then start NPM
cd ~/nginx-proxy-manager
docker compose up -d
```

Access the NPM admin panel at `http://your-server-ip:81`:
- Default login: `admin@example.com` / `changeme`
- Change the password immediately on first login

---

## Step 3: Configure Nginx Proxy Manager Proxy Hosts

Create four proxy hosts in NPM. Go to **Hosts > Proxy Hosts > Add Proxy Host** for each:

### 3.1 Dashboard + API (`daytona.example.com`)

| Field | Value |
|---|---|
| Domain Names | `daytona.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `api` |
| Forward Port | `3000` |
| Websockets Support | **ON** (required for notifications) |
| Block Common Exploits | ON |

> WebSockets must be enabled because the dashboard uses Socket.io for real-time notifications via `/api/socket.io/`.

### 3.2 Dex / OIDC (`dex.example.com`)

| Field | Value |
|---|---|
| Domain Names | `dex.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `dex` |
| Forward Port | `5556` |
| Websockets Support | OFF |
| Block Common Exploits | ON |

### 3.3 Proxy - Base Domain (`proxy.daytona.example.com`)

| Field | Value |
|---|---|
| Domain Names | `proxy.daytona.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `proxy` |
| Forward Port | `4000` |
| Websockets Support | **ON** (required for terminal/VNC) |
| Block Common Exploits | OFF |

> WebSockets must be enabled because the terminal (port 22222) and VNC (port 6080) are served through the proxy and rendered in iframes that use WebSocket connections.
>
> Keep "Block Common Exploits" OFF for the proxy -- it can interfere with the wide variety of traffic that passes through sandbox port previews.

### 3.4 Proxy - Wildcard (`*.proxy.daytona.example.com`)

| Field | Value |
|---|---|
| Domain Names | `*.proxy.daytona.example.com` |
| Scheme | `http` |
| Forward Hostname / IP | `proxy` |
| Forward Port | `4000` |
| Websockets Support | **ON** |
| Block Common Exploits | OFF |

This wildcard entry catches all sandbox port preview URLs like `22222-abc123def.proxy.daytona.example.com`.

### SSL Certificates in NPM

Since Cloudflare Tunnel handles SSL termination, traffic between Cloudflare and NPM is already encrypted within the tunnel. You have two options:

**Option A: No SSL in NPM (simpler)**
- Leave the SSL tab empty for all proxy hosts
- Set the Cloudflare Tunnel to use `http://` when forwarding to NPM
- This is fine because the tunnel itself is encrypted

**Option B: SSL in NPM (defense in depth)**
- In the SSL tab, select "Request a new SSL Certificate" with Let's Encrypt
- Use a Cloudflare API token for DNS challenge validation
- Set Cloudflare SSL/TLS mode to "Full (strict)"

Option A is recommended for simplicity.

---

## Step 4: Create the Cloudflare Tunnel

### 4.1 Create the Tunnel

1. Log into [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) (formerly Cloudflare Access)
2. Go to **Networks > Tunnels**
3. Click **Create a tunnel**
4. Choose **Cloudflared** as the connector type
5. Name it (e.g., `daytona-tunnel`)
6. Copy the tunnel token -- you'll need it for the `cloudflared` container

### 4.2 Configure Public Hostnames

In the tunnel configuration, add these public hostname entries. Each one maps an external hostname to your NPM instance:

| Public hostname | Service | Description |
|---|---|---|
| `daytona.example.com` | `http://npm:80` | Dashboard + API |
| `dex.example.com` | `http://npm:80` | OIDC authentication |
| `proxy.daytona.example.com` | `http://npm:80` | Sandbox proxy (base) |
| `*.proxy.daytona.example.com` | `http://npm:80` | Sandbox proxy (wildcard) |

> **Note**: All four entries point to NPM on port 80. NPM handles the routing based on the `Host` header. Replace `npm` with the actual container name or IP if your NPM service has a different name.

> **Important for the wildcard**: In Cloudflare's tunnel config UI, you can enter `*.proxy.daytona.example.com` as the hostname. Cloudflare will automatically create a wildcard DNS record pointing to the tunnel.

### 4.3 Cloudflare Tunnel Settings

For each hostname in the tunnel config, check these settings:

Under **Additional application settings > TLS**:
- **No TLS Verify**: ON (if using Option A from Step 3, since NPM serves HTTP)

Under **Additional application settings > HTTP Settings**:
- **HTTP Host Header**: leave empty (Cloudflare passes the original Host header)

### 4.4 Run cloudflared as a Docker Container

Create `~/cloudflared/docker-compose.yaml`:

```yaml
name: cloudflared
services:
  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: always
    command: tunnel --no-autoupdate run
    environment:
      - TUNNEL_TOKEN=your-tunnel-token-here
    networks:
      - daytona-network

networks:
  daytona-network:
    external: true
    name: daytona_daytona-network
```

Replace `your-tunnel-token-here` with the token from Step 4.1.

```bash
cd ~/cloudflared
docker compose up -d
```

The `cloudflared` container must be on the same Docker network as NPM so it can reach `npm:80`.

### 4.5 Verify DNS Records

After creating the tunnel with public hostnames, Cloudflare should automatically create DNS records. Verify in **Cloudflare Dashboard > DNS > Records**:

| Type | Name | Content | Proxy |
|---|---|---|---|
| CNAME | `daytona` | `<tunnel-id>.cfargotunnel.com` | Proxied (orange cloud) |
| CNAME | `dex` | `<tunnel-id>.cfargotunnel.com` | Proxied |
| CNAME | `proxy.daytona` | `<tunnel-id>.cfargotunnel.com` | Proxied |
| CNAME | `*.proxy.daytona` | `<tunnel-id>.cfargotunnel.com` | Proxied |

All records should show the orange cloud (Proxied). If any are missing, create them manually as CNAME records pointing to `<tunnel-id>.cfargotunnel.com`.

### 4.6 Cloudflare SSL/TLS Settings

In **Cloudflare Dashboard > SSL/TLS > Overview**:
- Set encryption mode to **Full** (or **Full (strict)** if using SSL in NPM)

In **SSL/TLS > Edge Certificates**:
- Verify "Always Use HTTPS" is ON
- Verify "Minimum TLS Version" is 1.2

---

## Step 5: Configure Daytona

### 5.1 Edit the `.env` File

Open `docker/.env` and replace all example domains with your actual domains:

```bash
cd /path/to/daytona/docker
cp .env .env.backup    # keep a backup
```

Edit `.env` and update these values:

```ini
# -- Main domain --
DAYTONA_DOMAIN=daytona.example.com
DAYTONA_URL=https://daytona.example.com

# -- Dashboard --
DASHBOARD_URL=https://daytona.example.com/dashboard
DASHBOARD_BASE_API_URL=https://daytona.example.com

# -- OIDC / Dex --
DEX_PUBLIC_URL=https://dex.example.com/dex
DEX_INTERNAL_URL=http://dex:5556/dex

# -- Proxy --
PROXY_DOMAIN=proxy.daytona.example.com
PROXY_PROTOCOL=https
PROXY_TEMPLATE_URL=https://{{PORT}}-{{sandboxId}}.proxy.daytona.example.com
PROXY_TOOLBOX_BASE_URL=https://proxy.daytona.example.com

# -- SSH Gateway --
SSH_GATEWAY_URL=your-ssh-host.example.com:65022
SSH_GATEWAY_COMMAND=ssh -p 65022 {{TOKEN}}@your-ssh-host.example.com
```

Also change the security-sensitive defaults:

```ini
ENCRYPTION_KEY=<generate-a-random-string>
ENCRYPTION_SALT=<generate-a-different-random-string>
PROXY_API_KEY=<generate-a-random-string>
SSH_GATEWAY_API_KEY=<generate-a-random-string>
DEFAULT_RUNNER_API_KEY=<generate-a-random-string>
OTEL_COLLECTOR_API_KEY=<generate-a-random-string>
HEALTH_CHECK_API_KEY=<generate-a-random-string>
DB_PASSWORD=<generate-a-random-string>
```

You can generate random strings with:

```bash
openssl rand -hex 32
```

### 5.2 Edit the Dex Configuration

Edit `docker/dex/config.yaml`. The **issuer** must exactly match your public Dex URL, and the **redirectURIs** must include your external domains:

```yaml
issuer: https://dex.example.com/dex

staticClients:
  - id: daytona
    redirectURIs:
      - 'https://daytona.example.com'
      - 'https://daytona.example.com/api/oauth2-redirect.html'
      - 'https://proxy.daytona.example.com/callback'
    name: 'Daytona'
    public: true
```

> **Critical**: The `issuer` URL in `dex/config.yaml` must be the public URL that browsers use to reach Dex. If this doesn't match, OIDC token validation will fail with issuer mismatch errors. This is a static YAML file -- it does not support environment variable interpolation.

### 5.3 Understanding the Internal vs. Public URL Split

The setup uses two different URLs for Dex:

- **`DEX_PUBLIC_URL`** (`https://dex.example.com/dex`) -- what browsers see. Used for OIDC discovery, login redirects, and token issuer validation. This is the URL in the `issuer` field of `dex/config.yaml`.
- **`DEX_INTERNAL_URL`** (`http://dex:5556/dex`) -- how the API and proxy services reach Dex inside the Docker network. Used for JWKS key fetching (JWT signature validation).

The API service reads both: `PUBLIC_OIDC_DOMAIN` (public) and `OIDC_ISSUER_BASE_URL` (internal). When it receives a JWT, it:
1. Validates the issuer claim against the public URL
2. Fetches the JWKS keys from the internal URL (rewriting the JWKS URI from the OIDC discovery response)

This is why Dex doesn't need to be reachable from the API container via the public URL.

---

## Step 6: Configure SSH Access

SSH uses raw TCP and **cannot go through Cloudflare Tunnel** (Cloudflare only proxies HTTP/HTTPS and some protocols via WARP, but not arbitrary TCP on custom ports in the free tier).

### Option A: Port Forwarding (home server / router)

1. Forward an external port (e.g., `65022`) on your router to port `2222` on your server
2. Use a dynamic DNS service (e.g., Dynu, DuckDNS, No-IP) if you don't have a static IP

```ini
SSH_GATEWAY_URL=your-ddns-hostname.dynu.net:65022
SSH_GATEWAY_COMMAND=ssh -p 65022 {{TOKEN}}@your-ddns-hostname.dynu.net
```

In `docker-compose.yaml`, the ssh-gateway port mapping should match:

```yaml
ssh-gateway:
  ports:
    - 65022:2222    # external:internal
```

### Option B: Direct Port (VPS with public IP)

If your server has a public IP, expose port 2222 directly (or remap it):

```ini
SSH_GATEWAY_URL=daytona.example.com:2222
SSH_GATEWAY_COMMAND=ssh -p 2222 {{TOKEN}}@daytona.example.com
```

Make sure the DNS record for `daytona.example.com` that points to your server's IP is set to **DNS only** (grey cloud) in Cloudflare, or use a different hostname with a grey-cloud A record.

> **Warning**: If `daytona.example.com` is proxied through Cloudflare (orange cloud), SSH connections to that hostname will fail. Use a separate hostname or a grey-cloud record for SSH.

### Option C: Cloudflare Tunnel with `cloudflared access` (advanced)

Cloudflare Tunnel can proxy TCP via `cloudflared access tcp`, but this requires `cloudflared` installed on every client machine. See [Cloudflare docs on SSH through Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/use-cases/ssh/).

---

## Step 7: Start Daytona

```bash
cd /path/to/daytona/docker

# Pull latest images
docker compose pull

# Start all services
docker compose up -d

# Check logs
docker compose logs -f api
```

Wait for all services to become healthy (30-60 seconds). The API service depends on the database, runner, redis, dex, and registry -- it may restart a few times while waiting for dependencies.

---

## Step 8: Verify the Deployment

### 8.1 Check Service Health

```bash
docker compose ps
```

All services should show `Up` or `Up (healthy)`.

### 8.2 Test the API Config Endpoint

```bash
curl -s https://daytona.example.com/api/config | jq .
```

Verify the response contains your external domains:

```json
{
  "oidc": {
    "issuer": "https://dex.example.com/dex",
    "clientId": "daytona",
    "audience": "daytona"
  },
  "proxyTemplateUrl": "https://{{PORT}}-{{sandboxId}}.proxy.daytona.example.com",
  "proxyToolboxUrl": "https://proxy.daytona.example.com/toolbox",
  "dashboardUrl": "https://daytona.example.com/dashboard",
  "sshGatewayCommand": "ssh -p 65022 {{TOKEN}}@your-ssh-host.example.com"
}
```

If you see `localhost` in any of these values, the `.env` file is not being loaded correctly. Check that the `.env` file is in the same directory as `docker-compose.yaml`.

### 8.3 Test OIDC Discovery

```bash
curl -s https://dex.example.com/dex/.well-known/openid-configuration | jq .issuer
```

Should return: `"https://dex.example.com/dex"`

### 8.4 Test the Dashboard

Open `https://daytona.example.com` in your browser. You should:
1. See the Daytona login page
2. Be redirected to `https://dex.example.com/dex/auth/...` for authentication
3. Log in with `dev@daytona.io` / `password`
4. Be redirected back to the dashboard at `https://daytona.example.com/dashboard`

### 8.5 Test Sandbox Terminal

1. Create a sandbox from the dashboard
2. Open the terminal tab
3. The terminal iframe should load from `https://22222-{token}.proxy.daytona.example.com`
4. If the terminal doesn't load, check that the wildcard proxy host and DNS are configured

### 8.6 Test SSH Access

```bash
# Get the SSH command from a sandbox's details in the dashboard
ssh -p 65022 <token>@your-ssh-host.example.com
```

---

## Troubleshooting

### Dashboard loads but shows "localhost" API errors

**Cause**: `DASHBOARD_BASE_API_URL` is not set correctly, so the dashboard falls back to `window.location.origin` but the API config still returns localhost URLs.

**Fix**: Verify `DASHBOARD_BASE_API_URL` in `.env` matches your external URL (e.g., `https://daytona.example.com`). Restart the API service -- it performs string replacement in the dashboard build files at startup.

```bash
docker compose restart api
```

### Login redirects to localhost:5556

**Cause**: The `/api/config` endpoint returns the wrong OIDC issuer. This happens when `PUBLIC_OIDC_DOMAIN` / `DEX_PUBLIC_URL` is not set.

**Fix**: Check that `DEX_PUBLIC_URL=https://dex.example.com/dex` is set in `.env` and that the API service has `PUBLIC_OIDC_DOMAIN=${DEX_PUBLIC_URL}` in its environment.

### Login fails with "issuer mismatch" or "invalid token"

**Cause**: The `issuer` in `dex/config.yaml` doesn't match `DEX_PUBLIC_URL`.

**Fix**: They must be identical. If `DEX_PUBLIC_URL=https://dex.example.com/dex`, then `dex/config.yaml` must have `issuer: https://dex.example.com/dex`. After changing, restart Dex:

```bash
docker compose restart dex
```

> If Dex previously stored a different issuer in its SQLite database, you may need to delete the volume: `docker compose down dex && docker volume rm daytona_dex_db && docker compose up -d dex`

### Login redirects fail with "redirect URI not registered"

**Cause**: The redirect URI the browser is using is not listed in `dex/config.yaml` under `redirectURIs`.

**Fix**: Add all external URLs to the `redirectURIs` list in `dex/config.yaml`. Common ones needed:
- `https://daytona.example.com` (dashboard callback)
- `https://daytona.example.com/api/oauth2-redirect.html` (Swagger UI)
- `https://proxy.daytona.example.com/callback` (proxy OIDC flow for sandbox auth)

### Terminal/sandbox preview doesn't load

**Cause**: Wildcard DNS or proxy host not configured. The terminal loads in an iframe from `https://22222-{token}.proxy.daytona.example.com`. If this hostname doesn't resolve or isn't routed to the proxy service, the iframe will fail.

**Fix**:
1. Verify DNS: `dig +short test.proxy.daytona.example.com` should return a Cloudflare IP
2. Verify the wildcard entry exists in Cloudflare Tunnel config
3. Verify NPM has a proxy host for `*.proxy.daytona.example.com` with WebSocket support ON
4. Check browser DevTools Network tab for the iframe URL and any errors

### Proxy returns 502 Bad Gateway

**Cause**: NPM can't reach the proxy container. This happens when NPM is not on the same Docker network.

**Fix**: Ensure both NPM and Daytona services share the `daytona_daytona-network` network. You can verify connectivity:

```bash
docker exec -it <npm-container> ping proxy
docker exec -it <npm-container> wget -qO- http://proxy:4000/health
```

### SSH connection refused or times out

**Cause**: SSH uses raw TCP, not HTTP. It cannot go through Cloudflare Tunnel (HTTP-only).

**Fix**:
1. Verify the port is forwarded: from an external network, run `nc -zv your-ssh-host 65022`
2. Check that `docker-compose.yaml` maps the correct external port to `2222`
3. If using a home server, check your router's port forwarding rules
4. Make sure the SSH hostname resolves to your server's actual IP, not a Cloudflare IP

### Changes to `.env` don't take effect

Docker Compose reads `.env` at `up` time. After editing `.env`:

```bash
docker compose down
docker compose up -d
```

Using `docker compose restart` alone will NOT re-read the `.env` file.

---

## File Reference

| File | Purpose |
|---|---|
| `docker/.env` | All configurable environment variables (domains, credentials, ports) |
| `docker/docker-compose.yaml` | Service definitions; references variables from `.env` |
| `docker/dex/config.yaml` | Dex OIDC configuration (issuer URL, redirect URIs, static passwords) |

---

## Security Checklist

Before exposing to the internet:

- [ ] Changed all default passwords in `.env` (`DB_PASSWORD`, `ENCRYPTION_KEY`, `ENCRYPTION_SALT`, API keys)
- [ ] Changed the default Dex login password in `dex/config.yaml` (the `staticPasswords` hash)
- [ ] Cloudflare SSL/TLS mode set to "Full" or "Full (strict)"
- [ ] SSH port is not exposed through Cloudflare (uses direct TCP)
- [ ] NPM admin panel (port 81) is not accessible from the internet
- [ ] PgAdmin (port 5050), Registry UI (port 5100), MinIO Console (port 9001), MailDev (port 1080), and Jaeger (port 16686) are not accessible from the internet
- [ ] Consider removing the `ports` mappings for internal-only services from `docker-compose.yaml` in production

To generate a new Dex password hash:

```bash
# Replace 'your-new-password' with your desired password
echo 'your-new-password' | htpasswd -BinC 10 admin | cut -d: -f2
```
