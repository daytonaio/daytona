# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
from __future__ import annotations

import asyncio
import codecs
import inspect
from collections.abc import AsyncIterator, Awaitable, Callable
from typing import TYPE_CHECKING, TypeVar, cast

import aiohttp
from httpx_ws import WebSocketDisconnect
from wsproto.events import BytesMessage, CloseConnection, TextMessage

from ..common.errors import DaytonaError
from ..common.process import MAX_PREFIX_LEN, STDERR_PREFIX, STDOUT_PREFIX, OutputHandler

if TYPE_CHECKING:
    from aiohttp import ClientSession, ClientWebSocketResponse
    from httpx_ws import WebSocketSession

T = TypeVar("T")

try:
    from builtins import anext  # Python 3.10+  # pyright: ignore[reportAttributeAccessIssue, reportUnknownVariableType]
except ImportError:
    # Python 3.9 fallback
    async def anext(ait: AsyncIterator[T], default: T | None = None) -> T:
        try:
            return await ait.__anext__()  # pylint: disable=unnecessary-dunder-call
        except StopAsyncIteration:
            if default is not None:
                return default
            raise


async def process_streaming_response(
    url: str,
    headers: dict[str, str],
    on_chunk: OutputHandler[str],
    should_terminate: Callable[[], bool] | Callable[[], Awaitable[bool]],
    method: str = "GET",
    chunk_timeout: float = 2.0,
    require_consecutive_termination: bool = True,
    session: "ClientSession | None" = None,
) -> None:
    """
    Process a streaming response from a URL using aiohttp. Stream will terminate if the
    server-side stream ends or if the should_terminate function returns True.

    When *session* is provided, the request reuses the caller's pooled aiohttp.ClientSession
    so this stream shares the SDK's single TCP/TLS pool. When *session* is None (sync code
    invoking via asyncio.run, or stand-alone use) a throwaway session is created per call.

    Args:
        url: The URL to stream from.
        headers: The headers to send with the request.
        on_chunk: A callback function to process each chunk of the response.
        should_terminate: A function to check if the response should be terminated.
        method: The HTTP method to use.
        chunk_timeout: The timeout for each chunk.
        require_consecutive_termination: Whether to require two consecutive termination signals
        to terminate the stream.
        session: Optional shared aiohttp session. When None, a per-call session is created.
    """
    owned_session: ClientSession | None = None
    if session is None:
        # No shared session means we're called from sync code (asyncio.run wrapper) or a
        # caller that wants a per-call session. Use a fresh aiohttp.ClientSession.
        owned_session = aiohttp.ClientSession(trust_env=True)
        active_session = owned_session
    else:
        active_session = session

    try:
        async with active_session.request(method, url, headers=headers) as response:
            response.raise_for_status()
            stream = response.content.iter_any()
            next_chunk: asyncio.Task[bytes | None] | None = None
            timeout_task: asyncio.Task[None] | None = None
            exit_check_streak = 0
            decoder = codecs.getincrementaldecoder("utf-8")(errors="replace")

            try:
                while True:
                    if next_chunk is None:
                        next_chunk = asyncio.create_task(anext(stream, None))
                    assert next_chunk is not None

                    timeout_task = asyncio.create_task(asyncio.sleep(chunk_timeout))
                    assert timeout_task is not None

                    done, pending = await asyncio.wait({next_chunk, timeout_task}, return_when=asyncio.FIRST_COMPLETED)
                    _ = pending

                    if next_chunk in done:
                        _ = timeout_task.cancel()
                        try:
                            await timeout_task
                        except asyncio.CancelledError:
                            pass
                        finally:
                            timeout_task = None

                        try:
                            chunk = cast(bytes | None, next_chunk.result())
                        except aiohttp.ClientPayloadError as e:
                            # Match the prior httpx.RemoteProtocolError compatibility — daemon
                            # closing the stream mid-flight is an expected end-of-logs signal.
                            if "Response payload is not completed" in str(e) or "TransferEncoding" in str(e):
                                break
                            raise

                        next_chunk = None

                        if not chunk:
                            break

                        await _invoke(on_chunk, decoder.decode(chunk, final=False))
                        exit_check_streak = 0

                    elif timeout_task in done:
                        timeout_task = None
                        should_end = should_terminate()
                        if inspect.isawaitable(should_end):
                            should_end = await should_end

                        if should_end:
                            exit_check_streak += 1
                            if not require_consecutive_termination or exit_check_streak > 1:
                                break
                        else:
                            exit_check_streak = 0
            finally:
                remaining = decoder.decode(b"", final=True)
                if remaining:
                    await _invoke(on_chunk, remaining)

                if timeout_task is not None:
                    _ = timeout_task.cancel()
                    try:
                        await timeout_task
                    except asyncio.CancelledError:
                        pass
                    finally:
                        timeout_task = None
                if next_chunk is not None:
                    _ = next_chunk.cancel()
                    try:
                        await next_chunk
                    except asyncio.CancelledError:
                        pass
                    except aiohttp.ClientPayloadError as e:
                        if "Response payload is not completed" not in str(e):
                            raise
    finally:
        if owned_session is not None:
            await owned_session.close()


async def _invoke(handler: OutputHandler[str], text: str) -> None:
    """Call an output handler and await the result if it is an awaitable."""
    result = handler(text)
    if inspect.isawaitable(result):
        await result


async def std_demux_stream_aio(
    ws: "ClientWebSocketResponse",
    on_stdout: OutputHandler[str],
    on_stderr: OutputHandler[str],
) -> None:
    """
    Demultiplex an aiohttp websocket stream into stdout/stderr.

    aiohttp returns WSMessage objects with a *type* that we normalize at this single
    boundary: TEXT/BINARY → bytes; ERROR → raise; CLOSE/CLOSED/CLOSING with code in
    {None, 1000} → clean EOF; any other close code → raise (so log/PTY consumers don't
    silently swallow daemon-side failures).

    Accepts both sync and async callbacks. Async callbacks are awaited. Blocking
    operations inside sync callbacks may delay WebSocket reads — use async callbacks
    or async libraries to avoid this.

    Args:
        ws: The aiohttp ClientWebSocketResponse to demultiplex.
        on_stdout: Callback function for stdout messages (sync or async).
        on_stderr: Callback function for stderr messages (sync or async).

    Raises:
        DaytonaError: If the WebSocket emits an ERROR frame or closes with a non-1000 code.
    """

    async def recv() -> bytes | str | None:
        msg = await ws.receive()
        msg_type = msg.type
        if msg_type == aiohttp.WSMsgType.BINARY:
            return cast(bytes, msg.data)
        if msg_type == aiohttp.WSMsgType.TEXT:
            return cast(str, msg.data)
        if msg_type == aiohttp.WSMsgType.ERROR:
            exc = ws.exception()
            raise DaytonaError(f"WebSocket error: {exc}") from exc
        if msg_type == aiohttp.WSMsgType.CLOSE:
            # Server-sent close: code in msg.data, reason in msg.extra. Treat 1000/None as
            # clean EOF; surface anything else (e.g. 1011 server error) so log streams
            # don't silently swallow daemon-side failures.
            close_code = cast(int, msg.data) if msg.data is not None else None
            close_reason = msg.extra if msg.extra else None
            if close_code in (None, 1000):
                return None
            detail = close_reason or "WebSocket closed unexpectedly"
            raise DaytonaError(f"{detail} (close code {close_code})")
        if msg_type in (aiohttp.WSMsgType.CLOSED, aiohttp.WSMsgType.CLOSING):
            close_code = ws.close_code
            if close_code in (None, 1000):
                return None
            raise DaytonaError(f"WebSocket closed unexpectedly (close code {close_code})")
        return None

    await _std_demux_loop(recv, on_stdout, on_stderr)


async def std_demux_stream_httpx_ws(
    ws: "WebSocketSession",
    on_stdout: OutputHandler[str],
    on_stderr: OutputHandler[str],
) -> None:
    """
    Demultiplex a sync httpx_ws WebSocketSession from inside an ``async def`` caller.

    Used by the *sync* SDK's ``async def`` log-streaming methods, which want to share the
    sync httpx.Client connection pool but still need to be awaited (so user-supplied
    callbacks can themselves be ``async def``). The blocking ``ws.receive()`` call runs
    via :func:`asyncio.to_thread` so it doesn't stall the event loop while waiting on
    the network.

    Close-code handling matches :func:`std_demux_stream_aio`: ``None`` / ``1000`` is clean
    EOF, anything else raises ``DaytonaError`` so daemon-side failures don't get swallowed.
    """

    async def recv() -> bytes | str | None:
        while True:
            try:
                event = await asyncio.to_thread(ws.receive)
            except WebSocketDisconnect as e:
                if e.code in (None, 1000):
                    return None
                raise DaytonaError(f"{e.reason or 'WebSocket closed unexpectedly'} (close code {e.code})") from e

            if isinstance(event, TextMessage):
                return event.data
            if isinstance(event, BytesMessage):
                # wsproto's data is bytes | bytearray; coerce for the typed return.
                return bytes(event.data)
            if isinstance(event, CloseConnection):
                if event.code in (None, 1000):
                    return None
                raise DaytonaError(f"{event.reason or 'WebSocket closed unexpectedly'} (close code {event.code})")
            # Ping / Pong / fragmentary frames — wsproto handles ping auto-reply for us;
            # keep looping until we get real payload or a close.

    await _std_demux_loop(recv, on_stdout, on_stderr)


async def _std_demux_loop(
    recv: Callable[[], Awaitable[bytes | str | None]],
    on_stdout: OutputHandler[str],
    on_stderr: OutputHandler[str],
) -> None:
    """Shared stdout/stderr demultiplexer body.

    Callers normalize their transport into a single ``recv()`` coroutine that yields
    bytes / str / None. This loop owns the binary framing protocol (STDOUT_PREFIX /
    STDERR_PREFIX markers, partial-prefix safe regions, and incremental UTF-8 decoding)
    so the transport adapter (currently only aiohttp) doesn't have to duplicate it.
    """
    buf = bytearray()
    current_data_type: str | None = None

    stdout_decoder = codecs.getincrementaldecoder("utf-8")(errors="replace")
    stderr_decoder = codecs.getincrementaldecoder("utf-8")(errors="replace")

    async def emit(payload: bytes):
        if not payload:
            return
        if current_data_type == "stdout":
            text = stdout_decoder.decode(payload, final=False)
            await _invoke(on_stdout, text)
        elif current_data_type == "stderr":
            text = stderr_decoder.decode(payload, final=False)
            await _invoke(on_stderr, text)

    try:
        while True:
            chunk = await recv()
            if chunk is None:
                break

            if isinstance(chunk, str):
                chunk = chunk.encode("utf-8", "ignore")

            if not chunk:
                continue

            buf += chunk

            while True:
                safe_len = len(buf)

                if len(buf) >= MAX_PREFIX_LEN:
                    last_byte = buf[-1]
                    if last_byte not in (0x01, 0x02):
                        safe_len = len(buf)
                    elif len(buf) >= MAX_PREFIX_LEN + 1:
                        second_last_byte = buf[-2]
                        if second_last_byte not in (0x01, 0x02):
                            safe_len = len(buf) - 1
                        else:
                            safe_len = len(buf) - (MAX_PREFIX_LEN - 1)
                    else:
                        safe_len = len(buf) - (MAX_PREFIX_LEN - 1)
                else:
                    safe_len = len(buf) - (MAX_PREFIX_LEN - 1)

                if safe_len <= 0:
                    break

                si = buf.find(STDOUT_PREFIX, 0, safe_len)
                ei = buf.find(STDERR_PREFIX, 0, safe_len)

                next_idx = -1
                next_kind: str | None = None
                next_len = 0

                if si != -1 and (ei == -1 or si < ei):
                    next_idx, next_kind, next_len = si, "stdout", len(STDOUT_PREFIX)
                elif ei != -1:
                    next_idx, next_kind, next_len = ei, "stderr", len(STDERR_PREFIX)

                if next_idx == -1:
                    to_emit = bytes(buf[:safe_len])
                    await emit(to_emit)
                    del buf[:safe_len]
                    break

                if next_idx > 0:
                    to_emit = bytes(buf[:next_idx])
                    await emit(to_emit)

                del buf[: next_idx + next_len]
                current_data_type = next_kind

    finally:
        if buf and current_data_type in ("stdout", "stderr"):
            if current_data_type == "stdout":
                text = stdout_decoder.decode(bytes(buf), final=True)
                await _invoke(on_stdout, text)
            else:
                text = stderr_decoder.decode(bytes(buf), final=True)
                await _invoke(on_stderr, text)
        else:
            stdout_flushed = stdout_decoder.decode(b"", final=True)
            stderr_flushed = stderr_decoder.decode(b"", final=True)
            if stdout_flushed:
                await _invoke(on_stdout, stdout_flushed)
            if stderr_flushed:
                await _invoke(on_stderr, stderr_flushed)
