# Runner-Android Implementation Status

Cuttlefish (Android Virtual Device) runner for Daytona - implementation status and roadmap.

## Overview

`runner-android` is a Cuttlefish-based runner designed for Android virtual devices with WebRTC display streaming, ADB-based device management, and multi-instance support.

**Hypervisor:** crosvm (via Cuttlefish/CVD)
**Target Host:** Bare-metal Linux x86_64 with KVM support
**Current Test Host:** `40.160.30.187` (32 cores, 125GB RAM, 878GB storage)

## ✅ Completed Features

### Core Infrastructure

- [x] Cuttlefish CVD client (create, start, stop, destroy via `cvd` CLI)
- [x] **Dual-mode operation: Local + Remote (SSH)**
  - Local mode: Runner on same host as Cuttlefish (production)
  - Remote mode: Runner connects via SSH (development)
- [x] API server with Gin framework
- [x] Authentication middleware
- [x] Prometheus metrics endpoint
- [x] Instance mapping persistence (`mappings.json`)

### VM Lifecycle

- [x] Create sandbox (`cvd create` with configurable CPUs, memory, snapshot)
- [x] Start sandbox (`cvd start` for stopped instances, re-create as fallback)
- [x] Stop sandbox (`cvd stop` by group name)
- [x] Destroy sandbox (`cvd rm` + process cleanup + directory cleanup)
- [x] Get sandbox info (state, ADB port, metadata)
- [x] **Multi-layered health detection** (CVD fleet → ADB liveness → crosvm process)
- [x] **Stale VM recovery** (handles CVD "Cancelled" status after server restarts)

### ADB Integration (instead of SSH)

All device interactions use ADB (Android Debug Bridge) instead of SSH:

- [x] ADB shell command execution
- [x] ADB file push/pull
- [x] ADB port forwarding info
- [x] ADB device discovery and serial management
- [x] Per-instance ADB ports (base port 6520 + instance_num - 1)

### WebRTC Display Streaming

- [x] **Cuttlefish WebRTC operator** integration (port 1443 HTTPS)
- [x] **WebRTC proxy** from runner API to operator
- [x] **Clean URL support** (`/` serves `client.html` without redirect)
  - Injects `window.__DAYTONA_DEVICE_ID` for device identification
  - Intercepts and modifies `server_connector.js` for clean URL compatibility
  - `<base>` tag injection for correct asset loading
- [x] **WebSocket proxying** for WebRTC signaling
- [x] **Device ID auto-discovery** from operator `/devices` endpoint
- [x] Port 6080 proxy routing (per-sandbox WebRTC access)

### Toolbox API (Partial — via ADB)

Toolbox commands are implemented via ADB shell instead of a daemon inside the VM.

**Supported commands:**

| Endpoint | Method | Status | Notes |
|----------|--------|--------|-------|
| `process/execute` | POST | ✅ | Execute shell command via `adb shell` |
| `process/commands/{id}` | GET | ❌ | Not supported (commands are synchronous) |
| `files` | GET | ✅ | List files via `adb shell ls` |
| `files/info` | GET | ✅ | File info via `adb shell stat` |
| `files/download` | GET | ✅ | Download via `adb pull` |
| `files/upload` | POST | ✅ | Upload via `adb push` |
| `files/folder` | POST | ✅ | Create folder via `adb shell mkdir` |
| `files/move` | POST | ✅ | Move/rename via `adb shell mv` |
| `files` | DELETE | ✅ | Delete via `adb shell rm` |
| `git/*` | ANY | ❌ | Not supported on Android |
| `workspace` | GET | ✅ | Returns `/sdcard` as default workspace |
| `computeruse/status` | GET | ✅ | WebRTC availability status |
| `computeruse/screenshot` | GET/POST | ✅ | Screenshot via `adb shell screencap` |
| `computeruse/keyboard/type` | POST | ✅ | Text input via `adb shell input text` |
| `computeruse/keyboard/key` | POST | ✅ | Key press via `adb shell input keyevent` |
| `computeruse/mouse/click` | POST | ✅ | Tap via `adb shell input tap` |
| `computeruse/mouse/move` | POST | ✅ | Move cursor |
| `computeruse/mouse/drag` | POST | ✅ | Swipe via `adb shell input swipe` |
| `computeruse/mouse/scroll` | POST | ✅ | Scroll via `adb shell input roll` |

### Android-Specific Endpoints

| Endpoint | Method | Status | Notes |
|----------|--------|--------|-------|
| `/sandboxes/:id/adb/info` | GET | ✅ | ADB port, serial, tunnel command |
| `/sandboxes/:id/android/install` | POST | ✅ | Install APK (multipart or base64) |
| `/sandboxes/:id/android/uninstall` | POST | ✅ | Uninstall app by package name |
| `/sandboxes/:id/android/packages` | GET | ✅ | List installed packages |
| `/sandboxes/:id/android/launch` | POST | ✅ | Launch app/activity |
| `/sandboxes/:id/android/stop` | POST | ✅ | Force stop app |
| `/sandboxes/:id/android/props` | GET | ✅ | Get system properties |
| `/sandboxes/:id/android/logcat` | GET | ✅ | Stream logcat via SSE |
| `/sandboxes/:id/android/device` | GET | ✅ | Device info (model, version, SDK) |

### Health Monitoring

- [x] **Multi-layered instance state detection:**
  1. CVD fleet status (`cvd fleet --json`) — fast but can show stale "Cancelled"
  2. ADB liveness (`adb shell getprop sys.boot_completed`) — ground truth
  3. crosvm process detection (`ps -eo args | grep crosvm`) — fallback
- [x] **Crash detection** with configurable retry count before reporting
- [x] **Automatic crash reporting** to Daytona API
- [x] **CVD state synchronization** (removes orphaned CVD instances)
- [x] **Operator device cleanup** (removes stale cuttlefish-operator registrations)

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

### Snapshots & S3

- [x] Snapshot existence check (local directory-based)
- [x] Base image support (Cuttlefish system images as snapshots)
- [x] Custom snapshot support (org-scoped, `{orgId}/{snapshotName}` format)
- [x] Snapshot directory management
- [x] **Create snapshot from running VM** — copies per-instance disk files (overlay.img, sdcard.img, etc.)
- [x] **S3 upload** — auto-uploads snapshot to S3 after local creation
- [x] **S3 download** — pulls snapshot from S3 if not present locally
- [x] **S3 existence check** — checks S3 for snapshot availability
- [x] **S3 delete** — removes snapshot from S3
- [x] Remote file streaming (SSH pipe to S3 for remote mode)
- [x] Symlink resolution (follows symlinks before uploading to S3)

**Snapshot structure (local):**

```
/var/lib/cuttlefish/artifacts/snapshots/{orgId}/{snapshotName}/
├── manifest.json       # Metadata (name, org, base image, creation time, size)
├── instance_data/      # Per-instance disk files (the actual VM state)
│   ├── overlay.img     # System overlay (~550 MB) — installed apps, system changes
│   ├── sdcard.img      # SD card data (2 GB sparse)
│   ├── metadata.img    # Metadata partition (64 MB)
│   ├── pflash.img      # UEFI/bootloader vars (3 MB)
│   ├── uboot_env.img   # U-Boot environment (72 KB)
│   ├── misc.img        # Misc partition (1 MB)
│   └── ap_overlay.img  # AP overlay (10 MB)
├── super.img → (symlink to base)
├── boot.img → (symlink to base)
└── ... other base image symlinks
```

**S3 structure:**

```
s3://{bucket}/snapshots/{orgId}/{snapshotName}/
├── manifest.json
├── instance_data/overlay.img
├── instance_data/sdcard.img
├── instance_data/metadata.img
├── ... (all files, symlinks resolved to real content)
├── super.img        # Full base image (~1.9 GB)
├── boot.img         # Kernel (64 MB)
└── ... all base images (self-contained for cross-runner restore)
```

**Known snapshot limitations:**

| Limitation | Cause | Impact |
|-----------|-------|--------|
| No memory state capture | crosvm virtio-fs doesn't support `virtio_sleep` | VM must cold-boot from snapshot (no instant restore) |
| CVD overlay path-specific | Overlays reference assembly paths internally | Overlay files backed up for DR but can't directly replace on new instance |
| Cold boot on restore | Custom snapshots launch from base images | ~60-120s boot time, same as fresh VM |
| No `--snapshot_compat` | `cvd_internal_start` doesn't recognize the flag | Can't use CVD's built-in snapshot_take/restore |

### Disk Management

- [x] CoW overlay images (only deltas stored per instance, ~570 MiB per VM)
- [x] Shared base artifacts (3.2 GiB shared across all VMs)
- [x] Configurable virtual disk size (default 20 GiB)

## 🚧 Partially Implemented

### Snapshots & Images

- [x] Pull snapshot from S3 (downloads if not present locally)
- [x] Push snapshot to S3 (auto-uploads on create)
- [ ] Restore VM state from snapshot disk files (overlay replacement after CVD assembly)
- [ ] Build snapshot from config (stub)
- [ ] Tag image (stub)

### SSH Gateway

- [x] SSH gateway service structure
- [ ] Full SSH-to-ADB bridging (partially implemented)

### Memory Ballooning

- [x] `crosvm_use_balloon: true` enabled in Cuttlefish config
- [x] `--balloon` flag passed to crosvm
- [ ] **Guest driver not available** — `CONFIG_VIRTIO_BALLOON=m` but `.ko` not shipped in stock images
- [ ] Balloon control via `crosvm balloon <bytes> <socket>` — works host-side but no guest response
- [ ] Requires custom Android kernel build with `CONFIG_VIRTIO_BALLOON=y`

## ❌ Not Yet Implemented

### High Priority

- [ ] **Warm snapshot restore** (blocked by crosvm virtio-fs `virtio_sleep` limitation)
- [ ] **Overlay restore from snapshot** (replace CVD overlays with backed-up files post-assembly)
- [ ] Memory ballooning (requires custom kernel with `CONFIG_VIRTIO_BALLOON=y`)

### Performance Optimizations

- [ ] VM pool (pre-warmed instances for instant creation)
- [ ] Snapshot restore (fast boot from memory state)
- [ ] Huge pages support

### Advanced Features

- [ ] Live resize (CPU/memory hotplug)
- [ ] GPU passthrough (Cuttlefish supports GPU acceleration)
- [ ] Live migration
- [ ] Nested virtualization

## Configuration

### Environment Variables

```bash
# Required
SERVER_URL=http://localhost:3000      # Daytona API URL
API_TOKEN=<token>                      # Runner API token

# Cuttlefish Host (remote mode only)
# Leave CVD_SSH_HOST empty for local mode
CVD_SSH_HOST=root@40.160.30.187
CVD_SSH_KEY_PATH=/path/to/id_rsa

# Optional
API_PORT=3107                          # Runner API port
LOG_LEVEL=info                         # debug, info, warn, error
RUNNER_DOMAIN=127.0.0.1                # Runner domain for API registration

# VM Defaults
CVD_DEFAULT_CPUS=4                     # vCPUs per VM
CVD_DEFAULT_MEMORY_MB=4096             # Memory per VM (MB)
CVD_DEFAULT_DISK_GB=20                 # Virtual disk size (GB)

# Cuttlefish Paths
CVD_INSTANCES_PATH=/var/lib/cuttlefish/instances
CVD_ARTIFACTS_PATH=/var/lib/cuttlefish/artifacts
CVD_HOME=/home/vsoc-01
CVD_PATH=/usr/bin/cvd
CVD_ADB_PATH=/usr/lib/cuttlefish-common/bin/adb
CVD_ADB_BASE_PORT=6520
CVD_MAX_INSTANCES=100

# SSH Gateway (optional)
SSH_GATEWAY_ENABLE=false

# S3 Configuration (for snapshot push/pull)
AWS_ENDPOINT_URL=https://s3.us-east-2.amazonaws.com
AWS_ACCESS_KEY_ID=<key>
AWS_SECRET_ACCESS_KEY=<secret>
AWS_REGION=us-east-2
AWS_DEFAULT_BUCKET=<bucket-name>
```

### Host Requirements

```bash
# Cuttlefish packages
sudo apt install cuttlefish-base cuttlefish-user

# User setup
sudo usermod -aG kvm,cvdnetwork,render vsoc-01

# TAP interface provisioning (default: 10, increase for more VMs)
# Edit /etc/default/cuttlefish-host-resources
num_cvd_accounts=20
sudo systemctl restart cuttlefish-host-resources

# Verify
cvd version
adb devices
```

### Directory Structure (Host)

```
/var/lib/cuttlefish/
├── artifacts/             # Cuttlefish system images
│   └── android-LATEST/    # Base image (symlink to cf_vm)
├── instances/             # Runner instance data
│   ├── mappings.json      # sandboxId ↔ instanceNum mappings
│   └── <sandbox-id>/      # Per-sandbox data
│       └── instance.json  # Instance configuration
└── snapshots/             # Custom snapshots (orgId/name)

/var/tmp/cvd/<uid>/        # CVD runtime (managed by cvd)
├── <run-id>/
│   ├── home/cuttlefish/
│   │   ├── assembly/      # Assembled config
│   │   └── instances/cvd-N/
│   │       ├── overlay.img   # CoW overlay (~545 MiB)
│   │       ├── sdcard.img    # SD card (2 GiB sparse)
│   │       └── logs/         # Instance logs
│   └── artifacts/         # Symlinks to base images
└── instance_database.binpb  # CVD instance registry

/tmp/cf_avd_<uid>/cvd-N/internal/
└── crosvm_control.sock    # crosvm control socket
```

## Resource Limits

### Per-VM Resource Usage

| Resource | Allocated | Actual Host Usage |
|----------|-----------|-------------------|
| CPU | 4 vCPUs (configurable) | Shared with host |
| Memory | 4 GiB (configurable) | ~4 GiB RSS per crosvm process |
| Disk (runtime) | 20 GiB virtual | ~570 MiB actual (CoW overlay) |
| Disk (shared base) | — | 3.2 GiB (shared across all VMs) |
| Network | 2 TAP interfaces per VM | etap + mtap |
| ADB port | 1 per VM (6520 + N-1) | — |
| vsock CID | 1 per VM | 32-bit space |

### Host Scaling Limits

| Limit | Default | How to Increase |
|-------|---------|----------------|
| **TAP interfaces** | 10 VMs | Set `num_cvd_accounts=N` in `/etc/default/cuttlefish-host-resources` |
| **RAM** | ~6-7 VMs (at 4 GiB each on 125 GiB host) | Add RAM, reduce per-VM allocation |
| **CPU** | 8 VMs (at 4 vCPUs on 32 cores, no overcommit) | CPU overcommit works for idle VMs |
| **Disk** | Hundreds of VMs | Not a practical concern |

## Known Issues

1. **CVD "Cancelled" status** — After runner restart, `cvd fleet` shows "Cancelled" for running VMs. The multi-layered health check handles this by falling back to ADB and process detection.
2. **Memory ballooning not functional** — Guest kernel has `CONFIG_VIRTIO_BALLOON=m` but the `.ko` module is not shipped in stock Cuttlefish images. Requires custom kernel build.
3. **crosvm balloon control fragile** — `crosvm balloon_stats` can hang or kill VMs if guest driver is not loaded.
4. **Operator device registration stale** — After `cvd rm`, the cuttlefish-operator may retain stale device entries. Runner proactively cleans these.
5. **No idle timeout in Cuttlefish** — VMs run indefinitely; no built-in auto-shutdown for idle devices.
6. **WebRTC vsock connection resets** — Occasional `vsock_connection.cpp: Failed to connect: Connection reset by peer` in WebRTC logs, causing temporary display disconnects.
7. **Snapshot restore is cold-boot only** — Custom snapshots create a fresh VM from the base images (cold boot). The backed-up disk files (overlay, sdcard) are stored in `instance_data/` for disaster recovery but not yet automatically restored into the new instance. This is because CVD overlays are path-specific and get wiped when base image paths change.
8. **crosvm virtio-fs blocks suspend/snapshot** — `cvd snapshot_take` and `cvd suspend` fail with `virtio_sleep not implemented for virtio-fs`. This blocks CVD's native memory+disk snapshot mechanism. A newer crosvm version may fix this.
9. **Stale CVD sockets cause boot failures** — Failed `cvd create` attempts leave stale sockets in `/tmp/cf_avd_*` that cause subsequent `run_cvd` processes to crash with `Failed to read a complete exit code`. Cleaning `/tmp/cf_avd_*` and `/tmp/cf_env_*` resolves this.

## Comparison with runner-ch

| Feature | runner-ch (Cloud Hypervisor) | runner-android (Cuttlefish) |
|---------|-----------------------------|-----------------------------|
| Hypervisor | Cloud Hypervisor | crosvm (via CVD) |
| OS Support | Linux only | Android only |
| Display | VNC/SPICE (planned) | ✅ WebRTC (built-in) |
| Shell Access | SSH | ADB |
| GPU Passthrough | 🚧 (untested) | ✅ (Cuttlefish GPU 2D) |
| Memory Ballooning | ✅ | ❌ (guest driver missing) |
| Live Fork | ✅ (local mode) | ❌ |
| Warm Snapshots | ✅ (~4s restore) | ❌ (blocked by virtio-fs) |
| Disk Snapshots to S3 | ✅ | ✅ (overlay + sdcard + metadata) |
| Boot Time | ~5-10s (cold), ~4s (warm) | ~60-120s |
| Memory Overhead | Lower | Higher (~18 GiB RSS for 4 GiB VM) |
| Disk Format | qcow2 (CoW) | qcow2-like (CVD overlay) |
| Per-VM Disk | ~20 MiB overlay | ~570 MiB overlay |
| Computer Use | ✅ (via daemon) | ✅ (via ADB + WebRTC) |
| File Operations | ✅ (via daemon) | ✅ (via ADB push/pull) |
| Process Execution | ✅ (via daemon) | ✅ (via ADB shell) |

## Architecture

### VM Management Flow

```
Daytona API → Runner API (port 3107) → CVD Client → cvd CLI → crosvm
                  │                         │
                  │                         ├── ADB Client (shell, push, pull)
                  │                         ├── Health Monitor (CVD + ADB + process)
                  │                         └── Instance Mapper (sandboxId ↔ instanceNum)
                  │
                  ├── WebRTC Proxy → Cuttlefish Operator (port 1443)
                  └── Toolbox Handler → ADB Shell
```

### WebRTC Proxy Flow

```
Browser → Proxy (port 4000) → Runner API (port 3107) → Cuttlefish Operator (port 1443)
   │                                │
   │  6080-{sandboxId}.proxy.localhost:4000/
   │                                │
   │  ┌─────────────────────────────┘
   │  │
   │  ├── / → Fetch client.html, inject device ID + base tag
   │  ├── /js/server_connector.js → Modify deviceId() to use injected ID
   │  ├── /js/*.js, /style.css → Proxy to operator with device path prefix
   │  ├── /infra_config → Pass through to operator
   │  └── /polled_connections → Pass through to operator (WebSocket)
```

### Health Check Flow

```
Health Monitor (every 30s)
    │
    ├── Layer 1: cvd fleet --json
    │   ├── "Running" → ✅ Trust it
    │   ├── "Stopped" → ✅ Trust it
    │   └── "Cancelled" or missing → ⚠️ Don't trust, check Layer 2
    │
    ├── Layer 2: adb shell getprop sys.boot_completed
    │   ├── "1" → ✅ VM is alive (CVD metadata is stale)
    │   └── No response → Check Layer 3
    │
    └── Layer 3: ps -eo args | grep crosvm.*cvd-N
        ├── Found → ⚠️ VM is booting (ADB not ready yet)
        └── Not found → ❌ VM is truly dead, report crash
```

---

_Last updated: 2026-02-08_
