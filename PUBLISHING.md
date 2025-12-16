# Publishing Daytona SDKs

This document describes how to publish the Daytona SDKs (Python, TypeScript, and Ruby) to their respective package registries.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Python SDK (PyPI)](#python-sdk-pypi)
- [TypeScript SDK (npm)](#typescript-sdk-npm)
- [Ruby SDK (RubyGems)](#ruby-sdk-rubygems)
- [Automated Publishing (CI/CD)](#automated-publishing-cicd)
- [Version Management](#version-management)

## Prerequisites

Before publishing any SDK, ensure you have:

1. **Maintainer Access**: Write access to the Daytona repository
2. **Package Registry Credentials**:
   - PyPI: Token with upload permissions
   - npm: Token with publish permissions
   - RubyGems: API key with push permissions
3. **Local Development Setup**:
   - All dependencies installed (`yarn install`)
   - SDKs built successfully
   - Tests passing

## Python SDK (PyPI)

### Manual Publishing

1. **Navigate to the SDK directory**:

   ```bash
   cd libs/sdk-python
   ```

2. **Update the version** in `pyproject.toml`:

   ```toml
   version = "0.126.0"
   ```

3. **Build the package**:

   ```bash
   poetry build
   ```

4. **Publish to PyPI**:

   ```bash
   poetry publish --username __token__ --password $PYPI_TOKEN
   ```

   The SDK is published under two names:
   - `daytona` (primary)
   - `daytona_sdk` (alias)

### Using Nx

```bash
# From repository root
export PYPI_TOKEN="your-pypi-token"
yarn nx run sdk-python:publish
```

## TypeScript SDK (npm)

### Manual Publishing

1. **Navigate to the SDK directory**:

   ```bash
   cd libs/sdk-typescript
   ```

2. **Update the version** in `package.json`:

   ```json
   {
     "version": "0.126.0"
   }
   ```

3. **Build the package**:

   ```bash
   yarn build
   ```

4. **Publish to npm**:

   ```bash
   npm publish --access public --registry https://registry.npmjs.org/ \
     --//registry.npmjs.org/:_authToken=$NPM_TOKEN
   ```

### Using Nx

```bash
# From repository root
export NPM_TOKEN="your-npm-token"
export NPM_TAG="latest"  # or "beta", "alpha", etc.
yarn nx run sdk-typescript:publish
```

## Ruby SDK (RubyGems)

The Ruby SDK consists of two gems that must be published in order:

1. **`daytona_api_client`** - Low-level API client (dependency)
2. **`daytona-sdk`** - High-level SDK

### Manual Publishing

#### Step 1: Configure Credentials

Create or update `~/.gem/credentials`:

```bash
mkdir -p ~/.gem
echo "---" > ~/.gem/credentials
echo ":rubygems_api_key: YOUR_API_KEY" >> ~/.gem/credentials
chmod 0600 ~/.gem/credentials
```

Or set the environment variable:

```bash
export RUBYGEMS_API_KEY="your-rubygems-api-key"
```

#### Step 2: Update Version

Update `libs/sdk-ruby/lib/daytona/sdk/version.rb`:

```ruby
module Daytona
  module Sdk
    VERSION = '0.126.0'
  end
end
```

#### Step 3: Publish API Client (if needed)

```bash
cd libs/api-client-ruby
gem build daytona_api_client.gemspec
gem push daytona_api_client-1.0.0.gem
cd ../..
```

#### Step 4: Publish SDK

```bash
cd libs/sdk-ruby
gem build daytona-sdk.gemspec
gem push daytona-sdk-0.126.0.gem
cd ../..
```

### Using the Publish Script

The Ruby SDK includes a convenient publish script:

```bash
# Set your RubyGems API key
export RUBYGEMS_API_KEY="your-api-key"

# Publish current version
./libs/sdk-ruby/scripts/publish.sh

# Or publish with a new version
./libs/sdk-ruby/scripts/publish.sh 0.126.0
```

The script will:

- Update the version if provided
- Check if gems already exist on RubyGems
- Build and publish the API client (if not already published)
- Build and publish the SDK
- Clean up build artifacts

### Using Nx

```bash
# From repository root
export RUBYGEMS_API_KEY="your-rubygems-api-key"

# Publish API client (builds and publishes)
yarn nx run api-client-ruby:publish

# Publish SDK (depends on api-client-ruby, builds and publishes)
yarn nx run sdk-ruby:publish
```

**Note**: The `build` target only installs dependencies. The `publish` target handles building the gem and pushing to RubyGems.

## Automated Publishing (CI/CD)

### GitHub Actions Workflow

The repository includes a GitHub Actions workflow for automated publishing: `.github/workflows/sdk_publish.yaml`

#### Triggering a Release

1. Go to **Actions** → **SDK and CLI Publish** in the GitHub repository
2. Click **Run workflow**
3. Fill in the parameters:
   - **version**: The version to release (e.g., `v0.126.0`)
   - **pypi_pkg_version**: (Optional) Override PyPI version
   - **npm_pkg_version**: (Optional) Override npm version
   - **rubygems_pkg_version**: (Optional) Override RubyGems version
   - **npm_tag**: npm dist-tag (default: `latest`)

#### Required Secrets

Ensure these secrets are configured in GitHub repository settings:

- `PYPI_TOKEN`: PyPI API token
- `NPM_TOKEN`: npm access token
- `RUBYGEMS_API_KEY`: RubyGems API key
- `GITHUBBOT_TOKEN`: GitHub token for Homebrew tap updates

### What the Workflow Does

1. Checks out the code
2. Sets up all required environments (Go, Java, Python, Node.js, Ruby)
3. Installs dependencies
4. Configures credentials for all package registries
5. Runs `yarn publish` which uses Nx to publish all SDKs in the correct order
6. Updates the Homebrew tap (for the CLI)

## Version Management

### Version Synchronization

All SDKs should use the same version number for consistency. When releasing:

1. Update version in all SDK version files:
   - Python: `libs/sdk-python/pyproject.toml`
   - TypeScript: `libs/sdk-typescript/package.json`
   - Ruby: `libs/sdk-ruby/lib/daytona/sdk/version.rb`

2. Ensure API clients have compatible versions:
   - Python API client: `libs/api-client-python`
   - TypeScript API client: `libs/api-client`
   - Ruby API client: `libs/api-client-ruby`

### Version Format

Follow semantic versioning (SemVer): `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

For pre-releases, use:

- `0.126.0-alpha.1` - Alpha release
- `0.126.0-beta.1` - Beta release
- `0.126.0-rc.1` - Release candidate

### Checking Published Versions

#### PyPI

```bash
pip index versions daytona
# or
curl -s https://pypi.org/pypi/daytona/json | jq -r .info.version
```

#### npm

```bash
npm view @daytonaio/sdk version
# or
npm info @daytonaio/sdk
```

#### RubyGems

```bash
gem search daytona-sdk --remote --exact
# or
gem info daytona-sdk --remote
```

## Troubleshooting

### Python: "Package already exists"

If the version already exists on PyPI, you must increment the version number. PyPI does not allow republishing to the same version.

### TypeScript: "Cannot publish over existing version"

Similar to PyPI, npm does not allow republishing. Increment the version or use a pre-release tag.

### Ruby: "Repushing of gem versions is not allowed"

RubyGems also prevents republishing. You must:

1. Increment the version
2. Or use `gem yank` to remove the existing version (not recommended for production releases)

### Ruby: "MFA required"

If you see "Rubygem requires owners to enable MFA":

1. Enable MFA on your RubyGems account: https://rubygems.org/settings/edit
2. Or temporarily disable MFA requirement in the gemspec:

   ```ruby
   # In daytona-sdk.gemspec
   spec.metadata['rubygems_mfa_required'] = 'false'
   ```

### Authentication Issues

If publishing fails with authentication errors:

1. **Verify credentials are set**:

   ```bash
   echo $PYPI_TOKEN
   echo $NPM_TOKEN
   echo $RUBYGEMS_API_KEY
   ```

2. **Check credential format**:
   - PyPI: Should start with `pypi-`
   - npm: Should be a valid npm token
   - RubyGems: Should start with `rubygems_`

3. **Verify token permissions**:
   - Ensure tokens have push/publish permissions
   - Check token hasn't expired

## Best Practices

1. **Test Before Publishing**:
   - Run all tests: `yarn test`
   - Build all packages: `yarn build`
   - Test installation locally

2. **Update Changelog**:
   - Document all changes in the release
   - Follow the existing changelog format

3. **Version Consistency**:
   - Keep all SDK versions synchronized
   - Update all version files before publishing

4. **Use CI/CD for Production**:
   - Prefer the GitHub Actions workflow for releases
   - Manual publishing is fine for testing and pre-releases

5. **Tag Releases**:
   - Create Git tags for releases
   - Use the format: `v0.126.0`

6. **Announce Releases**:
   - Update documentation
   - Notify users through appropriate channels

## Support

For issues with publishing:

1. Check this document first
2. Review CI/CD logs in GitHub Actions
3. Contact the maintainers team
4. Open an issue in the repository

## References

- [PyPI Publishing Guide](https://packaging.python.org/tutorials/packaging-projects/)
- [npm Publishing Guide](https://docs.npmjs.com/cli/v8/commands/npm-publish)
- [RubyGems Publishing Guide](https://guides.rubygems.org/publishing/)
- [Semantic Versioning](https://semver.org/)
