# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Asynchronous SessionService — the user-facing surface for the Sessions product.

The hot path (`run` / `run_stream`) skips the Daytona API and talks directly
to the in-sandbox session-daemon via the proxy chain — same shape
`sandbox.process.code_run` uses against the classic daytona-daemon. The API
is only hit on `createSession` / first one-shot / token refresh (every
~5 min) / on `SessionInvalidatedError`. See the SDK plan for the full
design.

The legacy `POST /sessions/code-run` and `POST /sessions/connect` paths are
preserved as transparent fallbacks for older API servers that don't yet
expose `/access` / `/transients`.
"""

from __future__ import annotations

import asyncio
import json
import time
from collections.abc import Awaitable
from datetime import datetime, timedelta, timezone
from typing import Any, Callable, Literal, Optional, Union, cast

import aiohttp

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

# Per-frame handlers may be sync OR async — we just await the result if it's awaitable.
_FrameHandler = Callable[[Any], Union[None, Awaitable[None]]]

_REFRESH_SKEW_SECONDS = 60


def _parse_iso(s: str) -> datetime:
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


def _load_websockets(error_prefix: str = "direct session execution") -> tuple[Any, Any, Any]:
    """Lazy-load the optional `websockets` dep and return (connect, InvalidStatusCode, WebSocketException).

    Returned as `Any` because the package surface drifts across versions
    (`websockets.client.connect` vs `websockets.connect`, `InvalidStatusCode`
    vs `InvalidStatus`), and the SDK only relies on a few duck-typed methods.
    """
    try:
        # pylint: disable-next=import-outside-toplevel
        websockets_client = __import__("websockets.client", fromlist=["connect"])
        # pylint: disable-next=import-outside-toplevel
        websockets_exc = __import__("websockets.exceptions", fromlist=["InvalidStatusCode", "WebSocketException"])
    except ImportError as exc:
        raise DaytonaError(
            f"{error_prefix} requires the 'websockets' package; install with `pip install websockets`"
        ) from exc
    ws_connect: Any = websockets_client.connect
    invalid_status: Any = websockets_exc.InvalidStatusCode
    ws_exception: Any = websockets_exc.WebSocketException
    return ws_connect, invalid_status, ws_exception


class _LegacyFallback(Exception):
    """Signal: API endpoint not available on this server; degrade to /code-run."""


class _WSAuthError(Exception):
    """Signal: WS handshake returned 401/403; refresh access bundle and retry once."""


class AsyncSessionService:
    """Async API surface for the Sessions product.

    The shape mirrors `SessionService` (sync) so user code switching from sync
    to async reads almost identically.
    """

    _base_url: str
    _headers: dict[str, str]

    def __init__(self, base_url: str, headers: dict[str, str]):
        self._base_url = base_url.rstrip("/")
        self._headers = dict(headers)
        _ = self._headers.setdefault("Content-Type", "application/json")
        self._ctx_access: dict[str, SessionAccess] = {}
        self._transient_access: dict[str, tuple[str, SessionAccess]] = {}
        self._transient_supported: Optional[bool] = None

    # -- one-shot run ----------------------------------------------------

    async def run(self, code: str, options: Optional[SessionRunOptions] = None) -> SessionRunResult:
        return await self._run_internal(code, options, handlers=None)

    # -- streaming run ---------------------------------------------------

    async def run_stream(
        self,
        code: str,
        options: Optional[SessionRunOptions] = None,
        on_stdout: Optional[_FrameHandler] = None,
        on_stderr: Optional[_FrameHandler] = None,
        on_error: Optional[_FrameHandler] = None,
        on_display: Optional[_FrameHandler] = None,
        on_control: Optional[_FrameHandler] = None,
    ) -> SessionRunResult:
        return await self._run_internal(
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

    async def create_session(
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
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.post(
                f"{self._base_url}/sessions",
                data=json.dumps(body),
                timeout=aiohttp.ClientTimeout(total=120),
            ) as resp:
                if resp.status >= 400:
                    text = await resp.text()
                    raise DaytonaError(
                        f"create_session returned {resp.status}: {text}",
                        status_code=resp.status,
                    )
                ctx = self._unmarshal_context(await resp.json())
                if ctx.access is not None:
                    self._ctx_access[ctx.id] = ctx.access
                return ctx

    async def list_sessions(self, template: Optional[str] = None) -> list[SessionRef]:
        params = {"template": template} if template else None
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.get(
                f"{self._base_url}/sessions",
                params=params,
                timeout=aiohttp.ClientTimeout(total=30),
            ) as resp:
                if resp.status >= 400:
                    raise DaytonaError(f"list_sessions returned {resp.status}", status_code=resp.status)
                return [self._unmarshal_context(c) for c in await resp.json()]

    async def delete_session(self, session_id: str) -> None:
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.delete(
                f"{self._base_url}/sessions/{session_id}",
                timeout=aiohttp.ClientTimeout(total=30),
            ) as resp:
                if resp.status not in (204, 404):
                    raise DaytonaError(f"delete_session returned {resp.status}", status_code=resp.status)
        _ = self._ctx_access.pop(session_id, None)

    async def list_templates(self) -> list[dict[str, Any]]:
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.get(
                f"{self._base_url}/sessions/templates",
                timeout=aiohttp.ClientTimeout(total=30),
            ) as resp:
                resp.raise_for_status()
                return await resp.json()

    async def list_packages(self, template_name: str, language: str) -> list[dict[str, Any]]:
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.get(
                f"{self._base_url}/sessions/templates/{template_name}/packages",
                params={"language": language},
                timeout=aiohttp.ClientTimeout(total=30),
            ) as resp:
                resp.raise_for_status()
                return await resp.json()

    # -- direct-to-sandbox hot path -------------------------------------

    async def _run_internal(
        self,
        code: str,
        options: Optional[SessionRunOptions],
        handlers: Optional[dict[str, Optional[_FrameHandler]]],
    ) -> SessionRunResult:
        if (options is None or options.context is None) and self._transient_supported is False:
            return await self._run_legacy(code, options, handlers)

        try:
            ctx_id, access, reset = await self._ensure_access(options)
        except _LegacyFallback:
            if options is None or options.context is None:
                self._transient_supported = False
            return await self._run_legacy(code, options, handlers)

        try:
            return await self._run_ws_direct(ctx_id, access, code, options, reset, handlers)
        except _WSAuthError:
            self._evict_access(options, ctx_id)
            try:
                ctx_id, access, reset = await self._ensure_access(options)
            except _LegacyFallback:
                if options is None or options.context is None:
                    self._transient_supported = False
                return await self._run_legacy(code, options, handlers)
            return await self._run_ws_direct(ctx_id, access, code, options, reset, handlers)

    async def _ensure_access(self, options: Optional[SessionRunOptions]) -> tuple[str, SessionAccess, bool]:
        if options is not None and options.context is not None:
            ctx_id = options.context.id
            if options.context.access is not None and ctx_id not in self._ctx_access:
                self._ctx_access[ctx_id] = options.context.access

            access = self._ctx_access.get(ctx_id)
            if access is not None and not self._is_expired(access):
                return ctx_id, access, False

            access = await self._fetch_session_access(ctx_id)
            self._ctx_access[ctx_id] = access
            return ctx_id, access, False

        tpl = options.template if options else None
        lang = options.language if options else None
        key = f"{tpl or ''}:{lang or ''}"
        cached = self._transient_access.get(key)
        if cached is not None and not self._is_expired(cached[1]):
            return cached[0], cached[1], True

        ctx_id, access = await self._fetch_transient(tpl, lang)
        self._transient_access[key] = (ctx_id, access)
        self._transient_supported = True
        return ctx_id, access, True

    async def _run_ws_direct(
        self,
        ctx_id: str,
        access: SessionAccess,
        code: str,
        options: Optional[SessionRunOptions],
        reset: bool,
        handlers: Optional[dict[str, Optional[_FrameHandler]]],
    ) -> SessionRunResult:
        ws_connect, invalid_status_code, websocket_exception = _load_websockets()

        ws_cm: Any = ws_connect(access.ws_url)
        try:
            # Acquire the WS via the context manager's __aenter__ directly so we can
            # surface InvalidStatusCode for the 401/403 -> _WSAuthError fallback
            # without an `async with` block that would have already swallowed it.
            ws: Any = await ws_cm.__aenter__()  # pylint: disable=unnecessary-dunder-call
        except invalid_status_code as exc:
            status: Any = getattr(exc, "status_code", None)
            if status in (401, 403):
                raise _WSAuthError() from exc
            # 404 (context gone), 400 (proxy: "Is the Sandbox started?"),
            # 5xx — all map to invalidation. The caller's recovery action is
            # the same: drop the context and create a fresh one.
            raise SessionInvalidatedError(ctx_id, _now_iso()) from exc
        except (ConnectionRefusedError, websocket_exception, OSError) as exc:
            raise SessionInvalidatedError(ctx_id, _now_iso()) from exc

        try:
            await ws.send(
                json.dumps(
                    {
                        "code": code,
                        "envs": (options.env if options else None) or {},
                        "timeout": (options.timeout if options else None) or 0,
                        "reset": reset,
                    }
                )
            )
            return await self._consume_ws(ws, handlers)
        finally:
            try:
                await ws_cm.__aexit__(None, None, None)
            except Exception:  # noqa: BLE001
                pass

    async def _consume_ws(
        self,
        ws: Any,
        handlers: Optional[dict[str, Optional[_FrameHandler]]],
    ) -> SessionRunResult:
        aggregated = SessionRunResult(stdout="", stderr="", error=None, displays=[], duration_ms=0)
        started_at = time.time()
        # The daemon emits a terminal control frame ("completed"/"interrupted")
        # to signal a clean end-of-run. A stream error / end BEFORE that frame
        # is an abnormal disconnect and must surface as an error, not a silent
        # success.
        completed = False
        disconnect_exc: Optional[BaseException] = None
        try:
            async for raw in ws:
                if isinstance(raw, bytes):
                    raw = raw.decode("utf-8", errors="replace")
                if not raw:
                    continue
                try:
                    frame = json.loads(raw)
                except json.JSONDecodeError:
                    continue
                if frame.get("type") == "control" and frame.get("text") in ("completed", "interrupted"):
                    completed = True
                await self._apply_frame(frame, aggregated, handlers)
            # Stream ended. Clean only if we already saw the terminal frame.
        except Exception as exc:  # noqa: BLE001
            disconnect_exc = exc
        return _finalize_disconnect(aggregated, completed, started_at, disconnect_exc)

    async def _fetch_session_access(self, session_id: str) -> SessionAccess:
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.get(
                f"{self._base_url}/sessions/{session_id}/access",
                timeout=aiohttp.ClientTimeout(total=30),
            ) as resp:
                if resp.status == 404:
                    if await self._is_route_missing_404(resp):
                        raise _LegacyFallback()
                    raise SessionInvalidatedError(session_id, _now_iso())
                if resp.status == 410:
                    raise await self._translate_410(resp)
                if resp.status >= 400:
                    text = await resp.text()
                    raise DaytonaError(
                        f"session access returned {resp.status}: {text}",
                        status_code=resp.status,
                    )
                return self._unmarshal_access(await resp.json())

    async def _fetch_transient(
        self,
        template: Optional[str],
        language: Optional[str],
    ) -> tuple[str, SessionAccess]:
        body: dict[str, Any] = {}
        if template is not None:
            body["template"] = template
        if language is not None:
            body["language"] = language
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.post(
                f"{self._base_url}/sessions/transients",
                data=json.dumps(body),
                timeout=aiohttp.ClientTimeout(total=120),
            ) as resp:
                if resp.status == 404 and await self._is_route_missing_404(resp):
                    raise _LegacyFallback()
                if resp.status >= 400:
                    text = await resp.text()
                    raise DaytonaError(
                        f"session transient returned {resp.status}: {text}",
                        status_code=resp.status,
                    )
                ctx = self._unmarshal_context(await resp.json())
                if ctx.access is None:
                    raise _LegacyFallback()
                return ctx.id, ctx.access

    async def _run_legacy(
        self,
        code: str,
        options: Optional[SessionRunOptions],
        handlers: Optional[dict[str, Optional[_FrameHandler]]],
    ) -> SessionRunResult:
        if handlers is not None:
            return await self._run_stream_legacy(code, options, handlers)
        return await self._run_code_run_legacy(code, options)

    async def _run_code_run_legacy(
        self,
        code: str,
        options: Optional[SessionRunOptions],
    ) -> SessionRunResult:
        body = self._build_run_body(code, options)
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.post(
                f"{self._base_url}/sessions/code-run",
                data=json.dumps(body),
                timeout=aiohttp.ClientTimeout(total=600),
            ) as resp:
                if resp.status == 410:
                    raise await self._translate_410(resp)
                if resp.status >= 400:
                    text = await resp.text()
                    raise DaytonaError(
                        f"session run returned {resp.status}: {text}",
                        status_code=resp.status,
                    )
                data: dict[str, Any] = await resp.json()
        displays_raw: list[dict[str, Any]] = data.get("displays") or []
        return SessionRunResult(
            stdout=data.get("stdout", ""),
            stderr=data.get("stderr", ""),
            error=self._unmarshal_error(data.get("error")),
            displays=[self._unmarshal_display(d) for d in displays_raw],
            duration_ms=int(data.get("durationMs", 0)),
        )

    async def _run_stream_legacy(
        self,
        code: str,
        options: Optional[SessionRunOptions],
        handlers: dict[str, Optional[_FrameHandler]],
    ) -> SessionRunResult:
        ws_connect, _invalid_status_code, _websocket_exception = _load_websockets(error_prefix="run_stream")

        connect_body = self._build_connect_body(options)
        async with aiohttp.ClientSession(headers=self._headers) as session:
            async with session.post(
                f"{self._base_url}/sessions/connect",
                data=json.dumps(connect_body),
                timeout=aiohttp.ClientTimeout(total=60),
            ) as resp:
                if resp.status == 410:
                    raise await self._translate_410(resp)
                if resp.status >= 400:
                    text = await resp.text()
                    raise DaytonaError(
                        f"session connect returned {resp.status}: {text}",
                        status_code=resp.status,
                    )
                connect_data: dict[str, Any] = await resp.json()

        aggregated = SessionRunResult(stdout="", stderr="", error=None, displays=[], duration_ms=0)
        started_at = time.time()
        async with ws_connect(connect_data["wsUrl"]) as ws_raw:
            ws: Any = ws_raw
            await ws.send(
                json.dumps(
                    {
                        "code": code,
                        "envs": (options.env if options else None) or {},
                        "timeout": options.timeout if options else None,
                    }
                )
            )
            completed = False
            disconnect_exc: Optional[BaseException] = None
            try:
                async for frame in ws:
                    raw: str = frame.decode("utf-8", errors="replace") if isinstance(frame, bytes) else cast(str, frame)
                    if not raw:
                        continue
                    try:
                        frame = json.loads(raw)
                    except json.JSONDecodeError:
                        continue
                    if frame.get("type") == "control" and frame.get("text") in ("completed", "interrupted"):
                        completed = True
                    await self._apply_frame(frame, aggregated, handlers)
            except Exception as exc:  # noqa: BLE001
                disconnect_exc = exc
        return _finalize_disconnect(aggregated, completed, started_at, disconnect_exc)

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
    async def _is_route_missing_404(resp: aiohttp.ClientResponse) -> bool:
        try:
            payload: Any = await resp.json()
        except (json.JSONDecodeError, ValueError, aiohttp.ContentTypeError):
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

    async def _apply_frame(
        self,
        frame: dict[str, Any],
        agg: SessionRunResult,
        handlers: Optional[dict[str, Optional[_FrameHandler]]],
    ) -> None:
        ftype = frame.get("type")
        if ftype == "stdout":
            text = frame.get("text") or ""
            agg.stdout += text
            await self._dispatch(handlers, "on_stdout", text)
        elif ftype == "stderr":
            text = frame.get("text") or ""
            agg.stderr += text
            await self._dispatch(handlers, "on_stderr", text)
        elif ftype == "error":
            err = SessionExecutionError(
                name=frame.get("name") or "Error",
                value=frame.get("value"),
                traceback=frame.get("traceback"),
            )
            agg.error = err
            await self._dispatch(handlers, "on_error", err)
        elif ftype == "display":
            d = SessionDisplay(formats=frame.get("formats") or [], data=frame.get("data") or {})
            agg.displays.append(d)
            await self._dispatch(handlers, "on_display", d)
        elif ftype == "control":
            await self._dispatch(handlers, "on_control", frame.get("text") or "")

    @staticmethod
    async def _dispatch(
        handlers: Optional[dict[str, Optional[_FrameHandler]]],
        name: str,
        payload: Any,
    ) -> None:
        if not handlers:
            return
        fn = handlers.get(name)
        if fn is None:
            return
        result = fn(payload)
        if asyncio.iscoroutine(result) or hasattr(result, "__await__"):
            await cast(Any, result)

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

    async def _translate_410(self, resp: aiohttp.ClientResponse) -> Exception:
        body: dict[str, Any]
        try:
            raw: Any = await resp.json()
            body = cast("dict[str, Any] | None", raw.get("error") if isinstance(raw, dict) else None) or {}
        except (json.JSONDecodeError, ValueError, aiohttp.ContentTypeError):
            body = {}
        name = cast("str | None", body.get("name"))
        sid = cast("str | None", body.get("sessionId"))
        if name == "SessionInvalidated" and sid and body.get("invalidatedAt"):
            return SessionInvalidatedError(sid, cast(str, body["invalidatedAt"]))
        if name == "SessionExpired" and sid and body.get("expiredAt"):
            reason_raw = cast("str | None", body.get("reason")) or "idle"
            reason: Literal["idle", "absolute"] = "absolute" if reason_raw == "absolute" else "idle"
            return SessionExpiredError(sid, cast(str, body["expiredAt"]), reason)
        text = await resp.text()
        return DaytonaError(f"session request returned 410: {text}", status_code=410)
