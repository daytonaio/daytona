# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import inspect
import json
from collections.abc import Awaitable, Callable
from typing import Any, cast

import aiohttp

from daytona_toolbox_api_client_async import PtySessionInfo

from ..common.errors import DaytonaConnectionError, DaytonaError, DaytonaTimeoutError
from ..common.pty import PtyResult, PtySize


class AsyncPtyHandle:
    """
    PTY session handle for managing a single PTY session asynchronously.

    Provides methods for sending input, waiting for connection and exit, and managing the WebSocket connection.

    Example usage:
    ```python
    # Create a PTY session with callback for handling output
    async def handle_output(data: bytes):
        print(data.decode('utf-8'), end='')

    # Connect to a PTY session
    pty_handle = await process.connect_pty_session('my-session', handle_output)

    # Wait for connection to be established
    await pty_handle.wait_for_connection()

    # Send commands
    await pty_handle.send_input('ls -la\n')
    await pty_handle.send_input('echo "Hello World"\n')

    # Resize the terminal
    pty_session_info = await pty_handle.resize(PtySize(cols=120, rows=30))
    print(f"PTY session resized to {pty_session_info.cols}x{pty_session_info.rows}")

    # Wait for process to exit or kill it
    try:
        result = await pty_handle.wait()
        print(f"PTY exited with code: {result.exit_code}")
    except Exception:
        # Kill the PTY if needed
        await pty_handle.kill()
    finally:
        # Always disconnect when done
        await pty_handle.disconnect()
    ```
    """

    def __init__(
        self,
        ws: aiohttp.ClientWebSocketResponse,
        on_data: Callable[[bytes], None] | Callable[[bytes], Awaitable[None]] | None = None,
        session_id: str | None = None,
        handle_resize: Callable[[PtySize], Awaitable[PtySessionInfo]] | None = None,
        handle_kill: Callable[[], Awaitable[None]] | None = None,
    ):
        self._ws: aiohttp.ClientWebSocketResponse = ws
        self._on_data: Callable[[bytes], None] | Callable[[bytes], Awaitable[None]] | None = on_data
        self._session_id: str | None = session_id
        self._handle_resize: Callable[[PtySize], Awaitable[PtySessionInfo]] | None = handle_resize
        self._handle_kill: Callable[[], Awaitable[None]] | None = handle_kill
        self._exit_code: int | None = None
        self._error: str | None = None
        self._connected: bool = False
        self._connection_established: bool = False

        # Start handling WebSocket events
        self._wait: asyncio.Task[None] = asyncio.create_task(self._handle_websocket())

    @property
    def session_id(self) -> str | None:
        """Session ID of the PTY session"""
        return self._session_id

    @property
    def exit_code(self) -> int | None:
        """Exit code of the PTY process (if terminated)"""
        return self._exit_code

    @property
    def error(self) -> str | None:
        """Error message if the PTY failed"""
        return self._error

    def is_connected(self) -> bool:
        """Check if connected to the PTY session"""
        return self._connected and not self._ws.closed

    async def wait_for_connection(self) -> None:
        """
        Wait for the WebSocket connection to be established.

        Raises:
            TimeoutError: If connection timeout (10 seconds) is reached
            ConnectionError: If connection fails or WebSocket is closed
        """
        if self._connection_established:
            return

        timeout = 10.0  # 10 seconds
        start_time = asyncio.get_event_loop().time()

        while not self._connection_established:
            if asyncio.get_event_loop().time() - start_time > timeout:
                raise DaytonaTimeoutError("PTY connection timeout")

            if self._ws.closed or self._error:
                raise DaytonaConnectionError(self._error or "Connection failed")

            await asyncio.sleep(0.1)

    async def send_input(self, data: str | bytes) -> None:
        """
        Send input data to the PTY.

        Args:
            data: Input data to send (string or bytes)

        Raises:
            ConnectionError: If PTY is not connected
            RuntimeError: If sending input fails
        """
        if not self.is_connected():
            raise DaytonaConnectionError("PTY is not connected")

        try:
            if isinstance(data, str):
                await self._ws.send_bytes(data.encode("utf-8"))
            else:
                await self._ws.send_bytes(data)
        except Exception as e:
            raise DaytonaConnectionError(f"Failed to send input to PTY: {e}") from e

    async def wait(self) -> PtyResult:
        """
        Wait for the PTY process to exit and return the result.

        Returns:
            PtyResult: Result containing exit code and error (if any)
        """
        await self._wait

        return PtyResult(
            exit_code=self._exit_code,
            error=self._error,
        )

    async def resize(self, pty_size: PtySize) -> PtySessionInfo:
        """
        Resize the PTY terminal.

        Args:
            pty_size: PtySize object containing the new terminal dimensions

        Raises:
            RuntimeError: If resize handler is not available or resize fails
        """
        if not self._handle_resize:
            raise DaytonaError("Resize handler not available")

        try:
            return await self._handle_resize(pty_size)
        except Exception as e:
            raise DaytonaError(f"Failed to resize PTY: {e}") from e

    async def kill(self) -> None:
        """
        Kill the PTY process.

        Raises:
            RuntimeError: If kill handler is not available or kill fails
        """
        if not self._handle_kill:
            raise DaytonaError("Kill handler not available")

        try:
            await self._handle_kill()
        except Exception as e:
            raise DaytonaError(f"Failed to kill PTY: {e}") from e

    async def disconnect(self) -> None:
        """Disconnect from the PTY session"""
        if self._wait and not self._wait.done():
            _ = self._wait.cancel()
            try:
                await self._wait
            except asyncio.CancelledError:
                pass

        if not self._ws.closed:
            try:
                _ = await self._ws.close()
            except Exception:
                pass

    async def _handle_websocket(self) -> None:
        """Handle WebSocket messages and connection lifecycle."""
        # We use ws.receive() rather than `async for` because aiohttp's iterator silently
        # stops on CLOSE without exposing the close code/reason — but the daytona daemon
        # encodes the PTY exit code in the close frame's reason payload, so we must
        # inspect every WSMessage including CLOSE/CLOSED to extract the exit data.
        try:
            self._connected = True

            close_code: int | None = None
            close_reason: str | None = None
            close_frame_seen = False

            while True:
                msg = await self._ws.receive()

                if msg.type == aiohttp.WSMsgType.TEXT:
                    await self._handle_message(cast(str, msg.data))
                    continue
                if msg.type == aiohttp.WSMsgType.BINARY:
                    await self._handle_message(cast(bytes, msg.data))
                    continue
                if msg.type == aiohttp.WSMsgType.CLOSE:
                    # Server-initiated close: code is in msg.data, reason is in msg.extra.
                    close_code = cast(int, msg.data) if msg.data is not None else None
                    close_reason = msg.extra if msg.extra else None
                    close_frame_seen = True
                    break
                if msg.type in (aiohttp.WSMsgType.CLOSED, aiohttp.WSMsgType.CLOSING):
                    break
                if msg.type == aiohttp.WSMsgType.ERROR:
                    exc = self._ws.exception()
                    self._error = f"WebSocket error: {exc}" if exc else "WebSocket error"
                    break

            # Fall back to ws.close_code when we exited via CLOSED rather than CLOSE.
            if not close_frame_seen and self._ws.close_code is not None:
                close_code = self._ws.close_code

            await self._handle_close(close_code, close_reason)

        except Exception as e:
            self._error = f"Unexpected error: {e}"
        finally:
            self._connected = False

    async def _handle_message(self, message: str | bytes) -> None:
        """Handle individual WebSocket messages"""
        try:
            if isinstance(message, str):
                # Try to parse as control message first
                try:
                    control_msg: dict[str, object] = json.loads(message)
                    if control_msg.get("type") == "control":
                        await self._handle_control_message(control_msg)
                        return
                except (json.JSONDecodeError, ValueError):
                    # Not a control message, treat as PTY output
                    pass

                # Regular PTY text output
                if self._on_data:
                    data = message.encode("utf-8")
                    await self._call_data_handler(data)
            else:
                # Binary PTY data
                if self._on_data:
                    await self._call_data_handler(message)

        except Exception as e:
            raise DaytonaError(f"Error handling PTY message: {e}") from e

    async def _handle_control_message(self, control_msg: dict[str, object]) -> None:
        """Handle control messages from the PTY server"""
        status = control_msg.get("status")

        if status == "connected":
            self._connection_established = True
        elif status == "error":
            self._error = cast(str, control_msg.get("error", "Unknown connection error"))
            self._connected = False

    async def _handle_close(self, close_code: int | None, close_reason: str | None) -> None:
        """Handle WebSocket close event."""
        self._connected = False

        # Parse structured exit data from close reason (daemon encodes exit_code/error in
        # the close frame so the client doesn't need a separate roundtrip to learn them).
        if close_reason:
            try:
                exit_data = json.loads(close_reason)
            except (json.JSONDecodeError, ValueError):
                if close_code == 1000:
                    self._exit_code = 0
            else:
                if isinstance(exit_data, dict):
                    exit_data = cast(dict[str, Any], exit_data)
                    exit_code_value: int | None = exit_data.get("exitCode")
                    if isinstance(exit_code_value, int):
                        self._exit_code = exit_code_value
                        exit_reason_value = exit_data.get("exitReason")
                        if isinstance(exit_reason_value, str):
                            self._error = exit_reason_value

                    error_value = exit_data.get("error")
                    if isinstance(error_value, str):
                        self._error = error_value

        # Default to exit code 0 if we can't parse it and it was a normal close
        if self._exit_code is None and close_code == 1000:
            self._exit_code = 0

    async def _call_data_handler(self, data: bytes) -> None:
        """Call the data handler, supporting both sync and async callbacks"""
        if self._on_data:
            result = self._on_data(data)
            if inspect.isawaitable(result):
                await result
