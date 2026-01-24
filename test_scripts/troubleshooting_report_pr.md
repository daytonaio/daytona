# Troubleshooting Report: Docker Hub Credential Passing & Runner Stability

## 1. Problem Symptoms
- **Snapshot/Sandbox Creation Pending**: When creating a snapshot or sandbox using an image without a registry prefix (e.g., `alpine`, `postgres`), the process would hang in `pending` or fail silently.
- **Missing Credentials Trace**: Runner logs did not show the "using credentials" message for Docker Hub pulls, indicating it was falling back to anonymous pulls (subject to rate limiting or private repo failure).
- **Runner Connectivity Errors**: Runner logs frequently showed `Is the docker daemon running?` or `events stream error`, causing the API to mark the runner as `UNRESPONSIVE` or `0` availability score.

## 2. Root Cause Analysis

### System Relationships & Workflow
The issue stemmed from a breakdown in the communication chain between the **API**, **Database**, and **Runner**:

1.  **Registry Mapping (API & DB)**: 
    - The `DockerRegistryService` in the API is responsible for mapping an image name (e.g., `postgres:17.2`) to a registry configuration in the DB.
    - If the image name lacks a prefix, the API looks for a registry entry specifically flagged as "Docker Hub" or matching `index.docker.io/v1/`.
    - **Failure Point**: The database entry for Docker Hub often used `https://index.docker.io/v1/` or `docker.io`, while the code logic expected a strictly formatted URL. This caused the mapping to fail, resulting in anonymous pull requests.

2.  **Credential Passing (API -> Runner)**:
    - Once the registry is identified, the API fetches the credentials and passes them to the Runner via its REST API.
    - **Failure Point**: Without the correct mapping in step 1, no credentials were sent.

3.  **Runner Readiness (DinD Architecture)**:
    - The Runner runs in a Docker-in-Docker (DinD) environment. The `daytona-runner` service depends on the internal `dockerd` daemon.
    - **Failure Point**: There was a race condition where `daytona-runner` started before `dockerd` created `/var/run/docker.sock`. This led to initialization failures, preventing the Runner from reporting its health to the API.

## 3. Resolution Steps

### Technical Fixes
1.  **Environment Stabilization**: 
    - Modified `apps/runner/Dockerfile.local` to include a `while` loop in the `ENTRYPOINT`. This ensures `daytona-runner` exclusively starts *after* the Docker socket is ready.
    - Added `LOG_LEVEL=info` to the Runner environment to ensure credential usage is visible in production logs.
2.  **Credential Logic Correction**:
    - Fixed `DockerRegistryService.ts` to correctly handle prefix-less images by defaulting to the "Docker Hub" registry entry.
    - Added explicit logging in the Runner's `image_pull.go` to audit credential usage:
      `Pulling image %s using credentials for registry %s (User: %s)`
3.  **Database Alignment**:
    - Standardized the Docker Hub registry URL to `index.docker.io/v1/` in the database to align with the API's mapping logic.

### User Considerations & Precautions ⚠️
- **Binary Rebuild Mandatory**: The `daytona-runner` binary must be compiled for Linux (`GOOS=linux GOARCH=amd64`) before building the Docker image. Changes to Go code will **not** take effect if simply running `docker compose build` without a fresh `go build`.
- **Registry URL Precision**: When adding manual registries to the database, the URL must match exactly what the `parseDockerImage` utility expects. For Docker Hub, always use `index.docker.io/v1/`.
- **Availability Score**: If the dashboard shows "No available runners", check the `availabilityScore` in the `runner` table. The API requires a score >= 10 (default) to schedule tasks.

## 4. Python SDK Troubleshooting

### Problem Symptoms
- **Region Not Found**: The Python SDK (`daytona-sdk==0.128.1`) failed during `daytona.create()` with a `Region not found` error.
- **Import Failures**: The script failed to import `daytona` because the package name is `daytona_sdk`.

### Root Cause Analysis
1.  **API Route Mismatch**:
    - The SDK attempts to fetch the default region via `GET /api/region` (singular).
    - The API only exposed `/api/regions` (plural).
    - This resulted in a 404 error, which the SDK interpreted as "Region not found".
2.  **SDK Parameter handling**:
    - The SDK's behavior changes based on whether `target` or `api_url` is used in `DaytonaConfig`. Using `target` caused issues in this environment; `api_url` is the correct parameter for direct API connection.

### Resolution Steps
1.  **Server-Side Fixes**:
    - Implemented a new `DefaultRegionController` to handle the `GET /api/region` route.
    - Updated `RegionController` to default to `includeShared=true` so that the default 'us' region is discoverable.
2.  **Client-Side Correction**:
    - Verified that using `from daytona_sdk import ...` is required.
    - Verified that configuring `DaytonaConfig` with `api_url` instead of `target` is necessary for stable operation.

### Verification of Fixes
- **Credential Logging**: Verified via `docker logs daytona-runner-1` that credential passing is active (seen in `postgres` test). For SDK tests using public images, caching prevented new pulls, but the mechanism remains validated.
- **SDK Success**: The test script `test_sdk.py` now successfully creates, uses, and returns a sandbox using the patched flow.

---
*Created on: 2026-01-24*

## 5. Final Verification Results

**Credential Logging Verification**:
- **Test 1**: Created a sandbox using `redis:7.0-alpine` (uncached) with valid Docker Hub credentials (`hyoungjunnoh`).
  - **Result**: `level=info msg="Pulling image redis:7.0-alpine using credentials for registry index.docker.io/v1/ (User: hyoungjunnoh)"`
- **Test 2**: Created a sandbox using `nginx:1.25-alpine` (uncached) to confirm consistency.
  - **Result**: `level=info msg="Pulling image nginx:1.25-alpine using credentials for registry index.docker.io/v1/ (User: hyoungjunnoh)"`
- **Conclusion**: This definitively proves that the `DockerRegistryService` correctly maps prefix-less images to the Docker Hub registry entry and that the Runner receives and uses the configured credentials consistently across different images.

**API Path Mismatch Investigation**:
- **Findings**: The `RegionController` was introduced in commit `af8fceb` (Jan 23, 2026) with the path `@Controller('regions')`.
- **Conclusion**: The mismatch occurred at this inception point. The server implementation used plural `/regions`, while the client SDK (likely generated from a spec or convention) expected singular `/region`. This divergence has been resolved by adding `DefaultRegionController` to handle the singular path.

---

## 6. Docker Hub Authentication Fix (Manifest API)

**Problem**: After the initial credential passing fix, sandbox creation still failed with "manifest unknown" errors in the audit logs. API logs showed:
```
Could not get image details for <image>: Failed to get manifest for image <image>: Not Found
```

**Root Cause**: The API was attempting to use **Basic Authentication directly on Docker Hub's manifest API**. However, Docker Hub's Registry API v2 **only accepts Bearer token authentication** for manifest requests, not Basic Auth. This caused all manifest lookup requests to fail with 401/404 errors.

**Understanding Docker Hub's Authentication**:
- Docker Hub uses a **two-step authentication process**:
  1. **Token Service** (`auth.docker.io`): Issues Bearer tokens (accepts Basic Auth with credentials)
  2. **Manifest API** (`index.docker.io/v2/...`): Serves image data (only accepts Bearer tokens)

**The Bug**: When Docker Hub credentials existed in the database, the original code sent Basic Auth directly to the manifest API, which Docker Hub rejected.

**Solution**: Modified `apps/api/src/docker-registry/services/docker-registry.service.ts`:

1. **Updated `getDockerHubToken` method** (lines 382-401):
   - Added optional `username` and `password` parameters
   - When credentials are provided, uses Basic Auth to **request an authenticated Bearer token** from `https://auth.docker.io/token`
   - Falls back to anonymous token if no credentials provided
   - **Key**: Basic Auth is used to GET a token, not for the manifest API

2. **Updated `getImageDetails` method** (lines 493-506):
   - **Separated Docker Hub from other registries**
   - For Docker Hub: Always use Bearer token (authenticated if credentials available, anonymous otherwise)
   - For other registries: Continue using Basic Auth (unchanged)

**Authentication Flow**:
```
Docker Hub:
  API → auth.docker.io (Basic Auth) → Bearer Token → index.docker.io (Bearer Token) → Manifest ✅

Other Registries:
  API → registry (Basic Auth) → Manifest ✅
```

**Verification**:
- **Test 1 (nginx:1.25-alpine)**: ✓ Sandbox created successfully
- **Test 2 (python:3.11-slim)**: ✓ Sandbox created successfully
- **Test 3 (redis:7.0-alpine)**: ✓ Sandbox created successfully
- **API Logs**: No "Failed to get manifest" errors after fix

**Impact**: Sandbox creation from Docker Hub images now works correctly. The API can successfully retrieve image manifests using authenticated Docker Hub credentials. Other registry authentication (Basic Auth) remains unchanged.
