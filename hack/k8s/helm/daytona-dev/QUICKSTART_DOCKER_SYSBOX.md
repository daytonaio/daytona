# Quick Start: Docker + Sysbox on Daytona Runner

This guide will help you quickly deploy the Daytona Runner with Docker and Sysbox support.

## Prerequisites

- Kubernetes cluster (GKE recommended)
- `kubectl` configured
- `helm` installed
- Nodes with:
  - Linux kernel 5.5+ (`uname -r`)
  - x86_64 architecture
  - Ubuntu/Debian OS
  - At least 4GB RAM per node
  - At least 20GB disk space per node

## Quick Deploy (5 minutes)

### 1. Deploy Runner with Docker + Sysbox

```bash
cd /workspaces/daytona

# Deploy
helm upgrade --install daytona-dev ./helm/daytona-dev \
  --namespace daytona-dev \
  --create-namespace

# Wait for runner to be ready
kubectl wait --for=condition=ready pod \
  -l app=daytona-runner \
  -n daytona-dev \
  --timeout=600s
```

**Note**: First deployment may take 5-10 minutes as Docker and Sysbox are installed on each node.

### 2. Validate Installation

```bash
# Run automated validation
./helm/daytona-dev/SYSBOX_VALIDATION.sh

# Expected output:
# ✓ Pod is running
# ✓ Docker is installed
# ✓ Sysbox is installed
# ✓ Kernel version is compatible
# ✓ Validation completed!
```

### 3. Manual Verification (Optional)

```bash
# Get pod name
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

# Check Docker
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker version

# Check Sysbox
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl status sysbox --no-pager

# Check runtime
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker info | grep -A5 "Runtimes:"
```

## Test Docker-in-Docker (2 minutes)

### Test 1: Basic Docker

```bash
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

# Run hello-world
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker run --rm hello-world

# Should output: "Hello from Docker!"
```

### Test 2: Docker-in-Docker (The Magic!)

```bash
# Create a container with Docker inside (no --privileged needed!)
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker run -d --name test-dind docker:dind

# Wait a few seconds for Docker to start inside
sleep 10

# Run Docker commands INSIDE the container
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker exec test-dind docker run --rm alpine echo "Docker-in-Docker works!"

# Cleanup
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker rm -f test-dind
```

### Test 3: Docker Compose

```bash
# Run a multi-container app
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i bash << 'EOF'
mkdir -p /tmp/compose-test
cd /tmp/compose-test

cat > docker-compose.yml << 'COMPOSE'
version: '3'
services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
  redis:
    image: redis:alpine
COMPOSE

docker compose up -d
docker compose ps
docker compose down
cd /
rm -rf /tmp/compose-test
EOF
```

## View Logs

```bash
# Installer logs (Docker + Sysbox installation)
kubectl logs -n daytona-dev -l app=daytona-runner -c docker-installer --tail=50

# Runner logs
kubectl logs -n daytona-dev -l app=daytona-runner -c runner --tail=50

# Follow logs in real-time
kubectl logs -n daytona-dev -l app=daytona-runner -c docker-installer -f
```

## Check Status

```bash
# Pod status
kubectl get pods -n daytona-dev -l app=daytona-runner -o wide

# Service status
kubectl get svc -n daytona-dev daytona-runner

# Node status
kubectl get nodes -l workload-type=daytona-sandbox-c
```

## Troubleshooting

### Issue: Pod is stuck in Init or Pending

```bash
# Check pod events
kubectl describe pod -n daytona-dev -l app=daytona-runner

# Common causes:
# - No nodes with matching labels
# - Insufficient resources
# - Image pull issues
```

**Fix:**

```bash
# Check node labels
kubectl get nodes --show-labels | grep daytona-sandbox

# If no nodes match, add label:
kubectl label nodes <node-name> workload-type=daytona-sandbox-c
```

### Issue: Docker installation fails

```bash
# View detailed logs
kubectl logs -n daytona-dev <pod-name> -c docker-installer --tail=100

# Common causes:
# - No internet access
# - Insufficient disk space
# - Unsupported OS
```

**Fix:**

```bash
# Check node OS
kubectl get nodes -o wide

# Check disk space
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i df -h
```

### Issue: Sysbox installation fails

```bash
# Check kernel version
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i uname -r

# Must be 5.5 or higher
```

**Fix:**

```bash
# If kernel too old, use GKE with newer version:
gcloud container clusters create my-cluster \
  --release-channel rapid \
  --machine-type n2-standard-4
```

### Issue: Services keep restarting

```bash
# Check systemd status
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i journalctl -u docker -n 50

kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i journalctl -u sysbox -n 50
```

## Configuration Options

### Disable Docker Installer

If you want to go back to containerd:

```bash
# Edit values.yaml
cat >> helm/daytona-dev/values.yaml << 'EOF'
runner:
  dockerInstaller:
    enabled: false
  config:
    containerRuntime: "containerd"
EOF

# Redeploy
helm upgrade daytona-dev ./helm/daytona-dev -n daytona-dev
```

### Change Base Image

```bash
# Use different base image for installer
cat >> helm/daytona-dev/values.yaml << 'EOF'
runner:
  dockerInstaller:
    enabled: true
    image: ubuntu:22.04  # Instead of alpine:latest
EOF

# Redeploy
helm upgrade daytona-dev ./helm/daytona-dev -n daytona-dev
```

### Adjust Resources

```bash
# Increase runner resources if needed
cat >> helm/daytona-dev/values.yaml << 'EOF'
runner:
  resources:
    requests:
      cpu: "1000m"
      memory: "2Gi"
    limits:
      cpu: "4000m"
      memory: "8Gi"
EOF

# Redeploy
helm upgrade daytona-dev ./helm/daytona-dev -n daytona-dev
```

## Monitoring

### Resource Usage

```bash
# Pod resource usage
kubectl top pods -n daytona-dev -l app=daytona-runner

# Node resource usage
kubectl top nodes

# Disk usage
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i df -h /var/lib/docker /var/lib/sysbox
```

### Container Count

```bash
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

# Docker containers
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Image}}"

# Container count
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker ps -q | wc -l
```

### Health Checks

```bash
# Check all services are active
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i bash -c '
    echo "Docker: $(systemctl is-active docker)"
    echo "Sysbox: $(systemctl is-active sysbox)"
  '
```

## Cleanup

### Remove Everything

```bash
# Delete runner
helm uninstall daytona-dev -n daytona-dev

# Delete namespace
kubectl delete namespace daytona-dev

# Optional: Remove Docker/Sysbox from nodes
# (If you want a complete cleanup)
kubectl run cleanup --rm -it --image=alpine --privileged --restart=Never -- sh -c '
  nsenter -t 1 -m -u -n -i bash << EOF
    systemctl stop docker sysbox
    apt-get remove -y docker.io sysbox-ce
    rm -rf /var/lib/docker /var/lib/sysbox
EOF
'
```

### Remove Just Docker/Sysbox (Keep Runner)

```bash
# Disable installer
helm upgrade daytona-dev ./helm/daytona-dev \
  --set runner.dockerInstaller.enabled=false \
  --set runner.config.containerRuntime=containerd \
  -n daytona-dev
```

## Production Checklist

Before deploying to production:

- [ ] Test in development/staging first
- [ ] Verify kernel version on all nodes (5.5+)
- [ ] Ensure sufficient disk space (20GB+ per node)
- [ ] Set up monitoring alerts
- [ ] Configure resource limits appropriately
- [ ] Set up log aggregation
- [ ] Configure network policies
- [ ] Set up backup/disaster recovery
- [ ] Document runbooks for your team
- [ ] Test failover scenarios
- [ ] Review security policies
- [ ] Set up auto-scaling if needed

## Performance Tips

1. **Use SSD storage** for `/var/lib/docker` and `/var/lib/sysbox`
2. **Prune regularly** to free up disk space
3. **Monitor disk I/O** to avoid bottlenecks
4. **Use appropriate node sizes** (4+ CPU, 8GB+ RAM recommended)
5. **Enable log rotation** (already configured)

## Next Steps

- 📖 Read full documentation: `/workspaces/daytona/DOCKER_INSTALLER_README.md`
- 📊 Review changes: `/workspaces/daytona/DOCKER_SYSBOX_CHANGES_SUMMARY.md`
- 🔍 Run validation: `./helm/daytona-dev/SYSBOX_VALIDATION.sh`
- 🧪 Test your workloads
- 📈 Monitor performance
- 🚀 Deploy to production

## Getting Help

If you encounter issues:

1. Run the validation script: `./helm/daytona-dev/SYSBOX_VALIDATION.sh`
2. Check logs: `kubectl logs -n daytona-dev <pod-name> -c docker-installer`
3. Review troubleshooting guide in `DOCKER_INSTALLER_README.md`
4. Check Sysbox GitHub issues: https://github.com/nestybox/sysbox/issues

## Success Indicators

You'll know everything is working when:

✅ Validation script passes all checks
✅ `docker version` works on host
✅ `systemctl status sysbox` shows active
✅ Docker runtime is `sysbox-runc`
✅ Docker-in-Docker works without `--privileged`
✅ Sandboxes can build and run containers
✅ No repeated restarts in logs
✅ Resource usage is within expected ranges

## Quick Reference Commands

```bash
# Get pod name
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

# Check everything
./helm/daytona-dev/SYSBOX_VALIDATION.sh

# View logs
kubectl logs -n daytona-dev $POD -c docker-installer --tail=50

# Check Docker
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker info

# Check Sysbox
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl status sysbox

# Test DinD
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker run --rm docker:dind docker version

# Restart services
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl restart sysbox docker

# Cleanup
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker system prune -af
```

---

**Deployment time**: ~5-10 minutes
**Setup difficulty**: Easy
**Maintenance**: Low (automated monitoring and recovery)
**Production ready**: Yes (with proper testing)
