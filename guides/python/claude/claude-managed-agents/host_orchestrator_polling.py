# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Long-polling entrypoint for the shared host orchestrator."""
from __future__ import annotations

import os
import signal
import threading

import orchestrator_lib

POLL_RECLAIM_OLDER_THAN_MS = int(os.environ.get("POLL_RECLAIM_OLDER_THAN_MS", "2000"))


def poll_block_ms_from_env() -> int:
    raw = os.environ.get("POLL_BLOCK_MS", "999")
    try:
        value = int(raw)
    except ValueError as e:
        raise ValueError(f"POLL_BLOCK_MS must be an integer in 1..999, got {raw!r}") from e
    if not 1 <= value <= 999:
        raise ValueError(f"POLL_BLOCK_MS must be in 1..999, got {value}")
    return value


POLL_BLOCK_MS = poll_block_ms_from_env()


def install_signal_handlers() -> None:
    def request_shutdown(signum, _frame) -> None:
        print(f"[poll] shutdown requested by signal {signum}", flush=True)
        orchestrator_lib.shutdown.set()

    signal.signal(signal.SIGTERM, request_shutdown)
    signal.signal(signal.SIGINT, request_shutdown)


def poll_loop() -> None:
    transient_attempts = 0
    while not orchestrator_lib.shutdown.is_set():
        try:
            orchestrator_lib.drain_work(
                block_ms=POLL_BLOCK_MS,
                reclaim_older_than_ms=POLL_RECLAIM_OLDER_THAN_MS,
                raise_poll_errors=True,
            )
            transient_attempts = 0
        except Exception as e:
            if orchestrator_lib.is_permanent_poll_error(e):
                raise
            transient_attempts += 1
            wait = min(60.0, 2.0 ** min(transient_attempts, 6))
            print(
                f"[poll] transient failure; retrying in {wait:.1f}s " f"({type(e).__name__}: {e})",
                flush=True,
            )
            orchestrator_lib.shutdown.wait(wait)


def main() -> None:
    orchestrator_lib.acquire_orchestrator_lock("polling")
    install_signal_handlers()
    threading.Thread(
        target=orchestrator_lib.janitor_loop,
        kwargs={"recover_crashed_runners": False},
        daemon=True,
    ).start()
    print(
        "host polling orchestrator running "
        f"env={orchestrator_lib.ENVIRONMENT_ID} "
        f"block_ms={POLL_BLOCK_MS} "
        f"reclaim_older_than_ms={POLL_RECLAIM_OLDER_THAN_MS}",
        flush=True,
    )
    poll_loop()


if __name__ == "__main__":
    main()
