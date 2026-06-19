# Daytona Sandbox Extension for Pi

This is a Pi extension that runs every Pi tool call inside a Daytona sandbox. The agent runs on your machine, while `bash`, file I/O, and search execute in a remote sandbox that is created when you launch Pi with `--daytona`, kept with your session (and reattached when you resume it), and deleted when you delete that session.

## Features

- Runs Pi's tool calls in an isolated Daytona sandbox while the agent stays on your machine
- Clones the repo you're in (or a `--repo` you pass) into the sandbox automatically
- Syncs each session to its own GitHub branch — the agent commits, the extension pushes
- Keeps one sandbox per session and reattaches it when you resume
- Generates live preview links when a server starts in the sandbox

## Usage

### Installation

First, install Pi:

```bash
npm install -g @earendil-works/pi-coding-agent
```

See [pi.dev](https://pi.dev) for other install options.

Then add the Daytona extension to Pi:

```bash
pi install npm:@daytona/pi
```

> [!NOTE]
> To update the extension later, run `pi update` — `pi install` won't refresh an existing install.

### Environment Configuration

This extension requires a [Daytona account](https://www.daytona.io/) and [Daytona API key](https://app.daytona.io/dashboard/keys) to create sandboxes.

Set your Daytona API key as an environment variable (e.g. in your shell profile):

```bash
export DAYTONA_API_KEY="your-api-key"
```

The extension also reads these optional variables:

- `DAYTONA_API_URL` — defaults to `https://app.daytona.io/api`.
- `DAYTONA_TARGET` — e.g. `us`.

If no key is set and a UI is available, Pi prompts you for one once per session.

### Running Pi

Run Pi from inside a git repository:

```bash
cd my-project
pi --daytona
```

The extension clones the repo you're in into the sandbox and syncs your work to a GitHub branch (see [GitHub branch sync](#github-branch-sync)).

Or point at a different repository:

```bash
pi --daytona --repo github.com/acme/api --branch dev
```

Or run outside a git repo to get a blank workspace.

#### Flags

| Flag                | Description                                                         |
| ------------------- | ------------------------------------------------------------------- |
| `--daytona`         | Run tools inside a Daytona sandbox                                  |
| `--repo <url>`      | Git repo to clone into the sandbox (defaults to the repo you're in) |
| `--branch <name>`   | Branch to clone (defaults to your current branch)                   |
| `--snapshot <name>` | Choose a Daytona snapshot / base image                              |
| `--public`          | Create a public sandbox so preview URLs need no token               |

#### Slash commands

While Pi is running with `--daytona`, you can manage the active sandbox:

- `/sandbox` — show the active sandbox's status: state, working directory, branch, sync status, and its GitHub branch link
- `/github` — open this session's branch on GitHub
- `/compare` — open this session's branch compare view on GitHub
- `/merge` — merge this session's branch into its base on GitHub
- `/pr` — open a GitHub pull request for this session's branch

## How It Works

The agent runs on your machine. Pi's tool layer is pluggable, so this extension substitutes Daytona-backed implementations of `bash`, `read`, `write`, `edit`, and `ls`, plus dedicated in-sandbox tools for `find` and `grep`. A footer badge is the always-visible signal that work is remote.

### Lifecycle

- **One sandbox per session, kept across runs.** A session's sandbox is recorded and **reattached** when you resume the session — your work and environment persist.
- **Idle pauses** the sandbox after 15 min (`autoStopInterval`, overridable with `--idle-stop <minutes>`). Its filesystem is preserved; resuming transparently restarts it.
- **Deleted when the session is.** When you delete a session from Pi's resume menu, its sandbox is reaped on the next Pi launch/exit (Pi has no session-deleted hook, so the extension reconciles live sessions against its sandboxes). There is no auto-delete timer — a sandbox lives until its session is gone.
- **In-memory sessions** (`--blank` / no session) can't be resumed, so their sandbox is deleted on exit.

### GitHub branch sync

If you're in a **github.com** repo and logged in via the GitHub CLI (`gh auth login`), each session gets its own branch and the agent's commits are pushed there automatically. The repo comes from `--repo`, or — when you omit it — is **detected from the git project you launched Pi in** (its `origin` and current branch).

- On start, the extension creates `pi/<short-session-id>` on GitHub (off your current branch, or `--branch`) and clones it into the sandbox over HTTPS.
- The agent **commits its own work** — it's prompted to commit after making changes, and not to push. After each turn the extension pushes those commits to the branch via the Daytona git API. A branch with nothing new is skipped.
- `/merge` merges the branch into its base, and `/pr` opens a pull request.
- **Forks** start a fresh sandbox and branch off the parent session's branch.

All network git operations (clone/push) run **inside the sandbox** through Daytona; the host only uses `gh` to mint a token and call the GitHub API. A temporary git identity is configured in the sandbox so commits work out of the box.

> [!NOTE]
> When you're not in a github.com repo (or `gh` isn't authenticated), push is disabled. The sandbox still gets a local git repo so the agent can commit, but nothing is pushed.

### Tools

| Tool                | What it does                                                                       |
| ------------------- | ---------------------------------------------------------------------------------- |
| `bash` (+ user `!`) | Run a command in the sandbox; backgrounded processes (`&`) don't hang the agent    |
| `read`              | Read a file from the sandbox                                                       |
| `write`             | Write a file to the sandbox                                                        |
| `edit`              | Edit a file (download → modify → upload; preserves Pi's exact-match semantics)     |
| `ls`                | List a sandbox directory                                                           |
| `find`              | Find files by glob inside the sandbox (gitignore-aware, supports path globs)       |
| `grep`              | Search file contents inside the sandbox                                            |
| `preview_url(port)` | Get a public preview URL for a port — the agent calls this after starting a server |

## Development

This extension is part of the Daytona monorepo.

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

To modify the extension, edit the source files in `libs/pi-extension`.

> [!NOTE]
> Because Pi loads extensions as TypeScript via [jiti](https://github.com/unjs/jiti), there is no build step — Pi runs the source directly.

#### Install dependencies

Install the extension's own dependencies once (needed for running it and for the tests):

```bash
cd libs/pi-extension && npm install
```

This is needed even after `yarn install` at the repo root, which doesn't make `@daytona/sdk` resolvable at runtime.

#### Run from source

Remove any previously installed copy (loading two copies makes every tool and flag conflict):

```bash
pi list                        # shows installed packages and their exact source
pi uninstall <source>          # e.g. npm:@daytona/pi — use the source shown by `pi list`
```

Install the local directory:

```bash
pi install ./libs/pi-extension    # add --local to scope it to the current project instead of globally
```

Run Pi:

```bash
DAYTONA_API_KEY=dtn_... pi --daytona
```

Edits to the source take effect on the next run — no reinstall needed.

Alternatively, load the source for a single run without installing:

```bash
DAYTONA_API_KEY=dtn_... pi -e ./libs/pi-extension/index.ts --daytona
```

#### Tests

```bash
npx nx run pi-extension:type-check   # type-check (from the repo root; needs the monorepo installed)

cd libs/pi-extension
npm run smoke                     # offline: load the extension and check it registers (no API key/network)
npm run test:live                 # end-to-end against real Daytona (needs DAYTONA_API_KEY)
```

### Publishing

Publish the TypeScript source to npm:

```bash
npx nx run pi-extension:publish
```

This will publish to npm with public access and use the version number from `package.json`.

## Project Structure

```
libs/pi-extension/
├── index.ts            # Extension entry point: flags, lifecycle, commands
├── src/                # Daytona-backed tool implementations
│   ├── tools.ts        # Tool registration (sandbox-backed tools + preview_url)
│   ├── auth.ts         # Daytona API key resolution
│   ├── sandbox.ts      # Sandbox resilience layer (auto-restart, exec)
│   ├── ops.ts          # Daytona-backed bash/read/write/edit/ls operations
│   ├── find-tool.ts    # In-sandbox find (ripgrep/find)
│   ├── grep-tool.ts    # In-sandbox grep (ripgrep/grep)
│   ├── github.ts       # Host gh control-plane (token + GitHub API)
│   ├── sync.ts         # Sandbox-side git push (Daytona git API)
│   └── util.ts         # Small shared helpers
├── scripts/            # Offline smoke + live integration tests
├── package.json        # Package metadata (includes the "pi" extensions field)
├── project.json        # Nx project configuration
├── tsconfig.json       # TypeScript config
└── README.md
```

## License

Apache-2.0
