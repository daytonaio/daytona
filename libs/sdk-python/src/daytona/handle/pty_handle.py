# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import json
import time
from typing import Callable, Generator, Optional, Union

from daytona_api_client_async import PtySessionInfo
from websockets.exceptions import ConnectionClosedError, ConnectionClosedOK
from websockets.sync.client import ClientConnection

from .._utils.errors import DaytonaError
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
        ws: ClientConnection,
        session_id: str,
        handle_resize: Optional[Callable[[PtySize], PtySessionInfo]] = None,
        handle_kill: Optional[Callable[[], None]] = None,
    ):
        """
        Initialize the PTY handle.

        Args:
            ws: Connected WebSocket client connection
            session_id: Session ID of the PTY session
            handle_resize: Optional callback for resizing the PTY
            handle_kill: Optional callback for killing the PTY
        """
        self._ws = ws
        self._session_id = session_id
        self._handle_resize = handle_resize
        self._handle_kill = handle_kill

        self._connected = True  # WebSocket is already connected
        self._connection_established = False  # Still need to wait for control message
        self._exit_code: Optional[int] = None
        self._error: Optional[str] = None

    @property
    def session_id(self) -> str:
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

        # Wait for connection established control message
        start_time = time.time()
        while not self._connection_established:
            if time.time() - start_time > timeout:
                raise DaytonaError("PTY connection timeout")

            if self._error:
                raise DaytonaError(self._error or "Connection failed")

            # Try to receive a control message
            try:
                message = self._ws.recv(timeout=0.1)
                if isinstance(message, str):
                    try:
                        control_msg = json.loads(message)
                        if control_msg.get("type") == "control":
                            self._handle_control_message(control_msg)
                    except (json.JSONDecodeError, ValueError):
                        pass
            except TimeoutError:
                continue  # Keep waiting
            except (ConnectionClosedOK, ConnectionClosedError) as e:
                raise DaytonaError("Connection closed during setup") from e

    def send_input(self, data: Union[str, bytes]) -> None:
        """
        Send input data to the PTY.

        Args:
            data: Input data to send (string or bytes)

        Raises:
            DaytonaError: If PTY is not connected or sending fails
        """
        if not self.is_connected():
            raise DaytonaError("PTY is not connected")

        try:
            if isinstance(data, str):
                self._ws.send(data.encode("utf-8"))
            else:
                self._ws.send(data)
        except Exception as e:
            raise DaytonaError(f"Failed to send input to PTY: {e}") from e

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
        if not self._ws:
            return

        try:
            for message in self._ws:
                if isinstance(message, str):
                    # Try to parse as control message
                    try:
                        control_msg = json.loads(message)
                        if control_msg.get("type") == "control":
                            self._handle_control_message(control_msg)
                            continue
                    except (json.JSONDecodeError, ValueError):
                        # Not a control message, treat as PTY output
                        pass

                    # Convert string to bytes for PTY output
                    yield message.encode("utf-8")
                else:
                    # Binary PTY data
                    yield message

            # If we exit the loop normally, the connection was closed gracefully
            # Simulate a close event with normal close code
            class CloseEvent:
                def __init__(self, code=1000, reason=""):
                    self.code = code
                    self.reason = reason

            self._handle_close(CloseEvent())

        except (ConnectionClosedOK, ConnectionClosedError) as e:
            # Handle connection close and extract exit data
            self._handle_close(e)
        except Exception as e:
            if not self._error:
                self._error = f"WebSocket error: {e}"

    def wait(self, on_data: Optional[Callable[[bytes], None]] = None, timeout: Optional[float] = None) -> PtyResult:
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

                # Check timeout
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
        if self._ws:
            try:
                self._ws.close()
            except Exception:
                pass  # Ignore close errors
            finally:
                self._ws = None
                self._connected = False

    def _handle_control_message(self, control_msg: dict) -> None:
        """Handle control messages from the PTY server"""
        status = control_msg.get("status")

        if status == "connected":
            self._connection_established = True
        elif status == "error":
            self._error = control_msg.get("error", "Unknown connection error")
            self._connected = False

    def _handle_close(self, close_event) -> None:
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
