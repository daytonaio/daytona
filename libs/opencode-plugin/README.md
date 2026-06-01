# Daytona Workspace Plugin for OpenCode

OpenCode plugin that provisions Daytona sandboxes as remote workspaces.

## Features

- Create Daytona sandboxes as remote OpenCode workspaces
- Automatic repository upload to sandbox
- Preview URLs for web servers running in sandboxes
- Sandbox cleanup when workspaces are removed

## Requirements

- OpenCode 1.14.x or later
- Daytona account and API key
- `OPENCODE_EXPERIMENTAL_WORKSPACES=true` environment flag

## Usage

### Installation

Add the plugin to your project's `.opencode/opencode.jsonc`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["@daytona/opencode"],
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

1. Type `/warp` in the prompt
2. Select "Daytona" as the workspace type
3. The plugin will create a sandbox, upload your repository, and start the OpenCode server

Once created, all commands run inside the remote sandbox.

### Removing a Workspace

When you delete a Daytona workspace from OpenCode, the associated sandbox is automatically cleaned up.

## Troubleshooting

### Source parse

Verify the plugin source parses and ESM-links through bun's CLI:

```bash
bun -e 'import("./.opencode/plugin/index.ts").catch(e => { console.error(e); process.exit(1) })'
```

Bun's CLI honors the project's `tsconfig.json` while OpenCode's embedded runtime does not, so a pass here is necessary but not sufficient.

### Adaptor registration

Confirm OpenCode itself registers the Daytona adaptor:

```bash
OPENCODE_EXPERIMENTAL_WORKSPACES=true DAYTONA_API_KEY=x opencode serve --port 4096 >/dev/null 2>&1 &
sleep 4
curl -s http://127.0.0.1:4096/experimental/workspace/adapter | grep -q daytona && echo OK || echo FAIL
kill %1 2>/dev/null
```

Starts a headless OpenCode server, queries `/experimental/workspace/adapter`, and prints `OK` if the `daytona` type appears in the response.

### Latest log

Grep the most recent OpenCode log for plugin-loading errors:

```bash
ls -t ~/.local/share/opencode/log/*.log | head -1 | xargs grep -E "ERROR|@daytona|opencode/plugin/index"
```

## Migrating from v0.167.0

Version 0.168.0 rewrites the plugin to use OpenCode's workspace adaptor API. This provides a simpler architecture but removes some features from the original implementation:

- Bidirectional git sync between local and sandbox
- Auto-commit on session idle
- Custom tool implementations (bash, edit, grep, etc.)

If you need these features, pin to the last version that includes them:

```json
{
  "dependencies": {
    "@daytona/opencode": "0.167.0"
  }
}
```

> **Note:** v0.167.0 will not receive updates and may break with future OpenCode versions.

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

### Running Tests

Run the test suite:

```bash
npx nx run opencode-plugin:test
```

Or directly with bun:

```bash
cd libs/opencode-plugin
bun test --timeout 180000
```

Set `DAYTONA_API_KEY` to run the full test suite including sandbox cleanup tests.

### Local Development

To test the plugin locally, create a symlink in your test project:

```bash
mkdir /tmp/myproject && cd /tmp/myproject
ln -s [ABSOLUTE_PATH_TO_DAYTONA]/libs/opencode-plugin/.opencode .opencode
git init
OPENCODE_EXPERIMENTAL_WORKSPACES=true opencode
```

> **Note:** When developing locally with a symlink, OpenCode loads the TypeScript source directly, so no build step is required.

### Running against a local OpenCode build

To test against a from-source OpenCode checkout (e.g. `~/opencode`) instead of the installed binary, run its `dev` script with the test project (`/tmp/myproject` from above) as a trailing argument:

```bash
cd ~/opencode
OPENCODE_EXPERIMENTAL_WORKSPACES=true bun dev /tmp/myproject
```

### Building

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
│       │   ├── index.ts
│       │   └── instructions.ts
│       └── index.ts
├── package.json
├── project.json
├── tsconfig.json
├── tsconfig.lib.json
└── README.md
```

## License

Apache-2.0
