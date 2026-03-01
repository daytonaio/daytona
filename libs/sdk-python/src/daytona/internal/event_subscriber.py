# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""WebSocket event subscriber for real-time sandbox events via Socket.IO."""

from __future__ import annotations

import asyncio
import logging
import threading
from typing import Any, Callable

import socketio

logger = logging.getLogger(__name__)

EventHandler = Callable[[str, Any], Any]


class AsyncEventSubscriber:
    """Async event subscriber that connects to the Socket.IO notification gateway."""

    _api_url: str
    _token: str
    _organization_id: str | None
    _sio: socketio.AsyncClient | None
    _connected: bool
    _failed: bool
    _fail_error: str | None
    _listeners: dict[str, list[EventHandler]]
    _lock: asyncio.Lock
    _disconnect_task: asyncio.Task[None] | None

    _DISCONNECT_DELAY: float = 30.0

    def __init__(self, api_url: str, token: str, organization_id: str | None = None):
        self._api_url = api_url
        self._token = token
        self._organization_id = organization_id
        self._sio = None
        self._connected = False
        self._failed = False
        self._fail_error = None
        self._listeners = {}
        self._lock = asyncio.Lock()
        self._disconnect_task = None

    async def connect(self, timeout: float = 5.0) -> None:
        """Establish the Socket.IO connection. Raises on failure."""
        if self._connected:
            return

        origin = self._api_url.rstrip("/")
        if origin.endswith("/api"):
            origin = origin[:-4]

        sio = socketio.AsyncClient(
            reconnection=True,
            reconnection_attempts=10,
            reconnection_delay=1,
            reconnection_delay_max=30,
            logger=False,
            engineio_logger=False,
        )
        self._sio = sio

        subscriber = self

        @sio.event  # type: ignore[misc]
        async def connect() -> None:  # pyright: ignore[reportUnusedFunction]
            subscriber._connected = True
            subscriber._failed = False
            subscriber._fail_error = None

        @sio.event  # type: ignore[misc]
        async def disconnect() -> None:  # pyright: ignore[reportUnusedFunction]
            subscriber._connected = False

        @sio.event  # type: ignore[misc]
        async def connect_error(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            subscriber._connected = False
            subscriber._failed = True
            subscriber._fail_error = f"WebSocket connection failed: {data}"

        @sio.on("sandbox.state.updated")  # pyright: ignore[reportOptionalCall, reportUntypedFunctionDecorator]
        async def on_state_updated(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            sandbox_id: str | None = _extract_sandbox_id(data)
            if sandbox_id:
                await subscriber._dispatch(sandbox_id, "state.updated", data)

        @sio.on("sandbox.desired-state.updated")  # pyright: ignore[reportOptionalCall, reportUntypedFunctionDecorator]
        async def on_desired_state_updated(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            sandbox_id: str | None = _extract_sandbox_id(data)
            if sandbox_id:
                await subscriber._dispatch(sandbox_id, "desired-state.updated", data)

        @sio.on("sandbox.created")  # pyright: ignore[reportOptionalCall, reportUntypedFunctionDecorator]
        async def on_created(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            sandbox_id: str | None = _extract_sandbox_id_from_root(data)
            if sandbox_id:
                await subscriber._dispatch(sandbox_id, "created", data)

        try:
            await asyncio.wait_for(
                sio.connect(
                    origin,
                    socketio_path="/api/socket.io/",
                    auth={"token": self._token},
                    transports=["websocket"],
                    headers={},
                    wait=True,
                ),
                timeout=timeout,
            )
        except asyncio.TimeoutError:
            self._failed = True
            self._fail_error = "WebSocket connection timed out"
            raise ConnectionError(self._fail_error) from None
        except Exception as e:
            self._failed = True
            self._fail_error = f"WebSocket connection failed: {e}"
            raise ConnectionError(self._fail_error) from e

    def subscribe(self, sandbox_id: str, handler: EventHandler) -> Callable[[], None]:
        """Register a handler for events targeting a specific sandbox. Returns unsubscribe function."""
        # Cancel any pending delayed disconnect
        if self._disconnect_task and not self._disconnect_task.done():
            _ = self._disconnect_task.cancel()
            self._disconnect_task = None

        if sandbox_id not in self._listeners:
            self._listeners[sandbox_id] = []
        self._listeners[sandbox_id].append(handler)

        def unsubscribe() -> None:
            handlers = self._listeners.get(sandbox_id)
            if handlers and handler in handlers:
                handlers.remove(handler)
                if not handlers:
                    del self._listeners[sandbox_id]
            # Schedule delayed disconnect when no sandboxes are listening anymore
            if not self._listeners:
                self._schedule_delayed_disconnect()

        return unsubscribe

    def _schedule_delayed_disconnect(self) -> None:
        async def _delayed() -> None:
            await asyncio.sleep(self._DISCONNECT_DELAY)
            if not self._listeners:
                await self.disconnect()

        try:
            self._disconnect_task = asyncio.get_event_loop().create_task(_delayed())
        except RuntimeError:
            pass  # No event loop - skip delayed disconnect

    @property
    def is_connected(self) -> bool:
        return self._connected

    @property
    def is_failed(self) -> bool:
        return self._failed

    @property
    def fail_error(self) -> str | None:
        return self._fail_error

    async def disconnect(self) -> None:
        """Disconnect and clean up resources."""
        if self._sio:
            await self._sio.disconnect()
        self._connected = False
        self._listeners.clear()

    async def _dispatch(self, sandbox_id: str, event_type: str, data: Any) -> None:
        handlers = self._listeners.get(sandbox_id, [])
        for handler in list(handlers):
            try:
                result = handler(event_type, data)
                if asyncio.iscoroutine(result):
                    await result
            except Exception:
                pass  # Don't let handler errors break other handlers


class SyncEventSubscriber:
    """Sync event subscriber using socketio.Client on a background thread."""

    _api_url: str
    _token: str
    _organization_id: str | None
    _sio: socketio.Client | None
    _connected: bool
    _failed: bool
    _fail_error: str | None
    _listeners: dict[str, list[EventHandler]]
    _lock: threading.Lock
    _bg_thread: threading.Thread | None
    _disconnect_timer: threading.Timer | None

    _DISCONNECT_DELAY: float = 30.0

    def __init__(self, api_url: str, token: str, organization_id: str | None = None):
        self._api_url = api_url
        self._token = token
        self._organization_id = organization_id
        self._sio = None
        self._connected = False
        self._failed = False
        self._fail_error = None
        self._listeners = {}
        self._lock = threading.Lock()
        self._bg_thread = None
        self._disconnect_timer = None

    def connect(self, timeout: float = 5.0) -> None:
        """Establish the Socket.IO connection. Raises on failure."""
        if self._connected:
            return

        origin = self._api_url.rstrip("/")
        if origin.endswith("/api"):
            origin = origin[:-4]

        sio = socketio.Client(
            reconnection=True,
            reconnection_attempts=10,
            reconnection_delay=1,
            reconnection_delay_max=30,
            logger=False,
            engineio_logger=False,
        )
        self._sio = sio

        connected_event = threading.Event()
        error_holder: list[str] = []
        subscriber = self

        @sio.event  # type: ignore[misc]
        def connect() -> None:  # pyright: ignore[reportUnusedFunction]
            subscriber._connected = True
            subscriber._failed = False
            subscriber._fail_error = None
            connected_event.set()

        @sio.event  # type: ignore[misc]
        def disconnect() -> None:  # pyright: ignore[reportUnusedFunction]
            subscriber._connected = False

        @sio.event  # type: ignore[misc]
        def connect_error(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            subscriber._connected = False
            subscriber._failed = True
            subscriber._fail_error = f"WebSocket connection failed: {data}"
            error_holder.append(str(data))
            connected_event.set()

        @sio.on("sandbox.state.updated")  # pyright: ignore[reportOptionalCall, reportUntypedFunctionDecorator]
        def on_state_updated(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            sandbox_id: str | None = _extract_sandbox_id(data)
            if sandbox_id:
                subscriber._dispatch(sandbox_id, "state.updated", data)

        @sio.on("sandbox.desired-state.updated")  # pyright: ignore[reportOptionalCall, reportUntypedFunctionDecorator]
        def on_desired_state_updated(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            sandbox_id: str | None = _extract_sandbox_id(data)
            if sandbox_id:
                subscriber._dispatch(sandbox_id, "desired-state.updated", data)

        @sio.on("sandbox.created")  # pyright: ignore[reportOptionalCall, reportUntypedFunctionDecorator]
        def on_created(data: Any) -> None:  # pyright: ignore[reportUnusedFunction]
            sandbox_id: str | None = _extract_sandbox_id_from_root(data)
            if sandbox_id:
                subscriber._dispatch(sandbox_id, "created", data)

        try:
            sio.connect(
                origin,
                socketio_path="/api/socket.io/",
                auth={"token": self._token},
                transports=["websocket"],
                headers={},
                wait=True,
                wait_timeout=int(timeout),
            )
        except Exception as e:
            self._failed = True
            self._fail_error = f"WebSocket connection failed: {e}"
            raise ConnectionError(self._fail_error) from e

        if not self._connected:
            self._failed = True
            err = error_holder[0] if error_holder else "unknown error"
            self._fail_error = f"WebSocket connection failed: {err}"
            raise ConnectionError(self._fail_error)

    def subscribe(self, sandbox_id: str, handler: EventHandler) -> Callable[[], None]:
        """Register a handler for events targeting a specific sandbox. Returns unsubscribe function."""
        # Cancel any pending delayed disconnect
        if self._disconnect_timer:
            self._disconnect_timer.cancel()
            self._disconnect_timer = None

        with self._lock:
            if sandbox_id not in self._listeners:
                self._listeners[sandbox_id] = []
            self._listeners[sandbox_id].append(handler)

        def unsubscribe() -> None:
            with self._lock:
                handlers = self._listeners.get(sandbox_id)
                if handlers and handler in handlers:
                    handlers.remove(handler)
                    if not handlers:
                        del self._listeners[sandbox_id]
                should_cleanup = not self._listeners

            # Schedule delayed disconnect when no sandboxes are listening anymore
            if should_cleanup:
                self._disconnect_timer = threading.Timer(self._DISCONNECT_DELAY, self._delayed_disconnect)
                self._disconnect_timer.daemon = True
                self._disconnect_timer.start()

        return unsubscribe

    def _delayed_disconnect(self) -> None:
        with self._lock:
            if self._listeners:
                return  # New subscribers arrived during the delay
        self.disconnect()

    @property
    def is_connected(self) -> bool:
        return self._connected

    @property
    def is_failed(self) -> bool:
        return self._failed

    @property
    def fail_error(self) -> str | None:
        return self._fail_error

    def disconnect(self) -> None:
        """Disconnect and clean up resources."""
        if self._sio:
            self._sio.disconnect()
        self._connected = False
        with self._lock:
            self._listeners.clear()

    def _dispatch(self, sandbox_id: str, event_type: str, data: Any) -> None:
        with self._lock:
            handlers = list(self._listeners.get(sandbox_id, []))
        for handler in handlers:
            try:
                handler(event_type, data)
            except Exception:
                pass  # Don't let handler errors break other handlers


def _extract_sandbox_id(data: Any) -> str | None:
    """Extract sandbox ID from an event payload with nested sandbox object."""
    if not isinstance(data, dict):
        return None
    sandbox_raw: object = data.get("sandbox")  # pyright: ignore[reportUnknownVariableType]
    if not isinstance(sandbox_raw, dict):
        return None
    sid: object = sandbox_raw.get("id")  # pyright: ignore[reportUnknownVariableType]
    return str(sid) if isinstance(sid, str) else None


def _extract_sandbox_id_from_root(data: Any) -> str | None:
    """Extract sandbox ID from a top-level event payload."""
    if not isinstance(data, dict):
        return None
    sid: object = data.get("id")  # pyright: ignore[reportUnknownVariableType]
    return str(sid) if isinstance(sid, str) else None
