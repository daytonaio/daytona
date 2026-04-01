# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from collections.abc import Mapping
from typing import Any


class DaytonaError(Exception):
    """Base error for Daytona SDK.

    Attributes:
        message (str): Error message
        status_code (int | None): HTTP status code if available
        headers (dict[str, Any]): Response headers
    """

    def __init__(
        self,
        message: str,
        status_code: int | None = None,
        headers: Mapping[str, Any] | None = None,
    ):
        """Initialize Daytona error.

        Args:
            message (str): Error message
            status_code (int | None): HTTP status code if available
            headers (Mapping[str, Any] | None): Response headers if available
        """
        super().__init__(message)
        self.status_code: int | None = status_code
        self.headers: dict[str, Any] = dict(headers or {})


class DaytonaBadRequestError(DaytonaError):
    """Error for malformed requests or invalid parameters (HTTP 400).

    Raised when the request is syntactically invalid or contains parameters
    that fail basic validation before reaching business logic.

    Example::

        try:
            sandbox = daytona.create(params)
        except DaytonaBadRequestError as e:
            print(f"Invalid request parameters: {e}")
    """


class DaytonaAuthenticationError(DaytonaError):
    """Error for authentication failures (HTTP 401).

    Raised when the API key or token is missing, expired, or invalid.

    Example::

        try:
            daytona = Daytona(DaytonaConfig(api_key="invalid"))
            sandbox = daytona.create()
        except DaytonaAuthenticationError:
            print("Invalid or missing API key")
    """


class DaytonaForbiddenError(DaytonaError):
    """Error for authorization failures (HTTP 403).

    Raised when the authenticated user does not have permission to perform
    the requested operation.

    Example::

        try:
            daytona.sandbox.delete(sandbox_id)
        except DaytonaForbiddenError:
            print("Not authorized to delete this sandbox")
    """


class DaytonaNotFoundError(DaytonaError):
    """Error for when a resource is not found (HTTP 404).

    Example::

        try:
            sandbox = daytona.sandbox.get("nonexistent-id")
        except DaytonaNotFoundError:
            print("Sandbox does not exist")
    """


class DaytonaConflictError(DaytonaError):
    """Error for resource conflicts (HTTP 409).

    Raised when an operation conflicts with the current state, such as
    creating a resource with a name that already exists.

    Example::

        try:
            daytona.snapshot.create(CreateSnapshotParams(name="my-snapshot"))
        except DaytonaConflictError:
            print("A snapshot with this name already exists")
    """


class DaytonaValidationError(DaytonaError):
    """Error for semantic validation failures (HTTP 422).

    Raised when the request is well-formed but the values fail business
    logic validation (e.g., unsupported resource class, invalid configuration).

    Example::

        try:
            sandbox = daytona.create(CreateSandboxFromImageParams(resources=...))
        except DaytonaValidationError as e:
            print(f"Validation failed: {e}")
    """


class DaytonaRateLimitError(DaytonaError):
    """Error for when rate limit is exceeded (HTTP 429).

    Example::

        try:
            for _ in range(1000):
                daytona.create()
        except DaytonaRateLimitError:
            print("Rate limit exceeded, back off and retry")
    """


class DaytonaServerError(DaytonaError):
    """Error for unexpected server-side failures (HTTP 5xx).

    Raised when the Daytona API encounters an internal error. These are
    typically transient and safe to retry with exponential backoff.

    Example::

        try:
            sandbox = daytona.create()
        except DaytonaServerError:
            print("Server error, retry later")
    """


class DaytonaTimeoutError(DaytonaError):
    """Error for when a timeout occurs.

    Raised when a polling operation (e.g., waiting for sandbox to start)
    exceeds the configured timeout.

    Example::

        try:
            sandbox = daytona.create(timeout=10)
        except DaytonaTimeoutError:
            print("Sandbox did not start within 10 seconds")
    """


class DaytonaConnectionError(DaytonaError):
    """Error for network-level connection failures.

    Raised when the SDK cannot reach the Daytona API due to network issues
    (DNS failure, connection refused, TLS error, etc.), with no HTTP response.

    Example::

        try:
            sandbox = daytona.create()
        except DaytonaConnectionError:
            print("Cannot reach Daytona API, check network connectivity")
    """
