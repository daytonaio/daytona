# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# pylint: disable=no-member

"""FastAPI webhook entrypoint for the shared host orchestrator."""
from __future__ import annotations

import asyncio
import os
import threading

import anthropic
import orchestrator_lib
import uvicorn
from fastapi import FastAPI, HTTPException, Request

PORT = int(os.environ.get("PORT", "5051"))
WEBHOOK_SECRET = os.environ.get("ANTHROPIC_WEBHOOK_SECRET")

app = FastAPI()


async def _handle_webhook(request: Request) -> dict:
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
    loop = asyncio.get_event_loop()
    drained = await loop.run_in_executor(None, orchestrator_lib.drain_work)
    return {"status": "ok", "drained": drained}


@app.post("/")
async def webhook(request: Request):
    return await _handle_webhook(request)


@app.get("/healthz")
def healthz():
    return {"ok": True, "environment_id": orchestrator_lib.ENVIRONMENT_ID}


@app.on_event("startup")
def on_startup() -> None:
    orchestrator_lib.acquire_orchestrator_lock("webhook")
    if WEBHOOK_SECRET is None:
        print(
            "WARNING: ANTHROPIC_WEBHOOK_SECRET is not set; webhook POSTs will 500. "
            "Set it in .env to enable signature verification.",
            flush=True,
        )
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
