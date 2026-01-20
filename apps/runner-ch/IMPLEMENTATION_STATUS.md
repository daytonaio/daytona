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

## 🚧 Partially Implemented

### Snapshots & Images

- [ ] Pull snapshot from S3 (stub)
- [ ] Push snapshot to S3 (stub)
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
- [ ] Memory state snapshots (for instant resume)

### Performance Optimizations

- [x] TAP pool (pre-created TAP interfaces for fast VM creation)
- [x] SSH command batching (single SSH call for disk creation + socket polling)
- [ ] btrfs/XFS reflink support for raw disks
- [ ] Pre-warmed VM pool for instant fork
- [ ] Memory ballooning
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

## Known Issues

1. **S3 integration missing** - Snapshot push/pull not implemented
2. **GPU passthrough untested** - VFIO code exists but needs testing
3. **Port conflicts** - SSH gateway (port 2220) may conflict with runner-win if both running

## Next Steps

1. **S3 Snapshots** - Implement push/pull for snapshot portability
2. **GPU Testing** - Test VFIO passthrough with actual GPUs
3. **Memory Snapshots** - Implement memory state save/restore for instant resume
4. **Testing** - End-to-end tests with actual workloads

## Comparison with runner-win

| Feature | runner-win (libvirt) | runner-ch |
|---------|---------------------|-----------|
| Hypervisor | QEMU/KVM | Cloud Hypervisor |
| OS Support | Windows + Linux | Linux only |
| GPU Passthrough | ✅ | 🚧 (untested) |
| Live Migration | ✅ | ❌ |
| Memory Hotplug | ✅ | ✅ |
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

---

_Last updated: 2026-01-20 (SSH Gateway + Local Mode)_
