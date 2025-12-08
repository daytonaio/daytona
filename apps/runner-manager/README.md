# Daytona Runner Manager

A Kubernetes-native service for managing Daytona runner infrastructure through dynamic node provisioning using placeholder pods and cluster autoscaling.

## Overview

The Runner Manager provides an API to dynamically add/remove runner nodes in a Kubernetes cluster by creating placeholder pods that trigger the cluster autoscaler. When runners are requested, it:

1. Creates placeholder pods with pod anti-affinity (one per node)
2. Placeholder pods trigger cluster autoscaler to provision new nodes
3. Runner DaemonSet automatically schedules on new nodes
4. Background jobs monitor pod scheduling and track node information
5. API exposes node internal IPs for cluster communication

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                          │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Cluster Autoscaler                                      │   │
│  │  (monitors pending pods, provisions nodes)               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Placeholder Pods (anti-affinity = 1 per node)          │   │
│  │  - node-placeholder-abc123 (triggers node creation)      │   │
│  │  - node-placeholder-def456 (triggers node creation)      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Runner DaemonSet                                         │   │
│  │  (automatically schedules on new nodes)                   │   │
│  │  - daytona-runner-xyz (on node 1)                        │   │
│  │  - daytona-runner-abc (on node 2)                        │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                           │
                           ▼
                  ┌──────────────────┐
                  │ Runner Manager   │
                  │ - Creates pods   │
                  │ - Tracks status  │
                  │ - Returns IPs    │
                  └──────────────────┘
```

## Project Structure

```
runner-manager/
├── cmd/
│   └── runner-manager/
│       └── main.go              # Application entry point
├── pkg/
│   ├── api/
│   │   ├── controllers/
│   │   │   └── runner.go        # API endpoints (Add, Remove, List, Get)
│   │   ├── dto/
│   │   │   └── runner.dto.go    # Request/response DTOs
│   │   └── middleware/
│   │       └── auth.go          # Bearer token authentication
│   └── provider/
│       ├── provider.go          # Provider interface
│       ├── manager.go           # Provider singleton manager
│       ├── types/
│       │   └── types.go         # Shared types (RunnerInfo, AddRunnerResponse)
│       ├── k8s/
│       │   └── k8s.go           # Kubernetes provider implementation
│       └── aws/
│           └── aws.go           # AWS provider (placeholder)
├── internal/
│   └── buildinfo.go             # Build version info
├── .env                         # Local environment configuration
├── .env.example                 # Environment configuration template
├── go.mod                       # Go module dependencies
└── README.md                    # This file
```

## Configuration

### Environment Variables

All configuration is done via environment variables or `.env` file:

| Variable | Default | Description |
|----------|---------|-------------|
| `API_PORT` | `3010` | HTTP API server port |
| `API_TOKEN` | _(required)_ | Bearer token for API authentication |
| `PROVIDER_TYPE` | `kubernetes` | Provider type (`kubernetes`, `k8s`, or `aws`) |
| `PROVIDER_NAMESPACE` | `default` | Kubernetes namespace for placeholder pods |
| `POD_WAIT_TIMEOUT` | `600` | Timeout in seconds for pod scheduling (10 minutes) |
| `KUBECONFIG_PATH` | _(optional)_ | Path to kubeconfig for local development |
| `LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |

### Generate API Token

For production, generate a secure random token:

```bash
openssl rand -base64 32
```

## Development Setup

### Prerequisites

- Go 1.25.4 or later
- Access to a Kubernetes cluster
- Kubeconfig file for cluster access
- (Optional) Telepresence for local development

### Local Development with Kubeconfig

1. **Copy environment template:**

```bash
cp .env.example .env
```

2. **Configure `.env` for development:**

```bash
# API Configuration
API_PORT=3010
API_TOKEN=dev-test-token-change-in-production

# Provider Configuration
PROVIDER_TYPE=kubernetes
PROVIDER_NAMESPACE=default

# Local Development - Use kubeconfig
KUBECONFIG_PATH=/path/to/your/kubeconfig.yaml

# Logging
LOG_LEVEL=debug

# Pod scheduling timeout (10 minutes)
POD_WAIT_TIMEOUT=600
```

3. **Build the application:**

```bash
go build -o runner-manager ./cmd/runner-manager
```

4. **Run locally:**

```bash
./runner-manager
```

5. **Test API:**

```bash
# Add a runner
curl -X POST http://localhost:3010/runners/add \
  -H "Authorization: Bearer dev-test-token-change-in-production" \
  -H "Content-Type: application/json" \
  -d '{"instances": 1}'

# List runners
curl http://localhost:3010/runners \
  -H "Authorization: Bearer dev-test-token-change-in-production"

# Get specific runner
curl http://localhost:3010/runners/{runner-id} \
  -H "Authorization: Bearer dev-test-token-change-in-production"
```

### Local Development with Telepresence

Telepresence allows running locally while connected to the cluster network:

1. **Connect Telepresence:**

```bash
export KUBECONFIG=/path/to/kubeconfig.yaml
telepresence connect --namespace your-namespace
```

2. **Run runner-manager:**

```bash
KUBECONFIG_PATH=/path/to/kubeconfig.yaml \
API_TOKEN=dev-token \
LOG_LEVEL=debug \
./runner-manager
```

Your local instance can now access cluster services and create pods directly.

## Production Deployment

### Kubernetes Deployment

Runner-manager is designed to run inside the Kubernetes cluster using in-cluster authentication.

1. **Create deployment YAML:**

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runner-manager
  namespace: your-namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: runner-manager
  namespace: your-namespace
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "create", "delete", "watch"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: runner-manager
  namespace: your-namespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: runner-manager
subjects:
- kind: ServiceAccount
  name: runner-manager
  namespace: your-namespace
---
apiVersion: v1
kind: Secret
metadata:
  name: runner-manager-secrets
  namespace: your-namespace
type: Opaque
stringData:
  API_TOKEN: "your-secure-random-token-here"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: runner-manager
  namespace: your-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: runner-manager
  template:
    metadata:
      labels:
        app: runner-manager
    spec:
      serviceAccountName: runner-manager
      containers:
      - name: runner-manager
        image: your-registry/runner-manager:latest
        ports:
        - containerPort: 3010
          name: http
        env:
        - name: API_PORT
          value: "3010"
        - name: API_TOKEN
          valueFrom:
            secretKeyRef:
              name: runner-manager-secrets
              key: API_TOKEN
        - name: PROVIDER_TYPE
          value: "kubernetes"
        - name: PROVIDER_NAMESPACE
          value: "your-namespace"
        - name: POD_WAIT_TIMEOUT
          value: "600"
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 3010
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 3010
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: runner-manager
  namespace: your-namespace
spec:
  selector:
    app: runner-manager
  ports:
  - port: 3010
    targetPort: 3010
    protocol: TCP
  type: ClusterIP
```

2. **Deploy:**

```bash
kubectl apply -f runner-manager-deployment.yaml
```

### Production Configuration Notes

- **No `KUBECONFIG_PATH`**: In production, omit this variable to use in-cluster config
- **Secure API_TOKEN**: Use a strong random token stored in Kubernetes Secret
- **Resource Limits**: Adjust based on expected load
- **Logging**: Set `LOG_LEVEL=info` or `warn` for production
- **RBAC**: Ensure ServiceAccount has necessary permissions

## Cluster Requirements

### Node Pool Configuration

The runner-manager expects a node pool configured for autoscaling:

```bash
# GKE Example
gcloud container node-pools create runner-pool \
  --cluster=your-cluster \
  --zone=your-zone \
  --machine-type=n2-standard-4 \
  --node-labels=daytona-sandbox-c=true \
  --node-taints=sandbox=true:NoSchedule \
  --enable-autoscaling \
  --min-nodes=0 \
  --max-nodes=10
```

**Requirements:**

- Node label: `daytona-sandbox-c=true`
- Node taint: `sandbox=true:NoSchedule`
- Autoscaling enabled with min/max bounds
- Cluster autoscaler installed and running

## API Reference

### Endpoints

#### Health Check

```
GET /health
```

Returns service health status (no authentication required).

#### Root

```
GET /
```

Returns service information and version (no authentication required).

#### Add Runners

```
POST /runners/add
Authorization: Bearer {token}
Content-Type: application/json

{
  "instances": 2
}
```

**Response:**

```json
{
  "job_id": "uuid-string",
  "pod_names": ["node-placeholder-abc", "node-placeholder-def"],
  "message": "Runner provisioning started",
  "instances": 2,
  "provider": "kubernetes"
}
```

Creates placeholder pods that trigger node provisioning. Returns immediately with job ID for tracking.

#### List Runners

```
GET /runners
Authorization: Bearer {token}
```

**Response:**

```json
[
  {
    "id": "node-placeholder-abc123",
    "status": "ready",
    "metadata": {
      "node_name": "gke-cluster-node-xyz",
      "node_internal_ip": "10.128.0.15",
      "placeholder_pod": "node-placeholder-abc123",
      "runner_pod": "daytona-runner-def456"
    }
  }
]
```

Returns all runners with their current status and node information.

#### Get Runner

```
GET /runners/{runnerId}
Authorization: Bearer {token}
```

Returns information for a specific runner (runnerId is the placeholder pod name).

#### Remove Runners

```
POST /runners/remove
Authorization: Bearer {token}
Content-Type: application/json

{
  "instances": 1
}
```

_(Not yet implemented)_

### Runner Status Values

- **`pending`**: Placeholder pod created but not scheduled to a node
- **`provisioning`**: Pod scheduled to node, waiting for runner DaemonSet
- **`ready`**: Runner DaemonSet pod is running on the node
- **`failed`**: Pod failed to schedule or run

## Background Job Monitoring

When runners are added, a background goroutine monitors pod scheduling:

1. Polls pods every 5 seconds
2. Detects when pods are scheduled to nodes
3. Logs scheduling events
4. Updates job status (pending → running → completed/timeout)
5. Timeout configurable via `POD_WAIT_TIMEOUT` (default 10 minutes)

**Logs Example:**

```
INFO Background job abc-123 started: waiting for 2 pods to be scheduled
INFO Pod node-placeholder-xyz scheduled to node gke-cluster-node-abc
INFO Pod node-placeholder-xyz successfully scheduled to node
INFO Background job abc-123 completed successfully
```

## Building

### Local Build

```bash
go build -o runner-manager ./cmd/runner-manager
```

### Docker Build

```bash
docker build -t runner-manager:latest .
```

### Multi-arch Build

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t your-registry/runner-manager:latest \
  --push .
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Test

```bash
# Start runner-manager
KUBECONFIG_PATH=/path/to/kubeconfig.yaml \
API_TOKEN=test-token \
./runner-manager

# In another terminal, test adding a runner
curl -X POST http://localhost:3010/runners/add \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"instances": 1}'

# Verify pod creation
kubectl get pods -l app=node-placeholder

# Verify autoscaling triggered (if enabled)
kubectl describe pod <placeholder-pod-name>
# Look for: "TriggeredScaleUp" event

# Monitor runner-manager logs for scheduling
tail -f /path/to/logs
```

## Troubleshooting

### Issue: Pods not scheduling

**Check node labels:**

```bash
kubectl get nodes --show-labels | grep daytona-sandbox-c
```

If no nodes match, add the label:

```bash
kubectl label nodes <node-name> daytona-sandbox-c=true
```

### Issue: Autoscaling not triggering

**Verify autoscaling is enabled:**

```bash
gcloud container node-pools describe <pool-name> \
  --cluster=<cluster> --zone=<zone>
```

Look for `autoscaling.enabled: true`

**Check cluster autoscaler logs:**

```bash
kubectl logs -n kube-system -l app=cluster-autoscaler
```

### Issue: Authentication failures

Ensure API_TOKEN is set and matches the token in requests:

```bash
echo $API_TOKEN
```

### Issue: In-cluster config fails

When running locally without kubeconfig:

```
failed to create in-cluster config and no KUBECONFIG found
```

**Solution:** Set `KUBECONFIG_PATH` environment variable for local development.

### Issue: Background job timeout

Pods not scheduling within timeout (default 10 minutes):

```
WARN Pod node-placeholder-xyz not scheduled within timeout
WARN Background job abc-123 timed out
```

**Solutions:**

- Verify autoscaling is enabled and configured correctly
- Increase `POD_WAIT_TIMEOUT` if nodes take longer to provision
- Check if cluster has reached max nodes limit
- Verify node pool has available capacity

## Security Considerations

1. **API Token**: Always use a strong random token in production
2. **RBAC**: Grant minimal permissions (only pods in specific namespace)
3. **Network Policies**: Restrict access to runner-manager API
4. **TLS**: Use ingress/service mesh for HTTPS in production
5. **Secrets**: Store sensitive config in Kubernetes Secrets, never in code

## Performance

- **Async Operations**: Pod creation is non-blocking, returns immediately
- **Background Monitoring**: Minimal overhead (5-second polling interval)
- **Scalability**: Can manage hundreds of runners concurrently
- **Memory**: ~30-50MB base + ~1MB per tracked job

## Contributing

### Code Style

- Follow Go conventions and `golint` standards
- Add comments for exported functions
- Include error handling for all operations

### Adding New Providers

To add a new cloud provider:

1. Create `pkg/provider/yourprovider/yourprovider.go`
2. Implement `IRunnerProvider` interface
3. Add provider to `manager.go` switch statement
4. Update configuration documentation

## License

AGPL-3.0 - See COPYRIGHT file for details.

## Support

For issues and questions:

- GitHub Issues: https://github.com/daytonaio/daytona
- Documentation: https://daytona.io/docs
