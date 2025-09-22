# Docker Compose Setup for Daytona

This folder contains a Docker Compose setup for running Daytona locally.

⚠️ **Important**:

- This setup is still in development and is **not safe to use in production**
- A separate deployment guide will be provided for production scenarios

## Overview

The Docker Compose configuration includes all the necessary services to run Daytona:

- **API**: Main Daytona application server
- **Proxy**: Request proxy service
- **Runner**: Service that hosts the Daytona Runner
- **SSH Gateway**: Service that handles sandbox SSH access
- **Database**: PostgreSQL database for data persistence
- **Redis**: In-memory data store for caching and sessions
- **Dex**: OIDC authentication provider
- **Registry**: Docker image registry with web UI
- **MinIO**: S3-compatible object storage
- **MailDev**: Email testing service
- **Jaeger**: Distributed tracing
- **PgAdmin**: Database administration interface

## Quick Start

1. Start all services (from the root of the Daytona repo):

   ```bash
   docker compose -f docker/docker-compose.yaml up -d
   ```

2. Access the services:
   - Daytona Dashboard: http://localhost:3000
     - Access Credentials: dev@daytona.io `password`
     - Make sure that the default snapshot is active at http://localhost:3000/dashboard/snapshots
   - PgAdmin: http://localhost:5050
   - Registry UI: http://localhost:5100
   - MinIO Console: http://localhost:9001 (minioadmin / minioadmin)

## DNS Setup for Proxy URLs

For local development, you need to resolve `*.proxy.localhost` domains to `127.0.0.1`:

```bash
./scripts/setup-proxy-dns.sh
```

This configures dnsmasq with `address=/proxy.localhost/127.0.0.1`.

**Without this setup**, SDK examples and direct proxy access won't work.

## Development Notes

- The setup uses shared networking for simplified service communication
- Database and storage data is persisted in Docker volumes
- The registry is configured to allow image deletion for testing
- Sandbox resource limits are disabled due to inability to partition cgroups in DinD environment where the sock is not mounted

<br><br><br>

# Auth0 Configuration Guide for Daytona

## Step 1: Create Your Auth0 Tenant

Begin by navigating to https://auth0.com/signup and start the signup process. Choose your account type based on your use case - select `Company` for business applications or `Personal` for individual projects.\
On the "Let's get setup" page, you'll need to enter your application name such as `My Daytona` and select `Single Page Application (SPA)` as the application type. For authentication methods, you can start with `Email and Password` since additional social providers like Google, GitHub, or Facebook can be added later. Once you've configured these settings, click `Create Application` in the bottom right corner.

## Step 2: Configure Your Single Page Application

Navigate to `Applications` > `Applications` in the left sidebar and select the application you just created. Click the `Settings` tab and scroll down to find the `Application URIs` section where you'll configure the callback and origin URLs.
In the `Allowed Callback URIs` field, add the following URLs:

```
http://localhost:3000
http://localhost:3000/api/oauth2-redirect.html
http://localhost:4000/callback
http://proxy.localhost:4000/callback
```

For `Allowed Logout URIs`, add:

```
http://localhost:3000
```

And for `Allowed Web Origins`, add:

```
http://localhost:3000
```

Remember to click `Save Changes` at the bottom of the page to apply these configurations.

## Step 3: Create Machine-to-Machine Application

You'll need a Machine-to-Machine application to interact with Auth0's Management API. Go to `Applications` > `Applications` and click `Create Application`. Choose `Machine to Machine Applications` as the type and provide a descriptive name like `My Management API M2M`.
After creating the application, navigate to the `APIs` tab within your new M2M application. Find and authorize the `Auth0 Management API` by clicking the toggle or authorize button.\
Once authorized, click the dropdown arrow next to the Management API to configure permissions. Grant the following permissions to your M2M application:

```
read:users
update:users
read:connections
create:guardian_enrollment_tickets
read:connections_options
```

Click `Save` to apply these permission changes.

## Step 4: Set Up Custom API

Your Daytona application will need a custom API to handle authentication and authorization. Navigate to `Applications` > `APIs` in the left sidebar and click `Create API`. Enter a descriptive name such as `My Daytona API` and provide an identifier like `my-daytona-api`. The identifier should be a unique string that will be used in your application configuration.\
After creating the API, go to the `Permissions` tab to define the scopes your application will use. Add each of the following permissions with their corresponding descriptions:

| Permission | Description |
|------------|-------------|
| `read:node` | Get workspace node info |
| `create:node` | Create new workspace node record |
| `create:user` | Create user account |
| `read:users` | Get all user accounts |
| `regenerate-key-pair:users` | Regenerate user SSH key-pair |
| `read:workspaces` | Read workspaces (user scope) |
| `create:registry` | Create a new docker registry auth record |
| `read:registries` | Get all docker registry records |
| `read:registry` | Get docker registry record |
| `write:registry` | Create or update docker registry record |

## Step 5: Configure Environment Variables

Once you've completed all the Auth0 setup steps, you'll need to configure environment variables in your Daytona deployment. These variables connect your application to the Auth0 services you've just configured.

### Finding Your Configuration Values

You can find the necessary values in the Auth0 dashboard. For your SPA application settings, go to `Applications` > `Applications`, select your SPA app, and click the `Settings` tab. For your M2M application, follow the same path but select your Machine-to-Machine app instead. Custom API settings are located under `Applications` > `APIs`, then select your custom API and go to `Settings`.

### API Service Configuration

Configure the following environment variables for your API service:

```bash
OIDC_CLIENT_ID=your_spa_app_client_id
OIDC_ISSUER_BASE_URL=your_spa_app_domain
OIDC_AUDIENCE=your_custom_api_identifier
OIDC_MANAGEMENT_API_ENABLED=true
OIDC_MANAGEMENT_API_CLIENT_ID=your_m2m_app_client_id
OIDC_MANAGEMENT_API_CLIENT_SECRET=your_m2m_app_client_secret
OIDC_MANAGEMENT_API_AUDIENCE=your_auth0_managment_api_identifier
```

### Proxy Service Configuration

For your proxy service, configure these environment variables:

```bash
OIDC_CLIENT_ID=your_spa_app_client_id
OIDC_CLIENT_SECRET=
OIDC_DOMAIN=your_spa_app_domain
OIDC_AUDIENCE=your_custom_api_identifier (with trailing slash)
```

Note that `OIDC_CLIENT_SECRET` should remain empty for your proxy environment.
