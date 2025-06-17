# Daytona SDK for Python

A Python SDK for interacting with the Daytona API, providing a simple interface for Daytona Sandbox management, Git operations, file system operations, and language server protocol support.

## Installation

You can install the package using pip:

```bash
pip install daytona
```

## Quick Start

Here's a simple example of using the SDK:

```python
from daytona import Daytona

# Initialize using environment variables
daytona = Daytona()

# Create a sandbox
sandbox = daytona.create()

# Run code in the sandbox
response = sandbox.process.code_run('print("Hello World!")')
print(response.result)

# Clean up when done
daytona.delete(sandbox)
```

## Configuration

The SDK can be configured using environment variables or by passing a configuration object:

```python
from daytona import Daytona, DaytonaConfig

# Initialize with configuration
config = DaytonaConfig(
    api_key="your-api-key",
    api_url="your-api-url",
    target="us"
)
daytona = Daytona(config)
```

Or using environment variables:

- `DAYTONA_API_KEY`: Your Daytona API key
- `DAYTONA_API_URL`: The Daytona API URL
- `DAYTONA_TARGET`: Your target environment

You can also customize sandbox creation:

```python
sandbox = daytona.create(CreateSandboxFromSnapshotParams(
    language="python",
    env_vars={"PYTHON_ENV": "development"},
    auto_stop_interval=60,  # Auto-stop after 1 hour of inactivity
    auto_archive_interval=60  # Auto-archive after a Sandbox has been stopped for 1 hour
))
```

## Features

- **Sandbox Management**: Create, manage and remove sandboxes
- **Git Operations**: Clone repositories, manage branches, and more
- **File System Operations**: Upload, download, search and manipulate files
- **Language Server Protocol**: Interact with language servers for code intelligence
- **Process Management**: Execute code and commands in sandboxes

## Examples

### Execute Commands

```python
# Execute a shell command
response = sandbox.process.exec('echo "Hello, World!"')
print(response.result)

# Run Python code
response = sandbox.process.code_run('''
x = 10
y = 20
print(f"Sum: {x + y}")
''')
print(response.result)
```

### File Operations

```python
# Upload a file
sandbox.fs.upload_file(b'Hello, World!', 'path/to/file.txt')

# Download a file
content = sandbox.fs.download_file('path/to/file.txt')

# Search for files
matches = sandbox.fs.find_files(root_dir, 'search_pattern')
```

### Git Operations

```python
# Clone a repository
sandbox.git.clone('https://github.com/example/repo', 'path/to/clone')

# List branches
branches = sandbox.git.branches('path/to/repo')

# Add files
sandbox.git.add('path/to/repo', ['file1.txt', 'file2.txt'])
```

### Language Server Protocol

```python
# Create and start a language server
lsp = sandbox.create_lsp_server('typescript', 'path/to/project')
lsp.start()

# Notify the lsp for the file
lsp.did_open('path/to/file.ts')

# Get document symbols
symbols = lsp.document_symbols('path/to/file.ts')

# Get completions
completions = lsp.completions('path/to/file.ts', {"line": 10, "character": 15})
```

## Contributing

Daytona is Open Source under the [Apache License 2.0](/libs/sdk-python/LICENSE), and is the [copyright of its contributors](/NOTICE). If you would like to contribute to the software, read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/). Afterwards, navigate to the [contributing guide](/CONTRIBUTING.md) to get started.

Code in [\_sync](/libs/sdk-python/src/daytona/_sync/) directory shouldn't be edited directly. It should be generated from the corresponding async code in the [\_async](/libs/sdk-python/src/daytona/_async/) directory using the [sync_generator.py](/libs/sdk-python/scripts/sync_generator.py) script.
