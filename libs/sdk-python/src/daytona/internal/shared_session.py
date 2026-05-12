# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from typing import Any

import aiohttp


class SharedAiohttpSession:
    """Funnels every async REST call through a single ``aiohttp.ClientSession``.

    ``aiohttp.TCPConnector`` requires a running event loop, so the session can only
    be created on the first awaited request. ``attach`` swaps each rest_client's
    ``request`` for a wrapper that lets the underlying call lazy-create its session,
    then propagates that session to every other attached rest_client so all SDK
    traffic shares one TCP/TLS pool.

    ``session``/``require_session`` expose the adopted session for direct WS and
    streaming use (multipart uploads, log follows, etc.).
    """

    def __init__(self) -> None:
        self._session: aiohttp.ClientSession | None = None
        self._rest_clients: list[Any] = []

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
            started_with_none = rest_client.pool_manager is None

            if not started_with_none:
                coordinator._adopt(rest_client.pool_manager)

            result = await original_request(*args, **kwargs)

            if started_with_none:
                coordinator._adopt(rest_client.pool_manager)

            return result

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
