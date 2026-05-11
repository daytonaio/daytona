# Sessions vs. classic sandboxes: a head-to-head latency comparison
# ==================================================================
# Same Python snippet, three execution surfaces:
#   - sandbox.process.exec()    — shell-command exec inside a full sandbox
#   - sandbox.process.code_run() — python-code exec inside a full sandbox
#   - daytona.session.run()      — V8/CPython session auto-allocated by the API
#
# Each section reports:
#   - time to be "ready to execute" (for sandboxes: create(); for sessions:
#     implicit auto-provisioning on the first call)
#   - per-call latency for N subsequent execs

import time
from typing import Any, Callable

from daytona import Daytona, SessionRunOptions

CODE = "print(2 ** 10)"
SHELL_CMD = "python3 -c 'print(2 ** 10)'"
N = 5


def fmt_ms(seconds: float) -> str:
    return f"{seconds * 1000:8.1f} ms"


def measure(fn: Callable[[], Any], n: int = N) -> list[float]:
    times: list[float] = []
    for _ in range(n):
        start = time.perf_counter()
        _ = fn()
        times.append(time.perf_counter() - start)
    return times


def report(label: str, times: list[float]) -> None:
    avg = sum(times) / len(times)
    print(f"{label:20s} {fmt_ms(avg)}  " + f"(n={len(times)}, min={fmt_ms(min(times))}, max={fmt_ms(max(times))})")


def classic_sandbox(daytona: Daytona) -> float:
    # Default snapshot, default language (python). No params = the same path
    # any first-time user would take.
    print("=== Classic sandbox: daytona.create() + process.{exec,code_run} ===")

    t0 = time.perf_counter()
    sandbox = daytona.create(timeout=180)
    create_secs = time.perf_counter() - t0
    print(f"{'create():':20s} {fmt_ms(create_secs)}  (one-time, blocks until the container is up)")

    try:
        # Shell-command exec: `python3 -c 'print(2 ** 10)'`. Each call spawns
        # a fresh OS process inside the container.
        t1 = time.perf_counter()
        _ = sandbox.process.exec(SHELL_CMD)
        first_exec = time.perf_counter() - t1
        print(f"{'first exec:':20s} {fmt_ms(first_exec)}")
        report("avg exec:", measure(lambda: sandbox.process.exec(SHELL_CMD)))

        # Code-run exec: same snippet but routed to the in-sandbox code
        # interpreter, which keeps a warm CPython around.
        t2 = time.perf_counter()
        _ = sandbox.process.code_run(CODE)
        first_code_run = time.perf_counter() - t2
        print(f"{'first code_run:':20s} {fmt_ms(first_code_run)}")
        report("avg code_run:", measure(lambda: sandbox.process.code_run(CODE)))

        # "time-to-first-exec" is the cheaper of the two — code_run is what a
        # user shipping Python code would actually use.
        time_to_first = create_secs + first_code_run
        print(f"\n{'time-to-first-exec:':20s} {fmt_ms(time_to_first)}  (create + first code_run)")
        return time_to_first
    finally:
        daytona.delete(sandbox)


def session(daytona: Daytona) -> float:
    # No create(), no handle. The first call lazily acquires a warm session
    # instance from the API's pool — that cold cost is folded into "first
    # session.run" below for an apples-to-apples comparison.
    print("\n=== Session: daytona.session.run() (no sandbox handle needed) ===")

    t0 = time.perf_counter()
    _ = daytona.session.run(CODE, SessionRunOptions(language="python"))
    first_secs = time.perf_counter() - t0
    print(f"{'first session.run:':20s} {fmt_ms(first_secs)}  (includes any cold-start provisioning)")
    report("avg session.run:", measure(lambda: daytona.session.run(CODE, SessionRunOptions(language="python"))))

    print(f"\n{'time-to-first-exec:':20s} {fmt_ms(first_secs)}")
    return first_secs


def main() -> None:
    daytona = Daytona()
    sandbox_ttfe = classic_sandbox(daytona)
    session_ttfe = session(daytona)

    # Summary: the headline number is "time-to-first-exec" — how long until
    # the user can run code. The steady-state numbers are reported per section.
    print("\n=== Summary ===")
    if session_ttfe > 0:
        ratio = sandbox_ttfe / session_ttfe
        print(f"session is {ratio:.1f}x faster to first exec ({fmt_ms(sandbox_ttfe)} vs {fmt_ms(session_ttfe)})")


if __name__ == "__main__":
    main()
