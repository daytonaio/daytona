# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import logging
import random
from typing import Any

import aiohttp

logger = logging.getLogger(__name__)

# Backoff = (base * attempt) + uniform(0, jitter).  Jitter de-correlates
# concurrent retries so a fleet that all hit the same transient blip doesn't
# come back in synchronized waves and hammer a recovering server.
_RETRY_BACKOFF_BASE_S = 0.25
_RETRY_BACKOFF_JITTER_S = 0.1


def _retry_backoff_seconds(attempt: int) -> float:
    """Return the backoff delay for ``attempt`` (1-based)."""
    return _RETRY_BACKOFF_BASE_S * attempt + random.random() * _RETRY_BACKOFF_JITTER_S


# Errors raised by ``aiohttp.TCPConnector`` *before* any bytes are written to
# the socket — TCP ``connect()`` itself failed (DNS resolution, RST during
# handshake, SSL failure on connect, …).  The application cannot have seen
# the request, so retrying is safe on any HTTP method, including non-idempotent
# ones.  ``ClientConnectorError`` is a subclass of ``ClientOSError``, so it
# must be matched first.
_ALWAYS_RETRYABLE_CONN_ERRORS: tuple[type[BaseException], ...] = (aiohttp.ClientConnectorError,)

# Errors that *could* fire after the request was partially or fully written to
# the socket — i.e. the server may have started processing it before the
# connection dropped.  Retrying is only safe for methods that are idempotent
# by HTTP semantics (RFC 9110 §9.2.2).
_IDEMPOTENT_ONLY_RETRYABLE_CONN_ERRORS: tuple[type[BaseException], ...] = (
    aiohttp.ServerDisconnectedError,
    aiohttp.ClientOSError,
    ConnectionResetError,
    BrokenPipeError,
)

# Methods that are idempotent under HTTP semantics (RFC 9110 §9.2.2).
# Matches urllib3's ``DEFAULT_ALLOWED_METHODS`` used by the sync client.
_IDEMPOTENT_METHODS: frozenset[str] = frozenset({"GET", "HEAD", "OPTIONS", "TRACE", "PUT", "DELETE"})

_MAX_CONN_RETRIES = 2


def _extract_method(args: tuple[object, ...], kwargs: dict[str, object]) -> str | None:
    """Pull the HTTP method out of a ``rest_client.request`` call.

    The generated ``RESTClientObject.request`` takes ``method`` as the first
    positional argument; the generated ``ApiClient.__call_api`` always passes
    it positionally.  We also accept ``method=`` as a keyword for robustness
    in case future generator changes shift to kwargs.
    """
    raw: object | None = None
    if args:
        raw = args[0]
    elif "method" in kwargs:
        raw = kwargs["method"]
    if isinstance(raw, str):
        return raw.upper()
    return None


class SharedAiohttpSession:
    """Funnels every async REST call through a single ``aiohttp.ClientSession``.

    ``aiohttp.TCPConnector`` requires a running event loop, so the session can only
    be created on the first awaited request. ``attach`` swaps each rest_client's
    ``request`` for a wrapper that lets the underlying call lazy-create its session,
    then propagates that session to every other attached rest_client so all SDK
    traffic shares one TCP/TLS pool.

    The wrapper also retries transient connection-level errors:

    - ``ClientConnectorError`` (TCP ``connect()`` failure — nothing was sent)
      is retried on **any** HTTP method, since the application cannot have
      seen the request.
    - ``ServerDisconnectedError``, ``ClientOSError``, ``ConnectionResetError``,
      and ``BrokenPipeError`` are retried **only for idempotent methods**
      (``GET``/``HEAD``/``OPTIONS``/``TRACE``/``PUT``/``DELETE``) because the
      server may already have started processing the request when the
      connection dropped.

    This is the async counterpart of ``urllib3_retry.RemoteDisconnectedRetry``
    used by the sync client, hardened against the rare double-execution window
    of non-idempotent requests.

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

            method = _extract_method(args, kwargs)
            is_idempotent = method in _IDEMPOTENT_METHODS

            for attempt in range(1, _MAX_CONN_RETRIES + 2):
                started_with_none = rest_client.pool_manager is None

                if not started_with_none:
                    coordinator._adopt(rest_client.pool_manager)

                # Shallow-copy mutable kwargs so a retry doesn't observe state the
                # first attempt left behind.  The generated rest_client mutates
                # ``headers`` in place (adds ``Content-Type`` for JSON, deletes it
                # for multipart) — without this copy the second attempt of a
                # multipart request would re-send without a Content-Type header.
                attempt_kwargs = dict(kwargs)
                headers = attempt_kwargs.get("headers")
                if isinstance(headers, dict):
                    attempt_kwargs["headers"] = headers.copy()

                try:
                    result = await original_request(*args, **attempt_kwargs)
                except _ALWAYS_RETRYABLE_CONN_ERRORS as e:
                    # Pre-flight failure (TCP connect): the request was never sent,
                    # so retrying is safe regardless of HTTP method.
                    if started_with_none and rest_client.pool_manager is not None:
                        coordinator._adopt(rest_client.pool_manager)
                    if attempt > _MAX_CONN_RETRIES:
                        raise
                    logger.debug(
                        "Retryable connect-time error (%s) for %s; retry %d/%d",
                        type(e).__name__,
                        method or "<unknown method>",
                        attempt,
                        _MAX_CONN_RETRIES,
                    )
                    await asyncio.sleep(_retry_backoff_seconds(attempt))
                    continue
                except _IDEMPOTENT_ONLY_RETRYABLE_CONN_ERRORS as e:
                    # The connection may have failed *after* the server began
                    # processing the request, so we only retry methods that are
                    # idempotent under HTTP semantics.  Non-idempotent failures
                    # surface to the caller unchanged.
                    if started_with_none and rest_client.pool_manager is not None:
                        coordinator._adopt(rest_client.pool_manager)
                    if not is_idempotent or attempt > _MAX_CONN_RETRIES:
                        raise
                    logger.debug(
                        "Retryable connection error (%s) for idempotent %s; retry %d/%d",
                        type(e).__name__,
                        method,
                        attempt,
                        _MAX_CONN_RETRIES,
                    )
                    await asyncio.sleep(_retry_backoff_seconds(attempt))
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
