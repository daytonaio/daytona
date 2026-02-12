# Daytona Sandbox Plugin for OpenCode

This is an OpenCode plugin that automatically runs OpenCode sessions in Daytona sandboxes. Each session has its own remote sandbox which is automatically synced to a local git branch.

## Features

- Securely isolate each OpenCode session in a sandbox environment
- Preserves sandbox environments indefinitely until the OpenCode session is deleted
- Generates live preview links when a server starts in the sandbox
- Synchronizes each OpenCode session to a local git branch

## Usage

### Installation

To add the plugin to a project, edit `opencode.json` in the project directory:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugins": ["@daytonaio/opencode"]
}
```

Now that the Daytona plugin is in the plugins list, it will automatically be downloaded when OpenCode starts.

To install the plugin globally, edit `~/.config/opencode/opencode.json`.

### Environment Configuration

This plugin requires a [Daytona account](https://www.daytona.io/) and [Daytona API key](https://app.daytona.io/dashboard/keys) to create sandboxes.

Set your Daytona API key and URL as environment variables:

```bash
export DAYTONA_API_KEY="your-api-key"
```

Or create a `.env` file in your project root:

```env
DAYTONA_API_KEY=your-api-key
```

### Running OpenCode

Before starting OpenCode, ensure that your project is a git repository:

```bash
git init
```

Now start OpenCode in your project using the OpenCode command:

```bash
opencode
```

To check that the plugin is working, type `pwd` in the chat. You should see a response like `/home/daytona/project`, and a toast notification that a new sandbox was created.

OpenCode will create new branches using the format `opencode/1`, `opencode/2`, etc. To work with these changes, use normal git commands in a separate terminal window. List branches:

```
git branch
```

Check out OpenCode's latest changes on your local system:

```
git checkout [branch]
```

To view live logs from the plugin for debugging, run this command in a separate terminal:

```bash
tail -f ~/.local/share/opencode/log/daytona.log
```

## How It Works

### File Synchronization

The plugin uses git to synchronize files between the sandbox and your local system. This happens automatically and in the background, keeping your copy of the code up-to-date without exposing your system to the agent.

#### Sandbox Setup

When a new Daytona sandbox is created:

1. The plugin looks for a git repository in the local directory. If none is found, file synchronization will be disabled.
2. In the sandbox, a parallel repository to the local repository is created in the sandbox. An `opencode` branch is created in the sandbox repository.
3. A new `sandbox` remote is added to the local repository using an SSH connection to the sandbox.
4. The `HEAD` of the local repository is pushed to `opencode`, and the sandbox repository is reset to match this initial state.
5. Each sandbox is assigned a unique incrementing branch number (1, 2, 3, etc.) that persists across sessions.

#### Synchronization

Each time the agent makes changes:

1. A new commit is created in the sandbox repository on the `opencode` branch.
2. The plugin pulls the latest commits from the sandbox remote into a unique local branch named `opencode/1`, `opencode/2`, etc. This keeps both environments in sync while isolating changes from different sandboxes in separate local branches.

The plugin only synchronizes changes from the sandbox to your system. To pass local changes to the agent, commit them to a local branch, and start a new OpenCode session with that branch checked out.

> [!CAUTION]
> When changes are synchronized to local `opencode` branches, any locally made changes will be overwritten.

### Session to sandbox mapping

The plugin keeps track of which sandbox belongs to each OpenCode project using local state files. This data is stored in a separate JSON file for each project:

- On macOS: `~/.local/share/opencode/storage/daytona/[projectid].json`.
- On Windows: `%LOCALAPPDATA%\opencode\storage\daytona\[projectid].json`.

Each JSON file contains the sandbox metadata for each session in the project, including when the sandbox was created, and when it was last used.

The plugin uses [XDG Base Directory](https://specifications.freedesktop.org/basedir/latest/) specifical to resolve the path to this directory, using the convention [set by OpenCode](https://github.com/anomalyco/opencode/blob/052f887a9a7aaf79d9f1a560f9b686d59faa8348/packages/opencode/src/global/index.ts#L4).

## Development

This plugin is part of the Daytona monorepo.

### Setup

First, clone the Daytona monorepo:

```bash
git clone https://github.com/daytonaio/daytona
cd daytona
```

Install dependencies:

```bash
yarn install
```

### Development and Testing

To modify the plugin, edit the source code files in `libs/opencode-plugin/.opencode`.

To test the OpenCode plugin, create a test project to run OpenCode in:

```bash
mkdir ~/myproject
```

Add a symlink from the project directory to the plugin source code:

```
ln -s libs/opencode-plugin/.opencode ~/myproject
```

Start OpenCode in the project directory:

```bash
cd ~/myproject && opencode
```

Use the instructions from [Running OpenCode](#running-opencode) above to check that the plugin is running and view live logs for debugging.

> [!NOTE]
> When developing locally with a symlink, OpenCode loads the TypeScript source directly, so no build step is required.

### Building

Build the plugin:

```bash
npx nx run opencode-plugin:build
```

This compiles the TypeScript source files in `.opencode/` to JavaScript in `dist/.opencode/`.

### Publishing

Log into npm:

```bash
npm login
```

Publish the compiled JavaScript package to npm:

```bash
npx nx run opencode-plugin:publish
```

This will publish to npm with public access and use the version number from `package.json`.

## Project Structure

```
libs/opencode-plugin/
├── .opencode/                     # Source TypeScript files
│   ├── plugin/
│   │   ├── daytona/               # Main Daytona integration
│   │   │   └── ...
│   │   └── index.ts               # Plugin entry point
├── dist/                          # Build output
│   └── .opencode/                 # Compiled JavaScript files
├── package.json                   # Package metadata (includes main/types)
├── project.json                   # Nx build configuration
├── tsconfig.json                  # TypeScript config
├── tsconfig.build.json            # TypeScript config
└── README.md
```

## License

Apache-2.0
