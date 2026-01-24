# Troubleshooting Log - Docker Hub Credentials & Runner Stability

## [2026-01-24 15:16] Runner Startup Issues
- **Problem**: Runner logs showed "is the docker daemon running?" and failed to initialize properly.
- **Cause**: The `daytona-runner` process was starting before the `dockerd` background process (DinD) had created the `/var/run/docker.sock`.
- **Fix**: Modified [Dockerfile.local](file:///c:/Users/hjsgo/Projects/daytona/apps/runner/Dockerfile.local) entrypoint to wait for the socket:
  ```bash
  while [ ! -S /var/run/docker.sock ]; do sleep 1; done && daytona-runner
  ```
- **Status**: Fixed. Runner now successfully tags images and monitors events.

## [2026-01-24 15:18] API "No available runners"
- **Problem**: API returns `400 Bad Request` with `No available runners` even though the runner container is up.
- **Diagnostics**:
    - Checked `runner` table in Postgres.
    - Runner state is `ready`, but `availabilityScore` was `0`.
    - API `RunnerService` requires `availabilityScore >= 10` (default config).
- **Target**: Manually bump `availabilityScore` to 100 to allow testing while the health check stabilizes.

## [2026-01-24 15:20] PowerShell Syntax Error
- **Problem**: Used `&&` in `run_command` which failed in PowerShell.
- **Fix**: Replaced with `;` or ran commands separately.

## [2026-01-24 15:30] Binary Rebuild and Logging
- **Problem**: New logging in `image_pull.go` was not appearing in `docker logs`.
- **Cause 1**: The Docker image was using a cached binary from `dist/apps/runner/daytona-runner`.
- **Cause 2**: Default `LOG_LEVEL` in the runner is `warn`, but logs were added as `info`.
- **Action**:
    1. Rebuilt binary with `GOOS=linux GOARCH=amd64`.
    2. Updated [Dockerfile.local](file:///c:/Users/hjsgo/Projects/daytona/apps/runner/Dockerfile.local) to include `ENV LOG_LEVEL=info`.
- **Status**: Rebuilding and redeploying.

## [2026-01-24 15:35] Artifact History Confirmation
- **User Question**: Are all task history entries since last night recorded as artifacts?
- **Answer**: Yes.
    - Conversation `20cf5d5c-b770-4eb0-b90c-6a3af5c5940d` (Last night) contains `task.md` and `implementation_plan.md`.
    - Current conversation `68143d23-8c88-4d5c-b59f-90798a08b737` also contains its own artifacts.
    - All progress is being tracked across blocks.

## [2026-01-24 16:30] Python SDK Connectivity and Region Mismatch
- **Problem**: Python SDK `create_sandbox()` failed with `Region not found`.
- **Root Cause 1 (Import)**: Installed package `daytona-sdk` exports as `daytona_sdk`, not `daytona`.
- **Root Cause 2 (Client Config)**: Getting "Region not found" because `DaytonaConfig(target=...)` was used. `target` parameter seems to trigger logic unsuitable for this environment. Using `api_url=...` is required.
- **Root Cause 3 (Server Endpoint)**: The SDK calls `GET /region` (singular) which returned 404. API only provided `GET /api/regions`.
- **Fixes**:
    1.  **Server**: Implemented `DefaultRegionController` to handle `GET /region` and return the default 'us' region. Includes update to `OrganizationModule`.
    2.  **Server**: Updated `RegionController` to default `includeShared=true`.
    3.  **Client Guidance**: Users must use `api_url` parameter in `DaytonaConfig`.
- **Status**: Verified. SDK now successfully creates sandboxes.
