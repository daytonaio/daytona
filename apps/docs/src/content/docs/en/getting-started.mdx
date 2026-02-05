---
title: Getting Started
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

This section introduces core concepts, common workflows, and next steps for using Daytona.

## Dashboard

[Daytona Dashboard ↗](https://app.daytona.io/) is a visual user interface where you can manage sandboxes, access API keys, view usage, and more.
It serves as the primary point of control for managing your Daytona resources.

## SDKs

Daytona provides [Python](/docs/en/python-sdk), [TypeScript](/docs/en/typescript-sdk), [Ruby](/docs/en/ruby-sdk), and [Go](/docs/en/go-sdk) SDKs to programmatically interact with sandboxes. They support sandbox lifecycle management, code execution, resource access, and more.

## CLI

Daytona provides command-line access to core features for interacting with Daytona Sandboxes, including managing their lifecycle, snapshots, and more.

To interact with Daytona Sandboxes from the command line, install the Daytona CLI:

<Tabs syncKey="language">
<TabItem label="Mac/Linux">

  ```bash
  brew install daytonaio/cli/daytona
  ```

</TabItem>
<TabItem label="Windows">

  ```bash
  powershell -Command "irm https://get.daytona.io/windows | iex"
  ```

</TabItem>
</Tabs>

After installing the Daytona CLI, use the `daytona` command to interact with Daytona Sandboxes from the command line.

To upgrade the Daytona CLI to the latest version:

<Tabs syncKey="language">
<TabItem label="Mac/Linux">

```bash
brew upgrade daytonaio/cli/daytona
```

</TabItem>
<TabItem label="Windows">

```bash
powershell -Command "irm https://get.daytona.io/windows | iex"
```

</TabItem>
</Tabs>

To view all available commands and flags, see the [CLI reference](/docs/en/tools/cli).

## API

Daytona provides a RESTful API for interacting with Daytona Sandboxes, including managing their lifecycle, snapshots, and more.
It serves as a flexible and powerful way to interact with Daytona from your own applications.

To interact with Daytona Sandboxes from the API, see the [API reference](/docs/en/tools/api).

## MCP server

Daytona provides a Model Context Protocol (MCP) server that enables AI agents to interact with Daytona Sandboxes programmatically. The MCP server integrates with popular AI agents including Claude, Cursor, and Windsurf.

To set up the MCP server with your AI agent:

```bash
daytona mcp init [claude/cursor/windsurf]
```

For more information, see the [MCP server documentation](/docs/en/mcp).

## Multiple runtime support

Daytona supports multiple programming language runtimes for direct code execution inside the sandbox.

[TypeScript SDK](/docs/en/typescript-sdk) works across multiple **JavaScript runtimes** including **Node.js**, **browsers**, and **serverless platforms**: Cloudflare Workers, AWS Lambda, Azure Functions, etc.

Using the Daytona SDK in browser-based environments or frameworks like [**Vite**](/docs/en/getting-started#daytona-in-vite-projects) and [**Next.js**](/docs/en/getting-started#daytona-in-nextjs-projects) requires configuring node polyfills.

### Daytona in Vite projects

When using Daytona SDK in a Vite-based project, configure node polyfills to ensure compatibility.

Add the following configuration to your `vite.config.ts` file in the `plugins` array:

```typescript
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  plugins: [
    // ... other plugins
    nodePolyfills({
      globals: { global: true, process: true, Buffer: true },
      overrides: {
        path: 'path-browserify-win32',
      },
    }),
  ],
  // ... rest of your config
})
```

### Daytona in Next.js projects

When using Daytona SDK in a Next.js project, configure node polyfills to ensure compatibility with Webpack and Turbopack bundlers.

Add the following configuration to your `next.config.ts` file:

```typescript
import type { NextConfig } from 'next'
import NodePolyfillPlugin from 'node-polyfill-webpack-plugin'
import { env, nodeless } from 'unenv'

const { alias: turbopackAlias } = env(nodeless, {})

const nextConfig: NextConfig = {
  // Turbopack
  experimental: {
    turbo: {
      resolveAlias: {
        ...turbopackAlias,
      },
    },
  },
  // Webpack
  webpack: (config, { isServer }) => {
    if (!isServer) {
      config.plugins.push(new NodePolyfillPlugin())
    }
    return config
  },
}

export default nextConfig
```

## Guides

Daytona provides a comprehensive set of guides to help you get started. The guides cover a wide range of topics, from basic usage to advanced topics, and showcase various types of integrations between Daytona and other tools.

For more information, see [guides](/docs/en/guides).

## Examples

Daytona provides quick examples for common sandbox operations and best practices. <br />
The examples are based on the Daytona [Python SDK](/docs/en/python-sdk/sync/process), [TypeScript SDK](/docs/en/typescript-sdk/process), [Go SDK](/docs/en/go-sdk/daytona#type-processservice), [Ruby SDK](/docs/en/ruby-sdk/process), [CLI](/docs/en/tools/cli), and [API](/docs/en/tools/api) references. More examples are available in the [GitHub repository ↗](https://github.com/daytonaio/daytona/tree/main/examples).

### Create a sandbox

Create a [sandbox](/docs/en/sandboxes) with default settings.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create()
print(f"Sandbox ID: {sandbox.id}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create();
console.log(`Sandbox ID: ${sandbox.id}`);
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    sandbox, err := client.Create(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Sandbox ID: %s\n", sandbox.ID)
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create
puts "Sandbox ID: #{sandbox.id}"
```

</TabItem>
<TabItem label="CLI" icon="seti:shell">

```shell
daytona create
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{}'
```

</TabItem>
</Tabs>

### Create and run code in a sandbox

Create a [sandbox](/docs/en/sandboxes) and run code securely in it.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create()
response = sandbox.process.exec("echo 'Hello, World!'")
print(response.result)
sandbox.delete()
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create();
const response = await sandbox.process.executeCommand('echo "Hello, World!"');
console.log(response.result);
await sandbox.delete();
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    sandbox, err := client.Create(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    response, err := sandbox.Process.ExecuteCommand(context.Background(), "echo 'Hello, World!'")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response.Result)
    sandbox.Delete(context.Background())
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create
response = sandbox.process.exec(command: "echo 'Hello, World!'")
puts response.result
daytona.delete(sandbox)
```

</TabItem>
<TabItem label="CLI" icon="seti:shell">

```shell
daytona create --name my-sandbox
daytona exec my-sandbox -- echo 'Hello, World!'
daytona delete my-sandbox
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
# Create a sandbox
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{}'

# Execute a command in the sandbox
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/process/execute' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "command": "echo '\''Hello, World!'\''"
}'

# Delete the sandbox
curl 'https://app.daytona.io/api/sandbox/{sandboxId}' \
  --request DELETE \
  --header 'Authorization: Bearer <API_KEY>'
```

</TabItem>
</Tabs>

### Create a sandbox with custom resources

Create a sandbox with [custom resources](/docs/en/sandboxes#resources) (CPU, memory, disk).

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona, CreateSandboxFromImageParams, Image, Resources

daytona = Daytona()
sandbox = daytona.create(
    CreateSandboxFromImageParams(
        image=Image.debian_slim("3.12"),
        resources=Resources(cpu=2, memory=4, disk=8)
    )
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona, Image } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    image: Image.debianSlim('3.12'),
    resources: { cpu: 2, memory: 4, disk: 8 }
});
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    sandbox, err := client.Create(context.Background(), types.ImageParams{
        Image: daytona.DebianSlim(nil),
        Resources: &types.Resources{
            CPU:    2,
            Memory: 4,
            Disk:   8,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

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

</TabItem>
<TabItem label="CLI" icon="seti:shell">

```shell
daytona create --class small
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "cpu": 2,
  "memory": 4,
  "disk": 8
}'
```

</TabItem>
</Tabs>

### Create an ephemeral sandbox

Create an [ephemeral sandbox](/docs/en/sandboxes#ephemeral-sandboxes) that is automatically deleted when stopped.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona, CreateSandboxFromSnapshotParams

daytona = Daytona()
sandbox = daytona.create(
    CreateSandboxFromSnapshotParams(ephemeral=True, auto_stop_interval=5)
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    ephemeral: true,
    autoStopInterval: 5
});
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    autoStop := 5
    sandbox, err := client.Create(context.Background(), types.SnapshotParams{
        SandboxBaseParams: types.SandboxBaseParams{
            Ephemeral:        true,
            AutoStopInterval: &autoStop,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(ephemeral: true, auto_stop_interval: 5)
)
```

</TabItem>

<TabItem label="CLI" icon="seti:shell">

```shell
daytona create --auto-stop 5 --auto-delete 0
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "autoStopInterval": 5,
  "autoDeleteInterval": 0
}'
```

</TabItem>
</Tabs>

### Create a sandbox from a snapshot

Create a sandbox from a pre-built [snapshot](/docs/en/snapshots) for faster sandbox creation with pre-installed dependencies.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
```python
from daytona import Daytona, CreateSandboxFromSnapshotParams

daytona = Daytona()
sandbox = daytona.create(
    CreateSandboxFromSnapshotParams(
        snapshot="my-snapshot-name",
        language="python"
    )
)
```

</TabItem>

<TabItem label="TypeScript" icon="seti:typescript">
```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    snapshot: 'my-snapshot-name',
    language: 'typescript'
});
```

</TabItem>

<TabItem label="Go" icon="seti:go">
```go
package main

import (
    "context"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    sandbox, err := client.Create(context.Background(), types.SnapshotParams{
        Snapshot: "my-snapshot-name",
        SandboxBaseParams: types.SandboxBaseParams{
            Language: types.CodeLanguagePython,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

</TabItem>

<TabItem label="Ruby" icon="seti:ruby">
```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(
        snapshot: 'my-snapshot-name',
        language: Daytona::CodeLanguage::PYTHON
    )
)
```

</TabItem>

<TabItem label="CLI" icon="seti:shell">
```shell
daytona create --snapshot my-snapshot-name
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "snapshot": "my-snapshot-name"
}'
```

</TabItem>
</Tabs>

### Create a sandbox with a declarative image

Create a sandbox with a [declarative image](/docs/en/declarative-builder) that defines dependencies programmatically.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona, CreateSandboxFromImageParams, Image

daytona = Daytona()
image = (
    Image.debian_slim("3.12")
    .pip_install(["requests", "pandas", "numpy"])
    .workdir("/home/daytona")
)
sandbox = daytona.create(
    CreateSandboxFromImageParams(image=image),
    on_snapshot_create_logs=print
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona, Image } from '@daytonaio/sdk';

const daytona = new Daytona();
const image = Image.debianSlim('3.12')
    .pipInstall(['requests', 'pandas', 'numpy'])
    .workdir('/home/daytona');
const sandbox = await daytona.create(
    { image },
    { onSnapshotCreateLogs: console.log }
);
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    image := daytona.DebianSlim(nil).
        PipInstall([]string{"requests", "pandas", "numpy"}).
        Workdir("/home/daytona")
    sandbox, err := client.Create(context.Background(), types.ImageParams{
        Image: image,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
image = Daytona::Image
    .debian_slim('3.12')
    .pip_install(['requests', 'pandas', 'numpy'])
    .workdir('/home/daytona')
sandbox = daytona.create(
    Daytona::CreateSandboxFromImageParams.new(image: image),
    on_snapshot_create_logs: proc { |chunk| puts chunk }
)
```

</TabItem>
<TabItem label="CLI" icon="seti:shell">

```shell
daytona create --dockerfile ./Dockerfile
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "buildInfo": {
    "dockerfileContent": "FROM python:3.12-slim\nRUN pip install requests pandas numpy\nWORKDIR /home/daytona"
  }
}'
```

</TabItem>
</Tabs>

### Create a sandbox with volumes

Create a sandbox with a [volume](/docs/en/volumes) mounted to share data across sandboxes.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona, CreateSandboxFromSnapshotParams, VolumeMount

daytona = Daytona()
volume = daytona.volume.get("my-volume", create=True)
sandbox = daytona.create(
    CreateSandboxFromSnapshotParams(
        volumes=[VolumeMount(volume_id=volume.id, mount_path="/home/daytona/data")]
    )
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const volume = await daytona.volume.get('my-volume', true);
const sandbox = await daytona.create({
    volumes: [{ volumeId: volume.id, mountPath: '/home/daytona/data' }]
});
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    volume, err := client.Volume.Get(context.Background(), "my-volume")
    if err != nil {
        volume, err = client.Volume.Create(context.Background(), "my-volume")
        if err != nil {
            log.Fatal(err)
        }
    }

    sandbox, err := client.Create(context.Background(), types.SnapshotParams{
        SandboxBaseParams: types.SandboxBaseParams{
            Volumes: []types.VolumeMount{{
                VolumeID:  volume.ID,
                MountPath: "/home/daytona/data",
            }},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">
```ruby
require 'daytona'

daytona = Daytona::Daytona.new
volume = daytona.volume.get('my-volume', create: true)
sandbox = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(
        volumes: [DaytonaApiClient::SandboxVolume.new(
            volume_id: volume.id,
            mount_path: '/home/daytona/data'
        )]
    )
)
```

</TabItem>
<TabItem label="CLI" icon="seti:shell">

```shell
daytona volume create my-volume
daytona create --volume my-volume:/home/daytona/data
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "volumes": [
    {
      "volumeId": "<VOLUME_ID>",
      "mountPath": "/home/daytona/data"
    }
  ]
}'
```

</TabItem>
</Tabs>

### Create a sandbox with a Git repository cloned

Create a sandbox with a [Git repository](/docs/en/typescript-sdk/git) cloned to manage version control.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create()

sandbox.git.clone("https://github.com/daytonaio/daytona.git", "/home/daytona/daytona")
status = sandbox.git.status("/home/daytona/daytona")
print(f"Branch: {status.current_branch}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create();

await sandbox.git.clone('https://github.com/daytonaio/daytona.git', '/home/daytona/daytona');
const status = await sandbox.git.status('/home/daytona/daytona');
console.log(`Branch: ${status.currentBranch}`);
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    sandbox, err := client.Create(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    sandbox.Git.Clone(context.Background(), "https://github.com/daytonaio/daytona.git", "/home/daytona/daytona")
    status, err := sandbox.Git.Status(context.Background(), "/home/daytona/daytona")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Branch: %s\n", status.CurrentBranch)
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create

sandbox.git.clone(url: "https://github.com/daytonaio/daytona.git", path: "/home/daytona/daytona")
status = sandbox.git.status("/home/daytona/daytona")
puts "Branch: #{status.current_branch}"
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
# Create a sandbox
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{}'

# Clone a Git repository in the sandbox
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/git/clone' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "url": "https://github.com/daytonaio/daytona.git",
  "path": "/home/daytona/daytona"
}'

# Get repository status
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/git/status?path=/home/daytona/daytona' \
  --header 'Authorization: Bearer <API_KEY>'
```

</TabItem>
</Tabs>

### Create a sandbox with labels

Create a sandbox with [labels](/docs/en/sandboxes#create-sandboxes) to organize and find sandboxes.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
```python
from daytona import Daytona, CreateSandboxFromSnapshotParams

daytona = Daytona()
sandbox = daytona.create(
    CreateSandboxFromSnapshotParams(labels={"project": "my-app", "env": "dev"})
)

found = daytona.find_one(labels={"project": "my-app"})
print(f"Found sandbox: {found.id}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
```typescript
import { Daytona } from '@daytonaio/sdk';

const daytona = new Daytona();
const sandbox = await daytona.create({
    labels: { project: 'my-app', env: 'dev' }
});

const found = await daytona.findOne({ labels: { project: 'my-app' } });
console.log(`Found sandbox: ${found.id}`);
```

</TabItem>
<TabItem label="Go" icon="seti:go">
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
    "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
    client, err := daytona.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    labels := map[string]string{"project": "my-app", "env": "dev"}
    sandbox, err := client.Create(context.Background(), types.SnapshotParams{
        SandboxBaseParams: types.SandboxBaseParams{Labels: labels},
    })
    if err != nil {
        log.Fatal(err)
    }

    found, err := client.FindOne(context.Background(), nil, map[string]string{"project": "my-app"})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found sandbox: %s\n", found.ID)
}
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">
```ruby
require 'daytona'

daytona = Daytona::Daytona.new
sandbox = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(labels: { 'project' => 'my-app', 'env' => 'dev' })
)

found = daytona.find_one(labels: { 'project' => 'my-app' })
puts "Found sandbox: #{found.id}"
```

</TabItem>

<TabItem label="CLI" icon="seti:shell">
```shell
daytona create --label project=my-app --label env=dev
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
  "labels": {
    "project": "my-app",
    "env": "dev"
  }
}'
```

</TabItem>
</Tabs>
