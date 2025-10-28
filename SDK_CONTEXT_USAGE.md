# Python SDK - Multi-Context Usage Guide

## Overview

The updated Python SDK now supports managing multiple isolated interpreter contexts. Each context has its own Python process with independent state and working directory.

## Installation

```bash
pip install daytona-sdk
```

## Basic Usage

### 1. Default Context (Backwards Compatible)

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create(...)

# Get code interpreter
interpreter = sandbox.code_interpreter()

# Execute code (uses default context automatically)
interpreter.execute(
    code="x = 42\nprint(x)",
    on_stdout=lambda text: print(f"Output: {text}"),
    on_error=lambda name, value, tb: print(f"Error: {name}: {value}")
)
```

## Context Management

### Create a Context

```python
# Create context with default settings
context = interpreter.create_context()
print(f"Context ID: {context['id']}")
print(f"Working Directory: {context['cwd']}")
print(f"Language: {context['language']}")
print(f"Active: {context['active']}")

# Create context with custom working directory
context = interpreter.create_context(cwd="/workspace/project")

# Create context with specific language (only "python" supported currently)
context = interpreter.create_context(language="python")
```

**Response Example:**
```python
{
    'id': '550e8400-e29b-41d4-a716-446655440000',
    'cwd': '/workspace/project',
    'language': 'python',
    'createdAt': '2025-01-28T10:30:00Z',
    'active': True
}
```

### Execute in Specific Context

```python
# Create context
ctx = interpreter.create_context(cwd="/my/project")

# Execute code in that context
interpreter.execute(
    code="import os\nprint(os.getcwd())",
    context_id=ctx["id"],
    on_stdout=lambda text: print(text)
)
# Output: /my/project
```

### List Contexts

```python
# List all user-created contexts (excludes default)
contexts = interpreter.list_contexts()

for ctx in contexts:
    print(f"ID: {ctx['id']}")
    print(f"CWD: {ctx['cwd']}")
    print(f"Active: {ctx['active']}")
    print("---")
```

### Delete Context

```python
# Create context
ctx = interpreter.create_context()

# ... use it ...

# Delete when done
interpreter.delete_context(ctx["id"])
```

## Complete Examples

### Example 1: Project-Specific Context

```python
from daytona import Daytona

def process_project(sandbox, project_path):
    interpreter = sandbox.code_interpreter()
    
    # Create context for this project
    ctx = interpreter.create_context(cwd=project_path)
    
    try:
        # Install dependencies
        interpreter.execute(
            code="""
import subprocess
subprocess.run(['pip', 'install', '-r', 'requirements.txt'])
            """,
            context_id=ctx["id"],
            timeout=300
        )
        
        # Run tests
        interpreter.execute(
            code="import pytest; pytest.main(['tests/'])",
            context_id=ctx["id"],
            on_stdout=lambda text: print(f"Test: {text}")
        )
        
    finally:
        # Clean up
        interpreter.delete_context(ctx["id"])

# Use it
daytona = Daytona()
sandbox = daytona.create(...)
process_project(sandbox, "/workspace/my-project")
```

### Example 2: Multiple Isolated Environments

```python
from daytona import Daytona

def compare_environments(sandbox):
    interpreter = sandbox.code_interpreter()
    
    # Create two contexts
    ctx1 = interpreter.create_context(cwd="/workspace/env1")
    ctx2 = interpreter.create_context(cwd="/workspace/env2")
    
    # Set different variables in each
    interpreter.execute(
        code="config = {'version': '1.0'}",
        context_id=ctx1["id"]
    )
    
    interpreter.execute(
        code="config = {'version': '2.0'}",
        context_id=ctx2["id"]
    )
    
    # Verify isolation
    def print_version(text):
        print(f"Version: {text}")
    
    interpreter.execute(
        code="print(config['version'])",
        context_id=ctx1["id"],
        on_stdout=print_version  # Output: 1.0
    )
    
    interpreter.execute(
        code="print(config['version'])",
        context_id=ctx2["id"],
        on_stdout=print_version  # Output: 2.0
    )
    
    # Clean up
    interpreter.delete_context(ctx1["id"])
    interpreter.delete_context(ctx2["id"])
```

### Example 3: Long-Running Context with State

```python
from daytona import Daytona

def data_processing_pipeline(sandbox):
    interpreter = sandbox.code_interpreter()
    
    # Create context for data processing
    ctx = interpreter.create_context(cwd="/workspace/data")
    
    try:
        # Load data (persists in context)
        interpreter.execute(
            code="""
import pandas as pd
df = pd.read_csv('data.csv')
print(f'Loaded {len(df)} rows')
            """,
            context_id=ctx["id"]
        )
        
        # Process data (uses loaded df)
        interpreter.execute(
            code="""
df['processed'] = df['value'] * 2
print(f'Processed {len(df)} rows')
            """,
            context_id=ctx["id"]
        )
        
        # Save results (still has df)
        interpreter.execute(
            code="""
df.to_csv('output.csv', index=False)
print('Saved results')
            """,
            context_id=ctx["id"]
        )
        
    finally:
        interpreter.delete_context(ctx["id"])
```

### Example 4: Context Auto-Restart After Exit

```python
from daytona import Daytona

def demonstrate_auto_restart(sandbox):
    interpreter = sandbox.code_interpreter()
    
    # Create context
    ctx = interpreter.create_context()
    
    # Set a variable
    interpreter.execute(
        code="x = 100\nprint(f'x = {x}')",
        context_id=ctx["id"],
        on_stdout=print  # Output: x = 100
    )
    
    # Exit the context
    interpreter.execute(
        code="exit()",
        context_id=ctx["id"]
    )
    # Context is now inactive
    
    # Execute again - context auto-restarts
    interpreter.execute(
        code="print(f'x = {x}')",  # Variable no longer exists (fresh environment)
        context_id=ctx["id"],
        on_error=lambda name, val, tb: print(f"Error: {name}")  # Output: Error: NameError
    )
    
    # Context is active again, but state was reset
    interpreter.delete_context(ctx["id"])
```

### Example 5: Context Management

```python
from daytona import Daytona

def manage_contexts(sandbox):
    interpreter = sandbox.code_interpreter()
    
    # Create multiple contexts
    contexts = []
    for i in range(3):
        ctx = interpreter.create_context(cwd=f"/workspace/project{i}")
        contexts.append(ctx)
        print(f"Created context: {ctx['id']}")
    
    # List all contexts
    all_contexts = interpreter.list_contexts()
    print(f"\nTotal contexts: {len(all_contexts)}")
    
    # Execute in each context
    for ctx in contexts:
        interpreter.execute(
            code=f"print('Hello from {ctx['id']}')",
            context_id=ctx["id"],
            on_stdout=print
        )
    
    # Clean up all
    for ctx in contexts:
        interpreter.delete_context(ctx["id"])
        print(f"Deleted context: {ctx['id']}")
```

## Error Handling

### Context Not Found

```python
try:
    interpreter.execute(
        code="print('test')",
        context_id="nonexistent-id"
    )
except Exception as e:
    print(f"Error: {e}")
    # WebSocket sends ContextError
```

### Context Creation Failed

```python
try:
    ctx = interpreter.create_context(language="javascript")
except requests.HTTPError as e:
    print(f"Failed to create context: {e}")
    # Error: Unsupported language
```

### Timeout Handling

```python
try:
    interpreter.execute(
        code="import time; time.sleep(100)",
        context_id=ctx["id"],
        timeout=5  # 5 second timeout
    )
except TimeoutError as e:
    print(f"Execution timed out: {e}")
    # WebSocket closed with code 4008
```

### Cannot Delete Default Context

```python
try:
    interpreter.delete_context("default")
except requests.HTTPError as e:
    print(f"Cannot delete: {e}")
    # Error: Cannot delete default context
```

## Best Practices

### 1. Always Clean Up Contexts

```python
# Good: Use try/finally
ctx = interpreter.create_context()
try:
    # ... use context ...
    pass
finally:
    interpreter.delete_context(ctx["id"])

# Even better: Use context manager (if available)
# with interpreter.context(cwd="/path") as ctx:
#     interpreter.execute(code, context_id=ctx["id"])
```

### 2. Reuse Contexts When Possible

```python
# Good: Create once, use multiple times
ctx = interpreter.create_context()
for script in scripts:
    interpreter.execute(script, context_id=ctx["id"])
interpreter.delete_context(ctx["id"])

# Bad: Create context per execution
for script in scripts:
    ctx = interpreter.create_context()
    interpreter.execute(script, context_id=ctx["id"])
    interpreter.delete_context(ctx["id"])
```

### 3. Use Default Context for Simple Cases

```python
# Good: Simple, one-off executions
interpreter.execute("print('Hello')")

# Overkill: Creating context for single execution
ctx = interpreter.create_context()
interpreter.execute("print('Hello')", context_id=ctx["id"])
interpreter.delete_context(ctx["id"])
```

### 4. Set Timeouts for Long-Running Code

```python
# Good: Prevent runaway processes
interpreter.execute(
    code=long_running_code,
    context_id=ctx["id"],
    timeout=300  # 5 minutes max
)
```

### 5. Handle Context Errors Gracefully

```python
try:
    interpreter.execute(code, context_id=ctx_id)
except requests.HTTPError as e:
    if e.response.status_code == 404:
        # Context doesn't exist - create it
        ctx = interpreter.create_context()
        interpreter.execute(code, context_id=ctx["id"])
    else:
        raise
```

## API Reference

### `create_context(cwd=None, language=None) -> Dict`

Creates a new isolated interpreter context.

**Parameters:**
- `cwd` (str, optional): Working directory for the context
- `language` (str, optional): Language ("python" only, default: "python")

**Returns:** Dict with `id`, `cwd`, `language`, `createdAt`, `active`

**Raises:** `requests.HTTPError` if creation fails

### `list_contexts() -> List[Dict]`

Lists all user-created contexts (excludes default).

**Returns:** List of context dictionaries

**Raises:** `requests.HTTPError` if request fails

### `delete_context(context_id: str) -> None`

Deletes a context and shuts down its worker process.

**Parameters:**
- `context_id` (str): ID of context to delete

**Raises:** `requests.HTTPError` if deletion fails or context not found

### `execute(code, context_id=None, ...)`

Executes code in a context (default context if not specified).

**Parameters:**
- `code` (str): Python code to execute
- `context_id` (str, optional): Context ID (defaults to "default")
- `envs` (Dict, optional): Environment variables
- `timeout` (int, optional): Timeout in seconds
- `on_stdout`, `on_stderr`, `on_error`, `on_artifact`, `on_control`: Callbacks

**Raises:** `TimeoutError`, `ConnectionError`, or other exceptions

## Migration from Single Context

### Before (Old SDK)
```python
# No context management
interpreter.execute("print('Hello')")
```

### After (New SDK - Backwards Compatible)
```python
# Option 1: Same as before (uses default context)
interpreter.execute("print('Hello')")

# Option 2: Use custom context
ctx = interpreter.create_context()
interpreter.execute("print('Hello')", context_id=ctx["id"])
interpreter.delete_context(ctx["id"])
```

**No breaking changes** - existing code continues to work!

## Summary

✅ **Create isolated contexts** with `create_context()`
✅ **Execute in specific contexts** with `context_id` parameter
✅ **List all contexts** with `list_contexts()`
✅ **Delete contexts** with `delete_context()`
✅ **Backwards compatible** - default context used when no contextId specified
✅ **Auto-restart** - contexts automatically restart after `exit()`
✅ **State isolation** - variables not shared between contexts
✅ **Clean error handling** - proper exceptions for all error cases

