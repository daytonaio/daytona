#!/bin/bash
# Script to create a temporary VM for configuring the Windows base image
# Uses linked clone (overlay) - same as runner-win does

set -e

LIBVIRT_HOST="root@h1001.blinkbox.dev"
BASE_IMAGE="/var/lib/libvirt/images/winserver-desktop-base.qcow2"
TEMPLATE_NVRAM="/var/lib/libvirt/qemu/nvram/winserver-desktop-base_VARS.fd"
WORK_IMAGE="/var/lib/libvirt/images/winserver-desktop-config.qcow2"
WORK_NVRAM="/var/lib/libvirt/qemu/nvram/winserver-desktop-config_VARS.fd"
VM_NAME="winserver-desktop-config"

echo "=== Windows Base Image Configuration Script ==="
echo ""
echo "Libvirt Host: $LIBVIRT_HOST"
echo "Base Image:   $BASE_IMAGE"
echo ""

case "${1:-create}" in
    create)
        echo "Step 1: Creating linked clone (overlay) from base image..."
        ssh "$LIBVIRT_HOST" "qemu-img create -f qcow2 -F qcow2 -b '$BASE_IMAGE' '$WORK_IMAGE' && chown libvirt-qemu:kvm '$WORK_IMAGE'"
        
        echo "Step 2: Copying NVRAM..."
        ssh "$LIBVIRT_HOST" "cp '$TEMPLATE_NVRAM' '$WORK_NVRAM' && chown libvirt-qemu:kvm '$WORK_NVRAM'"
        
        echo "Step 3: Creating VM definition..."
        cat << 'VMXML' | ssh "$LIBVIRT_HOST" "cat > /tmp/winserver-config.xml"
<domain type='kvm'>
  <name>winserver-desktop-config</name>
  <memory unit='GiB'>8</memory>
  <vcpu>4</vcpu>
  <os firmware='efi'>
    <type arch='x86_64' machine='q35'>hvm</type>
    <loader readonly='yes' secure='yes' type='pflash'>/usr/share/OVMF/OVMF_CODE_4M.ms.fd</loader>
    <nvram template='/usr/share/OVMF/OVMF_VARS_4M.ms.fd'>/var/lib/libvirt/qemu/nvram/winserver-desktop-config_VARS.fd</nvram>
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
      <source file='/var/lib/libvirt/images/winserver-desktop-config.qcow2'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'/>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
  </devices>
</domain>
VMXML
        
        echo "Step 4: Defining and starting VM..."
        ssh "$LIBVIRT_HOST" "virsh define /tmp/winserver-config.xml && virsh start $VM_NAME"
        
        echo ""
        echo "=== VM Created Successfully ==="
        echo ""
        
        # Get VNC port
        VNC_PORT=$(ssh "$LIBVIRT_HOST" "virsh vncdisplay $VM_NAME" | grep -oE '[0-9]+')
        ACTUAL_PORT=$((5900 + VNC_PORT))
        
        echo "VNC Connection Info:"
        echo "  Host: h1001.blinkbox.dev"
        echo "  Display: :$VNC_PORT"
        echo "  Port: $ACTUAL_PORT"
        echo ""
        echo "To connect via VNC:"
        echo "  1. SSH tunnel: ssh -L $ACTUAL_PORT:localhost:$ACTUAL_PORT $LIBVIRT_HOST"
        echo "  2. VNC client: Connect to localhost:$ACTUAL_PORT"
        echo ""
        echo "Or use virt-viewer:"
        echo "  virt-viewer -c qemu+ssh://$LIBVIRT_HOST/system $VM_NAME"
        echo ""
        echo "=== COMMANDS TO RUN INSIDE WINDOWS (as Administrator) ==="
        echo ""
        cat << 'WINCMD'
# Open PowerShell as Administrator and run:

# 1. Remove Administrator password
net user Administrator ""

# 2. Enable auto-login
reg add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" /v AutoAdminLogon /t REG_SZ /d 1 /f
reg add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" /v DefaultUserName /t REG_SZ /d Administrator /f
reg add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" /v DefaultPassword /t REG_SZ /d "" /f
reg add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" /v DefaultDomainName /t REG_SZ /d . /f

# 3. Disable lock screen
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\Personalization" /v NoLockScreen /t REG_DWORD /d 1 /f

# 4. Disable Ctrl+Alt+Del requirement
reg add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" /v DisableCAD /t REG_DWORD /d 1 /f

# 5. Disable screen saver lock
reg add "HKCU\Control Panel\Desktop" /v ScreenSaverIsSecure /t REG_SZ /d 0 /f

# 6. Shutdown cleanly
shutdown /s /t 0
WINCMD
        echo ""
        echo "After shutdown, run: $0 finalize"
        ;;
        
    finalize)
        echo "Finalizing: Committing overlay changes to base image..."
        
        # Make sure VM is stopped
        ssh "$LIBVIRT_HOST" "virsh destroy $VM_NAME 2>/dev/null || true"
        ssh "$LIBVIRT_HOST" "virsh undefine $VM_NAME 2>/dev/null || true"
        
        echo "Committing overlay changes back to base image..."
        ssh "$LIBVIRT_HOST" "qemu-img commit '$WORK_IMAGE'"
        
        echo "Updating NVRAM template..."
        ssh "$LIBVIRT_HOST" "cp '$WORK_NVRAM' '$TEMPLATE_NVRAM'"
        
        echo "Cleaning up overlay..."
        ssh "$LIBVIRT_HOST" "rm -f '$WORK_IMAGE' '$WORK_NVRAM'"
        
        echo ""
        echo "=== Base image updated successfully! ==="
        echo "Changes committed to: $BASE_IMAGE"
        ;;
        
    cleanup)
        echo "Cleaning up temporary VM..."
        ssh "$LIBVIRT_HOST" "virsh destroy $VM_NAME 2>/dev/null || true"
        ssh "$LIBVIRT_HOST" "virsh undefine $VM_NAME 2>/dev/null || true"
        ssh "$LIBVIRT_HOST" "rm -f '$WORK_IMAGE' '$WORK_NVRAM'"
        echo "Cleanup complete."
        ;;
        
    status)
        echo "VM Status:"
        ssh "$LIBVIRT_HOST" "virsh dominfo $VM_NAME 2>/dev/null || echo 'VM not found'"
        echo ""
        echo "VNC Display:"
        ssh "$LIBVIRT_HOST" "virsh vncdisplay $VM_NAME 2>/dev/null || echo 'VM not running'"
        ;;
        
    vnc)
        echo "Opening VNC tunnel..."
        VNC_PORT=$(ssh "$LIBVIRT_HOST" "virsh vncdisplay $VM_NAME" | grep -oE '[0-9]+')
        ACTUAL_PORT=$((5900 + VNC_PORT))
        echo "Tunneling port $ACTUAL_PORT - connect VNC client to localhost:$ACTUAL_PORT"
        ssh -L "$ACTUAL_PORT:localhost:$ACTUAL_PORT" "$LIBVIRT_HOST"
        ;;
        
    *)
        echo "Usage: $0 {create|finalize|cleanup|status|vnc}"
        echo ""
        echo "  create   - Create temporary VM from base image copy"
        echo "  finalize - Replace base image with configured image (after shutdown)"
        echo "  cleanup  - Remove temporary VM and image (discard changes)"
        echo "  status   - Show VM status and VNC info"
        echo "  vnc      - Open SSH tunnel for VNC access"
        ;;
esac

