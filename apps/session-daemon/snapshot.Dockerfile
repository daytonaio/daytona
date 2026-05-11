# Self-contained session-runtime image, designed to be built by the Daytona
# snapshot API (POST /api/snapshots with buildInfo.dockerfileContent) — not by a
# local `docker build` against the monorepo.
#
# Why this exists alongside ./Dockerfile:
#   - ./Dockerfile builds the session-daemon binary inline from the monorepo
#     (yarn install + nx build), then copies it into a runtime stage. That needs
#     the whole tree as build context, which the snapshot API isn't designed to
#     accept.
#   - This file is single-stage and self-contained: every byte it needs is
#     either in a public registry (Debian, NodeSource, PyPI, npm) or fetched at
#     build time from a URL the caller supplies via SESSION_DAEMON_URL.
#
# How to use this Dockerfile:
#   1. Build apps/session-daemon for linux/amd64 with CGO disabled (the same
#      `nx build session-daemon --configuration=production` that ./Dockerfile
#      runs in stage 1).
#   2. Host the resulting binary somewhere the runner that will execute the
#      build can fetch over HTTPS — e.g. a GitHub Release asset, a presigned
#      object-storage URL, or a tiny static server.
#   3. Render this Dockerfile by substituting the SESSION_DAEMON_URL build arg
#      with that URL, then submit it to the Daytona snapshot API. The
#      `create-snapshot` nx target in apps/session-daemon/project.json does
#      both steps; see apps/session-daemon/README.md for the full flow.
#
# Direct `docker build` for local validation also works:
#   docker build \
#     --build-arg SESSION_DAEMON_URL=https://.../session-daemon-amd64 \
#     -f apps/session-daemon/snapshot.Dockerfile .

FROM debian:bookworm-slim

ARG SESSION_DAEMON_URL
ARG SESSION_DAEMON_SHA256

ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHONUNBUFFERED=1
ENV NODE_ENV=production
ENV SESSION_DAEMON_PORT=2281
ENV SESSION_DAEMON_BIND_ADDR=127.0.0.1
ENV SESSION_DAEMON_USER_NODE_MODULES_ROOT=/workspace
ENV SESSION_DAEMON_NODE_BUNDLE_ROOT=/usr/lib/daytona/repl_host
ENV PATH="/opt/daytona/venv/bin:${PATH}"

# OS deps + Python (3.11, the version debian:bookworm ships) + Node.js 22.
# Mirrors the runtime stage of ./Dockerfile — keep these two in sync; the seed
# migration's expected Python/Node majors live here
# (see apps/api/src/migrations/post-deploy/1778367241001-migration.ts).
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl python3 python3-venv python3-pip build-essential \
      libffi-dev libssl-dev pkg-config \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# Curated Python venv. Pinning is intentional — see plan §1: snapshot is the
# only install path, runtime install is not supported in v1. Bumping these
# requires rebuilding the image (and republishing the python-default snapshot).
# Direct dependency versions are pinned for build stability; note transitive
# deps, OS packages, and the base image tag are not hash-pinned, so builds are
# stable but not byte-for-byte reproducible.
RUN python3 -m venv /opt/daytona/venv && \
    /opt/daytona/venv/bin/pip install --no-cache-dir --upgrade pip && \
    /opt/daytona/venv/bin/pip install --no-cache-dir \
      numpy==2.2.1 pandas==2.2.3 matplotlib==3.10.0 requests==2.32.3 \
      httpx==0.28.1 pydantic==2.10.4 openai==1.59.6 anthropic==0.42.0

# Curated user-side node_modules. Pure-JS only (no native bindings) — V8 session
# constraint, see plan §1. Versions pinned for reproducible builds.
RUN mkdir -p /workspace && cd /workspace \
    && npm init -y >/dev/null \
    && npm install --omit=optional --omit=peer \
        zod@3.24.1 lodash-es@4.17.21 date-fns@4.1.0 papaparse@5.4.1 \
        marked@15.0.6 openai@4.77.3 @anthropic-ai/sdk@0.33.1 \
    && rm -rf /root/.npm

# Host-side bundle dependencies (isolated-vm, esbuild-wasm, just-bash).
# isolated-vm has native bindings compiled against this image's Node version at
# install time; baking it into the image (rather than runtime install) is what
# makes cold-start fast. just-bash is the pure-JS virtual-bash engine backing
# the bash isolate + the in-isolate/Python bash() bridges (no native bindings,
# no real subprocesses). Per plan §1, the daemon's //go:embed only carries the
# entry scripts — heavy deps live in the image layer. Pinned to exact versions.
RUN mkdir -p /usr/lib/daytona/repl_host && cd /usr/lib/daytona/repl_host \
    && npm init -y >/dev/null \
    && npm install --omit=optional --omit=peer isolated-vm@5.0.3 esbuild-wasm@0.24.2 just-bash@3.0.2 \
    && rm -rf /root/.npm

# Fetch the daemon binary the caller pre-built. The URL scheme is enforced as
# https:// (CDN/object-storage URLs — presigned S3, GitHub Releases — work without
# extra auth wiring); plain http:// is rejected to prevent insecure transport, with
# an exception for loopback hosts so local builds can serve over http://localhost.
#
# Two delivery paths use the same Dockerfile:
#   - Daytona snapshot API path: the nx `create-snapshot` target runs envsubst
#     against the SESSION_DAEMON_URL and SESSION_DAEMON_SHA256 placeholders
#     before submitting, so both become literal strings in the rendered
#     Dockerfile. The snapshot API doesn't support docker --build-arg, so the
#     ARG lines above are unused on this path (Docker just ignores them).
#   - Direct `docker build --build-arg SESSION_DAEMON_URL=... -f snapshot.Dockerfile .`
#     path: the ARGs above are resolved by Docker at build time and inlined
#     here by Docker's own variable expansion.
#
# SESSION_DAEMON_SHA256 is optional but strongly recommended — the snapshot
# image is what backs every sandbox spawned from this template, so a silent
# binary swap upstream would be deeply unpleasant. When set, the build fails
# loudly if the digest doesn't match. When empty, the check is skipped.
RUN set -e; \
    case "${SESSION_DAEMON_URL}" in \
      https://*) : ;; \
      'http://localhost'|'http://localhost:'*|'http://localhost/'*|'http://127.0.0.1'|'http://127.0.0.1:'*|'http://127.0.0.1/'*|'http://[::1]'|'http://[::1]:'*|'http://[::1]/'*) : ;; \
      *) echo "SESSION_DAEMON_URL must use https:// (or http://localhost for local builds); refusing insecure transport: ${SESSION_DAEMON_URL}" >&2; exit 1 ;; \
    esac; \
    mkdir -p /opt/daytona; \
    curl -fsSL --retry 5 --retry-delay 2 \
      -o /opt/daytona/session-daemon \
      "${SESSION_DAEMON_URL}"; \
    if [ -n "${SESSION_DAEMON_SHA256}" ]; then \
      echo "${SESSION_DAEMON_SHA256}  /opt/daytona/session-daemon" | sha256sum -c -; \
    fi; \
    chmod +x /opt/daytona/session-daemon; \
    test -x /opt/daytona/session-daemon

# Run as an unprivileged user. The daemon binds loopback-only and needs no
# privileged ports; /workspace must stay writable for user code and node_modules.
RUN useradd --system --create-home --home-dir /home/daytona --shell /usr/sbin/nologin daytona \
    && chown -R daytona:daytona /workspace
USER daytona
# NOTE: do not invoke the daemon here as a smoke test. It takes no --version/
# --help flag (see apps/session-daemon/cmd/main.go); any invocation starts the
# blocking server and would hang the build forever. `test -x` is the only safe
# build-time check.

# Healthcheck targets the loopback port the daemon binds to. Because the daemon
# is loopback-only, only an in-container healthcheck makes sense.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -fsS http://127.0.0.1:${SESSION_DAEMON_PORT}/healthz || exit 1

# The runner spawns the snapshot's entrypoint as the sandbox's PID 1. The
# snapshot API extracts ENTRYPOINT from this Dockerfile (see
# snapshot.service.ts:getEntrypointFromDockerfile) and stores it on the
# snapshot row, so we don't need to pass --entrypoint on `daytona snapshot
# create`. The seed migration sets the same value as a fallback.
ENTRYPOINT ["/opt/daytona/session-daemon"]
