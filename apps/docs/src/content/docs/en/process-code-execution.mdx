---
title: Process and Code Execution
---

import { Tabs, TabItem } from '@astrojs/starlight/components';

Daytona provides process and code execution capabilities through the `process` module in sandboxes.

## Code execution

Daytona provides methods to execute code in sandboxes. You can run code snippets in multiple languages with support for both stateless execution and stateful interpretation with persistent contexts.

:::note

Stateless execution inherits the sandbox language that you choose at [sandbox creation](/docs/en/sandboxes#create-sandboxes) time. The stateful interpreter supports only Python.
  :::

### Run code (stateless)

Daytona provides methods to run code snippets in sandboxes using stateless execution. Each invocation starts from a clean interpreter, making it ideal for independent code snippets.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
```python
# Run Python code
response = sandbox.process.code_run('''
def greet(name):
    return f"Hello, {name}!"

print(greet("Daytona"))
''')

print(response.result)
```
</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Run TypeScript code
let response = await sandbox.process.codeRun(`
function greet(name: string): string {
    return \`Hello, \${name}!\`;
}

console.log(greet("Daytona"));
`);
console.log(response.result);

// Run code with argv and environment variables
response = await sandbox.process.codeRun(
    `
    console.log(\`Hello, \${process.argv[2]}!\`);
    console.log(\`FOO: \${process.env.FOO}\`);
    `,
    { 
      argv: ["Daytona"],
      env: { FOO: "BAR" }
    }
);
console.log(response.result);

// Run code with timeout (5 seconds)
response = await sandbox.process.codeRun(
    'setTimeout(() => console.log("Done"), 2000);',
    undefined,
    5
);
console.log(response.result);
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Run Python code
response = sandbox.process.code_run(code: <<~PYTHON)
  def greet(name):
      return f"Hello, {name}!"

  print(greet("Daytona"))
PYTHON

puts response.result
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Run code using shell command execution
// Note: For stateless code execution in Go, use ExecuteCommand with the appropriate interpreter
result, err := sandbox.Process.ExecuteCommand(ctx, `python3 -c '
def greet(name):
    return f"Hello, {name}!"

print(greet("Daytona"))
'`)
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.Result)

// Run code with environment variables
result, err = sandbox.Process.ExecuteCommand(ctx, `python3 -c 'import os; print(f"FOO: {os.environ.get(\"FOO\")}")'`,
	options.WithCommandEnv(map[string]string{"FOO": "BAR"}),
)
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.Result)

// Run code with timeout
result, err = sandbox.Process.ExecuteCommand(ctx, `python3 -c 'import time; time.sleep(2); print("Done")'`,
	options.WithExecuteTimeout(5*time.Second),
)
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.Result)
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/code-run' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "code": "def greet(name):\n    return f\"Hello, {name}!\"\n\nprint(greet(\"Daytona\"))",
  "env": {
    "FOO": "BAR"
  },
  "timeout": 5000
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**code_run (Python SDK)**](/docs/en/python-sdk/sync/process/#processcode_run)
>
> [**codeRun (TypeScript SDK)**](/docs/en/typescript-sdk/process/#coderun)
>
> [**code_run (Ruby SDK)**](/docs/en/ruby-sdk/process/#code_run)
>
> [**ExecuteCommand (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ExecuteCommand)
>
> [**execute command (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/execute)

### Run code (stateful)

Daytona provides methods to run code with persistent state using the code interpreter. You can maintain variables and imports between calls, create isolated contexts, and control environment variables.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona, OutputMessage

def handle_stdout(message: OutputMessage):
    print(f"[STDOUT] {message.output}")

daytona = Daytona()
sandbox = daytona.create()

# Shared default context
result = sandbox.code_interpreter.run_code(
    "counter = 1\nprint(f'Counter initialized at {counter}')",
    on_stdout=handle_stdout,
)

# Isolated context
ctx = sandbox.code_interpreter.create_context()
try:
    sandbox.code_interpreter.run_code(
        "value = 'stored in ctx'",
        context=ctx,
    )
    sandbox.code_interpreter.run_code(
        "print(value)",
        context=ctx,
        on_stdout=handle_stdout,
    )
finally:
    sandbox.code_interpreter.delete_context(ctx)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona } from '@daytonaio/sdk'

const daytona = new Daytona()

async function main() {
    const sandbox = await daytona.create()

    // Shared default context
    await sandbox.codeInterpreter.runCode(
`
counter = 1
print(f'Counter initialized at {counter}')
`,
        { onStdout: (msg) => process.stdout.write(`[STDOUT] ${msg.output}`)},
    )

    // Isolated context
    const ctx = await sandbox.codeInterpreter.createContext()
    try {
    await sandbox.codeInterpreter.runCode(
        `value = 'stored in ctx'`,
        { context: ctx },
    )
    await sandbox.codeInterpreter.runCode(
        `print(value)`,
        { context: ctx, onStdout: (msg) => process.stdout.write(`[STDOUT] ${msg.output}`) },
    )
    } finally {
    await sandbox.codeInterpreter.deleteContext(ctx)
    }
}

main()
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Ruby SDK uses process.code_run for code execution
# Stateful contexts are managed through the code interpreter API

require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create

# Run code (stateless in Ruby SDK)
response = sandbox.process.code_run(code: <<~PYTHON)
  counter = 1
  print(f'Counter initialized at {counter}')
PYTHON

puts response.result
```
</TabItem>
<TabItem label="Go" icon="seti:go">
```go
// Create a code interpreter context
ctxInfo, err := sandbox.CodeInterpreter.CreateContext(ctx, nil)
if err != nil {
	log.Fatal(err)
}
contextID := ctxInfo["id"].(string)

// Run code in the context
channels, err := sandbox.CodeInterpreter.RunCode(ctx,
	"counter = 1\nprint(f'Counter initialized at {counter}')",
	options.WithCustomContext(contextID),
)
if err != nil {
	log.Fatal(err)
}

// Read output
for msg := range channels.Stdout {
	fmt.Printf("[STDOUT] %s\n", msg.Text)
}

// Clean up context
err = sandbox.CodeInterpreter.DeleteContext(ctx, contextID)
if err != nil {
	log.Fatal(err)
}
```
</TabItem>
<TabItem label="API" icon="seti:json">

```bash
# Create context
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/interpreter/context' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{}'

# Run code in context (WebSocket endpoint)
# Connect via WebSocket to:
# wss://proxy.app.daytona.io/toolbox/{sandboxId}/process/interpreter/execute
# Send JSON message:
# {
#   "code": "counter = 1\nprint(f\"Counter initialized at {counter}\")",
#   "contextId": "your-context-id"
# }

# Delete context
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/interpreter/context/{contextId}' \
  --request DELETE
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**run_code (Python SDK)**](/docs/en/python-sdk/sync/code-interpreter/#codeinterpreterrun_code)
>
> [**create_context (Python SDK)**](/docs/en/python-sdk/sync/code-interpreter/#codeinterpretercreate_context)
>
> [**delete_context (Python SDK)**](/docs/en/python-sdk/sync/code-interpreter/#codeinterpreterdelete_context)
>
> [**runCode (TypeScript SDK)**](/docs/en/typescript-sdk/code-interpreter/#runcode)
>
> [**createContext (TypeScript SDK)**](/docs/en/typescript-sdk/code-interpreter/#createcontext)
>
> [**deleteContext (TypeScript SDK)**](/docs/en/typescript-sdk/code-interpreter/#deletecontext)
>
> [**code_run (Ruby SDK)**](/docs/en/ruby-sdk/process/#code_run)
>
> [**RunCode (Go SDK)**](/docs/en/go-sdk/daytona/#CodeInterpreterService.RunCode)
>
> [**CreateContext (Go SDK)**](/docs/en/go-sdk/daytona/#CodeInterpreterService.CreateContext)
>
> [**DeleteContext (Go SDK)**](/docs/en/go-sdk/daytona/#CodeInterpreterService.DeleteContext)
>
> [**code interpreter (API)**](/docs/en/tools/api/#daytona-toolbox/tag/interpreter)

## Command execution

Daytona provides methods to execute shell commands in sandboxes. You can run commands with working directory, timeout, and environment variable options.

The working directory defaults to the sandbox working directory. It uses the WORKDIR specified in the Dockerfile if present, or falls back to the user's home directory if not (e.g., `workspace/repo` implies `/home/daytona/workspace/repo`). You can override it with an absolute path by starting the path with `/`.

### Execute commands

Daytona provides methods to execute shell commands in sandboxes by providing the command string and optional parameters for working directory, timeout, and environment variables. You can also use the `daytona exec` CLI command for quick command execution.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Execute any shell command
response = sandbox.process.exec("ls -la")
print(response.result)

# Setting a working directory and a timeout

response = sandbox.process.exec("sleep 3", cwd="workspace/src", timeout=5)
print(response.result)

# Passing environment variables

response = sandbox.process.exec("echo $CUSTOM_SECRET", env={
        "CUSTOM_SECRET": "DAYTONA"
    }
)
print(response.result)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript

// Execute any shell command
const response = await sandbox.process.executeCommand("ls -la");
console.log(response.result);

// Setting a working directory and a timeout
const response2 = await sandbox.process.executeCommand("sleep 3", "workspace/src", undefined, 5);
console.log(response2.result);

// Passing environment variables
const response3 = await sandbox.process.executeCommand("echo $CUSTOM_SECRET", ".", {
        "CUSTOM_SECRET": "DAYTONA"
    }
);
console.log(response3.result);
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Execute any shell command
response = sandbox.process.exec(command: 'ls -la')
puts response.result

# Setting a working directory and a timeout
response = sandbox.process.exec(command: 'sleep 3', cwd: 'workspace/src', timeout: 5)
puts response.result

# Passing environment variables
response = sandbox.process.exec(
  command: 'echo $CUSTOM_SECRET',
  env: { 'CUSTOM_SECRET' => 'DAYTONA' }
)
puts response.result
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Execute any shell command
response, err := sandbox.Process.ExecuteCommand(ctx, "ls -la")
if err != nil {
	log.Fatal(err)
}
fmt.Println(response.Result)

// Setting a working directory and a timeout
response, err = sandbox.Process.ExecuteCommand(ctx, "sleep 3",
	options.WithCwd("workspace/src"),
	options.WithExecuteTimeout(5*time.Second),
)
if err != nil {
	log.Fatal(err)
}
fmt.Println(response.Result)

// Passing environment variables
response, err = sandbox.Process.ExecuteCommand(ctx, "echo $CUSTOM_SECRET",
	options.WithCommandEnv(map[string]string{"CUSTOM_SECRET": "DAYTONA"}),
)
if err != nil {
	log.Fatal(err)
}
fmt.Println(response.Result)
```

</TabItem>
<TabItem label="CLI" icon="seti:shell">

```bash
# Execute any shell command
daytona exec my-sandbox -- ls -la

# Setting a working directory and a timeout
daytona exec my-sandbox --cwd workspace/src --timeout 5 -- sleep 3

# Passing environment variables (use shell syntax)
daytona exec my-sandbox -- sh -c 'CUSTOM_SECRET=DAYTONA echo $CUSTOM_SECRET'
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/execute' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "command": "ls -la",
  "cwd": "workspace",
  "timeout": 5
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), [CLI](/docs/en/tools/cli/), and [API](/docs/en/tools/api/) references:

> [**exec (Python SDK)**](/docs/en/python-sdk/sync/process/#processexec)
>
> [**executeCommand (TypeScript SDK)**](/docs/en/typescript-sdk/process/#executecommand)
>
> [**exec (Ruby SDK)**](/docs/en/ruby-sdk/process/#exec)
>
> [**ExecuteCommand (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ExecuteCommand)
>
> [**daytona exec (CLI)**](/docs/en/tools/cli/#daytona-exec)
>
> [**execute command (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/execute)

## Session operations

Daytona provides methods to manage background process sessions in sandboxes. You can create sessions, execute commands, monitor status, and manage long-running processes.

### Get session status

Daytona provides methods to get session status and list all sessions in a sandbox by providing the session ID.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Check session's executed commands
session = sandbox.process.get_session(session_id)
print(f"Session {session_id}:")
for command in session.commands:
    print(f"Command: {command.command}, Exit Code: {command.exit_code}")

# List all running sessions

sessions = sandbox.process.list_sessions()
for session in sessions:
    print(f"Session: {session.session_id}, Commands: {session.commands}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Check session's executed commands
const session = await sandbox.process.getSession(sessionId);
console.log(`Session ${sessionId}:`);
for (const command of session.commands) {
    console.log(`Command: ${command.command}, Exit Code: ${command.exitCode}`);
}

// List all running sessions
const sessions = await sandbox.process.listSessions();
for (const session of sessions) {
    console.log(`Session: ${session.sessionId}, Commands: ${session.commands}`);
}
```

</TabItem>

<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Check session's executed commands
session = sandbox.process.get_session(session_id)
puts "Session #{session_id}:"
session.commands.each do |command|
  puts "Command: #{command.command}, Exit Code: #{command.exit_code}"
end

# List all running sessions
sessions = sandbox.process.list_sessions
sessions.each do |session|
  puts "Session: #{session.session_id}, Commands: #{session.commands}"
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Check session's executed commands
session, err := sandbox.Process.GetSession(ctx, sessionID)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Session %s:\n", sessionID)
commands := session["commands"].([]any)
for _, cmd := range commands {
	cmdMap := cmd.(map[string]any)
	fmt.Printf("Command: %s, Exit Code: %v\n", cmdMap["command"], cmdMap["exitCode"])
}

// List all running sessions
sessions, err := sandbox.Process.ListSessions(ctx)
if err != nil {
	log.Fatal(err)
}
for _, sess := range sessions {
	fmt.Printf("Session: %s, Commands: %v\n", sess["sessionId"], sess["commands"])
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
# Get session info
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}'

# List all sessions
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**get_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processget_session)
>
> [**list_sessions (Python SDK)**](/docs/en/python-sdk/sync/process/#processlist_sessions)
>
> [**getSession (TypeScript SDK)**](/docs/en/typescript-sdk/process/#getsession)
>
> [**listSessions (TypeScript SDK)**](/docs/en/typescript-sdk/process/#listsessions)
>
> [**get_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#get_session)
>
> [**list_sessions (Ruby SDK)**](/docs/en/ruby-sdk/process/#list_sessions)
>
> [**GetSession (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.GetSession)
>
> [**ListSessions (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ListSessions)
>
> [**get session (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/GET/process/session/{sessionId})
>
> [**list sessions (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/GET/process/session)

### Execute interactive commands

Daytona provides methods to execute interactive commands in sessions. You can send input to running commands that expect user interaction, such as confirmations or interactive tools like database CLIs and package managers.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
session_id = "interactive-session"
sandbox.process.create_session(session_id)

# Execute command that requires confirmation
command = sandbox.process.execute_session_command(
    session_id,
    SessionExecuteRequest(
        command='pip uninstall requests',
        run_async=True,
    ),
)

# Stream logs asynchronously
logs_task = asyncio.create_task(
    sandbox.process.get_session_command_logs_async(
        session_id,
        command.cmd_id,
        lambda log: print(f"[STDOUT]: {log}"),
        lambda log: print(f"[STDERR]: {log}"),
    )
)

await asyncio.sleep(1)
# Send input to the command
sandbox.process.send_session_command_input(session_id, command.cmd_id, "y")

# Wait for logs to complete
await logs_task
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const sessionId = 'interactive-session'
await sandbox.process.createSession(sessionId)

// Execute command that requires confirmation
const command = await sandbox.process.executeSessionCommand(sessionId, {
    command: 'pip uninstall requests',
    runAsync: true,
})

// Stream logs asynchronously
const logPromise = sandbox.process.getSessionCommandLogs(
    sessionId,
    command.cmdId!,
    (stdout) => console.log('[STDOUT]:', stdout),
    (stderr) => console.log('[STDERR]:', stderr),
)

await new Promise((resolve) => setTimeout(resolve, 1000))
// Send input to the command
await sandbox.process.sendSessionCommandInput(sessionId, command.cmdId!, 'y')

// Wait for logs to complete
await logPromise
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
session_id = "interactive-session"
sandbox.process.create_session(session_id)

# Execute command that requires confirmation
interactive_command = sandbox.process.execute_session_command(
  session_id: session_id,
  req: Daytona::SessionExecuteRequest.new(
    command: 'pip uninstall requests',
    run_async: true
  )
)

# Wait a moment for the command to start
sleep 1

# Send input to the command
sandbox.process.send_session_command_input(
  session_id: session_id,
  command_id: interactive_command.cmd_id,
  data: "y"
)

# Get logs for the interactive command asynchronously
sandbox.process.get_session_command_logs_async(
  session_id: session_id,
  command_id: interactive_command.cmd_id,
  on_stdout: ->(log) { puts "[STDOUT]: #{log}" },
  on_stderr: ->(log) { puts "[STDERR]: #{log}" }
)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
sessionID := "interactive-session"
err := sandbox.Process.CreateSession(ctx, sessionID)
if err != nil {
	log.Fatal(err)
}

// Execute command that requires confirmation
result, err := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "pip uninstall requests", true)
if err != nil {
	log.Fatal(err)
}
cmdID := result["cmdId"].(string)

// Stream logs asynchronously
stdout := make(chan string)
stderr := make(chan string)

go func() {
	err := sandbox.Process.GetSessionCommandLogsStream(ctx, sessionID, cmdID, stdout, stderr)
	if err != nil {
		log.Println("Log stream error:", err)
	}
}()

time.Sleep(1 * time.Second)

// Note: SendSessionCommandInput is not available in Go SDK
// Use the API endpoint directly for sending input

// Read logs
for msg := range stdout {
	fmt.Printf("[STDOUT]: %s\n", msg)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
# Create session
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{"sessionId": "interactive-session"}'

# Execute session command
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}/exec' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "command": "pip uninstall requests",
  "runAsync": true
}'

# Send input to command
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}/command/{commandId}/input' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "data": "y"
}'

# Get command logs
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}/command/{commandId}/logs'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**create_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processcreate_session)
>
> [**execute_session_command (Python SDK)**](/docs/en/python-sdk/sync/process/#processexecute_session_command)
>
> [**send_session_command_input (Python SDK)**](/docs/en/python-sdk/sync/process/#processsend_session_command_input)
>
> [**get_session_command_logs_async (Python SDK)**](/docs/en/python-sdk/sync/process/#processget_session_command_logs_async)
>
> [**createSession (TypeScript SDK)**](/docs/en/typescript-sdk/process/#createsession)
>
> [**executeSessionCommand (TypeScript SDK)**](/docs/en/typescript-sdk/process/#executesessioncommand)
>
> [**sendSessionCommandInput (TypeScript SDK)**](/docs/en/typescript-sdk/process/#sendsessioncommandinput)
>
> [**getSessionCommandLogs (TypeScript SDK)**](/docs/en/typescript-sdk/process/#getsessioncommandlogs)
>
> [**create_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#create_session)
>
> [**execute_session_command (Ruby SDK)**](/docs/en/ruby-sdk/process/#execute_session_command)
>
> [**send_session_command_input (Ruby SDK)**](/docs/en/ruby-sdk/process/#send_session_command_input)
>
> [**CreateSession (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.CreateSession)
>
> [**ExecuteSessionCommand (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.ExecuteSessionCommand)
>
> [**GetSessionCommandLogsStream (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.GetSessionCommandLogsStream)
>
> [**create session (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/session)
>
> [**execute session command (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/session/{sessionId}/exec)

## Resource management

Daytona provides methods to manage session resources. You should use sessions for long-running operations, clean up sessions after execution, and handle exceptions properly.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
   ```python
   # Python - Clean up session
   session_id = "long-running-cmd"
   try:
       sandbox.process.create_session(session_id)
       session = sandbox.process.get_session(session_id)
       # Do work...
   finally:
       sandbox.process.delete_session(session.session_id)
   ```
</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
   ```typescript
   // TypeScript - Clean up session
   const sessionId = "long-running-cmd";
   try {
       await sandbox.process.createSession(sessionId);
       const session = await sandbox.process.getSession(sessionId);
       // Do work...
   } finally {
       await sandbox.process.deleteSession(session.sessionId);
   }
   ```
</TabItem>

<TabItem label="Ruby" icon="seti:ruby">
   ```ruby
   # Ruby - Clean up session
   session_id = 'long-running-cmd'
   begin
     sandbox.process.create_session(session_id)
     session = sandbox.process.get_session(session_id)
     # Do work...
   ensure
     sandbox.process.delete_session(session.session_id)
   end
   ```
</TabItem>
<TabItem label="Go" icon="seti:go">
   ```go
   // Go - Clean up session
   sessionID := "long-running-cmd"
   err := sandbox.Process.CreateSession(ctx, sessionID)
   if err != nil {
   	log.Fatal(err)
   }
   defer sandbox.Process.DeleteSession(ctx, sessionID)

   session, err := sandbox.Process.GetSession(ctx, sessionID)
   if err != nil {
   	log.Fatal(err)
   }
   // Do work...
   ```
</TabItem>
<TabItem label="API" icon="seti:json">
   ```bash
   # Create session
   curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session' \
     --request POST \
     --header 'Content-Type: application/json' \
     --data '{"sessionId": "long-running-cmd"}'

   # Delete session when done
   curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/session/{sessionId}' \
     --request DELETE
   ```
</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**create_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processcreate_session)
>
> [**delete_session (Python SDK)**](/docs/en/python-sdk/sync/process/#processdelete_session)
>
> [**createSession (TypeScript SDK)**](/docs/en/typescript-sdk/process/#createsession)
>
> [**deleteSession (TypeScript SDK)**](/docs/en/typescript-sdk/process/#deletesession)
>
> [**create_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#create_session)
>
> [**delete_session (Ruby SDK)**](/docs/en/ruby-sdk/process/#delete_session)
>
> [**CreateSession (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.CreateSession)
>
> [**DeleteSession (Go SDK)**](/docs/en/go-sdk/daytona/#ProcessService.DeleteSession)
>
> [**create session (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/POST/process/session)
>
> [**delete session (API)**](/docs/en/tools/api/#daytona-toolbox/tag/process/DELETE/process/session/{sessionId})

## Error handling

Daytona provides methods to handle errors when executing processes. You should handle process exceptions properly, log error details for debugging, and use try-catch blocks for error handling.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
```python
from daytona import DaytonaError

try:
    response = sandbox.process.code_run("invalid python code")
    if response.exit_code != 0:
        print(f"Exit code: {response.exit_code}")
        print(f"Error output: {response.result}")
except DaytonaError as e:
    print(f"Execution failed: {e}")
```
</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
```typescript
import { DaytonaError } from '@daytonaio/sdk'

try {
    const response = await sandbox.process.codeRun("invalid typescript code");
    if (response.exitCode !== 0) {
        console.error("Exit code:", response.exitCode);
        console.error("Error output:", response.result);
    }
} catch (e) {
    if (e instanceof DaytonaError) {
        console.error("Execution failed:", e);
    }
}
```
</TabItem>

<TabItem label="Ruby" icon="seti:ruby">
```ruby
begin
  response = sandbox.process.code_run(code: 'invalid python code')
  if response.exit_code != 0
    puts "Exit code: #{response.exit_code}"
    puts "Error output: #{response.result}"
  end
rescue StandardError => e
  puts "Execution failed: #{e}"
end
```
</TabItem>
<TabItem label="Go" icon="seti:go">
```go
result, err := sandbox.Process.ExecuteCommand(ctx, "python3 -c 'invalid python code'")
if err != nil {
	fmt.Println("Execution failed:", err)
}
if result != nil && result.ExitCode != 0 {
	fmt.Println("Exit code:", result.ExitCode)
	fmt.Println("Error output:", result.Result)
}
```
</TabItem>
<TabItem label="API" icon="seti:json">
```bash
# API responses include exitCode field for error handling
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/execute' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "command": "python3 -c \"invalid python code\""
}'

# Response includes:
# {
#   "result": "",
#   "exitCode": 1
# }
```
</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), and [Go SDK](/docs/en/go-sdk/) references.

## Common issues

Daytona provides solutions for troubleshooting common issues related to process and code execution.

| **Issue**                | **Solutions**                                                                                                   |
| ------------------------ | --------------------------------------------------------------------------------------------------------------- |
| Process execution failed | • Check command syntax<br/>• Verify required dependencies<br/>• Ensure sufficient permissions                   |
| Process timeout          | • Adjust timeout settings<br/>• Optimize long-running operations<br/>• Consider using background processes      |
| Resource limits          | • Monitor process memory usage<br/>• Handle process cleanup properly<br/>• Use appropriate resource constraints |
