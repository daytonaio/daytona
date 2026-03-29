# TTY CLI Implementation

This document describes the implementation of TTY support for the Daytona CLI's `exec` command.

## Overview

The TTY functionality enables interactive terminal sessions with remote sandboxes, providing a true terminal experience with proper signal handling, terminal resizing, and raw mode support.

## Implementation Details

### 1. Command Line Interface

**New Flag**: `--tty` / `-t`

- **Description**: Allocate a pseudo-TTY for interactive command execution
- **Incompatible with**: `--timeout` flag (TTY sessions are inherently interactive)
- **Requires**: Interactive terminal (stdin must be a terminal)

### 2. Architecture

```
CLI Client                    Daemon
┌─────────────────┐          ┌──────────────────────┐
│                 │          │                      │
│  exec --tty     │ WebSocket│  TTY Exec Handler    │
│  ┌─────────────┐│ ◄──────► │  ┌──────────────────┐│
│  │Raw Terminal ││          │  │ PTY Process      ││
│  │Signal Handle││          │  │ Signal Forwarding││
│  │Resize Handle││          │  │ Terminal Resize  ││
│  └─────────────┘│          │  └──────────────────┘│
└─────────────────┘          └──────────────────────┘
```

### 3. Core Components

#### a) TTY Request Structure (`TTYExecuteRequest`)

```go
type TTYExecuteRequest struct {
    Command string  `json:"command"`
    Cwd     *string `json:"cwd,omitempty"`
    Cols    int     `json:"cols"`
    Rows    int     `json:"rows"`
}
```

#### b) WebSocket Client (`toolbox.go`)

- **Connection**: WebSocket to `/{sandboxId}/process/exec/tty`
- **Protocol**: Binary messages for I/O, JSON for control messages
- **Features**:
  - Connection establishment with timeout
  - Raw terminal mode handling
  - Bidirectional I/O forwarding
  - Signal handling and forwarding
  - Terminal resize support

#### c) Platform-Specific Signal Handling

- **Unix** (`tty_unix.go`): Supports SIGWINCH, SIGINT, SIGTERM
- **Windows** (`tty_windows.go`): Supports SIGINT, SIGTERM (no SIGWINCH)

### 4. Key Features

#### Terminal Raw Mode

- Automatically detects terminal capability
- Sets terminal to raw mode for true TTY experience
- Restores terminal state on exit (with error handling)

#### Signal Handling

- **SIGINT**: Forwards Ctrl+C (0x03) to remote process
- **SIGTERM**: Graceful termination
- **SIGWINCH** (Unix only): Terminal resize events

#### Resize Handling

- Detects terminal size changes
- Sends resize messages to daemon
- Format: `{"type": "resize", "cols": X, "rows": Y}`

#### Connection Management

- 30-second WebSocket handshake timeout
- Connection establishment confirmation
- Graceful cleanup on termination
- Exit code propagation from remote process

### 5. Error Handling

#### Validation

- TTY mode requires interactive terminal
- TTY and timeout flags are mutually exclusive
- WebSocket connection validation

#### Resilience

- Connection timeout handling
- Graceful goroutine cleanup
- Terminal state restoration
- Error propagation with context

### 6. Usage Examples

```bash
# Interactive shell
daytona sandbox exec my-sandbox -t -- bash

# Interactive Python REPL
daytona sandbox exec my-sandbox --tty -- python3

# Interactive editor
daytona sandbox exec my-sandbox -t -- vim file.txt

# With custom working directory
daytona sandbox exec my-sandbox --cwd /app -t -- bash
```

### 7. Compatibility

#### Supported Platforms

- ✅ Linux (full support including SIGWINCH)
- ✅ macOS (full support including SIGWINCH)
- ✅ Windows (limited - no SIGWINCH support)

#### Terminal Requirements

- Interactive terminal (stdin must be a TTY)
- Terminal emulator supporting raw mode
- For best experience: color support and proper terminal size detection

### 8. Integration Points

#### Daemon Side

- Connects to existing TTY exec WebSocket endpoint
- Uses existing authentication headers
- Compatible with existing toolbox proxy architecture

#### Client Side  

- Extends existing `toolbox.Client`
- Reuses existing configuration and authentication
- Follows existing CLI patterns and error handling

### 9. Dependencies

**New Dependencies Added:**

- `github.com/gorilla/websocket v1.5.3` - WebSocket client

**Existing Dependencies Used:**

- `golang.org/x/term` - Terminal handling
- Standard library: `os/signal`, `syscall`, `context`

### 10. Testing

#### Basic Functionality

- Terminal detection
- Command line validation
- Signal handling setup
- Terminal size detection

#### Integration Testing

- Requires running daemon with TTY exec support
- Test with various shells (bash, zsh, sh)
- Test with interactive applications (vim, nano, top)
- Test resize functionality
- Test signal forwarding (Ctrl+C, etc.)

## Future Enhancements

1. **Clipboard Integration**: Support for clipboard operations in TTY mode
2. **Session Persistence**: Ability to reconnect to existing TTY sessions
3. **Logging**: Optional session recording/playback
4. **Performance**: Optimize for high-throughput terminal applications

## Security Considerations

- Authentication uses existing token/API key system
- WebSocket connections are secured with same headers as HTTP
- No additional authentication requirements
- Raw terminal mode is local-only (doesn't affect security)

## Debugging

To debug TTY issues:

1. **Connection Issues**: Check WebSocket endpoint availability
2. **Terminal Issues**: Verify `term.IsTerminal()` returns true
3. **Signal Issues**: Check platform-specific signal support
4. **Size Issues**: Verify terminal size detection with `term.GetSize()`

Example debug command:

```bash
# Check if terminal is interactive
daytona sandbox exec my-sandbox -- test -t 0 && echo "TTY available" || echo "No TTY"
```
