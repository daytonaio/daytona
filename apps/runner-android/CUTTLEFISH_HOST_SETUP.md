# Cuttlefish Host Server Setup Guide

This guide covers setting up a bare-metal or VM server to run Android Cuttlefish virtual devices with the Daytona runner-android service.

## Prerequisites

- **Ubuntu 22.04 or 24.04** (tested on both)
- **KVM support** - server must have hardware virtualization enabled
- **Root access**
- **Minimum specs**: 4 CPU cores, 8GB RAM, 50GB disk space per instance

## Step 1: Verify KVM Support

```bash
# Check if KVM is available
ls -la /dev/kvm

# If not present, enable virtualization in BIOS/hypervisor settings
# For cloud VMs, ensure nested virtualization is enabled
```

## Step 2: Install Cuttlefish Host Packages

### Option A: Using PPA (Ubuntu 22.04)

```bash
# Add the Cuttlefish PPA
sudo apt install -y software-properties-common
sudo add-apt-repository ppa:cuttlefish-try/ppa -y
sudo apt update

# Install packages
sudo apt install -y cuttlefish-base cuttlefish-user cuttlefish-common cuttlefish-orchestration
```

### Option B: Download Pre-built Packages (Ubuntu 24.04+)

```bash
# Download latest host packages from Android CI
# Visit: https://ci.android.com/builds/branches/aosp-main/grid
# Look for aosp_cf_x86_64_phone-trunk_staging-userdebug builds
# Download cvd-host_package.tar.gz

# Or use cvd fetch (if cuttlefish-common is installed):
sudo apt install -y git devscripts equivs config-package-dev debhelper-compat golang curl

# Clone and build minimal packages
cd /tmp
git clone https://github.com/google/android-cuttlefish
cd android-cuttlefish
for dir in base frontend; do
  pushd $dir
  sudo mk-build-deps -i -t 'apt-get -o Debug::pkgProblemResolver=yes --no-install-recommends -y'
  dpkg-buildpackage -uc -us
  popd
done
sudo dpkg -i ./cuttlefish-base_*.deb ./cuttlefish-user_*.deb || sudo apt-get install -f -y
sudo dpkg -i ./cuttlefish-common_*.deb ./cuttlefish-orchestration_*.deb || sudo apt-get install -f -y
```

## Step 3: Create Cuttlefish User

```bash
# Create the vsoc-01 user with required groups
sudo useradd -m vsoc-01 -G cvdnetwork,kvm,render,video,audio

# Set a password (optional, for debugging)
sudo passwd vsoc-01

# Also add root to required groups (for runner)
sudo usermod -aG kvm,cvdnetwork root
```

## Step 4: Configure Kernel Modules

```bash
# Ensure vsock modules load at boot
cat << 'EOF' | sudo tee /etc/modules-load.d/cuttlefish.conf
vsock
vhost_vsock
vsock_loopback
vhost_net
EOF

# Load modules now
sudo modprobe vsock vhost_vsock vsock_loopback vhost_net

# Verify
lsmod | grep vsock
```

## Step 5: Configure Kernel Parameters (Ubuntu 24.04)

Ubuntu 24.04 has AppArmor restrictions that block crosvm. Fix with:

```bash
# Disable AppArmor restriction for unprivileged user namespaces
echo 'kernel.apparmor_restrict_unprivileged_userns = 0' | sudo tee /etc/sysctl.d/99-cuttlefish.conf

# Enable unprivileged user namespace cloning
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.d/99-cuttlefish.conf

# Apply now
sudo sysctl -p /etc/sysctl.d/99-cuttlefish.conf
```

## Step 6: Verify Device Permissions

```bash
# Check udev rules are in place
ls -la /lib/udev/rules.d/60-cuttlefish*.rules

# Verify device permissions
ls -la /dev/kvm /dev/vsock /dev/vhost-vsock /dev/vhost-net

# Expected output:
# /dev/kvm         - group kvm
# /dev/vhost-vsock - group cvdnetwork
# /dev/vhost-net   - group cvdnetwork (or kvm)
# /dev/vsock       - world readable (crw-rw-rw-)
```

## Step 7: ⚠️ REBOOT THE SERVER

**This is critical!** The kernel modules, udev rules, and group memberships only take full effect after a reboot.

```bash
sudo reboot
```

## Step 8: Download Android Images

After reboot, download the Android system images:

```bash
# Create directories
sudo mkdir -p /var/lib/cuttlefish/artifacts
sudo mkdir -p /var/lib/cuttlefish/instances
sudo chown -R vsoc-01:vsoc-01 /var/lib/cuttlefish

# Download images using cvd fetch
sudo -u vsoc-01 cvd fetch \
  --target_directory=/var/lib/cuttlefish/artifacts/android-LATEST \
  --default_build=aosp-main/aosp_cf_x86_64_phone-trunk_staging-userdebug

# Or download manually from Android CI:
# https://ci.android.com/builds/branches/aosp-main/grid
# Download: aosp_cf_x86_64_phone-img-*.zip and cvd-host_package.tar.gz
```

## Step 9: Test CVD Launch

```bash
# Switch to vsoc-01 user
sudo -u vsoc-01 -i

# Launch a test instance
cvd create \
  --config=phone \
  --host_path=/var/lib/cuttlefish/artifacts/android-LATEST \
  --product_path=/var/lib/cuttlefish/artifacts/android-LATEST \
  --instance_nums=1

# Check status
cvd fleet

# Expected output should show:
# "status" : "Running"
# "webrtc_device_id" : "cvd_1-1-1"
```

## Step 10: Verify Connectivity

### ADB

```bash
# Connect via ADB
adb connect localhost:6520
adb devices

# Expected: localhost:6520 device
```

### WebRTC

```bash
# Check WebRTC endpoint
curl -k https://localhost:1443/devices

# Access in browser (with SSH tunnel):
# ssh -L 1443:localhost:1443 root@<server-ip>
# Then open: https://localhost:1443/devices/cvd_1-1-1/files/client.html
```

## Step 11: Install Runner-Android

```bash
# Create runner directory
mkdir -p /root/apps/runner-android
cd /root/apps/runner-android

# Copy runner binary and config (from your build)
# scp runner user@server:/root/apps/runner-android/
# scp .env user@server:/root/apps/runner-android/

# Create .env file
cat << 'EOF' > .env
PORT=3007
API_TOKEN=secret_api_token_3007

# Cuttlefish paths
INSTANCES_PATH=/var/lib/cuttlefish/instances
ARTIFACTS_PATH=/var/lib/cuttlefish/artifacts
CVD_HOME=/home/vsoc-01

# Instance settings
ADB_BASE_PORT=6520
MAX_INSTANCES=100
DEFAULT_CPUS=4
DEFAULT_MEMORY=8192

# Local mode (no Daytona API)
LOCAL_MODE=true

# Or connect to Daytona API:
# DAYTONA_API_URL=https://api.daytona.io
# DAYTONA_API_KEY=your_api_key
# DOMAIN=your-runner-domain.example.com
EOF

# Start runner
./runner
```

## Troubleshooting

### "No such device" vsock errors

```bash
# Reload vsock modules
sudo modprobe -r vsock_loopback vhost_vsock
sudo modprobe vsock vhost_vsock vsock_loopback

# If persists, reboot the server
sudo reboot
```

### "User must be a member of cvdnetwork/kvm"

```bash
# Add user to groups
sudo usermod -aG cvdnetwork,kvm vsoc-01
sudo usermod -aG cvdnetwork,kvm root

# Log out and back in, or reboot
```

### AppArmor blocking crosvm (Ubuntu 24.04)

```bash
# Check dmesg for AppArmor denials
dmesg | grep -i apparmor | tail -20

# Disable restriction
sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0
```

### CVD launch fails with "Could not find host tools"

```bash
# Ensure host_path points to directory containing bin/cvd_internal_start
ls /var/lib/cuttlefish/artifacts/android-LATEST/bin/

# The cvd create command needs both paths:
cvd create \
  --host_path=/var/lib/cuttlefish/artifacts/android-LATEST \
  --product_path=/var/lib/cuttlefish/artifacts/android-LATEST \
  ...
```

### Multiple CVD groups conflict

```bash
# Stop all instances
cvd stop --clear_instance_dirs=true

# Kill any remaining processes
pkill -9 -f cuttlefish
pkill -9 -f cvd
pkill -9 -f crosvm

# Restart cuttlefish-operator
sudo systemctl restart cuttlefish-operator

# Clean up temp directories
rm -rf /var/tmp/cvd/*
rm -rf /tmp/cf_avd_*
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Server                                    │
│                                                                  │
│  ┌─────────────────┐    ┌─────────────────────────────────────┐ │
│  │  runner-android │    │         Cuttlefish Host             │ │
│  │   (port 3007)   │    │                                     │ │
│  │                 │    │  ┌─────────────┐  ┌──────────────┐  │ │
│  │  - REST API     │───▶│  │  Operator   │  │  CVD Manager │  │ │
│  │  - ADB proxy    │    │  │ (port 1443) │  │              │  │ │
│  │  - WebRTC proxy │    │  └──────┬──────┘  └──────┬───────┘  │ │
│  └─────────────────┘    │         │                │          │ │
│                         │         ▼                ▼          │ │
│                         │  ┌─────────────────────────────────┐│ │
│                         │  │      Android VM (crosvm)        ││ │
│                         │  │                                 ││ │
│                         │  │  - ADB: vsock:3:5555 → 6520     ││ │
│                         │  │  - WebRTC: port 8443            ││ │
│                         │  │  - Device ID: cvd_1-X-X         ││ │
│                         │  └─────────────────────────────────┘│ │
│                         └─────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Key Ports

| Port | Service | Protocol |
|------|---------|----------|
| 1443 | Cuttlefish Operator (WebRTC signaling) | HTTPS |
| 3007 | Runner-Android API | HTTP |
| 6520+ | ADB (per instance: 6520, 6521, ...) | TCP |
| 8443 | WebRTC (direct, per instance) | HTTPS |

## Service Management

```bash
# Cuttlefish operator
sudo systemctl status cuttlefish-operator
sudo systemctl restart cuttlefish-operator

# Runner (if installed as service)
sudo systemctl status runner-android
sudo systemctl restart runner-android

# View operator logs
tail -f /run/cuttlefish/operator.log

# View CVD instance logs
tail -f /var/tmp/cvd/*/home/cuttlefish/instances/cvd-1/logs/launcher.log
```

## References

- [Official Cuttlefish Documentation](https://source.android.com/docs/devices/cuttlefish)
- [Android Cuttlefish GitHub](https://github.com/google/android-cuttlefish)
- [Android CI Builds](https://ci.android.com/builds/branches/aosp-main/grid)
