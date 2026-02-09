# Cuttlefish Server Setup for Daytona Android Runner

This document describes the configuration required on the Cuttlefish host server to enable ADB and WebRTC connectivity for Android sandboxes.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              User's Machine                                  │
│  ┌─────────────┐     ┌─────────────┐                                        │
│  │ Android     │     │ Web Browser │                                        │
│  │ Studio/ADB  │     │ (WebRTC)    │                                        │
│  └──────┬──────┘     └──────┬──────┘                                        │
│         │                   │                                                │
│         │ SSH Tunnel        │ HTTPS                                         │
│         │ (port 2222)       │ (proxy.localhost:4000)                        │
└─────────┼───────────────────┼───────────────────────────────────────────────┘
          │                   │
          ▼                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Daytona Platform                                   │
│  ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐       │
│  │ Main SSH        │     │ Proxy           │     │ Runner-Android  │       │
│  │ Gateway (:2222) │────▶│ (:4000)         │────▶│ API (:3007)     │       │
│  └────────┬────────┘     └─────────────────┘     │ SSH GW (:2220)  │       │
│           │                                       └────────┬────────┘       │
│           │ SSH (direct-tcpip)                             │                │
│           └────────────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ SSH (port forwarding)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Cuttlefish Host Server                               │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    Host Orchestrator (operator)                       │  │
│  │                    Port 1443 (HTTPS) - WebRTC UI & Signaling         │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌──────────────────┐   │
│  │ CVD Instance 1      │  │ CVD Instance 2      │  │ CVD Instance N   │   │
│  │ ADB: 6520           │  │ ADB: 6521           │  │ ADB: 6520+N-1    │   │
│  │ WebRTC: 8443        │  │ WebRTC: 8444        │  │ WebRTC: 8443+N-1 │   │
│  │ Device: cvd_1-1-1   │  │ Device: cvd_1-2-2   │  │ Device: cvd_1-N-N│   │
│  └─────────────────────┘  └─────────────────────┘  └──────────────────┘   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    Kernel Modules (vsock)                             │  │
│  │  vsock, vhost_vsock, vsock_loopback, vmw_vsock_virtio_transport      │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Required Kernel Modules

The Cuttlefish host requires these kernel modules for proper vsock communication:

```bash
# Load required modules
modprobe vsock
modprobe vhost_vsock
modprobe vsock_loopback
modprobe vmw_vsock_virtio_transport_common

# Optional: for /proc/net/vsock diagnostics
modprobe vsock_diag

# Verify modules are loaded
lsmod | grep vsock
```

**Expected output:**

```
vsock_loopback         12288  0
vhost_vsock            24576  0
vmw_vsock_virtio_transport_common    57344  2 vhost_vsock,vsock_loopback
vsock                  61440  3 vmw_vsock_virtio_transport_common,vhost_vsock,vsock_loopback
```

### Persistent Module Loading

Add to `/etc/modules-load.d/cuttlefish.conf`:

```
vsock
vhost_vsock
vsock_loopback
vmw_vsock_virtio_transport_common
```

## Host Orchestrator (cuttlefish-operator)

The host orchestrator provides centralized WebRTC services for all CVD instances.

### Verify Service Status

```bash
systemctl status cuttlefish-operator
```

### Service Configuration

The operator listens on:

- **Port 1443 (HTTPS)**: WebRTC web interface and signaling
- **Port 1080 (HTTP)**: Health checks
- **Unix Socket**: `/run/cuttlefish/operator` (for CVD registration)

### Start/Restart Service

```bash
sudo systemctl restart cuttlefish-operator
```

## ADB Connection Flow

### Port Assignment

Each CVD instance gets a unique ADB port:

- Instance 1: `6520`
- Instance 2: `6521`
- Instance N: `6520 + N - 1`

### Connection Path

```
User's ADB Client
    │
    │ SSH tunnel: ssh -L 5555:localhost:<adb_port> -p 2222 adb-<sandbox>-<token>@localhost
    ▼
Main SSH Gateway (port 2222)
    │
    │ Authenticates token, identifies ADB request from "adb-" prefix
    ▼
Runner SSH Gateway (port 2220)
    │
    │ Public key authentication, forwards direct-tcpip
    ▼
Cuttlefish Host
    │
    │ localhost:<adb_port> (e.g., 6524)
    ▼
CVD Instance ADB Server
```

### SSH Tunnel Command Format

```bash
ssh -L <local_port>:localhost:<remote_adb_port> \
    -p 2222 \
    adb-<sandbox_id>-<access_token>@<gateway_host> \
    -N
```

Example:

```bash
ssh -L 5555:localhost:6524 -p 2222 adb-3df8183d-e7ac-4ac4-8a78-a3843ba25a8b-ABC123@localhost -N
```

### ADB Connection

```bash
adb kill-server
adb start-server
adb connect localhost:5555
adb devices
```

## WebRTC Connection Flow

### Port Assignment

- **Host Orchestrator**: Port 1443 (HTTPS) - serves all devices
- **Per-instance WebRTC**: Port 8443 + instance_num (internal signaling)

### Device ID Format

```
cvd_<group>-<instance>-<instance>
```

Example: `cvd_1-5-5` for instance 5 in group 1

### Connection Path

```
User's Web Browser
    │
    │ HTTPS: http://<port>-<sandbox_id>.proxy.localhost:4000/devices/cvd_1-N-N/files/client.html
    ▼
Daytona Proxy (port 4000)
    │
    │ Routes based on sandbox ID to runner
    ▼
Runner-Android API (port 3007)
    │
    │ Proxies WebRTC traffic to Cuttlefish host
    ▼
Host Orchestrator (port 1443)
    │
    │ Serves WebRTC UI and handles signaling
    ▼
CVD Instance WebRTC Process
    │
    │ vsock connection for video/audio streams
    ▼
Android Display Output
```

### WebRTC URL Format

```
http://6080-<sandbox_id>.proxy.localhost:4000/devices/cvd_1-<N>-<N>/files/client.html
```

Example:

```
http://6080-3df8183d-e7ac-4ac4-8a78-a3843ba25a8b.proxy.localhost:4000/devices/cvd_1-5-5/files/client.html
```

## Troubleshooting

### ADB Connection Issues

1. **Check if CVD is running:**

   ```bash
   cvd fleet --json
   ```

2. **Verify ADB port is listening:**

   ```bash
   netstat -tlnp | grep 652
   ```

3. **Test ADB directly on host:**

   ```bash
   adb connect localhost:<port>
   adb devices
   ```

4. **Check SSH tunnel is working:**
   - No errors in SSH command output
   - `nc -zv localhost 5555` succeeds on local machine

### WebRTC Connection Issues

1. **Check vsock modules:**

   ```bash
   lsmod | grep vsock
   ```

2. **Check host orchestrator:**

   ```bash
   systemctl status cuttlefish-operator
   curl -k https://localhost:1443/devices
   ```

3. **Check CVD WebRTC process:**

   ```bash
   tail -f /var/tmp/cvd/*/home/cuttlefish/instances/cvd-*/logs/launcher.log | grep -i webrtc
   ```

4. **Common error - vsock connection failed:**

   ```
   vsock_connection.cpp:222] Failed to connect:No such device
   ```

   **Fix:** Load vsock_loopback module:

   ```bash
   modprobe vsock_loopback
   ```

### CVD Instance Crashes

1. **Check launcher logs:**

   ```bash
   tail -100 /var/tmp/cvd/*/home/cuttlefish/instances/cvd-*/logs/launcher.log
   ```

2. **Check kernel logs:**

   ```bash
   dmesg | tail -50
   ```

3. **Restart with fresh state:**

   ```bash
   cvd stop --instance_nums=<N>
   cvd start --instance_nums=<N>
   ```

## Environment Variables (Runner)

The runner-android requires these environment variables for remote Cuttlefish:

```bash
# Remote CVD host
CVD_SSH_HOST=root@<cuttlefish_host_ip>
CVD_SSH_KEY_PATH=/path/to/ssh/private_key

# Cuttlefish paths on remote host
CVD_HOME=/home/vsoc-01
CVD_PATH=/home/vsoc-01/bin/cvd
CVD_INSTANCES_PATH=/var/lib/cuttlefish/instances
CVD_ARTIFACTS_PATH=/var/lib/cuttlefish/artifacts

# Port configuration
CVD_ADB_BASE_PORT=6520
CVD_WEBRTC_BASE_PORT=8443

# SSH Gateway (for ADB tunneling)
SSH_GATEWAY_ENABLE=true
SSH_PUBLIC_KEY=<base64_encoded_public_key>

# Runner registration
DOMAIN=localhost  # or runner's accessible domain
```

## Android Snapshots (System Images)

Cuttlefish supports multiple Android versions and form factors. Each snapshot is a directory of system images that CVD uses to boot a virtual device.

### Naming Convention

```
android_{version}_{formfactor}
```

### Available Snapshots

| Snapshot Name | Android | Form Factor | AOSP Branch |
|---------------|---------|-------------|-------------|
| `android_15_phone` | 15 (Baklava) | Phone 720×1280 | `aosp-main` |
| `android_14_phone` | 14 (UpsideDownCake) | Phone 720×1280 | `aosp-android14-gsi` |
| `android_13_phone` | 13 (Tiramisu) | Phone 720×1280 | `aosp-android13-gsi` |

Legacy aliases: `cf_vm` and `android_baklava` both point to `android_15_phone`.

**Note:** Only phone form factor is available through the public Android Build API.
Tablet, TV, and Wear OS targets require Google-internal API access or building from AOSP source.

### Fetching Snapshots

Each snapshot is ~2-3 GB. Use `cvd fetch` to download from Google's CI:

```bash
#!/bin/bash
# Run as root on the Cuttlefish host

ARTIFACTS="/var/lib/cuttlefish/artifacts"
CVD_HOME="/home/vsoc-01"

# Android 15 (latest) — uses the default phone target from aosp-main
echo "=== Fetching android_15_phone ==="
if [ ! -d "$ARTIFACTS/android-LATEST" ] || [ ! -f "$ARTIFACTS/android-LATEST/super.img" ]; then
  mkdir -p "$ARTIFACTS/android-LATEST"
  chown vsoc-01:vsoc-01 "$ARTIFACTS/android-LATEST"
  su - vsoc-01 -c "cvd fetch --default_build=aosp-main --target_directory=$ARTIFACTS/android-LATEST"
fi
ln -sf "$ARTIFACTS/android-LATEST" "$ARTIFACTS/android_15_phone"
ln -sf "$ARTIFACTS/android_15_phone" "$CVD_HOME/android_15_phone"
ln -sf "$ARTIFACTS/android_15_phone" "$CVD_HOME/cf_vm"
echo "DONE: android_15_phone"

# Android 14
echo "=== Fetching android_14_phone ==="
if [ ! -f "$ARTIFACTS/android_14_phone/super.img" ]; then
  mkdir -p "$ARTIFACTS/android_14_phone"
  chown vsoc-01:vsoc-01 "$ARTIFACTS/android_14_phone"
  su - vsoc-01 -c "cvd fetch --default_build=aosp-android14-gsi --target_directory=$ARTIFACTS/android_14_phone"
fi
ln -sf "$ARTIFACTS/android_14_phone" "$CVD_HOME/android_14_phone"
echo "DONE: android_14_phone"

# Android 13
echo "=== Fetching android_13_phone ==="
if [ ! -f "$ARTIFACTS/android_13_phone/super.img" ]; then
  mkdir -p "$ARTIFACTS/android_13_phone"
  chown vsoc-01:vsoc-01 "$ARTIFACTS/android_13_phone"
  su - vsoc-01 -c "cvd fetch --default_build=aosp-android13-gsi --target_directory=$ARTIFACTS/android_13_phone"
fi
ln -sf "$ARTIFACTS/android_13_phone" "$CVD_HOME/android_13_phone"
echo "DONE: android_13_phone"

# Fix ownership on all symlinks
chown -h vsoc-01:vsoc-01 $CVD_HOME/android_* $CVD_HOME/cf_vm 2>/dev/null

echo ""
echo "=== All snapshots ==="
ls -la $CVD_HOME/android_* $CVD_HOME/cf_vm 2>/dev/null
```

### Verifying Snapshots

```bash
# Check which snapshots the runner can see
curl -s -H "Authorization: Bearer <token>" http://localhost:3107/snapshots/exists?name=android_15_phone

# List all symlinks
ls -la /home/vsoc-01/android_*
```

### Disk Space Requirements

| Snapshot | Size |
|----------|------|
| `android_15_phone` | ~3.2 GB |
| `android_14_phone` | ~2.4 GB |
| `android_13_phone` | ~14 GB (includes host tools) |
| Per-VM runtime | ~570 MB |
| **Total** | **~20 GB** |

### Notes

- Each `cvd fetch` downloads from `ci.android.com` and takes 5-15 minutes
- The `--target_directory` must be writable by `vsoc-01`
- Snapshots are pure AOSP (no Google Play Services)
- Only the phone form factor is available publicly; other form factors (tablet, TV, wear) require Google-internal API access
- To verify a snapshot works: `cvd create --product_path=/path/to/snapshot --instance_nums=99`

## Security Notes

1. **SSH Key Authentication**: The main SSH gateway authenticates to the runner's SSH gateway using public key authentication. Keys must match.

2. **Token-based Access**: ADB access requires a valid token embedded in the SSH username (`adb-<sandbox_id>-<token>`).

3. **Self-signed Certificates**: Cuttlefish uses self-signed certificates for WebRTC. The runner proxy is configured to skip TLS verification when connecting to the host.

4. **Network Isolation**: CVD instances are isolated by port. Each instance has unique ADB and WebRTC ports.
