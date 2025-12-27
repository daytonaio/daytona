#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Run daemon-win.exe on win11-clone VM via SSH
# This script requires SSH access to h1001.blinkbox.dev and assumes
# OpenSSH Server is enabled on the Windows VM.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Configuration
LIBVIRT_HOST="h1001.blinkbox.dev"
VM_NAME="${WIN_VM_NAME:-winserver-core}"
WIN_USER="${WIN_USER:-Administrator}"
WIN_PASS="${WIN_PASS:-DaytonaWinAcc3ss!}"
WIN_DEPLOY_PATH="${WIN_DEPLOY_PATH:-C:\\daytona}"

echo "=== Running Daytona Windows Daemon ==="
echo "Libvirt Host: $LIBVIRT_HOST"
echo "VM Name: $VM_NAME"
echo ""

# Get VM IP address from libvirt
get_vm_ip() {
    local ip
    ip=$(ssh "$LIBVIRT_HOST" "virsh domifaddr $VM_NAME --source agent 2>/dev/null | grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' | head -1" || true)
    
    if [ -z "$ip" ]; then
        # Fallback: try to get IP from network leases
        local mac
        mac=$(ssh "$LIBVIRT_HOST" "virsh domiflist $VM_NAME | grep -oE '([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}' | head -1" || true)
        if [ -n "$mac" ]; then
            ip=$(ssh "$LIBVIRT_HOST" "virsh net-dhcp-leases default 2>/dev/null | grep -i '$mac' | grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' | head -1" || true)
        fi
    fi
    
    echo "$ip"
}

VM_IP="${WIN_VM_IP:-$(get_vm_ip)}"

if [ -z "$VM_IP" ]; then
    echo "Error: Could not determine IP address for VM '$VM_NAME'"
    echo "Set WIN_VM_IP environment variable manually."
    exit 1
fi

echo "VM IP: $VM_IP"
echo "Executing daemon-win.exe..."
echo ""
echo "--- Remote Output ---"

# SSH options for password auth
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

# Execute the daemon via SSH hop
sshpass -p "$WIN_PASS" ssh $SSH_OPTS -J "$LIBVIRT_HOST" "$WIN_USER@$VM_IP" "$WIN_DEPLOY_PATH\\daemon-win.exe"


