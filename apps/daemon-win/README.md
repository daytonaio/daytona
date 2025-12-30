# Daytona Windows Daemon

A Windows Go application for Daytona platform development. This daemon runs inside Windows VMs (sandboxes) and provides the Toolbox API for process execution, file operations, git commands, and remote desktop access via VNC.

## Features

- **SDK Compatibility**: Transparently handles Linux-style command wrappers (`sh -c "..."`) from the Python/TypeScript SDKs
- **Auto Firewall Configuration**: Automatically adds Windows Firewall rule on startup (port 2280)
- **Process Execution**: Execute commands via PowerShell with proper output capture
- **Session Management**: Persistent shell sessions for command execution
- **File System Operations**: Create, read, write, delete files and directories
- **Git Integration**: Clone, commit, push, pull, and branch operations
- **Computer Use Status**: Check VNC availability via `/computeruse/status` endpoint
- **Remote Desktop**: Web-based VNC access via noVNC on port 6080

## Architecture

The daemon is deployed as a scheduled task (`DaytonaDaemon`) that:

1. Listens on port 2280 for HTTP API requests
2. Automatically configures Windows Firewall on first start
3. Parses SDK command wrappers to extract actual commands
4. Executes commands via PowerShell

### Remote Desktop Stack

```
Browser ─────► noVNC (6080) ─────► websockify ─────► TightVNC (5900)
               (Web Client)        (WebSocket)        (VNC Server)
```

- **TightVNC Server**: Runs on port 5900, provides RFB protocol
- **websockify**: Python bridge that converts WebSocket to VNC protocol
- **noVNC**: Web-based VNC client accessible via browser

## Quick Start

### 1. Build for Windows

```bash
yarn nx build-windows daemon-win
```

This creates `dist/apps/daemon-win.exe` - a Windows AMD64 executable.

### 2. Base Image Setup

The Windows sandbox base image (`winserver-desktop-base.qcow2`) includes:

- Windows Server 2022 with **Desktop Experience** (full GUI)
- TightVNC Server (port 5900, no authentication)
- noVNC + websockify (port 6080)
- Daytona daemon (port 2280)
- Auto-start scheduled tasks for all services
- Firewall disabled for development

### 3. Deploy to Windows VM

```bash
# Auto-detect VM IP
yarn nx deploy daemon-win

# Or specify IP manually
WIN_VM_IP=10.100.12.205 yarn nx deploy daemon-win
```

### 4. Run on Windows VM

```bash
yarn nx run-remote daemon-win
```

## Configuration

### Deployment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WIN_VM_NAME` | `winserver-desktop` | VM name in libvirt |
| `WIN_VM_IP` | (auto-detected) | Override VM IP address |
| `WIN_USER` | `Administrator` | Windows SSH username |
| `WIN_PASS` | `Daytona123!` | Windows SSH/RDP password |
| `WIN_DEPLOY_PATH` | `C:\daytona` | Deployment directory on Windows |

### Runner Environment Variables

These are used by runner-win when connecting to the daemon:

| Variable | Default | Description |
|----------|---------|-------------|
| `LIBVIRT_URI` | `qemu:///system` | Libvirt connection URI |
| `LIBVIRT_SSH_TUNNEL` | `true` (if remote) | Enable/disable SSH tunneling for proxy |
| `DAEMON_START_TIMEOUT_SEC` | `60` | Seconds to wait for daemon startup |

### Service Ports

| Port | Service | Description |
|------|---------|-------------|
| 2280 | Daytona Daemon | Toolbox API |
| 5900 | TightVNC | VNC server (RFB protocol) |
| 6080 | noVNC | Web-based VNC client |
| 3389 | RDP | Remote Desktop (optional) |

## Available Targets

| Target | Description |
|--------|-------------|
| `build` | Build for current platform (Linux) |
| `build-windows` | Cross-compile for Windows AMD64 |
| `deploy` | Build and copy to Windows VM |
| `run-remote` | Deploy and execute on Windows VM |
| `serve` | Local development with hot-reload |
| `test` | Run Go tests |
| `lint` | Run Go linter |

## Development Workflow

```
Linux DevContainer                    h1001.blinkbox.dev
┌─────────────────┐                   ┌─────────────────┐
│  Edit Go code   │                   │ libvirt/QEMU    │
│        ↓        │                   │       ↓         │
│  nx build-win   │ ──── SSH ────►    │  Windows VM     │
│        ↓        │                   │  (Desktop Exp)  │
│  nx deploy      │ ──── SCP ────►    │  daemon-win.exe │
│        ↓        │                   │  TightVNC       │
│  nx run-remote  │ ◄─── Output ────  │  noVNC          │
└─────────────────┘                   └─────────────────┘
```

### Remote Libvirt Development

When developing with libvirt on a remote machine (e.g., h1001.blinkbox.dev):

1. **SSH Configuration**: Set up SSH config for passwordless access:

   ```bash
   # ~/.ssh/config (and /root/.ssh/config for sudo)
   Host h1001.blinkbox.dev
       IdentityFile /workspaces/daytona/.tmp/ssh/id_rsa
       StrictHostKeyChecking no
       BatchMode yes
   ```

2. **Runner Configuration**: Set `LIBVIRT_URI` to use SSH:

   ```bash
   export LIBVIRT_URI="qemu+ssh://root@h1001.blinkbox.dev/system"
   ```

3. **SSH Tunneling**: The runner-win automatically creates SSH tunnels to reach Windows VMs on the remote hypervisor. This is handled by `pkg/libvirt/ssh_tunnel.go`.

4. **Starting Development Server**:

   ```bash
   yarn serve
   ```

   The runner will:
   - Connect to libvirt over SSH
   - Create Windows sandboxes from the base image
   - Proxy API requests to the daemon via SSH tunnels

## API Endpoints

### Computer Use

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/computeruse/status` | Returns `{"status": "active"}` if VNC is running |
| POST | `/computeruse/start` | Start computer use (placeholder) |
| POST | `/computeruse/stop` | Stop computer use (placeholder) |

### VNC Access via Proxy

Access noVNC through the runner proxy:

```
http://localhost:3000/api/toolbox/{sandbox-id}/toolbox/6080/vnc.html
```

This proxies to the noVNC web client running on port 6080 inside the Windows VM.

## Troubleshooting

### Cannot connect to Windows VM

1. Check VM is running:

   ```bash
   ssh h1001.blinkbox.dev "virsh list --all"
   ```

2. Get VM IP:

   ```bash
   ssh h1001.blinkbox.dev "virsh domifaddr winserver-desktop"
   ```

3. Test connectivity:

   ```bash
   ssh h1001.blinkbox.dev "curl -s http://VM_IP:2280/health"
   ```

### VNC shows black screen

1. Verify Windows has Desktop Experience (not Server Core):

   ```bash
   # Via daemon API
   curl -X POST "http://VM_IP:2280/process/execute" \
     -H "Content-Type: application/json" \
     -d '{"command": "Get-ComputerInfo | Select WindowsInstallationType"}'
   ```

   Should return `Server` (Desktop Experience), not `Server Core`.

2. Check TightVNC is running:

   ```bash
   curl -X POST "http://VM_IP:2280/process/execute" \
     -H "Content-Type: application/json" \
     -d '{"command": "Get-Service tvnserver | Select Status"}'
   ```

3. Check noVNC/websockify:

   ```bash
   curl -X POST "http://VM_IP:2280/process/execute" \
     -H "Content-Type: application/json" \
     -d '{"command": "netstat -an | Select-String 6080"}'
   ```

### Daemon not starting on boot

Check the scheduled task:

```powershell
Get-ScheduledTask -TaskName "DaytonaDaemon"
Start-ScheduledTask -TaskName "DaytonaDaemon"
```

### Cross-compilation issues

Ensure CGO is disabled for Windows builds:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o daemon-win.exe ./cmd/daemon-win
```

## SDK Command Compatibility

The Python and TypeScript SDKs wrap commands for safe execution on Linux:

```
sh -c "echo 'BASE64_CMD' | base64 -d | sh"
```

The Windows daemon automatically:

1. Detects this pattern via `pkg/common/command_parser.go`
2. Extracts and decodes the base64 command
3. Extracts environment variables from `export KEY=$(echo 'BASE64' | base64 -d)`
4. Executes the actual command via PowerShell

This allows the SDK to remain unchanged while supporting both Linux and Windows sandboxes.

## Base Image Setup

To create or update the Windows base image for sandboxes:

### Prerequisites

- Windows Server 2022 VM with **Desktop Experience** (not Server Core)
- Administrator password: `Daytona123!`

### Components to Install

1. **Daytona Daemon**:

   ```powershell
   # Download daemon
   Invoke-WebRequest -Uri "http://10.100.0.1:8888/daemon-win.exe" -OutFile "C:\daemon-win.exe"
   
   # Create auto-start task
   $action = New-ScheduledTaskAction -Execute "C:\daemon-win.exe"
   $trigger = New-ScheduledTaskTrigger -AtStartup
   $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
   Register-ScheduledTask -TaskName "DaytonaDaemon" -Action $action -Trigger $trigger -Settings $settings -User "SYSTEM" -RunLevel Highest -Force
   ```

2. **TightVNC Server** (no authentication):

   ```powershell
   # Download TightVNC
   Invoke-WebRequest -Uri "https://www.tightvnc.com/download/2.8.85/tightvnc-2.8.85-gpl-setup-64bit.msi" -OutFile "$env:TEMP\tightvnc.msi"
   
   # Install silently without password, allow loopback
   Start-Process msiexec.exe -ArgumentList "/i `"$env:TEMP\tightvnc.msi`" /quiet /norestart ADDLOCAL=Server SET_USEVNCAUTHENTICATION=1 VALUE_OF_USEVNCAUTHENTICATION=0 SET_ALLOWLOOPBACK=1 VALUE_OF_ALLOWLOOPBACK=1" -Wait
   ```

3. **Python + noVNC + websockify**:

   ```powershell
   # Install Python
   Invoke-WebRequest -Uri "https://www.python.org/ftp/python/3.12.0/python-3.12.0-amd64.exe" -OutFile "$env:TEMP\python.exe"
   Start-Process "$env:TEMP\python.exe" -ArgumentList '/quiet', 'InstallAllUsers=1', 'PrependPath=1' -Wait
   
   # Refresh PATH
   $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine")
   
   # Install websockify
   & "C:\Program Files\Python312\Scripts\pip.exe" install websockify
   
   # Download noVNC
   Invoke-WebRequest -Uri "https://github.com/novnc/noVNC/archive/refs/tags/v1.4.0.zip" -OutFile "$env:TEMP\novnc.zip"
   Expand-Archive -Path "$env:TEMP\novnc.zip" -DestinationPath "C:\" -Force
   Rename-Item "C:\noVNC-1.4.0" "C:\noVNC"
   
   # Create noVNC auto-start task
   $action = New-ScheduledTaskAction -Execute "C:\Program Files\Python312\python.exe" -Argument "-m websockify --web C:\noVNC 6080 localhost:5900"
   $trigger = New-ScheduledTaskTrigger -AtStartup
   Register-ScheduledTask -TaskName "noVNC" -Action $action -Trigger $trigger -Settings $settings -User "SYSTEM" -RunLevel Highest -Force
   ```

4. **Disable Firewall** (for development):

   ```powershell
   Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled False
   ```

### Creating the Base Image

1. Shut down the VM:

   ```bash
   virsh shutdown winserver-desktop
   ```

2. Create standalone base image (flatten backing chain):

   ```bash
   qemu-img convert -O qcow2 /var/lib/libvirt/images/winserver-desktop.qcow2 /var/lib/libvirt/images/winserver-desktop-base.qcow2
   ```

3. Copy NVRAM template:

   ```bash
   cp /var/lib/libvirt/qemu/nvram/winserver-desktop_VARS.fd /var/lib/libvirt/qemu/nvram/winserver-desktop-base_VARS.fd
   ```

4. Set permissions:

   ```bash
   chown libvirt-qemu:kvm /var/lib/libvirt/images/winserver-desktop-base.qcow2
   chown libvirt-qemu:kvm /var/lib/libvirt/qemu/nvram/winserver-desktop-base_VARS.fd
   ```

### Verifying the Base Image

New sandboxes created from the base image should:

- Have the daemon running on port 2280
- Have TightVNC running on port 5900
- Have noVNC/websockify running on port 6080
- Return `{"status": "active"}` from `/computeruse/status`
- Show Windows desktop via noVNC at `http://VM_IP:6080/vnc.html`
