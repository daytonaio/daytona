# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import inspect
from typing import Callable

import httpx


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

                        on_chunk(chunk.decode("utf-8", "ignore"))
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
