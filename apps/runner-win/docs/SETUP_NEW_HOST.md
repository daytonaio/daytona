# Setting Up a New Windows Runner Host

This guide documents the steps to set up a new host for running Windows sandboxes with the Daytona runner-win.

## Prerequisites

- Ubuntu 24.04 or later
- Root SSH access to the new host
- SSH access to an existing host with the Windows base image (e.g., h1001.blinkbox.dev)
- At least 16GB RAM recommended
- CPU with virtualization support (VT-x/AMD-V)

## 1. Install Libvirt and KVM

SSH into the new host and install the required packages:

```bash
ssh root@NEW_HOST

# Install KVM, libvirt, and related tools
apt-get update
apt-get install -y qemu-kvm libvirt-daemon-system libvirt-clients \
    bridge-utils virtinst virt-manager cpu-checker ovmf iptables-persistent

# Verify KVM support
kvm-ok

# Enable and start libvirtd
systemctl enable --now libvirtd

# Verify libvirt is running
virsh list --all
```

## 2. Configure Libvirt Network

The runner expects VMs to use IPs in the `10.100.x.x` range. Configure the libvirt network:

```bash
# Create network configuration
cat > /tmp/daytona-network.xml << 'EOF'
<network>
  <name>default</name>
  <forward mode="nat">
    <nat>
      <port start="1024" end="65535"/>
    </nat>
  </forward>
  <bridge name="virbr0" stp="on" delay="0"/>
  <ip address="10.100.0.1" netmask="255.255.0.0">
    <dhcp>
      <range start="10.100.1.0" end="10.100.15.254"/>
    </dhcp>
  </ip>
</network>
EOF

# Remove existing default network (if any)
virsh net-destroy default 2>/dev/null || true
virsh net-undefine default 2>/dev/null || true

# Create and start the new network
virsh net-define /tmp/daytona-network.xml
virsh net-start default
virsh net-autostart default

# Verify
virsh net-dumpxml default
ip addr show virbr0 | grep inet
```

Expected output: `inet 10.100.0.1/16`

## 3. Set Up Windows Base Image

**CRITICAL**: The runner has hardcoded paths for base images (snapshots) and sandbox overlays:

- Snapshots directory: `/var/lib/libvirt/snapshots/` (base images / golden templates)
- Sandboxes directory: `/var/lib/libvirt/sandboxes/` (per-sandbox overlay disks)
- NVRAM template: `/var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd`

You have two options:

### Option A: Copy to Standard Location (Recommended for new hosts)

```bash
# Create directories
ssh root@NEW_HOST "mkdir -p /var/lib/libvirt/snapshots /var/lib/libvirt/sandboxes /var/lib/libvirt/qemu/nvram"

# Copy base image from existing host
scp root@SOURCE_HOST:/path/to/winserver-autologin-base.qcow2 \
    root@NEW_HOST:/var/lib/libvirt/snapshots/

# Create NVRAM template from OVMF defaults
ssh root@NEW_HOST "
cp /usr/share/OVMF/OVMF_VARS_4M.fd \
   /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd
"

# Set permissions
ssh root@NEW_HOST "
chown -R libvirt-qemu:kvm /var/lib/libvirt/snapshots
chmod 755 /var/lib/libvirt/snapshots
chmod 644 /var/lib/libvirt/snapshots/winserver-autologin-base.qcow2
chown -R libvirt-qemu:kvm /var/lib/libvirt/sandboxes
chmod 755 /var/lib/libvirt/sandboxes
chown libvirt-qemu:kvm /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd
chmod 644 /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd
"
```

### Option B: Use Custom Location with Symlink

If you prefer to store the base image elsewhere (e.g., `/home/vedran/daytona-win-snapshots/`):

```bash
# Create snapshot directory and copy image there
ssh root@NEW_HOST "mkdir -p /home/vedran/daytona-win-snapshots"

# Copy base image to custom location
scp root@SOURCE_HOST:/path/to/winserver-autologin-base.qcow2 \
    root@NEW_HOST:/home/vedran/daytona-win-snapshots/

# Create symlink to standard location (runner expects this path)
ssh root@NEW_HOST "
mkdir -p /var/lib/libvirt/snapshots /var/lib/libvirt/sandboxes
ln -sf /home/vedran/daytona-win-snapshots/winserver-autologin-base.qcow2 \
       /var/lib/libvirt/snapshots/winserver-autologin-base.qcow2
chown -R libvirt-qemu:kvm /var/lib/libvirt/sandboxes
"

# Create NVRAM template
ssh root@NEW_HOST "
mkdir -p /var/lib/libvirt/qemu/nvram && \
cp /usr/share/OVMF/OVMF_VARS_4M.fd \
   /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd
"
```

### Verify Setup

```bash
ssh root@NEW_HOST "
echo '=== Base Image (Snapshot) ===' && \
qemu-img info /var/lib/libvirt/snapshots/winserver-autologin-base.qcow2 | grep -E '(image:|virtual size:|disk size:)' && \
echo '' && \
echo '=== Sandboxes Directory ===' && \
ls -la /var/lib/libvirt/sandboxes/ && \
echo '' && \
echo '=== NVRAM Template ===' && \
ls -lh /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd
"
```

Expected output:

- Base image: ~13GB qcow2 file (or symlink pointing to it)
- NVRAM template: ~528KB file

## 4. Verify Base Image Works

Test that VMs can be created from the base image:

```bash
ssh root@NEW_HOST

# Create overlay disk in sandboxes directory
qemu-img create -f qcow2 -F qcow2 \
    -b /var/lib/libvirt/snapshots/winserver-autologin-base.qcow2 \
    /var/lib/libvirt/sandboxes/test-vm.qcow2
chown libvirt-qemu:kvm /var/lib/libvirt/sandboxes/test-vm.qcow2

# Copy NVRAM
cp /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd \
    /var/lib/libvirt/qemu/nvram/test-vm_VARS.fd
chown libvirt-qemu:kvm /var/lib/libvirt/qemu/nvram/test-vm_VARS.fd

# Create test VM
cat > /tmp/test-vm.xml << 'EOF'
<domain type='kvm'>
  <name>test-vm</name>
  <memory unit='KiB'>4194304</memory>
  <vcpu>2</vcpu>
  <os firmware='efi'>
    <type arch='x86_64' machine='q35'>hvm</type>
    <loader readonly='yes' secure='yes' type='pflash'>/usr/share/OVMF/OVMF_CODE_4M.ms.fd</loader>
    <nvram template='/usr/share/OVMF/OVMF_VARS_4M.ms.fd'>/var/lib/libvirt/qemu/nvram/test-vm_VARS.fd</nvram>
    <boot dev='hd'/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <hyperv mode='custom'>
      <relaxed state='on'/>
      <vapic state='on'/>
      <spinlocks state='on' retries='8191'/>
    </hyperv>
  </features>
  <cpu mode='host-passthrough'/>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/var/lib/libvirt/sandboxes/test-vm.qcow2'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'/>
  </devices>
</domain>
EOF

virsh define /tmp/test-vm.xml
virsh start test-vm

# Wait for boot and get IP
sleep 60
virsh domifaddr test-vm
```

Test daemon connectivity:

```bash
VM_IP=$(virsh domifaddr test-vm | grep -oP '10\.100\.\d+\.\d+')
curl -s http://$VM_IP:2280/version
curl -s http://$VM_IP:2280/computeruse/status
```

Clean up test VM:

```bash
virsh destroy test-vm
virsh undefine test-vm --nvram
rm -f /var/lib/libvirt/sandboxes/test-vm.qcow2
```

## 5. Deploy Runner Binary

Build and deploy the runner-win binary:

```bash
# On development machine
cd /workspaces/daytona
yarn nx build runner-win

# Copy to new host
scp dist/apps/runner-win root@NEW_HOST:/tmp/

# On new host
ssh root@NEW_HOST "
mkdir -p /opt/daytona /var/lib/daytona/runner /var/log/daytona /etc/daytona
mv /tmp/runner-win /opt/daytona/runner
chmod +x /opt/daytona/runner
"
```

## 6. Create Configuration

Create the environment configuration file:

```bash
# Replace these values with your actual configuration
RUNNER_TOKEN="your_runner_token_here"
RUNNER_DOMAIN="NEW_HOST_IP_OR_HOSTNAME"
SSH_PUBLIC_KEY=$(cat ~/.ssh/id_rsa.pub | base64 -w0)

ssh root@NEW_HOST "cat > /etc/daytona/runner.env << EOF
# Daytona Runner Windows Configuration

# API Configuration
DAYTONA_API_URL=http://localhost:3001/api
DAYTONA_RUNNER_TOKEN=$RUNNER_TOKEN
API_TOKEN=$RUNNER_TOKEN

# Runner API Port
API_PORT=8080

# Libvirt Configuration
LIBVIRT_URI=qemu:///system

# Environment
ENVIRONMENT=production
GIN_MODE=release

# Resource limits
RESOURCE_LIMITS_DISABLED=true

# Runner Domain (external IP for sandbox access)
RUNNER_DOMAIN=$RUNNER_DOMAIN

# SSH Gateway
SSH_GATEWAY_ENABLE=true
SSH_PUBLIC_KEY=$SSH_PUBLIC_KEY
EOF

chmod 600 /etc/daytona/runner.env"
```

## 7. Create Systemd Service

```bash
ssh root@NEW_HOST "cat > /etc/systemd/system/daytona-runner.service << 'EOF'
[Unit]
Description=Daytona Runner Service (Windows Sandboxes)
Documentation=https://github.com/daytonaio/daytona
After=network.target libvirtd.service
Wants=network-online.target
Requires=libvirtd.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/var/lib/daytona/runner

# Binary location
ExecStart=/opt/daytona/runner

# Environment file for configuration
EnvironmentFile=/etc/daytona/runner.env

# Restart policy
Restart=on-failure
RestartSec=5s

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=daytona-runner

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable daytona-runner
systemctl start daytona-runner
"
```

## 8. Register Runner in API

Add the runner to the API database:

```sql
INSERT INTO runners (
  id, 
  domain, 
  api_url, 
  api_key, 
  target_count,
  label,
  region,
  state,
  unschedulable,
  target,
  version
) VALUES (
  gen_random_uuid(),
  'NEW_HOST_IP',
  'http://NEW_HOST_IP:8080',
  'your_runner_token_here',
  8,
  'windows-NEW_HOST',
  'eu',
  'ready',
  false,
  'windows.remote',
  'v0.0.0-dev'
);
```

## 9. Set Up SSH Tunnel (if needed)

If the runner needs to reach the API via tunnel:

```bash
# From the API server
ssh -R 3001:localhost:3000 root@NEW_HOST -N &
```

Or create a persistent tunnel service on the API server.

## Verification

### Check runner status

```bash
ssh root@NEW_HOST "
systemctl status daytona-runner
curl -s http://localhost:8080/
journalctl -u daytona-runner -n 20 --no-pager
"
```

### Test sandbox creation

Use the Python SDK to create a sandbox targeting the new runner:

```python
from daytona import Daytona, SandboxTargetRegion

daytona = Daytona()
sandbox = daytona.create(
    CreateSandboxParams(
        language="python",
        target=SandboxTargetRegion(target="windows.remote", region="eu")
    )
)
print(f"Sandbox created: {sandbox.id}")
files = sandbox.fs.list_files("C:\\")
print(files)
```

## Troubleshooting

### Runner can't connect to API

Check SSH tunnel is active and runner logs:

```bash
journalctl -u daytona-runner -f
```

Look for "401 Unauthorized" (token mismatch) or connection errors.

### Sandbox creation fails

1. Check libvirt network is active:

   ```bash
   virsh net-list
   ```

2. Verify network uses correct IP range (`10.100.x.x`):

   ```bash
   virsh net-dumpxml default | grep -A5 dhcp
   ```

3. Check DHCP reservations are being added:

   ```bash
   virsh net-dumpxml default | grep host
   ```

### Proxy returns 502 Bad Gateway

1. VM IP doesn't match calculated IP - check network configuration
2. Daemon not running inside VM - check VM boot status
3. Test daemon directly:

   ```bash
   VM_IP=$(virsh domifaddr SANDBOX_ID | grep -oP '10\.100\.\d+\.\d+')
   curl -v http://$VM_IP:2280/version
   ```

### VM not getting expected IP

The network DHCP range must be `10.100.x.x`:

```bash
virsh net-dumpxml default | grep range
# Should show: <range start='10.100.1.0' end='10.100.15.254'/>
```

If not, reconfigure the network (see Step 2).

## Service Management

```bash
# Status
systemctl status daytona-runner

# Logs (live)
journalctl -u daytona-runner -f

# Restart
systemctl restart daytona-runner

# Stop
systemctl stop daytona-runner

# View configuration
cat /etc/daytona/runner.env
```

## Quick Reference

| Component | Path |
|-----------|------|
| Binary | `/opt/daytona/runner` |
| Config | `/etc/daytona/runner.env` |
| Service | `/etc/systemd/system/daytona-runner.service` |
| Snapshots Dir | `/var/lib/libvirt/snapshots/` (base images) |
| Sandboxes Dir | `/var/lib/libvirt/sandboxes/` (overlay disks) |
| Base Image | `/var/lib/libvirt/snapshots/winserver-autologin-base.qcow2` |
| NVRAM Template | `/var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd` |
| Logs | `journalctl -u daytona-runner` |

| Port | Service |
|------|---------|
| 8080 | Runner API |
| 2280 | Daemon (inside VMs) |
| 22220 | SSH Server (inside VMs) |
| 22222 | Web Terminal (inside VMs) |

| Network | Range |
|---------|-------|
| Bridge IP | 10.100.0.1 |
| DHCP Range | 10.100.1.0 - 10.100.15.254 |
