#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Deploy daemon-win.exe to win11-clone VM on h1001.blinkbox.dev
# This script requires SSH access to h1001.blinkbox.dev and assumes
# OpenSSH Server is enabled on the Windows VM.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKSPACE_ROOT="$(cd "$PROJECT_ROOT/../.." && pwd)"

# Configuration
LIBVIRT_HOST="h1001.blinkbox.dev"
VM_NAME="${WIN_VM_NAME:-winserver-core}"
BINARY_PATH="${WORKSPACE_ROOT}/dist/apps/daemon-win.exe"
WIN_USER="${WIN_USER:-Administrator}"
WIN_PASS="${WIN_PASS:-DaytonaWinAcc3ss!}"
WIN_DEPLOY_PATH="${WIN_DEPLOY_PATH:-C:\\daytona}"

echo "=== Daytona Windows Daemon Deployment ==="
echo "Libvirt Host: $LIBVIRT_HOST"
echo "VM Name: $VM_NAME"
echo "Binary: $BINARY_PATH"
echo ""

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Run 'nx build-windows daemon-win' first."
    exit 1
fi

# Get VM IP address from libvirt
echo "Getting VM IP address from libvirt..."
VM_IP=$(ssh "$LIBVIRT_HOST" "virsh domifaddr $VM_NAME --source agent 2>/dev/null | grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' | head -1" || true)

if [ -z "$VM_IP" ]; then
    # Fallback: try to get IP from ARP table
    echo "Trying to get IP from network leases..."
    VM_MAC=$(ssh "$LIBVIRT_HOST" "virsh domiflist $VM_NAME | grep -oE '([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}' | head -1" || true)
    if [ -n "$VM_MAC" ]; then
        VM_IP=$(ssh "$LIBVIRT_HOST" "virsh net-dhcp-leases default 2>/dev/null | grep -i '$VM_MAC' | grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' | head -1" || true)
    fi
fi

if [ -z "$VM_IP" ]; then
    echo "Error: Could not determine IP address for VM '$VM_NAME'"
    echo ""
    echo "You can manually set the IP with:"
    echo "  WIN_VM_IP=<ip> $0"
    echo ""
    echo "Or check VM status with:"
    echo "  ssh $LIBVIRT_HOST 'virsh domifaddr $VM_NAME --source agent'"
    exit 1
fi

# Allow override
VM_IP="${WIN_VM_IP:-$VM_IP}"
echo "VM IP: $VM_IP"
echo ""

# Copy binary to Windows VM via SSH hop through libvirt host
echo "Copying binary to Windows VM..."
echo "Target: $WIN_USER@$VM_IP:$WIN_DEPLOY_PATH"

# SSH options for password auth
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

# Create deploy directory on Windows if it doesn't exist
sshpass -p "$WIN_PASS" ssh $SSH_OPTS -J "$LIBVIRT_HOST" "$WIN_USER@$VM_IP" "if not exist \"$WIN_DEPLOY_PATH\" mkdir \"$WIN_DEPLOY_PATH\"" 2>/dev/null || true

# Copy the binary
sshpass -p "$WIN_PASS" scp $SSH_OPTS -o "ProxyJump=$LIBVIRT_HOST" "$BINARY_PATH" "$WIN_USER@$VM_IP:$WIN_DEPLOY_PATH\\daemon-win.exe"

echo ""
echo "=== Deployment Complete ==="
echo "Binary deployed to: $WIN_DEPLOY_PATH\\daemon-win.exe"
echo ""
echo "To run the daemon, use:"
echo "  nx run-remote daemon-win"
echo "  # or directly:"
echo "  ssh -J $LIBVIRT_HOST $WIN_USER@$VM_IP '$WIN_DEPLOY_PATH\\daemon-win.exe'"


