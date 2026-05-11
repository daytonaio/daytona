# Daytona Sessions: zero-setup, sub-second code execution
# =======================================================
# Unlike `sandbox.process.code_run`, sessions do NOT require you to provision
# a sandbox first. The Daytona API auto-allocates an ephemeral execution
# context per call (or a persistent one if you ask for it). Each function
# below highlights one product capability.

from daytona import Daytona, SessionExpiredError, SessionInvalidatedError, SessionRunOptions


def one_shot_python(daytona: Daytona) -> None:
    # The simplest possible call: no sandbox, no context, no setup.
    print("=== one-shot Python ===")
    result = daytona.session.run("print(2 ** 10)", SessionRunOptions(language="python"))
    print("stdout:", result.stdout.rstrip())
    print("took:", result.duration_ms, "ms")


def one_shot_typescript_with_imports_and_tla(daytona: Daytona) -> None:
    # Showcases the headline TypeScript capabilities of the V8-session runtime:
    #   - real ESM `import` from a curated, pre-installed registry (zod here)
    #   - top-level await, including over network calls
    #   - the trailing expression's value comes back as a `display` payload
    print("\n=== one-shot TypeScript with ESM imports + top-level await ===")
    code = """
import { z } from "zod"
const schema = z.object({ name: z.string(), age: z.number() })
const parsed = schema.safeParse({ name: "Ada", age: 36 })
console.log("parsed.success:", parsed.success)
console.log("data:", JSON.stringify(parsed.data))

const r = await fetch("https://httpbin.org/get?from=session")
const j = await r.json()
console.log("fetched OK, status:", r.status, "from:", j.args.from);

// Trailing expression — surfaced as a 'display' frame on the result.
({ ok: r.ok, status: r.status })
""".strip()
    result = daytona.session.run(code, SessionRunOptions(language="typescript"))
    print(result.stdout, end="")
    if result.error:
        print(f"error: {result.error.name} {result.error.value or ''}")
        return
    for display in result.displays:
        print(f"display formats: {', '.join(display.formats)}")
        if "text/plain" in display.data:
            print(f"display value: {display.data['text/plain']}")


def stateful_context(daytona: Daytona) -> None:
    # Persistent contexts keep variables alive across runs — perfect for
    # notebook-style workflows or interactive REPL frontends.
    print("\n=== stateful Python context (state survives across runs) ===")
    ctx = daytona.session.create_session(language="python")
    try:
        _ = daytona.session.run("counter = 0", SessionRunOptions(context=ctx))
        for _ in range(3):
            r = daytona.session.run(
                'counter += 1\nprint(f"counter = {counter}")',
                SessionRunOptions(context=ctx),
            )
            print(r.stdout, end="")
    finally:
        daytona.session.delete_session(ctx.id)


def multi_context(daytona: Daytona) -> None:
    # Multiple persistent contexts can coexist on the same session instance.
    # Each one keeps its own globals — variables defined in `ctx_a` are
    # invisible to `ctx_b`, even within the same Python language. You can also
    # mix languages; here we run a Python pair alongside a TypeScript context
    # and list them all via list_sessions() to confirm they're live.
    print("\n=== multiple coexisting contexts (independent state per context) ===")
    ctx_a = daytona.session.create_session(language="python")
    ctx_b = daytona.session.create_session(language="python")
    ctx_ts = daytona.session.create_session(language="typescript")
    try:
        _ = daytona.session.run('identity = "alpha"', SessionRunOptions(context=ctx_a))
        _ = daytona.session.run('identity = "beta"', SessionRunOptions(context=ctx_b))
        # TypeScript: each `run()` compiles a fresh ES module, so `const`/`let`
        # declarations don't leak between runs. Use `globalThis` (or `var`) to
        # persist values across runs in the same context.
        _ = daytona.session.run('globalThis.identity = "gamma"', SessionRunOptions(context=ctx_ts))

        a = daytona.session.run('print(f"A says: {identity}")', SessionRunOptions(context=ctx_a))
        b = daytona.session.run('print(f"B says: {identity}")', SessionRunOptions(context=ctx_b))
        ts = daytona.session.run("console.log(`TS says: ${globalThis.identity}`)", SessionRunOptions(context=ctx_ts))
        print(a.stdout, end="")
        print(b.stdout, end="")
        print(ts.stdout, end="")

        # list_sessions() returns every persistent context in the org. Filter
        # to the three we just created so the example is deterministic when
        # other contexts happen to be live.
        ours = {ctx_a.id, ctx_b.id, ctx_ts.id}
        live = [c for c in daytona.session.list_sessions() if c.id in ours]
        print(f"live contexts: {len(live)} ({', '.join(sorted(c.language for c in live))})")
    finally:
        for ctx in (ctx_a, ctx_b, ctx_ts):
            daytona.session.delete_session(ctx.id)


def streaming_with_handlers(daytona: Daytona) -> None:
    # run_stream() opens a WebSocket and invokes the per-frame handlers as
    # each chunk arrives. The returned aggregated result is the same shape
    # as run(), so callers can use one or the other interchangeably.
    print("\n=== streaming Python with per-frame handlers ===")
    code = """
for i in range(5):
    print(f"line {i}", flush=True)
import pandas as pd
pd.DataFrame({"x": [1, 2, 3], "y": [10, 20, 30]})
""".strip()
    result = daytona.session.run_stream(
        code,
        SessionRunOptions(language="python"),
        on_stdout=lambda chunk: print(f"[stream] {chunk}", end=""),
        on_error=lambda err: print(f"[stream-error] {err.name}: {err.value or ''}"),
        on_display=lambda d: print(f"[display] formats={','.join(d.formats)}"),
    )
    print(f"({len(result.stdout)} stdout chars, " + f"{len(result.displays)} display frames in {result.duration_ms}ms)")


def error_surfacing(daytona: Daytona) -> None:
    # User errors — including TypeScript imports of unsupported Node builtins
    # — come back as a structured `error` frame, never as a thrown exception
    # on the SDK side.
    print("\n=== structured errors from user code ===")
    py = daytona.session.run("1 / 0", SessionRunOptions(language="python"))
    print(f"Python: {py.error.name if py.error else None}: {py.error.value if py.error else ''}")

    ts = daytona.session.run("import 'fs'", SessionRunOptions(language="typescript"))
    print(f"TypeScript: {ts.error.name if ts.error else None}: {ts.error.value if ts.error else ''}")


def main() -> None:
    # The Daytona client reads DAYTONA_API_URL / DAYTONA_API_KEY from env
    # (or pass them via the `Daytona(DaytonaConfig(...))` config object).
    # No sandbox to create, no image to build — the API server handles
    # session provisioning internally.
    daytona = Daytona()

    try:
        one_shot_python(daytona)
        one_shot_typescript_with_imports_and_tla(daytona)
        stateful_context(daytona)
        multi_context(daytona)
        streaming_with_handlers(daytona)
        error_surfacing(daytona)
    except SessionInvalidatedError as exc:
        # The underlying sandbox was rolled (death / autostop). Drop the
        # context ref and re-create on next call.
        print(f"context was invalidated: {exc}")
    except SessionExpiredError as exc:
        # Hit the idle / absolute TTL. `exc.reason` is 'idle' | 'absolute'.
        print(f"context expired ({exc.reason}): {exc}")


if __name__ == "__main__":
    main()
