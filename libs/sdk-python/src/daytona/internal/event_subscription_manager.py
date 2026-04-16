# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import threading
import uuid
from typing import Callable

from .event_dispatcher import AsyncEventDispatcher, AsyncEventHandler, EventHandler, SyncEventDispatcher

_SUBSCRIPTION_TTL: float = 300.0


class _Subscription:
    __slots__: tuple[str, ...] = ("resource_id", "unsubscribe_fn", "timer")

    resource_id: str
    unsubscribe_fn: Callable[[], None]
    timer: threading.Timer | asyncio.TimerHandle | None

    def __init__(
        self,
        resource_id: str,
        unsubscribe_fn: Callable[[], None],
    ) -> None:
        self.resource_id = resource_id
        self.unsubscribe_fn = unsubscribe_fn
        self.timer = None


class AsyncEventSubscriptionManager:
    """Tracks subscriptions by unique sub_id with optional TTL auto-expiry.

    Multiple callers subscribing to the same resource_id get independent sub_ids.
    """

    _dispatcher: AsyncEventDispatcher
    _subscriptions: dict[str, _Subscription]

    def __init__(self, dispatcher: AsyncEventDispatcher) -> None:
        self._dispatcher = dispatcher
        self._subscriptions = {}

    @property
    def dispatcher(self) -> AsyncEventDispatcher:
        return self._dispatcher

    def subscribe(
        self,
        resource_id: str,
        handler: AsyncEventHandler,
        events: list[str],
    ) -> str:
        unsubscribe_fn = self._dispatcher.subscribe(resource_id, handler, events)

        sub_id = uuid.uuid4().hex
        sub = _Subscription(resource_id=resource_id, unsubscribe_fn=unsubscribe_fn)
        self._subscriptions[sub_id] = sub
        self._start_timer(sub_id)

        return sub_id

    def refresh(self, sub_id: str) -> bool:
        sub = self._subscriptions.get(sub_id)
        if sub is None:
            return False

        self._start_timer(sub_id)
        return True

    def unsubscribe(self, sub_id: str) -> None:
        sub = self._subscriptions.pop(sub_id, None)
        if sub is None:
            return

        if sub.timer is not None:
            sub.timer.cancel()
        sub.unsubscribe_fn()

    def _start_timer(self, sub_id: str) -> None:
        sub = self._subscriptions.get(sub_id)
        if sub is None:
            return

        if sub.timer is not None:
            sub.timer.cancel()

        try:
            loop = asyncio.get_running_loop()
        except RuntimeError:
            return

        def _expire() -> None:
            popped = self._subscriptions.pop(sub_id, None)
            if popped is not None:
                popped.unsubscribe_fn()

        sub.timer = loop.call_later(_SUBSCRIPTION_TTL, _expire)

    def shutdown(self) -> None:
        for sub in self._subscriptions.values():
            if sub.timer is not None:
                sub.timer.cancel()
            sub.unsubscribe_fn()
        self._subscriptions.clear()


class SyncEventSubscriptionManager:
    """Thread-safe variant of AsyncEventSubscriptionManager."""

    _dispatcher: SyncEventDispatcher
    _subscriptions: dict[str, _Subscription]
    _lock: threading.Lock

    def __init__(self, dispatcher: SyncEventDispatcher) -> None:
        self._dispatcher = dispatcher
        self._subscriptions = {}
        self._lock = threading.Lock()

    @property
    def dispatcher(self) -> SyncEventDispatcher:
        return self._dispatcher

    def subscribe(
        self,
        resource_id: str,
        handler: EventHandler,
        events: list[str],
    ) -> str:
        unsubscribe_fn = self._dispatcher.subscribe(resource_id, handler, events)

        sub_id = uuid.uuid4().hex
        sub = _Subscription(resource_id=resource_id, unsubscribe_fn=unsubscribe_fn)

        with self._lock:
            self._subscriptions[sub_id] = sub
            self._start_timer_locked(sub_id)

        return sub_id

    def refresh(self, sub_id: str) -> bool:
        with self._lock:
            sub = self._subscriptions.get(sub_id)
            if sub is None:
                return False

            self._start_timer_locked(sub_id)
            return True

    def unsubscribe(self, sub_id: str) -> None:
        with self._lock:
            sub = self._subscriptions.pop(sub_id, None)
            if sub is None:
                return
            if sub.timer is not None:
                sub.timer.cancel()

        sub.unsubscribe_fn()

    # ------------------------------------------------------------------
    # Timer management — must be called while holding self._lock.
    # ------------------------------------------------------------------

    def _start_timer_locked(self, sub_id: str) -> None:
        sub = self._subscriptions.get(sub_id)
        if sub is None:
            return

        if sub.timer is not None:
            sub.timer.cancel()

        current_timer: threading.Timer | None = None

        def _expire() -> None:
            with self._lock:
                s = self._subscriptions.get(sub_id)
                if s is None or s.timer is not current_timer:
                    return
                _ = self._subscriptions.pop(sub_id, None)

            s.unsubscribe_fn()

        current_timer = threading.Timer(_SUBSCRIPTION_TTL, _expire)
        current_timer.daemon = True
        sub.timer = current_timer
        current_timer.start()

    def shutdown(self) -> None:
        with self._lock:
            subs = list(self._subscriptions.values())
            self._subscriptions.clear()

        for sub in subs:
            if sub.timer is not None:
                sub.timer.cancel()
            sub.unsubscribe_fn()
