# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import logging
from typing import Any

import aiohttp

logger = logging.getLogger(__name__)

# Connection-level errors that indicate the request was never processed by the
# server. Safe to retry on any HTTP method — async counterpart of the sync
# client's urllib3 RemoteDisconnected retry.
_RETRYABLE_CONN_ERRORS: tuple[type[BaseException], ...] = (
    aiohttp.ServerDisconnectedError,
    aiohttp.ClientOSError,
    aiohttp.ClientConnectorError,
    ConnectionResetError,
    BrokenPipeError,
)
_MAX_CONN_RETRIES = 2


class SharedAiohttpSession:
    """Funnels every async REST call through a single ``aiohttp.ClientSession``.

    ``aiohttp.TCPConnector`` requires a running event loop, so the session can only
    be created on the first awaited request. ``attach`` swaps each rest_client's
    ``request`` for a wrapper that lets the underlying call lazy-create its session,
    then propagates that session to every other attached rest_client so all SDK
    traffic shares one TCP/TLS pool.

    The wrapper also retries transient connection-level errors
    (``ServerDisconnectedError``, ``ClientOSError``/broken pipe,
    ``ClientConnectorError``) on any HTTP method, since those failures indicate
    the request was never processed by the server.  This is the async equivalent
    of ``urllib3_retry.RemoteDisconnectedRetry`` used by the sync client.

    ``session``/``require_session`` expose the adopted session for direct WS and
    streaming use (multipart uploads, log follows, etc.).
    """

    def __init__(self, **connector_overrides: float | None) -> None:
        """
        Args:
            **connector_overrides: Forwarded verbatim to
                ``aiohttp.TCPConnector``.  Only the keys explicitly passed
                override aiohttp's own defaults — e.g. callers can pass
                ``happy_eyeballs_delay=None`` to disable the RFC 8305
                IPv4/IPv6 race, or omit the kwarg entirely to inherit
                whatever default aiohttp ships with.
        """
        self._session: aiohttp.ClientSession | None = None
        self._rest_clients: list[Any] = []
        self._connector_overrides: dict[str, float | None] = dict(connector_overrides)

    def _ensure_session(self, rest_client: Any) -> None:
        """Lazily create the shared aiohttp session on first request.

        Runs before the rest_client's own lazy-create branch so the SDK
        controls the connector settings, not the generated ``rest.py``.
        """
        if self._session is not None and not self._session.closed:
            return

        connector_kwargs: dict[str, Any] = {
            "limit": getattr(rest_client, "maxsize", 100),
            "keepalive_timeout": 30,
        }
        ssl_context = getattr(rest_client, "ssl_context", None)
        if ssl_context is not None:
            connector_kwargs["ssl"] = ssl_context
        connector_kwargs.update(self._connector_overrides)

        self._session = aiohttp.ClientSession(
            connector=aiohttp.TCPConnector(**connector_kwargs),
            trust_env=True,
        )
        for client in self._rest_clients:
            client.pool_manager = self._session

    def attach(self, rest_client: Any) -> None:
        if not (hasattr(rest_client, "request") and hasattr(rest_client, "pool_manager")):
            raise RuntimeError(
                "rest_client API surface changed; SharedAiohttpSession needs updating"
                + " (expected `request` and `pool_manager` attributes)."
            )

        original_request = rest_client.request
        coordinator = self

        async def request_wrapper(*args: object, **kwargs: object) -> object:
            # No timeout injection: callers that need one pass _request_timeout
            # explicitly; everyone else gets aiohttp's session DEFAULT_TIMEOUT. The
            # TCPConnector limit on the shared session is what bounds concurrency.
            coordinator._ensure_session(rest_client)

            for attempt in range(1, _MAX_CONN_RETRIES + 2):
                started_with_none = rest_client.pool_manager is None

                if not started_with_none:
                    coordinator._adopt(rest_client.pool_manager)

                try:
                    result = await original_request(*args, **kwargs)
                except _RETRYABLE_CONN_ERRORS as e:
                    # On first-call failure the rest_client created its session but
                    # the request never completed; still adopt it so the next attempt
                    # reuses the shared pool.
                    if started_with_none and rest_client.pool_manager is not None:
                        coordinator._adopt(rest_client.pool_manager)
                    if attempt > _MAX_CONN_RETRIES:
                        raise
                    logger.debug(
                        "Retryable connection error (%s); retry %d/%d",
                        type(e).__name__,
                        attempt,
                        _MAX_CONN_RETRIES,
                    )
                    await asyncio.sleep(0.25 * attempt)
                    continue

                if started_with_none:
                    coordinator._adopt(rest_client.pool_manager)

                return result

            raise AssertionError("unreachable")

        rest_client.request = request_wrapper
        self._rest_clients.append(rest_client)

    def _adopt(self, candidate: aiohttp.ClientSession) -> None:
        # First-writer-wins: races only matter when both rest_clients fire concurrent
        # first calls; whichever's session lands first is kept, the other is replaced.
        if self._session is not None and not self._session.closed:
            return
        self._session = candidate
        for rest_client in self._rest_clients:
            if rest_client.pool_manager is not candidate:
                rest_client.pool_manager = candidate

    @property
    def session(self) -> aiohttp.ClientSession | None:
        """The shared session, or ``None`` until the first request fires."""
        return self._session

    def require_session(self) -> aiohttp.ClientSession:
        """Return the shared session; raise if no request has fired yet.

        Reachable when a ``Sandbox`` is built outside ``daytona.create/get/list``
        and a toolbox call is made before any main-api call has populated the pool.
        """
        if self._session is None:
            raise RuntimeError(
                "Shared aiohttp session not initialized; this happens when toolbox"
                + " APIs are called before any main API call. Use AsyncDaytona.create/get/list"
                + " to obtain a Sandbox first."
            )
        return self._session

    async def close(self) -> None:
        """Close the shared session. Idempotent."""
        if self._session is not None and not self._session.closed:
            await self._session.close()


def http_session_of(api_client: Any) -> aiohttp.ClientSession:
    """Return the shared ``aiohttp.ClientSession`` reachable through a generated ApiClient.

    The session attribute lives on ``ToolboxApiClientProxy`` at runtime but isn't
    declared on the generated stub, so direct access trips pyright. This helper
    localizes the duck-typed access and returns a typed session for chained calls.
    """
    return api_client.http_session
