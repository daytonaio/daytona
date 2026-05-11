# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Common types for the Daytona Sessions product."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Callable, Literal, Optional

from .errors import DaytonaError


@dataclass
class SessionDisplay:
    """One display payload (e.g. an HTML table from pandas, a PNG from matplotlib)."""

    formats: list[str] = field(default_factory=list)
    """Mime types present in `data` (e.g. ['text/html', 'text/plain'])."""

    data: dict[str, str] = field(default_factory=dict)
    """Mime → payload (base64-encoded for binary mimes like image/png)."""


@dataclass
class SessionExecutionError:
    """A raised-error frame from the daemon (e.g. ZeroDivisionError)."""

    name: str
    value: Optional[str] = None
    traceback: Optional[str] = None


@dataclass
class SessionRunResult:
    """Aggregated result of a one-shot or streaming run."""

    stdout: str
    stderr: str
    error: Optional[SessionExecutionError]
    displays: list[SessionDisplay]
    duration_ms: int


@dataclass
class SessionAccess:
    """Signed direct-to-sandbox access bundle returned by the API.

    The SDK opens a WebSocket directly to `ws_url` on the in-sandbox session-daemon
    via the proxy chain — the URL is self-authenticating (the signed token lives in
    the proxy subdomain). `token` is exposed for revocation / observability; the
    SDK does NOT send it as an `Authorization` header.

    The SDK refreshes this bundle (via `GET /sessions/:id/access` for
    persistent contexts or `POST /sessions/transients` for one-shot handles)
    before `token_expires_at` minus a small skew.
    """

    http_url: str
    ws_url: str
    token: str
    token_expires_at: str


@dataclass
class SessionRef:
    """A user-facing reference to a persistent context."""

    id: str
    language: str
    cwd: Optional[str]
    created_at: str
    last_used_at: Optional[str]
    expires_at: str
    access: Optional[SessionAccess] = None


@dataclass
class SessionRunOptions:
    """Per-call options for one-shot run / streaming run."""

    language: Optional[Literal["python", "typescript", "javascript"]] = None
    template: Optional[str] = None
    context: Optional[SessionRef] = None
    env: Optional[dict[str, str]] = None
    timeout: Optional[int] = None


# Streaming handler callbacks. Kept loose-typed (callable) so users can pass either
# sync or async callables — the service internally awaits / calls as appropriate.
SessionStreamHandler = Callable[[Any], None]


class SessionInvalidatedError(DaytonaError):
    """SDK projection of HTTP 410 `error.name=ContextInvalidated`.

    Raised when the underlying sandbox has been rolled (death / snapshot drift /
    autostop). Callers should drop the context and re-create.
    """

    session_id: str
    invalidated_at: str

    def __init__(self, session_id: str, invalidated_at: str):
        super().__init__(f"Session {session_id} has been invalidated at {invalidated_at}")
        self.session_id = session_id
        self.invalidated_at = invalidated_at


class SessionExpiredError(DaytonaError):
    """SDK projection of HTTP 410 `error.name=ContextExpired`.

    `reason` distinguishes idle vs absolute TTL.
    """

    session_id: str
    expired_at: str
    reason: Literal["idle", "absolute"]

    def __init__(self, session_id: str, expired_at: str, reason: Literal["idle", "absolute"]):
        super().__init__(f"Session {session_id} expired ({reason}) at {expired_at}")
        self.session_id = session_id
        self.expired_at = expired_at
        self.reason = reason
