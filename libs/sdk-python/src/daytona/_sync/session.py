# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Synchronous SessionService — the user-facing surface for the Sessions product.

The hot path (`run` / `run_stream`) skips the Daytona API and talks directly to
the in-sandbox session-daemon via the proxy chain — same shape
`sandbox.process.code_run` uses against the classic daytona-daemon. The API is
only hit on `createSession` / first one-shot / token refresh (every ~5 min) /
on `SessionInvalidatedError`. See the SDK plan for the full design.

The legacy `POST /sessions/code-run` and `POST /sessions/connect` paths are
preserved as transparent fallbacks for older API servers that don't yet
expose `/access` / `/transients`.
"""

# The optional websocket-client dependency ships no type stubs, so its symbols are
# untyped here; scoped to this file to keep the rest of the SDK strict.
# pyright: reportUnknownVariableType=false, reportUnknownArgumentType=false

from __future__ import annotations

import json
import time
from datetime import datetime, timedelta, timezone
from typing import Any, Callable, Literal, Optional, cast

import requests

from ..common.errors import DaytonaError
from ..common.session import (
    SessionAccess,
    SessionDisplay,
    SessionExecutionError,
    SessionExpiredError,
    SessionInvalidatedError,
    SessionRef,
    SessionRunOptions,
    SessionRunResult,
)

# Refresh `SessionAccess` this many seconds before the token TTL hits — keeps
# the SDK from racing the proxy's expiry check on a tight in-flight call.
_REFRESH_SKEW_SECONDS = 60


def _parse_iso(s: str) -> datetime:
    """Parse an ISO 8601 string into an aware UTC datetime, tolerating trailing Z."""
    if not s:
        return datetime.now(timezone.utc)
    return datetime.fromisoformat(s.replace("Z", "+00:00"))


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def _finalize_disconnect(
    aggregated: SessionRunResult,
    completed: bool,
    started_at: float,
    exc: Optional[BaseException] = None,
) -> SessionRunResult:
    """Finalize a streamed run: surface an abnormal disconnect and stamp duration.

    `completed` is True once the daemon's terminal control frame
    ("completed"/"interrupted") was seen. If the stream ended — or `exc` was
    raised — before that frame, the run disconnected abnormally and we surface a
    `SessionDisconnected` error, unless the daemon already reported a real error.
    """
    if not completed:
        disconnect_error = (
            f"session websocket disconnected mid-run: {exc}"
            if exc is not None
            else "session websocket closed before run completed"
        )
        if aggregated.error is None:
            aggregated.error = SessionExecutionError(name="SessionDisconnected", value=disconnect_error)
    aggregated.duration_ms = int((time.time() - started_at) * 1000)
    return aggregated


class _LegacyFallback(Exception):
    """Signal: API endpoint not available on this server; degrade to /code-run."""


class _WSAuthError(Exception):
    """Signal: WS handshake returned 401/403; refresh access bundle and retry once."""


class SessionService:
    """Sync API surface for the Sessions product.

    The service is constructed by the top-level `Daytona` client; users never
    instantiate it directly.
    """

    _base_url: str
    _headers: dict[str, str]

    def __init__(self, base_url: str, headers: dict[str, str]):
        self._base_url = base_url.rstrip("/")
        self._headers = dict(headers)
        _ = self._headers.setdefault("Content-Type", "application/json")
        # SDK-side caches for direct-to-sandbox access bundles. Keyed by:
        # - context id for persistent contexts
        # - "{template}:{language}" for one-shot transients (empty parts ok)
        self._ctx_access: dict[str, SessionAccess] = {}
        self._transient_access: dict[str, tuple[str, SessionAccess]] = {}
        # Latches once the API confirms `POST /sessions/transients` is missing
        # (older server). Subsequent one-shots short-circuit to /code-run.
        self._transient_supported: Optional[bool] = None

    # -- one-shot run ----------------------------------------------------

    def run(self, code: str, options: Optional[SessionRunOptions] = None) -> SessionRunResult:
        """Execute `code` once and return aggregated stdout / stderr / displays.

        Talks straight to the in-sandbox daemon via a cached signed proxy URL.
        Falls back to the legacy `POST /sessions/code-run` API path on older
        servers that don't yet expose the direct surface.
        """
        return self._run_internal(code, options, handlers=None)

    # -- streaming run ---------------------------------------------------

    def run_stream(
        self,
        code: str,
        options: Optional[SessionRunOptions] = None,
        on_stdout: Optional[Callable[[str], None]] = None,
        on_stderr: Optional[Callable[[str], None]] = None,
        on_error: Optional[Callable[[SessionExecutionError], None]] = None,
        on_display: Optional[Callable[[SessionDisplay], None]] = None,
        on_control: Optional[Callable[[str], None]] = None,
    ) -> SessionRunResult:
        """Stream code execution; per-frame handlers fire as frames arrive.

        Same direct-to-sandbox path as `run()`. Returns an aggregated result
        after the daemon emits a terminal `control` frame and the WS closes.
        """
        return self._run_internal(
            code,
            options,
            handlers={
                "on_stdout": on_stdout,
                "on_stderr": on_stderr,
                "on_error": on_error,
                "on_display": on_display,
                "on_control": on_control,
            },
        )

    # -- context CRUD ----------------------------------------------------

    def create_session(
        self,
        template: Optional[str] = None,
        language: Optional[str] = None,
        cwd: Optional[str] = None,
    ) -> SessionRef:
        body: dict[str, Any] = {}
        if template is not None:
            body["template"] = template
        if language is not None:
            body["language"] = language
        if cwd is not None:
            body["cwd"] = cwd
        resp = requests.post(
            f"{self._base_url}/sessions",
            headers=self._headers,
            data=json.dumps(body),
            timeout=120,
        )
        if not resp.ok:
            raise DaytonaError(
                f"create_session returned {resp.status_code}: {resp.text}",
                status_code=resp.status_code,
            )
        ctx = self._unmarshal_context(resp.json())
        # Prime the cache so the very first run() against this context skips
        # the GET /access round-trip.
        if ctx.access is not None:
            self._ctx_access[ctx.id] = ctx.access
        return ctx

    def list_sessions(self, template: Optional[str] = None) -> list[SessionRef]:
        params = {"template": template} if template else None
        resp = requests.get(f"{self._base_url}/sessions", headers=self._headers, params=params, timeout=30)
        if not resp.ok:
            raise DaytonaError(f"list_sessions returned {resp.status_code}", status_code=resp.status_code)
        return [self._unmarshal_context(c) for c in resp.json()]

    def delete_session(self, session_id: str) -> None:
        resp = requests.delete(
            f"{self._base_url}/sessions/{session_id}",
            headers=self._headers,
            timeout=30,
        )
        if resp.status_code not in (204, 404):
            raise DaytonaError(f"delete_session returned {resp.status_code}", status_code=resp.status_code)
        # Evict any cached access for this id so a recreated context with the
        # same id doesn't reuse a stale signed URL.
        _ = self._ctx_access.pop(session_id, None)

    def list_templates(self) -> list[dict[str, Any]]:
        resp = requests.get(f"{self._base_url}/sessions/templates", headers=self._headers, timeout=30)
        resp.raise_for_status()
        return resp.json()

    def list_packages(self, template_name: str, language: str) -> list[dict[str, Any]]:
        resp = requests.get(
            f"{self._base_url}/sessions/templates/{template_name}/packages",
            headers=self._headers,
            params={"language": language},
            timeout=30,
        )
        resp.raise_for_status()
        return resp.json()

    # -- direct-to-sandbox hot path -------------------------------------

    def _run_internal(
        self,
        code: str,
        options: Optional[SessionRunOptions],
        handlers: Optional[dict[str, Optional[Callable[..., Any]]]],
    ) -> SessionRunResult:
        # If we already know the API doesn't expose /transients (older server)
        # and the caller didn't pin a context, skip straight to the legacy path.
        if (options is None or options.context is None) and self._transient_supported is False:
            return self._run_legacy(code, options, handlers)

        try:
            ctx_id, access, reset = self._ensure_access(options)
        except _LegacyFallback:
            if options is None or options.context is None:
                self._transient_supported = False
            return self._run_legacy(code, options, handlers)

        try:
            return self._run_ws_direct(ctx_id, access, code, options, reset, handlers)
        except _WSAuthError:
            # Stale signed URL — drop the cache entry and refresh once.
            self._evict_access(options, ctx_id)
            try:
                ctx_id, access, reset = self._ensure_access(options)
            except _LegacyFallback:
                if options is None or options.context is None:
                    self._transient_supported = False
                return self._run_legacy(code, options, handlers)
            return self._run_ws_direct(ctx_id, access, code, options, reset, handlers)

    def _ensure_access(self, options: Optional[SessionRunOptions]) -> tuple[str, SessionAccess, bool]:
        """Return (session_id, access, reset_flag). Refreshes near-expiry handles.

        Raises `_LegacyFallback` if the corresponding API endpoint is missing
        (older server) so the caller can degrade to /code-run.
        """
        if options is not None and options.context is not None:
            ctx_id = options.context.id
            # Inline access from a freshly-created context primes the cache —
            # avoids an immediate GET /access right after createSession().
            if options.context.access is not None and ctx_id not in self._ctx_access:
                self._ctx_access[ctx_id] = options.context.access

            access = self._ctx_access.get(ctx_id)
            if access is not None and not self._is_expired(access):
                return ctx_id, access, False

            access = self._fetch_session_access(ctx_id)
            self._ctx_access[ctx_id] = access
            return ctx_id, access, False

        tpl = options.template if options else None
        lang = options.language if options else None
        key = f"{tpl or ''}:{lang or ''}"
        cached = self._transient_access.get(key)
        if cached is not None and not self._is_expired(cached[1]):
            return cached[0], cached[1], True

        ctx_id, access = self._fetch_transient(tpl, lang)
        self._transient_access[key] = (ctx_id, access)
        self._transient_supported = True
        return ctx_id, access, True

    def _run_ws_direct(
        self,
        ctx_id: str,
        access: SessionAccess,
        code: str,
        options: Optional[SessionRunOptions],
        reset: bool,
        handlers: Optional[dict[str, Optional[Callable[..., Any]]]],
    ) -> SessionRunResult:
        try:
            # pylint: disable-next=import-outside-toplevel
            from websocket import (  # pyright: ignore[reportMissingImports]
                WebSocket,
                WebSocketBadStatusException,
                WebSocketException,
            )
        except ImportError as exc:
            msg = (
                "direct session execution requires the 'websocket-client' package; "
                "install with `pip install websocket-client`"
            )
            raise DaytonaError(msg) from exc

        ws = WebSocket()
        try:
            ws.connect(access.ws_url)
        except WebSocketBadStatusException as exc:
            status = getattr(exc, "status_code", None)
            if status in (401, 403):
                # Stale / revoked signed URL — caller will refresh & retry once.
                raise _WSAuthError() from exc
            # Any other handshake error — 404 (context gone from daemon), 400
            # (proxy: "Is the Sandbox started?"), 5xx — surfaces as an
            # invalidation. The caller's recovery action is the same: drop
            # the context and create a fresh one.
            raise SessionInvalidatedError(ctx_id, _now_iso()) from exc
        except (ConnectionRefusedError, WebSocketException, OSError) as exc:
            # Sandbox rolled / network gone — match the legacy contract.
            raise SessionInvalidatedError(ctx_id, _now_iso()) from exc

        try:
            _ = ws.send(
                json.dumps(
                    {
                        "code": code,
                        "envs": (options.env if options else None) or {},
                        "timeout": (options.timeout if options else None) or 0,
                        "reset": reset,
                    }
                )
            )
            return self._consume_ws(ws, handlers)
        finally:
            try:
                ws.close()
            except Exception:  # noqa: BLE001 — close errors are non-fatal here.
                pass

    def _consume_ws(
        self,
        ws: Any,
        handlers: Optional[dict[str, Optional[Callable[..., Any]]]],
    ) -> SessionRunResult:
        aggregated = SessionRunResult(stdout="", stderr="", error=None, displays=[], duration_ms=0)
        started_at = time.time()
        # The daemon emits a terminal control frame ("completed"/"interrupted")
        # to signal a clean end-of-run. A recv error / stream end BEFORE that
        # frame is an abnormal disconnect and must surface as an error, not a
        # silent success.
        completed = False
        disconnect_exc: Optional[BaseException] = None
        while True:
            try:
                raw = ws.recv()
            except Exception as exc:  # noqa: BLE001
                disconnect_exc = exc
                break
            if raw is None or raw == "":
                # End of stream. Clean only if we already saw the terminal frame.
                break
            try:
                frame = json.loads(raw)
            except json.JSONDecodeError:
                continue
            if frame.get("type") == "control" and frame.get("text") in ("completed", "interrupted"):
                completed = True
            self._apply_frame(frame, aggregated, handlers)
        return _finalize_disconnect(aggregated, completed, started_at, disconnect_exc)

    def _fetch_session_access(self, session_id: str) -> SessionAccess:
        try:
            resp = requests.get(
                f"{self._base_url}/sessions/{session_id}/access",
                headers=self._headers,
                timeout=30,
            )
        except requests.RequestException as exc:
            raise DaytonaError(f"session access request failed: {exc}") from exc

        if resp.status_code == 404:
            # Could be "endpoint missing" (older API) or "context missing". We
            # distinguish by looking at the body shape — the API returns a JSON
            # error object for missing contexts; the route layer returns a
            # plain Nest 404 for unknown routes.
            if self._is_route_missing_404(resp):
                raise _LegacyFallback()
            raise SessionInvalidatedError(session_id, _now_iso())
        if resp.status_code == 410:
            raise self._translate_410(resp)
        if not resp.ok:
            raise DaytonaError(
                f"session access returned {resp.status_code}: {resp.text}",
                status_code=resp.status_code,
            )
        return self._unmarshal_access(resp.json())

    def _fetch_transient(
        self,
        template: Optional[str],
        language: Optional[str],
    ) -> tuple[str, SessionAccess]:
        body: dict[str, Any] = {}
        if template is not None:
            body["template"] = template
        if language is not None:
            body["language"] = language
        try:
            resp = requests.post(
                f"{self._base_url}/sessions/transients",
                headers=self._headers,
                data=json.dumps(body),
                timeout=120,
            )
        except requests.RequestException as exc:
            raise DaytonaError(f"session transient request failed: {exc}") from exc

        if resp.status_code == 404 and self._is_route_missing_404(resp):
            raise _LegacyFallback()
        if not resp.ok:
            raise DaytonaError(
                f"session transient returned {resp.status_code}: {resp.text}",
                status_code=resp.status_code,
            )
        ctx = self._unmarshal_context(resp.json())
        if ctx.access is None:
            # Server returned a transient handle without access — older
            # behavior or misconfiguration. Treat as fallback-worthy.
            raise _LegacyFallback()
        return ctx.id, ctx.access

    def _run_legacy(
        self,
        code: str,
        options: Optional[SessionRunOptions],
        handlers: Optional[dict[str, Optional[Callable[..., Any]]]],
    ) -> SessionRunResult:
        """Pre-direct-access SDK code path. Used as transparent fallback.

        For `run()`, posts to `/sessions/code-run` and aggregates the response.
        For `run_stream()`, posts to `/sessions/connect` and streams the WS.
        """
        if handlers is not None:
            return self._run_stream_legacy(code, options, handlers)
        return self._run_code_run_legacy(code, options)

    def _run_code_run_legacy(
        self,
        code: str,
        options: Optional[SessionRunOptions],
    ) -> SessionRunResult:
        body = self._build_run_body(code, options)
        try:
            resp = requests.post(
                f"{self._base_url}/sessions/code-run",
                headers=self._headers,
                data=json.dumps(body),
                timeout=600,
            )
        except requests.RequestException as exc:
            raise DaytonaError(f"session run request failed: {exc}") from exc

        if resp.status_code == 410:
            raise self._translate_410(resp)
        if not resp.ok:
            raise DaytonaError(
                f"session run returned {resp.status_code}: {resp.text}",
                status_code=resp.status_code,
            )
        data: dict[str, Any] = resp.json()
        displays_raw: list[dict[str, Any]] = data.get("displays") or []
        return SessionRunResult(
            stdout=data.get("stdout", ""),
            stderr=data.get("stderr", ""),
            error=self._unmarshal_error(data.get("error")),
            displays=[self._unmarshal_display(d) for d in displays_raw],
            duration_ms=int(data.get("durationMs", 0)),
        )

    def _run_stream_legacy(
        self,
        code: str,
        options: Optional[SessionRunOptions],
        handlers: dict[str, Optional[Callable[..., Any]]],
    ) -> SessionRunResult:
        try:
            # pylint: disable-next=import-outside-toplevel
            from websocket import WebSocket  # pyright: ignore[reportMissingImports]
        except ImportError as exc:
            raise DaytonaError(
                "run_stream requires the 'websocket-client' package; install with `pip install websocket-client`"
            ) from exc

        connect_body = self._build_connect_body(options)
        try:
            connect_resp = requests.post(
                f"{self._base_url}/sessions/connect",
                headers=self._headers,
                data=json.dumps(connect_body),
                timeout=60,
            )
        except requests.RequestException as exc:
            raise DaytonaError(f"session connect request failed: {exc}") from exc

        if connect_resp.status_code == 410:
            raise self._translate_410(connect_resp)
        if not connect_resp.ok:
            raise DaytonaError(
                f"session connect returned {connect_resp.status_code}: {connect_resp.text}",
                status_code=connect_resp.status_code,
            )

        connect_data: dict[str, Any] = connect_resp.json()
        ws = WebSocket()
        ws.connect(connect_data["wsUrl"])
        try:
            _ = ws.send(
                json.dumps(
                    {
                        "code": code,
                        "envs": (options.env if options else None) or {},
                        "timeout": options.timeout if options else None,
                    }
                )
            )
            return self._consume_ws(ws, handlers)
        finally:
            try:
                ws.close()
            except Exception:  # noqa: BLE001
                pass

    # -- helpers --------------------------------------------------------

    def _is_expired(self, access: SessionAccess) -> bool:
        now = datetime.now(timezone.utc)
        expiry = _parse_iso(access.token_expires_at)
        return now + timedelta(seconds=_REFRESH_SKEW_SECONDS) >= expiry

    def _evict_access(self, options: Optional[SessionRunOptions], ctx_id: str) -> None:
        if options is not None and options.context is not None:
            _ = self._ctx_access.pop(ctx_id, None)
            return
        tpl = options.template if options else None
        lang = options.language if options else None
        _ = self._transient_access.pop(f"{tpl or ''}:{lang or ''}", None)

    @staticmethod
    def _is_route_missing_404(resp: requests.Response) -> bool:
        """Heuristic: an `application/json` body with `error.name=NotFound` and a
        sessionId or expiredAt field is "missing context". A plain text 404
        or one without those fields is "route missing" (older API)."""
        try:
            payload: Any = resp.json()
        except (json.JSONDecodeError, ValueError):
            return True
        if not isinstance(payload, dict):
            return True
        body_dict: dict[str, Any] = cast("dict[str, Any]", payload)
        # Nest's unknown-route 404 wraps an object like
        # {"message":"Cannot GET ...","error":"Not Found","statusCode":404}. A
        # genuine missing-session NotFoundException carries the SAME
        # error:"Not Found" but a descriptive message ("Session <id> not
        # found."), so we must NOT key off `error` — only the message (or its
        # absence) is a reliable route-missing signal.
        msg = body_dict.get("message")
        if isinstance(msg, str):
            return msg.startswith("Cannot ") or msg.strip() == ""
        # No message field at all → not a recognizable missing-session envelope.
        return True

    def _build_run_body(self, code: str, options: Optional[SessionRunOptions]) -> dict[str, Any]:
        body: dict[str, Any] = {"code": code}
        if not options:
            return body
        if options.language is not None:
            body["language"] = options.language
        if options.template is not None:
            body["template"] = options.template
        if options.context is not None:
            body["context"] = {"id": options.context.id}
        if options.env is not None:
            body["env"] = options.env
        if options.timeout is not None:
            body["timeout"] = options.timeout
        return body

    def _build_connect_body(self, options: Optional[SessionRunOptions]) -> dict[str, Any]:
        body: dict[str, Any] = {}
        if not options:
            return body
        if options.language is not None:
            body["language"] = options.language
        if options.template is not None:
            body["template"] = options.template
        if options.context is not None:
            body["context"] = {"id": options.context.id}
        if options.timeout is not None:
            body["timeout"] = options.timeout
        return body

    def _apply_frame(
        self,
        frame: dict[str, Any],
        agg: SessionRunResult,
        handlers: Optional[dict[str, Optional[Callable[..., Any]]]],
    ) -> None:
        ftype = frame.get("type")
        if ftype == "stdout":
            text = frame.get("text") or ""
            agg.stdout += text
            self._dispatch(handlers, "on_stdout", text)
        elif ftype == "stderr":
            text = frame.get("text") or ""
            agg.stderr += text
            self._dispatch(handlers, "on_stderr", text)
        elif ftype == "error":
            err = SessionExecutionError(
                name=frame.get("name") or "Error",
                value=frame.get("value"),
                traceback=frame.get("traceback"),
            )
            agg.error = err
            self._dispatch(handlers, "on_error", err)
        elif ftype == "display":
            d = SessionDisplay(formats=frame.get("formats") or [], data=frame.get("data") or {})
            agg.displays.append(d)
            self._dispatch(handlers, "on_display", d)
        elif ftype == "control":
            self._dispatch(handlers, "on_control", frame.get("text") or "")

    @staticmethod
    def _dispatch(
        handlers: Optional[dict[str, Optional[Callable[..., Any]]]],
        name: str,
        payload: Any,
    ) -> None:
        if not handlers:
            return
        fn = handlers.get(name)
        if fn is None:
            return
        fn(payload)

    def _unmarshal_error(self, raw: Optional[dict[str, Any]]) -> Optional[SessionExecutionError]:
        if not raw:
            return None
        return SessionExecutionError(
            name=raw.get("name") or "Error",
            value=raw.get("value"),
            traceback=raw.get("traceback"),
        )

    def _unmarshal_display(self, raw: dict[str, Any]) -> SessionDisplay:
        return SessionDisplay(formats=raw.get("formats") or [], data=raw.get("data") or {})

    def _unmarshal_context(self, raw: dict[str, Any]) -> SessionRef:
        access = self._unmarshal_access(raw["access"]) if raw.get("access") else None
        return SessionRef(
            id=raw["id"],
            language=raw["language"],
            cwd=raw.get("cwd"),
            created_at=raw["createdAt"],
            last_used_at=raw.get("lastUsedAt"),
            expires_at=raw["expiresAt"],
            access=access,
        )

    @staticmethod
    def _unmarshal_access(raw: dict[str, Any]) -> SessionAccess:
        return SessionAccess(
            http_url=raw["httpUrl"],
            ws_url=raw["wsUrl"],
            token=raw.get("token") or "",
            token_expires_at=raw["tokenExpiresAt"],
        )

    def _translate_410(self, resp: requests.Response) -> Exception:
        body: dict[str, Any]
        try:
            raw: Any = resp.json()
            body = cast("dict[str, Any] | None", raw.get("error") if isinstance(raw, dict) else None) or {}
        except (json.JSONDecodeError, ValueError):
            body = {}
        name = cast("str | None", body.get("name"))
        sid = cast("str | None", body.get("sessionId"))
        if name == "SessionInvalidated" and sid and body.get("invalidatedAt"):
            return SessionInvalidatedError(sid, cast(str, body["invalidatedAt"]))
        if name == "SessionExpired" and sid and body.get("expiredAt"):
            reason_raw = cast("str | None", body.get("reason")) or "idle"
            reason: Literal["idle", "absolute"] = "absolute" if reason_raw == "absolute" else "idle"
            return SessionExpiredError(sid, cast(str, body["expiredAt"]), reason)
        return DaytonaError(f"session request returned 410: {resp.text}", status_code=410)
