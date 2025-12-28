# Daytona Windows Daemon - Future Features

This document lists components and features that are planned for future implementation in the Windows daemon. These features exist in the Linux daemon (`apps/daemon`) and should be ported with Windows-specific adaptations.

---

## 1. SSH Server

**Reference**: `apps/daemon/pkg/ssh/`

Provides remote shell access and file transfer capabilities.

### Components to Implement

- **Authentication**: Public key and password authentication
- **SFTP Handler**: Secure file transfer protocol
- **Port Forwarding**: TCP port forwarding (local and remote)
- **Named Pipe Forwarding**: Windows equivalent of Unix socket forwarding (`streamlocal-forward`)
- **Interactive Shell**: Integration with ConPTY for interactive sessions

### Windows-Specific Considerations

- Replace Unix socket forwarding with Windows Named Pipes
- Use ConPTY instead of Unix PTY for interactive shells
- Windows signal handling for session termination
- Replace `golang.org/x/sys/unix` signals with Windows equivalents

### Key Dependencies

- `github.com/gliderlabs/ssh` - SSH server library
- `github.com/pkg/sftp` - SFTP implementation
- `golang.org/x/crypto/ssh` - SSH protocol

---

## 2. Terminal Server

**Reference**: `apps/daemon/pkg/terminal/`

Web-based terminal accessible via browser using xterm.js.

### Components to Implement

- **HTTP Server**: Serve xterm.js frontend (static files)
- **WebSocket Handler**: Bidirectional communication with terminal
- **ConPTY Integration**: Windows pseudo-terminal backend
- **Window Resize**: Handle terminal resize events
- **UTF-8 Decoder**: Handle multi-byte character sequences

### Windows-Specific Considerations

- Use Windows ConPTY API instead of Unix PTY
- Spawn `powershell.exe` or `cmd.exe` as the shell
- Handle Windows-specific escape sequences

### Key Dependencies

- `github.com/gorilla/websocket` - WebSocket support
- `github.com/UserExistsError/conpty` or similar - ConPTY wrapper

### Static Assets

- `xterm.js` - Terminal emulator
- `xterm.css` - Terminal styles
- `xterm-addon-fit.js` - Auto-fit addon

---

## 3. PTY Sessions (Toolbox API)

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

## 4. Interpreter REPL

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

## 5. Computer Use (Plugin System)

**Reference**: `apps/daemon/pkg/toolbox/computeruse/`

Screen capture and input automation via a plugin architecture.

### Components to Implement

- **Plugin Manager**: Load/unload computer use plugin
- **Screenshot Capture**: Full screen and region capture
- **Mouse Control**: Move, click, drag, scroll
- **Keyboard Control**: Type text, press keys, hotkeys
- **Display Info**: Screen resolution, window enumeration

### API Endpoints

```
GET    /computeruse/status              - Plugin status
POST   /computeruse/start               - Start plugin
POST   /computeruse/stop                - Stop plugin
GET    /computeruse/screenshot          - Take screenshot
GET    /computeruse/screenshot/region   - Region screenshot
GET    /computeruse/mouse/position      - Get mouse position
POST   /computeruse/mouse/move          - Move mouse
POST   /computeruse/mouse/click         - Click
POST   /computeruse/keyboard/type       - Type text
POST   /computeruse/keyboard/key        - Press key
GET    /computeruse/display/info        - Display information
GET    /computeruse/display/windows     - List windows
```

### Windows-Specific Considerations

- Use Windows GDI/GDI+ for screenshots
- Windows Input API for mouse/keyboard
- EnumWindows for window enumeration
- Plugin compiled as Windows DLL or separate executable

---

## 6. LSP Support (Language Server Protocol)

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
2. **Terminal Server** - Web-based terminal access
3. **SSH Server** - Remote access capability
4. **LSP Support** - Developer tooling
5. **Interpreter REPL** - Python execution
6. **Computer Use** - Automation features

---

## Related Documentation

- Linux Daemon Source: `apps/daemon/`
- Windows ConPTY: https://docs.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
- go-conpty: https://github.com/UserExistsError/conpty
