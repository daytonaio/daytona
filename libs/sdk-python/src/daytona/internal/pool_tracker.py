# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings

_SATURATION_WARNING = (
    "Connection pool is nearing saturation ({active}/{maxsize} concurrent requests). "
    "New requests may queue and time out. "
    "Consider increasing `connection_pool_maxsize` in DaytonaConfig or setting it to None.\n"
    "  Example: DaytonaConfig(connection_pool_maxsize=None)"
)


class AsyncPoolSaturationTracker:
    """Tracks in-flight async HTTP requests against the configured connection pool limit.

    aiohttp's TCPConnector(limit=N) is a hard cap — when all N connections are busy,
    new requests silently queue and may time out. This tracker emits a warning when the
    number of concurrent long-lived requests (e.g. process.exec) reaches the limit so
    the user knows why requests are stalling.

    The asyncio event loop is single-threaded, so no synchronization is needed.
    """

    def __init__(self, maxsize: int | None) -> None:
        self._maxsize: int | None = maxsize
        self._active: int = 0

    def acquire(self) -> None:
        if self._maxsize is None:
            return
        if self._active >= self._maxsize:
            warnings.warn(
                _SATURATION_WARNING.format(active=self._active, maxsize=self._maxsize),
                stacklevel=4,
            )
        self._active += 1

    def release(self) -> None:
        if self._maxsize is None:
            return
        self._active = max(0, self._active - 1)
