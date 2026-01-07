# Daytona Production Deployment

This directory contains scripts and configuration for deploying Daytona to a production Ubuntu server.

## Prerequisites

- A clean Ubuntu 24.04 server
- SSH access to the server (root or sudo user)
- Domain names configured:
  - `win.trydaytona.com` - Main domain for Dashboard and API
  - `*.preview.win.trydaytona.com` - Wildcard domain for sandbox proxy

## Quick Start

### 1. First-time Setup

Run the setup command to install Docker, nginx, and obtain SSL certificates:

```bash
./deploy.sh setup
```

This will:

- Install Docker and Docker Compose
- Install and configure nginx
- Obtain SSL certificates from Let's Encrypt
- Configure the firewall

### 2. Configure Environment

Copy the environment template and customize:

```bash
cp env.example .env
vim .env
```

Key variables to configure:

- `DB_PASSWORD` - PostgreSQL password
- `PROXY_API_KEY` - Secret key for proxy service
- `RUNNER_API_KEY` - Secret key for runner service
- `SSH_GATEWAY_API_KEY` - Secret key for SSH gateway
- `SSH_PRIVATE_KEY` - Base64-encoded SSH private key
- `SSH_HOST_KEY` - Base64-encoded SSH host key

### 3. Deploy

Run the full deployment:

```bash
./deploy.sh deploy
```

This will:

1. Sync project files to the remote server
2. Build Docker images
3. Start all services

## Commands

| Command | Description |
|---------|-------------|
| `./deploy.sh setup` | Initial server setup (Docker, nginx, SSL) |
| `./deploy.sh sync` | Sync project files to remote |
| `./deploy.sh build` | Build Docker images on remote |
| `./deploy.sh start` | Start all services |
| `./deploy.sh stop` | Stop all services |
| `./deploy.sh restart` | Restart all services |
| `./deploy.sh logs [service]` | View logs (optionally for specific service) |
| `./deploy.sh status` | Show service status |
| `./deploy.sh deploy` | Full deployment (sync + build + start) |

## Options

```bash
./deploy.sh [OPTIONS] COMMAND

Options:
  -h, --host HOST     Remote host (default: h1004.blinkbox.dev)
  -u, --user USER     Remote user (default: root)
  -k, --key KEY       SSH private key path
  -p, --path PATH     Remote path (default: /opt/daytona)
```

## Architecture

```
                    Internet
                        │
                        ▼
              ┌─────────────────┐
              │     nginx       │ (SSL termination)
              │   Port 80/443   │
              └────────┬────────┘
                       │
         ┌─────────────┼─────────────┐
         │             │             │
         ▼             ▼             ▼
    /dashboard      /api/*      *.proxy.*
    (static)     (API service)   (Proxy)
         │             │             │
         └──────┬──────┘             │
                │                    │
         ┌──────▼──────┐      ┌──────▼──────┐
         │     API     │      │    Proxy    │
         │  Port 3000  │      │  Port 4000  │
         └──────┬──────┘      └─────────────┘
                │
    ┌───────────┼───────────┬───────────┐
    │           │           │           │
    ▼           ▼           ▼           ▼
┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐
│  DB   │  │ Redis │  │  Dex  │  │ MinIO │
└───────┘  └───────┘  └───────┘  └───────┘

SSH Gateway (Port 2222) - Direct access, not through nginx
```

## Services

| Service | Internal Port | External Access |
|---------|---------------|-----------------|
| Dashboard | 3000 | https://win.trydaytona.com/ |
| API | 3000 | https://win.trydaytona.com/api/ |
| Proxy | 4000 | https://*.preview.win.trydaytona.com/ |
| SSH Gateway | 2222 | win.trydaytona.com:2222 |
| Dex (OIDC) | 5556 | https://win.trydaytona.com/dex/ |
| PostgreSQL | 5432 | Internal only |
| Redis | 6379 | Internal only |
| MinIO | 9000/9001 | Internal only |
| Registry | 6000 | Internal only |

## SSL Certificates

### Main Domain

The main domain certificate is obtained automatically via HTTP challenge:

```bash
certbot certonly --webroot -w /var/www/certbot -d win.trydaytona.com
```

### Wildcard Domain (Proxy)

The wildcard certificate requires DNS challenge:

```bash
certbot certonly --manual --preferred-challenges dns \
    -d 'preview.win.trydaytona.com' -d '*.preview.win.trydaytona.com' \
    --email admin@daytona.io --agree-tos
```

You'll need to create a DNS TXT record during the process.

## DNS Configuration

Add these DNS records:

| Type | Name | Value |
|------|------|-------|
| A | win.trydaytona.com | `<server-ip>` |
| A | *.preview.win.trydaytona.com | `<server-ip>` |

## Generating SSH Keys

To generate new SSH keys for the gateway:

```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 3072 -f ssh_key -N ""

# Generate host key
ssh-keygen -t rsa -b 3072 -f ssh_host_key -N ""

# Encode for .env file
echo "SSH_PRIVATE_KEY=$(base64 -w 0 ssh_key)"
echo "SSH_HOST_KEY=$(base64 -w 0 ssh_host_key)"
echo "SSH_GATEWAY_PUBLIC_KEY=$(base64 -w 0 ssh_key.pub)"
```

## Troubleshooting

### View Logs

```bash
# All services
./deploy.sh logs

# Specific service
./deploy.sh logs api
./deploy.sh logs proxy
./deploy.sh logs ssh-gateway
```

### Check Service Status

```bash
./deploy.sh status
```

### Restart Services

```bash
./deploy.sh restart
```

### Manual Docker Commands on Remote

```bash
ssh root@win.trydaytona.com
cd /opt/daytona/scripts/deploy
docker compose -f docker-compose.production.yaml ps
docker compose -f docker-compose.production.yaml logs -f api
```

### Nginx Issues

```bash
# Test configuration
nginx -t

# View error logs
tail -f /var/log/nginx/error.log

# Reload configuration
systemctl reload nginx
```

### Certificate Issues

```bash
# Check certificate status
certbot certificates

# Force renewal
certbot renew --force-renewal
```

## Default Credentials

| Service | Username | Password |
|---------|----------|----------|
| Dashboard | dev@daytona.io | password |
| PgAdmin | dev@daytona.io | pgadmin |
| MinIO | minioadmin | minioadmin |

**Important:** Change these credentials in production!

## Security Recommendations

1. **Change all default passwords** in `.env`
2. **Use strong API keys** for services
3. **Restrict SSH access** to specific IPs if possible
4. **Enable fail2ban** for brute force protection
5. **Set up monitoring** and alerting
6. **Regular backups** of PostgreSQL and MinIO data

## Backup

### Database Backup

```bash
docker exec daytona-db-1 pg_dump -U user daytona > backup.sql
```

### Restore Database

```bash
cat backup.sql | docker exec -i daytona-db-1 psql -U user daytona
```
