# Daytona Workspace Plugin for OpenCode

This is an OpenCode workspace plugin that provisions Daytona sandboxes as remote development environments. It uses the experimental workspace API to create isolated sandboxes where the OpenCode server runs remotely.

## Features

- Create Daytona sandboxes as remote OpenCode workspaces
- Automatic repository synchronization via tar archive upload
- OpenCode server runs inside the sandbox, proxying all tool calls
- Preview URLs for web servers running in sandboxes
- Clean sandbox teardown when workspaces are removed

## Requirements

- OpenCode 1.4.11 or later
- Daytona account and API key
- `OPENCODE_EXPERIMENTAL_WORKSPACES=true` environment flag

## Usage

### Installation

Add the plugin to your project's `.opencode/opencode.jsonc`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["@daytona/opencode"]
}
```

The plugin will be automatically downloaded when OpenCode starts.

To install globally, edit `~/.config/opencode/opencode.jsonc`.

### Environment Configuration

This plugin requires a [Daytona account](https://www.daytona.io/) and [Daytona API key](https://app.daytona.io/dashboard/keys).

Set your Daytona API key as an environment variable:

```bash
export DAYTONA_API_KEY="your-api-key"
```

Or create a `.env` file in your project root:

```env
DAYTONA_API_KEY=your-api-key
```

### Running OpenCode

Start OpenCode with the experimental workspaces flag:

```bash
OPENCODE_EXPERIMENTAL_WORKSPACES=true opencode
```

### Creating a Daytona Workspace

1. Press `Ctrl+W` in the session list to open the workspace creation menu
2. Select "Daytona" as the workspace type
3. The plugin will:
   - Create a new Daytona sandbox
   - Upload your repository as a tar archive
   - Install and start the OpenCode server inside the sandbox
   - Wait for the server to become healthy

Once created, all file operations and bash commands run inside the remote sandbox.

### Removing a Workspace

When you delete a Daytona workspace from OpenCode, the associated sandbox is automatically cleaned up.

## How It Works

### Architecture

Unlike traditional tool plugins that intercept and forward commands, this workspace plugin:

1. **Provisions a remote sandbox** with `create()` - uploads your repo and starts OpenCode server
2. **Returns a remote target** with `target()` - OpenCode proxies all requests to the sandbox's server
3. **Cleans up** with `remove()` - tears down the sandbox when the workspace is deleted

### Repository Upload

When creating a workspace:

1. Your local repository is cloned (shallow, depth 1)
2. The clone is packaged as a tar archive
3. The archive is uploaded to the sandbox and extracted
4. The OpenCode project ID is written to `.git/opencode` for session association

### Tool Execution

All OpenCode tools (bash, read, write, edit, glob, grep, etc.) work automatically through the remote target proxy. No custom tool implementations are needed.

## Development

This plugin is part of the Daytona monorepo.

### Setup

Clone the Daytona monorepo:

```bash
git clone https://github.com/daytonaio/daytona
cd daytona
```

Install dependencies:

```bash
yarn install
```

### Development and Testing

To test the plugin locally, create a symlink in your test project:

```bash
mkdir ~/myproject && cd ~/myproject
ln -s [ABSOLUTE_PATH_TO_DAYTONA]/libs/opencode-plugin/.opencode .opencode
git init
OPENCODE_EXPERIMENTAL_WORKSPACES=true opencode
```

> **Note:** When developing locally with a symlink, OpenCode loads the TypeScript source directly, so no build step is required.

### Building

Build the plugin:

```bash
npx nx run opencode-plugin:build
```

### Publishing

```bash
npm login
npx nx run opencode-plugin:publish
```

## Project Structure

```
libs/opencode-plugin/
├── .opencode/
│   └── plugin/
│       ├── daytona/
│       │   └── index.ts          # Workspace adaptor implementation
│       └── index.ts              # Plugin entry point
├── package.json
├── project.json
├── tsconfig.json
├── tsconfig.lib.json
└── README.md
```

## License

Apache-2.0
