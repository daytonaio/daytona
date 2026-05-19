# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# pylint: disable=no-member

"""FastAPI webhook entrypoint for the shared host orchestrator."""
from __future__ import annotations

import math
import os
import threading

import anthropic
import orchestrator_lib
import uvicorn
from fastapi import BackgroundTasks, FastAPI, HTTPException, Request


def _positive_float_from_env(name: str, default: str) -> float:
    raw = os.environ.get(name, default)
    try:
        value = float(raw)
    except ValueError as e:
        raise ValueError(f"{name} must be a positive number, got {raw!r}") from e
    if not math.isfinite(value) or value <= 0:
        raise ValueError(f"{name} must be a finite positive number, got {value}")
    return value


def _nonnegative_int_from_env(name: str, default: str) -> int:
    raw = os.environ.get(name, default)
    try:
        value = int(raw)
    except ValueError as e:
        raise ValueError(f"{name} must be a non-negative integer, got {raw!r}") from e
    if value < 0:
        raise ValueError(f"{name} must be a non-negative integer, got {value}")
    return value


PORT = int(os.environ.get("PORT", "5051"))
WEBHOOK_SECRET = os.environ.get("ANTHROPIC_WEBHOOK_SECRET")
WEBHOOK_DRAIN_SECONDS = _positive_float_from_env("WEBHOOK_DRAIN_SECONDS", "30")
WEBHOOK_RECLAIM_OLDER_THAN_MS = _nonnegative_int_from_env("WEBHOOK_RECLAIM_OLDER_THAN_MS", "2000")

app = FastAPI()


def webhook_fallback_drain_loop() -> None:
    """Poll occasionally so a delivered webhook is not the only drain trigger."""
    transient_attempts = 0
    while not orchestrator_lib.shutdown.is_set():
        try:
            processed = orchestrator_lib.drain_work(
                reclaim_older_than_ms=WEBHOOK_RECLAIM_OLDER_THAN_MS,
                raise_poll_errors=True,
            )
            transient_attempts = 0
            if processed:
                print(f"[webhook] fallback drain processed={len(processed)}", flush=True)
            wait = WEBHOOK_DRAIN_SECONDS
        except Exception as e:
            transient_attempts += 1
            wait = min(60.0, 2.0 ** min(transient_attempts, 6))
            print(
                f"[webhook] fallback drain failed; retrying in {wait:.1f}s " f"({type(e).__name__}: {e})",
                flush=True,
            )
        orchestrator_lib.shutdown.wait(wait)


async def _handle_webhook(request: Request, background_tasks: BackgroundTasks) -> dict:
    if WEBHOOK_SECRET is None:
        raise HTTPException(status_code=500, detail="ANTHROPIC_WEBHOOK_SECRET not set")
    raw = await request.body()
    try:
        event = orchestrator_lib.CLIENT.beta.webhooks.unwrap(
            raw.decode(),
            headers=dict(request.headers),
            key=WEBHOOK_SECRET,
        )
    except anthropic.APIWebhookValidationError as e:
        raise HTTPException(status_code=401, detail=str(e)) from e
    ev_type = event.data.type
    session_id = event.data.id
    print(f"[webhook] event={ev_type} session={session_id}", flush=True)
    if ev_type != "session.status_run_started":
        return {"status": "ignored", "type": ev_type}
    # Ack immediately and drain after the response; starting sandboxes can take
    # tens of seconds and would otherwise hold the webhook POST open long
    # enough for Anthropic to time out and retry. DRAIN_LOCK serializes
    # concurrent drains inside orchestrator_lib.
    background_tasks.add_task(
        orchestrator_lib.drain_work,
        reclaim_older_than_ms=WEBHOOK_RECLAIM_OLDER_THAN_MS,
    )
    return {"status": "queued"}


@app.post("/")
async def webhook(request: Request, background_tasks: BackgroundTasks):
    return await _handle_webhook(request, background_tasks)


@app.get("/healthz")
def healthz():
    return {"ok": True, "environment_id": orchestrator_lib.ENVIRONMENT_ID}


@app.on_event("startup")
def on_startup() -> None:
    if WEBHOOK_SECRET is None:
        raise RuntimeError(
            "ANTHROPIC_WEBHOOK_SECRET is not set; webhook mode cannot verify signatures. "
            "Set it in .env, or run host_orchestrator_polling.py if you don't want a webhook."
        )
    orchestrator_lib.acquire_orchestrator_lock("webhook")
    threading.Thread(
        target=webhook_fallback_drain_loop,
        daemon=True,
    ).start()
    threading.Thread(
        target=orchestrator_lib.janitor_loop,
        kwargs={"recover_crashed_runners": True},
        daemon=True,
    ).start()
    print(
        f"host webhook orchestrator listening on :{PORT} " f"env={orchestrator_lib.ENVIRONMENT_ID}",
        flush=True,
    )


@app.on_event("shutdown")
def on_shutdown() -> None:
    orchestrator_lib.shutdown.set()


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=PORT, log_level="info")
