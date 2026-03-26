---
title: "LspServer"
hideTitleOnPage: true
---

## LspServer

```python
class LspServer()
```

Provides Language Server Protocol functionality for code intelligence to provide
IDE-like features such as code completion, symbol search, and more.

#### LspServer.\_\_init\_\_

```python
def __init__(language_id: LspLanguageId | LspLanguageIdLiteral,
             path_to_project: str, api_client: LspApi)
```

Initializes a new LSP server instance.

**Arguments**:

- `language_id` _LspLanguageId | LspLanguageIdLiteral_ - The language server type
  (e.g., LspLanguageId.TYPESCRIPT).
- `path_to_project` _str_ - Absolute path to the project root directory.
- `api_client` _LspApi_ - API client for Sandbox operations.

#### LspServer.start

```python
@intercept_errors(message_prefix="Failed to start LSP server: ")
@with_instrumentation()
def start() -> None
```

Starts the language server.

This method must be called before using any other LSP functionality.
It initializes the language server for the specified language and project.

**Example**:

```python
lsp = sandbox.create_lsp_server("typescript", "workspace/project")
lsp.start()  # Initialize the server
# Now ready for LSP operations
```

#### LspServer.stop

```python
@intercept_errors(message_prefix="Failed to stop LSP server: ")
@with_instrumentation()
def stop() -> None
```

Stops the language server.

This method should be called when the LSP server is no longer needed to
free up system resources.

**Example**:

```python
# When done with LSP features
lsp.stop()  # Clean up resources
```

#### LspServer.did\_open

```python
@intercept_errors(message_prefix="Failed to open file: ")
@with_instrumentation()
def did_open(path: str) -> None
```

Notifies the language server that a file has been opened.

This method should be called when a file is opened in the editor to enable
language features like diagnostics and completions for that file. The server
will begin tracking the file's contents and providing language features.

**Arguments**:

- `path` _str_ - Path to the opened file. Relative paths are resolved based on the project path
  set in the LSP server constructor.
  

**Example**:

```python
# When opening a file for editing
lsp.did_open("workspace/project/src/index.ts")
# Now can get completions, symbols, etc. for this file
```

#### LspServer.did\_close

```python
@intercept_errors(message_prefix="Failed to close file: ")
@with_instrumentation()
def did_close(path: str) -> None
```

Notify the language server that a file has been closed.

This method should be called when a file is closed in the editor to allow
the language server to clean up any resources associated with that file.

**Arguments**:

- `path` _str_ - Path to the closed file. Relative paths are resolved based on the project path
  set in the LSP server constructor.
  

**Example**:

```python
# When done editing a file
lsp.did_close("workspace/project/src/index.ts")
```

#### LspServer.document\_symbols

```python
@intercept_errors(message_prefix="Failed to get symbols from document: ")
@with_instrumentation()
def document_symbols(path: str) -> list[LspSymbol]
```

Gets symbol information (functions, classes, variables, etc.) from a document.

**Arguments**:

- `path` _str_ - Path to the file to get symbols from. Relative paths are resolved based on the project path
  set in the LSP server constructor.
  

**Returns**:

- `list[LspSymbol]` - List of symbols in the document. Each symbol includes:
  - name: The symbol's name
  - kind: The symbol's kind (function, class, variable, etc.)
  - location: The location of the symbol in the file
  

**Example**:

```python
# Get all symbols in a file
symbols = lsp.document_symbols("workspace/project/src/index.ts")
for symbol in symbols:
    print(f"{symbol.kind} {symbol.name}: {symbol.location}")
```

#### LspServer.workspace\_symbols

```python
@deprecated(
    reason=
    "Method is deprecated. Use `sandbox_symbols` instead. This method will be removed in a future version."
)
@with_instrumentation()
def workspace_symbols(query: str) -> list[LspSymbol]
```

Searches for symbols matching the query string across all files
in the Sandbox.

**Arguments**:

- `query` _str_ - Search query to match against symbol names.
  

**Returns**:

- `list[LspSymbol]` - List of matching symbols from all files.

#### LspServer.sandbox\_symbols

```python
@intercept_errors(message_prefix="Failed to get symbols from sandbox: ")
@with_instrumentation()
def sandbox_symbols(query: str) -> list[LspSymbol]
```

Searches for symbols matching the query string across all files
in the Sandbox.

**Arguments**:

- `query` _str_ - Search query to match against symbol names.
  

**Returns**:

- `list[LspSymbol]` - List of matching symbols from all files. Each symbol
  includes:
  - name: The symbol's name
  - kind: The symbol's kind (function, class, variable, etc.)
  - location: The location of the symbol in the file
  

**Example**:

```python
# Search for all symbols containing "User"
symbols = lsp.sandbox_symbols("User")
for symbol in symbols:
    print(f"{symbol.name} in {symbol.location}")
```

#### LspServer.completions

```python
@intercept_errors(message_prefix="Failed to get completions: ")
@with_instrumentation()
def completions(path: str, position: LspCompletionPosition) -> CompletionList
```

Gets completion suggestions at a position in a file.

**Arguments**:

- `path` _str_ - Path to the file. Relative paths are resolved based on the project path
  set in the LSP server constructor.
- `position` _LspCompletionPosition_ - Cursor position to get completions for.
  

**Returns**:

- `CompletionList` - List of completion suggestions. The list includes:
  - isIncomplete: Whether more items might be available
  - items: List of completion items, each containing:
  - label: The text to insert
  - kind: The kind of completion
  - detail: Additional details about the item
  - documentation: Documentation for the item
  - sortText: Text used to sort the item in the list
  - filterText: Text used to filter the item
  - insertText: The actual text to insert (if different from label)
  

**Example**:

```python
# Get completions at a specific position
pos = LspCompletionPosition(line=10, character=15)
completions = lsp.completions("workspace/project/src/index.ts", pos)
for item in completions.items:
    print(f"{item.label} ({item.kind}): {item.detail}")
```


## LspLanguageId

```python
class LspLanguageId(str, Enum)
```

Language IDs for Language Server Protocol (LSP).

**Enum Members**:
    - `PYTHON` ("python")
    - `TYPESCRIPT` ("typescript")
    - `JAVASCRIPT` ("javascript")

## LspCompletionPosition

```python
@dataclass
class LspCompletionPosition()
```

Represents a zero-based completion position in a text document,
specified by line number and character offset.

**Attributes**:

- `line` _int_ - Zero-based line number in the document.
- `character` _int_ - Zero-based character offset on the line.

