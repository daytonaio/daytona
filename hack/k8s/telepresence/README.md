# Telepresence Configuration for Daytona Development

This directory contains Telepresence configuration for running Daytona services locally while connected to the Kubernetes cluster network.

## Prerequisites

1. **Telepresence is pre-installed** in the devcontainer (v2.21.1)

   Verify installation:

   ```bash
   telepresence version
   ```

2. Set KUBECONFIG:

   ```bash
   export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml
   ```

## Quick Start

### Install Traffic Manager (First Time Only)

The first time you use Telepresence with a cluster, install the traffic manager:

```bash
telepresence helm install --namespace daytona-dev
```

### Connect to the Cluster

Then connect Telepresence to the cluster:

```bash
telepresence connect --namespace daytona-dev
```

### Verify Connection

Check that you can access cluster services:

```bash
# These should resolve and return responses
curl http://db:5432 -v 2>&1 | head -5
curl http://redis:6379 -v 2>&1 | head -5
curl http://dex:5556/dex/.well-known/openid-configuration
```

## Running Services Locally (Recommended)

Since the app services (api, proxy, ssh-gateway) are NOT deployed in the cluster,
you simply run them locally while connected to Telepresence. Your local services
can access all cluster services (db, redis, dex, minio, etc.).

```bash
# Terminal 1: Ensure Telepresence is connected
telepresence connect -n daytona-dev

# Terminal 2: Run API locally (can reach db:5432, redis:6379, dex:5556, etc.)
yarn nx serve api

# Terminal 3: Run Proxy locally
yarn nx serve proxy

# Terminal 4: Run SSH Gateway locally
yarn nx serve ssh-gateway
```

## Intercept Services (When Deployed in Cluster)

If services ARE deployed in the cluster and you want to redirect traffic to your
local instance for debugging:

### Intercept API Service

```bash
# Start intercept (redirects cluster traffic to local port 3000)
telepresence intercept api -n daytona-dev --port 3000:3000

# Now run the API locally
yarn nx serve api
```

### Intercept Proxy Service

```bash
# Start intercept (redirects traffic to local port 4000)
telepresence intercept proxy -n daytona-dev --port 4000:4000

# Now run the proxy locally
yarn nx serve proxy
```

### Intercept SSH Gateway Service

```bash
# Start intercept (redirects traffic to local port 2222)
telepresence intercept ssh-gateway -n daytona-dev --port 2222:2222

# Now run the ssh-gateway locally
yarn nx serve ssh-gateway
```

## Environment Variables for Local Development

When running services locally with Telepresence, use these environment configurations:

### API Service (.env.local)

```bash
PORT=3000
NODE_ENV=development

# Database (accessible via telepresence)
DB_HOST=db
DB_PORT=5432
DB_USERNAME=user
DB_PASSWORD=pass
DB_DATABASE=application_ctx

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# OIDC (Dex)
OIDC_CLIENT_ID=daytona
OIDC_ISSUER_BASE_URL=http://dex:5556/dex
PUBLIC_OIDC_DOMAIN=http://localhost:5556/dex
OIDC_AUDIENCE=daytona

# S3 (Minio)
S3_ENDPOINT=http://minio:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_DEFAULT_BUCKET=daytona

# SMTP (Maildev)
SMTP_HOST=maildev
SMTP_PORT=1025

# Registry
TRANSIENT_REGISTRY_URL=http://registry:5000
INTERNAL_REGISTRY_URL=http://registry:5000
```

### Proxy Service (.env.local)

```bash
DAYTONA_API_URL=http://api:3000/api
PROXY_PORT=4000
PROXY_DOMAIN=proxy:4000
PROXY_PROTOCOL=http
PROXY_API_KEY=super_secret_key
OIDC_CLIENT_ID=daytona
OIDC_DOMAIN=http://dex:5556/dex
OIDC_AUDIENCE=daytona
REDIS_HOST=redis
REDIS_PORT=6379
```

### SSH Gateway Service (.env.local)

```bash
API_URL=http://api:3000/api
API_KEY=ssh_secret_api_token
SSH_GATEWAY_PORT=2222
```

## Useful Commands

```bash
# List active intercepts
telepresence list

# Leave an intercept
telepresence leave api

# Disconnect from the cluster
telepresence quit

# View connection status
telepresence status

# View telepresence logs
telepresence gather-logs
```

## Troubleshooting

### DNS Resolution Issues

If services don't resolve, ensure you're connected:

```bash
telepresence status
```

Try reconnecting:

```bash
telepresence quit
telepresence connect --namespace daytona-dev
```

### Port Conflicts

If a port is already in use locally, specify a different local port:

```bash
telepresence intercept api --port 3001:3000
```

This maps cluster port 3000 to local port 3001.

### Traffic Manager Issues

If the traffic manager has issues, reset it:

```bash
telepresence uninstall --agent api --namespace daytona-dev
telepresence quit
telepresence connect --namespace daytona-dev
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   daytona-dev namespace                  │   │
│  │                                                          │   │
│  │  ┌─────┐  ┌─────┐  ┌───────┐  ┌─────┐  ┌────────────┐  │   │
│  │  │ db  │  │redis│  │  dex  │  │minio│  │  registry  │  │   │
│  │  └─────┘  └─────┘  └───────┘  └─────┘  └────────────┘  │   │
│  │                                                          │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │     Telepresence Traffic Manager                │    │   │
│  │  │  (intercepts traffic and routes to local)       │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  │           │              │              │                │   │
│  │      [api:3000]    [proxy:4000]  [ssh-gateway:2222]     │   │
│  └──────────│──────────────│──────────────│────────────────┘   │
│             │              │              │                     │
└─────────────│──────────────│──────────────│─────────────────────┘
              │              │              │
              ▼              ▼              ▼
    ┌─────────────────────────────────────────────────────────┐
    │                    Local Machine                         │
    │                                                          │
    │   ┌──────────┐   ┌──────────┐   ┌───────────────┐       │
    │   │ API      │   │ Proxy    │   │ SSH Gateway   │       │
    │   │ :3000    │   │ :4000    │   │ :2222         │       │
    │   │          │   │          │   │               │       │
    │   │ (debug)  │   │ (debug)  │   │ (debug)       │       │
    │   └──────────┘   └──────────┘   └───────────────┘       │
    │                                                          │
    └─────────────────────────────────────────────────────────┘
```
