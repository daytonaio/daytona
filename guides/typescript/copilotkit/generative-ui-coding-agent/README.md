# CopilotKit Generative-UI Coding Agent

## Overview

A [CopilotKit](https://docs.showcase.copilotkit.ai/) Built-in Agent backed by a [Daytona](https://www.daytona.io/) sandbox with full shell and filesystem access. The agent can build apps, debug or analyze code, run scripts, work with data, install packages. Every tool call streams into the chat as generative UI: shell commands render as terminal cards, file edits show up as syntax-highlighted code, directory listings and grep results render as structured cards, and any hosted process (dev server, static site preview, API) embeds as a live `<iframe>` in the message stream.

## Prerequisites

- **Node.js 20 or newer**
- A **Daytona API key**, from the [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- An **OpenAI API key**, from [platform.openai.com](https://platform.openai.com/api-keys)

## Setup

Clone the repository and switch into the guide:

```bash
git clone https://github.com/daytonaio/daytona.git
cd daytona/guides/typescript/copilotkit/generative-ui-coding-agent
```

Copy `.env.example` to `.env` and fill in your keys:

```bash
cp .env.example .env
```

```bash
DAYTONA_API_KEY=your_daytona_key
OPENAI_API_KEY=your_openai_key
```

Install dependencies:

```bash
npm install
```

## Run

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) and ask the agent for something. For example:

> Build the classic Snake game in Vite + React using HTML canvas. Use arrow keys to control the snake, count score, end on collision with wall or self, with a restart button. Make the game area dark green and the snake bright green.

The agent creates a sandbox, scaffolds the project (it chooses Vite, Next.js, plain HTML, whatever fits), installs dependencies, starts the dev server, and surfaces the preview as an iframe in the chat. Follow-up edits like "make it red themed" hot-reload the iframe in place. You can also ask it non-app things like "find every TODO comment under src/" or "write a Python script that downloads X and shows Y", and the same tool cards render the work.

## Tool surface

The backend in `app/api/copilotkit/route.ts` exposes 11 tools to the `BuiltInAgent`, all defined with `defineTool` from `@copilotkit/runtime/v2`:

| Tool | What it does | Daytona SDK call |
|---|---|---|
| `createSandbox({envVars?, labels?, autoStopInterval?})` | Create an ephemeral Daytona sandbox (auto-deletes on stop) with optional creation params | `daytona.create({ public: true, ephemeral: true, envVars, labels, autoStopInterval })` |
| `runCommand({sandboxId, command, background?})` | Run a shell command; `background:true` is for fire-and-forget processes (test/build watchers, log followers). Dev servers should use `startWebServer` instead | `sandbox.process.executeCommand` / `executeSessionCommand(..., {runAsync:true})` |
| `writeFile({sandboxId, path, content})` | Overwrite a file with the FULL new content | `sandbox.fs.uploadFile` |
| `readFile({sandboxId, path})` | Read a file's text content | `sandbox.fs.downloadFile` |
| `listFiles({sandboxId, path})` | Directory listing with metadata | `sandbox.fs.listFiles` |
| `findFiles({sandboxId, path, pattern})` | Grep file CONTENTS, returns file/line matches | `sandbox.fs.findFiles` |
| `searchFiles({sandboxId, path, pattern})` | Glob file NAMES, returns matching paths | `sandbox.fs.searchFiles` |
| `replaceInFiles({sandboxId, files, pattern, newValue})` | Codemod-style find-and-replace across multiple files | `sandbox.fs.replaceInFiles` |
| `getFileDetails({sandboxId, path})` | File metadata (size, mode, owner, modifiedAt) | `sandbox.fs.getFileDetails` |
| `startWebServer({sandboxId, command, port})` | Start a dev server in the background AND return its preview URL atomically. Polls the session logs for a ready signal (URL with the port, or a `ready / listening / listen / started / running / serving` phrase near the port) for up to 90 s. | `process.createSession` + `executeSessionCommand({runAsync:true})` + `getSessionCommandLogs` polling + `getPreviewLink` |
| `getPreviewUrl({sandboxId, port})` | Public preview URL for an already-open port; fallback for when `startWebServer` isn't the right shape | `sandbox.getPreviewLink` |

`createSandbox` creates an ephemeral sandbox (auto-deletes on stop, so chats don't leave stopped sandboxes accumulating in your Daytona account) and exposes the most commonly useful Daytona sandbox-creation params: `envVars`, `labels`, and `autoStopInterval` (idle-stop in minutes, default 15).

Filesystem operations like `mkdir`, `mv`, `rm`, and `chmod` are intentionally left to `runCommand` since they only need success/fail, not structured output.

## How it works

The system prompt tells the agent to call `createSandbox` once at session start, reuse the returned `sandboxId` across every subsequent tool call (creating a new one if the sandbox auto-deletes after a long pause), and work under `/home/daytona` by default. For web apps, it tells the agent to prefer Vite over the deprecated `create-react-app`, bind any dev server to `0.0.0.0` so the Daytona preview proxy can reach it, and hand it to `startWebServer` so the preview URL comes back in one shot.

On the React side, `app/page.tsx` registers `useRenderTool` per tool. Each renderer reads streaming `{ status, parameters, result }` and produces a card:

- `createSandbox` → sandbox card with the new `sandboxId`, env vars, labels, and auto-stop interval
- `runCommand` → terminal card (dark, monospace) with command + stdout
- `writeFile` / `readFile` → collapsible file card with `react-syntax-highlighter` (`vscDarkPlus`)
- `listFiles` / `searchFiles` → file-list card with directory entries or matching paths
- `findFiles` → grep card with `file:line:content` matches
- `replaceInFiles` → codemod card showing `pattern → newValue` and per-file success/fail
- `getFileDetails` → compact metadata card (type / size / mode / permissions / owner / modified)
- `startWebServer` / `getPreviewUrl` → live `<iframe>` (~520px tall) with reload + open-in-tab actions

The `PreviewCard` iframe stays mounted across subsequent turns (its React `key` only bumps when the user hits the reload button), so when the agent calls `writeFile` to update a file in a running dev server, the dev server's HMR picks up the change and the iframe content reloads in place.

## File-naming aside

Next.js App Router has special file conventions inside `app/`:

- `app/page.tsx` renders a UI page at that URL
- `app/layout.tsx` wraps children in a layout
- `app/api/copilotkit/route.ts` makes `/api/copilotkit` an API route handler exporting HTTP methods (`POST`, etc.)

## References

- [CopilotKit](https://docs.showcase.copilotkit.ai/)
- [Daytona](https://www.daytona.io/)
