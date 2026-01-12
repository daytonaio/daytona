# Windows VM Snapshots

This document explains how VM snapshots work in the Daytona Windows runner, including how they're stored, pushed, and pulled.

## Overview

The snapshot system provides a way to:

- Create reusable base images from configured sandboxes
- Store snapshots in an S3-compatible object store
- Distribute base images across multiple runner hosts
- Enable fast sandbox creation from pre-configured templates

## Architecture

The runner can connect to libvirt either locally or via SSH to a remote host. Snapshot operations automatically handle both scenarios.

### Local Libvirt Connection

When the runner and libvirt are on the same machine (`LIBVIRT_URI=qemu:///system`):

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Runner + Libvirt Host (same machine)                  │
├─────────────────────────────────────────────────────────────────────────┤
│  Runner Process                                                          │
│  ├── S3 Client ──────────────────────────────▶ S3 Store                 │
│  └── Libvirt Client ─────────────────────────▶ Local QEMU/KVM           │
│                                                                          │
│  /var/lib/libvirt/                                                      │
│  ├── snapshots/                    ← Base images (golden templates)     │
│  │   ├── winserver-autologin-base.qcow2                                │
│  │   └── myapp-v1.0.qcow2          ← Custom snapshots                  │
│  ├── sandboxes/                    ← Per-sandbox overlay disks          │
│  │   ├── sandbox-001.qcow2         ← Copy-on-write overlay              │
│  │   └── sandbox-002.qcow2                                              │
│  └── qemu/                                                              │
│      ├── nvram/                    ← UEFI variables                     │
│      └── save/                     ← Memory snapshots (managed save)    │
└─────────────────────────────────────────────────────────────────────────┘
```

### Remote Libvirt Connection (Development/Distributed Setup)

When the runner connects to a remote libvirt host via SSH (`LIBVIRT_URI=qemu+ssh://root@host/system`):

```
┌─────────────────────────────┐         ┌─────────────────────────────────┐
│      Runner Machine          │         │      Remote Libvirt Host         │
├─────────────────────────────┤   SSH   ├─────────────────────────────────┤
│                             │◀───────▶│                                  │
│  Runner Process             │         │  /var/lib/libvirt/               │
│  ├── S3 Client              │         │  ├── snapshots/                  │
│  │   └── Upload/Download    │         │  │   └── *.qcow2                 │
│  └── Libvirt Client         │         │  ├── sandboxes/                  │
│      └── SSH Tunnel ────────┼────────▶│  │   └── *.qcow2                 │
│                             │         │  └── qemu/                       │
│  File streaming via SSH:    │         │      ├── nvram/                  │
│  - Push: remote → runner    │         │      └── save/                   │
│  - Pull: runner → remote    │         │                                  │
│                             │         │  QEMU/KVM processes              │
└─────────────────────────────┘         └─────────────────────────────────┘
              │
              │ S3 API
              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    S3-Compatible Object Store                            │
├─────────────────────────────────────────────────────────────────────────┤
│  bucket: daytona                                                        │
│  └── snapshots/                                                         │
│      ├── winserver-autologin-base.qcow2                                │
│      └── myapp-v1.0.qcow2                                               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Data Flow for Remote Hosts

**Push (creating snapshot from sandbox):**

1. `qemu-img convert` runs on remote host via SSH
2. Flattened image is streamed FROM remote host TO runner via `ssh cat`
3. Runner uploads to S3

**Pull (downloading snapshot to host):**

1. Runner downloads from S3
2. Image is streamed FROM runner TO remote host via `ssh cat >`
3. Permissions set on remote host via SSH

## Snapshot Types

### 1. Disk Snapshots (qcow2 images)

Disk snapshots are complete VM disk images in qcow2 format:

- **Base images**: Read-only golden templates stored in `/var/lib/libvirt/snapshots/`
- **Overlay disks**: Per-sandbox copy-on-write disks in `/var/lib/libvirt/sandboxes/`

When a sandbox is created, an overlay disk is created that uses the base image as a backing file. Only changes are stored in the overlay, making sandbox creation fast and storage-efficient.

### 2. Memory Snapshots (Managed Save)

Memory snapshots preserve the complete VM state (RAM + CPU registers):

- **Pause/Resume**: Keeps memory in host RAM (fast, but uses RAM)
- **SuspendToDisk/ResumeFromDisk**: Saves memory to disk (slower, frees RAM)

Memory snapshots are managed automatically by libvirt and stored in `/var/lib/libvirt/qemu/save/`.

## Configuration

### Libvirt Connection

The runner connects to libvirt using the `LIBVIRT_URI` environment variable:

```bash
# Local connection (runner and libvirt on same machine)
LIBVIRT_URI=qemu:///system

# Remote connection via SSH (runner connects to remote libvirt host)
LIBVIRT_URI=qemu+ssh://root@libvirt-host.example.com/system
```

For remote connections, ensure:

- SSH key-based authentication is configured (no password prompts)
- The remote user has access to libvirt and the snapshots/sandboxes directories
- `qemu-img` is installed on the remote host

### S3 Configuration

The snapshot store uses an S3-compatible object store. Configure it in the runner environment:

```bash
# S3-compatible storage configuration
AWS_ENDPOINT_URL=http://minio:9000    # S3 endpoint (MinIO, AWS S3, etc.)
AWS_REGION=us-east-1                   # Region
AWS_ACCESS_KEY_ID=your-access-key      # Access key
AWS_SECRET_ACCESS_KEY=your-secret-key  # Secret key
AWS_DEFAULT_BUCKET=daytona             # Bucket name
```

Snapshots are stored under the `snapshots/` prefix in the bucket:

```
s3://daytona/snapshots/winserver-autologin-base.qcow2
s3://daytona/snapshots/myapp-v1.0.qcow2
```

## Push Snapshot

Push creates a new snapshot from an existing sandbox and uploads it to the store.

### API Endpoint

```
POST /snapshots/push
```

### Request Body

```json
{
  "sandboxId": "sandbox-001",
  "snapshotName": "myapp-v1.0"
}
```

### Response

```json
{
  "snapshotName": "myapp-v1.0",
  "snapshotPath": "snapshots/myapp-v1.0.qcow2",
  "sizeBytes": 15032385536
}
```

### Process

1. **Validate State**: Sandbox must be stopped or paused
2. **Flatten Disk**: Convert overlay disk to standalone qcow2 using `qemu-img convert`
3. **Upload**: Stream the flattened image to S3
4. **Cleanup**: Remove temporary files

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Sandbox Disk   │     │  Flatten        │     │  Upload to S3   │
│  (overlay)      │────▶│  qemu-img       │────▶│                 │
│                 │     │  convert        │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │
        ▼                       ▼
   base.qcow2 ◀── backing ── overlay.qcow2
        │
        └── Changes merged into standalone image
```

### Requirements

- Sandbox must be **stopped** or **paused** for a consistent snapshot
- Sufficient disk space for the flattened image (temporary)
- S3 credentials must be configured

### Example: Create a Custom Base Image

```bash
# 1. Create a sandbox and configure it
curl -X POST http://runner:8080/sandboxes \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id": "config-vm", "snapshot": "winserver-autologin-base", ...}'

# 2. Configure the VM (install software, etc.)
# ... (via VNC or remote desktop)

# 3. Stop the sandbox
curl -X POST http://runner:8080/sandboxes/config-vm/stop \
  -H "Authorization: Bearer $TOKEN"

# 4. Push as a new snapshot
curl -X POST http://runner:8080/snapshots/push \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"sandboxId": "config-vm", "snapshotName": "myapp-configured-v1.0"}'

# 5. Destroy the config VM
curl -X POST http://runner:8080/sandboxes/config-vm/destroy \
  -H "Authorization: Bearer $TOKEN"
```

## Pull Snapshot

Pull downloads a snapshot from the store to the local snapshots directory.

### API Endpoint

```
POST /snapshots/pull
```

### Request Body

```json
{
  "snapshot": "myapp-v1.0"
}
```

### Process

1. **Check Local**: Skip if snapshot already exists and is valid
2. **Verify Remote**: Confirm snapshot exists in store
3. **Download**: Stream from S3 to temporary file
4. **Validate**: Verify with `qemu-img check`
5. **Finalize**: Move to snapshots directory with proper permissions

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  S3 Store       │     │  Download       │     │  Local          │
│  snapshots/     │────▶│  + Validate     │────▶│  /var/lib/      │
│  myapp-v1.0     │     │                 │     │  libvirt/       │
│  .qcow2         │     │                 │     │  snapshots/     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### Features

- **Progress logging**: Shows download progress every 100MB
- **Atomic download**: Uses temp file to prevent partial/corrupted files
- **Validation**: Verifies image integrity with `qemu-img check`
- **Idempotent**: Skips download if valid local copy exists
- **Auto-repair**: Re-downloads if local copy fails validation

### Example: Distribute Snapshot to Multiple Hosts

```bash
# On each runner host, pull the snapshot
curl -X POST http://runner1:8080/snapshots/pull \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"snapshot": "myapp-v1.0"}'

curl -X POST http://runner2:8080/snapshots/pull \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"snapshot": "myapp-v1.0"}'
```

## Create Snapshot

Create Snapshot is similar to Push Snapshot but can work with **running VMs**. It supports two modes:

- **Safe mode** (`live: false`, default): Pauses the VM briefly for a consistent snapshot, then resumes it before uploading to S3
- **Optimistic mode** (`live: true`): Takes the snapshot without pausing using `qemu-img --force-share`

### API Endpoint

```
POST /snapshots/create
```

### Request Body

```json
{
  "sandboxId": "sandbox-001",
  "name": "myapp-v1.0",
  "live": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `sandboxId` | string | Yes | ID of the sandbox to snapshot |
| `name` | string | Yes | Name for the snapshot |
| `live` | boolean | No | `false` (default) = pause VM for consistency; `true` = optimistic, no pause |

### Response

```json
{
  "name": "myapp-v1.0",
  "snapshotPath": "snapshots/myapp-v1.0.qcow2",
  "sizeBytes": 15032385536,
  "liveMode": false
}
```

### Process

```
                              ┌─────────────────────────────────────────┐
                              │           CreateSnapshot                 │
                              │                                          │
                              │  live=false (default)    live=true       │
                              └──────────┬─────────────────┬─────────────┘
                                         │                 │
                    ┌────────────────────▼───┐    ┌───────▼────────────────┐
                    │     SAFE MODE          │    │   OPTIMISTIC MODE      │
                    ├────────────────────────┤    ├────────────────────────┤
                    │ 1. Pause VM            │    │ 1. VM keeps running    │
                    │ 2. qemu-img convert    │    │ 2. qemu-img convert    │
                    │ 3. Resume VM           │    │    --force-share       │
                    │ 4. Upload to S3        │    │ 3. Upload to S3        │
                    └────────────────────────┘    └────────────────────────┘
                              │                              │
                              │     Guaranteed consistent    │  May be inconsistent
                              │     ~30-60s downtime         │  Zero downtime
                              └──────────────────────────────┘
```

### Mode Comparison

| Parameter | VM Downtime | Consistency | Use Case |
|-----------|-------------|-------------|----------|
| `live: false` | ~30-60s pause | Guaranteed | Production snapshots, base images |
| `live: true` | None | Best-effort | Quick dev snapshots, read-mostly VMs |

### Warning: Live Mode Consistency

When using `live: true`, the snapshot is taken while the VM is running. This uses `qemu-img convert --force-share` which reads the disk without exclusive access.

**Risks of live mode:**

- If the VM is actively writing to disk, the snapshot may be inconsistent
- File system may be in an unclean state
- Applications with open files may have incomplete data

**When live mode is safe:**

- VM is mostly idle (read-only workloads)
- Quick development snapshots where consistency is not critical
- You plan to validate the snapshot before using it as a base image

**Recommendation:** Use `live: false` (default) for production snapshots and base images.

### Example: Snapshot a Running VM

```bash
# Safe mode (default) - pauses VM briefly
curl -X POST http://runner:8080/snapshots/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "sandboxId": "my-running-sandbox",
    "name": "myapp-checkpoint-v1"
  }'

# Optimistic mode - no pause, may be inconsistent
curl -X POST http://runner:8080/snapshots/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "sandboxId": "my-running-sandbox",
    "name": "myapp-quick-snapshot",
    "live": true
  }'
```

### Difference from Push Snapshot

| Feature | Push Snapshot | Create Snapshot |
|---------|---------------|-----------------|
| Requires stopped/paused VM | Yes | No |
| Auto-pauses running VM | No (fails) | Yes (in safe mode) |
| Auto-resumes after flatten | No | Yes |
| Live/optimistic mode | No | Yes (`live: true`) |
| Use case | Manual snapshot of stopped VM | Snapshot running VM with minimal downtime |

## Using Custom Snapshots

Once a snapshot is available locally, sandboxes can be created from it by specifying the snapshot name:

```bash
# Create sandbox from custom snapshot
curl -X POST http://runner:8080/sandboxes \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "my-sandbox",
    "snapshot": "myapp-v1.0",
    "cpuQuota": 2,
    "memoryQuota": 4096,
    "storageQuota": 50
  }'
```

> **Note**: The snapshot must be pulled to the local snapshots directory before creating sandboxes from it.

## File Permissions

All snapshot files should have proper permissions for libvirt:

```bash
# Ownership
chown libvirt-qemu:kvm /var/lib/libvirt/snapshots/*.qcow2
chown libvirt-qemu:kvm /var/lib/libvirt/sandboxes/*.qcow2

# Permissions
chmod 644 /var/lib/libvirt/snapshots/*.qcow2
chmod 644 /var/lib/libvirt/sandboxes/*.qcow2
```

## Troubleshooting

### Push fails with "sandbox must be stopped or paused"

The sandbox must not be running when creating a snapshot:

```bash
# Stop the sandbox first
curl -X POST http://runner:8080/sandboxes/$SANDBOX_ID/stop \
  -H "Authorization: Bearer $TOKEN"

# Then push
curl -X POST http://runner:8080/snapshots/push \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"sandboxId": "...", "snapshotName": "..."}'
```

### Pull fails with "snapshot not found in store"

Verify the snapshot exists in S3:

```bash
# Using AWS CLI or mc (MinIO client)
aws s3 ls s3://daytona/snapshots/ --endpoint-url $AWS_ENDPOINT_URL

# Or with mc
mc ls myminio/daytona/snapshots/
```

### "qemu-img check failed" during pull

The downloaded file may be corrupted. Delete and retry:

```bash
rm -f /var/lib/libvirt/snapshots/myapp-v1.0.qcow2
# Retry pull
```

### Sandbox creation fails after pull

Ensure the snapshot is in the correct directory:

```bash
ls -la /var/lib/libvirt/snapshots/
# Should show:
# -rw-r--r-- 1 libvirt-qemu kvm ... myapp-v1.0.qcow2

# Verify the image
qemu-img check /var/lib/libvirt/snapshots/myapp-v1.0.qcow2
qemu-img info /var/lib/libvirt/snapshots/myapp-v1.0.qcow2
```

### Insufficient disk space during push

The push operation needs temporary space for the flattened image:

```bash
# Check available space on the libvirt host (local or remote)
df -h /tmp

# Flattened images can be large (10-50GB for Windows)
# Ensure sufficient space in /tmp on the libvirt host
```

### Remote host connection issues

If using a remote libvirt host, verify SSH connectivity:

```bash
# Test SSH connection (should not prompt for password)
ssh root@libvirt-host.example.com "echo 'SSH OK'"

# Test libvirt connection
virsh -c qemu+ssh://root@libvirt-host.example.com/system list

# Test file operations
ssh root@libvirt-host.example.com "ls -la /var/lib/libvirt/snapshots/"
```

Common issues:

- **SSH key not configured**: Set up passwordless SSH authentication
- **Permission denied**: Ensure the SSH user has access to libvirt directories
- **qemu-img not found**: Install `qemu-utils` on the remote host

### Push/Pull slow on remote hosts

Large snapshots (10-50GB) take time to stream over SSH. Monitor progress in the runner logs:

```
PullSnapshot: Downloading 'myapp-v1.0': 45.2% (4831838208 / 10685726720 bytes)
```

Tips:

- Use a fast network connection between runner and libvirt host
- Consider running the runner on the same machine as libvirt for production
- Use compression if bandwidth is limited (not currently implemented)

## Best Practices

1. **Name snapshots descriptively**: Use version numbers or dates (e.g., `myapp-v1.2.0`, `base-2025-01-12`)

2. **Keep base images minimal**: Install only essential software in base images

3. **Test snapshots before distribution**: Verify a sandbox can be created from the snapshot

4. **Clean up old snapshots**: Remove unused snapshots from both local storage and S3

5. **Use managed save for development**: Pause VMs instead of stopping them for faster resume

6. **Document snapshot contents**: Keep notes about what's installed in each snapshot

7. **For production, use local libvirt**: Running the runner on the same machine as libvirt avoids SSH streaming overhead for push/pull operations

8. **Pre-pull snapshots**: Pull commonly used snapshots to all runner hosts during deployment, not on first sandbox creation

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/snapshots/create` | POST | Create snapshot from running/stopped sandbox (supports live mode) |
| `/snapshots/push` | POST | Create snapshot from stopped/paused sandbox and upload to store |
| `/snapshots/pull` | POST | Download snapshot from store to local |
| `/snapshots/exists` | GET | Check if snapshot exists locally |
| `/snapshots/info` | GET | Get snapshot information |
| `/snapshots/remove` | POST | Remove local snapshot |

## Related Documentation

- [SETUP_NEW_HOST.md](./SETUP_NEW_HOST.md) - Setting up a new runner host
- Libvirt documentation: https://libvirt.org/
- QEMU disk images: https://qemu.readthedocs.io/en/latest/system/images.html
