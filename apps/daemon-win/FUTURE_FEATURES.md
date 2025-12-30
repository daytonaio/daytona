# Daytona Windows Daemon - Future Features

This document lists components and features that are planned for future implementation in the Windows daemon. These features exist in the Linux daemon (`apps/daemon`) and should be ported with Windows-specific adaptations.

## Recently Implemented

- ✅ **SDK Command Compatibility** (`pkg/common/command_parser.go`): Transparent handling of Linux-style `sh -c "..."` wrappers from Python/TypeScript SDKs
- ✅ **Auto Firewall Rule**: Daemon automatically configures Windows Firewall on startup (ports 2280, 22220, 22222)
- ✅ **Process Execution**: PowerShell-based command execution with proper output capture
- ✅ **Session Management**: Persistent shell sessions for command execution
- ✅ **Computer Use Status** (`pkg/toolbox/computeruse/`): VNC-based status endpoint returning `active`/`inactive`
- ✅ **Remote Desktop via VNC**: TightVNC + noVNC for web-based remote desktop access
- ✅ **WebSocket Proxy Support**: Proxy correctly handles WebSocket connections for noVNC
- ✅ **SSH Server** (`pkg/ssh/`): Built-in SSH server with password/public key auth, SFTP, TCP port forwarding, and ConPTY interactive shells
- ✅ **Web Terminal** (`pkg/terminal/`): Browser-based terminal using xterm.js on port 22222 with ConPTY backend
- ✅ **Toolbox Proxy** (`pkg/toolbox/proxy/`): Internal proxy route to access daemon services (e.g., `/proxy/22222/` for terminal)

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

## 4. Computer Use - Full Implementation

**Reference**: `apps/daemon/pkg/toolbox/computeruse/`

Screen capture and input automation. Currently implemented as VNC-based remote desktop; full programmatic control is planned.

### Currently Implemented ✅

| Endpoint | Status | Description |
|----------|--------|-------------|
| `GET /computeruse/status` | ✅ Done | Returns VNC status (`active`/`inactive`) |
| `POST /computeruse/start` | ✅ Stub | Placeholder (VNC auto-starts) |
| `POST /computeruse/stop` | ✅ Stub | Placeholder (VNC managed by Windows) |

### Remote Desktop Stack (Implemented)

```
Browser ─────► noVNC (6080) ─────► websockify ─────► TightVNC (5900)
               (Web Client)        (WebSocket)        (VNC Server)
```

### Components to Implement

- **Screenshot Capture**: Full screen and region capture via Windows API
- **Mouse Control**: Move, click, drag, scroll via SendInput
- **Keyboard Control**: Type text, press keys, hotkeys via SendInput
- **Display Info**: Screen resolution, DPI, monitor enumeration
- **Window Enumeration**: List and focus windows via EnumWindows

### Planned API Endpoints

| Endpoint | Status | Description |
|----------|--------|-------------|
| `GET /computeruse/screenshot` | ❌ Planned | Take full screenshot |
| `GET /computeruse/screenshot/region` | ❌ Planned | Region screenshot |
| `GET /computeruse/mouse/position` | ❌ Planned | Get mouse position |
| `POST /computeruse/mouse/move` | ❌ Planned | Move mouse |
| `POST /computeruse/mouse/click` | ❌ Planned | Click |
| `POST /computeruse/mouse/drag` | ❌ Planned | Drag |
| `POST /computeruse/mouse/scroll` | ❌ Planned | Scroll |
| `POST /computeruse/keyboard/type` | ❌ Planned | Type text |
| `POST /computeruse/keyboard/key` | ❌ Planned | Press key |
| `POST /computeruse/keyboard/hotkey` | ❌ Planned | Key combination |
| `GET /computeruse/display/info` | ❌ Planned | Display information |
| `GET /computeruse/display/windows` | ❌ Planned | List windows |

### Windows-Specific Considerations

- Use Windows GDI/GDI+ or DXGI for screenshots
- Windows SendInput API for mouse/keyboard
- EnumWindows/GetWindowText for window enumeration
- Handle DPI scaling for high-DPI displays
- Session 0 isolation if running as a service

---

## 5. LSP Support (Language Server Protocol)

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
2. **Computer Use (Full)** - Programmatic screenshot/input control
3. **LSP Support** - Developer tooling
4. **Interpreter REPL** - Python execution

### Already Implemented

- ✅ **SSH Server** - Built-in SSH with SFTP and port forwarding
- ✅ **Web Terminal** - Browser-based terminal via xterm.js

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

All services are configured to auto-start via Windows Scheduled Tasks (DaytonaDaemon task runs `C:\daemon-win.exe` at boot).

---

## Related Documentation

- Linux Daemon Source: `apps/daemon/`
- Windows ConPTY: https://docs.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
- go-conpty: https://github.com/UserExistsError/conpty
- TightVNC: https://www.tightvnc.com/
- noVNC: https://novnc.com/
- websockify: https://github.com/novnc/websockify
