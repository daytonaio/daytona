# TTY Execution Feature

This document describes the TTY (terminal) execution feature added to the Daytona daemon backend.

## Overview

The TTY execution feature allows commands to be executed in an interactive pseudo-terminal (PTY) environment via WebSocket connections. This enables real-time interactive command execution with features like:

- Full terminal emulation with ANSI escape sequences
- Interactive command input/output
- Real-time streaming of command output
- Support for interactive programs (shells, editors, etc.)
- Multiple WebSocket clients can attach to the same TTY session

## API Changes

### ExecuteRequest Type

The `ExecuteRequest` type has been extended with a new `tty` field:

```go
type ExecuteRequest struct {
    Command string  `json:"command" validate:"required"`
    Timeout *uint32 `json:"timeout,omitempty" validate:"optional"` 
    Cwd     *string `json:"cwd,omitempty" validate:"optional"`
    Tty     bool    `json:"tty,omitempty" validate:"optional"`  // New field
}
```

### Response Types

When `tty: false` (or omitted), the response is the existing `ExecuteResponse`:

```go
type ExecuteResponse struct {
    ExitCode int    `json:"exitCode"`
    Result   string `json:"result"`
}
```

When `tty: true`, the response is a new `ExecuteTTYResponse`:

```go
type ExecuteTTYResponse struct {
    SessionID string `json:"sessionId"`
}
```

## API Endpoints

### Execute Command

**POST** `/process/execute`

Execute a command. When `tty: true` is specified, returns a session ID for WebSocket connection.

**Request Body:**

```json
{
    "command": "ls -la",
    "tty": true,
    "cwd": "/home/user",
    "timeout": 30
}
```

**Response (TTY mode):**

```json
{
    "sessionId": "uuid-string"
}
```

### Connect to TTY Session

**GET** `/process/exec/{sessionId}/connect`

WebSocket endpoint to connect to a TTY execution session. Upgrades the HTTP connection to WebSocket.

**WebSocket Protocol:**

- **Control Messages (Text):** JSON messages with session control information
- **Data Messages (Binary):** Raw terminal data (input/output)

**Control Message Example:**

```json
{
    "type": "control",
    "status": "connected"
}
```

**Close Message:** Contains exit information

```json
{
    "exitCode": 0,
    "exitReason": "success"
}
```

## Usage Examples

### Simple Command Execution

```bash
# Standard execution (existing behavior)
curl -X POST http://localhost:8080/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command": "echo hello", "tty": false}'

# TTY execution (new feature)
curl -X POST http://localhost:8080/process/execute \
  -H "Content-Type: application/json" \
  -d '{"command": "bash", "tty": true}'
```

### WebSocket Connection

After receiving a session ID from TTY execution, connect via WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8080/process/exec/{sessionId}/connect');

ws.onmessage = (event) => {
    if (typeof event.data === 'string') {
        // Control message
        const control = JSON.parse(event.data);
        console.log('Control:', control);
    } else {
        // Binary terminal data
        const reader = new FileReader();
        reader.onload = () => {
            const text = reader.result;
            console.log('Output:', text);
        };
        reader.readAsText(event.data);
    }
};

// Send input to terminal
ws.send('ls -la\r');
```

## Implementation Details

### Architecture

The TTY execution feature is implemented as:

1. **TTY Exec Sessions:** Managed pseudo-terminal sessions for command execution
2. **Session Manager:** Global manager for TTY execution sessions
3. **WebSocket Handler:** Handles WebSocket connections to TTY sessions
4. **PTY Integration:** Uses `github.com/creack/pty` for pseudo-terminal functionality

### Key Components

- `tty_exec.go`: Core TTY execution session implementation
- `tty_exec_test.go`: Comprehensive test suite
- `types.go`: Extended request/response types
- `execute.go`: Updated execution handler
- `server.go`: WebSocket endpoint registration

### Session Lifecycle

1. **Creation:** POST to `/process/execute` with `tty: true` creates a session
2. **Connection:** WebSocket connects to `/process/exec/{sessionId}/connect`
3. **Execution:** Command is executed in PTY when first client connects
4. **I/O:** Bidirectional data flow between WebSocket clients and PTY
5. **Termination:** Session ends when command exits, notifying all clients
6. **Cleanup:** Session is removed from manager after exit notification

### Features

- **Multi-client Support:** Multiple WebSocket clients can attach to the same session
- **Timeout Support:** Sessions respect the timeout parameter from the request
- **Working Directory:** Commands executed in specified working directory
- **Environment Variables:** TERM=xterm-256color set for proper terminal emulation
- **Error Handling:** Comprehensive error handling and client notification
- **Resource Cleanup:** Proper cleanup of PTY resources and WebSocket connections

## Testing

The implementation includes comprehensive tests covering:

- TTY vs non-TTY execution paths
- Session creation and management
- Error handling and edge cases
- WebSocket handler functionality
- Timeout and working directory support

Run tests with:

```bash
go test ./apps/daemon/pkg/toolbox/process -v
```

## Backward Compatibility

This feature maintains full backward compatibility:

- Existing clients continue to work unchanged
- `tty` field defaults to `false` when omitted
- Non-TTY execution behavior is unchanged
- All existing API endpoints remain functional

## Security Considerations

- TTY sessions are isolated per session ID (UUID)
- WebSocket connections require valid session IDs
- Sessions automatically clean up on timeout or exit
- Process execution respects working directory restrictions
- Resource limits (timeout) prevent runaway processes
