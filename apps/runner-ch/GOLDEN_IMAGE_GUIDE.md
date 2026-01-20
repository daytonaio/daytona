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

Upload the golden image to S3 for distribution:

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

## Image Versioning

Use semantic versioning for golden images:

- `ubuntu-base.1.qcow2` - Initial release
- `ubuntu-base.2.qcow2` - Updated daemon version
- `ubuntu-base.3.qcow2` - Security patches

Always keep at least the previous version available for rollback.

## References

- [Cloud Hypervisor Documentation](https://github.com/cloud-hypervisor/cloud-hypervisor)
- [Ubuntu Cloud Images](https://cloud-images.ubuntu.com/)
- [QEMU qcow2 format](https://qemu-project.gitlab.io/qemu/system/images.html)
- [libguestfs tools](https://libguestfs.org/)
