# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0


class DaytonaError(Exception):
    """Base error for Daytona SDK."""


class DaytonaNotFoundError(DaytonaError):
    """Error for when a resource is not found."""
