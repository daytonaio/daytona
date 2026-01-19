# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import TypeVar

T = TypeVar("T")


def docs_ignore(obj: T) -> T:
    """Decorator to flag for documentation exclusion."""
    return obj
