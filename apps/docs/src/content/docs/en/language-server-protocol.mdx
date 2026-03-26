---
title: Language Server Protocol
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

Daytona provides Language Server Protocol (LSP) support through sandbox instances. This enables advanced language features like code completion, diagnostics, and more.

## Create LSP servers

Daytona provides methods to create LSP servers. The `path_to_project` argument is relative to the current sandbox working directory when no leading `/` is used. The working directory is specified by WORKDIR when it is present in the Dockerfile, and otherwise falls back to the user's home directory.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona, LspLanguageId

# Create Sandbox
daytona = Daytona()
sandbox = daytona.create()

# Create LSP server for Python
lsp_server = sandbox.create_lsp_server(
    language_id=LspLanguageId.PYTHON,
    path_to_project="workspace/project"
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona, LspLanguageId } from '@daytonaio/sdk'

// Create sandbox
const daytona = new Daytona()
const sandbox = await daytona.create({
  language: 'typescript',
})

// Create LSP server for TypeScript
const lspServer = await sandbox.createLspServer(
  LspLanguageId.TYPESCRIPT,
  'workspace/project'
)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

# Create Sandbox
daytona = Daytona::Daytona.new
sandbox = daytona.create

# Create LSP server for Python
lsp_server = sandbox.create_lsp_server(
  language_id: Daytona::LspServer::Language::PYTHON,
  path_to_project: 'workspace/project'
)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Create sandbox
client, err := daytona.NewClient()
if err != nil {
	log.Fatal(err)
}

ctx := context.Background()
sandbox, err := client.Create(ctx, nil)
if err != nil {
	log.Fatal(err)
}

// Get LSP service for Python
lsp := sandbox.Lsp(types.LspLanguagePython, "workspace/project")
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), and [Go SDK](/docs/en/go-sdk/) references:

> [**create_lsp_server (Python SDK)**](/docs/en/python-sdk/sync/sandbox#sandboxcreate_lsp_server)
>
> [**createLspServer (TypeScript SDK)**](/docs/en/typescript-sdk/sandbox#createlspserver)
>
> [**create_lsp_server (Ruby SDK)**](/docs/en/ruby-sdk/sandbox#create_lsp_server)

### Supported languages

The supported languages for creating LSP servers with Daytona are defined by the `LspLanguageId` enum:

| Enum Value                     | Description                            |
| ------------------------------ | -------------------------------------- |
| **`LspLanguageId.PYTHON`**     | Python language server                 |
| **`LspLanguageId.TYPESCRIPT`** | TypeScript/JavaScript language server  |

For more information, see the [Python SDK](/docs/en/python-sdk/sync/lsp-server/#lsplanguageid) and [TypeScript SDK](/docs/en/typescript-sdk/lsp-server/#lsplanguageid) references:

> [**LspLanguageId (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lsplanguageid)
>
> [**LspLanguageId (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#lsplanguageid)

## Start LSP servers

Daytona provides methods to start LSP servers.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
lsp = sandbox.create_lsp_server("typescript", "workspace/project")
lsp.start()  # Initialize the server
# Now ready for LSP operations
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const lsp = await sandbox.createLspServer('typescript', 'workspace/project')
await lsp.start() // Initialize the server
// Now ready for LSP operations
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
lsp = sandbox.create_lsp_server(
  language_id: Daytona::LspServer::Language::PYTHON,
  path_to_project: 'workspace/project'
)
lsp.start  # Initialize the server
# Now ready for LSP operations
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
lsp := sandbox.Lsp(types.LspLanguagePython, "workspace/project")
err := lsp.Start(ctx)  // Initialize the server
if err != nil {
	log.Fatal(err)
}
// Now ready for LSP operations
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/start' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "languageId": "",
  "pathToProject": ""
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp/) references:

> [**start (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspserverstart)
>
> [**start (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#start)
>
> [**start (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#start)
>
> [**Start (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.Start)
>
> [**start (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/POST/lsp/start)

## Stop LSP servers

Daytona provides methods to stop LSP servers.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# When done with LSP features
lsp.stop()  # Clean up resources
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// When done with LSP features
await lsp.stop() // Clean up resources
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# When done with LSP features
lsp.stop  # Clean up resources
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// When done with LSP features
err := lsp.Stop(ctx)  // Clean up resources
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/stop' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "languageId": "",
  "pathToProject": ""
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp/) references:

> [**stop (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspserverstop)
>
> [**stop (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#stop)
>
> [**stop (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#stop)
>
> [**Stop (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.Stop)
>
> [**stop (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/POST/lsp/stop)

## Code completions

Daytona provides methods to get code completions for a specific position in a file.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
completions = lsp_server.completions(
    path="workspace/project/main.py",
    position={"line": 10, "character": 15}
)
print(f"Completions: {completions}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const completions = await lspServer.completions('workspace/project/main.ts', {
  line: 10,
  character: 15,
})
console.log('Completions:', completions)
```
</TabItem>

<TabItem label="Ruby" icon="seti:ruby">

```ruby
completions = lsp_server.completions(
  path: 'workspace/project/main.py',
  position: { line: 10, character: 15 }
)
puts "Completions: #{completions}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
completions, err := lsp.Completions(ctx, "workspace/project/main.py",
	types.Position{Line: 10, Character: 15},
)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Completions: %v\n", completions)
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/completions' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "context": {
    "triggerCharacter": "",
    "triggerKind": 1
  },
  "languageId": "",
  "pathToProject": "",
  "position": {
    "character": 1,
    "line": 1
  },
  "uri": ""
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp/) references:

> [**completions (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspservercompletions)
>
> [**completions (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#completions)
>
> [**completions (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#completions)
>
> [**Completions (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.Completions)
>
> [**completions (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/POST/lsp/completions)

## File notifications

Daytona provides methods to notify the LSP server when files are opened or closed. This enables features like diagnostics and completion tracking for the specified files.

### Open file

Notifies the language server that a file has been opened for editing.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Notify server that a file is open
lsp_server.did_open("workspace/project/main.py")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Notify server that a file is open
await lspServer.didOpen('workspace/project/main.ts')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Notify server that a file is open
lsp_server.did_open('workspace/project/main.py')
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Notify server that a file is open
err := lsp.DidOpen(ctx, "workspace/project/main.py")
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/did-open' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "languageId": "",
  "pathToProject": "",
  "uri": ""
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp/) references:

> [**did_open (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspserverdid_open)
>
> [**didOpen (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#didopen)
>
> [**did_open (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#did_open)
>
> [**DidOpen (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.DidOpen)
>
> [**did_open (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/POST/lsp/did-open)

### Close file

Notifies the language server that a file has been closed. This allows the server to clean up resources associated with that file.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Notify server that a file is closed
lsp_server.did_close("workspace/project/main.py")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Notify server that a file is closed
await lspServer.didClose('workspace/project/main.ts')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Notify server that a file is closed
lsp_server.did_close('workspace/project/main.py')
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Notify server that a file is closed
err := lsp.DidClose(ctx, "workspace/project/main.py")
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/did-close' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "languageId": "",
  "pathToProject": "",
  "uri": ""
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp/) references:

> [**did_close (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspserverdid_close)
>
> [**didClose (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#didclose)
>
> [**did_close (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#did_close)
>
> [**DidClose (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.DidClose)
>
> [**did_close (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/POST/lsp/did-close)

## Document symbols

Daytona provides methods to retrieve symbols (functions, classes, variables, etc.) from a document.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
symbols = lsp_server.document_symbols("workspace/project/main.py")
for symbol in symbols:
    print(f"Symbol: {symbol.name}, Kind: {symbol.kind}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const symbols = await lspServer.documentSymbols('workspace/project/main.ts')
symbols.forEach((symbol) => {
  console.log(`Symbol: ${symbol.name}, Kind: ${symbol.kind}`)
})
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
symbols = lsp_server.document_symbols('workspace/project/main.py')
symbols.each do |symbol|
  puts "Symbol: #{symbol.name}, Kind: #{symbol.kind}"
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
symbols, err := lsp.DocumentSymbols(ctx, "workspace/project/main.py")
if err != nil {
	log.Fatal(err)
}
for _, symbol := range symbols {
	fmt.Printf("Symbol: %v\n", symbol)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/document-symbols?languageId=&pathToProject=&uri='
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp/) references:

> [**document_symbols (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspserverdocument_symbols)
>
> [**documentSymbols (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#documentsymbols)
>
> [**document_symbols (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#document_symbols)
>
> [**DocumentSymbols (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.DocumentSymbols)
>
> [**document_symbols (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/GET/lsp/document-symbols)

## Sandbox symbols

Daytona provides methods to search for symbols across all files in the sandbox.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
symbols = lsp_server.sandbox_symbols("MyClass")
for symbol in symbols:
    print(f"Found: {symbol.name} at {symbol.location}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const symbols = await lspServer.sandboxSymbols('MyClass')
symbols.forEach((symbol) => {
  console.log(`Found: ${symbol.name} at ${symbol.location}`)
})
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
symbols = lsp_server.sandbox_symbols('MyClass')
symbols.each do |symbol|
  puts "Found: #{symbol.name} at #{symbol.location}"
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
symbols, err := lsp.SandboxSymbols(ctx, "MyClass")
if err != nil {
	log.Fatal(err)
}
for _, symbol := range symbols {
	fmt.Printf("Found: %v\n", symbol)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/lsp/workspacesymbols?query=&languageId=&pathToProject='
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/lsp-server/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox/tag/lsp) references:

> [**sandbox_symbols (Python SDK)**](/docs/en/python-sdk/sync/lsp-server/#lspserversandbox_symbols)
>
> [**sandboxSymbols (TypeScript SDK)**](/docs/en/typescript-sdk/lsp-server/#sandboxsymbols)
>
> [**sandbox_symbols (Ruby SDK)**](/docs/en/ruby-sdk/lsp-server/#sandbox_symbols)
>
> [**SandboxSymbols (Go SDK)**](/docs/en/go-sdk/daytona#LspServerService.SandboxSymbols)
>
> [**sandbox_symbols (API)**](/docs/en/tools/api/#daytona-toolbox/tag/lsp/GET/lsp/workspacesymbols)
