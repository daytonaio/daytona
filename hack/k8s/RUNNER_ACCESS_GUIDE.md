# Accessing Daytona Runner from Local Environment

This guide provides multiple methods to access the Daytona runner DaemonSet from your local environment on localhost:8080.

## Current Setup

- **Runner Pod**: `daytona-runner-kw6nq`
- **Node**: `gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g`
- **Namespace**: `daytona-dev`
- **API Port**: 8080
- **SSH Gateway Port**: 2220

## Method 1: Direct Port Forward (Recommended) ⭐

This is the simplest method using kubectl port-forward to the specific pod.

### Quick Start

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

# Get the pod name on the specific node
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o json | \
  jq -r '.items[] | select(.spec.nodeName=="gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g") | .metadata.name')

# Port forward to localhost
kubectl port-forward -n daytona-dev pod/$POD 8080:8080 2220:2220
```

### Access

- **Runner API**: http://localhost:8080
- **SSH Gateway**: localhost:2220
- **Health Check**: `curl http://localhost:8080`

### Run in Background

```bash
# Run in background
kubectl port-forward -n daytona-dev pod/$POD 8080:8080 2220:2220 &

# Save PID for later cleanup
echo $! > /tmp/kubectl-port-forward.pid

# To stop later
kill $(cat /tmp/kubectl-port-forward.pid)
```

### Using the Helper Script

```bash
# Interactive script
./TELEPRESENCE_SETUP.sh

# Or run directly
bash TELEPRESENCE_SETUP.sh
```

## Method 2: Port Forward to Service (Round-Robin)

Forward to the service (will load-balance between all runner pods).

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

# Forward to the service
kubectl port-forward -n daytona-dev svc/daytona-runner 8080:8080 2220:2220
```

**Note**: This may connect to a different pod on each request due to load balancing.

## Method 3: Telepresence Intercept

Telepresence provides more advanced features like intercepting traffic and environment variable injection.

### Install Telepresence

```bash
# macOS
brew install datawire/blackbird/telepresence

# Linux
sudo curl -fL https://app.getambassador.io/download/tel2oss/releases/download/v2.20.2/telepresence-linux-amd64 \
  -o /usr/local/bin/telepresence
sudo chmod +x /usr/local/bin/telepresence

# Verify installation
telepresence version
```

### Connect to Cluster

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

# Connect telepresence to the cluster
telepresence connect --namespace daytona-dev

# Check connection
telepresence status
```

### Option A: Intercept Headless Service (Specific Pod)

```bash
# Intercept the headless service with specific pod
telepresence intercept daytona-runner-headless \
  --port 8080:8080 \
  --namespace daytona-dev \
  --env-file /tmp/runner-env.env
```

### Option B: Intercept with Pod Selection

```bash
# For DaemonSet, intercept by pod selector
telepresence intercept daytona-runner-headless \
  --port 8080:8080 \
  --namespace daytona-dev \
  --workload daytona-runner \
  --local-port 8080:8080
```

### Access via Telepresence

Once intercepted:

- **Runner API**: http://localhost:8080
- **Cluster DNS**: All K8s services accessible by name
- **Environment**: Get runner's environment variables from `/tmp/runner-env.env`

### Cleanup Telepresence

```bash
# Leave intercept
telepresence leave daytona-runner-headless

# Disconnect
telepresence quit
```

## Method 4: Direct LoadBalancer Access

The runner service has a LoadBalancer with external IP.

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

# Get external IP
EXTERNAL_IP=$(kubectl get svc -n daytona-dev daytona-runner -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "External IP: $EXTERNAL_IP"

# Access directly
curl http://$EXTERNAL_IP:8080
```

**Current External IP**: `34.61.117.88`

**Access**:

- Runner API: http://34.61.117.88:8080
- SSH Gateway: 34.61.117.88:2220

**Note**: This accesses all runner pods (round-robin), not just the specific node.

## Method 5: Node Port Access

Access via node's external IP and NodePort.

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

# Get node external IP
NODE_IP=$(kubectl get node gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g \
  -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}')

# Get NodePort
NODE_PORT=$(kubectl get svc -n daytona-dev daytona-runner \
  -o jsonpath='{.spec.ports[?(@.name=="api")].nodePort}')

echo "Node IP: $NODE_IP"
echo "Node Port: $NODE_PORT"

# Access via NodePort
curl http://$NODE_IP:$NODE_PORT
```

## Method 6: Local Proxy with Telepresence

Access all cluster services via local proxy.

```bash
# Connect telepresence
telepresence connect --namespace daytona-dev

# Access service by K8s DNS name
curl http://daytona-runner.daytona-dev.svc.cluster.local:8080

# Or use service IP
curl http://34.118.226.10:8080
```

## Comparison Table

| Method | Specific Pod | Simplicity | Background | Traffic Intercept | Cluster DNS |
|--------|--------------|------------|------------|-------------------|-------------|
| Port Forward (Pod) | ✅ | ⭐⭐⭐⭐⭐ | ✅ | ❌ | ❌ |
| Port Forward (Svc) | ❌ | ⭐⭐⭐⭐⭐ | ✅ | ❌ | ❌ |
| Telepresence | ✅ | ⭐⭐⭐ | ✅ | ✅ | ✅ |
| LoadBalancer | ❌ | ⭐⭐⭐⭐ | ✅ | ❌ | ❌ |
| NodePort | ✅ | ⭐⭐⭐ | ✅ | ❌ | ❌ |

## Quick Commands

### Get Specific Pod on Node

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

kubectl get pod -n daytona-dev -l app=daytona-runner -o wide | \
  grep "gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g"
```

### Port Forward (One-liner)

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml && \
  kubectl port-forward -n daytona-dev \
  $(kubectl get pod -n daytona-dev -l app=daytona-runner -o json | \
    jq -r '.items[] | select(.spec.nodeName=="gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g") | .metadata.name') \
  8080:8080 2220:2220
```

### Check Runner Health

```bash
# Via port-forward
curl http://localhost:8080

# Via LoadBalancer
curl http://34.61.117.88:8080

# Via telepresence
curl http://daytona-runner.daytona-dev:8080
```

## Troubleshooting

### Port Already in Use

```bash
# Find what's using port 8080
lsof -i :8080

# Kill the process
kill -9 $(lsof -t -i:8080)

# Or use different local port
kubectl port-forward -n daytona-dev pod/$POD 8081:8080
```

### Connection Refused

```bash
# Check pod is running
kubectl get pod -n daytona-dev -l app=daytona-runner

# Check pod logs
kubectl logs -n daytona-dev $POD -c runner --tail=20

# Check service
kubectl get svc -n daytona-dev daytona-runner

# Check endpoints
kubectl get endpoints -n daytona-dev daytona-runner
```

### Telepresence Issues

```bash
# Check status
telepresence status

# Reset connection
telepresence quit
telepresence connect --namespace daytona-dev

# Check logs
telepresence loglevel debug
telepresence status
```

### Can't Find Pod on Node

```bash
# List all runner pods with nodes
kubectl get pod -n daytona-dev -l app=daytona-runner -o wide

# Check if node exists
kubectl get nodes | grep ubuntu-sandbox

# Check DaemonSet status
kubectl get daemonset -n daytona-dev daytona-runner
```

## Environment Variables from Runner

When using telepresence intercept with `--env-file`, you get the runner's environment:

```bash
# After intercept with --env-file /tmp/runner-env.env
cat /tmp/runner-env.env

# Use in your local app
source /tmp/runner-env.env
env | grep DAYTONA
```

## Recommended Workflow

### For Quick Testing

```bash
# Quick port-forward for testing
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o json | \
  jq -r '.items[] | select(.spec.nodeName=="gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g") | .metadata.name')

kubectl port-forward -n daytona-dev pod/$POD 8080:8080 &

# Test
curl http://localhost:8080

# Cleanup
kill %1
```

### For Development

```bash
# Connect telepresence for full cluster access
telepresence connect --namespace daytona-dev

# Your app can now:
# - Access http://daytona-runner.daytona-dev:8080
# - Access any K8s service by DNS
# - Use cluster DNS resolution

# When done
telepresence quit
```

### For Production Testing

```bash
# Use LoadBalancer for external access
EXTERNAL_IP=$(kubectl get svc -n daytona-dev daytona-runner -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl http://$EXTERNAL_IP:8080
```

## Summary

**Best Option for Your Use Case**:

Since you want to access the **specific DaemonSet pod** on node `gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g` from **localhost:8080**:

### 🎯 Recommended Command

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml

# One command to rule them all
kubectl port-forward -n daytona-dev \
  $(kubectl get pod -n daytona-dev -l app=daytona-runner -o json | \
    jq -r '.items[] | select(.spec.nodeName=="gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g") | .metadata.name') \
  8080:8080 2220:2220
```

Then access:

- **Runner API**: http://localhost:8080
- **SSH Gateway**: localhost:2220

This gives you direct access to the specific pod on the target node! 🚀
