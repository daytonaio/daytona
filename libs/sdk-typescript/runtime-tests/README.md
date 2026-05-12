# Runtime compatibility tests

End-to-end smoke tests that prove `@daytona/sdk` builds and runs in every supported JavaScript runtime. Each subdirectory is a self-contained mini-project that:

1. Installs the locally-built SDK (via `npm pack` tarball)
2. Builds (when the runtime requires bundling)
3. Exercises `Image.base().env({...})` and `daytona.list()` against a real Daytona API
4. Asserts the responses are well-formed

## Running locally

Build the SDK first, then run the orchestrator:

```bash
yarn nx build sdk-typescript

DAYTONA_API_KEY=<your-key> \
DAYTONA_API_URL=https://app.daytona.io/api \
  bash libs/sdk-typescript/runtime-tests/run-all.sh
```

To run a subset:

```bash
ONLY=node-cjs,node-esm,bun bash libs/sdk-typescript/runtime-tests/run-all.sh
```

The orchestrator iterates every subdirectory that contains a `run.sh`, packs the local SDK build, and runs each test in isolation. It exits 0 only if every runtime passes.

## How each runtime is exercised

| Runtime | What runs | Engine |
| --- | --- | --- |
| `node-cjs` | `node test.js` | Node.js |
| `node-esm` | `node test.mjs` | Node.js |
| `bun` | `bun run test.ts` | Bun (real interpreter) |
| `deno` | `deno run test.ts` | Deno (real interpreter) |
| `vite-ssr` | `vite build --ssr` then `node` executes the bundle | Node.js |
| `vite-browser` | `vite build` + `vite preview` + Playwright drives **headless Chromium** to load the bundle and assert | Real browser |
| `nextjs` | `next build` + `next start` + `curl` hits the API route | Production Next.js server |
| `nuxt` | `nuxt build` + `node .output/server/index.mjs` + `curl` | Production Nitro server |
| `remix` | `remix vite:build` + `remix-serve` + `curl` | Production Remix server |
| `cloudflare-workers` | `wrangler dev --local` + `curl` | Real `workerd` (Cloudflare's actual Workers runtime) |
| `aws-lambda` | `esbuild` bundle + [`lambda-local`](https://github.com/ashiina/lambda-local) (real Lambda event/context shapes, env vars, timeout) | Lambda emulator |
| `azure-functions` | `func start` (Azure Functions Core Tools host) + `curl` | Real Functions host |

## Required external tools

Most runtimes ship as npm devDeps. These external binaries are required for specific tests, which skip gracefully when missing:

| Binary | Used by | Install |
| --- | --- | --- |
| `bun` | `bun` | `curl -fsSL https://bun.sh/install \| bash` |
| `deno` | `deno` | `curl -fsSL https://deno.land/install.sh \| sh` |
| `func` | `azure-functions` | `npm i -g azure-functions-core-tools@4` |

Playwright auto-installs Chromium on first run for `vite-browser`. On Linux CI, system libs may be needed: `npx playwright install --with-deps chromium`.

Everything else (`wrangler`, `lambda-local`, `vite`, `next`, `nuxt`, `remix`, `esbuild`, `typescript`, `playwright`, `@azure/functions`) is an npm devDep installed automatically per-runtime.

## Adding a new runtime

1. Create `runtime-tests/<runtime-name>/`
2. Add a minimal `package.json` (no `@daytona/sdk` in deps — the orchestrator injects it via tarball)
3. Add a test file that imports `@daytona/sdk`, exercises `Image.base()` + `daytona.list()`, and prints `PASS` / exits non-zero on failure
4. Add `run.sh` that installs deps, runs `npm install --silent "$SDK_TARBALL"`, builds (if needed), and runs the test
5. `chmod +x run.sh`

The orchestrator auto-discovers any directory with a `run.sh`.

## CI

These tests run in `.github/workflows/e2e_pr_tests.yaml` after the SDK Jest e2e suite, gated by the same `nx affected` check (only when `sdk-typescript`, `api-client`, or `toolbox-api-client` change). They use the local Daytona stack the workflow already brings up.
