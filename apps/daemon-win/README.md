# Daytona Windows Daemon

A Windows Go application for Daytona platform development. This daemon runs inside Windows VMs (sandboxes) and provides the Toolbox API for process execution, file operations, and git commands.

## Features

- **SDK Compatibility**: Transparently handles Linux-style command wrappers (`sh -c "..."`) from the Python/TypeScript SDKs
- **Auto Firewall Configuration**: Automatically adds Windows Firewall rule on startup (port 2280)
- **Process Execution**: Execute commands via PowerShell with proper output capture
- **Session Management**: Persistent shell sessions for command execution
- **File System Operations**: Create, read, write, delete files and directories
- **Git Integration**: Clone, commit, push, pull, and branch operations

## Architecture

The daemon is deployed as a Windows Service (`DaytonaDaemon`) that:

1. Listens on port 2280 for HTTP API requests
2. Automatically configures Windows Firewall on first start
3. Parses SDK command wrappers to extract actual commands
4. Executes commands via PowerShell

## Quick Start

### 1. Build for Windows

```bash
yarn nx build-windows daemon-win
```

This creates `dist/apps/daemon-win.exe` - a Windows AMD64 executable.

### 2. Setup Windows VM (First Time Only)

The Windows VM needs OpenSSH Server enabled. Copy `scripts/setup-windows-ssh.ps1` to the Windows VM and run it as Administrator:

```powershell
# On Windows VM (via VNC/RDP):
PowerShell -ExecutionPolicy Bypass -File setup-windows-ssh.ps1
```

Or manually enable OpenSSH:

1. Settings → Apps → Optional Features → Add a feature → OpenSSH Server
2. Start the service: `Start-Service sshd`
3. Set to auto-start: `Set-Service -Name sshd -StartupType 'Automatic'`

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
| `WIN_VM_NAME` | `winserver-core` | VM name in libvirt |
| `WIN_VM_IP` | (auto-detected) | Override VM IP address |
| `WIN_USER` | `Administrator` | Windows SSH username |
| `WIN_PASS` | `DaytonaWinAcc3ss!` | Windows SSH password |
| `WIN_DEPLOY_PATH` | `C:\daytona` | Deployment directory on Windows |

### Runner Environment Variables

These are used by runner-win when connecting to the daemon:

| Variable | Default | Description |
|----------|---------|-------------|
| `LIBVIRT_URI` | `qemu:///system` | Libvirt connection URI |
| `LIBVIRT_SSH_TUNNEL` | `true` (if remote) | Enable/disable SSH tunneling for proxy |
| `DAEMON_START_TIMEOUT_SEC` | `60` | Seconds to wait for daemon startup |

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
│  nx build-win   │ ──── SSH ────►    │  win11-clone    │
│        ↓        │                   │   (Windows VM)  │
│  nx deploy      │ ──── SCP ────►    │  daemon-win.exe │
│        ↓        │                   │       ↓         │
│  nx run-remote  │ ◄─── Output ────  │   Execution     │
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

## Troubleshooting

### Cannot connect to Windows VM

1. Check VM is running:

   ```bash
   ssh h1001.blinkbox.dev "virsh list --all"
   ```

2. Get VM IP:

   ```bash
   ssh h1001.blinkbox.dev "virsh domifaddr win11-clone --source agent"
   ```

3. Test SSH connectivity:

   ```bash
   ssh -J h1001.blinkbox.dev daytona@<VM_IP> "hostname"
   ```

### SSH connection times out

The Windows VM may not have OpenSSH Server enabled. Connect via VNC/RDP and run the setup script.

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

- Windows Server VM (`winserver-core`) with:
  - OpenSSH Server enabled
  - Daemon installed at `C:\daytona\daemon-win.exe`
  - Service configured via `install-service.ps1`
  - Windows Firewall rule for port 2280

### Updating the Base Image

1. Deploy the new daemon to the running VM:

   ```bash
   # Build for Windows
   GOOS=windows GOARCH=amd64 go build -o /tmp/daemon-win.exe apps/daemon-win/cmd/daemon-win/main.go
   
   # Copy to hypervisor then to VM (replace IP)
   scp /tmp/daemon-win.exe h1001.blinkbox.dev:/tmp/
   ssh h1001.blinkbox.dev "
     sshpass -p 'PASSWORD' scp -o StrictHostKeyChecking=no /tmp/daemon-win.exe Administrator@VM_IP:/C:/daytona/daemon-win.exe
   "
   ```

2. Ensure firewall rule exists:

   ```bash
   ssh h1001.blinkbox.dev "
     sshpass -p 'PASSWORD' ssh Administrator@VM_IP 'netsh advfirewall firewall add rule name=\"Daytona Daemon\" dir=in action=allow protocol=tcp localport=2280'
   "
   ```

3. Shutdown the VM and commit changes:

   ```bash
   ssh h1001.blinkbox.dev "
     # Shutdown gracefully
     virsh shutdown winserver-core
     sleep 60
     
     # Commit overlay to base (if using overlay disk)
     qemu-img commit /var/lib/libvirt/images/winserver-core-overlay.qcow2
     
     # Copy to sandbox base image
     cp /var/lib/libvirt/images/winserver-core.qcow2 /var/lib/libvirt/images/winserver-sandbox-base.qcow2
     
     # Restart VM
     virsh start winserver-core
   "
   ```

### Verifying the Base Image

New sandboxes created from the base image should:

- Have the daemon running on port 2280
- Have the firewall rule configured
- Successfully execute SDK commands (e.g., `sandbox.process.exec('dir')`)
