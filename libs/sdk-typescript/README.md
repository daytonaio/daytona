# Daytona TypeScript SDK

The official TypeScript SDK for [Daytona](https://daytona.io), an open-source, secure and elastic infrastructure for running AI-generated code. Daytona provides full composable computers — [sandboxes](https://www.daytona.io/docs/en/sandboxes/) — that you can manage programmatically using the Daytona SDK.

The SDK provides an interface for sandbox management, file system operations, Git operations, language server protocol support, process and code execution, and computer use. For more information, see the [documentation](https://www.daytona.io/docs/en/typescript-sdk/).

## Installation

Install the package using **npm**:

```bash
npm install @daytonaio/sdk
```

or using **yarn**:

```bash
yarn add @daytonaio/sdk
```

## Get API key

Generate an API key from the [Daytona Dashboard ↗](https://app.daytona.io/dashboard/keys) to authenticate SDK requests and access Daytona services. For more information, see the [API keys](https://www.daytona.io/docs/en/api-keys/) documentation.

## Configuration

Configure the SDK using [environment variables](https://www.daytona.io/docs/en/configuration/#environment-variables) or by passing a [configuration object](https://www.daytona.io/docs/en/configuration/#configuration-in-code):

- `DAYTONA_API_KEY`: Your Daytona [API key](https://www.daytona.io/docs/en/api-keys/)
- `DAYTONA_API_URL`: The Daytona [API URL](https://www.daytona.io/docs/en/tools/api/)
- `DAYTONA_TARGET`: Your target [region](https://www.daytona.io/docs/en/regions/) environment (e.g. `us`, `eu`)

```typescript
import { Daytona } from '@daytonaio/sdk'

// Initialize with environment variables
const daytona = new Daytona();

// Initialize with configuration object
const daytona = new Daytona({
  apiKey: 'YOUR_API_KEY',
  apiUrl: 'YOUR_API_URL',
  target: 'us',
});
```

## Create a sandbox

Create a sandbox to run your code securely in an isolated environment.

```typescript
import { Daytona } from '@daytonaio/sdk'

const daytona = new Daytona({apiKey: "YOUR_API_KEY"});
const sandbox = await daytona.create({
  language: 'typescript'
});
const response = await sandbox.process.codeRun('console.log("Hello World!")');
console.log(response.result);
```

## Examples and guides

Daytona provides [examples](https://www.daytona.io/docs/en/getting-started/#examples) and [guides](https://www.daytona.io/docs/en/guides/) for common sandbox operations, best practices, and a wide range of topics, from basic usage to advanced topics, showcasing various types of integrations between Daytona and other tools.

### Create a sandbox with custom resources

Create a sandbox with [custom resources](https://www.daytona.io/docs/en/sandboxes/#resources) (CPU, memory, disk).

```typescript
import { Daytona, Image } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    image: Image.debianSlim('3.12'),
    resources: { cpu: 2, memory: 4, disk: 8 }
});
```

### Create an ephemeral sandbox

Create an [ephemeral sandbox](https://www.daytona.io/docs/en/sandboxes/#ephemeral-sandboxes) that is automatically deleted when stopped.

```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    ephemeral: true,
    autoStopInterval: 5
});
```

### Create a sandbox from a snapshot

Create a sandbox from a [snapshot](https://www.daytona.io/docs/en/snapshots/).

```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    snapshot: 'my-snapshot-name',
    language: 'typescript'
});
```

### Execute commands

Execute commands in the sandbox.

```typescript
// Execute a shell command
const response = await sandbox.process.executeCommand('echo "Hello, World!"')
console.log(response.result)

// Run TypeScript code
const response = await sandbox.process.codeRun(`
const x = 10
const y = 20
console.log(\`Sum: \${x + y}\`)
`)
console.log(response.result)
```

### File operations

Upload, download, and search files in the sandbox.

```typescript
// Upload a file
await sandbox.fs.uploadFile(Buffer.from('Hello, World!'), 'path/to/file.txt')

// Download a file
const content = await sandbox.fs.downloadFile('path/to/file.txt')

// Search for files
const matches = await sandbox.fs.findFiles(root_dir, 'search_pattern')
```

### Git operations

Clone, list branches, and add files to the sandbox.

```typescript
// Clone a repository
await sandbox.git.clone('https://github.com/example/repo', 'path/to/clone')

// List branches
const branches = await sandbox.git.branches('path/to/repo')

// Add files
await sandbox.git.add('path/to/repo', ['file1.txt', 'file2.txt'])
```

### Language server protocol

Create and start a language server to get code completions, document symbols, and more.

```typescript
// Create and start a language server
const lsp = await sandbox.createLspServer('typescript', 'path/to/project')
await lsp.start()

// Notify the lsp for the file
await lsp.didOpen('path/to/file.ts')

// Get document symbols
const symbols = await lsp.documentSymbols('path/to/file.ts')

// Get completions
const completions = await lsp.completions('path/to/file.ts', {
  line: 10,
  character: 15,
})
```

## Contributing

Daytona is Open Source under the [Apache License 2.0](./LICENSE), and is the [copyright of its contributors](../../NOTICE). If you would like to contribute to the software, read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/). Afterwards, navigate to the [contributing guide](../../CONTRIBUTING.md) to get started.
