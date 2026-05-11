# Sessions

The sessions product gives organizations warm, pre-provisioned sandboxes for low-latency code
execution (one-shot `code-run` and streaming `connect`), backed by an in-sandbox `session-daemon`.

Core pieces:

- `services/session.service.ts` — request entrypoints (`codeRun`, `connect`, `createSession`,
  transient sessions) and the API-internal daemon client.
- `services/session-repository.service.ts` — context-id → sandbox resolution (Redis-backed).
- `services/session-pool.service.ts` — the warm-sandbox fleet lifecycle.
- `services/session-scheduler.service.ts` + `services/session-load.service.ts` — instance selection
  and load tracking.
- `services/session-gc.service.ts` — idle/absolute TTL GC for contexts.

## Scale-out

The pool autoscales from one warm sandbox per `(org, template)` to a hybrid-autoscaled fleet that
distributes concurrent load and provisions/reaps sandboxes on demand. See
[docs/scale-out.md](./docs/scale-out.md) for the full design: request routing & stickiness, the load
model, cgroup/PSI methodology, the autoscale algorithm, the config reference, and the ops runbook.

## Bash (isolated) executions

`bash` is a first-class session language, run as a **`just-bash` virtual-interpreter isolate**: each
session is an in-process bash (grep/sed/awk/jq/pipes, no real binaries) over an OverlayFs that reads
the real `/workspace` but keeps writes private + ephemeral per isolate. Python and TypeScript isolates
can also shell out via a `bash()` builtin (in-process bridge for TS, stdio-RPC bridge for Python).
See [docs/bash-isolation.md](./docs/bash-isolation.md) for the full story: the threat model and the
two isolation modes (one sandbox per principal vs. shared-sandbox isolates), the survey of
alternatives (`just-bash`, bashkit, E2B, Modal, microsandbox), and **§9 for the implemented design**
(engine, daemon wiring, both bridges, and tests).
