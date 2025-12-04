# Mock Runner

The Mock Runner is a development/testing variant of the Daytona Runner that simulates sandbox operations without creating individual Docker containers for each sandbox. Instead, it uses **in-memory state management** for sandboxes and a **single shared Docker container** (the "toolbox container") for all toolbox operations.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Mock Runner                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────┐    ┌─────────────────────────────┐ │
│  │   In-Memory Sandboxes   │    │  Shared Toolbox Container   │ │
│  │                         │    │                             │ │
│  │  ┌───────┐ ┌───────┐    │    │  ┌─────────────────────┐    │ │
│  │  │ sbx-1 │ │ sbx-2 │    │    │  │  Docker Container   │    │ │
│  │  │(mock) │ │(mock) │    │    │  │  - Daemon process   │    │ │
│  │  └───────┘ └───────┘    │    │  │  - Port 2280        │    │ │
│  │       ...               │    │  │  - ubuntu:22.04     │    │ │
│  └─────────────────────────┘    │  └──────────┬──────────┘    │ │
│              │                  │             │               │ │
│              │                  │             │               │ │
│              ▼                  │             ▼               │ │
│  ┌─────────────────────────────────────────────────────────┐  │ │
│  │                      API Server                         │  │ │
│  │  - All sandbox CRUD operations use mock state           │  │ │
│  │  - Proxy routes ALL toolbox requests to shared container│  │ │
│  │  - SSH gateway connects to shared container             │  │ │
│  └─────────────────────────────────────────────────────────┘  │ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## How It Works

### Sandbox Operations (Mocked)

All sandbox lifecycle operations are **simulated in memory** - no actual Docker containers are created per sandbox:

| Operation | Behavior |
|-----------|----------|
| `Create`  | Stores sandbox metadata in memory, sets state to "started" |
| `Start`   | Updates state to "started" in memory |
| `Stop`    | Updates state to "stopped" in memory |
| `Destroy` | Removes sandbox from memory, sets state to "destroyed" |
| `Resize`  | Updates resource quotas in memory (no actual enforcement) |

### Image Operations (Mocked)

Image operations are also simulated with in-memory tracking:

| Operation | Behavior |
|-----------|----------|
| `PullImage`   | Adds image name to in-memory registry |
| `PushImage`   | Logs the operation (no-op) |
| `BuildImage`  | Adds image to registry, writes mock build logs |
| `TagImage`    | Creates new tag reference in memory |
| `ImageExists` | Checks in-memory registry |
| `RemoveImage` | Removes from in-memory registry |

### Toolbox Operations (Real Container)

Unlike sandbox operations, **toolbox operations use a real Docker container**:

1. On startup, Mock Runner creates a single "toolbox container" (`mock-runner-toolbox`)
2. The Daytona daemon is copied into and started inside this container
3. **All** proxy requests (regardless of sandbox ID) are routed to this container
4. SSH gateway connections also route to this shared container

This means:

- File operations, process execution, and other toolbox APIs actually work
- All sandboxes share the same filesystem and process space
- State in the toolbox container persists across sandbox operations

## Configuration

The Mock Runner uses environment variables for configuration:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_URL` | Yes | - | Daytona server URL |
| `API_TOKEN` | Yes | - | Authentication token |
| `API_PORT` | No | `8080` | API server port |
| `ENVIRONMENT` | No | - | Set to `development` for debug mode |
| `LOG_LEVEL` | No | `warn` | Logging level (debug, info, warn, error) |
| `LOG_FILE_PATH` | No | - | Path to log file |
| `TOOLBOX_IMAGE` | No | `ubuntu:22.04` | Base image for toolbox container |
| `SSH_GATEWAY_ENABLE` | No | `false` | Enable SSH gateway |
| `CACHE_RETENTION_DAYS` | No | `7` | Days to retain sandbox state in cache |
| `TLS_CERT_FILE` | No | - | TLS certificate file |
| `TLS_KEY_FILE` | No | - | TLS key file |
| `ENABLE_TLS` | No | `false` | Enable HTTPS |

## Building

```bash
# Build using NX
npx nx build mock-runner

# Build directly with Go
cd apps/mock-runner
go build -o mock-runner ./cmd/mock-runner/main.go
```

## Setup & Usage

### Step 1: Start the Mock Runner

First, start the mock runner service with the required environment variables:

```bash
# Required environment variables for the Mock Runner
export SERVER_URL="https://your-daytona-server"
export API_TOKEN="your-api-token"

# Optional configuration
export API_PORT=8080
export LOG_LEVEL="info"
export TOOLBOX_IMAGE="ubuntu:22.04"

# Run the mock runner
./mock-runner
```

The mock runner will:

1. Start an API server on the configured port (default: 8080)
2. Create and start the shared toolbox container
3. Copy the daemon binary into the toolbox container
4. Wait for the daemon to be ready

### Step 2: Register with Daytona API (Automatic)

The Daytona API can automatically register the mock runner on startup. Set these environment variables on the **API server**:

```bash
# Mock Runner registration (on the Daytona API)
export MOCK_RUNNER_DOMAIN="mock-runner.example.com"
export MOCK_RUNNER_API_URL="http://mock-runner:8080"
export MOCK_RUNNER_PROXY_URL="http://mock-runner-proxy:8080"
export MOCK_RUNNER_API_KEY="your-mock-runner-api-key"

# Optional resource configuration
export MOCK_RUNNER_CPU=4
export MOCK_RUNNER_MEMORY=8
export MOCK_RUNNER_DISK=50
export MOCK_RUNNER_GPU=0
export MOCK_RUNNER_CLASS="small"
export MOCK_RUNNER_VERSION="0"
```

When the API starts, it will automatically:

1. Check if a runner with `MOCK_RUNNER_DOMAIN` exists
2. If not, create the runner with the configured settings
3. Wait for the mock runner to be healthy (up to 30 seconds)

### Step 3: Register with Daytona API (Manual)

Alternatively, you can manually register the mock runner using the API:

```bash
curl -X POST "https://your-daytona-server/runners" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "mock-runner.example.com",
    "apiUrl": "http://mock-runner:8080",
    "proxyUrl": "http://mock-runner-proxy:8080",
    "apiKey": "your-mock-runner-api-key",
    "cpu": 4,
    "memoryGiB": 8,
    "diskGiB": 50,
    "gpu": 0,
    "region": "us",
    "class": "small"
  }'
```

### Docker Compose Example

Here's a complete example using Docker Compose:

```yaml
version: '3.8'

services:
  mock-runner:
    build:
      context: ../..
      dockerfile: apps/mock-runner/Dockerfile
    environment:
      - SERVER_URL=http://api:3000
      - API_TOKEN=your-api-token
      - API_PORT=8080
      - LOG_LEVEL=info
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # Required for toolbox container
    depends_on:
      - api

  api:
    # Your Daytona API service
    environment:
      # ... other API config ...
      - MOCK_RUNNER_DOMAIN=mock-runner
      - MOCK_RUNNER_API_URL=http://mock-runner:8080
      - MOCK_RUNNER_PROXY_URL=http://mock-runner:8080
      - MOCK_RUNNER_API_KEY=your-api-token
```

### Verifying the Setup

Check if the mock runner is healthy:

```bash
# Health check
curl http://localhost:8080/

# Expected response:
# {"status":"ok","version":"v0.0.0-dev","mode":"mock"}
```

Check runner info:

```bash
curl -H "Authorization: Bearer your-api-token" http://localhost:8080/info
```

### Using with the Daytona SDK

Once registered, you can use the mock runner with any Daytona SDK. Sandboxes created on the mock runner will:

- Appear to work normally through the API
- Share the same toolbox container for all operations
- Have their state tracked in memory (not persisted)

```python
from daytona import Daytona

# The SDK works the same way - just target the mock runner
daytona = Daytona()

# Create a sandbox (mocked - no real container created)
sandbox = daytona.create()

# Execute commands (routes to shared toolbox container)
result = sandbox.process.exec("echo 'Hello from mock runner!'")
print(result.output)

# All sandboxes share the same toolbox container
sandbox2 = daytona.create()
# sandbox2's toolbox operations go to the same container as sandbox1
```

## Use Cases

The Mock Runner is ideal for:

1. **Development**: Test API integrations without heavy Docker overhead
2. **CI/CD**: Run integration tests without needing Docker-in-Docker for each sandbox
3. **Resource-constrained environments**: Run multiple "sandboxes" with minimal resources
4. **Debugging**: Easier to inspect shared state and logs

## Limitations

Since sandboxes are mocked:

- **No isolation**: All sandboxes share the same toolbox container
- **No resource limits**: CPU/memory quotas are tracked but not enforced
- **No network isolation**: Network settings are recorded but not applied
- **Shared filesystem**: All sandboxes see the same files in the toolbox container
- **Single architecture**: Only supports amd64/linux

## API Compatibility

The Mock Runner exposes the **same API** as the regular Runner:

- `POST /sandboxes` - Create sandbox
- `GET /sandboxes/{id}` - Get sandbox info
- `POST /sandboxes/{id}/start` - Start sandbox
- `POST /sandboxes/{id}/stop` - Stop sandbox
- `POST /sandboxes/{id}/destroy` - Destroy sandbox
- `POST /sandboxes/{id}/resize` - Resize sandbox
- `POST /sandboxes/{id}/backup` - Create backup (mocked)
- `ANY /sandboxes/{id}/toolbox/*` - Proxy to toolbox container
- `POST /snapshots/pull` - Pull snapshot (mocked)
- `POST /snapshots/build` - Build snapshot (mocked)
- `GET /snapshots/exists` - Check if snapshot exists
- `GET /snapshots/info` - Get snapshot info
- `GET /info` - Runner info with metrics
- `GET /metrics` - Prometheus metrics
- `GET /` - Health check

## Differences from Regular Runner

| Aspect | Regular Runner | Mock Runner |
|--------|---------------|-------------|
| Container per sandbox | Yes | No (in-memory) |
| Toolbox container | Per sandbox | Shared (one for all) |
| Resource limits | Enforced | Tracked only |
| Network rules | Applied | Recorded only |
| Image operations | Real Docker | In-memory tracking |
| Daemon | Per container | Single shared instance |
| Docker dependency | Full | Minimal (toolbox only) |

## Troubleshooting

### Mock Runner won't start

**Problem**: Mock runner fails to start with Docker errors.

**Solution**: Ensure Docker is running and the mock runner has access to the Docker socket:

```bash
# Check Docker is running
docker ps

# If using Docker-in-Docker, mount the socket
docker run -v /var/run/docker.sock:/var/run/docker.sock mock-runner
```

### Toolbox container not starting

**Problem**: Mock runner starts but toolbox operations fail.

**Solution**: Check if the toolbox container exists and is running:

```bash
# Check for the toolbox container
docker ps -a | grep mock-runner-toolbox

# Check container logs
docker logs mock-runner-toolbox

# Manually remove and let mock-runner recreate it
docker rm -f mock-runner-toolbox
```

### Daemon not responding

**Problem**: Toolbox operations timeout waiting for daemon.

**Solution**: The daemon binary might not be available. Check:

```bash
# Ensure daemon binary is built
npx nx build daemon

# Check if it's copied to the right location
ls -la apps/mock-runner/pkg/daemon/static/daemon-amd64
```

### API registration fails

**Problem**: Daytona API can't connect to mock runner.

**Solution**: Verify network connectivity and configuration:

```bash
# Test mock runner health from API server
curl http://mock-runner:8080/

# Check environment variables on API
echo $MOCK_RUNNER_API_URL
echo $MOCK_RUNNER_DOMAIN
```

### State not persisting

**Problem**: Sandbox state is lost after restart.

**Expected behavior**: This is by design. The mock runner stores all sandbox state in memory. When the mock runner restarts:

- All sandbox metadata is cleared
- The toolbox container may persist (unless manually removed)
- Images tracked in memory are cleared

For persistent state, use the regular Runner.

## Environment Variables Reference

### Mock Runner Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_URL` | Yes | - | Daytona server URL |
| `API_TOKEN` | Yes | - | Authentication token for API requests |
| `API_PORT` | No | `8080` | Port for the mock runner API |
| `ENVIRONMENT` | No | - | Set to `development` for debug mode |
| `LOG_LEVEL` | No | `warn` | Logging level (debug, info, warn, error) |
| `LOG_FILE_PATH` | No | - | Path to log file |
| `TOOLBOX_IMAGE` | No | `ubuntu:22.04` | Docker image for toolbox container |
| `SSH_GATEWAY_ENABLE` | No | `false` | Enable SSH gateway routing |
| `CACHE_RETENTION_DAYS` | No | `7` | Days to retain sandbox state |
| `TLS_CERT_FILE` | No | - | TLS certificate file path |
| `TLS_KEY_FILE` | No | - | TLS private key file path |
| `ENABLE_TLS` | No | `false` | Enable HTTPS |

### Daytona API Environment Variables (for auto-registration)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MOCK_RUNNER_DOMAIN` | No | - | Domain identifier (enables auto-registration) |
| `MOCK_RUNNER_API_URL` | Yes* | - | Mock runner API URL |
| `MOCK_RUNNER_PROXY_URL` | Yes* | - | Mock runner proxy URL |
| `MOCK_RUNNER_API_KEY` | Yes* | - | API key for mock runner auth |
| `MOCK_RUNNER_CPU` | No | `4` | CPU allocation |
| `MOCK_RUNNER_MEMORY` | No | `8` | Memory in GiB |
| `MOCK_RUNNER_DISK` | No | `50` | Disk in GiB |
| `MOCK_RUNNER_GPU` | No | `0` | GPU count |
| `MOCK_RUNNER_GPU_TYPE` | No | - | GPU type |
| `MOCK_RUNNER_CLASS` | No | - | Sandbox class |
| `MOCK_RUNNER_VERSION` | No | `0` | Runner version |

*Required only if `MOCK_RUNNER_DOMAIN` is set.
