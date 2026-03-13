use daytona::{Client, Config, CreateSandboxParams};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Option 1: Use environment variables (DAYTONA_API_KEY, DAYTONA_API_URL, etc.)
    // let config = Config::from_env();

    // Option 2: Use builder pattern (recommended for explicit configuration)
    let config = Config::builder()
        .api_key(std::env::var("DAYTONA_API_KEY").expect("DAYTONA_API_KEY must be set"))
        .build();

    let client = Client::new(config)?;

    let sandbox = client.create(CreateSandboxParams::default()).await?;
    println!(
        "Created sandbox: {} (state: {:?})",
        sandbox.id, sandbox.state
    );

    let process = sandbox.process().await?;
    let result = process
        .execute_command("echo 'Hello from Daytona Rust SDK'", None, None, None)
        .await?;
    println!("Command output: {}", result.result);

    sandbox.delete().await?;
    println!("Sandbox deleted");

    Ok(())
}
