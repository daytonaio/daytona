# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings
from enum import Enum, EnumMeta
from typing import Any, Callable, TypeVar, cast

from typing_extensions import override

E = TypeVar("E", bound=Enum)
T = TypeVar("T")


def deprecated_alias(old_name: str, new_name: str) -> Callable[[type[T]], type[T]]:
    """Decorator to mark a class or enum as deprecated with an alias.

    Args:
        old_name (str): The deprecated name/alias
        new_name (str): The new name that should be used instead

    Returns:
        A wrapped class that issues deprecation warnings when used
    """

    def decorator(cls: type[T]) -> type[T]:
        # Create warning message once
        warning_message = (
            f"`{old_name}` is deprecated. Please use `{new_name}` instead. "
            + "This will be removed in a future version."
        )

        if issubclass(cls, Enum):

            class DeprecatedEnumMeta(EnumMeta):  # pylint: disable=unused-variable
                @override
                def __getattribute__(cls, name: str) -> Any:
                    if not name.startswith("_"):
                        warnings.warn(warning_message, DeprecationWarning, stacklevel=2)
                    return super().__getattribute__(name)

            # Create the deprecated enum class with optimized creation
            class DeprecatedEnum(Enum, metaclass=DeprecatedEnumMeta):
                def __new__(cls, value: object) -> "DeprecatedEnum":
                    obj = object.__new__(cls)
                    obj._value_ = value
                    return obj

                @override
                def __eq__(self, other: object) -> bool:
                    return self.value == getattr(other, "value", other)

            # Add enum members and copy metadata in one pass
            for item in cls:
                setattr(DeprecatedEnum, item.name, item.value)

            # Copy metadata attributes directly
            for attr in ("__module__", "__qualname__", "__name__", "__doc__"):
                setattr(
                    DeprecatedEnum,
                    attr,
                    getattr(cls, attr) if attr != "__name__" else old_name,
                )

            return cast(type[T], DeprecatedEnum)

        # For non-enum classes, create a wrapper class that preserves type hints
        class WrappedClass(cls):  # type: ignore
            def __new__(cls, *args: object, **kwargs: object) -> "WrappedClass":
                warnings.warn(warning_message, DeprecationWarning, stacklevel=2)
                return super().__new__(cls)  # pylint: disable=no-value-for-parameter

            def __init__(self, *args: object, **kwargs: object) -> None:
                warnings.warn(warning_message, DeprecationWarning, stacklevel=2)
                super().__init__(*args, **kwargs)

        # Copy class attributes and metadata
        WrappedClass.__name__ = old_name
        WrappedClass.__qualname__ = cls.__qualname__
        WrappedClass.__module__ = cls.__module__
        WrappedClass.__doc__ = cls.__doc__

        # Copy annotations if they exist
        if hasattr(cls, "__annotations__"):
            WrappedClass.__annotations__ = dict(cls.__annotations__)

        # Copy any additional attributes from the original class
        for attr, value in cls.__dict__.items():
            if not attr.startswith("__"):
                setattr(WrappedClass, attr, value)

        return cast(type[T], WrappedClass)

    return decorator
