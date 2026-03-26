# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import http.client

import urllib3
from typing_extensions import override


class RemoteDisconnectedRetry(urllib3.Retry):
    """urllib3.Retry subclass that retries RemoteDisconnected on any HTTP method.

    urllib3 classifies both RemoteDisconnected and IncompleteRead as "read errors"
    (ProtocolError) and, by default, only retries them for idempotent methods (GET,
    HEAD, etc.) — not POST.

    RemoteDisconnected ("Remote end closed connection without response") means the
    server sent zero bytes, so the request was never processed (stale connection
    pool). Retrying on any HTTP method is safe.

    IncompleteRead means the server already started sending a response, so it did
    process the request. We must NOT retry POST on IncompleteRead to avoid executing
    an operation twice.

    Note: in an extremely rare case the daemon (sandbox) could crash after
    processing a request but before writing any response bytes, which would
    also surface as RemoteDisconnected. However, this is not a practical
    concern — if the daemon crashes, it will be down when the retry arrives,
    so the retried request will fail with a connection error rather than
    executing the operation a second time.

    Implementation: we override ``_is_read_error`` to return ``False`` for
    RemoteDisconnected. This causes urllib3's ``increment()`` to fall into the
    generic "other error" branch, which retries regardless of HTTP method.
    All other errors (IncompleteRead, ReadTimeoutError, etc.) keep their
    default behavior.  No mutable state, so this is thread-safe.
    """

    @override
    def _is_read_error(self, err: Exception) -> bool:
        if _is_remote_disconnected(err):
            return False
        return super()._is_read_error(err)


def _is_remote_disconnected(err: Exception) -> bool:
    """Return True if ``err`` is a ProtocolError wrapping RemoteDisconnected."""
    if not isinstance(err, urllib3.exceptions.ProtocolError):
        return False
    cause = err.args[1] if len(err.args) > 1 else None
    return isinstance(cause, http.client.RemoteDisconnected)
