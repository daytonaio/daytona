# Runner-CH Deployment Guide

This guide covers deploying runner-ch to a remote Cloud Hypervisor host in **local mode** for optimal performance (including live fork with memory state preservation).

## Prerequisites

### On the Remote Host (Cloud Hypervisor Server)

1. **Cloud Hypervisor** installed and configured
2. **Linux kernel** with KVM support
3. **Required packages**: `jq`, `qemu-img`, `curl`, `iproute2`, `iptables`
4. **Network bridge** (`br0`) configured
5. **Base image** at `/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1.qcow2`

### On the Development/Control Machine

1. SSH access to the remote host (key-based authentication)
2. Built runner-ch binary (`yarn nx build runner-ch`)

## Directory Structure (Remote Host)

```bash
# Create required directories
mkdir -p /var/lib/cloud-hypervisor/{sandboxes,snapshots,kernels,firmware}
mkdir -p /var/run/cloud-hypervisor
mkdir -p /var/log
```

## Step 1: Build the Runner

On your development machine:

```bash
cd /workspaces/daytona
yarn nx build runner-ch
```

The binary will be at: `dist/apps/runner-ch`

## Step 2: Deploy the Binary

```bash
# Set variables
SSH_KEY="/path/to/your/ssh/key"
REMOTE_HOST="root@<remote-ip>"

# Copy binary to remote host
scp -i "$SSH_KEY" dist/apps/runner-ch "$REMOTE_HOST:/usr/local/bin/runner-ch-local"

# Make executable
ssh -i "$SSH_KEY" "$REMOTE_HOST" "chmod +x /usr/local/bin/runner-ch-local"
```

## Step 3: Create Configuration

Create the configuration file on the remote host:

```bash
ssh -i "$SSH_KEY" "$REMOTE_HOST" 'cat > /etc/runner-ch-local.env << EOF
# Runner-CH Local Mode Configuration

# Environment
ENVIRONMENT=production
LOG_LEVEL=info

# IMPORTANT: Do NOT set CH_SSH_HOST for local mode
# CH_SSH_HOST is only for remote mode (development)

# Daytona API Configuration
# Use localhost:3000 if using reverse SSH tunnel
SERVER_URL=http://localhost:3000/api
API_PORT=3006
API_TOKEN=<your-runner-api-token>

# SSH Gateway (optional)
SSH_GATEWAY_ENABLE=false

# Cloud Hypervisor Paths
CH_SOCKETS_PATH=/var/run/cloud-hypervisor
CH_SANDBOXES_PATH=/var/lib/cloud-hypervisor/sandboxes
CH_SNAPSHOTS_PATH=/var/lib/cloud-hypervisor/snapshots
CH_KERNEL_PATH=/var/lib/cloud-hypervisor/kernels/vmlinuz
CH_INITRAMFS_PATH=/var/lib/cloud-hypervisor/kernels/initrd.img
CH_BASE_IMAGE_PATH=/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1/disk.qcow2

# Network
CH_BRIDGE_NAME=br0

# Default VM Resources
CH_DEFAULT_CPUS=2
CH_DEFAULT_MEMORY_MB=2048
CH_DEFAULT_DISK_GB=20
EOF'
```

## Step 4: Set Up Reverse SSH Tunnel

The runner needs to connect back to your Daytona API server. If your API server is not publicly accessible, use a reverse SSH tunnel.

### Option A: Manual Reverse Tunnel

On your development/API machine, run:

```bash
# This creates a tunnel where port 3000 on the remote host 
# forwards to localhost:3000 on your machine
ssh -i "$SSH_KEY" \
    -o ServerAliveInterval=30 \
    -o ServerAliveCountMax=3 \
    -o ExitOnForwardFailure=yes \
    -N -R 3000:localhost:3000 \
    "$REMOTE_HOST"
```

**Flags explained:**

- `-N`: Don't execute remote command (tunnel only)
- `-R 3000:localhost:3000`: Remote port forwarding
- `-o ServerAliveInterval=30`: Keep connection alive
- `-o ExitOnForwardFailure=yes`: Exit if tunnel fails

### Option B: Persistent Tunnel with autossh

Install `autossh` for automatic reconnection:

```bash
# Install autossh on your machine
apt-get install autossh  # Debian/Ubuntu
brew install autossh     # macOS

# Start persistent tunnel
autossh -M 0 \
    -o "ServerAliveInterval 30" \
    -o "ServerAliveCountMax 3" \
    -N -R 3000:localhost:3000 \
    -i "$SSH_KEY" \
    "$REMOTE_HOST"
```

### Option C: Systemd Service for Tunnel

Create a systemd service on your API machine:

```bash
sudo cat > /etc/systemd/system/daytona-tunnel.service << EOF
[Unit]
Description=SSH Tunnel to Cloud Hypervisor Host
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/ssh -N -R 3000:localhost:3000 \
    -o ServerAliveInterval=30 \
    -o ServerAliveCountMax=3 \
    -o ExitOnForwardFailure=yes \
    -i /path/to/ssh/key \
    root@<remote-ip>
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable daytona-tunnel
sudo systemctl start daytona-tunnel
```

## Step 5: Start the Runner

### Manual Start

```bash
ssh -i "$SSH_KEY" "$REMOTE_HOST" << 'EOF'
cd /tmp
set -a
source /etc/runner-ch-local.env
set +a
nohup /usr/local/bin/runner-ch-local > /var/log/runner-ch-local.log 2>&1 &
echo "Runner started with PID: $!"
EOF
```

### Systemd Service (Recommended for Production)

Create systemd service on the remote host:

```bash
ssh -i "$SSH_KEY" "$REMOTE_HOST" 'cat > /etc/systemd/system/runner-ch.service << EOF
[Unit]
Description=Daytona Runner CH (Local Mode)
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/runner-ch-local.env
ExecStart=/usr/local/bin/runner-ch-local
Restart=on-failure
RestartSec=5
StandardOutput=append:/var/log/runner-ch-local.log
StandardError=append:/var/log/runner-ch-local.log

[Install]
WantedBy=multi-user.target
EOF'

# Enable and start
ssh -i "$SSH_KEY" "$REMOTE_HOST" "systemctl daemon-reload && systemctl enable runner-ch && systemctl start runner-ch"
```

## Step 6: Verify Deployment

### Check Runner Status

```bash
# Check if running
ssh -i "$SSH_KEY" "$REMOTE_HOST" "systemctl status runner-ch"

# Or check process
ssh -i "$SSH_KEY" "$REMOTE_HOST" "pgrep -fa runner-ch-local"

# Check health endpoint
ssh -i "$SSH_KEY" "$REMOTE_HOST" "curl -s http://localhost:3006/"
# Expected: {"status":"ok","version":"..."}
```

### Check Tunnel (from remote host)

```bash
ssh -i "$SSH_KEY" "$REMOTE_HOST" "curl -s http://localhost:3000/api/health || echo 'Tunnel not connected'"
```

### View Logs

```bash
ssh -i "$SSH_KEY" "$REMOTE_HOST" "tail -f /var/log/runner-ch-local.log"
```

## Updating the Runner

```bash
# Build new version
yarn nx build runner-ch

# Stop runner
ssh -i "$SSH_KEY" "$REMOTE_HOST" "systemctl stop runner-ch"

# Copy new binary
scp -i "$SSH_KEY" dist/apps/runner-ch "$REMOTE_HOST:/usr/local/bin/runner-ch-local"

# Start runner
ssh -i "$SSH_KEY" "$REMOTE_HOST" "systemctl start runner-ch"
```

## Troubleshooting

### Runner won't start

```bash
# Check logs
ssh -i "$SSH_KEY" "$REMOTE_HOST" "journalctl -u runner-ch -n 50"

# Check environment
ssh -i "$SSH_KEY" "$REMOTE_HOST" "cat /etc/runner-ch-local.env"
```

### Can't connect to API (401 Unauthorized)

1. Verify `API_TOKEN` in config matches Daytona API expectations
2. Check reverse tunnel is running
3. Test tunnel: `curl -s http://localhost:3000/api/health`

### Fork fails with disk lock errors

This happens when the snapshot config still references the source VM's disk. The runner should automatically patch disk paths, but verify:

```bash
# Check if jq is installed
ssh -i "$SSH_KEY" "$REMOTE_HOST" "which jq || apt-get install -y jq"
```

### Network namespace issues

```bash
# List namespaces
ssh -i "$SSH_KEY" "$REMOTE_HOST" "ip netns list"

# Check veth interfaces
ssh -i "$SSH_KEY" "$REMOTE_HOST" "ip link show type veth"

# Check iptables rules
ssh -i "$SSH_KEY" "$REMOTE_HOST" "iptables -t nat -L -n"
```

## Security Considerations

1. **Firewall**: Only expose necessary ports (3006 for runner API if needed)
2. **SSH Keys**: Use dedicated keys for deployment, not personal keys
3. **API Token**: Use strong, unique tokens for each runner
4. **Reverse Tunnel**: The tunnel only exposes what you specify (port 3000)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Development Machine                          │
│  ┌─────────────┐                                                │
│  │ Daytona API │◄──────┐                                        │
│  │ :3000       │       │                                        │
│  └─────────────┘       │ Reverse SSH Tunnel                     │
│                        │ (R 3000:localhost:3000)                │
└────────────────────────┼────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│              Remote Host (Cloud Hypervisor)                      │
│                                                                  │
│  ┌─────────────────┐      ┌─────────────────────────────────┐   │
│  │ runner-ch-local │      │ Cloud Hypervisor VMs            │   │
│  │ :3006           │─────►│                                 │   │
│  │                 │      │  ┌─────────┐  ┌─────────┐      │   │
│  │ • Job Poller    │      │  │ VM 1    │  │ VM 2    │      │   │
│  │ • Fork Handler  │      │  │ ns-vm1  │  │ ns-vm2  │      │   │
│  │ • Proxy         │      │  └─────────┘  └─────────┘      │   │
│  └────────┬────────┘      └─────────────────────────────────┘   │
│           │                                                      │
│           ▼                                                      │
│  localhost:3000 ◄── Tunnel ── SSH ── Your Machine               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Reference

| Item | Location/Command |
|------|------------------|
| Binary | `/usr/local/bin/runner-ch-local` |
| Config | `/etc/runner-ch-local.env` |
| Logs | `/var/log/runner-ch-local.log` |
| Systemd | `systemctl {start,stop,status} runner-ch` |
| Health check | `curl http://localhost:3006/` |
| Start tunnel | `ssh -N -R 3000:localhost:3000 root@host` |

---

_Last updated: 2026-01-21_
