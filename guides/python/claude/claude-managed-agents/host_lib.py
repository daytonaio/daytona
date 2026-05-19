# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""
Shared host-side helpers for ensuring a Daytona sandbox is running and the
in-sandbox EnvironmentWorker runner (sandbox_runner.py) is alive inside it.

Used by the host orchestrator entrypoints and manual diagnostic drivers.
Keeping these in one place avoids drift between the paths.
"""
from __future__ import annotations

import contextlib
import hashlib
import json
import os
import pathlib
import shlex
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any, Callable

from daytona import CreateSandboxFromSnapshotParams, Daytona, DaytonaConflictError, DaytonaNotFoundError

REPO = pathlib.Path(__file__).resolve().parent
RUNNER_LOG = "/tmp/sandbox_runner.log"
RUNNER_PIDFILE = "/home/daytona/sandbox_runner.pid"
RUNNER_EXITCODE_FILE = "/home/daytona/sandbox_runner.exit"
STOPPED_AT_LABEL = "byoc.stopped_at"
WORK_ID_LABEL = "byoc.work_id"
DEFAULT_DOCKERFILE = REPO / "Dockerfile.default"
SNAPSHOT_NAME_METADATA_KEY = "daytona.snapshot_name"
SANDBOX_ID_METADATA_KEY = "daytona.sandbox_id"
SESSION_ID_LABEL = "byoc.session_id"
ENVIRONMENT_ID_LABEL = "byoc.environment_id"
MODE_LABEL = "byoc.mode"
MODE_IN_SANDBOX = "in-sandbox"
MODE_PREPARED = "prepared"
MODE_TERMINAL = "terminal"
AUTO_ARCHIVE_INTERVAL_MINUTES = 1440


class SandboxSelectionError(ValueError):
    """Raised when Daytona sandbox-selection metadata is invalid."""


class SandboxResurrectionError(SandboxSelectionError):
    """Raised when a session-named sandbox is not orchestrator-managed."""


class RunnerAlreadyAlive(RuntimeError):
    """Raised when a live runner is already handling this session."""


class RunnerLaunchDeferred(RuntimeError):
    """Raised when an ambiguous live runner should be left untouched."""


@dataclass(frozen=True)
class SandboxSelection:
    snapshot_name: str | None = None
    sandbox_id: str | None = None


@dataclass(frozen=True)
class RunnerProcessState:
    state: str
    pid: int | None = None
    exit_code: int | None = None
    detail: str | None = None


def default_snapshot_name() -> str:
    explicit = os.environ.get("DEFAULT_SNAPSHOT_NAME")
    if explicit:
        return explicit
    sha = hashlib.sha256(DEFAULT_DOCKERFILE.read_bytes()).hexdigest()[:8]
    return f"byoc-env-default-{sha}"


def sandbox_name(session_id: str) -> str:
    return f"byoc-{session_id}"


def sandbox_selection_from_metadata(metadata: dict[str, Any] | None) -> SandboxSelection:
    metadata = metadata or {}
    snapshot_name = _metadata_string(metadata, SNAPSHOT_NAME_METADATA_KEY)
    sandbox_id = _metadata_string(metadata, SANDBOX_ID_METADATA_KEY)

    if snapshot_name and sandbox_id:
        raise SandboxSelectionError(
            f"{SNAPSHOT_NAME_METADATA_KEY} and {SANDBOX_ID_METADATA_KEY} are mutually exclusive"
        )
    return SandboxSelection(
        snapshot_name=snapshot_name,
        sandbox_id=sandbox_id,
    )


def get_or_create_sandbox(
    daytona: Daytona,
    session_id: str,
    environment_id: str,
    *,
    metadata: dict[str, Any] | None = None,
):
    selection = sandbox_selection_from_metadata(metadata)
    if selection.sandbox_id:
        normal_name = sandbox_name(session_id)
        with contextlib.suppress(DaytonaNotFoundError):
            normal_sandbox = daytona.get(normal_name)
            if normal_sandbox.id != selection.sandbox_id:
                raise SandboxSelectionError(
                    f"session sandbox {normal_name!r} already exists as {normal_sandbox.id!r}; "
                    f"refusing to attach {selection.sandbox_id!r}"
                )
        return attach_prepared_sandbox(
            daytona,
            session_id,
            environment_id,
            selection.sandbox_id,
        )

    name = sandbox_name(session_id)
    with contextlib.suppress(DaytonaNotFoundError):
        return _managed_existing_sandbox(daytona.get(name), name)

    params = CreateSandboxFromSnapshotParams(
        name=name,
        snapshot=selection.snapshot_name or default_snapshot_name(),
        labels={
            SESSION_ID_LABEL: session_id,
            ENVIRONMENT_ID_LABEL: environment_id,
            MODE_LABEL: MODE_IN_SANDBOX,
        },
    )
    try:
        sandbox = daytona.create(params, timeout=180)
    except DaytonaConflictError:
        return _managed_existing_sandbox(daytona.get(name), name)
    try:
        sandbox.set_auto_archive_interval(AUTO_ARCHIVE_INTERVAL_MINUTES)
    except Exception as e:
        print(
            f"[host] set_auto_archive_interval failed for {sandbox.id}: " f"{type(e).__name__}: {e}",
            flush=True,
        )
    return sandbox


def _managed_existing_sandbox(sandbox, name: str):
    sandbox.refresh_data()
    existing_mode = (getattr(sandbox, "labels", None) or {}).get(MODE_LABEL)
    if existing_mode != MODE_IN_SANDBOX:
        raise SandboxResurrectionError(
            f"sandbox {name!r} exists with {MODE_LABEL}={existing_mode!r}; "
            f"refusing to resurrect (only {MODE_IN_SANDBOX!r} sandboxes are managed)"
        )
    return sandbox


def attach_prepared_sandbox(
    daytona: Daytona,
    session_id: str,
    environment_id: str,
    sandbox_id: str,
):
    sandbox = daytona.get(sandbox_id)
    sandbox.refresh_data()
    labels = dict(getattr(sandbox, "labels", None) or {})

    existing_session_id = labels.get(SESSION_ID_LABEL)
    if existing_session_id == session_id:
        if labels.get(ENVIRONMENT_ID_LABEL) != environment_id:
            raise SandboxSelectionError(
                f"sandbox {sandbox_id!r} is attached to environment "
                f"{labels.get(ENVIRONMENT_ID_LABEL)!r}, expected {environment_id!r}"
            )
        if labels.get(MODE_LABEL) != MODE_IN_SANDBOX:
            raise SandboxSelectionError(
                f"sandbox {sandbox_id!r} is attached to this session but mode is "
                f"{labels.get(MODE_LABEL)!r}, expected {MODE_IN_SANDBOX!r}"
            )
        return sandbox

    if existing_session_id:
        raise SandboxSelectionError(f"sandbox {sandbox_id!r} is already attached to session {existing_session_id!r}")

    if labels.get(ENVIRONMENT_ID_LABEL) != environment_id:
        raise SandboxSelectionError(
            f"sandbox {sandbox_id!r} has {ENVIRONMENT_ID_LABEL}={labels.get(ENVIRONMENT_ID_LABEL)!r}, "
            f"expected {environment_id!r}"
        )
    if labels.get(MODE_LABEL) != MODE_PREPARED:
        raise SandboxSelectionError(f"sandbox {sandbox_id!r} must have {MODE_LABEL}={MODE_PREPARED!r} before attach")
    merge_labels(
        sandbox,
        **{
            SESSION_ID_LABEL: session_id,
            MODE_LABEL: MODE_IN_SANDBOX,
            STOPPED_AT_LABEL: None,
        },
    )
    sandbox.refresh_data()
    refreshed_labels = dict(getattr(sandbox, "labels", None) or {})
    if (
        refreshed_labels.get(SESSION_ID_LABEL) != session_id
        or refreshed_labels.get(ENVIRONMENT_ID_LABEL) != environment_id
        or refreshed_labels.get(MODE_LABEL) != MODE_IN_SANDBOX
    ):
        raise SandboxSelectionError(f"sandbox {sandbox_id!r} attach label verification failed")
    return sandbox


def _metadata_string(metadata: dict[str, Any], key: str) -> str | None:
    value = metadata.get(key)
    if value is None:
        return None
    if not isinstance(value, str):
        raise SandboxSelectionError(f"{key} must be a string")
    value = value.strip()
    if not value:
        raise SandboxSelectionError(f"{key} must be non-empty")
    return value


def ensure_started(sandbox) -> None:
    sandbox.refresh_data()
    if sandbox.state in ("stopped", "archived"):
        sandbox.start(timeout=300)
    elif sandbox.state != "started":
        sandbox.wait_for_sandbox_start(timeout=300)


def iso8601_utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def merge_labels(sandbox, **changes) -> None:
    labels = dict(getattr(sandbox, "labels", None) or {})
    for key, value in changes.items():
        if value is None:
            labels.pop(key, None)
        else:
            labels[key] = str(value)
    sandbox.set_labels(labels)


def stop_sandbox(sandbox) -> None:
    sandbox.stop(timeout=300)
    # Clear WORK_ID_LABEL too: any work item the runner was holding is no
    # longer in-flight from this sandbox after a stop. The orchestrator
    # force-stops the queued lease separately before calling here.
    merge_labels(
        sandbox,
        **{
            STOPPED_AT_LABEL: iso8601_utc_now(),
            WORK_ID_LABEL: None,
        },
    )


def archive_sandbox(sandbox) -> None:
    sandbox.refresh_data()
    if sandbox.state == "started":
        stop_sandbox(sandbox)

    merge_labels(sandbox, **{MODE_LABEL: MODE_TERMINAL})

    sandbox.refresh_data()
    if sandbox.state not in ("archived", "archiving"):
        sandbox.archive()


def clear_stop_state(sandbox) -> None:
    merge_labels(
        sandbox,
        **{
            STOPPED_AT_LABEL: None,
        },
    )


def preflight_runner_environment(sandbox) -> None:
    script = """
    set -eu
    mkdir -p /home/daytona /mnt/session
    test -w /home/daytona
    test -w /mnt/session
    command -v bash >/dev/null
    command -v python3.12 >/dev/null
    python3.12 -m pip --version >/dev/null
    command -v setsid >/dev/null
    command -v ps >/dev/null
    command -v awk >/dev/null
    command -v tr >/dev/null
    """
    r = sandbox.process.exec("bash -lc " + shlex.quote(script), timeout=30)
    if r.exit_code not in (0, None):
        raise RuntimeError(
            "sandbox is missing runner prerequisites " f"(exit {r.exit_code}); output:\n{r.result or '<empty>'}"
        )


def install_runner(sandbox) -> None:
    preflight_runner_environment(sandbox)
    r = sandbox.process.exec("mkdir -p /mnt/session && chmod 777 /mnt/session", timeout=10)
    if r.exit_code not in (0, None):
        raise RuntimeError(f"failed to prepare /mnt/session (exit {r.exit_code}); " f"output:\n{r.result or '<empty>'}")
    sandbox.fs.upload_file(
        (REPO / "sandbox_runner.py").read_bytes(),
        "/home/daytona/sandbox_runner.py",
    )
    r = sandbox.process.exec(
        "python3.12 -m pip install --quiet --user 'anthropic>=0.103'",
        timeout=180,
    )
    if r.exit_code not in (0, None):
        raise RuntimeError(f"sandbox pip install failed (exit {r.exit_code}); " f"output:\n{r.result or '<empty>'}")


def runner_process_state(sandbox) -> RunnerProcessState:
    script = f"""
    set +e
    pidfile={shlex.quote(RUNNER_PIDFILE)}
    exitfile={shlex.quote(RUNNER_EXITCODE_FILE)}

    emit() {{
      printf '{{"state":"%s","pid":%s,"exit_code":%s,"detail":"%s"}}\\n' "$1" "$2" "$3" "$4"
    }}

    if test -f "$exitfile"; then
      code=$(head -n 1 "$exitfile" 2>/dev/null | tr -d '[:space:]')
      case "$code" in
        0)
          emit exited_zero null 0 ""
          exit 0
          ;;
        ''|*[!0-9]*)
          emit unknown null null bad_exit_code
          exit 0
          ;;
        *)
          emit exited_nonzero null "$code" ""
          exit 0
          ;;
      esac
    fi

    find_stale_runner() {{
      ps -eo pid=,comm=,args= |
        awk '$2 == "python3.12" && $0 ~ "/home/daytona/sandbox_runner.py" {{print $1; exit}}'
    }}

    if ! test -s "$pidfile"; then
      stale_pid=$(find_stale_runner)
      if test -n "$stale_pid"; then
        emit running "$stale_pid" null missing_pidfile_stale_process
      else
        emit missing_pidfile null null ""
      fi
      exit 0
    fi

    pid=$(head -n 1 "$pidfile" 2>/dev/null | tr -d '[:space:]')
    case "$pid" in
      ''|*[!0-9]*)
        emit unknown null null bad_pidfile
        exit 0
        ;;
    esac

    if kill -0 "$pid" 2>/dev/null; then
      # PID 1 in Daytona sandboxes is `daytona sleep infinity`, which does
      # not reap orphans; an exited runner can linger as a zombie while
      # `kill -0` still succeeds.
      proc_state=$(awk '{{print $3}}' /proc/"$pid"/stat 2>/dev/null)
      if test -n "$proc_state" && test "$proc_state" != "Z"; then
        emit running "$pid" null ""
      else
        emit missing_process "$pid" null zombie_or_missing_stat
      fi
    else
      emit missing_process "$pid" null ""
    fi
    """
    r = sandbox.process.exec("bash -lc " + shlex.quote(script), timeout=10)
    if r.exit_code not in (0, None):
        return RunnerProcessState("unknown", detail=f"probe_exit_{r.exit_code}")
    raw = (r.result or "").strip().splitlines()
    if not raw:
        return RunnerProcessState("unknown", detail="empty_probe_output")
    try:
        payload = json.loads(raw[-1])
    except json.JSONDecodeError:
        return RunnerProcessState("unknown", detail="malformed_probe_output")
    if not isinstance(payload, dict):
        return RunnerProcessState("unknown", detail="non_object_probe_output")
    state = payload.get("state")
    if state not in {
        "running",
        "exited_zero",
        "exited_nonzero",
        "missing_pidfile",
        "missing_process",
        "unknown",
    }:
        return RunnerProcessState("unknown", detail=f"bad_state_{state!r}")
    pid = payload.get("pid")
    exit_code = payload.get("exit_code")
    return RunnerProcessState(
        state=state,
        pid=pid if isinstance(pid, int) else None,
        exit_code=exit_code if isinstance(exit_code, int) else None,
        detail=str(payload.get("detail") or "") or None,
    )


def runner_alive(sandbox) -> bool:
    return runner_process_state(sandbox).state == "running"


def stop_existing_runner_process_group(sandbox) -> None:
    script = f"""
    set +e
    pidfile={shlex.quote(RUNNER_PIDFILE)}
    if test -s "$pidfile"; then
      pid=$(cat "$pidfile" 2>/dev/null)
      if test -n "$pid" && kill -0 "$pid" 2>/dev/null; then
        pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')
        if test -n "$pgid"; then
          kill -TERM "-$pgid" 2>/dev/null || true
          for _ in 1 2 3 4 5 6 7 8 9 10; do
            kill -0 "$pid" 2>/dev/null || break
            sleep 0.5
          done
          if kill -0 "$pid" 2>/dev/null; then
            kill -KILL "-$pgid" 2>/dev/null || true
          fi
        else
          kill -TERM "$pid" 2>/dev/null || true
          sleep 1
          kill -KILL "$pid" 2>/dev/null || true
        fi
      fi
      rm -f "$pidfile" {shlex.quote(RUNNER_EXITCODE_FILE)}
      exit 0
    fi

    stale=$(ps -eo pid=,comm=,args= | awk '$2 == "python3.12" && $0 ~ "/home/daytona/sandbox_runner.py" {{print $1}}')
    if test -n "$stale"; then
      echo "pidfile missing; stopping stale sandbox_runner.py"
      for pid in $stale; do
        pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')
        if test -n "$pgid"; then
          kill -TERM "-$pgid" 2>/dev/null || true
        else
          kill -TERM "$pid" 2>/dev/null || true
        fi
      done
      sleep 1
      for pid in $stale; do
        if kill -0 "$pid" 2>/dev/null; then
          pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')
          if test -n "$pgid"; then
            kill -KILL "-$pgid" 2>/dev/null || true
          else
            kill -KILL "$pid" 2>/dev/null || true
          fi
        fi
      done
    fi
    rm -f "$pidfile" {shlex.quote(RUNNER_EXITCODE_FILE)}
    exit 0
    """
    r = sandbox.process.exec("bash -lc " + shlex.quote(script), timeout=15)
    if r.exit_code not in (0, None):
        raise RuntimeError(f"failed to stop existing runner (exit {r.exit_code}); " f"output:\n{r.result or '<empty>'}")


def _runner_start_failure(sandbox, message: str) -> RuntimeError:
    with contextlib.suppress(Exception):
        stop_existing_runner_process_group(sandbox)
    log = fetch_runner_log(sandbox)
    return RuntimeError(f"{message}\n{RUNNER_LOG}:\n{log}")


def _env_float(name: str, default: float) -> float:
    raw = os.environ.get(name)
    if raw is None:
        return default
    return float(raw)


def _runner_max_idle_seconds(value: float | None) -> float:
    if value is not None:
        return float(value)
    return _env_float("RUNNER_MAX_IDLE_SECONDS", 300.0)


def _runner_launch_probe_seconds(value: float | None) -> float:
    return _env_float("RUNNER_LAUNCH_PROBE_SECONDS", 3.0) if value is None else float(value)


def _runner_replace_grace_seconds(value: float | None) -> float:
    return _env_float("RUNNER_REPLACE_GRACE_SECONDS", 10.0) if value is None else float(value)


def _clear_stale_runner_files(sandbox) -> None:
    r = sandbox.process.exec(
        f"rm -f {shlex.quote(RUNNER_PIDFILE)} {shlex.quote(RUNNER_EXITCODE_FILE)}",
        timeout=10,
    )
    if r.exit_code not in (0, None):
        raise RuntimeError(
            f"failed to clear stale runner files (exit {r.exit_code}); " f"output:\n{r.result or '<empty>'}"
        )


def _wait_for_existing_runner_to_exit(sandbox, grace_seconds: float) -> RunnerProcessState:
    state = runner_process_state(sandbox)
    if state.state != "running" or grace_seconds <= 0:
        return state
    deadline = time.monotonic() + grace_seconds
    while time.monotonic() < deadline:
        time.sleep(min(0.5, max(0.0, deadline - time.monotonic())))
        state = runner_process_state(sandbox)
        if state.state != "running":
            return state
    return state


def _handle_existing_runner_before_launch(
    sandbox,
    *,
    session_id: str,
    replace_grace_seconds: float,
    session_status_getter: Callable[[str], str | None] | None,
) -> None:
    state = _wait_for_existing_runner_to_exit(sandbox, replace_grace_seconds)
    if state.state == "running":
        if session_status_getter is None:
            raise RunnerLaunchDeferred(f"runner pid={state.pid} is live for session={session_id}; no status getter")
        try:
            status = session_status_getter(session_id)
        except Exception as e:
            raise RunnerLaunchDeferred(
                f"runner pid={state.pid} is live and session status is ambiguous: " f"{type(e).__name__}: {e}"
            ) from e

        if status in {"idle", "running"}:
            raise RunnerAlreadyAlive(f"runner pid={state.pid} already handling session={session_id} status={status!r}")
        raise RunnerLaunchDeferred(f"runner pid={state.pid} is live and session={session_id} has status={status!r}")

    if state.state == "unknown":
        raise RunnerLaunchDeferred(f"runner state is ambiguous for session={session_id}: {state.detail or 'unknown'}")

    _clear_stale_runner_files(sandbox)


def start_runner(
    sandbox,
    work_id: str,
    session_id: str,
    environment_key: str,
    environment_id: str,
    *,
    runner_max_idle_seconds: float | None = None,
    runner_launch_probe_seconds: float | None = None,
    runner_replace_grace_seconds: float | None = None,
    session_status_getter: Callable[[str], str | None] | None = None,
) -> bool:
    """Start a runner for this claimed work item after live-runner safety checks."""
    max_idle = _runner_max_idle_seconds(runner_max_idle_seconds)
    probe_seconds = _runner_launch_probe_seconds(runner_launch_probe_seconds)
    replace_grace = _runner_replace_grace_seconds(runner_replace_grace_seconds)
    _handle_existing_runner_before_launch(
        sandbox,
        session_id=session_id,
        replace_grace_seconds=replace_grace,
        session_status_getter=session_status_getter,
    )
    env = {
        "ANTHROPIC_ENVIRONMENT_ID": environment_id,
        "ANTHROPIC_WORK_ID": work_id,
        "ANTHROPIC_SESSION_ID": session_id,
        "ANTHROPIC_ENVIRONMENT_KEY": environment_key,
        "RUNNER_MAX_IDLE_SECONDS": str(max_idle),
        "ENVIRONMENT_ID": environment_id,
        "WORK_ID": work_id,
        "SESSION_ID": session_id,
    }
    runner_wrapper = f"""
      python3.12 /home/daytona/sandbox_runner.py
      code=$?
      printf '%s\\n' "$code" > {shlex.quote(RUNNER_EXITCODE_FILE)}
      exit "$code"
    """
    script = f"""
    set -eu
    cd /mnt/session
    rm -f {shlex.quote(RUNNER_PIDFILE)} {shlex.quote(RUNNER_EXITCODE_FILE)}
    setsid bash -lc {shlex.quote(runner_wrapper)} > {shlex.quote(RUNNER_LOG)} 2>&1 < /dev/null &
    runner_pid=$!
    printf '%s\\n' "$runner_pid" > {shlex.quote(RUNNER_PIDFILE)}
    printf '%s\\n' "$runner_pid"
    """
    # Stamp the work_id on the sandbox before we launch. If the janitor (or
    # an external call) stops the sandbox while the runner is mid-tool-call,
    # the runner's SIGKILL skips its `finally` and the work item lease lingers
    # on Anthropic's side. The orchestrator's stop paths read this label and
    # force-stop the leftover work item to keep the queue clean.
    merge_labels(sandbox, **{WORK_ID_LABEL: work_id})
    launched_at = time.monotonic()
    r = sandbox.process.exec("bash -lc " + shlex.quote(script), env=env, timeout=15)
    if r.exit_code not in (0, None):
        raise RuntimeError(f"sandbox runner launch failed (exit {r.exit_code}); " f"output:\n{r.result or '<empty>'}")
    try:
        launched_pid = int((r.result or "").strip().splitlines()[-1])
    except (IndexError, ValueError) as e:
        raise _runner_start_failure(
            sandbox,
            f"sandbox runner launch did not report a valid pid: {r.result!r}",
        ) from e
    print(
        f"[runner] launched work={work_id} pid={launched_pid} " f"launch={time.monotonic() - launched_at:.2f}s",
        flush=True,
    )
    time.sleep(probe_seconds)
    state = runner_process_state(sandbox)
    if state.state == "running" and state.detail != "missing_pidfile_stale_process":
        print(
            f"[runner] probe ok work={work_id} pid={launched_pid} "
            f"state={state.state} probe_delay={probe_seconds:.1f}s",
            flush=True,
        )
        return True
    if state.state == "exited_zero":
        print(
            f"[runner] probe complete work={work_id} pid={launched_pid} "
            f"state={state.state} probe_delay={probe_seconds:.1f}s",
            flush=True,
        )
        return True
    raise _runner_start_failure(
        sandbox,
        f"sandbox runner launch probe failed work={work_id} pid={launched_pid} "
        f"state={state.state} exit_code={state.exit_code!r} detail={state.detail!r}",
    )


def prepare_and_start_runner(
    sandbox,
    work_id: str,
    session_id: str,
    environment_key: str,
    environment_id: str,
    *,
    session_status_getter: Callable[[str], str | None] | None = None,
) -> bool:
    clear_stop_state(sandbox)
    _handle_existing_runner_before_launch(
        sandbox,
        session_id=session_id,
        replace_grace_seconds=_runner_replace_grace_seconds(None),
        session_status_getter=session_status_getter,
    )
    install_started = time.monotonic()
    install_runner(sandbox)
    print(f"[runner] install={time.monotonic() - install_started:.2f}s", flush=True)
    return start_runner(
        sandbox,
        work_id,
        session_id,
        environment_key,
        environment_id,
        session_status_getter=session_status_getter,
    )


def fetch_runner_log(sandbox) -> str:
    try:
        return sandbox.fs.download_file(RUNNER_LOG).decode(errors="replace")
    except Exception as e:
        return f"(could not fetch {RUNNER_LOG}: {type(e).__name__}: {e})"
