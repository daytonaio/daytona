#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Deploy daemon-win.exe to Windows VM and configure as a service
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
SERVICE_NAME="DaytonaDaemon"

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

# SSH options for password auth
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

run_remote() {
    sshpass -p "$WIN_PASS" ssh $SSH_OPTS -J "$LIBVIRT_HOST" "$WIN_USER@$VM_IP" "$@" 2>/dev/null
}

run_remote_ps() {
    sshpass -p "$WIN_PASS" ssh $SSH_OPTS -J "$LIBVIRT_HOST" "$WIN_USER@$VM_IP" "powershell -Command \"$@\"" 2>/dev/null
}

# Stop existing service/process to release file lock
echo "Stopping existing daemon..."
run_remote_ps "Stop-Service -Name $SERVICE_NAME -Force -ErrorAction SilentlyContinue" || true
run_remote "taskkill /IM daemon-win.exe /F" || true
sleep 2

# Create deploy directory on Windows if it doesn't exist
echo "Creating deploy directory..."
run_remote "if not exist \"$WIN_DEPLOY_PATH\" mkdir \"$WIN_DEPLOY_PATH\"" || true

# Copy the binary
echo "Copying binary to Windows VM..."
echo "Target: $WIN_USER@$VM_IP:$WIN_DEPLOY_PATH"
sshpass -p "$WIN_PASS" scp $SSH_OPTS -o "ProxyJump=$LIBVIRT_HOST" "$BINARY_PATH" "$WIN_USER@$VM_IP:$WIN_DEPLOY_PATH\\daemon-win.exe"

# Copy service installation scripts
echo "Copying service scripts..."
sshpass -p "$WIN_PASS" scp $SSH_OPTS -o "ProxyJump=$LIBVIRT_HOST" \
    "$SCRIPT_DIR/install-service.ps1" \
    "$WIN_USER@$VM_IP:$WIN_DEPLOY_PATH\\install-service.ps1"

sshpass -p "$WIN_PASS" scp $SSH_OPTS -o "ProxyJump=$LIBVIRT_HOST" \
    "$SCRIPT_DIR/uninstall-service.ps1" \
    "$WIN_USER@$VM_IP:$WIN_DEPLOY_PATH\\uninstall-service.ps1"

# Install/configure as Windows service
echo ""
echo "Installing/configuring Windows service..."
# Run the install script - it handles its own error reporting
sshpass -p "$WIN_PASS" ssh $SSH_OPTS -J "$LIBVIRT_HOST" "$WIN_USER@$VM_IP" "powershell -ExecutionPolicy Bypass -File $WIN_DEPLOY_PATH\\install-service.ps1" || {
    echo "Warning: Service installation may have had issues, checking status..."
}

# Verify service is running
echo ""
echo "Verifying service status..."
SERVICE_STATUS=$(run_remote_ps "Get-Service -Name $SERVICE_NAME -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Status" || echo "NotFound")
SERVICE_STATUS=$(echo "$SERVICE_STATUS" | tr -d '\r\n ')
if [ "$SERVICE_STATUS" = "Running" ]; then
    echo "✓ Service is running successfully!"
else
    echo "⚠ Warning: Service status is '$SERVICE_STATUS'"
fi

echo ""
echo "=== Deployment Complete ==="
echo "Binary deployed to: $WIN_DEPLOY_PATH\\daemon-win.exe"
echo "Service: $SERVICE_NAME (auto-start, auto-restart on crash)"
echo ""
echo "Useful commands:"
echo "  Check status:  ssh ... 'powershell Get-Service $SERVICE_NAME'"
echo "  View logs:     ssh ... 'powershell Get-Content C:\\daytona\\logs\\daemon-stdout.log -Tail 50'"
