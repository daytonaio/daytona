# Java SDK Examples

This directory contains example projects demonstrating how to use the Daytona Java SDK.

## Prerequisites

1. **Environment Variables** - Configure your API credentials:

   ```bash
   export DAYTONA_API_KEY="your-api-key"
   export DAYTONA_API_URL="https://app.daytona.io/api"  # optional, this is the default
   export DAYTONA_TARGET="us"  # optional
   ```

2. **Java** - Ensure JDK 11+ is installed (the devcontainer includes JDK 21)

## Running Examples

```bash
examples/java/gradlew -p examples/java/<example-name> run
```

For example:

```bash
examples/java/gradlew -p examples/java/lifecycle run
examples/java/gradlew -p examples/java/exec-command run
examples/java/gradlew -p examples/java/file-operations run
```

## How It Works

Each example is a standalone Gradle project. The shared Gradle wrapper (`gradlew`) in this directory handles downloading Gradle and building examples.

Examples resolve the SDK from source via Gradle composite builds — any changes you make to the SDK or API clients in `libs/` are reflected immediately without any install or publish step.
