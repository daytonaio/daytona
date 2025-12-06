# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from enum import Enum
from typing import TypeVar

E = TypeVar("E", bound=Enum)


def to_enum(enum_class: type[E], value: str) -> E | None:
    """Convert a string to an enum.

    Args:
        enum_class (type): The enum class to convert to.
        value (str): The value to convert to an enum.

    Returns:
        The enum value, or None if the value is not a valid enum.
    """
    if isinstance(value, enum_class):
        return value
    str_value = str(value)
    if str_value in enum_class._value2member_map_:
        return enum_class(str_value)
    return None
