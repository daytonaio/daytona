---
title: Pseudo Terminal (PTY)
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

Daytona provides powerful pseudo terminal (PTY) capabilities through the `process` module in sandboxes. PTY sessions allow you to create interactive terminal sessions that can execute commands, handle user input, and manage terminal operations.

A PTY (Pseudo Terminal) is a virtual terminal interface that allows programs to interact with a shell as if they were connected to a real terminal. PTY sessions in Daytona enable:

- **Interactive Development**: REPLs, debuggers, and development tools
- **Build Processes**: Running and monitoring compilation, testing, or deployment
- **System Administration**: Remote server management and configuration
- **User Interfaces**: Terminal-based applications requiring user interaction

## Create PTY session

Daytona provides methods to create an interactive terminal session that can execute commands and handle user input.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona.common.pty import PtySize

pty_handle = sandbox.process.create_pty_session(
    id="my-session",
    cwd="/workspace",
    envs={"TERM": "xterm-256color"},
    pty_size=PtySize(cols=120, rows=30)
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Create a PTY session with custom configuration
const ptyHandle = await sandbox.process.createPty({
  id: 'my-interactive-session',
  cwd: '/workspace',
  envs: { TERM: 'xterm-256color', LANG: 'en_US.UTF-8' },
  cols: 120,
  rows: 30,
  onData: (data) => {
    // Handle terminal output
    const text = new TextDecoder().decode(data)
    process.stdout.write(text)
  },
})

// Wait for connection to be established
await ptyHandle.waitForConnection()

// Send commands to the terminal
await ptyHandle.sendInput('ls -la\n')
await ptyHandle.sendInput('echo "Hello, PTY!"\n')
await ptyHandle.sendInput('exit\n')

// Wait for completion and get result
const result = await ptyHandle.wait()
console.log(`PTY session completed with exit code: ${result.exitCode}`)

// Clean up
await ptyHandle.disconnect()
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
pty_size = Daytona::PtySize.new(rows: 30, cols: 120)
pty_handle = sandbox.process.create_pty_session(
  id: 'my-interactive-session',
  cwd: '/workspace',
  envs: { 'TERM' => 'xterm-256color' },
  pty_size: pty_size
)

# Use the PTY session
pty_handle.send_input("ls -la\n")
pty_handle.send_input("echo 'Hello, PTY!'\n")
pty_handle.send_input("exit\n")

# Handle output
pty_handle.each { |data| print data }

puts "PTY session completed with exit code: #{pty_handle.exit_code}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Create a PTY session with custom configuration
handle, err := sandbox.Process.CreatePty(ctx, "my-interactive-session",
	options.WithCreatePtySize(types.PtySize{Cols: 120, Rows: 30}),
	options.WithCreatePtyEnv(map[string]string{"TERM": "xterm-256color"}),
)
if err != nil {
	log.Fatal(err)
}
defer handle.Disconnect()

// Wait for connection to be established
if err := handle.WaitForConnection(ctx); err != nil {
	log.Fatal(err)
}

// Send commands to the terminal
handle.SendInput([]byte("ls -la\n"))
handle.SendInput([]byte("echo 'Hello, PTY!'\n"))
handle.SendInput([]byte("exit\n"))

// Read output from channel
for data := range handle.DataChan() {
	fmt.Print(string(data))
}

// Wait for completion and get result
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("PTY session completed with exit code: %d\n", *result.ExitCode)
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/pty' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "cols": 1,
  "cwd": "",
  "envs": {
    "additionalProperty": ""
  },
  "id": "",
  "lazyStart": true,
  "rows": 1
}'
```

</TabItem>

</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/sync/process/), [TypeScript SDK](/docs/en/typescript-sdk/process/#createpty), [Ruby SDK](/docs/en/ruby-sdk/process/), [Go SDK](/docs/en/go-sdk/daytona/#type-processservice), and [API](/docs/en/tools/api#daytona-toolbox/tag/process) references:

> [**create_pty_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processcreate_pty_session)
>
> [**createPty (TypeScript SDK)**](/docs/en/typescript-sdk/process/#createpty)
>
> [**create_pty_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#create_pty_session)
>
> [**CreatePty (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.CreatePty)
>
> [**Create PTY session (API)**](/docs/en/tools/api#daytona-toolbox/tag/process/POST/process/pty)

## Connect to PTY session

Daytona provides methods to establish a connection to an existing PTY session, enabling interaction with a previously created terminal.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
pty_handle = sandbox.process.connect_pty_session("my-session")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Connect to an existing PTY session
const handle = await sandbox.process.connectPty('my-session', {
  onData: (data) => {
    // Handle terminal output
    const text = new TextDecoder().decode(data)
    process.stdout.write(text)
  },
})

// Wait for connection to be established
await handle.waitForConnection()

// Send commands to the existing session
await handle.sendInput('pwd\n')
await handle.sendInput('ls -la\n')
await handle.sendInput('exit\n')

// Wait for completion
const result = await handle.wait()
console.log(`Session exited with code: ${result.exitCode}`)

// Clean up
await handle.disconnect()
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Connect to an existing PTY session
pty_handle = sandbox.process.connect_pty_session('my-session')
pty_handle.send_input("echo 'Hello World'\n")
pty_handle.send_input("exit\n")

# Handle output
pty_handle.each { |data| print data }

puts "Session exited with code: #{pty_handle.exit_code}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Connect to an existing PTY session
handle, err := sandbox.Process.ConnectPty(ctx, "my-session")
if err != nil {
	log.Fatal(err)
}
defer handle.Disconnect()

// Wait for connection to be established
if err := handle.WaitForConnection(ctx); err != nil {
	log.Fatal(err)
}

// Send commands to the existing session
handle.SendInput([]byte("pwd\n"))
handle.SendInput([]byte("ls -la\n"))
handle.SendInput([]byte("exit\n"))

// Read output
for data := range handle.DataChan() {
	fmt.Print(string(data))
}

// Wait for completion
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Session exited with code: %d\n", *result.ExitCode)
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/pty/{sessionId}/connect'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/sync/process/), [TypeScript SDK](/docs/en/typescript-sdk/process/), [Ruby SDK](/docs/en/ruby-sdk/process/), [Go SDK](/docs/en/go-sdk/daytona/#type-processservice), and [API](/docs/en/tools/api#daytona-toolbox/tag/process) references:

> [**connect_pty_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processconnect_pty_session)
>
> [**connectPty (TypeScript SDK)**](/docs/en/typescript-sdk/process/#connectpty)
>
> [**connect_pty_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#connect_pty_session)
>
> [**ConnectPty (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ConnectPty)
>
> [**Connect to PTY session (API)**](/docs/en/tools/api#daytona-toolbox/tag/process/GET/process/pty/{sessionId}/connect)

## List PTY sessions

Daytona provides methods to list PTY sessions, allowing you to retrieve information about all PTY sessions, both active and inactive, that have been created in the sandbox.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# List all PTY sessions
sessions = sandbox.process.list_pty_sessions()

for session in sessions:
    print(f"Session ID: {session.id}")
    print(f"Active: {session.active}")
    print(f"Created: {session.created_at}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// List all PTY sessions
const sessions = await sandbox.process.listPtySessions()

for (const session of sessions) {
  console.log(`Session ID: ${session.id}`)
  console.log(`Active: ${session.active}`)
  console.log(`Created: ${session.createdAt}`)
  console.log('---')
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# List all PTY sessions
sessions = sandbox.process.list_pty_sessions

sessions.each do |session|
  puts "Session ID: #{session.id}"
  puts "Active: #{session.active}"
  puts "Terminal Size: #{session.cols}x#{session.rows}"
  puts '---'
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// List all PTY sessions
sessions, err := sandbox.Process.ListPtySessions(ctx)
if err != nil {
	log.Fatal(err)
}

for _, session := range sessions {
	fmt.Printf("Session ID: %s\n", session.Id)
	fmt.Printf("Active: %t\n", session.Active)
	fmt.Printf("Terminal Size: %dx%d\n", session.Cols, session.Rows)
	fmt.Println("---")
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/pty'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/sync/process/), [TypeScript SDK](/docs/en/typescript-sdk/process/), [Ruby SDK](/docs/en/ruby-sdk/process/), [Go SDK](/docs/en/go-sdk/daytona/#type-processservice), and [API](/docs/en/tools/api#daytona-toolbox/tag/process) references:

> [**list_pty_sessions (Python SDK)**](/docs/en/python-sdk/sync/process/#processlist_pty_sessions)
>
> [**listPtySessions (TypeScript SDK)**](/docs/en/typescript-sdk/process/#listptysessions)
>
> [**list_pty_sessions (Ruby SDK)**](/docs/en/ruby-sdk/process/#list_pty_sessions)
>
> [**ListPtySessions (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ListPtySessions)
>
> [**List PTY sessions (API)**](/docs/en/tools/api#daytona-toolbox/tag/process/GET/process/pty)

## Get PTY session info

Daytona provides methods to get information about a specific PTY session, allowing you to retrieve comprehensive information about a specific PTY session including its current state, configuration, and metadata.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Get details about a specific PTY session
session_info = sandbox.process.get_pty_session_info("my-session")

print(f"Session ID: {session_info.id}")
print(f"Active: {session_info.active}")
print(f"Working Directory: {session_info.cwd}")
print(f"Terminal Size: {session_info.cols}x{session_info.rows}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Get details about a specific PTY session
const session = await sandbox.process.getPtySessionInfo('my-session')

console.log(`Session ID: ${session.id}`)
console.log(`Active: ${session.active}`)
console.log(`Working Directory: ${session.cwd}`)
console.log(`Terminal Size: ${session.cols}x${session.rows}`)

if (session.processId) {
  console.log(`Process ID: ${session.processId}`)
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Get details about a specific PTY session
session_info = sandbox.process.get_pty_session_info('my-session')

puts "Session ID: #{session_info.id}"
puts "Active: #{session_info.active}"
puts "Working Directory: #{session_info.cwd}"
puts "Terminal Size: #{session_info.cols}x#{session_info.rows}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Get details about a specific PTY session
session, err := sandbox.Process.GetPtySessionInfo(ctx, "my-session")
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Session ID: %s\n", session.Id)
fmt.Printf("Active: %t\n", session.Active)
fmt.Printf("Working Directory: %s\n", session.Cwd)
fmt.Printf("Terminal Size: %dx%d\n", session.Cols, session.Rows)

if session.ProcessId != nil {
	fmt.Printf("Process ID: %d\n", *session.ProcessId)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/sync/process/), [TypeScript SDK](/docs/en/typescript-sdk/process/), [Ruby SDK](/docs/en/ruby-sdk/process/), [Go SDK](/docs/en/go-sdk/daytona/#type-processservice), and [API](/docs/en/tools/api#daytona-toolbox/tag/process) references:

> [**get_pty_session_info (Python SDK)**](/docs/en/python-sdk/sync/process/#processget_pty_session_info)
>
> [**getPtySessionInfo (TypeScript SDK)**](/docs/en/typescript-sdk/process/#getptysessioninfo)
>
> [**get_pty_session_info (Ruby SDK)**](/docs/en/ruby-sdk/process/#get_pty_session_info)
>
> [**GetPtySessionInfo (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.GetPtySessionInfo)
>
> [**Get PTY session info (API)**](/docs/en/tools/api#daytona-toolbox/tag/process/GET/process/session/{sessionId})

## Kill PTY session

Daytona provides methods to kill a PTY session, allowing you to forcefully terminate a PTY session and cleans up all associated resources.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Kill a specific PTY session
sandbox.process.kill_pty_session("my-session")

# Verify the session no longer exists
pty_sessions = sandbox.process.list_pty_sessions()
for pty_session in pty_sessions:
    print(f"PTY session: {pty_session.id}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Kill a specific PTY session
await sandbox.process.killPtySession('my-session')

// Verify the session is no longer active
try {
  const info = await sandbox.process.getPtySessionInfo('my-session')
  console.log(`Session still exists but active: ${info.active}`)
} catch (error) {
  console.log('Session has been completely removed')
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Delete a specific PTY session
sandbox.process.delete_pty_session('my-session')

# Verify the session no longer exists
sessions = sandbox.process.list_pty_sessions
sessions.each do |session|
  puts "PTY session: #{session.id}"
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Kill a specific PTY session
err := sandbox.Process.KillPtySession(ctx, "my-session")
if err != nil {
	log.Fatal(err)
}

// Verify the session is no longer active
sessions, err := sandbox.Process.ListPtySessions(ctx)
if err != nil {
	log.Fatal(err)
}

for _, session := range sessions {
	fmt.Printf("PTY session: %s\n", session.Id)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}' \
  --request DELETE
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/sync/process/), [TypeScript SDK](/docs/en/typescript-sdk/process/), [Ruby SDK](/docs/en/ruby-sdk/process/#delete_pty_session), [Go SDK](/docs/en/go-sdk/daytona/#type-processservice), and [API](/docs/en/tools/api#daytona-toolbox/tag/process) references:

> [**kill_pty_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processkill_pty_session)
>
> [**killPtySession (TypeScript SDK)**](/docs/en/typescript-sdk/process/#killptysession)
>
> [**delete_pty_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#delete_pty_session)
>
> [**KillPtySession (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.KillPtySession)
>
> [**Kill PTY session (API)**](/docs/en/tools/api#daytona-toolbox/tag/process/DELETE/process/session/{sessionId})

## Resize PTY session

Daytona provides methods to resize a PTY session, allowing you to change the terminal dimensions of an active PTY session. This sends a SIGWINCH signal to the shell process, allowing terminal applications to adapt to the new size.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona.common.pty import PtySize

# Resize a PTY session to a larger terminal
new_size = PtySize(rows=40, cols=150)
updated_info = sandbox.process.resize_pty_session("my-session", new_size)

print(f"Terminal resized to {updated_info.cols}x{updated_info.rows}")

# You can also use the PtyHandle's resize method
pty_handle.resize(new_size)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Resize a PTY session to a larger terminal
const updatedInfo = await sandbox.process.resizePtySession('my-session', 150, 40)
console.log(`Terminal resized to ${updatedInfo.cols}x${updatedInfo.rows}`)

// You can also use the PtyHandle's resize method
await ptyHandle.resize(150, 40) // cols, rows
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Resize a PTY session to a larger terminal
pty_size = Daytona::PtySize.new(rows: 40, cols: 150)
session_info = sandbox.process.resize_pty_session('my-session', pty_size)

puts "Terminal resized to #{session_info.cols}x#{session_info.rows}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Resize a PTY session to a larger terminal
updatedInfo, err := sandbox.Process.ResizePtySession(ctx, "my-session", types.PtySize{
	Cols: 150,
	Rows: 40,
})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Terminal resized to %dx%d\n", updatedInfo.Cols, updatedInfo.Rows)

// You can also use the PtyHandle's Resize method
info, err := handle.Resize(ctx, 150, 40)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Terminal resized to %dx%d\n", info.Cols, info.Rows)
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/pty/{sessionId}/resize' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "cols": 1,
  "rows": 1
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/sync/process/), [TypeScript SDK](/docs/en/typescript-sdk/process/#resizeptysession), [Ruby SDK](/docs/en/ruby-sdk/process/), [Go SDK](/docs/en/go-sdk/daytona/#type-processservice), and [API](/docs/en/tools/api#daytona-toolbox/tag/process) references:

> [**resize_pty_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processresize_pty_session)
>
> [**resizePtySession (TypeScript SDK)**](/docs/en/typescript-sdk/process/#resizeptysession)
>
> [**resize_pty_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#resize_pty_session)
>
> [**ResizePtySession (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ResizePtySession)
>
> [**Resize PTY session (API)**](/docs/en/tools/api#daytona-toolbox/tag/process/POST/process/pty/{sessionId}/resize)

## Interactive commands

Daytona provides methods to handle interactive commands with PTY sessions, allowing you to handle interactive commands that require user input and can be resized during execution.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
import time
from daytona import Daytona, Sandbox
from daytona.common.pty import PtySize

def handle_pty_data(data: bytes):
    text = data.decode("utf-8", errors="replace")
    print(text, end="")

# Create PTY session
pty_handle = sandbox.process.create_pty_session(
    id="interactive-session",
    pty_size=PtySize(cols=300, rows=100)
)

# Send interactive command
pty_handle.send_input('printf "Are you accepting the terms and conditions? (y/n): " && read confirm && if [ "$confirm" = "y" ]; then echo "You accepted"; else echo "You did not accept"; fi\n')
time.sleep(1)
pty_handle.send_input("y\n")

# Resize terminal
pty_session_info = pty_handle.resize(PtySize(cols=210, rows=110))
print(f"PTY session resized to {pty_session_info.cols}x{pty_session_info.rows}")

# Exit the session
pty_handle.send_input('exit\n')

# Handle output using iterator
for data in pty_handle:
    handle_pty_data(data)

print(f"Session completed with exit code: {pty_handle.exit_code}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona, Sandbox } from '@daytonaio/sdk'

// Create PTY session
const ptyHandle = await sandbox.process.createPty({
  id: 'interactive-session',
  cols: 300,
  rows: 100,
  onData: data => {
    const text = new TextDecoder().decode(data)
    process.stdout.write(text)
  },
})

await ptyHandle.waitForConnection()

// Send interactive command
await ptyHandle.sendInput(
  'printf "Are you accepting the terms and conditions? (y/n): " && read confirm && if [ "$confirm" = "y" ]; then echo "You accepted"; else echo "You did not accept"; fi\n'
)
await new Promise(resolve => setTimeout(resolve, 1000))
await ptyHandle.sendInput('y\n')

// Resize terminal
const ptySessionInfo = await sandbox.process.resizePtySession(
  'interactive-session',
  210,
  110
)
console.log(
  `\nPTY session resized to ${ptySessionInfo.cols}x${ptySessionInfo.rows}`
)

// Exit the session
await ptyHandle.sendInput('exit\n')

// Wait for completion
const result = await ptyHandle.wait()
console.log(`Session completed with exit code: ${result.exitCode}`)
```

</TabItem>

<TabItem label="Ruby" icon="seti:ruby">
```ruby
require 'daytona'

# Create PTY session
pty_handle = sandbox.process.create_pty_session(
  id: 'interactive-session',
  pty_size: Daytona::PtySize.new(cols: 300, rows: 100)
)

# Handle output in a separate thread
thread = Thread.new do
  pty_handle.each { |data| print data }
end

# Send interactive command
pty_handle.send_input('printf "Are you accepting the terms and conditions? (y/n): " && read confirm && if [ "$confirm" = "y" ]; then echo "You accepted"; else echo "You did not accept"; fi' + "\n")
sleep(1)
pty_handle.send_input("y\n")

# Resize terminal
pty_handle.resize(Daytona::PtySize.new(cols: 210, rows: 110))
puts "\nPTY session resized"

# Exit the session
pty_handle.send_input("exit\n")

# Wait for the thread to finish
thread.join

puts "Session completed with exit code: #{pty_handle.exit_code}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Create PTY session
handle, err := sandbox.Process.CreatePty(ctx, "interactive-session",
	options.WithCreatePtySize(types.PtySize{Cols: 300, Rows: 100}),
)
if err != nil {
	log.Fatal(err)
}
defer handle.Disconnect()

if err := handle.WaitForConnection(ctx); err != nil {
	log.Fatal(err)
}

// Handle output in a goroutine
go func() {
	for data := range handle.DataChan() {
		fmt.Print(string(data))
	}
}()

// Send interactive command
handle.SendInput([]byte(`printf "Are you accepting the terms and conditions? (y/n): " && read confirm && if [ "$confirm" = "y" ]; then echo "You accepted"; else echo "You did not accept"; fi` + "\n"))
time.Sleep(1 * time.Second)
handle.SendInput([]byte("y\n"))

// Resize terminal
info, err := handle.Resize(ctx, 210, 110)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("\nPTY session resized to %dx%d\n", info.Cols, info.Rows)

// Exit the session
handle.SendInput([]byte("exit\n"))

// Wait for completion
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Session completed with exit code: %d\n", *result.ExitCode)
```

</TabItem>
</Tabs>

## Long-running processes

Daytona provides methods to manage long-running processes with PTY sessions, allowing you to manage long-running processes that need to be monitored or terminated.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
import time
import threading
from daytona import Daytona, Sandbox
from daytona.common.pty import PtySize

def handle_pty_data(data: bytes):
    text = data.decode("utf-8", errors="replace")
    print(text, end="")

# Create PTY session
pty_handle = sandbox.process.create_pty_session(
    id="long-running-session",
    pty_size=PtySize(cols=120, rows=30)
)

# Start a long-running process
pty_handle.send_input('while true; do echo "Running... $(date)"; sleep 1; done\n')

# Using thread and wait() method to handle PTY output
thread = threading.Thread(target=pty_handle.wait, args=(handle_pty_data, 10))
thread.start()

time.sleep(3)  # Let it run for a bit

print("Killing long-running process...")
pty_handle.kill()

thread.join()

print(f"\nProcess terminated with exit code: {pty_handle.exit_code}")
if pty_handle.error:
    print(f"Termination reason: {pty_handle.error}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona, Sandbox } from '@daytonaio/sdk'

// Create PTY session
const ptyHandle = await sandbox.process.createPty({
  id: 'long-running-session',
  cols: 120,
  rows: 30,
  onData: (data) => {
    const text = new TextDecoder().decode(data)
    process.stdout.write(text)
  },
})

await ptyHandle.waitForConnection()

// Start a long-running process
await ptyHandle.sendInput('while true; do echo "Running... $(date)"; sleep 1; done\n')
await new Promise(resolve => setTimeout(resolve, 3000)) // Let it run for a bit

console.log('Killing long-running process...')
await ptyHandle.kill()

// Wait for termination
const result = await ptyHandle.wait()
console.log(`\nProcess terminated with exit code: ${result.exitCode}`)
if (result.error) {
    console.log(`Termination reason: ${result.error}`)
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

# Create PTY session
pty_handle = sandbox.process.create_pty_session(
  id: 'long-running-session',
  pty_size: Daytona::PtySize.new(cols: 120, rows: 30)
)

# Handle output in a separate thread
thread = Thread.new do
  pty_handle.each { |data| print data }
end

# Start a long-running process
pty_handle.send_input("while true; do echo \"Running... $(date)\"; sleep 1; done\n")
sleep(3) # Let it run for a bit

puts "Killing long-running process..."
pty_handle.kill

thread.join

puts "\nProcess terminated with exit code: #{pty_handle.exit_code}"
puts "Termination reason: #{pty_handle.error}" if pty_handle.error
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Create PTY session
handle, err := sandbox.Process.CreatePty(ctx, "long-running-session",
	options.WithCreatePtySize(types.PtySize{Cols: 120, Rows: 30}),
)
if err != nil {
	log.Fatal(err)
}
defer handle.Disconnect()

if err := handle.WaitForConnection(ctx); err != nil {
	log.Fatal(err)
}

// Handle output in a goroutine
go func() {
	for data := range handle.DataChan() {
		fmt.Print(string(data))
	}
}()

// Start a long-running process
handle.SendInput([]byte(`while true; do echo "Running... $(date)"; sleep 1; done` + "\n"))
time.Sleep(3 * time.Second) // Let it run for a bit

fmt.Println("Killing long-running process...")
if err := handle.Kill(ctx); err != nil {
	log.Fatal(err)
}

// Wait for termination
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("\nProcess terminated with exit code: %d\n", *result.ExitCode)
if result.Error != nil {
	fmt.Printf("Termination reason: %s\n", *result.Error)
}
```

</TabItem>
</Tabs>

## Resource management

Daytona provides methods to manage resource leaks with PTY sessions, allowing you to always clean up PTY sessions to prevent resource leaks.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
```python
# Python: Use try/finally
pty_handle = None
try:
    pty_handle = sandbox.process.create_pty_session(id="session", pty_size=PtySize(cols=120, rows=30))
    # Do work...
finally:
    if pty_handle:
        pty_handle.kill()
```
</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
```typescript
// TypeScript: Use try/finally
let ptyHandle
try {
  ptyHandle = await sandbox.process.createPty({
    id: 'session',
    cols: 120,
    rows: 30,
  })
  // Do work...
} finally {
  if (ptyHandle) await ptyHandle.kill()
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Ruby: Use begin/ensure
pty_handle = nil
begin
  pty_handle = sandbox.process.create_pty_session(
    id: 'session',
    pty_size: Daytona::PtySize.new(cols: 120, rows: 30)
  )
  # Do work...
ensure
  pty_handle&.kill
end
```
</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Go: Use defer for cleanup
handle, err := sandbox.Process.CreatePty(ctx, "session",
	options.WithCreatePtySize(types.PtySize{Cols: 120, Rows: 30}),
)
if err != nil {
	log.Fatal(err)
}
defer handle.Disconnect()

// Do work...

// Or use Kill to terminate the process
defer handle.Kill(ctx)
```

</TabItem>
</Tabs>

## PtyHandle methods

Daytona provides methods to interact with PTY sessions, allowing you to send input, resize the terminal, wait for completion, and manage the WebSocket connection to a PTY session.

### Send input

Daytona provides methods to send input to a PTY session, allowing you to send input data (keystrokes or commands) to the PTY session.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Send a command
pty_handle.send_input("ls -la\n")

# Send user input
pty_handle.send_input("y\n")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Send a command
await ptyHandle.sendInput('ls -la\n')

// Send raw bytes
await ptyHandle.sendInput(new Uint8Array([3])) // Ctrl+C
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Send a command
pty_handle.send_input("ls -la\n")

# Send user input
pty_handle.send_input("y\n")
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Send a command
handle.SendInput([]byte("ls -la\n"))

// Send Ctrl+C
handle.SendInput([]byte{0x03})
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**sendInput (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#sendinput)
>
> [**SendInput (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.SendInput)

### Wait for completion

Daytona provides methods to wait for a PTY process to exit and return the result, allowing you to wait for a PTY process to exit and return the result.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Wait with a callback for output data
def handle_data(data: bytes):
    print(data.decode("utf-8", errors="replace"), end="")

result = pty_handle.wait(on_data=handle_data, timeout=30)
print(f"Exit code: {result.exit_code}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Wait for process to complete
const result = await ptyHandle.wait()

if (result.exitCode === 0) {
  console.log('Process completed successfully')
} else {
  console.log(`Process failed with code: ${result.exitCode}`)
  if (result.error) {
    console.log(`Error: ${result.error}`)
  }
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Wait by iterating over output (blocks until PTY session ends)
pty_handle.each { |data| print data }

if pty_handle.exit_code == 0
  puts 'Process completed successfully'
else
  puts "Process failed with code: #{pty_handle.exit_code}"
  puts "Error: #{pty_handle.error}" if pty_handle.error
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Wait for process to complete
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}

if result.ExitCode != nil && *result.ExitCode == 0 {
	fmt.Println("Process completed successfully")
} else {
	fmt.Printf("Process failed with code: %d\n", *result.ExitCode)
	if result.Error != nil {
		fmt.Printf("Error: %s\n", *result.Error)
	}
}
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**wait (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#wait)
>
> [**Wait (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.Wait)

### Wait for connection

Daytona provides methods to wait for the WebSocket connection to be established before sending input, allowing you to wait for the WebSocket connection to be established before sending input.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Python handles connection internally during creation
# No explicit wait needed
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Wait for connection to be established
await ptyHandle.waitForConnection()

// Now safe to send input
await ptyHandle.sendInput('echo "Connected!"\n')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Ruby handles connection internally during creation
# No explicit wait needed - can send input immediately after creation
pty_handle.send_input("echo 'Connected!'\n")
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Wait for connection to be established
if err := handle.WaitForConnection(ctx); err != nil {
	log.Fatal(err)
}

// Now safe to send input
handle.SendInput([]byte("echo 'Connected!'\n"))
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**waitForConnection (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#waitforconnection)
>
> [**WaitForConnection (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.WaitForConnection)

### Kill PTY process

Daytona provides methods to kill a PTY process and terminate the session from the handle.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
pty_handle.kill()
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Kill a long-running process
await ptyHandle.kill()

// Wait to confirm termination
const result = await ptyHandle.wait()
console.log(`Process terminated with exit code: ${result.exitCode}`)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Kill a long-running process
pty_handle.kill

puts "Process terminated with exit code: #{pty_handle.exit_code}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Kill a long-running process
if err := handle.Kill(ctx); err != nil {
	log.Fatal(err)
}

// Wait to confirm termination
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Process terminated with exit code: %d\n", *result.ExitCode)
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**kill (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#kill)
>
> [**Kill (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.Kill)

### Resize from handle

Daytona provides methods to resize the PTY terminal dimensions directly from the handle.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona.common.pty import PtySize

pty_handle.resize(PtySize(cols=120, rows=30))
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Resize to 120x30
await ptyHandle.resize(120, 30)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Resize to 120x30
pty_handle.resize(Daytona::PtySize.new(cols: 120, rows: 30))
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Resize to 120x30
info, err := handle.Resize(ctx, 120, 30)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Resized to %dx%d\n", info.Cols, info.Rows)
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**resize (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#resize)
>
> [**Resize (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.Resize)

### Disconnect

Daytona provides methods to disconnect from a PTY session and clean up resources without killing the process.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Python: Use kill() to terminate, or let the handle go out of scope
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Always clean up when done
try {
  // ... use PTY session
} finally {
  await ptyHandle.disconnect()
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Ruby: Use begin/ensure or kill the session
begin
  # ... use PTY session
ensure
  pty_handle.kill
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Always clean up when done using defer
handle, err := sandbox.Process.CreatePty(ctx, "session")
if err != nil {
	log.Fatal(err)
}
defer handle.Disconnect()

// ... use PTY session
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**disconnect (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#disconnect)
>
> [**Disconnect (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.Disconnect)

### Check connection status

Daytona provides methods to check if a PTY session is still connected.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Python: Check by attempting operations or using session info
session_info = sandbox.process.get_pty_session_info("my-session")
print(f"Session active: {session_info.active}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
if (ptyHandle.isConnected()) {
  console.log('PTY session is active')
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Ruby: Check by using session info
session_info = sandbox.process.get_pty_session_info('my-session')
puts 'PTY session is active' if session_info.active
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
if handle.IsConnected() {
	fmt.Println("PTY session is active")
}
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**isConnected (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#isconnected)
>
> [**IsConnected (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.IsConnected)

### Exit code and error

Daytona provides methods to access the exit code and error message after a PTY process terminates.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# After iteration or wait completes
print(f"Exit code: {pty_handle.exit_code}")
if pty_handle.error:
    print(f"Error: {pty_handle.error}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Access via getters after process terminates
console.log(`Exit code: ${ptyHandle.exitCode}`)
if (ptyHandle.error) {
  console.log(`Error: ${ptyHandle.error}`)
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Access after process terminates
puts "Exit code: #{pty_handle.exit_code}"
puts "Error: #{pty_handle.error}" if pty_handle.error
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Access via methods after process terminates
if exitCode := handle.ExitCode(); exitCode != nil {
	fmt.Printf("Exit code: %d\n", *exitCode)
}
if errMsg := handle.Error(); errMsg != nil {
	fmt.Printf("Error: %s\n", *errMsg)
}
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/daytona/#type-processservice) references:

> [**exitCode (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#exitcode)
>
> [**error (TypeScript SDK)**](/docs/en/typescript-sdk/pty-handle/#error)
>
> [**ExitCode (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.ExitCode)
>
> [**Error (Go SDK)**](/docs/en/go-sdk/daytona/#PtyHandle.Error)

### Iterate over output (Python)

Daytona provides methods to iterate over a PTY handle to receive output data.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Iterate over PTY output
for data in pty_handle:
    text = data.decode("utf-8", errors="replace")
    print(text, end="")

print(f"Session ended with exit code: {pty_handle.exit_code}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// TypeScript uses the onData callback instead
const ptyHandle = await sandbox.process.createPty({
  id: 'my-session',
  onData: (data) => {
    const text = new TextDecoder().decode(data)
    process.stdout.write(text)
  },
})
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Iterate over PTY output
pty_handle.each do |data|
  print data
end

puts "Session ended with exit code: #{pty_handle.exit_code}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Go uses a channel to receive output data
for data := range handle.DataChan() {
	fmt.Print(string(data))
}

// Or use as io.Reader
io.Copy(os.Stdout, handle)

fmt.Printf("Session ended with exit code: %d\n", *handle.ExitCode())
```

</TabItem>
</Tabs>

## Error handling

Daytona provides methods to monitor exit codes and handle errors appropriately with PTY sessions.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Python: Check exit codes
result = pty_handle.wait()
if result.exit_code != 0:
    print(f"Command failed: {result.exit_code}")
    print(f"Error: {result.error}")
```
</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
```typescript
// TypeScript: Check exit codes
const result = await ptyHandle.wait()
if (result.exitCode !== 0) {
  console.log(`Command failed: ${result.exitCode}`)
  console.log(`Error: ${result.error}`)
}
```
</TabItem>

<TabItem label="Ruby" icon="seti:ruby">
```ruby
# Ruby: Check exit codes
# The handle blocks until the PTY session completes
pty_handle.each { |data| print data }

if pty_handle.exit_code != 0
  puts "Command failed: #{pty_handle.exit_code}"
  puts "Error: #{pty_handle.error}"
end
```
</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Go: Check exit codes
result, err := handle.Wait(ctx)
if err != nil {
	log.Fatal(err)
}

if result.ExitCode != nil && *result.ExitCode != 0 {
	fmt.Printf("Command failed: %d\n", *result.ExitCode)
	if result.Error != nil {
		fmt.Printf("Error: %s\n", *result.Error)
	}
}
```

</TabItem>
</Tabs>

## Troubleshooting

- **Connection issues**: verify sandbox status, network connectivity, and proper session IDs
- **Performance issues**: use appropriate terminal dimensions and efficient data handlers
- **Process management**: use explicit `kill()` calls and proper timeout handling for long-running processes
