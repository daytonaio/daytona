# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""WebSocket event subscriber for real-time events via Socket.IO."""

from __future__ import annotations

import asyncio
import logging
import threading
from collections.abc import Awaitable
from typing import Any, Callable, Union

import socketio

logger = logging.getLogger(__name__)

# Handler receives (event_name, raw_data).
EventHandler = Callable[[str, Any], None]
AsyncEventHandler = Callable[[str, Any], Union[Awaitable[None], None]]


class AsyncEventSubscriber:
    """Async event subscriber that connects to the Socket.IO notification gateway."""

    _api_url: str
    _token: str
    _organization_id: str | None
    _sio: socketio.AsyncClient | None
    _connected: bool
    _failed: bool
    _fail_error: str | None
    _listeners: dict[str, list[AsyncEventHandler]]
    _registered_events: set[str]
    _disconnect_task: asyncio.Task[None] | None
    _subscription_timers: dict[str, threading.Timer]
    _subscription_ttls: dict[str, float]
    _disconnect_generation: int
    _lock: asyncio.Lock

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
        self._registered_events = set()
        self._disconnect_task = None
        self._subscription_timers = {}
        self._subscription_ttls = {}
        self._connect_task: asyncio.Task[None] | None = None
        self._disconnect_generation = 0
        self._lock = asyncio.Lock()

    def ensure_connected(self) -> None:
        """Idempotent: ensure a connection attempt is in progress or already established.

        Non-blocking. Creates a background task if not already connected and no
        attempt is currently running.
        """
        if self._connected:
            return
        if self._connect_task is not None and not self._connect_task.done():
            return

        async def _connect() -> None:
            try:
                await self.connect()
            except Exception:
                pass  # Callers check is_connected when they need it

        try:
            loop = asyncio.get_running_loop()
        except RuntimeError:
            pass  # No event loop — will connect on first await
        else:
            self._connect_task = loop.create_task(_connect())

    async def connect(self, timeout: float = 5.0) -> None:
        """Establish the Socket.IO connection. Raises on failure."""
        async with self._lock:
            if self._connected:
                return
            old_sio = self._sio
            self._sio = None

        if old_sio:
            await old_sio.disconnect()

        origin = self._api_url.rstrip("/")
        if origin.endswith("/api"):
            origin = origin[:-4]

        sio = socketio.AsyncClient(
            reconnection=True,
            reconnection_attempts=0,
            reconnection_delay=1,
            reconnection_delay_max=30,
            logger=False,
            engineio_logger=False,
        )

        async with self._lock:
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

        # Re-register any events that were added before the socket was created
        async with self._lock:
            pending_events = list(self._registered_events)
            self._registered_events.clear()
            self._register_events(pending_events)

        connect_url = origin
        if self._organization_id:
            connect_url = f"{origin}?organizationId={self._organization_id}"

        try:
            await asyncio.wait_for(
                sio.connect(
                    connect_url,
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

    def subscribe(
        self,
        resource_id: str,
        handler: AsyncEventHandler,
        events: list[str],
        ttl_seconds: float = 0,
    ) -> Callable[[], None]:
        """Subscribe to specific events for a resource.

        Args:
            resource_id: The ID of the resource (e.g. sandbox ID).
            handler: Callback receiving (event_name, raw_data).
            events: List of Socket.IO event names to listen for (e.g. ["sandbox.state.updated", "sandbox.created"]).

        Returns:
            Unsubscribe function.
        """
        self.ensure_connected()

        self._disconnect_generation += 1

        if self._disconnect_task and not self._disconnect_task.done():
            _ = self._disconnect_task.cancel()
            self._disconnect_task = None

        # Register any new events with the Socket.IO client
        self._register_events(events)

        if resource_id not in self._listeners:
            self._listeners[resource_id] = []
        self._listeners[resource_id].append(handler)

        if ttl_seconds > 0:
            self._subscription_ttls[resource_id] = ttl_seconds
            self._start_subscription_timer(resource_id)
        else:
            self._clear_subscription_timer(resource_id)
            self._subscription_ttls.pop(resource_id, None)

        def unsubscribe() -> None:
            handlers = self._listeners.get(resource_id)
            if handlers and handler in handlers:
                handlers.remove(handler)
                if not handlers:
                    self._unsubscribe_resource(resource_id)

        return unsubscribe

    def refresh_subscription(self, resource_id: str) -> bool:
        if resource_id not in self._subscription_ttls:
            return False

        self._start_subscription_timer(resource_id)
        return True

    def _register_events(self, events: list[str]) -> None:
        """Register Socket.IO event handlers (idempotent — each event is registered once)."""
        subscriber = self

        for event_name in events:
            if event_name in self._registered_events:
                continue
            self._registered_events.add(event_name)

            # If socket isn't created yet, the event will be registered when connect() runs
            if not self._sio:
                continue

            def _make_handler(evt: str) -> Callable[..., Any]:
                async def _handler(data: Any) -> None:
                    resource_id = _extract_id_from_event(data)
                    if resource_id:
                        await subscriber._dispatch(resource_id, evt, data)

                return _handler

            self._sio.on(event_name, _make_handler(event_name))  # pyright: ignore[reportUnusedCallResult]

    def _schedule_delayed_disconnect(self) -> None:
        generation = self._disconnect_generation

        async def _delayed() -> None:
            await asyncio.sleep(self._DISCONNECT_DELAY)
            if generation != self._disconnect_generation or self._listeners:
                return
            await self._disconnect(expected_generation=generation)

        try:
            self._disconnect_task = asyncio.get_event_loop().create_task(_delayed())
        except RuntimeError:
            pass

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
        await self._disconnect()

    async def _disconnect(self, expected_generation: int | None = None) -> None:
        if self._connect_task and not self._connect_task.done():
            self._connect_task.cancel()

        async with self._lock:
            if expected_generation is not None:
                if expected_generation != self._disconnect_generation or self._listeners:
                    return

            sio = self._sio
            self._sio = None
            self._clear_subscription_timers()
            self._listeners.clear()
            self._registered_events.clear()
            self._connected = False

        if sio:
            await sio.disconnect()

    async def _dispatch(self, resource_id: str, event_name: str, data: Any) -> None:
        for handler in list(self._listeners.get(resource_id, [])):
            try:
                result = handler(event_name, data)
                if result is not None and asyncio.iscoroutine(result):
                    await result
            except Exception:
                pass

    def _start_subscription_timer(self, resource_id: str) -> None:
        ttl_seconds = self._subscription_ttls.get(resource_id)
        if ttl_seconds is None or ttl_seconds <= 0:
            return

        self._clear_subscription_timer(resource_id)

        try:
            loop = asyncio.get_running_loop()
        except RuntimeError:
            loop = None

        current_timer: threading.Timer | None = None

        def _expire() -> None:
            if self._subscription_timers.get(resource_id) is not current_timer:
                return

            self._clear_subscription_timer(resource_id)
            self._subscription_ttls.pop(resource_id, None)
            if loop is not None and loop.is_running():
                loop.call_soon_threadsafe(self._unsubscribe_resource, resource_id)
            else:
                self._unsubscribe_resource(resource_id)

        current_timer = threading.Timer(ttl_seconds, _expire)
        current_timer.daemon = True
        self._subscription_timers[resource_id] = current_timer
        current_timer.start()

    def _clear_subscription_timer(self, resource_id: str) -> None:
        timer = self._subscription_timers.pop(resource_id, None)
        if timer is not None:
            timer.cancel()

    def _clear_subscription_timers(self) -> None:
        for timer in self._subscription_timers.values():
            timer.cancel()
        self._subscription_timers.clear()
        self._subscription_ttls.clear()

    def _unsubscribe_resource(self, resource_id: str) -> None:
        if resource_id in self._listeners:
            del self._listeners[resource_id]

        self._clear_subscription_timer(resource_id)
        self._subscription_ttls.pop(resource_id, None)

        if not self._listeners:
            self._schedule_delayed_disconnect()


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
    _registered_events: set[str]
    _lock: threading.Lock
    _bg_thread: threading.Thread | None
    _disconnect_timer: threading.Timer | None
    _subscription_timers: dict[str, threading.Timer]
    _subscription_ttls: dict[str, float]
    _disconnect_generation: int
    _closed: bool

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
        self._registered_events = set()
        self._lock = threading.Lock()
        self._bg_thread = None
        self._disconnect_timer = None
        self._subscription_timers = {}
        self._subscription_ttls = {}
        self._disconnect_generation = 0
        self._closed = False

    def ensure_connected(self) -> None:
        """Idempotent: ensure a connection attempt is in progress or already established.

        Non-blocking. Starts a background thread if not already connected and no
        attempt is currently running.
        """
        if self._closed or self._connected:
            return
        if self._bg_thread is not None and self._bg_thread.is_alive():
            return

        def _connect() -> None:
            try:
                self.connect()
            except Exception:
                pass  # Callers check is_connected when they need it

        self._bg_thread = threading.Thread(target=_connect, daemon=True)
        self._bg_thread.start()

    def connect(self, timeout: float = 5.0) -> None:
        """Establish the Socket.IO connection. Raises on failure."""
        with self._lock:
            if self._closed or self._connected:
                return
            old_sio = self._sio
            self._sio = None

        if old_sio:
            old_sio.disconnect()

        origin = self._api_url.rstrip("/")
        if origin.endswith("/api"):
            origin = origin[:-4]

        sio = socketio.Client(
            reconnection=True,
            reconnection_attempts=0,
            reconnection_delay=1,
            reconnection_delay_max=30,
            logger=False,
            engineio_logger=False,
        )
        with self._lock:
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

        try:
            # Re-register any events that were added before the socket was created
            with self._lock:
                pending_events = list(self._registered_events)
                self._registered_events.clear()
            self._register_events(pending_events)

            connect_url = origin
            if self._organization_id:
                connect_url = f"{origin}?organizationId={self._organization_id}"

            sio.connect(
                connect_url,
                socketio_path="/api/socket.io/",
                auth={"token": self._token},
                transports=["websocket"],
                headers={},
                wait=True,
                wait_timeout=int(timeout),
            )

            if self._closed:
                self._sio = None
                self._connected = False
                sio.disconnect()
                return
        except Exception as e:
            self._failed = True
            self._fail_error = f"WebSocket connection failed: {e}"
            raise ConnectionError(self._fail_error) from e

        if not self._connected:
            self._failed = True
            err = error_holder[0] if error_holder else "unknown error"
            self._fail_error = f"WebSocket connection failed: {err}"
            raise ConnectionError(self._fail_error)

    def subscribe(
        self,
        resource_id: str,
        handler: EventHandler,
        events: list[str],
        ttl_seconds: float = 0,
    ) -> Callable[[], None]:
        """Subscribe to specific events for a resource.

        Args:
            resource_id: The ID of the resource (e.g. sandbox ID).
            handler: Callback receiving (event_name, raw_data).
            events: List of Socket.IO event names to listen for.

        Returns:
            Unsubscribe function.
        """
        self.ensure_connected()

        if self._disconnect_timer:
            self._disconnect_timer.cancel()
            self._disconnect_timer = None

        # Register any new events with the Socket.IO client
        self._register_events(events)

        with self._lock:
            self._disconnect_generation += 1
            if resource_id not in self._listeners:
                self._listeners[resource_id] = []
            self._listeners[resource_id].append(handler)

            if ttl_seconds > 0:
                self._subscription_ttls[resource_id] = ttl_seconds
                self._start_subscription_timer_locked(resource_id)
            else:
                self._clear_subscription_timer_locked(resource_id)
                self._subscription_ttls.pop(resource_id, None)

        def unsubscribe() -> None:
            with self._lock:
                handlers = self._listeners.get(resource_id)
                if handlers and handler in handlers:
                    handlers.remove(handler)
                    if not handlers:
                        self._unsubscribe_resource_locked(resource_id)
                should_cleanup = not self._listeners

            if should_cleanup:
                generation = self._disconnect_generation
                self._disconnect_timer = threading.Timer(
                    self._DISCONNECT_DELAY,
                    self._delayed_disconnect,
                    args=(generation,),
                )
                self._disconnect_timer.daemon = True
                self._disconnect_timer.start()

        return unsubscribe

    def refresh_subscription(self, resource_id: str) -> bool:
        with self._lock:
            if resource_id not in self._subscription_ttls:
                return False

            self._start_subscription_timer_locked(resource_id)
            return True

    def _register_events(self, events: list[str]) -> None:
        """Register Socket.IO event handlers (idempotent — each event is registered once)."""
        subscriber = self

        with self._lock:
            for event_name in events:
                if event_name in self._registered_events:
                    continue
                self._registered_events.add(event_name)

                # If socket isn't created yet, the event will be registered when connect() runs
                if not self._sio:
                    continue

                def _make_handler(evt: str) -> Callable[..., Any]:
                    def _handler(data: Any) -> None:
                        resource_id = _extract_id_from_event(data)
                        if resource_id:
                            subscriber._dispatch(resource_id, evt, data)

                    return _handler

                self._sio.on(event_name, _make_handler(event_name))  # pyright: ignore[reportUnusedCallResult]

    def _delayed_disconnect(self, generation: int) -> None:
        self._disconnect(permanent=False, expected_generation=generation)

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
        self._disconnect(permanent=True)

    def _disconnect(self, permanent: bool, expected_generation: int | None = None) -> None:
        sio = None

        with self._lock:
            if expected_generation is not None:
                if expected_generation != self._disconnect_generation or self._listeners:
                    return
            if permanent:
                self._closed = True
            if self._disconnect_timer is not None:
                self._disconnect_timer.cancel()
                self._disconnect_timer = None
            sio = self._sio
            self._sio = None
            self._connected = False
            self._clear_subscription_timers_locked()
            self._listeners.clear()
            self._registered_events.clear()

        if sio:
            sio.disconnect()

    def _dispatch(self, resource_id: str, event_name: str, data: Any) -> None:
        with self._lock:
            handlers = list(self._listeners.get(resource_id, []))
        for handler in handlers:
            try:
                handler(event_name, data)
            except Exception:
                pass

    def _start_subscription_timer_locked(self, resource_id: str) -> None:
        ttl_seconds = self._subscription_ttls.get(resource_id)
        if ttl_seconds is None or ttl_seconds <= 0:
            return

        self._clear_subscription_timer_locked(resource_id)

        current_timer: threading.Timer | None = None

        def _expire() -> None:
            self._expire_subscription(resource_id, current_timer)

        current_timer = threading.Timer(ttl_seconds, _expire)
        current_timer.daemon = True
        self._subscription_timers[resource_id] = current_timer
        current_timer.start()

    def _clear_subscription_timer_locked(self, resource_id: str) -> None:
        timer = self._subscription_timers.pop(resource_id, None)
        if timer is not None:
            timer.cancel()

    def _clear_subscription_timers_locked(self) -> None:
        for timer in self._subscription_timers.values():
            timer.cancel()
        self._subscription_timers.clear()
        self._subscription_ttls.clear()

    def _expire_subscription(self, resource_id: str, current_timer: threading.Timer | None) -> None:
        with self._lock:
            if self._subscription_timers.get(resource_id) is not current_timer:
                return
            self._unsubscribe_resource_locked(resource_id)
            should_cleanup = not self._listeners
            generation = self._disconnect_generation

        if should_cleanup:
            self._disconnect_timer = threading.Timer(self._DISCONNECT_DELAY, self._delayed_disconnect, args=(generation,))
            self._disconnect_timer.daemon = True
            self._disconnect_timer.start()

    def _unsubscribe_resource_locked(self, resource_id: str) -> None:
        self._listeners.pop(resource_id, None)
        self._clear_subscription_timer_locked(resource_id)
        self._subscription_ttls.pop(resource_id, None)


def _extract_id_from_event(data: Any) -> str | None:
    """Extract resource ID from an event payload.

    Handles two payload shapes:
      - Wrapper: {sandbox: {id: ...}, ...} → nested resource ID
      - Direct: {id: ...} → top-level ID
    """
    if not isinstance(data, dict):
        return None
    for key in ("sandbox", "volume", "snapshot", "runner"):
        nested: object = data.get(key)  # pyright: ignore[reportUnknownVariableType]
        if isinstance(nested, dict):
            sid: object = nested.get("id")  # pyright: ignore[reportUnknownVariableType]
            if isinstance(sid, str):
                return sid
    top_id: object = data.get("id")  # pyright: ignore[reportUnknownVariableType]
    if isinstance(top_id, str):
        return top_id
    return None
