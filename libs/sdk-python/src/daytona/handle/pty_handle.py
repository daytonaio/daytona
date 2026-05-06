# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
import time
from collections.abc import Callable, Generator
from contextlib import AbstractContextManager
from typing import Any, cast

from httpx_ws import WebSocketDisconnect, WebSocketSession
from wsproto.events import BytesMessage, CloseConnection, TextMessage

from daytona_toolbox_api_client import PtySessionInfo

from ..common.errors import DaytonaConnectionError, DaytonaError, DaytonaTimeoutError
from ..common.pty import PtyResult, PtySize


class PtyHandle:
    """
    Synchronous PTY session handle for managing a single PTY session.

    Provides methods for sending input, waiting for connection and exit, and managing WebSocket connections.
    Uses an iterator-based approach for handling PTY events.

    Example usage:
    ```python
    # Connect to a PTY session
    pty_handle = sandbox.process.connect_pty_session('my-session', handle_output)

    # Wait for connection
    pty_handle.wait_for_connection()

    # Send commands
    pty_handle.send_input('ls -la\n')

    # Wait for completion with callbacks
    def handle_data(data: bytes):
        print(data.decode('utf-8'), end='')

    result = pty_handle.wait(on_data=handle_data)
    print(f"PTY exited with code: {result.exit_code}")

    # Clean up
    pty_handle.disconnect()
    ```
    """

    def __init__(
        self,
        ws: WebSocketSession,
        session_id: str,
        handle_resize: Callable[[PtySize], PtySessionInfo] | None = None,
        handle_kill: Callable[[], None] | None = None,
        ws_context_manager: AbstractContextManager[WebSocketSession] | None = None,
    ):
        """
        Initialize the PTY handle.

        Args:
            ws: Open httpx_ws WebSocketSession (long-lived; the handle owns its lifecycle)
            session_id: Session ID of the PTY session
            handle_resize: Optional callback for resizing the PTY
            handle_kill: Optional callback for killing the PTY
            ws_context_manager: The httpx_ws.connect_ws() context manager that produced ``ws``.
                When provided, ``disconnect()`` calls its ``__exit__`` so the underlying HTTP
                stream is released back to the httpx connection pool. Pass ``None`` only if
                the caller manages the context manager itself.
        """
        self._ws: WebSocketSession | None = ws
        self._ws_cm: AbstractContextManager[WebSocketSession] | None = ws_context_manager
        self._session_id: str = session_id
        self._handle_resize: Callable[[PtySize], PtySessionInfo] | None = handle_resize
        self._handle_kill: Callable[[], None] | None = handle_kill

        self._connected: bool = True  # WebSocket is already connected
        self._connection_established: bool = False  # Still need to wait for control message
        self._exit_code: int | None = None
        self._error: str | None = None

    @property
    def session_id(self) -> str:
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
        return self._connected and self._ws is not None

    def wait_for_connection(self, timeout: float = 10.0) -> None:
        """
        Wait for the WebSocket connection to be established.

        Args:
            timeout: Connection timeout in seconds

        Raises:
            DaytonaError: If connection timeout or connection fails
        """
        if self._connection_established:
            return

        if self._ws is None:
            raise DaytonaConnectionError("WebSocket connection is not available")

        start_time = time.time()
        while not self._connection_established:
            if time.time() - start_time > timeout:
                raise DaytonaTimeoutError("PTY connection timeout")

            if self._error:
                raise DaytonaConnectionError(self._error or "Connection failed")

            try:
                event = self._ws.receive(timeout=0.1)
            except TimeoutError:
                continue
            except WebSocketDisconnect as e:
                raise DaytonaConnectionError("Connection closed during setup") from e

            if isinstance(event, TextMessage):
                try:
                    control_msg: dict[str, object] = json.loads(event.data)
                    if control_msg.get("type") == "control":
                        self._handle_control_message(control_msg)
                except (json.JSONDecodeError, ValueError):
                    pass

    def send_input(self, data: str | bytes) -> None:
        """
        Send input data to the PTY.

        Args:
            data: Input data to send (string or bytes)

        Raises:
            DaytonaError: If PTY is not connected or sending fails
        """
        if not self.is_connected():
            raise DaytonaConnectionError("PTY is not connected")

        if self._ws is None:
            raise DaytonaConnectionError("WebSocket connection is not available")

        try:
            if isinstance(data, str):
                self._ws.send_bytes(data.encode("utf-8"))
            else:
                self._ws.send_bytes(data)
        except Exception as e:
            raise DaytonaConnectionError(f"Failed to send input to PTY: {e}") from e

    def resize(self, pty_size: PtySize) -> PtySessionInfo:
        """
        Resize the PTY terminal.

        Args:
            pty_size: PtySize object containing the new terminal dimensions

        Raises:
            DaytonaError: If resize handler is not available or resize fails
        """
        if not self._handle_resize:
            raise DaytonaError("Resize handler not available")

        try:
            return self._handle_resize(pty_size)
        except Exception as e:
            raise DaytonaError(f"Failed to resize PTY: {e}") from e

    def kill(self) -> None:
        """
        Kill the PTY process.

        Raises:
            DaytonaError: If kill handler is not available or kill fails
        """
        if not self._handle_kill:
            raise DaytonaError("Kill handler not available")

        try:
            self._handle_kill()
        except Exception as e:
            raise DaytonaError(f"Failed to kill PTY: {e}") from e

    def __iter__(self):
        """Iterator protocol for handling PTY events"""
        return self._handle_events()

    def _handle_events(self) -> Generator[bytes, None, None]:
        """
        Generator that yields PTY data events.

        Yields:
            bytes: PTY output data
        """
        ws = self._ws
        if ws is None:
            return

        close_code: int | None = None
        close_reason: str | None = None

        try:
            while True:
                try:
                    event = ws.receive()
                except WebSocketDisconnect as e:
                    close_code = e.code
                    close_reason = e.reason
                    break

                if isinstance(event, TextMessage):
                    text = event.data
                    try:
                        control_msg: dict[str, object] = json.loads(text)
                        if control_msg.get("type") == "control":
                            self._handle_control_message(control_msg)
                            continue
                    except (json.JSONDecodeError, ValueError):
                        pass
                    yield text.encode("utf-8")
                elif isinstance(event, BytesMessage):
                    # wsproto's data is bytes | bytearray; coerce for the typed yield.
                    yield bytes(event.data)
                elif isinstance(event, CloseConnection):
                    close_code = event.code
                    close_reason = event.reason
                    break
                # Ping/Pong/etc. — wsproto auto-handles ping; just continue.

        except Exception as e:
            if not self._error:
                self._error = f"WebSocket error: {e}"
            return

        self._handle_close(close_code, close_reason)

    def wait(self, on_data: Callable[[bytes], None] | None = None, timeout: float | None = None) -> PtyResult:
        """
        Wait for the PTY process to exit and return the result.

        Args:
            on_data: Optional callback for handling PTY output data
            timeout: Optional timeout in seconds

        Returns:
            PtyResult: Result containing exit code and error (if any)
        """
        start_time = time.time()

        try:
            for data in self:
                if on_data:
                    on_data(data)

                if timeout and (time.time() - start_time) > timeout:
                    break

        except StopIteration:
            pass
        except Exception as e:
            if not self._error:
                self._error = str(e)

        return PtyResult(
            exit_code=self._exit_code,
            error=self._error,
        )

    def disconnect(self) -> None:
        """Disconnect from the PTY session"""
        # Close via the original context manager so httpx releases its stream resources
        # back to the connection pool. Falling back to ws.close() covers cases where the
        # caller passed in a bare WebSocketSession (e.g. tests).
        if self._ws_cm is not None:
            try:
                _ = self._ws_cm.__exit__(None, None, None)
            except Exception:
                pass
            self._ws_cm = None
            self._ws = None
            self._connected = False
        elif self._ws is not None:
            try:
                self._ws.close()
            except Exception:
                pass
            self._ws = None
            self._connected = False

    def _handle_control_message(self, control_msg: dict[str, object]) -> None:
        """Handle control messages from the PTY server"""
        status = control_msg.get("status")

        if status == "connected":
            self._connection_established = True
        elif status == "error":
            self._error = cast(str, control_msg.get("error", "Unknown connection error"))
            self._connected = False

    def _handle_close(self, close_code: int | None, close_reason: str | None) -> None:
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

        if self._exit_code is None and close_code == 1000:
            self._exit_code = 0
