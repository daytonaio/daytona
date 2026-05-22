# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import weakref

import aiohttp
import httpx

DEFAULT_POOL_SIZE = 250

# TCP three-way handshake cap (httpx ``connect`` / aiohttp ``sock_connect``).
# Matches aiohttp's session DEFAULT_TIMEOUT.sock_connect so streaming callers
# and regular REST calls share the same TCP budget.
CONNECT_TIMEOUT_S = 30.0
# Overall connection-establishment cap (httpx ``pool`` / aiohttp ``connect``):
# pool wait + DNS + TCP + TLS handshake. Generous enough to absorb slow TLS
# handshakes during concurrent-upload bursts.
POOL_TIMEOUT_S = 60.0


def _build_limits(pool_size: int | None) -> httpx.Limits:
    if pool_size is None:
        return httpx.Limits(max_connections=None, max_keepalive_connections=None)
    keepalive = min(max(pool_size // 4, 20), pool_size)
    return httpx.Limits(max_connections=pool_size, max_keepalive_connections=keepalive)


def _build_timeout() -> httpx.Timeout:
    return httpx.Timeout(connect=CONNECT_TIMEOUT_S, read=None, write=None, pool=POOL_TIMEOUT_S)


def request_timeout(timeout: float) -> httpx.Timeout:
    """Build a per-request timeout that preserves connect/pool defaults.

    httpx replaces the entire timeout config when you pass ``timeout=N``,
    which would negate the client-level connect and pool timeouts.  This
    helper sets read/write to the caller's value while keeping connect and
    pool at the safe defaults.

    A *timeout* of ``0`` is treated as "no timeout" (``None``).
    """
    rw = None if timeout == 0 else timeout
    return httpx.Timeout(connect=CONNECT_TIMEOUT_S, read=rw, write=rw, pool=POOL_TIMEOUT_S)


def build_async_http_client(pool_size: int | None = DEFAULT_POOL_SIZE) -> httpx.AsyncClient:
    return httpx.AsyncClient(limits=_build_limits(pool_size), timeout=_build_timeout())


def build_sync_http_client(pool_size: int | None = DEFAULT_POOL_SIZE) -> httpx.Client:
    client = httpx.Client(limits=_build_limits(pool_size), timeout=_build_timeout())
    _attach_self_finalizer(client)
    return client


def _attach_self_finalizer(client: httpx.Client) -> None:
    """Close the underlying transports when *client* itself is garbage-collected.

    Anchoring the finalizer on the client (rather than on the owning ``Daytona``)
    means the pool survives as long as anything still references it — the
    parent ``Daytona`` or any ``Sandbox`` it returned — and closes
    deterministically only when the last reference drops. This prevents
    ``Daytona().create()``-style usage from prematurely closing the shared
    client while the returned sandbox is still in use.

    The callback captures only the transport objects (not the client), so it
    introduces no self-reference that would keep the client alive forever.
    """
    transport = client._transport
    mounts = tuple(t for t in client._mounts.values() if t is not None)

    def _close() -> None:
        for t in (transport, *mounts):
            try:
                t.close()
            except Exception:  # noqa: BLE001 - best-effort cleanup
                pass

    _ = weakref.finalize(client, _close)


def aiohttp_request_timeout(timeout: float | None) -> aiohttp.ClientTimeout:
    """Build a per-request aiohttp timeout that keeps connect/pool defaults intact.

    Aiohttp mirror of :func:`request_timeout`. ``timeout=0`` or ``None`` means
    no read deadline; otherwise ``sock_read`` is the caller's value. ``total``
    stays unbounded so streaming callers (long downloads/uploads) aren't wall-
    clock-capped, while ``connect``/``sock_connect`` keep a slow op from
    starving the pool.
    """
    sock_read = None if not timeout else float(timeout)
    return aiohttp.ClientTimeout(
        total=None,
        connect=POOL_TIMEOUT_S,
        sock_connect=CONNECT_TIMEOUT_S,
        sock_read=sock_read,
    )
