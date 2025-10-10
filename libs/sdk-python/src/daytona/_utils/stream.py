# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import codecs
import inspect
from typing import Callable

import httpx
import websockets
from websockets.asyncio.client import Connection

from ..common.process import MAX_PREFIX_LEN, STDERR_PREFIX, STDOUT_PREFIX
from .errors import DaytonaError


async def process_streaming_response(
    url: str,
    headers: dict,
    on_chunk: Callable[[str], None],
    should_terminate: Callable[[], bool],
    method: str = "GET",
    chunk_timeout: float = 2.0,
    require_consecutive_termination: bool = True,
) -> None:
    """
    Process a streaming response from a URL. Stream will terminate if the server-side stream
    ends or if the should_terminate function returns True.

    Args:
        url: The URL to stream from.
        headers: The headers to send with the request.
        on_chunk: A callback function to process each chunk of the response.
        should_terminate: A function to check if the response should be terminated.
        method: The HTTP method to use.
        chunk_timeout: The timeout for each chunk.
        require_consecutive_termination: Whether to require two consecutive termination signals
        to terminate the stream.
    """
    async with httpx.AsyncClient(timeout=None) as client:
        async with client.stream(method, url, headers=headers) as response:
            stream = response.aiter_bytes()
            next_chunk = None
            exit_check_streak = 0
            # Use incremental decoder to properly handle UTF-8 sequences split across chunks
            decoder = codecs.getincrementaldecoder("utf-8")(errors="replace")

            try:
                while True:
                    if next_chunk is None:
                        next_chunk = asyncio.create_task(anext(stream, None))
                    timeout_task = asyncio.create_task(asyncio.sleep(chunk_timeout))

                    done, _ = await asyncio.wait([next_chunk, timeout_task], return_when=asyncio.FIRST_COMPLETED)

                    if next_chunk in done:
                        # Cancel timeout task and handle any cancellation errors
                        timeout_task.cancel()
                        try:
                            await timeout_task
                        except asyncio.CancelledError:
                            pass

                        try:
                            chunk = next_chunk.result()
                        except httpx.RemoteProtocolError as e:
                            if "peer closed connection without sending complete message body" in str(e):
                                break
                            raise e

                        next_chunk = None

                        if chunk is None:
                            break

                        # Use final=False to buffer incomplete UTF-8 sequences for the next chunk
                        on_chunk(decoder.decode(chunk, final=False))
                        exit_check_streak = 0  # Reset on activity

                    elif timeout_task in done:
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
                # Flush any remaining buffered bytes from the decoder
                remaining = decoder.decode(b"", final=True)
                if remaining:
                    on_chunk(remaining)

                # Final cleanup - ensure any remaining tasks are cancelled
                if timeout_task:
                    timeout_task.cancel()
                    try:
                        await timeout_task
                    except asyncio.CancelledError:
                        pass
                if next_chunk:
                    next_chunk.cancel()
                    try:
                        await next_chunk
                    except asyncio.CancelledError:
                        pass
                    except httpx.RemoteProtocolError as e:
                        if "peer closed connection without sending complete message body" not in str(e):
                            raise e


async def std_demux_stream(
    connection: Connection,
    on_stdout: Callable[[str], None],
    on_stderr: Callable[[str], None],
) -> None:
    """
    Demultiplex a WebSocket stream into separate stdout and stderr streams.

    Args:
        connection: The WebSocket connection to demultiplex.
        on_stdout: Callback function for stdout messages.
        on_stderr: Callback function for stderr messages.

    Raises:
        DaytonaError: If the WebSocket connection closed error occurs.
    """
    buf = bytearray()
    current_data_type = None  # None | "stdout" | "stderr"

    # Separate incremental decoders for stdout and stderr to maintain independent UTF-8 decoding state
    stdout_decoder = codecs.getincrementaldecoder("utf-8")(errors="replace")
    stderr_decoder = codecs.getincrementaldecoder("utf-8")(errors="replace")

    def emit(payload: bytes):
        if not payload:
            return
        # Use final=False to buffer incomplete UTF-8 sequences for the next chunk
        if current_data_type == "stdout":
            text = stdout_decoder.decode(payload, final=False)
            on_stdout(text)
        elif current_data_type == "stderr":
            text = stderr_decoder.decode(payload, final=False)
            on_stderr(text)
        # If current is None, drop unlabeled bytes (shouldn't happen with proper labeling)

    try:
        while True:
            try:
                chunk = await connection.recv()
            except websockets.exceptions.ConnectionClosedOK:
                break
            except websockets.exceptions.ConnectionClosedError as e:
                raise DaytonaError(f"WebSocket error: {e}") from e

            # WS server sends text frames; convert to bytes so we can match control markers.
            if isinstance(chunk, str):
                chunk = chunk.encode("utf-8", "ignore")

            if not chunk:
                continue

            buf += chunk

            # Process as much as we can, preserving only bytes that could be part of a prefix
            while True:
                # Calculate how many bytes we can safely process
                # We need to keep bytes that could potentially be the start of a prefix marker
                safe_len = len(buf)

                # Check if the last few bytes could be part of a prefix marker
                if len(buf) >= MAX_PREFIX_LEN:
                    # Check if the last byte could be part of a prefix (must be \x01 or \x02)
                    last_byte = buf[-1]
                    if last_byte not in (0x01, 0x02):
                        # Last byte can't be part of any prefix, safe to process everything
                        safe_len = len(buf)
                    elif len(buf) >= MAX_PREFIX_LEN + 1:
                        # Check second-to-last byte if buffer is long enough
                        second_last_byte = buf[-2]
                        if second_last_byte not in (0x01, 0x02):
                            # Second-to-last byte can't be part of any prefix, safe to process all but last byte
                            safe_len = len(buf) - 1
                        else:
                            # Both last bytes could be part of prefix, keep MAX_PREFIX_LEN - 1 bytes
                            safe_len = len(buf) - (MAX_PREFIX_LEN - 1)
                    else:
                        # Buffer is exactly MAX_PREFIX_LEN, keep MAX_PREFIX_LEN - 1 bytes
                        safe_len = len(buf) - (MAX_PREFIX_LEN - 1)
                else:
                    # Buffer shorter than MAX_PREFIX_LEN, keep MAX_PREFIX_LEN - 1 bytes
                    safe_len = len(buf) - (MAX_PREFIX_LEN - 1)

                if safe_len <= 0:
                    break

                # Find earliest next marker within the safe region
                si = buf.find(STDOUT_PREFIX, 0, safe_len)
                ei = buf.find(STDERR_PREFIX, 0, safe_len)

                next_idx = -1
                next_kind = None
                next_len = 0

                if si != -1 and (ei == -1 or si < ei):
                    next_idx, next_kind, next_len = si, "stdout", len(STDOUT_PREFIX)
                elif ei != -1:
                    next_idx, next_kind, next_len = ei, "stderr", len(STDERR_PREFIX)

                if next_idx == -1:
                    # No full marker in safe region: emit everything we safely can as payload
                    to_emit = bytes(buf[:safe_len])
                    emit(to_emit)
                    del buf[:safe_len]
                    break  # wait for more data to resolve any partial marker at the end

                # We found a marker. Emit preceding bytes (if any) under the current stream.
                if next_idx > 0:
                    to_emit = bytes(buf[:next_idx])
                    emit(to_emit)

                # Advance past the marker and switch current stream
                del buf[: next_idx + next_len]
                current_data_type = next_kind

    finally:
        # Flush any remaining buffered payload on clean close
        if buf and current_data_type in ("stdout", "stderr"):
            if current_data_type == "stdout":
                # Use final=True to flush any buffered incomplete UTF-8 sequences
                text = stdout_decoder.decode(bytes(buf), final=True)
                on_stdout(text)
            else:
                text = stderr_decoder.decode(bytes(buf), final=True)
                on_stderr(text)
        else:
            # Flush any remaining bytes in the decoders even if buf is empty
            stdout_flushed = stdout_decoder.decode(b"", final=True)
            stderr_flushed = stderr_decoder.decode(b"", final=True)
            if stdout_flushed:
                on_stdout(stdout_flushed)
            if stderr_flushed:
                on_stderr(stderr_flushed)
