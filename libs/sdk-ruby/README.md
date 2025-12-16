# Daytona Ruby SDK

The official Ruby SDK for [Daytona](https://daytona.io) - a platform for secure, isolated sandbox environments.

## Installation

Add this line to your application's Gemfile:

```ruby
gem 'daytona-sdk'
```

And then execute:

```bash
bundle install
```

Or install it yourself as:

```bash
gem install daytona-sdk
```

## Quick Start

```ruby
require 'daytona'

# Initialize the client (uses DAYTONA_API_KEY environment variable)
daytona = Daytona::Daytona.new

# Or with explicit configuration
config = Daytona::Config.new(
  api_key: 'your-api-key',
  target: 'us'
)
daytona = Daytona::Daytona.new(config)

# Create a sandbox
sandbox = daytona.create

# Execute code
response = sandbox.process.code_run(code: 'print("Hello, World!")')
puts response.result

# Clean up
daytona.delete(sandbox)
```

## Configuration

The SDK can be configured using environment variables:

| Variable | Description |
|----------|-------------|
| `DAYTONA_API_KEY` | API key for authentication |
| `DAYTONA_API_URL` | URL of the Daytona API (defaults to `https://app.daytona.io/api`) |
| `DAYTONA_TARGET` | Target location for Sandboxes |

## Documentation

- [Ruby SDK Reference](https://www.daytona.io/docs/en/ruby-sdk)
- [Getting Started Guide](https://www.daytona.io/docs/en/getting-started)
- [API Documentation](https://www.daytona.io/docs/en/tools/api)

## Examples

See the [examples/ruby](https://github.com/daytonaio/daytona/tree/main/examples/ruby) directory for more usage examples:

- [Lifecycle management](https://github.com/daytonaio/daytona/tree/main/examples/ruby/lifecycle)
- [File operations](https://github.com/daytonaio/daytona/tree/main/examples/ruby/file-operations)
- [Git operations](https://github.com/daytonaio/daytona/tree/main/examples/ruby/git-lsp)
- [Process execution](https://github.com/daytonaio/daytona/tree/main/examples/ruby/exec-command)
- [PTY sessions](https://github.com/daytonaio/daytona/tree/main/examples/ruby/pty)
- [Volumes](https://github.com/daytonaio/daytona/tree/main/examples/ruby/volumes)

## Development

After checking out the repo, run `bin/setup` to install dependencies. Then, run `rake spec` to run the tests. You can also run `bin/console` for an interactive prompt.

### Publishing a New Version

1. Update the version in `lib/daytona/sdk/version.rb`
2. Run the publish script:

```bash
# Set your RubyGems API key
export RUBYGEMS_API_KEY=your_api_key

# Publish
./scripts/publish.sh

# Or publish with a specific version
./scripts/publish.sh 0.126.0
```

## Requirements

- Ruby >= 3.2.0

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/daytonaio/daytona.

## Code of Conduct

Everyone interacting in the Daytona SDK project's codebases, issue trackers, chat rooms and mailing lists is expected to follow the [code of conduct](https://github.com/daytonaio/daytona/blob/main/CODE_OF_CONDUCT.md).

## License

See [LICENSE](https://github.com/daytonaio/daytona/blob/main/LICENSE) for details.
