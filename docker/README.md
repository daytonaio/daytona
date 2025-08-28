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

## Development Notes

- The setup uses shared networking for simplified service communication
- Database and storage data is persisted in Docker volumes
- The registry is configured to allow image deletion for testing
- Sandbox resource limits are disabled due to inability to partition cgroups in DinD environment where the sock is not mounted
