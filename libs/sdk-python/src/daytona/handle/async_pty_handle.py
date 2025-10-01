# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import inspect
import json
from typing import Awaitable, Callable, Optional, Union

import websockets
from daytona_api_client_async import PtySessionInfo
from websockets.asyncio.client import Connection

from .._utils.errors import DaytonaError
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
        ws: Connection,
        on_data: Optional[Union[Callable[[bytes], None], Callable[[bytes], Awaitable[None]]]] = None,
        session_id: Optional[str] = None,
        handle_resize: Optional[Callable[[PtySize], Awaitable[PtySessionInfo]]] = None,
        handle_kill: Optional[Callable[[], Awaitable[None]]] = None,
    ):
        """
        Initialize the PTY handle.

        Args:
            ws: WebSocket connection to the PTY session
            on_data: Optional callback function to handle PTY output data
            session_id: Optional session ID for resize/kill operations
            handle_resize: Optional callback for resizing the PTY
            handle_kill: Optional callback for killing the PTY
        """
        self._ws = ws
        self._on_data = on_data
        self._session_id = session_id
        self._handle_resize = handle_resize
        self._handle_kill = handle_kill
        self._exit_code: Optional[int] = None
        self._error: Optional[str] = None
        self._connected = False
        self._connection_established = False

        # Start handling WebSocket events
        self._wait = asyncio.create_task(self._handle_websocket())

    @property
    def session_id(self) -> Optional[str]:
        """Session ID of the PTY session"""
        return self._session_id

    @property
    def exit_code(self) -> Optional[int]:
        """Exit code of the PTY process (if terminated)"""
        return self._exit_code

    @property
    def error(self) -> Optional[str]:
        """Error message if the PTY failed"""
        return self._error

    def is_connected(self) -> bool:
        """Check if connected to the PTY session"""
        # For websockets ClientConnection, check if the connection is not closed
        try:
            return self._connected and not self._ws.close_code
        except AttributeError:
            # Fallback if close_code is not available
            return self._connected

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
                raise DaytonaError("PTY connection timeout")

            # Check if WebSocket is closed (handle different websocket implementations)
            is_closed = False
            try:
                is_closed = bool(self._ws.close_code)
            except AttributeError:
                # Fallback - assume not closed if we can't check
                pass

            if is_closed or self._error:
                raise DaytonaError(self._error or "Connection failed")

            await asyncio.sleep(0.1)

    async def send_input(self, data: Union[str, bytes]) -> None:
        """
        Send input data to the PTY.

        Args:
            data: Input data to send (string or bytes)

        Raises:
            ConnectionError: If PTY is not connected
            RuntimeError: If sending input fails
        """
        if not self.is_connected():
            raise DaytonaError("PTY is not connected")

        try:
            if isinstance(data, str):
                await self._ws.send(data.encode("utf-8"))
            else:
                await self._ws.send(data)
        except Exception as e:
            raise DaytonaError(f"Failed to send input to PTY: {e}") from e

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
            self._wait.cancel()
            try:
                await self._wait
            except asyncio.CancelledError:
                pass

        # Close WebSocket if not already closed
        try:
            if not self._ws.close_code:
                await self._ws.close()
        except AttributeError:
            # Fallback - try to close anyway
            try:
                await self._ws.close()
            except Exception:
                pass  # Ignore close errors

    async def _handle_websocket(self) -> None:
        """Handle WebSocket messages and connection lifecycle"""
        try:
            self._connected = True

            async for message in self._ws:
                await self._handle_message(message)

            # If we exit the loop normally, the connection was closed gracefully
            # Simulate a close event with normal close code
            class CloseEvent:
                def __init__(self, code=1000, reason=""):
                    self.code = code
                    self.reason = reason

            await self._handle_close(CloseEvent())

        except websockets.exceptions.ConnectionClosedOK as e:
            await self._handle_close(e)
        except websockets.exceptions.ConnectionClosedError as e:
            await self._handle_close(e)
        except Exception as e:
            self._error = f"Unexpected error: {e}"
        finally:
            self._connected = False

    async def _handle_message(self, message: Union[str, bytes]) -> None:
        """Handle individual WebSocket messages"""
        try:
            if isinstance(message, str):
                # Try to parse as control message first
                try:
                    control_msg = json.loads(message)
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

    async def _handle_control_message(self, control_msg: dict) -> None:
        """Handle control messages from the PTY server"""
        status = control_msg.get("status")

        if status == "connected":
            self._connection_established = True
        elif status == "error":
            self._error = control_msg.get("error", "Unknown connection error")
            self._connected = False

    async def _handle_close(self, close_event) -> None:
        """Handle WebSocket close event"""
        self._connected = False

        # In websockets library, the close event is a ConnectionClosed exception
        # The close code is available as close_event.code and reason as close_event.reason
        close_code = getattr(close_event, "code", None)
        close_reason = getattr(close_event, "reason", None)

        # Parse structured exit data from close reason
        if close_reason:
            try:
                exit_data = json.loads(close_reason)
                if isinstance(exit_data.get("exitCode"), int):
                    self._exit_code = exit_data["exitCode"]
                    # Store exit reason if provided
                    if exit_data.get("exitReason"):
                        self._error = exit_data["exitReason"]

                # Handle error messages from server
                if exit_data.get("error"):
                    self._error = exit_data["error"]

            except (json.JSONDecodeError, ValueError):
                # Default to exit code 0 for normal close
                if close_code == 1000:
                    self._exit_code = 0

        # Default to exit code 0 if we can't parse it and it was a normal close
        if self._exit_code is None and close_code == 1000:
            self._exit_code = 0

    async def _call_data_handler(self, data: bytes) -> None:
        """Call the data handler, supporting both sync and async callbacks"""
        if self._on_data:
            result = self._on_data(data)
            if inspect.isawaitable(result):
                await result
