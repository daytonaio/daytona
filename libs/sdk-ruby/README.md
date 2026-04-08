# Daytona Ruby SDK

The official Ruby SDK for [Daytona](https://daytona.io), an open-source, secure and elastic infrastructure for running AI-generated code. Daytona provides full composable computers — [sandboxes](https://www.daytona.io/docs/en/sandboxes/) — that you can manage programmatically using the Daytona SDK.

The SDK provides an interface for sandbox management, file system operations, Git operations, language server protocol support, process and code execution, and computer use. For more information, see the [documentation](https://www.daytona.io/docs/en/ruby-sdk/).

## Installation

Install the package using **gem**:

```bash
gem install daytona
```

## Get API key

Generate an API key from the [Daytona Dashboard ↗](https://app.daytona.io/dashboard/keys) to authenticate SDK requests and access Daytona services. For more information, see the [API keys](https://www.daytona.io/docs/en/api-keys/) documentation.

## Configuration

Configure the SDK using [environment variables](https://www.daytona.io/docs/en/configuration/#environment-variables) or by passing a [configuration object](https://www.daytona.io/docs/en/configuration/#configuration-in-code):

- `DAYTONA_API_KEY`: Your Daytona [API key](https://www.daytona.io/docs/en/api-keys/)
- `DAYTONA_API_URL`: The Daytona [API URL](https://www.daytona.io/docs/en/tools/api/)
- `DAYTONA_TARGET`: Your target [region](https://www.daytona.io/docs/en/regions/) environment (e.g. `us`, `eu`)

```ruby
require 'daytona'

# Initialize with environment variables
daytona = Daytona::Daytona.new

# Initialize with configuration object
config = Daytona::Config.new(
  api_key: 'YOUR_API_KEY',
  api_url: 'YOUR_API_URL',
  target: 'us'
)
```

## Create a sandbox

Create a sandbox to run your code securely in an isolated environment.

```ruby
require 'daytona'

config = Daytona::Config.new(api_key: 'YOUR_API_KEY')
daytona = Daytona::Daytona.new(config)
sandbox = daytona.create
```

## Examples and guides

Daytona provides [examples](https://www.daytona.io/docs/en/getting-started/#examples) and [guides](https://www.daytona.io/docs/en/guides/) for common sandbox operations, best practices, and a wide range of topics, from basic usage to advanced topics, showcasing various types of integrations between Daytona and other tools.

### Create a sandbox with custom resources

Create a sandbox with [custom resources](https://www.daytona.io/docs/en/sandboxes/#resources) (CPU, memory, disk).

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create(
    Daytona::CreateSandboxFromImageParams.new(
        image: Daytona::Image.debian_slim('3.12'),
        resources: Daytona::Resources.new(cpu: 2, memory: 4, disk: 8)
    )
)
```

### Create an ephemeral sandbox

Create an [ephemeral sandbox](https://www.daytona.io/docs/en/sandboxes/#ephemeral-sandboxes) that is automatically deleted when stopped.

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(ephemeral: true, auto_stop_interval: 5)
)
```

### Create a sandbox from a snapshot

Create a sandbox from a [snapshot](https://www.daytona.io/docs/en/snapshots/).

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(
        snapshot: 'my-snapshot-name'
    )
)
```

### Execute commands

Execute commands in the sandbox.

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

### File operations

Upload, download, and search files in the sandbox.

```ruby
# Upload a text file from string content
content = "Hello, World!"
sandbox.fs.upload_file(content, "tmp/hello.txt")

# Download and get file content
content = sandbox.fs.download_file("workspace/data/file.txt")
puts content

# Get file metadata
info = sandbox.fs.get_file_info("workspace/data/file.txt")
puts "Size: #{info.size} bytes"
puts "Modified: #{info.mod_time}"
puts "Mode: #{info.mode}"
```

### Git operations

Clone, list branches, and add files to the sandbox.

```ruby
# Basic clone
sandbox.git.clone(
  url: 'https://github.com/daytonaio/daytona.git',
  path: 'workspace/repo'
)

# List branches
response = sandbox.git.branches('workspace/repo')
puts "Branches: #{response.branches}"

# Add files
sandbox.git.add('workspace/repo', ['README.md'])
```

### Language server protocol

Create and start a language server to get code completions, document symbols, and more.

```ruby
# Create a language server
lsp_server = sandbox.create_lsp_server(
  language_id: Daytona::LspServer::Language::PYTHON,
  path_to_project: 'workspace/project'
)
lsp_server.start

# Notify server that a file is open
lsp_server.did_open('workspace/project/main.py')

# Get document symbols
symbols = lsp_server.document_symbols('workspace/project/main.py')

# Get completions
completions = lsp_server.completions(
  path: 'workspace/project/main.py',
  position: Daytona::LspServer::Position.new(line: 10, character: 15)
)
```

## Contributing

Daytona is Open Source under the [Apache License 2.0](https://github.com/daytonaio/daytona/blob/main/libs/sdk-ruby/LICENSE), and is the [copyright of its contributors](https://github.com/daytonaio/daytona/blob/main/NOTICE). If you would like to contribute to the software, read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/). Afterwards, navigate to the [contributing guide](../../CONTRIBUTING.md) to get started.
