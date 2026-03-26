# Ruby SDK Examples

This directory contains example scripts demonstrating how to use the Daytona Ruby SDK.

## Prerequisites

1. **Environment Variables** - Configure your API credentials using one of these methods:

   **Option A: Using .env files (Recommended)**

   Create a `.env.local` file in the directory where you run your code:

   ```bash
   # Required (choose one authentication method)
   DAYTONA_API_KEY=your-api-key
   # OR
   DAYTONA_JWT_TOKEN=your-jwt-token
   DAYTONA_ORGANIZATION_ID=your-org-id  # required when using JWT token

   # Optional
   DAYTONA_API_URL=https://app.daytona.io/api  # defaults to this if not specified
   DAYTONA_TARGET=us  # defaults to your organization's default region
   ```

   The SDK automatically loads only Daytona-specific variables from `.env` and `.env.local` files in the current working directory, where `.env.local` overrides `.env`. Runtime environment variables always take precedence over `.env` files.

   **Option B: Export manually**

   ```bash
   export DAYTONA_API_KEY="your-api-key"
   export DAYTONA_API_URL="https://app.daytona.io/api"  # optional, this is the default
   export DAYTONA_TARGET="us"  # optional
   ```

2. **Ruby** - Ensure Ruby is installed (the devcontainer includes Ruby 3.4.5)

3. **Devcontainer Setup** - The devcontainer automatically sets up the Ruby environment with the SDK libraries in your `RUBYLIB` path

## Running Examples

Use the `ruby` command to run any example:

```bash
ruby examples/ruby/<example-folder>/<script>.rb
```

For example:

```bash
ruby examples/ruby/exec-command/exec_session.rb
ruby examples/ruby/lifecycle/lifecycle.rb
ruby examples/ruby/file-operations/main.rb
```

The SDK and all client libraries are loaded from source files in the `libs/` directory, so any changes you make to the SDK will be reflected immediately when you run examples.

## How It Works

The devcontainer sets up the following environment variables:

- **`RUBYLIB`** - Includes paths to the SDK and client library source files
- **`BUNDLE_GEMFILE`** - Points to the SDK's Gemfile for dependency management

This allows you to use plain `ruby` commands while still loading everything from source, ensuring all changes are reflected automatically.
