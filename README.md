<div align="center">

[![Documentation](https://img.shields.io/github/v/release/daytonaio/docs?label=Docs&color=23cc71)](https://www.daytona.io/docs)
![License](https://img.shields.io/badge/License-AGPL--3-blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/daytonaio/daytona)](https://goreportcard.com/report/github.com/daytonaio/daytona)
[![Issues - daytona](https://img.shields.io/github/issues/daytonaio/daytona)](https://github.com/daytonaio/daytona/issues)
![GitHub Release](https://img.shields.io/github/v/release/daytonaio/daytona)

</div>

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

<p align="center">
    <a href="https://www.producthunt.com/posts/daytona-2?embed=true&utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-daytona&#0045;2" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=957617&theme=neutral&period=daily&t=1746176740150" alt="Daytona&#0032; - Secure&#0032;and&#0032;elastic&#0032;infra&#0032;for&#0032;running&#0032;your&#0032;AI&#0045;generated&#0032;code&#0046; | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
    <a href="https://www.producthunt.com/posts/daytona-2?embed=true&utm_source=badge-top-post-topic-badge&utm_medium=badge&utm_souce=badge-daytona&#0045;2" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-topic-badge.svg?post_id=957617&theme=neutral&period=monthly&topic_id=237&t=1746176740150" alt="Daytona&#0032; - Secure&#0032;and&#0032;elastic&#0032;infra&#0032;for&#0032;running&#0032;your&#0032;AI&#0045;generated&#0032;code&#0046; | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
</p>

---

## Installation

### Python SDK

```bash
pip install daytona
```

### TypeScript SDK

```bash
npm install @daytonaio/sdk
```

### Rust SDK

```toml
[dependencies]
daytona = "0.0.0"
```

---

## Features

- **Lightning-Fast Infrastructure**: Sub-90ms Sandbox creation from code to execution.
- **Separated & Isolated Runtime**: Execute AI-generated code with zero risk to your infrastructure.
- **Massive Parallelization for Concurrent AI Workflows**: Fork Sandbox filesystem and memory state (Coming soon!)
- **Programmatic Control**: File, Git, LSP, and Execute API
- **Unlimited Persistence**: Your Sandboxes can live forever
- **OCI/Docker Compatibility**: Use any OCI/Docker image to create a Sandbox

---

## Quick Start

1. Create an account at https://app.daytona.io
1. Generate a [new API key](https://app.daytona.io/dashboard/keys)
1. Follow the [Getting Started docs](https://www.daytona.io/docs/getting-started/) to start using the Daytona SDK

## Creating your first Sandbox

### Python SDK

```py
from daytona import Daytona, DaytonaConfig, CreateSandboxBaseParams

# Initialize the Daytona client
daytona = Daytona(DaytonaConfig(api_key="YOUR_API_KEY"))

# Create the Sandbox instance
sandbox = daytona.create(CreateSandboxBaseParams(language="python"))

# Run code securely inside the Sandbox
response = sandbox.process.code_run('print("Sum of 3 and 4 is " + str(3 + 4))')
if response.exit_code != 0:
    print(f"Error running code: {response.exit_code} {response.result}")
else:
    print(response.result)

# Clean up the Sandbox
daytona.delete(sandbox)
```

### Typescript SDK

```jsx
import { Daytona } from '@daytonaio/sdk'

async function main() {
  // Initialize the Daytona client
  const daytona = new Daytona({
    apiKey: 'YOUR_API_KEY',
  })

  let sandbox
  try {
    // Create the Sandbox instance
    sandbox = await daytona.create({
      language: 'typescript',
    })
    // Run code securely inside the Sandbox
    const response = await sandbox.process.codeRun('console.log("Sum of 3 and 4 is " + (3 + 4))')
    if (response.exitCode !== 0) {
      console.error('Error running code:', response.exitCode, response.result)
    } else {
      console.log(response.result)
    }
  } catch (error) {
    console.error('Sandbox flow error:', error)
  } finally {
    if (sandbox) await daytona.delete(sandbox)
  }
}

main().catch(console.error)
```

### Go SDK

```go
package main

import (
 "context"
 "fmt"
 "log"
 "time"

 "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
 "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
 // Initialize the Daytona client with DAYTONA_API_KEY in env
  // Alternative is to use daytona.NewClientWithConfig(...) for more specific config
 client, err := daytona.NewClient()
 if err != nil {
  log.Fatalf("Failed to create client: %v", err)
 }

 ctx := context.Background()

 // Create the Sandbox instance
 params := types.SnapshotParams{
  SandboxBaseParams: types.SandboxBaseParams{
   Language: types.CodeLanguagePython,
  },
 }

 sandbox, err := client.Create(ctx, params, daytona.WithTimeout(90*time.Second))
 if err != nil {
  log.Fatalf("Failed to create sandbox: %v", err)
 }

 // Run code securely inside the Sandbox
 response, err := sandbox.Process.ExecuteCommand(ctx, `python3 -c "print('Sum of 3 and 4 is', 3 + 4)"`)
 if err != nil {
  log.Fatalf("Failed to execute command: %v", err)
 }

 if response.ExitCode != 0 {
  fmt.Printf("Error running code: %d %s\n", response.ExitCode, response.Result)
 } else {
  fmt.Println(response.Result)
 }

 // Clean up the Sandbox
 if err := sandbox.Delete(ctx); err != nil {
  log.Fatalf("Failed to delete sandbox: %v", err)
 }
}
```

### Rust SDK

```rust
use daytona::{Client, Config, CreateSandboxParams, CodeLanguage};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Config::builder().api_key("YOUR_API_KEY").build();
    let client = Client::new(config)?;

    let params = CreateSandboxParams {
        language: Some(CodeLanguage::Python),
        ..Default::default()
    };

    let sandbox = client.create(params).await?;
    let result = sandbox.process().await?.execute_command(
        "python3 -c 'print(\"Sum of 3 and 4 is\", 3 + 4)'",
        None, None, None
    ).await?;

    if result.exit_code != 0 {
        eprintln!("Error running code: {} {}", result.exit_code, result.result);
    } else {
        println!("{}", result.result);
    }

    sandbox.delete().await?;
    Ok(())
}
```

---

## Contributing

Daytona is Open Source under the [GNU AFFERO GENERAL PUBLIC LICENSE](LICENSE), and is the [copyright of its contributors](NOTICE). If you would like to contribute to the software, read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/). Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.
