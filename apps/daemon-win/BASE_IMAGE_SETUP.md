# Windows Sandbox Base Image Setup

This document provides instructions for setting up and configuring the Windows base image used for Daytona sandboxes.

## Overview

The Windows sandbox base image is a pre-configured Windows Server 2022 QCOW2 disk image that includes:

- Daytona daemon with all toolbox features
- Auto-logon configuration for headless operation
- VNC server for remote desktop access
- Proper session configuration for computer use features

## Prerequisites

### On the Runner Host (Linux)

```bash
# Required packages
apt-get install -y libvirt-daemon-system qemu-kvm ovmf

# Verify OVMF (UEFI firmware) is available
ls /usr/share/OVMF/OVMF_VARS_4M.fd
```

### Base Image Location

The runner expects the base image at:

```
/var/lib/libvirt/images/winserver-autologin-base.qcow2
```

If your image is stored elsewhere, create a symlink:

```bash
ln -sf /path/to/your/winserver-autologin-base.qcow2 \
       /var/lib/libvirt/images/winserver-autologin-base.qcow2
```

### NVRAM Template

UEFI boot requires an NVRAM template:

```bash
mkdir -p /var/lib/libvirt/qemu/nvram
cp /usr/share/OVMF/OVMF_VARS_4M.fd \
   /var/lib/libvirt/qemu/nvram/winserver-autologin-base_VARS.fd
```

---

## Critical Configuration: Interactive Session

### Understanding Windows Sessions

Windows uses session isolation for security:

| Session | Type | Desktop Access | Use Case |
|---------|------|----------------|----------|
| Session 0 | Non-interactive | ❌ No | Windows Services |
| Session 1+ | Interactive | ✅ Yes | User logon sessions |

**The daemon MUST run in Session 1 (interactive) for computer use features to work.**

If the daemon runs in Session 0, these operations will fail:

- Screenshots (`BitBlt failed`)
- Mouse control (`This operation requires an interactive window station`)
- Keyboard input (`Access is denied`)

### Solution: Auto-Logon + Interactive Scheduled Task

The base image must be configured with:

1. **Auto-logon**: Windows automatically logs in as Administrator at boot
2. **Interactive scheduled task**: Daemon starts in the user's interactive session

---

## Step-by-Step Base Image Configuration

### 1. Configure Auto-Logon

Enable automatic logon so Windows boots directly to the desktop:

```powershell
# Run in PowerShell as Administrator
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" -Name "AutoAdminLogon" -Value "1" -Type String
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" -Name "DefaultUserName" -Value "Administrator" -Type String
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon" -Name "DefaultPassword" -Value "YourPassword" -Type String
```

### 2. Create Interactive Daemon Scheduled Task

Create a scheduled task that runs the daemon in the user's interactive session:

```powershell
# Remove any existing task
Unregister-ScheduledTask -TaskName "DaytonaDaemon" -Confirm:$false -ErrorAction SilentlyContinue

# Create task action
$action = New-ScheduledTaskAction -Execute "C:\daemon-win.exe"

# Trigger at user logon
$trigger = New-ScheduledTaskTrigger -AtLogOn -User "Administrator"

# CRITICAL: Use Interactive logon type for GUI access
$principal = New-ScheduledTaskPrincipal -UserId "Administrator" -LogonType Interactive -RunLevel Highest

# Task settings
$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -StartWhenAvailable `
    -ExecutionTimeLimit ([TimeSpan]::Zero)

# Register the task
Register-ScheduledTask -TaskName "DaytonaDaemon" `
    -Action $action `
    -Trigger $trigger `
    -Principal $principal `
    -Settings $settings `
    -Force
```

**Key parameter**: `-LogonType Interactive` ensures the daemon runs in the user's session with desktop access.

### 3. Alternative: Startup Folder (Backup Method)

As a backup, also add a startup script:

```powershell
$startupPath = "C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp"
$batchContent = @"
@echo off
start /b C:\daemon-win.exe
"@
Set-Content -Path "$startupPath\start-daemon.bat" -Value $batchContent
```

**Note**: The Startup folder runs after user logon, which is ideal for interactive session requirements.

### 4. Install VNC Server (TightVNC)

```powershell
# Download and install TightVNC silently
$vncInstaller = "tightvnc-2.8.85-gpl-setup-64bit.msi"
Invoke-WebRequest -Uri "https://www.tightvnc.com/download/$vncInstaller" -OutFile "C:\$vncInstaller"

msiexec /i "C:\$vncInstaller" /quiet /norestart `
    ADDLOCAL="Server" `
    SET_USEVNCAUTHENTICATION=1 `
    VALUE_OF_USEVNCAUTHENTICATION=1 `
    SET_PASSWORD=1 `
    VALUE_OF_PASSWORD=daytona `
    SET_USECONTROLAUTHENTICATION=1 `
    VALUE_OF_USECONTROLAUTHENTICATION=1 `
    SET_CONTROLPASSWORD=1 `
    VALUE_OF_CONTROLPASSWORD=daytona
```

### 5. Install noVNC and websockify

```powershell
# Install Python (for websockify)
# Download from python.org and install

# Install websockify
pip install websockify

# Download noVNC
Invoke-WebRequest -Uri "https://github.com/novnc/noVNC/archive/refs/tags/v1.4.0.zip" -OutFile "C:\novnc.zip"
Expand-Archive -Path "C:\novnc.zip" -DestinationPath "C:\"
Rename-Item "C:\noVNC-1.4.0" "C:\noVNC"

# Create websockify startup script
$websockifyScript = @"
@echo off
cd C:\noVNC
python -m websockify --web . 6080 localhost:5900
"@
Set-Content -Path "C:\start-novnc.bat" -Value $websockifyScript
```

### 6. Copy Daemon Binary

```powershell
# Copy daemon to C:\
Copy-Item "path\to\daemon-win.exe" -Destination "C:\daemon-win.exe"
```

### 7. Configure Windows Firewall

The daemon auto-configures firewall rules on startup, but you can pre-configure them:

```powershell
# Daemon API
New-NetFirewallRule -DisplayName "Daytona Daemon" -Direction Inbound -Port 2280 -Protocol TCP -Action Allow

# SSH Server
New-NetFirewallRule -DisplayName "Daytona SSH" -Direction Inbound -Port 22220 -Protocol TCP -Action Allow

# Web Terminal
New-NetFirewallRule -DisplayName "Daytona Terminal" -Direction Inbound -Port 22222 -Protocol TCP -Action Allow

# VNC
New-NetFirewallRule -DisplayName "TightVNC" -Direction Inbound -Port 5900 -Protocol TCP -Action Allow

# noVNC
New-NetFirewallRule -DisplayName "noVNC" -Direction Inbound -Port 6080 -Protocol TCP -Action Allow
```

### 8. Verify Configuration

After rebooting, verify the daemon is running in the correct session:

```powershell
# Check daemon process session
Get-Process -Name daemon-win | Select-Object Id, SessionId

# Expected output:
# Id   SessionId
# --   ---------
# 1234 1          <-- Session 1 = Interactive ✅
```

If `SessionId` is `0`, the daemon is in the wrong session and computer use will fail.

---

## Updating an Existing Sandbox

To update the daemon in an existing sandbox:

### 1. Build New Daemon

```bash
cd /workspaces/daytona
yarn nx build-windows daemon-win
```

### 2. Upload via Daemon API

```bash
VM_IP="10.100.x.x"
cat dist/apps/daemon-win.exe | \
  ssh hypervisor "curl -s -X POST 'http://$VM_IP:2280/files/upload?path=C%3A%5Cdaemon-win.exe.new' -F 'file=@-'"
```

### 3. Replace via WinRM

```python
import winrm
s = winrm.Session(f'http://{vm_ip}:5985/wsman', auth=('Administrator', 'Password'), transport='ntlm')

# Stop daemon
s.run_cmd('taskkill', ['/F', '/IM', 'daemon-win.exe'])

# Replace binary
s.run_cmd('copy', ['C:\\daemon-win.exe.new', 'C:\\daemon-win.exe', '/Y'])

# Reboot to restart in interactive session
s.run_cmd('shutdown', ['/r', '/t', '3', '/f'])
```

---

## Creating a New Base Image

### From an Existing VM

1. Configure the VM as described above
2. Shut down the VM cleanly:

   ```bash
   virsh shutdown <vm-id>
   ```

3. Wait for shutdown to complete:

   ```bash
   while [ "$(virsh domstate <vm-id>)" != "shut off" ]; do sleep 3; done
   ```

4. Create the base image:

   ```bash
   qemu-img convert -p -O qcow2 \
     /var/lib/libvirt/images/<vm-id>.qcow2 \
     /var/lib/libvirt/images/winserver-autologin-base.qcow2
   ```

### Verify Base Image

```bash
# Check image info
qemu-img info /var/lib/libvirt/images/winserver-autologin-base.qcow2

# Expected: qcow2 format, ~15-25GB virtual size
```

---

## Troubleshooting

### Computer Use Returns Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `BitBlt failed` | Daemon in Session 0 | Configure interactive scheduled task |
| `This operation requires an interactive window station` | Session 0 isolation | Reboot with auto-logon enabled |
| `Access is denied` | Insufficient privileges | Run with `-RunLevel Highest` |

### Verify Session Configuration

```powershell
# Check current user sessions
query user

# Expected output:
# USERNAME              SESSIONNAME        ID  STATE
# administrator         console            1   Active

# Check daemon session
Get-Process -Name daemon-win | Select-Object Id, SessionId

# Expected: SessionId = 1 (not 0)
```

### Daemon Not Starting After Reboot

1. Check scheduled task:

   ```powershell
   Get-ScheduledTask -TaskName "DaytonaDaemon" | Select-Object TaskName, State
   ```

2. Check task history:

   ```powershell
   Get-ScheduledTaskInfo -TaskName "DaytonaDaemon"
   ```

3. Manually run task:

   ```powershell
   Start-ScheduledTask -TaskName "DaytonaDaemon"
   ```

### VNC Not Working

1. Check TightVNC service:

   ```powershell
   Get-Service -Name "tvnserver"
   ```

2. Check noVNC/websockify:

   ```powershell
   Get-Process -Name python | Where-Object { $_.CommandLine -like "*websockify*" }
   ```

---

## Quick Reference

### Ports

| Port | Service | Protocol |
|------|---------|----------|
| 2280 | Daemon API | HTTP |
| 22220 | SSH Server | SSH |
| 22222 | Web Terminal | HTTP/WebSocket |
| 5900 | TightVNC | VNC |
| 6080 | noVNC | HTTP/WebSocket |

### File Locations

| File | Path |
|------|------|
| Daemon binary | `C:\daemon-win.exe` |
| noVNC | `C:\noVNC\` |
| VNC password | Stored in TightVNC registry |
| Startup script | `C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\start-daemon.bat` |

### Registry Keys

| Key | Purpose |
|-----|---------|
| `HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\AutoAdminLogon` | Enable auto-logon |
| `HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\DefaultUserName` | Auto-logon username |
| `HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\DefaultPassword` | Auto-logon password |

---

## Related Documentation

- Feature documentation: `apps/daemon-win/FUTURE_FEATURES.md`
- Runner deployment: `.tmp/deploy/RUNNER-WIN-DEPLOYMENT.md`
- Linux daemon: `apps/daemon/`
