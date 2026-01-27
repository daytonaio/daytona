# Cloud Hypervisor Golden Image Creation Guide

This guide documents how to create a golden base image for Cloud Hypervisor VMs, including the Daytona daemon installation.

## Prerequisites

### Runner Host Requirements

- Ubuntu 24.04 LTS (or compatible Linux distribution)
- KVM support enabled (`/dev/kvm` available)
- Root or sudo access
- At least 20GB free disk space
- Network connectivity for downloading images

### Required Packages

```bash
apt-get update
apt-get install -y \
    cloud-hypervisor \
    qemu-utils \
    libguestfs-tools \
    cloud-image-utils \
    bridge-utils \
    dnsmasq \
    wget \
    curl
```

## Directory Structure

The runner uses the following directory structure:

```
/var/lib/cloud-hypervisor/
├── firmware/           # Hypervisor firmware files
├── kernels/            # Linux kernel images
├── images/             # Base OS images (Ubuntu cloud images)
├── snapshots/          # Golden images and snapshots
└── sandboxes/          # Per-VM runtime directories
```

Create the directories:

```bash
mkdir -p /var/lib/cloud-hypervisor/{firmware,kernels,images,snapshots,sandboxes}
```

## Step 1: Download Base Ubuntu Cloud Image

Download the Ubuntu 24.04 cloud image:

```bash
cd /var/lib/cloud-hypervisor/images

wget -O ubuntu-24.04-server-cloudimg-amd64.img \
    https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img

# Verify the download
qemu-img info ubuntu-24.04-server-cloudimg-amd64.img
```

## Step 2: Build the Daemon Binary

Build the daemon for linux/amd64:

```bash
# From the repository root
cd /workspaces/daytona

# Build using nx
npx nx build daemon --configuration=linux-amd64

# Or build directly with Go
cd apps/daemon
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o daemon-amd64 ./cmd/daemon
```

The binary will be at:

- Using nx: `dist/apps/daemon/daemon-amd64`
- Using go build: `apps/daemon/daemon-amd64`

### Build SSH Backdoor (Optional but Recommended)

The SSH backdoor provides reliable VM access for debugging and updates:

```bash
cd apps/ssh-backdoor
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ssh-backdoor ./cmd/main.go
```

The binary will be at: `apps/ssh-backdoor/ssh-backdoor`

## Step 3: Create the Golden Image

### Option A: Flattened Single Image (Recommended)

This creates a standalone qcow2 image with no backing file dependencies.

```bash
#!/bin/bash
set -eux

BASE_DIR=/var/lib/cloud-hypervisor
IMAGES_DIR=$BASE_DIR/images
SNAPSHOTS_DIR=$BASE_DIR/snapshots
WORK_DIR=$BASE_DIR/sandboxes/golden-build
DAEMON_BIN=/path/to/daemon-amd64

# Clean up previous builds
rm -rf "$WORK_DIR"
mkdir -p "$WORK_DIR"
cd "$WORK_DIR"

# Copy base image
cp "$IMAGES_DIR/ubuntu-24.04-server-cloudimg-amd64.img" golden-temp.qcow2

# Resize to 10GB
qemu-img resize golden-temp.qcow2 10G

# Inject daemon binary
virt-copy-in -a golden-temp.qcow2 "$DAEMON_BIN" /usr/local/bin/
virt-customize -a golden-temp.qcow2 --run-command 'mv /usr/local/bin/daemon-amd64 /usr/local/bin/daytona-daemon'
virt-customize -a golden-temp.qcow2 --run-command 'chmod +x /usr/local/bin/daytona-daemon'

# Create systemd service file
cat > /tmp/daytona-daemon.service <<'EOF'
[Unit]
Description=Daytona Daemon Service
After=network.target

[Service]
ExecStart=/usr/local/bin/daytona-daemon
Restart=always
RestartSec=3
User=root
Group=root
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
EOF

# Inject and enable systemd service
virt-copy-in -a golden-temp.qcow2 /tmp/daytona-daemon.service /etc/systemd/system/
virt-customize -a golden-temp.qcow2 --run-command 'systemctl enable daytona-daemon.service'

# Clean cloud-init to allow re-initialization on first boot
virt-customize -a golden-temp.qcow2 --run-command 'cloud-init clean'

# Optional: Install additional packages
# virt-customize -a golden-temp.qcow2 --install net-tools,htop,curl

# Flatten and compress the image
qemu-img convert -c -O qcow2 golden-temp.qcow2 "$SNAPSHOTS_DIR/ubuntu-base.1.qcow2"

# Set read-only permissions
chmod 444 "$SNAPSHOTS_DIR/ubuntu-base.1.qcow2"

# Verify the golden image
qemu-img info "$SNAPSHOTS_DIR/ubuntu-base.1.qcow2"

# Clean up
rm -rf "$WORK_DIR"

echo "Golden image created: $SNAPSHOTS_DIR/ubuntu-base.1.qcow2"
```

### Option B: Layered Image Architecture

This uses qcow2 backing files for a copy-on-write architecture:

```
Layer 1 (read-only): Ubuntu cloud image (shared by all VMs)
Layer 2 (read-only): Golden image with daemon (shared by all VMs)  
Layer 3 (per-VM):    Thin overlay for each VM instance
```

```bash
#!/bin/bash
set -eux

BASE_DIR=/var/lib/cloud-hypervisor
UBUNTU_BASE="$BASE_DIR/images/ubuntu-24.04-server-cloudimg-amd64.img"
GOLDEN_IMAGE="$BASE_DIR/snapshots/golden-with-daemon.qcow2"
DAEMON_BIN=/path/to/daemon-amd64

# Create Layer 2 - Golden image with daemon as overlay on Ubuntu
qemu-img create -f qcow2 -b "$UBUNTU_BASE" -F qcow2 "$GOLDEN_IMAGE"
qemu-img resize "$GOLDEN_IMAGE" 10G

# Inject daemon (same as Option A)
virt-copy-in -a "$GOLDEN_IMAGE" "$DAEMON_BIN" /usr/local/bin/
virt-customize -a "$GOLDEN_IMAGE" --run-command 'mv /usr/local/bin/daemon-amd64 /usr/local/bin/daytona-daemon'
virt-customize -a "$GOLDEN_IMAGE" --run-command 'chmod +x /usr/local/bin/daytona-daemon'

# Create and inject systemd service (same as Option A)
cat > /tmp/daytona-daemon.service <<'EOF'
[Unit]
Description=Daytona Daemon Service
After=network.target

[Service]
ExecStart=/usr/local/bin/daytona-daemon
Restart=always
RestartSec=3
User=root
Group=root
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
EOF

virt-copy-in -a "$GOLDEN_IMAGE" /tmp/daytona-daemon.service /etc/systemd/system/
virt-customize -a "$GOLDEN_IMAGE" --run-command 'systemctl enable daytona-daemon.service'
virt-customize -a "$GOLDEN_IMAGE" --run-command 'cloud-init clean'

# Make golden image read-only
chmod 444 "$GOLDEN_IMAGE"

echo "Golden image with backing chain created: $GOLDEN_IMAGE"
qemu-img info --backing-chain "$GOLDEN_IMAGE"
```

## Step 4: Extract Kernel and Initramfs (Required)

Cloud Hypervisor requires direct kernel boot with a bzImage kernel and initramfs. The hypervisor-fw firmware doesn't properly set the `root=` kernel parameter, causing boot failures.

Extract the kernel and initramfs from the golden image:

```bash
#!/bin/bash
set -eux

BASE_DIR=/var/lib/cloud-hypervisor
KERNELS_DIR=$BASE_DIR/kernels
GOLDEN_IMAGE=$BASE_DIR/snapshots/ubuntu-base.1.qcow2

mkdir -p "$KERNELS_DIR"

# List available kernels in the image
virt-ls -la "$GOLDEN_IMAGE" /boot/ | grep -E "vmlinuz|initrd"

# Extract kernel and initramfs (adjust version as needed)
virt-copy-out -a "$GOLDEN_IMAGE" \
    /boot/vmlinuz-6.8.0-90-generic \
    /boot/initrd.img-6.8.0-90-generic \
    "$KERNELS_DIR/"

# Verify extraction
ls -la "$KERNELS_DIR/"
file "$KERNELS_DIR/vmlinuz-6.8.0-90-generic"

echo "Kernel and initramfs extracted to $KERNELS_DIR"
```

The runner configuration should then point to these files:

```bash
CH_KERNEL_PATH=/var/lib/cloud-hypervisor/kernels/vmlinuz-6.8.0-90-generic
CH_INITRAMFS_PATH=/var/lib/cloud-hypervisor/kernels/initrd.img-6.8.0-90-generic
```

## Step 5: Upload to S3 (Optional)

Upload the golden image to S3 for distribution.

### Important: Flatten Before Upload

Before uploading, ensure the disk image is **flattened** (has no backing file dependencies). Otherwise, the snapshot won't work on other machines.

```bash
SNAP_DIR=/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1

# Check for backing file
qemu-img info $SNAP_DIR/disk.qcow2 | grep -i backing

# If backing file exists, flatten it:
cd $SNAP_DIR
qemu-img convert -O qcow2 disk.qcow2 disk-flat.qcow2
mv disk.qcow2 disk.qcow2.bak
mv disk-flat.qcow2 disk.qcow2
rm disk.qcow2.bak
```

### Upload Cold Snapshot (disk only)

```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_DEFAULT_REGION=us-east-2

BUCKET=your-snapshots-bucket
IMAGE=/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1.qcow2

aws s3 cp "$IMAGE" "s3://$BUCKET/snapshots/ubuntu-base.1.qcow2"

# Verify upload
aws s3 ls "s3://$BUCKET/snapshots/"
```

### Upload Warm Snapshot (disk + memory state)

Warm snapshots include multiple files that must all be uploaded:

```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_DEFAULT_REGION=us-east-2

BUCKET=your-snapshots-bucket
SNAP_NAME=ubuntu-base.1
SNAP_DIR=/var/lib/cloud-hypervisor/snapshots/$SNAP_NAME

# Upload all snapshot files
aws s3 cp $SNAP_DIR/disk.qcow2 s3://$BUCKET/snapshots/$SNAP_NAME/disk.qcow2
aws s3 cp $SNAP_DIR/config.json s3://$BUCKET/snapshots/$SNAP_NAME/config.json
aws s3 cp $SNAP_DIR/state.json s3://$BUCKET/snapshots/$SNAP_NAME/state.json
aws s3 cp $SNAP_DIR/memory-ranges s3://$BUCKET/snapshots/$SNAP_NAME/memory-ranges

# Verify upload
aws s3 ls "s3://$BUCKET/snapshots/$SNAP_NAME/"
```

## Step 5: Creating VM Overlays

For each new VM, create a thin overlay from the golden image:

```bash
GOLDEN_IMAGE=/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1.qcow2
VM_DIR=/var/lib/cloud-hypervisor/sandboxes/vm-001
VM_ID=vm-001

mkdir -p "$VM_DIR"
cd "$VM_DIR"

# Create thin overlay (copy-on-write)
qemu-img create -f qcow2 -b "$GOLDEN_IMAGE" -F qcow2 disk.qcow2

# Create cloud-init ISO for VM-specific configuration
cat > meta-data <<EOF
instance-id: $VM_ID
local-hostname: $VM_ID
EOF

cat > user-data <<EOF
#cloud-config
users:
  - name: daytona
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - ssh-rsa YOUR_PUBLIC_KEY_HERE

network:
  version: 2
  ethernets:
    eth0:
      dhcp4: true

runcmd:
  - systemctl start daytona-daemon || true
EOF

cloud-localds cloud-init.iso user-data meta-data

echo "VM overlay created at $VM_DIR"
ls -lh "$VM_DIR"
```

## Daemon Configuration

The daemon listens on port **2280** by default. Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `TOOLBOX_API_PORT` | API port | 2280 |
| `HOME` | Home directory | /root |
| `DAYTONA_SANDBOX_USER` | User context for operations | root |

### Service Ports

| Service | Port | Description |
|---------|------|-------------|
| Daemon API | 2280 | Toolbox REST API |
| Daemon SSH | 22220 | Daemon's built-in SSH server |
| Recording Dashboard | 6090 | Screen recording web UI |
| SSH Backdoor | 2222 | Minimal SSH for debugging (password: `sandbox-ssh`) |

## VNC Desktop Support (noVNC)

To enable browser-based VNC desktop access, install the VNC stack and the `daytona-computer-use` daemon plugin.

### VNC Components

| Component | Port | Description |
|-----------|------|-------------|
| Xvfb | N/A | Virtual framebuffer (creates virtual display :0) |
| XFCE4 | N/A | Lightweight desktop environment |
| x11vnc | 5901 | VNC server exposing the X display |
| websockify | 6080 | WebSocket-to-TCP bridge for browser access |
| noVNC | 6080 | Web-based VNC client (served by websockify) |

### Install VNC Packages

Install via apt in the golden image or running VM:

```bash
DEBIAN_FRONTEND=noninteractive apt-get install -y \
    xvfb \
    x11vnc \
    novnc \
    python3-websockify \
    xfce4 \
    xfce4-terminal \
    dbus-x11
```

### Xvfb GLX Wrapper (Required)

Cloud Hypervisor VMs don't have GPU support, causing Xvfb to crash when GLX is enabled. Create a wrapper script to disable GLX:

```bash
# Move original Xvfb binary
mv /usr/bin/Xvfb /usr/bin/Xvfb.real

# Create wrapper script
cat > /usr/bin/Xvfb << 'EOF'
#!/bin/bash
exec /usr/bin/Xvfb.real -extension GLX "$@"
EOF

chmod +x /usr/bin/Xvfb
```

The `-extension GLX` flag disables the GLX extension, preventing crashes in environments without GPU.

### Install Computer-Use Plugin

The `daytona-computer-use` plugin manages the VNC stack (Xvfb, XFCE4, x11vnc, websockify). Build and install it:

```bash
# Build the plugin (from repository root)
cd /workspaces/daytona
npx nx build computer-use --configuration=linux-amd64

# Copy to runner host
scp dist/apps/computer-use/computer-use-amd64 root@<runner-host>:/tmp/daytona-computer-use

# Install in VM (via SSH or daemon API)
cp /tmp/daytona-computer-use /usr/local/lib/daytona-computer-use
chmod +x /usr/local/lib/daytona-computer-use
```

The daemon automatically detects and loads the plugin from `/usr/local/lib/daytona-computer-use`.

### DNS Configuration

VMs need DNS configured for package installation. Configure systemd-resolved:

```bash
mkdir -p /etc/systemd/resolved.conf.d

cat > /etc/systemd/resolved.conf.d/dns.conf << 'EOF'
[Resolve]
DNS=8.8.8.8 8.8.4.4
FallbackDNS=1.1.1.1
EOF

systemctl restart systemd-resolved
```

### Start VNC Services

The daemon exposes endpoints to control VNC:

```bash
VM_ID="your-vm-id"
NS="ns-${VM_ID:0:8}"

# Start VNC desktop
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/computeruse/start

# Check status
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:2280/computeruse/status
# Returns: {"status":"active"}

# Stop VNC desktop
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/computeruse/stop
```

### Verify VNC Access

Test VNC from the browser:

```
http://6080-<sandbox-id>.proxy.localhost:4000/vnc.html?autoconnect=true
```

Or test directly from the runner host:

```bash
VM_ID="your-vm-id"
NS="ns-${VM_ID:0:8}"

# Test noVNC web client
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:6080/vnc.html | head -5

# Test WebSocket upgrade
nsenter --net=/var/run/netns/$NS curl -s -v \
  -H "Upgrade: websocket" \
  -H "Connection: Upgrade" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  -H "Sec-WebSocket-Version: 13" \
  http://192.168.0.2:6080/websockify 2>&1 | grep "101 Switching"
# Should return: HTTP/1.1 101 Switching Protocols

# Test x11vnc directly
nsenter --net=/var/run/netns/$NS curl -s telnet://192.168.0.2:5901 2>&1 | head -1
# Should return: RFB 003.008
```

### VNC in Golden Image Script

Add VNC installation to your golden image creation script:

```bash
# After installing daemon, add VNC support
virt-customize -a golden-temp.qcow2 --install \
    xvfb,x11vnc,novnc,python3-websockify,xfce4,xfce4-terminal,dbus-x11

# Create Xvfb wrapper
virt-customize -a golden-temp.qcow2 --run-command '
    mv /usr/bin/Xvfb /usr/bin/Xvfb.real
    cat > /usr/bin/Xvfb << "WRAPPER"
#!/bin/bash
exec /usr/bin/Xvfb.real -extension GLX "$@"
WRAPPER
    chmod +x /usr/bin/Xvfb
'

# Install computer-use plugin
virt-copy-in -a golden-temp.qcow2 /path/to/daytona-computer-use /usr/local/lib/
virt-customize -a golden-temp.qcow2 --run-command 'chmod +x /usr/local/lib/daytona-computer-use'

# Configure DNS
virt-customize -a golden-temp.qcow2 --run-command '
    mkdir -p /etc/systemd/resolved.conf.d
    cat > /etc/systemd/resolved.conf.d/dns.conf << "DNSCONF"
[Resolve]
DNS=8.8.8.8 8.8.4.4
FallbackDNS=1.1.1.1
DNSCONF
'
```

### VNC Troubleshooting

1. **Xvfb crashes with "Unrecognized option"**: Ensure the Xvfb wrapper is installed correctly

   ```bash
   file /usr/bin/Xvfb  # Should show "ASCII text executable"
   cat /usr/bin/Xvfb   # Should show wrapper script
   ```

2. **websockify not listening on 6080**: Check if the computer-use plugin started successfully

   ```bash
   curl -s http://192.168.0.2:2280/computeruse/status
   ss -tlnp | grep 6080
   ```

3. **noVNC shows blank page or null bytes**: Page cache corruption from warm snapshot. Re-install noVNC package or replace `/usr/share/novnc/` with fresh files.

4. **"Failed to connect to server" in browser**: WebSocket connection failed. Check:
   - websockify is running on port 6080
   - x11vnc is running on port 5901
   - Proxy chain is correctly forwarding WebSocket upgrades

5. **Desktop not appearing**: XFCE4 may have failed to start. Check processes:

   ```bash
   ps aux | grep -E "Xvfb|xfce|x11vnc|websockify"
   ```

## Verification

To verify the daemon is running in a VM:

```bash
# From the host, curl the VM's daemon port
VM_IP=10.0.0.x
curl -v "http://$VM_IP:2280/health"
curl -v "http://$VM_IP:2280/version"
```

## Troubleshooting

### Check daemon status inside VM

```bash
# SSH into the VM
ssh daytona@$VM_IP

# Check daemon status
systemctl status daytona-daemon
journalctl -u daytona-daemon -f
```

### Common Issues

1. **Daemon not starting**: Check if the binary has execute permissions

   ```bash
   ls -la /usr/local/bin/daytona-daemon
   ```

2. **Network not configured**: Ensure cloud-init ran successfully

   ```bash
   cloud-init status
   cat /var/log/cloud-init-output.log
   ```

3. **VM not getting IP**: Check DHCP server on host bridge

   ```bash
   # On host
   ip neigh show
   cat /var/lib/misc/dnsmasq.leases
   ```

## SSH Backdoor for VM Access

The SSH backdoor provides reliable root access to VMs for debugging and updates, bypassing potential issues with the daemon's process execution API (which can fail in warm snapshots due to page cache corruption).

### Overview

| Component | Port | Description |
|-----------|------|-------------|
| ssh-backdoor | 2222 | Minimal SSH server with password auth |

**Credentials:**

- **Port:** 2222
- **User:** root (or any user - not validated)
- **Password:** `sandbox-ssh`

### Building the SSH Backdoor

```bash
cd /workspaces/daytona/apps/ssh-backdoor
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ssh-backdoor ./cmd/main.go
```

### Installing in Golden Image

#### Option A: Via virt-customize (Cold Image)

```bash
SNAP_DIR=/var/lib/cloud-hypervisor/snapshots/ubuntu-base.X

# Copy binary
virt-copy-in -a $SNAP_DIR/disk.qcow2 /path/to/ssh-backdoor /usr/local/bin/

# Create systemd service
cat > /tmp/ssh-backdoor.service << 'EOF'
[Unit]
Description=SSH Backdoor Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/ssh-backdoor
Restart=always
RestartSec=3
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF

virt-copy-in -a $SNAP_DIR/disk.qcow2 /tmp/ssh-backdoor.service /etc/systemd/system/

# Enable service
virt-customize -a $SNAP_DIR/disk.qcow2 \
  --run-command 'chmod +x /usr/local/bin/ssh-backdoor' \
  --run-command 'systemctl enable ssh-backdoor.service'
```

#### Option B: Via Daemon API (Running VM)

```bash
VM_ID="your-vm-id"
NS="ns-${VM_ID:0:8}"

# Upload binary via files API
curl -X POST "http://localhost:3005/sandboxes/$VM_ID/toolbox/files/upload?path=/home/daytona/ssh-backdoor" \
  -F "file=@/path/to/ssh-backdoor"

# Install and enable
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo cp /home/daytona/ssh-backdoor /usr/local/bin/ssh-backdoor"}'

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo chmod +x /usr/local/bin/ssh-backdoor"}'

# Create and install service file (upload via files API, then copy)
# ... similar to Option A service file ...

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo systemctl daemon-reload"}'

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo systemctl enable ssh-backdoor"}'

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo systemctl start ssh-backdoor"}'
```

### Using the SSH Backdoor

#### From Runner Host

```bash
VM_ID="your-vm-id"
NS="ns-${VM_ID:0:8}"

# Test connectivity
nsenter --net=/var/run/netns/$NS nc -zv 192.168.0.2 2222

# SSH into VM
nsenter --net=/var/run/netns/$NS sshpass -p 'sandbox-ssh' \
  ssh -o StrictHostKeyChecking=no -p 2222 root@192.168.0.2 'whoami'

# Copy files to VM
nsenter --net=/var/run/netns/$NS sshpass -p 'sandbox-ssh' \
  scp -o StrictHostKeyChecking=no -P 2222 /path/to/file root@192.168.0.2:/tmp/
```

#### Interactive Session

```bash
nsenter --net=/var/run/netns/$NS sshpass -p 'sandbox-ssh' \
  ssh -o StrictHostKeyChecking=no -p 2222 root@192.168.0.2
```

### Warm Snapshot Considerations

For warm snapshots, the SSH backdoor service must be **running** when the snapshot is taken. If it's only enabled but not started, new VMs from the snapshot won't have the backdoor running.

```bash
# Ensure service is running before taking snapshot
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo systemctl start ssh-backdoor"}'

# Verify
nsenter --net=/var/run/netns/$NS nc -zv 192.168.0.2 2222
# Should return: Connection to 192.168.0.2 2222 port [tcp/*] succeeded!

# Now take the snapshot
```

### Troubleshooting

1. **Connection refused on port 2222**: Service not running

   ```bash
   # Start manually via daemon API or check service status
   nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
     -H "Content-Type: application/json" \
     -d '{"command":"systemctl status ssh-backdoor"}'
   ```

2. **Host key verification failed**: Clear old host key

   ```bash
   ssh-keygen -f '/root/.ssh/known_hosts' -R '[192.168.0.2]:2222'
   ```

3. **Permission denied**: Ensure password is exactly `sandbox-ssh`

---

## Updating Existing Golden Images (Warm Snapshots)

When you need to update the daemon binary in an existing warm snapshot (which includes memory state), follow this process. This is more complex than creating a new image because the warm snapshot contains the page cache with the old binary.

### Prerequisites

- A running VM created from the current golden snapshot
- The new daemon binary built for linux/amd64
- Access to the runner host
- **Recommended:** SSH backdoor installed in the snapshot

### Step 1: Copy New Daemon to Runner Host

```bash
# Build the new daemon
cd /workspaces/daytona
npx nx build daemon --configuration=linux-amd64

# Copy to runner host
scp dist/apps/daemon/daemon-amd64 root@<runner-host>:/tmp/daemon-new
```

### Step 2: Update Daemon in Running VM

#### Method A: Via SSH Backdoor (Recommended)

The SSH backdoor provides reliable file transfer and command execution, bypassing potential issues with the daemon's API in warm snapshots.

```bash
VM_ID="your-vm-id"
NS="ns-${VM_ID:0:8}"

# Copy daemon binary to VM via SCP
nsenter --net=/var/run/netns/$NS sshpass -p 'sandbox-ssh' \
  scp -o StrictHostKeyChecking=no -P 2222 /tmp/daemon-new root@192.168.0.2:/tmp/daytona-daemon

# SSH in and update
nsenter --net=/var/run/netns/$NS sshpass -p 'sandbox-ssh' \
  ssh -o StrictHostKeyChecking=no -p 2222 root@192.168.0.2 '
    # Stop daemon
    systemctl stop daytona-daemon || true
    pkill -9 daytona-daemon || true
    
    # Replace binary
    cp /tmp/daytona-daemon /usr/local/bin/daytona-daemon
    chmod +x /usr/local/bin/daytona-daemon
    rm /tmp/daytona-daemon
    
    # Start daemon (run as daytona user)
    su - daytona -c "HOME=/home/daytona DAYTONA_SANDBOX_USER=daytona nohup /usr/local/bin/daytona-daemon > /home/daytona/daemon.log 2>&1 &"
    
    sleep 3
    ps aux | grep daytona-daemon | grep -v grep
  '

# Verify daemon is working
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:2280/version
```

**Note:** Running the daemon manually (not via systemd) avoids issues with stuck systemd cgroups from warm snapshots.

#### Method B: Via Daemon API with Direct I/O

The key challenge is that warm snapshots restore the kernel's page cache. Even if you update the disk, `systemctl restart` will load the cached (old) binary. Use **direct I/O** to bypass the page cache:

```bash
# Get the VM's network namespace (replace VM_ID)
VM_ID="your-vm-id"
NS="ns-${VM_ID:0:8}"

# First, copy new daemon to the host's sandbox directory
cp /tmp/daemon-new /var/lib/cloud-hypervisor/sandboxes/$VM_ID/daemon-new

# Use the daemon's API to copy with direct I/O (bypasses page cache)
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"dd if=/tmp/daemon-new of=/tmp/daemon-fresh iflag=direct bs=1M","timeout":60}'

# Verify the new binary has your changes (e.g., check for new endpoint)
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"grep -ao memory-stats /tmp/daemon-fresh","timeout":30}'

# Make executable and copy to final location
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"chmod +x /tmp/daemon-fresh","timeout":10}'

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"cp /tmp/daemon-fresh /usr/local/bin/daytona-daemon-new","timeout":10}'

# Update systemd to use new binary path
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sed -i s,/usr/local/bin/daytona-daemon,/usr/local/bin/daytona-daemon-new,g /etc/systemd/system/daytona-daemon.service","timeout":10}'

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"systemctl daemon-reload","timeout":10}'

# Restart daemon (connection will reset - this is expected)
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"systemctl restart daytona-daemon","timeout":30}' || echo "Connection reset (expected)"

# Wait for daemon to start
sleep 5

# Verify new daemon is working
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:2280/version
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:2280/memory-stats  # or your new endpoint
```

### Step 3: Restore Original Daemon Path

Before taking the snapshot, restore the service to use the original binary path:

```bash
# Copy new daemon to original location
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"cp /usr/local/bin/daytona-daemon-new /usr/local/bin/daytona-daemon","timeout":10}'

# Restore systemd service path
nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"sed -i s,/usr/local/bin/daytona-daemon-new,/usr/local/bin/daytona-daemon,g /etc/systemd/system/daytona-daemon.service","timeout":10}'

nsenter --net=/var/run/netns/$NS curl -s -X POST http://192.168.0.2:2280/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command":"systemctl daemon-reload","timeout":10}'
```

### Step 4: Take New Warm Snapshot

```bash
VM_ID="your-vm-id"
NEW_SNAP_DIR="/var/lib/cloud-hypervisor/snapshots/ubuntu-base.X"  # Increment version

# Create snapshot directory
mkdir -p $NEW_SNAP_DIR

# Pause the VM (required for snapshot)
ch-remote --api-socket /var/run/cloud-hypervisor/$VM_ID.sock pause

# Take the snapshot
ch-remote --api-socket /var/run/cloud-hypervisor/$VM_ID.sock snapshot file://$NEW_SNAP_DIR

# Copy the disk
cp /var/lib/cloud-hypervisor/sandboxes/$VM_ID/disk.qcow2 $NEW_SNAP_DIR/disk.qcow2

# Resume the VM
ch-remote --api-socket /var/run/cloud-hypervisor/$VM_ID.sock resume

# Verify snapshot contents
ls -la $NEW_SNAP_DIR/
```

### Step 5: Fix Snapshot Configuration

The snapshot's `config.json` contains hardcoded paths that must be made generic:

```bash
NEW_SNAP_DIR="/var/lib/cloud-hypervisor/snapshots/ubuntu-base.X"

# Fix disk path (use placeholder that runner will replace)
cat $NEW_SNAP_DIR/config.json | jq '.disks[0].path = "DISK_PATH_PLACEHOLDER"' > $NEW_SNAP_DIR/config.json.tmp
mv $NEW_SNAP_DIR/config.json.tmp $NEW_SNAP_DIR/config.json

# Fix serial console (set to Tty mode, no file)
cat $NEW_SNAP_DIR/config.json | jq '.serial.file = null | .serial.mode = "Tty"' > $NEW_SNAP_DIR/config.json.tmp
mv $NEW_SNAP_DIR/config.json.tmp $NEW_SNAP_DIR/config.json

# Fix network config (use generic tap name)
cat $NEW_SNAP_DIR/config.json | jq '.net[0].tap = "tap0" | .net[0].id = "_net0"' > $NEW_SNAP_DIR/config.json.tmp
mv $NEW_SNAP_DIR/config.json.tmp $NEW_SNAP_DIR/config.json

# Verify the configuration
cat $NEW_SNAP_DIR/config.json | jq '{disks: .disks[0].path, serial: .serial.mode, net_tap: .net[0].tap}'
```

### Step 6: Flatten Disk Image (Required for Distribution)

When creating VMs from snapshots, the disk is created as a copy-on-write overlay. This means the `disk.qcow2` has a **backing file** reference to the parent snapshot. For the snapshot to work on a fresh runner machine (or after uploading to S3), you must **flatten** the disk to remove this dependency.

#### Check for Backing File

```bash
NEW_SNAP_DIR="/var/lib/cloud-hypervisor/snapshots/ubuntu-base.X"

# Check if disk has a backing file
qemu-img info $NEW_SNAP_DIR/disk.qcow2
```

If you see a `backing file:` line in the output, the image needs to be flattened:

```
backing file: /var/lib/cloud-hypervisor/snapshots/ubuntu-base.Y/disk.qcow2
backing file format: qcow2
```

#### Flatten the Disk

```bash
NEW_SNAP_DIR="/var/lib/cloud-hypervisor/snapshots/ubuntu-base.X"

cd $NEW_SNAP_DIR

# Convert to standalone image (removes backing file dependency)
qemu-img convert -O qcow2 disk.qcow2 disk-flat.qcow2

# Replace original with flattened version
mv disk.qcow2 disk.qcow2.bak
mv disk-flat.qcow2 disk.qcow2

# Verify no backing file
qemu-img info disk.qcow2 | grep -i backing
# Should return nothing (no backing file)

# Clean up backup (after verifying)
rm disk.qcow2.bak
```

**Note:** Flattening increases the disk size significantly (e.g., from 2.5GB to 8GB) because it merges all data from the backing chain into a single file.

#### Optional: Compress the Flattened Disk

To reduce storage and transfer time:

```bash
# Convert with compression
qemu-img convert -c -O qcow2 disk.qcow2 disk-compressed.qcow2
mv disk-compressed.qcow2 disk.qcow2
```

### Step 7: Update Database

Update the snapshot reference in the database to use the new version:

```sql
-- Example: Update snapshot file path
UPDATE snapshots 
SET file = 'ubuntu-base.X' 
WHERE name = 'your-snapshot-name';
```

### Step 8: Verify New Snapshot

Create a test VM from the new snapshot and verify:

```bash
# After creating a new VM from the updated snapshot
NEW_VM_ID="test-vm-id"
NS="ns-${NEW_VM_ID:0:8}"

# Test daemon endpoints
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:2280/version
nsenter --net=/var/run/netns/$NS curl -s http://192.168.0.2:2280/memory-stats  # New endpoint should work immediately
```

### Important Notes

1. **Page Cache Behavior**: Warm snapshots include the kernel's page cache. Simply updating the disk and restarting the daemon won't work - you must use `iflag=direct` with `dd` to bypass the cache.

2. **Keep Previous Versions**: Always keep at least one previous snapshot version for rollback.

3. **Test Before Production**: Always test the new snapshot by creating a VM and verifying all daemon endpoints work before updating production.

4. **Snapshot Timing**: The snapshot captures the exact memory state. Ensure the daemon is fully started and stable before taking the snapshot.

5. **Flatten Before Distribution**: Always flatten the disk image before uploading to S3 or copying to other machines. Snapshots created from running VMs have backing file dependencies that won't exist on other systems.

## Image Versioning

Use semantic versioning for golden images:

- `ubuntu-base.1` - Initial release
- `ubuntu-base.2` - Updated daemon version
- `ubuntu-base.3` - Security patches
- `ubuntu-base.X` - Always increment for new versions

Always keep at least the previous version available for rollback.

## References

- [Cloud Hypervisor Documentation](https://github.com/cloud-hypervisor/cloud-hypervisor)
- [Ubuntu Cloud Images](https://cloud-images.ubuntu.com/)
- [QEMU qcow2 format](https://qemu-project.gitlab.io/qemu/system/images.html)
- [libguestfs tools](https://libguestfs.org/)
- SSH Backdoor source: `apps/ssh-backdoor/` in this repository
