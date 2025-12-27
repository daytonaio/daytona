# Daytona Windows Daemon

A Windows Go application for Daytona platform development. This project provides cross-compilation support and deployment tooling to run Go apps on Windows VMs hosted on h1001.blinkbox.dev.

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

Environment variables for deployment:

| Variable | Default | Description |
|----------|---------|-------------|
| `WIN_VM_NAME` | `winserver-core` | VM name in libvirt |
| `WIN_VM_IP` | (auto-detected) | Override VM IP address |
| `WIN_USER` | `Administrator` | Windows SSH username |
| `WIN_PASS` | `DaytonaWinAcc3ss!` | Windows SSH password |
| `WIN_DEPLOY_PATH` | `C:\daytona` | Deployment directory on Windows |

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
