# Daytona Sessions

**Sub-second, stateful code execution without provisioning a sandbox.**

Sessions are Daytona's lightweight code-execution primitive: the SDK calls
`daytona.session.run(code, ...)` and the API takes care of routing the call to
a warm in-cluster sandbox running an embedded interpreter. There is no
`create()`, no handle, no image pull, no port wiring. A one-shot Python call
returns in under 50 ms on first use and ~25–40 ms steady-state against a warm
pool (see [Performance characteristics](#performance-characteristics) for the
breakdown). A persistent context keeps the same CPython process alive across
N calls so variables, imports, open files, and loaded models survive between
turns.

This document is the engineering overview for the feature. It covers what
sessions are, the components that implement them, the request flows, and
the operational characteristics.

- Python example: [`examples/python/sessions/main.py`](../../examples/python/sessions/main.py)
- TypeScript example: [`examples/typescript/sessions/index.ts`](../../examples/typescript/sessions/index.ts)
- Performance comparison vs classic sandbox: [`examples/python/sessions/vs_sandbox.py`](../../examples/python/sessions/vs_sandbox.py)

---

## Why this exists

Building "many parallel isolated code executions" on top of classic sandboxes
is doable but not cheap. The pieces a caller would otherwise have to write
themselves:

- **A pool of sandboxes** sized to peak concurrency, with provisioning,
  health checks, idle reclamation, and replacement of failed instances.
- **A scheduler** that decides which sandbox a given execution lands on
  and what happens when the pool is exhausted (queue, scale up, reject).
- **Per-context isolation across executions.** The classic-sandbox
  `process.session` API can run consecutive commands in the same shell, but
  it doesn't give you an isolation boundary between contexts — one user's
  state can leak into the next. Getting clean isolation means provisioning
  a fresh sandbox per context, which then re-raises the lifecycle problem
  at higher granularity.
- **Cleanup.** Tracking which sandboxes are still in use, which can be
  reaped, and which leaked because a caller died mid-execution.

None of that is novel work, but it's a meaningful amount of operational
plumbing for what the caller usually thinks of as "run this code." And
the naive version — one sandbox per concurrent execution — has poor
utilization: most sandboxes sit idle most of the time, and provisioning a
fresh one is multi-second.

**Sessions move that plumbing into the platform.** A warm sandbox per
`(organization, template)` holds many concurrent interpreter sessions,
each with its own heap. The API tracks instances, the daemon enforces
per-session isolation at the interpreter level (CPython subprocess or V8
isolate), and the pool grows / recycles itself based on demand. From the
caller's perspective there is just `daytona.session.run(...)` — no pool
to size, no scheduler to write, no idle sandboxes to reap, no per-context
provisioning cost.

The trade-off is real and explicit: sessions don't give you the full
sandbox surface (filesystem, shell, custom images) — that's still what
classic sandboxes are for. They give you fast, isolated code execution
at high concurrency without the orchestration burden.

---

## What it is

A single API call, two execution modes:

```python
# One-shot: no sandbox, no handle, no setup.
daytona.session.run("print(2 ** 10)", SessionRunOptions(language="python"))

# Persistent context: state survives across runs.
ctx = daytona.session.create_context(language="python")
daytona.session.run("import pandas as pd; df = load()", context=ctx)
daytona.session.run("df = df[df.amount > 100]",         context=ctx)
daytona.session.run("result = df.groupby('user').sum()", context=ctx)
```

Both modes run in a **real CPython 3.11 interpreter** (or the V8 ESM module
loader for TypeScript) inside a Daytona sandbox provisioned from a small,
shared, in-cluster warm pool. The sandbox is invisible to the caller — the
API tracks instances internally, and the SDK talks directly to the in-sandbox
daemon over a signed WebSocket once the first call has primed the bundle.

### What you get out of the box

- **Real CPython + pip ecosystem** — `numpy`, `pandas`, `polars`, `pyarrow`,
  `openai`, anything with native C extensions. Not Pyodide / WASM.
- **TypeScript with real ESM** — `import` from a curated registry,
  top-level `await`, the last expression surfaces as a `display` frame.
- **Stateful contexts** — Python globals (and TS `globalThis`) survive across
  `run()` calls in the same context.
- **Multiple concurrent contexts** per organization, each with its own
  interpreter heap.
- **Streaming** — `run_stream()` opens a WS and invokes `on_stdout` /
  `on_display` / `on_error` per frame, with the same aggregated return value
  shape as `run()`.
- **Structured errors** — user `1 / 0` comes back as an `error` frame with
  `name`, `value`, and `traceback`, not a thrown SDK exception.

### What it is _not_

- **Not a long-running container.** The user never owns a sandbox — pool
  instances are shared across the org and recycled on idle / drift.
- **Not a Linux shell.** No `exec`, no file system you carry around, no
  port forwarding. For those, use the [classic sandbox](#vs-classic-sandboxes-sandboxprocesscode_run).
- **Not automatically region-routed.** A context binds to one sandbox in
  one region; the API doesn't (yet) pick the closest region to the SDK
  caller. See [trade-offs](#trade-offs) below for the RTT consequence.

---

## Key concepts

| Term | What it is | Persisted |
|---|---|---|
| **Template** | Named runtime profile — image, supported languages, default packages. Resolved by name (`python-default` by default). | Yes (`session_template`) |
| **Instance** | A warm sandbox bound to `(organization, template)`. Created lazily on first use. Exactly one READY instance per pair. | Yes (`session_instance`) |
| **Session** | An interpreter session inside an instance. Has an id, a language, and (for persistent contexts) a DB row with idle / absolute TTLs. | Yes (persistent only) |
| **Transient context** | A deterministic, _unpersisted_ per-(instance, language) context the API reuses for one-shot `run()` calls. Resets globals on each call. | No |
| **Access bundle** (`SessionAccessDto`) | Signed proxy URL + short-lived token the SDK uses to talk directly to the in-sandbox daemon. Refreshed lazily. | No (cache in SDK) |

The lifecycle invariant: `Template → Instance → Session`. A request that names
a context resolves that context's instance and bypasses the pool. A request
that doesn't carries `template`/`language`, the pool returns the org's warm
instance, and the API picks a transient context to dispatch into.

---

## Architecture

```
                        ┌──────────────────────────────────────┐
                        │                SDK                   │
                        │   (Python / TypeScript / ...)        │
                        │                                      │
                        │   daytona.session.run(code, ctx)     │
                        │   ────────────────────────────────   │
                        │   caches SessionAccess per (ctx)     │
                        │   (signed wsUrl + token, TTL ~5m)    │
                        └───────┬──────────────────────┬───────┘
                                │                      │
              first call /      │                      │  steady-state
              cache miss / 401  │                      │  run() / run_stream()
              (control plane)   │                      │  (WS, direct)
                                ▼                      ▼
       ╔════════════════════════════════════╗    ╔══════════════════════╗
       ║              API                   ║    ║        PROXY         ║
       ║       (NestJS, port 3001)          ║    ║     (port 4000)      ║
       ║                                    ║    ║  host-based routing  ║
       ║  POST   /sessions/transients       ║    ║  ws://<token>.<host> ║
       ║  GET    /sessions/:id/access       ║    ║      /execute        ║
       ║  POST   /sessions                  ║    ║                      ║
       ║  POST   /sessions/code-run (legacy)║    ╚══════════┬═══════════╝
       ║  POST   /sessions/connect  (legacy)║               │
       ║                                    ║               │
       ║  ┌──────────────────────────────┐  ║               │
       ║  │ SessionService               │  ║               │
       ║  │  • resolve template          │  ║               │
       ║  │  • pool.acquire(org, tpl)    │  ║               │
       ║  │  • build signed sandbox URL  │  ║               │
       ║  └──────────────────────────────┘  ║               │
       ║  ┌──────────────────────────────┐  ║               │
       ║  │ SessionPoolService           │  ║               │
       ║  │  • READY instance? fast-path │  ║               │
       ║  │  • else: createFromSnapshot  │  ║               │
       ║  │  • drift detection / rolling │  ║               │
       ║  └──────────────────────────────┘  ║               │
       ║  ┌──────────────────────────────┐  ║               │
       ║  │ SessionRepository            │  ║               │
       ║  │  • idle / absolute TTLs      │  ║               │
       ║  │  • mark INVALID on roll      │  ║               │
       ║  └──────────────────────────────┘  ║               │
       ║  ┌──────────────────────────────┐  ║               │
       ║  │ SessionGcService (cron)      │  ║               │
       ║  │  • sweep idle / expired      │  ║               │
       ║  └──────────────────────────────┘  ║               │
       ╚══════════════════╤═════════════════╝               │
                          │                                 │
              Postgres ◀──┤                                 │
              Redis    ◀──┤ (TypeORM cache, pool lock)      │
                          │                                 │
                          │ knob #3:                        │
                          │ management traffic              │
                          │ (POST/DELETE /sessions)         │
                          │ goes API → Runner directly,     │
                          │ via runner.apiUrl,              │
                          │ bypassing the public proxy      │
                          │                                 │
                          ▼                                 ▼
                  ╔═══════════════════════════════════════════════╗
                  ║                  RUNNER                       ║
                  ║              (Go, port 4000)                  ║
                  ║   picks runner by region, snapshot, score     ║
                  ╚═══════════════════════╤═══════════════════════╝
                                          │
                                          ▼
              ┌───────────────────────────────────────────────────┐
              │       SANDBOX (container, OCI runtime)            │
              │                                                   │
              │  ┌─────────────────────────────────────────────┐  │
              │  │  session-daemon  (Go, port 2281)            │  │
              │  │                                             │  │
              │  │  HTTP:  GET    /healthz                     │  │
              │  │         POST   /sessions                    │  │
              │  │         GET    /sessions                    │  │
              │  │         DELETE /sessions/:id                │  │
              │  │         GET    /packages                    │  │
              │  │  WS:    /sessions/:id/execute               │  │
              │  │                                             │  │
              │  │  Manager owns:                              │  │
              │  │   • One Worker per session                  │  │
              │  │   • Python factory: CPython subprocess      │  │
              │  │     (one pre-warmed pool per daemon)        │  │
              │  │   • TS factory: isolated-vm (V8) +          │  │
              │  │     Isolate.compileModule + ESM resolver    │  │
              │  │     against on-disk catalog                 │  │
              │  │   • Idle sweeper goroutine                  │  │
              │  └─────────────────────────────────────────────┘  │
              │                                                   │
              │  ┌──────────────┐    ┌──────────────────────────┐ │
              │  │   CPython    │    │  Node (esbuild +         │ │
              │  │   3.11       │    │  isolated-vm host)       │ │
              │  │   workers    │    │                          │ │
              │  └──────────────┘    └──────────────────────────┘ │
              └───────────────────────────────────────────────────┘
```

### Components

- **SDK (`libs/sdk-{python,typescript}`)** — surface API: `session.run`,
  `session.run_stream`, `session.create_context`, etc. Owns the
  `SessionAccess` cache (per-context and per-(template, language) for
  transients) and a 401/403-refresh-and-retry loop. Falls back to the
  REST-only endpoints `POST /sessions/code-run` and `POST /sessions/connect`
  if the direct-WS endpoints (`POST /sessions/transients`,
  `GET /sessions/:id/access`) return 404 — see [legacy endpoints](#legacy-endpoints).

- **API (`apps/api/src/session`)** — control plane. Resolves templates,
  manages the instance pool, persists contexts, mints signed access bundles.
  Talks to the daemon over a **proxy-bypass** internal URL (knob #3) for its
  own management calls. Public SDK traffic goes through the proxy.

- **Proxy (`apps/proxy`)** — host-based routing layer that maps
  `<token>.<host>/<port>/<path>` → `runner → sandbox → daemon`. Authenticates
  via the signed subdomain (not `Authorization: Bearer`). Same chain as
  `sandbox.process.code_run`.

- **Runner (`apps/runner`)** — picks the runner with the lowest load that
  already has the snapshot pulled. Calls the OCI runtime to start the sandbox
  container.

- **Sandbox image (`daytonaio/session-runtime:python-default-*`)** — base
  image carrying CPython 3.11, Node 22, esbuild, isolated-vm, and the
  curated package catalog. The `session-daemon` is the container entrypoint.

- **session-daemon (`apps/session-daemon`)** — Go service inside each warm
  sandbox. Owns a `Manager` with per-session `Worker`s. Python workers are
  pre-warmed CPython subprocesses (one shared pool per daemon, knob #2);
  TypeScript workers are V8 isolates that compile each `run()` as an ES
  module (`Isolate.compileModule`, the isolated-vm API) and resolve a
  curated set of `import`s from a pre-installed registry on disk.

---

## Request flows

### One-shot `run()` (no context)

```
SDK                  API                      Daemon (in sandbox)
 │  run(code)         │                              │
 │  cache miss        │                              │
 ├──POST /transients ▶│                              │
 │                    │  templates.resolve()         │
 │                    │  pool.acquire(org, tpl)      │
 │                    │  buildSandboxAccess()        │
 │                    │  ensureTransientContext() ──▶│  POST /sessions (idempotent)
 │                    │◀─ access bundle ─────────────│
 │◀── SessionDto (id=transient-<inst>-<lang>, access={wsUrl,token,expiresAt})
 │                    │                              │
 │  cache access      │                              │
 ├────────────────────WS /sessions/:id/execute──────▶│  reset=true, code=...
 │                                                   │  worker rebuilds globals
 │◀─ stdout frames ──────────────────────────────────│
 │◀─ display frame ──────────────────────────────────│
 │◀─ control: done ──────────────────────────────────│
```

**Latency budget (warm pool, localhost):** under 50 ms for the first call
(the `POST /transients` round-trip + WS handshake + first exec); ~25–40 ms
steady state. The steady-state cost decomposes into a WS handshake per
call + one WS round-trip + interpreter dispatch — see [Performance
characteristics](#performance-characteristics) below. The API is _only_
on the first-call path; subsequent calls hit the proxy + daemon directly.

### Persistent context `create_context()` + `run(context=ctx)`

```
SDK                       API
 │  create_context(lang)   │
 ├──POST /sessions ───────▶│
 │                         │  templates.resolve()
 │                         │  pool.acquire(org, tpl)
 │                         │  contexts.create() → row in DB
 │                         │  daemon POST /sessions (best effort)
 │                         │  buildSandboxAccess()
 │◀── SessionDto {id, access, expiresAt} ───┤
 │  cache access by ctx.id │
 │                         │
 │  run(code, context=ctx) │
 ├─── WS exec ─────────────────────────────────────▶ (proxy → daemon)
 │  (reset=false, globals survive between runs)
```

Persistent contexts have idle TTL (default 30 min) and absolute TTL
(default 24 h). The `SessionGcService` cron sweeps expired rows;
the daemon's idle sweeper independently tears down dormant workers. When
the underlying sandbox is rolled (drift, death), the pool marks all
dependent contexts `INVALID` so the SDK sees a clean
`SessionInvalidatedError` rather than getting silently routed to a
fresh process with empty globals.

### Streaming `run_stream()`

Same WS path as `run()`; the SDK invokes per-frame handlers
(`on_stdout`, `on_display`, `on_error`) as frames arrive, _and_ aggregates
them into the same `SessionRunResult` shape `run()` returns. The two are
interchangeable from the caller's perspective.

Note: a streaming exec naturally keeps the WS open for the lifetime of
the run, so it doesn't pay the per-call handshake that dominates short
`run()`'s steady-state floor. Pooling WS across multiple sequential
`run()` calls (see [future work](#future-work) item #1) would extend the
same property to non-streaming calls.

---

## Performance characteristics

Numbers below are localhost / single-machine `vs_sandbox.py` runs.
Re-run [`vs_sandbox.py`](../../examples/python/sessions/vs_sandbox.py) on
your own deploy for ground-truth — the values here are representative
ranges, not committed SLOs. Cross-region traffic adds RTT to every leg.

| Metric | Classic sandbox | Session (one-shot) | Session (persistent ctx) |
|---|---|---|---|
| Time to first exec (warm pool) | ~250–300 ms (`create()` always blocks) | < 50 ms | < 50 ms |
| Time to first exec (cold-cold, pool eviction) | ~250–300 ms (same — `create()` doesn't benefit from a session-side warm pool) | ~2,000–2,500 ms (full sandbox provision) | ~2,000–2,500 ms |
| Steady-state per call | ~5 ms (`code_run`)¹ | ~25–40 ms | ~25–40 ms |
| Survives state across calls | yes (long-lived container) | no — globals reset each call | **yes — Python globals / `globalThis` retained** |

¹ Classic-sandbox `code_run` is only this fast _after_ a sandbox already
exists; every workflow pays the ~250–300 ms `create()` once. Sessions
amortize that creation cost across the org's warm pool, but pay a higher
per-call floor (proxy hop + per-call WS handshake). The two rows are
comparable end-to-end once you fold the classic `create()` into a single
workflow's first call.

The steady-state floor on sessions is dominated by **two** components:
the WS round-trip through the proxy (~10–20 ms localhost, +RTT cross-region)
**plus a fresh WS handshake per call** because the SDK opens and closes a
new WS for every `run()`. Interpreter dispatch itself is < 1 ms once the
[CPython pre-warmed pool](#knob-2--pre-warmed-cpython-pool) (knob #2)
removed the ~300 ms per-call subprocess spawn. Pooling WS per
`(session_id, access)` is on the roadmap and would push steady-state
closer to ~10 ms — see [future work](#future-work) item #1; this is the
single largest near-term win.

### Latency optimizations actually in the build today

The performance characteristics above are the result of four deliberate
choices, all in the codebase:

#### Knob 1 — Transient context reuse

One-shot `run()` calls all dispatch into a single deterministic
`transient-<instance>-<language>` context per (instance, language) instead of
creating + tearing down a UUID context per call. The WS frame carries
`reset: true` so the worker rebuilds globals — same isolation semantics, no
subprocess spawn / V8 context teardown. About ~14× faster for Python
one-shot than the naive create/delete-per-call model.

#### Knob 2 — Pre-warmed CPython pool

The Python `WorkerFactory` keeps a small pool of long-lived CPython
subprocesses inside the daemon. New contexts attach to a pool member instead
of forking a fresh `python3` per request. This is the difference between
~300 ms per call and ~25 ms.

#### Knob 3 — Direct API→runner for management traffic

The API talks to the daemon over `runner.apiUrl/sandboxes/<id>/toolbox/proxy/<port>`
(internal) for its own `POST /sessions` / `DELETE /sessions` calls, bypassing
the public proxy on port 4000. Saves one local TCP hop + one signed-URL
mint + one auth round-trip per management call.

#### Knob 4 — SDK direct-to-sandbox WS path

The SDK caches an `SessionAccess` bundle (signed proxy URL + token) per
context and per (template, language) for transients, and opens the WS
**directly to the proxy** for `run()` / `run_stream()`. The API is only on
the first-call (cache miss) / refresh / recovery paths. Subsequent
steady-state calls don't transit the API at all.

---

## Where it fits

### Vs. classic sandboxes (`sandbox.process.code_run`)

Use **classic sandboxes** when you need:

- A handle the user "owns" for the duration of a session (files persist,
  shell history, port forwards, mounted volumes).
- `exec()` for arbitrary shell commands, not just `code_run`.
- Workloads that don't fit in a shared warm pool (custom images, heavy GPU,
  long-running daemons).
- Anything the user might inspect via SSH or the dashboard.

Use **sessions** when you need:

- Fast time-to-first-execution without a `create()` step.
- Notebook / REPL semantics with state across calls.
- Many parallel short-lived "tool" executions that don't deserve a sandbox
  each.
- A predictable, curated Python or TypeScript runtime (vs "whatever image
  the user picked").

The two coexist: an agent can hold a classic sandbox for filesystem work
_and_ fire session `run()` calls for stateless evaluation.

### Vs. Cloudflare Workers / Dynamic Workers / Workers Python

A summary of what each model is shaped for; full comparison lives in
[`/.cursor/plans/`](../../.cursor/plans) and in the analysis written
alongside `examples/python/sessions/vs_sandbox.py`.

| | Stateless edge dispatch | Stateful multi-call REPL | Real Python ecosystem |
|---|---|---|---|
| **Daytona Session** | < 50 ms cold, ~25–40 ms warm, single-region | **yes — interpreter stays live** | **yes — CPython + pip** |
| CF Dynamic Workers (JS/TS) | ~5 ms cold, sub-ms warm, edge-local | no — fresh session per request | n/a |
| CF Workers Python (Pyodide) | ~1,000 ms cold, tens of ms warm | no — fresh Pyodide per request | partial (Pyodide-only packages) |
| CF Durable Objects (+ Dynamic Worker for code) | ~5-10 ms within colo | yes, but region-pinned + JS only | n/a |
| CF Sandbox SDK (containers) | hundreds of ms cold | yes — but you write the in-container kernel | yes |

CF Sandbox SDK gives you a container substrate and a process-spawning
API — getting a real REPL out of it means writing the long-lived
multiplexing interpreter host yourself (subprocess pool, V8/Pyodide
hosting, WS dispatch, per-context isolation, stdout streaming). That
host is exactly what `session-daemon` ships. The trade-off is _which
substrate runs the kernel_: CF's containers + your kernel, or Daytona's
sandboxes + ours.

Cloudflare wins on raw cold-start for **stateless JS/TS at the edge**. For
**stateful, multi-call, real-CPython** workloads — what `stateful_context`
in `main.py` shows — Daytona session contexts are the more direct primitive
(no DO-pin trade-off you don't already pay, no Pyodide ceiling, no
in-sandbox-kernel to build yourself).

### Trade-offs

- **Region-pinned.** A context binds to one sandbox in one region. If your
  user is 150 ms away, every WS exec pays that on top of the ~25–40 ms
  in-cluster steady-state. Edge-distributed session contexts would require
  a separate architectural play (see [future work](#future-work) item #3).
- **Curated runtime.** Templates ship a fixed package catalog. Users can't
  `pip install` arbitrary packages mid-session today — that's reserved for
  classic sandboxes. (Custom templates per org are the planned escape hatch.)
- **Pool warmth ≠ guaranteed.** The first request after pool eviction pays
  the full ~2 s sandbox-provision cost. In steady multi-tenant operation
  the pool stays warm; for a quiet single-tenant deployment, expect
  occasional cold starts.

---

## Code map

```
apps/api/src/session/
├── controllers/session.controller.ts      ← REST surface (/sessions/*)
├── services/
│   ├── session.service.ts                 ← Facade: routes to template/pool/repository, builds access
│   ├── session-template.service.ts        ← Resolve name → SessionTemplate row (org > general)
│   ├── session-pool.service.ts            ← Warm-instance lifecycle, drift detection, reconcile cron
│   ├── session-repository.service.ts      ← Persistent session-row CRUD + TTL invariants;
│   │                                        marks dependent sessions INVALID on instance roll
│   └── session-gc.service.ts              ← Cron: idle / absolute TTL sweeper
├── entities/                              ← TypeORM rows for template / instance / session
├── dto/                                   ← DTOs: create-session, create-session-transient, code-run, connect, access
└── enums/                                 ← session-language / session-instance-state / session-state

apps/session-daemon/
├── cmd/main.go                            ← Entrypoint, in-sandbox HTTP+WS server
├── internal/server/server.go              ← Gin router; /healthz, /sessions CRUD, /packages, /sessions/:id/execute (WS)
├── internal/interpreter/
│   ├── manager.go                         ← Session registry; idle sweeper goroutine
│   ├── session.go                         ← Session struct: one Worker + FIFO exec queue per session
│   ├── worker.go                          ← Worker interface (Run, Reset, Close)
│   ├── wsclient.go                        ← /sessions/:id/execute WS frame plumbing
│   ├── python_subprocess.go               ← CPython worker (knob #2 pre-warmed pool)
│   ├── ts_host.go                         ← V8/isolated-vm host: compileModule + ESM resolver
│   └── types.go                           ← Wire types: CreateSessionRequest, OutputMessage
├── Dockerfile                             ← Multi-stage build of the session-runtime image (monorepo-aware, builds the daemon inline)
└── snapshot.Dockerfile                    ← Self-contained variant for the Daytona snapshot API path — see "Installing the snapshot via the Daytona API" below

libs/sdk-python/src/daytona/_sync/session.py     ← Sync SDK: caches access, WS-direct, retry/recovery
libs/sdk-python/src/daytona/_async/session.py    ← Async version (aiohttp + websockets)
libs/sdk-typescript/src/Session.ts                ← TS SDK: same shape; ensureAccess + runWsDirect
```

---

## Operational notes

### Templates

Seeded by migration `1778367241001-migration.ts` — registers a
`python-default` general template pointing at
`daytonaio/session-runtime:python-default-<version>`. Org-specific templates
(custom snapshot + curated packages) are coming; the resolver already
prefers org-scoped over general so the plumbing is in place.

### Installing the snapshot via the Daytona API

The seed migration registers the template row but ships an `imageName`
placeholder (`daytonaio/session-runtime:python-default-placeholder`) because
the actual runtime image has to be built and pushed into your Daytona's
internal registry. There are two supported ways to do that.

**Option A — via the Daytona snapshot API (no local Docker required).** This
is the path most operators want: you build the daemon binary somewhere with
internet access, host the binary at an HTTPS URL the build-runner can reach
(GitHub Release asset, presigned S3 URL, internal artifact store), and the
Daytona snapshot API builds the runtime image on-cluster from a single
self-contained Dockerfile.

```bash
nix develop .#go --command bash -c "VERSION=v0.0.0-dev nx run session-daemon:build-amd64 --configuration=production"

# 2. Host dist/apps/session-daemon-amd64 somewhere the runner can curl.
#    Examples (pick whichever fits your deploy):
#      gh release upload v0.0.0-dev dist/apps/session-daemon-amd64
#      aws s3 presign --expires-in 3600 s3://my-bucket/session-daemon-amd64
#    Capture the URL (and ideally a sha256).
export SESSION_DAEMON_URL=https://github.com/daytonaio/daytona/releases/download/v0.0.0-dev/session-daemon-amd64
export SESSION_DAEMON_SHA256=$(sha256sum dist/apps/session-daemon-amd64 | awk '{print $1}')

# 3. Create the snapshot via the Daytona API. This calls
#    `daytona snapshot create python-default --dockerfile ... --cpu 2 --memory 2 --disk 5`
#    under the hood, which POSTs the rendered Dockerfile to /api/snapshots and
#    streams the build logs until the snapshot reaches state=ACTIVE.
nix develop --command bash -c "nx run session-daemon:create-snapshot"
```

What this does end-to-end:

1. `render-snapshot-dockerfile` runs `envsubst` against
   [`apps/session-daemon/snapshot.Dockerfile`](snapshot.Dockerfile),
   substituting `${SESSION_DAEMON_URL}` and `${SESSION_DAEMON_SHA256}` into
   `dist/apps/session-daemon-snapshot/snapshot.rendered.Dockerfile`. This
   pre-rendering is necessary because the snapshot API's
   `buildInfo.dockerfileContent` field is a plain string — there's no
   equivalent of `docker --build-arg`.
2. `create-snapshot` shells out to the Daytona CLI, which uploads the
   rendered Dockerfile to `POST /api/snapshots`. There is no build context
   to upload (the Dockerfile has no `COPY` statements that reference local
   files) so `contextHashes` is empty.
3. The API picks a runner via the existing builder scheduling and emits a
   `BUILD_SNAPSHOT` job. The runner builds the image inside its sandbox
   builder, pushes the result to the internal registry, and updates the
   snapshot row's `imageName` to that registry-qualified tag plus
   `state=ACTIVE`.
4. The CLI streams the build logs and waits for `state=ACTIVE`.

Overridable env vars consumed by `create-snapshot`:

| Var | Default | Purpose |
|---|---|---|
| `SESSION_DAEMON_URL` | _(required)_ | HTTPS URL the build runner curls to fetch the prebuilt daemon binary |
| `SESSION_DAEMON_SHA256` | _(empty)_ | Optional sha256 digest; when set, the build aborts on mismatch |
| `SNAPSHOT_NAME` | `python-default` | Snapshot row name. Override to publish a versioned tag like `python-default-v0.3.1` |
| `SNAPSHOT_CPU` / `SNAPSHOT_MEMORY` / `SNAPSHOT_DISK` | `2` / `2` / `5` | Resource sizing baked into the snapshot row, used by the pool when spawning sandboxes |

The snapshot the API creates is per-org by default. To make it the new
`general=true` `python-default` snapshot that the seed migration's template
points at, an operator currently has to `UPDATE` the seeded row's
`imageName` to match the new snapshot's `imageName`. A dedicated
"promote-to-general" admin endpoint is on the operational backlog;
once it lands this README will switch to recommending it.

**Option B — via `docker build` (CI and local validation).** When you're
already in a build environment with the monorepo on disk and a working
Docker daemon, the multi-stage [`apps/session-daemon/Dockerfile`](Dockerfile)
is the path of least resistance: stage 1 builds the daemon from source,
stage 2 bakes it into the runtime image, and you push the result to whatever
registry your Daytona installation is configured to pull from.

```bash
docker build -f apps/session-daemon/Dockerfile \
  --build-arg VERSION=v0.0.0-dev \
  -t <your-registry>/daytonaio/session-runtime:python-default-v0.0.0-dev .
docker push <your-registry>/daytonaio/session-runtime:python-default-v0.0.0-dev
```

Then point the seeded snapshot row at the tag:

```sql
UPDATE snapshot
   SET "imageName" = '<your-registry>/daytonaio/session-runtime:python-default-v0.0.0-dev'
 WHERE name = 'python-default' AND general = true;
```

CI uses Option B (see [`.github/workflows/e2e_pr_tests.yaml`](../../.github/workflows/e2e_pr_tests.yaml)
— it builds the monorepo-aware Dockerfile, pushes to a local registry, and
patches the seed row). The CI also runs a render-only check on
`snapshot.Dockerfile` to make sure the Option-A path stays valid.

### Instance lifecycle

- One **READY** `session_instance` per `(organizationId, templateId)` at a
  time. Provisioning is serialized with a Redis lock so concurrent first
  calls don't race.
- The pool's reconcile cron (every 30 s) detects dead / destroyed sandboxes
  and snapshot drift (template's `snapshotId` no longer matches the
  instance's). On drift the instance rolls and all dependent contexts are
  marked `INVALID`.
- The `SESSION_DAEMON_API_IDLE_TTL_SECONDS_HINT` env var is propagated into
  the sandbox at create-time so the daemon can compare its own
  `SESSION_DAEMON_IDLE_TTL_SECONDS` against `1.5×` that hint and log a
  startup warning on inverted ratios (contexts being reaped before the API
  expects). The earlier plan put this on `GET /healthz` as a degraded
  indicator; the runtime check was dropped in favour of the boot-time
  warning since the ratio is a deployment-config invariant, not a
  runtime-drifting signal. `/healthz` today is a binary liveness probe
  (`{"ok": true}`), nothing more.

### Session lifecycle

- **Idle TTL** — default 30 min of no `run()` (configurable). Bumped on
  every `run()` and on `GET /access` (SDK keep-alive when streaming).
- **Absolute TTL** — default 24 h since `createdAt`.
- **Invalidation** — the underlying sandbox rolling or being destroyed
  immediately marks all dependent contexts `INVALID`. The SDK surfaces this
  as `SessionInvalidatedError`.

### Security model

- Public SDK → daemon path: signed proxy subdomain
  (`<token>.<proxy-host>/sessions/:id/execute`). Token is short-lived (TTL
  ~5 min; SDK refreshes lazily via `GET /access`).
- API → daemon path: internal `runner.apiUrl/sandboxes/.../toolbox/proxy/...`
  with `Authorization: Bearer <runnerApiKey>`. Never exposed externally.
- The daemon listens on `127.0.0.1:2281` inside the sandbox — the proxy
  chain is the only reachable surface.

### Legacy endpoints

`POST /sessions/code-run` and `POST /sessions/connect` are kept for
backward compatibility with API deployments that predate the
direct-to-sandbox WS path (knob #4). They route through the API on every
call — no `SessionAccess` cache, no direct WS to the proxy — so every
exec pays an extra API round-trip and signed-URL mint on top of the
proxy/WS hop the new path already pays. Steady-state is meaningfully
slower than the direct path (call it roughly 2× the new-path floor — but
re-run `vs_sandbox.py` against your deploy if you need a real number,
this hasn't been measured under the current build). The SDK falls back to
them automatically when `POST /sessions/transients` or
`GET /sessions/:id/access` return 404, so an SDK against an older API
still works (slower). These can be removed once all in-use API
deployments are running the version that ships `/transients` + `/:id/access`.

### Failure modes the SDK distinguishes

| Symptom | Error | What it means | Recovery |
|---|---|---|---|
| WS 401 / 403 on handshake | (internal) `_WSAuthError` | Token expired between cache fill and use | Refresh access, retry once |
| WS 400 / 404 / ECONNREFUSED on handshake | `SessionInvalidatedError` | The sandbox is gone or the context was rolled | Drop context, recreate |
| Daemon execution `error` frame | `result.error` (no exception) | User code raised | Surface to caller |
| API 410 with `reason='idle'` / `'absolute'` | `SessionExpiredError` | Session hit its TTL | Recreate |
| API 404 on `/transients` or `/:id/access` | (internal) `_LegacyFallback` | New endpoints not deployed on this API yet | Transparent fallback to `/code-run` / `/connect` — see [Legacy endpoints](#legacy-endpoints) |

---

## Future work

Things that aren't built yet but the architecture admits cleanly:

1. **Pooled WS sessions in the SDK.** Today each `run()` opens and closes a
   fresh WS to the proxy, so steady-state pays a per-call handshake on top
   of the round-trip. Holding one WS open per `(session_id, access)` would
   eliminate that handshake — pushing the ~25–40 ms steady-state range
   toward ~10 ms. This is the dominant near-term latency win and should be
   weighed against the call-pattern of typical agent code (many short
   independent `run()` calls, where the connection-reuse win is real but
   gated on the SDK keeping the WS alive between calls).
2. **Org-custom templates.** Resolver already prefers org-scoped; needs an
   admin path to upload a Dockerfile + package list and trigger a build.
3. **Cross-region routing.** Pool an instance per (org, template, region)
   and have the API pick the closest region to the SDK caller.
4. **`pip install` mid-session.** Persistent contexts could expose a
   `install_packages()` call that the daemon executes against an isolated
   per-context venv. Today this is a classic-sandbox primitive only.
5. **Multi-instance autoscaling per template.** Today one warm sandbox
   serves all concurrent sessions for a given `(organization, template)`;
   peak concurrency is bounded by what fits in that single sandbox
   (interpreter count, RAM, CPU). The pool service already keys on
   `(organization, template)` and the SDK already treats `instance.id` as
   opaque, so growing this to N instances behind a single template — with
   the scheduler picking the least-loaded — is an in-place extension. The
   moving pieces are: per-instance load tracking, a placement policy, and
   eviction rules for scaling back down. Mentioned in the motivation
   section as the headline post-v1 capability.
