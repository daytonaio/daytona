# Docker + Sysbox Installer for Daytona Runner DaemonSet

## Overview

This implementation adds a sidecar container to the Daytona Runner DaemonSet that automatically installs Docker and Sysbox on GKE nodes and configures the runner to use Docker with Sysbox runtime instead of containerd.

### What is Sysbox?

Sysbox is an enhanced container runtime developed by Nestybox that provides:

- **True isolation**: Containers run with enhanced security and isolation
- **Systemd support**: Run systemd inside containers
- **Docker-in-Docker**: Securely run Docker inside containers without privileged mode
- **Kubernetes-in-Docker**: Run K8s clusters inside containers (kind, k3s)
- **Better resource control**: More granular control over container resources

### Key Benefits

| Category | Benefit |
|----------|---------|
| **Security** | No privileged containers needed for Docker-in-Docker, user namespace isolation, reduced attack surface |
| **Functionality** | Full Docker API support, Docker Compose works natively, systemd inside containers |
| **Storage** | XFS filesystem with project quota support for per-container filesystem size limits |
| **Developer Experience** | Familiar Docker commands, standard workflows, extensive tooling ecosystem |
| **Automation** | Automatic installation, recovery on failure, health monitoring |

## Architecture

```
┌─────────────────────────────────────────────────┐
│           GKE Node (Ubuntu)                     │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │  Kubernetes (containerd)                 │  │
│  │  - System Pods                           │  │
│  │  - Runner DaemonSet Pod                  │  │
│  │    ├─ docker-installer (sidecar)         │  │
│  │    └─ runner (main)                      │  │
│  └──────────────────────────────────────────┘  │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │  Docker Daemon + Sysbox Runtime          │  │
│  │  - Daytona Sandboxes                     │  │
│  │  - User Containers                       │  │
│  │  - Docker-in-Docker Support              │  │
│  └──────────────────────────────────────────┘  │
│                                                 │
│  /var/run/docker.sock ←─ Runner mounts this    │
│  /run/containerd/containerd.sock ←─ K8s uses   │
└─────────────────────────────────────────────────┘
```

## System Requirements

### Kernel Requirements

- Linux kernel 5.5 or later (5.12+ recommended)
- Kernel modules: `configfs`, `fuse`, `overlayfs`, `vxlan`
- Kernel features: user namespaces, cgroup v2 (optional but recommended)

### OS Requirements

- Ubuntu 18.04, 20.04, 22.04, 24.04
- Debian 10, 11, 12
- CentOS 8, Rocky Linux 8, AlmaLinux 8
- Fedora (recent versions)

### Node Requirements

- x86_64 architecture (ARM support limited)
- Minimum 2GB RAM (4GB+ recommended)
- Root access (handled by privileged DaemonSet)
- Sufficient disk space (1-2GB overhead)

### GKE Specific

- GKE uses Ubuntu-based node images (compatible)
- Kernel version 5.4+ on most GKE versions
- Ensure node pools have sufficient resources

---

## What Gets Installed

### Docker

| Item | Location |
|------|----------|
| Package | `docker.io` (Ubuntu/Debian) or `docker` (RHEL/CentOS) |
| Socket | `/var/run/docker.sock` |
| Storage | `/var/lib/docker` |
| Service | `docker.service` (systemd) |
| Config | `/etc/docker/daemon.json` |

### Sysbox

| Item | Location |
|------|----------|
| Package | `sysbox-ce` (Community Edition) |
| Binary | `/usr/bin/sysbox-runc` |
| Storage | `/var/lib/sysbox` |
| Service | `sysbox.service` (systemd) |
| Config | `/etc/sysbox/sysbox.conf` |
| Repository | Nestybox official repos |

### Docker Daemon Configuration

The installer creates `/etc/docker/daemon.json`:

```json
{
  "runtimes": {
    "sysbox-runc": {
      "path": "/usr/bin/sysbox-runc"
    }
  },
  "default-runtime": "sysbox-runc",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2"
}
```

This configuration:

- Registers Sysbox as an available runtime
- Sets Sysbox as the default runtime for all containers
- Configures log rotation to prevent disk exhaustion
- Uses overlay2 storage driver for better performance

### XFS Filesystem Storage

XFS filesystem is configured for Docker storage to enable filesystem quota support for limiting container filesystem sizes.

| Item | Location |
|------|----------|
| Loopback File | `/var/lib/docker-xfs.img` (50GB default) |
| Mount Point | `/var/lib/docker` |
| Mount Options | `loop,prjquota` |
| Backup Location | `/var/lib/docker.backup.*` |

**Current Status:**

```bash
=== XFS Mount ===
Filesystem: /dev/loop0
Type: xfs
Size: 50G
Used: 1.0G
Available: 49G
Mount Point: /var/lib/docker

=== Storage Driver ===
Storage Driver: overlay2
Backing Filesystem: xfs
Supports d_type: true
```

---

## Files Modified

### 1. `helm/daytona-dev/templates/runner.yaml`

**Changes:**

- ✅ Added `docker-installer` sidecar container
- ✅ Conditional volume mounts for Docker vs containerd socket
- ✅ Added host root filesystem volume mount for installer
- ✅ Conditional Docker/containerd socket volumes

**Key Features of Docker Installer Sidecar:**

```yaml
- name: docker-installer
  image: alpine:latest
  securityContext:
    privileged: true
  volumeMounts:
    - name: host-root
      mountPath: /host
      mountPropagation: Bidirectional
```

### 2. `helm/daytona-dev/values.yaml`

**Changes:**

```yaml
runner:
  # NEW: Docker installer configuration
  dockerInstaller:
    enabled: true
    image: alpine:latest
    xfsStorageSize: "50G"  # Configurable XFS storage size
  
  config:
    # CHANGED: From "containerd" to "docker"
    containerRuntime: "docker"
```

---

## How It Works

### Installation Process

1. **Pod Startup**: When the runner DaemonSet pod starts, the docker-installer sidecar launches
2. **Host Access**: Uses `nsenter -t 1 -m -u -n -i` to enter the host's namespaces
3. **Check Existing**: Checks if Docker and Sysbox are already installed
4. **Install Docker**:
   - For Debian/Ubuntu (most GKE nodes): `apt-get install -y docker.io`
   - For RHEL/CentOS: `yum install -y docker`
5. **Install Sysbox**:
   - Adds Nestybox package repository
   - Stops Docker temporarily
   - Installs `sysbox-ce` (Community Edition)
   - For Ubuntu: Uses official Nestybox Ubuntu repository
   - For Debian: Uses official Nestybox Debian repository
   - For RHEL/CentOS: Uses official Nestybox RPM repository
6. **Configure Docker**: Creates `/etc/docker/daemon.json` with Sysbox as default runtime
7. **Enable Services**: Ensures Docker and Sysbox start on node boot
8. **Start Services**: Starts Sysbox first, then Docker daemon
9. **Verify**: Checks Docker and Sysbox are running correctly
10. **Monitor**: Continuously monitors both Docker and Sysbox service health every 60 seconds

### Runner Configuration

- Runner container mounts `/var/run/docker.sock` from the host
- `CONTAINER_RUNTIME` environment variable set to "docker"
- Runner uses Docker API to manage containers instead of containerd
- Docker configured to use Sysbox runtime by default

### Coexistence with Containerd

- Docker and Sysbox install alongside containerd without conflicts
- Docker uses its own daemon and storage (`/var/lib/docker`)
- Sysbox uses its own storage (`/var/lib/sysbox`)
- Kubernetes continues using containerd for system pods
- Docker creates its own bridge network (`docker0`)
- All runtimes can operate independently

---

## Deployment

### 1. Review Configuration

Check `helm/daytona-dev/values.yaml`:

```yaml
runner:
  dockerInstaller:
    enabled: true  # Ensure this is true
  config:
    containerRuntime: "docker"  # Ensure this is "docker"
```

### 2. Deploy with Helm

```bash
# Install/upgrade the runner with Docker installer
helm upgrade --install daytona-dev ./helm/daytona-dev \
  --namespace daytona-dev \
  --create-namespace

# Wait for pods to be ready
kubectl wait --for=condition=ready pod \
  -l app=daytona-runner \
  -n daytona-dev \
  --timeout=300s
```

### 3. Verify Installation

```bash
# Check runner pods are running
kubectl get pods -n daytona-dev -l app=daytona-runner

# Check docker-installer sidecar logs
kubectl logs -n daytona-dev <runner-pod-name> -c docker-installer

# Verify Docker is installed on the node
kubectl exec -n daytona-dev <runner-pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker version

# Verify Sysbox is installed and running
kubectl exec -n daytona-dev <runner-pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl status sysbox

# Check Docker runtime configuration
kubectl exec -n daytona-dev <runner-pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker info | grep -i runtime

# Verify Sysbox runtime is available
kubectl exec -n daytona-dev <runner-pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker info | grep sysbox

# Check runner logs to confirm it's using Docker
kubectl logs -n daytona-dev <runner-pod-name> -c runner
```

### 4. Run Validation Script (Optional)

```bash
./helm/daytona-dev/SYSBOX_VALIDATION.sh
# Or with custom namespace:
NAMESPACE=my-namespace ./helm/daytona-dev/SYSBOX_VALIDATION.sh
```

---

## Configuration Options

### Enable/Disable Docker Installer

In `values.yaml`:

```yaml
runner:
  dockerInstaller:
    enabled: true  # Set to false to disable
    image: alpine:latest  # Change to use different base image
```

### Switch Back to Containerd

To revert to containerd:

```yaml
runner:
  dockerInstaller:
    enabled: false
  config:
    containerRuntime: "containerd"
```

Then upgrade the deployment:

```bash
helm upgrade daytona-dev ./helm/daytona-dev -n daytona-dev
```

### Custom Docker Configuration

To add custom Docker daemon configuration, modify the sidecar args to include:

```yaml
echo '{"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}' > /etc/docker/daemon.json
systemctl restart docker
```

### Custom Sysbox Configuration

Sysbox can be configured via `/etc/sysbox/sysbox.conf`. Common configurations:

```bash
# Example: Enable Sysbox debug logging
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i bash -c 'cat > /etc/sysbox/sysbox.conf << EOF
{
  "log-level": "debug",
  "allow-trusted-xattr": true
}
EOF
systemctl restart sysbox'
```

### Disable Sysbox (Use Standard Docker Runtime)

If you want Docker without Sysbox, update the Docker daemon config:

```bash
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i bash -c 'cat > /etc/docker/daemon.json << EOF
{
  "runtimes": {
    "sysbox-runc": {
      "path": "/usr/bin/sysbox-runc"
    }
  },
  "default-runtime": "runc",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
systemctl restart docker'
```

---

## XFS Filesystem Storage

XFS filesystem has been configured for Docker storage on the runner nodes to enable filesystem quota support for limiting container filesystem sizes.

### XFS Setup Process

When the docker-installer sidecar starts, it automatically sets up XFS:

1. **Check Current Filesystem**
   - Detects if `/var/lib/docker` is on XFS
   - If already XFS with quotas: skips setup
   - If not XFS: proceeds with migration

2. **Backup Existing Data**
   - Moves `/var/lib/docker` to `/var/lib/docker.backup.<timestamp>`
   - Preserves existing container data

3. **Create XFS Loopback**
   - Creates sparse file at `/var/lib/docker-xfs.img`
   - Size determined by `XFS_STORAGE_SIZE` environment variable (default: 50GB)
   - Sparse file doesn't consume full size until used

4. **Format and Mount**
   - Formats file as XFS with `-n ftype=1` (required for overlay2)
   - Mounts with `loop,prjquota` options
   - Adds to `/etc/fstab` for persistence

5. **Initialize Quotas**
   - Runs `xfs_quota` to initialize project quota system
   - Ready for per-container filesystem limits

### Changing XFS Storage Size

To modify the XFS storage size, update `values.yaml`:

```yaml
runner:
  dockerInstaller:
    xfsStorageSize: "100G"  # Increase to 100GB
```

Then redeploy:

```bash
helm upgrade daytona-dev ./helm/daytona-dev -n daytona-dev
kubectl delete pod -n daytona-dev -l app=daytona-runner
```

**Note**: Changing size only affects NEW deployments. Existing XFS files won't be automatically resized.

### Using XFS Quotas for Containers

#### Enable Project Quotas

To use XFS project quotas for limiting container filesystem size:

```bash
# Set up a project ID for a container
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i sh -c '
    # Create project
    echo "100:/var/lib/docker/overlay2/<container-id>" >> /etc/projects
    echo "mycontainer:100" >> /etc/projid
    
    # Setup project quota
    xfs_quota -x -c "project -s mycontainer" /var/lib/docker
    
    # Set limit (e.g., 10GB)
    xfs_quota -x -c "limit -p bhard=10g mycontainer" /var/lib/docker
  '
```

#### Check Quota Usage

```bash
# View all quotas
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i xfs_quota -x -c "report -pbih" /var/lib/docker
```

### XFS Verification Commands

```bash
export KUBECONFIG=/workspaces/daytona/.tmp/kubeconfig.yaml
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

# Check XFS mount
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i df -Th /var/lib/docker

# Check quota state
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i xfs_quota -x -c 'state' /var/lib/docker

# Check Docker storage info
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker info | grep -A10 "Storage Driver"

# Check loopback file size
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i ls -lh /var/lib/docker-xfs.img
```

### XFS Benefits

| Benefit | Description |
|---------|-------------|
| **Filesystem Size Limiting** | Set hard limits on container filesystem usage, prevent one container from filling entire disk |
| **Better Multi-tenancy** | Isolate storage usage between tenants, predictable storage allocation |
| **Performance** | XFS optimized for large files, good performance with many small files (overlayfs layers) |
| **Flexibility** | Can grow XFS file if needed, project quotas more flexible than device quotas |

### Expanding XFS Storage

If you need more than 50GB, you can expand the XFS file:

```bash
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i sh -c '
    # Stop Docker
    systemctl stop docker
    
    # Unmount XFS
    umount /var/lib/docker
    
    # Expand file
    truncate -s 100G /var/lib/docker-xfs.img
    
    # Grow XFS filesystem
    xfs_growfs /var/lib/docker-xfs.img
    
    # Remount
    mount /var/lib/docker
    
    # Start Docker
    systemctl start docker
  '
```

### XFS Limitations

| Limitation | Description |
|------------|-------------|
| **Quota Not Auto-Enforced** | Project quotas initialized but not enforced by default; need explicit setup per container |
| **Fixed Size** | 50GB default (configurable); sparse file grows as used; must manually expand if needed |
| **Performance Overhead** | Loopback device adds slight overhead vs. native filesystem; usually negligible |

**Not Implemented Yet:**

- ❌ Automatic quota setting during sandbox creation
- ❌ Dynamic XFS file resizing
- ❌ Quota enforcement in runner application
- ❌ Quota monitoring/alerting

### Future Enhancement: Runner Integration

To fully integrate XFS quotas with the runner, the runner application would need to:

1. **On Sandbox Creation:**

   ```go
   // Set up project quota for container
   projectID := getSandboxProjectID(sandboxID)
   setupProjectQuota(projectID, containerPath, sizeLimit)
   ```

2. **Set Quota Limit:**

   ```bash
   xfs_quota -x -c "project -s -p <path> <projectID>" /var/lib/docker
   xfs_quota -x -c "limit -p bhard=<size> <projectID>" /var/lib/docker
   ```

3. **Monitor Usage:**

   ```bash
   xfs_quota -x -c "report -p <projectID>" /var/lib/docker
   ```

---

## Troubleshooting

### Issue: Docker Installation Fails

**Symptoms**: Sidecar logs show installation errors

**Solutions**:

- Check node OS compatibility
- Verify node has internet access for package downloads
- Check node disk space
- Review sidecar logs for specific errors

```bash
kubectl logs -n daytona-dev <pod-name> -c docker-installer
```

### Issue: Docker Service Won't Start

**Symptoms**: "Docker service is down, restarting..." repeatedly in logs

**Solutions**:

- Check systemd status on the node
- Verify no port conflicts
- Check Docker daemon logs on the node:

  ```bash
  kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
    nsenter -t 1 -m -u -n -i journalctl -u docker
  ```

### Issue: Sysbox Service Won't Start

**Symptoms**: "Sysbox service is down, restarting..." in logs

**Solutions**:

- Check Sysbox requirements (kernel version, modules)
- Verify Sysbox service status:

  ```bash
  kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
    nsenter -t 1 -m -u -n -i journalctl -u sysbox
  ```

- Check kernel compatibility:

  ```bash
  kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
    nsenter -t 1 -m -u -n -i uname -r
  ```

- Sysbox requires kernel 5.5+ with specific features
- Ensure node has necessary kernel modules loaded

### Issue: Sysbox Installation Fails

**Symptoms**: "Failed to install sysbox-ce" in logs

**Solutions**:

- Verify node OS version is supported (Ubuntu 18.04+, Debian 10+, CentOS 8+)
- Check internet connectivity to Nestybox repositories
- Verify GPG key can be added
- Check available disk space
- View detailed installation logs in sidecar logs

### Issue: Node Disk Space Exhaustion

**Symptoms**: Pods evicted, Docker operations failing

**Solutions**:

- Clean up unused Docker images:

  ```bash
  kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
    nsenter -t 1 -m -u -n -i docker system prune -af
  ```

- Monitor disk usage:

  ```bash
  kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
    nsenter -t 1 -m -u -n -i df -h /var/lib/docker
  ```

### Issue: Docker Removed After Node Update

**Symptoms**: Docker not found after GKE node pool upgrade

**Solution**: Docker will be automatically reinstalled when the pod restarts after node update

### Issue: XFS Mount Failed

**Symptoms**: Docker won't start after XFS setup

**Check logs:**

```bash
kubectl logs -n daytona-dev <pod-name> -c docker-installer --tail=50
```

**Common causes:**

- XFS mount failed
- /etc/fstab syntax error
- Insufficient space for XFS file

**Fix:**

```bash
# Check mount
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i mount | grep docker

# Check fstab
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i cat /etc/fstab | grep docker
```

### Issue: XFS Quota Not Working

**Check quota state:**

```bash
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i xfs_quota -x -c 'state' /var/lib/docker
```

**Enable quotas:**

```bash
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i sh -c '
    mount -o remount,prjquota /var/lib/docker
    xfs_quota -x -c "project -s -d 1" /var/lib/docker
  '
```

### Issue: XFS Out of Space

**Check usage:**

```bash
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i df -h /var/lib/docker
```

**Cleanup:**

```bash
kubectl exec -n daytona-dev $POD -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker system prune -af --volumes
```

**Expand XFS**: See "Expanding XFS Storage" section above.

### Common Issues Summary

| Issue | Quick Fix |
|-------|-----------|
| Installer pod stuck | Check logs, verify node meets requirements |
| Docker won't start | Check journalctl, verify port 2375/2376 available |
| Sysbox won't start | Verify kernel 5.5+, check journalctl -u sysbox |
| Out of disk space | Run `docker system prune -af` |
| Runtime not sysbox | Check /etc/docker/daemon.json, restart docker |
| Can't build images | Verify Docker socket mounted, check permissions |
| XFS mount failed | Check /etc/fstab, verify loopback file exists |
| XFS quota not working | Remount with prjquota, check xfs_quota state |
| XFS out of space | Prune Docker, expand XFS file if needed |

---

## Known Limitations

### Sysbox Limitations

1. **Kernel Version**: Requires Linux kernel 5.5+
   - Older GKE versions may not be compatible
   - Check with `uname -r` before deploying

2. **Architecture**: Primarily x86_64
   - Limited ARM support
   - Not compatible with ARM-based GKE nodes

3. **Storage**: Uses more disk space than standard runtimes
   - Plan for additional 1-2GB per node
   - Monitor disk usage regularly

4. **Nested Containers**: Limited nesting depth
   - Docker-in-Docker works well
   - Deep nesting (3+ levels) may have issues

5. **Performance**: Slight overhead compared to runc
   - 10-20% slower container start times
   - Worth the tradeoff for enhanced security

6. **AppArmor/SELinux**: May need adjustments
   - Some security profiles may conflict
   - Test thoroughly in your environment

### Docker + Sysbox Specific Issues

1. **Node Updates**: Docker and Sysbox removed on node replacement
   - Automatically reinstalled when pod starts
   - Brief downtime during reinstallation

2. **Resource Overhead**: Higher memory usage
   - Docker daemon: ~100-200MB
   - Sysbox: ~100-200MB
   - Factor into node sizing

3. **Compatibility**: Not all Docker features supported
   - Most common features work fine
   - Test your specific workloads

---

## Security Considerations

### Privileged Access

- Docker installer sidecar runs with `privileged: true`
- Has full access to host filesystem and namespaces
- Can modify node configuration
- Ensure proper RBAC and network policies are in place
- Consider node isolation for sandbox workloads

### Sysbox Security Benefits

Despite the privileged installer, Sysbox provides enhanced security:

- **Containers don't need privileged mode** for Docker-in-Docker
- **User namespace isolation** prevents container root from being host root
- **Reduced attack surface** compared to `--privileged` containers
- **Better multi-tenancy** with stronger isolation between containers

### Security Posture Summary

**Threats Mitigated:**

- ✅ Container breakout (improved isolation)
- ✅ Privilege escalation (user namespaces)
- ✅ Resource exhaustion (better limits)
- ✅ Cross-tenant contamination (stronger isolation)

**Remaining Considerations:**

- ⚠️ Installer runs privileged (required, limited scope)
- ⚠️ Docker daemon has host access (by design)
- ⚠️ Monitor for CVEs in Docker/Sysbox
- ⚠️ Regular security updates needed

### Best Practices

1. **Node Isolation**: Use dedicated node pools for sandbox workloads

   ```bash
   # Example node pool with taints
   gcloud container node-pools create sandbox-pool \
     --cluster=your-cluster \
     --machine-type=n2-standard-4 \
     --node-taints=sandbox=true:NoSchedule \
     --node-labels=workload-type=daytona-sandbox-c
   ```

2. **Network Policies**: Restrict network access

   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: sandbox-isolation
   spec:
     podSelector:
       matchLabels:
         app: daytona-runner
     policyTypes:
     - Ingress
     - Egress
     ingress:
     - from:
       - namespaceSelector:
           matchLabels:
             name: daytona-dev
   ```

3. **Resource Quotas**: Limit resource consumption

   ```yaml
   apiVersion: v1
   kind: ResourceQuota
   metadata:
     name: sandbox-quota
   spec:
     hard:
       requests.cpu: "100"
       requests.memory: 200Gi
       persistentvolumeclaims: "10"
   ```

4. **Audit Logging**: Monitor Docker and Sysbox activity

   ```bash
   # Enable Docker audit logging
   kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
     nsenter -t 1 -m -u -n -i bash -c \
     'echo "audit" >> /etc/docker/daemon.json && systemctl restart docker'
   ```

---

## Resource Considerations

### Disk Space

- Docker images stored in `/var/lib/docker`
- Sysbox data stored in `/var/lib/sysbox`
- Monitor disk usage to prevent node storage exhaustion
- Consider setting up Docker image cleanup policies

### Memory & CPU

- Docker daemon consumes additional node resources
- Default runner resource limits:
  - Requests: 500m CPU, 512Mi RAM
  - Limits: 2000m CPU, 4Gi RAM

### Network

- Docker creates `docker0` bridge network
- Doesn't conflict with Kubernetes CNI networking
- Sandboxes run in Docker's network namespace

### Performance Comparison

| Metric | Containerd | Docker | Docker + Sysbox |
|--------|-----------|---------|-----------------|
| Container Start Time | ~50-100ms | ~100-200ms | ~150-300ms |
| Memory Overhead | Lower | Higher (~100-200MB) | Highest (~300-400MB) |
| Disk Space | Lower | Higher (~500MB) | Highest (~1GB) |
| API Compatibility | Limited | Full Docker API | Full Docker API |
| Tooling Support | Growing | Extensive | Extensive |
| Security Isolation | Good | Good | Excellent |
| Docker-in-Docker | No | Requires privileged | Yes, secure |
| Systemd Support | Limited | No | Yes |

### Optimization Tips

1. **Prune regularly**: Set up CronJob for `docker system prune`
2. **Log rotation**: Already configured (10MB max, 3 files)
3. **Monitor disk**: Set alerts for >80% disk usage
4. **Resource limits**: Configure appropriate pod limits
5. **Node sizing**: Plan for additional overhead

---

## Maintenance

### Monitoring

#### Key Metrics to Watch

```bash
# Check Sysbox resource usage
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl status sysbox

# View Sysbox logs
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i journalctl -u sysbox -n 100

# Check Docker daemon status
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl status docker

# Monitor disk usage
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i du -sh /var/lib/docker /var/lib/sysbox

# Container count
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker ps -q | wc -l

# Resource usage
kubectl top pod -n daytona-dev -l app=daytona-runner
kubectl top node -l workload-type=daytona-sandbox-c
```

#### Log Locations

- **Installer Logs**: `kubectl logs -n daytona-dev <pod-name> -c docker-installer`
- **Runner Logs**: `kubectl logs -n daytona-dev <pod-name> -c runner`
- **Docker Logs**: Via `journalctl -u docker` on host
- **Sysbox Logs**: Via `journalctl -u sysbox` on host

### Regular Cleanup

Set up a cleanup CronJob:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: docker-sysbox-cleanup
  namespace: daytona-dev
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          hostNetwork: true
          hostPID: true
          containers:
          - name: cleanup
            image: alpine:latest
            securityContext:
              privileged: true
            command:
            - /bin/sh
            - -c
            - |
              nsenter -t 1 -m -u -n -i bash << 'EOF'
              echo "Cleaning up Docker resources..."
              docker system prune -af --volumes
              
              echo "Cleaning up old Sysbox data..."
              # Sysbox cleanup is automatic, but you can check for orphaned data
              find /var/lib/sysbox -type f -mtime +7 -name "*.log" -delete 2>/dev/null || true
              
              echo "Cleanup completed"
              df -h /var/lib/docker /var/lib/sysbox
              EOF
          restartPolicy: OnFailure
```

---

## Advanced Usage

### Manual Docker Commands

Execute Docker commands on the host:

```bash
# List containers
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker ps

# View images
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker images

# Check Docker info
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker info

# Run a container with Sysbox runtime (explicit)
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker run --runtime=sysbox-runc -it ubuntu bash

# Check Sysbox version
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i sysbox --version
```

### Testing Sysbox Capabilities

Test Docker-in-Docker with Sysbox:

```bash
# Run a container with Docker inside (no privileged mode needed!)
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker run -d --name test-dind \
  --hostname test-dind \
  docker:dind

# Execute into the container and run Docker commands
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker exec -it test-dind docker ps

# Run systemd inside a container
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker run -d --name test-systemd \
  jrei/systemd-ubuntu:20.04
```

---

## Use Cases

### Use Case 1: Docker Build in Sandbox

Create a sandbox that can build Docker images:

```bash
# The sandbox will automatically use Sysbox runtime
# No need for --privileged flag!

# Inside the sandbox:
docker build -t myapp:latest .
docker run myapp:latest
```

### Use Case 2: Running Development Environment with Docker Compose

```bash
# Inside a Daytona sandbox:
cat > docker-compose.yml << EOF
version: '3'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  db:
    image: postgres:13
    environment:
      POSTGRES_PASSWORD: example
EOF

docker-compose up -d
docker-compose ps
```

### Use Case 3: Testing Kubernetes Manifests with kind

```bash
# Run a local Kubernetes cluster inside the sandbox
kind create cluster --name test-cluster
kubectl cluster-info --context kind-test-cluster
kubectl apply -f manifests/
```

### Use Case 4: Multi-stage CI/CD Pipeline

```bash
# Build, test, and push images all within a sandbox
docker build -t myapp:test .
docker run myapp:test npm test
docker tag myapp:test registry.example.com/myapp:latest
docker push registry.example.com/myapp:latest
```

---

## Rollback Plan

To revert to containerd:

### 1. Update values.yaml

```yaml
runner:
  dockerInstaller:
    enabled: false  # Disable installer
  config:
    containerRuntime: "containerd"  # Switch back
```

### 2. Redeploy

```bash
helm upgrade daytona-dev ./helm/daytona-dev -n daytona-dev
```

### 3. Optional: Remove Docker from Nodes

```bash
# If you want to remove Docker from nodes
POD=$(kubectl get pod -n daytona-dev -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}')

kubectl exec -n daytona-dev $POD -c runner -- \
  nsenter -t 1 -m -u -n -i bash -c '
    systemctl stop docker sysbox
    apt-get remove -y docker.io sysbox-ce
    rm -rf /var/lib/docker /var/lib/sysbox
  '
```

---

## Testing Checklist

- [ ] Runner pods deploy successfully
- [ ] Docker installer completes without errors
- [ ] Docker service is running on nodes
- [ ] Sysbox service is running on nodes
- [ ] Docker runtime is sysbox-runc
- [ ] XFS filesystem mounted at /var/lib/docker
- [ ] XFS project quota support enabled
- [ ] Runner can create sandboxes
- [ ] Sandboxes can run Docker commands
- [ ] Docker-in-Docker works without --privileged
- [ ] Docker Compose works in sandboxes
- [ ] systemd works in containers (if needed)
- [ ] Services recover after restart
- [ ] Disk space is monitored
- [ ] No impact on Kubernetes system pods

---

## Quick Reference

### Common Commands

```bash
# Check pod status
kubectl get pods -n daytona-dev -l app=daytona-runner -o wide

# View installer logs (full output)
kubectl logs -n daytona-dev <pod-name> -c docker-installer --tail=100

# Check Docker version on host
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker version

# Check Sysbox status
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl status sysbox

# View running containers on host
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker ps -a

# Check disk usage
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i df -h

# Cleanup Docker
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i docker system prune -af

# Restart services
kubectl exec -n daytona-dev <pod-name> -c docker-installer -- \
  nsenter -t 1 -m -u -n -i systemctl restart sysbox docker
```

### Configuration Files

| File | Purpose |
|------|---------|
| `/etc/docker/daemon.json` | Docker daemon configuration |
| `/etc/sysbox/sysbox.conf` | Sysbox configuration |
| `/var/lib/docker` | Docker storage (mounted on XFS) |
| `/var/lib/sysbox` | Sysbox storage |
| `/var/lib/docker-xfs.img` | XFS loopback file (50GB) |
| `/etc/projects` | XFS project definitions |
| `/etc/projid` | XFS project ID mappings |
| `helm/daytona-dev/values.yaml` | Helm values |
| `helm/daytona-dev/templates/runner.yaml` | DaemonSet template |

### Important Values.yaml Settings

```yaml
runner:
  dockerInstaller:
    enabled: true              # Enable/disable Docker+Sysbox installer
    image: alpine:latest       # Base image for installer
    xfsStorageSize: "50G"      # XFS loopback file size for Docker storage
  
  config:
    containerRuntime: "docker" # Use "docker" or "containerd"
```

### Troubleshooting Checklist

- [ ] Check runner pod is running: `kubectl get pods -n daytona-dev`
- [ ] View installer logs: `kubectl logs -n daytona-dev <pod-name> -c docker-installer`
- [ ] Verify Docker installed: `docker version` via nsenter
- [ ] Verify Sysbox installed: `systemctl status sysbox` via nsenter
- [ ] Check kernel version: `uname -r` (should be 5.5+)
- [ ] Check disk space: `df -h /var/lib/docker /var/lib/sysbox`
- [ ] Check XFS mount: `df -Th /var/lib/docker` (should show xfs)
- [ ] Check XFS quota state: `xfs_quota -x -c 'state' /var/lib/docker`
- [ ] Review node resources: `kubectl describe node <node-name>`
- [ ] Check container runtime: `docker info | grep -i runtime`

---

## Next Steps

1. ✅ Deploy to development environment
2. ⏳ Run validation script
3. ⏳ Test sandbox creation with Docker
4. ⏳ Test Docker-in-Docker functionality
5. ⏳ Monitor resource usage for 24-48 hours
6. ⏳ Deploy to staging
7. ⏳ Performance testing
8. ⏳ Security review
9. ⏳ Deploy to production

---

## Support and Resources

### Documentation

- **Sysbox Documentation**: https://github.com/nestybox/sysbox
- **Sysbox Quick Start**: https://github.com/nestybox/sysbox/blob/master/docs/quickstart/README.md
- **Docker Documentation**: https://docs.docker.com/

### Getting Help

For issues or questions:

- Check runner logs: `kubectl logs -n daytona-dev <pod-name> -c runner`
- Check installer logs: `kubectl logs -n daytona-dev <pod-name> -c docker-installer`
- Review Docker daemon logs: `journalctl -u docker` via nsenter
- Review Sysbox logs: `journalctl -u sysbox` via nsenter
- Check node resources and disk space
- Verify kernel compatibility
- Review Sysbox GitHub issues: https://github.com/nestybox/sysbox/issues

### Validation Script

```bash
./helm/daytona-dev/SYSBOX_VALIDATION.sh
```
