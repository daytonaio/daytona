# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import os
import sys
import time

import requests
from dotenv import load_dotenv

from daytona import (
    CreateSandboxFromImageParams,
    Daytona,
    DaytonaConfig,
    GpuType,
    Image,
    Resources,
    SessionExecuteRequest,
)

load_dotenv()

MODEL = "openai/gpt-oss-20b"
SERVED_AS = "gpt-oss-20b"
SGLANG_IMAGE = "lmsysorg/sglang:v0.5.12.post1-cu130"
PORT = 8000
TARGET = "us-east-1"  # current region for GPU sandboxes
SESSION = "sglang"  # name of the background session the server runs in
BOOT_TIMEOUT = 900  # max seconds to wait for the server to come up

daytona = Daytona(DaytonaConfig(target=TARGET))
env_vars = {"HF_TOKEN": os.environ["HF_TOKEN"]} if os.environ.get("HF_TOKEN") else {}
print(f"creating GPU sandbox from {SGLANG_IMAGE} ...", flush=True)
sb = daytona.create(
    CreateSandboxFromImageParams(
        image=Image.base(SGLANG_IMAGE),
        resources=Resources(
            gpu=1,
            gpu_type=[GpuType.H100, GpuType.RTX_PRO_6000],  # preference order
        ),
        auto_stop_interval=0,
        ephemeral=True,
        env_vars=env_vars,
    ),
    timeout=600,
)
print(f"sandbox {sb.id} up", flush=True)


def dump_log(command_id):
    log_path = os.path.abspath(f"sglang-{sb.id}.log")
    with open(log_path, "w") as f:
        f.write(sb.process.get_session_command_logs(SESSION, command_id).output or "")
    return log_path


try:
    sb.process.create_session(SESSION)
    cmd = sb.process.execute_session_command(
        SESSION,
        SessionExecuteRequest(
            command=(
                f"python3 -m sglang.launch_server --model-path {MODEL} "
                f"--served-model-name {SERVED_AS} "
                f"--port {PORT} "
                # Parser names must match the model family and your SGLang version,
                # or tool calls and reasoning come back unparsed in `content`.
                "--tool-call-parser gpt-oss --reasoning-parser gpt-oss "
                # report RadixAttention cache hits in usage.prompt_tokens_details
                "--enable-cache-report"
            ),
            run_async=True,
        ),
    )
    cmd_id = cmd.cmd_id

    pv = sb.get_preview_link(PORT)
    hdr = {"x-daytona-preview-token": pv.token}
    print(f"preview: {pv.url}  (waiting for /health_generate, up to {BOOT_TIMEOUT}s)", flush=True)

    deadline = time.time() + BOOT_TIMEOUT
    ready = False
    printed = 0
    while time.time() < deadline:
        # logs are a cumulative snapshot; print only the new tail
        out = sb.process.get_session_command_logs(SESSION, cmd_id).output or ""
        if len(out) > printed:
            sys.stdout.write(out[printed:])
            sys.stdout.flush()
            printed = len(out)
        # the server runs until killed; an exit code means it died
        exit_code = sb.process.get_session_command(SESSION, cmd_id).exit_code
        if exit_code is not None:
            print(f"!! sglang exited with code {exit_code}. Full log saved to {dump_log(cmd_id)}", flush=True)
            sys.exit(1)
        try:
            # /health_generate runs a real forward pass, not just a liveness check
            if requests.get(f"{pv.url}/health_generate", headers=hdr, timeout=10).status_code == 200:
                ready = True
                break
        except requests.RequestException:
            pass
        time.sleep(10)

    if not ready:
        print(f"!! server never became healthy. Full log saved to {dump_log(cmd_id)}", flush=True)
        sys.exit(1)

    print("\nready - paste into your shell:", flush=True)
    print(f"export ENDPOINT={pv.url}", flush=True)
    print(f"export TOKEN={pv.token}", flush=True)

finally:
    # auto_stop_interval=0 keeps it from idle-stopping; on failure this also
    # preserves the downloaded weights. Reconnect to reuse, delete when done.
    print(f"\nsandbox left UP: {sb.id}", flush=True)
    print(f"  reconnect:  daytona.get('{sb.id}')", flush=True)
    print(f"  delete:     daytona.get('{sb.id}').delete()", flush=True)
