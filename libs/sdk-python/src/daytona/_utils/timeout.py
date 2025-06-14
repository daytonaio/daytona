# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import concurrent.futures
import functools
from typing import Any, Callable, Optional, ParamSpec, TypeVar

from .._utils.errors import DaytonaError

P = ParamSpec("P")
T = TypeVar("T")


def with_timeout(
    error_message: Optional[Callable[[Any, float], str]] = None,
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """Decorator to add a timeout mechanism with an optional custom error message.

    Args:
        error_message (Optional[Callable[[Any, float], str]]): A callable that accepts `self` and `timeout`,
                                                               and returns a string error message.
    """

    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        @functools.wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            # Get function argument names
            arg_names = func.__code__.co_varnames[: func.__code__.co_argcount]
            arg_dict = dict(zip(arg_names, args))

            # Extract self if method is bound
            self_instance = args[0] if args else None

            # Check for 'timeout' in kwargs first, then in positional arguments
            timeout = kwargs.get("timeout", arg_dict.get("timeout", None))

            if timeout is None or timeout == 0:
                # If timeout is None or 0, run the function normally
                return func(*args, **kwargs)

            if timeout < 0:
                raise DaytonaError("Timeout must be a non-negative number or None.")

            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(func, *args, **kwargs)
                try:
                    return future.result(timeout=timeout)
                except concurrent.futures.TimeoutError:
                    # Use custom error message if provided, otherwise default
                    msg = (
                        error_message(self_instance, timeout)
                        if error_message
                        else f"Function '{func.__name__}' exceeded timeout of {timeout} seconds."
                    )
                    raise TimeoutError(msg)  # pylint: disable=raise-missing-from

        return wrapper

    return decorator
