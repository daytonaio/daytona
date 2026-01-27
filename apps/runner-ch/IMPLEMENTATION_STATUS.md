# Runner-CH Implementation Status

Cloud Hypervisor runner for Daytona - implementation status and roadmap.

## Overview

`runner-ch` is a Cloud Hypervisor-based runner designed for Linux VMs with fast startup times, GPU passthrough support, and efficient resource usage.

**Target Host:** `206.223.225.17` (48 cores, 377GB RAM, 438GB storage)

## ✅ Completed Features

### Core Infrastructure

- [x] Cloud Hypervisor REST API client
- [x] **Dual-mode operation: Local + Remote (SSH)**
  - Local mode: Runner on same host as Cloud Hypervisor (production)
  - Remote mode: Runner connects via SSH (development)
- [x] API server with Gin framework
- [x] Authentication middleware
- [x] Prometheus metrics endpoint

### SSH Gateway

- [x] SSH gateway service (port 2220)
- [x] Public key authentication (matches global SSH gateway)
- [x] Dual-mode: Direct connection (local) or SSH tunnel (remote)
- [x] Channel and request forwarding to daemon
- [x] Proper channel rejection on connection failure

### VM Lifecycle

- [x] Create sandbox (qcow2 overlay, instant CoW)
- [x] Start sandbox (boot VM)
- [x] Stop sandbox (pause VM)
- [x] Destroy sandbox (cleanup resources)
- [x] Get sandbox info

### Disk Management

- [x] qcow2 overlay with backing file (instant creation)
- [x] Disk quota via qcow2 virtual size
- [x] Support for custom base snapshots

### Networking

- [x] TAP interface creation/deletion
- [x] Bridge networking (br0)
- [x] MAC address generation
- [x] Network settings update (stub)
- [x] **Network namespace pool** for VM isolation
  - Per-VM network namespaces (ns-<sandbox-id>)
  - veth pairs for host connectivity
  - NAT for external access
  - Automatic IP allocation (10.0.X.0/24 per namespace)

### Metrics & Healthcheck

- [x] Remote metrics collection via SSH (CPU, RAM, disk)
- [x] Allocated resources tracking
- [x] Healthcheck service (reports to Daytona API)
- [x] Snapshot count tracking

### Job System (v2 API)

- [x] Job poller service
- [x] Job executor with OpenTelemetry tracing
- [x] CREATE_SANDBOX job handler
- [x] START_SANDBOX job handler
- [x] STOP_SANDBOX job handler
- [x] DESTROY_SANDBOX job handler

### Snapshots

- [x] Create snapshot from VM
- [x] List snapshots
- [x] Get snapshot info
- [x] Delete snapshot
- [x] Fork VM from snapshot

### Memory Ballooning

- [x] **Balloon device support** in VM creation (deflate_on_oom, free_page_reporting)
- [x] **Memory stats collection** from guest via daemon `/memory-stats` endpoint
- [x] **Memory controller** background loop (configurable interval, default 30s)
- [x] **Automatic balloon inflation** to reclaim unused guest memory
- [x] **Automatic balloon deflation** when guest needs more memory
- [x] **StatsStore** for persistent memory statistics history
- [x] **Stats API endpoints**: `/stats/memory` (JSON), `/stats/memory/view` (HTML dashboard)
- [x] Network namespace-aware daemon queries (curl inside ns-<id>)

**Configuration:**

```bash
MEMORY_BALLOONING_ENABLED=true        # Enable/disable ballooning
MEMORY_BALLOONING_INTERVAL_SEC=30     # Collection/rebalance interval
MEMORY_BALLOONING_MIN_VM_GB=1         # Minimum memory per VM (GB)
MEMORY_BALLOONING_BUFFER_GB=2         # Safety buffer (GB)
MEMORY_BALLOONING_BUFFER_RATIO=0.25   # Safety ratio (25% of used)
```

**Memory Target Formula:**

```
target = max(used_memory × (1 + buffer_ratio), min_vm_memory) + buffer_gb
balloon_size = allocated_memory - target
```

**Example:** VM with 4GB allocated, 400MB used:

- Target = max(400MB × 1.25, 1GB) + 2GB = 3GB
- Balloon = 4GB - 3GB = 1GB reclaimed

### Live Fork (Local Mode)

- [x] **Fork VM with memory state preservation** (vm.snapshot/vm.restore)
- [x] CoW disk overlay (instant copy-on-write, no full disk copy)
- [x] Network namespace isolation (each fork gets own namespace)
- [x] Automatic disk path patching in snapshot config
- [x] Parent sandbox tracking (`parent` file in sandbox dir)
- [x] Fork cleanup on failure
- [ ] TAP FD passing via SCM_RIGHTS (falls back to cold restore - CH returns 400)

**Fork Process:**

1. Pause source VM
2. Create snapshot via vm.snapshot (memory + device state)

### Warm Snapshots (Instant Restore)

- [x] **Warm snapshot detection** (checks for `memory-ranges` + `disk.qcow2` in snapshot directory)
- [x] **Instant VM restore** (~4 seconds vs ~16 seconds cold boot)
- [x] CoW disk overlay from golden disk
- [x] Network namespace creation
- [x] Automatic disk path patching in restore config

**Performance Comparison:**

| Snapshot Type | Time | Notes |
|---------------|------|-------|
| Warm snapshot | ~4 sec | Memory state restored, daemon instant |
| Cold boot | ~16 sec | Full kernel + systemd boot |

**Warm Snapshot Structure:**

```
/var/lib/cloud-hypervisor/snapshots/ubuntu-base.2/
├── disk.qcow2       # Golden disk image (base for CoW overlays)
├── memory-ranges    # VM memory state
├── state.json       # Device state
└── config.json      # VM config (disk path patched at restore time)
```

**Creating a Warm Snapshot:**

```bash
# 1. Create source VM and wait for daemon
curl -X POST -d '{"id":"golden-source","snapshot":"ubuntu-base.1",...}' /sandboxes

# 2. Pause VM and take snapshot
curl -X PUT --unix-socket /var/run/cloud-hypervisor/golden-source.sock \
  http://localhost/api/v1/vm.pause

curl -X PUT -d '{"destination_url":"file:///path/to/snapshot"}' \
  --unix-socket /var/run/cloud-hypervisor/golden-source.sock \
  http://localhost/api/v1/vm.snapshot

# 3. Copy disk to snapshot directory
cp /var/lib/cloud-hypervisor/sandboxes/golden-source/disk.qcow2 \
   /var/lib/cloud-hypervisor/snapshots/ubuntu-base.2/disk.qcow2
```

**Fork Process (continued):**
3. Create qcow2 overlay disk with source as backing file
4. Patch snapshot config.json to use new overlay disk
5. Create network namespace for forked VM
6. Start cloud-hypervisor in namespace
7. Restore VM via vm.restore (memory state preserved)
8. Resume both VMs
9. Cleanup temporary snapshot

## 🚧 Partially Implemented

### Snapshots & Images

- [ ] Pull snapshot from S3 (stub)
- [x] Push snapshot to S3 (uploads to `{bucket}/{orgId}/{snapshotName}/`)
- [ ] Build snapshot from Dockerfile (stub)
- [ ] Tag image (stub)

### Networking & Proxy

- [x] Static IP pool (10.0.0.2 - 10.0.0.254)
- [x] Cloud-init ISO for static IP configuration (no DHCP wait)
- [x] IP allocation instant (0ms overhead)
- [x] Toolbox proxy (`/sandboxes/:id/toolbox/*`) - dual mode
- [x] Port proxy (`/sandboxes/:id/proxy/:port/*`) - dual mode
- [x] IP cache for sandbox IPs
- [x] SOCKS5 proxy via SSH (remote mode)
- [x] Direct HTTP proxy (local mode)
- [x] Persistent IP storage in sandbox directory
- [x] SSH gateway for sandbox SSH access

### VM Features

- [ ] Live resize (CPU/memory hotplug) - implemented but untested
- [ ] GPU passthrough - API ready, needs VFIO testing

## ❌ Not Yet Implemented

### High Priority

- [ ] Backup to S3
- [x] Memory state snapshots (implemented via fork - vm.snapshot/vm.restore)

### Performance Optimizations

- [x] TAP pool (pre-created TAP interfaces for fast VM creation)
- [x] SSH command batching (single SSH call for disk creation + socket polling)
- [ ] btrfs/XFS reflink support for raw disks
- [ ] Pre-warmed VM pool for instant fork
- [x] **Memory ballooning** (active reclamation of unused guest memory)
- [ ] Huge pages support

### Advanced Features

- [ ] Live migration
- [ ] Nested virtualization
- [ ] Custom kernel boot (PVH)
- [ ] Serial console access
- [ ] VNC/SPICE display

## Configuration

### Environment Variables

```bash
# Required
SERVER_URL=http://localhost:3000      # Daytona API URL
API_TOKEN=<token>                      # Runner API token

# Cloud Hypervisor Host (remote mode only)
# Leave CH_SSH_HOST empty for local mode
CH_SSH_HOST=root@206.223.225.17
CH_SSH_KEY_PATH=/path/to/id_rsa

# Optional
API_PORT=3005                          # Runner API port
LOG_LEVEL=info                         # debug, info, warn, error

# VM Defaults
CH_DEFAULT_CPUS=2
CH_DEFAULT_MEMORY_MB=2048
CH_DEFAULT_DISK_GB=20

# SSH Gateway (optional)
SSH_GATEWAY_ENABLE=true               # Enable SSH gateway
SSH_GATEWAY_PORT=2220                 # SSH gateway port
SSH_PUBLIC_KEY=<base64-encoded-key>   # Public key for authentication
SSH_HOST_KEY_PATH=/root/.ssh/id_rsa   # Host key path
SANDBOX_SSH_USER=daytona              # User for sandbox SSH
SANDBOX_SSH_PORT=22220                # SSH port inside sandbox

# Performance
TAP_POOL_ENABLED=true                 # Pre-create TAP interfaces
TAP_POOL_SIZE=10                      # Number of TAPs to pre-create

# Memory Ballooning
MEMORY_BALLOONING_ENABLED=true        # Enable memory ballooning
MEMORY_BALLOONING_INTERVAL_SEC=30     # Stats collection interval
MEMORY_BALLOONING_MIN_VM_GB=1         # Minimum memory per VM
MEMORY_BALLOONING_BUFFER_GB=2         # Safety buffer in GB
MEMORY_BALLOONING_BUFFER_RATIO=0.25   # Safety ratio (25% of used)
```

### Directory Structure (Remote Host)

```
/var/lib/cloud-hypervisor/
├── firmware/           # OVMF/hypervisor-fw
├── kernels/            # vmlinux (for PVH boot)
├── images/             # Base images
├── sandboxes/          # VM working directories
│   └── <sandbox-id>/
│       ├── disk.qcow2  # VM disk (overlay)
│       └── config.json # VM configuration
└── snapshots/          # Snapshot storage
    └── <snapshot-name>/
        ├── disk.qcow2  # Snapshot disk
        └── memory.bin  # Memory state (optional)

/var/run/cloud-hypervisor/
└── <sandbox-id>.sock   # API sockets
```

## Technical Decisions

### Disk Format: qcow2 (not raw)

**Chosen:** qcow2 with backing file

- ✅ Instant creation (CoW)
- ✅ Built-in quota (virtual size)
- ✅ Efficient storage (only deltas)
- ⚠️ ~20-30% slower I/O than raw

**Future option:** raw + btrfs/XFS reflink

- Requires filesystem support
- Native I/O performance
- Needs separate quota mechanism

### Memory Alignment

Cloud Hypervisor's virtio-mem requires 128 MiB alignment.
Memory is automatically rounded up to nearest 128 MiB boundary.
Minimum memory: 1 GB (1024 MB)

### TAP Interface Naming

Linux limits interface names to 15 characters.
Format: `tap-<11 chars from sandbox ID>` = 15 chars max

## API Endpoints

| Endpoint | Method | Status |
|----------|--------|--------|
| `/` | GET | ✅ Health check |
| `/info` | GET | ✅ Runner info |
| `/metrics` | GET | ✅ Prometheus metrics |
| `/sandboxes` | POST | ✅ Create sandbox |
| `/sandboxes/:id` | GET | ✅ Get sandbox info |
| `/sandboxes/:id` | DELETE | ✅ Remove destroyed |
| `/sandboxes/:id/start` | POST | ✅ Start sandbox |
| `/sandboxes/:id/stop` | POST | ✅ Stop sandbox |
| `/sandboxes/:id/destroy` | POST | ✅ Destroy sandbox |
| `/sandboxes/:id/fork` | POST | ✅ Fork sandbox (local mode) |
| `/sandboxes/:id/resize` | POST | 🚧 Resize (untested) |
| `/sandboxes/:id/backup` | POST | 🚧 Stub |
| `/sandboxes/:id/network-settings` | POST | 🚧 Stub |
| `/sandboxes/:id/toolbox/*` | ANY | ✅ SSH tunnel proxy |
| `/sandboxes/:id/proxy/:port/*` | ANY | ✅ SSH tunnel proxy |
| `/snapshots/pull` | POST | 🚧 Stub |
| `/snapshots/push` | POST | 🚧 Stub |
| `/snapshots/create` | POST | ✅ Works |
| `/snapshots/build` | POST | 🚧 Stub |
| `/snapshots/exists` | GET | ✅ Works |
| `/snapshots/info` | GET | ✅ Works |
| `/snapshots/remove` | POST | ✅ Works |
| `/stats/memory` | GET | ✅ Memory stats JSON |
| `/stats/memory/view` | GET | ✅ Memory stats HTML dashboard |

## Known Issues

1. **S3 pull not implemented** - Snapshot pull from S3 not yet implemented (push works)
2. **GPU passthrough untested** - VFIO code exists but needs testing
3. **Port conflicts** - SSH gateway (port 2220) may conflict with runner-win if both running
4. **Fork FD passing** - TAP FD passing via SCM_RIGHTS returns 400 from CH; falls back to cold restore (which still works with memory state)
5. **Fork in remote mode** - Live fork only works in local mode; remote mode cannot pass TAP FDs over SSH

## Next Steps

1. **S3 Pull** - Implement snapshot pull from S3 for portability
2. **GPU Testing** - Test VFIO passthrough with actual GPUs
3. **Fork FD Passing** - Investigate CH API format for proper SCM_RIGHTS FD passing
4. **Testing** - End-to-end tests with actual workloads

## Comparison with runner-win

| Feature | runner-win (libvirt) | runner-ch |
|---------|---------------------|-----------|
| Hypervisor | QEMU/KVM | Cloud Hypervisor |
| OS Support | Windows + Linux | Linux only |
| GPU Passthrough | ✅ | 🚧 (untested) |
| Live Migration | ✅ | ❌ |
| Live Fork | ❌ | ✅ (local mode) |
| Memory Hotplug | ✅ | ✅ |
| Memory Ballooning | ✅ | ✅ |
| Boot Time | ~10-30s | ~5-10s |
| Memory Overhead | Higher | Lower |
| Disk Format | qcow2 | qcow2 |

## Architecture

### SSH Connection Flow

```
User → Global SSH Gateway (2222) → Runner SSH Gateway (2220) → Daemon SSH (22220)
         │                              │                           │
         │ token as username            │ sandboxId as username     │ password auth
         │ public key auth              │ public key auth           │ "sandbox-ssh"
```

### Local vs Remote Mode

| Operation | Local Mode | Remote Mode |
|-----------|------------|-------------|
| Shell commands | `/bin/sh -c` | `ssh user@host` |
| File operations | `os.Stat`, `os.ReadFile` | SSH commands |
| Proxy to VM | Direct HTTP | SOCKS5 via SSH |
| SSH Gateway | Direct dial | SSH tunnel dial |
| Metrics | Local `gopsutil` | SSH to remote host |
| Fork | ✅ Full (memory state) | ⚠️ Cold only (disk) |

### Fork Architecture

```
Source VM (Running)
     │
     ├─1─► vm.pause
     │
     ├─2─► vm.snapshot ──► /snapshots/fork-<id>-<timestamp>/
     │                         ├── config.json (patched disk path)
     │                         └── memory-*
     │
     ├─3─► qcow2 create ──► /sandboxes/<fork-id>/disk.qcow2
     │                         (backing: source disk)
     │
     ├─4─► Create NetNS ──► ns-<fork-id>
     │                         ├── tap0 (192.168.0.1/24)
     │                         └── veth pair (10.0.X.0/24)
     │
     ├─5─► Start CH in NS ──► cloud-hypervisor --api-socket
     │
     ├─6─► vm.restore ────► Forked VM (memory state preserved)
     │
     └─7─► vm.resume (both VMs)

Forked VM: Independent disk (CoW), independent network, same memory state
```

### Network Namespace Isolation

Each VM runs in its own network namespace:

```
Host Network
    │
    ├── br0 (bridge)
    │
    ├── veth-<id> ◄──► veth-<id>-n (in namespace)
    │                        │
    │                   ns-<sandbox-id>
    │                        │
    │                   tap0 (192.168.0.1)
    │                        │
    │                   VM (192.168.0.2)
    │
    └── NAT rules for external access
```

---

_Last updated: 2026-01-25 (Memory Ballooning Implementation)_
