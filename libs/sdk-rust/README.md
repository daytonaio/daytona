# Daytona Rust SDK

Rust SDK for Daytona sandbox lifecycle and toolbox operations.

## Installation

```toml
[dependencies]
daytona = "0.0.0"
```

## Configuration

Supported environment variables:

- `DAYTONA_API_KEY`
- `DAYTONA_JWT_TOKEN`
- `DAYTONA_ORGANIZATION_ID`
- `DAYTONA_API_URL` (default: `https://app.daytona.io/api`)
- `DAYTONA_TARGET`

## Quick Start

```rust
use daytona::{Client, Config, CreateSandboxParams};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Config::builder().api_key("YOUR_API_KEY").build();
    let client = Client::new(config)?;

    let sandbox = client.create(CreateSandboxParams::default()).await?;
    let result = sandbox.process().execute_command("echo hello", None, None, None).await?;
    println!("{}", result.result);

    sandbox.delete().await?;
    Ok(())
}
```

## Implemented Modules

- `Client`: create/get/findOne/list/delete/start/stop
- `Sandbox`: filesystem, git, process, code interpreter, computer use, lsp server
- `VolumeService`, `SnapshotService`, `ObjectStorageService`

## License

Apache-2.0
