# Claude Managed Agents on Daytona

Reference implementation for running [Claude Managed Agents](https://platform.claude.com/docs/en/managed-agents/overview) inside Daytona sandboxes as a self-hosted environment. See the [full guide](https://www.daytona.io/docs/en/guides/claude/claude-managed-agents) for the architecture story; this README covers how to run the reference.

## Prerequisites

- An Anthropic workspace with Claude Managed Agents self-hosted access. Create a self-hosted environment from the Claude Console under _Workspace → Environments → New → Self-hosted_, then click _Generate environment key_ on the environment page.
- An Anthropic API key for `create_agent.py` and for the application that creates sessions.
- A Daytona account and API key.
- For webhook mode only: an Anthropic webhook configured with `session.status_run_started` enabled (Claude Console → _Workspace → Manage → Webhooks_), plus its signing secret.

## Setup

```
python3.12 -m venv .venv
.venv/bin/pip install -e .
# Webhook mode also needs:
.venv/bin/pip install -e ".[webhook]"

cp .env.example .env
# Then fill in the values.
```

## Build the default snapshot

```
.venv/bin/python build_default_snapshot.py
```

Builds a Daytona snapshot from `Dockerfile.default`, naming it `byoc-env-default-{sha8}` from the Dockerfile's SHA-256. Idempotent: skips the build if a snapshot with the same hash already exists. The snapshot is provisioned with 2 vCPU / 8 GB memory / 10 GB disk — edit `build_default_snapshot.py` if you need different resources. The orchestrator uses this snapshot as the per-session sandbox image unless `session.metadata` overrides it (see below).

To use your own image, edit `Dockerfile.default` and rerun the script, or build any snapshot you like and pass its name as `daytona.snapshot_name` on `session.metadata`. Custom snapshots must include the runtime prerequisites the in-sandbox worker needs; see `Dockerfile.minimal` for the smallest viable example.

## Create an agent

```
.venv/bin/python create_agent.py my-agent
```

Prints an agent id. Use it when creating sessions.

The script creates a Sonnet 4.6 agent with the `bash`, `read`, `write`, `edit`, `glob`, `grep`, `web_fetch`, and `web_search` tools enabled, all set to `always_allow`. Edit `create_agent.py` to change the model, system prompt, or tool selection — or skip this script entirely and create your agent however you normally would.

## Run the orchestrator

Run exactly one orchestrator per `ENVIRONMENT_ID`. Both entrypoints enforce this with an exclusive `flock` on `/tmp/anthropic-selfhosted-orchestrator-{ENVIRONMENT_ID}.lock` (override the path with `ORCHESTRATOR_LOCK_FILE`) and fail fast if a second instance starts on the same host.

Pick polling when you don't want to expose an inbound endpoint or manage a webhook secret; pick webhook for lower per-event latency.

Both modes share the same sandbox lifecycle: ensure a labeled Daytona sandbox is running for each active session, start the in-sandbox runner, and let a janitor thread sweep idle and crashed sandboxes every minute.

### Polling mode

```
.venv/bin/python host_orchestrator_polling.py
```

No inbound network requirements. The main poll loop also covers crash recovery — replacing runners whose sandbox is still up but whose process exited — so the janitor leaves that case alone.

### Webhook mode

```
.venv/bin/python host_orchestrator_webhook.py
```

Listens on `:5051` and requires `ANTHROPIC_WEBHOOK_SECRET`. The server exposes `POST /` for Anthropic webhook deliveries and `GET /healthz` for liveness probes.

Anthropic requires the registered webhook endpoint to be a public HTTPS URL on port 443, so expose `:5051` through a TLS-terminating reverse proxy, load balancer, or tunnel. Register that public endpoint URL with Anthropic, not the raw orchestrator address.

In this mode the janitor thread also handles crash recovery: it polls the work queue to replace runners whose sandbox is still up but whose process exited.
Webhook-triggered drains are backed by a periodic safety-net drain loop, so a process restart after acknowledging a webhook cannot leave newly queued work waiting indefinitely for another webhook.

## Drive a session

You build the application that creates sessions and tails the event stream. A minimal example:

```py
import anthropic
client = anthropic.Anthropic()

session = client.beta.sessions.create(
    agent="agent_01...",
    environment_id="env_01...",
)

with client.beta.sessions.events.stream(session.id) as stream:
    client.beta.sessions.events.send(
        session.id,
        events=[{"type": "user.message",
                 "content": [{"type": "text", "text": "what python version is installed?"}]}],
    )
    for ev in stream:
        ...  # render events as they arrive
```

See [Events and streaming](https://platform.claude.com/docs/en/managed-agents/events-and-streaming) for the full event vocabulary.

## Per-session sandbox customisation

Two keys on `session.metadata` override the default sandbox source:

- `daytona.snapshot_name`: create this session's sandbox from a named Daytona snapshot instead of the default.
- `daytona.sandbox_id`: attach an already-prepared Daytona sandbox. Label it with `byoc.environment_id=<env>` and `byoc.mode=prepared` before passing the id. The orchestrator validates the labels, binds the session, starts the sandbox, and installs the runner.

The two keys are mutually exclusive. The full guide covers the prepared-sandbox flow in detail.

## Sandbox naming and labels

Each session's sandbox is named `byoc-{session_id}`, so you can find it directly in the Daytona dashboard from the session id. The orchestrator also writes a small set of `byoc.*` labels onto every sandbox it manages. The janitor uses these to scope its cleanup; you can use the same labels to filter sandboxes or to script your own tooling.

- `byoc.environment_id`: the self-hosted environment this sandbox belongs to.
- `byoc.session_id`: the Anthropic session id bound to this sandbox.
- `byoc.mode`: `in-sandbox` (orchestrator-owned), `prepared` (operator-attached via `daytona.sandbox_id`, until the orchestrator binds it and flips it to `in-sandbox`), or `terminal` (session has ended; the sandbox is being aged out).
- `byoc.work_id`: the currently-claimed work item id, cleared when the runner exits.
- `byoc.stopped_at`: ISO-8601 timestamp of when the sandbox was stopped or archived; drives the `MAX_IDLE_DAYS` deletion.

### Detaching a sandbox from the orchestrator

To take a sandbox out of management — for debugging, manual investigation, or any reason you want the janitor to leave it alone — change or remove its `byoc.environment_id` label. The orchestrator's Daytona list query filters on this label, so a sandbox without the matching value isn't returned to it at all. Restore the label when you want it managed again.

```py
from daytona import Daytona

dayt = Daytona()
sb = dayt.get("byoc-sess_01...")  # sandbox id from the Daytona dashboard
labels = dict(sb.labels or {})
labels.pop("byoc.environment_id", None)
sb.set_labels(labels)
```

## Files

- `Dockerfile.default` — default snapshot image. Mirrors the Claude Managed Agents [container reference](https://platform.claude.com/docs/en/managed-agents/cloud-containers).
- `Dockerfile.minimal` — smallest viable snapshot example for custom images.
- `build_default_snapshot.py` — builds the default snapshot in Daytona.
- `create_agent.py` — creates a long-lived agent.
- `sandbox_runner.py` — runs inside each Daytona sandbox. Wraps the SDK's `EnvironmentWorker.handle_item()`, which owns the session event stream for one residency.
- `host_lib.py` — host-side helpers for sandbox lifecycle and runner startup.
- `orchestrator_lib.py` — shared orchestration (work-queue draining, session locking, janitor thread).
- `host_orchestrator_polling.py` — long-polling entrypoint.
- `host_orchestrator_webhook.py` — FastAPI webhook entrypoint.

## Configuration

`.env.example` is the source of truth for environment variables.

Required:

- `ENVIRONMENT_ID` — the self-hosted environment id from the Console.
- `ANTHROPIC_ENVIRONMENT_KEY` — the environment-scoped key. Authenticates the orchestrator's work-queue and event-stream calls.
- `DAYTONA_API_KEY` — for the snapshot builder and the orchestrator.
- `ANTHROPIC_API_KEY` — for `create_agent.py` and for the application that creates sessions. Never read by the orchestrator.

Webhook mode also requires:

- `ANTHROPIC_WEBHOOK_SECRET` — the signing secret for the configured Anthropic webhook.

Optional tuning (see `.env.example` for defaults):

- `JANITOR_SECONDS`: how often the janitor sweeps labeled Daytona sandboxes.
- `RUNNER_MAX_IDLE_SECONDS`: SDK worker's idle window before it exits.
- `RUNNER_LAUNCH_PROBE_SECONDS`: delay between launching a runner and probing its state to confirm it started.
- `RUNNER_REPLACE_GRACE_SECONDS`: grace window for an existing runner to exit before a new launch on the same sandbox treats it as still live.
- `MAX_IDLE_DAYS`: delete stopped/archived sandboxes this long after they were stopped. `0` disables.
- `PORT`: webhook receiver port.
- `POLL_BLOCK_MS`, `POLL_RECLAIM_OLDER_THAN_MS`: long-poll timing for polling mode. `POLL_BLOCK_MS` must be in `1..999`.
- `WEBHOOK_DRAIN_SECONDS`, `WEBHOOK_RECLAIM_OLDER_THAN_MS`: safety-net drain timing for webhook mode.
- `LOG_LEVEL`: log level for the in-sandbox runner (default `INFO`).
- `DEFAULT_SNAPSHOT_NAME`: override the hash-derived default snapshot name.
- `ORCHESTRATOR_LOCK_FILE`: override the single-instance lock file path (default `/tmp/anthropic-selfhosted-orchestrator-{ENVIRONMENT_ID}.lock`).

## Production notes

In production, the orchestrator and the customer application would run as separate processes with separate credentials: the orchestrator never needs `ANTHROPIC_API_KEY`, and the application never needs `ANTHROPIC_ENVIRONMENT_KEY` or `DAYTONA_API_KEY`. The single `.env` here is for local-machine simplicity.

## See also

- [Full Claude Managed Agents on Daytona guide](https://www.daytona.io/docs/en/guides/claude/claude-managed-agents)
- [Claude Managed Agents documentation](https://platform.claude.com/docs/en/managed-agents/overview)
- [Daytona documentation](https://www.daytona.io/docs/)
