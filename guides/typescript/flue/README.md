# Flue Bug-Fix Agent

## Overview

This example builds an autonomous bug-fix agent with [Flue](https://flueframework.com/) and [Daytona](https://www.daytona.io/). Given a GitHub issue, the agent:

1. Spins up a fresh Daytona sandbox.
2. Clones the target repository (your fork) into the sandbox.
3. Reads the issue and the relevant source code.
4. Writes a **failing test** that reproduces the bug.
5. Implements the minimal fix.
6. Runs the full test suite to verify the fix.
7. Pushes a new branch to your fork and opens a pull request via `gh`.

The whole TDD loop runs inside an isolated Daytona sandbox, so the agent can install dependencies and execute arbitrary code from any repository without ever touching your host machine. A sandbox is essential for this workflow: the agent clones unknown code, installs unknown dependencies, and executes the project's test suite — operations that need strict isolation from your host. Daytona provisions a fresh isolated environment for every run and tears it down on completion, so an untrusted repository can never affect your host.

## Features

- **Test-driven by construction.** The agent's skill enforces a strict Reproduce → Fix → Verify → PR sequence: no fix lands without a failing test that turns green.
- **Real pull requests.** The agent uses `gh` inside the sandbox to push a branch and open a real PR on your fork. You can review and merge in the GitHub UI.
- **Skill-driven logic.** The TDD workflow lives in markdown (`.agents/skills/bug-fix/SKILL.md`), not code. Tweak the workflow without changing TypeScript.
- **Structured output.** The agent returns a typed result (`prUrl`, `branch`, `filesChanged`, `summary`) you can pipe into downstream automation.
- **Sandbox-isolated execution.** Cloning, dependency installation, and test runs happen entirely inside Daytona, so your host stays clean.

## Prerequisites

- **Node.js 22+** (Flue requires it)
- A **Daytona** account (sign up at [app.daytona.io](https://app.daytona.io))
- An **Anthropic** API key (for the default model `claude-sonnet-4-6`)
- A **GitHub Personal Access Token** with `repo` scope
- A **fork** of a target repository to demo against

### Recommended demo target

Fork [`vercel/ms`](https://github.com/vercel/ms), the well-known millisecond conversion utility (5.5k+ stars). It's small (single ~244-line source file), uses Jest, has an MIT license, and ships with real open bug issues you can demo against.

```bash
gh repo fork vercel/ms --clone=false
```

The agent will clone _your fork_ (so the PR you open lands on your fork, not upstream).

## Environment Variables

Copy `.env.example` to `.env` and fill in:

| Variable | Required | Purpose |
|---|---|---|
| `DAYTONA_API_KEY` | yes | Get one from the [Daytona Dashboard](https://app.daytona.io/dashboard/keys) |
| `ANTHROPIC_API_KEY` | yes | For this agent's default model, `anthropic/claude-sonnet-4-6`. Required only if you don't override `MODEL`. (Flue itself has no default; our `bug-fix.ts` picks one.) |
| `GITHUB_TOKEN` | yes | Personal Access Token with `repo` scope. Create at [github.com/settings/tokens](https://github.com/settings/tokens) |
| `MODEL` | no | Override this agent's default. Any `provider/model-id` recognized by [`@mariozechner/pi-ai`](https://www.npmjs.com/package/@mariozechner/pi-ai). Examples: `anthropic/claude-opus-4-7`, `openai/gpt-5.5`, `openrouter/moonshotai/kimi-k2.6` |
| `DEMO_REPO` | no¹ | Default target fork in `<owner>/<repo>` form (e.g. `your-username/your-fork`). Used when the webhook payload omits `repo` |
| `DEMO_ISSUE` | no¹ | Default issue number (e.g. `284`). Used when the webhook payload omits `issueNumber` |
| `ISSUE_REPO` | no | Override the issue source, in `<owner>/<repo>` form. By default the agent auto-detects the upstream parent of `DEMO_REPO` via `gh repo view`; set this if `DEMO_REPO` is not a fork or you want to point at a different repo |

¹ Either set both `DEMO_REPO`/`DEMO_ISSUE` in `.env` and trigger with an empty payload, **or** omit them from `.env` and pass `repo` / `issueNumber` in every webhook payload. One or the other must be present.

> **Why two repos?** GitHub forks have **issues disabled by default** and don't carry over issues from the upstream. The agent therefore reads the issue from the fork's upstream parent (which it auto-detects), but pushes its branch and opens the PR against your fork. So you keep the demo isolated to your account: no spam to `vercel/ms` maintainers, no extra setup on your side.

```bash
cp .env.example .env
# edit .env with your values
```

## Getting Started

### 1. Install dependencies

```bash
npm install
```

### 2. Start the Flue dev server

```bash
npm run dev
```

You should see Flue's webhook server start on port `3583`:

```
Flue dev server listening on http://localhost:3583
  POST /agents/bug-fix/<session-id>
```

### 3. Trigger the agent

There are three ways to invoke the agent. Pick one.

**Option A: drive everything from `.env`** (simplest):

With `DEMO_REPO` and `DEMO_ISSUE` set in `.env`, fire an empty payload:

```bash
curl -X POST http://localhost:3583/agents/bug-fix/run-1 \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Option B: pass target per call** (override `.env`):

```bash
curl -X POST http://localhost:3583/agents/bug-fix/run-1 \
  -H "Content-Type: application/json" \
  -d '{
    "repo": "your-username/your-fork",
    "issueNumber": <number>
  }'
```

Replace `your-username/your-fork` with your fork's slug and `<number>` with the issue number you want to target.

The payload always wins over `.env`, so you can keep one default in `.env` and freely override per request when running multiple targets.

Watch the first terminal. You'll see the agent provision the sandbox, clone, write a failing test, fix the bug, run tests, and open a PR.

**Option C: one-shot with live tool tracing** (recommended when iterating or debugging):

Stop `flue dev` (or leave it; doesn't matter) and run:

```bash
npm run run
```

That maps to `flue run bug-fix --target node --id run-1 --env .env --payload '{}'` and does the following:

1. Builds a fresh `dist/server.mjs`.
2. Spawns its own ephemeral Node server on a random port.
3. POSTs to the agent's webhook with `Accept: text/event-stream`.
4. Streams every event (`tool:start`, `tool:done`, agent text deltas) decorated to **stderr** as readable lines.
5. Prints the final `result` JSON to **stdout**.
6. Tears the server down on completion.

Use this for the best introspection into what the LLM is doing tool-by-tool; use Options A/B when you want quiet sync mode against a long-running `flue dev` server.

> **What if the issue you target no longer exists?** If the issue you point the agent at has been closed, deleted, or never existed, the run will fail honestly instead of fabricating a fix. Two failure modes:
>
> - **Issue not found**: the setup phase's `gh issue view` returns a non-zero exit code and the agent throws before reaching the LLM. Pick a different issue and rerun.
> - **Issue is already fixed**: the LLM proceeds through Phase 1, then in Phase 2 writes a "failing" test that actually passes. The `test-driven-developer` role's hard rule (_a test that doesn't fail isn't a reproduction_) makes the agent stop and return early with `prUrl: ""` and a `summary` explaining the situation. No PR is opened.
>
> Update `DEMO_ISSUE` in `.env` (or pass `issueNumber` in the payload) and try another open issue.

## How It Works

```
                ┌─────────────────────────────────────────────────┐
                │            Flue dev server (host)               │
                │                                                 │
   POST /run ──▶│  bug-fix.ts orchestrator                        │
                │     │                                           │
                │     ├─ create Daytona sandbox (GH_TOKEN env)    │
                │     ├─ install gh + setup-git auth              │
                │     ├─ resolve token owner (gh api user)        │
                │     ├─ gh repo clone <user-fork>                │
                │     ├─ detect package manager from lockfile     │
                │     ├─ install deps (pnpm/yarn/bun/npm)         │
                │     ├─ resolve issue source (fork's upstream)   │
                │     ├─ fetch issue body via gh                  │
                │     ├─ upload SKILL.md into project sandbox     │
                │     └─ exclude .agents/ via .git/info/exclude   │
                │     │                                           │
                │     ▼                                           │
                │  session.skill('bug-fix', { args, role, ... })  │
                └─────┬───────────────────────────────────────────┘
                      │
                      ▼
                ┌──────────────────────────────────────┐
                │       Daytona sandbox (remote)       │
                │                                      │
                │  test-driven-developer role          │
                │     │                                │
                │     ├─ Phase 1: Understand           │
                │     ├─ Phase 2: Reproduce            │
                │     ├─ Phase 3: Fix                  │
                │     └─ Phase 4: Pull Request         │
                │                                      │
                │  returns { branch, prUrl,            │
                │            testFile, filesChanged,   │
                │            summary }                 │
                └──────────────────────────────────────┘
```

### Project layout

```
guides/typescript/flue/
├── .env.example
├── .gitignore
├── package.json
├── tsconfig.json
├── README.md                 # this file
├── .flue/
│   ├── agents/
│   │   └── bug-fix.ts        # orchestrator: sandbox + setup + skill invocation
│   ├── connectors/
│   │   └── daytona.ts        # adapts Daytona SDK to Flue's SandboxFactory
│   └── roles/
│       └── test-driven-developer.md   # subagent persona (TDD principles)
└── .agents/
    └── skills/
        └── bug-fix/
            └── SKILL.md      # the TDD workflow (markdown, not code)
```

The harness is intentionally minimal: a TypeScript orchestrator (`bug-fix.ts`), a Flue-spec connector (`daytona.ts`), one role markdown, and one skill markdown. There is **no `AGENTS.md`** at the guide root — Flue's runtime would prepend it to the LLM's system prompt for every agent call, but every guardrail it would contain (TDD discipline, minimal change, match host code style, etc.) is already covered by the `test-driven-developer` role and the `bug-fix` skill body. Adding `AGENTS.md` would be redundant and would also overwrite the target repository's `AGENTS.md` if it has one.

### Why split logic between `.ts` and `.md`?

Flue's design encourages keeping orchestration (sandbox lifecycle, payload validation, structured outputs) in TypeScript and **agent reasoning** in markdown. The `.ts` file is ~90 lines of plumbing; the actual TDD logic the LLM follows lives in the skill file and can be tuned without recompiling.

## Example Output

A successful run against `vercel/ms` issue #284 produces output like this in the `flue dev` terminal:

```
[bug-fix] target: your-username/your-fork#284 (model: anthropic/claude-sonnet-4-6)
[bug-fix] sandbox ready (id: a44a184e-cf0a-4407-bb1a-02f1b8000466)
[bug-fix] installing gh CLI in sandbox...
[bug-fix] commits will be authored as Your Name <12345+your-username@users.noreply.github.com>
[bug-fix] cloning your-username/your-fork into sandbox...
[bug-fix] detected package manager: pnpm
[bug-fix] installing pnpm...
[bug-fix] installing project dependencies...
[bug-fix] resolving issue source: vercel/ms
[bug-fix] fetching issue #284 from vercel/ms...
[bug-fix] uploading skill into sandbox + excluding it from git...
[bug-fix] running TDD workflow (reproduce → fix → PR)...
[bug-fix] PR opened: https://github.com/your-username/your-fork/pull/1
[bug-fix] branch: flue/fix-issue-284
[bug-fix] files changed: src/index.ts, src/parse.test.ts
[bug-fix] tearing down agents + sandbox...
```

The four-phase TDD work happens inside the LLM's session in the sandbox, so it doesn't surface line by line in the dev-server log. To stream those events live, use Option C above (`npm run run`), or hit the webhook with `Accept: text/event-stream` from any HTTP client.

The HTTP response body returned to your `curl` is the structured result the agent emits:

```json
{
  "result": {
    "branch": "flue/fix-issue-284",
    "prUrl": "https://github.com/your-username/your-fork/pull/1",
    "testFile": "src/parse.test.ts",
    "filesChanged": ["src/index.ts", "src/parse.test.ts"],
    "summary": "The parse() regex only matched plain decimal numbers in the value group (`-?\\d*\\.?\\d+`), so when format() produced scientific notation (e.g. `5.696545792019405e+297y`) via JavaScript's default number serialisation for very large Math.round() results, parse() returned NaN; the fix extends the value capture group with an optional exponent part (`(?:e[+-]?\\d+)?`) so scientific notation is accepted transparently."
  }
}
```

The PR (and the underlying commit) are authored under the GitHub account that owns your `GITHUB_TOKEN`. The agent resolves your login + numeric ID via `gh api user` at startup and sets `git config user.email` to the GitHub-recommended `<id>+<login>@users.noreply.github.com` noreply format, so the commit attaches to your profile.

## Common Issues

- **`gh: command not found` after install**: the agent installs `gh` if the sandbox image lacks it. If installation fails, switch to a Daytona snapshot that includes `gh` (or extend the install fallback in `bug-fix.ts`).
- **`Failed to clone` errors**: make sure the `repo` payload field points at _your fork_ (e.g. `your-username/your-fork`), not the upstream, and that your `GITHUB_TOKEN` is a classic PAT with the `repo` scope.
- **Test framework not detected**: the skill instructs the agent to read `package.json` and one existing test file. If the agent picks the wrong runner, tighten the Phase 1 instructions in `.agents/skills/bug-fix/SKILL.md`.
- **PR not opened**: check that your fork has push protection disabled for branches matching `flue/*`, and that the token user is the fork owner.

## Cleanup

The orchestrator wraps the entire two-agent flow in a `try { ... } finally { ... }` block. On both successful runs and failures, the `finally` block:

1. Destroys the **project** agent first (closes its session, no sandbox impact).
2. Destroys the **setup** agent next, which fires `cleanup: true`'s registered `sandbox.delete()` callback and tears the Daytona sandbox down.
3. As a fallback, if `init()` threw before the setup agent was created (so `cleanup: true` never armed), the orchestrator calls `sandbox.delete()` directly so we never leak a sandbox.

This is necessary because Flue **does not auto-destroy sessions on handler return** — sessions persist for resumability via the same `<id>`, and the registered cleanup callback only runs when `agent.destroy()` is explicitly called. Without the `try/finally`, the Daytona sandbox would stay up indefinitely.

Cleanup is best-effort: if `destroy()` or `sandbox.delete()` throws inside the `finally` block, the error is logged but the sandbox stays up. Verify in your [Daytona Dashboard](https://app.daytona.io/dashboard) after each run and remove any orphans manually if you see them.

## License

MIT. See the root project `LICENSE`.

## References

- [Flue Documentation](https://flueframework.com/)
- [Flue GitHub](https://github.com/withastro/flue)
- [Daytona Documentation](https://www.daytona.io/docs)
- [Daytona TypeScript SDK](https://www.daytona.io/docs/typescript-sdk/)
