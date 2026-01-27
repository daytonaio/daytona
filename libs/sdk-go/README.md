# Daytona Go SDK

The official Go SDK for Daytona, enabling programmatic interaction with Daytona Sandboxes

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    // Create a new Daytona client (uses DAYTONA_API_KEY from environment)
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a sandbox
    params := &types.ImageParams{
        SandboxBaseParams: types.SandboxBaseParams{
            Language: types.CodeLanguagePython,
            EnvVars: map[string]string{
                "NODE_ENV": "development",
            },
        },
    }

    buildLogs := make(chan string, 100)
    go func() {
  for logLine := range logChan {
   fmt.Printf("[BUILD] %s\n", logLine)
  }
 }()
  
    // Default WaitForStart is true, but can be overriden for more async behavior
    sandbox, buildLogs, err := client.Create(ctx, params,
  daytona.WithTimeout(90*time.Second),
 )
 if err != nil {
  log.Fatal(err)
 }

    log.Printf("✓ Created sandbox: %s (ID: %s)\n", sandbox.Name, sandbox.ID)

    // Execute a command
    result, err := sandbox.Process.ExecuteCommand(ctx, "echo 'Hello, World!'")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Output: %s\n", result.Result)

    // Clean up
    if err := sandbox.Delete(ctx); err != nil {
        log.Printf("Failed to delete: %v", err)
    }
}
```

## Configuration

The SDK can be configured using environment variables or a configuration object.

### Environment Variables

Set the following environment variables:

```bash
export DAYTONA_API_KEY=your-api-key
```

Then create the client:

```go
client, err := daytona.NewClient()
```

### Configuration Object

```go
config := &types.DaytonaConfig{
    APIKey: "your-api-key",
}

client, err := daytona.NewClientWithConfig(config)
```

## Usage Examples

All of the usage examples are maintained in the `/examples` folder. Please check it out for latest patterns of SDK usage.

## Examples

The `examples/` directory contains comprehensive examples demonstrating various SDK features:

- **sandbox** - Basic sandbox creation and lifecycle management
- **filesystem** - File system operations (upload, download, list)
- **git_operations** - Git operations (clone, status, branches)
- **fromimage** - Creating sandboxes from custom images
- **code_interpreter** - Python code execution with WebSocket streaming
- **lsp_usage** - Language Server Protocol integration
- **pty_interactive** - Interactive PTY sessions
- **snapshots/simple** - Basic snapshot operations
- **snapshots/withlogstreaming** - Snapshot creation with real-time build log streaming
- **volumes** - Volume management

To run an example:

```bash
export DAYTONA_API_KEY=your-api-key
go run examples/sandbox/main.go
go run examples/code_interpreter/main.go
go run examples/fromimage/main.go
go run examples/snapshots/withlogstreaming/snapshot_with_logs.go
```

## Best Practices

### Resource Cleanup

Always clean up sandboxes when done:

```go
sandbox, err := client.Create(ctx, params)
if err != nil {
    log.Fatal(err)
}
defer func() {
    if err := sandbox.Delete(ctx); err != nil {
        log.Printf("Failed to delete sandbox: %v", err)
    }
}()

// Use the sandbox...
```

### Error Handling

Always check errors and handle them appropriately:

```go
result, err := sandbox.Process.ExecuteCommand(ctx, "some-command")
if err != nil {
    log.Printf("Command failed: %v", err)
    return
}

if result.ExitCode != 0 {
    log.Printf("Command exited with code %d: %s", result.ExitCode, result.Result)
    return
}

log.Printf("Success: %s\n", result.Result)
```

### Context Usage

Use contexts appropriately for timeouts and cancellation:

```go
// Create a context with timeout for long-running operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// Use the context for all operations
sandbox, err := client.Create(ctx, params)
if err != nil {
    log.Fatal(err)
}

result, err := sandbox.Process.ExecuteCommand(ctx, "long-running-command")
```

## API Reference

### Client Methods and Properties

**Properties:**

- `Volume *VolumeService` - Access volume management operations
- `Snapshot *SnapshotService` - Access snapshot management operations

**Methods:**

- `NewClient() (*Client, error)` - Create a new Daytona client with default configuration
- `NewClientWithConfig(config *types.DaytonaConfig) (*Client, error)` - Create a new Daytona client with custom configuration
- `Create(ctx, params, options...) (*Sandbox, <-chan string, error)` - Create a sandbox and returns a channel for streaming build logs
  - Options: `WithTimeout(time.Duration)`
- `Get(ctx, sandboxIDOrName) (*Sandbox, error)` - Get a sandbox by ID or name
- `FindOne(ctx, idOrName, labels) (*Sandbox, error)` - Find a sandbox by ID/name or labels
- `List(ctx, labels, page, limit) (*PaginatedSandboxes, error)` - List sandboxes with pagination

### Sandbox Properties and Methods

**Properties:**

- `FileSystem *FileSystemService` - Access file system operations
- `Git *GitService` - Access Git operations
- `Process *ProcessService` - Access process execution
- `CodeInterpreter *CodeInterpreterService` - Access code interpreter
- `ComputerUse *ComputerUseService` - Access desktop automation
- `Name string` - Sandbox name
- `State apiclient.SandboxState` - Current state
- `ID string` - Sandbox ID
- `ToolboxClient *toolbox.APIClient` - Toolbox API client

**Methods:**

- `RefreshData(ctx) error` - Refresh sandbox data from API
- `GetUserHomeDir(ctx) (string, error)` - Get user home directory path
- `GetWorkingDir(ctx) (string, error)` - Get working directory path
- `SetLabels(ctx, labels) error` - Set custom labels
- `GetPreviewLink(ctx, port) (string, error)` - Get port preview URL
- `Start(ctx) error` - Start this sandbox (60s default timeout)
- `StartWithTimeout(ctx, timeout time.Duration) error` - Start with custom timeout
- `Stop(ctx) error` - Stop this sandbox (60s default timeout)
- `StopWithTimeout(ctx, timeout time.Duration) error` - Stop with custom timeout
- `Delete(ctx) error` - Delete this sandbox (60s default timeout)
- `DeleteWithTimeout(ctx, timeout time.Duration) error` - Delete with custom timeout
- `Archive(ctx) error` - Archive the sandbox to object storage
- `WaitForStart(ctx, timeoutSec int) error` - Wait for sandbox to start
- `WaitForStop(ctx, timeoutSec int) error` - Wait for sandbox to stop

### Snapshot Service Methods

- `List(ctx, page, limit) (*PaginatedSnapshots, error)` - List snapshots with pagination
- `Get(ctx, nameOrID) (*Snapshot, error)` - Get a snapshot by name or ID
- `Create(ctx, params, timeout) (*Snapshot, <-chan string, error)` - Create a snapshot from an image with real-time build log streaming
  - Returns the snapshot object, a channel for streaming build logs, and an error
  - The log channel will be closed when the build is complete
- `Delete(ctx, snapshot) error` - Delete a snapshot

### Volume Service Methods

- `List(ctx) ([]*Volume, error)` - List all volumes
- `Get(ctx, name) (*Volume, error)` - Get a volume by name
- `Create(ctx, name) (*Volume, error)` - Create a new volume
- `Delete(ctx, volume) error` - Delete a volume

## License

Apache-2.0

For issues and questions:

- **GitHub Issues**: https://github.com/daytonaio/daytona/issues
- **Documentation**: https://www.daytona.io/docs
