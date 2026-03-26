# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from collections.abc import Mapping
from typing import Any


class DaytonaError(Exception):
    """Base error for Daytona SDK.

    Example:
        ```python
        try:
            sandbox = daytona.get("missing-sandbox")
        except DaytonaError as exc:
            print(exc.status_code)
            print(exc.error_code)
            print(exc.message)
        ```

    Attributes:
        message (str): Error message
        status_code (int | None): HTTP status code if available
        error_code (str | None): Machine-readable error code if available
        headers (dict[str, Any]): Response headers
    """

    def __init__(
        self,
        message: str,
        status_code: int | None = None,
        headers: Mapping[str, Any] | None = None,
        error_code: str | None = None,
    ):
        """Initialize Daytona error.

        Args:
            message (str): Error message
            status_code (int | None): HTTP status code if available
            headers (Mapping[str, Any] | None): Response headers if available
            error_code (str | None): Machine-readable error code if available
        """
        super().__init__(message)
        self.message: str = message
        self.status_code: int | None = status_code
        self.error_code: str | None = error_code
        self.headers: dict[str, Any] = dict(headers or {})


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
            daytona.list()
        except DaytonaAuthenticationError as exc:
            print(exc.status_code)
        ```
    """


class DaytonaAuthorizationError(DaytonaError):
    """Error for when the request is forbidden (HTTP 403).

    Example:
        ```python
        try:
            daytona.get("sandbox-without-access")
        except DaytonaAuthorizationError as exc:
            print(exc.message)
        ```
    """


class DaytonaRateLimitError(DaytonaError):
    """Error for when rate limit is exceeded (HTTP 429).

    Example:
        ```python
        try:
            daytona.list()
        except DaytonaRateLimitError as exc:
            print(exc.error_code)
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
            print(exc.error_code)
        ```
    """


class DaytonaValidationError(DaytonaError):
    """Error for when input validation fails (HTTP 400 or client-side validation).

    Example:
        ```python
        try:
            Image.debian_slim("3.8")
        except DaytonaValidationError as exc:
            print(exc.message)
        ```
    """


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
    """Error for when a network connection fails.

    Example:
        ```python
        try:
            pty_handle.wait_for_connection()
        except DaytonaConnectionError as exc:
            print(exc.message)
        ```
    """


STATUS_CODE_TO_ERROR: dict[int, type[DaytonaError]] = {
    400: DaytonaValidationError,
    401: DaytonaAuthenticationError,
    403: DaytonaAuthorizationError,
    404: DaytonaNotFoundError,
    409: DaytonaConflictError,
    429: DaytonaRateLimitError,
}


def error_class_from_status_code(status_code: int | None) -> type[DaytonaError]:
    """Map an HTTP status code to the corresponding DaytonaError subclass."""

    if status_code is None:
        return DaytonaError

    return STATUS_CODE_TO_ERROR.get(status_code, DaytonaError)


def create_daytona_error(
    message: str,
    status_code: int | None = None,
    headers: Mapping[str, Any] | None = None,
    error_code: str | None = None,
) -> DaytonaError:
    """Create the appropriate DaytonaError subclass from structured error metadata."""

    error_cls = error_class_from_status_code(status_code)
    return error_cls(message, status_code=status_code, headers=headers, error_code=error_code)
