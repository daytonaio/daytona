# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# pylint: disable=no-member

"""Shared host-side orchestration for self-hosted Managed Agents work."""
from __future__ import annotations

import fcntl
import os
import pathlib
import threading
import time
from datetime import datetime, timedelta, timezone

import anthropic
import dotenv
import host_lib

from daytona import Daytona, DaytonaNotFoundError

dotenv.load_dotenv(override=True)

ENVIRONMENT_ID = os.environ["ENVIRONMENT_ID"]
ANTHROPIC_ENVIRONMENT_KEY = os.environ["ANTHROPIC_ENVIRONMENT_KEY"]
JANITOR_SECONDS = int(os.environ.get("JANITOR_SECONDS", "60"))
MAX_IDLE_DAYS = float(os.environ.get("MAX_IDLE_DAYS", "30"))

CLIENT = anthropic.Anthropic(auth_token=ANTHROPIC_ENVIRONMENT_KEY)
DAYT = Daytona()

DRAIN_LOCK = threading.RLock()
SESSION_LOCKS_LOCK = threading.Lock()
SESSION_LOCKS: dict[str, threading.RLock] = {}
shutdown = threading.Event()

_ORCHESTRATOR_LOCK_FD: int | None = None


def session_lock(session_id: str) -> threading.RLock:
    with SESSION_LOCKS_LOCK:
        lock = SESSION_LOCKS.get(session_id)
        if lock is None:
            lock = threading.RLock()
            SESSION_LOCKS[session_id] = lock
        return lock


def acquire_orchestrator_lock(mode: str) -> None:
    """Fail fast if another same-host orchestrator owns this environment."""
    global _ORCHESTRATOR_LOCK_FD  # pylint: disable=global-statement
    lock_path = pathlib.Path(
        os.environ.get(
            "ORCHESTRATOR_LOCK_FILE",
            f"/tmp/anthropic-selfhosted-orchestrator-{ENVIRONMENT_ID}.lock",
        )
    )
    fd = os.open(lock_path, os.O_CREAT | os.O_RDWR, 0o600)
    try:
        fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError as exc:
        existing = os.read(fd, 4096).decode(errors="replace")
        os.close(fd)
        raise RuntimeError(f"another orchestrator already holds {lock_path}: {existing or '<unknown>'}") from exc
    os.ftruncate(fd, 0)
    os.write(fd, f"pid={os.getpid()} mode={mode} env={ENVIRONMENT_ID}\n".encode())
    _ORCHESTRATOR_LOCK_FD = fd


def is_permanent_poll_error(err: Exception) -> bool:
    return (
        isinstance(err, anthropic.APIStatusError) and 400 <= err.status_code < 500 and err.status_code not in (408, 429)
    )


def work_created_at_key(work) -> tuple[float, str]:
    value = getattr(work, "created_at", None)
    try:
        dt = datetime.fromisoformat(str(value).replace("Z", "+00:00"))
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        created_ts = dt.timestamp()
    except Exception:
        print(
            f"[poll] work {getattr(work, 'id', '<unknown>')} has unparsable created_at={value!r}",
            flush=True,
        )
        created_ts = float("-inf")
    return created_ts, getattr(work, "id", "")


def claim_work_once(
    *,
    block_ms: int | None,
    reclaim_older_than_ms: int | None,
):
    poll_kwargs = {"timeout": 30.0}
    if block_ms is not None:
        poll_kwargs["block_ms"] = block_ms
    if reclaim_older_than_ms is not None:
        poll_kwargs["reclaim_older_than_ms"] = reclaim_older_than_ms

    item = CLIENT.beta.environments.work.poll(ENVIRONMENT_ID, **poll_kwargs)
    if item is None:
        return None
    print(
        f"[poll] claimed work={item.id} session={getattr(item.data, 'id', '?')} "
        f"type={getattr(item.data, 'type', '?')}",
        flush=True,
    )
    try:
        CLIENT.beta.environments.work.ack(
            item.id,
            environment_id=ENVIRONMENT_ID,
            timeout=30.0,
        )
    except Exception as e:
        print(f"[poll] work={item.id} ack failed: {type(e).__name__}: {e}", flush=True)
        return "skip"
    print(f"[poll] ack ok work={item.id}", flush=True)
    return item


def release_in_flight_work_for(sb) -> None:
    """Force-stop the queued work item this sandbox was running, if any.

    The runner stamps `byoc.work_id` at launch. If the orchestrator stops
    the sandbox externally (idle path, crashed-runner path, anything that
    SIGKILLs the runner before its `finally` can call `work.stop`), the
    work-item lease lingers and blocks Anthropic from enqueuing fresh
    work for the same session. Calling this before `stop_sandbox` keeps
    the queue clean.
    """
    labels = getattr(sb, "labels", None) or {}
    work_id = labels.get(host_lib.WORK_ID_LABEL)
    if not work_id:
        return
    try:
        CLIENT.beta.environments.work.stop(
            work_id,
            environment_id=ENVIRONMENT_ID,
            force=True,
            timeout=30.0,
        )
    except anthropic.ConflictError:
        pass  # already stopped; nothing to do
    except Exception as e:
        print(
            f"[work] force-stop on sandbox-stop failed for work={work_id}: " f"{type(e).__name__}: {e}",
            flush=True,
        )


def stop_work_item(work) -> bool:
    # force=True transitions the work item to state=stopped immediately.
    # Without force, items stick in state=stopping and block the server from
    # enqueuing a new work item for the same session on the next prompt.
    try:
        CLIENT.beta.environments.work.stop(
            work.id,
            environment_id=ENVIRONMENT_ID,
            force=True,
            timeout=30.0,
        )
        return True
    except anthropic.ConflictError:
        print(f"[work] stop skipped work={work.id}: already stopped", flush=True)
        return True
    except Exception as e:
        print(f"[work] stop failed work={work.id}: {type(e).__name__}: {e}", flush=True)
        return False


def stop_claimed_work_item(work, reason: str) -> bool:
    stopped = stop_work_item(work)
    if stopped:
        print(f"[poll] claimed work={work.id} stopped ({reason})", flush=True)
    return stopped


def session_status_for_launch(session_id: str) -> str | None:
    session = CLIENT.beta.sessions.retrieve(session_id, timeout=10.0)
    return getattr(session, "status", None)


def ensure_session_runner(work) -> dict:
    session_id = work.data.id
    with session_lock(session_id):
        # Skip if the session is already terminated. Late-arriving webhooks
        # for archived sessions otherwise spawn a runner that immediately
        # sees session.status_terminated and exits, wasting a sandbox start.
        try:
            session = CLIENT.beta.sessions.retrieve(session_id, timeout=10.0)
            if getattr(session, "status", None) == "terminated":
                print(f"[ensure] {session_id} terminated; skipping", flush=True)
                stop_work_item(work)
                return {"session_id": session_id, "skipped": "terminated"}
            metadata = getattr(session, "metadata", None) or {}
            if not isinstance(metadata, dict):
                raise TypeError(f"session metadata must be a dict, got {type(metadata).__name__}")
        except anthropic.NotFoundError:
            print(f"[ensure] {session_id} not found; skipping", flush=True)
            stop_work_item(work)
            return {"session_id": session_id, "skipped": "not_found"}
        except Exception as e:
            print(
                f"[ensure] {session_id} session metadata check failed " f"({type(e).__name__}: {e}); stopping",
                flush=True,
            )
            stop_work_item(work)
            raise

        try:
            sb = host_lib.get_or_create_sandbox(DAYT, session_id, ENVIRONMENT_ID, metadata=metadata)
            t0 = time.monotonic()
            host_lib.ensure_started(sb)
            print(f"[ensure] {session_id} sandbox_start={time.monotonic() - t0:.2f}s", flush=True)
            started = host_lib.prepare_and_start_runner(
                sb,
                work.id,
                session_id,
                ANTHROPIC_ENVIRONMENT_KEY,
                ENVIRONMENT_ID,
                session_status_getter=session_status_for_launch,
            )
        except host_lib.SandboxResurrectionError as e:
            print(f"[ensure] {session_id} sandbox mode mismatch; skipping: {e}", flush=True)
            stop_work_item(work)
            return {"session_id": session_id, "skipped": "terminal_sandbox"}
        except host_lib.RunnerAlreadyAlive as e:
            print(f"[ensure] {session_id} runner already alive; stopping duplicate work: {e}", flush=True)
            stop_work_item(work)
            return {"session_id": session_id, "skipped": "already_running"}
        except host_lib.RunnerLaunchDeferred as e:
            print(f"[ensure] {session_id} runner launch deferred; stopping duplicate work: {e}", flush=True)
            stop_work_item(work)
            return {"session_id": session_id, "skipped": "runner_ambiguous"}
        except Exception:
            stop_work_item(work)
            raise
        info = {
            "session_id": session_id,
            "work_id": work.id,
            "sandbox_id": sb.id,
            "runner_started": started,
        }
        print(f"[ensure] {info}", flush=True)
        return info


def drain_work(
    max_items: int = 25,
    target_session_id: str | None = None,
    *,
    block_ms: int | None = None,
    reclaim_older_than_ms: int | None = 2000,
    raise_poll_errors: bool = False,
) -> list[dict]:
    drained = []
    poll_error: Exception | None = None

    with DRAIN_LOCK:
        while not shutdown.is_set():
            try:
                claimed = claim_work_once(
                    block_ms=block_ms,
                    reclaim_older_than_ms=reclaim_older_than_ms,
                )
            except Exception as e:
                print(f"[poll] failed: {type(e).__name__}: {e}", flush=True)
                poll_error = e
                break

            if claimed is None:
                break
            if claimed == "skip":
                continue

            work = claimed
            if work.data.type != "session":
                print(f"[poll] stop non-session work={work.id} type={work.data.type}", flush=True)
                stop_claimed_work_item(work, "non-session work")
                continue
            if target_session_id is not None and work.data.id != target_session_id:
                print(
                    f"[poll] recovery for session={target_session_id} also claimed "
                    f"work={work.id} session={work.data.id}; processing claimed work",
                    flush=True,
                )
            drained.append(work)
            if len(drained) >= max_items:
                print(
                    f"[poll] max_items={max_items} reached; " "leaving remaining queue items for the next drain",
                    flush=True,
                )
                break

        by_session = {}
        for work in drained:
            session_id = work.data.id
            current = by_session.get(session_id)
            if current is None or work_created_at_key(work) > work_created_at_key(current):
                by_session[session_id] = work

        kept_work_ids = {work.id for work in by_session.values()}
        for work in drained:
            if work.id not in kept_work_ids:
                stop_claimed_work_item(work, f"newer work kept for session {work.data.id}")

        spawned = []
        for work in by_session.values():
            try:
                spawned.append(ensure_session_runner(work))
            except Exception as e:
                print(f"[ensure] failed for {work.data.id}: {type(e).__name__}: {e}", flush=True)

        if poll_error is not None and raise_poll_errors:
            raise poll_error
        return spawned


def janitor_once(*, recover_crashed_runners: bool = True) -> None:
    """One pass over labeled Daytona sandboxes."""
    page_number = 1
    deleted = 0
    archived = 0
    stopped = 0
    restarted = 0
    backfilled = 0
    recovery_requests: list[tuple[str, str]] = []

    while True:
        try:
            page = DAYT.list(
                labels={host_lib.ENVIRONMENT_ID_LABEL: ENVIRONMENT_ID},
                page=page_number,
                limit=100,
            )
        except Exception as e:
            print(f"[janitor] daytona list failed: {type(e).__name__}: {e}", flush=True)
            return

        for sb in page.items:
            listed_sid = (getattr(sb, "labels", None) or {}).get(host_lib.SESSION_ID_LABEL)
            if not listed_sid:
                continue
            with session_lock(listed_sid):
                try:
                    sb.refresh_data()
                except Exception as e:
                    print(f"[janitor] refresh {sb.id} failed: {type(e).__name__}: {e}", flush=True)
                    continue

                labels = getattr(sb, "labels", None) or {}
                sid = labels.get(host_lib.SESSION_ID_LABEL)
                if not sid:
                    continue
                if sid != listed_sid:
                    print(
                        f"[janitor] sandbox {sb.id} session label changed "
                        f"{listed_sid!r} -> {sid!r}; skipping this pass",
                        flush=True,
                    )
                    continue

                if labels.get(host_lib.ENVIRONMENT_ID_LABEL) != ENVIRONMENT_ID:
                    print(
                        f"[janitor] sandbox {sb.id} environment label changed; skipping this pass",
                        flush=True,
                    )
                    continue

                mode = labels.get(host_lib.MODE_LABEL)
                if mode == host_lib.MODE_IN_SANDBOX:
                    d, s, a, b = _janitor_handle_in_sandbox(
                        sb,
                        labels,
                        sid,
                        recovery_requests,
                        recover_crashed_runners=recover_crashed_runners,
                    )
                elif mode == host_lib.MODE_TERMINAL:
                    d, s, a, b = _janitor_handle_terminal(sb, labels, sid)
                elif mode == host_lib.MODE_PREPARED:
                    continue
                else:
                    print(f"[janitor] sandbox {sb.id} has unknown mode {mode!r}; skipping", flush=True)
                    continue

                deleted += d
                stopped += s
                archived += a
                backfilled += b

        if page_number >= int(page.total_pages):
            break
        page_number += 1

    for sid, sandbox_id in recovery_requests:
        try:
            recovered = drain_work(max_items=5, target_session_id=sid)
        except Exception as e:
            print(f"[janitor] recover {sandbox_id} failed: {type(e).__name__}: {e}", flush=True)
            continue

        with session_lock(sid):
            try:
                sb = DAYT.get(sandbox_id)
            except DaytonaNotFoundError:
                print(f"[janitor] recovery sandbox {sandbox_id} no longer exists", flush=True)
                continue
            except Exception as e:
                print(f"[janitor] recovery get {sandbox_id} failed: {type(e).__name__}: {e}", flush=True)
                continue
            try:
                sb.refresh_data()
                state = getattr(sb, "state", None)
                alive = state == "started" and host_lib.runner_alive(sb)
            except Exception as e:
                print(f"[janitor] recovery refresh/probe {sandbox_id} failed: {type(e).__name__}: {e}", flush=True)
                alive = False

            if alive:
                restarted += 1
                print(
                    f"[janitor] recovered runner for sandbox {sandbox_id} " f"(session {sid}, result={recovered})",
                    flush=True,
                )
                continue
            try:
                session = CLIENT.beta.sessions.retrieve(sid, timeout=10.0)
                status = getattr(session, "status", None)
                archived_at = getattr(session, "archived_at", None)
            except anthropic.NotFoundError:
                print(
                    f"[janitor] recovery session {sid} no longer exists; " "next janitor pass will archive the sandbox",
                    flush=True,
                )
                continue
            except Exception as e:
                print(
                    f"[janitor] recovery status check {sandbox_id} failed: " f"{type(e).__name__}: {e}",
                    flush=True,
                )
                continue

            if archived_at or status == "terminated":
                print(
                    f"[janitor] recovery session {sid} is terminal; " "next janitor pass will archive the sandbox",
                    flush=True,
                )
            elif status == "idle":
                print(
                    f"[janitor] recovery session {sid} is idle; " "deferring to normal idle handling",
                    flush=True,
                )
            elif status == "running":
                print(
                    f"[janitor] recovery still pending for sandbox {sandbox_id} "
                    f"(session {sid}); leaving sandbox started",
                    flush=True,
                )
            else:
                print(
                    f"[janitor] recovery saw unknown session status {status!r} "
                    f"for sandbox {sandbox_id}; leaving unchanged",
                    flush=True,
                )

    if deleted or archived or stopped or restarted or backfilled:
        print(
            f"[janitor] cleaned up deleted={deleted} archived={archived} stopped={stopped} "
            f"restarted={restarted} backfilled={backfilled}",
            flush=True,
        )


def _janitor_handle_in_sandbox(
    sb,
    labels: dict,
    sid: str,
    recovery_requests: list[tuple[str, str]],
    *,
    recover_crashed_runners: bool,
) -> tuple[int, int, int, int]:
    deleted = 0
    stopped = 0
    archived = 0
    backfilled = 0

    state = getattr(sb, "state", None)
    if state == "started":
        try:
            if host_lib.runner_alive(sb):
                return deleted, stopped, archived, backfilled
        except Exception as e:
            print(f"[janitor] runner probe {sb.id} failed: {type(e).__name__}: {e}", flush=True)
            return deleted, stopped, archived, backfilled

        try:
            session = CLIENT.beta.sessions.retrieve(sid, timeout=10.0)
        except anthropic.NotFoundError:
            try:
                host_lib.merge_labels(sb, **{host_lib.WORK_ID_LABEL: None})
                host_lib.archive_sandbox(sb)
                archived += 1
                print(f"[janitor] archived sandbox {sb.id} for missing session {sid}", flush=True)
            except Exception as e:
                print(f"[janitor] archive missing-session {sb.id} failed: {type(e).__name__}: {e}", flush=True)
            return deleted, stopped, archived, backfilled
        except Exception as e:
            print(f"[janitor] retrieve session {sid} failed: {type(e).__name__}: {e}", flush=True)
            return deleted, stopped, archived, backfilled

        status = getattr(session, "status", None)
        if getattr(session, "archived_at", None) or status == "terminated":
            try:
                host_lib.merge_labels(sb, **{host_lib.WORK_ID_LABEL: None})
                host_lib.archive_sandbox(sb)
                archived += 1
                print(f"[janitor] archived terminal sandbox {sb.id} (session {sid})", flush=True)
            except Exception as e:
                print(f"[janitor] archive {sb.id} failed: {type(e).__name__}: {e}", flush=True)
            return deleted, stopped, archived, backfilled

        if status == "idle":
            try:
                host_lib.merge_labels(sb, **{host_lib.WORK_ID_LABEL: None})
                host_lib.stop_sandbox(sb)
                stopped += 1
                print(f"[janitor] stopped idle sandbox {sb.id} (session {sid})", flush=True)
            except Exception as e:
                print(f"[janitor] idle handling {sb.id} failed: {type(e).__name__}: {e}", flush=True)
            return deleted, stopped, archived, backfilled

        if status == "running":
            try:
                release_in_flight_work_for(sb)
            except Exception as e:
                print(f"[janitor] release crashed work {sb.id} failed: {type(e).__name__}: {e}", flush=True)
            if recover_crashed_runners:
                print(
                    f"[janitor] runner missing for running session {sid} "
                    f"on sandbox {sb.id}; polling for replacement work",
                    flush=True,
                )
                recovery_requests.append((sid, sb.id))
            else:
                print(
                    f"[janitor] runner missing for running session {sid} " "and polling loop owns work recovery",
                    flush=True,
                )
            return deleted, stopped, archived, backfilled

        print(
            f"[janitor] sandbox {sb.id} session {sid} has unknown status {status!r}; skipping",
            flush=True,
        )
        return deleted, stopped, archived, backfilled

    if state in ("stopped", "archived"):
        deleted, backfilled = _janitor_reap_by_stopped_at(sb, labels, sid, delete_reason="long-idle")

    return deleted, stopped, archived, backfilled


def _janitor_handle_terminal(sb, labels: dict, sid: str) -> tuple[int, int, int, int]:
    deleted, backfilled = _janitor_reap_by_stopped_at(sb, labels, sid, delete_reason="terminal")
    return deleted, 0, 0, backfilled


def _janitor_reap_by_stopped_at(sb, labels: dict, sid: str, *, delete_reason: str) -> tuple[int, int]:
    deleted = 0
    backfilled = 0
    stopped_at = labels.get(host_lib.STOPPED_AT_LABEL)
    if not stopped_at:
        try:
            host_lib.merge_labels(sb, **{host_lib.STOPPED_AT_LABEL: host_lib.iso8601_utc_now()})
            backfilled += 1
            print(f"[janitor] backfilled stopped_at for sandbox {sb.id}", flush=True)
        except Exception as e:
            print(f"[janitor] backfill {sb.id} failed: {type(e).__name__}: {e}", flush=True)
        return deleted, backfilled

    if MAX_IDLE_DAYS <= 0:
        return deleted, backfilled

    try:
        stopped_at_dt = _parse_utc_datetime(stopped_at)
    except ValueError as e:
        print(
            f"[janitor] invalid stopped_at for sandbox {sb.id}: {stopped_at!r} ({e}); " "backfilling from now",
            flush=True,
        )
        try:
            host_lib.merge_labels(sb, **{host_lib.STOPPED_AT_LABEL: host_lib.iso8601_utc_now()})
            backfilled += 1
        except Exception as label_error:
            print(
                f"[janitor] backfill {sb.id} failed: " f"{type(label_error).__name__}: {label_error}",
                flush=True,
            )
        return deleted, backfilled

    if datetime.now(timezone.utc) - stopped_at_dt >= timedelta(days=MAX_IDLE_DAYS):
        try:
            sb.delete(timeout=120)
            deleted += 1
            print(f"[janitor] deleted {delete_reason} sandbox {sb.id} (session {sid})", flush=True)
        except Exception as e:
            print(f"[janitor] delete {sb.id} failed: {type(e).__name__}: {e}", flush=True)

    return deleted, backfilled


def _parse_utc_datetime(value: str) -> datetime:
    dt = datetime.fromisoformat(value.replace("Z", "+00:00"))
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def janitor_loop(*, recover_crashed_runners: bool = True) -> None:
    while not shutdown.wait(JANITOR_SECONDS):
        try:
            janitor_once(recover_crashed_runners=recover_crashed_runners)
        except Exception as e:
            print(f"[janitor] pass failed: {type(e).__name__}: {e}", flush=True)
