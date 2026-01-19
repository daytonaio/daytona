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


class DaytonaNotFoundError(DaytonaError):
    """Error for when a resource is not found."""


class DaytonaRateLimitError(DaytonaError):
    """Error for when rate limit is exceeded."""


class DaytonaTimeoutError(DaytonaError):
    """Error for when a timeout occurs."""
