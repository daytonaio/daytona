# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import Optional


class DaytonaError(Exception):
    """Base error for Daytona SDK."""

    def __init__(self, message: str, status_code: Optional[int] = None, headers: Optional[dict] = None):
        """Initialize Daytona error.

        Args:
            message: Error message
            status_code: HTTP status code if available
            headers: Response headers if available
        """
        super().__init__(message)
        self.status_code = status_code
        self.headers = headers or {}


class DaytonaNotFoundError(DaytonaError):
    """Error for when a resource is not found."""


class DaytonaRateLimitError(DaytonaError):
    """Error for when rate limit is exceeded."""


class DaytonaTimeoutError(DaytonaError):
    """Error for when a timeout occurs."""
