# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from collections.abc import Mapping
from typing import Any

from typing_extensions import override

#: Wire-format ``source`` identifiers set by the translation layer when a
#: Daytona service stamps them on the wire envelope. ``source = None`` means
#: the response did not carry a structured envelope (treat as opaque).
SOURCE_API = "DAYTONA_API"
SOURCE_DAEMON = "DAYTONA_DAEMON"
SOURCE_PROXY = "DAYTONA_PROXY"


class DaytonaError(Exception):
    """Base error for Daytona SDK.

    Example:
        ```python
        try:
            sandbox = daytona.get("missing-sandbox")
        except DaytonaError as exc:
            print(exc.status_code)
            print(exc.code)
            print(exc.message)
        ```

    Attributes:
        message (str): Error message
        status_code (int | None): HTTP status code (set only for errors translated
            from an HTTP response; ``None`` for client-side errors).
        code (str | None): Machine-readable error code from the server envelope
            (``None`` for client-side errors).
        source (str | None): Originating service. ``None`` when the response
            did not carry a structured envelope. Otherwise one of
            :data:`SOURCE_API`, :data:`SOURCE_DAEMON`, :data:`SOURCE_PROXY`.
        headers (dict[str, Any]): Response headers (empty for client-side errors).
    """

    def __init__(
        self,
        message: str,
        status_code: int | None = None,
        headers: Mapping[str, Any] | None = None,
        code: str | None = None,
        source: str | None = None,
    ):
        """Initialize Daytona error.

        Args:
            message (str): Error message
            status_code (int | None): HTTP status code if the error came from a
                Daytona service response.
            headers (Mapping[str, Any] | None): Response headers if available
            code (str | None): Machine-readable error code from the wire envelope
            source (str | None): Originating service from the wire envelope.
                Left as ``None`` for SDK-side errors and for responses from
                services that don't emit the envelope.
        """
        super().__init__(message)
        self.message: str = message
        self.status_code: int | None = status_code
        self.code: str | None = code
        self.source: str | None = source
        self.headers: dict[str, Any] = dict(headers or {})

    @override
    def __repr__(self) -> str:
        parts = [
            f"message={self.message!r}",
            f"status_code={self.status_code!r}",
            f"code={self.code!r}",
        ]
        if self.source is not None:
            parts.append(f"source={self.source!r}")
        if self.headers:
            parts.append(f"headers={self.headers!r}")
        return f"{self.__class__.__name__}({', '.join(parts)})"


class DaytonaNotFoundError(DaytonaError):
    """Error for when a resource is not found (HTTP 404).

    Example:
        ```python
        try:
            sandbox.fs.download_file("/workspace/missing.txt")
        except DaytonaNotFoundError as exc:
            print(exc.status_code)
        ```
    """


class DaytonaAuthenticationError(DaytonaError):
    """Error for when authentication fails (HTTP 401).

    Example:
        ```python
        try:
            for sandbox in daytona.list():
                print(sandbox.id)
        except DaytonaAuthenticationError as exc:
            print(exc.status_code)
        ```
    """


class DaytonaForbiddenError(DaytonaError):
    """Error for when the request is forbidden (HTTP 403).

    Example:
        ```python
        try:
            daytona.get("sandbox-without-access")
        except DaytonaForbiddenError as exc:
            print(exc.message)
        ```
    """


#: Deprecated alias for :class:`DaytonaForbiddenError`. Kept so existing
#: ``except DaytonaAuthorizationError`` blocks continue to work; new code
#: should use ``DaytonaForbiddenError`` directly.
DaytonaAuthorizationError = DaytonaForbiddenError


class DaytonaRateLimitError(DaytonaError):
    """Error for when rate limit is exceeded (HTTP 429).

    Example:
        ```python
        try:
            for sandbox in daytona.list():
                print(sandbox.id)
        except DaytonaRateLimitError as exc:
            print(exc.code)
        ```
    """


class DaytonaConflictError(DaytonaError):
    """Error for when a resource conflict occurs (HTTP 409).

    Example:
        ```python
        try:
            params = CreateSandboxFromSnapshotParams(name="existing-sandbox")
            daytona.create(params)
        except DaytonaConflictError as exc:
            print(exc.code)
        ```
    """


class DaytonaBadRequestError(DaytonaError):
    """Error for when the request is malformed or fails server-side validation (HTTP 400).

    Example:
        ```python
        try:
            Image.debian_slim("3.8")
        except DaytonaBadRequestError as exc:
            print(exc.message)
        ```
    """


#: Deprecated alias for :class:`DaytonaBadRequestError`. Kept so existing
#: ``except DaytonaValidationError`` blocks continue to work; new code
#: should use ``DaytonaBadRequestError`` directly.
DaytonaValidationError = DaytonaBadRequestError


class DaytonaTimeoutError(DaytonaError):
    """Error for when a timeout occurs.

    Example:
        ```python
        try:
            sandbox.wait_for_sandbox_start(timeout=1)
        except DaytonaTimeoutError as exc:
            print(exc.message)
        ```
    """


class DaytonaConnectionError(DaytonaError):
    """Error for when a network connection fails (can't connect or mid-request drop)."""


class DaytonaConnectionTimeoutError(DaytonaConnectionError):
    """Error for when the transport layer times out connecting or reading from a Daytona service."""


class DaytonaGoneError(DaytonaError):
    """Error for HTTP 410 — the target resource is permanently gone."""


class DaytonaUnprocessableEntityError(DaytonaError):
    """Error for HTTP 422 — request is well-formed but semantically invalid."""


class DaytonaInternalServerError(DaytonaError):
    """Error for HTTP 500 — server-side bug or unhandled condition."""


class DaytonaBadGatewayError(DaytonaError):
    """Error for HTTP 502 — an upstream dependency rejected or dropped the request."""


class DaytonaServiceUnavailableError(DaytonaError):
    """Error for HTTP 503 — the service is temporarily refusing traffic."""


# Domain-specific error classes. Each inherits from the HTTP-status class
# that matches its server-side ``HTTPStatusCode()``, so callers can branch on
# either the precise class or the broader status class.


# Git.
class DaytonaGitAuthFailedError(DaytonaAuthenticationError):
    """Git auth credentials were rejected by the remote."""


class DaytonaGitRepoNotFoundError(DaytonaNotFoundError):
    """The requested git repository does not exist."""


class DaytonaGitBranchNotFoundError(DaytonaNotFoundError):
    """The requested git branch does not exist."""


class DaytonaGitBranchExistsError(DaytonaConflictError):
    """A git branch with this name already exists."""


class DaytonaGitPushRejectedError(DaytonaConflictError):
    """Git push was rejected (non-fast-forward / stale ref)."""


class DaytonaGitDirtyWorktreeError(DaytonaConflictError):
    """Worktree has uncommitted changes."""


class DaytonaGitMergeConflictError(DaytonaConflictError):
    """Git merge has conflicts that need manual resolution."""


# Filesystem.
class DaytonaFileNotFoundError(DaytonaNotFoundError):
    """Filesystem entry was not found."""


class DaytonaFileAccessDeniedError(DaytonaForbiddenError):
    """Insufficient permissions for the filesystem operation."""


# LSP.
class DaytonaLspServerNotInitializedError(DaytonaBadRequestError):
    """LSP server must be started via /lsp/start first."""


# Process / session.
class DaytonaProcessExecutionTimeoutError(DaytonaTimeoutError):
    """A process exceeded its configured execution timeout."""


class DaytonaProcessNotFoundError(DaytonaNotFoundError):
    """The requested process is not running."""


class DaytonaSessionEndedError(DaytonaGoneError):
    """The shell session has ended."""


class DaytonaCommandAlreadyCompletedError(DaytonaGoneError):
    """The shell command already finished."""


# Computer-use.
class DaytonaA11yUnavailableError(DaytonaServiceUnavailableError):
    """The accessibility (AT-SPI) bus is not reachable."""


class DaytonaRecordingStillActiveError(DaytonaConflictError):
    """The recording is still running; stop it first."""


class DaytonaRecordingFfmpegNotFoundError(DaytonaServiceUnavailableError):
    """ffmpeg binary is not installed; required for recording."""


STATUS_CODE_TO_ERROR: dict[int, type[DaytonaError]] = {
    400: DaytonaBadRequestError,
    401: DaytonaAuthenticationError,
    403: DaytonaForbiddenError,
    404: DaytonaNotFoundError,
    408: DaytonaTimeoutError,
    409: DaytonaConflictError,
    410: DaytonaGoneError,
    422: DaytonaUnprocessableEntityError,
    429: DaytonaRateLimitError,
    500: DaytonaInternalServerError,
    502: DaytonaBadGatewayError,
    503: DaytonaServiceUnavailableError,
    504: DaytonaTimeoutError,
}

# (source, code) → precise DaytonaError subclass. Lookup runs before the
# HTTP-status fallback so a domain code (e.g. ``FILE_NOT_FOUND``) yields the
# precise subclass rather than the generic 404 class.
#
# Codes are kept as inline string literals to mirror the TypeScript SDK and
# avoid coupling this module to the generated client packages. Drift is
# guarded by the cross-language code-catalog generator in ``hack/error-codes``.
CODE_TO_ERROR: dict[tuple[str, str], type[DaytonaError]] = {
    # Daemon: git
    (SOURCE_DAEMON, "GIT_AUTH_FAILED"): DaytonaGitAuthFailedError,
    (SOURCE_DAEMON, "GIT_REPO_NOT_FOUND"): DaytonaGitRepoNotFoundError,
    (SOURCE_DAEMON, "GIT_BRANCH_NOT_FOUND"): DaytonaGitBranchNotFoundError,
    (SOURCE_DAEMON, "GIT_BRANCH_EXISTS"): DaytonaGitBranchExistsError,
    (SOURCE_DAEMON, "GIT_PUSH_REJECTED"): DaytonaGitPushRejectedError,
    (SOURCE_DAEMON, "GIT_DIRTY_WORKTREE"): DaytonaGitDirtyWorktreeError,
    (SOURCE_DAEMON, "GIT_MERGE_CONFLICT"): DaytonaGitMergeConflictError,
    # Daemon: filesystem
    (SOURCE_DAEMON, "FILE_NOT_FOUND"): DaytonaFileNotFoundError,
    (SOURCE_DAEMON, "FILE_ACCESS_DENIED"): DaytonaFileAccessDeniedError,
    # Daemon: lsp
    (SOURCE_DAEMON, "LSP_SERVER_NOT_INITIALIZED"): DaytonaLspServerNotInitializedError,
    # Daemon: process / session
    (SOURCE_DAEMON, "PROCESS_EXECUTION_TIMEOUT"): DaytonaProcessExecutionTimeoutError,
    (SOURCE_DAEMON, "PROCESS_NOT_FOUND"): DaytonaProcessNotFoundError,
    (SOURCE_DAEMON, "SESSION_ENDED"): DaytonaSessionEndedError,
    (SOURCE_DAEMON, "COMMAND_ALREADY_COMPLETED"): DaytonaCommandAlreadyCompletedError,
    # Daemon: computer-use
    (SOURCE_DAEMON, "A11Y_UNAVAILABLE"): DaytonaA11yUnavailableError,
    (SOURCE_DAEMON, "RECORDING_STILL_ACTIVE"): DaytonaRecordingStillActiveError,
    (SOURCE_DAEMON, "RECORDING_FFMPEG_NOT_FOUND"): DaytonaRecordingFfmpegNotFoundError,
}


def error_class_from_status_code(status_code: int | None) -> type[DaytonaError]:
    """Map an HTTP status code to the corresponding DaytonaError subclass."""

    if status_code is None:
        return DaytonaError

    return STATUS_CODE_TO_ERROR.get(status_code, DaytonaError)


def _resolve_error_class(status_code: int | None, code: str | None, source: str | None) -> type[DaytonaError]:
    """(source, code) override first, HTTP status code otherwise."""
    if code and source:
        cls = CODE_TO_ERROR.get((source, code))
        if cls is not None:
            return cls
    return error_class_from_status_code(status_code)


def create_daytona_error(
    message: str,
    status_code: int | None = None,
    headers: Mapping[str, Any] | None = None,
    code: str | None = None,
    source: str | None = None,
) -> DaytonaError:
    """Create the appropriate DaytonaError subclass from structured error metadata.

    Resolution order: ``(source, code)`` exact match → HTTP status code → base
    :class:`DaytonaError`.
    """

    error_cls = _resolve_error_class(status_code, code, source)
    return error_cls(
        message,
        status_code=status_code,
        headers=headers,
        code=code,
        source=source,
    )
