# Daytona Windows Daemon - Future Features

This document lists components and features that are planned for future implementation in the Windows daemon. These features exist in the Linux daemon (`apps/daemon`) and should be ported with Windows-specific adaptations.

## Recently Implemented

- ✅ **SDK Command Compatibility** (`pkg/common/command_parser.go`): Transparent handling of Linux-style `sh -c "..."` wrappers from Python/TypeScript SDKs
- ✅ **Auto Firewall Rule**: Daemon automatically configures Windows Firewall on startup (ports 2280, 22220, 22222)
- ✅ **Process Execution**: PowerShell-based command execution with proper output capture
- ✅ **Session Management**: Persistent shell sessions for command execution
- ✅ **Remote Desktop via VNC**: TightVNC + noVNC for web-based remote desktop access
- ✅ **WebSocket Proxy Support**: Proxy correctly handles WebSocket connections for noVNC
- ✅ **SSH Server** (`pkg/ssh/`): Built-in SSH server with password/public key auth, SFTP, TCP port forwarding, and ConPTY interactive shells
- ✅ **Web Terminal** (`pkg/terminal/`): Browser-based terminal using xterm.js on port 22222 with ConPTY backend
- ✅ **Toolbox Proxy** (`pkg/toolbox/proxy/`): Internal proxy route to access daemon services (e.g., `/proxy/22222/` for terminal)
- ✅ **Computer Use** (`pkg/toolbox/computeruse/`): Full programmatic desktop automation (see below)

---

## Computer Use - Full Implementation ✅

**Reference**: `apps/daemon/pkg/toolbox/computeruse/`

Screen capture and input automation via Windows API syscalls.

### Implementation Details

The Windows computer use implementation uses direct Windows API calls (syscalls) instead of CGO-dependent libraries, enabling cross-compilation from Linux. Key files:

| File | Purpose |
|------|---------|
| `winapi.go` | Windows API syscalls (user32.dll, kernel32.dll) |
| `winapi_stub.go` | Stub implementations for non-Windows builds |
| `screenshot.go` | Screenshot capture using `kbinani/screenshot` |
| `mouse.go` | Mouse control via SendInput API |
| `keyboard.go` | Keyboard control via SendInput API |
| `display.go` | Display and window enumeration |

### API Endpoints

| Endpoint | Status | Description |
|----------|--------|-------------|
| `GET /computeruse/status` | ✅ Done | Returns VNC status (`active`/`inactive`) |
| `POST /computeruse/start` | ✅ Done | Placeholder (VNC auto-starts) |
| `POST /computeruse/stop` | ✅ Done | Placeholder (VNC managed by Windows) |
| `GET /computeruse/screenshot` | ✅ Done | Take full screenshot |
| `GET /computeruse/screenshot/region` | ⚠️ Partial | Region screenshot (may fail on some configs) |
| `GET /computeruse/screenshot/compressed` | ✅ Done | Compressed screenshot (JPEG) |
| `GET /computeruse/mouse/position` | ✅ Done | Get mouse position |
| `POST /computeruse/mouse/move` | ✅ Done | Move mouse |
| `POST /computeruse/mouse/click` | ✅ Done | Click (left/right/double) |
| `POST /computeruse/mouse/drag` | ✅ Done | Drag operation |
| `POST /computeruse/mouse/scroll` | ✅ Done | Scroll wheel |
| `POST /computeruse/keyboard/type` | ✅ Done | Type text |
| `POST /computeruse/keyboard/key` | ✅ Done | Press key |
| `POST /computeruse/keyboard/hotkey` | ✅ Done | Key combination |
| `GET /computeruse/display/info` | ✅ Done | Display information |
| `GET /computeruse/display/windows` | ✅ Done | List windows |

### ⚠️ Critical: Interactive Session Requirement

**Computer use operations require the daemon to run in an interactive Windows session (Session 1), NOT Session 0.**

Windows has session isolation:

- **Session 0**: Non-interactive, used by services. No access to desktop/GUI.
- **Session 1+**: Interactive sessions where users log in. Has desktop access.

If the daemon runs in Session 0 (e.g., as a Windows Service or via certain scheduled tasks), GUI operations will fail with errors like:

- `BitBlt failed` (screenshots)
- `This operation requires an interactive window station` (mouse/keyboard)
- `Access is denied` (input injection)

**Solution**: The daemon uses a two-part approach:

1. **Built as GUI application** (`-H windowsgui` linker flag):
   - No console window appears when daemon starts
   - Users cannot accidentally close the daemon
   - Daemon runs invisibly in the background
   - Logging continues to file (`C:\Windows\Temp\daytona-daemon.log`)

2. **Interactive scheduled task**:
   - Trigger: At user logon
   - Principal: Interactive logon type (`-LogonType Interactive`)
   - Combined with: Auto-logon for the Administrator user

This ensures the daemon runs **invisibly** in Session 1 with full desktop access, and users cannot accidentally terminate it by closing a window.

See `BASE_IMAGE_SETUP.md` for detailed configuration instructions.

### Remote Desktop Stack

```
Browser ─────► noVNC (6080) ─────► websockify ─────► TightVNC (5900)
               (Web Client)        (WebSocket)        (VNC Server)
```

---

## 2. PTY Sessions (Toolbox API)

**Reference**: `apps/daemon/pkg/toolbox/process/pty/`

Interactive pseudo-terminal sessions accessible via the Toolbox API.

### Components to Implement

- **Session Manager**: Create, list, delete PTY sessions
- **ConPTY Session**: Windows pseudo-terminal wrapper
- **WebSocket Client**: Connect to PTY sessions
- **Multi-Client Broadcasting**: Multiple clients per session
- **Resize Support**: Dynamic terminal size changes

### API Endpoints

```
GET    /process/pty           - List PTY sessions
POST   /process/pty           - Create PTY session
GET    /process/pty/:id       - Get session info
DELETE /process/pty/:id       - Delete session
GET    /process/pty/:id/connect - WebSocket connection
POST   /process/pty/:id/resize  - Resize terminal
```

### Windows-Specific Considerations

- ConPTY API for pseudo-terminal creation
- Windows process creation flags
- Handle Windows-specific exit codes

---

## 3. Interpreter REPL

**Reference**: `apps/daemon/pkg/toolbox/process/interpreter/`

Persistent Python interpreter contexts for code execution.

### Components to Implement

- **Context Manager**: Create and manage Python contexts
- **REPL Worker**: Python subprocess for execution
- **WebSocket Interface**: Real-time execution feedback
- **Output Streaming**: Stream stdout/stderr to clients

### API Endpoints

```
POST   /process/interpreter/context     - Create context
GET    /process/interpreter/context     - List contexts
DELETE /process/interpreter/context/:id - Delete context
GET    /process/interpreter/execute     - Execute code (WebSocket)
```

### Windows-Specific Considerations

- Locate Python installation (`python.exe`, `python3.exe`)
- Handle Windows path separators in Python scripts
- Process termination using Windows APIs

---

## 4. LSP Support (Language Server Protocol)

**Reference**: `apps/daemon/pkg/toolbox/lsp/`

Language server integration for code intelligence features.

### Components to Implement

- **LSP Client**: JSON-RPC communication with language servers
- **Server Management**: Start/stop language servers
- **TypeScript LSP**: Integration with `typescript-language-server`
- **Python LSP**: Integration with `pylsp` or `pyright`
- **Document Sync**: Open/close document notifications
- **Completions**: Code completion requests
- **Symbols**: Document and workspace symbol queries

### API Endpoints

```
POST   /lsp/start           - Start language server
POST   /lsp/stop            - Stop language server
POST   /lsp/completions     - Get completions
POST   /lsp/did-open        - Notify document open
POST   /lsp/did-close       - Notify document close
GET    /lsp/document-symbols - Get document symbols
GET    /lsp/workspacesymbols - Get workspace symbols
```

### Windows-Specific Considerations

- Locate Node.js/npm for TypeScript LSP
- Locate Python for Python LSP
- Windows process spawning for language servers
- Handle Windows paths in URI conversions

---

## Implementation Priority Suggestion

1. **PTY Sessions** - Enables interactive terminal via API
2. **LSP Support** - Developer tooling
3. **Interpreter REPL** - Python execution

### Already Implemented

- ✅ **SSH Server** - Built-in SSH with SFTP and port forwarding
- ✅ **Web Terminal** - Browser-based terminal via xterm.js
- ✅ **Computer Use** - Full programmatic desktop automation

---

## Base Image Components

The Windows sandbox base image includes these pre-installed components:

| Component | Version | Port | Purpose |
|-----------|---------|------|---------|
| Windows Server | 2022 Desktop Experience | - | Full GUI support |
| Daytona Daemon | Latest | 2280 | Toolbox API |
| Daytona SSH Server | (built-in) | 22220 | SSH, SFTP, port forwarding |
| Daytona Web Terminal | (built-in) | 22222 | Browser-based terminal |
| TightVNC | 2.8.85 | 5900 | VNC server |
| Python | 3.12 | - | For websockify |
| websockify | Latest | - | WebSocket bridge |
| noVNC | 1.4.0 | 6080 | Web VNC client |

All services are configured to auto-start. The daemon runs via a scheduled task that triggers at user logon with interactive session privileges.

**See `BASE_IMAGE_SETUP.md` for detailed base image configuration.**

---

## Related Documentation

- Base Image Setup: `apps/daemon-win/BASE_IMAGE_SETUP.md`
- Linux Daemon Source: `apps/daemon/`
- Windows ConPTY: https://docs.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
- go-conpty: https://github.com/UserExistsError/conpty
- TightVNC: https://www.tightvnc.com/
- noVNC: https://novnc.com/
- websockify: https://github.com/novnc/websockify
