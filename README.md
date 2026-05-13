&nbsp;

<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-white.png">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-black.png">
    <img alt="Daytona logo" src="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-black.png" width="50%">
  </picture>
</div>

<h3 align="center">
  Run AI Code.
  <br/>
  Secure and Elastic Infrastructure for
  Running Your AI-Generated Code.
</h3>

<p align="center">
    <a href="https://www.daytona.io/docs"> Documentation </a>·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+"> Report Bug </a>·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+"> Request Feature </a>·
    <a href="https://go.daytona.io/slack"> Join our Slack </a>·
    <a href="https://x.com/daytonaio"> Connect on X </a>
</p>

&nbsp;

Daytona is a secure and elastic infrastructure runtime for AI-generated code execution and agent workflows. Our open-source platform provides [sandboxes](https://www.daytona.io/docs/en/sandboxes/), full composable computers with complete isolation, a dedicated kernel, filesystem, network stack, and allocated vCPU, RAM, and disk.

Sandboxes are the core component of the Daytona platform, spinning up in under 90ms from code to execution and running any code in Python, TypeScript, and JavaScript. Built on OCI/Docker compatibility, massive parallelization, and unlimited persistence, sandboxes deliver consistent, predictable environments for agent workflows.

Agents and developers interact with sandboxes programmatically using the Daytona [SDKs](https://www.daytona.io/docs/en/#3-install-the-sdk), [API](https://www.daytona.io/docs/en/tools/api/#daytona/), and [CLI](https://www.daytona.io/docs/en/tools/cli/). Operations span sandbox lifecycle management, filesystem operations, process and code execution, and runtime configuration through base images, packages, and tooling. Our stateful environment [snapshots](https://www.daytona.io/docs/en/snapshots/) enable persistent agent operations across sessions, making Daytona the ideal foundation for AI agent architectures.

## Features

Daytona provides an extensive set of features and tools for interacting with sandboxes.

- **Platform**: governance and operational controls for organizations standardizing on Daytona
- **Sandboxes**: isolated full composable computers that execute workloads and retain state
- **Agent tools**: programmatic capabilities for application code, agents, and integrations
- **Human tools**: interfaces and remote sessions for interacting with sandboxes
- **System tools**: platform-level hooks and controls for lifecycle events and network access

| Platform                                                                   | Sandboxes                                                               | Agent tools                                                                       | Human tools                                                               | System tools                                                  |
| :------------------------------------------------------------------------- | :---------------------------------------------------------------------- | :-------------------------------------------------------------------------------- | :------------------------------------------------------------------------ | :------------------------------------------------------------ |
| [Organizations](https://www.daytona.io/docs/en/organizations/)                | [Environment](https://www.daytona.io/docs/en/configuration/)               | [Process & code execution](https://www.daytona.io/docs/en/process-code-execution/)   | [Dashboard](https://www.daytona.io/docs/en/getting-started#dashboard)        | [Webhooks](https://www.daytona.io/docs/en/webhooks/)             |
| [API Keys](https://www.daytona.io/docs/en/api-keys/)                          | [Snapshots](https://www.daytona.io/docs/en/snapshots/)                     | [File system operations](https://www.daytona.io/docs/en/file-system-operations/)     | [Web terminal](https://www.daytona.io/docs/en/web-terminal/)                 | [Network limits](https://www.daytona.io/docs/en/network-limits/) |
| [Limits](https://www.daytona.io/docs/en/limits/)                              | [Declarative builder](https://www.daytona.io/docs/en/declarative-builder/) | [Language server protocol](https://www.daytona.io/docs/en/language-server-protocol/) | [SSH access](https://www.daytona.io/docs/en/ssh-access/)                     |                                                               |
| [Billing](https://www.daytona.io/docs/en/billing/)                            | [Volumes](https://www.daytona.io/docs/en/volumes/)                         | [Computer use](https://www.daytona.io/docs/en/computer-use/)                         | [VNC access](https://www.daytona.io/docs/en/vnc-access/)                     |                                                               |
| [Audit logs](https://www.daytona.io/docs/en/audit-logs/)                      | [Regions](https://www.daytona.io/docs/en/regions/)                         | [MCP server](https://www.daytona.io/docs/en/mcp/)                                    | [VPN connection](https://www.daytona.io/docs/en/vpn-connections/)            |                                                               |
| [OpenTelemetry](https://www.daytona.io/docs/en/experimental/otel-collection/) |                                                                         | [Git operations](https://www.daytona.io/docs/en/git-operations/)                     | [Preview](https://www.daytona.io/docs/en/preview/)                           |                                                               |
| [Integrations](https://www.daytona.io/docs/en/guides/)                        |                                                                         | [Pseudo terminal (PTY)](https://www.daytona.io/docs/en/pty/)                         | [Custom preview proxy](https://www.daytona.io/docs/en/custom-preview-proxy/) |                                                               |
| [Security exhibit](https://www.daytona.io/docs/en/security-exhibit/)          |                                                                         | [Log streaming](https://www.daytona.io/docs/en/log-streaming/)                       | [Playground](https://www.daytona.io/docs/en/playground/)                     |                                                               |

## Architecture

Daytona platform is organized into multiple plane components, each serving a specific purpose. A detailed overview of each component is available in the [architecture documentation](https://www.daytona.io/docs/en/architecture/).

- **Interface plane**: provides client interfaces for interacting with Daytona
- **Control plane**: orchestrates all sandbox operations
- **Compute plane**: runs and manages sandbox instances

### Applications

Runnable applications and services for the Daytona platform. Each directory is a deployable or buildable component, available in the [apps](apps) directory.

- [`api`](apps/api): NestJS-based RESTful service; primary entry point for all platform operations
- [`cli`](apps/cli): Go command-line interface access to core features for interacting with sandboxes
- [`daemon`](apps/daemon): code execution agent that runs inside each sandbox
- [`dashboard`](apps/dashboard): web user interface for visual sandbox management
- [`docs`](apps/docs): documentation content; website published to [daytona.io/docs](https://www.daytona.io/docs/en/)
- [`otel-collector`](apps/otel-collector): trace and metric collection for Daytona SDK operations
- [`proxy`](apps/proxy): reverse proxy for custom routing and preview URLs
- [`runner`](apps/runner): compute nodes that power Daytona's compute plane and run sandboxes
- [`snapshot-manager`](apps/snapshot-manager): orchestrates the creation of sandbox snapshots
- [`ssh-gateway`](apps/ssh-gateway): standalone SSH gateway that accepts authenticated `ssh` connections

### Client libraries

Client libraries integrate the Daytona platform from application code through developer-facing SDKs backed by OpenAPI-generated REST clients and toolbox API clients. Each directory is a publishable package for a specific language or runtime, available in the [libs](libs) directory.

#### Python

```bash
pip install daytona
```

Standalone packages and libraries for interacting with Daytona using Python:

> [`sdk-python`](libs/sdk-python) • [`api-client-python`](libs/api-client-python) • [`api-client-python-async`](libs/api-client-python-async) • [`toolbox-api-client-python`](libs/toolbox-api-client-python) • [`toolbox-api-client-python-async`](libs/toolbox-api-client-python-async)

#### TypeScript

```bash
npm install @daytona/sdk
```

Standalone packages and libraries for interacting with Daytona using TypeScript:

> [`sdk-typescript`](libs/sdk-typescript) • [`api-client`](libs/api-client) • [`toolbox-api-client`](libs/toolbox-api-client)

#### Ruby

```bash
gem install daytona
```

Standalone packages and libraries for interacting with Daytona using Ruby:

> [`sdk-ruby`](libs/sdk-ruby) • [`api-client-ruby`](libs/api-client-ruby) • [`toolbox-api-client-ruby`](libs/toolbox-api-client-ruby)

#### Go

```bash
go get github.com/daytonaio/daytona/libs/sdk-go
```

Standalone packages and libraries for interacting with Daytona using Go:

> [`sdk-go`](libs/sdk-go) • [`api-client-go`](libs/api-client-go) • [`toolbox-api-client-go`](libs/toolbox-api-client-go)

#### Java

Gradle (`build.gradle.kts`):

```kotlin
dependencies {
    implementation("io.daytona:sdk:0.1.0")
}
```

Maven (`pom.xml`):

```xml
<dependency>
  <groupId>io.daytona</groupId>
  <artifactId>sdk</artifactId>
  <version>0.1.0</version>
</dependency>
```

Standalone packages and libraries for interacting with Daytona using Java:

> [`sdk-java`](libs/sdk-java) • [`api-client-java`](libs/api-client-java) • [`toolbox-api-client-java`](libs/toolbox-api-client-java)

## Deployments

Daytona is available as a managed service on [app.daytona.io](https://app.daytona.io). Daytona can run as a fully hosted service, as an open-source stack you operate, or in a hybrid setup where Daytona orchestrates sandboxes while execution happens on machines you manage.

- [Open source deployment](https://www.daytona.io/docs/en/oss-deployment/): full local stack from the [`docker`](docker) directory using Docker Compose
- [Customer managed compute](https://www.daytona.io/docs/en/runners/): custom regions and runner machines that operate Daytona sandboxes on your own compute infrastructure

## Quick Start

1. Create an account at [app.daytona.io](https://app.daytona.io)
2. Generate an [API key](https://app.daytona.io/dashboard/keys)
3. Create a sandbox

### Python SDK

```py
from daytona import Daytona, DaytonaConfig

config = DaytonaConfig(api_key="YOUR_API_KEY")
daytona = Daytona(config)
sandbox = daytona.create()
response = sandbox.process.code_run('print("Hello World!")')
print(response.result)
```

### Typescript SDK

```jsx
import { Daytona } from "@daytona/sdk";

const daytona = new Daytona({ apiKey: "YOUR_API_KEY" });
const sandbox = await daytona.create();
const response = await sandbox.process.codeRun('print("Hello World!")');
console.log(response.result);
```

### Ruby SDK

```ruby
require 'daytona'

config = Daytona::Config.new(api_key: 'YOUR_API_KEY')
daytona = Daytona::Daytona.new(config)
sandbox = daytona.create
response = sandbox.process.code_run(code: 'print("Hello World!")')
puts response.result
```

### Go SDK

```go
package main

import (
  "context"
  "fmt"
  "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
  "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
  config := &types.DaytonaConfig{APIKey: "YOUR_API_KEY"}
  client, _ := daytona.NewClientWithConfig(config)
  ctx := context.Background()
  sandbox, _ := client.Create(ctx, nil)
  response, _ := sandbox.Process.ExecuteCommand(ctx, "echo 'Hello World!'")
  fmt.Println(response.Result)
}
```

### Java SDK

```java
import io.daytona.sdk.Daytona;
import io.daytona.sdk.DaytonaConfig;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.ExecuteResponse;

public class Main {
  public static void main(String[] args) {
    DaytonaConfig config = new DaytonaConfig.Builder()
        .apiKey("YOUR_API_KEY")
        .build();
    try (Daytona daytona = new Daytona(config)) {
      Sandbox sandbox = daytona.create();
      ExecuteResponse response = sandbox.getProcess().executeCommand("echo 'Hello World!'");
      System.out.println(response.getResult());
    }
  }
}
```

### API

```bash
curl 'https://app.daytona.io/api/sandbox' \
  --request POST \
  --header 'Authorization: Bearer <YOUR_API_KEY>' \
  --header 'Content-Type: application/json' \
  --data '{}'
```

### CLI

```bash
daytona create
```

## Development

### Devcontainer (full environment)

Open this repository in a [devcontainer](https://containers.dev/)-compatible editor (VS Code, GitHub Codespaces) for a batteries-included setup with all languages, tools, and supporting services.

### Nix (lightweight, agent-friendly)

If you prefer working outside the devcontainer — or are an AI agent executing build commands — use the Nix dev shells:

```bash
# Enter the full dev shell (Go + Node + Python + Ruby + JDK)
nix develop

# Or pick a language-specific shell
nix develop .#go       # Go services & libs
nix develop .#node     # TypeScript / Node.js apps & libs
nix develop .#python   # Python SDKs & libs
nix develop .#ruby     # Ruby SDKs & libs
nix develop .#java     # Java SDKs & libs
```

**Prerequisites:** [Nix](https://nixos.org/download/) with flakes enabled (`experimental-features = nix-command flakes` in `~/.config/nix/nix.conf`).

For non-interactive / CI usage:

```bash
nix develop .#go --command bash -c "go build ./..."
```

Optional: Install [direnv](https://direnv.net/) + [nix-direnv](https://github.com/nix-community/nix-direnv) for automatic shell activation when you `cd` into the project.

See [`AGENTS.md`](AGENTS.md) for the full shell reference, project-to-shell mapping, and common commands.

> **Note:** Supporting services (PostgreSQL, Redis, etc.) are still managed via `docker compose -f .devcontainer/docker-compose.yaml up`.

---

## Contributing

> [!NOTE]
> Daytona is Open Source under the [GNU AFFERO GENERAL PUBLIC LICENSE](LICENSE), and is the [copyright of its contributors](NOTICE). If you would like to contribute to the software, read the [Developer Certificate of Origin Version 1.1](https://developercertificate.org/) and the [contributing guide](CONTRIBUTING.md) to get started.
