# Docker Hub Authentication Fix - Detailed Summary

## Problem
User reported sandbox creation failures in the dashboard with "manifest unknown" errors in the audit logs. Investigation revealed:

1. **API Logs showed**: `Could not get image details for <image>: Failed to get manifest for image <image>: Not Found`
2. **Root Cause**: The API was using **Basic Authentication directly on Docker Hub's manifest API**, but Docker Hub's Registry API v2 **only accepts Bearer token authentication** for manifest requests.

## Understanding Docker Hub Authentication

### Docker Hub's Two-Step Authentication Process

Docker Hub uses a **token-based authentication system**:

1. **Token Service** (`auth.docker.io`): Issues Bearer tokens
   - Accepts Basic Auth with username/password
   - Returns a Bearer token for specific repository access
   
2. **Registry API** (`index.docker.io/v2/...`): Serves image manifests
   - **Only accepts Bearer tokens**
   - Rejects Basic Auth requests

### The Problem with the Original Code

**Before Fix** (lines 484-493):
```typescript
if (registry.username && registry.password) {
  // ❌ Tried to use Basic Auth for ALL registries including Docker Hub
  const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
  baseHeaders.set('Authorization', `Basic ${encodedCredentials}`)
} else if (registry.url.includes(DOCKER_HUB_REGISTRY)) {
  // Only used Bearer token when NO credentials (anonymous)
  bearerToken = await this.getDockerHubToken(dockerHubRepo)
}
```

**Issue**: When Docker Hub credentials existed in the database, the code sent Basic Auth directly to `index.docker.io/v2/.../manifests/...`, which Docker Hub rejected with 401/404.

## Solution

### Code Changes to `apps/api/src/docker-registry/services/docker-registry.service.ts`

#### 1. Updated `getDockerHubToken` Method (lines 382-401)

**Added credential support for authenticated token requests**:

```typescript
private async getDockerHubToken(repository: string, username?: string, password?: string): Promise<string | null> {
  const tokenUrl = `https://auth.docker.io/token?service=${DOCKER_HUB_REGISTRY}&scope=repository:${repository}:pull`
  
  const config: any = { timeout: 10000 }
  if (username && password) {
    // Use Basic Auth to REQUEST a Bearer token from auth.docker.io
    const encodedCredentials = Buffer.from(`${username}:${password}`).toString('base64')
    config.headers = { 'Authorization': `Basic ${encodedCredentials}` }
  }
  
  const response = await axios.get(tokenUrl, config)
  return response.data.token  // Returns Bearer token
}
```

**Key Point**: Basic Auth is used **only to request a token from `auth.docker.io`**, not for the manifest API.

#### 2. Updated `getImageDetails` Method (lines 493-506)

**Separated Docker Hub from other registries**:

```typescript
// Docker Hub requires Bearer tokens, not Basic Auth
if (registry.url.includes(DOCKER_HUB_REGISTRY)) {
  // ✅ Always use Bearer token for Docker Hub
  const dockerHubRepo = repoPath.includes('/') ? repoPath : `library/${repoPath}`
  bearerToken = await this.getDockerHubToken(dockerHubRepo, registry.username, registry.password)
  if (bearerToken) {
    baseHeaders.set('Authorization', `Bearer ${bearerToken}`)
  }
} else if (registry.username && registry.password) {
  // ✅ Other registries continue using Basic Auth
  const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
  baseHeaders.set('Authorization', `Basic ${encodedCredentials}`)
}
```

### Authentication Flow Comparison

#### Docker Hub Flow (After Fix)
```
1. API → auth.docker.io/token (Basic Auth: username:password)
2. auth.docker.io → API (Bearer token)
3. API → index.docker.io/v2/.../manifests/... (Bearer token)
4. index.docker.io → API (Manifest data) ✅
```

#### Other Registries Flow (Unchanged)
```
1. API → registry/v2/.../manifests/... (Basic Auth: username:password)
2. registry → API (Manifest data) ✅
```

## What Changed vs. What Stayed the Same

### Changed
- **Docker Hub authentication**: Switched from Basic Auth to Bearer Token for manifest API
- **Code structure**: Separated Docker Hub logic from other registries

### Stayed the Same
- **Other registries**: Still use Basic Auth (line 502-506)
- **Anonymous Docker Hub access**: Still works (when no credentials provided)
- **Overall authentication logic**: Just reorganized, not fundamentally changed

## Deployment
```bash
docker compose -f docker/docker-compose.yaml build api
docker compose -f docker/docker-compose.yaml up -d api
```

## Verification

### Test Results
Tested sandbox creation with three different Docker Hub images:
- ✅ `nginx:1.25-alpine` - Success
- ✅ `python:3.11-slim` - Success
- ✅ `redis:7.0-alpine` - Success

### API Logs
- **Before fix**: Multiple "Failed to get manifest" errors
- **After fix**: No manifest-related errors in logs

## Impact
- ✅ Sandbox creation from Docker Hub images now works correctly
- ✅ API can successfully retrieve image manifests using authenticated Docker Hub credentials
- ✅ No more "manifest unknown" errors in audit logs
- ✅ Other registry authentication (Basic Auth) remains unchanged
