# Ruby SDK Examples

This directory contains example scripts demonstrating how to use the Daytona Ruby SDK.

## Prerequisites

1. **Environment Variables** - Set the following before running examples:

   ```bash
   export DAYTONA_API_KEY="your-api-key"
   export DAYTONA_API_URL="https://app.daytona.io/api"  # optional, this is the default
   ```

2. **Ruby** - Ensure Ruby is installed (the devcontainer includes Ruby 3.4.5)

## Running Examples

Use the `daytona-ruby` command to run any example:

```bash
daytona-ruby examples/ruby/<example-folder>/<script>.rb
```

For example:

```bash
daytona-ruby examples/ruby/exec-command/exec_session.rb
daytona-ruby examples/ruby/lifecycle/lifecycle.rb
daytona-ruby examples/ruby/file-operations/main.rb
```

## Manual Setup (if alias not available)

If the `daytona-ruby` alias isn't available, you can run examples manually:

```bash
ruby -I/workspaces/daytona/libs/sdk-ruby/lib \
     -I/workspaces/daytona/libs/api-client-ruby/lib \
     -r daytona/sdk \
     examples/ruby/lifecycle/lifecycle.rb
```

Or using bundler from the SDK directory:

```bash
cd /workspaces/daytona/libs/sdk-ruby
bundle exec ruby -r bundler/setup -r daytona/sdk ../../examples/ruby/lifecycle/lifecycle.rb
```
