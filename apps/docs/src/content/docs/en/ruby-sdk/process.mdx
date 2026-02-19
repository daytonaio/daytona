---
title: "Process"
hideTitleOnPage: true
---

## Process

Initialize a new Process instance

### Constructors

#### new Process()

```ruby
def initialize(code_toolbox:, sandbox_id:, toolbox_api:, get_preview_link:, otel_state:)

```

Initialize a new Process instance

**Parameters**:

- `code_toolbox` _Daytona:SandboxPythonCodeToolbox, Daytona:SandboxTsCodeToolbox_ -
- `sandbox_id` _String_ - The ID of the Sandbox
- `toolbox_api` _DaytonaToolboxApiClient:ProcessApi_ - API client for Sandbox operations
- `get_preview_link` _Proc_ - Function to get preview link for a port
- `otel_state` _Daytona:OtelState, nil_ -

**Returns**:

- `Process` - a new instance of Process

### Methods

#### code_toolbox()

```ruby
def code_toolbox()

```

**Returns**:

- `Daytona:SandboxPythonCodeToolbox, ` - Daytona::SandboxPythonCodeToolbox,

#### sandbox_id()

```ruby
def sandbox_id()

```

**Returns**:

- `String` - The ID of the Sandbox

#### toolbox_api()

```ruby
def toolbox_api()

```

**Returns**:

- `DaytonaToolboxApiClient:ProcessApi` - API client for Sandbox operations

#### get_preview_link()

```ruby
def get_preview_link()

```

**Returns**:

- `Proc` - Function to get preview link for a port

#### exec()

```ruby
def exec(command:, cwd:, env:, timeout:)

```

Execute a shell command in the Sandbox

**Parameters**:

- `command` _String_ - Shell command to execute
- `cwd` _String, nil_ - Working directory for command execution. If not specified, uses the sandbox working directory
- `env` _Hash\<String, String\>, nil_ - Environment variables to set for the command
- `timeout` _Integer, nil_ - Maximum time in seconds to wait for the command to complete. 0 means wait indefinitely

**Returns**:

- `ExecuteResponse` - Command execution results containing exit_code, result, and artifacts

**Examples:**

```ruby
# Simple command
response = sandbox.process.exec("echo 'Hello'")
puts response.artifacts.stdout
=> "Hello\n"

# Command with working directory
result = sandbox.process.exec("ls", cwd: "workspace/src")

# Command with timeout
result = sandbox.process.exec("sleep 10", timeout: 5)

```

#### code_run()

```ruby
def code_run(code:, params:, timeout:)

```

Execute code in the Sandbox using the appropriate language runtime

**Parameters**:

- `code` _String_ - Code to execute
- `params` _CodeRunParams, nil_ - Parameters for code execution
- `timeout` _Integer, nil_ - Maximum time in seconds to wait for the code to complete. 0 means wait indefinitely

**Returns**:

- `ExecuteResponse` - Code execution result containing exit_code, result, and artifacts

**Examples:**

```ruby
# Run Python code
response = sandbox.process.code_run(<<~CODE)
  x = 10
  y = 20
  print(f"Sum: {x + y}")
CODE
puts response.artifacts.stdout  # Prints: Sum: 30

```

#### create_session()

```ruby
def create_session(session_id)

```

Creates a new long-running background session in the Sandbox

Sessions are background processes that maintain state between commands, making them ideal for
scenarios requiring multiple related commands or persistent environment setup.

**Parameters**:

- `session_id` _String_ - Unique identifier for the new session

**Returns**:

- `void`

**Examples:**

```ruby
# Create a new session
session_id = "my-session"
sandbox.process.create_session(session_id)
session = sandbox.process.get_session(session_id)
# Do work...
sandbox.process.delete_session(session_id)

```

#### get_session()

```ruby
def get_session(session_id)

```

Gets a session in the Sandbox

**Parameters**:

- `session_id` _String_ - Unique identifier of the session to retrieve

**Returns**:

- `DaytonaApiClient:Session` - Session information including session_id and commands

**Examples:**

```ruby
session = sandbox.process.get_session("my-session")
session.commands.each do |cmd|
  puts "Command: #{cmd.command}"
end

```

#### get_session_command()

```ruby
def get_session_command(session_id:, command_id:)

```

Gets information about a specific command executed in a session

**Parameters**:

- `session_id` _String_ - Unique identifier of the session
- `command_id` _String_ - Unique identifier of the command

**Returns**:

- `DaytonaApiClient:Command` - Command information including id, command, and exit_code

**Examples:**

```ruby
cmd = sandbox.process.get_session_command(session_id: "my-session", command_id: "cmd-123")
if cmd.exit_code == 0
  puts "Command #{cmd.command} completed successfully"
end

```

#### execute_session_command()

```ruby
def execute_session_command(session_id:, req:)

```

Executes a command in the session

**Parameters**:

- `session_id` _String_ - Unique identifier of the session to use
- `req` _Daytona:SessionExecuteRequest_ - Command execution request containing command and run_async

**Returns**:

- `Daytona:SessionExecuteResponse` - Command execution results containing cmd_id, output, stdout, stderr, and exit_code

**Examples:**

```ruby
# Execute commands in sequence, maintaining state
session_id = "my-session"

# Change directory
req = Daytona::SessionExecuteRequest.new(command: "cd /workspace")
sandbox.process.execute_session_command(session_id:, req:)

# Create a file
req = Daytona::SessionExecuteRequest.new(command: "echo 'Hello' > test.txt")
sandbox.process.execute_session_command(session_id:, req:)

# Read the file
req = Daytona::SessionExecuteRequest.new(command: "cat test.txt")
result = sandbox.process.execute_session_command(session_id:, req:)
puts "Command stdout: #{result.stdout}"
puts "Command stderr: #{result.stderr}"

```

#### get_session_command_logs()

```ruby
def get_session_command_logs(session_id:, command_id:)

```

Get the logs for a command executed in a session

**Parameters**:

- `session_id` _String_ - Unique identifier of the session
- `command_id` _String_ - Unique identifier of the command

**Returns**:

- `Daytona:SessionCommandLogsResponse` - Command logs including output, stdout, and stderr

**Examples:**

```ruby
logs = sandbox.process.get_session_command_logs(session_id: "my-session", command_id: "cmd-123")
puts "Command stdout: #{logs.stdout}"
puts "Command stderr: #{logs.stderr}"

```

#### get_session_command_logs_async()

```ruby
def get_session_command_logs_async(session_id:, command_id:, on_stdout:, on_stderr:)

```

Asynchronously retrieves and processes the logs for a command executed in a session as they become available

**Parameters**:

- `session_id` _String_ - Unique identifier of the session
- `command_id` _String_ - Unique identifier of the command
- `on_stdout` _Proc_ - Callback function to handle stdout log chunks as they arrive
- `on_stderr` _Proc_ - Callback function to handle stderr log chunks as they arrive

**Returns**:

- `WebSocket:Client:Simple:Client`

**Examples:**

```ruby
sandbox.process.get_session_command_logs_async(
  session_id: "my-session",
  command_id: "cmd-123",
  on_stdout: ->(log) { puts "[STDOUT]: #{log}" },
  on_stderr: ->(log) { puts "[STDERR]: #{log}" }
)

```

#### send_session_command_input()

```ruby
def send_session_command_input(session_id:, command_id:, data:)

```

Sends input data to a command executed in a session

This method allows you to send input to an interactive command running in a session,
such as responding to prompts or providing data to stdin.

**Parameters**:

- `session_id` _String_ - Unique identifier of the session
- `command_id` _String_ - Unique identifier of the command
- `data` _String_ - Input data to send to the command

**Returns**:

- `void`

#### list_sessions()

```ruby
def list_sessions()

```

**Returns**:

- `Array\<DaytonaApiClient:Session\>` - List of all sessions in the Sandbox

**Examples:**

```ruby
sessions = sandbox.process.list_sessions
sessions.each do |session|
  puts "Session #{session.session_id}:"
  puts "  Commands: #{session.commands.length}"
end

```

#### delete_session()

```ruby
def delete_session(session_id)

```

Terminates and removes a session from the Sandbox, cleaning up any resources associated with it

**Parameters**:

- `session_id` _String_ - Unique identifier of the session to delete

**Examples:**

```ruby
# Create and use a session
sandbox.process.create_session("temp-session")
# ... use the session ...

# Clean up when done
sandbox.process.delete_session("temp-session")

```

#### create_pty_session()

```ruby
def create_pty_session(id:, cwd:, envs:, pty_size:)

```

Creates a new PTY (pseudo-terminal) session in the Sandbox.

Creates an interactive terminal session that can execute commands and handle user input.
The PTY session behaves like a real terminal, supporting features like command history.

**Parameters**:

- `id` _String_ - Unique identifier for the PTY session. Must be unique within the Sandbox.
- `cwd` _String, nil_ - Working directory for the PTY session. Defaults to the sandbox's working directory.
- `envs` _Hash\<String, String\>, nil_ - Environment variables to set in the PTY session. These will be merged with
the Sandbox's default environment variables.
- `pty_size` _PtySize, nil_ - Terminal size configuration. Defaults to 80x24 if not specified.

**Returns**:

- `PtyHandle` - Handle for managing the created PTY session. Use this to send input,
receive output, resize the terminal, and manage the session lifecycle.

**Raises**:

- `Daytona:Sdk:Error` - If the PTY session creation fails or the session ID is already in use.

**Examples:**

```ruby
# Create a basic PTY session
pty_handle = sandbox.process.create_pty_session(id: "my-pty")

# Create a PTY session with specific size and environment
pty_size = Daytona::PtySize.new(rows: 30, cols: 120)
pty_handle = sandbox.process.create_pty_session(
  id: "my-pty",
  cwd: "/workspace",
  envs: {"NODE_ENV" => "development"},
  pty_size: pty_size
)

# Use the PTY session
pty_handle.wait_for_connection
pty_handle.send_input("ls -la\n")
result = pty_handle.wait
pty_handle.disconnect

```

#### connect_pty_session()

```ruby
def connect_pty_session(session_id)

```

Connects to an existing PTY session in the Sandbox.

Establishes a WebSocket connection to an existing PTY session, allowing you to
interact with a previously created terminal session.

**Parameters**:

- `session_id` _String_ - Unique identifier of the PTY session to connect to.

**Returns**:

- `PtyHandle` - Handle for managing the connected PTY session.

**Raises**:

- `Daytona:Sdk:Error` - If the PTY session doesn't exist or connection fails.

**Examples:**

```ruby
# Connect to an existing PTY session
pty_handle = sandbox.process.connect_pty_session("my-pty-session")
pty_handle.wait_for_connection
pty_handle.send_input("echo 'Hello World'\n")
result = pty_handle.wait
pty_handle.disconnect

```

#### resize_pty_session()

```ruby
def resize_pty_session(session_id, pty_size)

```

Resizes a PTY session to the specified dimensions

**Parameters**:

- `session_id` _String_ - Unique identifier of the PTY session
- `pty_size` _PtySize_ - New terminal size

**Returns**:

- `DaytonaApiClient:PtySessionInfo` - Updated PTY session information

**Examples:**

```ruby
pty_size = Daytona::PtySize.new(rows: 30, cols: 120)
session_info = sandbox.process.resize_pty_session("my-pty", pty_size)
puts "PTY resized to #{session_info.cols}x#{session_info.rows}"

```

#### delete_pty_session()

```ruby
def delete_pty_session(session_id)

```

Deletes a PTY session, terminating the associated process

**Parameters**:

- `session_id` _String_ - Unique identifier of the PTY session to delete

**Returns**:

- `void`

**Examples:**

```ruby
sandbox.process.delete_pty_session("my-pty")

```

#### list_pty_sessions()

```ruby
def list_pty_sessions()

```

Lists all PTY sessions in the Sandbox

**Returns**:

- `Array\<DaytonaApiClient:PtySessionInfo\>` - List of PTY session information

**Examples:**

```ruby
sessions = sandbox.process.list_pty_sessions
sessions.each do |session|
  puts "PTY Session #{session.id}: #{session.cols}x#{session.rows}"
end

```

#### get_pty_session_info()

```ruby
def get_pty_session_info(session_id)

```

Gets detailed information about a specific PTY session

Retrieves comprehensive information about a PTY session including its current state,
configuration, and metadata.

**Parameters**:

- `session_id` _String_ - Unique identifier of the PTY session to retrieve information for

**Returns**:

- `DaytonaApiClient:PtySessionInfo` - Detailed information about the PTY session including ID, state,
creation time, working directory, environment variables, and more

**Examples:**

```ruby
# Get details about a specific PTY session
session_info = sandbox.process.get_pty_session_info("my-session")
puts "Session ID: #{session_info.id}"
puts "Active: #{session_info.active}"
puts "Working Directory: #{session_info.cwd}"
puts "Terminal Size: #{session_info.cols}x#{session_info.rows}"

```
