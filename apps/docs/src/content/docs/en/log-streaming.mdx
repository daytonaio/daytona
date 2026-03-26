---
title: Log Streaming
---

Log streaming allows you to access and process logs as they are being produced, while the process is still running. When executing long-running processes in a sandbox, you often want to access and process their logs in **real-time**. 

Real-time streaming is especially useful for **debugging**, **monitoring**, or integrating with **observability tools**.

- [**Log streaming**](#stream-logs-with-callbacks): stream logs as they are being produced, while the process is still running.
- [**Fetching log snapshot**](#retrieve-all-existing-logs): retrieve all logs up to a certain point.

This guide covers how to use log streaming with callbacks and fetching log snapshots in both asynchronous and synchronous modes.

:::note
Starting with version `0.27.0`, you can retrieve session command logs in two distinct streams: **stdout** and **stderr**.
:::

## Stream logs with callbacks

If your sandboxed process is part of a larger system and is expected to run for an extended period (or indefinitely),
you can process logs asynchronously **in the background**, while the rest of your system continues executing.

This is ideal for:

- Continuous monitoring
- Debugging long-running jobs
- Live log forwarding or visualizations

import { TabItem, Tabs } from '@astrojs/starlight/components'

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
import asyncio
from daytona import Daytona, SessionExecuteRequest

async def main():
  daytona = Daytona()
  sandbox = daytona.create()

  session_id = "streaming-session"
  sandbox.process.create_session(session_id)

  command = sandbox.process.execute_session_command(
    session_id,
    SessionExecuteRequest(
      command='for i in {1..5}; do echo "Step $i"; echo "Error $i" >&2; sleep 1; done',
      var_async=True,
    ),
  )

  # Stream logs with separate callbacks
  logs_task = asyncio.create_task(
    sandbox.process.get_session_command_logs_async(
      session_id,
      command.cmd_id,
      lambda stdout: print(f"[STDOUT]: {stdout}"),
      lambda stderr: print(f"[STDERR]: {stderr}"),
    )
  )

  print("Continuing execution while logs are streaming...")
  await asyncio.sleep(3)
  print("Other operations completed!")

  # Wait for the logs to complete
  await logs_task

  sandbox.delete()
  
if __name__ == "__main__":
    asyncio.run(main())
```

</TabItem>
<TabItem label="Typescript" icon="seti:typescript">

```typescript
import { Daytona, SessionExecuteRequest } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()
  const sandbox = await daytona.create()
  const sessionId = "exec-session-1"
  await sandbox.process.createSession(sessionId)

  const command = await sandbox.process.executeSessionCommand(
    sessionId,
    {
      command: 'for i in {1..5}; do echo "Step $i"; echo "Error $i" >&2; sleep 1; done',
      runAsync: true,
    },
  )

  // Stream logs with separate callbacks
  const logsTask = sandbox.process.getSessionCommandLogs(
    sessionId,
    command.cmdId!,
    (stdout) => console.log('[STDOUT]:', stdout),
    (stderr) => console.log('[STDERR]:', stderr),
  )

  console.log('Continuing execution while logs are streaming...')
  await new Promise((resolve) => setTimeout(resolve, 3000))
  console.log('Other operations completed!')

  // Wait for the logs to complete
  await logsTask

  await sandbox.delete()
}

main()
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create

session_id = 'streaming-session'
sandbox.process.create_session(session_id)

command = sandbox.process.execute_session_command(
  session_id,
  Daytona::SessionExecuteRequest.new(
    command: 'for i in {1..5}; do echo "Step $i"; echo "Error $i" >&2; sleep 1; done',
    var_async: true
  )
)

# Stream logs using a thread
log_thread = Thread.new do
  sandbox.process.get_session_command_logs_stream(
    session_id,
    command.cmd_id,
    on_stdout: ->(stdout) { puts "[STDOUT]: #{stdout}" },
    on_stderr: ->(stderr) { puts "[STDERR]: #{stderr}" }
  )
end

puts 'Continuing execution while logs are streaming...'
sleep(3)
puts 'Other operations completed!'

# Wait for the logs to complete
log_thread.join

daytona.delete(sandbox)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
	client, _ := daytona.NewClient()
	ctx := context.Background()
	sandbox, _ := client.Create(ctx, nil)

	sessionID := "streaming-session"
	sandbox.Process.CreateSession(ctx, sessionID)

	// Execute async command that outputs to stdout and stderr
	cmd := `for i in 1 2 3 4 5; do echo "Step $i"; echo "Error $i" >&2; sleep 1; done`
	cmdResult, _ := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, cmd, true)
	cmdID, _ := cmdResult["id"].(string)

	// Create channels for stdout and stderr
	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	// Stream logs in a goroutine
	go func() {
		err := sandbox.Process.GetSessionCommandLogsStream(ctx, sessionID, cmdID, stdout, stderr)
		if err != nil {
			log.Printf("Stream error: %v", err)
		}
	}()

	fmt.Println("Continuing execution while logs are streaming...")

	// Read from channels until both are closed
	stdoutOpen, stderrOpen := true, true
	for stdoutOpen || stderrOpen {
		select {
		case chunk, ok := <-stdout:
			if !ok {
				stdoutOpen = false
			} else {
				fmt.Fprintf(os.Stdout, "[STDOUT]: %s", chunk)
			}
		case chunk, ok := <-stderr:
			if !ok {
				stderrOpen = false
			} else {
				fmt.Fprintf(os.Stderr, "[STDERR]: %s", chunk)
			}
		}
	}

	fmt.Println("Streaming completed!")
	sandbox.Delete(ctx)
}
```
</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}/command/{commandId}/logs'
```
</TabItem>

</Tabs>

For more information, see the [Python SDK](/docs/python-sdk/sync/process/), [TypeScript SDK](/docs/typescript-sdk/process/), [Ruby SDK](/docs/ruby-sdk/process/), [Go SDK](/docs/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/process) references.

> [**get_session_command_logs_async (Python SDK)**](/docs/python-sdk/sync/process/#processget_session_command_logs_async)
>
> [**getSessionCommandLogs (TypeScript SDK)**](/docs/typescript-sdk/process/#getsessioncommandlogs)
>
> [**get_session_command_logs_async (Ruby SDK)**](/docs/ruby-sdk/process/#get_session_command_logs_async)
>
> [**GetSessionCommandLogsStream (Go SDK)**](/docs/go-sdk/daytona/#ProcessService.GetSessionCommandLogsStream)
>
> [**get session command logs (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/session/{sessionId}/exec)

## Retrieve all existing logs

If the command has a predictable duration, or if you don't need to run it in the background but want to
periodically check all existing logs, you can use the following example to get the logs up to the current point in time.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
import time
from daytona import Daytona, SessionExecuteRequest

daytona = Daytona()
sandbox = daytona.create()
session_id = "exec-session-1"
sandbox.process.create_session(session_id)

# Execute a blocking command and wait for the result
command = sandbox.process.execute_session_command(
  session_id, SessionExecuteRequest(command="echo 'Hello from stdout' && echo 'Hello from stderr' >&2")
)
print(f"[STDOUT]: {command.stdout}")
print(f"[STDERR]: {command.stderr}")
print(f"[OUTPUT]: {command.output}")

# Or execute command in the background and get the logs later
command = sandbox.process.execute_session_command(
  session_id, 
  SessionExecuteRequest(
    command='while true; do if (( RANDOM % 2 )); then echo "All good at $(date)"; else echo "Oops, an error at $(date)" >&2; fi; sleep 1; done',
    run_async=True
  )
)
time.sleep(5)
# Get the logs up to the current point in time
logs = sandbox.process.get_session_command_logs(session_id, command.cmd_id)
print(f"[STDOUT]: {logs.stdout}")
print(f"[STDERR]: {logs.stderr}")
print(f"[OUTPUT]: {logs.output}")

sandbox.delete()
```

</TabItem>
<TabItem label="Typescript" icon="seti:typescript">

```typescript
import { Daytona, SessionExecuteRequest } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()
  const sandbox = await daytona.create()
  const sessionId = "exec-session-1"
  await sandbox.process.createSession(sessionId)

  // Execute a blocking command and wait for the result
  const command = await sandbox.process.executeSessionCommand(
    sessionId,
    {
      command: 'echo "Hello from stdout" && echo "Hello from stderr" >&2',
    },
  )
  console.log(`[STDOUT]: ${command.stdout}`)
  console.log(`[STDERR]: ${command.stderr}`)
  console.log(`[OUTPUT]: ${command.output}`)

  // Or execute command in the background and get the logs later
  const command2 = await sandbox.process.executeSessionCommand(
    sessionId,
    {
      command: 'while true; do if (( RANDOM % 2 )); then echo "All good at $(date)"; else echo "Oops, an error at $(date)" >&2; fi; sleep 1; done',
      runAsync: true,
    },
  )
  await new Promise((resolve) => setTimeout(resolve, 5000))
  // Get the logs up to the current point in time
  const logs = await sandbox.process.getSessionCommandLogs(sessionId, command2.cmdId!)
  console.log(`[STDOUT]: ${logs.stdout}`)
  console.log(`[STDERR]: ${logs.stderr}`)
  console.log(`[OUTPUT]: ${logs.output}`)

  await sandbox.delete()
}

main()
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create
session_id = 'exec-session-1'
sandbox.process.create_session(session_id)

# Execute a blocking command and wait for the result
command = sandbox.process.execute_session_command(
  session_id,
  Daytona::SessionExecuteRequest.new(
    command: 'echo "Hello from stdout" && echo "Hello from stderr" >&2'
  )
)
puts "[STDOUT]: #{command.stdout}"
puts "[STDERR]: #{command.stderr}"
puts "[OUTPUT]: #{command.output}"

# Or execute command in the background and get the logs later
command = sandbox.process.execute_session_command(
  session_id,
  Daytona::SessionExecuteRequest.new(
    command: 'while true; do if (( RANDOM % 2 )); then echo "All good at $(date)"; else echo "Oops, an error at $(date)" >&2; fi; sleep 1; done',
    var_async: true
  )
)
sleep(5)
# Get the logs up to the current point in time
logs = sandbox.process.get_session_command_logs(session_id, command.cmd_id)
puts "[STDOUT]: #{logs.stdout}"
puts "[STDERR]: #{logs.stderr}"
puts "[OUTPUT]: #{logs.output}"

daytona.delete(sandbox)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
	client, _ := daytona.NewClient()
	ctx := context.Background()
	sandbox, _ := client.Create(ctx, nil)

	sessionID := "exec-session-1"
	sandbox.Process.CreateSession(ctx, sessionID)

	// Execute a blocking command and wait for the result
	cmd1, _ := sandbox.Process.ExecuteSessionCommand(ctx, sessionID,
		`echo "Hello from stdout" && echo "Hello from stderr" >&2`, false)
	if stdout, ok := cmd1["stdout"].(string); ok {
		fmt.Printf("[STDOUT]: %s\n", stdout)
	}
	if stderr, ok := cmd1["stderr"].(string); ok {
		fmt.Printf("[STDERR]: %s\n", stderr)
	}

	// Or execute command in the background and get the logs later
	cmd := `counter=1; while (( counter <= 5 )); do echo "Count: $counter"; ((counter++)); sleep 1; done`
	cmdResult, _ := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, cmd, true)
	cmdID, _ := cmdResult["id"].(string)

	time.Sleep(5 * time.Second)

	// Get the logs up to the current point in time
	logs, err := sandbox.Process.GetSessionCommandLogs(ctx, sessionID, cmdID)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}
	if logContent, ok := logs["logs"].(string); ok {
		fmt.Printf("[LOGS]: %s\n", logContent)
	}

	sandbox.Delete(ctx)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}/command/{commandId}/logs'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/python-sdk/sync/process/), [TypeScript SDK](/docs/typescript-sdk/process/), [Ruby SDK](/docs/ruby-sdk/process/), [Go SDK](/docs/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/process) references.

> [**get_session_command_logs (Python SDK)**](/docs/python-sdk/sync/process/#processget_session_command_logs)
>
> [**getSessionCommandLogs (TypeScript SDK)**](/docs/typescript-sdk/process/#getsessioncommandlogs)
>
> [**get_session_command_logs (Ruby SDK)**](/docs/ruby-sdk/process/#get_session_command_logs)
>
> [**GetSessionCommandLogs (Go SDK)**](/docs/go-sdk/daytona/#ProcessService.GetSessionCommandLogs)
>
> [**get session command logs (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/session/{sessionId}/exec)
