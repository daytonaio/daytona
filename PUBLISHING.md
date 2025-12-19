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

### Using Nx

```bash
# From repository root
export PYPI_TOKEN="your-pypi-token"
export PYPI_PKG_VERSION="X.Y.Z" # pre-release format example: "X.Y.Za1"
yarn nx publish sdk-python
```

**Note**: [Guide](https://packaging.python.org/en/latest/discussions/versioning/) for versioning Python packages.

## TypeScript SDK (npm)

### Using Nx

```bash
# From repository root
export NPM_TOKEN="your-npm-token"
export NPM_PKG_VERSION="X.Y.Z" # pre-release format example: "X.Y.Z-alpha.1"
export NPM_TAG="latest"  # or "beta", "alpha", etc.
yarn nx publish sdk-typescript
```

**Note**: NPM packages must have [SemVer-aligned formats](https://semver.org/).

## Ruby SDK (RubyGems)

### Using Nx

```bash
# From repository root
export RUBYGEMS_API_KEY="your-rubygems-api-key"
export RUBYGEMS_PKG_VERSION="X.Y.Z" # pre-release format example: "X.Y.Z.alpha.1"
yarn nx publish sdk-ruby
```

**Note**: [Guide](https://guides.rubygems.org/patterns/#prerelease-gems) for versioning Ruby gems.

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

### Version Format

`MAJOR.MINOR.PATCH` releases follow semantics:

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

Prerelease formats depend on SDK language:

1. For **Typescript** (npm) and **Ruby** (gem) follow semantic versioning ([SemVer](https://semver.org/)): `MAJOR.MINOR.PATCH`

   For pre-releases, use:

   - `0.126.0-alpha.1` - Alpha release
   - `0.126.0-beta.1` - Beta release
   - `0.126.0-rc.1` - Release candidate

2. For **Python** (PyPI) follow Python packages versioning [guide](https://packaging.python.org/en/latest/discussions/versioning/):

   For pre-releases, use:

   - `1.2.0a1` - Alpha release
   - `1.2.0b1` - Beta release
   - `1.2.0rc1` - Release candidate

3. For **Ruby** (gem) follow Ruby gems versioning [guide](https://guides.rubygems.org/patterns/#prerelease-gems):

   For pre-releases, use:

   - `0.126.0.alpha.1` - Alpha release
   - `0.126.0.beta.1` - Beta release
   - `0.126.0.rc.1` - Release candidate

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
gem search daytona --remote --exact
# or
gem info daytona --remote
```

## References

- [Semantic Versioning](https://semver.org/)
- [Python packages versioning](https://packaging.python.org/en/latest/discussions/versioning/)
- [Ruby gems versioning guide](https://guides.rubygems.org/patterns/#prerelease-gems)
